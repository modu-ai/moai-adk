---
id: SPEC-WEB-CONSOLE-015
title: "moai web three-axis improvement — recorded session telemetry, todo queue section, per-lane factory progress"
version: "0.1.0"
status: draft
created: 2026-08-24
updated: 2026-08-24
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: internal/web
lifecycle: spec-anchored
tags: web-console, kanban, factory, todo, telemetry
era: V3R6
tier: L
related_specs: [SPEC-KANBAN-TODO-CLI-001, SPEC-WEB-CONSOLE-REDESIGN-001, SPEC-HANDOFF-THRESHOLD-001, SPEC-FACTORY-WORKER-FANOUT-001]
---

# SPEC-WEB-CONSOLE-015 — moai web three-axis improvement

## HISTORY

| Version | Date | Change |
|---|---|---|
| 0.1.0 | 2026-08-24 | Initial draft from operator card t207. |

## §A Background

The operator's card t207 names three axes for `moai web`:

> "moai web 개선 3축 (실시간 갱신: htmx 폴백 폴링 vs SSE 택일, kanban.Record 확장으로 모델·effort·CW 기록 / todo 섹션: moai todo 큐 스토어 연동 / 팩토리 진행상황: lane별 카드·단계 표시)"

That card text is the **requirements SSOT** for this SPEC. Two screenshots accompanied the
card; **they were not consulted** and form no part of these requirements.

### A.1 The card's first axis carries a false premise, and this SPEC corrects it

The card frames live refresh as "htmx fallback polling **vs** SSE — pick one". Both are
already built, and polling is SSE's degraded mode rather than its alternative. Measured in
this tree at `28bde4022`:

| Claim | Evidence |
|---|---|
| SSE hub exists | `internal/web/events.go` — `Hub`, `ServeEvents` (`text/event-stream`, 25s keepalive at line 95), `Hub.Watch` (fsnotify, 250ms debounce at line 22) |
| It is wired | `internal/web/server.go:251` calls `s.app.hub.Watch(...)`; `internal/web/app.go:162` routes `GET /events` |
| Client is complete | `internal/web/assets/app.js` — `EVENTS = ["spec","session","goal","verify","kanban","config"]`, `POLL_MS = 30000`, `connect()` opens `new EventSource("/events")`, `es.onerror` calls `startPolling()` at `failures >= 3`, a `ready` event calls `stopPolling()` |
| Polling is the fallback, not a rival | `startPolling` is reachable only from `es.onerror` (3 consecutive failures) or from the `!window.EventSource` branch |
| Events carry no payload | each listener calls `refresh(area)`, which re-fetches the fragment through htmx — render truth stays on the server |
| The htmx constraint is already handled | bundled htmx is 2.0.4 core with no SSE extension, so `hx-ext="sse"` is unavailable and `EventSource` is hand-wired — documented at `app.js:626-627` |

There is therefore **no transport decision left to take**. Axis 1 reduces to its data half:
the console renders empty cells because the values are not recorded, not because the refresh
channel is missing.

### A.2 The data half is already marked in the code

`internal/web/viewmodel_ops.go:250-256` builds every `RoleVM` with literal placeholders:

```go
Model:      "",  // 3단계: Record 에 모델 스냅샷이 추가되면 채운다
Effort:     "",  // 3단계
ContextPct: -1,  // 3단계: context-usage/<session-id>.json 분리 후
```

`kanban.Record` (`internal/kanban/record.go:55-99`) carries
`{SessionID, SpecID, Role, Backend, EnteredAt, DeepScanDir, VerifyRung, VerifyReentries}` —
no model, no effort, no context. The producers of model and effort exist
(`internal/config/profile.go:73-76` `ModelEffort`; `internal/settings/agentfm` `AgentInfo`),
they are simply not threaded into `recordKanbanSession` (`internal/cli/kanban.go:472`), whose
signature is `(specID, backend, role string)` and whose eight callers live in
`internal/cli/cc.go` (161/175/192/208) and `internal/cli/glm.go` (224/237/250/264).

### A.3 Context-window usage is structurally unreadable per-session

`internal/statusline/context_usage.go:134` writes exactly one file per project root,
`.moai/state/context-usage.json`, overwritten wholesale by whichever session last rendered
its statusline — last-writer-wins, with `session_id` and `writer_pid` inside the record
rather than in the path. A prior investigation observed the race live
(`.moai/reports/webredesign/moai-web-menu-spec.md` §4.6): session `368a2bd9…` at 260,000
tokens replaced by session `e463a3c9…` at 0. Reading N lanes' usage from one slot is
impossible by construction.

