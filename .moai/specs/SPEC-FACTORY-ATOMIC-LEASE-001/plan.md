# SPEC-FACTORY-ATOMIC-LEASE-001 — Plan

Tier M. Card t1458, plan-start HEAD `2de0a2cb613b04765a1554f86685a3b48e0be806` (branch
`WT-atomic-lease`, worktree `.claude/worktrees/t1458`). Version 0.2.0 of the plan (iteration 2 repair
after the independent plan-audit; the defect map is in `progress.md` §E.1). Sections and milestones are
ordered by **decision reversibility** — the decisions most likely to change come first (the lock
primitive's interface, the bounded-claim semantics a lane sees, the creator interface shared by four
callers), mechanical edits last. No time estimates; priority labels and ordering only.

**Decisions most likely to change** are D1 to D3 below: the primitive's shape, the bounded-claim
semantics, and the worktree-step lock. D4 and D5 are structural and bookkeeping.

**Two id spaces, kept apart on purpose.** The card's five scope items keep their ids **M1–M5**
(spec.md §A.1: M1 F3(a), M2 F3(c), M3 F14, M4 t1407 option B, M5 rename) and are what the
requirements and criteria trace to. The plan's ordered work steps are **WM1–WM6**, so "M3" never means
two things.

## 1. Approach in one paragraph

Give the lease path one critical section: a new primitive in `internal/kanban` holds the queue's
existing cross-process lock across a caller-supplied sequence, and `internal/cli/factory_card.go` runs
the whole selection pass (bare form, all four arms) or the nominated lease (validate, promote, seam,
claim, compensate) inside it, so the decision and the claim share an exclusion with every queue writer
and the compensation can no longer be separated from its promotion. Because the section now contains a
record write whose wait is not bounded by a context deadline, the lease path opens its own record
connection with a low busy timeout in its DSN **and** bounds the whole claim by a deadline, the two
summing to a cap derived from the queue lock's wait budget (both are needed — D2); on the cap the
nominated form restores its promotion and reports the existing `raced` refusal and the bare form ends
its pass with an error. Separately, serialize the post-lease worktree step (creation plus the
`git branch -m` rename, in the one function all four callers share) behind a dedicated cross-process
lock with a bounded wait — measured to remove both observed git races, which dropping the rename alone
does not. Everything is TDD-first, each milestone's RED commit precedes its fix commit so the commit
graph, not a commit message, witnesses the order (`verification-claim-integrity.md` §2.3).

## 2. Decisions most likely to change (most reversible first)

**D1 — the primitive's shape (WM2).** `BacklogStore.WithLock(fn func(*LockedBacklog) error) error`:
acquire the lock through the existing `acquireLock` (same wait policy, same timeout error), refuse a
relocated queue exactly as `Mutate` does, run `fn`, release and join the release error as `Mutate`
does. `LockedBacklog` offers **`LoadPure()`** and `Mutate(func(*BacklogRecord) error)` (today's `Mutate`
body without the lock acquisition). It offers no adopting read. Iteration 1 of this plan wrote
`Load()` ("a fresh read under the lock"); `BacklogStore.Load` is the adopting read, and the lease path
reads with `LoadPure` today (`factory_card.go` lines 470, 535, 810 and inside `queueItemState`, line
1647), which never adopts, migrates or relocates (`backlog_store.go`, `LoadPure`). REQ-FAL-009 ("behave as
before") therefore requires the locked handle's read to be `LoadPure`; `queueItemState(root, cardID)`,
which also searches the archive, is replaced inside the section by the same search over the record the
handle returned, so the section builds no second store value. The store the section locks is
`todoStoreAt(root)` (the adopting path the promotion uses today); spec §F R13 states what that adds for
the bare arms that never opened it. `Mutate` itself becomes `WithLock` around one `LockedBacklog.Mutate`,
so its behavior is byte-preserved. The lock is a non-reentrant `flock` on a separate descriptor: calling
the *public* `Mutate` from inside the section contends with the section's own descriptor and waits out
the budget. That is a trap, not a feature, and §5 lists the existing tests it breaks. An exported read
accessor for the wait budget (`LockWaitBudget()`), so the claim cap is derived from the budget and never
copied from it. A test pins the non-adopting read (`TestLockedBacklogLoadIsPure`, WM2: a queue in the
legacy layout shows no layout change after a `WithLock` that only calls `LoadPure`; mutant MU8).

**D2 — the bounded claim (WM4).** Two mechanisms, **both required**, and the reason is measured:

- the **claim deadline** — a context deadline over `RecordPicked` plus the two transitions taken as
  one claim. It is the only thing that bounds the claim as a whole, because `retryFactoryBusy`
  (`internal/homestate/factory.go` lines 473–486) re-enters a busy transaction up to 100 times: with a
  200 ms busy timeout in the DSN and no deadline, a claim waited out a 3 s holder whole (ledger L14:
  `3.0079325s`, no error);
- the **claim busy timeout** — a `busy_timeout` carried **in the lease record connection's DSN**, which
  bounds the overshoot past the deadline: with an 800 ms deadline and a 200 ms DSN timeout the claim
  returned after 0.98–1.03 s, 182–227 ms past the deadline, against both a 3 s and a 7 s holder (L14). A
  runtime `PRAGMA` does not substitute: it did not survive a context-cancelled call (spec §A.2 O6, the
  iteration-1 measurement, which used a runtime PRAGMA and never a DSN-carried value). Without the
  timeout the deadline's overshoot is the default 5000 ms (L5: a 500 ms deadline returned after
  `3.107s`).

