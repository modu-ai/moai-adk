---
id: SPEC-SPECLINT-ARTIFACT-STATUS-001
title: "Implementation plan — two-line Artifact Statelessness repair"
version: "0.1.0"
created: 2026-09-06
updated: 2026-09-06
author: manager-spec
---

# plan.md — SPEC-SPECLINT-ARTIFACT-STATUS-001

## §A Context

CI job `spec-lint` fails on exactly 2 ERROR-severity findings, both `ArtifactStatusFieldForbidden`,
both inside one SPEC's sibling artifacts (`SPEC-CODEX-E2E-MEASURE-001/plan.md:5`, `acceptance.md:5`).
Card t462's close commit `185ce7d57` wrote `status: completed` into them. The schema's Artifact
Statelessness rule (`spec-frontmatter-schema.md` § Artifact Statelessness) makes non-spec.md
artifacts stateless on the status axis; the lint error message itself prescribes deleting the line.
Measured baseline: `.moai/reports/t490/red-baseline.md` (2 errors / 4,333 warnings, exit 1, tree
`615d18c1f`).

## §B Known Issues

- The CLI exits 1 on ERROR findings only (`internal/cli/spec_lint.go:97`); the 4,333 warnings never
  flip the job. Post-fix, the command will still exit 1 — verdicts read `^ERROR` match counts.
- The StatusTransitionInvalid 99 / MissingExclusions 25 warning families are a separate card's
  concern (rule-semantics design decisions); do not touch them here.

## §C Pre-flight

- [ ] `git rev-parse --short HEAD` — confirm working tree is the card worktree on
      `WT-speclint-status-transition`, base `615d18c1f`.
- [ ] Blob pin check: `git rev-parse HEAD:.moai/specs/SPEC-CODEX-E2E-MEASURE-001/plan.md` = `15d5ef6ae478b97fc301f9c778e472aa7bd6c8d8` and `.../acceptance.md` = `13d830137cd34b4fcb3616a92f0099e4880f2d54`. Mismatch ⇒ STOP and re-measure (premise moved).
- [ ] Confirm `.moai/reports/t490/red-baseline.md` exists and is uncommitted (it is the baseline that M1 commits).

## §D Constraints

- Two files, one deleted line each (`status: completed`, line 5). No other edit — no reformatting,
  no field reordering, no trailing-whitespace cleanup.
- No `spec.md` edits anywhere; no lint-rule code edits; no HISTORY edits to completed SPECs.
- The lane commits with explicit pathspecs (never `git add -A`); every commit message carries `t490`.

## §E Self-Verification

Judged on match counts, never exit codes. Each step re-runs `go run ./cmd/moai spec lint --strict`
with output redirected to a file and the exit code captured separately.

- E1: GREEN — `grep -c '^ERROR' <output>` = 0 (AC-GREEN).
- E2: MUTANT — restore the two lines from `615d18c1f`, re-run, `grep -c '^ERROR'` = 2 with both lines
  `ArtifactStatusFieldForbidden`; re-apply the fix, re-confirm 0 (AC-MUTANT).
- E3: ORDERING — `git log --oneline` shows the red-baseline commit before the fix commit (AC-ORDERING).
- E4: SCOPE — `git diff 615d18c1f..HEAD -- .moai/specs/SPEC-CODEX-E2E-MEASURE-001/` shows exactly two
  `-status: completed` deletions and nothing else (AC-SCOPE).
- E5: EVIDENCE — GREEN + mutant outputs exported under `.moai/reports/t490/`; every branch commit
  message carries `t490` (AC-EVIDENCE).

## §F Milestones

Ordered by decision-reversibility — the only genuinely reversible-risk decision (the two-line edit
itself) lands after the ordering-sensitive baseline is committed, so the commit graph can witness
§2.3 ordering.

- **M1 (Priority High) — Commit the RED baseline.** `git add .moai/reports/t490/red-baseline.md`
  (explicit pathspec) and commit with `t490` in the message. This commit MUST precede M2's —
  verification-claim-integrity §2.3: the commit graph is the only sequencing witness.
- **M2 (Priority High) — Apply the two-line fix.** Delete line 5 (`status: completed`) from
  `plan.md` and `acceptance.md`. Nothing else.
- **M3 (Priority High) — Verify.** Run the E1–E4 batch (GREEN → mutant → re-fix → diff scope), one
  redirect per measurement, verdicts on counts.
- **M4 (Priority Medium) — Export evidence.** Write GREEN/mutant outputs + commit list to
  `.moai/reports/t490/`, commit with `t490` in the message.
- **M5 (Priority Medium) — Commit the SPEC directory.** `git add
  .moai/specs/SPEC-SPECLINT-ARTIFACT-STATUS-001/` (explicit pathspec; the directory is currently
  untracked) and commit with `t490` in the message — on this card's git-flow protocol, untracked
  artifacts never reach develop and are lost at worktree disposal.

## §G Anti-Patterns

- Reading the lint exit code as the verdict — it stays 1 on the residual warnings forever.
- "Fixing" the warnings while in the file (StatusTransitionInvalid / MissingExclusions are out of scope).
- `git add -A` or any sweep-stage; explicit pathspecs only.
- Reformatting the frontmatter (alphabetizing fields, changing quoting) — AC-SCOPE fails on any extra diff line.
- Committing evidence and fix in one commit — breaks AC-ORDERING.

## §H Cross-References

- RED baseline: `.moai/reports/t490/red-baseline.md`
- Schema rule violated: `.claude/rules/moai/development/spec-frontmatter-schema.md` § Artifact Statelessness
- Ordering doctrine: `.claude/rules/moai/core/verification-claim-integrity.md` §2.3
- Offending close commit: `185ce7d57` (card t462); historical warning source: `ce779f9ee` (#939, out of scope)
- CI invocation: `.github/workflows/spec-lint.yml:58` (`go run ./cmd/moai spec lint --strict`)
