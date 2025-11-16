package controllers

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/heru-oktafian/api-retail/models"
	"github.com/heru-oktafian/scafold/config"
	"github.com/heru-oktafian/scafold/framework"
	"github.com/heru-oktafian/scafold/helpers"
	"github.com/heru-oktafian/scafold/middlewares"
	"github.com/heru-oktafian/scafold/responses"
	"github.com/heru-oktafian/scafold/utils"
	"gorm.io/gorm"
)

// CreateSaleReturnTransaction adalah fungsi untuk membuat transaksi retur penjualan baru
func CreateSaleReturnTransaction(c *framework.Ctx) error {
	nowWIB := time.Now().In(utils.Location)

	subscriptionType, _ := middlewares.GetClaimsToken(c.Request, "subscription_type")
	branchID, _ := middlewares.GetClaimsToken(c.Request, "branch_id")
	userID, _ := middlewares.GetClaimsToken(c.Request, "user_id")

	db := config.DB
	var req models.SaleReturnRequest // pastikan struct request ini ada
	err := c.BodyParser(&req)
	if err != nil {
		return responses.JSONResponse(c, http.StatusBadRequest, "Body permintaan tidak valid", err.Error())
	}

	if req.SaleReturn.Payment == "" {
		req.SaleReturn.Payment = "paid_by_cash"
	}

	req.SaleReturn.UserID = userID
	req.SaleReturn.BranchID = branchID

	err = utils.ValidateStruct(req.SaleReturn)
	if err != nil {
		return responses.JSONResponse(c, http.StatusBadRequest, "Validasi input retur penjualan gagal", err.Error())
	}

	for _, item := range req.SaleReturnItems {
		err = utils.ValidateStruct(item)
		if err != nil {
			return responses.JSONResponse(c, http.StatusBadRequest, "Validasi salah satu item retur penjualan gagal", err.Error())
		}
	}

	var returnDate time.Time
	if req.SaleReturn.ReturnDate == "" {
		returnDate = nowWIB
	} else {
		returnDate, err = time.Parse("2006-01-02", req.SaleReturn.ReturnDate)
		if err != nil {
			return responses.JSONResponse(c, http.StatusBadRequest, "Format return_date tidak valid. Gunakan YYYY-MM-DD.", err.Error())
		}
	}

	tx := db.Begin()
	if tx.Error != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal memulai transaksi database", tx.Error.Error())
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Validasi apakah sale_id valid
	var sale models.Sales
	err = tx.Where("id = ?", req.SaleReturn.SaleId).First(&sale).Error
	if err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return responses.JSONResponse(c, http.StatusNotFound, fmt.Sprintf("Penjualan dengan ID %s tidak ditemukan", req.SaleReturn.SaleId), err.Error())
		}
		return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal mengambil data penjualan", err.Error())
	}

	saleReturnID := helpers.GenerateID("SRT")
	saleReturn := models.SaleReturns{
		ID:         saleReturnID,
		SaleId:     req.SaleReturn.SaleId,
		ReturnDate: returnDate,
		BranchID:   branchID,
		Payment:    req.SaleReturn.Payment,
		UserID:     userID,
		CreatedAt:  nowWIB,
		UpdatedAt:  nowWIB,
	}

	var totalReturn int
	var saleReturnItems []models.SaleReturnItems

	for _, item := range req.SaleReturnItems {
		parsedExpiredDate, err := time.Parse("2006-01-02", item.ExpiredDate)
		if err != nil {
			tx.Rollback()
			return responses.JSONResponse(c, http.StatusBadRequest, fmt.Sprintf("expired_date tidak valid untuk produk %s", item.ProductId), err.Error())
		}

		// Validasi item berasal dari sale_id
		// Ambil item penjualan untuk sale_id + product_id
		var saleItem models.SaleItems
		err = tx.Where("sale_id = ? AND product_id = ?", req.SaleReturn.SaleId, item.ProductId).First(&saleItem).Error
		if err != nil {
			tx.Rollback()
			return responses.JSONResponse(c, http.StatusBadRequest, fmt.Sprintf("Produk %s tidak ditemukan pada penjualan asal", item.ProductId), err.Error())
		}

		// Ambil total qty yang sudah diretur sebelumnya
		var totalReturnedQty int64
		err = tx.Model(&models.SaleReturnItems{}).
			Select("COALESCE(SUM(qty), 0)").
			Where("product_id = ? AND sale_return_id IN (SELECT id FROM sale_returns WHERE sale_id = ?)", item.ProductId, req.SaleReturn.SaleId).
			Scan(&totalReturnedQty).Error

		if err != nil {
			tx.Rollback()
			return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal memeriksa retur sebelumnya", err.Error())
		}

		// Validasi jika qty retur melebihi qty penjualan
		if int(totalReturnedQty)+item.Qty > saleItem.Qty {
			tx.Rollback()
			return responses.JSONResponse(c, http.StatusBadRequest, fmt.Sprintf("Total qty retur untuk produk %s melebihi jumlah yang dijual. Dijual: %d, Sudah Diretur: %d, Retur Ini: %d",
				item.ProductId, saleItem.Qty, totalReturnedQty, item.Qty), nil)
		}

		// Ambil informasi produk
		var product models.Product
		err = tx.Where("id = ?", item.ProductId).First(&product).Error
		if err != nil {
			tx.Rollback()
			return responses.JSONResponse(c, http.StatusInternalServerError, fmt.Sprintf("Gagal mengambil info produk untuk %s", item.ProductId), err.Error())
		}

		actualQtyToReduce := item.Qty

		// Update stok
		err = tx.Model(&models.Product{}).Where("id = ?", item.ProductId).
			Update("stock", gorm.Expr("stock + ?", actualQtyToReduce)).Error
		if err != nil {
			tx.Rollback()
			return responses.JSONResponse(c, http.StatusInternalServerError, fmt.Sprintf("Gagal memperbarui stok untuk produk %s", item.ProductId), err.Error())
		}

		subTotal := saleItem.Price * item.Qty
		totalReturn += subTotal

		saleReturnItems = append(saleReturnItems, models.SaleReturnItems{
			ID:           helpers.GenerateID("SRI"),
			SaleReturnId: saleReturnID,
			ProductId:    item.ProductId,
			Price:        saleItem.Price,
			Qty:          item.Qty,
			SubTotal:     subTotal,
			ExpiredDate:  parsedExpiredDate,
		})

	}

	saleReturn.TotalReturn = totalReturn

	err = tx.Create(&saleReturn).Error
	if err != nil {
		tx.Rollback()
		return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal membuat retur penjualan", err.Error())
	}

	err = tx.CreateInBatches(&saleReturnItems, len(saleReturnItems)).Error
	if err != nil {
		tx.Rollback()
		return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal membuat item retur penjualan", err.Error())
	}

	// Tambahkan ke laporan transaksi
	transactionReportID := helpers.GenerateID("TRX")
	transactionReport := models.TransactionReports{
		ID:              transactionReportID,
		TransactionType: models.SaleReturn,
		UserID:          userID,
		BranchID:        branchID,
		Total:           saleReturn.TotalReturn,
		Payment:         saleReturn.Payment,
		CreatedAt:       nowWIB,
		UpdatedAt:       nowWIB,
	}
	err = tx.Create(&transactionReport).Error
	if err != nil {
		tx.Rollback()
		return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal membuat laporan transaksi retur penjualan", err.Error())
	}

	// Kurangi kuota jika berlangganan quota
	if subscriptionType == "quota" {
		var branch models.Branch
		err = tx.Where("id = ?", branchID).First(&branch).Error
		if err != nil {
			tx.Rollback()
			return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal mengambil info cabang untuk kuota", err.Error())
		}

		if branch.Quota > 0 {
			branch.Quota -= 1
			err = tx.Save(&branch).Error
			if err != nil {
				tx.Rollback()
				return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal memperbarui kuota cabang", err.Error())
			}
		} else {
			tx.Rollback()
			return responses.JSONResponse(c, http.StatusBadRequest, "Kuota cabang sudah habis", nil)
		}
	}

	err = tx.Commit().Error
	if err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal melakukan commit transaksi", err.Error())
	}

	return responses.JSONResponse(c, http.StatusOK, "Transaksi retur penjualan berhasil dibuat", framework.Map{
		"id":           saleReturn.ID,
		"sale_id":      saleReturn.SaleId,
		"return_date":  utils.FormatIndonesianDate(saleReturn.ReturnDate),
		"total_return": saleReturn.TotalReturn,
		"payment":      saleReturn.Payment,
		"items":        saleReturnItems,
	})
}

