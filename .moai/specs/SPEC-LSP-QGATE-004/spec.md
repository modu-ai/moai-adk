---
id: SPEC-LSP-QGATE-004
version: "1.0.0"
status: completed
created: "2026-04-11"
updated: "2026-04-11"
author: GOOS
priority: P2
issue_number: 0
phase: "Phase 4 - Multi-Language LSP"
module: "internal/lsp/hook/, internal/hook/quality/"
estimated_loc: 2400
dependencies:
  - SPEC-LSP-AGG-003
lifecycle: spec-anchored
tags: lsp, quality-gate, phase-aware, trust5
---

# SPEC-LSP-QGATE-004: Phase-Aware LSP Quality Gates

## HISTORY

| Date | Version | Change |
|------|---------|--------|
| 2026-04-11 | 1.0.0 | Initial draft — resolves 11 orphaned config keys found in audit A3 |
| 2026-04-11 | 1.0.1 | LOC reconciliation: `estimated_loc` raised from 800 → 2400 to match expanded plan.md W12 scope (PR #631). Original 800 estimate reflected narrower initial scope; actual ceiling including test code and golden fixtures is ~2400, still within P2 budget. |

---

## Overview

| Field | Value |
|-------|-------|
| SPEC ID | SPEC-LSP-QGATE-004 |
| Title | Phase-Aware LSP Quality Gates |
| Status | Draft |
| Priority | P2 |
| Depends on | SPEC-LSP-AGG-003 |

## Problem Statement

The previous audit (report A3) found **11 of 13 config keys in `lsp_quality_gates` are orphaned** — defined in `.moai/config/sections/quality.yaml` but never read by any code:

- `plan.require_baseline` — not enforced
- `run.max_type_errors` — not read
- `run.max_lint_errors` — not read
- `run.allow_regression` — not enforced
- `sync.max_errors` — not read (only sync.max_warnings is)
- `sync.require_clean_lsp` — not enforced (a documented TRUST 5 gate)
- `cache_ttl_seconds` — not implemented
- `timeout_seconds` — not read
- `lsp_integration.trust5_integration.*` — not wired
- `lsp_integration.regression_detection.*` — not implemented

Additionally, `internal/lsp/hook/gate.go` bypasses the centralized `ConfigManager`, reads yaml directly, and uses a duplicate struct `qualityYAMLConfig` instead of `pkg/models.QualityConfig`.

## Goal

1. **Wire all 11 orphaned keys** to real enforcement code
2. **Phase context** (plan/run/sync) passed to gate enforcer
3. **Centralized config access** via `ConfigManager`
4. **Regression detection** using baseline comparison
5. **Clean LSP state enforcement** for sync phase

## Requirements (EARS Format)

**REQ-QG-001**: The `QualityGate` struct SHALL accept a `Phase` field (plan/run/sync/auto).

**REQ-QG-002**: When phase is `plan`, the system SHALL capture a baseline LSP state if `require_baseline: true`.

**REQ-QG-003**: When phase is `run`, the system SHALL enforce `max_errors`, `max_type_errors`, `max_lint_errors`, and `allow_regression` thresholds.

**REQ-QG-004**: When phase is `sync`, the system SHALL enforce `max_errors`, `max_warnings`, and `require_clean_lsp` (all diagnostics must be fixed or explicitly dismissed).

**REQ-QG-005**: The gate SHALL use `ConfigManager.Load()` instead of direct yaml reads.

**REQ-QG-006**: The duplicate `qualityYAMLConfig` struct in `gate.go` SHALL be removed; all consumers SHALL use `pkg/models.QualityConfig`.

**REQ-QG-007**: Baseline comparison SHALL detect regression per `regression_detection.error_increase_threshold` and `warning_increase_threshold`.

**REQ-QG-008**: Cache TTL and timeouts SHALL be delegated to SPEC-LSP-AGG-003 Aggregator via config passthrough.

**REQ-QG-009**: The `trust5_integration` mapping SHALL influence which dimensions (tested/readable/understandable/secured/trackable) include LSP diagnostics in their scoring.

**REQ-QG-010**: Phase context SHALL be discoverable from the hook invocation context (which workflow is active).

## Non-Goals

- New LSP client (uses SPEC-LSP-AGG-003)
- UI for viewing diagnostics (already in PostToolUse hook)

## References

- Audit report A3 (orphaned config keys)
- `pkg/models/config.go` (canonical QualityConfig)
- `.moai/config/sections/quality.yaml` lsp_quality_gates section
