---
id: SPEC-V3R6-I18N-VALIDATOR-BUDGET-001
title: "i18n-validator TestBudget Threshold 30s → 35s"
version: "1.0.0"
status: implemented
created: 2026-05-24
updated: 2026-05-24
author: manager-spec
priority: P3
phase: "v3.0.0"
module: "scripts/i18n-validator"
lifecycle: spec-anchored
tags: "i18n, test, budget, tier-s, lcl-001-followup"
issue_number: null
tier: S
depends_on:
  - SPEC-V3R6-LEGACY-CLEANUP-001
  - SPEC-V3R3-CI-AUTONOMY-001
---

# SPEC-V3R6-I18N-VALIDATOR-BUDGET-001 — i18n-validator TestBudget Threshold 30s → 35s

## §A Background, Goal, Scope

### §A.1 Background

SPEC-V3R6-LEGACY-CLEANUP-001 (merged 2026-05-23, commit `19bc873ff`) closed with one outstanding `PASS-WITH-DEBT` acceptance criterion:

- **AC-LCL-005** (LCL-001 progress.md / CHANGELOG `[Unreleased]` line 61): `scripts/i18n-validator/TestBudget_FullRepoScanWithin30Sec` elapsed **31.18s** vs declared **30s budget** = **+4% over** (POST_PASS=84 vs PRE=85, marginal).
- The test passed on the maintainer machine on subsequent re-runs but the budget threshold is brittle for variable CI environments (shared runners, cold caches, parallel job pressure).
- LCL-001 follow-up note: "Follow-up `SPEC-V3R6-I18N-VALIDATOR-BUDGET-001` (Tier S, budget bump 30s→35s) deferred post-merge."

This SPEC clears the debt with a minimal threshold bump and coherent renaming so the test name and Japanese godoc comment match the new budget value.

### §A.2 Goal

Raise the `TestBudget_FullRepoScanWithin30Sec` wall-clock budget from `30*time.Second` to `35*time.Second` (≈ +17 % headroom) and rename the function and its godoc comment to reflect the new threshold, so the test passes deterministically on slower CI runners while remaining a meaningful regression guard against full-repo-scan latency growth.

### §A.3 Out of Scope

- **No change to validator logic** in `scripts/i18n-validator/main.go` — only the test threshold + naming/comment.
- **No change to `TestBudget_TimeoutExitOnExcess`** at line 382 (uses 1ns synthetic budget for exit-code-4 enforcement, unrelated to wall-clock budget).
- **No change to `defaultBudget` const** in `main.go` (production validator `--budget` default stays 30s; only the *test perf harness* moves to 35s for tolerance).
- **No Makefile, CI workflow, or GitHub Actions edits** (the test budget is read from the test source only).
- **No edits to archival SPEC bodies** (`SPEC-V3R3-CI-AUTONOMY-001/strategy-wave6.md`, `tasks-wave6.md`, `progress.md`, `SPEC-V3R6-LEGACY-CLEANUP-001/progress.md`) — these are completed/archived documentation references per LCL-001 §A.6 [Unwanted] retention pattern, immutable.
- **No change to synthetic corpus generation** at the 100-file loop in `TestBudget_TimeoutExitOnExcess`.

### §A.4 Scope Decision (Option 1 vs Option 2)

Two reasonable scope options were considered:

| Option | Edits | Rationale | Decision |
|--------|------:|-----------|----------|
| 1 | 1 (line 376 budget value only) | Matches user paste-ready resume "1-line" verbatim; minimal blast radius | Rejected |
| 2 | 4 (line 359 JP comment + line 360 function name + line 376 budget value + line 377 error message) | Coherent naming — `Within30Sec` becomes misleading after threshold raised to 35s; Karpathy Surgical Changes (rename to match new meaning). Line 377 error message text bundled with line 376 threshold as single logical edit unit (per plan.md §1). | **Selected** |

**Cascade evidence** (verified 2026-05-24 via `grep -rn "TestBudget_FullRepoScanWithin30Sec"`):

- **0 code invocations** (Go test functions are discovered via reflection on `func TestX(t *testing.T)` signature; no explicit call sites exist in Go source).
- **5 documentation references** across archived SPECs (4 in `SPEC-V3R3-CI-AUTONOMY-001/{strategy-wave6,tasks-wave6,progress}.md`, 1 in `SPEC-V3R6-LEGACY-CLEANUP-001/progress.md`) — all archival narrative; per §A.3 exclusion, NOT edited.

Cascade risk for function rename = **zero in code, immutable in docs**. Option 2 is safe and self-documenting.

### §A.5 Tier S Rationale

