---
id: SPEC-LSP-LOOP-005
version: "1.0.0"
status: completed
created: "2026-04-11"
updated: "2026-04-11"
author: GOOS
priority: P2
issue_number: 0
phase: "Phase 4 - Multi-Language LSP"
module: "internal/loop/, internal/ralph/, internal/hook/post_tool.go"
estimated_loc: 1200
dependencies:
  - SPEC-LSP-AGG-003
lifecycle: spec-anchored
tags: lsp, loop, ralph, feedback, classification
---

# SPEC-LSP-LOOP-005: Loop/Ralph LSP Integration

## HISTORY

| Date | Version | Change |
|------|---------|--------|
| 2026-04-11 | 1.0.0 | Initial draft — resolves audit A4 findings |

---

## Overview

| Field | Value |
|-------|-------|
| SPEC ID | SPEC-LSP-LOOP-005 |
| Title | Loop/Ralph LSP Integration |
| Status | Completed |
| Priority | P2 |
| Depends on | SPEC-LSP-AGG-003 |

## Problem Statement

Audit report A4 found:

- `GoFeedbackGenerator` uses only `go test` + `go vet` CLI output; **LSP diagnostics never feed into Ralph's decisions** (Gap A4-2)
- Ralph's `ClassifyFeedback` reads only integer counts, not LSP Diagnostic metadata (severity, source, code) — Gap A4-3
- PostTool hook collects LSP diagnostics for `systemMessage` injection but **data evaporates** — never reaches loop controller (Gap A4-4)
- `RalphConfig.LintAsInstruction` / `WarnAsInstruction` flags exist in config but are unused by `Decide()` (Gap A4-6)

SPEC-GOPLS-BRIDGE-001 provides the Go-only bridge; this SPEC wires Go-language feedback from that bridge (and its Aggregator facade in SPEC-LSP-AGG-003) into Ralph's decision loop.

### Scope Clarification: Go-Only

This SPEC is **explicitly scoped to the Go language**. `GoFeedbackGenerator` remains Go-specific by name and implementation. The `Aggregator` dependency is used because it provides the facade for Go diagnostic collection (via SPEC-LSP-AGG-003), not because this SPEC attempts to generalize feedback across all 16 MoAI-supported languages. Multi-language feedback (e.g., a `PythonFeedbackGenerator` or a generic `FeedbackGenerator` interface with per-language implementations) is deferred to a future SPEC.

## Goal

Complete the diagnostic feedback loop:

```
Edit file
  ↓
PostTool hook → Aggregator.GetDiagnostics
  ↓
Feedback.Diagnostics populated
  ↓
Ralph.ClassifyFeedback (severity/source-aware)
  ↓
Decision (auto-fix / approval / manual / blocker)
  ↓
Loop controller
```

## Requirements (EARS Format)

**REQ-LL-001**: `Feedback` struct SHALL carry `Diagnostics []lsp.Diagnostic` populated from Aggregator output.

**REQ-LL-002**: `GoFeedbackGenerator.Collect` SHALL invoke `Aggregator.GetDiagnostics` for Go source files when the aggregator is available, filter the returned diagnostics to Go-only results, and merge them into the Feedback struct.

**REQ-LL-003**: PostTool hook SHALL emit LSP diagnostics to both: (a) agent conversation via `systemMessage` (existing), and (b) LoopController via a new `RecordFeedback` channel.

**REQ-LL-004**: `Ralph.ClassifyFeedback(feedback)` SHALL inspect `feedback.Diagnostics` when non-empty, classifying by `Severity` and `Source` instead of integer counts.

**REQ-LL-005**: Classification rules:
- `Severity=Error` + `Source=compiler` → ErrorLevelBlocker
- `Severity=Error` + `Source=staticcheck SA*` → ErrorLevelApproval
- `Severity=Warning` + `Source=staticcheck` → ErrorLevelAutoFix
- `Severity=Information` → ErrorLevelSkip
- `Severity=Hint` → ErrorLevelSkip

**REQ-LL-006**: `RalphConfig.LintAsInstruction` SHALL influence classification when true, treating warning-severity as instruction-level input to the agent (no decision block).

**REQ-LL-007**: `RalphConfig.WarnAsInstruction` SHALL cause warning-severity diagnostics to be injected as agent instruction rather than gate block.

**REQ-LL-008**: Backwards compatibility: when `feedback.Diagnostics` is empty, existing integer-based classification SHALL remain as fallback (no regression).

**REQ-LL-009**: Loop controller SHALL receive feedback events via a bounded channel; overflow drops oldest events with a warn log.

**REQ-LL-010**: Re-planning trigger (stagnation detection) SHALL consider diagnostic counts trending up as a stagnation signal.

## Non-Goals

- **Multi-language feedback generators**: Python, TypeScript, Rust, and other languages are NOT in scope. `GoFeedbackGenerator` remains Go-specific. A generic `FeedbackGenerator` interface is explicitly deferred to a future SPEC.
- **Generic Aggregator consumption for non-Go languages**: Although `Aggregator` (SPEC-LSP-AGG-003) is multi-language capable, this SPEC only wires Go diagnostics into Ralph's classification logic.
- Web UI for diagnostic visualization
- Historical diagnostic trends across sessions

## References

- Audit report A4 (Gaps A4-1 through A4-7)
- `internal/loop/feedback.go`, `internal/loop/go_feedback.go`
- `internal/ralph/engine.go`
- `.claude/rules/moai/core/agent-common-protocol.md`
