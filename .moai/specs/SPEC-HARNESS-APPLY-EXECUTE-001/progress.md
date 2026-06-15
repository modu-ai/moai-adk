# Progress — SPEC-HARNESS-APPLY-EXECUTE-001

## §E — Phase 0.95 Mode Selection

- **Input parameters**: tier=M, scope=4 files (execute.go + 2 test files + harness.go edit), domain count=2 (internal/cli/harness + internal/harness), file language mix=100% Go, concurrency benefit=LOW (coding-heavy), Agent Teams prereqs=not met.
- **Decision**: sub-agent (Mode 5)
- **Justification**: coding-heavy single-SPEC implementation with sequential milestone dependencies (M1→M2→M3→M4→M5). Per Anthropic's coding-task parallelism caveat, sequential sub-agent is the correct default. No multi-domain research fan-out benefit; scope well below the ~30-file mechanical-transform Mode 6 threshold.

## §E.1 Milestone Log

| Milestone | Description | Status | Commit |
|-----------|-------------|--------|--------|
| M1 | execute.go verb-factory + RunExecute skeleton + proposal loader + error→exit mapping | in-progress | (pending) |
| M2 | Apply pipeline wiring + autoApply contract + canonical paths | pending | — |
| M3 | Error surface → exit code mapping (merged into M1 test file) | pending | — |
| M4 | apply --execute UX delegation + payload-only preservation | pending | — |
| M5 | T1 integration telemetry test + MX tags + honest framing | pending | — |

## §E.2 Run-phase Evidence (AC Matrix)

The acceptance.md §D AC Matrix (16 ACs: MUST 14 / SHOULD 2) is the verification SSOT.
This table is populated as each AC is verified during run-phase.

| AC ID | REQ | Status | Verification Command | Actual Output |
|-------|-----|--------|---------------------|---------------|
| AC-AEX-001 | REQ-AEX-001 | (pending) | `test -f internal/cli/harness/execute.go && go test -run 'TestPropose_NoAskUserQuestion$'` | — |
| AC-AEX-002 | REQ-AEX-002 | (pending) | router registration grep | — |
| AC-AEX-003 | REQ-AEX-003 | (pending) | `go test -run 'TestApply_PayloadOnly...$\|TestApply_DelegatesToGoPath...$'` | — |
| AC-AEX-004 | REQ-AEX-004 | (pending) | `go test -run 'TestExecute_LoadsProposalByID$'` | — |
| AC-AEX-005 | REQ-AEX-005 | (pending) | `grep -nE 'AutoApply:[[:space:]]*true' execute.go` | — |
| AC-AEX-006 | REQ-AEX-006 | (pending) | harness.yaml byte-identical + no config write grep | — |
| AC-AEX-007 | REQ-AEX-007 | (pending) | `go test -run 'TestExecute_ProductionPipeline_UsesAutoApplyTrue$'` | — |
| AC-AEX-008 | REQ-AEX-008 | (pending) | `grep -nE 'NewApplierWithRegressionGate\|WithOutcomeObserver' execute.go` | — |
| AC-AEX-009 | REQ-AEX-009 | (pending) | `go test -run 'TestExecute_ResolvesCanonicalHarnessPaths$'` | — |
| AC-AEX-010 | REQ-AEX-010 | (pending) | `go test -run 'TestExecute_FirstApply_WritesApplyOutcomeTelemetry$'` | — |
| AC-AEX-011 | REQ-AEX-011 | (pending) | `go test -run 'TestExecute_NilSessions_CanaryDoesNotReject$'` | — |
| AC-AEX-012 | REQ-AEX-012 | (pending) | `go test -run 'TestExecute_MissingProposal_ExitsUserError$\|TestExecute_RejectionError...$'` | — |
| AC-AEX-013 | REQ-AEX-013 | (pending) | `go test -run 'TestExecute_RegressionError_ExitsUserError$'` | — |
| AC-AEX-014 | REQ-AEX-014 | (pending) | `go test -run 'TestExecute_PendingErrorUnderAutoApply_ClassifiedAsInvariantExit2$'` | — |
| AC-AEX-015 | REQ-AEX-015 | (pending) | `go test -run 'TestExecute_MeasurementExecError_ExitsSystemError$'` | — |
| AC-AEX-016 | REQ-AEX-016 | (pending) | `go test -run 'TestPropose_NoAskUserQuestion$'` | — |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: (pending)
run_commit_sha: (pending — M-final)
run_status: in-progress
ac_pass_count: 0
ac_fail_count: 0
preserve_list_post_run_count: (pending)
l44_pre_commit_fetch: (n/a — worktree, orchestrator controls push)
l44_post_push_fetch: (n/a — orchestrator controls push)
new_warnings_or_lints_introduced: (pending)
cross_platform_build:
  linux_amd64: (pending)
  windows_amd64: (pending)
total_run_phase_files: (pending)
m1_to_mN_commit_strategy: per-milestone commits, local only (orchestrator controls push)
```
