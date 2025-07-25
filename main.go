package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberRecover "github.com/gofiber/fiber/v2/middleware/recover"
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
	// Initialize logger
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("🚀 Starting High Performance Go API...")

	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Load configuration
	config.LoadConfig()

	// Connect to databases with retry logic
	if err := connectWithRetry(); err != nil {
		log.Fatalf("Failed to connect to databases after retries: %v", err)
	}

	// Create Fiber app with enhanced configuration
	app := createFiberApp()

	// Setup middleware
	setupMiddleware(app)

	// Serve static files
	app.Static("/assets", "./assets")

	// Setup routes
	routes.SetupRoutesApp(app)

	// Setup graceful shutdown
	setupGracefulShutdown(app)
}

func createFiberApp() *fiber.App {
	return fiber.New(fiber.Config{
		AppName:               "High Performance API",
		ServerHeader:          "Go-Fiber",
		ReadTimeout:           config.AppConfig.ReadTimeout,
		WriteTimeout:          config.AppConfig.WriteTimeout,
		IdleTimeout:           config.AppConfig.IdleTimeout,
		BodyLimit:             int(config.AppConfig.BodyLimit),
		EnablePrintRoutes:     false,
		DisableStartupMessage: false,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			log.Printf("Error: %v", err)
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
				"code":  code,
			})
		},
	})
}

func setupMiddleware(app *fiber.App) {
	// Recovery middleware to handle panics
	app.Use(fiberRecover.New(fiberRecover.Config{
		EnableStackTrace: true,
	}))

	// CORS middleware
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))
}

func connectWithRetry() error {
	maxRetries := 5
	retryDelay := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		log.Printf("Attempting to connect to databases (attempt %d/%d)", i+1, maxRetries)

		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("Panic during database connection: %v", r)
				}
			}()
			database.ConnectDB()
		}()

		// Test connections
		if database.DB != nil && database.Redis != nil {
			// Test database connection
			sqlDB, err := database.DB.DB()
			if err == nil {
				if err := sqlDB.Ping(); err == nil {
					// Test Redis connection
					ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
					defer cancel()
					if _, err := database.Redis.Ping(ctx).Result(); err == nil {
						log.Println("✅ Successfully connected to all databases")
						return nil
					}
				}
			}
		}

		if i < maxRetries-1 {
			log.Printf("Connection failed, retrying in %v...", retryDelay)
			time.Sleep(retryDelay)
			retryDelay *= 2 // Exponential backoff
		}
	}

	return fmt.Errorf("failed to connect to databases after %d attempts", maxRetries)
}

func setupGracefulShutdown(app *fiber.App) {
	// Create a channel to receive OS signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// Start server in a goroutine
	go func() {
		port := os.Getenv("PORT")
		if port == "" {
			port = "3000"
		}

		log.Printf("🌟 Server starting on port %s", port)
		if err := app.Listen(":" + port); err != nil {
			log.Printf("❌ Server startup error: %v", err)
		}
	}()

	// Wait for interrupt signal
	sig := <-quit
	log.Printf("🛑 Received signal: %v. Starting graceful shutdown...", sig)

	// Create shutdown context with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Start shutdown process
	shutdownComplete := make(chan bool, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Panic during shutdown: %v", r)
			}
			shutdownComplete <- true
		}()

		log.Println("📝 Closing database connections...")
		closeDatabaseConnections()

		log.Println("🔌 Shutting down HTTP server...")
		if err := app.ShutdownWithContext(shutdownCtx); err != nil {
			log.Printf("❌ Server shutdown error: %v", err)
		}

		log.Println("✅ Graceful shutdown completed")
	}()

	// Wait for shutdown to complete or timeout
	select {
	case <-shutdownComplete:
		log.Println("🎯 Application stopped gracefully")
	case <-shutdownCtx.Done():
		log.Println("⚠️ Shutdown timeout exceeded, forcing exit")
	}

	// Final cleanup
	log.Println("👋 Application exited")
	os.Exit(0)
}

func closeDatabaseConnections() {
	// Close Redis connection
	if database.Redis != nil {
		if err := database.Redis.Close(); err != nil {
			log.Printf("Error closing Redis connection: %v", err)
		} else {
			log.Println("✅ Redis connection closed")
		}
	}

	// Close database connection pool
	if database.Pool != nil {
		database.Pool.Close()
		log.Println("✅ Database pool closed")
	}

	// Close GORM database connection
	if database.DB != nil {
		if sqlDB, err := database.DB.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				log.Printf("Error closing database connection: %v", err)
			} else {
				log.Println("✅ Database connection closed")
			}
		}
	}
}
