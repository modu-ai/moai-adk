---
spec_id: SPEC-V3R2-MIG-002
phase: plan
status: audit-ready
plan_complete_at: 2026-05-18
---

# Implementation Plan — SPEC-V3R2-MIG-002

## 1. Context

The original spec.md was authored 2026-04-23 against the pre-RT-006 state of the hook registry. Between authoring and this plan-phase (2026-05-18), `SPEC-V3R2-RT-006` shipped the per-handler resolution program: 4 retirements (RETIRE-OBS-ONLY), 5 stub→real upgrades (UPGRADE), the `subagentStop` tmux-pane-kill fix (FIX), and the per-file `// Resolution: <CATEGORY>` header convention. See research.md §9 for the drift table.

Consequently most of spec.md REQ-MIG002-004 through REQ-MIG002-008 are **already shipped** in main. The residual work for MIG-002 is now:

- Retire the **EventSetup orphan** (constant + CLI binding + tests) — research.md §8.
- Add a **3-way sync invariant test** preventing future Go ↔ wrapper ↔ settings.json drift — research.md §6, §11, §14.
- Decide and execute the **retired-event wrapper template policy** — research.md §4, §14.
- Provide a **migration path** for user-local settings.json upgrading from pre-RT-006 v3 builds — research.md §14, spec.md REQ-MIG002-012.
- Normalize the SPEC's **own frontmatter** to the canonical 12-field schema (`.claude/rules/moai/development/spec-frontmatter-schema.md`).

This plan turns the obsolete REQs into characterization tests where the implementation already exists, and codifies the genuine residual work as M1–M4 milestones.

## 2. Approach

TDD across M1–M4. Each milestone has an explicit RED (failing test capturing intent) → GREEN (minimal change to pass) → REFACTOR (cleanup + invariant check) loop.

We do NOT re-implement subagent_stop, config_change, instructions_loaded, file_changed, or post_tool_failure — those are RT-006 work product. We DO add characterization tests that lock the current behavior so MIG-002's regression surface includes them.

## 3. Milestones

### M1 — RED: orphan + drift detection tests

**Priority: High**

Failing tests that name the residual cleanup surface. All tests in this milestone MUST fail at M1 entry and pass at M2 exit.

Deliverables:
- `internal/hook/audit_test.go` — extend with `TestAuditThreeWaySync`:
  - Collects Go-registered events by parsing `internal/cli/deps.go` registrations (or, simpler, by reading from `Dependencies.HookRegistry.Handlers()` after `InitDependencies()`).
  - Collects settings.json hook keys from `internal/template/templates/.claude/settings.json.tmpl` (parse with `gopkg.in/yaml.v3` after rendering, OR scan with a regex anchored on `"Event":` style keys).
  - Asserts: `(Go events) == (settings.json keys) ∪ retiredEventNames`.
  - On mismatch: report `HOOK_SYNC_DRIFT: <event-name>` for Go-only entries; `HOOK_WRAPPER_ORPHAN: <event-name>` for settings-only entries.
- `internal/hook/audit_test.go` — extend with `TestAuditNoEventSetupOrphan`:
  - Asserts `EventSetup` constant does NOT exist in `hook` package (compile-time via `_ = hook.EventSetup` reference being absent OR runtime via parsing `types.go`).
  - Asserts `moai hook setup` cobra subcommand binding does NOT exist (parse `internal/cli/hook.go` for the `"setup"` entry).
- `internal/hook/audit_test.go` — extend with `TestAuditNoStubHandlers`:
  - For each of {subagent_stop, config_change, instructions_loaded, file_changed, post_tool_failure}, assert the file's `// Resolution:` header is one of {`UPGRADE`, `FIX`}, NOT a hypothetical `STUB` value.
  - Locks RT-006 work product against silent regression.
- `internal/template/render_test.go` (or extend an existing template renderer test) — assert that the 4 retired-event wrapper templates' fate matches the wrapper-ship decision (see M2 §Decision Gate).