Named constants in `factory_card.go`, each defined once here: `factoryLeaseClaimDeadline` (starting value
800 ms), `factoryLeaseClaimBusyTimeout` (200 ms) and `factoryLeaseClaimWaitCap`, which **is** their sum
(1.0 s at the starting values, so the claim returns at about the cap, within about 30 ms of it in L14).
The derivation rule is `3 × factoryLeaseClaimWaitCap ≤ kanban.LockWaitBudget()` (3.3 s today, so the cap
may not exceed 1.1 s); `TestFactoryLeaseCapWithinBoardBudget` reads the accessor, never a literal. The
section's worst hold is the cap plus one promotion and one restore (two mutation costs, 33 ms each at the
lock's sizing figure) plus the decision reads, not the cap alone; AC-FAL-009 clause (iv) records the
measured hold. The lease path opens its record connection through a new `homestate` open variant that
puts the busy timeout in the DSN (sharing the existing DSN builder). The open happens **before** the
section, so a slow open never lengthens the hold (a stalled record can delay the open itself — its
initialization runs under a 10 s context today too — and that is not part of the hold and is unchanged).
Reads inside the section (`ListCards`, `LoadCard`) run on the parent context: a WAL reader is not blocked
by a writer (spec §A.2 O4).

Outcomes (spec REQ-FAL-006, §H DL-1/DL-2): on the cap or on a queue-lock timeout the **nominated form**
runs its existing compensation inside the section and returns the `raced` refusal (so `--wait` retries)
with a detail that says the record or the queue lock was busy; the **bare form** ends its pass at once,
with no re-selection, and returns an error (exit 1; `factoryNextLeaseOnceGated` returns it, the verb
prints nothing on standard output). The bare form's re-selection loop (`factoryNextSelectionAttempts`
= 5) therefore stays reserved for the genuine races it serves; had a lock timeout been mapped to a
re-selection, a held queue lock would have kept the verb busy for 5 × 3.3 s = 16.5 s before it printed
`no card is available` (spec §H DL-1). The bare form's worst case is one wait budget plus one retry wait
of at most 50 ms (`boardLockWaitMax`). **No new refusal token** (the closed set of twelve is pinned by
SPEC-TODO-AUTO-PICK-001). Reversible: a thirteenth token is a later, explicit amendment of that set.

