---
id: SPEC-QUOTA-AWARE-SCHEDULING-001
title: "Quota-aware scheduling — Claude 5h/7d usage in the session record, and a Factory lane gate that steers the next card away from a nearly exhausted Claude account"
version: "0.3.0"
status: draft
created: 2026-10-02
updated: 2026-10-02
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: "internal/statusline, internal/config, internal/cli"
lifecycle: spec-anchored
tags: "quota, rate-limit, factory, lane, self-dispatch, statusline, session-telemetry, reset-time, card-t1347"
tier: L
card: t1347
depends_on: [SPEC-SESSION-TELEMETRY-001, SPEC-FACTORY-SELF-DISPATCH-001]
related_specs: [SPEC-STATE-ANCHOR-001, SPEC-FACTORY-LANE-AUTONOMY-001]
---

# SPEC-QUOTA-AWARE-SCHEDULING-001 — Quota-aware scheduling (card t1347)

## HISTORY

| Version | Date | Author | Change |
|---|---|---|---|
| 0.1.0 | 2026-10-02 | manager-spec | Initial plan-phase draft for card t1347 on `WT-quota-aware-scheduling`, base develop `c50da9c2f`. Designed for Factory mode only (self-dispatching lanes through `moai factory next`); no kanban-specific path is added or touched. Open decisions were marked in-line and routed in `decision-index.md`. |
| 0.2.0 | 2026-10-02 | manager-spec | The lane resolved DO-1..DO-11 with the decision oracle (Jev). Changed: DO-5 (the hold reuses exit status 3; the stderr hold line is the discriminator) and DO-6 (the gate applies to leasing NEW cards only; a card already assigned to the lane is never held). DO-3 and DO-7 were PROVISIONAL (low oracle confidence, escalated to the leader). REQ-QAS-009 and -011 and AC-QAS-008, -009, -010 changed. |
| 0.3.0 | 2026-10-02 | manager-spec | Leader verdict on the escalations (the leader owns card scope). DO-3 accepted as provisional: the four thresholds were already configuration keys and are now stated as unmeasured defaults, with a follow-up entry. DO-7 FINAL: warn-only, never blocks taking the integration window. DO-8 REJECTED as `status block only`: steering is IN SCOPE in its minimal form — with quota pressure on, `moai todo --auto` and the `moai factory status` leader surface recommend non-Claude lanes (codex/glm) for the next card, a warning is printed when none is live, and nothing is dispatched or reassigned. Added D8, a lane inventory, REQ-QAS-017..022 and AC-QAS-017..022, a design.md and a research.md; reclassified Tier M to Tier L (22 REQ / 22 AC, more than the Tier M ceilings of 16). New open decision DO-12 (lane backend source). |

**Tier L** (reclassified from M by the leader verdict of 0.3.0). With steering in scope the change is about 1,100-1,500 lines across more than 15 files in three packages (`internal/statusline`, `internal/config`, `internal/cli`, including the new lane inventory, the `--auto` cycle and its wiring, and the status command) plus the shipped template and its local twin: above the Tier M bands on both files and lines, and 22 REQ / 22 AC exceed the Tier M ceilings of 16 each, so folding the new criteria into old ones was not an option. Tier L artifacts: spec.md, plan.md, acceptance.md, design.md, research.md, plus progress.md and decision-index.md (plan-auditor threshold 0.85; ceilings 25 / 25).

## §A Background

The card (queue text): record the 429 / weekly-limit exhaustion time as session state, share it, and plan dispatch and integration-window enforcement around the reset time. Measured by the card: on a 429 5-hour window two lanes stalled the same 554 minutes twice, and one session ended on the weekly limit. The leader narrowed the scope (operator-approved): store Claude's 5h / 7d usage in the session snapshot, and when usage nears the limit route the NEXT card to a Codex or GLM lane instead of a Claude lane. Kanban mode is being removed, so this SPEC targets Factory mode only.

Facts measured in this tree (HEAD `c50da9c2f`, every line cite below comes from a command run in this plan phase):

