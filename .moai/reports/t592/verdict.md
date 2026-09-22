# t592 run-phase verdict — SPEC-HOME-STATE-ROLLOUT-001

## Claim

F6-R2 보강으로 worktree inventory discovery 오류와 malformed/empty/path/canonicalization 이상이 모두 indeterminate error로 닫힌다. F16은 intervening merge 전체를 포함하던 연속 diff를 audited commit별 first-parent delta union으로 교체했다. 현재 exact diff-line validator는 **1202/1413 = 85.067%**다. 기존 구현은 독립 sync-audit PASS 100/100을 받았지만 이번 후속 변경은 **독립 재감사 대기**이며, 실제 사용자 홈 apply와 post-apply AC-HSR-022는 실행하지 않았다.

## Evidence

### TDD RED 원문

```text
go test ./internal/cli -run '^TestHomeStateDryRunNoMutation$' -count=1 -v
internal/cli/migrate_home_state_test.go:43:7: undefined: homeStateRunner
internal/cli/migrate_home_state_test.go:126:28: undefined: homestate.AcquireMigrationAdmission
FAIL github.com/modu-ai/moai-adk/internal/cli [build failed]
```

```text
go test ./internal/homestate -run '^TestFactoryV1ClaimedRowsUpgradeToV2$' -count=1 -v
internal/homestate/handoff_lease_test.go:61:22: f.ClaimResume undefined
internal/homestate/handoff_lease_test.go:101:14: f.FinishResume undefined
FAIL github.com/modu-ai/moai-adk/internal/homestate [build failed]
```

```text
go test ./internal/homestate -run '^TestProfileLeasesAreGlobalAndPrivate$' -count=1 -v
internal/homestate/profile_lease_test.go:13:16: undefined: OpenProfileLeases
internal/homestate/profile_lease_test.go:18:60: undefined: ProfileLease
FAIL github.com/modu-ai/moai-adk/internal/homestate [build failed]
```

```text
go test ./internal/cli -run '^TestHomeStateDryRunReport$' -count=1 -v
missing "active-census: sessions=1 factory=0 mcp=0" in ...
active-census: determinate zero (isolated runner)
--- FAIL: TestHomeStateDryRunReport
```

```text
go test ./internal/cli -run '^TestHomeStateRollbackVerifiedBackup$' -count=1 -v
parity rollback mutated target: entries 1 -> 3
--- FAIL: TestHomeStateRollbackVerifiedBackup
```

```text
go test ./internal/cli -run '^TestHomeStateVerdictEvidenceValidator$' -count=1 -v
internal/cli/migrate_home_state_test.go:407:9: ledger.Checks undefined
FAIL github.com/modu-ai/moai-adk/internal/cli [build failed]
```

### GREEN / quality gates

F6-R2 RED/GREEN:

```text
RED: internal/homestate/runtime_census_test.go:31:17: undefined: readRuntimeCensusWithWorktreeList
FAIL github.com/modu-ai/moai-adk/internal/homestate [build failed]

GREEN: TestRuntimeCensusFailsClosedWhenWorktreeInventoryUnavailable PASS
GREEN: TestRuntimeCensusFailsClosedOnMalformedOrUnavailableWorktreeInventory PASS
GREEN: TestRuntimeCensusCountsLiveAndIgnoresProvablyDead PASS
GREEN: TestRuntimeCensusIncludesLinkedWorktreeLocalRegistry PASS
ok github.com/modu-ai/moai-adk/internal/homestate 2.125s
```

F14 dedup mutant RED/GREEN:

```text
RED (dedup condition removed):
dedup failed: linked={ActiveSessions:3 ... Fingerprint:3:0:0} primary={ActiveSessions:3 ... Fingerprint:3:0:0}
--- FAIL: TestRuntimeCensusDeduplicatesSameLivePIDAcrossPrimaryAndLinkedRegistries

GREEN (production dedup restored):
--- PASS: TestRuntimeCensusDeduplicatesSameLivePIDAcrossPrimaryAndLinkedRegistries (0.87s)
PASS
ok github.com/modu-ai/moai-adk/internal/homestate 3.958s
```

