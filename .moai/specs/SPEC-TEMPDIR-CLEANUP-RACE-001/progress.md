# SPEC-TEMPDIR-CLEANUP-RACE-001 — progress

Card: t352 · Branch: `WT-tempdir-cleanup-race` · Tier S

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts authored: `spec.md`, `plan.md`, `acceptance.md`, `progress.md`.
- Evidence base: `.moai/reports/t352/reproduction.md` (base `origin/develop` @ `77b2bcae6`).
- SPEC ID regex self-check executed as Bash; output `PASS`.
- Mechanism rung chosen in `plan.md` §D.1: rung 1 (exported synchronous-deferred-scan option),
  constrained to a **variadic** option parameter so `internal/cli/deps.go:221` keeps compiling
  unchanged (`plan.md` §D.3).
- Plan audit iter-1: PASS 0.88; iter-2: PASS 0.93 (monotonic). Repairs D1-D8 landed at spec version
  0.1.1, D9-D10 at 0.1.2 (`.moai/reports/t352/plan-audit.md`); the Tier S budget overrun (9 ACs, separate `acceptance.md`) is recorded and
  justified in `plan.md` §D.4 rather than resolved by deleting a criterion.
- AC-TCR-002b base SHA at plan time: `git merge-base origin/develop HEAD` → `77b2bcae6`
  (`origin/develop` tip had already moved to `c6aa61346`; the three-dot form pins the fork point).
- Status: `draft`. Awaiting Implementation Kickoff Approval.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
