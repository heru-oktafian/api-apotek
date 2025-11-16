package controllers

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/heru-oktafian/api-retail/models"
	"github.com/heru-oktafian/api-retail/reports"
	"github.com/heru-oktafian/api-retail/tools"
	"github.com/heru-oktafian/scafold/config"
	"github.com/heru-oktafian/scafold/framework"
	"github.com/heru-oktafian/scafold/helpers"
	"github.com/heru-oktafian/scafold/middlewares"
	"github.com/heru-oktafian/scafold/responses"
	"github.com/heru-oktafian/scafold/utils"
	"gorm.io/gorm"
)

// CreateSaleTransaction controller
func CreateSaleTransaction(c *framework.Ctx) error {
	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(utils.Location)

	db := config.DB
	var req SaleTransactionRequest
	// Deklarasi 'err' pertama kali di sini
	err := c.BodyParser(&req)
	if err != nil {
		return responses.BadRequest(c, "Invalid request body", err)
	}

	// Get default_member id dari token
	defaultMember, _ := middlewares.GetClaimsToken(c.Request, "default_member")

	// Get subscription_type dari token
	subscriptionType, _ := middlewares.GetClaimsToken(c.Request, "subscription_type")

	//Get BranchID from token
	branchID, _ := middlewares.GetBranchID(c.Request)

	// Get UserID from token
	userID, _ := middlewares.GetUserID(c.Request)

	// --- VALIDASI INPUT ---
	// Menggunakan 'err =' karena 'err' sudah dideklarasikan di atas
	err = utils.ValidateStruct(req)
	if err != nil {
		return responses.BadRequest(c, "Validate failed", err)
	}
	// --- AKHIR VALIDASI INPUT ---

	// Modifikasi agar jika `member_id` tidak dikirim dalam request,
	// maka `member_id` diisi `defaultMember` dari deklarasi tersebut.
	if req.Sale.MemberId == "" {
		req.Sale.MemberId = defaultMember
	}

	if req.Sale.Payment == "" {
		req.Sale.Payment = "paid_by_cash"
	}

	// --- Proses Penyimpanan Data ---
	// Mulai transaksi database
	tx := db.Begin()
	if tx.Error != nil {
		return responses.InternalServerError(c, "Failed to begin database transaction", err)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Simpan data Sales (induk)
	saleID := helpers.GenerateID("SAL")
	req.Sale.ID = saleID
	req.Sale.SaleDate = nowWIB
	req.Sale.UserID = userID
	req.Sale.BranchID = branchID
	req.Sale.CreatedAt = nowWIB
	req.Sale.UpdatedAt = nowWIB

	// Inisialisasi total_sale dan profit_estimate untuk kalkulasi
	var calculatedTotalSale int
	var calculatedProfitEstimate int

	// 2. Simpan data SaleItems (anak-anak) dan Update Stok
	// var stockTracksToCreate []models.StockTracks

	for i := range req.SaleItems {
		itemID := helpers.GenerateID("SIT") // Generate ID untuk setiap SaleItem
		req.SaleItems[i].ID = itemID
		req.SaleItems[i].SaleId = saleID // Kaitkan dengan Sale ID yang baru dibuat

		// Dapatkan detail produk untuk stok dan perhitungan profit
		var product models.Product
		err = tx.Where("id = ?", req.SaleItems[i].ProductId).First(&product).Error
		if err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				return responses.NotFound(c, "Product with ID %s not found")
			}

			return responses.InternalServerError(c, "Failed to retrieve product details", err)
		}

		// Periksa ketersediaan stok
		if product.Stock < req.SaleItems[i].Qty {
			tx.Rollback()
			return responses.BadRequest(c, fmt.Sprintf("Insufficient stock for product %s. Available: %d, Requested: %d", product.Name, product.Stock, req.SaleItems[i].Qty), err)
		}

		// Kurangi stok produk
		newStock := product.Stock - req.SaleItems[i].Qty
		err = tx.Model(&models.Product{}).Where("id = ?", product.ID).Update("stock", newStock).Error
		if err != nil {
			tx.Rollback()
			return responses.InternalServerError(c, fmt.Sprintf("Failed to update stock for product %s", product.Name), err)
		}

		// Kalkulasi total_sale dan profit_estimate dari item_sales
		calculatedTotalSale += req.SaleItems[i].SubTotal
		// Profit per item = (Harga Jual - Harga Beli) * Qty
		calculatedProfitEstimate += (req.SaleItems[i].Price - product.PurchasePrice) * req.SaleItems[i].Qty
	}

	// Set nilai total_sale dan profit_estimate pada struct Sales
	req.Sale.TotalSale = calculatedTotalSale - req.Sale.Discount // Kurangi profit dengan diskon keseluruhan
	req.Sale.ProfitEstimate = calculatedProfitEstimate

	// Simpan data Sales setelah kalkulasi total dan profit
	err = tx.Create(&req.Sale).Error
	if err != nil {
		tx.Rollback()
		return responses.InternalServerError(c, "Failed to create sale", err)
	}

	// Simpan SaleItems dalam batch
	err = tx.CreateInBatches(&req.SaleItems, len(req.SaleItems)).Error
	if err != nil {
		tx.Rollback()
		return responses.InternalServerError(c, "Failed to create sale items", err)
	}

	// 3. Simpan data di TransactionReports
	transactionReportID := helpers.GenerateID("TRX")
	transactionReport := models.TransactionReports{
		ID:              transactionReportID,
		TransactionType: models.Sale, // Tipe transaksi adalah "sale"
		UserID:          req.Sale.UserID,
		BranchID:        req.Sale.BranchID,
		Total:           req.Sale.TotalSale - req.Sale.Discount,
		Payment:         req.Sale.Payment,
		CreatedAt:       nowWIB,
		UpdatedAt:       nowWIB,
	}
	err = tx.Create(&transactionReport).Error
	if err != nil {
		tx.Rollback()
		return responses.InternalServerError(c, "Failed to create transaction report", err)
	}

	// 4. Update/Simpan data di DailyProfitReport
	var dailyProfit models.DailyProfitReport
	// Pastikan SaleDate tidak nol saat diakses (validasi required sudah ada, tapi jaga-jaga)
	if req.Sale.SaleDate.IsZero() {
		tx.Rollback()
		return responses.BadRequest(c, "SaleDate cannot be zero for daily profit report calculation. Please provide a valid date.", nil)
	}

	reportDate := req.Sale.SaleDate.Format("2006-01-02") // Format tanggal menjadi "YYYY-MM-DD"
	err = tx.Where("report_date = ? AND branch_id = ? AND user_id = ?", reportDate, req.Sale.BranchID, req.Sale.UserID).First(&dailyProfit).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		tx.Rollback()
		return responses.InternalServerError(c, "Failed to check daily profit report", err)
	}

	if err == gorm.ErrRecordNotFound {
		// Jika belum ada, buat entri baru
		dailyProfitID := helpers.GenerateID("DPR")
		dailyProfit = models.DailyProfitReport{
			ID:             dailyProfitID,
			ReportDate:     req.Sale.SaleDate,
			UserID:         req.Sale.UserID,
			BranchID:       req.Sale.BranchID,
			TotalSales:     req.Sale.TotalSale,
			ProfitEstimate: req.Sale.ProfitEstimate,
			CreatedAt:      nowWIB,
			UpdatedAt:      nowWIB,
		}
		err = tx.Create(&dailyProfit).Error
		if err != nil {
			tx.Rollback()
			return responses.InternalServerError(c, "Failed to create daily profit report", err)
		}
	} else {
		// Jika sudah ada, update total_sales dan profit_estimate
		dailyProfit.TotalSales += req.Sale.TotalSale
		dailyProfit.ProfitEstimate += req.Sale.ProfitEstimate
		dailyProfit.UpdatedAt = time.Now()
		err = tx.Save(&dailyProfit).Error
		if err != nil {
			tx.Rollback()
			return responses.InternalServerError(c, "Failed to update daily profit report", err)
		}
	}

	// b. Cek `subscription_type` jika type nya adalah `quota`
	// maka setiap transaksi Sale tersebut akan mengurangi 1 jumlah pada kolom `quota` yang ada di tabel `branches`.
	if subscriptionType == "quota" {
		var branch models.Branch
		err = tx.Where("id = ?", req.Sale.BranchID).First(&branch).Error
		if err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				return responses.NotFound(c, fmt.Sprintf("Branch with ID %s not found", req.Sale.BranchID))
			}
			return responses.InternalServerError(c, "Failed to retrieve branch details for quota update", err)
		}

		if branch.Quota > 0 {
			branch.Quota -= 1
			err = tx.Save(&branch).Error
			if err != nil {
				tx.Rollback()
				// return c.Status(framework.StatusInternalServerError).JSON(framework.Map{
				// 	"message": fmt.Sprintf("Failed to update quota for branch %s", branch.BranchName),
				// 	"error":   err.Error(),
				// })
				return responses.InternalServerError(c, fmt.Sprintf("Failed to update quota for branch %s", branch.BranchName), err)
			}
		} else {
			tx.Rollback()
			return responses.BadRequest(c, fmt.Sprintf("No quota available for branch %s", branch.BranchName), nil)
		}
	}

	// c. Cek jika `member_id` diisi tidak sama dengan `defaultMember` yang kita ambil dari klaim token tersebut,
	// maka akan mengecek `points_conversion_rate` yang ada di tabel `member_categories`
	// dengan acuan `member_id` yang dimasukan tersebut.
	// Kemudian melakukan perhitungan (`total_sale` : `points_conversion_rate`) = x
	// kemudian menambahkan x tersebut di kolom `points` di tabel `members`
	if req.Sale.MemberId != "" && req.Sale.MemberId != defaultMember {
		var member models.Member
		err = tx.Where("id = ?", req.Sale.MemberId).First(&member).Error
		if err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				return responses.NotFound(c, fmt.Sprintf("Member with ID %s not found", req.Sale.MemberId))
			}
			return responses.InternalServerError(c, "Failed to retrieve member details for points calculation", err)
		}

		var memberCategory models.MemberCategory
		err = tx.Where("id = ?", member.MemberCategoryId).First(&memberCategory).Error
		if err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				return responses.NotFound(c, fmt.Sprintf("Member category with ID %d not found for member %s", member.MemberCategoryId, member.ID))
			}
			return responses.InternalServerError(c, "Failed to retrieve member category for points calculation", err)
		}

		if memberCategory.PointsConversionRate > 0 {
			// Pastikan total_sale adalah float untuk perhitungan poin
			pointsEarned := float64(req.Sale.TotalSale) / float64(memberCategory.PointsConversionRate)
			member.Points += int(pointsEarned) // Tambahkan poin yang didapat (gunakan int jika kolom points int)

			err = tx.Save(&member).Error
			if err != nil {
				tx.Rollback()
				return responses.InternalServerError(c, fmt.Sprintf("Failed to update points for member %s", member.ID), err)
			}
		} else {
			// Optional: Handle case where PointsConversionRate is 0 or less
			// You might want to log this or return a specific error
			fmt.Printf("Warning: PointsConversionRate for member category %d is zero or negative. Points not calculated.\n", member.MemberCategoryId)
		}
	}

	// Commit transaksi jika semua berhasil
	err = tx.Commit().Error
	if err != nil {
		return responses.InternalServerError(c, "Failed to commit database transaction", err)
	}

	// Berhasil
	return responses.JSONResponse(c, http.StatusOK, "Sale transaction created successfully", req)
}