**D3 — the worktree-step lock (WM5).** `factoryEnsureCardWorktree` takes a dedicated cross-process lock
around the creator call and the `git branch -m` rename together, and releases it before
`RecordCardWorktree`. The creator interface does not change (spec §B.3: creating the branch with its
final name failed 2 of 280 concurrent creations, so it is not a fix). Substrate: reuse the board lock's
path-parameterized `flock` / atomic-create implementation (`acquireBoardLockImpl`) behind a small exported
helper in `internal/kanban` that polls with the same retry policy (`boardLockRetryWait`) against a
**separate** lock file, `<root>/.moai/state/factory-worktree-step.lock` — not the queue's lock, not
`board.lock` (the leader's board writes), and not `moai slot` (a named lease with state and an audit trail,
too heavy for a sub-second critical section). No existing generic named lock exists to reuse
(`internal/spec/lock.go` is scoped to a SPEC id). Wait budget, derived like the queue lock's and stated as a
sizing heuristic and not a worst-case bound: lanes (10) × the worst observed step (2.9 s, from a busy
machine, ledger L7) × headroom (2) ≈ 58 s, taken as 60 s (`factoryWorktreeStepWait`) — a named constant
with a derivation test. The wait is generous on purpose: it bounds a failure, and the usual wait for the
last of N lanes is about N × the median step (0.73–1.22 s). On timeout the step fails with its existing
error shape and creates nothing.

Failure outcomes, stated so nothing about them is implied:

- **A lane killed inside the step.** Unix: the kernel releases the `flock`, nothing wedges; the half-built
  directory or branch it leaves is today's M5 residue (spec §F R14). Windows: the lock file stays, no
  stale clear is wired for it (spec §F R3), and every later step waits the full 60 s and fails, each time,
  until an operator removes the file. No recovery for that case is planned; wiring
  `clearStaleLockAtPath` to the new lock is out of scope (spec §D).
- **A step-lock timeout after a successful lease.** The card is `leased` to the lane with no recorded
  worktree (a serial card then holds the serial slot until its lease runs out) and the verb errors. No
  caller re-runs the step for an already-leased card: `factory_lane_relaunch.go` line 110 calls
  `factoryEnsureCardWorktree` on the card its own lease just returned (the 0.1.0 pointer to it as a
  recovery path was imprecise and is withdrawn). The way out is lease expiry plus a transition request
  that applies it, then arm (a); spec §F R15 states the gap, including that whether anything issues that
  request was not established.

Reversible: if the run phase shows the lock's start-up cost matters, final-name creation under the same
lock is the next step and needs the creator interface change this plan declines.

**D4 — where the section's boundaries sit (WM3).** Inside: the pass's reads, the keep-set and slot
decisions, `Mutate` promotion, the seams, the claim, the compensation. Outside: `OpenFactory`, the quota
latch, `factoryEnsureCardWorktree` and everything after it (REQ-FAL-007). The only filesystem read the
section makes outside the two stores is `factoryRefuseForeignWorktree` (a stat of the card's landing
directory), which the claim and `factoryNextValidate` already run; REQ-FAL-007 carves it out and
AC-FAL-009 pins the allowed set. Arm (b)'s `queueItemState` read and arm (b2)'s queue read go through the
locked handle's `LoadPure()`; no code inside the section calls the public `Mutate` or `Add`. A guard test
(`TestFactoryLeaseSectionRejectsNestedMutate`) calls `Mutate` from inside a section and asserts the
timeout error, so a future edit that does is caught (§6 R-B).

**D5 — the supersession records (WM6 + sync).** The plan phase edits neither completed SPEC. At sync,
manager-spec adds one Amendments row to each (§E of the spec). Accepting the `completed → in-progress →
completed` round trip on both is a cost the operator may veto; the alternative (a pointer only in this
SPEC) leaves readers of the old SPECs reading "not closed". Decision recorded here so it can be
overruled.

## 3. What was observed at plan time

All of it is in spec.md §A.2 with the commands in `acceptance.md`'s evidence ledger and the probes under
`evidence/` (committed as `.go.txt` so the build never sees them, mapped in with `go test -overlay`).
The five symptoms reproduced RED on the unmodified tree; the cost and stall figures that size the
critical section were measured; several hypotheses were refuted or corrected by measurement and are kept
in the record rather than smoothed over:

- the naive "hold the lock across the claim" form starves other queue writers under a stalled record
  (O5), so the cap in D2 is a requirement and not polish;
- the t1448 audit's 1.65 s lock budget is stale — the tree derives 3.3 s (O1);
- the expected M5 fix (create the branch with its final name) is not a fix: 2 of 280 concurrent
  creations still failed in git's own `worktree add` (O11), where serializing the step failed 0 of 80
  (O14);
