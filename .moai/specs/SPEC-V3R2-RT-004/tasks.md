# SPEC-V3R2-RT-004 Task Breakdown

> Granular task decomposition of M1-M5 milestones from `plan.md` §2.
> Companion to `spec.md` v0.1.0, `research.md` v0.1.0, `plan.md` v0.1.0, `acceptance.md` v0.1.0.

## HISTORY

| Version | Date       | Author                            | Description                                                            |
|---------|------------|-----------------------------------|------------------------------------------------------------------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow)      | Initial task breakdown — 17 tasks (T-RT004-01..17) across M1-M5         |

---

## Task ID Convention

- ID format: `T-RT004-NN`
- Priority: P0 (blocker), P1 (required), P2 (recommended), P3 (optional)
- Owner role: `manager-tdd`, `manager-docs`, `expert-backend` (Go), `manager-git` (commit/PR boundary)
- Dependencies: explicit task ID list; tasks with no deps may run in parallel within their milestone
- DDD/TDD alignment: per `.moai/config/sections/quality.yaml` `development_mode: tdd`, M1 (RED) precedes M2-M5 (GREEN/REFACTOR)

[HARD] No time estimates per `.claude/rules/moai/core/agent-common-protocol.md` §Time Estimation. Priority + dependencies only.

---

## M1: Test Scaffolding (RED phase) — Priority P0

Goal: Add ~11 failing tests + 1 audit lint that exercise the gaps identified in research.md §2.1. Per `spec-workflow.md` TDD: write failing test first.

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-RT004-01 | Add `TestCheckpoint_BlockerOutstanding` to `internal/session/store_test.go`. Creates an outstanding blocker file, attempts Checkpoint for same (phase, SPECID), asserts `ErrBlockerOutstanding`. | expert-backend | `internal/session/store_test.go` (extend, ~30 LOC) | none | 1 file (extend) | RED — fails today (skeleton checks inline `state.BlockerRpt` but not blocker files on disk) |
| T-RT004-02 | Add `TestHydrate_StaleCheckpoint`, `TestHydrate_WithinTTL`, `TestHydrate_ResumeBypass`, `TestHydrate_CorruptedJSON`, `TestHydrate_MalformedJSON`, `TestHydrate_NotExists` to `internal/session/store_test.go`. | expert-backend | same file (extend, ~80 LOC) | T-RT004-01 | 1 file (extend) | RED — most fail today (no `HydrateWithOpts`, no validator on read) |
| T-RT004-03 | Add `TestCheckpoint_ConcurrentRace` to `internal/session/store_test.go` using 2 goroutines + `sync.WaitGroup`. Confirms exactly one wins; the other gets `ErrCheckpointConcurrent`. | expert-backend | same file (extend, ~50 LOC) | T-RT004-01 | 1 file (extend) | RED — fails today (no advisory lock; both writes succeed via os.Rename) |
| T-RT004-04 | Add `TestCheckpoint_ValidatorRejectsBadHarness`, `TestCheckpoint_ValidatorAcceptsGoodHarness`, `TestCheckpoint_ValidatorRejectsEmptyHarness` to `internal/session/store_test.go`. Asserts `err.Error()` contains `"Harness"` substring on rejection. | expert-backend | same file (extend, ~40 LOC) | T-RT004-01 | 1 file (extend) | RED — fails today (no validator/v10 tags on RunCheckpoint) |
| T-RT004-05 | Add `TestProvenanceRoundTrip` to `internal/session/state_test.go`. Marshals + unmarshals `PhaseState` with non-empty `Provenance{Source, Origin, Loaded}`, asserts round-trip equality. | expert-backend | `internal/session/state_test.go` (extend, ~25 LOC) | T-RT004-01 | 1 file (extend) | RED initially (existing test does not assert Provenance; baseline OK after M2) |
| T-RT004-06 | Add `TestWriteRunArtifact_BinaryUnderArtifacts`, `TestWriteRunArtifact_TextEncodingMismatch` to `internal/session/store_test.go`. | expert-backend | same file (extend, ~35 LOC) | T-RT004-01 | 1 file (extend) | RED — encoding mismatch fails today (no UTF-8 validation) |
| T-RT004-07 | Create `internal/session/lock_test.go` with `TestFileLock_HappyPath`, `TestFileLock_HighContention` (10 goroutines). Skip on Windows if flaky (`t.Skip("requires fast disk")`). | expert-backend | new file (~80 LOC) | T-RT004-01 | 1 file (create) | RED — fails until M3 lock implementation |
| T-RT004-08 | Add `TestCheckpoint_HappyPath` regression check + `TestHydrate_RoundTrip` extension to `internal/session/store_test.go`. Verifies skeleton baseline still passes. | expert-backend | same file (extend, ~20 LOC) | T-RT004-01 | 1 file (extend) | GREEN at baseline — regression sentinel |
| T-RT004-09 | Add `TestRecordBlocker_FilenameFormat`, `TestCheckpoint_AfterBlockerResolved` to `internal/session/store_test.go`. Asserts filename format `blocker-{phase}-{SPECID}-{timestamp}.json`. | expert-backend | same file (extend, ~30 LOC) | T-RT004-01 | 1 file (extend) | RED — current skeleton uses `{Kind}-{Source}` filename (mismatch with spec REQ-012) |
| T-RT004-10 | Create `internal/cli/state_test.go` with `TestStateDump_HappyPath`, `TestStateDump_FormatJSON`, `TestStateDump_NoMatch`, `TestStateShowBlocker_*`, `TestRun_ResumeFlag`. | expert-backend | new file (~120 LOC) | T-RT004-01 | 1 file (create) | RED — fails until M5 CLI implementation |
| T-RT004-11 | Run `go test ./internal/session/ ./internal/cli/ ./internal/template/` and confirm RED state for all new tests; existing tests still GREEN (regression sentinel). | manager-tdd | n/a (verification only) | T-RT004-01..10 | 0 files | RED gate verification |

