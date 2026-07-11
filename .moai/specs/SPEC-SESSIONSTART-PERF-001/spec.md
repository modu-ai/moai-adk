---
id: SPEC-SESSIONSTART-PERF-001
title: "Session-Start Performance Durability"
version: "0.1.0"
status: completed
created: 2026-07-11
updated: 2026-07-11
author: manager-spec
priority: P0
phase: "v3.1.0 performance durability"
module: "internal/spec"
lifecycle: spec-anchored
tags: "performance, session-start, drift-detection, git-subprocess, archive, regression-guard"
tier: L
era: V3R6
depends_on: []
related_specs: [SPEC-DRIFT-001, SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001, SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-002, SPEC-INTERNAL-PERF-001, SPEC-HOOK-SESSIONSTART-PROBE-001]
---

# SPEC-SESSIONSTART-PERF-001 — Session-Start Performance Durability

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-11 | manager-spec | Initial plan-phase authoring — 3-milestone Tier L SPEC targeting session-start latency caused by O(n) `git log` subprocess spawns in drift detection |

---

## §A Context and Motivation

### §A.1 The problem (measured, reproduced this session)

`moai hook session-start` blocks Claude Code session initialization by roughly **13 seconds** and the cost grows **linearly with the number of SPEC directories**. The delay is CPU-bound process-spawn thrash, not I/O wait.

**Verified baseline** (reproduced this session on git HEAD `0303e8c7`, 477 SPEC directories):

```
$ time moai spec drift --count
78
moai spec drift --count  7.99s user  4.02s system  91% cpu  13.140 total
```

The `moai spec drift` wall-clock (13.14s) matches the observed session-start block, confirming drift detection is the sole bottleneck. The task-provided independent reproduction measured real 12.78s / user 7.9s / sys 4.3s with 50,282 involuntary context switches (process-spawn thrash) — consistent with this session's measurement.

### §A.2 Root cause (verified in source at HEAD 0303e8c7)

The blocking chain is fully synchronous:

1. `.claude/settings.json` `SessionStart` hook runs `handle-session-start.sh` → `moai hook session-start` as a `type=command` hook (synchronous, `timeout: 30`). *(Verified: `.claude/settings.json` line 4-10.)*
2. `internal/hook/session_start.go` calls `detectStatusDrift(input.ProjectDir)` (session_start.go line 185) → `detectStatusDrift` (line 935) → `spec.DriftCount(projectDir)` (line 937).
3. `internal/spec/drift.go` `DriftCount` → `DetectDrift` (line 31) does `os.ReadDir(.moai/specs)` over all SPEC directories (477), and for each non-terminal, non-grandfather SPEC calls `getGitImpliedStatus(specID)` (line 101).
4. `getGitImpliedStatus` (line 178) spawns `exec.Command("git", "log", ...--grep=specID...)` (line 184) — one subprocess per SPEC. For SPECs whose frontmatter is `completed` but whose primary walk did not find a `completed` commit, `resolveCombinedScopeClose(specID)` (line 122) spawns a **second** `git log --grep=<scope-prefix>` subprocess (line 468).

The result is O(n) `git log` subprocess spawns (up to ~2× the active-SPEC count). As SPECs accumulate, every user hits a worsening wall on every session start.

### §A.3 Why this is durable (not a one-time cleanup)

Deleting or archiving SPECs is a palliative, not a fix — the O(n) subprocess pattern re-manifests as the corpus regrows. A durable fix must (a) make drift detection asymptotically independent of SPEC count, (b) bound the working dataset structurally, and (c) install a regression guard so the O(n) pattern cannot silently return.

---

## §B Requirements (GEARS)

Requirements are grouped by milestone. `<subject>` is generalized per GEARS; the drift detector, archive capability, CLI, and regression guard are the named subjects.

### §B.1 Milestone M1 — Algorithmic hardening of drift detection (primary)

