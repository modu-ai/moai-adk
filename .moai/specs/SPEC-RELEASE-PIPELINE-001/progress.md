# progress.md — SPEC-RELEASE-PIPELINE-001

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
audit_ready: true
plan_complete_at: 2026-10-16T19:56:40Z
plan_artifacts: spec.md, plan.md, acceptance.md, decision-index.md (Tier M + decision gate on)
spec_id: SPEC-RELEASE-PIPELINE-001
card: t1536
baseline_head: 0bd1da4973d1d9539172f90587ca1ee84c8d9426

Plan-phase evidence: item-5 jq defect reproduced RED at the baseline head
(acceptance.md §B EV-3, exit 5) with fixed-form divergence (EV-4); parity
guard baseline green (EV-1); re-anchor table for all five card items in
plan.md §A.1 (item 1 and half of item 3 already landed on main).

Iter-1 repair (audit verdict FAIL 0.69 → this revision, spec v0.2.0): all
six blocking defects + MP-8 disposed — D1 both required-name SSoT surfaces
(gtmpl + required-checks.yml) named; D2 dry-run dispatch wiring added to
M5; D3 REQ-RPC-012 pins both sides' cancel policy; D4 REQ-RPC-009
redefined + M4 upload extension, AC-009 release-blocking; D5 REQ-RPC-001
fail-open narrowed to the release consumer; D6/MP-8 RED-now cells
EV-7/EV-8/EV-9 measured at c8ede48f90fbfe4c6873f300c91243f03928d254
(grep -c observed 0/0/0, exit 1 each).

Iter-2 repair (audit verdict FAIL 0.69, `.moai/reports/t1536/plan-audit-iter2.md`
→ this revision, spec v0.3.0): D7 (critical) — boolean→"1" conversion specified
at the MOAI_CLEANUP_DRY_RUN wiring point (`${{ inputs.dry_run && '1' || '0' }}`);
TestCleanupDryRunDispatchReachable asserts zero gh release delete under the
wired "true" (mock-gh); §D.2 carries the value contract. D8 — ref-independent
literal concurrency group name on both workflows + group-key-invariance
assertion in TestDrafterCleanupConcurrency; REQ-RPC-012 names ref-independence.
D9 — .github/required-checks.yml restated as the single required-name SSoT;
the gtmpl + template mirror reclassified shape-only PRESERVE (§D-11, M2 steps
1+4, AC-RPC-004). O5 — retry-wrapper anchors corrected to the seven measured
ci.yml sites (:222/:370/:422/:473/:525/:578/:635, verified this run).

Hold-and-split revision (iter-3 FAIL 0.69 at the Tier M ceiling →
operator-approved split → this revision, spec v0.4.0, 2026-10-22):

plan_status: audit-ready
audit_ready: true
plan_complete_at: 2026-10-22T12:26:32Z
plan_tree: bf0a0f35fc280e83c2f9fd76d0e2a5de39b7e8b7
scope: M1 (verify-and-preserve) + M2 + M4 + M5 (with D10); M3 + D11 moved to a follow-up card
requirements: 11 live (REQ-RPC-001..004, 007..013); REQ-RPC-005/006 [RETIRED]
acceptance_criteria: 11 live (9 release-blocking, 2 regression-guard); AC-RPC-005/006 [RETIRED]
authorized_audit_rounds: 1 (operator-approved, scope = this revision incl. D10)

Changes: card item 3 removed (REQ/AC-005/006, EV-2, E5, M3; identifiers retired
in place); REQ-RPC-013 / AC-RPC-013 added for D10 (EV-10 RED: `gh release list`
rejects `targetCommitish`, exit 1; EV-12 green-path: `gh release view` supplies
it); every RED-now cell re-measured at plan_tree (acceptance.md §B, document-level
pin); line-citation drift table in plan.md §A.2 (release-pr-multi-os.yml moved
by t1892 sharding; required-checks.yml gained `release_gate` by t1858; ci.yml,
codeql.yml, release-drafter*.yml unchanged). Decision-index: Q4 re-opened and
reversed to retained names (release_gate subset invariant), Q7 (fetch source)
and Q8 (event-class gating scope) added and default-applied. AC-RPC-004
reclassified regression-guard (no RED-now cell exists on the real tree).

## Binding run conditions

