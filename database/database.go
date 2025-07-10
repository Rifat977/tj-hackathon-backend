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
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	// Create pgxpool for raw queries
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("Error parsing pool config: %v", err)
	}

	poolConfig.MaxConns = 20
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	Pool, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalf("Error creating connection pool: %v", err)
	}

	log.Println("Database Connected successfully")

	// Connect to Redis
	Redis = redis.NewClient(&redis.Options{
		Addr:     config.AppConfig.REDIS_URL,
		Password: config.AppConfig.REDIS_PASSWORD,
		DB:       0,
		PoolSize: 10,
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

	log.Println("Search indexes created successfully")
}
