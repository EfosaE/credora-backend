# Include environment variables from .env file
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# Build DATABASE_URL from components or use direct value
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

.PHONY: migrate-up migrate-down migrate-force migrate-version sqlc-generate

migrate-up:
	migrate -database $(DB_URL) -path internal/db/migrations up

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

sqlc-generate:
	sqlc generate

dev-setup: migrate-up sqlc-generate
# Check migration status
migrate-status:
	migrate -database $(DATABASE_URL) -path internal/db/migrations version

# Rollback one migration
migrate-rollback:
	migrate -database $(DATABASE_URL) -path internal/db/migrations down 1

# Drop all migrations
migrate-drop:
	migrate -database $(DATABASE_URL) -path internal/db/migrations drop -f

run:
	go run cmd/server/main.go

build:
	go build -o bin/server cmd/main.go
test-app:
	go test -v ./...