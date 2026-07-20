---
id: SPEC-SESSIONSTART-PERF-001
title: "Session-Start Performance Durability — Acceptance Criteria"
version: "0.1.0"
status: completed
created: 2026-07-11
updated: 2026-07-11
author: manager-spec
---

# Acceptance Criteria — SPEC-SESSIONSTART-PERF-001

Every REQ-SSP-NNN maps to at least one AC. ACs are Given-When-Then and observable (test output, wall-clock, grep result, file existence). Sub-IDs (e.g. AC-SSP-005a/005b) denote paired sub-criteria of one logical AC.

## §D.1 Milestone M1 — Algorithmic hardening

### AC-SSP-001 — Constant subprocess count (REQ-SSP-001, REQ-SSP-002)
- **Given** a repository with N SPEC directories and an injected/observable git command runner,
- **When** `DetectDrift` runs,
- **Then** the number of `git log` subprocess invocations is a small constant (independent of N) — verified by a test that counts git invocations through a seam (injected runner or command counter) and asserts the count does NOT grow with N.

### AC-SSP-002 — In-memory index build (REQ-SSP-002)
- **Given** a single full-message `git log --no-merges` pass output (subject + body captured),
- **When** the drift detector builds its index,
- **Then** every SPEC ID whose full commit message (subject or body) contains it is resolvable from the in-memory index as a candidate (full-message substring match, then subject re-filter via `commitMatchesSPECID`; scope-prefix indexed via `deriveScopePrefix`) without a further subprocess — replicating the current two-stage `git log --grep` + subject-refilter semantics (design.md §M1.3).

### AC-SSP-003 — Terminal + grandfather pre-filter (REQ-SSP-003)
- **Given** a SPEC in terminal status (`completed`/`superseded`/`archived`/`rejected`) or grandfather-protected (era-final),
- **When** `DetectDrift` scans it,
- **Then** the SPEC is excluded from the git-checked working set (no per-SPEC git work is performed for it), AND its record is still emitted with `Drifted: false` and the ACTUAL sentinel `GitImpliedStatus` the current code uses — `"terminal-exempt"` (drift.go:75) for terminal-status SPECs, `"era-exempt"` (drift.go:93) for era-final SPECs — NOT an empty string (the existing record-emission contract at drift.go:71-98 is preserved verbatim).

### AC-SSP-004 — HEAD-SHA cache hit skips git work (REQ-SSP-004)
- **Given** a prior drift computation persisted under `.moai/state/` keyed on HEAD SHA X,
- **When** `DetectDrift` runs again while HEAD is still X,
- **Then** the cached drift count is returned and zero `git` subprocesses are invoked — verified via the injected-runner git-invocation counter reading 0.

### AC-SSP-005 — Behavior preservation on representative fixture (REQ-SSP-005) [HARD]
- **AC-SSP-005a Given** the characterization baseline (`DetectDrift` output captured on the representative fixture BEFORE the M1 refactor),
- **When** the refactored `DetectDrift` runs on the same fixture,
- **Then** the drift `count` and every per-SPEC record (`SPECID`, `FrontmatterStatus`, `GitImpliedStatus`, `Drifted`) are identical to the baseline.
- **AC-SSP-005b Given** the fixture includes a drifted SPEC, a non-drifted SPEC, a terminal SPEC, a grandfather SPEC, and a combined-scope-close SPEC,
- **When** classification runs,
- **Then** each category classifies identically pre- and post-refactor (chore-skip LSCSK-001, word-boundary LSGF-001, and combined-scope 3-gate behavior all preserved).
- **AC-SSP-005c Given** the REAL `.moai/specs` corpus baseline captured in plan.md §C BEFORE the M1 refactor (currently `count == 78` at HEAD `0303e8c7`, per `moai spec drift --json`), **When** the refactored `DetectDrift` runs against the same real corpus at the same HEAD, **Then** the drift `count` equals that baseline (78) AND every per-SPEC record (`{SPECID, FrontmatterStatus, GitImpliedStatus, Drifted}`) is equivalent to the baseline — the empirical body-dependency backstop the synthetic fixture (AC-SSP-005a) cannot reach (guards against the two-stage full-message-vs-subject divergence, design.md §M1.3). Bound to REQ-SSP-005 [HARD].

