---
id: SPEC-LAI-001
title: "Lint-as-Instruction - Acceptance Criteria"
---

# Acceptance Criteria

## AC-LAI-001: Error Injection
**Given** a PostToolUse event for Write tool
**When** LSP diagnostics contain 2 errors
**Then** HookOutput.SystemMessage SHALL contain formatted error summary with both errors

## AC-LAI-002: Message Format
**Given** LSP diagnostics with error at main.go:42 "undefined: foo"
**When** systemMessage is generated
**Then** message SHALL contain "[Quality Gate] 1 error(s) detected in main.go:" and "main.go:42: undefined: foo"

## AC-LAI-003: Config Disable
**Given** `ralph.lint_as_instruction` is false
**When** PostToolUse fires with LSP errors
**Then** HookOutput.SystemMessage SHALL be empty (metrics still collected in Data)

## AC-LAI-004: Max 10 Errors
**Given** LSP diagnostics contain 15 errors
**When** systemMessage is generated
**Then** message SHALL show 10 errors and "... and 5 more errors"

## AC-LAI-005: Metrics Preserved
**Given** lint_as_instruction is true
**When** PostToolUse completes
**Then** HookOutput.Data SHALL still contain lsp_diagnostics metrics (backward compatible)

## AC-LAI-006: Warnings Optional
**Given** diagnostics contain 3 warnings and 0 errors, warn_as_instruction is false
**When** PostToolUse fires
**Then** HookOutput.SystemMessage SHALL be empty

## AC-LAI-007: Warnings Enabled
**Given** diagnostics contain 3 warnings and 0 errors, warn_as_instruction is true
**When** PostToolUse fires
**Then** HookOutput.SystemMessage SHALL contain formatted warning summary

## AC-LAI-008: Clean File
**Given** diagnostics are clean (0 errors, 0 warnings)
**When** PostToolUse fires
**Then** HookOutput.SystemMessage SHALL be empty

## AC-LAI-009: AST Security Integration
**Given** AST-grep finds 1 security error after Write
**When** systemMessage is generated
**Then** message SHALL include both LSP errors and AST security findings

## AC-LAI-010: Edit Tool Support
**Given** a PostToolUse event for Edit tool (not just Write)
**When** LSP diagnostics contain errors
**Then** systemMessage SHALL be generated (same as Write)

## AC-LAI-011: Test Coverage
Unit tests SHALL achieve >= 85% coverage for `internal/hook/quality/lint_instruction.go`
