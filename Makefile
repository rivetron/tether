.PHONY: help dev run test test-coverage lint format clean install setup

# Variables
BINARY_NAME=tether
BINARY_PATH=./bin/$(BINARY_NAME)

# Colors
GREEN=\033[0;32m
YELLOW=\033[0;33m
RED=\033[0;31m
NC=\033[0m

help: ## 📋 Show available commands
	@echo "🔗 Tether - Available Commands:"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install: ## 📦 Install dependencies
	@echo "$(YELLOW)📦 Installing dependencies...$(NC)"
	@go mod download && go mod tidy
	@echo "$(GREEN)✅ Dependencies installed$(NC)"

dev: ## 🚀 Run in development mode
	@echo "$(YELLOW)🚀 Running in development mode...$(NC)"
	@go run src/cmd/main.go

run: dev ## 🏃 Alias for dev

test: ## 🧪 Run tests
	@echo "$(YELLOW)🧪 Running tests...$(NC)"
	@go test ./... -v

test-coverage: ## 📊 Run tests with coverage
	@echo "$(YELLOW)📊 Running tests with coverage...$(NC)"
	@go test ./... -coverprofile=coverage.out -covermode=atomic
	@go tool cover -func=coverage.out | grep total | awk '{print "Coverage: " $$3}'
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)📊 Coverage report: coverage.html$(NC)"

test-watch: ## 👀 Watch tests (requires entr)
	@echo "$(YELLOW)👀 Watching for changes...$(NC)"
	@find . -name '*.go' | entr -d -c go test ./... -v

coverage-check: test-coverage ## ✅ Check 80% coverage threshold
	@echo "$(YELLOW)✅ Checking coverage threshold...$(NC)"
	@coverage=$$(go tool cover -func=coverage.out | grep total | awk '{print substr($$3, 1, length($$3)-1)}'); \
	echo "Current coverage: $${coverage}%"; \
	threshold=80; \
	if (( $$(echo "$${coverage} < $$threshold" | bc -l) )); then \
		echo "$(RED)❌ Coverage $${coverage}% below $$threshold%$(NC)"; \
		exit 1; \
	else \
		echo "$(GREEN)✅ Coverage threshold met: $${coverage}%$(NC)"; \
	fi

lint: ## 🔍 Run linter
	@echo "$(YELLOW)🔍 Running linter...$(NC)"
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)⚠️  golangci-lint not installed, using go vet$(NC)"; \
		go vet ./...; \
	fi
	@echo "$(GREEN)✅ Linting complete$(NC)"

format: ## 🎨 Format code
	@echo "$(YELLOW)🎨 Formatting code...$(NC)"
	@go fmt ./...
	@echo "$(GREEN)✅ Code formatted$(NC)"

format-check: ## 🎭 Check code formatting
	@echo "$(YELLOW)🎭 Checking formatting...$(NC)"
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "$(RED)❌ Unformatted files: $$unformatted$(NC)"; \
		exit 1; \
	else \
		echo "$(GREEN)✅ All files formatted$(NC)"; \
	fi

vet: ## 🔍 Run go vet
	@echo "$(YELLOW)�� Running go vet...$(NC)"
	@go vet ./...
	@echo "$(GREEN)✅ Go vet passed$(NC)"

clean: ## 🧹 Clean build artifacts
	@echo "$(YELLOW)🧹 Cleaning...$(NC)"
	@rm -rf bin/ coverage.out coverage.html
	@go clean -cache
	@echo "$(GREEN)✅ Cleaned$(NC)"

setup: ## 🛠️ Setup development environment
	@echo "$(YELLOW)🛠️  Setting up dev environment...$(NC)"
	@go mod tidy
	@echo "$(GREEN)✅ Setup complete$(NC)"

# Tether specific commands
docker-up: ## 🐳 Start Docker services (MongoDB + Redis)
	@echo "$(YELLOW)🐳 Starting Docker services...$(NC)"
	@docker-compose up -d mongodb redis
	@echo "$(GREEN)✅ Services started$(NC)"

docker-down: ## 🛑 Stop Docker services
	@echo "$(YELLOW)🛑 Stopping Docker services...$(NC)"
	@docker-compose down
	@echo "$(GREEN)✅ Services stopped$(NC)"

docker-logs: ## 📋 View Docker logs
	@docker-compose logs -f

air-install: ## 🌪️ Install Air for hot reload
	@echo "$(YELLOW)🌪️ Installing Air...$(NC)"
	@go install github.com/air-verse/air@latest
	@echo "$(GREEN)✅ Air installed$(NC)"

dev-air: air-install ## 🔥 Run with Air hot reload
	@echo "$(YELLOW)🔥 Starting with Air hot reload...$(NC)"
	@air

swagger: ## 📚 Generate Swagger docs
	@echo "$(YELLOW)📚 Generating Swagger docs...$(NC)"
	@go install github.com/swaggo/swag/cmd/swag@latest
	@swag init -g src/cmd/main.go -o docs/
	@echo "$(GREEN)✅ Swagger docs generated$(NC)"

swagger-validate: ## ✅ Validate Swagger docs
	@echo "$(YELLOW)✅ Validating Swagger docs...$(NC)"
	@if command -v swagger > /dev/null; then \
		swagger validate docs/swagger.yaml; \
	else \
		echo "$(YELLOW)⚠️  swagger-cli not installed, skipping validation$(NC)"; \
	fi

test-api: ## 🧪 Test API endpoints
	@echo "$(YELLOW)🧪 Testing API endpoints...$(NC)"
	@echo "Health check:"
	@curl -s http://localhost:5000/health | head -1 || echo "❌ Service not running"
	@echo ""
	@echo "Swagger docs:"
	@curl -s http://localhost:5000/docs/swagger.json | head -1 || echo "❌ Swagger not available"
	@echo "$(GREEN)✅ API test complete$(NC)"

all: format lint vet test

ci: format-check lint vet coverage-check ## 🤖 CI pipeline

.DEFAULT_GOAL := help
