# Progress — SPEC-HARNESS-APPLY-EXECUTE-001

## §E — Phase 0.95 Mode Selection

- **Input parameters**: tier=M, scope=4 files (execute.go + 2 test files + harness.go edit), domain count=2 (internal/cli/harness + internal/harness), file language mix=100% Go, concurrency benefit=LOW (coding-heavy), Agent Teams prereqs=not met.
- **Decision**: sub-agent (Mode 5)
- **Justification**: coding-heavy single-SPEC implementation with sequential milestone dependencies (M1→M2→M3→M4→M5). Per Anthropic's coding-task parallelism caveat, sequential sub-agent is the correct default. No multi-domain research fan-out benefit; scope well below the ~30-file mechanical-transform Mode 6 threshold.

## §E.1 Milestone Log

| Milestone | Description | Status | Commit |
|-----------|-------------|--------|--------|
| M1 | execute.go verb-factory + RunExecute skeleton + proposal loader + error→exit mapping | completed | 81b50c9bd |
| M2 | Apply pipeline wiring + autoApply contract + canonical paths | completed | a2d26c811 |
| M3 | Error surface → exit code mapping (merged into M1 test file) | completed | 81b50c9bd |
| M4 | apply --execute UX delegation + execute verb registration | completed | a49b9c21a |
| M5 | T1 integration telemetry test + MX tags + honest framing + coverage | completed | (this commit) |

## §E.2 Run-phase Evidence (AC Matrix)

The acceptance.md §D AC Matrix (16 ACs: MUST 14 / SHOULD 2) is the verification SSOT.
This table is populated as each AC is verified during run-phase.

| AC ID | REQ | Status | Verification Command | Actual Output |
|-------|-----|--------|---------------------|---------------|
| AC-AEX-001 | REQ-AEX-001 | PASS | `test -f execute.go` + `TestPropose_NoAskUserQuestion$` | file PRESENT; PASS — guard scans execute.go |
| AC-AEX-002 | REQ-AEX-002 | PASS | `TestExecuteCmd_RegisteredInRouter$` | PASS — execute registered in newHarnessRouterCmd() |
| AC-AEX-003 | REQ-AEX-003 | PASS | `TestApply_PayloadOnly...$` + `TestApply_DelegatesToGoPath...$` | both PASS — payload-only preserved, --execute delegates |
| AC-AEX-004 | REQ-AEX-004 | PASS | `TestExecute_LoadsProposalByID$` | PASS — proposal loaded from .moai/harness/proposals/<id>.json |
| AC-AEX-005 | REQ-AEX-005 | PASS | `grep -nE 'AutoApply:[[:space:]]*true' execute.go` | line 95: `AutoApply:        true,` |
| AC-AEX-006 | REQ-AEX-006 | PASS | no file-write code in execute.go + harness.yaml diff=0 | NO_FILE_WRITE_CODE_IN_EXECUTE_GO; harness.yaml diff empty |
| AC-AEX-007 | REQ-AEX-007 | PASS | `TestExecute_ProductionPipeline_UsesAutoApplyTrue$` | PASS — production PipelineConfig.AutoApply==true (non-vacuous) |
| AC-AEX-008 | REQ-AEX-008 | PASS | `grep -nE 'NewApplierWithRegressionGate\|WithOutcomeObserver'` | lines 140-141: chained wiring present |
| AC-AEX-009 | REQ-AEX-009 | PASS | `TestExecute_ResolvesCanonicalHarnessPaths$` (5 subtests) | PASS — all 4 canonical paths + proposalDir join correct |
| AC-AEX-010 | REQ-AEX-010 | PASS | `TestExecute_FirstApply_WritesApplyOutcomeTelemetry$` | PASS — 1 apply_outcome line, verdict="kept", proposal_id matches |
| AC-AEX-011 | REQ-AEX-011 | PASS | `TestExecute_NilSessions_CanaryDoesNotReject$` | PASS — production Pipeline L1~L5 passed with nil sessions |
| AC-AEX-012 | REQ-AEX-012 | PASS | `TestExecute_MissingProposal_ExitsUserError$` + `TestExecute_RejectionError...$` | both PASS — exit 1 |
| AC-AEX-013 | REQ-AEX-013 | PASS | `TestExecute_RegressionError_ExitsUserError$` | PASS — *ApplyRegressionError → exit 1 (+ errors.Join walk) |
| AC-AEX-014 | REQ-AEX-014 | PASS | `TestExecute_PendingErrorUnderAutoApply_ClassifiedAsInvariantExit2$` | PASS — stub PendingApproval → *ApplyPendingError → exit 2 |
| AC-AEX-015 | REQ-AEX-015 | PASS | `TestExecute_MeasurementExecError_ExitsSystemError$` | PASS — measurement-exec wrapped error → exit 2 |
| AC-AEX-016 | REQ-AEX-016 | PASS | `TestPropose_NoAskUserQuestion$` + boundary grep | PASS — no AskUserQuestion( in internal/cli/harness/ |

## §E.2 Sync-phase Audit-Ready Signal

```yaml
sync_commit_sha: (pending — orchestrator will backfill after sync commit)
sync_status: implemented
```

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-06-15
run_commit_sha: 2dc365c35
run_status: implemented
ac_pass_count: 16
ac_fail_count: 0
preserve_list_post_run_count: 8  # applier/pipeline/regression_gate/outcome/observer/canary/lineage + harness.yaml — diff=0
l44_pre_commit_fetch: (n/a — worktree-isolated; orchestrator controls push + pre-push fetch)
l44_post_push_fetch: (n/a — orchestrator controls push)
new_warnings_or_lints_introduced: 0  # golangci-lint 0 issues on target packages; go vet clean
cross_platform_build:
  linux_amd64: exit 0
  windows_amd64: exit 0
total_run_phase_files: 7  # spec.md + progress.md + execute.go + execute_test.go + harness_execute_test.go + harness.go + harness_route.go
m1_to_mN_commit_strategy: per-milestone commits, local only (orchestrator controls push)
coverage:
  internal/cli/harness/execute.go: 86.0%   # ≥85% DoD met (80/93 statements)
  internal/harness: 87.5%                   # package, no regression
frozen_file_diff: 0                          # applier/pipeline/regression_gate/outcome/observer/canary/lineage/harness.yaml untouched
honest_framing: telemetry-only               # CHANGELOG/docs framing reserved for sync-phase; NO "prevents regressions" claim
c_hra_008_boundary: clean                    # TestPropose_NoAskUserQuestion PASS (scans execute.go)
real_go_test_recursion: none                 # T1 uses stubMeasurer; T2 uses stub SafetyEvaluator
```
