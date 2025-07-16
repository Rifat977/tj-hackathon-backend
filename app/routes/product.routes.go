package routes

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/rizkyizh/go-fiber-boilerplate/app/controllers"
	"github.com/rizkyizh/go-fiber-boilerplate/database"
)

// Rate limiting middleware
func rateLimit(requests int, window time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		clientIP := c.IP()
		key := "rate_limit:" + clientIP

		ctx := c.Context()
		count, err := database.Redis.Incr(ctx, key).Result()
		if err != nil {
			return c.Next()
		}

		if count == 1 {
			database.Redis.Expire(ctx, key, window)
		}

		if count > int64(requests) {
			return c.Status(429).JSON(fiber.Map{
				"error":       "Rate limit exceeded",
				"retry_after": window.Seconds(),
			})
		}

		return c.Next()
	}
}

func SetupProductRoutes(app *fiber.App) {
	productController := controllers.NewProductController()

	// Product routes with rate limiting
	products := app.Group("/api/products")
	// products.Use(rateLimit(100, time.Minute)) // 100 requests per minute
	products.Get("/", productController.GetProducts)
	products.Get("/search", productController.SearchProducts)
	products.Get("/:id", productController.GetProductByID)

	// Category routes with rate limiting
	categories := app.Group("/api/categories")
	// categories.Use(rateLimit(200, time.Minute)) // 200 requests per minute for categories
	categories.Get("/", productController.GetCategories)
	categories.Get("/:id/products", productController.GetProductsByCategory)
}
