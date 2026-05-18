---
spec_id: SPEC-V3R2-MIG-002
phase: tasks
status: draft
---

# Implementation Tasks — SPEC-V3R2-MIG-002

Each task has explicit REQ ↔ AC traceability and points to the file:line anchors documented in research.md.

## Milestone M1 — RED tests (orphan + drift + stub-regression guards)

### T-MIG002-01 — Add `TestAuditThreeWaySync`

- **Action**: Extend `internal/hook/audit_test.go` with a new test that asserts `Go-registered events ≡ settings.json hook keys ∪ retiredEventNames`.
- **Traceability**: REQ-MIG002-009 (HOOK_SYNC_DRIFT), REQ-MIG002-010 (HOOK_WRAPPER_ORPHAN) → AC-MIG002-A1.
- **Anchors**: `internal/hook/audit_test.go:59-103` (existing parity test as reference), `internal/cli/deps.go:152-185` (Go-side registration source), `internal/template/templates/.claude/settings.json.tmpl:4-369` (settings keys).
- **Acceptance**: test compiles, FAILS at M1 (drift exists pre-M2.1), PASSES after M2.1.

### T-MIG002-02 — Add `TestAuditNoEventSetupOrphan`

- **Action**: Extend `internal/hook/audit_test.go` with a static scan of `types.go` and `internal/cli/hook.go` for the EventSetup symbol and the `"setup"` cobra binding.
- **Traceability**: REQ-MIG002-003 (reconciled to EventSetup retire) → AC-MIG002-A2, AC-MIG002-A3.
- **Anchors**: `internal/hook/types.go:83-85, 139`, `internal/cli/hook.go:58`, `internal/hook/types_test.go:36`, `internal/cli/hook_e2e_test.go:183, 305`.
- **Acceptance**: test compiles, FAILS at M1 (constant still present), PASSES after M2.1.

### T-MIG002-03 — Add `TestAuditNoStubHandlers`

- **Action**: Extend `internal/hook/audit_test.go` with a per-file regex assertion that each of {subagent_stop, config_change, instructions_loaded, file_changed, post_tool_failure}.go carries `// Resolution: (UPGRADE|FIX)`.
- **Traceability**: REQ-MIG002-004, REQ-MIG002-005, REQ-MIG002-006, REQ-MIG002-007, REQ-MIG002-008 (reconciled to characterization tests) → AC-MIG002-A8.
- **Anchors**: `internal/hook/subagent_stop.go:1`, `internal/hook/config_change.go:1`, `internal/hook/instructions_loaded.go:1`, `internal/hook/file_changed.go:1`, `internal/hook/post_tool_failure.go:1`.
- **Acceptance**: test compiles and PASSES immediately (RT-006 already shipped); test locks the work product against future regression.

### T-MIG002-04 — Add `TestRenderRetiredWrappers` (template-side)

- **Action**: Add or extend `internal/template/render_test.go` with assertion that the 4 retired-wrapper templates (`handle-notification.sh.tmpl`, `handle-elicitation.sh.tmpl`, `handle-elicitation-result.sh.tmpl`, `handle-task-created.sh.tmpl`) render successfully and contain a documented RETIRE-OBS-ONLY inline comment.
- **Traceability**: REQ-MIG002-001 (3-way sync intent — the wrappers are retained as forward-compat surface; their presence must be documented) → AC-MIG002-A5 indirectly.
- **Anchors**: `internal/template/templates/.claude/hooks/moai/handle-notification.sh.tmpl`, `handle-elicitation.sh.tmpl`, `handle-elicitation-result.sh.tmpl`, `handle-task-created.sh.tmpl`.
- **Acceptance**: test compiles, FAILS at M1 (no inline RETIRE-OBS-ONLY comment in the tmpls yet), PASSES after M2.2.

## Milestone M2 — GREEN: EventSetup retirement + sync

### T-MIG002-05 — Remove EventSetup constant + active-event entry

- **Action**: Edit `internal/hook/types.go` to delete the `EventSetup` declaration block (lines 83-85) and remove `EventSetup,` from the active-event slice at line 139.
- **Traceability**: REQ-MIG002-003 → AC-MIG002-A2.
- **Anchors**: `internal/hook/types.go:83-85, 139`.
- **Acceptance**: `grep -rn "EventSetup" internal/hook/types.go` returns empty; `go build ./...` succeeds.

### T-MIG002-06 — Remove EventSetup truth-table entry

- **Action**: Edit `internal/hook/types_test.go` to drop the `EventSetup: true,` entry at line 36.
- **Traceability**: REQ-MIG002-003 → AC-MIG002-A2.
- **Anchors**: `internal/hook/types_test.go:36`.
- **Acceptance**: `go test ./internal/hook/ -run TestEventType` passes.

