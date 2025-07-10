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
			"image": "banner1.jpg",
			"url":   "/assets/images/Banners/banner1.jpg",
		},
		{
			"id":    "2",
			"name":  "Banner 2",
			"image": "banner2.jpg",
			"url":   "/assets/images/Banners/banner2.jpg",
		},
		{
			"id":    "3",
			"name":  "Banner 3",
			"image": "banner3.jpg",
			"url":   "/assets/images/Banners/banner3.jpg",
		},
	}

	return ctx.JSON(fiber.Map{
		"success": true,
		"banners": banners,
	})
}