// GetSaleItemsForReturn digunakan untuk mengambil item penjualan yang bisa diretur
func GetSaleItemsForReturn(c *framework.Ctx) error {
	saleId := c.Query("sale_id")
	if saleId == "" {
		return responses.JSONResponse(c, http.StatusBadRequest, "sale_id wajib diisi", nil)
	}

	var results []struct {
		ProID    string `json:"pro_id"`
		ProName  string `json:"pro_name"`
		Stock    int    `json:"stock"`
		UnitID   string `json:"unit_id"`
		UnitName string `json:"unit_name"`
		Price    int    `json:"price"`
	}

	err := config.DB.Raw(`
        SELECT 
            A.product_id AS pro_id,
            B.name AS pro_name,
            A.qty AS stock,
            B.unit_id,
            C.name AS unit_name,
            A.price
        FROM sale_items A
        LEFT JOIN products B ON B.id = A.product_id
        LEFT JOIN units C ON C.id = B.unit_id
        LEFT JOIN (
            SELECT 
                sri.product_id, 
                SUM(sri.qty) AS total_returned
            FROM sale_return_items sri
            INNER JOIN sale_returns sr ON sri.sale_return_id = sr.id
            WHERE sr.sale_id = ?
            GROUP BY sri.product_id
        ) R ON R.product_id = A.product_id
        WHERE A.sale_id = ? 
        AND COALESCE(R.total_returned, 0) < A.qty
    `, saleId, saleId).Scan(&results).Error

	if err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal mengambil item penjualan", err.Error())
	}

	if len(results) == 0 {
		return responses.JSONResponse(c, http.StatusNotFound, "Tidak ada item yang bisa diretur untuk penjualan ini", nil)
	}

	return responses.JSONResponse(c, http.StatusOK, "Data item retur ditemukan", results)
}

