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
| Author: new or changed code | the full check set on the exact tree | no; a CI run is extra evidence, never a replacement |
| Repair lane: the repair changes code | the full check set on the repair head | no |
| Auditor: the audited head changes code | tests that exercise the diff, mutations, behavioural probes, fresh-clone reproduction, and the full check set | no |
| Delta auditor: merge-only delta (the merge-tree result equals the pushed tree and the lane changed nothing itself) | only the tests that target the files the merge brought in | yes, under the citation conditions; the local full suite is recorded as `not_run` |

Citation conditions — all must hold, or the CI result is not cited:

1. The run's `head_sha` equals the PR head exactly. Read it with `gh api "repos/<owner>/<repo>/actions/runs?head_sha=<head>"`; a `pull_request` run reports the PR head in `head_sha` and builds the merge ref, whose SHA is the run's `GITHUB_SHA`.
2. Every required job of that run is `completed` with conclusion `success` (in this repository: the four `rc-ci / <target>` jobs, the `rc-ci` aggregate and `windows-build`; take the names from the workflow file, not from memory).
3. The run id is written in the verdict and in the PR's `## Self-Audit` section as one line: `ci: rc-ci run=<run id> head=<40-hex head SHA>`. `tools/check_pr_self_audit.py` rejects a `ci:` line without both keyed fields; it cannot see GitHub, so the auditor still reads the run back.

A CI result is never cited when the repair changed code, when no run exists for the head (a `pull_request` run does not start while the PR has a merge conflict; write `ci: not_run` with the reason), or for checks CI does not perform (mutations, behavioural probes, fresh-clone reproduction). Citing never relaxes a gate: the independent audit, `rc-ci` and `windows-build` success, and exact head and base equality stay required.

Heavy checks are serialised: on one machine, at most one full build or full test run executes at a time. Before starting one, check that no other is running; if one is, wait with a bounded poll and record the wait instead of starting a second. Reuse the existing target directory instead of creating a new one, and record free disk space before the run. Numeric load and disk thresholds are not decided yet (HOLD); until they are, a lane reports high load or low disk to the leader instead of choosing a threshold itself.

A test must not depend on wall-clock time. When an assertion depends on elapsed time or a time budget, inject the clock or the budget (a seam such as a `--render-budget-ms` style flag or a clock parameter) and fix it in the test. Auditors treat a direct wall-clock budget in a test assertion as a finding.

## Exceptions and Permissions

A command that was not run is not_run. A missing required guard is blocked before anything runs, and exit code 0 alone does not satisfy every AC.

## Completion Check

Bind the output to the task and the exact snapshot. Never report a verification or external effect that was not performed as complete.