**M1 priority: P0** — blocks all subsequent milestones. TDD discipline.

T-RT004-01 through T-RT004-10 may execute in parallel (touch independent file regions) where the same file is not edited concurrently; otherwise sequential within `store_test.go`.

---

## M2: Validator/v10 + atomic-write helper (GREEN, part 1) — Priority P0

Goal: Add validator/v10 schema tags + `Validate()` method + extract atomic-write helper to satisfy AC-01, AC-09, AC-15.

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-RT004-12 | Extend `internal/session/checkpoint.go`: add `validate:"..."` tags to `PlanCheckpoint`, `RunCheckpoint`, `SyncCheckpoint`. Add `Harness string` field to `RunCheckpoint` with `validate:"required,oneof=minimal standard thorough"`. Add `Validate() error` method to `Checkpoint` interface; implement on each concrete type. | expert-backend | `internal/session/checkpoint.go` (extend, ~30 LOC) | T-RT004-11 | 1 file (edit) | GREEN — exposes validator tag surface |
| T-RT004-13 | Create `internal/session/atomic.go` with `writeAtomic(path string, data []byte, perm os.FileMode) error` helper. Pattern: write `<path>.tmp` → `os.Rename(<path>.tmp, <path>)`. Refactor 4 call sites in `internal/session/store.go` (`Checkpoint`, `WriteRunArtifact`, `RecordBlocker`, `ResolveBlocker`) to use the helper. | expert-backend | `internal/session/atomic.go` (new, ~25 LOC) + `store.go` (4 call sites) | T-RT004-12 | 2 files (create + edit) | REFACTOR — DRY consolidation |
| T-RT004-14 | Wire `Validate()` into `FileSessionStore.Checkpoint`: call before atomic write; return `ErrCheckpointInvalid` wrapped with offending field name on failure. Wire into `Hydrate` after JSON decode for symmetry (AC-09). | expert-backend | `internal/session/store.go` (extend Checkpoint + Hydrate) | T-RT004-12, T-RT004-13 | 1 file (edit) | GREEN — AC-15, AC-09 |

**M2 priority: P0** — blocks M3-M5. AC-01, AC-09, AC-15 turn GREEN.

T-RT004-12 must complete before T-RT004-14 (validator tags must exist before wiring `Validate()`).

---

## M3: Cross-platform advisory locking (GREEN, part 2) — Priority P0

Goal: Implement `fileLock` interface + Unix flock + Windows LockFileEx to satisfy AC-10.

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-RT004-15 | Create `internal/session/lock.go` with `fileLock` interface (`acquire`, `release`) + `newFileLock()` factory. Create `internal/session/lock_unix.go` (`//go:build unix`) using `golang.org/x/sys/unix.Flock` with `LOCK_EX\|LOCK_NB`. Create `internal/session/lock_windows.go` (`//go:build windows`) using `golang.org/x/sys/windows.LockFileEx`. | expert-backend | 3 new files (~140 LOC total) | T-RT004-14 | 3 files (create) | GREEN — lock primitive |
| T-RT004-16 | Wire lock acquisition into `FileSessionStore.Checkpoint`: 3-retry / 10ms-backoff loop. On repeated loss, return `ErrCheckpointConcurrent`. Lock companion file is `<checkpoint-path>.lock` (sibling of the target file). Best-effort cleanup on release. | expert-backend | `internal/session/store.go` (extend Checkpoint, ~20 LOC) | T-RT004-15 | 1 file (edit) | GREEN — AC-10 |

