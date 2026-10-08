# progress.md — SPEC-CANDIDATE-CI-001

status: in-progress
card: t1478
phase: plan

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-09
artifacts: spec.md, plan.md, acceptance.md, design.md, research.md, progress.md
tier: L
req_count: 17
ac_count: 25
evidence_tree: db0c514d3 (WT-10-03-tier)
red_now_measured: AC-CCI-001-1 (live verb help, 2026-10-09), AC-CCI-006-1 (grep, 0 rows), AC-CCI-007-1 (ci.yml:16-25), AC-CCI-011-1 (integration_merge.go:95-99 placeholder), AC-CCI-014-1 (ci.yml:35-37 cancels every ref)
open_clarifications: none — the operator approval embedded in the card text resolves the §4.1 replacement and the green-gated integration-branch rule

## §E.2 Run-phase Evidence

### M1 — Config gate + candidate record (card t1478)

| Item | What changed |
|------|--------------|
| Files | `.moai/config/sections/workflow.yaml`, `internal/template/templates/.moai/config/sections/workflow.yaml` (candidate_ci block, both mirrors); `internal/config/types.go` (CandidateCIConfig + WorkflowConfig field); `internal/config/defaults.go` (Enabled false, GuardBundleRequired true); `internal/factory/candidate_record.go` (new — record type + store, keyed (card, pinned SHA), atomic writes); `internal/cli/integration_merge.go` (candidateCIEnabled reads the real key); tests: `internal/config/workflow_candidate_ci_test.go`, `internal/factory/candidate_record_test.go`, `internal/cli/integration_candidate_gate_test.go` |

- RED (tree a4944fb3f, this run): `go test -list '^TestCandidateRecord$' ./internal/factory/` → `ok  	github.com/modu-ai/moai-adk/internal/factory	0.195s` (zero test names; same shape for `^(TestWorkflowCandidateCIDefaultFalse|TestWorkflowCandidateCILoaderRoundTrip|TestCandidateCIEnabled)$` across config/cli). After writing the tests, pre-implementation RED: `cfg.Workflow.CandidateCI undefined (type WorkflowConfig has no field or method CandidateCI)`, `undefined: ReadCandidateRecord`, and `--- FAIL: TestCandidateCIEnabled/explicit_true_reads_true ... candidateCIEnabled: got false, want true with enabled: true` (the constant-false placeholder refuses the key).
- GREEN (this run): `go test ./internal/config/ -run '^(TestAuditLoaderCompleteness|TestWorkflowConfigFields|TestNewDefaultWorkflowConfig|TestNewDefaultWorkflowConfigNestedDefaults)$' -count=1` → `ok  ...  0.145s`; `go test ./internal/factory/ -run '^TestCandidateRecord$' -count=1` → `ok  ...  0.155s`; `go vet ./internal/cli/` exit 0; gofmt clean on all touched files.

### Mid-run plan amendment observed (absorbed)

While M1 was in flight, the SPEC artifacts gained the landing-check TARGET BINDING amendment, landed by the spawner lane as commit 11c169028 (`docs(spec): M5 pipeline hygiene and the candidate-to-merge-target binding (card t1478)`, applied by manager-spec): design.md D10 (branch identity refuses outright; tip advance voids with re-candidate guidance), spec.md REQ-CCI-010/011 additions, plan.md M4 (+ `TestLandingCheckRefusesTargetMismatch`), acceptance.md AC-CCI-011-2 (target-mismatch class), M5 pipeline hygiene (pipefail + mktemp -d). The same commit carries spec.md's draft → in-progress transition. Coordinator confirmed: applies at M4; no impact on M1/M2 in flight. M4 below implements the amended contract.

## §E.3 Run-phase Audit-Ready Signal

_pending run-phase_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase_

## Lane Kickoff Decision Record (2026-10-09, lane-11)

decision record: decided_by=claude+lane-11 (self-dispatch lane, ladder terminal) evidence_refs=.moai/reports/t1478/verdict.md (PASS-WITH-DEBT 1.0, must_pass 0, blocking 0; receipts rcpt-69038b8e4d8f0c9ad4499861, rcpt-37a7694cccd61c7df7470e64, rcpt-b38299720acf23c31b7a497a; Addendum 2 records the final-tree codex re-run) ladder_path=plan→run Kickoff gate, autonomous form (audit cross + evidence criteria met)

Judgment: PASS-WITH-DEBT is a passing form — 0 must-pass failures, 0 blocking; the codex required-backend machine disagreement is transparently adjudicated per-item in the verdict (P2-A refuted with tree evidence ci.yml:67/:234/:246, P2-B/P2-C accepted as optional/debts N2/N1). Post-verdict deltas (bc4d46f63, b10bcfd0e) are M5 verification-command mechanics the auditor confirmed non-conflicting at b10bcfd0e. Debts 1-3 and optional N5-N7 travel with the SPEC; N5 (one-word doc fix) rides the sync phase.

Run-phase entry: M1 first (config gate workflow.candidate_ci.enabled + candidate record schema), RED-first per plan.md test-name conventions.
