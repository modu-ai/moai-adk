---
id: SPEC-FACTORY-RECORD-001
title: "Harness-neutral factory F1 — card record layer and version-checked card state machine"
version: "0.2.1"
status: completed
created: 2026-09-26
updated: 2026-09-27
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/homestate, internal/cli, internal/kanban, .claude/agents/moai"
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
| 0.2.0 | 2026-09-26 | manager-spec | Plan-audit iteration 1 (FAIL 0.71, `.moai/reports/t1239/plan-audit.md`) revision. D1: `kickoff` is a lease-released decision-pending state, approval returns the card to `assigned` with stage `run` (REQ-FR-018/019). D2: the resume edge is split into six concrete stage-guarded edges; AC-005 carries a fixed edge count (65). D3: `pushed → ci-green` and `ci-green → done` are reserved and refused in F1 (REQ-FR-006). D4: `unblock` (REQ-FR-020) and `failed` (REQ-FR-015) get their own requirements and AC-017. D5: AC-024 / AC-025 cover both dispatch paths. D6: the auditor verdict-line producer is a requirement of this SPEC (REQ-FR-011, AC-020, milestone M3b). D7: some compound requirements merged or split within the 25 ceiling. D8/D9: abandon of a plain card and an AC-002 mutation line added. Lead updates: A1 re-read at `de8aee456` (v0.5.1); contract pointer extended with a store-event locator into `$MOAI_HOME/db/<project-key>/contract/` (lead decision R10). REQ and AC renumbered; 25 REQ, 25 AC. |
| 0.2.1 | 2026-09-26 | manager-spec | Plan-audit iteration 2 (PASS-WITH-DEBT 0.94, `.moai/reports/t1239/plan-audit-iter2.md`) debt notes, no REQ or AC count change: research.md R16 records the existing `AUDIT-VERDICT:` chat-message line (`internal/auditreceipt`, SPEC-CODEX-AUDIT-GATE-AXES-001) and corrects R13; verdict-file-only guardrail added to design.md, plan.md M3b, and AC-020 (D10); rationale for excluding `blocked` from T21 (D11); forward note that computing `contract_event` belongs to A3/F3 (D12). Lead decisions of the same day recorded: all ten plan.md §C defaults adopted; decision 5 strengthened in REQ-FR-025 (readable unavailable-record log, reported by `status`, and a `record.drift` event on the next successful write; AC-024 / AC-025); decision 10 cites A3 at `710530d67`; A1 re-pinned to `8a7cb0e22` (v0.5.2, F1 shape unchanged). |

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
- Auditors end their final chat message with a machine-readable `AUDIT-VERDICT:` line
  (`research.md` § R16), but no auditor writes a machine-readable verdict line into the exported
  verdict **file** (§ R13), so an evidence gate that reads verdict files needs its producer built in
  the same SPEC.

F1 is the first of three cards. It delivers the record layer only: the schema, the state machine, the
transition API with its evidence gates, the lease, the auditor verdict-line producer the gates need,
and three operator commands. F2 (self-dispatch lane verbs and launcher) and F3 (controller, Decider,
CI verdict reader) consume it.

## §B. Requirements (GEARS)

### Record schema and migration

- **REQ-FR-001** (Ubiquitous): The factory card record shall carry, per card within a run, its
  state, the last working stage it reached, a version counter, the owner label, the lease holder and
  lease expiry, the last heartbeat time, the pending decision (gate, question, state to resume, and
  once decided the decider and decision time), the failure reason, the assignment hints `prefer` and
  `after`, the SPEC identifier, the card worktree path, the last accepted evidence path and commit
  SHA, the local merge commit SHA, and an optional contract pointer consisting of a SPEC identifier, a
  signed contract digest, a signing time, and a locator of the signing event in the moai-owned contract
  store `$MOAI_HOME/db/<project-key>/contract/`.
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
  `abandoned`, shall accept exactly the requested edges enumerated in `design.md` § Transition Table
  when each edge's guard holds, and shall refuse every other requested transition with an
  illegal-transition error that leaves the card row and the event log unchanged.
- **REQ-FR-005** (Ubiquitous): The transition API shall apply each transition as one transaction that
  compares the caller's expected version with the stored version, refuses the transition with a
  stale-version error and no change when they differ, and otherwise writes the new state, increments
  the version by one, and appends one `card.transition` event naming the card, the source and target
  states, the new version, the actor, and the evidence read — committing all of these or none.
