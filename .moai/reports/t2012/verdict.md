# t2012 verdict — ADK own CI: bot-merge push CI, release matrix default, PR race (Class B, lane-3)

Policy applied: `verification-claim-integrity.md` §1.1 surface 3, §2, §3.1; `verification-completeness.md` §1.1, §2. Tree: branch WT-ci-push-ci on origin/main 5c3f0ebf27 (fast-forwarded from cb30e05b3e, 14 behind). Code commit: 4bd6a28e15.

## Claim 1 — bot-merged commits get no ci.yml push run (REPRODUCED, no patch)
Evidence (read-only):
```
gh api repos/modu-ai/moai-adk/pulls/{1942,1947,1922,1940} --jq .merged_by.login
  1942 github-actions[bot]   1947 github-actions[bot]   1922 GoosLab   1940 GoosLab
gh run list --workflow ci.yml --limit 100 ... headSha startswith 5c3f0ebf27|ee5763349e  -> count 0
gh run list --workflow ci.yml --branch main --limit 12: 9aab38f500 push success, f8415a5d8e push success (human merges)
```
Cause (code): `.github/workflows/auto-merge.yml:643` merges with `secrets.GITHUB_TOKEN`; docs.github.com triggering-a-workflow: "events triggered by the `GITHUB_TOKEN` will not create a new workflow run" except `workflow_dispatch`/`repository_dispatch`. A dispatch needs `actions: write`, but `auto-merge.yml:19-24` pins `actions: read` and `auto_merge_codex_test.go:1686`, `:3838` pin the permission set. Not patched: token-scope widening and the `workflow_run` alternative are operator decisions (leader). Docs do not state whether `workflow_run` fires for GITHUB_TOKEN-started upstream runs, nor which commit its check runs attach to (progress.md, Item 1 (b)).

## Claim 2 — release.sh runs the 3-OS matrix gate by default (DONE, local commit 4bd6a28e15)
Change: `REQUIRE_MATRIX_RUN` default true; `--skip-matrix-run` is the one way past it and the summary prints `Matrix gate: OVERRIDDEN (--skip-matrix-run)`; a passed gate prints `verified (run <id>)`; `--require-matrix-run` still accepted; both flags together refused. Gate logic (3 OS x 6 shards, dispatch, exact SHA) unchanged.

RED first (before the script change), `go test ./internal/template -run 'TestReleaseScriptMatrixGate$' -count=1`: `--- FAIL` for all 15 `switch_skip/*` rows (unknown flag `--skip-matrix-run`) and all 15 `switch_default/*` rows (no flag: the matrix was never consulted, so refusal scenarios exited 0 and the green scenario lacked `Matrix gate: run 101 …`); `FAIL github.com/modu-ai/moai-adk/internal/template 58.445s`.

Side RED (default flipped, other-gate tests unchanged): TestReleaseScriptCIGateSSoTFailsClosed, TestReleaseScriptCIGate, TestReleaseScriptCIGateWithoutGh, TestReleaseScriptTagsTheVerifiedCommit, TestReleaseScriptCharacterization, TestReleaseScriptHeadAdmission, TestReleaseScriptRcChangelogSkip failed — each ran release.sh with no matrix stub. They now pass the recorded override.

GREEN, `go test ./internal/template -run 'TestReleaseScript|TestReleaseVerifyShardParity|TestTestInputFilters|TestCICostJobsInputGated' -count=1 -v`:
```
--- PASS: TestCICostJobsInputGated  TestReleaseScriptCIGateSSoTFailsClosed  TestReleaseScriptRealSSoTDemandsEveryPinnedCheck
    TestReleaseScriptCIGate  TestReleaseScriptCIGateWithoutGh  TestReleaseScriptTagsTheVerifiedCommit
    TestReleaseScriptCharacterization  TestReleaseScriptHeadAdmission  TestReleaseScriptRcChangelogSkip
    TestReleaseScriptMatrixGate (39.44s)  TestReleaseScriptMatrixGateLargeJobList  TestReleaseVerifyShardParity
    TestTestInputFiltersSingleSource  TestTestInputFiltersCoversGoTestInputs
ok  github.com/modu-ai/moai-adk/internal/template  100.011s
```
Scoped gate: `gofmt -l internal/template` empty; `go vet ./internal/template/` ok; `golangci-lint run ./internal/template/` → `0 issues.`; `bash -n scripts/release.sh` ok. Swept set non-empty (`-v` rows above).

## Claim 3 — PR race blast radius (MEASURED only, no change)
Run 41283886701 (pull_request, head 32e79e33ec): six Race Test legs 12m11, 11m13, 15m07, 8m49, 18m35, 12m55; runner time ≈ 78m50s, wall ≈ 18m35s. Race jobs are advisory/non-required (ci.yml:269-284). With Claim 1 unresolved, narrowing PR race would leave race mostly to the release gate.

## Claim 4 — required checks: live vs file (RECORDED, nothing changed)
Live main protection = Lint, Build (linux/amd64), Analyze (Go) (go); `.github/required-checks.yml` branches.main = 10 (adds Test (ubuntu-latest), four more Builds, Release PR Multi-OS Gate, Guard Bundle). The file is the declared SSoT rendered by `branch_protection.go` and read by ciwatch, candidate verdicts and release.sh; GitHub enforces only the live 3. Who reduced live protection and when: not observed.

## Baseline attribution
Commands above ran in this run against tree 5c3f0ebf27 + commit 4bd6a28e15 (local, unpushed). Tool provenance: `go test`/`go vet`/`golangci-lint` are the installed toolchain run over this tree (the Go tests exec this tree's `scripts/release.sh`); no moai binary measurement is cited.

## Gaps
- Claim 1: GITHUB_TOKEN no-push-trigger is documented behaviour, not re-measured; the `workflow_run` check-run attachment is unmeasured and undocumented.
- shellcheck not installed (`command not found`); release.sh only `bash -n`.
- Not run: the full `./internal/template/...` suite and any CI (local commit only, no push).
- Refusals by the worktree-isolation guard: one compound `for … gh api --jq` loop and one `sed -i` call did not run; redone as plain commands / the Edit tool.
- Customer-template impact: grep over the 4 `.github` template files and 46 hooks found no mirror/reference of release.sh, auto-merge.yml, ci.yml or required-checks.yml; the `branch-protection.json.gtmpl` vs ADK 3-vs-10 comparison was not done.

## Residual-risk
- Every release now needs a dispatched green multi-OS run on the exact SHA before tagging; anyone releasing without it must pass `--skip-matrix-run` (recorded). Release docs/harness text calling `release.sh` without the flag were not found (`grep require-matrix-run` over .claude, .moai/docs, docs-site/ko empty) but external callers are unobserved.
- Claim 1 stays open until the operator picks (a), (b) or acceptance; until then bot-merged main commits carry no push CI.
