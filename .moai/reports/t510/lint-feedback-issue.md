## Summary

`moai spec lint` reports `CoverageIncomplete` findings for Tier M SPECs whose acceptance
criteria live in `acceptance.md` — the linter appears to scan `spec.md` only, so ACs that
exist (in `acceptance.md`, the standard V3R6 artifact layout) are counted as missing.

## Reproduction

1. Author a Tier M SPEC with the standard 4-artifact layout: `spec.md`, `plan.md`,
   `acceptance.md`, `progress.md`, ACs authored in `acceptance.md` (12 ACs in the case at
   hand).
2. Run `moai spec lint` on the SPEC directory.
3. Observed: `CoverageIncomplete` × 11 — one per requirement, despite every requirement
   having a traced AC in `acceptance.md`.

An independent plan-audit (iteration 1, score 0.85) confirmed all 11 findings as false
positives: AC↔REQ traceability exists in full, the linter just does not read
`acceptance.md`.

## Expected

Either the linter reads ACs from `acceptance.md` (the artifact where the V3R6 3-phase
contract puts them), or it suppresses `CoverageIncomplete` when `acceptance.md` exists in
the same directory.

## Environment

- moai v3.2.0-rc.0 (dev, dogfood)
- SPEC: `.moai/specs/SPEC-STATE-ANCHOR-001/` (11 REQ / 12 AC)
- Discovered during card t510 plan-audit (2026-09-07)
