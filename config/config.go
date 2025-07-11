package config

import (
	"log"
	"os"
)

type Config struct {
	DB_URL         string
	REDIS_URL      string
	REDIS_PASSWORD string
	JWT_SECRET     string
}

var AppConfig Config

func LoadConfig() {
	AppConfig = Config{
		DB_URL:         os.Getenv("DATABASE_URL"),
		REDIS_URL:      os.Getenv("REDIS_URL"),
		REDIS_PASSWORD: os.Getenv("REDIS_PASSWORD"),
		JWT_SECRET:     os.Getenv("JWT_SECRET"),
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
}
