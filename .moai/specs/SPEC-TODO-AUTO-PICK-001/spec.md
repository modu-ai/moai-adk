---
id: SPEC-TODO-AUTO-PICK-001
title: "Autonomous card selection under --auto — the invoked session judges, the lease is the only pick path, the keep-set is skipped and reported"
version: "0.2.0"
status: draft
created: 2026-10-02
updated: 2026-10-02
author: manager-spec (card t1448)
priority: P1
phase: "v3.2.0 target"
module: "internal/cli, internal/template/templates, .claude/rules, .claude/skills, .claude/agents"
lifecycle: spec-anchored
tags: "todo-auto, factory-next, card-selection, keep-set, lease, decision-record, doctrine-amendment, card-t1448"
tier: M
card: t1448
depends_on: [SPEC-FACTORY-SELF-DISPATCH-001, SPEC-TODO-CLASSIFY-DISPATCH-001, SPEC-TODO-HOLD-STATE-001, SPEC-TODO-AUTO-PRIORITY-001]
related_specs: [SPEC-AUTONOMY-BATCH-GATE-001, SPEC-JEV-AUTO-EXCEPTION-001, SPEC-FACTORY-LANE-AUTONOMY-001, SPEC-RELATION-PICKUP-FILTER-001, SPEC-MANAGER-TODO-001, SPEC-AUTONOMY-GATE-REWIRE-001]
---

# SPEC: autonomous card selection under `--auto`

## HISTORY

- 0.2.0 — 2026-10-02 — plan-audit iteration 1 repair (FAIL 0.73, MP-9 failed on two ordering
  conflicts; audited commit `b3646de10`). Live and mirror doc edits now land in one milestone
  (D1); the bare-path golden is GREEN on the unmodified tree with a seeded-perturbation
  non-vacuity cell (D2); the lane `--auto` refusal gets a defined predicate, a dedicated refusal
  text, a lease-path routing sentence, and the shipped `factory fallback declare` instruction is
  corrected (D3, D4); the record's location is stated and the compensating-control claim is
  retracted (D5); a surgical-diff criterion and the full list of pins to move are added (D6); the
  nominee leaves queue and record unchanged on refusal, with a normative token set (D7); the
  mutant holes are closed (D8); two guard criteria are reclassified (D9); PR/landed state is a
  skip input only for `queued` candidates (D10). The requirement count stays 16 (Tier M ceiling);
  the criterion count rises from 12 to 14. Resolution map: `plan.md` § Audit-1 resolution map.
- 0.1.0 — 2026-10-02 — plan-phase artifact set authored (card t1448; worktree
  `.moai/worktrees/t1448`, branch `WT-todo-auto-pick-autonomy`, plan-start HEAD `4bf547bca`).
  Tier M: `spec.md` + `plan.md` + `acceptance.md`, plus `research.md` and `spec-compact.md`.
  Operator inputs: the card text (items 1-6) and the leader's constraint that the selection
  inputs are an open extension point for card t1454 (§B.7).

## §A Context

### A.1 Problem

A measured trigger (mo.ai.kr factory screenshot, 2026-10-02): six lanes entered `--auto` and each
stopped on an `AskUserQuestion` — "which card should I take" — so no lane made concurrent
progress. The standing rule reads `--auto` as authorizing **serial queue-order consumption and
nothing else** (`kanban-dispatch.md` § Entry into the board; the `gtd.md` `--auto` section;
`auto-semantics.md` §9 card-pick row), and states that the leader "never picks for the operator".
A session reading that text is correct to ask: nothing authorizes it to choose, and nothing lets
six lanes consume a queue serially at once.

The mechanical side has the opposite gap. `moai factory next` chooses for the lane by stored
priority order (`factory_card.go:402-542`); it has no flag to name a card, so even a session that
is allowed to judge cannot express the judgment (`research.md` R1). And a lane session may run
`moai todo --auto` today (`research.md` R2/O4, and the re-measurement in R11) — a second,
queue-level pick path outside the lease — while the shipped `moai factory fallback declare` tells
a lane in fallback to do exactly that (`factory_messaging.go:255`).

