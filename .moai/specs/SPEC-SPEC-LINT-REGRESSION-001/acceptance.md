---
id: SPEC-SPEC-LINT-REGRESSION-001
title: "Acceptance criteria — spec-lint baseline regression repair"
version: "0.3.0"
created: 2026-10-23
updated: 2026-10-23
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: ".moai/specs"
lifecycle: spec-anchored
tags: "spec-lint, baseline-gate, status-transition, lifecycle-record, vacuous-assertion, regression-repair"
---

# acceptance.md — SPEC-SPEC-LINT-REGRESSION-001

> Verification layer. Every entry is Given-When-Then and binary-testable. GEARS obligations live in `spec.md` §D; they are not restated here. Stateless on the status axis.

## §A — Traceability matrix

| AC | Requirement | Milestone | Severity | Class |
|---|---|---|---|---|
| AC-SLR-001 | REQ-SLR-001, REQ-SLR-007 | M2 | Must | release-blocking |
| AC-SLR-002 | REQ-SLR-002 | M1 | Must | release-blocking |
| AC-SLR-003 | REQ-SLR-003 | M1 | Must | release-blocking |
| AC-SLR-004 | REQ-SLR-004 | M1/M2 | Must | regression-guard (preserve) |
| AC-SLR-005 | REQ-SLR-005 | M2 | Must | regression-guard (already-green surface) |
| AC-SLR-006 | REQ-SLR-006 | M2 | Should | regression-guard (preserve) |
| AC-SLR-007 | REQ-SLR-007 | M2 | Must | release-blocking |
| AC-SLR-008 | REQ-SLR-008 | M2 | Must | regression-guard (provenance) |

## §B — Acceptance criteria

