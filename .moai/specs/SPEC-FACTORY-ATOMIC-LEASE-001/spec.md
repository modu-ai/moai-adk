---
id: SPEC-FACTORY-ATOMIC-LEASE-001
title: "Atomic lease across the queue store and the factory record — one critical section from the selection read to the claim, a bounded record wait, and a worktree step that no longer renames"
version: "0.1.0"
status: draft
created: 2026-10-03
updated: 2026-10-03
author: manager-spec (card t1458)
priority: P1
phase: "v3.2.0 target"
module: "internal/cli, internal/kanban, internal/homestate"
lifecycle: spec-anchored
tags: "factory-next, atomic-lease, queue-lock, serial-slot, compensation, worktree-rename, card-t1458"
tier: M
card: t1458
depends_on: [SPEC-TODO-AUTO-PICK-001, SPEC-TODO-CLASSIFY-DISPATCH-001]
related_specs: [SPEC-FACTORY-SELF-DISPATCH-001, SPEC-TODO-HOLD-STATE-001, SPEC-BACKLOG-LOCK-BUDGET-001, SPEC-RESOURCE-SLOT-LEASE-001, SPEC-WORKTREE-BASEREF-001, SPEC-TODO-CLAIM-LEASE-001]
---

# SPEC: atomic lease across the queue store and the factory record

## HISTORY

- 0.1.0 — 2026-10-03 — plan-phase artifacts authored (card t1458, Class C, Tier M; worktree
  `.claude/worktrees/t1458`, branch `WT-atomic-lease`, plan-start HEAD `2de0a2cb6`, full SHA
  `2de0a2cb613b04765a1554f86685a3b48e0be806`). Inputs: the t1448 sync-audit iteration 2
  (`.moai/reports/t1448/sync-audit-iter2.md`) findings F3(a), F3(c), F14 and gap G-rename, which the
  operator accepted as residual risk on 2026-10-03 and handed to this card; the t1407 verdict's
  option B (arm (a) reads a sibling snapshot, then claims non-atomically); and the card's own
  `git branch -m` hazard. The design question the card poses — three options, one hypothesis to test
  — is answered in §B with observations made in this tree; the probes that produced them are kept as
  re-executable evidence in `evidence/`. Two decisions were **corrected by measurement during
  planning** and the record keeps both: the naive "hold the queue lock across the claim" form starves
  other queue writers under a stalled record (§B.2), and the expected M5 fix — create the branch with its
  final name — still fails 2 of 280 concurrent creations, so the step is serialized instead (§B.3).
  Requirement count 13, criterion count 14 (Tier M ceilings 16 / 16).

## §A Context

### A.1 Problem

`moai factory next` leases a card in two stores that share no transaction. The **queue**
(`backlog.db`) is guarded by a cross-process lock that every queue writer takes through
`BacklogStore.Mutate`. The **factory record** (`factory.db`) is a separate SQLite file with its own
write lock and version-checked edges. The lease path reads both, decides, promotes the card in the
queue, and claims it in the record — three steps with nothing around them. Two lanes, or a lane and
the operator, can interleave between any two of the steps, and nothing the record offers can undo a
claim that landed (§A.2 O7).

Five symptoms are in scope. Each was **reproduced in this tree** (§A.2; `acceptance.md` evidence
ledger), not taken from the audit's reading.