**M3 priority: P0** — blocks M4 (in-flight detection uses the same lock primitive for idempotent reads). AC-10 turns GREEN.

[HARD] T-RT004-15 must NOT introduce new direct module dependencies beyond `golang.org/x/sys/{unix,windows}` (already in `go.mod` indirect deps).

---

## M4: Provenance + blocker-outstanding + stale-check + STALE_SECONDS config + in-flight detection + team merge (GREEN, part 3) — Priority P0

Goal: Wire the 6 remaining behavioural REQs to satisfy AC-04, AC-05, AC-06, AC-07 (provenance round-trip portion), AC-08, AC-14.

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-RT004-17 | Extend `FileSessionStore.Checkpoint` to scan `.moai/state/blocker-{phase}-{SPECID}-*.json` for unresolved blockers BEFORE writing. If any unresolved blocker matches `(phase, SPECID)`, return `ErrBlockerOutstanding`. | expert-backend | `internal/session/store.go` (extend Checkpoint, ~20 LOC) | T-RT004-16 | 1 file (edit) | GREEN — AC-04 |
| T-RT004-18 | Rename blocker file format in `RecordBlocker` from `blocker-{Kind}-{Source}-{ts}.json` to `blocker-{phase}-{SPECID}-{timestamp}.json` per spec.md REQ-012. Backfill the phase + SPECID into `BlockerReport` (add `Phase Phase` and `SPECID string` fields if not already inferable). | expert-backend | `internal/session/blocker.go` (extend struct), `internal/session/store.go` (extend RecordBlocker) | T-RT004-17 | 2 files (edit) | GREEN — AC-03 (filename), AC-04 (scan key) |
| T-RT004-19 | Add `HydrateWithOpts(phase Phase, specID string, opts HydrateOpts) (*PhaseState, error)` method. `HydrateOpts{SkipStaleCheck bool}`. Existing `Hydrate(phase, specID)` calls `HydrateWithOpts(phase, specID, HydrateOpts{})`. When `SkipStaleCheck: true`, skip the stale check; log stderr WARN if hydrating a stale checkpoint. | expert-backend | `internal/session/store.go` (extend, ~20 LOC) | T-RT004-17 | 1 file (edit) | GREEN — AC-06 |
| T-RT004-20 | Add `SessionConfig{StaleSeconds int}` to `internal/config/types.go`. Extend `internal/config/loader.go` to read `.moai/config/sections/ralph.yaml` `stale_seconds:` key and populate `Config.Session.StaleSeconds`. Default: 3600. | expert-backend | 2 files (edit, ~30 LOC) | T-RT004-19 | 2 files (edit) | GREEN — AC-05 (configurable) |
| T-RT004-21 | Add `DetectInFlightTransition(specID string) (fromPhase Phase, toPhase Phase, found bool, err error)` method. Scans `.moai/state/checkpoint-*-{specID}.json`, returns the most recent (Phase) without a corresponding next-phase file. Phase ordering: plan → run → sync. | expert-backend | `internal/session/store.go` (extend, ~40 LOC) | T-RT004-19 | 1 file (edit) | GREEN — AC-14 |
| T-RT004-22 | Create `internal/session/team_merge.go` with `MergeTeamCheckpoints(specID string, phase Phase, agentNames []string) (*PhaseState, error)`. Reads each `checkpoint-{phase}-{specID}-{agent}.json`, validates each, short-circuits on unresolved blocker (REQ-051 bubble-mode), merges per phase-specific union rule (research.md §7.3), sets `Provenance.Source = "session"` + audit subfield. | expert-backend | new file (~80 LOC) | T-RT004-19 | 1 file (create) | GREEN — AC-08 |

**M4 priority: P0** — blocks M5 (CLI subcommand consumes these methods). AC-04, AC-05, AC-06, AC-07 (provenance), AC-08, AC-14 turn GREEN.

T-RT004-19, T-RT004-20, T-RT004-21, T-RT004-22 may execute in parallel (touch independent file regions). T-RT004-17 + T-RT004-18 must precede the others (filename format change is structural).

---

