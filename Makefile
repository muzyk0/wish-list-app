# Makefile for Wish List Application

.PHONY: help
help: ## Show this help message
	@echo "Wish List Application - Development Commands"
	@echo ""
	@grep -E '^[a-zA-Z_0-9%-]+:.*?## .*$$' $(word 1,$(MAKEFILE_LIST)) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "OpenAPI Commands:"
	@echo "  \033[36mopenapi-bundle\033[0m               Bundle OpenAPI specification from split files"
	@echo "  \033[36mopenapi-validate\033[0m             Validate OpenAPI specification"
	@echo "  \033[36mopenapi-preview\033[0m              Preview OpenAPI specification in browser"
	@echo "  \033[36mopenapi-info\033[0m                 Show information about the OpenAPI specification"

.PHONY: setup
setup: ## Set up the development environment
	@echo "Setting up development environment..."
	@cd backend && go mod tidy
	@cd frontend && pnpm install
	@cd mobile && pnpm install
	@echo "Ensure golangci-lint is installed (recommended: install via package manager like brew)"
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "golangci-lint not found. Install with: brew install golangci-lint"; \
		exit 1; \
	fi

.PHONY: db-up
db-up: ## Start the database with Docker
	@echo "Starting database with Docker..."
	@cd database && docker-compose up -d postgres redis

.PHONY: db-down
db-down: ## Stop the database
	@echo "Stopping database..."
	@cd database && docker-compose down

.PHONY: docker-up
docker-up: ## Start all services (database + backend) with Docker
	@echo "Starting all services with Docker..."
	@cd database && docker-compose up -d

.PHONY: docker-down
docker-down: ## Stop all Docker services
	@echo "Stopping all Docker services..."
	@cd database && docker-compose down

.PHONY: docker-build
docker-build: ## Build the backend Docker image
	@echo "Building backend Docker image..."
	@cd database && docker-compose build backend

.PHONY: docker-logs
docker-logs: ## Show logs from all Docker services
	@cd database && docker-compose logs -f

.PHONY: docker-logs-backend
docker-logs-backend: ## Show logs from backend service
	@cd database && docker-compose logs -f backend

.PHONY: docker-restart
docker-restart: ## Restart all Docker services
	@echo "Restarting all Docker services..."
	@cd database && docker-compose restart

.PHONY: docker-restart-backend
docker-restart-backend: ## Restart backend service
	@echo "Restarting backend service..."
	@cd database && docker-compose restart backend

.PHONY: docker-ps
docker-ps: ## Show running Docker containers
	@cd database && docker-compose ps

.PHONY: docker-clean
docker-clean: ## Remove all containers, volumes, and images
	@echo "Cleaning up Docker resources..."
	@cd database && docker-compose down -v --rmi all

.PHONY: migrate-up
migrate-up: ## Run database migrations
	@echo "Running database migrations..."
	@cd backend && go run cmd/migrate/main.go -action up

.PHONY: migrate-down
migrate-down: ## Rollback database migrations
	@echo "Rolling back database migrations..."
	@cd backend && go run cmd/migrate/main.go -action down

.PHONY: backend
backend: ## Start the backend server
	@echo "Starting backend server..."
	@cd backend && go run cmd/server/main.go

.PHONY: frontend
frontend: ## Start the frontend server
	@echo "Starting frontend server..."
	@cd frontend && pnpm run dev

.PHONY: mobile
mobile: ## Start the mobile development server
	@echo "Starting mobile development server..."
	@cd mobile && pnpm expo start

.PHONY: lint
lint: ## Run lint for all components
	@echo "Running lint for all components..."
	@cd backend && golangci-lint run
	@cd frontend && pnpm run lint
	@cd mobile && pnpm run lint

.PHONY: format
format: ## Format all components with biome
	@echo "Formatting all components with biome..."
	@cd backend && go fmt ./...
	@cd frontend && pnpm run format
	@cd mobile && pnpm run format

.PHONY: format-backend
format-backend: ## Format backend with go fmt
	@echo "Formatting backend..."
	@cd backend && go fmt ./...

.PHONY: test
test: ## Run tests for all components
	@echo "Running tests..."
	@cd backend && go test ./...
	@cd frontend && pnpm test
	@cd mobile && pnpm test

.PHONY: lint-backend
lint-backend: ## Run golangci-lint on backend
	@echo "Running golangci-lint on backend..."
	@cd backend && golangci-lint run

.PHONY: lint-frontend
lint-frontend: ## Run lint on frontend
	@echo "Running lint on frontend..."
	@cd frontend && pnpm run lint

.PHONY: lint-mobile
lint-mobile: ## Run lint on mobile
	@echo "Running lint on mobile..."
	@cd mobile && pnpm run lint

.PHONY: test-backend
test-backend: ## Run backend tests
	@echo "Running backend tests..."
	@cd backend && go test ./...

.PHONY: test-frontend
test-frontend: ## Run frontend tests
	@echo "Running frontend tests..."
	@cd frontend && pnpm test

.PHONY: test-mobile
test-mobile: ## Run mobile tests
	@echo "Running mobile tests..."
	@cd mobile && pnpm test

.PHONY: build
build: ## Build all components
	@echo "Building all components..."
	@cd backend && go build -o bin/server cmd/server/main.go
	@cd frontend && pnpm run build
	@cd mobile && pnpm expo export:web

.PHONY: build-backend
build-backend: ## Build backend
	@echo "Building backend..."
	@cd backend && go build -o bin/server cmd/server/main.go

.PHONY: migrate-create
migrate-create: ## Create a new migration
	@echo "Enter migration name:"
	@read name; \
	cd backend && migrate create -ext sql -dir internal/db/migrations $$name

.PHONY: build-frontend
build-frontend: ## Build frontend
	@echo "Building frontend..."
	@cd frontend && pnpm run build

.PHONY: openapi-bundle
openapi-bundle: ## Bundle OpenAPI specification from split files
	@echo "Bundling OpenAPI specification..."
	@pnpm install @redocly/cli || echo "Installing Redocly CLI..."
	@redocly bundle api/openapi.json --output api/generated/openapi.json
	@echo "OpenAPI specification bundled to api/generated/openapi.json"

.PHONY: openapi-validate
openapi-validate: ## Validate OpenAPI specification
	@echo "Validating OpenAPI specification..."
	@redocly lint api/openapi.json

.PHONY: openapi-preview
openapi-preview: ## Preview OpenAPI specification in browser
	@echo "Starting OpenAPI documentation preview..."
	@redocly preview-docs api/openapi.json

.PHONY: openapi-info
openapi-info: ## Show information about the OpenAPI specification
	@echo "OpenAPI specification information:"
	@echo "Main file: api/openapi.json"
	@echo "Split files:"
	@find api/ -name "*.json" | grep -v generated | wc -l | xargs -I {} echo "  Total JSON files (excluding generated): {}"
	@echo "Path files: $$(find api/paths -name '*.json' | wc -l)"
	@echo "Schema files: $$(find api/components/schemas -name '*.json' | wc -l)"
	@echo "Response files: $$(find api/components/responses -name '*.json' | wc -l)"

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf backend/bin
	@cd frontend && rm -rf .next
	@cd frontend && rm -rf out
	@rm -rf dist
	@rm -rf api/generated
