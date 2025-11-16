package routes

import (
	"os"

	"github.com/heru-oktafian/api-retail/controllers"
	"github.com/heru-oktafian/scafold/framework"
	"github.com/heru-oktafian/scafold/middlewares"
)

// AuditFirstStock sets up the routes for managing first stock audits.
func AuditFirstStockRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// Grup rute yang dilindungi JWT
	firstStock := app.Group("/api/first-stock", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("operator", "cashier", "finance", "superadmin", "administrator"))
	firstStock.Get("/", controllers.GetAllFirstStocks)
	firstStock.Post("/", controllers.CreateFirstStockTransaction)
	firstStock.Put("/:id", controllers.UpdateFirstStock)
	firstStock.Delete("/:id", controllers.DeleteFirstStock)
}

func AuditFirstStockWithItems(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// FirstStock routes
	firstStock := app.Group("/api/first-stock-with-items", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("operator", "cashier", "finance", "superadmin", "administrator"))
	firstStock.Get("/:id", controllers.GetFirstStockWithItems)
}

func AuditFirstStockItemRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// FirstStock Items routes
	firstStockItems := app.Group("/api/first-stock-items", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("operator", "cashier", "finance", "superadmin", "administrator"))
	firstStockItems.Get("/:id", controllers.GetAllFirstStockItems)   // Fungsi ini untuk mendapatkan semua item dari first stock
	firstStockItems.Post("/", controllers.CreateFirstStockItem)      // Fungsi ini untuk membuat item baru pada first stock
	firstStockItems.Put("/:id", controllers.UpdateFirstStockItem)    // Fungsi ini untuk memperbarui item pada first stock
	firstStockItems.Delete("/:id", controllers.DeleteFirstStockItem) // Fungsi ini untuk menghapus item dari first stock
}
