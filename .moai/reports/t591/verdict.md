# t591 홈 상태 SQLite·handoff 검증 보고서

판정: **부분 통과**. 새로 확인한 결함을 RED→GREEN으로 보완했고 변경 관련 테스트,
race, vet, build, SQLite 백업·복원을 통과했다. 전체 저장소 CI와 실제 `~/.moai`
전환·강제 정리는 실행하지 않았다.

## Claim

1. 명시적 `MOAI_HOME`은 임시 경로 프로젝트에서도 우선하며 새 프로젝트별 DB 경로를 사용한다.
2. Todo와 Factory의 DB/WAL/SHM은 열린 동안 모두 0600이다.
3. 구 `~/.moai/todo/<project-key>` 큐와 비ASCII·symlink 도입 전 키를 새 Todo DB로
   논리 복사하고 원본을 롤백 자료로 보존한다.
4. 큐 이전은 source와 target lock을 모두 잡아 신규 홈 큐 쓰기를 덮어쓰지 않는다.
5. `modernc.org/sqlite`는 직접 의존성으로 선언됐고 필요한 체크섬이 `go.sum`에 있다.
6. SessionEnd memory는 DB의 오래된 pending보다 다른 legacy 파일을 우선 저장하며,
   처리한 SHA가 그대로인 파일만 원자적 claim 후 삭제한다.
7. Factory 진입은 run metadata를 기록하고, 실제 `todo next --spec` 카드 배정 이벤트가
   canonical SPEC 경로·SHA256·git commit·captured_at 스냅샷의 SSOT다. 카드 상태 변경도
   version 증가와 append-only event를 한 트랜잭션으로 기록한다.
8. 실패한 Todo 이전은 불완전 target DB artifacts를 제거하며 workers legacy import는 한 번만 실행된다.
9. resume TTL 만료는 관찰한 row id와 pending 상태를 조건으로 하는 CAS이고, 활성 Claude
   프로필은 홈 정리 후보에서 제외되며 retention=0에서도 force 권한 보정은 실행된다.

## Evidence

### RED

```text
env GOCACHE=/private/tmp/go-build-t591-red go test -count=1 ./internal/homestate -run 'Test(ProjectDirHonorsExplicitMoaiHomeForTemporaryProject|FactorySQLiteArtifactsArePrivateWhileOpen)'
--- FAIL: TestProjectDirHonorsExplicitMoaiHomeForTemporaryProject (0.14s)
    handoff_test.go:134: ProjectDir() = "/private/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestProjectDirHonorsExplicitMoaiHomeForTemporaryProject819170950/002/.moai/db/002-538623f2", want explicit MOAI_HOME path "/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestProjectDirHonorsExplicitMoaiHomeForTemporaryProject819170950/001/db/002-538623f2"
--- FAIL: TestFactorySQLiteArtifactsArePrivateWhileOpen (0.29s)
    handoff_test.go:157: factory.db-wal mode=0644, want 0600
    handoff_test.go:157: factory.db-shm mode=0644, want 0600
FAIL
FAIL github.com/modu-ai/moai-adk/internal/homestate 0.889s

env GOCACHE=/private/tmp/go-build-t591-red go test -count=1 ./internal/kanban -run TestRelocateQueueArtifactsSerializesOnTargetQueueLock
--- FAIL: TestRelocateQueueArtifactsSerializesOnTargetQueueLock (0.01s)
    state_dir_test.go:154: relocation bypassed target queue lock: <nil>
FAIL
FAIL github.com/modu-ai/moai-adk/internal/kanban 0.439s

env GOCACHE=/private/tmp/go-build-t591-red go test -count=1 ./internal/kanban -run 'Test(AdoptingPathMigratesLegacyHomeQueue|BacklogSQLiteArtifactsArePrivateWhileOpen)'
--- FAIL: TestAdoptingPathMigratesLegacyHomeQueue (0.09s)
    state_dir_test.go:179: adopting path="/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestAdoptingPathMigratesLegacyHomeQueue218594695/002/.moai/state/todo/backlog.json", want "/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestAdoptingPathMigratesLegacyHomeQueue218594695/001/db/002-84ee9965/todo/backlog.json"
--- FAIL: TestBacklogSQLiteArtifactsArePrivateWhileOpen (0.01s)
    state_dir_test.go:210: backlog.db-wal mode=0644, want 0600
    state_dir_test.go:210: backlog.db-shm mode=0644, want 0600
FAIL
FAIL github.com/modu-ai/moai-adk/internal/kanban 0.573s

env GOCACHE=/private/tmp/go-build-t591-red go test -count=1 ./internal/kanban -run TestAdoptingPathFindsLegacyHomeQueueWithPreSanitizationKey
--- FAIL: TestAdoptingPathFindsLegacyHomeQueueWithPreSanitizationKey (0.05s)
    state_dir_test.go:226: migrated items=0, want 4
FAIL
FAIL github.com/modu-ai/moai-adk/internal/kanban 0.478s
```

