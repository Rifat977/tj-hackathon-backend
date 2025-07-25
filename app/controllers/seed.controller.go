package controllers

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/rizkyizh/go-fiber-boilerplate/database"
)

type SeedController struct{}

func NewSeedController() *SeedController {
	return &SeedController{}
}

// SeedProducts seeds the database with products from the JSON file
func (c *SeedController) SeedProducts(ctx *fiber.Ctx) error {
	// Get counts before seeding
	productsBefore, categoriesBefore, err := database.GetDataCounts()
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to get initial data counts",
			"error":   err.Error(),
		})
	}

	// Perform seeding
	if err := database.SeedProductsFromJSON(); err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to seed products",
			"error":   err.Error(),
		})
	}

	// Get counts after seeding
	productsAfter, categoriesAfter, err := database.GetDataCounts()
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to get final data counts",
			"error":   err.Error(),
		})
	}

	return ctx.Status(http.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Products seeded successfully from JSON file",
		"data": fiber.Map{
			"products_before":   productsBefore,
			"products_after":    productsAfter,
			"products_added":    productsAfter - productsBefore,
			"categories_before": categoriesBefore,
			"categories_after":  categoriesAfter,
			"categories_added":  categoriesAfter - categoriesBefore,
		},
	})
}

// ClearProducts clears all products and categories from the database
func (c *SeedController) ClearProducts(ctx *fiber.Ctx) error {
	// Get counts before clearing
	productsBefore, categoriesBefore, err := database.GetDataCounts()
	if err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to get initial data counts",
			"error":   err.Error(),
		})
	}

	// Perform clearing
	if err := database.ClearProductsData(); err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Failed to clear products",
			"error":   err.Error(),
		})
	}

	// Verify the clearing worked
	if err := database.VerifyDataCleared(); err != nil {
		return ctx.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Clear operation completed but verification failed",
			"error":   err.Error(),
		})
	}

	// Get final counts for response
	productsAfter, categoriesAfter, _ := database.GetDataCounts()

	return ctx.Status(http.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Products and categories cleared successfully",
		"data": fiber.Map{
			"products_before":    productsBefore,
			"products_after":     productsAfter,
			"products_removed":   productsBefore,
			"categories_before":  categoriesBefore,
			"categories_after":   categoriesAfter,
			"categories_removed": categoriesBefore,
		},
	})
}
