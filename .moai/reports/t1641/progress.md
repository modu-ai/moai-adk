# Card t1641 — progress record

Card: t1641 (run tn1p4e, Class C, P1, operator-approved). Branch `WT-todo-unify`. SPEC: `SPEC-TODO-UNIFY-001`.

## Stage: plan (authored 2026-10-18T06:37Z, tree `026f48658f`)

- Artifacts: `.moai/specs/SPEC-TODO-UNIFY-001/` holds spec.md, plan.md, acceptance.md, design.md, research.md, decision-index.md, progress.md. Tier L; 24 requirements, 24 criteria.
- `moai spec lint` (plain and `--strict`) printed no findings, using a binary built from this tree.
- Plan audit: not yet run.
- Kickoff: blocked on six product-level decision rows (Q1, Q3, Q4, Q5, Q6, Q7 of `decision-index.md`); Q2, Q8, Q9 carry `DEFAULT-APPLIED`.

## Decisions made without asking (recorded, per the decision ladder)

- Tier L, because 73 files carry the command spelling and the change touches the transition table, the approval gate, and always-loaded surfaces.
- Internal names and persisted formats stay (reading (a) of Q3) as the working assumption of the plan; the plan builds nothing that blocks reading (b).
- Scope of the done path: the lane own-card path and the receipt seam only (reading (a) of Q4); observers are not built.
- spec_id setter: a new verb, leader-only (Q8, Q9 defaults).

## Gaps at the end of this stage

- No `go test` run (card constraint): every class B criterion is unadopted until its RED is recorded.
- `go list` reverse-dependency query refused by the worktree guard; importer lists come from `git grep` of import paths.
- The installed `moai` is 516 commits behind the tree; observations through it are labelled in `acceptance.md` and `research.md`.
- `moai todo done` not executed against the live queue; the lane refusal was read from source.
- The board mod's JavaScript tests and the docs-site build were not examined.
