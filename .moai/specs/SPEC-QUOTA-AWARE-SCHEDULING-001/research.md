# research.md — SPEC-QUOTA-AWARE-SCHEDULING-001 (card t1347)

Codebase and platform analysis behind the SPEC. Everything cited was read or measured in this plan phase on tree `c50da9c2f` (base develop at planning time); the Claude Code facts come from one WebFetch of https://code.claude.com/docs/en/statusline. Claims not established by a command are in `plan.md` §G.

## 1. Where the usage data enters

| Fact | Source |
|---|---|
| The statusline stdin carries `rate_limits{five_hour,seven_day}` each `{used_percentage 0-100, resets_at epoch seconds}` | `internal/statusline/types.go:68,112-124` |
| It is copied into the render data and rendered, never persisted | `internal/statusline/builder.go:310-313`, `renderer.go:384-405` |
| Present only for claude.ai Pro/Max (or a gateway spend limit), only after the first API response; each window independently absent; a window is dropped once its `resets_at` passes | statusline doc (WebFetch) |
| The command re-runs on a new assistant message, `/compact`, permission mode change, vim toggle, `refreshInterval`, a window reaching `resets_at`, a warm cache reaching `expires_at`; it can go quiet while the session is idle; updates are debounced 300 ms | statusline doc (WebFetch) |
| A second, account-level 5H/7D cache exists at `~/.moai/cache/usage.json` but runs only when stdin has no `rate_limits`, and its probe sends a one-token Haiku request | `usage.go:48-55,77-80,453-479`, `builder.go:386` |

Rejected carrier: that cache (probe spends quota, not fed on the supplied path, home-scoped across projects).

## 2. The per-session record

- Schema constant 2 at `internal/statusline/context_usage.go:32`; record fields `:83-114`; atomic temp+rename writer `:176-214` with a write-if-changed throttle `sameSemanticPayload` `:265-272` that compares session id, stage, window size, integer-rounded raw percentage, model, effort and excludes capture time and writer pid.
- The single writer call site is `builder.go:181`; the record lands under the state anchor (`state_anchor.go:22-35`).
- Readers: `ReadSessionTelemetry` `context_usage.go:239-249`; consumers `factory_card.go:987-1005` (`factoryContextAtHandoffThreshold`), `internal/cli/tokens.go:383`, `internal/web/viewmodel_ops.go:289`.
- Records of dead sessions are not reaped (the rule `context-window-management-detail.md` § Detection Heuristics), so an aggregator must filter by age.
- Precedent for additive schema bumps and previous-version tolerance: `TestReadsPreviousSchemaRecord` (`session_telemetry_payload_test.go:121`) and the schema-2 comment at `context_usage.go:28-31`.

Consequence found: a steady reading never rewrites (the throttle), so `captured_at` ages while the session is alive; the design adds a heartbeat for window-carrying records only.

## 3. The 429 as an event

The StopFailure hook decodes `error_type` and `error_message` only (`internal/hook/types.go:251-253`); `stop_failure.go:46-47` handles `rate_limit` with a fixed message and no reset time; `goal/schema.go:162-188` classifies `rate_limit` as recoverable. A GLM 429 is a different account (the t1290 report quotes a z.ai `Usage limit reached for 5 hour` body), so a bare `rate_limit` event does not say which quota ran out. Hence exhaustion is derived from the windows (DO-1: not now).

## 4. Factory lane lease and the arms

`factory_card.go`: lane admission `:50-52`, refusal `:60-64`; `factoryNextLeaseOnce` `:238-258` (one call to the selection pass at `:246`, five attempts); selection `:384-521`: (a) card assigned to this lane `:426-438`, (b) picked unassigned `:442-460`, (b2) picked with no record row `:461-475`, (c) oldest eligible queued card, promoted in the same write `:476-521`; per-backend skip `:585-592` (Codex skips merge-ready or later); the verb `:608-664`; no-card exit status 3 `:181`; wait interval 5s and bound 15m `:165-168`; MCP twin `mcp_factory_card.go:111-147`. Other callers of the lease function: the CLI verb `:638`, the MCP handler `:130`, the Claude relaunch loop `factory_lane_relaunch.go:62`, the Codex loop `codex_launcher.go:968`, and tests. The `skip` function is per card and cannot tell the arms apart, so exempting arm (a) needs one boolean on the selection pass.

