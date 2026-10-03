---
id: SPEC-FACTORY-ATOMIC-LEASE-001
title: "Atomic lease across the queue store and the factory record — one critical section from the selection read to the claim, a bounded record wait, and a worktree step that no longer renames"
version: "0.3.1"
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
depends_on: [SPEC-TODO-AUTO-PICK-001, SPEC-TODO-CLASSIFY-DISPATCH-001, SPEC-FACTORY-RECORD-001]
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
- 0.2.0 — 2026-10-03 — iteration 2 repair after the independent plan-audit returned FAIL 0.79 against the
  Tier M threshold 0.80 (defects D1 to D18; the audit report is an untracked local file and is not cited
  as authority here — each repaired point below stands on its own measurement). Changes in this file:
  REQ-FAL-006 restated as an outcome, with the bare form's busy-record and busy-lock outcome decided
  (§H DL-1); REQ-FAL-007 carves out the existing foreign-worktree directory check; REQ-FAL-003 and
  REQ-FAL-010 qualified; REQ-FAL-004 reworded into a positive form; REQ-FAL-009 names the non-adopting
  reads and its two carve-outs; §A.2 O6 corrected (the iteration-1 probe used a runtime PRAGMA, not a
  DSN-carried timeout), O10 re-measured with its commands, O13 and O14 reordered; §F R3 and R4
  corrected, R1 to R12 put in order, R13 to R15 added; §H added for the decisions only the operator can
  take. Requirement count stays 13 and criterion count 14; nothing was added to either. The
  defect-by-defect map is in `progress.md` §E.1.