// GetSaleReturnWithItems menampilkan satu retur penjualan beserta semua item-nya
func GetSaleReturnWithItems(c *framework.Ctx) error {
	db := config.DB

	saleReturnID := c.Param("id")

	// Gunakan models.AllSaleReturns untuk mengambil data dari DB
	var saleReturn models.AllSaleReturns

	err := db.Table("sale_returns A").
		Select("A.id, A.sale_id, A.return_date, A.payment, A.total_return").
		Where("A.id = ?", saleReturnID).
		Scan(&saleReturn).Error

	if err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal mengambil data retur penjualan", err.Error())
	}

	// Ambil item retur penjualan terkait
	var items []models.AllSaleReturnItems
	err = db.Table("sale_return_items A").
		Select("A.id, A.sale_return_id, A.product_id AS pro_id, B.name AS pro_name, B.unit_id, C.name AS unit_name, A.qty, A.price, A.sub_total, A.expired_date").
		Joins("LEFT JOIN products B on B.id=A.product_id").
		Joins("LEFT JOIN units C on C.id=B.unit_id").
		Where("A.sale_return_id = ?", saleReturnID).
		Order("B.name ASC").
		Scan(&items).Error

	if err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal mengambil item retur penjualan", err.Error())
	}

	// Format tanggal secara manual untuk respons ini
	formattedSaleReturnDate := utils.FormatIndonesianDate(saleReturn.ReturnDate)

	// Buat objek respons menggunakan struct SaleItemResponse yang baru
	responseDetail := models.SaleReturnItemResponse{
		ID:          saleReturn.ID,
		SaleId:      saleReturn.SaleId,
		ReturnDate:  formattedSaleReturnDate,
		TotalReturn: saleReturn.TotalReturn,
		Payment:     string(saleReturn.Payment),
		Items:       items,
	}

	// Panggil JSONResponse yang sudah ada, meneruskan SaleItemResponse sebagai 'data'
	return responses.JSONResponse(c, http.StatusOK, "Retur penjualan berhasil diambil", responseDetail)
}

