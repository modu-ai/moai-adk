# Card t1641 — progress record

Card: t1641 (run tn1p4e, Class C, P1, operator-approved). Branch `WT-todo-unify`. SPEC: `SPEC-TODO-UNIFY-001`.

## Stage: plan (authored 2026-10-18T06:37Z, tree `026f48658f`; revised 2026-10-18 for audit iteration 1)

- Artifacts: `.moai/specs/SPEC-TODO-UNIFY-001/` holds spec.md, plan.md, acceptance.md, design.md, research.md, decision-index.md, progress.md. Tier L; 25 requirements, 25 criteria (revision 2).
- Plan audit iteration 1: FAIL 0.62 (MP-7 failed, ten blocking findings; receipt `rcpt-90e345b1d5488c65cc03140e`; report `.moai/reports/t1641/plan-audit.md`). Revision 2 repairs D1 to D15 (map in the SPEC's `progress.md` §E.1); it needs plan-audit iteration 2 (of 3).
- `moai spec lint` (plain and `--strict`) printed `0 error(s), 0 warning(s)` plus one INFO finding, using a binary built from this tree.
- Kickoff: no open decision row remains. Q5 (the automatic receipt) is the one trust-model decision, marked `lane-judgment, leader-confirmed 10-18 (conditional)`.

## Decisions made without asking (recorded, per the decision ladder; evidence in `decision-index.md`)

- Q1 absorb all six stage verbs on `moai todo`, delete none. Q2 the warning states removal one release after the release that first carries it. Q3 user-facing spelling only; the workflow document keeps its path (it is an identifier wired into five rule files, an embed line, a recorded fixture, and seven test files). Q4 lane own-card path plus the automatic receipt inside the two close paths plus edge T29; no observers. Q6 remove T18 and the T19 reservation now; declare T29. Q7 both records end `done`, factory edge first, queue archive second, with a stated recovery. Q8 a new leader-side `moai todo spec` verb, set-once per carrier. Q9 leader only.
- Q5 (leader-confirmed, conditional): the receipt is issued automatically only when three machine facts all hold (merged-local; a committed PASS-family `sync-audit.md` whose `audited_sha` is an ancestor of the integration branch; a valid re-measure record); issuer `auto`; revocable; otherwise fail-closed to the leader's manual approval.
- Tier L, because 73 files carry the command spelling and the change touches the transition table, the approval gate, and always-loaded surfaces.

## Gaps at the end of this stage

- No `go test` run (card constraint): every class B criterion is unadopted until its RED is recorded.
- `go list` reverse-dependency query refused by the worktree guard; importer lists come from `git grep` of import paths. Two more commands were refused in revision 2 (a `python3 -c` calling git after `cd`; a `git grep --include=*.go`).
- The installed `moai` is 516 commits behind the tree; observations through it are labelled in `acceptance.md` and `research.md`.
- `moai todo done` not executed against the live queue; the lane refusal was read from source.
- Cobra's execution order was read from the v1.10.2 source, not run.
- The exit status of empty `git grep` runs was not displayed by the tool.
- The board mod's JavaScript tests and the docs-site build were not examined.
