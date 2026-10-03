# SPEC-FACTORY-ATOMIC-LEASE-001 — Plan

Tier M. Card t1458, plan-start HEAD `2de0a2cb613b04765a1554f86685a3b48e0be806` (branch
`WT-atomic-lease`, worktree `.claude/worktrees/t1458`). Sections and milestones are ordered by
**decision reversibility** — the decisions most likely to change come first (the lock primitive's
interface, the bounded-claim semantics a lane sees, the creator interface shared by four callers),
mechanical edits last. No time estimates; priority labels and ordering only.

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
connection with a low busy timeout in its DSN and bounds the whole claim by a deadline derived from the
queue lock's wait budget; on the cap it restores its promotion and reports the existing `raced`
refusal. Separately, serialize the post-lease worktree step (creation plus the `git branch -m` rename,
in the one function all four callers share) behind a dedicated cross-process lock with a bounded wait —
measured to remove both observed git races, which dropping the rename alone does not. Everything is
TDD-first, each milestone's RED commit precedes its fix commit so the commit graph, not a
commit message, witnesses the order (`verification-claim-integrity.md` §2.3).

## 2. Decisions most likely to change (most reversible first)

**D1 — the primitive's shape (WM2).** `BacklogStore.WithLock(fn func(*LockedBacklog) error) error`:
acquire the lock through the existing `acquireLock` (same wait policy, same timeout error), refuse a
relocated queue exactly as `Mutate` does, run `fn`, release and join the release error as `Mutate`
does. `LockedBacklog` offers `Load()` (a fresh read under the lock) and `Mutate(func(*BacklogRecord)
error)` (today's `Mutate` body without the lock acquisition). `Mutate` itself becomes `WithLock` around
one `LockedBacklog.Mutate`, so its behavior is byte-preserved. The lock is a non-reentrant `flock` on
a separate descriptor: calling the *public* `Mutate` from inside the section contends with the
section's own descriptor and waits out the budget. That is a trap, not a feature, and §5 lists the
existing tests it breaks. An exported read accessor for the wait budget (`LockWaitBudget()`), so the
claim cap is derived from the budget and never copied from it.

**D2 — the bounded claim (WM4).** Constants in `factory_card.go`: a claim deadline and a claim busy
timeout whose sum is at most `LockWaitBudget()/3` (today 3.3 s / 3 = 1.1 s; starting values 800 ms and
200 ms). The lease path opens its record connection through a new `homestate` open variant that takes
the busy timeout into the DSN (O6: only an open-time value lasts). The open happens **before** the
section, so a slow open never lengthens the hold. The deadline context covers `RecordPicked` + the two
transitions as one claim. Cap or lock-acquire timeout → restore the promotion (inside the section),
return the lost-race outcome: `raced` for the nominated form (so `--wait` retries), a re-selection for
the bare form. **No new refusal token** (the closed set of twelve is pinned by SPEC-TODO-AUTO-PICK-001);
the detail line says the record or the queue lock was busy. Reversible: a thirteenth token is a later,
explicit amendment of that set.

**D3 — the worktree-step lock (WM5).** `factoryEnsureCardWorktree` takes a dedicated cross-process lock
around the creator call and the `git branch -m` rename together, and releases it before
`RecordCardWorktree`. The creator interface does not change (spec §B.3: creating the branch with its
final name failed 2 of 280 concurrent creations, so it is not a fix). Substrate: reuse the board lock's
path-parameterized `flock` / atomic-create implementation (`acquireBoardLockImpl`) behind a small exported
helper in `internal/kanban` that polls with the same retry policy (`boardLockRetryWait`) against a
**separate** lock file under `<root>/.moai/state/` — not the queue's lock, not `board.lock` (the leader's
board writes), and not `moai slot` (a named lease with state and an audit trail, too heavy for a
sub-second critical section). No existing generic named lock exists to reuse (`internal/spec/lock.go` is
scoped to a SPEC id). Wait budget, derived like the queue lock's and stated as a sizing heuristic and not
a worst-case bound: lanes (10) × the worst observed step (2.9 s, from a busy machine) × headroom (2) ≈
58 s, taken as 60 s — a named constant with a derivation test. The wait is generous on purpose: it bounds
a failure, and the usual wait for the last of N lanes is about N × the median step (0.73–1.22 s). On timeout the step fails with its existing error shape and creates nothing.
Reversible: if the run phase shows the lock's start-up cost matters, final-name creation under the same
lock is the next step and needs the creator interface change this plan declines.

