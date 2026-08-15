---
description: CI watch loop protocol — HARD invocation contract for hns-workflow-ci-loop skill (watch phase). Auto-loaded on /moai sync and moai pr watch invocations.
paths: ".claude/skills/hns-workflow-ci-loop/SKILL.md,scripts/ci-watch/run.sh"
---

# CI Watch Protocol Rule

> This file is the single source of truth for CI watch loop invocation rules.
> Cross-referenced by: SKILL.md, hns-workflow-ci-loop (unified watch + autofix skill).

---

<!-- anchor: #watch-loop-entry -->
## Auto-Invocation Contract

[ZONE:Frozen] [HARD] The orchestrator MUST invoke the CI watch loop after `/moai sync` Phase 4
(PR creation) completes successfully and returns a PR number.

```
/moai sync → Phase 4 (gh pr create returns PR_NUMBER) → invoke ci-watch loop
```

**Invocation command** (Bash tool):
```bash
MOAI_CIWATCH_GH=gh sh scripts/ci-watch/run.sh <PR_NUMBER> <BRANCH>
```

**Prerequisites** (all must be satisfied before invocation):
1. `gh` CLI is authenticated (`gh auth status` exits 0)
2. `.github/required-checks.yml` exists (required-checks SSoT layer)
3. PR number is a positive integer (not zero, not empty)
4. No active watch for a different PR (heartbeat < 90s)

---

<!-- anchor: #poll-interval -->
## Polling Cadence

[ZONE:Frozen] [HARD] Poll interval MUST be 30 seconds minimum. GitHub Actions API rate limits
apply; polling faster than 30s risks 429 responses.

[WARN] `CIWATCH_POLL_INTERVAL` env var overrides the interval. Do not set below
30 in production. Test mode uses `MOAI_CIWATCH_NO_SLEEP=1` (single-tick exit).

---

<!-- anchor: #timeout -->
## 30-Minute Hard Timeout

[ZONE:Frozen] [HARD] The watch loop MUST exit with code 3 after 30 minutes wall-clock time
regardless of check states. Token budget guard: a watch loop running indefinitely
would exhaust the orchestrator context window.

Default: `CIWATCH_TIMEOUT_SECONDS=1800` (30 minutes).

On exit 3, the orchestrator MUST surface a blocker message to the user and
return control. Do NOT auto-restart the watch loop after timeout.

---

<!-- anchor: #required-checks-ssot -->
## Required vs Auxiliary Discrimination

[ZONE:Frozen] [HARD] Required checks are defined ONLY in `.github/required-checks.yml`
`branches.<pattern>.contexts`. Hardcoding check names in scripts is prohibited.

[ZONE:Frozen] [HARD] Auxiliary checks listed under `auxiliary:` in `.github/required-checks.yml`
MUST NOT block the ready-to-merge decision. They are advisory only.

[WARN] If `.github/required-checks.yml` is missing, the watch loop exits 1 with
a remediation message. Run `moai github init` to restore the SSoT.

---

<!-- anchor: #failed-checks-reporting -->
## Exit Code Handling

| Exit | Meaning | Orchestrator MUST |
|------|---------|-------------------|
| 0 | All required checks passed | Present ready-to-merge AskUserQuestion |
| 1 | Fatal error | Surface error + remediation to user |
| 2 | Required check(s) failed | Parse JSON handoff → autofix layer `Agent(general-purpose)` diagnostic scope (ci-autofix loop entry) |
| 3 | 30-min timeout | Emit blocker → return control to user |

---

<!-- anchor: #emit-ready-to-merge-report -->
## AskUserQuestion Boundary

[ZONE:Frozen] [HARD] The CLI (`moai pr watch`, `EmitReadyToMergeReport`) MUST NOT call
AskUserQuestion. Interaction is strictly orchestrator territory.

The orchestrator presents the emitted markdown report via AskUserQuestion:
- Option 1 (권장): Merge PR
- Option 2: Hold
- Option 3: Investigate

The `(권장)` label MUST be on the first option per `.claude/rules/moai/core/askuser-protocol.md`.

---

## T3 Handoff Format

On exit 2, stdout contains JSON (see `.claude/skills/hns-workflow-ci-loop/SKILL.md` (the **Handoff schema on exit 2** marker)):

```json
{
  "prNumber": 785,
  "branch": "feat/...",
  "failedChecks": [{"name": "Lint", "runId": "...", "logUrl": "..."}],
  "auxiliaryFailCount": 0,
  "totalRequired": 6
}
```

[ZONE:Frozen] [HARD] Only required failures appear in `failedChecks`. Auxiliary failures are
counted in `auxiliaryFailCount` but MUST NOT be passed to the diagnostic scope as
blocking failures.

---

## Abort Protocol

[WARN] If user requests abort mid-watch: `moai pr watch --abort` sets
`abort_requested: true` in `.moai/state/ci-watch-active.flag`. The loop polls
this flag at the start of each tick and exits cleanly (exit 0).

Heartbeat staleness reclaim: if `heartbeat_at` is older than 90 seconds,
a new invocation may take over without explicit abort.

---

## Required-Checks SSoT Contract Preservation

[ZONE:Frozen] [HARD] The watch loop layer MUST NOT modify `.github/required-checks.yml` (the required-checks SSoT layer).
The SSoT is read-only for the watch layer. Modifications require `moai github init` re-run.

