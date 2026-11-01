# t2012 verdict — ADK own CI: bot-merge push CI, release matrix gate, PR race (Class B, lane-3)

Policy applied: `verification-claim-integrity.md` §1.1 surface 3, §2, §3.1; `verification-completeness.md` §1.1, §2. Tree: branch WT-ci-push-ci on origin/main 5c3f0ebf27 (fast-forwarded from cb30e05b3e, 14 behind). Commits: 4bd6a28e15 (first draft, superseded by the leader's correction, kept in history), then the forward repair commit named in the card report.

## Claim 1 — bot-merged commits get no ci.yml push run (REPRODUCED, no patch)
Evidence (read-only):
```
gh api repos/modu-ai/moai-adk/pulls/{1942,1947,1922,1940} --jq .merged_by.login
  1942 github-actions[bot]   1947 github-actions[bot]   1922 GoosLab   1940 GoosLab
gh run list --workflow ci.yml --limit 100 ... headSha startswith 5c3f0ebf27|ee5763349e  -> count 0
gh run list --workflow ci.yml --branch main --limit 12: 9aab38f500 push success, f8415a5d8e push success (human merges)
```
Cause (code): `.github/workflows/auto-merge.yml:643` merges with `secrets.GITHUB_TOKEN`; docs.github.com triggering-a-workflow: "events triggered by the `GITHUB_TOKEN` will not create a new workflow run" except `workflow_dispatch`/`repository_dispatch`. A dispatch needs `actions: write`, but `auto-merge.yml:19-24` pins `actions: read` and `auto_merge_codex_test.go:1686`, `:3838` pin the permission set. Not patched. Leader decision 11-01: keep current behaviour (no `actions: write`, `workflow_run` unmeasured, rely on the release-time refusal). The docs do not state whether `workflow_run` fires for GITHUB_TOKEN-started upstream runs, nor which commit its check runs attach to (progress.md, Item 1 (b)).

## Claim 2 — release.sh always runs the 3-OS matrix gate (forward repair; supersedes 4bd6a28e15)
Leader correction 11-01: the first draft added an explicit off flag; that conflicts with the operator's "release full verification is mandatory" instruction and is withdrawn (forward commit, no amend). Now `scripts/release.sh` Validation 9b runs unconditionally: no flag, no environment variable, no test hook in the script. `--require-matrix-run` is accepted and does nothing. The diff against origin/main is mostly de-indentation of the gate block (a whitespace-ignoring diff shows the real change); gate logic (3 OS x 6 shards, dispatch, exact SHA, every leg success) is unchanged. The gh-missing message now reads `Matrix gate: the gh CLI is needed …`. `--skip-ci-check` is NOT changed (see Gaps).

RED first (new negative tests run against the 4bd6a28e15 script), `go test ./internal/template -run 'TestReleaseScriptRetiredBypassIsAnUnknownFlag|TestReleaseScriptRealTagPathRefusesWithoutAGreenMatrix|TestReleaseScriptNoSwitchTurnsTheMatrixOff' -count=1 -v` → `--- FAIL: TestReleaseScriptRetiredBypassIsAnUnknownFlag` (`/dry_run`, `/real_tag_path`: the retired flag was still accepted). The real-tag and no-switch tests passed there because that script already defaulted on: they are guards, not red-then-green.

Mutant for the guards (temporary copy, restored byte-identical, checked with `cmp`): the gate wrapped in `if [[ "${REQUIRE_MATRIX_RUN:-true}" != false && "$SKIP_CI_CHECK" != true && "$HOTFIX" != true ]]` → `--- FAIL` for `env_require_matrix_run_false`, `skip_ci_check_does_not_skip_the_matrix`, `hotfix_does_not_skip_the_matrix`. `env_matrix_gate_off_spellings` and `require_matrix_run_is_a_noop_flag` passed under that mutant (it honoured neither spelling), so they pin the contract but were not shown red.

New tests (`internal/template/release_matrix_nobypass_test.go`): the retired flag (spelled by concatenation in the test, so that string is absent from the diff) is `Unknown flag`, exit 1, no local and no remote tag, on the dry-run and on the real tag path (no `--dry-run`, push blocked by a shim); on the real tag path no run / green run on another SHA / red run on the target SHA each exit 1 with no tag, and a green-matrix control shows the harness does reach tag creation; with no matrix run, `REQUIRE_MATRIX_RUN=false` in the environment, `--skip-ci-check`, `--hotfix` and `--require-matrix-run` all still end at `Matrix gate: … no run of release-pr-multi-os.yml found for …`, exit 1. The other-gate tests bypass nothing: the gh double supplies a green dispatched matrix run on HEAD (`rlsGreenMatrix`); `TestReleaseScriptCIGateWithoutGh/override` now expects exit 1 and `Matrix gate: the gh CLI is needed` (skipping the CI gate does not skip the matrix); `TestReleaseScriptCharacterization` now requires a passing row to have read the matrix run.

GREEN, `go test ./internal/template -run 'TestReleaseScript|TestReleaseVerifyShardParity|TestTestInputFilters|TestCICostJobsInputGated' -count=1 -v -timeout 540s` → exit 0, `ok  github.com/modu-ai/moai-adk/internal/template  93.686s`; 17 `--- PASS` rows, 0 `--- FAIL` (RetiredBypass, RealTagPathRefuses, NoSwitchTurnsTheMatrixOff, MatrixGate, MatrixGateLargeJobList, Characterization, CIGate, CIGateSSoTFailsClosed, RealSSoTDemandsEveryPinnedCheck, CIGateWithoutGh, TagsTheVerifiedCommit, HeadAdmission, RcChangelogSkip, VerifyShardParity, TestInputFilters x2, CICostJobsInputGated).
Scoped gate: `gofmt -l internal/template` empty (exit 0); `go vet ./internal/template/` ok; `golangci-lint run ./internal/template/` → `0 issues.`; `bash -n scripts/release.sh` ok.

## Claim 3 — PR race blast radius (MEASURED only, no change)
Run 41283886701 (pull_request, head 32e79e33ec): six Race Test legs 12m11, 11m13, 15m07, 8m49, 18m35, 12m55; runner time ≈ 78m50s, wall ≈ 18m35s. Race jobs are advisory/non-required (ci.yml:269-284). With Claim 1 left as is, narrowing PR race would leave race mostly to the release gate.

## Claim 4 — required checks: live vs file (RECORDED, nothing changed)
Live main protection = Lint, Build (linux/amd64), Analyze (Go) (go); `.github/required-checks.yml` branches.main = 10 (adds Test (ubuntu-latest), four more Builds, Release PR Multi-OS Gate, Guard Bundle). The file is the declared SSoT rendered by `branch_protection.go` and read by ciwatch, candidate verdicts and release.sh; GitHub enforces only the live 3. Who reduced live protection and when: not observed.

## Baseline attribution
Commands above ran in this run against tree 5c3f0ebf27 plus the local commits (unpushed). Tool provenance: `go test`/`go vet`/`golangci-lint` are the installed toolchain run over this tree (the Go tests exec this tree's `scripts/release.sh`); no moai binary measurement is cited.

## Gaps
- `--skip-ci-check` is unchanged and remains an existing bypass of the CI gate (not of the matrix gate). Not touched, per the leader's ruling.
- Claim 1: GITHUB_TOKEN no-push-trigger is documented behaviour, not re-measured; the `workflow_run` check-run attachment is unmeasured and undocumented.
- Mutant evidence covers three of the five no-switch rows (see Claim 2).
- shellcheck not installed (`command not found`); release.sh only `bash -n`.
- Not run: the full `./internal/template/...` suite and any CI (local commits only, no push).
- Refusals by the worktree-isolation guard (none changed a result): a compound `for … gh api --jq` loop, a `sed -i`, a test pipeline using `PIPESTATUS`, and a Python call whose text named git; each redone as plain commands, the Edit/Write tools, or a redirect to a file.
- Customer-template impact: grep over the 4 `.github` template files and 46 hooks found no mirror or reference of release.sh, auto-merge.yml, ci.yml or required-checks.yml; the `branch-protection.json.gtmpl` vs ADK 3-vs-10 comparison was not done.

## Residual-risk
- Every release needs a dispatched green multi-OS run on the exact SHA before tagging, with no off switch in the script; `--skip-ci-check` still skips the CI gate only. External callers of release.sh that never dispatched the matrix are unobserved (a search of .claude, .moai/docs and docs-site/ko found no mention of the old flag).
- Until the Claim 1 decision changes, bot-merged main commits carry no push CI and race verification relies on PR runs and the release-time gate.
