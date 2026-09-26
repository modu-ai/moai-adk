---
id: SPEC-FACTORY-RECORD-001
title: "Harness-neutral factory F1 — card record layer and version-checked card state machine"
version: "0.1.0"
status: draft
created: 2026-09-26
updated: 2026-09-26
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/homestate, internal/cli, internal/kanban"
lifecycle: spec-anchored
tags: "factory, record-layer, state-machine, lease, heartbeat, evidence-gate, optimistic-concurrency, migration, factory-f1, card-t1239"
tier: L
card: t1239
related_specs: [SPEC-FACTORY-RUN-RETIRE-001, SPEC-FACTORY-MODE-001, SPEC-FACTORY-WORKER-NAMING-001, SPEC-AUTONOMY-CONTRACT-001, SPEC-GTD-AUTONOMY-001]
---

# SPEC-FACTORY-RECORD-001 — Factory card record layer and state machine (F1)

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-09-26 | manager-spec | Initial plan-phase draft (card t1239, FACTORY-F1). Baseline: worktree `.claude/worktrees/t1239`, branch `WT-factory-record-state`, base develop `553e224f3`. Operator decisions of 2026-09-26 recorded in §C. |

## §A. Background and Motivation

The operator decided on 2026-09-26 to rebuild Factory Mode as a harness-neutral pipeline in which the
**record is the source of truth**: the queue holds the cards the operator picked, a factory database
holds each card's position in the pipeline, and `.moai/reports/<card>/` holds the evidence. Every
movement of a card is a version-checked atomic update of that record, decided by Go from files and
commit SHAs it reads itself — never from an agent's claim that a phase passed.

The pieces exist but are not connected (measured in `research.md`):

- `factory.db` already has a `cards` table with `version` and `evidence_path` columns
  (`internal/homestate/factory.go:44-52`) and a `workers` table with `heartbeat_at`
  (`factory.go:24-32`). The only writer of `cards`, `RecordCard` (`internal/homestate/runtime.go:53-84`),
  upserts unconditionally and never compares the version, and it has **no production caller**.
  Nothing ever updates `workers.heartbeat_at` after registration.
- Today's per-card factory state is written to a different store — the queue's
  `todo_runtime_assignments` table in `backlog.db` (`internal/kanban/factory_runtime.go:58-71`,
  `internal/kanban/todo_runtime.go:53-57`) — which that code itself declares "the latest execution
  report, not card completion authority" (`todo_runtime.go:14`).
- `moai factory` has no `status`, `assign`, or `decide` subcommand; an unknown subcommand prints help
  and exits 0 (measured, `research.md` § R9).

F1 is the first of three cards. It delivers the record layer only: the schema, the state machine, the
transition API with its evidence gates, the lease, and three operator commands. F2 (self-dispatch
lane verbs and launcher) and F3 (controller and Decider) consume it.

## §B. Requirements (GEARS)

### Record schema and migration

- **REQ-FR-001** (Ubiquitous): The factory card record shall carry, per card within a run, its
  state, the last working stage it reached, a version counter, the owner label, the lease holder and
  lease expiry, the last heartbeat time, the pending decision (gate, question, state to resume, and
  once decided the decider and decision time), the assignment hints `prefer` and `after`, the SPEC
  identifier, the card worktree path, the last accepted evidence path and commit SHA, the local merge
  commit SHA, and an optional contract reference consisting of a SPEC identifier, a contract digest,
  and a signing time.
- **REQ-FR-002** (Event-driven): **When** a factory database at schema version 3 is opened, the
  factory store shall migrate it to the F1 schema version in one transaction that adds only the
  missing columns, preserves every existing row of every table, and records the new schema version.
- **REQ-FR-003** (State-driven): **While** a card row holds a state value outside the F1 state set
  (a row written before the migration), the transition API shall refuse every transition from that
  row except an operator decision to `abandoned`.

### State machine and atomic transitions

- **REQ-FR-004** (Ubiquitous): The card state machine shall admit exactly the states `picked`,
  `assigned`, `leased`, `plan`, `plan-audit`, `kickoff`, `run`, `sync`, `sync-audit`, `merge-ready`,
  `merging`, `merged-local`, `pushed`, `ci-green`, `done`, `needs-decision`, `blocked`, `failed`, and
  `abandoned`, and exactly the edges enumerated in `design.md` § Transition Table.
- **REQ-FR-005** (Ubiquitous): The transition API shall apply each transition as one transaction that
  compares the caller's expected version with the stored version, writes the new state, increments the
  version by one, and appends one `card.transition` event naming the card, the source and target
  states, the new version, the actor, and the evidence read — committing all of these or none.
