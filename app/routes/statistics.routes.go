package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/rizkyizh/go-fiber-boilerplate/app/controllers"
)

func SetupStatisticsRoutes(app *fiber.App) {
	statisticsController := controllers.NewStatisticsController()

	// Statistics routes
	statistics := app.Group("/api/statistics")
	statistics.Get("/download", statisticsController.DownloadStatistics)
}