- **REQ-FR-006** (Unwanted): The F1 transition API shall not accept the edges `pushed → ci-green` and
  `ci-green → done`; it shall refuse them with a reserved-edge error naming F3 as the owner of the CI
  verdict reader, and leave the card row and the event log unchanged.

### Evidence gates

- **REQ-FR-007** (Event-driven): **When** a card enters `plan-audit` or `sync-audit`, the transition
  API shall accept the transition only after it has itself verified that the supplied commit SHA
  resolves in the card's worktree, is an ancestor of or equal to that worktree's HEAD, and contains
  the supplied artifact path at that commit; it shall record that SHA and path as the card's audited
  evidence.
- **REQ-FR-008** (Event-driven): **When** a card leaves `plan-audit` for `kickoff` or leaves
  `sync-audit` for `merge-ready`, the transition API shall read the card's audit verdict file under
  `.moai/reports/<card-id>/` and accept the transition only when the file's machine-readable verdict is
  `PASS` or `PASS-WITH-DEBT` and its machine-readable audited SHA equals the SHA recorded at audit
  entry; a verdict file that names the recorded SHA with any verdict shall admit the return edge to
  `plan` or `sync`.
- **REQ-FR-009** (Event-driven): **When** a card enters `merged-local`, the transition API shall
  accept the transition only after verifying that the supplied merge commit is reachable from the local
  integration branch, that the merge commit's tree equals its second parent's tree, and that the
  supplied re-measure evidence file exists and names that merge commit; it shall record the merge
  commit SHA, the tree hash, and the re-measure evidence path.
- **REQ-FR-010** (Unwanted): The transition API shall not accept a caller-supplied verdict, pass flag,
  tree hash, or ancestry claim as a substitute for the verification it performs itself.
- **REQ-FR-011** (Ubiquitous): The plan-auditor and sync-auditor agent definitions (the local copies,
  their template mirrors, and the emitted Codex definitions regenerated from those mirrors) and the
  audit-artifact convention document shall instruct the auditor to write, at the start of a line in
  every exported card verdict file, `verdict: <PASS|PASS-WITH-DEBT|FAIL>` and `audited_sha: <commit
  SHA the audit read>`.

### Lease, heartbeat, and failure

- **REQ-FR-012** (Event-driven): **When** a worker acquires the lease on an `assigned` card, the
  transition API shall accept only a worker label registered in the factory worker roster and equal to
  the card's owner label, move the card to `leased`, and record the holder, the heartbeat time, and an
  expiry equal to the heartbeat time plus the lease duration; **when** the holder renews its heartbeat,
  the record shall extend the expiry and update the roster's heartbeat time, and a renewal by any other
  label shall be refused with no change.
- **REQ-FR-013** (State-driven): **While** a card's lease has expired and its state is a
  lease-holding state other than `merging`, the transition API shall, before any other transition on
  that card, return the card to `assigned` with the holder cleared and the stage, worktree path, and
  evidence fields preserved, append a `lease.expired` event, and modify no file or git ref in the card's
  worktree.
- **REQ-FR-014** (State-driven): **While** a card's lease has expired and its state is `merging`, the
  transition API shall move the card to `blocked` rather than `assigned`, and append a `lease.expired`
  event naming the interrupted merge.
- **REQ-FR-015** (Event-driven): **When** the lease holder moves a lease-holding card to `failed`, the
  transition API shall require a non-empty reason, record it on the card, clear the lease, and refuse
  every later transition from `failed`.

### Assignment hints

- **REQ-FR-016** (Where): **Where** a card carries an `after` hint naming a predecessor card, the
  transition from `picked` to `assigned` shall be refused until the predecessor has a factory record in
  `merged-local`, `pushed`, `ci-green`, or `done`; a predecessor with no factory record shall refuse the
  transition with an unknown-predecessor error.
- **REQ-FR-017** (Where): **Where** a card carries a `prefer` hint (for example `backend=codex`), the
  record shall store it verbatim and report it, and F1 shall not use it to refuse or reorder any
  transition.

### Decisions and gates

