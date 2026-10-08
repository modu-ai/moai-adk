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

## §E.3 Run-phase Audit-Ready Signal

_Pending run-phase (manager-develop)._

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