The split the code comment already anticipates —
`.moai/state/context-usage/<session-id>.json` — is **not a rename**. Its blast radius is
enumerated in §C.3.

### A.4 The todo queue has no console consumer, and one path trap

Nothing under `internal/web` references the backlog. The store is ready to be read:
`internal/kanban/backlog_store.go` exports `NewBacklogStore`, `Load`, `BacklogPathForRoot`,
and `QueuedBacklogCountForRoot`, all lock-free on the read path. `BacklogItem` is the frozen
five-field contract `{ID, Text, AddedAt, SpecID *string, State}`, with `BacklogState` in
`queued | picked | dropped`.

The trap is root resolution. `resolveTodoQueueRoot()` is unexported at
`internal/cli/todo.go:66` and deliberately resolves the **primary checkout** through git's
common directory, never the worktree the process runs in. Its own comment records the
measured incident: 30 queued cards on the primary, "queue is empty" from a linked worktree
(2026-08-17). `moai web` can be launched from inside a worktree, so it must use the same
resolution — and a second implementation of it is a second chance to fork the queue.

### A.5 The factory lane join is derivable; the card id is not

`internal/kanban/factory_slots.go` holds `map[laneLabel]FactoryWorkerEntry{PID, RegisteredAt}`
at `.moai/state/factory/workers.json` — liveness only, no card, no spec, no stage. Its loader
is fail-open (unreadable ⇒ every slot free).

`internal/session/registry.go:86-95` entries carry `{session_id, spec_id, phase, started_at,
last_heartbeat, pid, host, cwd}` — **including `PID`**. So
`workers.json[lane-N].PID → active-sessions entry → session_id → kanban.Record`
is a join that closes on today's data with no new state file.

Three things are genuinely missing:

1. **The card id.** `Record.SpecID` is a SPEC ID, not a card id (`t207`), and Class A / Class B
   cards never acquire a SPEC at all — precisely where the board most needs a label.
   `MOAI_KANBAN_ID` is **not** it: `internal/config/envkeys.go:167-173` documents it as the
   *run* identifier (`tk4ntu`), set once per run at `internal/cli/factory.go:255` and
   `kanban.go:173`. No per-lane card id exists on disk anywhere.
2. **The lane number.** `internal/kanban/role.go:42` defines `RoleLane = "lane"` as a bare
   constant, so every factory lane writes `role: "lane"` and N lanes are mutually
   indistinguishable. `WithRole` (`record.go:116-130`) silently drops unrecognised roles, so a
   naive `lane-3` value would vanish without error.
3. **Lane visibility in the view.** `internal/web/viewmodel_ops.go:46` fixes
   `ChainRoles = []string{"lead","plan","run","sync"}` and the view iterates only those.

Stage is estimated rather than recorded: `estimateStage` maps heartbeat state to
`StageActive` / `StageWait` / `StageBlocked` and returns an `estimated bool` the UI surfaces.
The prior investigation weighed inference-from-SPEC-status, lead-recorded transitions, and
heartbeat estimation, and recommended starting at estimation and upgrading later. This SPEC
keeps estimation and keeps the honesty flag.

## §B Requirements (GEARS)

### B.0 Framing

- **REQ-WC15-001** — The implementation shall not add, replace, or re-select a live-refresh
  transport mechanism; `internal/web/events.go` and the `connect()` / `startPolling()` block
  of `internal/web/assets/app.js` shall be unchanged in transport behaviour.
- **REQ-WC15-002** — The console shall not perform any write, mutation, or state transition
  against SPEC files, the backlog queue, kanban records, the session registry, or the factory
  registry.

### B.1 Axis 1 — recorded model, effort, and context-window usage

- **REQ-WC15-010** — The `kanban.Record` struct shall carry a model field and an effort field,
  each a string with `omitempty`, added without renaming or removing any existing JSON key.
- **REQ-WC15-011** — **When** a session enters kanban or factory mode through `moai cc` or
  `moai glm`, the launcher shall record into that session's `kanban.Record` the model and the
  effort resolved for the session at launch.
- **REQ-WC15-012** — **When** a `kanban.Record` is read whose model or effort field is empty,
  the console shall render that cell as an explicit "not recorded" marker and shall not
  substitute a blank cell, a placeholder value, or an inferred value.
