# Progress — SPEC-POWERSHELL-DENY-PARITY-001

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts (Tier M): spec.md, plan.md, acceptance.md, progress.md. Status: draft. Version 0.2.0.
- Requirements: 18 GEARS (REQ-PSD-001..018), split by branch (P = parity, N = no-rule). ACs: 14 matrix rows + 7 Given-When-Then scenarios.
- SPEC ID self-check: `SPEC-POWERSHELL-DENY-PARITY-001` → PASS; uniqueness: 0 existing matches.
- Plan-audit iter-1: FAIL 0.66 (`.moai/reports/plan-audit/SPEC-POWERSHELL-DENY-PARITY-001-review-1.md`); v0.2.0 addresses D1–D10 (blocking) and D11–D15 (optional). Awaiting iter-2 re-audit.
- Premise status: built-in PowerShell removal protection verified from vendor docs (spec §A.2); "Bash denies do not reach the PowerShell tool" is UNVERIFIED; run-phase M1 measures it on a residual-set command before any fix.
- Dependency: absorb local `develop` carrying t1207 (`baa054586`) in M0.
- Open decisions D1 (operator), D2 (lead/run-phase), D3 (operator card issuance), D4 (operator) — plan.md § Open Decisions. Awaiting Implementation Kickoff Approval.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
