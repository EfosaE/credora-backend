package router

import (
	// "fmt"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/EfosaE/credora-backend/domain/logger"
	"github.com/EfosaE/credora-backend/infrastructure"
	"github.com/EfosaE/credora-backend/internal/config"
	"github.com/EfosaE/credora-backend/internal/handler"
	"github.com/EfosaE/credora-backend/internal/metrics"
	custmiddleware "github.com/EfosaE/credora-backend/internal/middleware"
	"github.com/EfosaE/credora-backend/internal/openapi"
	"github.com/EfosaE/credora-backend/internal/response"
	accountsvc "github.com/EfosaE/credora-backend/service/account"
	authsvc "github.com/EfosaE/credora-backend/service/auth"
	idempotencysvc "github.com/EfosaE/credora-backend/service/idempotency"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/go-chi/jwtauth/v5"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	// "github.com/go-chi/jwtauth/v5"
)

type RouterSetupParams struct {
	AuthHandler            *handler.AuthHandler
	OperationsHandler      *handler.OperationHandler
	UserHandler            *handler.UserHandler
	MonnifyHandler         *handler.MonnifyHandler
	Auth                   *authsvc.JWTTokenService
	WbHkHandler            *handler.WebHookHandler
	SimHandler             *handler.SimulatorHandler
	HealthHandler          *handler.HealthHandler
	IdempHandler           *handler.IdempotencyHandler
	Cache                  *infrastructure.IdempotencyCache
	AcctSvc                *accountsvc.AccountService
	IdempSvc               *idempotencysvc.IdempotencyService
	BackPressureMiddleware *custmiddleware.BackpressureMiddleware
	NotificationHandler    *handler.NotificationHandler
}

func SetupRouter(params RouterSetupParams) chi.Router {
	log := logger.Get()
	registry := prometheus.NewRegistry()
	metrics.Register(registry)

	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(custmiddleware.RequestLogger(log))
	r.Use(custmiddleware.MetricsMiddleware)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.NotFound(response.NotFoundHandler())
	r.MethodNotAllowed(response.MethodNotAllowedHandler())

	r.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	// ── PASS 1: Register all the api routes ──────────────────────────────
	r.Route("/api/v1", func(api chi.Router) {
		api.Get("/health/liveness", params.HealthHandler.Liveness)
		api.Get("/health/readiness", params.HealthHandler.Readiness)

		api.Post("/auth/register", params.AuthHandler.RegisterUserHandler)
		api.Post("/auth/login", params.AuthHandler.LoginUserHandler)
		api.Post("/auth/reset-password", params.AuthHandler.ResetPasswordHandler)
		api.Post("/auth/reset-password/validate", params.AuthHandler.ValidatePasswordRequestHandler)
		api.Post("/webhooks/monnify", params.WbHkHandler.HandleMonnifyWebhook)
		api.Post("/simulator/ext", params.SimHandler.SimulateTransferExt)
		api.Post("/tests/idempotency", params.IdempHandler.CreateRecordHandler)
		api.Get("/tests/idempotency", params.IdempHandler.CheckRecordHandler)

		RegisterMonnifyRoutes(api, params.MonnifyHandler)

		api.Group(func(r chi.Router) {
			r.Use(params.Auth.Verifier())
			r.Use(params.Auth.Authenticator())

			r.Get("/users/info", params.UserHandler.GetUserInfo)
			r.Get("/users/balance", params.UserHandler.GetUserBalance)
			r.Get("/users/{email}", params.UserHandler.GetUserByEmailHandler)
			r.Get("/users/transactions", params.UserHandler.GetTransactionHistoryHandler)
			r.Post("/users/register-token", params.UserHandler.RegisterDeviceToken)
			r.Get("/recipient/internal/{acctNum}", params.UserHandler.GetRecipientName)
			r.Get("/transfers/{trxID}/status", params.OperationsHandler.GetTransferStatus)
			r.Post("/notifications/{token}", params.NotificationHandler.SendNotificationHandler)

			r.Group(func(r chi.Router) {
				r.Use(httprate.Limit(
					config.App.Job.RateLimitPerMinute,
					time.Minute,
					httprate.WithKeyFuncs(func(r *http.Request) (string, error) {
						_, claims, _ := jwtauth.FromContext(r.Context())
						userId, ok := claims["userId"].(string)
						if !ok {
							return httprate.KeyByRealIP(r)
						}
						return userId, nil
					}),
				))
				r.Use(custmiddleware.InternalTransferMiddleware(*params.AcctSvc, *params.IdempSvc))
				r.Use(params.BackPressureMiddleware.Handler)
				r.Post("/transfers/internal", params.OperationsHandler.InternalTransfer)
			})
		})

		api.Get("/", handler.LandingPageHandler)
	})

	// ── PASS 2: Now that all routes exist, generate the spec ───────────────
	// spec := openapi.GenerateSpec(r, openapi.Info{
	// 	Title:       "Credora API",
	// 	Description: "Financial simulation and banking platform",
	// 	Version:     "1.0.0",
	// }, []openapi.Server{
	// 	{URL: "https://api.credora.app/api/v1", Description: "Production"},
	// 	{URL: "http://localhost:8080/api/v1", Description: "Local"},
	// })
	var (
		specBytes []byte
		specOnce  sync.Once
	)

	// these can be defined inside the api (api/v1) groupI just wanted clarity of the process hence why I put it outside the gruo but referenced the path
	r.Get("/api/v1/openapi.json", func(w http.ResponseWriter, req *http.Request) {
		specOnce.Do(func() {
			spec := openapi.GenerateSpec(r, openapi.Info{
				Title:       "Credora API",
				Description: "Financial simulation and management platform",
				Version:     "1.0.0",
			}, []openapi.Server{
				{URL: "https://api.credora.app/api/v1", Description: "Production"},
				{URL: "http://localhost:8080/api/v1", Description: "Local"},
			})
			specBytes, _ = json.MarshalIndent(spec, "", "  ")

			// optionally write to disk
			// if bytes, err := json.MarshalIndent(spec, "", "  "); err == nil {
			// 	os.WriteFile("docs/openapi.json", bytes, 0644)
			// }
		})
		w.Header().Set("Content-Type", "application/json")
		w.Write(specBytes)
	})

	// ── PASS 3: Mount the doc routes onto the already-existing /api/v1 group
	r.Route("/api/v1/documentation", func(api chi.Router) {
		openapi.RegisterOpenAPIRoutes(api)
	})

	return r
}

