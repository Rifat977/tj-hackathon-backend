package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/rizkyizh/go-fiber-boilerplate/app/controllers"
)

func SetupAdminRoutes(app *fiber.App) {
	adminController := controllers.NewAdminController()

	// Admin dashboard route
	app.Get("/admin", adminController.Dashboard)

	// Admin API routes
	adminAPI := app.Group("/admin/api")
	adminAPI.Get("/products", adminController.GetProducts)
	adminAPI.Get("/products/:id", adminController.GetProductByID)
	adminAPI.Post("/products", adminController.CreateProduct)
	adminAPI.Put("/products/:id", adminController.UpdateProduct)
	adminAPI.Delete("/products/:id", adminController.DeleteProduct)
	adminAPI.Post("/products/bulk", adminController.BulkUploadProducts)
	adminAPI.Post("/products/bulk-delete", adminController.DeleteAllProducts)
	adminAPI.Get("/categories", adminController.GetCategories)
	adminAPI.Post("/cache/clear", adminController.ClearCache)
}
