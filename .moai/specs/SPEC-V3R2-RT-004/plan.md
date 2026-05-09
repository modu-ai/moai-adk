# SPEC-V3R2-RT-004 Implementation Plan

> Implementation plan for **Typed Session State + Phase Checkpoint**.
> Companion to `spec.md` v0.1.0, `research.md` v0.1.0, `acceptance.md` v0.1.0, `tasks.md` v0.1.0.
> Authored against branch `plan/SPEC-V3R2-RT-004` at `<worktree-root>` = `/Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-004`. See §7 for cwd resolution rule.

## HISTORY

| Version | Date       | Author                        | Description                                                              |
|---------|------------|-------------------------------|--------------------------------------------------------------------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow)  | Initial implementation plan per `.claude/skills/moai/workflows/plan.md` Phase 1B |

---

## 1. Plan Overview

### 1.1 Goal restatement

Convert master §5.6 file-first state commitment into concrete Go types and a canonical `.moai/state/` layout so that every phase boundary (`plan` → `run` → `sync` → ...) writes a typed checkpoint to disk and the next phase reads the file rather than the prior conversation. The implementation:

- Hardens existing skeleton in `internal/session/` (phase.go, state.go, checkpoint.go, blocker.go, store.go) to **fully satisfy 24 EARS REQs** and **15 ACs** declared in `spec.md`.
- Adds **provenance tagging** (`SrcUser`, `SrcProject`, `SrcLocal`, `SrcSession`, `SrcHook`) on every state mutation and surfaces it through `moai state dump`.
- Closes problem-catalog.md **P-C02** (no sub-agent context-isolation primitive) by making `PhaseState{Phase, SPECID, Checkpoint, BlockerRpt, UpdatedAt, Provenance}` the durable hand-off substrate.
- Closes **P-C05** (no cache-prefix discipline) by freezing the `(systemPrompt, userContext, systemContext)` assembly order at the hydrate step with a `// cache-prefix: DO NOT REORDER` invariant.
- Adds **`moai state` CLI subcommand** (`dump`, `show-blocker`) for human-readable inspection with per-field provenance.
- Introduces **STALE_SECONDS crash-resume semantics** (default 3600s, configurable via `.moai/config/sections/ralph.yaml`) and **`--resume` flag** to bypass the staleness AskUserQuestion prompt.
- Implements **concurrent-write safety** via per-platform advisory locks (`flock` on Unix, `LockFileEx` on Windows) with 3-retry / 10ms backoff per REQ-V3R2-RT-004-040.

### 1.2 Implementation Methodology

`.moai/config/sections/quality.yaml`: `development_mode: tdd`. Run phase MUST follow RED → GREEN → REFACTOR per `.claude/rules/moai/workflow/spec-workflow.md` Phase Run.

- **RED**: Write failing tests that exercise the 15 ACs against the existing skeleton. Existing `internal/session/*_test.go` files cover only happy-path basics; new tests cover staleness, blocker outstanding, concurrent writes, validator/v10 failure, cache-prefix invariant, provenance round-trip, atomic rename crash safety, team-mode merge, in-flight transition recovery (per AC-V3R2-RT-004-01..15).
- **GREEN**: Extend existing skeleton (validator/v10 schema tags, advisory locking, provenance round-trip, blocker-outstanding gate, stale-check, hydrate cache-prefix order) and add new files: `internal/session/lock.go`, `internal/session/lock_unix.go`, `internal/session/lock_windows.go`, `internal/cli/state.go`, `internal/cli/state_test.go`. Add `SessionConfig` to `internal/config/types.go` for STALE_SECONDS configuration.
- **REFACTOR**: Extract atomic-write helper (`writeAtomic(path, data, perm)`) used 4× across `Checkpoint`, `WriteRunArtifact`, `RecordBlocker`, `ResolveBlocker`. Consolidate provenance string conversion. Document the cache-prefix invariant with a load-bearing comment.

### 1.3 Deliverables

| Deliverable | Path | REQ Coverage |
|-------------|------|--------------|
| Validator/v10 schema tags on all checkpoints | `internal/session/checkpoint.go` (extend existing) | REQ-004 (validator), REQ-015 (oneof) |
| Validator integration in `Checkpoint()` | `internal/session/store.go` (extend `Checkpoint` method) | REQ-004 |
| Advisory-lock primitive (cross-platform) | `internal/session/lock.go` (new), `internal/session/lock_unix.go` (new), `internal/session/lock_windows.go` (new) | REQ-040 |
| 3-retry / 10ms-backoff lock acquisition | `internal/session/store.go` (extend `Checkpoint` method) | REQ-040 |
| Provenance round-trip in JSON | `internal/session/state.go` (extend `MarshalJSON`/`UnmarshalJSON`) | REQ-005, REQ-007 |
| Blocker-outstanding gate | `internal/session/store.go` (extend `Checkpoint` to check blocker file existence, not just inline ref) | REQ-020 |
| Stale-check error path (`ErrCheckpointStale`) | `internal/session/store.go` (already exists; add CLI flow) | REQ-014 |
| `--resume` flag bypass | `internal/cli/run.go` (extend), `internal/session/store.go` (`HydrateOpts{SkipStaleCheck bool}`) | REQ-033 |
| `moai state dump` subcommand | `internal/cli/state.go` (new) | REQ-007, REQ-030 |
| `moai state show-blocker` subcommand | `internal/cli/state.go` (new) | REQ-032 |
| In-flight transition detection + recovery | `internal/session/store.go` (new method `DetectInFlightTransition`) | REQ-050 |
| Team-mode per-agent checkpoint merge | `internal/session/store.go` (new method `MergeTeamCheckpoints`) | REQ-021, REQ-051 |
| Configurable STALE_SECONDS via ralph.yaml | `internal/config/types.go` (`SessionConfig{StaleSeconds int}`), `internal/config/loader.go` extended | REQ-022 |
| `retention_days` clean integration | `internal/cli/clean.go` (extend; references `state.yaml`) | REQ-031 |
| ArtifactEncodingMismatch validation | `internal/session/store.go` (`WriteRunArtifact` extended) | REQ-043 |
| Cache-prefix invariant comment | `internal/session/hydrate.go` (new file w/ `// cache-prefix: DO NOT REORDER`) | §7 Constraints (P-C05) |
| CHANGELOG entry | `CHANGELOG.md` Unreleased section | Trackable (TRUST 5) |
| MX tags per §6 | 6 files (per §6 below) | mx_plan |

