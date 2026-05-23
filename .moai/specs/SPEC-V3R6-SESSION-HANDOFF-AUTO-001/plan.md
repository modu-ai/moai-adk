---
spec_id: SPEC-V3R6-SESSION-HANDOFF-AUTO-001
plan_version: "0.1.0"
created: 2026-05-23
updated: 2026-05-23
status: draft
---

# Plan — SPEC-V3R6-SESSION-HANDOFF-AUTO-001

This plan-phase document defines the milestone sequence, file paths touched per milestone, and AC mapping for `manager-develop` run-phase execution. Tier S minimal — 4 milestones (M1, M2, M3, M4), no design.md/research.md.

## Milestone Overview

| Milestone | Purpose | AC Coverage | Files Touched (NEW or MOD) |
|---|---|---|---|
| M1 | Handoff package skeleton + path resolvers | AC-SHA-001, AC-SHA-002, AC-SHA-009 (static guard) | NEW: `internal/hook/handoff/persist.go` (skeleton + `PersistIfPending` signature + pending-path resolver + memoryDir parameter passing) |
| M2 | Parser + writer + supersede marker | AC-SHA-003, AC-SHA-004, AC-SHA-005, AC-SHA-006, AC-SHA-008 | MOD: `internal/hook/handoff/persist.go` (frontmatter parser, atomic write helper, MEMORY.md prepend, supersede marker logic) |
| M3 | `session_end.go` integration call site | AC-SHA-002 (integration), AC-SHA-010 | MOD: `internal/hook/session_end.go` (insert `handoff.PersistIfPending` call after existing cleanup steps, before final `slog.Info`) |
| M4 | Test suite + integration smoke | AC-SHA-001 through AC-SHA-010 (all binary verification) | NEW: `internal/hook/handoff/persist_test.go` (table-driven cases, race-compatible concurrency test, sentinel grep guard) |

## M1 — Handoff Package Skeleton + Path Resolvers

### M1 Deliverables

1. NEW file `internal/hook/handoff/persist.go` with package declaration, doc-comment (linking to SPEC ID), and `PersistIfPending(ctx context.Context, sessionID, projectDir, memoryDir string) error` signature
2. NEW unexported helper `pendingPath(projectDir string) string` returning `filepath.Join(projectDir, ".moai", "state", "session-handoff", "pending.md")`
3. Early-return path for absent pending file (REQ-SHA-002): `os.Stat` returns `os.IsNotExist(err)` → return `nil` silently without `slog.Warn` and without creating any directory
4. Stub return `nil` for all other paths (M2 fills in)

### M1 Files Touched

- NEW: `internal/hook/handoff/persist.go` (~50 lines: package, imports, doc, signature, pendingPath helper, IsNotExist short-circuit, stub return)

### M1 AC Mapping

- AC-SHA-001 — pendingPath helper centralizes the contract path; no other paths read
- AC-SHA-002 — IsNotExist short-circuit verified
- AC-SHA-009 (partial) — package has zero imports of AskUserQuestion-related symbols (static guard)

### M1 Exit Criteria

- `go build ./internal/hook/handoff/` succeeds
- `grep -r "AskUserQuestion\|mcp__askuser" internal/hook/handoff/` exits with code 1 (no matches)
- Skeleton compiles without test file (M4 adds tests)

## M2 — Parser + Writer + Supersede Marker

### M2 Deliverables

