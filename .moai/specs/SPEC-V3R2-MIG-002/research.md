# Research — SPEC-V3R2-MIG-002

> **Goal**: Document the *current* (2026-05-18) state of MoAI hook registration so plan.md can scope an honest residual cleanup. The original spec.md (2026-04-23) was authored BEFORE SPEC-V3R2-RT-006 shipped the per-handler resolution program; many of its premises are now obsolete and must be reconciled.

## 1. Architectural anchors (current registry)

- `internal/hook/registry.go:13-26` — `registry` struct is the central event-router (cfg + handlers map + timeout + traceWriter). `@MX:ANCHOR fan_in=20+`.
- `internal/hook/registry.go:48-57` — `Register(handler)` is the single entry-point for handler installation; appends to per-event slice; no de-duplication.
- `internal/hook/registry.go:67-173` — `Dispatch()` runs handlers sequentially under a per-event timeout (default 30s); short-circuits on `Decision == block` or `ExitCode == 2`; merges `SystemMessage` and `HookSpecificOutput`.
- `internal/hook/registry.go:207-232` — `defaultOutputForEvent()` enumerates the canonical event set (Stop, SessionEnd, SessionStart, PreCompact, SubagentStop, PostToolUseFailure, Notification, SubagentStart, UserPromptSubmit, TeammateIdle, TaskCompleted, WorktreeCreate, WorktreeRemove, PermissionRequest, PreToolUse, PostToolUse).
- `internal/hook/types.go:20-115` — `EventType` constants (27 events) — domain enum.
- `internal/hook/types.go:85` — `EventSetup` constant still declared (no handler, no shell wrapper, no settings.json entry).
- `internal/hook/types.go:139` — `EventSetup` listed in the canonical active-event slice (used by `hook list`).

## 2. Composition root (Go-side registrations)

- `internal/cli/deps.go:42-52` — `Dependencies` struct holds `HookRegistry hook.Registry` plus other domain services.
- `internal/cli/deps.go:64-186` — `InitDependencies()` is the composition root.
- `internal/cli/deps.go:115-117` — `hook.NewRegistry(deps.Config)` instantiates the registry.
- `internal/cli/deps.go:152-185` — 27 explicit `deps.HookRegistry.Register(...)` calls. The active Go-side roster:
  1. `NewSessionStartHandler(deps.Config)` — line 152
  2. `buildSessionEndHandler(cwd)` — line 154
  3. `NewAutoUpdateHandler(buildAutoUpdateFunc())` — line 157 (composite, shares SessionStart event key per audit_test.go:99-101)
  4. `NewStopHandler()` — line 159
  5. `NewPreToolHandlerWithScanner(...)` — line 163
  6. `NewPostToolHandlerWithMxValidatorAndTimeout(...)` — line 164
  7. `NewCompactHandler()` — line 165
  8. `NewPostToolUseFailureHandler()` — line 166
  9. `NewNotificationHandler()` — line 167 (RETIRE-OBS-ONLY)
  10. `NewSubagentStartHandler()` — line 168
  11. `NewUserPromptSubmitHandler(deps.Config)` — line 169
  12. `NewPermissionRequestHandler()` — line 170
  13. `NewTeammateIdleHandler()` — line 171
  14. `NewTaskCompletedHandler()` — line 172
  15. `NewWorktreeCreateHandler()` — line 173
  16. `NewWorktreeRemoveHandler()` — line 174
  17. `NewPostCompactHandler()` — line 175
  18. `NewInstructionsLoadedHandler()` — line 176
  19. `NewStopFailureHandler()` — line 177
  20. `NewSubagentStopHandler()` — line 178
  21. `NewTaskCreatedHandler()` — line 179 (RETIRE-OBS-ONLY)
  22. `NewPermissionDeniedHandler()` — line 180
  23. `NewConfigChangeHandler()` — line 181
  24. `NewCwdChangedHandler()` — line 182
  25. `NewFileChangedHandler()` — line 183
  26. `NewElicitationHandler()` — line 184 (RETIRE-OBS-ONLY)
  27. `NewElicitationResultHandler()` — line 185 (RETIRE-OBS-ONLY)

