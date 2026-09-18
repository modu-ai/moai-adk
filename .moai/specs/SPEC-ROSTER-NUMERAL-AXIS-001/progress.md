# SPEC-ROSTER-NUMERAL-AXIS-001 — Progress

card t930 · branch `WT-numeral-roster-guard` · base `690dfe369`

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts authored: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier M, Class C).
- Requirements: REQ-RNA-001 … REQ-RNA-013 (GEARS). Acceptance: AC-RNA-001 … AC-RNA-016.
- Decisions D1 (noun class), D2 (discharge rule) and D3 (mirror derivation) recorded in
  `spec.md` §D with their rejected alternatives and measured cost. All three are
  **operator-unanswered** — put to the operator, no answer inside the window — so they are the
  lead's / this SPEC's judgment and stay reviewable at the Implementation Kickoff Approval gate.
- Plan-audit iteration 1 returned FAIL (0.775 vs Tier M 0.80). Revision v0.2.0 clears D1-D7
  (blocking) and D8-D12 (minor); see `.moai/reports/t930/plan-audit.md` (gitignored by operator
  directive — never committed).
- Corrected cost arithmetic (measured by the dispatching lead in this tree at HEAD `6abcc85fa`):
  63 hit paths, 20 registry paths carrying `ClaimCount`, 17 hits discharged, **46 residual**,
  1 extra path the rejected any-row rule would free.
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
