package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/rizkyizh/go-fiber-boilerplate/app/models"
	"github.com/rizkyizh/go-fiber-boilerplate/database"
)

type ProductService struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewProductService() *ProductService {
	return &ProductService{
		db:    database.DB,
		redis: database.Redis,
	}
}

func (s *ProductService) GetProducts(page, limit int, categoryID *uint) ([]models.Product, int64, error) {
	cacheKey := fmt.Sprintf("products:page:%d:limit:%d", page, limit)
	if categoryID != nil {
		cacheKey += fmt.Sprintf(":category:%d", *categoryID)
	}

	// Try to get from cache
	ctx := context.Background()
	cached, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var products []models.Product
		json.Unmarshal([]byte(cached), &products)
		return products, 0, nil
	}

	// Optimize query with specific field selection
	query := s.db.Model(&models.Product{}).
		Select("id, index, name, description, short_description, brand, category, price, currency, stock, ean, color, size, availability, image, internal_id, slug, sku, category_id, active, created_at, updated_at").
		Where("active = ?", true).
		Preload("CategoryModel", "active = ?", true)

	if categoryID != nil {
		query = query.Where("category_id = ?", *categoryID)
	}

	var total int64
	query.Count(&total)

	offset := (page - 1) * limit
	var products []models.Product
	err = query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&products).Error
	if err != nil {
		return nil, 0, err
	}

	// Cache for 5 minutes
	if data, err := json.Marshal(products); err == nil {
		s.redis.Set(ctx, cacheKey, data, 5*time.Minute)
	}

	return products, total, nil
}

func (s *ProductService) GetProductByID(id uint) (*models.Product, error) {
	cacheKey := fmt.Sprintf("product:%d", id)

	// Try to get from cache
	ctx := context.Background()
	cached, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var product models.Product
		json.Unmarshal([]byte(cached), &product)
		return &product, nil
	}

	// Optimize query with specific field selection
	var product models.Product
	err = s.db.Select("id, index, name, description, short_description, brand, category, price, currency, stock, ean, color, size, availability, image, internal_id, slug, sku, category_id, active, created_at, updated_at").
		Preload("CategoryModel", "active = ?", true).
		First(&product, id).Error
	if err != nil {
		return nil, err
	}

	// Cache for 10 minutes
	if data, err := json.Marshal(product); err == nil {
		s.redis.Set(ctx, cacheKey, data, 10*time.Minute)
	}

	return &product, nil
}

func (s *ProductService) SearchProducts(query, category, minPrice, maxPrice, sortBy, sortOrder string, page, limit int) ([]models.Product, int64, error) {
	cacheKey := fmt.Sprintf("search:%s:%s:%s:%s:%s:%s:%d:%d", query, category, minPrice, maxPrice, sortBy, sortOrder, page, limit)

	// Try to get from cache
	ctx := context.Background()
	cached, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var products []models.Product
		json.Unmarshal([]byte(cached), &products)
		return products, 0, nil
	}

	// Optimize query with specific field selection
	dbQuery := s.db.Model(&models.Product{}).
		Select("id, index, name, description, short_description, brand, category, price, currency, stock, ean, color, size, availability, image, internal_id, slug, sku, category_id, active, created_at, updated_at").
		Where("active = ?", true).
		Preload("CategoryModel", "active = ?", true)

	// Full-text search
	if query != "" {
		dbQuery = dbQuery.Where("to_tsvector('english', name || ' ' || description) @@ plainto_tsquery('english', ?)", query)
	}

	// Category filter
	if category != "" {
		dbQuery = dbQuery.Joins("JOIN categories ON products.category_id = categories.id").
			Where("categories.slug = ?", category)
	}

	// Price filters
	if minPrice != "" {
		if price, err := strconv.ParseFloat(minPrice, 64); err == nil {
			dbQuery = dbQuery.Where("price >= ?", price)
		}
	}
	if maxPrice != "" {
		if price, err := strconv.ParseFloat(maxPrice, 64); err == nil {
			dbQuery = dbQuery.Where("price <= ?", price)
		}
	}

	// Sorting
	if sortBy != "" {
		order := "ASC"
		if strings.ToUpper(sortOrder) == "DESC" {
			order = "DESC"
		}
		dbQuery = dbQuery.Order(fmt.Sprintf("%s %s", sortBy, order))
	} else {
		dbQuery = dbQuery.Order("created_at DESC")
	}

	var total int64
	dbQuery.Count(&total)

	offset := (page - 1) * limit
	var products []models.Product
	err = dbQuery.Offset(offset).Limit(limit).Find(&products).Error
	if err != nil {
		return nil, 0, err
	}

	// Cache for 2 minutes
	if data, err := json.Marshal(products); err == nil {
		s.redis.Set(ctx, cacheKey, data, 2*time.Minute)
	}

	return products, total, nil
}

func (s *ProductService) GetCategories() ([]models.Category, error) {
	cacheKey := "categories"

	// Try to get from cache
	ctx := context.Background()
	cached, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var categories []models.Category
		json.Unmarshal([]byte(cached), &categories)
		return categories, nil
	}

	// Debug: Check if there are any categories at all (without active filter)
	var allCategories []models.Category
	err = s.db.Find(&allCategories).Error
	if err != nil {
		return nil, err
	}

	// Debug: Log the count of all categories
	fmt.Printf("Debug: Found %d total categories in database\n", len(allCategories))

	// Debug: Log the first few categories to see their active status
	for i, cat := range allCategories {
		if i < 5 { // Only log first 5
			fmt.Printf("Debug: Category %d: ID=%d, Name=%s, Active=%t\n", i, cat.ID, cat.Name, cat.Active)
		}
	}

	// Optimize query with specific field selection
	var categories []models.Category
	err = s.db.Select("id, name, description, slug, active, created_at, updated_at").
		Where("active = ?", true).
		Find(&categories).Error
	if err != nil {
		return nil, err
	}

	// Debug: Log the count of active categories
	fmt.Printf("Debug: Found %d active categories\n", len(categories))

	// Cache for 30 minutes
	if data, err := json.Marshal(categories); err == nil {
		s.redis.Set(ctx, cacheKey, data, 30*time.Minute)
	}

	return categories, nil
}

func (s *ProductService) GetCategoryByID(id uint) (*models.Category, error) {
	var category models.Category
	err := s.db.Preload("Products", "active = ?", true).First(&category, id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}
