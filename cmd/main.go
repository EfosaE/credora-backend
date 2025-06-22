package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/EfosaE/credora-backend/domain/logger"
	"github.com/EfosaE/credora-backend/domain/monnify"
	"github.com/EfosaE/credora-backend/infrastructure"
	"github.com/EfosaE/credora-backend/internal/config"
	"github.com/EfosaE/credora-backend/internal/db"
	"github.com/EfosaE/credora-backend/internal/handler"
	"github.com/EfosaE/credora-backend/internal/router"
	"github.com/EfosaE/credora-backend/internal/server"
	"github.com/EfosaE/credora-backend/service"
)

func main() {

	ctx := context.Background()
	config.Load()
	// Create logger configuration
	loggerConfig := logger.LoggerConfig{
		LogFilePath:   "logs/app.log",
		LogLevel:      logger.INFO,
		EnableConsole: true,
		EnableFile:    true,
		MaxFileSize:   1024 * 1024, // 1MB for demo
		MaxFiles:      3,
		IncludeSource: true,
	}
	// Create logger
	logger, err := logger.NewLogger(loggerConfig)
	if err != nil {
		log.Fatal("Failed to create logger:", err)
	}
	defer logger.Close()

	db, err := db.InitDB(ctx)

	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Pool.Close()

	// Initialize Monnify configuration
	monnifyConfig := &monnify.MonnifyConfig{
		ApiKey:       config.App.MonnifyApiKey,
		SecretKey:    config.App.MonnifySecretKey,
		ContractCode: config.App.MonnifyContractCode,
		BaseURL:      config.App.MonnifyBaseURL,
	}
	
	// Initialize Monnify client
	client := infrastructure.NewMonnifyClient(monnifyConfig, &http.Client{Timeout: 10 * time.Second})
	monnifySvc := service.NewMonnifyService(client, logger)

	userRepo := infrastructure.NewSqlcUserRepository(ctx, db.Queries)
	userService := service.NewUserService(userRepo, logger, monnifySvc)
	userHandler := handler.NewUserHandler(userService)

	r := router.SetupRouter(userHandler)

	// Option 1: Use default configuration
	srv := server.New(r, nil)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/api/v1", http.StatusFound)
	})

	// Start server with graceful shutdown
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
}
