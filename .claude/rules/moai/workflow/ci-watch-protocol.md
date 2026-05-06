---
description: CI watch loop protocol — HARD invocation contract for moai-workflow-ci-watch skill. Auto-loaded on /moai sync and moai pr watch invocations.
paths:
  - ".claude/skills/moai-workflow-ci-watch/SKILL.md"
  - "scripts/ci-watch/run.sh"
---

# CI Watch Protocol Rule

> Source: SPEC-V3R3-CI-AUTONOMY-001 Wave 2 (T2)
> This file is the single source of truth for CI watch loop invocation rules.
> Cross-referenced by: SKILL.md, moai-workflow-ci-watch modules.

---

## Auto-Invocation Contract

[HARD] The orchestrator MUST invoke the CI watch loop after `/moai sync` Phase 4
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

[HARD] Poll interval MUST be 30 seconds minimum. GitHub Actions API rate limits
apply; polling faster than 30s risks 429 responses.

[WARN] `CIWATCH_POLL_INTERVAL` env var overrides the interval. Do not set below
30 in production. Test mode uses `MOAI_CIWATCH_NO_SLEEP=1` (single-tick exit).

---

## 30-Minute Hard Timeout

[HARD] The watch loop MUST exit with code 3 after 30 minutes wall-clock time
regardless of check states. Token budget guard: a watch loop running indefinitely
would exhaust the orchestrator context window.

Default: `CIWATCH_TIMEOUT_SECONDS=1800` (30 minutes).

On exit 3, the orchestrator MUST surface a blocker message to the user and
return control. Do NOT auto-restart the watch loop after timeout.

---

## Required vs Auxiliary Discrimination

[HARD] Required checks are defined ONLY in `.github/required-checks.yml`
`branches.<pattern>.contexts`. Hardcoding check names in scripts is prohibited.

[HARD] Auxiliary checks listed under `auxiliary:` in `.github/required-checks.yml`
MUST NOT block the ready-to-merge decision. They are advisory only.

[WARN] If `.github/required-checks.yml` is missing, the watch loop exits 1 with
a remediation message. Run `moai github init` to restore the SSoT.

---

## Exit Code Handling

| Exit | Meaning | Orchestrator MUST |
|------|---------|-------------------|
| 0 | All required checks passed | Present ready-to-merge AskUserQuestion |
| 1 | Fatal error | Surface error + remediation to user |
| 2 | Required check(s) failed | Parse JSON handoff → Wave 3 expert-debug |
| 3 | 30-min timeout | Emit blocker → return control to user |

---

## AskUserQuestion Boundary

[HARD] The CLI (`moai pr watch`, `EmitReadyToMergeReport`) MUST NOT call
AskUserQuestion. Interaction is strictly orchestrator territory.

The orchestrator presents the emitted markdown report via AskUserQuestion:
- Option 1 (권장): Merge PR
- Option 2: Hold
- Option 3: Investigate

The `(권장)` label MUST be on the first option per `.claude/rules/moai/core/askuser-protocol.md`.

---

## T3 Handoff Format

On exit 2, stdout contains JSON (see `modules/trigger-handoff.md` in the skill):

```json
{
  "prNumber": 785,
  "branch": "feat/...",
  "failedChecks": [{"name": "Lint", "runId": "...", "logUrl": "..."}],
  "auxiliaryFailCount": 0,
  "totalRequired": 6
}
```

[HARD] Only required failures appear in `failedChecks`. Auxiliary failures are
counted in `auxiliaryFailCount` but MUST NOT be passed to expert-debug as
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

[HARD] Wave 2 watch loop MUST NOT modify `.github/required-checks.yml` (Wave 1 SSoT).
The SSoT is read-only for Wave 2. Modifications require `moai github init` re-run.

---

Version: 1.0.0
Source: SPEC-V3R3-CI-AUTONOMY-001 Wave 2 (2026-05-06)
Classification: HARD operational rule, applies to all /moai sync workflows
