package controllers

import (
	"math"
	"net/http"
	"strconv"
	strings "strings"

	"github.com/heru-oktafian/api-retail/models"
	"github.com/heru-oktafian/scafold/config"
	"github.com/heru-oktafian/scafold/framework"
	"github.com/heru-oktafian/scafold/helpers"
	"github.com/heru-oktafian/scafold/middlewares"
	"github.com/heru-oktafian/scafold/responses"
)

// CreateSupplier buat Supplier
func CreateSupplier(c *framework.Ctx) error {
	// Get branch id
	branch_id, _ := middlewares.GetBranchID(c.Request)

	// Creating new Supplier using helpers
	return helpers.CreateResource(c, config.DB, &models.Supplier{}, branch_id, "SPL")
}

// UpdateSupplier update Supplier
func UpdateSupplier(c *framework.Ctx) error {
	id := c.Param("id")
	// Updating Supplier using helpers
	return helpers.UpdateResource(c, config.DB, &models.Supplier{}, id)
}

// DeleteSupplier hapus Supplier
func DeleteSupplier(c *framework.Ctx) error {
	id := c.Param("id")
	// Deleting Supplier using helpers
	return helpers.DeleteResource(c, config.DB, &models.Supplier{}, id)
}

// GetSupplierByID tampilkan Supplier berdasarkan id
func GetSupplierByID(c *framework.Ctx) error {
	id := c.Param("id")
	// Getting Supplier using helpers
	return helpers.GetResource(c, config.DB, &models.Supplier{}, id)
}

// GetAllSuppliers tampilkan semua Supplier
func GetAllSupplier(c *framework.Ctx) error {
	// Get branch id
	branch_id, _ := middlewares.GetBranchID(c.Request)

	// Ambil parameter page dan search dari query URL
	pageParam := c.Query("page")
	search := strings.TrimSpace(c.Query("search"))

	// Konversi page ke int, default ke 1 jika tidak valid
	page := 1
	if p, err := strconv.Atoi(pageParam); err == nil && p > 0 {
		page = p
	}

	limit := 10                  // Tetapkan limit ke 10 data per halaman
	offset := (page - 1) * limit // Hitung offset untuk pagination

	var Supplier []models.SupplierDetail
	var total int64

	// Query dasar
	query := config.DB.Table("suppliers s").
		Select("s.id, s.name, s.phone, s.address, s.pic, s.supplier_category_id, sc.name AS supplier_category").
		Joins("LEFT JOIN supplier_categories sc ON sc.id = s.supplier_category_id").
		Where("s.branch_id = ?", branch_id)

	// Jika ada search key, tambahkan filter WHERE
	if search != "" {
		search = strings.ToLower(search) // Konversi search ke lowercase
		query = query.Where("LOWER(s.name) LIKE ?", "%"+search+"%")
	}

	// Hitung total supplier yang sesuai dengan filter
	if err := query.Count(&total).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Get Suppliers failed", "Failed to count Suppliers")
	}

	// Ambil data dengan pagination
	if err := query.Offset(offset).Limit(limit).Scan(&Supplier).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Get Suppliers failed", "Failed to fetch Suppliers")
	}

	// Hitung total halaman berdasarkan hasil filter
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	// Mengembalikan response data Supplier
	return responses.JSONResponseGetAll(c, http.StatusOK, "Suppliers retrieved successfully", search, int(total), page, int(totalPages), int(limit), Supplier)

}

// CmbSupplierCategory mendapatkan semua kategori supplier
func CmbSupplierCategory(c *framework.Ctx) error {
	// Get branch id
	branch_id, _ := middlewares.GetBranchID(c.Request)

	var cmbSupplierCategories []models.SupplierCategoryCombo

	// Query untuk mendapatkan semua kategori supplier
	if err := config.DB.Table("supplier_categories").
		Select("id AS supplier_category_id, name AS supplier_category_name").
		Where("branch_id = ?", branch_id).
		Order("name ASC").
		Find(&cmbSupplierCategories).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Failed to get data", "Failed to get data")
	}

	return responses.JSONResponse(c, http.StatusOK, "Data berhasil ditemukan", cmbSupplierCategories)
}

// CmbSupplier mendapatkan semua supplier untuk combo box
func CmbSupplier(c *framework.Ctx) error {
	// Get branch id
	branch_id, _ := middlewares.GetBranchID(c.Request)

	// Ambil parameter search dari query URL
	search := strings.ToLower(c.Query("search"))

	var cmbSuppliers []models.CmbSupplierModel

	// Query untuk mendapatkan semua supplier
	query := config.DB.Table("suppliers").
		Select("id AS supplier_id, name AS supplier_name").
		Where("branch_id = ?", branch_id)
	// Jika ada search key, tambahkan filter WHERE
	if search != "" {
		query = query.Where("LOWER(name) LIKE ?", "%"+search+"%")
	}

	// Urutkan hasil secara ascending berdasarkan name
	query = query.Order("name ASC")

	if err := query.Find(&cmbSuppliers).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Failed to get data", "Failed to get data")
	}

	return responses.JSONResponse(c, http.StatusOK, "Data berhasil ditemukan", cmbSuppliers)
}
