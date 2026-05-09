# SPEC-V3R2-RT-004 Acceptance Criteria — Given/When/Then

> Detailed acceptance scenarios for each AC declared in `spec.md` §6.
> Companion to `spec.md` v0.1.0, `research.md` v0.1.0, `plan.md` v0.1.0, `tasks.md` v0.1.0.

## HISTORY

| Version | Date       | Author                        | Description                                                            |
|---------|------------|-------------------------------|------------------------------------------------------------------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow)  | Initial G/W/T conversion of 15 ACs (AC-V3R2-RT-004-01 through -15)     |

---

## Scope

This document converts each of the 15 ACs from `spec.md` §6 into Given/When/Then format with happy-path + edge-case + test-mapping notation.

Notation:
- **Test mapping** identifies which Go test function (or manual verification step) covers the AC.
- **Sentinel** is the literal error string the test expects on the negative path.

---

## AC-V3R2-RT-004-01 — Plan checkpoint atomically written and validator-valid

Maps to: REQ-V3R2-RT-004-002, REQ-V3R2-RT-004-004, REQ-V3R2-RT-004-010.

### Happy path

- **Given** a fresh `.moai/state/` directory and a valid `PlanCheckpoint{TaskDAG: [...], RiskSummary: "...", SelectedHarness: "standard"}`
- **When** the orchestrator calls `SessionStore.Checkpoint(PhaseState{Phase: PhasePlan, SPECID: "SPEC-V3R2-RT-004", Checkpoint: &plan, UpdatedAt: now})`
- **Then** the file `.moai/state/checkpoint-plan-SPEC-V3R2-RT-004.json` exists
- **And** the file passes JSON syntax validation
- **And** the file passes validator/v10 schema validation (no missing required fields, all `oneof=...` constraints satisfied)
- **And** no `.moai/state/checkpoint-plan-SPEC-V3R2-RT-004.json.tmp` file remains (atomic rename completed)

### Edge case — atomic crash safety

- **Given** the system crashes (process killed) between writing `*.tmp` and `os.Rename`
- **When** the next process inspects `.moai/state/`
- **Then** either the old checkpoint or no `.tmp` artifact remains (depending on crash timing)
- **And** the previous valid `checkpoint-plan-*.json` is unchanged

### Test mapping

- `internal/session/store_test.go::TestCheckpoint_HappyPath` (existing, extends with provenance assertion)
- `internal/session/atomic_test.go::TestWriteAtomic_NoTmpResidue` (new, M2)

---

## AC-V3R2-RT-004-02 — Hydrate returns prior PlanCheckpoint

Maps to: REQ-V3R2-RT-004-011.

### Happy path

- **Given** `.moai/state/checkpoint-plan-SPEC-V3R2-RT-004.json` was written by a prior plan phase with `PlanCheckpoint{Status: "approved", ResearchPath: "research.md"}`
- **When** the run phase begins and calls `SessionStore.Hydrate(PhaseRun, "SPEC-V3R2-RT-004")`
- **Then** the call returns `(*PhaseState, nil)` with a populated `Checkpoint` field
- **Wait, this is wrong** — Hydrate is per-phase, not "what did the prior phase write". Re-reading spec.md §5.2 REQ-011: "the orchestrator SHALL call `SessionStore.Hydrate(Phase, SPECID)` to load the prior-phase checkpoint". So Hydrate(PhaseRun, ...) returns the run-phase state that was previously written, OR the prior plan-phase state if run hasn't run yet.

### Refined Happy path (per skeleton semantics in `internal/session/store.go:87-109`)

- **Given** `.moai/state/checkpoint-plan-SPEC-V3R2-RT-004.json` exists with valid `PlanCheckpoint`
- **When** `SessionStore.Hydrate(PhasePlan, "SPEC-V3R2-RT-004")` is called
- **Then** returned `*PhaseState` has `Phase == PhasePlan`, `SPECID == "SPEC-V3R2-RT-004"`, and `Checkpoint` of concrete type `*PlanCheckpoint`
- **And** the `*PlanCheckpoint` field values match what was previously written (round-trip)

### Edge case — no prior checkpoint

- **Given** no `.moai/state/checkpoint-plan-SPEC-XXX.json` exists
- **When** `SessionStore.Hydrate(PhasePlan, "SPEC-XXX")` is called
- **Then** the call returns `(nil, nil)` (per skeleton `store.go:91-93`: `os.IsNotExist` → no checkpoint)

