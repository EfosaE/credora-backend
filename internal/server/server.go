// Package server provides HTTP server configuration and management
package server

import (
	"context"

	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EfosaE/credora-backend/internal/config"
	"github.com/go-chi/chi/v5"
)

// ServerConfig holds server configuration options
type ServerConfig struct {
	Port              string
	ReadTimeout       time.Duration
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	MaxHeaderBytes    int
}

// DefaultConfig returns a server configuration with sensible defaults
func DefaultConfig() *ServerConfig {
	return &ServerConfig{
		Port:              ":" + config.App.Port,
		ReadTimeout:       60 * time.Second,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1MB
	}
}

// Server wraps http.Server with additional functionality
type Server struct {
	*http.Server
	config *ServerConfig
}

// New creates a new server instance with the given router and config
func New(router chi.Router, config *ServerConfig) *Server {
	if config == nil {
		config = DefaultConfig()
	}

	srv := &http.Server{
		Addr:              config.Port,
		Handler:           router,
		ReadTimeout:       config.ReadTimeout,
		ReadHeaderTimeout: config.ReadHeaderTimeout,
		WriteTimeout:      config.WriteTimeout,
		IdleTimeout:       config.IdleTimeout,
		MaxHeaderBytes:    config.MaxHeaderBytes,
	}

	return &Server{
		Server: srv,
		config: config,
	}
}

// Start starts the server and handles graceful shutdown
func (s *Server) Start() error {
	log.Printf("Launching HTTP server on %s...", s.config.Port)

	// Channel to listen for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on %s", s.config.Port)
		if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	<-quit
	log.Println("Server shutting down...")

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := s.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
		return err
	}

	log.Println("Server stopped")
	return nil
}

// StartWithoutGracefulShutdown starts the server without graceful shutdown handling
// Useful when you want to handle shutdown logic yourself
func (s *Server) StartWithoutGracefulShutdown() error {
	log.Printf("Server starting on %s", s.config.Port)
	return s.ListenAndServe()
}
