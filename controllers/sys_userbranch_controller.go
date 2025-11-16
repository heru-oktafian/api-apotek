package controllers

import (
	"net/http"

	"github.com/heru-oktafian/api-retail/models"
	"github.com/heru-oktafian/scafold/config"
	"github.com/heru-oktafian/scafold/framework"
	"github.com/heru-oktafian/scafold/middlewares"
	"github.com/heru-oktafian/scafold/responses"
)

// CreateUserBranch menangani penambahan userbranch
func CreateUserBranch(c *framework.Ctx) error {
	// Buat instance baru untuk UserBranch
	var userbranch models.UserBranch

	// Parse input JSON menjadi struct UserBranch
	if err := c.BodyParser(&userbranch); err != nil {
		return responses.JSONResponse(c, http.StatusBadRequest, "Invalid input", err)
	}

	// Simpan user ke database
	if err := config.DB.Create(&userbranch).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Failed to create user", err)
	}
	return responses.JSONResponse(c, http.StatusOK, "UserBranch created successfully", userbranch)
}

// GetUserBranch menangani penampilan userbranch
func GetUserBranch(c *framework.Ctx) error {
	userID := c.Param("user_id")
	branch_id, _ := middlewares.GetBranchID(c.Request)
	var userBranchDetails []models.UserBranchDetail

	// Melakukan LEFT OUTER JOIN menggunakan GORM
	if err := config.DB.
		Table("user_branches").
		Select("user_branches.user_id, users.name AS user_name, user_branches.branch_id, branches.branch_name, branches.address, branches.phone").
		Joins("LEFT JOIN users ON users.id = user_branches.user_id").
		Joins("LEFT JOIN branches ON branches.id = user_branches.branch_id").
		Where("branches.branch_status = 'active' AND user_branches.branch_id = ? AND user_branches.user_id = ?", branch_id, userID).
		Scan(&userBranchDetails).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Get userbranches failed", "Failed to fetch user branches with details")
	}

	// Mengembalikan response data userbranch
	return responses.JSONResponse(c, http.StatusOK, "UserBranch found", userBranchDetails)
}

// GetUserBranch menangani penampilan userbranch
func GetBranchByUserId(c *framework.Ctx) error {
	// get user id
	userID := c.Param("user_id")

	// Menampilkan semua userbranch
	var userBranchDetails []models.UserBranchDetail

	// Melakukan LEFT OUTER JOIN menggunakan GORM
	if err := config.DB.
		Table("user_branches").
		Select("user_branches.user_id, users.name AS user_name, user_branches.branch_id, branches.branch_name, branches.sia_name, branches.sipa_name, branches.phone").
		Joins("LEFT JOIN users ON users.id = user_branches.user_id").
		Joins("LEFT JOIN branches ON branches.id = user_branches.branch_id").
		Where("branches.branch_status = 'active' AND user_branches.user_id = ?", userID).
		Scan(&userBranchDetails).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Get userbranches failed", "Failed to fetch user branches with details")
	}

	// Mengembalikan response data userbranch
	return responses.JSONResponse(c, http.StatusOK, "UserBranch found", userBranchDetails)
}

// UpdateUserBranch menangani pembaruan userbranch
func UpdateUserBranch(c *framework.Ctx) error {
	userID := c.Param("user_id")
	branch_id, _ := middlewares.GetBranchID(c.Request)

	var userbranch models.UserBranch

	// Cari userbranch berdasarkan ID
	if err := config.DB.Where("user_id	= ? AND branch_id = ?", userID, branch_id).First(&userbranch).Error; err != nil {
		return responses.JSONResponse(c, http.StatusNotFound, "UserBranch not found", err)
	}

	// Parsing data body langsung ke struct `userbranch`
	// Namun, ini hanya akan mengupdate field-field tertentu.
	if err := c.BodyParser(&userbranch); err != nil {
		return responses.JSONResponse(c, http.StatusBadRequest, "Invalid input", err)
	}

	// Pastikan hanya field yang ingin diperbarui yang diubah.
	// Gunakan `Model` untuk menghindari overwrite seluruh object.
	if err := config.DB.Model(&userbranch).Where("user_id	= ? AND branch_id = ?", userID, branch_id).Updates(userbranch).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Failed to update userbranch", err)
	}

	// Mengembalikan response sukses dengan data userbranch yang diperbarui
	return responses.JSONResponse(c, http.StatusOK, "UserBranch updated successfully", userbranch)
}