Total: **26 distinct Go handlers** (`NewAutoUpdateHandler` shares the SessionStart key with `NewSessionStartHandler` per the `audit_test.go:89-101` formula).

## 3. settings.json.tmpl native event keys

`internal/template/templates/.claude/settings.json.tmpl` lines 4-369 enumerate the native hook event keys. The exhaustive list:

- `SessionStart` — line 4
- `PreCompact` — line 20
- `SessionEnd` — line 36
- `PreToolUse` — line 51
- `PostToolUse` — line 67
- `Stop` — line 84 (2 wrappers: handle-stop + handle-harness-observe-stop)
- `SubagentStop` — line 108 (2 wrappers: handle-subagent-stop + handle-harness-observe-subagent-stop)
- `PostToolUseFailure` — line 132
- `SubagentStart` — line 147
- `UserPromptSubmit` — line 162 (2 wrappers: handle-user-prompt-submit + handle-harness-observe-user-prompt-submit)
- `TeammateIdle` — line 186
- `TaskCompleted` — line 201
- `WorktreeCreate` — line 216
- `WorktreeRemove` — line 231
- `ConfigChange` — line 246
- `StopFailure` — line 262
- `PostCompact` — line 278
- `InstructionsLoaded` — line 294
- `CwdChanged` — line 309
- `FileChanged` — line 324
- `PermissionDenied` — line 340
- `PermissionRequest` — line 355

Total: **22 native settings.json event keys** (matches `audit_test.go:93` `const expectedNative = 22`).

Conspicuously **absent** from settings.json.tmpl (intentional RETIRE-OBS-ONLY per `audit_test.go:44-51`):
- `Notification`
- `Elicitation`
- `ElicitationResult`
- `TaskCreated`

Also **absent** (true orphan, never had a settings.json entry):
- `Setup`

## 4. Shell wrappers (local + template)

Local installed wrappers — `.claude/hooks/moai/`:

- 32 `handle-*.sh` files present (verified by `ls .claude/hooks/moai/`).
- Includes: `handle-notification.sh`, `handle-elicitation.sh`, `handle-elicitation-result.sh`, `handle-task-created.sh` (4 wrappers whose event keys are absent from settings.json — they ship via template even though they will never be invoked by Claude Code).
- Does NOT include any `handle-setup.sh` wrapper.

Template-source wrappers — `internal/template/templates/.claude/hooks/moai/`:

- 32 `handle-*.sh.tmpl` files (verified by `ls`).
- Mirrors the local-installed set 1:1.
- Critically the 4 retired-event wrappers are still shipped in the template: `handle-notification.sh.tmpl`, `handle-elicitation.sh.tmpl`, `handle-elicitation-result.sh.tmpl`, `handle-task-created.sh.tmpl`.

## 5. Per-handler current state — overview

Each handler file declares `// Resolution: <CATEGORY>` at line 1 per `audit_test.go:110-170` (`TestAuditPerFileCategoryHeader`). The valid categories are: `KEEP | UPGRADE | FIX | REMOVE | RETIRE-OBS-ONLY | COMPOSITE`.

### 5.1 RETIRE-OBS-ONLY handlers (4)

- `internal/hook/notification.go:1` — `// Resolution: RETIRE-OBS-ONLY` header.
- `internal/hook/notification.go:36-48` — `Handle()` returns silently when `observabilityOptIn(h.cfg, "notification")` is false (Pattern A).
- `internal/hook/elicitation.go:1` — RETIRE-OBS-ONLY header (file contains BOTH `elicitationHandler` and `elicitationResultHandler`).
- `internal/hook/elicitation.go:36-48` — `Handle()` returns silently when opt-in is false.
- `internal/hook/elicitation.go:73-85` — `elicitationResultHandler.Handle()` mirror logic.
- `internal/hook/task_created.go:1` — RETIRE-OBS-ONLY header.
- `internal/hook/task_created.go:36-48` — silent return when not opted in.

