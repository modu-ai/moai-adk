auditor-model: claude-sonnet-5-5

# Sync audit, card t1833, iteration 2

verdict: FAIL
audited_sha: 505749a927a31338e3a5cb625e91af6825b87957
overall_score: 0.76
must_pass_failed: 1
blocking_count: 1
required_backend: codex fail
convergence_overall: fail
required_backend_fail: codex

Per-dimension scores (flat weighted mode; `evaluator_mode` is not hierarchical):

| Dimension | Weight | Result | Note |
|---|---|---|---|
| Functionality | 40% | 78/100, FAIL (must-pass) | F1-F6 fixes all reproduce as fixed, but the documented guarantee "resets on every path out" is false on the interrupt path (N1) |
| Security | 25% | 78/100, PASS | no Critical/High finding of my own; two hardening points (N2, N3) that codex rates P1 |
| Craft | 20% | 85/100, PASS | 12/12 `TestBisect*`, 8 template selectors green, vet/gofmt/bash -n clean; coverage not re-measured (Gap) |
| Consistency | 15% | 88/100, PASS | Korean runbook reads natively; local-only claim matches the tree; scope is clean |

Must-pass firewall: Functionality is below its threshold because of N1 (blocking). Independently, the required codex gate returned `fail`, so the convergence result is `fail` and the verdict file records `required_backend_fail: codex`. Both routes lead to FAIL.

Audit receipt (required codex gate): `rcpt-e45273c3e30caaae5dc3d6d8`.

## Iteration history

- Iteration 1 (`.moai/reports/t1833/sync-audit-iter1.md`): FAIL 0.62 at `8f1873c58`, blocking F1-F5, optional F6-F9.
- Fix round 1 landed as `e86799369` (F1-F4, F7), `315fc6dc9` (F5, F6), `8a2c39bc9` and `505749a92` (runbook, CHANGELOG, progress). This audit reads `505749a92`.
- This iteration: F1-F5 confirmed fixed, F6 confirmed fixed for its stated scope, F7-F9 confirmed fixed. One new blocking finding (N1) and five optional ones. Codex adds two P1 findings and one P2 (the P2 is the same defect as N1, reproduced with SIGTERM).

## Defect list (this iteration)

- N1 [medium] [blocking] internal/cli/integration_bisect.go:197-206 and CHANGELOG.md entry "resets on every path out" and runbook section 2.2 "성공이든 실패든 끝에서 `git bisect reset`을 실행" - `bisectRestore` runs only on return paths. A signal (Ctrl-C = SIGINT to the process group, or SIGTERM) kills the verb with no restoration: the card worktree is left detached at an intermediate commit with a bisect in progress. Reproduced with SIGINT to the group (moai exit -2, HEAD `3c5c239` DETACHED, BISECT_LOG present) and, per codex, with SIGTERM. This is the same class as iteration-1 F3 (documented guarantee versus behavior). Required fix: either handle SIGINT/SIGTERM (wait for the children to exit, then run `bisectRestore` on an uncancelled path), or reword the CHANGELOG entry and runbook to "on every return path (an interrupt leaves the bisect running; end it with `git bisect reset`)". Add a red/green test for whichever is chosen.
- N2 [medium] [optional] scripts/release.sh:251-252, 269 - the required check names are matched by name only; the jq does not read `.app.slug`, so a check run with the right name from a different app would satisfy the gate. Codex rates this P1 (F6 incomplete). My assessment: publishing such a check needs an installed GitHub App with check-write on this repository, a trusted party, and `--skip-ci-check` already exists as a deliberate bypass; so it is hardening, not a break. Suggested fix: print and compare `.app.slug == "github-actions"` in the jq, and extend the stub rows.
- N3 [low] [optional] internal/cli/integration_bisect.go:156 - the dirty-tree check does not cover ignored files. A plain `git checkout` of a commit that tracks a path silently overwrote an ignored local file (reproduced with git alone: `private.env` changed from `LOCAL-SECRET` to `T`). Codex reports the same loss through the verb. I could not make the verb itself visit the colliding commit in my scratch history (the search converged without it), so the verb-level loss is codex-reported, not independently reproduced. Suggested fix: a runbook line naming the risk (card worktree only, no uncommitted ignored work), or a pre-check on ignored files that collide with tracked paths.
- N4 [low] [optional] internal/cli/integration_bisect.go (endpoints) - the verb names the `--bad` commit as the culprit when the command passes everywhere (`-- true` gave `first bad ... t1004`, the tip) and names the oldest landing when the command fails everywhere (`-- false` gave `t1001`). The runbook records this as limit 1 and requires the pre-run at both ends (section 2.1), so the behavior is documented; a cheap endpoint check in the verb would remove the false culprit.
- N5 [low] [optional] internal/cli/integration_bisect.go (judged command) - a judged command that itself moves HEAD (`git checkout --detach main; sh t.sh`) makes `git bisect run` mark the same commit bad forever; the verb hangs with no timeout (reproduced; I had to kill it). Pathological input, but there is no bound.
- N6 [low] [optional] internal/cli/integration_bisect.go:22 and :170-177 - two cosmetic points: the culprit regex accepts only 40 hex digits (a sha256 repository would report "did not converge" after a successful search), and `--dry-run` prints `git bisect run sh -c <command>` without quoting the command, so the printed line is not copy-pasteable when the command contains spaces (the real run is correct: git logs `'sh' '-c' '...'`).

