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

// CreateDuplicateReceipe handles the creation of a new duplicate receipe record.
func CreateDuplicateReceipe(c *framework.Ctx) error {
	// Hitung waktu saat ini di zona WIB
	nowWIB := time.Now().In(utils.Location)

	db := config.DB
	var req DuplicateReceipeRequest

	// Deklarasi variabel ''err' untuk menangani error
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
	if req.DuplicationReceipe.MemberId == "" {
		req.DuplicationReceipe.MemberId = defaultMember
	}

	if req.DuplicationReceipe.Payment == "" {
		req.DuplicationReceipe.Payment = "paid_by_cash"
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
	req.DuplicationReceipe.ID = durID
	req.DuplicationReceipe.DuplicateReceipeDate = nowWIB
	req.DuplicationReceipe.UserID = userID
	req.DuplicationReceipe.BranchID = branchID
	req.DuplicationReceipe.CreatedAt = nowWIB
	req.DuplicationReceipe.UpdatedAt = nowWIB

	// Inisilisasi total & profit duplicate receipe
	totalDUR := 0
	totalProfDUR := 0

	for i := range req.Items {
		itemID := helpers.GenerateID("DRI")
		req.Items[i].ID = itemID
		req.Items[i].DuplicateReceipeId = durID

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

		// Kalkulasi total_duplicate_recipe dan profit_estimate dari item-item
		totalDUR += req.Items[i].SubTotal
		// Profit per item = (Harga Jual - Harga Beli) * Qty
		totalProfDUR += (req.Items[i].Price - product.PurchasePrice) * req.Items[i].Qty
	}

	// Hitung total duplicate receipe dari item-item yang ada
	req.DuplicationReceipe.TotalDuplicateReceipe = totalDUR
	req.DuplicationReceipe.ProfitEstimate = totalProfDUR

	err = tx.Create(&req.DuplicationReceipe).Error
	if err != nil {
		tx.Rollback()
		return responses.InternalServerError(c, "Failed to create duplicate receipe", err)
	}

	err = tx.CreateInBatches(&req.Items, len(req.Items)).Error
	if err != nil {
		tx.Rollback()
		return responses.InternalServerError(c, "Failed to create duplicate items", err)
	}

	transactionReportID := helpers.GenerateID("TRX")
	transactionReport := models.TransactionReports{
		ID:              transactionReportID,
		TransactionType: models.Sale, // Tipe transaksi adalah "sale"
		UserID:          req.DuplicationReceipe.UserID,
		BranchID:        req.DuplicationReceipe.BranchID,
		Total:           req.DuplicationReceipe.TotalDuplicateReceipe,
		Payment:         req.DuplicationReceipe.Payment,
		CreatedAt:       nowWIB,
		UpdatedAt:       nowWIB,
	}
	err = tx.Create(&transactionReport).Error
	if err != nil {
		tx.Rollback()
		return responses.InternalServerError(c, "Failed to create transaction report", err)
	}

	var dailyProfit models.DailyProfitReport
	// Pastikan Duplicate date tidak nol saat diakses (validasi required sudah ada, tapi jaga-jaga)
	if req.DuplicationReceipe.DuplicateReceipeDate.IsZero() {
		tx.Rollback()
		return responses.BadRequest(c, "Date cannot be zero for daily profit report calculation. Please provide a valid date.", nil)
	}

	reportDate := req.DuplicationReceipe.DuplicateReceipeDate.Format("2006-01-02") // Format tanggal menjadi "YYYY-MM-DD"
	err = tx.Where("report_date = ? AND branch_id = ? AND user_id = ?", reportDate, req.DuplicationReceipe.BranchID, req.DuplicationReceipe.UserID).First(&dailyProfit).Error

	if err != nil && err != gorm.ErrRecordNotFound {
		tx.Rollback()
		return responses.InternalServerError(c, "Failed to check daily profit report", err)
	}

	if err == gorm.ErrRecordNotFound {
		// Jika belum ada, buat entri baru
		dailyProfitID := helpers.GenerateID("DPR")
		dailyProfit = models.DailyProfitReport{
			ID:             dailyProfitID,
			ReportDate:     req.DuplicationReceipe.DuplicateReceipeDate,
			UserID:         req.DuplicationReceipe.UserID,
			BranchID:       req.DuplicationReceipe.BranchID,
			TotalSales:     req.DuplicationReceipe.TotalDuplicateReceipe,
			ProfitEstimate: req.DuplicationReceipe.ProfitEstimate,
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
		dailyProfit.TotalSales += req.DuplicationReceipe.TotalDuplicateReceipe
		dailyProfit.ProfitEstimate += req.DuplicationReceipe.ProfitEstimate
		dailyProfit.UpdatedAt = time.Now()
		err = tx.Save(&dailyProfit).Error
		if err != nil {
			tx.Rollback()
			return responses.InternalServerError(c, "Failed to update daily profit report", err)
		}
	}

	if subscriptionType == "quota" {
		var branch models.Branch
		err = tx.Where("id = ?", req.DuplicationReceipe.BranchID).First(&branch).Error
		if err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				return responses.NotFound(c, fmt.Sprintf("Branch with ID %s not found", req.DuplicationReceipe.BranchID))
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

	if req.DuplicationReceipe.MemberId != "" && req.DuplicationReceipe.MemberId != defaultMember {
		var member models.Member
		err = tx.Where("id = ?", req.DuplicationReceipe.MemberId).First(&member).Error
		if err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				return responses.NotFound(c, fmt.Sprintf("Member with ID %s not found", req.DuplicationReceipe.MemberId))
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
			pointsEarned := float64(req.DuplicationReceipe.TotalDuplicateReceipe) / float64(memberCategory.PointsConversionRate)
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
	return responses.JSONResponse(c, http.StatusOK, "Duplicate receipe transaction created successfully", req)
}

// UpdateDuplicateReceipe Function (Modified)
func UpdateDuplicateReceipe(c *framework.Ctx) error {

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(utils.Location)

	db := config.DB
	id := c.Param("id")

	var duplicate_receipe models.DuplicateReceipes
	if err := db.First(&duplicate_receipe, "id = ?", id).Error; err != nil {
		return responses.NotFound(c, "Receipe not found")
	}

	var input models.DuplicateReceipeInput
	if err := c.BodyParser(&input); err != nil {
		return responses.BadRequest(c, "Invalid input", err)
	}

	if input.MemberId != nil {
		var member models.Member
		if err := db.Where("id = ?", *input.MemberId).First(&member).Error; err != nil {
			// Jika ID tidak valid, fallback ke default
			memberId, _ := middlewares.GetClaimsToken(c.Request, "default_member")
			duplicate_receipe.MemberId = memberId
		} else {
			duplicate_receipe.MemberId = *input.MemberId
		}
	}
	// Jika nil → tidak diubah, tetap pakai MemberID yang sudah ada

	if input.Payment != "" {
		duplicate_receipe.Payment = models.PaymentStatus(input.Payment)
	}

	duplicate_receipe.UpdatedAt = nowWIB

	var items []models.DuplicateReceipeItems
	if err := db.Where("duplicate_recipe_id = ?", id).Find(&items).Error; err != nil {
		return responses.InternalServerError(c, "Failed to fetch sale items", err)
	}

	total := 0
	for _, item := range items {
		total += item.SubTotal
	}

	if err := db.Save(&duplicate_receipe).Error; err != nil {
		return responses.InternalServerError(c, "Failed to update Duplicate receipe", err)
	}

	if err := reports.SyncDuplicateReceipeReport(db, duplicate_receipe); err != nil {
		return responses.InternalServerError(c, "Failed to sync Duplicate receipe report", err)
	}

	// _ = reports.AutoCleanupSales(db)
	_ = reports.SyncDuplicateReceipeReport(db, duplicate_receipe)

	return responses.JSONResponse(c, http.StatusOK, "Duplicate receipe updated successfully", duplicate_receipe)
}

// DeleteDuplicateReceipe Function
func DeleteDuplicateReceipe(c *framework.Ctx) error {
	db := config.DB
	id := c.Param("id")

	// Ambil duplicate receipe
	var duplicate_receipe models.DuplicateReceipes
	if err := db.First(&duplicate_receipe, "id = ?", id).Error; err != nil {
		return responses.NotFound(c, "Duplicate receipe not found")
	}

	// Ambil & hapus item, serta rollback stok
	var items []models.DuplicateReceipeItems
	if err := db.Where("duplicate_receipe_id = ?", id).Find(&items).Error; err == nil {
		for _, item := range items {
			_ = tools.SubtractProductStock(db, item.ProductId, item.Qty)
		}
		db.Where("duplicate_receipe_id = ?", id).Delete(&models.DuplicateReceipeItems{})
	}

	// Hapus laporan transaksi
	if err := db.Where("id = ? AND transaction_type = ?", duplicate_receipe.ID, models.Sale).Delete(&models.TransactionReports{}).Error; err != nil {
		return responses.InternalServerError(c, "Failed to delete transaction report", err)
	}

	// Hapus data penjualan
	if err := db.Delete(&duplicate_receipe).Error; err != nil {
		return responses.InternalServerError(c, "Failed to delete sale", err)
	}

	// Delete laporan profit harian
	_ = reports.DeleteDailyProfitReport(db, id)

	// (Opsional) Sync laporan penjualan agar tetap konsisten
	_ = reports.SyncDuplicateReceipeReport(db, duplicate_receipe)

	return responses.JSONResponse(c, http.StatusOK, "Duplicate receipe deleted successfully", duplicate_receipe)
}

type DuplicateReceipeRequest struct {
	DuplicationReceipe models.DuplicateReceipes       `json:"duplication_receipe"`
	Items              []models.DuplicateReceipeItems `json:"items"`
}

// CreateDuplicateRecipeItem Function
func CreateDuplicateRecipeItem(c *framework.Ctx) error {
	db := config.DB
	var item models.DuplicateReceipeItems

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

	// Cek apakah item dengan duplicate_receipe_id dan product_id sudah ada
	var existing models.DuplicateReceipeItems
	err := db.Where("duplicate_receipe_id = ? AND product_id = ?", item.DuplicateReceipeId, item.ProductId).First(&existing).Error
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

		if err := reports.RecalculateTotalSale(db, item.DuplicateReceipeId); err != nil {
			return responses.InternalServerError(c, "Failed to recalculate total sale", err)
		}

		// Sync laporan profit harian
		var duplicateReceipe models.DuplicateReceipes
		if err := db.First(&duplicateReceipe, "id = ?", item.DuplicateReceipeId).Error; err != nil {
			return responses.InternalServerError(c, "Failed to fetch duplicate receipe", err)
		}

		_ = reports.SyncDailyDuplicateProfitReport(db, duplicateReceipe)

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

	if err := reports.RecalculateTotalSale(db, item.DuplicateReceipeId); err != nil {
		return responses.InternalServerError(c, "Failed to recalculate total sale", err)
	}

	// Sync laporan profit harian
	var duplicateReceipe models.DuplicateReceipes
	if err := db.First(&duplicateReceipe, "id = ?", item.DuplicateReceipeId).Error; err != nil {
		return responses.InternalServerError(c, "Failed to fetch duplicate receipe", err)
	}

	_ = reports.SyncDailyDuplicateProfitReport(db, duplicateReceipe)
	return responses.JSONResponse(c, http.StatusOK, "Item added successfully", item)
}

// UpdateDuplicateRecipeItem Function
func UpdateDuplicateRecipeItem(c *framework.Ctx) error {
	db := config.DB
	id := c.Param("id")

	var existingItem models.DuplicateReceipeItems
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

	if err := reports.RecalculateTotalSale(db, existingItem.DuplicateReceipeId); err != nil {
		return responses.InternalServerError(c, "Failed to recalculate total sale", err)
	}

	// Sync laporan profit harian
	var duplicateReceipe models.DuplicateReceipes
	if err := db.First(&duplicateReceipe, "id = ?", existingItem.DuplicateReceipeId).Error; err != nil {
		return responses.InternalServerError(c, "Failed to fetch duplicate receipe", err)
	}

	_ = reports.SyncDailyDuplicateProfitReport(db, duplicateReceipe)

	return responses.JSONResponse(c, http.StatusOK, "Item updated successfully", existingItem)
}

// Delete DuplicateReceipeItem
func DeleteDuplicateReceipeItem(c *framework.Ctx) error {
	db := config.DB
	id := c.Param("id")

	var item models.DuplicateReceipeItems
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
	if err := reports.RecalculateTotalSale(db, item.DuplicateReceipeId); err != nil {
		return responses.InternalServerError(c, "Failed to recalculate total sale", err)
	}

	return responses.JSONResponse(c, http.StatusOK, "Item deleted successfully", item)
}

// GetAllDuplicateReceipes tampilkan semua duplicate receipe items
func GetAllDuplicateReceipes(c *framework.Ctx) error {
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

	var salesFromDB []models.AllDuplicateReceipes // Gunakan models.AllDuplicateReceipes untuk mengambil data dari DB
	var total int64

	query := config.DB.Table("duplicate_receipes dr").
		Select("dr.id, dr.member_id, mbr.name AS member_name, dr.duplicate_receipe_date, dr.total_sale, dr.discount, dr.profit_estimate, dr.payment").
		Joins("LEFT JOIN members mbr on mbr.id = dr.member_id").
		Where("dr.branch_id = ? AND dr.total_sale > 0", branchID).
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
		query = query.Where("sl.duplicate_receipe_date BETWEEN ? AND ?", startDate, endDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return responses.InternalServerError(c, "Get sale failed", err)
	}

	if err := query.Offset(offset).Limit(limit).Scan(&salesFromDB).Error; err != nil {
		return responses.InternalServerError(c, "Get sales failed", err)
	}

	// Buat slice baru untuk menampung data yang sudah diformat
	var formattedDuplicateData []models.DuplicateDetailResponse
	for _, duplicate_receipe := range salesFromDB {
		formattedDuplicateData = append(formattedDuplicateData, models.DuplicateDetailResponse{
			ID:                    duplicate_receipe.ID,
			MemberId:              duplicate_receipe.MemberId,
			MemberName:            duplicate_receipe.MemberName,
			DuplicateReceipeDate:  utils.FormatIndonesianDate(duplicate_receipe.DuplicateReceipeDate), // Format tanggal di sini
			TotalDuplicateReceipe: duplicate_receipe.TotalDuplicateReceipe,
			ProfitEstimate:        duplicate_receipe.ProfitEstimate,
			Payment:               string(duplicate_receipe.Payment),
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
		formattedDuplicateData, // Kirim data yang sudah diformat (slice dari DuplicateDetailResponse)
	)
}
