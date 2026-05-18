---
spec_id: SPEC-V3R2-MIG-002
phase: acceptance
status: draft
---

# Acceptance Criteria — SPEC-V3R2-MIG-002

Each criterion is **binary verifiable** with a single shell command or Go test invocation. Mapping to REQs follows the spec.md ↔ plan.md reconciliation in research.md §9.

## AC-MIG002-A1 — Three-way sync invariant holds

**Given** the merged main branch after MIG-002
**When** `go test ./internal/hook/ -run TestAuditThreeWaySync -v` is invoked
**Then** the test PASSES with zero `HOOK_SYNC_DRIFT` and zero `HOOK_WRAPPER_ORPHAN` diagnostics.

Maps: REQ-MIG002-009, REQ-MIG002-010.

Binary verification:
```bash
go test ./internal/hook/ -run TestAuditThreeWaySync -v 2>&1 | grep -E "PASS|FAIL"
# Expected: contains "PASS:", does not contain "FAIL:"
```

## AC-MIG002-A2 — EventSetup constant retired

**Given** the merged main branch after MIG-002 M2.1
**When** `grep -rn "EventSetup" internal/hook/ internal/cli/` is invoked
**Then** output is empty (zero matches).

Maps: REQ-MIG002-003 (`setupHandler` retire — reconciled to "EventSetup constant retire" per research.md §8).

Binary verification:
```bash
test -z "$(grep -rn 'EventSetup' internal/hook/ internal/cli/ 2>/dev/null)"
# Expected: exit code 0 (no matches)
```

## AC-MIG002-A3 — `moai hook setup` CLI subcommand removed

**Given** the built `moai` binary after MIG-002
**When** `moai hook setup --help` is invoked
**Then** the command fails with cobra's "unknown command" error.

Maps: REQ-MIG002-003.

Binary verification:
```bash
moai hook setup --help 2>&1 | grep -q "unknown command"
# Expected: exit code 0
```

## AC-MIG002-A4 — Per-file Resolution headers preserved

**Given** the merged main branch after MIG-002
**When** `go test ./internal/hook/ -run TestAuditPerFileCategoryHeader -v` is invoked
**Then** all 25 handler files (down from 26 after EventSetup retirement) carry a valid `// Resolution: <CATEGORY>` header.

Maps: REQ-MIG002-001 (3-way sync), implicitly REQ-MIG002-002.

Binary verification:
```bash
go test ./internal/hook/ -run TestAuditPerFileCategoryHeader -v 2>&1 | grep -E "FAIL"
# Expected: empty output (no FAIL lines)
```

## AC-MIG002-A5 — Retired events absent from settings.json.tmpl

**Given** the rendered `settings.json` from `internal/template/templates/.claude/settings.json.tmpl`
**When** parsed as JSON and `Object.keys(result.hooks)` is enumerated
**Then** none of `Notification`, `Elicitation`, `ElicitationResult`, `TaskCreated`, `Setup` appear.

Maps: REQ-MIG002-001, REQ-MIG002-003.

Binary verification:
```bash
go test ./internal/hook/ -run TestAuditRetiredEventsNotInSettings -v 2>&1 | grep -q "PASS"
# Expected: exit code 0
```

## AC-MIG002-A6 — Registration parity (22 native + 4 retired = 26 → 25 post-retirement)

**Given** the merged main branch after MIG-002
**When** `go test ./internal/hook/ -run TestAuditRegistrationParity -v` is invoked
**Then** the test PASSES with `nativeCount == 22` (settings.json keys unchanged) AND the Go handler count adjusts to **25** (was 26; `EventSetup` no longer contributes since the constant is removed; `setupHandler` never existed so no actual handler is removed).

Maps: REQ-MIG002-002 (handler count ≤ 22 reconciled — the spec.md target was off-by-one; correct target is 22 native settings.json keys + 4 retired-obs-only).

Binary verification:
```bash
go test ./internal/hook/ -run TestAuditRegistrationParity -v 2>&1 | grep -q "PASS"
# Expected: exit code 0
```

