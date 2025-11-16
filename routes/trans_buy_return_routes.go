package routes

import (
	"os"

	"github.com/heru-oktafian/api-retail/controllers"
	"github.com/heru-oktafian/scafold/framework"
	"github.com/heru-oktafian/scafold/middlewares"
)

// TransBuyReturnRoutes mengatur rute-rute untuk resource transaksi pembelian retur
func TransBuyReturnRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// Grup rute yang DILINDUNGI dengan JWT dan ROLE Authorization
	transBuyReturnAPI := app.Group("/api/buy-returns", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("superadmin", "administrator"))

	// GET /api/buy-returns - Mengambil semua transaksi pembelian retur
	transBuyReturnAPI.Get("/", controllers.GetAllBuyReturns)

	// GET /api/buy-returns/:id - Mengambil transaksi pembelian retur berdasarkan ID
	transBuyReturnAPI.Get("/:id", controllers.GetBuyReturnWithItems)

	// POST /api/buy-returns - Membuat transaksi pembelian retur baru
	transBuyReturnAPI.Post("/", controllers.CreateBuyReturnTransaction)
}

func CmbProdBuyReturn(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// Grup rute yang DILINDUNGI dengan JWT dan ROLE Authorization
	cmbProdBuyReturnAPI := app.Group("/api/cmb-prod-buy-returns", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("superadmin", "administrator"))

	// GET /api/cmb-prod-buy-returns - Mengambil semua kombinasi produk dari transaksi pembelian retur
	cmbProdBuyReturnAPI.Get("/", controllers.GetBuyItemsForReturn)
}

func CmbPurchaseRoute(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// Grup rute yang DILINDUNGI dengan JWT dan ROLE Authorization
	cmbPurchaseAPI := app.Group("/api/cmb-purchases", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("superadmin", "administrator"))

	// GET /api/cmb-purchases - Mengambil semua pembelian
	cmbPurchaseAPI.Get("/", controllers.CmbPurchase)
}
