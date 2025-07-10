package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/rizkyizh/go-fiber-boilerplate/app/controllers"
)

func SetupProductRoutes(app *fiber.App) {
	productController := controllers.NewProductController()

	// Product routes
	products := app.Group("/api/products")
	products.Get("/", productController.GetProducts)
	products.Get("/search", productController.SearchProducts)
	products.Get("/:id", productController.GetProductByID)

	// Category routes
	categories := app.Group("/api/categories")
	categories.Get("/", productController.GetCategories)
	categories.Get("/:id/products", productController.GetProductsByCategory)
}
