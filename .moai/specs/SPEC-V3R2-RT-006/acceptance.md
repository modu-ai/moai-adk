# SPEC-V3R2-RT-006 Acceptance Criteria

> Detailed acceptance criteria for **Hook Handler Completeness and 27-Event Coverage**.
> Companion to `spec.md` § 6, `plan.md`, `tasks.md`, `research.md`.
> Each AC is independently verifiable; spec §6 enumerated 17 ACs (AC-V3R2-RT-006-01..17). This file expands each AC into Given-When-Then scenarios with verification commands and edge-case enumeration.

## HISTORY

| Version | Date       | Author                        | Description                                                                                  |
|---------|------------|-------------------------------|----------------------------------------------------------------------------------------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow)  | Initial acceptance criteria. 17 baseline ACs + 4 derived edge-case ACs (AC-18..21).         |

---

## 1. Acceptance Criteria Format

Each AC is structured as:
- **Given**: precondition + environment state
- **When**: triggering action or input
- **Then**: observable evidence (command output, file content, test result)
- **Maps to**: REQ ID(s) from spec.md § 5
- **Verification command**: explicit shell or `go test` invocation
- **Edge cases**: failure modes considered

[HARD] AC verification is performed during run-phase M5-T28..T35 (per `tasks.md`); plan-auditor uses this file to confirm REQ coverage breadth.

---

## 2. Acceptance Criteria

### AC-V3R2-RT-006-01: SubagentStop tmux pane teardown end-to-end

- **Given**: A team mode session is active with 3 teammates whose `~/.claude/teams/<team-name>/config.json` contains valid `tmuxPaneId` entries (e.g., `%12`, `%13`, `%14`).
- **When**: A SubagentStop event fires for any teammate (e.g., `teammateName: "researcher"`).
- **Then**:
  - `tmux kill-pane -t <correct-id>` is invoked.
  - The teammate entry is removed from `~/.claude/teams/<team-name>/config.json`.
  - HookOutput.SystemMessage = `"Teammate researcher shut down, pane %12 released"`.
- **Maps to**: REQ-V3R2-RT-006-006, REQ-V3R2-RT-006-010
- **Verification command**: `go test -run TestSubagentStop_FullPipeline ./internal/hook/` + manual integration test on live tmux session (M5-T31).
- **Edge cases**:
  - Teammate already removed by concurrent SubagentStop call → AC-02 covers.
  - Pane already gone (tmux returns "pane not found") → AC-02 covers.
  - Windows host (no tmux) → AC-03 covers.
  - kill-pane exceeds 500 ms timeout → handler still updates registry (M3-T14).

### AC-V3R2-RT-006-02: SubagentStop pane-not-found graceful handling

- **Given**: Team config indicates teammate `researcher` has `tmuxPaneId: %99`, but tmux returns "can't find pane" because pane was killed externally.
- **When**: SubagentStop fires for `researcher`.
- **Then**:
  - Handler detects "pane not found" / "can't find pane" in stderr.
  - Treats cleanup as successful (no error returned).
  - Registry entry is still removed.
  - slog.Debug emits "tmux pane not found (may already be removed)".
- **Maps to**: REQ-V3R2-RT-006-061
- **Verification command**: `go test -run TestSubagentStop_PaneNotFoundGraceful ./internal/hook/`
- **Edge cases**:
  - tmux returns "no server running" (server already exited) → also graceful.
  - tmux returns syntax error (impossible in practice but tested defensively) → returns error but registry still updated.

### AC-V3R2-RT-006-03: SubagentStop Windows no-op

- **Given**: `runtime.GOOS == "windows"`.
- **When**: SubagentStop fires.
- **Then**:
  - Handler returns `HookOutput{}` immediately (no tmux invocation).
  - slog.Debug emits "tmux cleanup skipped on Windows".
