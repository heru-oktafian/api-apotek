package controllers

import (
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

// CreateBranch is function for create new branch
func CreateBranch(c *framework.Ctx) error {
	// Get branch id
	branch_id, _ := middlewares.GetBranchID(c.Request)

	// Creating new unit using helpers
	return helpers.CreateResource(c, config.DB, &models.Branch{}, branch_id, "BRC")
}

// UpdateBranch is function for update branch
func UpdateBranch(c *framework.Ctx) error {
	id := c.Param("id")
	// Updating branch using helpers
	return helpers.UpdateResource(c, config.DB, &models.Branch{}, id)
}

// DeleteBranch is function for delete branch
func DeleteBranch(c *framework.Ctx) error {
	id := c.Param("id")
	// Deleting branch using helpers
	return helpers.DeleteResource(c, config.DB, &models.Branch{}, id)
}

// GetBranch is function for get branch
func GetBranch(c *framework.Ctx) error {
	id := c.Param("id")
	// Getting branch using helpers
	return helpers.GetResource(c, config.DB, &models.Branch{}, id)
}

// GetAllBranch is function for get all branch
func GetAllBranch(c *framework.Ctx) error {
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

	// Query builder untuk mengambil data branch
	db := config.DB.Model(&models.Branch{})

	// Pencarian berdasarkan branch_name, address, phone, email, owner_name, bank_name atau account_name
	if search != "" {
		// search = strings.ToLower(search) // Konversi search ke lowercase
		db = db.Where("branch_name LIKE ? OR address LIKE ? OR phone LIKE ? OR email LIKE ? OR owner_name LIKE ? OR bank_name LIKE ? OR account_name LIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	var branches []models.Branch
	var total int64

	// Menghitung total data branch
	if err := db.Count(&total).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Error", "Gagal menghitung data branch")
	}

	// Mengambil data branch sesuai paginasi
	if err := db.Limit(limit).Offset(offset).Find(&branches).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Error", "Gagal mengambil data branch")
	}

	// Mengembalikan response JSON dengan data branch
	return responses.JSONResponseGetAll(c, http.StatusOK, "Data berhasil ditemukan", search, int(total), page, int((total+int64(limit)-1)/int64(limit)), limit, branches)
}
