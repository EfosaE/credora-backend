# ============================================
# Load environment variables from .env
# ============================================
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# ============================================
# Database URL setup
# ============================================
ifdef DATABASE_URL
    DB_URL = $(DATABASE_URL)
else
    DB_HOST ?= localhost
    DB_PORT ?= 5432
    DB_USER ?= postgres
    DB_PASSWORD ?= password
    DB_NAME ?= mydb
    DB_SSL ?= disable
    DB_URL = postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL)
endif

# ============================================
# Migration commands
# ============================================
.PHONY: migrate-up migrate-down migrate-down-all migrate-force migrate-version migrate-create migrate-status migrate-rollback migrate-drop

# migrate-up:
# 	migrate -database $(DB_URL) -path internal/db/migrations up

migrate-up:
	@read -p "Enter DATABASE_URL: " DB_URL; \
	migrate -database "$$DB_URL" -path internal/db/migrations up

migrate-down:
	migrate -database $(DB_URL) -path internal/db/migrations down

migrate-down-all:
	migrate -database $(DB_URL) -path internal/db/migrations down -all

migrate-force:
	@echo "Enter the version to force (e.g., 4, 5, 6):"
	@read -p "> " version; \
	if [ -z "$$version" ]; then \
		echo "Error: Version cannot be empty"; \
		exit 1; \
	fi; \
	migrate -database $(DB_URL) -path internal/db/migrations force $$version; \
	echo "Database forced to version $$version successfully!"

migrate-version:
	migrate -database $(DB_URL) -path internal/db/migrations version

migrate-create:
	@echo "Enter migration name (e.g., add_users_table, update_posts_index):"
	@read -p "> " name; \
	if [ -z "$$name" ]; then \
		echo "Error: Migration name cannot be empty"; \
		exit 1; \
	fi; \
	migrate create -ext sql -dir internal/db/migrations -seq $$name; \
	echo "Migration created successfully!"

migrate-status:
	migrate -database $(DB_URL) -path internal/db/migrations version

migrate-rollback:
	migrate -database $(DB_URL) -path internal/db/migrations down 1

migrate-drop:
	migrate -database $(DB_URL) -path internal/db/migrations drop -f

# ============================================
# SQLC
# ============================================
.PHONY: sqlc-generate
sqlc-generate:
	sqlc generate

# ============================================
# Run & Build
# ============================================
.PHONY: run build start-worker
run:
	go run cmd/server/main.go

build:
	go build -o bin/server cmd/server/main.go

start-worker:
	go run cmd/worker/main.go

	
# ============================================
# Tests: unit vs integration
# ============================================
UNIT_TAGS=
INTEGRATION_TAGS=integration

.PHONY: test
test:
	go test -v -tags=$(UNIT_TAGS) ./...

.PHONY: test-integration
test-integration:
	go test -v -tags=$(INTEGRATION_TAGS) ./...

.PHONY: test-all
test-all: test test-integration
