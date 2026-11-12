# VERIFY

## Input

An exact tree, the SPEC and policy digests, and a check inventory.

## Output

A receipt (returned on stdout by the internal `run-verify` operation; not stored in 2026.11.0-rc.1) with each check marked passed, failed, not_run, blocked, stale, or unknown.

## Procedure

1. Confirm the verification commands and environment the repository actually uses.
2. Before running anything, confirm permissions and the required guards.
3. Run the commands through a trusted runner, and record start, end, exit code, cancellation, and timeout.
4. Bind the task, attempt, tree, SPEC, policy, and tool version to the raw evidence.
5. State skips, coverage gaps, and artifact expiry explicitly.

## Check-run rules

This section is the single source for who runs which check locally and when an exact-head CI result may be cited instead (#2385). AUDIT, SYNC and the PR form point here and do not restate it. The integration queue, merge order and open-PR cap are the integration-flow rules (#2388); build-artifact sharing is #2380.

| Role and change | Run locally | May cite CI for the full Rust suite |
| --- | --- | --- |
| Author: new or changed code | the full check set of every language and suite the change touches, on the exact tree | no for what the change touches; a CI run there is extra evidence, never a replacement |
| Author: a language or suite the change does not touch (for example the Rust suite of a PR that changes only Python and documents) | nothing for that suite | yes, under the citation conditions; the local run of that suite is recorded as `not_run` (leader decision 2026-11-10 on the #2395 question) |
| Repair lane: the repair changes code | the full check set on the repair head | no |
| Auditor: the audited head changes code | tests that exercise the diff, mutations, behavioural probes, fresh-clone reproduction, and the full check set | no |
| Delta auditor: merge-only delta (the merge-tree result equals the pushed tree and the lane changed nothing itself) | only the tests that target the files the merge brought in | yes, under the citation conditions; the local full suite is recorded as `not_run` |
| Author: a skill package file changes | the file's sha256 entry updated in the package's source manifest in the same commit, then `cargo test --offline --locked --test product_cli skill_package::` (the package's manifest and route checks) | no for the package tests; the rest of the Rust suite follows the rows above |

Citation conditions — all must hold, or the CI result is not cited:

1. The run's `head_sha` equals the PR head exactly. Read it with `gh api "repos/<owner>/<repo>/actions/runs?head_sha=<head>"`; a `pull_request` run reports the PR head in `head_sha` and builds the merge ref, whose SHA is the run's `GITHUB_SHA`.
2. Every required job of that run is `completed` with conclusion `success` (in this repository a pull request run has `rc-ci / linux-x86_64` and the `rc-ci` aggregate; take the names from the workflow file, not from memory).
3. The run id is written in the verdict and in the PR's `## Self-Audit` section as one line: `ci: rc-ci run=<run id> head=<40-hex head SHA>`. `tools/check_pr_self_audit.py` rejects a `ci:` line without both keyed fields; it cannot see GitHub, so the auditor still reads the run back.

Where each target runs (user decision U10, Issue #2554): a pull request to the integration branch runs `rc-ci / linux-x86_64` only; the push that merges it runs all four `rc-ci / <target>` jobs and `windows-build` with the same check set. The scope moves to after merge, it does not shrink. When a post-merge run fails, the leader's first act — before any other merge — is to revert the pull request that caused it or to have it fixed, and the failing run id and the reverted or fixing PR are recorded on that PR's Issue.

A CI result is never cited for a language or suite the change touches, when the repair changed code in it, when no run exists for the head (a `pull_request` run does not start while the PR has a merge conflict; write `ci: not_run` with the reason), or for checks CI does not perform (mutations, behavioural probes, fresh-clone reproduction). Citing never relaxes a gate: the independent audit, `rc-ci` success with no failed `windows-build` on the head, and exact head and base equality stay required.

Heavy checks are serialised: on one machine, at most one full build or full test run executes at a time. Before starting one, check that no other is running; if one is, wait with a bounded poll and record the wait instead of starting a second. Use one build folder per worktree and never share `CARGO_TARGET_DIR` between worktrees (a shared folder once ran another tree's binary without a rebuild): reuse the worktree's own folder, confirm the `Compiling` line names this worktree, and record free disk space before the run. Build-folder cleanup follows the leader's canonical wording (Issue #2385 comment 6777290466; user decision, Issue #2389 comment 6774899518): (1) when the audits and tests of its own worktree are done, a lane removes only the cargo output it made itself — the `debug`, `release` and `<target triple>` folders in `target/` or in a cargo folder directly under it; (2) it never removes `target/` itself or the notes, evidence and scripts under `target/<name>/`, and it never touches another lane's worktree; (3) cleaning several merged worktrees at once is done by a script the leader writes after checking the list and the operator runs, never by a lane; (4) a worktree itself is removed only after a separate confirmation. Numeric load and disk thresholds are not decided yet (HOLD); until they are, a lane reports high load or low disk to the leader instead of choosing a threshold itself.

A test must not depend on wall-clock time. When an assertion depends on elapsed time or a time budget, inject the clock or the budget (a seam such as a `--render-budget-ms` style flag or a clock parameter) and fix it in the test. Auditors treat a direct wall-clock budget in a test assertion as a finding.

A child-process budget in a test is load-independent: bound the child by progress — an invocation count or a log signal that resets the deadline — plus a hard hang ceiling and an invocation cap, never a bare wall-clock timeout (`ci_install_ripgrep_test` #2469, `mutate_test` #2502). The budget numbers are pinned by tests, and wall-clock appears only as the generous hang guard.

Docs-only PRs get light CI (#2527, design comment 6822993459): when every changed path is under a verified-safe doc root the heavy steps (PyYAML install, ripgrep install, rust toolchain setup, build/test) are skipped and only `git diff --check` runs, but the aggregate `rc-ci` check, the `ci: rc-ci run=<id> head=<sha>` line contract and every citation condition above are unchanged — a light run still counts as the required CI result for that head. The scope classifier is `tools/ci_scope.py`; its verified-safe list is `docs/` with `.md`/`.txt`/`.png`/`.svg` suffixes (minus files tests read) plus `.moai-cli/reports/`. Any failure or unrecognized path fails safe to the full scope.

## Exceptions and Permissions

A command that was not run is not_run. A missing required guard is blocked before anything runs, and exit code 0 alone does not satisfy every AC. A test-name filter applies to every target named in the same `cargo test` call, so filter one target per call; a result that ran zero tests (`0 passed`) is not_run, never a pass.

## Completion Check

Bind the output to the task and the exact snapshot. Never report a verification or external effect that was not performed as complete.
