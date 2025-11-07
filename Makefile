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

all: format lint vet test

ci: format-check lint vet coverage-check ## 🤖 CI pipeline

.DEFAULT_GOAL := help