// UpdateSale Function (Modified)
func UpdateSale(c *framework.Ctx) error {

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(utils.Location)

	db := config.DB
	id := c.Param("id")

	// branchId, _ := middlewares.GetBranchID(c.Request)

	var sale models.Sales
	if err := db.First(&sale, "id = ?", id).Error; err != nil {
		return responses.NotFound(c, "Sale not found")
	}

	var input models.SaleInput
	if err := c.BodyParser(&input); err != nil {
		return responses.BadRequest(c, "Invalid input", err)
	}

	if input.MemberId != nil {
		var member models.Member
		if err := db.Where("id = ?", *input.MemberId).First(&member).Error; err != nil {
			// Jika ID tidak valid, fallback ke default
			memberId, _ := middlewares.GetClaimsToken(c.Request, "default_member")
			sale.MemberId = memberId
		} else {
			sale.MemberId = *input.MemberId
		}
	}
	// Jika nil → tidak diubah, tetap pakai MemberID yang sudah ada

	if input.Payment != "" {
		sale.Payment = models.PaymentStatus(input.Payment)
	}

	sale.UpdatedAt = nowWIB

	var items []models.SaleItems
	if err := db.Where("sale_id = ?", id).Find(&items).Error; err != nil {
		return responses.InternalServerError(c, "Failed to fetch sale items", err)
	}

	total := 0
	for _, item := range items {
		total += item.SubTotal
	}

	// Gunakan diskon baru jika dikirim, jika tidak tetap pakai yang lama
	if input.Discount != nil {
		sale.Discount = *input.Discount
	}
	sale.TotalSale = total - sale.Discount

	if err := db.Save(&sale).Error; err != nil {
		return responses.InternalServerError(c, "Failed to update sale", err)
	}

	if err := reports.SyncSaleReport(db, sale); err != nil {
		return responses.InternalServerError(c, "Failed to sync sale report", err)
	}

	_ = reports.AutoCleanupSales(db)
	_ = reports.SyncDailyProfitReport(db, sale)

	return responses.JSONResponse(c, http.StatusOK, "Sale updated successfully", sale)
}

