---
id: SPEC-SESSIONSTART-PERF-001
title: "Session-Start Performance Durability — Progress"
version: "0.1.0"
status: in-progress
created: 2026-07-11
updated: 2026-07-11
author: manager-spec
---

# Progress — SPEC-SESSIONSTART-PERF-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-07-11
tier: L
artifacts: spec.md, plan.md, acceptance.md, design.md, research.md, progress.md
milestones: M1 (ddd, algorithmic hardening — highest priority), M2 (tdd, SPEC auto-archive), M3 (tdd, regression guard)
resolved_decisions: D1 archive-location-path → .moai/archive/specs/<year>/, D2 grace-window-default → 90 days (user-approved 2026-07-11 "proceed as proposed"); 0 open markers

## §E.2 Run-phase Evidence

### M1 — Algorithmic hardening of drift detection (cycle_type=ddd) — COMPLETE

Baseline attribution: worktree at HEAD `378613ba8481ab6944c6976b6ea09b9a4cbe4ef4` (2846 commits, 478 SPEC dirs). The pre-refactor baseline was captured from this same tree BEFORE any edit, and independently reproduced the shared-checkout baseline bit-for-bit (canonical record-set SHA256 `19a35bf3aa8dc1f00dd39139ea639f49a0778ae74f1ae8d938c37be70f89c8b1`).

**Characterization baseline (DDD PRESERVE gate, captured pre-refactor):**

| Metric | Pre-refactor | Post-refactor |
|--------|--------------|---------------|
| `moai spec drift --json` Count | 78 | 78 |
| Records emitted | 474 | 474 |
| Canonical record-set SHA256 | `19a35bf3…89c8b1` | `19a35bf3…89c8b1` (identical) |
| GitImplied distribution | terminal-exempt 38 / era-exempt 239 / completed 120 / in-progress 56 / implemented 21 | identical |

**Run-phase AC matrix (M1 scope):**

| AC | Status | Verification command | Actual output |
|----|--------|----------------------|---------------|
| AC-SSP-001 (constant subprocess count) | PASS | `go test -run TestDetectDrift_ConstantGitLogInvocations ./internal/spec/` | PASS for N=1, N=10, N=100 — `git log` invocations = 1 in every case. Empirical corroboration via PATH shim: pre-refactor 274 `git log` (548 git subprocesses total) → post-refactor **1** `git log` (3 total). |
| AC-SSP-002 (in-memory index) | PASS | `go test -run TestParseCommitRecords\|TestInMemImpliedStatus ./internal/spec/` | PASS — index built from ONE `git log <branch> --no-merges --format=%s%x1f%B%x1e` pass; `%B` (full raw message) preserves the body-dependent candidate set. |
| AC-SSP-003 (terminal + grandfather pre-filter) | PASS | `go test -run TestDetectDrift_PreFilterPerformsNoGitWork ./internal/spec/` | PASS — a fully pre-filtered corpus performs **0** `git log` invocations; sentinels `terminal-exempt` / `era-exempt` emitted verbatim with `Drifted: false`. |
| AC-SSP-004 (cache hit → zero git work) | PASS | `go test -run TestDetectDrift_CacheHitPerformsZeroGitLogWork ./internal/spec/` | PASS — warm run performs 0 `git log` invocations. PATH-shim corroboration: warm run = 1 total git subprocess (`rev-parse HEAD` cache key only), 0 `git log`. |
| AC-SSP-005a/005b (synthetic behavior preservation) | PASS | `go test -run TestDetectDrift_Characterization ./internal/spec/` | PASS — 5-category fixture (drifted / non-drifted / terminal / grandfather / combined-scope-close) + chore-skip (LSCSK-001) + word-boundary (LSGF-001). Written and confirmed GREEN on pre-refactor code first, still GREEN after. |
| AC-SSP-005c (real-corpus equivalence) [HARD] | PASS | `moai spec drift --json` (post-refactor binary, cold cache) | PASS — Count **78 == 78**; 474 records; canonical record-set SHA256 **identical** to the pre-refactor baseline. Zero per-record divergence. |
| AC-SSP-006 (recompute + persist on miss) | PASS | `go test -run TestDetectDrift_CacheHitPerformsZeroGitLogWork ./internal/spec/` | PASS — cold run computes and persists `.moai/state/drift-cache.json` keyed on HEAD SHA (gitignored per `.gitignore:265`). |
| AC-SSP-006a (`--no-cache` bypass) | PASS | `moai spec drift --count --no-cache` with a valid same-HEAD cache present | PASS — with `--no-cache`: 1 `git log` invocation (recomputed); without: 0 (cache hit). Both return 78. Unit: `TestDetectDrift_NoCacheBypassesValidCacheEntry`. |
| AC-SSP-019 (wall-clock) [HARD] | PASS | `/usr/bin/time -p moai spec drift --count` | PASS — pre-refactor **15.33s / 16.14s / 15.32s** real; post-refactor **0.23s / 0.23s / 0.22s** cold, **0.01s** warm. Target "well under 2s" met with ~67x cold / ~1500x warm improvement. |
| AC-SSP-020 (non-git graceful) | PASS | `go test -run TestDetectDrift_Characterization_NonGitEnvironment ./internal/spec/` | PASS — no panic, no error; pre-filter records still emitted, git-classified SPECs skipped. |
| AC-SSP-021 (absent specs dir) | PASS | `go test -run TestDetectDrift_Characterization_EmptySpecsDir ./internal/spec/` | PASS — empty report, Count 0, no error. |
| AC-SSP-022 (cache invalidation on new commit) | PASS | `go test -run TestDetectDrift_CacheInvalidatedWhenHeadAdvances ./internal/spec/` + live stale-key injection | PASS — a stale `head_sha` (with an injected bogus `count: 9999`) is ignored; the live binary recomputed and printed **78**, not 9999. |
| AC-SSP-024 (quality gate + coverage) | PASS | `go test ./...` / `golangci-lint run` / `go test -cover ./internal/spec/` | PASS — full suite **96 ok / 0 FAIL**; golangci-lint **0 issues**; `go vet ./...` exit 0; `internal/spec` coverage **88.9%** (≥85%). Cross-platform: `GOOS=windows` and `GOOS=linux` builds exit 0. |