## M5: CLI subcommand + cache-prefix invariant + clean retention + ArtifactEncodingMismatch + AskUserQuestion audit + CHANGELOG + MX tags (GREEN, part 4 + REFACTOR + Trackable) — Priority P1

Goal: User-facing surface, anti-regression CI lint, documentation, MX tag insertions.

| ID | Subject | Owner role | File:line target | Dependency | Touch points | TDD alignment |
|----|---------|-----------|-------------------|------------|--------------|---------------|
| T-RT004-23 | Create `internal/cli/state.go` registering `moai state dump [--phase] [--spec] [--format json\|human]` and `moai state show-blocker [--phase] [--spec]` subcommands. Wire into `internal/cli/root.go` (or wherever moai root is composed). Pretty-print PhaseState with per-field provenance. | expert-backend | new + 1 edit (~150 LOC) | T-RT004-22 | 2 files (create + edit) | GREEN — AC-07, REQ-007, REQ-030, REQ-032 |
| T-RT004-24 | Create `internal/session/hydrate.go` with the load-bearing comment `// cache-prefix: DO NOT REORDER` and `HydrateForPrompt(phase, specID) (system, user, sys2 string, err error)` returning the 3 pieces in fixed order. | expert-backend | new file (~40 LOC) | T-RT004-22 | 1 file (create) | GREEN — P-C05 closure |
| T-RT004-25 | Extend `internal/session/store.go::WriteRunArtifact` to validate UTF-8 for text-declared extensions (`.md\|.txt\|.json\|.yaml\|.yml`); on invalid UTF-8 return `ErrArtifactEncodingMismatch` (new sentinel) without writing the file. Binary extensions exempt. | expert-backend | `internal/session/store.go` (extend, ~20 LOC) | T-RT004-22 | 1 file (edit) | GREEN — REQ-043, AC-12 edge case |
| T-RT004-26 | Extend `internal/cli/clean.go` to read `.moai/config/sections/state.yaml` `retention_days: N`. When set, mark `runs/{iter-id}/` directories with `mtime` older than N days as eligible for prune. Default behaviour: dry-run; require `--force` + AskUserQuestion confirmation to actually delete. | expert-backend | `internal/cli/clean.go` (extend, ~40 LOC) | T-RT004-22 | 1 file (edit) | GREEN — REQ-031, AC-13 |
| T-RT004-27 | Extend `internal/cli/run.go` (and `internal/cli/loop.go` if exists) to pass `--resume` flag through to `HydrateWithOpts(phase, specID, HydrateOpts{SkipStaleCheck: true})`. On `ErrCheckpointStale` without `--resume`, return error to orchestrator (AskUserQuestion routing happens at orchestrator level, NOT inside CLI). | expert-backend | 1-2 files (edit, ~20 LOC) | T-RT004-23 | 2 files (edit) | GREEN — AC-06 wiring |
| T-RT004-28 | Create `internal/template/agent_askuser_audit_test.go` with `TestNoAskUserQuestionInSubagents` walking `.claude/agents/**.md` (NOT skills) asserting body contains no literal `AskUserQuestion(`. Sentinel: `ASKUSERQUESTION_IN_SUBAGENT: <file> body line <N> contains AskUserQuestion(; subagents must use BlockerReport, not AskUserQuestion. See agent-common-protocol.md §User Interaction Boundary.`. | expert-backend | new file (~30 LOC) | T-RT004-23 | 1 file (create) | GREEN — AC-11 |
| T-RT004-29 | Insert MX tags per `plan.md` §6: 3 `@MX:ANCHOR` (Checkpoint, HydrateForPrompt, fileLock interface) + 2 `@MX:NOTE` (PhaseState, BlockerReport) + 2 `@MX:WARN` (Hydrate stale check, WriteRunArtifact UTF-8 check) across 6 distinct files. Tag content per `plan.md` §6 verbatim. | manager-docs | 6 files (edit, MX tag insertion only) | T-RT004-23..28 | 6 files (edit) | Trackable — plan §6 |
| T-RT004-30 | Add CHANGELOG entry under `## [Unreleased] / ### Added` per `plan.md` §M5f text. | manager-docs | `CHANGELOG.md` (extend) | T-RT004-29 | 1 file (edit) | Trackable |
| T-RT004-31 | Run `make build` from worktree root to regenerate `internal/template/embedded.go`. Verify diff scope (only audit test addition expected; no skills/rules content changes). | manager-docs | `internal/template/embedded.go` (regenerated) | T-RT004-30 | 1 file (regenerated) | Build verification |
| T-RT004-32 | Run full `go test ./...` from worktree root. Verify ALL audit tests pass + 0 cascading failures (per `CLAUDE.local.md` §6 HARD rule). Run `go vet ./...` + `golangci-lint run` — zero warnings. | manager-tdd | n/a (verification only) | T-RT004-31 | 0 files | GREEN gate (final) |
| T-RT004-33 | Update `progress.md` with `run_complete_at: <timestamp>` and `run_status: implementation-complete`. | manager-docs | `progress.md` (extend) | T-RT004-32 | 1 file (edit) | Trackable closure |

