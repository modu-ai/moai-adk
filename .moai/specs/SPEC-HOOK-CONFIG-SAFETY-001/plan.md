---
id: SPEC-HOOK-CONFIG-SAFETY-001
title: "moai update Configuration Safety /hooks Review Guidance"
version: "0.1.0"
status: completed
created: 2026-07-07
updated: 2026-07-07
author: manager-spec (P5 of hooks-hardening Epic)
priority: P3
phase: "v3.0.0"
module: "internal/cli"
lifecycle: spec-anchored
tags: "hooks, cli, configuration-safety, ux, discoverability"
---

# plan.md — SPEC-HOOK-CONFIG-SAFETY-001

> Tier S implementation plan. Single domain (CLI completion UX). 2-3 files. No architectural change. P5 of hooks-hardening Epic.

## §A Context — Design Decision: Guidance Trigger Strategy

### §A.1 The three options

This SPEC adds `/hooks` review guidance to `moai update` completion output. The design decision is WHICH update-flow branch(es) trigger the guidance emission. Three options were considered.

**Option A — Conditional on hooks-section actual change (JSON-diff)**
- Trigger: emit guidance IFF the `.claude/settings.json` hooks section actually changed between the pre-sync and post-sync state.
- Mechanism: parse pre/post settings.json, compare `hooks` key JSON equality, emit guidance only on diff.
- Pros: maximum precision — guidance fires only when CC's snapshot is genuinely stale.
- Cons: adds JSON-diff complexity (new helper, edge cases for empty hooks, key ordering, formatting normalization). False-negative risk if JSON comparison is too strict (formatting-only diff) or too loose (semantically-meaningful diff missed). Disproportionate implementation cost for a UX hint.

**Option B — Unconditional on "Template sync complete" branch (RECOMMENDED)**
- Trigger: emit guidance whenever the update flow takes the "Template sync complete" branch (`internal/cli/update.go`, content token `Label: "Template sync complete"`).
- Mechanism: append one stdout guidance line immediately after the existing "Template sync complete" pill.
- Pros: simplest implementation (single insertion point, no JSON-diff helper). Covers primary maintainer use case (mid-session `moai update` after a template change). Does NOT fire on no-op paths by construction (the no-op branches are distinct pills at `Label: "Already up to date"`, `Label: "Template version up-to-date · Skipping sync"`, `Label: "Binary updated (template sync skipped)"`).
- Cons: may emit guidance when template sync ran but the hooks section did not actually change (e.g., template re-render produced byte-identical hooks JSON). Acceptable for a UX hint — the cost of a spurious reminder is one terminal line, and reviewing `/hooks` is harmless.

**Option C — Hybrid (update full + init brief hint)**
- Trigger: Option B for `moai update`, PLUS a brief one-line hint at `moai init` completion (`internal/cli/init.go:434-444` success card) pointing at `/hooks`.
- Mechanism: two insertion points (`update.go` "Template sync complete" pill + `init.go` success card `details` slice).
- Pros: covers the rarer edge case of running `moai init` from inside an existing CC session (via Bash tool) where the new project's settings.json would not be picked up by the running session's snapshot.
- Cons: slightly larger surface area (2 source files instead of 1). The common `moai init` workflow (terminal → `cd` → start CC) has NO stale-snapshot risk because CC captures the snapshot at fresh session start, so the init-time hint is addressing an edge case.

### §A.2 Recommendation: Option B (Tier S fit)

Per `moai-constitution.md § Agent Core Behaviors #4 Enforce Simplicity` (reach for the cheapest sufficient capability via the reuse-and-dependency-avoidance decision ladder), **Option B is the recommended trigger strategy** for this Tier S SPEC. Reasoning:

1. **Sufficient coverage** — Option B fires on the canonical "template sync ran" branch. The maintainer's noted frequent pattern (mid-session `moai update`) is fully covered.
2. **No spurious fire on no-op paths** — The update flow has multiple no-op branches (`update.go` content tokens `"Already up to date"`, `"Template version up-to-date · Skipping sync"`, `"Binary updated (template sync skipped)"`). Option B's single trigger condition (the "Template sync complete" branch) excludes all no-op paths by construction — they are distinct code branches with distinct pill labels.
3. **Avoids Option A's gold-plating** — For a UX hint, JSON-diff precision is over-engineering. The cost of a spurious reminder (when template sync ran but hooks didn't change) is one terminal line; the cost of Option A's implementation (JSON-diff helper, edge cases, false-negative risk) is disproportionately high.
4. **Defers Option C's edge case** — The common `moai init` workflow has no stale-snapshot risk. The edge case (running `moai init` from inside an existing CC session via Bash) is rarer and can be addressed in a follow-up SPEC if user feedback indicates confusion.

Option C is the second choice if the user explicitly wants init-time coverage in this SPEC. Option A is rejected as over-engineering for Tier S.

