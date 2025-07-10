package dto

type ProductResponse struct {
	ID               uint             `json:"id"`
	Index            int              `json:"index"`
	Name             string           `json:"name"`
	Description      string           `json:"description"`
	ShortDescription string           `json:"short_description"`
	Brand            string           `json:"brand"`
	Category         string           `json:"category"`
	Price            float64          `json:"price"`
	Currency         string           `json:"currency"`
	Stock            int              `json:"stock"`
	EAN              string           `json:"ean"`
	Color            string           `json:"color"`
	Size             string           `json:"size"`
	Availability     string           `json:"availability"`
	Image            string           `json:"image"`
	ImageURL         string           `json:"image_url"`
	InternalID       string           `json:"internal_id"`
	Slug             string           `json:"slug"`
	SKU              string           `json:"sku"`
	CategoryModel    CategoryResponse `json:"category_model,omitempty"`
	Active           bool             `json:"active"`
	CreatedAt        string           `json:"created_at"`
	UpdatedAt        string           `json:"updated_at"`
}

type CategoryResponse struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Slug        string `json:"slug"`
	Active      bool   `json:"active"`
}

type ProductListResponse struct {
	Products   []ProductResponse `json:"products"`
	Pagination PaginationInfo    `json:"pagination"`
}

type PaginationInfo struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

type ProductSearchRequest struct {
	Query     string `query:"q"`
	Category  string `query:"category"`
	MinPrice  string `query:"min_price"`
	MaxPrice  string `query:"max_price"`
	SortBy    string `query:"sort_by"`
	SortOrder string `query:"sort_order"`
	Page      int    `query:"page"`
	Limit     int    `query:"limit"`
}

type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Database  string `json:"database"`
	Redis     string `json:"redis"`
}