### Test mapping

- `internal/session/store_test.go::TestHydrate_RoundTrip` (existing baseline; extends with concrete-type assertion)
- `internal/session/store_test.go::TestHydrate_NotExists` (new, M1)

---

## AC-V3R2-RT-004-03 — Subagent records BlockerReport on disk

Maps to: REQ-V3R2-RT-004-012.

### Happy path

- **Given** a subagent encounters missing acceptance-criteria context during the run phase
- **When** the subagent calls `SessionStore.RecordBlocker(BlockerReport{Kind: "missing_input", Message: "...", RequestedAction: "...", Provenance: ProvenanceTag{Source: "session"}})`
- **Then** the file `.moai/state/blocker-run-SPEC-V3R2-RT-004-{timestamp}.json` exists
- **And** the file contains the BlockerReport JSON
- **And** `RecordBlocker()` returns nil (no error)
- **And** the subagent does NOT call `AskUserQuestion` directly (verified by REQ-042 audit lint)

### Edge case — concurrent blocker writes

- **Given** two subagents record blockers within 1 second
- **When** both call `RecordBlocker` with different timestamps
- **Then** both files exist with distinct filenames (timestamp uniqueness)
- **And** neither write fails

### Test mapping

- `internal/session/blocker_test.go::TestRecordBlocker_HappyPath` (existing, extends with file-exists assertion)
- `internal/session/store_test.go::TestRecordBlocker_FilenameFormat` (new, M1) — asserts filename matches `blocker-{phase}-{SPECID}-{timestamp}.json`. Note: the existing skeleton uses `{Kind}-{Source}` not `{phase}-{SPECID}`; M4 corrects this.

---

## AC-V3R2-RT-004-04 — Outstanding blocker prevents Checkpoint advance

Maps to: REQ-V3R2-RT-004-020.

### Happy path

- **Given** an unresolved `BlockerReport` exists at `.moai/state/blocker-run-SPEC-V3R2-RT-004-*.json`
- **When** the orchestrator calls `SessionStore.Checkpoint(PhaseState{Phase: PhaseRun, SPECID: "SPEC-V3R2-RT-004", ...})`
- **Then** `Checkpoint()` returns `ErrBlockerOutstanding`
- **And** no new `.moai/state/checkpoint-run-SPEC-V3R2-RT-004.json` is written

### Edge case — resolved blocker permits advance

- **Given** a previously outstanding blocker has been marked `Resolved: true` via `ResolveBlocker(...)`
- **When** the orchestrator retries `Checkpoint()` for the same phase
- **Then** the call succeeds and the checkpoint is written

### Test mapping

- `internal/session/store_test.go::TestCheckpoint_BlockerOutstanding` (new, M1) — sentinel `ErrBlockerOutstanding`
- `internal/session/store_test.go::TestCheckpoint_AfterBlockerResolved` (new, M1)

---

## AC-V3R2-RT-004-05 — Stale checkpoint triggers AskUserQuestion prompt

Maps to: REQ-V3R2-RT-004-014, REQ-V3R2-RT-004-022.

### Happy path

- **Given** `.moai/state/checkpoint-plan-SPEC-XXX.json` has `UpdatedAt` 2 hours old
- **And** `.moai/config/sections/ralph.yaml` `stale_seconds: 3600` (default)
- **When** the CLI invokes `Hydrate(PhasePlan, "SPEC-XXX")`
- **Then** `Hydrate()` returns `ErrCheckpointStale`
- **And** the orchestrator surfaces the staleness via AskUserQuestion (orchestrator-level, NOT inside `SessionStore` — agent-common-protocol.md HARD rule)

### Edge case — within TTL window

- **Given** the same file with `UpdatedAt` 30 minutes old
- **When** `Hydrate()` is called
- **Then** the call succeeds and the checkpoint is returned (no stale error)

### Test mapping

- `internal/session/store_test.go::TestHydrate_StaleCheckpoint` (new, M1) — uses `time.Now().Add(-2*time.Hour)` for UpdatedAt; sentinel `ErrCheckpointStale`
- `internal/session/store_test.go::TestHydrate_WithinTTL` (new, M1) — uses `time.Now().Add(-30*time.Minute)`; expects no error

---

## AC-V3R2-RT-004-06 — `--resume` flag bypasses staleness prompt

Maps to: REQ-V3R2-RT-004-033.

