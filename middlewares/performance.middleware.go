package middlewares

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
)

func PerformanceMonitor() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		// Process request
		err := c.Next()

		// Calculate response time
		duration := time.Since(start)

		// Log slow requests (>1 second)
		if duration > time.Second {
			log.Printf("Slow request detected: %s %s took %v", c.Method(), c.Path(), duration)
		}

		// Log performance metrics for all requests
		statusCode := c.Response().StatusCode()
		if statusCode == 0 && err != nil {
			statusCode = 500
		}
		log.Printf("%s %s - %v - %d", c.Method(), c.Path(), duration, statusCode)

		// Return the error from the handler chain
		return err
	}
}
