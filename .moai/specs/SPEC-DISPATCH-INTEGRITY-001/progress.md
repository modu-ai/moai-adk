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

### Post-close repair row (turn-end gate round on the completed SPEC — abandonment family, fourth member)

The SPEC reached `completed` (sync commit `d6caefb81`); this row records a
post-close follow-up fix landed on the same branch (same pre-PR shape as
the D3 backfill; §E.3/§E.4 untouched — manager-docs surfaces).

- **Finding — an abandoned fold still published.** The bounded reads of
  the final byte comparison can complete INSIDE a poll interval and return
  normally after the caller's deadline has passed; control then flowed
  straight to the rename and the abandoned fold's rename published
  (gate repro: `abandoned=true err=<nil> final="fold output\n"`). Fix: the
  abandonment flag is re-checked at the last observable moment before the
  rename in `atomicWriteFoldFile`.
  RED `TestReviewFindingAbandonedFoldDoesNotPublish` (unix FIFO; the flip
  lands inside the poll gap and the plan-time bytes are delivered before
  the next tick): verbatim RED "the abandoned fold published: err=<nil>
  final=\"fold output\\n\"" (exit 1) → GREEN ×3 (0.20s each; the refusal
  names the abandonment).

Re-measure judgment (recorded per the standing instruction): the fix adds
a boolean flag check inside the in-lock write path — no new goroutine, no
lock-internal or acquisition change — so the proportionate sweep is the
fold family at `-count=2 -race` (not the full `-count=5` gate); outcome
below.

**Post-close sweep outcome**: PASS — single-pass family green first
(`ok … 120.998s`), then `-count=2 -race -v`: `ok
github.com/modu-ai/moai-adk/internal/cli 244.765s`, exit 0, **0 data
races**, 0 failures, 108 `=== RUN` lines (the fold family + the new
publish-guard test × 2). Builds native + windows green; gofmt/vet/lint
clean. The full-suite verdict remains CI's job (C2) and is PENDING.

### Post-close repair row 2 (gate finding 4 — the lock file must not follow symlinks)

The same turn-end gate round, queued mid-repair and folded as the second
post-close row:

- **Finding — the lock open followed symlinks.** `os.OpenFile` with
  O_CREATE on `.moai-store-lock` followed a symlink placed at the path and
  created/locked the TARGET outside the store (gate repro: the symlink
  pointed at another repository's `.git/index.lock`; the fold succeeded
  and the target repo's `git add` then failed with `File exists`, exit
  128). Fix (unix): the open carries `syscall.O_NOFOLLOW` and the opened
  file is fstat-verified REGULAR — a symlinked lock is refused (ELOOP),
  and every other non-regular shape (a FIFO standing in for the lock, a
  device node) refuses the same way.
  RED `TestReviewFindingLockFileRefusesSymlink`: verbatim RED "the lock
  acquisition followed a symlinked lock file and locked the target outside
  the store" (exit 1) → GREEN ×2 ("symlinked lock refused: … too many
  levels of symbolic links"; the target never created; a regular lock
  file still acquires). unix-only test (the O_NOFOLLOW surface).
- **Windows path statement** (named, not silent): `os.OpenFile` on Windows
  passes no `OPEN_REPARSE_POINT` flag, so the open WOULD follow a symlink
  the same way; the Windows half guards the acquisition with a preceding
  `os.Lstat` symlink refusal. A narrow TOCTOU tail remains between the
  lstat and the open on that platform (the unix half is O_NOFOLLOW-exact);
  the Windows test surface is compile-verified only — the race gates run
  on darwin.

Re-measure judgment: the change is INSIDE the lock acquisition's open path
— lock internals — so the fold family sweep re-ran at `-count=2 -race`;
outcome below.

**Finding-4 sweep outcome**: PASS — single-pass family green first
(`ok … 110.652s`), then `-count=2 -race -v`: `ok
github.com/modu-ai/moai-adk/internal/cli 280.254s`, exit 0, **0 data
races**, 0 failures, 110 `=== RUN` lines (the fold family + the publish
guard + the symlink refusal × 2). Builds native + windows green;
gofmt/vet/lint clean (0 issues). The full-suite verdict remains CI's job
(C2) and is PENDING.

### Post-close repair row 3 (gate finding — the --yes no-op path required the lock to discover nothing to do)

The preview fix's mirror on the apply path:

- **Finding — a no-op fold on a read-only store failed at lock creation.**
  With no lock file on a readable-but-not-writable store, `--yes` failed
  with `permission denied` at `.moai-store-lock` creation BEFORE checking
  whether the target card's line even exists; the baseline answered the
  same repro with the no-fold line and exit 0.
  RED `TestReviewFindingNoOpFoldSucceedsOnReadOnlyStore`: verbatim RED
  "open …/.moai-store-lock: permission denied" (exit 1) → GREEN (the
  no-fold line prints; exit 0).
- **Fix — optimistic-outside / authoritative-inside double-read.** The
  no-change discovery is read-only: the plan is computed from the snapshot
  OUTSIDE the lock and a fold with nothing to do answers there (preview
  and `--yes` alike). Only when actual application is needed does the
  write transaction acquire the lock and RE-COMPUTE the plan inside it —
  the plan the apply executes is always computed from the post-wait
  store; the optimistic snapshot is discarded. The double-read under the
  lock is the authoritative plan computation every apply now performs
  (one extra SnapshotStore per changing fold, the cost of the unlocked
  discovery).

Re-measure judgment: the lock internals are unchanged (acquisition,
release, hold span identical); the verb's flow around them changed —
every `--yes` invocation now double-reads — so the family sweep re-ran at
`-count=2 -race`; outcome below.

**Row-3 sweep outcome**: PASS — single-pass family green first
(`ok … 111.095s`), then `-count=2 -race -v`: `ok
github.com/modu-ai/moai-adk/internal/cli 246.668s`, exit 0, **0 data
races**, 0 failures, 112 `=== RUN` lines (the fold family + the no-op and
preview tests × 2). Builds native + windows green; gofmt/vet/lint clean
(0 issues). The full-suite verdict remains CI's job (C2) and is PENDING.

### Post-close repair row 4 (gate findings on lock scope + on-done parity — with the self-sweep the round's main ask)

Two findings, each RED-first, plus the self-sweep over every entry path
and every in-hold region:

