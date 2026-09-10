---
id: SPEC-SPECLINT-ARTIFACT-STATUS-001
title: "Remove forbidden status field from SPEC-CODEX-E2E-MEASURE-001 sibling artifacts to clear 2 spec-lint ERRORs"
version: "0.1.0"
status: completed
created: 2026-09-06
updated: 2026-09-07
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: ".moai/specs/SPEC-CODEX-E2E-MEASURE-001"
lifecycle: spec-anchored
tier: S
tags: "speclint, ci, frontmatter, artifact-statelessness, schema-repair"
---

# SPEC-SPECLINT-ARTIFACT-STATUS-001

## 1. Background

CI job `spec-lint` (run 34014859906, tree `615d18c1f`) FAILS. The lane reproduced locally with the
identical CI invocation `go run ./cmd/moai spec lint --strict`: **2 errors / 4,333 warnings**, exit 1.
The CLI fails the job on ERROR-severity findings only (`internal/cli/spec_lint.go:97`). The 2 errors:

- `ArtifactStatusFieldForbidden` at `.moai/specs/SPEC-CODEX-E2E-MEASURE-001/plan.md:5` — frontmatter carries `status: completed`
- same finding at `.moai/specs/SPEC-CODEX-E2E-MEASURE-001/acceptance.md:5`

Root cause: card t462's 3-phase close (commit `185ce7d57`) wrote `status: completed` into the sibling
artifacts, violating the schema's **Artifact Statelessness** rule
(`.claude/rules/moai/development/spec-frontmatter-schema.md` § Artifact Statelessness — non-spec.md
artifacts are stateless on the status axis). `spec.md` itself legitimately carries the status and is
untouched. The lint error message itself prescribes the exact repair: delete the `status:` line.

Full measured baseline: `.moai/reports/t490/red-baseline.md` (RED baseline, this run).

## 2. Requirements (GEARS)

- **REQ-001**: When the spec-lint `--strict` scan runs on the repaired tree, the lint engine shall report zero `ArtifactStatusFieldForbidden` ERROR-severity findings (grep count of `^ERROR` lines in the output = 0).
- **REQ-002**: The repair shall delete ONLY the `status: completed` line (line 5) from `.moai/specs/SPEC-CODEX-E2E-MEASURE-001/plan.md` and `.moai/specs/SPEC-CODEX-E2E-MEASURE-001/acceptance.md`; every other frontmatter field (`id`, `title`, `version`, `created`, `updated`, `author`, `tier`) shall remain byte-identical.
- **REQ-003**: The repair shall not modify any file other than the two named artifacts.
- **REQ-004**: When the two `status:` lines are restored from base `615d18c1f` (mutant), the lint engine shall report exactly 2 ERROR findings, both `ArtifactStatusFieldForbidden` — the anti-vacuous guard proving the zero-error state is caused by this repair, not pre-existing.
- **REQ-005**: The RED baseline artifact `.moai/reports/t490/red-baseline.md` shall be committed in a commit that precedes the fix commit in the branch history (ordering attribution per verification-claim-integrity §2.3).
- **REQ-006**: The AC verdicts shall be judged on `^ERROR` match counts of the lint output, never on the command's exit code — the command is EXPECTED to still exit 1 on the 4,333 residual WARNING-severity findings, and that exit code is not a failure of this SPEC.
- **REQ-007**: The repair shall export the GREEN and mutant lint outputs under `.moai/reports/t490/`, and every branch commit message on `WT-speclint-status-transition` shall carry `t490`.

## 3. Acceptance Criteria (inline — Tier S)

- **AC-GREEN**: Given the repaired tree, When `go run ./cmd/moai spec lint --strict` runs (output redirected to a file, exit code captured separately), Then the grep count of `^ERROR` on the output file = 0. The command MAY exit 1 on the ~4,333 warnings — that is EXPECTED and is not this AC's verdict.
- **AC-MUTANT**: Given the two `status:` lines restored via `git checkout 615d18c1f -- .moai/specs/SPEC-CODEX-E2E-MEASURE-001/plan.md .moai/specs/SPEC-CODEX-E2E-MEASURE-001/acceptance.md`, When the same command runs, Then the output carries exactly 2 lines matching `^ERROR`, both `ArtifactStatusFieldForbidden`. Then the fix is re-applied and AC-GREEN re-confirmed (count back to 0).
- **AC-ORDERING**: Given the branch history, When `git log --oneline` is read, Then the commit carrying `.moai/reports/t490/red-baseline.md` appears BEFORE the commit carrying the two-line fix (§2.3 — the commit graph, not a commit message, witnesses the ordering).
- **AC-SCOPE**: Given the branch diff for the two files, When `git diff <base>..HEAD -- .moai/specs/SPEC-CODEX-E2E-MEASURE-001/` is read, Then it shows exactly two deleted lines (`-status: completed`, one per file) and no other change; and the commits carrying the two-line repair modify no file other than the two named artifacts.
- **AC-EVIDENCE** (parent: REQ-007): Given `.moai/reports/t490/`, When the evidence files are read, Then the GREEN + mutant re-measurement outputs are recorded there, and every branch commit message carries `t490`.

## 4. Constraints

- Tier S — two artifacts (spec.md + plan.md, AC inline above). This is a two-line repair with a documented rationale.
- Blob pins at the measured tree `615d18c1f`: `plan.md` = `15d5ef6ae478b97fc301f9c778e472aa7bd6c8d8`, `acceptance.md` = `13d830137cd34b4fcb3616a92f0099e4880f2d54`. If the blobs at fix time differ, STOP and re-measure before editing (the premise has moved).
- All five numbers (2 errors / 4,333 warnings / 99 StatusTransitionInvalid / 25 MissingExclusions / 48 `ce779f9ee` citations) were measured by the lane in this run — recorded in the RED baseline, not re-derived here.

## 5. Exclusions

### Out of Scope — The 4,333 WARNING-severity findings

- `StatusTransitionInvalid` 99 findings (48 citing the deliberate 2026-05 batch-transition commit `ce779f9ee` = SPEC-V3R4-STATUS-DRIFT-FOLLOWUP-001 #939) — warnings correctly report history; repair needs rule-semantics design decisions (edge addition / exemption / per-SPEC fix) and is a separate card.
- `MissingExclusions` 25 findings — terminal-status downgraded, non-blocking.

### Out of Scope — History and rule-code edits

- No HISTORY-section edits to completed SPECs.
- No `spec.md` status changes — `SPEC-CODEX-E2E-MEASURE-001/spec.md` legitimately carries `status:` and stays untouched.
- No lint-rule code changes (`internal/spec/`, `internal/cli/spec_lint.go` untouched).
