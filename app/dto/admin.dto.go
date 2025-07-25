package dto

// CreateProductRequest represents the request to create a new product
type CreateProductRequest struct {
	Name             string  `json:"name" validate:"required"`
	Description      string  `json:"description"`
	ShortDescription string  `json:"short_description"`
	Brand            string  `json:"brand"`
	Category         string  `json:"category"`
	Price            float64 `json:"price" validate:"required,gt=0"`
	Currency         string  `json:"currency"`
	Stock            int     `json:"stock" validate:"gte=0"`
	EAN              string  `json:"ean"`
	Color            string  `json:"color"`
	Size             string  `json:"size"`
	Availability     string  `json:"availability"`
	Image            string  `json:"image"`
	InternalID       string  `json:"internal_id"`
	Slug             string  `json:"slug"`
	SKU              string  `json:"sku"`
	CategoryID       uint    `json:"category_id" validate:"required"`
	Active           bool    `json:"active"`
}

// UpdateProductRequest represents the request to update an existing product
type UpdateProductRequest struct {
	Name             *string  `json:"name"`
	Description      *string  `json:"description"`
	ShortDescription *string  `json:"short_description"`
	Brand            *string  `json:"brand"`
	Category         *string  `json:"category"`
	Price            *float64 `json:"price" validate:"omitempty,gt=0"`
	Currency         *string  `json:"currency"`
	Stock            *int     `json:"stock" validate:"omitempty,gte=0"`
	EAN              *string  `json:"ean"`
	Color            *string  `json:"color"`
	Size             *string  `json:"size"`
	Availability     *string  `json:"availability"`
	Image            *string  `json:"image"`
	InternalID       *string  `json:"internal_id"`
	Slug             *string  `json:"slug"`
	SKU              *string  `json:"sku"`
	CategoryID       *uint    `json:"category_id"`
	Active           *bool    `json:"active"`
}

// BulkUploadResult represents the result of a bulk upload operation
type BulkUploadResult struct {
	Uploaded              int      `json:"uploaded"`
	Failed                int      `json:"failed"`
	Errors                []string `json:"errors,omitempty"`
	ProcessingTimeSeconds float64  `json:"processing_time_seconds,omitempty"`
}

// AdminStats represents admin dashboard statistics
type AdminStats struct {
	TotalProducts    int64   `json:"total_products"`
	ActiveProducts   int64   `json:"active_products"`
	TotalCategories  int64   `json:"total_categories"`
	AveragePrice     float64 `json:"average_price"`
	LowStockProducts int64   `json:"low_stock_products"`
}
