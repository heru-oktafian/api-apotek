package routes

import (
	"os"

	"github.com/heru-oktafian/api-retail/controllers"
	"github.com/heru-oktafian/scafold/framework"
	"github.com/heru-oktafian/scafold/middlewares"
)

// MasterSupplierRoutes mengatur rute-rute terkait supplier di API.
func MasterSupplierRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// Grup rute yang dilindungi JWT
	supplierAPI := app.Group("/api/suppliers", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("operator", "cashier", "finance", "superadmin", "administrator"))
	supplierAPI.Get("/", controllers.GetAllSupplier)
	supplierAPI.Get("/:id", controllers.GetSupplierByID)
	supplierAPI.Post("/", controllers.CreateSupplier)
	supplierAPI.Put("/:id", controllers.UpdateSupplier)
	supplierAPI.Delete("/:id", controllers.DeleteSupplier)
}

// CmbSupplierCategoryRoutes mengatur rute untuk combo box kategori supplier
func CmbSupplierCategoryRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// Grup rute yang dilindungi JWT

	cmbSupplierAPI := app.Group("/api/supplier-categories-combo", middlewares.Protected(JWTSecret))
	cmbSupplierAPI.Get("/", controllers.CmbSupplierCategory)
}

// CmbSupplierRoutes mengatur rute untuk combo box supplier
func CmbSupplierRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// Grup rute yang dilindungi JWT
	cmbSupplierAPI := app.Group("/api/suppliers-combo", middlewares.Protected(JWTSecret))
	cmbSupplierAPI.Get("/", controllers.CmbSupplier)
}