1. MOD `internal/hook/handoff/persist.go`: NEW unexported helper `parsePending(data []byte) (*pendingEntry, error)` where `pendingEntry` struct has `Sprint string`, `Spec string`, `Status string`, `Supersedes string` (optional), `IndexLine string`, `Body string` fields
2. Frontmatter parser uses `gopkg.in/yaml.v3` (already a project dependency — verify in `go.mod`); on parse failure returns error with descriptive message routed to REQ-SHA-004 path
3. Required-field validation: missing or empty `sprint`/`spec`/`status`/`index_line` returns error routed to REQ-SHA-004 path
4. Field-format validation (REQ-SHA-011 added to spec.md §C per plan-auditor iter-1 SF-1): `sprint` and `spec` MUST match `^[a-z0-9_-]+$`; failures route to REQ-SHA-004 path with `reason: invalid_field_format` (AC-SHA-004 subtest `InvalidFieldFormat` covers this)
5. Body validation: extract substring between `## Next Session Entry Point` heading and end of file; verify presence of fenced ```` ```text ```` ... ```` ``` ```` block; failure routes to REQ-SHA-005 path
6. NEW unexported helper `atomicWriteFile(dir, baseName string, data []byte, perm os.FileMode) error` using `os.CreateTemp(dir, baseName+".tmp.*")` + write + `os.Rename` (REQ-SHA-006)
7. NEW unexported helper `prependToMemoryMD(memoryDir, indexLine string, supersedes string) error` with read-modify-write + max 3 retries on mtime/hash drift (REQ-SHA-007) + supersede marker application (REQ-SHA-008)
8. Wire parser + writer + MEMORY.md updater into `PersistIfPending` main body
9. `slog.Warn` prefix convention: all log messages start with `"session_end: handoff: "` per §A.1.4

### M2 Files Touched

- MOD: `internal/hook/handoff/persist.go` (add ~200 lines: parsePending, atomicWriteFile, prependToMemoryMD, supersede logic, wiring)

### M2 AC Mapping

- AC-SHA-003 — full happy path
- AC-SHA-004 — malformed frontmatter detection + log
- AC-SHA-005 — missing heading/fenced block detection + log
- AC-SHA-006 — atomic write contract
- AC-SHA-008 — supersede marker application

### M2 Exit Criteria

- `go build ./internal/hook/handoff/` succeeds
- `go vet ./internal/hook/handoff/` zero warnings
- Manual smoke: write a valid pending file in `/tmp` and invoke `PersistIfPending` via a tiny `cmd/_smoke/main.go` (deleted after verification, NOT committed) — confirms memory file appears and MEMORY.md prepends

## M3 — session_end.go Integration Call Site

### M3 Deliverables

1. MOD `internal/hook/session_end.go` `Handle()` — insert call to `handoff.PersistIfPending(ctx, input.SessionID, projectDir, memoryDir)` after the existing MX tag validation block (line ~108) and before `slog.Info("session_end: cleanup complete", ...)` (line ~110)
2. memoryDir resolution stub: introduce a NEW unexported helper `resolveMemoryDir(homeDir, projectDir string) (string, error)` in `session_end.go` that returns a TODO placeholder path `filepath.Join(homeDir, ".claude", "projects", "TODO-project-hash", "memory")` with a `@MX:TODO: [AUTO] resolve Claude Code project-hash convention; placeholder fails to produce real memory writes until resolved` annotation. This unblocks integration shipping while §E.1 open question awaits resolution
3. Error handling: `PersistIfPending` return value is discarded (function is best-effort by contract); errors from `resolveMemoryDir` are logged via `slog.Warn("session_end: could not resolve memory directory", "error", err)` and persistence is skipped
4. NEW import in `session_end.go`: `"github.com/modu-ai/moai-adk/internal/hook/handoff"`

### M3 Files Touched

- MOD: `internal/hook/session_end.go` (~20 lines added: resolveMemoryDir helper + integration call + import)

### M3 AC Mapping

- AC-SHA-002 (integration-level) — when pending file absent, session_end still completes successfully with no new warnings
- AC-SHA-010 — cleanup on success verified end-to-end

### M3 Exit Criteria

- `go build ./internal/hook/...` succeeds
- `go test ./internal/hook/...` — pre-existing tests still pass (zero regressions)
- `go vet ./internal/hook/...` zero warnings
- Integration smoke: real session-end invocation with absent pending file completes silently (no new slog.Warn records)

## M4 — Test Suite + Integration Smoke

### M4 Deliverables

