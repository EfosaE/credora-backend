package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DbUrl     string
	Port      string
	Env       string
	JwtSecret string

	// Monnify settings
	MonnifyApiKey       string
	MonnifySecretKey    string
	MonnifyContractCode string
	MonnifyBaseURL      string

	ResendApiKey string

	MailtrapHost string
	MailtrapPass string
	MailtrapPort string
	MailtrapUser string

	RedisAddr string
}

var App Config

func Load() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, relying on system environment variables")
	}

	App = Config{
		DbUrl:     mustGetEnv("DATABASE_URL"),
		Port:      getEnv("PORT", "8080"),
		Env:       getEnv("APP_ENV", "development"),
		JwtSecret: mustGetEnv("JWT_SECRET"),

		MonnifyApiKey:       mustGetEnv("MONNIFY_API_KEY"),
		MonnifySecretKey:    mustGetEnv("MONNIFY_SECRET_KEY"),
		MonnifyContractCode: mustGetEnv("MONNIFY_CONTRACT_CODE"),
		MonnifyBaseURL:      mustGetEnv("MONNIFY_BASE_URL"),

		// Resend
		ResendApiKey: mustGetEnv("RESEND_API_KEY"),

		// Mailtrap
		MailtrapHost: mustGetEnv("MAILTRAP_HOST"),
		MailtrapPort: mustGetEnv("MAILTRAP_PORT"),
		MailtrapUser: mustGetEnv("MAILTRAP_USER"),
		MailtrapPass: mustGetEnv("MAILTRAP_PASSWORD"),

		// Redis
		RedisAddr: mustGetEnv("REDIS_ADDR"),
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