### A.2 What this SPEC changes

It moves the **owner of the selection judgment** to the invoked session, gives that judgment a
mechanically validated way to become a lease, fixes the lease as the only pick path for a lane, and
writes down — and where possible enforces — the set of cards autonomous selection must skip. It
leaves queue **admission** (card production) with the operator.

### A.3 Verified basis (pinned tree `4bf547bca`; repair measurements on `b3646de10`)

`research.md` carries the evidence and marks each fact OBSERVED, READ, or INFERRED.

- `factory next` has three flags and no nomination (R1). Arm (c) skips `blocked` cards and
  held-slot serial cards, and **does not skip** a `[보류`-opening card (probe O1) or a
  relation-blocked card (probe O2); the `hold` state is skipped by positive state enumeration (O3).
- The concurrency core is already pinned by four existing tests (R3); the new tests assert it
  through the nominated path.
- No operator-decision-queue marker exists anywhere in code or docs (R4-ii); the queue schema is
  frozen additive-only, so none will be added (§B.3).
- No Go code writes the §11 decision board (R4-iii); the decision record is a self-attested line
  kept as evidence, and no party executes a re-read of it today (§B.5, §G).
- The sentences to amend are pinned by existing tests, one of which is a passage start marker (R6).
- A session whose only lane variable is the lane label, and a non-lane session whose only marker
  is the Codex backend, both run `moai todo --auto` today (R11); the refusal must separate them.

## §B Decisions

### B.1 Q1 — mechanism: an opt-in nomination form of `factory next`

`moai factory next --card <id>` (and the MCP `factory_next` parameter `card`) names one card. The
CLI validates the nominee **before any write**, and leases it through the same version-checked
edges the arms use, or refuses with a distinct exit code and a one-line reason token so the
session re-selects. With no flag, behavior is the current priority order (with the single
exception D-DEF in B.4). Alternatives measured in `research.md` R8: a candidate-list hand-back
(state store, stale list) and a separate `peek` verb (duplicates the lane-readable `list --json`,
`why`, `pr`, `show`). The read half of "judge then pick" already exists; only the validated write
half was missing.

### B.2 Q4 — boundaries that stay true

- Queue **admission** stays the operator's: `moai gtd add/drop/edit/done/relate` stay refused to a
  lane (`todoLaneReadOnlyVerbs`), `moai contract sign` stays forbidden, and the leader remains the
  queue's sole producer.
- The self-dispatch lane exception stays and becomes the **sole** lane pick path: `moai factory
  next` is the lane's only promotion.
- Jev stays display-only. Nothing in this SPEC lets a Jev answer choose a card or write the queue.
- The `--auto` invocation remains the operator's batch approval; a lane's self-service pickup is
  that approval acted on, never a self-grant.
