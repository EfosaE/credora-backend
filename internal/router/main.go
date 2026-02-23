package router

import (
	"fmt"
	"net/http"
	"time"

	"github.com/EfosaE/credora-backend/infrastructure"
	"github.com/EfosaE/credora-backend/internal/config"
	"github.com/EfosaE/credora-backend/internal/handler"
	custmiddleware "github.com/EfosaE/credora-backend/internal/middleware"
	"github.com/EfosaE/credora-backend/internal/response"
	accountsvc "github.com/EfosaE/credora-backend/service/account"
	authsvc "github.com/EfosaE/credora-backend/service/auth"
	idempotencysvc "github.com/EfosaE/credora-backend/service/idempotency"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
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
}

func SetupRouter(params RouterSetupParams) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	// r.Use(middleware.Recoverer)

	// Basic CORS
	r.Use(cors.Handler(cors.Options{
		// AllowedOrigins:   []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"https://*", "http://*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	r.NotFound(response.NotFoundHandler())
	r.MethodNotAllowed(response.MethodNotAllowedHandler())

	r.Route("/api/v1", func(api chi.Router) {
		RegisterOpenAPIRoutes(api)

		api.Get("/health/liveness", params.HealthHandler.Liveness)
		api.Get("/health/readiness", params.HealthHandler.Readiness)

		// Public Routes
		api.Post("/auth/register", params.AuthHandler.RegisterUserHandler)
		api.Post("/auth/login", params.AuthHandler.LoginUserHandler)
		api.Post("/auth/reset-password", params.AuthHandler.ResetPasswordHandler)
		api.Post("/auth/reset-password/validate", params.AuthHandler.ValidatePasswordRequestHandler)
		api.Post("/webhooks/monnify", params.WbHkHandler.HandleMonnifyWebhook)
		api.Post("/simulator/ext", params.SimHandler.SimulateTransferExt)
		api.Post("/tests/idempotency", params.IdempHandler.CreateRecordHandler)
		api.Get("/tests/idempotency", params.IdempHandler.CheckRecordHandler)

		RegisterMonnifyRoutes(api, params.MonnifyHandler)

		// JWT Auth Middleware ----- Protected Routes -----
		api.Group(func(r chi.Router) {
			r.Use(params.Auth.Verifier())
			r.Use(params.Auth.Authenticator())

			r.Get("/user/info", params.UserHandler.GetUserInfo)
			r.Get("/user/balance", params.UserHandler.GetUserBalance)
			r.Get("/user/transactions", params.UserHandler.GetTransactionHistoryHandler) // with cursor pagination
			r.Get("/recipient/internal/{acctNum}", params.UserHandler.GetRecipientName)
			r.Get("/transfers/{trxID}/status", params.OperationsHandler.GetTransferStatus)
			// Internal transfer route
			r.Group(func(r chi.Router) {
				r.Use(httprate.LimitByIP(config.App.Job.RateLimitPerMinute,
					time.Minute))
				r.Use(custmiddleware.InternalTransferMiddleware(*params.AcctSvc, *params.IdempSvc))
				r.Use(params.BackPressureMiddleware.Handler)
				r.Post("/transfers/internal", params.OperationsHandler.InternalTransfer)
			})

		})

		api.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprintf(w, `
        <!DOCTYPE html>
        <html lang="en">
        <head>
            <meta charset="UTF-8" />
            <meta name="viewport" content="width=device-width, initial-scale=1.0" />
            <title>Credora API</title>
            <style>
                body {
                    font-family: system-ui, sans-serif;
                    background: #f9fafb;
                    color: #111827;
                    display: flex;
                    flex-direction: column;
                    align-items: center;
                    justify-content: center;
                    height: 100vh;
                    margin: 0;
                    text-align: center;
                }
                h1 {
                    font-size: 2rem;
                    margin-bottom: 1rem;
                }
                p {
                    margin-bottom: 2rem;
                    color: #4b5563;
                }
                a {
                    display: inline-block;
                    margin: 0.5rem;
                    padding: 0.75rem 1.5rem;
                    border-radius: 0.5rem;
                    text-decoration: none;
                    color: white;
                    background-color: #2563eb;
                    transition: background-color 0.2s;
                }
                a:hover {
                    background-color: #1d4ed8;
                }
            </style>
        </head>
        <body>
            <h1>Welcome to Credora API</h1>
            <p>Your financial simulation and management platform backend.</p>
            <div>
                <a href="https://credora.app" target="_blank">Go to Frontend</a>
                <a href="/api/v1/documentation" target="_blank">View API Documentation</a>
            </div>
        </body>
        </html>
    `)
		})

	})

	return r
}
