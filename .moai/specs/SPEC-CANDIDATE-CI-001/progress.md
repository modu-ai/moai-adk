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

### M2 edge repairs — tip equality, verdict preservation, caller tree, serialization (card t1478, commit ccddd442e)

| Item | What changed |
|------|--------------|
| Files | `internal/cli/integration_candidate.go` (tip-equality refusal; identical-SHA verdict preservation inside the push section; candidateCallerTreeGuard; per-card critical section), `internal/factory/candidate_mutation_lock.go` + unix/windows split (new — WithCandidateMutation over the shared state-lock substrate, ErrCandidateMutationBusy), `internal/cli/integration_candidate_test.go`, `internal/factory/candidate_mutation_lock_test.go`, `internal/config/testdata/shipped_key_inventory.yaml` (the two candidate_ci keys triaged — enabled W/reader, guard_bundle_required R/workflow-layer) |

- RED (this run): `go vet` → `undefined: WithCandidateMutation`, `undefined: factory.ErrCandidateMutationBusy`, `undefined: candidateCallerTreeGuard`.
- GREEN: `go test ./internal/cli/ -run '^TestIntegrationCandidate$' -count=1` → `ok ... 5.723s`; `go test ./internal/factory/ -run '^TestCandidateMutationLock$|^TestCandidateRecord$' -count=1` → `ok ... 3.510s`; `go test ./internal/config/ -run '^(TestShippedConfigKeysHaveReaders|TestAuditLoaderCompleteness|TestWorkflowConfigFields)$'` → `ok ... 7.288s`; vet exit 0. The inventory repair flipped the M1-introduced `TestShippedConfigKeysHaveReaders` red (`workflow.candidate_ci.enabled`/`guard_bundle_required` not triaged) green.

### M2 path-hygiene + git-env repairs — gitenv scrub, run selection, store confinement (card t1478, commits d027dfe4e / f6523784d)

| Item | What changed |
|------|--------------|
| Files | `internal/factorylane/merge.go` (ExecGitRunner.Git carries `cmd.Env = gitenv.Env()` — the shared substrate every candidate-path git child flows through), `internal/cli/integration_candidate.go` (candidateScrubbedGit; --run flag + MOAI_KANBAN_ID selection; candidateCallerTreeGuard runID), `internal/factory/candidate_record.go` (candidateStoreRealPath symlink confinement; openCandidateRecordFile non-blocking regular-verified open), `internal/factory/candidate_record_unix.go`/`_windows.go` (new), `internal/factory/candidate_record_hygiene_unix_test.go` (new), tests |

- GREEN: `go test ./internal/cli/ -run '^(TestIntegrationCandidateIgnoresRepoScopingEnv|TestIntegrationCandidateRunSelection|...)$' -count=1` → `ok ... 23.352s` (7 subtests); `go test ./internal/factory/ -run '^(TestCandidateRecord...|TestLandingCheck...|TestMergeStep)' -count=1` → `ok ... 71.534s`; `go test ./internal/factorylane/ -count=1` → `ok ... 64.472s`; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0; gofmt clean.

### M3 — ci/** trigger + concurrency exception (card t1478, commits 29c81d5fa / 6e968bf93)

| Item | What changed |
|------|--------------|
| Files | `.github/workflows/ci.yml` — push.branches gains `ci/**` (no paths: filter, by design); concurrency.cancel-in-progress becomes `$\{\{ github.ref != 'refs/heads/main' && github.ref != 'refs/heads/develop' \}\}` (main's CI IS the integration verdict per AGENTS.local.md §4.1 canon — the follow-up commit added main after the first cut exempted only develop); B6 re-read: both race-job `if:` conditions branch on `startsWith(github.head_ref, 'release/')`, false on push events → candidate pushes RUN the race jobs |

- Baseline→post: `actionlint .github/workflows/ci.yml` exit 0 on BOTH sides of each edit; `git grep -n "ci/\*\*" .github/workflows/ci.yml` → 3 rows (trigger :25, comment :18/:44).

### M4 — landing gate at both call sites + per-card red hold + verdict observation (card t1478, commits 745bfae64 / b02ba30e3)

| Item | What changed |
|------|--------------|
| Files | `internal/factory/candidate_landing_check.go` (new — the shared check: record read, branch identity refuses outright, tip equality voids with re-candidate guidance, green-or-refuse, ancestry), `internal/factory/integration_merge_step.go` (fail-closed nil-seam guard at gate 5; in-section landing RECHECK under the candidate mutation lock — the batch TOCTOU repair), `internal/cli/integration_merge.go` (real check wired, placeholder gone), `internal/cli/factory_card.go` (complete call site wires the SAME check), `internal/cli/integration.go` (candidateAcquirePrecondition before any record mutation; root-resolution git child scrubbed), `internal/factory/candidate_record.go` (ObserveCandidateVerdict with the verdict-SHA binding + under-lock re-read; LatestCandidateRecord; Seq; scan opens non-blocking), `internal/cli/integration_candidate.go` (--observe; ghRunStates with the workflow filter; required-check verdict from required-checks.yml over the SHA's cross-workflow check runs; Seq stamping), tests: candidate_landing_check_test.go, candidate_landing_gate_test.go (new) |

- RED (this run): `go vet` → `undefined: CandidateLandingCheck` (factory), `undefined: factory.LatestCandidateRecord` (cli).
- GREEN: `go test ./internal/factory/ ./internal/cli/ -run '^(TestLandingCheck|TestCandidateVerdict|TestLandingCheckRefusesTargetMismatch|TestCandidateAcquirePrecondition|TestCompleteRefusesRedCandidate|TestCandidateObserveWalk|TestCandidateObserveCommand|TestMergeStepHappyPathCreatesNoFFMergeAndReleases|TestMergeStepPreMergeCausesReleaseWithDistinctCodes|TestIntegrationCandidate)$' -count=1` → `ok internal/factory 68.704s` / `ok internal/cli 16.453s`; `TestLandingCheckRecheckCatchesRedAfterGate` (the gate→observe→merge TOCTOU) → cause 5 at the merge point.
- Coverage (this run, coverprofile over the candidate families): factory candidate surfaces avg 84.4% over 16 funcs (CandidateLandingCheck 90.5%, WithCandidateMutation 90.9%, candidateStoreRealPath 91.7%); cli runIntegrationCandidate 87.2%, candidateCallerTreeGuard 91.7%. Sub-85 residuals are guard branches: candidateRecordPath 60% (invalid-key error branches), shortSHAFull 66.7% (≤12-char branch), WriteCandidateRecord 70% (error paths), ghRunStates/candidateGhRunner 0% (network shims by design, seams tested instead).
- KNOWN CONSEQUENCE (reported to coordinator, config-level): required-checks.yml's main contexts include `Release PR Multi-OS Gate` and `Analyze (Go) (go)`, which never publish on ci/** pushes (codeql.yml/release workflows are main-PR-triggered) — the required-check verdict holds every candidate red until the SSoT gains a candidate-scoped set. The verdict names the unpublished contexts in its `why` — honest red, not a false green.

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
