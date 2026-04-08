# Intelligent Test Orchestrator - Makefile

.PHONY: build test clean install run help deps

# Variables
BINARY_NAME=orchestrator
BUILD_DIR=bin
GO=go
GOFLAGS=-v

# Default target
all: deps build

# Help target
help:
	@echo "Intelligent Test Orchestrator - Available targets:"
	@echo "  make build      - Build the orchestrator binary"
	@echo "  make test       - Run all tests"
	@echo "  make clean      - Clean build artifacts"
	@echo "  make install    - Install the orchestrator"
	@echo "  make run        - Run the orchestrator"
	@echo "  make deps       - Download dependencies"
	@echo "  make coverage   - Run tests with coverage"
	@echo "  make lint       - Run linters"

# Download dependencies
deps:
	@echo "📦 Downloading dependencies..."
	$(GO) mod download
	$(GO) mod tidy

# Build the orchestrator
build: deps
	@echo "🔨 Building orchestrator..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for production (static binary)
build-prod:
	@echo "🔨 Building production binary..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=linux $(GO) build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "✅ Production build complete"

# Run tests
test:
	@echo "🧪 Running tests..."
	$(GO) test ./... -v

# Run tests with coverage
coverage:
	@echo "📊 Running tests with coverage..."
	$(GO) test ./... -coverprofile=coverage.out
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

# Run linters
lint:
	@echo "🔍 Running linters..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "⚠️  golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@rm -rf generated-tests
	@rm -f coverage-report.json
	@echo "✅ Clean complete"

# Install the orchestrator
install: build
	@echo "📦 Installing orchestrator..."
	$(GO) install .
	@echo "✅ Installation complete"

# Run the orchestrator
run: build
	@echo "🚀 Running orchestrator..."
	./$(BUILD_DIR)/$(BINARY_NAME) --help

# Run with example configuration
run-example: build
	@echo "🚀 Running orchestrator with example config..."
	./$(BUILD_DIR)/$(BINARY_NAME) \
		--base main \
		--head $$(git rev-parse --abbrev-ref HEAD) \
		--config config/test-orchestrator-config.yaml \
		--output generated-tests \
		--root .

# Format code
fmt:
	@echo "🎨 Formatting code..."
	$(GO) fmt ./...
	@echo "✅ Format complete"

# Vet code
vet:
	@echo "🔍 Vetting code..."
	$(GO) vet ./...
	@echo "✅ Vet complete"

# Check for security issues
security:
	@echo "🔒 Checking for security issues..."
	@if command -v gosec > /dev/null; then \
		gosec ./...; \
	else \
		echo "⚠️  gosec not installed. Install with: go install github.com/securego/gosec/v2/cmd/gosec@latest"; \
	fi

# Generate documentation
docs:
	@echo "📚 Generating documentation..."
	@if command -v godoc > /dev/null; then \
		echo "Starting godoc server at http://localhost:6060"; \
		godoc -http=:6060; \
	else \
		echo "⚠️  godoc not installed. Install with: go install golang.org/x/tools/cmd/godoc@latest"; \
	fi

# Docker build
docker-build:
	@echo "🐳 Building Docker image..."
	docker build -t test-orchestrator:latest .
	@echo "✅ Docker image built"

# Show project structure
tree:
	@echo "📁 Project structure:"
	@tree -I 'vendor|node_modules|.git' -L 3