- **REQ-FR-018** (Event-driven): **When** a card enters a decision-pending state (`kickoff` on the
  `kickoff` gate, `needs-decision` on the `question` gate), the record shall store the gate, the
  question for `needs-decision`, and the state to resume, clear the lease holder and expiry, and
  thereafter apply no lease expiry to that card until it leaves the decision-pending state.
- **REQ-FR-019** (Event-driven): **When** an operator runs `moai factory decide` on the `kickoff` gate
  with the choice `approve` or `reject`, the record shall move each named card in `kickoff` to
  `assigned` with stage `run` or to `blocked` respectively, recording the decider as `human` and the
  decision time; one invocation naming several cards shall apply one independent version-checked
  transition per card, and a decider value other than `human` shall be refused.
- **REQ-FR-020** (Event-driven): **When** the operator decides `resume`, `block`, `unblock`, or
  `abandon`, the record shall move a `needs-decision` card to its resume target (per `design.md`
  § Transition Table T22), a `needs-decision` card to `blocked`, a `blocked` card to `assigned` with its
  stage preserved, or any non-terminal card to `abandoned`, respectively.
- **REQ-FR-021** (Event-driven): **When** the lead runs `moai factory decide` on the `push` gate for
  cards in `merged-local`, the record shall move each card to `pushed` only after verifying that its
  recorded merge commit is an ancestor of the remote-tracking ref of the integration branch; **while**
  the repository has no remote configured, it shall instead move the card to `done` with the event note
  `no remote — no CI verdict`, and **while** a remote is configured it shall refuse `merged-local → done`.

### Queue boundary and commands

- **REQ-FR-022** (Unwanted): The F1 record layer shall not alter the queue database schema and shall
  not write any queue item row; it shall admit a card to `picked` only when the queue item for that
  card id is in the queue state `picked`.
- **REQ-FR-023** (Ubiquitous): The `moai factory` command shall provide `assign` (record a picked card
  with optional `--to`, `--prefer`, `--after`, SPEC, worktree, and contract-pointer values, moving it to
  `assigned` when `--to` is given), `status` (report every card record of the selected run with state,
  stage, version, owner, lease expiry and whether it has expired, pending gate, hints, and contract
  pointer, in text and JSON), and `decide` (the gates and choices of REQ-FR-019 through REQ-FR-021).
- **REQ-FR-024** (Unwanted): `moai factory status` shall not write to the factory database, including
  the lease-expiry return of REQ-FR-013; it shall report an expired lease as expired.
- **REQ-FR-025** (Event-driven): **When** either existing lead dispatch path records a card assignment
  — the queue dispatch (`internal/cli/gtd.go`) and the auto-mission dispatch (`internal/cli/goal.go`) —
  it shall also record the card in the factory record as `assigned` to that lane through the transition
  API, while the lane-visible dispatch, the message delivery, and the existing queue runtime report
  remain as they are today; a factory-record write failure shall not fail the dispatch, and shall be
  recorded both as a `FACTORY_RECORD_UNAVAILABLE` line on stderr and as an entry (time, run, card,
  lane, error) in a readable log under the project's factory state directory, which
  `moai factory status` shall report; and the next successful factory-record write for that run shall
  append one `record.drift` event per unreconciled entry, naming the dispatched card and lane and the
  factory record's actual state for that card (or its absence), and then mark the entry reconciled.

## §C. Operator decisions (2026-09-26) and how F1 applies them

- **All `-f` launches become self-dispatch; no headless engine; the numeric `-f N` form is removed;
  the lead is harness-agnostic (cc, glm, or codex).** F1 is the record these rest on and changes no
  launcher. Measured: `-f N` is already rejected (`internal/cli/factory.go:67-73`, `:237-240`), but the
  `-k N` factory shape is still accepted (`internal/cli/kanban.go:124-131`) and the conflict error still
  teaches `-f [N]` (`factory.go:202-203`). Both are launcher work and belong to F2 (plan.md §C.1).
- **In autonomous mode the local merge is automatic.** The actor that performs it is the F3
  controller; F1 supplies the `merge-ready → merging → merged-local` edges and their evidence gate.
- **The Kickoff decider.** A1 v0.5.1 (read at `8a7cb0e22`, v0.5.2) records the operator's re-decision: the decider set
  is `human | llm | llm+jev`. The Decider interface and the non-human deciders are F3. F1 records the
  decider field and accepts only `human` (REQ-FR-019), so no non-human approval can reach the record
  before F3 defines how it is evidenced. A Kickoff wait is a human wait, so `kickoff` holds no lease
  (REQ-FR-018): no lease timer can overturn a pending decision.