- **REQ-SSP-001** (Ubiquitous): The drift detector **shall** infer git-implied status for all scanned SPECs using a bounded, constant number of `git log` subprocess invocations, independent of the SPEC count.
- **REQ-SSP-002** (Event-driven): **When** `DetectDrift` runs, the drift detector **shall** execute a single `git log` pass capturing full commit messages (subject + body), build an in-memory index, and match every SPEC ID against that index in memory while replicating the current two-stage semantics (full-message candidate match + subject word-boundary re-filter — see design.md §M1.3).
- **REQ-SSP-003** (Where — capability gate): **Where** a SPEC is in a terminal status (`completed`, `superseded`, `archived`, `rejected`) or is grandfather-protected (era-final), the drift detector **shall** exclude it from the git-checked working set before performing any `git log` work.
- **REQ-SSP-004** (While — state-driven): **While** the current git HEAD SHA is unchanged since the last drift computation, the drift detector **shall** return the cached drift result without invoking any `git` subprocess.
- **REQ-SSP-005** (Ubiquitous — HARD behavior preservation): The drift detector **shall** produce a drift `count` and per-SPEC classification (`FrontmatterStatus`, `GitImpliedStatus`, `Drifted`) semantically equivalent to the pre-refactor implementation on a representative fixture.
- **REQ-SSP-006** (Where — capability gate): **Where** the HEAD-SHA result cache is absent or stale, the drift detector **shall** recompute the result and persist it keyed on the current HEAD SHA under `.moai/state/`.
- **REQ-SSP-006a** (Where — capability gate): **Where** a user invokes `moai spec drift --no-cache`, the drift detector **shall** bypass the HEAD-SHA result cache and recompute the drift result freshly (the authoritative on-demand path — ensures an operator can always obtain a fresh count regardless of cache state).

### §B.2 Milestone M2 — SPEC auto-archive (bounds the dataset)

- **REQ-SSP-007** (Ubiquitous): The archive capability **shall** relocate terminal-status SPECs whose terminal transition is older than a configurable grace window out of `.moai/specs/` into `.moai/archive/specs/<year>/` (decided path — research.md §F D1), keeping them git-tracked and grep-discoverable.
- **REQ-SSP-008** (Event-driven): **When** a user invokes `moai spec archive`, the CLI **shall** relocate the eligible SPECs and report the moved set; **When** invoked with `--dry-run`, the CLI **shall** report the eligible set without moving any file.
- **REQ-SSP-009** (Where — capability gate): **Where** a `--grace-days N` flag (defaulting to the configured grace window, default **90 days** — research.md §F D2) applies, only SPECs whose terminal transition predates the window **shall** be eligible for archival.
- **REQ-SSP-010** (Unwanted behavior — HARD grandfather protection): The archive capability **shall not** relocate a SPEC that the era-classification logic protects, unless that SPEC independently satisfies the terminal-status + grace-window eligibility criteria.
- **REQ-SSP-011** (Unwanted behavior — critical path): The archive capability **shall not** execute inside the session-start critical path.
- **REQ-SSP-012** (Where — Template-First): **Where** the grace-window default is configured, the default **shall** live in the template source under `internal/template/templates/.moai/config/` and be mirrored to the local `.moai/config/`.
- **REQ-SSP-013** (Event-driven — trigger points): **When** a SPEC is closed during `/moai sync`, the archive capability **shall** be reachable as an on-close trigger point, in addition to the on-demand `moai spec archive` invocation.

### §B.3 Milestone M3 — Regression guard

