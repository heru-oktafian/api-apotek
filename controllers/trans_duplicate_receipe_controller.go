package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/heru-oktafian/api-apotek/models"
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

	//--- VALIDASI INPUT ---

	subscriptionType, _ := middlewares.GetClaimsToken(c.Request, "subscription_type")

	branchID, _ := middlewares.GetBranchID(c.Request)

	userID, _ := middlewares.GetUserID(c.Request)

	err = utils.ValidateStruct(req)
	if err != nil {
		return responses.BadRequest(c, "Validate failed", err)
	}
	// --- AKHIR VALIDASI INPUT ---

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
	req.DuplicationReceipe.Description = req.DuplicationReceipe.Description
	req.DuplicationReceipe.DuplicateReceipeDate = nowWIB
	req.DuplicationReceipe.Payment = req.DuplicationReceipe.Payment
	req.DuplicationReceipe.UserID = userID
	req.DuplicationReceipe.BranchID = branchID
	req.DuplicationReceipe.CreatedAt = nowWIB

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

	// Commit transaksi jika semua berhasil
	err = tx.Commit().Error
	if err != nil {
		return responses.InternalServerError(c, "Failed to commit database transaction", err)
	}

	// Berhasil
	return responses.JSONResponse(c, http.StatusOK, "Sale transaction created successfully", req)
}

type DuplicateReceipeRequest struct {
	DuplicationReceipe models.DuplicateReceipes      `json:"duplication_receipe"`
	Items              []models.DuplicateRecipeItems `json:"items"`
}