- 0.3.0 — 2026-10-03 — override-round repair after the independent plan-audit's iteration 2 returned FAIL
  0.81 against the Tier M threshold 0.80 on three must-fix defects and seven documentation defects (the
  audit's own ids PA2-M1 to PA2-M3 and PA2-N1 to PA2-N7, kept apart from the symptom ids M1–M5 of §A.1);
  the leader granted one extra repair round and one delta audit past the Tier M ceiling of 2. Changes in
  this file: **PA2-M1** — every record write also reconciles the drift log beside the
  record, and the log's file lock is waited for with no bound (§A.2 O15, measured: a claim returned after
  3.00–3.02 s against an 800 ms deadline), so REQ-FAL-006's bound was false on that path; the leader
  decided the design (§H DL-5): the lease claim's record writes try the log's lock without waiting and
  skip the reconciliation on contention. New REQ-FAL-014; REQ-FAL-006, -007, -009, -012 and -013 amended;
  §E adds SPEC-FACTORY-RECORD-001 REQ-FR-025 (narrowed for the claim's writes); §F R16; §H DL-7 records
  the scope choice (the bounded form applies to the claim's writes only). **PA2-M2** — AC-FAL-009's
  per-path table is rebuilt in `acceptance.md` (no change here). **PA2-M3** — REQ-FAL-003's second clause
  is narrowed to the arms that read the queue item's state; arm (a)'s re-lease of its own assigned card
  is a named residual (§F R17, §H DL-6). **PA2-N1** — DL-1's reasoning restated, §F R5 says a lane
  process ends on a busy-store result. **PA2-N2** — the AUTO-PICK §C.2 `raced` definition added to §E.
  **PA2-N3** — §F R13 and REQ-FAL-009 locate the legacy-directory adoption at store construction, before
  any lock, and name its race. **PA2-N6** — two absolute measurement sentences qualified (§A.1 M5 row).
  **PA2-N8** — plan D5 listed in §H as DL-4. **PA2-N12/N13** — probe ranges stated as ranges seen over
  named runs; REQ-FAL-007 says "file or process I/O". §H records DL-1 to DL-3 as decided (leader,
  2026-10-03). Requirement count 14, criterion count 15 (Tier M ceilings 16 / 16). The defect-by-defect
  map is in `progress.md` §E.1.
- 0.3.1 — 2026-10-03 — exception repair (iteration 4) after the independent plan-audit's iteration 3
  returned FAIL 0.83 against the Tier M threshold 0.80 on one must-fix defect (the audit's own id I3-M1) and
  three should-fix defects (I3-S1 to I3-S3); the leader approved an exception to "a second ceiling hit parks
  the card" for exactly four hunks and a re-read of those four hunks only (the leader's decision record,
  "Decision 2", an untracked local file named for provenance and not cited as authority). Changes in this
  file: **I3-S1** — REQ-FAL-014's "no entry reconciled twice" is replaced by what the claim's flow
  guarantees (it re-reads the log under the lock and reconciles an entry only if it is still unreconciled),
  and §F R16 names the cross-writer window the flow does not close. **I3-S2** — REQ-FAL-009's carve-outs
  are an enumerated list of four, with the skip stated once. **I3-S3** — §H DL-6's provenance is the
  leader's Decision 2; DL-7 is recorded as decided with the leader's added instruction, whose limit is the
  new §F R18. **I3-M1** — repaired in `acceptance.md` AC-FAL-015 (clauses restated at the lease function's
  return, two clauses added), with the mirror in `plan.md` §7; no requirement changed. Mechanical
  consequences: §B.5 and §F's introduction mention R18, and AC-FAL-012 counts it. Requirement count 14,
  criterion count 15 (Tier M ceilings 16 / 16). The finding-by-finding map is in `progress.md` §E.1.

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
| M2 | F3(c) | The operator holds a card after the lease path decided to take it and before the claim; the card is leased anyway. Same class in unnominated arm (c), which promotes and claims with no re-check. Bare arm (a), a lane re-leasing the card its own row assigns to it, never reads the queue item's state and is outside what this SPEC closes for this symptom (§F R17; ledger L20). | `queue=hold record=leased holder=lane-1 err=<nil>` on both arms |
| M3 | F14 | The compensation of a failed nominated claim cannot tell its own promotion from an operator's later fresh pick of the same card (ABA) and reverts the operator's pick to `queued`. | `queue=queued` after the operator's `picked` |
| M4 | t1407 option B | Arm (a), a lane leasing the serial card assigned to itself, reads the sibling snapshot and then claims non-atomically; two lanes with two assigned serial cards both lease. | both `leased` |
| M5 | rename | `factoryEnsureCardWorktree` renames the new branch with `git branch -m`; two lanes in one repository collide on git's shared reflog temp file, the verb errors after the lease succeeded, and the half-built directory is then refused by the foreign-worktree precheck that a re-lease runs (the precheck was observed to refuse; a full re-lease was not run). | Unforced: 44 of 160 concurrent iterations failed over eight runs, per-run 1 to 10 of 20 (5% to 50%; ledger L2 and L15 — the plan-audit's own run of 2 of 20 is cited from it, the other seven were measured by this SPEC's author). With the two lanes' renames forced to start together: 79 of 120 over six runs, per-run 11 to 15 of 20 (55% to 75%; L13). In the runs whose stranded-directory count was recorded here (18 + 13 + 79 = 110 failed steps) every failed step left a directory the foreign-worktree precheck refuses; one independent re-execution of the L13 command (its output is not in the ledger) reported 27 failed steps, 26 of them rename failures that left such a directory and one a creation failure that left none (the mechanism of O11), so that tally is a count over recorded runs and not a law |

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
  re-enters it (read: `retryFactoryBusy`, `internal/homestate/factory.go` lines 473–486, up to 100
  attempts 20 ms apart, each opening a fresh immediate transaction).
- **O6 — a busy timeout carried in the DSN lasts but does not bound the claim alone; the deadline
  does not bound it either, alone.** Two measurements, kept apart because the first was mis-described in
  the 0.1.0 text. (1) The iteration-1 probe opened its "low" handle through `OpenFactoryPath` (DSN
  5000 ms) and then set 200 ms with a runtime `PRAGMA busy_timeout`: with a 500 ms deadline it returned after
  0.68–0.71 s; with a 1 s deadline against a 7 s holder it ran 5.09–5.14 s and the handle read back 5000 ms,
  so a runtime PRAGMA did not survive a context-cancelled call (an inference: the connection was
  replaced). That probe never carried the timeout in the DSN. (2) The iteration-2 probe
  (`evidence/probe-homestate-dsn_test.go.txt`, ledger L14) opens the handle with `busy_timeout(200)`
  in the DSN, the form this SPEC adopts: the value read back 200 ms after the cancelled calls (it lasts);
  with NO context deadline the claim still waited out a 3 s holder whole (3.008 s, no error), because
  `retryFactoryBusy` re-enters; with an 800 ms deadline it returned after 0.98–1.03 s against a 3 s and a
  7 s holder, 182–227 ms past the deadline (the in-flight attempt's busy wait plus about 30 ms; ranges seen
  over the runs of L14 — an independent re-execution of that command printed 0.976–1.009 s, 176–209 ms past,
  against the 3 s holder and 0.934 s against the 7 s holder, and the ranges are not envelopes). So the
  deadline alone bounds nothing past the in-flight attempt and the DSN timeout alone bounds nothing past
  one attempt: the claim needs both, the deadline for the whole claim and the DSN timeout for the
  overshoot (plan D2). The plan-audit reported the same no-deadline observation with its own probe.
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
- **O10 — the two stores are consumed by different code.** At the patterns of ledger L17, 14 non-test Go
  files under `internal` and `cmd` open the factory record (`homestate.OpenFactory(` or
  `homestate.OpenFactoryPath(`), 10 construct a queue store (`kanban.NewBacklogStore(`, `todoStoreAt(`,
  `todoReadStoreAt(`), 23 in union and 1 in both (`internal/cli/factory_card.go`). The 0.1.0 text said
  "31 / 35 / 64 / 2"; the command behind those numbers was not recorded and they were not reproduced
  by the patterns of L17, so they are withdrawn and nothing in this SPEC rests on them. The new
  counts are a floor, not a census: they count direct call sites and not callers reached through
  helpers in other packages.
- **O11 — removing the rename does not remove the race.** `WorktreeCreator(name)` uses `name` as the
  directory leaf and as the branch (`session_worktree.go` `materializeSessionWorktree`,
  `gitWorktreeAddReal`), which is why the branch is renamed afterwards; a creator that took the branch
  separately would need an interface change. Two lanes creating worktrees concurrently in one repository
  with their FINAL `WT-` names failed in 2 of 280 iterations over seven runs (0, 1, 0, 1, 0, 0, 0 of
  40), with `fatal: failed to read .git/worktrees/wt-b/commondir` — `git worktree add` itself races a
  sibling `git worktree add`. The rename step failed in 44 of 160 unforced iterations (§A.1).
- **O12 — the worktree step has four production callers** (`factory_card.go`, `mcp_factory_card.go`,
  `factory_lane_relaunch.go`, `codex_launcher.go`), all through one function, so one lock around that
  function covers them.
- **O13 — SPEC-TODO-AUTO-PICK-001 is at its Tier M ceiling.** Its HISTORY states "requirement count
  stays 16 (Tier M ceiling)". An amendment that adds any requirement would force a tier-up.
- **O14 — serializing the whole step removes both failures, and it is cheap.** Two lanes' worktree steps
  (create plus rename) run one at a time under one lock: 0 of 80 iterations failed over four runs; one
  step took 0.42–0.69 s at best, 0.73–1.22 s at the median and 1.14–2.88 s at worst (the slower figures
  came from the later runs, made while the machine was busier; four runs, of which L7 displays two — the
  other two are not in the ledger). The probe used an
  in-process mutex, which orders the two git processes exactly as a cross-process lock would.
- **O15 — every record write also reconciles the drift log, and the log's lock wait has no bound.**
  `withCardTx`, which each of the claim's three writes goes through (`RecordPicked`, then `Transition`
  twice), reads `record-unavailable.jsonl` beside `factory.db` before the write and, when an unreconciled
  entry for the run exists, appends one `record.drift` event per entry inside the write transaction; after
  the commit it rewrites the log under `flock(LOCK_EX)` on `record-unavailable.jsonl.lock`, taken with no
  timeout and no context (read: `card_transition.go` `withCardTx`, `card_unavailable.go`
  `reconcileUnavailable` and `markRecordUnavailableReconciled`, `admission_lock_unix.go`
  `acquireAdmissionLock`). Measured (ledger L19, three runs): with one unreconciled entry for the run and
  that lock held by another party for 3 s, the claim's three writes under an 800 ms context deadline
  returned after 3.002–3.016 s, the second write failing with `context deadline exceeded` because its
  context expired while the first write waited, and the first write had reconciled the entry by waiting;
  the control with no lock held returned in 4.4–9.8 ms and reconciled the entry. Neither the busy timeout
  nor the deadline bounds this wait, so inside the section it would hold the queue lock for as long as the
  log's lock is held. The reconciliation's own work, uncontended, measured 0.002–0.028 s for one entry,
  0.018–0.100 s for 200 and 0.151–0.260 s for 2000 (L19, second test, three runs). Callers of
  `withCardTx` other than the claim — `factory stage` and its lease renewal, `factory complete`,
  `factory assign`, `factory decide`, the dispatch mirror (`factory_mirror.go`) and the lease path's own
  `RecordCardWorktree` after the section — hold no queue lock while they wait, and
  SPEC-FACTORY-RECORD-001 REQ-FR-025 states the log's contract for them: the next successful write for the
  run reconciles. The Windows counterpart of the lock primitive (`admission_lock_windows.go`,
  `LockFileEx` without `LOCKFILE_FAIL_IMMEDIATELY`) waits the same way (read, not run).

## §B Decisions

### B.1 Q1 — how the two stores become one step

The card names three options. All three were weighed against the measurements above.

| Option | Closes | Does not close | Measured cost |
|---|---|---|---|
| **1. Merge the two stores into one** | M1–M4 in principle | M5 | Both stores must move under one engine and one schema owner. 23 non-test files call the two APIs directly at the O10 patterns, more through helpers; `kanban` already imports `homestate`, so the merge picks a direction and inverts the other package's layering (O8). The factory schema is at version 5 with four sequential migrations; the queue has its own migration and downgrade machinery. SQLite `ATTACH` is not a shortcut: the SQLite documentation states that in WAL mode a transaction across attached databases is "atomic for each individual database, but … not atomic across all databases as a set" (sqlite.org/wal.html), and both stores are WAL (O2). Highest cost; touches two live operator stores. |
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
the end state (AC-FAL-002). **One arm is outside the "a hold committed before the section opened is
read" half:** bare arm (a), a lane re-leasing the card its own row assigns to it, never reads the queue
item's state, so it leases a card held beforehand whose row is assigned to the lane (ledger L20:
`queue=hold record=leased`, 3 of 3 runs; §F R17). It sits inside the section and so still orders any
*later* operator write after its lease.

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
- The repair is a bounded claim, and it takes two mechanisms because neither bounds the claim alone
  (O5, O6): the lease path opens its own record connection with a low busy timeout set in the DSN (it
  bounds the overshoot past the deadline) and bounds the whole claim by a deadline (it bounds the claim as
  a whole, three writes and the retry loop between them). The section's worst hold is therefore a named
  cap — the deadline plus the busy timeout — derived to be at most one third of the queue lock's wait
  budget (1.1 s at today's constants), the convention `slot_lease_cross_test.go` already uses for its
  stall-release timeout. On the cap the lease gives up, restores its promotion, and reports the outcome
  REQ-FAL-006 states for each form. A third wait lives inside every record write and neither mechanism
  reaches it: the drift log's file lock (O15). It is bounded by a different device — the claim's record
  writes try that lock without waiting and skip the reconciliation on contention (REQ-FAL-014, §H DL-5) —
  which adds no wait to the cap.
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

- **Avoid the rename — measured insufficient.** The rename collision (44 of 160 unforced iterations) is in `git branch -m`:
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

A NEW SPEC, citing the completed SPECs it supersedes or narrows through `depends_on`.

1. **It spans two completed SPECs and neither owns the lease path.** SPEC-TODO-CLASSIFY-DISPATCH-001
   owns serial exclusivity (REQ-TCD-008); SPEC-TODO-AUTO-PICK-001 owns the nominated lease and its
   compensation (REQ-TAU-004–006). The change is the section that surrounds both. A third completed
   SPEC is touched at one clause: SPEC-FACTORY-RECORD-001 REQ-FR-025 (the next successful write
   reconciles the drift log), narrowed for the claim's own writes by REQ-FAL-014 (§E).
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
lock, a stalled record, a skipped drift-log reconciliation (R16) and the record writes that keep waiting
for a held drift-log lock (R18) are named in §F and are not claimed closed.

## §C Requirements

Fourteen requirements, GEARS notation. Every `REQ-FAL-NNN` is traced by at least one `AC-FAL-NNN` in
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
  complete first and the operator write shall apply after it, within the queue lock's wait budget (a
  write that waits longer fails with the lock-held error, §F R6); **when** such a write committed before
  the critical section opened, an arm that reads the card's queue state — the nominated form and bare
  arms (b), (b2) and (c) — shall read it and shall not lease the card (the `held`, `dropped` or `raced`
  refusal for the nominated form, a skipped candidate for the bare arms). Bare arm (a), a lane
  re-leasing the card its own row assigns to it, reads no queue item state and is not covered by this
  second clause (§F R17). *(M2)*