Optional findings are discretionary and do not by themselves drive the verdict.

## Confirmation of the iteration-1 defects (reproduced by running, not by reading the author's tests)

All scratch work is under `/private/tmp/lane3-audit2` (a scratch repository `s/P` with a linked worktree `s/W`, and `rel/` with a bare local origin).

| Defect | Scenario I built | Observed |
|---|---|---|
| F1 culprit from mixed output | judged command prints `1111...1 is the first bad commit` and a forged `# first bad commit: [2222...]` line, then runs the real test | verb printed `547d37df90b4`, `card: t1003` (the real culprit); the forged lines did not move it |
| F2 GIT_* split | from the primary with `GIT_DIR`/`GIT_WORK_TREE` aimed at the card worktree; from the primary with `GIT_INDEX_FILE`; from the card worktree with `GIT_DIR`/`GIT_WORK_TREE` aimed at the primary; dirty tree with and without `GIT_DIR` | primary: refused in all three primary cases; card worktree with GIT_DIR set: converged on the card tree, primary HEAD `8b33d31 main` unchanged and no bisect state; dirty tree: refused with and without `GIT_DIR` |
| F3 reset failure swallowed | `git` shim on PATH that fails `bisect reset` | exit 1; error "git bisect reset failed (exit status 1); a bisect is still in progress in ...; run git bisect reset by hand (and check out the starting commit 8b33d31be9d6)" |
| F4 failed start leaves HEAD moved | `git` shim that runs the real `bisect start` and then exits 1 | exit 1 "git bisect start failed"; HEAD back on `WT-aud` at the starting commit, no bisect state |
| F5 tag names no commit | non-dry run of `release.sh` in a scratch clone with a bare local origin and a stub `gh` (tag and push stay local); second run moves HEAD during the y/N prompt | tag `v9.9.9-rc.1` commit and the origin tag both equal the verified `a731fbae6979...`; moved-HEAD run: "HEAD moved from a731fbae6979 after the CI gate; refusing to tag", no tag created |
| F6 single green check run | 12 dry-run pages through the real script | green (3 required + extra) pass; neutral/skipped required runs pass; Build missing, only-CodeRabbit, lookalike names (`lint`, trailing space, suffix) refused naming the missing names; pending, failed, cancelled, action_required, empty-conclusion-completed, empty page, api failure refused; gh absent refused; `--skip-ci-check` on a failing page passes with `CI gate OVERRIDDEN` in the warning and `CI gate:     OVERRIDDEN (--skip-ci-check)` in the summary |
| F7-F9 | read runbook and progress | exit-code sentence now matches the observed `bogus exit code 127`; no `동사`/`끝점`; chore count states 3 |

## 1. Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim A: the author's red/green tests and static checks are green on this tree

Evidence (commands run in `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1833`, env `-u MOAI_FACTORY_WORKER -u CLAUDE_PROJECT_DIR TMPDIR=/private/tmp/lane3-audit2 GOTMPDIR=/private/tmp/lane3-audit2`):