**M5 priority: P1** — completes the SPEC. AC-11, AC-12, AC-13 turn GREEN.

T-RT004-23, T-RT004-24, T-RT004-25, T-RT004-26 may execute in parallel (touch independent file regions). T-RT004-27 depends on T-RT004-23 (CLI flag wiring). T-RT004-28 may run anytime after T-RT004-23.

T-RT004-29 must wait for all M5 source code edits to land (MX tag positions reference final code structure).

---

## Task summary by milestone

| Milestone | Task IDs | Total tasks | Priority | Owner role mix |
|-----------|----------|-------------|----------|----------------|
| M1 (RED) | T-RT004-01..11 | 11 | P0 | expert-backend (10) + manager-tdd (1 verification) |
| M2 (GREEN part 1) | T-RT004-12..14 | 3 | P0 | expert-backend |
| M3 (GREEN part 2) | T-RT004-15..16 | 2 | P0 | expert-backend |
| M4 (GREEN part 3) | T-RT004-17..22 | 6 | P0 | expert-backend |
| M5 (GREEN part 4 + REFACTOR + Trackable) | T-RT004-23..33 | 11 | P1 | expert-backend (5) + manager-docs (4) + manager-tdd (1 verification) + manager-git (closure) |
| **TOTAL** | T-RT004-01..33 | **33 tasks** | — | — |

> NOTE: 33 tasks span the 5 milestones. M1 (RED) opens with 11 test-only tasks; M2-M4 (GREEN core) deliver the typed-state subsystem in 11 tasks; M5 (final GREEN + REFACTOR + Trackable) closes with 11 tasks of CLI, lint, MX tags, and trackability work.

---

## Dependency graph (critical path)

```
T-RT004-01..10 (M1 tests, parallel)
   ↓
T-RT004-11 (M1 verification gate)
   ↓
T-RT004-12 (validator tags, GREEN start)
   ↓
T-RT004-13 (atomic.go helper) ─┐
   ↓                           │
T-RT004-14 (Validate wiring) ─┘
   ↓
T-RT004-15 (lock.{go,_unix.go,_windows.go}, M3 start)
   ↓
T-RT004-16 (lock wiring in Checkpoint)
   ↓
T-RT004-17 (blocker-file scan)
   ↓
T-RT004-18 (filename rename)
   ↓
T-RT004-19, 20, 21, 22 (M4 parallel)
   ↓
T-RT004-23, 24, 25, 26 (M5 parallel)
   ↓
T-RT004-27, 28 (M5 dependents)
   ↓
T-RT004-29 (MX tags) → T-RT004-30 (CHANGELOG) → T-RT004-31 (make build) → T-RT004-32 (go test ./... + vet + lint) → T-RT004-33 (progress.md closure)
```

Critical path: 11 sequential gates from T-RT004-11 → T-RT004-33 (M1 → M5 closure).

---

## Cross-task constraints

[HARD] All file edits use absolute paths under the worktree root `/Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-004` per CLAUDE.md §Worktree Isolation Rules.

[HARD] All tests use `t.TempDir()` per `CLAUDE.local.md` §6 — no test creates files in the project root.

[HARD] All filesystem operations use `filepath.Join` / `filepath.Abs`; no `filepath.Join(cwd, absPath)` patterns per `CLAUDE.local.md` §6.

[HARD] No new direct module dependencies beyond `validator/v10` (already in go.mod via SCH-001) and `golang.org/x/sys/{unix,windows}` (already indirect).

[HARD] No `internal/template/templates/.claude/...` files modified by this SPEC (no skills, no rules, no commands changes). Only the embed regeneration via `make build` is required.

[HARD] Code comments in Korean (per `.moai/config/sections/language.yaml` `code_comments: ko`). Godoc and exported identifier docstrings remain English (industry standard).

[HARD] Commit messages in Korean (per `.moai/config/sections/language.yaml` `git_commit_messages: ko`).

---

End of tasks.md.