- **Maps to**: REQ-V3R2-RT-006-007
- **Verification command**: `GOOS=windows go test -run TestSubagentStop_WindowsNoOp ./internal/hook/` (build constraint test).
- **Edge cases**:
  - Cross-compiled binary running on Linux but `GOOS=windows` set → test passes (runtime.GOOS reads compiled value).

### AC-V3R2-RT-006-04: ConfigChange RT-005 reload integration

- **Given**: `.moai/config/sections/quality.yaml` is edited externally to add `coverage_threshold: 90` (valid field).
- **When**: ConfigChange fires with `input.ConfigFilePath: ".moai/config/sections/quality.yaml"`.
- **Then**:
  - Handler re-reads the file (after 20 ms debounce).
  - Calls `Manager.Reload(path)` from RT-005.
  - HookOutput.AdditionalContext = `".moai/config/sections/quality.yaml reloaded successfully"` (or SystemMessage with same text).
  - Continue:true.
- **Maps to**: REQ-V3R2-RT-006-011
- **Verification command**: `go test -run TestConfigChange_RT005ReloadIntegration ./internal/hook/`
- **Edge cases**:
  - File deleted between event fire and ReadFile → handler returns Continue:false + SystemMessage with read error.
  - File mid-write (truncated YAML) → debounce of 20 ms mitigates; if still mid-write, validation catches.
  - Manager.Reload is not yet wired (RT-005 not in context) → handler skips reload, emits AdditionalContext warning.

### AC-V3R2-RT-006-05: ConfigChange invalid YAML keeps old settings

- **Given**: `.moai/config/sections/quality.yaml` is edited to introduce invalid field `coverage_threshold: not-a-number`.
- **When**: ConfigChange fires.
- **Then**:
  - Handler reads file successfully but YAML parse + validator/v10 raises validation error.
  - HookOutput.SystemMessage = `"Config reload failed: ..."` or `"Config reload rejected: ...; old settings retained"`.
  - HookOutput.Continue = false.
  - Manager.Reload is NOT called (or its return error triggers retention semantics).
  - In-memory cfg remains the pre-edit version.
- **Maps to**: REQ-V3R2-RT-006-011, REQ-V3R2-RT-006-062
- **Verification command**: `go test -run TestConfigChange_InvalidYAMLKeepsOldSettings ./internal/hook/`
- **Edge cases**:
  - YAML syntax error (not just type mismatch) → same handling.
  - Field type matches but value out of validator range (e.g., `coverage_threshold: 200`) → also rejected.

### AC-V3R2-RT-006-06: InstructionsLoaded 40,000 char overage warning

- **Given**: A synthetic CLAUDE.md of 42,000 UTF-8 characters exists at `<cwd>/CLAUDE.md`.
- **When**: InstructionsLoaded fires with `input.InstructionFilePath: "<cwd>/CLAUDE.md"`.
- **Then**:
  - Handler reads file, counts UTF-8 characters (utf8.RuneCount).
  - Detects 42000 > 40000.
  - HookOutput.SystemMessage names the file path and the character count: `"<path> exceeds 40,000 char budget at 42000; split content per coding-standards.md"`.
  - Continue:true (warning, not block).
- **Maps to**: REQ-V3R2-RT-006-012
- **Verification command**: `go test -run TestInstructionsLoaded_42kCLAUDE ./internal/hook/`
- **Edge cases**:
  - Multi-byte UTF-8 (Korean text) → character count uses runes, not bytes.
  - File exactly 40,000 chars → no warning (boundary).
  - File 40,001 chars → warning fires.
  - File missing → handler returns `HookOutput{}` silently (existing behavior).

### AC-V3R2-RT-006-07: FileChanged MX tag delta detection

- **Given**: `internal/auth/handler.go` exists with no @MX tags. External edit adds `// @MX:WARN at line 42` comment.
- **When**: FileChanged fires with `input.FilePath: "internal/auth/handler.go"`, `input.ChangeType: "modified"`.
- **Then**:
  - Handler verifies `.go` extension is in supportedExtensions.
  - Invokes TagScanner from SPC-002 (or stub interface during run-phase).
  - Detects new @MX:WARN marker.
  - HookOutput.AdditionalContext = `"MX tag delta on internal/auth/handler.go: +1 WARN"` (or similar summary).
