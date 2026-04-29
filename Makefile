BINARY_NAME=aix
BUILD_DIR=build
GO=go
LDFLAGS=-s -w

# Platforms for cross-compilation
PLATFORMS=darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64 windows/arm64

.PHONY: all dev build build-all test test-coverage lint fmt install clean

all: build-all

# Run for development (like npm start)
dev:
	$(GO) run .

# Build for current platform
build:
	$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) .

# Cross-compile all platforms
build-all:
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*}; \
		GOARCH=$${platform#*/}; \
		EXT=$$( [ "$$GOOS" = "windows" ] && echo ".exe" || echo "" ); \
		OUTPUT=$(BUILD_DIR)/$(BINARY_NAME)-$$GOOS-$$GOARCH$$EXT; \
		echo "Building $$GOOS/$$GOARCH -> $$OUTPUT"; \
		GOOS=$$GOOS GOARCH=$$GOARCH $(GO) build -ldflags "$(LDFLAGS)" -o $$OUTPUT . || exit 1; \
	done

# Run tests
test:
	$(GO) test ./...

# Run tests with coverage
test-coverage:
	$(GO) test -cover ./...

# Lint
lint:
	golangci-lint run

# Format code
fmt:
	gofmt -w .
	goimports -w .

# Install locally
install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
