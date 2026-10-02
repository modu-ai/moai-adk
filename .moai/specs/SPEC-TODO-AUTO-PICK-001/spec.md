---
id: SPEC-TODO-AUTO-PICK-001
title: "Autonomous card selection under --auto — the invoked session judges, the lease is the only pick path, the keep-set is skipped and reported"
version: "0.1.0"
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

- 0.1.0 — 2026-10-02 — plan-phase artifact set authored (card t1448; worktree
  `.moai/worktrees/t1448`, branch `WT-todo-auto-pick-autonomy`, plan-start HEAD `4bf547bca`).
  Tier M: `spec.md` + `plan.md` + `acceptance.md`, plus `research.md` (observed-versus-inferred
  evidence, probe history, alternatives) and `spec-compact.md`. Operator inputs: the card text
  (items 1-6) and the leader's added constraint that the selection inputs are an open extension
  point for card t1454 (§B.7).

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
`moai todo --auto` today (`research.md` R2/O4) — a second, queue-level pick path outside the lease.

### A.2 What this SPEC changes

It moves the **owner of the selection judgment** to the invoked session, gives that judgment a
mechanically validated way to become a lease, fixes the lease as the only pick path for a lane, and
writes down — and where possible enforces — the set of cards autonomous selection must skip. It
leaves queue **admission** (card production) with the operator.

### A.3 Verified basis (pinned tree `4bf547bca`)

The facts below were read or observed in this tree; `research.md` carries the evidence and marks
each as OBSERVED, READ, or INFERRED.

- `factory next` has three flags and no nomination (R1). Arm (c) skips `blocked` cards and
  held-slot serial cards, and **does not skip** a `[보류`-opening card (probe O1) or a
  relation-blocked card (probe O2); the `hold` state is skipped by positive state enumeration (O3).
- The concurrency core is already pinned by four existing tests (R3); the new tests assert it
  through the nominated path.
- No operator-decision-queue marker exists anywhere in code or docs (R4-ii); the queue schema is
  frozen additive-only, so none will be added (§B.3).
- No Go code writes the §11 decision board (R4-iii); the decision record is a self-attested,
  policy-layer line.
- The sentences to amend are pinned by existing tests (R6).

## §B Decisions

### B.1 Q1 — mechanism: an opt-in nomination form of `factory next`

`moai factory next --card <id>` (and the MCP `factory_next` parameter `card`) names one card. The
CLI validates the nominee and leases it through the same version-checked edges the arms use, in
one pass, or refuses with a distinct exit code and a one-line reason token so the session
re-selects. With no flag, behavior is the current priority order (with the single exception D-DEF
in B.4). Alternatives measured in `research.md` R8: a candidate-list hand-back (state store, stale
list) and a separate `peek` verb (duplicates the lane-readable `list --json`, `why`, `pr`,
`show`). The read half of "judge then pick" already exists; only the validated write half was
missing.

### B.2 Q4 — boundaries that stay true

- Queue **admission** stays the operator's: `moai gtd add/drop/edit/done/relate` stay refused to a
  lane (`todoLaneReadOnlyVerbs`), `moai contract sign` stays forbidden, and the leader remains the
  queue's sole producer.
- The self-dispatch lane exception stays and becomes the **sole** lane pick path: `moai factory
  next` is the lane's only promotion (§C module A, B).
- Jev stays display-only. Nothing in this SPEC lets a Jev answer choose a card or write the queue.
- The `--auto` invocation remains the operator's batch approval; a lane's self-service pickup is
  that approval acted on, never a self-grant.
- **Reconciliation with the pre-dispatch PR cross-check.** The cross-check "reports, never
  vetoes" applies to a card the **operator** already picked: the leader reports PR/landed state
  and the operator confirms or withdraws. An **autonomously chosen** card has no operator pick to
  override, so there the cross-check is a selection *input* (`pr=` in the record): a candidate with
  an open pull request or a landed fix is not nominated and is reported as skipped. That is the
  selection judgment doing its job, not a veto of an operator act. The two clauses govern
  different cards and do not conflict; the stub's `confirms or withdraws` literal stays untouched.