- **REQ-FAL-004** (Event-detected): **When** a nominated claim fails, the compensation shall restore
  only a queue state written by the invocation that made the promotion, and shall run inside the
  critical section of the promotion it undoes, so no operator write can fall between the two. *(M3)*
- **REQ-FAL-005** (Event-driven): **When** two lanes each lease the serial card assigned to themselves
  at the same time and nothing is in flight, the lease path shall lease exactly one; the other lane
  shall read the first lane's live lease inside its own critical section and skip its card, which stays
  `assigned`. *(M4)*

### Module B — Bounded hold

- **REQ-FAL-006** (Event-detected): **When** the factory record's write lock is not obtained inside the
  critical section within the lease-claim wait cap, or the queue's lock is not obtained within its own
  wait budget, the lease path shall stop waiting, release whatever it holds, run the compensation the
  nominated lease already has (the queue item restored, inside the section; §F R4 names the cases in
  which an item stays `picked`), and report that no lease was taken by this invocation: the `raced`
  refusal for the nominated form, whose detail says the record or the queue lock was busy and does not
  say that another lane took the card; an error for the bare form, which ends the selection pass at once
  without re-selecting and is never reported as an empty queue (§H DL-1). The claim shall return no later than the lease-claim wait cap (to within the
  margin AC-FAL-007 states), the cap shall not exceed one third of the queue lock's wait budget, and the
  claim's drift-log reconciliation shall add no wait to the cap (REQ-FAL-014: it never waits for the
  log's lock; the work it does once it holds the lock is not bounded by the cap, §F R16).
  *(The mechanism that meets this is plan D2; the measurements behind it are §A.2 O5, O6 and O15.)*
- **REQ-FAL-007** (Ubiquitous): The critical section shall run no git subprocess, create no worktree,
  write nothing outside the queue store and the factory record, and perform no file or process I/O
  outside them other than two named exceptions: the existing foreign-worktree directory check
  (`factoryRefuseForeignWorktree`, a stat of the card's landing directory, which the claim and the
  nominated validation already run) and the claim's bounded reconciliation of the drift log that sits
  beside the factory record (`record-unavailable.jsonl`, its lock file and the temporary file renamed
  over it; REQ-FAL-014) — reading the environment and the clock is not I/O in this sense; it shall issue at most
  one queue promotion, at most one queue restore and at most three factory-record write transactions (the
  record step, the assign edge and the lease edge of a claim), none of them before the queue's lock is
  held; worktree creation and the card-worktree record write shall run after the lock is released.

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
  of twelve refusal tokens (none added), exit codes, the quota hold, the Codex backend skip, a lane's
  own re-lease of its assigned card (which reads no queue item state, §F R17), and the non-adopting
  reads the lease path makes today (its decision reads stay pure reads and migrate no queue layout).
  The carve-outs are exactly the four below and no others:
  1. a busy store ends the pass (REQ-FAL-006) — the factory record's write lock not obtained within the
     lease-claim wait cap, or the queue's lock not obtained within its wait budget: the nominated form
     reports the `raced` refusal and the bare form an error, which ends a lane process's run (§F R5);
  2. the detail of the `raced` refusal when the cause is a cap hit (REQ-FAL-006);
  3. the claim's skip of the drift-log reconciliation on contention (REQ-FAL-014); and
  4. the one-time adoption of a legacy queue state directory, which constructing the section's queue
     store performs, before any lock opens, on the bare arms that built only non-adopting stores before
     (§F R13).
