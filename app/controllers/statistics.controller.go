package controllers

import (
	"log"

	"github.com/gofiber/fiber/v2"

	"github.com/rizkyizh/go-fiber-boilerplate/app/services"
)

type StatisticsController struct {
	statisticsService *services.StatisticsService
}

func NewStatisticsController() *StatisticsController {
	return &StatisticsController{
		statisticsService: services.NewStatisticsService(),
	}
}

// DownloadStatistics handles downloading product statistics as CSV
// @Summary Download product statistics CSV
// @Description Downloads a CSV file containing product statistics including totals, averages, and counts
// @Tags statistics
// @Accept json
// @Produce application/octet-stream
// @Success 200 {file} csv "CSV file containing product statistics"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /api/statistics/download [get]
func (c *StatisticsController) DownloadStatistics(ctx *fiber.Ctx) error {
	log.Println("Generating product statistics CSV...")

	// Generate CSV data
	csvData, err := c.statisticsService.GenerateCSV()
	if err != nil {
		log.Printf("Error generating CSV: %v", err)
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to generate statistics",
			"message": err.Error(),
		})
	}

	// Set response headers for file download
	ctx.Set("Content-Type", "text/csv")
	ctx.Set("Content-Disposition", "attachment; filename=product_statistics.csv")
	ctx.Set("Content-Length", string(rune(len(csvData))))

	// Send CSV data
	return ctx.Send(csvData)
}
