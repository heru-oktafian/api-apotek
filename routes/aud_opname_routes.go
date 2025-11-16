package routes

import (
	"os"

	"github.com/heru-oktafian/api-retail/controllers"
	"github.com/heru-oktafian/scafold/framework"
	"github.com/heru-oktafian/scafold/middlewares"
)

// AuditMobileOpnamesRoutes sets up the routes for managing mobile opname.
func AuditMobileOpnamesRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// MobileOpnames routes
	mobileOpname := app.Group("/api/mobile-opnames", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("operator", "cashier", "finance", "superadmin", "administrator"))
	mobileOpname.Get("/", controllers.GetAllMobileOpnames)
	mobileOpname.Get("-active/", controllers.GetAllActiveMobileOpnames)
	mobileOpname.Get("-item-details/", controllers.GetMobileOpnameItemDetails)
	mobileOpname.Get("-items-glimpse/", controllers.GetMobileOpnameItemsGlimpse)
}

// AuditOpnameRoutes sets up the routes for managing opname.
func AuditOpnameRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// AuditOpname routes
	auditOpname := app.Group("/api/opnames", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("superadmin", "administrator"))
	auditOpname.Get("/", controllers.GetAllOpnames)
	auditOpname.Post("/", controllers.CreateOpname)
	auditOpname.Get("/:id", controllers.GetOpnameWithItems)
	auditOpname.Put("/:id", controllers.UpdateOpnameByID)
	auditOpname.Delete("/:id", controllers.DeleteOpnameByID)
}

func AuditOpnameItemRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// AuditOpnameItem routes
	auditOpnameItem := app.Group("/api/opname-items", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("operator", "cashier", "finance", "superadmin", "administrator"))
	auditOpnameItem.Get("/:id", controllers.GetAllOpnameItems)
	auditOpnameItem.Post("/", controllers.CreateOpnameItem)
	auditOpnameItem.Put("/:id", controllers.UpdateOpnameItemByID)
	auditOpnameItem.Delete("/:id", controllers.DeleteOpnameItemByID)
}

func CmbProductOpnameRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// CmbProductOpname routes
	cmbProductOpname := app.Group("/api/cmb-product-opname", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("operator", "cashier", "finance", "superadmin", "administrator"))
	cmbProductOpname.Get("/", controllers.GetProductsComboboxByName)
}
