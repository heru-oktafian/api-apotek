package routes

import (
	//"os"

	"github.com/heru-oktafian/scafold/framework"
)

func CobaRoutes(app *framework.Fiber) {
	// JWTSecret := os.Getenv("JWT_SECRET_KEY")
	cobaAPI := app.Group("/coba")
	cobaAPI.Get("/", func(c *framework.Ctx) error {
		return c.JSON(200, map[string]interface{}{
			"status":  200,
			"message": "Coba endpoint hit successfully : ", //+ JWTSecret,
			"data":    nil,
		})
	})
}
