package routes

import (
	"os"

	"github.com/heru-oktafian/api-apotek/controllers"
	"github.com/heru-oktafian/scafold/framework"
	"github.com/heru-oktafian/scafold/middlewares"
)

// TransDuplicateReceiptRoutes mengatur rute-rute untuk resource transaksi duplicate receipt
func TransDuplicateReceiptRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")
	// Grup rute yang DILINDUNGI dengan JWT dan ROLE Authorization
	transDuplicateAPI := app.Group("/api/duplicate-receipts", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("operator", "cashier", "finance", "superadmin", "administrator"))

	// GET /api/duplicate-receipts - Mengambil semua transaksi kopi resep
	transDuplicateAPI.Get("/", controllers.GetAllDuplicateReceipts)

	// GET /api/duplicate-receipts/:id - Mengambil transaksi kopi resep berdasarkan ID
	transDuplicateAPI.Get("/:id", controllers.GetDuplicateWithItems)

	// POST /api/duplicate-receipts - Membuat transaksi kopi resep baru
	transDuplicateAPI.Post("/", controllers.CreateDuplicateReceipt)

	// PUT /api/duplicate-receipts/:id - Memperbarui transaksi kopi resep
	transDuplicateAPI.Put("/:id", controllers.UpdateDuplicateReceipt)

	// DELETE /api/duplicate-receipts/:id - Menghapus transaksi kopi resep (soft delete)
	transDuplicateAPI.Delete("/:id", controllers.DeleteDuplicateReceipt)
}

// TransDuplicateItemRoutes mengatur rute untuk resource item transaksi kopi resep
func TransDuplicateItemRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// Grup rute yang DILINDUNGI dengan JWT dan ROLE Authorization
	transDuplicateItemAPI := app.Group("/api/duplicate-receipts-items", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("operator", "cashier", "finance", "superadmin", "administrator"))

	// POST /api/duplicate-receipts-items - Membuat item penjualan baru
	transDuplicateItemAPI.Post("/", controllers.CreateDuplicateReceiptItem)

	// GET /api/duplicate-receipts-items/all/:id - Mengambil semua item penjualan berdasarkan ID transaksi
	transDuplicateItemAPI.Get("/all/:id", controllers.GetAllDuplicateItems)

	// PUT /api/duplicate-receipts-items/:id - Memperbarui item penjualan
	transDuplicateItemAPI.Put("/:id", controllers.UpdateDuplicateReceiptItem)

	// DELETE /api/duplicate-receipts-items/:id - Menghapus item penjualan (soft delete)
	transDuplicateItemAPI.Delete("/:id", controllers.DeleteDuplicateReceiptItem)
}

// TransDuplicateDetailRoutes mengatur rute untuk resource transaksi kopi resep dengan detail item di dalamnya
func TransDuplicateDetailRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// Grup rute yang DILINDUNGI dengan JWT dan ROLE Authorization
	transDuplicateDetailAPI := app.Group("/api/duplicate-receipts-details", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("operator", "cashier", "finance", "superadmin", "administrator"))

	// GET /api/duplicate-receipts-details - Mengambil semua detail transaksi duplicate receipt
	transDuplicateDetailAPI.Get("/", controllers.GetAllDuplicateDetail)
}