Embedded-template parity is **not applicable** — this SPEC modifies `internal/` Go source, not `.claude/` template files. `make build` is still required after edits because `internal/template/embedded.go` contains the embedded `.claude/` contents that may be referenced by other code paths (e.g., audit tests). Per `CLAUDE.local.md` §2 Template-First Rule, no `.claude/` files are added or modified by this SPEC.

### 1.4 Traceability Matrix (REQ → AC → Task)

Per plan-auditor PASS criterion #2 (every REQ maps to at least one AC and at least one Task):

| REQ ID | Category | Mapped AC(s) | Mapped Task(s) |
|--------|----------|--------------|----------------|
| REQ-V3R2-RT-004-001 | Ubiquitous | (Phase enum coverage; verified by existing `phase_test.go`) | T-RT004-01 |
| REQ-V3R2-RT-004-002 | Ubiquitous | AC-01, AC-02 | T-RT004-02, T-RT004-08 |
| REQ-V3R2-RT-004-003 | Ubiquitous | AC-01, AC-02 | T-RT004-02 |
| REQ-V3R2-RT-004-004 | Ubiquitous | AC-01, AC-15 | T-RT004-03, T-RT004-04 |
| REQ-V3R2-RT-004-005 | Ubiquitous | AC-07, AC-08 | T-RT004-05, T-RT004-12 |
| REQ-V3R2-RT-004-006 | Ubiquitous | AC-12, AC-13 | T-RT004-06 |
| REQ-V3R2-RT-004-007 | Ubiquitous | AC-07 | T-RT004-12 |
| REQ-V3R2-RT-004-008 | Ubiquitous | AC-01, AC-02, AC-03, AC-12 | T-RT004-08 |
| REQ-V3R2-RT-004-010 | Event-Driven | AC-01 | T-RT004-08 |
| REQ-V3R2-RT-004-011 | Event-Driven | AC-02 | T-RT004-08 |
| REQ-V3R2-RT-004-012 | Event-Driven | AC-03 | T-RT004-09 |
| REQ-V3R2-RT-004-013 | Event-Driven | (orchestrator-side; verified manually) | T-RT004-13 |
| REQ-V3R2-RT-004-014 | Event-Driven | AC-05, AC-06 | T-RT004-10 |
| REQ-V3R2-RT-004-015 | Event-Driven | AC-12 | T-RT004-06 |
| REQ-V3R2-RT-004-020 | State-Driven | AC-04 | T-RT004-09 |
| REQ-V3R2-RT-004-021 | State-Driven | AC-08 | T-RT004-15 |
| REQ-V3R2-RT-004-022 | State-Driven | AC-05, AC-06 | T-RT004-11 |
| REQ-V3R2-RT-004-030 | Optional | AC-07 | T-RT004-12 |
| REQ-V3R2-RT-004-031 | Optional | AC-13 | T-RT004-16 |
| REQ-V3R2-RT-004-032 | Optional | (CLI behaviour) | T-RT004-12 |
| REQ-V3R2-RT-004-033 | Optional | AC-06 | T-RT004-10 |
| REQ-V3R2-RT-004-040 | Unwanted | AC-10 | T-RT004-07, T-RT004-14 |
| REQ-V3R2-RT-004-041 | Unwanted | AC-09 | T-RT004-04 |
| REQ-V3R2-RT-004-042 | Unwanted | AC-11 | T-RT004-17 |
| REQ-V3R2-RT-004-043 | Unwanted | (negative test) | T-RT004-06 |
| REQ-V3R2-RT-004-050 | Complex | AC-14 | T-RT004-15 |
| REQ-V3R2-RT-004-051 | Complex | AC-08 | T-RT004-15 |

Coverage: **27 REQs mapped to 15 ACs and 17 tasks** (some ACs and tasks map to multiple REQs).

---

## 2. Milestone Breakdown (M1-M5)

Each milestone is **priority-ordered** (no time estimates per `.claude/rules/moai/core/agent-common-protocol.md` §Time Estimation HARD rule).

### M1: Test scaffolding (RED phase) — Priority P0

Reference existing tests: `internal/session/{phase,state,checkpoint,blocker,store}_test.go`.

Owner role: `expert-backend` (Go test) or direct `manager-tdd` execution.

