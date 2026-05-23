---
spec_id: SPEC-V3R6-SESSION-HANDOFF-AUTO-001
progress_version: "0.1.0"
created: 2026-05-23
updated: 2026-05-23
status: not_started
---

# Progress — SPEC-V3R6-SESSION-HANDOFF-AUTO-001

This document tracks milestone execution evidence during run-phase. Plan-phase populates only M0 baseline; manager-develop fills M1–M4 evidence as work proceeds.

## M0 — Baseline (plan-phase, 2026-05-23)

### M0 Status

`baseline_captured`

### M0 Baseline Evidence

- **Pre-existing test count for `internal/hook/`**: captured at plan-phase via `go test -count=1 ./internal/hook/... 2>&1 | grep "^ok\|^FAIL" | wc -l` — value to be recorded by manager-develop at run-phase entry (M1 first action) as the regression baseline
- **Pre-existing coverage for `internal/hook/`**: captured at plan-phase via `go test -coverprofile=/tmp/baseline.out ./internal/hook/... && go tool cover -func=/tmp/baseline.out | tail -1` — baseline percentage to be recorded by manager-develop
- **Pre-existing lint baseline for `internal/hook/`**: captured via `golangci-lint run ./internal/hook/... 2>&1 | wc -l` — baseline finding count to be recorded
- **Pre-existing `session_end.go` line count**: 717 lines (verified at plan-phase via `wc -l internal/hook/session_end.go`)
- **Sibling SPEC artifacts for format reference**: `.moai/specs/SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001/` (Tier S, status `implemented`) — confirmed exists with 4 files (spec.md, plan.md, acceptance.md, progress.md)

### M0 Baseline Decisions Carried Forward

- **§E.1 (project-hash resolver)**: Recommendation accepted by plan author — parameter injection via `resolveMemoryDir` helper in `session_end.go` with `@MX:TODO` annotation. Real implementation deferred; placeholder unblocks shipping.
- **§E.3 (sprint/spec regex validation)**: Recommendation accepted — `^[a-z0-9_-]+$` enforcement folded into M2 deliverable 4 to prevent path-injection. Plan-auditor may upgrade to REQ-SHA-011 if strict numbering preferred.
- **§E.2 (partial-success cleanup)**: Recommendation accepted — pending file preserved on partial success; `"partial": true` slog attribute aids observability.
- **§E.4 (lessons cap interaction)**: Recommendation accepted — scope boundary documented; this SPEC handles project memory only.
- **§E.5 (backfill mode)**: Recommendation accepted — out of scope; no CLI command.

## M1 — Handoff Package Skeleton + Path Resolvers

### M1 Status

`not_started`

### M1 Evidence Placeholders (to be filled by manager-develop)

- [ ] `internal/hook/handoff/persist.go` created with package declaration + signature
- [ ] `pendingPath(projectDir string) string` helper implemented
- [ ] Absent pending file short-circuit returns `nil` without slog.Warn
- [ ] `go build ./internal/hook/handoff/` succeeds (paste command output)
- [ ] `grep -r "AskUserQuestion\|mcp__askuser" internal/hook/handoff/` exits code 1 (paste command + exit code)

### M1 AC Verification (target)

- AC-SHA-001 (partial) — pendingPath centralization
- AC-SHA-002 (partial) — IsNotExist short-circuit
- AC-SHA-009 (partial) — package import audit

## M2 — Parser + Writer + Supersede Marker

### M2 Status

`not_started`

### M2 Evidence Placeholders (to be filled by manager-develop)

