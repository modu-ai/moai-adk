---
id: SPEC-CI-VERDICT-INTEGRITY-001
title: "CI verdict integrity — repair five success-misjudgment gates and align the required-checks SSoT with published check names"
version: "0.1.0"
status: draft
created: 2026-10-06
updated: 2026-10-06
author: manager-spec
priority: P1
phase: "v3.3.0"
module: ".github/workflows"
lifecycle: spec-anchored
tags: "ci, github-actions, verdict-integrity, required-checks, auto-merge, ci-watch, branch-protection"
tier: M
related_specs: [SPEC-V3R3-CI-AUTONOMY-001]
---

# SPEC-CI-VERDICT-INTEGRITY-001

## A. History

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-10-06 | manager-spec | Initial draft. Card t1534 (operator directive 2026-10-06, Class C, codex-analysis-based). All card claims re-verified as RED-now observations on tree `a158b4b5f` before adoption; evidence ledger in plan.md §B. Codex source (read-only provenance): `.moai/reports/t1534/codex-ci-analysis-2026-10-06.md` (primary checkout), measured on `main@ec13872f3`. |

Problem statement: five CI verdict surfaces misjudge failure-class results as success — a required release gate passes on a cancelled matrix, the auto-merge flow can merge a head that was never verified, an install test summary prints "All tests passed" when every test was cancelled, and the two required-checks consumers (validator, watch loop) silently fail open. Independently, the required-checks SSoT names checks that are no longer published and omits one that is. Every requirement below is grounded in a reproduced RED observation (plan.md §B, tree `a158b4b5f`).

## B. Requirements (GEARS)

### M1 — gate verdict repairs

- **REQ-CI-001** — **While** the `Release PR Multi-OS Gate` job is the required check that blocks a release PR merge, **when** the gate job evaluates its dependency results, the gate shall exit non-zero unless the detect job and the full matrix job both concluded `success`, or the matrix produced no run because of a positively observed intentional exclusion (a docs-only diff reported by the paths filter, or a non-release head branch) that the gate can name.

- **REQ-CI-002** — **When** the gate passes because of an intentional exclusion, the gate shall emit a distinct output line naming the exclusion and its reason, and shall not reuse the unconditional "verification PASSED" wording for that case.

- **REQ-CI-003** — **When** the auto-merge workflow issues its merge command, the workflow shall pin the merge to the exact CI-verified head commit as a hard merge precondition, and shall re-read the PR head and confirm it still equals that commit immediately before the merge call; on any mismatch or read failure the workflow shall withhold the merge.

- **REQ-CI-004** — **When** the auto-merge required-checks lookup fails, returns empty output, or returns a result that does not cover every check the target branch's protection requires, the workflow shall withhold the merge and record the incomplete observation, and shall not treat lookup absence as check completion.

- **REQ-CI-005** — **While** the auto-merge job waits for required checks and for the review verdict, the workflow shall enforce one overall deadline that is strictly smaller than the job's own timeout budget, and the merge step shall be unreachable once that deadline passes.

### M1 — install test verdict repairs

- **REQ-CI-006** — **When** the test-install summary job evaluates its dependency results, the summary shall exit non-zero unless every job it lists — the install-script-parity job included — concluded `success`, and shall treat `cancelled`, `timed_out`, and `skipped` as non-success.

- **REQ-CI-007** — **When** a required install-feature probe finds its target absent, the compatibility-check step shall exit non-zero; printing a "missing" marker with a zero exit shall not count as a verdict.

### M2 — required-checks SSoT

- **REQ-CI-008** — The required-checks SSoT shall list, per branch key, only context names that the workflows actually publish as check runs on that branch's pull requests. The `main` list shall contain `Release PR Multi-OS Gate` and `Analyze (Go) (go)`, and shall not contain `Test (macos-latest)`, `Test (windows-latest)`, or `CodeQL`. Grounding: live check-run names read from PR 1748 (plan.md §B ledger E13).

- **REQ-CI-009** — **When** the corrected SSoT values are ready to reach live branch protection, the run phase shall deliver the corrected file, the rendered apply payload, and the exact apply command to the operator/leader as a keep-set gate; the run phase shall not execute the GitHub branch-protection apply itself, and the post-apply GET re-verification (read-back of the live protection contexts) shall be the only accepted evidence that live state matches the SSoT. Re-applying protection from the pre-correction file values is prohibited.

- **REQ-CI-010** — **When** the validator runs, it shall exit non-zero when its YAML parser is unavailable or the SSoT file fails to parse, and shall verify that every listed required context is publishable by the workflows; an empty parse result shall never produce a passing verdict.