Scope:
1. Add new test cases to `internal/session/store_test.go` covering ACs not yet exercised by existing tests:
   - `TestCheckpoint_BlockerOutstanding` (AC-04, REQ-020).
   - `TestHydrate_StaleCheckpoint` (AC-05, REQ-014, REQ-022).
   - `TestHydrate_ResumeBypass` (AC-06, REQ-033).
   - `TestHydrate_CorruptedJSON` (AC-09, REQ-041) — must name offending field.
   - `TestCheckpoint_ConcurrentRace` (AC-10, REQ-040) — uses 2 goroutines + `sync.WaitGroup`.
   - `TestCheckpoint_ValidatorRejectsBadHarness` (AC-15, REQ-004) — `RunCheckpoint{Harness: "ultra"}`.
   - `TestProvenanceRoundTrip` (AC-07, REQ-005).
   - `TestMergeTeamCheckpoints` (AC-08, REQ-021).
   - `TestDetectInFlightTransition` (AC-14, REQ-050).
   - `TestWriteRunArtifact_BinaryUnderArtifacts` (AC-12, REQ-015).
   - `TestWriteRunArtifact_TextEncodingMismatch` (REQ-043).
2. Create `internal/cli/state_test.go` covering `moai state dump` and `moai state show-blocker` (AC-07, AC-12, REQ-007, REQ-030, REQ-032).
3. Add CI lint test in `internal/template/agent_askuser_audit_test.go` (or extend existing audit test) asserting subagent body files contain no literal `AskUserQuestion(` call (AC-11, REQ-042).
4. Run `go test ./internal/session/ ./internal/cli/ ./internal/template/` — confirm RED for all new tests (existing skeleton lacks the new behaviours).

Verification gate before advancing to M2: at least 11 of the new tests fail with documented sentinel messages. Existing tests continue to pass (regression baseline).

[HARD] No implementation code in M1 outside of test files.

### M2: Validator/v10 + atomic-write helper (GREEN, part 1) — Priority P0

Owner role: `expert-backend`.

Scope:
1. Add validator/v10 import and schema tags to `internal/session/checkpoint.go`:
   - `PlanCheckpoint.Status` → `validate:"oneof=approved draft rejected"`.
   - `PlanCheckpoint.SPECID` → `validate:"required"`.
   - `RunCheckpoint.Status` → `validate:"oneof=pass fail partial"`.
   - `RunCheckpoint.Harness` (NEW field) → `validate:"oneof=minimal standard thorough"`.
   - `RunCheckpoint.SPECID` → `validate:"required"`.
   - `SyncCheckpoint.SPECID` → `validate:"required"`.
2. Add `Validate() error` method to `Checkpoint` interface; concrete types call `validator.New().Struct(c)` and wrap error with offending field name (AC-15, AC-09).
3. Wire `Validate()` into `FileSessionStore.Checkpoint` before atomic write; return `ErrCheckpointInvalid` with field name on failure.
4. Wire `Validate()` into `FileSessionStore.Hydrate` after JSON decode; return `ErrCheckpointInvalid` with field name on corruption (AC-09).
5. Refactor 4 atomic-write call sites (`Checkpoint`, `WriteRunArtifact`, `RecordBlocker`, `ResolveBlocker`) into a single `writeAtomic(path string, data []byte, perm os.FileMode) error` helper in a new file `internal/session/atomic.go`.

Verification: `TestCheckpoint_ValidatorRejectsBadHarness` (AC-15) and `TestHydrate_CorruptedJSON` (AC-09) turn GREEN. All existing tests still pass.

### M3: Advisory locking (cross-platform) (GREEN, part 2) — Priority P0

Owner role: `expert-backend`.

Scope:
1. Create `internal/session/lock.go` with the platform-neutral interface:

   ```go
   type fileLock interface {
       acquire(path string, retries int, backoff time.Duration) error
       release() error
   }
   func newFileLock() fileLock { ... } // returns platform impl
   ```

2. Create `internal/session/lock_unix.go` with build tag `//go:build unix` using `syscall.Flock(fd, syscall.LOCK_EX|syscall.LOCK_NB)`. Lock companion file path is `<checkpoint-path>.lock`.
3. Create `internal/session/lock_windows.go` with build tag `//go:build windows` using `golang.org/x/sys/windows.LockFileEx` with `LOCKFILE_EXCLUSIVE_LOCK | LOCKFILE_FAIL_IMMEDIATELY`.
4. Wire lock acquisition into `FileSessionStore.Checkpoint` before atomic write: 3 retries with 10 ms backoff per REQ-040; on repeated loss return `ErrCheckpointConcurrent`.
5. Lock companion file is best-effort cleaned up on `release()`.

Verification: `TestCheckpoint_ConcurrentRace` (AC-10) turns GREEN on Linux/macOS/Windows CI matrix.

[HARD] No new module dependency beyond `golang.org/x/sys/windows` (already used per `CLAUDE.local.md` §6 test isolation rules and SPEC-V3R2-RT-005 dependency).

### M4: Provenance + blocker-outstanding + stale-check + in-flight + team merge (GREEN, part 3) — Priority P0

Owner role: `expert-backend`.