**D4 — where the section's boundaries sit (WM3).** Inside: the pass's reads, the keep-set and slot
decisions, `Mutate` promotion, the seams, the claim, the compensation. Outside: `OpenFactory`, the quota
latch, `factoryEnsureCardWorktree` and everything after it (REQ-FAL-007). Arm (b)'s
`queueItemState` read and arm (b2)'s queue read go through the locked handle's `Load()`; no code
inside the section calls the public `Mutate` or `Add`. A guard test calls `Mutate` from inside a
section and asserts the timeout error, so a future edit that does is caught (§6 R-B).

**D5 — the supersession records (WM6 + sync).** The plan phase edits neither completed SPEC. At sync,
manager-spec adds one Amendments row to each (§E of the spec). Accepting the `completed → in-progress →
completed` round trip on both is a cost the operator may veto; the alternative (a pointer only in this
SPEC) leaves readers of the old SPECs reading "not closed". Decision recorded here so it can be
overruled.

## 3. What was observed at plan time

All of it is in spec.md §A.2 with the commands in `acceptance.md`'s evidence ledger and the probes under
`evidence/` (committed as `.go.txt` so the build never sees them, mapped in with `go test -overlay`).
The five symptoms reproduced RED on the unmodified tree; the cost and stall figures that size the
critical section were measured; two hypotheses were refuted or corrected by measurement and are kept in
the record rather than smoothed over:

- the naive "hold the lock across the claim" form starves other queue writers under a stalled record
  (O5), so the cap in D2 is a requirement and not polish;
- the t1448 audit's 1.65 s lock budget is stale — the tree derives 3.3 s (O1);
- the expected M5 fix (create the branch with its final name) is not a fix: 2 of 280 concurrent
  creations still failed in git's own `worktree add` (O11), where serializing the step failed 0 of 80
  (O14);
- a hang placed after the arm-(c) decision does **not** reproduce the two-serial-card breach (the
  second lane adopts the first lane's promoted card through arm (b2) and loses the version-checked
  edge); the breach needs both snapshots to precede both promotions. The RED test for arm (c)
  therefore holds lanes at the pass entry, before the snapshot.

## 4. Milestones

### WM1 — RED first, for every scope item (Priority High)

Nothing here changes behavior. Two commits, in this order: (1) the seam commit — one package variable
`factoryLeaseBeforeClaim func(arm, cardID string) error` (inert default) called in arms (a), (b), (b2)
and (c) immediately before the claim, plus the pass-entry hook the arm (c) test needs
(`factoryLeaseAtPassEntry`, inert); (2) the RED commit — the new tests below, failing on the commit
before them for the stated reason, and the tolerant harness.

- **Tolerant seam harness `nmRaceAtSeamTolerant`.** The existing `nmRaceAtSeam` waits for every lane to
  arrive at the seam before releasing any. Inside an exclusive section two lanes cannot both be at the
  seam, so it would hang for the wrong reason on the fixed tree. The tolerant form releases the lanes
  when all have arrived **or** a grace window (750 ms) after the last arrival has passed. On the
  unfixed tree both lanes arrive within milliseconds and are released together (RED, as observed); on
  the fixed tree the second lane is blocked at the lock, the window elapses, and it proceeds in order.
- **Operator-write probe `nmOperatorWriteDuring`.** From inside a seam, start the operator write in a
  goroutine and wait up to 400 ms for it; report whether it completed inside the section. Unfixed: it
  completes at once. Fixed: it stays blocked until the verb returns, then applies.
- New tests (names the criteria cite): `TestFactoryLeaseSerialDistinctNomineesExactlyOne` (M1,
  nominated), `TestFactoryLeaseSerialBareLanesExactlyOne` (M1, arm (c), held at the pass entry),
  `TestFactoryLeaseOperatorWriteWaitsForSection` (M2, nominated), `TestFactoryLeaseArmCOperatorHold`
  (M2, arm (c)), `TestFactoryLeaseCompensationKeepsOperatorPick` (M3), `TestFactoryLeaseOwnAssignedSerialSiblingsExactlyOne`
  (M4), `TestFactoryEnsureCardWorktreeConcurrentRealMaterializer` (M5, real materializer, no stub),
  `TestFactoryEnsureCardWorktreeStepLockBounded` and `TestFactoryWorktreeStepWaitDerivation` (M5, the
  step lock's bounded wait and its derivation), `TestFactoryLeaseRecordStallBounded` (bound), `TestFactoryLeaseCapWithinBoardBudget` (derivation),
  `TestFactoryLeaseSectionExcludesWorktreeStep` (REQ-FAL-007).
- Record the verbatim RED output of each in `progress.md` §E.2 and the **baseline pass count of the
  preservation selector** (AC-FAL-010) before any fix.

### WM2 — the primitive (Priority High)

`internal/kanban`: `WithLock`, `LockedBacklog`, `LockWaitBudget()`. Tests first: two mutations inside one
`WithLock` see each other's writes; the lock is held until `fn` returns (a contender blocks, then
proceeds); it is released on an error return and on a panic; a relocated queue is refused as `Mutate`
refuses it; `Mutate`'s existing tests stay green unchanged. `GOOS=windows GOARCH=amd64 go build ./...`.

