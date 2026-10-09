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

run_status: audit-ready (four review findings and the jobs-view pin; the full AC matrix, CI, live gh, and codex review are not measured here, see Gaps)
run_complete_at: 2026-10-10
run_commit_sha: 07323c990 (last run-phase code commit; this §E.3 commit does not cite its own SHA)
run_branch: WT-10-03-tier
run_range: 4efb4212e..07323c990 (absorb base 4efb4212e excluded)
run_commits: fbc48c021 (F1), 47034f5e9 (F2), ad313aaeb (F3), 03061fe21 (F4), 08731c463 (the leader's ci.yml count-comment commit, not lane work), 07323c990 (F5)
run_diff_stat: 12 files changed, 982 insertions(+), 97 deletions(-) (git diff --stat 4efb4212e HEAD)
run_spec_body_touched: none in the range (the stat lists Go files and .github/workflows/ci.yml only)
run_push: none (no push in this close)
go_files_changed_in_this_close: none (this close changes progress.md only)
red_green_ordering: a run-record claim, not a commit-graph fact (each finding's RED test and its fix share one commit; verification-claim-integrity.md §2.3)
ac_matrix: not re-measured in this close (see Gaps)

Attribution. Agent-run is what this agent ran. The RED outputs for F1 to F5 were observed earlier in this run, before each fix, and are quoted from that output. They were not re-executed here, because re-executing RED needs the Go files reverted and this close changes no Go file. The GREEN, build, format, vet, selector, and gh help results were observed again in this close at HEAD 07323c990. Lane-run is the lane's re-run as reported by the coordinator; it is not observed in this record (see Lane-run below). WARN config-notice lines and go-test run-header lines are omitted from the quoted outputs; every other line is verbatim.

### F1 — the first binding run decides the verdict observation (commit fbc48c021)

Files (git log --stat): internal/cli/candidate_landing_gate_test.go (89 insertions), internal/cli/integration_candidate.go (28 changed lines).

RED — Agent-run, earlier in this run, before the fix:

```bash
go test -count=1 -v -run '^TestObserveCandidateRunsFirstBindingRunDecides$' ./internal/cli/
```

Exit code 1:

```text
    candidate_landing_gate_test.go:132: wrote=true, stored verdict "green": the older run 100 must not decide over the newer binding run 200 — want wrote=false, pending
    candidate_landing_gate_test.go:149: wrote=true, stored verdict "green": the in-progress binding run 200 decides — want wrote=false, pending
--- FAIL: TestObserveCandidateRunsFirstBindingRunDecides (0.01s)
    --- FAIL: TestObserveCandidateRunsFirstBindingRunDecides/newest_binding_run_with_an_empty_conclusion_decides:_an_older_success_writes_nothing (0.00s)
    --- FAIL: TestObserveCandidateRunsFirstBindingRunDecides/newest_binding_run_still_in_progress_decides:_an_older_success_writes_nothing (0.00s)
    --- PASS: TestObserveCandidateRunsFirstBindingRunDecides/newest_binding_run_failed_decides_red;_a_non-binding_run_is_skipped_without_stopping_the_walk (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.092s
FAIL
```

GREEN — Agent-run, this close, HEAD 07323c990:

```bash
go test -count=1 -v -run '^TestObserveCandidateRunsFirstBindingRunDecides$' ./internal/cli/ ; echo "CLOSE_F1_GREEN_EXIT=$?"
```

```text
--- PASS: TestObserveCandidateRunsFirstBindingRunDecides (0.00s)
    --- PASS: TestObserveCandidateRunsFirstBindingRunDecides/newest_binding_run_with_an_empty_conclusion_decides:_an_older_success_writes_nothing (0.00s)
    --- PASS: TestObserveCandidateRunsFirstBindingRunDecides/newest_binding_run_still_in_progress_decides:_an_older_success_writes_nothing (0.00s)
    --- PASS: TestObserveCandidateRunsFirstBindingRunDecides/newest_binding_run_failed_decides_red;_a_non-binding_run_is_skipped_without_stopping_the_walk (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.931s
CLOSE_F1_GREEN_EXIT=0
```

### F2 — a stale observation cannot overwrite a newer verdict (commit 47034f5e9)

Files (git log --stat): internal/factory/candidate_record.go (93 changed lines), internal/factory/candidate_record_order_test.go (118 insertions), internal/factory/candidate_landing_check_test.go (14 changed lines), internal/cli/candidate_landing_gate_test.go (33 changed lines), internal/cli/integration_candidate.go (9 changed lines).

Fixture note: the F1 and F2 test fixtures used non-numeric run IDs (r-mine, r-1, r-9, r-red), which the numeric rule refuses. They were changed to numeric IDs (7001, 1, 9, 9). The assertions were kept. The refusal of non-numeric IDs is asserted by the subtest "a non-numeric or empty run id is refused and never green" below.

RED — Agent-run, earlier in this run, before the fix:

```bash
go test -count=1 -v -run '^TestObserveCandidateRunOrder$' ./internal/factory/
```

Exit code 1:

```text
    candidate_record_order_test.go:48: wrote=true, stored verdict "green" run "100": run 100 is older than the recorded run 200 — want wrote=false, red, run 200
    candidate_record_order_test.go:70: wrote=true, stored verdict "red" attempt 0: attempt 1 arriving after attempt 2 must be refused — want wrote=false, green, attempt 2
    candidate_record_order_test.go:80: run id "": wrote=true err=<nil> — an unorderable run must be refused with an error
    candidate_record_order_test.go:87: run id "": stored verdict "green", want pending
    candidate_record_order_test.go:80: run id "r-9": wrote=true err=<nil> — an unorderable run must be refused with an error
    candidate_record_order_test.go:87: run id "r-9": stored verdict "green", want pending
    candidate_record_order_test.go:80: run id " 9": wrote=true err=<nil> — an unorderable run must be refused with an error
    candidate_record_order_test.go:87: run id " 9": stored verdict "green", want pending
    candidate_record_order_test.go:80: run id "-9": wrote=true err=<nil> — an unorderable run must be refused with an error
    candidate_record_order_test.go:87: run id "-9": stored verdict "green", want pending
    candidate_record_order_test.go:97: first observation: wrote=true err=<nil> verdict "green" run "300" attempt 0, want green from run 300 attempt 1
    candidate_record_order_test.go:115: run 199 against the legacy run 200: wrote=true err=<nil>, want refused
--- FAIL: TestObserveCandidateRunOrder (0.04s)
    --- FAIL: TestObserveCandidateRunOrder/a_stale_successful_run_cannot_overwrite_a_newer_red_verdict (0.01s)
    --- FAIL: TestObserveCandidateRunOrder/a_re-run_attempt_may_replace_its_earlier_attempt,_never_the_reverse (0.01s)
    --- FAIL: TestObserveCandidateRunOrder/a_non-numeric_or_empty_run_id_is_refused_and_never_green (0.01s)
    --- FAIL: TestObserveCandidateRunOrder/the_first_observation_of_a_record_with_no_prior_run_keeps_working (0.00s)
    --- FAIL: TestObserveCandidateRunOrder/a_record_written_before_attempts_reads_as_attempt_1_of_its_run (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/factory	0.462s
FAIL
```

```bash
go test -count=1 -v -run '^TestGhRunStatesMapsRunAttempt$' ./internal/cli/
```

Exit code 1:

```text
    candidate_landing_gate_test.go:251: gh args "run list --workflow ci.yml --branch ci/t9001 --limit 20 --json databaseId,headSha,headBranch,status,conclusion": want the run query to ask for the attempt field
    candidate_landing_gate_test.go:254: runs [{RunID:200 HeadSHA:c1 Ref:ci/t9001 Status:completed Conclusion:success Attempt:0}]: want run 200 at attempt 2
--- FAIL: TestGhRunStatesMapsRunAttempt (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.325s
FAIL
```

GREEN — Agent-run, this close, HEAD 07323c990:

```bash
go test -count=1 -v -run '^TestObserveCandidateRunOrder$' ./internal/factory/ ; echo "CLOSE_F2A_GREEN_EXIT=$?"
```

```text
--- PASS: TestObserveCandidateRunOrder (0.02s)
    --- PASS: TestObserveCandidateRunOrder/a_stale_successful_run_cannot_overwrite_a_newer_red_verdict (0.01s)
    --- PASS: TestObserveCandidateRunOrder/a_re-run_attempt_may_replace_its_earlier_attempt,_never_the_reverse (0.00s)
    --- PASS: TestObserveCandidateRunOrder/a_non-numeric_or_empty_run_id_is_refused_and_never_green (0.00s)
    --- PASS: TestObserveCandidateRunOrder/the_first_observation_of_a_record_with_no_prior_run_keeps_working (0.00s)
    --- PASS: TestObserveCandidateRunOrder/a_record_written_before_attempts_reads_as_attempt_1_of_its_run (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/factory	0.374s
CLOSE_F2A_GREEN_EXIT=0
```

```bash
go test -count=1 -v -run '^TestGhRunStatesMapsRunAttempt$' ./internal/cli/ ; echo "CLOSE_F2B_GREEN_EXIT=$?"
```

```text
--- PASS: TestGhRunStatesMapsRunAttempt (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.665s
CLOSE_F2B_GREEN_EXIT=0
```

### F3 — one guard set in both selectors; the census check runs once (commit ad313aaeb)

Files (git log --stat): .github/workflows/ci.yml (12 insertions, 14 deletions).

RED — Agent-run, earlier in this run, before the fix. This is a probe over the workflow file, not a Go test. Its exit status was not captured.

```bash
python3 -c '
import re
t = open(".github/workflows/ci.yml", encoding="utf-8").read()
skip = re.search(r"-skip .\^\((.*?)\)\$.", t).group(1)
run = re.search(r"-run .\^\((.*?)\)\$. \./internal/cli/", t).group(1)
names = ["TestDestructiveTargetRegistry_CoversAllSites", "TestCodexCommand_GuardFileLiteralsNeutral", "TestProductionStringLiteralsUseLeaderLaneVocabulary"]
for n in names:
    in_skip = re.fullmatch(skip, n) is not None
    in_run = re.fullmatch(run, n) is not None
    print(n, "| ordinary job skips it:", in_skip, "| guard job runs it:", in_run)
print("census-check.sh occurrences in ci.yml:", t.count("census-check.sh"))
'
```

```text
TestDestructiveTargetRegistry_CoversAllSites | ordinary job skips it: False | guard job runs it: False
TestCodexCommand_GuardFileLiteralsNeutral | ordinary job skips it: False | guard job runs it: False
TestProductionStringLiteralsUseLeaderLaneVocabulary | ordinary job skips it: False | guard job runs it: False
census-check.sh occurrences in ci.yml: 3
```

GREEN — Agent-run, this close, HEAD 07323c990. Same probe, with the name-set comparison and the run-invocation count added:

```bash
python3 -I -c '
import re
t = open(".github/workflows/ci.yml", encoding="utf-8").read()
skip = re.search(r"-skip .\^\((.*?)\)\$.", t).group(1)
run = re.search(r"-run .\^\((.*?)\)\$. \./internal/cli/", t).group(1)
names = ["TestDestructiveTargetRegistry_CoversAllSites", "TestCodexCommand_GuardFileLiteralsNeutral", "TestProductionStringLiteralsUseLeaderLaneVocabulary"]
for n in names:
    print(n, "| ordinary skips:", re.fullmatch(skip, n) is not None, "| guard runs:", re.fullmatch(run, n) is not None)
print("same name set:", sorted(skip.split("|")) == sorted(run.split("|")), "| guard names:", len(run.split("|")))
print("census-check.sh run invocations:", len(re.findall(r"run: bash scripts/ci-census/census-check.sh", t)))
' ; echo "CLOSE_SELECTOR_PROBE_EXIT=$?"
```

```text
TestDestructiveTargetRegistry_CoversAllSites | ordinary skips: True | guard runs: True
TestCodexCommand_GuardFileLiteralsNeutral | ordinary skips: True | guard runs: True
TestProductionStringLiteralsUseLeaderLaneVocabulary | ordinary skips: True | guard runs: True
same name set: True | guard names: 8
census-check.sh run invocations: 1
CLOSE_SELECTOR_PROBE_EXIT=0
```

The two census figures are different measures: the RED line counts every textual occurrence (comments included), and the GREEN line counts `run:` invocations.

```bash
actionlint .github/workflows/ci.yml ; echo "CLOSE_ACTIONLINT_EXIT=$?"
```

```text
CLOSE_ACTIONLINT_EXIT=0
```

### F4 — the per-card red hold holds at every window grant (commit 03061fe21)

Files (git log --stat): internal/cli/candidate_landing_gate_test.go (140 insertions), internal/cli/factory_card.go (18 insertions), internal/cli/integration_candidate.go (23 changed lines), internal/cli/integration_wait.go (30 changed lines), internal/factory/candidate_window_hold_test.go (169 insertions), internal/factory/integration_lock.go (18 changed lines), internal/factory/integration_merge_step.go (2 changed lines), internal/factory/integration_window_ops.go (162 changed lines).

RED — Agent-run, earlier in this run, before the fix:

```bash
go test -count=1 -v -run '^TestWindowHold' ./internal/factory/
```

Exit code 1:

```text
    candidate_window_hold_test.go:62: no gate under the enabled key
    candidate_window_hold_test.go:86: no gate under the enabled key
--- FAIL: TestWindowHoldGate (0.01s)
    --- PASS: TestWindowHoldGate/the_gate_is_off_unless_workflow.candidate_ci.enabled_is_true (0.00s)
    --- FAIL: TestWindowHoldGate/with_the_key_on,_a_red_card_is_refused_and_every_other_card_is_admitted (0.00s)
    --- FAIL: TestWindowHoldGate/an_unreadable_candidate_record_refuses_fail-closed (0.00s)
    candidate_window_hold_test.go:103: acquire for a red card: err <nil>, want a candidate-hold refusal
--- FAIL: TestWindowHoldDirectGrant (0.00s)
    --- FAIL: TestWindowHoldDirectGrant/a_red_card's_direct_acquire_is_refused_before_any_window_record_is_written (0.00s)
    --- PASS: TestWindowHoldDirectGrant/a_green_card's_direct_acquire_is_granted_as_before (0.00s)
    candidate_window_hold_test.go:135: holder "sess-red" on card "tRed" after the refresh, want sess-green on tGreen: the red ticket must not take the window
    candidate_window_hold_test.go:160: holder "sess-red" after the refresh, want the window free: the only queued ticket is refused by the red hold
--- FAIL: TestWindowHoldQueuePromotion (0.00s)
    --- FAIL: TestWindowHoldQueuePromotion/a_red_ticket_is_withdrawn_at_promotion_and_the_next_admitted_ticket_takes_the_window (0.00s)
    --- FAIL: TestWindowHoldQueuePromotion/a_stale_holder_with_only_refused_tickets_queued_leaves_the_window_free (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/factory	0.424s
FAIL
```

```bash
go test -count=1 -v -run '^(TestWaitPromotionRefusesCardTurnedRedWhileQueued|TestCompleteAcquisitionRefusesRedCandidate)$' ./internal/cli/
```

Exit code 1:

```text
    candidate_landing_gate_test.go:658: tB did not return: the refused waiter blocked the queue behind it
--- FAIL: TestWaitPromotionRefusesCardTurnedRedWhileQueued (5.07s)
--- PASS: TestCompleteAcquisitionRefusesRedCandidate (1.24s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	7.226s
FAIL
```

The PASS line above for TestCompleteAcquisitionRefusesRedCandidate is a vacuous pass: the hook seam was declared but not yet called. The seam was then wired before the acquisition, and the run below is the RED that counts.

```bash
go test -count=1 -v -run '^TestCompleteAcquisitionRefusesRedCandidate$' ./internal/cli/
```

Exit code 1:

```text
    candidate_landing_gate_test.go:709: complete attempted the window acquisition 1 time(s) for a red card — the refusal must precede every grant
--- FAIL: TestCompleteAcquisitionRefusesRedCandidate (0.97s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	2.176s
FAIL
```

GREEN — Agent-run, this close, HEAD 07323c990:

```bash
go test -count=1 -v -run '^TestWindowHold' ./internal/factory/ ; echo "CLOSE_F4A_GREEN_EXIT=$?"
```

```text
--- PASS: TestWindowHoldGate (0.02s)
    --- PASS: TestWindowHoldGate/the_gate_is_off_unless_workflow.candidate_ci.enabled_is_true (0.01s)
    --- PASS: TestWindowHoldGate/with_the_key_on,_a_red_card_is_refused_and_every_other_card_is_admitted (0.01s)
    --- PASS: TestWindowHoldGate/an_unreadable_candidate_record_refuses_fail-closed (0.00s)
--- PASS: TestWindowHoldDirectGrant (0.00s)
    --- PASS: TestWindowHoldDirectGrant/a_red_card's_direct_acquire_is_refused_before_any_window_record_is_written (0.00s)
    --- PASS: TestWindowHoldDirectGrant/a_green_card's_direct_acquire_is_granted_as_before (0.00s)
--- PASS: TestWindowHoldQueuePromotion (0.00s)
    --- PASS: TestWindowHoldQueuePromotion/a_red_ticket_is_withdrawn_at_promotion_and_the_next_admitted_ticket_takes_the_window (0.00s)
    --- PASS: TestWindowHoldQueuePromotion/a_stale_holder_with_only_refused_tickets_queued_leaves_the_window_free (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/factory	0.264s
CLOSE_F4A_GREEN_EXIT=0
```

```bash
go test -count=1 -v -run '^(TestWaitPromotionRefusesCardTurnedRedWhileQueued|TestCompleteAcquisitionRefusesRedCandidate)$' ./internal/cli/ ; echo "CLOSE_F4B_GREEN_EXIT=$?"
```

```text
--- PASS: TestWaitPromotionRefusesCardTurnedRedWhileQueued (0.05s)
--- PASS: TestCompleteAcquisitionRefusesRedCandidate (0.90s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.692s
CLOSE_F4B_GREEN_EXIT=0
```

### F5 — the run jobs view pinned to the observed attempt (commit 07323c990)

Files (git log --stat): internal/cli/candidate_landing_gate_test.go (84 changed lines), internal/cli/integration_candidate.go (23 changed lines).

Caller change, verbatim from the commit diff (observed this run):

```diff
-			if adjusted := candidateRunVerdict(root, run.RunID, run.Status, run.Conclusion, targetBranch); adjusted.authoritative || adjusted.conclusion == "" {
+			if adjusted := candidateRunVerdict(root, run.RunID, candidateObservedAttempt(run.Attempt), run.Status, run.Conclusion, targetBranch); adjusted.authoritative || adjusted.conclusion == "" {
```

Jobs read at HEAD 07323c990 (grep, this close): `internal/cli/integration_candidate.go:533: jobsRaw, err := candidateGhRunsListFn(root, "run", "view", runID, "--attempt", fmt.Sprintf("%d", attempt), "--json", "jobs")`

RED — Agent-run, earlier in this run, before the fix:

```bash
go test -count=1 -v -run '^TestCandidateRunVerdictPinsTheObservedAttempt$' ./internal/cli/ ; echo "RED_EXIT=$?"
```

```text
=== RUN   TestCandidateRunVerdictPinsTheObservedAttempt
    candidate_landing_gate_test.go:313: jobs view argv "run view 200 --json jobs": want --attempt 1, the observed attempt, so the view cannot answer with the rerun
    candidate_landing_gate_test.go:316: observed: wrote=true verdict "green" attempt 1, want red recorded under the observed attempt 1 — the rerun's green must be neither judged nor recorded under attempt 1
--- FAIL: TestCandidateRunVerdictPinsTheObservedAttempt (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.967s
FAIL
RED_EXIT=1
```

GREEN — Agent-run, this close, HEAD 07323c990:

```bash
go test -count=1 -v -run '^TestCandidateRunVerdictPinsTheObservedAttempt$' ./internal/cli/ ; echo "CLOSE_F5_GREEN_EXIT=$?"
```

```text
--- PASS: TestCandidateRunVerdictPinsTheObservedAttempt (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.717s
CLOSE_F5_GREEN_EXIT=0
```

### Agent-run — build, format, vet, callers, gh help (this close, HEAD 07323c990)

```bash
go build ./... ; echo "CLOSE_GO_BUILD_EXIT=$?"
```

```text
CLOSE_GO_BUILD_EXIT=0
```

```bash
GOOS=windows GOARCH=amd64 go build ./... ; echo "CLOSE_WINDOWS_BUILD_EXIT=$?"
```

```text
CLOSE_WINDOWS_BUILD_EXIT=0
```

```bash
gofmt -l internal/cli/candidate_landing_gate_test.go internal/cli/factory_card.go internal/cli/integration_candidate.go internal/cli/integration_wait.go internal/factory/candidate_landing_check_test.go internal/factory/candidate_record.go internal/factory/candidate_record_order_test.go internal/factory/candidate_window_hold_test.go internal/factory/integration_lock.go internal/factory/integration_merge_step.go internal/factory/integration_window_ops.go ; echo "CLOSE_GOFMT_EXIT=$?"
```

```text
CLOSE_GOFMT_EXIT=0
```

(gofmt -l listed no file; the 11 files are every Go file in 4efb4212e..07323c990.)

```bash
go vet ./internal/cli/ ./internal/factory/ ; echo "CLOSE_VET_EXIT=$?"
```

```text
CLOSE_VET_EXIT=0
```

```bash
go test -count=1 -run '^(TestCandidate|TestIntegrationCandidate|TestCompleteRefusesRedCandidate|TestObserveCandidateRuns|TestGhRunStates|TestWaitPromotionRefuses|TestCompleteAcquisitionRefuses)' ./internal/cli/ ; echo "CLOSE_CALLERS_EXIT=$?"
```

```text
ok  	github.com/modu-ai/moai-adk/internal/cli	21.508s
CLOSE_CALLERS_EXIT=0
```

```bash
gh run view --help
```

Local help text, no network call. The relevant lines:

```text
  -a, --attempt uint      The attempt number of the workflow run
  # View a specific run with specific attempt number
  $ gh run view 12345 --attempt 3
```

The help does not state what an unpinned `gh run view` returns, so the default attempt is not established by this output.

### Lane-run — reported by the coordinator, not observed in this record

- The coordinator reports commit 07323c990 verified in the lane. This agent did not observe the lane's commands or their output, and no lane output is transcribed here.
- Commands the lane re-ran, as specified by the coordinator: `go build ./...`; `gofmt -l` on internal/cli/integration_candidate.go and internal/cli/candidate_landing_gate_test.go; the new test TestCandidateRunVerdictPinsTheObservedAttempt; the callers' tests.

### Residual risk

- RED-before-GREEN ordering is a run-record claim. Each finding's RED test and its fix share one commit (see the git log --stat figures above), so the commit graph cannot show RED before GREEN (verification-claim-integrity.md §2.3).
- The pin's premise is stated in code. The comment at internal/cli/integration_candidate.go ("THE PINNED ATTEMPT") says an unpinned jobs view answers with the newest attempt. That premise came from the codex finding and is not observed here (see Gaps). The pin is correct whatever the default is, but the comment's claim is unverified.
- The known consequence recorded in the M4 entry (required-checks contexts that never publish on ci/** pushes) is carried forward and was not re-measured in this close.

### Gaps

- Full `./...` test suite: not run.
- CI: no CI run observed for any commit in 4efb4212e..07323c990.
- Live `gh` call: none made. The `--attempt` behavior against GitHub is not observed, and the default attempt of an unpinned view is unverified.
- Codex review: not re-run on 07323c990 or on the final tree.
- Windows runtime tests: not run. Only the cross-build `GOOS=windows GOARCH=amd64 go build ./...` was observed (exit 0).
- Not re-measured in this close: the 25 acceptance criteria (ac_count 25 in §E.1), test coverage, and golangci-lint.
- The RED outputs above are quoted from this run's earlier output and were not re-executed here.
- No compile-stage RED output for F2 and F4 is in the record; only the behavioral RED output is quoted.
- Lane-run results (above) are not observed in this record.

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase_

## Lane Kickoff Decision Record (2026-10-09, lane-11)

decision record: decided_by=claude+lane-11 (self-dispatch lane, ladder terminal) evidence_refs=.moai/reports/t1478/verdict.md (PASS-WITH-DEBT 1.0, must_pass 0, blocking 0; receipts rcpt-69038b8e4d8f0c9ad4499861, rcpt-37a7694cccd61c7df7470e64, rcpt-b38299720acf23c31b7a497a; Addendum 2 records the final-tree codex re-run) ladder_path=plan→run Kickoff gate, autonomous form (audit cross + evidence criteria met)

Judgment: PASS-WITH-DEBT is a passing form — 0 must-pass failures, 0 blocking; the codex required-backend machine disagreement is transparently adjudicated per-item in the verdict (P2-A refuted with tree evidence ci.yml:67/:234/:246, P2-B/P2-C accepted as optional/debts N2/N1). Post-verdict deltas (bc4d46f63, b10bcfd0e) are M5 verification-command mechanics the auditor confirmed non-conflicting at b10bcfd0e. Debts 1-3 and optional N5-N7 travel with the SPEC; N5 (one-word doc fix) rides the sync phase.

Run-phase entry: M1 first (config gate workflow.candidate_ci.enabled + candidate record schema), RED-first per plan.md test-name conventions.