---

## Background watch standardization

[ZONE:Evolvable] [HARD] When the orchestrator monitors CI checks on a long-running PR
(typically 5+ minutes), it MUST use `gh pr checks --watch` invoked via
`run_in_background: true`. Idle polling loops (e.g., `sleep N && gh pr checks`)
are prohibited because they block the orchestrator's main session and waste
both wall-time and tokens. Motivation: reduces serial CI wait during long-running PRs.

### Canonical Pattern

```bash
# Background watch — returns immediately, the orchestrator continues other work.
gh pr checks <PR> --watch
```

Invoked via the Bash tool with `run_in_background: true`. The background task
emits a notification when checks resolve, at which point the orchestrator
foreground-polls `BashOutput` to retrieve the final state.

### Anti-pattern: Sleep + Poll

```bash
# PROHIBITED — blocks the orchestrator's main turn for N seconds.
sleep 60 && gh pr checks <PR>
```

The `sleep N && check` idiom locks the orchestrator into an idle wait. Use
`gh pr checks --watch` with `run_in_background: true` instead, then poll
`BashOutput` only when other productive work runs out.

### Notification Pattern (foreground recovery)

If `gh pr checks --watch` hangs beyond a reasonable threshold (typically 20+ min),
the orchestrator MAY foreground-poll:

```bash
gh pr checks <PR> --json name,state,conclusion | jq '.[] | select(.state != "COMPLETED")'
```

This is a fallback. The default path is background watch + concurrent productive
work in the orchestrator's main turn.

### When NOT to Background-Watch

- Pre-merge final check (synchronous gate at the end of /moai sync): use the
  blocking ci-watch loop CLI (`scripts/ci-watch/run.sh`) instead of `--watch`.
- Test/CI fixtures that must observe state immediately: use synchronous
  `gh pr checks` without `--watch`.

Cross-reference: the canonical CI-watch acceptance criterion (recorded in
the predecessor workflow optimization rule) verifies this section contains
both `gh pr checks --watch` and `run_in_background: true` literals.

---

## Merge Convergence Under a Strict Base

Applies when the base branch requires branches to be up to date before merging
(GitHub branch protection `required_status_checks.strict: true`). Under that
setting a merge into the base makes every other open pull request `BEHIND`, and
bringing one up to date **restarts its entire check suite from zero**.

Two failure modes follow, and both are self-inflicted.

### Failure mode 1 — update-restart livelock

Updating a branch whose own checks are still running discards their progress and
starts them again. Repeat that while waiting for a long-tail check — a
cross-platform test matrix routinely runs ten minutes or more — and the check
can never complete. The pull request appears permanently `pending`, and the loop
looks like slow CI rather than the caller resetting it.

[ZONE:Evolvable] [HARD] Do NOT update a branch while any of its own required
checks is `pending`. Update only when the check set has settled — every required
check `pass` or `fail`, none `pending` — AND merge state is `BEHIND`.

### Failure mode 2 — parallel update convergence collapse

Updating several pull requests in the same round makes them all up to date at
once. The first to merge returns every other one to `BEHIND`, and each of those
restarts a full check suite. For N pull requests against a base whose slowest
check takes T, updating in parallel costs on the order of N²·T; serializing
costs N·T.

[ZONE:Evolvable] [HARD] Drive **one** pull request to merge at a time. Do not
update a second branch until the first has merged or has been abandoned.

### Procedure

```bash
# 1. Enable auto-merge once per PR. GitHub merges each one as soon as its
#    requirements are met, so no polling is needed to complete the merge.
gh pr merge <PR> --squash --auto

# 2. Read state without mutating it.
gh pr view <PR> --json state,mergeStateStatus --jq '"\(.state) \(.mergeStateStatus)"'
gh pr checks <PR> --json name,bucket

# 3. Update ONLY when checks have settled and the branch is BEHIND.
gh pr update-branch <PR>
```

Between steps, follow § Background watch standardization — `gh pr checks --watch`
under `run_in_background: true`, never a foreground `sleep N` loop. The
sleep-and-poll anti-pattern is what turns a slow merge into an hour of blocked
wall-time, and it compounds this section's failure modes by hiding them behind
apparent progress.

### Bounded retries

[ZONE:Evolvable] [HARD] Allow at most **3** update cycles per pull request.
Exceeding that means the base is advancing faster than the check suite
completes, which no amount of retrying resolves. Stop, report the situation, and
let the user choose: pause the other merge sources, adopt a merge queue, batch
the changes into one pull request, or accept a longer settle window.

Never leave an unbounded update-or-poll loop running. A loop with no exit
condition other than success cannot report the case where success is
unreachable.

### Anti-patterns

```bash
# PROHIBITED — updates while the PR's own checks are still running.
for pr in 1 2 3; do gh pr update-branch $pr; done; sleep 60
```

- Updating a branch on a `pending` check set — resets the work being waited on.
- Updating several pull requests in one round under a strict base — guarantees
  that all but one are immediately invalidated again.
- Treating `mergeStateStatus: UNKNOWN` as a failure. It means GitHub has not
  finished computing mergeability; re-read it rather than acting on it.
- Foreground `sleep N` polling of a check that takes minutes.

---

Version: 1.2.0
Classification: HARD operational rule, applies to all /moai sync workflows
