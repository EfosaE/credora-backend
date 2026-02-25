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
	config.Load()

	seedFlag := flag.Bool("seed", false, "Seed database")
	usersFlag := flag.Int("users", 500, "Number of users to create")
	hotFlag := flag.Bool("hot", true, "Create hot/system accounts")
	dbURLFlag := flag.String("db", config.App.DbUrl, "Database URL override")

	flag.Parse()

	if !*seedFlag {
		return
	}

	ctx := context.Background()

	conn, err := pgx.Connect(ctx, *dbURLFlag)
	if err != nil {
		log.Fatalf("DB connect failed: %v", err)
	}
	defer conn.Close(ctx)

	s := seeder.NewSeeder(conn, ctx)

	// if *usersFlag > 60 {
	// 	log.Printf("Seeding %d users...", *usersFlag)
	// 	if err := s.SeedUsersAndAccounts(*usersFlag); err != nil {
	// 		log.Fatalf("User seeding failed: %v", err)
	// 	}
	// }

	if err := s.SeedUsersAndAccounts(*usersFlag); err != nil {
		log.Fatalf("User seeding failed: %v", err)
	}

	if *hotFlag {
		log.Println("Seeding hot/system accounts...")
		if err := s.SeedHotAccounts(); err != nil {
			log.Fatalf("Hot account seeding failed: %v", err)
		}
	}

	log.Println("Seeding complete.")
}
