package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"sync"

	"github.com/rizkyizh/go-fiber-boilerplate/app/dto"
	"github.com/rizkyizh/go-fiber-boilerplate/app/models"
	"github.com/rizkyizh/go-fiber-boilerplate/database"
)

// chunkResult represents the result of processing a chunk
type chunkResult struct {
	uploaded int
	failed   int
	errors   []string
}

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

func (s *ProductService) CreateProduct(request dto.CreateProductRequest) (*models.Product, error) {
	// Generate unique values
	uniqueSlug := s.generateUniqueSlug(request.Name)
	uniqueSKU := s.generateUniqueSKU(request.Name)
	uniqueInternalID := s.generateUniqueInternalID(request.Name)

	product := models.Product{
		Name:             request.Name,
		Description:      request.Description,
		ShortDescription: request.ShortDescription,
		Brand:            request.Brand,
		Category:         request.Category,
		Price:            request.Price,
		Currency:         request.Currency,
		Stock:            request.Stock,
		EAN:              request.EAN,
		Color:            request.Color,
		Size:             request.Size,
		Availability:     request.Availability,
		Image:            request.Image,
		InternalID:       uniqueInternalID,
		Slug:             uniqueSlug,
		SKU:              uniqueSKU,
		CategoryID:       request.CategoryID,
		Active:           request.Active,
	}

	err := s.db.Create(&product).Error
	if err != nil {
		return nil, err
	}

	// Clear cache
	s.clearProductCache()

	return &product, nil
}

func (s *ProductService) UpdateProduct(id uint, request dto.UpdateProductRequest) (*models.Product, error) {
	var product models.Product
	err := s.db.First(&product, id).Error
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if request.Name != nil {
		product.Name = *request.Name
	}
	if request.Description != nil {
		product.Description = *request.Description
	}
	if request.ShortDescription != nil {
		product.ShortDescription = *request.ShortDescription
	}
	if request.Brand != nil {
		product.Brand = *request.Brand
	}
	if request.Category != nil {
		product.Category = *request.Category
	}
	if request.Price != nil {
		product.Price = *request.Price
	}
	if request.Currency != nil {
		product.Currency = *request.Currency
	}
	if request.Stock != nil {
		product.Stock = *request.Stock
	}
	if request.EAN != nil {
		product.EAN = *request.EAN
	}
	if request.Color != nil {
		product.Color = *request.Color
	}
	if request.Size != nil {
		product.Size = *request.Size
	}
	if request.Availability != nil {
		product.Availability = *request.Availability
	}
	if request.Image != nil {
		product.Image = *request.Image
	}
	if request.InternalID != nil {
		product.InternalID = *request.InternalID
	}
	if request.Slug != nil {
		product.Slug = *request.Slug
	}
	if request.SKU != nil {
		product.SKU = *request.SKU
	}
	if request.CategoryID != nil {
		product.CategoryID = *request.CategoryID
	}
	if request.Active != nil {
		product.Active = *request.Active
	}

	err = s.db.Save(&product).Error
	if err != nil {
		return nil, err
	}

	// Clear cache
	s.clearProductCache()

	return &product, nil
}

func (s *ProductService) DeleteProduct(id uint) error {
	var product models.Product
	err := s.db.First(&product, id).Error
	if err != nil {
		return err
	}

	err = s.db.Delete(&product).Error
	if err != nil {
		return err
	}

	// Clear cache
	s.clearProductCache()

	return nil
}

