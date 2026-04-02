.PHONY: build build-all clean test lint fmt deps run frontend-install frontend-build frontend-dev help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet
BINARY_NAME=ms
BINARY_DIR=bin
GO_FILES=$(shell find . -name '*.go' -type f)

# Version
VERSION=v0.1.0
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Build for current platform
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME) ./cmd/ms

# Build for all platforms
build-all: build-linux build-darwin build-windows

build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BINARY_DIR)/linux
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/linux/$(BINARY_NAME) ./cmd/ms

build-darwin:
	@echo "Building for macOS..."
	@mkdir -p $(BINARY_DIR)/darwin
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/darwin/$(BINARY_NAME) ./cmd/ms
	@mkdir -p $(BINARY_DIR)/darwin-arm64
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/darwin-arm64/$(BINARY_NAME) ./cmd/ms

build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BINARY_DIR)/windows
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/windows/$(BINARY_NAME).exe ./cmd/ms

# Clean
clean:
	@echo "Cleaning..."
	@rm -rf $(BINARY_DIR)

# Test
test:
	$(GOTEST) -v ./...

# Lint
lint:
	$(GOVET) ./...

# Format
fmt:
	$(GOFMT) $(GO_FILES)

# Install dependencies
deps:
	$(GOCMD) mod download
	$(GOCMD) mod tidy

# Run locally
run: build
	./$(BINARY_DIR)/$(BINARY_NAME) $(ARGS)

# Frontend
frontend-install:
	cd frontend && npm install

frontend-build:
	cd frontend && npm run build

frontend-dev:
	cd frontend && npm run dev

# Help
help:
	@echo "Available targets:"
	@echo "  build       - Build for current platform"
	@echo "  build-all   - Build for all platforms (linux, darwin, windows)"
	@echo "  clean       - Clean build artifacts"
	@echo "  test        - Run tests"
	@echo "  lint        - Run go vet"
	@echo "  fmt         - Format code"
	@echo "  deps        - Download and tidy dependencies"
	@echo "  run         - Build and run locally"
	@echo "  frontend-install - Install frontend dependencies"
	@echo "  frontend-build   - Build frontend"
	@echo "  frontend-dev     - Run frontend dev server"
	@echo "  help        - Show this help"
