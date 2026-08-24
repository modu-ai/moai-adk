---
id: SPEC-WEB-CONSOLE-015
title: "moai web console — session telemetry cells and per-lane factory progress (consumer)"
version: "0.3.0"
status: in-progress
created: 2026-08-24
updated: 2026-08-24
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: internal/web
lifecycle: spec-anchored
tags: web-console, kanban, factory, telemetry
era: V3R6
tier: M
depends_on: [SPEC-SESSION-TELEMETRY-001, SPEC-KANBAN-RECORD-SESSION-KEY-001]
related_specs: [SPEC-WEB-TODO-QUEUE-001, SPEC-WEB-CONSOLE-REDESIGN-001, SPEC-FACTORY-WORKER-FANOUT-001]
---

# SPEC-WEB-CONSOLE-015 — console consumer: session telemetry cells and per-lane factory progress

## HISTORY

| Version | Date | Change |
|---|---|---|
| 0.3.0 | 2026-08-24 | Iteration-3 audit repairs (`.moai/reports/t207/plan-audit-web-console-015-iter3.md`, seven blocking findings, all pre-run) plus the operator-ruled reclassification to Tier M. The measurement supporting M was already recorded in `plan.md` §B at 0.2.0 and is unchanged; what resolved it is that both dependency SPECs passed their own audits (0.91 / 0.857), which is the condition that section named. `design.md` and `research.md` leave the required artifact set and are retained as reference material. §A.4 withdraws the categorical claim that a record is always keyed by the launching session: measured, the join completes for some lanes and returns another lane's record. |
| 0.2.0 | 2026-08-24 | Three-way carve-out. Session telemetry (the context-usage path split, the exported reader, `moai tokens`, the doctrine and docs-site sweeps, and the model/effort producer) moved to `SPEC-SESSION-TELEMETRY-001`; the kanban record's session key, lane number, and card identifier moved to `SPEC-KANBAN-RECORD-SESSION-KEY-001`; the `/todo` route and queue-root resolution moved to `SPEC-WEB-TODO-QUEUE-001`. What remains is consumer-only. REQ/AC-WC15-012 deleted outright (it observed nothing). Surviving requirement ids keep their numbers so the two audit iterations stay traceable; the sequence is therefore gapped. |
| 0.1.0 | 2026-08-24 | Initial draft from operator card t207 (three axes), plus the iteration-2 revision landing decisions G-1..G-6. |

## §A Background

Operator card t207 named three axes for `moai web`. Two of them remain here, as **console
consumption only**:

1. the per-session telemetry cells the kanban chain view leaves blank, and
2. a per-lane factory progress section the view does not have at all.

The third axis (the todo queue section) and every producer-side change the first two need are
owned by the three sibling SPECs named in the frontmatter. This SPEC cites them and restates
none of their content.

Two screenshots accompanied the card; they were **not consulted** and form no part of these
requirements.

### A.1 The card's transport premise is false, and this SPEC corrects it

The card frames live refresh as "htmx fallback polling **vs** SSE — pick one". Both are already
built, and polling is SSE's degraded mode rather than its alternative. Measured in this tree:

| Claim | Evidence |
|---|---|
| SSE hub exists and is wired | `internal/web/events.go` — `text/event-stream` at `:81`, 25s keepalive at `:95`, 250ms debounce at `:22`; `internal/web/server.go` starts `Hub.Watch`, `internal/web/app.go` routes `GET /events` |
| Client is complete | `internal/web/assets/app.js` — `POLL_MS = 30000` (`:638`), `new EventSource("/events")` (`:721`), `startPolling()` (`:700`) reachable only from `es.onerror` at `failures >= 3` (`:743`) or the missing-`EventSource` branch |
| Events carry no payload | each listener calls `refresh(area)`, which re-fetches the fragment through htmx — render truth stays on the server |
| The htmx constraint is already handled | bundled htmx 2.0.4 core has no SSE extension, so `EventSource` is hand-wired — documented in `app.js` beside `EVENTS` |

There is therefore no transport decision left to take. The axis reduces to its data half: the
console renders empty cells because the values are not recorded, not because the refresh channel
is missing.

### A.2 The card's second premise is also false: the kanban record is not the model/effort producer

The card named "kanban.Record 확장으로 모델·effort·CW 기록" as the mechanism. Measured, the
launcher never holds the session's model on either backend: `internal/cli/cc.go` neither parses
nor sets a model — its only `--model` occurrence is the help string at `:36` — and
`internal/cli/glm.go:350-353` sets a four-slot model map (High / Medium / Low / Fable) rather than
one model, so which slot a session runs in is unknown to the launcher. The actual producer is the
statusline, which receives the session's model, effort, and context percentage per render from
the runtime.