- **REQ-FR-006** (Event-driven): **When** a transition's expected version differs from the stored
  version, the transition API shall refuse it with a stale-version error and leave the card row and the
  event log unchanged.
- **REQ-FR-007** (Event-driven): **When** a requested transition is not an edge of the transition
  table, the transition API shall refuse it with an illegal-transition error and leave the card row
  and the event log unchanged.

### Evidence gates

- **REQ-FR-008** (Event-driven): **When** a card enters `plan-audit` or `sync-audit`, the transition
  API shall accept the transition only after it has itself verified that the supplied commit SHA
  resolves in the card's worktree, is an ancestor of or equal to that worktree's HEAD, and contains
  the supplied artifact path at that commit; it shall record that SHA and path as the card's audited
  evidence.
- **REQ-FR-009** (Event-driven): **When** a card leaves `plan-audit` for `kickoff` or leaves
  `sync-audit` for `merge-ready`, the transition API shall read the card's audit verdict file under
  `.moai/reports/<card-id>/` and accept the transition only when the file's machine-readable verdict is
  `PASS` or `PASS-WITH-DEBT` and its machine-readable audited SHA equals the SHA recorded at audit
  entry; a verdict file that names the recorded SHA with any verdict shall admit the return edge to
  `plan` or `sync`.
- **REQ-FR-010** (Event-driven): **When** a card enters `merged-local`, the transition API shall
  accept the transition only after verifying that the supplied merge commit is reachable from the local
  integration branch, that the merge commit's tree equals its second parent's tree, and that the
  supplied re-measure evidence file exists and names that merge commit; it shall record the merge
  commit SHA, the tree hash, and the re-measure evidence path.
- **REQ-FR-011** (Unwanted): The transition API shall not accept a caller-supplied verdict, pass flag,
  tree hash, or ancestry claim as a substitute for the verification it performs itself.

### Lease and heartbeat

- **REQ-FR-012** (Event-driven): **When** a worker acquires the lease on an `assigned` card, the
  transition API shall accept only a worker label registered in the factory worker roster and equal to
  the card's owner label, move the card to `leased`, and record the holder, the heartbeat time, and an
  expiry equal to the heartbeat time plus the lease duration.
- **REQ-FR-013** (Event-driven): **When** the lease holder renews its heartbeat, the record shall
  extend the lease expiry and update the worker roster's heartbeat time; **when** any other label
  attempts the renewal, the record shall refuse it unchanged.
- **REQ-FR-014** (State-driven): **While** a card's lease has expired and its state is a
  lease-holding state other than `merging`, the transition API shall, before any other transition on
  that card, return the card to `assigned` with the holder cleared and the stage, worktree path, and
  evidence fields preserved, append a `lease.expired` event, and modify no file or git ref in the card's
  worktree.
- **REQ-FR-015** (State-driven): **While** a card's lease has expired and its state is `merging`, the
  transition API shall move the card to `blocked` rather than `assigned`, and append a `lease.expired`
  event naming the interrupted merge.

### Assignment hints

- **REQ-FR-016** (Where): **Where** a card carries an `after` hint naming a predecessor card, the
  transition from `picked` to `assigned` shall be refused until the predecessor has a factory record in
  `merged-local`, `pushed`, `ci-green`, or `done`; a predecessor with no factory record shall refuse the
  transition with an unknown-predecessor error.
- **REQ-FR-017** (Where): **Where** a card carries a `prefer` hint (for example `backend=codex`), the
  record shall store it verbatim and report it, and F1 shall not use it to refuse or reorder any
  transition.

### Decisions and gates

- **REQ-FR-018** (Event-driven): **When** an operator runs `moai factory decide` on the `kickoff` gate
  with the choice `approve` or `reject`, the record shall move each named card in `kickoff` to `run` or
  to `blocked` respectively, recording the decider as `human` and the decision time; one invocation
  naming several cards shall apply one independent version-checked transition per card, and a decider
  value other than `human` shall be refused.
- **REQ-FR-019** (Event-driven): **When** a card enters `needs-decision`, the record shall store the
  question and the state to resume and clear the lease; **when** the operator decides `resume`, `block`,
  or `abandon`, the record shall move the card to the resume state (or to `assigned` with that stage
  preserved when the resume state is lease-holding), to `blocked`, or to `abandoned`.
- **REQ-FR-020** (Event-driven): **When** the lead runs `moai factory decide` on the `push` gate for
  cards in `merged-local`, the record shall move each card to `pushed` only after verifying that its
  recorded merge commit is an ancestor of the remote-tracking ref of the integration branch.
