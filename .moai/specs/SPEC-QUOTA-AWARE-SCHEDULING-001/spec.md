---
id: SPEC-QUOTA-AWARE-SCHEDULING-001
title: "Quota-aware scheduling — Claude 5h/7d usage in the session record, and a Factory lane gate that steers the next card away from a nearly exhausted Claude account"
version: "0.1.0"
status: draft
created: 2026-10-02
updated: 2026-10-02
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: "internal/statusline, internal/config, internal/cli"
lifecycle: spec-anchored
tags: "quota, rate-limit, factory, lane, self-dispatch, statusline, session-telemetry, reset-time, card-t1347"
tier: M
card: t1347
depends_on: [SPEC-SESSION-TELEMETRY-001, SPEC-FACTORY-SELF-DISPATCH-001]
related_specs: [SPEC-STATE-ANCHOR-001, SPEC-FACTORY-LANE-AUTONOMY-001]
---

# SPEC-QUOTA-AWARE-SCHEDULING-001 — Quota-aware scheduling (card t1347)

## HISTORY

| Version | Date | Author | Change |
|---|---|---|---|
| 0.1.0 | 2026-10-02 | manager-spec | Initial plan-phase draft for card t1347 on `WT-quota-aware-scheduling`, base develop `c50da9c2f`. Designed for Factory mode only (self-dispatching lanes through `moai factory next`); no kanban-specific path is added or touched. Open decisions are marked `[DECISION-OPEN]` here and routed in `decision-index.md`. |

**Tier M** — about 600-900 changed lines across 12-15 files in three packages (`internal/statusline`, `internal/config`, `internal/cli`) plus the shipped template and its local twin; the 5-15 file band of the tier table, so three artifacts (spec.md, plan.md, acceptance.md) plus progress.md. REQ count 16 and AC count 16 sit at the Tier M ceilings.

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

## §B Decisions

