# Progress — SPEC-TODO-LANDING-STATE-001

Card t331. Plan-phase artifacts authored at tree `3de2f85a2` (worktree `.claude/worktrees/t331`,
branch `WT-card-landing-state`).

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier M set).
- SPEC ID regex self-check executed: `SPEC-TODO-LANDING-STATE-001` → `PASS`.
- ID uniqueness confirmed against `.moai/specs/SPEC-TODO-*`.
- Root cause measured, not inferred: `LandedRef = "origin/main"` against an integration branch of
  `develop`. Re-measured at 0.2.0 against `origin/main` `48239c7dc` / `origin/develop` `c6aa61346`,
  observed 2026-08-28T13:15Z: `git rev-list --count --left-right origin/main...origin/develop` →
  `0	349` (`0	329` at 0.1.0 — the count moves, the leading zero does not).

### Version 0.2.0 — plan-audit iteration 1 remediation (FAIL 0.74)

- **Scope split by operator ruling.** The landing-evidence half — storage, the recording verb, an
  observed commit on `todo pr`, the live SPEC-status read — moved to card **t359**, which depends on
  this SPEC landing first. What remains is half A, the discriminator.
- Requirements: **REQ-TLS-001..011** (11, Tier M ceiling 16), GEARS notation, traceable to
  **AC-TLS-001..010** (10, ceiling 16) in spec.md §E. Was 26 REQ / 16 AC at 0.1.0.
- The storage deviation that required a gate ruling at 0.1.0 is **withdrawn with the B half**: this
  SPEC persists nothing, so there is no deviation from the dispatch's storage direction to rule on.
- Open clarifications: **none**. Two were ruled from the sources (plan.md §C); the third retired
  with the storage half.
- AC-TLS-008's RED **observed by planting a mutant** in `runTodoPR`, running
  `TestTodoPR_QueueDirUnchanged` (4/4 sub-cases FAIL), and reverting; production tree restored and
  re-verified GREEN. Full record: acceptance.md §D.1.
- Citations: seven corrected, every remaining one re-opened at its address at HEAD `11426a128`.
- Not committed; awaiting lead review.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
