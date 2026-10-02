---
id: SPEC-TODO-AUTO-PICK-001
title: "Autonomous card selection under --auto — the invoked session judges, the lease is the only pick path, the keep-set is skipped and reported"
version: "0.4.1"
status: in-progress
created: 2026-10-02
updated: 2026-10-03
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

- 0.4.1 — 2026-10-03 — post-audit mechanical re-pin of measurement baselines after develop `7109e0900`
  was absorbed (merge `095ac6c3e`) **following** the iteration-4 delta plan-audit PASS (audited hash
  `724841520`, base `4bf547bca`). REQ-TAU-016 and AC-TAU-011 now measure the always-loaded
  `kanban-dispatch.md` against the blob at the card's merge base with develop (re-derived at reading
  time, never pinned) instead of fixed figures that develop's own +1,349 B per copy made unsatisfiable;
  the criterion cells L14 and L21 are re-based the same way; line numbers moved by develop's
  insertions are given with their text (the live-only `moai worktree sweep …` sentence is line 181, was
  177). **No requirement or criterion changed meaning** — requirement count 16, criterion count 14,
  `status:` untouched; the audit resolution map entry is `plan.md` § 10 "Post-audit re-pin". The plan
  artifacts differ from the audited hash only by this re-pin (and the run phase's sanctioned `status:`
  transition); **no re-audit was run**.
- 0.4.0 — 2026-10-03 — plan-audit iteration 3 finding F1 and notes S1-S8 corrected. The iteration-3
  audit (extension; audited commit `625f01718`) returned **FAIL, score 0.80, all nine must-pass
  criteria PASS, one blocking finding (F1)**; this version corrects F1 and the notes **without a
  re-audit** — the leader decides the next step, and nothing here claims a PASS. F1: the template
  edits change the stored hashes in `internal/template/catalog.yaml` (the `moai` and
  `moai-kanban-foreman` skill directories and the template `manager-todo.md`), so the plan now
  regenerates the catalog with the repository's generator in the same M5 commit and carries the two
  catalog guard tests through M5, M6, AC-TAU-010, -012, -013 and Definition of Done 7; the red was
  observed on a reversible perturbation of the tree, not inferred. F2: the user-facing pages that
  still say the pick is always the operator are named and assigned to the sync phase (plan §4a, §7;
  Out of Scope here). S1-S8: wording and bookkeeping corrections (`plan.md` § Audit resolution map).
  Requirement count 16, criterion count 14 — unchanged.
- 0.3.0 — 2026-10-02 — plan-audit iteration 2 repair (FAIL 0.77, MP-9 failed on N1; audited commit
  `63daaf6a7`; the leader approved one extra delta audit scoped to N1-N6 plus a full ordering
  re-read). N1: M1 declares the one seam variable, so the seam tests compile at M1 and are RED at
  runtime while the golden is GREEN. N2: REQ-TAU-005 and AC-TAU-014 are narrowed to what the
  factory-record API can undo (it has no delete), with the seam point, the post-record residue and
  the compensation failure specified (§B.8). N3: the generated Codex `manager-todo.toml` and the
  second stale sentence in `auto-semantics.md` join the sweep; a whole-tree sweep for the old
  authority wording is classified in `plan.md`. N4: every release-blocking test name has its own
  RED-now cell. N5: the marker predicate's mid-text and leading-whitespace cards are fixtures. N6:
  the refusal tokens are closed over the producible refusals (twelve) and `owned`/`recorded` are
  defined against the nineteen record states. N7: §G states that the §10 sync-audit wording stands
  for the other gates. The handoff wording for the three operator-decision cards is made identical
  everywhere (§B.3, REQ-TAU-011, §G, Definition of Done 6, `plan.md` §7). REQ-TAU-013, the open
  input set for card t1454, is unchanged. Requirement count 16, criterion count 14. Resolution map:
  `plan.md` § Audit resolution map.
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
**How the three cards the memory index calls operator-decision-queue cards are identified.** Read
through the installed binary with the read-only `moai gtd show` (plan audit, and again in the
iteration-2 repair): t810 is `picked`, t1294 and t1383 are `queued`, none with text that begins
with the marker. The handoff sentence, identical in every place it appears: **Today t810
(`picked`), t1294 and t1383 (`queued`, ordinary text) carry neither the structural `hold` state
nor a leading `[보류` marker, so the keep-set does not mechanically identify them. The SPEC
supports exactly two identification forms — (a) the structural `hold` state, written only by
`moai gtd hold`, and (b) card text that begins with the `[보류` marker — and a lane may write
neither. The operator or leader must apply one of them to each such card before lanes exercise
this doctrine.** This SPEC does not touch the queue.

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
(`.moai/specs/<SPEC-ID>/progress.md` in a fresh top-level section — `§J`, because the schema's
section map allocates `§F` to the Phase 4 Mode Selection log and tells a new concern to claim an
unallocated letter — where a SPEC exists, else under `.moai/reports/<card-id>/`).
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

