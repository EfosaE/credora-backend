package main

import (
	"context"
	"log"

	"github.com/EfosaE/credora-backend/infrastructure"
	"github.com/EfosaE/credora-backend/internal/config"
	"github.com/EfosaE/credora-backend/internal/queues"
	"github.com/EfosaE/credora-backend/service"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
)

func main() {
	// Load app configuration
	config.Load()

	// -----------------------------
	// 1. go-redis client for EventBus
	// -----------------------------
	appRedis := redis.NewClient(&redis.Options{
		Addr: config.App.RedisAddr,
		OnConnect: func(ctx context.Context, cn *redis.Conn) error {
			log.Println("✅ Redis connection established for EventBus")
			return nil
		},
	})

	eventBus := infrastructure.NewStreamEventBus(appRedis)

	// Initialize Email service
	emailAdapter := infrastructure.NewEmailAdapter()
	emailSvc := service.NewEmailService(emailAdapter, eventBus)

	// -----------------------------
	// 2. Asynq Redis client for queue
	// -----------------------------
	asynqRedis := asynq.RedisClientOpt{
		Addr: config.App.RedisAddr,
	}

	// Create Asynq server (worker)
	srv := asynq.NewServer(
		asynqRedis,
		asynq.Config{
			Concurrency: 10, // number of parallel jobs
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
		},
	)

	// -----------------------------
	// 3. Handlers
	// -----------------------------
	handlers := queues.NewHandlers(
		emailSvc,
		// wallet.NewService(),
		// transfer.NewService(),
		// webhook.NewService(),
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(queues.TypeWelcomeEmail, handlers.HandleSendEmail)
	mux.HandleFunc(queues.TypeAccountNumberEmail, handlers.HandleSendAccountNumberEmail)
	// Uncomment and add other handlers when ready
	// mux.HandleFunc(queues.TaskProcessExternalTransfer, handlers.HandleExternalTransfer)
	// mux.HandleFunc(queues.TaskProcessInternalTransfer, handlers.HandleInternalTransfer)
	// mux.HandleFunc(queues.TaskProcessVAWebhook, handlers.HandleVACredit)

	log.Println("🚀 Asynq worker running...")
	if err := srv.Run(mux); err != nil {
		log.Fatal(err)
	}
}
