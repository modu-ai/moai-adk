---
name: moai-workflow-ci-loop
description: >
  Unified CI watch + auto-fix loop skill. Polls gh pr checks after /moai sync PR creation,
  classifies required vs auxiliary failures, attempts safe automated patches (max 3 iterations),
  and escalates semantic failures to the user. Use for CI loop workflow — NOT for general
  loop iteration patterns (see moai-workflow-loop).

when_to_use: >
  Use for the CI watch and auto-fix loop after /moai sync PR creation:
  polling gh pr checks, classifying required vs auxiliary failures, safe
  automated patch attempts (max 3 iterations), and escalation of semantic
  failures.

license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Bash, Read
user-invocable: false
metadata:
  version: "0.1.0"
  category: "workflow"
  status: "active"
  updated: "2026-05-22"
  tags: "ci, watch, autofix, polling, github-actions, required-checks, force-push-prohibited"

progressive_disclosure:
  enabled: true
  level1_tokens: 120
  level2_tokens: 5000
---

# CI Loop (`moai-workflow-ci-loop`)

Unified CI watch + auto-fix loop. The orchestrator invokes this skill after `/moai sync`
Phase 4 (`gh pr create`) returns a PR number. The skill polls required checks, classifies
failures into mechanical vs semantic, attempts safe patches up to 3 iterations, and
escalates semantic failures via AskUserQuestion.

## Quick Reference

**Trigger**: `/moai sync` Phase 4 PR-create returns a PR number, or an existing PR needs
CI monitoring.

**Two phases, one skill**:
1. **Watch** — Poll `gh pr checks` every 30s, classify required vs auxiliary via
   `.github/required-checks.yml` SSoT; exit on green/fail/timeout.
2. **Auto-fix** — On required-fail (exit 2), receive JSON handoff, run up to 3 patch
   iterations, escalate semantic failures immediately.

**One-liner**:
```bash
MOAI_CIWATCH_GH=gh sh scripts/ci-watch/run.sh <PR_NUMBER> <BRANCH>
```

**Exit codes**:
- `0` — all required passed → ready-to-merge AskUserQuestion
- `1` — fatal error → surface remediation
- `2` — required failure → JSON handoff → auto-fix phase
- `3` — 30-min hard timeout → blocker message, return control

**HARD invariants**:
- AskUserQuestion is orchestrator-only — CLI, shell scripts, and the `manager-develop` (cycle_type=autofix) subagent MUST NOT call it.
- Force-push is absolutely prohibited (`--force`, `-f`, `--force-with-lease` all banned).
- Max 3 auto-fix iterations; iteration 4+ triggers mandatory blocking AskUserQuestion.
- Semantic failures (race, deadlock, panic, assertion) are never auto-patched.
- Protected files: `.env*`, credentials, `.claude/settings*.json`, `.github/required-checks.yml`,
  `scripts/ci-watch/run.sh`.

## Implementation Guide

### Phase 1 — Watch Loop

**Polling cadence**: 30 seconds minimum (GitHub API rate-limit). Override via
`CIWATCH_POLL_INTERVAL` env, never below 30 in production. Test mode uses
`MOAI_CIWATCH_NO_SLEEP=1` (single-tick exit).

**30-minute hard timeout**: `CIWATCH_TIMEOUT_SECONDS=1800` default. On timeout, exit 3.
Do not auto-restart.

**Required vs auxiliary**: Required checks live in `.github/required-checks.yml`
`branches.<pattern>.contexts`. Auxiliary checks listed under `auxiliary:` MUST NOT block
ready-to-merge. Hardcoding check names in scripts is prohibited.

**State file**: `.moai/state/ci-watch-active.flag` (YAML). Tracks `pr_number`, `started_at`,
`heartbeat_at`, `required_checks`, `abort_requested`. Heartbeat staleness > 90s allows
takeover. Abort: `moai pr watch --abort`.

**Background watch standardization**: For long-running PRs (5+ min), use
`gh pr checks <PR> --watch` invoked via `run_in_background: true`. Sleep + poll loops are
prohibited — they block the orchestrator's main session.

**Status report format** (stderr, state-change ticks only, no ANSI):
```
[ci-watch] PR #<N>: required 4/6 pass, 2 pending; advisory 0 fail
```

**Handoff schema on exit 2** — JSON with stable fields: `prNumber`, `branch`,
`failedChecks[]` (each entry `{name, runId, logUrl}`), `auxiliaryFailCount`,
`totalRequired`. Field stability: `name`, `runId`, `logUrl` are stable contract — do not
rename. Schema source: the CI-watch handoff struct.

### Phase 2 — Auto-Fix Loop

**Entry condition**: `ci-watch` exit 2 + valid JSON handoff. State file:
`.moai/state/ci-autofix-<PR>.json` (PR-scoped, 24-hour staleness threshold).

**OQ2 cadence matrix** (single source of truth for iteration behavior):

- iter 1, any mechanical sub_class → confirm + apply via AskUserQuestion (1st option =
  "패치 적용 (권장)").
- iter 1, semantic/unknown → escalate (no patch attempt) via AskUserQuestion with
  diagnosis report.