```text
go test ./internal/cli ./internal/homestate -run '^(TestHomeState.*|TestRuntimeCensus.*|TestAdmission.*|TestFactory.*|TestResume.*|TestProfile.*|TestCleanHome.*|TestPlatformProcessIdentity.*)$' -count=1
ok github.com/modu-ai/moai-adk/internal/cli 48.308s
ok github.com/modu-ai/moai-adk/internal/homestate 4.302s
```

```text
go test -race ./internal/homestate -run '^(TestResumeLatestPendingThenExpiredReclaim|TestProfileLeaseReconcilePIDFingerprint|TestRuntimeCensusCountsLiveAndIgnoresProvablyDead)$' -count=1
ok github.com/modu-ai/moai-adk/internal/homestate 2.115s
go test -race ./internal/cli -run '^(TestHomeStateStartVsMigrateSerialized|TestProfileLeaseLifecycleAndNonExecCleanerRace)$' -count=1
ok github.com/modu-ai/moai-adk/internal/cli 27.602s
```

```text
go vet ./internal/homestate ./internal/hook/handoff ./internal/hook ./internal/kanban ./internal/cli
(empty), exit 0
go run ./cmd/moai spec lint SPEC-HOME-STATE-ROLLOUT-001 --strict --json
[]
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go test -c ... ./internal/homestate ./internal/cli
homestate.test.exe
cli.test.exe
git diff --check
(empty), exit 0
```

### Sync-audit 보강 focused GREEN

```text
TestHomeStateRefusesDivergentTarget PASS
TestHomeStateDryRunReportsUnreadableTargetWithoutMutation PASS
TestRunNamedTestsRejectsSkipFailAndZeroJSONEvents PASS (skip/fail/zero/duplicate-pass)
TestHomeStateVerdictEvidenceValidator PASS
TestPersistedHomeStateEvidenceIsImmutableAndReadsRealDatabases PASS
TestContinueLaunchCreatesSingleProvisionalLeaseBeforeChildStart PASS
ok github.com/modu-ai/moai-adk/internal/cli 5.618s
```

### Coverage

정적 allowlist를 제거하고 `git diff HEAD`와 untracked production Go 파일을 자동 도출했다. 플랫폼 전용 파일은 cross-compile disposition으로 분리했다.

```text
timeout 150s go test ./internal/cli ./internal/homestate -run '^(TestHomeState.*|TestRuntimeCensus.*|TestAdmission.*|TestFactory.*|TestResume.*|TestProfile.*|TestCleanHome.*|TestPlatformProcessIdentity.*|TestParseChangedSurfaceCoverage.*)$' -count=1 -coverpkg=./internal/cli,./internal/homestate -coverprofile="$tmp_profile"
ok  github.com/modu-ai/moai-adk/internal/cli        117.322s  coverage: 11.3% of statements in ./internal/cli, ./internal/homestate
ok  github.com/modu-ai/moai-adk/internal/homestate  3.137s    coverage: 2.4% of statements in ./internal/cli, ./internal/homestate

home_state_coverage_test.go:127: auto-diff changed production coverage: 1174/1408 = 83.381%
--- PASS: TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite (136.93s)
```

추가 테스트는 omitted/rename/deletion/platform disposition/malformed diff·profile, handoff CAS, evidence ledger와 실제 readback 실패 분기를 직접 실행한다. 대상은 정적 allowlist가 아니라 current HEAD production diff에서 매번 자동 도출한다.

Coverage RED는 `1174/1408 = 83.381%`였다. 테스트 selector에서 빠져 있던 기존 safety test 이름을 current selector 계약에 맞추고, malformed diff parser, named-test JSON event, backup manifest/recovery/rollback/evidence path, dry-run equivalent/divergent 분기를 테스트 전용으로 보강했다.

```text
timeout 300s go test ./internal/cli -run '^TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite$' -count=1 -v
    home_state_coverage_test.go:127: auto-diff changed production coverage: 1214/1425 = 85.193%
--- PASS: TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite (142.75s)
PASS
ok  github.com/modu-ai/moai-adk/internal/cli  143.831s
```

검증 중 첫 재시도는 hook의 500 ms timing assertion, 둘째 재시도는 내부 4분 deadline으로 각각 실패했다. hook selector 단독 PASS 뒤 동일 exact command를 다시 실행해 위 결과를 얻었다. validator, threshold, diff parser, denominator와 platform disposition은 변경하지 않았다.