- **Share the store with A1 (t1234) and the moai-owned contract store (lead decision R10).** A1 keeps
  its contract as a file (`contract.yaml`) and touches no database; A2 armed state and A3 signing
  events live under `$MOAI_HOME/db/<project-key>/contract/`, the sibling of `factory/` and `todo/` under
  the same project directory. The card record holds a pointer into that store and nothing else
  (REQ-FR-001). Detail: `research.md` § R11-R12.

## §D. Exclusions

### Out of Scope — F2 self-dispatch and launcher

- Lane-facing verbs (`moai factory next`, `stage`, `complete`, and a heartbeat verb) that let a worker
  acquire, renew, and advance its own card. F1 exposes these as a Go API exercised by tests only.
- Launcher changes: removing the `-k N` factory shape, correcting the `-f [N]` error text, and making
  every `-f` launch self-dispatch.
- Moving lane stage reporting (plan → run → sync) onto the record; F1 moves only the assignment.

### Out of Scope — F3 controller, Decider, and CI verdict

- `moai factory run`, the controller loop that performs lease reclamation, automatic local merges,
  and CI polling.
- The Decider interface and the `llm` and `llm+jev` deciders, including their evidence and receipts.
- The CI verdict reader and the `pushed → ci-green → done` edges, reserved by REQ-FR-006.
- Optional notifications.

### Out of Scope — Queue and contract

- Any change to the queue `items` table or any other queue table schema, and any automatic promotion
  of a queue item (promotion stays the operator's act).
- Verifying a contract's signature or digest, and writing to the contract store; those are
  `moai contract verify` (A1) and the A2/A3 store writers. F1 stores the pointer only.

### Out of Scope — Documentation surfaces

- docs-site pages and README text describing the new commands; sync-phase work for manager-docs. The
  auditor verdict-line instructions (REQ-FR-011) are not documentation for this purpose: they are the
  producer the evidence gate depends on, and they are run-phase deliverables.

## §E. Residual risk

- **Evidence files are local.** `.moai/reports/` is gitignored, so a verdict file is evidence on the
  machine that wrote it. The audited SHA binding (REQ-FR-008) prevents a verdict for another commit
  from passing; it does not prevent a verdict file being written by an agent rather than an auditor.
- **Two stores during migration.** Until F2 moves lane reporting, the queue runtime report and the
  factory record both describe assignments. The factory record is authoritative for state; the queue
  report stays a report (plan.md §C.3).
- **`pushed` is a dead end in F1.** With `pushed → ci-green` reserved, a card with a remote stops at
  `pushed` until F3 lands, leaving only `needs-decision` and `abandoned` as exits. This is intentional:
  an F1 reader for CI evidence would be an unverified claim.
- **The unavailable-record log shares a directory with factory.db.** A failure that makes the whole
  factory state directory unwritable also prevents the log entry; then the stderr line is the only
  record, and the next successful write has nothing to reconcile. `status` and the drift event cover
  the common case (the database write fails, the directory is fine), not that one.
- **The contract pointer is not verified.** A stale or fabricated pointer is readable by `status`; A1's
  `moai contract verify` and the A3 store are the authorities.
- **Schema v4 is a one-way migration (F5, Opus sync-audit).** `factorySchemaVersion = 4` has no
  downgrade path: a v3 binary rejects a v4 database with `unsupported factory schema version "4"`,
  and the handoff hooks (`internal/hook/handoff/pending.go`, `persist.go`) and `factorymsg` open
  the same database. In a repository where lanes run different builds, the first v4 open locks
  every v3-binary path (handoff + factory) out of the project. Operational order: update all lane
  binaries first, then let the new database be opened.
- **`--decider human` is a self-report (F1, Opus sync-audit).** `moai factory decide` records the
  decider the caller names; nothing yet distinguishes an operator terminal from an agent lane
  running the same command through Bash, and no hook restricts the command. F1 has no consumer
  that grants run entry on this record, so there is no present harm — but when F2/F3 start
  treating this record as an authority basis, they must first require proof of a human decision
  (for example a PreToolUse denial of `moai factory decide` from agent sessions, or an
  interactive TTY check).