- **Maps to**: REQ-V3R2-RT-006-013
- **Verification command**: `go test -run TestFileChanged_MXTagDelta ./internal/hook/`
- **Edge cases**:
  - Unsupported extension (e.g., `.md`, `.json`) → handler returns `HookOutput{}` without scan.
  - File deleted (ChangeType: "removed") → existing tags removed; AdditionalContext shows negative delta.
  - SPC-002 not merged → mock TagScanner returns deterministic delta in test.

### AC-V3R2-RT-006-08: PostToolUseFailure timeout classification

- **Given**: A Bash tool execution fails with `exit code 124` and stderr containing `"command timed out after 30s"`.
- **When**: PostToolUseFailure fires with `input.ToolName: "Bash"`, `input.Error: "exit status 124"`, `input.Stderr: "command timed out after 30s"`.
- **Then**:
  - Handler classifies as `TimeoutError`.
  - HookOutput.SystemMessage = `"TimeoutError: command timed out after 30s; consider --timeout flag or reduce input scope"` (actionable hint).
  - HookOutput.AdditionalContext contains diagnostic hints.
- **Maps to**: REQ-V3R2-RT-006-014
- **Verification command**: `go test -run TestPostToolFailure_TimeoutClassification ./internal/hook/`
- **Edge cases**:
  - Exit code 137 (OOM) → classified as `OOMKilled`.
  - Exit code 1 with no stderr signature → classified as `ExitError` (default).
  - context.Cancelled in error → classified as `ContextCancelled`.
  - Permission denied stderr → classified as `PermissionDenied`.
  - SandboxViolation signature → classified separately.

### AC-V3R2-RT-006-09: setup.go orphan removal

- **Given**: Repository at `feat/SPEC-V3R2-RT-006` worktree.
- **When**: `go build ./...` runs.
- **Then**:
  - `internal/hook/setup.go` does not exist (`ls internal/hook/setup.go` returns "No such file").
  - `grep -r "NewSetupHandler" internal/` returns no matches.
  - Build succeeds.
- **Maps to**: REQ-V3R2-RT-006-005
- **Verification command**: `[ ! -f internal/hook/setup.go ] && go build ./... && ! grep -rq NewSetupHandler internal/`
- **Edge cases**:
  - Git history retains old setup.go → harmless; only working tree state matters.
  - Test files reference NewSetupHandler → grep would catch; tests must be updated.

### AC-V3R2-RT-006-10: settings.json retire-event audit

- **Given**: `.claude/settings.json` is post-M3-T10 state.
- **When**: TestAuditRetiredEventsNotInSettings runs.
- **Then**:
  - Test parses settings.json hooks map.
  - For each of the 4 retire events (Notification, Elicitation, ElicitationResult, TaskCreated), test asserts the key is absent.
  - If any retire key is present (e.g., due to manual user edit), test FAILS naming the event.
- **Maps to**: REQ-V3R2-RT-006-004, REQ-V3R2-RT-006-063
- **Verification command**: `go test -run TestAuditRetiredEventsNotInSettings ./internal/hook/`
- **Edge cases**:
  - User has manually opted in via system.yaml `hook.observability_events: ["notification"]` → settings.json regenerated at next `moai update` adds Notification key back; test should pass when run inline (read-only check).
  - Test runs after manual rollback edit → fails fast naming the event.

### AC-V3R2-RT-006-11: Observability opt-in structured-log only

- **Given**: `system.yaml` has `hook.observability_events: ["notification"]` AND user has run `moai update` so settings.json contains the Notification hook entry.
- **When**: A Notification event fires.
- **Then**:
  - Handler reads `cfg.System.Hook.ObservabilityEvents` via context.
  - `observabilityOptIn(ctx, "notification")` returns true.
  - Handler emits `slog.Info("notification received", ...)` only.
  - HookOutput.SystemMessage is empty.
  - HookOutput.Continue is unset (zero value = true by Claude Code default).
  - No user-facing side effect (no chat injection, no permission decision, no continue:false).