- a hang placed after the arm-(c) decision does **not** reproduce the two-serial-card breach (the
  second lane adopts the first lane's promoted card through arm (b2) and loses the version-checked
  edge); the breach needs both snapshots to precede both promotions. The RED test for arm (c)
  therefore holds lanes at the pass entry, before the snapshot;
- (iteration 2) the busy timeout carried in the DSN lasts but bounds only the overshoot, not the claim
  (L14), which is why D2 states both mechanisms; the iteration-1 probe never measured a DSN-carried value;
- (iteration 2) the iteration-1 RED probes for the operator-write criteria failed on an end state the fix
  leaves reachable; clause (i) itself was observed red separately (L11) with a positive control;
- (iteration 2) the M5 collision's per-iteration rate moved between 5% and 50% unforced, so the criterion
  rests on a forced overlap whose occurrence is deterministic (L13), not on that rate.

## 4. Milestones

### WM1 — RED first, for every scope item (Priority High)

Nothing here changes behavior. Three commits, in this order, so the commit graph — not a commit message —
witnesses the order (`verification-claim-integrity.md` §2.3):

1. **The seam-and-stub commit.** It must compile on its own and carry every symbol a WM1 or WM2 test
   names, because one undefined symbol in a `_test.go` file fails the whole package's test build, and
   every test in it — the preservation family included — would then be red for the wrong reason:
   - `internal/cli`: the package variable `factoryLeaseBeforeClaim func(arm, cardID string) error` (inert
     default) called in arms (a), (b), (b2) and (c) immediately before the claim; the pass-entry hook
     `factoryLeaseAtPassEntry func()` (inert); the constants `factoryLeaseClaimDeadline`,
     `factoryLeaseClaimBusyTimeout`, `factoryLeaseClaimWaitCap` and `factoryWorktreeStepWait` with their
     starting values (unused by production code at this commit).
   - `internal/kanban`: compile-only stubs, each a signature with the body
     `panic("not implemented: SPEC-FACTORY-ATOMIC-LEASE-001 WM2")` — `LockWaitBudget() time.Duration`,
     `(*BacklogStore).WithLock`, the `LockedBacklog` type with `LoadPure` and `Mutate`, and the
     step-lock helper (`AcquireFactoryStepLock(root string, wait time.Duration) (release func() error,
     err error)`). A test that reaches one fails with the panic text, which is the stated reason.
   - Not in this commit: any behavior, and any test.
2. **The baseline commit.** On the seam-and-stub tree, take the AC-FAL-010 baseline — the family's
   `-list` dump and the `--- PASS` count of one `-v` run of it (68 names and 68 passes were measured at
   plan time on the pin, ledger L9; the run phase re-measures on this tree and compares name for name) —
   and record it in `progress.md` §E.2 in a commit of its own, before any commit that changes behavior.
3. **The RED commit.** The tests below plus the harnesses, failing on the commit before them for the
   stated reason, and the tolerant harness.

- **Tolerant seam harness `nmRaceAtSeamTolerant`.** The existing `nmRaceAtSeam` waits for every lane to
  arrive at the seam before releasing any. Inside an exclusive section two lanes cannot both be at the
  seam, so it would hang for the wrong reason on the fixed tree. The tolerant form releases the lanes
  when all have arrived **or** a grace window after the last arrival has passed. The grace (750 ms) is a
  heuristic, not derived: it must exceed the time the first lane needs to leave the seam, which was
  milliseconds on the unfixed tree. On the unfixed tree both lanes arrive within milliseconds and are
  released together (RED, as observed); on the fixed tree the second lane is blocked at the lock, the
  window elapses, and it proceeds in order.
- **Operator-write probe `nmOperatorWriteDuring`.** From inside a seam, start the operator write in a
  goroutine and wait up to 400 ms for it; report whether it completed inside the section. Unfixed: it
  completes at once. Fixed: it stays blocked until the verb returns, then applies. The probe is
  accompanied by its positive control (a write started under a held queue lock stays pending across
  the same window — ledger L11 observed it) so the window is known to discriminate.
- **Ordered event log for the step.** A creator wrapper (around the real materializer) and a `git` shim
  first on `PATH` append to one file opened for append: creator-enter, creator-exit, and the shim's
  start and end of each `git` call. The log's line order is a total order, so "no creator-enter of lane B
  between lane A's creator-enter and the end of lane A's `branch -m`" is checkable without timestamps.
  The same shim, with a marker line from the creator stub, pins the no-git-in-the-section clause
  (AC-FAL-009 clause (iii)).