- Scope: 4 textual edits in 1 file (`scripts/i18n-validator/main_test.go` lines 359, 360, 376, 377) — 3 logical edits (comment + name + threshold/message coherence unit).
- Risk: minimal (test threshold change, no behavior change in validator logic, no API surface).
- Verification: single `go test -timeout 60s ./scripts/i18n-validator/...` invocation with elapsed-time capture.
- Per `.claude/rules/moai/workflow/spec-workflow.md` Tier S criteria: ≤ 3-file edit + ≤ 1 milestone + ≤ 5 ACs + no external service touch.

## §B Requirements (EARS Format)

### REQ-IVB-001 (Ubiquitous, mandatory)

**The system shall** enforce a `35 * time.Second` wall-clock budget at `scripts/i18n-validator/main_test.go:376` when `TestBudget_FullRepoScanWithin35Sec` (renamed) runs `runAllFilesOracle(repoRoot)` against the full repository worktree.

**Rationale**: 35s provides ≈ +17 % headroom over the observed 31.18s baseline measured on the maintainer machine during SPEC-V3R6-LEGACY-CLEANUP-001 run-phase. CI runners with shared resources may run slower; 35s absorbs that variance while remaining a meaningful upper bound (catastrophic regression would still trip the threshold at ≥ 36s, ≥ 14 % above current baseline).

### REQ-IVB-002 (Ubiquitous, mandatory)

**The system shall** name the test function `TestBudget_FullRepoScanWithin35Sec` (renamed from `TestBudget_FullRepoScanWithin30Sec`) and its godoc comment shall read `// TestBudget_FullRepoScanWithin35Sec は実際の repo で35秒以内に完了することを検証します。` (the literal `30秒` is replaced by `35秒` to match the new threshold).

**Rationale**: Self-documenting code per Karpathy Surgical Changes. The function name and comment encode the budget value; leaving them at "30Sec" / "30秒" after the threshold moves to 35s creates a naming-vs-implementation drift hazard for future readers and grep-based audits.

### REQ-IVB-003 (Event-Driven, mandatory)

**When** `go test -timeout 60s ./scripts/i18n-validator/...` is invoked locally with the renamed test, **the system shall** complete `TestBudget_FullRepoScanWithin35Sec` with elapsed time strictly less than 35 seconds and report `PASS` for the test, and the full i18n-validator test package shall not regress (POST_PASS count = PRE_PASS count for all non-budget tests).

**Rationale**: Verification anchor — the budget bump must not silently mask a real regression in `runAllFilesOracle` or other i18n-validator tests. Elapsed time must be recorded in plan-phase progress evidence (post-run).

## §C Acceptance Criteria

See [acceptance.md](./acceptance.md) for the canonical 5 AC definitions (AC-IVB-001 through AC-IVB-005) with independently verifiable commands and expected outcomes. acceptance.md is the SSOT for AC enumeration; spec.md §B-REQ rationale references AC IDs but does not duplicate AC body text.

## §D Constraints

- **D-1**: No behavior change to `scripts/i18n-validator/main.go` validator logic (the `defaultBudget = 30 * time.Second` const remains untouched; only the test harness threshold moves).
- **D-2**: No impact on `TestBudget_TimeoutExitOnExcess` (line 382): synthetic 1ns budget for exit-code-4 enforcement is orthogonal to the wall-clock budget under change.
- **D-3**: No edits to archival SPEC body files (`SPEC-V3R3-CI-AUTONOMY-001/*`, `SPEC-V3R6-LEGACY-CLEANUP-001/progress.md`) — historical narrative immutable per LCL-001 §A.6 [Unwanted] retention pattern.
- **D-4**: Tier S — plan-auditor threshold 0.80. MP-2 EARS format obligation applies (3 REQs all in SHALL form, §B.REQ-IVB-001/002/003 verified).
- **D-5**: Code comments language policy — godoc comment at line 359 is Japanese (existing file convention preserved). New code comments in this SPEC's edits shall follow the existing file's Japanese-comment convention for the budget test cluster only; English summary for spec.md/plan.md/acceptance.md follows project documentation policy (Korean conversation, English code identifiers, English EARS REQs).

## §E Risks

- **Risk-E1 (Low)**: Future regression in `runAllFilesOracle` could exceed 35s and still pass other gates, masked by the bump. **Mitigation**: AC-IVB-004 records elapsed time in progress evidence post-run; future LCL-002+ sweeps can detect drift toward the new 35s ceiling. If elapsed time exceeds 33s (94 % of new budget) in plan-phase verification, trigger optimization SPEC instead of accepting silently.
- **Risk-E2 (Negligible)**: 5 archival .md references to the old function name `TestBudget_FullRepoScanWithin30Sec` will diverge from the renamed function. **Mitigation**: documented in §A.3 as out of scope (archived SPEC immutability); a single one-line note in CHANGELOG `[Unreleased]` "### Changed" entry during sync-phase will record the rename for auditability.
- **Risk-E3 (Zero)**: Cascade grep verified 0 code invocations; Go test reflection-based discovery means function rename does not break `go test` resolution.
