package controllers

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"runtime"

	"github.com/rizkyizh/go-fiber-boilerplate/app/dto"
	"github.com/rizkyizh/go-fiber-boilerplate/app/models"
	"github.com/rizkyizh/go-fiber-boilerplate/app/services"
	"github.com/rizkyizh/go-fiber-boilerplate/config"
)

type AdminController struct {
	productService *services.ProductService
}

func NewAdminController() *AdminController {
	return &AdminController{
		productService: services.NewProductService(),
	}
}

// Helper function to convert product to response DTO
func (c *AdminController) convertProductToResponse(product models.Product) dto.ProductResponse {
	// Generate image URL
	imageURL := ""
	if product.Image != "" {
		imageURL = fmt.Sprintf("/assets/images/Products/%s", product.Image)
	}

	response := dto.ProductResponse{
		ID:               product.ID,
		Index:            product.Index,
		Name:             product.Name,
		Description:      product.Description,
		ShortDescription: product.ShortDescription,
		Brand:            product.Brand,
		Category:         product.Category,
		Price:            product.Price,
		Currency:         product.Currency,
		Stock:            product.Stock,
		EAN:              product.EAN,
		Color:            product.Color,
		Size:             product.Size,
		Availability:     product.Availability,
		Image:            product.Image,
		ImageURL:         imageURL,
		InternalID:       product.InternalID,
		Slug:             product.Slug,
		SKU:              product.SKU,
		Active:           product.Active,
		CreatedAt:        product.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        product.UpdatedAt.Format(time.RFC3339),
	}

	// Add category model if available
	if product.CategoryModel.ID != 0 {
		response.CategoryModel = dto.CategoryResponse{
			ID:          product.CategoryModel.ID,
			Name:        product.CategoryModel.Name,
			Description: product.CategoryModel.Description,
			Slug:        product.CategoryModel.Slug,
			Active:      product.CategoryModel.Active,
		}
	}

	return response
}

// Helper function to convert products to responses
func (c *AdminController) convertProductsToResponses(products []models.Product) []dto.ProductResponse {
	responses := make([]dto.ProductResponse, len(products))
	for i, product := range products {
		responses[i] = c.convertProductToResponse(product)
	}
	return responses
}

// @Summary Admin Dashboard
// @Description Serve admin dashboard HTML page
// @Tags admin
// @Accept html
// @Produce html
// @Success 200 {string} string "Admin Dashboard"
// @Router /admin [get]
func (c *AdminController) Dashboard(ctx *fiber.Ctx) error {
	// Add timeout context
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Return the admin dashboard HTML
	return ctx.SendFile("./views/admin/dashboard.html")
}

// @Summary Get products for admin
// @Description Get paginated list of products for admin panel
// @Tags admin
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Items per page" default(20) minimum(1) maximum(100)
// @Param search query string false "Search query"
// @Param category_id query int false "Filter by category ID"
// @Success 200 {object} dto.ProductListResponse "Success"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /admin/api/products [get]
func (c *AdminController) GetProducts(ctx *fiber.Ctx) error {
	// Add timeout context
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "20"))
	search := ctx.Query("search")
	categoryIDStr := ctx.Query("category_id")

	var categoryID *uint
	if categoryIDStr != "" {
		if id, err := strconv.ParseUint(categoryIDStr, 10, 32); err == nil {
			catID := uint(id)
			categoryID = &catID
		}
	}

	// Use search if provided, otherwise use regular get products
	var products []models.Product
	var total int64
	var err error

	if search != "" {
		products, total, err = c.productService.SearchProducts(search, "", "", "", "", "", page, limit)
	} else {
		// For admin dashboard, always fetch fresh data without cache
		products, total, err = c.productService.GetProductsWithoutCache(page, limit, categoryID)
	}

	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch products",
		})
	}

	// Convert to response DTOs using helper function
	productResponses := c.convertProductsToResponses(products)

	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	pagination := dto.PaginationInfo{
		Page:       page,
		Limit:      limit,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}

	response := dto.ProductListResponse{
		Products:   productResponses,
		Pagination: pagination,
	}

	return ctx.JSON(response)
}

// @Summary Get categories for admin
// @Description Get list of all categories for admin panel
// @Tags admin
// @Accept json
// @Produce json
// @Success 200 {array} dto.CategoryResponse "Success"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /admin/api/categories [get]
func (c *AdminController) GetCategories(ctx *fiber.Ctx) error {
	// Add timeout context
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	categories, err := c.productService.GetCategories()
	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch categories",
		})
	}

	var categoryResponses []dto.CategoryResponse
	for _, category := range categories {
		categoryResponses = append(categoryResponses, dto.CategoryResponse{
			ID:          category.ID,
			Name:        category.Name,
			Description: category.Description,
			Slug:        category.Slug,
			Active:      category.Active,
		})
	}

	return ctx.JSON(categoryResponses)
}

