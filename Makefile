# Traffic Control Go - Simplified Makefile

.PHONY: help build test clean install dev release
.DEFAULT_GOAL := help

# Variables
MAIN_BINARY := traffic-control
DEMO_BINARY := tcctl
BIN_DIR := bin
VERSION := $(shell grep -o 'version = "[^"]*"' cmd/traffic-control/main.go | cut -d'"' -f2 2>/dev/null || echo "dev")
BUILD_DATE := $(shell TZ=Asia/Tokyo date '+%Y%m%d%H%M')
BUILD_TIME := $(shell TZ=Asia/Tokyo date '+%Y-%m-%d_%H:%M_JST')

# Build flags
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.commit=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown") -X main.buildDate=$(BUILD_TIME)

help: ## Show available commands
	@echo "Traffic Control Go - Available Commands:"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@echo ""
	@echo "Examples:"
	@echo "  make build               # Build both binaries"
	@echo "  make test                # Run tests"
	@echo "  make install             # Install to system"
	@echo "  make release-auto VERSION=0.2.0  # Automated release (v202412021430)"
	@echo "  make release-with-date VERSION=0.2.0  # Manual release with date"

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

test-integration: ## Run integration tests (requires root and iperf3)
	@echo "Running integration tests (requires root privileges and iperf3)..."
	@sudo go test -v -tags=integration ./test/integration/...

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

security: ## Run security scanner
	@if command -v gosec >/dev/null 2>&1; then \
		gosec -quiet -fmt json ./... || true; \
	else \
		echo "gosec not installed, skipping security scan"; \
	fi

check: fmt lint test ## Run all quality checks

# Version management (simplified)
version: ## Show current version
	@echo "Current version: $(VERSION)"

# Release management (enhanced with date support)
release-simple: ## Simple release (manual)
ifndef VERSION
	@echo "Usage: make release-simple VERSION=0.2.0"
	@exit 1
endif
	@echo "Creating simple release $(VERSION)..."
	@$(MAKE) clean test build
	@echo "✓ Release $(VERSION) ready in $(BIN_DIR)/"

release-with-date: ## Create release with date tag
ifndef VERSION
	@echo "Usage: make release-with-date VERSION=0.2.0"
	@exit 1
endif
	@echo "Creating release $(VERSION) with date $(BUILD_DATE)..."
	@$(MAKE) clean test build
	@echo "Version: $(VERSION)"
	@echo "Build Date: $(BUILD_TIME)"
	@echo "Commit: $(shell git rev-parse --short HEAD)"
	@echo ""
	@echo "To create a git tag with date:"
	@echo "  git tag -a v$(BUILD_DATE) -m 'Release v$(VERSION) built on $(BUILD_TIME)'"
	@echo "  git push origin v$(BUILD_DATE)"
	@echo ""
	@echo "✓ Release v$(BUILD_DATE) (version $(VERSION)) ready in $(BIN_DIR)/"

release-auto: ## Trigger automated GitHub release
ifndef VERSION
	@echo "Usage: make release-auto VERSION=0.2.0 [PRE_RELEASE=beta]"
	@echo ""
	@echo "Examples:"
	@echo "  make release-auto VERSION=0.2.0           # Creates v202412021430 (v0.2.0)"
	@echo "  make release-auto VERSION=1.0.0           # Creates v202412021430 (v1.0.0)"
	@echo "  make release-auto VERSION=0.2.0 PRE_RELEASE=beta  # Creates v202412021430 (v0.2.0-beta)"
	@exit 1
endif
	@echo "Triggering automated GitHub release for version $(VERSION)..."
	@if command -v gh >/dev/null 2>&1; then \
		if [ -n "$(PRE_RELEASE)" ]; then \
			gh workflow run auto-release.yml -f custom_version=$(VERSION) -f pre_release=$(PRE_RELEASE); \
		else \
			gh workflow run auto-release.yml -f custom_version=$(VERSION); \
		fi; \
		echo "✓ GitHub Actions workflow triggered"; \
		echo "Check progress at: https://github.com/$(shell git config --get remote.origin.url | sed 's|.*github.com[:/]||' | sed 's|\.git||')/actions"; \
	else \
		echo "GitHub CLI (gh) not found. Please install it or use GitHub web interface."; \
		echo "Manually trigger workflow at: https://github.com/$(shell git config --get remote.origin.url | sed 's|.*github.com[:/]||' | sed 's|\.git||')/actions/workflows/auto-release.yml"; \
	fi

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