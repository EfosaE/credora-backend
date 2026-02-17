package app

import (
	"context"
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
	custmiddleware "github.com/EfosaE/credora-backend/internal/middleware"
	"github.com/EfosaE/credora-backend/internal/queues"
	"github.com/EfosaE/credora-backend/service"
	accountsvc "github.com/EfosaE/credora-backend/service/account"
	authsvc "github.com/EfosaE/credora-backend/service/auth"
	idempotencysvc "github.com/EfosaE/credora-backend/service/idempotency"
	operationsvc "github.com/EfosaE/credora-backend/service/operation"
	transactionsvc "github.com/EfosaE/credora-backend/service/transaction"
	usersvc "github.com/EfosaE/credora-backend/service/user"
	"github.com/hibiken/asynq"

	"github.com/redis/go-redis/v9"
)

type AppDependencies struct {
	// Infrastructure
	Logger      *logger.Logger
	DB          *db.DB
	Redis       *redis.Client
	EventBus    *infrastructure.StreamEventBus
	QueueClient *queues.QueueClient

	//Transaction Manager
	TxManager infrastructure.TxManager

	// Repositories
	AcctRepo          *infrastructure.SqlcAccountRepository
	TrxRepo           *infrastructure.SqlcTransactionRepository
	UserRepo          *infrastructure.SqlcUserRepository
	PasswordResetRepo *infrastructure.SqlcPasswordResetRepository
	IdempotencyRepo   *infrastructure.SqlcIdempotencyRepository
	IdempotencyCache  *infrastructure.IdempotencyCache

	// Services
	MonnifySvc     *service.MonnifyService
	EmailSvc       *service.EmailServiceImpl
	AcctSvc        *accountsvc.AccountService
	TrxSvc         *transactionsvc.TransactionService
	UserSvc        *usersvc.UserService
	TokenSvc       *authsvc.JWTTokenService
	AuthSvc        *authsvc.AuthService
	OperationSvc   *operationsvc.OperationService
	SimulatorSvc   *service.SimulatorService
	IdempotencySvc *idempotencysvc.IdempotencyService

	// Middleware
	BackPressureMiddleware *custmiddleware.BackpressureMiddleware

	// Handlers
	AuthHandler        *handler.AuthHandler
	UserHandler        *handler.UserHandler
	WebhookHandler     *handler.WebHookHandler
	OperationsHandler  *handler.OperationHandler
	MonnifyHandler     *handler.MonnifyHandler
	SimHandler         *handler.SimulatorHandler
	HealthHandler      *handler.HealthHandler
	IdempotencyHandler *handler.IdempotencyHandler

	// Contexts
	Ctx context.Context
}

type AppBuilder struct {
	cfg  config.Config
	deps *AppDependencies
	err  error
}

// Create builder
func NewAppBuilder(cfg config.Config) *AppBuilder {
	return &AppBuilder{
		cfg:  cfg,
		deps: &AppDependencies{Ctx: context.Background()},
	}
}

//
// ─── INFRASTRUCTURE BUILDERS ─────────────────────────────────────────
//

func (b *AppBuilder) WithLogger() *AppBuilder {
	if b.err != nil {
		return b
	}

	loggerCfg := logger.LoggerConfig{
		LogFilePath:   "logs/app.log",
		LogLevel:      logger.INFO,
		EnableConsole: true,
		EnableFile:    true,
		MaxFileSize:   1024 * 1024,
		MaxFiles:      3,
		IncludeSource: true,
	}

	b.deps.Logger, b.err = logger.NewLogger(loggerCfg)
	return b
}

func (b *AppBuilder) WithDB() *AppBuilder {
	if b.err != nil {
		return b
	}

	b.deps.DB, b.err = db.InitDB(b.deps.Ctx)
	return b
}

func (b *AppBuilder) WithRedis() *AppBuilder {
	if b.err != nil {
		return b
	}

	b.deps.Redis = redis.NewClient(&redis.Options{
		Addr: b.cfg.RedisAddr,
		OnConnect: func(ctx context.Context, cn *redis.Conn) error {
			log.Println("Redis connected")
			return nil
		},
	})

	b.err = b.deps.Redis.Ping(b.deps.Ctx).Err()
	return b
}

func (b *AppBuilder) WithEventBus() *AppBuilder {
	if b.err != nil {
		return b
	}

	b.deps.EventBus = infrastructure.NewStreamEventBus(b.deps.Redis)
	return b
}

