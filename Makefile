.PHONY: help migrate-up migrate-down migrate-status migrate-create promote-user db-recreate

# Database configuration (can be overridden with environment variables)
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= familytree
DB_NAME ?= familytree
DB_PASSWORD ?= secret

help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

migrate-up: ## Run database migrations (up) using goose
	@echo "Running migrations with goose..."
	@docker run --rm --network host \
		-e DB_HOST=$(DB_HOST) \
		-e DB_PORT=$(DB_PORT) \
		-e DB_USER=$(DB_USER) \
		-e DB_PASSWORD=$(DB_PASSWORD) \
		-e DB_NAME=$(DB_NAME) \
		-e DB_SSL=disable \
		-v $(PWD)/be/migrations:/migrations \
		gomicro/goose:latest \
		/bin/bash -c "cd /migrations && goose postgres 'host=$(DB_HOST) port=$(DB_PORT) user=$(DB_USER) password=$(DB_PASSWORD) dbname=$(DB_NAME) sslmode=disable' up"

migrate-down: ## Run database migrations (down) using goose
	@echo "Rolling back migrations with goose..."
	@docker run --rm --network host \
		-e DB_HOST=$(DB_HOST) \
		-e DB_PORT=$(DB_PORT) \
		-e DB_USER=$(DB_USER) \
		-e DB_PASSWORD=$(DB_PASSWORD) \
		-e DB_NAME=$(DB_NAME) \
		-e DB_SSL=disable \
		-v $(PWD)/be/migrations:/migrations \
		gomicro/goose:latest \
		/bin/bash -c "cd /migrations && goose postgres 'host=$(DB_HOST) port=$(DB_PORT) user=$(DB_USER) password=$(DB_PASSWORD) dbname=$(DB_NAME) sslmode=disable' down"

migrate-status: ## Check migration status using goose
	@echo "Checking migration status with goose..."
	@docker run --rm --network host \
		-e DB_HOST=$(DB_HOST) \
		-e DB_PORT=$(DB_PORT) \
		-e DB_USER=$(DB_USER) \
		-e DB_PASSWORD=$(DB_PASSWORD) \
		-e DB_NAME=$(DB_NAME) \
		-e DB_SSL=disable \
		-v $(PWD)/be/migrations:/migrations \
		gomicro/goose:latest \
		/bin/bash -c "cd /migrations && goose postgres 'host=$(DB_HOST) port=$(DB_PORT) user=$(DB_USER) password=$(DB_PASSWORD) dbname=$(DB_NAME) sslmode=disable' status"

migrate-create: ## Create a new migration file (Usage: make migrate-create NAME=migration_name)
	@if [ -z "$(NAME)" ]; then \
		echo "Error: NAME is required"; \
		echo "Usage: make migrate-create NAME=migration_name"; \
		exit 1; \
	fi
	@echo "Creating new migration: $(NAME)"
	@docker run --rm \
		-v $(PWD)/be/migrations:/migrations \
		gomicro/goose:latest \
		/bin/bash -c "cd /migrations && goose create $(NAME) sql"

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

# Production deployment commands
prod-setup: ## Setup VPS for production (run on VPS)
	@echo "Setting up VPS for production..."
	chmod +x scripts/setup-vps.sh
	sudo ./scripts/setup-vps.sh

prod-init: ## Initialize production environment and SSL certificates
	@echo "Initializing production environment..."
	chmod +x scripts/init-ssl.sh
	./scripts/init-ssl.sh

prod-deploy: ## Deploy application to production
	@echo "Deploying to production..."
	chmod +x scripts/deploy.sh
	./scripts/deploy.sh

prod-up: ## Start production services
	docker-compose -f docker-compose.prod.yml up -d

prod-down: ## Stop production services
	docker-compose -f docker-compose.prod.yml down

prod-logs: ## View production logs
	docker-compose -f docker-compose.prod.yml logs -f

prod-status: ## Check production services status
	docker-compose -f docker-compose.prod.yml ps

prod-restart: ## Restart production services
	docker-compose -f docker-compose.prod.yml restart

prod-backup: ## Backup production data
	@echo "Creating backup..."
	chmod +x scripts/backup.sh
	./scripts/backup.sh

prod-maintenance: ## Run maintenance menu
	@echo "Starting maintenance menu..."
	chmod +x scripts/maintenance.sh
	./scripts/maintenance.sh

prod-shell-db: ## Access production database shell
	docker-compose -f docker-compose.prod.yml exec postgres psql -U familytree -d familytree

prod-shell-redis: ## Access production Redis shell
	docker-compose -f docker-compose.prod.yml exec redis redis-cli

prod-clean: ## Clean up production Docker resources
	docker system prune -f