- **REQ-FAL-010** (Ubiquitous): The lease path shall take the queue's lock before it writes the factory
  record and never the reverse; no code path shall wait for the queue's lock while holding a
  factory-record write transaction; and the non-test files of `internal/homestate` shall not depend on
  `internal/kanban` (a test file of that package may, as one does today).
- **REQ-FAL-011** (Ubiquitous): Every artifact of this card shall state the windows the critical
  section does not close (§F) and shall claim no atomicity beyond the critical section; a window it
  cannot close shall be named, not omitted.
- **REQ-FAL-012** (Ubiquitous): The change shall be surgical: lease-path code and the worktree step in
  `internal/cli`, the lock primitives (the queue's lock section and the worktree-step lock) in
  `internal/kanban`, the lease-path record open and the claim-scoped bounded reconciliation of the drift
  log in `internal/homestate` (the reconciliation step of the record-write transaction and a
  non-waiting form of its file-lock primitive), their tests, this SPEC's records, and any doctrine line
  that states the old residual risk — and nothing else; no queue verb, no queue or record schema, no
  transition-table row, no worktree-creator interface and no refusal token changes, and no change to
  how any record write other than the claim's reconciles the drift log.
- **REQ-FAL-013** (Ubiquitous): This SPEC shall name each clause of SPEC-TODO-AUTO-PICK-001,
  SPEC-TODO-CLASSIFY-DISPATCH-001 and SPEC-FACTORY-RECORD-001 that it supersedes or narrows (§E), shall
  leave the three completed bodies unedited in the plan phase, and shall record the supersession in
  each through its Amendments mechanism at the sync phase.

### Module E — The drift log

- **REQ-FAL-014** (Event-detected): **When** a record write of the lease claim — the record step, the
  assign edge or the lease edge, the three writes REQ-FAL-007 counts — finds an unreconciled entry for
  its run in the drift log beside the factory record (the log of failed dispatch-mirror writes that a
  later write reconciles, SPEC-FACTORY-RECORD-001 REQ-FR-025), the write shall try the log's lock
  without waiting; **when** that lock is held by another party, the write shall skip the reconciliation
  for that write — it shall append no `record.drift` event and mark no entry reconciled — and the
  entries shall stay unreconciled for a later write, the claim's own next write included; **when** the
  lock is obtained, the write shall re-read the log under the lock and reconcile an entry only if it is
  still unreconciled there (one `record.drift` event per such entry, committed with the write or not at
  all, and the entries marked reconciled once the write has committed). That re-read is the whole of
  the guarantee against a repeat: two writers that overlap between one's commit and its mark can still
  reconcile one entry twice, as every record write can today (§F R16). The skip shall apply
  to the claim's three record writes only: every other record write — `factory stage`, its lease
  renewal, `factory complete`, `factory assign`, `factory decide`, the dispatch mirror and the lease
  path's own card-worktree record write after the section — shall keep today's behavior, including
  waiting for the log's lock (AC-FAL-015 clause (vii) observes the last of these at the verb; §F R18).
  *(Plan-audit iteration 2 defect PA2-M1, not the symptom M1 of §A.1; the mechanism is plan D2, the
  measurement §A.2 O15, the choice of scope §H DL-7.)*

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
- The skip of the drift-log reconciliation (REQ-FAL-014) applies to the lease claim's three record
  writes only; these writers, and the lease path's own card-worktree record write, keep reconciling the
  drift log by waiting for its lock, as today.

### Out of Scope — arm (a) reading the queue item's state

- Making a lane's re-lease of the card its own row assigns to it (bare arm (a)) refuse a card whose queue
  item is held, dropped or otherwise not leasable. It would change a behavior REQ-FAL-009 preserves, and
  it is named for a follow-up card (§F R17, §H DL-6).

### Out of Scope — lane-loop handling of a busy-store result

- Making the lane loops (`factory_lane_relaunch.go`, `codex_launcher.go`) treat the busy-store error of
  REQ-FAL-006 as retryable instead of ending the lane's run (§F R5, §H DL-1). Neither file is within
  REQ-FAL-012's scope.

### Out of Scope — schemas, verbs, tokens and the transition table

- No queue or factory schema change, no new queue verb, no new refusal token (the `raced` token covers
  both a lost race and a bounded wait), and no row added to or removed from the card transition table.

### Out of Scope — worktree lifecycle beyond the lease's own step

- Worktree disposal, the `moai worktree` verbs and the session entry points (they do not take the new
  step lock, §F R11), the creator interface (unchanged — creating the branch with its final name was
  measured insufficient, §B.3), stranded-directory cleanup for failure causes other than the collision,
  two cards deriving one branch slug, and any automatic recovery of a card that was leased and then
  failed its worktree step (§F R15).

### Out of Scope — a stale-lock clear for the new locks on Windows

- Wiring `clearStaleLockAtPath` (the Windows-only primitive in `internal/kanban`) to the queue's lock or
  to the new worktree-step lock. The tree wires it to three other locks only (§F R3); a holder killed on
  Windows leaves the step lock's artifact until an operator removes it (§F R14). A follow-up card may
  wire it.

### Out of Scope — unrelated t1448 follow-ups

- The operator-decision cards t810, t1294 and t1383, the `kanban-dispatch-detail.md` budget split, the
  `moai todo --help` `--auto` wording, and the lane-session doctrine wording of the keep-set. They are
  separate cards named in the t1448 verdict.

## §E Supersession record

Nothing in any of the three completed SPECs below is edited by the plan phase. The sync phase records
each row below through the Amendments mechanism (`completed → in-progress` per
`spec-frontmatter-schema.md` § Status Enum, with `prior_completed_sha`), through a manager-spec
re-delegation. SPEC-TODO-AUTO-PICK-001 has no `## Amendments` section today (its HISTORY carries the
versions), so its amendment creates one; SPEC-FACTORY-RECORD-001 has none either (`grep -n Amendments`
over its `spec.md` printed nothing in the override round).

| Completed SPEC | Clause | Disposition once this SPEC is closed |
|---|---|---|
| SPEC-TODO-CLASSIFY-DISPATCH-001 | Amendments 0.4.0, "Known limitation (not closed by this amendment …)": the serial slot is read from a snapshot and the claim happens afterwards, two lanes can both lease a serial card; an atomic lease across both stores is future work | **Superseded for every lease arm** (REQ-FAL-002, -005). The paragraph's sentence that the leader named the follow-up t1458 without verifying it is also answered: the card exists. |
| SPEC-TODO-CLASSIFY-DISPATCH-001 | Same paragraph: "two lanes each leasing their OWN assigned serial card at the same moment is now an ordinary path … not closed here" | **Superseded** (REQ-FAL-005). |
| SPEC-TODO-CLASSIFY-DISPATCH-001 | Amendments 0.4.0, "Residual not repaired": an ownerless `picked` serial row and an `assigned` serial row wait on each other | **Untouched.** This SPEC does not repair or alter it. |
| SPEC-TODO-AUTO-PICK-001 | §B.8 "Non-atomicity": the record read and the queue write of the compensation are two stores with no shared lock, accepted and stated | **Superseded** (REQ-FAL-001, -004): the compensation runs inside the promotion's section. |
| SPEC-TODO-AUTO-PICK-001 | §G third bullet, "the record read and the queue write of the compensation are two stores without a shared lock" | **Superseded**, same reason. The post-`RecordPicked` residue half of that bullet stays (§F R4). |
| SPEC-TODO-AUTO-PICK-001 | REQ-TAU-005's compensation clause ("one queue write that acts only if the item is still `picked`") | **Strengthened, not contradicted:** the guard stays, and the write now cannot interleave. The F14 window (the guard cannot tell its own promotion from an operator's re-pick) is closed (REQ-FAL-004). |
| SPEC-TODO-AUTO-PICK-001 | REQ-TAU-006 (single ownership under concurrency) | **Unchanged and now also true for the serial slot** (REQ-FAL-002). |
| SPEC-TODO-AUTO-PICK-001 | §C.2 definition of the `raced` token: "the nominee was promoted or claimed by another lane first, or the version-checked edge reported a stale version" | **Amended:** the token also covers a bounded wait on a busy store — the record's write lock past the lease-claim wait cap, or the queue lock past its wait budget; the detail line distinguishes the two and in that case never says that another lane took the card (REQ-FAL-006). The closed set of twelve tokens is unchanged. |
| SPEC-FACTORY-RECORD-001 | REQ-FR-025, last clause: "the next successful factory-record write for that run shall append one `record.drift` event per unreconciled entry … and then mark the entry reconciled" | **Narrowed for the lease claim's three record writes** (REQ-FAL-014): such a write reconciles only when the drift log's lock is free, skips on contention, and a later write reconciles. Every other record write keeps the clause as written. |
| the t1448 run and sync records (`progress.md` §E.4, `CHANGELOG.md` limitations (a) and (b)) | The three windows and the `git branch -m` hazard stated as accepted residual risk | Run- and sync-owned; this card does not edit them. The sync phase of THIS card states, in its own CHANGELOG entry, which windows closed and which stay (REQ-FAL-011). |

