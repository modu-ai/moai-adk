---
id: SPEC-V3R6-SESSION-HANDOFF-AUTO-001
title: "Session handoff auto-persistence (paste-ready resume → memory + MEMORY.md)"
version: "0.1.0"
status: implemented
created: 2026-05-23
updated: 2026-05-23
author: Author Name
priority: P1
phase: "v3.0.0"
module: "internal/hook, internal/hook/handoff"
lifecycle: spec-anchored
tags: "hook, session-handoff, persistence, memory, sprint-2, v3.0"
tier: S
issue_number: null
depends_on: []
related_specs: [SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001, SPEC-V3R6-HOOK-ASYNC-EXPAND-001]
---

# SPEC-V3R6-SESSION-HANDOFF-AUTO-001: Session Handoff Auto-Persistence

## Section A — Pre-existing State Survey (codebase audit prior to scoping)

This section enumerates pre-existing infrastructure facts discovered during plan-phase scoping (2026-05-23). Every fact below was verified against working-tree HEAD on `main` via grep/Read; partial verifications are self-disclosed.

### A.1 — Pre-existing infrastructure facts (5)

1. **`internal/hook/session_end.go` `Handle()` exists with established best-effort cleanup contract** (verified via Read, 717 lines total). The handler performs cleanup steps in strict sequence: `cleanupCurrentSessionTeam(input.SessionID, homeDir)` → `garbageCollectStaleTeams(homeDir)` → `garbageCollectOrphanedTasks(homeDir)` → `cleanupOrphanedTmuxSessions(ctx)` → `clearTmuxSessionEnv(ctx)` → `cleanupGLMSettingsLocal(projectDir)` → `cleanupBogusRootDir(projectDir)` → MX tag validation under `context.WithTimeout(ctx, 4*time.Second)` → conditional `generateSessionSummary(...)`. Doc comment line 59 establishes the contract verbatim: "All cleanup is best-effort: errors are logged with slog.Warn, never returned." The handler always returns `&HookOutput{}, nil` regardless of internal step failures. This is the exact integration template the new persistence call must follow.

2. **`os.UserHomeDir()` is the canonical home-directory resolver used in `session_end.go:66`** and three other `internal/hook/` call sites (`pre_tool.go:274`, `subagent_stop.go:58`, `session_start.go:661`) plus eight `internal/cli/` call sites (verified via grep). The handler already extracts `homeDir` early and bails with a single `slog.Warn` if it fails — line 66-72. The new persistence helper MUST reuse `os.UserHomeDir()` (no abstraction layer exists) and inherit the same "log + return nil" failure semantics.

3. **No existing project-hash resolver exists in `internal/hook/` or `internal/cli/`** for resolving `~/.claude/projects/{hash}/memory/` paths (verified via grep `crypto/sha256`, `projectHash`, `/projects/`). The only sha256 use sites are `internal/cli/design_folder.go` (file content hashing for design drift detection), `internal/cli/hook.go:701,915` (message/prompt hashing for dedupe), `internal/cli/update_noise.go:142-150` (file hash for noise audit), and `internal/hook/quality/change_detector.go:35` (content hash for change detection). **None of these compute the Claude-Code project-hash convention.** The persistence helper MUST either (a) accept the memory directory as an injected parameter (resolved by the orchestrator before writing the pending file), or (b) introduce a NEW project-hash resolver in `internal/hook/handoff/`. Decision deferred to plan.md M1 (recommended: option (a), simpler and avoids re-implementing Claude Code's hash convention which is not documented in our codebase).

4. **`slog.Warn` + best-effort log-only pattern is the universal session_end convention** (verified via Read). Line 67-71 example: `slog.Warn("session_end: could not determine home directory", "error", err)`. Every cleanup function follows the same shape: `slog.Warn("session_end: <human-readable failure description>", "path", <path>, "error", err)` for failures and `slog.Info("session_end: <action verb past tense>", ...)` for successes. The new persistence helper MUST use the prefix `"session_end: handoff: ..."` for slog messages so log greppability matches the existing convention.

5. **Atomic write pattern (temp-file + rename) is NOT pre-existing in `internal/hook/`** — `session_end.go:150` uses `os.WriteFile(reportPath, []byte(content), 0o644)` directly without atomic-rename. This is acceptable for session summary because the file name is session-id-unique. **For MEMORY.md update the persistence helper MUST introduce atomic rename** (`os.CreateTemp` in target dir + `os.Rename`) because parallel sessions can race on the shared MEMORY.md file. For the per-entry memory file, atomic rename is also recommended to prevent partial reads by sibling sessions that scan the memory dir. Reference convention exists in `internal/cli/` (verify in M1).

### A.2 — Trigger detection scope clarification

This SPEC does NOT implement trigger detection. The 5 triggers defined in `.claude/rules/moai/workflow/session-handoff.md` §When To Generate (model-specific threshold / SPEC phase completion / user session-end request / PR creation with pending SPECs / multi-milestone checkpoint) remain orchestrator self-discipline. The hook intervenes only AFTER the orchestrator has already written the pending file to a known buffer path. Absence of the pending file is a no-op — the hook does not generate a resume message from inferred state.

### A.3 — Pending file contract (orchestrator-emitted)

The orchestrator MUST write the pending resume message to `<projectDir>/.moai/state/session-handoff/pending.md` immediately before emitting the fenced text block to the user. The file format is YAML frontmatter + Markdown body:

```yaml
---
sprint: "<sprint-id>"          # e.g., "wave6" or "sprint2"
spec: "<spec-id>"              # e.g., "myproj001" — lowercase, no SPEC- prefix
status: "<status>"             # e.g., "plan_ready" | "complete" | "blocked"
supersedes: "<old-memory-name>" # optional; omit if no prior entry
index_line: "<≤200-char one-line MEMORY.md index entry>"
---
## Next Session Entry Point

```text
ultrathink. <SPEC-ID> <phase> 진입.
applied lessons: <files>.

전제 검증:
1) <precondition 1>
N) <precondition N>