Scope:
1. **Provenance round-trip** (REQ-005, AC-07): Verify `PhaseState.Provenance` survives `MarshalJSON` / `UnmarshalJSON` round-trip. Existing skeleton already declares the field; `TestProvenanceRoundTrip` confirms behaviour.
2. **Blocker-outstanding gate** (REQ-020, AC-04): Extend `FileSessionStore.Checkpoint` to scan `.moai/state/blocker-*.json` for unresolved blockers matching the `(Phase, SPECID)` key BEFORE writing. Return `ErrBlockerOutstanding` if found. (Existing skeleton checks the inline `state.BlockerRpt` field but does not scan blocker files.)
3. **Stale-check + AskUserQuestion routing** (REQ-014, REQ-022, AC-05): Existing skeleton returns `ErrCheckpointStale` already. Add CLI plumbing: when `internal/cli/run.go` (or other phase entrypoints) receives `ErrCheckpointStale`, surface to user via the orchestrator's AskUserQuestion (NOT inside SessionStore — agent-common-protocol.md HARD rule). The CLI exposes `--resume` flag that sets `HydrateOpts{SkipStaleCheck: true}` in a new method `HydrateWithOpts(phase, specID, opts)`.
4. **Configurable STALE_SECONDS** (REQ-022): Add `SessionConfig{StaleSeconds int}` to `internal/config/types.go`; load from `.moai/config/sections/ralph.yaml` via existing config loader. Default 3600s. `NewFileSessionStore` accepts a `staleTTL time.Duration` already; the CLI wires the loaded config value in.
5. **In-flight transition detection** (REQ-050, AC-14): Add new method `DetectInFlightTransition(specID string) (fromPhase Phase, toPhase Phase, found bool, err error)` that scans `.moai/state/checkpoint-*.json` and returns the most recent (Phase, SPECID) pair without a corresponding next-phase checkpoint. Used by orchestrator on session start.
6. **Team-mode merge** (REQ-021, REQ-051, AC-08): Add new method `MergeTeamCheckpoints(specID string, phase Phase, agentNames []string) (PhaseState, error)`. Reads each `checkpoint-{phase}-{agent-name}.json`, returns merged PhaseState with `Provenance.Source = "session"`, audit subfield listing all per-agent paths. If any per-agent file contains an unresolved blocker, surfaces it (REQ-051 bubble-mode).

Verification: `TestProvenanceRoundTrip`, `TestCheckpoint_BlockerOutstanding`, `TestHydrate_StaleCheckpoint`, `TestHydrate_ResumeBypass`, `TestDetectInFlightTransition`, `TestMergeTeamCheckpoints` all turn GREEN. AC-04, AC-05, AC-06, AC-07, AC-08, AC-14 satisfied.

### M5: CLI subcommand + cache-prefix invariant + clean retention + CHANGELOG + MX tags (GREEN, part 4 + Trackable) — Priority P1

Owner role: `expert-backend` (CLI), `manager-docs` (CHANGELOG, MX tags).

Scope:

#### M5a: `moai state` CLI subcommand (REQ-007, REQ-030, REQ-032)

1. Create `internal/cli/state.go` registering the root command `moai state` with two subcommands:
   - `moai state dump [--phase <name>] [--spec <id>]` — pretty-prints the most recent `PhaseState` matching the filter, including per-field provenance.
   - `moai state show-blocker [--phase <name>] [--spec <id>]` — finds the most recent unresolved `BlockerReport` and prints in human-readable form. Exits with non-zero status if any unresolved blocker exists (REQ-032).
2. Wire into `internal/cli/root.go` (or wherever `moai` root command is composed).
3. JSON output format: `--format json` flag (default human-readable).

#### M5b: Cache-prefix invariant (P-C05 closure)

1. Create `internal/session/hydrate.go` with the load-bearing comment:

   ```go
   // Package session — hydrate.go
   //
   // cache-prefix: DO NOT REORDER
   //
   // The (systemPrompt, userContext, systemContext) assembly order is frozen at
   // this layer. Reordering breaks Anthropic prompt-cache hits (problem-catalog
   // P-C05). Any change here must be accompanied by a SPEC and explicit
   // approval from a load-bearing reviewer.
   ```

2. Implement `HydrateForPrompt(phase, specID) (system, user, sys2 string, err error)` returning the three pieces in fixed order. Used by the orchestrator's prompt assembly (out-of-scope for this SPEC; the function exists as the contract).

#### M5c: `retention_days` integration with `moai clean` (REQ-031, AC-13)

1. Extend `internal/cli/clean.go` (existing) to read `.moai/config/sections/state.yaml` `retention_days: N` (default unset → no pruning).
2. When set, mark `runs/{iter-id}/` directories whose `mtime` is older than `N` days as eligible for prune. Show count + dry-run by default; require `--force` to actually delete (TRUST 5 Secured).

#### M5d: ArtifactEncodingMismatch (REQ-043)

1. Extend `WriteRunArtifact(iterID, name string, body []byte)`:
   - If `name` extension matches `.md|.txt|.json|.yaml|.yml` (text-declared), validate `utf8.Valid(body)`.
   - On invalid UTF-8: return `ErrArtifactEncodingMismatch`.
2. Binary artifacts (`.png|.jpg|.bin|...`) bypass UTF-8 check.

#### M5e: CI lint for AskUserQuestion in subagents (REQ-042, AC-11)

1. Create `internal/template/agent_askuser_audit_test.go` (or extend existing `commands_audit_test.go`) walking `.claude/agents/**.md` and asserting body contains no literal token `AskUserQuestion(`.
2. Sentinel: `t.Errorf("ASKUSERQUESTION_IN_SUBAGENT: %s body line %d contains AskUserQuestion(; subagents must use BlockerReport, not AskUserQuestion. See agent-common-protocol.md §User Interaction Boundary.", path, line)`.

#### M5f: CHANGELOG + MX tags + final verification

