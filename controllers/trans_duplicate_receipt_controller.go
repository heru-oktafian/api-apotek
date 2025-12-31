package controllers

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/heru-oktafian/api-apotek/models"
	"github.com/heru-oktafian/api-apotek/reports"
	"github.com/heru-oktafian/api-apotek/tools"
	"github.com/heru-oktafian/scafold/config"
	"github.com/heru-oktafian/scafold/framework"
	"github.com/heru-oktafian/scafold/helpers"
	"github.com/heru-oktafian/scafold/middlewares"
	"github.com/heru-oktafian/scafold/responses"
	"github.com/heru-oktafian/scafold/utils"
	"gorm.io/gorm"
)

// CreateDuplicateReceipt handles the creation of a new duplicate receipt record.
func CreateDuplicateReceipt(c *framework.Ctx) error {
	// Hitung waktu saat ini di zona WIB
	nowWIB := time.Now().In(utils.Location)

	db := config.DB
	var req DuplicateReceiptRequest

	// Deklarasi variabel 'err' untuk menangani error
	err := c.BodyParser(&req)
	if err != nil {
		return responses.BadRequest(c, "Invalid request body", err)
	}

	// Get default_member id dari token
	defaultMember, _ := middlewares.GetClaimsToken(c.Request, "default_member")

	//--- VALIDASI INPUT ---

	subscriptionType, _ := middlewares.GetClaimsToken(c.Request, "subscription_type")

	branchID, _ := middlewares.GetBranchID(c.Request)

	userID, _ := middlewares.GetUserID(c.Request)

	err = utils.ValidateStruct(req)
	if err != nil {
		return responses.BadRequest(c, "Validate failed", err)
	}
	// --- AKHIR VALIDASI INPUT ---
	// Modifikasi agar jika `member_id` tidak dikirim dalam request,
	// maka `member_id` diisi `defaultMember` dari deklarasi tersebut.
	if req.DuplicateReceipt.MemberId == "" {
		req.DuplicateReceipt.MemberId = defaultMember
	}

	if req.DuplicateReceipt.Payment == "" {
		req.DuplicateReceipt.Payment = "paid_by_cash"
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
	durID := helpers.GenerateID("DUR")
	req.DuplicateReceipt.ID = durID
	req.DuplicateReceipt.DuplicateReceiptDate = nowWIB
	req.DuplicateReceipt.UserID = userID
	req.DuplicateReceipt.BranchID = branchID
	req.DuplicateReceipt.CreatedAt = nowWIB
	req.DuplicateReceipt.UpdatedAt = nowWIB

	// Inisilisasi total & profit duplicate receipt
	totalDUR := 0
	totalProfDUR := 0

	for i := range req.Items {
		itemID := helpers.GenerateID("DRI")
		req.Items[i].ID = itemID
		req.Items[i].DuplicateReceiptId = durID

		// Dapatkan detail produk dari database
		var product models.Product
		err = tx.Where("id = ?", req.Items[i].ProductId).First(&product).Error
		if err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				return responses.NotFound(c, "Product with ID %s not found")
			}
			return responses.InternalServerError(c, "Failed to retrieve product details", err)
		}

		// Periksa ketersediaan stok produk
		if product.Stock < req.Items[i].Qty {
			tx.Rollback()
			return responses.BadRequest(c, fmt.Sprintf("Insufficient stock for product %s. Available: %d, Requested: %d", product.Name, product.Stock, req.Items[i].Qty), err)
		}

		// Kurangi stok produk
		newStock := product.Stock - req.Items[i].Qty
		err = tx.Model(&models.Product{}).Where("id = ?", product.ID).Update("stock", newStock).Error
		if err != nil {
			tx.Rollback()
			return responses.InternalServerError(c, fmt.Sprintf("Failed to update stock for product %s", product.Name), err)
		}

		// Update stock in Redis
		cacheKey := fmt.Sprintf("%s:%s:%s:%s", branchID, userID, product.ID, newStock)
		tools.UpdateProductStockInRedisAsync(cacheKey, product.ID, newStock)

		// Kalkulasi total_duplicate_recipe dan profit_estimate dari item-item
		totalDUR += req.Items[i].SubTotal
		// Profit per item = (Harga Jual - Harga Beli) * Qty
		totalProfDUR += (req.Items[i].Price - product.PurchasePrice) * req.Items[i].Qty
	}

	// Hitung total duplicate receipt dari item-item yang ada
	req.DuplicateReceipt.TotalDuplicateReceipt = totalDUR
	req.DuplicateReceipt.ProfitEstimate = totalProfDUR

	err = tx.Create(&req.DuplicateReceipt).Error
	if err != nil {
		tx.Rollback()
		return responses.InternalServerError(c, "Failed to create duplicate receipt", err)
	}

	err = tx.CreateInBatches(&req.Items, len(req.Items)).Error
	if err != nil {
		tx.Rollback()
		return responses.InternalServerError(c, "Failed to create duplicate items", err)
	}

	transactionReportID := durID // Gunakan ID yang sama dengan DuplicateReceipt ID
	transactionReport := models.TransactionReports{
		ID:              transactionReportID,
		TransactionType: models.Sale, // Tipe transaksi adalah "sale"
		UserID:          req.DuplicateReceipt.UserID,
		BranchID:        req.DuplicateReceipt.BranchID,
		Total:           req.DuplicateReceipt.TotalDuplicateReceipt,
		Payment:         req.DuplicateReceipt.Payment,
		CreatedAt:       nowWIB,
		UpdatedAt:       nowWIB,
	}
	err = tx.Create(&transactionReport).Error
	if err != nil {
		tx.Rollback()
		return responses.InternalServerError(c, "Failed to create transaction report", err)
	}

	// 4. Sinkronisasi laporan duplicate receipt
	// Inisialisasi laporan duplicate receipt
	var dailyProfit models.DailyProfitReport

	// Pastikan Duplicate date tidak nol saat diakses (validasi required sudah ada, tapi jaga-jaga)
	if req.DuplicateReceipt.DuplicateReceiptDate.IsZero() {
		tx.Rollback()
		return responses.BadRequest(c, "Date cannot be zero for daily profit report calculation. Please provide a valid date.", nil)
	}

	reportDate := req.DuplicateReceipt.DuplicateReceiptDate.Format("2006-01-02") // Format tanggal menjadi "YYYY-MM-DD"
	err = tx.Where("report_date = ? AND branch_id = ? AND user_id = ?", reportDate, branchID, userID).First(&dailyProfit).Error

	// Cek error selain record not found
	if err != nil && err != gorm.ErrRecordNotFound {
		tx.Rollback()
		return responses.InternalServerError(c, "Failed to check daily profit report", err)
	}

	if err == gorm.ErrRecordNotFound {
		// Jika belum ada, buat entri baru
		dailyProfitID := helpers.GenerateID("DPR")
		dailyProfit = models.DailyProfitReport{
			ID:             dailyProfitID,
			ReportDate:     req.DuplicateReceipt.DuplicateReceiptDate,
			UserID:         userID,
			BranchID:       branchID,
			TotalSales:     req.DuplicateReceipt.TotalDuplicateReceipt,
			ProfitEstimate: req.DuplicateReceipt.ProfitEstimate,
			CreatedAt:      nowWIB,
			UpdatedAt:      nowWIB,
		}
		err = tx.Create(&dailyProfit).Error
		if err != nil {
			tx.Rollback()
			return responses.InternalServerError(c, "Failed to create daily profit report", err)
		}
	} else {
		// Jika sudah ada, update total_sales dan profit_estimate yang sudah ada
		dailyProfit.TotalSales += req.DuplicateReceipt.TotalDuplicateReceipt
		dailyProfit.ProfitEstimate += req.DuplicateReceipt.ProfitEstimate
		dailyProfit.UpdatedAt = time.Now()
		err = tx.Save(&dailyProfit).Error
		if err != nil {
			tx.Rollback()
			return responses.InternalServerError(c, "Failed to update daily profit report", err)
		}
	}

	if subscriptionType == "quota" {
		var branch models.Branch
		err = tx.Where("id = ?", req.DuplicateReceipt.BranchID).First(&branch).Error
		if err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				return responses.NotFound(c, fmt.Sprintf("Branch with ID %s not found", req.DuplicateReceipt.BranchID))
			}
			return responses.InternalServerError(c, "Failed to retrieve branch details for quota update", err)
		}

		if branch.Quota > 0 {
			branch.Quota -= 1
			err = tx.Save(&branch).Error
			if err != nil {
				tx.Rollback()
				return responses.InternalServerError(c, fmt.Sprintf("Failed to update quota for branch %s", branch.BranchName), err)
			}
		} else {
			tx.Rollback()
			return responses.BadRequest(c, fmt.Sprintf("No quota available for branch %s", branch.BranchName), nil)
		}
	}

	if req.DuplicateReceipt.MemberId != "" && req.DuplicateReceipt.MemberId != defaultMember {
		var member models.Member
		err = tx.Where("id = ?", req.DuplicateReceipt.MemberId).First(&member).Error
		if err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				return responses.NotFound(c, fmt.Sprintf("Member with ID %s not found", req.DuplicateReceipt.MemberId))
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
			pointsEarned := float64(req.DuplicateReceipt.TotalDuplicateReceipt) / float64(memberCategory.PointsConversionRate)
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
	return responses.JSONResponse(c, http.StatusOK, "Duplicate receipt transaction created successfully", req)
}