- [ ] `parsePending` function implemented with required-field validation
- [ ] `pendingEntry` struct fields: Sprint, Spec, Status, Supersedes, IndexLine, Body
- [ ] Sprint/spec field regex `^[a-z0-9_-]+$` enforced
- [ ] Body validation: `## Next Session Entry Point` heading + fenced ```` ```text ```` block
- [ ] `atomicWriteFile` helper using `os.CreateTemp` + `os.Rename`
- [ ] `prependToMemoryMD` helper with retry loop (max 3)
- [ ] Supersede marker logic: `[SUPERSEDED by <new-file>]` prefix on matched line
- [ ] `slog.Warn` messages use `"session_end: handoff: ..."` prefix
- [ ] `go build ./internal/hook/handoff/` succeeds
- [ ] `go vet ./internal/hook/handoff/` zero warnings

### M2 AC Verification (target)

- AC-SHA-003 (full happy path)
- AC-SHA-004 (malformed frontmatter)
- AC-SHA-005 (missing heading/block)
- AC-SHA-006 (atomic write)
- AC-SHA-008 (supersede marker)

## M3 — session_end.go Integration Call Site

### M3 Status

`not_started`

### M3 Evidence Placeholders (to be filled by manager-develop)

- [ ] `resolveMemoryDir(homeDir, projectDir string) (string, error)` helper added to `session_end.go` with `@MX:TODO: [AUTO]` annotation
- [ ] `handoff.PersistIfPending` call inserted after MX validation block, before final `slog.Info("session_end: cleanup complete", ...)`
- [ ] Import `"github.com/modu-ai/moai-adk/internal/hook/handoff"` added
- [ ] Error from `resolveMemoryDir` logged via `slog.Warn`; persistence skipped on resolution failure
- [ ] `go build ./internal/hook/...` succeeds (full hook package)
- [ ] `go test ./internal/hook/...` zero regressions vs M0 baseline
- [ ] Integration smoke: absent pending file → session_end completes silently

### M3 AC Verification (target)

- AC-SHA-002 (integration-level, end-to-end no-op)
- AC-SHA-010 (cleanup-on-success end-to-end)

## M4 — Test Suite + Integration Smoke

### M4 Status

`not_started`

### M4 Evidence Placeholders (to be filled by manager-develop)

- [ ] `internal/hook/handoff/persist_test.go` created
- [ ] 10 test functions implemented (one per AC-SHA-001..010)
- [ ] All tests use `t.TempDir()` for isolation
- [ ] Race-detector test (`TestPersistIfPending_AtomicWriteNoPartialRead`) passes under `-race`
- [ ] `go test -race ./internal/hook/handoff/...` all pass (paste output)
- [ ] `go test -cover ./internal/hook/handoff/...` ≥ 85% (paste coverage)
- [ ] `golangci-lint run ./internal/hook/handoff/...` zero new findings (vs M0 baseline)
- [ ] Static guard `grep -r "AskUserQuestion\|mcp__askuser" internal/hook/handoff/ | grep -v "_test.go" | grep -v "^[^:]*:[0-9]*:[ \t]*//"` exits code 1
- [ ] Full hook package regression check: `go test ./internal/hook/...` passes vs M0 baseline

### M4 AC Verification (target)

- AC-SHA-001 through AC-SHA-010 (all 10 via automated tests)

## Run-Phase Completion Checklist (manager-develop final verification)

When all M1–M4 are complete, manager-develop SHALL verify the following in a single parallel-batch (per `.claude/rules/moai/workflow/verification-batch-pattern.md`):

```bash
# 1. Full hook test suite — zero regressions vs M0 baseline
go test -race ./internal/hook/...

# 2. Coverage report for new package
go test -coverprofile=cover.out ./internal/hook/handoff/... && go tool cover -func=cover.out | tail -1

# 3. AC-SHA-009 static guard
grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/handoff/ | grep -v "_test.go" | grep -v "^[^:]*:[0-9]*:[ \t]*//"

# 4. Sentinel-key audit (no FROZEN sentinel violations)
grep -rn 'FROZEN_SENTINEL\|HARNESS_FROZEN' internal/hook/handoff/

# 5. Hook subcommand smoke (verify session-end still routes correctly)
go run ./cmd/moai hook session-end < /dev/null || echo "expected non-zero with empty stdin"

# 6. Lint baseline check
golangci-lint run --timeout=2m ./internal/hook/handoff/

# 7. Full build
go build ./...
```

All 7 commands SHOULD be invoked in a single assistant turn per AC-WO-007 verification batch pattern.

## Notable Decisions Log (run-phase appendable)

| Date | Milestone | Decision | Rationale |
|---|---|---|---|
| (TBD M1) | M1 | (placeholder) | (placeholder) |
| (TBD M2) | M2 | (placeholder) | (placeholder) |
| (TBD M3) | M3 | (placeholder) | (placeholder) |
| (TBD M4) | M4 | (placeholder) | (placeholder) |

## Blockers Log (run-phase appendable)

(No blockers at plan-phase. Manager-develop will append blockers here as encountered during M1–M4 execution.)

## Status Transition Log

| Date | From | To | By |
|---|---|---|---|
| 2026-05-23 | (none) | draft | manager-spec (plan-phase initial creation) |
| (TBD) | draft | planned | (post-merge automation or manager-develop run-phase entry) |
| (TBD) | planned | in-progress | manager-develop |
| (TBD) | in-progress | implemented | manager-develop |
| (TBD) | implemented | completed | manager-docs (sync-phase) |