## 5. Backend of a lane

- Carried, not inferred: `MOAI_KANBAN_BACKEND` (`envkeys.go:227-238`), exported for the Claude and GLM launchers by `kanban.go:565-567`, set to `gpt` for Codex (`codex_launcher.go:1038`); `MOAI_LAUNCH_PROVIDER` is stamped by every launcher (`launcher.go:231-248`) and read first, with the kanban variable as fallback, by `session_start_factory.go:144-156`. `resolveMode` defaults an empty mode to `claude` (`launcher.go:55-60`).
- Constants `claude`, `glm`, `gpt` at `internal/kanban/record.go:22-24`.

## 6. The lane inventory data (for steering)

- Registry table `workers(label, pid, backend, session_id, run_id, registered_at, heartbeat_at)` `internal/homestate/factory.go:24-32`; project-scoped roster shared by runs `internal/kanban/factory_slots.go:45-47`; loader reads label, pid, registered time only `:70`; the two inserts omit `backend` `:106,:329`. The `backend` column is therefore never written.
- Liveness: `kanban.FactoryProcessAlive` (`factory_alive_unix.go:22`); the launcher exec's into Claude/GLM so the registered pid is the session's (`factory_slots.go:36-40`).
- Backend per session also lives in the kanban session record written by the session's own SessionStart hook for a lane (`session_start_record.go:60-116`; role and lane number from `kanbanRoleFromEnv` `:180-187`; backend from `MOAI_KANBAN_BACKEND`, or the launch provider for gateway sessions `:104-107`); `kanban.ReadAll` reads the directory (`record.go:258-285`). The web console joins registry pid → session registry → record and refuses ambiguous or missing joins (`internal/web/factory_lanes.go:7-13,62-155`). This route was the 0.3.0 design and is dropped at 0.4.0 (DO-12): it needs the SessionStart hook to have written a record, unobserved for Codex.

### 6.1 The claim path verified for DO-12 (0.4.0)

Every lane claim goes through one engine, `claimFactoryLane` (`internal/kanban/factory_slots.go:180-337`), entered by `ClaimFactoryLane` (`:155`), `ClaimFactoryLaneWithin` (`:173`) and the thin `ClaimFactoryLaneName` (`:115`). Production call sites (grep of `ClaimFactoryLane`, `claimFactoryLane(`, `resolveFactoryLaneName(` over non-test Go files):

| Site | Entry | Backend known there |
|---|---|---|
| `internal/cli/cc.go:250` via `resolveFactoryLaneName` (`factory.go:790-809`, which calls `kanban.ClaimFactoryLane` at `:792`) | `moai cc -f lane` | the `backend` parameter of `runClaudeEntry` (`cc.go:146`), `kanban.BackendClaude` from `cc.go:143`; the same value is stamped at `cc.go:260` |
| `internal/cli/glm.go:294` via `resolveFactoryLaneName` | `moai glm -f lane` | literal `kanban.BackendGLM` (stamped at `glm.go:304`) |
| `internal/cli/codex_launcher.go:932` via `resolveFactoryLaneName` | `moai codex -f lane` supervising loop | literal `kanban.BackendGPT` (stamped at `:951`, after the claim, in the same function) |
| `internal/cli/codex_factory.go:149` direct `kanban.ClaimFactoryLaneWithin`, reached from `codex_launcher.go:1161` | Codex factory entry | `BackendCodex` = `"codex"` (`mcp_convergence.go:65`), stamped at `codex_factory.go:144` before the claim; the loop path uses the token `gpt` — two spellings of one harness, so the claim normalizes to `gpt` |

So the value is a launcher-local literal at each site, available before the claim; none of them reads the kanban launch-facts carrier (`exportKanbanLaunchFacts`, `kanban.go:543-567`). The kanban-mode registries (`claimName` over `companions.json` and `leads.json`, `kanban.go:357-422,467-497`) are separate files and the only production callers of `SaveFactoryRegistry`; they do not touch the project factory registry rows.

