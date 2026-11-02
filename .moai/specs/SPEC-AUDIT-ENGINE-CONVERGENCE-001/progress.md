# Progress — SPEC-AUDIT-ENGINE-CONVERGENCE-001

status: draft

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_complete: true
plan_author: manager-spec (card t2017)
plan_revision: 0.5.2 (leader rulings R287, R288, R290, R308, R313, R320 applied; plan-audit round 1 repaired — .moai/reports/t2017/plan-audit-1.md, FAIL 0.75 at 36148aeec9; plan-audit round 2 repaired — .moai/reports/t2017/plan-audit-2.md, FAIL at cc100c2909; plan-audit round 3 repaired — .moai/reports/t2017/plan-audit-3.md, FAIL at 44b21e558e, G1)
follow_up:
  - "agentemit local target expansion (add plan-auditor and sync-auditor to the project targets) - input for card t2018"
  - "agentemit local .codex target expansion (add plan-auditor and sync-auditor to the codex_project target as well) - input for card t2018; the two local TOML copies themselves are in scope here and fixed by hand in M6 (leader ruling R313)"
measured_tree: cb30e05b3e5f64774b8dabd87e80c24332f4551d
tier: M
req_count: 16   # Tier M ceiling 16
ac_count: 6 logical / 16 sub-criteria   # Tier M ceiling 16
artifacts: [spec.md, plan.md, acceptance.md, decision-index.md, progress.md]
open_decisions: []   # Q1, Q3, Q6, Q7 settled by leader ruling R287
open_questions: []   # Q8 reconciliation confirmed by leader ruling R288
plan_audit: not run (separate stage)
```

## §E.2 Run-phase Evidence

### M1 — B1 delta chain gate, the delta record, and its persistence

Measured on tree `e4047e75471fd1ea52d1ff0a1a893d4ca750176c` (branch WT-audit-engine-repairs, working tree clean; the commit that records the GREEN code). RED observed first on `e3757ebaf1` (43 sub-rows FAIL, 8 controls PASS) and `c7d235a94b` (carrier row FAIL). GREEN commits: `dce4134d59` (record fields, `blocking_defect:` parse), `e4047e7547` (gate, wiring, persistence). Verbatim output file: `.moai/reports/t2017/green-m1.md` (machine-local, gitignored).

Command (one invocation, anchored, `-count=1`):

```
go test -count=1 -timeout 600s -v -run '^(TestAEC|TestClosingCommitRefusal|TestAuditMulti_|TestRunMultiAudit_|TestAC_CLOSURE_012|TestSelfReview_AuditToolSchemasUnchanged|TestConverge|TestMultiAudit|TestHandleAuditMulti|TestAuditDelta|TestCodexAudit_DeltaBase)' ./internal/cli/
```

Result: `ok  	github.com/modu-ai/moai-adk/internal/cli	150.443s`, exit 0; 125 top-level `--- PASS`, 0 `--- FAIL`.

| AC | Status | Actual Output (verbatim, same run) |
|---|---|---|
| AC-AEC-001c | PASS | `--- PASS: TestAEC001c_BrokenChain_ReviewsFullScope (31.73s)` with 11 sub-rows PASS (gap; proposed base differs from last head; no full-scope start; R0/R1 inconclusive; R0/R1 no glm entry; R0/R1 `gate_unmet`; R0/R1 codex inconclusive). Control: `--- PASS: TestAEC001c_UnbrokenReviewedChain_ReviewsTheDeltaScope_Control (2.12s)` |
| AC-AEC-001d | PASS | `--- PASS: TestAEC001d_ClosureConfirmation (38.39s)` with 13 sub-rows PASS (9 refusals, 4 complete-confirmation controls: omitted; subset; new-id out of range; count mismatch; no `blocking_count`; no FAIL file; Z-only; inheritance with R1 confirming only X; re-report without a closing commit; and the controls); `--- PASS: TestAEC001d_AdmittedDelta_EveryLegRequestCarriesTheConfirmation (3.01s)`; `--- PASS: TestAEC001d_ClosingCommitEdges_ReviewFullScope (5.41s)`; `--- PASS: TestAEC001d_EvidenceFilesNotTiedToTheChain_AreIgnored (2.47s)`; `--- PASS: TestAEC001d_UnreadableEvidenceFile_RefusesTheDelta (2.28s)`; `--- PASS: TestClosingCommitRefusal_NonSHAIsRefused (1.44s)` |
| AC-AEC-001e | PASS | `--- PASS: TestAEC001e_ChainIsReadFromServerWrittenStateOnly (19.85s)` with 7 leaf rows PASS (receipt absent R0 and R1; other tool; other tree root; receipt head differs; chain only in verdict files; chain as a call argument) |
| AC-AEC-001f | PASS | `--- PASS: TestAEC001f_GatedDelta_FilesOneDeltaRecord (1.96s)` (5 sub-rows); `--- PASS: TestAEC001f_GatedDelta_RecordCarriesTheDigestOfEachReliedOnFile (2.32s)`; `--- PASS: TestAEC001f_GatedDelta_RecordIsKeptBySelectSecondReview (2.54s)`; `--- PASS: TestAEC001f_GatedDelta_PersistsItsPassResult (2.42s)`; `--- PASS: TestAEC001f_GateRejectedDelta_IsAFullScopeCall (2.39s)` (2 sub-rows); `--- PASS: TestAEC001f_AdmittedDelta_NoteDoesNotDenyItsRecord (2.02s)`. Controls: `TestAEC001f_DeltaPassWithoutCardID_LeavesTheSessionFileAlone_Control`, `TestAEC001f_FullScopeCall_RecordIsShapedAsToday_Control`, `TestAEC001f_FullScopeRecord_IsKeptBySelectSecondReview_Control` PASS |

Other M1 checks on the same code (each run on the tree before commit, code identical to `e4047e7547`): `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` exit 0; `go vet` of `internal/cli`, `internal/closure`, `internal/auditreceipt`, `internal/auditverdict` on both OS exit 0; `golangci-lint run --timeout=3m` on those four packages exit 0, `0 issues.`; `gofmt -l` empty; `go test ./internal/auditverdict/ ./internal/closure/ ./internal/auditreceipt/` all `ok`; `audit_delta_chain.go` statement coverage 93.2% (every function at or above 87.5%).

Recorded deviations and open items for M1: (1) `TestRunMultiAudit_DeltaBase_WritesNoSecondReviewRecord` was narrowed to the `card_id`-less delta and `TestRunMultiAudit_DeltaBase_WithCardID_NoChain_IsAFullScopeCall` added, because REQ-AEC-005 makes a gate-refused `card_id` delta a full-scope call that files a full-scope record; (2) `testdata/audit-tools-schema-pre.json` `audit_multi` entry updated for the new `blocking_closed` input and the reworded `delta_base` text; (3) `golangci-lint run ./...` over the whole repository reports 1 issue (SA1019 in `internal/discovery/leader_readers_darwin.go`, a file this milestone does not touch); (4) the gate's full-suite test run was not executed locally, CI owns it.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
