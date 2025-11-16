package routes

import (
	"os"

	"github.com/heru-oktafian/api-retail/controllers"
	"github.com/heru-oktafian/scafold/framework"
	"github.com/heru-oktafian/scafold/middlewares"
)

// MasterSupplierRoutes mengatur rute-rute terkait supplier category di API.
func MasterSupplierCategoryRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// Grup rute yang dilindungi JWT
	supplierCatAPI := app.Group("/api/supplier-categories", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("operator", "cashier", "finance", "superadmin", "administrator"))

	supplierCatAPI.Get("/", controllers.GetAllSupplierCategory)
	supplierCatAPI.Get("/:id", controllers.GetSupplierCategoryByID)
	supplierCatAPI.Post("/", controllers.CreateSupplierCategory)
	supplierCatAPI.Put("/:id", controllers.UpdateSupplierCategory)
	supplierCatAPI.Delete("/:id", controllers.DeleteSupplierCategory)
}