- **REQ-WC15-020** — The statusline shall persist its context-usage snapshot to a per-session
  path `.moai/state/context-usage/<session-id>.json`.
- **REQ-WC15-021** — The `internal/statusline` package shall export exactly one reader for the
  context-usage record, and the console shall consume that exported reader.
- **REQ-WC15-022** — The implementation shall not define a second declaration of the
  context-usage record schema outside `internal/statusline`.
- **REQ-WC15-025** — The pre-existing duplicate declaration in `internal/cli/tokens.go`
  (`tokensContextSnapshotFilename`, `tokensContextSnapshot`) shall be removed and that call
  site migrated onto the exported reader of REQ-WC15-021, in the same milestone that moves the
  path. Verification: after the change, `grep -rn '"context-usage' internal/` returns the
  statusline declaration and no other, and `moai tokens` still emits a context snapshot for a
  session that has one. Rationale: this reader shares no constant with the statusline, so the
  path move breaks it with no compile error and no runtime error — the snapshot simply stops
  appearing.
- **REQ-WC15-023** — **When** no readable per-session context-usage record exists for a
  session, the console shall render that session's context percentage as "not recorded" and
  shall not fall back to another session's record.
- **REQ-WC15-024** — **Where** the per-session context-usage path is adopted, the consumer
  doctrine at `.claude/rules/moai/workflow/context-window-management.md` § Detection
  Heuristics shall be updated to name the new path and to drop the single-slot validation
  steps the split makes unreachable, and its template mirror at
  `internal/template/templates/.claude/rules/moai/workflow/context-window-management.md`
  shall be updated in the same change.

### B.2 Axis 2 — todo queue section

- **REQ-WC15-030** — The console shall present a section listing the backlog queue's items
  with, per item, its id, its text, its state, and its SPEC id where one is attached.
- **REQ-WC15-031** — The console shall resolve the backlog queue root through the same
  primary-checkout resolution `moai todo` uses, relocated into a package both `internal/cli`
  and `internal/web` import, with `internal/cli/todo.go` delegating to that relocated
  resolution rather than retaining its own copy.
- **REQ-WC15-032** — The console shall not call any mutating backlog operation and shall not
  acquire the backlog lock.
- **REQ-WC15-033** — **When** the backlog queue file is absent, empty, or unreadable, the
  console shall render an empty-state for the section and shall not return an error response.
- **REQ-WC15-034** — **When** a `kanban` live-refresh event fires, the todo section shall be
  re-fetched through the existing refresh path.

### B.3 Axis 3 — per-lane factory progress

- **REQ-WC15-040** — A factory lane's `kanban.Record` shall carry the lane's number as data
  distinct from its role value.
- **REQ-WC15-041** — **When** a lane label carrying a lane number is recorded, the record
  writer shall preserve the lane number and shall not discard it as an unrecognised role.
- **REQ-WC15-042** — A `kanban.Record` shall carry the card identifier the session is working,
  as a field distinct from the SPEC identifier, populated for cards that never acquire a SPEC.
- **REQ-WC15-043** — The console shall resolve each registered factory lane to its session by
  joining the factory registry's recorded PID to the active-sessions registry entry bearing
  that PID, and shall not introduce a new state file for that join.
- **REQ-WC15-044** — The console shall present, for each registered factory lane, the lane
  number, the card identifier, the SPEC identifier where one exists, the session state, and
  the stage.
- **REQ-WC15-045** — **While** a lane's stage is derived by heartbeat estimation rather than
  read from a recorded transition, the console shall mark that stage as estimated.
- **REQ-WC15-046** — **When** the factory registry is absent or unreadable, the console shall
  render the factory section as carrying no registered lanes and shall not return an error
  response.

### B.4 Cross-cutting

- **REQ-WC15-050** — Every user-visible string added by this SPEC shall be present in all four
  locale maps of `internal/web/assets/i18n.js` (`en`, `ko`, `ja`, `zh`).
- **REQ-WC15-051** — **When** a record is written by a build predating this SPEC's schema
  additions, every reader shall treat the absent fields as "not recorded" and shall not fail,
  and rewriting such a record shall not alter its pre-existing keys.

## §C Constraints

### C.1 Schema additivity

`internal/kanban/record.go:45` carries an `@MX:ANCHOR` whose reason states that the launcher,
the orchestrator, and the sync-phase dedup gate all bind to these JSON keys, so a renamed key
breaks readers the package cannot see. Every field this SPEC adds is additive and
`omitempty`; no existing key is renamed or removed.

