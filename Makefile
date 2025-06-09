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

# Performance benchmarks
bench: ## Run benchmarks
	@echo "Running performance benchmarks..."
	@go test -bench=. -benchmem ./...

bench-compare: ## Run benchmarks with comparison (5 iterations)
	@echo "Running benchmark comparison (5 iterations)..."
	@go test -bench=. -benchmem -count=5 ./...

bench-profile: ## Run benchmarks with CPU profiling
	@echo "Running benchmarks with CPU profiling..."
	@go test -bench=. -cpuprofile=cpu.prof -memprofile=mem.prof ./...
	@echo "✓ Profiles generated: cpu.prof, mem.prof"
	@echo "View with: go tool pprof cpu.prof"

bench-value-objects: ## Run benchmarks for value objects only
	@echo "Running value object benchmarks..."
	@go test -bench=. -benchmem ./pkg/tc/...

bench-eventstore: ## Run benchmarks for event store only
	@echo "Running event store benchmarks..."
	@go test -bench=. -benchmem ./internal/infrastructure/eventstore/...

bench-api: ## Run benchmarks for API layer only
	@echo "Running API layer benchmarks..."
	@go test -bench=. -benchmem ./api/...

bench-report: ## Generate benchmark report
	@echo "Generating benchmark report..."
	@go test -bench=. -benchmem ./... > benchmark-report.txt
	@echo "✓ Benchmark report saved to benchmark-report.txt"

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
	@cd examples && go build basic/main.go
	@cd examples && go build production/main.go
	@cd examples && go build priority_demo.go
	@cd examples && go build filter_management_demo.go
	@cd examples && go build htb_advanced_demo.go
	@echo "✓ All examples built successfully"

# Quick info
info: ## Show project info
	@echo "Project: Traffic Control Go Library"
	@echo "Version: $(VERSION)"
	@echo "Go version: $(shell go version)"
	@echo "Type: Go Library (import github.com/RNG999/traffic-control-go)"