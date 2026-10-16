---
id: SPEC-RELEASE-PIPELINE-001
title: "Release pipeline coverage and cost — acceptance criteria"
version: "0.1.0"
created: 2026-10-17
author: manager-spec
---

# acceptance.md — SPEC-RELEASE-PIPELINE-001

## §A Verification model

Every AC is binary-testable through one of: a Go guard test in `internal/template/` (reads the workflow YAML / executes the extracted jq program against the committed fixture), a single-invocation read-only command, or a workflow-level observation. ACs are classed **release-blocking** (two-cell adoption: RED-now cell + green-path cell, per `.claude/rules/moai/development/verification-completeness.md` §2) or **regression-guard** (preserve an already-green invariant).

## §B Evidence ledger (RED-now cells; tree `0bd1da4973d1d9539172f90587ca1ee84c8d9426`, captured 2026-10-16T19:1xZ)

**EV-1 — parity guard baseline (regression-guard pin).**
Command: `go test ./internal/template/ -run '^(TestTestInputFiltersSingleSource|TestTestInputFiltersCoversGoTestInputs)$' -count=1`
Verbatim stdout: `ok  github.com/modu-ai/moai-adk/internal/template	0.355s` · Exit code: 0 · Tree: `0bd1da4973d1d9539172f90587ca1ee84c8d9426`

**EV-2 — release workflow cancellation flag (RED for AC-RPC-005).**
Command: `grep -n "cancel-in-progress" .github/workflows/release-pr-multi-os.yml`
Verbatim stdout: `20:  cancel-in-progress: false` · Exit code: 0 · Tree: `0bd1da4973d1d9539172f90587ca1ee84c8d9426`
Why red: the value is `false` on a read-only verification workflow — the defect AC-RPC-005 exists to remove.

**EV-3 — pre-fix jq program from release-drafter-cleanup.yml:65 (RED for AC-RPC-010).**
Command: `jq -r --arg threshold_days 30 'now | floor as $now | (($threshold_days | tonumber) * 86400) as $threshold_secs | .[] | select(.isDraft == true and (($now - (.createdAt | fromdate)) > $threshold_secs) and (.targetCommitish | startswith("release/") | not)) | .tagName' .moai/reports/t1536/red-drafter-fixture.json`
Verbatim stderr: `jq: error (at .moai/reports/t1536/red-drafter-fixture.json:6): Cannot iterate over number (1792178114....)` · Exit code: 5 · Tree: `0bd1da4973d1d9539172f90587ca1ee84c8d9426`
Why red: the `as` binding rebinds the iteration input from the releases array to `now`'s numeric output — the selection never executes, exactly the card's "array→number conversion never executes (reproduced)".

**EV-4 — fixed-form divergence on the same fixture (green-path pin for AC-RPC-010).**
Command: `jq -r --arg threshold_days 30 '(now|floor) as $now | (($threshold_days | tonumber) * 86400) as $threshold_secs | .[] | select(.isDraft == true and (($now - (.createdAt | fromdate)) > $threshold_secs) and (.targetCommitish | startswith("release/") | not)) | .tagName' .moai/reports/t1536/red-drafter-fixture.json`
Verbatim stdout: `stale-draft` · Exit code: 0 · Tree: `0bd1da4973d1d9539172f90587ca1ee84c8d9426`
The fixture holds 4 drafts (stale/fresh/release-branch-targeted/published); only `stale-draft` is selected — threshold, draft-flag, and release-branch exemption all verified in one divergence against EV-3.

**EV-5 — cost jobs carry no input gate (RED for AC-RPC-002/003/004).**
Command: `grep -nE "^  [a-z][a-z0-9-]*:|^    if:|^    needs:|^    continue-on-error:" .github/workflows/ci.yml`
Verbatim stdout (excerpt, job → gate map): `772:  lint:` (no following if/needs), `819:  build:` (none), `889:  constitution-check:` (only `893: continue-on-error: true`), `950:  test-browser:` (none); contrast `102: test:` → `105: needs: detect` + `106: if: needs.detect.outputs.go_code == 'true'` · Exit code: 0 · Tree: `0bd1da4973d1d9539172f90587ca1ee84c8d9426`
Why red: the four cost jobs are ungated — every PR class pays them; no aggregate exists.