### B.8 The nominee's state semantics — what is undone and what is not

The factory record API cannot delete: `RecordPicked` INSERTs a `cards` row at `picked` and appends a
`card.transition` event (`homestate/card_picked.go:114-145`), and the package has no delete verb
(`grep` for delete verbs in non-test `internal/homestate` finds none; plan audit N2). The
compensation therefore promises only what the stores allow:

- **Before any write**, every § C.2 token that can be read is decided (one pure queue read, one
  record read, the foreign-worktree precheck). A refusal here changes nothing.
- **Seam point — before `RecordPicked`.** Between the queue promotion and the first record write
  the lease path calls one test seam (`factoryNominateBeforeRecord`, declared in M1). A seam that
  injects a refusal, or a competing lane that leases the card there, exercises the two
  compensation outcomes the tests can observe: with no record row and no other holder the queue
  item is restored to `queued`; with another holder the winner's state is left alone and the
  invocation refuses `raced`.
- **After `RecordPicked` the record cannot be rolled back.** A failure there (a later
  version-checked edge refused for a non-race reason, or a store error) leaves the card's row at
  `picked`, unowned, with its `card.transition` event, and the queue item restored to `queued`
  only if the record shows **no row, or a row at `picked` with no owner**. That residue differs from
  what arm (c)'s own failed claim leaves: arm (c) leaves the queue item `picked` and the row
  `picked`, while a failed nominated claim leaves the queue item `queued` with a `picked` row. Arm
  (b) skips it (the item is `queued`); it **re-adopts through arm (c)**, or through a later
  nomination (`RecordPicked` returns the existing row unchanged); it is visible as a `picked` row
  with no owner in `moai factory status`. A row `assigned` to this lane (a first claim edge landed
  and a later edge failed for a non-race reason) meets neither restore alternative and is left as
  it is — it is this lane's own arm (a) lease target. None of this is **tested** — the single M1
  seam sits before `RecordPicked` — and it is listed in §G.
- **Non-atomicity.** The record read ("no other holder") and the queue write (the restore) are two
  stores with no shared lock. A lane that adopts the unowned `picked` row between the two leaves
  the queue `queued` while the record says `leased`; no second lease results (arm (c) re-promoting
  the item cannot claim a row that is no longer `picked`), and the inconsistency is the one a
  crashed lane already leaves. It is accepted and stated, not hidden.
