package database

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rizkyizh/go-fiber-boilerplate/app/models"
	"gorm.io/gorm"
)

// ProductData represents the structure from products.json
type ProductData struct {
	Index            int         `json:"Index"`
	Name             string      `json:"Name"`
	Description      string      `json:"Description"`
	Brand            string      `json:"Brand"`
	Category         string      `json:"Category"`
	Price            int         `json:"Price"`
	Currency         string      `json:"Currency"`
	Stock            int         `json:"Stock"`
	EAN              interface{} `json:"EAN"` // Can be string or number
	Color            string      `json:"Color"`
	Size             string      `json:"Size"`
	Availability     string      `json:"Availability"`
	ShortDescription string      `json:"ShortDescription"`
	Image            string      `json:"Image"`
	InternalID       string      `json:"Internal ID"`
}

// SeedProductsFromJSON seeds the database with products from the JSON file
func SeedProductsFromJSON() error {
	log.Println("Seeding products from JSON file...")

	// Read the JSON file
	jsonPath := filepath.Join("assets", "products.json")
	jsonData, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("error reading products.json: %v", err)
	}

	log.Printf("Read JSON file with %d bytes", len(jsonData))

	// Parse JSON data
	var productsData []ProductData
	if err := json.Unmarshal(jsonData, &productsData); err != nil {
		return fmt.Errorf("error parsing products.json: %v", productsData)
	}

	log.Printf("Parsed %d products from JSON", len(productsData))

	// Track operations
	categoriesCreated := 0
	productsCreated := 0
	productsErrored := 0

	// Create categories first (outside of main transaction)
	categoriesMap := make(map[string]uint)
	log.Println("Creating categories...")

	for _, productData := range productsData {
		if _, exists := categoriesMap[productData.Category]; !exists {
			category := models.Category{
				Name:        productData.Category,
				Description: fmt.Sprintf("Products in %s category", productData.Category),
				Slug:        generateUniqueSlug(productData.Category, "categories"),
				Active:      true,
			}

			// Create category in separate transaction
			result := DB.Where("name = ?", productData.Category).FirstOrCreate(&category)
			if result.Error != nil {
				log.Printf("Error creating category %s: %v", productData.Category, result.Error)
				continue
			}

			categoriesMap[productData.Category] = category.ID
			if result.RowsAffected > 0 {
				categoriesCreated++
				log.Printf("Created category: %s (ID: %d)", productData.Category, category.ID)
			}
		}
	}

	log.Printf("Categories setup complete. Created: %d", categoriesCreated)

	// Process products in batches to avoid large transactions
	batchSize := 100
	for i := 0; i < len(productsData); i += batchSize {
		end := i + batchSize
		if end > len(productsData) {
			end = len(productsData)
		}

		batch := productsData[i:end]
		log.Printf("Processing batch %d-%d of %d products...", i+1, end, len(productsData))

		// Process batch in transaction
		err := DB.Transaction(func(tx *gorm.DB) error {
			for j, productData := range batch {
				actualIndex := i + j + 1

				// Get category ID
				categoryID, exists := categoriesMap[productData.Category]
				if !exists {
					log.Printf("Category not found for product %s: %s", productData.Name, productData.Category)
					continue
				}

				// Handle EAN conversion
				var ean string
				switch v := productData.EAN.(type) {
				case string:
					ean = v
				case float64:
					ean = strconv.FormatInt(int64(v), 10)
				case int:
					ean = strconv.Itoa(v)
				case int64:
					ean = strconv.FormatInt(v, 10)
				default:
					ean = ""
				}

				// Generate unique values for constrained fields
				uniqueEAN := generateUniqueEAN(tx, ean, productData.Index)
				uniqueInternalID := generateUniqueInternalID(tx, productData.InternalID, productData.Index)
				uniqueSlug := generateUniqueProductSlug(tx, productData.Name, productData.Index)
				uniqueSKU := generateUniqueSKU(tx, productData.InternalID, productData.Index)

				// Create product
				product := models.Product{
					Index:            productData.Index,
					Name:             productData.Name,
					Description:      productData.Description,
					ShortDescription: productData.ShortDescription,
					Brand:            productData.Brand,
					Category:         productData.Category,
					Price:            float64(productData.Price),
					Currency:         productData.Currency,
					Stock:            productData.Stock,
					EAN:              uniqueEAN,
					Color:            productData.Color,
					Size:             productData.Size,
					Availability:     productData.Availability,
					Image:            productData.Image,
					InternalID:       uniqueInternalID,
					Slug:             uniqueSlug,
					SKU:              uniqueSKU,
					CategoryID:       categoryID,
					Active:           true,
				}

				// Create product
				if err := tx.Create(&product).Error; err != nil {
					log.Printf("Error creating product %s (Index: %d): %v", productData.Name, productData.Index, err)
					return err // This will rollback the batch
				}

				productsCreated++
				if actualIndex <= 20 || actualIndex%100 == 0 {
					log.Printf("✅ Created product %d: %s (Index: %d)", actualIndex, productData.Name, productData.Index)
				}
			}
			return nil
		})

		if err != nil {
			log.Printf("Error processing batch %d-%d: %v", i+1, end, err)
			productsErrored += len(batch)
			continue
		}
	}

	log.Printf("Product seeding completed!")
	log.Printf("📊 Final Results:")
	log.Printf("  📂 Categories created: %d", categoriesCreated)
	log.Printf("  ✅ Products created: %d", productsCreated)
	log.Printf("  ❌ Products with errors: %d", productsErrored)
	log.Printf("  📈 Total processed: %d", len(productsData))

	return nil
}