### AC-SSP-006 — Cache recompute + persist on miss (REQ-SSP-006)
- **Given** no cache file, or a cache keyed on a stale HEAD SHA,
- **When** `DetectDrift` runs,
- **Then** the result is recomputed and persisted keyed on the CURRENT HEAD SHA under `.moai/state/`, and a subsequent same-HEAD run is served from cache (AC-SSP-004).

### AC-SSP-006a — On-demand cache bypass (REQ-SSP-006a)
- **Given** a valid HEAD-SHA cache entry exists for the current HEAD,
- **When** a user invokes `moai spec drift --no-cache`,
- **Then** the drift detector ignores the cache and recomputes freshly (the full single-pass path runs; the cache is neither read as authoritative nor short-circuited) — verified by a test asserting the compute path executes despite a present same-HEAD cache.

### AC-SSP-019 — Wall-clock improvement (REQ-SSP-001, aggregate) [HARD]
- **Given** the current corpus (≈477 SPECs, baseline `moai spec drift --count` ≈ 13.1s at HEAD 0303e8c7),
- **When** `moai spec drift --count` runs after M1 (cold cache, single pass),
- **Then** wall-clock is well under 2 seconds — an order-of-magnitude improvement — with the before/after measurements cited verbatim in progress.md §E.2.

## §D.2 Milestone M2 — SPEC auto-archive

### AC-SSP-007 — Archive relocates and keeps git-tracked (REQ-SSP-007)
- **Given** a terminal-status SPEC older than the grace window,
- **When** `moai spec archive` runs (non-dry-run),
- **Then** the SPEC directory is moved to `.moai/archive/specs/<year>/SPEC-...` (decided path), remains git-tracked (appears in `git ls-files`), and its content is still grep-discoverable.

### AC-SSP-008 — Dry-run vs apply (REQ-SSP-008)
- **AC-SSP-008a Given** eligible SPECs, **When** `moai spec archive --dry-run` runs, **Then** the eligible set is reported and NO file is moved (`git status` clean).
- **AC-SSP-008b Given** eligible SPECs, **When** `moai spec archive` runs without `--dry-run`, **Then** the eligible SPECs are moved and the moved set is reported.

### AC-SSP-009 — Grace window gates eligibility (REQ-SSP-009)
- **AC-SSP-009a Given** two terminal SPECs, one whose terminal transition predates `--grace-days N` and one within the window, **When** `moai spec archive --grace-days N --dry-run` runs, **Then** only the older SPEC is reported eligible; the within-window SPEC is excluded.
- **AC-SSP-009b Given** no `--grace-days` flag is passed, **When** `moai spec archive --dry-run` runs, **Then** the effective grace window defaults to the configured value (default **90 days** from Template-First config / `config/defaults.go` fallback) — verified by a test asserting the flag-absent default equals 90.

### AC-SSP-010 — Grandfather protection guard (REQ-SSP-010) [HARD]
- **Given** an era-final (grandfather-protected) SPEC that does NOT independently satisfy terminal-status + grace-window,
- **When** `moai spec archive` runs,
- **Then** that SPEC is NOT moved. **And** an era-final SPEC that DOES independently satisfy terminal+grace IS eligible — grandfather status neither forces nor forbids archival on its own.

### AC-SSP-011 — Archive absent from session-start path (REQ-SSP-011) [HARD]
- **Given** the session-start handler source,
- **When** `grep -n 'archive' internal/hook/session_start.go` runs,
- **Then** there is no archive invocation in the session-start critical path (0 archive call sites).

### AC-SSP-012 — Grace-window default is Template-First (REQ-SSP-012)
- **Given** the grace-window default config,
- **When** the repository is inspected,
- **Then** the default exists in `internal/template/templates/.moai/config/` AND is mirrored to local `.moai/config/` (both present; template is the source of truth).

### AC-SSP-013 — On-close trigger reachability (REQ-SSP-013)
- **Given** a SPEC closed during `/moai sync`,
- **When** the close path is exercised,
- **Then** the archive eligibility+relocation entry point is reachable as an on-close trigger (in addition to the on-demand `moai spec archive` command), and this trigger is NOT the session-start path.

## §D.3 Milestone M3 — Regression guard

