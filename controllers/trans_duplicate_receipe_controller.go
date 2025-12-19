package controllers

import (
	"fmt"
	"net/http"
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
		req.Items[i].DuplicateRecipeId = durID

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

	var input models.SaleInput
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

	var items []models.DuplicateRecipeItems
	if err := db.Where("duplicate_recipe_id = ?", id).Find(&items).Error; err != nil {
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
		return responses.InternalServerError(c, "Failed to update Duplicate receipe", err)
	}

	if err := reports.SyncSaleReport(db, sale); err != nil {
		return responses.InternalServerError(c, "Failed to sync Duplicate receipe report", err)
	}

	// _ = reports.AutoCleanupSales(db)
	_ = reports.SyncDailyProfitReport(db, sale)

	return responses.JSONResponse(c, http.StatusOK, "Duplicate receipe updated successfully", sale)
}

// DeleteDuplicateReceipe Function
func DeleteDuplicateReceipe(c *framework.Ctx) error {
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

type DuplicateReceipeRequest struct {
	DuplicationReceipe models.DuplicateReceipes      `json:"duplication_receipe"`
	Items              []models.DuplicateRecipeItems `json:"items"`
}