- **Finding 1 — the applied result's stdout write ran INSIDE the lock
  hold** (the verb's apply closure): a blocked stdout pipe kept every
  other fold on the store waiting even after the file update completed.
  Fix: the closure CAPTURES the result render (`var render func() error`)
  and runs it AFTER `withFoldStoreLock` returns — the hold spans exactly
  the index-write critical section; the no-fold answer's render is
  deferred the same way.
  RED `TestReviewFindingResultPrintedOutsideLockHold` (a stdout writer
  that passes the store line and blocks every later write; the apply's
  rename is observed in the store, then a follow-up locker must acquire):
  verbatim RED "the store lock stayed held while the fold was blocked
  printing its result — the stdout write happened inside the lock hold"
  (probe timeout, exit 1) → GREEN ×2.
- **Finding 2 — the on-done auto-fold required the lock to discover
  nothing to do** (the no-op parity of row 3's --yes fix): the step
  created `.moai-store-lock` before checking whether the target card's
  line exists — a read-only store failed with `permission denied` where
  the baseline quietly succeeded, and a no-op waited out another fold's
  hold needlessly. Fix: bounded snapshot + plan OUTSIDE the lock first —
  no line → quiet success; only an actual change takes the lock and
  re-plans inside it (the bounded re-plan is already the
  abandonment-polled read).
  RED `TestReviewFindingOnDoneNoOpSucceedsOnReadOnlyStore`: verbatim RED
  "open …/.moai-store-lock: permission denied" (exit 1) → GREEN ×2
  (quiet success, empty summary).

**Self-sweep enumeration** (every fold entry path × every in-hold region,
against theme (a) lock-scope discipline and theme (b) no-op parity):

| Entry path | No-op discovery | In-hold regions | Disposition |
|---|---|---|---|
| Verb preview (`--yes` off, ±`--json`) | unlocked plan (row 3) | none — no lock taken | (a) ✓ (b) ✓ |
| Verb apply (`--yes`, ±`--json`) | unlocked plan (row 3) | authoritative re-plan (bounded reads n/a — the verb is unbounded by design), `mutateDisk` seam (test-only), `applyFold`'s temp/cmp/rename critical section; result print DEFERRED outside (finding 1) | (a) ✓ after this pass (b) ✓ |
| On-done auto-fold (`foldOnDoneStep`, all close paths) | bounded unlocked plan (finding 2) | bounded re-plan (`snapshotStoreBounded`), `applyFold` with forbidden-threaded reads; the summary string is RETURNED and printed after release | (a) ✓ after this pass (b) ✓ |
| Direct `atomicWriteFoldFile` callers | n/a (write primitive) | the write's own cmp/probe/final-cmp/rename geometry | (a) ✓ |
| `acquireFoldStoreLock` production callers | — | `withFoldStoreLock` (verb, on-done) only; tests pass nil/plain | ✓ |

No further deviations found: every read inside either hold is
bounded/cancellable on the bounded path; every output write, log line,
and no-op answer now happens outside the hold; both no-op-parity entry
paths discover no-op before requiring write access.

Re-measure judgment: the hold spans SHRANK (printing moved out) and the
on-done path restructured (double-read on the bounded path) — the family
sweep re-ran at `-count=2 -race`; outcome below.

**Row-4 sweep outcome**: PASS — single-pass family green first
(`ok … 96.320s`), then `-count=2 -race -v`: `ok
github.com/modu-ai/moai-adk/internal/cli 242.724s`, exit 0, **0 data
races**, 0 failures, 116 `=== RUN` lines (the fold family + the
print-outside and on-done no-op tests × 2). Builds native + windows
green; gofmt/vet/lint clean (0 issues). The full-suite verdict remains
CI's job (C2) and is PENDING.

### Post-close repair row 5 (committed-instrument race in the FIFO fixture — test-fixture concurrency, not production)

- **Finding — the fixture's cleanup raced the fold worker on
  `memoryFoldSeam`.**
  `TestReviewFindingAbandonedFoldReleasesLockOnApplyReads`'s cleanup
  restored the global seam without synchronizing the spawned fold
  worker's exit (the caller's timeout return creates no happens-before
  edge with the worker; the `fifoCh` default branch had no termination
  sync) — the worker's probe write/read could race the restore under
  `-race` (gate-reproduced on the target test; scheduling-dependent —
  single short runs passed).
- **RED account (honest)**: NOT reproduced locally — 70 race iterations
  across three bound configurations of the unfixed test (300ms ×20,
  1ms ×20 — where the worker abandons before the apply and never touches
  the seam — and 200ms ×30) all came back clean on this machine; the
  gate's reproduction stands as the defect evidence, and the window is
  scheduling-dependent (the local machine's worker reliably reaches the
  probe far before the cleanup). Recorded as a Gap, never as a pass.
- **Fix (the coordinator's directed sync)**: the fixture now sets
  `foldOnDoneExit` to its own channel and registers a cleanup that —
  BEFORE the seam restore (registered first, runs last) — releases the
  FIFO (if the probe fired) and waits `<-exit` for the worker's own
  deferred exit signal: every worker seam access happens-before the
  restore, in both the `fifoCh` receive and default branches.
- **Post-fix verification**: the target test `-count=5 -race` → clean
  (0 races, 0.48s/iter); then the fixture's family at `-count=5 -race`
  (`-run '^TestReviewFinding'`, the full owned set) → PASS, `ok …
  499.177s`, **0 data races**, 0 failures, 140 `=== RUN` lines.

### Sync-stage — AC-DI-012 test-family list, recorded before the re-run (card t1595)

The sync audit found this family list absent from §E.2 before the re-run, and not derived from the touched functions' callers. This record derives it now, at the tree the re-run measures: HEAD `08b9b31c0`, base `81786284e` (`git merge-base HEAD main`). It corrects the M1 record above, which named `factory_bundle_test.go` as the only entered family. The caller trace adds `factory_t1533_test.go` (21 tests), `memory_budget_test.go` (6 tests, through a touched constant), and the review-observation families.

**Pre-flight (re-verified for this record).**

- `git rev-parse --show-toplevel` → `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1595`; `git branch --show-current` → `WT-dispatch-integrity`; `git rev-parse --short HEAD` → `08b9b31c0`; `git status --short` → empty.
- `git diff --name-only 05a812410..HEAD` → `.moai/specs/SPEC-DISPATCH-INTEGRITY-001/progress.md` (the only file).
- `git diff --name-only 81786284e..HEAD -- '*.go'` → production: `internal/cli/factory_bundle.go`, `internal/cli/memory_fold.go`, `internal/cli/memory_fold_lock_unix.go`, `internal/cli/memory_fold_lock_windows.go`; tests: `internal/cli/memory_budget_test.go`, `internal/cli/memory_fold_test.go`, `internal/cli/review_observation_test.go`, `internal/cli/review_observation_fifo_unix_test.go`, `internal/cli/review_observation_publish_fifo_unix_test.go`, `internal/hook/review_observation_test.go`.
- The hunk headers of `git diff -U0 81786284e..HEAD` on the four production files name the touched symbols below. A symbol counts as touched when a hunk lands inside its body; every symbol of a new file counts as touched.

**Touched production symbols and their callers** (line numbers at HEAD; "changed" is the set of changed lines inside the symbol):

| Symbol | Location | Changed lines | Direct callers (file:line) |
|---|---|---|---|
| `newMemoryFoldCmd` | memory_fold.go:100 | 129–234, inside the RunE closure opened at :112 (the constructor at 100–111 is unchanged) | memory.go:231 (in `newMemoryCmd`, memory.go:210; registered by `init` at memory.go:559–560); tests memory_fold_test.go:51, review_observation_test.go:420 |
| `applyFold` | memory_fold.go:466 | 477, 489, 505–508, 530, 549 | memory_fold.go:216 (inside the `withFoldStoreLock` closure opened at :201); memory_fold.go:1365 (inside the closure opened at :1350) |
| `verifyArchiveEffectiveState` | memory_fold.go:563 | 563–570 | memory_fold.go:530, :549 (applyFold) |
| `readFileBounded` (new) | memory_fold.go:598 | 593–625 (with its doc comment) | memory_fold.go:508 (applyFold), :630 (checkFoldUnchanged) |
| `checkFoldUnchanged` | memory_fold.go:629 | 629–630 (gains the `forbidden` parameter) | memory_fold.go:477, :489 (applyFold); :729, :733, :751, :755 (atomicWriteFoldFile) |
| `atomicWriteFoldFile` | memory_fold.go:665 | 729, 733, 743–765 | memory_fold.go:497, :545 (applyFold); tests review_observation_test.go:123, review_observation_publish_fifo_unix_test.go:65 |
| `foldLockName` (const, new) | memory_fold.go:784 | 784 | memory_fold.go:788 (foldLockPath); tests memory_fold_test.go:218, review_observation_publish_fifo_unix_test.go:88, memory_budget_test.go:204 |
| `foldLockPath` (new) | memory_fold.go:787 | 787–789 | memory_fold_lock_unix.go:35; memory_fold_lock_windows.go:29 |
| `withFoldStoreLock` | memory_fold.go:812 | 812–819 | memory_fold.go:201 (newMemoryFoldCmd RunE); memory_fold.go:1350 (foldOnDoneStep) |
| `foldOnDoneStep` | memory_fold.go:1290 | 1323–1329, 1343–1372, 1375–1387 | through the seam below; tests memory_fold_wiring_test.go:680, review_observation_test.go:480, :564 |
| `foldAbandonedReadPoll` (const, new) | memory_fold.go:1380 | 1375–1387 | memory_fold.go:611, :621, :1398, :1408; memory_fold_lock_unix.go:67; memory_fold_lock_windows.go:62 |
| `snapshotStoreBounded` (new) | memory_fold.go:1388 | 1375–1410 | memory_fold.go:567 (verifyArchiveEffectiveState); memory_fold.go:1329, :1351 (foldOnDoneStep) |
| `acquireFoldStoreLock` (new, unix and windows) | memory_fold_lock_unix.go:34; memory_fold_lock_windows.go:28 | whole files | memory_fold.go:813 (withFoldStoreLock); tests review_observation_fifo_unix_test.go:81, :128; review_observation_test.go:439, :553, :587, :597; review_observation_publish_fifo_unix_test.go:92, :109 |
| `releaseFoldStoreLockFunc` (new, unix and windows) | memory_fold_lock_unix.go:71; memory_fold_lock_windows.go:66 | whole files | memory_fold_lock_unix.go:52, :57; memory_fold_lock_windows.go:47, :52 |
| `runFactoryBundleLocked` | factory_bundle.go:75 | 76–89 (the duplicate-member refusal) | factory_bundle.go:63 (closure passed to `factoryLeaseSection` by `runFactoryBundle`, factory_bundle.go:55); reached by tests through `runFactory(t, "bundle", …)` (family F3, F6, F7) |

**Seam chain (traced through function values, not only direct calls).** `foldOnDoneStep` is reached in production only through the seam `foldOnDoneStepFn` (memory_fold.go:1096), called at memory_fold.go:1180 inside the worker of `foldClosedCardMemory` (memory_fold.go:1154). That call runs only when `foldOnDoneGateOpen()` (memory_fold.go:1140–1146, reading `MOAI_MEMORY_FOLD_ON_DONE`) is true; the gate is checked first, at memory_fold.go:1155–1157. `foldClosedCardMemory` is itself reached through the seam `foldClosedCardMemoryFn` (memory_fold.go:1089), whose production callers are `todo.go:1382` (inside the RunE opened at todo.go:1277, in newTodoDoneCmd at todo.go:1270), `todo_auto.go:368` (runAutoCycle, todo_auto.go:252), and `todo_autodone.go:260` (runTodoAutoDone, todo_autodone.go:187). The package-init registrations (memory.go:559–560; factory_handoff_recover.go:113) construct commands but run no RunE, so they are not an entry for the changed lines. The function-value seams `memoryFoldSeam` and `factoryBundleRecord` are data or callee seams, not callers, and do not extend the chain.

**Families entered by the change (re-run these).** A family is one test file, as in AC-DI-012. A family is entered when one of its tests executes a changed line, directly or through a caller chain whose steps all run at test time. Grep counts are `^func Test`; `go test -list` counts were taken at HEAD `08b9b31c0`, env-scrubbed, all exit 0.

| # | Family (file) | Tests | Entry chain (file:line) | Grep count | `go test -list` count |
|---|---|---|---|---|---|
| F1 | memory_fold_test.go | 14 | `runMemoryFold` (:48) calls `newMemoryFoldCmd` (:51) → RunE → applyFold (memory_fold.go:216) → checkFoldUnchanged, readFileBounded, verifyArchiveEffectiveState, atomicWriteFoldFile → acquireFoldStoreLock (memory_fold.go:813) | 14 | 14 |
| F2 | memory_fold_wiring_test.go | 15 | direct seam set at :661 and direct call at :680 (foldOnDoneStep); gate open at :505–902; close path runWireAutoCycle (:313) → runAutoCycle (:323) → todo_auto.go:368 → foldClosedCardMemory → memory_fold.go:1180 → foldOnDoneStep | 15 | 15 |
| F3 | review_observation_test.go | 19 | :123 atomicWriteFoldFile; :420 newMemoryFoldCmd + Execute; :439, :553, :587, :597 acquireFoldStoreLock; :480, :564 foldOnDoneStep; :632, :670, :721 `runFactory(t, "bundle", …)` → runFactoryBundleLocked | 19 | 19 |
| F4 | review_observation_fifo_unix_test.go | 2 | :81, :128 acquireFoldStoreLock; gate open at :31, :109; :72, :122 foldClosedCardMemory → foldOnDoneStep | 2 | 2 |
| F5 | review_observation_publish_fifo_unix_test.go | 2 | :65 atomicWriteFoldFile; :92, :109 acquireFoldStoreLock | 2 | 2 |
| F6 | factory_bundle_test.go | 13 | fbBundle (:43–49, :46 `"bundle"`) → runFactory (factory_card_test.go:66–75) → newFactoryCommand (factory_handoff_recover.go:19; registers newFactoryBundleCommand at :56) → factory_bundle.go:26 RunE (:35) → runFactoryBundle (:55) → closure (:62–63) → runFactoryBundleLocked (:75) | 13 | 13 |
| F7 | factory_t1533_test.go | 21 | fbBundle at :48, :131, :596; `runFactory(t, "bundle", …)` at :62; same chain as F6. Not in the M1 record. | 21 | 21 |
| F8 | memory_budget_test.go | 6 | the listing test compares names with the touched constant `foldLockName` (memory_fold.go:784) at :204. The file itself changed in this diff. A constant reference rather than a function caller, so included conservatively. | 6 | 6 |
| E3 | AC-DI-011 memory-fold selector (F1 + F2) | 29 | the family named by AC-DI-011: memory_fold_test.go 14 + memory_fold_wiring_test.go 15 | 29 (14 + 15) | 29 |

**Selectors** (exact names from the grep output; each is an anchored alternation over the family's own tests):

```
F1 memory_fold_test.go (14):
^(TestMemoryFold_DryRunWritesNothing|TestMemoryFold_Classification|TestMemoryFold_ReachabilityPreserved|TestMemoryFold_VerbatimFiling|TestMemoryFold_ArchiveSelection|TestMemoryFold_Idempotent|TestMemoryFold_EdgeInputs|TestMemoryFold_ApplyOrderAndAbort|TestMemoryFold_ArchiveEntryLinkKept|TestMemoryFold_MoveBeforeDeletionKeepsLine|TestMemoryFold_NeverExistedTargetRefusedCleanly|TestMemoryFold_MoveDuringMemoryPrepKeepsLine|TestReviewArchiveUpdateDuringEffectiveScan|TestMemoryFold_DuplicateStrongLineFiledOnce)$

F2 memory_fold_wiring_test.go (15):
^(TestMemoryFoldOnDone_ExistingClosePathsContained|TestMemoryFold_ArchiveRecheckedBeforeMemoryRename|TestMemoryFoldOnDone_DisabledDifferential|TestMemoryFoldOnDone_EnabledFolds|TestMemoryFoldOnDone_FailOpen|TestMemoryFoldOnDone_SeededPanic|TestMemoryFoldOnDone_BlockedRead|TestMemoryFoldOnDone_RunsAfterQueueWrite|TestMemoryFoldOnDone_ThreeClosePaths|TestMemoryFoldOnDone_DisabledNeverOpensStore|TestMemoryFoldOnDone_ProductionBoundEffective|TestMemoryFoldOnDone_AbandonedStepWritesNothing|TestMemoryFoldOnDone_TimeoutRecoversTempFile|TestReviewSequentialAbandonedTempOwnership|TestMemoryFoldOnDone_BoundConstantCeiling)$

F3 review_observation_test.go (19):
^(TestReviewFindingMergedPRPredecessor|TestReviewFindingNominatedOverwritesDependency|TestReviewFindingNominatedLeasesAfterPredecessorMerges|TestReviewFindingFoldConcurrentWrite|TestReviewFindingFoldInterleavedArchiveLoss|TestReviewFindingFoldArchiveConcurrentWrite|TestReviewFindingFoldGuardArchiveChange|TestFoldSubprocessHelper|TestReviewFindingResultPrintedOutsideLockHold|TestReviewFindingOnDoneNoOpSucceedsOnReadOnlyStore|TestReviewFindingNoOpFoldSucceedsOnReadOnlyStore|TestReviewFindingPreviewNeedsNoWriteAccess|TestReviewFindingAbandonedFoldExitsLockWait|TestReviewFindingStoreLockIndependentOfTempDir|TestReviewFindingBundleDuplicateMemberRefused|TestReviewFindingBundleHeadHubConstraint|TestReviewFindingBundleMultiHubMemberWaits|TestReviewFindingNoRecordArmSkipsBlockedCandidate|TestReviewFindingMergingRetryValidatesLeaseBeforeRemote)$

F4 review_observation_fifo_unix_test.go (2):
^(TestReviewFindingAbandonedFoldReleasesLockOnApplyReads|TestReviewFindingAbandonedFoldReleasesStoreLock)$

F5 review_observation_publish_fifo_unix_test.go (2):
^(TestReviewFindingAbandonedFoldDoesNotPublish|TestReviewFindingLockFileRefusesSymlink)$

F6 factory_bundle_test.go (13):
^(TestFactoryNextBundleSerialLane|TestFactoryBundleKeepsSerialSlot|TestFactoryAssignBundleOrderGuard|TestFactoryAssignBundleHubChain|TestFactorySerialBundleHeadLeasesDespitePickedMembers|TestFactoryBundleLoadAtomicWhenAMemberIsLeased|TestFactoryNextSkipsHubCandidateWhosePredecessorIsUnmerged|TestFactoryNextAfterGuardCountsOtherRunsMerges|TestFactoryHubChainRequiresSharedHubPath|TestFactoryHubChainChainsToTheLastPredecessor|TestFactoryNominateRefusesForeignBundleMember|TestFactoryBundleRecordsUnderTheQueueLock|TestFactoryKeepSetReadsNoFileOverlap)$

F7 factory_t1533_test.go (21):
^(TestFactoryBundleRefusesAnotherLanesMember|TestFactoryNextHubChainPickedRowReleasesSerialSlot|TestReviewNominatedBundlePreservesHint|TestFactoryBundleHeadCarriesHubCondition|TestFactoryNextMergedPRPredecessorReleasesFollower|TestReviewPickedHubSkip|TestFactoryCompleteMergingRetryRefusesForeignLane|TestFactoryCompleteMergingRetryRefusesExpiredLease|TestReviewNominatedWaitsOnAllHubPredecessors|TestReviewNominatedLeaseDoesNotFillAnEmptyHint|TestReviewUnrecordedPickedHubWait|TestFactoryCompletePRMergePinnedToCardTip|TestFactoryBundleHeadIgnoresOwnMembers|TestReviewReversedBundleLeasesInBundleOrder|TestFactoryCompletePRMergePinnedToCheckTimeTip|TestFactoryCompleteRechecksLeaseBeforeRemoteMutations|TestReviewHubWaitNeverClosesACycle|TestReviewGenerationNeverReversesAnAfterChain|TestReviewNominatedCreationFollowsTheAfterRelation|TestReviewHubWaitCoversLaterSharers|TestReviewWaitFollowsTheAfterRelation)$

F8 memory_budget_test.go (6):
^(TestMemoryDoctor_Measures|TestMemoryDoctor_TopicCapUnchanged|TestMemoryDoctor_BudgetBoundaries|TestMemoryDoctor_BytesProxyWarns|TestMemoryDoctor_LinkClasses|TestMemoryDoctor_IndexSelfLinkNotDangling)$

E3 AC-DI-011 (F1 then F2, 29 names, one alternation):
^(TestMemoryFold_DryRunWritesNothing|TestMemoryFold_Classification|TestMemoryFold_ReachabilityPreserved|TestMemoryFold_VerbatimFiling|TestMemoryFold_ArchiveSelection|TestMemoryFold_Idempotent|TestMemoryFold_EdgeInputs|TestMemoryFold_ApplyOrderAndAbort|TestMemoryFold_ArchiveEntryLinkKept|TestMemoryFold_MoveBeforeDeletionKeepsLine|TestMemoryFold_NeverExistedTargetRefusedCleanly|TestMemoryFold_MoveDuringMemoryPrepKeepsLine|TestReviewArchiveUpdateDuringEffectiveScan|TestMemoryFold_DuplicateStrongLineFiledOnce|TestMemoryFoldOnDone_ExistingClosePathsContained|TestMemoryFold_ArchiveRecheckedBeforeMemoryRename|TestMemoryFoldOnDone_DisabledDifferential|TestMemoryFoldOnDone_EnabledFolds|TestMemoryFoldOnDone_FailOpen|TestMemoryFoldOnDone_SeededPanic|TestMemoryFoldOnDone_BlockedRead|TestMemoryFoldOnDone_RunsAfterQueueWrite|TestMemoryFoldOnDone_ThreeClosePaths|TestMemoryFoldOnDone_DisabledNeverOpensStore|TestMemoryFoldOnDone_ProductionBoundEffective|TestMemoryFoldOnDone_AbandonedStepWritesNothing|TestMemoryFoldOnDone_TimeoutRecoversTempFile|TestReviewSequentialAbandonedTempOwnership|TestMemoryFoldOnDone_BoundConstantCeiling)$
```

The M4 AC-DI-011 selector (prefix alternation) is superseded for this re-run by the exact-name E3 selector. Its prefixes also reached review_observation_test.go's `TestReviewFindingFold*` instruments, which belong to F3 and not to the AC-DI-011 family.

**Families not entered (grep evidence for each exclusion).**

- factory_card_test.go (10): its only todo call is `runTodo(t, "add", …)` (:50). `bundle`, `fold`, `--auto`, `runAutoCycle`, and `"done"` each match 0 times in the file. Not entered.
- factory_card_pr_test.go (15): `runTodo(t, "add", …)` at :807 and :814. Its three `--auto` matches (:321, :322, :460) are `gh pr merge … --auto` argument literals, not the todo auto cycle. `bundle` and `fold` each match 0 times. Not entered.
- factory_card_pr_guard_test.go (1): `runTodo`, `fold`, `bundle`, `--auto`, `"done"`, and memory-command references all match 0 times. Not entered.
- factory_nominate_test.go (26): reaches the fold seam through `runTodo(t, "--auto", "--auto-wait", "1ms")` (:1317, :1368, :1388) → todo.go:290 → runAutoCycle → todo_auto.go:368 → foldClosedCardMemory, which returns at memory_fold.go:1155 because the file never sets the gate. `newFactoryCommand().Find` at :395 is a lookup only. Reached with the gate closed: not entered.
- The other 23 test files that reach the seam through `done`, `auto-done`, `--auto`, `runAutoCycle`, or `newTodoDoneCmd`: factory_quota_lanes_test.go, todo_analysis_test.go, todo_audit_regression_test.go, todo_auto_lane_test.go, todo_auto_doc_test.go, todo_auto_rank_test.go, todo_auto_test.go, todo_autodone_test.go, todo_done_ref_test.go, todo_done_verdict_test.go, todo_history_test.go, todo_history_stamps_test.go, todo_hold_predicate_test.go, todo_issuance_test.go, todo_landed_archived_test.go, todo_landing_test.go, todo_landing_roundtrip_test.go, todo_lazy_landedref_test.go, todo_relation_filter_test.go, todo_test.go, todo_transition_stamps_test.go, todo_undone_test.go, todo_verdict_readjudication_test.go. Each reaches foldClosedCardMemory and returns at the gate. The gate evidence covers all of them: `grep -rn 'EnvMemoryFoldOnDone|MOAI_MEMORY_FOLD_ON_DONE' internal cmd` lists only the config declaration (envkeys.go:340–347), the default comment (defaults.go:530), the gate reader (memory_fold.go:1136–1141), and the two test files that set it: memory_fold_wiring_test.go (:45, :204, :471, :505, :533, :580, :625, :682, :696, :782, :819, :902) and review_observation_fifo_unix_test.go (:31, :109). No other test file sets the gate. Not entered.
- memory_drain_test.go: `newMemoryCmd()` (:43, :56) constructs newMemoryFoldCmd, but only the constructor runs, since the changed lines sit inside the RunE. The file runs `drain` (:47) and `--help` (:60), and its `fold` count is 0. Not entered.
- memory_test.go, memory_worktree_key_test.go, update_test.go, update_preserve_reach_test.go, migrate_profiles_test.go: `fold` (case-insensitive) matches 0 times; none calls a touched symbol. Not entered.
- factory_two_hub_cycle_test.go, factory_lease_operator_test.go: use only the helpers `fbSeedFiles` and `fbLeasedCard`; `fbBundle` and `"bundle"` each match 0 times. Not entered.
- memory_fold_wiring_fifo_unix_test.go and memory_fold_wiring_fifo_windows_test.go: contain 0 `^func Test` (helpers `wireFIFO`, `blockOnRead`, `blockOnReadRaw`). A helper file is not a family.
- internal/hook/review_observation_test.go (1 test, TestReviewFindingZoneExistingDotDot): a different package. Its only match against the cli package path in internal/hook is a string literal (internal/hook/mx/validator_fanin_source_test.go:182, inside `strings.HasPrefix`), so there is no import edge. Not entered. This file is the AC-DI-013 control and runs in every verification pass as a separate check.

The M2 and M3 confirmation sweeps (above) ran factory_card, factory_nominate, and factory_card_pr families as confirmation runs. This record does not enter those families, so those passes stand as confirmation only, not as entered-family evidence.

**Gaps (not observed in this record).**

- Only `go test -list` was run. The names and counts are observed; no test in any family was executed here, and no pass is claimed for any family.
- The caller trace is lexical: grep over `internal/` and `cmd/` for each touched identifier and each function-value seam, with the seam assignments and call sites read directly. No type-checked call graph was built, because only `go test -list` is permitted in this record.
- The windows-only symbols (`acquireFoldStoreLock` and `releaseFoldStoreLockFunc` in memory_fold_lock_windows.go) and the windows-tagged test file were not compiled on this darwin host. No GOOS=windows measurement was taken.
- The gate-closed exclusions assume `MOAI_MEMORY_FOLD_ON_DONE` is unset in the re-run environment. It was unset in the worktree shell when this record was taken. The env-scrub form in this section does not unset it.
- The AC-DI-013 control (`TestReviewFindingZoneExistingDotDot`, internal/hook) was not run here.

**Residual-risk.**

- The 24 gate-closed files (factory_nominate_test.go, factory_quota_lanes_test.go, and 22 todo test files, with the todo group listed above) would enter the fold body if `MOAI_MEMORY_FOLD_ON_DONE` were set in the ambient environment. Their exclusion holds only while the gate stays closed. A later change to the todo close paths or to `foldClosedCardMemory` needs a fresh derivation.
- `-list` proves selection, not execution. The re-run must count `=== RUN` lines per family (run with `-v`): a skipped or under-selected family otherwise reads as green. The M4 selector-defect record above shows that failure mode.

### Sync-stage — AC-DI-012 re-run results and Craft measurement (card t1595, tree dbadebc1c)

**Attribution.** The figures below were measured by the coordinator with the tree clean at each measurement, at HEAD `dbadebc1c`, except the lint run in item 4, which was measured at HEAD `08b9b31c0`. This subsection re-executed no command. It re-read each evidence file under `.moai/reports/t1595/` at 2026-10-09T19:24Z and checked the counts and test names stated here against that file. Exit codes, which the evidence files do not carry, are the coordinator's statements and are marked as such. Code identity: `git diff --name-only 08b9b31c0..HEAD` (checked at 2026-10-09T19:22Z) lists `.moai/specs/SPEC-DISPATCH-INTEGRITY-001/progress.md` only.

**1. AC-DI-011 (the 29-name selector recorded above, `-count=5 -race`).**

- Command: `unset MOAI_PROJECT_DIR MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -count=5 -race -v -timeout 30m -run '<AC-DI-011 selector>' ./internal/cli/`, exit 0 (coordinator's statement).
- Evidence: `.moai/reports/t1595/ac-di-011-dbadebc1c.txt`. Read back: top-level `=== RUN` 145; top-level `--- PASS` 145; `--- FAIL` 0; `DATA RACE` 0; distinct top-level test names 29, identical to the selector's names.
- Final lines (verbatim):

```
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	105.491s
```

- Verdict: PASS.

**2. AC-DI-012 families.** Command per family: `go test -count=1 -v -timeout 30m -run '<Fn>' ./internal/cli/`, where `<Fn>` is the family's selector recorded above. Evidence: `.moai/reports/t1595/ac-di-012-F<n>-dbadebc1c.txt`. For each run, the distinct top-level test names equal the family's name list.

| Family | Top-level RUN | Top-level PASS | SKIP | Subtests RUN / PASS | `ok` time (package `github.com/modu-ai/moai-adk/internal/cli`) |
|---|---|---|---|---|---|
| F1 memory_fold_test.go | 14 | 14 | 0 | 0 / 0 | 2.556s (exit 0, coordinator's statement) |
| F2 memory_fold_wiring_test.go | 15 | 15 | 0 | 0 / 0 | 13.719s |
| F3 review_observation_test.go | 19 | 18 | 1 | 6 / 6 | 88.212s |
| F4 review_observation_fifo_unix_test.go | 2 | 2 | 0 | 0 / 0 | 2.417s |
| F5 review_observation_publish_fifo_unix_test.go | 2 | 2 | 0 | 0 / 0 | 1.092s |
| F6 factory_bundle_test.go | 13 | 13 | 0 | 0 / 0 | 26.822s |
| F7 factory_t1533_test.go | 21 | 21 | 0 | 8 / 8 | 44.794s |
| F8 memory_budget_test.go | 6 | 6 | 0 | 0 / 0 | 0.681s |
| Sum | 92 | 91 | 1 | 14 / 14 | |

- The RUN figures are top-level test counts. The raw `=== RUN` line count is 25 for F3 and 29 for F7, because subtests print `=== RUN` lines too. All 14 subtests (6 in F3, 8 in F7) PASS. No other family has subtests.
- The SKIP is `--- SKIP: TestFoldSubprocessHelper (0.00s)`, line 24 of `ac-di-012-F3-dbadebc1c.txt`. Its source, `internal/cli/review_observation_test.go:354–364`, skips when the environment variable `foldSubprocessDirEnv` is empty (lines 355–358), with the reason "fold subprocess entry — only reached via TestReviewFindingFoldInterleavedArchiveLoss". That parent PASSes in the same file (line 15, `(60.02s)`).
- Verdict: every executed test PASS; the one SKIP is by design and must be reported as is.

**3. AC-DI-013 (control).** Command: `go test -count=1 -v -run '^TestReviewFindingZoneExistingDotDot$' ./internal/hook/`, exit 0 (coordinator's statement). Evidence: `.moai/reports/t1595/ac-di-013-dbadebc1c.txt`. Read back: top-level `=== RUN` 1; `--- PASS` 1. Final lines (verbatim):

```
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook	0.566s
```

- Verdict: PASS.

**4. Craft E5 (lint).** Command: `golangci-lint run ./internal/cli/... ./internal/hook/...` with the project config, at HEAD `08b9b31c0`, exit 0 (coordinator's statement). Evidence: `.moai/reports/t1595/craft-lint-08b9b31c0.txt`, which reads exactly `0 issues.` Code identity to `dbadebc1c` holds (see Attribution).

- Verdict: PASS.

**5. Craft E3 (coverage).** Verdict for the coverage measurement: NOT MET.

**5a. `internal/hook`.** Command: `go test -count=1 -timeout 15m -coverprofile=<profile> ./internal/hook/` (profile `.moai/reports/t1595/cover-hook-dbadebc1c.out`), exit 0 (coordinator's statement). Evidence `craft-cover-hook-dbadebc1c.txt` reads:

```
ok  	github.com/modu-ai/moai-adk/internal/hook	179.649s	coverage: 87.3% of statements
```

`go tool cover -func` (`craft-cover-hook-func-dbadebc1c.txt`, line 833) reads `total: (statements) 87.3%`.

- Against the configured package target (`.moai/config/sections/quality.yaml` `test_coverage_target: 85`): MET (87.3% ≥ 85%).
- Against the documented critical target (`.moai/docs/local-dev-guide.md:169`, 90%): NOT MET. That line says its mechanical basis is only the strict evaluator profile's global "Coverage >= 90%" gate, and that no per-package rule was found.
- Source of the 90% target: `.moai/docs/local-dev-guide.md:169` lists "Critical packages (cli, template, hook): 90%+ coverage". The same line says its mechanical basis is only the strict evaluator profile's global "Coverage >= 90%" gate, and that no per-package rule was found. `.moai/config/sections/quality.yaml` sets `test_coverage_target: 85`; the verdict above is against the 90% critical target.

**5b. `internal/cli` (run did not complete).** Command: `go test -count=1 -timeout 40m -coverprofile=<profile> ./internal/cli/` (profile `.moai/reports/t1595/cover-cli-dbadebc1c.out`), exit 1 (coordinator's statement). Evidence `craft-cover-cli-dbadebc1c.txt`:

- line 1761: `coverage: 77.6% of statements`
- line 1762: `panic: test timed out after 40m0s`
- line 4824: `FAIL	github.com/modu-ai/moai-adk/internal/cli	2401.194s`

- Running at the alarm (lines 1763–1765): `TestTodoReadSurface_SilentWithoutJSON (2s)` and `TestTodoReadSurface_SilentWithoutJSON/list (1s)`. The coordinator reads both as progressing, not stuck. The file gives only these elapsed times at the alarm and no other progress trace.
- `--- FAIL` lines: 14, all before the alarm: 10 top-level and 4 subtests. Top-level (file line): TestCountCodexAgentTOMLs_BadPatternRoot (111), TestCountCodexAgentTOMLs (113), TestCheckCodexWiring_MirrorAbsentAdvisesRedeploy (743), TestCheckCodexWiring_DanglingMirrorEntriesCounted (745), TestCheckCodexWiring_CopyModeDetailOnly (747), TestCheckCodexWiring_UnmirroredSkillsDetailOnly (749), TestCheckCodexWiring_MirrorUnreadableIndeterminate (751), TestCheckCodexWiring_MirrorSummaryWidth (753), TestCheckCodexWiring_MirrorFindingParticipatesInTailDrop (758), TestCheckCodexWiring_MirrorUsesExistingRowTwoRegisters (763). Subtests: TestCheckCodexWiring_MirrorSummaryWidth/absent (754) and /dangling (756); TestCheckCodexWiring_MirrorFindingParticipatesInTailDrop/tail_drop_keeps_the_lead_and_detail_keeps_everything (759) and /lead_summary_exception_stays_unreachable_for_mirror_summaries (761).
- First message, line 112: `codex_readiness_test.go:545: count = 6, want 0 (a glob error degrades to zero, never an error)`.
- The ten top-level codex tests are defined in `internal/cli/codex_readiness_test.go` and `internal/cli/doctor_codex_test.go`. Neither file is an AC-DI-012 family file, and neither is among the Go files changed since `81786284e`.

Partial figures from the incomplete profile (not a completed-package measurement): `go tool cover -func` total 77.8% (`craft-cover-cli-func-dbadebc1c.txt`, line 2689). The go-test line says 77.6%. Not reconciled; both are recorded.

Touched-function coverage, same partial profile (`craft-cover-cli-func-dbadebc1c.txt`):

| Symbol | Location | Coverage |
|---|---|---|
| newMemoryFoldCmd | memory_fold.go:100 | 90.0% |
| applyFold | memory_fold.go:466 | 88.9% |
| verifyArchiveEffectiveState | memory_fold.go:563 | 94.7% |
| readFileBounded | memory_fold.go:598 | 93.3% |
| checkFoldUnchanged | memory_fold.go:629 | 83.3% |
| atomicWriteFoldFile | memory_fold.go:665 | 86.8% |
| foldLockPath | memory_fold.go:787 | 100.0% |
| withFoldStoreLock | memory_fold.go:812 | 100.0% |
| foldOnDoneStep | memory_fold.go:1290 | 82.7% |
| snapshotStoreBounded | memory_fold.go:1388 | 100.0% |
| acquireFoldStoreLock (unix) | memory_fold_lock_unix.go:34 | 69.2% |
| releaseFoldStoreLockFunc (unix) | memory_fold_lock_unix.go:71 | 100.0% |
| runFactoryBundleLocked | factory_bundle.go:75 | 83.9% |

Read against the 85% floor (`quality.yaml` `test_coverage_target: 85`; the manager-develop postcondition "Coverage >= 85% on modified files"), four touched symbols are below 85% in this partial profile: checkFoldUnchanged (83.3%), foldOnDoneStep (82.7%), acquireFoldStoreLock (unix) (69.2%), and runFactoryBundleLocked (83.9%).

- Verdict: internal/cli coverage NOT MET (the run did not complete, so no completed-package figure exists).
- Craft E3 verdict: NOT MET.

**Gaps.**

- No command was re-executed in this subsection. Exit codes are the coordinator's statements, and the evidence files do not carry them.
- The internal/cli run did not complete. Its per-function figures come from an incomplete profile.
- The ten top-level codex failures (14 `--- FAIL` lines with subtests) are not attributed to the base commit. Their tests are outside the eight AC-DI-012 families and outside the Go files changed since `81786284e`.
- "Progressing, not stuck" for TestTodoReadSurface_SilentWithoutJSON is the coordinator's characterization. The file gives only the elapsed times at the alarm.
- Windows-tagged code (memory_fold_lock_windows.go and memory_fold_wiring_fifo_windows_test.go) was not measured on darwin.
- The 90% hook target rests on the documented target at `.moai/docs/local-dev-guide.md:169`, whose own note says no per-package rule was found.

### Coverage debt and pre-existing finding (leader ruling B)

**Ruling.** Leader ruling B, as stated by the coordinator in a cross-session message received 2026-10-09T23:52Z: coverage is recorded as debt, and the full package suite is CI's job (AGENTS.md §4, "Scope verification to the change"). This subsection claims no coverage target beyond the ones stated below.

**Coverage debt.**

- `internal/hook`: 87.3% (`craft-cover-hook-dbadebc1c.txt`: `ok  	github.com/modu-ai/moai-adk/internal/hook	179.649s	coverage: 87.3% of statements`). Configured target 85 (`.moai/config/sections/quality.yaml` `test_coverage_target: 85`): met. Documented critical target 90 (`.moai/docs/local-dev-guide.md:169`): not met.
- `internal/cli` full package: not measured to completion. The run hit the 40-minute bound (`craft-cover-cli-dbadebc1c.txt`: `coverage: 77.6% of statements`, `panic: test timed out after 40m0s`, `FAIL	github.com/modu-ai/moai-adk/internal/cli	2401.194s`). The 77.6% is a partial figure, not a package figure.
- Touched-function coverage, partial profile (`craft-cover-cli-func-dbadebc1c.txt`): `foldOnDoneStep` 82.7%, `acquireFoldStoreLock` (unix) 69.2%, `checkFoldUnchanged` 83.3%, `runFactoryBundleLocked` 83.9%, `atomicWriteFoldFile` 86.8%. The other touched symbols are 88.9–100%.
- Uncovered blocks in `acquireFoldStoreLock` (`internal/cli/memory_fold_lock_unix.go`; count-0 blocks in `cover-cli-dbadebc1c.out`):
  - lines 41–42 (fstat error path);
  - lines 44–45 (refusal of a non-regular lock file);
  - lines 49–50 (flock error, blocking path);
  - lines 60–61 (non-EWOULDBLOCK flock error, polling path).

  Correction to the first statement of this item, which said all four blocks need an injected system-call failure to reach. That holds for lines 41–42, 49–50, and 60–61. It does not hold for lines 44–45 (verification note below).

  **Verification note (code read at write time; not run).** The open at line 36 uses `O_CREATE|O_RDWR|O_NOFOLLOW`. The comment at lines 32–33 says a FIFO standing in for the lock is refused by the regular-file check at line 43. A FIFO at the lock path should therefore reach lines 44–45 without an injected failure. No test was run to confirm this route.

**Pre-existing finding, codex family (standing rule d-20261009T171628Z-fc47).**

- Scope: `internal/cli/codex_readiness_test.go` and `internal/cli/doctor_codex_test.go`, 59 test functions. Card diff: `git diff --stat 81786284e -- internal/cli/codex_readiness_test.go internal/cli/doctor_codex_test.go` printed nothing (empty output; checked at 2026-10-09T23:55Z).
- Base reproduction (measured by the coordinator; not re-run in this subsection): export of `81786284e`, `go -C <export> test -count=1 -v -timeout 20m -run '<the 59 names>' ./internal/cli/`, exit 1 (as stated by the coordinator; the evidence file carries no exit code). Result: 59 RUN, 49 PASS, 10 FAIL top-level, 4 failing subtests. The same 10 top-level names as the card-tree run. The same first message: `codex_readiness_test.go:545: count = 6, want 0 (a glob error degrades to zero, never an error)`. Evidence: `.moai/reports/t1595/codex-attribution-81786284e.txt`, present at write time. Its top-level RUN, PASS, and FAIL counts, its 4 failing subtests, its first message, and its 10 top-level FAIL names agree with the figures above.
- Classification: a pre-existing finding at the base commit, reproduced there. Cause not diagnosed. The rule's conditions (1)–(4), which the lane record at `lane-wait-20261010.md:87` says must all hold, are held on the decision board and are not restated in this file. The classification follows the coordinator's ruling.
- The gate legs this finding prevented: the full-package run is CI's under ruling B. The affected family tests and the AC-DI-013 control were run at `dbadebc1c`; see the subsections above.

**Process note.** The first full-suite run reported 14 `--- FAIL` lines (10 top-level, 4 subtests), all inside the tests of those two files.

**Supersession.** The Gaps item at line 780 of the preceding subsection ("The ten top-level codex failures ... are not attributed to the base commit") is superseded by this subsection: the attribution is now measured, see the codex finding above.

**Rule conditions, restated with evidence (rule d-20261009T171628Z-fc47).** The conditions are restated here because they are held on the decision board.

- (1) The card did not touch the flagged path: `git diff --stat 81786284e -- internal/cli/codex_readiness_test.go internal/cli/doctor_codex_test.go` printed nothing (2026-10-09T23:55Z).
- (2) The same finding reproduces on the base commit: the export run recorded above (`codex-attribution-81786284e.txt`).
- (3) Both are recorded in the progress record: this subsection and `lane-progress.md`.
- (4) The gate legs the finding prevented: the affected family tests (AC-DI-011, AC-DI-012) and the AC-DI-013 control were run at `dbadebc1c`; vet and build were run at `dbadebc1c` as below; the full-package `go test` stays with CI under ruling B.

**Gate legs, run at `dbadebc1c`.** Background run `bx197cc3e` in the go-test-heavy slot. `lane-progress.md` (entry of 2026-10-09T23:58:37Z) records the three commands below as exit 0 with 0-byte output: build native, build windows, and vet. The output files are empty, which is the normal result for a clean `go build` or `go vet`.

- `go build ./...` → `.moai/reports/t1595/build-native-dbadebc1c.txt`
- `GOOS=windows GOARCH=amd64 go build ./...` → `.moai/reports/t1595/build-windows-dbadebc1c.txt`
- `go vet ./internal/cli/ ./internal/hook/` → `.moai/reports/t1595/vet-dbadebc1c.txt`

## §E.3 Run-phase Audit-Ready Signal

run_complete_at: 2026-10-09T06:34:36+09:00 (source: `git show -s --format=%cI 281d490bc`, the committer time of the run commit; sync re-audit r2 F2. Superseded value 2026-10-09T22:30+09:00, not reconcilable with the commit history)
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
total_run_phase_files: 14 (observed: `git diff --name-only 5a91b5758 281d490bc` lists 14 paths: .moai/specs/SPEC-DISPATCH-INTEGRITY-001/acceptance.md, .moai/specs/SPEC-DISPATCH-INTEGRITY-001/owned-tests/README.md, .moai/specs/SPEC-DISPATCH-INTEGRITY-001/owned-tests/owned_red_tests.go.txt, .moai/specs/SPEC-DISPATCH-INTEGRITY-001/progress.md, .moai/specs/SPEC-DISPATCH-INTEGRITY-001/spec.md, internal/cli/factory_bundle.go, internal/cli/memory_budget_test.go, internal/cli/memory_fold.go, internal/cli/memory_fold_lock_unix.go, internal/cli/memory_fold_lock_windows.go, internal/cli/memory_fold_test.go, internal/cli/review_observation_fifo_unix_test.go, internal/cli/review_observation_test.go, internal/hook/review_observation_test.go; split 4 production (non-test .go) / 5 test (*_test.go) / 5 SPEC (paths under .moai/specs/) = 14; sync re-audit r2 F1 supersedes the earlier 15 with a 5 / 6 / 3 split)
m1_to_mN_commit_strategy: per-milestone commits (M0 345eb6483, M1 96f392d06, M2 271d71ab9, M3 c0a0d7cbd, M4 c97d50a1e) + the post-report repair pass 281d490bc + this audit-ready stamp
notes: 13/13 ACs PASS (6 as committed regression guards with recorded not-reproduced observations per C1 — defects (1)(2)(4)(5)(6)(7); 2 production repairs — defects (3) and (8) — plus the process ACs held green with measured cells EL-001..EL-027). Two in-gate repair rounds, one selector defect, one default-timeout gate kill, and the post-report gate round (2 findings + the abandonment trio's third member) are recorded in §E.2 as failed measurements and repair records, never as passes. The mid-flight gate's 2 findings and the post-report gate's 3 findings were all folded with RED-first tests. The full-suite verdict is CI's job (C2) and PENDING at report time.

## §E.4 Sync-phase Audit-Ready Signal

sync_complete_at: 2026-10-09T17:16:03Z
sync_commit_sha: c282c3920
sync_status: complete
sync_resync_note: this pass re-synced the CHANGELOG entry for the post-close repairs; the previous close sync commit was d6caefb81.
changelog_entry_position: [Unreleased] §Fixed — single SPEC-DISPATCH-INTEGRITY-001 entry (newest-first), covering the four user-observable behaviors (duplicate-member bundle refusal, cross-process `.moai-store-lock` + abandonment-aware waits, read-only preview without write access, merging-retry lease validation) and the four post-close items (5)-(8) (abandoned bounded fold does not publish, store lock file refuses a symlink, `--yes` no-op needs no store lock, fold result printed after unlock)
b12_self_test_a: PASS (first-close record, historical) — pre-emission `grep -c 'SPEC-DISPATCH-INTEGRITY-001' CHANGELOG.md` = 0 (exit 1) before the append; one entry emitted, no duplicate
b12_self_test_b: PASS (first-close record, historical) — reserved-token-aware AC counter on acceptance.md returned live=13 excluded=0 ambiguous=0; the entry cites 13 acceptance criteria AC-DI-001..013 (matches §E.3 ac_pass_count 13; acceptance.md is the SSOT)
b12_self_test_c: PASS — every file path named in the CHANGELOG entry verified with `ls -l` at the re-sync (eight paths): .moai/specs/SPEC-DISPATCH-INTEGRITY-001/spec.md, internal/cli/factory_bundle.go, internal/cli/memory_fold.go, internal/cli/memory_fold_lock_unix.go, internal/cli/memory_fold_lock_windows.go, internal/cli/review_observation_test.go (merging-retry guard for item 4, whose validation landed with card t1533; guards for items 7 and 8), internal/cli/review_observation_publish_fifo_unix_test.go (publish test for items 5 and 6, build-tagged !windows), .moai/specs/SPEC-DISPATCH-INTEGRITY-001/progress.md; the runtime lock name `.moai-store-lock` is not a repository path and is not listed
b12_self_test_resync: PASS — re-sync self-test run after commit 05a812410: `grep -c 'SPEC-DISPATCH-INTEGRITY-001' CHANGELOG.md` printed 1; the reserved-token-aware AC counter on acceptance.md printed 13 (stderr live=13 excluded=0 ambiguous=0); the CHANGELOG entry cites 13 acceptance criteria AC-DI-001..013
frontmatter_status_transitions.in_progress_to_implemented: merged into the sync commit (3-phase close)
frontmatter_status_transitions.implemented_to_completed: merged into the sync commit (3-phase close — no separate Mx commit)
spec_frontmatter: status in-progress → completed; updated 2026-10-09 (unchanged — already the sync date)
plan_acceptance_frontmatter: neither artifact carries a frontmatter block (stateless per the schema; omission permitted), so no `updated:` refresh was applicable
canary_compliance_check: N/A (this SPEC defines no forward-looking policy carrying its own sync tests)
mx_tag_validation: sync sub-step — the touched production files carry @MX:WARN on withFoldStoreLock (with @MX:REASON, cross-process file lock) and @MX:NOTE/@MX:ANCHOR annotations from the run phase; no new tags owed by the sync edit (docs-only change), no stale tags found on the changed symbols
readme_docs_judgment: no README/docs-site staleness — the 4-locale command tables list `moai memory <doctor|archive>` only (`moai memory fold` and the factory verbs are not README-documented), and none of the four behaviors changes a documented surface; no edit emitted (edit-for-its-own-sake avoided)

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

## §G Override and Refusal Record

- 2026-10-09T00:47:01Z SPEC-DISPATCH-INTEGRITY-001 ceiling-refusal outcome=hold reasons="plan-audit ceiling reached (round count 7 >= tier ceiling 2); the verdict matches no admitting arm and holds, entry blocked (REQ-ACE-006) — release path: the split/new-SPEC route of REQ-ACE-005 or an operator decision recorded in progress.md §G" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1595/.moai/reports/t1595/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1595/.moai/reports/t1595/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1595/.moai/reports/t1595/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1595/.moai/reports/t1595/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1595/.moai/reports/t1595/plan-audit-iter6.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1595/.moai/reports/t1595/plan-audit-iter7.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1595/.moai/reports/t1595/plan-audit.md

- 2026-10-09T16:34:46Z SPEC-DISPATCH-INTEGRITY-001 ceiling-exception outcome=granted decided_by=operator relay=leader(run tmnboq, session 069be28e) decision_id=d-20261009T163446Z-cfa7 reasons="operator approved the plan-audit ceiling exception for card t1595 (kickoff refused REQ-ACE-006: round 7 >= ceiling 2); scope: the ceiling hold only; the leader read plan-audit-iter9.md (PASS 0.96, audited_sha 85ee109c6) and lane-progress.md:154-158. The relay text dates the decision 2026-10-10 KST (2026-10-09T16:2xZ); the board record was written 16:34:46Z." evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1595/.moai/reports/t1595/plan-audit-iter9.md

### Post-close repair row 6 (the iter-8 auditor's routed P3 — repair-count phrase miscounted the card's own record)

- **Finding**: the CHANGELOG's t1595 close entry said "five repaired on
  this card, three measured not-reproduced" — miscounting the card's own
  §E.2 record, which shows 2 production repairs (defects (3) and (8)) and
  6 regression guards confirmed (defects (1)(2)(4)(5)(6)(7)). The iter-8
  verdict is PASS 0.96 with the debt disposal directly observed; its
  codex receipt recorded this P3 as overall fail, which the admission
  predicate refuses unconditionally — the auditor routed the unlock.
- **Fix (two lines)**: the CHANGELOG phrase corrected to the §E.2 framing
  ("2 production repairs on this card (defects (3) and (8)), 6 measured
  not-reproduced … confirmed with committed regression guards"), and the
  §E.3 repair-count phrase aligned to the same corrected framing
  (6 regression guards named by defect; 2 production repairs named by
  defect). No other changes.
- **Verification**: the corrected CHANGELOG phrase greps 1-of-1 against
  the corrected counts; the §E.3 phrase names the same defect split as
  §E.2's rows; `golangci-lint` exit 0. No production bytes, no test
  changes.
- 2026-10-09T16:40:10Z SPEC-DISPATCH-INTEGRITY-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 9 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1595/.moai/reports/t1595/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1595/.moai/reports/t1595/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1595/.moai/reports/t1595/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1595/.moai/reports/t1595/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1595/.moai/reports/t1595/plan-audit-iter6.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1595/.moai/reports/t1595/plan-audit-iter7.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1595/.moai/reports/t1595/plan-audit-iter8.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1595/.moai/reports/t1595/plan-audit-iter9.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1595/.moai/reports/t1595/plan-audit.md
- 2026-10-09T16:41:59Z SPEC-DISPATCH-INTEGRITY-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 9 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1595/.moai/reports/t1595/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1595/.moai/reports/t1595/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1595/.moai/reports/t1595/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1595/.moai/reports/t1595/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1595/.moai/reports/t1595/plan-audit-iter6.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1595/.moai/reports/t1595/plan-audit-iter7.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1595/.moai/reports/t1595/plan-audit-iter8.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1595/.moai/reports/t1595/plan-audit-iter9.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1595/.moai/reports/t1595/plan-audit.md
