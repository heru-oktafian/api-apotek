package routes

import (
	"os"

	"github.com/heru-oktafian/api-retail/controllers"
	"github.com/heru-oktafian/scafold/framework"
	"github.com/heru-oktafian/scafold/middlewares"
)

func TransExpenseRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// Expense routes
	expense := app.Group("/api/expenses", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("operator", "cashier", "finance", "superadmin", "administrator"))
	expense.Post("/", controllers.CreateExpense)
	expense.Put("/:id", controllers.UpdateExpense)
	expense.Delete("/:id", controllers.DeleteExpense)
	expense.Get("/", controllers.GetAllExpenses)
}
