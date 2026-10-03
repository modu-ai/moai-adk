# SPEC-FACTORY-ATOMIC-LEASE-001 — Acceptance

Tier M. Every release-blocking criterion carries a **RED-now cell** (a read-only single-invocation
command, its verbatim output, its exit code, the pinned tree — held in the evidence ledger and cited by
row) and a **green-path cell** (the milestone that flips it and what the passing output becomes), per
`.claude/rules/moai/development/verification-completeness.md` §2 and §2.1, and a **mutation criterion**
(a named mutant of plan §7 that must turn a named test red). A criterion that cannot be red on arrival is
a **regression-guard**, is labelled so, and is never recorded as release-blocking. Given-When-Then is the
verification layer's format; the requirements are GEARS in `spec.md`.

Concurrency criteria run under repetition **and** `-race`, with the lane environment scrubbed in the same
compound call (setup row S1), under a heavy-run lease (S2), and are judged by the pass condition of S4,
not by an exit status. 14 criteria (Tier M ceiling 16), tracing all 13 requirements.

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
  Rows L1–L8 were observed on it in iteration 1 (the probes were mapped in with `go test -overlay`; the
  SPEC directory itself was untracked at the time, and no tracked file differed from the pin;
  `git merge-base develop HEAD` returned the same SHA). Rows L9–L18 were observed in iteration 2 on HEAD
  `db692601307c28b6d1dd905ab1ab6f6d9bd1e974`, whose Go files equal the pin (`git diff --name-only
  2de0a2cb613b04765a1554f86685a3b48e0be806..HEAD -- internal` printed nothing, exit 0); the iteration-2
  repair adds only files under `.moai/specs/SPEC-FACTORY-ATOMIC-LEASE-001/`.
- **S4 — the pass condition of every `go test` Command in this file.** Every Command carries `-v`, and
  its output is judged by what it printed, not by its exit status or its package line, because a selector
  that sweeps nothing prints `ok` and exits 0 (ledger L10 shows both, verbatim). A run passes only when:
  (a) each test name the criterion selects prints `--- PASS: <Name>` at the left margin **exactly N
  times**, N being the Command's `-count` value (subtests print indented lines and are not counted) — so
  a name that is mistyped, missing or skipped fails the criterion where an alternation would otherwise
  print `ok`; (b) there is no `--- FAIL` line and no `--- SKIP` line; (c) the output does not contain
  `no tests to run`; and (d) for a `-race` Command, no `DATA RACE` line. The **swept count** of a selector
  is recorded where its tests exist: the one selector that names existing tests today (AC-FAL-010) swept
  68 (L9); every other selector names tests that arrive in WM1 and swept 0 at the pin (L10), so its cell
  reads "to be recorded at WM1" and is not given a number here.

## Evidence ledger (RED-now observations and context)

Rows `L1`–`L8` were observed on the pin; `L9`–`L18` in iteration 2 on a tree whose Go files equal it. After
the run they are history by design: each describes the pinned tree and prints something else on a tree
that carries the linked milestone. Output is quoted exactly as the tool printed it; where a stream is
abridged the row says so and says what was elided. Every command below is the exact command that
produced the output beside it, with the anchored pattern `moai spec lint` asks for. Several probes are
timing- or interleaving-dependent, so a re-run prints different durations and counts; the sentence under
each row gives the range seen across all runs. A row whose command is a pipeline is marked **(context,
pipeline form)**: §2.1's single-invocation form excludes it, so it is never a RED-now cell.

### L1 — M1 (nominated), M2 (nominated, context), M3, and the final-name counter-observation

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

The M1 probe is the RED-now cell of AC-FAL-001 (nominated clause) and the M3 probe the RED-now cell of
AC-FAL-004 (its failing predicate `queue == picked` is the criterion's own final state, so a fix
flips it); both failed identically on all runs made from the committed probe file (and on the scratch run
before it). **The M2 probe is context, not the RED-now cell of AC-FAL-002:** it fails on the end state
`queue == hold && record == leased`, and that end state stays reachable after the fix (spec §F R10), so
it is red before and after for a reason the fix does not touch; and its seam calls the public `Mutate`
synchronously, which after the fix contends with the section's own lock. Clause (i) of AC-FAL-002 is
observed in L11. The fourth probe is a counter-observation, not a RED-now cell: creating the branch with
its final name passed 40 of 40 in this run and failed once in 40 in two others (`fatal: failed to read
.git/worktrees/wt-b/commondir`) — 2 failures in 280 iterations over seven runs, which is why spec §B.3
does not adopt it.

### L2 — M5 rename collision, real materializer, two lanes started together (abridged)

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

Four iteration-1 runs of the same step failed 6, 5, 8 and 10 of 20 iterations (29 of 80). The
`stranded-…=10` count is the number of failed steps that left a directory with no recorded worktree for
which `factoryRefuseForeignWorktree` returns a refusal; the earlier counted run gave 8 of 8. Iteration 2
re-ran it (L15) and the plan-audit ran it once (2 of 20, cited from it): across the eight unforced runs
the per-run rate was 5% to 50%. L13 measures the same step with the overlap forced.

### L3 — M2 (arm (c), context) and M4 (arm (a)), unmodified tree

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

The arm (a) probe is the RED-now cell of AC-FAL-005; both probes failed on both runs made with the final
probe file. The arm (c) hold probe is **context** for AC-FAL-003 for the reason given under L1 (its failing
end state stays reachable after the fix); clause (i) of AC-FAL-003 is observed in L11. The probes hang the
lane on the existing `factoryCardNow` variable at the first claim write, keyed on the calling function
names — a probe device; the real tests use a seam (plan WM1).

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
returned after 3.107 s. **Reading note kept from the plan-audit:** the "busy_timeout=200ms" lines here set
the value with a runtime `PRAGMA busy_timeout=200` after opening through `OpenFactoryPath` (DSN 5000 ms);
a DSN-carried value was never measured in this row — L14 measures it. Across three runs the uncontended
claim ranged p50 0.96–2.03 ms, p95 1.48–16.77 ms, max 7.78–72.52 ms (load-dependent); the stall lines
repeated within 0.1 s.

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
returning 1 on a zero count is the observed behavior of rows 3 and 5, stated here as an inference for row 4
(the plan-audit re-ran all five and printed `0` with exit 1 for rows 3 to 5, confirming the inference).

### L9 — the preservation family: the names and the swept count (AC-FAL-010 baseline at plan time)

Two commands, on HEAD `db692601307c28b6d1dd905ab1ab6f6d9bd1e974` (Go files equal to the pin), lane
environment scrubbed in the same call, heavy-run lease held (S1, S2).

```
$ go test ./internal/cli -list '^(TestFactoryNext|TestTodoLane|TestTodoNonLane|TestFactoryFallback|TestAutoPick|TestAutoRank|TestAutoHelp)'
[abridged: 68 test-name lines, then the package line; the names are exactly the alternation of the next command, in that order]
ok  	github.com/modu-ai/moai-adk/internal/cli	1.625s
exit 0
```