// UpdateDuplicateReceipt Function (Modified)
func UpdateDuplicateReceipt(c *framework.Ctx) error {

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(utils.Location)

	branchID, _ := middlewares.GetBranchID(c.Request)
	userID, _ := middlewares.GetUserID(c.Request)

	total_before := 0
	profit_before := 0

	db := config.DB
	id := c.Param("id")

	var duplicate_receipt models.DuplicateReceipts
	if err := db.First(&duplicate_receipt, "id = ?", id).Error; err != nil {
		return responses.NotFound(c, "Receipt not found")
	}

	total_before += duplicate_receipt.TotalDuplicateReceipt
	profit_before += duplicate_receipt.ProfitEstimate

	var input models.DuplicateReceiptInput
	if err := c.BodyParser(&input); err != nil {
		return responses.BadRequest(c, "Invalid input", err)
	}

	if input.MemberId != nil {
		var member models.Member
		if err := db.Where("id = ?", *input.MemberId).First(&member).Error; err != nil {
			// Jika ID tidak valid, fallback ke default
			memberId, _ := middlewares.GetClaimsToken(c.Request, "default_member")
			duplicate_receipt.MemberId = memberId
		} else {
			duplicate_receipt.MemberId = *input.MemberId
		}
	}
	// Jika nil → tidak diubah, tetap pakai MemberID yang sudah ada

	if input.Payment != "" {
		duplicate_receipt.Payment = models.PaymentStatus(input.Payment)
	}

	duplicate_receipt.UpdatedAt = nowWIB

	var items []models.DuplicateReceiptItems
	if err := db.Where("duplicate_receipt_id = ?", id).Find(&items).Error; err != nil {
		return responses.InternalServerError(c, "Failed to fetch sale items", err)
	}

	total := 0
	for _, item := range items {
		total += item.SubTotal
	}

	profit := 0
	for _, item := range items {
		profit += item.SubTotal - item.Price*item.Qty
	}

	if err := db.Save(&duplicate_receipt).Error; err != nil {
		return responses.InternalServerError(c, "Failed to update Duplicate receipt", err)
	}

	if err := reports.SyncDuplicateReceiptReport(db, duplicate_receipt); err != nil {
		return responses.InternalServerError(c, "Failed to sync Duplicate receipt report", err)
	}

	// Sync laporan penjualan agar tetap konsisten
	_ = reports.SyncDailyProfitReport(db, branchID, userID, duplicate_receipt.DuplicateReceiptDate, total, profit, total_before, profit_before)

	return responses.JSONResponse(c, http.StatusOK, "Duplicate receipt updated successfully", duplicate_receipt)
}

