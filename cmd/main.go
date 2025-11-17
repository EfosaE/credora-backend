package main

import (
	"context"
	"flag"
	"fmt"
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
	"github.com/EfosaE/credora-backend/internal/seeder"
	"github.com/EfosaE/credora-backend/internal/server"
	"github.com/EfosaE/credora-backend/service"
	accountsvc "github.com/EfosaE/credora-backend/service/account"
	authsvc "github.com/EfosaE/credora-backend/service/auth"
	operationsvc "github.com/EfosaE/credora-backend/service/operation"
	transactionsvc "github.com/EfosaE/credora-backend/service/transaction"
	usersvc "github.com/EfosaE/credora-backend/service/user"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
)

func main() {
	config.Load()

	// Define CLI flag
	seedFlag := flag.Bool("seed", false, "Seed the database with fake data")
	flag.Parse()

	// Check if --seed flag was passed
	if *seedFlag {
		ctx := context.Background()

		conn, err := pgx.Connect(ctx, config.App.TestDbUrl)
		if err != nil {
			log.Fatalf("DB connect failed: %v", err)
		}
		defer conn.Close(ctx)
		seeder := seeder.NewSeeder(conn, ctx)
		if err := seeder.SeedUsersAndAccounts(5); err != nil {
			log.Fatalf("Seeding failed: %v", err)
		}
		log.Println("✅ Seeding complete")
		return // exit after seeding
	}

	dbCtx := context.Background()
	qCtx := context.Background()
	evtCtx := context.Background()

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

	db, err := db.InitDB(dbCtx)

	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Pool.Close()

	// redis init
	rdb := redis.NewClient(&redis.Options{
		Addr: config.App.RedisAddr,
		OnConnect: func(ctx context.Context, cn *redis.Conn) error {
			fmt.Println("✅ Redis connection established")
			return nil
		},
	})

	// Optional: ping to test right away
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		fmt.Printf("❌ Redis ping failed: %v\n", err)
	}

	eventBus := infrastructure.NewStreamEventBus(rdb)

	// Initialize Monnify configuration
	monnifyConfig := &monnify.MonnifyConfig{
		ApiKey:       config.App.MonnifyApiKey,
		SecretKey:    config.App.MonnifySecretKey,
		ContractCode: config.App.MonnifyContractCode,
		BaseURL:      config.App.MonnifyBaseURL,
	}

	// Initialize Monnify client
	monnifyClient := infrastructure.NewMonnifyClient(monnifyConfig, &http.Client{Timeout: 10 * time.Second})
	monnifySvc := service.NewMonnifyService(monnifyClient, logger)
	monnifyHandler := handler.NewMonnifyHandler(monnifySvc)

	// initialize email service
	emailAdapter := infrastructure.NewEmailAdapter()
	emailSvc := service.NewEmailService(emailAdapter, eventBus)

	//initialize acct service
	acctRepo := infrastructure.NewSqlcAccountRepository(db.Pool)
	acctSvc := accountsvc.NewAccountService(acctRepo, logger, eventBus)

	//initialize trx service
	trxRepo := infrastructure.NewSqlcTransactionRepository(db.Pool)
	trxSvc := transactionsvc.NewTransactionService(trxRepo, logger)

	//initialize user service
	userRepo := infrastructure.NewSqlcUserRepository(qCtx, db.Queries)
	userService := usersvc.NewUserService(userRepo, logger, eventBus, monnifySvc)

	// initialize auth service
	tokenSvc := authsvc.NewJWTTokenService(config.App.JwtSecret, time.Hour*24)
	authSvc := authsvc.NewAuthService(tokenSvc, acctRepo)

	// initialize my simulator service
	simRepo := infrastructure.NewInMemoryRepo(config.App.WebhookURL)
	simSvc := service.NewSimulatorService(simRepo)

	operationSvc := operationsvc.NewOperationService(acctRepo, trxRepo, logger)

	// Initialize route handlers
	authHandler := handler.NewAuthHandler(userService, authSvc)
	userHandler := handler.NewUserHandler(userService)
	wbHkHandler := handler.NewWebHookHandler(acctSvc, trxSvc)
	operationsHandler := handler.NewOperationHandler(operationSvc)
	simHandler := handler.NewSimulatorHandler(simSvc)

	// Subscribe to events
	if err := emailSvc.SubscribeToUserCreatedEvents(evtCtx); err != nil {
		panic(err)
	}

	if err := acctSvc.SubscribeToUserCreatedEvents(evtCtx); err != nil {
		panic(err)
	}

	r := router.SetupRouter(authHandler, operationsHandler, userHandler, monnifyHandler, tokenSvc, wbHkHandler, simHandler)

	// Option 1: Use default configuration
	srv := server.New(r, nil)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/api/v1", http.StatusFound)
	})

	// log.Printf("Launching HTTP server on %s...", config.App.Port)

	// Start server with graceful shutdown
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
}