- **Cross-process lane helper (what it needs).** The package already re-executes its own test binary for
  helpers (`exec.Command(os.Args[0], "-test.run=^<Helper>$")`, e.g. `gate_lock_cli_test.go`), and the
  existing `TestMain` (`main_test.go`) runs first in the child, so **no new `TestMain` is needed**. The
  probe that proved the form (`evidence/probe-xproc_test.go.txt`, ledger L12) shows what the helper does
  need: (a) the child re-establishes `CLAUDE_PROJECT_DIR`, the `MOAI_HOME` the parent chose and the lane
  environment inside the helper test function, because `TestMain` unsets the first and clears the factory
  ambient family; (b) a file rendezvous in a shared temp directory in place of the channels, because a
  channel cannot cross a process boundary (each child touches a file and waits until both exist or a
  grace passes); (c) the worktree creator stub rebuilt in the child from repository paths the parent
  prepared; (d) results returned on the child's standard output as one tagged line, because `t.Logf` of
  a child is not visible to the parent. It costs about 5.5 s per iteration at the pin (ledger L12, 10
  iterations in 55.7 s), so the criterion's `-count` is 10, not 20 (AC-FAL-001). Spec §H DL-3.
- New tests (names the criteria cite): `TestFactoryLeaseSerialDistinctNomineesExactlyOne` (M1,
  nominated), `TestFactoryLeaseSerialBareLanesExactlyOne` (M1, arm (c), held at the pass entry),
  `TestFactoryLeaseSerialCrossProcessExactlyOne` (M1, two processes), `TestFactoryLeaseOperatorWriteWaitsForSection`
  (M2, nominated), `TestFactoryLeaseArmCOperatorHold` (M2, arm (c)), `TestFactoryLeaseCompensationKeepsOperatorPick`
  (M3), `TestFactoryLeaseOwnAssignedSerialSiblingsExactlyOne` (M4),
  `TestFactoryEnsureCardWorktreeConcurrentRealMaterializer` (M5, real materializer, no stub, forced
  overlap, ordered event log), `TestFactoryEnsureCardWorktreeStepLockBounded` and
  `TestFactoryWorktreeStepWaitDerivation` (M5, the step lock's bounded wait and its derivation),
  `TestFactoryLeaseRecordStallBounded`, `TestFactoryLeaseMidClaimStallBounded` and
  `TestFactoryLeaseQueueLockStallBounded` (the bound, the T2-landed shape, the queue-lock timeout in both
  forms), `TestFactoryLeaseCapWithinBoardBudget` (derivation), `TestFactoryLeaseSectionExcludesWorktreeStep`,
  `TestFactoryLeaseSectionRecordWritesPerArm` and `TestFactoryLeaseSectionAllowedSet` (REQ-FAL-007), and
  the layering and re-entrancy guards `TestHomestateDoesNotImportKanban` (in `internal/homestate`; the
  import guard covers **non-test files only** — `go list -deps` of the package without `-test`, whose
  kanban count is 0 today; with `-test` it is 1, through `internal/homestate/temp_parity_test.go`
  line 10, ledger L16, so a guard written the natural way would be red on arrival) and
  `TestFactoryLeaseSectionRejectsNestedMutate`.
- Record the verbatim RED output of each in `progress.md` §E.2 and the **swept count** of each selector
  (every selector in `acceptance.md` sweeps 0 tests at the pin, ledger L10; the run phase records the
  count at the RED commit and again at each green).

### WM2 — the primitive (Priority High)

`internal/kanban`: `WithLock`, `LockedBacklog` (`LoadPure`, `Mutate`), `LockWaitBudget()`, replacing the
WM1 stubs. Tests first: two mutations inside one `WithLock` see each other's writes; the lock is held until
`fn` returns (a contender blocks, then proceeds); it is released on an error return and on a panic; a
relocated queue is refused as `Mutate` refuses it; `TestLockedBacklogLoadIsPure`; `Mutate`'s existing tests
stay green unchanged. `GOOS=windows GOARCH=amd64 go build ./...`.

### WM3 — the lease path inside the section (Priority High)

`factory_card.go`: `factoryNextNominate` runs validate, promotion, seam, claim and compensation inside
`WithLock`; `factoryNextSelectAndLease` runs inside `WithLock` per attempt, its decision reads taken
under the lock; the compensation takes the locked handle. The snapshot-reuse paths (`nom.serialHeld`
carried from the pre-lock validation into the in-lock check) are deleted: the in-lock check reads the
record itself. Flips AC-FAL-001 to -005 and AC-FAL-009. Move the existing tests §5 lists.

### WM4 — the bounded claim (Priority High; lands with WM3 — WM3 alone puts an unbounded record wait
inside the queue lock and must not be integrated without it)