// Helper functions to generate unique values
func generateUniqueSlug(name, table string) string {
	baseSlug := generateSlug(name)
	slug := baseSlug
	counter := 1

	for {
		var count int64
		var err error

		if table == "categories" {
			err = DB.Model(&models.Category{}).Where("slug = ?", slug).Count(&count).Error
		} else {
			err = DB.Model(&models.Product{}).Where("slug = ?", slug).Count(&count).Error
		}

		if err != nil || count == 0 {
			break
		}
		slug = fmt.Sprintf("%s-%d", baseSlug, counter)
		counter++
	}
	return slug
}

func generateUniqueEAN(tx *gorm.DB, originalEAN string, index int) string {
	if originalEAN == "" {
		return fmt.Sprintf("EAN%010d", index+1000000000)
	}

	ean := originalEAN
	counter := 1
	for {
		var count int64
		if err := tx.Model(&models.Product{}).Where("ean = ?", ean).Count(&count).Error; err != nil || count == 0 {
			break
		}
		ean = fmt.Sprintf("%s-%d", originalEAN, counter)
		counter++
	}
	return ean
}

func generateUniqueInternalID(tx *gorm.DB, originalID string, index int) string {
	baseID := originalID
	if baseID == "" {
		baseID = fmt.Sprintf("PROD-%d", index)
	}

	internalID := baseID
	counter := 1
	for {
		var count int64
		if err := tx.Model(&models.Product{}).Where("internal_id = ?", internalID).Count(&count).Error; err != nil || count == 0 {
			break
		}
		internalID = fmt.Sprintf("%s-%d", baseID, counter)
		counter++
	}
	return internalID
}

func generateUniqueProductSlug(tx *gorm.DB, name string, index int) string {
	baseSlug := generateSlug(name)
	slug := baseSlug
	counter := 1

	for {
		var count int64
		if err := tx.Model(&models.Product{}).Where("slug = ?", slug).Count(&count).Error; err != nil || count == 0 {
			break
		}
		slug = fmt.Sprintf("%s-%d", baseSlug, counter)
		counter++
	}
	return slug
}

func generateUniqueSKU(tx *gorm.DB, originalID string, index int) string {
	baseSKU := originalID
	if baseSKU == "" {
		baseSKU = fmt.Sprintf("PROD-%d", index)
	}

	sku := baseSKU
	counter := 1
	for {
		var count int64
		if err := tx.Model(&models.Product{}).Where("sku = ?", sku).Count(&count).Error; err != nil || count == 0 {
			break
		}
		sku = fmt.Sprintf("%s-%d", baseSKU, counter)
		counter++
	}
	return sku
}

