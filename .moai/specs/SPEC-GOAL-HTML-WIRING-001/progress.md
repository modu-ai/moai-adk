# progress.md — SPEC-GOAL-HTML-WIRING-001

> Run-phase evidence is populated by manager-develop (§E.2/§E.3) and manager-docs (§E.4). manager-spec emits only the §E.1 plan-phase signal and the §E.2-§E.4 placeholder headings.

## §A. Branch + Tier + Mode

- **Tier**: M (3 counted artifacts: spec.md + plan.md + acceptance.md)
- **Branch**: `plan/SPEC-GOAL-HTML-WIRING-001` (per repo-local PR-mandatory policy — `enforce_admins: true`)
- **Mode**: Mode 5 (sub-agent sequential) — default for coding-heavy Tier M work per orchestration-mode-selection.md §B
- **Predecessors**: SPEC-GOAL-HTML-FLOW-001 (completed), SPEC-INFINITE-GOAL-001 (completed)

## §B. Working Tree State

- Main checkout, feature branch (opt-in L2 worktree not used for plan-phase per spec-workflow.md Step 1)
- Unmodified at plan-phase start: `.moai/specs/SPEC-GOAL-HTML-FLOW-001/**`, `internal/report/planhtml/`, `internal/goal/dashboard.go`, `internal/hook/handoff_inject.go`

## §C. Plan-phase Decisions

- **Verdict persistence shape**: sidecar file `.moai/state/goal/<session-id>.verdict.json` (single writer: stop-goal evaluator; single reader: `runGoalRender`). NOT extending `<session>.json`; NOT recomputing at render time.
- **CLI namespace**: NEW `moai plan` Cobra parent command (currently does not exist as a CLI verb — `/moai plan` is a slash-command skill routing through the skill system, not the binary). `render-html` is the first subcommand.
- **Surface 1 c2 deferral**: per user-confirmed scope decision 1, the per-turn Stop-hook `.html` auto-refresh is out of scope — separate follow-up SPEC.

## §D. Open Questions

None — intent 100% drained upstream (3 user-confirmed scope decisions: c1-only, new SPEC, CLI verb + closeout rewrite). No `[NEEDS CLARIFICATION]` markers.

## §E.1 Plan-phase Audit-Ready Signal

_<pending plan-auditor — emit verdict + score here after audit>_

## §E.2 Run-phase Evidence

_<pending run-phase — populated by manager-develop>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — populated by manager-develop>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — populated by manager-docs; carries `sync_commit_sha`>_

## §F. Phase 4 Mode Selection

_<pending run-phase — populated by orchestrator before first Agent() spawn>_