| Id | Name | What happens | Reproduced as |
|---|---|---|---|
| M1 | F3(a) | Two lanes lease two DIFFERENT serial cards at once and both end `leased`. The nominated path reuses the `serialHeld` snapshot it read before its queue lock; arm (c) reads its serial slot from a record snapshot taken before either lane promoted. | both `leased` — nominated probe 2 of 2 runs, bare-form probe 8 of 8 runs |
| M2 | F3(c) | The operator holds a card after the lease path decided to take it and before the claim; the card is leased anyway. Same class in unnominated arm (c), which promotes and claims with no re-check. | `queue=hold record=leased holder=lane-1 err=<nil>` on both arms |
| M3 | F14 | The compensation of a failed nominated claim cannot tell its own promotion from an operator's later fresh pick of the same card (ABA) and reverts the operator's pick to `queued`. | `queue=queued` after the operator's `picked` |
| M4 | t1407 option B | Arm (a), a lane leasing the serial card assigned to itself, reads the sibling snapshot and then claims non-atomically; two lanes with two assigned serial cards both lease. | both `leased` |
| M5 | rename | `factoryEnsureCardWorktree` renames the new branch with `git branch -m`; two lanes in one repository collide on git's shared reflog temp file, the verb errors after the lease succeeded, and the half-built directory is then refused by the foreign-worktree precheck that a re-lease runs (the precheck was observed to refuse; a full re-lease was not run). | 29 of 80 concurrent iterations failed across four runs (6, 5, 8 and 10 of 20); in the two runs that counted them, all 18 failures left a directory the foreign-worktree precheck refuses |

M1 to M4 are one root cause seen from four sides: the decision and the claim are not inside one
exclusion. M5 is a different hazard in the step that follows the lease, reached by the same callers.

### A.2 Measured basis (pinned tree `2de0a2cb6`)

Measurements were taken in this run, against this tree, through `go test -overlay` probes that are
committed under `evidence/` (the Gaps in §F say what each probe does not show). "Read" marks a fact
established by reading source, not by running it.

- **O1 — the queue lock's wait budget is 3.3 s.** `boardLockWaitBudget` printed
  `3.3s` (`boardLockSupportedWriters=10`, `boardLockCIMutationCost=33ms`, `boardLockHeadroom=10`).
  The t1448 audit quoted 1.65 s; that was the figure at headroom 5, before commit `0e98ffe6c` doubled
  it.
- **O2 — both stores are WAL SQLite.** The factory record opens with `busy_timeout(5000)`,
  `journal_mode(WAL)`, `_txlock=immediate`, one open connection (read, `factory.go`); the queue engine
  also sets WAL (read, `backlog_sqlite.go`).
- **O3 — an uncontended claim is cheap, and load moves it.** RecordPicked plus the two transitions:
  n=100 per run, p50 1.0–2.0 ms, p95 1.5–16.8 ms, max 7.8–72.5 ms over three runs on a machine other
  sessions were also using.
- **O4 — a reader is not blocked by a writer.** `LoadCard` while another connection held a write
  transaction: 0.05–0.63 ms.
- **O5 — a claim's wait for the record is not bounded by a context deadline.** Against a 3 s write
  holder a claim carrying a 500 ms deadline returned after 3.03–3.11 s; with no deadline against a
  7 s holder it succeeded after 7.00–7.08 s, past the 5000 ms busy timeout, because the retry loop
  re-enters it (read, `factory.go` `retryFactoryBusy`).
- **O6 — a low busy timeout set at open time does bound it; a runtime PRAGMA does not last.** A
  connection opened and set to 200 ms plus a 500 ms deadline returned after 0.68–0.71 s; the same handle
  with a 1 s deadline against a 7 s holder ran 5.09–5.14 s, and the handle read back 5000 ms afterwards. The
  per-handle PRAGMA did not survive a context-cancelled call (an inference: the connection was
  replaced), so a cap has to come from the connection's own DSN.
- **O7 — a landed claim cannot be undone by a lane.** The edges leaving `leased` are the stage
  resumes T4a–f, T21 (needs-decision), T25 (abandon — a human decider) and T26 (fail — `failed` is
  terminal, nothing leaves it) (read, `card_transition.go`). There is no un-lease, so a design that
  claims first and verifies afterwards has no compensation to run.
- **O8 — the layering forbids one inversion by construction.** `internal/kanban` imports
  `internal/homestate`; `internal/homestate` imports no `internal/kanban` (`go list -deps` of both).
  A factory transaction therefore cannot call into queue code or wait for the queue lock.