Schema: `workers.backend TEXT NOT NULL DEFAULT ''` was introduced with the table itself (`git log -S` finds only commit `449b1c993`, "centralize home SQLite state (t591)"). `factorySchemaVersion` is 5 (`factory.go:20`); the migrations add columns only to `resume_handoffs`, `runs`, and `cards` (`factory.go:261-423`) and never to `workers`. Existing rows carry the empty default: the lane insert at `factory_slots.go:329`, the legacy `workers.json` import at `factory.go:540`, and the registry rewrite at `factory_slots.go:106` all omit the column. Other writers of `workers` after insert: heartbeat updates (`card_transition.go:400,445`) and the Codex pid re-stamp (`codex_factory.go:252`) touch pid and heartbeat only; `runtime_census.go:110` and `factory_run_retire.go:325-326` only read.

Consequences: feasible without kanban-mode code and without a migration; the registry reader (`:70`) and the entry type (`:36-42`) need the column added; `SaveFactoryRegistry`'s delete-and-reinsert (`:98-110`) would reset a column it does not carry, so it must round-trip the backend (its production callers operate on other files, but the factory registry's tests call it on the project registry).

## 7. The `--auto` cycle and the leader surfaces

- `runAutoCycle` `internal/cli/todo_auto.go:214-348`; options with nil-inert seams `:195-209`; production wiring `internal/cli/todo.go:275-279`; first output line is the Jev display line `:239`; per card `accept` `:262`, directive `:374-382` naming ONE isolated in-session Agent() worker, no lease, no slot. Existing tests: `todo_auto_test.go` (`TestTodoAutoSerialCycle` `:247`, `TestTodoAutoCreatesNoFactoryLease` `:395`, `TestTodoAutoClearGuidancePerCard` `:339`).
- Leader notification in code: no dynamic surface. The SessionStart leader notice (`factoryLeaderNotice` in `internal/hook/session_start_factory.go`) is composed once at start; the message store and MCP tools exist (`internal/cli/mcp_factory_msg.go:48-119`, `internal/factorymsg`) but a message is a nudge. The leader's read surface is `moai factory status` (`factory_card.go:1351`, report type `:1319-1328`, text writer `:1416-1448`).

## 8. Integration window

`moai integration acquire` is `internal/cli/integration.go:325-433`; it already carries warn-only stderr lines for configuration disagreement (`:387-406`) that leave stdout, `--json`, the record, and the exit status unchanged — the pattern the quota warning follows.

## 9. Configuration pattern

Workflow config field and struct (`types.go:501-505`, `:747-756`), defaults (`defaults.go:1161-1164`), single-block reader that defaults on every failure (`loader_slot_lease.go:19-30`), template block (`templates/.moai/config/sections/workflow.yaml:160-181`, 4-space indent, ships `enabled: false`), shipped-key inventory rows (`testdata/shipped_key_inventory.yaml:2403-2408`, class W, evidence reader) enforced by `TestShippedConfigKeysHaveReaders` (`shipped_key_reader_test.go:74`), cache schema bump (`cache.go:15-44`, now 10), template-decode test (`workflow_jev_test.go:64-89`).

## 10. The "manual quota-reset gate"

Searched Go code, rules, skills, docs, and the auto-memory store for `quota`, `reset`, `rate.?limit`, `weekly`, `limit reached`, plus Korean terms. No code or configuration gate exists. Operator practice recorded: kickoff option "승인 — 리셋 후 착수" and a scheduled wakeup at the reset (`SPEC-CI-FLAKE-SERIES-001/progress.md:9-10`); a one-shot cron resume at quota reset and "Opus re-audit at quota reset" notes in the auto-memory topics for cards t465 and t1256. The SPEC supplies the reset instant for these; it builds no scheduler.

## 11. Alternatives considered and set aside

| Alternative | Why set aside |
|---|---|
| Separate shared account file | second last-writer-wins writer, second staleness story |
| Projected-burn model | needs per-card consumption data that does not exist |
| Persisted hysteresis latch | the hold is evaluated per lease; an in-wait latch covers the only repeated evaluation |
| Message to the leader on a hold | a message is a nudge, not state; new messaging use for a printed line's job |
| Automatic re-dispatch or reassignment | excluded by the leader verdict |
| Joining registry pid → session → session record for the backend | needs the lane's SessionStart hook to have written a record (unobserved for Codex), ambiguous on both sides (`DO-12`, dropped at 0.4.0) |
