# Deferred finding — lane-probe (productionLaneFilesProbe)

Deferral record for the finding SPEC-DISPATCH-INTEGRITY-001 declined
(decision-index.md Q1, DEFAULT-APPLIED 2026-10-08T15:14Z). This directory is
the durable, TRACKED side of the inheritance — the deferring card cannot edit
the owner card's queue body (queue mutation is prohibited for factory lanes
and their spawned agents), and the overlay archive under `.moai/reports/` is
gitignored, so without this directory the finding would survive only in
machine-local artifacts.

- **Finding**: `TestReviewFindingLaneProbe` — `productionLaneFilesProbe`
  mis-measures a lane's changed files: github-flow repos return an empty
  list, and whitespace-containing paths are split on spaces (unicode-path
  also failed on the 2026-10-08 re-run).
- **Owner**: card t1596 (named first among candidate owners by the
  predecessor handoff; t1597 is the fallback candidate if t1596 is closed
  otherwise).
- **Evidence (RED)**: all 3 subtests FAIL on tree 81786284e (2026-10-08 run)
  and re-observed FAIL by the round-1 plan audit on 544462a8d; original
  observation record `.moai/reports/t1595/overlay/final-results.txt`
  (2026-10-07, different tree — labeled as such).
- **Intake (t1596)**: take `lane-probe_test.go.txt` from this directory into
  `internal/cli/` as a `*_test.go` file; re-verify RED on the intake tree
  BEFORE fixing (same-tree baseline rule); classify per the two-cell rule
  (`verification-completeness.md` §2); rename the `reviewGit` helper on
  collision.
- **Surface**: `productionLaneFilesProbe` in `internal/cli` (github-flow
  empty-list branch + whitespace path splitting in the changed-files probe).