// DeleteSale Function
func DeleteSale(c *framework.Ctx) error {
	db := config.DB
	id := c.Param("id")

	// Ambil sale
	var sale models.Sales
	if err := db.First(&sale, "id = ?", id).Error; err != nil {
		return responses.NotFound(c, "Sale not found")
	}

	// Ambil & hapus item, serta rollback stok
	var items []models.SaleItems
	if err := db.Where("sale_id = ?", id).Find(&items).Error; err == nil {
		for _, item := range items {
			_ = tools.SubtractProductStock(db, item.ProductId, item.Qty)
		}
		db.Where("sale_id = ?", id).Delete(&models.SaleItems{})
	}

	// Hapus laporan transaksi
	if err := db.Where("id = ? AND transaction_type = ?", sale.ID, models.Sale).Delete(&models.TransactionReports{}).Error; err != nil {
		return responses.InternalServerError(c, "Failed to delete transaction report", err)
	}

	// Hapus data penjualan
	if err := db.Delete(&sale).Error; err != nil {
		return responses.InternalServerError(c, "Failed to delete sale", err)
	}

	// Delete laporan profit harian
	_ = reports.DeleteDailyProfitReport(db, id)

	// (Opsional) Sync laporan penjualan agar tetap konsisten
	_ = reports.SyncSaleReport(db, sale)

	return responses.JSONResponse(c, http.StatusOK, "Sale deleted successfully", sale)
}