- **Maps to**: REQ-V3R2-RT-006-040
- **Verification command**: `go test -run TestNotification_ObservabilityOptIn ./internal/hook/`
- **Edge cases**:
  - Opt-in list contains invalid event name (e.g., "foo") → validator/v10 rejects at config load; handler never reached.
  - Opt-in list is empty → AC-16 covers (silent return).
  - Settings.json has Notification entry but system.yaml is empty → handler returns silent (AC-16).

### AC-V3R2-RT-006-12: doctor hook 27-event table output

- **Given**: A clean repository with default system.yaml (no observability opt-in).
- **When**: User runs `moai doctor hook`.
- **Then**:
  - Stdout contains a 27-row table with columns: #, Event, Resolution, Status.
  - Header row labels match.
  - Footer summary: `"17 KEEP, 4 UPGRADE, 1 FIX, 4 RETIRE-OBS-ONLY (0 opted-in), 1 REMOVED"` (or equivalent).
  - Exit code 0.
- **Maps to**: REQ-V3R2-RT-006-050, REQ-V3R2-RT-006-051
- **Verification command**: `go test -run TestDoctorHook_TableOutput ./internal/cli/` + manual `moai doctor hook` invocation.
- **Edge cases**:
  - `moai doctor hook --json` → outputs valid JSON parseable by `jq`.
  - `moai doctor hook --observability` → highlights opt-in events.
  - `moai doctor hook --trace SubagentStop` → tail of `.moai/logs/hook.log` filtered to SubagentStop entries (or no-op message if log absent).

### AC-V3R2-RT-006-13: audit_test detects undocumented handler

- **Given**: A developer adds `internal/hook/foobar.go` with `NewFooBarHandler()` registered in `deps.go`, but does not update spec.md §5.7 table or add a `// Resolution:` header.
- **When**: TestAuditPerFileCategoryHeader + TestAuditRegistrationParity runs in CI.
- **Then**:
  - TestAuditPerFileCategoryHeader FAILS naming `foobar.go` as missing Resolution header.
  - TestAuditRegistrationParity FAILS reporting count mismatch (e.g., 27 deps vs 22 settings + composite + obs-residual).
  - CI build blocks merge.
- **Maps to**: REQ-V3R2-RT-006-003, REQ-V3R2-RT-006-042
- **Verification command**: `go test -run "TestAuditPerFileCategoryHeader|TestAuditRegistrationParity" ./internal/hook/`
- **Edge cases**:
  - Developer adds Resolution header but forgets deps.go register → TestAuditRegistrationParity catches.
  - Developer forgets header but registers → TestAuditPerFileCategoryHeader catches.
  - Both forgotten → both tests fail.

### AC-V3R2-RT-006-14: per-file Resolution header coverage

- **Given**: All non-test, non-aux handler files in `internal/hook/`.
- **When**: `grep -E '^// Resolution: (KEEP|UPGRADE|FIX|REMOVE|RETIRE-OBS-ONLY|COMPOSITE)$' internal/hook/<handler>.go` runs.
- **Then**:
  - All 26 remaining handler files (27 - 1 setup removed) match the pattern.
  - No file has unrecognized category value.