## §F Gaps and residual risks

Stated plainly. A window not closed here is not closed; nothing in this SPEC's records may say
otherwise (REQ-FAL-011). R1 to R12 keep their 0.1.0 numbers; R13 to R15 were added in 0.2.0, R16 and
R17 in 0.3.0 and R18 in 0.3.1.

- **R1 — a lane process killed inside the section.** The lock is a `flock` released when the process dies
  (read, `board_lock_unix.go`), so nothing wedges, but a death between the queue promotion and the claim
  leaves the item `picked` with no row (or a `picked` unowned row). The next `next` re-adopts it through
  arm (b) or (b2). This is today's behavior for a crashed lane; it is not closed.
- **R2 — record writers that do not take the queue lock.** `factory assign` and the dispatch mirror read
  the queue without the lock and then write the record (`requireQueuePicked`, `writeFactoryAssignment`):
  an operator `unpick` between their read and their write is the M2 class of window in another verb. They
  and `stage`, `complete`, `decide` and lease renewal race the lease's claim at the record level, where
  the version-checked edges still arbitrate. Not closed; named for a follow-up card.
- **R3 — Windows: the locks have no stale clear that this SPEC can rely on.** On Unix the queue lock and
  the new step lock are `flock` on a descriptor the kernel releases at process exit. On Windows the lock
  file IS the lock: `acquireBoardLockImpl` (`board_lock_windows.go`, the `os.IsExist` case) returns
  `ErrBoardLockHeld` when the file exists and performs no stale clear on acquisition. The only clear code,
  `clearStaleLockAtPath` (`board_lock_clear_windows.go`), is wired to three locks — the board lock
  through the operator-invoked `ClearStaleBoardLock(root)`, the integration mutation lock and the
  slot-lease mutation lock — and a search of its callers found none for the queue's sibling lock
  (`backlog.lock`) or for the new step lock (ledger L17). `ClearStaleBoardLock` is a no-op on Unix by
  design. So a holder killed on Windows leaves a queue lock that this tree has no command to clear (an
  existing hazard, now held for longer inside the section) and, for the step lock, a new one (R14). The
  section's guarantee is cross-compile verified only on Windows (`GOOS=windows go build`), not exercised;
  the probes in this card ran on macOS. The 0.1.0 text said "bounded stale clear on Windows"; that
  described the board lock's operator clear and not an acquisition-time clear, and is withdrawn.
- **R4 — the claim is three record transactions, not one.** A claim that fails after `RecordPicked` leaves
  the row `picked` and unowned (SPEC-TODO-AUTO-PICK-001 §B.8, which stays accurate on this point). The
  section restores the queue item; the row stays, as today. **A second shape, named here because the
  cap makes it reachable:** when the lease-claim wait cap fires after the assign edge committed and
  before the lease edge, the row is `assigned` to this lane and the compensation deliberately leaves the
  queue item `picked` (the row is the lane's own, so the compensation reads no other holder and leaves
  the queue state alone — read, `factoryNominateCompensate`). The verb reports the `raced` refusal; its
  detail says the record was busy (REQ-FAL-006) and not that another lane took the card, which in this
  shape nobody did. The lane's next `factory next` leases the card through arm (a) (read:
  `factoryRecordRefusal` returns no refusal for a row `assigned` to the lane; to be observed in AC-FAL-007
  and not yet observed), and until then a serial card's `assigned` row holds the serial slot against every
  other lane's new take (the existing rule, pinned by
  `TestFactoryNextAssignedSerialCardStillHoldsSlotAgainstNewTakes`). **A third shape is the bare
  form's arm (c):** it promotes and then claims with no compensation today, so a claim that stops on the
  cap leaves the promoted item `picked` with at most an unowned `picked` row, and the next pass adopts
  it through arm (b) or (b2) (the recovery R1 already relies on). This SPEC adds no compensation to arm
  (c). None of the three shapes is closed.
