---
id: SPEC-RELEASE-PIPELINE-001
title: "Release pipeline coverage and cost — shared classification preservation, input-gated cost jobs, stage budgets, and the drafter-cleanup fetch and filter repair"
version: "0.4.0"
status: draft
created: 2026-10-17
updated: 2026-10-22
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: ".github/workflows, internal/template"
lifecycle: spec-anchored
tags: "ci, release-pipeline, paths-filter, input-classification, cost, budget, jq, release-drafter, card-t1536"
tier: M
era: V3R6
card: t1536
related_specs: [SPEC-CANDIDATE-CI-001, SPEC-CI-DOCTOR-BIN-001, SPEC-V3R4-CI-FASTTRACK-001, SPEC-APPJS-FIRE-GUARD-001, SPEC-GITHUB-FLOW-CI-RESIDUE-001, SPEC-CI-RELEASE-VERIFY-SHARD-001]
---

# SPEC-RELEASE-PIPELINE-001

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-10-17 | manager-spec | Initial plan-phase authoring (card t1536, Class C, operator instruction 10-06, codex-analysis basis). Five card items re-anchored to HEAD `0bd1da4973d1d9539172f90587ca1ee84c8d9426`: item 1 found already landed on main under this card id (commit `0990c922f`, PR #1770, plus t1478 follow-ups `446eaee95`/`a73559090`/`e588f9882`), and the graph-freshness half of item 3 found already landed under card t1298 (`7a9d1fa29`/`0ec7c7679`); the remaining open work is items 2, 4, 5, the release-workflow half of item 3, and preservation of the landed surfaces. Item 5's jq defect reproduced RED in this tree (acceptance.md §B EV-3/EV-4). |
| 0.2.0 | 2026-10-17 | manager-spec | Plan-audit iter-1 repair (verdict FAIL 0.69, `.moai/reports/t1536/plan-audit-iter1.md`): D1 — both required-name SSoT surfaces named in plan M2/§D-11/AC-004; D2 — dry-run dispatch input wired into M5 scope; D3 — REQ-RPC-012 pins `cancel-in-progress: false` on BOTH workflows; D4 — REQ-RPC-009 redefined (local preserve-before-retry + post-wrapper evidence upload), AC-009 reclassified release-blocking; D5 — REQ-RPC-001 fail-open clause scoped to the release consumer; D6/MP-8 — RED-now cells EV-7/EV-8/EV-9 added. |
| 0.3.0 | 2026-10-17 | manager-spec | Plan-audit iter-2 repair (verdict FAIL 0.69, `.moai/reports/t1536/plan-audit-iter2.md`): D7 — explicit boolean→"1" conversion at the `MOAI_CLEANUP_DRY_RUN` wiring point; D8 — ref-independent literal group name on both draft-mutating workflows; D9 — `.github/required-checks.yml` restated as the single required-name SSoT, gtmpl reclassified shape-only PRESERVE; O5 — retry-wrapper anchors corrected to the seven measured ci.yml sites. |
| 0.4.0 | 2026-10-22 | manager-spec | Hold-and-split applied (plan-audit iter-3 FAIL 0.69 at the Tier M ceiling, `.moai/reports/t1536/plan-audit-iter3.md`; operator-approved split relayed by leader lane-17, recorded in `.moai/reports/t1536/progress.md` § Split 결정). Scope narrowed to M1 (verify-and-preserve) + M2 + M4 + M5. **Removed:** card item 3 in full — M3, REQ-RPC-005 (read-only cancellation) and REQ-RPC-006 (graph-freshness slot-split guard), AC-RPC-005/006, EV-2, and the D11 concern — moved to a named follow-up card (Out of Scope below); identifiers retired in place, not reused. **Added (D10):** REQ-RPC-013 / AC-RPC-013 — the cleanup's listing fetch requests a field `gh release list` does not support (`targetCommitish`; reproduced `Unknown JSON field`, exit 1, EV-10), so the fetch step fails before the jq filter runs; repair replaces the fetch source with a field set the source supports while keeping the `release/*` exemption, with payload-shape mapping (decision-index Q7). **Re-anchored** to tree `bf0a0f35fc280e83c2f9fd76d0e2a5de39b7e8b7` (absorbed origin/main): `release-pr-multi-os.yml` was sharded 6 shards × 3 OS by card t1892 (SPEC-CI-RELEASE-VERIFY-SHARD-001, PR #1854) and every line citation into it moved; `.github/required-checks.yml` gained the `release_gate` block (card t1858), which re-opened decision-index Q4 and surfaced Q8 (push-event gating scope); `.github/test-input-filters.yml` gained one `go_code` root. ci.yml, codeql.yml, release-drafter.yml and release-drafter-cleanup.yml are byte-unchanged since the prior baseline. All RED-now cells re-measured at the new tree (acceptance.md §B). |

## §A Background (Context)

The release verification pipeline spans these GitHub Actions workflow files plus shared classification and required-name inputs:

- `.github/workflows/release-pr-multi-os.yml` — the 3-OS × 6-shard race matrix (`Release Verify (<os>, <shard>)`, 18 legs since card t1892) aggregated by `Release PR Multi-OS Gate`, the ONLY required ubuntu race verdict on `release/*` → main PRs.
- `.github/workflows/ci.yml` — PR/push CI: `detect` (paths-filter), `test`, six `test-race-*` legs, `guard-bundle`, `test-skip-marker`, `test-integration`, and the always-on cost jobs `lint` / `build` (5-leg cross-compile matrix) / `constitution-check` / `test-browser`.
- `.github/workflows/release-drafter-cleanup.yml` — weekly cleanup of stale Release Drafter drafts (deletes drafts > 30 days old, exempting `release/*` targets).
- `.github/workflows/release-drafter.yml` — the draft-list writer the cleanup must serialize with.
- `.github/test-input-filters.yml` — the single-source `go_code` classifier consumed by ci.yml, codeql.yml, and release-pr-multi-os.yml (landed under this card id, PR #1770), guarded by `internal/template/test_input_filter_parity_test.go`.
- `.github/required-checks.yml` — the single required-name SSoT: `branches.main.contexts`, `branches."ci/**".contexts` (candidate verdict), and since card t1858 the `release_gate` list (`Lint`, `Test (ubuntu-latest)`, `Build (linux/amd64)`) that `scripts/release.sh` demands on the main commit it tags, held a subset of main's contexts by `internal/config/required_checks_release_gate_test.go`.

**Landed before this revision** (verified by direct reads at tree `bf0a0f35fc280e83c2f9fd76d0e2a5de39b7e8b7`):

1. Card item 1 (shared test-input list) is LANDED; the shared file carries the evidence-backed root set, classification fails open on the release consumer, and the parity guard fails on consumer drift (EV-1 green, 2 tests swept).
2. The release matrix was sharded by card t1892 (SPEC-CI-RELEASE-VERIFY-SHARD-001, completed): each of 18 legs keeps `-timeout 35m` (release-pr-multi-os.yml:300/:302) inside the unchanged job `timeout-minutes: 40` (:121), with no retry wrapper and an empty-sweep refusal (:306). That SPEC fixed the per-shard ceiling and job budget as not-to-be-raised.
3. ci.yml's race family is the six-leg SPEC-CANDIDATE-CI-001 layout with `retry-flaky.sh` wired at seven sites (:222/:370/:422/:473/:525/:578/:635).

**Still open at this tree** (each verified by direct read; evidence in acceptance.md §B):

- Item 2: `lint` (ci.yml:772), `build` ×5 legs (:819), `constitution-check` (:889), and `test-browser` (:950) carry no input gate — every PR, including docs-only diffs, pays the full cost surface (EV-5).
- Item 4: the browser job runs its 17-minute green phase (:1013-:1018) and the mutation cycle (:1020-) inside one 20-minute job budget (:953) with no step-level ceilings; the release race step (:271-:307) carries no step-level ceiling, so the `always()` compress/upload steps (:309/:315) rely on the 5-minute remainder of the job budget (EV-6, EV-11). No ci.yml upload path carries the first-failure record `retry-flaky.sh` preserves (EV-9).
- Item 5: the cleanup's fetch step requests `targetCommitish` from `gh release list` (release-drafter-cleanup.yml:56-59), a field that command does not support — the step exits 1 before the filter (EV-10) — and the filter at :65 carries `now | floor as $now |`, which rebinds the iteration input to a number (EV-3). The deletion pass cannot execute at this tree. No re-query guards the delete loop (EV-7), and the workflow has no concurrency block (EV-8), while `release-drafter.yml`'s group (:19-21, `release-drafter-${{ github.ref }}`) carries `cancel-in-progress: true`.

## §B Problem Statement

Two failure classes remain in the release pipeline:

1. **Coverage/cost mismatch.** Verification coverage was repaired (item 1 landed), but the cost side still runs every always-on job on every PR class: the pipeline pays release-grade cost for docs-grade diffs.
2. **Budget and correctness defects inside the verification stages.** Long stages carry no stage ceiling of their own, leaving the `always()` evidence steps that make a failure diagnosable to the job remainder; the browser job's two phases compete for one budget; the retry wrapper's first-failure record never leaves the runner; and the drafter-cleanup workflow fails at its fetch step and would mis-evaluate at its filter step, so the safety-maintenance workflow this pipeline relies on cannot perform its deletion pass.

## §C Requirements (GEARS)

- **REQ-RPC-001** (shared classifier preservation) — Where a CI workflow classifies a PR diff through the `go_code` filter, the workflow shall consume the shared single-source filter at `.github/test-input-filters.yml`, and when any consumer stops referencing the shared file, retains an inline classifier copy, or the shared list loses a required evidence-backed root, the parity guard test shall fail the suite. Where the consuming workflow is the release verification matrix (`release-pr-multi-os.yml`), an unevaluable classification shall run the matrix — fail-open through the filter step's `continue-on-error: true` (:64) plus the `!= 'false'` gate (:120); the ci.yml and codeql.yml consumers keep their landed gating (`if: … == 'true'`, ci.yml:106; codeql.yml:53-56) unchanged, and this SPEC introduces no change to their evaluation-failure behavior (fail-open for the newly gated cost family is REQ-RPC-002's obligation).
- **REQ-RPC-002** (input-class gating) — **When** a pull request event is classified, the always-on cost jobs of ci.yml (`lint`, the `build` matrix, `constitution-check`, `test-browser`) shall run only when the PR diff positively matches at least one of that job's declared verification input classes, with the class definitions sourced from the same shared filter file as `go_code`; **when** the classification for a cost job cannot be evaluated, that job shall run (an unclassifiable diff never skips a cost job silently).
- **REQ-RPC-003** (gated-family verdict semantics) — Every required check name the gated cost-job family publishes shall be satisfied only through a positively observed outcome: **when** an input-relevant job in the family concludes `failure` or `cancelled`, its verdict shall fail; **while** a job is skipped under a positively observed non-matching classification, its verdict shall pass and name the intentional exclusion; a hard failure of the classification step shall fail the verdict rather than pass it as a skip. The verdict surface (retained per-job names or one aggregate) follows decision-index Q4.
- **REQ-RPC-004** (required-name continuity and 3-OS harness scope) — **When** the input-class gating lands, every name in `.github/required-checks.yml` — `branches.main.contexts`, `branches."ci/**".contexts`, and `release_gate` — shall remain satisfied on every event class (pull request of any classification, push to main, candidate push to `ci/**`), the `release_gate` ⊆ `branches.main.contexts` invariant shall hold, and harness-class inputs shall retain full multi-OS verification through the release matrix's shared classifier — input splitting shall narrow only ci.yml's always-on cost jobs on the event classes decision-index Q8 names, never the release 3-OS coverage keyed to the shared `go_code` filter.
- **REQ-RPC-005** [RETIRED] — moved to the follow-up card with the rest of card item 3 (v0.4.0 split; see Out of Scope). Identifier not reused.
- **REQ-RPC-006** [RETIRED] — moved to the follow-up card with the rest of card item 3 (v0.4.0 split; see Out of Scope). Identifier not reused.
- **REQ-RPC-007** (per-stage budget headroom) — Each long-running verification stage shall carry a per-stage time ceiling strictly below its job's `timeout-minutes`, sized so the `always()` evidence steps that follow it (stream compression, artifact upload) complete even after a stage-failing or near-ceiling run; a stage ceiling that consumes the job budget shall not pass review, and the release matrix's per-shard `-timeout 35m` and job `timeout-minutes: 40` shall not be raised (SPEC-CI-RELEASE-VERIFY-SHARD-001).
- **REQ-RPC-008** (browser phase split) — The browser fire-guard verification shall run its green phase and its mutation (red/restore) phase as separately bounded stages, so a slow green phase cannot consume the mutation phase's budget and a mutation-phase failure retains its own diagnosable evidence trail.
- **REQ-RPC-009** (retry discipline) — **When** a verification stage retries a failure, the retry shall target only transient failures admitted by the flaky registry; the first-failure record shall be preserved locally before the retry re-runs the stage (`scripts/ci/retry-flaky.sh` writes `<log-stem>.attempt1.json` at :166 before the re-run at :169), and the run's post-wrapper evidence upload shall include that first-failure record alongside the final stream.
- **REQ-RPC-010** (drafter-cleanup filter repair) — The cleanup filter shall bind its timestamp through a parenthesized `(now|floor)` so the release list remains the iteration input, and **when** the filter executes against a draft list it shall select exactly the drafts that are drafts, older than the threshold, and not targeting `release/*`.
- **REQ-RPC-011** (re-query before deletion) — Immediately before each deletion, the cleanup shall re-query that tag's draft state and shall delete only while it is still a draft, still older than the threshold, and still not targeting `release/*`; a draft whose state changed between listing and deletion shall be skipped and the skip recorded.
- **REQ-RPC-012** (drafter/cleanup serialization) — The Release Drafter workflow and the cleanup workflow shall serialize their draft-list mutations through one shared concurrency group whose key is a ref-independent literal name — the group key shall not vary with `github.ref`, so a dispatched drafter on any ref lands in the same group as the scheduled main cleanup — with `cancel-in-progress: false` on BOTH sides, so an incoming run — drafter or cleanup, on any ref — never cancels an in-flight run of the other (in a shared group the incoming workflow's own flag governs, so pinning the cleanup side alone cannot deliver the guarantee).
- **REQ-RPC-013** (cleanup fetch supplies the filter's fields) — The cleanup's listing fetch shall request only fields its fetch source supports, and the payload the filter reads shall supply the draft flag, the creation time, the tag name, and the target branch of every candidate draft under the names the filter reads (a source with a different payload shape is mapped to those names before the filter runs); **when** the target branch is unavailable for a draft, that draft shall not be selected for deletion — the `release/*` exemption shall never be dropped to make the fetch succeed.

## Out of Scope

### Out of Scope — card item 3: read-only verification cancellation and the D11 concern (follow-up card)

- REQ-RPC-005 / REQ-RPC-006, AC-RPC-005 / AC-RPC-006, milestone M3, and the plan-audit iter-3 concern D11 move to a follow-up card (operator issue pending; leader candidate list) per the v0.4.0 hold-and-split decision. That card owns: the `release-pr-multi-os.yml:20` `cancel-in-progress` flip, the read-only-workflow cancellation guard test, and the graph-freshness slot-split regression guard.
- Binding note carried to the follow-up: its cancellation requirement must be scoped so ci.yml's conditional cancellation (ci.yml:52-55 — no cancel on `refs/heads/main` and `refs/heads/develop`, cancel elsewhere) and SPEC-CANDIDATE-CI-001 REQ-CCI-014 are preserved; ci.yml must not be flipped to unconditional `cancel-in-progress: true`.
- This SPEC changes no `cancel-in-progress` value outside the two draft-mutating workflows of REQ-RPC-012.

### Out of Scope — candidate CI, race sharding, and flaky-retry ownership

- The `ci/**` candidate-branch flow, landing-check gating, `guard-bundle` family, the six-leg ci.yml race sharding, and the flaky registry/retry wrapper implementation — owned by SPEC-CANDIDATE-CI-001 (implemented, card t1478). This SPEC consumes and preserves those surfaces.
- The release matrix's 6-shard × 3-OS layout, its selector parity guard, its per-shard `-timeout 35m`, and its job `timeout-minutes: 40` — owned by SPEC-CI-RELEASE-VERIFY-SHARD-001 (completed, card t1892). REQ-RPC-007 adds a stage ceiling around them; it does not re-shard, rebalance, or raise them.
- Extending `retry-flaky.sh` wiring to the macOS/Windows release legs — blocked on measured flaky signatures for those platforms (decision-index Q6, EVIDENCE-NEEDED).

### Out of Scope — release-targeting PR trigger expansion

- `.github/required-checks.yml` `release/*` contexts stay `[]`. The file's comment names "trigger expansion" as release-pipeline scope (t1536); the v0.4.0 split fixed this card's scope at M2/M4/M5, and none of them expands a workflow's `pull_request` branches to release-base PRs. A later card that expands triggers revisits that list.

### Out of Scope — doctor and shell-guard domains

- The doctor agent-emit embed verdict surface (SPEC-CI-DOCTOR-BIN-001, completed) and the protected-zone shell guard parser (SPEC-ZONE-SHELL-PARSING-001, completed) — disjoint modules.

### Out of Scope — deploy-path concurrency and release execution

- `release.yml` (the deploy/release execution path) and its no-cancel semantics are untouched.
- Release Drafter configuration content (label/category taxonomy) — this SPEC touches only the cleanup workflow's fetch, filter, delete-loop safety, dry-run reachability, and the drafter/cleanup concurrency relationship.

### Out of Scope — new verification content

- No new test suites, jobs, or verification axes are added to the verified product. This SPEC reorganizes when existing verifications run, how their budgets are bounded, and repairs one workflow; the verified content itself is unchanged (browser ADDITIVE-ONLY constraint from SPEC-APPJS-FIRE-GUARD-001 preserved).

## §E Preservation Constraints (summary)

The PRESERVE list (files and behaviors this SPEC must not regress) is enumerated in plan.md §D and is normative for the run phase. Headline invariants: the shared filter + parity guard; `release-pr-gate`'s cancelled-never-passes semantics; the release matrix's sharded layout and not-raised ceilings; the release/* race exclusion rationale; `test-skip-marker` intentional-skip reporting; the required-checks.yml name sets including `release_gate`; least-privilege `permissions:` blocks; the `|| rc=$?` census pattern and `-json` stream artifact chain.

## §F Cross-References

- plan.md §A carries the landed-vs-open re-anchor table and the line-citation drift table (0bd1da497 → bf0a0f35f).
- acceptance.md §B carries the evidence ledger, every live cell pinned to tree `bf0a0f35fc280e83c2f9fd76d0e2a5de39b7e8b7`.
- decision-index.md carries the assembly decisions (Q1-Q8) routed per the decision-gate labels (interview.yaml `decision_gate: on`).