Each decision states the options considered, the chosen default (kept minimal), and why. A decision whose default is not settled by the evidence above is marked `[DECISION-OPEN]` and routed in `decision-index.md` (no recommendation lives there; the default written here is the author's, for the lane to confirm or replace).

### D1 — Record shape

- Options: (A) extend the per-session telemetry record with the 5h/7d windows; (B) a separate shared account-level file under `.moai/state/`; (C) reuse the home-level `~/.moai/cache/usage.json`.
- Chosen default: (A) `[DECISION-OPEN DO-11]`. Each window carries used percentage, reset time (epoch seconds, the stdin unit), and, once observed exhausted, the first-observed-exhausted capture time. Schema version 2 becomes 3; the fields are optional and omitted when empty, so a version-1 or version-2 record reads as "no windows" and a window-less record differs from today's bytes only in the version number.
- Why: the record is already per-session, atomic, throttled, and read by a factory seam (`factory_card.go:987-1005`); (B) adds a second last-writer-wins writer and a second staleness story; (C) is fed by a network probe that spends quota and by a path that does not run when stdin supplies the data (`builder.go:386`).
- Exhaustion time: derived from the windows (first capture at which a window reads at or above the exhaustion percentage, kept while the window's reset time is unchanged), not from a new event writer. Under a 429 the session is quiet, so a final 100% render is not guaranteed; the hold threshold sits well below it, which is why the gate does not need the 429 event. `[DECISION-OPEN DO-1]` — stamping the StopFailure `rate_limit` turn-end into the record is deferred: the hook payload carries no reset time, the provider of the failing session (a GLM 429 is a different account) is not in the payload, and whether the hook process inherits the session environment is unmeasured. Default: not in this SPEC.

### D2 — Aggregation, staleness, rollover, fail-open

- Options: (A) freshest record per window by capture time; (B) maximum over all fresh records; (C) a single freshest record for both windows.
- Chosen default: (A), per window. A record is fresh when its capture time is within `max_age` (default 30m). A window whose reset time is not after now is **reset**: it contributes nothing and is never read as a stale high. Absence of any fresh record, an unreadable directory or record, or an unreadable configuration yields **unknown**, and unknown never holds a lane (fail open). Only records that carry the window contribute, so a GLM or Codex session (no `rate_limits`) contributes nothing.
- Why: the account quota is shared, so any live Claude session's reading describes it; freshest-wins avoids a held lane reading its own old number; per-window avoids letting a record carrying only the seven-day window mask a fresher five-hour one. Only records whose files changed within `max_age` are parsed, so the cost does not grow with the count of dead-session records.
- Interaction with the throttle: a steady reading never rewrites, so its capture time would age toward "stale" while the session is alive. The writer therefore rewrites a record that carries a window when its capture time is older than a 5-minute heartbeat (compiled constant); a window-less record is throttled exactly as before.
- `[DECISION-OPEN DO-2]` account identity: the stdin payload gives no account identifier, and profile directories suggest more than one Claude account can exist on a machine. Default: one account per project checkout, with the reset time as the only separator (windows of different accounts differ in reset time, so a mismatch is visible in `moai factory status`). `[DECISION-OPEN DO-9]` a high reading older than `max_age` but whose window has not reset: default is to treat it as stale (unknown), because the "usage never falls inside a window" premise is unmeasured and the doc calls the 5h window rolling; the alternative keeps a stale high binding until reset.

### D3 — Policy

- Options: (A) one hold percentage for both windows; (B) a hold percentage per window plus a release margin; (C) a projected-burn model.
- Chosen default: (B). Keys under `workflow.quota_gate`: `enabled` (false), `five_hour_hold_pct` (90), `seven_day_hold_pct` (95), `release_margin_pct` (5), `max_age` (30m). "Near the limit" means a fresh window at or above its own hold percentage; either window suffices. The compiled constants are the heartbeat (5m) and the exhaustion percentage (100).
- Hysteresis: a latch that lives inside one `moai factory next --wait` invocation. Once held, the verb stays held until the held window falls below `hold - release_margin` or its reset time passes. Across invocations (spaced by a card's duration) each evaluation starts fresh at the hold percentage. No persisted latch.
- Why (B): the seven-day window resets days away, so a stricter five-hour default and a laxer weekly one avoid parking Claude lanes for days on a weekly reading that the five-hour window cannot cure; (C) needs consumption-per-card data that does not exist.
- `[DECISION-OPEN DO-3]` the numeric defaults (90 / 95 / 5 / 30m) are unmeasured: no data on usage consumed per card exists in this tree. `[DECISION-OPEN DO-4]` the shipped default of `enabled`: default false in the template (sibling opt-in family: slot lease, branch guard, served-model gate), true in this repository's local `workflow.yaml`.

### D4 — Routing in Factory mode only

- Options: (A) a Claude lane declines to lease a new card while held; (B) the leader re-assigns cards away from Claude lanes; (C) per-class diversion tables.
- Chosen default: (A). While held, `moai factory next` (CLI and the `factory_next` MCP tool) leases nothing, changes no factory record, and prints one hold line (D5 carries its content); Codex and GLM lanes run the unchanged path and take unassigned cards through the existing arms (b), (b2), (c). The gate applies to every arm, including arm (a) (a card already assigned to the held lane stays with it; the leader's existing `moai factory assign --to` is the route to move it). It gates leasing only: `stage`, `complete`, and a card already leased are untouched. No card class is named; H1 (automatic class routing) and t1241 (controller) are neighbours, not part of this.
- How the leader learns: the lane's hold line in its own report, and the read-only quota block in `moai factory status` (D6). `moai todo --auto` is not steered: `internal/cli/todo_auto.go` contains no lane or backend reference (grep of `assign|lane|factory` returns only an in-session Agent() worker note at `:372,381`), and an `--auto` cycle runs inside the Claude session whose quota it would measure.
- A "Claude lane" is a lane session whose carried backend names Claude. The predicate fails open: a glm, gpt, empty, or unrecognised value is never held.
- `[DECISION-OPEN DO-5]` the CLI hold status: default exit 4 (distinct from the no-card status 3 at `factory_card.go:181`); alternative reuses 3. `[DECISION-OPEN DO-6]` arm (a): default gates it like the others (single predicate, one seam); alternative exempts a card the lane already started. `[DECISION-OPEN DO-8]` leader/`--auto` steering beyond the status block: default none. `[DECISION-OPEN DO-10]` the variable that defines the Claude predicate: default `MOAI_LAUNCH_PROVIDER` with `MOAI_KANBAN_BACKEND` as fallback (the pattern of `session_start_factory.go:144-156`), because the launch-provider variable does not depend on the kanban launcher that the removal will delete; the tokens a `moai cc` lane actually carries are measured in the M3 pre-flight.

### D5 — Reset-time planning

- Options: (A) carry the reset instant in every quota surface and warn at the integration window; (B) refuse the integration window to a pressured Claude lane; (C) add a scheduler that wakes lanes at the reset.
- Chosen default: (A). The hold line, the `moai factory status` quota block, and a warn-only line at `moai integration acquire` all carry the reset instant (RFC 3339 UTC). The gate releases by itself at the reset (a reset window reads as reset; Claude Code also re-runs the statusline at `resets_at` and then drops the window from stdin). The existing manual practice found in §A (a scheduled wakeup at the reset time) reads that instant instead of the operator's eye; no scheduler, no wakeup, no new gate is built.
- Why: the measured stall is a lane holding the integration window through a 429; a warning at the moment the window is taken is the cheapest place to make the cost visible, and a refusal would be a new gate with doctrine weight (`kanban-dispatch.md` § Integration into the release branch is self-served) that this card does not carry. `[DECISION-OPEN DO-7]` warn-only versus refuse: default warn-only.

### D6 — Observability

- Chosen default: a read-only quota block in `moai factory status` (text line and a JSON `quota` key present only when some window has data), per window: used percentage, reset time, source capture time, state (fresh, reset, unknown), whether it would hold a Claude lane. No `moai doctor` check (a new doctor check pulls in the binary-lag allowlist and its test, which is more than "cheap").

### D7 — Exclusions

Kanban mode, cross-host sessions, Jev engine choice, automatic class routing, the Codex-lane slot defects, and the StopFailure stamp are listed in §E.

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
- **REQ-QAS-009** (Event-driven) — Where the quota gate is enabled, when a Claude lane runs `moai factory next` or the `factory_next` MCP tool and a fresh window is at or above its hold percentage, the verb shall lease no card, change no factory record, print one hold line, and on the CLI exit with the hold status.
- **REQ-QAS-010** (Ubiquitous) — The hold line shall be one line beginning `quota hold:` that names each held window, its used percentage, and its reset time as an RFC 3339 UTC instant.
- **REQ-QAS-011** (State-driven) — While `moai factory next --wait` is held, the verb shall remain held until the held window's used percentage falls below its hold percentage minus the release margin or its reset time passes, re-checking at the existing wait interval until the wait bound.
- **REQ-QAS-012** (Unwanted) — The quota gate shall not hold a lane whose carried backend is not Claude, shall not gate `moai factory stage`, `moai factory complete`, or any other verb, and shall not alter a card already leased.

### C.4 Surfaces

- **REQ-QAS-013** (Ubiquitous) — `moai factory status` shall report the quota aggregate — per window used percentage, reset time, source capture time, state (fresh, reset, unknown), and whether the window would hold a Claude lane — in text and in JSON, shall omit the JSON key when no window has data, and shall change no existing line or key.
- **REQ-QAS-014** (Event-driven) — Where the quota gate is enabled, when a Claude lane runs `moai integration acquire` and a fresh window is at or above its hold percentage, acquire shall print one warn-only line naming the window and its reset time on the error stream and shall leave the lock record, the exit status, and standard output (including `--json`) unchanged.

### C.5 Scope and ordering

- **REQ-QAS-015** (Unwanted) — The change shall not add to or alter kanban-mode launch, dispatch, or companion code paths.
- **REQ-QAS-016** (Event-driven) — When the run phase begins, the baseline artifacts the acceptance criteria compare against (the serialized window-less record and the `moai factory status` output of a fixed fixture) shall be measured on the pre-change tree and committed in their own commit before the first implementation commit.

## §D Success criteria

The 16 acceptance criteria in `acceptance.md` (Given-When-Then, one decider command each) are the verification layer of REQ-QAS-001..016; the traceability table is in `acceptance.md` §C. The SPEC is done when every AC decider passes in a run whose swept count is non-empty, the shipped-key guard and the config-symmetry tests stay green, and `moai spec lint` reports no error.

## §E Exclusions

### Out of Scope — kanban mode and cross-host sessions

- Kanban mode launch, dispatch, companion sessions, and every kanban-specific code path: kanban mode is being removed; no path is added or touched (REQ-QAS-015).
- Cross-host managed sessions (card t1375): a quota record shared across machines is not designed here.

### Out of Scope — engine choice and class routing

- Jev-based engine choice (the M5 proposal) and any model-judged routing.
- Automatic card-class routing (the H1 proposal) and the factory controller (card t1241): no card class is named by the gate; a held Claude lane declines every lease, whatever the class.
- Steering `moai todo --auto` or the leader's dispatch beyond the read-only status block.

### Out of Scope — recording and scheduling

- Stamping the StopFailure `rate_limit` event into the record (`DO-1`), account identification, and any multi-account separation (`DO-2`).
- A scheduler, auto-wake, or cron at the reset time; the integration window is never refused or released by the gate (`DO-7` default warn-only).
- Codex-lane slot defects, the Codex quota (the Codex statusline carries its own `five-hour-limit` and `weekly-limit` items, `internal/codexwiring/configtoml.go:40`), GLM's own 5-hour quota, and the gateway `spend_limit` window.
- A `moai doctor` check for the quota block, and any change to the home-level `~/.moai/cache/usage.json` collector.

## §F Residual risk

- The data exists only for claude.ai Pro and Max subscribers or behind a gateway spend limit, and only after the first API response (the Claude Code statusline doc, this run). An API-key or Bedrock user gets no window, so the gate never holds: the design degrades to today's behavior, by construction.
- A Codex lane cannot advance a card at merge-ready or later (`factory_card.go:585-603`), and a card assigned to the held lane stays assigned: such cards wait through the hold. The leader's `moai factory assign --to` is the existing move; this SPEC does not automate it.
- The hold thresholds are unmeasured (`DO-3`); a too-low value parks Claude lanes early, a too-high one reproduces today's stall. The status block makes the reading visible so the values can be tuned from observation.
- Statusline updates ride assistant messages and `resets_at` and can go quiet while a session is idle; a lease decision may read a number up to `max_age` old. The heartbeat keeps a live session's record fresh; an idle session's record ages out and reads as unknown (fail open).