### Happy path

- **Given** the same stale checkpoint as AC-05
- **And** the user invokes `moai run SPEC-XXX --resume`
- **When** the CLI calls `HydrateWithOpts(PhasePlan, "SPEC-XXX", HydrateOpts{SkipStaleCheck: true})`
- **Then** the staleness check is skipped
- **And** the checkpoint is returned successfully
- **And** a stderr WARN is logged ("hydrating stale checkpoint under --resume; this may be unsafe")

### Edge case — `--resume` with valid checkpoint

- **Given** a fresh checkpoint within TTL
- **When** `HydrateWithOpts(...)` is called with `SkipStaleCheck: true`
- **Then** no stale error and no WARN — flag is a no-op when checkpoint is fresh

### Test mapping

- `internal/session/store_test.go::TestHydrate_ResumeBypass` (new, M1)
- `internal/cli/state_test.go::TestRun_ResumeFlag` (new, M5)

---

## AC-V3R2-RT-004-07 — `moai state dump` prints PhaseState with provenance

Maps to: REQ-V3R2-RT-004-007, REQ-V3R2-RT-004-030.

### Happy path

- **Given** `.moai/state/checkpoint-plan-SPEC-V3R2-RT-004.json` exists with a complete `PhaseState` including `Provenance{Source: "user", Origin: "/Users/goos/.../spec.md", Loaded: ...}`
- **When** the user runs `moai state dump --phase plan --spec SPEC-V3R2-RT-004`
- **Then** stdout contains the `PlanCheckpoint` field values
- **And** stdout contains the `Provenance.Source` value `"user"`
- **And** stdout contains the `Provenance.Origin` value (file path)
- **And** the format is human-readable (multi-line, indented)

### Edge case — `--format json`

- **Given** the same file
- **When** the user runs `moai state dump --phase plan --spec SPEC-V3R2-RT-004 --format json`
- **Then** stdout is valid JSON
- **And** parsing the JSON yields the same `PhaseState` as `Hydrate()` would return

### Edge case — no matching checkpoint

- **Given** no checkpoint file matches the filter
- **When** the user runs `moai state dump --phase plan --spec NON-EXISTENT`
- **Then** the command exits with code 1
- **And** stderr contains `"no checkpoint found for phase=plan spec=NON-EXISTENT"`

### Test mapping

- `internal/cli/state_test.go::TestStateDump_HappyPath` (new, M5)
- `internal/cli/state_test.go::TestStateDump_FormatJSON` (new, M5)
- `internal/cli/state_test.go::TestStateDump_NoMatch` (new, M5)

---

## AC-V3R2-RT-004-08 — Team-mode merge produces SrcSession provenance with audit subfield

Maps to: REQ-V3R2-RT-004-021, REQ-V3R2-RT-004-051.

### Happy path

- **Given** team mode active with 3 teammates, each having written `checkpoint-plan-SPEC-XXX-team-001.json`, `...-team-002.json`, `...-team-003.json`
- **When** the orchestrator calls `MergeTeamCheckpoints("SPEC-XXX", PhasePlan, ["team-001", "team-002", "team-003"])`
- **Then** the returned `PhaseState.Provenance.Source` equals `"session"`
- **And** the audit subfield (Provenance.Origin) lists the 3 per-agent paths comma-joined and sorted alphabetically
- **And** the merged checkpoint reflects the union of per-agent state per the merge algorithm (research.md §7.3)

### Edge case — one teammate has unresolved blocker

- **Given** team-002 wrote a checkpoint AND a blocker file for the same phase/SPECID
- **When** `MergeTeamCheckpoints(...)` is called
- **Then** the merge surfaces the blocker (returns `ErrBlockerOutstanding` with team-002 origin) — bubble-mode short-circuit per REQ-051

### Test mapping

- `internal/session/team_merge_test.go::TestMergeTeamCheckpoints_HappyPath` (new, M1/M4)
- `internal/session/team_merge_test.go::TestMergeTeamCheckpoints_BlockerShortCircuit` (new, M1/M4)

---

## AC-V3R2-RT-004-09 — Corrupted checkpoint returns CheckpointInvalid naming offending field

Maps to: REQ-V3R2-RT-004-041.

### Happy path

