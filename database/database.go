package database

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/rizkyizh/go-fiber-boilerplate/app/models"
	"github.com/rizkyizh/go-fiber-boilerplate/config"
)

var (
	DB    *gorm.DB
	Pool  *pgxpool.Pool
	Redis *redis.Client
)

func ConnectDB() {
	var err error
	dsn := config.AppConfig.DB_URL

	// Optimize GORM configuration
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt: true, // Enable prepared statements for better performance
		DryRun:      false,
	})
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	// Get underlying SQL DB for connection pool configuration
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("Error getting SQL DB: %v", err)
	}

	// Configure connection pool for high concurrency
	sqlDB.SetMaxOpenConns(config.AppConfig.DBMaxOpenConns)
	sqlDB.SetMaxIdleConns(config.AppConfig.DBMaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.AppConfig.DBConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(config.AppConfig.DBConnMaxIdleTime)

	// Create pgxpool for raw queries with better configuration
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("Error parsing pool config: %v", err)
	}

	poolConfig.MaxConns = int32(config.AppConfig.DBMaxOpenConns)
	poolConfig.MinConns = int32(config.AppConfig.DBMaxIdleConns / 2) // Half of max idle
	poolConfig.MaxConnLifetime = config.AppConfig.DBConnMaxLifetime
	poolConfig.MaxConnIdleTime = config.AppConfig.DBConnMaxIdleTime
	poolConfig.HealthCheckPeriod = 30 * time.Second

	Pool, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalf("Error creating connection pool: %v", err)
	}

	log.Println("Database Connected successfully")

	// Connect to Redis with optimized settings
	Redis = redis.NewClient(&redis.Options{
		Addr:         config.AppConfig.REDIS_URL,
		Password:     config.AppConfig.REDIS_PASSWORD,
		DB:           0,
		PoolSize:     20, // Increase pool size for high concurrency
		MinIdleConns: 5,  // Keep minimum idle connections
		MaxRetries:   3,
		DialTimeout:  config.AppConfig.RedisDialTimeout,
		ReadTimeout:  config.AppConfig.RedisReadTimeout,
		WriteTimeout: config.AppConfig.RedisWriteTimeout,
		PoolTimeout:  config.AppConfig.RedisPoolTimeout,
	})

	// Test Redis connection
	ctx := context.Background()
	_, err = Redis.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Error connecting to Redis: %v", err)
	}

	log.Println("Redis Connected successfully")

	// Run migrations
	err = DB.AutoMigrate(
		&models.User{},
		&models.UserProfile{},
		&models.Category{},
		&models.Product{},
	)
	if err != nil {
		log.Fatalf("Error AutoMigrate database: %v", err)
	}

	// Run custom migrations
	RunMigrations()

	// Create full-text search indexes
	createSearchIndexes()

	// Note: Database seeding is now done via API endpoints
	// Use POST /api/seed/products for full product catalog from JSON
	// Use DELETE /api/seed/clear to clear all products and categories
}

func createSearchIndexes() {
	// Create full-text search index for products
	DB.Exec(`
		CREATE INDEX IF NOT EXISTS idx_products_search 
		ON products USING gin(to_tsvector('english', name || ' ' || description))
	`)

	// Create full-text search index for categories
	DB.Exec(`
		CREATE INDEX IF NOT EXISTS idx_categories_search 
		ON categories USING gin(to_tsvector('english', name || ' ' || description))
	`)

	// Create additional performance indexes
	DB.Exec(`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_active_category ON products(active, category_id)`)
	DB.Exec(`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_price_active ON products(price, active)`)
	DB.Exec(`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_brand_active ON products(brand, active)`)
	DB.Exec(`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_products_created_at_active ON products(created_at DESC, active)`)
	DB.Exec(`CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_categories_slug_active ON categories(slug, active)`)

	log.Println("Search indexes created successfully")
}
