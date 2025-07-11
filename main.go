package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"

	"github.com/rizkyizh/go-fiber-boilerplate/config"
	"github.com/rizkyizh/go-fiber-boilerplate/database"
	"github.com/rizkyizh/go-fiber-boilerplate/routes"
)

// @title High Performance Go API with Assets Management
// @version 2.0
// @description A high-performance Go API with authentication, products, assets management, and caching. Features include product seeding from JSON, image serving, and comprehensive product management.
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email support@api.com
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:3000
// @BasePath /
// @schemes http https
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	config.LoadConfig()
	database.ConnectDB()

	// Create Fiber app with performance optimizations
	app := fiber.New(fiber.Config{
		AppName:      "High Performance API",
		ServerHeader: "Go-Fiber",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
		BodyLimit:    10 * 1024 * 1024, // 10MB
	})

	// Add CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins:     config.AppConfig.ALLOWED_ORIGINS,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-Requested-With",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS, PATCH",
		AllowCredentials: true,
		ExposeHeaders:    "Content-Length, Content-Type",
		MaxAge:           86400, // 24 hours
	}))

	// Serve static files (product images, banner images)
	app.Static("/assets", "./assets")

	// Setup routes
	routes.SetupRoutesApp(app)

	// Graceful shutdown
	go func() {
		if err := app.Listen(":" + config.AppConfig.PORT); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	if err := app.Shutdown(); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