### T-MIG002-07 — Remove `moai hook setup` cobra binding

- **Action**: Edit `internal/cli/hook.go` to delete the `{"setup", "Handle setup event", hook.EventSetup},` entry at line 58.
- **Traceability**: REQ-MIG002-003 → AC-MIG002-A3.
- **Anchors**: `internal/cli/hook.go:58`.
- **Acceptance**: `moai hook setup --help` reports "unknown command"; `go test ./internal/cli/` passes.

### T-MIG002-08 — Remove EventSetup references from e2e test

- **Action**: Edit `internal/cli/hook_e2e_test.go` to drop EventSetup entries at lines 183 and 305.
- **Traceability**: REQ-MIG002-003 → AC-MIG002-A2, AC-MIG002-A3.
- **Anchors**: `internal/cli/hook_e2e_test.go:183, 305`.
- **Acceptance**: `go test ./internal/cli/ -run TestHookE2E` passes.

### T-MIG002-09 — Update doctor coverage table

- **Action**: Edit `internal/hook/coverage_table.go` to drop the setup row; verify `Total` and `Remove` count adjustments. Update `internal/cli/doctor_hook_test.go:18-36, 127-131` expectations to match.
- **Traceability**: REQ-MIG002-002 (reconciled) → AC-MIG002-A6, AC-MIG002-A9.
- **Anchors**: `internal/hook/coverage_table.go`, `internal/cli/doctor_hook_test.go:18-36, 127-131`.
- **Acceptance**: `go test ./internal/cli/ -run TestDoctorHook` passes.

### T-MIG002-10 — Document RETIRE-OBS-ONLY contract in 4 wrapper tmpls

- **Action**: Edit each of `internal/template/templates/.claude/hooks/moai/handle-{notification,elicitation,elicitation-result,task-created}.sh.tmpl` to add a 2-line inline comment: `# RETIRE-OBS-ONLY (SPEC-V3R2-RT-006): not registered in settings.json. Enable via system.yaml hook.observability_events.` near the file header.
- **Traceability**: REQ-MIG002-001 (3-way sync intent / Decision Gate M2.2 Option Keep) → AC-MIG002-A14 indirectly.
- **Anchors**: `internal/template/templates/.claude/hooks/moai/handle-notification.sh.tmpl`, `handle-elicitation.sh.tmpl`, `handle-elicitation-result.sh.tmpl`, `handle-task-created.sh.tmpl`.
- **Acceptance**: `TestRenderRetiredWrappers` PASSES.

### T-MIG002-11 — Run `make build` to regenerate embedded templates

- **Action**: Per CLAUDE.local.md §2 Template-First rule, regenerate `internal/template/embedded.go` after the template wrapper edits.
- **Traceability**: implicit (Template-First discipline) → AC-MIG002-A12.
- **Anchors**: `internal/template/embedded.go` (generated), `Makefile` build target.
- **Acceptance**: `make build` exits 0; `git diff internal/template/embedded.go` shows the embedded-blob delta corresponding to the wrapper comments added in T-MIG002-10.

### T-MIG002-12 — Normalize spec.md frontmatter to canonical 12-field schema

- **Action**: Edit `.moai/specs/SPEC-V3R2-MIG-002/spec.md` frontmatter:
  - Replace `id` (keep), drop legacy `dependencies`/`related_gap`/`related_theme`/`breaking`/`bc_id` keys.
  - Add `title: "Hook Registration Cleanup (orphan EventSetup, 3-way sync invariant, migration step)"`.
  - Keep `created`/`updated` (canonical names).
  - Convert `tags` to comma-separated string: `tags: "hooks, cleanup, orphan, drift, mig, v3"`.
  - Add `lifecycle: spec-anchored` (already present).
  - Ensure `priority: P1` (rewrite from `"P1 High"`).
- **Traceability**: `.claude/rules/moai/development/spec-frontmatter-schema.md` → AC-MIG002-A10.
- **Anchors**: `.moai/specs/SPEC-V3R2-MIG-002/spec.md:1-22`.
- **Acceptance**: `moai spec lint --strict .moai/specs/SPEC-V3R2-MIG-002/` reports `✓ No findings`.

## Milestone M3 — GREEN: migration step

### T-MIG002-13 — Promote `retiredEventNames` to exported package symbol

