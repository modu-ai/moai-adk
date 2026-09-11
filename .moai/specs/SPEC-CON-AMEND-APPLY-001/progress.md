# progress.md — SPEC-CON-AMEND-APPLY-001

Card t659 · branch `WT-amend-apply` · worktree `.claude/worktrees/t659`

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: draft — awaiting `moai spec lint` and plan-audit
- plan artifacts: `spec.md`, `plan.md`, `acceptance.md` (revision 0.1.3), this file
- plan commits: `7b4d1ac89` (0.1.0, rulings §7); `01af243ae` (0.1.1, rulings §8, adds this file); `e693f0583` (0.1.2, ruling §9); 0.1.3 (ruling §11) is the commit that updates this section
- rulings encoded: `.moai/reports/t659/verdict.md` §7 (Q1–Q5), §8 (G1–G5, scope addition (a)), §9 (G6, option (ii)), §11 (G7, option A)
- counts: 21 requirements, 25 acceptance criteria, 29 mutants
- added in 0.1.3: REQ-CAA-021, AC-CAA-024, AC-CAA-025, M-21, M-22, M-23, M-24
- changed in 0.1.3: REQ-CAA-020 (its closing sentence points at REQ-CAA-021), AC-CAA-023 (asserts the refused path, not the loader's wording), M-20 (also kills AC-CAA-024 `relative_env_escape`)
- open question: none
- spec lint: 0.1.3 _<pending>_ — the lane runs `moai spec lint` after this commit. The 0.1.2 result is recorded by the lane in `.moai/reports/t659/lint-0.1.2.txt` and verdict §10.2 (tree `e693f0583`); it is not evidence for 0.1.3.
- plan-audit verdict: _<pending plan-audit>_
- Implementation Kickoff Approval: _<pending>_

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