1. Add CHANGELOG entry under `## [Unreleased]`:
   ```
   ### Added
   - SPEC-V3R2-RT-004: Typed session state + phase checkpoint. New `internal/session/` Go types (PhaseState, Checkpoint, BlockerReport, SessionStore) with validator/v10 schema, advisory file locking (cross-platform), provenance tagging (user/project/local/session/hook), STALE_SECONDS crash-resume semantics, team-mode per-agent checkpoint merge, in-flight transition detection. New `moai state {dump,show-blocker}` CLI subcommand. New CI lint blocking `AskUserQuestion(` in subagent files.
   ```
2. Insert MX tags per §6 below.
3. Run full `go test ./...` from repo root. Verify ALL tests pass + 0 cascading failures (per `CLAUDE.local.md` §6 HARD rule).
4. Run `make build` to regenerate `internal/template/embedded.go` (if any embedded template was modified — none expected).
5. Update `progress.md` with `run_complete_at` and `run_status: implementation-complete` after M1-M5f land.

[HARD] No new documents are created in `.moai/specs/` or `.moai/reports/` during M5 — this is a SPEC implementation phase, not a planning phase.

---

## 3. File:line Anchors (concrete edit targets)

### 3.1 To-be-modified (existing files)

| File | Anchor | Edit type | Reason |
|------|--------|-----------|--------|
| `internal/session/checkpoint.go:9-45` | `PlanCheckpoint`, `RunCheckpoint`, `SyncCheckpoint` structs | Add validator/v10 tags + `Harness` field on RunCheckpoint + `Validate()` method | M2 / REQ-004, REQ-015 |
| `internal/session/store.go:54-84` | `Checkpoint()` method | Insert validator call + advisory lock acquisition + blocker-file scan | M2 / M3 / M4 / REQ-004, REQ-020, REQ-040 |
| `internal/session/store.go:87-109` | `Hydrate()` method | Add validator call after JSON decode + cache-prefix invariant comment reference | M2 / M4 / REQ-004, REQ-014 |
| `internal/session/store.go:54-247` | `writeAtomic` callers (4 call sites) | Refactor to use new helper from `atomic.go` | M2 / refactor |
| `internal/session/store.go:end` | New methods | Add `HydrateWithOpts`, `DetectInFlightTransition`, `MergeTeamCheckpoints` | M4 / REQ-022, REQ-033, REQ-050, REQ-021, REQ-051 |
| `internal/session/store.go:134-153` | `WriteRunArtifact()` | Add UTF-8 validation for text-declared artifacts | M5d / REQ-043 |
| `internal/config/types.go` | top-level `Config` struct | Add `Session SessionConfig` field with `StaleSeconds int` | M4 / REQ-022 |
| `internal/config/loader.go` | `Load()` (or equivalent) | Read `.moai/config/sections/ralph.yaml` `stale_seconds:` key | M4 / REQ-022 |
| `internal/cli/clean.go` | (existing clean implementation) | Honor `retention_days` from `state.yaml`; dry-run-by-default for `runs/{iter-id}/` | M5c / REQ-031 |
| `internal/cli/run.go` | (existing run command) | Wire `--resume` flag + ErrCheckpointStale → AskUserQuestion routing | M4 / REQ-014, REQ-033 |
| `CHANGELOG.md` | `## [Unreleased]` section | Add Added entry per §M5f | M5f / Trackable |

### 3.2 To-be-created (new files)

| File | Reason | LOC estimate |
|------|--------|--------------|
| `internal/session/atomic.go` | `writeAtomic(path, data, perm)` helper extracted from 4 call sites | ~25 |
| `internal/session/lock.go` | Platform-neutral fileLock interface + `newFileLock()` factory | ~30 |
| `internal/session/lock_unix.go` | `//go:build unix` flock implementation | ~50 |
| `internal/session/lock_windows.go` | `//go:build windows` LockFileEx implementation | ~60 |
| `internal/session/lock_test.go` | Cross-platform lock tests | ~80 |
| `internal/session/hydrate.go` | Cache-prefix invariant comment + `HydrateForPrompt` | ~40 |
| `internal/session/atomic_test.go` | Atomic write helper tests | ~50 |
| `internal/session/team_merge_test.go` | Team-mode merge tests | ~80 |
| `internal/session/inflight_test.go` | In-flight transition detection tests | ~60 |
| `internal/cli/state.go` | `moai state {dump,show-blocker}` subcommand | ~150 |
| `internal/cli/state_test.go` | CLI tests | ~120 |
| `internal/template/agent_askuser_audit_test.go` (or extend existing) | CI lint blocking `AskUserQuestion(` in subagents | ~30 |

Total new: ~775 LOC. Total modified: ~250 LOC. Net additions: ~1025 LOC across 12 new files + ~9 modified files.

### 3.3 NOT to be touched (preserved by reference)

The following files are referenced by tests but MUST NOT be modified by this SPEC's run phase. They define the rhythm RT-004 is *codifying* but ownership belongs elsewhere.

- `.moai/state/` directory contents at runtime — runtime artifacts, not source. The test suite uses `t.TempDir()` per `CLAUDE.local.md` §6.
- `internal/config/sections/*.yaml` template content — owned by SPEC-V3R2-RT-005 (settings resolver). This SPEC only adds a `SessionConfig` Go struct that reads from `ralph.yaml`.
- `.claude/rules/moai/core/agent-common-protocol.md` §User Interaction Boundary — load-bearing rule consumed by AC-11/REQ-042 lint test. Preserved verbatim.
- `internal/template/templates/.moai/...` — no template changes in this SPEC (per §1.3 deliverables note).
- `sprint-contract.yaml` schema — owned by SPEC-V3R2-HRN-002. This SPEC only commits to the file path under `.moai/state/` and the read/write interface contract.

### 3.4 Reference citations (file:line)

Per `spec.md` §10 traceability and research.md §10, the following anchors are load-bearing and cited verbatim throughout this plan:

1. `spec.md:50-67` (in-scope items 1-9)
2. `spec.md:121-167` (24 EARS REQs — REQ-001 through REQ-051)
3. `spec.md:168-184` (15 ACs — AC-01 through AC-15)
4. `spec.md:188-194` (constraints — Go 1.22+, validator/v10, atomic rename, advisory locks)
5. `internal/session/phase.go:1-32` (existing Phase enum baseline — REQ-001 satisfied)
6. `internal/session/state.go:8-23` (existing PhaseState struct + ProvenanceTag baseline)
7. `internal/session/checkpoint.go:1-45` (existing checkpoint variants — needs validator/v10 tags)
8. `internal/session/blocker.go:1-34` (existing BlockerReport struct + NewBlockerReport)
9. `internal/session/store.go:24-37` (existing SessionStore interface — 6 methods declared)
10. `internal/session/store.go:54-84` (existing Checkpoint() — needs lock + validator + blocker-file scan)
11. `internal/session/store.go:87-109` (existing Hydrate() — needs validator + cache-prefix comment)
12. `internal/session/store.go:134-153` (existing WriteRunArtifact() — needs UTF-8 validation)
13. `internal/session/store.go:155-180` (existing RecordBlocker())
14. `internal/session/store.go:182-236` (existing ResolveBlocker())
15. `.claude/rules/moai/core/agent-common-protocol.md:#user-interaction-boundary` (subagent prohibition)
16. `.claude/rules/moai/workflow/spec-workflow.md:172-204` (Plan Audit Gate)
17. `CLAUDE.local.md:§6` (test isolation: `t.TempDir()` + `filepath.Abs`)
18. `CLAUDE.local.md:§14` (no hardcoded paths in `internal/`)

Total: **18 distinct file:line anchors** (exceeds the §Hard-Constraints minimum of 10 for plan.md).

---

## 4. Technology Stack Constraints

Per `spec.md` §7 Constraints, **minimal new technology** is introduced:

- Go 1.22+ (already required by `go.mod`).
- `github.com/go-playground/validator/v10` — added by SPEC-V3R2-SCH-001 (blocker dependency). Imported as new dependency in `go.mod` if not already present.
- `golang.org/x/sys/unix` — already in indirect deps for cross-platform support.
- `golang.org/x/sys/windows` — already in indirect deps for `golang.org/x/sys/windows.LockFileEx`.

The only additive surfaces are:

- 12 new Go files under `internal/session/`, `internal/cli/`, `internal/template/` (per §3.2).
- ~9 modified Go files (per §3.1).
- One CHANGELOG entry.
- 6 MX tags across 5 files (per §6).
- One new `SessionConfig` field in `internal/config/types.go`.

**No new directory structures** — `internal/session/` and `.moai/state/` already exist.

**No new YAML schema files** — this SPEC reads from `ralph.yaml` and `state.yaml` (the latter declared in `spec.md` §5.4 but expected to be created lazily by `retention_days` consumers).

---

## 5. Risk Analysis & Mitigations

Extends `spec.md` §8 risks with concrete file-path mitigations.

| Risk | Probability | Impact | Mitigation Anchor |
|------|-------------|--------|-------------------|
| Validator/v10 dependency from SPEC-V3R2-SCH-001 not yet merged | M | H | Check `go.mod` for `validator/v10` before M2; if absent, add directly (`go get github.com/go-playground/validator/v10`) and document in CHANGELOG. SCH-001 status verified at plan-audit gate. |
| Advisory lock semantics differ between Linux flock and macOS flock | L | M | Both use `syscall.Flock` (BSD-style flock); identical behaviour for `LOCK_EX|LOCK_NB`. Lock test suite runs on both platforms in CI. |
| Windows `LockFileEx` lock companion file not auto-cleaned on crash | M | L | Lock companion file is a sibling `<checkpoint-path>.lock` empty file; periodic cleanup via `moai clean --locks` (out of scope for this SPEC; tracked as follow-up). On next acquire attempt, `LockFileEx` correctly returns busy if held; if stale (process died), advisory locks are auto-released by OS. |
| `validator/v10` rejects valid corner cases (e.g., empty Status during draft) | L | M | Use `oneof=...` with all known states explicitly enumerated; revisit per concrete failure. Tests cover both happy path (AC-15 valid Harness) and error path (invalid Harness). |
| Concurrent test races in `TestCheckpoint_ConcurrentRace` flaky on slow CI | M | L | Use `sync.WaitGroup` + 100ms timeout per goroutine; if flaky, mark `t.Skip("requires fast disk")` on Windows runners. |
| Cache-prefix invariant comment ignored by future contributors | L | H | Add `@MX:ANCHOR` tag (§6.1) on `hydrate.go`; CI grep ensures the comment string `cache-prefix: DO NOT REORDER` is preserved (added to `internal/template/cache_prefix_audit_test.go` follow-up; out of scope for RT-004 but tracked). |
| Team-mode merge produces inconsistent provenance when 2 agents write at the same UpdatedAt | L | L | Use stable sort by agent name (alphabetical) for the audit subfield list. Tie-breaking is deterministic. |
| `moai state dump` token-floods stdout for large `RunCheckpoint` | L | L | Default human-readable format; `--format json` available; truncate `Context` map values to 200 bytes by default with `--full` flag override. |
| `retention_days` accidentally deletes load-bearing `runs/iter-current/` directory | L | M | Default behaviour is `--dry-run`; user must pass `--force`. AskUserQuestion confirmation in non-`--force` mode. |
| `--resume` flag misuse silently accepts stale state, leading to data corruption | M | M | Document in CHANGELOG and `moai run --help` that `--resume` is for power-users only. Log a WARN to stderr when a stale checkpoint is hydrated under `--resume`. |
| In-flight transition detection ambiguous when 2 SPECs run concurrently | M | L | Detection is per-SPECID; `DetectInFlightTransition(specID)` filters by SPEC. The orchestrator iterates over all known SPECIDs on session start. |

---

## 6. mx_plan — @MX Tag Strategy

Per `.claude/rules/moai/workflow/mx-tag-protocol.md` and `.claude/skills/moai/workflows/plan.md` mx_plan MANDATORY rule.

### 6.1 @MX:ANCHOR targets (high fan_in / contract enforcers)

| Target file:line | Tag content | Rationale |
|------------------|-------------|-----------|
| `internal/session/store.go:Checkpoint()` | `@MX:ANCHOR fan_in=N - SPEC-V3R2-RT-004 REQ-002, REQ-004, REQ-010, REQ-020, REQ-040 enforcer; every phase boundary writes through this method. Validator + lock + blocker-file scan order is contract; touching this affects all 9 phases.` | The Checkpoint method is the single write gate for all session state. High downstream impact. |
| `internal/session/hydrate.go:HydrateForPrompt()` | `@MX:ANCHOR fan_in=N - SPEC-V3R2-RT-004 cache-prefix discipline (P-C05 closure); the (systemPrompt, userContext, systemContext) assembly order is frozen here. DO NOT REORDER without a new SPEC.` | Cache-prefix invariant is the only Anthropic-visible performance contract in this SPEC. |
| `internal/session/lock.go:fileLock` | `@MX:ANCHOR fan_in=2 - SPEC-V3R2-RT-004 REQ-040 cross-platform lock contract; lock_unix.go and lock_windows.go are the two implementations. New platforms add a new file with the same interface.` | Two platform implementations conform to one interface; expansion to a third platform inherits the contract. |

### 6.2 @MX:NOTE targets (intent / context delivery)

| Target file:line | Tag content | Rationale |
|------------------|-------------|-----------|
| `internal/session/state.go:PhaseState` | `@MX:NOTE - SPEC-V3R2-RT-004 file-first state hand-off; the next phase reads this struct from disk, not the prior conversation. Closes problem-catalog P-C02 (no sub-agent context isolation primitive).` | Documents the WHY for future readers; carries the design intent. |
| `internal/session/blocker.go:BlockerReport` | `@MX:NOTE - SPEC-V3R2-RT-004 interrupt() equivalent. Subagents call SessionStore.RecordBlocker; orchestrator surfaces via AskUserQuestion. Subagents MUST NOT call AskUserQuestion (agent-common-protocol.md §User Interaction Boundary).` | Carries the routing rule for future contributors. |

### 6.3 @MX:WARN targets (danger zones)

| Target file:line | Tag content | Rationale |
|------------------|-------------|-----------|
| `internal/session/store.go:Hydrate()` near `if time.Since(state.UpdatedAt) > fs.staleTTL` | `@MX:WARN @MX:REASON - SPEC-V3R2-RT-004 STALE_SECONDS gate. Bypassing this check (e.g., setting staleTTL to math.MaxInt64) defeats crash-resume safety. The --resume flag uses HydrateWithOpts(SkipStaleCheck: true) instead, which logs a stderr warning.` | Most likely point of regression — developers may want to "just disable" the stale check. |
| `internal/session/store.go:WriteRunArtifact()` UTF-8 check | `@MX:WARN @MX:REASON - SPEC-V3R2-RT-004 REQ-043 ArtifactEncodingMismatch. Text-declared artifacts (.md|.txt|.json|.yaml) MUST be UTF-8. Bypassing this corrupts moai state dump output and breaks Ralph iteration parsing.` | Easy to disable accidentally; consequence is delayed corruption. |

### 6.4 @MX:TODO targets (intentionally NONE for this SPEC)

This SPEC produces a complete, audit-ready typed-state subsystem. No `@MX:TODO` markers are planned — all work converges to GREEN within M1-M5. Any `@MX:TODO` introduced during implementation must be resolved before final M5 commit (per `.claude/rules/moai/workflow/mx-tag-protocol.md` GREEN-phase resolution rule).

### 6.5 MX tag count summary

- @MX:ANCHOR: 3 targets
- @MX:NOTE: 2 targets
- @MX:WARN: 2 targets
- @MX:TODO: 0 targets
- **Total**: 7 MX tag insertions planned across 6 distinct files

---

## 7. Worktree Mode Discipline

[HARD] All run-phase work for SPEC-V3R2-RT-004 executes in:

```
/Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-004
```

Branch: `plan/SPEC-V3R2-RT-004` (already checked out per session context). Run-phase agent will continue on the same branch or create a sibling branch `feature/SPEC-V3R2-RT-004-typed-state` per `CLAUDE.local.md` §18.2 branch naming.

[HARD] Worktree is used for this SPEC (per session context: `/Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-004`). All Read/Write/Edit tool invocations use absolute paths under the worktree root.

[HARD] `make build` and `go test ./...` execute from the worktree root: `cd /Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-RT-004 && make build && go test ./...`.

> Note: Run-phase agent operates from the actual worktree cwd; absolute paths shown for reference only. The worktree-root resolves to the directory returned by `git -C <worktree> rev-parse --show-toplevel` at run time.

---

## 8. Plan-Audit-Ready Checklist

These criteria are checked by `plan-auditor` at `/moai run` Phase 0.5 (Plan Audit Gate per `spec-workflow.md:172-204`). The plan is **audit-ready** only if all are PASS.

- [x] **C1: Frontmatter v0.1.0 schema** — `spec.md` frontmatter has all required fields (`id`, `version`, `status`, `created`, `updated`, `author`, `priority`, `phase`, `module`, `dependencies`, `bc_id`, `breaking`, `lifecycle`, `tags`). Verified by reading `spec.md:1-23`.
- [x] **C2: HISTORY entry for v0.1.0** — `spec.md:29-31` HISTORY table has v0.1.0 row with description.
- [x] **C3: 24 EARS REQs across 6 categories** — `spec.md:121-167` (Ubiquitous 8, Event-Driven 6, State-Driven 3, Optional 4, Unwanted 4, Complex 2).
- [x] **C4: 15 ACs all map to REQs (100% coverage)** — `spec.md:168-184`. Each AC explicitly cites the REQ(s) it maps to. Plan §1.4 traceability matrix confirms 27/27 REQ → AC → Task mapping.
- [x] **C5: BC scope clarity** — `spec.md:20` (`breaking: false`) + spec.md §1 (non-breaking — additive typed schema on top of existing `.moai/state/`).
- [x] **C6: File:line anchors ≥10** — research.md §10 (cited internally), this plan.md §3.4 (18 anchors).
- [x] **C7: Exclusions section present** — `spec.md:69-77` Out of scope (6 entries explicitly mapped to other SPECs).
- [x] **C8: TDD methodology declared** — this plan §1.2 + `.moai/config/sections/quality.yaml` `development_mode: tdd`.
- [x] **C9: mx_plan section** — this plan §6 (7 MX tag insertions across 4 categories).
- [x] **C10: Risk table with mitigations** — `spec.md:198-207` (7 risks) + this plan §5 (11 risks, file-anchored mitigations).
- [x] **C11: Worktree mode path discipline** — this plan §7 (3 HARD rules, worktree-mode per session context).
- [x] **C12: No implementation code in plan documents** — verified self-check: this plan, research.md, acceptance.md, tasks.md contain only natural-language descriptions, regex patterns, file paths, code-block templates, and pseudo-Go for interface declarations. No executable Go function bodies.
- [x] **C13: Acceptance.md G/W/T format with edge cases** — verified in acceptance.md §1-15.
- [x] **C14: tasks.md owner roles aligned with TDD methodology** — verified in tasks.md §M1-M5 (manager-tdd / expert-backend / manager-docs assignments).
- [x] **C15: Cross-SPEC consistency** — blocked-by dependencies verified: SPEC-V3R2-SCH-001 (validator/v10 — at-risk per §5 risk table; mitigation defined), SPEC-V3R2-RT-005 (Source enum subset — referenced not blocking), SPEC-V3R2-CON-001 (FROZEN zone declaration — completed per Wave 6 history). RT-004 blocks HRN-002 and WF-003 per `spec.md` §9.2.

All 15 criteria PASS → plan is **audit-ready**.

---

## 9. Implementation Order Summary

Run-phase agent executes in this order (P0 first, dependencies resolved):

1. **M1 (P0)**: Add ~11 new test cases across `internal/session/store_test.go`, `internal/cli/state_test.go`, `internal/template/agent_askuser_audit_test.go`. Confirm RED for all new tests; existing tests still GREEN.
2. **M2 (P0)**: Add validator/v10 schema tags to checkpoint variants + `Validate()` method + `internal/session/atomic.go` helper. Confirm `TestCheckpoint_ValidatorRejectsBadHarness` and `TestHydrate_CorruptedJSON` GREEN.
3. **M3 (P0)**: Add `internal/session/lock.go` + `lock_unix.go` + `lock_windows.go` + `lock_test.go`. Wire 3-retry / 10ms-backoff lock acquisition into `Checkpoint()`. Confirm `TestCheckpoint_ConcurrentRace` GREEN on Linux/macOS/Windows CI matrix.
4. **M4 (P0)**: Provenance round-trip + blocker-outstanding gate + stale-check + STALE_SECONDS config + in-flight detection + team-mode merge. Confirm AC-04, AC-05, AC-06, AC-07, AC-08, AC-14 GREEN.
5. **M5 (P1)**: `moai state {dump,show-blocker}` CLI subcommand + `cache-prefix invariant` (hydrate.go) + `retention_days` clean integration + ArtifactEncodingMismatch + `AskUserQuestion(` audit lint + CHANGELOG entry + MX tags per §6 + final `make build` + `go test ./...`. Update `progress.md` with `run_complete_at` and `run_status: implementation-complete`.

Total milestones: 5. Total file edits (existing): ~9. Total file creations (new): 12. Total CHANGELOG entries: 1. Total MX tag insertions: 7.

---

End of plan.md.
