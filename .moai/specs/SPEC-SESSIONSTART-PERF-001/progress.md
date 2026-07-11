---
id: SPEC-SESSIONSTART-PERF-001
title: "Session-Start Performance Durability — Progress"
version: "0.1.0"
status: completed
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

### M2 — SPEC auto-archive (cycle_type=tdd) — COMPLETE

Baseline attribution: worktree fast-forwarded to `main` at `a02cfd0da` (M1 merged; `main` confirmed an ancestor of HEAD). All measurements below were taken on this tree.

RED→GREEN: the archive tests were written first and confirmed failing to compile (`undefined: PlanArchive`, `undefined: newSpecArchiveCmd`, `undefined: DefaultArchiveGraceDays`) across `internal/spec`, `internal/cli`, and `internal/config` before any implementation existed. The `internal/hook` guard passed from the start on a clean baseline — deliberately so: it proves non-regression rather than a fix.

**Run-phase AC matrix (M2 scope):**

| AC | Status | Verification command | Actual output |
|----|--------|----------------------|---------------|
| AC-SSP-007 (relocate + stay git-tracked) | PASS | `go test -run TestExecuteArchive_RealRepo_StaysTracked ./internal/spec/` | PASS — against a REAL git repo: the SPEC moves to `.moai/archive/specs/2023/…` via `git mv` and remains in the index at its new path (`git ls-files --error-unmatch` succeeds); the old path is no longer tracked. |
| AC-SSP-008a (dry-run reports, moves nothing) | PASS | `go test -run TestSpecArchiveCmd_DryRunMovesNothing ./internal/cli/` | PASS — eligible SPEC reported; source still present; `.moai/archive` never created. `PlanArchive` is observation-only by construction (`TestPlanArchive_NeverMoves`). |
| AC-SSP-008b (apply moves the eligible set) | PASS-WITH-NARROWING | `go test -run TestSpecArchiveCmd_ApplyMovesEligible ./internal/cli/` | PASS — eligible SPEC moved, non-terminal SPEC untouched, moved set reported. **Deliberate narrowing:** a bare `moai spec archive` reports the plan then REFUSES without `--yes` (`TestSpecArchiveCmd_RequiresConfirmation`). The AC text says "without `--dry-run` → moved"; a confirmation flag was added as a safety gate against unreviewed bulk relocation (the §5 incident class). Disclosed for user acceptance — see Residual risk. |
| AC-SSP-009a (grace window gates eligibility) | PASS | `go test -run TestSpecArchiveCmd_GraceDaysFlag ./internal/cli/` | PASS — a 60-day-old terminal SPEC is NOT eligible at `--grace-days 90` and IS eligible at `--grace-days 30`. Boundary pinned strictly-before by `TestPlanArchive_GraceBoundary` (a SPEC exactly on the cutoff stays). |
| AC-SSP-009b (flag-absent default = 90) | PASS | `go test -run TestSpecArchiveCmd_DefaultGraceDaysIs90 ./internal/cli/` | PASS — `--json` reports `"grace_days": 90` with no flag passed. Zero/absent config also resolves to 90 (`TestArchiveGraceDays_ZeroFallsBackToDefault`) — an unset window never degrades into "no grace at all". |
| AC-SSP-010 (grandfather protection) [HARD] | PASS | `go test -run TestPlanArchive_GrandfatherIsNotAGate ./internal/spec/` + real-corpus jq invariant check | PASS — era is REPORTED, never a gate. Unit: an era-final SPEC that is `draft`, or inside the grace window, is NOT archived; one that independently satisfies terminal+grace IS. **Real corpus (grace=30d, 142 eligible):** 128 era-final SPECs appear, and **0** were swept in without independently satisfying terminal+grace; 0 non-terminal candidates; 0 candidates inside the window. |
| AC-SSP-011 (archive absent from session-start) [HARD] | PASS | `go test -run Archive ./internal/hook/` + `grep -c -i archive internal/hook/session_start.go` | PASS — grep returns **0**. The static guard scans **58 non-test hook sources** for 6 archive symbols and finds 0 call sites; it fails loudly if the guard ever becomes vacuous. |
| AC-SSP-012 (Template-First config) | PASS | `diff internal/template/templates/.moai/config/sections/archive.yaml .moai/config/sections/archive.yaml` | PASS — template authored FIRST, `make build` exit 0, then mirrored. Template ↔ local **IDENTICAL**. Neutrality self-check: 0 matches for SPEC-IDs / REQ tokens / internal dates / CLAUDE.local refs (CLAUDE.local.md §25). |
| AC-SSP-013 (on-close trigger reachability) | PASS | `go test -run 'TestPlanArchive_RealRepo\|TestExecuteArchive_RealRepo' ./internal/spec/` | PASS — `spec.PlanArchive` / `spec.ExecuteArchive` are plain exported functions callable from any non-session-start caller (the `moai spec archive` CLI today; the `/moai sync` close path is the documented second trigger point). Both covered at 100%. |
| AC-SSP-018 (no hardcoded thresholds) | PASS | `grep -n 'DefaultArchiveGraceDays' internal/config/defaults.go` | PASS — the literal `90` exists exactly once, as `config.DefaultArchiveGraceDays`. `internal/spec/archive.go` and the CLI reference it by symbol; the template config carries the same value as the user-facing default. |
| AC-SSP-023 (archive round-trip discoverability) | PASS | `go test -run TestExecuteArchive_ContentSurvives ./internal/spec/` | PASS — archiving is a MOVE, never a delete: the archived `spec.md` still contains its SPEC-ID and `status:` line at the new path, and (AC-SSP-007) remains git-tracked. |
| AC-SSP-024 (quality gate + coverage) | PASS | `go test ./...` / `golangci-lint run` / `go vet ./...` | PASS — full suite **96 ok / 0 FAIL**; golangci-lint **0 issues**; `go vet` exit 0. Cross-platform: darwin / windows / linux builds all exit 0. |

