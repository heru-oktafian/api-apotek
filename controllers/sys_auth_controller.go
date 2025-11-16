package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/heru-oktafian/api-retail/models"
	"github.com/heru-oktafian/scafold/config"
	"github.com/heru-oktafian/scafold/framework"
	"github.com/heru-oktafian/scafold/middlewares"
	"github.com/heru-oktafian/scafold/responses"
	"github.com/heru-oktafian/scafold/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// LoginUser mengautentikasi user dan mengembalikan JWT
func LoginUser(c *framework.Ctx) error {
	loginReq := new(struct {
		Username string `json:"username"`
		Password string `json:"password"`
	})

	if err := c.BodyParser(loginReq); err != nil {
		return responses.BadRequest(c, "Format data yang dikirim tidak valid", err)
	}

	var user models.User
	result := config.DB.Where("username = ? AND user_status = 'active'", loginReq.Username).First(&user)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return responses.InternalServerError(c, "An error occurred during login", result.Error)
		}
		// log.Printf("Error retrieving user during login: %v", result.Error)
		return responses.InternalServerError(c, "An error occurred during login", result.Error)
	}

	// Bandingkan password yang di-hash
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginReq.Password))
	if err != nil {
		return responses.InternalServerError(c, "An error occurred during login", err)
	}

	// Buat token JWT
	claims := jwt.MapClaims{
		"user_id":   user.UserID,
		"user_role": user.UserRole,                        // Tambahkan role ke claims JWT
		"exp":       time.Now().Add(time.Hour * 8).Unix(), // Token berlaku 8 jam
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
	if err != nil {
		// log.Printf("Error generating JWT: %v", err)
		return responses.InternalServerError(c, "Could not generate token", err)
	}

	// return utils.SuccessResponse(c, "Login successful", fiber.Map{"token": t})
	return responses.JSONResponse(c, http.StatusOK, "Login successful", t)
}

// GenerateBranchJWTWithRole menghasilkan JWT untuk branch dengan peran tertentu
func generateBranchJWTWithRole(userID string, branchID string, userRole string, defaultMember string, quota int, subscriptionType string, namaUser string) (string, error) {

	// Hitung waktu sekarang dalam WIB
	nowWIB := time.Now().In(utils.Location)

	// Definisikan klaim untuk token baru
	claims := jwt.MapClaims{
		"user_id":           userID,                           // User ID
		"name":              namaUser,                         // Nama User
		"branch_id":         branchID,                         // Branch ID
		"user_role":         userRole,                         // User Role
		"exp":               nowWIB.Add(8 * time.Hour).Unix(), // Expired dalam 8 jam
		"default_member":    defaultMember,
		"quota":             quota,
		"subscription_type": subscriptionType,
	}

	// Buat token baru dengan klaim
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Gunakan secret key untuk menandatangani token
	secretKey := []byte(os.Getenv("JWT_SECRET_KEY"))
	return token.SignedString(secretKey)
}

// blacklistToken untuk menambahkan token ke blacklist di redis
func blacklistToken(token string) error {
	// Parse token untuk mendapatkan waktu kedaluwarsa (exp)
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		secretKey := []byte(os.Getenv("JWT_SECRET_KEY"))
		return secretKey, nil
	})

	if err != nil || !parsedToken.Valid {
		log.Printf("Failed to parse token: %v", err)
		return err
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok || claims["exp"] == nil {
		// log.Println("Invalid token claims, no exp found")
		return fmt.Errorf("invalid token claims")
	}

	// Hitung waktu kedaluwarsa token
	expiryUnix := int64(claims["exp"].(float64)) // Klaim `exp` adalah float64
	expiryTime := time.Unix(expiryUnix, 0)
	ttl := time.Until(expiryTime)

	// Pastikan TTL valid
	if ttl <= 0 {
		// log.Println("Token already expired")
		return fmt.Errorf("token already expired")
	}

	// Tambahkan token ke Redis dengan TTL
	ctx := context.Background()
	redisKey := fmt.Sprintf("blacklist:%s", token)
	err = config.RDB.Set(ctx, redisKey, "blacklisted", ttl).Err()
	if err != nil {
		// log.Printf("Failed to blacklist token: %v", err)
		return err
	}

	// log.Printf("Token blacklisted successfully with TTL: %v", ttl)
	return nil
}

