# Build Scan Go - Makefile

# Variables
BINARY_NAME=cleansource-sca-cli
GO_VERSION=1.21
BUILD_DIR=build
MAIN_FILE=main.go

# Build information
VERSION?=4.0.0
BUILD_TIME=$(shell date +%Y-%m-%d_%H:%M:%S)
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Linker flags
LDFLAGS=-ldflags="-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Default target
.PHONY: all
all: clean build

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME) $(BINARY_NAME).exe

# Create build directory
$(BUILD_DIR):
	@mkdir -p $(BUILD_DIR)

# Build for current platform
.PHONY: build
build: $(BUILD_DIR)
	@echo "Building $(BINARY_NAME) for current platform..."
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)

# Build for all platforms
.PHONY: build-all
build-all: build-windows build-linux build-macos

# Build for Windows
.PHONY: build-windows
build-windows: $(BUILD_DIR)
	@echo "Building $(BINARY_NAME) for Windows..."
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_FILE)

# Build for Linux
.PHONY: build-linux
build-linux: $(BUILD_DIR)
	@echo "Building $(BINARY_NAME) for Linux..."
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_FILE)

# Build for macOS
.PHONY: build-macos
build-macos: $(BUILD_DIR)
	@echo "Building $(BINARY_NAME) for macOS..."
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_FILE)
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_FILE)

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

# Run unit tests only (exclude integration tests)
.PHONY: test-unit
test-unit:
	@echo "Running unit tests..."
	@go test -v -short ./...

# Run integration tests only
.PHONY: test-integration
test-integration:
	@echo "Running integration tests..."
	@go test -v -run TestIntegration ./...

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Lint code
.PHONY: lint
lint:
	@echo "Linting code..."
	@golint ./...

# Vet code
.PHONY: vet
vet:
	@echo "Vetting code..."
	@go vet ./...

# Run all checks
.PHONY: check
check: fmt vet lint test

# Download dependencies
.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Install binary to GOPATH/bin
.PHONY: install
install:
	@echo "Installing $(BINARY_NAME)..."
	@go install $(LDFLAGS) .

# Run the application (for development)
.PHONY: run
run:
	@go run $(MAIN_FILE) $(ARGS)

# Create release archives
.PHONY: package
package: build-all
	@echo "Creating release packages..."
	@mkdir -p $(BUILD_DIR)/releases
	@cd $(BUILD_DIR) && tar -czf releases/$(BINARY_NAME)-windows-amd64.tar.gz $(BINARY_NAME)-windows-amd64.exe
	@cd $(BUILD_DIR) && tar -czf releases/$(BINARY_NAME)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64
	@cd $(BUILD_DIR) && tar -czf releases/$(BINARY_NAME)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64
	@cd $(BUILD_DIR) && tar -czf releases/$(BINARY_NAME)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64

# Development build (with debug info)
.PHONY: dev-build
dev-build: $(BUILD_DIR)
	@echo "Building $(BINARY_NAME) for development..."
	@go build -race -o $(BUILD_DIR)/$(BINARY_NAME)-dev $(MAIN_FILE)

# Run with race detection (for development)
.PHONY: dev-run
dev-run:
	@go run -race $(MAIN_FILE) $(ARGS)

# Benchmark tests
.PHONY: bench
bench:
	@echo "Running benchmarks..."
	@go test -bench=. -benchmem ./...

# Security check
.PHONY: security
security:
	@echo "Running security checks..."
	@govulncheck ./...

# Generate documentation
.PHONY: docs
docs:
	@echo "Generating documentation..."
	@godoc -http=:6060

# Docker build
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(BINARY_NAME):$(VERSION) .

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all          - Clean and build for current platform"
	@echo "  build        - Build for current platform"
	@echo "  build-all    - Build for all platforms"
	@echo "  build-windows- Build for Windows"
	@echo "  build-linux  - Build for Linux"
	@echo "  build-macos  - Build for macOS"
	@echo "  test         - Run tests"
	@echo "  test-coverage- Run tests with coverage"
	@echo "  test-unit    - Run unit tests only"
	@echo "  test-integration- Run integration tests only"
	@echo "  fmt          - Format code"
	@echo "  lint         - Lint code"
	@echo "  vet          - Vet code"
	@echo "  check        - Run all checks"
	@echo "  deps         - Download dependencies"
	@echo "  install      - Install binary"
	@echo "  run          - Run application"
	@echo "  package      - Create release packages"
	@echo "  dev-build    - Development build with debug info"
	@echo "  dev-run      - Run with race detection"
	@echo "  bench        - Run benchmarks"
	@echo "  security     - Run security checks"
	@echo "  docs         - Generate documentation"
	@echo "  docker-build - Build Docker image"
	@echo "  clean        - Clean build artifacts"
	@echo "  help         - Show this help"
