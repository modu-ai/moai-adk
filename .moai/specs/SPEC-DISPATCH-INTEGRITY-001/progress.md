# progress.md — SPEC-DISPATCH-INTEGRITY-001 (card t1595)

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-09T18:56Z
tier: M
artifact_set: spec.md, plan.md, acceptance.md, progress.md, decision-index.md
baseline_tree: 81786284e
baseline_classification: measured 2026-10-08 (plan.md §B); four-element ledger cells EL-001..004 captured 2026-10-09, input-bound to owned-tests/ mirror (round-2); M0 re-affirms on the current tree before any fix
open_operator_decisions: 0 blocking (2 default-applied — decision-index.md)

## §E.2 Run-phase Evidence

Run-phase owner: manager-develop (card t1595). Tree: worktree
`.moai/worktrees/t1595`, branch `WT-dispatch-integrity`. Baseline HEAD at
run entry: `5a91b5758` (re-read before every commit). All measurements are
env-scrubbed compound invocations
(`unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test …`),
worktree-guard refusals named where they occurred, full-suite runs excluded
(C2).

### M0 — Overlay intake & baseline re-verification (complete)

Pre-flight: `git branch --show-current` → `WT-dispatch-integrity`;
`git rev-parse --short HEAD` → `5a91b5758`; `go build ./...` → exit 0;
`GOOS=windows GOARCH=amd64 go build ./...` → exit 0; `golangci-lint run
--timeout=2m` → `0 issues.` (baseline, no NEW findings after the M0 test
files either — re-run below); Go-bytes identity vs the classification tree
verified: `git diff 81786284e..HEAD --stat -- internal/ cmd/` → empty.

Canonical drop-in measured (superset + worktree file + hook file from
`.moai/reports/t1595/overlay/`), env-scrubbed `-run 'TestReviewFinding'
-count=1 -v` per package, then trimmed to the owned subset before commit:

| Test | Pkg | §B @ 81786284e | This run @ 5a91b5758 | Divergence |
|---|---|---|---|---|
| MergedPRPredecessor (defect 5) | cli | PASS | PASS | none |
| NominatedOverwritesDependency (defect 4) | cli | PASS | PASS | none |
| FoldConcurrentWrite (defect 8a) | cli | FAIL | FAIL | none |
| FoldInterleavedArchiveLoss (defect 8b) | cli | FAIL | FAIL | none |
| LaneProbe (deferred → t1596) | cli | FAIL 3/3 | FAIL 3/3 | none |
| ExplicitCandidateFiles (t1596) | cli | FAIL | FAIL | none |
| DryRunFactoryMigration (t1596) | cli | FAIL | FAIL | none |
| WhitespaceLanding (t1561) | cli/worktree | PASS | PASS | none |
| ZoneExistingDotDot (control) | hook | PASS | PASS | none |
| BackgroundReceiptRecycling (t1562) | hook | FAIL | FAIL | none |

Classification re-affirmed on THIS tree (AC-DI-001): (4)(5) NOT REPRODUCED
→ regression guards (M2); (8a)(8b) RED-LIVE → fix in M4; (1)(2)(3)(6)(7)
UNDETERMINED (no overlay test) → RED-first characterization in M1–M3.
Zero divergences; no re-classification.

M0's authored instruments (committed in the M0 commit as
`internal/cli/review_observation_test.go` +
`internal/hook/review_observation_test.go`; foreign tests trimmed per
decision-index Q2):

- Re-authored `TestReviewFindingFoldInterleavedArchiveLoss` — AC-DI-010's
  serialized shape: fold B is a separate OS process
  (re-exec of the test binary through `TestFoldSubprocessHelper`, env-gated,
  skipped in ordinary runs) started at fold A's archive-write probe; A's
  critical section is held open until B's whole fold finishes or a
  60s-bounded wait expires. RED-now measured (EL-009): B completed inside
  A's window, B's line landed in NEITHER index, A's line duplicated — M4's
  lock flips it green. Selector anchored; the exit-code capture of the
  guard-refused compound shape (`${PIPESTATUS[…]}`) was split to a plain
  invocation + separate exit read — deviation named per §C.