- **AC-SLR-001 — the finding is gone, by canonical pair only**. Given the repaired tree, when the tree-built linter runs targeted at SPEC-PROGRESS-RECORD-IO-001 in JSON form, then the output carries no StatusTransitionInvalid finding, and the SPEC's git-recorded most-recent transition pair reads `implemented → completed` — the canonical row-3 edge; and given the full repair-window history of the SPEC's spec.md, then no commit in it rebase-rewrites, reverts, or filter-edits the recorded past (the `fa1f04b9f4` commit remains reachable and untouched in history).
- **AC-SLR-002 — the single canonical-edge restore (executed at `35cccfc217`)**. Given the repair window's commit log for `.moai/specs/SPEC-PROGRESS-RECORD-IO-001/spec.md`, when the repair commit is read, then it shows the status line changing from `implemented` to `completed` (the canonical row-3 edge) and nothing else in frontmatter but `updated:`, the HISTORY record rides the same commit, and no commit in the window records any other status transition on this SPEC — the one-commit form the leader disposition of 2026-10-23 ordered (산출물 실제 보완 + the one-commit simplification; spec.md §A v0.3.0).
- **AC-SLR-003 — the HISTORY record is complete**. Given the repaired spec.md, when its new HISTORY row is inspected, then it names all five facts: the restored recorded close (commit `8d10007ca9`, progress.md §E.4's `sync_commit_sha` `8d10007ca9804eeb58bcb0a34aca922989603198`), the mis-toked revert it undoes (commit `fa1f04b9f4`, card t1598, pair `completed → implemented` outside the canonical set), the restored edge (Status Transition Ownership Matrix row 3), the leader disposition of 2026-10-23 (산출물 실제 보완 — not a new close claim), and AC-CI-007 remaining the SPEC's own pending surface with progress.md untouched; and with the SPEC terminal at `completed`, the audit's drift check is sheltered by the terminal-status class, so no amendment record is required (the original `amendment_of`/`## Amendments` machinery is retracted with the two-commit design).
- **AC-SLR-004 — the pending-CI record survives untouched**. Given the repair-start commit, when `<ID>/progress.md` is diffed against it at M1 close and again at M2 close, then the diff is empty — §E.4's `sync-ready-pending-CI` state, the AC-CI-007 PENDING row, and the recorded `sync_commit_sha` are byte-identical — and no artifact added by this card asserts AC-CI-007 satisfied or the SPEC closed.
- **AC-SLR-005 — the 13 sites verified on the real checkout**. Given the real checkout (git history resolvable for the three SPEC paths), when the tree-built linter runs over SPEC-FACTORY-RECORD-REPAIR-001, SPEC-PROGRESS-RECORD-IO-001, and SPEC-GLM-JEV-KEY-001, then zero VacuousTestAssertion findings are reported for all three, and the landed repair footprint of commit `17a3b98de2` is present — the 14 line-pairs enumerated in plan.md §A.2 carry the repaired forms (the space-delimited outcome-assertion delimiter; the fully-anchored alternation selectors); and the five files repaired at `17a3b98de2` — SPEC-FACTORY-RECORD-REPAIR-001/acceptance.md, SPEC-GLM-JEV-KEY-001/acceptance.md and plan.md, SPEC-PROGRESS-RECORD-IO-001/acceptance.md and plan.md (the plan.md §A.4 PRESERVE set) — receive no edit in this card's window, the card's only edit among the three SPEC directories being SPEC-PROGRESS-RECORD-IO-001/spec.md, M1's mandated status-record repair (AC-SLR-002/003).
- **AC-SLR-006 — propagation recorded, nothing pushed**. Given the card's progress record at M2, when it is read, then it carries the release-branch propagation note (the six PR head SHAs with the identical 14-line-pair staleness footprint against the repair tree, clearing when they absorb the repair); and given the card worktree's state at close, then no remote ref moved from this card: no push, no PR, no release-branch write.
- **AC-SLR-007 — the gate is green on the repaired tree**. Given the repair commit landed, when the spec-lint baseline gate (the lint command with `--baseline .moai/spec-lint-baseline.json`) runs in the real checkout with a linter built from the repaired tree, then the observable green form holds: it exits 0, the printed verdict line reads `baseline: OK — .moai/spec-lint-baseline.json`, no per-rule increase line appears anywhere in the output (those lines print only on the gate's EXCEEDED branch, `internal/cli/spec_lint.go:255-259`, so the OK verdict plus exit 0 is the assertion of their absence), and the inventory line reports the non-advisory tracked set unchanged at its recorded TierArtifactMissing 4 — the repaired rule drops out of the tracked set — with no new rule entering it; and when a targeted JSON lint runs as a separate invocation on the same pinned tree with the same tree-built binary (plan.md §E item 4 — one invocation cannot carry both outputs: the CLI refuses `--baseline` with `--json` by design, `internal/cli/spec_lint.go:182`, measured exit 3), then that census reports zero StatusTransitionInvalid findings for SPEC-PROGRESS-RECORD-IO-001; and when the commit-pinned diff `git diff <repair-start-sha> -- .moai/spec-lint-baseline.json` runs with the pre-flight-recorded SHA, then the output is empty.
- **AC-SLR-008 — provenance on every citation**. Given the M2 evidence report, when each cited measurement is read, then it names the tree commit SHA it was taken on and states the judging linter was built from that tree (or states the installed-build comparison where one was used), and no measurement in the report originates from a git-archive export or other history-less tree.

## §C — Edge cases (covered by the tests or explicitly excluded)

- **A moved tree mid-repair** — the pre-flight re-observation (plan.md §C) catches it: if the targeted JSON no longer shows the finding, the card premise changed and the run stops for re-scoping, never repairs blind.
- **A foreign commit lands adjacent to the repair commit** — the sequence requirement (AC-SLR-002) reads the SPEC-file log, which only this card's commit touches in-window; a foreign commit elsewhere in the tree does not interleave the pair.
- **The gate green but a new rule appears** — AC-SLR-007's "no new rule appearing" clause fails the gate even at exit 0's neighborhood: an increase from any rule is a failure, not just the repaired rule's line.
- **The legacy `planned` token and terminal targets** — both are lint-tolerated transitions (`lint_transition.go:80-100`); neither is admitted as a repair target by REQ-SLR-001's canonical-edge wording, which names the matrix rows, not the tolerances.

## §D — Quality gate criteria

- The card's own three plan-phase artifacts pass `spec lint` with zero findings (verified at plan phase; re-verified at M2 on the final content).
- No Go source, workflow, template, or baseline file changes anywhere in the card's diff: the card's full diff against its base is confined to this SPEC directory, `<ID>/spec.md`, and the card's evidence and progress records.
- Commit discipline: each repair commit stages exactly one path (`<ID>/spec.md`); conventional-commit subjects; no sweep staging.

## §E — Adoption classification (two-cell discipline)

### RED-now evidence ledger (tree `c3c86e795d`, measured 2026-10-23)

All four cells were captured on the real checkout at worktree `.moai/worktrees/t1929`, HEAD `c3c86e795d`, with a linter binary built from that tree (`go build -o <scratch>/moai ./cmd/moai` — build provenance per REQ-SLR-008; the canonical `go run ./cmd/moai ...` form is stated per cell and is invocation-equivalent).

- **E-SLR-R1 — the defect is live and non-advisory (flips AC-SLR-001, AC-SLR-002)**.
  - Command: `go run ./cmd/moai spec lint --json .moai/specs/SPEC-PROGRESS-RECORD-IO-001`
  - Verbatim stdout (raw, complete — two findings; note the StatusTransitionInvalid object carries no `advisory` key, i.e. non-advisory, while the companion StatusGitConsistency carries `"advisory":true`):
    ```json
    [{"file":".moai/specs/SPEC-PROGRESS-RECORD-IO-001/spec.md","line":1,"severity":"warning","code":"StatusGitConsistency","message":"SPEC SPEC-PROGRESS-RECORD-IO-001 frontmatter status 'implemented' disagrees with git-implied status 'in-progress'","advisory":true},{"file":".moai/specs/SPEC-PROGRESS-RECORD-IO-001/spec.md","line":1,"severity":"warning","code":"StatusTransitionInvalid","message":"SPEC SPEC-PROGRESS-RECORD-IO-001 status transition \"completed\" → \"implemented\" is not a canonical lifecycle edge (commit fa1f04b9f4954fe514961c61f4edb4a1386e9a00)"}]
    ```
  - Exit code: 0 (warning-severity findings; the defect is the baseline-gate delta, not an error).
  - Tree: `c3c86e795d`. Red reason: the pair `completed → implemented` is outside the canonical set and the finding is non-advisory — exactly the gated defect; it is red for the right stated reason and flips green only through M1's two commits.
- **E-SLR-R2 — the gate verdict (flips AC-SLR-007)**.
  - Command: `go run ./cmd/moai spec lint --baseline .moai/spec-lint-baseline.json`
  - Verbatim stdout — the gate's summary lines, preserved from the lane's gate run on this tree at `.moai/reports/t1929/baseline-lint-measurement.txt` lines 8783-8785:
    ```text
    baseline: EXCEEDED — .moai/spec-lint-baseline.json
      StatusTransitionInvalid: recorded 0 -> current 1 (+1)
      inventory: 8557 warning(s) total (advisory included), 5 non-advisory tracked across 1 recorded rule(s)
    ```
  - Exit code: 1 — the baseline-gate contract for an EXCEEDED verdict (CompareBaseline.Failed), independently corroborated by the red main-push CI run 39680374012 at head `bf0a35f10e`; the preserved capture records stdout.
  - Tree: `c3c86e795d`. Red reason: the +1 is the identified non-advisory finding of E-SLR-R1 — the flip is M1+M2.
- **E-SLR-R3 — the corpus census: exactly one (backs AC-SLR-001's uniqueness claim)**.
  - Command: `go run ./cmd/moai spec lint --json` (full corpus), counts extracted from the JSON report.
  - Observed output (deciding lines): `warnings total: 8557`; `non-advisory warnings by rule: {'TierArtifactMissing': 4, 'StatusTransitionInvalid': 1}`; `errors: 0`; the single non-advisory StatusTransitionInvalid names SPEC-PROGRESS-RECORD-IO-001 at commit `fa1f04b9f4`.
  - Exit code: 0. Tree: `c3c86e795d`.
- **E-SLR-R4 — release-head staleness identity (backs AC-SLR-005's verification-only framing and AC-SLR-006's note)**.
  - Command (per head): `git diff --numstat HEAD <head> -- .moai/specs/SPEC-FACTORY-RECORD-REPAIR-001/ .moai/specs/SPEC-PROGRESS-RECORD-IO-001/ .moai/specs/SPEC-GLM-JEV-KEY-001/` for heads `07d1e7ff5e`, `97533db7d`, `2316237ef`, `a037f6b98`, `8476cd424`, `a6d6ceb5b`.
  - Observed output: all six return the identical block — `6 6` FACTORY acceptance.md, `1 1` + `1 1` GLM-JEV-KEY acceptance.md + plan.md, `4 4` + `2 2` PROGRESS-RECORD-IO acceptance.md + plan.md (14 line-pairs).
  - Exit code: 0. Tree: `c3c86e795d`.

### Green-path cells

- **AC-SLR-001 / AC-SLR-002 / AC-SLR-003** flip at **M1**: the targeted JSON re-run reports no StatusTransitionInvalid; the repair commit `35cccfc217` records the canonical pair (`implemented → completed`, row 3); the HISTORY record carries the five facts of AC-SLR-003.
- **AC-SLR-007** flips at **M2**: the gate re-run on the repaired tree exits 0 with the printed `baseline: OK` verdict, its inventory line, no increase lines anywhere in the output, and the JSON census reporting zero StatusTransitionInvalid for SPEC-PROGRESS-RECORD-IO-001.
- **AC-SLR-004 / AC-SLR-005 / AC-SLR-006 / AC-SLR-008** are preserve, already-green, and provenance classes — green on the pre-repair tree by observation (E-SLR-R4; the empty baseline diff; the three-SPEC zero-VTA census), re-asserted at M2 on the repaired tree. They are regression-guards by the undecidable-RED disposition (no failing input exists on this tree for them); M2 re-measures each rather than carrying the plan-phase observation forward.

## §F — Definition of Done

- All Must ACs PASS with verbatim evidence per §E's attribution format (command + raw output + tree SHA); Should ACs PASS or carry a recorded gap with owner.
- The gate re-run on the repaired tree prints the `baseline: OK` verdict inside an exit-0 run with no increase lines, the JSON census reports zero StatusTransitionInvalid for SPEC-PROGRESS-RECORD-IO-001, and the release-branch propagation note exists in the card's progress record.
- The card's full diff touches only: this SPEC directory, `<ID>/spec.md`, and the card's evidence/progress records. Gaps stated explicitly; residual risk (the release heads stay red until they absorb the repair; AC-CI-007 remains t1598's pending surface) recorded, not silently dropped.