// ClearProductsData clears all products and categories from the database using GORM
func ClearProductsData() error {
	log.Println("Clearing products and categories data...")

	// First, let's check how many records we have
	var productCount, categoryCount int64
	if err := DB.Model(&models.Product{}).Count(&productCount).Error; err != nil {
		return fmt.Errorf("error counting products: %v", err)
	}
	if err := DB.Model(&models.Category{}).Count(&categoryCount).Error; err != nil {
		return fmt.Errorf("error counting categories: %v", err)
	}

	log.Printf("Found %d products and %d categories to delete", productCount, categoryCount)

	if productCount == 0 && categoryCount == 0 {
		log.Println("No data to clear")
		return nil
	}

	// Use transaction for data integrity
	return DB.Transaction(func(tx *gorm.DB) error {
		// Delete products first (this removes the foreign key references)
		log.Println("Deleting products...")

		// Use Unscoped().Delete to permanently delete records (bypass soft delete)
		result := tx.Unscoped().Where("1 = 1").Delete(&models.Product{})
		if result.Error != nil {
			return fmt.Errorf("error deleting products: %v", result.Error)
		}
		log.Printf("Successfully deleted %d products", result.RowsAffected)

		// Then delete all categories
		log.Println("Deleting categories...")
		result = tx.Unscoped().Where("1 = 1").Delete(&models.Category{})
		if result.Error != nil {
			return fmt.Errorf("error deleting categories: %v", result.Error)
		}
		log.Printf("Successfully deleted %d categories", result.RowsAffected)

		// Reset auto-increment sequences using GORM's native approach
		log.Println("Resetting auto-increment sequences...")
		if err := tx.Exec("ALTER SEQUENCE products_id_seq RESTART WITH 1").Error; err != nil {
			log.Printf("Warning: Could not reset products sequence: %v", err)
		}
		if err := tx.Exec("ALTER SEQUENCE categories_id_seq RESTART WITH 1").Error; err != nil {
			log.Printf("Warning: Could not reset categories sequence: %v", err)
		}

		return nil
	})
}

// VerifyDataCleared verifies that all data has been cleared
func VerifyDataCleared() error {
	var productCount, categoryCount int64

	if err := DB.Model(&models.Product{}).Count(&productCount).Error; err != nil {
		return fmt.Errorf("error verifying product count: %v", err)
	}
	if err := DB.Model(&models.Category{}).Count(&categoryCount).Error; err != nil {
		return fmt.Errorf("error verifying category count: %v", err)
	}

	log.Printf("Verification: %d products and %d categories remaining", productCount, categoryCount)

	if productCount > 0 || categoryCount > 0 {
		return fmt.Errorf("clear operation incomplete: %d products and %d categories still exist", productCount, categoryCount)
	}

	log.Println("✅ Data cleared successfully - all products and categories removed")
	return nil
}

// GetDataCounts returns the current count of products and categories
func GetDataCounts() (int64, int64, error) {
	var productCount, categoryCount int64

	if err := DB.Model(&models.Product{}).Count(&productCount).Error; err != nil {
		return 0, 0, fmt.Errorf("error counting products: %v", err)
	}
	if err := DB.Model(&models.Category{}).Count(&categoryCount).Error; err != nil {
		return 0, 0, fmt.Errorf("error counting categories: %v", err)
	}

	return productCount, categoryCount, nil
}

// generateSlug creates a URL-friendly slug from a string
func generateSlug(input string) string {
	// Convert to lowercase and replace spaces with hyphens
	slug := strings.ToLower(input)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "&", "and")
	slug = strings.ReplaceAll(slug, "amp;", "")

	// Remove special characters
	var result strings.Builder
	for _, char := range slug {
		if (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9') || char == '-' {
			result.WriteRune(char)
		}
	}

	// Remove multiple consecutive hyphens
	cleanSlug := strings.ReplaceAll(result.String(), "--", "-")
	cleanSlug = strings.Trim(cleanSlug, "-")

	return cleanSlug
}
