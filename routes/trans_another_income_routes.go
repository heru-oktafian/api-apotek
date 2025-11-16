package routes

import (
	"os"

	"github.com/heru-oktafian/api-retail/controllers"
	"github.com/heru-oktafian/scafold/framework"
	"github.com/heru-oktafian/scafold/middlewares"
)

func TransAnotherIncomeRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// Another Income routes
	anotherIncome := app.Group("/api/another-incomes", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("operator", "cashier", "finance", "superadmin", "administrator"))
	anotherIncome.Post("/", controllers.CreateAnotherIncome)
	anotherIncome.Put("/:id", controllers.UpdateAnotherIncome)
	anotherIncome.Delete("/:id", controllers.DeleteAnotherIncome)
	anotherIncome.Get("/", controllers.GetAllAnotherIncomes)
}