// CreateSaleItem Function
func CreateSaleItem(c *framework.Ctx) error {
	db := config.DB
	var item models.SaleItems

	if err := c.BodyParser(&item); err != nil {
		return responses.BadRequest(c, "Invalid input", err)
	}

	// Ambil harga jual produk dari tabel products
	var product models.Product
	if err := db.Select("sales_price").Where("id = ?", item.ProductId).First(&product).Error; err != nil {
		return responses.InternalServerError(c, "Failed to fetch product price", err)
	}

	// Gunakan sales_price dari produk, abaikan inputan frontend
	item.Price = product.SalesPrice

	// Cek apakah item dengan sale_id dan product_id sudah ada
	var existing models.SaleItems
	err := db.Where("sale_id = ? AND product_id = ?", item.SaleId, item.ProductId).First(&existing).Error
	if err == nil {
		// Sudah ada: update qty dan sub_total
		existing.Qty += item.Qty
		existing.Price = product.SalesPrice
		existing.SubTotal = existing.Qty * existing.Price

		if err := db.Save(&existing).Error; err != nil {
			return responses.InternalServerError(c, "Failed to update sale item", err)
		}

		if err := tools.ReduceProductStock(db, item.ProductId, item.Qty); err != nil {
			return responses.InternalServerError(c, "Failed to reduce product stock", err)
		}

		if err := reports.RecalculateTotalSale(db, item.SaleId); err != nil {
			return responses.InternalServerError(c, "Failed to recalculate total sale", err)
		}

		// Sync laporan profit harian
		var sale models.Sales
		if err := db.First(&sale, "id = ?", item.SaleId).Error; err != nil {
			return responses.InternalServerError(c, "Failed to fetch sale", err)
		}

		_ = reports.SyncDailyProfitReport(db, sale)

		return responses.JSONResponse(c, http.StatusOK, "Item updated successfully", existing)

	} else if err != gorm.ErrRecordNotFound {
		return responses.InternalServerError(c, "Failed to find existing sale item", err)
	}

	// Data belum ada, buat item baru
	if item.ID == "" {
		item.ID = helpers.GenerateID("SIT")
	}
	item.SubTotal = item.Qty * item.Price

	if err := db.Create(&item).Error; err != nil {
		return responses.InternalServerError(c, "Failed to create sale item", err)
	}

	if err := tools.ReduceProductStock(db, item.ProductId, item.Qty); err != nil {
		return responses.InternalServerError(c, "Failed to reduce product stock", err)
	}

	if err := reports.RecalculateTotalSale(db, item.SaleId); err != nil {
		return responses.InternalServerError(c, "Failed to recalculate total sale", err)
	}

	// Sync laporan profit harian
	var sale models.Sales
	if err := db.First(&sale, "id = ?", item.SaleId).Error; err != nil {
		return responses.InternalServerError(c, "Failed to fetch sale", err)
	}

	_ = reports.SyncDailyProfitReport(db, sale)

	return responses.JSONResponse(c, http.StatusOK, "Item added successfully", item)
}