// SetBranch mengatur branch dengan peran tertentu
func SetBranch(c *framework.Ctx) error {
	// Ambil token dari header Authorization
	token := c.Get("Authorization")

	// Hapus prefix "Bearer " jika ada
	token = strings.TrimPrefix(token, "Bearer ")

	// Periksa jika token kosong
	if token == "" {
		return responses.JSONResponse(c, http.StatusUnauthorized, "Missing token", "Insert valid token to access this endpoint!")
	}

	// Verifikasi token JWT untuk mendapatkan user ID
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		secretKey := []byte(os.Getenv("JWT_SECRET_KEY"))
		return secretKey, nil
	})

	if err != nil || !parsedToken.Valid {
		return responses.JSONResponse(c, http.StatusUnauthorized, "Invalid token", "Try to login again!")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok || claims["user_id"] == nil {
		return responses.JSONResponse(c, http.StatusUnauthorized, "Invalid token claims", "Try to login again!")
	}

	// Ambil user ID dari klaim token
	userID := string(claims["user_id"].(string))

	// Parse input JSON untuk mendapatkan branch ID
	var request struct {
		BranchID string `json:"branch_id" validate:"required"`
	}
	if err := c.BodyParser(&request); err != nil {
		return responses.BadRequest(c, "Invalid input", err)
	}

	// Periksa apakah branch_id valid untuk user ini
	var userBranch models.UserBranch
	if err := config.DB.Where("user_id = ? AND branch_id = ?", userID, request.BranchID).First(&userBranch).Error; err != nil {
		return responses.JSONResponse(c, http.StatusForbidden, "Invalid branch ID", "Branch not associated with this user!")
	}

	// Ambil user_role dari tabel users berdasarkan user_id
	var user models.User
	if err := config.DB.Select("name AS name, user_role AS user_role").Where("user_id = ?", userID).First(&user).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Failed to set branch", "Unable to retrieve user role")
	}

	// Ambil default_member, quota, dan subscription_type dari branch
	var branch models.Branch
	if err := config.DB.Select("default_member, quota, subscription_type").Where("id = ?", request.BranchID).First(&branch).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Failed to set branch", "Unable to retrieve branch details")
	}

	// Buat token JWT baru dengan klaim branch_id dan user_role
	newToken, err := generateBranchJWTWithRole(userID, request.BranchID, string(user.UserRole), branch.DefaultMember, branch.Quota, string(branch.SubscriptionType), user.Name)
	if err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Failed to set branch", "Failed to generate new token")
	}

	// Tambahkan token lama ke Redis blacklist
	if err := blacklistToken(token); err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Failed to set branch", "Failed to blacklist old token")
	}

	// Berikan token baru ke pengguna
	return responses.JSONResponse(c, http.StatusOK, "Branch set successfully", newToken)
}

func GetProfile(c *framework.Ctx) error {

	branchID, _ := middlewares.GetBranchID(c.Request)
	userID, _ := middlewares.GetUserID(c.Request)
	userRole, _ := middlewares.GetUserRole(c.Request)
	var profilStruct models.Profile

	// Melakukan LEFT OUTER JOIN menggunakan GORM
	if err := config.DB.
		Table("user_branches usrbrc").
		Select("usrbrc.user_id AS user_id, usr.name AS profile_name, usrbrc.branch_id AS branch_id, brc.branch_name AS branch_name, brc.address, brc.phone, brc.email, brc.bank_name, brc.account_name, brc.account_number, brc.tax_percentage, brc.journal_method, brc.default_member AS default_member, mbr.name AS member_name, brc.branch_status, brc.owner_id, brc.owner_name").
		Joins("LEFT JOIN users usr ON usr.user_id = usrbrc.user_id").
		Joins("LEFT JOIN branches brc ON brc.id = usrbrc.branch_id").
		Joins("LEFT JOIN members mbr ON mbr.id = brc.default_member").
		Where("usrbrc.branch_id = ? AND usrbrc.user_id = ?", branchID, userID).
		Scan(&profilStruct).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Get userbranches failed", "Failed to fetch user branches with details")
	}

	// Mengembalikan response data branch
	return responses.JSONResponse(c, http.StatusOK, "Otoritas : "+userRole, profilStruct)
}

