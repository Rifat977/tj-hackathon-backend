package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/rizkyizh/go-fiber-boilerplate/app/routes"
	"github.com/rizkyizh/go-fiber-boilerplate/middlewares"
)

func SetupRoutesApp(app *fiber.App) {
	// Add performance middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(middlewares.PerformanceMonitor())
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	// Setup route groups
	routes.SetupAuthRoutes(app)
	routes.SetupProductRoutes(app)
	routes.SetupHealthRoutes(app)
	routes.SetupSeedRoutes(app)
	routes.SetupBannerRoutes(app)
	routes.SetupAdminRoutes(app)
}