// UpdateSaleItem
func UpdateSaleItem(c *framework.Ctx) error {
	db := config.DB
	id := c.Param("id")

	var existingItem models.SaleItems
	if err := db.First(&existingItem, "id = ?", id).Error; err != nil {
		return responses.NotFound(c, "Item not found")
	}

	// Parsing data baru dari body (hanya untuk ambil ProductId dan Qty baru)
	var updatedData struct {
		ProductId string `json:"product_id"`
		Qty       int    `json:"qty"`
	}
	if err := c.BodyParser(&updatedData); err != nil {
		return responses.BadRequest(c, "Invalid input", err)
	}

	// Rollback stok lama
	if err := tools.AddProductStock(db, existingItem.ProductId, existingItem.Qty); err != nil {
		return responses.InternalServerError(c, "Failed to add product stock", err)
	}

	// Ambil harga jual dari produk baru
	var product models.Product
	if err := db.Select("sales_price").Where("id = ?", updatedData.ProductId).First(&product).Error; err != nil {
		return responses.InternalServerError(c, "Failed to get product price", err)
	}

	// Kurangi stok baru
	if err := tools.ReduceProductStock(db, updatedData.ProductId, updatedData.Qty); err != nil {
		return responses.InternalServerError(c, "Failed to reduce product stock", err)
	}

	// Update item
	existingItem.ProductId = updatedData.ProductId
	existingItem.Qty = updatedData.Qty
	existingItem.Price = product.SalesPrice
	existingItem.SubTotal = product.SalesPrice * updatedData.Qty

	if err := db.Save(&existingItem).Error; err != nil {
		return responses.InternalServerError(c, "Failed to update sale item", err)
	}

	if err := reports.RecalculateTotalSale(db, existingItem.SaleId); err != nil {
		return responses.InternalServerError(c, "Failed to recalculate total sale", err)
	}

	// Sync laporan profit harian
	var sale models.Sales
	if err := db.First(&sale, "id = ?", existingItem.SaleId).Error; err != nil {
		return responses.InternalServerError(c, "Failed to fetch sale", err)
	}

	_ = reports.SyncDailyProfitReport(db, sale)

	return responses.JSONResponse(c, http.StatusOK, "Item updated successfully", existingItem)
}

// Delete SaleItem
func DeleteSaleItem(c *framework.Ctx) error {
	db := config.DB
	id := c.Param("id")

	var item models.SaleItems
	if err := db.First(&item, "id = ?", id).Error; err != nil {
		return responses.NotFound(c, "Item not found")
	}

	// Rollback stok
	if err := tools.AddProductStock(db, item.ProductId, item.Qty); err != nil {
		return responses.InternalServerError(c, "Failed to add product stock", err)
	}

	// Hapus item
	if err := db.Delete(&item).Error; err != nil {
		return responses.InternalServerError(c, "Failed to delete sale item", err)
	}

	// Recalculate total
	if err := reports.RecalculateTotalSale(db, item.SaleId); err != nil {
		return responses.InternalServerError(c, "Failed to recalculate total sale", err)
	}

	return responses.JSONResponse(c, http.StatusOK, "Item deleted successfully", item)
}