### GREEN 및 회귀 검사

```text
env GOCACHE=/private/tmp/go-build-t591-green go test -count=1 ./internal/homestate -run 'Test(ProjectDirHonorsExplicitMoaiHomeForTemporaryProject|FactorySQLiteArtifactsArePrivateWhileOpen)'
ok github.com/modu-ai/moai-adk/internal/homestate 0.857s
env GOCACHE=/private/tmp/go-build-t591-green go test -count=1 ./internal/kanban -run 'Test(RelocateQueueArtifactsSerializesOnTargetQueueLock|AdoptingPathMigratesLegacyHomeQueue|AdoptingPathFindsLegacyHomeQueueWithPreSanitizationKey|BacklogSQLiteArtifactsArePrivateWhileOpen)'
ok github.com/modu-ai/moai-adk/internal/kanban 1.522s

env GOCACHE=/private/tmp/go-build-t591 go test -race -count=1 ./internal/homestate -run 'Test(FactorySchemaAndResumeClaimOnce|ProjectDirHonorsExplicitMoaiHomeForTemporaryProject|FactorySQLiteArtifactsArePrivateWhileOpen)'
ok github.com/modu-ai/moai-adk/internal/homestate 2.527s
env GOCACHE=/private/tmp/go-build-t591 go test -race -count=1 ./internal/kanban -run 'Test(RelocateQueueArtifacts|AdoptingPath|BacklogSQLiteArtifactsArePrivateWhileOpen|ClaimFactoryWorkerNameConcurrentClaimsAreUnique)'
ok github.com/modu-ai/moai-adk/internal/kanban 3.173s

env GOCACHE=/private/tmp/go-build-t591 go test -timeout=5m -count=1 ./internal/kanban
ok github.com/modu-ai/moai-adk/internal/kanban 151.282s

env GOCACHE=/private/tmp/go-build-t591 go test -count=1 ./internal/hook/handoff
ok github.com/modu-ai/moai-adk/internal/hook/handoff 9.424s
env GOCACHE=/private/tmp/go-build-t591 go test -count=1 ./internal/hook -run 'Test(Handoff|ClaimThenInject|ConcurrentConsume|RenderHandoff|SessionStartRecord)'
ok github.com/modu-ai/moai-adk/internal/hook 2.480s
env GOCACHE=/private/tmp/go-build-t591 go test -count=1 ./internal/cli -run 'Test(CleanHome|ScanHomeCleanable|CheckHomeDisk|HomeDiskReport|SecureHomeDirectories|Handoff|TodoHelp|DeprecatedPaths|TempOriginGuidance|ResolveFactoryWorkerName|CC_FactoryEntryThroughRunCC|GLM_FactoryWorkerEntry)'
ok github.com/modu-ai/moai-adk/internal/cli 8.031s
env GOCACHE=/private/tmp/go-build-t591 go test -count=1 ./internal/web -run 'Test(ResolvedWatchPaths|LaneSection|FactoryLanes|HomeState)'
ok github.com/modu-ai/moai-adk/internal/web 10.055s

env GOCACHE=/private/tmp/go-build-t591 go vet ./internal/homestate ./internal/kanban ./internal/hook/handoff ./internal/hook ./internal/cli ./internal/profile ./internal/web ./internal/config ./internal/goal ./internal/template
(출력 없음, exit 0)
env GOCACHE=/private/tmp/go-build-t591 go build -o /private/tmp/moai-t591 ./cmd/moai
(출력 없음, exit 0)
go mod tidy -diff
(출력 없음, exit 0; 누락 체크섬 반영 후 재실행)
git diff --check
(출력 없음, exit 0)
```