### Coverage 이후 재검증

```text
timeout 120s go test -race ./internal/homestate -run '^(TestResumeLatestPendingThenExpiredReclaim|TestProfileLeaseReconcilePIDFingerprint|TestRuntimeCensusCountsLiveAndIgnoresProvablyDead)$' -count=1
ok  github.com/modu-ai/moai-adk/internal/homestate  1.984s

timeout 120s go test -race ./internal/cli -run '^(TestHomeStateStartVsMigrateSerialized|TestProfileLeaseLifecycleAndNonExecCleanerRace)$' -count=1
ok  github.com/modu-ai/moai-adk/internal/cli  30.597s

timeout 120s go vet ./internal/homestate ./internal/hook/handoff ./internal/hook ./internal/kanban ./internal/cli
(empty), exit 0

GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go test -c -o "$out_dir/homestate.test.exe" ./internal/homestate
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go test -c -o "$out_dir/cli.test.exe" ./internal/cli
homestate.test.exe 12048384 bytes
cli.test.exe       73379840 bytes

go run ./cmd/moai spec lint SPEC-HOME-STATE-ROLLOUT-001 --strict --json
[]

git diff --check
(empty), exit 0
```

### Post-commit clean-tree coverage 회귀와 보강

기존 resolver는 `git diff HEAD`만 사용해 clean commit/merge에서 `zero changed production files`로 실패했다. RED fixture는 clean repository에서 committed t592 change set을 요구했지만 `resolveHomeStateCoverageChangeSet`이 없어 build fail했다.

```text
timeout 30s go test ./internal/cli -run '^TestCommittedCoverageChangeSet' -count=1 -v
internal/cli/home_state_coverage_test.go:46:20: undefined: resolveHomeStateCoverageChangeSet
internal/cli/home_state_coverage_test.go:58:15: undefined: resolveHomeStateCoverageChangeSet
internal/cli/home_state_coverage_test.go:69:20: undefined: resolveHomeStateCoverageChangeSet
FAIL github.com/modu-ai/moai-adk/internal/cli [build failed]
```

GREEN 설계는 caller env/ref를 받지 않는다. Git 이력에서 exact original subject를 유일하게 찾고 그 첫 부모를 base로 고정한다. 이번 보강 commit은 exact remediation subject `fix(state): stabilize committed coverage evidence (t592)`로 식별하며 original tip의 descendant여야 한다. audited tip→HEAD의 모든 covered production blob ID가 동일해야 하고, HEAD→worktree tracked/untracked production diff만 합산한다. duplicate/unreachable/non-descendant/stale evidence는 fail closed다.

```text
timeout 90s go test ./internal/cli -run '^TestCommittedCoverage' -count=1 -v
--- PASS: TestCommittedCoverageChangeSetWorksInCleanRepositoryAndAfterUnrelatedCommit
--- PASS: TestCommittedCoverageChangeSetWorksAfterMergeCommit
--- PASS: TestCommittedCoverageChangeSetWorksCleanAfterRemediationCommit
--- PASS: TestCommittedCoverageChangeSetMergesDirtyProductionDiff
--- PASS: TestCommittedCoverageChangeSetRejectsStaleOrAmbiguousEvidence
--- PASS: TestCommittedCoverageChangeSetRejectsInvalidGitAndMalformedEvidence
PASS
ok github.com/modu-ai/moai-adk/internal/cli 10.946s

timeout 300s go test ./internal/cli -run '^TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite$' -count=1 -v
home_state_coverage_test.go:341: auto-diff changed production coverage: 1197/1408 = 85.014%
--- PASS: TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite (136.89s)
PASS
ok github.com/modu-ai/moai-adk/internal/cli 137.654s

timeout 120s go test -race ./internal/cli -run '^TestCommittedCoverage' -count=1
ok github.com/modu-ai/moai-adk/internal/cli 12.336s

timeout 120s go vet ./internal/homestate ./internal/hook/handoff ./internal/hook ./internal/kanban ./internal/cli
(empty), exit 0

GOOS=windows GOARCH=amd64 go test -c ./internal/cli -o /tmp/t592-cli.test.exe
GOOS=windows GOARCH=amd64 go test -c ./internal/homestate -o /tmp/t592-homestate.test.exe
(empty), exit 0

go run ./cmd/moai spec lint SPEC-HOME-STATE-ROLLOUT-001 --strict --json
[{"severity":"info","code":"OwnershipTransitionUnmeasured","message":"... commit 0c86e61d... has no Authored-By-Agent trailer ..."}]
exit 0; warning/error 0, 기존 provenance info 1

git diff --check
(empty), exit 0
```

