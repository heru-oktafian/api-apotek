package controllers

import (
	fmt "fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/heru-oktafian/api-retail/models"
	"github.com/heru-oktafian/scafold/config"
	"github.com/heru-oktafian/scafold/framework"
	"github.com/heru-oktafian/scafold/helpers"
	"github.com/heru-oktafian/scafold/middlewares"
	"github.com/heru-oktafian/scafold/responses"
	"gorm.io/gorm"
)

// CreateUnitConversion controller
// Endpoint: POST /api/unit-conversions
func CreateUnitConversion(c *framework.Ctx) error {
	db := config.DB
	var req models.UnitConversionRequest

	// Parsing request body
	if err := c.BodyParser(&req); err != nil {
		return responses.BadRequest(c, "Format data yang dikirim tidak valid", err)
	}

	// Mendapatkan BranchID dari token (asumsi UnitConversion spesifik per cabang)
	branchID, _ := middlewares.GetBranchID(c.Request)
	if branchID == "" {
		return responses.NotFound(c, "Branch ID not found in token. Unauthorized")
	}

	// --- VALIDASI INPUT ---
	// if err := helpers.ValidateStruct(req); err != nil {
	// 	return responses.BadRequest(c, "Validation failed", err)
	// }
	// --- AKHIR VALIDASI INPUT ---

	// Mulai transaksi database
	tx := db.Begin()
	if tx.Error != nil {
		return responses.InternalServerError(c, "Failed to begin database transaction", tx.Error)
	}
	// Pastikan transaksi di-rollback jika terjadi kesalahan atau panic
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// --- Pengecekan Duplikasi ---
	var existingConversion models.UnitConversion
	checkErr := tx.Where("product_id = ? AND init_id = ? AND final_id = ? AND branch_id = ?",
		req.ProductId,
		req.InitId,
		req.FinalId,
		branchID,
	).First(&existingConversion).Error

	if checkErr == nil {
		// Jika record sudah ditemukan (checkErr == nil), berarti duplikat
		tx.Rollback()
		return responses.Conflict(c, fmt.Errorf("unit conversion from '%s' to '%s' for product '%s' already exists in this branch: duplicate entry",
			req.InitId, req.FinalId, req.ProductId))
	} else if checkErr != gorm.ErrRecordNotFound {
		// Jika ada error lain selain record not found
		tx.Rollback()
		return responses.InternalServerError(c, "Failed to check for existing unit conversion", checkErr)
		// "error":   checkErr.Error(),
		// })
	}
	// Jika checkErr == gorm.ErrRecordNotFound, berarti tidak ada duplikat, lanjutkan proses

	// --- 1. Dapatkan Product untuk memastikan ProductId valid ---
	var product models.Product
	if err := tx.Where("id = ? AND branch_id = ?", req.ProductId, branchID).First(&product).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return responses.NotFound(c, fmt.Sprintf("Product with ID %s not found in branch %s.", req.ProductId, branchID))
		}
		return responses.InternalServerError(c, "Failed to retrieve product for validation", err)
	}

	// --- 2. Dapatkan Init Unit untuk memastikan InitId valid ---
	var initUnit models.Unit
	if err := tx.Where("id = ? AND branch_id = ?", req.InitId, branchID).First(&initUnit).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return responses.NotFound(c, fmt.Sprintf("Initial unit (InitId) with ID %s not found in branch %s.", req.InitId, branchID))
		}
		return responses.InternalServerError(c, "Failed to retrieve initial unit for validation", err)
	}

	// --- 3. Dapatkan Final Unit untuk memastikan FinalId valid ---
	var finalUnit models.Unit
	if err := tx.Where("id = ? AND branch_id = ?", req.FinalId, branchID).First(&finalUnit).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return responses.NotFound(c, fmt.Sprintf("Final unit (FinalId) with ID %s not found in branch %s.", req.FinalId, branchID))
		}
		return responses.InternalServerError(c, "Failed to retrieve final unit for validation", err)
	}

	// --- Buat objek UnitConversion baru ---
	unitConversion := models.UnitConversion{
		ID:        helpers.GenerateID("UNC"), // Generate ID untuk Unit Conversion
		ProductId: req.ProductId,
		InitId:    req.InitId,
		FinalId:   req.FinalId,
		ValueConv: req.ValueConv,
		BranchID:  branchID, // Set BranchID dari token
	}

	// Simpan UnitConversion ke database
	if err := tx.Create(&unitConversion).Error; err != nil {
		tx.Rollback()
		return responses.InternalServerError(c, "Failed to create unit conversion", err)
	}

	// Commit transaksi jika semua berhasil
	if err := tx.Commit().Error; err != nil {
		return responses.InternalServerError(c, "Failed to commit database transaction", err)
	}

	// Berhasil
	return responses.JSONResponse(c, http.StatusCreated, "Unit conversion created successfully", framework.Map{
		"id":              unitConversion.ID,
		"product_id":      unitConversion.ProductId,
		"init_id":         unitConversion.InitId,
		"init_unit_name":  initUnit.Name, // Menambahkan nama unit
		"final_id":        unitConversion.FinalId,
		"final_unit_name": finalUnit.Name, // Menambahkan nama unit
		"value_conv":      unitConversion.ValueConv,
		"branch_id":       unitConversion.BranchID,
	})
}