### 추가 감사 RED→GREEN

```text
env GOCACHE=/private/tmp/go-build-t591-red2 go test -count=1 ./internal/hook/handoff -run TestPersistIfPending_NewerLegacyFileSupersedesPendingDatabaseRow
--- FAIL: TestPersistIfPending_NewerLegacyFileSupersedesPendingDatabaseRow (0.66s)
    persist_test.go:240: newer legacy handoff was not persisted: open .../project_sprint2_session-handoff-auto-001_plan_ready.md: no such file or directory
FAIL

env GOCACHE=/private/tmp/go-build-t591-red3 go test -count=1 ./internal/hook/handoff -run TestRemovePendingIfUnchangedPreservesConcurrentReplacement
persist_test.go:262:18: undefined: removePendingIfUnchanged
FAIL github.com/modu-ai/moai-adk/internal/hook/handoff [build failed]

env GOCACHE=/private/tmp/go-build-t591-red4 go test -count=1 ./internal/cli -run TestCCFactoryEntryRecordsFailOpenRunProvenance
--- FAIL: TestCCFactoryEntryRecordsFailOpenRunProvenance
    factory_test.go:763: factory run row missing: sql: no rows in result set
FAIL

env GOCACHE=/private/tmp/go-build-t591-red5 go test -count=1 ./internal/cli -run TestTodoPickInFactoryRecordsCardAndEvent
--- FAIL: TestTodoPickInFactoryRecordsCardAndEvent
    todo_test.go:120: factory card row missing: sql: no rows in result set
FAIL

env GOCACHE=/private/tmp/go-build-t591-red6 go test -count=1 ./internal/kanban -run TestRelocateQueueArtifactsRemovesIncompleteTargetOnWriteFailure
--- FAIL: TestRelocateQueueArtifactsRemovesIncompleteTargetOnWriteFailure
    state_dir_test.go:121: incomplete target survived at .../global/backlog.db: <nil>
FAIL

env GOCACHE=/private/tmp/go-build-t591-red7 go test -count=1 ./internal/homestate -run TestImportLegacyWorkersRunsOnlyOnceAfterRosterBecomesEmpty
--- FAIL: TestImportLegacyWorkersRunsOnlyOnceAfterRosterBecomesEmpty
    handoff_test.go:190: legacy workers reimported after roster emptied: count=1
FAIL

env GOCACHE=/private/tmp/go-build-t591-red8 go test -count=1 ./internal/homestate -run TestExpireResumeIfPendingDoesNotExpireNewerReplacement
internal/homestate/handoff_test.go:212:21: db.ExpireResumeIfPending undefined
FAIL github.com/modu-ai/moai-adk/internal/homestate [build failed]

env GOCACHE=/private/tmp/go-build-t591-red9 go test -count=1 ./internal/cli -run 'Test(CleanHome_RetentionFromHomeTier|ScanHomeCleanable_PreservesActiveProfile|TempOriginGuidance_SilentWithExplicitAbsoluteMOAIHome)'
--- FAIL: TestCleanHome_RetentionFromHomeTier/explicit_zero_disables_cleaning
    clean_home_test.go:250: force must repair directory permissions even when retention is disabled: mode=-rwxr-xr-x err=<nil>
--- FAIL: TestScanHomeCleanable_PreservesActiveProfile
    clean_home_test.go:271: active profile path selected for deletion: .../claude-profiles/active/debug/old.log
--- FAIL: TestTempOriginGuidance_SilentWithExplicitAbsoluteMOAIHome
    todo_temp_guard_test.go:114: explicit absolute MOAI_HOME was falsely reported as refused
FAIL

env GOCACHE=/private/tmp/go-build-t591-red10 go test -count=1 ./internal/cli -run TestTodoPickInFactoryRecordsCardAndEvent
--- FAIL: TestTodoPickInFactoryRecordsCardAndEvent
    todo_test.go:136: updated card=("lane-2","picked",1), want (lane-2,queued,2)
FAIL

env GOCACHE=/private/tmp/go-build-t591-green3 go test -count=1 ./internal/hook/handoff -run 'Test(PersistIfPending_NewerLegacyFileSupersedesPendingDatabaseRow|RemovePendingIfUnchangedPreservesConcurrentReplacement)'
ok github.com/modu-ai/moai-adk/internal/hook/handoff 1.142s
env GOCACHE=/private/tmp/go-build-t591-green4 go test -count=1 ./internal/cli -run 'Test(CCFactoryEntryRecordsFailOpenRunProvenance|TodoPickInFactoryRecordsCardAndEvent)'
ok github.com/modu-ai/moai-adk/internal/cli 3.350s
env GOCACHE=/private/tmp/go-build-t591-green9 go test -count=1 ./internal/cli -run 'Test(CleanHome_RetentionFromHomeTier|ScanHomeCleanable_PreservesActiveProfile|TempOriginGuidance_SilentWithExplicitAbsoluteMOAIHome)'
ok github.com/modu-ai/moai-adk/internal/cli 2.061s
env GOCACHE=/private/tmp/go-build-t591-green10 go test -count=1 ./internal/cli -run TestTodoPickInFactoryRecordsCardAndEvent
ok github.com/modu-ai/moai-adk/internal/cli 3.407s
```