```
$ go test ./internal/cli -run '^(TestFactoryNextSerialMutualExclusivity|TestFactoryNextSkipsClassificationBlocked|TestFactoryNextParallelizableConcurrentLeases|TestFactoryNextRecordAndClaimRaceOnLeasedRow|TestFactoryNextDuplicateDispatchGuard|TestFactoryNextClaimRefusedMapsRace|TestFactoryFallbackDeclareRestoreLifecycle|TestFactoryFallbackQueryPrintsCountByLane|TestFactoryFallbackAllLanesCountsAcrossLanes|TestFactoryFallbackAllNeedsNoLaneBareStillDoes|TestFactoryFallbackDeclareRefusesUnknownTrigger|TestFactoryNextNominateLeasesNominee|TestFactoryNextNominateUnknownCard|TestFactoryNextFlagSet|TestFactoryNextNominateMCPParity|TestFactoryNextNominateConcurrentLanes|TestFactoryNextNominateSameCardExactlyOne|TestFactoryNextNominateRefusesKeepSet|TestFactoryNextArmCSkipsHoldMarker|TestFactoryNextNominateQuotaHold|TestFactoryNextNominateBackendSkip|TestFactoryNextNominateRecordStateTokens|TestFactoryNextNominateRefusalLeavesStateUnchanged|TestFactoryNextNominatePromoteThenLose|TestFactoryNextNominateClaimRefusedRollsBack|TestFactoryNextNominateCompensationFailure|TestFactoryNextNominateCompensateRechecksRecord|TestFactoryNextNominateBlankCardIsAnError|TestFactoryNextAllMarkerQueueExitsNoCard|TestFactoryNextBareUnchanged|TestTodoLaneRefusesAutoCycle|TestTodoLaneAutoRefusalText|TestTodoNonLaneGPTSessionNotRefused|TestFactoryFallbackDeclarePrintsLeasePath|TestFactoryNextExpiredLeaseReleasesSerialSlot|TestFactoryNextSerialSlotLeaseExpiryBoundary|TestFactoryNextLiveLeaseHoldsSerialSlotInEveryState|TestFactoryNextFailedSerialRowReleasesSlot|TestFactoryNextOwnAssignedSerialCardLeasesPastSiblingAssigned|TestFactoryNextOwnAssignedSerialCardBlockedByLiveSerialLease|TestFactoryNextAssignedSerialCardStillHoldsSlotAgainstNewTakes|TestFactoryNextAssignedSerialCardHoldsSlotInPickedArms|TestFactoryNextOwnAssignedSerialCardBlockedByPickedSibling|TestFactoryNextParallelizableLeasesBesideLiveSerial|TestFactoryNextPickedOwnerlessRowHoldsSlot_OutOfExpiryScope|TestAutoRankDoctrineAmendment|TestAutoRankMirrorParity|TestAutoRankMarkerDisclosure|TestAutoHelpAndRefusalDoNotAssertPickOrder|TestAutoRankAgentDoctrine|TestAutoPickDocDoctrine|TestAutoPickMirrorParity|TestAutoRankFallbackOrder|TestAutoRankDemotion|TestAutoRankJevFindingIsNotASignal|TestAutoRankUnmeasuredSignal|TestAutoRankBlockedExcluded|TestAutoRankSelectionRecord|TestAutoRankJevOrdering|TestAutoRankFallbackReasons|TestAutoRankJevMalformedAnswer|TestAutoRankRescueFirst|TestAutoRankQueueUnchanged|TestAutoRankNoQueueWriteGuard|TestAutoPickTargetsRelationBlocked|TestAutoPickTargetsReturnsAfterDone|TestAutoPickTargetsRescueArmUnfiltered|TestAutoPickTargetsNonSequencingRelation)$' -count=1 -v -timeout 25m
[output went to a scratch file; the deciding lines are quoted from it]
--- PASS: TestAutoPickTargetsNonSequencingRelation (0.73s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	757.794s
exit 0
```

The run's output was then counted with plain `grep -c` calls on the scratch file: `'^--- PASS: '` printed `68`,
`'^--- FAIL'` printed `0`, `-- '--- SKIP'` printed `0`, `'no tests to run'` printed `0`, `'DATA RACE'` printed
`0` (the SKIP count is `grep -c -- '--- SKIP'`, the `--` ending the options); and the set of passing top-level names sorted against the sorted `-list` names printed no difference
(`diff` exit 0). **Swept count: 68, equal to the 68 listed.** The 68-name alternation above is the
AC-FAL-010 selector. WM1 re-lists on the seam-and-stub tree and compares name for name before the baseline
is recorded.

### L10 — the selectors of the new criteria sweep nothing at the pin

```
$ go test ./internal/cli -run '^(TestFactoryLeaseSerialDistinctNomineesExactlyOne|TestFactoryLeaseSerialBareLanesExactlyOne|TestFactoryLeaseSerialCrossProcessExactlyOne)$' -count=2 -v -timeout 25m
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.661s [no tests to run]
exit 0
$ go test ./internal/cli -run '^(TestFactoryLeaseOperatorWriteWaitsForSection|TestFactoryLeaseArmCOperatorHold|TestFactoryLeaseCompensationKeepsOperatorPick|TestFactoryLeaseOwnAssignedSerialSiblingsExactlyOne|TestFactoryEnsureCardWorktreeConcurrentRealMaterializer|TestFactoryEnsureCardWorktreeStepLockBounded|TestFactoryWorktreeStepWaitDerivation|TestFactoryLeaseRecordStallBounded|TestFactoryLeaseMidClaimStallBounded|TestFactoryLeaseQueueLockStallBounded|TestFactoryLeaseCapWithinBoardBudget|TestFactoryLeaseSectionExcludesWorktreeStep|TestFactoryLeaseSectionRecordWritesPerArm|TestFactoryLeaseSectionAllowedSet)$' -count=1 -v -timeout 25m
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.684s [no tests to run]
exit 0
$ go test ./internal/homestate ./internal/cli ./internal/kanban -run '^(TestHomestateDoesNotImportKanban|TestFactoryLeaseSectionRejectsNestedMutate|TestLockedBacklogLoadIsPure)$' -count=1 -v -timeout 25m
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/homestate	0.497s [no tests to run]
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.710s [no tests to run]
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/kanban	0.459s [no tests to run]
exit 0
tree: db692601307c28b6d1dd905ab1ab6f6d9bd1e974
```

Exit status 0 and an `ok` line, with nothing run: this is the vacuous green that S4 refuses. At the pin every
selector of AC-FAL-001 to -009 and -011 fails S4(a) (0 passes against N expected) and S4(c); the cells below
cite this row as the criterion's own swept-count RED, and cite a behavioral row where one exists. Output
quoted verbatim; nothing elided.

### L11 — clause (i) of AC-FAL-002 and AC-FAL-003, observed red at the pin, with a positive control (abridged)

