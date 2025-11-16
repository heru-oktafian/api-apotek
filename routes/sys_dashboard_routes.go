package routes

import (
	"os"

	"github.com/heru-oktafian/api-retail/controllers"
	"github.com/heru-oktafian/scafold/framework"
	"github.com/heru-oktafian/scafold/middlewares"
)

func SysDashboardRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// Dashboards routes
	dashboards := app.Group("/api/dashboard", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("operator", "cashier", "finance", "superadmin", "administrator"))
	dashboards.Get("/monthly-profit-report", controllers.MonthlyProfitReport)
	dashboards.Get("/daily-profit-report", controllers.DailyProfitReport)
	dashboards.Get("/weekly-profit-report", controllers.WeeklyProfitReport)
	dashboards.Get("/profit-today-by-user", controllers.GetDailyProfitReportByUser)
	dashboards.Get("/top-selling-report", controllers.GetTopSellingProducts)
	dashboards.Get("/least-selling-report", controllers.GetLeastSellingProducts)
	dashboards.Get("/neared-report", controllers.GetExpiringProducts)
}
