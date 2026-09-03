# SPEC-ACHWD-STRIP-EXEMPT-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- Plan authority: `SPEC-HOOK-WIRING-DRIFT-001` plan.md §I (iteration-2 patch
  `8eb5b9102`).
- Plan audit: PASS-WITH-DEBT 0.80 (Tier M threshold met marginally), verdict
  `.moai/reports/t469/plan-audit.md` @ `7a8ae6744`; D1–D5 patched, confirming
  check textual only.
- Implementation Kickoff Approval: operator approval 2026-09-03, relayed by
  the lead; run phase GO.
- Pre-run absorb: local develop `6765a75c0` merged into the card branch
  (merge `a122e7568`, one conflict in CHANGELOG.md, both t216/t456 entries
  kept). All run work is on top of `a122e7568`.
- Scope statement: this run's file deltas are `.moai/specs/**` only, so the
  verification scope (file-delta packages ∪ reverse-dep packages via
  `go list -deps`) contains **no Go packages** — no package tests, no
  `go vet`, and no build are in scope. The §I.4 perl command and the spec
  lint are the entire check set. No local full suite is run.
- A5 verification outputs: _appended below after payload application._

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## Carried Debt (explicit carry-forward per lead instruction)

Source: `SPEC-HOOK-WIRING-DRIFT-001` plan.md §I.5 (residual risks recorded at
plan phase; plan-audit PASS-WITH-DEBT 0.80 was boundary-value — these are the
debts that survive the amendment).

| # | Debt | Owner |
|---|---|---|
| 1 | Normalization over-absorption at parenthetical granularity — a local-side and a template-side parenthetical containing a forbidden token can differ in non-token content invisibly to the mirror check | Neutrality doctrine (`.moai/docs/template-internal-isolation-doctrine.md` §25) / future fleet-wide strip-aware SPEC |
| 2 | Internal-date class is NOT normalized (no agreed regex) — a future date-mandated template strip will report a neutrality-mandated divergence as MISMATCH, the same false-FAIL shape one class over | Neutrality doctrine / future fleet-wide SPEC |
| 3 | The (ii-bare) absorption gap — a bare forbidden token inserted into the template copy passes the mirror check (exit 0) and is closed only by AC-HWD-016's template-side neutrality scan; either criterion alone does not close the space | AC-HWD-016 (standing compensating control) + future fleet-wide SPEC |