- **A failed compensation** (the restoring queue write errors, or finds the item no longer
  `picked`) is not a refusal token. If the item is no longer `picked` the compensation does
  nothing and the original token is reported; if the write errors the invocation exits with status
  1 and `factory next: compensation failed: <cause>` naming the card, and the card stays `picked`
  and unowned (adoptable, nothing lost).

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
  and shall write nothing once the refusal is decided: every readable refusal is decided before the
  first write; a promotion of the nominee from `queued` to `picked` that the invocation made is
  undone, in one queue write that acts only if the item is still `picked`, when the claim fails
  and the record shows no row, or a row at `picked` with no owner; a row `RecordPicked`
  already wrote cannot be undone and stays at `picked`, unowned (§B.8); and a failed compensation
  exits with status 1 and `factory next: compensation failed: <cause>`, not with a token (the
  re-selection that follows a refusal is the session's act, *doctrine-only*).
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
  queue, in state `hold` or `dropped`, whose factory-record row is in a state § C.2 maps to `owned`
  or `recorded`, that is `queued` with text whose trimmed form opens with the `[보류` marker,
  whose effective classification is `blocked`, whose landing directory belongs to no card, or that
  is a serial card while another serial card holds the serial slot; each refusal shall carry its
  own § C.2 token.
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
  lease path, and shall carry the handoff sentence **defined in §B.3, which this requirement does
  not restate** (it names the three operator-decision cards t810, t1294 and t1383, says none carries
  either identification form today, and says the operator or leader must apply one before lanes
  exercise the doctrine); §B.3, §G, Definition of Done 6 and `plan.md` §7 carry that sentence in
  identical words.

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
  (the live-copy sentence that carries `moai worktree sweep …`, line 177 at plan start and line 181
  after the develop absorption — located by its text) shall be preserved exactly and reported, not
  absorbed; and the two
  artifacts generated from the edited template files — the Codex agent artifact
  `manager-todo.toml` and the stored hashes in `internal/template/catalog.yaml` — shall be
  regenerated by their generators in the same change, never hand-edited.
- **REQ-TAU-016** (Ubiquitous): The edit shall not grow the always-loaded `kanban-dispatch.md` —
  live or mirror — by any byte or any character against the blob of that file at the card's merge
  base with develop (re-derived at reading time, never pinned), with new detail placed in lazy
  files.

### C.1 Contract literals (the doc criteria pin these verbatim; other wording is free)

Matching is done after the shared normalization (backticks dropped, whitespace collapsed), so the
literals below are written without backticks.

| Surface (live and mirror) | Must contain | Must no longer contain |
|---|---|---|
| `kanban-dispatch.md` | `Outside an --auto authorization the leader never picks for the operator`; `authorizes the invoked session to take cards from the queue on its own judgment`; `moai factory next — bare, or --card <id>` | `Promotion is the operator's act, always.`; `The leader never picks for the operator`; `consumption of the queue and nothing else` |
| `gtd.md` (`--auto` section) | `on its own judgment`; `a lane session exercises the --auto authorization through moai factory next`; `factory next --card`; `keep-set`; `payments, secrets, or irreversible external-shared work`; `decision record:`; `ladder_path=gate-row card pick`; `unmeasured`; `demotes a card in the serial cycle and excludes it on the lease path` | `consumption of the queue and nothing else` |
| `manager-todo.md` | `on its own judgment`; `keep-set` | `process cards in queue order`; `consumption of the queue and nothing else` |
| `.codex/agents/moai/manager-todo.toml` (template tree only; **generated** from the template `manager-todo.md` by `make agents-emit`, never hand-edited) | the same two literals as `manager-todo.md` | the same two absent literals |
| `moai-kanban-foreman/SKILL.md` | `on its own judgment`; `outside a batch authorization` | `consumption in queue order is authorized`; `not yours to pick.` |
| `kanban-dispatch-detail.md` | `a pull request or landed state is a skip input for a queued candidate the session chose and is report-only for an operator-picked card` | — |
| `auto-semantics.md` | `ladder_path=gate-row card pick`; `adds an input without amending the keep-set or the lease path`; `unmeasured`; a `9.3` heading; `evidence, not the decision board`; `no party re-reads the card-pick record` | `authorizes serial queue consumption and nothing else` (the §9.2 sentence at line 186, the second place the old authority is stated) |

Literals that a raw `git grep -F` cell checks (L9, L19, L20, L24-L26 of the ledger) are written
**without interior backticks**, so the raw-byte cell and the normalized Go test agree. The §9.1
and §10 sentences of `auto-semantics.md` that call the sync audit's re-read "the compensating
control" are **not edited**: they stand for the other gates (§G); the new §9.3 sentence above
says they do not hold for the card-pick record.

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
| `owned` | the card's factory-record row is `assigned` to another lane, or is in an in-flight state — `leased`, `plan`, `plan-audit`, `kickoff`, `run`, `sync`, `sync-audit`, `merge-ready`, `merging`, `merged-local`, `pushed`, `ci-green` — for **any** holder, this lane included |
| `recorded` | the card's factory-record row is in a terminal or parked state — `done`, `abandoned`, `failed`, `blocked`, `needs-decision` — or is a legacy row outside the nineteen states; only an operator unblock or re-pick moves it |
| `foreign-worktree` | the card has no recorded worktree and its landing directory already exists and belongs to no card (the `factoryRefuseForeignWorktree` refusal, decided by the read-only precheck) |
| `hold-marker` | a `queued` card whose trimmed text opens with `[보류` |
| `blocked` | the card's effective classification is `blocked` |
| `serial-slot` | a serial card while another serial card is in flight |
| `quota-hold` | the quota gate holds new leases (arm (a) — a card already assigned to this lane — is not a new lease and stays leasable) |
| `backend-skip` | a Codex lane and a card at or past merge-ready |
| `raced` | the nominee was promoted or claimed by another lane first, or the version-checked edge reported a stale version |

**Leasable by nomination:** a card with no record row; a card whose row is `picked` with no owner
(an operator pick, arms (b)/(b2)); a card whose row is `assigned` to **this** lane (arm (a)); and a
`queued` card that passes every row above. The factory record has nineteen states — `picked`,
`assigned`, the twelve in-flight states of the `owned` row, and the five of the `recorded` row — and
every one is mapped here, so no record state produces a refusal without a token. The queue has four
states: `queued`, `picked`, `dropped`, `hold`. An error reading or opening either store (not a
refusal of the card) exits with status 1 and no token, as every other `factory next` error does;
the closed set covers every refusal of the nominee, not infrastructure failures. A failed
compensation is the other non-token exit (REQ-TAU-005, §B.8). The claim-lease columns of the queue
(`picked_by`, `lease_expires_at`) are not consulted, exactly as the unnominated arms do not.

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
  SPEC-JEV-AUTO-EXCEPTION-001); no absorption of the `kanban-dispatch.md` live-only-sentence drift (line 177 at plan start, line 181
  after the develop absorption).
