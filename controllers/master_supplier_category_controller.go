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

// CreateSupplierCategory buat Supplier category
func CreateSupplierCategory(c *framework.Ctx) error {
	// Get branch id
	branch_id, _ := middlewares.GetBranchID(c.Request)

	// Creating new SupplierCategory using helpers
	return helpers.CreateResourceInc(c, config.DB, branch_id, &models.SupplierCategory{})
}

// UpdateSupplierCategory update SupplierCategory
func UpdateSupplierCategory(c *framework.Ctx) error {
	id := c.Param("id")
	// Updating SupplierCategory using helpers
	return helpers.UpdateResource(c, config.DB, &models.SupplierCategory{}, id)
}

// DeleteSupplierCategory hapus SupplierCategory
func DeleteSupplierCategory(c *framework.Ctx) error {
	id := c.Param("id")
	// Deleting SupplierCategory using helpers
	return helpers.DeleteResource(c, config.DB, &models.SupplierCategory{}, id)
}

// GetSupplierCategoryByID tampilkan SupplierCategory berdasarkan id
func GetSupplierCategoryByID(c *framework.Ctx) error {
	id := c.Param("id")
	// Getting SupplierCategory using helpers
	return helpers.GetResource(c, config.DB, &models.SupplierCategory{}, id)
}

// GetSupplierCategory tampilkan SupplierCategory
func GetAllSupplierCategory(c *framework.Ctx) error {
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

	var SupplierCategory []models.SupplierCategory
	var total int64

	// Query dasar
	query := config.DB.Table("supplier_categories sc").
		Select("sc.id, sc.name, sc.branch_id").
		Where("sc.branch_id = ?", branch_id)

	// Jika ada search key, tambahkan filter WHERE
	if search != "" {
		search = strings.ToLower(search) // Konversi search ke lowercase
		query = query.Where("LOWER(sc.name) LIKE ?", "%"+search+"%")
	}

	// Hitung total data yang sesuai dengan filter
	if err := query.Count(&total).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Get data failed", "Failed to count data")
	}

	// Ambil data dengan pagination dan urutkan hasil
	if err := query.Order("sc.name ASC").Offset(offset).Limit(limit).Scan(&SupplierCategory).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Get data failed", "Failed to fetch data")
	}

	// Hitung total halaman berdasarkan hasil filter
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return responses.JSONResponseGetAll(c, http.StatusOK, "Supplier Category retrieved successfully", search, int(total), page, int(totalPages), int(limit), SupplierCategory)
}