- **R5 — a stalled record costs liveness, not correctness — and a lane process ends on it.** While the
  record's write lock is held longer than the cap, leases give up as `raced` (nominated form) or as an
  error (bare form), and `--wait` retries the former. The lane loops end on the bare form's error: the
  relaunch loop (`factory_lane_relaunch.go`, the `return err` on a lease error) and the Codex loop
  (`codex_launcher.go`, `return fmt.Errorf("codex lane: %w", err)`) both stop exactly as they stop on a
  not-leased result, so one transient stall that returns the error ends that lane's run. It ends
  visibly (an error return) where a not-leased stop would have ended it as if the queue were empty; before
  this change the same stall made the claim wait — up to 100 re-entries of a 5 s busy timeout, and a 7 s
  hold was waited out whole in the measurement of O5. This is a carve-out of REQ-FAL-009 (§H DL-1). The stall source is out of scope, and
  so is a retry in the loops (§D). Not closed.
- **R6 — multi-lane stall.** Under a persistent record stall each lane holds the lock up to one cap, so
  lanes queue behind one another; with more than about three waiting (budget 3.3 s over a cap of at most
  1.1 s) a later lane exhausts the queue lock's wait budget, and an operator queue write shares that
  exposure. The single-stall bound is tested (AC-FAL-007); the multi-lane case was not measured.
- **R7 — the stall evidence is constructed.** The probes hold a write transaction open to model a stall;
  no real WAL checkpoint stall or long git-in-transaction hold was observed (O9 is read, not measured).
- **R8 — the serialized step is measured at 0 of 80**, an upper bound of about 3.7% at 95% confidence,
  on one machine, with an in-process mutex standing in for the cross-process lock. AC-FAL-006 repeats it
  with the real lock and the real materializer.
- **R9 — the section's hold time on a real run is not measured yet.** B.2 gives an estimate from the
  component costs; AC-FAL-009 records the measured distribution in the run phase.
- **R10 — operator end state.** A hold that arrives while the section is open applies after the lease:
  `queue=hold record=leased` is reachable by that order. It is the ordering of two events, not a lease of
  a held card; the criteria test the ordering. (Through arm (a) the same end state is also reachable by a
  hold that preceded the section: R17.)
- **R11 — the step lock orders only this card's own lanes.** `moai worktree new`, the session entry
  points, and any other git process reading `.git/worktrees/*` in the same repository (the observed
  failure is a read of a sibling's half-written `commondir`) can still race a lane's creation. Not
  closed, and not claimed.
- **R12 — lane start-up latency.** Lanes that begin at the same instant queue behind the step lock at
  0.42–2.88 s per step; the wait budget is derived from the worst observed step (plan D3), not measured
  at ten lanes.
- **R13 — a stuck queue lock now stops every bare arm, every `--wait` poll takes the lock, and the
  section's store adopts a legacy directory.** Arms (a), (b) and (b2) read the queue purely and took no
  queue lock before; inside the section each takes it, so a queue lock held beyond its wait budget now
  ends a bare pass at once as an error on those arms too (REQ-FAL-006), where only arm (c) and the
  nominated promotion needed the lock before. A bare `--wait` re-checks every `factoryNextWaitInterval`
  (5 s) and each re-check takes the lock once per selection attempt, which is new lock traffic that was
  absent for arms (a), (b) and (b2). **The adoption is a side effect of constructing the section's
  store, not of opening its lock.** The store is built by `todoStoreAt(root)` — the adopting form arm (c)
  and the nominated promotion already used — whose path comes from `BacklogPathForRootAdopting`
  (`internal/cli/todo.go`), which calls `resolveStateDir(root, true)` (`internal/kanban/state_dir.go`):
  when only a legacy state directory exists it renames the whole directory (queue and session registries
  together, `relocateStateDir`) or moves the queue's artifacts into the global project directory
  (`relocateQueueArtifacts`). That runs at construction, BEFORE any lock is taken, so the lock cannot
  order it and REQ-FAL-007 keeps it outside the section: the section's store is constructed before
  `WithLock` is called (plan D1). Arms (a), (b) and (b2) built only non-adopting stores before, so for
  them the adoption is new; the decision reads inside the section stay non-adopting (`LoadPure`).
  **The race on that one-time adoption is not closed.** The loser of two concurrent first adoptions can
  find its rename refused because the winner already moved the directory; `resolveStateDir` then returns
  the legacy directory and the caller serves from it (the relocation-refused branch of `state_dir.go`).
  Read, not reproduced: the two lanes can then hold different lock files for that window, which is the
  exclusion the section is built on. A smaller change to REQ-FAL-009 exists — build the section's store
  non-adopting — and is not taken, because the store is chosen once before the arm is known and arm (c)'s
  promotion needs the adopting form. Not closed; AC-FAL-007 tests the bound, not the exposure.
- **R14 — a lane killed inside the worktree step.** On Unix the kernel releases the step lock, nothing
  wedges, and the half-built directory or branch it leaves is today's M5 residue (a re-lease is then
  refused by the foreign-worktree precheck — observed to refuse, a full re-lease not run). On Windows
  the step lock's file stays, no stale clear is wired for it (R3), and every later step waits the full
  step wait (plan D3) and then fails with the step's error, each time, until an operator removes the
  file by hand. No recovery exists for that case in this SPEC. Not closed.
- **R15 — a card leased and then failed at the worktree step stays leased without a worktree.** When the
  step fails after a successful lease — its lock wait elapsing, a creation failure, or today's rename
  collision — the row is `leased` to the lane with no recorded worktree, and a live lease on a serial
  card holds the serial slot. No caller re-runs the step for an already-leased card: the four callers of
  `factoryEnsureCardWorktree` (`factory_card.go`, `mcp_factory_card.go`, `factory_lane_relaunch.go` at the
  call after a NEW lease, `codex_launcher.go`) each pass it the card their lease just returned. The way
  out that exists is the lease running out and a transition request applying the expiry (T27,
  `applyLeaseExpiry`: the card returns to `assigned`, stage kept), after which arm (a) re-leases it and the
  step runs again; whether anything issues such a request for a card with no worktree was not
  established. This is the outcome of today's M5 failures too; this SPEC adds one more way to reach it and
  bounds that way by the step wait. Not closed.
