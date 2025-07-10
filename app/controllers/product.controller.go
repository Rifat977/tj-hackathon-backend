package controllers

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/rizkyizh/go-fiber-boilerplate/app/dto"
	"github.com/rizkyizh/go-fiber-boilerplate/app/services"
)

type ProductController struct {
	productService *services.ProductService
}

func NewProductController() *ProductController {
	return &ProductController{
		productService: services.NewProductService(),
	}
}

// @Summary Get products list
// @Description Get paginated list of products with filtering and sorting options
// @Tags products
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Items per page" default(10) minimum(1) maximum(100)
// @Param category_id query int false "Filter by category ID"
// @Success 200 {object} dto.ProductListResponse "Success"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /api/products [get]
func (c *ProductController) GetProducts(ctx *fiber.Ctx) error {
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "10"))
	categoryIDStr := ctx.Query("category_id")

	var categoryID *uint
	if categoryIDStr != "" {
		if id, err := strconv.ParseUint(categoryIDStr, 10, 32); err == nil {
			catID := uint(id)
			categoryID = &catID
		}
	}

	products, total, err := c.productService.GetProducts(page, limit, categoryID)
	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch products",
		})
	}

	// Convert to response DTOs
	var productResponses []dto.ProductResponse
	for _, product := range products {
		// Generate image URL
		imageURL := ""
		if product.Image != "" {
			imageURL = fmt.Sprintf("/assets/images/Products/%s", product.Image)
		}

		productResponse := dto.ProductResponse{
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
			productResponse.CategoryModel = dto.CategoryResponse{
				ID:          product.CategoryModel.ID,
				Name:        product.CategoryModel.Name,
				Description: product.CategoryModel.Description,
				Slug:        product.CategoryModel.Slug,
				Active:      product.CategoryModel.Active,
			}
		}

		productResponses = append(productResponses, productResponse)
	}

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

// @Summary Get product by ID
// @Description Get detailed information about a specific product by its ID
// @Tags products
// @Accept json
// @Produce json
// @Param id path int true "Product ID" minimum(1)
// @Success 200 {object} dto.ProductResponse "Success"
// @Failure 400 {object} map[string]interface{} "Bad Request - Invalid product ID"
// @Failure 404 {object} map[string]interface{} "Not Found - Product not found"
// @Router /api/products/{id} [get]
func (c *ProductController) GetProductByID(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 32)
	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"error": "Invalid product ID",
		})
	}

	product, err := c.productService.GetProductByID(uint(id))
	if err != nil {
		return ctx.Status(404).JSON(fiber.Map{
			"error": "Product not found",
		})
	}

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

	return ctx.JSON(response)
}

// @Summary Search products
// @Description Search products with advanced filters, sorting, and pagination
// @Tags products
// @Accept json
// @Produce json
// @Param q query string false "Search query for product name and description"
// @Param category query string false "Category slug for filtering"
// @Param min_price query number false "Minimum price filter" minimum(0)
// @Param max_price query number false "Maximum price filter" minimum(0)
// @Param sort_by query string false "Sort field (name, price, created_at, etc.)"
// @Param sort_order query string false "Sort order" Enums(ASC, DESC) default(DESC)
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Items per page" default(10) minimum(1) maximum(100)
// @Success 200 {object} dto.ProductListResponse "Success"
// @Failure 400 {object} map[string]interface{} "Bad Request"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /api/products/search [get]
func (c *ProductController) SearchProducts(ctx *fiber.Ctx) error {
	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "10"))
	query := ctx.Query("q")
	category := ctx.Query("category")
	minPrice := ctx.Query("min_price")
	maxPrice := ctx.Query("max_price")
	sortBy := ctx.Query("sort_by")
	sortOrder := ctx.Query("sort_order")

	products, total, err := c.productService.SearchProducts(query, category, minPrice, maxPrice, sortBy, sortOrder, page, limit)
	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"error": "Failed to search products",
		})
	}

	// Convert to response DTOs
	var productResponses []dto.ProductResponse
	for _, product := range products {
		// Generate image URL
		imageURL := ""
		if product.Image != "" {
			imageURL = fmt.Sprintf("/assets/images/Products/%s", product.Image)
		}

		productResponse := dto.ProductResponse{
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
			productResponse.CategoryModel = dto.CategoryResponse{
				ID:          product.CategoryModel.ID,
				Name:        product.CategoryModel.Name,
				Description: product.CategoryModel.Description,
				Slug:        product.CategoryModel.Slug,
				Active:      product.CategoryModel.Active,
			}
		}

		productResponses = append(productResponses, productResponse)
	}

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

// @Summary Get categories
// @Description Get list of all active categories
// @Tags categories
// @Accept json
// @Produce json
// @Success 200 {array} dto.CategoryResponse "Success"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /api/categories [get]
func (c *ProductController) GetCategories(ctx *fiber.Ctx) error {
	fmt.Println("Debug: GetCategories controller method called")

	categories, err := c.productService.GetCategories()
	if err != nil {
		fmt.Printf("Debug: Error getting categories: %v\n", err)
		return ctx.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch categories",
		})
	}

	fmt.Printf("Debug: Got %d categories from service\n", len(categories))

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

	fmt.Printf("Debug: Returning %d category responses\n", len(categoryResponses))
	return ctx.JSON(categoryResponses)
}

// @Summary Get products by category
// @Description Get products filtered by category ID with pagination
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID" minimum(1)
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Items per page" default(10) minimum(1) maximum(100)
// @Success 200 {object} dto.ProductListResponse "Success"
// @Failure 400 {object} map[string]interface{} "Bad Request - Invalid category ID"
// @Failure 404 {object} map[string]interface{} "Not Found - Category not found"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /api/categories/{id}/products [get]
func (c *ProductController) GetProductsByCategory(ctx *fiber.Ctx) error {
	id, err := strconv.ParseUint(ctx.Params("id"), 10, 32)
	if err != nil {
		return ctx.Status(400).JSON(fiber.Map{
			"error": "Invalid category ID",
		})
	}

	page, _ := strconv.Atoi(ctx.Query("page", "1"))
	limit, _ := strconv.Atoi(ctx.Query("limit", "10"))
	categoryID := uint(id)

	products, total, err := c.productService.GetProducts(page, limit, &categoryID)
	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch products",
		})
	}

	// Convert to response DTOs
	var productResponses []dto.ProductResponse
	for _, product := range products {
		// Generate image URL
		imageURL := ""
		if product.Image != "" {
			imageURL = fmt.Sprintf("/assets/images/Products/%s", product.Image)
		}

		productResponse := dto.ProductResponse{
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
			productResponse.CategoryModel = dto.CategoryResponse{
				ID:          product.CategoryModel.ID,
				Name:        product.CategoryModel.Name,
				Description: product.CategoryModel.Description,
				Slug:        product.CategoryModel.Slug,
				Active:      product.CategoryModel.Active,
			}
		}

		productResponses = append(productResponses, productResponse)
	}

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
