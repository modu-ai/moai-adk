---
id: SPEC-GOAL-STOPFAILURE-CLEAR-001
artifact: plan
version: "0.1.0"
status: draft
created: 2026-08-19
updated: 2026-08-19
---

## Approach

One handler gains one responsibility. `stopFailureHandler.Handle` already classifies `error_type` to pick a `systemMessage`; the disarm rides that same classification rather than introducing a second one.

## Change set

| File | Change |
|---|---|
| `internal/goal/schema.go` | `IsUnrecoverableStopFailure(errorType string) bool` — the classifier, owned by the goal package because it answers a goal-lifecycle question |
| `internal/hook/stop_failure.go` | on an unrecoverable type, resolve the project root, clear the session's goal, append the disarm to the message |
| `.claude/settings.json` + `internal/template/templates/.claude/settings.json.tmpl` | add `oauth_org_not_allowed` to the `StopFailure` matcher |

Import direction: `internal/hook` → `internal/goal`. Verified acyclic — `internal/goal` imports neither `internal/hook` nor `internal/cli` (`grep -rn 'moai-adk/internal/hook' internal/goal/*.go` → no match).

## Disarm mechanism: reuse `ClearGoal`, do not add a status

`goal.ClearGoal(projectRoot, sessionID)` already removes the state file, the dashboard sibling, and the verdict sidecar, and is idempotent on absence. It is what `moai goal clear` uses, so the disarm lands the session in a state the rest of the system already understands.

The alternative — writing `Status: cleared` with a new `cleared_reason` field and keeping the file — was considered and rejected for this card. It would preserve a post-hoc explanation, but it needs a schema field, a migration story for readers that do not know it, and a decision about what `moai goal status` should print for a goal that exists but is inert. The notice in REQ-GSF-003 carries the explanation at the moment it matters, which is what the upstream behavior does too ("clears itself **with a notice**"). If a durable record is later wanted, that is a schema change worth its own card.

## Project-root resolution

`HookInput.CWD` is the hook's own cwd field and is the same value the other handlers resolve from. The existing `resolveProjectRootFromInputOrEnv` helper (`input.CWD` → `$CLAUDE_PROJECT_DIR` → `os.Getwd()`) is reused rather than re-derived, so a worktree-resident session resolves the same way here as everywhere else.

## Failure handling

Every step is best-effort. A root that does not resolve, a goal that is absent, a remove that fails — each skips the disarm and leaves the pre-existing message untouched. The handler's contract is already "always non-blocking — never returns an error", and this change does not weaken it.

## TDD order

1. `TestStopFailureHandler_DisarmsGoalOnUnrecoverable` — arm a goal, deliver `authentication_failed`, assert the state file is gone and the message names the disarm. **Expected to fail before the implementation.**
2. `TestStopFailureHandler_KeepsGoalOnRecoverable` — arm a goal, deliver `rate_limit`, assert the goal is still armed. This is the regression that stops a later "just clear on any StopFailure" simplification.
3. Implement.
4. `TestIsUnrecoverableStopFailure` — the classifier over the full documented enum, so a new error type added upstream is a deliberate decision rather than an accident of a default branch.
