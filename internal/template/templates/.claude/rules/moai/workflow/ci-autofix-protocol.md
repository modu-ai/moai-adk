---
description: CI auto-fix loop protocol — HARD invocation contract for moai-workflow-ci-loop skill (auto-fix phase). Auto-loaded when the ci-loop skill is active.
paths:
  - ".claude/skills/moai-workflow-ci-loop/SKILL.md"
---

# CI Auto-Fix Protocol Rule

> This file is the single source of truth for the CI auto-fix loop invocation rules.
> Cross-referenced by: SKILL.md, moai-workflow-ci-loop (unified watch + autofix skill).

---

## Entry Condition

[ZONE:Frozen] [HARD] The CI auto-fix loop MUST be entered ONLY when `scripts/ci-watch/run.sh`
exits with code 2 and emits a valid JSON handoff to stdout.

```
ci-watch exit 2 → JSON handoff → ci-autofix loop entry
```

**Prerequisites** (all must be satisfied before loop entry):
1. Handoff JSON is valid (contains `prNumber`, `branch`, `failedChecks[]`)
2. `failedChecks[]` is non-empty (at least one required check failed)
3. `scripts/ci-autofix/log-fetch.sh` is executable
4. `scripts/ci-autofix/classify.sh` is executable
5. State file `.moai/state/ci-autofix-<PR>.json` is writable

---

## Iteration Cap

[ZONE:Frozen] [HARD] The auto-fix loop MUST attempt at most **3 iterations**. The iteration
counter is persisted in `.moai/state/ci-autofix-<PR>.json`.

```
iteration 1, 2, 3 → allowed
iteration 4+ → MANDATORY BLOCKING AskUserQuestion (no patch attempt, no timer)
```

After 3 failed iterations, the orchestrator MUST present a mandatory blocking
AskUserQuestion with three options:
1. (권장) 직접 수동 수정 — Investigate and fix manually, then push
2. SPEC 수정 — Revise the SPEC and restart the implementation
3. PR 포기 — Close the PR and abandon this approach

[ZONE:Frozen] [HARD] The AskUserQuestion at iteration > 3 MUST be a blocking call with no
silent timeout. The orchestrator waits indefinitely for user response before
taking any further action.

---

## Patch Commit Rule — No Force-Push

[ZONE:Frozen] [HARD] Every auto-fix patch MUST be applied as a **new commit** on the PR branch.
Do not force-push.

Prohibited commands:
- `git push --force`
- `git push -f`
- `git push --force-with-lease`

The orchestrator MUST use standard `git add && git commit && git push` workflow.
Commit message format: `fix(ci): auto-fix <classification> failure (iter <N>)`

Example:
```bash
git add -p  # or specific files from patch
git commit -m "fix(ci): auto-fix mechanical/trivial failure (iter 2)"
git push origin <branch>
```

After push, the orchestrator MUST re-invoke `scripts/ci-watch/run.sh` to restart
the watch loop for the same PR.

---

## AskUserQuestion Boundary

[ZONE:Frozen] [HARD] AskUserQuestion is the **exclusive user interaction channel** for the
auto-fix loop. All user confirmations and escalations go through AskUserQuestion.
The CLI, shell scripts, and any per-spawn `Agent(general-purpose)` diagnostic scoped to the loop MUST NOT call AskUserQuestion.

[ZONE:Frozen] [HARD] The orchestrator MUST preload AskUserQuestion via
`ToolSearch(query: "select:AskUserQuestion")` before every AskUserQuestion call.

Interaction surfaces:
- **Mechanical (iter 1)**: Confirm patch apply — options: apply (권장) / reject / escalate manually
- **Mechanical (iter 2-3 non-trivial)**: Same as iter 1
- **Mechanical (iter 2-3 trivial)**: Silent apply — NO AskUserQuestion
- **Semantic / unknown (any iter)**: Escalation with diagnosis report — NO patch attempt
- **Post-iter-3**: Mandatory blocking AskUserQuestion — options: manual fix / revise SPEC / abandon PR

---

## Semantic Failure — No Auto-Patch

[ZONE:Frozen] [HARD] Semantic failures (data race, deadlock, panic, test assertion failure) MUST
NOT be automatically patched. The orchestrator MUST immediately escalate via
AskUserQuestion with the diagnosis report.

Semantic classification is determined by `scripts/ci-autofix/classify.sh`:
- `classification=semantic` → immediate escalation
- `classification=unknown` → treated as semantic (conservative) → immediate escalation

The diagnosis is produced by a per-spawn `Agent(general-purpose)` with a diagnostic scope (read-only investigation of the semantic failure), returning diagnosis only (no patch field). The orchestrator presents the diagnosis to the user and waits for user decision. Per `.claude/rules/moai/workflow/archived-agent-rejection.md` §C, the archived `manager-quality` agent is replaced by this `Agent(general-purpose)` diagnostic scope; the sync-phase quality gate is mechanically enforced by the Stop hook `sync-phase-quality-gate.sh` (see `.claude/rules/moai/core/agent-common-protocol.md` § Hook Invocation Surface).

---

## Secrets and Credentials Protection

[ZONE:Frozen] [HARD] The auto-fix loop MUST NOT modify `.env`, `.env.*`, credentials files,
API key files, or any file matching common secrets patterns.

File patterns that MUST NOT be touched by auto-fix:
- `**/.env`, `**/.env.*`
- `**/credentials*`, `**/*_key.json`, `**/*secret*`
- `.claude/settings.json`, `.claude/settings.local.json`

If a patch proposed by the diagnostic `Agent(general-purpose)` scope touches these files, the orchestrator MUST
reject the patch and escalate to the user.

---

## Audit Log Requirement

[ZONE:Frozen] [HARD] Every auto-fix iteration MUST be logged to:
```
.moai/reports/ci-autofix/<PR-NNN>-<YYYY-MM-DD>.md
```

Each log entry MUST include:
- Iteration number
- classification and sub_class
- action taken (applied / escalated / aborted)
- patch SHA (if applied)
- escalation reason (if escalated)

The log file is append-only. The first iteration creates the file with a header.
The log file is a local artifact (gitignored via `.moai/reports/` pattern).

---

## State File Lifecycle

The state file `.moai/state/ci-autofix-<PR>.json` tracks loop state:

- Created at loop entry (iteration=1)
- Updated after each iteration (iteration++)
- Deleted on successful CI green (exit 0 from ci-watch)
- Staleness threshold: 24 hours (new invocation may reclaim a stale state file)
- PR-scoped filename prevents conflicts between concurrent PRs

---

## Wave 2 Contract Preservation

[ZONE:Frozen] [HARD] The auto-fix loop MUST NOT modify `scripts/ci-watch/run.sh` or any Wave 2
artifacts. Wave 3 is a read-only consumer of Wave 2 outputs.

The handoff schema fields `name`, `runId`, `logUrl` in `failedChecks[]` are
stable contract fields. Rename or removal requires simultaneous update of both
Wave 2 and Wave 3 code.

Cross-reference: `internal/template/templates/.claude/rules/moai/workflow/ci-watch-protocol.md`
§"T3 Handoff Format" for the authoritative schema definition.

---

Version: 1.0.0
Classification: HARD operational rule, applies to all T3 auto-fix invocations
