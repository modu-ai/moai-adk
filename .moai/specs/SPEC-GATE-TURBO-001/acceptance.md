---
id: SPEC-GATE-TURBO-001
title: "Turbo-safe Node test command in quality gate - Acceptance Criteria"
---

# Acceptance Criteria

## AC-GATE-TURBO-001: Turbo monorepo uses turbo-safe test command
**Given** a project dir containing both `package.json` and `turbo.json`
**When** `detectToolchain()` matches the Node.js toolchain
**Then** the returned `testStep` args SHALL be `[]string{"test"}` and SHALL NOT contain `--passWithNoTests`

## AC-GATE-TURBO-002: Non-turbo Node project unchanged
**Given** a project dir containing `package.json` but NO `turbo.json`
**When** `detectToolchain()` matches the Node.js toolchain
**Then** the returned `testStep` args SHALL remain `[]string{"test", "--", "--passWithNoTests"}` (no regression)

## AC-GATE-TURBO-003: Shared toolchains slice not mutated
**Given** the package-level `toolchains` slice
**When** `resolveNodeTurbo` returns a turbo variant
**Then** the original Node.js entry in `toolchains` SHALL retain its `--passWithNoTests` args (variant returned, slice unmutated)

## AC-GATE-TURBO-004: Lint step preserved
**Given** a turbo monorepo Node.js project
**When** the turbo-safe variant is returned
**Then** the eslint `lintSteps` SHALL be unchanged from the original Node.js toolchain entry

## AC-GATE-TURBO-005: Other languages unaffected
**Given** a Go, Python, or Rust project
**When** `detectToolchain()` runs
**Then** its toolchain SHALL be returned unchanged (resolveNodeTurbo applies only to the Node.js toolchain)

## AC-GATE-TURBO-006: Test Coverage
Unit tests SHALL cover both paths: `turbo.json` present (no `--passWithNoTests`) and `turbo.json` absent (flag retained) for `internal/hook/quality/gate.go`.
