# SPEC-FACTORY-ATOMIC-LEASE-001 — Acceptance

Tier M. Every release-blocking criterion carries a **RED-now cell** (a read-only single-invocation
command, its verbatim output, its exit code, the pinned tree — held in the evidence ledger and cited by
row) and a **green-path cell** (the milestone that flips it and what the passing output becomes), per
`.claude/rules/moai/development/verification-completeness.md` §2 and §2.1, and a **mutation criterion**
(a named mutant of plan §7 that must turn a named test red). A criterion that cannot be red on arrival is
a **regression-guard**, is labelled so, and is never recorded as release-blocking. Given-When-Then is the
verification layer's format; the requirements are GEARS in `spec.md`.

Concurrency criteria run under repetition **and** `-race`, with the lane environment scrubbed in the same
compound call (setup row S1), under a heavy-run lease (S2). 14 criteria (Tier M ceiling 16), tracing all
13 requirements.

## Setup rows

- **S1 — scrub.** Every `internal/cli` command in this file runs as one compound call whose prefix is
  `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER && `
  (a separate `unset` does not reach the next call). The ledger's command column holds the `go test`
  invocation itself, because the single-invocation form that §2.1 admits cannot carry the prefix; the
  prefix is part of how each run was made.
- **S2 — heavy-run lease.** `moai slot acquire --resource go-test-cli-shared --max-duration 25m` before an
  `internal/cli` run and `moai slot release --resource go-test-cli-shared` straight after; `-timeout 25m`
  (the default 10 m was exceeded in the t1448 run).
- **S3 — pins.** Plan-start tree `2de0a2cb613b04765a1554f86685a3b48e0be806` (branch `WT-atomic-lease`).
  The probes were mapped in with `go test -overlay`; the SPEC directory itself was untracked at the time,
  and no tracked file differed from the pin. `git merge-base develop HEAD` returned the same SHA.

## Evidence ledger (RED-now observations and context)

Rows `L1`–`L8` were observed on the pin. After the run they are history by design: each describes the
pinned tree and prints something else on a tree that carries the linked milestone. Output is quoted
exactly as the tool printed it; where a stream is abridged the row says so and says what was elided.
Every command below is the exact command that produced the output beside it, with the anchored pattern
`moai spec lint` asks for. Several probes are timing- or interleaving-dependent, so a re-run prints
different durations and counts; the sentence under each row gives the range seen across all runs.

### L1 — M1 (nominated), M2 (nominated), M3, and the final-name counter-observation

```
$ go test ./internal/cli -overlay=.moai/specs/SPEC-FACTORY-ATOMIC-LEASE-001/evidence/overlay-tree.json -run '^(TestT1458ProbeM1SerialTwoDistinct|TestT1458ProbeM2HoldAfterPromotion|TestT1458ProbeM3ABAOperatorPick|TestT1458ProbeM5FinalNameCreation)$' -count=1 -v -timeout 8m
=== RUN   TestT1458ProbeM1SerialTwoDistinct
2026/10/03 12:30:27 WARN config sections directory not found, using defaults path=/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestT1458ProbeM1SerialTwoDistinct545844063/002/.moai/config/sections
    zz_t1458_probe_test.go:57: PROBE M1 serial cards leased = 2 rows=[t1=leased/lane-1 t2=leased/lane-2]
    zz_t1458_probe_test.go:59: serial exclusivity breached: 2 serial cards are leased at once ([t1=leased/lane-1 t2=leased/lane-2])
--- FAIL: TestT1458ProbeM1SerialTwoDistinct (4.77s)
=== RUN   TestT1458ProbeM2HoldAfterPromotion
2026/10/03 12:30:32 WARN config sections directory not found, using defaults path=/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestT1458ProbeM2HoldAfterPromotion968543491/002/.moai/config/sections
    zz_t1458_probe_test.go:79: PROBE M2 err=<nil> queue=hold record=leased holder=lane-1 stdout="t1 stage=- worktree=003\nt1\tunknown\t\t\thold\t\tfactory card 1\n" stderr="note: open pull requests unavailable (exit status 1); link column left empty, landed check still ran\nnote: the landed check against origin/main could not answer for t1; those cards report unknown rather than no-link, because an unanswerable query is not evidence of not-landed\n"
    zz_t1458_probe_test.go:81: an operator write committed between the decision and the claim and the card was leased anyway: queue=hold record=leased holder=lane-1
--- FAIL: TestT1458ProbeM2HoldAfterPromotion (2.76s)
=== RUN   TestT1458ProbeM3ABAOperatorPick
2026/10/03 12:30:34 WARN config sections directory not found, using defaults path=/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestT1458ProbeM3ABAOperatorPick2046487017/002/.moai/config/sections
    zz_t1458_probe_test.go:108: PROBE M3 err=factory next: injected seam failure queue=queued (the operator's fresh pick was picked)
    zz_t1458_probe_test.go:110: the operator's fresh pick was reverted by the compensation: queue=queued want picked
--- FAIL: TestT1458ProbeM3ABAOperatorPick (1.98s)
=== RUN   TestT1458ProbeM5FinalNameCreation
    zz_t1458_probe_test.go:261: PROBE M5-final-name iterations=40 iterations with a failed creation=0
--- PASS: TestT1458ProbeM5FinalNameCreation (24.23s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	35.144s
FAIL
exit status 1
tree: 2de0a2cb613b04765a1554f86685a3b48e0be806
```