// DeleteDuplicateReceipt Function
func DeleteDuplicateReceipt(c *framework.Ctx) error {
	db := config.DB
	id := c.Param("id")

	// Ambil duplicate receipt
	var duplicate_receipt models.DuplicateReceipts
	if err := db.First(&duplicate_receipt, "id = ?", id).Error; err != nil {
		return responses.NotFound(c, "Duplicate receipt not found")
	}

	// Ambil & hapus item, serta rollback stok
	var items []models.DuplicateReceiptItems
	if err := db.Where("duplicate_receipt_id = ?", id).Find(&items).Error; err == nil {
		for _, item := range items {
			_ = tools.SubtractProductStock(db, item.ProductId, item.Qty)
		}
		// Update stock in Redis asynchronously
		go func(items []models.DuplicateReceiptItems) {
			cacheKey := fmt.Sprintf("%s:%s", duplicate_receipt.BranchID, duplicate_receipt.UserID)
			for _, item := range items {
				var prod models.Product
				if err := db.Select("stock").Where("id = ?", item.ProductId).First(&prod).Error; err == nil {
					tools.UpdateProductStockInRedisAsync(cacheKey, item.ProductId, prod.Stock)
					tools.UpdatePurchaseProductStockInRedisAsync(cacheKey, item.ProductId, prod.Stock)
				}
			}
		}(items)
		db.Where("duplicate_receipt_id = ?", id).Delete(&models.DuplicateReceiptItems{})
	}

	// Hapus laporan transaksi
	if err := db.Where("id = ? AND transaction_type = ?", duplicate_receipt.ID, models.Sale).Delete(&models.TransactionReports{}).Error; err != nil {
		return responses.InternalServerError(c, "Failed to delete transaction report", err)
	}

	// Hapus data penjualan
	if err := db.Delete(&duplicate_receipt).Error; err != nil {
		return responses.InternalServerError(c, "Failed to delete sale", err)
	}

	// Delete laporan profit harian asynchronously
	go func() {
		if err := reports.DeleteDailyProfitReport(db, id, "duplicate_receipt"); err != nil {
			fmt.Printf("Failed to delete daily profit report asynchronously: %v\n", err)
		}
	}()

	return responses.JSONResponse(c, http.StatusOK, "Duplicate receipt deleted successfully", duplicate_receipt)
}

