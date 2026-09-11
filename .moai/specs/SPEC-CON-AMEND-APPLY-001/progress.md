# progress.md — SPEC-CON-AMEND-APPLY-001

Card t659 · branch `WT-amend-apply` · worktree `.claude/worktrees/t659`

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: draft — awaiting `moai spec lint` and plan-audit (Tier L PASS threshold 0.85)
- tier: L, revision 0.1.4 — raised from M by operator decision (verdict §13.3): 21 requirements and 25 acceptance criteria each exceed the Tier M ceiling of 16; 25 acceptance criteria equals the Tier L ceiling
- plan artifacts: six files — `spec.md`, `plan.md`, `acceptance.md`, `design.md`, `research.md` (revision 0.1.4), and this file
- plan commits: `7b4d1ac89` (0.1.0, rulings §7); `01af243ae` (0.1.1, rulings §8, adds this file); `e693f0583` (0.1.2, ruling §9); `4e9273d0b` (0.1.3 partial: spec.md, plan.md, and this file; acceptance.md not yet edited); `fa966740d` (0.1.3 completion: acceptance.md, reconciliation of all four files); the 0.1.4 tier raise (adds `design.md` and `research.md`) is the commit that updates this section
- rulings encoded: `.moai/reports/t659/verdict.md` §7 (Q1–Q5), §8 (G1–G5, scope addition (a)), §9 (G6, option (ii)), §11 (G7, option A), §12.5 (two plan-phase extensions accepted), §13.3 (Tier L)
- counts: 21 requirements, 25 acceptance criteria, 29 mutants (unchanged in 0.1.4)
- added in 0.1.4: `design.md`, `research.md`; frontmatter `tier: L`, `version: "0.1.4"`; no requirement, acceptance criterion, mutant, ruling, or scope changed, no ID renumbered
- added in 0.1.3: REQ-CAA-021, AC-CAA-024, AC-CAA-025, M-21, M-22, M-23, M-24
- changed in 0.1.3: REQ-CAA-020 (its closing sentence points at REQ-CAA-021), AC-CAA-023 (asserts the refused path, not the loader's wording), M-20 (names the registry-path site of the one check; also kills AC-CAA-024 `relative_env_escape` and its CLI case)
- reconciled in the 0.1.3 completion: plan.md declared `M-21a` / `M-21b` while spec.md and this file declared one M-21; folded into M-21 with two variants, and M-22 given two variants (no-boundary prefix; no-Clean prefix), so M-15, M-18, M-21, and M-22 carry variants and the distinct mutant count stays 29; AC-CAA-024 gains a CLI dry-run case for the shared-check clause of REQ-CAA-021
- open question: none
- spec lint: 0.1.4 _<pending>_ — the lane runs `moai spec lint` after this commit. Earlier results are recorded by the lane: 0.1.2 in `.moai/reports/t659/lint-0.1.2.txt` and verdict §10.2 (tree `e693f0583`), 0.1.3 in `.moai/reports/t659/lint-0.1.3.txt` and verdict §12.3 (tree `fa966740d`); neither is evidence for 0.1.4.
- plan-audit verdict: _<pending plan-audit>_ — Tier L threshold 0.85; the handoff states that 25 acceptance criteria equals the Tier L ceiling; verdict file `.moai/reports/t659/plan-audit-iter1.md` (verdict §13.3). The audit interrupted on `f2ba39c94` produced no artifact and is not a verdict basis (verdict §13.2).
- Implementation Kickoff Approval: _<pending>_

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
