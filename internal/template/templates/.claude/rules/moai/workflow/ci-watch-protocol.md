---
description: CI watch loop protocol — HARD invocation contract for moai-workflow-ci-loop skill (watch phase). Auto-loaded on /moai sync and moai pr watch invocations.
paths:
  - ".claude/skills/moai-workflow-ci-loop/SKILL.md"
  - "scripts/ci-watch/run.sh"
---

# CI Watch Protocol Rule

> This file is the single source of truth for CI watch loop invocation rules.
> Cross-referenced by: SKILL.md, moai-workflow-ci-loop (unified watch + autofix skill).

---

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
2. `.github/required-checks.yml` exists (Wave 1 SSoT)
3. PR number is a positive integer (not zero, not empty)
4. No active watch for a different PR (heartbeat < 90s)

---

## Polling Cadence

[ZONE:Frozen] [HARD] Poll interval MUST be 30 seconds minimum. GitHub Actions API rate limits
apply; polling faster than 30s risks 429 responses.

[WARN] `CIWATCH_POLL_INTERVAL` env var overrides the interval. Do not set below
30 in production. Test mode uses `MOAI_CIWATCH_NO_SLEEP=1` (single-tick exit).

---

## 30-Minute Hard Timeout

[ZONE:Frozen] [HARD] The watch loop MUST exit with code 3 after 30 minutes wall-clock time
regardless of check states. Token budget guard: a watch loop running indefinitely
would exhaust the orchestrator context window.

Default: `CIWATCH_TIMEOUT_SECONDS=1800` (30 minutes).

On exit 3, the orchestrator MUST surface a blocker message to the user and
return control. Do NOT auto-restart the watch loop after timeout.

---

## Required vs Auxiliary Discrimination

[ZONE:Frozen] [HARD] Required checks are defined ONLY in `.github/required-checks.yml`
`branches.<pattern>.contexts`. Hardcoding check names in scripts is prohibited.

[ZONE:Frozen] [HARD] Auxiliary checks listed under `auxiliary:` in `.github/required-checks.yml`
MUST NOT block the ready-to-merge decision. They are advisory only.

[WARN] If `.github/required-checks.yml` is missing, the watch loop exits 1 with
a remediation message. Run `moai github init` to restore the SSoT.

---

## Exit Code Handling

| Exit | Meaning | Orchestrator MUST |
|------|---------|-------------------|
| 0 | All required checks passed | Present ready-to-merge AskUserQuestion |
| 1 | Fatal error | Surface error + remediation to user |
| 2 | Required check(s) failed | Parse JSON handoff → Wave 3 `Agent(general-purpose)` diagnostic scope (ci-autofix loop entry) |
| 3 | 30-min timeout | Emit blocker → return control to user |

---

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

On exit 2, stdout contains JSON (see `.claude/skills/moai-workflow-ci-loop/SKILL.md` (the **Handoff schema on exit 2** marker)):

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

## Wave 1 SSoT Contract Preservation

[ZONE:Frozen] [HARD] Wave 2 watch loop MUST NOT modify `.github/required-checks.yml` (Wave 1 SSoT).
The SSoT is read-only for Wave 2. Modifications require `moai github init` re-run.

---

## Background watch standardization

[ZONE:Evolvable] [HARD] When the orchestrator monitors CI checks on a long-running PR
(typically 5+ minutes), it MUST use `gh pr checks --watch` invoked via
`run_in_background: true`. Idle polling loops (e.g., `sleep N && gh pr checks`)
are prohibited because they block the orchestrator's main session and waste
both wall-time and tokens. This rule was added in response to a meta-analysis
which found 15 min of serial CI wait during a single run-phase.

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

Version: 1.1.0
Classification: HARD operational rule, applies to all /moai sync workflows
