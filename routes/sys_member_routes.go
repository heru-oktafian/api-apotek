package routes

import (
	"os"

	"github.com/heru-oktafian/api-retail/controllers"
	"github.com/heru-oktafian/scafold/framework"
	"github.com/heru-oktafian/scafold/middlewares"
)

func SysMemberRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// Grup rute yang dilindungi dengan JWT dan ROLE Authorization
	memberAPI := app.Group("/api/members", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("operator", "cashier", "finance", "superadmin", "administrator"))

	// Member Endpoints
	memberAPI.Post("/", controllers.CreateMember)
	memberAPI.Get("/", controllers.GetAllMember)
	memberAPI.Get("/:id", controllers.GetMember)
	memberAPI.Put("/:id", controllers.UpdateMember)
	memberAPI.Delete("/:id", controllers.DeleteMember)
}

// CmbMemberCategoryRoutes mengatur rute untuk mendapatkan daftar kategori member.
// Rute ini tidak memerlukan autentikasi, sehingga tidak dilindungi oleh middleware.
func CmbMemberCategoryRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// Rute ini tidak memerlukan autentikasi, sehingga tidak dilindungi oleh middleware.
	cmbMemberCategoryAPI := app.Group("/api/member-categories-combo", middlewares.Protected(JWTSecret))
	cmbMemberCategoryAPI.Get("/", controllers.CmbMemberCategory)
}

func CmbMemberRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// Rute ini tidak memerlukan autentikasi, sehingga tidak dilindungi oleh middleware.
	cmbMemberAPI := app.Group("/api/members-combo", middlewares.Protected(JWTSecret))
	cmbMemberAPI.Get("/", controllers.CmbMember)
}