- **Given** `.moai/state/checkpoint-run-SPEC-XXX.json` has `RunCheckpoint.Status: "completed"` (invalid; not in `oneof=pass fail partial`)
- **When** `SessionStore.Hydrate(PhaseRun, "SPEC-XXX")` is called
- **Then** the call returns an error wrapping `ErrCheckpointInvalid`
- **And** the error message names the offending field — substring `"Status"` present in `err.Error()`

### Edge case — multiple invalid fields

- **Given** the same file with both `Status: "completed"` AND `Harness: "ultra"` (both invalid)
- **When** `Hydrate()` is called
- **Then** the error names at least one offending field
- **And** the test does NOT require a specific ordering (validator/v10 returns all errors)

### Edge case — JSON syntax error

- **Given** the file is not valid JSON (truncated, malformed)
- **When** `Hydrate()` is called
- **Then** the call returns a JSON unmarshal error (NOT `ErrCheckpointInvalid`; that's reserved for validator failures)
- **And** the error message indicates the JSON parse failure

### Test mapping

- `internal/session/store_test.go::TestHydrate_CorruptedJSON` (new, M1) — asserts `err.Error()` contains `"Status"` substring
- `internal/session/store_test.go::TestHydrate_MalformedJSON` (new, M1)

---

## AC-V3R2-RT-004-10 — Concurrent Checkpoint races; one wins, one fails

Maps to: REQ-V3R2-RT-004-040.

### Happy path

- **Given** a `FileSessionStore` with state directory `tmp/state/`
- **When** 2 goroutines simultaneously call `Checkpoint(PhaseState{...same key...})` for the same `(phase, specID)`
- **Then** at least one call succeeds (returns nil)
- **And** at most one call fails with `ErrCheckpointConcurrent` after 3 retries
- **And** the final on-disk file matches one of the 2 attempted writes (no partial write)

### Edge case — high contention (10 goroutines)

- **Given** 10 goroutines compete for the same checkpoint key
- **When** all call `Checkpoint(...)` simultaneously
- **Then** exactly 1 succeeds first; the remaining 9 each retry up to 3 times; some succeed (sequential) and some fail with `ErrCheckpointConcurrent`
- **And** the final on-disk file is one of the 10 attempted writes

### Test mapping

- `internal/session/store_test.go::TestCheckpoint_ConcurrentRace` (new, M1) — uses `sync.WaitGroup` + 2 goroutines
- `internal/session/lock_test.go::TestFileLock_HighContention` (new, M3) — 10 goroutines stress test

[NOTE] May skip on slow CI runners (Windows): `t.Skip("requires fast disk")` if the timing margin is insufficient.

---

## AC-V3R2-RT-004-11 — CI lint blocks `AskUserQuestion(` in subagent body files

Maps to: REQ-V3R2-RT-004-042.

### Happy path

- **Given** a subagent file at `.claude/agents/<name>.md` containing the literal string `AskUserQuestion(`
- **When** CI runs `go test ./internal/template/ -run TestNoAskUserQuestionInSubagents`
- **Then** the test fails with sentinel `ASKUSERQUESTION_IN_SUBAGENT: <file> body line <N> contains AskUserQuestion(; subagents must use BlockerReport, not AskUserQuestion. See agent-common-protocol.md §User Interaction Boundary.`

### Edge case — orchestrator file is exempt

- **Given** the orchestrator file (e.g., `.claude/skills/moai/SKILL.md`) contains `AskUserQuestion(`
- **When** the audit test runs
- **Then** the orchestrator file is excluded from the audit (only `.claude/agents/**.md` is scanned)

### Edge case — fenced code blocks

- **Given** a subagent file contains `AskUserQuestion(` only inside a markdown fenced code block (illustrating a forbidden pattern)
- **When** the audit test runs
- **Then** the test still fails (sentinel triggers on raw substring; documentation must use a non-literal placeholder like `Ask​UserQuestion(` with zero-width separator OR use indirect phrasing)

### Test mapping

- `internal/template/agent_askuser_audit_test.go::TestNoAskUserQuestionInSubagents` (new, M5)

---

## AC-V3R2-RT-004-12 — Ralph iteration writes prompt.md/response.md/artifacts under runs/

Maps to: REQ-V3R2-RT-004-015.

### Happy path

- **Given** a Ralph iteration with `iter-007`
- **When** the runner calls:
  - `WriteRunArtifact("iter-007", "prompt.md", []byte("..."))`
  - `WriteRunArtifact("iter-007", "response.md", []byte("..."))`
- **Then** files exist at:
  - `.moai/state/runs/iter-007/prompt.md`
  - `.moai/state/runs/iter-007/response.md`
- **And** file contents match the bytes passed
- **And** atomic-rename guaranteed (no `.tmp` residue)

### Edge case — binary artifact

- **Given** the same iteration writes `WriteRunArtifact("iter-007", "screenshot.png", pngBytes)`
- **When** the call completes
- **Then** the file `.moai/state/runs/iter-007/screenshot.png` exists with the byte content
- **And** UTF-8 validation is NOT applied (binary extension exempt)

### Edge case — text-declared artifact with invalid UTF-8

- **Given** the runner calls `WriteRunArtifact("iter-007", "prompt.md", invalidUTF8Bytes)`
- **When** the call is processed
- **Then** the call returns `ErrArtifactEncodingMismatch`
- **And** no file is created (atomic abort)

### Test mapping

- `internal/session/store_test.go::TestWriteRunArtifact_HappyPath` (existing, extends with content assertion)
- `internal/session/store_test.go::TestWriteRunArtifact_BinaryUnderArtifacts` (new, M1)
- `internal/session/store_test.go::TestWriteRunArtifact_TextEncodingMismatch` (new, M1) — sentinel `ErrArtifactEncodingMismatch`

---

## AC-V3R2-RT-004-13 — `moai clean` honors `retention_days` for runs/

Maps to: REQ-V3R2-RT-004-031.

### Happy path

- **Given** `.moai/config/sections/state.yaml` `retention_days: 14`
- **And** `.moai/state/runs/iter-old/` exists with `mtime` 30 days old
- **And** `.moai/state/runs/iter-new/` exists with `mtime` 1 day old
- **When** the user runs `moai clean --dry-run`
- **Then** stdout reports that `iter-old/` is eligible for prune
- **And** stdout reports that `iter-new/` is preserved
- **And** no files are actually deleted (dry-run default)

### Edge case — `--force` actually deletes

- **Given** the same setup
- **When** the user runs `moai clean --force`
- **Then** the AskUserQuestion confirmation appears
- **And** on user confirm, `iter-old/` is deleted; `iter-new/` is preserved

### Edge case — `retention_days` unset

- **Given** `state.yaml` does not declare `retention_days`
- **When** `moai clean` is invoked
- **Then** no `runs/` directory is marked for prune (default behaviour: opt-in)

### Test mapping

- `internal/cli/clean_test.go::TestClean_RetentionDryRun` (new, M5c) — uses `t.TempDir()` + `os.Chtimes` to set mtime
- `internal/cli/clean_test.go::TestClean_RetentionForce` (new, M5c)
- `internal/cli/clean_test.go::TestClean_RetentionUnset` (new, M5c)

---

## AC-V3R2-RT-004-14 — In-flight transition detected on session start

Maps to: REQ-V3R2-RT-004-050.

### Happy path

- **Given** `.moai/state/checkpoint-plan-SPEC-XXX.json` exists (plan completed)
- **And** no `.moai/state/checkpoint-run-SPEC-XXX.json` exists (run not yet completed)
- **When** the orchestrator calls `DetectInFlightTransition("SPEC-XXX")` on session start
- **Then** the call returns `(fromPhase: PhasePlan, toPhase: PhaseRun, found: true, err: nil)`

### Edge case — no in-flight transition

- **Given** all phase checkpoints for SPEC-XXX are complete (plan + run + sync all exist)
- **When** `DetectInFlightTransition("SPEC-XXX")` is called
- **Then** the call returns `(found: false, err: nil)`

### Edge case — multiple SPECs concurrently in-flight

- **Given** SPEC-XXX has plan complete + run pending; SPEC-YYY has run complete + sync pending
- **When** the orchestrator iterates over all known SPECIDs
- **Then** each call returns the correct (from, to) pair for that SPECID

### Test mapping

- `internal/session/inflight_test.go::TestDetectInFlightTransition_HappyPath` (new, M1/M4)
- `internal/session/inflight_test.go::TestDetectInFlightTransition_NoInFlight` (new, M1/M4)
- `internal/session/inflight_test.go::TestDetectInFlightTransition_MultipleSPECs` (new, M1/M4)

---

## AC-V3R2-RT-004-15 — Validator/v10 rejects invalid Harness value

Maps to: REQ-V3R2-RT-004-004.

### Happy path

- **Given** `RunCheckpoint{SPECID: "SPEC-XXX", Status: "pass", Harness: "ultra", ...}` (invalid Harness)
- **When** `SessionStore.Checkpoint(PhaseState{Phase: PhaseRun, ..., Checkpoint: &rc})` is called
- **Then** the call returns an error wrapping `ErrCheckpointInvalid`
- **And** the error message names the offending field — substring `"Harness"` present in `err.Error()`
- **And** no file is written to disk

### Edge case — valid Harness value

- **Given** the same RunCheckpoint with `Harness: "thorough"` (valid)
- **When** `Checkpoint()` is called
- **Then** the call succeeds and the file is written

### Edge case — empty Harness

- **Given** `RunCheckpoint{Harness: ""}` (empty; required field)
- **When** `Checkpoint()` is called
- **Then** the call returns validator error naming `Harness` with `"required"` constraint

### Test mapping

- `internal/session/store_test.go::TestCheckpoint_ValidatorRejectsBadHarness` (new, M1) — sentinel substring `"Harness"`
- `internal/session/store_test.go::TestCheckpoint_ValidatorAcceptsGoodHarness` (new, M1)
- `internal/session/store_test.go::TestCheckpoint_ValidatorRejectsEmptyHarness` (new, M1)

---

## Summary table — AC → REQ → Test

| AC | REQs covered | Test files |
|----|--------------|------------|
| AC-01 | REQ-002, REQ-004, REQ-010 | `store_test.go::TestCheckpoint_HappyPath`, `atomic_test.go::TestWriteAtomic_NoTmpResidue` |
| AC-02 | REQ-011 | `store_test.go::TestHydrate_RoundTrip`, `TestHydrate_NotExists` |
| AC-03 | REQ-012 | `blocker_test.go::TestRecordBlocker_HappyPath`, `store_test.go::TestRecordBlocker_FilenameFormat` |
| AC-04 | REQ-020 | `store_test.go::TestCheckpoint_BlockerOutstanding`, `TestCheckpoint_AfterBlockerResolved` |
| AC-05 | REQ-014, REQ-022 | `store_test.go::TestHydrate_StaleCheckpoint`, `TestHydrate_WithinTTL` |
| AC-06 | REQ-033 | `store_test.go::TestHydrate_ResumeBypass`, `cli/state_test.go::TestRun_ResumeFlag` |
| AC-07 | REQ-007, REQ-030, REQ-005 | `cli/state_test.go::TestStateDump_*` (3 cases) |
| AC-08 | REQ-021, REQ-051 | `team_merge_test.go::TestMergeTeamCheckpoints_*` (2 cases) |
| AC-09 | REQ-041 | `store_test.go::TestHydrate_CorruptedJSON`, `TestHydrate_MalformedJSON` |
| AC-10 | REQ-040 | `store_test.go::TestCheckpoint_ConcurrentRace`, `lock_test.go::TestFileLock_HighContention` |
| AC-11 | REQ-042 | `template/agent_askuser_audit_test.go::TestNoAskUserQuestionInSubagents` |
| AC-12 | REQ-015, REQ-043 | `store_test.go::TestWriteRunArtifact_*` (3 cases) |
| AC-13 | REQ-031 | `cli/clean_test.go::TestClean_Retention*` (3 cases) |
| AC-14 | REQ-050 | `inflight_test.go::TestDetectInFlightTransition_*` (3 cases) |
| AC-15 | REQ-004 | `store_test.go::TestCheckpoint_Validator*` (3 cases) |

Total new test functions: **~32 across 6 new test files + 4 existing files extended**.

---

## Definition of Done

This SPEC is considered done when ALL of the following are true:

1. All 15 ACs above pass under `go test ./internal/session/ ./internal/cli/ ./internal/template/`.
2. Full `go test ./...` from the worktree root passes with zero failures and zero cascading regressions.
3. `make build` succeeds and `internal/template/embedded.go` regenerates cleanly.
4. `go vet ./...` and `golangci-lint run` pass with zero warnings.
5. `progress.md` is updated with `run_complete_at: <timestamp>` and `run_status: implementation-complete`.
6. CHANGELOG entry is present under `## [Unreleased] / ### Added`.
7. 7 MX tags are inserted per `plan.md` §6 (3 ANCHOR, 2 NOTE, 2 WARN).
8. The PR opened by `manager-git` has all required CI checks green (Lint, Test ubuntu/macos/windows, Build all 5, CodeQL).

---

End of acceptance.md.
