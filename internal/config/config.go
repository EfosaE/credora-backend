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
	// YAML-backed inputs.
	// NOTE: processing_rate_per_sec holds the p95 service time in seconds
	// (T_s) — i.e. how long one job takes end-to-end. The YAML key name is
	// kept for backwards compatibility; internally it is always treated as T_s.
	ServiceTimeSec       float64 `mapstructure:"processing_rate_per_job" validate:"gt=0"`
	SafetyFactor         float64 `mapstructure:"safety_factor"           validate:"gt=0,lte=1"`
	BacklogWindowSeconds int     `mapstructure:"backlog_window_seconds"  validate:"gt=0"`
	QueueName            string  `mapstructure:"queue_name"              validate:"required"`

	// Derived — computed by computeJobCapacity, never read from YAML/ENV.
	HWCeilingPerSec    float64 `mapstructure:"-" validate:"gt=0"` // μ_max  = cores / T_s
	RateLimitPerMinute int     `mapstructure:"-" validate:"gt=0"` // floor(sustainable/sec * 60)
	QueueMaxSize       int     `mapstructure:"-" validate:"gt=0"` // floor(μ_max * backlog_window)
}

type WorkerConfig struct {
	ConcurrencyLimit int `mapstructure:"concurrency_limit" validate:"gt=0"`
	Replicas         int `mapstructure:"replicas"          validate:"gt=0"`
	// NOTE: do not use validate:"required" on bool — validator treats false
	// as "not set" and will fail. Use *bool if you need to distinguish
	// "unset" from false.
	StrictPriority bool `mapstructure:"strict_priority"`
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
| Global app config
|--------------------------------------------------------------------------
*/

var App Config

/*
|--------------------------------------------------------------------------
| Load
|--------------------------------------------------------------------------
*/

func Load() {
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
}

/*
|--------------------------------------------------------------------------
| Capacity derivation
|--------------------------------------------------------------------------
*/

func computeJobCapacity(cfg *Config) {
	// T_s: p95 service time per job in seconds.
	ts := cfg.Job.ServiceTimeSec // e.g. 3.3975 s/job

	cores := float64(cfg.Machine.CPUCores) // e.g. 4

	// μ_max = C / T_s — the physical throughput ceiling of this machine.
	// No amount of goroutines or replicas can exceed this for CPU-bound work.
	// e.g. 4 / 3.3975 = 1.1775 jobs/sec
	hwCeilingPerSec := cores / ts
	cfg.Job.HWCeilingPerSec = hwCeilingPerSec

	// Effective sustainable rate depends on job bound type:
	//
	// cpu-bound: goroutines beyond (cores/replicas) add context-switch
	//            overhead with zero throughput gain. Hard-cap at μ_max.
	//
	// io-bound:  goroutines block on network/DB/Redis while the CPU is free,
	//            so concurrency above cores IS beneficial. The real ceiling is
	//            downstream saturation (DB connections, Redis bandwidth), not
	//            CPU. We use μ_max as a conservative anchor and do not clamp
	//            below it.
	//
	// In both cases the anchor is hwCeilingPerSec — for cpu the cap is hard,
	// for io it is a floor-safe reference point.
	var effectiveRatePerSec float64
	if cfg.Machine.JobBound == "cpu" {
		effectiveRatePerSec = math.Min(hwCeilingPerSec, hwCeilingPerSec) // hard cap at ceiling
	} else {
		effectiveRatePerSec = hwCeilingPerSec // io-bound: ceiling is the anchor, not a hard cap
	}

	// Apply safety factor to leave headroom for GC pauses, retry spikes,
	// and p99 tail latency exceeding p95.
	// e.g. 1.1775 * 0.85 = 1.0009 jobs/sec
	sustainablePerSec := effectiveRatePerSec * cfg.Job.SafetyFactor

	// RateLimitPerMinute: what we advertise to the HTTP rate limiter.
	// e.g. floor(1.0009 * 60) = 60 req/min
	cfg.Job.RateLimitPerMinute = int(sustainablePerSec * 60)

	// QueueMaxSize: maximum backlog depth before new jobs are rejected.
	// = μ_max * W_q_max  (how many jobs fit inside the acceptable wait budget)
	// e.g. floor(1.1775 * 20) = 23 jobs
	cfg.Job.QueueMaxSize = int(hwCeilingPerSec * float64(cfg.Job.BacklogWindowSeconds))

	// Concurrency sanity check — auto-correct and warn if over- or under-provisioned.
	var recommendedConcurrency int
	if cfg.Machine.JobBound == "cpu" {
		// Ideal: one goroutine per core per replica. Beyond this you pay
		// context-switch cost with no throughput gain.
		recommendedConcurrency = max(cfg.Machine.CPUCores/cfg.Worker.Replicas, 1)
	} else {
		// IO-bound rule of thumb: concurrency ≈ cores * (1 + wait/service). Go to the lesson notes.txt to undertsnd what this is line 18
		// Without measuring wait time, cores*3 is a practical starting point.
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
	// All constraints (including derived fields) are expressed via struct tags.
	// No manual checks needed.
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

// mustEnv panics instead of log.Fatalf so the compiler correctly treats
// every call site as a termination point — no dead return needed.
func mustEnv(key string) string {
	value := viper.GetString(key)
	if value == "" {
		panic(fmt.Sprintf("required environment variable %q is not set", key))
	}
	return value
}
