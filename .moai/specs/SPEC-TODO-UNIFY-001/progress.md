# Progress: SPEC-TODO-UNIFY-001

## §E.1 Plan-phase Audit-Ready Signal

plan_authored_at: 2026-10-18T06:37:06Z
plan_status: audit-ready
plan_revision: 1
audit_ready: true
tier: L (spec.md, plan.md, acceptance.md, design.md, research.md)
tree_basis: 026f48658f (branch WT-todo-unify)

Lint evidence (this run, this tree): `moai spec lint SPEC-TODO-UNIFY-001` and `moai spec lint --strict SPEC-TODO-UNIFY-001`, each printed `✓ No findings — all SPEC documents are valid`. Judging build: the lint binary was built from this tree at HEAD `026f48658f` with a plain `go build -o <scratch> ./cmd/moai` (no Makefile LDFLAGS, so it reports `v3.1.3 none`; the version stamp is not the build's identity, the tree is). The installed `moai` (`ee981a932-dirty`, a strict ancestor 516 commits behind) first reported one `VacuousTestAssertion` warning (an unanchored `-run` placeholder in `acceptance.md`); the placeholder was replaced by the real test name, and both builds then reported no findings.

Requirement and criterion counts: 24 requirements (REQ-TU-001 to REQ-TU-024), 24 criteria (AC-TU-001 to AC-TU-024); Tier L ceilings are 25 each.

Open decisions: nine rows in `decision-index.md` (Q1 to Q9). Three implementation-level rows (Q2, Q8, Q9) carry `DEFAULT-APPLIED`. Six product-level rows (Q1, Q3, Q4, Q5, Q6, Q7) have an empty verdict and block the autonomous Kickoff; they go to the operator.

Not done at plan time, by the card's constraints: no `go test` was run, so every class B criterion in `acceptance.md` is unadopted until its milestone records the RED output; no plan-audit has run.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
