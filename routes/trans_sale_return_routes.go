package routes

import (
	"os"

	"github.com/heru-oktafian/api-retail/controllers"
	"github.com/heru-oktafian/scafold/framework"
	"github.com/heru-oktafian/scafold/middlewares"
)

// TransSaleReturnRoutes mengatur rute-rute untuk resource transaksi penjualan retur
func TransSaleReturnRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// Grup rute yang DILINDUNGI dengan JWT dan ROLE Authorization
	transSaleReturnAPI := app.Group("/api/sale-returns", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("superadmin", "administrator"))

	// GET /api/sale-returns - Mengambil semua transaksi penjualan retur
	transSaleReturnAPI.Get("/", controllers.GetAllSaleReturns)

	// GET /api/sale-returns/:id - Mengambil transaksi penjualan retur berdasarkan ID
	transSaleReturnAPI.Get("/:id", controllers.GetSaleReturnWithItems)

	// POST /api/sale-returns - Membuat transaksi penjualan retur baru
	transSaleReturnAPI.Post("/", controllers.CreateSaleReturnTransaction)
}

func CmbProdSaleReturn(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// Grup rute yang DILINDUNGI dengan JWT dan ROLE Authorization
	cmbProdSaleReturnAPI := app.Group("/api/cmb-prod-sale-returns", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("superadmin", "administrator"))

	// GET /api/cmb-prod-sale-returns - Mengambil semua kombinasi produk dari transaksi penjualan retur
	cmbProdSaleReturnAPI.Get("/", controllers.GetSaleItemsForReturn)
}

func CmbSaleRoute(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// Grup rute yang DILINDUNGI dengan JWT dan ROLE Authorization
	cmbSaleAPI := app.Group("/api/cmb-sales", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("superadmin", "administrator"))

	// GET /api/cmb-sales - Mengambil semua penjualan
	cmbSaleAPI.Get("/", controllers.CmbSale)
}