type DuplicateReceiptRequest struct {
	DuplicateReceipt models.DuplicateReceipts       `json:"duplicate_receipt"`
	Items            []models.DuplicateReceiptItems `json:"items"`
}

// CreateDuplicateReceiptItem Function
func CreateDuplicateReceiptItem(c *framework.Ctx) error {
	// Get branch and user IDs from middleware
	branchID, _ := middlewares.GetBranchID(c.Request)
	userID, _ := middlewares.GetUserID(c.Request)

	var item models.DuplicateReceiptItems

	db := config.DB

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

	// Cek apakah item dengan duplicate_receipt_id dan product_id sudah ada
	var existing models.DuplicateReceiptItems
	err := db.Where("duplicate_receipt_id = ? AND product_id = ?", item.DuplicateReceiptId, item.ProductId).First(&existing).Error
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

		// Supporting operations asynchronously
		go func() {
			// Update stock in Redis
			cacheKey := fmt.Sprintf("%s:%s", branchID, userID)
			var prod models.Product
			if err := db.Select("stock").Where("id = ?", item.ProductId).First(&prod).Error; err == nil {
				tools.UpdateProductStockInRedisAsync(cacheKey, item.ProductId, prod.Stock)
			}

			if err := reports.RecalculateTotalDuplicate(db, item.DuplicateReceiptId); err != nil {
				fmt.Printf("Failed to recalculate total duplicate asynchronously: %v\n", err)
			}
		}()

		return responses.JSONResponse(c, http.StatusOK, "Item updated successfully", existing)

	} else if err != gorm.ErrRecordNotFound {
		return responses.InternalServerError(c, "Failed to find existing sale item", err)
	}

	// Data belum ada, buat item baru
	if item.ID == "" {
		item.ID = helpers.GenerateID("DRI")
	}
	item.SubTotal = item.Qty * item.Price

	if err := db.Create(&item).Error; err != nil {
		return responses.InternalServerError(c, "Failed to create sale item", err)
	}

	if err := tools.ReduceProductStock(db, item.ProductId, item.Qty); err != nil {
		return responses.InternalServerError(c, "Failed to reduce product stock", err)
	}

	// Recalculate total and sync reports asynchronously
	go func() {
		// Update stock in Redis
		cacheKey := fmt.Sprintf("%s:%s", branchID, userID)
		var prod models.Product
		if err := db.Select("stock").Where("id = ?", item.ProductId).First(&prod).Error; err == nil {
			tools.UpdateProductStockInRedisAsync(cacheKey, item.ProductId, prod.Stock)
		}

		if err := reports.RecalculateTotalDuplicate(db, item.DuplicateReceiptId); err != nil {
			fmt.Printf("Failed to recalculate total duplicate asynchronously: %v\n", err)
		}

		// Sync laporan profit harian
		var duplicateReceipt models.DuplicateReceipts
		if err := db.First(&duplicateReceipt, "id = ?", item.DuplicateReceiptId).Error; err != nil {
			fmt.Printf("Failed to fetch duplicate receipt asynchronously: %v\n", err)
			return
		}

		if err := reports.SyncDailyProfitReport(db, branchID, userID, duplicateReceipt.DuplicateReceiptDate, duplicateReceipt.TotalDuplicateReceipt, duplicateReceipt.ProfitEstimate, 0, 0); err != nil {
			fmt.Printf("Failed to sync daily profit report asynchronously: %v\n", err)
		}
	}()

	return responses.JSONResponse(c, http.StatusOK, "Item added successfully", item)
}

