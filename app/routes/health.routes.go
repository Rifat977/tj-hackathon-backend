package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"

	"github.com/rizkyizh/go-fiber-boilerplate/app/controllers"
	_ "github.com/rizkyizh/go-fiber-boilerplate/docs"
)

func SetupHealthRoutes(app *fiber.App) {
	healthController := controllers.NewHealthController()

	app.Get("/api/health", healthController.HealthCheck)

	// Swagger documentation
	app.Get("/swagger/*", swagger.HandlerDefault)
}