`homestate`: the busy-timeout open variant sharing the existing DSN builder. `factory_card.go`: the cap
constants become live, the claim deadline, the outcome mapping of D2 (nominated: `raced` with the
busy-store detail; bare: stop at once with an error). Flips AC-FAL-007 and AC-FAL-008.

### WM5 — the worktree step (Priority Medium)

The lock helper in `internal/kanban` (tests first: two contenders serialize; release on error and
panic; the wait budget bounds a blocked contender; the Windows build), replacing the WM1 stub, then
`factoryEnsureCardWorktree` taking it around creation and rename and releasing it before the record
write. `nmIsolatedWorktrees` needs no change (the stub still replaces the creator; the lock is taken
around it). Flips AC-FAL-006.

### WM6 — closure (Priority Medium)

Preservation run against the WM1 baseline; the mutants of §7 each executed with `go test -overlay` on a
mutated copy kept in scratch (repository untouched), the failing top-level test recorded; the measured
hold-time distribution of the section on a real lease (10 sequential and 2/4/10 concurrent lanes) in
`progress.md` §E.2 (REQ-FAL-011, R9); the layering test; the doctrine sweep (`grep` of `.claude/` and
`internal/template/templates/` for text stating the old residual risk — none was found at plan time);
the surgical-diff measurement. Sync phase: the §E supersession rows, and this card's CHANGELOG entry
stating which windows closed and which (spec §F) remain.

## 5. Existing tests that must move (listed, so none is a surprise)

All in `internal/cli`. The assertions each test makes are unchanged; what moves is how it reaches the
interleaving, because the section removes interleavings the old tests built by hand.

| Test | Why it moves |
|---|---|
| `nmRaceAtSeam` (helper) and its two users `TestFactoryNextNominateConcurrentLanes`, `TestFactoryNextNominateSameCardExactlyOne` | The helper requires every lane to reach the seam before any is released; inside the section only one can. Switch to `nmRaceAtSeamTolerant`. Assertions (each lane holds its own card; same card → exactly one holder and one refusal) are unchanged. |
| `TestFactoryNextNominateCompensationFailure/item-moved` | Its seam writes the queue synchronously with `nmSetState` → the public `Mutate` from inside the section waits out the budget and errors. Becomes `nmOperatorWriteDuring`: the operator's `dropped` applies after the verb, and the compensation must not have overwritten it. |
| `TestFactoryNextNominateCompensateRechecksRecord` (table + the lock-wait subtest) | Calls `factoryNominateCompensate(ctx, db, root, …)` directly; the signature gains the locked handle. The table keeps its six cases against the new signature. The lock-wait subtest models a lease landing while the compensation waits for the queue lock, a window the section removes — it is replaced by AC-FAL-004's test and the removal is recorded, not silent. |
| `TestFactoryNextNominatePromoteThenLose`, `TestFactoryNextNominateClaimRefusedRollsBack` | Seam writes touch only the factory record or return an error; they stay as they are and are the control that the seam still works inside the section. |
| `factory_classify_test.go:312` | Calls `factoryNextRecordAndClaim` directly; its signature is kept. |

## 6. Risks

- **R-A — the tolerant harness weakens what AC-TAU-002 asserts.** Mitigation: assertions untouched,
  and mutant MU2 (release the lock before the claim) must fail the moved tests.
- **R-B — self-contention.** A seam or helper that calls the public `Mutate` inside the section stalls
  for the full budget and errors. Mitigation: the guard test of D4 and the list in §5.
- **R-C — Windows.** The lock is an atomic-create file there with no stale clear this SPEC can rely on;
  compile-verified only (spec §F R3, R14).
