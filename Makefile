.PHONY: help migrate-up migrate-down promote-user db-recreate

# Database configuration (can be overridden with environment variables)
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= familytree
DB_NAME ?= familytree
DB_PASSWORD ?= secret

help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

migrate-up: ## Run database migrations (up)
	@echo "Running migrations..."
	@for file in be/migrations/*_*.up.sql; do \
		echo "Applying migration: $$file"; \
		PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -f $$file || exit 1; \
	done
	@echo "✅ Migrations completed successfully!"

migrate-down: ## Run database migrations (down)
	@echo "Rolling back migrations..."
	@for file in $$(ls -r be/migrations/*_*.down.sql); do \
		echo "Rolling back migration: $$file"; \
		PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME) -f $$file || exit 1; \
	done
	@echo "✅ Migrations rolled back successfully!"

db-recreate: ## Drop and recreate database, then run migrations
	@echo "⚠️  WARNING: This will DELETE all data in the database!"
	@echo "Press Ctrl+C to cancel, or Enter to continue..."
	@read confirm
	@echo "Dropping database..."
	@PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U postgres -c "DROP DATABASE IF EXISTS $(DB_NAME);"
	@echo "Creating database..."
	@PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U postgres -c "CREATE DATABASE $(DB_NAME);"
	@PGPASSWORD=$(DB_PASSWORD) psql -h $(DB_HOST) -p $(DB_PORT) -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE $(DB_NAME) TO $(DB_USER);"
	@echo "Running migrations..."
	@$(MAKE) migrate-up
	@echo "✅ Database recreated successfully!"

promote-user: ## Promote a user to super_admin (Usage: make promote-user EMAIL=user@example.com)
	@if [ -z "$(EMAIL)" ]; then \
		echo "Error: EMAIL is required"; \
		echo "Usage: make promote-user EMAIL=user@example.com"; \
		exit 1; \
	fi
	@echo "Promoting user: $(EMAIL)"
	@DB_HOST=$(DB_HOST) DB_PORT=$(DB_PORT) DB_USER=$(DB_USER) DB_NAME=$(DB_NAME) DB_PASSWORD=$(DB_PASSWORD) \
		./be/scripts/promote-user.sh $(EMAIL)

# Docker commands
docker-up: ## Start all services with docker-compose
	docker-compose up -d

docker-down: ## Stop all services
	docker-compose down

docker-logs: ## Show docker logs
	docker-compose logs -f

docker-recreate: ## Recreate all containers
	docker-compose down -v
	docker-compose up -d

# Development commands
dev-backend: ## Run backend in development mode
	cd be && go run cmd/main.go

dev-frontend: ## Run frontend in development mode
	cd fe && npm run dev

test-backend: ## Run backend tests
	cd be && go test ./...

build-backend: ## Build backend binary
	cd be && go build -o main cmd/main.go

build-frontend: ## Build frontend for production
	cd fe && npm run build
