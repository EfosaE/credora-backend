package test

import "github.com/EfosaE/credora-backend/internal/config"

func NewTestConfig() config.Config {
	return config.Config{
		DbUrl:                        "postgres://postgres:postgres@localhost:5432/credora_test",
		TestDbUrl:                    "postgres://postgres:postgres@localhost:5432/credora_test",
		Port:                         "8080",
		Env:                          "test",
		JwtSecret:                    "test-secret",
		GoogleApplicationCredentials: "test-credentials",

		MonnifyApiKey:       "test-monnify-key",
		MonnifySecretKey:    "test-monnify-secret",
		MonnifyContractCode: "test-contract-code",
		MonnifyBaseURL:      "https://sandbox.monnify.com",

		ResendApiKey: "test-resend-key",

		MailtrapHost: "sandbox.smtp.mailtrap.io",
		MailtrapPort: "2525",
		MailtrapUser: "test-user",
		MailtrapPass: "test-pass",

		RedisAddr:   "localhost:6379",
		WebhookURL:  "http://localhost:8080/webhook",
		FrontendURL: "http://localhost:3000",
		LogLevel:    "1",

		Email: config.EmailConfig{
			Provider: "mailtrap",
			Rate:     10,
			Burst:    20,
		},
		Job: config.JobConfig{
			ServiceTimeSec:       1.0,
			SafetyFactor:         0.85,
			BacklogWindowSeconds: 20,
			QueueName:            "test-queue",
			HWCeilingPerSec:      4.0,
			RateLimitPerMinute:   200,
			QueueMaxSize:         80,
		},
		Worker: config.WorkerConfig{
			ConcurrencyLimit: 4,
			Replicas:         1,
			StrictPriority:   false,
		},
		Machine: config.MachineConfig{
			CPUCores: 4,
			JobBound: "cpu",
		},
	}
}
