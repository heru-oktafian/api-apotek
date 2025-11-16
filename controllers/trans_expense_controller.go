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

// CreateExpense Function
func CreateExpense(c *framework.Ctx) error {

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(utils.Location)

	db := config.DB

	// Ambil informasi dari token
	branchID, _ := middlewares.GetBranchID(c.Request)
	userID, _ := middlewares.GetUserID(c.Request)
	generatedID := helpers.GenerateID("EXP")

	// Ambil input dari body
	var input models.ExpenseInput
	if err := c.BodyParser(&input); err != nil {
		return responses.BadRequest(c, "Invalid input", err)
	}

	// Parse tanggal
	layout := "2006-01-02" // format harus YYYY-MM-DD
	parsedDate, err := time.Parse(layout, input.ExpenseDate)
	description := input.Description
	payment := input.Payment
	total := input.TotalExpense
	if err != nil {
		return responses.BadRequest(c, "Invalid date format. Use YYYY-MM-DD", err)
	}

	// Map ke struct model
	expense := models.Expenses{
		ID:           generatedID,
		Description:  description,
		BranchID:     branchID,
		UserID:       userID,
		ExpenseDate:  parsedDate,
		TotalExpense: total,
		Payment:      models.PaymentStatus(payment),
		CreatedAt:    nowWIB,
		UpdatedAt:    nowWIB,
	}

	// Simpan expense
	if err := db.Create(&expense).Error; err != nil {
		return responses.InternalServerError(c, "Failed to create Expense", err)
	}

	// Buat laporan
	if err := SyncExpenseReport(db, expense); err != nil {
		return responses.InternalServerError(c, "Failed to create Expense Report", err)
	}

	return responses.JSONResponse(c, http.StatusOK, "Expense created successfully", expense)
}

// UpdateExpenseItem Function
func UpdateExpense(c *framework.Ctx) error {

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(utils.Location)

	db := config.DB
	id := c.Param("id")

	// Cari data expense
	var expense models.Expenses
	if err := db.First(&expense, "id = ?", id).Error; err != nil {
		return responses.NotFound(c, "Expense not found")
	}

	// Gunakan struct khusus input
	var input models.ExpenseInput
	if err := c.BodyParser(&input); err != nil {
		return responses.BadRequest(c, "Invalid input", err)
	}

	// Parse tanggal dari string ke time.Time
	layout := "2006-01-02"
	parsedDate, err := time.Parse(layout, input.ExpenseDate)
	if err != nil {
		return responses.BadRequest(c, "Invalid date format. Use YYYY-MM-DD", err)
	}

	// Update field dasar
	expense.ExpenseDate = parsedDate
	expense.Description = input.Description
	expense.TotalExpense = input.TotalExpense
	expense.Payment = models.PaymentStatus(input.Payment)
	expense.UpdatedAt = nowWIB

	// Simpan update
	if err := db.Save(&expense).Error; err != nil {
		return responses.InternalServerError(c, "Failed to update Expense", err)
	}

	// Sync report
	if err := SyncExpenseReport(db, expense); err != nil {
		return responses.InternalServerError(c, "Failed to sync Expense Report", err)
	}

	return responses.JSONResponse(c, http.StatusOK, "Expense updated successfully", expense)
}

// DeleteExpenseItem Function
func DeleteExpense(c *framework.Ctx) error {
	db := config.DB
	id := c.Param("id")

	// Ambil expense
	var expense models.Expenses
	if err := db.First(&expense, "id = ?", id).Error; err != nil {
		return responses.NotFound(c, "Expense not found")
	}

	// Hapus laporan
	if err := db.Where("id = ? AND transaction_type = ?", expense.ID, models.Expense).Delete(&models.TransactionReports{}).Error; err != nil {
		return responses.InternalServerError(c, "Failed to delete Transaction Report", err)
	}

	// Hapus expense
	if err := db.Delete(&expense).Error; err != nil {
		return responses.InternalServerError(c, "Failed to delete Expense", err)
	}

	return responses.JSONResponse(c, http.StatusOK, "Expense deleted successfully", expense)
}

// GetAllExpenses tampilkan semua Expense
func GetAllExpenses(c *framework.Ctx) error {
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

	var expenses []models.Expenses // Mengganti nama menjadi 'expenses' untuk kejelasan
	var total int64

	// Buat builder kueri yang bersih untuk menghitung dan mengambil data
	countQuery := config.DB.Table("expenses ex").
		Where("ex.branch_id = ?", branchID)

	dataQuery := config.DB.Table("expenses ex").
		Select("ex.id, ex.description, ex.expense_date, ex.total_expense, ex.payment").
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
			return responses.BadRequest(c, "Invalid month format. Month should be in format YYYY-MM", err)
		}
		startDate := parsedMonth
		endDate := startDate.AddDate(0, 1, 0).Add(-time.Nanosecond)
		countQuery = countQuery.Where("ex.expense_date BETWEEN ? AND ?", startDate, endDate)
		dataQuery = dataQuery.Where("ex.expense_date BETWEEN ? AND ?", startDate, endDate)
	}

	// Pertama, hitung total catatan yang sesuai dengan filter
	if err := countQuery.Count(&total).Error; err != nil {
		return responses.InternalServerError(c, "Failed to count expenses", err)
	}

	// Kemudian, ambil data yang dipaginasi dengan pengurutan
	if err := dataQuery.Order("ex.created_at DESC").Limit(limit).Offset(offset).Find(&expenses).Error; err != nil {
		return responses.InternalServerError(c, "Failed to get expenses data", err)
	}

	// Format data pengeluaran yang diambil
	var formattedExpenseData []models.ExpenseDetailResponse
	for _, expense := range expenses { // Iterasi melalui 'expenses'
		formattedExpenseData = append(formattedExpenseData, models.ExpenseDetailResponse{
			ID:           expense.ID,
			Description:  expense.Description,
			ExpenseDate:  utils.FormatIndonesianDate(expense.ExpenseDate), // Format tanggal di sini
			TotalExpense: expense.TotalExpense,
			Payment:      string(expense.Payment),
		})
	}

	// Hitung total halaman
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return responses.JSONResponseGetAll(c, http.StatusOK, "Expenses retrieved successfully", search, int(total), page, totalPages, limit, formattedExpenseData)
}

// Insert atau update laporan transaksi berdasarkan Expenses / Pengeluaran
func SyncExpenseReport(db *gorm.DB, expense models.Expenses) error {

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(utils.Location)

	// Siapkan data report dari Expense
	report := models.TransactionReports{
		ID:              expense.ID,
		TransactionType: models.Expense,
		UserID:          expense.UserID,
		BranchID:        expense.BranchID,
		Total:           expense.TotalExpense,
		CreatedAt:       expense.CreatedAt,
		UpdatedAt:       expense.UpdatedAt,
		Payment:         expense.Payment,
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