```
$ go test ./internal/cli -overlay=.moai/specs/SPEC-FACTORY-ATOMIC-LEASE-001/evidence/overlay-clause-i.json -run '^(TestT1458ProbeClauseINominated|TestT1458ProbeClauseIArmC|TestT1458ProbeClauseIControl)$' -count=3 -v -timeout 8m
[abridged to the first of three identical-in-predicate runs; the WARN lines and the second and third runs are elided]
=== RUN   TestT1458ProbeClauseINominated
    zz_t1458_clausei_probe_test.go:119: PROBE clause-i nominated: completed inside section = true operator-write-err=<nil> after the verb: queue=hold record=leased holder=lane-1 verb-err=<nil>
    zz_t1458_clausei_probe_test.go:121: clause (i) not met: the operator write completed inside the section (the seam returned only after the write finished)
--- FAIL: TestT1458ProbeClauseINominated (3.37s)
=== RUN   TestT1458ProbeClauseIArmC
    zz_t1458_clausei_probe_test.go:164: PROBE clause-i arm-c: completed at claim point = true operator-write-err=<nil> after the verb: queue=hold record=leased holder=lane-1 verb-err=<nil>
    zz_t1458_clausei_probe_test.go:169: clause (i) not met: the operator write completed at the claim point (arm (c) held no queue lock there)
--- FAIL: TestT1458ProbeClauseIArmC (3.10s)
=== RUN   TestT1458ProbeClauseIControl
    zz_t1458_clausei_probe_test.go:191: PROBE clause-i control (a write started while another Mutate holds the queue lock): completed inside section = false operator-write-err=<nil>
--- PASS: TestT1458ProbeClauseIControl (2.05s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	26.952s
exit 1
tree: db692601307c28b6d1dd905ab1ab6f6d9bd1e974
```

The two probes start the operator write **in a goroutine** from inside the seam (nominated form) or at the
claim point (arm (c)) and record whether it completed within 400 ms, i.e. before the seam returned; they
assert clause (i) alone and not the end state. All three runs read `true`, `true`, `false` in that order
(nominated, arm (c), control) — the predicate of AC-FAL-002 (i) and AC-FAL-003 (i) is red at the pin for the
stated reason (the section does not exist, so no lock is held at the seam). The control proves the 400 ms
window tells a held lock from a free one: a write started under another `Mutate`'s lock stayed pending across
it. After the fix the first two lines must read `false`, which is what the criteria's green-path cells state.

### L12 — M1 across two PROCESSES (bare form, held at the pass entry)

```
$ go test ./internal/cli -overlay=.moai/specs/SPEC-FACTORY-ATOMIC-LEASE-001/evidence/overlay-xproc.json -run '^TestT1458ProbeXProcSerialTwoLanes$' -count=10 -v -timeout 10m
[abridged to the ten PROBE lines; the interleaved "=== RUN", WARN, XPROC-RESULT, "--- FAIL" and "serial exclusivity breached…" lines are elided]
    zz_t1458_xproc_probe_test.go:189: PROBE xproc arm-c serial cards leased=1 rows=[t1=leased/lane-1]
    zz_t1458_xproc_probe_test.go:189: PROBE xproc arm-c serial cards leased=2 rows=[t1=leased/lane-2 t2=leased/lane-1]
    zz_t1458_xproc_probe_test.go:189: PROBE xproc arm-c serial cards leased=2 rows=[t1=leased/lane-2 t2=leased/lane-1]
    zz_t1458_xproc_probe_test.go:189: PROBE xproc arm-c serial cards leased=2 rows=[t1=leased/lane-2 t2=leased/lane-1]
    zz_t1458_xproc_probe_test.go:189: PROBE xproc arm-c serial cards leased=2 rows=[t1=leased/lane-1 t2=leased/lane-2]
    zz_t1458_xproc_probe_test.go:189: PROBE xproc arm-c serial cards leased=2 rows=[t1=leased/lane-2 t2=leased/lane-1]
    zz_t1458_xproc_probe_test.go:189: PROBE xproc arm-c serial cards leased=2 rows=[t1=leased/lane-2 t2=leased/lane-1]
    zz_t1458_xproc_probe_test.go:189: PROBE xproc arm-c serial cards leased=2 rows=[t1=leased/lane-2 t2=leased/lane-1]
    zz_t1458_xproc_probe_test.go:189: PROBE xproc arm-c serial cards leased=2 rows=[t1=leased/lane-2 t2=leased/lane-1]
    zz_t1458_xproc_probe_test.go:189: PROBE xproc arm-c serial cards leased=1 rows=[t1=leased/lane-1]
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	55.676s
exit 1
tree: db692601307c28b6d1dd905ab1ab6f6d9bd1e974
```

Eight of ten iterations leased two serial cards from two separate OS processes (the other two leased one: the
processes did not overlap enough that time). An earlier run of `-count=3` of the same probe printed two
breaches in three (`leased=1`, `leased=2`, `leased=2`). Combined: 10 breaches in 13 iterations, a 77% rate;
ten iterations took 55.7 s (about 5.5 s each). The probe is the RED-now cell of the cross-process clause of
AC-FAL-001 and the source of plan WM1's "what the helper needs".

### L13 — M5 with the two lanes' renames forced to overlap (deterministic overlap, collisions as corroboration)

```
$ go test ./internal/cli -overlay=.moai/specs/SPEC-FACTORY-ATOMIC-LEASE-001/evidence/overlay-barrier.json -run '^TestT1458ProbeM5BarrierRename$' -count=3 -v -timeout 20m
[abridged to the PROBE and result lines; 60 WARN lines are elided]
    zz_t1458_barrier_probe_test.go:128: PROBE M5-barrier iterations=20 creator-overlap-iterations=20 failed-iterations=15 rename-step-failures=15 create-step-failures=0 stranded-dir-and-foreign-worktree-precheck-refusal=15
--- FAIL: TestT1458ProbeM5BarrierRename (111.50s)
    zz_t1458_barrier_probe_test.go:128: PROBE M5-barrier iterations=20 creator-overlap-iterations=20 failed-iterations=13 rename-step-failures=13 create-step-failures=0 stranded-dir-and-foreign-worktree-precheck-refusal=13
--- FAIL: TestT1458ProbeM5BarrierRename (130.82s)
    zz_t1458_barrier_probe_test.go:128: PROBE M5-barrier iterations=20 creator-overlap-iterations=20 failed-iterations=11 rename-step-failures=11 create-step-failures=0 stranded-dir-and-foreign-worktree-precheck-refusal=11
--- FAIL: TestT1458ProbeM5BarrierRename (99.48s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	343.833s
exit 1
tree: db692601307c28b6d1dd905ab1ab6f6d9bd1e974
```

The probe wraps the real materializer so that each lane, once its worktree is created and before its
branch rename runs, waits until the other lane has also created (or 1.5 s pass). `creator-overlap-iterations`
counts iterations in which both lanes were inside the wrapper at once: **20 of 20 in each of three runs, 60
of 60** — the overlap is forced by construction and does not depend on how two git processes happen to
interleave. The collision itself is still a rate: 15, 13 and 11 of 20 here (55% to 75%), and 12, 13 and 15
of 20 in an earlier run of the same probe made before the overlap counter was added (79 of 120 over the six
runs). A failed step was always a rename failure (`create-step-failures=0`) and always left a refused
directory (`stranded…` equals `failed-iterations` in every run).

### L14 — the busy timeout carried in the DSN: what it bounds and what it does not

