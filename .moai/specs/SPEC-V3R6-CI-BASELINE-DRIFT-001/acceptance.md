---
id: SPEC-V3R6-CI-BASELINE-DRIFT-001
title: "CI baseline drift cleanup — acceptance"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: P3
phase: v3.0.0
module: "internal/cli, internal/config, internal/statusline, internal/cli/wizard, internal/constitution, internal/template, internal/merge, internal/tmux"
lifecycle: spec-anchored
tier: S
tags: "ci, lint, baseline, drift, technical-debt, tier-s, v3.0"
---

# Acceptance Criteria — CI baseline drift cleanup

## Binary Acceptance Criteria

Each AC is binary (PASS/FAIL). The verification command MUST be the exact command listed; output is captured verbatim and compared to the expected pattern.

### AC-CBD-001 — TestStatus PASS (golden match)

**Given** golden files at `internal/cli/testdata/status-{nocolor,light,dark}.golden`
**When** `go test -count=1 -run TestStatus ./internal/cli/...` is executed
**Then** all TestStatus subtests SHALL PASS (exit code 0) with no golden diff reported.

**Verification command**:
```
go test -count=1 -run TestStatus ./internal/cli/... 2>&1 | tail -10
```
**Expected**: Last line matches `ok ` for `internal/cli` package, exit code 0.

### AC-CBD-002 — Lint errcheck = 0

**Given** the codebase at the SPEC's final commit
**When** `golangci-lint run --timeout=2m` is executed
**Then** the output footer SHALL report `* errcheck: 0` or omit the errcheck line entirely.

**Verification command**:
```
golangci-lint run --timeout=2m 2>&1 | grep -E "^\* errcheck:" || echo "errcheck: 0 (line omitted)"
```
**Expected**: `errcheck: 0` or `errcheck: 0 (line omitted)`.

### AC-CBD-003 — Lint ineffassign = 0

**Given** the codebase at the SPEC's final commit
**When** `golangci-lint run --timeout=2m` is executed
**Then** the output footer SHALL report `* ineffassign: 0` or omit the ineffassign line.

**Verification command**:
```
golangci-lint run --timeout=2m 2>&1 | grep -E "^\* ineffassign:" || echo "ineffassign: 0 (line omitted)"
```
**Expected**: `ineffassign: 0` or `ineffassign: 0 (line omitted)`.

### AC-CBD-004 — Lint staticcheck = 0

**Given** the codebase at the SPEC's final commit
**When** `golangci-lint run --timeout=2m` is executed
**Then** the output footer SHALL report `* staticcheck: 0` or omit the staticcheck line.

**Verification command**:
```
golangci-lint run --timeout=2m 2>&1 | grep -E "^\* staticcheck:" || echo "staticcheck: 0 (line omitted)"
```
**Expected**: `staticcheck: 0` or `staticcheck: 0 (line omitted)`.

### AC-CBD-005 — Lint unused ≤ scope-deferred

**Given** the codebase at the SPEC's final commit
**When** `golangci-lint run --timeout=2m` is executed
**Then** every remaining `unused` issue SHALL have a `//nolint:unused // <SPEC-XXX-YYY rationale>` directive on the matching line in source, AND the total unused count SHALL be ≤ the count documented in `plan.md` §D.1 (M2.4) "Scope-defer with `//nolint:unused`" list (initial target: ≤8 — 6 wizard/review + 2 statusline v3, or fewer if branch_protection.go and init_layout.go are also deferred).

**Verification command**:
```
golangci-lint run --timeout=2m 2>&1 | grep -E "\(unused\)" | wc -l
```
**Expected**: Count ≤ 8 (or whatever number plan.md §D.1 final scope decision records). Each remaining entry MUST be paired with a `//nolint:unused` directive verifiable by `grep -A1 "//nolint:unused" <file>`.

### AC-CBD-006 — ConfigManager race-free under -race

**Given** the codebase at the SPEC's final commit
**When** `go test -race -count=1 ./internal/config/...` is executed
**Then** the command SHALL exit with code 0 AND output SHALL NOT contain `race detected during execution of test`.

**Verification command**:
```
go test -race -count=1 ./internal/config/... 2>&1 | tail -20
```
**Expected**: Last line matches `ok 	github.com/modu-ai/moai-adk/internal/config 	<duration>s`. No `FAIL` or `race detected` strings in output.

### AC-CBD-007 — DORMANT notice once-per-process preserved (REQ-MIG003-018)

**Given** the post-fix `internal/config/sunset_notice.go`
**When** `emitSunsetDormantNotice(sectionsDir)` is invoked multiple times in the same process (without `resetSunsetNoticeOnce()`)
**Then** the `slog.Info("SUNSET_CONFIG_DORMANT_NOTICE", ...)` SHALL be called exactly once.

**Verification**: Existing test `TestEmitSunsetDormantNotice*` in `internal/config/` package SHALL PASS unchanged. Command:
```
go test -count=1 -run TestEmitSunsetDormantNotice ./internal/config/...
```
**Expected**: PASS, exit code 0.

If no such existing test, M3 deliverable creates one in `sunset_notice_test.go` that:
1. Calls `emitSunsetDormantNotice(tmpDirWithSunsetYAML)` twice
2. Captures `slog` output via `slog.SetDefault(...)` with a test handler
3. Asserts exactly one `SUNSET_CONFIG_DORMANT_NOTICE` record

### AC-CBD-008 — Full test suite PASS under -race

**Given** the codebase at the SPEC's final commit
**When** `go test -race -count=1 ./...` is executed
**Then** all packages SHALL PASS with exit code 0, no `race detected`, no `FAIL` lines.

**Verification command**:
```
go test -race -count=1 ./... 2>&1 | tail -30
```
**Expected**: No `FAIL` lines, exit code 0. (Pre-existing skipped tests `t.Skip(...)` are acceptable.)

---

## Definition of Done

All 8 binary ACs MUST PASS in the same commit (or final PR head commit). The PR description MUST embed the verbatim verification command output for each AC. M2.4 deferred unused entries MUST be enumerated in the PR description with their follow-up SPEC IDs (created as plan-stub SPECs in the same PR or a sibling PR).

## Edge Cases Documented

- **EC-1**: A lint issue surfaces in unchanged file mid-SPEC. Resolution: per Risk R1, fix in same SPEC if mechanical, else document in `progress.md`.
- **EC-2**: Race reproducer (REQ-CBD-006) added but flaky. Resolution: per Risk R4, mark `t.Skip` with rationale; primary verification remains AC-CBD-006 (full package race).
- **EC-3**: `pkg/version` SoT bumps during SPEC (e.g., to `v3.0.0-rc2`). Resolution: re-run M1 with new value before merge.
- **EC-4**: Removing `unused` symbol breaks an external test or import. Resolution: revert removal, mark as deferred per §D.1 M2.4 with `//nolint:unused` directive.

## Out-of-Scope Reminders (Negative ACs)

The following are explicitly NOT validated in this SPEC's acceptance:
- `moai update --verbose` flag behavior (SPEC-V3R6-UPDATE-NOISE-001 scope)
- TTY/non-TTY progress line rendering (SPEC-V3R6-UPDATE-PROGRESS-001 scope)
- Namespace protection contract (SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001 scope)
- `moai status` behavioral changes (only golden file content updated)
- pkg/version GA release path (rc → GA)