The first three probes are the RED-now cells of AC-FAL-001 (nominated clause), -002 and -004; they failed
identically on all three runs made from the committed probe file (and on the scratch run before it). The fourth is a counter-observation, not a
RED-now cell: creating the branch with its final name passed 40 of 40 in this run and failed once in 40 in
two others (`fatal: failed to read .git/worktrees/wt-b/commondir`) — 2 failures in 280 iterations over
seven runs, which is why spec §B.3 does not adopt it.

### L2 — M5 rename collision, real materializer (abridged)

```
$ go test ./internal/cli -overlay=.moai/specs/SPEC-FACTORY-ATOMIC-LEASE-001/evidence/overlay-tree.json -run '^TestT1458ProbeM5RenameCollision$' -count=1 -v -timeout 12m
=== RUN   TestT1458ProbeM5RenameCollision
[20 lines elided, one per iteration: "2026/10/03 12:31:23 WARN config sections directory not found, using defaults path=<tmp>/…/.moai/config/sections"]
    zz_t1458_probe_test.go:162: PROBE M5 iterations=20 failed-iterations=10 stranded-dir-and-foreign-worktree-precheck-refusal=10
    zz_t1458_probe_test.go:164: 10 of 20 concurrent worktree steps failed; first error: rename the card worktree branch: exit status 128: error: unable to move logfile /private/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/Tes
--- FAIL: TestT1458ProbeM5RenameCollision (63.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	64.336s
FAIL
exit status 1
tree: 2de0a2cb613b04765a1554f86685a3b48e0be806
```

RED-now cell of AC-FAL-006. Four runs of the same step failed 6, 5, 8 and 10 of 20 iterations (29 of 80,
36%). The `stranded-…=10` count is the number of failed steps that left a directory with no recorded
worktree for which `factoryRefuseForeignWorktree` returns a refusal; the earlier counted run gave 8 of 8.

### L3 — M2 (arm (c)) and M4 (arm (a)), unmodified tree

```
$ go test ./internal/cli -overlay=.moai/specs/SPEC-FACTORY-ATOMIC-LEASE-001/evidence/overlay-arms.json -run '^(TestT1458ProbeArmCHoldAfterPromotion|TestT1458ProbeArmASerialSiblings)$' -count=1 -v -timeout 12m
=== RUN   TestT1458ProbeArmCHoldAfterPromotion
2026/10/03 12:32:35 WARN config sections directory not found, using defaults path=/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestT1458ProbeArmCHoldAfterPromotion770360404/002/.moai/config/sections
    zz_t1458_arms_probe_test.go:199: PROBE arm-c-hold err=<nil> queue=hold record=leased holder=lane-1 stdout="t1 stage=- worktree=003\nt1\tunknown\t\t\thold\t\tfactory card 1\n" stderr="note: open pull requests unavailable (exit status 1); link column left empty, landed check still ran\nnote: the landed check against origin/main could not answer for t1; those cards report unknown rather than no-link, because an unanswerable query is not evidence of not-landed\n"
    zz_t1458_arms_probe_test.go:201: arm (c) leased a card whose hold was committed before its claim: queue=hold record=leased holder=lane-1
--- FAIL: TestT1458ProbeArmCHoldAfterPromotion (2.95s)
=== RUN   TestT1458ProbeArmASerialSiblings
2026/10/03 12:32:38 WARN config sections directory not found, using defaults path=/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestT1458ProbeArmASerialSiblings2527311129/002/.moai/config/sections
    zz_t1458_arms_probe_test.go:221: PROBE arm-a serial cards leased=2 rows=[t1=leased/lane-1 t2=leased/lane-2] errs=[<nil> <nil>]
    zz_t1458_arms_probe_test.go:223: arm (a): 2 serial cards are leased at once ([t1=leased/lane-1 t2=leased/lane-2])
--- FAIL: TestT1458ProbeArmASerialSiblings (6.49s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	10.933s
FAIL
exit status 1
tree: 2de0a2cb613b04765a1554f86685a3b48e0be806
```

