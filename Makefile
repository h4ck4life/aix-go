BINARY_NAME=aix
BUILD_DIR=build
GO=go
LDFLAGS=-s -w

# Platforms for cross-compilation
PLATFORMS=darwin/amd64 darwin/arm64 linux/amd64 linux/arm64 windows/amd64 windows/arm64

.PHONY: all dev build build-all test test-coverage lint fmt install clean tag release

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

# Create a git tag (usage: make tag VERSION=v1.0.0)
tag:
	@if [ -z "$(VERSION)" ]; then \
		echo "Usage: make tag VERSION=v1.0.0"; \
		exit 1; \
	fi
	@echo "Creating tag $(VERSION)..."
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)

# Build and publish release (usage: make release VERSION=v1.0.0)
release: build-all
	@if [ -z "$(VERSION)" ]; then \
		VERSION=$$(git describe --tags --abbrev=0 2>/dev/null); \
		if [ -z "$$VERSION" ]; then \
			echo "No tag found. Create one first: make tag VERSION=v1.0.0"; \
			exit 1; \
		fi; \
		echo "Using latest tag: $$VERSION"; \
	else \
		VERSION=$(VERSION); \
	fi; \
	echo "Publishing release $$VERSION..."; \
	gh release create $$VERSION $(BUILD_DIR)/* \
		--repo h4ck4life/aix-go \
		--title "aix $$VERSION" \
		--generate-notes \
		--verify-tag

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)