```
$ go test ./internal/homestate -overlay=.moai/specs/SPEC-FACTORY-ATOMIC-LEASE-001/evidence/overlay-homestate-dsn.json -run '^TestT1458ProbeDSNClaim$' -count=1 -v -timeout 4m
=== RUN   TestT1458ProbeDSNClaim
    zz_t1458_dsn_probe_test.go:93: PROBE DSN-carried handle busy_timeout=200ms
    zz_t1458_dsn_probe_test.go:98: PROBE claim DSN busy_timeout=200ms + NO ctx deadline vs 3s write holder: elapsed=3.0079325s err=<nil>
    zz_t1458_dsn_probe_test.go:107: PROBE claim DSN busy_timeout=200ms + ctx-deadline=800ms vs 3s write holder (run 1): elapsed=1.027451375s overshoot-past-deadline=227.451375ms err=context deadline exceeded
        database is locked (5) (SQLITE_BUSY)
    zz_t1458_dsn_probe_test.go:107: PROBE claim DSN busy_timeout=200ms + ctx-deadline=800ms vs 3s write holder (run 2): elapsed=1.010474458s overshoot-past-deadline=210.474458ms err=context deadline exceeded
        database is locked (5) (SQLITE_BUSY)
    zz_t1458_dsn_probe_test.go:107: PROBE claim DSN busy_timeout=200ms + ctx-deadline=800ms vs 3s write holder (run 3): elapsed=982.322625ms overshoot-past-deadline=182.322625ms err=context deadline exceeded
        database is locked (5) (SQLITE_BUSY)
    zz_t1458_dsn_probe_test.go:116: PROBE claim DSN busy_timeout=200ms + ctx-deadline=800ms vs 7s write holder: elapsed=984.171458ms err=context deadline exceeded
        database is locked (5) (SQLITE_BUSY)
    zz_t1458_dsn_probe_test.go:122: PROBE busy_timeout on the DSN-carried handle after cancelled calls=200ms
--- PASS: TestT1458ProbeDSNClaim (29.76s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/homestate	30.192s
exit 0
tree: db692601307c28b6d1dd905ab1ab6f6d9bd1e974
```

The value in the DSN lasts (200 ms after the cancelled calls, where a runtime PRAGMA fell back to 5000 ms
in L5). With no deadline the claim waited the 3 s holder out whole (`3.0079325s`, no error), because the
retry loop re-enters: the DSN timeout alone does not bound the claim. With an 800 ms deadline it returned
after 0.98–1.03 s, 182–227 ms past the deadline: the deadline bounds the claim and the DSN timeout bounds
the overshoot. Both are required (plan D2).

### L15 — M5 rename collision re-run (iteration 2)

```
$ go test ./internal/cli -overlay=.moai/specs/SPEC-FACTORY-ATOMIC-LEASE-001/evidence/overlay-tree.json -run '^TestT1458ProbeM5RenameCollision$' -count=3 -v -timeout 20m
[abridged to the PROBE and result lines; 60 WARN lines are elided]
    zz_t1458_probe_test.go:162: PROBE M5 iterations=20 failed-iterations=1 stranded-dir-and-foreign-worktree-precheck-refusal=1
--- FAIL: TestT1458ProbeM5RenameCollision (160.54s)
    zz_t1458_probe_test.go:162: PROBE M5 iterations=20 failed-iterations=6 stranded-dir-and-foreign-worktree-precheck-refusal=6
--- FAIL: TestT1458ProbeM5RenameCollision (193.61s)
    zz_t1458_probe_test.go:162: PROBE M5 iterations=20 failed-iterations=6 stranded-dir-and-foreign-worktree-precheck-refusal=6
--- FAIL: TestT1458ProbeM5RenameCollision (123.23s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	478.903s
exit 1
tree: db692601307c28b6d1dd905ab1ab6f6d9bd1e974
```

One of the three runs failed one iteration of twenty (5%): the unforced collision rate moves with load.
With L2 (6, 5, 8, 10), the plan-audit's run (2) and these (1, 6, 6): 44 of 160, per-run 5% to 50%.

### L16 — the import graph (context, pipeline form)

```
$ go list -deps ./internal/homestate | grep -c 'moai-adk/internal/kanban'
0
$ go list -deps -test ./internal/homestate | grep -c 'moai-adk/internal/kanban'
1
$ go list -deps ./internal/kanban | grep -c 'moai-adk/internal/homestate'
1
$ grep -n 'internal/kanban' internal/homestate/temp_parity_test.go
10:	"github.com/modu-ai/moai-adk/internal/kanban"
tree: db692601307c28b6d1dd905ab1ab6f6d9bd1e974
```

The layering holds for the package's non-test dependencies (0) and fails by exactly one edge when test files
are included (1: `temp_parity_test.go`, package `homestate_test`). `AC-FAL-011`'s import guard therefore
scopes to non-test files; the third line is the positive control that the listing finds an import that
exists. The exit statuses were not displayed for these lines; the printed counts are what the row cites.

### L17 — callers and counts (context, pipeline form)

```
$ grep -rlE 'homestate\.(OpenFactory|OpenFactoryPath)\(' internal cmd --include='*.go' | grep -v '_test.go' | sort > files-factory.txt
$ wc -l < files-factory.txt
14
$ grep -rlE 'kanban\.NewBacklogStore\(|todoStoreAt\(|todoReadStoreAt\(' internal cmd --include='*.go' | grep -v '_test.go' | sort > files-queue.txt
$ wc -l < files-queue.txt
10
$ cat files-factory.txt files-queue.txt | sort -u | wc -l
23
$ comm -12 files-factory.txt files-queue.txt
internal/cli/factory_card.go
$ grep -rn 'clearStaleLockAtPath(' internal --include='*.go' | grep -v _test
internal/kanban/integration_lock_mutation_windows.go:25:	return clearStaleLockAtPath(path, "integration mutation lock")
internal/kanban/board_lock_clear_windows.go:76:	return clearStaleLockAtPath(boardLockPath(root), "board lock")
internal/kanban/board_lock_clear_windows.go:92:func clearStaleLockAtPath(path, label string) (*ClearStaleReport, error) {
internal/kanban/slot_lease_mutation_windows.go:15:	return clearStaleLockAtPath(path, "slot lease mutation lock")
tree: db692601307c28b6d1dd905ab1ab6f6d9bd1e974
```

14 files open the factory record, 10 construct a queue store, 23 in union, 1 in both (spec §A.2 O10, which
replaces the unattributed 31 / 35 / 64 / 2). The last command lists the definition and three callers of the
Windows stale-clear primitive: none is the queue's `backlog.lock` or the new step lock (spec §F R3). The two
saved lists (`files-factory.txt`, `files-queue.txt`) live in the scratchpad of the run, not in the
repository; the counts are what the row cites.

### L18 — which build judged the tree (`verification-claim-integrity.md` §2.2)

```
$ moai version
[the banner box above the version line is elided]
 v3.2.0-rc.27   archive/t1401-504-g0732cc699   built 2026-10-03T03:34:50Z
$ git merge-base --is-ancestor 0732cc699 HEAD
exit 1
$ git merge-base --is-ancestor HEAD 0732cc699
exit 1
$ git diff --name-only HEAD 0732cc699 -- internal/spec
(no output)
exit 0
tree: db692601307c28b6d1dd905ab1ab6f6d9bd1e974
```

