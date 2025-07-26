package services

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strconv"

	"gorm.io/gorm"

	"github.com/rizkyizh/go-fiber-boilerplate/app/models"
	"github.com/rizkyizh/go-fiber-boilerplate/database"
)

type StatisticsService struct {
	db *gorm.DB
}

func NewStatisticsService() *StatisticsService {
	return &StatisticsService{
		db: database.DB,
	}
}

// ProductStatistics holds all the calculated statistics
type ProductStatistics struct {
	TotalProducts     int64   `json:"total_products"`
	UniqueBrands      int64   `json:"unique_brands"`
	UniqueCategories  int64   `json:"unique_categories"`
	AveragePrice      float64 `json:"average_price"`
	PriceMin          float64 `json:"price_min"`
	PriceMax          float64 `json:"price_max"`
	InStockCount      int64   `json:"in_stock_count"`
	LimitedStockCount int64   `json:"limited_stock_count"`
	OutOfStockCount   int64   `json:"out_of_stock_count"`
}

// CalculateProductStatistics calculates all required product statistics
func (s *StatisticsService) CalculateProductStatistics() (*ProductStatistics, error) {
	stats := &ProductStatistics{}

	// Calculate total products
	var totalProducts int64
	if err := s.db.Model(&models.Product{}).Where("active = ?", true).Count(&totalProducts).Error; err != nil {
		return nil, fmt.Errorf("failed to count total products: %w", err)
	}
	stats.TotalProducts = totalProducts

	// Calculate unique brands
	var uniqueBrands int64
	if err := s.db.Model(&models.Product{}).Where("active = ? AND brand != ''", true).Distinct("brand").Count(&uniqueBrands).Error; err != nil {
		return nil, fmt.Errorf("failed to count unique brands: %w", err)
	}
	stats.UniqueBrands = uniqueBrands

	// Calculate unique categories
	var uniqueCategories int64
	if err := s.db.Model(&models.Product{}).Where("active = ? AND category != ''", true).Distinct("category").Count(&uniqueCategories).Error; err != nil {
		return nil, fmt.Errorf("failed to count unique categories: %w", err)
	}
	stats.UniqueCategories = uniqueCategories

	// Calculate price statistics (average, min, max)
	var priceStats struct {
		Avg float64
		Min float64
		Max float64
	}

	if err := s.db.Model(&models.Product{}).
		Where("active = ? AND price > 0", true).
		Select("AVG(price) as avg, MIN(price) as min, MAX(price) as max").
		Scan(&priceStats).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate price statistics: %w", err)
	}

	stats.AveragePrice = priceStats.Avg
	stats.PriceMin = priceStats.Min
	stats.PriceMax = priceStats.Max

	// Calculate stock availability counts
	var inStockCount int64
	if err := s.db.Model(&models.Product{}).Where("active = ? AND availability = ?", true, "in_stock").Count(&inStockCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count in_stock products: %w", err)
	}
	stats.InStockCount = inStockCount

	var limitedStockCount int64
	if err := s.db.Model(&models.Product{}).Where("active = ? AND availability = ?", true, "limited_stock").Count(&limitedStockCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count limited_stock products: %w", err)
	}
	stats.LimitedStockCount = limitedStockCount

	var outOfStockCount int64
	if err := s.db.Model(&models.Product{}).Where("active = ? AND availability = ?", true, "out_of_stock").Count(&outOfStockCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count out_of_stock products: %w", err)
	}
	stats.OutOfStockCount = outOfStockCount

	return stats, nil
}

// GenerateCSV generates CSV content from product statistics
func (s *StatisticsService) GenerateCSV() ([]byte, error) {
	stats, err := s.CalculateProductStatistics()
	if err != nil {
		return nil, fmt.Errorf("failed to calculate statistics: %w", err)
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	// Write CSV header
	if err := writer.Write([]string{"metric", "value"}); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write statistics data
	records := [][]string{
		{"total_products", strconv.FormatInt(stats.TotalProducts, 10)},
		{"unique_brands", strconv.FormatInt(stats.UniqueBrands, 10)},
		{"unique_categories", strconv.FormatInt(stats.UniqueCategories, 10)},
		{"average_price", fmt.Sprintf("%.2f", stats.AveragePrice)},
		{"price_min", fmt.Sprintf("%.2f", stats.PriceMin)},
		{"price_max", fmt.Sprintf("%.2f", stats.PriceMax)},
		{"in_stock_count", strconv.FormatInt(stats.InStockCount, 10)},
		{"limited_stock_count", strconv.FormatInt(stats.LimitedStockCount, 10)},
		{"out_of_stock_count", strconv.FormatInt(stats.OutOfStockCount, 10)},
	}

	for _, record := range records {
		if err := writer.Write(record); err != nil {
			return nil, fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("CSV writer error: %w", err)
	}

	return buf.Bytes(), nil
}
