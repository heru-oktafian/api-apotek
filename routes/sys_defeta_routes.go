package routes

import (
	"os"

	"github.com/heru-oktafian/api-apotek/controllers"
	"github.com/heru-oktafian/scafold/framework"
	"github.com/heru-oktafian/scafold/middlewares"
)

func SysDefectaRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// Grup rute yang DILINDUNGI dengan JWT dan ROLE Authorization
	sysDefectaAPI := app.Group("/api/sys-defectas", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("superadmin", "administrator"))

	// GET /api/sys-defectas - Mengambil semua data defecta sistem
	sysDefectaAPI.Get("/", controllers.GetAllDefectas)

	// GET /api/sys-defectas/:id - Mengambil data defecta sistem berdasarkan ID
	sysDefectaAPI.Get("/:id", controllers.GetDefetaWithItems)

	// POST /api/sys-defectas - Membuat data defecta sistem baru
	sysDefectaAPI.Post("/", controllers.CreateDefecta)

	// PUT /api/sys-defectas/:id - Memperbarui data defecta sistem
	sysDefectaAPI.Put("/:id", controllers.UpdateDefecta)

	// DELETE /api/sys-defectas/:id - Menghapus data defecta sistem (soft delete)
	sysDefectaAPI.Delete("/:id", controllers.DeleteDefecta)
}

func SysDefectaItemRoutes(app *framework.Fiber) {
	// Load Secret Key from environment
	JWTSecret := os.Getenv("JWT_SECRET_KEY")

	// Grup rute yang DILINDUNGI dengan JWT dan ROLE Authorization
	sysDefectaItemAPI := app.Group("/api/sys-defecta-items", middlewares.Protected(JWTSecret), middlewares.AuthorizeRole("superadmin", "administrator"))

	// POST /api/sys-defecta-items - Membuat item defecta sistem baru
	sysDefectaItemAPI.Post("/", controllers.CreateDefectaItem)

	// GET /api/sys-defecta-items/all/:id - Mengambil semua item defecta sistem berdasarkan ID defecta
	sysDefectaItemAPI.Get("/all/:id", controllers.GetAllDefectaItems)

	// PUT /api/sys-defecta-items/:id - Memperbarui item defecta sistem
	sysDefectaItemAPI.Put("/:id", controllers.UpdateDefectaItem)

	// DELETE /api/sys-defecta-items/:id - Menghapus item defecta sistem (soft delete)
	sysDefectaItemAPI.Delete("/:id", controllers.DeleteDefectaItem)
}