### 5.2 UPGRADE handlers (5) — formerly stub, now real logic

- `internal/hook/config_change.go:1` — `// Resolution: UPGRADE — re-render + diff-aware reload via SPEC-V3R2-RT-005 + 20ms debounce.`
- `internal/hook/config_change.go:44-91` — `Handle()` performs 20ms debounce + YAML validation + `ConfigManager.Reload()` via dependency injection.
- `internal/hook/instructions_loaded.go:1` — `// Resolution: UPGRADE — CLAUDE.md 40,000-char budget check per coding-standards.md.`
- `internal/hook/instructions_loaded.go:29-56` — checks both `input.InstructionFilePath` and a fallback `<cwd>/CLAUDE.md` against the 40,000 char budget.
- `internal/hook/file_changed.go:1` — `// Resolution: UPGRADE — MX re-scan for 16 supported language extensions on FileChanged.`
- `internal/hook/file_changed.go:14-37` — `supportedExtensions` map covers 19 extensions across 16 languages (per `language.yaml` neutrality policy in CLAUDE.local.md §15).
- `internal/hook/file_changed.go:55-114` — `Handle()` scans the file with `mx.NewScanner()` and updates the sidecar via `mx.NewManager(stateDir).UpdateFile(...)`.
- `internal/hook/post_tool_failure.go:1` — `// Resolution: UPGRADE — error classification (...).`
- `internal/hook/post_tool_failure.go:13-34` — 7 `ErrorCategory` constants (`TimeoutError`, `PermissionDenied`, `ContextCancelled`, `SandboxViolation`, `OOMKilled`, `ExitError`, `UnknownFailure`).
- `internal/hook/post_tool_failure.go:52-71` — `Handle()` runs `classifyError()` and emits a user-friendly `SystemMessage`.

### 5.3 FIX handlers (1) — bug repair

- `internal/hook/subagent_stop.go:1` — `// Resolution: FIX — P-H02 bug fix: read tmuxPaneId, kill-pane with 500ms timeout, update team registry.`
- `internal/hook/subagent_stop.go:32-98` — `Handle()` reads `~/.claude/teams/<team-name>/config.json`, locates the teammate's `tmuxPaneId`, runs `tmux kill-pane -t <id>` in a goroutine bounded by a 500ms timeout (SPEC-V3R2-RT-006 AC-15, AC-02), then strips the teammate from the team config.
- `internal/hook/subagent_stop.go:42-45` — Windows short-circuit (`runtime.GOOS == "windows"` → no-op).
- `internal/hook/subagent_stop.go:76-86` — "pane not found" graceful-degradation path.

### 5.4 KEEP handlers — observability gate helper

- `internal/hook/observability.go:1` — `// Resolution: KEEP — observability gate helper for RETIRE-OBS-ONLY handlers.`
- `internal/hook/observability.go:23-44` — `observabilityOptIn(cfg, eventName)` reads `cfg.Get().System.Hook.ObservabilityEvents` (declared at `internal/config/types.go:49-52` as `ObservabilityEvents []string \`yaml:"observability_events"\``).
- `internal/hook/observability.go:21-22` — `@MX:ANCHOR fan_in=4` (notification, elicitation, elicitationResult, taskCreated).

### 5.5 COMPOSITE handler — autoUpdate

- `internal/hook/auto_update.go:1` — `// Resolution: COMPOSITE` (registered under SessionStart event key, not a separate event).

## 6. Audit test suite (the de-facto invariant set)