RED-now cells of AC-FAL-003 and AC-FAL-005; both failed on both runs made with the final probe file. The probes hang the lane on the
existing `factoryCardNow` variable at the first claim write, keyed on the calling function names — a probe
device; the real tests use a seam (plan WM1).

### L4 — M1 (bare form, arm (c)): both snapshots precede both promotions (abridged)

```
$ go test ./internal/cli -overlay=.moai/specs/SPEC-FACTORY-ATOMIC-LEASE-001/evidence/overlay-arms.json -run '^TestT1458ProbeArmCSerialTwoLanes$' -count=8 -v -timeout 12m
[abridged to the eight PROBE lines; the interleaved "=== RUN", WARN, "--- FAIL" and "arm (c): 2 serial cards…" lines are elided]
    PROBE arm-c serial cards leased=2 rows=[t1=leased/lane-1 t2=leased/lane-2] errs=[<nil> <nil>]
    PROBE arm-c serial cards leased=2 rows=[t1=leased/lane-2 t2=leased/lane-1] errs=[<nil> <nil>]
    PROBE arm-c serial cards leased=2 rows=[t1=leased/lane-1 t2=leased/lane-2] errs=[<nil> <nil>]
    PROBE arm-c serial cards leased=2 rows=[t1=leased/lane-2 t2=leased/lane-1] errs=[<nil> <nil>]
    PROBE arm-c serial cards leased=2 rows=[t1=leased/lane-2 t2=leased/lane-1] errs=[<nil> <nil>]
    PROBE arm-c serial cards leased=2 rows=[t1=leased/lane-1 t2=leased/lane-2] errs=[<nil> <nil>]
    PROBE arm-c serial cards leased=2 rows=[t1=leased/lane-2 t2=leased/lane-1] errs=[<nil> <nil>]
    PROBE arm-c serial cards leased=2 rows=[t1=leased/lane-1 t2=leased/lane-2] errs=[<nil> <nil>]
FAIL	github.com/modu-ai/moai-adk/internal/cli	48.996s
exit status 1
tree: 2de0a2cb613b04765a1554f86685a3b48e0be806
```

RED-now cell of AC-FAL-001 (bare-form clause): 8 of 8 here, 8 of 8 in an earlier run. **A correction kept on
the record:** when the same probe held each lane after its decision instead of at the pass entry, only one
serial card leased — the second lane adopted the first lane's promoted card through arm (b2) and lost the
version-checked edge. The breach needs both snapshots to precede both promotions, which is why the RED test
holds lanes at the pass entry.

### L5 — the factory record claim: cost and stall behavior

```
$ go test ./internal/homestate -overlay=.moai/specs/SPEC-FACTORY-ATOMIC-LEASE-001/evidence/overlay-homestate.json -run '^TestT1458ProbeRecordClaim$' -count=1 -v -timeout 4m
=== RUN   TestT1458ProbeRecordClaim
    zz_t1458_probe_test.go:46: PROBE uncontended claim n=100 p50=964.917µs p95=1.479292ms max=7.781167ms
    zz_t1458_probe_test.go:65: PROBE handle busy_timeout=5000ms
    zz_t1458_probe_test.go:70: PROBE read while another connection holds a write tx: elapsed=52.708µs err=<nil>
    zz_t1458_probe_test.go:77: PROBE claim ctx-deadline=500ms vs 3s write holder: elapsed=3.107069958s err=context deadline exceeded
    zz_t1458_probe_test.go:82: PROBE claim no-deadline vs 7s write holder: elapsed=7.002348792s err=<nil>
    zz_t1458_probe_test.go:98: PROBE claim busy_timeout=200ms + ctx-deadline=500ms vs 3s write holder: elapsed=712.49125ms err=context deadline exceeded
        database is locked (5) (SQLITE_BUSY)
    zz_t1458_probe_test.go:104: PROBE claim busy_timeout=200ms (runtime PRAGMA) + ctx-deadline=1s vs 7s write holder: elapsed=5.092852125s err=context deadline exceeded
        database is locked (5) (SQLITE_BUSY)
    zz_t1458_probe_test.go:108: PROBE busy_timeout on that handle afterwards=5000ms
--- PASS: TestT1458ProbeRecordClaim (24.70s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/homestate	25.147s
exit 0
tree: 2de0a2cb613b04765a1554f86685a3b48e0be806
```

Context for spec §A.2 O3–O6 and the RED-now cell of AC-FAL-007: no wait cap exists — a 500 ms deadline
returned after 3.107 s. Across three runs the uncontended claim ranged p50 0.96–2.03 ms, p95 1.48–16.77 ms,
max 7.78–72.52 ms (load-dependent); the stall lines repeated within 0.1 s.

### L6 — the queue lock's wait budget as the tree derives it