- **REQ-SSP-014** (Ubiquitous — performance budget): The regression-guard test **shall** assert that drift detection completes under a defined time budget for a synthetic fixture of N SPECs, and **shall** fail CI on a regression that violates the budget.
- **REQ-SSP-015** (Event-driven — time-box): **When** the session-start drift computation exceeds a bounded deadline, the session-start handler **shall** skip the computation and emit the existing advisory ("Run `moai spec drift` for details.") instead of blocking.
- **REQ-SSP-016** (Ubiquitous — codified principle): The codified rule **shall** state that session-start / Stop advisory checks must be cached, asynchronous, or on-demand — never unbounded-blocking.
- **REQ-SSP-017** (Where — Template-First): **Where** the codified principle is added to a rule file under `.claude/rules/`, it **shall** be edited in `internal/template/templates/.claude/rules/` first, then mirrored to the local `.claude/rules/`.
- **REQ-SSP-018** (Ubiquitous — no hardcoding): The implementation **shall** extract new tunable thresholds (grace-window default, performance budget, time-box deadline) to `config/defaults.go` or, for environment overrides, to `internal/config/envkeys.go`, with no hardcoded literals embedded in business logic.

---

## §C Exclusions

This section satisfies the `OutOfScopeRule` lint (`MissingExclusions`). Each excluded item is a routing decision — what this SPEC does NOT build and where that concern belongs instead.

### Out of Scope — Web-tooling documentation references

- This SPEC does NOT touch `internal/template/templates/` documentation references to `web_search_prime` / `web_reader`. That is a separate follow-up explicitly declared out of scope by the task brief.

### Out of Scope — Drift classification semantics changes

- This SPEC does NOT change WHAT drift means or HOW a commit subject classifies to a lifecycle status. The `ClassifyPRTitle`, `shouldSkipCommitTitle`, `commitMatchesSPECID`, word-boundary (LSGF-001), and chore-skip (LSCSK-001) semantics are PRESERVED verbatim. Any change to classification semantics belongs to the drift-convention SPEC line (`SPEC-V3R6-DRIFT-*`), not here.

### Out of Scope — Era classification and grandfather policy changes

- This SPEC REUSES the existing era-classification helpers (`LoadEraSignalsFromDir`, `ClassifyEra`, `EraFinal`) read-only. It does NOT alter the era heuristic table or grandfather clause. Changes to era policy belong to `SPEC-V3R6-LIFECYCLE-SYNC-GATE-001`'s SSOT (`lifecycle-sync-gate.md`).

### Out of Scope — Making session-start hook asynchronous

- This SPEC does NOT convert the `SessionStart` hook from `type=command` (synchronous) to an async hook, nor does it change the hook timeout from 30s. The fix is algorithmic (make the work fast + time-boxed), not a hook-execution-model change. A hook-execution-model change would belong to a hook-runtime SPEC.

### Out of Scope — Retroactive archival of the current 477-SPEC backlog

- This SPEC delivers the archive CAPABILITY and its eligibility rules. It does NOT mandate a one-time bulk archival of the existing backlog as an acceptance criterion; the operator runs `moai spec archive` (with `--dry-run` first) at their discretion. A bulk-archival operation is a maintainer runbook step, not a code deliverable.

### Out of Scope — CI wall-clock optimization of the full test suite

- This SPEC adds ONE performance-budget regression test. It does NOT undertake a broad CI-time optimization pass. Broad CI optimization belongs to the CI-baseline SPEC line.

---

## §D Acceptance Criteria Reference

The full Given-When-Then acceptance criteria (AC-SSP-001 .. AC-SSP-024) are enumerated in `acceptance.md`. Each REQ-SSP-NNN maps to at least one AC. The behavior-preservation criterion (REQ-SSP-005) is verified against a representative fixture via characterization tests captured BEFORE the M1 refactor.

---

## §E Success Definition

This SPEC is successful when:

1. `moai spec drift` and session-start drift detection complete in well under 2 seconds at the current corpus size, with subprocess count independent of SPEC count (M1).
2. Terminal SPECs past the grace window can be archived out of the active scan set via `moai spec archive`, respecting grandfather protection (M2).
3. A CI performance-budget test fails if the O(n) subprocess pattern regresses, and session-start is time-boxed so drift computation can never block the critical path unboundedly (M3).
4. The "advisory checks must be cached/async/on-demand, never unbounded-blocking" principle is codified in a rule so future advisory checks inherit the constraint.
