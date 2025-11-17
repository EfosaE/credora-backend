package router

import (
	"fmt"
	"net/http"

	"github.com/EfosaE/credora-backend/internal/handler"
	"github.com/EfosaE/credora-backend/internal/response"
	authsvc "github.com/EfosaE/credora-backend/service/auth"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	// "github.com/go-chi/jwtauth/v5"
)

func SetupRouter(authHandler *handler.AuthHandler, operationsHandler *handler.OperationHandler, userHandler *handler.UserHandler, monnifyHandler *handler.MonnifyHandler, auth *authsvc.JWTTokenService, wbHkHandler *handler.WebHookHandler, simHandler *handler.SimulatorHandler) chi.Router {
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

		api.Post("/auth/register", authHandler.RegisterUserHandler)
		api.Post("/auth/login", authHandler.LoginUserHandler)
		api.Post("/webhooks/monnify", wbHkHandler.HandleMonnifyWebhook)
		api.Post("/simulator/ext", simHandler.SimulateTransferExt)

		// RegisterUserRoutes(api, userHandler)
		RegisterMonnifyRoutes(api, monnifyHandler)

		// JWT Auth Middleware ----- Protected Routes -----
		api.Group(func(r chi.Router) {
			r.Use(auth.Verifier())
			r.Use(auth.Authenticator())

			r.Get("/user/info", userHandler.GetUserInfo)
			// Internal transfer route
			r.Post("/transfer/internal", operationsHandler.InternalTransfer)
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