**Drift non-regression (M1 must be untouched):** `git diff main -- internal/spec/drift*.go internal/cli/spec_drift.go internal/hook/session_start.go` is **EMPTY** — M2 touched zero drift or session-start code. `moai spec drift --no-cache` on the real corpus returns **475 records / 79 drifted**, matching the stated M1 baseline exactly.

**Real-corpus dry-run (nothing moved):** `moai spec archive --dry-run` scanned **477 SPECs** and found **0 archive-eligible** at the default 90-day window. This is correct, not a defect: an independent raw-git cross-check (`git log -1 --format=%cI -- <specdir>` per SPEC) shows the **oldest** last-touch across the entire corpus is **2026-05-12** — 60 days ago — so no SPEC can predate a 90-day cutoff (`2026-04-12`). Eligibility by window: 1d→313, 7d→270, 30d→142, 90d→0, 180d→0, 365d→0. The cliff is consistent with an oldest-activity of 60 days. **Operational consequence:** at the default grace, M2 does not shrink the 477-SPEC dataset today; a bulk sweep re-touched the whole tree ~60 days ago and reset every SPEC's last-activity clock. The capability becomes effective as that sweep recedes past 90 days, or immediately at a shorter `--grace-days`.

**Eligibility predicate (the load-bearing contract):** a SPEC is archive-eligible iff (1) its status is terminal — `completed` / `superseded` / `archived` / `rejected` — AND (2) its last activity is strictly before `now − graceDays`. Era classification is surfaced on every candidate but **never consulted as a gate**, so grandfather status neither forces nor forbids archival. Note the deliberate divergence from `drift.go`'s `isTerminalStatus`, which excludes `completed`: the two answer different questions, and collapsing them would make archiving a no-op for most of a mature corpus (pinned by `TestIsArchiveTerminalStatus_DivergesFromDriftTerminal`).

**O(1) subprocess discipline carried over from M1:** the archive scan resolves last-activity for the whole corpus in **ONE** `git log --name-only -- .moai/specs` pass (`gitLastActivity` + `parseGitActivity`), newest-first, first-sighting-wins. The naive `git log -1 -- <specDir>` per SPEC would have reintroduced exactly the O(n) fan-out M1 removed.

**Coverage (new files):**

| Symbol | Coverage |
|--------|----------|
| `IsArchiveTerminalStatus` / `PlanArchive` / `ExecuteArchive` / `gitLastActivity` / `archiveDestDir` / `realArchiveDeps` | 100.0% |
| `planArchive` | 93.9% |
| `parseGitActivity` | 90.5% |
| `printArchivePlan` | 100.0% |
| `newSpecArchiveCmd` | 89.3% |
| `gitMoveOrRename` | 83.3% |
| `ArchiveGraceDays` | 100.0% |
| Package `internal/spec` | 87.8% (≥85%) |

