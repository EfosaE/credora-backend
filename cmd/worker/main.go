package main

import (
	"log"

	app "github.com/EfosaE/credora-backend/di"

	"github.com/EfosaE/credora-backend/internal/config"
	"github.com/EfosaE/credora-backend/internal/queues"

	"github.com/hibiken/asynq"
)

func main() {
	// Load app configuration
	config.Load()

	builder := app.NewAppBuilder(config.App)
	deps, err := builder.BuildForWorker()
	if err != nil {
		log.Fatal(err)
	}

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
			Concurrency: 20, // number of parallel jobs
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},

			StrictPriority: true,
		},
	)

	// -----------------------------
	// 3. Handlers
	// -----------------------------
	handlers := queues.NewHandlers(
		deps.EmailSvc,
		*deps.OperationSvc,
		deps.AcctSvc,
		deps.TrxSvc,
		deps.Logger,
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(queues.TypeWelcomeEmail, handlers.HandleSendEmail)
	mux.HandleFunc(queues.TypeAccountNumberEmail, handlers.HandleSendAccountNumberEmail)
	mux.HandleFunc(queues.TypeInternalTransfer, handlers.HandleInternalTransfer)
	mux.HandleFunc(queues.TypeWebhookInboundTransfer, handlers.HandleInboundTransferWebhook)
	// Uncomment and add other handlers when ready

	// mux.HandleFunc(queues.TaskProcessInternalTransfer, handlers.HandleInternalTransfer)
	// mux.HandleFunc(queues.TaskProcessVAWebhook, handlers.HandleVACredit)

	log.Println("Asynq worker running...")
	if err := srv.Run(mux); err != nil {
		log.Fatal(err)
	}
}