### 최종 변경 범위 검증

```text
env GOCACHE=/private/tmp/go-build-t591-focus go test -count=1 ./internal/homestate ./internal/hook/handoff ./internal/hook ./internal/kanban ./internal/cli
ok github.com/modu-ai/moai-adk/internal/homestate 2.816s
ok github.com/modu-ai/moai-adk/internal/hook/handoff 10.443s
ok github.com/modu-ai/moai-adk/internal/hook 54.715s
ok github.com/modu-ai/moai-adk/internal/kanban 151.483s
internal/cli 전체 패키지는 장시간 실행으로 중단; 아래 focused CLI 검증으로 대체

env GOCACHE=/private/tmp/go-build-t591-race go test -race -count=1 ./internal/homestate -run 'Test(ExpireResumeIfPendingDoesNotExpireNewerReplacement|ImportLegacyWorkersRunsOnlyOnceAfterRosterBecomesEmpty|FactorySQLiteArtifactsArePrivateWhileOpen)'
ok github.com/modu-ai/moai-adk/internal/homestate 2.153s
env GOCACHE=/private/tmp/go-build-t591-race go test -race -count=1 ./internal/hook/handoff -run 'Test(PersistIfPending_NewerLegacyFileSupersedesPendingDatabaseRow|RemovePendingIfUnchangedPreservesConcurrentReplacement)'
ok github.com/modu-ai/moai-adk/internal/hook/handoff 2.058s
env GOCACHE=/private/tmp/go-build-t591-race go test -race -count=1 ./internal/kanban -run 'Test(RelocateQueueArtifactsRemovesIncompleteTargetOnWriteFailure|RecordFactoryRunStartCapturesSpecAndGitProvenance)'
ok github.com/modu-ai/moai-adk/internal/kanban 2.497s
env GOCACHE=/private/tmp/go-build-t591-race go test -race -count=1 ./internal/cli -run 'Test(CCFactoryEntryRecordsFailOpenRunProvenance|TodoPickInFactoryRecordsCardAndEvent|CleanHome_RetentionFromHomeTier|ScanHomeCleanable_PreservesActiveProfile|TempOriginGuidance_SilentWithExplicitAbsoluteMOAIHome)'
ok github.com/modu-ai/moai-adk/internal/cli 6.576s

go vet ./internal/homestate ./internal/hook/handoff ./internal/hook ./internal/kanban ./internal/cli
(출력 없음, exit 0)
go build ./cmd/...
(출력 없음, exit 0)
go mod tidy 후 go.mod/go.sum SHA256 비교
(변경 없음, exit 0)
git diff --check
(출력 없음, exit 0)
git diff | shasum -a 256
889c5049ea796f9d04b825b557f8a5169afe5d3158700b1386b5f1d865c127fd  -
```

### 재감사: 카드 배정 시점 SPEC provenance