// GetAllSaleReturns menampilkan semua retur penjualan
func GetAllSaleReturns(c *framework.Ctx) error {
	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(utils.Location)

	branchID, _ := middlewares.GetBranchID(c.Request)

	// Ambil parameter page dan search dari query URL
	pageParam := c.Query("page")
	search := strings.TrimSpace(c.Query("search"))

	// Konversi page ke int, default ke 1 jika tidak valid
	page := 1
	if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
		page = p
	}

	limit := 10                  // Tetapkan limit ke 10 data per halaman
	offset := (page - 1) * limit // Hitung offset berdasarkan halaman dan limit

	month := strings.TrimSpace(c.Query("month"))

	// Jika month kosong, isi dengan bulan ini (format YYYY-MM)
	if month == "" {
		month = nowWIB.Format("2006-01")
	}

	var saleReturnsFromDB []models.AllSaleReturns // Gunakan models.AllSaleReturns untuk mengambil data dari DB
	var total int64

	query := config.DB.Table("sale_returns A").
		Select("A.id, A.sale_id, A.return_date, A.payment, A.total_return").
		Where("A.branch_id = ? ", branchID).
		Order("A.created_at DESC")

	if search != "" {
		search = strings.ToLower(search)
		query = query.Where("LOWER(A.sale_id) LIKE ?", "%"+search+"%")
	}

	if month != "" {
		parsedMonth, err := time.Parse("2006-01", month)
		if err != nil {
			return responses.JSONResponse(c, http.StatusBadRequest, "Format bulan tidak valid", "Bulan harus dalam format YYYY-MM")
		}
		startDate := parsedMonth
		endDate := startDate.AddDate(0, 1, 0).Add(-time.Nanosecond)
		query = query.Where("A.return_date BETWEEN ? AND ?", startDate, endDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal mengambil retur pembelian", "Gagal menghitung retur pembelian")
	}

	if err := query.Offset(offset).Limit(limit).Scan(&saleReturnsFromDB).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal mengambil retur penjualan", "Gagal mengambil data retur penjualan")
	}

	// Buat slice baru untuk menampung data yang sudah diformat
	var formattedSaleReturnsData []models.SaleReturnsResponse
	for _, saleReturn := range saleReturnsFromDB {
		formattedSaleReturnsData = append(formattedSaleReturnsData, models.SaleReturnsResponse{
			ID:          saleReturn.ID,
			SaleId:      saleReturn.SaleId,
			ReturnDate:  utils.FormatIndonesianDate(saleReturn.ReturnDate), // Format tanggal di sini
			TotalReturn: saleReturn.TotalReturn,
			Payment:     string(saleReturn.Payment),
		})
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	// Gunakan JSONResponseGetAll helper dengan data yang sudah diformat
	return responses.JSONResponseGetAll(
		c,
		http.StatusOK,
		"Data retur pembelian berhasil diambil",
		search,
		int(total),
		page,
		totalPages,
		limit,
		formattedSaleReturnsData, // Kirim data yang sudah diformat (slice dari SaleReturnsResponse)
	)
}

// CmbSale mengambil data penjualan
func CmbSale(c *framework.Ctx) error {
	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(utils.Location)

	branchID, _ := middlewares.GetBranchID(c.Request)

	// Ambil parameter page dan search dari query URL
	search := strings.TrimSpace(c.Query("search"))

	month := strings.TrimSpace(c.Query("month"))

	// Jika month kosong, isi dengan bulan ini (format YYYY-MM)
	if month == "" {
		month = nowWIB.Format("2006-01")
	}

	var sales []models.Sales

	query := config.DB.Table("sales").
		Where("branch_id = ?", branchID)

	// Filter by month (sale_date)
	if month != "" {
		parsedMonth, err := time.Parse("2006-01", month)
		if err != nil {
			return responses.JSONResponse(c, http.StatusBadRequest, "Format bulan tidak valid", "Bulan harus dalam format YYYY-MM")
		}
		startDate := parsedMonth
		endDate := startDate.AddDate(0, 1, 0).Add(-time.Nanosecond)
		query = query.Where("sale_date BETWEEN ? AND ?", startDate, endDate)
	}

	// Optional search by sales.id
	if search != "" {
		search = strings.ToLower(search)
		query = query.Where("LOWER(id) LIKE ?", "%"+search+"%")
	}

	query = query.Order("sale_date DESC")

	if err := query.Find(&sales).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal mengambil data penjualan", err.Error())
	}

	return responses.JSONResponse(
		c,
		http.StatusOK,
		"Data penjualan berhasil diambil",
		sales,
	)
}
