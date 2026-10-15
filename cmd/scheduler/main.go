package main

import (
	app "github.com/EfosaE/credora-backend/di"
	"github.com/EfosaE/credora-backend/internal/config"
	"github.com/EfosaE/credora-backend/internal/queues"
	"github.com/hibiken/asynq"
)

func main() {
	config.Load()
	builder := app.NewAppBuilder(config.App)

	deps, err := builder.BuildForWorker()
	schedulerLogger := deps.Logger
	if err != nil {
		// fallback because logger not yet available
		panic(err)
	}

	scheduler := asynq.NewScheduler(
		asynq.RedisClientOpt{
			Addr: config.App.RedisAddr,
		},
		&asynq.SchedulerOpts{},
	)

	defer scheduler.Shutdown()

	// register cron jobs here
	task := asynq.NewTask(queues.TypeReconciliationReport, nil)
	// You can use "@every <duration>" to specify the interval.
	entryID, err := scheduler.Register("@every 30s", task)
	if err != nil {
		schedulerLogger.Fatal().Err(err).Msg("Failed to register cron job")
	}
	schedulerLogger.Info().Str("entryID", entryID).Msg("Cron job registered")

	if err := scheduler.Run(); err != nil {
		schedulerLogger.Fatal().Err(err).Msg("Failed to run scheduler")
	}
}
