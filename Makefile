# Traffic Control Go - Simplified Makefile

.PHONY: help build test clean install dev release
.DEFAULT_GOAL := help

# Variables
MAIN_BINARY := traffic-control
DEMO_BINARY := tcctl
BIN_DIR := bin
VERSION := $(shell grep -o 'version = "[^"]*"' cmd/traffic-control/main.go | cut -d'"' -f2 2>/dev/null || echo "dev")

# Build flags
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.commit=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

help: ## Show available commands
	@echo "Traffic Control Go - Available Commands:"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "Examples:"
	@echo "  make build               # Build both binaries"
	@echo "  make test                # Run tests"
	@echo "  make install             # Install to system"
	@echo "  make release VERSION=0.2.0  # Create release"

# Core development commands
build: ## Build both binaries
	@mkdir -p $(BIN_DIR)
	@echo "Building $(MAIN_BINARY)..."
	@go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(MAIN_BINARY) ./cmd/traffic-control
	@echo "Building $(DEMO_BINARY)..."
	@go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(DEMO_BINARY) ./cmd/tcctl
	@echo "✓ Build completed"

test: ## Run tests
	@go test -v ./...

clean: ## Clean build artifacts
	@rm -rf $(BIN_DIR) dist release coverage.out
	@echo "✓ Cleaned"

install: build ## Install binaries to system
	@echo "Installing to /usr/local/bin..."
	@sudo cp $(BIN_DIR)/$(MAIN_BINARY) /usr/local/bin/
	@sudo cp $(BIN_DIR)/$(DEMO_BINARY) /usr/local/bin/
	@echo "✓ Installed"

# Development helpers
dev: ## Set up development environment
	@go mod download
	@go mod tidy
	@echo "✓ Development setup completed"

fmt: ## Format code
	@go fmt ./...

lint: ## Run basic linting
	@go vet ./...
	@echo "✓ Linting completed"

check: fmt lint test ## Run all quality checks

# Version management (simplified)
version: ## Show current version
	@echo "Current version: $(VERSION)"

# Release management (choose one approach)
release-simple: ## Simple release (manual)
ifndef VERSION
	@echo "Usage: make release-simple VERSION=0.2.0"
	@exit 1
endif
	@echo "Creating simple release $(VERSION)..."
	@$(MAKE) clean test build
	@echo "✓ Release $(VERSION) ready in $(BIN_DIR)/"

release-goreleaser: ## Release with GoReleaser (requires setup)
	@if ! command -v goreleaser >/dev/null 2>&1; then \
		echo "GoReleaser not found. Install with: go install github.com/goreleaser/goreleaser@latest"; \
		exit 1; \
	fi
	@goreleaser build --snapshot --clean
	@echo "✓ GoReleaser build completed"

# Advanced (optional - only if needed)
build-all: ## Build for multiple platforms
	@mkdir -p $(BIN_DIR)
	@echo "Building for multiple platforms..."
	@GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(MAIN_BINARY)-linux-amd64 ./cmd/traffic-control
	@GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(MAIN_BINARY)-darwin-amd64 ./cmd/traffic-control
	@GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(MAIN_BINARY)-windows-amd64.exe ./cmd/traffic-control
	@echo "✓ Multi-platform build completed"

docker: ## Build Docker image
	@docker build -t traffic-control-go .

# Quick info
info: ## Show project info
	@echo "Project: Traffic Control Go"
	@echo "Version: $(VERSION)"
	@echo "Binaries: $(MAIN_BINARY), $(DEMO_BINARY)"
	@echo "Go version: $(shell go version)"