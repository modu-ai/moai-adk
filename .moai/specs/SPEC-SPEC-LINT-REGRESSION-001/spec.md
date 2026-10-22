---
id: SPEC-SPEC-LINT-REGRESSION-001
title: "Spec-lint baseline regression repair: restore the SPEC-PROGRESS-RECORD-IO-001 canonical status record and re-verify the main-push spec-lint gate green"
version: "0.2.0"
status: draft
created: 2026-10-23
updated: 2026-10-23
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: ".moai/specs"
lifecycle: spec-anchored
tags: "spec-lint, baseline-gate, status-transition, lifecycle-record, vacuous-assertion, regression-repair"
tier: M
---

# SPEC-SPEC-LINT-REGRESSION-001

## §A — History

- **2026-10-23 (v0.1.0)** — plan-phase authored from factory card t1929 (Class C, plan phase from scratch, P1). Scope: the common-cause repair for the red main-push spec-lint baseline gate. Task 1 identification landed during plan phase with parser-grade evidence (three independent lines: source mechanism, corpus era census, parser JSON) — the one non-advisory StatusTransitionInvalid on the merged tree `c3c86e795d` is SPEC-PROGRESS-RECORD-IO-001 (pair `completed → implemented`, commit `fa1f04b9f4`); this SPEC specifies its status-record repair, the verification of the already-landed 13-vacuous-assertion repair set on the three release-branch SPECs, and the green-gate re-measurement. Tier M: 8 requirements (ceiling 16) and 8 acceptance criteria (ceiling 16). Operator rulings carried from the card as hard constraints: the baseline file is never raised (기준선 상향 금지), the lint rules are never edited (the rules are correct; the SPEC records drifted), and recorded git history is never rewritten (forward commits only).

- **2026-10-23 (v0.2.0)** — plan-audit iteration 1 repairs (verdict FAIL 0.76, receipt `rcpt-87f3623a41a8217a1e32909b`; blocking B1, B2 and the same-delta N1, applied per fix_scope). (B1) The gate-green success assertion no longer demands the per-rule line `StatusTransitionInvalid: recorded 0 -> current 0`: that line prints only on the gate's EXCEEDED branch (`internal/cli/spec_lint.go:255-259` — `len(Increases) > 0`), so a correctly repaired tree (0 = 0) prints no such line and the former wording made every correct repair permanently fail the criterion (the impossible direction, verification-completeness §2). The success assertion at every echo site (§C Goal, REQ-SLR-007, plan.md §E item 2 and M2, AC-SLR-007, its green-path cell, and §F DoD) is restated in the observable green form: exit 0, the `baseline: OK` verdict line for `.moai/spec-lint-baseline.json`, its inventory line (the non-advisory tracked set unchanged at its recorded TierArtifactMissing 4 — the repaired rule drops out of the tracked set, so the total is stated unnumbered), zero StatusTransitionInvalid findings for SPEC-PROGRESS-RECORD-IO-001 in the JSON census, and no rule increase — whose absence the OK verdict plus exit 0 assert, because any increase produces the EXCEEDED verdict and exit 1 instead. (B2) AC-SLR-005's trailing no-edit clause is scoped to the five files repaired at `17a3b98de2` — the card's only edit among the three SPEC directories is SPEC-PROGRESS-RECORD-IO-001/spec.md, M1's mandated status-record repair (REQ-SLR-002/003) — resolving the former contradiction. (N1, applied in the same delta) The baseline byte-identity check is pinned to the commit form `git diff <repair-start-sha> -- .moai/spec-lint-baseline.json` — the bare `git diff -- <path>` form reads worktree-vs-index and misses a staged change (auditor-measured). Requirement and AC counts unchanged (8 / 8).

## §B — Problem

The main push run 39680374012 (head `bf0a35f10e`) failed the spec-lint baseline gate. The lane's preserved measurement on the merged tree `c3c86e795d` (`.moai/reports/t1929/baseline-lint-measurement.txt`) records the verdict verbatim: `baseline: EXCEEDED — .moai/spec-lint-baseline.json`, `StatusTransitionInvalid: recorded 0 -> current 1 (+1)`, `inventory: 8557 warning(s) total (advisory included), 5 non-advisory tracked across 1 recorded rule(s)`, with 0 errors. Because one non-advisory finding stands on the main baseline comparison, every open pull request touching SPEC input (24 at card time) shows the same `baseline: EXCEEDED` — one record defect reds a repo-wide gate.

