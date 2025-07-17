# Anubis Task Executor Makefile

.PHONY: help build test test-unit test-integration test-coverage clean demo lint fmt vet

# Default target
help:
	@echo "Available targets:"
	@echo "  build            - Build the binary"
	@echo "  test             - Run all unit tests"
	@echo "  test-unit        - Run unit tests only"
	@echo "  test-integration - Run integration tests (requires INTEGRATION_TESTS=1)"
	@echo "  test-coverage    - Run tests with coverage report"
	@echo "  demo             - Run demo with test cases"
	@echo "  clean            - Clean build artifacts"
	@echo "  lint             - Run golangci-lint"
	@echo "  fmt              - Format code"
	@echo "  vet              - Run go vet"
	@echo "  deps             - Download dependencies"

# Build the binary
build: bin
	@echo "Building anubis-executer..."
	go build -o bin/anubis-executer main.go

# Run all unit tests
test: test-unit

# Run unit tests only
test-unit:
	@echo "Running unit tests..."
	go test ./executer -v

# Run integration tests (requires real API access)
test-integration:
	@echo "Running integration tests..."
	INTEGRATION_TESTS=1 go test ./executer -v -run TestIntegration

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test ./executer -v -cover -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run demo
demo:
	@echo "Running demo..."
	go run main.go demo

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f bin/anubis-executer
	rm -f coverage.out coverage.html

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Run golangci-lint (if available)
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, skipping..."; \
	fi

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Development workflow
dev: fmt vet test-unit

# CI workflow
ci: fmt vet test-unit test-coverage

# Create bin directory
bin:
	mkdir -p bin
