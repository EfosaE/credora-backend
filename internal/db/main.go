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
	Queries *sqlc.Queries  // my sqlc package supports pgx instead of database/sql
)

type DB struct {
	Pool    *pgxpool.Pool
	Queries *sqlc.Queries
}

// InitDB initializes the database connection for production/development
func InitDB(ctx context.Context) (*DB, error) {
	pingCtx, cancel := context.WithTimeout(ctx, time.Second*60)
	defer cancel()

	// Create a connection pool
	pool, err := pgxpool.New(pingCtx, config.App.DbUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify the connection
	if err = pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	fmt.Printf("Connection to database successful ✅\n")

	// Initialize queries with the connection pool
	queries := sqlc.New(pool)

	return &DB{
		Pool:    pool,
		Queries: queries,
	}, nil
}

// InitTestDB initializes the database connection for local testing
func InitTestDB(ctx context.Context) (*DB, error) {
	pingCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	// Create a connection pool
	pool, err := pgxpool.New(pingCtx, "postgres://efosa:secret@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify the connection
	if err = pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	fmt.Printf("Connection to test database successful ✅\n")

	// Initialize queries with the connection pool
	queries := sqlc.New(pool)

	return &DB{
		Pool:    pool,
		Queries: queries,
	}, nil
}

// InitContainerDB initializes the database from an existing pool (for testcontainers)
func InitContainerDB(ctx context.Context, pool *pgxpool.Pool) (*DB, error) {
	pingCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	// Verify the connection
	if err := pool.Ping(pingCtx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	fmt.Printf("Connection to container database successful ✅\n")

	// Initialize queries with the connection pool
	queries := sqlc.New(pool)

	return &DB{
		Pool:    pool,
		Queries: queries,
	}, nil
}

// Close closes the database connection pool
func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}