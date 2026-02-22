package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

/*
|--------------------------------------------------------------------------
| Typed sub-configs
|--------------------------------------------------------------------------
*/

type EmailConfig struct {
	Provider string `mapstructure:"provider" validate:"required"`
	Rate     int    `mapstructure:"rate" validate:"gt=0"`
	Burst    int    `mapstructure:"burst" validate:"gte=0"`
}

/*
|--------------------------------------------------------------------------
| Main Config (matches your existing fields)
|--------------------------------------------------------------------------
*/

type Config struct {
	DbUrl     string
	TestDbUrl string
	Port      string
	Env       string
	JwtSecret string

	// Monnify
	MonnifyApiKey       string
	MonnifySecretKey    string
	MonnifyContractCode string
	MonnifyBaseURL      string

	// Resend
	ResendApiKey string

	// Mailtrap
	MailtrapHost string
	MailtrapPass string
	MailtrapPort string
	MailtrapUser string

	// Redis
	RedisAddr string

	WebhookURL  string
	FrontendURL string
	LogLevel    string

	GoogleApplicationCredentials string

	// NEW (from config.yaml)
	Email EmailConfig `mapstructure:"email"`
}

/*
|--------------------------------------------------------------------------
| Global app config
|--------------------------------------------------------------------------
*/

var App Config

/*
|--------------------------------------------------------------------------
| Load config (Viper + .env)
|--------------------------------------------------------------------------
*/

func Load() {
	// 1️⃣ Load .env (optional in prod)
	_ = godotenv.Load()

	// 2️⃣ Viper setup
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".") // project root

	// ENV overrides YAML
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 3️⃣ Read config.yaml (optional but recommended)
	if err := viper.ReadInConfig(); err != nil {
		log.Println("config.yaml not found, relying on environment variables")
	}

	// 4️⃣ Load ENV-backed values (exactly like your old code)
	App = Config{
		DbUrl:                        mustEnv("DATABASE_URL"),
		TestDbUrl:                    getEnv("TEST_DATABASE_URL", "test"),
		Port:                         getEnv("PORT", "8080"),
		Env:                          getEnv("APP_ENV", "development"),
		JwtSecret:                    mustEnv("JWT_SECRET"),
		GoogleApplicationCredentials: mustEnv("GOOGLE_APPLICATION_CREDENTIALS"),

		MonnifyApiKey:       mustEnv("MONNIFY_API_KEY"),
		MonnifySecretKey:    mustEnv("MONNIFY_SECRET_KEY"),
		MonnifyContractCode: mustEnv("MONNIFY_CONTRACT_CODE"),
		MonnifyBaseURL:      mustEnv("MONNIFY_BASE_URL"),

		ResendApiKey: mustEnv("RESEND_API_KEY"),

		MailtrapHost: mustEnv("MAILTRAP_HOST"),
		MailtrapPort: mustEnv("MAILTRAP_PORT"),
		MailtrapUser: mustEnv("MAILTRAP_USER"),
		MailtrapPass: mustEnv("MAILTRAP_PASSWORD"),

		RedisAddr: mustEnv("REDIS_ADDR"),

		WebhookURL:  mustEnv("WEBHOOK_URL"),
		FrontendURL: mustEnv("FRONTEND_URL"),
		LogLevel:    getEnv("LOG_LEVEL", "1"),
	}

	// 5️⃣ Unmarshal YAML-only sections (non-secrets)
	if err := viper.Unmarshal(&App); err != nil {
		log.Fatalf("failed to unmarshal config.yaml: %v", err)
	}

	// 6️⃣ Validate config
	validateConfig(App)

	log.Println("✅ configuration loaded successfully")
}

/*
|--------------------------------------------------------------------------
| Validation
|--------------------------------------------------------------------------
*/

func validateConfig(cfg Config) {
	validate := validator.New()

	if err := validate.Struct(cfg); err != nil {
		panic(fmt.Errorf("config validation failed: %w", err))
	}

	// Extra guarantees
	if cfg.Email.Rate < 1 {
		panic("email.rate must be >= 1")
	}

	if cfg.Email.Burst < 1 {
		panic("email.burst must be >= 1")
	}
}

/*
|--------------------------------------------------------------------------
| Helpers
|--------------------------------------------------------------------------
*/

func getEnv(key, fallback string) string {
	if value := viper.GetString(key); value != "" {
		return value
	}
	return fallback
}

func mustEnv(key string) string {
	if value := viper.GetString(key); value != "" {
		return value
	}
	log.Fatalf("Environment variable %s is required but not set", key)
	return ""
}
