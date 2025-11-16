package routes

import (
	"os"

	"github.com/heru-oktafian/api-retail/controllers"
	"github.com/heru-oktafian/scafold/framework"
	"github.com/heru-oktafian/scafold/middlewares"
)

// MasterUnitConversionRoutes mengatur rute-rute untuk resource konversi unit
func MasterUnitConversionRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// Grup rute yang DILINDUNGI dengan JWT dan ROLE Authorization
	unitConversionAPI := app.Group("/api/unit-conversions", middlewares.Protected(JWTSecret))

	// GET /api/unit-conversions - Mengambil semua konversi unit
	unitConversionAPI.Get("/", controllers.GetAllUnitConversion, middlewares.AuthorizeRole("operator", "cashier", "finance", "superadmin", "administrator"))

	// GET /api/unit-conversions/:id - Mengambil konversi unit berdasarkan ID
	unitConversionAPI.Get("/:id", controllers.GetUnitConversionByID, middlewares.AuthorizeRole("operator", "cashier", "finance", "superadmin", "administrator"))

	// POST /api/unit-conversions - Membuat konversi unit baru
	unitConversionAPI.Post("/", controllers.CreateUnitConversion, middlewares.AuthorizeRole("operator", "superadmin", "administrator"))

	// PUT /api/unit-conversions/:id - Memperbarui konversi unit
	unitConversionAPI.Put("/:id", controllers.UpdateUnitConversion, middlewares.AuthorizeRole("operator", "superadmin", "administrator"))

	// DELETE /api/unit-conversions/:id - Menghapus konversi unit (soft delete)
	unitConversionAPI.Delete("/:id", controllers.DeleteUnitConversion, middlewares.AuthorizeRole("superadmin", "administrator"))
}

// CmbProdConvRoutes mengatur rute untuk mendapatkan daftar produk konversi.
func CmbProdConvRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	cmbProdConvAPI := app.Group("/api/conversion-products-combo", middlewares.Protected(JWTSecret))
	cmbProdConvAPI.Get("/", controllers.CmbProdConv)
}

// CmbUnitConvRoutes mengatur rute untuk mendapatkan daftar unit konversi.
func CmbUnitConvRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	cmbUnitConvAPI := app.Group("/api/conversion-units-combo", middlewares.Protected(JWTSecret))
	cmbUnitConvAPI.Get("/", controllers.CmbUnit)
}