### §A.3 Insertion point (Option B)

**File**: `internal/cli/update.go`
**Content token anchor**: `Label: "Template sync complete"` (the canonical "template sync ran" branch; grep-date line ~874, but run-phase MUST locate by content token per `feedback_line_number_drift_asymmetry`).

The existing code emits a single pill:
```go
_, _ = fmt.Fprintln(out, tui.Pill(tui.PillOpts{Kind: tui.PillOk, Solid: false, Label: "Template sync complete", Theme: &th}))
```

Proposed: append a stdout guidance line immediately after this pill. The guidance string includes the literal token `/hooks` and conveys that running CC sessions require `/hooks` review (or a CC restart) for the new/changed hooks to take effect.

The exact guidance wording is deferred to run-phase (per SPEC scope-boundary rule — WHAT not HOW). The substring `/hooks` is the stable anchor asserted by the characterization test (acceptance.md AC-CS-001, AC-CS-003).

## §B Known Issues

### §B.1 update.go "Configuration updated successfully" branch
Grep also surfaced a `Label: "Configuration updated successfully"` pill (~line 2591 at grep date). The context of this branch is NOT fully clear from grep alone — it may be a separate config-update path that also touches `.claude/settings.json` hooks. Run-phase MUST characterize this branch (read surrounding code) and decide whether guidance should also fire there. Default: emit guidance if and only if the branch is reached via template-sync-equivalent path (i.e., template was actually re-rendered); do NOT emit if it's a no-op confirmation. This is a characterization-test decision, not a design decision.

### §B.2 update.go "Updated X → Y" version-bump branch
A `Label: "Updated " + result.PreviousVersion + " → " + result.NewVersion` pill exists (~line 449). Whether template sync also runs on this branch is unclear from grep alone. Run-phase MUST characterize this branch. If template sync runs here too, guidance should fire; if not (pure binary-bump path with template sync skipped), guidance should not fire. Default: do NOT emit guidance on the pure binary-bump path.

### §B.3 Insertion-point line drift
The line numbers cited in this plan (~874, ~239, ~534, ~927, ~228, ~449, ~2591) are accurate as of the grep date (2026-07-07). Run-phase MUST re-locate every insertion point by content token (the `Label:` string), NOT by line number, per `feedback_line_number_drift_asymmetry`.

### §B.4 `--json` output interaction
If `moai update --json` is supported, the guidance MUST NOT corrupt the JSON output (acceptance.md §D.3). Run-phase MUST characterize the `--json` behavior on `moai update` before finalizing. If `--json` is not supported on `moai update`, this issue is moot.

## §C Pre-flight (run-phase entry conditions)

Before M1 implementation begins, the run-phase agent MUST verify:
1. The insertion point at `internal/cli/update.go` is located by content token (`Label: "Template sync complete"`), NOT by line number.
2. The characterization test compiles against the current `internal/cli/update_test.go` test scaffolding.
3. The three no-op paths identified in §A.1 (content tokens `"Already up to date"`, `"Template version up-to-date · Skipping sync"`, `"Binary updated (template sync skipped)"`) are still distinct branches in the current update flow (no refactor has merged them into the template-sync-complete path).
4. The `--json` behavior (§B.4) is characterized before test finalization.

## §D Constraints

(Carried from spec.md §C — Tier S scope; CC policy alignment not re-implementation; simplicity ladder compliance; subagent boundary; test isolation. Not duplicated here per SSOT principle — spec.md §C is authoritative.)

## §E Self-Verification

Plan-phase self-verification — the load-bearing plan-phase outputs are:
1. The design decision (Option B recommended, Option A rejected, Option C deferred) — justified via the simplicity ladder.
2. The insertion-point identification (content token `Label: "Template sync complete"`).
3. The no-op-path enumeration (3 distinct branches by content token).

Run-phase produces the §E.1 audit-ready signal after M1 implementation + characterization test (see progress.md §E.2 / §E.3, populated by manager-develop).

## §F Milestones

### M1 — Implementation + characterization test (single milestone, Tier S scope)

**Tasks**:
1. Locate the "Template sync complete" branch in `internal/cli/update.go` by content token.
2. Append a stdout guidance line immediately after the "Template sync complete" pill. Guidance string includes literal `/hooks` token.
3. Write characterization tests in `internal/cli/update_test.go` (or a sibling `_test.go` file following the existing convention):
   - `TestUpdate_Characterize_TemplateSyncComplete_EmitsHooksGuidance` — asserts `/hooks` substring is present (AC-CS-001, AC-CS-003).
   - `TestUpdate_Characterize_AlreadyUpToDate_NoHooksGuidance` — asserts `/hooks` absent on `"Already up to date"` branch (AC-CS-002).
   - `TestUpdate_Characterize_SkippingSync_NoHooksGuidance` — asserts `/hooks` absent on `"Template version up-to-date · Skipping sync"` branch (AC-CS-002).
   - `TestUpdate_Characterize_BinaryOnly_NoHooksGuidance` — asserts `/hooks` absent on `"Binary updated (template sync skipped)"` branch (AC-CS-002).
   - `TestUpdate_NoAskUserQuestion` static guard — asserts no `AskUserQuestion` / `mcp__askuser__*` references in `update.go` non-comment lines (AC-CS-004).
