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
  pass: the AC-ILT-012 positive-control pipeline `printf '+%s\n' '<sample with a card-id token>' | grep '^+[^+]' | grep -cE 't[0-9]{3}|SPEC-|20[0-9]{2}-[0-9]{2}-[0-9]{2}|[0-9a-f]{9}'`
  → `1` (the AC-ILT-012 positive control matches); `grep -c '^\[moai:integration-lock\] warning:'`
  → `1` on a sample holding the line, `0` on a sample without it.
- D3 / D4 — AC-ILT-007 adds github-flow + EMPTY develop and personal-mode git-flow + EMPTY develop
  cells; §D.3 row b names the first (plus absent config), new row g names the second.
- D5 — plan M5 wording carries lowercase `integration target`; AC-ILT-010 checks it
  case-sensitively.
- Every pre-existing test name cited in CMD-ILT-002/007/015/016 exists: a `git grep` for the 13
  `func <Name>(` declarations over `internal/cli` and `internal/template` found 13.

### iter-2 audit repair (v0.3.0)

plan-auditor iter2 returned FAIL 0.88 with three blocking findings (report
`.moai/reports/t637/plan-audit-SPEC-INTEGRATION-LOCK-TARGET-SOURCE-001-iter2.md`). Operator granted
one extra iteration (iter3 final). Repaired on HEAD `e51428db1`; acceptance.md and plan.md only,
no requirement or scope change:

- N1 — every fixture call and CMD-ILT-010(b) spells `/tmp/t637-fx-bin/moai`; the `BIN` variable is
  gone. `EV` remains, used only as a redirect/file argument (the iter2 guard probe accepted that
  form). Measured in this pass: `grep -n '\$BIN'` over acceptance.md matches only the §A.4 prose
  that explains the refusal.
- N2 — row b names only `TestIntegrationAcquire_GitHubFlowEmptyDevelopDoesNotWarn`, with the
  parenthetical corrected; new row i (absent config treated as git-flow) names
  `TestIntegrationAcquire_NoConfigCallerFallbackDoesNotWarn`; §D.6 reads "rows a-i".
- N3 — the three fixture cells run `acquire` inside `( cd /tmp/t637-fx-wt/cardA && … )`; every
  command using worktree-relative paths (CMD-ILT-001..010, 014, 015, 016, §D.5) begins with
  `cd /Users/goos/MoAI/moai-adk-go/.claude/worktrees/t637 &&` in its own invocation.
- N4 — the seam test is `TestLoadGitFlowIntegrationConfig_SeparatesGitFlowFromBranch`, selected with
  `^(…)$`.
- N5 — CMD-ILT-014 adds a fourth count, `--card <` on the release line, expected `0`.
- N6 — the positive control is `'see t637 here'`; measured in this pass through the same pipeline
  → `1`. The earlier SPEC-ID-shaped sample string was also removed from this file.
- iter1 D13 remains **declined**. The `OwnershipTransitionUnmeasured` INFO names the
  `(none) → draft` creation commit `a3b913b85`, which has already landed; a trailer on any later
  commit does not measure that transition, and rewriting the landed commit is out of bounds. In
  addition, git only reads a custom trailer from the message's final paragraph, and this card's
  dispatch requires `🗿 MoAI` to be the final line, so a trailer cannot be both parsed and
  compliant. The INFO does not affect the lint exit status.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