// func SetupRouter(params RouterSetupParams) chi.Router {
// 	log := logger.Get()
// 	registry := prometheus.NewRegistry()

// 	metrics.Register(registry) //Run your apop specific registry

// 	r := chi.NewRouter()

// 	r.Use(middleware.RequestID)
// 	r.Use(middleware.RealIP)
// 	r.Use(middleware.Recoverer)
// 	r.Use(custmiddleware.RequestLogger(log))
// 	r.Use(custmiddleware.MetricsMiddleware)

// 	// Basic CORS
// 	r.Use(cors.Handler(cors.Options{
// 		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
// 		AllowedOrigins: []string{"https://*", "http://*"},
// 		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
// 		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
// 		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
// 		ExposedHeaders:   []string{"Link"},
// 		AllowCredentials: true,
// 		MaxAge:           300, // Maximum value not ignored by any of major browsers
// 	}))

// 	r.NotFound(response.NotFoundHandler())
// 	r.MethodNotAllowed(response.MethodNotAllowedHandler())

// 	r.Handle("/metrics",
// 		promhttp.HandlerFor(registry, promhttp.HandlerOpts{}),
// 	)

// 	r.Route("/api/v1", func(api chi.Router) {
// 		RegisterOpenAPIRoutes(api)

// 		api.Get("/health/liveness", params.HealthHandler.Liveness)
// 		api.Get("/health/readiness", params.HealthHandler.Readiness)

// 		// Public Routes
// 		api.Post("/auth/register", params.AuthHandler.RegisterUserHandler)
// 		api.Post("/auth/login", params.AuthHandler.LoginUserHandler)
// 		api.Post("/auth/reset-password", params.AuthHandler.ResetPasswordHandler)
// 		api.Post("/auth/reset-password/validate", params.AuthHandler.ValidatePasswordRequestHandler)
// 		api.Post("/webhooks/monnify", params.WbHkHandler.HandleMonnifyWebhook)
// 		api.Post("/simulator/ext", params.SimHandler.SimulateTransferExt)
// 		api.Post("/tests/idempotency", params.IdempHandler.CreateRecordHandler)
// 		api.Get("/tests/idempotency", params.IdempHandler.CheckRecordHandler)

