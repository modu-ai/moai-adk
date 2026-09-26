# progress.md — SPEC-SPEC-LINT-VACUOUS-ASSERT-001

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-09-26
card: t1269
tier: M
artifacts: [spec.md, plan.md, acceptance.md, spec-compact.md, progress.md]
baseline_head: e464fd5d0
req_count: 14
ac_count: 14
run_evidence_target: .moai/reports/t1269/verdict.md
severity_rollout: "warning; non-advisory only for SPECs created on/after a pinned cutoff = day after newest existing created (plan.md §B.1)"
estimate_not_measurement: "plan.md §B.1 corpus figures are a throwaway-script estimate; AC-VTA-010 records the measured count"
audit_history: "iter-1 FAIL 0.82 (.moai/reports/t1269/plan-audit-iter1.md); v0.2.0 repairs D1-D10 and D11-D16"
```

## §E.2 Run-phase Evidence

Tree: run commit `171634ee7` (parent `d1d19bc17`, merge base with develop `e464fd5d0`).
Full verbatim outputs: `.moai/reports/t1269/verdict.md` (run section).

| AC | Status | Command | Actual Output |
|---|---|---|---|
| AC-VTA-001 | PASS | `go test ./internal/spec/ -run '^TestVacuousAssertionRule_RunPatternAxis$' -v` | `--- PASS: TestVacuousAssertionRule_RunPatternAxis (0.03s)` |
| AC-VTA-002 | PASS | `go test ./internal/spec/ -run '^TestVacuousAssertionRule_OutcomeAssertionAxis$' -v` | `--- PASS: TestVacuousAssertionRule_OutcomeAssertionAxis (0.04s)` (N1 and N3 arms included) |
| AC-VTA-003 | PASS | `go test ./internal/spec/ -run '^TestVacuousAssertionRule_MarkdownContexts$' -v` | `--- PASS: TestVacuousAssertionRule_MarkdownContexts (0.02s)` |
| AC-VTA-004 | PASS | `go test ./internal/spec/ -run '^TestVacuousAssertionRule_D15Regression$' -v` | `--- PASS: TestVacuousAssertionRule_D15Regression (0.00s)` |
| AC-VTA-005 | PASS | `go test ./internal/spec/ -run '^TestVacuousAssertionRule_ArtifactScope$' -v` | `--- PASS: TestVacuousAssertionRule_ArtifactScope (0.00s)` |
| AC-VTA-006 | PASS | `go test ./internal/spec/ -run '^TestVacuousAssertionRule_GateCutoff$' -v` | `--- PASS: TestVacuousAssertionRule_GateCutoff (0.01s)` (N4 malformed arms included) |
| AC-VTA-007 | PASS | `go test ./internal/spec/ -run '^TestVacuousAssertionRule_LintSkip$' -v` | `--- PASS: TestVacuousAssertionRule_LintSkip (0.87s)` |
| AC-VTA-008 | PASS | `go test ./internal/spec/ -run '^TestVacuousAssertionRule_Registered$' -v` | `--- PASS: TestVacuousAssertionRule_Registered (0.26s)` |
| AC-VTA-009 | PASS | CI argument vector from each fixture root | red: `baseline: EXCEEDED — .moai/spec-lint-baseline.json` then `  VacuousTestAssertion: recorded 0 -> current 1 (+1)`, `rc=1`; green: `baseline: OK`, `rc=0` |
| AC-VTA-010 | PASS | `go run ./cmd/moai spec lint --baseline .moai/spec-lint-baseline.json` + newest-created + census | gate `rc=0`; newest created `2026-09-26` < cutoff `2026-09-27`; census `{"total": 3335, "gated": 0}` |
| AC-VTA-011 | PASS | per-SPEC `--json` count for `VacuousTestAssertion` | `0` (positive control SPEC-INSTRUCTION-FILES-UNIFY-001 → `5`) |
| AC-VTA-012 | PASS | protected-path diff from merge base `e464fd5d0` | empty |
| AC-VTA-013 | PASS | mutants M1-M5 (`.moai/reports/t1269/mutants.py`) | M1/M2/M4 fail RunPatternAxis and D15Regression; M3 fails MarkdownContexts only; M5 fails RunPatternAxis `detect/sibling_groups_top-level_pipe` only; restored tree green and byte-identical |
| AC-VTA-014 | PASS | coverage + `go vet` + golangci-lint v2.1.6 | rule file 136/146 statements = 93.2%; vet rc=0; `0 issues.` |

| Invariant | Status | Evidence |
|---|---|---|
| No new gate / workflow / dependency / baseline edit | PASS | AC-VTA-012 empty diff |
| No other card's SPEC edited | PASS | commit `171634ee7` touches only this SPEC's spec.md and acceptance.md under `.moai/specs/` |
| RED before GREEN | PASS | nil-returning stub run: all 8 tests `--- FAIL` (`.moai/reports/t1269/red-stub.txt`) |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-26
run_commit_sha: 171634ee7
run_status: audit-ready
ac_pass_count: 14
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not-run (lane does not push; lead batch-pushes develop)
l44_post_push_fetch: not-applicable (no push)
new_warnings_or_lints_introduced: "0 golangci issues; 3335 advisory VacuousTestAssertion corpus warnings, 0 gated"
cross_platform_build.darwin: "go test ./internal/spec/... rc=0"
cross_platform_build.windows: not-run (no build-tagged code added)
total_run_phase_files: 10
m1_to_mN_commit_strategy: "single run commit + evidence commit"
gate_cutoff: 2026-09-27
debts_discharged: [N1, N2, N3, N4]
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
