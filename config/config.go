package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

type Config struct {
	DB_URL         string
	REDIS_URL      string
	REDIS_PASSWORD string
	JWT_SECRET     string

	// Timeout configurations for high-performance bulk operations
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	BodyLimit    int64

	// Database timeout configurations
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration
	DBConnMaxIdleTime time.Duration

	// Redis timeout configurations
	RedisDialTimeout  time.Duration
	RedisReadTimeout  time.Duration
	RedisWriteTimeout time.Duration
	RedisPoolTimeout  time.Duration
}

var AppConfig Config

func LoadConfig() {
	AppConfig = Config{
		DB_URL:         os.Getenv("DATABASE_URL"),
		REDIS_URL:      os.Getenv("REDIS_URL"),
		REDIS_PASSWORD: os.Getenv("REDIS_PASSWORD"),
		JWT_SECRET:     os.Getenv("JWT_SECRET"),

		// HTTP Server timeouts - optimized for bulk uploads
		ReadTimeout:  getDurationEnv("READ_TIMEOUT", 10*time.Minute),  // Increased to 10 minutes for large file reads
		WriteTimeout: getDurationEnv("WRITE_TIMEOUT", 15*time.Minute), // Increased to 15 minutes for bulk operations
		IdleTimeout:  getDurationEnv("IDLE_TIMEOUT", 20*time.Minute),  // Increased to 20 minutes idle
		BodyLimit:    getInt64Env("BODY_LIMIT", 500*1024*1024),        // Increased to 500MB for large JSON files

		// Database connection pool - optimized for high concurrency
		DBMaxOpenConns:    getIntEnv("DB_MAX_OPEN_CONNS", 200), // Increased for bulk operations
		DBMaxIdleConns:    getIntEnv("DB_MAX_IDLE_CONNS", 50),  // More idle connections
		DBConnMaxLifetime: getDurationEnv("DB_CONN_MAX_LIFETIME", 2*time.Hour),
		DBConnMaxIdleTime: getDurationEnv("DB_CONN_MAX_IDLE_TIME", 1*time.Hour),

		// Redis timeouts - optimized for caching
		RedisDialTimeout:  getDurationEnv("REDIS_DIAL_TIMEOUT", 10*time.Second),
		RedisReadTimeout:  getDurationEnv("REDIS_READ_TIMEOUT", 5*time.Second),
		RedisWriteTimeout: getDurationEnv("REDIS_WRITE_TIMEOUT", 5*time.Second),
		RedisPoolTimeout:  getDurationEnv("REDIS_POOL_TIMEOUT", 10*time.Second),
	}

	if AppConfig.DB_URL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	if AppConfig.REDIS_URL == "" {
		AppConfig.REDIS_URL = "localhost:6379"
	}

	if AppConfig.JWT_SECRET == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	// Log timeout configurations
	log.Printf("📊 Timeout Configuration:")
	log.Printf("   HTTP Read Timeout: %v", AppConfig.ReadTimeout)
	log.Printf("   HTTP Write Timeout: %v", AppConfig.WriteTimeout)
	log.Printf("   HTTP Idle Timeout: %v", AppConfig.IdleTimeout)
	log.Printf("   Body Limit: %d MB", AppConfig.BodyLimit/(1024*1024))
	log.Printf("   DB Max Open Conns: %d", AppConfig.DBMaxOpenConns)
	log.Printf("   DB Max Idle Conns: %d", AppConfig.DBMaxIdleConns)
}

// Helper functions to parse environment variables with defaults
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
		log.Printf("Warning: Invalid duration format for %s, using default: %v", key, defaultValue)
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Printf("Warning: Invalid integer format for %s, using default: %d", key, defaultValue)
	}
	return defaultValue
}

func getInt64Env(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
		log.Printf("Warning: Invalid integer format for %s, using default: %d", key, defaultValue)
	}
	return defaultValue
}
