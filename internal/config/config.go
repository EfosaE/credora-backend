package config

import (
	"fmt"
	"log"
	"math"
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
	Rate     int    `mapstructure:"rate"     validate:"gt=0"`
	Burst    int    `mapstructure:"burst"    validate:"gt=0"`
}

type JobConfig struct {
	ServiceTimeSec       float64 `mapstructure:"processing_rate_per_job" validate:"gt=0"`
	SafetyFactor         float64 `mapstructure:"safety_factor"           validate:"gt=0,lte=1"`
	BacklogWindowSeconds int     `mapstructure:"backlog_window_seconds"  validate:"gt=0"`
	QueueName            string  `mapstructure:"queue_name"              validate:"required"`

	HWCeilingPerSec    float64 `mapstructure:"-" validate:"gt=0"`
	RateLimitPerMinute int     `mapstructure:"-" validate:"gt=0"`
	QueueMaxSize       int     `mapstructure:"-" validate:"gt=0"`
}

type WorkerConfig struct {
	ConcurrencyLimit int  `mapstructure:"concurrency_limit" validate:"gt=0"`
	Replicas         int  `mapstructure:"replicas"          validate:"gt=0"`
	StrictPriority   bool `mapstructure:"strict_priority"`
}

type MachineConfig struct {
	CPUCores int    `mapstructure:"cpu_cores" validate:"gt=0"`
	JobBound string `mapstructure:"job_bound" validate:"oneof=cpu io"`
}

/*
|--------------------------------------------------------------------------
| Main Config
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

	// YAML-backed sections (non-secrets)
	Email   EmailConfig   `mapstructure:"email"`
	Job     JobConfig     `mapstructure:"job"`
	Worker  WorkerConfig  `mapstructure:"worker"`
	Machine MachineConfig `mapstructure:"machine"`
}

/*
|--------------------------------------------------------------------------
| Global app config — still accessible as config.App from other packages
|--------------------------------------------------------------------------
*/

var App Config

/*
|--------------------------------------------------------------------------
| Load
|--------------------------------------------------------------------------
*/

// Load reads all configuration sources, assigns the result to the global
// App variable, and returns a pointer to it.
//
// Usage:
//
//	cfg := config.Load()   // bind locally in main
//	config.App.Port        // global access in any other package
func Load() Config {
	// 1. Load .env (optional in prod)
	_ = godotenv.Load()

	// 2. Viper setup
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".") // project root

	// ENV overrides YAML
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 3. Read config.yaml (logs a warning if absent, does not fatal)
	if err := viper.ReadInConfig(); err != nil {
		log.Println("config.yaml not found, relying on environment variables")
	}

	// 4. Load ENV-backed values (secrets — never stored in YAML)
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

	// 5. Unmarshal only the YAML-backed sections into a scoped struct, then
	//    assign onto App. Using a scoped struct prevents viper.Unmarshal from
	//    clobbering the ENV-loaded fields above with zero values when YAML
	//    keys overlap.
	var yamlSections struct {
		Email   EmailConfig   `mapstructure:"email"`
		Job     JobConfig     `mapstructure:"job"`
		Worker  WorkerConfig  `mapstructure:"worker"`
		Machine MachineConfig `mapstructure:"machine"`
	}
	if err := viper.Unmarshal(&yamlSections); err != nil {
		log.Fatalf("failed to unmarshal config.yaml sections: %v", err)
	}
	App.Email = yamlSections.Email
	App.Job = yamlSections.Job
	App.Worker = yamlSections.Worker
	App.Machine = yamlSections.Machine

	// 6. Derive capacity values from hardware + job timing.
	//    Must run before validateConfig so derived fields pass gt=0 checks.
	computeJobCapacity(&App)

	// 7. Validate the fully-assembled config (inputs + derived fields)
	validateConfig(App)

	log.Println("✅ configuration loaded successfully")

	return App
}

/*
|--------------------------------------------------------------------------
| Capacity derivation
|--------------------------------------------------------------------------
*/

func computeJobCapacity(cfg *Config) {
	ts := cfg.Job.ServiceTimeSec
	cores := float64(cfg.Machine.CPUCores)

	hwCeilingPerSec := cores / ts
	cfg.Job.HWCeilingPerSec = hwCeilingPerSec

	var effectiveRatePerSec float64
	if cfg.Machine.JobBound == "cpu" {
		effectiveRatePerSec = math.Min(hwCeilingPerSec, hwCeilingPerSec)
	} else {
		effectiveRatePerSec = hwCeilingPerSec
	}

	sustainablePerSec := effectiveRatePerSec * cfg.Job.SafetyFactor
	cfg.Job.RateLimitPerMinute = int(sustainablePerSec * 60)
	cfg.Job.QueueMaxSize = int(hwCeilingPerSec * float64(cfg.Job.BacklogWindowSeconds))

	var recommendedConcurrency int
	if cfg.Machine.JobBound == "cpu" {
		recommendedConcurrency = max(cfg.Machine.CPUCores/cfg.Worker.Replicas, 1)
	} else {
		recommendedConcurrency = cfg.Machine.CPUCores * 3
	}

	if cfg.Worker.ConcurrencyLimit != recommendedConcurrency {
		var reason string
		if cfg.Worker.ConcurrencyLimit > recommendedConcurrency {
			if cfg.Machine.JobBound == "cpu" {
				reason = "excess goroutines add context-switch overhead with zero throughput gain"
			} else {
				reason = "excess goroutines risk downstream saturation (DB connections, Redis bandwidth)"
			}
		} else {
			if cfg.Machine.JobBound == "cpu" {
				reason = "under-provisioned relative to core count — throughput will be left on the table"
			} else {
				reason = "IO-bound jobs block on network/DB; too few goroutines starves the CPU of work"
			}
		}

		log.Printf(
			"⚠️  worker.concurrency_limit=%d does not match recommended=%d for %s-bound jobs on %d cores\n"+
				"    reason: %s\n"+
				"    auto-correcting to recommended value",
			cfg.Worker.ConcurrencyLimit,
			recommendedConcurrency,
			cfg.Machine.JobBound,
			cfg.Machine.CPUCores,
			reason,
		)
		cfg.Worker.ConcurrencyLimit = recommendedConcurrency
	}

	log.Printf(
		"📊 job capacity — T_s=%.4fs  μ_max=%.4f/s  sustainable=%.4f/s  rate_limit=%d/min  queue_max=%d jobs",
		ts,
		hwCeilingPerSec,
		sustainablePerSec,
		cfg.Job.RateLimitPerMinute,
		cfg.Job.QueueMaxSize,
	)
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
	value := viper.GetString(key)
	if value == "" {
		panic(fmt.Sprintf("required environment variable %q is not set", key))
	}
	return value
}