Note: the spec.md REQ-MIG002-002 said "≤ 22 handlers". That number conflated settings.json keys with Go handlers; the verifiable invariant after reconciliation is "22 settings.json keys + 4 obs-only Go handlers" (the truth-table from `audit_test.go:88-101`).

## AC-MIG002-A7 — User-local settings.json cleanup is idempotent and archival

**Given** a fixture `<projectRoot>/.claude/settings.json` with `Notification` + `Elicitation` entries
**When** `internal/migrate.CleanupUserSettings(projectRoot)` is invoked
**Then**:
1. `<projectRoot>/.claude/settings.json` no longer contains those entries.
2. `<projectRoot>/.moai/archive/hooks/v3.0/migration-<YYYY-MM-DD>.json` is created with the removed entries.
3. Re-invoking `CleanupUserSettings(projectRoot)` produces no further modifications (idempotent).

Maps: REQ-MIG002-011, REQ-MIG002-012, REQ-MIG002-019.

Binary verification:
```bash
go test ./internal/migrate/ -run TestCleanupUserSettings -v 2>&1 | grep -q "PASS"
# Expected: exit code 0
```

## AC-MIG002-A8 — RT-006 work product locked (no stub regression)

**Given** the merged main branch after MIG-002
**When** `go test ./internal/hook/ -run TestAuditNoStubHandlers -v` is invoked
**Then** the test PASSES, confirming each of `subagent_stop.go`, `config_change.go`, `instructions_loaded.go`, `file_changed.go`, `post_tool_failure.go` carries a `UPGRADE` or `FIX` Resolution header (NOT a `STUB`-like marker).

Maps: REQ-MIG002-004, REQ-MIG002-005, REQ-MIG002-006, REQ-MIG002-007, REQ-MIG002-008 (reconciled to characterization tests — the implementation already shipped via RT-006 per research.md §5.2 / §5.3).

Binary verification:
```bash
go test ./internal/hook/ -run TestAuditNoStubHandlers -v 2>&1 | grep -q "PASS"
# Expected: exit code 0
```

## AC-MIG002-A9 — Doctor coverage table summary consistent

**Given** the merged main branch after MIG-002
**When** `go test ./internal/cli/ -run TestDoctorHook -v` is invoked
**Then** all doctor-hook tests PASS, with the counts adjusted to reflect EventSetup retirement (the existing assertion `summary.Remove == 1 (setupHandler)` MUST be updated to either `0` (preferred) or kept at `1` with `setupHandler` row preserved in `coverage_table.go` as a historical record; the choice MUST be consistent between code and test).

Maps: implicit dependency from REQ-MIG002-002, REQ-MIG002-003.

Binary verification:
```bash
go test ./internal/cli/ -run TestDoctorHook -v 2>&1 | grep -E "FAIL"
# Expected: empty output
```

## AC-MIG002-A10 — SPEC frontmatter lint clean

**Given** the post-MIG-002 spec.md with normalized 12-field frontmatter
**When** `moai spec lint --strict .moai/specs/SPEC-V3R2-MIG-002/` is invoked
**Then** the lint reports `✓ No findings`.

Maps: implicit — `.claude/rules/moai/development/spec-frontmatter-schema.md` (canonical 12-field schema).

Binary verification:
```bash
moai spec lint --strict .moai/specs/SPEC-V3R2-MIG-002/ 2>&1 | grep -q "No findings"
# Expected: exit code 0
```

## AC-MIG002-A11 — Coverage thresholds met

**Given** the merged main branch after MIG-002
**When** `go test -cover ./internal/hook/ ./internal/migrate/` is invoked
**Then**:
- `internal/hook/` coverage ≥ 90% (critical-package tier per CLAUDE.local.md §6).
- `internal/migrate/` coverage ≥ 85% (standard tier).

Maps: TRUST 5 Tested pillar; implicit per CLAUDE.local.md §6.

