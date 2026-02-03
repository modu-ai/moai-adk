# MoAI-ADK Go Edition
# Build and development automation

BINARY_NAME := moai
MODULE := github.com/modu-ai/moai-adk-go
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-s -w -X $(MODULE)/pkg/version.Version=$(VERSION) -X $(MODULE)/pkg/version.Commit=$(COMMIT) -X $(MODULE)/pkg/version.Date=$(DATE)"

.PHONY: all build test lint clean install generate help

all: lint test build ## Run lint, test, and build

build: ## Build the binary
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/moai

install: ## Install the binary
	go install $(LDFLAGS) ./cmd/moai

test: ## Run tests with race detection
	go test -race -coverprofile=coverage.out -covermode=atomic ./...

test-verbose: ## Run tests with verbose output
	go test -race -v -coverprofile=coverage.out -covermode=atomic ./...

coverage: test ## Show test coverage report
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

lint: ## Run linters
	golangci-lint run ./...

vet: ## Run go vet
	go vet ./...

fmt: ## Format code
	gofumpt -l -w .

generate: ## Run go generate
	go generate ./...

clean: ## Remove build artifacts
	rm -rf bin/ coverage.out coverage.html

tidy: ## Tidy go modules
	go mod tidy

run: build ## Build and run
	./bin/$(BINARY_NAME)

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
