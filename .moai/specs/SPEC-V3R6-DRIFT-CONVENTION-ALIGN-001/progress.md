# Progress — SPEC-V3R6-DRIFT-CONVENTION-ALIGN-001

> Tier S — `internal/spec` drift classifier convention alignment (run-phase record).

## §D — Plan Audit Record (recorded by orchestrator)

- plan-auditor verdict: **PASS 0.87** (Tier S PASS threshold 0.75).
- GATE-2: user-approved (run-phase entry confirmed).
- Run-phase entry HEAD: `de2f1ac40` (origin/main synced, 0 0 divergence).

## §E — Phase 0.95 Mode Selection

### Input parameters

- **tier**: S (Simple — `internal/spec/*.go` classifier + walker fix, 2 source files + tests)
- **scope (file count)**: 4 files (`transitions.go`, `drift.go`, `transitions_test.go`, new `drift_test.go`) + progress.md
- **domain count**: 1 (Go source code in a single package `internal/spec`)
- **file language mix**: 100% Go
- **concurrency benefit**: LOW (coding-heavy single-package work; Finding A4 caveat — coding tasks have fewer truly-parallelizable subtasks than research)
- **Agent Teams prereqs status**: not evaluated (single-domain coding task; Mode 3 not a candidate)

### Mode evaluation table

| Mode | Selected? | Rationale |
|------|-----------|-----------|
| 1 trivial | not selected | Multi-file semantic change (classifier rule + walker skip + tests); not a typo/single-line edit |
| 2 background | not selected | Requires Write/Edit (CONST-V3R2-020 forbids background write); not read-only |
| 3 agent-team | not selected | Single-domain (1 package), not multi-domain (≥3) — capability gate not met |
| 4 parallel | not selected | Coding-heavy single-package work; Finding A4 caveat prefers sequential |
| 5 sub-agent | **selected** | Default fallback; single sequential manager-develop (cycle_type=tdd) per milestone |

### Decision

`Decision: sub-agent`

### Justification

This is a single-domain, coding-heavy Tier S task confined to one Go package (`internal/spec`). Per Finding A4 ("most coding tasks involve fewer truly parallelizable tasks than research"), sequential sub-agent execution (Mode 5) is the correct default. Agent Teams (Mode 3) and parallel multi-spawn (Mode 4) are not warranted because the work is not multi-domain (1 domain, < 3) and the RED-GREEN-REFACTOR milestones (M1 tests → M2 fix → M3 verify) are dependency-ordered, not parallelizable.

## §E.2 — Run-phase Evidence

### AC PASS/FAIL matrix

| AC | Status | Actual Output |
|----|--------|---------------|
| AC-DCA-001 (PRIMARY — 4 named exemplars aligned) | PASS | All 4 (`GIT-STRATEGY-SCHEMA-001`, `CI-FLAKY-STABILIZE-002`, `ANTHROPIC-AUDIT-TIER3-001`, `CI-FLAKY-STABILIZE-001`) git-implied=`completed`, `Drifted: false` |
| AC-DCA-001b (SECONDARY — count strictly < 67) | PASS | `moai spec drift --count` = 51 (baseline 67; 16 sub-class-1 false-positives resolved) |
| AC-DCA-002 (close commit → completed) | PASS | `TestClassifyPRTitle_CloseInfix` ok — `chore(...): Mx-phase audit-ready signal + 4-phase close` → `(mx-close, completed)` |
| AC-DCA-003 (SPEC-ID-scoped chore disambiguated) | PASS | `TestShouldSkipCommitTitle_BackfillChore` ok — backfill chore skipped, `chore(spec):`/`chore(specs):` still skipped |
| AC-DCA-004 (walker returns completed before sync docs) | PASS | `TestGetGitImpliedStatus_CloseInfixWinsBeforeSyncDocs` + `_CloseInfixDirect` ok |
| AC-DCA-005 (metadata-sweep skip regression) | PASS | `TestClassifyPRTitle_ChoreSpecUnchanged` (AC-LSCSK-003) ok |
| AC-DCA-006 (word-boundary unchanged + audit clean) | PASS | LSGF tests ok; `moai spec audit` MUST-FIX count = 0 |
| AC-DCA-007 (full package suite) | PASS-WITH-DEBT | All `internal/spec` tests green EXCEPT pre-existing `TestCatalogHashParity` (sync-auditor.md template/catalog hash drift at HEAD de2f1ac40 — out of scope, see blocker note in run report) |
| AC-DCA-008 (backfill-no-regression) | PASS | `TestGetGitImpliedStatus_BackfillNoRegression` ok — `implemented` (NOT `completed`) when no close-infix; `_CombinedBackfillCloseSubject` D5 ok |

### Cross-platform build

- `go build ./...` (darwin) → exit 0
- `GOOS=windows GOARCH=amd64 go build ./...` → exit 0

### Coverage / Lint / Boundary

- `go test -cover ./internal/spec/...` → 85.9% (≥ 85% threshold)
- `golangci-lint run ./internal/spec/...` → 0 issues (NEW == baseline)
- subagent-boundary grep (`AskUserQuestion` in non-test `internal/spec/`) → empty (pass)

### Design notes

- `transitions.go`: added pre-loop close-infix check (`4-phase close` / `mx-phase audit-ready` → `(mx-close, completed)`) + narrow `docs(SPEC-...): ...sync-phase` → `(sync-merge, implemented)` rule (AC-DCA-008 required this — plan.md §M2 escape hatch "unless a failing in-scope fixture requires it"). Generic `docs`→`in-progress` rule left untouched.
- `drift.go`: extended `shouldSkipCommitTitle` with a narrow SPEC-ID-scoped backfill-skip (skip when SPEC-ID-scoped AND contains `backfill` AND NOT close-infix — D5 guard). `chore(spec):`/`chore(specs):` metadata-sweep skip + LSGF-001 word-boundary preserved verbatim.
- New positive `completed` signal is the close-infix ONLY (REQ-DCA-005 / AP-2 anti-goal honored — no `completed` inferred for SPECs lacking a close commit).

## §E.2 Sync-phase Audit-Ready Signal

- `sync_commit_sha:` (left EMPTY for backfill by orchestrator's `moai spec close`)
- Frontmatter transition: `status: in-progress` → `status: implemented` (applied at sync commit)
- CHANGELOG entry: single line under `## [Unreleased]` → `### Fixed` (AC-LSCSK-003 + no-op status clause per plan.md §A.5)
- README: no update (Tier S internal/spec, user-facing surface unchanged)
- docs-site: no update (internal classifier, user docs stable)

## §E.5 Mx-phase Audit-Ready Signal

- `mx_commit_sha:` (left EMPTY for backfill by orchestrator's `moai spec close`)
- Mx determination: SKIP-eligible — classifier fix is purely declarative Go code (transitions.go + drift.go rewrites), no fan_in≥3 functions, no goroutines, no per-entity cross-module routing → Mx tags not required per MX-tag-protocol.md § Tag Necessity Heuristic
- Plan-phase commit count: 1 (`de2f1ac40` plan-phase create) + run-phase 2 (`69966e8a9` final commit) = 3 commits total → audit eligible
- version: append `v0.2.0 (sync-phase, 2026-06-03)` to §A.3 Version list
