# SPEC-FACTORY-ATOMIC-LEASE-001 — Progress

Card t1458, Tier M. This file is the phase record. Plan-phase writes only §E.1; §E.2 and §E.3 belong to the
run phase (manager-develop) and §E.4 to the sync phase (manager-docs).

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-10-03
- plan_start_head: 2de0a2cb613b04765a1554f86685a3b48e0be806
- artifacts: spec.md (14 requirements), plan.md (6 work milestones), acceptance.md (15 criteria), evidence/
  (ten probes and their overlays, re-executable from the repository root)
- plan_audit: iteration 1 FAIL 0.79; iteration 2 FAIL 0.81; iteration 3 FAIL 0.83; four-hunk confirmation
  not run — to be run by plan-auditor; no verdict is claimed here. (Iteration 1: Tier M threshold 0.80,
  audited tree `db692601307c28b6d1dd905ab1ab6f6d9bd1e974`, no must-pass criterion failed, the score was
  driven by Testability 0.55 and Clarity 0.70. Iteration 2: audited tree
  `c8b716fed24564a685188dc94b3f446ac9fc79c8`, no must-pass criterion failed, three must-fix defects PA2-M1
  to PA2-M3 and seven PA2-N1 to PA2-N7; the leader granted one extra repair round and one delta audit past
  the Tier M ceiling of 2, and accepts no PASS-with-debt on it. Iteration 3, that delta audit: audited tree
  `9f73f4cdf7af24af493edfb9e629f70aac915133`, no must-pass criterion failed, one must-fix defect I3-M1 and
  three should-fix defects I3-S1 to I3-S3, five notes I3-N1 to I3-N5. The leader's "Decision 2" approved an
  exception to "a second ceiling hit parks the card": this repair of four hunks, then a re-read of those
  four hunks only; if that confirmation is also blocked the card is parked.)

### 2026-10-03 — iteration 2 repair (spec.md 0.2.0)

Each audit defect id, where it was fixed, and how a re-auditor checks it. Measurements made in this
iteration are ledger rows L9–L18 in `acceptance.md`, on HEAD `db692601307c28b6d1dd905ab1ab6f6d9bd1e974`
(Go files equal to the plan-start tree), judged by `moai` build `0732cc699` (L18).

