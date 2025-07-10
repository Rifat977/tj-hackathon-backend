package controllers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/rizkyizh/go-fiber-boilerplate/database"
)

type HealthController struct{}

func NewHealthController() *HealthController {
	return &HealthController{}
}

// @Summary Health check
// @Description Check system health status including database and Redis connectivity
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Success - System health status"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /api/health [get]
func (c *HealthController) HealthCheck(ctx *fiber.Ctx) error {
	// Check database connection
	dbStatus := "healthy"
	if err := database.DB.Raw("SELECT 1").Error; err != nil {
		dbStatus = "unhealthy"
	}

	// Check Redis connection
	redisStatus := "healthy"
	redisCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := database.Redis.Ping(redisCtx).Err(); err != nil {
		redisStatus = "unhealthy"
	}

	response := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
		"database":  dbStatus,
		"redis":     redisStatus,
		"version":   "1.0.0",
	}

	return ctx.JSON(response)
}