- **O9 — a factory write transaction can contain git work.** The transition guards
  `verifyAuditEntry`, `verifyCommitAtHead` and `verifyMerge` run inside the write transaction (read,
  `card_transition.go` `planTransition`), so `stage` and `complete` can hold the record's write lock
  across git subprocesses. Their duration was not measured.
- **O10 — the two stores are consumed by different code.** 31 non-test Go files name the factory-record
  API, 35 name the queue-store API, 64 in union, 2 in both (`factory_card.go`, `factory_mirror.go`) —
  counted with a pipeline at this tree.
- **O11 — removing the rename does not remove the race.** `WorktreeCreator(name)` uses `name` as the
  directory leaf and as the branch (`session_worktree.go` `materializeSessionWorktree`,
  `gitWorktreeAddReal`), which is why the branch is renamed afterwards; a creator that took the branch
  separately would need an interface change. Two lanes creating worktrees concurrently in one repository
  with their FINAL `WT-` names failed in 2 of 280 iterations over seven runs (0, 1, 0, 1, 0, 0, 0 of
  40), with `fatal: failed to read .git/worktrees/wt-b/commondir` — `git worktree add` itself races a
  sibling `git worktree add`. The rename step failed in 29 of 80 (§A.1).
- **O12 — the worktree step has four production callers** (`factory_card.go`, `mcp_factory_card.go`,
  `factory_lane_relaunch.go`, `codex_launcher.go`), all through one function, so one lock around that
  function covers them.
- **O14 — serializing the whole step removes both failures, and it is cheap.** Two lanes' worktree steps
  (create plus rename) run one at a time under one lock: 0 of 80 iterations failed over four runs; one
  step took 0.42–0.69 s at best, 0.73–1.22 s at the median and 1.14–2.88 s at worst (the slower figures
  came from the later runs, made while the machine was busier). The probe used an
  in-process mutex, which orders the two git processes exactly as a cross-process lock would.
- **O13 — SPEC-TODO-AUTO-PICK-001 is at its Tier M ceiling.** Its HISTORY states "requirement count
  stays 16 (Tier M ceiling)". An amendment that adds any requirement would force a tier-up.

## §B Decisions

### B.1 Q1 — how the two stores become one step

The card names three options. All three were weighed against the measurements above.

| Option | Closes | Does not close | Measured cost |
|---|---|---|---|
| **1. Merge the two stores into one** | M1–M4 in principle | M5 | Both stores must move under one engine and one schema owner. 64 files across the two APIs (O10); `kanban` already imports `homestate`, so the merge picks a direction and inverts the other package's layering (O8). The factory schema is at version 5 with four sequential migrations; the queue has its own migration and downgrade machinery. SQLite `ATTACH` is not a shortcut: the SQLite documentation states that in WAL mode a transaction across attached databases is "atomic for each individual database, but … not atomic across all databases as a set" (sqlite.org/wal.html), and both stores are WAL (O2). Highest cost; touches two live operator stores. |
| **2. Version-conditional writes plus re-validation across stores** | M3 (a promotion token), part of M2 | M1, M4, and the final check-to-claim gap of M2 | The record side is already version-checked. The queue side has no version; `PickedAt` is the episode stamp but the lease promotion never writes it (read), so a token needs a stamp or a schema change. More importantly there is no un-lease edge (O7), so a claim that lands after a stale check cannot be compensated: a compare-and-set narrows the window and cannot close it. |
| **3. One lease-side lock** | M1, M4, arm (c) between lanes | M2, M3 | Serializes lease against lease and leaves every operator queue write (`unpick`, `hold`, `drop`, `add`, …) outside it, because those take only the queue lock. Making the operator writers take it too is the queue lock under another name. |
| **3′. Hold the existing queue lock across select, promote and claim** (the hypothesis) | **M1–M4** | the windows in §F | See B.2. No token and no un-lease edge are needed: the decision and the claim sit inside one exclusion shared with every other queue writer. |

**Recommendation: option 3′.** It is the only option that closes all four related symptoms by
exclusion rather than by detection, it reuses a lock every writer already honors, and its cost is
measurable (B.2). Option 2's promotion token stays available as defence in depth and is not adopted:
under 3′ the compensation runs inside the promotion's own section, so there is no interleaving left
for a token to detect.

