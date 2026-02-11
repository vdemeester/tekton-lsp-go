.PHONY: all build test clean install fmt lint

# Build variables
BINARY_NAME=tekton-lsp
VERSION?=0.1.0-dev
BUILD_DIR=build
INSTALL_DIR?=/usr/local/bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOINSTALL=$(GOCMD) install
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	@$(GOBUILD) -o $(BINARY_NAME) ./cmd/tekton-lsp
	@echo "Build complete: $(BINARY_NAME)"

build-release:
	@echo "Building release binary..."
	@CGO_ENABLED=1 $(GOBUILD) -ldflags="-s -w" -o $(BINARY_NAME) ./cmd/tekton-lsp
	@echo "Release build complete: $(BINARY_NAME)"

test:
	@echo "Running unit tests..."
	@$(GOTEST) -race -v ./pkg/...

test-integration:
	@echo "Running integration tests..."
	@$(GOTEST) -race -v -timeout 60s ./test/integration/

test-all:
	@echo "Running all tests..."
	@$(GOTEST) -race -v -timeout 60s ./...

test-coverage:
	@echo "Running tests with coverage..."
	@$(GOTEST) -race -coverprofile=coverage.out ./pkg/...
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

clean:
	@echo "Cleaning..."
	@$(GOCLEAN)
	@rm -f $(BINARY_NAME)
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

install: build
	@echo "Installing to $(INSTALL_DIR)..."
	@install -m 755 $(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@echo "Install complete"

fmt:
	@echo "Formatting code..."
	@$(GOFMT) -s -w .
	@echo "Format complete"

lint:
	@echo "Running linter..."
	@$(GOLINT) run ./...
	@echo "Lint complete"

tidy:
	@echo "Tidying go.mod..."
	@$(GOMOD) tidy
	@echo "Tidy complete"

version:
	@./$(BINARY_NAME) --version

run:
	@./$(BINARY_NAME)

# Development helpers
dev-deps:
	@echo "Installing development dependencies..."
	@$(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Dev dependencies installed"
