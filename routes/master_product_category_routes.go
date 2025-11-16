package routes

import (
	"os"

	"github.com/heru-oktafian/api-retail/controllers"
	"github.com/heru-oktafian/scafold/framework"
	"github.com/heru-oktafian/scafold/middlewares"
)

func MasterProductCategoryRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// Grup rute yang dilindungi JWT
	productCategoryAPI := app.Group("/api/product-categories", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("operator", "cashier", "finance", "superadmin", "administrator"))

	// Definisikan rute untuk product categories
	productCategoryAPI.Post("/", controllers.CreateProductCategory)
	productCategoryAPI.Get("/", controllers.GetAllProductCategory)
	productCategoryAPI.Get("/:id", controllers.GetProductCategory)
	productCategoryAPI.Put("/:id", controllers.UpdateProductCategory)
	productCategoryAPI.Delete("/:id", controllers.DeleteProductCategory)
}

func CmbProductCategoryRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// Grup rute yang dilindungi JWT
	cmbProductCategoryAPI := app.Group("/api/product-categories-combo", middlewares.Protected(JWTSecret))
	cmbProductCategoryAPI.Get("/", controllers.CmbProductCategory)
}
