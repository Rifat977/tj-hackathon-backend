package controllers

import (
	"github.com/gofiber/fiber/v2"
)

type BannerController struct{}

func NewBannerController() *BannerController {
	return &BannerController{}
}

// GetBanners returns a list of available banner images
// @Summary Get banner images
// @Description Get list of available banner images with their URLs for frontend display
// @Tags banners
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Success - List of banner images"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /api/banners [get]
func (c *BannerController) GetBanners(ctx *fiber.Ctx) error {
	banners := []map[string]string{
		{
			"id":    "1",
			"name":  "Banner 1",
			"image": "banner1.webp",
			"url":   "/assets/images/Banners/banner1.webp",
		},
		{
			"id":    "2",
			"name":  "Banner 2",
			"image": "banner2.webp",
			"url":   "/assets/images/Banners/banner2.webp",
		},
		{
			"id":    "3",
			"name":  "Banner 3",
			"image": "banner3.webp",
			"url":   "/assets/images/Banners/banner3.webp",
		},
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"banners": banners,
	})
}