The installed `moai` build (`0732cc699`) and the tree's HEAD have diverged: neither is an ancestor of the
other, so the build is not a strict ancestor of the tree (and it is not older in the sense §2.2 treats as a
lag). `internal/spec` — the lint's source — is identical between them, so the installed build's lint result
equals the tree's. The tree's own build was not made.

## Criteria

Classification: AC-FAL-001 to -009 and -011 are **release-blocking** (AC-FAL-011's tests do not exist at the
pin, so it is red until WM1 and WM2 land them — it is not a guard that is green today); AC-FAL-010 and -013
are **regression-guards**; AC-FAL-012 and -014 are release-blocking at the sync phase (their RED-now cells
are in L8).

### AC-FAL-001 — two serial leases at once: exactly one wins (M1)

**Covers**: AC-FAL-001 maps REQ-FAL-001, REQ-FAL-002

**Given** a queue holding two queued serial cards `t1` and `t2`, no serial card in flight, and two
registered lanes,
**When** (a) the lanes nominate `t1` and `t2` in one test process and are held at the nomination seam by
the tolerant harness, (b) the lanes run bare `factory next` in one test process held at the pass entry, and
(c) two separate helper processes run bare `factory next`, each held at the pass entry by a file
rendezvous (plan WM1, "cross-process lane helper"),
**Then** exactly one record row is `leased`; in (a) the other lane exits 4 with one stderr line
`factory next: refused serial-slot: …`, in (b) and (c) it exits 3 with `no card is available`; the other
card has no `leased` row.

- **Command**: `go test ./internal/cli -run '^(TestFactoryLeaseSerialDistinctNomineesExactlyOne|TestFactoryLeaseSerialBareLanesExactlyOne|TestFactoryLeaseSerialCrossProcessExactlyOne)$' -count=10 -race -v -timeout 25m` (S1, S2). Pass condition: S4 with N = 10 — each of the three names prints `--- PASS` ten times.
- **RED-now cell**: L1 (`PROBE M1 serial cards leased = 2`, nominated clause), L4 (`leased=2`, 8 of 8, bare
  form) and L12 (`leased=2` in 8 of 10 iterations from two processes, 10 of 13 over both runs). Red for the
  stated reason: the in-lock check reuses the pre-lock `serialHeld`, arm (c)'s record snapshot predates both
  promotions, and no lock orders two processes. The selector's own swept count at the pin is 0 (L10), so
  S4(a) fails by construction as well. Swept count at WM1: to be recorded at WM1.
- **Green-path cell**: WM3. S4 holds: each named test `--- PASS` in all 10 repetitions, `DATA RACE` count 0.
- **Mutation**: MU1 (compute the serial slot before `WithLock`) turns the first two tests red; MU7 (a
  process-local `sync.Mutex` and no `flock`) leaves the first two green by construction — both lanes are
  goroutines of one process and the mutex serializes them — and turns the third red, which is the reason the
  third test exists. Detection power: the probe's per-iteration breach rate across processes was 77% (10 of
  13), so ten clean iterations of a process-local mutex have probability about 0.23^10 ≈ 4 × 10^-7 if
  iterations are independent (assumed, not tested) and about 10^-3 at a pessimistic 50% (plan R-I).

### AC-FAL-002 — an operator write cannot interleave with a nominated lease (M2)

**Covers**: AC-FAL-002 maps REQ-FAL-001, REQ-FAL-003

**Given** a queued nominee `t1` and a seam inside the lease that starts an operator `hold` of `t1` in a
goroutine,
**When** a lane nominates `t1`,
**Then** (i) the operator write has not completed when the seam returns (it is blocked on the section's
lock), (ii) after the verb returns the operator's hold is applied (queue `hold`, not lost), and (iii) a
hold committed before the verb starts is refused with `refused held`, writing no record row and changing
nothing in the queue.

- **Command**: `go test ./internal/cli -run '^TestFactoryLeaseOperatorWriteWaitsForSection$' -count=20 -race -v -timeout 25m`. Pass condition: S4, N = 20.
- **RED-now cell**: L11, first test: `completed inside section = true`, observed at the pin with a positive
  control that reads `false` under a held queue lock — clause (i) itself, red because no lock is held at the
  seam. The iteration-1 probe of L1 (`queue=hold record=leased`) is context only: it measures an end state
  the fix leaves reachable (spec §F R10). Clause (iii) is a regression-guard (green today) and is not what
  makes the criterion red. Swept count at the pin: 0 (L10); at WM1: to be recorded at WM1.
- **Green-path cell**: WM3. Clause (i) reads `completed inside section = false` in all 20 repetitions.
- **Mutation**: MU2 (release the lock before the claim) turns clause (i) red.
- **Not claimed**: the end state `queue=hold record=leased` is still reachable when the operator's write
  arrives mid-section (spec §F R10); the criterion tests the ordering, not the end state.

### AC-FAL-003 — the same for the unnominated arm (c) (M2)

**Covers**: AC-FAL-003 maps REQ-FAL-001, REQ-FAL-003

**Given** queued parallelizable cards `t1`, `t2` and a hook that starts an operator `hold` of `t1` in a
goroutine at arm (c)'s claim point,
**When** a lane runs bare `factory next`,
**Then** (i) the operator write has not completed at the claim point, and is applied after the verb
returns; and (ii) a hold committed before the verb starts makes the bare form skip `t1` and lease `t2`.

- **Command**: `go test ./internal/cli -run '^TestFactoryLeaseArmCOperatorHold$' -count=20 -race -v -timeout 25m`. Pass condition: S4, N = 20.
- **RED-now cell**: L11, second test: `completed at claim point = true` at the pin, same control as AC-FAL-002.
  L3's arm (c) hold probe is context only, for the reason given under L1. Swept count at the pin: 0 (L10);
  at WM1: to be recorded at WM1.
- **Green-path cell**: WM3; clause (i) prints `completed at claim point = false` in all 20 repetitions.
- **Mutation**: MU2.

### AC-FAL-004 — the compensation keeps the operator's fresh pick (M3)

**Covers**: AC-FAL-004 maps REQ-FAL-004

**Given** a queued nominee, a seam that fails the claim, and a seam goroutine that after the promotion
writes the queue item to `queued` then to `picked` (an operator unpick and re-pick, as direct store
writes),
**When** a lane nominates the card,
**Then** the invocation fails with the injected error (not exit 4), the operator goroutine's writes have
not completed when the seam returns, and after the verb returns the card's queue state is `picked`, the
operator's last write, with no `leased` row.

- **Command**: `go test ./internal/cli -run '^TestFactoryLeaseCompensationKeepsOperatorPick$' -count=20 -race -v -timeout 25m`. Pass condition: S4, N = 20.
- **RED-now cell**: L1 (`PROBE M3 … queue=queued (the operator's fresh pick was picked)`): the probe's
  failing predicate `queue == picked` is this criterion's own final state, which the fix flips. Swept count
  at the pin: 0 (L10); at WM1: to be recorded at WM1.