- **Reconciliation with the pre-dispatch PR cross-check.** The cross-check "reports, never
  vetoes" a card the **operator** picked. Therefore the pull-request/landed state is a **skip
  input only for a `queued` candidate the session chose itself**; for a `picked` card (an operator
  pick) it is **report-only** — the session reports what it read and never passes the card over
  for it. The two clauses govern different cards and do not conflict; the stub's `confirms or
  withdraws` literal stays untouched. The sentence lands in the amended
  `kanban-dispatch-detail.md` paragraph and in § C.1.

### B.3 Q2 — the keep-set marker: reuse, add nothing

The queue record's per-item contract is frozen additive-only (R4-ii), and no operator-decision
marker exists. **No new marker, state, or field is introduced.** An operator-decision-queue card is
expressed with what exists: the structural `hold` state (written only by `moai gtd hold`, "operator
(or lead) only: no lane session or machine leaser can hold or unhold") and, for prose parking, the
`[보류` text marker. Both are mechanically enforced on the lease path (REQ-TAU-009, -007). A card
the operator means to keep out but has neither held nor marked is **not** mechanically protected; it
falls to the text-judgement half of the keep-set, whose evidence is recorded (REQ-TAU-010). That
failure mode is stated, not hidden (§G). The same marker has two treatments, which the doctrine
names: it **demotes** a card in the serial cycle and **excludes** it on the lease path (REQ-TAU-011).
The cards the memory index calls operator-decision-queue cards (t810, t1294, t1383) were read
through the installed binary during the plan audit: t1294 and t1383 `queued` with ordinary text,
t810 `picked` — so they are selectable today; a lane cannot mutate the queue, so the operator or
leader must hold or mark them before lanes exercise this doctrine (Definition of Done item 6).

### B.4 The two deliberate deltas, each isolated so the leader can veto it

- **D-DEF — the default arm skips `[보류`-marked cards (second clause of REQ-TAU-007).** Today the
  bare `factory next` arm (c) leases a card whose text opens with `[보류` (probe O1; readable at
  `factory_card.go:511-523`, which has no marker check), so the keep-set guard would be bypassed by
  omitting the flag. REQ-TAU-007 states the unchanged default for every queue that holds no
  marker-bearing queued card and names the skip as the single exception; the skipped card counts
  as seen, so a queue holding only such cards ends on the no-card exit and not on a retry. The
  skip has its own criterion subtest (AC-TAU-004, AC-TAU-006), so dropping the clause and those
  subtests leaves the rest intact.
- **D-LANE — a lane session is refused `moai todo --auto` (REQ-TAU-008), with the lease path as
  the replacement.** A lane can run the serial cycle today (probe O4, R11), picking in the queue
  outside the lease. The SPEC defines **who is a lane session** for this refusal, fixes **what the
  refusal says**, and corrects **what the shipped fallback verb tells a lane to do**:
  - *Predicate.* A lane session is one whose lane-role marker equals `lane`, or whose lane-label
    variable is non-empty. The Codex backend marker alone does **not** make a session a lane for
    this refusal: a Codex-backend leader or non-lane session keeps its batch approval (R11 M1 shows
    it runs `--auto` today and has no lease alternative, since `factory next` refuses outside a
    lane). A Codex lane carries its lane label and is refused. The shared predicate
    `factoryLaneRefusal()` (four other call sites) is **not** changed.
  - *Text.* A dedicated refusal — not the queue-mutation text — says the `--auto` authorization is
    exercised through `moai factory next --card <id>` (or bare `moai factory next`), so a lane that
    reads it proceeds instead of stopping to ask.
  - *Routing.* The amended `gtd.md` `--auto` section routes a lane to the lease path in a pinned
    sentence, and carries the keep-set list and the record form itself (a lane reads gtd.md and
    may never load `auto-semantics.md` §9.3).
  - *Reconciliation with SPEC-FACTORY-LANE-AUTONOMY-001.* That SPEC's REQ-FLA-001 says a lane in
    messaging fallback switches to `/moai:todo --auto` self-service pickup, and `factory fallback
    declare` prints that string. **In lane sessions that reference is superseded by the lease
    path** (`moai factory next [--card <id>]`); REQ-FLA-006/-007's classification consumption is
    arm (c)'s rule and is unchanged. That completed SPEC's body is not edited; the printed string
    is.
  - *Card item 6.* "Two concurrent `--auto` lanes lease different cards" is satisfied through the
    nominated path (AC-TAU-002), because the literal `moai todo --auto` is no longer runnable in a
    lane — and the serial cycle it ran was exactly the double-pick surface item 2 closes.
  Dropping D-LANE (REQ-TAU-008, M3, AC-TAU-005) removes only that refusal; the doctrine then reads
  "in a lane `--auto` means the lease path" without a mechanical guard.

### B.5 Q3 — the decision record

Reuse the §10 one-line form (`decision record: decided_by= evidence_refs= ladder_path=`). The
card-pick row's `ladder_path` token is `gate-row card pick (AUTONOMOUS, auto-semantics §9)`,
following the precedent of the plan→run Kickoff row. `evidence_refs` carries `key=value` items
joined by `;`: `card`, `class`, `relate`, `pr`, `wt`, `overlap`, `skipped`, each `unmeasured` when
the input could not be read.

**Location, and what it is not.** The line is written in the card's progress record
(`.moai/specs/<SPEC-ID>/progress.md` §F where a SPEC exists, else under `.moai/reports/<card-id>/`).
It is **evidence, not the §11 decision board**: no lane-writable board verb exists (R4-iii), and
`auto-semantics.md` §11 calls a tree-local copy "a ghost — never a board". The line makes no claim
to be the board. It is **self-attested**; **no party executes a re-read of it today** — the
sync-audit instructions carry no decision-record step (measured in the plan audit) and this SPEC
does not add one (scope growth on surfaces the t1453/t1454 cards may touch). That is an
**unimplemented residual risk** (§G), not a compensating control.

### B.6 Q5 — the foreman skill

Amended minimally, not restructured: Boundary 1's "serial consumption in queue order is
authorized" becomes "serial consumption … on the iteration's own judgment, within the keep-set",
and step 4's "`queued` items are not yours to pick" gains "outside a batch authorization". The
mechanical removal of `AskUserQuestion`, the one-worker-at-a-time serialization, and "never run
`moai gtd next <n>`" are unchanged. The foreman stays a serial, unattended dispatcher.

### B.7 Forward-compatibility with t1454 — the selection inputs are an open extension point

The leader's constraint: card t1454 will add inputs (unified relation records, expected-file
fields and card→file edges, parent/spawned-by, size). This SPEC therefore names the inputs it
**needs** and states that the set is open (REQ-TAU-013); the keep-set and the lease path read none
of the optional inputs, so adding an input never amends them. Relation records are not bound to
one of the two existing stores; file overlap is a pluggable input whose fallback is INFERRED and
whose expected value until measured is `unmeasured`. `plan.md` § Forward-compatibility with t1454
lists what this SPEC leaves open and must not pre-empt.

## §C Requirements

Five modules. GEARS notation; every `REQ-TAU-NNN` is traced by at least one `AC-TAU-NNN` in
`acceptance.md`. Parts that no mechanical check can reach are marked *(doctrine-only)* and the
criteria say so.

### Module A — Selection authority and the single pick path

- **REQ-TAU-001** (Event-driven): **When** the operator has authorized `--auto` — by typing
  `/moai:todo --auto`, or through a lane acting on that authorization — the invoked session shall
  choose the next card on its own judgment and shall not ask the operator which card to take
  *(doctrine-only)*.
- **REQ-TAU-002** (Ubiquitous): The lease through `moai factory next` (CLI or its MCP form) shall
  be the only path by which a lane session takes a card; its bare form shall take the CLI's
  priority-order choice and its nominated form shall take the session's judged choice.
- **REQ-TAU-003** (Ubiquitous): The system shall keep queue admission with the operator — no lane
  queue mutation, no admission by `--auto`, no Jev-chosen card — and shall change no todo or gtd
  schema, relation kind, `moai graph` behavior, or queue verb, its only CLI additions being the
  `--card` flag, the MCP `card` parameter, and the lane refusal of `--auto`.

### Module B — The nominated lease

- **REQ-TAU-004** (Event-driven): **When** `--card <id>` is given to `moai factory next` (or `card`
  to the MCP `factory_next` tool), the lease path shall validate the nominee before any write,
  lease it atomically through the same version-checked edges the unnominated arms use, honor the
  quota hold and the Codex backend skip exactly as the unnominated new-card arms do, and refuse
  instead of leasing any other card.
- **REQ-TAU-005** (Event-detected): **When** a nominee is refused or its lease race is lost, the
  invocation shall exit with status 4, print one line `factory next: refused <token>: <detail>` on
  its error stream with `<token>` from the closed set of § C.2 and nothing on its output stream,
  and leave the queue and the factory record as they were before the invocation, apart from a
  change made by a concurrent holder: validation completes before any write, and a promotion of
  the nominee from `queued` to `picked` that the invocation made is undone when the claim then
  fails with no other holder (the re-selection that follows is the session's act, *doctrine-only*).
- **REQ-TAU-006** (Ubiquitous): The nominated lease shall preserve single ownership under
  concurrency: two lanes nominating different cards shall each hold their own card, and two lanes
  nominating the same card shall end with exactly one holder and one refusal.
- **REQ-TAU-007** (Ubiquitous): Without `--card`, `moai factory next` shall behave exactly as
  before — arm order, selection, output, exit codes — for every queue that holds no queued card
  whose text opens with the `[보류` marker; for a queue that does, the arm that promotes a queued
  card shall skip such a card **and count it as seen**, so a queue holding only such cards ends on
  the no-card exit (status 3), and that skip is the only default-path change (D-DEF, §B.4).
- **REQ-TAU-008** (Unwanted): **While** the session is a lane session — its lane-role marker
  equals `lane`, or its lane-label variable is non-empty, the Codex backend marker alone not
  sufficing — `moai todo --auto` shall not run: the guard shall refuse it with a dedicated text
  naming `moai factory next --card <id>` and saying the `--auto` authorization is exercised
  through it, leave the queue byte-identical, and `moai factory fallback declare` shall print the
  lease path `moai factory next [--card <id>]` in place of `/moai:todo --auto`; REQ-FLA-001's
  reference to `/moai:todo --auto` is superseded in lane sessions by the lease path.

### Module C — The keep-set

- **REQ-TAU-009** (Unwanted): The nominated lease shall not lease a card that is absent from the
  queue, in state `hold` or `dropped`, owned by another holder, whose text opens with the `[보류`
  marker, whose effective classification is `blocked`, or that is a serial card while another
  serial card holds the serial slot; each refusal shall carry its own § C.2 token.
- **REQ-TAU-010** (Event-driven): **When** the session screens a candidate it chose, it shall not
  nominate a card whose text hinges on an operator confirmation — payments, secrets, or irreversible
  external-shared work — nor a `queued` card whose pull-request or landed state it read as open or
  landed, shall report each as skipped with its reason, and shall change no queue state; for a
  `picked` card (an operator pick) the pull-request and landed state is report-only
  *(doctrine-only)*.
- **REQ-TAU-011** (Ubiquitous): The system shall express the operator-decision queue with the
  existing `hold` state and `[보류` marker only, shall introduce no new marker, state, per-item
  field, or queue verb, and shall state in its doctrine that a card marked by neither is judged by
  REQ-TAU-010 alone and that the marker demotes a card in the serial cycle and excludes it on the
  lease path.

### Module D — The decision record and the open input set

- **REQ-TAU-012** (Event-driven): **When** a lane takes a card by lease under an `--auto`
  authorization, the lane shall write one `decision record:` line carrying the three §10 fields in
  the card's progress record — as evidence, not as the §11 decision board — with
  `ladder_path=gate-row card pick (AUTONOMOUS, auto-semantics §9)` and `evidence_refs` naming at
  least the card, its class, its relation records, its pull-request/landing state, its worktree
  presence, its file overlap with in-flight lanes, and every candidate skipped with the reason;
  each input that could not be read shall appear as `unmeasured`, never omitted and never `none`.
- **REQ-TAU-013** (Ubiquitous): The selection-input set shall be open: a later card shall add an
  input without amending the keep-set or the lease path; relation records shall not be tied to one
  store; and file overlap shall be a pluggable input whose fallback to the paths changed by each
  in-flight lane's card branch is inferred and unmeasured, whose expected value until measured is
  `unmeasured`, and which the record shall not require.

### Module E — Surfaces, mirrors, budget

- **REQ-TAU-014** (Ubiquitous): The amendment shall replace — and not append to — the stated
  sentences in `kanban-dispatch.md`, `gtd.md`, `auto-semantics.md`, `manager-todo.md`, and the
  `moai-kanban-foreman` skill and add one paragraph to `kanban-dispatch-detail.md`, applying each
  edit to the live file and its template mirror in one change; carry the contract literals of
  § C.1, including in `gtd.md` the keep-set list, the record form, and the lane routing sentence
  themselves and not a pointer; confine each file's diff to the bound of AC-TAU-013; and move the
  existing doc-pin markers that assert a removed sentence in the same change: the
  `kanban-dispatch.md` prohibition sentence in `TestAutoRankDoctrineAmendment` (live and mirror)
  and the passage start marker `[HARD] **Promotion is the operator's act, always.**` in
  `TestAutoRankMirrorParity`.
- **REQ-TAU-015** (Ubiquitous): Every edited rule, skill, and agent file shall stay byte-identical
  to its template mirror, except that the one pre-existing difference in `kanban-dispatch.md`
  (line 177 of the live copy) shall be preserved exactly and reported, not absorbed.
- **REQ-TAU-016** (Ubiquitous): The edit shall not grow the always-loaded `kanban-dispatch.md` —
  live or mirror — by any byte or any character against the pinned baselines, with new detail
  placed in lazy files.

### C.1 Contract literals (the doc criteria pin these verbatim; other wording is free)

Matching is done after the shared normalization (backticks dropped, whitespace collapsed), so the
literals below are written without backticks.

| Surface (live and mirror) | Must contain | Must no longer contain |
|---|---|---|
| `kanban-dispatch.md` | `Outside an --auto authorization the leader never picks for the operator`; `authorizes the invoked session to take cards from the queue on its own judgment`; `moai factory next — bare, or --card <id>` | `Promotion is the operator's act, always.`; `The leader never picks for the operator`; `consumption of the queue and nothing else` |
| `gtd.md` (`--auto` section) | `on its own judgment`; `a lane session exercises the --auto authorization through moai factory next`; `factory next --card`; `keep-set`; `payments, secrets, or irreversible external-shared work`; `decision record:`; `ladder_path=gate-row card pick`; `unmeasured`; `demotes a card in the serial cycle and excludes it on the lease path` | `consumption of the queue and nothing else` |
| `manager-todo.md` | `on its own judgment`; `keep-set` | `process cards in queue order`; `consumption of the queue and nothing else` |
| `moai-kanban-foreman/SKILL.md` | `on its own judgment`; `outside a batch authorization` | `consumption in queue order is authorized`; `not yours to pick.` |
| `kanban-dispatch-detail.md` | `a pull request or landed state is a skip input for a queued candidate the session chose and is report-only for an operator-picked card` | — |
| `auto-semantics.md` | `ladder_path=gate-row card pick`; `adds an input without amending the keep-set or the lease path`; `unmeasured`; a `9.3` heading; `evidence, not the decision board` | — |

The `gtd.md` `--auto` section keeps its heading `### --auto — the serial batch consumption`
(the passage start marker of `TestAutoRankMirrorParity`) and every existing pinned clause; the
template mirror of `gtd.md` and of `kanban-dispatch.md` stays free of SPEC identifiers,
requirement tokens, ISO dates and 9+-hex runs (the neutrality contract two existing tests apply).

### C.2 Refusal reason tokens (normative, closed set)

| Token | Refused because |
|---|---|
| `unknown-card` | the id is in no live queue row |
| `dropped` | the card's queue state is `dropped` |
| `held` | the card's queue state is `hold` |
| `owned` | the card is `picked` and owned by another lane, or leased/assigned to another holder, or already leased by this lane |
| `hold-marker` | a `queued` card whose trimmed text opens with `[보류` |
| `blocked` | the card's effective classification is `blocked` |
| `serial-slot` | a serial card while another serial card is in flight |
| `quota-hold` | the quota gate holds new leases (arm (a) — a card already assigned to this lane — is not a new lease and stays leasable) |
| `backend-skip` | a Codex lane and a card at or past merge-ready |
| `raced` | the nominee was promoted or claimed by another lane first, or the version-checked edge reported a stale version |

A card assigned to this lane (arm (a) shape) is leasable by nomination; every other state outside
this table does not exist (the queue has four states: `queued`, `picked`, `dropped`, `hold`).

## §D Out of Scope

### Out of Scope — queue admission and card production

- No change to who may add, drop, edit, reorder, relate, hold, unhold, or finish a card; no card
  is created by `--auto`.
- No promotion path other than the lease for a lane; `moai gtd claim` and `moai gtd next <n>`
  keep their operator-only semantics.

### Out of Scope — schemas, relation kinds, graph, and the t1454 inputs

- No change to the todo or gtd queue schema, no new relation kind, no merge of the queue-findings
  store and the `gtd_relations` store, no change to `moai graph` (all card t1454's).
- Expected-file fields, card→file edges, parent/spawned-by and size fields are not added here; this
  SPEC only names the open slot (§B.7).

### Out of Scope — Jev, ranking, and the audit surfaces

- No Jev capability change; the ranking exception of SPEC-TODO-AUTO-PRIORITY-001 stays as is, and
  `moai todo --auto`'s own ranking, `selection:` record and serial cycle stay unchanged for a
  non-lane session.
- No change to the relation-blocked pickup filter (SPEC-RELATION-PICKUP-FILTER-001); that the
  default arm does not apply it is recorded (probe O2) and left for t1454's relation unification.
- No edit to `sync-auditor`, `sync-audit-4dim`, or any audit instruction: the decision record has
  no executing re-reader today (§B.5, §G) and this SPEC does not add one.

### Out of Scope — surfaces not named by the card

- No edit to the lane bootstrap notice (`internal/hook/session_start_factory_i18n.go`, four
  locales); bare `moai factory next` stays a valid instruction there. Recorded as a follow-up
  candidate (plan.md).
- No edit to a completed SPEC's body (SPEC-FACTORY-LANE-AUTONOMY-001, SPEC-TODO-AUTO-PRIORITY-001,
  SPEC-JEV-AUTO-EXCEPTION-001); no absorption of the `kanban-dispatch.md` line-177 drift.
- No board writer for decision records; no change to `moai factory decide`.

## §G Gaps and Residual Risks

- **Self-attested, unread record (unimplemented residual risk).** No Go code writes the §11 board
  and no party executes a re-read of the card-pick record: the sync-audit instructions carry no
  decision-record step. The record is evidence a reader may check; nothing checks it
  (`research.md` R4-iii, plan audit measurement).
- **Unmarked operator-decision cards.** A card kept on an operator decision queue by arrangement
  but neither `hold`ed nor `[보류`-marked is selectable until the operator marks it; t810, t1294,
  t1383 were read as selectable during the plan audit (Definition of Done item 6).
- **Text-judgement keep-set is a model act.** The text-judgement and PR-skip parts of REQ-TAU-010
  are enforceable by evidence (the record's `skipped=` entries and the card text read), not by a
  mechanical check; so are the "does not ask" part of REQ-TAU-001 and the re-selection of
  REQ-TAU-005.
- **Default-arm relation blindness persists** (probe O2) — a bare `factory next` may still lease a
  relation-blocked card; the nominated path leaves relation judgment to the session (an input), by
  design, until t1454 unifies the records.
- **Probe history is not re-executable** (research R2): its four rows motivate the criteria; the
  criteria's own RED-now cells are the re-executable ones. The lane-predicate measurements of
  R11 use a scratch binary and scratch queue (setup rows S1, S2 of `acceptance.md`).
- **File-overlap fallback is inferred.** Reading another lane's branch by name from a lane's own
  tree is plausible but unmeasured; `unmeasured` is the expected value until it is.
