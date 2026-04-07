---
id: SPEC-LAI-001
title: "Lint-as-Instruction Pattern"
status: draft
priority: P0
created: "2026-04-07"
harness_pillar: "P3: Verification Loop"
---

# SPEC-LAI-001: Lint-as-Instruction Pattern

## Overview

PostToolUse 훅에서 Write/Edit 후 수집된 LSP 에러를 `systemMessage`로 포맷팅하여 AI의 다음 프롬프트에 자동 주입하는 패턴.

하네스 엔지니어링의 핵심: "에러 메시지가 AI의 다음 프롬프트가 된다" (Lint-as-Instruction)

## Requirements (EARS Format)

### REQ-LAI-001 (Event-Driven)
When PostToolUse event fires for Write or Edit tool AND LSP diagnostics contain errors (severity: error), the system SHALL include a formatted error summary in the `systemMessage` field of HookOutput.

### REQ-LAI-002 (Ubiquitous)
The systemMessage SHALL follow this format:
```
[Quality Gate] {N} error(s) detected in {filename}:
- {file}:{line}: {message} ({source})
- {file}:{line}: {message} ({source})
Fix these errors before proceeding.
```

### REQ-LAI-003 (State-Driven)
When `ralph.lint_as_instruction` is false in ralph.yaml, the system SHALL skip systemMessage injection (observation-only mode preserved).

### REQ-LAI-004 (Ubiquitous)
The systemMessage SHALL be limited to max 10 error entries. If more errors exist, append "... and {N} more errors".

### REQ-LAI-005 (Ubiquitous)
The system SHALL continue to include diagnostics in the `Data` field (existing metrics behavior unchanged).

### REQ-LAI-006 (State-Driven)
When diagnostics contain only warnings (no errors), the system SHALL include warnings in `systemMessage` only if `ralph.warn_as_instruction` is true (default: false).

### REQ-LAI-007 (Ubiquitous)
The system SHALL NOT include `systemMessage` when diagnostics are clean (0 errors, 0 warnings or warn_as_instruction false).

### REQ-LAI-008 (Event-Driven)
When AST-grep scan finds security issues after Write, the security findings SHALL also be included in the systemMessage alongside LSP errors.

## Architecture

```
PostToolUse(Write/Edit)
  → collectDiagnostics() [existing — unchanged]
  → formatDiagnosticsAsInstruction() [NEW]
    → errors > 0 → systemMessage = formatted errors
    → warnings only → systemMessage only if warn_as_instruction
    → clean → no systemMessage
  → HookOutput{Data: metrics, SystemMessage: instruction}
```

## Implementation Scope

### New Files
- `internal/hook/quality/lint_instruction.go` — FormatDiagnosticsAsInstruction()
- `internal/hook/quality/lint_instruction_test.go` — Unit tests

### Modified Files
- `internal/hook/post_tool.go` — Add systemMessage injection after collectDiagnostics
- `internal/hook/post_tool_test.go` — Add LAI integration tests
- `internal/template/templates/.moai/config/sections/ralph.yaml` — Add lint_as_instruction: true

## Non-Goals
- Blocking Write/Edit operations (observation-only, instruction injection only)
- Custom formatting per language (generic format for now)
- Integration with MX tag warnings (separate concern)