**The intent is unchanged** — the console shows per-session model, effort, and context — but the
recording mechanism moves to `SPEC-SESSION-TELEMETRY-001`, which owns the record and the one
exported reader this SPEC consumes. This is the second correction of a card premise by
measurement; §A.1 is the first.

### A.3 What the view shows today

`internal/web/viewmodel_ops.go:250-256` builds every `RoleVM` with literal placeholders:

```go
Model:      "",  // 3단계: Record 에 모델 스냅샷이 추가되면 채운다
Effort:     "",  // 3단계
ContextPct: -1,  // 3단계: context-usage/<session-id>.json 분리 후
```

The honest rendering of those placeholders **already exists and already renders**:
`internal/web/screens.templ:165-175` selects `@missing()` when the model or effort string is
empty, and `internal/web/widgets.templ:122-124` defines that marker. Nothing in this SPEC adds a
"not recorded" marker; it fills the values the marker currently stands in for, and keeps the
marker for the cases that stay genuinely absent (§B).

`internal/web/screens.templ:192` additionally renders a hard-coded English note banner stating
that model, effort, and context usage "are not recorded yet … they fill in once kanban.Record is
extended". This change makes the first half of that sentence false and §A.2 makes the second half
false, and its third argument — the translation key — is the empty string, so it is also the one
user-visible string in this view that no locale map covers.

Lanes are invisible for a different reason: `internal/web/viewmodel_ops.go:46` fixes
`ChainRoles = []string{"lead","plan","run","sync"}` and the view iterates only those. Measured,
`internal/web` contains **zero** references to the factory registry and **zero** to a lane role.

Stage is estimated rather than recorded: `estimateStage` (`viewmodel_ops.go:266-275`) maps
heartbeat state to `StageActive` / `StageWait` / `StageBlocked` and returns an `estimated bool`
the view surfaces through `RoleVM.StageEstimated`. This SPEC keeps estimation and keeps the flag.

### A.4 The lane join — and the correction of what version 0.1.0 claimed about it

`internal/kanban/factory_slots.go` holds `map[laneLabel]FactoryWorkerEntry{PID, RegisteredAt}`
(`:37`, `:55`) — liveness only, no card, no spec, no stage — and its loader is fail-open.
`internal/session/registry.go` entries carry `PID`. So the shape of the join is
`workers.json[lane-N].PID → active-sessions entry → session_id → kanban record`.

Version 0.1.0 §A.5 asserted this "closes on today's data with no new state file". **The first half
of that claim is withdrawn; the second half survives** — nothing in this SPEC introduces a state
file.

The withdrawal is narrower and sharper than "the join returns nothing", and the difference matters
because the wrong version of it is falsifiable by a single counter-example. The record is **not
reliably** keyed by the session it describes: the launcher writes it before the launched session
exists, keyed from a single-slot side-channel file, so the key is whichever session's SessionStart
wrote that slot most recently. Sometimes that is the launched session; usually it is not. A lookup
therefore returns **nothing, the right record, or another session's record, with nothing on disk
distinguishing the second case from the third.**

Measured in this tree, with two lanes registered three minutes apart:

```
$ cat .moai/state/factory/workers.json
lane-5:  pid 87705, registered_at 2026-08-24T09:22:12Z
lane-10: pid 10793, registered_at 2026-08-24T09:25:29Z

$ .moai/state/active-sessions.json → record present?
55cdc796…  pid 87705  record: YES     ← lane-5's session
e995be8e…  pid 10793  record: no      ← lane-10's session
34740be0…  pid 31329  record: YES
e46fcfef…  pid 51045  record: no

$ cat .moai/state/kanban/55cdc796-….json     ← reached by joining lane-5
{ "session_id": "55cdc796-…", "role": "lane", "backend": "glm",
  "entered_at": "2026-08-24T09:25:29Z", … }  ← lane-TEN's registration instant

$ cat .moai/state/kanban/34740be0-….json
{ "entered_at": "2026-08-24T09:22:12Z", … }  ← lane-FIVE's registration instant
```

For `lane-5` the join **completes** — and returns a record describing a different lane, identifiable
only because the two `entered_at` values are each other's registration instants. That is the failure
this SPEC's consumer must survive, and it is worse than an empty lookup: an empty lookup renders as
"unresolved", while a completed lookup renders a confident wrong row. `lane-10` shows the empty case
in the same measurement.

The identifiers above are live and will age out. What does not age out is the property: **the record's
key is not a function of the session it describes**, so a consumer cannot tell a correct hit from an
incorrect one. `SPEC-KANBAN-RECORD-SESSION-KEY-001` is what makes the key a function of that session;
until it lands, this SPEC's rows are unreliable rather than absent.