`internal/session/registry.go` `Entry` is frozen per REQ-COORD-002 / REQ-COORD-024 and is
**read only** by this SPEC — the lane join (REQ-WC15-043) consumes its existing `PID` field
and adds nothing to it.

### C.2 Read-only console

Read-only is a standing console rule (`.moai/reports/webredesign/moai-web-menu-spec.md` §2.2,
§7). `SPEC-KANBAN-TODO-CLI-001` (`status: in-progress`) **owns** the backlog store — lock-guarded
writes and id issuance. This SPEC is a read-only consumer of that store and takes no ownership
of it.

### C.3 Context-usage split blast radius (not a rename)

| Surface | Change |
|---|---|
| `internal/statusline/context_usage.go` | writer path (line 134), `readContextUsage` (line 186), `contextUsageRecord` (line 56) — reader must be exported per REQ-WC15-021; `isFreshForSession` / `sameSemanticPayload` / `isRealSessionID` validation written for a single slot becomes partly unreachable |
| `internal/statusline/builder.go:157` | call site of the persistence step |
| `internal/cli/tokens.go` | **a second, independent reader** — `tokensContextSnapshotFilename = "context-usage.json"` (line 30) is its own hardcoded filename constant, and `tokensContextSnapshot` (line 79) its own duplicate of the record schema. It embeds the snapshot into `moai tokens` output. Because it neither imports the statusline reader nor shares the constant, the split breaks it **silently**: the file simply stops being found and the snapshot silently drops out of the output. Exporting one reader (REQ-WC15-021) is not enough — this call site must be migrated onto it, and REQ-WC15-022's single-declaration rule applied to `tokensContextSnapshot` as well |
| `internal/statusline/{builder,context_usage}_test.go` | assert the literal path `.moai/state/context-usage.json` in at least five places |
| `.claude/rules/moai/workflow/context-window-management.md` | § Detection Heuristics names the file as the authoritative snapshot and specifies session-id match, `writer_pid` discriminator, and freshness check — a read procedure built around the single-slot shape (line 100) |
| `internal/template/templates/.claude/rules/moai/workflow/context-window-management.md` | Template-First mirror of the above |
| docs-site | `content/{en,ko,ja,zh}/advanced/statusline.md`, `advanced/token-budget.md`, `cli-reference/tokens.md` reference the path — 12 files, four locales |
| `.moai/README.md` | references the path |

### C.4 Template-First

Go source under `internal/web`, `internal/kanban`, `internal/cli`, and `internal/statusline`
has **no** mirror under `internal/template/templates/` (that tree carries only `.claude/`,
`.moai/`, and root config), so the Template-First rule does **not** apply to the code changes.
It applies to exactly one artifact in this SPEC: the doctrine-rule edit of REQ-WC15-024, whose
mirror must be updated and followed by `make build`.

### C.5 Existing view constraint

`internal/web/viewmodel_ops.go:46` `ChainRoles` fixes the four chain roles the kanban view
iterates. Factory lanes are not chain roles (`role.go:42` states lanes never occupy the
three-role chain), so lane presentation is an addition beside that iteration, not a widening
of it.

## §D Exclusions

Explicitly out of scope. Each may be taken up separately.

### Out of Scope — live-refresh transport

- Choosing between SSE and polling, or replacing either. Both ship and are wired (§A.1);
  polling is SSE's degraded mode.
- Adding an htmx SSE extension or migrating to `hx-ext="sse"`.
- Changing the event vocabulary, the 250ms debounce, the 25s keepalive, or the 30s poll period.

### Out of Scope — write paths

- Any web-initiated state change: SPEC status edits, card moves, queue mutation, command
  execution.
- Backlog id issuance or lock-guarded writes — owned by `SPEC-KANBAN-TODO-CLI-001`.

### Out of Scope — stage recording

- Lead-recorded stage transitions. This SPEC keeps heartbeat estimation and its honesty flag;
  upgrading to recorded transitions is a separate change.

### Out of Scope — cost and usage billing

- Displaying GLM metered spend or any monetary figure. No code in this tree reads z.ai usage.

### Out of Scope — multi-project and auth

- Switching between checkouts, authentication, multi-user access. The console stays a
  loopback single-checkout single-user surface.

### Out of Scope — the accompanying screenshots

- The two images attached to card t207 were not opened and contributed nothing to these
  requirements. Any visual change they may have implied is not specified here.
