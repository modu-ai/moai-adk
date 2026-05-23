---
spec_id: SPEC-V3R6-SESSION-HANDOFF-AUTO-001
acceptance_version: "0.1.0"
created: 2026-05-23
updated: 2026-05-23
status: draft
---

# Acceptance Criteria — SPEC-V3R6-SESSION-HANDOFF-AUTO-001

Each AC is a binary (pass/fail) test assertion with a Given-When-Then scenario, an automated test reference (file:func), and a manual reproducer where applicable. ACs are 1:1 traceable to REQ-SHA-NNN in spec.md §C.

## AC-SHA-001 — Pending file location contract (REQ-SHA-001)

**GIVEN** `PersistIfPending` is invoked with `projectDir = <tmpDir>`, `sessionID = "test-session-1"`, `memoryDir = <memTmpDir>`

**WHEN** the function executes

**THEN** the only file path under `<tmpDir>` that the function reads (via `os.Stat` or `os.ReadFile`) is `<tmpDir>/.moai/state/session-handoff/pending.md`

**Automated test**: `internal/hook/handoff/persist_test.go::TestPersistIfPending_ReadsOnlyContractPath` — implemented via the decoy-mtime approach: populate `<tmpDir>` with multiple decoy files at known mtimes BEFORE invoking `PersistIfPending`; AFTER the call, assert each decoy file's mtime is unchanged (verified via `os.Stat(decoy).ModTime().Equal(...)`). This black-box approach avoids injecting a filesystem abstraction into the production `PersistIfPending` signature (which stays `(ctx, sessionID, projectDir, memoryDir string) error` per spec.md §B.1).

**Manual reproducer**: Create `<tmpDir>/.moai/state/session-handoff/pending.md` AND `<tmpDir>/other-file.md`. Invoke `PersistIfPending`. Check `<tmpDir>/other-file.md` mtime is unchanged.

## AC-SHA-002 — Absent pending file no-op (REQ-SHA-002)

**GIVEN** `<tmpDir>/.moai/state/session-handoff/pending.md` does NOT exist

**WHEN** `PersistIfPending` is invoked

**THEN**:
1. Return value is `nil`
2. No `slog` records with severity ≥ Warn are emitted (verified via `slog.NewTextHandler` with `slog.LevelWarn` minimum and a `bytes.Buffer` capture)
3. `<tmpDir>/.moai/state/session-handoff/` directory does NOT exist after the call (function did not create it)

**Automated test**: `internal/hook/handoff/persist_test.go::TestPersistIfPending_AbsentPendingNoOp`

**Manual reproducer**: Create empty `<tmpDir>`. Invoke `PersistIfPending`. Run `ls -la <tmpDir>/.moai/state/`. Assert directory not found.

## AC-SHA-003 — Valid pending writes both files (REQ-SHA-003)

**GIVEN** pending file content (verbatim):
```
---
sprint: wave6
spec: myproj001
status: plan_ready
index_line: "- [Wave 6 myproj001 plan ready](project_wave6_myproj001_plan_ready.md) — short hook"
---
## Next Session Entry Point

```text
ultrathink. SPEC-MYPROJ-001 plan 진입.
applied lessons: project_wave5_complete.

전제 검증:
1) gh pr view 42 → MERGED

실행: /moai run SPEC-MYPROJ-001

머지 후: SPEC-MYPROJ-002
```
```

**WHEN** `PersistIfPending` is invoked with `memoryDir = <memTmpDir>`

**THEN**:
1. `<memTmpDir>/project_wave6_myproj001_plan_ready.md` exists
2. File content is byte-equal to the Markdown body from the heading `## Next Session Entry Point` onward (frontmatter stripped)
3. `<memTmpDir>/MEMORY.md` first non-blank line equals `- [Wave 6 myproj001 plan ready](project_wave6_myproj001_plan_ready.md) — short hook`
4. Return value is `nil`

**Automated test**: `internal/hook/handoff/persist_test.go::TestPersistIfPending_ValidPendingWritesBoth` — subtests: `NoSupersedes`, `WithSupersedes`, `MinimumBody`, `MaximumBody`

**Manual reproducer**: Write the pending content above to `<tmpDir>/.moai/state/session-handoff/pending.md`. Invoke `PersistIfPending`. Run `cat <memTmpDir>/project_wave6_myproj001_plan_ready.md` and `head -1 <memTmpDir>/MEMORY.md`.

## AC-SHA-004 — Malformed frontmatter preserved (REQ-SHA-004; subtest `InvalidFieldFormat` also covers REQ-SHA-011)

**GIVEN** pending file with frontmatter missing the `spec:` field (or unparseable YAML, or invalid field format `sprint: "wave 6"` with space):
```
---
sprint: wave6
status: plan_ready
index_line: "- short hook"
---
## Next Session Entry Point

```text
body...
```
```

