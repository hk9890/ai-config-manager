.PHONY: build test integration-test test-integration install clean help

# Binary name
BINARY=aimgr

# Version information
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS=-ldflags "-s -w -X github.com/hk9890/ai-config-manager/pkg/version.Version=$(VERSION) \
	-X github.com/hk9890/ai-config-manager/pkg/version.GitCommit=$(GIT_COMMIT) \
	-X github.com/hk9890/ai-config-manager/pkg/version.BuildDate=$(BUILD_DATE)"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet
GOMOD=$(GOCMD) mod
INSTALL_PATH=~/bin

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

build: ## Build the binary
	$(GOBUILD) $(LDFLAGS) -o $(BINARY) -v

test: unit-test integration-test vet ## Run all tests

unit-test: ## Run only unit tests (fast, no external dependencies)
	@echo "Running unit tests..."
	$(GOTEST) -v -short ./pkg/...

integration-test: ## Run only integration tests (slower, requires git/network)
	@echo "Running integration tests..."
	$(GOTEST) -v -tags=integration ./pkg/...
	$(GOTEST) -v ./test/...

test-integration: ## Run integration tests (requires network, uses real GitHub repos)
	@echo "Running integration tests with real Git operations..."
	$(GOTEST) -v -tags=integration ./test/...

install: build ## Install binary to ~/bin
	mkdir -p $(INSTALL_PATH)
	cp $(BINARY) $(INSTALL_PATH)/
	@echo "Installed $(BINARY) to $(INSTALL_PATH)"
	@echo "Make sure $(INSTALL_PATH) is in your PATH"

clean: ## Clean build artifacts
	rm -f $(BINARY)
	rm -rf bin/ dist/
	$(GOCMD) clean

deps: ## Download dependencies
	$(GOMOD) download
	$(GOMOD) tidy

fmt: ## Format Go code
	$(GOCMD) fmt ./...

vet: ## Run go vet
	$(GOVET) ./...

all: clean deps fmt vet test build ## Run all checks and build