실행: <command>

머지 후: <next>
```
```

The hook reads this file in SessionEnd, validates frontmatter, writes to `<memoryDir>/project_<sprint>_<spec>_<status>.md`, and prepends `index_line` to MEMORY.md.

## Section B — Scope (in-scope vs out-of-scope)

### B.1 — In-scope (Tier S minimal)

1. NEW package `internal/hook/handoff/` containing `persist.go` and `persist_test.go`
2. NEW function `handoff.PersistIfPending(ctx context.Context, sessionID, projectDir, memoryDir string) error` — best-effort persistence with slog warnings, never blocking
3. NEW integration call site in `internal/hook/session_end.go` `Handle()` — invoked after existing best-effort cleanup steps, before the final `slog.Info("session_end: cleanup complete", ...)`
4. NEW pending file contract at `<projectDir>/.moai/state/session-handoff/pending.md` with YAML frontmatter (sprint, spec, status, optional supersedes, index_line) + Markdown body containing the verbatim 6-block resume message
5. NEW frontmatter parser that validates required fields (sprint, spec, status, index_line) + file-name pattern derivation (`project_<sprint>_<spec>_<status>.md`)
6. NEW atomic write helper using `os.CreateTemp` in the target directory + `os.Rename` to prevent partial reads by parallel sessions
7. NEW MEMORY.md prepend helper with read-modify-write + brief retry (max 3) on file-change detection; falls back to slog.Warn on persistent contention
8. NEW supersede marker logic — when frontmatter `supersedes:` is present AND the named file exists in MEMORY.md index, prepend `[SUPERSEDED by <new-file-name>]` to the old entry per Lessons Protocol
9. NEW test file `persist_test.go` with table-driven cases for: absent pending file no-op, present + valid → memory + MEMORY.md written, present + malformed frontmatter → log warning + leave pending untouched, present + missing fenced text block → log warning + leave pending untouched, parallel-write contention smoke test
10. Pending file is left in place after successful persistence for orchestrator follow-up (orchestrator deletes it on next session-start, OR the hook deletes it on success — decision deferred to M3, recommendation: hook deletes on success to prevent stale-state confusion)

### B.2 — Out-of-scope (explicit, bullet form)

