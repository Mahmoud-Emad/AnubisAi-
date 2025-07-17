# Anubis AI Core-Backend Makefile

.PHONY: help build test test-unit test-integration clean run dev deps swagger lint fmt vet docker-build docker-run

# Default target
help:
	@echo "Available targets:"
	@echo "  build            - Build the binary"
	@echo "  run              - Run the application"
	@echo "  dev              - Run in development mode with hot reload"
	@echo "  test             - Run all tests"
	@echo "  test-unit        - Run unit tests only"
	@echo "  test-integration - Run integration tests"
	@echo "  clean            - Clean build artifacts"
	@echo "  deps             - Download dependencies"
	@echo "  swagger          - Generate Swagger documentation"
	@echo "  lint             - Run linter"
	@echo "  fmt              - Format code"
	@echo "  vet              - Run go vet"
	@echo "  docker-build     - Build Docker image"
	@echo "  docker-run       - Run Docker container"

# Build the binary
build:
	@echo "Building anubis-backend..."
	go build -o bin/anubis-backend main.go

# Run the application
run: build
	@echo "Starting Anubis AI Core-Backend..."
	./bin/anubis-backend

# Run in development mode
dev:
	@echo "Starting development server..."
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "Air not found. Install with: go install github.com/cosmtrek/air@latest"; \
		go run main.go; \
	fi

# Run all tests
test: test-unit

# Run unit tests
test-unit:
	@echo "Running unit tests..."
	go test ./... -v

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	INTEGRATION_TESTS=1 go test ./... -v -tags=integration

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f bin/anubis-backend
	rm -f coverage.out coverage.html
	rm -rf data/

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Generate Swagger documentation
swagger:
	@echo "Generating Swagger documentation..."
	@if command -v swag >/dev/null 2>&1; then \
		swag init; \
	else \
		echo "Swag not found. Install with: go install github.com/swaggo/swag/cmd/swag@latest"; \
	fi

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Run linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, skipping..."; \
	fi

# Development workflow
dev-setup: deps swagger

# CI workflow
ci: fmt vet test-unit

# Create bin directory
bin:
	mkdir -p bin

# Build with bin directory
build: bin
	go build -o bin/anubis-backend main.go

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t anubis-backend:latest .

# Docker run
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 --env-file .env anubis-backend:latest

# Database operations
db-migrate:
	@echo "Running database migrations..."
	go run main.go migrate

db-seed:
	@echo "Seeding database..."
	go run main.go seed

db-reset: clean
	@echo "Resetting database..."
	rm -rf data/
	$(MAKE) db-migrate db-seed
