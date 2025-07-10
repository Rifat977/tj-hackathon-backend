package routes

import (
	"github.com/gofiber/fiber/v2"

	"github.com/rizkyizh/go-fiber-boilerplate/app/controllers"
	"github.com/rizkyizh/go-fiber-boilerplate/middlewares"
)

func SetupAuthRoutes(app *fiber.App) {
	authController := controllers.NewAuthController()

	auth := app.Group("/api/auth")

	// Public routes
	auth.Post("/register", authController.Register)
	auth.Post("/login", authController.Login)

	// Protected routes
	auth.Post("/logout", middlewares.AuthMiddleware(), authController.Logout)
	auth.Get("/profile", middlewares.AuthMiddleware(), authController.GetProfile)
	auth.Put("/profile", middlewares.AuthMiddleware(), authController.UpdateProfile)
}