**EV-6 — stage-vs-job budget map (RED pin for AC-RPC-007/008).**
Command: `grep -nE "timeout-minutes:|-timeout (17m|35m|25m)" .github/workflows/ci.yml .github/workflows/release-pr-multi-os.yml`
Verbatim stdout (excerpt): `release-pr-multi-os.yml:94: timeout-minutes: 40`, `:217: go test -json -race -timeout 35m ./...`, `ci.yml:222: retry-flaky.sh ... -timeout 25m ...` (test job, `timeout-minutes: 30` at :107); test-browser job `timeout-minutes: 20` with green step `-timeout 17m` (read directly at :950-:1017) · Exit code: 0 · Tree: `0bd1da4973d1d9539172f90587ca1ee84c8d9426`
Why red: browser green phase (17m) + mutation cycle share one 20m job budget; the release race step (35m) leaves the `always()` census/compress/upload steps to compete for the 5m remainder.

## §C Edge cases

- E1: mixed diff (docs + browser inputs) — browser job runs, build matrix does not (class matched per job, not per PR).
- E2: filter file itself edited — every class including `go_code` re-matches (the filter file is a root of its own classification; harness verification re-runs).
- E3: cleanup listing lands mid-drafter-update — shared group queues; no mid-deletion cancellation (REQ-RPC-012).
- E4: draft published between listing and deletion — re-query skips it and logs the skip (REQ-RPC-011).
- E5: rapid re-push on a release PR — prior-HEAD run cancelled; newest-SHA run is the verdict; the cancelled run's gate failure belongs to the superseded SHA only.
- E6: green phase at max ceiling — mutation phase retains its declared budget (headroom arithmetic holds).

## §D AC matrix (Given-When-Then)

