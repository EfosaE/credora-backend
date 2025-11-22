package main

import (
	"context"
	"flag"
	"log"

	"github.com/EfosaE/credora-backend/internal/config"
	"github.com/EfosaE/credora-backend/internal/seeder"
	"github.com/jackc/pgx/v5"
)

func main() {
	// Load app configuration
	config.Load()

	// Seed mode
	seedFlag := flag.Bool("seed", false, "Seed the database with fake data")
	flag.Parse()

	if *seedFlag {
		ctx := context.Background()
		conn, err := pgx.Connect(ctx, config.App.TestDbUrl)
		if err != nil {
			log.Fatalf("DB connect failed: %v", err)
		}
		defer conn.Close(ctx)

		s := seeder.NewSeeder(conn, ctx)
		if err := s.SeedUsersAndAccounts(5); err != nil {
			log.Fatalf("Seeding failed: %v", err)
		}

		log.Println("✅ Seeding complete")
		return
	}

}
