---
id: SPEC-GATE-001
title: "Deterministic Quality Gate - Acceptance Criteria"
---

# Acceptance Criteria

## AC-GATE-001: Git Commit Detection
**Given** a PreToolUse event with Bash tool
**When** the command contains `git commit`
**Then** the quality gate SHALL execute before allowing the command

## AC-GATE-002: Go Vet Gate
**Given** quality gate is enabled
**When** `go vet ./...` fails with exit code != 0
**Then** return Decision "deny" with the go vet error output

## AC-GATE-003: Golangci-lint Gate
**Given** quality gate is enabled and go vet passes
**When** `golangci-lint run` fails with exit code != 0
**Then** return Decision "deny" with the lint error output

## AC-GATE-004: Go Test Gate
**Given** quality gate is enabled, go vet passes, lint passes, and skip_tests is false
**When** `go test ./...` fails with exit code != 0
**Then** return Decision "deny" with the test failure output

## AC-GATE-005: Sequential Execution
**Given** go vet fails
**When** the quality gate processes the commit
**Then** golangci-lint and go test SHALL NOT be executed

## AC-GATE-006: Config Disable
**Given** `gate.enabled` is false in gate.yaml
**When** a git commit is attempted
**Then** the quality gate SHALL be skipped and the commit allowed

## AC-GATE-007: Skip Tests Flag
**Given** `gate.skip_tests` is true
**When** the quality gate runs
**Then** only go vet and golangci-lint SHALL execute, go test is skipped

## AC-GATE-008: Timeout Handling
**Given** go test is running and exceeds 120s timeout
**When** timeout fires
**Then** return Decision "deny" with message "quality gate timed out: go test exceeded 120s"

## AC-GATE-009: No Bypass
**Given** a command `git commit --no-verify -m "test"`
**When** PreToolUse processes it
**Then** the quality gate SHALL still execute (--no-verify does not bypass)

## AC-GATE-010: All Pass
**Given** go vet, golangci-lint, and go test all pass
**When** the quality gate completes
**Then** return allow (original PreToolUse flow continues)

## AC-GATE-011: Error Output Formatting
**Given** golangci-lint fails with 3 issues
**When** deny decision is returned
**Then** the reason SHALL contain the full linter output with file:line references

## AC-GATE-012: Test Coverage
Unit tests SHALL achieve >= 85% coverage for `internal/hook/quality/gate.go`
