package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/rizkyizh/go-fiber-boilerplate/app/controllers"
)

func SetupSeedRoutes(app *fiber.App) {
	seedController := controllers.NewSeedController()

	// Seed routes
	seed := app.Group("/api/seed")
	seed.Post("/products", seedController.SeedProducts)
	seed.Delete("/clear", seedController.ClearProducts)
}