```
go test ./internal/cli -run '^TestBisect' -count=1 -v        -> exit 0
  --- PASS: TestBisectNamesTheCulpritCard (3.02s)
  --- PASS: TestBisectRefusesThePrimaryCheckout (2.45s)
  --- PASS: TestBisectRefusesADirtyTree (1.69s)
  --- PASS: TestBisectRefusesWhenOneIsInProgress (1.50s)
  --- PASS: TestBisectDryRunPrintsTheGitCommandsAndMovesNothing (1.71s)
  --- PASS: TestBisectRequiresGoodAndAResolvableRef (1.75s)
  --- PASS: TestBisectReportsAnAbortedSearchAndStillResets (2.27s)
  --- PASS: TestBisectCmdRejectsAnEmptyCommand (0.00s)
  --- PASS: TestBisectIgnoresACulpritClaimPrintedByTheJudgedCommand (2.00s)
  --- PASS: TestBisectActsOnTheCheckedTreeWhenGitDirIsSet (2.85s)
  --- PASS: TestBisectFailsWhenTheResetFails (3.02s)
  --- PASS: TestBisectRestoresHeadWhenStartFailsPartway (2.67s)
  ok  github.com/modu-ai/moai-adk/internal/cli  25.982s

go test ./internal/template -run '^(TestReleaseScriptCIGate|TestReleaseScriptCIGateWithoutGh|TestReleaseScriptTagsTheVerifiedCommit|TestReleaseScriptCharacterization|TestReleaseScriptMatrixGate|TestReleaseScriptHeadAdmission|TestReleaseScriptRcChangelogSkip|TestGitHubFlowSweepGuard)$' -count=1 -v  -> exit 0
  --- PASS: TestGitHubFlowSweepGuard (5.42s)
  --- PASS: TestReleaseScriptCIGate (11.14s)
  --- PASS: TestReleaseScriptCIGateWithoutGh (1.33s)
  --- PASS: TestReleaseScriptTagsTheVerifiedCommit (1.35s)
  --- PASS: TestReleaseScriptCharacterization (8.64s)
  --- PASS: TestReleaseScriptHeadAdmission (3.81s)
  --- PASS: TestReleaseScriptRcChangelogSkip (3.80s)
  --- PASS: TestReleaseScriptMatrixGate (12.76s)
  ok  github.com/modu-ai/moai-adk/internal/template  48.545s

bash -n scripts/release.sh                      -> exit 0
go vet ./internal/cli                           -> exit 0
go vet ./internal/template                      -> exit 0
gofmt -l <5 touched Go files>                   -> exit 0, no output
```

Baseline-attribution: tree HEAD `505749a927a31338e3a5cb625e91af6825b87957` (`git rev-parse HEAD` first command of this audit, equal to the audited commit; `git status --short` clean). Judging build for the scratch reproductions: `/private/tmp/lane3-audit2/moai-t1833`, built in this audit with `go build -o ... ./cmd/moai` from this tree at `505749a92`; its `version` prints `v3.1.3 none` (no ldflags), so its identity is the build path plus the tree HEAD. The installed `moai` (`/Users/goos/go/bin/moai`, `v3.2.0-rc.36-t1822c`, commit `ee981a932`) is a strict ancestor of the tree HEAD (`git merge-base --is-ancestor ee981a932 HEAD` exit 0) and predates the verb, so I used it only for `moai verify audit-plan`, never for the verb under test. The `audit_multi` result reports `build_commit: ee981a932` and a build-lag note for the MCP server; its codex review reads the tree by `project_root`, and its findings quote HEAD `505749a92`.

Gaps: coverage was not re-measured (no package-wide run, by rule). Test counts are what the targeted selectors swept; I did not run the other `internal/cli` tests.

### Claim B: no live remote was contacted and the committed tests cannot reach the real `gh` (item 6)

Evidence: a canary `gh` that appends to a log and exits 9 was put FIRST on the PATH of the `go test` process:

```
PATH=/private/tmp/lane3-audit2/canary:$PATH go test ./internal/template -run '^(TestReleaseScriptCIGate|TestReleaseScriptTagsTheVerifiedCommit)$' -count=1  -> exit 0, ok 11.869s
cat /private/tmp/lane3-audit2/canary.log  -> No such file (never invoked)
```

