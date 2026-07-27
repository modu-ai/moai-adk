# SPEC-WORKTREE-BRANCH-GUARD-001 — Progress

> Plan-phase artifact set created 2026-07-28 v0.1.0; audit-fix iter-2 v0.1.1.
> Status: draft (manager-spec owns the `(none) → draft` transition).
> Next phase owner: manager-develop (run-phase, `draft → in-progress`).

## §A. Plan-Phase Artifact Status

| Artifact | Status | Path |
|----------|--------|------|
| spec.md | draft | `.moai/specs/SPEC-WORKTREE-BRANCH-GUARD-001/spec.md` |
| plan.md | draft | `.moai/specs/SPEC-WORKTREE-BRANCH-GUARD-001/plan.md` |
| acceptance.md | draft | `.moai/specs/SPEC-WORKTREE-BRANCH-GUARD-001/acceptance.md` |
| progress.md | draft | `.moai/specs/SPEC-WORKTREE-BRANCH-GUARD-001/progress.md` |

## §B. Plan-Phase Research Summary

- All 5 verified preconditions confirmed (settings defaultMode, PreToolUse
  infra, static deny entries, `.worktreeinclude` divergence, `workflow.worktree`
  config).
- 3 design questions resolved: manager-git exemption boundary (hybrid
  identity-based), discriminant portability (`--path-format=absolute` primary,
  `--absolute-git-dir` + cwd-normalize fallback), fail-open with advisory.
- 1 research contradiction surfaced: `HookInput.AgentType` is main-thread
  `--agent`-only (NOT populated for `Agent(manager-git, ...)` sub-agent spawns).
  Closed by REQ-WBG-011b — sentinel env var `MOAI_BRANCH_GUARD_EXEMPT=1`.
- Census P1-B constraint analyzed: the deny rides the fast regex path
  (`checkBashCommand`), NOT the 10.58s quality-gate path (which only fires for
  `quality.IsGitCommit`).

## §C. Plan-Phase Decisions (encoded in spec.md)

| Decision | Resolution | Rationale |
|----------|------------|-----------|
| manager-git exemption boundary | Identity-based, unconditional within agent identity | Phase-D scoping rejected (not observable at PreToolUse time) |
| Discriminant portability | `--path-format=absolute` (git 2.31+) + `--absolute-git-dir` fallback | macOS Apple Git-155 = 2.50.1 supports the flag |
| Fail-open vs fail-deny | Fail-open with advisory (stderr + audit log) | Bash Risk-Amplifier Doctrine + Recovery-Signal Carve-Out norm |
| Sentinel deny reason prefix | `BRANCH_GUARD_VIOLATION:` | Orchestrator pattern-match without parsing full reason |
| Sentinel env var | `MOAI_BRANCH_GUARD_EXEMPT=1` | Closes the `agent_type` main-thread-only gap |

## §D. NEEDS-CLARIFICATION (orchestrator to resolve before run-phase)

- **B-1 (minor, non-blocking)**: orchestrator-side spawn contract for setting
  `MOAI_BRANCH_GUARD_EXEMPT=1` when spawning `manager-git` for Phase D work.
  The hook works correctly without it (deny fires as designed); the env var
  merely provides the exemption path. Plan.md §B-1 documents the one-line
  orchestrator-side edit required.

## §E.1 Plan-phase Audit-Ready Signal

_Populated by manager-spec at plan-phase completion; updated v0.1.1 (iter-2)._

- All 4 plan-phase artifacts (spec.md + plan.md + acceptance.md + progress.md)
  exist with `status: draft`, `version: 0.1.1`.
- `moai spec lint --strict` PASS on all 3 main artifacts (0 findings).
- All 13 REQs written in GEARS notation (no residual `IF/THEN` modality —
  verified via `grep -cE '^\s*(IF|THEN)\s' spec.md` → 0).
- All 13 ACs cite mechanical verification commands with falsification arms.
- Out of Scope sections satisfy `OutOfScopeRule` in all 3 files (spec.md 6
  H3 sub-headings; plan.md §H, acceptance.md §F — each with `-` bullets).
