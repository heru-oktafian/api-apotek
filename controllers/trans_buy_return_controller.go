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

// CreateBuyReturnTransaction adalah fungsi untuk membuat transaksi retur pembelian baru
func CreateBuyReturnTransaction(c *framework.Ctx) error {
	nowWIB := time.Now().In(utils.Location)

	subscriptionType, _ := middlewares.GetClaimsToken(c.Request, "subscription_type")
	branchID, _ := middlewares.GetClaimsToken(c.Request, "branch_id")
	userID, _ := middlewares.GetClaimsToken(c.Request, "user_id")

	db := config.DB
	var req models.BuyReturnRequest // pastikan struct request ini ada
	err := c.BodyParser(&req)
	if err != nil {
		return responses.JSONResponse(c, http.StatusBadRequest, "Body permintaan tidak valid", err.Error())
	}

	if req.BuyReturn.Payment == "" {
		req.BuyReturn.Payment = "paid_by_cash"
	}

	req.BuyReturn.UserID = userID
	req.BuyReturn.BranchID = branchID

	err = utils.ValidateStruct(req.BuyReturn)
	if err != nil {
		return responses.JSONResponse(c, http.StatusBadRequest, "Validasi input retur pembelian gagal", err.Error())
	}

	for _, item := range req.BuyReturnItems {
		err = utils.ValidateStruct(item)
		if err != nil {
			return responses.JSONResponse(c, http.StatusBadRequest, "Validasi salah satu item retur pembelian gagal", err.Error())
		}
	}

	var returnDate time.Time
	if req.BuyReturn.ReturnDate == "" {
		returnDate = nowWIB
	} else {
		returnDate, err = time.Parse("2006-01-02", req.BuyReturn.ReturnDate)
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

	// Validasi apakah purchase_id valid
	var buy models.Purchases
	err = tx.Where("id = ?", req.BuyReturn.PurchaseId).First(&buy).Error
	if err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return responses.JSONResponse(c, http.StatusNotFound, fmt.Sprintf("Pembelian dengan ID %s tidak ditemukan", req.BuyReturn.PurchaseId), err.Error())
		}
		return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal mengambil data pembelian", err.Error())
	}

	buyReturnID := helpers.GenerateID("BRT")
	buyReturn := models.BuyReturns{
		ID:         buyReturnID,
		PurchaseId: req.BuyReturn.PurchaseId,
		ReturnDate: returnDate,
		BranchID:   branchID,
		Payment:    req.BuyReturn.Payment,
		UserID:     userID,
		CreatedAt:  nowWIB,
		UpdatedAt:  nowWIB,
	}

	var totalReturn int
	var buyReturnItems []models.BuyReturnItems

	for _, item := range req.BuyReturnItems {
		parsedExpiredDate, err := time.Parse("2006-01-02", item.ExpiredDate)
		if err != nil {
			tx.Rollback()
			return responses.JSONResponse(c, http.StatusBadRequest, fmt.Sprintf("expired_date tidak valid untuk produk %s", item.ProductId), err.Error())
		}

		// Validasi item berasal dari purchase_id
		// Ambil item pembelian untuk purchase_id + product_id
		var buyItem models.PurchaseItems
		err = tx.Where("purchase_id = ? AND product_id = ?", req.BuyReturn.PurchaseId, item.ProductId).First(&buyItem).Error
		if err != nil {
			tx.Rollback()
			return responses.JSONResponse(c, http.StatusBadRequest, fmt.Sprintf("Produk %s tidak ditemukan pada pembelian asal", item.ProductId), err.Error())
		}

		// Ambil total qty yang sudah diretur sebelumnya
		var totalReturnedQty int64
		err = tx.Model(&models.BuyReturnItems{}).
			Select("COALESCE(SUM(qty), 0)").
			Where("product_id = ? AND buy_return_id IN (SELECT id FROM buy_returns WHERE purchase_id = ?)", item.ProductId, req.BuyReturn.PurchaseId).
			Scan(&totalReturnedQty).Error

		if err != nil {
			tx.Rollback()
			return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal memeriksa retur sebelumnya", err.Error())
		}

		// Validasi jika qty retur melebihi qty pembelian
		if int(totalReturnedQty)+item.Qty > buyItem.Qty {
			tx.Rollback()
			return responses.JSONResponse(c, http.StatusBadRequest, fmt.Sprintf("Total qty retur untuk produk %s melebihi jumlah yang dibeli. Dibeli: %d, Sudah Diretur: %d, Retur Ini: %d",
				item.ProductId, buyItem.Qty, totalReturnedQty, item.Qty), nil)
		}

		// Ambil informasi produk
		var product models.Product
		err = tx.Where("id = ?", item.ProductId).First(&product).Error
		if err != nil {
			tx.Rollback()
			return responses.JSONResponse(c, http.StatusInternalServerError, fmt.Sprintf("Gagal mengambil info produk untuk %s", item.ProductId), err.Error())
		}

		actualQtyToReduce := item.Qty

		// Lakukan konversi unit jika diperlukan
		if buyItem.UnitId != product.UnitId {
			var unitConv models.UnitConversion
			err = tx.Where("product_id = ? AND init_id = ? AND final_id = ? AND branch_id = ?",
				buyItem.ProductId, buyItem.UnitId, product.UnitId, branchID).First(&unitConv).Error

			if err != nil {
				if err == gorm.ErrRecordNotFound {
					actualQtyToReduce = item.Qty
				} else {
					tx.Rollback()
					return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal mengambil konversi satuan", err.Error())
				}
			} else {
				actualQtyToReduce = item.Qty * unitConv.ValueConv
			}
		}

		// Update stok
		err = tx.Model(&models.Product{}).Where("id = ?", item.ProductId).
			Update("stock", gorm.Expr("stock - ?", actualQtyToReduce)).Error
		if err != nil {
			tx.Rollback()
			return responses.JSONResponse(c, http.StatusInternalServerError, fmt.Sprintf("Gagal memperbarui stok untuk produk %s", item.ProductId), err.Error())
		}

		subTotal := buyItem.Price * item.Qty
		totalReturn += subTotal

		buyReturnItems = append(buyReturnItems, models.BuyReturnItems{
			ID:          helpers.GenerateID("BRI"),
			BuyReturnId: buyReturnID,
			ProductId:   item.ProductId,
			Price:       buyItem.Price,
			Qty:         item.Qty,
			SubTotal:    subTotal,
			ExpiredDate: parsedExpiredDate,
		})

	}

	buyReturn.TotalReturn = totalReturn

	err = tx.Create(&buyReturn).Error
	if err != nil {
		tx.Rollback()
		return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal membuat retur pembelian", err.Error())
	}

	err = tx.CreateInBatches(&buyReturnItems, len(buyReturnItems)).Error
	if err != nil {
		tx.Rollback()
		return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal membuat item retur pembelian", err.Error())
	}

	// Tambahkan ke laporan transaksi
	transactionReportID := helpers.GenerateID("TRX")
	transactionReport := models.TransactionReports{
		ID:              transactionReportID,
		TransactionType: models.BuyReturn,
		UserID:          userID,
		BranchID:        branchID,
		Total:           buyReturn.TotalReturn,
		Payment:         buyReturn.Payment,
		CreatedAt:       nowWIB,
		UpdatedAt:       nowWIB,
	}
	err = tx.Create(&transactionReport).Error
	if err != nil {
		tx.Rollback()
		return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal membuat laporan transaksi retur pembelian", err.Error())
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
			return responses.JSONResponse(c, http.StatusBadRequest, "Kuota cabang sudah habis", "Silakan tingkatkan paket berlangganan Anda")
		}
	}

	err = tx.Commit().Error
	if err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal melakukan commit transaksi", err.Error())
	}

	return responses.JSONResponse(c, http.StatusOK, "Transaksi retur pembelian berhasil dibuat", framework.Map{
		"id":           buyReturn.ID,
		"purchase_id":  buyReturn.PurchaseId,
		"return_date":  utils.FormatIndonesianDate(buyReturn.ReturnDate),
		"total_return": buyReturn.TotalReturn,
		"payment":      buyReturn.Payment,
		"items":        buyReturnItems,
	})
}

