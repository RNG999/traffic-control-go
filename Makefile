.PHONY: help build test clean lint security deps check-deps install-tools
.DEFAULT_GOAL := help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=tcctl
BINARY_DIR=bin

# Version and build info
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags="-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME) -s -w"

help: ## Show this help message
	@echo 'Usage: make <target>'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

deps: ## Download dependencies
	$(GOMOD) download
	$(GOMOD) verify

tidy: ## Tidy up dependencies
	$(GOMOD) tidy

build: deps ## Build the binary
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME) ./cmd/tcctl

build-all: deps ## Build for all platforms
	@mkdir -p $(BINARY_DIR)
	# Linux
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/tcctl
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/tcctl
	# macOS
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/tcctl
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/tcctl
	# Windows
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/tcctl

test: ## Run tests
	$(GOTEST) -race -coverprofile=coverage.out -covermode=atomic ./internal/domain/valueobjects/... ./internal/infrastructure/netlink/... ./pkg/... ./api/... ./test/...

test-unit: ## Run unit tests
	$(GOTEST) -v ./test/unit/...

test-integration: ## Run integration tests
	$(GOTEST) -v ./test/integration/...

test-examples: ## Run example tests
	$(GOTEST) -v ./test/examples/...

test-coverage: test ## Run tests and show coverage
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

install-tools: ## Install development tools
	@echo "Installing development tools..."
	$(GOGET) github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	$(GOGET) github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

lint: ## Run linter
	golangci-lint run

security: ## Run security scanner
	gosec ./...

vet: ## Run go vet
	$(GOCMD) vet ./...

fmt: ## Format code
	$(GOCMD) fmt ./...

check: fmt vet lint test ## Run all checks (format, vet, lint, test)

clean: ## Clean build artifacts
	$(GOCLEAN)
	rm -rf $(BINARY_DIR)
	rm -f coverage.out coverage.html

docker-build: ## Build Docker image
	docker build -t tcctl:latest .

docker-run: docker-build ## Run Docker container
	docker run --rm tcctl:latest --help

install: build ## Install binary to $GOPATH/bin
	cp $(BINARY_DIR)/$(BINARY_NAME) $(GOPATH)/bin/

uninstall: ## Remove binary from $GOPATH/bin
	rm -f $(GOPATH)/bin/$(BINARY_NAME)

# Development helpers
dev-setup: deps install-tools ## Set up development environment
	@echo "Development environment set up successfully!"

ci: check build-all ## Run CI pipeline locally
	@echo "CI pipeline completed successfully!"

release-check: ## Check if ready for release
	@echo "Checking release readiness..."
	@$(MAKE) test
	@$(MAKE) lint
	@$(MAKE) security
	@$(MAKE) build-all
	@echo "✓ All checks passed - ready for release!"