### A.5 Dependencies

- **`SPEC-SESSION-TELEMETRY-001`** must land first, because without it no per-session telemetry
  record exists to read: the single-slot file holds whichever session rendered last. Observed
  today — while the three sessions above were live, `.moai/state/context-usage.json` held
  `{"session_id": "d281730e-…", "writer_pid": 58721, "raw_pct": 60}`, one session's telemetry,
  the other three unreadable by construction. (A reading taken minutes earlier showed
  `writer_pid 41575, raw_pct 55` for the same session id — the slot had been overwritten in
  between, which is the defect itself.)
- **`SPEC-KANBAN-RECORD-SESSION-KEY-001`** must land first, because without it the lane join does
  not close — §A.4.

## §B Requirements (GEARS)

Twelve requirements. Every one is a property of the console; no requirement here changes a
producer.

### B.0 Framing

- **REQ-WC15-001** — The console shall not add, replace, or re-select a live-refresh transport
  mechanism, and shall preserve the existing transport behaviour.
- **REQ-WC15-002** — The console shall not perform any write, mutation, or state transition
  against SPEC files, kanban records, the session registry, the factory registry, or any other
  file beneath the project's state directory.

### B.1 Session telemetry cells

- **REQ-WC15-021** — The console shall obtain each session's model, effort, and context-window
  percentage through the single reader exported by `SPEC-SESSION-TELEMETRY-001`, and shall
  declare no reader and no copy of that record's schema of its own.
- **REQ-WC15-023** — **When** no readable telemetry record exists for a session, the console
  shall render that session's model, effort, and context cells with the existing "not recorded"
  marker, and shall not substitute another session's values, a placeholder, or an inferred value.

### B.2 Per-lane factory progress

- **REQ-WC15-043** — The console shall resolve each registered factory lane to a session by
  joining the factory registry's recorded process identifier to the active-sessions entry
  uniquely carrying that identifier (REQ-WC15-047 governs the non-unique cases, which are
  reachable on both sides of the join), and shall introduce no new state file for that join.
  **When** a
  registered lane resolves to no session, or the resolved session has no record, the console
  shall still present that lane, carrying its lane number and an explicit unresolved marker.
- **REQ-WC15-044** — The console shall present, for each registered factory lane, the lane
  number, the card identifier, the SPEC identifier where one exists, the session state, and the
  stage.
- **REQ-WC15-045** — **While** a lane's stage is derived by heartbeat estimation rather than read
  from a recorded transition, the console shall mark that stage as estimated.
- **REQ-WC15-046** — **When** the factory registry is absent or unreadable, the console shall
  present the factory section as carrying no registered lanes and shall not return an error
  response.
- **REQ-WC15-047** — **When** a process identifier does not resolve to exactly one session — because
  two or more registered lanes carry it, or because the active-sessions registry carries more than
  one entry bearing it — the console shall attribute a record to none of the affected lanes and
  shall present each with an explicit unresolved marker. Both sides of the join are covered because
  neither is unique by construction: lanes may share a stale identifier in the factory registry, and
  the session registry deduplicates by session identifier alone, so two of its entries may carry one
  live process identifier.

### B.3 Cross-cutting

- **REQ-WC15-050** — Every user-visible string this SPEC adds or touches shall be present in all
  four locale maps of the console's translation catalogue.
- **REQ-WC15-051** — **When** a record or a telemetry snapshot written by a build predating this
  SPEC's dependencies is read, the console shall treat the absent fields as "not recorded", shall
  render the row, and shall not fail.
- **REQ-WC15-052** — Every user-visible explanatory string in the kanban view that today describes
  the telemetry cells as unrecorded — the section's note banner and the not-recorded marker's
  hover text — shall, after this change, state what a not-recorded cell now means (that this
  session has no telemetry record yet, not that the values are unrecordable) and shall name no
  producer that is not one. Each shall carry a translation key, like every other user-visible
  string in the view. Removing such a string rather than correcting it does not satisfy this
  requirement: the explanation is what stops a reader misreading an honest blank as a bug.

## §C Constraints

### C.1 Read-only console

Read-only is a standing console rule. It is stricter than "does not write state": the console
must not reach a producer's write path indirectly either — a resolver or loader that mutates on a
fallback branch is a write the console performs. This SPEC's dependencies are consumed through
read-only entry points only.

`internal/session.Entry` is a frozen schema (REQ-COORD-002 / REQ-COORD-024) and is **read** here;
the lane join consumes its existing `PID` field and adds nothing to it.

### C.2 No schema change remains in this SPEC