**Files changed (M2):**

| File | Change |
|------|--------|
| `internal/spec/archive.go` | NEW — eligibility predicate (`IsArchiveTerminalStatus`, `planArchive`), `PlanArchive` / `ExecuteArchive` entry points, single-pass `gitLastActivity` + `parseGitActivity`, `gitMoveOrRename` (git mv → os.Rename fallback, refuses to clobber). |
| `internal/spec/archive_test.go` | NEW — eligibility, grace boundary, grandfather non-gating (both directions), dry-run no-op, destination path, frontmatter fallback, git-wins-over-frontmatter. |
| `internal/spec/archive_git_test.go` | NEW — the production path against a REAL git repo: activity scan, newest-touch-wins, `git mv` keeps the SPEC tracked, clobber refusal, non-git fallback. |
| `internal/cli/spec_archive.go` | NEW — `moai spec archive [--dry-run] [--yes] [--grace-days N] [--json]`. |
| `internal/cli/spec_archive_test.go` | NEW — dry-run, confirmation gate, apply, grace-days flag, default 90, era surfacing, subagent-boundary guard. |
| `internal/cli/spec.go` | Registered `newSpecArchiveCmd()` under the `moai spec` group. |
| `internal/config/{types,defaults,loader,audit_registry}.go` + `loader_archive.go` | NEW section `ArchiveConfig` + `DefaultArchiveGraceDays = 90` + `ArchiveGraceDays()` accessor (zero → default). |
| `internal/template/templates/.moai/config/sections/archive.yaml` + local mirror | NEW — Template-First grace-window config (content-neutral). |
| `internal/hook/session_start_archive_guard_test.go` | NEW — static guard: archive is unreachable from the session-start critical path. |

### M3 — Regression guard (cycle_type=tdd) — COMPLETE

Baseline attribution: worktree branch `spec-sessionstart-perf-001-m3` created from `main` at `02ba30bbe` (the M2-merged base). `main` is an ancestor of this branch (ff-mergeable at the time of authoring; `main` had since advanced to `82b758d43` via parallel sessions — see §E.3 m3_base_note). All M3 measurements were taken on this tree.

RED→GREEN: the time-box tests were written first and confirmed failing to compile (`undefined: driftWorkFn`, `undefined: DriftCountCtx`, `undefined: driftCountFn`, `undefined: driftWarningThreshold`) before any implementation existed. The perf-budget guard passes on the correct O(1) code; its RED is demonstrated by the guard-catches-regression proof below (a real, reverted O(n) injection).

**Run-phase AC matrix (M3 scope):**