// GetBuyItemsForReturn digunakan untuk mengambil item pembelian yang bisa diretur
func GetBuyItemsForReturn(c *framework.Ctx) error {
	purchaseId := c.Query("purchase_id")
	if purchaseId == "" {
		return responses.JSONResponse(c, http.StatusBadRequest, "purchase_id wajib diisi", nil)
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
        FROM purchase_items A
        LEFT JOIN products B ON B.id = A.product_id
        LEFT JOIN units C ON C.id = B.unit_id
        LEFT JOIN (
            SELECT 
                sri.product_id, 
                SUM(sri.qty) AS total_returned
            FROM buy_return_items sri
            INNER JOIN buy_returns sr ON sri.buy_return_id = sr.id
            WHERE sr.purchase_id = ?
            GROUP BY sri.product_id
        ) R ON R.product_id = A.product_id
        WHERE A.purchase_id = ? 
        AND COALESCE(R.total_returned, 0) < A.qty
    `, purchaseId, purchaseId).Scan(&results).Error

	if err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal mengambil item pembelian", err.Error())
	}

	if len(results) == 0 {
		return responses.JSONResponse(c, http.StatusOK, "Tidak ada item yang bisa diretur untuk pembelian ini", results)
	}

	return responses.JSONResponse(c, http.StatusOK, "Data item retur ditemukan", results)
}

// GetBuyReturnWithItems menampilkan satu retur pembelian beserta semua item-nya
func GetBuyReturnWithItems(c *framework.Ctx) error {
	db := config.DB

	buyReturnID := c.Param("id")

	// Gunakan models.AllBuyReturns untuk mengambil data dari DB
	var buyReturn models.AllBuyReturns

	err := db.Table("buy_returns A").
		Select("A.id, A.purchase_id, A.return_date, A.payment, A.total_return").
		Where("A.id = ?", buyReturnID).
		Scan(&buyReturn).Error

	if err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal mengambil data retur pembelian", err.Error())
	}

	// Ambil item retur pembelian terkait
	var items []models.AllBuyReturnItems
	err = db.Table("buy_return_items A").
		Select("A.id, A.buy_return_id, A.product_id AS pro_id, B.name AS pro_name, B.unit_id, C.name AS unit_name, A.qty, A.price, A.sub_total, A.expired_date").
		Joins("LEFT JOIN products B on B.id=A.product_id").
		Joins("LEFT JOIN units C on C.id=B.unit_id").
		Where("A.buy_return_id = ?", buyReturnID).
		Order("B.name ASC").
		Scan(&items).Error

	if err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal mengambil item retur pembelian", err.Error())
	}

	// Format tanggal secara manual untuk respons ini
	formattedBuyReturnDate := utils.FormatIndonesianDate(buyReturn.ReturnDate)

	// Buat objek respons menggunakan struct BuyItemResponse yang baru
	responseDetail := models.BuyReturnItemResponse{
		ID:          buyReturn.ID,
		PurchaseId:  buyReturn.PurchaseId,
		ReturnDate:  formattedBuyReturnDate,
		TotalReturn: buyReturn.TotalReturn,
		Payment:     string(buyReturn.Payment),
		Items:       items,
	}

	// Panggil JSONResponse yang sudah ada, meneruskan BuyItemResponse sebagai 'data'
	return responses.JSONResponse(c, http.StatusOK, "Retur pembelian berhasil diambil", responseDetail)
}

// GetAllBuyReturns menampilkan semua retur pembelian
func GetAllBuyReturns(c *framework.Ctx) error {
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

	var buyReturnsFromDB []models.AllBuyReturns // Gunakan models.AllBuyReturns untuk mengambil data dari DB
	var total int64

	query := config.DB.Table("buy_returns A").
		Select("A.id, A.purchase_id, A.return_date, A.payment, A.total_return").
		Where("A.branch_id = ? ", branchID).
		Order("A.created_at DESC")

	if search != "" {
		search = strings.ToLower(search)
		query = query.Where("LOWER(A.purchase_id) LIKE ?", "%"+search+"%")
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

	if err := query.Offset(offset).Limit(limit).Scan(&buyReturnsFromDB).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal mengambil retur pembelian", "Gagal mengambil data retur pembelian")
	}

	// Buat slice baru untuk menampung data yang sudah diformat
	var formattedBuyReturnsData []models.BuyReturnsResponse
	for _, buyReturn := range buyReturnsFromDB {
		formattedBuyReturnsData = append(formattedBuyReturnsData, models.BuyReturnsResponse{
			ID:          buyReturn.ID,
			PurchaseId:  buyReturn.PurchaseId,
			ReturnDate:  utils.FormatIndonesianDate(buyReturn.ReturnDate), // Format tanggal di sini
			TotalReturn: buyReturn.TotalReturn,
			Payment:     string(buyReturn.Payment),
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
		formattedBuyReturnsData, // Kirim data yang sudah diformat (slice dari BuyDetailResponse)
	)
}

// CmbPurchase mengambil data pembelian
func CmbPurchase(c *framework.Ctx) error {
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

	var purchases []models.Purchases

	query := config.DB.Table("purchases").
		Where("branch_id = ?", branchID)

	// Filter by month (purchase_date)
	if month != "" {
		parsedMonth, err := time.Parse("2006-01", month)
		if err != nil {
			return responses.JSONResponse(c, http.StatusBadRequest, "Format bulan tidak valid", "Bulan harus dalam format YYYY-MM")
		}
		startDate := parsedMonth
		endDate := startDate.AddDate(0, 1, 0).Add(-time.Nanosecond)
		query = query.Where("purchase_date BETWEEN ? AND ?", startDate, endDate)
	}

	// Optional search by purchases.id
	if search != "" {
		search = strings.ToLower(search)
		query = query.Where("LOWER(id) LIKE ?", "%"+search+"%")
	}

	query = query.Order("purchase_date DESC")

	if err := query.Find(&purchases).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal mengambil data pembelian", err.Error())
	}

	return responses.JSONResponse(
		c,
		http.StatusOK,
		"Data pembelian berhasil diambil",
		purchases,
	)
}