| AC | REQ | Class | Given / When / Then |
|---|---|---|---|
| AC-RPC-001 | REQ-RPC-001 | regression-guard | Given HEAD carries the shared filter and its three consumers, When the parity guard tests run (`go test ./internal/template/ -run '^(TestTestInputFiltersSingleSource|TestTestInputFiltersCoversGoTestInputs)$' -count=1`), Then both pass (EV-1) and any consumer drift, inline copy, or lost required root fails the suite. |
| AC-RPC-002 | REQ-RPC-002 | release-blocking | Given ci.yml at plan-phase HEAD (EV-5), When `TestCICostJobsInputGated` runs, Then it FAILS (no gate on lint/build/constitution-check/test-browser); green path: M2 gates each cost job on its shared-filter input class, fail-open on unevaluated classification, and the test passes. |
| AC-RPC-003 | REQ-RPC-003 | release-blocking | Given no aggregate exists (EV-5), When `TestCIAggregateGateSemantics` runs, Then it FAILS; green path: exactly one aggregate for the gated family — fails on member `failure`/`cancelled`, fails on classification hard-failure, passes with a NAMED exclusion only under positively observed non-match. |
| AC-RPC-004 | REQ-RPC-004 | release-blocking | Given the branch-protection SSoT, When `TestCIRequiredNamesSatisfied` runs (asserting contexts remain satisfied for every classification outcome and that harness roots stay inside `go_code`), Then it FAILS pre-M2 (no aggregate surface) and passes post-M2 with the SSoT updated in the same milestone; release 3-OS coverage unchanged (grep: harness roots present under `go_code`). |
| AC-RPC-005 | REQ-RPC-005 | release-blocking | Given release-pr-multi-os.yml:20 reads `cancel-in-progress: false` (EV-2), When `TestReadOnlyVerificationCancels` runs, Then it FAILS; green path: flag true, `release.yml` (deploy path) still exempt, test passes. |
| AC-RPC-006 | REQ-RPC-006 | regression-guard | Given the landed t1298 split, When `TestGraphFreshnessSlotSplit` runs, Then the main group keys on event kind + PR number with `cancel-in-progress: true` and the conflict-guard job holds its own group — any regression to a shared undifferentiated group fails the test. |
| AC-RPC-007 | REQ-RPC-007 | release-blocking | Given the budget map of EV-6, When `TestStageCeilingBelowJobBudget` runs, Then it FAILS (browser 17m-in-20m single budget; release 35m-in-40m with evidence steps in the remainder); green path: every long stage's ceiling is strictly below its job budget by a declared headroom covering census+compress+upload, test passes. |
| AC-RPC-008 | REQ-RPC-008 | release-blocking | Given the browser job runs green+mutation under one budget (EV-6), When `TestBrowserPhaseSplit` runs, Then it FAILS; green path: green and mutation phases carry independent stage bounds (E6 holds), SPEC-APPJS-FIRE-GUARD-001 phase order and restore assertions unchanged. |
| AC-RPC-009 | REQ-RPC-009 | regression-guard | Given registry-gated retries wired at ci.yml:222/:578, When `TestRetryPolicyPreservesFirstFailure` runs, Then retry invocation appears only through the flaky-registry wrapper and the first-failure stream is uploaded before any retry — a bare retry or a retry-without-preserved-first-failure fails the test. |
| AC-RPC-010 | REQ-RPC-010 | release-blocking | Given the workflow's committed jq program (EV-3: exit 5, `Cannot iterate over number`), When `TestCleanupFilterSelectsStaleDrafts` executes the extracted program against its testdata fixture (run phase adds `internal/template/testdata/release-drafter-cleanup.json`, same four-draft shape as the plan-phase evidence fixture), Then pre-fix it FAILS and post-fix it selects exactly `stale-draft` (EV-4 divergence embedded in the test). |
| AC-RPC-011 | REQ-RPC-011 | release-blocking | Given the delete loop deletes straight from the listing, When `TestCleanupRequeriesBeforeDelete` runs, Then it FAILS; green path: each deletion is preceded by a state re-query; a changed draft (E4) is skipped and the skip recorded. |
| AC-RPC-012 | REQ-RPC-012 | release-blocking | Given the cleanup workflow has no concurrency block, When `TestDrafterCleanupConcurrency` runs, Then it FAILS; green path: both draft-mutating workflows share one group, no-cancel on the destructive side (E3), test passes. |

## §D.1 Severity and traceability

All twelve ACs are release-blocking for their milestones except AC-RPC-001/006/009 (regression-guards pinning landed invariants). Every REQ-RPC-001..012 maps to exactly one AC; no orphan REQ, no orphan AC. Traceability chain: card item # → M# (plan.md §F) → REQ-RPC-### (spec.md §C) → AC-RPC-### (here) → EV-# / guard test.

## §D.2 Indirect verification (workflow-level, post-merge)

- After M2 merges: open a docs-only diff PR; observe zero cost-job runs and every required name satisfied (aggregate or marker).
- After M5 merges: `workflow_dispatch` the cleanup with `MOAI_CLEANUP_DRY_RUN=1`; observe a non-zero-or-correctly-zero stale list computed (pre-fix this step errored) and no deletions executed.

## §D.3 Closure gates (Definition of Done)

1. All guard tests of M2-M5 green with RED outputs captured (E8) and committed.
2. EV-1 re-run green at the final run-phase HEAD.
3. `moai spec lint --spec SPEC-RELEASE-PIPELINE-001` clean; golangci-lint clean; both builds green.
4. PRESERVE list (plan.md §D) re-verified item by item at the final HEAD.
5. Branch protection SSoT consistent with the workflow surface (required names satisfied on every classification outcome).

## §D.4 Forward-looking checks

- The input-class root set cites, per root, the job whose verdict reads it (same evidence-comment discipline as `go_code`) — a future input with no cited verdict owner must not be added.
- The concurrency guard test enumerates workflows by mutation class, so a NEW verification workflow added without a cancel decision fails `TestReadOnlyVerificationCancels` until classified (continued-firing, `verification-completeness.md` §1.3).
