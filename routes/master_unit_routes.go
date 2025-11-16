package routes

import (
	"os"

	"github.com/heru-oktafian/api-retail/controllers"
	"github.com/heru-oktafian/scafold/framework"
	"github.com/heru-oktafian/scafold/middlewares"
)

// MasterUnitRoutes mengatur rute untuk master unit
func MasterUnitRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// Grup rute yang dilindungi JWT
	unitAPI := app.Group("/api/units", middlewares.Protected(JWTSecret))

	// GET /api/units -> ambil semua unit
	unitAPI.Get("/", controllers.GetAllUnit)

	// GET /api/units/:id -> ambil unit berdasarkan ID
	unitAPI.Get("/:id", controllers.GetUnitByID)

	// POST /api/units -> buat unit baru
	unitAPI.Post("/", controllers.CreateUnit, middlewares.AuthorizeRole("superadmin", "administrator"))

	// PUT /api/unites/:id - Memperbarui unit
	unitAPI.Put("/:id", controllers.UpdateUnit, middlewares.AuthorizeRole("superadmin", "administrator"))

	// DELETE /api/unites/:id - Menghapus unit (soft delete)
	unitAPI.Delete("/:id", controllers.DeleteUnit, middlewares.AuthorizeRole("superadmin", "administrator"))
}

func CmbUnitRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// Grup rute yang dilindungi JWT
	cmbUnitAPI := app.Group("/api/units-combo", middlewares.Protected(JWTSecret))
	cmbUnitAPI.Get("/", controllers.CmbUnit)
}
