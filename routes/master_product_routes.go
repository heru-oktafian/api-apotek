package routes

import (
	"os"

	"github.com/heru-oktafian/api-retail/controllers"
	"github.com/heru-oktafian/scafold/framework"
	"github.com/heru-oktafian/scafold/middlewares"
)

// ProductRoutes mengatur rute-rute terkait produk di API.
func MasterProductRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// Grup rute yang dilindungi JWT
	productAPI := app.Group("/api/products", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("operator", "cashier", "finance", "superadmin", "administrator"))

	productAPI.Post("/", controllers.CreateProduct)
	productAPI.Get("/", controllers.GetAllProduct)
	productAPI.Get("/:id", controllers.GetProduct)
	productAPI.Put("/:id", controllers.UpdateProduct)
	productAPI.Delete("/:id", controllers.DeleteProduct)
}

func CmbProdSaleRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// Grup rute yang dilindungi JWT
	cmbProductSaleAPI := app.Group("/api/sale-products-combo", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("operator", "cashier", "finance", "superadmin", "administrator"))

	cmbProductSaleAPI.Get("/", controllers.CmbProdSale)
}

func CmbProdPurchaseRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// Grup rute yang dilindungi JWT
	cmbProductPurchaseAPI := app.Group("/api/purchase-products-combo", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("operator", "cashier", "finance", "superadmin", "administrator"))

	cmbProductPurchaseAPI.Get("/", controllers.CmbProdPurchase)
}