- SPEC ID regex self-check PASS: `SPEC-WORKTREE-BRANCH-GUARD-001`.
- Frontmatter carries all 12 canonical fields.
- 1 research contradiction surfaced (agent_type main-thread-only) and closed
  via REQ-WBG-011b.
- Plan-audit iter-1 verdict: PASS-With-Debt 0.84 (Tier M threshold 0.80
  cleared). Iter-2 fixes applied: D1-D7 + N1-N2 (8 findings total).
- 1 NEEDS-CLARIFICATION remains (B-1/D7: orchestrator-side spawn contract —
  deferred to follow-up SPEC per orchestrator iteration-2 decision;
  `[NEEDS CLARIFICATION: ...]` canonical marker form in plan.md).
- Ready for Implementation Kickoff Approval (plan→run HUMAN gate).

## §E.2 Run-phase Evidence

- **Run-phase entry 2026-07-28** (orchestrator). Pre-Spawn Sync Check executed:
  `git fetch origin main` + `git rev-list --count --left-right origin/main...HEAD`
  returned `3 1` (diverged — main advanced 3 commits via #1184/#1185/#1186,
  branch 1 ahead). File-overlap check vs planned work files = **0 overlap**.
  Active-session query `moai session list --filter-spec=...` = `[]`.
  Resolution: `git merge origin/main` (merge commit `c6139581f`, clean — no
  rebase, preserves plan commit SHA `0e9e3ee26` per
  `feedback_rebase_orphans_recorded_shas`). Post-merge divergence = `0 2`
  (clean; main fully absorbed). Working-tree dirty file `llm.yaml` is
  runtime-managed (`team_mode: glm`, §22.3) — excluded from pathspec commits.
  Implementation Kickoff Approval: granted (prior session, semi-autonomous
  progression — checkpoint after M1-M2).
- _<M1-M2 evidence to be populated by manager-develop>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

Recorded by orchestrator before the first run-phase `Agent()` spawn
(`orchestration-mode-selection.md` §D logging contract).

### Input parameters

| Parameter | Value |
|-----------|-------|
| tier | M |
| scope (file count) | ~8-10 (branch_guard.go + test, pre_tool.go ext, pre_tool_test.go ext, integration test, rule + template mirror, rule_template_mirror_test.go, .worktreeinclude x2, init/update/web.go) |
| domain count | 4 (hook Go code, rule/template mirror, CLI advisory, worktree config) |
| file language mix | Go source + markdown rules + YAML — mixed |
| concurrency benefit | LOW (coding-heavy implementation: regex logic, exec.Command dispatch, hook wiring) |

### Mode evaluation

| Mode | Selected? | Rationale |
|------|-----------|-----------|
| 1 trivial | no | multi-file Go implementation, not a typo |
| 2 background | no | write-capable, must commit |
| 3 agent-team | no (RETIRED) | tombstone |
| 4 parallel | no | multi-domain yes, but coding-heavy → Anthropic coding-task parallelism caveat |
| 5 sub-agent | **YES** | sequential coding work; M1→M2 tightly coupled (M2 wires M1 functions) |
| 6 workflow | no | <30 files, semantic/new-code not uniform-mechanical |

### Decision

`Decision: sub-agent` (Mode 5, sequential)

### Justification

This is coding-heavy implementation (regex matching, `exec.Command` dispatch
indirection, PreToolUse handler wiring) where sequential sub-agent delegation
is the safe default per Anthropic's coding-task parallelism caveat. M1
(discriminant + exemption + branch-state regex in `branch_guard.go`) and M2
(wiring `checkBranchState` into `pre_tool.go` Handle) are tightly coupled —
M2 calls M1's functions — so a single `manager-develop` spawn covers both to
avoid context handoff. Progression mode: semi-autonomous (checkpoint with the
user after M1-M2, per the kickoff-approved resume). M3-M6 (rule/template
mirror, CLI advisory, `.worktreeinclude` reconciliation, tests + sync) are
deferred to after the M1-M2 checkpoint.
