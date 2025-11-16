package routes

import (
	"os"

	"github.com/heru-oktafian/api-retail/controllers"
	"github.com/heru-oktafian/scafold/framework"
	"github.com/heru-oktafian/scafold/middlewares"
)

// TransSaleRoutes mengatur rute-rute untuk resource transaksi penjualan
func TransSaleRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// Grup rute yang DILINDUNGI dengan JWT dan ROLE Authorization
	transSaleAPI := app.Group("/api/sales", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("operator", "cashier", "finance", "superadmin", "administrator"))

	// GET /api/sales - Mengambil semua transaksi penjualan
	transSaleAPI.Get("/", controllers.GetAllSales)

	// GET /api/sales/:id - Mengambil transaksi penjualan berdasarkan ID
	transSaleAPI.Get("/:id", controllers.GetSaleWithItems)

	// POST /api/sales - Membuat transaksi penjualan baru
	transSaleAPI.Post("/", controllers.CreateSaleTransaction)

	// PUT /api/sales/:id - Memperbarui transaksi penjualan
	transSaleAPI.Put("/:id", controllers.UpdateSale)

	// DELETE /api/sales/:id - Menghapus transaksi penjualan (soft delete)
	transSaleAPI.Delete("/:id", controllers.DeleteSale)
}

func TransSaleItemRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// Grup rute yang DILINDUNGI dengan JWT dan ROLE Authorization
	transSaleItemAPI := app.Group("/api/sale-items", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("operator", "cashier", "finance", "superadmin", "administrator"))

	// POST /api/sale_items - Membuat item penjualan baru
	transSaleItemAPI.Post("/", controllers.CreateSaleItem)

	// GET /api/sale-items/all/:id - Mengambil semua item penjualan berdasarkan ID transaksi
	transSaleItemAPI.Get("/all/:id", controllers.GetAllSaleItems)

	// PUT /api/sale-items/:id - Memperbarui item penjualan
	transSaleItemAPI.Put("/:id", controllers.UpdateSaleItem)

	// DELETE /api/sale-items/:id - Menghapus item penjualan (soft delete)
	transSaleItemAPI.Delete("/:id", controllers.DeleteSaleItem)
}

// TransSaleDetailRoutes mengatur rute untuk resource transaksi penjualan dengan detail item di dalamnya
func TransSaleDetailRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// Grup rute yang DILINDUNGI dengan JWT dan ROLE Authorization
	transSaleDetailAPI := app.Group("/api/sales-details", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("operator", "cashier", "finance", "superadmin", "administrator"))

	// GET /api/sales - Mengambil semua transaksi penjualan
	transSaleDetailAPI.Get("/", controllers.GetAllSalesDetail)
}