// UpdateDuplicateReceiptItem Function
func UpdateDuplicateReceiptItem(c *framework.Ctx) error {
	db := config.DB
	id := c.Param("id")

	branchID, _ := middlewares.GetBranchID(c.Request)
	userID, _ := middlewares.GetUserID(c.Request)

	var existingItem models.DuplicateReceiptItems
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

	// Supporting operations asynchronously
	go func() {
		cacheKey := fmt.Sprintf("%s:%s", branchID, userID)
		// Update stock in Redis for old product
		var oldProd models.Product
		if err := db.Select("stock").Where("id = ?", existingItem.ProductId).First(&oldProd).Error; err == nil {
			tools.UpdateProductStockInRedisAsync(cacheKey, existingItem.ProductId, oldProd.Stock)
		}

		// Update stock in Redis for new product
		var newProd models.Product
		if err := db.Select("stock").Where("id = ?", updatedData.ProductId).First(&newProd).Error; err == nil {
			tools.UpdateProductStockInRedisAsync(cacheKey, updatedData.ProductId, newProd.Stock)
		}

		if err := reports.RecalculateTotalDuplicate(db, existingItem.DuplicateReceiptId); err != nil {
			fmt.Printf("Failed to recalculate total duplicate asynchronously: %v\n", err)
		}

		// Sync laporan profit harian
		var duplicateReceipt models.DuplicateReceipts
		if err := db.First(&duplicateReceipt, "id = ?", existingItem.DuplicateReceiptId).Error; err != nil {
			fmt.Printf("Failed to fetch duplicate receipt asynchronously: %v\n", err)
			return
		}

		if err := reports.SyncDailyProfitReport(db, branchID, userID, duplicateReceipt.DuplicateReceiptDate, duplicateReceipt.TotalDuplicateReceipt, duplicateReceipt.ProfitEstimate, 0, 0); err != nil {
			fmt.Printf("Failed to sync daily profit report asynchronously: %v\n", err)
		}
	}()

	return responses.JSONResponse(c, http.StatusOK, "Item updated successfully", existingItem)
}

// Delete DuplicateReceiptItem
func DeleteDuplicateReceiptItem(c *framework.Ctx) error {
	db := config.DB
	id := c.Param("id")

	branchID, _ := middlewares.GetBranchID(c.Request)
	userID, _ := middlewares.GetUserID(c.Request)

	var item models.DuplicateReceiptItems
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

	// Supporting operations asynchronously
	go func() {
		// Update stock in Redis
		cacheKey := fmt.Sprintf("%s:%s", branchID, userID)
		var prod models.Product
		if err := db.Select("stock").Where("id = ?", item.ProductId).First(&prod).Error; err == nil {
			tools.UpdateProductStockInRedisAsync(cacheKey, item.ProductId, prod.Stock)
		}

		if err := reports.RecalculateTotalDuplicate(db, item.DuplicateReceiptId); err != nil {
			fmt.Printf("Failed to recalculate total duplicate asynchronously: %v\n", err)
		}
	}()

	return responses.JSONResponse(c, http.StatusOK, "Item deleted successfully", item)
}

