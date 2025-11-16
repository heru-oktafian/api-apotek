package controllers

import (
	"errors"
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

// CreateAnotherIncome Function
func CreateAnotherIncome(c *framework.Ctx) error {

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(utils.Location)

	db := config.DB

	// Ambil informasi dari token
	branchID, _ := middlewares.GetBranchID(c.Request)
	userID, _ := middlewares.GetUserID(c.Request)
	generatedID := helpers.GenerateID("ANI")

	// Ambil input dari body
	var input models.AnotherIncomeInput
	if err := c.BodyParser(&input); err != nil {
		return responses.BadRequest(c, "Invalid input", err)
	}

	// Parse tanggal
	layout := "2006-01-02" // format harus YYYY-MM-DD
	parsedDate, err := time.Parse(layout, input.IncomeDate)
	description := input.Description
	payment := input.Payment
	total := input.TotalIncome
	if err != nil {
		return responses.BadRequest(c, "Invalid date format. Use YYYY-MM-DD", err)
	}

	// Map ke struct model
	another_income := models.AnotherIncomes{
		ID:          generatedID,
		Description: description,
		BranchID:    branchID,
		UserID:      userID,
		IncomeDate:  parsedDate,
		TotalIncome: total,
		Payment:     models.PaymentStatus(payment),
		CreatedAt:   nowWIB,
		UpdatedAt:   nowWIB,
	}

	// Simpan another_income
	if err := db.Create(&another_income).Error; err != nil {
		return responses.InternalServerError(c, "Failed to create Another Income", err)
	}

	// Buat laporan
	if err := SyncAnotherIncomeReport(db, another_income); err != nil {
		return responses.InternalServerError(c, "Failed to sync Another Income report", err)
	}

	return responses.JSONResponse(c, http.StatusOK, "Another Income created successfully", another_income)
}

// UpdateAnotherIncomeItem Function
func UpdateAnotherIncome(c *framework.Ctx) error {
	db := config.DB
	id := c.Param("id")

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(utils.Location)

	// Cari data another_income
	var another_income models.AnotherIncomes
	if err := db.First(&another_income, "id = ?", id).Error; err != nil {
		return responses.NotFound(c, "Another Income not found")
	}

	// Gunakan struct khusus input
	var input models.AnotherIncomeInput
	if err := c.BodyParser(&input); err != nil {
		return responses.BadRequest(c, "Invalid input", err)
	}

	// Parse tanggal dari string ke time.Time
	layout := "2006-01-02"
	parsedDate, err := time.Parse(layout, input.IncomeDate)
	if err != nil {
		return responses.BadRequest(c, "Invalid date format. Use YYYY-MM-DD", err)
	}

	// Update field dasar
	another_income.IncomeDate = parsedDate
	another_income.Description = input.Description
	another_income.TotalIncome = input.TotalIncome
	another_income.Payment = models.PaymentStatus(input.Payment)
	another_income.UpdatedAt = nowWIB

	// Simpan update
	if err := db.Save(&another_income).Error; err != nil {
		return responses.InternalServerError(c, "Failed to update Another Income", err)
	}

	// Sync report
	if err := SyncAnotherIncomeReport(db, another_income); err != nil {
		return responses.InternalServerError(c, "Failed to sync Another Income report", err)
	}

	return responses.JSONResponse(c, http.StatusOK, "Another Income updated successfully", another_income)
}

// DeleteAnotherIncomeItem Function
func DeleteAnotherIncome(c *framework.Ctx) error {
	db := config.DB
	id := c.Param("id")

	// Ambil another_income
	var another_income models.AnotherIncomes
	if err := db.First(&another_income, "id = ?", id).Error; err != nil {
		return responses.NotFound(c, "Another Income not found")
	}

	// Hapus laporan
	if err := db.Where("id = ? AND transaction_type = ?", another_income.ID, models.Income).Delete(&models.TransactionReports{}).Error; err != nil {
		return responses.InternalServerError(c, "Failed to delete transaction report", err)
	}

	// Hapus another_income
	if err := db.Delete(&another_income).Error; err != nil {
		return responses.InternalServerError(c, "Failed to delete Another Income", err)
	}

	return responses.JSONResponse(c, http.StatusOK, "Another Income deleted successfully", another_income)
}

// GetAllAnotherIncome tampilkan semua AnotherIncome
func GetAllAnotherIncomes(c *framework.Ctx) error {
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

	var AnotherIncome []models.AnotherIncomes
	var total int64

	// Buat builder kueri yang bersih untuk menghitung dan mengambil data
	countQuery := config.DB.Table("another_incomes ex").
		Where("ex.branch_id = ?", branchID)

	dataQuery := config.DB.Table("another_incomes ex").
		Select("ex.id, ex.description, ex.income_date, ex.total_income, ex.payment").
		Where("ex.branch_id = ?", branchID)

	// Terapkan filter pencarian
	if search != "" {
		search = strings.ToLower(search)
		countQuery = countQuery.Where("LOWER(ex.description) LIKE ? ", "%"+search+"%")
		dataQuery = dataQuery.Where("LOWER(ex.description) LIKE ? ", "%"+search+"%")
	}

	// Terapkan filter bulan
	if month != "" {
		parsedMonth, err := time.Parse("2006-01", month)
		if err != nil {
			return responses.BadRequest(c, "Invalid month format", err)
		}
		startDate := parsedMonth
		endDate := startDate.AddDate(0, 1, 0).Add(-time.Nanosecond)
		countQuery = countQuery.Where("ex.income_date BETWEEN ? AND ?", startDate, endDate)
		dataQuery = dataQuery.Where("ex.income_date BETWEEN ? AND ?", startDate, endDate)
	}

	// Pertama, hitung total catatan yang sesuai dengan filter
	if err := countQuery.Count(&total).Error; err != nil {
		return responses.InternalServerError(c, "Failed to count another income", err)
	}

	// Kemudian, ambil data yang dipaginasi dengan pengurutan
	if err := dataQuery.Order("ex.created_at DESC").Limit(limit).Offset(offset).Find(&AnotherIncome).Error; err != nil {
		return responses.InternalServerError(c, "Failed to get another income data", err)
	}

	// Buat slice baru untuk menampung data yang sudah diformat
	var formattedAnotherIncomesData []models.AnotherIncomeDetailResponse
	for _, anothIn := range AnotherIncome {
		formattedAnotherIncomesData = append(formattedAnotherIncomesData, models.AnotherIncomeDetailResponse{
			ID:          anothIn.ID,
			Description: anothIn.Description,
			IncomeDate:  utils.FormatIndonesianDate(anothIn.IncomeDate), // Format tanggal di sini
			TotalIncome: anothIn.TotalIncome,
			Payment:     string(anothIn.Payment),
		})
	}

	// Hitung total halaman berdasarkan hasil filter
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return responses.JSONResponseGetAll(c, http.StatusOK, "Another Incomes retrieved successfully", search, int(total), page, int(totalPages), int(limit), formattedAnotherIncomesData)
}

// Insert atau update laporan transaksi berdasarkan Another Income / Pendapatan Lain
func SyncAnotherIncomeReport(db *gorm.DB, anotherIncome models.AnotherIncomes) error {

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(utils.Location)

	// Siapkan data report dari AnotherIncome
	report := models.TransactionReports{
		ID:              anotherIncome.ID,
		TransactionType: models.Income,
		UserID:          anotherIncome.UserID,
		BranchID:        anotherIncome.BranchID,
		Total:           anotherIncome.TotalIncome,
		CreatedAt:       anotherIncome.CreatedAt,
		UpdatedAt:       anotherIncome.UpdatedAt,
		Payment:         anotherIncome.Payment,
	}

	var existing models.TransactionReports
	err := db.Take(&existing, "id = ?", report.ID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Insert
		return db.Create(&report).Error
	}
	if err != nil {
		return err
	}

	// Jika ditemukan, lakukan update pada kolom yang dibutuhkan
	existing.Total = report.Total
	existing.UpdatedAt = nowWIB
	existing.Payment = report.Payment

	return db.Save(&existing).Error
}
