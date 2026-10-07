# SPEC-AUDIT-CEILING-REPAIR-001 — progress

status: in-progress
card: t1560 (lane-2, run tmhxo0) · tree .moai/worktrees/t1560 · branch WT-internal-runtime-audit @ 903ccd028
evidence: this file (§G carries any ceiling records for THIS spec if the engine ever evaluates it) + .moai/reports/t1560/

## §E.1 Plan-phase Audit-Ready Signal

plan_status: pending-plan-audit
plan_complete_at: 2026-10-07
비고: plan artifacts (spec.md · plan.md · acceptance.md · decision-index.md · references/t1500-seal-excerpt.md) authored 2026-10-07 by manager-spec (card t1560, Tier M). Revision v0.3.0 (2026-10-07): plan-audit-1 repairs integrated (verdict FAIL 0.81, `.moai/reports/t1560/plan-audit-1.md`) + leader ruling #2 (D4 fold-in). `plan_status: audit-ready` is set after the plan-auditor verdict; iteration 2 re-audits delta-scoped to the fix_scope anchors (mapping in plan.md §H).

## §E.2 Run-phase Evidence

### M1 — the eight reproduction REDs (tree `8233644f0`, code-identical to the RED baseline `903ccd028`)

Authored at M1 against the still-pristine engine per the family convention
("RED is a new test — E8 evidence required"); MP-8 basis restated: no repro
test existed at plan phase by design, so the plan-phase `go test -run
'^Test…$'` selectors matched zero tests (`ok … [no tests to run]`, exit 0) —
the expected plan-phase state, never a RED observation and never a pass.
Every RED below ran on `8233644f0` with only test files added (engine files
untouched); `go build`/`go vet` clean and the AC-ACR-012 sealed-surface
selection green before the tests were authored (baseline at 10:26 KST).

| AC | Test | Command (all with `./internal/runtime/`, `-count=1`) | Verbatim RED output (assertion line) | Exit |
|---|---|---|---|---|
| AC-ACR-001 (D1) | `TestPreviousAuditedSHALatestLegacyRound` | `go test -run '^TestPreviousAuditedSHALatestLegacyRound$' -v` | `audit_counter_review_test.go:240: previous audited SHA "", want sha-rev2 (largest round strictly below the legacy latest)` | 1 |
| AC-ACR-002 (D1) | `TestPreviousAuditedSHAMixedFamilyPrior` | `go test -run '^TestPreviousAuditedSHAMixedFamilyPrior$' -v` | `audit_counter_review_test.go:264: previous audited SHA "", want sha-iter2 (convention prior below the legacy latest)` | 1 |
| AC-ACR-004 (D1 e2e) | `TestEvaluateCeilingLegacyLatestDeltaGranted` | `go test -run '^TestEvaluateCeilingLegacyLatestDeltaGranted$' -v` | `audit_ceiling_test.go:1216: outcome &{Outcome:hold Reasons:[plan-audit ceiling reached (round count 2 >= tier ceiling 2); the verdict matches no admitting arm and holds, entry blocked (REQ-ACE-006) — release path: …] Blocked:true Debts:[]} (override false), want nil — the delta round is granted: count 2 reaches the tier ceiling 2 and the eligibility conditions hold` | 1 |
| AC-ACR-005 (D2) | `TestPersistOutcomeDebtAdmitCarriesDebtInventory` | `go test -run '^TestPersistOutcomeDebtAdmitCarriesDebtInventory$' -v` | `audit_ceiling_test.go:1253: §G record carries no debts= token: - 2026-10-07T10:31:23Z SPEC-ACE-ENG-001 ceiling-outcome outcome=debt-admit reasons="plan-audit ceiling reached (round count 1 >= tier ceiling 1); the verdict fails admission on the label alone and debt-admits — findings recorded as PASS-WITH-DEBT debts, entry admitted without a question (REQ-ACE-004)" evidence=…` | 1 |
| AC-ACR-013 (D3) | `TestAppendProgressRecordConcurrentSurvival` | `go test -run '^TestAppendProgressRecordConcurrentSurvival$' -count=3 -v` | run 1/3: `audit_ceiling_test.go:1303: pre-existing progress content lost:` followed by a progress.md holding only records 18/13/08/20 under the §G heading; runs 2/3 failed the same way (run 2's content additionally interleaved a partial `fusal Record` fragment — torn write) | 1 |
| AC-ACR-014 (D4 a-c) | `TestCountAuditRoundsOverflowOwnRound` | `go test -run '^TestCountAuditRoundsOverflowOwnRound$' -v` | `(a) audit_counter_review_test.go:306: count 1, want 2 (the base and the overflow suffix are two rounds)` · `(b) audit_counter_review_test.go:325: count 1, want 2 (the base never merges with iter1)` · `(c) audit_counter_review_test.go:344: count 1 (sources 3), want 3 (1 normal + 2 unparseable)` | 1 |
| AC-ACR-014 (D4 e) | `TestPreviousAuditedSHABaseRoundBaseline` | `go test -run '^TestPreviousAuditedSHABaseRoundBaseline$' -v` | `audit_counter_review_test.go:410: previous audited SHA "", want sha-base (the base is round 0, the previous audited round)` | 1 |
| AC-ACR-015 (C3) | `TestReqACSetsUnchangedReadsAcceptance` | `go test -run '^TestReqACSetsUnchangedReadsAcceptance$' -v` | `audit_ceiling_test.go:1413: an acceptance-only AC id rename verified as unchanged — the delta must be refused (fail-closed)` | 1 |
| AC-ACR-016 (D3) | `TestAppendProgressRecordInsertsAtSectionEnd` | `go test -run '^TestAppendProgressRecordInsertsAtSectionEnd$' -v` | `audit_ceiling_test.go:1340: record did not land inside the §G block:` (the record appended past the following `## §E.2` heading) | 1 |

**AC-ACR-003 keep-green guard (pre-fix run)**: `go test -run
'^TestPreviousAuditedSHAUnparseableLatestStaysEmpty$' -v ./internal/runtime/`
→ `--- PASS: TestPreviousAuditedSHAUnparseableLatestStaysEmpty (0.00s)`,
exit 0. The preserve arms of `TestCountAuditRoundsOverflowOwnRound`
(`bare_base_alone_counts_1_and_stays_latest`,
`explicit_zero_keeps_own_round_parity`) also passed inside the pre-fix RED
batch — only the (a)/(b)/(c) subtests failed.

**Measurement note (arm (e) count parenthetical)**: acceptance.md's arm (e)
RED cell says "Count=2 but the scan skips the base". The mechanically
measured count for the arm-(e) fixture (base + iter1) at `8233644f0` is
**1**, not 2 — the same base/iter1 collapse arm (b) documents as "RED today:
1" (both files parse to round 1 and dedupe into `seen[1]`). The Count=2
figure could not be reproduced on this tree; the RED was re-anchored to the
stated previous-baseline reason (the test asserts the previous-audited-SHA
behavior before the count assertions). The defect class, the fix, and the
green state are unaffected.

Full verbatim outputs: `.moai/state/verify/t1560/m1-red-batch.txt`,
`m1-red-e2e.txt`, `m1-red-concurrent.txt`, `m1-red-base-baseline.txt`
(machine-local scratch, this run).

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — manager-develop owns this section.>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs owns this section. sync_commit_sha is written pending-backfill at the sync commit and backfilled in a follow-up commit.>_