### WM3 — the lease path inside the section (Priority High)

`factory_card.go`: `factoryNextNominate` runs validate, promotion, seam, claim and compensation inside
`WithLock`; `factoryNextSelectAndLease` runs inside `WithLock` per attempt, its decision reads taken
under the lock; the compensation takes the locked handle. The snapshot-reuse paths (`nom.serialHeld`
carried from the pre-lock validation into the in-lock check) are deleted: the in-lock check reads the
record itself. Flips AC-FAL-001 to -005 and AC-FAL-009. Move the existing tests §5 lists.

### WM4 — the bounded claim (Priority High; lands with WM3 — WM3 alone puts an unbounded record wait
inside the queue lock and must not be integrated without it)

`homestate`: the busy-timeout open variant sharing the existing DSN builder. `factory_card.go`: the cap
constants, the claim deadline, the lost-race mapping for a cap or a lock-acquire timeout. Flips
AC-FAL-007 and AC-FAL-008.

### WM5 — the worktree step (Priority Medium)

The lock helper in `internal/kanban` (tests first: two contenders serialize; release on error and
panic; the wait budget bounds a blocked contender; the Windows build), then `factoryEnsureCardWorktree`
taking it around creation and rename and releasing it before the record write. `nmIsolatedWorktrees`
needs no change (the stub still replaces the creator; the lock is taken around it). Flips AC-FAL-006.

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
- **R-C — Windows.** The lock is an atomic-create file there; compile-verified only (spec §F R3).
- **R-D — the section is slower than estimated.** The estimate is one extra queue mutation plus the
  claim (spec B.2); WM6 measures it, and the cap bounds the worst case at one third of the budget.
- **R-E — multi-lane stall.** Bounded per lease, not across lanes (spec §F R6); not measured.
- **R-F — the step lock covers this card's lanes only.** Four production callers share one function
  (O12) and all are covered; `moai worktree new`, session entry points and other git readers are not
  (spec §F R11). The AC-FAL-006 test calls the shared function directly.
- **R-F2 — start-up latency.** Lanes starting together queue at 0.42–2.88 s per step; the wait budget
  (D3) is a derived heuristic, not measured at ten lanes (spec §F R12).
- **R-G — heavy runs.** `internal/cli` suites are heavy: take a `moai slot` lease named for the shared
  target (`go-test-cli-shared`), pass `-timeout 25m` (the default 10 m was exceeded in t1448), scope
  `-run` to the change, and scrub the lane environment in the same compound call.
- **R-H — amendments reopen two completed SPECs** (D5); mitigated by doing it once, at sync, in
  one manager-spec delegation.

## 7. Mutants the run phase executes (REQ per milestone; each must turn a named test red)

| Mutant | Edit | Must fail |
|---|---|---|
| MU1 | compute the serial slot / `serialHeld` before `WithLock` and pass it in | AC-FAL-001, -005 |
| MU2 | claim after `WithLock` returns (release the lock before the claim) | AC-FAL-002, -003, -004, -009 |
| MU3 | run the compensation in a second public `Mutate` after the section | AC-FAL-004 |
| MU4 | take no step lock around creation and rename | AC-FAL-006 |
| MU5 | open the lease record connection with the default 5000 ms busy timeout | AC-FAL-007 |
| MU6 | remove the lock-acquire-timeout → lost-race mapping | AC-FAL-007 |

## 8. Cross-references

- spec.md §B (decisions and the hypothesis test), §E (supersession), §F (residuals).
- `acceptance.md` — criteria, evidence ledger with the RED-now cells.
- `evidence/` — the probes and the overlays that run them.
- `.claude/rules/moai/workflow/resource-slot-lease.md` — the heavy-run lease used in R-G.
