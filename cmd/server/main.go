package main

import (
	"log"
	"net/http"

	app "github.com/EfosaE/credora-backend/di"

	"github.com/EfosaE/credora-backend/internal/config"

	"github.com/EfosaE/credora-backend/internal/router"
	"github.com/EfosaE/credora-backend/internal/server"
)

func main() {
	cfg := config.Load()

	// Build dependencies using DI
	builder := app.NewAppBuilder(cfg)
	deps, err := builder.BuildForServer()
	if err != nil {
		log.Fatal("Failed to initialize dependencies:", err)
	}
	defer deps.DB.Pool.Close()
	log.Println("✅ Dependencies initialized")

	logger := deps.Logger
	// Router
	r := router.SetupRouter(router.RouterSetupParams{
		AuthHandler:            deps.AuthHandler,
		OperationsHandler:      deps.OperationsHandler,
		UserHandler:            deps.UserHandler,
		MonnifyHandler:         deps.MonnifyHandler,
		Auth:                   deps.TokenSvc,
		WbHkHandler:            deps.WebhookHandler,
		SimHandler:             deps.SimHandler,
		HealthHandler:          deps.HealthHandler,
		Cache:                  deps.IdempotencyCache,
		AcctSvc:                deps.AcctSvc,
		IdempSvc:               deps.IdempotencySvc,
		IdempHandler:           deps.IdempotencyHandler,
		BackPressureMiddleware: deps.BackPressureMiddleware,
		NotificationHandler:    deps.NotificationHandler,
	})

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/api/v1", http.StatusFound)
	})

	srv := server.New(r, nil)
	if err := srv.Start(); err != nil {
		logger.Fatal().Err(err).Msg("Failed to start server")
	}
}
