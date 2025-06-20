package main

import (
	"context"
	"log"
	"net/http"

	// "net/http"

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
	db, err := db.InitDB(ctx)

	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Pool.Close()

	userRepo := infrastructure.NewSqlcUserRepository(ctx, db.Queries)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	// user, err := userService.CreateUser(ctx, &domain.CreateUserRequest{
	// 	Name:  "John Doe",
	// 	Email: "john.doe@example.com",
	// })
	// if err != nil {
	// 	log.Fatalf("Failed to create user: %v", err)
	// }

	// fmt.Printf("User created: %+v\n", user)

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