// DeleteUserBranch menangani penghapusan userbranch
func DeleteUserBranch(c *framework.Ctx) error {
	user_id := c.Param("user_id")
	branch_id := c.Param("branch_id")
	var userbranch models.UserBranch

	// Cari userbranch berdasarkan ID
	if err := config.DB.Where("user_id	= ? AND branch_id = ?", user_id, branch_id).First(&userbranch).Error; err != nil {
		return responses.JSONResponse(c, http.StatusNotFound, "userbranch not found", err)
	}

	// Hapus userbranch
	// if err := config.DB.Where("user_id	= ? AND branch_id = ?", user_id, branch_id).Delete(&userbranch).Error; err != nil {
	// 	return helpers.JSONResponse(c, fiber.StatusInternalServerError, "Failed to delete userbranch", err)
	// }

	// Hapus userbranch secara permanen
	if err := config.DB.Unscoped().Where("user_id = ? AND branch_id = ?", user_id, branch_id).Delete(&userbranch).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Failed to delete userbranch permanently", err)
	}

	// Mengembalikan response sukses
	return responses.JSONResponse(c, http.StatusOK, "UserBranch deleted successfully", userbranch)
}

// GetAllUserBranch menangani penampilan semua userbranch
func GetAllUserBranch(c *framework.Ctx) error {
	// get branch id
	// branch_id, _ := services.GetBranchID(c)

	// Menampilkan semua userbranch
	var userBranchDetails []models.UserBranchDetail

	// Melakukan LEFT OUTER JOIN menggunakan GORM
	if err := config.DB.
		Table("user_branches usrb").
		Select("usrb.user_id, usr.name AS user_name, usrb.branch_id, brc.branch_name AS branch_name, brc.sia_name, brc.sipa_name, brc.phone").
		Joins("LEFT JOIN users usr ON usr.id = usrb.user_id").
		Joins("LEFT JOIN branches brc ON brc.id = usrb.branch_id").
		// Where("usrb.branch_id = ?", branch_id).
		Scan(&userBranchDetails).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Get user branches failed", "Failed to fetch user branches with details")
	}

	// Mengembalikan response dengan data hasil JOIN
	return responses.JSONResponse(c, http.StatusOK, "UserBranches retrieved successfully", userBranchDetails)
}

// New Function of GetUser in controller used to get user with branch
func GetUserDetails(c *framework.Ctx) error {
	userID := c.Param("user_id")

	// Ambil user
	var user models.User
	if err := config.DB.First(&user, "user_id = ?", userID).Error; err != nil {
		return responses.JSONResponse(c, http.StatusNotFound, "Pengguna tidak ditemukan", err)
	}

	// Ambil relasi cabang dari user melalui UserBranch
	var userBranches []models.UserBranch
	if err := config.DB.Where("user_id = ?", userID).Find(&userBranches).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal mendapatkan cabang", err)
	}

	// Ambil detail branch berdasarkan hasil dari userBranches
	var branchIDs []string
	for _, ub := range userBranches {
		branchIDs = append(branchIDs, ub.BranchID)
	}

	var branches []models.Branch
	if len(branchIDs) > 0 {
		if err := config.DB.Where("id IN ?", branchIDs).Find(&branches).Error; err != nil {
			return responses.JSONResponse(c, http.StatusInternalServerError, "Gagal memuat detail cabang", err)
		}
	}

	type BranchResponse struct {
		BranchID   string `json:"branch_id"`
		BranchName string `json:"branch_name"`
		Address    string `json:"address"`
		Phone      string `json:"phone"`
	}

	// Membuat response format yang diinginkan
	var branchResponses []BranchResponse
	for _, b := range branches {
		branchResponses = append(branchResponses, BranchResponse{
			BranchID:   b.ID,
			BranchName: b.BranchName,
			Address:    b.Address,
			Phone:      b.Phone,
		})
	}

	// Response format yang diinginkan dari API
	type GetUserResponse struct {
		User           models.User      `json:"user"`
		DetailBranches []BranchResponse `json:"detail_branches"`
	}

	// Return user + detail branches
	response := GetUserResponse{
		User:           user,
		DetailBranches: branchResponses,
	}

	return responses.JSONResponse(c, http.StatusOK, "Data berhasil ditemukan", response)
}