```text
env GOCACHE=/private/tmp/go-build-t591-red11 go test -count=1 ./internal/cli -run TestTodoPickInFactoryRecordsCardAndEvent
--- FAIL: TestTodoPickInFactoryRecordsCardAndEvent (1.69s)
    todo_test.go:156: assignment provenance=map[card_id:t1 owner:lane-2 spec_id:SPEC-CARD-001 state:picked]
FAIL
FAIL github.com/modu-ai/moai-adk/internal/cli 2.759s

env GOCACHE=/private/tmp/go-build-t591-green11 go test -count=1 ./internal/cli -run TestTodoPickInFactoryRecordsCardAndEvent
ok github.com/modu-ai/moai-adk/internal/cli 3.597s
env GOCACHE=/private/tmp/go-build-t591-green11 go test -count=1 ./internal/cli -run 'TestTodoPickInFactory(RecordsCardAndEvent|ProvenanceFailsOpenWithoutSpecOrGit)'
ok github.com/modu-ai/moai-adk/internal/cli 5.646s

env GOCACHE=/private/tmp/go-build-t591-race2 go test -race -count=1 ./internal/cli -run 'Test(CCFactoryEntryRecordsFailOpenRunMetadata|TodoPickInFactory(RecordsCardAndEvent|ProvenanceFailsOpenWithoutSpecOrGit))'
ok github.com/modu-ai/moai-adk/internal/cli 8.071s
env GOCACHE=/private/tmp/go-build-t591-race2 go test -race -count=1 ./internal/kanban -run TestRecordFactoryRunStartRecordsMetadataWithoutClaimingSpecProvenance
ok github.com/modu-ai/moai-adk/internal/kanban 2.452s
env GOCACHE=/private/tmp/go-build-t591-race2 go test -count=1 ./internal/template -run 'TestBacklogJSONDisclosure_(EmbeddedTemplatesMatchSource|TemplateMirrorIsComplete)'
ok github.com/modu-ai/moai-adk/internal/template 0.472s

go vet ./internal/cli ./internal/kanban ./internal/homestate
(출력 없음, exit 0)
go build ./cmd/...
(출력 없음, exit 0)
go mod tidy 전후 go.mod/go.sum SHA256 비교
(변경 없음, exit 0)
git diff --check
(출력 없음, exit 0)
git diff | shasum -a 256
9ad133ca969634f78f9c00c7420c792550492f8f3867a923942f59915886f9a1  -
```

### 최종 감사: linked lane 실행 checkout provenance

```text
env GOCACHE=/private/tmp/go-build-t591-red12 go test -count=1 ./internal/cli -run 'TestTodoPickInFactory(ProvenanceFailsOpenWithoutSpecOrGit|CapturesLinkedWorktreeSpecAndHEAD)'
--- FAIL: TestTodoPickInFactoryProvenanceFailsOpenWithoutSpecOrGit (2.03s)
    todo_test.go:210: fail-open provenance=map[captured_at:2026-09-09T04:24:26.448681Z card_id:t1 git_commit: owner:lead spec_id:../../../outside spec_path:/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestTodoPickInFactoryProvenanceFailsOpenWithoutSpecOrGit3378785822/outside/spec.md spec_sha256: state:picked]
--- FAIL: TestTodoPickInFactoryCapturesLinkedWorktreeSpecAndHEAD (2.18s)
    todo_test.go:274: linked-lane provenance=map[captured_at:2026-09-09T04:24:28.760086Z card_id:t1 git_commit:3cc317a0b152e1762b62eb468a5df135663ec8d5 owner:lane-1 spec_id:SPEC-LANE-001 spec_path:/private/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestTodoPickInFactoryCapturesLinkedWorktreeSpecAndHEAD1697775946/001/.moai/specs/SPEC-LANE-001/spec.md spec_sha256:84307656ed68df5515adb78ef97fe650924f109895e5d0b8f98fcbc44ae37342 state:picked], want path=/private/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestTodoPickInFactoryCapturesLinkedWorktreeSpecAndHEAD1697775946/002/lane/.moai/specs/SPEC-LANE-001/spec.md commit=8b9932d79c42bcefd6060c49c3d49a57f50b78c9
FAIL
FAIL github.com/modu-ai/moai-adk/internal/cli 5.301s

env GOCACHE=/private/tmp/go-build-t591-green12 go test -count=1 ./internal/cli -run 'TestTodoPickInFactory(ProvenanceFailsOpenWithoutSpecOrGit|CapturesLinkedWorktreeSpecAndHEAD|RecordsCardAndEvent)'
ok github.com/modu-ai/moai-adk/internal/cli 7.805s

env GOCACHE=/private/tmp/go-build-t591-race3 go test -race -count=1 ./internal/cli -run 'TestTodoPickInFactory(ProvenanceFailsOpenWithoutSpecOrGit|CapturesLinkedWorktreeSpecAndHEAD|RecordsCardAndEvent)'
ok github.com/modu-ai/moai-adk/internal/cli 9.325s
env GOCACHE=/private/tmp/go-build-t591-race3 go test -race -count=1 ./internal/kanban -run TestRecordFactoryRunStartRecordsMetadataWithoutClaimingSpecProvenance
ok github.com/modu-ai/moai-adk/internal/kanban 2.534s
go vet ./internal/cli ./internal/kanban ./internal/homestate
(출력 없음, exit 0)
go build ./cmd/...
(출력 없음, exit 0)
go mod tidy 전후 go.mod/go.sum SHA256 비교
(변경 없음, exit 0)
git diff --check
(출력 없음, exit 0)
git diff | shasum -a 256
33ef562651ae652dd09a19ff2fe775e3ce875ebd17138d28361681a15313e697  -
```