**WHEN** `PersistIfPending` is invoked

**THEN**:
1. Return value is `nil`
2. Exactly one `slog.Warn` record emitted with attributes `path=<pendingPath>` and `reason=<descriptive-string>` (e.g., `"missing required frontmatter field: spec"`)
3. Pending file is byte-for-byte unchanged on disk (verified via SHA256 before and after)
4. No memory file is created in `<memTmpDir>`
5. MEMORY.md is unchanged

**Automated test**: `internal/hook/handoff/persist_test.go::TestPersistIfPending_MalformedFrontmatterPreserved` — subtests: `MissingSprint`, `MissingSpec`, `MissingStatus`, `MissingIndexLine`, `UnparseableYAML`, `InvalidFieldFormat`

**Manual reproducer**: Write malformed pending file. Compute `sha256sum <pendingPath>`. Invoke `PersistIfPending`. Re-compute hash and assert equal.

## AC-SHA-005 — Missing heading or fenced block preserved (REQ-SHA-005)

**GIVEN** pending file with valid frontmatter but body missing the `## Next Session Entry Point` heading (or missing fenced ```` ```text ```` block, or fenced with wrong language tag):
```
---
sprint: wave6
spec: myproj001
status: plan_ready
index_line: "- short hook"
---
body without heading...
```

**WHEN** `PersistIfPending` is invoked

**THEN**:
1. Return value is `nil`
2. Exactly one `slog.Warn` record emitted describing the structural defect (e.g., `"missing heading: ## Next Session Entry Point"` or `"missing fenced text block"`)
3. Pending file unchanged on disk
4. No memory file created
5. MEMORY.md unchanged

**Automated test**: `internal/hook/handoff/persist_test.go::TestPersistIfPending_MissingHeadingPreserved` — subtests: `MissingHeading`, `MissingFencedBlock`, `WrongFenceLanguage`

**Manual reproducer**: Write pending file missing heading. Compute hash. Invoke. Assert hash unchanged + slog.Warn captured.

## AC-SHA-006 — Atomic write no partial read (REQ-SHA-006)

**GIVEN** valid pending file AND a goroutine that concurrently performs `os.ReadFile(<memTmpDir>/project_wave6_myproj001_plan_ready.md)` in a tight loop while `PersistIfPending` runs

**WHEN** `PersistIfPending` completes

**THEN** every observation by the concurrent reader is either:
1. `os.IsNotExist(err) == true` (file not yet renamed into place), OR
2. File content is byte-equal to the verbatim Markdown body (file fully written)

The reader MUST NEVER observe partial content (truncated, mid-write).

**Automated test**: `internal/hook/handoff/persist_test.go::TestPersistIfPending_AtomicWriteNoPartialRead` — implemented via `sync.WaitGroup` with 100 reader iterations during 10 sequential `PersistIfPending` calls; race-detector compatible (`go test -race`)

**Manual reproducer**: Difficult to reproduce manually due to timing window. Trust automated test under `-race`.

## AC-SHA-007 — MEMORY.md contention retry (REQ-SHA-007)

**GIVEN** valid pending file AND a test-injected hook that modifies MEMORY.md between the read and write of the prepend operation

**WHEN** `PersistIfPending` retries

**THEN**:
- Subtest `RetrySucceeds`: parallel write stops after 2 iterations → 3rd retry succeeds → MEMORY.md final state contains the new index_line on top
- Subtest `RetryExhausts`: parallel write continues indefinitely → after 3 retries function emits exactly one `slog.Warn` with `reason="memory_md_contention_exhausted"` AND returns `nil`

**Automated test**: `internal/hook/handoff/persist_test.go::TestPersistIfPending_MemoryMdContentionRetry`

**Manual reproducer**: Run two terminal sessions that write to the same `MEMORY.md` simultaneously while `PersistIfPending` runs. Verify either consistent prepend or `slog.Warn` exhaustion.

## AC-SHA-008 — Supersede marker applied (REQ-SHA-008)

**GIVEN** pending file frontmatter `supersedes: project_wave5_old_complete.md` AND existing `<memTmpDir>/MEMORY.md` containing the line:
```
- [Wave 5 old complete](project_wave5_old_complete.md) — prev hook
```

**WHEN** `PersistIfPending` completes successfully

**THEN** the existing line is rewritten to:
```
[SUPERSEDED by project_wave6_myproj001_plan_ready.md] - [Wave 5 old complete](project_wave5_old_complete.md) — prev hook
```

AND the new index_line is prepended to MEMORY.md above the superseded entry.

**Automated test**: `internal/hook/handoff/persist_test.go::TestPersistIfPending_SupersedeMarkerApplied` — subtests: `SupersedesExistingLine`, `SupersedesMissingLine` (no-op, no error), `MultipleMatches` (first match only rewritten)

**Manual reproducer**: Seed MEMORY.md with the old line. Write pending with supersedes field. Invoke. Run `cat <memTmpDir>/MEMORY.md` and verify `[SUPERSEDED by ...]` prefix on old line.

## AC-SHA-009 — No user interaction from hook context (REQ-SHA-009)

**GIVEN** any test scenario above

**WHEN** `PersistIfPending` executes

**THEN**:
1. The function makes zero calls to `AskUserQuestion` (verified via static import audit: `internal/hook/handoff/persist.go` MUST NOT import `mcp__askuser`, `AskUserQuestion`, or any user-prompt-related symbol)
2. Stdout is empty (verified via captured `os.Stdout` redirect during test)
3. Stderr contains only `slog.*` records (no raw prose writes)

**Automated test**: `internal/hook/handoff/persist_test.go::TestPersistIfPending_NoUserInteraction` PLUS static guard via shell:
```bash
grep -r "AskUserQuestion\|mcp__askuser" internal/hook/handoff/ | grep -v "_test.go" | grep -v "^[^:]*:[0-9]*:[ \t]*//"
```
Exit code MUST be 1 (no matches).

**Manual reproducer**: Run the grep above. Confirm exit code 1.

## AC-SHA-010 — Pending file cleaned on success (REQ-SHA-010)

**GIVEN** valid pending file AND `PersistIfPending` completes both memory write AND MEMORY.md update successfully

**WHEN** the function returns

**THEN**:
1. Return value is `nil`
2. `<tmpDir>/.moai/state/session-handoff/pending.md` no longer exists on disk (verified via `os.Stat` returning `os.IsNotExist(err) == true`)
3. The parent directory `<tmpDir>/.moai/state/session-handoff/` still exists (only the file is removed, not the directory)

If removal fails (e.g., permission error), exactly one additional `slog.Warn` is emitted with `reason="cleanup_failed"` but return value is still `nil`.

**Automated test**: `internal/hook/handoff/persist_test.go::TestPersistIfPending_PendingCleanedOnSuccess` — subtests: `NormalCleanup`, `CleanupFailureLogged` (simulated via read-only directory permission)

**Manual reproducer**: Write valid pending file. Invoke `PersistIfPending`. Run `ls <tmpDir>/.moai/state/session-handoff/pending.md`. Assert "No such file".

## Definition of Done

All 10 ACs (AC-SHA-001 through AC-SHA-010) pass via `go test -race ./internal/hook/handoff/...` AND `golangci-lint run ./internal/hook/handoff/...` produces zero new findings AND the static grep sentinel for AC-SHA-009 exits code 1 AND `go test ./internal/hook/...` (full hook package) produces zero regressions from pre-SPEC baseline.

## Quality Gate Criteria (TRUST 5)

- **Tested**: ≥85% coverage on `internal/hook/handoff/` package (per CLAUDE.local.md §6 Coverage Targets)
- **Readable**: All exported symbols have godoc; SPEC ID `SPEC-V3R6-SESSION-HANDOFF-AUTO-001` referenced in `persist.go` doc-comment
- **Unified**: `gofmt -d ./internal/hook/handoff/` produces zero output; imports organized via `goimports`
- **Secured**: No path traversal possible (sprint/spec regex validation per M2 deliverable 4); no user input flows to filesystem without validation
- **Trackable**: Commit message references SPEC ID; `@MX:NOTE: [AUTO]` and `@MX:TODO: [AUTO]` annotations applied where applicable per `.claude/rules/moai/workflow/mx-tag-protocol.md`

## Edge Cases Catalogue

| Edge case | Handling |
|---|---|
| Pending file exists but empty (0 bytes) | REQ-SHA-004 path → slog.Warn `reason="empty_pending_file"` |
| Pending file is a directory, not a regular file | `os.ReadFile` returns error → slog.Warn `reason="pending_not_regular_file"` |
| memoryDir does not exist | slog.Warn `reason="memory_dir_absent"`; function returns `nil` without creating directory (Claude Code owns it) |
| memoryDir is read-only | atomicWriteFile fails → slog.Warn `reason="memory_write_permission_denied"` |
| MEMORY.md absent in memoryDir | Create MEMORY.md with single index_line (no read-modify-write needed); not a contention scenario |
| Pending body contains nested fenced blocks (triple-backtick inside) | Verbatim copy — body content opaque to parser beyond heading+fence detection |
| Pending body contains binary content | Parser does not interpret body bytes; verbatim copy preserves them |
| sessionID parameter is empty string | Function still proceeds; sessionID is logged in slog attributes but not used for path derivation |
| projectDir parameter is empty string | `os.Stat("/.moai/state/...")` → IsNotExist → no-op (treated as absent pending) |
| Sprint/spec field contains `..` (path traversal attempt) | Regex `^[a-z0-9_-]+$` rejects → REQ-SHA-004 path |
