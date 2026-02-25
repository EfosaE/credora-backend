package main

import (
	"math/rand"
	"strings"
	"time"

	app "github.com/EfosaE/credora-backend/di"
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

	workerLogger := deps.Logger

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
			Concurrency: config.App.Worker.ConcurrencyLimit,
			Queues: map[string]int{
				"critical": 6,
				"default":  3,
				"low":      1,
			},
			RetryDelayFunc: func(n int, err error, t *asynq.Task) time.Duration {
				if strings.Contains(err.Error(), "could not obtain lock") {
					// Give the competing transaction time to commit
					// n=1 → 400-700ms, n=2 → 800-1200ms, n=3 → default
					base := time.Duration(300+n*200) * time.Millisecond
					jitter := time.Duration(rand.Intn(300)) * time.Millisecond
					return base + jitter
				}
				return asynq.DefaultRetryDelayFunc(n, err, t)
			},
			StrictPriority: config.App.Worker.StrictPriority,
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
		Int("concurrency", config.App.Worker.ConcurrencyLimit).
		Str("redis_addr", config.App.RedisAddr).
		Bool("strict_priority", config.App.Worker.StrictPriority).
		Msg("Asynq worker started")

	if err := srv.Run(mux); err != nil {
		workerLogger.Fatal().
			Err(err).
			Msg("Asynq worker crashed")
	}
}