### 임시 홈 백업·복원

실사용 홈과 분리한 `MOAI_HOME=/private/tmp/moai-t591-e2e.XEWElI`에서 실행했다.

```text
todo=t592 44
handoff=handoff saved: /private/tmp/moai-t591-e2e.XEWElI/db/moai-adk-go-1bd3d038/factory/factory.db
todo_integrity=ok
factory_integrity=ok
restore_integrity=ok
source_counts=75,228
restore_counts=75,228
600 .../factory/factory.db
600 .../factory/factory.db-shm
600 .../factory/factory.db-wal
600 .../todo/backlog.db
600 .../todo/backlog.db-shm
600 .../todo/backlog.db-wal
```

## Baseline-attribution

```text
HEAD: 817990f67
branch: WT-home-sqlite-handoff
origin/main...HEAD: 0 2518
session: 9cc4f7d4-a9c0-4b4b-8493-216a1782edef
```

측정 대상은 위 워크트리의 커밋되지 않은 변경이다. primary checkout은 수정하지 않았다.

## Gaps

- 실제 `~/.moai` force cleanup, 전역 마이그레이션, 바이너리 설치는 실행하지 않았다.
- 저장소 전체 테스트와 통합 브랜치 CI는 실행하지 않았다.
- `~/.moai/cache/search/<project-key>/sessions.db`는 경로 계약만 있고 생산자 전환은 확인하지 못했다.
- claimed 상태에서 프로세스가 중단된 resume를 lease 만료 후 재회수하는 정책은 구현하지 않았다.
- legacy `pending.json`을 read-only로 읽은 stale record는 DB row id가 없어 CAS 만료하지 않는다.
- 활성 프로필 보호는 현재 프로세스의 `CLAUDE_CONFIG_DIR`과 일치하는 프로필에 한정된다.
  별도 프로세스의 활성 프로젝트를 판별할 공용 lease/registry는 없다.
- `docs-site/README.md`의 Rank 섹션 대량 삭제는 이전 작업자의 변경이다. Rank 폐기와
  연관은 있으나 이 실행에서 문서 빌드·소유권·다국어 링크 무결성을 확인하지 않았다.
- handoff 테스트 일부가 SQLite 전환 과정에서 크게 축소됐다. 대체된 DB 동시성 검사는
  통과했지만 삭제된 모든 파일 오류 주입 시나리오와 의미상 동등하다고 주장하지 않는다.

## Residual-risk

- 구/신 바이너리를 동시에 실행하면 구 경로와 새 경로가 다시 갈라질 수 있으므로 실제
  전환 전에 활성 세션 정지와 단일 버전 재기동 절차가 필요하다.
- source→target 이전은 두 lock으로 직렬화하지만, 여러 legacy 후보가 동시에 존재할 때는
  project-local 후보를 우선한다. 서로 다른 카드가 양쪽에 존재하는 병합 정책은 없다.
- 실제 Claude lead/Codex lane 다중 프로세스 통신은 이번 범위 밖이며 검증하지 않았다.
- Factory DB 기록 실패는 CLI 본래 동작을 보존하기 위해 fail-open이다. 이벤트 누락은 가능하며
  후속 진단/재시도 큐는 이번 최소 구현에 포함하지 않았다.