1. NEW file `internal/hook/handoff/persist_test.go` with the following table-driven test functions:
   - `TestPersistIfPending_ReadsOnlyContractPath` → AC-SHA-001
   - `TestPersistIfPending_AbsentPendingNoOp` → AC-SHA-002
   - `TestPersistIfPending_ValidPendingWritesBoth` → AC-SHA-003 (with parameterized subtests for variants: no supersedes / with supersedes / minimum-length body / maximum-length body)
   - `TestPersistIfPending_MalformedFrontmatterPreserved` → AC-SHA-004 (subtests: missing sprint / missing spec / missing status / missing index_line / unparseable YAML / invalid field format)
   - `TestPersistIfPending_MissingHeadingPreserved` → AC-SHA-005 (subtests: missing heading / missing fenced block / wrong fence language)
   - `TestPersistIfPending_AtomicWriteNoPartialRead` → AC-SHA-006 (race-detector-compatible; uses goroutine reader during write)
   - `TestPersistIfPending_MemoryMdContentionRetry` → AC-SHA-007 (subtests: retry succeeds within 3 attempts / retry exhausts after 3)
   - `TestPersistIfPending_SupersedeMarkerApplied` → AC-SHA-008 (subtests: supersedes matches existing line / supersedes references missing line / multiple lines match — only first replaced)
   - `TestPersistIfPending_NoUserInteraction` → AC-SHA-009 (verifies via stdout capture + import audit)
   - `TestPersistIfPending_PendingCleanedOnSuccess` → AC-SHA-010
2. All tests use `t.TempDir()` for filesystem isolation per CLAUDE.local.md §6 Test Isolation
3. NEW build-time guard: `cohabitation_guard_test.go`-style static check that `grep -r "AskUserQuestion\|mcp__askuser" internal/hook/handoff/` returns no matches (AC-SHA-009 sentinel)
4. Run with `-race` flag enabled in CI to validate AC-SHA-006 concurrency contract

### M4 Files Touched

- NEW: `internal/hook/handoff/persist_test.go` (~400 lines: 10 test functions, table-driven subtests, race-compatible concurrency)

### M4 AC Mapping

- All 10 ACs (AC-SHA-001 through AC-SHA-010) verified via automated tests

### M4 Exit Criteria

- `go test -race ./internal/hook/handoff/...` all pass
- `go test -cover ./internal/hook/handoff/...` coverage ≥ 85% per CLAUDE.local.md §6 Coverage Targets
- `golangci-lint run ./internal/hook/handoff/...` zero new findings (baseline drift acceptable per §22.4 if pre-existing baseline issues unrelated)
- Static guard: `grep -r "AskUserQuestion\|mcp__askuser" internal/hook/handoff/ | grep -v "_test.go" | grep -v "^[^:]*:[0-9]*:[ \t]*//"` exits code 1

## Technical Approach Summary

- **Best-effort contract**: All failure paths log via `slog.Warn` and return `nil`. This mirrors the existing `session_end.go` cleanup convention (§A.1.1) and ensures hook never blocks session end
- **Atomic writes**: `os.CreateTemp` + `os.Rename` in the target directory prevents parallel-session partial reads (POSIX guarantees rename atomicity within the same filesystem)
- **MEMORY.md retry**: read-modify-write with mtime check + max 3 retries handles the realistic case of parallel sessions persisting near-simultaneously; persistent contention falls back to log-only
- **Supersede marker**: simple in-place line rewrite by file-name match avoids over-engineering; multi-match scenario explicitly chooses "first match wins" (M2 deliverable 7)
- **Project-hash deferral**: §E.1 open question deferred — integration site uses TODO-marked placeholder per M3 deliverable 2; real persistence requires follow-up resolution but skeleton ships safely

## Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| Project-hash placeholder ships without resolution → no real memory writes | High | TODO annotation + plan-auditor §E.1 must address before run-phase begins |
| Parallel sessions race on MEMORY.md → entries lost | Low | 3-retry + atomic rename; persistent contention logged for observability follow-up |
| Pending file written by orchestrator with invalid YAML → silent failure | Medium | REQ-SHA-004 mandates slog.Warn with `reason` key for grep-based debugging; pending file preserved for inspection |
| Hook adds latency to session-end → SessionEnd timeout exceeded | Low | Best-effort + log-only; persistence is bounded to file I/O on local FS (<10ms typical); no network calls |
| Frontmatter sprint/spec injection → path traversal write | Low | M2 deliverable 4 enforces regex `^[a-z0-9_-]+$` on sprint/spec field values |

## Out-of-Scope Reminders (cross-ref §B.2)

- No trigger detection (orchestrator self-discipline)
- No full structural validation (minimal sanity only)
- No CLI command
- No AskUserQuestion (hook context prohibition)
- No project-hash resolver implementation (deferred via §E.1)
