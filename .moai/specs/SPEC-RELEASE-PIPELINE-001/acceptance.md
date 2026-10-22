---
id: SPEC-RELEASE-PIPELINE-001
title: "Release pipeline coverage and cost — acceptance criteria"
version: "0.6.1"
created: 2026-10-17
updated: 2026-10-23
author: manager-spec
---

# acceptance.md — SPEC-RELEASE-PIPELINE-001

## §A Verification model

Every AC is binary-testable through one of: a Go guard test in `internal/template/` (reads the workflow YAML / executes the extracted jq program against the committed fixture), a single-invocation read-only command, or a workflow-level observation. ACs are classed **release-blocking** (two-cell adoption: RED-now cell + green-path cell, per `.claude/rules/moai/development/verification-completeness.md` §2 and §2.1) or **regression-guard** (preserve an already-green invariant). Every GREEN command is run with `-v` and its swept test count read before its verdict (§1.1 empty-sweep clause).

## §B Evidence ledger

**Document-level pin (binds every cell below that carries no pin of its own):** tree `bf0a0f35fc280e83c2f9fd76d0e2a5de39b7e8b7`, measured 2026-10-22 (session clock `2026-10-22T12:26:32Z` at capture), worktree `.moai/worktrees/t1536`, clean tracked state. This replaces the v0.3.0 pins (`0bd1da497…` / `c8ede48f9…`): every cell was re-executed at this tree, none carried over. Cells added or re-measured in v0.5.0 (EV-12, EV-14, EV-16) carry their own pin, tree `c1a53b483d380f0f2ef00defe0e555a2681e7bd5`, measured 2026-10-22 (session clock `2026-10-22T12:54:18Z`); `git diff --stat bf0a0f35fc HEAD -- .github` at that tree printed nothing, so every `.github/` citation holds at both pins. Where stdout and stderr were not separated by the harness, the cell says "combined output".

**EV-1 — parity guard baseline (regression-guard pin, AC-RPC-001).**
Command: `go test ./internal/template/ -run '^(TestTestInputFiltersSingleSource|TestTestInputFiltersCoversGoTestInputs)$' -count=1 -v`
Verbatim stdout:
```
=== RUN   TestTestInputFiltersSingleSource
--- PASS: TestTestInputFiltersSingleSource (0.01s)
=== RUN   TestTestInputFiltersCoversGoTestInputs
--- PASS: TestTestInputFiltersCoversGoTestInputs (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/template	0.924s
```
Exit code: 0 · Swept: 2 of 2 named.

**EV-2 [RETIRED]** — release workflow cancellation flag; moved with card item 3 to the follow-up card (v0.4.0 split).

**EV-3 — pre-fix jq program from release-drafter-cleanup.yml:65 (RED for AC-RPC-010).**
Command: `jq -r --arg threshold_days 30 'now | floor as $now | (($threshold_days | tonumber) * 86400) as $threshold_secs | .[] | select(.isDraft == true and (($now - (.createdAt | fromdate)) > $threshold_secs) and (.targetCommitish | startswith("release/") | not)) | .tagName' .moai/reports/t1536/red-drafter-fixture.json`
Verbatim output (stderr): `jq: error (at .moai/reports/t1536/red-drafter-fixture.json:6): Cannot iterate over number (1792671846....)` · Exit code: 5
Why red: the `as` binding rebinds the iteration input from the releases array to `now`'s numeric output, so `.[]` iterates a number. The fixture is the gitignored plan-phase evidence copy; its shape is the four-draft camelCase list.

**EV-4 — fixed-form divergence on the same fixture (green-path pin for AC-RPC-010).**
Command: `jq -r --arg threshold_days 30 '(now|floor) as $now | (($threshold_days | tonumber) * 86400) as $threshold_secs | .[] | select(.isDraft == true and (($now - (.createdAt | fromdate)) > $threshold_secs) and (.targetCommitish | startswith("release/") | not)) | .tagName' .moai/reports/t1536/red-drafter-fixture.json`
Verbatim stdout: `stale-draft` · Exit code: 0
Four drafts (stale / fresh / release-branch-targeted / published); only `stale-draft` is selected — threshold, draft flag, and release-branch exemption verified in one divergence against EV-3.

