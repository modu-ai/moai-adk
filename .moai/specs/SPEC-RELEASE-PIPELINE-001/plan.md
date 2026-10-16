---
id: SPEC-RELEASE-PIPELINE-001
title: "Release pipeline coverage and cost — implementation plan"
version: "0.1.0"
created: 2026-10-17
author: manager-spec
---

# plan.md — SPEC-RELEASE-PIPELINE-001

## §A Context

Card t1536 (operator instruction 10-06, Class C). Worktree `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536`, branch `WT-10-06-codex`, plan-phase baseline HEAD `0bd1da4973d1d9539172f90587ca1ee84c8d9426` (clean main tip; no card commits before this plan-phase commit).

### §A.1 Re-anchor table (10-06 card citation → current HEAD)

| Card item | 10-06 citation | Current anchor at `0bd1da497` | State |
|---|---|---|---|
| 1. Filter lacks template/config inputs | release-pr-multi-os.yml:62; ci.yml:292 | Shared filter `.github/test-input-filters.yml` consumed at ci.yml:79, codeql.yml:43, release-pr-multi-os.yml:68; release/* race exclusions at ci.yml:340/394/446/497/549/602 (by design) | **LANDED** (PR #1770 `0990c922f` + t1478 follow-ups) — M1 verifies and preserves |
| 2. Doc-only PRs run lint/build/constitution/browser | ci.yml:523/:570/:640/:701 | `lint` :772, `build` ×5 :819, `constitution-check` :889, `test-browser` :950 — all carry no `needs:`/`if:` (EV-5) | **OPEN** — M2 |
| 3. Stale prior-HEAD verification + freshness slots | release-pr-multi-os.yml:18; graph-freshness.yml:23 | release-pr-multi-os.yml:20 `cancel-in-progress: false` (EV-2); graph-freshness split LANDED (:24-26 event-kind+PR group, :130-132 guard group, card t1298) | **HALF LANDED** — M3 (release workflow flip + guard) |
| 4. Browser 17m/20m + mutation; release single long race | ci.yml:769; release-pr-multi-os.yml:217 | test-browser job budget 20m with green `-timeout 17m` + mutation cycle in-job; release race step `-timeout 35m` inside `timeout-minutes: 40` (:94/:217, EV-6); retry wrapper wired only at ci.yml:222/:578 | **OPEN** — M4 |
| 5. jq `now\|floor as $now` never executes | release-drafter-cleanup.yml:65 | :65 broken form present, reproduced RED (EV-3/EV-4); no re-query in delete loop (:82-98, EV-7); no concurrency block (EV-8); release-drafter.yml:19-21 group exists with `cancel-in-progress: true` (flip required, D3) | **OPEN** — M5 |

### §A.2 Overlap boundaries (sibling SPECs)

- **SPEC-CANDIDATE-CI-001** (implemented, t1478): owns candidate-branch CI, guard-bundle, six race legs, flaky registry/retry implementation. This SPEC is a consumer/preserver of those surfaces; the shared filter's authorship history is joint (t1536 landed the sharing, t1478 extended the root set) but no REQ here redesigns them.
- **SPEC-CI-DOCTOR-BIN-001** (completed, t346): doctor embed-check verdict — disjoint.
- **SPEC-V3R4-CI-FASTTRACK-001**: origin of release-pr-multi-os.yml and the release-gate aggregate semantics this SPEC's REQ-RPC-003 generalizes to ci.yml's cost family.
- **SPEC-APPJS-FIRE-GUARD-001**: owns the browser job's ADDITIVE-ONLY + phase semantics; M4's split preserves its verdict family and pinned-Chrome discipline.
- **SPEC-GITHUB-FLOW-CI-RESIDUE-001**: owns pr-multi-os-gate.yml — the in-tree read-only `cancel-in-progress: true` precedent REQ-RPC-005 generalizes.

### §A.3 Development mode

`quality.yaml` → `constitution.development_mode: tdd`; coverage target 85. All milestone guards are RED-first Go tests in `internal/template/` (the established pattern: `test_input_filter_parity_test.go` reads the workflow YAML from the repo tree).

## §B Known Issues

- **B1 — golangci-lint from source**: the lint job installs golangci-lint via `go install` (~minutes). Input-gating it (M2) changes when that cost is paid, not the check itself; do not swap the install method in this SPEC.
- **B2 — skipped required checks**: GitHub reports a path-filtered required check as skipped, not success. The repo's existing answers are `test-skip-marker` (success stub) and `release-pr-gate` (aggregate). M2 must pick one coherent surface per decision-index Q4 and keep branch protection satisfied on every class.
- **B3 — workflow YAML is repo input**: any guard test parsing workflow YAML must read the file from the module root relative path (`../../.github/workflows/...` from `internal/template/`), as the parity test does; a copied fixture would drift.
- **B4 — cancel flip vs required gate**: flipping release-pr-multi-os.yml to `cancel-in-progress: true` cancels superseded-SHA runs; `release-pr-gate`'s AC-CI-001 logic already treats `cancelled` as failure for THAT run, and branch protection reads the newest SHA's run — no gate-semantic change is needed, but M3's guard test must pin this reasoning next to the flip.
- **B5 — jq availability**: guard tests executing jq programs require the `jq` binary (present on ubuntu runners; used by graph-freshness.yml). Skip-with-loud-message is not acceptable for the filter-semantics test; if jq is absent locally the test must fail, not skip.
- **B6 — shared group side effects (REQ-RPC-012)**: joining cleanup into a group with the push-triggered drafter can queue a Monday cleanup behind a draft-update burst; the cleanup job budget (15m) and schedule tolerance absorb this. With both sides `cancel-in-progress: false` (D3), the reverse also holds — rapid main pushes queue drafter updates serially (each a minutes-scale draft update converging on the final state) instead of cancelling superseded ones. Record the queuing behavior in BOTH workflow comments.

## §C Pre-flight (baseline commands, run at plan phase)

```
$ go test ./internal/template/ -run '^(TestTestInputFiltersSingleSource|TestTestInputFiltersCoversGoTestInputs)$' -count=1
ok  github.com/modu-ai/moai-adk/internal/template	0.355s        # EV-1, exit 0

$ grep -n "cancel-in-progress" .github/workflows/release-pr-multi-os.yml
20:  cancel-in-progress: false                                   # EV-2, exit 0

$ jq -r --arg threshold_days 30 'now | floor as $now | ...' .moai/reports/t1536/red-drafter-fixture.json
jq: error (at .moai/reports/t1536/red-drafter-fixture.json:6): Cannot iterate over number (1792178114....)   # EV-3, exit 5

$ jq -r --arg threshold_days 30 '(now|floor) as $now | ...' .moai/reports/t1536/red-drafter-fixture.json
stale-draft                                                     # EV-4, exit 0

$ grep -n "gh release view" .github/workflows/release-drafter-cleanup.yml
(no output)                                                    # EV-7, exit 1

$ grep -n "concurrency" .github/workflows/release-drafter-cleanup.yml
(no output)                                                    # EV-8, exit 1

$ grep -n "attempt1" .github/workflows/ci.yml
(no output)                                                    # EV-9, exit 1
```

EV-1..EV-4 executed at tree `0bd1da4973d1d9539172f90587ca1ee84c8d9426`; EV-7..EV-9 (iter-1 repair) at `c8ede48f90fbfe4c6873f300c91243f03928d254` — the plan-artifacts commit changed no workflow file (`git diff --name-only` between the two trees lists only SPEC artifacts), so all seven are interchangeable across both trees. acceptance.md §B carries full verbatim text. Run-phase pre-flight re-runs EV-1 and `git rev-parse HEAD` before the first M-commit.

## §D Constraints — PRESERVE list (normative)

1. `.github/test-input-filters.yml` root set and its three consumers; `internal/template/test_input_filter_parity_test.go` green (EV-1 baseline).
2. `release-pr-gate` semantics (release-pr-multi-os.yml:274-315): cancelled/timed-out/skipped-without-named-exclusion never passes; detect-release hard failure fails the gate; `non-release-pr` and `docs-only-release-pr` are the only named intentional exclusions.
3. ci.yml release/* race exclusion rationale (the release matrix supersedes; its ubuntu leg is the only REQUIRED ubuntu race — do not drop it).
4. `test-skip-marker` intentional-skip reporting for the test family on non-`go_code` PRs.
5. SPEC-CANDIDATE-CI-001 surfaces: `ci/**` push trigger without `paths:` filter, guard-bundle partition, six race legs' exact-cover selectors and per-leg ceilings, `retry-flaky.sh` wiring at ci.yml:222/:578.
6. Least-privilege `permissions:` blocks in every touched workflow (contents: read for verification workflows; the cleanup workflow's `contents: write` stays scoped to `gh release delete`).
7. The release matrix evidence chain: `-json` stream to file, `|| rc=$?` census-on-failure pattern, gzip compression, per-OS artifact names, windows `if-no-files-found: error`.
8. SPEC-APPJS-FIRE-GUARD-001 invariants: pinned Chrome-for-Testing + sha256, AppArmor profile step, probe manifest lint, mutation restore `RESTORED_BYTE_IDENTICAL` assertion, `0 → 1 → 0` cycle.
9. `release.yml` deploy path concurrency (no-cancel) — untouched.
10. graph-freshness landed split (event-kind + PR-number groups; separate guard group) — M3 adds a guard test, changes nothing.
11. Required-name SSoT surfaces (BOTH) — `.github/branch-protection.json.gtmpl` (the contexts `internal/cli/branch_protection.go` renders into the GitHub branch-protection PUT) AND `.github/required-checks.yml` (the SSoT read by `scripts/ci-watch/run.sh`, `scripts/ci-mirror/validate-required-checks.sh`, `scripts/ci-mirror/run.sh`, and `scripts/cutover-protection-compare.sh` for required-vs-auxiliary verdicts) — any required-name change lands as a coordinated update of BOTH files in the same milestone, never as a workflow-only edit and never as a one-file SSoT edit: a name present in one surface and stale in the other leaves the candidate/ci-watch verdict tooling judging an outdated required set (D1).

Forbidden: modifying workflow files during plan phase (done — this phase writes SPEC artifacts only); `git add -A` (stage by explicit pathspec); force-push; touching `.moai/state/**`, `.moai/logs/**`, other SPEC directories.

## §E Self-Verification (run-phase E-matrix summary)

E1 AC matrix from acceptance.md §D (command + verbatim output per AC). E2 `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...`. E3 `go test -cover ./internal/template/` (guard-test package). E4 boundary grep (no AskUserQuestion/mcp__askuser in non-test code touched — none expected, YAML+tests only). E5 `golangci-lint run --timeout=5m` + `moai spec lint --spec SPEC-RELEASE-PIPELINE-001`. E6 commit SHAs + push state. E8 RED outputs per milestone (each M's guard test failing output captured before GREEN).

## §F Milestones (M1-M5, card-item mapping, priority order)

> M1-M5 map 1:1 onto card items 1-5 for traceability. Within the remaining open work, M5 is the only live functional defect (silent no-op of a safety workflow) and is small; M2 is the largest cost win. Priority labels below reflect remaining risk, and the executor may run M5 ahead of M2-M4 if sequencing lanes free up — the card-item mapping (M# ↔ item #) is the traceability contract, not a wall-clock order.

### M1 — Item 1: verify and preserve the landed shared classifier (Priority High, verification-only)

1. Re-run EV-1 at the run-phase HEAD; record in progress.md §E.2.
2. Grep-verify the three consumers still pass `filters: .github/test-input-filters.yml` (ci.yml, codeql.yml, release-pr-multi-os.yml detect jobs).
3. Confirm fail-open semantics intact: `continue-on-error: true` on the release detect's filter step + `!= 'false'` matrix comparison (release-pr-multi-os.yml:62-68, :93).
4. Repair only comment drift found (none expected); no behavior change. Exit criteria: EV-1 green + three-consumer grep recorded.

### M2 — Item 2: input classes + single aggregate gate (Priority High)

1. RED: author `internal/template/ci_cost_gating_test.go` — `TestCICostJobsInputGated` (asserts lint/build/constitution-check/test-browser each carry an input-class gate sourced from the shared filter), `TestCIAggregateGateSemantics` (asserts exactly one fail-propagating aggregate for the family; detect hard-failure propagates; intentional-skip passes with named exclusion), `TestCIRequiredNamesSatisfied` (asserts BOTH required-name surfaces — `.github/branch-protection.json.gtmpl` contexts and `.github/required-checks.yml` `branches.main.contexts` — remain satisfied AND mutually consistent for every classification outcome). Observe RED (EV-5 is the pin).
2. Add the input classes (`binary_inputs`, `browser_inputs`, `harness_inputs`) to `.github/test-input-filters.yml` beside `go_code` — same file, single source; each root carries its evidence comment (which job's verdict reads it).
3. Gate the four cost jobs on their classes (fail-open: unevaluable classification runs the job).
4. Add the aggregate job; update BOTH required-name SSoT surfaces per decision-index Q4's applied default (single aggregate required name for the gated family) in the SAME milestone — `.github/branch-protection.json.gtmpl` contexts AND `.github/required-checks.yml` (`branches.main.contexts`, plus `branches."ci/**"` wherever the aggregate publishes on candidates) — the ci-watch/ci-mirror tooling reads the latter, so a gtmpl-only update leaves the verdict tooling on a stale required set (D1).
5. Preserve REQ-RPC-004: harness roots stay inside `go_code` (release 3-OS coverage unchanged); add one guard-test assertion for that containment.
GREEN: the three tests pass; `go test ./internal/template/ -run '^(TestCICostJobsInputGated|TestCIAggregateGateSemantics|TestCIRequiredNamesSatisfied)$' -count=1`.

### M3 — Item 3: read-only verification cancels superseded runs (Priority Medium)

1. RED: author `internal/template/verification_concurrency_test.go` — `TestReadOnlyVerificationCancels` (every workflow whose jobs mutate nothing outside CI state carries `cancel-in-progress: true`; `release.yml` exempt as deploy path), `TestGraphFreshnessSlotSplit` (pins the landed t1298 split: event-kind + PR-number group, separate guard group). Observe RED on the release workflow flip (EV-2 is the pin).
2. Flip release-pr-multi-os.yml:20 to `cancel-in-progress: true` with the B4 rationale comment (newest-SHA verdict; AC-CI-001 cancelled-never-passes unaffected for the newest run).
3. GREEN: both tests pass.

### M4 — Item 4: stage budgets, browser phase split, retry discipline (Priority Medium)

1. RED: author `internal/template/verification_budget_test.go` — `TestStageCeilingBelowJobBudget` (per-stage ceilings strictly below job `timeout-minutes` with declared headroom for the `always()` evidence steps; covers the release race step and both browser phases), `TestBrowserPhaseSplit` (green and mutation phases carry independent stage bounds; a max-duration green phase leaves the mutation phase its declared budget), `TestRetryPolicyPreservesFirstFailure` (three assertions: retry invocation appears only through the flaky-registry wrapper — green today; `scripts/ci/retry-flaky.sh` writes the attempt1 record before the re-run — green today, script :166 before :169; every ci.yml job that invokes the wrapper carries the attempt1 record in its evidence-upload path — RED today, EV-9). Observe RED (EV-9 is the pin; the third assertion fails while the first two hold).
2. Re-budget the release race step/job relationship so census+compress+upload have explicit headroom (stage ceiling strictly < job ceiling; document the arithmetic in the workflow comment).
3. Split the browser job's phases into separately bounded stages per decision-index Q5's applied default (one job preserved; internal stage ceilings + upload headroom; the SPEC-APPJS-FIRE-GUARD-001 phase order `0 → 1 → 0` and restore assertions unchanged).
4. Keep the registry-gated retry semantics as-is (wrapper behavior unchanged — `retry-flaky.sh` already preserves attempt1 locally before the re-run); ADD the missing evidence half — extend each retry-wrapped job's `Upload test event stream` step path (test job, rosterguard leg, and every other leg invoking `retry-flaky.sh`) to include `.moai/logs/ci-flaky-retry.attempt1.json` (uploads stay `if: always()` with `if-no-files-found: warn`, so a no-retry run stays green). Do NOT extend retries to the macOS/Windows release legs (decision-index Q6, EVIDENCE-NEEDED).
GREEN: the three tests pass.

### M5 — Item 5: drafter-cleanup repair (Priority High — live silent no-op)

1. RED: author `internal/template/release_drafter_cleanup_test.go` — `TestCleanupFilterSelectsStaleDrafts` executes the workflow's extracted jq program against a committed testdata fixture (`internal/template/testdata/release-drafter-cleanup.json`; the plan-phase evidence copy at `.moai/reports/t1536/red-drafter-fixture.json` is gitignored and dies with the worktree — M5 commits the fixture into testdata) and asserts exactly `stale-draft` is selected; `TestCleanupRequeriesBeforeDelete` (delete loop re-queries each tag's draft state before `gh release delete`; EV-7 is the RED pin — zero occurrences today); `TestDrafterCleanupConcurrency` (both draft-mutating workflows share one group with `cancel-in-progress: false` on BOTH sides — the cleanup's new block AND release-drafter.yml:21 after the flip — including the drafter-arrives-while-cleanup-in-flight case: an incoming drafter run queues, never cancels; EV-8 is the RED pin); `TestCleanupDryRunDispatchReachable` (the workflow declares `workflow_dispatch.inputs.dry_run` wired into the env that sets `MOAI_CLEANUP_DRY_RUN` — unreachable today, D2). Observe RED (EV-3/EV-7/EV-8 pin the first three; the fourth fails on the absent input).
2. Fix release-drafter-cleanup.yml:65 to `(now|floor) as $now`.
3. Add the per-tag re-query guard to the delete step (skip + log on state change).
4. Join both draft-mutating workflows to the shared concurrency group (`release-drafter-${{ github.ref }}`) with `cancel-in-progress: false` on BOTH sides: add the cleanup's concurrency block AND flip release-drafter.yml:21 to `false` — in a shared group the INCOMING workflow's own flag governs, so a cleanup-side-only `false` leaves an incoming drafter run cancelling the in-flight cleanup (D3); per decision-index Q3's applied default (queue, never cancel), extended to both sides by the iter-1 repair. Record the queuing trade-off in both workflow comments (B6).
5. Make the dry-run path reachable (D2): add `workflow_dispatch.inputs.dry_run` (boolean) to release-drafter-cleanup.yml and wire it into the env step that sets `MOAI_CLEANUP_DRY_RUN` — the delete step's dry-run branch (:92-93) already exists, but a dispatch cannot set runner env, so the pre-repair instruction (`MOAI_CLEANUP_DRY_RUN=1`) left DRY_RUN=0 and would have performed REAL deletions; `TestCleanupDryRunDispatchReachable` guards the wiring. Post-merge validation per acceptance.md §D.2: dispatch with dry-run enabled and observe NO `gh release delete` executes.
GREEN: the four tests pass; EV-3/EV-4 divergence preserved inside `TestCleanupFilterSelectsStaleDrafts`.

## §G Anti-Patterns

- AP-1: adding an inline classifier copy anywhere (parity guard exists to kill this).
- AP-2: gating a cost job with fail-closed classification (an unevaluated filter must RUN the job — the #1557 stub-over-broken-tree hazard).
- AP-3: satisfying a required name with a success stub over an UNEVALUATED run (stub is legitimate only for positively observed non-matching classification).
- AP-4: flipping `cancel-in-progress` on a workflow that pushes/merges/deploys.
- AP-5: raising a job `timeout-minutes` instead of splitting/re-budgeting stages (headroom is for evidence steps, not for the stage itself).
- AP-6: retrying a semantic failure (assertion/race/panic) — registry-admitted transient failures only; first failure uploaded first.
- AP-7: editing the branch-protection SSoT in a different milestone/commit than the required-name change it accompanies.
- AP-8: "simplifying" `|| rc=$?` to `; rc=$?` (the `-e` exit-at-failure hazard documented in the release workflow).

## §H Cross-References

- spec.md §C REQ-RPC-001..012; acceptance.md §D AC-RPC-001..012 + §B evidence ledger EV-1..EV-9.
- decision-index.md Q1-Q6 (interview.yaml `decision_gate: on`; defaults applied per the Default rule, operator-overridable at Kickoff).
- `.moai/reports/t1536/progress.md` — card progress record (§E.1 audit-ready signal).
- Sibling SPECs: §A.2 above.