- Claude Code already hands the statusline the usage. `internal/statusline/types.go:112-124` defines `RateLimitInfo{FiveHour, SevenDay *RateLimitWindow}` and `RateLimitWindow{UsedPercentage float64 (0-100), ResetsAt int64 (Unix epoch seconds)}`; `types.go:68` reads it from stdin; `builder.go:310-313` copies it into the render data; `renderer.go:384-405` renders it. Nothing persists it.
- The per-session record is the natural carrier. `internal/statusline/context_usage.go:83-114` is `SessionTelemetryRecord` (schema constant `:32` = 2; fields session id, writer pid, capture time, window size, tokens, raw percentage, stage, band, model, effort); `:176-214` writes it atomically, throttled by `:265-272` (`sameSemanticPayload`, which excludes the capture time and the writer pid on purpose); `builder.go:181` is the single call site. Records of dead sessions are not reaped.
- A lane already reads its own record for a policy line: `internal/cli/factory_card.go:987-1005` (`factoryContextAtHandoffThreshold`) resolves the record by `config.EnvClaudeCodeSessionID` and treats an unreadable record as below threshold (the conservative direction).
- The lease seam: `moai factory next` is `factory_card.go:608-664`; selection is `:384-521` with arms (a) a card assigned to this lane, (b) an operator-picked unassigned card, (b2) a picked card with no record row, (c) the oldest eligible queued card; the per-backend selection seam is `:585-592` (a Codex lane skips cards at merge-ready or later); the no-card exit status is `:181` (3); the wait interval and bound are `:165-168` (5s, 15m). The MCP twin is `internal/cli/mcp_factory_card.go:111-147`, which returns plain text.
- Another account-level cache exists but is the wrong carrier: `internal/statusline/usage.go:48-55,77-80` caches 5H/7D usage at `~/.moai/cache/usage.json`, but `builder.go:386` runs it only when stdin carries no `rate_limits`, and its probe (`usage.go:453-479`) sends a one-token Haiku request — it spends the very quota under watch.
- A 429 is observable only as a turn-end event: the StopFailure hook decodes `error_type` and `error_message` (`internal/hook/types.go:251-253`); `internal/hook/stop_failure.go:46-47` handles `rate_limit` with a fixed message and no reset time.
- Claude Code documents the field (WebFetch of https://code.claude.com/docs/en/statusline in this run): `rate_limits` appears only for claude.ai Pro and Max subscribers (or behind a Claude apps gateway with a spend limit), only after the first API response in the session; each window may be independently absent; Claude Code drops a window once its `resets_at` passes; the statusline command re-runs on a new assistant message and when a window reaches `resets_at`, and can go quiet while a session is idle.
- The backend of a lane is carried, not inferred: `internal/config/envkeys.go:227-238` documents `MOAI_KANBAN_BACKEND`; `internal/cli/kanban.go:565-567` (`exportFactoryLaunchFacts`) preserves it for the Claude and GLM launchers; `internal/cli/launcher.go:231-248` stamps `MOAI_LAUNCH_PROVIDER` for every launcher; `internal/hook/session_start_factory.go:144-156` already reads the provider with the kanban backend as fallback.
- Configuration pattern for a new gate key: `internal/config/types.go:501-505` (`SlotLease` field in the workflow config), `defaults.go:1161-1164`, `loader_slot_lease.go:19-30` (single-key reader that falls back to the default on every failure), template `internal/template/templates/.moai/config/sections/workflow.yaml:160-181`, inventory `internal/config/testdata/shipped_key_inventory.yaml:2403-2408`, cache schema bump `internal/config/cache.go:44` (`configCacheSchemaVersion = 10`), template-decode test `internal/config/workflow_jev_test.go:64-89`.
- The "manual quota-reset gate" named by the card was searched for in Go code, rules, skills, docs, and the auto-memory store (patterns `quota`, `reset`, `rate.?limit`, `weekly`, `limit reached`). No code or config gate exists. What exists is operator practice recorded in SPEC progress files: a kickoff option "승인 — 리셋 후 착수" and a scheduled wakeup at the usage-limit reset time (`.moai/specs/SPEC-CI-FLAKE-SERIES-001/progress.md:9-10`), and a one-shot cron resume at quota reset (auto-memory topic for card t465). This SPEC supplies the reset instant those practices read by eye; it adds no scheduler.
- The `--auto` cycle (leader-verdict scope): `moai todo --auto` is `runAutoCycle` (`internal/cli/todo_auto.go:214-348`), wired at `internal/cli/todo.go:275-279` with `autoOptions` whose nil seams are inert (`todo_auto.go:195-209`). It prints a labelled Jev line first (`:239`), the liveness notes, the ranking record (`:255`), then per card `accept <id> …` (`:262`) and a directive (`writeAutoDirective`, `:374-382`) that names ONE isolated in-session Agent() worker — it leases no card and claims no slot. Nothing in the cycle knows a lane or a backend.
- The lane inventory data that exists: the factory registry table `workers(label, pid, backend, session_id, run_id, registered_at, heartbeat_at)` (`internal/homestate/factory.go:24-32`) is project-scoped, so one project's runs share one roster (`internal/kanban/factory_slots.go:45-47`); but its `backend` column is never written — the only inserts omit it (`factory_slots.go:106` and `:329`) and the loader reads only label, pid, and registered time (`:70`). Liveness is `kanban.FactoryProcessAlive(pid)` (`internal/kanban/factory_alive_unix.go:22`). The backend of a lane lives in the session's own record, written by its SessionStart hook with role `lane` and the lane number (`internal/hook/session_start_record.go:60-116`, role from `kanbanRoleFromEnv` at `:180-187`), whose `Backend` is `claude`, `glm`, or `gpt` (`internal/kanban/record.go:22-24,78-79`); `kanban.ReadAll` reads them (`record.go:258`). The web console already joins registry pid → session → record, and refuses an ambiguous or missing join rather than guessing (`internal/web/factory_lanes.go:7-13,62-155`).
- The leader notification surface in code today: there is no dynamic one. The leader's SessionStart notice (`factoryLeaderNotice` in `internal/hook/session_start_factory.go`) is composed once at session start (including a free-slot line from `kanban.FactoryFreeSlots`); a message store and MCP tools exist for lane-to-leader messages (`internal/cli/mcp_factory_msg.go:48-119`, `internal/factorymsg`), but a message is a nudge, never state (`cross-session-messaging.md`), and using it would be a new messaging use, not a notification surface. The leader's live read surface is the status command (`factory_card.go:1351`). D8 therefore defines the minimal notification as lines in that command and in the `--auto` output.

## §B Decisions

Each decision states the options considered, the chosen default (kept minimal), and why. Open decisions were routed in `decision-index.md` (no recommendation lives there) and have since been resolved by the lane with the decision oracle (Jev); each resolution is noted beside its decision as "resolved by decision oracle, confidence x" or, for the three escalations (DO-3, DO-7, DO-8), as the leader's verdict. DO-12 (lane backend source) is the one decision still open.

### D1 — Record shape

- Options: (A) extend the per-session telemetry record with the 5h/7d windows; (B) a separate shared account-level file under `.moai/state/`; (C) reuse the home-level `~/.moai/cache/usage.json`.
- Chosen default: (A) (DO-11 — resolved by decision oracle, confidence 1.00). Each window carries used percentage, reset time (epoch seconds, the stdin unit), and, once observed exhausted, the first-observed-exhausted capture time. Schema version 2 becomes 3; the fields are optional and omitted when empty, so a version-1 or version-2 record reads as "no windows" and a window-less record differs from today's bytes only in the version number.
- Why: the record is already per-session, atomic, throttled, and read by a factory seam (`factory_card.go:987-1005`); (B) adds a second last-writer-wins writer and a second staleness story; (C) is fed by a network probe that spends quota and by a path that does not run when stdin supplies the data (`builder.go:386`).
- Exhaustion time: derived from the windows (first capture at which a window reads at or above the exhaustion percentage, kept while the window's reset time is unchanged), not from a new event writer. Under a 429 the session is quiet, so a final 100% render is not guaranteed; the hold threshold sits well below it, which is why the gate does not need the 429 event. DO-1 (resolved by decision oracle, "not now", confidence 1.00): stamping the StopFailure `rate_limit` turn-end into the record is deferred — the hook payload carries no reset time, the provider of the failing session (a GLM 429 is a different account) is not in the payload, and whether the hook process inherits the session environment is unmeasured. Not in this SPEC.

### D2 — Aggregation, staleness, rollover, fail-open

- Options: (A) freshest record per window by capture time; (B) maximum over all fresh records; (C) a single freshest record for both windows.
- Chosen default: (A), per window. A record is fresh when its capture time is within `max_age` (default 30m). A window whose reset time is not after now is **reset**: it contributes nothing and is never read as a stale high. Absence of any fresh record, an unreadable directory or record, or an unreadable configuration yields **unknown**, and unknown never holds a lane (fail open). Only records that carry the window contribute, so a GLM or Codex session (no `rate_limits`) contributes nothing.
- Why: the account quota is shared, so any live Claude session's reading describes it; freshest-wins avoids a held lane reading its own old number; per-window avoids letting a record carrying only the seven-day window mask a fresher five-hour one. Only records whose files changed within `max_age` are parsed, so the cost does not grow with the count of dead-session records.
- Interaction with the throttle: a steady reading never rewrites, so its capture time would age toward "stale" while the session is alive. The writer therefore rewrites a record that carries a window when its capture time is older than a 5-minute heartbeat (compiled constant); a window-less record is throttled exactly as before.
- DO-2 (resolved by decision oracle, "reset time only", confidence 0.97) account identity: the stdin payload gives no account identifier, and profile directories suggest more than one Claude account can exist on a machine. One account per project checkout, with the reset time as the only separator (windows of different accounts differ in reset time, so a mismatch is visible in `moai factory status`). DO-9 (resolved by decision oracle, "treat as unknown", confidence 0.86) a high reading older than `max_age` but whose window has not reset is treated as stale (unknown), because the "usage never falls inside a window" premise is unmeasured and the doc calls the 5h window rolling.

### D3 — Policy

- Options: (A) one hold percentage for both windows; (B) a hold percentage per window plus a release margin; (C) a projected-burn model.
- Chosen default: (B). Keys under `workflow.quota_gate`: `enabled` (false), `five_hour_hold_pct` (90), `seven_day_hold_pct` (95), `release_margin_pct` (5), `max_age` (30m). "Near the limit" means a fresh window at or above its own hold percentage; either window suffices. The compiled constants are the heartbeat (5m) and the exhaustion percentage (100).
- Hysteresis: a latch that lives inside one `moai factory next --wait` invocation. Once held, the verb stays held until the held window falls below `hold - release_margin` or its reset time passes. Across invocations (spaced by a card's duration) each evaluation starts fresh at the hold percentage. No persisted latch.
- Why (B): the seven-day window resets days away, so a stricter five-hour default and a laxer weekly one avoid parking Claude lanes for days on a weekly reading that the five-hour window cannot cure; (C) needs consumption-per-card data that does not exist.
- DO-3 (accepted as provisional by the leader; oracle confidence 0.32, LOW): the numeric defaults (90 / 95 / 5 / 30m) are **unmeasured defaults** — no data on usage consumed per card exists in this tree, and nobody measured them. All four are configuration keys (`five_hour_hold_pct`, `seven_day_hold_pct`, `release_margin_pct`, `max_age`), not code constants, so the first real measurement changes a config value and no code; only the 5-minute heartbeat and the 100% exhaustion percentage are compiled constants. Adjusting them after the first measurement is listed in §E. DO-4 (resolved by decision oracle, confidence 0.99): the shipped default of `enabled` is false in the template (sibling opt-in family: slot lease, branch guard, served-model gate) and true in this repository's local `workflow.yaml`.

### D4 — Routing in Factory mode only

- Options: (A) a Claude lane declines to lease a new card while held; (B) the leader re-assigns cards away from Claude lanes; (C) per-class diversion tables.
- Chosen default: (A). While held, `moai factory next` (CLI and the `factory_next` MCP tool) leases no NEW card — arms (b), (b2), and (c) are bypassed — changes no factory record, and prints one hold line on the error stream (D5 carries its content); Codex and GLM lanes run the unchanged path and take unassigned cards through the existing arms (b), (b2), (c). Arm (a), a card already assigned to this lane, is NOT held: a held Claude lane still receives its own assigned card (DO-6, below). The gate leases nothing else: `stage`, `complete`, and a card already leased are untouched. No card class is named; H1 (automatic class routing) and t1241 (controller) are neighbours, not part of this.
- Seam (verified in `factory_card.go:384-521`): the arms run in a fixed order — (a) `:426-438`, (b) `:442-460`, (b2) `:461-475`, (c) `:476-521` — and the one selection function has a single caller, `factoryNextLeaseOnce` at `:246`. The hold is evaluated once per `next` invocation, outside the arms, and handed to the selection pass as one boolean that bypasses (b), (b2), and (c); no arm re-evaluates the predicate, so it is expressible without duplicating it. The per-card `skip` function cannot serve as the seam, because it cannot tell the arms apart. `factoryNextLeaseOnce` keeps its signature (five other call sites: the CLI verb, the MCP handler, `factory_lane_relaunch.go:62`, `codex_launcher.go:968`, and tests) and a gated variant carries the boolean, so only the CLI verb and the MCP handler change behaviour.
- How the leader and `moai todo --auto` learn and steer: D8 (steering end to end). Before 0.3.0 this bullet recorded that `internal/cli/todo_auto.go` has no lane or backend reference (grep of `assign|lane|factory` returns only the in-session Agent() worker note at `:372,381`); that fact is why D8 has to define a lane inventory rather than reuse one.
- A "Claude lane" is a lane session whose carried backend names Claude. The predicate fails open: a glm, gpt, empty, or unrecognised value is never held.
- DO-5 (resolved by decision oracle, "reuse exit status 3", confidence 0.81): the CLI hold status reuses the no-card status `factoryNextNoCardExit` (3, `factory_card.go:181`, defined "so a supervising launcher can distinguish it from failure"). The hold line on the error stream is what tells a hold from a true no-card (which prints `no card is available` on standard output). A supervising launcher that treats 3 as "back off" is the intended behaviour: during a hold there is nothing for it to do either way.
- DO-6 (resolved by decision oracle, "exempt cards the lane already holds or started", confidence 0.99): the gate applies only to leasing NEW cards; arm (a) is never held. A card the lane already owns is work in progress, not new quota spend decided at lease time.
- DO-8 (resolved by LEADER VERDICT, not by the oracle; the oracle leaned toward steering at probability 0.66 and confidence 0.32, and the lane's `status block only` override was rejected): steering is in scope in its minimal form, defined in D8.
- DO-10 (resolved by decision oracle, confidence 0.96): the Claude predicate reads `MOAI_LAUNCH_PROVIDER` first and `MOAI_KANBAN_BACKEND` as the fallback (the pattern of `session_start_factory.go:144-156`), because the launch-provider variable does not depend on the kanban launcher that the removal will delete; the tokens a `moai cc` lane actually carries are measured in the M3 pre-flight.

### D5 — Reset-time planning

- Options: (A) carry the reset instant in every quota surface and warn at the integration window; (B) refuse the integration window to a pressured Claude lane; (C) add a scheduler that wakes lanes at the reset.
- Chosen default: (A). The hold line, the `moai factory status` quota block, and a warn-only line at `moai integration acquire` all carry the reset instant (RFC 3339 UTC). The gate releases by itself at the reset (a reset window reads as reset; Claude Code also re-runs the statusline at `resets_at` and then drops the window from stdin). The existing manual practice found in §A (a scheduled wakeup at the reset time) reads that instant instead of the operator's eye; no scheduler, no wakeup, no new gate is built.
- Why: the measured stall is a lane holding the integration window through a 429; a warning at the moment the window is taken is the cheapest place to make the cost visible, and a refusal would be a new gate with doctrine weight (`kanban-dispatch.md` § Integration into the release branch is self-served) that this card does not carry. DO-7 (FINAL by leader verdict; the oracle was a coin flip at 0.54 / 0.46): the line is warn-only and acquiring the integration window is never blocked, refused, or delayed by quota state.

### D6 — Observability

- Chosen default: a read-only quota block in `moai factory status` (text line and a JSON `quota` key present only when some window has data), per window: used percentage, reset time, source capture time, state (fresh, reset, unknown), whether it would hold a Claude lane, plus the quota-pressure recommendation of D8. No `moai doctor` check (a new doctor check pulls in the binary-lag allowlist and its test, which is more than "cheap").

### D7 — Exclusions

Kanban mode, cross-host sessions, Jev engine choice, automatic class routing, the Codex-lane slot defects, and the StopFailure stamp are listed in §E.

### D8 — Steering end to end (leader verdict on DO-8)

- Options: (A) print-only recommendation in the two surfaces a leader reads — `moai todo --auto` output and the `moai factory status` block — plus the Claude-lane hold of D4; (B) additionally send a factory message to the leader when a Claude lane is held; (C) re-dispatch or reassign cards automatically. The leader fixed the form: (A), minimal; (C) is excluded outright ("no forced re-dispatch of any card"); (B) is not chosen because no messaging use is needed to make the recommendation visible.
- Quota pressure: one evaluation, shared by every surface — the gate is enabled and some fresh window is at or above its own hold percentage (the same predicate as the lane hold, minus the "caller is a Claude lane" test). A hold is pressure seen by a Claude lane. Pressure off means no steering output of any kind.
- Recommendation, in words: when pressure is on, each surface prints one line `quota pressure: <window> used=<pct>% resets_at=<RFC 3339 UTC>; recommend non-Claude lane(s): <label> (<backend>), …` naming the live registered lanes whose recorded backend is not Claude. `moai todo --auto` prints it immediately before every `accept` line (so it is re-evaluated per card); `moai factory status` prints it under the quota block and carries it in JSON under the `quota` key (pressure flag, the candidate lanes with label and backend, a warning marker, and the count of lanes of unknown backend). The line is advice for the operator or leader; the `--auto` directive still names an in-session worker and is unchanged.
- No candidate lane: when pressure is on and no live lane has a resolved non-Claude backend, the recommendation line becomes a warning line — `quota pressure: <window> used=<pct>% resets_at=<…Z>; warning: no live non-Claude lane (unknown backend: <n>); nothing is re-dispatched` — and nothing else changes. A Claude lane that is held still receives its own assigned card (DO-6) and, once the window resets, leases again.
- Lane inventory: a lane is a candidate when it is registered in the project's factory registry, its registered process is alive, and the Backend of its newest session record (role lane, same lane number) is a resolved non-Claude value. A lane with no such record, or whose newest record is ambiguous, is not a candidate and is counted as unknown; an unreadable registry or record store yields no candidates. The inventory is read-only. `[DECISION-OPEN DO-12]` the source of the lane's backend: default is the session record (the data the web console already joins), because the registry's `backend` column is never written; the alternative is to write that column when the lane claims its label, which changes the claim path and the launchers.
- Nothing is dispatched: no card is assigned, reassigned, leased, unpicked, or queued by any steering code; no lane is started or stopped; no queue entry, factory record, or session registry row is written. Card classes stay out of scope (H1 automatic class routing, t1241 controller): the recommendation does not look at the card.
- Why this form: the card's core is "route the next card to a codex or glm lane when the limit is near", and in Factory mode the next card is taken either by a self-dispatching lane (gated by D4) or chosen by the leader or the `--auto` cycle; the two surfaces above are the only places a person or leader reads before choosing, and each can be extended by one printed line without a new subsystem.

## §C Requirements (GEARS)

### C.1 Record

- **REQ-QAS-001** (Capability gate) — Where Claude Code supplies a rate-limit window (five-hour or seven-day) on the statusline stdin, the session telemetry record shall carry that window's used percentage and reset time, and shall carry no window the stdin did not supply.
- **REQ-QAS-002** (Ubiquitous) — The session telemetry record shall carry schema version 3, and the record reader shall read a record written at schema version 1 or 2 as a record with no windows rather than as an error; a record carrying no window shall serialize byte-identically to the pre-change record apart from the schema version value.
- **REQ-QAS-003** (Event-driven) — When the statusline renders, the writer shall write the record only when the throttle payload differs from the on-disk record, the throttle payload gaining each window's presence, integer-rounded used percentage, and reset time, or when the record carries a window and the on-disk capture time is older than the heartbeat interval; a record carrying no window shall be throttled exactly as before this change.
- **REQ-QAS-004** (Event-driven) — When a carried window's used percentage is at or above the exhaustion percentage, the record shall carry that window's first-observed-exhausted time, keep it unchanged across later writes while the window's reset time is unchanged, and drop it when the window is no longer carried or its reset time changes.

### C.2 Aggregation

- **REQ-QAS-005** (Ubiquitous) — The quota aggregator shall report, per window, the used percentage and reset time of the freshest record that carries the window and whose capture time lies within the max age, and shall report a window whose reset time is not after now as reset, never as stale.
- **REQ-QAS-006** (Event-driven) — When no fresh record carries a window, or the record directory, a record, or the quota configuration cannot be read, the aggregator shall report the window as unknown, and an unknown window shall never hold a lane.
- **REQ-QAS-007** (Unwanted) — The quota aggregator and the lane gate shall not open a network connection, spawn a process, or write under the state directory.

### C.3 Policy and lane gate

- **REQ-QAS-008** (Ubiquitous) — The quota gate shall be configured under `workflow.quota_gate` with the keys `enabled`, `five_hour_hold_pct`, `seven_day_hold_pct`, `release_margin_pct`, and `max_age`, whose defaults (false, 90, 95, 5, 30m) shall be identical in the Go defaults and in the shipped template, the template shall ship `enabled: false`, and an absent key, an absent file, an unparseable file, or a value outside its valid range (a hold percentage outside 1-100, a release margin outside 0-50, a non-positive or unparseable duration) shall yield the default.
- **REQ-QAS-009** (Event-driven) — Where the quota gate is enabled, when a Claude lane runs `moai factory next` or the `factory_next` MCP tool and a fresh window is at or above its hold percentage, the verb shall lease no new card (no card selected by the picked-unassigned, recorded-picked, or queued-promotion arms), change no factory record, print one hold line on the error stream, and on the CLI exit with the no-card status (3); a card already assigned to the lane shall still be leased and returned exactly as without the gate.
- **REQ-QAS-010** (Ubiquitous) — The hold line shall be one line beginning `quota hold:` that names each held window, its used percentage, and its reset time as an RFC 3339 UTC instant, and shall be the only output that distinguishes a hold from an empty queue.
- **REQ-QAS-011** (State-driven) — While `moai factory next --wait` is held, the verb shall remain held until the held window's used percentage falls below its hold percentage minus the release margin or its reset time passes, re-checking at the existing wait interval until the wait bound, and shall lease a card assigned to the lane during the wait.
- **REQ-QAS-012** (Unwanted) — The quota gate shall not hold a lane whose carried backend is not Claude, shall not gate `moai factory stage`, `moai factory complete`, or any other verb, and shall not alter a card already leased.

### C.4 Surfaces

- **REQ-QAS-013** (Ubiquitous) — `moai factory status` shall report the quota aggregate — per window used percentage, reset time, source capture time, state (fresh, reset, unknown), and whether the window would hold a Claude lane — in text and in JSON, shall omit the JSON key when no window has data, and shall change no existing line or key.
- **REQ-QAS-014** (Event-driven) — Where the quota gate is enabled, when a Claude lane runs `moai integration acquire` and a fresh window is at or above its hold percentage, acquire shall print one warn-only line naming the window and its reset time on the error stream, shall never block, refuse, or delay taking the integration window, and shall leave the lock record, the exit status, and standard output (including `--json`) unchanged.

### C.5 Scope and ordering

- **REQ-QAS-015** (Unwanted) — The change shall not add to or alter kanban-mode launch, dispatch, or companion code paths.
- **REQ-QAS-016** (Event-driven) — When the run phase begins, the baseline artifacts the acceptance criteria compare against (the serialized window-less record, the `moai factory status` output of a fixed fixture, and the `moai todo --auto` output of a fixed fixture with the Jev line stubbed) shall be measured on the pre-change tree and committed in their own commit before the first implementation commit.

### C.6 Steering (leader verdict, D8)

REQ numbers are stable: REQ-QAS-017..022 were added at 0.3.0 and sit here by topic.

- **REQ-QAS-017** (Ubiquitous) — The lane gate, the `moai factory status` quota block, the `moai todo --auto` recommendation, and the integration-window warning shall derive quota pressure from one evaluation: the gate is enabled and some fresh window is at or above its own hold percentage; a hold shall be that pressure seen by a Claude lane, and quota pressure shall be off whenever the gate is disabled or every window is below its hold percentage, reset, or unknown.
- **REQ-QAS-018** (Ubiquitous) — The steering surfaces shall list as candidate lanes the registered factory lanes whose registered process is alive and whose newest session record carries a resolved backend other than Claude, shall count every other registered live lane (no record, ambiguous record) as unknown and never list it as a candidate, and shall read the registry and the records without writing them; an unreadable registry or record store shall yield no candidates.
- **REQ-QAS-019** (Event-driven) — When quota pressure is on, `moai todo --auto` shall print, immediately before each `accept` line, one `quota pressure:` line naming each candidate lane with its label and recorded backend, and `moai factory status` shall print the same line under its quota block and carry the pressure flag, the candidate lanes, and the unknown-lane count in its JSON `quota` key.
- **REQ-QAS-020** (Event-driven) — When quota pressure is on and there is no candidate lane, the recommendation line shall be replaced by one `quota pressure:` warning line stating that no live non-Claude lane exists and how many lanes are of unknown backend, and no other output or behaviour shall change.
- **REQ-QAS-021** (Unwanted) — The steering surfaces shall not dispatch, assign, reassign, lease, unpick, queue, or otherwise change any card, queue entry, factory record, lane, or session registry row, shall not look at the card's class, and shall add no process, network call, or message.
- **REQ-QAS-022** (State-driven) — While quota pressure is off, the output of `moai todo --auto` and of `moai factory status` shall be byte-identical to its pre-change output for the same state.

## §D Success criteria

The 22 acceptance criteria in `acceptance.md` (Given-When-Then, one decider command each) are the verification layer of REQ-QAS-001..022; the traceability table is in `acceptance.md` §C. The SPEC is done when every AC decider passes in a run whose swept count is non-empty, the shipped-key guard and the config-symmetry tests stay green, and `moai spec lint` reports no error.

## §E Exclusions

### Out of Scope — kanban mode and cross-host sessions

- Kanban mode launch, dispatch, companion sessions, and every kanban-specific code path: kanban mode is being removed; no path is added or touched (REQ-QAS-015).
- Cross-host managed sessions (card t1375): a quota record shared across machines is not designed here.

### Out of Scope — engine choice and class routing

- Jev-based engine choice (the M5 proposal) and any model-judged routing.
- Automatic card-class routing (the H1 proposal) and the factory controller (card t1241): no card class is named by the gate; a held Claude lane declines every lease, whatever the class.
- Forced re-dispatch, reassignment, or automatic lease of any card by the steering code (the leader verdict: recommend and warn only), and any message sent to the leader or a lane (the factory message store is not used).
- Preferring an idle lane over a busy one in the recommendation, a lane-count or capacity model, and any use of the card's text or class.

### Out of Scope — recording and scheduling

- Stamping the StopFailure `rate_limit` event into the record (`DO-1`), account identification, and any multi-account separation (`DO-2`).
- Gating the Claude relaunch supervising loop (`internal/cli/factory_lane_relaunch.go:62`, the `--clear-policy relaunch` launcher): it leases through `factoryNextLeaseOnce`, which stays ungated, so a lane on that policy is not held. Reported to the lane as an amendment candidate; not added here to keep the change to the `next` verb and its MCP twin.
- A scheduler, auto-wake, or cron at the reset time; the integration window is never refused or released by the gate (`DO-7`, final: warn-only).
- Adjusting the thresholds after the first real measurement: the 90 / 95 / 5 / 30m defaults are unmeasured (`DO-3`); re-setting them from observed quota consumption is a follow-up, done through the four configuration keys, not part of this SPEC.
- Codex-lane slot defects, the Codex quota (the Codex statusline carries its own `five-hour-limit` and `weekly-limit` items, `internal/codexwiring/configtoml.go:40`), GLM's own 5-hour quota, and the gateway `spend_limit` window.
- A `moai doctor` check for the quota block, and any change to the home-level `~/.moai/cache/usage.json` collector.

## §F Residual risk

- The data exists only for claude.ai Pro and Max subscribers or behind a gateway spend limit, and only after the first API response (the Claude Code statusline doc, this run). An API-key or Bedrock user gets no window, so the gate never holds: the design degrades to today's behavior, by construction.
- A Codex lane cannot advance a card at merge-ready or later (`factory_card.go:585-603`); a held Claude lane still receives cards already assigned to it (DO-6), so a card the leader assigned to a Claude lane is not stranded by the hold — it simply spends quota the gate would otherwise have saved. The new cards the gate diverts reach non-Claude lanes through the unchanged arms.
- The steering is only as good as the lane inventory: a lane's backend comes from its session record, which is written by the lane's own SessionStart hook; a lane whose record is missing is counted as unknown, so the warning ("no live non-Claude lane") can fire while a usable Codex or GLM lane exists. The warning states the unknown count so the operator can see the gap (`DO-12`).
- A recommendation names lanes, not idleness: a listed lane may be busy. The card says "route the next card"; choosing which lane is free stays with the leader.
- Exit status 3 is shared by a hold and an empty queue (DO-5): a caller reading only the status cannot tell them apart; the stderr hold line is the discriminator.
- The hold thresholds are unmeasured defaults (`DO-3`, accepted as provisional); a too-low value parks Claude lanes early, a too-high one reproduces today's stall. The status block makes the reading visible so the values can be tuned from observation.
- Statusline updates ride assistant messages and `resets_at` and can go quiet while a session is idle; a lease decision may read a number up to `max_age` old. The heartbeat keeps a live session's record fresh; an idle session's record ages out and reads as unknown (fail open).
