---
description: CI auto-fix loop protocol — HARD invocation contract for the manager-develop autofix cycle (cycle_type=autofix). Loaded when working on the autofix cycle or on CI workflow definitions.
paths: ".claude/agents/moai/manager-develop.md,.claude/rules/moai/development/manager-develop-prompt-template.md,.github/workflows/**"
---

# CI Auto-Fix Protocol Rule

> This file is the single source of truth for the CI auto-fix loop invocation rules.
> Cross-referenced by: `.claude/agents/moai/manager-develop.md` (cycle_type=autofix) and
> `.claude/rules/moai/development/manager-develop-prompt-template.md`.

---

<!-- anchor: #ci-auto-fix-loop-entry-condition -->
## Entry Condition

[ZONE:Frozen] [HARD] The CI auto-fix loop MUST be entered ONLY when the orchestrator hands off
a failing required check. The orchestrator is the sole entry point: it observes the failing
check by whatever means the project provides, then delegates to `manager-develop` with
`cycle_type=autofix`.

```
failing required check -> orchestrator handoff -> ci-autofix loop entry
```

**Prerequisites** (all must be satisfied before loop entry):
1. The handoff names the pull request and branch under repair
2. At least one required check is failing (an empty failure set is not a loop entry)
3. The failing check's log output is available to the loop
4. State file `.moai/state/ci-autofix-<PR>.json` is writable

---

<!-- anchor: #iteration-limit -->
## Iteration Cap

[ZONE:Frozen] [HARD] The auto-fix loop MUST attempt at most **3 iterations**. The iteration
counter is persisted in `.moai/state/ci-autofix-<PR>.json`.

```
iteration 1, 2, 3 -> allowed
iteration 4+ -> MANDATORY BLOCKING AskUserQuestion (no patch attempt, no timer)
```

After 3 failed iterations, the orchestrator MUST present a mandatory blocking
AskUserQuestion with three options:
1. (Recommended) Fix manually — investigate and fix by hand, then push
2. Revise the SPEC and restart the implementation
3. Close the PR and abandon this approach

[ZONE:Frozen] [HARD] The AskUserQuestion at iteration > 3 MUST be a blocking call with no
silent timeout. The orchestrator waits indefinitely for user response before
taking any further action.

---

<!-- anchor: #commit-strategy -->
## Patch Commit Rule — No Force-Push

[ZONE:Frozen] [HARD] Every auto-fix patch MUST be applied as a **new commit** on the PR branch.
Do not force-push.

Prohibited commands:
- `git push --force`
- `git push -f`
- `git push --force-with-lease`
- `git commit --amend`

The orchestrator MUST use the standard `git add && git commit && git push` workflow.
Commit message format: `fix(ci): auto-fix <classification> failure (iter <N>)`

Example:
```bash
git add <specific files from the patch>
git commit -m "fix(ci): auto-fix mechanical/trivial failure (iter 2)"
git push origin <branch>
```

After the push, the orchestrator waits for the re-run of the same required check before
deciding whether another iteration is warranted.

---

<!-- anchor: #user-interaction-channel -->
## AskUserQuestion Boundary

[ZONE:Frozen] [HARD] AskUserQuestion is the **exclusive user interaction channel** for the
auto-fix loop. All user confirmations and escalations go through AskUserQuestion.
Any per-spawn `Agent(general-purpose)` diagnostic scoped to the loop MUST NOT call AskUserQuestion.

[ZONE:Frozen] [HARD] The orchestrator MUST preload AskUserQuestion via
`ToolSearch(query: "select:AskUserQuestion")` before every AskUserQuestion call.

Interaction surfaces:
- **Mechanical (iter 1)**: Confirm patch apply — options: apply (Recommended) / reject / escalate manually
- **Mechanical (iter 2-3 non-trivial)**: Same as iter 1
- **Mechanical (iter 2-3 trivial)**: Silent apply — NO AskUserQuestion
- **Semantic / unknown (any iter)**: Escalation with diagnosis report — NO patch attempt
- **Post-iter-3**: Mandatory blocking AskUserQuestion — options: manual fix / revise SPEC / abandon PR

---

<!-- anchor: #semantic-failure-handling -->
## Semantic Failure — No Auto-Patch

[ZONE:Frozen] [HARD] Semantic failures (data race, deadlock, panic, test assertion failure) MUST
NOT be automatically patched. The orchestrator MUST immediately escalate via
AskUserQuestion with the diagnosis report.

Classification is decided from the failing check's own output:
- A failure whose root cause is a lint rule, a formatting rule, a build error, a type error,
  or a missing dependency is **mechanical** — a patch attempt is allowed
- A failure whose root cause is a data race, deadlock, panic, or test assertion is
  **semantic** — immediate escalation
- A failure that cannot be classified is treated as semantic (conservative) — immediate escalation

The diagnosis is produced by a per-spawn `Agent(general-purpose)` with a diagnostic scope
(read-only investigation of the semantic failure), returning diagnosis only (no patch field).
The orchestrator presents the diagnosis to the user and waits for the user decision.

---

<!-- anchor: #protected-files -->
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

<!-- anchor: #audit-log -->
## Audit Log Requirement

[ZONE:Frozen] [HARD] Every auto-fix iteration MUST be logged to:
```
.moai/logs/ci-autofix/<PR-NNN>-<YYYY-MM-DD>.md
```

Each log entry MUST include:
- Iteration number
- classification and sub_class
- action taken (applied / escalated / aborted)
- patch SHA (if applied)
- escalation reason (if escalated)

The log file is append-only. The first iteration creates the file with a header.
The log file is a local artifact (gitignored via the `.moai/logs/` pattern).

---

## State File Lifecycle

The state file `.moai/state/ci-autofix-<PR>.json` tracks loop state:

- Created at loop entry (iteration=1)
- Updated after each iteration (iteration++)
- Deleted once the required checks pass
- Staleness threshold: 24 hours (a new invocation may reclaim a stale state file)
- PR-scoped filename prevents conflicts between concurrent PRs

---

## CI Infrastructure Preservation

[ZONE:Frozen] [HARD] The auto-fix loop MUST NOT modify CI watch infrastructure scripts or
workflow definitions. The autofix layer repairs the code under test; it never repairs the
harness that reports the failure, because a patch to the reporting layer can turn a real
failure into a false green.

Concretely, the loop MUST NOT touch workflow definition files, required-check
configuration, or any script whose role is to observe or report CI status.

---

Version: 2.0.0
Classification: HARD operational rule, applies to every `cycle_type=autofix` invocation