// GetAllSales tampilkan semua sale
func GetAllSales(c *framework.Ctx) error {
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

	var salesFromDB []models.AllSales // Gunakan models.AllSales untuk mengambil data dari DB
	var total int64

	query := config.DB.Table("sales sl").
		Select("sl.id, sl.member_id, mbr.name AS member_name, sl.sale_date, sl.total_sale, sl.discount, sl.profit_estimate, sl.payment").
		Joins("LEFT JOIN members mbr on mbr.id = sl.member_id").
		Where("sl.branch_id = ? AND sl.total_sale > 0", branchID).
		Order("sl.created_at DESC")

	if search != "" {
		search = strings.ToLower(search)
		query = query.Where("LOWER(mbr.name) LIKE ?", "%"+search+"%")
	}

	if month != "" {
		parsedMonth, err := time.Parse("2006-01", month)
		if err != nil {
			// return helpers.JSONResponse(c, framework.StatusBadRequest, "Invalid month format", "Month should be in format YYYY-MM")
			return responses.BadRequest(c, "Invalid month format. Month should be in format YYYY-MM", err)
		}
		startDate := parsedMonth
		endDate := startDate.AddDate(0, 1, 0).Add(-time.Nanosecond)
		query = query.Where("sl.sale_date BETWEEN ? AND ?", startDate, endDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return responses.InternalServerError(c, "Get sale failed", err)
	}

	if err := query.Offset(offset).Limit(limit).Scan(&salesFromDB).Error; err != nil {
		return responses.InternalServerError(c, "Get sales failed", err)
	}

	// Buat slice baru untuk menampung data yang sudah diformat
	var formattedSalesData []models.SaleDetailResponse
	for _, sale := range salesFromDB {
		formattedSalesData = append(formattedSalesData, models.SaleDetailResponse{
			ID:             sale.ID,
			MemberId:       sale.MemberId,
			MemberName:     sale.MemberName,
			SaleDate:       utils.FormatIndonesianDate(sale.SaleDate), // Format tanggal di sini
			TotalSale:      sale.TotalSale,
			Discount:       sale.Discount,
			ProfitEstimate: sale.ProfitEstimate,
			Payment:        string(sale.Payment),
		})
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	// Gunakan JSONResponseGetAll helper dengan data yang sudah diformat
	return responses.JSONResponseGetAll(
		c,
		http.StatusOK,
		"Sales retrieved successfully",
		search,
		int(total),
		page,
		totalPages,
		limit,
		formattedSalesData, // Kirim data yang sudah diformat (slice dari SaleDetailResponse)
	)
}

// GetAllSalesDetail menampilkan sales dengan kolom description yang
// berisi daftar nama item (dipisah koma) diikuti dengan sale_date (DD-MM-YYYY HH:MM)
// di mana waktu ditambah 7 jam sesuai permintaan.
func GetAllSalesDetail(c *framework.Ctx) error {
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

	limit := 10
	offset := (page - 1) * limit

	month := strings.TrimSpace(c.Query("month"))
	if month == "" {
		month = nowWIB.Format("2006-01")
	}

	// Struct sementara untuk menampung hasil query
	type saleSummary struct {
		ID        string
		TotalSale int
		Payment   string
		SaleDate  time.Time
	}

	var salesFromDB []saleSummary
	var total int64

	query := config.DB.Table("sales sl").
		Select("sl.id, sl.total_sale, sl.payment, sl.sale_date").
		Joins("LEFT JOIN members mbr on mbr.id = sl.member_id").
		Where("sl.branch_id = ? AND sl.total_sale > 0", branchID).
		Order("sl.created_at DESC")

	if search != "" {
		s := strings.ToLower(search)
		query = query.Where("LOWER(mbr.name) LIKE ?", "%"+s+"%")
	}

	if month != "" {
		parsedMonth, err := time.Parse("2006-01", month)
		if err != nil {
			return responses.BadRequest(c, "Invalid month format. Month should be in format YYYY-MM", err)
		}
		startDate := parsedMonth
		endDate := startDate.AddDate(0, 1, 0).Add(-time.Nanosecond)
		query = query.Where("sl.sale_date BETWEEN ? AND ?", startDate, endDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return responses.InternalServerError(c, "Get sale failed", err)
	}

	if err := query.Offset(offset).Limit(limit).Scan(&salesFromDB).Error; err != nil {
		return responses.InternalServerError(c, "Get sales failed", err)
	}

	// Bentuk response dengan description
	var formatted []map[string]interface{}
	for _, s := range salesFromDB {
		// Ambil nama item untuk sale ini
		var itemNames []string
		if err := config.DB.Table("sale_items sit").
			Select("pro.name").
			Joins("LEFT JOIN products pro ON pro.id = sit.product_id").
			Where("sit.sale_id = ?", s.ID).
			Order("pro.name ASC").
			Pluck("pro.name", &itemNames).Error; err != nil {
			return responses.InternalServerError(c, "Failed to get sale items", err)
		}

		// Gabungkan nama item, lalu tambahkan tanggal yang ditambah 7 jam
		descItems := strings.Join(itemNames, ", ")
		dateWith7 := s.SaleDate.Add(7 * time.Hour).Format("02-01-2006 15:04")
		var description string
		if descItems != "" {
			description = descItems + " ; " + dateWith7
		} else {
			description = dateWith7
		}

		formatted = append(formatted, map[string]interface{}{
			"id":          s.ID,
			"total_sale":  s.TotalSale,
			"payment":     s.Payment,
			"description": description,
		})
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return responses.JSONResponseGetAll(
		c,
		http.StatusOK,
		"Sales retrieved successfully",
		search,
		int(total),
		page,
		totalPages,
		limit,
		formatted,
	)
}

// GetAllSaleItems tampilkan semua item berdasarkan sale_id tanpa pagination
func GetAllSaleItems(c *framework.Ctx) error {
	// Get sale id dari param
	saleID := c.Param("id")

	// Parsing body JSON ke struct
	var SaleItems []models.AllSaleItems

	// Query dasar
	query := config.DB.Table("sale_items sit").
		Select("sit.id, sit.sale_id, sit.product_id, pro.name AS product_name, sit.price, sit.qty, un.name AS unit_name, sit.sub_total").
		Joins("LEFT JOIN products pro ON pro.id = sit.product_id").
		Joins("LEFT JOIN units un ON un.id = pro.unit_id").
		Where("sit.sale_id = ?", saleID).
		Order("pro.name ASC")

	// Eksekusi query
	if err := query.Scan(&SaleItems).Error; err != nil {
		return responses.InternalServerError(c, "Get items failed", err)
	}

	return responses.JSONResponse(c, http.StatusOK, "Items retrieved successfully", SaleItems)
}

// GetSaleWithItems menampilkan satu sale beserta semua item-nya
func GetSaleWithItems(c *framework.Ctx) error {
	db := config.DB

	saleID := c.Param("id")

	// Gunakan models.AllSales untuk mengambil data dari DB
	var sale models.AllSales

	err := db.Table("sales sl").
		Select("sl.id, sl.member_id, mbr.name AS member_name, sl.sale_date, sl.discount, sl.total_sale, sl.profit_estimate, sl.payment").
		Joins("LEFT JOIN members mbr ON mbr.id = sl.member_id").
		Where("sl.id = ?", saleID).
		Scan(&sale).Error

	if err != nil {
		return responses.InternalServerError(c, "Failed to get sale", err)
	}

	// Ambil item pembelian terkait
	var items []models.AllSaleItems
	err = db.Table("sale_items sit").
		Select("sit.id, sit.sale_id, sit.product_id, pro.name AS product_name, sit.price, sit.qty, un.name AS unit_name, sit.sub_total").
		Joins("LEFT JOIN products pro ON pro.id = sit.product_id").
		Joins("LEFT JOIN units un ON un.id = pro.unit_id").
		Where("sit.sale_id = ?", saleID).
		Order("pro.name ASC").
		Scan(&items).Error

	if err != nil {
		return responses.InternalServerError(c, "Failed to get sale items", err)
	}

	// Format tanggal secara manual untuk respons ini
	// Menggunakan helper FormatIndonesianDate yang sudah kita buat
	formattedSaleDate := utils.FormatIndonesianDate(sale.SaleDate)

	// Buat objek respons menggunakan struct SaleItemResponse yang baru
	// dan isi field-fieldnya
	responseDetail := models.SaleItemResponse{
		ID:             sale.ID,
		MemberId:       sale.MemberId,
		MemberName:     sale.MemberName,
		SaleDate:       formattedSaleDate, // Gunakan tanggal yang sudah diformat
		TotalSale:      sale.TotalSale,
		Discount:       sale.Discount,
		ProfitEstimate: sale.ProfitEstimate,
		Payment:        string(sale.Payment),
		Items:          items,
	}

	// Panggil JSONResponse yang sudah ada, meneruskan SaleItemResponse sebagai 'data'
	return responses.JSONResponse(c, http.StatusOK, "Sale retrieved successfully", responseDetail)
}

// Request body struct untuk transaksi penjualan
type SaleTransactionRequest struct {
	Sale      models.Sales       `json:"sale" validate:"required"`
	SaleItems []models.SaleItems `json:"sale_items" validate:"required,min=1,dive"` // dive untuk validasi setiap item di slice
}