// @Summary Get product by ID for admin
// @Description Get detailed information about a specific product for admin panel
// @Tags admin
// @Accept json
// @Produce json
// @Param id path int true "Product ID" minimum(1)
// @Success 200 {object} dto.ProductResponse "Success"
// @Failure 400 {object} map[string]interface{} "Bad Request - Invalid product ID"
// @Failure 404 {object} map[string]interface{} "Not Found - Product not found"
// @Router /admin/api/products/{id} [get]
func (c *AdminController) GetProductByID(ctx *fiber.Ctx) error {
	// Add timeout context
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id, err := strconv.ParseUint(ctx.Params("id"), 10, 32)
	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"error": "Invalid product ID",
		})
	}

	// Use non-cached version for admin dashboard
	product, err := c.productService.GetProductByIDWithoutCache(uint(id))
	if err != nil {
		return ctx.Status(404).JSON(fiber.Map{
			"error": "Product not found",
		})
	}

	// Convert to response using helper function
	response := c.convertProductToResponse(*product)

	return ctx.JSON(response)
}

// @Summary Create new product
// @Description Create a new product
// @Tags admin
// @Accept json
// @Produce json
// @Param product body dto.CreateProductRequest true "Product data"
// @Success 201 {object} dto.ProductResponse "Product created"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /admin/api/products [post]
func (c *AdminController) CreateProduct(ctx *fiber.Ctx) error {
	// Add timeout context
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var createRequest dto.CreateProductRequest
	if err := ctx.BodyParser(&createRequest); err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if createRequest.Name == "" {
		return ctx.Status(400).JSON(fiber.Map{
			"error": "Product name is required",
		})
	}

	if createRequest.Price <= 0 {
		return ctx.Status(400).JSON(fiber.Map{
			"error": "Product price must be greater than 0",
		})
	}

	if createRequest.CategoryID == 0 {
		return ctx.Status(400).JSON(fiber.Map{
			"error": "Category is required",
		})
	}

	// Create product using service
	product, err := c.productService.CreateProduct(createRequest)
	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"error": "Failed to create product",
		})
	}

	// Convert to response
	response := c.convertProductToResponse(*product)

	return ctx.Status(201).JSON(response)
}

// @Summary Update product
// @Description Update an existing product
// @Tags admin
// @Accept json
// @Produce json
// @Param id path int true "Product ID" minimum(1)
// @Param product body dto.UpdateProductRequest true "Product data"
// @Success 200 {object} dto.ProductResponse "Product updated"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 404 {object} map[string]interface{} "Not Found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /admin/api/products/{id} [put]
func (c *AdminController) UpdateProduct(ctx *fiber.Ctx) error {
	// Add timeout context
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id, err := strconv.ParseUint(ctx.Params("id"), 10, 32)
	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"error": "Invalid product ID",
		})
	}

	var updateRequest dto.UpdateProductRequest
	if err := ctx.BodyParser(&updateRequest); err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Update product using service
	product, err := c.productService.UpdateProduct(uint(id), updateRequest)
	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"error": "Failed to update product",
		})
	}

	// Convert to response
	response := c.convertProductToResponse(*product)

	return ctx.JSON(response)
}

// @Summary Delete product
// @Description Delete a product
// @Tags admin
// @Accept json
// @Produce json
// @Param id path int true "Product ID" minimum(1)
// @Success 200 {object} map[string]interface{} "Product deleted"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 404 {object} map[string]interface{} "Not Found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /admin/api/products/{id} [delete]
func (c *AdminController) DeleteProduct(ctx *fiber.Ctx) error {
	// Add timeout context
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	id, err := strconv.ParseUint(ctx.Params("id"), 10, 32)
	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"error": "Invalid product ID",
		})
	}

	// Delete product using service
	err = c.productService.DeleteProduct(uint(id))
	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"error": "Failed to delete product",
		})
	}

	return ctx.JSON(fiber.Map{
		"message": "Product deleted successfully",
	})
}