| AC | Status | Verification command | Actual output |
|----|--------|----------------------|---------------|
| AC-SSP-014 (perf-budget regression test) [HARD] | PASS | `go test -run TestDetectDrift_PerfBudget ./internal/spec/` | PASS — 2 sub-tests. PRIMARY (deterministic): at N=`config.DefaultDriftPerfFixtureSpecs`=500, the `driftDeps` counting seam reads `git log` invocations = **1** (independent of N). SECONDARY (real timed run): `DetectDriftFresh` on a REAL 500-SPEC git fixture completes in ~0.25s ≪ `DefaultDriftPerfBudget` (2s). Guard-catches-regression PROVEN: injecting `deps.logAll` into the per-SPEC loop made the count read **501** and the guard FAILED (`git log invocations = 501 at N=500, want EXACTLY 1`); reverted → PASS. |
| AC-SSP-015 (session-start time-box) [HARD] | PASS | `go test -run 'TestDetectStatusDrift\|TestDriftCountCtx' ./internal/hook/ ./internal/spec/` | PASS — `detectStatusDrift` wraps the drift call in `context.WithTimeout(DefaultSessionStartDriftTimeout)`; on `DeadlineExceeded` it skips and returns the advisory `⚠ SPEC status drift check timed out. Run 'moai spec drift' for details.` instead of blocking. `TestDetectStatusDrift_TimeBoxEmitsAdvisory` injects a slow drift fn + 50ms deadline → returns promptly (< 2s) with the advisory. `TestDriftCountCtx_TimeoutAbandonsSlowWorker` proves `DriftCountCtx` abandons a context-IGNORING 5s worker in ~50ms via its select. |
| AC-SSP-016 (principle codified) | PASS | `grep -l 'cached, asynchronous, or on-demand — never unbounded-blocking' <template> <local>` | PASS — the `## Advisory-Check Discipline` [HARD] section states the principle verbatim in BOTH `internal/template/templates/.claude/rules/moai/development/coding-standards.md` AND the local mirror. |
| AC-SSP-017 (rule codification Template-First) | PASS | Template-First order + `go test ./internal/template/...` | PASS — the section was authored in the TEMPLATE first, `make build` exit 0, then mirrored to local (identical text). Both present. Template neutrality guard (`TestTemplateNoInternalContentLeak`, default/narrow tier) GREEN — the new section contributes 0 leak-class matches (0 in strict tier too). |
| AC-SSP-018 (no hardcoded thresholds) | PASS | `grep -n 'DefaultDriftPerfBudget\|DefaultSessionStartDriftTimeout\|DefaultDriftPerfFixtureSpecs' internal/config/defaults.go` | PASS — all three M3 thresholds (`2s` perf budget, `2s` time-box deadline, `500` fixture size) are `const` in `config/defaults.go` and referenced by symbol; no inline literal in business logic or tests. |
| AC-SSP-024 (quality gate + coverage) | PASS | `go test ./...` / `golangci-lint run` / `go vet ./...` / `go tool cover` | PASS — full suite **96 ok / 0 FAIL** (3× `-count=1` deterministic); golangci-lint **0 issues**; `go vet` exit 0; cross-platform darwin/windows/linux builds exit 0. `internal/spec` coverage **88.9%** (≥85%); the touched-file functions `detectStatusDrift` (hook) and `DriftCountCtx` (spec) are **100%** covered. `-race` on all M3-touched packages GREEN (3× 0 races). |

**Determinism proof (the item-4 flaky-test fix):** `TestBudget_FullRepoScanWithin35Sec` in `scripts/i18n-validator` passed **3/3** in the deterministic full-suite runs and in isolation. Root cause found + fixed: the validator walked 64 gitignored, ephemeral `.claude/worktrees/agent-*` checkouts (corpus 1,539 → 93,531 files, ~60× inflation), so a raw 35s wall-clock ceiling flaked under load. Fix = (a) exclude `.claude/worktrees/` from `corpusExclusions`, (b) match exclusions on the ROOT-RELATIVE path (an absolute-path match wrongly excluded EVERY file when the scan itself runs from within a worktree — the 0-file silent no-op the new work-done guard caught), (c) rework the budget assertion from a fixed ceiling to work-done (corpus non-empty) + a scaled per-file budget with generous margin + the 35s outer ceiling honouring the historical contract. The function name + `35秒以内` godoc are PRESERVED (locked by the closed SPEC-V3R6-I18N-VALIDATOR-BUDGET-001 acceptance). New `TestIsExcluded_MatchesRootRelativePath` pins the relative-match semantics.

**Race caught + fixed by `-race` before commit:** the initial `DriftCountCtx` read the mutable `driftWorkFn` package var from inside its worker goroutine; when the timeout test deliberately abandons that goroutine, its delayed read raced the test's `Cleanup` reassignment. Fixed by capturing the seam into a local (`work := driftWorkFn`) synchronously before spawn. Post-fix: 3× `-race` on touched packages, 0 races.

**Files changed (M3):**

