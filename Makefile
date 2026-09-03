# MoAI-ADK Go Edition
# Build and development automation

BINARY_NAME := moai
MODULE := github.com/modu-ai/moai-adk
VERSION ?= $(shell git describe --tags --abbrev=0 2>/dev/null || git rev-parse --short HEAD 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
# BUILD_ID is the MONOTONE build identity, and it is deliberately separate from
# VERSION. VERSION derives with --abbrev=0, which drops the commit-distance
# suffix and so collapses every commit since the last tag onto one string — it
# is a tag floor, not a build identity, and two builds in an ancestor relation
# read as identical through it. Worse, an explicit release-candidate VERSION
# reads HIGHER than a later default build, so comparing version strings reaches
# the opposite conclusion about which binary is newer.
# VERSION stays as it is because it reaches outward (RELEASE_BINARY below,
# version.json, internal/update/local.go); the identity that has to be monotone
# goes here instead, where nothing else consumes it.
BUILD_ID := $(shell git describe --tags --dirty 2>/dev/null || git rev-parse --short HEAD 2>/dev/null || echo "dev")
LDFLAGS := -ldflags "-s -w -X $(MODULE)/pkg/version.Version=$(VERSION) -X $(MODULE)/pkg/version.Commit=$(COMMIT) -X $(MODULE)/pkg/version.Date=$(DATE) -X $(MODULE)/pkg/version.BuildID=$(BUILD_ID)"

# Local release configuration
LOCAL_RELEASE_DIR ?= $(HOME)/.moai/releases
PLATFORM := $(shell go env GOOS)-$(shell go env GOARCH)
RELEASE_BINARY := moai-$(VERSION)-$(PLATFORM)

.PHONY: all build test lint fix clean install generate templ-generate help release-local constitution-check ci-local pr-merge ci-disable verify-required-checks tui-snapshot tui-snapshot-verify preflight lint-fast test-race-short agents-emit agents-emit-check embed-check fmt-check

all: lint test build ## Run lint, test, and build

templ-generate: ## Generate *_templ.go from *.templ sources (pure-Go codegen, no Node)
	go run github.com/a-h/templ/cmd/templ generate -path ./internal/web

build: agents-emit-check templ-generate ## Build the binary
	@go run ./internal/template/scripts/gen-catalog-hashes.go --all
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/moai

agents-emit: ## Regenerate the .codex/agents/moai TOMLs from the neutral .md layer
	AGENTEMIT_UPDATE=1 go test ./internal/template/agentemit/... -run TestGoldenCommittedArtifactsMatchEmission

# Read-only source-layer drift check, wired ahead of `build` so a missed
# regeneration turns red locally instead of waiting for CI. It NEVER writes:
# regeneration stays behind the explicit `agents-emit` verb, because a build
# that silently overwrote a hand edit would erase the evidence CI needs to see.
# AGENTEMIT_UPDATE is scrubbed so an inherited value cannot flip this into the
# regeneration branch.
agents-emit-check: ## Verify the committed .codex TOMLs match the .md source layer (read-only; never regenerates)
	@AGENTEMIT_UPDATE= go test ./internal/template/agentemit/... -run TestGoldenCommittedArtifactsMatchEmission -count=1 \
		|| { printf 'agent-emit drift: committed .codex/agents/moai/*.toml differ from the .md source layer — run `make agents-emit`\n' >&2; exit 1; }

# Embed-axis judgment point: compares the .codex artifacts carried by an
# ALREADY-BUILT binary against the committed ones. It deliberately has no
# `build` prerequisite — a freshly built binary matches the committed set by
# construction, so a check that could only run right after a build would be
# the same tautology it exists to close. Same reason it is not attached to a
# CI build job: CI builds from the commit it checks.
# The runner is built from source; the JUDGMENT TARGET is $(BIN), never rebuilt.
embed-check: ## Verify a built binary's embedded .codex TOMLs match the committed set (BIN=<path>, default bin/moai)
	@MOAI_EMBED_CHECK_BIN=$(or $(BIN),bin/$(BINARY_NAME)) go run ./cmd/moai doctor --check "Agent Emit Embed"

release-local: build ## Create a local release for development updates
	@echo "Creating local release: $(VERSION)"
	@mkdir -p $(LOCAL_RELEASE_DIR)
	@cp bin/$(BINARY_NAME) $(LOCAL_RELEASE_DIR)/$(RELEASE_BINARY)
	@chmod +x $(LOCAL_RELEASE_DIR)/$(RELEASE_BINARY)
	@echo '{"version":"$(VERSION)","date":"$(DATE)","platform":"$(PLATFORM)","binary":"$(RELEASE_BINARY)"}' > $(LOCAL_RELEASE_DIR)/version.json
	@echo "Local release created at: $(LOCAL_RELEASE_DIR)"
	@echo "  Binary: $(RELEASE_BINARY)"
	@echo "  Version: $(VERSION)"

install: ## Install the binary
	go install $(LDFLAGS) ./cmd/moai

test: templ-generate ## Run tests with race detection
	go test -race -coverprofile=coverage.out -covermode=atomic ./...

test-verbose: templ-generate ## Run tests with verbose output
	go test -race -v -coverprofile=coverage.out -covermode=atomic ./...

coverage: test ## Show test coverage report
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

lint: ## Run linters
	golangci-lint run ./...

fix: ## Run go fix modernizers (twice for synergistic fixes)
	go fix ./...
	go fix ./...

vet: ## Run go vet
	go vet ./...

fmt: ## Format code
	gofumpt -l -w .

# Format gate (SPEC-FMT-GATE-001): tracked-files variant of `gofmt -l .` —
# untracked scratch .go files must not flip the local verdict. Silent on a
# clean tree; lists offending files and exits non-zero otherwise. gofumpt
# output (`make fmt`) is gofmt-clean, so the existing fix path still applies.
fmt-check: ## Verify tracked .go files are gofmt-clean (gate predicate; silent on success)
	@files="$$(git ls-files -z '*.go' | xargs -0 gofmt -l)"; \
	if [ -n "$$files" ]; then \
		printf 'gofmt violations found (run gofmt -w or make fmt):\n%s\n' "$$files" >&2; \
		exit 1; \
	fi

generate: ## Run go generate
	go generate ./...

clean: ## Remove build artifacts
	rm -rf bin/ coverage.out coverage.html

tidy: ## Tidy go modules
	go mod tidy

constitution-check: build ## Verify zone registry integrity (SPEC-V3R2-CON-001)
	MOAI_CONSTITUTION_REGISTRY=.claude/rules/moai/core/zone-registry.md \
		./bin/$(BINARY_NAME) constitution list --format json > /dev/null && \
		echo "constitution-check: OK" || echo "constitution-check: WARN (zone-registry.md not found)"

run: build ## Build and run
	./bin/$(BINARY_NAME)

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

ci-local: ## Run CI mirror locally (lint + vet + test + cross-compile)
	@./scripts/ci-mirror/run.sh

pr-merge: ## Enable GitHub auto-merge for a PR (Usage: make pr-merge PR=N [STRATEGY=squash|merge])
	@test -n "$(PR)" || (printf 'Usage: make pr-merge PR=<number> [STRATEGY=squash|merge]\n' >&2 && exit 1)
	@gh pr merge $(PR) --auto --$(or $(STRATEGY),squash)

ci-disable: ## Disable a workflow (set on: workflow_dispatch only). Usage: make ci-disable WORKFLOW=name
	@command -v yq >/dev/null || { echo "Error: yq not found. Install: brew install yq (macOS) or apt install yq (Linux)"; exit 1; }
	@test -n "$(WORKFLOW)" || { echo "Usage: make ci-disable WORKFLOW=<workflow-basename>"; exit 1; }
	@test -f .github/workflows/$(WORKFLOW).yml || { echo "Not found: .github/workflows/$(WORKFLOW).yml"; exit 1; }
	yq -i '.on = {"workflow_dispatch": null}' .github/workflows/$(WORKFLOW).yml
	git add .github/workflows/$(WORKFLOW).yml
	git commit -m "chore(ci): disable $(WORKFLOW) (workflow_dispatch only)"
	@echo "Disabled $(WORKFLOW). Re-enable: edit .github/workflows/$(WORKFLOW).yml on: section."

verify-required-checks: ## Verify SSoT integrity of .github/required-checks.yml
	@./scripts/ci-mirror/validate-required-checks.sh

tui-snapshot: ## Regenerate all internal/tui golden snapshots (UPDATE_GOLDEN=1)
	UPDATE_GOLDEN=1 go test ./internal/tui/... ./internal/tui/golden/... -v
	@echo "Golden snapshots regenerated. Review diffs before committing."

tui-snapshot-verify: ## Verify all internal/tui golden snapshots match on-disk state (no update)
	go test ./internal/tui/... ./internal/tui/golden/... -v -count=1
	@echo "All golden snapshots verified."

preflight: lint-fast test-race-short build ## Run pre-push preflight: lint-fast + test-race-short + build
	@echo "✓ ready to push"

lint-fast: ## Run golangci-lint --fast (preflight gate)
	@golangci-lint run --fast || (echo "preflight: lint-fast FAIL"; exit 1)

test-race-short: ## Run go test -race -short (preflight gate)
	@go test -race -short ./... || (echo "preflight: test-race-short FAIL"; exit 1)

.DEFAULT_GOAL := help