- **Action**: Move the `retiredEventNames` slice from `internal/hook/audit_test.go:46-51` to a new `internal/hook/retired_events.go` file. Export as `RetiredEventNames []string`. Update `audit_test.go` to reference the exported symbol.
- **Traceability**: enables reuse by `internal/migrate/hook_cleanup.go` (T-MIG002-14).
- **Anchors**: `internal/hook/audit_test.go:46-51`.
- **Acceptance**: `go test ./internal/hook/` continues to pass; new file headed with `// Resolution: KEEP — shared canonical list of retired event names (SPEC-V3R2-RT-006 + SPEC-V3R2-MIG-002).`.

### T-MIG002-14 — Implement `CleanupUserSettings`

- **Action**: Create `internal/migrate/hook_cleanup.go` with `CleanupUserSettings(projectRoot string) error`:
  - Reads `<projectRoot>/.claude/settings.json`.
  - Iterates `hook.RetiredEventNames`; removes matching `hooks.<EventName>` entries.
  - Archives removed entries to `<projectRoot>/.moai/archive/hooks/v3.0/migration-<YYYY-MM-DD>.json`.
  - Writes cleaned settings.json atomically (temp + rename).
- **Traceability**: REQ-MIG002-011, REQ-MIG002-012, REQ-MIG002-019 → AC-MIG002-A7.
- **Anchors**: research.md §9 (drift table — REQ-MIG002-012 row), `.moai/archive/hooks/v3.0/` per spec.md §7.
- **Acceptance**: function compiles; godoc present; error paths return wrapped errors.

### T-MIG002-15 — Unit tests for `CleanupUserSettings`

- **Action**: Create `internal/migrate/hook_cleanup_test.go` table-driven test covering:
  - All 4 retired entries present → all removed + archive file written.
  - 0 retired entries present → no-op + no archive file.
  - Mixed retired + active entries → only retired removed.
  - Malformed JSON → wrapped error + no write.
  - Re-invocation idempotency (subsequent call is a no-op).
- **Traceability**: REQ-MIG002-011, REQ-MIG002-012, REQ-MIG002-019 → AC-MIG002-A7.
- **Anchors**: `internal/migrate/hook_cleanup_test.go` (new), `internal/migrate/testdata/` (fixtures).
- **Acceptance**: `go test ./internal/migrate/ -run TestCleanupUserSettings -v -cover` PASSES with coverage ≥ 85%.

### T-MIG002-16 — Wire CleanupUserSettings into migration framework

- **Action**: Locate the SPEC-V3R2-EXT-004 migration entry point (search `internal/migrate/` for the v2→v3 dispatcher) and add a call to `CleanupUserSettings(projectRoot)` as a step.
- **Traceability**: spec.md §9.1 (blocked-by EXT-004), spec.md §9.2 (blocks MIG-001).
- **Anchors**: `internal/migrate/` (locate dispatcher during implementation).
- **Acceptance**: integration smoke test invoking the full v2→v3 path on a fixture project shows the retired-event cleanup executed.

## Milestone M4 — REFACTOR + verification gates

### T-MIG002-17 — Update `hooks-system.md` rules document

- **Action**: Edit `.claude/rules/moai/core/hooks-system.md` (referenced in spec.md §3) to reflect the post-MIG-002 event roster: 22 native settings.json keys + 4 retired-obs-only + 0 orphans = 26 total Go handlers.
- **Traceability**: documentation alignment with spec.md §3.
- **Anchors**: `.claude/rules/moai/core/hooks-system.md` (verify path during implementation; may be at a different location).
- **Acceptance**: doc reflects the current truth-table; cross-referenced from spec.md.

### T-MIG002-18 — Negative-test the drift detector

- **Action**: In a throwaway branch, inject a fake `EventFoo EventType = "Foo"` constant + register a dummy handler in `deps.go`. Run `TestAuditThreeWaySync`; verify it FAILS with `HOOK_SYNC_DRIFT: Foo`. Discard the branch.
- **Traceability**: REQ-MIG002-009 verification (regression detection actually works) → AC-MIG002-A1.
- **Anchors**: `internal/hook/types.go`, `internal/cli/deps.go:152-185`.
- **Acceptance**: documented in PR description / `progress.md` that the negative test was run and the detector fired correctly.

### T-MIG002-19 — Full verification suite

- **Action**: Run:
  - `go test -race -count=1 ./internal/hook/ ./internal/cli/ ./internal/template/ ./internal/migrate/`
  - `go test -cover ./internal/hook/ ./internal/migrate/`
  - `golangci-lint run ./internal/hook/... ./internal/cli/... ./internal/migrate/...`
  - `go vet ./...`
  - `moai spec lint --strict .moai/specs/SPEC-V3R2-MIG-002/`