- `internal/hook/audit_test.go:46-51` — `retiredEventNames = []string{"Notification", "Elicitation", "ElicitationResult", "TaskCreated"}` (canonical list, used by 3 test cases).
- `internal/hook/audit_test.go:59-103` — `TestAuditRegistrationParity` asserts:
  - `nativeCount == 22` (settings.json hook keys).
  - The 4 retired names MUST NOT appear in `settingsJSON.Hooks`.
  - Expected total Go handlers = `22 + 4 = 26`.
- `internal/hook/audit_test.go:110-170` — `TestAuditPerFileCategoryHeader` requires every handler file to carry the `// Resolution: CATEGORY` header in its first ~10 lines.
- `internal/hook/audit_test.go:177-196` — `TestAuditRetiredEventsNotInSettings` independently asserts retired-event absence from local `settings.json`.
- `internal/hook/audit_test.go:203-255` — `TestAuditObservabilityWhitelist` covers 5 subcases: empty/listed/case-insensitive/nil-cfg/strict-mode + a `notification_handler_silent_when_not_opted_in` integration assertion.
- `internal/hook/audit_test.go:259-299` — `TestAuditRetiredHandlersNotActive` cross-checks `EventType` constants vs. an enumerated active list.

## 7. Doctor coverage table (separate invariant surface)

- `internal/cli/doctor_hook_test.go:18-36` — `TestDoctorHook_27EventTableCount` asserts the coverage table has exactly 27 entries (matches `hook.CoverageTable`).
- `internal/cli/doctor_hook_test.go:113-135` — `TestDoctorHook_SummaryCountsConsistent` asserts the per-resolution counts:
  - `RetireObsOnly == 4`
  - `Fix == 1` (subagentStop P-H02)
  - `Remove == 1` (setupHandler)
  - `Composite == 1` (autoUpdate)
- `internal/hook/coverage_table.go` (sized ~5.8 KB per `ls`) — source of `hook.CoverageTable` and `hook.Summarize()`.

## 8. EventSetup orphan footprint

- `internal/hook/types.go:83-85` — `EventSetup` constant retained (declared "triggered via --init, --init-only, or --maintenance CLI flags").
- `internal/hook/types.go:139` — `EventSetup` still in active event slice.
- `internal/hook/types_test.go:36` — `EventSetup: true` in the active-event truth table.
- `internal/cli/hook.go:58` — `{"setup", "Handle setup event", hook.EventSetup}` cobra binding still wires a `moai hook setup` CLI subcommand.
- `internal/cli/hook_e2e_test.go:183, 305` — references the setup event in e2e tests.
- `internal/cli/doctor_hook_test.go:130` — explicitly expects `summary.Remove == 1 (setupHandler)`.
- No `internal/hook/setup.go` exists (verified by `find . -name "setup*.go" -print` returning empty).
- No `NewSetupHandler()` constructor exists (verified by `grep -rn "NewSetupHandler" internal/`).
- No `handle-setup.sh` shell wrapper exists in `.claude/hooks/moai/` or `internal/template/templates/.claude/hooks/moai/`.
- No `Setup` key in `internal/template/templates/.claude/settings.json.tmpl`.

This is the **only true orphan** in the codebase post-RT-006: an `EventType` constant + CLI binding without a handler implementation.

## 9. Original spec.md premises vs. reality (drift table)

The `.moai/specs/SPEC-V3R2-MIG-002/spec.md` document was authored 2026-04-23 — before SPEC-V3R2-RT-006 landed (RT-006 references appear in `subagent_stop.go:64`, `audit_test.go:3`, `notification.go:3`, etc.). The drift between spec.md assumptions and the current tree:

