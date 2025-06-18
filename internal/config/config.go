package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DbUrl       string
	Port       string
	Env        string
	JwtSecret  string
}

var App Config

func Load() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, relying on system environment variables")
	}

	App = Config{
		DbUrl:  mustGetEnv("DATABASE_URL"),
		Port: getEnv("PORT", "8080"),
		Env:  getEnv("APP_ENV", "development"),
		// JwtSecret: mustGetEnv("JWT_SECRET"),
	}
}

// getEnv returns a fallback if variable not set
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// mustGetEnv panics if variable is not set
func mustGetEnv(key string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	log.Fatalf("Environment variable %s is required but not set", key)
	return ""
}
