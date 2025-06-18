package main

import (
	"context"
	"log"

	"github.com/EfosaE/credora-backend/internal/config"
	"github.com/EfosaE/credora-backend/internal/db"
	"github.com/EfosaE/credora-backend/internal/server"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func main() {

	ctx := context.Background()
	config.Load()
	db, err := db.InitDB(ctx)

	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Pool.Close()

	r := chi.NewRouter()

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

	// Option 1: Use default configuration
	srv := server.New(r, nil)
	// Creating a Sub Router
	apiRouter := chi.NewRouter()
	// Mounting the new Sub Router on the main router
	r.Mount("/api/v1", apiRouter)

	// apiRouter.Post("/upload", handlers.UploadHandler)

	// // File upload endpoint
	// r.HandleFunc("/api/upload", uploadHandler).Methods("POST")

	// // Protected download endpoint
	// r.HandleFunc("/api/download/{fileId}", protectedDownloadHandler).Methods("GET")

	// Start server with graceful shutdown
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
}