- Trigger detection — remains orchestrator self-discipline per `.claude/rules/moai/workflow/session-handoff.md` §When To Generate
- Full 6-block structural validation — minimal sanity (heading presence + fenced block presence) only; structural validation is deferred to a future SPEC if the hook discovers persistent malformed-pending issues in production
- Memory file rotation, archiving, or 50-lesson cap enforcement — existing Lessons Protocol rules apply (`.claude/rules/moai/core/moai-constitution.md`)
- AskUserQuestion integration — hooks cannot prompt users per `.claude/rules/moai/core/agent-common-protocol.md` §User Interaction Boundary; strict best-effort with slog warnings only
- CLI command for manual persistence (`moai handoff persist`, `moai handoff status`, etc.) — deferred to future SPEC if needed
- Backward-compatibility shim for sessions that never write the pending file — none needed; absent file is a no-op matching the existing baseline
- Modification to `.claude/rules/moai/workflow/session-handoff.md` or `.claude/output-styles/moai/moai.md` — Step A of corrective work (output-styles v5.2.0) is already completed in the same orchestrator session prior to this SPEC; this SPEC introduces ONLY Go hook automation (Step B)
- Project-hash directory creation — if `<memoryDir>` does not exist the hook MUST log warning and return; it MUST NOT call `os.MkdirAll(memoryDir, ...)` because the project-hash directory is owned by Claude Code, not by MoAI hooks
- Block 0 (cwd anchoring) handling for L3 `--worktree` resume — the pending file content is opaque to the hook; the orchestrator is responsible for including Block 0 when applicable

## Section C — EARS Requirements (11)

### REQ-SHA-001 (Ubiquitous): Pending file location contract

The `handoff.PersistIfPending` function SHALL read the pending resume message from `<projectDir>/.moai/state/session-handoff/pending.md` and SHALL NOT read from any other path.

### REQ-SHA-002 (Event-Driven): Absent pending file is a no-op

WHEN `<projectDir>/.moai/state/session-handoff/pending.md` does not exist at the time `PersistIfPending` is invoked, the function SHALL return `nil` without producing slog warnings and SHALL NOT create the pending directory.

### REQ-SHA-003 (Event-Driven): Valid pending file triggers persistence

WHEN the pending file exists and contains valid YAML frontmatter with required fields (`sprint`, `spec`, `status`, `index_line`) AND the Markdown body contains a `## Next Session Entry Point` heading followed by a fenced ```` ```text ```` block, the function SHALL write the verbatim Markdown body to `<memoryDir>/project_<sprint>_<spec>_<status>.md` AND SHALL prepend `<index_line>` to `<memoryDir>/MEMORY.md`.

### REQ-SHA-004 (Event-Driven): Malformed frontmatter is logged and pending preserved

WHEN the pending file exists but YAML frontmatter cannot be parsed OR a required field is missing or empty, the function SHALL emit `slog.Warn("session_end: handoff: ...", "path", pendingPath, "reason", <details>)` AND SHALL leave the pending file untouched for orchestrator follow-up AND SHALL return `nil`.

### REQ-SHA-005 (Event-Driven): Missing fenced text block is logged and pending preserved

WHEN the pending file exists with valid frontmatter but the Markdown body does not contain a `## Next Session Entry Point` heading OR does not contain a fenced ```` ```text ```` block, the function SHALL emit `slog.Warn` describing the structural defect AND SHALL leave the pending file untouched AND SHALL return `nil`.

### REQ-SHA-006 (Ubiquitous): Atomic write to prevent partial reads

The function SHALL write the memory file using `os.CreateTemp` in the target directory followed by `os.Rename` such that parallel sessions reading the memory directory never observe a partial file. The function SHALL apply the same atomic-rename pattern to MEMORY.md updates.

### REQ-SHA-007 (Event-Driven): MEMORY.md contention is retried then logged

WHEN MEMORY.md is modified between the read and write of the prepend operation (detected by re-reading and comparing mtime or content hash), the function SHALL retry the read-modify-write up to 3 times. WHEN all 3 retries fail, the function SHALL emit `slog.Warn` AND SHALL return `nil` without aborting.

### REQ-SHA-008 (Event-Driven): Supersede marker is applied to prior entry

WHEN the pending frontmatter contains a non-empty `supersedes:` field AND `MEMORY.md` contains a line referencing that file name, the function SHALL prepend `[SUPERSEDED by <new-file-name>] ` to the matched line per Lessons Protocol in `.claude/rules/moai/core/moai-constitution.md` §Lessons Protocol.

### REQ-SHA-009 (Unwanted): No user prompting from hook context

The function SHALL NOT invoke `AskUserQuestion`, SHALL NOT write to stdout/stderr in a way that would surface to the user, AND SHALL NOT block on user input under any condition. All failure paths SHALL be log-only via `slog.Warn`.

