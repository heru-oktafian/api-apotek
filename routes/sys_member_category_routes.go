package routes

import (
	"os"

	"github.com/heru-oktafian/api-retail/controllers"
	"github.com/heru-oktafian/scafold/framework"
	"github.com/heru-oktafian/scafold/middlewares"
)

func SysMemberCategoryRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// Grup rute yang dilindungi JWT
	memberCategoryAPI := app.Group("/api/member-categories", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("operator", "cashier", "finance", "superadmin", "administrator"))

	memberCategoryAPI.Post("/", controllers.CreateMemberCategory)
	memberCategoryAPI.Get("/", controllers.GetAllMemberCategory)
	memberCategoryAPI.Get("/:id", controllers.GetMemberCategory)
	memberCategoryAPI.Put("/:id", controllers.UpdateMemberCategory)
	memberCategoryAPI.Delete("/:id", controllers.DeleteMemberCategory)
}