Version 0.1.0 carried two cross-cutting schema changes (the kanban record's field additions and
the context-usage path relocation). Both moved out. Nothing in §B alters an on-disk format, so
the `@MX:ANCHOR` additivity constraint at `internal/kanban/record.go:45` binds
`SPEC-KANBAN-RECORD-SESSION-KEY-001`, not this SPEC.

### C.3 Surfaces this SPEC changes

Enumerated so the change has no unlisted landing site. Why REQ-WC15-047 is a requirement rather
than an edge-case note is visible in the first two rows of §A.4's mechanism: the registry loader
is fail-open and pruning dead claims is a separate call the console is not required to make, so a
stale duplicate process identifier is reachable at render time.

| Surface | Change |
|---|---|
| `internal/web/viewmodel_ops.go:250-256` | the three `RoleVM` placeholders are filled from the telemetry reader |
| `internal/web/viewmodel_ops.go` `loadSessions` (`:409-435`) | the process identifier is currently dropped when a registry entry is mapped to `SessionVM`; the lane join needs it |
| `internal/web/viewmodel_ops.go:46` `ChainRoles` and the view model beside it | a lane collection is added **beside** the chain-role iteration, not by widening it — lanes are not chain roles |
| `internal/web/screens.templ` kanban section | the lane section is added; the generated `screens_templ.go` follows |
| `internal/web/screens.templ:192` note banner | its text no longer asserts the values are unrecorded and no longer names the kanban record as their producer; its empty third argument becomes a translation key (REQ-WC15-052) |
| `internal/web/widgets.templ:122` `@missing()` hover text | `title="not recorded anywhere yet (kanban.Record extension required)"` names the same producer §A.2 measures as false, so it is falsified by this change exactly as the banner is (REQ-WC15-052). It is also hard-coded English with no translation key |
| `internal/web/widgets.templ` | the unresolved-lane marker, if `@missing()` (`:122-124`) is not reused verbatim |
| `internal/web/assets/i18n.js` | every new key in the `en` / `ko` / `ja` / `zh` maps (REQ-WC15-050) |

Go source under `internal/web` has no mirror under `internal/template/templates/`, so the
Template-First rule does not apply to any file in this table.

### C.4 What the console may not assume about its dependencies

The two dependency SPECs land independently of this one. Until both have landed, the telemetry
cells and the lane rows read as "not recorded" — which REQ-WC15-023 and REQ-WC15-051 already
require, so a partially-landed tree renders honestly rather than incorrectly.

## §D Exclusions

Explicitly out of scope. The first three headings carve out work that was in version 0.1.0 of
this SPEC and now has a named owner, so a reader arriving from card t207 finds where each axis
went.

### Out of Scope — session telemetry production (owner: `SPEC-SESSION-TELEMETRY-001`)

- Splitting the context-usage snapshot to a per-session path, and the choice of a hard cut.
- Recording the session's model and effort, including the backend-specific model resolution.
- Exporting the reader, removing the duplicate reader in `moai tokens`, and deleting the
  single-slot validation the split makes unreachable.
- The consumer-doctrine mirror pairs and the published documentation pages naming the old path.

### Out of Scope — kanban record keying and lane identity (owner: `SPEC-KANBAN-RECORD-SESSION-KEY-001`)

- Keying a record by the session it describes, and moving the write to a point where that
  session's identifier is known.
- The lane number and the card identifier as record fields, and their derivation.
- Any launcher change. Version 0.1.0 required the launcher to record the session's model and
  effort; that requirement was **not implementable** (§A.2) and is closed by relocation rather
  than by rewording.

### Out of Scope — the todo queue axis (owner: `SPEC-WEB-TODO-QUEUE-001`)

- The `/todo` route, its nav entry, its icon case, and its area value.
- Queue-root resolution, its relocation, and the separation of resolution from queue adoption.
- Reading, listing, or badging backlog items anywhere in the console.

### Out of Scope — live-refresh transport

- Choosing between SSE and polling, or replacing either. Both ship and are wired (§A.1); polling
  is SSE's degraded mode.
- Changing the event vocabulary, the debounce, the keepalive, or the poll period.

### Out of Scope — stage recording

- Lead-recorded stage transitions. This SPEC keeps heartbeat estimation and its honesty flag;
  upgrading to recorded transitions is a separate change.

### Out of Scope — write paths, cost, and multi-project

- Any web-initiated state change: SPEC status edits, card moves, queue mutation, command
  execution.
- Displaying metered spend or any monetary figure. No code in this tree reads provider usage.
- Switching between checkouts, authentication, multi-user access. The console stays a loopback
  single-checkout single-user surface.

### Out of Scope — the accompanying screenshots

- The two images attached to card t207 were not opened and contributed nothing to these
  requirements. Any visual change they may have implied is not specified here.