- iter 2-3, mechanical + sub_class=trivial → silent apply + log (no AskUserQuestion).
- iter 2-3, mechanical + sub_class=non-trivial → confirm + apply via AskUserQuestion.
- iter 2-3, semantic/unknown → escalate (no patch) via AskUserQuestion.
- iter 4+ → mandatory blocking AskUserQuestion (no timer, options: manual fix / revise
  SPEC / abandon PR).

"trivial" = whitespace, gofmt/goimports, import-order (matches `classify.sh` `RX_TRIVIAL_*`).

**Patch commit rule**: Every patch = new commit. Format:
`fix(ci): auto-fix <classification> failure (iter <N>)`. After push, re-invoke
`scripts/ci-watch/run.sh` to restart the watch loop.

**Iteration 4+ escalation** (mandatory blocking, no silent timeout):
1. (권장) 직접 수동 수정 — investigate and fix manually
2. SPEC 수정 — revise the SPEC and restart implementation
3. PR 포기 — close the PR and abandon this approach

**`manager-develop` (cycle_type=autofix) spawn prompt** injects: handoff JSON, classification + sub_class,
failed CI log + PR diff, mode directive (mechanical → propose unified-diff patch;
semantic/unknown → return diagnosis only, no patch). HARD: no AskUserQuestion call from
the subagent — return Markdown only.

**Audit log**: `.moai/reports/ci-autofix/<PR-NNN>-<YYYY-MM-DD>.md`. Append-only. Each
iteration records classification, sub_class, action, patch_sha, escalation_reason.

### Protected Files (never auto-modified)

- `**/.env`, `**/.env.*`
- `**/credentials*`, `**/*_key.json`, `**/*secret*`
- `.claude/settings.json`, `.claude/settings.local.json`
- `.github/required-checks.yml` (Wave 1 SSoT, read-only for Wave 2/3)
- `scripts/ci-watch/run.sh` (Wave 2 invariant)

If the `manager-develop` (cycle_type=autofix) subagent proposes a patch touching any of these, reject and escalate.

### Go Helpers and Shell Scripts

Go helpers: the CI-watch classifier (required-vs-auxiliary), the handoff JSON-schema
struct, the watch state file, and the PR-watch report emitters (`EmitReadyToMergeReport`,
`EmitFailureHandoff`). Shell: `scripts/ci-watch/run.sh` (main
loop, mock via `MOAI_CIWATCH_GH`); `scripts/ci-watch/lib/classify.sh` (yq + grep
fallback); `scripts/ci-autofix/log-fetch.sh` (failure log + PR diff);
`scripts/ci-autofix/classify.sh` (mechanical vs semantic).

**gh CLI compat**: requires `gh >= 2.50` for the `workflow` JSON field. On older `gh`,
`classify.sh` falls back to name-based heuristics from the `required:` list.

## Works Well With

- `manager-develop` (cycle_type=autofix) — failure diagnosis + patch proposal subagent
- `manager-git` — commit/push of auto-fix patches
- `.claude/rules/moai/workflow/ci-watch-protocol.md` — HARD watch invocation contract
- `.claude/rules/moai/workflow/ci-autofix-protocol.md` — HARD autofix invocation contract
- `.github/required-checks.yml` — Wave 1 SSoT
- `scripts/ci-mirror/run.sh` — Wave 1 local CI mirror

## Common Rationalizations

- "Skip watch loop for small PRs" — small PRs fail CI too. Loop costs nothing, saves manual polling.
- "Auxiliary failures should block merge" — advisory by SSoT definition. Edit `required-checks.yml` to change classification.
- "semantic 실패도 auto-patch 시도해보자" — semantic failures (race, assertion) cannot be auto-patched without context; wrong patch is worse.
- "force-push로 히스토리 정리하면 깔끔하다" — force-push destroys reviewer diff visibility. Always a new commit.
- "iter 3 이후 timeout" — silent timeout prohibited; user must decide explicitly.
- "trivial fix는 confirm 없이 바로 적용하자" — iter 1 always confirms; iter 2+ trivial may silent apply.

## Red Flags

- `manager-develop` (cycle_type=autofix) subagent calls AskUserQuestion (HARD: orchestrator-only)
- Iteration 4 auto-continues without blocking AskUser
- `git push --force` / `-f` / `--force-with-lease` anywhere in scripts
- Semantic classification produces a patch attempt
- State file missing → iteration counter lost → infinite loop risk
- Watch polling interval < 30s (rate-limit risk)
- `required-checks.yml` modified without `moai github init` re-run

## Verification

- [ ] `bash scripts/ci-watch/test/run_test.sh` passes all shell tests
- [ ] `go test ./internal/ciwatch/... ./internal/cli/pr/... -race` passes
- [ ] `internal/ciwatch/` coverage >= 85%
- [ ] No ANSI codes in `FormatStatusUpdate()` output
- [ ] `EmitReadyToMergeReport` first option carries `(권장)`
- [ ] CLI does NOT call AskUserQuestion
- [ ] `grep -r 'push -f\|push --force' scripts/ci-autofix/ scripts/ci-watch/` returns no matches
- [ ] Audit log `.moai/reports/ci-autofix/<PR>-<DATE>.md` contains every iteration

<!-- absorbed from moai-workflow-ci-watch + moai-workflow-ci-autofix per the skill consolidation policy -->
