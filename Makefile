# Traffic Control Go - Library Makefile

.PHONY: help test clean dev
.DEFAULT_GOAL := help

# Variables
VERSION := $(shell git describe --tags --always 2>/dev/null || echo "dev")

help: ## Show available commands
	@echo "Traffic Control Go Library - Available Commands:"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "Examples:"
	@echo "  make test                # Run tests"
	@echo "  make test-coverage       # Run tests with coverage"
	@echo "  make lint                # Run linting"

# Core development commands
test: ## Run tests
	@go test -v ./...

test-coverage: ## Run tests with coverage
	@go test -v -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report generated: coverage.html"

test-integration: ## Run integration tests (requires root and iperf3)
	@echo "Running integration tests (requires root privileges and iperf3)..."
	@sudo go test -v -tags=integration ./test/integration/...

clean: ## Clean build artifacts
	@rm -rf dist coverage.out coverage.html
	@echo "✓ Cleaned"

# Development helpers
dev: ## Set up development environment
	@go mod download
	@go mod tidy
	@echo "✓ Development setup completed"

fmt: ## Format code
	@go fmt ./...

lint: ## Run linting
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		go vet ./...; \
		echo "✓ Basic linting completed (install golangci-lint for comprehensive checks)"; \
	fi

security: ## Run security scanner
	@if command -v gosec >/dev/null 2>&1; then \
		gosec -quiet -fmt json ./... || true; \
	else \
		echo "gosec not installed, skipping security scan"; \
	fi

check: fmt lint test ## Run all quality checks

# Version management
version: ## Show current version
	@echo "Current version: $(VERSION)"

# Documentation
docs: ## Generate documentation
	@echo "Generating documentation..."
	@go doc -all > docs/api-reference.txt
	@echo "✓ Documentation generated"

# Examples
examples: ## Build and test examples
	@echo "Testing examples..."
	@go run examples/basic_demo.go
	@go run examples/priority_demo.go
	@go run examples/qdisc_types_demo.go
	@go run examples/statistics_demo.go
	@go run examples/improved_api_demo.go
	@echo "✓ All examples tested"

# Quick info
info: ## Show project info
	@echo "Project: Traffic Control Go Library"
	@echo "Version: $(VERSION)"
	@echo "Go version: $(shell go version)"
	@echo "Type: Go Library (import github.com/RNG999/traffic-control-go)"