F15 RED/GREEN은 production coverage runner가 각 child `go test`에 recursion guard를 직접 설정하는지 가짜 `go` executable로 검증한다. 기존 환경의 `PATH`와 log 경로는 유지하고, 다섯 child 모두 `MOAI_HOME_STATE_COVERAGE_CHILD=1`을 관측해야 한다.

```text
RED: TestCoverageRunnerSetsRecursionGuardOnEveryChild
coverage cli tests: exit status 42
coverage child recursion guard missing
--- FAIL: TestCoverageRunnerSetsRecursionGuardOnEveryChild

GREEN: TestCoverageRunnerSetsRecursionGuardOnEveryChild
--- PASS: TestCoverageRunnerSetsRecursionGuardOnEveryChild (0.13s)
PASS
ok github.com/modu-ai/moai-adk/internal/cli 1.116s
```

F16 RED는 original t592 commit 뒤 unrelated production branch가 merge되고 remediation/docs commit이 이어지는 그래프에서 연속 base→tip diff가 `internal/codexwiring/configtoml.go`를 잘못 포함하는 것을 재현했다. GREEN은 각 exact marker의 `marker^1 → marker` delta만 union한다. marker 순서·유일성·ancestry를 검증하고, path별 마지막 audited blob이 HEAD와 같은지 확인한 뒤 working production diff만 추가한다.

```text
RED: TestCommittedCoverageChangeSetExcludesInterveningMergedProduction
intervening merged production included: Native:[internal/codexwiring/configtoml.go internal/x/a.go]
--- FAIL: TestCommittedCoverageChangeSetExcludesInterveningMergedProduction

GREEN: timeout 120s go test ./internal/cli -run '^TestCommittedCoverage' -count=1 -v
--- PASS: TestCommittedCoverageChangeSetExcludesInterveningMergedProduction
--- PASS: TestCommittedCoverageChangeSetSupportsVersionedRemediationChain
--- PASS: TestCommittedCoverageChangeSetRejectsStaleOrAmbiguousEvidence
PASS
ok github.com/modu-ai/moai-adk/internal/cli 15.463s

timeout 300s go test ./internal/cli -run '^TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite$' -count=1 -v
home_state_coverage_test.go:433: auto-diff changed production coverage: 1202/1413 = 85.067%
--- PASS: TestHomeStateChangedSurfaceCoverageRunsBoundedFocusedSuite (143.77s)
PASS
ok github.com/modu-ai/moai-adk/internal/cli 144.550s
```

### 최종 focused F1–F13 및 gate 재검증

