# SPEC-WORKTREE-BASEREF-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: CLOSED — plan-audit complete, operator kickoff approval granted, run-phase entry authorized
plan_complete_at: 2026-08-27
plan_audit_iter: 2 complete — verdict **PASS-WITH-DEBT 0.92 harmonic, 7/7 must-pass clear, zero blocking defects** (`.moai/reports/t313/plan-audit-iter2.md`). Iteration 1: FAIL 0.78 harmonic, 7/7 must-pass clear (`.moai/reports/t313/plan-audit-iter1.md`).
kickoff_approval: GRANTED by the operator, conditional on repairing debt item N1 first. N1 repaired in spec version 0.3.1 (below) before run-phase entry; the condition is discharged. Run-phase progression mode: **autonomous** — there is no milestone checkpoint at which a human reviews an intermediate reading, so criterion wording that admits a permissive interpretation was pinned rather than left open (see N2).
artifacts: spec.md (0.3.1, draft), plan.md, acceptance.md, progress.md — Tier M
tree: 48eb945df (worktree t313, branch WT-worktree-baseref)
counts: 16 GEARS requirements / 16 acceptance criteria (all MUST) — exactly at the Tier M ceiling of 16 REQ / 16 AC (`spec-workflow.md:148`). Unchanged by the 0.3.1 repairs: both are criterion-layer wording fixes, no id added, renamed, or renumbered.
iter2_repairs: D1 (REQ-WBR-004 now covered by AC-WBR-016, read-seam call count), D2 (vacuous `-run` pass mode closed in §D.3), D3 (AC-WBR-012 restated as REQ-WBR-015's three-part conjunction + guard-existence + mutation check), D4 (R1 premise corrected, consumer 1 narrowed to the primary checkout), A5 (REQ-WBR-011 bound to REQ-WBR-009's predicate as sole authority, AC-WBR-008 asserts the shared helper), plus debt D5-D10 folded in
iter2_changes: `.moai/reports/t313/spec-iter2-changes.md`
post_audit_repairs (spec 0.3.1): `.moai/reports/t313/spec-n1n2-changes.md`
  - **N1 (major, was blocking-carried)** — AC-WBR-013 check (1) probed only `internal/template/templates/$f`, so a correct implementation of plan §D emitted a false `NO-TEMPLATE-COUNTERPART` for `.moai/config/sections/git-strategy.yaml` (counterpart ships only as `.tmpl`; plan §B G5). Probe now accepts the plain OR the `.tmpl` form and reports only when neither changed. Re-measured against real history: old probe emits the false line on `664cd6eae^..664cd6eae`, new probe emits nothing; negative control `63b4628a6^..63b4628a6` (genuine orphan) still emits it, so detection was not weakened.
  - **N2 (minor-to-major, was blocking-carried)** — AC-WBR-016's "read seam" pinned to the **alignment-entry (configured-value) read**; the narrowing alternative was rejected because it would let the criterion lapse on the shipped empty default, which is the common case. Recorded in the criterion text and at plan.md §A D3.2.
debt_carried_into_run_phase (7 items, from the iter-2 verdict § Debt Carried Into Run-Phase):
  1. N1 — REPAIRED in 0.3.1 (no longer carried).
  2. N2 — REPAIRED in 0.3.1 (no longer carried).
  3. **Seam obligation is not optional (M2/M3)** — `internal/hook` carries no function-variable seam today; M2 must introduce one in its own new file. Budget for it; it is what makes AC-WBR-016 and AC-WBR-008's third assertion testable at all.
  4. **N4 — AC-WBR-013 check (2) is dormant, not correct.** Path mixing: the `[ -f ]` guard reconstructs a template-tree path while `diff -q` compares a sibling in `$b`'s own tree. Enumerates nothing for this SPEC (plan §D lists no hook wrapper); fix before any scope growth that touches one.
  5. **N5 — AC-WBR-002's first command has a prose-only pass condition.** Carry-over, not an iter-2 regression. Add an explicit value-emptiness assertion if run-phase wants it mechanical.
  6. **N3 — REQ-WBR-012 carries two GEARS modalities in one id.** Cosmetic; no run-phase action.
  7. **Folding traceability thinness (§ B1)** — REQ-WBR-004's narrowing and REQ-WBR-013's preservation clause are each carried by one AC half. Treat **both halves of AC-WBR-016 as separately mandatory**; an id-level matrix check will not notice a dropped half.
  Plus inherited and unchanged from iteration 1: **G2** (`EnterWorktree`'s `origin/HEAD` read is inferred, not read from source — the doctor item is the stated fallback), **G4** (`moai doctor` never executed), **G6** (rendered attribute order unmeasured — AC-WBR-011's two-branch condition MUST be collapsed to the single true form in run-phase).
blocker: none — D2 (widget shape) CLOSED by operator ruling 2026-08-27: `TypeText` with `main` / `develop` named in the field description; the new-combo-widget alternative rejected on template-neutrality grounds. Ruling + rationale recorded at plan.md §A D2.1; consequences at REQ-WBR-009 / REQ-WBR-012 / REQ-WBR-014, AC-WBR-009 / AC-WBR-011 / AC-WBR-015.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
