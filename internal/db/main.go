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
	Pool    *pgxpool.Pool // Connection pool for PostgreSQL (DB)
	Queries *sqlc.Queries // my sqlc package supports pgx instead of database/sql
)

type DB struct {
	Pool    *pgxpool.Pool
	Queries *sqlc.Queries
}

// InitDB initializes the database connection for production/development
func InitDB(ctx context.Context) (*DB, error) {
	pingCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// Parse pool config
	cfg, err := pgxpool.ParseConfig(config.App.DbUrl)
	if err != nil {
		return nil, fmt.Errorf("db: failed to parse db url: %w", err)
	}

	// Explicit pool configuration (IMPORTANT)
	cfg.MinConns = 2
	cfg.MaxConns = 8
	cfg.MaxConnLifetime = time.Hour
	cfg.MaxConnIdleTime = 30 * time.Minute

	fmt.Println("🛢️  Initializing PostgreSQL connection pool...")
	fmt.Printf("   ➜ MinConns: %d\n", cfg.MinConns)
	fmt.Printf("   ➜ MaxConns: %d\n", cfg.MaxConns)
	fmt.Printf("   ➜ MaxConnLifetime: %s\n", cfg.MaxConnLifetime)
	fmt.Printf("   ➜ MaxConnIdleTime: %s\n", cfg.MaxConnIdleTime)

	// Create pool
	pool, err := pgxpool.NewWithConfig(pingCtx, cfg)
	if err != nil {
		return nil, fmt.Errorf("db: failed to create pool: %w", err)
	}

	// Verify connectivity
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("db: failed to ping database: %w", err)
	}

	fmt.Println("✅ Database connection established")

	// Log initial pool stats
	logPoolStats("startup", pool)

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

func logPoolStats(stage string, pool *pgxpool.Pool) {
	stats := pool.Stat()

	fmt.Println("📊 DB Pool Stats (" + stage + ")")
	fmt.Printf("   ➜ MaxConns:       %d\n", stats.MaxConns())
	fmt.Printf("   ➜ TotalConns:     %d\n", stats.TotalConns())
	fmt.Printf("   ➜ IdleConns:      %d\n", stats.IdleConns())
	fmt.Printf("   ➜ AcquiredConns:  %d\n", stats.AcquiredConns())
	fmt.Printf("   ➜ Constructing:   %d\n", stats.ConstructingConns())
	fmt.Printf("   ➜ EmptyAcquire:   %d\n", stats.EmptyAcquireCount())
	fmt.Println()
}