func (b *AppBuilder) WithQueueClient() *AppBuilder {
	if b.err != nil {
		return b
	}
	b.deps.QueueClient = queues.NewQueueClient(b.cfg.RedisAddr)
	return b
}

//
// ─── REPOSITORIES ─────────────────────────────────────────
//

func (b *AppBuilder) WithRepositories() *AppBuilder {
	if b.err != nil {
		return b
	}

	if b.deps.DB == nil {
		log.Println("[ERROR] DB dependency is missing: DB must be initialized before repositories")
		b.err = fmt.Errorf("DB must be initialized before repositories")
		return b
	}
	if b.deps.Redis == nil {
		log.Println("[ERROR] Redis dependency is missing: Redis must be initialized before repositories")
		b.err = fmt.Errorf("Redis must be initialized before repositories")
		return b
	}

	b.deps.TxManager = infrastructure.NewTxManager(b.deps.DB.Pool)

	//POOL VS QUERIES: We use the DB pool to create repositories that need direct access to the connection pool (like those that need transactions), and we use the DB queries for repositories that only need to execute queries without managing transactions. This allows us to optimize resource usage while still providing the necessary functionality for each repository.
	b.deps.AcctRepo = infrastructure.NewSqlcAccountRepository(b.deps.DB.Pool)
	b.deps.TrxRepo = infrastructure.NewSqlcTransactionRepository(b.deps.DB.Pool)
	b.deps.UserRepo = infrastructure.NewSqlcUserRepository(b.deps.DB.Pool)
	b.deps.IdempotencyRepo = infrastructure.NewSqlcIdempotencyRepository(b.deps.DB.Pool)
	b.deps.PasswordResetRepo = infrastructure.NewSqlcPasswordResetRepository(b.deps.DB.Pool)
	b.deps.IdempotencyCache = infrastructure.NewIdempotencyCache(b.deps.Redis, 5*time.Minute)
	return b
}

//
// ─── SERVICES ─────────────────────────────────────────
//

func (b *AppBuilder) WithMonnifyService() *AppBuilder {
	if b.err != nil {
		return b
	}

	if b.deps.Logger == nil {
		log.Println("[ERROR] Logger dependency is missing: Logger must be initialized before Monnify service")
		b.err = fmt.Errorf("logger must be initialized before Monnify service")
		return b
	}
	mcfg := &monnify.MonnifyConfig{
		ApiKey:       b.cfg.MonnifyApiKey,
		SecretKey:    b.cfg.MonnifySecretKey,
		ContractCode: b.cfg.MonnifyContractCode,
		BaseURL:      b.cfg.MonnifyBaseURL,
	}

	client := infrastructure.NewMonnifyClient(mcfg, &http.Client{Timeout: 10 * time.Second})
	b.deps.MonnifySvc = service.NewMonnifyService(client, b.deps.Logger)
	return b
}

func (b *AppBuilder) WithEmailService() *AppBuilder {
	if b.err != nil {
		return b
	}

	if b.deps.EventBus == nil {
		log.Println("[ERROR] EventBus dependency is missing: EventBus must be initialized before email service")
		b.err = fmt.Errorf("event bus must be initialized before email service")
		return b
	}
	adapter := infrastructure.NewEmailAdapter()
	b.deps.EmailSvc = service.NewEmailService(adapter, b.deps.EventBus)
	return b
}

func (b *AppBuilder) WithAccountService() *AppBuilder {
	if b.err != nil {
		return b
	}
	if b.deps.EventBus == nil {
		log.Println("[ERROR] EventBus dependency is missing: EventBus must be initialized before account service")
		b.err = fmt.Errorf("event bus must be initialized before account service")
		return b
	}
	if b.deps.AcctRepo == nil {
		log.Println("[ERROR] AccountRepo dependency is missing: Account repository must be initialized before account service")
		b.err = fmt.Errorf("account repository must be initialized before account service")
		return b
	}
	if b.deps.Logger == nil {
		log.Println("[ERROR] Logger dependency is missing: Logger must be initialized before account service")
		b.err = fmt.Errorf("logger must be initialized before account service")
		return b
	}
	b.deps.AcctSvc = accountsvc.NewAccountService(b.deps.AcctRepo, b.deps.Logger, b.deps.EventBus)
	return b
}

