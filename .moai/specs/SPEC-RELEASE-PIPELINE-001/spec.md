---
id: SPEC-RELEASE-PIPELINE-001
title: "Release pipeline coverage and cost — shared classification preservation, input-gated cost jobs, verification concurrency, stage budgets, and the drafter-cleanup filter repair"
version: "0.1.0"
status: draft
created: 2026-10-17
updated: 2026-10-17
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: ".github/workflows, internal/template"
lifecycle: spec-anchored
tags: "ci, release-pipeline, paths-filter, input-classification, concurrency, cost, budget, jq, card-t1536"
tier: M
era: V3R6
card: t1536
related_specs: [SPEC-CANDIDATE-CI-001, SPEC-CI-DOCTOR-BIN-001, SPEC-V3R4-CI-FASTTRACK-001, SPEC-APPJS-FIRE-GUARD-001, SPEC-GITHUB-FLOW-CI-RESIDUE-001]
---

# SPEC-RELEASE-PIPELINE-001

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-10-17 | manager-spec | Initial plan-phase authoring (card t1536, Class C, operator instruction 10-06, codex-analysis basis). Five card items re-anchored to HEAD `0bd1da4973d1d9539172f90587ca1ee84c8d9426`: item 1 found already landed on main under this card id (commit `0990c922f`, PR #1770, plus t1478 follow-ups `446eaee95`/`a73559090`/`e588f9882`), and the graph-freshness half of item 3 found already landed under card t1298 (`7a9d1fa29`/`0ec7c7679`); the remaining open work is items 2, 4, 5, the release-workflow half of item 3, and preservation of the landed surfaces. Item 5's jq defect reproduced RED in this tree (acceptance.md §B EV-3/EV-4). |

## §A Background (Context)

The release verification pipeline spans four GitHub Actions workflow files plus one shared classification input:

- `.github/workflows/release-pr-multi-os.yml` — the 3-OS race matrix that is the ONLY required ubuntu race verdict on `release/*` → main PRs (`Release PR Multi-OS Gate`, required).
- `.github/workflows/ci.yml` — PR-time CI: `detect` (paths-filter), `test`, six `test-race-*` legs, `guard-bundle`, `test-skip-marker`, `test-integration`, and the always-on cost jobs `lint` / `build` (5-leg cross-compile matrix) / `constitution-check` / `test-browser`.
- `.github/workflows/graph-freshness.yml` — codemaps stamp reachability + freshness, with a `pull_request_target` conflict-guard job.
- `.github/workflows/release-drafter-cleanup.yml` — weekly cleanup of stale Release Drafter drafts (deletes drafts > 30 days old, exempting `release/*` targets).
- `.github/test-input-filters.yml` — the single-source `go_code` classifier consumed by ci.yml, codeql.yml, and release-pr-multi-os.yml (landed under this card id, PR #1770), guarded by `internal/template/test_input_filter_parity_test.go`.

**Landed since the 10-06 card text** (verified via `git log -S "card t1536"` and direct file reads at HEAD `0bd1da497`):

1. Card item 1 (shared test-input list; template/config release PRs skipping all race verification) is LANDED. The stale narrow inline filter is gone; the shared file carries the evidence-backed root set (`.moai/**`, `internal/template/templates/**`, catalog, `.claude/**`, census/release scripts, workflows, the filter file itself); classification fails open to running the matrix; a parity guard test fails on consumer drift.
2. The graph-freshness half of card item 3 is LANDED under card t1298: the concurrency group is split by event kind (`pull_request` vs `pull_request_target`) and PR number, and the conflict-guard job carries its own group.
3. ci.yml's race family was restructured by SPEC-CANDIDATE-CI-001 (card t1478) into six legs with per-leg ceilings and a flaky-registry retry wrapper (`scripts/ci/retry-flaky.sh`) wired into the `test` job and the rosterguard leg.

**Still open at this HEAD** (each verified by direct read; evidence in acceptance.md §B):

- Item 2: `lint` (ci.yml:772), `build` ×5 legs (:819), `constitution-check` (:889), and `test-browser` (:950) carry no input gate — every PR, including docs-only diffs, pays the full cost surface (golangci-lint installed from source, 5 cross-compiles, browser Chrome provisioning + 17-minute green phase + mutation cycle).
- Item 3 (release half): release-pr-multi-os.yml:20 still reads `cancel-in-progress: false` — superseded prior-HEAD verification runs pile up instead of being replaced, unlike the read-only `pr-multi-os-gate.yml` (:25, `true`) and `graph-freshness.yml` (:26, `true`) precedents already in the tree.
- Item 4: the browser job runs its 17-minute green phase and the mutation (red/restore) cycle inside one 20-minute job budget, and the release matrix runs a single 35-minute race step inside a 40-minute job whose `always()` census/compress/upload steps compete for the remaining headroom.
- Item 5: release-drafter-cleanup.yml:65 carries `now | floor as $now |` — the jq binding rebinds the iteration input from the releases array to a number, the filter errors (`Cannot iterate over number`, reproduced at this HEAD), and the cleanup this workflow exists for silently never executes; no re-query guards the delete loop, and the workflow has no concurrency relationship with `release-drafter.yml`.

## §B Problem Statement

Two failure classes remain in the release pipeline:

1. **Coverage/cost mismatch.** Verification coverage was repaired (item 1 landed), but the cost side still runs every always-on job on every PR class, and superseded read-only verification runs are not replaced on the newest HEAD. The pipeline pays release-grade cost for docs-grade diffs, and pays it again for stale HEADs.
2. **Budget and correctness defects inside the verification stages.** Long stages leave no headroom for the `always()` evidence steps that make a failure diagnosable (census, stream compression, artifact upload); the browser job's two phases compete for one budget; and the drafter-cleanup filter is a reproduced silent no-op, so the safety-maintenance workflow this pipeline relies on has not been executing its deletion pass at all.

## §C Requirements (GEARS)

- **REQ-RPC-001** (shared classifier preservation) — Where a CI workflow classifies a PR diff through the `go_code` filter, the workflow shall consume the shared single-source filter at `.github/test-input-filters.yml`, and when any consumer stops referencing the shared file, retains an inline classifier copy, or the shared list loses a required evidence-backed root, the parity guard test shall fail the suite. When the classifier cannot evaluate a diff, the consumer shall run the verification (uncertain classification never skips).
- **REQ-RPC-002** (input-class gating) — The always-on cost jobs of ci.yml (`lint`, the `build` matrix, `constitution-check`, `test-browser`) shall run only when the PR diff positively matches at least one of that job's declared verification input classes, with the class definitions sourced from the same shared filter file as `go_code`; **when** the classification for a cost job cannot be evaluated, that job shall run (an unclassifiable diff never skips a cost job silently).
- **REQ-RPC-003** (aggregate gate semantics) — The gated cost-job family shall report through exactly one fail-propagating aggregate check: **when** any input-relevant job in the family concludes `failure` or `cancelled`, the aggregate shall fail; **while** every relevant job is skipped under a positively observed non-matching classification, the aggregate shall pass and name the intentional exclusion; a hard failure of the classification step shall fail the aggregate rather than pass it as a skip.
- **REQ-RPC-004** (required-name continuity and 3-OS harness scope) — **When** the input-class gating lands, every branch-protection required check name shall remain satisfied on every PR class (no mergeability regression on any classification outcome), and harness-class inputs shall retain full multi-OS verification through the release matrix's shared classifier — input splitting shall narrow only ci.yml's always-on cost jobs, never the release 3-OS coverage keyed to the shared `go_code` filter.
- **REQ-RPC-005** (read-only cancellation) — **While** a verification workflow's runs mutate nothing outside CI state (no deploy, no merge, no push), its concurrency group shall replace superseded runs on the newest HEAD (`cancel-in-progress: true`); the no-cancel rule shall survive only for workflows whose runs mutate or publish shared state.
- **REQ-RPC-006** (graph-freshness slot split preservation) — The graph-freshness concurrency slots shall stay split by event kind and PR number with the conflict-guard job in its own group, so a `pull_request_target` guard run is never cancelled by an ordinary verification run for the same PR and updates to another PR never cancel it.
- **REQ-RPC-007** (per-stage budget headroom) — Each long-running verification stage shall carry a per-stage time ceiling strictly below its job's `timeout-minutes`, sized so the `always()` evidence steps that follow it (census, stream compression, artifact upload) complete even after a stage-failing or near-ceiling run — a stage ceiling that consumes the job budget shall not pass review.
- **REQ-RPC-008** (browser phase split) — The browser fire-guard verification shall run its green phase and its mutation (red/restore) phase as separately bounded stages, so a slow green phase cannot consume the mutation phase's budget and a mutation-phase failure retains its own diagnosable evidence trail.
- **REQ-RPC-009** (retry discipline) — **When** a verification stage retries a failure, the retry shall target only transient failures admitted by the flaky registry, and the first-failure record shall be preserved (captured and uploaded) before any retry re-runs the stage.
- **REQ-RPC-010** (drafter-cleanup filter repair) — The cleanup filter shall bind its timestamp through a parenthesized `(now|floor)` so the release list remains the iteration input, and **when** the filter executes against a draft list it shall select exactly the drafts that are `isDraft`, older than the threshold, and not targeting `release/*`.
- **REQ-RPC-011** (re-query before deletion) — Immediately before each deletion, the cleanup shall re-query that tag's draft state and shall delete only while it is still a draft, still older than the threshold, and still not targeting `release/*`; a draft whose state changed between listing and deletion shall be skipped and the skip recorded.
- **REQ-RPC-012** (drafter/cleanup serialization) — The Release Drafter workflow and the cleanup workflow shall serialize their draft-list mutations through one shared concurrency group, and an incoming run shall never cancel an in-flight cleanup run — the destructive side queues, it is never cancelled mid-deletion.

## Out of Scope

### Out of Scope — candidate CI, race sharding, and flaky-retry ownership (SPEC-CANDIDATE-CI-001)

- The `ci/**` candidate-branch flow, landing-check gating, `guard-bundle` family, the six-leg race sharding, and the flaky registry/retry wrapper implementation — owned by SPEC-CANDIDATE-CI-001 (implemented, card t1478). This SPEC consumes and preserves those surfaces; it does not redesign them.
- Extending `retry-flaky.sh` wiring to the macOS/Windows release legs — blocked on measured flaky signatures for those platforms (decision-index Q6, EVIDENCE-NEEDED); not implemented under this SPEC without that evidence.

### Out of Scope — doctor and shell-guard domains

- The doctor agent-emit embed verdict surface (SPEC-CI-DOCTOR-BIN-001, completed) and the protected-zone shell guard parser (SPEC-ZONE-SHELL-PARSING-001, completed) — disjoint modules; the workflow-file jq defect in REQ-RPC-010 is a CI pipeline defect, not a shell-guard parsing defect.

### Out of Scope — deploy-path concurrency and release execution

- `release.yml` (the deploy/release execution path) keeps its no-cancel semantics; REQ-RPC-005 binds read-only verification workflows only.
- Release Drafter configuration content (label/category taxonomy in `release-drafter.yml` config) — this SPEC touches only the cleanup workflow's filter, delete-loop safety, and concurrency relationship.

### Out of Scope — new verification content

- No new test suites, jobs, or verification axes are added. This SPEC reorganizes when existing verifications run, how their budgets are bounded, and repairs one workflow's filter; the verified content itself is unchanged (browser ADDITIVE-ONLY constraint from SPEC-APPJS-FIRE-GUARD-001 preserved).

## §E Preservation Constraints (summary)

The PRESERVE list (files and behaviors this SPEC must not regress) is enumerated in plan.md §D and is normative for the run phase. Headline invariants: the shared filter + parity guard; `release-pr-gate`'s cancelled-never-passes semantics; the release/* race exclusion rationale; `test-skip-marker` intentional-skip reporting; the release matrix's ubuntu leg as the only required ubuntu race; least-privilege `permissions:` blocks; the `|| rc=$?` census pattern and `-json` stream artifact chain.

## §F Cross-References

- plan.md §A carries the landed-vs-open re-anchor table with 10-06 → current line-number mapping.
- acceptance.md §B carries the evidence ledger (EV-1..EV-6) pinned to tree `0bd1da4973d1d9539172f90587ca1ee84c8d9426`.
- decision-index.md carries the six assembly decisions routed per the decision-gate labels (interview.yaml `decision_gate: on`).