- **REQ-CI-011** — **When** a repository-side test reads an input file (for example the branch-protection parity test reading `.github/branch-protection.json.gtmpl`), the ci.yml detect filter shall match that path so the test runs, and a correspondence guard shall fail CI when any test-read input path is absent from the filter.

### M3 — ci-watch verdict integrity

- **REQ-CI-012** — The ci-watch poll shall request only check-run fields the installed gh CLI supports for `pr checks` (`name`, `state`, `bucket`, `link`), shall process the response as one JSON array, and shall classify each check by its `bucket` value; a field-validation rejection or any other gh failure shall exit non-zero and shall never be read as an all-pass signal.

- **REQ-CI-013** — **When** the ci-watch loop classifies a poll tick, it shall determine required checks from the SSoT required list for the PR's base branch (the existing `is_required` classifier), shall count an expected required check that is absent from the poll response as pending, and shall iterate check names without splitting them on whitespace.

## C. Acceptance Criteria

Tier M — the full Given-When-Then acceptance set lives in `acceptance.md` (14 criteria, AC-CI-001..AC-CI-014, each carrying a RED-now cell pinned to tree `a158b4b5f` and a green-path cell naming the flipping milestone). The evidence ledger that the RED cells cite is plan.md §B.

## D. Constraints

1. **Keep-set gate (irreversible external-shared operation).** The GitHub branch-protection apply is executed only by the operator or leader session. No run-phase agent applies protection, in any milestone. See REQ-CI-009 and plan.md §F (M2 keep-set subsection).
2. **Dev-repo paths only.** Touched files are limited to: `.github/workflows/release-pr-multi-os.yml`, `.github/workflows/auto-merge.yml`, `.github/workflows/test-install.yml`, `.github/workflows/required-checks.yml` (the SSoT lives at `.github/required-checks.yml`, not under `workflows/`), `.github/workflows/ci.yml`, `scripts/ci-mirror/validate-required-checks.sh`, `internal/template/branch_protection_parity_test.go`, `scripts/ci-watch/run.sh` (plus its sibling `lib/classify.sh` and `test/run_test.sh` only as needed to keep the M3 repair and its test consistent).
3. **No drive-by fixes** to any other workflow file; scope discipline per the card list only.
4. Local test rules: new Go test code uses `t.TempDir()`; temporary probe fixtures live under `/tmp`.
5. Toolchain grounding: the repaired field list must hold against the gh CLI's own published field set (observed 2026-10-06, ledger E12) — no field list is hardcoded beyond what the CLI validates.

## E. Out of Scope

### Out of Scope — template tree

- `internal/template/templates/.github/**` is excluded entirely — generic-value template fixes belong to card t1537. No template mirror is edited by this SPEC (verified: `required-checks.yml` has no mirror under the template tree; `branch-protection.json.gtmpl` is shape-only and takes contexts at render time).

### Out of Scope — GitHub-Flow transition residues (card t1535)

- `spec-lint.yml` `develop` references and the other GitHub-Flow-transition residues.
- The release-flow question of required multi-OS coverage for ordinary feature-to-main PRs (codex §C P1 item 2).

### Out of Scope — release-pipeline cost and coverage classification (card t1536)

- ci.yml job classification for docs-only PRs (lint / Build / constitution / browser jobs), binary-browser-harness input splitting, 3-OS harness scoping, mutation/browser step splitting, `release-pr-multi-os.yml` concurrency (`cancel-in-progress`) policy, and the release-pr paths-filter input-list gap (codex §A P1 item 2 — a real gap, deliberately excluded from this card's list).
- `release-drafter-cleanup.yml` jq expression repair (codex §A P2 item 9).

### Out of Scope — live API operations during run phase

- Any GitHub branch-protection read or write executed by a run-phase agent (reads needed for the pre-apply diff are packaged as operator-executable commands, per REQ-CI-009).

## F. Success Criteria

1. Every M1 gate fails closed on non-success conclusions (cancelled/timed_out included) and passes only on success or a named intentional exclusion — demonstrated by the acceptance.md criteria flipping from their RED cells (plan.md §B) to green on the repaired tree.
2. The SSoT lists only published context names; the corrected values reach live protection only through the operator keep-set gate with a post-apply GET read-back recorded as card evidence.
3. The validator and the watch loop fail closed on parser/fetch failure and classify required checks from the SSoT, with the watch loop's own test suite green.
4. `actionlint` stays clean on all touched workflow files; the affected Go package (`internal/template`) tests pass; no template-tree file changes (verified by `git status` at close).
