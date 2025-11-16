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

// CreateProductCategory buat product category
func CreateProductCategory(c *framework.Ctx) error {
	// Get branch id
	branch_id, _ := middlewares.GetBranchID(c.Request)
	// Creating new ProductCategory using helpers
	return helpers.CreateResourceInc(c, config.DB, branch_id, &models.ProductCategory{})
}

// UpdateProductCategory update ProductCategory
func UpdateProductCategory(c *framework.Ctx) error {
	id := c.Param("id")
	// Updating ProductCategory using helpers
	return helpers.UpdateResource(c, config.DB, &models.ProductCategory{}, id)
}

// DeleteProductCategory hapus ProductCategory
func DeleteProductCategory(c *framework.Ctx) error {
	id := c.Param("id")
	// Deleting ProductCategory using helpers
	return helpers.DeleteResource(c, config.DB, &models.ProductCategory{}, id)
}

// GetProductCategory tampilkan ProductCategory berdasarkan id
func GetProductCategory(c *framework.Ctx) error {
	id := c.Param("id")
	// Getting ProductCategory using helpers
	return helpers.GetResource(c, config.DB, &models.ProductCategory{}, id)
}

// GetAllProductCategory tampilkan semua ProductCategory
func GetAllProductCategory(c *framework.Ctx) error {
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

	limit := 10 // Tetapkan limit ke 10 data per halaman
	offset := (page - 1) * limit

	var ProductCategory []models.ComboProductCategory
	var total int64

	// Query dasar
	query := config.DB.Table("product_categories pc").Select("pc.id AS product_category_id, pc.name AS product_category_name").Where("pc.branch_id = ?", branch_id)

	// Jika ada search key, tambahkan filter WHERE
	if search != "" {
		search = strings.ToLower(search) // Konversi search ke lowercase
		query = query.Where("LOWER(pc.name) LIKE ?", "%"+search+"%")
	}

	// Hitung total data yang sesuai dengan filter
	if err := query.Count(&total).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Get data failed", "Failed to count data")
	}

	// Ambil data dengan pagination
	if err := query.Offset(offset).Limit(limit).Scan(&ProductCategory).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Get data failed", "Failed to fetch data")
	}

	// Hitung total halaman berdasarkan hasil filter
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return responses.JSONResponseGetAll(c, http.StatusOK, "Product Categories retrieved successfully", search, int(total), page, int(totalPages), int(limit), ProductCategory)
}

// CmbProductCategory mendapatkan semua kategori produk
func CmbProductCategory(c *framework.Ctx) error {
	// Get branch id
	branch_id, _ := middlewares.GetBranchID(c.Request)

	// Parsing query parameter "search"
	search := strings.TrimSpace(c.Query("search"))

	var categories []models.ComboProductCategory

	// Query untuk mendapatkan semua kategori produk
	query := config.DB.Table("product_categories").
		Select("product_categories.id as product_category_id, product_categories.name as product_category_name").
		Where("branch_id = ?", branch_id)

	// Jika ada search key, tambahkan filter WHERE
	if search != "" {
		search = strings.ToLower(search) // Konversi search ke lowercase
		query = query.Where("LOWER(product_categories.name) LIKE ?", "%"+search+"%")
	}

	// Tambahkan urutan ascending berdasarkan nama
	query = query.Order("product_categories.name ASC")

	// Eksekusi query
	if err := query.Find(&categories).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Failed to get data", "Failed to get data")
	}

	return responses.JSONResponse(c, http.StatusOK, "Data berhasil ditemukan", categories)
}