### AC-SSP-014 — Performance-budget regression test (REQ-SSP-014) [HARD]
- **Given** a synthetic fixture of N=500 SPEC directories in `t.TempDir()`,
- **When** the perf-budget test runs drift detection,
- **Then** it asserts completion under the configured budget (default 2s) and FAILS if exceeded — the assertion is a real timed run, not a mocked claim.

### AC-SSP-015 — Session-start time-box (REQ-SSP-015) [HARD]
- **Given** a drift computation that would exceed the bounded deadline,
- **When** the session-start handler runs it under `context.WithTimeout`,
- **Then** the handler skips the computation on deadline exceed and emits the existing advisory ("Run `moai spec drift` for details.") instead of blocking — verified by a test injecting a slow drift function and asserting the advisory-emitted, non-blocking outcome.

### AC-SSP-016 — Principle codified (REQ-SSP-016)
- **Given** the codified rule file,
- **When** it is read,
- **Then** it states that session-start / Stop advisory checks must be cached, asynchronous, or on-demand — never unbounded-blocking (grep for the principle keywords returns the codified statement).

### AC-SSP-017 — Rule codification is Template-First (REQ-SSP-017)
- **Given** the codified principle,
- **When** the repository is inspected,
- **Then** the rule text exists in `internal/template/templates/.claude/rules/` AND is mirrored to local `.claude/rules/` (template is the source).

### AC-SSP-018 — No hardcoded thresholds (REQ-SSP-018)
- **Given** the new tunable thresholds (grace-window default, perf budget, time-box deadline),
- **When** the source is inspected,
- **Then** each is defined in `config/defaults.go` (or `internal/config/envkeys.go` for env overrides) and referenced by symbol — no inline literal of these values in business logic.

## §D.4 Cross-cutting / edge cases

### AC-SSP-020 — Non-git environment graceful no-op (REQ-SSP-005)
- **Given** a project checkout without a git directory (e.g. `t.TempDir()` fixture),
- **When** drift detection runs,
- **Then** it degrades gracefully (returns 0 / empty, no panic, no error-severity escalation) — preserving current `DriftCount` error-swallowing behavior in the session-start caller.

### AC-SSP-021 — Empty / absent specs directory (REQ-SSP-005)
- **Given** no `.moai/specs` directory,
- **When** `DetectDrift` runs,
- **Then** it returns an empty report with count 0 (preserving drift.go:35-37 behavior).

### AC-SSP-022 — Cache invalidation on new commit (REQ-SSP-004 / REQ-SSP-006)
- **Given** a cached result at HEAD X,
- **When** a new commit advances HEAD to Y and drift runs,
- **Then** the cache is treated as stale and recomputed against Y (no stale count served).

### AC-SSP-023 — Archive round-trip discoverability
- **Given** a SPEC archived by `moai spec archive` (now under `.moai/archive/specs/<year>/`),
- **When** an operator greps the repository for the SPEC ID,
- **Then** the archived SPEC is still found (moved, not deleted) — confirming "keep git-tracked and grep-discoverable" (REQ-SSP-007).

### AC-SSP-024 — Quality gate + coverage (REQ-SSP-018 — implementation quality; cross-cutting gate over M1-M3)
- **Given** the completed implementation,
- **When** the full suite runs,
- **Then** `go test ./...` passes, `golangci-lint run` is clean, and coverage is ≥ 85% package-level (≥ 90% for `internal/cli` and `internal/hook` touched files).

## §D.5 Definition of Done

- [ ] All 18 REQ-SSP have at least one passing AC.
- [ ] AC-SSP-005 (behavior preservation) proven against the pre-refactor characterization baseline.
- [ ] AC-SSP-019 wall-clock improvement measured and cited verbatim (before/after).
- [ ] AC-SSP-014 perf-budget test wired into `go test ./...` (fails CI on regression).
- [ ] AC-SSP-011 grep guard confirms archive is not in the session-start path.
- [ ] Template-First verified for M2 config (AC-SSP-012) and M3 rule (AC-SSP-017).
- [ ] No hardcoded thresholds (AC-SSP-018).
- [ ] `go test ./...` green, `golangci-lint run` clean, coverage targets met (AC-SSP-024).
- [x] Prior open decisions resolved (user-approved 2026-07-11): archive location → `.moai/archive/specs/<year>/`; grace-window default → 90 days. 0 open markers remain.