func (b *AppBuilder) WithIdempotencyService() *AppBuilder {
	if b.err != nil {
		return b
	}
	if b.deps.IdempotencyRepo == nil {
		log.Println("[ERROR] IdempotencyRepo dependency is missing: Idempotency repo must be initialized before Idempotency Service")
		b.err = fmt.Errorf("the idempotency repo must be initialized before Idemp Service")
		return b
	}
	b.deps.IdempotencySvc = idempotencysvc.NewIdempotencyService(b.deps.IdempotencyRepo)
	return b
}

func (b *AppBuilder) WithTransactionService() *AppBuilder {
	if b.err != nil {
		return b
	}

	if b.deps.TrxRepo == nil {
		log.Println("[ERROR] TrxRepo dependency is missing: Transaction repository must be initialized before transaction service")
		b.err = fmt.Errorf("transaction repository must be initialized before transaction service")
		return b
	}
	if b.deps.Logger == nil {
		log.Println("[ERROR] Logger dependency is missing: Logger must be initialized before transaction service")
		b.err = fmt.Errorf("logger must be initialized before transaction service")
		return b
	}
	b.deps.TrxSvc = transactionsvc.NewTransactionService(b.deps.TrxRepo, b.deps.Logger)
	return b
}
func (b *AppBuilder) WithSimulatorService() *AppBuilder {
	if b.err != nil {
		return b
	}
	b.deps.SimulatorSvc = service.NewSimulatorService(
		infrastructure.NewInMemoryRepo(b.cfg.WebhookURL),
	)
	return b
}
func (b *AppBuilder) WithUserService() *AppBuilder {
	if b.err != nil {
		return b
	}

	if b.deps.UserRepo == nil {
		log.Println("[ERROR] UserRepo dependency is missing: User repository must be initialized before user service")
		b.err = fmt.Errorf("user repository must be initialized before user service")
		return b
	}
	if b.deps.Logger == nil {
		log.Println("[ERROR] Logger dependency is missing: Logger must be initialized before user service")
		b.err = fmt.Errorf("logger must be initialized before user service")
		return b
	}
	if b.deps.EventBus == nil {
		log.Println("[ERROR] EventBus dependency is missing: EventBus must be initialized before user service")
		b.err = fmt.Errorf("event bus must be initialized before user service")
		return b
	}
	if b.deps.MonnifySvc == nil {
		log.Println("[ERROR] MonnifySvc dependency is missing: Monnify service must be initialized before user service")
		b.err = fmt.Errorf("Monnify service must be initialized before user service")
		return b
	}
	if b.deps.QueueClient == nil {
		log.Println("[ERROR] QueueClient dependency is missing: Queue client must be initialized before user service")
		b.err = fmt.Errorf("queue client must be initialized before user service")
		return b
	}
	b.deps.UserSvc = usersvc.NewUserService(
		b.deps.UserRepo,
		b.deps.Logger,
		b.deps.EventBus,
		b.deps.MonnifySvc,
		b.deps.QueueClient,
	)
	return b
}

func (b *AppBuilder) WithAuthService() *AppBuilder {
	if b.err != nil {
		return b
	}

	if b.deps.TxManager == nil {
		log.Println("[ERROR] TxManager dependency is missing")
		b.err = fmt.Errorf("transaction manager must be initialized before auth service")
		return b
	}
	if b.deps.AcctRepo == nil {
		log.Println("[ERROR] AcctRepo dependency is missing: Account repository must be initialized before auth service")
		b.err = fmt.Errorf("account repository must be initialized before auth service")
		return b
	}
	if b.deps.PasswordResetRepo == nil {
		log.Println("[ERROR] PasswordResetRepo dependency is missing: Password reset repository must be initialized before auth service")
		b.err = fmt.Errorf("password reset repository must be initialized before auth service")
		return b
	}
	if b.deps.EmailSvc == nil {
		log.Println("[ERROR] EmailSvc dependency is missing")
		b.err = fmt.Errorf("email service must be initialized before auth service")
		return b
	}
	b.deps.TokenSvc = authsvc.NewJWTTokenService(b.cfg.JwtSecret, 24*time.Hour)
	b.deps.AuthSvc = authsvc.NewAuthService(b.deps.TxManager, b.deps.UserRepo, b.deps.PasswordResetRepo, b.deps.TokenSvc, b.deps.AcctRepo, b.deps.EmailSvc)
	return b
}