### B.2 The hypothesis, tested

Hypothesis: hold the existing cross-process queue lock across select, promote and claim, so the
factory-record write happens inside the queue lock. Two questions: does it close the symptoms, and
what does it do to the lock's other users?

**Does it close them — yes by construction, and each symptom was reproduced RED first (§A.1).** After
the change the operator's write waits for the lease, and the second lane's decision reads the first
lane's lease, because both are inside the same exclusion. What exclusion cannot give is an operator
write ordered *before* a lease that it arrived after: a `hold` that arrives mid-section applies after
the lease, so the same end state `queue=hold record=leased` can still be reached by that order. That is
a lease followed by a hold, which is what any later hold does, and the criteria test the ordering, not
the end state (AC-FAL-002).

**What it costs — confirmed typical, refuted in its naive form.**

- Typical: an uncontended claim is p50 1.0–2.0 ms, p95 1.5–16.8 ms, load-dependent (O3). With the one queue
  promotion (the lock's sizing figure is 33 ms per mutation, O1) the section is of the order of one
  extra mutation. Ten lanes serialized behind it stay far inside the 3.3 s budget (a derived estimate,
  not a measurement; the run phase measures the real distribution, AC-FAL-009).
- Under a stalled record the naive form fails. The claim's wait for the record is not bounded by a
  context deadline (O5), so an unbounded claim inside the lock extends the queue lock's hold to the
  full stall — 7.0 s against a 3.3 s budget in the 7 s probe, which makes every other queue writer
  fail with a lock-held error. A stall source exists in the code: stage and complete transitions run git
  inside the record's write transaction (O9).
- The repair is a bounded claim: the lease path opens its own record connection with a low busy timeout
  set in the DSN (O6) and bounds the whole claim by a deadline, so the section's worst hold is a named
  cap, and the cap is derived to be at most one third of the queue lock's wait budget (1.1 s at
  today's constants) — the convention `slot_lease_cross_test.go` already uses for its stall-release
  timeout. On the cap the lease gives up as a lost race and restores its promotion.
- A stalled record still costs liveness: concurrent lanes serialize behind the lock at up to one cap
  each, so with more than three lanes waiting (3.3 s / 1.1 s) the later ones exhaust the lock's own wait
  budget. That is stated as a residual (§F R6) and was not measured under multi-lane stall.

**Lock order.** Queue lock then record is the only order the lease path uses. The t1448 audit found no
path that holds a record write and then waits for the queue lock; this tree gives the stronger
statement that none can exist inside `internal/homestate` (O8). The new criteria pin it (AC-FAL-011).

### B.3 Q2 — M5: serialize the whole worktree step

Three approaches were weighed: avoid the rename by creating the branch with its final name, retry on the
specific `unable to move logfile` error, or serialize the step with a repository-level lock. The first
was the expected winner and **the measurements overturned it**.

- **Avoid the rename — measured insufficient.** The rename collision (29 of 80) is in `git branch -m`:
  git moves the reflog through one temp name in the common git directory. Creating the branch with its
  final name removes that, and what is left is git's own race between two `git worktree add` runs
  (2 of 280, O11). It would also need the shared creator to take the branch and the directory leaf
  separately — an interface change across four callers (O12) for a fix that still fails.
- **Retry on the error string — rejected.** It guards one symptom of two different races, and a failed
  rename leaves a half-built directory that the foreign-worktree precheck refuses (§A.1 M5), so a retry
  would first need cleanup logic.
- **Serialize the step — chosen.** One lock, distinct from the queue's lock, held around the creator
  call and the rename together, in the one function all four callers use (O12). Measured: 0 of 80
  failed (O14). Cost: one step is 0.42–2.88 s depending on machine load (median 0.73–1.22 s), so lanes
  that start at the same instant queue behind each other — roughly N × the median step for the last of N
  lanes (a derived figure; not measured at N=10). Lane startup is not latency-critical and the wait is
  bounded; it is a start-up cost, not a per-lease cost.
- **The lock must not be the queue lock** and must be released before the card-worktree record write:
  holding the queue lock across `git worktree add` (seconds) would starve every queue writer, which is
  what REQ-FAL-007 forbids.
- Honest limits: the lock orders this card's own lanes only (§F R11); two cards whose queue titles derive
  the same slug still collide on the branch name at rename, unchanged in kind and outside this SPEC;
  and a rename that fails for a reason other than the collision still strands the directory.

### B.4 Q3 — a new SPEC, not an amendment

A NEW SPEC, citing both completed SPECs through `depends_on`.

1. **It spans two completed SPECs and neither owns the lease path.** SPEC-TODO-CLASSIFY-DISPATCH-001
   owns serial exclusivity (REQ-TCD-008); SPEC-TODO-AUTO-PICK-001 owns the nominated lease and its
   compensation (REQ-TAU-004–006). The change is the section that surrounds both.
2. **AUTO-PICK cannot take another requirement.** It is at 16 of 16 for Tier M (O13); an in-place
   amendment that adds requirements forces a tier-up of a completed SPEC.
3. **CLASSIFY-DISPATCH already carries one in-place amendment** (0.4.0, card t1407) with its own
   known-limitation paragraph that this SPEC closes; a second, structurally larger one would bury the
   serial-slot text under lease-path mechanics.
4. **Supersession is recorded, not edited.** §E names each clause. The completed bodies are not
   touched in the plan phase. Recording the supersession in them through the Amendments mechanism is a
   sync-phase deliverable (REQ-FAL-013): the mechanism itself reopens the SPEC
   (`completed → in-progress`) and closes it again, which is a cost to accept deliberately and not to pay
   in a plan.

### B.5 What the critical section is not

It is not an atomic commit across two databases. The queue commit and the record commits are still two
separate commits; the section makes every other participant wait while they happen and orders the
compensation inside it. A lane killed inside the section, a record writer that never takes the queue
lock, and a stalled record are named in §F and are not claimed closed.

## §C Requirements

Thirteen requirements, GEARS notation. Every `REQ-FAL-NNN` is traced by at least one `AC-FAL-NNN` in
`acceptance.md`.

### Module A — The critical section

- **REQ-FAL-001** (Ubiquitous): The lease path of `moai factory next` — every arm of the bare form and
  the nominated form, the command line and the MCP tool alike — shall run its decision reads (the queue
  record, the factory record, the keep-set and the serial slot), its queue promotion, its claim of the
  factory-record row, and any compensation of that promotion inside one critical section that holds the
  queue's cross-process lock from the first decision read to the last claim write, and shall release
  that lock before it creates a worktree.
- **REQ-FAL-002** (Event-driven): **When** two lanes lease at the same time and the card each takes is
  serial — two different nominees, the two top queued serial cards through the bare form, or either
  mix — and no serial card is in flight, the lease path shall end with exactly one of those cards
  leased, and the lane that enters the critical section second shall read the first lane's lease and be
  refused with `serial-slot` (nominated form) or take no serial card (bare form). *(M1)*
- **REQ-FAL-003** (Event-detected): **When** an operator queue write — hold, unpick, drop, or any other
  write that takes the queue's lock — arrives while a lease critical section is open, the lease shall
  complete first and the operator write shall apply after it; **when** such a write committed before the
  critical section opened, the lease shall read it and shall not lease the card (the `held`, `dropped`
  or `raced` refusal for the nominated form, a skipped candidate for the bare form). *(M2)*
- **REQ-FAL-004** (Unwanted): The compensation of a failed nominated claim shall not revert a queue
  state written by anyone other than the invocation that made the promotion: it shall run inside the
  critical section of the promotion it undoes, so no operator write can fall between the two. *(M3)*
- **REQ-FAL-005** (Event-driven): **When** two lanes each lease the serial card assigned to themselves
  at the same time and nothing is in flight, the lease path shall lease exactly one; the other lane
  shall read the first lane's live lease inside its own critical section and skip its card, which stays
  `assigned`. *(M4)*

### Module B — Bounded hold

- **REQ-FAL-006** (Event-detected): **When** the factory record's write lock is not obtained inside the
  critical section within the lease-claim wait cap, the lease path shall stop waiting, restore any
  promotion it made, and report a lost race — the `raced` refusal for the nominated form, a
  re-selection for the bare form; **when** the queue's lock is not obtained within its own wait budget,
  the lease path shall report the same lost race and not an infrastructure failure. The cap shall be
  enforced by the connection the lease path opens for its record writes (not by a context deadline),
  shall bound the whole claim and not each write, and shall not exceed one third of the queue lock's
  wait budget.
- **REQ-FAL-007** (Ubiquitous): The critical section shall contain no git subprocess, no worktree
  creation, and no read or write outside the queue store and the factory record, and shall hold at most
  one queue promotion, one queue restore and the three record writes of a claim; worktree creation and
  the card-worktree record write shall run after the lock is released.

### Module C — The worktree step

- **REQ-FAL-008** (Event-driven): **When** two lanes run the post-lease worktree step at the same time
  in one repository, both shall succeed: the step — worktree creation and branch rename together — shall
  run under a dedicated cross-process lock that is not the queue's lock, with a bounded wait, and the lock
  shall be released before the card-worktree record write; the directory-leaf and recorded-path rules of
  SPEC-FACTORY-SELF-DISPATCH-001 REQ-SD-011 are unchanged. A lane that cannot obtain the lock within the
  wait shall fail the step with the step's existing error and create nothing. *(M5)*

### Module D — What stays true

- **REQ-FAL-009** (Ubiquitous): Outside the interleavings REQ-FAL-002 to REQ-FAL-006 close,
  `moai factory next` shall behave as before — arm order, candidate choice, output text, the closed set
  of twelve refusal tokens (none added), exit codes, the quota hold, the Codex backend skip, and a
  lane's own re-lease of its assigned card.
- **REQ-FAL-010** (Ubiquitous): The lease path shall take the queue's lock before it writes the factory
  record and never the reverse; no code path shall wait for the queue's lock while holding a
  factory-record write transaction; and `internal/homestate` shall not import `internal/kanban`.
- **REQ-FAL-011** (Ubiquitous): Every artifact of this card shall state the windows the critical
  section does not close (§F) and shall claim no atomicity beyond the critical section; a window it
  cannot close shall be named, not omitted.
- **REQ-FAL-012** (Ubiquitous): The change shall be surgical: lease-path code and the worktree step in
  `internal/cli`, the lock primitives (the queue's lock section and the worktree-step lock) in
  `internal/kanban`, the lease-path record open in `internal/homestate`, their tests, this SPEC's
  records, and any doctrine line that states the old residual risk — and nothing else; no queue verb, no
  queue or record schema, no transition-table row, no worktree-creator interface and no refusal token
  changes.
- **REQ-FAL-013** (Ubiquitous): This SPEC shall name each clause of SPEC-TODO-AUTO-PICK-001 and
  SPEC-TODO-CLASSIFY-DISPATCH-001 that it supersedes (§E), shall leave both completed bodies unedited
  in the plan phase, and shall record the supersession in both through their Amendments mechanism at
  the sync phase.

## §D Out of Scope

### Out of Scope — merging the two stores

- Option 1 of §B.1: moving the factory tables into the queue database, or the reverse, and any
  `ATTACH`-based arrangement. Reasons are measured in §B.1; a future card may revisit it.

### Out of Scope — a single-transaction lease edge in the factory record

- A combined `RecordPicked + assigned + leased` edge in `internal/homestate`. The claim stays three
  record writes, so its partial residue (§F R4) stays. Adding the edge would touch the transition table,
  whose size is pinned by a criterion of the SPEC that owns it.

### Out of Scope — record writers outside the lease

- `moai factory assign`, the leader's dispatch mirror (`mirrorFactoryAssignment`), `stage`, `complete`,
  `decide` and lease renewal do not take the queue's lock and are not changed. `assign` and the mirror
  read the queue lock-free and then write the record, which is the same class of window as M2 in a
  different verb (§F R2).

### Out of Scope — schemas, verbs, tokens and the transition table

- No queue or factory schema change, no new queue verb, no new refusal token (the `raced` token covers
  both a lost race and a bounded wait), and no row added to or removed from the card transition table.

### Out of Scope — worktree lifecycle beyond the lease's own step

- Worktree disposal, the `moai worktree` verbs and the session entry points (they do not take the new
  step lock, §F R11), the creator interface (unchanged — creating the branch with its final name was
  measured insufficient, §B.3), stranded-directory cleanup for failure causes other than the collision,
  and two cards deriving one branch slug.

### Out of Scope — unrelated t1448 follow-ups

- The operator-decision cards t810, t1294 and t1383, the `kanban-dispatch-detail.md` budget split, the
  `moai todo --help` `--auto` wording, and the lane-session doctrine wording of the keep-set. They are
  separate cards named in the t1448 verdict.

## §E Supersession record

Nothing in either completed SPEC is edited by the plan phase. The sync phase records each row below
through the Amendments mechanism (`completed → in-progress` per
`spec-frontmatter-schema.md` § Status Enum, with `prior_completed_sha`), through a manager-spec
re-delegation. SPEC-TODO-AUTO-PICK-001 has no `## Amendments` section today (its HISTORY carries the
versions), so its amendment creates one.

| Completed SPEC | Clause | Disposition once this SPEC is closed |
|---|---|---|
| SPEC-TODO-CLASSIFY-DISPATCH-001 | Amendments 0.4.0, "Known limitation (not closed by this amendment …)": the serial slot is read from a snapshot and the claim happens afterwards, two lanes can both lease a serial card; an atomic lease across both stores is future work | **Superseded for every lease arm** (REQ-FAL-002, -005). The paragraph's sentence that the leader named the follow-up t1458 without verifying it is also answered: the card exists. |
| SPEC-TODO-CLASSIFY-DISPATCH-001 | Same paragraph: "two lanes each leasing their OWN assigned serial card at the same moment is now an ordinary path … not closed here" | **Superseded** (REQ-FAL-005). |
| SPEC-TODO-CLASSIFY-DISPATCH-001 | Amendments 0.4.0, "Residual not repaired": an ownerless `picked` serial row and an `assigned` serial row wait on each other | **Untouched.** This SPEC does not repair or alter it. |
| SPEC-TODO-AUTO-PICK-001 | §B.8 "Non-atomicity": the record read and the queue write of the compensation are two stores with no shared lock, accepted and stated | **Superseded** (REQ-FAL-001, -004): the compensation runs inside the promotion's section. |
| SPEC-TODO-AUTO-PICK-001 | §G third bullet, "the record read and the queue write of the compensation are two stores without a shared lock" | **Superseded**, same reason. The post-`RecordPicked` residue half of that bullet stays (§F R4). |
| SPEC-TODO-AUTO-PICK-001 | REQ-TAU-005's compensation clause ("one queue write that acts only if the item is still `picked`") | **Strengthened, not contradicted:** the guard stays, and the write now cannot interleave. The F14 window (the guard cannot tell its own promotion from an operator's re-pick) is closed (REQ-FAL-004). |
| SPEC-TODO-AUTO-PICK-001 | REQ-TAU-006 (single ownership under concurrency) | **Unchanged and now also true for the serial slot** (REQ-FAL-002). |
| the t1448 run and sync records (`progress.md` §E.4, `CHANGELOG.md` limitations (a) and (b)) | The three windows and the `git branch -m` hazard stated as accepted residual risk | Run- and sync-owned; this card does not edit them. The sync phase of THIS card states, in its own CHANGELOG entry, which windows closed and which stay (REQ-FAL-011). |

## §F Gaps and residual risks

Stated plainly. A window not closed here is not closed; nothing in this SPEC's records may say
otherwise (REQ-FAL-011).

- **R1 — a lane process killed inside the section.** The lock is a `flock` released when the process dies
  (read, `board_lock_unix.go`), so nothing wedges, but a death between the queue promotion and the claim
  leaves the item `picked` with no row (or a `picked` unowned row). The next `next` re-adopts it through
  arm (b) or (b2). This is today's behavior for a crashed lane; it is not closed.
- **R2 — record writers that do not take the queue lock.** `factory assign` and the dispatch mirror read
  the queue without the lock and then write the record (`requireQueuePicked`, `writeFactoryAssignment`):
  an operator `unpick` between their read and their write is the M2 class of window in another verb. They
  and `stage`, `complete`, `decide` and lease renewal race the lease's claim at the record level, where
  the version-checked edges still arbitrate. Not closed; named for a follow-up card.
- **R3 — Windows.** The lock is `flock` on Unix and an atomic-create file with a bounded stale clear on
  Windows. The section's guarantee is cross-compile verified only on Windows (`GOOS=windows go build`), not
  exercised; the probes in this card ran on macOS.
- **R4 — the claim is three record transactions, not one.** A claim that fails after `RecordPicked` leaves
  the row `picked` and unowned (SPEC-TODO-AUTO-PICK-001 §B.8, which stays accurate on this point).
  The section restores the queue item; the row stays, as today.
- **R5 — a stalled record costs liveness, not correctness.** While the record's write lock is held longer
  than the cap, leases give up as `raced` and `--wait` retries. The stall source is out of scope.
- **R6 — multi-lane stall.** Under a persistent record stall each lane holds the lock up to one cap, so
  lanes queue behind one another; with more than about three waiting (budget 3.3 s over a cap of at most
  1.1 s) a later lane exhausts the queue lock's wait budget, and an operator queue write shares that
  exposure. The single-stall bound is tested (AC-FAL-007); the multi-lane case was not measured.
- **R7 — the stall evidence is constructed.** The probes hold a write transaction open to model a stall;
  no real WAL checkpoint stall or long git-in-transaction hold was observed (O9 is read, not measured).
- **R8 — the serialized step is measured at 0 of 80**, an upper bound of about 3.7% at 95% confidence,
  on one machine, with an in-process mutex standing in for the cross-process lock. AC-FAL-006 repeats it
  with the real lock and the real materializer.
- **R11 — the step lock orders only this card's own lanes.** `moai worktree new`, the session entry
  points, and any other git process reading `.git/worktrees/*` in the same repository (the observed
  failure is a read of a sibling's half-written `commondir`) can still race a lane's creation. Not
  closed, and not claimed.
- **R12 — lane start-up latency.** Lanes that begin at the same instant queue behind the step lock at
  0.42–2.88 s per step; the wait budget is derived from the worst observed step (plan D3), not measured
  at ten lanes.
- **R9 — the section's hold time on a real run is not measured yet.** B.2 gives an estimate from the
  component costs; AC-FAL-009 records the measured distribution in the run phase.
- **R10 — operator end state.** A hold that arrives while the section is open applies after the lease:
  `queue=hold record=leased` is reachable by that order. It is the ordering of two events, not a lease of
  a held card; the criteria test the ordering.

## §G Cross-references

- `.moai/reports/t1448/sync-audit-iter2.md` — F3(a), F3(c), F14, G-rename, the lock-order analysis.
- `.moai/worktrees/t1407/.moai/reports/t1407/verdict.md` and `progress.md` — option B and the arm (a)
  snapshot.
- `.claude/rules/moai/development/verification-completeness.md` § 2, § 2.1 — the two-cell adoption and
  RED-now cell content `acceptance.md` follows.
- `.claude/rules/moai/core/verification-claim-integrity.md` § 2.3 — the baseline-before-change ordering
  the milestones use (a RED commit that precedes each fix).
- `internal/kanban/board_store.go` — the lock's wait policy and its sizing derivation.