```
$ go test ./internal/kanban -overlay=.moai/specs/SPEC-FACTORY-ATOMIC-LEASE-001/evidence/overlay-kanban.json -run '^TestT1458ProbeBoardLockBudget$' -count=1 -v
=== RUN   TestT1458ProbeBoardLockBudget
    zz_t1458_probe_test.go:11: PROBE boardLockWaitBudget=3.3s writers=10 mutationCost=33ms headroom=10
--- PASS: TestT1458ProbeBoardLockBudget (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/kanban	0.364s
exit 0
tree: 2de0a2cb613b04765a1554f86685a3b48e0be806
```

### L7 — M5 control: the serialized step, and final-name creation again (abridged)

```
$ go test ./internal/cli -overlay=.moai/specs/SPEC-FACTORY-ATOMIC-LEASE-001/evidence/overlay-tree.json -run '^(TestT1458ProbeM5SerializedStep|TestT1458ProbeM5FinalNameCreation)$' -count=2 -v -timeout 9m
[abridged to the four PROBE lines; 40 WARN lines and the RUN/PASS lines are elided]
    PROBE M5-serialized iterations=20 failed-iterations=0 step-duration min=611.024ms p50=912.092542ms max=2.257248125s
    PROBE M5-final-name iterations=40 iterations with a failed creation=0
    PROBE M5-serialized iterations=20 failed-iterations=0 step-duration min=692.143417ms p50=1.217420209s max=2.877402333s
    PROBE M5-final-name iterations=40 iterations with a failed creation=0
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	364.602s
exit 0
tree: 2de0a2cb613b04765a1554f86685a3b48e0be806
```

The serialized step used an in-process mutex in the probe. This row is not a RED-now cell; it is the
measurement behind the choice of fix (spec §B.3). Over four runs the serialized step failed 0 of 80
iterations; one step took 0.42–0.69 s at best, 0.73–1.22 s at the median and 1.14–2.88 s at worst — the
slower figures came from the later runs, made while the machine was busier.

### L8 — RED-now for the structural and documentation criteria

```
$ grep -rn factoryLeaseClaimWaitCap internal/cli
(no output)
exit 1
$ grep -n WithLock internal/kanban/backlog_store.go
(no output)
exit 1
$ grep -c SPEC-FACTORY-ATOMIC-LEASE-001 .moai/specs/SPEC-TODO-AUTO-PICK-001/spec.md
0
exit 1
$ grep -c SPEC-FACTORY-ATOMIC-LEASE-001 .moai/specs/SPEC-TODO-CLASSIFY-DISPATCH-001/spec.md
0
$ grep -c SPEC-FACTORY-ATOMIC-LEASE-001 CHANGELOG.md
0
exit 1
tree: 2de0a2cb613b04765a1554f86685a3b48e0be806
```

The exit status was printed with `; echo "exit=$?"` for rows 1, 2, 3 and 5. Row 4, the
`SPEC-TODO-CLASSIFY-DISPATCH-001` count, printed `0` and its exit status was not displayed; `grep -c`
returning 1 on a zero count is the observed behavior of rows 3 and 5, stated here as an inference for row 4.

## Criteria

Classification: AC-FAL-001 to -009 are **release-blocking**; AC-FAL-010, -011 and -013 are
**regression-guards**; AC-FAL-012 and -014 are release-blocking at the sync phase (their RED-now cells are
in L8).

### AC-FAL-001 — two serial leases at once: exactly one wins (M1)

**Covers**: maps REQ-FAL-001, REQ-FAL-002

**Given** a queue holding two queued serial cards `t1` and `t2`, no serial card in flight, and two
registered lanes,
**When** (a) the lanes nominate `t1` and `t2` and are held at the nomination seam by the tolerant
harness, and (b) the lanes run bare `factory next` held at the pass entry,
**Then** exactly one record row is `leased`; in (a) the other lane exits 4 with one stderr line
`factory next: refused serial-slot: …`, in (b) it exits 3 with `no card is available`; the other card has
no `leased` row.

- **Command**: `go test ./internal/cli -run '^(TestFactoryLeaseSerialDistinctNomineesExactlyOne|TestFactoryLeaseSerialBareLanesExactlyOne)$' -count=20 -race -timeout 25m` (S1, S2).
- **RED-now cell**: L1 (`PROBE M1 serial cards leased = 2`) and L4 (`leased=2`, 8 of 8). Red for the stated
  reason: the in-lock check reuses the pre-lock `serialHeld`, and arm (c)'s record snapshot predates both
  promotions.
- **Green-path cell**: WM3. Output becomes `ok  github.com/modu-ai/moai-adk/internal/cli`, each named test
  `--- PASS` in all 20 repetitions, `DATA RACE` count 0.
- **Mutation**: MU1 (compute the serial slot before `WithLock`) turns both tests red.

### AC-FAL-002 — an operator write cannot interleave with a nominated lease (M2)

**Covers**: maps REQ-FAL-001, REQ-FAL-003