```text
timeout 150s go test ./internal/cli ./internal/homestate ./internal/hook ./internal/hook/handoff -run '^(TestCC_FactoryEntryThroughRunCC|TestGLM_FactoryWorkerEntry|TestContinueLaunchCreatesSingleProvisionalLeaseBeforeChildStart|TestDeadProvisionalLeaseRemainsIndeterminateUntilTransferDeadline|TestResumeFinishRejectsABAToken|TestAdmissionMarkerClearFailureIsObservable|TestRuntimeCensusIncludesLinkedWorktreeLocalRegistry|TestHomeStateDryRunClassifiesEquivalentAndDivergentTargets|TestHomeStateRunNamedTestsRejectsSkipFailAndZeroJSONEvents|TestHomeStateVerdictEvidenceValidator|TestPersistedHomeStateEvidenceIsImmutableAndReadsRealDatabases|TestCleanHomeReportsProtectedProfileReasonInDryRunAndForce|TestSessionStartProfileLeaseDirectAndTokenFallback)$' -count=1 -v
ok  github.com/modu-ai/moai-adk/internal/cli          9.454s
ok  github.com/modu-ai/moai-adk/internal/homestate    1.909s
ok  github.com/modu-ai/moai-adk/internal/hook         3.073s
ok  github.com/modu-ai/moai-adk/internal/hook/handoff 1.076s [no tests to run]

timeout 120s go test -race ./internal/homestate -run '^(TestResumeLatestPendingThenExpiredReclaim|TestProfileLeaseReconcilePIDFingerprint|TestRuntimeCensusCountsLiveAndIgnoresProvablyDead)$' -count=1
ok  github.com/modu-ai/moai-adk/internal/homestate  2.209s

timeout 120s go test -race ./internal/cli -run '^(TestHomeStateStartVsMigrateSerialized|TestProfileLeaseLifecycleAndNonExecCleanerRace)$' -count=1
ok  github.com/modu-ai/moai-adk/internal/cli  38.556s

timeout 120s go vet ./internal/homestate ./internal/hook/handoff ./internal/hook ./internal/kanban ./internal/cli
(empty), exit 0

GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go test -c -o /tmp/t592-final-homestate.exe ./internal/homestate
/tmp/t592-final-homestate.exe 12084736 bytes
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go test -c -o /tmp/t592-final-cli.exe ./internal/cli
/tmp/t592-final-cli.exe 73520640 bytes

go run ./cmd/moai spec lint SPEC-HOME-STATE-ROLLOUT-001 --strict --json
[]

git diff --check
(empty), exit 0
```

## Baseline-attribution

검증 수치는 worktree `.claude/worktrees/t592`, branch `WT-home-state-rollout`, HEAD `01e7a612e`와 현재 미커밋 diff에서 이번 run에 직접 측정했다. audited commit별 first-parent delta와 현재 working diff를 합친 exact validator의 1202/1413=85.067%를 유효한 기준으로 사용한다. intervening production merge, versioned remediation, later docs, dirty union은 모두 `t.TempDir()` Git fixture에서 검증했고 실제 사용자 홈은 사용하지 않았다.

## Gaps

- production `validateLivePreApply`의 fresh coverage 계산과 low/zero/error/tampered/missing profile 거부 및 exact changed-line coverage 85.067%는 GREEN이다. 기준 초과 여유가 작으므로 변경 시 즉시 재측정해야 한다.
- 다음 clean post-remediation 동작은 commit subject `fix(state): isolate committed coverage deltas (t592)`를 세 번째 trusted marker로 사용한다. 다른 제목이면 현재 working delta가 commit 후 evidence chain에서 사라지므로 exact 제목을 보존해야 한다.
- 이번 post-commit remediation은 기존 PASS 100/100 이후 변경이므로 독립 sync re-audit는 PENDING이다.
- 소유한 5개 변경 경로의 `git diff --check`는 exit 0이다. 전체 tree 검사는 동시 세션이 수정한 `.moai/reports/t592/sync-audit.md`의 기존 trailing whitespace 4건 때문에 exit 2였으며, 해당 외부 변경은 수정하거나 되돌리지 않았다.
- exact validator 실행 중 기존 hook timing assertion과 4분 내부 deadline 실패를 각각 한 번 관측했다. 최종 동일 command는 PASS했지만 느린 개발 머신에서의 timing 변동 위험은 남아 있다.
- 실제 live apply, post-apply AC-HSR-022 readback, integration-branch CI 전체 suite는 실행하지 않았다. 이 셋의 verdict는 PENDING이다.
- 전체 kanban package 회귀는 실행하지 않았고 변경 영향 selector와 요청된 최소 race set만 PASS다.

## Residual-risk

handoff 전달은 at-least-once다. 주입이 성공한 뒤 `consumed` CAS 전에 프로세스가 종료되면 같은 claimed/pending envelope가 재전달되어 중복 실행될 수 있으며, receiver-side exactly-once/deduplication은 구현되어 있지 않다. 또한 run-phase quality PASS는 live 데이터 이전 성공을 뜻하지 않는다. 실제 apply는 fresh zero-active census, current HEAD one-shot authorization, backup/restore probe를 다시 통과한 경우에만 실행하고, 이후 AC-HSR-022 readback과 독립 sync audit가 필요하다.
