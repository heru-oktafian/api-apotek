package routes

import (
	"os"

	"github.com/heru-oktafian/api-retail/controllers"
	"github.com/heru-oktafian/scafold/framework"
	"github.com/heru-oktafian/scafold/middlewares"
)

// SysUserBranchRoutes mengatur rute-rute untuk resource user branch
func SysUserBranchRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")
	// Grup rute yang DILINDUNGI dengan JWT dan ROLE Authorization
	// Hanya 'administrator' dan 'superadmin' yang bisa mengakses rute ini
	userBranchAPI := app.Group("/api/user-branches", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("administrator", "superadmin"))

	// GET /api/user-branches - Mengambil semua user branch
	userBranchAPI.Get("/", controllers.GetAllUserBranch)

	// GET /api/user-branches/:user_id/:branch_id - Mengambil user branch berdasarkan ID
	userBranchAPI.Get("/:user_id/:branch_id", controllers.GetUserBranch)

	// POST /api/user-branches - Membuat user branch baru
	userBranchAPI.Post("/", controllers.CreateUserBranch)

	// PUT /api/user-branches/:user_id/:branch_id - Memperbarui user branch
	userBranchAPI.Put("/:user_id/:branch_id", controllers.UpdateUserBranch)

	// DELETE /api/user-branches/:user_id/:branch_id - Menghapus user branch (soft delete)
	userBranchAPI.Delete("/:user_id/:branch_id", controllers.DeleteUserBranch)

	// GET /api/user-branches/:id - Mengambil user branch berdasarkan ID
	userBranchAPI.Get("/:id", controllers.GetUserDetails)
}

func GetBranchByUserId(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")
	// Grup rute yang DILINDUNGI dengan JWT dan ROLE Authorization
	// Hanya 'administrator' dan 'superadmin' yang bisa mengakses rute ini
	userBranchByUserId := app.Group("/api/detail-users", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("administrator", "superadmin"))

	// GET /api/detail-users/ - Mengambil semua cabang berdasarkan user_id
	userBranchByUserId.Get("/:user_id", middlewares.Protected(JWTSecret), controllers.GetUserDetails)
}