**EV-5 — cost jobs carry no input gate (RED for AC-RPC-002/003; baseline for the AC-RPC-004 guard).**
Command: `grep -nE "^  [a-z][a-z0-9-]*:|^    if:|^    needs:|^    continue-on-error:" .github/workflows/ci.yml`
Verbatim stdout (excerpt; the full output was redirected to a session scratch file, which is not cited — the excerpt lines below are copied from it): consecutive lines `772:  lint:` / `819:  build:` / `889:  constitution-check:` / `893:    continue-on-error: true` / `950:  test-browser:` — no `if:`/`needs:` between any of them; contrast `102:  test:` / `105:    needs: detect` / `106:    if: needs.detect.outputs.go_code == 'true'` · Exit code: 0
Why red: the four cost jobs are ungated on every event class; no gated-family verdict surface exists.

**EV-6 — stage-vs-job budget map (RED pin for AC-RPC-007/008).**
Command: `grep -nE "timeout-minutes:|-timeout (17m|35m|25m)" .github/workflows/ci.yml .github/workflows/release-pr-multi-os.yml`
Verbatim stdout (excerpt of the 27 lines): `.github/workflows/ci.yml:953:    timeout-minutes: 20` · `.github/workflows/ci.yml:1018:        run: go test ./internal/web/ -run 'AppJs.*Fire' -v -count=1 -timeout 17m` · `.github/workflows/release-pr-multi-os.yml:121:    timeout-minutes: 40` · `.github/workflows/release-pr-multi-os.yml:300:            go test -json -race -count=1 -timeout 35m $pkgs > test-stream.json || rc=$?` · `.github/workflows/release-pr-multi-os.yml:302:            go test -json -race -count=1 -timeout 35m -run "$SHARD_RUN" "$SHARD_PKG" > test-stream.json || rc=$?` · Exit code: 0
Why red: in ci.yml no `timeout-minutes` appears between :953 (job) and :1018 (green step) or after it for the mutation phase (:1020-) — the browser green phase (17m) and mutation cycle share one 20m job budget with no stage bounds.

**EV-7 — no re-query before deletion (RED for AC-RPC-011).**
Command: `grep -c "gh release view" .github/workflows/release-drafter-cleanup.yml`
Verbatim stdout: `0` · Exit code: 1
Why red: the delete loop (:89-98) deletes straight from the listing.

**EV-8 — no concurrency block on the cleanup side (RED for AC-RPC-012).**
Command: `grep -c "concurrency" .github/workflows/release-drafter-cleanup.yml`
Verbatim stdout: `0` · Exit code: 1
Why red: no concurrency relationship with `release-drafter.yml`, whose own group (:19-21) is ref-keyed with `cancel-in-progress: true` (read directly, file byte-unchanged since v0.3.0).

**EV-9 — first-failure record absent from every upload path (RED for AC-RPC-009, upload clause).**
Command: `grep -c "attempt1" .github/workflows/ci.yml`
Verbatim stdout: `0` · Exit code: 1
Why red: `scripts/ci/retry-flaky.sh:166` copies the first-failure record to `${log_file%.log}.attempt1.json` (default `.moai/logs/ci-flaky-retry.attempt1.json`) before the re-run at :169, but no ci.yml upload step names it.

