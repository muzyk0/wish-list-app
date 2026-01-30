# Makefile for Wish List Application

.PHONY: help
help: ## Show this help message
	@echo "Wish List Application - Development Commands"
	@echo ""
	@grep -E '^[a-zA-Z_0-9%-]+:.*?## .*$$' $(word 1,$(MAKEFILE_LIST)) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "📚 Documentation Commands:"
	@echo "  \033[36mdocs\033[0m                         Generate complete API docs (Swagger 2.0 → OpenAPI 3.0 → Split)"
	@echo "  \033[36mswagger-generate\033[0m            Generate OpenAPI docs from Go annotations"
	@echo "  \033[36mswagger-convert-v3\033[0m          Convert Swagger 2.0 to OpenAPI 3.0"
	@echo "  \033[36mswagger-split\033[0m               Split OpenAPI 3.0 into organized files"
	@echo "  \033[36mswagger-preview\033[0m             Preview OpenAPI specification in browser"

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

.PHONY: docs
docs: ## Generate complete API documentation (Swagger 2.0 → OpenAPI 3.0 → Split)
	@echo "================================================"
	@echo "📚 Generating Complete API Documentation"
	@echo "================================================"
	@echo ""
	@echo "Step 1/3: Generating Swagger 2.0 from Go annotations..."
	@if ! command -v swag >/dev/null 2>&1; then \
		echo "Installing swag..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	@$(shell go env GOPATH)/bin/swag init -g cmd/server/main.go -d backend -o backend/docs --parseDependency --parseInternal
	@echo "✓ Swagger 2.0 generated"
	@echo ""
	@echo "Step 2/3: Converting to OpenAPI 3.0..."
	@if ! command -v swagger2openapi >/dev/null 2>&1; then \
		echo "Installing swagger2openapi..."; \
		npm install -g swagger2openapi; \
	fi
	@swagger2openapi backend/docs/swagger.json -o backend/docs/openapi3.json
	@swagger2openapi backend/docs/swagger.yaml -o backend/docs/openapi3.yaml
	@echo "✓ OpenAPI 3.0 converted"
	@echo ""
	@echo "Step 3/3: Splitting into organized files..."
	@if ! command -v redocly >/dev/null 2>&1; then \
		echo "Installing Redocly CLI..."; \
		npm install -g @redocly/cli; \
	fi
	@mkdir -p backend/docs/split
	@redocly split backend/docs/openapi3.yaml --outDir=backend/docs/split
	@echo "✓ Split files created"
	@echo ""
	@echo "================================================"
	@echo "✅ Documentation Complete!"
	@echo "================================================"
	@echo ""
	@echo "📁 Generated files:"
	@echo "  • backend/docs/swagger.{json,yaml}    (Swagger 2.0)"
	@echo "  • backend/docs/openapi3.{json,yaml}   (OpenAPI 3.0)"
	@echo "  • backend/docs/split/                 (Split files)"
	@echo ""
	@echo "🌐 View documentation:"
	@echo "  • Swagger UI: http://localhost:8080/swagger/index.html"
	@echo "  • Preview:    make swagger-preview"
	@echo ""

.PHONY: swagger-generate
swagger-generate: ## Generate OpenAPI 3.0 documentation from Go code annotations
	@echo "Generating OpenAPI 3.0 documentation..."
	@if ! command -v swag >/dev/null 2>&1; then \
		echo "Installing swag..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	@$(shell go env GOPATH)/bin/swag init -g cmd/server/main.go -d backend -o backend/docs --parseDependency --parseInternal
	@echo "Converting to OpenAPI 3.0..."
	@$(MAKE) swagger-convert-v3
	@echo "OpenAPI 3.0 documentation generated at backend/docs/openapi3.{json,yaml}"
	@echo "Swagger UI available at http://localhost:8080/swagger/index.html"

.PHONY: swagger-convert-v3
swagger-convert-v3: ## Convert OpenAPI 2.0 to 3.0
	@echo "Converting OpenAPI 2.0 to 3.0..."
	@if ! command -v swagger2openapi >/dev/null 2>&1; then \
		echo "Installing swagger2openapi..."; \
		npm install -g swagger2openapi; \
	fi
	@swagger2openapi backend/docs/swagger.json -o backend/docs/openapi3.json
	@swagger2openapi backend/docs/swagger.yaml -o backend/docs/openapi3.yaml
	@echo "✓ OpenAPI 3.0 files generated: backend/docs/openapi3.{json,yaml}"

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

.PHONY: swagger-split
swagger-split: ## Split generated OpenAPI 3.0 spec into organized files
	@echo "Splitting OpenAPI 3.0 specification..."
	@if ! command -v redocly >/dev/null 2>&1; then \
		echo "Installing Redocly CLI..."; \
		pnpm install -g @redocly/cli; \
	fi
	@if [ ! -f backend/docs/openapi3.yaml ]; then \
		echo "OpenAPI 3.0 not found. Generating..."; \
		$(MAKE) swagger-convert-v3; \
	fi
	@mkdir -p backend/docs/split
	@redocly split backend/docs/openapi3.yaml --outDir=backend/docs/split
	@echo "✓ OpenAPI 3.0 specification split into backend/docs/split/"

.PHONY: swagger-preview
swagger-preview: ## Preview OpenAPI 3.0 specification in browser
	@echo "Starting OpenAPI 3.0 documentation preview..."
	@if ! command -v redocly >/dev/null 2>&1; then \
		echo "Installing Redocly CLI..."; \
		pnpm install -g @redocly/cli; \
	fi
	@if [ ! -f backend/docs/openapi3.yaml ]; then \
		echo "OpenAPI 3.0 not found. Generating..."; \
		$(MAKE) swagger-convert-v3; \
	fi
	@redocly preview-docs backend/docs/openapi3.yaml

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf backend/bin
	@cd frontend && rm -rf .next
	@cd frontend && rm -rf out
	@rm -rf dist
	@rm -rf api/generated
