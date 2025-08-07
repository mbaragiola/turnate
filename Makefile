# Turnate Makefile

.PHONY: help build run test clean dev install deps fmt lint docs docker

# Default target
.DEFAULT_GOAL := help

# Variables
BINARY_NAME := turnate
BUILD_DIR := bin
DOCKER_IMAGE := turnate:latest
GO_PATH := $(HOME)/go/bin
GO_CMD := $(shell which go 2>/dev/null || echo "$(GO_PATH)/go")
GO_VERSION := $(shell $(GO_CMD) version 2>/dev/null | cut -d' ' -f3 || echo "not-found")

## Help
help: ## Show this help message
	@echo "Turnate Development Commands"
	@echo "=========================="
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

## Build
build: ## Build the application binary
	@echo "🔨 Building Turnate..."
	@mkdir -p $(BUILD_DIR)
	@if [ "$(GO_VERSION)" = "not-found" ]; then \
		echo "❌ Go not found. Please install Go or add it to PATH"; \
		echo "   You can also set PATH=$(GO_PATH):$$PATH"; \
		exit 1; \
	fi
	@PATH="$(GO_PATH):$$PATH" CGO_ENABLED=1 $(GO_CMD) build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/turnate
	@echo "✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

build-static: ## Build static binary for distribution
	@echo "🔨 Building static binary..."
	@mkdir -p $(BUILD_DIR)
	@PATH="$(GO_PATH):$$PATH" CGO_ENABLED=1 GOOS=linux $(GO_CMD) build -a -ldflags '-extldflags "-static"' -o $(BUILD_DIR)/$(BINARY_NAME)-static ./cmd/turnate
	@echo "✅ Static build complete: $(BUILD_DIR)/$(BINARY_NAME)-static"

## Development
dev: ## Run in development mode with auto-reload (requires air)
	@echo "🚀 Starting development server..."
	@which air > /dev/null || (echo "Installing air..." && go install github.com/air-verse/air@latest)
	@air

run: build ## Build and run the application
	@echo "🚀 Starting Turnate..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

## Dependencies
deps: ## Download and install dependencies
	@echo "📦 Installing dependencies..."
	@PATH="$(GO_PATH):$$PATH" $(GO_CMD) mod download
	@PATH="$(GO_PATH):$$PATH" $(GO_CMD) mod tidy
	@echo "✅ Dependencies installed"

install-tools: ## Install development tools
	@echo "🛠️  Installing development tools..."
	@go install github.com/air-verse/air@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "✅ Development tools installed"

## Testing
test: ## Run unit tests
	@echo "🧪 Running tests..."
	@PATH="$(GO_PATH):$$PATH" $(GO_CMD) test ./tests/unit/... -v

test-coverage: ## Run tests with coverage
	@echo "🧪 Running tests with coverage..."
	@go test ./tests/unit/... -coverprofile=coverage.out
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated: coverage.html"

test-race: ## Run tests with race detection
	@echo "🧪 Running tests with race detection..."
	@go test -race ./tests/unit/...

benchmark: ## Run benchmarks
	@echo "⚡ Running benchmarks..."
	@go test -bench=. -benchmem ./...

## Code Quality
fmt: ## Format Go code
	@echo "🎨 Formatting code..."
	@gofmt -s -w .
	@goimports -w .
	@echo "✅ Code formatted"

lint: ## Run linter
	@echo "🔍 Running linter..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	@golangci-lint run
	@echo "✅ Linting complete"

vet: ## Run go vet
	@echo "🔍 Running go vet..."
	@go vet ./...
	@echo "✅ Vet complete"

check: fmt lint vet test ## Run all code quality checks

## Documentation
docs: ## Generate documentation
	@echo "📚 Generating documentation..."
	@go doc -all ./... > docs/GODOC.md
	@echo "✅ Documentation generated"

## Docker
docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	@docker build -t $(DOCKER_IMAGE) .
	@echo "✅ Docker image built: $(DOCKER_IMAGE)"

docker-run: docker-build ## Build and run Docker container
	@echo "🐳 Running Docker container..."
	@docker run -p 8080:8080 --rm $(DOCKER_IMAGE)

docker-compose-up: ## Start with docker-compose
	@echo "🐳 Starting with docker-compose..."
	@docker-compose up -d
	@echo "✅ Application started with docker-compose"

docker-compose-down: ## Stop docker-compose
	@echo "🐳 Stopping docker-compose..."
	@docker-compose down
	@echo "✅ Application stopped"

## Database
db-reset: ## Reset database (WARNING: Deletes all data)
	@echo "⚠️  Resetting database..."
	@read -p "Are you sure? This will delete all data (y/N): " confirm && [ "$$confirm" = "y" ]
	@rm -f turnate.db
	@echo "✅ Database reset"

db-backup: ## Backup database
	@echo "💾 Backing up database..."
	@mkdir -p backups
	@cp turnate.db backups/turnate-backup-$(shell date +%Y%m%d_%H%M%S).db
	@echo "✅ Database backed up to backups/"

## Deployment
deploy-build: ## Build for deployment
	@echo "🚀 Building for deployment..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -a -ldflags '-extldflags "-static" -s -w' -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/turnate
	@echo "✅ Deployment build complete: $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64"

release: clean check build-static ## Create release build
	@echo "📦 Creating release..."
	@mkdir -p release
	@cp $(BUILD_DIR)/$(BINARY_NAME)-static release/$(BINARY_NAME)
	@cp -r web release/
	@cp -r docs release/
	@cp README.md release/
	@cp LICENSE release/
	@tar -czf release/turnate-$(shell date +%Y%m%d).tar.gz -C release .
	@echo "✅ Release created: release/turnate-$(shell date +%Y%m%d).tar.gz"

## Cleanup
clean: ## Clean build artifacts
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -rf release
	@rm -f coverage.out coverage.html
	@echo "✅ Clean complete"

clean-all: clean ## Clean everything including dependencies
	@echo "🧹 Cleaning everything..."
	@go clean -modcache
	@docker system prune -f
	@echo "✅ Deep clean complete"

## Information
info: ## Show project information
	@echo "Turnate Project Information"
	@echo "=========================="
	@echo "Go Version: $(GO_VERSION)"
	@echo "Binary Name: $(BINARY_NAME)"
	@echo "Build Directory: $(BUILD_DIR)"
	@echo "Docker Image: $(DOCKER_IMAGE)"
	@echo ""
	@echo "Project Structure:"
	@tree -I 'bin|node_modules|.git|tmp' -L 2

status: ## Show git and build status
	@echo "Git Status:"
	@git status --short
	@echo ""
	@echo "Recent Commits:"
	@git log --oneline -5
	@echo ""
	@echo "Build Status:"
	@if [ -f $(BUILD_DIR)/$(BINARY_NAME) ]; then \
		echo "✅ Binary exists: $(BUILD_DIR)/$(BINARY_NAME)"; \
		echo "   Built: $(shell stat -c %y $(BUILD_DIR)/$(BINARY_NAME) 2>/dev/null || echo 'unknown')"; \
	else \
		echo "❌ Binary not found. Run 'make build'"; \
	fi

## Quick Development Workflow
quick-start: deps build run ## Full setup and run (first time)

quick-test: fmt test ## Quick format and test

quick-deploy: check deploy-build ## Quick quality check and deploy build