Copied from the PASS-WITH-DEBT verdict `.moai/reports/t1536/plan-audit-iter7.md` (audited b0f6783e434f, score 0.93, codex receipt rcpt-8e4e368d127d6d4389f9b43f); both sync verdict owners re-read these, and an undisposed condition caps the sync verdict at FAIL. Kickoff 2026-10-23: autonomous audit-decider approve (`moai factory decide t1536 --gate kickoff --choice approve --decider audit`, card v12), under the operator ceiling exception recorded in §G.

- debt: N1 dispose_in=run REQ-RPC-002 / AC-RPC-002 fail-open text lacks the "run not itself cancelled" qualifier stated in REQ-RPC-003
- debt: N3 dispose_in=run AC-RPC-003 cell (i) (real failure, marker skipped) pair relies on the landed same-name resolution; state it in (i)
- debt: N4 dispose_in=run acceptance.md §D.0 "verdict step without fail-open" probe cites AC-RPC-004 as a red source that no longer reads red

Binding run conditions disposition (manager-spec, 2026-10-23, SPEC v0.7.0, text-only; the debt lines above are kept as recorded):

- N1 disposed 2026-10-23 — spec.md §C REQ-RPC-002 (fail-open clause now "in a run that is not itself cancelled", with a pointer to REQ-RPC-003's cancelled-run rule) and acceptance.md §D AC-RPC-002 (green path "fail-open on unevaluated classification in a run that is not itself cancelled (a cancelled run is AC-RPC-003 cell (iv))").
- N3 disposed 2026-10-23 — acceptance.md §D AC-RPC-003 cell (i) (states the (real job `failure`/`cancelled`, marker `skipped`) pair relies on the landed same-name check-run resolution the `Test (ubuntu-latest)` / `test-skip-marker` pair already relies on; accepted pre-existing dependency, not introduced by this SPEC).
- N4 disposed 2026-10-23 — acceptance.md §D.0, probe "AC-RPC-003 vs verdict step without fail-open" (the "and by AC-RPC-004's reported-name assertion" citation removed; the probe reads red through AC-RPC-003 cell (iii) alone).
- Run-phase scope notes 1, 3, 4 (§E.2) reconciled in the same revision: REQ-RPC-004 / AC-RPC-004 `Test (ubuntu-latest)` exemption; REQ-RPC-008, plan M4 step 3, AC-RPC-007/008 at 15m/16m (n=4 green main runs); plan citations touched refreshed to HEAD `3b3fca7d0736879c23771b30db80c3a95c0b3b1c`.

## §E.2 Run-phase Evidence

Measured by manager-develop (cycle_type=tdd, serial) in worktree `.moai/worktrees/t1536`, branch `WT-10-06-codex`. Tree for every GREEN cell below: HEAD `1258076d11102e596dadb86a80aa657cbbf6174f` (M1 commit) unless a cell names another; plan-tree baseline `4c387f96bc`. Run clock `2026-10-22T20:27Z` UTC (2026-10-23 local). gh `2.92.0 (2026-04-28)`, jq present. Nothing was pushed (lane holds pushes).

Commits (plan order M2, M5, M4, M1): M2 `8a48e715f7` · M5 `4fc2d0a8d5` · M4 `81563e6f3f` · M1 `1258076d11`. The first commit carries the `draft -> in-progress` transition (spec.md `status:` + `updated:`; plan.md / acceptance.md `updated:` only).

### AC matrix (regression-guard rows are labelled; none is counted as a release gate)

| AC | Class | Verifying command (this run, HEAD 1258076d11) | Actual output | Status |
|---|---|---|---|---|
| AC-RPC-001 | regression-guard | `go test ./internal/template/ -run '^(TestTestInputFiltersSingleSource\|TestTestInputFiltersCoversGoTestInputs)$' -count=1 -v` | both `--- PASS`, `ok ... internal/template 0.231s` (swept 2 of 2); consumers: `grep -n "filters: .github/test-input-filters.yml"` -> ci.yml:85, codeql.yml:43, release-pr-multi-os.yml:68; fail-open wiring: release-pr-multi-os.yml:64 `continue-on-error: true`, :120 `!= 'false'` | PASS (guard holds) |
| AC-RPC-002 | release-blocking | `go test ./internal/template/ -run '^TestCICostJobsInputGated$' -count=1 -v` | RED before M2 (see E8); after: `--- PASS: TestCICostJobsInputGated`, 6 negative controls observed red | PASS |
| AC-RPC-003 | release-blocking | `... -run '^TestCIGatedFamilyVerdictSemantics$'` | RED before M2: `Lint: (iii) detect=failure mode=0: published [success], want the real job and the marker both failure`; after: PASS, 5 negative controls observed red | PASS |
| AC-RPC-004 | regression-guard | `... -run '^TestCIRequiredNamesSatisfied$'`; `go test ./internal/config/ -run '^TestRequiredChecks_ReleaseGateSubsetOfMain$' -v` | green BEFORE and AFTER M2; both negative controls (a) lint gated + no Lint marker -> `pull_request to main x docs-only: required name "Lint" is not reported`, (b) release_gate + `Release PR Multi-OS Gate` -> `push to main x docs-only: ... not reported` observed red on the pre-M2 tree; config test `--- PASS` | PASS (guard holds; see scope note 1) |
| AC-RPC-007 | release-blocking | `... -run '^TestStageCeilingBelowJobBudget$'` | RED before M4 (browser half); after: PASS; release-half controls T=38 (`38 + 5 + 1 = 44 exceeds the job budget 40`) and upload without `if: always()` observed red | PASS |
| AC-RPC-008 | release-blocking | `... -run '^TestBrowserPhaseSplit$'` | RED before M4; after: PASS, 6 controls red | PASS |
| AC-RPC-009 | release-blocking | `... -run '^TestRetryPolicyPreservesFirstFailure$'` | RED before M4 (7 jobs without the attempt1 upload); after: PASS, 6 controls red | PASS |
| AC-RPC-010 | release-blocking | `... -run '^TestCleanupFilterSelectsStaleDrafts$'` | RED before M5 (no extractable program); EV-3/EV-4 divergence re-observed; after: PASS (selects exactly `stale-draft`), 5 controls red | PASS |
| AC-RPC-011 | release-blocking | `... -run '^TestCleanupRequeriesBeforeDelete$'` | RED before M5; after: PASS (4 scenarios: unchanged deleted; published / retargeted / refreshed skipped and logged), 2 controls red | PASS |
| AC-RPC-012 | release-blocking | `... -run '^TestDrafterCleanupConcurrency$'` | RED before M5 (no cleanup group; ref-keyed drafter group; cancel true); after: PASS, 5 controls red | PASS |
| AC-RPC-013 | release-blocking | `... -run '^TestCleanupFetchFieldsSupported$'` | RED before M5 (`gh release list requests --json field "targetCommitish"`); after: PASS, 4 controls red | PASS |

Final GREEN sweep, one invocation: `go test ./internal/template/ -run '^(TestCICostJobsInputGated|TestCIGatedFamilyVerdictSemantics|TestCIRequiredNamesSatisfied|TestCleanupFetchFieldsSupported|TestCleanupFilterSelectsStaleDrafts|TestCleanupRequeriesBeforeDelete|TestDrafterCleanupConcurrency|TestCleanupDryRunDispatchReachable|TestStageCeilingBelowJobBudget|TestBrowserPhaseSplit|TestRetryPolicyPreservesFirstFailure|TestReleaseVerifyShardParity|TestTestInputFiltersSingleSource|TestTestInputFiltersCoversGoTestInputs)$' -count=1 -v` -> 14 `--- PASS` lines (11 new guards incl. the dry-run reachability guard, `TestReleaseVerifyShardParity` swept and green, the 2 EV-1 guards), `ok ... 18.620s`.

### RED-now cells re-observed at plan tree 4c387f96bc (before any change)

- EV-14 `grep -n "needs.detect.result" .github/workflows/ci.yml` -> empty, exit 1. EV-5 `grep -nE "^  [a-z][a-z0-9-]*:|^    if:|^    needs:|^    continue-on-error:" .github/workflows/ci.yml` -> `lint`/`build`/`constitution-check`/`test-browser` with no `needs`/`if` (the `test` family has them).
- EV-10 `gh release list --repo modu-ai/moai-adk --json tagName,isDraft,createdAt,targetCommitish -L 1` -> `Unknown JSON field: "targetCommitish"`, exit 1 (gh 2.92.0); EV-12 `gh release view v3.1.3 --repo modu-ai/moai-adk --json tagName,isDraft,createdAt,targetCommitish` -> `{"createdAt":"2026-08-21T19:28:08Z","isDraft":true,"tagName":"v3.1.3","targetCommitish":"main"}`, exit 0; `gh release view ... --json bogusField` lists the view field set (apiUrl assets author body createdAt databaseId id isDraft isImmutable isPrerelease name publishedAt tagName tarballUrl targetCommitish uploadUrl url zipballUrl) — pinned in the guard.
- EV-3 `jq -r --arg threshold_days 30 'now | floor as $now | ...' internal/template/testdata/release-drafter-cleanup.json` -> `jq: error (at ...:6): Cannot iterate over number (1792699450....)`, exit 5; EV-4 `(now|floor) as $now` form -> `stale-draft`, exit 0 (the committed fixture stands in for the gitignored plan-phase copy).
- EV-7 `grep -c "gh release view" .github/workflows/release-drafter-cleanup.yml` -> `0` exit 1; EV-8 `grep -c "concurrency" ...` -> `0` exit 1; EV-9 `grep -c "attempt1" .github/workflows/ci.yml` -> `0` exit 1.
- EV-6/EV-11: `grep -n "timeout-minutes" .github/workflows/release-pr-multi-os.yml` -> :41 :121 (40) :367 plus comments, no race-step ceiling (the compliant Q9 state). ci.yml line numbers moved after M2 (browser job now at :1110, green step :1187).

### E8 — verbatim RED before GREEN (the first violation lines per guard, copied from the run's test output; the full RED outputs were captured in the run session's scratch files, which are machine-local and not cited)

```
M2  CI_COST_GATE_VIOLATION: lint: runs on a docs-only pull_request (no input class matched)
M2  CI_COST_GATE_VIOLATION: lint: does not need detect, so it cannot read the input classification
M2  CI_VERDICT_SEMANTICS_VIOLATION: Lint: (iii) detect=failure mode=0: published [success], want the real job and the marker both failure
M2  --- FAIL: TestCICostJobsInputGated / --- FAIL: TestCIGatedFamilyVerdictSemantics / --- PASS: TestCIRequiredNamesSatisfied (regression-guard green; controls (a)(b) red)
M5  CLEANUP_FETCH_VIOLATION: gh release list requests --json field "targetCommitish", which that subcommand does not support (gh 2.92.0: "Unknown JSON field", exit 1)
M5  CLEANUP_REQUERY_VIOLATION: the delete step does not re-query (gh release view) before gh release delete (view at -1, delete at 225)
M5  CLEANUP_CONCURRENCY_VIOLATION: ... release-drafter-cleanup.yml declares no concurrency group
M5  CLEANUP_CONCURRENCY_VIOLATION: ... release-drafter.yml: the group key varies with github.ref ...
M5  CLEANUP_DRYRUN_VIOLATION: workflow_dispatch declares no dry_run input, so the dry-run path is unreachable by dispatch
M5  CLEANUP_FILTER_VIOLATION: the workflow carries no stale-draft jq program (job env STALE_FILTER)
M4  STAGE_BUDGET_VIOLATION: "Green phase — gated driver tests (in-process surface)" has no step-level timeout-minutes
M4  STAGE_BUDGET_VIOLATION: no always() evidence upload follows the mutation phase, so a failing phase leaves no trail
M4  RETRY_EVIDENCE_VIOLATION: job test-race-cli-1 runs the retry wrapper but no upload step carries .moai/logs/ci-flaky-retry.attempt1.json   (x7 jobs)
M4  release half: green before M4; controls: race-step ceiling 38 + pre-step 5 + evidence 1 = 44 exceeds the job budget 40 / "Upload test event stream" is not an `if: always()` step
```

### E2 / E3 / E4 / E5

- E2 `go build ./...` -> exit 0; `GOOS=windows GOARCH=amd64 go build ./...` -> exit 0; both re-run at HEAD 1258076d11; `GOOS=windows GOARCH=amd64 go vet ./internal/template/` -> exit 0 (run at the M5 commit, the test files' windows compile check).
- E3 `go test -cover ./internal/template/ -count=1` -> `coverage: 82.8% of statements`, package FAIL only on three tests that also fail at the plan tree `4c387f96bc` (`TestBindingLedgerIntegrity`, `TestCatalogHashCoversSkillSubfiles`, `TestManifestHashFormat`: moai-lane-watchdog catalog hash, ledger rows ALB-0139/0148). No non-test Go file was touched, so there is no modified-file coverage to hold at 85%.
- E4 `grep -rn 'AskUserQuestion'` over the four new Go files, comment lines excluded -> no match.
- E5 `golangci-lint run --timeout=10m ./internal/template/...` -> `0 issues.`; `moai spec lint .moai/specs/SPEC-RELEASE-PIPELINE-001` -> 0 errors, 1 warning (`VacuousTestAssertion` at acceptance.md:52, an EV-6 cell in SPEC body text this agent may not edit).
- `actionlint` over ci.yml, release-pr-multi-os.yml, release-drafter-cleanup.yml, release-drafter.yml -> exit 0. `git diff --name-only 4c387f96bc..HEAD` lists only: test-input-filters.yml, the four workflows above, the three SPEC artifacts (frontmatter), and the four Go test files plus the fixture; codeql.yml and required-checks.yml are untouched.

### Scope notes for manager-spec disposition (SPEC text; not edited here)

1. AC-RPC-004's "classification failure" outcome cannot hold for `Test (ubuntu-latest)`: when `detect` fails, both `test` and `test-skip-marker` are skipped (no failure arm; PRESERVE item 4), so the name is unreported today and stays so. The guard names that name as an explicit pre-existing fail-closed-by-absence exemption (`rpcFailClosedByAbsence`) and asserts every other name in every outcome.
2. Binding conditions N1 (REQ/AC-RPC-002 lacks the not-cancelled qualifier), N3 (AC-RPC-003 cell (i) same-name resolution unstated), N4 (stale AC-RPC-004 citation in the "verdict step without fail-open" probe) remain open SPEC-text debts. The implementation is consistent with REQ-RPC-003: real jobs run under `!cancelled()`, the marker runs and fails on a failed classification, the verdict step is `if: always()`.
3. Plan line citations shifted after M2 (ci.yml consumer line :79 -> :85; browser job :950 -> :1110).
4. The browser green phase's Go `-timeout` moved 17m -> 15m (step ceiling 16) because REQ-RPC-007/008's arithmetic (setup 1 + green + mutation + evidence 1 <= 20 with a mutation phase measured up to 1m16s) cannot hold a 17-minute green phase; the measured green maximum is 13m43s (n=4).

### Sync-audit iter1 repair (F1-F7 dispositions; audit file `.moai/reports/t1536/sync-audit-iter1.md`, audited sha f3c09a12e7)

All commands run in this worktree; repair commits `687bbceb33` (F1) and `7468f1601c` (F7) on top of f3c09a12e7.

- F1 RED (EXPECTED_RED), `go test ./internal/template/ -count=1 -run TestWorkflowRunScriptsNoUntrustedInterpolation` on the unrepaired workflows -> `--- FAIL`, `swept 77 run steps across 4 workflows`, seven `RUN_INTERPOLATION_VIOLATION` lines, among them `release-drafter-cleanup.yml: job "cleanup" step "Show results" interpolates ${{ steps.drafts.outputs.stale_tags }} into its run script` (reported twice, one per occurrence), the `stale_count` / `dry_run` interpolations of "Update cleanup log" and "Show results", and `release-pr-multi-os.yml: job "release-pr-gate" step "Verify all OS legs passed" interpolates ${{ needs.detect-release.outputs.go_code }}`.
- F1 GREEN, same command after the repair (the three steps read the values from the step `env:` and print with `printf '%s\n'`) -> `--- PASS: TestWorkflowRunScriptsNoUntrustedInterpolation (0.03s)` with seven mutation controls red as intended (one injects `steps.drafts.outputs.stale_tags` into "Show results" itself) and the allowed-enum control green (`needs.detect.result`, `matrix.goos`, `github.event_name` stay accepted).
- F2 disposition: NO permission change. `gh run list --repo modu-ai/moai-adk --workflow ci.yml --event pull_request --status success --limit 3 --json databaseId,headSha,conclusion` -> run `39848631205` (head `94b72685717466617fc93ef73c6efcf6422188c5`, branch `WT-fail-safe-serial`, event `pull_request`, created 2026-10-22T20:24:50Z, conclusion `success`). `gh run view 39848631205 --repo modu-ai/moai-adk --json jobs --jq '.jobs[] | [.name,.conclusion] | @tsv'` -> `Detect (paths-filter)	success`, `Lint	success`, `Build (linux/amd64)	success`. That run's ci.yml has top-level `permissions: contents: read` (line 38-39), no job-level override, and `dorny/paths-filter@ceb8a2b8f2d89434be7ff52d3de7ec3738c5cc9d # v4.0.3` (read through `gh api repos/modu-ai/moai-adk/contents/.github/workflows/ci.yml?ref=94b72685717466617fc93ef73c6efcf6422188c5`); this tree's ci.yml carries the same permission block and the same paths-filter pin. Scope of the observation: a same-repository pull_request run; a fork pull_request run was not observed.
- F7 RED (EXPECTED_RED): eight added mutants (`always() && false` / `if: false` on the release compress and upload steps, the browser phases and browser evidence upload, the split-guard phase and upload, the retry attempt1 upload) printed `mutant ... passes the guard` against the previous guards (`TestStageCeilingBelowJobBudget`, `TestBrowserPhaseSplit`, `TestRetryPolicyPreservesFirstFailure`). GREEN: the guards now compare the normalized condition to exactly `always()` and require the two browser phase steps to carry no `if:`; the same three tests pass and all eight mutants are red.
- F4 (input-class guard stays green when all 17 classification paths are replaced by non-matching paths): NOT repaired. The cheap strengthening is a root-existence check over the filter file, and it is red on the real tree: `ls -d CLAUDE.md` -> `No such file or directory` and `git ls-files CLAUDE.md | wc -l` -> `0`, yet `.github/test-input-filters.yml` lists `CLAUDE.md` as a `go_code` root. Whether that root is stale or forward-looking is a SPEC-covered filter-content decision. Remaining debt.
- F5 (required-name guard ignores `paths:` filters on external workflows) and F6 (`exit 1` matched textually, `if false; then exit 1; fi` passes): NOT repaired. Both need workflow-semantics evaluation rather than a text match; remaining debt, one follow-up card on guard strength.
- F3 (discretionary concurrency queue) untouched.
- Gate set at the repair tree, one plain invocation (the ambient environment was NOT scrubbed; no `unset` was applied, a Gap against the card's env-scrubbed instruction): `go test ./internal/template/ -count=1 -run 'TestCICostJobsInputGated|TestCIGatedFamilyVerdictSemantics|TestCIRequiredNamesSatisfied|TestStageCeilingBelowJobBudget|TestBrowserPhaseSplit|TestRetryPolicyPreservesFirstFailure|TestCleanup|TestDrafterCleanupConcurrency|Parity|TestWorkflowRunScriptsNoUntrustedInterpolation'` -> `ok  	github.com/modu-ai/moai-adk/internal/template	36.509s`; `golangci-lint run ./internal/template/...` -> `0 issues.`; `actionlint` (repo-wide) -> exit 0.

### Sync-audit iter2 repair (G1; audit file `.moai/reports/t1536/sync-audit-iter2.md`, audited sha 9f4e3078c4)

Repair commit `98599969ca` on top of 9f4e3078c4; all commands run in this worktree with `unset GITHUB_ACTIONS CI MOAI_TEST_MODE` applied in the same invocation.

- G1 (the run-script interpolation guard was a denylist): turned into an allowlist. Each `${{ ... }}` in a run body is matched with `(?s)`, whitespace-collapsed and lower-cased, and must equal `needs.<id>.result`, `matrix.<key>`, `github.event_name`, `github.run_id|run_number|run_attempt|sha`, `runner.os|arch` or `job.status`; anything else is refused under `RUN_INTERPOLATION_VIOLATION`. Allowlist sizing evidence: a scratch dump of every expression in the four workflows' run bodies printed only `needs.detect.result`, `needs.detect-release.result`, `needs.full-matrix-test.result`, `matrix.goos`, `matrix.goarch`; `secrets.*` occurs only in `env:`/`with:` (grep), so it is not allowed in a run body. The scratch test was deleted before the commit.
- G1 RED, run on the OLD predicate with the 12 new controls added (`go test ./internal/template/ -count=1 -run TestWorkflowRunScriptsNoUntrustedInterpolation -v`): `--- FAIL`, and 11 of the 12 printed `mutant interpolating ${{ <expr> }} passes the guard` for `steps['drafts'].outputs.stale_tags`, `STEPS.drafts.outputs.stale_tags`, `toJSON(github.event)`, `env.STALE_TAGS`, `fromJSON(toJSON(steps)).drafts.outputs.stale_tags`, the multi-line `format('{0}',\n  github.event.head_commit.message)`, `github.ref_name`, `github.actor`, `secrets.CODECOV_TOKEN`, `join(matrix.goos, ',')`, `needs['detect-release'].outputs.go_code`. The twelfth control (`\n  github.event.pull_request.title\n`, multi-line but a bare payload field) was already caught by the old `\s*(.+?)\s*` pattern and stayed PASS; it is kept as a regression control, not as a bypass.
- G1 GREEN, same command on the allowlist: `--- PASS: TestWorkflowRunScriptsNoUntrustedInterpolation (0.04s)`, `swept 77 run steps across 4 workflows`, 19 refusal controls observed red plus `control/every allowlisted form stays allowed` green. No workflow edit was needed: the real run bodies already use only allowlisted forms.
- Gate set at the repair tree: `go test ./internal/template/ -count=1 -run 'TestWorkflowRunScriptsNoUntrustedInterpolation|TestCICostJobsInputGated|TestCIGatedFamilyVerdictSemantics|TestCIRequiredNamesSatisfied|TestStageCeilingBelowJobBudget|TestBrowserPhaseSplit|TestRetryPolicyPreservesFirstFailure|TestCleanup|TestDrafterCleanupConcurrency'` -> `ok  	github.com/modu-ai/moai-adk/internal/template	11.875s`; `golangci-lint run ./internal/template/...` -> `0 issues.`; `actionlint` on ci.yml, release-drafter-cleanup.yml, release-drafter.yml, release-pr-multi-os.yml -> exit 0.
- Residual: the allowlist trusts `matrix.<key>` because every matrix in the four workflows is a literal; a future matrix built with `fromJSON` from a step or job output would make `matrix.*` attacker-influenceable and the guard would not notice. Expressions in other positions than `run:` (`env:`, `with:`, `if:`, `shell:`) are outside this guard.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-10-22T20:27:18Z
run_commit_sha: pending-backfill   # M2 8a48e715f7, M5 4fc2d0a8d5, M4 81563e6f3f, M1 1258076d11; the evidence commit cannot cite itself
run_status: complete-pending-audit
ac_pass_count: 11                  # 9 release-blocking + 2 regression-guard
ac_fail_count: 0
preserve_list_post_run_count: 12   # plan.md section D items; spot-verified: filter + parity (1), release-pr-gate untouched (2), release/* exclusions untouched (3), test-skip-marker untouched (4), retry sites (5, TestRetryPolicy... 7 sites), permissions blocks unchanged (6), race evidence chain (7, TestStageCeiling... release half), browser invariants (8, TestBrowserPhaseSplit), release.yml untouched (9), item 10 untouched, required-checks.yml and gtmpl untouched (11), shard parity (12, TestReleaseVerifyShardParity PASS)
l44_pre_commit_fetch: not-run    # no fetch performed; the lane holds pushes and reads no remote state
l44_post_push_fetch: not-applicable  # no push
new_warnings_or_lints_introduced: 0  # golangci-lint 0 issues; spec lint warning at acceptance.md:52 is pre-existing SPEC text
cross_platform_build:
  linux_amd64: exit 0
  windows_amd64: exit 0
total_run_phase_files: 13
m1_to_mN_commit_strategy: per-milestone local commits in plan order M2, M5, M4, M1; evidence commit after; no push
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-10-23
sync_commit_sha: pending-backfill   # the sync commit cannot cite its own hash; backfilled in a following commit
sync_status: complete-pending-audit
tier: M
ac_source: .moai/specs/SPEC-RELEASE-PIPELINE-001/acceptance.md   # B12 counter: stdout 11, stderr live=11 excluded=2 ambiguous=0
b12_self_test_a: pass   # grep -c 'SPEC-RELEASE-PIPELINE-001' CHANGELOG.md was 0 before emission
b12_self_test_b: pass   # CHANGELOG states 11 acceptance criteria (9 release-blocking + 2 regression-guard) = counter output
b12_self_test_c: pass   # every file path claimed in the CHANGELOG entry verified by ls
changelog_entry_position: "[Unreleased] > Added, first bullet"
frontmatter_status_transitions:
  spec_md: "in-progress -> implemented -> completed (single sync commit; updated: already 2026-10-23, unchanged)"
  plan_md: "no status field; updated: already 2026-10-23, unchanged"
  acceptance_md: "no status field; updated: already 2026-10-23, unchanged"
canary_compliance_check: not-applicable   # this SPEC defines no forward-looking policy that its own sync tests
mx_review: no tag added   # no non-test Go source changed; no @MX rule fires (test files and CI workflow YAML only)
docs_impact: CHANGELOG.md only; no tracked codemap, README, or docs-site page names a changed workflow or guard file
known_warning: "moai spec lint: 1 pre-existing VacuousTestAssertion at acceptance.md:52 (verbatim EV-6 evidence, SPEC body text; left unchanged)"
```

## §F Phase 4 Mode Selection

- Input parameters: tier M; scope ~10-14 files (`.github/workflows/*.yml`, `.github/required-checks.yml` consumers, Go guard tests under `internal/template/`, `internal/template/testdata/`); domains 2 (CI workflows, Go guard tests); file mix YAML + Go + jq; concurrency benefit LOW (coding-heavy, M2 -> M5 -> M4 -> M1 ordered, M4 depends on M2's job names).
- Mode evaluation: direct not selected (semantic change); serial selected; fanout not selected (coding-heavy, <3 domains); sweep not selected (not mechanical-uniform); agent-team not requested.
- Decision: serial
- Justification: coding-heavy work with ordered milestones stays sequential per Anthropic's coding-task parallelism caveat; one manager-develop (cycle_type=tdd, effort high for Class C) carries M2, M5, M4, M1 in plan order, one writer in this tree.
- Kickoff: autonomous audit-decider approve 2026-10-23 (card v12) on the PASS-WITH-DEBT verdict plan-audit-iter7.md, under the operator ceiling exception recorded in §G; Binding run conditions above.

## §G Override and Refusal Record

- 2026-10-22T18:41:00Z SPEC-RELEASE-PIPELINE-001 ceiling-refusal outcome=hold reasons="plan-audit ceiling reached (round count 6 >= tier ceiling 2); the verdict matches no admitting arm and holds, entry blocked (REQ-ACE-006) — release path: the split/new-SPEC route of REQ-ACE-005 or an operator decision recorded in progress.md §G" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter6.md
- 2026-10-23 SPEC-RELEASE-PIPELINE-001 operator-decision (relayed by the leader session 3936623d to lane-9 on 2026-10-23, recorded verbatim at the leader's request; the lane did not observe the operator directly): "운영자 결정 2026-10-23: plan-audit iter6 PASS-WITH-DEBT 0.93(rcpt-78c759130e822df7b345659e)을 수용, REQ-ACE-006 천장(6회>=2) 예외 승인, 리더 세션 3936623d 경유."
- 2026-10-22T19:20:15Z SPEC-RELEASE-PIPELINE-001 ceiling-refusal outcome=hold reasons="plan-audit ceiling reached (round count 6 >= tier ceiling 2); the verdict matches no admitting arm and holds, entry blocked (REQ-ACE-006) — release path: the split/new-SPEC route of REQ-ACE-005 or an operator decision recorded in progress.md §G" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter6.md
- 2026-10-23 SPEC-RELEASE-PIPELINE-001 round-7 note (instruction relayed by the leader session, which relayed the operator decision above): 7th round = hash-binding re-audit to enforce the operator exception, scope limited to the relabel delta 71b227b088..8e097c9d84 with a regression glance only; the verdict must record the engine plan-artifact hash.
- 2026-10-22T19:32:29Z SPEC-RELEASE-PIPELINE-001 ceiling-refusal outcome=hold reasons="plan-audit ceiling reached (round count 7 >= tier ceiling 2); the verdict matches no admitting arm and holds, entry blocked (REQ-ACE-006) — release path: the split/new-SPEC route of REQ-ACE-005 or an operator decision recorded in progress.md §G" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter6.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter7.md
- 2026-10-22T19:35:50Z SPEC-RELEASE-PIPELINE-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 7 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter6.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter7.md
- 2026-10-22T19:36:02Z SPEC-RELEASE-PIPELINE-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 7 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter6.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1536/.moai/reports/t1536/plan-audit-iter7.md