**Given** a queued nominee `t1` and a seam inside the lease that starts an operator `hold` of `t1` in a
goroutine,
**When** a lane nominates `t1`,
**Then** (i) the operator write has not completed when the seam returns (it is blocked on the section's
lock), (ii) after the verb returns the operator's hold is applied (queue `hold`, not lost), and (iii) a
hold committed before the verb starts is refused with `refused held`, writing no record row and changing
nothing in the queue.

- **Command**: `go test ./internal/cli -run '^TestFactoryLeaseOperatorWriteWaitsForSection$' -count=20 -race -timeout 25m`.
- **RED-now cell**: L1 (`PROBE M2 err=<nil> queue=hold record=leased holder=lane-1`; message "an operator
  write committed between the decision and the claim and the card was leased anyway"). Clause (iii) is a
  regression-guard (green today) and is not what makes the criterion red.
- **Green-path cell**: WM3. Clause (i) reads `completed inside section = false` in all 20 repetitions.
- **Mutation**: MU2 (release the lock before the claim) turns clause (i) red.
- **Not claimed**: the end state `queue=hold record=leased` is still reachable when the operator's write
  arrives mid-section (spec §F R10); the criterion tests the ordering, not the end state.

### AC-FAL-003 — the same for the unnominated arm (c) (M2)

**Covers**: maps REQ-FAL-001, REQ-FAL-003

**Given** queued parallelizable cards `t1`, `t2` and a hook that starts an operator `hold` of `t1` in a
goroutine at arm (c)'s claim point,
**When** a lane runs bare `factory next`,
**Then** the operator write has not completed at the claim point and is applied after the verb returns;
and a hold committed before the verb starts makes the bare form skip `t1` and lease `t2`.

- **Command**: `go test ./internal/cli -run '^TestFactoryLeaseArmCOperatorHold$' -count=20 -race -timeout 25m`.
- **RED-now cell**: L3 (`PROBE arm-c-hold err=<nil> queue=hold record=leased holder=lane-1`).
- **Green-path cell**: WM3; the first clause prints `completed at claim point = false`.
- **Mutation**: MU2.

### AC-FAL-004 — the compensation keeps the operator's fresh pick (M3)

**Covers**: maps REQ-FAL-004

**Given** a queued nominee, a seam that fails the claim, and a seam goroutine that after the promotion
writes the queue item to `queued` then to `picked` (an operator unpick and re-pick, as direct store
writes),
**When** a lane nominates the card,
**Then** the invocation fails with the injected error (not exit 4), the operator goroutine's writes have
not completed when the seam returns, and after the verb returns the card's queue state is `picked`, the
operator's last write, with no `leased` row.

- **Command**: `go test ./internal/cli -run '^TestFactoryLeaseCompensationKeepsOperatorPick$' -count=20 -race -timeout 25m`.
- **RED-now cell**: L1 (`PROBE M3 … queue=queued (the operator's fresh pick was picked)`).
- **Green-path cell**: WM3; final queue state `picked` in all 20 repetitions.
- **Mutation**: MU3 (compensation as a second public `Mutate` after the section) and MU2 each turn it red.

### AC-FAL-005 — two lanes lease their own assigned serial cards: exactly one (M4)

**Covers**: maps REQ-FAL-005

**Given** serial cards `t1` and `t2` recorded `assigned` to `lane-1` and `lane-2`, nothing in flight,
**When** both lanes run bare `factory next` held at their claim (arm (a)),
**Then** exactly one card is `leased`; the other lane exits 3 and its card stays `assigned` with owner and
version unchanged.

- **Command**: `go test ./internal/cli -run '^TestFactoryLeaseOwnAssignedSerialSiblingsExactlyOne$' -count=20 -race -timeout 25m`.
- **RED-now cell**: L3 (`PROBE arm-a serial cards leased=2`).
- **Green-path cell**: WM3; one `leased`, one exit 3, in all 20 repetitions.
- **Mutation**: MU1.

### AC-FAL-006 — concurrent worktree steps both succeed (M5)

**Covers**: maps REQ-FAL-008

**Given** two leased cards with no recorded worktree in one repository and the real worktree materializer
(no stub),
**When** both lanes run the worktree step at the same instant, 12 iterations inside the test,
**Then** every step returns nil, each card records its own worktree path and its branch is `WT-<slug>`;
and, with the step lock held by the test, a step called with a shortened wait returns the wait error
within that wait plus 500 ms, creates no directory and leaves the card's record row unchanged; and the
wait budget constant is not smaller than lanes (10) × the worst observed step (2.9 s) × headroom (2).

- **Command**: `go test ./internal/cli -run '^(TestFactoryEnsureCardWorktreeConcurrentRealMaterializer|TestFactoryEnsureCardWorktreeStepLockBounded|TestFactoryWorktreeStepWaitDerivation)$' -count=3 -race -timeout 25m` (36 concurrent iterations).
- **RED-now cell**: L2 (`failed-iterations=10` of 20, `stranded…=10`); at the 36% per-iteration rate
  seen over four runs, a clean 36-iteration run is chance at about one in ten million, so a green is not
  luck. L1 and L7 show the alternative (final-name creation alone) also fails, at a much lower rate.
- **Green-path cell**: WM5; 0 failed iterations in 36. A clean 36 bounds the fixed rate at about 8%
  (95%); the claim is "removed the observed 36% and the observed 0.7%", not "zero".
- **Mutation**: MU4 (no step lock) turns the concurrency test red.
- **Control**: L7, the same step serialized in-process, 0 of 80 over four runs.

### AC-FAL-007 — a stalled record bounds the lease and does not starve the queue (REQ-FAL-006)

**Covers**: maps REQ-FAL-006, REQ-FAL-001

**Given** a record write transaction held by a second connection for three times the lease claim wait cap,
and a queued nominee,
**When** a lane nominates it and, at the same instant, a queue writer (`Mutate`) starts,
**Then** the lease returns within the cap plus 750 ms with exit 4 and `factory next: refused raced: …`,
its queue item is `queued` again, the queue writer completes without a lock-held error, and the queue's
lock is acquirable immediately afterwards; with the queue's lock held by the test for longer than its wait
budget the lease reports `raced` (exit 4), not an infrastructure error (exit 1); and the bare form
reports no card or re-selects (exit 3).

- **Command**: `go test ./internal/cli -run '^TestFactoryLeaseRecordStallBounded$' -count=20 -race -timeout 25m` (each repetition waits out about three caps).
- **RED-now cell**: L5 — a claim with a 500 ms deadline against a 3 s holder returned after
  `3.107069958s`: no bound exists. The "queue writer is not starved" clause is green on arrival (today
  the claim holds no lock) and becomes the constraint on the fix; it is what MU5 breaks.
- **Green-path cell**: WM3 + WM4; the lease returns in under about 1.1 s in all repetitions.
- **Mutation**: MU5 (default 5000 ms busy timeout on the lease connection) turns the bound red; MU6
  (drop the lock-timeout mapping) turns the second clause red (exit 1, not 4).

### AC-FAL-008 — the cap is derived from the queue lock's budget (REQ-FAL-006)

**Covers**: maps REQ-FAL-006

**Given** the lease claim deadline and busy-timeout constants, **When**
`TestFactoryLeaseCapWithinBoardBudget` runs, **Then** (deadline + busy timeout) × 3 is not greater than
the queue lock's wait budget as `kanban.LockWaitBudget()` returns it, and the test reads that accessor
rather than a literal (the precedent is `slot_lease_cross_test.go`'s stall-release guard).

- **Command**: `go test ./internal/cli -run '^TestFactoryLeaseCapWithinBoardBudget$' -count=1`.
- **RED-now cell**: L8 first row (`factoryLeaseClaimWaitCap` absent, exit 1); L6 gives the budget the
  guard will read, 3.3 s.
- **Green-path cell**: WM4; `ok`.
- **Mutation**: raising the claim deadline to 1.5 s turns it red.

### AC-FAL-009 — the section contains only the two stores, and its hold is measured (REQ-FAL-007)

**Covers**: maps REQ-FAL-007, REQ-FAL-001

**Given** a lease through each path (nominated, bare arm (a), bare arm (c)) with the worktree creator
stubbed and the record writes counted,
**When** the lease runs,
**Then** (i) a `Mutate` issued from inside the creator stub completes within 500 ms (no queue lock is
held during worktree creation) and likewise at the card-worktree record write; (ii) while the lock is
held at most one queue promotion commit, at most one restore commit and exactly three record write
transactions are issued, and none is issued before the lock is held; and (iii) the hold-time
distribution of the section — 10 sequential leases, then 2, 4 and 10 concurrent lanes — is recorded in
`progress.md` §E.2 with p50, p95, max and the command that produced it (no wall-clock threshold is
asserted).

- **Command**: `go test ./internal/cli -run '^TestFactoryLeaseSectionExcludesWorktreeStep$' -count=20 -race -timeout 25m`.
- **RED-now cell**: L8 second row (`WithLock` absent, exit 1) — today no section exists, so clause (ii)
  cannot hold. Clause (i) alone is green on arrival; it is the constraint on the fix and is what the
  creator-inside-the-section mutant below breaks.
- **Green-path cell**: WM2 + WM3; counters read `promotions=1 restores=0 record_writes=3
  writes_before_lock=0`.
- **Mutation**: moving the creator call inside the section turns clause (i) red.

### AC-FAL-010 — selection behavior is unchanged where nothing interleaves (REQ-FAL-009) — regression-guard

**Covers**: maps REQ-FAL-009

**Given** the preservation family swept before any change — the exact test names that
`go test ./internal/cli -list` prints for the family `TestFactoryNext`, `TestTodoLane`,
`TestTodoNonLane`, `TestFactoryFallback`, `TestAutoPick`, `TestAutoRank` and `TestAutoHelp`, dumped at
WM1 into `progress.md` §E.2 together with the pass count of one full run of them,
**When** the anchored alternation of exactly those names is run after WM6,
**Then** it reports 0 `--- FAIL` and 0 `--- SKIP`, a pass count at least the baseline, and the tests plan
§5 moves keep every assertion line they had (checked by reading the diff).

- **Command**: `go test ./internal/cli -run '^(<the names from the WM1 -list dump, joined with |>)$' -count=1 -timeout 25m`.
  The family is the selector t1448 used; it includes the t1407 serial-slot tests
  (`TestFactoryNextOwnAssignedSerialCardLeasesPastSiblingAssigned` and its siblings) and the arm-ordering
  tests. A prefix pattern is not used because an anchored prefix selects nothing and an unanchored one
  also sweeps longer names; the enumerated list is the selector.
- **Classification**: regression-guard — green on arrival by design, never recorded as release-blocking.
  Its non-vacuity is the swept-count floor: a run that sweeps fewer tests than the baseline fails the
  criterion. No mutant is claimed for it.
- **Green-path cell**: n/a (green throughout); the criterion fails if any milestone turns it red.

### AC-FAL-011 — lock order and layering hold (REQ-FAL-010) — regression-guard

**Covers**: maps REQ-FAL-010

**Given** the tree, **When** `TestHomestateDoesNotImportKanban` runs and a test calls the public `Mutate`
from inside a lease section, **Then** `internal/homestate` has no dependency path to `internal/kanban`, and
the in-section `Mutate` returns the lock-held timeout error within the budget plus 750 ms (the
self-contention guard of plan D4).

- **Command**: `go test ./internal/homestate ./internal/cli -run '^(TestHomestateDoesNotImportKanban|TestFactoryLeaseSectionRejectsNestedMutate)$' -count=1`.
- **Classification**: regression-guard — the layering holds today (spec §A.2 O8, read from `go list -deps`
  of both packages in this session, not a filtered count). Non-vacuity: adding an import of
  `internal/kanban` to a `homestate` file must turn the first test red; that mutant is executed in WM6
  and its failing test name recorded.
- The ordering half of REQ-FAL-010 (the lease takes the queue lock before any record write) is carried by
  AC-FAL-009 clause (ii).

### AC-FAL-012 — the records state what is not closed (REQ-FAL-011)

**Covers**: maps REQ-FAL-011

**Given** the run and sync records, **When** read at the sync phase, **Then** `spec.md` §F still lists
R1–R12; the `progress.md` §E.4 signal and this SPEC's `CHANGELOG.md` entry each name which of M1–M5 closed
and each §F window that remains open; and every use of "atomic" in that CHANGELOG entry carries
"within the critical section" (or "across the critical section" for the section itself) in the same
sentence.

- **Command**: `grep -c SPEC-FACTORY-ATOMIC-LEASE-001 CHANGELOG.md` (at least 1 at sync) and
  `grep -n -i atomic CHANGELOG.md` read against this SPEC's entry.
- **RED-now cell**: L8 last row (`0`, exit 1). The sentence-level qualifier check is a reading, so that
  part is *(doctrine-only)* and is stated as such in the sync record.
- **Green-path cell**: sync phase, after the CHANGELOG entry is written.
- **Mutation**: deleting the R6 paragraph from `spec.md` makes `grep -c 'R6 — multi-lane stall' spec.md`
  return 0, which the sync check reads.

### AC-FAL-013 — the diff is surgical (REQ-FAL-012) — regression-guard

**Covers**: maps REQ-FAL-012

**Given** the card branch, **When** measured at WM6 against `git merge-base develop HEAD` re-derived at
reading time, **Then** the changed paths are only: `internal/cli/factory_card.go` and its test files,
`internal/kanban` lock files and their tests, `internal/homestate/factory.go` and its test, and
`.moai/specs/SPEC-FACTORY-ATOMIC-LEASE-001/**`; and **zero** paths under `internal/template/`,
`.claude/rules/` (unless the doctrine sweep named a line), `internal/cli/todo.go`, `internal/cli/gtd.go`,
`internal/homestate/card_transition.go`, `internal/cli/session_worktree.go`, `internal/cli/worktree/`,
`internal/cli/root.go`, or any queue or factory schema file.

- **Command** (two plain calls — a compound shell around git is refused in a worktree session):
  `git merge-base develop HEAD`, then `git diff --name-only <that sha>..HEAD -- <pathspec>` once for the
  allowed set (control: non-empty) and once for the forbidden set (probe: empty).
- **Classification**: regression-guard — the forbidden-path probe is empty on arrival. It is only
  meaningful before the card merges (after a merge the base is the tip and the range is empty); the
  control must be non-empty or the result is "unmeasured", not "clean".
- **Mutation**: touching `internal/cli/todo.go` makes the forbidden-set probe non-empty.

### AC-FAL-014 — supersession recorded, completed bodies untouched in the plan (REQ-FAL-013)

**Covers**: maps REQ-FAL-013

**Given** spec §E, **When** the sync phase closes this SPEC, **Then** both
`.moai/specs/SPEC-TODO-AUTO-PICK-001/spec.md` and `.moai/specs/SPEC-TODO-CLASSIFY-DISPATCH-001/spec.md`
carry an Amendments entry that cites `SPEC-FACTORY-ATOMIC-LEASE-001`, records `prior_completed_sha`, and
leaves their `status` at `completed`; and, at every point before that, neither body differs from the pin.

- **Command** (plan-phase half): `git diff --name-only 2de0a2cb613b04765a1554f86685a3b48e0be806..HEAD -- .moai/specs/SPEC-TODO-AUTO-PICK-001 .moai/specs/SPEC-TODO-CLASSIFY-DISPATCH-001`
  prints nothing. Sync half: `grep -c SPEC-FACTORY-ATOMIC-LEASE-001` on each `spec.md` is at least 1.
- **RED-now cell**: L8 rows 3 and 4 (`0` each) for the sync half; the plan-phase half is green by
  construction and is the reason the sync half cannot be mistaken for a plan-time edit.
- **Green-path cell**: sync phase, one manager-spec re-delegation (spec §B.4, plan D5).

## Edge cases

- **E1 — the same card nominated by two lanes at once.** Still exactly one holder and one refusal
  (`TestFactoryNextNominateSameCardExactlyOne`, moved to the tolerant harness, assertions unchanged;
  AC-FAL-010).
- **E2 — an operator-picked nominee (no promotion).** The section still reads and claims; the compensation
  does not run for a promotion the invocation did not make (`TestFactoryNextNominateCompensateRechecksRecord`
  table row 6, kept).
- **E3 — a relocated queue.** `WithLock` refuses it as `Mutate` does; the lease reports the relocation
  error, not `raced` (WM2 primitive test).
- **E4 — a claim that fails after `RecordPicked`.** The row stays `picked` and unowned, the queue item is
  restored (spec §F R4); unchanged and not tested here.
- **E5 — Windows.** `GOOS=windows GOARCH=amd64 go build ./...` passes; the section and the step lock are
  not exercised there (spec §F R3).
- **E6 — a lane killed inside the section.** Not tested (spec §F R1).

## Quality gate criteria

- `gofmt -l` over the changed Go files is empty; `go vet ./internal/cli ./internal/kanban ./internal/homestate`
  exits 0; `golangci-lint run` at the CI version (v2.1.6) over the three packages reports 0 issues;
  `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- @MX annotations: the critical-section function and `WithLock` carry `@MX:WARN` with `@MX:REASON` (a lock
  held across I/O) and `@MX:ANCHOR` where fan-in reaches 3, per `mx-tag-protocol.md`.
- Coverage of the changed functions is measured and reported per function (package-wide coverage is not
  claimed).
- Tool provenance: every `moai` or `go` figure cited in `progress.md` names the judging build's commit next
  to the tree's HEAD (`verification-claim-integrity.md` §2.2). At plan time the installed `moai` build was
  `45600e4ee` (`archive/t1401-293-g45600e4ee`, built 2026-10-02); `git rev-list --count 45600e4ee..HEAD`
  printed 187, and `git diff --name-only 45600e4ee HEAD -- internal/spec` printed nothing, so the lint
  source the installed build carries equals the tree's. The installed build is the judging build for
  `moai spec lint`; the tree's own build was not made.

## Definition of Done

1. WM1 RED output recorded verbatim in `progress.md` §E.2, each RED commit before its fix commit.
2. AC-FAL-001 to -009 green on the final tree with the repetition counts above, each mutant of plan §7
   executed and its failing test recorded.
3. AC-FAL-010, -011, -013 green; the preservation baseline and final counts both recorded.
4. The section's measured hold-time distribution recorded (REQ-FAL-011, spec §F R9).
5. `spec.md` §F unchanged in substance, or amended with the reason, by manager-spec; no residual window
   claimed closed that this card did not close.
6. Sync: the two Amendments records (AC-FAL-014) and the CHANGELOG entry (AC-FAL-012).
7. `moai spec lint SPEC-FACTORY-ATOMIC-LEASE-001` reports no findings.
