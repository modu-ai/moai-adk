---
name: moai-workflow-ci-watch
description: CI watch loop skill — monitors gh pr checks after /moai sync PR creation, classifies required vs auxiliary failures, emits ready-to-merge handoff or T3 expert-debug trigger. HARD invocation contract in .claude/rules/moai/workflow/ci-watch-protocol.md.
version: "1.0.0"
tools: Bash,Read
level1_tokens: 120
level2_tokens: 4800
triggers:
  - "/moai sync.*PR.*created"
  - "moai pr watch"
  - "ci watch"
  - "check.*status.*PR"
paths:
  - ".claude/rules/moai/workflow/ci-watch-protocol.md"
---

# CI Watch Loop (`moai-workflow-ci-watch`)

## Quick Reference

Automatically invoked after `/moai sync` creates a PR. Polls `gh pr checks`
every 30 seconds, classifies required vs auxiliary failures using the SSoT at
`.github/required-checks.yml`, and presents a ready-to-merge decision report.

**Trigger**: After `/moai sync` Phase 4 PR-create step completes.
**Exit paths**: all-pass (ready-to-merge) | required-fail (T3 handoff) | 30-min timeout.

### One-liner

```bash
# Orchestrator invokes via Bash tool after gh pr create returns PR number.
MOAI_CIWATCH_GH=gh sh scripts/ci-watch/run.sh <PR_NUMBER> <BRANCH>
```

### State Machine

```
[idle] → [arm-watch] → [watching] → ┬─→ [all-required-pass] → emit ready-to-merge
                                    ├─→ [required-fail]      → JSON handoff to T3
                                    ├─→ [auxiliary-fail-only] → emit advisory → continue
                                    └─→ [timeout-30m]         → emit blocker
```

### Required vs Auxiliary

- **Required** checks: defined in `.github/required-checks.yml` `branches.<pattern>.contexts`
- **Auxiliary** checks: defined in `.github/required-checks.yml` `auxiliary` list
- Auxiliary failures emit advisory messages but do NOT block ready-to-merge

Detailed Reference: `modules/ci-watch-protocol.md`

---

## Implementation Guide

### Orchestrator-Side Protocol

When `/moai sync` Phase 4 returns a PR number, the orchestrator MUST:

1. Start the watch loop via Bash tool:
   ```bash
   MOAI_CIWATCH_GH=gh sh scripts/ci-watch/run.sh <PR_NUMBER> <BRANCH>
   ```

2. Monitor the exit code:
   - **Exit 0**: All required checks passed → present ready-to-merge report via AskUserQuestion
   - **Exit 2**: Required failure detected → parse JSON from stdout → invoke T3 auto-fix loop
   - **Exit 3**: 30-min hard timeout → emit blocker message, return control to user

3. On exit 0, emit ready-to-merge report:
   ```bash
   # Go CLI generates the markdown report; orchestrator reads and presents via AskUserQuestion.
   moai pr watch --report <PR_NUMBER> --branch <BRANCH>
   ```

4. On exit 2, pipe JSON handoff to Wave 3 expert-debug:
   ```bash
   # stdout contains: {"prNumber":N,"branch":"...","failedChecks":[...],"auxiliaryFailCount":N}
   # Orchestrator injects this JSON into expert-debug spawn prompt.
   ```

### State File Management

The watch loop maintains `.moai/state/ci-watch-active.flag` (YAML):

```yaml
pr_number: 785
started_at: "2026-05-06T08:30:00Z"
heartbeat_at: "2026-05-06T08:35:00Z"
required_checks: [Lint, "Test (ubuntu-latest)", ...]
abort_requested: false
```

**Abort** an active watch:
```bash
moai pr watch --abort
```

**Staleness detection**: If `heartbeat_at` is older than 90 seconds, the state
is considered stale (crashed watch) and a new invocation may take over.

### Go Helpers (`internal/ciwatch/`)

| Package | File | Purpose |
|---------|------|---------|
| ciwatch | `classifier.go` | `IsRequired(check, branch)` via SSoT YAML |
| ciwatch | `handoff.go` | `CIState`, `CheckResult`, `Handoff`, `NewHandoff()`, `FormatStatusUpdate()` |
| ciwatch | `state.go` | `WatchState`, `ReadState()`, `WriteState()`, `Touch()`, `SetAbortFlag()` |
| cli/pr | `watch.go` | `EmitReadyToMergeReport()`, `EmitFailureHandoff()` |

