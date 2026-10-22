---
id: SPEC-RELEASE-PIPELINE-001
title: "Release pipeline coverage and cost — implementation plan"
version: "0.6.1"
created: 2026-10-17
updated: 2026-10-23
author: manager-spec
---

# plan.md — SPEC-RELEASE-PIPELINE-001

## §A Context

Card t1536 (operator instruction 10-06, Class C). Worktree `.moai/worktrees/t1536`, branch `WT-10-06-codex`. Plan-phase tree for this revision: `bf0a0f35fc280e83c2f9fd76d0e2a5de39b7e8b7` (merge of origin/main into the card branch; prior baseline `0bd1da4973d1d9539172f90587ca1ee84c8d9426`).

**Scope after the v0.4.0 hold-and-split** (operator-approved; `.moai/reports/t1536/progress.md` § Split 결정): M1 (verify-and-preserve) + M2 + M4 + M5 (with D10). Card item 3 (M3, REQ-RPC-005/006, D11) moved to a follow-up card — spec.md § Out of Scope. Milestone numbers keep their card-item mapping; M3 is vacant, not renumbered.

### §A.1 Re-anchor table (card item → current tree)

| Card item | Current anchor at `bf0a0f35f` | State |
|---|---|---|
| 1. Filter lacks template/config inputs | Shared filter `.github/test-input-filters.yml` consumed at ci.yml:79, codeql.yml:43, release-pr-multi-os.yml:68; release/* race exclusions at ci.yml:340/394/446/497/549/602 (by design) | **LANDED** (PR #1770 + t1478 follow-ups; t1858 added one root) — M1 verifies and preserves |
| 2. Doc-only PRs run lint/build/constitution/browser | `lint` :772, `build` ×5 :819, `constitution-check` :889, `test-browser` :950 — no `needs:`/`if:` (EV-5) | **OPEN** — M2 |
| 3. Stale prior-HEAD verification + freshness slots | release-pr-multi-os.yml:20 `cancel-in-progress: false`; graph-freshness split landed (t1298) | **MOVED** to follow-up card (v0.4.0 split) — no milestone here |
| 4. Browser 17m/20m + mutation; release single long race | test-browser job 20m (:953), green step `-timeout 17m` (:1018) + mutation cycle (:1020-) in-job, no step ceilings; release race now sharded (t1892): per-leg step :271-:307 with `-timeout 35m` (:300/:302), job `timeout-minutes: 40` (:121), no step ceiling (EV-6, EV-11); retry wrapper at seven ci.yml sites, attempt1 record uploaded nowhere (EV-9) | **OPEN** — M4 |
| 5. jq `now\|floor as $now` never executes | fetch requests unsupported `targetCommitish` from `gh release list` (:56-59, EV-10 — D10); filter :65 broken form (EV-3/EV-4); no re-query in delete loop (:89-98, EV-7); no concurrency block (EV-8); release-drafter.yml:19-21 ref-keyed group with `cancel-in-progress: true` | **OPEN** — M5 |

### §A.2 Line-citation drift table (baseline `0bd1da497` → `bf0a0f35f`)

Measured by `git diff --stat 0bd1da4973d1d9539172f90587ca1ee84c8d9426 HEAD -- .github/ …` and `git diff --numstat` over the same range (changed: `release-pr-multi-os.yml` +96/−8 via t1892, `required-checks.yml` +8/−0 via t1858, `test-input-filters.yml` +6/−0 via t1858, plus three workflows this SPEC does not cite: claude-code-review.yml, claude.yml, label-sync.yml) and direct reads at `bf0a0f35f`.

| Citation (v0.3.0) | Now | Moved? |
|---|---|---|
| release-pr-multi-os.yml:20 `cancel-in-progress` | :20 | no (now out of scope) |
| release-pr-multi-os.yml:62-68 filter step (`continue-on-error` :64, `filters:` :68) | :62-68 | no |
| release-pr-multi-os.yml:93 `!= 'false'` matrix gate | :120 | **yes** |
| release-pr-multi-os.yml:94 job `timeout-minutes: 40` | :121 | **yes** |
| release-pr-multi-os.yml:217 race `-timeout 35m` | :300 (remainder) / :302 (selector shards); step :271-:307 | **yes** (single step → sharded step) |
| release-pr-multi-os.yml:274-315 `release-pr-gate` | :362-403 | **yes** |
| ci.yml :52-55 concurrency, :79, :102/:105/:106/:107, :222/:370/:422/:473/:525/:578/:635, :340…:602, :772, :819, :889/:893, :950/:953, :1013-:1018 | same | no (ci.yml byte-unchanged) |
| codeql.yml:43, :53-56 | same | no |
| release-drafter-cleanup.yml :24, :46, :65, :82-98, :92 | same; new citation :56-59 (fetch) | no |
| release-drafter.yml :14, :19-21 | same | no |
| scripts/ci/retry-flaky.sh :166 / :169 | same | no |
| test-input-filters.yml `go_code` roots | +6 lines at :79-84 (root `.github/required-checks.yml`, card t1858) | appended |
| required-checks.yml | +8 lines at :58-65 (`release_gate`, card t1858) | appended — new constraint (§D-11, decision-index Q4/Q8) |

### §A.3 Overlap boundaries (sibling SPECs)

- **SPEC-CANDIDATE-CI-001** (implemented, t1478): owns candidate-branch CI, guard-bundle, six ci.yml race legs, flaky registry/retry implementation, and REQ-CCI-014 (ci.yml conditional cancellation). Consumed and preserved.
- **SPEC-CI-RELEASE-VERIFY-SHARD-001** (completed, t1892): owns the release matrix sharding, its selector parity guard (`internal/template/release_verify_shard_parity_test.go`), and the not-raised ceilings (its Q8) and no-retry decision (its Q5). M4 adds a step-level ceiling around the race step without touching any of those.
- **SPEC-CI-DOCTOR-BIN-001** (completed, t346): disjoint.
- **SPEC-V3R4-CI-FASTTRACK-001**: origin of the release-gate aggregate semantics REQ-RPC-003 generalizes.
- **SPEC-APPJS-FIRE-GUARD-001**: owns the browser job's ADDITIVE-ONLY + phase semantics; M4's split preserves its verdict family and pinned-Chrome discipline.
- **SPEC-GITHUB-FLOW-CI-RESIDUE-001**: owns pr-multi-os-gate.yml — no longer cited after the item-3 move.

### §A.4 Development mode

`quality.yaml` → `constitution.development_mode: tdd`; coverage target 85. All milestone guards are RED-first Go tests in `internal/template/` (the established pattern: `test_input_filter_parity_test.go` reads the workflow YAML from the repo tree).

## §B Known Issues

- **B1 — golangci-lint from source**: the lint job installs golangci-lint via `go install`. Input-gating it (M2) changes when that cost is paid, not the check itself; do not swap the install method in this SPEC.
- **B2 — unreported required names**: a required name whose job is gated off and that nothing else publishes is never reported, and the PR cannot merge (the reason `test-skip-marker` exists, ci.yml:660-:716). `Lint` and the five `Build (…)` names are required on main and on `ci/**`, and `Lint` / `Build (linux/amd64)` are `release_gate` names; M2 must keep each set reported on the event class it governs — `branches.main.contexts` on pull requests to main, `branches."ci/**".contexts` on `ci/**` pushes, `release_gate` on pushes to main — per REQ-RPC-004 and decision-index Q4/Q8 (`Release PR Multi-OS Gate` exists only on pull_request/workflow_dispatch, acceptance.md EV-16).
- **B3 — workflow YAML is repo input**: guard tests parse workflow YAML from the module-root relative path (`../../.github/workflows/...` from `internal/template/`), as the parity test does; a copied fixture would drift.
- **B4 — `release_gate` subset invariant**: `internal/config/required_checks_release_gate_test.go` enforces `release_gate` ⊆ `branches.main.contexts`, and `scripts/release.sh` reads `release_gate` to decide which check runs to demand on the tagged main commit. Any required-name change (only under Q4's alternate) must keep both true in the same commit.
- **B5 — jq availability**: guard tests executing jq programs require the `jq` binary. If jq is absent locally the filter-semantics test fails, not skips.
- **B6 — shared group side effects (REQ-RPC-012)**: joining cleanup into a group with the push-triggered drafter can queue a Monday cleanup behind a draft-update burst; the cleanup job budget (15m) absorbs it. With both sides `cancel-in-progress: false`, rapid main pushes queue drafter updates serially. Record the queuing behavior in BOTH workflow comments.
- **B7 — gh field support drifts by version (D10)**: the `gh release list` JSON field set was measured on gh 2.92.0 (EV-10: no `targetCommitish`); `gh release view` supports it (EV-12). The runner's preinstalled gh may differ; the guard test pins the field contract from the workflow text, and the post-merge dispatch (§D.2 of acceptance.md) is the runner-side observation.
- **B8 — REST payload shape**: if decision-index Q7's alternate (`gh api …/releases`) is chosen, the payload is snake_case (`tag_name`, `draft`, `created_at`, `target_commitish` — measured, acceptance.md §B EV-13); the filter reads camelCase, so a missing mapping turns every predicate null and selects nothing — a silent no-op of the same class item 5 exists to fix.

## §C Pre-flight (baseline commands, re-run at `bf0a0f35f`)

```
$ go test ./internal/template/ -run '^(TestTestInputFiltersSingleSource|TestTestInputFiltersCoversGoTestInputs)$' -count=1 -v
--- PASS: TestTestInputFiltersSingleSource (0.01s)
--- PASS: TestTestInputFiltersCoversGoTestInputs (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/template	0.924s         # EV-1, exit 0, 2 swept

$ gh release list --json tagName,isDraft,createdAt,targetCommitish -L 1
Unknown JSON field: "targetCommitish"                               # EV-10, exit 1

$ jq … 'now | floor as $now | …' .moai/reports/t1536/red-drafter-fixture.json
jq: error (…:6): Cannot iterate over number (1792671846....)        # EV-3, exit 5

$ jq … '(now|floor) as $now | …' .moai/reports/t1536/red-drafter-fixture.json
stale-draft                                                         # EV-4, exit 0

$ grep -c "gh release view" .github/workflows/release-drafter-cleanup.yml   → 0, exit 1   # EV-7
$ grep -c "concurrency" .github/workflows/release-drafter-cleanup.yml       → 0, exit 1   # EV-8
$ grep -c "attempt1" .github/workflows/ci.yml                               → 0, exit 1   # EV-9
```

All executed at tree `bf0a0f35fc280e83c2f9fd76d0e2a5de39b7e8b7` on 2026-10-22; acceptance.md §B carries full verbatim text. Run-phase pre-flight re-runs EV-1 and `git rev-parse HEAD` before the first M-commit.

## §D Constraints — PRESERVE list (normative)

1. `.github/test-input-filters.yml` root set (including the t1858 root `.github/required-checks.yml`) and its three consumers; `internal/template/test_input_filter_parity_test.go` green (EV-1).
2. `release-pr-gate` semantics (release-pr-multi-os.yml:362-403): cancelled/timed-out/skipped-without-named-exclusion never passes; detect-release hard failure fails the gate; `non-release-pr` and `docs-only-release-pr` are the only named intentional exclusions.
3. ci.yml release/* race exclusion rationale (the release matrix supersedes; its ubuntu legs are the only REQUIRED ubuntu race — do not drop them).
4. `test-skip-marker` intentional-skip reporting for the test family on non-`go_code` changes.
5. SPEC-CANDIDATE-CI-001 surfaces: `ci/**` push trigger without `paths:` filter, guard-bundle partition, six race legs' exact-cover selectors and per-leg ceilings, `retry-flaky.sh` wiring at its seven ci.yml sites (:222, :370, :422, :473, :525, :578, :635), and ci.yml's conditional cancellation (:52-55, REQ-CCI-014).
6. Least-privilege `permissions:` blocks in every touched workflow (contents: read for verification workflows; the cleanup workflow's `contents: write` stays scoped to `gh release delete`).
7. The release matrix evidence chain: `-json` stream to file, `|| rc=$?` census-on-failure pattern (:300/:302/:304), the empty-sweep refusal (:306), gzip compression (:309), per-OS-and-shard artifact names (:319), windows `if-no-files-found: error` (:322).
8. SPEC-APPJS-FIRE-GUARD-001 invariants: pinned Chrome-for-Testing + sha256, AppArmor profile step, probe manifest lint, mutation restore `RESTORED_BYTE_IDENTICAL` assertion (:1112), `0 → 1 → 0` cycle.
9. `release.yml` deploy path concurrency (no-cancel) — untouched.
10. graph-freshness landed split and `release-pr-multi-os.yml:20` — untouched here (owned by the item-3 follow-up card).
11. Required-name SSoT — `.github/required-checks.yml` is the SINGLE required-name source: it feeds the branch-protection apply path (`internal/cli/branch_protection.go`, error string at :103 "branch %q not found in required-checks.yml"), the candidate/ci-watch verdict tooling, and `scripts/release.sh` via `release_gate`. `.github/branch-protection.json.gtmpl` and its distributed-template mirror are SHAPE-ONLY (no check names; guarded by `internal/template/branch_protection_parity_test.go`) and MUST NOT be edited. Under decision-index Q4's applied default the file is NOT edited at all; under the alternate, any name change lands in the same commit as the workflow change and keeps `release_gate` ⊆ `branches.main.contexts`.
12. SPEC-CI-RELEASE-VERIFY-SHARD-001 surfaces: the 18-leg `os × shard` matrix and its `include` entries (:117-:153), selector parity with ci.yml (`release_verify_shard_parity_test.go`), per-shard `-timeout 35m` and job `timeout-minutes: 40` not raised, no retry wrapper on release legs, `go vet` once per OS on the cli-2 leg (:210-:212).

Forbidden: modifying workflow files during plan phase; `git add -A`; force-push; touching `.moai/state/**`, `.moai/logs/**`, other SPEC directories.

## §E Self-Verification (run-phase E-matrix summary)

E1 AC matrix from acceptance.md §D (command + verbatim output per AC). E2 `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...`. E3 `go test -cover ./internal/template/`. E4 boundary grep (no AskUserQuestion in non-test code touched — none expected). E5 `golangci-lint run --timeout=5m` + `moai spec lint --spec SPEC-RELEASE-PIPELINE-001`. E6 commit SHAs + local state (lanes do not push). E8 RED outputs per milestone, captured before GREEN.

## §F Milestones (M1, M2, M4, M5 — card-item mapping, ordered by decision reversibility)

> M# ↔ card item # is the traceability contract; M3 is vacant (moved). Ordering below leads with the decisions most likely to change: M2 carries the required-name surface (decision-index Q4/Q8), M5 the fetch-source choice (Q7). The executor may run M5 first — it is the only live functional defect.

### M2 — Item 2: input classes + gated-family verdict surface (Priority High)

1. RED: author `internal/template/ci_cost_gating_test.go` — `TestCICostJobsInputGated` (each of lint/build/constitution-check/test-browser carries an input-class gate sourced from the shared filter on pull_request events, runs when the classification did not succeed in a run that is not itself cancelled, carries a `!cancelled()`-guarded rather than `always()`-based condition, and runs unconditionally on the event classes Q8 exempts), `TestCIGatedFamilyVerdictSemantics` (renamed from `TestCIAggregateGateSemantics` in v0.5.0, D15 — the surface is the retained per-job names, not an aggregate; asserts acceptance.md AC-RPC-003 cells (i)-(iv) per retained name, including the counter-case **classification failure × verification success → `failure`**, plus the push cells equal to the real job's outcome), `TestCIRequiredNamesSatisfied` (per event class, for every classification outcome: `branches.main.contexts` reported on a pull_request to main, `branches."ci/**".contexts` by ci.yml on a `ci/**` push, `release_gate` by ci.yml on a push to main; no name asserted on an event its set does not govern; `release_gate` ⊆ `branches.main.contexts`; the gtmpl + mirror untouched — a regression guard, green on the real tree today, completed by negative controls (a) an in-memory ci.yml copy with `lint` gated and no `Lint` marker, and (b) an in-memory required-checks copy that adds `Release PR Multi-OS Gate` to `release_gate`, both observed red). Observe RED on the first two (EV-5 pins the first; EV-14 pins the second's counter-case) and on both negative controls of the third.
2. Add the input classes (`binary_inputs`, `browser_inputs`, `harness_inputs`) to `.github/test-input-filters.yml` beside `go_code`, each root with its evidence comment (which job's verdict reads it).
3. Gate the four cost jobs on their classes for pull_request events only, per Q8's applied default; push (main, `ci/**`) and `workflow_dispatch` keep the always-run behavior. Fail-open: because `detect` (ci.yml:61-79) carries no `continue-on-error`, a failed classification fails `detect` itself, and a job that merely `needs: detect` would be skipped with its marker — so each cost job's run condition is `!cancelled()`-guarded: unless the run itself was cancelled, it runs when `detect` did not conclude `success`, or when its class matched, or on a Q8-exempt event. No cost job or marker condition is built on `always()`, which would start the job, and keep it running, after its run is cancelled — reversing ci.yml's preserved conditional cancellation (§D PRESERVE item 5) and the cost goal of this milestone.
4. Publish the gated required names per Q4's applied default — retain `Lint` and the five `Build (…)` names. (a) Each name is also published by a skip-marker job in the `test-skip-marker` pattern that runs on pull_request events only, under a `!cancelled()`-guarded condition, when `detect` did not conclude `success` OR the class output is a positive non-match. The marker passes, with a step-summary explanation naming the intentional exclusion, only when `detect` concluded `success` (the positively observed non-match, unchanged from the existing pattern); when `detect` did not conclude `success` the marker's step fails. So on a classification failure both check runs published under the same required name — the real job (b) and the marker — conclude `failure`, and no rule for resolving two same-named check runs is relied on. On Q8-exempt events the marker does not run. (b) Each real cost job ends with a classification-verdict step that runs after its verification steps whatever they concluded (`if: always()`) and fails the job — and so the same required name — on a pull_request event when `detect` did not conclude `success`. This is the single rule REQ-RPC-003 states: classification failure × verification success publishes `failure`. required-checks.yml stays unedited. (Q4 alternate, if the operator overrides: one aggregate name, with `branches.main.contexts`, `branches."ci/**".contexts`, and `release_gate` edited in the same commit, and the aggregate publishing on push to main.)
5. Preserve REQ-RPC-004: harness roots stay inside `go_code` (release 3-OS coverage unchanged); one guard-test assertion for that containment.
GREEN: `go test ./internal/template/ -run '^(TestCICostJobsInputGated|TestCIGatedFamilyVerdictSemantics|TestCIRequiredNamesSatisfied)$' -count=1 -v` shows 3 tests run and passing; `go test ./internal/config/ -run '^TestRequiredChecks_ReleaseGateSubsetOfMain$' -count=1 -v` stays green.

### M5 — Item 5: drafter-cleanup repair, fetch included (Priority High — live defect)

1. RED: author `internal/template/release_drafter_cleanup_test.go` —
   - `TestCleanupFetchFieldsSupported` (D10, AC-RPC-013): extracts every `gh release list --json` / `gh release view --json` / `gh api` fetch in the cleanup workflow; asserts each `--json` field list is a subset of the field set that subcommand supports (list: `createdAt,isDraft,isImmutable,isLatest,isPrerelease,name,publishedAt,tagName`; view adds `targetCommitish` among others — measured gh 2.92.0, EV-10/EV-12); asserts the payload the filter reads supplies `isDraft`, `createdAt`, `tagName`, `targetCommitish` under those names (via the requested fields or an explicit mapping step); asserts the filter still carries the `release/*` exemption and a draft with no resolvable target branch is not selected. RED today: `targetCommitish` requested from `gh release list`.
   - `TestCleanupFilterSelectsStaleDrafts` (AC-RPC-010): executes the workflow's extracted jq program against a committed testdata fixture (`internal/template/testdata/release-drafter-cleanup.json`, four drafts: stale / fresh / release-branch-targeted / published, keyed exactly as the repaired fetch delivers them) and asserts exactly `stale-draft` is selected.
   - `TestCleanupRequeriesBeforeDelete` (AC-RPC-011), `TestDrafterCleanupConcurrency` (AC-RPC-012), `TestCleanupDryRunDispatchReachable` (dry-run reachability + literal-"1" value contract, mock-gh, zero `gh release delete` under a dispatched `dry_run: true`) — unchanged from v0.3.0.
   Observe RED (EV-10, EV-3, EV-7, EV-8 pin the first four; the fifth fails on the absent input).
2. Replace the fetch source per decision-index Q7's applied default: list with supported fields (`gh release list --json tagName,isDraft,createdAt -L 100`), then resolve `targetCommitish` per candidate draft with `gh release view <tag> --json tagName,isDraft,createdAt,targetCommitish` and assemble the filter input from those payloads (camelCase end to end — no mapping layer). The per-tag view call is the same call REQ-RPC-011's re-query needs; the run phase may share it, provided the re-query still happens immediately before each delete. (Q7 alternate: `gh api repos/{owner}/{repo}/releases` with an explicit `tag_name→tagName`, `draft→isDraft`, `created_at→createdAt`, `target_commitish→targetCommitish` mapping before the filter.)
3. Fix release-drafter-cleanup.yml:65 to `(now|floor) as $now`.
4. Add the per-tag re-query guard to the delete step (skip + log on state change).
5. Join both draft-mutating workflows to one ref-independent LITERAL concurrency group (e.g. `release-drafter-mutations`) with `cancel-in-progress: false` on BOTH sides: add the cleanup's block, replace release-drafter.yml's ref-keyed group value (:19-21), and flip :21 to `false`. Record the queuing trade-off in both comments (B6).
6. Make the dry-run path reachable AND value-correct: add `workflow_dispatch.inputs.dry_run` (boolean) under :24 and wire it into the env step (:41-:48) with an explicit boolean→"1" conversion (`${{ inputs.dry_run && '1' || '0' }}` or an equivalent in-shell coercion) — the delete loop honors only the literal "1" (:92), while GitHub boolean inputs stringify to "true"/"false".
GREEN: `go test ./internal/template/ -run '^(TestCleanupFetchFieldsSupported|TestCleanupFilterSelectsStaleDrafts|TestCleanupRequeriesBeforeDelete|TestDrafterCleanupConcurrency|TestCleanupDryRunDispatchReachable)$' -count=1 -v` shows 5 tests run and passing; EV-3/EV-4 divergence preserved inside `TestCleanupFilterSelectsStaleDrafts`.

### M4 — Item 4: stage budgets, browser phase split, retry evidence (Priority Medium)

1. RED: author `internal/template/verification_budget_test.go` — `TestStageCeilingBelowJobBudget` with two halves. Browser half: both browser phases (:1013/:1020) carry a step-level `timeout-minutes` strictly below the job's 20, with declared headroom for the following `always()` compress/upload steps — RED today (EV-6). Release half, built on the measured budget model rather than on the per-binary timeout: a leg's 40-minute job holds pre-step time (checkout, toolchain, `go vet` on cli-2) + the race step (compilation plus every package binary — `-timeout 35m` bounds one binary, not the step) + the evidence steps; measured (EV-15: failing run 39729541531, n=1, windows, non-main branch) pre-step ≤ 4m21s, race step ≤ 37m28s on failing legs, evidence tail 6-8s. The half asserts (i) compress (:309) and upload (:315) stay separate `if: always()` steps after the race step, so evidence collection sits outside any budgeted test step; (ii) `-timeout 35m` (:300/:302) and job `timeout-minutes: 40` (:121) unchanged; (iii) if the race step carries `timeout-minutes: T`, then T + 5 + 1 ≤ 40, with 5 (pre-step allowance, 4m21s rounded up) and 1 (evidence allowance) held as test constants citing EV-15. The test asserts no lower bound on T and makes no "never pre-empts `-timeout 35m`" claim: EV-15's race-step durations come from failing legs and bound nothing about green legs, so no lower bound for T is established, while the upper bound is ≤ 34 (pre-step and evidence headroom); which T, if any, to place is therefore decision-index Q9. The release half holds today; completion is by negative controls over an in-memory release-pr-multi-os.yml copy (T = 38; upload without `if: always()`), observed red. `TestBrowserPhaseSplit` (green and mutation phases carry independent stage bounds; a max-duration green phase leaves the mutation phase its declared budget), `TestRetryPolicyPreservesFirstFailure` (retry only through the registry wrapper — green today; script writes attempt1 before the re-run, :166 before :169 — green today; every ci.yml job invoking the wrapper uploads `.moai/logs/ci-flaky-retry.attempt1.json` — RED today, EV-9). Observe RED (EV-6 pins the first test's browser half and the second, EV-9 the third's upload assertion) and both release-half negative controls.
2. Release race step per decision-index Q9's applied default: add no step ceiling; record the EV-15 budget arithmetic (pre-step + race step + evidence ≤ 40, measured values and run id, labelled as a failing-run, n=1, windows, non-main-branch measurement) in the workflow comment above the race step so the inherited headroom risk is visible in place; touch nothing in PRESERVE item 12. (Q9 alternate, if the operator overrides: `timeout-minutes: 34` on the race step, with the same comment, recording that the failing-run cli-1/cli-3 race steps exceeded 34 minutes and that the green-leg effect is unmeasured.)
3. Split the browser job's phases into separately bounded stages per decision-index Q5's applied default (one job; internal stage ceilings + upload headroom; phase order `0 → 1 → 0` and restore assertions unchanged).
4. Extend each retry-wrapped ci.yml job's `Upload test event stream` path (test job + six race legs) to include `.moai/logs/ci-flaky-retry.attempt1.json` (uploads stay `if: always()` with `if-no-files-found: warn`). Do NOT add retries to the release legs (Q6, and SPEC-CI-RELEASE-VERIFY-SHARD-001 Q5).
GREEN: `go test ./internal/template/ -run '^(TestStageCeilingBelowJobBudget|TestBrowserPhaseSplit|TestRetryPolicyPreservesFirstFailure)$' -count=1 -v` shows 3 tests run and passing; `go test ./internal/template/ -run '^TestReleaseVerifyShardParity$' -count=1 -v` (the t1892 selector parity guard, `release_verify_shard_parity_test.go:938`) stays green with a swept count of 1.

### M1 — Item 1: verify and preserve the landed shared classifier (Priority Medium, verification-only)

1. Re-run EV-1 at the run-phase HEAD (swept count 2); record in progress.md §E.2.
2. Grep-verify the three consumers still pass `filters: .github/test-input-filters.yml` (ci.yml:79, codeql.yml:43, release-pr-multi-os.yml:68).
3. Confirm release fail-open intact: `continue-on-error: true` (:64) + `!= 'false'` gate (:120).
4. Repair only comment drift found; no behavior change.

## §G Anti-Patterns

- AP-1: adding an inline classifier copy anywhere.
- AP-2: gating a cost job with fail-closed classification (an unevaluated filter must RUN the job).
- AP-3: satisfying a required name with a success stub over an UNEVALUATED run (stub legitimate only for positively observed non-match).
- AP-4: changing any `cancel-in-progress` value other than the two draft-mutating workflows (item 3 belongs to the follow-up card).
- AP-5: raising a job `timeout-minutes` or the release per-shard `-timeout 35m` instead of bounding stages.
- AP-6: retrying a semantic failure — registry-admitted transient failures only; first failure uploaded.
- AP-7: editing required-checks.yml in a different commit than the workflow change it accompanies, or breaking `release_gate` ⊆ `branches.main.contexts`.
- AP-8: "simplifying" `|| rc=$?` to `; rc=$?`.
- AP-9: making the cleanup fetch succeed by dropping `targetCommitish` without re-sourcing it — that silently deletes `release/*`-targeted drafts.
- AP-10: switching the fetch to the REST payload without the snake_case→camelCase mapping (filter predicates read null; nothing is selected).

## §H Cross-References

- spec.md §C REQ-RPC-001..013 (005/006 retired); acceptance.md §D AC-RPC-001..013 (005/006 retired) + §B evidence ledger.
- decision-index.md Q1-Q9 (interview.yaml `decision_gate: on`).
- `.moai/reports/t1536/progress.md` — card progress record (Hold Record, Split 결정).
- Sibling SPECs: §A.3 above.