Verification (M1 exit):
- `go test ./internal/hook/ -run TestAuditThreeWaySync` — FAIL (drift exists per research.md §9).
- `go test ./internal/hook/ -run TestAuditNoEventSetupOrphan` — FAIL (EventSetup still present per research.md §8).
- `go test ./internal/hook/ -run TestAuditNoStubHandlers` — PASS (RT-006 already shipped, but the test is added to lock it).
- `go test ./internal/template/ -run TestRenderRetiredWrappers` — FAIL until M2 §Decision Gate resolves.

### M2 — GREEN: EventSetup retirement + settings.json sync

**Priority: High**

Minimal-change retirement of the EventSetup orphan and reconciliation of the 4 retired-wrapper templates per the decision gate.

#### M2.1 — EventSetup retirement (mandatory)

Files to modify:
- `internal/hook/types.go:83-85` — remove the `EventSetup EventType = "Setup"` constant and the surrounding comment block.
- `internal/hook/types.go:139` — remove `EventSetup,` from the active-event slice.
- `internal/hook/types_test.go:36` — remove the `EventSetup: true,` entry from the truth table.
- `internal/cli/hook.go:58` — remove the `{"setup", "Handle setup event", hook.EventSetup},` cobra subcommand binding.
- `internal/cli/hook_e2e_test.go:183, 305` — remove EventSetup references (both the active-set entry and the slug mapping).
- `internal/hook/coverage_table.go` — drop the setup row; decrement `Total` and `Remove` accordingly.
- `internal/cli/doctor_hook_test.go:127-131` — update `summary.Remove` expectation from `1` to `0` AND update `summary.Total` expectation if asserted.
- `internal/cli/doctor_hook_test.go:18-36` — update `TestDoctorHook_27EventTableCount` if the count drops to 26.

#### M2.2 — Wrapper template decision gate

Decision Gate: **Option Keep** is the recommended choice. Rationale: the 4 retired-wrapper templates are harmless (Claude Code will never invoke them because no `settings.json` entry binds them) AND they preserve forward-compat with the planned `system.yaml hook.observability_events` opt-in path (`internal/hook/observability.go:21-44`).

Under Option Keep:
- No `.tmpl` files are deleted.
- `internal/template/render_test.go` — adjust `TestRenderRetiredWrappers` to assert presence (the failing test from M1 flips to assert presence under Option Keep).
- Add inline comment to the 4 retired `.sh.tmpl` files documenting the RETIRE-OBS-ONLY contract: "This wrapper is shipped for forward-compat with `system.yaml hook.observability_events` opt-in. Claude Code will not invoke it unless the corresponding `settings.json` entry is added by the user."

Under Option Retire (deferred — not pursued in this plan):
- Remove the 4 `.tmpl` files.
- Add a migration step (M3) that removes user-local `.sh` copies.

Recommendation rationale: lessons #16 (sync-prefix-correction) — silent files are forgivable; loud files (settings.json entries) are not. We err toward preserving silent surface area until user opt-in evidence appears.

#### M2.3 — Frontmatter normalization

- `.moai/specs/SPEC-V3R2-MIG-002/spec.md` — bring frontmatter to the canonical 12-field schema per `.claude/rules/moai/development/spec-frontmatter-schema.md`:
  - Required: `id`, `title`, `version`, `status`, `created`, `updated`, `author`, `priority`, `phase`, `module`, `lifecycle`, `tags`.
  - Tags as comma-separated string (NOT YAML list): `"hooks, cleanup, orphan, drift, mig"`.
  - Optional `depends_on` retained for SPEC-V3R2-EXT-004 / blocked-by relation.
  - Drop legacy keys (`related_gap`, `related_theme`, `breaking`, `bc_id` if empty).

Verification (M2 exit):
- All M1 tests pass.
- `go test ./internal/hook/ ./internal/cli/ ./internal/template/` — PASS.
- `moai spec lint --strict .moai/specs/SPEC-V3R2-MIG-002/` — `✓ No findings`.

### M3 — GREEN: migration step for user-local settings.json

**Priority: Medium**