- **R16 — a skipped drift-log reconciliation delays the cleanup it owes, and the work done once the lock
  is obtained is not bounded by the cap.** On contention the claim's record write skips the
  reconciliation (REQ-FAL-014): the `record.drift` events and the reconciled marks wait for a later
  write, so an entry can stay unreconciled across the claim and across every claim in which the log's
  lock happens to be held. The skip decision is a non-blocking try, so a momentary holder (an append by
  the dispatch mirror, another process's rewrite) also causes a skip. A lane whose last record write
  was the claim leaves its entries to another lane's or another verb's later write; if none follows,
  they stay unreconciled (SPEC-FACTORY-RECORD-001 REQ-FR-025 has them reported by `moai factory status`,
  unchanged here). Once the lock is obtained the reconciliation does its work inside the write — a file
  read, one event append per entry inside the write transaction (under the claim's deadline context), and
  a rewrite of the log after the commit (file I/O, outside that context) — measured uncontended at
  0.002–0.028 s for one entry, 0.018–0.100 s for 200 and 0.151–0.260 s for 2000 (L19, three runs): inside
  AC-FAL-007's 500 ms margin at those sizes and proportional to the log's length beyond them, which was
  not measured. While the claim's write holds the log's lock another process's append to the log waits
  for it, bounded by one write. **A second window, named here because REQ-FAL-014's re-read does not
  close it: an entry can be reconciled twice when two writers overlap.** The re-read under the lock
  (REQ-FAL-014) stops the claim from reconciling an entry that is already marked. It cannot stop two
  writers that overlap between one's commit and its mark: an ordinary record write reads the log without
  the lock, appends its events inside its transaction and marks the entries only after the commit
  (`withCardTx` and `reconcileUnavailable`, read), so a second writer that reads the log in that gap sees
  the entry still unreconciled and appends a second `record.drift` event for it. The claim's bounded
  flow can do the same when it obtains the lock after another writer's commit and before that writer's
  mark — its re-read still finds the entry unreconciled (read, not run). This is today's behavior on every
  record write and is not changed here. Measured by the plan-audit's iteration 3, not by this SPEC's
  author (audit-measured, probe A, a scratch probe that is not in the repository): two ordinary writes
  for one run, overlapping between the first's commit and its mark, appended 2 `record.drift` events for
  the one entry with `unreconciled-after=0`, in 3 of 3 runs. No window beyond these is claimed closed.
  Not closed.
- **R17 — arm (a) re-leases its own assigned card without reading the queue item's state.** A card whose
  row is `assigned` to the lane and whose queue item is `hold` is leased by bare arm (a): the arm never
  reads the item's state (`factory_card.go`, the loop over `cards` that precedes the `noNewCards`
  check; it reads the queue only for the card's classification). Wrapping it in the section orders the
  operator's *later* writes after the lease and does not make it refuse a card held earlier. Observed
  (ledger L20, 3 of 3 runs): `queue=hold record=leased holder=lane-1`. REQ-FAL-003's second clause is
  narrowed to the arms that read the queue state and REQ-FAL-009 preserves arm (a) as it is; reading the
  item in arm (a) would change preserved behavior and is out of scope (§D). The end state
  `queue=hold record=leased` is therefore reachable through arm (a) by an operator write that preceded
  the section as well as, through every arm, by one that followed it (R10). Not closed.
- **R18 — the record writes that keep waiting for a held drift-log lock: a lane's verb can still stall on
  it, only the queue is protected.** REQ-FAL-014 bounds the claim's three record writes and nothing else.
  Every other record write is an ordinary `withCardTx` write that waits for the log's lock, with no bound,
  when an unreconciled entry for its run exists: the lease path's own card-worktree record write after the
  section (`RecordCardWorktree`, run by every caller of `factoryEnsureCardWorktree` after a lease), `factory
  stage` and its lease renewal, `factory complete`, `factory assign`, `factory decide` and the dispatch
  mirror (the callers listed in §H DL-7). The card-worktree write was measured: with one unreconciled entry
  and the log's lock held 2 s it returned after 2.000–2.004 s with the entry reconciled afterwards
  (audit-measured, iteration 3's probe B: 2.004 s, 2.001 s, 2.003 s, 3 of 3; re-measured by this author,
  ledger L24: 2.0007 s, 2.0015 s, 2.0006 s, 3 of 3). The others share that code path and were not measured
  one by one (read). So a log lock held for a long time still holds a lane's `factory next` after its
  lease succeeded, and holds `stage`, `complete` and the renewal, while it no longer holds the queue or
  the claim. This SPEC does not protect the unattended lane loop from those waits: that belongs to cards
  t1480 and t1482 (the lane supervisor), as the leader decided (§H DL-7). Not closed.

## §G Cross-references

- `.moai/reports/t1448/sync-audit-iter2.md` — F3(a), F3(c), F14, G-rename, the lock-order analysis.
- `.moai/worktrees/t1407/.moai/reports/t1407/verdict.md` and `progress.md` — option B and the arm (a)
  snapshot.
- `.claude/rules/moai/development/verification-completeness.md` § 2, § 2.1 — the two-cell adoption and
  RED-now cell content `acceptance.md` follows.
- `.claude/rules/moai/core/verification-claim-integrity.md` § 2.3 — the baseline-before-change ordering
  the milestones use (a RED commit that precedes each fix).
- `internal/kanban/board_store.go` — the lock's wait policy and its sizing derivation.

## §H Decisions

Decisions that only the leader or the operator can overrule. Each states what was chosen, what it changes
against the preserved behavior of REQ-FAL-009, and the alternative. No requirement or criterion waits on
an answer; each is written to the choice below, so an overrule is an edit to the named clauses.

### Decided (leader, 2026-10-03)

DL-1 to DL-3 were put to the leader as open decisions in 0.2.0 and approved at the defaults written
below; DL-5 was decided in the override round; DL-6 and DL-7 were decided by the leader's second decision
(iteration 4). Provenance: the leader's decision record for card t1458 dated 2026-10-03 (an untracked
local file, named for provenance and not cited as authority — the text below states what was decided):
its first decision (items 1 to 5) for DL-1 to DL-3 and DL-5, and its "Decision 2" for DL-6 and DL-7.

- **DL-1 — the bare form on a busy record or a busy queue lock (audit D8). Decided: default.** Chosen:
  the pass ends at once, without re-selecting, and the verb reports an error (exit 1; the queue lock's
  own timeout error or an error naming the record wait; nothing on standard output; never the text `no
  card is available`). Why: today an arm (c) queue-lock timeout is already that error (the promotion's
  `Mutate` fails and `factoryNextLeaseOnceGated` returns it), so this is the smallest change to
  REQ-FAL-009's preserved behavior. The rejected design in 0.1.0 mapped the timeout to a re-selection:
  the bare path retries `factoryNextSelectionAttempts` (5, `factory_card.go` line 184) times and each
  attempt can wait the queue lock's whole budget (3.3 s), so a held lock would have kept the verb busy
  for about 5 × 3.3 s = 16.5 s and then printed `no card is available` with exit 3, which reports a lock
  stall as an empty queue (the retry loop is at `factory_card.go` lines 304–316). The adopted form waits
  one budget plus one retry wait of at most 50 ms (`boardLockWaitMax`) and then stops. **The true
  difference from the alternative is visibility, not whether a lane's run ends.** Alternative: exit 3
  with one stderr line in the shape the quota hold uses (at the loops' level, a not-leased result with a
  note). The lane loops stop on that result (`factory_lane_relaunch.go` lines 107–109,
  `codex_launcher.go` lines 992–994) — and they also stop on an error (`factory_lane_relaunch.go` lines
  76–79 and 103–106, `codex_launcher.go` lines 988–991), so the adopted design ends a lane's run on a transient
  stall too. What differs is how: the error ends it visibly, as an error return the lane's operator
  sees, where the not-leased result would end it silently as if the queue were empty (§F R5). No third
  option is offered here: making the loops treat a busy-store error as retryable is outside REQ-FAL-012's
  scope and is listed in §D.
  Clauses that follow the choice: REQ-FAL-006, AC-FAL-007, plan D2.
- **DL-2 — the nominated form on a cap hit or a queue-lock timeout. Decided: default.** Kept as 0.1.0
  had it: the `raced` refusal, exit 4, so `--wait` retries and a session that nominated can tell a busy
  store from a permanent refusal. Today both cases are plain errors (exit 1: `promote the nominee: …`, or
  the claim's error), so this is a deliberate carve-out of REQ-FAL-009 that REQ-FAL-006 states.
  Alternative: keep exit 1, which makes the two forms agree and drops the `--wait` retry on a stall.
  The `raced` token's widened meaning is recorded in §E.
- **DL-3 — a cross-process test helper is added to the package's test code (audit D6). Decided:
  default.** It costs the package one re-executing test (about 5.5 s per iteration at the pin, ledger
  L12) and is the only way a process-local mutex stands a chance of being caught (MU7). Alternative:
  accept that the criteria run the lanes as goroutines and say so — rejected, because REQ-FAL-001 is
  about a cross-process lock.
- **DL-5 — the drift-log reconciliation inside the claim is bounded, not named as a residual.** The
  leader's decision (replacing the author's default of naming the unbounded wait as a residual): the
  reconciliation inside the lease claim tries its file lock without waiting (or with a short deadline)
  and, on contention, skips the reconciliation, which a later write retries; REQ-FAL-006's time bound
  then holds unchanged. Reason given: an unbounded lock on every record write would block the
  factory-autonomy unattended lane loop that is starting. Written as REQ-FAL-014, with §F R16 for what
  stays open. The choice between the two forms of the bound is DL-7.
- **DL-6 — bare arm (a) is narrowed out of REQ-FAL-003's second clause.** Decided (leader, 2026-10-03,
  decision.md Decision 2): a hold that preceded the section still being leased through arm (a) is as
  intended, per REQ-FAL-009's "arm (a) unchanged". REQ-FAL-003's "read it and refuse" half binds the
  arms that read the queue item's state, and arm (a)'s re-lease of its own assigned card stays as it is,
  with the residual stated (§F R17). The serial-slot decision itself is changed by card t1480, not
  here (the leader's statement). Alternative not taken: arm (a) reads the item's state and refuses a
  held card — a behavior change to REQ-FAL-009, a candidate for a follow-up card.
- **DL-7 — the form and the scope of REQ-FAL-014.** Decided (leader, 2026-10-03, decision.md Decision 2):
  the smaller change stands — a non-blocking try, opted into by the claim's three writes only. The leader
  left two choices open in the first decision: non-blocking or a short deadline, and global or opt-in.
  Chosen: **non-blocking** — it adds nothing to the lease-claim
  wait cap, where a deadline would add to it and force `3 × cap ≤ budget` to be re-derived; the cost is
  that any momentary holder of the log's lock (an append in flight, another process's rewrite) causes a
  skip. **Opt-in per call, for the claim's three writes only** — not every record write. Evidence from
  reading the callers of the shared `withCardTx` (O15): `factory stage` and its lease renewal
  (`factory_card.go` 1244, 1248), `factory complete` (1402, 1422), `factory assign` (1739, 1744),
  `factory decide` (2059), the dispatch mirror (`factory_mirror.go` 47, 54) and the lease path's own
  `RecordCardWorktree` (428) hold no queue lock while they wait, so the starvation REQ-FAL-006 prevents
  does not arise for them; REQ-FR-025 makes reconciliation the duty of the next successful write, and the
  mirror's failed writes are what the log exists to reconcile; a global skip would change reconciliation
  for all of those callers to bound a wait only the claim suffers inside the queue lock. The scope is
  per call and not per connection because the card-worktree record write after the section may share the
  lease path's connection and must keep waiting. **The leader's added instruction:** the writes that stay
  unbounded — the lease path's own `RecordCardWorktree` and `factory stage`, `complete` and lease
  renewal — keep waiting for a held drift-log lock; protecting the unattended lane loop from those waits
  belongs to cards t1480 and t1482 (the lane supervisor), and that limit is written into the SPEC's
  residual risks (§F R18). Overrule: an edit to REQ-FAL-014, plan D2, AC-FAL-015 and mutants MU16 and
  MU20.

### Open decisions for the leader

- **DL-4 — reopening three completed SPECs through the Amendments mechanism (plan D5, audit PA2-N8).**
  Recording the supersession at sync takes SPEC-TODO-AUTO-PICK-001, SPEC-TODO-CLASSIFY-DISPATCH-001 and
  — since the drift-log change narrows REQ-FR-025 — SPEC-FACTORY-RECORD-001 through `completed →
  in-progress → completed` once each, in one manager-spec delegation (§E, REQ-FAL-013, AC-FAL-014).
  This is the largest governance cost in the SPEC and the operator may veto it. Default: record it.
  Alternative: a pointer in this SPEC only, which leaves readers of the three completed SPECs reading
  clauses this SPEC supersedes or narrows as if open. No requirement or criterion waits on the answer;
  a veto edits REQ-FAL-013 and AC-FAL-014.
