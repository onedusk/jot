# Jot Makefile for building and releasing
VERSION := $(shell cat VERSION)
BINARY_NAME := jot
BUILD_DIR := build
DIST_DIR := dist-release

# Build flags
LDFLAGS := -ldflags "-X main.version=$(VERSION) -s -w"
GOFLAGS := -trimpath

# Platforms to build for
PLATFORMS := \
	darwin/amd64 \
	darwin/arm64 \
	linux/amd64 \
	linux/arm64 \
	windows/amd64

.PHONY: all
all: clean test build

.PHONY: build
build:
	@echo "Building jot v$(VERSION)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/jot

.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

.PHONY: clean
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR) $(DIST_DIR) $(BINARY_NAME)
	go clean

.PHONY: install
install: build
	@echo "Installing jot to /usr/local/bin..."
	sudo mv $(BINARY_NAME) /usr/local/bin/

.PHONY: release
release: clean test release-build release-package
	@echo "Release v$(VERSION) complete!"
	@echo "Release packages are in $(DIST_DIR)/"

.PHONY: release-build
release-build:
	@echo "Building releases for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		GOOS=$$(echo $$platform | cut -d/ -f1); \
		GOARCH=$$(echo $$platform | cut -d/ -f2); \
		output=$(BUILD_DIR)/$(BINARY_NAME)-$$(echo $$platform | tr / -); \
		echo "Building $$platform..."; \
		if [ "$$GOOS" = "windows" ]; then \
			GOOS=$$GOOS GOARCH=$$GOARCH go build $(LDFLAGS) $(GOFLAGS) -o $$output.exe ./cmd/jot; \
		else \
			GOOS=$$GOOS GOARCH=$$GOARCH go build $(LDFLAGS) $(GOFLAGS) -o $$output ./cmd/jot; \
		fi; \
	done

.PHONY: release-package
release-package:
	@echo "Creating release packages..."
	@mkdir -p $(DIST_DIR)
	@for file in $(BUILD_DIR)/*; do \
		base=$$(basename $$file); \
		name=$$(basename $$file .exe); \
		case "$$base" in \
			*.exe) \
				cd $(BUILD_DIR) && zip ../$(DIST_DIR)/$$name.zip $$base && cd ..; \
				if [ -f README.md ]; then zip $(DIST_DIR)/$$name.zip README.md; fi; \
				if [ -f LICENSE ]; then zip $(DIST_DIR)/$$name.zip LICENSE; fi; \
				if [ -d docs ]; then zip -r $(DIST_DIR)/$$name.zip docs; fi;; \
			*) \
				tar czf $(DIST_DIR)/$$base.tar.gz -C $(BUILD_DIR) $$base; \
				if [ -f README.md ]; then tar rzf $(DIST_DIR)/$$base.tar.gz README.md; fi; \
				if [ -f LICENSE ]; then tar rzf $(DIST_DIR)/$$base.tar.gz LICENSE; fi; \
				if [ -d docs ]; then tar rzf $(DIST_DIR)/$$base.tar.gz docs; fi;; \
		esac; \
	done

.PHONY: dev
dev:
	@echo "Running in development mode with live reload..."
	go run ./cmd/jot watch

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

.PHONY: lint
lint:
	@echo "Running linter..."
	golangci-lint run

.PHONY: coverage
coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: deps
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

.PHONY: update-deps
update-deps:
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t jot:$(VERSION) .

.PHONY: help
help:
	@echo "Jot v$(VERSION) - Build Commands"
	@echo ""
	@echo "Usage:"
	@echo "  make [command]"
	@echo ""
	@echo "Commands:"
	@echo "  all          - Clean, test, and build"
	@echo "  build        - Build the binary for current platform"
	@echo "  test         - Run tests"
	@echo "  clean        - Remove build artifacts"
	@echo "  install      - Build and install to /usr/local/bin"
	@echo "  release      - Build releases for all platforms"
	@echo "  dev          - Run in development mode"
	@echo "  fmt          - Format code"
	@echo "  lint         - Run linter"
	@echo "  coverage     - Generate test coverage report"
	@echo "  deps         - Download dependencies"
	@echo "  update-deps  - Update dependencies"
	@echo "  docker-build - Build Docker image"
	@echo "  help         - Show this help message"