`release_fixture_test.go:364` builds `PATH = <stub bin>:<os PATH>` (stub first); `TestReleaseScriptTagsTheVerifiedCommit` prefixes only a git shim to that stub PATH; `TestReleaseScriptCIGateWithoutGh` uses a symlink-only PATH with no `gh`. My own release.sh runs used `PATH=<stub>:/usr/bin:/bin` and the no-`gh` run a symlink-only directory; the real `gh` was not on any of them (`gh` call log held only stub calls).

Baseline-attribution: same tree and test binary as Claim A. Gaps: the stub does not model GitHub's real paging or HTTP errors. Residual-risk: low.

### Claim C: CI_REQUIRED_CHECKS names match what the CI workflow publishes on a push to main (item 2, gate)

Evidence (reading plus runs):

```
.github/workflows/ci.yml: on.push.branches [main, 'ci/**']; jobs: lint -> name: Lint (no condition);
  build -> name: Build (${{ matrix.goos }}/${{ matrix.goarch }}), matrix linux/amd64 first, no if;
  test -> name: Test (${{ matrix.os }}) if go_code=='true'; test-skip-marker -> same name if go_code!='true'
  (strict negation: exactly one of the two publishes `Test (ubuntu-latest)`; matrix os [ubuntu-latest]).
.github/required-checks.yml branches.main.contexts: Lint, "Test (ubuntu-latest)", "Build (linux/amd64)", ... (the three are a subset)
```

So a legitimately green push-to-main commit publishes all three names, and the exact-name awk match (checked with lookalike and trailing-space names) does not over-accept. Paging: the script uses `gh api --paginate ... --jq` (applied per page), `per_page=100`; the HEAD re-check precedes `git tag -a "$VERSION" -F "$TMP_NOTES" "$LOCAL_SHA"`.

Baseline-attribution: `.github/workflows/ci.yml` and `.github/required-checks.yml` at tree HEAD `505749a92`, read in this run.

Gaps (cannot be confirmed without live GitHub, per the task): that GitHub reports these exact names for a real main-push commit; that `Test (ubuntu-latest)` from the skip marker is a check run (not a status) on a docs-only push; the real `--paginate --jq` output format over multiple pages; whether other workflows publishing check runs on main pushes (auto-merge, graph-freshness, CodeQL, race tests) can be failed, cancelled or still running at release time and so produce a false refusal. The gate is stricter than branch protection by design (runbook limit 8); a false refusal fails safe and `--skip-ci-check` exists. Residual-risk: false refusal possible, false accept limited to N2.

### Claim D: the runbook and CHANGELOG claims match the code and tests (item 3)

Evidence:

- Landing shape (item 4), re-run at `fb749d372` (equal to local `main`, `git rev-parse main` = `fb749d372e61a6b10c53091362c76b4412827427`):

```
git log --first-parent --format='%P|%s' fb749d372 -n 40   -> two-parent + card id: 40, two-parent without: 0, direct: 0
git log --first-parent --format='%P|%s' fb749d372 -n 100  -> 64 / 4 / 32
the 4 two-parent without card id: 1 integration merge ("Merge WT-codex-moai-home-add-dir into main (codex --add-dir moai home, integration merge)") + 3 PR merges (#1816, #1812, #1772)
one-parent subjects starting with "chore": 3; ending in "(#N)": 29
oldest of the 100: 2026-10-07
```

These match `progress.md` and runbook section 1 (40/0/0 and 64/4/32; 3 chore and the PR-squash remainder).