// GetAllDuplicateReceipts tampilkan semua duplicate receipt items
func GetAllDuplicateReceipts(c *framework.Ctx) error {
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

	var salesFromDB []models.AllDuplicateReceipts // Gunakan models.AllDuplicateReceipts untuk mengambil data dari DB
	var total int64

	query := config.DB.Table("duplicate_receipts dr").
		Select("dr.id, dr.member_id, mbr.name AS member_name, dr.duplicate_receipt_date, dr.total_duplicate_receipt, dr.profit_estimate, dr.payment").
		Joins("LEFT JOIN members mbr on mbr.id = dr.member_id").
		Where("dr.branch_id = ? AND dr.total_duplicate_receipt > 0", branchID).
		Order("dr.created_at DESC")

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
		query = query.Where("dr.duplicate_receipt_date BETWEEN ? AND ?", startDate, endDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return responses.InternalServerError(c, "Get sale failed", err)
	}

	if err := query.Offset(offset).Limit(limit).Scan(&salesFromDB).Error; err != nil {
		return responses.InternalServerError(c, "Get sales failed", err)
	}

	// Buat slice baru untuk menampung data yang sudah diformat
	var formattedDuplicateData []models.DuplicateDetailResponse
	for _, duplicate_receipt := range salesFromDB {
		formattedDuplicateData = append(formattedDuplicateData, models.DuplicateDetailResponse{
			ID:                    duplicate_receipt.ID,
			MemberId:              duplicate_receipt.MemberId,
			MemberName:            duplicate_receipt.MemberName,
			DuplicateReceiptDate:  utils.FormatIndonesianDate(duplicate_receipt.DuplicateReceiptDate), // Format tanggal di sini
			TotalDuplicateReceipt: duplicate_receipt.TotalDuplicateReceipt,
			ProfitEstimate:        duplicate_receipt.ProfitEstimate,
			Payment:               string(duplicate_receipt.Payment),
		})
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	// Gunakan JSONResponseGetAll helper dengan data yang sudah diformat
	return responses.JSONResponseGetAll(
		c,
		http.StatusOK,
		"Duplicate receipts retrieved successfully",
		search,
		int(total),
		page,
		totalPages,
		limit,
		formattedDuplicateData, // Kirim data yang sudah diformat (slice dari DuplicateDetailResponse)
	)
}

// GetAllDuplicateDetail menampilkan sales dengan kolom description yang
// berisi daftar nama item (dipisah koma) diikuti dengan duplicate_receipt_date (DD-MM-YYYY HH:MM)
// di mana waktu ditambah 7 jam sesuai permintaan.
func GetAllDuplicateDetail(c *framework.Ctx) error {
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
	type duplicateSummary struct {
		ID                    string
		TotalDuplicateReceipt int
		Payment               string
		SaleDate              time.Time
	}

	var salesFromDB []duplicateSummary
	var total int64

	query := config.DB.Table("duplicate_receipts dr").
		Select("dr.id, dr.total_duplicate_receipt, dr.payment, dr.duplicate_receipt_date").
		Joins("LEFT JOIN members mbr on mbr.id = dr.member_id").
		Where("dr.branch_id = ? AND dr.total_duplicate_receipt > 0", branchID).
		Order("dr.created_at DESC")

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
		query = query.Where("dr.duplicate_receipt_date BETWEEN ? AND ?", startDate, endDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return responses.InternalServerError(c, "Get duplicate receipt failed", err)
	}

	if err := query.Offset(offset).Limit(limit).Scan(&salesFromDB).Error; err != nil {
		return responses.InternalServerError(c, "Get duplicate receipt failed", err)
	}

	// Bentuk response dengan description
	var formatted []map[string]interface{}
	for _, s := range salesFromDB {
		// Ambil nama item untuk sale ini
		var itemNames []string
		if err := config.DB.Table("duplicate_receipt_items dri").
			Select("pro.name").
			Joins("LEFT JOIN products pro ON pro.id = dri.product_id").
			Where("dri.duplicate_receipt_id = ?", s.ID).
			Order("pro.name ASC").
			Pluck("pro.name", &itemNames).Error; err != nil {
			return responses.InternalServerError(c, "Failed to get duplicate receipt items", err)
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
			"id":                      s.ID,
			"total_duplicate_receipt": s.TotalDuplicateReceipt,
			"payment":                 s.Payment,
			"description":             description,
		})
	}

	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return responses.JSONResponseGetAll(
		c,
		http.StatusOK,
		"Duplicate receipts retrieved successfully",
		search,
		int(total),
		page,
		totalPages,
		limit,
		formatted,
	)
}