1. **The gated defect, identified.** The baseline ratchet (`internal/spec/lint_baseline.go` § CompareBaseline) gates on per-rule NON-advisory warning counts; a rule absent from the checked-in baseline has a recorded count of 0, so its first occurrence is an increase. The Advisory flag on a StatusTransitionInvalid warning is set by exactly one path: `applyEraDemotion` (`internal/spec/lint.go:412-430`) marks every warning advisory on a protected SPEC — protected means grandfather-era (V2.x / V3R2-R4 / V3R5) OR terminal lifecycle status (`terminalStatusEnum` at `internal/spec/lint.go:1766-1771`: superseded, archived, rejected, completed). Active modern-era non-terminal SPECs pass through untouched. Measured on tree `c3c86e795d`: 112 StatusTransitionInvalid warnings fire across 111 SPECs; 110 of those SPECs are protected (terminal status, or grandfather era — including all 15 modern-dated non-terminal candidates, 14 of which carry no V3R6 progress.md markers and classify V2.x/V3R2-R4 under heuristics H-1/H-2). Exactly one SPEC fails both protection tests: SPEC-PROGRESS-RECORD-IO-001 — created 2026-10-08, phase v3.2.0, full V3R6 progress.md (§E.2 + §E.4 + sync_commit_sha present, H-4 modern), frontmatter status `implemented` (non-terminal). Parser confirmation: the tree-built linter's full-corpus JSON reports exactly one non-advisory StatusTransitionInvalid, and a targeted run reports it `advisory=false`: `SPEC SPEC-PROGRESS-RECORD-IO-001 status transition "completed" → "implemented" is not a canonical lifecycle edge (commit fa1f04b9f4)`.
2. **The record defect's origin.** The canonical transition set (`internal/spec/lint_transition.go:65-77`) admits `completed → in-progress` as the declared amendment edge (matrix row 7) and `implemented → completed` (row 3), but NOT `completed → implemented`. Commit `8d10007ca9` (2026-10-08) closed SPEC-PROGRESS-RECORD-IO-001 with the 3-phase close (status `completed`, progress.md §E.4 `sync_commit_sha: 8d10007ca9804eeb58bcb0a34aca922989603198`). Commit `fa1f04b9f4` (2026-10-09, card t1598) then reverted the status to `implemented` "pending AC-CI-007 CI evidence" — an in-fact amendment reopen recorded with the wrong token. The parking itself is deliberate and recorded (progress.md §E.4: `sync_status: sync-ready-pending-CI`; the completed transition re-lands when AC-CI-007's landing-flow evidence arrives); the TOKEN is the defect. Because the lint reads the most recent recorded transition from git history (`internal/spec/lint_ownership.go:204-247`, newest-first walk), no file edit clears the finding — only a forward status transition with a canonical pair does. Card t1914 (PR #1859, commit `17a3b98de2`) already repaired the 13 vacuous assertions on main and named this finding "not file-repairable and escalated separately" — that escalation is this SPEC.
3. **The 13 release-branch findings are staleness, not main defects.** The card's second task — 13 VacuousTestAssertion findings on the 6 release-branch PR heads (#1828 #1843 #1844 #1848 #1853 #1858; SPEC-FACTORY-RECORD-REPAIR-001 6, SPEC-PROGRESS-RECORD-IO-001 5, SPEC-GLM-JEV-KEY-001 2) — is already satisfied on main: commit `17a3b98de2` landed the repair set (5 files, 14 changed line-pairs: 6 outcome-assertion delimiter sites in SPEC-FACTORY-RECORD-REPAIR-001/acceptance.md, 6 test-selector anchoring sites across SPEC-PROGRESS-RECORD-IO-001, 2 in SPEC-GLM-JEV-KEY-001 — the card counts 13 findings; one of the 14 pairs is the associated anchor-wording line). All six PR heads carry the identical pre-repair footprint (measured: the same 14-line numstat against `c3c86e795d` on each head), and the tree's own lint measurement reports zero VacuousTestAssertion findings for the three SPECs. The remaining work for this task is verification plus the propagation note — the release heads clear when they absorb main.

## §C — Goal

- Restore SPEC-PROGRESS-RECORD-IO-001's status record to a canonical lifecycle edge by forward commits only, ending in the truthful mid-amendment state (in-progress, amendment machinery complete) — never by rewriting history, raising the baseline, or editing lint rules.
- Verify on a real checkout that the 13 vacuous-assertion sites carry the repaired forms as landed at `17a3b98de2`, and record the release-branch propagation path.
- Re-measure the spec-lint baseline gate on the repaired tree and observe the observable green form — exit 0 with the `baseline: OK` verdict for `.moai/spec-lint-baseline.json`, the JSON census reporting zero StatusTransitionInvalid findings for SPEC-PROGRESS-RECORD-IO-001, and no rule increase — with the baseline file byte-identical to its pre-repair state.

## §D — Requirements (GEARS)

8 requirements (Tier M ceiling 16). `<ID>` is SPEC-PROGRESS-RECORD-IO-001 throughout. A "real checkout" is a working tree with git history present (`git log` resolves for the SPEC path) — a git-archive export has no history and reads the transition walk as empty, so every measurement requirement below is bound to a real checkout (card caution, from the lane's measurement discipline).

### D.1 — The status-record repair

- **REQ-SLR-001** (State-driven) — While the git-recorded most-recent status transition of `<ID>` reads `completed → implemented` (commit `fa1f04b9f4`), the spec-lint baseline gate shall report exactly one non-advisory StatusTransitionInvalid finding for it, and the repair shall eliminate that finding only by landing forward status-transition commits whose recorded pair is a canonical lifecycle edge (`internal/spec/lint_transition.go:65-77`, including the row-7 amendment edge and terminal targets) — never by editing `.moai/spec-lint-baseline.json`, never by editing any file under `internal/spec/`, and never by rewriting recorded history (no rebase, no filter, no revert of `fa1f04b9f4`).
- **REQ-SLR-002** (Event-driven) — When the repair re-records `<ID>`'s status, it shall do so in exactly two forward commits: first a commit transitioning the status from `implemented` to `completed` (the canonical row-3 edge) that restores the recorded close state of commit `8d10007ca9` — a record repair, not a new close claim; then a commit transitioning the status from `completed` to `in-progress` (the canonical row-7 amendment edge) that reopens the SPEC as the in-place amendment of that close, which is its factual state (AC-CI-007 landing-flow CI evidence still pending, recorded in progress.md §E.4). A single-commit `implemented` to `in-progress` transition is prohibited: that pair is outside the canonical set and would re-fire the finding under a different message.
- **REQ-SLR-003** (Ubiquitous) — The reopened amendment record shall carry the full in-place amendment machinery so the spec audit's SyncStatusDrift check exempts it: spec.md frontmatter `amendment_of: SPEC-PROGRESS-RECORD-IO-001` (self-referential), a `## Amendments` sub-section in the spec.md body recording the prior completed version (the `8d10007ca9` close), a `prior_completed_sha` equal to progress.md §E.4's recorded `sync_commit_sha` (`8d10007ca9804eeb58bcb0a34aca922989603198`), the rationale (AC-CI-007 landing-flow CI evidence pending; spec-lint StatusTransitionInvalid record repair), and the amendment scope.
- **REQ-SLR-004** (State-driven) — Throughout the repair, `<ID>`'s progress.md §E.4 content (the `sync-ready-pending-CI` state and the AC-CI-007 pending record) shall remain intact and uncontradicted: the repair changes the status TOKEN and adds the amendment record; it shall not record a completion claim, shall not mark AC-CI-007 satisfied, and shall not alter the pending-CI evidence record — the close re-lands through its own recorded path when AC-CI-007's evidence arrives.

### D.2 — The vacuous-assertion verification

- **REQ-SLR-005** (Event-driven) — When the run phase verifies the card's 13 VacuousTestAssertion sites (SPEC-FACTORY-RECORD-REPAIR-001: 6 outcome-assertion delimiter sites in acceptance.md; SPEC-PROGRESS-RECORD-IO-001: 5 test-selector anchoring sites across acceptance.md and plan.md; SPEC-GLM-JEV-KEY-001: 2 test-selector anchoring sites across acceptance.md and plan.md), it shall do so in the real checkout and shall observe both: the landed repair footprint of commit `17a3b98de2` present in the working tree (the space-delimited outcome-assertion form — the test name followed by a whitespace delimiter before the closing delimiter; the fully-anchored alternation selector form), and zero VacuousTestAssertion findings reported by the tree-built linter for the three SPECs. Where a site already satisfies this — measured at plan phase for all 14 line-pairs on `c3c86e795d` — the requirement is verification-only and no edit shall be made.
- **REQ-SLR-006** (State-driven) — While the six release-branch PR heads (#1828 `07d1e7ff5e`, #1843 `97533db7d`, #1844 `2316237ef`, #1848 `a037f6b98`, #1853 `8476cd424`, #1858 `a6d6ceb5b`) predate their absorption of `17a3b98de2`, they shall continue to report the +13 until they absorb main; propagating the repair into release branches is out of scope for this SPEC. The run phase shall record the propagation note (the identical 14-line-pair staleness footprint, measured per head) in the card's progress record, and shall open no release-branch PR and push no branch.

### D.3 — The green gate and measurement provenance

- **REQ-SLR-007** (Event-driven) — When the status-record repair has landed, running the tree-built linter with the checked-in baseline (the form the spec-lint workflow runs: the lint command with `--baseline .moai/spec-lint-baseline.json`) in the real checkout shall produce the observable green form: exit 0, the printed verdict line `baseline: OK — .moai/spec-lint-baseline.json`, its inventory line with the non-advisory tracked set unchanged at its recorded TierArtifactMissing 4 (the repaired rule drops out of the tracked set) and no new rule entering it, zero StatusTransitionInvalid findings for SPEC-PROGRESS-RECORD-IO-001 in the same run's JSON census, and no rule increase — the per-rule `recorded -> current` increase lines print only on the gate's EXCEEDED branch (`internal/cli/spec_lint.go:255-259`), so no such line may appear and the OK verdict with exit 0 is what asserts its absence. The baseline file shall be byte-identical to its pre-repair state: an empty `git diff <repair-start-sha> -- .moai/spec-lint-baseline.json` output, the commit-pinned form (N1 — the bare `git diff -- <path>` form reads worktree-vs-index and misses a staged change).
- **REQ-SLR-008** (Ubiquitous) — Every measurement cited by this SPEC's verification shall name the tree it was taken on (a commit SHA, never a branch name) and a linter built from that measured tree; a measurement taken through a git-archive export or any history-less tree is void, and a tool binary resolved from an installed path rather than built from the measured tree shall have its provenance stated beside the citation.

## §E — Constraints

- The operator ruling binds: the baseline file is never raised and never rewritten by this SPEC (기준선 상향 금지) — the repair clears the +1 by eliminating the finding, not by recording it.
- The lint rules are correct; the SPEC records drifted. No file under `internal/spec/` changes.
- Forward commits only; recorded history is immutable. The lane commits (the plan-phase SPEC author does not).
- Work happens in the card worktree only; nothing is pushed to any remote.
- The two repair commits are the minimum status-record surgery: any additional status transition on `<ID>` during this card's window re-fires the pair judgment and is prohibited.

## §F — Out of Scope

### Out of Scope — resolving AC-CI-007

- The landing-flow CI evidence SPEC-PROGRESS-RECORD-IO-001 parks on is t1598's acceptance surface, not this card's. This SPEC restores the record; it neither triggers the release-pr-multi-os workflow nor records its verdict.

### Out of Scope — closing SPEC-PROGRESS-RECORD-IO-001

- No `implemented`/`in-progress` to `completed` close lands in this card's window beyond REQ-SLR-002's intermediate restore commit; the close re-lands through the SPEC's own recorded path when AC-CI-007 evidence arrives.

### Out of Scope — release-branch propagation

- Amending the six release-branch PR heads, their base branch, or any release branch is outside this card.

### Out of Scope — baseline and rule changes

- Raising, reshaping, or re-deriving `.moai/spec-lint-baseline.json`; changing any lint rule, severity, or demotion cause under `internal/spec/`.

### Out of Scope — the other 111 advisory StatusTransitionInvalid records

- The advisory findings on grandfathered and terminal SPECs are demoted by design; no bulk record normalization runs here.

## §G — Known Relations and Non-Goals

- Card t1914 (PR #1859, commit `17a3b98de2`) is the direct predecessor: it landed the 13-vacuous-assertion repair set on main and explicitly escalated the StatusTransitionInvalid +1 as "not file-repairable". This SPEC is that escalation's execution; no work item of t1914 is repeated here.
- The spec audit SyncStatusDrift exemption for in-place amendments (`spec-frontmatter-schema.md` § Status Enum amendment note) is the mechanism REQ-SLR-003 relies on; the exemption requires all four conditions, which is why the amendment record is complete rather than a bare token change.
- The gate stays green only while no NEW non-advisory warning lands with a recorded count of 0. The repair commits touch frontmatter only; the M2 gate re-measurement is the check that the repair introduced no such rule.

## §H — Cross-References

- `.moai/reports/t1929/baseline-lint-measurement.txt` — the lane's preserved 8788-line measurement on tree `c3c86e795d` (verbatim gate verdict at lines 8783-8785).
- `internal/spec/lint_baseline.go` § CompareBaseline — the ratchet: recorded 0 for an absent rule, first occurrence is an increase.
- `internal/spec/lint_transition.go:65-77` — the canonical edge set and matrix-row citations.
- `internal/spec/lint.go:412-430` (applyEraDemotion), `:1766-1771` (terminalStatusEnum) — the advisory boundary this defect was found through.
- `internal/spec/lint_ownership.go:204-247` — the newest-first transition walk that makes the finding history-recorded.
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix — rows 3 and 7 and the amendment record requirements.
- Factory card t1929; predecessor card t1914 (PR #1859).
