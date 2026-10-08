# Owned RED test sources — measurement-input mirror for the evidence ledger

Tracked mirror of the four t1595-owned test sources measured in the
acceptance.md evidence ledger (EL-001..004). Plan-audit iter-2 D10: the
ledger's tree SHA alone does not make a cell replayable — at measurement
time these sources were untracked drop-ins from the gitignored overlay
archive, and the codex gate demonstrated the hazard concretely
(`testing: warning: no tests to run`, exit 0, when replaying the recorded
selectors against the cited revision's checkout). Every ledger cell binds
to the tree SHA AND this input.

- `owned_red_tests.go.txt` — the four test functions verbatim:
  `TestReviewFindingMergedPRPredecessor`,
  `TestReviewFindingNominatedOverwritesDependency`,
  `TestReviewFindingFoldConcurrentWrite`,
  `TestReviewFindingFoldInterleavedArchiveLoss`.

## Drop-in procedure

1. Copy `owned_red_tests.go.txt` into `internal/cli/` as
   `review_observation_test.go` (M0's canonical committed drop-in — after
   M0 the sources live in the branch and this mirror becomes the provenance
   record).
2. Target package: `internal/cli`. The helper symbols the tests call live
   in the repo's existing internal/cli test files (not in this mirror):
   `fcFixture`, `fcQueue`, `fcClassify`, `fcPlace`, `fbSeedFiles`,
   `sdRegisterLane`, `sdLaneEnv`, `runFactory`, `fcCard`, `fcRun`,
   `fbLeasedCard`, `minimalFiles`, `minimalMemory`, `minimalArchive`,
   `seedFoldStore`, `runMemoryFold`, `foldRead`, `foldTestSeam`,
   `fixtureArchive`, `line9001`, plus production test seams
   `atomicWriteFoldFile` / `memoryFoldSeam`. On collision with a helper
   renamed since 81786284e, adapt the call sites — the test bodies are the
   contract, the helper names are not.
3. Re-verify on the intake tree BEFORE any fix (same-tree baseline rule):
   run the ledger's recorded selector, expect the recorded outcome.

## Provenance

Extracted verbatim from the t1498 overlay suite `cli_all_test.go`
(origin `/tmp/t1498-review-overlay/`, archived local-only under
`.moai/reports/t1595/overlay/`). The EL-001..004 measurements dropped the
FULL overlay superset file (7 tests, including 3 foreign-card tests); the
single-test selectors mean the foreign tests did not participate in the
four cells. This mirror carries the owned four only — the t1596-owned
three are mirrored separately under `../deferred/` per the lane-probe
deferral (decision-index Q1).
