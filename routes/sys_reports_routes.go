package routes

import (
	"os"

	"github.com/heru-oktafian/api-retail/controllers"
	"github.com/heru-oktafian/scafold/framework"
	"github.com/heru-oktafian/scafold/middlewares"
)

func SysReportRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// Expense routes
	reports := app.Group("/api/report", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("operator", "cashier", "finance", "superadmin", "administrator"))
	reports.Get("/neraca-saldo", controllers.GetNeracaSaldo)
	reports.Get("/profit-by-month", controllers.GetProfitGraphByMonth)
}