**EV-10 — cleanup fetch requests an unsupported field (RED for AC-RPC-013; D10).**
Command: `gh release list --json tagName,isDraft,createdAt,targetCommitish -L 1`
Verbatim combined output:
```
Unknown JSON field: "targetCommitish"
Available fields:
  createdAt
  isDraft
  isImmutable
  isLatest
  isPrerelease
  name
  publishedAt
  tagName
```
Exit code: 1 · gh version: `gh version 2.92.0 (2026-04-28)` (local; the runner's gh is unmeasured — Gap)
Why red: the workflow's fetch (:56-59) requests the same field list with `-L 100`; the field validation is limit-independent, so the step exits 1 before the filter at :62-76 is reached. EV-3 is therefore a second, independent defect behind this one.

**EV-11 — release race step carries no step-level ceiling (baseline for the AC-RPC-007 release-half guard; not a RED cell since v0.5.0 — under decision-index Q9's applied default the absence of a race-step ceiling is the compliant state).**
Command: `grep -n "timeout-minutes" .github/workflows/release-pr-multi-os.yml`
Verbatim stdout:
```
41:    timeout-minutes: 5
102:  # this job keeps `timeout-minutes: 40` (no ceiling raised), no test is skipped
121:    timeout-minutes: 40
218:      # the job's own timeout-minutes; this value exists to stop a package from
283:          # job's timeout-minutes: 40 budget. It is NOT raised. The value is only
367:    timeout-minutes: 5
```
Exit code: 0
What it shows: the only non-comment occurrences are job-level (:41 detect, :121 matrix, :367 gate); the race step at :271-:307 has none, so the `always()` compress (:309) and upload (:315) steps run on whatever of the 40-minute job budget the step leaves (EV-15 measures that remainder).

**EV-12 — the repair's fetch source supports the field on a draft (green-path pin for AC-RPC-013; re-pinned to a draft tag in v0.5.0, plan-audit iter-4 D16).**
Command: `gh release view v3.1.3 --repo modu-ai/moai-adk --json tagName,isDraft,createdAt,targetCommitish`
Verbatim stdout: `{"createdAt":"2026-08-21T19:28:08Z","isDraft":true,"tagName":"v3.1.3","targetCommitish":"main"}` · Exit code: 0 · Pin: `c1a53b483d380f0f2ef00defe0e555a2681e7bd5` (remote release state; the tree pin records where it was measured from)
Shows the per-draft view Q7's default relies on returns all four filter fields for a release with `isDraft: true` (the v0.4.0 form, with no tag, read the latest published release instead).

**EV-13 — REST payload shape for decision-index Q7's alternate (shape pin, not a RED cell).**
Command: `gh api "repos/modu-ai/moai-adk/releases?per_page=1" --jq '.[0] | [.tag_name, .draft, .created_at, .target_commitish] | @json'`
Verbatim stdout: `["v3.1.3",true,"2026-08-21T19:28:08Z","main"]` · Exit code: 0
Shows the REST payload carries the four needed values under snake_case keys — the mapping AP-10 requires if the alternate is chosen.

**EV-14 — no ci.yml job reads the classification step's result (RED for AC-RPC-003, counter-case "classification failure × verification success"; v0.5.0, D12).**
Command: `grep -n "needs.detect.result" .github/workflows/ci.yml`
Verbatim stdout: (empty) · Exit code: 1 (observed in the same run by appending `; echo "exit=$?"`, which printed `exit=1`; the single-invocation re-run printed nothing) · Pin: `c1a53b483d380f0f2ef00defe0e555a2681e7bd5`
Why red, and why for the right reason: every reference to `detect` in ci.yml reads `needs.detect.outputs.go_code` (:106, :340-:602, :682) and the cost jobs `lint` (:772) and `build` (:819) carry no `needs:` at all (EV-5). So when `detect` fails and the lint/build steps succeed, `Lint` and the five `Build (…)` names publish `success` — the outcome REQ-RPC-003 forbids. The cell flips when M2's verdict step uses `needs.detect.result` (AC-RPC-003's test, not this grep, is the verdict); nothing outside this SPEC's ci.yml change can flip it.

**EV-15 — release leg budget measurement (budget input for AC-RPC-007 and decision-index Q9; an externally observed CI result, so per `verification-completeness.md` §2.1 it is a measurement pin, never a RED cell).**
Command: `gh run view 39729541531 --repo modu-ai/moai-adk --json jobs --jq '.jobs[] | select(.name|startswith("Release Verify (windows")) | [.name, .startedAt, .completedAt, (.steps[] | select(.name|startswith("Run tests")) | .startedAt, .completedAt)] | @tsv'`
Verbatim stdout (columns: job, job start, job end, race-step start, race-step end):
```
Release Verify (windows-latest, cli-3)	2026-10-22T06:41:22Z	2026-10-22T07:19:20Z	2026-10-22T06:42:32Z	2026-10-22T07:19:14Z
Release Verify (windows-latest, cli-1)	2026-10-22T06:41:23Z	2026-10-22T07:20:32Z	2026-10-22T06:42:56Z	2026-10-22T07:20:24Z
Release Verify (windows-latest, cli-4)	2026-10-22T06:41:23Z	2026-10-22T07:18:10Z	2026-10-22T06:43:07Z	2026-10-22T07:18:02Z
Release Verify (windows-latest, rosterguard)	2026-10-22T06:41:22Z	2026-10-22T06:49:56Z	2026-10-22T06:43:18Z	2026-10-22T06:49:48Z
Release Verify (windows-latest, cli-2)	2026-10-22T06:41:23Z	2026-10-22T07:08:51Z	2026-10-22T06:45:44Z	2026-10-22T07:08:43Z
Release Verify (windows-latest, remainder)	2026-10-22T06:41:24Z	2026-10-22T07:17:14Z	2026-10-22T06:42:41Z	2026-10-22T07:17:08Z
```
Exit code: 0 · Measured 2026-10-22 from tree `c1a53b483d380f0f2ef00defe0e555a2681e7bd5`.
Arithmetic (from the rows above): pre-step = race-step start − job start, maximum 4m21s (cli-2, which runs `go vet`); race step longest 37m28s (cli-1) and 36m42s (cli-3); evidence tail = job end − race-step end, 6-8s; cli-1 used 39m09s of the 40-minute job. The race step's wall time includes compilation and every package binary, while `-timeout 35m` bounds one binary. A ceiling leaving a 5-minute pre-step allowance (4m21s rounded up) and a 1-minute evidence allowance needs ≤ 34.
Run classification (v0.6.1, plan-audit iter-5 D20): run 39729541531 is a FAILING run — `gh run view 39729541531 --repo modu-ai/moai-adk --json conclusion,event,headBranch` gives `failure` / `workflow_dispatch` / `WT-cloud-windows-portability-repair` (not main) — and the windows legs cli-1, cli-2, cli-3, cli-4 and remainder concluded `failure`, race step included. Every figure above is therefore an n=1, windows, non-main-branch, failing-run measurement. The race-step durations (37m28s, 36m42s) are durations of failing legs and bound nothing about a green leg's race step; no lower bound for a ceiling that leaves green legs green is derived from them. The pre-step figure is kept as the 5-minute constant because the pre-step precedes the tests and is outcome-independent, under the same n=1 / windows / non-main-branch label. Assumption (unmeasured, not a requirement): a green-run re-measurement — `gh run list --workflow release-pr-multi-os.yml --status success`, then this EV's jq on a listed run — is the follow-up that would establish the green-leg bound. Only windows legs were fetched; ubuntu/macOS pre-step durations are a Gap.

**EV-16 — event classes on which each required-name producer runs (baseline for AC-RPC-004; v0.5.0, D14).**
Command: `grep -nE "^on:|^  (push|pull_request|workflow_dispatch|schedule):|^    branches:|^      - (main|'ci/\*\*')$|name: (Analyze \(Go\)|Guard Bundle|Release PR Multi-OS Gate)$" .github/workflows/ci.yml .github/workflows/codeql.yml .github/workflows/release-pr-multi-os.yml`
Verbatim stdout:
```
.github/workflows/codeql.yml:3:on:
.github/workflows/codeql.yml:4:  push:
.github/workflows/codeql.yml:5:    branches: [main]
.github/workflows/codeql.yml:6:  pull_request:
.github/workflows/codeql.yml:7:    branches: [main]
.github/workflows/codeql.yml:8:  schedule:
.github/workflows/codeql.yml:49:    name: Analyze (Go)
.github/workflows/codeql.yml:104:    name: Analyze (Go)
.github/workflows/release-pr-multi-os.yml:12:on:
.github/workflows/release-pr-multi-os.yml:13:  pull_request:
.github/workflows/release-pr-multi-os.yml:14:    branches: [main]
.github/workflows/release-pr-multi-os.yml:16:  workflow_dispatch:
.github/workflows/release-pr-multi-os.yml:363:    name: Release PR Multi-OS Gate
.github/workflows/ci.yml:16:on:
.github/workflows/ci.yml:17:  push:
.github/workflows/ci.yml:23:    branches:
.github/workflows/ci.yml:24:      - main
.github/workflows/ci.yml:25:      - 'ci/**'
.github/workflows/ci.yml:26:  pull_request:
.github/workflows/ci.yml:27:    branches: [main]  # main으로 향하는 모든 PR에서 CI 실행
.github/workflows/ci.yml:32:  workflow_dispatch:
.github/workflows/ci.yml:285:    name: Guard Bundle
```
Exit code: 0 · Pin: `c1a53b483d380f0f2ef00defe0e555a2681e7bd5`
Shows: `Release PR Multi-OS Gate` runs only on `pull_request` and `workflow_dispatch`, and `Analyze (Go)` never on a `ci/**` push, so `branches.main.contexts` is satisfiable only on pull requests to main — the per-event-class binding AC-RPC-004 asserts. `ci/**` contexts and `release_gate` name ci.yml jobs only, and ci.yml runs on push to `main` and to `ci/**`.

## §C Edge cases

- E1: mixed diff (docs + browser inputs) — browser job runs, build matrix does not (class matched per job, not per PR).
- E2: filter file itself edited — every class including `go_code` re-matches.
- E3: cleanup listing lands mid-drafter-update — shared group queues; no mid-deletion cancellation (REQ-RPC-012).
- E4: draft published between listing and deletion — re-query skips it and logs the skip (REQ-RPC-011).
- E5 [RETIRED] — rapid re-push cancellation; moved with card item 3.
- E6: green phase at max ceiling — mutation phase retains its declared budget.
- E7: a draft whose target branch cannot be resolved by the fetch — not selected (REQ-RPC-013).
- E8: docs-only push to main — Lint / Test / Build (linux/amd64) still publish, so `scripts/release.sh`'s `release_gate` check finds them (REQ-RPC-004, Q8).
- E9: pull request whose classification step fails, in a run that is not itself cancelled, while every lint/build step succeeds — the real jobs run and conclude `failure`, the markers run and conclude `failure`, so every check run under `Lint` and each `Build (…)` concludes `failure` (REQ-RPC-003; EV-14 is the RED-now cell).
- E10: push to `ci/**` — only `branches."ci/**".contexts` is expected; `Release PR Multi-OS Gate` and `Analyze (Go) (go)` are absent on that event by trigger design and are not a failure (REQ-RPC-004; EV-16).

## §D AC matrix (Given-When-Then)

| AC | REQ | Class | Given / When / Then |
|---|---|---|---|
| AC-RPC-001 | REQ-RPC-001 | regression-guard | Given the tree carries the shared filter and its three consumers, When the parity guard tests run (EV-1 command), Then both pass with 2 swept and any consumer drift, inline copy, or lost required root fails the suite; the release consumer's fail-open wiring (`continue-on-error: true` :64 + `!= 'false'` gate :120) stays intact, and ci.yml:106 / codeql.yml:53-56 keep their landed gating. |
| AC-RPC-002 | REQ-RPC-002 | release-blocking | Given ci.yml at the plan tree (EV-5), When `TestCICostJobsInputGated` runs, Then it FAILS (no gate on lint/build/constitution-check/test-browser); green path: M2 gates each cost job on its shared-filter input class for pull_request events, fail-open on unevaluated classification, always-run on the Q8-exempt event classes, and the test passes (swept 1). |
| AC-RPC-003 | REQ-RPC-003 | release-blocking | Given no gated-family verdict surface exists and no ci.yml job reads the classification step's result (EV-5; EV-14: `grep -n "needs.detect.result"` → empty, exit 1, pin `c1a53b483d`), When `TestCIGatedFamilyVerdictSemantics` runs, Then it FAILS — its counter-case cell **classification failure × verification success** (event pull_request, `detect` = failure, every lint/build step = success) evaluates the conclusions of the check runs published under `Lint` and each `Build (…)` name from ci.yml's job and step conditions and obtains `success` today, where REQ-RPC-003 requires `failure` (E9). Green path (M2): for each retained name over the pull_request cells — (i) real job `failure`/`cancelled` → `failure`; (ii) `detect` success with a positively observed non-match → the marker publishes `success` naming the exclusion and the real job is skipped; (iii) `detect` concluded anything other than `success` (failure or cancellation; a skip is unreachable today because `detect` carries no `if:`) in a run that is not itself cancelled → the real job runs and concludes `failure` both with every verification step succeeding (the counter-case) and with a verification step failing, and the marker also runs and concludes `failure`, so every check run published under the name concludes `failure`; (iv) the run itself cancelled (ci.yml:52-54 conditional cancellation) → every cost-job and marker condition is `!cancelled()`-guarded and none is `always()`-based, so no cost job or marker is required to start after the cancellation, and no required name concludes `success` through a job that had not completed when the run was cancelled; and on the push cells (Q8) no marker runs and the name equals the real job's own outcome whatever `detect` concluded. The test passes with its swept count read (1). |
| AC-RPC-004 | REQ-RPC-004 | regression-guard | Given `.github/required-checks.yml` and the triggers of the workflows that publish its names (EV-16: ci.yml on push to `main`/`ci/**`, pull_request to main, workflow_dispatch; codeql.yml on push to main, pull_request to main, schedule; release-pr-multi-os.yml on pull_request to main and workflow_dispatch) — each set satisfied today only because the cost jobs are ungated (EV-5) — When `TestCIRequiredNamesSatisfied` runs, asserting for every classification outcome (docs-only, go_code, harness-only, classification failure) that every `branches.main.contexts` name is reported by some job on a pull_request to main, every `branches."ci/**".contexts` name by some ci.yml job on a push to `ci/**`, and every `release_gate` name by some ci.yml job on a push to main — "reported" meaning published with any conclusion; pass/fail semantics belong to AC-RPC-003 — plus `release_gate` ⊆ `branches.main.contexts`, harness roots inside `go_code`, and gtmpl + mirror untouched, Then it passes before and after M2. No name is asserted on an event class whose set does not list it (E10). Completion (§1.1): two negative controls over in-memory copies, both observed red and captured in E8 before M2's gating lands — (a) ci.yml with `lint` gated on a class and no marker publishing `Lint` must report `Lint` missing for the docs-only pull_request cell; (b) a required-checks copy that adds `Release PR Multi-OS Gate` to `release_gate` must report that name unpublished on push to main, which proves the per-event-class binding is checked rather than assumed. `TestRequiredChecks_ReleaseGateSubsetOfMain` stays green. (Reclassified from release-blocking in v0.4.0: under Q4's retained-names default no RED-now cell exists on the real tree, so a release-blocking label would be vacuous.) |
| AC-RPC-005 [RETIRED] | REQ-RPC-005 [RETIRED] | — | Moved to the follow-up card with card item 3 (v0.4.0 split). |
| AC-RPC-006 [RETIRED] | REQ-RPC-006 [RETIRED] | — | Moved to the follow-up card with card item 3 (v0.4.0 split). |
| AC-RPC-007 | REQ-RPC-007 | release-blocking | Given the browser budget map of EV-6 and the release budget measurement of EV-11/EV-15, When `TestStageCeilingBelowJobBudget` runs, Then it FAILS on its browser half (neither the green phase :1013-:1018 nor the mutation phase :1020- carries a step-level ceiling, EV-6). Green path (M4): browser — each phase carries a step ceiling strictly below the job's 20 with declared headroom for compress+upload; release — independent of decision-index Q9, the test asserts that compress (:309) and upload (:315) remain separate `if: always()` steps after the race step, that `-timeout 35m` (:300/:302) and job `timeout-minutes: 40` (:121) are unchanged, and that **if** the race step carries `timeout-minutes: T` then T + 5 + 1 ≤ 40, with the 5-minute pre-step and 1-minute evidence allowances held as constants in the test that cite EV-15 rather than read from the workflow (under Q9's applied default no T is present). The release half holds today, so it has no RED-now cell; it is completed by two negative controls over an in-memory copy of release-pr-multi-os.yml, both observed red and captured in E8 — T = 38 (the withdrawn v0.4.0 window) and an upload step whose `if: always()` is removed. The test makes no "T ≥ 35" or "never pre-empts `-timeout 35m`" assertion. |
| AC-RPC-008 | REQ-RPC-008 | release-blocking | Given the browser job runs green+mutation under one budget (EV-6), When `TestBrowserPhaseSplit` runs, Then it FAILS; green path: green and mutation phases carry independent stage bounds (E6 holds), SPEC-APPJS-FIRE-GUARD-001 phase order and restore assertions unchanged. |
| AC-RPC-009 | REQ-RPC-009 | release-blocking | Given registry-gated retries through `retry-flaky.sh` whose attempt1 record no upload path carries (EV-9), When `TestRetryPolicyPreservesFirstFailure` runs, Then its upload assertion FAILS while the wrapper-only and preserve-before-retry assertions hold; green path: each retry-wrapped job's upload includes `.moai/logs/ci-flaky-retry.attempt1.json` (always(), warn-on-absent) and all three assertions pass. |
| AC-RPC-010 | REQ-RPC-010 | release-blocking | Given the workflow's jq program (EV-3: exit 5), When `TestCleanupFilterSelectsStaleDrafts` executes the extracted program against `internal/template/testdata/release-drafter-cleanup.json` (four drafts, keyed as the repaired fetch delivers them), Then pre-fix it FAILS and post-fix it selects exactly `stale-draft` (EV-4 divergence embedded). |
| AC-RPC-011 | REQ-RPC-011 | release-blocking | Given the delete loop deletes straight from the listing (EV-7), When `TestCleanupRequeriesBeforeDelete` runs, Then it FAILS; green path: each deletion is preceded by a state re-query; a changed draft (E4) is skipped and the skip recorded. |
| AC-RPC-012 | REQ-RPC-012 | release-blocking | Given no concurrency block on the cleanup (EV-8) and a ref-keyed `cancel-in-progress: true` drafter group (release-drafter.yml:19-21), When `TestDrafterCleanupConcurrency` runs, Then it FAILS; green path: both draft-mutating workflows share one ref-independent LITERAL group name (invariant under at least two distinct `github.ref` values) with `cancel-in-progress: false` on BOTH sides, and the test passes. |
| AC-RPC-013 | REQ-RPC-013 | release-blocking | Given the fetch requests `targetCommitish` from `gh release list` (EV-10: `Unknown JSON field`, exit 1), When `TestCleanupFetchFieldsSupported` runs, Then it FAILS (requested field outside the list-supported set); green path: every fetch requests only fields its subcommand supports, the filter's four read fields (`isDraft`, `createdAt`, `tagName`, `targetCommitish`) are supplied under those names (EV-12 shows `gh release view` supplies `targetCommitish`; a REST source is mapped first, EV-13), the `release/*` exemption is still present, a draft with an unresolvable target is not selected (E7), and the test passes. |

