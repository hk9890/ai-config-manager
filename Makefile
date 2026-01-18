.PHONY: build test integration-test install clean help

# Binary name
BINARY=ai-repo

# Version information
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS=-ldflags "-s -w -X github.com/hans-m-leitner/ai-config-manager/pkg/version.Version=$(VERSION) \
	-X github.com/hans-m-leitner/ai-config-manager/pkg/version.GitCommit=$(GIT_COMMIT) \
	-X github.com/hans-m-leitner/ai-config-manager/pkg/version.BuildDate=$(BUILD_DATE)"

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

test: ## Run all tests
	$(GOTEST) -v ./pkg/...
	$(GOTEST) -v ./test/...
	$(GOVET) ./...

unit-test: ## Run only unit tests
	$(GOTEST) -v ./pkg/...

integration-test: ## Run only integration tests
	$(GOTEST) -v ./test/...

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