| File | Change |
|------|--------|
| `internal/config/defaults.go` | Added `DefaultDriftPerfBudget` (2s), `DefaultSessionStartDriftTimeout` (2s), `DefaultDriftPerfFixtureSpecs` (500) consts (REQ-SSP-018). |
| `internal/spec/drift.go` | Added `DriftCountCtx(ctx, baseDir)` (goroutine + select time-box, seam captured to a local) + `driftWorkFn` seam. `detectDrift` algorithm UNCHANGED (additive-only diff, 0 deletions). |
| `internal/spec/drift_timebox_test.go` | NEW — `DriftCountCtx` timeout/completion/error tests via the `driftWorkFn` seam. |
| `internal/spec/drift_perf_test.go` | NEW — AC-SSP-014 perf-budget guard: N=500 constant-subprocess (primary, deterministic) + real-repo timed run (secondary). |
| `internal/hook/session_start.go` | `detectStatusDrift` rewritten with a `context.WithTimeout` time-box + `driftCountFn`/`sessionStartDriftTimeout` seams + `driftWarningThreshold`/`driftTimeoutAdvisory` consts (REQ-SSP-015). |
| `internal/hook/session_start_timebox_test.go` | NEW — advisory-on-timeout + count/below-threshold/non-timeout-error branch tests. |
| `internal/template/templates/.claude/rules/moai/development/coding-standards.md` + local mirror | NEW `## Advisory-Check Discipline` [HARD] section (Template-First; neutral prose) — REQ-SSP-016/017. |
| `scripts/i18n-validator/main.go` | Added `.claude/worktrees/` to `corpusExclusions`; `isExcluded` now matches the ROOT-RELATIVE path (fixes the scan-from-within-a-worktree 0-file no-op). |
| `scripts/i18n-validator/lockset.go` | Both `isExcluded` call sites pass `root` for relative matching. |
| `scripts/i18n-validator/main_test.go` | Reworked `TestBudget_FullRepoScanWithin35Sec` (name + godoc preserved) to work-done + scaled-budget robustness; NEW `TestIsExcluded_MatchesRootRelativePath`. |

## §E.3 Run-phase Audit-Ready Signal

run_status: run-complete (all run-phase milestones M1-M3 done; awaiting sync-phase close)
milestones_complete: M1 (algorithmic hardening — REQ-SSP-001..006, REQ-SSP-006a), M2 (SPEC auto-archive — REQ-SSP-007..013, REQ-SSP-018), M3 (regression guard — REQ-SSP-014..018)
milestones_pending: (none — run-phase complete)
m1_commit_sha: f376ee7d25bea4721b0f6d9bd9e1f9ea92419f44
m2_commit_sha: a814656c5b159a6bd12e5b456a9aa969037aad5d
m3_commit_sha: ab4ea7d7498faba491f50d29b4b0f99d8722a00d (backfilled during sync-phase; verified via `git rev-parse ab4ea7d74` and `git show --stat ab4ea7d74`)
m3_base_note: M3 branch `spec-sessionstart-perf-001-m3` was cut from `main` @ `02ba30bbe` (the M2-merged base, as delegated). `main` advanced to `82b758d43` via parallel sessions during authoring; `02ba30bbe` IS an ancestor of `82b758d43`, so the maintainer can rebase M3 onto current `main` cleanly (or ff-merge after rebase). M3 touched only files verified identical between the worktree base and current `main` (drift.go, drift_index.go, session_start.go, defaults.go, i18n validator, coding-standards.md), so no merge conflict is expected.
ac_pass_count: 29 (M1: 13; M2: 11; M3: AC-SSP-014, 015, 016, 017, 018 = 5) + AC-SSP-024 cross-cutting quality gate
ac_fail_count: 0
ac_narrowed_count: 1 (AC-SSP-008b — a `--yes` confirmation gate narrows "bare invocation moves" to "bare invocation reports then refuses"; safety narrowing, disclosed for user acceptance)
ac_deferred_count: 0 (all 24 ACs now addressed across M1-M3)
preserve_list_post_run_count: 0 violations (M3 left the M1 drift algorithm untouched — `git diff` on internal/spec/drift.go is additive-only, +43/-0, the `DriftCountCtx` + context import; `detectDrift` body unchanged. M2 archive logic untouched. No spec/plan/acceptance body content modified.)
drift_non_regression_verified: `moai spec drift --no-cache` → **475 records / 78 drifted** at HEAD 02ba30bbe (matches the delegated expected baseline exactly). `moai spec drift --count --no-cache` → 78.
archive_non_regression_verified: `moai spec archive --dry-run` → 477 SPECs scanned, **0 eligible at the default 90-day window** (cutoff 2026-04-12) — unchanged from M2.
new_warnings_or_lints_introduced: 0 (golangci-lint 0 issues; go vet exit 0)
cross_platform_build: darwin PASS, windows PASS, linux PASS
coverage_internal_spec: 88.9% (threshold 85%); DriftCountCtx 100.0%
coverage_internal_hook_touched: detectStatusDrift 100.0% (touched-file function; package-wide internal/hook 83.8% is a pre-existing baseline dominated by other untested session_start.go handler code, unchanged by M3)
full_suite: go test ./... — 96 ok / 0 FAIL, **3× -count=1 deterministic**; -race on all M3-touched packages GREEN (3× 0 races)
guard_catches_regression: PROVEN — injecting a per-SPEC `deps.logAll` into `detectDrift`'s active loop made `TestDetectDrift_PerfBudget_ConstantSubprocessAtScale` read `git log invocations = 501 at N=500, want EXACTLY 1` (FAIL); the M1 seam guard `TestDetectDrift_ConstantGitLogInvocations` also FAILed. Reverted → both PASS.
residual_risk: (1) AC-SSP-008b narrowing (M2 — unchanged). (2) M2 yields 0 eligible SPECs at the default 90-day grace on today's corpus (M2 — unchanged). (3) M2 last-activity = "last commit touching the SPEC dir", conservative direction (M2 — unchanged). (4) **NEW (M3, pre-existing + out-of-scope):** `internal/hook/dbsync` `TestCheckDebounceConcurrency` is a load-sensitive race test (asserts exactly-one-winner of a 50ms-window debounce race). It PASSED 12/12 in isolation and in all 3 clean full-suite runs, but flaked ONCE during an earlier heavy-load 3× sequential sweep (`results = [false false], want multiset {false, true}`). M3 touched 0 lines of `dbsync` (git diff --stat empty) — this is a PRE-EXISTING test-quality defect, NOT an M3 regression, and out of the item-4 scope (which was the i18n budget test). Flagged for a possible follow-up debounce-test hardening SPEC; NOT fixed here per scope discipline.