- Arch-coverage debt instruments (iter-7 debt, dispose_in=run):
  `TestReviewFindingFoldArchiveConcurrentWrite` (the guard=nil archive
  write's post-probe window; RED-now EL-010 — the rename published over a
  concurrent author's bytes) and `TestReviewFindingFoldGuardArchiveChange`
  (the guard file's post-probe recheck; RED-now EL-011 — the codex gate's
  data-loss mutant reproduced through the fold verb, line in neither
  index). M4's post-probe final comparisons flip both green.
- EL-005 replaced with the measured four elements (control PASS, exit 0,
  tree 5a91b5758, input e7c0e4791 mirror + the M0-committed drop-in).
- M0 intake observations recorded: strengthened nomination body GREEN
  (EL-006), positive control's FIRST compile+run GREEN-at-adoption
  (EL-007), strengthened fold body RED (EL-008).

Family state after M0 (confirmation only — the per-fix re-measures are
AC-DI-012's, at each milestone): the owned set runs
(`^TestReviewFindingMergedPRPredecessor$|…|^TestFoldSubprocessHelper$` →
3 PASS + 1 SKIP + FoldConcurrentWrite FAIL-as-expected, exit 1 from the
expected RED); existing families untouched by M0 (no production bytes
changed in M0).

### M1 — Bundle predecessor semantics, defects (1)(2)(3) (complete)

Characterizations authored RED-first into
`internal/cli/review_observation_test.go` from the existing helper surface
(`fcFixture`/`fcQueue`/`fcClassify`/`fbSeedFiles`/`fcPlace`/`runFactory`/
`fbLeasedCard`), each measured on THIS tree (HEAD `345eb6483`, production
bytes of the M0 commit) before any fix:

- **(3) `TestReviewFindingBundleDuplicateMemberRefused` — LIVE, RED**
  (EL-012): the duplicate-id load failed only via an incidental version
  conflict (`stale card version: card t1 is at version 2, request expected
  1`), not a duplicate-member refusal; exit 1. Nothing recorded (the
  rollback held), but the refusal named no duplicate — the AC demands the
  named refusal before any record.
- **(2) `TestReviewFindingBundleHeadHubConstraint` — NOT REPRODUCED**
  (EL-013, regression guard per C1): the load REFUSED with
  `predecessor card not merged: t2 has not reached merged-local …` — the
  head-hint T2 guard (t1533 r2a) refuses the load outright; AC-DI-004
  admits refusal as the constraint. exit 0.
- **(1) `TestReviewFindingBundleMultiHubMemberWaits` — NOT REPRODUCED**
  (EL-014, regression guard per C1): with the member RECORDED (loaded while
  the sharers were rowless) and the sharers' open rows placed afterwards,
  all three partially-blocked states refuse the lease (`""`) and both-merged
  leases (`"t3"`) — the selection-level multi-hub sweep
  (`factoryHubWaitUnmerged`, t1533 r10) satisfies REQ-DISPATCH-001's
  behavioral requirement in all four cells. exit 0.

Fix (b) — the LIVE one only: `runFactoryBundleLocked` now refuses a
duplicate member id BY NAME before any read or record
(`duplicate member %s in the bundle member list`); GREEN re-measured
(EL-015, exit 0). Re-implementation of (1)/(2) is prohibited (§G); their
tests stand as the committed regression guards.

AC-DI-012 re-measurement — family list recorded BEFORE the run: the touched
function `runFactoryBundleLocked`'s callers are `runFactoryBundle` (same
file); selection code untouched, so the entered family is
`factory_bundle_test.go` — 13 tests, selector
`^(TestFactoryNextBundleSerialLane|TestFactoryBundleKeepsSerialSlot|
TestFactoryAssignBundleOrderGuard|TestFactoryAssignBundleHubChain|
TestFactorySerialBundleHeadLeasesDespitePickedMembers|
TestFactoryBundleLoadAtomicWhenAMemberIsLeased|
TestFactoryNextSkipsHubCandidateWhosePredecessorIsUnmerged|
TestFactoryNextAfterGuardCountsOtherRunsMerges|
TestFactoryHubChainRequiresSharedHubPath|
TestFactoryHubChainChainsToTheLastPredecessor|
TestFactoryNominateRefusesForeignBundleMember|
TestFactoryBundleRecordsUnderTheQueueLock|
TestFactoryKeepSetReadsNoFileOverlap)$` → exit 0, `ok … 36.175s`, all 13
green. Slot lease taken for the family run
(`moai slot acquire --resource internal-cli-suite --max-duration 15m`;
displaced one expired holder from 2026-10-06).

### M2 — Card selection & after-overwrite, defects (4)(5)(6) (complete)

No production fix in M2 — all three defects classified NOT REPRODUCED and
their tests stand as the committed regression guards (§G prohibits
re-implementation):

- **(4) `TestReviewFindingNominatedOverwritesDependency`** — the M0
  strengthened two-arm body (EL-006) re-affirmed on the M2 tree HEAD
  `96f392d06`: the refusal names the unmerged predecessor, t3 stays picked
  with the stored hint, and the positive control
  (TestReviewFindingNominatedLeasesAfterPredecessorMerges, EL-007) proves
  the merged-predecessor path actually leases. Combined selector run exit 0
  (both PASS, `ok … 5.630s`).
- **(5) `TestReviewFindingMergedPRPredecessor`** — re-affirmed in the same
  run (`predecessor=merged-pr; factory next leased="t2"`).
- **(6) `TestReviewFindingNoRecordArmSkipsBlockedCandidate`** — authored
  RED-first, measured on `96f392d06` (EL-016): the no-record arm (b2) skips
  the blocked candidate (no record row created — `fcHasCard` false — no
  claim) and progresses the ready card behind it (`lease "t3"`), the whole
  pass unerrored. exit 0. The b2 pre-filter (t1533 r6/r7) satisfies
  REQ-DISPATCH-006.

M2 family re-measure (plan M2; no production change — confirmation run
over the selection/nomination families the regression guards exercise):
selector `^TestFactoryCard|^TestFactoryNextNominate|^TestFactoryNominated|
^TestFactoryNext` → exit 0, `ok github.com/modu-ai/moai-adk/internal/cli
276.772s`. Deviation, named: the `^TestFactoryNext` prefix over-selected —
the run swept the whole factory-next family (hundreds of seconds), far
beyond factory_card_test.go's 10 tests and the nomination family; every
swept test passed. Slot lease: the M1 lease on `internal-cli-suite` still
held (15m window).

### M3 — Merging-retry lease validation, defect (7) (complete)

No production fix in M3 — the defect classified NOT REPRODUCED and its test
stands as the committed regression guard (§G):

- **(7) `TestReviewFindingMergingRetryValidatesLeaseBeforeRemote`** —
  authored RED-first from the `factory_card_pr_test.go` delivery fixtures
  (ghf local-remote + gh double), measured on HEAD `271d71ab9` (EL-018):
  the merging-state retry refuses in BOTH ineligible shapes — the
  foreign-lane holder (`refused — card t1 belongs to lane-1 (lease lane-1),
  not lane-2 …` — the entry owner/holder check) and the expired caller
  lease (`refused — card t1's merging lease held by lane-1 expired at
  2026-09-01T00:00:00Z …`) — and the fixture remote observes ZERO mutation
  in both: no push (`remoteBranchTip` empty), no `gh pr create`, no
  `gh pr merge` (double counts 0/0), the card still at merging. exit 0.
  The entry checks + `verifyLease` (t1533 r2c/r5/r13) satisfy
  REQ-DISPATCH-007.

AC-DI-012 re-measurement — family list recorded BEFORE the run: no
production function touched (classification-only milestone); the plan's
family sweep ran as confirmation: `factory_card_pr_test.go` (15 tests) +
`factory_card_pr_guard_test.go` (1), anchored selector over all 16 names →
exit 0, `ok … 79.079s`, all green. Slot lease: the M1
`internal-cli-suite` lease still held (15m window).

### M4 — Memory-fold cross-process serialization, defect (8) (fix landed)

Fix (M4, plan §F; the run-phase lock-mechanism design the SPEC deferred):

- **Cross-process store lock** (C5): `withFoldStoreLock(dir, fn)` spans the
  fold's WHOLE transaction — snapshot read, plan, and both index writes —
  in BOTH fold paths (the verb's RunE and `foldOnDoneStep`). The lock is
  flock(2) on `<store>/.moai-fold-lock` (unix,
  `memory_fold_lock_unix.go`) / LockFileEx (windows,
  `memory_fold_lock_windows.go`, x/sys) — exclusive, blocking, kernel-
  released on process exit (a crashed fold leaves no stale lock). The lock
  file is dot-prefixed non-.md: invisible to SnapshotStore and the index
  guard. A second fold's process waits at the lock and re-plans against
  the post-transaction store. The bounded on-done step keeps its own
  discipline: the lock wait happens inside the acquire and the abandoned
  step's forbidden() checks still refuse the write after it.
- **Write geometry** (D9, plan §G): `atomicWriteFoldFile` now runs the
  FINAL byte comparison AFTER the pinned `orderProbe("bytes-done")` seam
  and before the rename — re-judging BOTH the file being written and the
  guard file. Geometry: cmp1 → seam probe → NEW final byte comparison →
  rename; "final byte comparison" is the role the post-probe comparison
  takes. This is the defense against a non-cooperating writer; the
  irreducible TOCTOU tail between it and the rename is AC-DI-009's stated
  residual risk — the lock closes the window for cooperating writers only.

GREEN verification (env-scrubbed, anchored, HEAD `c0a0d7cbd` + this fix
uncommitted — the fix and cells land together in the M4 commit):

- `TestReviewFindingFoldConcurrentWrite` → PASS (refused at the new final
  comparison, the concurrent author's bytes preserved).
- `TestReviewFindingFoldArchiveConcurrentWrite` → PASS (the archive write's
  final comparison detected the author's bytes — the arch-coverage debt's
  archive-write path closed).
- `TestReviewFindingFoldGuardArchiveChange` → PASS (the guard file's
  post-probe recheck detected the stripped line — the codex gate's
  data-loss mutant closed).
- `TestReviewFindingFoldInterleavedArchiveLoss` → PASS (60.02s: fold B's
  process made NO progress inside A's window — the bounded wait expired
  with in-window=false — then completed normally against the post-A store
  after A released; both lines exactly once; B's exit observed after A's
  transaction closed).

AC-DI-011 selector, recorded BEFORE the race run: the memory-fold family
sweep is

`-run '^(TestMemoryFold_|TestMemoryFoldOnDone_|TestReviewArchiveUpdateDuringEffectiveScan|TestReviewSequentialAbandonedTempOwnership|TestReviewFindingFold|TestFoldSubprocessHelper)$'
-count=5 -race`

covering `memory_fold_test.go`'s 14 tests (`TestMemoryFold_*` ×13 +
`TestReviewArchiveUpdateDuringEffectiveScan`), `memory_fold_wiring_test.go`'s
15 (`TestMemoryFoldOnDone_*` ×14 + `TestReviewSequentialAbandonedTempOwnership`),
and the owned fold instruments (`TestReviewFindingFold*` ×5 +
`TestFoldSubprocessHelper`, which skips outside the interleaved test's
child). Outcome recorded below the run.

Repair round (in-gate, first race run): the first `-count=5 -race` gate
run FAILED — zero data races, but `TestReviewArchiveUpdateDuringEffectiveScan`
failed all 5 iterations: the lock file's original name `.moai-fold-lock`
collided with the temp-file sweep's `.moai-fold` prefix
(`requireNoTempFiles` — the namespace reserved for the fold's leak-check of
actual temporaries). Repair: the lock renamed to `.moai-store-lock`, the
prefix stays reserved, and the sweep keeps reading a leftover lock as a
foreign file, not as a leaked temp.

Selector-defect record (found because the repair run finished in 3.018s —
impossibly fast for a family whose interleaved member alone bounds at 60s):
the FIRST two gate selectors were `$`-anchored alternations whose
alternatives are PREFIXES (`^…(TestMemoryFold_|…)$`), so the pattern
matched only whole-name equals — the two runs swept just 3 tests
(`TestReviewArchiveUpdateDuringEffectiveScan`,
`TestReviewSequentialAbandonedTempOwnership`, the skipping
`TestFoldSubprocessHelper`) × 5 and never reached the 60s interleaved
member. Both runs are VOID as family measurements — recorded here as the
failed measurements they are, per the selector discipline (a selector
sweeping less than the family is a failed measurement, not a pass). The
corrected selector drops the `$`: front-anchored prefix alternation
`^(TestMemoryFold_|TestMemoryFoldOnDone_|TestReviewFindingFold|TestFoldSubprocessHelper|TestReviewArchiveUpdateDuringEffectiveScan|TestReviewSequentialAbandonedTempOwnership)`
— and the corrected gate runs with `-v`, so the output file itself carries
the swept `=== RUN` set as the measurement's own swept-count evidence.

Gate-finding repairs (the turn-end codex gate reviewed the in-progress
lock and found 2 defects, folded into M4 as its own deliverable; both
RED-first measured, both fixed, both flipped green):

- **Finding 1 — the lock path followed the process TMPDIR.** The temp-dir
  lock keyed by the store path gave two processes with different TMPDIR
  env different lock files: serialization silently broke
  (`TestReviewFindingStoreLockIndependentOfTempDir` RED: "a
  different-TMPDIR locker entered the store's critical section while it
  was held"). Fix: the lock file is `.moai-store-lock` BESIDE the store's
  index — derived from the store alone. Store-shape helpers
  (`storeHashes`, `requireNoTempFiles`) exempt the lock by name as durable
  named infrastructure.
- **Finding 2 — an abandoned bounded fold pinned the store lock.** The
  on-done worker blocked on the store read kept holding the lock after its
  caller's timeout (`TestReviewFindingAbandonedFoldReleasesStoreLock`,
  unix-only FIFO fixture, RED: "the store lock stayed held after the
  bounded fold was abandoned"). Fix: the locked section's snapshot read is
  abandonment-polled (`snapshotStoreBounded`, 100ms poll) — the abandoned
  step releases the lock and returns; the blocked read's goroutine holds
  nothing and its result is discarded.
- Both GREEN re-measured (EL-023/EL-024 below). The gate's external probe
  scenarios (TestReviewProbeDifferentTempLock,
  TestReviewProbeAbandonedStoreLock) are covered on-tree by these two
  committed tests.

AC-DI-011 selector, FINAL (recorded BEFORE the closing race run — extends
the corrected selector with the two finding tests):
`^(TestMemoryFold_|TestMemoryFoldOnDone_|TestReviewFindingFold|TestReviewFindingStoreLock|TestReviewFindingAbandonedFold|TestFoldSubprocessHelper|TestReviewArchiveUpdateDuringEffectiveScan|TestReviewSequentialAbandonedTempOwnership)`
`-count=5 -race -v`. Outcome below.

**AC-DI-011 gate outcome (closing run)**: PASS — `ok
github.com/modu-ai/moai-adk/internal/cli 461.530s`, exit 0, all 5
iterations green, **0 data races** reported, 0 test failures, swept set
180 `=== RUN` lines in the run's own output (the fold family's 29 tests ×
5 + subtests + the two finding tests × 5). The full-suite verdict remains
CI's job (C2) and is PENDING at report time.

### Post-report repair pass (turn-end gate round on 2f000f923 — 2 + 1 findings, all folded)

The turn-end codex gate re-judged the final tree and returned findings in
the abandonment family; each was RED-first tested, minimally fixed, and
flipped green:

- **Finding 1 — only the initial snapshot was cancellation-guarded.** The
  reads INSIDE the lock hold (applyFold's archive re-read, the
  checkFoldUnchanged comparisons, verifyArchiveEffectiveState's store
  snapshot) could still block and pin the lock past the caller's timeout.
  Fix: `readFileBounded` (polling read, nil-forbidden = plain) threaded
  through `checkFoldUnchanged`, `verifyArchiveEffectiveState`, and the
  apply re-read — every in-lock read on the bounded path is
  abandonment-polled; the verb path passes nil and reads plainly.
  RED `TestReviewFindingAbandonedFoldReleasesLockOnApplyReads` (unix FIFO
  swapped in at the archive write's seam): "the store lock stayed held
  after the bounded fold was abandoned — a blocking read inside the apply
  pins the lock past its caller's timeout" → GREEN 0.56s.
- **Finding 2 — the read-only preview required write access.** The locked
  verb opened the lock file (O_CREATE|O_RDWR) before the preview decision:
  a readable-but-not-writable store failed the preview with
  `permission denied` where the baseline previewed. Fix: a preview without
  `--yes` computes the plan from the snapshot WITHOUT acquiring the write
  lock or creating the lock file; the apply path re-computes the plan
  INSIDE the lock (the applied plan is always computed post-wait).
  RED `TestReviewFindingPreviewNeedsNoWriteAccess` (0555 store):
  "open …/.moai-store-lock: permission denied" → GREEN (the plan prints).
- **Finding 3 (same round, third member of the abandonment trio) — the
  blocking lock-ACQUISITION wait ignored the deadline.** A bounded fold
  whose worker waited for a long-held lock stayed stuck inside blocking
  flock/LockFileEx after its caller's timeout; over an auto-done batch
  close abandoned workers and descriptors accumulated. Fix: a non-nil
  forbidden makes `acquireFoldStoreLock` poll a non-blocking try
  (LOCK_NB / LOCKFILE_FAIL_IMMEDIATELY) and abandon — releasing the
  descriptor, holding nothing; nil keeps the plain blocking wait (verb,
  C5's cross-process symmetry on both GOOS implementations).
  RED `TestReviewFindingAbandonedFoldExitsLockWait` (long-holder +
  timed-out acquirer; the compile refusal against the one-arg signature is
  the pre-fix API state the fix introduces) → GREEN 0.53s (the worker
  exits with the abandonment error, no external release).

Re-measure judgment (recorded per the gate's instruction): the lock
mechanism itself (acquire/release primitives' happy paths) is unchanged
for the blocking form, but the bounded path now spawns polling goroutines
inside the lock hold and the acquisition wait changed shape — the FULL
`-count=5 -race` gate is re-run (new goroutines inside the hold are
exactly race-detector territory), with an explicit `-timeout=30m` after
the interim gate run was killed by the 10m default mid-iteration
("panic: test timed out after 10m0s" while the 60s interleaved member was
at 25s of an iteration — a budget problem, recorded as the failed
measurement it is; single-pass family runs before it were green:
103.8s, 100.0s).

**Post-report AC-DI-011 gate outcome (final run)**: PASS — `ok
github.com/modu-ai/moai-adk/internal/cli 608.171s`, exit 0, all 5
iterations green, **0 data races**, 0 test failures, swept set 265
`=== RUN` lines (the fold family, both post-report finding tests, and the
abandonment trio × 5). The full-suite verdict remains CI's job (C2) and is
PENDING at report time.

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-10-09T22:30+09:00
run_commit_sha: 281d490bc
run_status: complete
ac_pass_count: 13
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not-run (lane-local worktree; no pre-spawn fetch duty on a depth-1 worker — the lane's landing sequence owns integration reads)
l44_post_push_fetch: not-run (this worker never pushes; the lane lands the branch)
new_warnings_or_lints_introduced: 0 (golangci-lint 0 issues at every measurement; gofmt clean; go vet clean)
cross_platform_build.native: PASS (go build ./..., exit 0)
cross_platform_build.windows: PASS (GOOS=windows GOARCH=amd64 go build ./..., exit 0)
total_run_phase_files: 15 (5 production: memory_fold.go, factory_bundle.go, memory_fold_lock_unix.go, memory_fold_lock_windows.go, + the verb preview restructure in memory_fold.go counted once; 6 test: review_observation_test.go, review_observation_fifo_unix_test.go, memory_fold_test.go, memory_budget_test.go, internal/hook/review_observation_test.go; 3 SPEC artifacts: acceptance.md ledger EL-001..EL-027, progress.md §E.2/§E.3, spec.md frontmatter transition)
m1_to_mN_commit_strategy: per-milestone commits (M0 345eb6483, M1 96f392d06, M2 271d71ab9, M3 c0a0d7cbd, M4 c97d50a1e) + the post-report repair pass 281d490bc + this audit-ready stamp
notes: 13/13 ACs PASS (5 as committed regression guards with recorded not-reproduced observations per C1; 8 flipped or held green with measured cells EL-001..EL-027). Two in-gate repair rounds, one selector defect, one default-timeout gate kill, and the post-report gate round (2 findings + the abandonment trio's third member) are recorded in §E.2 as failed measurements and repair records, never as passes. The mid-flight gate's 2 findings and the post-report gate's 3 findings were all folded with RED-first tests. The full-suite verdict is CI's job (C2) and PENDING at report time.

## §E.4 Sync-phase Audit-Ready Signal

_Pending sync-phase (manager-docs)._

## §F Phase 4 Mode Selection

Input parameters: tier=M; scope≈4 production files + 5 owned test sources + SPEC artifacts; domains=1 (Go `internal/cli` + `internal/hook` control mirror); language mix=Go + Markdown artifacts; concurrency benefit=LOW (coding-heavy sequential defect repair, single worktree, M0 gates every later classification); agent-teams prereqs=not requested (never auto-selected).

| Mode | Selected | Rationale |
|---|---|---|
| direct | no | multi-milestone semantic fixes exceed direct execution |
| serial | **YES** | coding-heavy repair — sequential manager-develop delegation per milestone (Anthropic coding-task parallelism caveat) |
| fanout | no | not multi-domain research; coding-heavy work stays serial |
| sweep | no | not a uniform mechanical transform; scope is semantic |
| agent-team | no | explicit-request-only; not requested |

Decision: serial

Justification: the card is coding-heavy defect repair across four production files with a strict milestone order — M0's re-classification gates M1–M4 — on a single worktree under one-writer discipline. The serial envelope (one manager-develop delegation at a time, milestones in plan order) is the default fallback and no other mode's selection criteria are met.
