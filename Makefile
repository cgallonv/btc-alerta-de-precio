# Variables
APP_NAME := btc-price-alert
DOCKER_IMAGE := $(APP_NAME)
DOCKER_TAG := latest
PORT := 8080

# Default target
.DEFAULT_GOAL := help

## Help
help: ## Show this help message
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

## Development
dev: ## Run the application in development mode
	@echo "Starting development server..."
	go run main.go

build: ## Build the application
	@echo "Building application..."
	go build -o $(APP_NAME) main.go

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -f $(APP_NAME)
	rm -f *.db
	go clean

test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

test-cover: ## Run tests with coverage
	@echo "Running tests with coverage..."
	go test -v -cover ./...
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...

lint: ## Lint code
	@echo "Linting code..."
	golangci-lint run

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

## Docker
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run: docker-build ## Run Docker container
	@echo "Starting Docker container..."
	docker run -d \
		--name $(APP_NAME) \
		-p $(PORT):$(PORT) \
		-v $(PWD)/data:/app/data \
		$(DOCKER_IMAGE):$(DOCKER_TAG)

docker-stop: ## Stop Docker container
	@echo "Stopping Docker container..."
	docker stop $(APP_NAME) || true
	docker rm $(APP_NAME) || true

docker-logs: ## Show Docker container logs
	docker logs -f $(APP_NAME)

## Docker Compose
compose-up: ## Start services with docker-compose
	@echo "Starting services with docker-compose..."
	docker-compose up -d

compose-down: ## Stop services with docker-compose
	@echo "Stopping services with docker-compose..."
	docker-compose down

compose-logs: ## Show docker-compose logs
	docker-compose logs -f

compose-build: ## Build services with docker-compose
	docker-compose build

## Database
db-reset: ## Reset database (delete and recreate)
	@echo "Resetting database..."
	rm -f alerts.db
	@echo "Database reset. It will be recreated on next run."

db-backup: ## Backup database
	@echo "Backing up database..."
	cp alerts.db alerts_backup_$(shell date +%Y%m%d_%H%M%S).db

## Deployment
deploy-build: ## Build for deployment
	@echo "Building for deployment..."
	CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o $(APP_NAME) main.go

## Setup
setup: deps ## Setup development environment
	@echo "Setting up development environment..."
	@if [ ! -f .env ]; then \
		echo "Creating .env file from .env.example..."; \
		cp .env.example .env; \
		echo "Please edit .env file with your configuration"; \
	fi
	@echo "Setup complete!"

## API Testing
test-api: ## Test API endpoints
	@echo "Testing API endpoints..."
	@echo "Health check:"
	curl -s http://localhost:$(PORT)/api/v1/health | jq .
	@echo "Current price:"
	curl -s http://localhost:$(PORT)/api/v1/price | jq .
	@echo "Stats:"
	curl -s http://localhost:$(PORT)/api/v1/stats | jq .

## Performance
benchmark: ## Run benchmarks
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

## Security
security-scan: ## Run security scan
	@echo "Running security scan..."
	gosec ./...

## Monitoring
monitor: ## Monitor application (requires running app)
	@echo "Monitoring application..."
	@while true; do \
		echo "=== $(shell date) ==="; \
		curl -s http://localhost:$(PORT)/api/v1/health | jq .; \
		echo ""; \
		sleep 30; \
	done

.PHONY: help dev build clean test test-cover fmt lint deps docker-build docker-run docker-stop docker-logs compose-up compose-down compose-logs compose-build db-reset db-backup deploy-build setup test-api benchmark security-scan monitor 