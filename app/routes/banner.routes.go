package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/rizkyizh/go-fiber-boilerplate/app/controllers"
)

func SetupBannerRoutes(app *fiber.App) {
	bannerController := controllers.NewBannerController()

	// Banner routes
	banners := app.Group("/api/banners")
	banners.Get("/", bannerController.GetBanners)
}