| Defect | Where it was fixed | Check |
|---|---|---|
| D1 | spec REQ-FAL-006 restated as an outcome, mechanism removed from it; plan D2 states both mechanisms and why (L14); spec §A.2 O6 corrected (the iteration-1 probe used a runtime PRAGMA) and O5 cites `factory.go` 473–486 | read REQ-FAL-006 last sentences; plan D2; O6; L14 |
| D2 | acceptance AC-FAL-009 clause (ii): per-path table plus the invariant; checked against `factoryNextClaim`, `factoryNextRecordAndClaim`, `RecordPicked`, `withCardTx`. **One cell differs from the audit's table:** bare arm (b2) makes 0 promotions, not 1, because its queue item is already `picked` | AC-FAL-009 (ii) |
| D3 | spec REQ-FAL-007 names the foreign-worktree directory check as the only outside read; AC-FAL-009 clause (iii) pins the allowed set (the read half stated doctrine-only); plan D4 | REQ-FAL-007; AC-FAL-009 (iii) |
| D4 | acceptance S4 (pass condition) and `-v` on every Command; per-name PASS count N; AC-FAL-010 given a literal selector (L9, 68 names swept, run exit 0); L10 shows the new selectors sweep 0 at the pin | S4; L9; L10 |
| D5 | AC-FAL-002/003 RED cells now cite L11 (clause (i), goroutine form, positive control; `completed inside section = true` at the pin); L1/L3 M2 probes relabeled context; evidence `probe-clause-i_test.go.txt`, `overlay-clause-i.json` | L11; AC-FAL-002/003 |
| D6 | plan §7 MU7 (and MU8–MU11); plan WM1 "cross-process lane helper (what it needs)"; AC-FAL-001 clause (c) and `TestFactoryLeaseSerialCrossProcessExactlyOne`; L12 (10 of 13 breaches from two processes); spec §H DL-3; evidence `probe-xproc_test.go.txt` | AC-FAL-001; plan WM1, §7; L12 |
| D7 | spec REQ-FAL-003 adds "within the queue lock's wait budget" and cites §F R6 | REQ-FAL-003 |
| D8 | spec REQ-FAL-006 bare-form outcome (end the pass at once, error, never an empty-queue report); §H DL-1 for the leader (worst cases: rejected design 5 × 3.3 s = 16.5 s; adopted one budget plus ≤ 50 ms); AC-FAL-007 (c) single outcome with a bound; plan D2; spec §F R13 | §H DL-1; AC-FAL-007 (c) |
| D9 | plan WM1 test list adds `TestHomestateDoesNotImportKanban` (non-test files only; L16: 0 vs 1 with `-test`) and `TestFactoryLeaseSectionRejectsNestedMutate`; AC-FAL-011 reclassified release-blocking until the tests exist (L10); spec REQ-FAL-010 scoped | AC-FAL-011; plan WM1; L16 |
| D10 | plan WM1 now three commits: seam-and-stub (compile-only stubs for every symbol a test names), baseline, RED; the AC-FAL-010 baseline is taken before the RED commit; Definition of Done item 1 | plan WM1; DoD 1 |
| D11 | spec §F R3 corrected (no stale clear on acquisition; the clear is wired to three other locks only, L17), R14 (step-lock crash), R15 (leased without a worktree and its recovery, with the audit's pointer to `factory_lane_relaunch.go` ~110 corrected: that line follows a new lease and is not a recovery); plan D3 failure outcomes; spec §D new Out-of-Scope bullets | §F R3, R14, R15 |
| D12 | AC-FAL-006: deterministic overlap criterion (ordered event log) with the forced-overlap probe (L13: 60 of 60 iterations overlapped); statistics restated as ranges (unforced 5% to 50% over eight runs, forced 55% to 75% over six; "one in ten million" withdrawn); spec §A.1 M5 row | AC-FAL-006; L13; L15 |
| D13 | spec §F R4 second shape (T2 landed, T3 timed out: row `assigned`, item `picked`, verb says `raced`) and third shape (arm (c) has no compensation); REQ-FAL-006 detail wording; AC-FAL-007 (b) mid-claim stall | §F R4; AC-FAL-007 (b) |
| D14 | spec REQ-FAL-004 reworded into the positive form | REQ-FAL-004 |
| D15 | AC-FAL-007 margin stated once (500 ms, a labeled heuristic); `factoryLeaseClaimWaitCap` defined in plan D2 as deadline plus busy timeout and used by AC-FAL-007/-008 | plan D2; AC-FAL-007/-008 |
| D16 | spec O13 before O14, §F R1–R15 in order; O10 replaced by re-measured, attributed counts (L17) | spec §A.2, §F |
| D17 | every `Covers:` line begins with its AC id | acceptance.md |
| D18 | plan D1: the locked handle offers `LoadPure`, no adopting read; spec REQ-FAL-009 names the non-adopting reads; `TestLockedBacklogLoadIsPure` in AC-FAL-011 | plan D1; AC-FAL-011 |

Decisions for the leader are in spec §H (DL-1 bare-form outcome, DL-2 nominated-form outcome, DL-3
cross-process helper). Not observed in this iteration, stated so nobody reads silence as a pass: the
mutants (none can run before implementation); any behavior after the fix; Windows; a multi-lane stall;
whether anything issues the request that applies a lease expiry to a card leased without a worktree.

### 2026-10-03 — iteration 3 (override round) repair (spec.md 0.3.0)

Defect ids below are the plan-audit's iteration-2 ids, written `PA2-…` so they do not collide with the
card's symptom ids M1–M5. Observations are ledger rows L19–L23 in `acceptance.md`, made on HEAD
`c8b716fed24564a685188dc94b3f446ac9fc79c8` (Go files equal to the plan-start tree) plus this round's
uncommitted evidence files; the installed `moai` build is `0732cc699`, diverged from HEAD (neither is an
ancestor of the other) with an identical `internal/spec` (L22). No Go source, other SPEC, doctrine file or
`.moai/reports/` file was edited.

| Finding | Where it was fixed | How a re-auditor checks it |
|---|---|---|
| PA2-M1 (drift-log lock wait is unbounded) | spec §A.2 O15 (source reading + L19), REQ-FAL-006 (time bound includes it), REQ-FAL-007 (the named, bounded exception), new REQ-FAL-014, REQ-FAL-009 and -012 (carve-out and scope), §E (REQ-FR-025 narrowed), §F R16, §H DL-5/DL-7; plan D2 (third device, bounded flow, lock order), WM1 stub, WM4 (the milestone that changes `homestate`), §6 R-K/R-L, §7 MU15–MU17; acceptance AC-FAL-015 (new, release-blocking), AC-FAL-009 (iii), L19, L21, L23 | read REQ-FAL-014 and AC-FAL-015; re-run L19's command (red: elapsed 3.0 s against a 1.3 s limit, skip predicate red, retry green) |
| PA2-M2 (AC-FAL-009 table) | acceptance AC-FAL-009 (ii): table rebuilt with queue-item and row state per row, `queued`/`picked` split for the assigned-row nominee, two failed-claim rows (before and after the assign edge), fixture-builder column | compare each row with `factoryNextClaim`, `factoryNextRecordAndClaim`, `RecordPicked`, `factoryNominateCompensate` (factory_card.go 604–650, 972–1003; card_picked.go 114–177) |
| PA2-M3 (REQ-FAL-003 vs REQ-FAL-009, arm (a)) | spec REQ-FAL-003 second clause narrowed, REQ-FAL-009 (arm (a) preserved), §A.1 M2 row, §B.2, §F R17, §D (arm (a) out of scope), §H DL-6; acceptance AC-FAL-003 clause (iii) (guard) and L20; plan MU18 | read REQ-FAL-003; L20 (`queue=hold record=leased`, 3 of 3) |
| PA2-N1 (DL-1 reasoning) | spec §H DL-1 restated (visible error vs silent stop; both loops end on an error), §F R5, §D (lane-loop retry out of scope) | read DL-1 and R5 against `factory_lane_relaunch.go` 76–79, 103–109 and `codex_launcher.go` 988–994 |
| PA2-N2 (`raced` supersession) | spec §E new row; acceptance AC-FAL-014 | read the §E row; AC-FAL-014 Then |
| PA2-N3 (legacy adoption) | spec §F R13 and REQ-FAL-009 (store construction, before any lock; the race named); plan D1 (the store is constructed before `WithLock`) | read R13 against `todo.go` and `state_dir.go` |
| PA2-N4 (AC-FAL-010 vs plan §5) | acceptance AC-FAL-010 Then (the one recorded removal is excepted); plan §5 intro | read both |
| PA2-N5 (plan §7) | plan §7: MU8 → AC-FAL-011; MU12, MU13, MU14 added (the audit's two AC-only mutants and N14(iii)); MU15–MU18 added for this round | read the §7 table against the acceptance Mutation lines |
| PA2-N6 (absolute sentences) | spec §A.1 M5 row; acceptance L13 text | read both sentences against the re-execution counts they cite |
| PA2-N7 (panic stubs) | plan WM1 (stubs return a zero value or a sentinel error; `release` never nil; homestate marker is the identity); acceptance AC-FAL-008 cell | read plan WM1 step 1 |
| PA2-N8 (D5 not in §H) | spec §H DL-4 (open, non-blocking; now three SPECs); plan D5 | read §H |
| PA2-N9 (shortened wait vs constant) | plan D3 (`factoryWorktreeStepWaitDefault` const + `factoryWorktreeStepWait` var), WM1; acceptance AC-FAL-006 | read plan D3 |
| PA2-N10, PA2-N11 (AC-FAL-014) | acceptance AC-FAL-014: Mutation line added; Command gains `prior_completed_sha` and `status:` counts | read AC-FAL-014 |
| PA2-N12 (probe ranges) | spec §A.2 O6, O14; acceptance L7, L14 (ranges stated as seen over named runs) | read the sentences |
| PA2-N13 (REQ-FAL-007 "read") | spec REQ-FAL-007 ("file or process I/O", environment and clock excepted); acceptance AC-FAL-009 (iii) | read both |
| PA2-N14 (mutant holes) | only (iii) taken: plan §7 MU14; holes (i) and (ii) have no criterion and are named as not claimed (plan §7 note) — skipped because each needs a new criterion clause and fixture, not a one-line edit | read plan §7 note |
| PA2-N15 ("only" unmeasured) | acceptance AC-FAL-013: allowed set enumerated, a complement probe added, sync-phase paths stated as outside it | read AC-FAL-013 |
| PA2-N16 (volume) | nothing asked; the round added L19–L23, REQ-FAL-014 and AC-FAL-015 | — |

Observations made in this round (commands and verbatim output in the ledger): L19 — drift-log lock held
3 s, claim under an 800 ms deadline: 3.016 s, 3.009 s, 3.002 s (exit 1, red for the stated reason); the
uncontended reconciliation costs 0.002–0.028 s (1 entry), 0.018–0.100 s (200), 0.151–0.260 s (2000). L20 —
arm (a) leases a held card with an assigned row, 3 of 3. L21 — the four new test names sweep 0 at the pin.
L22 — tool provenance. L23 — the six existing drift-log tests pass (6 of 6). Also: `grep -c
SPEC-FACTORY-ATOMIC-LEASE-001 .moai/specs/SPEC-FACTORY-RECORD-001/spec.md` printed `0`, exit 1.

Decisions taken in this round that the leader should see: spec §H DL-7 (non-blocking try rather than a
deadline; opt-in per call for the claim's three writes rather than global; both left open by the leader's
design and chosen to change the fewest other callers' behavior); the new third completed SPEC
(SPEC-FACTORY-RECORD-001, REQ-FR-025's "next successful write reconciles", narrowed for the claim's
writes) that the Amendments mechanism now reopens — DL-4's cost grew from two SPECs to three.

Not observed in this round, stated so nobody reads silence as a pass: any behavior after the fix; the
lease-level tests of AC-FAL-015 (they arrive in WM1; L19 is the record-write-level probe); the mutants
(none can run before implementation); Windows (the non-waiting form is compile-verified only once built);
a drift log larger than 2000 entries; whether another process's log append is delayed in practice by the
held lock; the `moai` build from this tree (the installed `0732cc699` judged the lint); the plan-audit's
own L13 and L14 re-executions that §A.1 and L13/L14 cite (they are reported to this author, not in the
ledger).

### 2026-10-03 — exception repair (iteration 4) (spec.md 0.3.1)

Scope: exactly the four hunks of the leader's "Decision 2" (the plan-audit's iteration-3 ids I3-M1 and
I3-S1 to I3-S3; the notes I3-N1 to I3-N5 were left alone, they are optional). No Go source, other SPEC,
doctrine file or `.moai/reports/` file was edited; the HEAD this repair read and measured is
`9f73f4cdf7af24af493edfb9e629f70aac915133` (Go files equal the plan-start tree). Tool provenance for
`moai spec lint` (`verification-claim-integrity.md` §2.2): the installed build is `0732cc699`
(`v3.2.0-rc.27`, built 2026-10-03T03:34:50Z, from `moai version`); `git merge-base --is-ancestor 0732cc699
HEAD` and `git merge-base --is-ancestor HEAD 0732cc699` each exited 1, so the build is not a strict ancestor
of HEAD, and `git diff --name-only HEAD 0732cc699 -- internal/spec` printed nothing, so the lint source it
carries equals the tree's; the tree's own build was not made. With that build, `moai spec lint
SPEC-FACTORY-ATOMIC-LEASE-001` and the same with `--strict` each printed `✓ No findings — all SPEC
documents are valid` and exited 0 after the edits above.

| Finding | Where it was fixed | How a re-reader checks it |
|---|---|---|
| I3-M1 (AC-FAL-015 (i)–(iii) unsatisfiable at the verb) | acceptance AC-FAL-015: a new "Observation points" paragraph, five fixtures (a)–(e) each naming its level, clauses (i)–(iii) restated at the return of `factoryNextNominate` / `factoryNextLeaseOnceGated` with "exit 0 / the verb's output" expressly not observed there, new clause (vii) (the verb's own card-worktree record write waits for the held lock and then reconciles, with probe B's measured expectation), Command, RED-now, green-path, mutation and Not-claimed cells updated; ledger L24; plan §7 MU20, WM1 test list | read AC-FAL-015 "Observation points" and clauses (i)–(iii), (vii) against `factory_card.go` 1083–1112 and 407–430 (the verb calls `factoryEnsureCardWorktree` only after a lease function returned, and that step ends in `db.RecordCardWorktree`); re-run L24's first command shape (the probe is a scratch file, see L24) |
| I3-S1 ("no entry reconciled twice") | spec REQ-FAL-014 (the over-claim replaced by the re-read guarantee); spec §F R16 (the cross-writer window named, audit-measured 2 duplicate events in 3 of 3 runs); acceptance AC-FAL-015 fixture (e) and clause (vi); plan §7 MU19, D2 step 3, WM1 seam and test list | read REQ-FAL-014's third clause and R16's "second window"; AC-FAL-015 (vi) and MU19 |
| I3-S2 (REQ-FAL-009 carve-outs) | spec REQ-FAL-009: the lead-in no longer names the skip; the four carve-outs are a numbered list | read REQ-FAL-009; each of the four items equals one of the four things the audit listed |
| I3-S3 (DL-6 provenance) | spec §H: DL-6 and DL-7 under "Decided (leader, 2026-10-03)" with the provenance "decision.md Decision 2"; the "Taken by the repair author" heading removed (it would be empty); DL-7 carries the leader's added instruction; spec §F R18 added for the limit | read §H DL-6, DL-7 and R18; DL-4 is still open |

Mechanical consequences of those four, listed so none is a surprise: spec HISTORY 0.3.1 and the frontmatter
`version`; spec §B.5 and the §F introduction mention R18; acceptance AC-FAL-012 counts R1–R18; acceptance
S3 and the ledger header name L24; plan.md's header note and §6 R-K. The counts stay at 14 requirements and
15 criteria (Tier M ceilings 16 and 16); the mutants are now MU1–MU20.

Observations made in this repair (commands and verbatim output are ledger row L24): the card-worktree
record write, with one unreconciled entry and the drift log's lock held 2 s, returned after 2.0007 s,
2.0015 s and 2.0006 s with `unreconciled-after=0` (3 of 3; the plan-audit's own probe B printed 2.004 s,
2.001 s and 2.003 s — audit-measured, not a ledger row); the two test names added in this repair
(`TestRecordWriteReconcileBoundedRereadsUnderLock`, `TestFactoryLeaseDriftLogVerbWorktreeWriteWaits`)
sweep 0 tests at the pin. The probe is a scratch file outside the repository (`evidence/` was outside the
file list of this repair); the exit status of both `go test` runs was not captured (the worktree guard
refuses a trailing `echo`), and the rows say so.

Decisions taken in this repair beyond the leader's record, for the leader to see: (1) the test seam of
AC-FAL-015 fixture (e) — an inert package variable between the log's unlocked read and the claim's try for
its lock — is a new line item of the WM1 seam-and-stub commit, because no deterministic test of the
re-read step exists without one; (2) the cross-writer window is named in R16 and the leader's DL-7 limit
is a separate R18, not folded into R16, so each reads as one residual; (3) new clauses (vi) and (vii)
were appended after (v) so no existing clause number, and no cross-reference to one, changed; (4) two
mutants were added, MU19 (no re-read under the lock) and MU20 (the skip reaching the verb's worktree write),
the second being the "connection-scoped cousin" of MU16 the audit named; (5) the verb-level fixture holds
the lock 2 s, the value of the measured probe, and bounds the verb's return by that hold plus the same
500 ms margin as AC-FAL-007, a labeled heuristic.

Not observed in this repair, stated so nobody reads silence as a pass: any test of AC-FAL-015 clauses (i)–(iii),
(vi) or (vii) (they arrive in WM1; clause (vi) has no RED-now observation and none is claimed — its test needs
the seam); the overlap of two live writers in the narrow gap between a commit and its mark (only the audit's
state-forcing probe A exists); the mutants; Windows; any behavior after the fix; the verb itself under a held
drift-log lock (L24 is record level); the four-hunk confirmation by a plan-auditor.

## §E.2 Run-phase Evidence

### WM1 commit 2 — AC-FAL-010 preservation baseline (card t1458)

Taken on the seam-and-stub tree (HEAD 883a3a205 with no test file added; Go files equal commit 0eb3d5b0d),
lane environment scrubbed in the same invocation (S1), slot `internal-cli-suite` held (S2). Tool: `go` from
PATH, not the project build.

- `-list` of the family (`^(TestFactoryNext|TestTodoLane|TestTodoNonLane|TestFactoryFallback|TestAutoPick|TestAutoRank|TestAutoHelp)`): exit 0, 68 test names. Compared with ledger L9's 68-name alternation (the
  AC-FAL-010 selector): identical name for name and in the same order (sorted `diff` exit 0; ordered diff
  differs only by the trailing newline). No name is added or removed, so no explanation entry is owed.
- `-run` of the literal 68-name alternation, `-count=1 -v -timeout 25m`: exit 0, package line
  `ok  github.com/modu-ai/moai-adk/internal/cli  367.501s`. `grep -c '^--- PASS: '` = **68**,
  `'^--- FAIL'` = 0, `-- '--- SKIP'` = 0, `'no tests to run'` = 0, `'DATA RACE'` = 0; the sorted passing
  top-level names diffed against the L9 names: exit 0 (no difference).
- **Floor for the final run: swept count 68.**
- Evidence (git-ignored, this worktree): `.moai/reports/t1458/baseline-list.txt`, `baseline-run.txt`.

### WM1 commit 3 — RED (card t1458)

Tests restored from the previous worker's draft and reviewed against plan WM1; they compile (`go vet` of
`internal/cli` and `internal/homestate` exit 0). Run on the seam-and-stub tree plus the new test files only,
lane environment scrubbed, slot held, `-count=1 -v -timeout 30m`.

- `internal/cli` `-run '^(TestFactoryLease|TestFactoryEnsureCardWorktree|TestFactoryWorktreeStepWaitDerivation)'`:
  exit 1, `FAIL internal/cli 102.300s`, no panic. 17 FAIL, 4 PASS, 1 SKIP (the cross-process helper, by design).
  FAIL: TestFactoryLeaseRecordStallBounded (3.45 s against a 1.5 s limit), TestFactoryLeaseMidClaimStallBounded,
  TestFactoryLeaseQueueLockStallBounded, TestFactoryLeaseCapWithinBoardBudget (LockWaitBudget stub is 0),
  TestFactoryLeaseSectionRejectsNestedMutate (in-section Mutate returned nil),
  TestFactoryLeaseOperatorWriteWaitsForSection, TestFactoryLeaseArmCOperatorHold,
  TestFactoryLeaseCompensationKeepsOperatorPick, TestFactoryLeaseSectionRecordWritesPerArm (writes_before_lock>0
  in every arm), TestFactoryLeaseSerialDistinctNomineesExactlyOne, TestFactoryLeaseSerialBareLanesExactlyOne,
  TestFactoryLeaseSerialCrossProcessExactlyOne (2 serial cards leased), TestFactoryLeaseOwnAssignedSerialSiblingsExactlyOne,
  TestFactoryEnsureCardWorktreeConcurrentRealMaterializer, TestFactoryLeaseSectionAllowedSet,
  TestFactoryLeaseDriftLogStallBounded, TestFactoryEnsureCardWorktreeStepLockBounded.
  PASS on the unfixed tree: TestFactoryLeaseArmAKeepsHeldAssignedCard, TestFactoryLeaseSectionExcludesWorktreeStep,
  TestFactoryLeaseDriftLogVerbWorktreeWriteWaits, TestFactoryWorktreeStepWaitDerivation.
- `internal/homestate` four tests: exit 1. FAIL: TestRecordWriteReconcileBoundedSkipsOnContention (1.503 s, want
  within 500 ms), TestRecordWriteReconcileBoundedRereadsUnderLock (seam never called; 1 duplicate event). PASS:
  TestRecordWriteReconcileDefaultStillWaits, TestHomestateDoesNotImportKanban.
- Evidence (git-ignored): `.moai/reports/t1458/red-cli.txt`, `red-homestate.txt`.

### WM2 — the primitive (card t1458)

Commit order (the graph witnesses it): RED `2997abc29` (tests only), then GREEN `f8cc4b978` (`internal/kanban`
`backlog_store.go`, `factory_step_lock.go`). Tool: `go` from PATH, lane environment scrubbed in each invocation.

- RED at `2997abc29` (`go test ./internal/kanban -run '^(TestWithLock|TestLockedBacklog|TestLockWaitBudget)'`): exit 1,
  8 FAIL, 0 PASS, no panic. Reasons: the WM2 sentinel error (6 tests), a panic that did not propagate out of
  `WithLock` (`TestWithLockReleasesOnPanic`), `LockWaitBudget() = 0s`. Evidence `wm2-red-kanban.txt`.
- GREEN at `f8cc4b978` (same selector, `-race`): exit 0, 8 PASS: TestWithLockMutationsSeeEachOther,
  TestWithLockHoldsLockUntilFnReturns (positive control for its 300 ms window included), TestWithLockReleasesOnErrorReturn,
  TestWithLockReleasesOnPanic, TestWithLockRefusesRelocatedQueue (acceptance edge E3), TestLockedBacklogMutateRefusalKeepsFileUnchanged,
  TestLockedBacklogLoadIsPure (AC-FAL-011 non-adopting half), TestLockWaitBudgetIsTheQueueLockBudget. Evidence `wm2-green-kanban-race.txt`.
- Mutant MU8 (handle read made `Load`, scratch edit restored byte-identical): `TestLockedBacklogLoadIsPure` FAIL (the adopting
  read inside the section stalled on the section's own lock, `lock ... held`), exit 1. Evidence `wm2-mu8.txt`. Not claimed:
  a mutant where the adopting read ran without self-contention.
- Preservation: full `go test ./internal/kanban -count=1 -timeout 30m`: exit 0, `ok ... 233.953s` (`wm2-kanban-pkg.txt`);
  `Mutate`'s existing tests unchanged and green. AC-FAL-010 (68-name selector, `-count=1 -v`, slot held): exit 0,
  `ok internal/cli 225.049s`, `--- PASS` = **68**, `--- FAIL` 0, `--- SKIP` 0, `no tests to run` 0, `DATA RACE` 0; sorted passing
  names diff against the WM1 baseline list: identical. Floor 68 met. Evidence `wm2-ac010-run.txt`.
- `gofmt -l internal/kanban` empty; `go vet` of kanban, cli, homestate exit 0; `golangci-lint run ./internal/kanban/...`
  (v2.1.6) exit 0, 0 issues; `GOOS=windows GOARCH=amd64 go build ./internal/kanban/... ./internal/cli/... ./internal/homestate/...`
  exit 0 and `GOOS=windows go vet ./internal/kanban` (compiles tests) exit 0.
- WM1 RED set re-run after WM2 (`wm2-cli.txt`, `wm2-homestate.txt`): moved RED to GREEN: `TestFactoryLeaseCapWithinBoardBudget`
  (needs `LockWaitBudget`). Still RED, by milestone: WM3 — SectionRejectsNestedMutate, OperatorWriteWaitsForSection,
  ArmCOperatorHold, CompensationKeepsOperatorPick, SectionRecordWritesPerArm, SerialDistinctNomineesExactlyOne,
  SerialBareLanesExactlyOne, SerialCrossProcessExactlyOne, OwnAssignedSerialSiblingsExactlyOne, SectionAllowedSet; WM4 —
  RecordStallBounded, MidClaimStallBounded, QueueLockStallBounded (with WM3), DriftLogStallBounded, and in
  `internal/homestate` BoundedSkipsOnContention, BoundedRereadsUnderLock; WM5 — EnsureCardWorktreeConcurrentRealMaterializer,
  EnsureCardWorktreeStepLockBounded. Unchanged green guards: ArmAKeepsHeldAssignedCard, SectionExcludesWorktreeStep,
  WorktreeStepWaitDerivation, DefaultStillWaits, HomestateDoesNotImportKanban; helper test skips by design.
- Load-sensitive guard, stated so nobody reads it as a WM2 regression: `TestFactoryLeaseDriftLogVerbWorktreeWriteWaits`
  passed at RED (returns 250 and 379 ms after the release, bound 500 ms) and failed once in the WM2 run (`bare`, 691.7 ms).
  Nothing in `internal/cli` calls `WithLock` yet. A `-count=3` re-run on the same machine (`uptime` load averages 53.95 39.42
  28.80, many concurrent `cli.test` processes from other lanes) passed 2 of 3 (the failing iteration `nominated`, 658.5 ms).
  The 500 ms margin is the plan's labeled heuristic and is not loosened here. Evidence `wm2-cli-rerun-driftverb.txt`.

### WM3 + WM4 — the lease section and the bounded claim (card t1458)

Commits in order: `fd70b9123` (homestate: bounded drift-log reconcile, DSN busy-timeout open `OpenFactoryBounded`,
non-waiting admission lock, settle step in `withCardTx`), `2016bad1a` (test repair, below), `daed5dd09` (`factory_card.go`
section + bounded claim, moved tests). Tool: `go` from PATH, lane environment scrubbed in each invocation (scratchpad scripts).

- RED to GREEN, `internal/cli` (`wm3-cli-lease2.txt`, `wm3-ac010-final.txt`, `wm4-cli-lease-race.txt`): SectionRejectsNestedMutate,
  OperatorWriteWaitsForSection, ArmCOperatorHold, CompensationKeepsOperatorPick, SectionRecordWritesPerArm,
  SerialDistinctNomineesExactlyOne, SerialBareLanesExactlyOne, SerialCrossProcessExactlyOne, OwnAssignedSerialSiblingsExactlyOne,
  SectionAllowedSet, RecordStallBounded, MidClaimStallBounded, QueueLockStallBounded, DriftLogStallBounded. `internal/homestate`
  (`wm4-homestate-reconcile.txt`, `-race`, exit 0): BoundedSkipsOnContention, BoundedRereadsUnderLock. Guards still green:
  ArmAKeepsHeldAssignedCard, SectionExcludesWorktreeStep, WorktreeStepWaitDerivation, CapWithinBoardBudget, DriftLogVerbWorktreeWriteWaits,
  DefaultStillWaits, HomestateDoesNotImportKanban.
- Still RED (WM5 only): TestFactoryEnsureCardWorktreeConcurrentRealMaterializer, TestFactoryEnsureCardWorktreeStepLockBounded.
- Two WM1 tests were red for the wrong reason and were repaired in `2016bad1a` (no timing bound touched): `SectionAllowedSet` scanned git
  from the verb's start (the record open and store construction run 68 project-root git lines before any section, already in
  `red-cli.txt`) and compared `/var` with `/private/var`; it now measures from the pass-entry seam (the nominated lease calls it
  too) and compares canonical paths. `DriftLogStallBounded` clause (iii) used the wall clock against a lease issued under the
  fixture's frozen clock; it uses `factoryCardNow`.
- Plan section 5 moves done: tolerant race helper (`nmRaceAtSeamTolerant`), item-moved test via operator goroutine, compensate table
  inside a section; the lock-wait subtest was REMOVED (recorded in the test's comment), as plan section 5 states.
- Preservation: AC-FAL-010 68-name selector, `-count=1 -v`, exit 0, `ok internal/cli 158.882s`, PASS 68, FAIL 0, SKIP 0, no-tests 0,
  names identical to the WM1 baseline (`wm3-ac010-final.txt`). DATA RACE 0 there is vacuous (no `-race`); the lease selector under `-race`
  exited 0 with 0 DATA RACE (`wm4-cli-lease-race.txt`). All factory/mcp-factory test files: PASS 257, FAIL 2 (the WM5 two), SKIP 4
  (`wm3-cli-factoryfiles.txt`, taken before the two lint-only test edits). Full `internal/homestate` `-race`: exit 0, 137 PASS (`wm4-homestate-pkg.txt`).
- gofmt clean, `go vet` (kanban, homestate, cli) rc 0, `golangci-lint` v2.1.6 rc 0 (0 issues), `GOOS=windows` build and vet rc 0 (`wm3-static-*.txt`).
- Not claimed: mutants MU1-MU7, MU9-MU20 (WM6); hold-time distribution (AC-FAL-009 iv, WM6); the Windows non-waiting lock beyond compile;
  a lock-release failure after a committed lease is not reported (the lease stays valid); a claim on a loaded machine can read as busy at
  the 800 ms deadline by design.

### WM5 — the worktree-step lock (card t1458, round R4)

Commit `da0c9d039` (code and tests), then the docs commit carrying this section. Tool: `go` from PATH, lane environment scrubbed in each
invocation (scratchpad scripts); the `internal-cli-suite` slot held for each `internal/cli` run and released after (`moai slot status`: free).

- Built: `kanban.AcquireFactoryStepLock` (`internal/kanban/factory_step_lock.go`) over `acquireBoardLockImpl` and `boardLockRetryWait`, own file
  `<root>/.moai/state/factory-worktree-step.lock`; `factoryCreateAndRenameWorktree` in `factory_card.go` takes it around the creator and
  `git branch -m` for `factoryWorktreeStepWait` and releases it (defer) before `RecordCardWorktree`. The constants of D3 already existed from WM1.
  `@MX:WARN` (with REASON) on the helper, `@MX:NOTE` on the step function.
- RED to GREEN (`wm5-cli-ac006.txt`, `-race -count=3 -timeout 25m`, exit 0, `ok internal/cli 166.625s`):
  `TestFactoryEnsureCardWorktreeConcurrentRealMaterializer` (3 of 3 PASS, 49-55 s each, `overlap-iterations=0 failed-iterations=0` per run),
  `TestFactoryEnsureCardWorktreeStepLockBounded` (1.4-1.65 s), `TestFactoryWorktreeStepWaitDerivation`. New `internal/kanban` tests
  (`wm5-kanban-steplock.txt`, `-race -count=3`, exit 0): `TestFactoryStepLockSerializesContenders`, `...ReleasesOnErrorAndPanic`,
  `...WaitIsBounded`, `...IsItsOwnFile`. Not observed RED against the stub for the kanban tests (written beside the implementation); the CLI
  pair was RED at WM1 (`red-cli.txt`).
- Test edit, stated because the round was told not to touch the WM1 tests further: `TestFactoryLeaseSectionAllowedSet` went red
  (`wm5-cli-lease-race.txt`) because its tree snapshot at the creator stub reaches into the worktree step, which creates the step lock's file after
  the section has ended. The test now allows that one path (`factory-worktree-step.lock`) in its outside-the-stores scan; no git-line clause and
  no timing bound changed. Re-run: PASS x2 under `-race` (`wm5-cli-allowedset.txt`). The edit is in `da0c9d039` and can be vetoed by the
  leader; the alternative that keeps the test untouched is a lock file outside the project root, which D3 does not allow.
- Lease/worktree selector under `-race` (`wm5-cli-lease-race.txt`, `^(TestFactoryLease|TestFactoryEnsureCardWorktree|TestFactoryWorktreeStepWait|TestFactoryNextNominate)`): the only FAIL was
  `SectionAllowedSet` above; every other test PASS, including SectionExcludesWorktreeStep and DriftLogVerbWorktreeWriteWaits (returned 135-157 ms
  after the release). `TestHomestateDoesNotImportKanban` PASS (`wm5-homestate-layer.txt`); `go test ./internal/kanban -run Lock -race` exit 0
  (`wm5-kanban-lock.txt`).
- Preservation: AC-FAL-010 68-name selector, `-count=1 -v`, exit 0, `ok internal/cli 164.592s`, `--- PASS` **68**, `--- FAIL` 0, `--- SKIP` 0,
  no-tests 0, names identical to the WM1 baseline (`wm5-ac010-final.txt`; `DATA RACE` 0 is vacuous there, no `-race`). All
  factory/mcp-factory test files (`wm5-cli-factoryfiles.txt`): exit 0, `ok internal/cli 603.587s`, PASS 259, FAIL 0, SKIP 4.
- Static (`wm5-static-*.txt`): gofmt empty; `go vet` (kanban, homestate, cli) rc 0; `golangci-lint` v2.1.6 rc 0, 0 issues; `GOOS=windows GOARCH=amd64`
  build and vet of kanban, homestate, cli rc 0.
- Not claimed: mutants (WM6, MU4 in particular); the Windows lock beyond compile; a Windows stale lock file (D3: no recovery planned); the step
  lock's wait of 60 s was not measured at ten lanes (spec §F R12); a timeout after a successful lease leaves the card `leased` with no recorded
  worktree (D3, spec §F R15).

### WM6 — closure (card t1458, round R5)

No Go source changed in WM6: the repository was never edited by a mutant. Every mutant ran through `go test -overlay` (via `GOFLAGS`, so the
`go list` subprocess of the layering test sees it too) on a mutated copy kept in scratch, and every run was bracketed by a sha256 of the five
tracked source files it could touch (`factory_card.go`, `backlog_store.go`, `card_unavailable.go`, `card_transition.go`, `factory.go`): all runs
printed `source-sha256=IDENTICAL` and the tree carried no tracked change (`git status --short` listed only the new evidence files). Judging build:
`go` from PATH (no project build is involved in a `go test` run), tree HEAD `6c374496d`. Generators and runner are in
`evidence/wm6-mutants-gen.py.txt`, `wm6-mutants-gen2.py.txt`, `wm6-mutant-run.sh.txt`; per-run output is `.moai/reports/t1458/wm6-<MU>-<pkg>*.txt`
(git-ignored), the one-page roll-up `wm6-mutants-summary.txt`. A control run of the same selectors on the unmutated tree is green
(`wm6-CONTROL-*.txt`: 17 cli lease tests, kanban lock tests, four homestate tests, the nominate moved tests, cross-process x10).

**Mutants (plan §7).** RED means the named top-level test printed `--- FAIL` and the package exited 1.

| MU | Selector (scoped, -count=1 unless noted) | Result | Why it turned red (decisive line) |
|---|---|---|---|
| MU1 | AC-001 trio | RED for Distinct nominees and Own-assigned siblings; **BareLanes survived** (also 10 of 10 at -count=10) | `serial cards leased=2`; see finding F1 |
| MU1q | same trio, queue read also taken before the lock (the pin's shape) | RED, all three | `serial exclusivity: 2 serial cards are leased` |
| MU2 | AC-002/-003/-004/-009 selector | RED x3 (OperatorWriteWaitsForSection, CompensationKeepsOperatorPick, SectionRecordWritesPerArm); **ArmCOperatorHold survived** | see finding F2 |
| MU2p | same, the claim-point seam moves with the claim | RED, all four | `completed at claim point = true` |
| MU3 | `TestFactoryLeaseCompensationKeepsOperatorPick` | RED | `the operator's fresh pick was reverted by the compensation: queue=queued want picked` |
| MU4 | `TestFactoryEnsureCardWorktreeConcurrentRealMaterializer` | RED | `overlap-iterations=12 failed-iterations=10`; git `unable to move logfile` |
| MU5 | AC-007 trio | RED for RecordStallBounded and MidClaimStallBounded (3.05 s, limit 1.5 s); QueueLockStallBounded stays green (not its bound) | `returned after 3.049696083s with the record held, want within 1.5s` |
| MU6 | `TestFactoryLeaseQueueLockStallBounded` | RED | `exit code = -1, want 4 for refused raced` |
| MU7 | `TestFactoryLeaseSerialCrossProcessExactlyOne` -count=10, then -count=50 | RED: 2 of 10, then 17 of 50 iterations; goroutine-lane tests stay GREEN (as the plan predicts) | `serial exclusivity across two processes: 2 serial cards are leased`; see finding F4 |
| MU8 | `TestLockedBacklogLoadIsPure` (kanban), the adopting read runs with the lock held, no self-contention | RED | `layout after the section = {dbExists:true jsonExists:false}; the read adopted (migrated) the queue` |
| MU9 | `TestHomestateDoesNotImportKanban` | RED, but at build: `import cycle not allowed` | finding F6 |
| MU10 | `TestFactoryLeaseSectionAllowedSet` | RED, both forms | `a git subprocess ran inside the section ... "git-start --version"` |
| MU11 | `TestFactoryLeaseQueueLockStallBounded` | RED | `the bare lease returned after 5.668254041s, want within 3.8s` |
| MU12 | `TestFactoryLeaseSectionExcludesWorktreeStep` (creator called inside AND outside) | **survived** | finding F3 |
| MU12p | same, the creator and the card-worktree record write moved inside the section | RED | `a Mutate issued from inside the creator stub did not complete within 500ms` |
| MU13 | `TestFactoryLeaseCapWithinBoardBudget` | RED | `3 x factoryLeaseClaimWaitCap = 5.1s > kanban.LockWaitBudget() = 3.3s` |
| MU14 | OperatorWriteWaitsForSection, ArmCOperatorHold | RED, both | `the operator write completed inside the section` |
| MU15 | cli `TestFactoryLeaseDriftLogStallBounded`; homestate `...BoundedSkipsOnContention` | RED, both | cli `returned after 3.504187083s ... want within 1.5s`; homestate `elapsed=1.502s ... record.drift-events=1` |
| MU16 | homestate `TestRecordWriteReconcileDefaultStillWaits` | RED | `the ordinary write returned 1.501619125s before the lock was released` |
| MU17 | homestate `...BoundedSkipsOnContention`; cli `TestFactoryLeaseDriftLogStallBounded` | RED, both | `left unreconciled=1 (want 1) and record.drift-events=1 (want 0)` |
| MU18 | `TestFactoryLeaseArmAKeepsHeldAssignedCard` | RED | `arm (a) no longer leases a held card whose row is assigned to the lane` |
| MU19 | homestate `...BoundedRereadsUnderLock` | RED | `appended 1 record.drift event(s) for an entry another writer had already marked, want 0` |
| MU20 | `TestFactoryLeaseDriftLogVerbWorktreeWriteWaits` | RED | `clause (vii): the verb returned before the timer released the lock` |
| MU21 (extra) | AllowedSet; the section writes `.moai/state/zz-outside-section.txt` | RED | `files created or modified ... outside the two stores: [.moai/state/zz-outside-section.txt]` |
| MU22 (extra) | AllowedSet; the section writes the allowed `.moai/state/factory-worktree-step.lock` | **survived** | finding F5 |
| MU10b (extra) | AllowedSet; a subprocess between lock acquisition and the pass-entry seam | **survived** | finding F5 |

Plan R-A, checked separately (`wm6-MU2-cli-nominate-moved.txt`, `wm6-MU2p-...`): of the four moved nominate tests, only
`TestFactoryNextNominateCompensationFailure` (item-moved) turns red under MU2 and MU2p; `ConcurrentLanes`, `SameCardExactlyOne` and
`CompensateRechecksRecord` stay green. The plan §7 does not map MU2 to them; R-A's mitigation ("MU2 must fail the moved tests") is met by that one test.

**Findings, stated plainly (none was fixed; each is debt for the sync-phase audit to weigh).**

- **F1 (MU1, AC-001 BareLanes).** The mutant as the plan words it (the record snapshot, hence the serial slot, computed before `WithLock`) does
  not turn `TestFactoryLeaseSerialBareLanesExactlyOne` red. Mechanism: arm (b2) reads the fresh queue inside the section, finds the winner's
  `picked` item, and claims it through `RecordPicked`, whose existing "row is no longer picked" guard returns a race and the attempt re-selects.
  The test goes red only when the queue read is also stale (MU1q, the pin's shape). The AC-001 pass condition (all three green x10) is met on
  the real tree; the criterion's coverage of the bare form rests on a defence in depth the mutant cannot isolate.
- **F2 (MU2, AC-003 clause (i)).** With the claim moved after the release but the `factoryLeaseBeforeClaim` seam left inside the section, the
  operator write the test starts at the seam is still blocked at that moment, so `TestFactoryLeaseArmCOperatorHold` stays green. It is red when the
  seam moves with the claim (MU2p). The test observes the seam's position, not the claim's.
- **F3 (MU12).** `TestFactoryLeaseSectionExcludesWorktreeStep` keeps the last creator-stub observation (`inCreator, creatorDone = op, ok` is
  overwritten by each call), so a mutant that calls the creator inside the section and again outside is masked by the later, successful call. A
  mutant that moves the call (MU12p) is caught, which is what the plan's wording says.
- **F4 (MU7 detection rate).** Plan R-I estimated a per-iteration breach rate of 77% (10 of 13 at the pin) and so a survival chance of about
  4 x 10^-7 over ten iterations. Measured under this mutant on this machine: 19 of 60 iterations failed (31.7%; 2 of 10 and 17 of 50, load average
  about 6 to 14), which puts the chance that a ten-iteration run passes a process-local-mutex implementation at about 0.683^10 = 2% if iterations
  are independent (assumed). The criterion's command still failed under MU7 in both runs. The unmutated cross-process test passed 10 of 10.
- **F5 (repairs of 2016bad1a and da0c9d039 versus detection).** Both repairs narrowed `TestFactoryLeaseSectionAllowedSet` in a measured way.
  Still caught: a section writing a path outside the allowed set (MU21, the path named) and a git subprocess after the pass-entry seam (MU10).
  No longer caught: (a) a subprocess started between the lock's acquisition and the pass-entry seam (MU10b survives, the git scan now starts at
  the seam) and (b) a section that writes `.moai/state/factory-worktree-step.lock`, the one path da0c9d039 allows (MU22 survives); in particular
  the step lock taken inside the section, a lock-order inversion, leaves the same file and passes. The wall-clock repair of
  `TestFactoryLeaseDriftLogStallBounded` did not weaken it: it still fails under MU15 (clause (i)) and MU17 (clause (ii)); MU16 and MU19 are the
  homestate tests' mutants and pass the cli test by design.
- **F6 (MU9).** `internal/kanban` already imports `internal/homestate`, so any import of kanban from a homestate non-test file is an import cycle:
  the compiler refuses it before `TestHomestateDoesNotImportKanban` runs (`go test` reported `[setup failed]`). The guard test's own `Errorf`
  branch is therefore unreachable by this mutant; the layering the test names is already enforced by the build.

**AC-FAL-009 (iv) hold-time distribution** (REQ-FAL-011, spec §F R9). Probe: `evidence/probe-hold_test.go.txt` mapped into `internal/cli`
with an instrumented copy of `factory_card.go` (`evidence/probe-hold-instrument.patch.txt`: two clock reads around the section body, nothing else
changed). It drives the real lease function (`factoryNextLeaseOnceGated`, bare form, arm (c): record read, queue read, promotion, three record
writes) on a real queue store and a real factory record in a temp project; the worktree creator is not part of the section. "hold" is the time
`fn` ran with the queue lock held; "lock-wait" is the time from the call to the lock's acquisition. Command (scrubbed environment, slot held):
`GOFLAGS=-overlay=<overlay mapping both files> PROBE_ROUNDS=<n> go test -count=1 -timeout 30m ./internal/cli -run '^TestZZProbeSectionHoldTime$' -v`.
Machine load average before/after: 8.63 / 7.60 (1 round), 7.31 / 6.94 (20 rounds).

| Configuration | n holds | hold p50 | hold p95 | hold max | lock-wait p50 | lock-wait p95 | lock-wait max |
|---|---:|---:|---:|---:|---:|---:|---:|
| 10 sequential leases (one lane) | 10 | 4.08 ms | 4.64 ms | 4.64 ms | 70 us | 80 us | 80 us |
| 2 concurrent lanes (1 round) | 2 | 3.47 ms | 3.87 ms | 3.87 ms | 90 us | 12.57 ms | 12.57 ms |
| 4 concurrent lanes (1 round) | 4 | 3.39 ms | 4.50 ms | 4.50 ms | 8.33 ms | 46.9 ms | 46.9 ms |
| 10 concurrent lanes (1 round) | 10 | 3.62 ms | 8.98 ms | 8.98 ms | 15.09 ms | 80.43 ms | 80.43 ms |

Evidence `wm6-hold-r1.txt` (the criterion's own shape, p95 equals max at these counts). A repeat with 20 fresh rounds per concurrent size
(`wm6-hold-r20.txt`, sequential 10 again): sequential hold p50 4.2 ms, p95 13.3 ms, max 13.3 ms; 2 lanes (n=40) p50 3.87 / p95 4.36 / max 13.23 ms;
4 lanes (n=80) p50 3.63 / p95 5.23 / max 26.03 ms; 10 lanes (n=200) p50 3.44 / p95 6.66 / max 17.5 ms; lock-wait for 10 lanes p50 23.97 / p95 83.26 /
max 114.35 ms. Reading, within what was measured: the section holds the queue lock for single-digit milliseconds typically (largest observed 26 ms)
on an idle drift log and a local disk; the section was not measured with a held drift-log lock or a stalled record (those bounds are
AC-FAL-007 and AC-FAL-015's, at the claim wait cap). No wall-clock threshold is asserted.

**AC-FAL-013 (diff measurement)** (`wm6-ac013.txt`, `git merge-base develop HEAD` re-derived at reading time = `2b9e4a4d067ce4277d7a3ea5df1ec16c7ab231dc`;
HEAD `6c374496d`; local `develop` was `1da5e4fc6`, ahead of the merge-base, which stays at the absorbed tip). Allowed set (control): 43 paths, non-empty.
Forbidden set (probe): 0 paths. Complement of the allowed set (probe): 0 paths. Pathspecs used for "their tests": kanban `backlog_*_test.go`,
`board_lock*_test.go`; homestate `factory*_test.go`, `card_*_test.go`, `admission_lock*_test.go`. Read: the three `card_transition.go` hunks are
at lines 300 to 319 and the hunk headers name `withCardTx` as their function; the `factory.go` hunks are the open path only
(`OpenFactoryBounded`, `openFactoryPathBusy`), and `card_unavailable.go`'s `withCardTx`-reachable change is gated by the context marker (MU16, MU20).
No path outside the enumerated set changed, so no explanation entry is owed. `internal/cli/factory_card_test.go` and `factory_classify_test.go`
are in the allowed set and were not changed.

**Doctrine sweep** (`wm6-doctrine-sweep.txt`): twelve patterns (the old residual risk: both lanes lease, no shared lock, two stores, serial slot snapshot,
non-atomic claim or compensation, accepted residual, creation race) over `.claude/` and `internal/template/templates/`
(`*.md *.tmpl *.yaml *.json *.sh`). Positive control: the pattern `both lease` finds the line in the SPEC-TODO-CLASSIFY-DISPATCH-001 amendment and in
`CHANGELOG.md`. Hits under `.claude/` and the template tree: this card's own planning memory (`.claude/agent-memory/manager-spec/MEMORY.md`, pattern
`atomic lease`), an unrelated accepted-residual note in `rule-authoring.md` (and its template mirror), and an unrelated `git branch -m` line in
`manager-git.md` (and its mirror). No doctrine line states the old residual risk, so no `.claude/rules` or template path is edited (consistent with
the empty forbidden-set probe above). A text sweep is a hypothesis; the claim is bounded to those patterns.

**Final verification on the final tree** (HEAD `6c374496d`, lane environment scrubbed, slot `internal-cli-suite` held only around each `internal/cli`
run and released after each: `moai slot status` free). Summary files `wm6-final-summary.txt`, `wm6-ac-summary.txt`.

- AC-FAL-010, the 68-name selector, `-count=1 -v`: exit 0, `ok internal/cli 148.078s`, `--- PASS` **68**, FAIL 0, SKIP 0, no-tests 0, names identical
  to the WM1 baseline (`wm6-ac010-final.txt`; `DATA RACE` 0 is vacuous there, no `-race`). Floor 68 met.
- All factory/mcp-factory test files: exit 0, `ok internal/cli 503.705s`, PASS 259, FAIL 0, SKIP 4 (`wm6-cli-factoryfiles.txt`).
- Lease/worktree selector `^(TestFactoryLease|TestFactoryEnsureCardWorktree|TestFactoryWorktreeStepWait|TestFactoryNextNominate)` under `-race`: exit 0,
  `ok internal/cli 238.502s`, 36 top-level PASS, 0 FAIL, 0 DATA RACE (`wm6-cli-lease-race.txt`).
- `internal/kanban` whole package `-race`: exit 0, `ok 270.906s`, 598 PASS, 0 FAIL, 0 DATA RACE. `internal/homestate` whole package `-race`: exit 0,
  `ok 52.415s`, 137 PASS, 0 FAIL, 0 DATA RACE.
- Each criterion's own command with its repetition count, judged by S4 (name printed exactly N times, no FAIL/SKIP/no-tests, no DATA RACE for `-race`):
  AC-FAL-001 x10 -race, 002 x20, 003 x20, 004 x20, 005 x20, 009 x20 (all -race), 006 x3 -race, 007 x10 -race, 008 x1, 011 x1 (three packages),
  015 (1) x10 -race over five names in two packages and 015 (2) x1 over six names: every one `exit=0 S4=yes FAIL=0 SKIP=0 notests=0 DATARACE=0`
  (`wm6-ac-<id>.txt`). Slowest: AC-FAL-009, 395.141 s.
- Static (`wm6-static-*.txt`, `wm6-gobuild*.txt`): `gofmt -l` over the changed Go files empty (rc 0); `go vet` kanban, homestate, cli rc 0;
  `golangci-lint run` v2.1.6 over the three packages rc 0, `0 issues.`; `GOOS=windows GOARCH=amd64 go build` and `go vet` of the three packages rc 0;
  `go build ./...` rc 0 and `GOOS=windows GOARCH=amd64 go build ./...` rc 0.
- Not observed in WM6: per-function coverage of the changed functions (quality-gate criterion; not measured here); the Windows lock beyond compile;
  AC-FAL-012 and -014 (sync phase); a plan-auditor or sync-auditor reading of any of this; the section's hold under a held drift-log lock beyond
  the AC-FAL-007/-015 bounds.

## §E.3 Run-phase Audit-Ready Signal

run_status: complete
run_complete_at: 2026-10-04
run_commit_sha: da0c9d039 (last code commit; WM6 changed no Go source — the WM6 docs commit that carries this block follows it, find it with `git log`)
run_commits: 0eb3d5b0d (WM1 seam-and-stub) · 807eabe20 (AC-FAL-010 baseline) · 1a4ef6402 (WM1 RED) · 2997abc29 (WM2 RED) · f8cc4b978 (WM2) · fd70b9123 (WM4 homestate) · 2016bad1a (two WM1 test repairs) · daed5dd09 (WM3 + WM4 cli) · da0c9d039 (WM5)
preservation_baseline: 68 swept (`.moai/reports/t1458/baseline-list.txt`), final 68 PASS 0 FAIL, names identical
ac_pass_count: 13
ac_fail_count: 0
ac_pass_list: AC-FAL-001 to -011, -013, -015 (the S4 pass condition at each criterion's repetition count; AC-FAL-010 and -013 are regression-guards)
ac_sync_phase_pending: AC-FAL-012 (the records state what is not closed), AC-FAL-014 (supersession recorded) — both release-blocking at the sync phase
mutants: plan §7 MU1-MU20 executed, each turned its named test red in at least one faithful form; survivors and weaker-than-planned results F1-F6 above, two extra mutants (MU21 red, MU22 survived) and MU10b survived
hold_time: recorded (section E.2, WM6), single-digit milliseconds typical, largest observed 26 ms
ac_fal_013: control 43 paths, forbidden 0, complement 0
cross_platform_build: GOOS=windows GOARCH=amd64 go build ./... rc 0; go vet of kanban, homestate, cli rc 0
lint: golangci-lint v2.1.6 rc 0, 0 issues; gofmt empty; go vet rc 0
new_warnings_or_lints_introduced: none observed
l44_pre_commit_fetch: not run (nothing is pushed or merged by this card's lane; the branch absorbed develop 2b9e4a4d0 at merge 09faf2965 and stays unpushed)
l44_post_push_fetch: not applicable (no push)
total_run_phase_files: 19 Go files changed under internal/ (`git diff --stat 2b9e4a4d0..HEAD -- internal`: 8 sources, 11 test files)
m1_to_mN_commit_strategy: WM1 as three commits (seam-and-stub, baseline, RED), WM2 RED then GREEN, WM3 and WM4 together with a homestate commit first, WM5 one commit, WM6 docs and evidence only; RED before GREEN is witnessed by the commit graph for WM1, WM2 and WM5's pair (WM5's kanban tests were written beside the implementation, stated in its section)
open_findings_for_sync_audit: F1 to F6 of the WM6 section

## §E.4 Sync-phase Audit-Ready Signal

sync_status: audit-ready
sync_complete_at: 2026-10-04
sync_commit_sha: ae96088b1b7a6674beb85af5b1dd652f8c1bccd5
status_transition: in-progress → implemented → completed, frontmatter `status` + `updated` only, riding the single sync commit (spec.md `status: completed`, `updated: 2026-10-04`; plan.md and acceptance.md carry no `status:` field, `updated:` not present there either, so neither file changed)
sync_commits: the one sync commit (subject `docs(SPEC-FACTORY-ATOMIC-LEASE-001): sync-phase artifacts ... 3-phase close (card t1458)`); its own SHA is backfilled by a following commit per the D3 exemption
files_changed_by_sync: `CHANGELOG.md` (one `[Unreleased]` / Fixed entry, AC-FAL-012), `.moai/specs/SPEC-FACTORY-ATOMIC-LEASE-001/spec.md` (frontmatter `status`, `updated` only), this file (§E.4 and §G)
no_other_doc_surface_changed: README, docs-site (4 locales), `.claude/rules/**`, `internal/template/templates/**`, codemaps — none states the old residual risk or the old lease behavior; grep evidence in `.moai/reports/t1458/sync-doc-sweep.txt`
b12_self_test: pre-emission `grep -c SPEC-FACTORY-ATOMIC-LEASE-001 CHANGELOG.md` was 0 before the entry and is 1 after; live AC count 15 (`live=15 excluded=0 ambiguous=0`, `AC_FILE=acceptance.md`, tier M; `.moai/reports/t1458/sync-b12-ac-count.txt`)
ac_fal_012: CHANGELOG entry written; it names M1 to M5 as closed within the critical section, lists the open windows R1 to R18 by name with R6 (multi-lane stall) spelled out, and its only use of "atomic" carries "within the critical section" in the same sentence; `grep -c 'R6 — multi-lane stall' CHANGELOG.md` is 1 (the spec.md count is the criterion's mutation check, spec.md §F is untouched). The sentence-level qualifier check is a reading (doctrine-only)
ac_fal_014: NOT MET by this commit — OPEN, release-blocking. The three Amendments records (SPEC-TODO-AUTO-PICK-001, SPEC-TODO-CLASSIFY-DISPATCH-001, SPEC-FACTORY-RECORD-001) edit the bodies of three other completed SPECs through `completed → in-progress (amendment)` and need a manager-spec re-delegation (spec.md §B.4, plan D5; the leader may veto the reopening, spec §H DL-4). manager-docs may not edit SPEC bodies and may not spawn; it is returned to the orchestrator as a blocker
open_windows_named_in_the_changelog: R1 to R18 of spec.md §F stay open as written; M1 to M5 are the closed set

### Run-phase findings F1 to F6 — candidate debt, disposition proposed (recorded, none fixed)

Source: §E.2 WM6, mutation pass `wm6-mutants-summary.txt`. The sync audit weighs them; the proposals below are this phase's, not a verdict.

| Id | Finding | Kind | Proposed disposition |
|---|---|---|---|
| F1 | MU1 (record snapshot read before `WithLock`) does not turn `TestFactoryLeaseSerialBareLanesExactlyOne` red; only the variant that also reads the queue stale (MU1q) does. The bare form's exclusivity rests on `RecordPicked`'s "no longer picked" guard as a second defence | test sensitivity (AC-FAL-001 BareLanes) | Record as debt, owned by a follow-up test card: add a bare-lane case that makes the guard defence unavailable so the test isolates the section. Accept for this close — AC-FAL-001's stated pass condition is met and MU1q is red |
| F2 | MU2 (claim after the release, the `factoryLeaseBeforeClaim` seam left inside) leaves `TestFactoryLeaseArmCOperatorHold` green; red only when the seam moves with the claim (MU2p). The test observes the seam position, not the claim position | test sensitivity (AC-FAL-003 clause (i)) | Record as debt: the follow-up card should place the operator write at a seam that is moved by the claim itself. Accept for this close — three other tests are red under MU2 |
| F3 | `TestFactoryLeaseSectionExcludesWorktreeStep` keeps only the last creator-stub observation, so a mutant that calls the creator both inside and outside the section is masked by the later call; a moved call (MU12p) is caught | test sensitivity (REQ-FAL-007/-008) | Record as debt: record every call's in-section flag, not the last. Accept — plan §7 words MU12 as a move, and that form is red |
| F4 | Cross-process serial-slot test detection rate is about 32% per iteration (19 of 60) against the plan's estimated 77%; a ten-iteration run would pass a process-local-mutex implementation about 2% of the time (independence assumed). The criterion's own command failed under MU7 in both runs and the real lock passed 10 of 10 | detection power (AC-FAL-001 / MU7) | Record as debt; propose raising the AC-FAL-001 cross-process repetition count in a later amendment of acceptance.md (a manager-spec edit, not made here). Accept for this close with the number stated |
| F5 | Two repairs narrowed `TestFactoryLeaseSectionAllowedSet`: a subprocess between lock acquisition and the pass-entry seam (MU10b) and a section that writes the allowed step-lock path (MU22, a lock-order inversion) both survive | test sensitivity (REQ-FAL-007, REQ-FAL-010 ordering) | Record as debt, highest of the six because a lock-order inversion is a real hazard: the follow-up should add a check that the step lock is never held while the queue lock is, independent of the file-set scan. Recommend the sync audit weigh this one first |
| F6 | `internal/kanban` already imports `internal/homestate`, so MU9 fails at build with an import cycle before `TestHomestateDoesNotImportKanban` runs; the guard test's `Errorf` branch is unreachable | redundant guard (REQ-FAL-010) | Accept: the layering is enforced by the compiler; no action. Optionally note the test as a documentation guard |

### Sync gate evidence (this run, tree HEAD `c278fa520` plus the uncommitted sync edits; judging build: `moai` built from this tree with `go build -o <scratch>/moai ./cmd/moai`, exit 0, so the build commit equals the tree HEAD; raw files under `.moai/reports/t1458/sync-*`, git-ignored and local to this worktree)

- `moai spec lint SPEC-FACTORY-ATOMIC-LEASE-001` → exit 0, output `✓ No findings — all SPEC documents are valid`. Positive control: the same verb on a copy of spec.md with the `tags:` line removed printed `FrontmatterInvalid` (count 1) and `CoverageIncomplete` warnings, so the rule set fires.
- `moai spec audit --json` → exit 0; the only entry naming this SPEC is `era V3R6, finding_type EraAutoDetected, severity INFO, heuristic H-4 (§E.2 + §E.4 + sync_commit_sha)`; no MUST-FIX or status-drift finding for it.
- `moai spec drift --no-cache --json` → exit 0; this SPEC does not appear in its records (count 0), so the drift detector made no statement about it (not read as a clean result).
- `go test -count=1 -timeout 30m ./internal/spec/...` → exit 0, `ok github.com/modu-ai/moai-adk/internal/spec 112.200s`. No `internal/cli` test was run, so no `internal-cli-suite` slot was taken.
- No Go file, template file or `.claude`/`.moai/config` mirror is touched by the sync commit, so `gofmt`, `go vet` and `make build` have nothing to judge for it; `git status --short` before the commit lists exactly `CHANGELOG.md`, spec.md and this file.
- acceptance.md is unchanged, so no AC baseline snapshot is owed in this commit.

residual_risk: the sync phase observed only what is listed in the sync gate evidence above; the run-phase measurements (68-test baseline, repetition runs, mutants, hold time) were not re-run here. Gaps: AC-FAL-014 open; per-function coverage of the changed functions not measured in run or sync; the Windows lock beyond compile; no plan-auditor or sync-auditor reading of this card yet (the sync audit follows this commit); codemaps (`.moai/project/codemaps/*`) were not regenerated and now lag the three packages' file lists by the new files (`internal/kanban/factory_step_lock.go`, the lease test files) — codemaps refresh is a separate card with its own fold guard.

## §F Phase 4 Mode Selection

Input parameters: tier M; scope about 12 Go files in `internal/cli`, `internal/kanban`, `internal/homestate` plus their tests;
domain count 3 (CLI lease verbs, queue store, factory record); file mix 100% Go; concurrency benefit LOW (coding-heavy, ordered
milestones with a shared file `internal/cli/factory_card.go`); Agent Teams prerequisites not applicable (not requested).

| Mode | Selected | Rationale |
|---|---|---|
| direct | not selected | semantic change across three packages |
| serial | **selected** | coding-heavy, milestones depend on each other (WM2 primitive before WM3, WM3 with WM4) |
| fanout | not selected | write-capable work in one tree; one writer per working tree |
| sweep | not selected | not a uniform mechanical transform |

Decision: serial

Justification: the milestones share one tree and one hot file; a single write-capable agent runs at a time. The orchestrator
verifies each round (own `git status`, HEAD, scoped test runs) before the next spawn. Rounds: R1 = WM1 (seam-and-stub commit,
baseline commit, RED commit); R2 = WM2; R3 = WM3 + WM4 (they land together); R4 = WM5; R5 = WM6 closure. The run starts after
the autonomous plan->run Kickoff recorded in `.moai/reports/t1458/decision.md` (git-ignored) at HEAD b27652922, and after
absorbing develop 2b9e4a4d0 (merge 09faf2965; the plan-artifact hashes were re-checked unchanged).

## §G Resume Point (after the sync commit, 2026-10-04)

- Sync landed (manager-docs): commit `ae96088b1` carried the CHANGELOG entry, spec.md `status: completed` + `updated: 2026-10-04`, and §E.4; its SHA is backfilled in §E.4 by the commit after the amendment landing. AC-FAL-014 amendment records landed: the three Amendments commits `d879b4b04` (SPEC-TODO-AUTO-PICK-001), `47afa50a6` (SPEC-TODO-CLASSIFY-DISPATCH-001), `770e99146` (SPEC-FACTORY-RECORD-001), and the re-close commits `48610afde`, `85caed9ff`, `eaaecb893` (each prior SPEC back to `status: completed`). AC-FAL-014 sync-half greps: AUTO-PICK 7/1/completed, CLASSIFY 6/2/completed, RECORD 5/1/completed. DL-4 (spec §H) was executed at its default (record it) per `.moai/reports/t1458/decision.md` line 47, and remains open to the leader's veto. **Next steps:** independent sync-audit (weighing F1 to F6, §E.4), then develop absorption plus merge-tree re-measurement, then the completion report and the leader's integration window (no push by the lane). The WM6-era resume notes below are kept as written.

### Earlier resume notes (after WM6, 2026-10-04)

- WM1 landed in order: 0eb3d5b0d (seam-and-stub), 807eabe20 (AC-FAL-010 baseline, 68 PASS), 1a4ef6402 (RED). WM2 landed in order: 2997abc29 (RED tests), f8cc4b978 (primitive), then the docs commit carrying this section (find its SHA with `git log`).
- Evidence (git-ignored, this worktree only): .moai/reports/t1458/{baseline-*.txt, red-*.txt, plan-audit*.md, decision.md, park.md}.
- WM3+WM4 landed: fd70b9123 (homestate), 2016bad1a (two WM1 test repairs), daed5dd09 (lease section + bounded claim), then the docs commit carrying this section. Slot internal-cli-suite released. Nothing pushed or merged.
- WM5 landed (round R4): `da0c9d039` (step lock helper, `factoryCreateAndRenameWorktree`, helper tests, the one-path allowance in `TestFactoryLeaseSectionAllowedSet`), then the docs commit carrying this section. No test is RED now: AC-FAL-006's pair is GREEN, AC-FAL-010 68 PASS, all factory test files 259 PASS / 0 FAIL. Slot internal-cli-suite released. Nothing pushed or merged.
- WM6 landed (round R5, closure): no Go source changed. Mutants MU1-MU20 plus MU1q, MU2p, MU10b, MU12p, MU21, MU22 executed through `go test -overlay` with the repository untouched (sha256 identical before and after every run); hold-time distribution, AC-FAL-013, the doctrine sweep and the final verification are in section E.2 (WM6), and the run-phase signal is section E.3 (`run_status: complete`). The docs commit carrying this section is the one after `da0c9d039` (find its SHA with `git log`). Slot internal-cli-suite released (`moai slot status`: free). Nothing pushed or merged. The extra evidence files are `evidence/probe-hold_test.go.txt`, `probe-hold-instrument.patch.txt`, `wm6-*.txt`; the per-run outputs are git-ignored under `.moai/reports/t1458/wm6-*`.
- [Superseded: the sync phase and the AC-FAL-014 amendment records have landed; see the first bullet of this section.] Resume at the sync phase (manager-docs, via `/moai sync SPEC-FACTORY-ATOMIC-LEASE-001`, after the plan-to-sync audit path the card's route requires): (1) write the CHANGELOG entry (AC-FAL-012: which windows this card closed, which stay open per spec §F, R6 multi-lane stall named; `grep -c 'R6 — multi-lane stall'` and `grep -c SPEC-FACTORY-ATOMIC-LEASE-001 CHANGELOG.md`); (2) the three Amendments records through manager-spec re-delegation (AC-FAL-014: SPEC-TODO-CLASSIFY-DISPATCH-001, SPEC-TODO-AUTO-PICK-001, SPEC-FACTORY-RECORD-001, `prior_completed_sha` equal to each prior close's `sync_commit_sha`); (3) the §E.4 sync-phase signal; (4) `moai spec lint SPEC-FACTORY-ATOMIC-LEASE-001`; (5) weigh findings F1 to F6 of section E.2 (WM6) in the sync audit: F1 to F3 are test-sensitivity survivors of variant mutants (AC-FAL-001 BareLanes, AC-FAL-003 clause (i), the creator-stub last-write-wins), F4 is the measured cross-process detection rate (about 32% per iteration against the planned 77%), F5 the two AllowedSet blind spots, F6 the compile-time layering. No fix was applied; each is a debt candidate the card's owner may take in scope or defer. Then the leader's integration window (merge into local `develop` after the merge-tree re-measurement, report the merge SHA; no push by the lane).
