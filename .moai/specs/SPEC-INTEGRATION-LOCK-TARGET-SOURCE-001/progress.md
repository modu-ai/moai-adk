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

### iter-1 audit repair (v0.2.0)

plan-auditor iter1 returned FAIL 0.74 (Tier M threshold 0.80; Testability 0.55; report
`.moai/reports/t637/plan-audit-SPEC-INTEGRATION-LOCK-TARGET-SOURCE-001-iter1.md`). Repaired on HEAD
`fe92308e3` (the audit-report commit on top of `a3b913b85`):

- D1 — acceptance.md §A.4 defines `EV` and `BIN` as absolute paths with a one-invocation preamble;
  every fixture `acquire` / `status` / `release` carries `CLAUDE_PROJECT_DIR=/tmp/t637-fx` (C3′
  included).
- D2 — every deciding command moved out of the matrix into fenced blocks CMD-ILT-001..016; §A.2
  requires `-v`, no `no tests to run`, and one `--- PASS:` line per named test. Measured in this
  pass: `printf '+%s\n' 'see t637 and SPEC-X-001' | grep '^+[^+]' | grep -cE 't[0-9]{3}|SPEC-|20[0-9]{2}-[0-9]{2}-[0-9]{2}|[0-9a-f]{9}'`
  → `1` (the AC-ILT-012 positive control matches); `grep -c '^\[moai:integration-lock\] warning:'`
  → `1` on a sample holding the line, `0` on a sample without it.
- D3 / D4 — AC-ILT-007 adds github-flow + EMPTY develop and personal-mode git-flow + EMPTY develop
  cells; §D.3 row b names the first (plus absent config), new row g names the second.
- D5 — plan M5 wording carries lowercase `integration target`; AC-ILT-010 checks it
  case-sensitively.
- Every pre-existing test name cited in CMD-ILT-002/007/015/016 exists: a `git grep` for the 13
  `func <Name>(` declarations over `internal/cli` and `internal/template` found 13.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