### REQ-SHA-010 (Event-Driven): Successful persistence cleans pending file

WHEN persistence completes successfully (both memory file write AND MEMORY.md update succeed), the function SHALL remove `<projectDir>/.moai/state/session-handoff/pending.md` to prevent re-persistence on the next session-end. WHEN removal fails, the function SHALL emit `slog.Warn` but still SHALL return `nil` (best-effort).

### REQ-SHA-011 (Ubiquitous): Sprint and spec field format validation prevents path injection

The function SHALL enforce the regular expression `^[a-z0-9_-]+$` on both the `sprint:` and `spec:` frontmatter field values before deriving the memory file name. WHEN either field fails the regex check, the function SHALL route the failure through REQ-SHA-004 (malformed-frontmatter path) with `slog.Warn` `reason="invalid_field_format"` AND SHALL NOT proceed to memory write. This requirement provides path-injection resistance: file-name derivation feeds directly into `filepath.Join(memoryDir, ...)` and unvalidated input could escape the memory directory.

## Section D — Acceptance Criteria (10)

Each AC is binary (pass/fail) and 1:1 traceable to a REQ. Test references are package-relative.

### AC-SHA-001 → REQ-SHA-001

**GIVEN** `PersistIfPending` is invoked with `projectDir = <tmpDir>`
**WHEN** the function reads from any path
**THEN** the only file read is `<tmpDir>/.moai/state/session-handoff/pending.md`
**Test**: `internal/hook/handoff/persist_test.go::TestPersistIfPending_ReadsOnlyContractPath` (verifies via decoy-mtime approach: pre-populate `<tmpDir>` with multiple decoy files at known mtimes and assert mtimes are unchanged after the call; black-box, no filesystem abstraction needed)

### AC-SHA-002 → REQ-SHA-002

**GIVEN** `<tmpDir>/.moai/state/session-handoff/pending.md` does NOT exist
**WHEN** `PersistIfPending` is invoked
**THEN** the function returns `nil`, no slog records of severity ≥ Warn are emitted, AND `<tmpDir>/.moai/state/session-handoff/` directory is NOT created
**Test**: `internal/hook/handoff/persist_test.go::TestPersistIfPending_AbsentPendingNoOp`

### AC-SHA-003 → REQ-SHA-003

**GIVEN** pending file exists with valid frontmatter (`sprint: wave6`, `spec: myproj001`, `status: plan_ready`, `index_line: "- [Wave 6 myproj001 plan ready](project_wave6_myproj001_plan_ready.md) — short hook"`) AND valid Markdown body
**WHEN** `PersistIfPending` is invoked with `memoryDir = <memTmpDir>`
**THEN** `<memTmpDir>/project_wave6_myproj001_plan_ready.md` exists with verbatim Markdown body AND `<memTmpDir>/MEMORY.md` first line is the `index_line` value
**Test**: `internal/hook/handoff/persist_test.go::TestPersistIfPending_ValidPendingWritesBoth`

### AC-SHA-004 → REQ-SHA-004

**GIVEN** pending file exists with frontmatter missing `spec:` field
**WHEN** `PersistIfPending` is invoked
**THEN** the function returns `nil`, exactly one `slog.Warn` record is emitted with `path` and `reason` keys, AND the pending file is unchanged on disk (byte-for-byte equal to pre-call state)
**Test**: `internal/hook/handoff/persist_test.go::TestPersistIfPending_MalformedFrontmatterPreserved`

### AC-SHA-005 → REQ-SHA-005

**GIVEN** pending file exists with valid frontmatter but Markdown body missing the `## Next Session Entry Point` heading
**WHEN** `PersistIfPending` is invoked
**THEN** the function returns `nil`, exactly one `slog.Warn` record is emitted describing the structural defect, AND the pending file is unchanged
**Test**: `internal/hook/handoff/persist_test.go::TestPersistIfPending_MissingHeadingPreserved`

### AC-SHA-006 → REQ-SHA-006

**GIVEN** pending file is valid AND a competing goroutine reads `<memTmpDir>/project_*.md` concurrently during the write
**WHEN** `PersistIfPending` completes
**THEN** the concurrent reader either observes the file as absent OR observes it as complete (byte-equal to verbatim body) — never partial
**Test**: `internal/hook/handoff/persist_test.go::TestPersistIfPending_AtomicWriteNoPartialRead` (uses `go test -race`-compatible goroutine race)