- "resets on every path and fails the command when it cannot confirm restoration": the second half is true (F3 and F4 reproduced). The first half is true for every return path and false for an interrupt (N1). The CHANGELOG sentence says "resets on every path out", the runbook "성공이든 실패든 끝에서"; neither names the interrupt exclusion.
- Test counts in the CHANGELOG (12 `TestBisect*`, 11 `TestReleaseScriptCIGate` rows) match the runs above.
- Local-only claim: the runbook says it is not mirrored into the template. `.moai/docs` holds 42 files and `internal/template/templates/.moai/docs` 9; `comm` shows many local-only docs (for example `hook-development.md`, `kickoff-autonomy.md`, `local-linear-integration.md`), and `git ls-files` shows the runbook tracked only at `.moai/docs/`. The claim is consistent with how the tree treats other local-only docs.
- Korean register: the runbook reads as clean written Korean. `동사` and `끝점` are gone. The one mildly translated phrase is "과잉 거부가 과잉 허용보다 싸기 때문이다" (cost as "cheap"); optional wording only.
- Scope hygiene (item 5): `git diff --name-only fb749d372 HEAD` lists exactly the 10 files: the runbook, `progress.md`, `sync-audit-iter1.md`, `CHANGELOG.md`, `integration.go`, `integration_bisect.go`, `integration_bisect_test.go`, `release_ci_gate_test.go`, `release_fixture_test.go`, `scripts/release.sh`. `git status --short` is clean before and after my runs (this verdict file is the only new file). The only tracked path under `.moai/state/` is `.moai/state/.gitkeep`. No `.orig/.rej/.tmp/.bak` files tracked.

Baseline-attribution: tree HEAD `505749a92`; commands run in this audit.

Gaps: I did not re-render Korean text through a humanize pass; the register judgment is mine. Residual-risk: low.

### Claim E: verb behavior around the edges (item 2, verb)

Evidence (scratch repository, judging build as above): annotated-tag refs for `--good` and `--bad` resolved (`^{commit}` peel) and converged on t1003; `--bad` omitted defaults to HEAD; starting from detached HEAD restored `DETACHED` at the starting commit; exit code 128 from the judged command gave "the search did not converge" and a restored tree; multi-argument commands are quoted by the shared `shellJoinArgs`; N1, N4, N5 as above.

Gaps: git-version dependence (`--first-parent` for `git bisect start` needs a recent git; observed on the installed git only). Residual-risk: see the N-list.

### Audit plan and audit tool (item 7), tool refusals

`moai verify audit-plan --project-root /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1833` returned `config_status: ok`, `model: codex`, `model_source: config`, `cross_model_active: true`, `cross_model_required: true`, gates claude=off, codex=required, glm=off, `enforced_required: ["codex"]`. I called `mcp__moai__audit_multi` once with `project_root` set to this tree, `target: baseBranch`, and no `gates` argument (the plan was applied by the server, `plan_source: config`).

audit_multi result digest, for the orchestrator's check:

- `overall_verdict`: fail
- `gate_unmet`: absent from the result (the overall verdict is fail from the backend verdict)
- `plan_source`: config
- `per_backend_verdicts`: backend codex, gate required, verdict fail
- `audit_receipt`: rcpt-e45273c3e30caaae5dc3d6d8
- my verdict is not final until the orchestrator's `moai verify audit-plan --result-file` check passes.

Tool refusals: none. One environment limit: `os.setsid` was denied in my sandbox (PermissionError), so the interrupt test used `os.setpgid` instead, which gives the same process-group SIGINT a terminal sends. I ran no `claude_audit` or `glm_audit` call because the plan turns both off.

## Gaps (consolidated)

- Live GitHub behavior: real check-run names, skip-marker publication, paging output, other workflows' check runs on a main push (Claim C).
- Coverage percentage not measured; only targeted selectors ran (machine-load rule).
- N3 through the verb is codex-reported only; my verb-level scratch history never visited the colliding commit.
- Git-version dependence of `git bisect start --first-parent`.
- The reuse of a verified result: none (`moai verify run` not used; every observation above was made directly in this run).

## Residual-risk

- After N1 is fixed or reworded, an interrupted run in a card worktree is the main remaining operational hazard; it is recoverable by hand (`git bisect reset`), and the next run refuses with a message naming that command.
- The gate matches check names only (N2); a trusted-app impostor is the residual.
- The release path is verified end to end only against a local bare origin and a stub `gh`.

## Required repair list for the fix round

1. N1 (blocking): handle SIGINT/SIGTERM in `runIntegrationBisect`, or reword the CHANGELOG and runbook claim to return paths only; add a test.
2. Optional, at the author's discretion: N2 (`app.slug` in the jq), N4 (endpoint check), N3 (runbook line), N5/N6 (hang bound, dry-run quoting).

Re-audit scope: the N1 delta (the signal path and the two claim sentences) plus a re-run of the `TestBisect*` selector and the codex gate; the confirmed F1-F9 fixes need no re-measurement unless the verb body changes beyond signal handling.
