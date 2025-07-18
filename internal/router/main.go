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

func SetupRouter(authHandler *handler.AuthHandler, userHandler *handler.UserHandler, monnifyHandler *handler.MonnifyHandler, auth *authsvc.JWTTokenService) chi.Router {
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
		api.Post("/auth/register", authHandler.RegisterUserHandler)
		api.Post("/auth/login", authHandler.LoginUserHandler)

		// RegisterUserRoutes(api, userHandler)
		RegisterMonnifyRoutes(api, monnifyHandler)


		// JWT Auth Middleware ----- Protected Routes -----
		api.Group(func(r chi.Router) {
			r.Use(auth.Verifier())
			r.Use(auth.Authenticator())

			r.Get("/user/info", userHandler.GetUserInfo)
		})

		// 2. Catch-all {name} route last
		api.Get("/{name}", func(w http.ResponseWriter, r *http.Request) {
			name := chi.URLParam(r, "name")
			fmt.Fprintf(w, "Welcome To Creadora API %s", name)
		})
	})

	return r
}
