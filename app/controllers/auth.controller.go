package controllers

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"

	"github.com/rizkyizh/go-fiber-boilerplate/app/dto"
	"github.com/rizkyizh/go-fiber-boilerplate/app/services"
)

type AuthController struct {
	authService *services.AuthService
	validate    *validator.Validate
}

func NewAuthController() *AuthController {
	return &AuthController{
		authService: services.NewAuthService(),
		validate:    validator.New(),
	}
}

// formatValidationError formats validation errors into a user-friendly message
func formatValidationError(err error) string {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var errors []string
		for _, e := range validationErrors {
			switch e.Tag() {
			case "required":
				errors = append(errors, e.Field()+" is required")
			case "email":
				errors = append(errors, e.Field()+" must be a valid email address")
			case "min":
				errors = append(errors, e.Field()+" must be at least "+e.Param()+" characters")
			default:
				errors = append(errors, e.Field()+" failed validation: "+e.Tag())
			}
		}
		return strings.Join(errors, "; ")
	}
	return err.Error()
}

// @Summary Register a new user
// @Description Register a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param body body dto.RegisterRequest true "Register request"
// @Success 201 {object} dto.AuthResponse
// @Failure 400 {object} map[string]interface{}
// @Router /api/auth/register [post]
func (c *AuthController) Register(ctx *fiber.Ctx) error {
	var req dto.RegisterRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := c.validate.Struct(req); err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"error": formatValidationError(err),
		})
	}

	user, err := c.authService.Register(req.Email, req.Password, req.FirstName, req.LastName)
	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	token, err := c.authService.GenerateJWT(*user)
	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	response := dto.AuthResponse{
		Token: token,
		User: dto.UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Role:      user.Role,
		},
	}

	return ctx.Status(201).JSON(response)
}

// @Summary Login user
// @Description Login with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param body body dto.LoginRequest true "Login request"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} map[string]interface{}
// @Router /api/auth/login [post]
func (c *AuthController) Login(ctx *fiber.Ctx) error {
	var req dto.LoginRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := c.validate.Struct(req); err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"error": formatValidationError(err),
		})
	}

	user, token, err := c.authService.Login(req.Email, req.Password)
	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	response := dto.AuthResponse{
		Token: token,
		User: dto.UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Role:      user.Role,
		},
	}

	return ctx.JSON(response)
}

// @Summary Logout user
// @Description Logout and invalidate session
// @Tags auth
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Router /api/auth/logout [post]
func (c *AuthController) Logout(ctx *fiber.Ctx) error {
	token := ctx.Get("Authorization")
	if token != "" {
		token = token[7:] // Remove "Bearer " prefix
		c.authService.Logout(token)
	}

	return ctx.JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}

// @Summary Get user profile
// @Description Get current user profile
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.UserResponse
// @Failure 401 {object} map[string]interface{}
// @Router /api/auth/profile [get]
func (c *AuthController) GetProfile(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(uint)

	user, err := c.authService.GetUserByID(userID)
	if err != nil {
		return ctx.Status(404).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	response := dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
	}

	return ctx.JSON(response)
}

// @Summary Update user profile
// @Description Update current user profile
// @Tags auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body dto.UpdateProfileRequest true "Update profile request"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} map[string]interface{}
// @Router /api/auth/profile [put]
func (c *AuthController) UpdateProfile(ctx *fiber.Ctx) error {
	userID := ctx.Locals("user_id").(uint)

	var req dto.UpdateProfileRequest
	if err := ctx.BodyParser(&req); err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err := c.authService.UpdateProfile(userID, req.FirstName, req.LastName, req.Phone, req.Address, req.City, req.Country, req.PostalCode)
	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"error": "Failed to update profile",
		})
	}

	user, err := c.authService.GetUserByID(userID)
	if err != nil {
		return ctx.Status(404).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	response := dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      user.Role,
	}

	return ctx.JSON(response)
}