4. Run full test suite: `go test ./internal/cli/...` (verify no regressions) and `go test ./...` (CLAUDE.local.md §6 HARD rule — full suite).
5. Run lint: `golangci-lint run ./internal/cli/...` and `go vet ./internal/cli/...`.

**Tier S discipline**: single milestone. If §B.1 (`"Configuration updated successfully"` branch) or §B.2 (`"Updated X → Y"` branch) characterization reveals additional insertion points, those are absorbed into M1 (not separate milestones) — Tier S = single milestone.

**Exit criteria for M1**:
- All 5 characterization/static-guard tests PASS with observed `go test` output.
- Coverage on modified branch measured by actual `go test -cover` (not table assertion per `feedback_coverage_audit_table_not_actually_run`).
- `golangci-lint run` 0 errors on `internal/cli/`.
- `go test ./...` full suite 0 failures.

No M2/M3 — Tier S scope closes at M1.

## §G Anti-Patterns

### §G.1 DO NOT use Option A (JSON-diff conditional)
JSON-diff-based conditional emission is over-engineering for a UX hint. The simplicity ladder (`moai-constitution.md § Agent Core Behaviors #4`) requires the cheapest sufficient capability; Option B is sufficient.

### §G.2 DO NOT add guidance to no-op paths
The no-op paths (`"Already up to date"`, `"Template version up-to-date · Skipping sync"`, `"Binary updated (template sync skipped)"`) MUST NOT emit guidance — they do not re-render templates, so no stale-snapshot risk exists. Emitting guidance on no-op paths would be noise and would violate AC-CS-002.

### §G.3 DO NOT call AskUserQuestion
The guidance is a stdout message. Per `internal/cli/CLAUDE.md` (C-HRA-008 / REQ-PGN-012), CLI code MUST NOT call `AskUserQuestion`, `mcp__askuser__*`, or any interactive prompt. The static-guard test (AC-CS-004) enforces this.

### §G.4 DO NOT trust line numbers
The line numbers in this plan are accurate as of 2026-07-07 but WILL drift. Run-phase MUST locate insertion points by content token (the `Label:` string value), per `feedback_line_number_drift_asymmetry`.

### §G.5 DO NOT re-implement CC's snapshot mechanism
This SPEC emits advisory guidance only. CC owns the snapshot lifecycle. Any attempt to programmatically trigger snapshot refresh (via IPC, signal, hook, or any other mechanism) is OUT OF SCOPE (see spec.md §D).

### §G.6 DO NOT skip the verify-before-author step
This SPEC exists BECAUSE the verify-before-author pipeline confirmed the defect at code level (spec.md §A.2). If run-phase discovers the defect has been silently fixed in the interim (e.g., another commit added the guidance), run-phase MUST halt and return a blocker report — do NOT implement duplicate guidance.

### §G.7 DO NOT use coverage-table assertions
Per `feedback_coverage_audit_table_not_actually_run`, coverage claims MUST cite actual `go test -cover` output, not table-row assertions.

## §H Cross-references

- **spec.md**: `SPEC-HOOK-CONFIG-SAFETY-001/spec.md` (this directory) — §A.2 grep evidence, §B GEARS requirements, §C constraints, §D Out of Scope, §E rationale.
- **acceptance.md**: `SPEC-HOOK-CONFIG-SAFETY-001/acceptance.md` (this directory) — AC-CS-001..007, Given-When-Then scenarios, edge cases.
- **progress.md**: `SPEC-HOOK-CONFIG-SAFETY-001/progress.md` (this directory) — §E.1 plan-phase audit-ready signal.
- **CC policy**: https://docs.claude.com/en/docs/claude-code/hooks § Configuration Safety.
- **Epic context**: P5 of moai-adk hooks-hardening Epic (P1 = SPEC-HOOK-SESSIONSTART-PROBE-001 completed).
- **Verification doctrine**: `.claude/rules/moai/core/verification-claim-integrity.md`.
- **Simplicity ladder**: `.claude/rules/moai/core/moai-constitution.md § Agent Core Behaviors #4`.
- **Subagent boundary**: `internal/cli/CLAUDE.md`.
- **Test isolation**: CLAUDE.local.md §6.
- **Line-drift lesson**: `feedback_line_number_drift_asymmetry`.
- **Coverage-audit lesson**: `feedback_coverage_audit_table_not_actually_run`.
- **Hypothesis-as-defect lesson**: `feedback_hypothesis_as_defect`.