**Invariants preserved (verbatim, no semantic change):** `ClassifyPRTitle`, `shouldSkipCommitTitle` (chore-skip LSCSK-001), `commitMatchesSPECID` (word-boundary LSGF-001), `combinedScopeCloseMatches` (3-gate), `deriveScopePrefix`, `isTerminalStatus` (mechanism ③), `LoadEraSignalsFromDir` + `ClassifyEra` (mechanism ④), the `gitLogWindowSize = 50` per-SPEC window, and the combined-scope FALLBACK-ONLY gate (mechanism ①). Only the commit SOURCE changed — from a per-SPEC subprocess to a shared in-memory index.

**Divergence guard (design.md §M1.3 / AP-6):** `TestInMemImpliedStatus_BodyOnlyMatchesExhaustWindow` + `TestInMemImpliedStatus_WindowBoundaryReachesCloseCommit` pin the candidate window to the FULL-MESSAGE-matched set (exactly where git applies `-N` after `--grep`). A subject-only index — the rejected alternative — would reach a close commit these tests prove must be unreachable, flipping the drift count. The pair fails loudly if that "simplification" is ever made.

**Files changed (M1):**

| File | Change |
|------|--------|
| `internal/spec/drift.go` | `DetectDrift` rewritten: `driftDeps` seam + HEAD-SHA cache short-circuit + cheap pre-filter + single-pass + in-memory walk. Added `DetectDriftFresh` / `DriftCountFresh`. Removed `resolveCombinedScopeClose` (superseded by `inMemCombinedScopeClose`; would otherwise be dead code). `getGitImpliedStatus` PRESERVED unchanged (still used by `lint.go` `StatusGitConsistencyRule`). |
| `internal/spec/drift_index.go` | NEW — `commitRecord`, `gitLogAllFullMessage` (the single pass), `parseCommitRecords`, `gitHeadSHA`, `inMemImpliedStatus` (exact two-stage replication), `inMemCombinedScopeClose`. |
| `internal/spec/drift_cache.go` | NEW — HEAD-SHA-keyed `.moai/state/drift-cache.json`, fail-open on every error path. |
| `internal/cli/spec_drift.go` | Added `--no-cache` flag + `detectDriftFn` / `driftCountFn` routing. |
| `internal/spec/drift_characterization_test.go` | NEW — DDD PRESERVE safety net (5 categories + chore-skip + word-boundary + edge cases). |
| `internal/spec/drift_seam_test.go` | NEW — O(1)-subprocess proof, cache behavior, two-stage window guard. |
| `internal/spec/drift_cache_test.go` | NEW — cache round-trip + fail-open paths. |
| `internal/spec/drift_entrypoints_test.go` | NEW — the four public entry points against a real git fixture. |

### M2 — SPEC auto-archive — NOT STARTED (out of this delegation's scope)

### M3 — Regression guard — NOT STARTED (out of this delegation's scope)

## §E.3 Run-phase Audit-Ready Signal

run_status: in-progress
milestones_complete: M1 (algorithmic hardening — REQ-SSP-001..006, REQ-SSP-006a)
milestones_pending: M2 (SPEC auto-archive — REQ-SSP-007..013), M3 (regression guard — REQ-SSP-014..018)
m1_commit_sha: pending-backfill-m1
ac_pass_count: 13 (AC-SSP-001, 002, 003, 004, 005a, 005b, 005c, 006, 006a, 019, 020, 021, 022 + AC-SSP-024 quality gate)
ac_fail_count: 0
ac_deferred_count: 11 (AC-SSP-007..018, 023 — M2 + M3 scope)
preserve_list_post_run_count: 0 violations (getGitImpliedStatus + all classification helpers preserved verbatim; no SPEC body content modified)
behavior_preservation_verified: real-corpus count 78 == 78, 474 records, record-set SHA256 identical to pre-refactor baseline
new_warnings_or_lints_introduced: 0 (golangci-lint 0 issues; go vet exit 0)
cross_platform_build: darwin PASS, windows PASS, linux PASS
coverage_internal_spec: 88.9% (threshold 85%)
full_suite: go test ./... — 96 ok / 0 FAIL
residual_risk: the HEAD-SHA cache serves a stale count while frontmatter is edited but uncommitted (HEAD unchanged). Accepted and documented per design.md §M1.4; `moai spec drift --no-cache` (REQ-SSP-006a) is the authoritative fresh path. M3's session-start time-box remains unimplemented, so a pathological repo could in principle still block session-start — M1 removes the observed cause, M3 adds the structural bound.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
