---
id: SPEC-GATE-001
title: "Deterministic Quality Gate"
status: draft
priority: P0
created: "2026-04-07"
harness_pillar: "P3: Verification Loop"
---

# SPEC-GATE-001: Deterministic Quality Gate

## Overview

git commit 실행 전 go vet, golangci-lint, go test를 기계적으로 실행하여 실패 시 commit을 차단하는 결정론적 품질 게이트.

하네스 엔지니어링 프레임워크의 핵심 원칙: "AI는 결코 결정론적 게이트를 건너뛸 수 없다."

## Requirements (EARS Format)

### REQ-GATE-001 (Event-Driven)
When PreToolUse event fires for Bash tool AND the command contains `git commit`, the system SHALL execute quality gate checks before allowing the commit.

### REQ-GATE-002 (Ubiquitous)
The quality gate SHALL run `go vet ./...` and return Decision "deny" with the error output if it fails (exit code != 0).

### REQ-GATE-003 (Ubiquitous)
The quality gate SHALL run `golangci-lint run` and return Decision "deny" with the error output if it fails (exit code != 0).

### REQ-GATE-004 (Ubiquitous)
The quality gate SHALL run `go test ./...` and return Decision "deny" with the error output if it fails (exit code != 0).

### REQ-GATE-005 (Ubiquitous)
The quality gate checks SHALL execute sequentially: go vet → golangci-lint → go test. If any step fails, subsequent steps SHALL be skipped.

### REQ-GATE-006 (State-Driven)
When `gate.enabled` is false in `.moai/config/sections/gate.yaml`, the quality gate SHALL be skipped entirely.

### REQ-GATE-007 (State-Driven)
When `gate.skip_tests` is true, the go test step SHALL be skipped (for quick commits during development).

### REQ-GATE-008 (Ubiquitous)
Each quality gate step SHALL have a configurable timeout (default: go vet 30s, golangci-lint 60s, go test 120s).

### REQ-GATE-009 (Event-Driven)
When a quality gate step times out, the system SHALL treat it as a failure and return Decision "deny" with a timeout message.

### REQ-GATE-010 (Ubiquitous)
The deny reason SHALL include the full error output from the failed step, formatted for readability.

### REQ-GATE-011 (Event-Driven)
When the command is `git commit --amend` or contains `--no-verify`, the quality gate SHALL still execute (no bypass).

## Architecture

```
Bash(git commit) detected in PreToolUse
  → preToolHandler.checkBashCommand()
  → isGitCommit(command) == true
  → qualityGate.Run(ctx)
    → Step 1: go vet ./...     (fail → deny)
    → Step 2: golangci-lint run (fail → deny)
    → Step 3: go test ./...    (fail → deny, skippable)
  → All pass → allow
```

## Implementation Scope

### New Files
- `internal/hook/quality/gate.go` — QualityGate struct with Run() method
- `internal/hook/quality/gate_test.go` — Unit tests
- `internal/template/templates/.moai/config/sections/gate.yaml` — Default config

### Modified Files
- `internal/hook/pre_tool.go` — Add git commit detection in checkBashCommand()
- `internal/hook/pre_tool_test.go` — Add gate integration tests

## Non-Goals
- Dynamic per-language gate detection (Go only for now)
- Parallel step execution (sequential is simpler and sufficient)
- Integration with CI/CD (this is local-only)
