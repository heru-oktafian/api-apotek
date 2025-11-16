package routes

import (
	"os"

	"github.com/heru-oktafian/api-retail/controllers"
	"github.com/heru-oktafian/scafold/framework"
	"github.com/heru-oktafian/scafold/middlewares"
)

// DailyAssetRoutes mengatur rute-rute untuk resource transaksi penjualan
func DailyAssetRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// Grup rute yang DILINDUNGI dengan JWT dan ROLE Authorization
	sysDailyAssetAPI := app.Group("/api/daily_asset", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("operator", "cashier", "finance", "superadmin", "administrator"))

	// GET /api/daily_asset - Mengambil laporan aset harian
	sysDailyAssetAPI.Get("/", controllers.GetAllAssets)
}