// 		RegisterMonnifyRoutes(api, params.MonnifyHandler)

// 		// JWT Auth Middleware ----- Protected Routes -----
// 		api.Group(func(r chi.Router) {
// 			r.Use(params.Auth.Verifier())
// 			r.Use(params.Auth.Authenticator())

// 			r.Get("/users/info", params.UserHandler.GetUserInfo)
// 			r.Get("/users/balance", params.UserHandler.GetUserBalance)
// 			r.Get("/users/{email}", params.UserHandler.GetUserByEmailHandler)
// 			r.Get("/users/transactions", params.UserHandler.GetTransactionHistoryHandler) // with cursor pagination
// 			r.Post("/users/register-token", params.UserHandler.RegisterDeviceToken)
// 			r.Get("/recipient/internal/{acctNum}", params.UserHandler.GetRecipientName)
// 			r.Get("/transfers/{trxID}/status", params.OperationsHandler.GetTransferStatus)
// 			r.Post("/notifications/{token}", params.NotificationHandler.SendNotificationHandler) // test notification
// 			// Internal transfer route
// 			r.Group(func(r chi.Router) {
// 				// r.Use(httprate.LimitByIP(config.App.Job.RateLimitPerMinute,
// 				// 	time.Minute))
// 				r.Use(httprate.Limit(
// 					config.App.Job.RateLimitPerMinute,
// 					time.Minute,
// 					httprate.WithKeyFuncs(func(r *http.Request) (string, error) {
// 						_, claims, _ := jwtauth.FromContext(r.Context())

// 						userAcctNum, ok := claims["accountNumber"].(string)
// 						if !ok {
// 							return httprate.KeyByRealIP(r)
// 						}
// 						return userAcctNum, nil
// 					}),
// 				))
// 				r.Use(custmiddleware.InternalTransferMiddleware(*params.AcctSvc, *params.IdempSvc))
// 				r.Use(params.BackPressureMiddleware.Handler)
// 				r.Post("/transfers/internal", params.OperationsHandler.InternalTransfer)
// 			})

// 		})

// 		api.Get("/", func(w http.ResponseWriter, r *http.Request) {
// 			w.Header().Set("Content-Type", "text/html; charset=utf-8")
// 			fmt.Fprintf(w, `
//         <!DOCTYPE html>
//         <html lang="en">
//         <head>
//             <meta charset="UTF-8" />
//             <meta name="viewport" content="width=device-width, initial-scale=1.0" />
//             <title>Credora API</title>
//             <style>
//                 body {
//                     font-family: system-ui, sans-serif;
//                     background: #f9fafb;
//                     color: #111827;
//                     display: flex;
//                     flex-direction: column;
//                     align-items: center;
//                     justify-content: center;
//                     height: 100vh;
//                     margin: 0;
//                     text-align: center;
//                 }
//                 h1 {
//                     font-size: 2rem;
//                     margin-bottom: 1rem;
//                 }
//                 p {
//                     margin-bottom: 2rem;
//                     color: #4b5563;
//                 }
//                 a {
//                     display: inline-block;
//                     margin: 0.5rem;
//                     padding: 0.75rem 1.5rem;
//                     border-radius: 0.5rem;
//                     text-decoration: none;
//                     color: white;
//                     background-color: #2563eb;
//                     transition: background-color 0.2s;
//                 }
//                 a:hover {
//                     background-color: #1d4ed8;
//                 }
//             </style>
//         </head>
//         <body>
//             <h1>Welcome to Credora API</h1>
//             <p>Your financial simulation and management platform backend.</p>
//             <div>
//                 <a href="https://credora.app" target="_blank">Go to Frontend</a>
//                 <a href="/api/v1/documentation" target="_blank">View API Documentation</a>
//             </div>
//         </body>
//         </html>
//     `)
// 		})

// 	})

// 	return r
// }
