package controllers

import (
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
)

// CreateUnit buat unit
func CreateUnit(c *framework.Ctx) error {
	branch_id, _ := middlewares.GetBranchID(c.Request)

	// Creating new unit using helpers
	return helpers.CreateResource(c, config.DB, &models.Unit{}, branch_id, "UNT")
}

// UpdateUnit update unit
func UpdateUnit(c *framework.Ctx) error {
	id := c.Param("id")
	// Updating unit using helpers
	return helpers.UpdateResource(c, config.DB, &models.Unit{}, id)
}

// DeleteUnit hapus unit
func DeleteUnit(c *framework.Ctx) error {
	id := c.Param("id")
	// Deleting unit using helpers
	return helpers.DeleteResource(c, config.DB, &models.Unit{}, id)
}

// GetUnitByID tampilkan unit berdasarkan id
func GetUnitByID(c *framework.Ctx) error {
	id := c.Param("id")

	// Getting unit using helpers
	return helpers.GetResource(c, config.DB, &models.Unit{}, id)
}

// GetAllUnit tampilkan semua unit
func GetAllUnit(c *framework.Ctx) error {
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

	limit := 10 // Tetapkan batas data per halaman ke 10
	offset := (page - 1) * limit

	var Unit []models.AllUnit // Gunakan AllUnit untuk mengambil data unit tanpa branch_id
	var total int64

	// Query dasar
	query := config.DB.Table("units un").Select("un.id AS unit_id, un.name AS unit_name").Where("un.branch_id = ?", branch_id)

	// Jika ada kata kunci pencarian, tambahkan filter WHERE
	if search != "" {
		search = strings.ToLower(search) // Konversi kata kunci pencarian ke huruf kecil
		query = query.Where("LOWER(un.name) LIKE ?", "%"+search+"%")
	}

	// Hitung total unit yang sesuai dengan filter
	if err := query.Count(&total).Error; err != nil {
		return responses.InternalServerError(c, "Gagal mengambil data Unit", err)
	}

	// Ambil data dengan pagination
	if err := query.Offset(offset).Limit(limit).Scan(&Unit).Error; err != nil {
		return responses.InternalServerError(c, "Gagal mengambil data Unit", err)
	}

	// Hitung total halaman berdasarkan hasil filter
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return responses.JSONResponseGetAll(c, http.StatusOK, "Data Unit berhasil diambil", search, int(total), page, int(totalPages), int(limit), Unit)
}

// CmbUnit mendapatkan semua kategori unit
func CmbUnit(c *framework.Ctx) error {
	// Get branch id
	branch_id, _ := middlewares.GetBranchID(c.Request)

	// Parse query parameters for search
	search := strings.TrimSpace(c.Query("search"))

	var cmbUnits []models.UnitCombo

	// Base query to get all unit categories
	query := config.DB.Table("units").
		Select("id as unit_id, name as unit_name").
		Where("branch_id = ?", branch_id)

	// If search parameter is provided, add a filter
	if search != "" {
		search = strings.ToLower(search) // Convert search keyword to lowercase
		query = query.Where("LOWER(name) LIKE ?", "%"+search+"%")
	}

	// Add order by name ascending
	query = query.Order("name ASC")

	// Execute the query
	if err := query.Find(&cmbUnits).Error; err != nil {
		return responses.InternalServerError(c, "Failed to get data", err)
	}

	return responses.JSONResponse(c, http.StatusOK, "Data berhasil ditemukan", cmbUnits)
}