### §D.0 Mutant probes (§2 of verification-completeness.md)

- AC-RPC-013 vs "drop the field": a mutant that removes `targetCommitish` from the fetch passes a field-support-only check but deletes `release/*` drafts. Blocked by the test's exemption-present and supplied-field assertions together with AC-RPC-010's fixture, whose release-targeted draft must stay unselected.
- AC-RPC-013 vs "REST without mapping": a mutant fetching snake_case passes field-support (REST has no `--json` field list) but makes every predicate null. Blocked by the supplied-under-read-names assertion, and by AC-RPC-010's fixture being keyed as the fetch delivers — a snake_case fixture with the camelCase filter selects nothing, which fails "exactly `stale-draft`".
- AC-RPC-007 vs "step ceiling equals job": the strictly-below assertion rejects `timeout-minutes: 20` on a browser phase; on the release step, T + 5 + 1 ≤ 40 rejects any T above 34 (control: T = 38). Because the allowances are test constants citing EV-15, a mutant that declares a smaller pre-step in a workflow comment cannot widen the bound.
- AC-RPC-007 vs "evidence inside the budget": a mutant that folds the upload into the race step, or drops `if: always()` from compress or upload, fails the separate-always()-step assertion (control: `if: always()` removed).
- AC-RPC-003 vs "run on failure, publish its success": a mutant giving the real job a fail-open condition without a final verdict step passes the "job runs when classification fails" check but publishes `success` in the counter-case cell — rejected by (iii).
- AC-RPC-003 vs "marker covers the failure": a marker that passes when `detect` did not conclude `success` (for example one whose step does not read the classification's result) publishes `success` under classification failure — rejected by (iii)'s "the marker also runs and concludes `failure`".
- AC-RPC-003 vs "marker skipped on classification failure": a marker whose condition requires a successful classification is skipped when `detect` fails, leaving a (`failure`, `skipped`) pair under one required name whose resolution this SPEC does not rely on — rejected by (iii)'s "the marker also runs".
- AC-RPC-003 vs "always() job condition": a cost job or marker whose condition uses `always()` starts, and keeps running, after the run is cancelled — rejected by (iv) (control: one `!cancelled()` replaced by `always()`).
- AC-RPC-003 vs "verdict step without fail-open": a verdict step on a job that is still skipped when `detect` fails leaves the name unreported — rejected by (iii)'s "the real job runs" and by AC-RPC-004's reported-name assertion for the classification-failure outcome.
- AC-RPC-004 vs "gate every event": a mutant gating push-to-main too fails the push-to-main × docs-only cell for `release_gate` names unless a classification-aware surface publishes them (Q8 applied default).
- AC-RPC-004 vs "event-blind assertion": a guard that checked only a union of names across events would pass control (b); the per-event-class assertion is what turns it red.
- AC-RPC-002 vs "nothing-matching gate": an `if: false` mutant passes a "gate exists" check; blocked by the fail-open and matched-class-runs assertions (two-direction pairing).

## §D.1 Severity and traceability

Count, live identifiers only: release-blocking = 002, 003, 007, 008, 009, 010, 011, 012, 013 (9); regression-guard = 001, 004 (2); total live = 11. Retired: AC-RPC-005, AC-RPC-006 (moved, not reused). Every live REQ-RPC maps to exactly one live AC; no orphan. Traceability chain: card item # → M# (plan.md §F) → REQ-RPC-### (spec.md §C) → AC-RPC-### (here) → EV-# / guard test.

## §D.2 Indirect verification (workflow-level, post-merge)

- After M2 lands: open a docs-only diff PR; observe zero cost-job runs and every required name reported (real job or marker). Push a docs-only commit to a `ci/**` candidate; observe Lint and the five Build names run for real (Q8).
- After M5 lands: `workflow_dispatch` the cleanup with `dry_run: true`; observe the fetch step exit 0 on the runner's gh (closes the EV-10 runner Gap), a computed stale list, and ONLY `[DRY-RUN] Would delete:` lines — no `gh release delete` executes.

## §D.3 Closure gates (Definition of Done)

1. All guard tests of M2, M4, M5 green with RED outputs captured (E8) and committed, each GREEN run's swept count read.
2. EV-1 re-run green at the final run-phase HEAD.
3. `moai spec lint --spec SPEC-RELEASE-PIPELINE-001` clean; golangci-lint clean; both builds green.
4. PRESERVE list (plan.md §D, 12 items) re-verified item by item at the final HEAD.
5. `TestRequiredChecks_ReleaseGateSubsetOfMain` and `TestReleaseVerifyShardParity` green at the final HEAD.

## §D.4 Forward-looking checks

- The input-class root set cites, per root, the job whose verdict reads it — a future input with no cited verdict owner must not be added.
- `TestCleanupFetchFieldsSupported` names the gh field sets it pins; a runner gh upgrade that drops a field turns the post-merge dispatch red before any deletion runs (continued-firing answer, `verification-completeness.md` §1.3).
