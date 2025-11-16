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

// CreateMemberCategory buat member category
func CreateMemberCategory(c *framework.Ctx) error {
	// Get branch id
	branch_id, _ := middlewares.GetBranchID(c.Request)

	// Creating new MemberCategory using helpers
	return helpers.CreateResourceInc(c, config.DB, branch_id, &models.MemberCategory{})
}

// UpdateMemberCategory update MemberCategory
func UpdateMemberCategory(c *framework.Ctx) error {
	id := c.Param("id")
	// Updating MemberCategory using helpers
	return helpers.UpdateResource(c, config.DB, &models.MemberCategory{}, id)
}

// DeleteMemberCategory hapus MemberCategory
func DeleteMemberCategory(c *framework.Ctx) error {
	id := c.Param("id")
	// Deleting MemberCategory using helpers
	return helpers.DeleteResource(c, config.DB, &models.MemberCategory{}, id)
}

// GetMemberCategory tampilkan MemberCategory berdasarkan id
func GetMemberCategory(c *framework.Ctx) error {
	id := c.Param("id")
	// Getting MemberCategory using helpers
	return helpers.GetResource(c, config.DB, &models.MemberCategory{}, id)
}

// GetAllMemberCategories tampilkan semua MemberCategory
func GetAllMemberCategory(c *framework.Ctx) error {
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

	var MemberCategory []models.MemberCategory
	var total int64

	// Query dasar
	query := config.DB.Table("member_categories mc").Select("mc.id, mc.name, mc.points_conversion_rate, mc.branch_id").Where("mc.branch_id = ?", branch_id)

	// Jika ada search key, tambahkan filter WHERE
	if search != "" {
		search = strings.ToLower(search) // Konversi search ke lowercase
		query = query.Where("LOWER(mc.name) LIKE ?", "%"+search+"%")
	}

	// Hitung total data
	if err := query.Count(&total).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Failed to get total data", "Failed to get total data")
	}

	// Ambil data sesuai limit dan offset
	if err := query.Offset(offset).Limit(limit).Find(&MemberCategory).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Failed to get data", "Failed to get data")
	}

	// Hitung total halaman berdasarkan hasil filter
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return responses.JSONResponseGetAll(c, http.StatusOK, "Member Categories retrieved successfully", search, int(total), page, int(totalPages), int(limit), MemberCategory)

}
