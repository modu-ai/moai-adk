# SPEC-INTEGRATION-LOCK-TARGET-SOURCE-001 — progress (card t637)

Plan-phase artifacts authored 2026-09-11 on HEAD `1ad0fdc09` @ `WT-acquire-branch-record`
(worktree `.claude/worktrees/t637`). Tier M. Status: draft.

## §E.1 Plan-phase Audit-Ready Signal

- Artifact set for Tier M: spec.md, plan.md, acceptance.md, progress.md (this file).
- SPEC ID regex check, run as Bash in this plan pass:
  `ID="SPEC-INTEGRATION-LOCK-TARGET-SOURCE-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL`
  → observed output `PASS`.
- ID uniqueness: `ls .moai/specs/SPEC-INTEGRATION-LOCK-TARGET-SOURCE-001` → "No such file or
  directory" before creation; `git grep -l "SPEC-INTEGRATION-LOCK-TARGET-SOURCE"` → no output.
- Frontmatter: 12 canonical fields plus `tier: M`; `status: draft`; `phase` is a release target
  ("v3.2.0 target").
- Requirements: REQ-ILT-001..013 in GEARS notation (no IF/THEN modality) — within the Tier M
  ceiling of 16.
- Acceptance: AC-ILT-001..014 — within the Tier M ceiling of 16; includes both lead-mandated
  fixture controls (AC-ILT-005 same tree / own branch, AC-ILT-006 other tree / other branch), the
  no-warn negative (AC-ILT-007), flag and config sources (AC-ILT-003/004), old-record
  compatibility (AC-ILT-009), and the mutation guard (AC-ILT-014, rows a-f).
- Exclusions: spec.md §E carries three `### Out of Scope — <topic>` sub-headings, each with `-`
  bullets.
- Baseline: `git diff --stat 4c99d973e origin/develop -- <12 target paths>` and
  `git diff --stat 4c99d973e HEAD -- <same paths>` both printed nothing (exit 0) in this pass.
- AC-ILT-011 base: `git grep -n "moai integration acquire" -- <5 doc files>` → 8 lines, of which
  0 contain `--card`.
- Lint: `moai spec lint .moai/specs/SPEC-INTEGRATION-LOCK-TARGET-SOURCE-001/spec.md` (installed
  binary v3.2.0-rc.7) → exit 0, "No findings — all SPEC documents are valid" (a first run
  flagged four REQ bullets whose `shall` sat on a wrapped second line and one traceability table;
  both were reshaped, then re-run). `spec_audit` for this ID → one INFO `EraAutoDetected` (V3R6),
  no drift.
- Decisions recorded in plan.md: warning on standard error (§B1), status text shape (§B2),
  git-flow predicate (§B3), release invocation carries no card (§E2), no catalog entry covers the
  kanban-dispatch rule (§E4).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
