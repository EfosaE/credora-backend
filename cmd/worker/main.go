package main

import (
	app "github.com/EfosaE/credora-backend/di"
	"github.com/EfosaE/credora-backend/domain/logger"
	"github.com/EfosaE/credora-backend/internal/config"
	"github.com/EfosaE/credora-backend/internal/queues"

	"github.com/hibiken/asynq"
)

func main() {
	config.Load()

	builder := app.NewAppBuilder(config.App)

	deps, err := builder.BuildForWorker()
	if err != nil {
		// fallback because logger not yet available
		panic(err)
	}

	l := logger.Get()

	workerLogger := l.With().
		Str("component", "worker").
		Logger()

	if deps.EmailSvc == nil {
		workerLogger.Fatal().Msg("EmailSvc is nil — check BuildForWorker()")
	}
	if deps.OperationSvc == nil {
		workerLogger.Fatal().Msg("OperationSvc is nil — check BuildForWorker()")
	}
	if deps.AcctSvc == nil {
		workerLogger.Fatal().Msg("AcctSvc is nil — check BuildForWorker()")
	}
	if deps.TrxSvc == nil {
		workerLogger.Fatal().Msg("TrxSvc is nil — check BuildForWorker()")
	}

	asynqRedis := asynq.RedisClientOpt{
		Addr: config.App.RedisAddr,
	}

	srv := asynq.NewServer(
		asynqRedis,
		asynq.Config{
			Concurrency: 12,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			StrictPriority: true,
		},
	)

	handlers := queues.NewHandlers(
		deps.EmailSvc,
		*deps.OperationSvc,
		deps.AcctSvc,
		deps.TrxSvc,
		workerLogger,
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(queues.TypeWelcomeEmail, handlers.HandleSendWelcomeEmail)
	mux.HandleFunc(queues.TypeAccountNumberEmail, handlers.HandleSendAccountNumberEmail)
	mux.HandleFunc(queues.TypeInternalTransfer, handlers.HandleInternalTransfer)
	mux.HandleFunc(queues.TypeWebhookInboundTransfer, handlers.HandleInboundTransferWebhook)

	workerLogger.Info().
		Int("concurrency", 12).
		Str("redis_addr", config.App.RedisAddr).
		Msg("Asynq worker starting")

	if err := srv.Run(mux); err != nil {
		workerLogger.Fatal().
			Err(err).
			Msg("Asynq worker crashed")
	}
}
