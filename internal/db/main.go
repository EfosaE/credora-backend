package db

import (
	"context"
	"fmt"
	"time"

	"github.com/EfosaE/credora-backend/internal/config"
	"github.com/EfosaE/credora-backend/internal/db/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	Pool    *pgxpool.Pool  // Connection pool for PostgreSQL (DB)
	Queries *sqlc.Queries // my sqlc package supports pgx instead of database/sql
)

type DB struct {
    Pool    *pgxpool.Pool
    Queries *sqlc.Queries
}

func InitDB(ctx context.Context) (*DB, error) {
    pingCtx, cancel := context.WithTimeout(ctx, time.Second*10) // Set a timeout for the database connection
    defer cancel()
    // Create a connection pool
    pool, err := pgxpool.New(pingCtx, config.App.DbUrl)
    if err != nil {
        return nil, fmt.Errorf("failed to create connection pool: %w", err)
    }

   
    // Verify the connection
    if err = pool.Ping(pingCtx); err != nil {
        return nil, fmt.Errorf("%w", err)
    }
    
    fmt.Printf("Connection to database successful ✅\n")

    // Initialize queries with the connection pool
    queries := sqlc.New(pool)

    return &DB{
        Pool:    pool,
        Queries: queries,
    }, nil
}