Provide a one-shot cleanup step for users upgrading from pre-RT-006 v3 builds whose local `settings.json` still carries the 4 retired event entries.

Deliverables:
- `internal/migrate/hook_cleanup.go` (new) — exposes `CleanupUserSettings(projectRoot string) error`:
  - Reads `<projectRoot>/.claude/settings.json`.
  - For each name in `hook.RetiredEventNames` (move the slice from `audit_test.go:46-51` to a non-test export in `internal/hook/`):
    - If the key exists under `hooks`, remove it.
    - Append the original entry to `<projectRoot>/.moai/archive/hooks/v3.0/migration-<YYYY-MM-DD>.json` (Pattern preserved per spec.md REQ-MIG002-019).
  - Writes the cleaned settings.json back if any change occurred (atomic write via temp + rename).
  - No-op when no retired entries are present.
- Wire the new function into the migration framework provided by SPEC-V3R2-EXT-004 (the actual integration point lives in EXT-004's deliverable; MIG-002 contributes `CleanupUserSettings` as the unit-of-work).
- `internal/migrate/hook_cleanup_test.go` — table-driven tests covering:
  - User-settings.json with all 4 retired entries → all removed, archive file created.
  - User-settings.json with 0 retired entries → no-op, no archive file.
  - Mixed retired + active entries → only retired removed, active preserved.
  - Malformed JSON → return wrapped error, do not write.

Verification (M3 exit):
- `go test ./internal/migrate/ -run TestCleanupUserSettings` — PASS.
- Integration smoke: hand-craft a fixture settings.json with `Notification` + `Elicitation` entries, invoke `CleanupUserSettings`, assert removal + archive.

### M4 — REFACTOR + verification gates

**Priority: Medium**

Cross-cutting polish and final gate verification.

Deliverables:
- Promote `retiredEventNames` from `internal/hook/audit_test.go:46-51` to an exported `internal/hook/retired_events.go` so both the audit test and the migration step in M3 share a single source of truth.
- Update `.claude/rules/moai/core/hooks-system.md` (referenced in spec.md §3) to reflect the post-MIG-002 event roster: 22 native settings.json keys + 4 retired-obs-only + 0 orphans.
- Add a `@MX:ANCHOR` on `InitDependencies` (already present at `internal/cli/deps.go:58-62`) — verify the `fan_in` annotation reflects the post-cleanup state.
- Run the full audit suite + doctor_hook suite end-to-end:
  - `go test -race -count=1 ./internal/hook/ ./internal/cli/ ./internal/template/ ./internal/migrate/`
  - `golangci-lint run ./internal/hook/... ./internal/cli/... ./internal/migrate/...`
  - `go vet ./...`
- Confirm the audit_test.go drift detector continues to fire on artificially injected drift (negative test: temporarily add a fake `EventFoo` constant to types.go in a separate branch; `TestAuditThreeWaySync` MUST fail).

Verification (M4 exit, equivalent to MIG-002 "DONE"):
- All M1–M3 tests pass.
- Coverage on `internal/hook/` ≥ 90% (per CLAUDE.local.md §6 critical-package tier).
- `moai spec lint --strict .moai/specs/SPEC-V3R2-MIG-002/` → `✓ No findings`.
- `audit_test.go` `TestAuditRegistrationParity` continues to hold `nativeCount == 22` and Go handler count == 25 (was 26 pre-MIG-002; setupHandler row removed via M2.1).
- `doctor_hook_test.go` `TestDoctorHook_SummaryCountsConsistent` updated to `summary.Remove == 0` (or kept at 1 if `coverage_table.go` retains the setup row as a historical record — decision logged in M2.1 commit).

## 4. Technical Approach

### 4.1 Detection strategy for `TestAuditThreeWaySync`

Three input sources:

1. **Go event set**: Bootstrap by calling `cli.InitDependencies()` in the test fixture (or by reading `deps.go` source statically). The runtime path is safer because it captures registration intent exactly as the binary sees it. Iterate the `Handlers(event)` map across all `EventType` constants; an event is "Go-registered" if `len(Handlers(e)) > 0`.
2. **settings.json key set**: Render `internal/template/templates/.claude/settings.json.tmpl` against a representative `TemplateContext` (GoBinPath = "/tmp/test/go/bin", HomeDir = "/tmp/test/home"), parse the JSON, extract `Object.keys(result.hooks)`.
3. **Retired-event set**: Import from `internal/hook/retired_events.go` (the M4 promotion).

Assertion:
- `goEvents ∖ (settingsKeys ∪ retiredEvents)` MUST be empty (no orphan Go handlers).
- `settingsKeys ∖ goEvents` MUST be empty (no orphan settings entries).
- Reports per-event diagnostic on failure with `HOOK_SYNC_DRIFT` / `HOOK_WRAPPER_ORPHAN` prefixes (matches spec.md REQ-MIG002-009/010 error codes).

### 4.2 Why we keep retired-wrapper templates

Retiring the 4 `.sh.tmpl` files now is reversible but pre-mature:

- Forward-compat path: `internal/hook/observability.go:23-44` already supports re-enabling these handlers via `system.yaml hook.observability_events`. Removing the wrappers severs the path.
- Disk footprint: 4 × ~1.3 KB = ~5 KB total; negligible.
- Risk surface: With no settings.json binding, Claude Code never invokes them. The Go handlers themselves return silently when not opted in (`audit_test.go:237-246` `TestAuditObservabilityWhitelist` subcase `notification_handler_silent_when_not_opted_in`).

The decision is documented in M2.2 and captured by `TestRenderRetiredWrappers` asserting presence.

### 4.3 Why `CleanupUserSettings` ships in `internal/migrate/`

The migration framework lives in EXT-004 (per spec.md §9.1). MIG-002 contributes a single unit-of-work function. Co-locating with `internal/migrate/` avoids tangling user-data cleanup logic into `internal/hook/` (which should remain a pure event-handler package). Test fixtures under `internal/migrate/testdata/` keep handler tests fast.

## 5. Risks & Mitigations

| Risk | Impact | Mitigation |
|---|---|---|
| Removing `EventSetup` breaks `hook_e2e_test.go` flows that exercise the setup subcommand | Test-suite red after M2 | Update `hook_e2e_test.go:183, 305` and `types_test.go:36` in the same commit as the constant removal; verify `go test ./internal/cli/...` before commit. |
| `TestAuditThreeWaySync` becomes flaky if settings.json.tmpl parser is sensitive to template directives | Intermittent CI failure | Render via the production `template.Render*` path (not a hand-rolled regex); commit a golden-file fixture in `testdata/`. |
| Users on pre-RT-006 v3 builds run `CleanupUserSettings` and silently lose customizations to their `Notification` hook | User data loss | Archive every removed entry to `.moai/archive/hooks/v3.0/migration-<YYYY-MM-DD>.json` (already in deliverable); document the archive path in CLAUDE.local.md §2 "Protected Directories". |
| The wrapper-keep decision is wrong (users expect retired wrappers to be gone) | Conceptual debt | Captured as M2.2 decision gate with rationale; revisit in a follow-up SPEC if user feedback signals otherwise. |
| `coverage_table.go` decrements break `TestDoctorHook_27EventTableCount` assertion | Test-suite red after M2 | Update the doctor test in lockstep with `coverage_table.go` edits; bump expectation to 26. |
| Frontmatter normalization invalidates other tooling that reads spec.md frontmatter | Indirect tooling break | Run `moai spec lint --strict` post-change; verify no downstream consumers parse the legacy keys (`related_gap`, `related_theme`). |

## 6. Out of Scope (residual non-goals beyond spec.md §2.2)

In addition to spec.md §2.2 exclusions, this plan explicitly does NOT cover:

- Re-implementing any of the 5 UPGRADE-resolved handlers (config_change, instructions_loaded, file_changed, post_tool_failure, and any other RT-006 work product).
- Re-implementing the `subagent_stop` FIX (RT-006 P-H02 work product — locked by characterization test only).
- Adding new hook events beyond the 22 native + 4 retired = 26 (now 25 after EventSetup retirement) roster.
- Modifying the hook protocol (JSON-OR-ExitCode contract per `internal/hook/protocol.go`).
- Modifying `hookSpecificOutput` schema or block-decision short-circuit logic in `internal/hook/registry.go:127-202`.
- Removing the 4 retired-event `.sh.tmpl` wrappers (Option Retire — deferred per M2.2 decision).
- Changes to `system.yaml hook.observability_events` schema or `observabilityOptIn()` semantics.
- Changes to harness-observer wrappers (`handle-harness-observe-*.sh`) added by SPEC-V3R4-HARNESS-002.
- Reordering or merging existing handler chains under shared event keys (e.g., the `Stop` event's two-wrapper composition with `handle-harness-observe-stop`).

## 7. Verification Matrix (input to acceptance.md)

| AC | Verification command | Expected outcome |
|---|---|---|
| AC-MIG002-A1 | `go test ./internal/hook/ -run TestAuditThreeWaySync -v` | PASS; zero drift report |
| AC-MIG002-A2 | `go test ./internal/hook/ -run TestAuditNoEventSetupOrphan -v` | PASS; `EventSetup` absent |
| AC-MIG002-A3 | `grep -rn 'EventSetup\|"setup"' internal/hook/ internal/cli/hook.go internal/cli/hook_e2e_test.go` | Empty output |
| AC-MIG002-A4 | `go test ./internal/hook/ -run TestAuditPerFileCategoryHeader -v` | PASS; all 26 handler files headed correctly |
| AC-MIG002-A5 | `go test ./internal/hook/ -run TestAuditRetiredEventsNotInSettings -v` | PASS |
| AC-MIG002-A6 | `go test ./internal/hook/ -run TestAuditRegistrationParity -v` | PASS; nativeCount == 22, Go handler total adjusts to 25 |
| AC-MIG002-A7 | `go test ./internal/migrate/ -run TestCleanupUserSettings -v` | PASS; archive file written; user-active entries preserved |
| AC-MIG002-A8 | `go test ./internal/hook/ -run TestAuditNoStubHandlers -v` | PASS; all 5 RT-006-touched handlers show UPGRADE or FIX header |
| AC-MIG002-A9 | `go test ./internal/cli/ -run TestDoctorHook -v` | PASS; counts consistent with post-cleanup state |
| AC-MIG002-A10 | `moai spec lint --strict .moai/specs/SPEC-V3R2-MIG-002/` | `✓ No findings` |
| AC-MIG002-A11 | `go test -cover ./internal/hook/ ./internal/migrate/` | Coverage ≥ 90% on hook; ≥ 85% on migrate |
| AC-MIG002-A12 | `golangci-lint run ./internal/hook/... ./internal/cli/... ./internal/migrate/...` | Zero findings |

## 8. Dependencies

- **Blocks**: nothing internally; downstream `SPEC-V3R2-MIG-001` consumes `CleanupUserSettings`.
- **Blocked by**: `SPEC-V3R2-EXT-004` (migration framework path) — Already shipped per existing `internal/migrate/` package presence; verify integration surface during M3.
- **Side-effect-coupled with**: `SPEC-V3R2-RT-006` (already merged); MIG-002 ships the *guardrails* against RT-006 regression.

## 9. Lessons Applied

- **Lesson #16 (sync-prefix-correction)**: silent files are forgivable; loud files are not. Keep wrapper templates; remove only the loud `EventSetup` binding.
- **Lesson #17 (verify-only T-deliverable)**: M1's "RED tests" include `TestAuditNoStubHandlers` even though it passes immediately — the goal is to lock RT-006 work product, not re-implement it.
- **Lesson #18 (plan-auditor race prevention)**: this plan does NOT add new HARD constitutional rules; it codifies existing invariants as test surface.
- **Lessons #12/13 (worktree isolation)**: per user policy 2026-05-17, plan-phase runs in main checkout (`plan/SPEC-V3R2-MIG-002` branch), not in an L2 worktree.

End of plan.