### B.3 Q2 — the keep-set marker: reuse, add nothing

The queue record's per-item contract is frozen additive-only (R4-ii), and no operator-decision
marker exists. **No new marker, state, or field is introduced.** An operator-decision-queue card is
expressed with what exists: the structural `hold` state (written only by `moai gtd hold`, "operator
(or lead) only: no lane session or machine leaser can hold or unhold") and, for prose parking, the
`[보류` text marker. Both are mechanically enforced on the lease path (REQ-TAU-009, -007). A card
the operator means to keep out but has neither held nor marked is **not** mechanically protected; it
falls to the text-judgement half of the keep-set, whose evidence is recorded (REQ-TAU-010). That
failure mode is stated, not hidden (§G).

### B.4 The two deliberate deltas, each isolated so the leader can veto it

- **D-DEF — the default arm skips `[보류`-marked cards (second clause of REQ-TAU-007).** Today the
  bare `factory next` arm (c) leases a card whose text opens with `[보류` (probe O1), so the
  keep-set guard would be bypassed by omitting the flag. REQ-TAU-007 states the unchanged default
  for every queue that holds no marker-bearing queued card and names the skip as the single
  exception. The skip has its own criterion subtest (AC-TAU-004), so dropping the clause and that
  subtest leaves the rest intact.
- **D-LANE — a lane session is refused `moai todo --auto` (REQ-TAU-008).** A lane can run the
  serial cycle today (probe O4), picking in the queue outside the lease. The refusal reuses the
  existing lane-boundary refusal text, which already says "a lane takes its next card through
  `moai factory next`". SPEC-FACTORY-LANE-AUTONOMY-001's "self-service pickup" is read as the lease
  path (research R5); that completed SPEC's body is not edited.

### B.5 Q3 — the decision record

Reuse the §10 one-line form (`decision record: decided_by= evidence_refs= ladder_path=`). The
card-pick row's `ladder_path` token is `gate-row card pick (AUTONOMOUS, auto-semantics §9)`,
following the precedent of the plan→run Kickoff row. `evidence_refs` carries `key=value` items
joined by `;`: `card`, `class`, `relate`, `pr`, `wt`, `overlap`, `skipped`, each `unmeasured` when
the input could not be read. The record is self-attested (no board writer exists, R4-iii); the
sync-audit re-read is the compensating control (§10). The record's home in this tree is the card's
progress record; this SPEC does not claim a board write.

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
one of the two existing stores; file overlap is a pluggable input with a defined fallback.
`plan.md` § Forward-compatibility with t1454 lists what this SPEC leaves open and must not
pre-empt.

## §C Requirements

Five modules. GEARS notation; every `REQ-TAU-NNN` is traced by at least one `AC-TAU-NNN` in
`acceptance.md`.

### Module A — Selection authority and the single pick path

- **REQ-TAU-001** (Event-driven): **When** the operator has authorized `--auto` — by typing
  `/moai:todo --auto`, or through a lane's self-service pickup — the invoked session shall choose the
  next card on its own judgment and shall not ask the operator which card to take.
- **REQ-TAU-002** (Ubiquitous): The lease through `moai factory next` (CLI or its MCP form) shall
  be the only path by which a lane session takes a card; its bare form shall take the CLI's
  priority-order choice and its nominated form shall take the session's judged choice.
- **REQ-TAU-003** (Ubiquitous): The system shall keep queue admission with the operator and its
  schemas untouched: no lane session shall run `add`, `drop`, `edit`, `done`, `relate` or any other
  queue mutation, no `--auto` authorization shall admit a card, Jev output shall never choose a
  card or write the queue, the pre-dispatch cross-check shall keep reporting without vetoing a
  card the operator picked while serving as a selection input for a card the session chose itself,
  and no todo or gtd schema, relation kind, `moai graph` behavior, or queue verb shall change —
  the only CLI additions being the `--card` flag, the MCP `card` parameter, and the lane refusal
  of `--auto`.

### Module B — The nominated lease

- **REQ-TAU-004** (Where): **Where** `--card <id>` is given to `moai factory next` (or `card` to
  the MCP `factory_next` tool), the lease path shall validate the nominee and lease it atomically
  through the same version-checked edges the unnominated arms use, shall honor the quota hold and
  the Codex backend skip exactly as the unnominated new-card arms do, and shall refuse instead of
  leasing any other card.
- **REQ-TAU-005** (Event-detected): **When** a nominee is refused or its lease race is lost, the
  lease path shall exit with a status distinct from success and from the no-card status, print one
  line naming a closed reason token on its error stream, change no queue or factory-record
  state, and the session shall re-select a different candidate.
- **REQ-TAU-006** (Ubiquitous): The nominated lease shall preserve single ownership under
  concurrency: two lanes nominating different cards shall each hold their own card, and two lanes
  nominating the same card shall end with exactly one holder and one refusal.
- **REQ-TAU-007** (Ubiquitous): Without `--card`, `moai factory next` shall behave exactly as
  before — arm order, selection, output, exit codes — for every queue that holds no queued card
  whose text opens with the `[보류` marker; for a queue that does, the arm that promotes a queued
  card shall skip such a card, and that skip is the only default-path change (D-DEF, §B.4).
- **REQ-TAU-008** (Unwanted): **While** the lane-refusal predicate holds, `moai todo --auto` shall
  not run: the todo guard shall refuse it with the lane-boundary refusal naming `moai factory next`
  and shall leave the queue byte-identical.

### Module C — The keep-set

- **REQ-TAU-009** (Unwanted): The nominated lease shall not lease a card that is not in the queue,
  in state `hold`, `dropped`, or owned by another holder, a queued card whose text opens with the
  `[보류` marker, a card whose effective classification is `blocked`, or a serial card while another
  serial card holds the serial slot; each refusal shall carry its own reason token.
- **REQ-TAU-010** (Event-driven): **When** the session reads a candidate's card text and the body
  hinges on an operator confirmation — payments, secrets, or irreversible external-shared work —
  the session shall not nominate that card, shall report it as skipped with its reason, and shall
  change no queue state.
- **REQ-TAU-011** (Ubiquitous): The system shall express the operator-decision queue with the
  existing `hold` state and `[보류` marker only, shall introduce no new marker, state, per-item
  field, or queue verb, and shall state in its doctrine that a card marked by neither is judged by
  REQ-TAU-010 alone.

### Module D — The decision record and the open input set

- **REQ-TAU-012** (Event-driven): **When** a lane takes a card by lease under an `--auto`
  authorization, the lane shall write one `decision record:` line carrying the three §10 fields,
  with `ladder_path=gate-row card pick (AUTONOMOUS, auto-semantics §9)` and `evidence_refs`
  naming at least the card, its class, its relation records, its pull-request/landing state, its
  worktree presence, its file overlap with in-flight lanes, and every candidate skipped with the
  reason; each input that could not be read shall appear as `unmeasured`, never omitted and never
  `none`.
- **REQ-TAU-013** (Ubiquitous): The selection-input set shall be open: a later card shall add an
  input without amending the keep-set or the lease path; relation records shall not be tied to one
  store; and file overlap shall be a pluggable input that, when expected-file data is absent, falls
  back to the paths changed by each in-flight lane's card branch and, when that is unavailable
  too, reads `unmeasured`.

### Module E — Surfaces, mirrors, budget

- **REQ-TAU-014** (Ubiquitous): The amendment shall replace — and not append to — the stated
  sentences in `kanban-dispatch.md`, `gtd.md`, `auto-semantics.md`, `manager-todo.md`, and the
  `moai-kanban-foreman` skill, carry the contract literals of § C.1, and move the existing doc-pin
  tests to the new wording in the same milestone, so no pin asserts a sentence this SPEC removes.
- **REQ-TAU-015** (Ubiquitous): Every edited rule, skill, and agent file shall stay byte-identical
  to its template mirror, except that the one pre-existing difference in `kanban-dispatch.md`
  (line 177 of the live copy) shall be preserved exactly and reported, not absorbed.
- **REQ-TAU-016** (Ubiquitous): The edit shall not grow the always-loaded `kanban-dispatch.md` —
  live or mirror — by any byte against the pinned baseline, with new detail placed in lazy files.

### C.1 Contract literals (the doc criteria pin these verbatim; other wording is free)

| Surface (live and mirror) | Must contain | Must no longer contain |
|---|---|---|
| `kanban-dispatch.md` | `Outside an` + backtick-`--auto`-backtick + ` authorization the leader never picks for the operator`; `authorizes the invoked session to take cards from the queue on its own judgment`; `moai factory next` + ` — bare, or ` + backtick-`--card <id>`-backtick | `Promotion is the operator's act, always.`; `The leader never picks for the operator`; `consumption of the queue and nothing else` |
| `gtd.md` | `on its own judgment`; `factory next --card`; `keep-set`; `decision record` (in the `--auto` section) | `consumption of the queue and nothing else` |
| `manager-todo.md` | `on its own judgment`; `keep-set` | `process cards in queue order`; `consumption of the queue and nothing else` |
| `moai-kanban-foreman/SKILL.md` | `on its own judgment`; `outside a batch authorization` | `consumption in queue order is authorized`; `not yours to pick.` |
| `auto-semantics.md` | `ladder_path=gate-row card pick`; `adds an input without amending the keep-set or the lease path`; `unmeasured`; a `9.3` heading | — |

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

### Out of Scope — Jev and ranking

- No Jev capability change; the ranking exception of SPEC-TODO-AUTO-PRIORITY-001 stays as is, and
  `moai todo --auto`'s own ranking, `selection:` record and serial cycle stay unchanged.
- No change to the relation-blocked pickup filter (SPEC-RELATION-PICKUP-FILTER-001); that the
  default arm does not apply it is recorded (probe O2) and left for t1454's relation unification.

### Out of Scope — surfaces not named by the card

- No edit to the lane bootstrap notice (`internal/hook/session_start_factory_i18n.go`, four
  locales); bare `moai factory next` stays a valid instruction there. Recorded as a follow-up
  candidate (plan.md).
- No edit to a completed SPEC's body (SPEC-FACTORY-LANE-AUTONOMY-001, SPEC-TODO-AUTO-PRIORITY-001,
  SPEC-JEV-AUTO-EXCEPTION-001); no absorption of the `kanban-dispatch.md` line-177 drift.
- No board writer for decision records; no change to `moai factory decide`.

## §G Gaps and Residual Risks

- **Self-attested record.** No Go code writes the §11 board; the card-pick record is a line the
  lane writes and the sync audit re-reads — detection, not prevention (research R4-iii).
- **Unmarked operator-decision cards.** A card kept on an operator decision queue by arrangement
  but neither `hold`ed nor `[보류`-marked is selectable until the operator marks it (plan.md
  § Operational follow-ups).
- **Text-judgement keep-set is a model act.** REQ-TAU-010 is enforceable by evidence (the record's
  `skipped=` entries and the card text it read), not by a mechanical check.
- **Default-arm relation blindness persists** (probe O2) — a bare `factory next` may still lease a
  relation-blocked card; the nominated path leaves relation judgment to the session (an input), by
  design, until t1454 unifies the records.
- **Probe history is not re-executable** (research R2): its four rows motivate the criteria; the
  criteria's own RED-now cells are the re-executable ones.