- **Traceability**: AC-MIG002-A1 through AC-MIG002-A14.
- **Anchors**: CI surface; CLAUDE.local.md §3 + §6.
- **Acceptance**: all commands exit 0; coverage thresholds met.

### T-MIG002-20 — Update `@MX:ANCHOR` on InitDependencies

- **Action**: Re-verify the `@MX:ANCHOR fan_in=5` annotation at `internal/cli/deps.go:58-62` still reflects post-cleanup reality. If fan_in count changed (e.g., e2e test references dropped), update the annotation.
- **Traceability**: MX-Tag Protocol (Quality Gates).
- **Anchors**: `internal/cli/deps.go:58-62`.
- **Acceptance**: `@MX:REASON` text matches current call-site count verified by `grep -rn "InitDependencies\b" internal/ | wc -l`.

## Traceability Matrix

| REQ | AC | Tasks |
|---|---|---|
| REQ-MIG002-001 | AC-MIG002-A1, A4 | T-MIG002-01, T-MIG002-04, T-MIG002-10 |
| REQ-MIG002-002 | AC-MIG002-A6, A9 | T-MIG002-09 |
| REQ-MIG002-003 | AC-MIG002-A2, A3, A5 | T-MIG002-02, T-MIG002-05, T-MIG002-06, T-MIG002-07, T-MIG002-08 |
| REQ-MIG002-004 | AC-MIG002-A8 | T-MIG002-03 (characterization only — RT-006 already shipped impl) |
| REQ-MIG002-005 | AC-MIG002-A8 | T-MIG002-03 |
| REQ-MIG002-006 | AC-MIG002-A8 | T-MIG002-03 |
| REQ-MIG002-007 | AC-MIG002-A8 | T-MIG002-03 |
| REQ-MIG002-008 | AC-MIG002-A8 | T-MIG002-03 |
| REQ-MIG002-009 | AC-MIG002-A1 | T-MIG002-01, T-MIG002-18 |
| REQ-MIG002-010 | AC-MIG002-A1 | T-MIG002-01 |
| REQ-MIG002-011 | AC-MIG002-A7 | T-MIG002-14, T-MIG002-15 |
| REQ-MIG002-012 | AC-MIG002-A7 | T-MIG002-14, T-MIG002-15, T-MIG002-16 |
| REQ-MIG002-013 | AC-MIG002-A8 | T-MIG002-03 (characterization — RT-006 P-H02 impl at subagent_stop.go:32-98) |
| REQ-MIG002-014 | AC-MIG002-A8 | T-MIG002-03 (characterization — instructions_loaded.go:29-78) |
| REQ-MIG002-015 | AC-MIG002-A14 | T-MIG002-10 (reconciled to observability_events opt-in, not retain_noop) |
| REQ-MIG002-016 | AC-MIG002-A8 | T-MIG002-03 (characterization — file_changed.go:55-114) |
| REQ-MIG002-017 | AC-MIG002-A8 | T-MIG002-03 (characterization — subagent_stop.go:76-86 graceful degradation) |
| REQ-MIG002-018 | AC-MIG002-A8 | T-MIG002-03 (characterization — config_change.go:56-71 reload-fail path) |
| REQ-MIG002-019 | AC-MIG002-A7 | T-MIG002-14 (archive path `.moai/archive/hooks/v3.0/`) |

Cross-cutting:
- **TRUST 5 Tested**: AC-MIG002-A11 → T-MIG002-15, T-MIG002-19.
- **TRUST 5 Readable + Unified**: AC-MIG002-A12 → T-MIG002-19.
- **TRUST 5 Secured**: AC-MIG002-A13 → T-MIG002-19 (race detector).
- **TRUST 5 Trackable**: implicit via `feat(SPEC-V3R2-MIG-002):` commit prefix per CLAUDE.local.md §18.3.

## Execution Order

Sequential within milestones; parallel across non-conflicting tasks:

1. M1: T-MIG002-01, T-MIG002-02, T-MIG002-03, T-MIG002-04 (parallel — all add to test files; no production code mutation).
2. M2 (sequential due to interlocking file edits): T-MIG002-05 → T-MIG002-06 → T-MIG002-07 → T-MIG002-08 → T-MIG002-09 → T-MIG002-10 → T-MIG002-11 → T-MIG002-12.
3. M3: T-MIG002-13 → T-MIG002-14 → T-MIG002-15 → T-MIG002-16 (sequential — each task depends on the previous).
4. M4: T-MIG002-17, T-MIG002-18, T-MIG002-19, T-MIG002-20 (mostly parallel; T-MIG002-19 is the gate).

End of tasks.