- No board writer for decision records; no change to `moai factory decide`.
- No run-phase edit to the user-facing pages that state the operator-only pick — `README.md` and the
  four-locale docs-site pages `advanced/factory-mode.md` and `advanced/kanban-mode.md`. They are the
  **sync phase's** scope (`plan.md` §4a, §7), where the sentence "the actor that picks a card is
  always the operator" becomes "the operator, in person or in advance through `--auto`".

## §G Gaps and Residual Risks

- **Self-attested, unread record (unimplemented residual risk).** No Go code writes the §11 board
  and no party executes a re-read of the card-pick record: the sync-audit instructions carry no
  decision-record step. The record is evidence a reader may check; nothing checks it
  (`research.md` R4-iii, plan audit measurement).
- **Unmarked operator-decision cards.** Today t810 (`picked`), t1294 and t1383 (`queued`, ordinary
  text) carry neither the structural `hold` state nor a leading `[보류` marker, so the keep-set
  does not mechanically identify them. The SPEC supports exactly two identification forms — (a) the
  structural `hold` state, written only by `moai gtd hold`, and (b) card text that begins with the
  `[보류` marker — and a lane may write neither. The operator or leader must apply one of them to
  each such card before lanes exercise this doctrine (Definition of Done item 6).
- **The post-record residue of a failed claim is untested.** After `RecordPicked` the factory record
  cannot be rolled back (§B.8); the single M1 seam sits before it, so the case where a later claim
  edge fails and leaves a `picked`, unowned row plus its event is specified, accepted and **not
  tested**, and so is the branch where the compensating queue write itself errors (exit 1,
  `factory next: compensation failed`). The record read and the queue write of the compensation are
  two stores without a shared lock (the window is stated in §B.8).
- **The stranded `picked` row holds the serial slot (accepted, visible).** The residue of a failed
  claim after `RecordPicked` is a `picked`, unowned row, and `factorySerialSlotFree("picked")` is
  false (`factory_card.go` ~L202-L210): for a serial card that row holds the serial slot against
  other serial cards until it is re-adopted (`factory_card.go:202-210` read; the arm-(c)
  re-adoption order was read by the iteration-3 audit and not re-measured here). The state shows
  in `moai factory status` as a `picked` row with no owner.
- **auto-semantics §9.1 and §10 are outside this SPEC's scope.** Their wording that the sync audit's
  re-read is "the compensating control" for decision records (`auto-semantics.md` ~L179-L180 and
  ~L241-L242) still stands for the other gates; this SPEC retracts that claim **only for the
  card-pick record**, in the new §9.3 sentence, and does not edit those lines.
- **The `--auto` flag help still says the invocation is the operator's batch approval "of the queue
  and nothing else" and that the cycle processes the queue serially** (`todo.go` ~L315, read-only
  in this SPEC by REQ-TAU-003): accurate for the operator-session serial cycle, silent about the
  lane path; named here, not edited.
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
