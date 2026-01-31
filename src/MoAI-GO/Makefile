# Makefile for MoAI-ADK Go

# Variables
GO := go
CGO_ENABLED := 0
VERSION := $(shell git describe --tags --always 2>/dev/null || echo "dev")
LDFLAGS := -s -w -X main.Version=$(VERSION)
BINARY_NAME := moai-go
BUILD_DIR := bin

# Platform-specific binary names
ifeq ($(OS),Windows_NT)
	BINARY := $(BUILD_DIR)/$(BINARY_NAME).exe
else
	BINARY := $(BUILD_DIR)/$(BINARY_NAME)
endif

# Source directories
CMD_DIR := ./cmd/moai
INTERNAL_DIRS := ./internal/...
PKG_DIRS := ./pkg/...

.PHONY: build test lint clean build-all install help

# Default target
all: build

## build: Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build -ldflags="$(LDFLAGS)" -o $(BINARY) $(CMD_DIR)/
	@echo "Build complete: $(BINARY)"

## install: Install the binary to ~/.local/bin
install: build
	@echo "Installing $(BINARY_NAME) to ~/.local/bin..."
	@install -m 0755 $(BINARY) ~/.local/bin/$(BINARY_NAME)
	@echo "Installed: ~/.local/bin/$(BINARY_NAME)"

## test: Run tests with coverage
test:
	@echo "Running tests..."
	$(GO) test -v -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## lint: Run golangci-lint
lint:
	@echo "Running linter..."
	golangci-lint run ./...

## clean: Remove build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	@echo "Clean complete"

## build-all: Build for all platforms
build-all:
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 $(GO) build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_DIR)/
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 $(GO) build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_DIR)/
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO) build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)/
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 $(GO) build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(CMD_DIR)/
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 $(GO) build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_DIR)/
	@echo "Cross-platform builds complete"

## fmt: Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GO) vet ./...

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GO) mod download
	$(GO) mod tidy

## help: Show this help message
help:
	@echo "Available targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'

# Development targets
dev: build lint test

# CI targets
ci: fmt vet lint test