- **Green-path cell**: WM3; final queue state `picked` in all 20 repetitions.
- **Mutation**: MU3 (compensation as a second public `Mutate` after the section) and MU2 each turn it red.

### AC-FAL-005 — two lanes lease their own assigned serial cards: exactly one (M4)

**Covers**: AC-FAL-005 maps REQ-FAL-005

**Given** serial cards `t1` and `t2` recorded `assigned` to `lane-1` and `lane-2`, nothing in flight,
**When** both lanes run bare `factory next` held at their claim (arm (a)),
**Then** exactly one card is `leased`; the other lane exits 3 and its card stays `assigned` with owner and
version unchanged.

- **Command**: `go test ./internal/cli -run '^TestFactoryLeaseOwnAssignedSerialSiblingsExactlyOne$' -count=20 -race -v -timeout 25m`. Pass condition: S4, N = 20.
- **RED-now cell**: L3 (`PROBE arm-a serial cards leased=2`). Swept count at the pin: 0 (L10); at WM1: to be
  recorded at WM1.
- **Green-path cell**: WM3; one `leased`, one exit 3, in all 20 repetitions.
- **Mutation**: MU1.

### AC-FAL-006 — concurrent worktree steps both succeed, and never overlap (M5)

**Covers**: AC-FAL-006 maps REQ-FAL-008

**Given** two leased cards with no recorded worktree in one repository, the real worktree materializer
(no stub) wrapped by an event-logging wrapper that, once its worktree is created, holds the lane until the
other lane has also entered the creator or a 1.5 s grace has passed (the forced overlap of L13), and a
`git` shim on `PATH` that logs the start and end of each call to the same append-only file,
**When** both lanes run the worktree step at the same instant, 12 iterations inside the test,
**Then** every step returns nil; each card records its own worktree path and its branch is `WT-<slug>`;
**in every iteration the event log shows no creator-enter of one lane between the other lane's
creator-enter and the end of that lane's `git branch -m`** (the step is serialized — this is the
deterministic criterion, and it needs no collision to happen); and, with the step lock held by the test, a
step called with a shortened wait returns the wait error within that wait plus 500 ms, creates no directory
and leaves the card's record row unchanged; and the wait budget constant is not smaller than lanes (10) ×
the worst observed step (2.9 s) × headroom (2).

- **Command**: `go test ./internal/cli -run '^(TestFactoryEnsureCardWorktreeConcurrentRealMaterializer|TestFactoryEnsureCardWorktreeStepLockBounded|TestFactoryWorktreeStepWaitDerivation)$' -count=3 -race -v -timeout 25m` (36 concurrent iterations). Pass condition: S4, N = 3.
- **RED-now cell**: L2 and L15 for the collision (unforced, 5% to 50% per run, 44 of 160 over eight runs) and
  **L13 for the overlap**: `creator-overlap-iterations=20` of 20 in three runs (60 of 60) with 11 to 15 of
  20 iterations failing (55% to 75%). The unforced rate is **not** the discriminator and the criterion does
  not rest on it: at its low end (5%) a clean run of 36 iterations of the unfixed tree has probability
  0.95^36 ≈ 16% (computed from the lowest observed rate; iterations assumed independent). The forced
  overlap's occurrence is deterministic, so the event-log clause is red on the unfixed tree in every
  iteration. Swept count at the pin: 0 (L10); at WM1: to be recorded at WM1.
- **Green-path cell**: WM5; 0 overlap iterations and 0 failed iterations in 36. A clean 36 bounds a fixed
  collision rate at about 8% (rule of three, 95%); the claim is "removed the observed collisions (44 of 160
  unforced, 79 of 120 forced) and the observed 2 of 280 final-name failures", not "zero". Under the fix the
  wrapper cannot force an overlap (the second lane waits at the lock for the first), so the harness is
  tolerant: it releases the held lane after the grace.
- **Mutation**: MU4 (no step lock) turns the concurrency test red deterministically (the barrier forces the
  overlap the event log then reports).
- **Control**: L7, the same step serialized in-process, 0 of 80 over four runs.

### AC-FAL-007 — a busy record or a busy queue lock bounds the lease and does not starve the queue (REQ-FAL-006)

**Covers**: AC-FAL-007 maps REQ-FAL-006, REQ-FAL-001

Let `C` be `factoryLeaseClaimWaitCap` (the claim deadline plus the claim busy timeout, plan D2; 1.0 s at the
starting values) and `B` be `kanban.LockWaitBudget()` (3.3 s). The margin used below is **500 ms**, a stated
heuristic and not a derivation: it covers one lock retry wait (at most 50 ms, `boardLockWaitMax`), one queue
mutation at the lock's sizing figure (33 ms), the overshoot of the claim past its cap measured in L14 (about
30 ms beyond the sum), and scheduler delay; WM4 records the observed maximum and may tighten it.

**Given** (a) a record write transaction held by a second connection for three times `C` and a queued
nominee; (b) the same stall started only when the claim's third record write is about to begin, after the
assign edge committed (the third `factoryCardNow` call of the claim, the probe device of L3); (c) the queue
lock held by the test for longer than `B`,
**When** (a) a lane nominates the card and, at the same instant, a queue writer (`Mutate`) starts; (b) a
lane nominates the card, then the same lane nominates it again once the stall has ended; (c) a lane
nominates the card, and a second lane runs bare `factory next`,
**Then**

- (a) the lease returns within `C` + 500 ms with exit 4 and `factory next: refused raced: …` whose detail
  says the record or the queue lock was busy and does not contain "another lane"; the queue item is `queued`
  again; the queue writer completes without a lock-held error; the queue's lock is acquirable immediately
  afterwards;
- (b) the first invocation returns within `C` + 500 ms of the stall's start with exit 4 and the same detail;
  the card's row is `assigned` to the lane and its queue item stays `picked` (spec §F R4, second shape); the
  second invocation leases the card (exit 0) — that last expectation is read from the code and is the one
  thing here not yet observed;
- (c) the nominated lease returns within `B` + 500 ms with exit 4 and a detail that names the queue lock; the
  bare form returns within `B` + 500 ms — **one attempt, not five** — with **exit 1**, the queue lock's
  timeout error on standard error, nothing on standard output (in particular not `no card is available`),
  and no record row written for any card (spec §H DL-1; the rejected re-selecting design would have run for
  about 5 × 3.3 s = 16.5 s before printing exit 3 with `no card is available`); and
- the bare form's arm (c) under (a)'s record stall returns within `C` + 500 ms with exit 1 and the promoted
  item left `picked` (spec §F R4, third shape), which the next bare pass after the stall adopts through arm
  (b) or (b2).

- **Command**: `go test ./internal/cli -run '^(TestFactoryLeaseRecordStallBounded|TestFactoryLeaseMidClaimStallBounded|TestFactoryLeaseQueueLockStallBounded)$' -count=10 -race -v -timeout 25m` (each repetition waits out about three caps for (a) and (b) and a little more than `B` twice for (c)). Pass condition: S4, N = 10.
- **RED-now cell**: L5 — a claim with a 500 ms deadline against a 3 s holder returned after `3.107069958s`: no
  bound exists; L14 shows the DSN timeout alone does not provide one either. The "queue writer is not
  starved" clause is green on arrival (today the claim holds no lock) and becomes the constraint on the fix;
  it is what MU5 breaks. Swept count at the pin: 0 (L10); at WM1: to be recorded at WM1.