### Shell Helpers (`scripts/ci-watch/`)

| File | Purpose |
|------|---------|
| `run.sh` | Main polling loop (mock-injectable via `MOAI_CIWATCH_GH`) |
| `lib/_common.sh` | `log_step()`, `abort()`, `posix_now()` |
| `lib/classify.sh` | `is_required()`, `is_auxiliary()` (yq + grep fallback) |
| `lib/timeout.sh` | `ciwatch_start_timer()`, `ciwatch_check_timeout()` |

Detailed Reference: `modules/trigger-handoff.md`

---

## Advanced Patterns

### Abort and Recover

If the watch loop is interrupted (Ctrl-C, session end):
1. Heartbeat becomes stale after 90s
2. Next invocation auto-reclaims (no explicit cleanup needed)
3. For immediate reclaim: `moai pr watch --abort`

### gh CLI Compatibility

Requires `gh >= 2.50` for the `workflow` JSON field. Version check:
```bash
gh --version | grep -E '^gh version ([2-9]|[1-9][0-9])\.([5-9][0-9]|[6-9][0-9])'
```

On older `gh`, `workflow` field is absent; `lib/classify.sh` falls back to
name-based heuristics using the `required:` list from the SSoT directly.

### Concurrent Invocation Guard

Only one watch-per-repo at a time. If a second `moai pr watch` invocation
detects a fresh heartbeat (< 90s), it exits with:
```
[ci-watch][FATAL] watch already active for PR=785; run `moai pr watch --abort` to release
```

### Status Report Format

Every 30-second tick emits to stderr:
```
[ci-watch] PR #785: required 4/6 pass, 2 pending; advisory 0 fail
```

- No ANSI escape codes (orchestrator chat transport)
- Max 200 chars per line
- State-change ticks only (no repeat if unchanged)

### Wave 3 Handoff Schema

On exit 2, stdout contains stable JSON for `expert-debug` consumption:

```json
{
  "prNumber": 785,
  "branch": "feat/my-feature",
  "failedChecks": [
    {
      "name": "Lint",
      "runId": "12345678",
      "logUrl": "https://github.com/.../actions/runs/12345678",
      "conclusionDetail": ""
    }
  ],
  "auxiliaryFailCount": 1,
  "totalRequired": 6
}
```

Field stability: `name`, `runId`, `logUrl` are stable for Wave 3. Do not rename.

---

## Works Well With

- `moai-workflow-ci-watch` → `expert-debug` (Wave 3 auto-fix loop)
- `.github/required-checks.yml` (Wave 1 SSoT)
- `scripts/ci-mirror/run.sh` (Wave 1 local CI mirror)
- `.claude/rules/moai/workflow/ci-watch-protocol.md` (HARD invocation contract)

<!-- moai:evolvable-start id="rationalizations" -->
## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "I'll skip the watch loop for a small PR" | Small PRs fail CI too. The loop costs nothing and saves manual polling. |
| "Auxiliary failures should block merge" | Auxiliary checks are defined as advisory in the SSoT. Changing this requires editing required-checks.yml, not the skill. |
| "30 min is too long" | Extend via `CIWATCH_TIMEOUT_SECONDS` env if needed. Default covers tail-end CI runs. |
<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="red-flags" -->
## Red Flags

- Orchestrator calls AskUserQuestion inside the watch loop (HARD: orchestrator-only, not CLI)
- State file path hardcoded outside `ciwatch.StateFile` constant
- Watch loop polling interval < 30s (GitHub API rate-limit risk)
- required-checks.yml modified to add checks without updating branch protection via `moai github init`
<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="verification" -->
## Verification

- [ ] `bash scripts/ci-watch/test/run_test.sh` passes all 9 shell tests
- [ ] `go test ./internal/ciwatch/... ./internal/cli/pr/... -race` passes
- [ ] `internal/ciwatch/` coverage >= 85%
- [ ] No ANSI codes in `FormatStatusUpdate()` output
- [ ] `EmitReadyToMergeReport` contains `(권장)` on first option
- [ ] CLI does NOT call AskUserQuestion (orchestrator-only)
<!-- moai:evolvable-end -->
