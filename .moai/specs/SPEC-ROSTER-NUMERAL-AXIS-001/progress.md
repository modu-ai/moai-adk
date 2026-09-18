# SPEC-ROSTER-NUMERAL-AXIS-001 — Progress

card t930 · branch `WT-numeral-roster-guard` · base `690dfe369`

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts authored: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier M, Class C).
- Requirements: REQ-RNA-001 … REQ-RNA-013 (GEARS). Acceptance: AC-RNA-001 … AC-RNA-015
  (15 criteria, under the Tier M ceiling of 16).
- Decisions D1 (noun class) and D2 (discharge rule) are ADOPTED and recorded in `spec.md` §D
  with their rejected alternatives and measured cost. Both are **operator-unanswered** — put to
  the operator, no answer inside the window — so they are the lead's / this SPEC's judgment and
  stay reviewable at the Implementation Kickoff Approval gate.
- **D3 (mirror derivation) — WITHDRAWN by operator decision, 2026-09-18.** Kept in `spec.md` §D
  as a recorded rejected alternative with its four findings (wrong premise measured at 13 rows /
  33 result rather than 20 / 26; derivation blinds the guard to sibling divergence;
  `readmeSite()` precedent overstated; withdrawal shrinks the SPEC). M5 stands at 46 authored
  rows with no folding.
- Plan-audit iteration 1 FAIL (0.775) → revision v0.2.0 cleared D1-D12. Iteration 2 FAIL (0.825,
  score clears; verdict rested on an introduced Must-vs-Must contradiction plus the partially
  resolved D3) → revision v0.3.0, authorised by the operator as a third iteration past the Tier
  M cap of two. Verdicts at `.moai/reports/t930/plan-audit.md` and
  `.moai/reports/t930/plan-audit-iter2.md` — both gitignored by operator directive, never
  committed.
- Cost arithmetic (measured by the dispatching lead in this tree at HEAD `6abcc85fa`): 63 hit
  paths, 20 registry paths carrying `ClaimCount`, 17 hits discharged, **46 residual**, 1 extra
  path the rejected any-row rule would free. Mirror-fold measurement at HEAD `d8f140b25`: 17 of
  the 46 under `internal/template/templates/`, 14 with a local twin on disk, **13 foldable
  pairs** → 33 rows, which is what withdrew D3.
- The 16-vs-17 `ClaimCount` disagreement between the auditor's pass-1 parse and the lead's
  figures is SETTLED in favour of the lead: a text-level `Path:`/`Claims:` parse cannot see the
  four rows `readmeSite()` builds from a `path` parameter. AC-RNA-013 requires the run-phase
  re-derivation to read `Registry()` rather than the file for this reason.

### Implementation Kickoff Approval

- **Given 2026-09-18 by the operator, CONDITIONAL on this SPEC reaching plan-audit PASS.**
- Recorded here so run-phase entry rests on a written approval rather than a remembered one.
- It does NOT authorise run-phase entry at this moment: the condition is unmet while the SPEC
  sits at plan-audit FAIL, and this session is plan-phase only.
- Plan-phase measurements attributed in `spec.md` §A: population figures supplied by the
  dispatching lead (this tree, base `690dfe369`); independently re-measured here —
  `go test -count=1 ./internal/harness/rosterguard/...` → `ok … 1.096s`,
  `profileMatrixAgentOrder` = 13 names, `registry.go` = 30 `Site` rows.
- Status: `draft`.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

🗿 MoAI