- **Maps to**: REQ-V3R2-RT-006-002
- **Verification command**: `go test -run TestAuditPerFileCategoryHeader ./internal/hook/` (test does the grep internally).
- **Edge cases**:
  - Aux files (`generic_handler.go`, `dual_parse.go`, `glm_tmux.go`, `normalize.go`, `wrapper*.go`, `coverage_table.go`, `observability.go`, `errors.go`, `doc.go`, `contract.go`) are excluded from the check (they don't implement Handler interface or are infrastructure).

### AC-V3R2-RT-006-15: SubagentStop pane teardown ordering preservation

- **Given**: Team protocol per `.claude/rules/moai/workflow/team-protocol.md` § Team Discovery requires `shutdown_request → shutdown_response → SubagentStop` ordering.
- **When**: A teammate sends shutdown_response, then SubagentStop fires.
- **Then**:
  - Handler executes pane teardown (kill-pane + config update) AFTER shutdown_response is processed.
  - No race: orchestrator does not call SubagentStop before shutdown_response is acknowledged.
- **Maps to**: REQ-V3R2-RT-006-010 (sub-clause "preserves the protocol order")
- **Verification command**: Manual integration test (M5-T31). Codified via `subagent_stop_test.go` ordering test if mocking framework supports.
- **Edge cases**:
  - Teammate timeout exceeds shutdown grace → orchestrator fallback to forceful SubagentStop; pane teardown still occurs.

### AC-V3R2-RT-006-16: Empty observability_events silent retire

- **Given**: `system.yaml` has `hook.observability_events: []` (empty).
- **When**: A retire event (e.g., Notification) fires (theoretically — settings.json should not register it, but defensively tested).
- **Then**:
  - Handler returns `HookOutput{}` without logging.
  - No SystemMessage / no AdditionalContext / no Continue:false.
- **Maps to**: REQ-V3R2-RT-006-040
- **Verification command**: `go test -run TestNotification_EmptyObservabilityList ./internal/hook/`
- **Edge cases**:
  - Settings.json mistakenly contains Notification entry but observability_events is empty → handler silent return (AC-16 scenario).

### AC-V3R2-RT-006-17: PreToolUse PermissionDecision integration

- **Given**: A PreToolUse event fires with `input.ToolName: "Bash"`, `input.ToolInput: {"command": "rm -rf /"}`.
- **When**: PreToolUseHandler.Handle runs and consults RT-002 permission stack.
- **Then**:
  - HookOutput.PermissionDecision is populated with `{Decision: "deny", Reason: "..."}` (or similar SPEC-V3R2-RT-002 contract).
  - Optionally HookOutput.UpdatedInput rewrites the command (e.g., to `echo "blocked"`).
- **Maps to**: REQ-V3R2-RT-006-022
- **Verification command**: `go test -run TestPreToolUse_PermissionDecision ./internal/hook/`
- **Edge cases**:
  - Tool not in permission stack → PermissionDecision is nil; orchestrator falls back to default policy.
  - Stack returns "ask" → decision relayed to user.

---

## 3. Derived Edge-Case ACs

These are not in spec §6 but are implicit from REQ + risk analysis. Plan-auditor may treat as advisory.

### AC-V3R2-RT-006-18: Empty observability_events at SystemHookConfig zero value

- **Given**: `system.yaml` has no `hook:` section at all (key missing).
- **When**: RT-005 typed loader builds Config.
- **Then**:
  - `cfg.System.Hook` is `SystemHookConfig{ObservabilityEvents: nil, StrictMode: false}`.
  - `observabilityOptIn(ctx, anyEvent)` returns false for all events.
- **Maps to**: REQ-V3R2-RT-006-040 (zero-value behavior)
- **Verification command**: `go test -run TestSystemHookConfig_ZeroValue ./internal/config/`

### AC-V3R2-RT-006-19: Validator rejects unknown observability event

- **Given**: User edits `system.yaml` with `hook.observability_events: ["invalidEvent"]`.
- **When**: RT-005 typed loader runs.
- **Then**:
  - validator/v10 raises ValidationError naming the field and value.
  - Manager.Load returns error; old config retained.
- **Maps to**: REQ-V3R2-RT-006-040 (whitelist enforcement via validator)
- **Verification command**: `go test -run TestSystemHookConfig_InvalidEventRejected ./internal/config/`

### AC-V3R2-RT-006-20: ConfigChange debounce mitigates fsnotify mid-write

- **Given**: A file write is in progress when fsnotify triggers ConfigChange (truncated content visible).
- **When**: ConfigChange fires.
- **Then**:
  - Handler waits 20 ms before ReadFile.
  - File is fully written by then; YAML parse succeeds.
  - Reload completes normally.
- **Maps to**: REQ-V3R2-RT-006-011 (race mitigation per spec §8 risk row 4)
- **Verification command**: `go test -run TestConfigChange_DebounceMidWrite ./internal/hook/` (timed test using channel signaling).

### AC-V3R2-RT-006-21: SubagentStop kill-pane within 500 ms timeout

- **Given**: A misconfigured tmux server hangs on kill-pane (extremely rare in practice).
- **When**: SubagentStop fires and goroutine wraps killTmuxPane with `context.WithTimeout(500 ms)`.
- **Then**:
  - After 500 ms, ctx.Done() fires.
  - Handler logs slog.Warn("kill-pane timeout") and proceeds with config cleanup.
  - HookOutput.SystemMessage indicates pane release was best-effort.
  - Total handler latency < 1500 ms (SessionEnd budget).
- **Maps to**: REQ-V3R2-RT-006-006 + spec §8 risk row 1 mitigation
- **Verification command**: `go test -run TestSubagentStop_KillPaneTimeout ./internal/hook/` (uses channel-blocked exec mock).

---

## 4. AC ↔ REQ ↔ Task Reverse Map

| AC ID | Maps to REQ(s) | Run-phase task(s) |
|-------|----------------|--------------------|
| AC-01 | REQ-006, -010 | T-RT006-14, T-RT006-31 |
| AC-02 | REQ-061 | T-RT006-03, T-RT006-14 |
| AC-03 | REQ-007 | T-RT006-14 |
| AC-04 | REQ-011 | T-RT006-15 |
| AC-05 | REQ-011, -062 | T-RT006-16 |
| AC-06 | REQ-012 | T-RT006-17 |
| AC-07 | REQ-013 | T-RT006-18 |
| AC-08 | REQ-014 | T-RT006-20 |
| AC-09 | REQ-005 | T-RT006-01 |
| AC-10 | REQ-004, -063 | T-RT006-09, T-RT006-10, T-RT006-21 |
| AC-11 | REQ-040 | T-RT006-04, T-RT006-05, T-RT006-06, T-RT006-07 |
| AC-12 | REQ-050, -051 | T-RT006-23, T-RT006-24, T-RT006-25, T-RT006-26 |
| AC-13 | REQ-003, -042 | T-RT006-19 |
| AC-14 | REQ-002 | T-RT006-12, T-RT006-13 |
| AC-15 | REQ-010 (ordering) | T-RT006-31 (manual integration) |
| AC-16 | REQ-040 (zero-value) | T-RT006-07 |
| AC-17 | REQ-022 | (semantic reaffirmation; covered by existing pretool tests) |
| AC-18 | REQ-040 | T-RT006-05 |
| AC-19 | REQ-040 | T-RT006-05 |
| AC-20 | REQ-011 | T-RT006-16 |
| AC-21 | REQ-006 | T-RT006-14 |

## 5. Definition of Done

[HARD] All 17 baseline ACs verified before run-phase merge:

- [ ] AC-01 through AC-17 all GREEN (`go test ./internal/hook/ ./internal/cli/`)
- [ ] AC-18 through AC-21 derived edge cases verified
- [ ] Manual integration test (M5-T31) recorded in `progress.md`
- [ ] `moai doctor hook` output matches AC-12 expectation
- [ ] `golangci-lint run` clean
- [ ] `make build` regenerates embedded.go correctly
- [ ] CHANGELOG entry per M5-T32
- [ ] @MX tags applied per `plan.md` § 6
- [ ] All 35 tasks (T-RT006-01..35) marked complete in tasks.md or progress.md

[HARD] Sync-phase will:
- Update spec.md §5.7 footer count if reconciliation per `plan.md` § 1.2.1 is needed.
- Generate API documentation for `moai doctor hook` and observability_events schema.
- Create release-note bullet for BC-V3R2-018.

---

Version: 0.1.0
Last Updated: 2026-05-10
