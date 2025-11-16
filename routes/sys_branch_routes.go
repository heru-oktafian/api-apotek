package routes

import (
	"os"

	"github.com/heru-oktafian/api-retail/controllers"
	"github.com/heru-oktafian/scafold/framework"
	"github.com/heru-oktafian/scafold/middlewares"
)

// SysBranchRoutes mengatur rute-rute untuk resource cabang
func SysBranchRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")
	// Grup rute yang DILINDUNGI dengan JWT dan ROLE Authorization
	// Hanya 'administrator' dan 'superadmin' yang bisa mengakses rute ini
	branchAPI := app.Group("/api/branches", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("administrator", "superadmin"))
	// branchAPI := app.Group("/api/branches", middleware.AuthorizeRole("administrator", "superadmin"))

	// GET /api/branches - Mengambil semua cabang
	branchAPI.Get("/", controllers.GetAllBranch)

	// GET /api/branches/:id - Mengambil cabang berdasarkan ID
	branchAPI.Get("/:id", controllers.GetBranch)

	// POST /api/branches - Membuat cabang baru
	branchAPI.Post("/", controllers.CreateBranch)

	// PUT /api/branches/:id - Memperbarui cabang
	branchAPI.Put("/:id", controllers.UpdateBranch)

	// DELETE /api/branches/:id - Menghapus cabang (soft delete)
	branchAPI.Delete("/:id", controllers.DeleteBranch)
}