func (b *AppBuilder) WithOperationService() *AppBuilder {
	if b.err != nil {
		return b
	}

	if b.deps.TxManager == nil {
		log.Println("[ERROR] TxManager dependency is missing")
		b.err = fmt.Errorf("transaction manager must be initialized before operation service")
		return b
	}
	if b.deps.AcctRepo == nil {
		log.Println("[ERROR] AcctRepo dependency is missing: Account repo must be initialized before operation service")
		b.err = fmt.Errorf("account repo must be initialized before operation service")
		return b
	}
	if b.deps.TrxRepo == nil {
		log.Println("[ERROR] TrxRepo dependency is missing: Transaction repo must be initialized before operation service")
		b.err = fmt.Errorf("transaction repo must be initialized before operation service")
		return b
	}
	if b.deps.IdempotencyRepo == nil {
		log.Println("[ERROR] IdempotencyRepo dependency is missing: Idempotency repo must be initialized before operation service")
		b.err = fmt.Errorf("idempotency repo must be initialized before operation service")
		return b
	}
	if b.deps.Logger == nil {
		log.Println("[ERROR] Logger dependency is missing: Logger must be initialized before operation service")
		b.err = fmt.Errorf("logger must be initialized before operation service")
		return b
	}

	b.deps.OperationSvc = operationsvc.NewOperationService(
		b.deps.TxManager,
		b.deps.AcctRepo,
		b.deps.TrxRepo,
		b.deps.IdempotencyRepo,
		b.deps.Logger,
	)
	return b
}

//
// ─── EVENT SUBSCRIPTIONS ─────────────────────────────────────────
//

func (b *AppBuilder) WithEventSubscriptions() *AppBuilder {
	if b.err != nil {
		return b
	}

	if err := b.deps.EmailSvc.SubscribeToUserCreatedEvents(b.deps.Ctx); err != nil {
		b.err = err
	}
	if err := b.deps.AcctSvc.SubscribeToUserCreatedEvents(b.deps.Ctx); err != nil {
		b.err = err
	}

	return b
}

func (b *AppBuilder) WithMiddlewares() *AppBuilder {
	if b.err != nil {
		return b
	}
	inspector := asynq.NewInspector(asynq.RedisClientOpt{
		Addr: b.cfg.RedisAddr,
	})
	b.deps.BackPressureMiddleware = custmiddleware.NewBackpressure(inspector, 250, "critical")

	return b
}

// WithHandlers initializes all HTTP handlers
func (b *AppBuilder) WithHandlers() *AppBuilder {
	if b.err != nil {
		return b
	}

	b.deps.AuthHandler = handler.NewAuthHandler(b.deps.UserSvc, b.deps.AuthSvc)
	b.deps.UserHandler = handler.NewUserHandler(b.deps.UserSvc)
	b.deps.WebhookHandler = handler.NewWebHookHandler(b.deps.AcctSvc, b.deps.TrxSvc, b.deps.MonnifySvc, b.deps.IdempotencySvc, b.deps.QueueClient)
	b.deps.OperationsHandler = handler.NewOperationHandler(b.deps.OperationSvc, b.deps.IdempotencySvc, b.deps.QueueClient)
	b.deps.MonnifyHandler = handler.NewMonnifyHandler(b.deps.MonnifySvc)
	b.deps.SimHandler = handler.NewSimulatorHandler(b.deps.SimulatorSvc)
	b.deps.HealthHandler = handler.NewHealthHandler(b.deps.Redis)
	b.deps.IdempotencyHandler = handler.NewIdempotencyHandler(b.deps.IdempotencySvc)

	return b
}

//
// ─── BUILD METHODS ─────────────────────────────────────────
//

func (b *AppBuilder) Build() (*AppDependencies, error) {
	return b.deps, b.err
}

func (b *AppBuilder) BuildForServer() (*AppDependencies, error) {
	return b.
		WithLogger().
		WithDB().
		WithRedis().
		WithEventBus().
		WithQueueClient().
		WithRepositories().
		WithMonnifyService().
		WithEmailService().
		WithAccountService().
		WithTransactionService().
		WithUserService().
		WithAuthService().
		WithOperationService().
		WithIdempotencyService().
		WithEventSubscriptions().
		WithMiddlewares().
		WithHandlers().
		Build()
}

func (b *AppBuilder) BuildForWorker() (*AppDependencies, error) {
	return b.
		WithLogger().
		WithDB().
		WithRedis().
		WithEventBus().
		WithRepositories().
		WithOperationService().
		WithAccountService().
		WithTransactionService().
		Build()
}

func (b *AppBuilder) BuildForTests() (*AppDependencies, error) {
	return b.
		WithLogger().
		WithDB().
		WithRepositories().
		Build()
}