// UpdateUnit update unit
func UpdateUnitConversion(c *framework.Ctx) error {
	id := c.Param("id")
	// Updating unit using helpers
	return helpers.UpdateResource(c, config.DB, &models.UnitConversion{}, id)
}

// DeleteUnit hapus unit
func DeleteUnitConversion(c *framework.Ctx) error {
	id := c.Param("id")
	// Deleting unit using helpers
	return helpers.DeleteResource(c, config.DB, &models.UnitConversion{}, id)
}

// GetUnitConversionByID tampilkan unit berdasarkan id
func GetUnitConversionByID(c *framework.Ctx) error {
	id := c.Param("id")
	// Getting unit using helpers
	return helpers.GetResource(c, config.DB, &models.UnitConversion{}, id)
}

// GetAllUnit tampilkan semua unit
func GetAllUnitConversion(c *framework.Ctx) error {
	// Ambil ID cabang
	branch_id, _ := middlewares.GetBranchID(c.Request)

	// Ambil parameter page dan search dari query URL
	pageParam := c.Query("page")
	search := strings.TrimSpace(c.Query("search"))

	// Konversi page ke int, default ke 1 jika tidak valid
	page := 1
	if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
		page = p
	}

	limit := 10                  // Tetapkan batas data per halaman ke 10
	offset := (page - 1) * limit // Hitung offset berdasarkan halaman dan batas

	var unit_conversions []models.UnitConversionDetail
	var total int64

	// Query dasar
	query := config.DB.Table("unit_conversions unc").
		Select("unc.id, pro.name AS product_name, uin.name AS init_name, ufi.name AS final_name, unc.value_conv, unc.product_id, unc.init_id, unc.final_id, unc.branch_id").
		Joins("LEFT JOIN products pro on pro.id = unc.product_id").
		Joins("LEFT JOIN units uin on uin.id = unc.init_id").
		Joins("LEFT JOIN units ufi on ufi.id = unc.final_id").
		Where("unc.branch_id = ?", branch_id)

	// Jika ada search key, tambahkan filter WHERE
	if search != "" {
		search = strings.ToLower(search) // Konversi search ke lowercase
		query = query.Where("LOWER(pro.name) LIKE ? OR LOWER(uin.name) LIKE ? OR LOWER(ufi.name) LIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// Hitung total unit conversion yang sesuai dengan filter
	if err := query.Count(&total).Error; err != nil {
		return responses.JSONResponse(c, http.StatusNotFound, "Get unit conversions Failed", "Failed to count Unit Conversions")
	}

	// Ambil data unit conversion dengan paginasi
	if err := query.Offset(offset).Limit(limit).Scan(&unit_conversions).Error; err != nil {
		return responses.JSONResponse(c, http.StatusNotFound, "Get unit conversions Failed", "Failed to get Unit Conversions")
	}

	// Hitung total halaman
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	// Return response
	return responses.JSONResponseGetAll(c, http.StatusOK, "Unit conversions retrieved successfully", search, int(total), page, int(totalPages), int(limit), unit_conversions)

}

// CmbProdConv mendapatkan semua produk
func CmbProdConv(c *framework.Ctx) error {
	// Get branch id
	branch_id, _ := middlewares.GetBranchID(c.Request)

	// Parsing query parameter "search"
	search := strings.TrimSpace(c.Query("search"))

	var cmbProducts []models.ProdConvCombo

	// Query untuk mendapatkan semua produk
	query := config.DB.Table("products").
		Select("id as product_id, name as product_name").
		Where("branch_id = ?", branch_id)

	// Jika ada parameter search, tambahkan filter WHERE
	if search != "" {
		search = strings.ToLower(search) // Konversi search ke lowercase
		query = query.Where("LOWER(name) LIKE ?", "%"+search+"%")
	}

	if err := query.Find(&cmbProducts).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Failed to get data", "Failed to get data")
	}

	return responses.JSONResponse(c, http.StatusOK, "Data berhasil ditemukan", cmbProducts)
}
