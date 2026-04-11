---
id: SPEC-LSP-LOOP-005
version: "1.0.0"
status: draft
created: "2026-04-11"
updated: "2026-04-11"
author: GOOS
priority: P2
issue_number: 0
phase: "Phase 4 - Multi-Language LSP"
module: "internal/loop/, internal/ralph/, internal/hook/post_tool.go"
estimated_loc: 600
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
| Priority | P2 |
| Depends on | SPEC-LSP-AGG-003 |

## Problem Statement

Audit report A4 found:

- `GoFeedbackGenerator` uses only `go test` + `go vet` CLI output; **LSP diagnostics never feed into Ralph's decisions** (Gap A4-2)
- Ralph's `ClassifyFeedback` reads only integer counts, not LSP Diagnostic metadata (severity, source, code) — Gap A4-3
- PostTool hook collects LSP diagnostics for `systemMessage` injection but **data evaporates** — never reaches loop controller (Gap A4-4)
- `RalphConfig.LintAsInstruction` / `WarnAsInstruction` flags exist in config but are unused by `Decide()` (Gap A4-6)

SPEC-GOPLS-BRIDGE-001 provides the Go-only bridge; this SPEC wires the generalized path via the Aggregator.

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

**REQ-LL-002**: `GoFeedbackGenerator.Collect` SHALL invoke `Aggregator.GetDiagnostics` when bridge is available and merge results into the Feedback struct.

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

- Web UI for diagnostic visualization
- Historical diagnostic trends across sessions

## References

- Audit report A4 (Gaps A4-1 through A4-7)
- `internal/loop/feedback.go`, `internal/loop/go_feedback.go`
- `internal/ralph/engine.go`
- `.claude/rules/moai/core/agent-common-protocol.md`
