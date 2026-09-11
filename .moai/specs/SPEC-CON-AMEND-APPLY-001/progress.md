# progress.md — SPEC-CON-AMEND-APPLY-001

Card t659 · branch `WT-amend-apply` · worktree `.claude/worktrees/t659`

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: draft — revision 0.1.5 fixed plan-audit iteration 1 (FAIL 0.78) except D4; revision 0.1.6 applies the operator decision on D4 (resolved); awaiting `moai spec lint` and plan-audit iteration 2 (Tier L PASS threshold 0.85)
- tier: L, revision 0.1.6 — raised from M in revision 0.1.4 by operator decision (verdict §13.3): 21 requirements and 25 acceptance criteria each exceed the Tier M ceiling of 16; 25 acceptance criteria equals the Tier L ceiling
- plan artifacts: six files — `spec.md`, `plan.md`, `acceptance.md`, `design.md`, `research.md` (revision 0.1.6; `design.md` and `research.md` added in 0.1.4; `research.md` unchanged in 0.1.6), and this file
- plan commits: `7b4d1ac89` (0.1.0, rulings §7); `01af243ae` (0.1.1, rulings §8, adds this file); `e693f0583` (0.1.2, ruling §9); `4e9273d0b` (0.1.3 partial: spec.md, plan.md, and this file; acceptance.md not yet edited); `fa966740d` (0.1.3 completion: acceptance.md, reconciliation of all four files); `51454e5e5` (0.1.4 tier raise, adds `design.md` and `research.md`); `193136a6a` (0.1.5 plan-audit iteration 1 fixes D1–D3, D5–D16); the 0.1.6 D4 decision is the commit that updates this section
- rulings encoded: `.moai/reports/t659/verdict.md` §7 (Q1–Q5), §8 (G1–G5, scope addition (a)), §9 (G6, option (ii)), §11 (G7, option A), §12.5 (two plan-phase extensions accepted), §13.3 (Tier L); plan-audit iteration 1 defects D1–D3 and D5–D16 (`.moai/reports/t659/plan-audit-iter1.md`, verdict §15); operator decision D4 (amend path only)
- counts: 21 requirements, 25 acceptance criteria, 29 mutants (unchanged in 0.1.6, 0.1.5, and 0.1.4)
- changed in 0.1.6 (D4, operator decision): the containment check of REQ-CAA-021 runs on the amend path only — `Execute` and `runConstitutionAmend`; `LoadRegistry` keeps its present behaviour and its five non-amend call sites are unchanged; extending the check to read-only callers is a follow-up card candidate recorded by the lead. spec.md: REQ-CAA-020, REQ-CAA-021, §D terms, §A item 6, §C, §E.2, §F (new exclusion), §G (D4 resolved). plan.md: §A, §B, §C.1, §D verification scope (adds the `internal/spec` selector), M2 approach and exit, R-8. design.md: §A, §B, §C.2–§C.5. acceptance.md: AC-CAA-022 and AC-CAA-023 wording; AC-CAA-024 preservation case `loader_unchanged`; AC-CAA-025 preservation run of `TestLinter_AC08_DanglingRuleReference`; M-19 names AC-CAA-023's divergent rows; M-20 gains variant (iii) (check moved into `LoadRegistry`) and variant (i) drops AC-CAA-023; M-24 kills AC-CAA-025 `dry_run` instead of `load`. No ID added or renumbered
- changed in 0.1.5: REQ-CAA-004 (error names the registry path), REQ-CAA-008 and REQ-CAA-009 (`When` form), REQ-CAA-013 (`--dry-run` flag only), REQ-CAA-018 (temporary files removed on a failed restore), REQ-CAA-021 (lock file excepted), seven GEARS labels; AC-CAA-003, AC-CAA-005 (b), AC-CAA-007, AC-CAA-012 (rename-then-fail injector; subtests `first_rename_applied`, `third_rename_applied`), AC-CAA-014, AC-CAA-024 (rows `symlinked_registry`, `symlinked_file`; discriminating CLI case), AC-CAA-025 (count derived from the copy, target entry by rule, `load` entry point); M-5b and M-23 split into variants, the registry-path mutant gains a CLI-only variant; RED statements labelled as predictions; spec.md §E.4 gains an unapproved Gap (D7) and §G records D4; no ID added or renumbered
- added in 0.1.4: `design.md`, `research.md`; frontmatter `tier: L`, `version: "0.1.4"`; no requirement, acceptance criterion, mutant, ruling, or scope changed, no ID renumbered
- added in 0.1.3: REQ-CAA-021, AC-CAA-024, AC-CAA-025, M-21, M-22, M-23, M-24
- changed in 0.1.3: REQ-CAA-020 (its closing sentence points at REQ-CAA-021), AC-CAA-023 (asserts the refused path, not the loader's wording), M-20 (names the registry-path site of the one check; also kills AC-CAA-024 `relative_env_escape` and its CLI case)
- reconciled in the 0.1.3 completion: plan.md declared `M-21a` / `M-21b` while spec.md and this file declared one M-21; folded into M-21 with two variants, and M-22 given two variants (no-boundary prefix; no-Clean prefix), so M-15, M-18, M-21, and M-22 carry variants and the distinct mutant count stays 29; AC-CAA-024 gains a CLI dry-run case for the shared-check clause of REQ-CAA-021
- resolved question: D4 (plan-audit iteration 1) — operator decision in 0.1.6: amend path only (`spec.md` §G). No open question remains.
- spec lint: 0.1.6 _<pending>_ — the lane runs `moai spec lint` after this commit; no 0.1.5 result is recorded under `.moai/reports/t659/` (a working-tree listing of that directory taken with HEAD at `193136a6a` shows no `lint-0.1.5.txt`). 0.1.4 is recorded by the lane in `.moai/reports/t659/lint-0.1.4.txt` and verdict §14.2 (tree `51454e5e5`). Earlier results are recorded by the lane: 0.1.2 in `.moai/reports/t659/lint-0.1.2.txt` and verdict §10.2 (tree `e693f0583`), 0.1.3 in `.moai/reports/t659/lint-0.1.3.txt` and verdict §12.3 (tree `fa966740d`); none is evidence for 0.1.5 or 0.1.6.
- plan-audit verdict: iteration 1 FAIL 0.78 (`.moai/reports/t659/plan-audit-iter1.md`, audit HEAD `76144d40a`, verdict §15); iteration 2 _<pending plan-audit>_ — Tier L threshold 0.85, delta audit of D1–D16 with D4 applied as the operator decided it; the handoff states that 25 acceptance criteria equals the Tier L ceiling. The audit interrupted on `f2ba39c94` produced no artifact and is not a verdict basis (verdict §13.2).
- Implementation Kickoff Approval: _<pending>_

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
