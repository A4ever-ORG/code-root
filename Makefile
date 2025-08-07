# Telegram Store Hub - Makefile

# Application name
APP_NAME = telegram-store-hub

# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get
GOMOD = $(GOCMD) mod

# Build directory
BUILD_DIR = build

# Binary names for different platforms
BINARY_UNIX = $(BUILD_DIR)/$(APP_NAME)_unix
BINARY_LINUX = $(BUILD_DIR)/$(APP_NAME)_linux
BINARY_WINDOWS = $(BUILD_DIR)/$(APP_NAME)_windows.exe
BINARY_MAC = $(BUILD_DIR)/$(APP_NAME)_mac
BINARY_ARM = $(BUILD_DIR)/$(APP_NAME)_arm

# Version info
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Linker flags
LDFLAGS = -ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)"

.PHONY: build clean test deps help run dev

# Default target
all: clean deps test build-all

# Build for current platform
build:
	@echo "Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) .

# Build for all platforms
build-all: build-linux build-windows build-mac build-arm

# Build for Linux
build-linux:
	@echo "Building for Linux..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_LINUX) .

# Build for Windows
build-windows:
	@echo "Building for Windows..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_WINDOWS) .

# Build for macOS
build-mac:
	@echo "Building for macOS..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_MAC) .

# Build for ARM (Raspberry Pi, Android Termux)
build-arm:
	@echo "Building for ARM..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=7 $(GOBUILD) $(LDFLAGS) -o $(BINARY_ARM) .

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

# Run in development mode
dev: deps
	@echo "Running in development mode..."
	$(GOCMD) run .

# Run the application
run: build
	@echo "Running $(APP_NAME)..."
	./$(BUILD_DIR)/$(APP_NAME)

# Install the application
install: build
	@echo "Installing $(APP_NAME)..."
	sudo cp $(BUILD_DIR)/$(APP_NAME) /usr/local/bin/

# Uninstall the application
uninstall:
	@echo "Uninstalling $(APP_NAME)..."
	sudo rm -f /usr/local/bin/$(APP_NAME)

# Create release package
release: clean build-all
	@echo "Creating release package..."
	@mkdir -p $(BUILD_DIR)/release
	@cp $(BINARY_LINUX) $(BUILD_DIR)/release/$(APP_NAME)_linux_amd64
	@cp $(BINARY_WINDOWS) $(BUILD_DIR)/release/$(APP_NAME)_windows_amd64.exe
	@cp $(BINARY_MAC) $(BUILD_DIR)/release/$(APP_NAME)_darwin_amd64
	@cp $(BINARY_ARM) $(BUILD_DIR)/release/$(APP_NAME)_linux_arm
	@cp .env.example $(BUILD_DIR)/release/
	@cp README.md $(BUILD_DIR)/release/
	@cp INSTALLATION.md $(BUILD_DIR)/release/
	@cp USAGE.md $(BUILD_DIR)/release/
	@echo "Release package created in $(BUILD_DIR)/release/"

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(APP_NAME):$(VERSION) .

# Docker run
docker-run:
	@echo "Running Docker container..."
	docker run --rm -it --env-file .env $(APP_NAME):$(VERSION)

# Format code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run

# Migrate database
migrate:
	@echo "Running database migrations..."
	$(GOCMD) run . --migrate

# Show help
help:
	@echo "Available targets:"
	@echo "  build        - Build for current platform"
	@echo "  build-all    - Build for all platforms"
	@echo "  build-linux  - Build for Linux"
	@echo "  build-windows - Build for Windows"
	@echo "  build-mac    - Build for macOS"
	@echo "  build-arm    - Build for ARM"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Install dependencies"
	@echo "  test         - Run tests"
	@echo "  dev          - Run in development mode"
	@echo "  run          - Build and run"
	@echo "  install      - Install to /usr/local/bin"
	@echo "  uninstall    - Remove from /usr/local/bin"
	@echo "  release      - Create release package"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run   - Run Docker container"
	@echo "  fmt          - Format code"
	@echo "  lint         - Lint code"
	@echo "  migrate      - Run database migrations"
	@echo "  help         - Show this help"

# Version info
version:
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"