# Credora Backend

A Go-based backend for financial simulations, virtual accounts and transaction processing with event-driven and queued workers.

## Highlights (latest)
- Go 1.24+
- JWT authentication + account management
- Monnify integration for virtual/reserved accounts
- PostgreSQL (pgx) + SQLC for typed SQL
- Redis-based Event Bus + Asynq workers for background jobs
- Email delivery adapters (Resend / Mailtrap) and event-driven welcome/account emails
- Structured JSON logging with file rotation support

## Features
- Authentication & Authorization
  - JWT token service
  - Register / Login endpoints
- Virtual Accounts
  - Reserved virtual accounts via Monnify
  - Virtual account bookkeeping and webhooks
- Transactions & Transfers
  - Internal transfer processing with idempotency
  - Webhook handling for Monnify payment events
  - Transaction history and settlement calculation
- Asynchronous Processing
  - Redis + Asynq queue for email, account notifications and internal transfer tasks
- Developer tooling
  - Database migrations (golang-migrate)
  - SQLC for generating query code
  - Makefile helpers for common operations

## Tech stack & key libraries
- Language: Go (>= 1.24)
- Router: github.com/go-chi/chi
- DB driver: github.com/jackc/pgx/v5
- SQL generation: github.com/kyleconroy/sqlc
- Migrations: golang-migrate/migrate
- Redis client: github.com/redis/go-redis/v9
- Queue: github.com/hibiken/asynq
- Logging: custom structured logger (domain/logger)
- Monnify client: internal/infrastructure adapter (monnify)
- Utilities: github.com/shopspring/decimal, github.com/google/uuid, github.com/brianvoe/gofakeit, github.com/joho/godotenv

## Environment (example)
Create a `.env` with at least:
- DATABASE_URL (postgres url)
- TEST_DATABASE_URL
- PORT (e.g., 8080)
- JWT_SECRET
- REDIS_ADDR (e.g., localhost:6379)
- MONNIFY_API_KEY, MONNIFY_SECRET_KEY, MONNIFY_CONTRACT_CODE, MONNIFY_BASE_URL
- RESEND_API_KEY (or Mailtrap settings)
- WEBHOOK_URL

## Quick start — local development

1. Install dependencies:
   - Go 1.24+
   - PostgreSQL
   - Redis
   - golang-migrate (for migrations)
   - sqlc (if you need to regenerate queries)

2. Prepare DB & Redis and set environment variables (.env).

3. Run migrations:
   - make migrate-up
   - or provide DATABASE_URL interactively: make migrate-up

4. Generate SQLC code (if needed):
   - make sqlc-generate

5. Run the HTTP server:
   - make run
   - or: go run cmd/server/main.go
   - The server listens on PORT from config (default 8080). API root: /api/v1

6. Run the worker (background job processor):
   - make start-worker
   - or: go run cmd/worker/main.go
   - Ensure REDIS_ADDR is reachable; the worker processes Asynq tasks (email, internal transfer, etc.)

7. CLI utilities:
   - Seed database with fake data:
     go run cmd/cli/main.go -seed
   - The CLI uses TEST_DATABASE_URL from env for seeding by default.

## Docker
Build and run a container (example):
- Build:
  docker build -t credora:latest .
- Run (provide envs):
  docker run -e DATABASE_URL="postgres://..." -e REDIS_ADDR="redis:6379" -p 8080:8080 credora:latest

Notes:
- Use a separate worker container for Asynq workers (run the worker binary / cmd/worker).
- The Dockerfile builds a "server" binary under /app and exposes 8080.

## Testing
- Unit tests:
  make test
- Integration tests:
  make test-integration
- Run all:
  make test-all

## Useful Makefile targets
- make run            — run server
- make build          — build server binary
- make start-worker   — run worker (go run cmd/worker/main.go)
- make migrate-up     — run DB migrations (interactive or use DATABASE_URL)
- make migrate-create — scaffold a new SQL migration
- make sqlc-generate  — regenerate SQLC code
- make test           — run unit tests
- make test-integration — run integration tests

## Project layout (short)
- cmd/         — server, worker, cli entrypoints
- internal/    — config, router, server, handlers, queues, seeder, db
- infrastructure/ — adapters and repo implementations
- service/     — business logic
- domain/      — domain models and shared utilities


