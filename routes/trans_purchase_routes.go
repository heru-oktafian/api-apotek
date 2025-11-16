package routes

import (
	"os"

	"github.com/heru-oktafian/api-retail/controllers"
	"github.com/heru-oktafian/scafold/framework"
	"github.com/heru-oktafian/scafold/middlewares"
)

// TransPurchaseRoutes mengatur rute-rute untuk resource transaksi pembelian
func TransPurchaseRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// Grup rute yang DILINDUNGI dengan JWT dan ROLE Authorization
	transPurchaseAPI := app.Group("/api/purchases", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("operator", "cashier", "finance", "superadmin", "administrator"))

	// GET /api/purchases - Mengambil semua transaksi pembelian
	transPurchaseAPI.Get("/", controllers.GetAllPurchases)

	// GET /api/purchases/:id - Mengambil transaksi pembelian berdasarkan ID
	transPurchaseAPI.Get("/:id", controllers.GetPurchaseWithItems)

	// POST /api/purchases - Membuat transaksi pembelian baru
	transPurchaseAPI.Post("/", controllers.CreatePurchaseTransaction)

	// PUT /api/purchases/:id - Memperbarui transaksi pembelian
	transPurchaseAPI.Put("/:id", controllers.UpdatePurchase)

	// DELETE /api/purchases/:id - Menghapus transaksi pembelian (soft delete)
	transPurchaseAPI.Delete("/:id", controllers.DeletePurchase)
}

// TransPurchaseItemRoutes mengatur rute-rute untuk resource item pembelian
func TransPurchaseItemRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// Grup rute yang DILINDUNGI dengan JWT dan ROLE Authorization
	transPurchaseItemAPI := app.Group("/api/purchase-items", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("operator", "cashier", "finance", "superadmin", "administrator"))

	// POST /api/purchase_items - Membuat item pembelian baru
	transPurchaseItemAPI.Post("/", controllers.CreatePurchaseItem)

	// GET /api/purchase-items/all/:id - Mengambil semua item pembelian berdasarkan ID transaksi
	transPurchaseItemAPI.Get("/all/:id", controllers.GetAllPurchaseItems)

	// PUT /api/purchase-items/:id - Memperbarui item pembelian
	transPurchaseItemAPI.Put("/:id", controllers.UpdatePurchaseItem)

	// DELETE /api/purchase-items/:id - Menghapus item pembelian (soft delete)
	transPurchaseItemAPI.Delete("/:id", controllers.DeletePurchaseItem)
}
