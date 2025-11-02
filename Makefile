# Project information
APP_NAME    := jira-parser
VERSION     := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "dev")
COMMIT      := $(shell git rev-parse --short HEAD)
DATE        := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# Build flags
GC_FLAGS    := -trimpath
LD_FLAGS    := -s -w -X main.buildVersion=$(VERSION) -X main.buildCommit=$(COMMIT) -X main.buildDate=$(DATE)
LINK_MODE   := -buildmode=pie -ldflags="$(LD_FLAGS)"
OPTIMIZATION := -gcflags="$(GC_FLAGS)" $(LINK_MODE)

# Build directories
DIST_DIR    := dist
BIN_DIR     := bin

.PHONY: all
all: build

# Standard build with optimizations
.PHONY: build
build: $(BIN_DIR)/$(APP_NAME)

$(BIN_DIR)/$(APP_NAME):
	@mkdir -p $(BIN_DIR)
	@echo "Building $(APP_NAME) version: $(VERSION), commit: $(COMMIT)"
	go build $(OPTIMIZATION) -o $(BIN_DIR)/$(APP_NAME) ./cmd/jira-parser/main.go
	@echo "Build complete: $(BIN_DIR)/$(APP_NAME)"
	@echo "Check version: ./$(BIN_DIR)/$(APP_NAME) --version"

# Debug build (without optimizations)
.PHONY: build-debug
build-debug:
	@mkdir -p $(BIN_DIR)
	@echo "Building debug version..."
	go build -race -o $(BIN_DIR)/$(APP_NAME)-debug ./cmd/jira-parser/main.go
	@echo "Debug build complete: $(BIN_DIR)/$(APP_NAME)-debug"

# Minimal build with UPX compression
.PHONY: build-minimal
build-minimal: build
	@which upx > /dev/null || (echo "UPX not installed, skipping compression" && exit 0)
	@echo "Compressing binary with UPX..."
	upx --best --lzma $(BIN_DIR)/$(APP_NAME)
	@echo "Compression complete"

# Cross-compilation for releases
.PHONY: release
release: clean build-linux build-windows build-macos
	@echo "Release builds complete in $(DIST_DIR)/"

build-linux:
	@mkdir -p $(DIST_DIR)
	@echo "Building Linux binaries..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(OPTIMIZATION) -o $(DIST_DIR)/$(APP_NAME)-linux-amd64 ./cmd/jira-parser/main.go
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(OPTIMIZATION) -o $(DIST_DIR)/$(APP_NAME)-linux-arm64 ./cmd/jira-parser/main.go
	@echo "Linux binaries: $(DIST_DIR)/$(APP_NAME)-linux-*"

build-windows:
	@mkdir -p $(DIST_DIR)
	@echo "Building Windows binaries..."
	GOOS=windows GOARCH=amd64 go build $(OPTIMIZATION) -o $(DIST_DIR)/$(APP_NAME)-windows-amd64.exe ./cmd/jira-parser/main.go
	@echo "Windows binary: $(DIST_DIR)/$(APP_NAME)-windows-amd64.exe"

build-macos:
	@mkdir -p $(DIST_DIR)
	@echo "Building macOS binaries..."
	GOOS=darwin GOARCH=amd64 go build $(OPTIMIZATION) -o $(DIST_DIR)/$(APP_NAME)-darwin-amd64 ./cmd/jira-parser/main.go
	GOOS=darwin GOARCH=arm64 go build $(OPTIMIZATION) -o $(DIST_DIR)/$(APP_NAME)-darwin-arm64 ./cmd/jira-parser/main.go
	@echo "macOS binaries: $(DIST_DIR)/$(APP_NAME)-darwin-*"

# Install to GOPATH/bin
.PHONY: install
install:
	@echo "Installing $(APP_NAME) to GOPATH/bin..."
	go install -ldflags="$(LD_FLAGS)" ./cmd/jira-parser/main.go
	@echo "Installation complete"

# Development tools
.PHONY: test
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

.PHONY: test-coverage
test-coverage: test
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

.PHONY: lint
lint:
	@echo "Running linters..."
	golangci-lint run

.PHONY: tidy
tidy:
	@echo "Tidying Go modules..."
	go mod tidy

# Clean up
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BIN_DIR) $(DIST_DIR) coverage.out coverage.html
	@echo "Clean complete"

# Version info
.PHONY: version
version:
	@echo "Project: $(APP_NAME)"
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Date: $(DATE)"

# Help
.PHONY: help
help:
	@echo "Available targets for $(APP_NAME):"
	@echo ""
	@echo "  Build targets:"
	@echo "    build           - Build for current platform (to bin/)"
	@echo "    build-debug     - Build with race detector for debugging"
	@echo "    build-minimal   - Build with UPX compression (requires UPX)"
	@echo "    release         - Build releases for all platforms (to dist/)"
	@echo "    install         - Install to GOPATH/bin"
	@echo ""
	@echo "  Development:"
	@echo "    test            - Run tests with race detector"
	@echo "    test-coverage   - Run tests and generate HTML coverage report"
	@echo "    lint            - Run golangci-lint"
	@echo "    tidy            - Tidy Go modules"
	@echo ""
	@echo "  Utility:"
	@echo "    clean           - Remove all build artifacts"
	@echo "    version         - Show version information"
	@echo "    help            - Show this help message"
