---
id: SPEC-GOAL-SURFACE-UNIFY-001
title: Unify the goal surface on /moai goal and relocate goal presentation to the Implementation Kickoff Approval gate
version: 1.0.0
status: draft
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
priority: HIGH
phase: plan
module: doctrine
lifecycle: spec-anchored
tags: [goal, doctrine, session-handoff, slash-command, template-mirror]
tier: L
---

## §E.1 Plan-phase Audit-Ready Signal

Plan-phase artifact set complete (Tier L): `spec.md`, `plan.md`, `acceptance.md`, `design.md`, `research.md`, `progress.md`.

- SPEC ID regex check executed as Bash; observed output `SPEC-ID-CHECK: PASS`.
- **29** acceptance criteria authored; every judgment command executed against the worktree at `origin/main` = `e306e21a9` and its verbatim output recorded as the baseline in `acceptance.md` §B.
- All 29 ACs FAIL at the recorded baseline.
- Milestone ownership verified single-owner across all 35 run-phase paths plus 13 sync-phase paths (`plan.md` §F.1, §F.2).
- Scope spans three layers: doctrine (M1-M6), Go emission paths (M7, `cycle_type: tdd`), public docs (sync phase, owner `manager-docs`).

Five brief corrections surfaced and resolved during authoring, each verified by an executed command (see `research.md`):

| # | Correction | Evidence |
|---|---|---|
| 1 | Doctrine scope is 28 existing files, not 26 — the root `CLAUDE.md` pair was missed by a `.claude/`-scoped grep | `research.md` §A.3 |
| 2 | Four retention surfaces exist, not one — `internal/goal/evaluate.go`'s native-`/goal` yield invariant was mis-classified in the D4 brief as "implementation" | `research.md` §F.2 |
| 3 | The D4 "17 implementation references" figure is not reproducible (nearest pattern gives 35); no AC is built on it | `research.md` §F.3 |
| 4 | Sync-phase affected set is 13 files, not 9 — `advanced/self-evolving.md` (×4) is MoAI-surface primitive naming, not factual contrast | `research.md` §G.1 |
| 5 | `advanced/hooks-reference.md` exists in all four locales, not only en/ko — the gap is a content gap inside existing pages | `research.md` §G.3 |

Two D4 preconditions resolved in the safe direction: `PrimitiveGoal` occupancy measured **0** (hard rename safe, no back-compat needed), and the renderer confirmed to have **no test** (M7 must author the RED gate).

_Awaiting plan-audit and Implementation Kickoff Approval._

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