### AC-SHA-007 → REQ-SHA-007

**GIVEN** pending file is valid AND MEMORY.md is modified by a parallel write between the read and write of the prepend operation (simulated via a hook callback in test)
**WHEN** `PersistIfPending` retries
**THEN** if the parallel write stops within 3 retries the prepend succeeds; if it continues indefinitely the function emits `slog.Warn` describing contention AND returns `nil`
**Test**: `internal/hook/handoff/persist_test.go::TestPersistIfPending_MemoryMdContentionRetry` (with both retry-succeeds and retry-exhausts subtests)

### AC-SHA-008 → REQ-SHA-008

**GIVEN** pending file frontmatter contains `supersedes: project_wave5_old_complete.md` AND `MEMORY.md` contains a line `- [Wave 5 old complete](project_wave5_old_complete.md) — prev hook`
**WHEN** `PersistIfPending` completes successfully
**THEN** the MEMORY.md line is rewritten to `[SUPERSEDED by project_wave6_myproj001_plan_ready.md] - [Wave 5 old complete](project_wave5_old_complete.md) — prev hook`
**Test**: `internal/hook/handoff/persist_test.go::TestPersistIfPending_SupersedeMarkerApplied`

### AC-SHA-009 → REQ-SHA-009

**GIVEN** any test scenario above
**WHEN** `PersistIfPending` executes
**THEN** the function makes zero calls to `AskUserQuestion` (verified via package-import audit: `internal/hook/handoff/` MUST NOT import any AskUserQuestion-related package) AND produces no stdout output
**Test**: `internal/hook/handoff/persist_test.go::TestPersistIfPending_NoUserInteraction` + static guard: `grep -r "AskUserQuestion\|mcp__askuser" internal/hook/handoff/` MUST return exit code 1 (no matches)

### AC-SHA-010 → REQ-SHA-010

**GIVEN** pending file is valid AND `PersistIfPending` completes successfully
**WHEN** the function returns
**THEN** `<tmpDir>/.moai/state/session-handoff/pending.md` no longer exists on disk
**Test**: `internal/hook/handoff/persist_test.go::TestPersistIfPending_PendingCleanedOnSuccess`

## Section E — Open Questions

The following are open questions for plan-auditor or implementer discretion. Each has a recommended resolution but is not pre-decided.

1. **Project-hash resolver location**: Should the memoryDir be resolved inside `handoff.PersistIfPending` (introducing a new project-hash helper) or be passed as a parameter resolved by the caller in `session_end.go`? **Recommended**: parameter injection — keeps `handoff` package decoupled from Claude Code's undocumented hash convention. `session_end.go` resolves memoryDir via the existing `homeDir` plus a TODO-marked stub that the integration site can later upgrade.

2. **Pending file cleanup on partial failure**: REQ-SHA-010 specifies cleanup only on full success. Should partial success (memory file written but MEMORY.md update failed retry exhaustion) trigger cleanup? **Recommended**: NO — leave pending file in place so a subsequent session-end can retry. Track partial-success scenario explicitly in `slog.Warn` with `"partial": true` key for observability.

3. **Sprint/spec field validation regex**: Should the parser enforce regex on `sprint:` and `spec:` (e.g., lowercase alphanumeric only) to prevent path-injection via crafted frontmatter? **Recommended**: YES — apply `^[a-z0-9_-]+$` to both fields. Frontmatter that fails regex routes to REQ-SHA-004 (malformed) path with `reason: invalid_field_format`. Add as REQ-SHA-011 if plan-auditor concurs.

4. **MEMORY.md size cap interaction**: Lessons Protocol mentions 50-active-lesson cap with archival to `lessons-archive.md`. Should `PersistIfPending` enforce or check this cap? **Recommended**: NO — this SPEC handles project memory entries (`project_*.md`), not lesson entries (`lessons.md`). The cap applies only to lessons. Document this scope boundary in plan.md M2 comment.

5. **Backfill for already-emitted resume messages**: If the user retrofits this hook to an existing project where prior resume messages were never persisted, should `PersistIfPending` offer a one-time backfill mode? **Recommended**: NO — out of scope per §B.2; deferred to a CLI command if demand emerges.