Binary verification:
```bash
go test -cover ./internal/hook/ 2>&1 | grep -E "coverage: ([9][0-9]|100)\.[0-9]+% of statements"
go test -cover ./internal/migrate/ 2>&1 | grep -E "coverage: ([8-9][5-9]|9[0-9]|100)\.[0-9]+% of statements"
# Both expected: exit code 0
```

## AC-MIG002-A12 — Linter clean

**Given** the merged main branch after MIG-002
**When** `golangci-lint run ./internal/hook/... ./internal/cli/... ./internal/migrate/...` is invoked
**Then** zero findings reported.

Maps: TRUST 5 Readable + Unified pillars.

Binary verification:
```bash
golangci-lint run ./internal/hook/... ./internal/cli/... ./internal/migrate/... 2>&1 | tee /tmp/lint.out
test ! -s /tmp/lint.out
# Expected: exit code 0 (empty output)
```

## AC-MIG002-A13 — Race-free under concurrent dispatch

**Given** the merged main branch after MIG-002
**When** `go test -race ./internal/hook/` is invoked
**Then** all tests PASS with zero data-race reports.

Maps: TRUST 5 Secured pillar (concurrency safety).

Binary verification:
```bash
go test -race ./internal/hook/ 2>&1 | grep -E "DATA RACE|FAIL"
# Expected: empty output
```

## AC-MIG002-A14 — Observability opt-in path preserved

**Given** a fixture `system.yaml` with `hook.observability_events: ["notification"]`
**When** the `notificationHandler.Handle(...)` is invoked with that config
**Then** the handler logs the notification (covered by `TestAuditObservabilityWhitelist` subcase `listed_event_returns_true`).

Maps: REQ-MIG002-015 (`retain_noop` opt-in reconciled to `observability_events` opt-in per `internal/hook/observability.go:23-44`).

Binary verification:
```bash
go test ./internal/hook/ -run TestAuditObservabilityWhitelist -v 2>&1 | grep -q "PASS"
# Expected: exit code 0
```

## Definition of Done

The SPEC-V3R2-MIG-002 implementation is complete when ALL of the following hold:

- [ ] AC-MIG002-A1 through AC-MIG002-A14 all verified PASS.
- [ ] Plan-PR merged into main with `plan-auditor` PASS verdict.
- [ ] Run-PR merged into main; commits squashed with `feat(SPEC-V3R2-MIG-002):` prefix per CLAUDE.local.md §18.3.
- [ ] Sync-PR merged into main with CHANGELOG entry under `v3.0.0-rc.X` heading describing: EventSetup retirement + three-way sync invariant + CleanupUserSettings migration step.
- [ ] `.moai/specs/SPEC-V3R2-MIG-002/spec.md` status transitions `draft → implemented → completed`.
- [ ] `moai spec lint --strict` reports `✓ No findings`.
- [ ] Auto-memory project entry `project_v3r2_mig_002_complete.md` created with paste-ready resume + MEMORY.md index updated.

## Edge cases (forecast-only — implementation discretion)

- **Empty user-local settings.json**: `CleanupUserSettings` MUST handle a missing `hooks` key gracefully (no-op, no archive file).
- **Malformed JSON in user-local settings.json**: `CleanupUserSettings` MUST return a wrapped error (`fmt.Errorf("parse settings.json: %w", err)`) and MUST NOT write a partial result.
- **Race against concurrent user edit**: `CleanupUserSettings` writes atomically via temp file + rename; if rename fails, the temp file is removed and the error is returned.
- **Pre-RT-006 build artifact**: if a user's project still has `handle-setup.sh` (from an earlier draft), the migration step MUST detect it via filesystem scan and archive-move it alongside the settings.json cleanup.
- **`hook.RetiredEventNames` empty (future hypothetical)**: if MIG-002 is later extended to retire additional events, the migration logic MUST iterate the exported `RetiredEventNames` slice, not a hard-coded list.

End of acceptance.