- **Green-path cell**: WM3 + WM4; each lease returns in the stated bound in all repetitions.
- **Mutation**: MU5 (default 5000 ms busy timeout on the lease connection) turns the bound red; MU6 (drop the
  lock-timeout mapping) turns clause (c) red (a raw error in the nominated form, not exit 4); MU11 (map the
  bare form's lock timeout back to a re-selection) turns the bare-form clause of (c) red (the run lasts
  far beyond `B` + 500 ms and exits 3).

### AC-FAL-008 — the cap is derived from the queue lock's budget (REQ-FAL-006)

**Covers**: AC-FAL-008 maps REQ-FAL-006

**Given** the lease claim deadline and busy-timeout constants (`factoryLeaseClaimDeadline`,
`factoryLeaseClaimBusyTimeout`) and their sum `factoryLeaseClaimWaitCap`, **When**
`TestFactoryLeaseCapWithinBoardBudget` runs, **Then** `factoryLeaseClaimWaitCap` × 3 is not greater than
the queue lock's wait budget as `kanban.LockWaitBudget()` returns it, and the test reads that accessor
rather than a literal (the precedent is `slot_lease_cross_test.go`'s stall-release guard).

- **Command**: `go test ./internal/cli -run '^TestFactoryLeaseCapWithinBoardBudget$' -count=1 -v`. Pass condition: S4, N = 1.
- **RED-now cell**: L8 first row (`factoryLeaseClaimWaitCap` absent from the tree, exit 1); L6 gives the budget
  the guard will read, 3.3 s. Swept count at the pin: 0 (L10); at WM1: to be recorded at WM1 (at the WM1 seam
  commit the constant exists and the accessor is a stub, so the test fails with the stub's panic — the
  stated reason).
- **Green-path cell**: WM4; the test prints `--- PASS` once.
- **Mutation**: raising the claim deadline to 1.5 s turns it red.

### AC-FAL-009 — the section contains only the two stores, its writes are bounded per path, and its hold is measured (REQ-FAL-007)

**Covers**: AC-FAL-009 maps REQ-FAL-007, REQ-FAL-001

**Given** a lease through each path of the table below with the worktree creator stubbed, the record writes
counted (by the `factoryCardNow` hook keyed on the claim functions' names, the probe device of L3, which the
claim calls once per record write) and the lock state probed at each count (an operator write started at the
count and observed pending for 100 ms, the window of L11),
**When** the lease runs,
**Then**

(i) a `Mutate` issued from inside the creator stub completes within 500 ms (no queue lock is held during
worktree creation) and likewise at the card-worktree record write;

(ii) **per path, on a successful lease**, the section issues exactly the counts below, and **in every path**
at most one queue promotion, at most one queue restore and at most three factory-record write transactions,
none of them before the lock is held (the lock probe reads "pending" at every count):

| Path | queue promotions | queue restores | record write transactions |
|---|---:|---:|---:|
| bare arm (a) — the row is `assigned` to the lane | 0 | 0 | 1 |
| bare arm (b) — an unowned `picked` row, queue item `picked` | 0 | 0 | 2 |
| bare arm (b2) — a `picked` queue item with no row | 0 | 0 | 3 |
| bare arm (c) — a `queued` item | 1 | 0 | 3 |
| nominated — a `queued` item | 1 | 0 | 3 |
| nominated — a `picked` item with no row, or an unowned `picked` row | 0 | 0 | 3 |
| nominated — the row is `assigned` to the lane | 0 | 0 | 1 |
| nominated — a claim that fails after its promotion | 1 | 1 | at most 3 |

(a record write transaction is one `RecordPicked`, assign edge or lease edge; each opens an immediate
transaction, `RecordPicked` even when it changes no row — read, `withCardTx`, `card_transition.go` line 283.
Promotions and restores are read from the queue item's state at the hook points and after the verb; a second
write of an unchanged state is not observable and is not claimed);

(iii) **the allowed set**: between the verb's start and the creator stub's call, the section runs no `git`
subprocess (the shim's log holds no line before the stub's marker line) and writes no path under the project
root other than the queue store's and the factory record's own files (the set of paths created or modified
there, taken as a size-and-modification-time diff, is contained in the engine database, its write-ahead and
shared-memory files and its lock file, and the record database with its write-ahead and shared-memory files);
and the only read outside those stores is the foreign-worktree directory check
(`factoryRefuseForeignWorktree`) — **this last half is not mechanically observable by the test, is stated as
doctrine-only, and is reviewed at WM3 by listing the call sites of `exec.Command`, `os.ReadFile`, `os.Open`,
`os.Stat` and `os.Lstat` in `factory_card.go` against the section's functions**; and

(iv) the hold-time distribution of the section — 10 sequential leases, then 2, 4 and 10 concurrent lanes — is
recorded in `progress.md` §E.2 with p50, p95, max and the command that produced it (no wall-clock threshold
is asserted).

- **Command**: `go test ./internal/cli -run '^(TestFactoryLeaseSectionExcludesWorktreeStep|TestFactoryLeaseSectionRecordWritesPerArm|TestFactoryLeaseSectionAllowedSet)$' -count=20 -race -v -timeout 25m`. Pass condition: S4, N = 20.
- **RED-now cell**: L8 second row (`WithLock` absent, exit 1) — today no section exists, so the lock probe of
  clause (ii) cannot read "pending". The selector sweeps 0 tests at the pin (L10). Clause (i) alone is green
  on arrival; it is the constraint on the fix and is what the creator-inside-the-section mutant below breaks.
  Swept count at WM1: to be recorded at WM1.
- **Green-path cell**: WM2 + WM3; the per-path counters read the table's values with `writes_before_lock=0`.
- **Mutation**: moving the creator call inside the section turns clause (i) red; MU2 turns the lock probe of
  clause (ii) red; MU10 (one `git` subprocess inside the section) turns clause (iii) red.

### AC-FAL-010 — selection behavior is unchanged where nothing interleaves (REQ-FAL-009) — regression-guard

**Covers**: AC-FAL-010 maps REQ-FAL-009

**Given** the preservation family: the 68 test names that `go test ./internal/cli -list` prints for the
family `TestFactoryNext`, `TestTodoLane`, `TestTodoNonLane`, `TestFactoryFallback`, `TestAutoPick`,
`TestAutoRank` and `TestAutoHelp` — enumerated in ledger L9, which also records that one `-v` run of
exactly those names printed 68 `--- PASS` and no failure,
**When** the same anchored alternation is run after WM6,
**Then** it prints `--- PASS` once for each of the 68 names (a swept count of 68, not less), 0 `--- FAIL` and
0 `--- SKIP`, and the tests plan §5 moves keep every assertion line they had (checked by reading the diff);
and a fresh `-list` of the family, taken at WM1 on the seam-and-stub tree before the baseline is
recorded and again at WM6, differs from L9's names only by names the WM1 baseline entry explains.

- **Command** (two plain calls): the `-list` command and the `-run` command of ledger L9, with `-count=1 -v
  -timeout 25m` (S1, S2). The `-run` selector is the literal 68-name alternation of L9 — not a template —
  because the family exists today; a prefix pattern is not used because an anchored prefix selects nothing
  and an unanchored one also sweeps longer names. Pass condition: S4 with N = 1 for each of the 68 names,
  plus the swept count at least the baseline (68 at plan time; the WM1 baseline entry in `progress.md`
  §E.2 is the floor the final run is held to).
- **Classification**: regression-guard — green on arrival by design (L9 measured it: 68 of 68), never
  recorded as release-blocking. Its non-vacuity is the swept-count floor: a run that sweeps fewer tests than
  the baseline fails the criterion. No mutant is claimed for it.
- **Green-path cell**: n/a (green throughout); the criterion fails if any milestone turns it red. The run
  takes 757.794 s on the pin's machine (L9), so it is a WM6 closure step, not a per-commit check.

### AC-FAL-011 — lock order, layering and non-adopting reads hold (REQ-FAL-010, REQ-FAL-009's pure-read half)

**Covers**: AC-FAL-011 maps REQ-FAL-010, REQ-FAL-009

**Given** the tree, **When** `TestHomestateDoesNotImportKanban` runs, a test calls the public `Mutate` from
inside a lease section, and `TestLockedBacklogLoadIsPure` runs, **Then** `internal/homestate` has no
dependency path to `internal/kanban` **through its non-test files** (the guard lists `go list -deps` of the
package without `-test`; a test file that imports `internal/kanban`, as `temp_parity_test.go` does today, does
not turn it red — L16); the in-section `Mutate` returns the lock-held timeout error within the budget plus 500
ms (the margin AC-FAL-007 states; the self-contention guard of plan D4); and a queue in the legacy layout shows no layout change after a
`WithLock` that only calls `LoadPure` (the locked handle offers no adopting read, plan D1).

- **Command**: `go test ./internal/homestate ./internal/cli ./internal/kanban -run '^(TestHomestateDoesNotImportKanban|TestFactoryLeaseSectionRejectsNestedMutate|TestLockedBacklogLoadIsPure)$' -count=1 -v -timeout 25m`. Pass condition: S4, N = 1, one name per package.
- **Classification**: **release-blocking until the three tests exist and pass.** None of them exists at the
  pin: the layering holds today (spec §A.2 O8; L16: 0 non-test dependencies), but the guard that would catch a
  regression has not been written, so a green "regression-guard" label would have been a vacuous claim.
- **RED-now cell**: L10 third command — the selector sweeps 0 tests in all three packages, so S4(a) and S4(c)
  fail at the pin. That is the stated reason it is red now; the criterion's green is the three tests.
- **Green-path cell**: WM1 (stubs, tests), WM2 (`TestLockedBacklogLoadIsPure`, the primitive) and WM3
  (the nested-`Mutate` test against a real section); S4 holds.
- **Mutation**: MU9 (an import of `internal/kanban` added to a non-test file of `internal/homestate`) must turn
  `TestHomestateDoesNotImportKanban` red, and MU8 (the locked read made adopting) must turn
  `TestLockedBacklogLoadIsPure` red; both are executed in WM6 and the failing test names recorded.
- The ordering half of REQ-FAL-010 (the lease takes the queue lock before any record write) is carried by
  AC-FAL-009 clause (ii).

### AC-FAL-012 — the records state what is not closed (REQ-FAL-011)

**Covers**: AC-FAL-012 maps REQ-FAL-011

**Given** the run and sync records, **When** read at the sync phase, **Then** `spec.md` §F still lists
R1–R15; the `progress.md` §E.4 signal and this SPEC's `CHANGELOG.md` entry each name which of M1–M5 closed
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

**Covers**: AC-FAL-013 maps REQ-FAL-012

**Given** the card branch, **When** measured at WM6 against `git merge-base develop HEAD` re-derived at
reading time, **Then** the changed paths are only: `internal/cli/factory_card.go` and its test files,
`internal/kanban` lock files and their tests, `internal/homestate/factory.go` and its test files, and
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

**Covers**: AC-FAL-014 maps REQ-FAL-013

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
  restored (spec §F R4, first shape); unchanged and not tested here.
- **E5 — Windows.** `GOOS=windows GOARCH=amd64 go build ./...` passes; the section and the step lock are
  not exercised there (spec §F R3, R14).
- **E6 — a lane killed inside the section.** Not tested (spec §F R1).
- **E7 — a card leased and then failed at the worktree step.** Not tested; it stays leased without a worktree
  until its lease runs out (spec §F R15).

## Quality gate criteria

- `gofmt -l` over the changed Go files is empty; `go vet ./internal/cli ./internal/kanban ./internal/homestate`
  exits 0; `golangci-lint run` at the CI version (v2.1.6) over the three packages reports 0 issues;
  `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- @MX annotations: the critical-section function and `WithLock` carry `@MX:WARN` with `@MX:REASON` (a lock
  held across I/O) and `@MX:ANCHOR` where fan-in reaches 3, per `mx-tag-protocol.md`.
- Coverage of the changed functions is measured and reported per function (package-wide coverage is not
  claimed).
- Tool provenance: every `moai` or `go` figure cited in `progress.md` names the judging build's commit next
  to the tree's HEAD (`verification-claim-integrity.md` §2.2). At iteration 2 the installed `moai` build was
  `0732cc699` (`v3.2.0-rc.27`, built 2026-10-03T03:34:50Z), diverged from HEAD — neither an ancestor of the
  other — and `git diff --name-only HEAD 0732cc699 -- internal/spec` printed nothing (L18), so the lint
  source the installed build carries equals the tree's. (Iteration 1 used build `45600e4ee`, 187 commits
  behind HEAD, with the same empty `internal/spec` diff.) The installed build is the judging build for
  `moai spec lint`; the tree's own build was not made.

## Definition of Done

1. Three commits, in this order, before any behavior changes: the seam-and-stub commit (plan WM1 step 1),
   the baseline commit (the AC-FAL-010 `-list` dump and pass count taken on the seam-and-stub tree,
   recorded in `progress.md` §E.2), then the RED commit. WM1 RED output recorded verbatim in `progress.md`
   §E.2 with each selector's swept count, each later RED commit before its fix commit.
2. AC-FAL-001 to -009 and -011 green on the final tree with the repetition counts above, each judged by
   S4's pass condition, each mutant of plan §7 executed and its failing test recorded.
3. AC-FAL-010 and -013 green; the preservation baseline and final counts both recorded (swept count at least
   68).
4. The section's measured hold-time distribution recorded (REQ-FAL-011, spec §F R9).
5. `spec.md` §F unchanged in substance, or amended with the reason, by manager-spec; no residual window
   claimed closed that this card did not close.
6. Sync: the two Amendments records (AC-FAL-014) and the CHANGELOG entry (AC-FAL-012).
7. `moai spec lint SPEC-FACTORY-ATOMIC-LEASE-001` reports no findings.
