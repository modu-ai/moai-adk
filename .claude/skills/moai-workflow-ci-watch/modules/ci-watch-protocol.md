# CI Watch Protocol

Detailed protocol for orchestrator invocation of the CI watch loop.

## When to Invoke

[HARD] The orchestrator MUST invoke the CI watch loop after `/moai sync` Phase 4
(PR creation) completes and returns a PR number.

**Trigger conditions** (any one activates):
1. `/moai sync` Phase 4 returns PR number from `gh pr create`
2. User explicitly requests `moai pr watch <PR_NUMBER>`
3. Skill trigger keyword detected: "watch ci", "monitor pr", "check status"

**Do NOT invoke** when:
- No PR was created (sync exited early)
- PR is already in merged/closed state
- Watch loop is already active for the same PR (check heartbeat)

## Invocation Contract

```bash
# Primary invocation (orchestrator Bash tool call):
MOAI_CIWATCH_GH=gh \
MOAI_CIWATCH_REQUIRED_CHECKS_FILE=.github/required-checks.yml \
sh scripts/ci-watch/run.sh <PR_NUMBER> <BRANCH>
```

**Required prerequisites**:
- `gh` CLI authenticated (`gh auth status`)
- `.github/required-checks.yml` exists (created by Wave 1 `moai github init`)
- PR number is a positive integer
- State file is not active for a different PR (check `ciwatch.ReadState`)

## Exit Code Contract

| Exit | Meaning | Orchestrator Action |
|------|---------|---------------------|
| 0 | All required checks passed | Call `EmitReadyToMergeReport`, present via AskUserQuestion |
| 1 | Fatal error (gh auth, PR not found, SSoT missing) | Surface error to user; suggest remediation |
| 2 | Required check(s) failed | Parse JSON from stdout; invoke Wave 3 expert-debug |
| 3 | 30-minute hard timeout | Emit blocker; return control to user; suggest manual investigation |

## 30-Second Polling Cadence

Each tick:
1. `gh pr checks <PR> --json name,status,conclusion,detailsUrl`
2. Classify each check via `.github/required-checks.yml` SSoT
3. Compute aggregate state (required_pass, required_fail, required_pending, aux_fail)
4. Emit status line to stderr (no ANSI, max 200 chars)
5. Touch `.moai/state/ci-watch-active.flag` heartbeat
6. Evaluate state machine transition (see SKILL.md Quick Reference)
7. Sleep 30s (skip in `MOAI_CIWATCH_NO_SLEEP=1` test mode)

## Abort Protocol

```bash
# Set abort flag — watch loop polls and exits cleanly within 30s.
moai pr watch --abort

# Verify state cleared:
ls -la .moai/state/ci-watch-active.flag 2>/dev/null || echo "cleared"
```

Abort flag is written via `ciwatch.SetAbortFlag()` (atomic rename).
The watch loop polls `AbortRequested` at the start of each tick.

## Heartbeat Staleness

| Threshold | Action |
|-----------|--------|
| < 90s | Watch is active — new invocation must abort or wait |
| >= 90s | Watch is stale (crashed) — new invocation may take over |

Staleness check: `ciwatch.ReadState(path).IsStale(90 * time.Second)`

## Required-Checks SSoT Integration

The watch loop reads `.github/required-checks.yml` (Wave 1 artifact):

- `branches.<pattern>.contexts` — required checks for branch patterns
- `auxiliary` — advisory-only checks (never block merge)

Branch pattern matching: exact match + `filepath.Match` glob (supports `release/*`).

**Do NOT hardcode check names** in watch loop scripts. Always read from SSoT.

## AskUserQuestion Boundary

[HARD] The CLI (`moai pr watch`, `EmitReadyToMergeReport`) MUST NOT call
AskUserQuestion. The orchestrator presents the emitted markdown report via
AskUserQuestion with these options:

1. **(권장)** Merge PR — all required checks pass
2. Hold — keep PR open
3. Investigate advisory failures

The "(권장)" label MUST be on the first option per AskUserQuestion protocol.
See `.claude/rules/moai/core/askuser-protocol.md`.