| spec.md claim | Reality (2026-05-18) | Reconciliation needed |
|---|---|---|
| `setupHandler` is a Go-handler orphan (R6 §2.2) | `setup.go` never existed; only `EventSetup` constant + CLI binding | Cleanup target shifts from "delete setupHandler" to "remove EventSetup constant + CLI subcommand binding + type_test entry" |
| 14/27 handlers are "thin loggers" | All 4 retired handlers are observability taps; 5 former stubs are UPGRADE-resolved; subagent_stop is FIX-resolved | "Thin logger" framing is obsolete |
| `subagentStopHandler` lacks tmux pane kill (REQ-MIG002-004, REQ-MIG002-013) | Already implemented at `subagent_stop.go:32-98` with 500ms timeout + graceful degradation | REQ duplicates RT-006 work; lift to "characterization test confirming the behavior holds" |
| `configChangeHandler` lacks reload trigger (REQ-MIG002-005) | Implemented at `config_change.go:44-91` with diff-aware reload | Same — characterization test only |
| `instructionsLoadedHandler` lacks CLAUDE.md budget check (REQ-MIG002-006, REQ-MIG002-014) | Implemented at `instructions_loaded.go:29-78` with 40,000 char limit | Same — characterization test only |
| `fileChangedHandler` lacks MX rescan (REQ-MIG002-007, REQ-MIG002-016) | Implemented at `file_changed.go:55-114` with 19 supported extensions | Same — characterization test only |
| `postToolUseFailureHandler` lacks error classification (REQ-MIG002-008) | Implemented at `post_tool_failure.go:13-135` with 7 error categories | Same — characterization test only |
| Retire `setupHandler, notificationHandler, elicitationHandler, elicitationResultHandler, taskCreatedHandler` (REQ-MIG002-003) | 4 already retired-obs-only; `setupHandler` never existed | Only the EventSetup *constant/binding* removal remains |
| Migration should remove user-local `handle-notification.sh` (REQ-MIG002-011, REQ-MIG002-019) | Wrappers still ship in template (`handle-notification.sh.tmpl` et al.); user-local copies follow | Decide: keep wrappers (they're harmless no-ops without a settings.json entry) OR remove templates + add `moai migrate hook-cleanup` step |
| Migration should remove retired entries from user-local settings.json (REQ-MIG002-012) | Local settings.json already excludes the 4 retired keys (verified by `audit_test.go:177-196`) | Migration-step needed only for users upgrading from v3.0.0-pre-RT-006 |
| `HOOK_SYNC_DRIFT` / `HOOK_WRAPPER_ORPHAN` CI failure modes (REQ-MIG002-009, REQ-MIG002-010) | Not implemented; `audit_test.go` only asserts hard-coded counts | Real residual work — introduce a 3-way-sync assertion that scans (Go handlers ∩ shell wrappers ∩ settings.json keys) and reports orphans |

## 10. Hook protocol contract (boundary)

- `internal/hook/protocol.go` — JSON-OR-ExitCode dispatch contract (`Protocol` interface).
- `internal/hook/registry.go:127-148` — block-decision short-circuit + exit-code 2 propagation.
- `internal/hook/registry.go:179-202` — `isBlockDecision()` + `getBlockReason()` cover both top-level decisions (Stop, PostToolUse) and `hookSpecificOutput.permissionDecision` (PreToolUse).
- `internal/hook/types.go` — `HookInput` / `HookOutput` / `HookSpecificOutput` struct definitions (boundary types between Claude Code stdin and Go handlers).

## 11. Coverage / quality posture

- Per CLAUDE.local.md §6 the hook package targets 90%+ coverage (critical package tier).
- Existing tests for the modules in scope:
  - `notification_test.go` (1.3 KB) — verifies silent path.
  - `subagent_stop` covered by `coverage_boost_test.go`, `misc_coverage_test.go`, and `new_handlers_coverage_test.go`.
  - `config_change_test.go` (6.7 KB) — verifies debounce + reload trigger.
  - `file_changed_test.go` (3.8 KB) — verifies extension filter + scanner invocation.
  - `instructions_loaded_test.go` (3.8 KB) — verifies 40,000 char budget path.
  - `audit_test.go` (9.6 KB) — registration parity + per-file header + retire-event absence + observability whitelist.
- No file at `internal/template/settings_audit_test.go` (spec.md §3 referenced this name; actual audit-suite is `internal/hook/audit_test.go`).

## 12. Dependencies on this SPEC

- `SPEC-V3R2-MIG-001` blocks on this SPEC's "cleanup step" being callable from the v2→v3 migrator (spec.md §9.2).
- This SPEC blocks on `SPEC-V3R2-EXT-004` providing the migration framework path (spec.md §9.1).
- `SPEC-V3R2-RT-006` shipped the per-handler resolution program that supersedes spec.md REQ-MIG002-004 through REQ-MIG002-008 (and most of REQ-MIG002-003).

## 13. Memory anchors (out-of-tree)

- `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/feedback_team_tmux_cleanup.md` — original tmux-pane-cleanup gap that motivated subagent_stop FIX (now resolved at `subagent_stop.go:32-98`).
- `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/MEMORY.md` (index) — references the RT-006 lifecycle entry; no entry for MIG-002 yet (this plan-phase PR will populate one).

## 14. Residual work surface (input to plan.md)

After reconciliation the residual MIG-002 surface is:

1. **EventSetup constant removal**: drop `EventSetup` from `internal/hook/types.go:83-85, 139`; remove the `internal/cli/hook.go:58` cobra subcommand binding; update `internal/hook/types_test.go:36` and `internal/cli/hook_e2e_test.go:183, 305`; update `hook.CoverageTable` so `summary.Remove` becomes 0 (or update `doctor_hook_test.go:130-131` to match).
2. **3-way sync invariant test**: extend `internal/hook/audit_test.go` with `TestAuditThreeWaySync` that asserts the Go-registered event set (from `deps.go` invocations) ≡ the settings.json.tmpl hook-key set ∪ `retiredEventNames`. Surfaces both `HOOK_SYNC_DRIFT` (Go-only) and `HOOK_WRAPPER_ORPHAN` (settings-only) drift.
3. **Wrapper-ship decision**: confirm intent for the 4 retired-event template wrappers (`handle-notification.sh.tmpl`, `handle-elicitation.sh.tmpl`, `handle-elicitation-result.sh.tmpl`, `handle-task-created.sh.tmpl`). Two options:
   - **Keep** — wrappers exist for hypothetical observability opt-in users; Claude Code won't invoke them because no settings.json entry binds them.
   - **Retire** — remove the 4 tmpls + add a `moai migrate hook-cleanup` step that deletes user-local copies.
4. **Migration step for legacy local settings.json**: detect user-local `settings.json` entries for the 4 retired events; archive/move on `moai migrate v2-to-v3` invocation (REQ-MIG002-012).
5. **Frontmatter reconciliation**: spec.md frontmatter must be brought to the canonical 12-field schema (`.claude/rules/moai/development/spec-frontmatter-schema.md`). Current spec.md frontmatter uses 14 fields including legacy keys (`created`, `updated`, `dependencies`, `related_gap`, `related_theme`, `breaking`, `bc_id`) that need normalization to the 12-field schema before plan PR lints clean.

The plan.md milestones key off these 5 residual items, not the original spec.md "stub upgrade" framing.

## 15. Files modified by this SPEC (forecast)

Forecast — to be confirmed during run-phase:

- `internal/hook/types.go` — remove EventSetup constant + truth-table entry.
- `internal/hook/types_test.go` — remove EventSetup truth-table assertion.
- `internal/cli/hook.go` — remove `setup` cobra subcommand binding.
- `internal/cli/hook_e2e_test.go` — remove EventSetup references.
- `internal/cli/doctor_hook_test.go` — update `summary.Remove` expectation from 1 to 0.
- `internal/hook/coverage_table.go` — drop the setup row, decrement counts.
- `internal/hook/audit_test.go` — add `TestAuditThreeWaySync`.
- `.moai/specs/SPEC-V3R2-MIG-002/spec.md` — frontmatter normalization (12-field schema).
- Optional (pending decision in plan-phase): 4 wrapper `.tmpl` deletions + migrate step.

End of research.