- **REQ-FR-021** (State-driven): **While** the repository has no remote configured, the record shall
  admit `merged-local` → `done` directly and record the note `no remote — no CI verdict` on the
  transition event; **while** a remote is configured, it shall refuse that edge.

### Queue boundary and commands

- **REQ-FR-022** (Unwanted): The F1 record layer shall not alter the queue database schema and shall
  not write any queue item row; it shall admit a card to `picked` only when the queue item for that
  card id is in the queue state `picked`.
- **REQ-FR-023** (Ubiquitous): The `moai factory` command shall provide `assign` (record a picked card
  with optional `--to`, `--prefer`, `--after`, SPEC, and worktree values, moving it to `assigned` when
  `--to` is given), `status` (report every card record of the selected run with state, stage, version,
  owner, lease expiry and whether it has expired, pending gate, hints, and contract reference, in text
  and JSON), and `decide` (the gates of REQ-FR-018 through REQ-FR-020).
- **REQ-FR-024** (Unwanted): `moai factory status` shall not write to the factory database, including
  the lease-expiry return of REQ-FR-014; it shall report an expired lease as expired.
- **REQ-FR-025** (Event-driven): **When** an existing lead dispatch path records a card assignment
  (the queue dispatch and the auto-mission dispatch), it shall also record the card in the factory
  record as `assigned` to that lane through the transition API, while the lane-visible dispatch, the
  message delivery, and the existing queue runtime report remain as they are today.

## §C. Operator decisions (2026-09-26) and how F1 applies them

- **All `-f` launches become self-dispatch; no headless engine; the numeric `-f N` form is removed;
  the lead is harness-agnostic (cc, glm, or codex).** F1 is the record these rest on and changes no
  launcher. Measured: `-f N` is already rejected (`internal/cli/factory.go:67-73`, `:237-240`), but the
  `-k N` factory shape is still accepted (`internal/cli/kanban.go:124-131`) and the conflict error still
  teaches `-f [N]` (`factory.go:202-203`). Both are launcher work and belong to F2 (plan.md §C.1).
- **In autonomous mode the local merge is automatic.** The actor that performs it is the F3
  controller; F1 supplies the `merge-ready → merging → merged-local` edges and their evidence gate.
- **The Kickoff decider is a single choice `human | llm | jev`, default `llm`.** The Decider interface
  and the non-human deciders are F3. F1 records the decider field and accepts only `human`
  (REQ-FR-018), so no non-human approval can reach the record before F3 defines how it is evidenced.
- **Share the store with A1 (t1234).** A1 keeps its contract as a file (`contract.yaml`) and touches no
  database; its F1 reference shape is a three-field pointer. F1 stores that pointer and nothing else
  (REQ-FR-001). Detail and the A3 receipt store: `research.md` § R11-R12.

## §D. Exclusions

### Out of Scope — F2 self-dispatch and launcher

- Lane-facing verbs (`moai factory next`, `stage`, `complete`, and a heartbeat verb) that let a worker
  acquire, renew, and advance its own card. F1 exposes these as a Go API exercised by tests only.
- Launcher changes: removing the `-k N` factory shape, correcting the `-f [N]` error text, and making
  every `-f` launch self-dispatch.
- Moving lane stage reporting (plan → run → sync) onto the record; F1 moves only the assignment.

### Out of Scope — F3 controller and Decider

- `moai factory run`, the controller loop that performs lease reclamation, automatic local merges,
  and CI polling.
- The Decider interface and the `llm` and `jev` deciders, including their evidence and receipts.
- Producing the CI verdict evidence file consumed by `pushed → ci-green`.
- Optional notifications.

### Out of Scope — Queue and contract

- Any change to the queue `items` table or any other queue table schema, and any automatic promotion
  of a queue item (promotion stays the operator's act).
- Verifying a contract's signature or digest; that is `moai contract verify` (A1). F1 stores the
  pointer only.

### Out of Scope — Documentation surfaces

- docs-site pages, README text, and rule-file edits describing the new commands; sync-phase work for
  manager-docs.

## §E. Residual risk

- **Evidence files are local.** `.moai/reports/` is gitignored, so a verdict file is evidence on the
  machine that wrote it. The audited SHA binding (REQ-FR-009) prevents a verdict for another commit
  from passing; it does not prevent a verdict file being written by an agent rather than an auditor.
- **Two stores during migration.** Until F2 moves lane reporting, the queue runtime report and the
  factory record both describe assignments. The factory record is authoritative for state; the queue
  report stays a report (plan.md §C.3).
