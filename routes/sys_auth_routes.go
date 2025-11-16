package routes

import (
	"os"

	"github.com/heru-oktafian/api-retail/controllers"
	"github.com/heru-oktafian/scafold/framework"
	"github.com/heru-oktafian/scafold/middlewares"
)

// SysAuthRoutes mengatur rute untuk autentikasi (register dan login)
func SysAuthRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// Group routes under /api
	auth := app.Group("/api")

	// auth.Post("/register", controllers.RegisterUser)
	auth.Post("/login", controllers.LoginUser)
	auth.Post("/logout", controllers.Logout)
	auth.Post("/set_branch", middlewares.Protected(JWTSecret), controllers.SetBranch)
	auth.Get("/profile", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("administrator", "superadmin"), controllers.GetProfile)
	auth.Get("/list_branches", middlewares.Protected(JWTSecret), controllers.CmbBranch)
}

func SysMenuRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// Group routes under /api
	menu := app.Group("/menus", middlewares.Protected(JWTSecret))

	// Protected route to get menus, requires authentication and specific roles
	menu.Get("/", controllers.GetMenus)
}
