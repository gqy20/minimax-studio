.PHONY: build build-release build-all clean test lint fmt deps run dev frontend-install frontend-build frontend-dev help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet
BINARY_NAME=ms
BINARY_DIR=bin
GO_FILES=$(shell find . -name '*.go' -type f)
EMBED_DIR=internal/api/frontend_dist

# Version
VERSION=v0.1.0
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Build for current platform (no frontend embed)
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME) ./cmd/ms

# Build with frontend embedded into binary (~20 MB single file)
build-release: frontend-build embed-frontend
	@echo "Building $(BINARY_NAME) with embedded frontend..."
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) $(LDFLAGS) -tags embed_frontend -o $(BINARY_DIR)/$(BINARY_NAME) ./cmd/ms
	@echo "Done: $(BINARY_DIR)/$(BINARY_NAME)"
	@ls -lh $(BINARY_DIR)/$(BINARY_NAME)

# Copy frontend dist into embed directory
embed-frontend:
	@echo "Embedding frontend..."
	@rm -rf $(EMBED_DIR)
	@cp -r frontend/dist $(EMBED_DIR)

# Build for all platforms (with embedded frontend)
build-all: build-linux build-darwin build-windows

build-linux: frontend-build embed-frontend
	@echo "Building for Linux..."
	@mkdir -p $(BINARY_DIR)/linux
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -tags embed_frontend -o $(BINARY_DIR)/linux/$(BINARY_NAME) ./cmd/ms

build-darwin: frontend-build embed-frontend
	@echo "Building for macOS..."
	@mkdir -p $(BINARY_DIR)/darwin
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -tags embed_frontend -o $(BINARY_DIR)/darwin/$(BINARY_NAME) ./cmd/ms
	@mkdir -p $(BINARY_DIR)/darwin-arm64
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -tags embed_frontend -o $(BINARY_DIR)/darwin-arm64/$(BINARY_NAME) ./cmd/ms

build-windows: frontend-build embed-frontend
	@echo "Building for Windows..."
	@mkdir -p $(BINARY_DIR)/windows
	GOOS=windows GOARCH=amd64 $(GOBUILD) -ldflags "-H windowsgui -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)" -tags embed_frontend -o $(BINARY_DIR)/windows/$(BINARY_NAME).exe ./cmd/ms

# Clean
clean:
	@echo "Cleaning..."
	@rm -rf $(BINARY_DIR)
	@rm -rf $(EMBED_DIR)

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

dev:
	bash scripts/dev.sh

# Frontend
frontend-install:
	cd frontend && npm install

frontend-build:
	cd frontend && npm run build

frontend-dev:
	cd frontend && npm run dev

frontend-dev-lan:
	cd frontend && FRONTEND_HOST=0.0.0.0 npm run dev

# Help
help:
	@echo "Available targets:"
	@echo "  build           - Build for current platform (no frontend)"
	@echo "  build-release   - Build with embedded frontend (single binary)"
	@echo "  build-all       - Build for all platforms with embedded frontend"
	@echo "  clean           - Clean build artifacts"
	@echo "  test            - Run tests"
	@echo "  lint            - Run go vet"
	@echo "  fmt             - Format code"
	@echo "  deps            - Download and tidy dependencies"
	@echo "  run             - Build and run locally"
	@echo "  dev             - Start backend and frontend together"
	@echo "  frontend-install - Install frontend dependencies"
	@echo "  frontend-build   - Build frontend"
	@echo "  frontend-dev     - Run frontend dev server"
	@echo "  frontend-dev-lan - Run frontend dev server for LAN access"
	@echo "  help             - Show this help"
