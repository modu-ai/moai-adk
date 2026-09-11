# progress.md — SPEC-CON-AMEND-APPLY-001

Card t659 · branch `WT-amend-apply` · worktree `.claude/worktrees/t659`

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: draft — awaiting `moai spec lint` and plan-audit
- plan artifacts: `spec.md`, `plan.md`, `acceptance.md` (revision 0.1.3), this file
- plan commits: `7b4d1ac89` (0.1.0, rulings §7); `01af243ae` (0.1.1, rulings §8, adds this file); `e693f0583` (0.1.2, ruling §9); `4e9273d0b` (0.1.3 partial: spec.md, plan.md, and this file; acceptance.md not yet edited); the 0.1.3 completion (acceptance.md, reconciliation of all four files) is the commit that updates this section
- rulings encoded: `.moai/reports/t659/verdict.md` §7 (Q1–Q5), §8 (G1–G5, scope addition (a)), §9 (G6, option (ii)), §11 (G7, option A)
- counts: 21 requirements, 25 acceptance criteria, 29 mutants
- added in 0.1.3: REQ-CAA-021, AC-CAA-024, AC-CAA-025, M-21, M-22, M-23, M-24
- changed in 0.1.3: REQ-CAA-020 (its closing sentence points at REQ-CAA-021), AC-CAA-023 (asserts the refused path, not the loader's wording), M-20 (names the registry-path site of the one check; also kills AC-CAA-024 `relative_env_escape` and its CLI case)
- reconciled in the 0.1.3 completion: plan.md declared `M-21a` / `M-21b` while spec.md and this file declared one M-21; folded into M-21 with two variants, and M-22 given two variants (no-boundary prefix; no-Clean prefix), so M-15, M-18, M-21, and M-22 carry variants and the distinct mutant count stays 29; AC-CAA-024 gains a CLI dry-run case for the shared-check clause of REQ-CAA-021
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
