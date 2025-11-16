package routes

import (
	"os"

	"github.com/heru-oktafian/api-retail/controllers"
	"github.com/heru-oktafian/scafold/framework"
	"github.com/heru-oktafian/scafold/middlewares"
)

// SysUserRoutes mengatur rute-rute untuk resource pengguna
func SysUserRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")
	// Grup rute yang DILINDUNGI dengan JWT dan ROLE Authorization
	// Hanya 'administrator' dan 'superadmin' yang bisa mengakses rute ini
	userAPI := app.Group("/api/users", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("administrator", "superadmin"))

	// GET /api/users - Mengambil semua pengguna
	// userAPI.Get("/", middleware.CacheMiddleware(5*time.Minute), controllers.GetUsers)
	userAPI.Get("/", controllers.GetUsers)

	// GET /api/users/:user_id - Mengambil pengguna berdasarkan ID
	userAPI.Get("/:user_id", controllers.GetUserByID)

	// POST /api/users - Membuat pengguna baru
	userAPI.Post("/", controllers.CreateUser)

	// PUT /api/users/:user_id - Memperbarui pengguna
	userAPI.Put("/:user_id", controllers.UpdateUser)

	// DELETE /api/users/:user_id - Menghapus pengguna (soft delete)
	userAPI.Delete("/:user_id", controllers.DeleteUser)
}