// GetAllDuplicateItems tampilkan semua item berdasarkan duplicate_receipt_id tanpa pagination
func GetAllDuplicateItems(c *framework.Ctx) error {
	// Get duplicate receipt id dari param
	duplicateReceiptID := c.Param("id")

	// Parsing body JSON ke struct
	var DuplicateItems []models.AllDuplicateReceiptItems

	// Query dasar
	query := config.DB.Table("duplicate_receipt_items dri").
		Select("dri.id, dri.duplicate_receipt_id, dri.product_id, pro.name AS product_name, dri.price, dri.qty, un.name AS unit_name, dri.sub_total").
		Joins("LEFT JOIN products pro ON pro.id = dri.product_id").
		Joins("LEFT JOIN units un ON un.id = pro.unit_id").
		Where("dri.duplicate_receipt_id = ?", duplicateReceiptID).
		Order("pro.name ASC")

	// Eksekusi query
	if err := query.Scan(&DuplicateItems).Error; err != nil {
		return responses.InternalServerError(c, "Get items failed", err)
	}

	return responses.JSONResponse(c, http.StatusOK, "Items retrieved successfully", DuplicateItems)
}

// GetDuplicateWithItems menampilkan satu duplicate receipt beserta semua item-nya
func GetDuplicateWithItems(c *framework.Ctx) error {
	db := config.DB

	duplicateId := c.Param("id")

	// Gunakan models.AllDuplicateReceipts untuk mengambil data dari DB
	var duplicate_receipts models.AllDuplicateReceipts

	err := db.Table("duplicate_receipts dr").
		Select("dr.id, dr.member_id, mbr.name AS member_name, dr.duplicate_receipt_date, dr.total_duplicate_receipt, dr.profit_estimate, dr.payment").
		Joins("LEFT JOIN members mbr ON mbr.id = dr.member_id").
		Where("dr.id = ?", duplicateId).
		Scan(&duplicate_receipts).Error
	if err != nil {
		return responses.InternalServerError(c, "Failed to get duplicate receipt", err)
	}

	// Ambil item pembelian terkait
	var items []models.AllDuplicateReceiptItems
	err = db.Table("duplicate_receipt_items dri").
		Select("dri.id, dri.duplicate_receipt_id, dri.product_id, pro.name AS product_name, dri.price, dri.qty, un.name AS unit_name, dri.sub_total").
		Joins("LEFT JOIN products pro ON pro.id = dri.product_id").
		Joins("LEFT JOIN units un ON un.id = pro.unit_id").
		Where("dri.duplicate_receipt_id = ?", duplicateId).
		Order("pro.name ASC").
		Scan(&items).Error

	if err != nil {
		return responses.InternalServerError(c, "Failed to get duplicate receipt items", err)
	}

	// Format tanggal secara manual untuk respons ini
	// Menggunakan helper FormatIndonesianDate yang sudah kita buat
	formattedDuplicateDate := utils.FormatIndonesianDate(duplicate_receipts.DuplicateReceiptDate)

	// Buat objek respons menggunakan struct DuplicateItemResponse yang baru
	// dan isi field-fieldnya
	responseDetail := models.DuplicateItemResponse{
		ID:                    duplicate_receipts.ID,
		MemberId:              duplicate_receipts.MemberId,
		MemberName:            duplicate_receipts.MemberName,
		DuplicateReceiptDate:  formattedDuplicateDate, // Gunakan tanggal yang sudah diformat
		TotalDuplicateReceipt: duplicate_receipts.TotalDuplicateReceipt,
		ProfitEstimate:        duplicate_receipts.ProfitEstimate,
		Payment:               string(duplicate_receipts.Payment),
		Items:                 items,
	}

	// Panggil JSONResponse yang sudah ada, meneruskan DuplicateItemResponse sebagai 'data'
	return responses.JSONResponse(c, http.StatusOK, "Duplicate receipt retrieved successfully", responseDetail)
}