// @Summary Bulk upload products
// @Description Upload multiple products from JSON file with high-performance processing
// @Tags admin
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "JSON file with products"
// @Success 200 {object} map[string]interface{} "Products uploaded"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 408 {object} map[string]interface{} "Request Timeout"
// @Failure 413 {object} map[string]interface{} "Request Entity Too Large"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /admin/api/products/bulk [post]
func (c *AdminController) BulkUploadProducts(ctx *fiber.Ctx) error {
	// Get uploaded file first to calculate timeout
	file, err := ctx.FormFile("file")
	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"error": "No file uploaded",
		})
	}

	// Validate file type
	if !strings.HasSuffix(file.Filename, ".json") {
		return ctx.Status(400).JSON(fiber.Map{
			"error": "Only JSON files are allowed",
		})
	}

	// Validate file size (max 500MB for large bulk uploads)
	maxFileSize := int64(500 * 1024 * 1024) // 500MB
	if file.Size > maxFileSize {
		return ctx.Status(413).JSON(fiber.Map{
			"error":        fmt.Sprintf("File size too large. Maximum size is %d MB", maxFileSize/(1024*1024)),
			"file_size_mb": float64(file.Size) / (1024 * 1024),
			"max_size_mb":  maxFileSize / (1024 * 1024),
		})
	}

	// Calculate dynamic timeout based on file size and server configuration
	fileSizeMB := float64(file.Size) / (1024 * 1024)
	baseTimeout := config.AppConfig.WriteTimeout

	// Calculate additional time based on file size (1 second per MB for large files)
	additionalTime := time.Duration(fileSizeMB*1.0) * time.Second
	timeoutSeconds := int((baseTimeout + additionalTime).Seconds())

	// Cap at 30 minutes for very large files
	if timeoutSeconds > 1800 {
		timeoutSeconds = 1800
	}

	// Minimum 10 minutes for any file
	if timeoutSeconds < 600 {
		timeoutSeconds = 600
	}

	fmt.Printf("📊 Upload timeout calculated: %d seconds for %.2f MB file\n", timeoutSeconds, fileSizeMB)
	fmt.Printf("🚀 Using optimized processing with 4 workers for 2 vCPU environment\n")
	fmt.Printf("⚙️ Server Write Timeout: %v, Additional Time: %v\n", baseTimeout, additionalTime)

	// Create context with timeout
	uploadCtx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	// Create a channel to receive the result
	resultChan := make(chan *dto.BulkUploadResult, 1)
	errChan := make(chan error, 1)

	// Start bulk upload in a goroutine with memory optimization
	go func() {
		// Set memory limit for this goroutine
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		initialAlloc := m.Alloc

		result, err := c.productService.BulkUploadProducts(file)
		if err != nil {
			errChan <- err
			return
		}

		// Log memory usage
		runtime.ReadMemStats(&m)
		memoryUsed := m.Alloc - initialAlloc
		fmt.Printf("💾 Memory used during upload: %.2f MB\n", float64(memoryUsed)/(1024*1024))

		resultChan <- result
	}()

	// Wait for result or timeout
	select {
	case result := <-resultChan:
		// Return detailed results
		response := fiber.Map{
			"message":      "Bulk upload completed",
			"uploaded":     result.Uploaded,
			"failed":       result.Failed,
			"total":        result.Uploaded + result.Failed,
			"timeout":      timeoutSeconds,
			"file_size_mb": fileSizeMB,
		}

		// Include errors if any
		if len(result.Errors) > 0 {
			response["errors"] = result.Errors
			// Limit error messages to first 10 to avoid response size issues
			if len(result.Errors) > 10 {
				response["errors"] = result.Errors[:10]
				response["error_message"] = fmt.Sprintf("Showing first 10 errors. Total errors: %d", len(result.Errors))
			}
		}

		return ctx.JSON(response)

	case err := <-errChan:
		return ctx.Status(500).JSON(fiber.Map{
			"error":   "Failed to upload products",
			"details": err.Error(),
		})

	case <-uploadCtx.Done():
		// Handle timeout gracefully
		return ctx.Status(408).JSON(fiber.Map{
			"error":           "Request timeout",
			"message":         fmt.Sprintf("Upload timed out after %d seconds. Try uploading a smaller file or contact support.", timeoutSeconds),
			"timeout_seconds": timeoutSeconds,
			"file_size_mb":    fileSizeMB,
		})
	}
}

// @Summary Delete all products
// @Description Delete all products from the database
// @Tags admin
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "All products deleted"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /admin/api/products/bulk-delete [post]
func (c *AdminController) DeleteAllProducts(ctx *fiber.Ctx) error {
	// Add timeout context
	_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Delete all products using service
	err := c.productService.DeleteAllProducts()
	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"error": "Failed to delete all products",
		})
	}

	return ctx.JSON(fiber.Map{
		"message": "All products deleted successfully",
	})
}

// @Summary Clear all caches
// @Description Clear all Redis caches for better performance
// @Tags admin
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Cache cleared"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /admin/api/cache/clear [post]
func (c *AdminController) ClearCache(ctx *fiber.Ctx) error {
	// Add timeout context
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Clear all caches using service
	err := c.productService.ClearAllCaches()
	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"error": "Failed to clear cache",
		})
	}

	return ctx.JSON(fiber.Map{
		"message": "All caches cleared successfully",
	})
}