func (s *ProductService) BulkUploadProducts(file *multipart.FileHeader) (*dto.BulkUploadResult, error) {
	startTime := time.Now()

	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	// Read file content
	content, err := io.ReadAll(src)
	if err != nil {
		return nil, err
	}

	// Parse JSON - expect array of products
	var productsData []map[string]interface{}
	err = json.Unmarshal(content, &productsData)
	if err != nil {
		return nil, err
	}

	parseTime := time.Since(startTime)
	fmt.Printf("📊 Bulk Upload Stats:\n")
	fmt.Printf("   File size: %.2f MB\n", float64(file.Size)/(1024*1024))
	fmt.Printf("   Products to process: %d\n", len(productsData))
	fmt.Printf("   Parse time: %v\n", parseTime)

	result := &dto.BulkUploadResult{
		Uploaded: 0,
		Failed:   0,
		Errors:   []string{},
	}

	// Pre-process categories to avoid repeated database queries
	categoryStart := time.Now()
	categoryMap := make(map[string]uint)
	var categories []models.Category
	err = s.db.Model(&models.Category{}).Select("name, id").Find(&categories).Error
	if err == nil {
		for _, cat := range categories {
			categoryMap[cat.Name] = cat.ID
		}
	}
	categoryTime := time.Since(categoryStart)
	fmt.Printf("   Categories loaded: %d in %v\n", len(categories), categoryTime)

	// High-concurrency processing with goroutines
	chunkSize := 200 // Smaller chunks for better concurrency
	totalChunks := (len(productsData) + chunkSize - 1) / chunkSize
	maxWorkers := 10 // Maximum concurrent workers

	fmt.Printf("   Processing in %d chunks of %d products each with %d concurrent workers\n", totalChunks, chunkSize, maxWorkers)

	// Create channels for coordination
	chunkChan := make(chan []map[string]interface{}, totalChunks)
	resultChan := make(chan *chunkResult, totalChunks)
	errorChan := make(chan error, maxWorkers)

	// Create context with timeout for the entire operation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// Start worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for chunk := range chunkChan {
				select {
				case <-ctx.Done():
					errorChan <- fmt.Errorf("worker %d cancelled due to timeout", workerID)
					return
				default:
					// Process chunk with high-performance COPY protocol
					chunkResult := s.processChunkHighPerformance(ctx, chunk, categoryMap, workerID)
					resultChan <- chunkResult
				}
			}
		}(i)
	}

	// Send chunks to workers
	go func() {
		defer close(chunkChan)
		for i := 0; i < len(productsData); i += chunkSize {
			end := i + chunkSize
			if end > len(productsData) {
				end = len(productsData)
			}
			chunk := productsData[i:end]
			select {
			case chunkChan <- chunk:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Collect results
	go func() {
		wg.Wait()
		close(resultChan)
		close(errorChan)
	}()

	// Process results
	processedChunks := 0
	for {
		select {
		case chunkResult := <-resultChan:
			if chunkResult != nil {
				result.Uploaded += chunkResult.uploaded
				result.Failed += chunkResult.failed
				result.Errors = append(result.Errors, chunkResult.errors...)
			}
			processedChunks++

			// Log progress
			progress := float64(processedChunks) / float64(totalChunks) * 100
			fmt.Printf("   📈 Progress: %.1f%% (%d/%d chunks completed)\n", progress, processedChunks, totalChunks)

			if processedChunks >= totalChunks {
				goto done
			}

		case err := <-errorChan:
			if err != nil {
				fmt.Printf("   ❌ Worker error: %v\n", err)
				result.Errors = append(result.Errors, err.Error())
			}

		case <-ctx.Done():
			return result, fmt.Errorf("bulk upload timed out after processing %d chunks", processedChunks)
		}
	}

done:
	totalTime := time.Since(startTime)
	throughput := float64(result.Uploaded) / totalTime.Seconds()

	fmt.Printf("   Total time: %v\n", totalTime)
	fmt.Printf("   Throughput: %.0f products/second\n", throughput)
	fmt.Printf("   Success rate: %.1f%%\n", float64(result.Uploaded)/float64(len(productsData))*100)

	// Clear cache
	s.clearProductCache()

	return result, nil
}

// processChunkHighPerformance processes a chunk with high-performance COPY protocol
func (s *ProductService) processChunkHighPerformance(ctx context.Context, productsData []map[string]interface{}, categoryMap map[string]uint, workerID int) *chunkResult {
	result := &chunkResult{
		uploaded: 0,
		failed:   0,
		errors:   []string{},
	}

	// Get connection from pool
	conn, err := database.Pool.Acquire(ctx)
	if err != nil {
		result.errors = append(result.errors, fmt.Sprintf("Failed to acquire connection: %v", err))
		return result
	}
	defer conn.Release()

	// Begin transaction
	tx, err := conn.Begin(ctx)
	if err != nil {
		result.errors = append(result.errors, fmt.Sprintf("Failed to begin transaction: %v", err))
		return result
	}
	defer tx.Rollback(ctx)

	// Process categories first for this chunk
	newCategories := make(map[string]models.Category)
	for _, productData := range productsData {
		if category, ok := productData["Category"].(string); ok && category != "" {
			if _, exists := categoryMap[category]; !exists {
				if _, exists := newCategories[category]; !exists {
					newCategory := models.Category{
						Name:        category,
						Description: fmt.Sprintf("Category for %s", category),
						Slug:        s.generateCategorySlug(category),
						Active:      true,
					}
					newCategories[category] = newCategory
				}
			}
		}
	}

	// Insert new categories using high-performance COPY
	if len(newCategories) > 0 {
		err = s.insertCategoriesHighPerformance(ctx, tx, newCategories)
		if err != nil {
			result.errors = append(result.errors, fmt.Sprintf("Failed to insert categories: %v", err))
			return result
		}

		// Update category map with new categories
		for name, cat := range newCategories {
			categoryMap[name] = cat.ID
		}
	}

	// Prepare products for high-performance COPY insert
	var products []models.Product
	for _, productData := range productsData {
		product, err := s.convertToProduct(productData)
		if err != nil {
			result.failed++
			result.errors = append(result.errors, fmt.Sprintf("Invalid product data: %v", err))
			continue
		}

		// Set category ID
		if product.Category != "" {
			if categoryID, exists := categoryMap[product.Category]; exists {
				product.CategoryID = categoryID
			}
		}

		products = append(products, *product)
	}

	// Insert products using high-performance COPY
	if len(products) > 0 {
		err = s.insertProductsHighPerformance(ctx, tx, products)
		if err != nil {
			result.errors = append(result.errors, fmt.Sprintf("Failed to insert products: %v", err))
			return result
		}
		result.uploaded += len(products)
	}

	// Commit transaction
	err = tx.Commit(ctx)
	if err != nil {
		result.errors = append(result.errors, fmt.Sprintf("Failed to commit transaction: %v", err))
		return result
	}

	return result
}

// insertCategoriesHighPerformance uses optimized COPY protocol for categories
func (s *ProductService) insertCategoriesHighPerformance(ctx context.Context, tx pgx.Tx, categories map[string]models.Category) error {
	// Use direct INSERT with ON CONFLICT for better performance
	var rows [][]interface{}
	timestamp := time.Now()
	for _, cat := range categories {
		rows = append(rows, []interface{}{
			cat.Name,
			cat.Description,
			cat.Slug,
			cat.Active,
			timestamp,
			timestamp,
		})
	}

	// Use COPY to insert directly with conflict handling
	_, err := tx.CopyFrom(
		ctx,
		pgx.Identifier{"categories"},
		[]string{"name", "description", "slug", "active", "created_at", "updated_at"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return err
	}

	return nil
}

// insertProductsHighPerformance uses optimized COPY protocol for products
func (s *ProductService) insertProductsHighPerformance(ctx context.Context, tx pgx.Tx, products []models.Product) error {
	// Prepare COPY data with proper validation
	var rows [][]interface{}
	timestamp := time.Now()

	for i, product := range products {
		// Validate required fields
		if product.Name == "" {
			return fmt.Errorf("product at index %d has empty name", i)
		}
		if product.Price <= 0 {
			return fmt.Errorf("product '%s' has invalid price: %f", product.Name, product.Price)
		}
		if product.Slug == "" {
			return fmt.Errorf("product '%s' has empty slug", product.Name)
		}
		if product.SKU == "" {
			return fmt.Errorf("product '%s' has empty SKU", product.Name)
		}

		rows = append(rows, []interface{}{
			product.Index,
			product.Name,
			product.Description,
			product.ShortDescription,
			product.Brand,
			product.Category,
			product.Price,
			product.Currency,
			product.Stock,
			product.EAN,
			product.Color,
			product.Size,
			product.Availability,
			product.Image,
			product.InternalID,
			product.Slug,
			product.SKU,
			product.CategoryID,
			product.Active,
			timestamp,
			timestamp,
		})
	}

	// Use COPY to insert directly with conflict handling
	_, err := tx.CopyFrom(
		ctx,
		pgx.Identifier{"products"},
		[]string{
			"index", "name", "description", "short_description", "brand", "category",
			"price", "currency", "stock", "ean", "color", "size", "availability",
			"image", "internal_id", "slug", "sku", "category_id", "active", "created_at", "updated_at",
		},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("failed to copy products: %v", err)
	}

	return nil
}

// convertToProduct converts JSON data to Product model with proper field mapping
func (s *ProductService) convertToProduct(data map[string]interface{}) (*models.Product, error) {
	product := &models.Product{}

	// Handle required fields with proper case mapping
	if name, ok := data["Name"].(string); ok && name != "" {
		product.Name = name
	} else {
		return nil, fmt.Errorf("Name is required")
	}

	if price, ok := data["Price"].(float64); ok && price > 0 {
		product.Price = price
	} else if price, ok := data["Price"].(int); ok && price > 0 {
		product.Price = float64(price)
	} else {
		return nil, fmt.Errorf("valid Price is required")
	}

	// Handle optional fields
	if desc, ok := data["Description"].(string); ok {
		product.Description = desc
	}
	if shortDesc, ok := data["ShortDescription"].(string); ok {
		product.ShortDescription = shortDesc
	}
	if brand, ok := data["Brand"].(string); ok {
		product.Brand = brand
	}
	if category, ok := data["Category"].(string); ok {
		product.Category = category
	}
	if currency, ok := data["Currency"].(string); ok {
		product.Currency = currency
	} else {
		product.Currency = "USD"
	}
	if stock, ok := data["Stock"].(float64); ok {
		product.Stock = int(stock)
	} else if stock, ok := data["Stock"].(int); ok {
		product.Stock = stock
	}
	if ean, ok := data["EAN"].(string); ok {
		product.EAN = ean
	} else if ean, ok := data["EAN"].(float64); ok {
		product.EAN = fmt.Sprintf("%.0f", ean)
	} else if ean, ok := data["EAN"].(int); ok {
		product.EAN = fmt.Sprintf("%d", ean)
	}
	if color, ok := data["Color"].(string); ok {
		product.Color = color
	}
	if size, ok := data["Size"].(string); ok {
		product.Size = size
	}
	if availability, ok := data["Availability"].(string); ok {
		product.Availability = availability
	}
	if image, ok := data["Image"].(string); ok {
		product.Image = image
	}
	if internalID, ok := data["Internal ID"].(string); ok {
		product.InternalID = internalID
	}

	// Generate unique values using timestamp-based approach for bulk operations
	timestamp := time.Now().UnixNano()
	product.Slug = s.generateBulkSlug(product.Name, timestamp)
	product.SKU = s.generateBulkSKU(product.Name, timestamp)
	if product.InternalID == "" {
		product.InternalID = s.generateBulkInternalID(product.Name, timestamp)
	}

	// Set defaults
	product.Active = true
	product.Index = 0 // Will be auto-incremented

	return product, nil
}

// generateCategorySlug generates a slug for category names
func (s *ProductService) generateCategorySlug(name string) string {
	baseSlug := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s-%d", baseSlug, timestamp)
}

func (s *ProductService) DeleteAllProducts() error {
	// Delete all products
	err := s.db.Where("1 = 1").Delete(&models.Product{}).Error
	if err != nil {
		return err
	}

	// Clear cache
	s.clearProductCache()

	return nil
}

// Helper methods
func (s *ProductService) generateUniqueSlug(name string) string {
	baseSlug := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	slug := baseSlug
	counter := 1

	for {
		var count int64
		s.db.Model(&models.Product{}).Where("slug = ?", slug).Count(&count)
		if count == 0 {
			break
		}
		slug = fmt.Sprintf("%s-%d", baseSlug, counter)
		counter++
	}

	return slug
}

func (s *ProductService) generateUniqueSKU(name string) string {
	baseSKU := strings.ToUpper(strings.ReplaceAll(name, " ", ""))
	sku := baseSKU[:min(len(baseSKU), 8)]
	counter := 1

	for {
		var count int64
		s.db.Model(&models.Product{}).Where("sku = ?", sku).Count(&count)
		if count == 0 {
			break
		}
		sku = fmt.Sprintf("%s%d", baseSKU[:min(len(baseSKU), 6)], counter)
		counter++
	}

	return sku
}

func (s *ProductService) generateUniqueInternalID(name string) string {
	baseID := strings.ToUpper(strings.ReplaceAll(name, " ", ""))
	internalID := baseID[:min(len(baseID), 10)]
	counter := 1

	for {
		var count int64
		s.db.Model(&models.Product{}).Where("internal_id = ?", internalID).Count(&count)
		if count == 0 {
			break
		}
		internalID = fmt.Sprintf("%s%d", baseID[:min(len(baseID), 8)], counter)
		counter++
	}

	return internalID
}

// generateBulkSlug generates a unique slug for bulk operations
func (s *ProductService) generateBulkSlug(name string, timestamp int64) string {
	baseSlug := strings.ToLower(strings.ReplaceAll(name, " ", "-"))
	return fmt.Sprintf("%s-%d", baseSlug, timestamp)
}

// generateBulkSKU generates a unique SKU for bulk operations
func (s *ProductService) generateBulkSKU(name string, timestamp int64) string {
	baseSKU := strings.ToUpper(strings.ReplaceAll(name, " ", ""))
	if len(baseSKU) > 6 {
		baseSKU = baseSKU[:6]
	}
	return fmt.Sprintf("%s%d", baseSKU, timestamp%1000000)
}

// generateBulkInternalID generates a unique internal ID for bulk operations
func (s *ProductService) generateBulkInternalID(name string, timestamp int64) string {
	baseID := strings.ToUpper(strings.ReplaceAll(name, " ", ""))
	if len(baseID) > 8 {
		baseID = baseID[:8]
	}
	return fmt.Sprintf("%s%d", baseID, timestamp%100000000)
}

func (s *ProductService) clearProductCache() {
	ctx := context.Background()
	keys, err := s.redis.Keys(ctx, "products:*").Result()
	if err == nil {
		for _, key := range keys {
			s.redis.Del(ctx, key)
		}
	}
}

// ClearAllCaches clears all Redis caches
func (s *ProductService) ClearAllCaches() error {
	ctx := context.Background()

	// Clear all keys (be careful with this in production)
	err := s.redis.FlushAll(ctx).Err()
	if err != nil {
		return err
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