- **R-D — the section is slower than estimated.** The estimate is one extra queue mutation plus the
  claim (spec B.2); WM6 measures it, and the cap bounds the worst case at one third of the budget plus a
  promotion and a restore.
- **R-E — multi-lane stall.** Bounded per lease, not across lanes (spec §F R6); not measured.
- **R-F — the step lock covers this card's lanes only.** Four production callers share one function
  (O12) and all are covered; `moai worktree new`, session entry points and other git readers are not
  (spec §F R11). The AC-FAL-006 test calls the shared function directly.
- **R-F2 — start-up latency.** Lanes starting together queue at 0.42–2.88 s per step; the wait budget
  (D3) is a derived heuristic, not measured at ten lanes (spec §F R12).
- **R-G — heavy runs.** `internal/cli` suites are heavy: take a `moai slot` lease named for the shared
  target (`go-test-cli-shared`), pass `-timeout 25m` (the default 10 m was exceeded in t1448), scope
  `-run` to the change, and scrub the lane environment in the same compound call. The AC-FAL-010 family
  run took 757.794 s at the pin (ledger L9), which is why it is a WM6 closure step and not a per-commit
  check.
- **R-H — amendments reopen two completed SPECs** (D5); mitigated by doing it once, at sync, in
  one manager-spec delegation.
- **R-I — the cross-process helper is the heaviest new test** (WM1, about 5.5 s per iteration) and its
  detection power is a rate, not a certainty: the probe's two-process breach rate at the pin was 10 of 13
  (77%), so a process-local-mutex mutant (MU7) survives ten clean iterations with probability about
  0.23^10 ≈ 4 × 10^-7 if iterations are independent (assumed, not tested), and about 10^-3 at a pessimistic
  50% rate. Mitigation: `-count=10`, and MU7 is executed in WM6 and its failing test recorded.
- **R-J — the cap constants and the 500 ms margin are starting values.** The deadline (800 ms), the busy
  timeout (200 ms) and AC-FAL-007's margin are chosen from the measurements in L14 and are heuristics
  inside the derivation rule `3 × cap ≤ budget`; WM4 records the observed maximum return time and may
  tighten the margin, never loosen the rule.

## 7. Mutants the run phase executes (REQ per milestone; each must turn a named test red)

| Mutant | Edit | Must fail |
|---|---|---|
| MU1 | compute the serial slot / `serialHeld` before `WithLock` and pass it in | AC-FAL-001, -005 |
| MU2 | claim after `WithLock` returns (release the lock before the claim) | AC-FAL-002, -003, -004, -009 |
| MU3 | run the compensation in a second public `Mutate` after the section | AC-FAL-004 |
| MU4 | take no step lock around creation and rename | AC-FAL-006 |
| MU5 | open the lease record connection with the default 5000 ms busy timeout | AC-FAL-007 |
| MU6 | remove the lock-acquire-timeout → outcome mapping (the timeout surfaces as the raw error in the nominated form) | AC-FAL-007 |
| MU7 | serialize the section with a process-local `sync.Mutex` and hold no `flock` | AC-FAL-001 (the cross-process test; the goroutine-lane tests stay green under this mutant, which is the reason the process test exists) |
| MU8 | make the locked handle's read the adopting `Load` | AC-FAL-010 clause (iii) (`TestLockedBacklogLoadIsPure`) |
| MU9 | add an import of `internal/kanban` to a non-test file of `internal/homestate` | AC-FAL-011 (`TestHomestateDoesNotImportKanban`) |
| MU10 | run one `git` subprocess inside the section | AC-FAL-009 clause (iii) (`TestFactoryLeaseSectionAllowedSet`) |
| MU11 | map a bare-form queue-lock timeout back to a re-selection | AC-FAL-007 (`TestFactoryLeaseQueueLockStallBounded`) |

## 8. Cross-references

- spec.md §B (decisions and the hypothesis test), §E (supersession), §F (residuals), §H (open decisions).
- `acceptance.md` — criteria, evidence ledger with the RED-now cells.
- `evidence/` — the probes and the overlays that run them.
- `.claude/rules/moai/workflow/resource-slot-lease.md` — the heavy-run lease used in R-G.
