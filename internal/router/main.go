package router

import (
	"fmt"
	"net/http"

	"github.com/EfosaE/credora-backend/internal/handler"
	"github.com/EfosaE/credora-backend/internal/response"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func SetupRouter(userHandler *handler.UserHandler, monnifyHandler *handler.MonnifyHandler) chi.Router {
	r := chi.NewRouter()

	// Add recovery middleware first!
	// r.Use(middleware.Recoverer)
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

	// // Mounting the new Sub Router on the main router
	// r.Mount("/api/v1", apiRouter)

	// // Registering the routes on apiRouter
	// apiRouter.Get("/", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Write([]byte("Welcome to Credora API v1"))
	// })

	// // Register your app routes here
	// RegisterUserRoutes(apiRouter, userHandler)

	r.Route("/api/v1", func(api chi.Router) {
		// 1. Specific fixed routes first
		RegisterUserRoutes(api, userHandler)
		RegisterMonnifyRoutes(api, monnifyHandler)

		// 2. Catch-all {name} route last
		api.Get("/{name}", func(w http.ResponseWriter, r *http.Request) {
			name := chi.URLParam(r, "name")
			fmt.Fprintf(w, "Welcome To Creadora API %s", name)
		})
	})

	return r
}
