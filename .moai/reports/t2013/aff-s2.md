# AFF S2 evidence — card t2013 (SPEC-FACTORY-COMPLETION-RECOVERY-001, REQ-FCR-026 / AC-FCR-026)

Slice: under agentic-trunk, when the fixed release branch (`release_branch_prefix` + `main-batch`) is checked out in a non-primary worktree, `moai factory complete` lands the card through the integration window into that branch, after a fast required check. Otherwise the existing routes run unchanged. Decisions: `.moai/reports/pr-repair-20261101/decisions.md` R114 (user approval, original wording recorded), R116 (accepts release route before the main route, Tier L ceiling deviation Q11, branch name fixed `release/main-batch`, Q10).

## Claim

- SPEC amendment `7a28ea1f89` (REQ-FCR-026, AC-FCR-026 a-g, plan M4, decision-index Q8-Q11).
- RED `7341fcc49befc2dabcf6b6e857d861a83d118447`, GREEN `8d318e63f65ff3e13bbb07defebc911bea23def0`.
- The fast check, the release-branch merge through the integration window, the hold record on a failed check or a stopped merge, and the same-SHA no-op are new behavior.

## Evidence

Observed by me in this run (HEAD 8d318e63f6, tree clean):
- `go test ./internal/cli/ -run '^(TestAFFS2.*|TestAgenticTrunk.*|TestFactoryMergeReadyGitHubFlow|TestFactoryGitHubFlowStates|TestFactoryDecidePushGate)$' -count=1` -> `ok  github.com/modu-ai/moai-adk/internal/cli  50.871s`
- `git diff --stat 7a28ea1f89 HEAD` lists 5 files only: `internal/cli/factory_aff_s2_t2013_test.go` (new), `internal/cli/factory_card.go` (+73/-5), `internal/cli/factory_release_route.go` (new), `internal/config/loader_agentic_release_branch.go` (new), `internal/homestate/card_hold.go` (new). No existing test file, rule or template is in the list.

Reported by the implementing agent (not re-run by me):
- RED failing lines: all cases (a)-(f) failed because the old route sent a primary-held main to the PR edge ("merge-readiness failed on sync-audit against main"), hold records = 0, want 1; the regression guard `TestAFFS2CompleteWithoutProvisionedReleaseBranchKeepsMainRoute` was already green at RED.
- GREEN per case: (a) `TestAFFS2CompleteMergesCardIntoReleaseBranch` (b) `TestAFFS2CompleteCheckFailureHoldsAndReleases` (c) `TestAFFS2CompleteConflictHoldsReleasesAndCleans` (d) `TestAFFS2CompleteSameSHARerunIsNoOp` (e) `TestAFFS2CompleteAdoptsMergeAfterCrash` (f) `TestAFFS2CompleteSecondCompleterBlocked` — `ok internal/cli 23.060s`.
- Regression set (g): `-run 'TestAgenticTrunk|TestCodex|TestFactoryMergeReadyGitHubFlow|TestFactoryGitHubFlowStates|TestFactoryDecidePushGate|TestFactoryComplete'` ok 162.654s; `go test ./internal/factory/ -count=1` ok 308.467s; `go test ./internal/homestate/ -run 'Card|Wait|Factory'` ok 32.857s; `TestAgenticTrunkTarget` PASS; go vet and gofmt clean; golangci-lint 0 issues.
- Mutation checks (temporary, reverted, not committed): skip the fast check -> (a),(b) fail; merge into main -> (a)-(e) fail (on the "integration worktree for main is not provisioned" refusal, not on the main-SHA assertion); drop the window release after a failed check -> (b) fails, (c) still passes because RunMergeStep releases on its own pre-merge refusals; disable the no-op probe -> (d) fails, on the missing-lease gate, not on a second merge.
- Where the change lives: factory_card.go:3023-3026 release-route predicate before the PR-edge judgment; :3052-3058 no-op ahead of the not-merge-ready refusal; :3099-3112 branch resolution; :3157 `stop` hold writer; :3182 base-freshness stop; :3261 re-measure-record stop; :3268-3283 fast check then lease renewal; :3320-3333 merge-step stops (a post-merge failure, exit 8, writes no hold). New: factory_release_route.go, loader_agentic_release_branch.go (`LoadAgenticTrunkReleaseBranch`, a separate read so `IntegrationTarget` stays `main`), homestate card_hold.go (`RecordCardHold`, a `card.hold` event, no version bump, no new transition edge).
- Fast check steps (each a `moai verify run` in the card tree): `go build ./...`; lint of the packages the card changed since the release tip (`golangci-lint run --allow-parallel-runners <pkgs>` when binary and config exist, else `go vet <pkgs>`); `go test <pkgs>`. A card tree without `go.mod` fails closed (`check-failed`).

## Applied / not applied

Applied: release route when the release branch is provisioned; fast check; hold record; same-SHA no-op; crash adoption via the existing `factoryAdoptedTip`.
Not applied: retiring the old main route (S4), the release PR verb (S3), any push, creating the release branch, the `-YYYYMMDD` documentation cleanup (S4), a configurable check for non-Go projects.

## Baseline-attribution

Tests ran from this tree with the installed go and golangci-lint (not a build of this tree's `moai`); the tests call `verifyRun` in-process.

## Gaps

- The base-moved gate always fires before a literal git merge conflict in the fixture, so the `merge --abort` path inside RunMergeStep and the `merge-conflict` reason are not exercised by any test.
- The lease-renewal block after the fast check, the `golangci-lint` branch of the fast check, and the verify reuse line are not tested.
- (d) is stopped by the existing lease gate rather than by a second merge being attempted; the double-merge guard itself is not isolated.
- The `/verify` full gate and the full `./internal/cli/...` suite were not run; only scoped sets.
- Other AC-FCR-026 (g) commands (e.g. `TestT2013_PREdgeGuidanceDraftAndNonRequiredChecks`) were not part of the agent's regression set; they were last run at the previous head.
- No independent audit exists for this slice (audit withheld by the leader, R100).

## Residual-risk

- A non-Go project with a provisioned release branch gets `check-failed` on every complete until a configurable check exists.
- `go build ./...` runs on every complete in this repository while the integration window is held; only the untested lease renewal protects a long run.
- A diverged release branch still runs the full check before RunMergeStep refuses it.
- Q11: REQ and AC counts are 26, one over the Tier L ceiling (accepted by the leader, R116).