// Function Logout menangani logout pengguna
func Logout(c *framework.Ctx) error {
	// Ambil token dari header Authorization
	token := c.Get("Authorization")

	// Remove prefix "Bearer " jika ada
	token = strings.TrimPrefix(token, "Bearer ")

	if token == "" {
		return responses.JSONResponse(c, http.StatusUnauthorized, "Token tidak ditemukan", "Masukkan token yang valid !")
	}

	// Blacklist token JWT
	if err := blacklistToken(token); err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Logout failed", "Failed to blacklist token")
	}

	return responses.JSONResponse(c, http.StatusOK, "Logout successful", "Logout successful")
}

// CmbBranch menangani penampilan userbranch
func CmbBranch(c *framework.Ctx) error {
	// get user id
	userID, _ := middlewares.GetUserID(c.Request)

	// Menampilkan semua userbranch
	var userBranchDetails []models.UserBranchDetail

	// Melakukan LEFT OUTER JOIN menggunakan GORM
	if err := config.DB.
		Table("user_branches").
		Select("user_branches.user_id, users.name AS user_name, user_branches.branch_id, branches.branch_name, branches.address, branches.phone").
		Joins("LEFT JOIN users ON users.user_id = user_branches.user_id").
		Joins("LEFT JOIN branches ON branches.id = user_branches.branch_id").
		Where("branches.branch_status = 'active' AND user_branches.user_id = ?", userID).
		Scan(&userBranchDetails).Error; err != nil {
		return responses.JSONResponse(c, http.StatusInternalServerError, "Get userbranches failed", "Failed to fetch user branches with details")
	}

	// Mengembalikan response data userbranch
	return responses.JSONResponse(c, http.StatusOK, "UserBranch found", userBranchDetails)
}

// GetMenus handles the request to retrieve menu data with an optional user_role filter.
func GetMenus(c *framework.Ctx) error {
	userRoles, _ := middlewares.GetClaimsToken(c.Request, "user_role")

	// Read the menus.json file
	data, err := os.ReadFile("menus.json")
	if err != nil {
		log.Printf("Error reading menus.json: %v", err)
		return responses.InternalServerError(c, "Failed to read menu data", err)
	}

	var menuResponse models.MenuResponse // Gunakan struct pembungkus baru
	err = json.Unmarshal(data, &menuResponse)
	if err != nil {
		log.Printf("Error unmarshaling menus.json: %v", err)
		return responses.InternalServerError(c, "Failed to parse menu data", err)
	}

	// Data menu yang sebenarnya sekarang ada di menuResponse.Data
	menus := menuResponse.Data

	// Get the user_role query parameter
	userRoleFilter := userRoles

	if userRoleFilter == "" {
		// If no user_role filter is provided, return all menus
		return responses.JSONResponse(c, http.StatusOK, "Get All Menus Success", menus)
	}

	// Filter menus by user_role
	var filteredMenus []models.Menu // Gunakan models.Menu
	for _, menu := range menus {
		if strings.EqualFold(menu.UserRole, userRoleFilter) {
			filteredMenus = append(filteredMenus, menu)
		}
	}

	if len(filteredMenus) == 0 {
		return responses.JSONResponse(c, http.StatusNotFound, "No menu found for the specified user_role", []models.Menu{})
	}

	return responses.JSONResponse(c, http.StatusOK, "Get Menus by User Role Success", filteredMenus)
}