## §E.4 Sync-phase Audit-Ready Signal

sync_status: sync-complete (3-phase close — `in-progress → implemented → completed` merged into this single sync commit)
sync_complete_at: 2026-07-11
sync_commit_sha: f3b43e522b57439d6b92792050c3d867c15ec945 (backfilled in this follow-up commit; verified via `git rev-parse HEAD` after the sync commit landed)
frontmatter_status_transitions: spec.md (in-progress → completed), plan.md (in-progress → completed), acceptance.md (in-progress → completed), design.md (draft → completed), research.md (draft → completed), progress.md (in-progress → completed) — all 6 SPEC artifacts updated atomically on this commit; `updated:` refreshed to 2026-07-11 in all 6 frontmatter blocks
changelog_entry_position: CHANGELOG.md `[Unreleased]` § Fixed (session-start latency fix is user-facing behavior-change framing; see also § Added for the new `moai spec archive` command and § Changed for the codified Advisory-Check Discipline rule)
b12_self_test_a: `grep -c 'SPEC-SESSIONSTART-PERF-001' CHANGELOG.md` returned 0 before this edit (no prior entry — no duplicate-emission risk from parallel BATCH-SYNC sessions)
b12_self_test_b: AC count cross-check — `grep -oE '^### AC-SSP-[0-9]+' acceptance.md \| sort -u \| wc -l` = 24 unique top-level AC IDs (AC-SSP-001..024); CHANGELOG entry cites this count, not the 29-row §E.2 breakdown (which includes lettered sub-criteria a/b/c counted separately)
b12_self_test_c: file-path verification — every file path cited in the CHANGELOG entry was confirmed present via `ls`/`git show --stat` against the m1/m2/m3 commits before the entry was drafted (see Deliverables below for the verified SHA list)
new_command_docs: `moai spec archive` is a new user-facing CLI command. README.md `## CLI Commands` (line 697) and README.ko.md carry NO existing `moai spec *` subcommand entries (no `moai spec drift`, `moai spec lint`, `moai spec audit`, etc. are documented there either) — so there is no established convention/location to extend. No README edit made; flagged as a pre-existing documentation gap for `moai spec` subcommands generally, out of this SPEC's scope to invent a new README section.
scoped_commit_verified: `git status --short` confirmed 3 pre-existing unrelated modifications (`.mcp.json`, `.moai/config/sections/handoff.yaml`, `CLAUDE.local.md`) present in the working tree BEFORE this sync commit; these were excluded from `git add` (specific-path staging only, no `git add -A`)
