# SPEC-AGENTIC-FACTORY-LOCAL-001 — 수용 기준 (0.6.0 — revised from landed code, unaudited, no re-audit)

이전 문구(머리, 0.5.1): «# SPEC-AGENTIC-FACTORY-LOCAL-001 — 수용 기준 (0.5.1 — audited-FAIL 0.74 (plan debt accepted by leader decision), no re-audit)»

> 0.6.0: AC-AFL-001·007·008·009·014에 "0.6.0 사례" 줄을 더했다(spec HISTORY 0.6.0, progress §O). 각 줄의 "착지 시험"은 이 작성자가 HEAD `12eaa6c23f`에서 `grep -n '^func Test'`로 찾은 이름과 줄이다. 그 시험의 본문이 사례를 얼마나 덮는지는 판정하지 않았고(본문을 읽은 곳은 그 줄에 따로 적음), 시험은 하나도 돌리지 않았다 — 통과로 적지 않는다. 다섯 AC 모두 §N(release 게이트 밖)이며, 0.6.0 사례는 §D 완료 계산에 들지 않는다. AC마다 시험 이름이 하나라는 §A 규칙은 그대로다: 0.6.0 줄의 이름들은 AC의 시험 이름을 바꾸지 않는다.

이전 문구(머리, 0.5.0): «# SPEC-AGENTIC-FACTORY-LOCAL-001 — 수용 기준 (0.5.0 — audited-FAIL 0.70 → revised (A-na delta), unaudited)»

이전 문구(머리): «# SPEC-AGENTIC-FACTORY-LOCAL-001 — 수용 기준 (0.4.0)»

## §A 측정 규칙

- AC마다 검증 대상 시험 이름(Go 시험 함수 이름)은 **정확히 하나**다. 이름은 이 작성자가 계약의 패키지·소유 lane에 맞춰 붙인 제안 이름이다.
- 21개 이름은 모두 기준 커밋 `37b67f6d39`에 없다(§B — 이 부재는 at 37b67f6d39 한정이다; integration tip `db02401ad1`에는 그 가운데 셋이 있다, §B.2). 그래서 **실행 명령을 두지 않는다**: 아무 시험도 고르지 않는 `-run` 명령은 0개를 훑고도 `ok`를 낸다(`.claude/rules/moai/development/verification-completeness.md` §1.1).
- 각 AC의 두 칸(`verification-completeness.md` §2): **RED-now**(지금 무엇이 왜 빨간가)와 **green path**(어느 lane의 어떤 커밋이 뒤집는가). 둘 다 글로 적은 것이며 실행 결과가 아니다. (0.5.0) 두 칸은 release 게이트에 남은 AC 셋(013·015·017)에만 건다; 나머지 18개는 §N 사후 관측이다.
- (0.5.0) release 게이트와 사후 관측의 분리: 이 SPEC은 코드보다 늦게 쓰였다. integration tip `db02401ad1`(`git branch --list WT-agentic-factory-integration --format='%(refname:short) %(objectname:short)'` → `WT-agentic-factory-integration db02401ad1`; 감사가 인용한 `c83013938a`, 0.4.0이 읽은 `1a552af405`에서 움직임)에는 원장·병합 큐·흐름·보드·훅 heartbeat·importer·퇴역 코드가 이미 있다. 코드나 시험이 tip에 착지한 AC 18개(AC-AFL-001~012·014·016·018~021)는 release 게이트에서 빼서 §N '사후 관측 (release 게이트 밖)'에 둔다. 그 18개는 제목과 시험 이름 하나를 그대로 두며, 완료 계산에 들지 않는다(§D). 착지 시험이 무엇을 단언하는지는 §F·§N에 적었고, 어느 것도 RED를 관측했다고 적지 않는다. release 게이트에 남는 AC는 셋이다: AC-AFL-013·017은 시험이 쓰이지 않은 상태(이름이 `37b67f6d39`에도 tip `db02401ad1`에도 없다, §B·§B.2)이고, AC-AFL-015는 그 시험이 단언할 상태가 `37b67f6d39`에서 거짓임을 트리를 읽어 적었다(퇴역 대상 디렉터리 열 개가 그 트리에 있다, §B.2 [L-015a]). 이 SPEC은 어떤 RED도 관측했다고 적지 않는다.
- 이전 문구(0.4.0, §A 측정 시점 문제): «이 SPEC은 코드보다 늦게 쓰였다. integration tip `1a552af405`(조정자가 적은 `915a6524cb`에서 움직임)에는 원장·병합 큐·흐름·보드·훅 heartbeat·importer·퇴역이 이미 착지해 있다. 착지한 코드에 대해서는 RED 커밋이 GREEN보다 먼저였다는 증거를 이 SPEC의 이름으로 되살릴 수 없다. 그래서 코드가 착지한 AC(013·017을 뺀 19개)는 `verification-completeness.md` §2.1 undecidable disposition을 따른다: release-blocking RED가 아니라 regression-guard이고, 통과로 기록하지 않는다. 각 AC가 착지 시험으로 얼마나 덮이는지는 §F에 적었다. 착지 시험이 AC를 완전히 덮는 경우에만 그 이름을 AC의 이름으로 채택했다(AC-AFL-021 하나; 이전 이름은 '이전 문구'). AC-AFL-002·005는 이름이 이미 같고, 그 시험은 lane-18 커밋 `607bd17804`에 있으나 integration tip에는 아직 없다.»
- 이전 문구(0.3.0, RED-now 공통 형식): «RED-now: `<이름>`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.» 각 AC의 0.3.0 RED-now·green path는 그 AC의 '이전 문구(0.3.0, RED-now·green path)' 줄에 남겼다.
- RED-now의 이유는 21개 모두 "시험 함수가 `37b67f6d39`에 없다"이다. 그 시험이 놓일 패키지도 대부분 없다(`ledger`·`mergeq`·`monitor/flow`, spec §B). 이 빨강은 이 작업이 만들 시험이 없다는 뜻이며, 손대지 않는 파일 때문에 생긴 빨강이 아니다.

## §B 증거 원장 — 시험 이름 부재 확인

```text
명령: git grep -n -e 'func TestLedgerImportGraph_NoIssueOrRemoteClient' -e 'func TestStore_StaleEpochWrite_ReturnsErrStaleEpoch' -e 'func TestOpen_SetsBusyTimeoutAndChecksSQLiteVersion' -e 'func TestLedgerSource_TimestampColumnsUseDBClock' -e 'func TestReap_LivenessCases_ReclaimOnlyDead' -e 'func TestMove_RefusesEdgeRejectedByCheckMove' -e 'func TestReap_ThirdReclaim_ParksCard' -e 'func TestEnqueue_EvidenceSubjectSHAMismatch_ReturnsErrEvidenceStale' -e 'func TestSchema_CheckConstraints_RejectUnknownQueueStateAndNoteKind' -e 'func TestClaim_SingleTransaction_ReadyToBuildingByClass' -e 'func TestExpandKeys_Conflict_ReturnsErrKeyConflict' -e 'func TestClaim_DisjointConflictKeys_ClaimInParallel' -e 'func TestCardFlow_ClaimToDone_ThroughMergeQueue' -e 'func TestMergeq_Land_CASUpdateRefThenLandedNote' -e 'func TestRetirement_NoImportOfForgeMirrorGitref' -e 'func TestImporter_DryRun_SourcesUnchanged' -e 'func TestAFCommand_VerbsRegistered' -e 'func TestLedgerDefaults_MatchContract' -e 'func TestFlow_StalledDetection_FromReader' -e 'func TestBoard_ReaderOnly_NoStoreReference' -e 'func TestHeartbeatWriter_SkipsWriteWhenStampFresh' 37b67f6d39 -- internal
출력: (없음)
종료 코드: 1
tree: 37b67f6d39
측정: 0.2.0 작성 중 이 워크트리(HEAD c445ce9559)에서 다시 돌림. 이름 넷(AC-AFL-004·005·008·020)은 0.2.0에서 바뀐 이름이다.
대조(같은 형식이 찾을 수 있음을 보임): git grep -n -e 'func TestLiveOwnerExpiredLockStaysBusy' 37b67f6d39 -- internal
대조 출력: 37b67f6d39:internal/agenticfactory/claim/lock_test.go:63:func TestLiveOwnerExpiredLockStaysBusy(t *testing.T) {
대조 종료 코드: 0
```

이름마다 따로 낸 명령이 아니라 `-e` 21개를 묶은 명령 한 번이다. 묶음 명령은 이름 가운데 하나라도 있으면 그 줄을 출력하므로, 출력 없음은 21개 모두 없다는 뜻이다.

### §B.1 0.4.0 다시 잰 것 (AC-AFL-021 이름 채택 뒤의 이름 집합)

```text
명령(기준): git -C <worktree> grep -n -e 'func TestLedgerImportGraph_NoIssueOrRemoteClient' … -e 'func TestBoard_ReaderOnly_NoStoreReference' -e 'func TestLedgerHeartbeat_ThrottlesToTheStampFile' 37b67f6d39 -- internal
  (-e 21개: 위 §B 명령의 20개 이름 그대로 + AC-AFL-021의 채택 이름. 전체 문자열은 progress §K)
출력: (없음)
종료 코드: 관측하지 못함 — 이 실행 환경은 -e가 둘 이상인 git grep의 종료 코드를 보이지 않았다(지어낸 이름 둘만 넣은 대조도 같음; -e 하나의 대조는 'Exit code 1'을 보였다). 판정은 출력 없음과 아래 양성 대조에 기댄다.
대조: git -C <worktree> grep -n -e 'func TestLiveOwnerExpiredLockStaysBusy' 37b67f6d39 -- internal
대조 출력: 37b67f6d39:internal/agenticfactory/claim/lock_test.go:63:func TestLiveOwnerExpiredLockStaysBusy(t *testing.T) {
같은 21개 이름(+ 옛 이름 TestHeartbeatWriter_SkipsWriteWhenStampFresh) @ 1a552af405 출력:
  1a552af405:internal/hook/ledger_heartbeat_test.go:58:func TestLedgerHeartbeat_ThrottlesToTheStampFile(t *testing.T) {
같은 21개 이름 @ 607bd17804 출력:
  607bd17804:internal/agenticfactory/ledger/ac_names_test.go:30:func TestReap_LivenessCases_ReclaimOnlyDead(t *testing.T) {
  607bd17804:internal/agenticfactory/ledger/ac_names_test.go:117:func TestStore_StaleEpochWrite_ReturnsErrStaleEpoch(t *testing.T) {
측정: 0.4.0 작성 중(HEAD c83fdc04ff 위 작업 트리). 607bd17804는 가지 WT-agentic-factory-core-claim-runner에만 있다(git branch --contains 607bd17804 --list WT-agentic-factory-integration WT-agentic-factory-core-claim-runner → '+ WT-agentic-factory-core-claim-runner').
```

### §B.2 0.5.0 다시 잰 것 (integration tip `db02401ad1`, D6)

```text
[L-names] 명령: git -C <worktree> grep -n -e 'func TestLedgerImportGraph_NoIssueOrRemoteClient' … -e 'func TestLedgerHeartbeat_ThrottlesToTheStampFile' -e 'func TestHeartbeatWriter_SkipsWriteWhenStampFresh' db02401ad1 -- internal
  (-e 22개: §B.1의 21개 이름 + 옛 AC-AFL-021 이름. 전체 문자열은 progress §L)
출력:
  db02401ad1:internal/agenticfactory/ledger/ac_names_test.go:30:func TestReap_LivenessCases_ReclaimOnlyDead(t *testing.T) {
  db02401ad1:internal/agenticfactory/ledger/ac_names_test.go:117:func TestStore_StaleEpochWrite_ReturnsErrStaleEpoch(t *testing.T) {
  db02401ad1:internal/hook/ledger_heartbeat_test.go:58:func TestLedgerHeartbeat_ThrottlesToTheStampFile(t *testing.T) {
종료 코드: 관측하지 못함(도구가 따로 보이지 않음). 판정은 출력에 기댄다.
대조: git -C <worktree> grep -n -e 'func TestLiveOwnerExpiredLockStaysBusy' db02401ad1 -- internal
대조 출력: db02401ad1:internal/agenticfactory/claim/lock_test.go:63:func TestLiveOwnerExpiredLockStaysBusy(t *testing.T) {
읽은 것: tip에 있는 이름은 셋(AC-AFL-005·002·021의 이름)이고 나머지 18개와 옛 이름은 없다. lane-18 커밋 607bd17804의 ac_names_test.go는 tip에 들어왔다(위 출력의 경로).

[L-013] 명령: git -C <worktree> grep -n 'func TestCardFlow_ClaimToDone_ThroughMergeQueue' db02401ad1 -- internal
출력: (없음)
종료 코드: 관측하지 못함(이 실행에서 도구가 0이 아닌 종료 코드를 표시하지 않았다)
tree: db02401ad1

[L-017] 명령: git -C <worktree> grep -n 'func TestAFCommand_VerbsRegistered' db02401ad1 -- internal
출력: (없음)
종료 코드: 관측하지 못함
tree: db02401ad1

[L-015a] 명령: git -C <worktree> ls-tree -d --name-only 37b67f6d39 internal/agenticfactory/forge internal/agenticfactory/mirror internal/agenticfactory/gitref internal/agenticfactory/resident internal/agenticfactory/effect/ghintake internal/agenticfactory/effect/ctxledger internal/agenticfactory/effect/labels internal/agenticfactory/effect/ghexec internal/agenticfactory/effect/ghidentity internal/agenticfactory/effect/gitexec
출력:
  internal/agenticfactory/effect/ctxledger
  internal/agenticfactory/effect/ghexec
  internal/agenticfactory/effect/ghidentity
  internal/agenticfactory/effect/ghintake
  internal/agenticfactory/effect/gitexec
  internal/agenticfactory/effect/labels
  internal/agenticfactory/forge
  internal/agenticfactory/gitref
  internal/agenticfactory/mirror
  internal/agenticfactory/resident
종료 코드: 관측하지 못함(출력 있음)
tree: 37b67f6d39

[L-015b] 명령: 위 [L-015a]와 같고 트리만 db02401ad1
출력: (없음)
종료 코드: 관측하지 못함
tree: db02401ad1

[L-015c] 명령: git -C <worktree> grep -n -e 'factory_resident' -e 'newFactoryResidentCmd' 37b67f6d39 -- internal/cli
출력:
  37b67f6d39:internal/cli/factory_resident.go:3:// factory_resident.go — `moai factory resident`: runs the Agentic Factory resident monitor loop
  37b67f6d39:internal/cli/factory_resident_test.go:3:// factory_resident_test.go — `moai factory resident` (Issue #1948).
종료 코드: 관측하지 못함(출력 있음)
tree: 37b67f6d39

[L-015d] 명령: git -C <worktree> grep -n 'func TestRetirement_NoImportOfForgeMirrorGitref' db02401ad1 -- internal
출력: (없음)
종료 코드: 관측하지 못함
tree: db02401ad1
```

`verification-completeness.md` §2.1의 네 요소 가운데 종료 코드는 위 어느 항목에서도 관측하지 못했다. 그래서 이 원장을 인용하는 RED-now 칸은 §2.1 기준으로 완전하지 않다(§E 0.5.0).

## §C 수용 기준

#### AC-AFL-001 — 단일 로컬 권위, Issue 읽기·쓰기 없음

- 주 REQ: maps REQ-AFL-001.
- Given 원장 패키지와 그것이 가져오는 패키지 그래프, When 그래프를 훑을 때, Then `forge`·`mirror`·`gitref`·`effect/ghintake`·`effect/ctxledger`·`effect/labels` 가운데 어느 것도 그래프에 없고, 원장 파일 경로는 `<moaiHome>/db/<projectKey>/af/ledger.db` 하나다.
- 같은 AC의 0.6.0 사례(REQ-AFL-001 저장소 묶음): (e1) Given 저장소 R1에 묶인 원장, When `BindRepository(R2)`를 부를 때, Then `ErrWrongRepository`이고 묶음은 R1 그대로이며, 같은 R1을 다시 부르면 변화가 없고, `repository` 행의 UPDATE·DELETE는 트리거가 거절한다; (e2) Given 원장이 R1에 묶인 상태, When 다른 저장소에서(프로젝트 디렉터리·lease worktree·`--repo`로) 어떤 `moai af` 동사를 부를 때, Then 원장이 바뀌지 않고 거절된다; (e3) Given 아직 묶이지 않은 원장의 lease worktree가, 또는 `--db`로 준 원장 파일 자리가(symlink를 푼 자리), 또는 행이 이름 대는 커밋을 가진 저장소가 R1인 상태, When R2에서 동사를 부를 때, Then 둘을 이름 대어 거절하고 원장은 그대로다(기본 위치의 원장 파일 자리는 단서가 아니다); (e4) Given 프로젝트 저장소의 linked worktree, When 그 경로로 동사를 부를 때, Then 받아들여진다(같은 git common dir). 죽여야 할 변이: 두 번째 묶음이 첫 묶음을 덮어쓰는 구현 → (e1)에서 FAIL; 묶이지 않은 원장에서 아는 저장소와 비교하지 않는 구현 → (e3)에서 FAIL; 경로 문자열만 비교하는 `--repo` 검사 → (e4)에서 FAIL.
- 같은 AC의 2026-11-02 개정 사례(REQ-AFL-001 첫 묶음, 새 id 없음): (e5) Given 아직 묶이지 않은 원장 — 기본 위치의 원장, 또는 프로젝트 저장소 R1 안에 놓인 `--db` 원장 — 과 프로젝트 디렉터리 R1, When 원장에 아무것도 쓰지 않고 끝나는 동사(회수할 것이 없는 `reap`)를 `--bind-repo` 없이 부를 때, Then 동사는 끝나고 원장은 묶이지 않은 채다(`Repository` = ""); (e5') Given 같은 원장, When 같은 동사를 `--bind-repo <R1 안의 경로>`로 부를 때, Then 원장은 R1의 git common dir에 묶인다; (e5'') Given 같은 원장, When 같은 동사를 다른 저장소 R2 안의 경로나 git 저장소가 아닌 경로를 가리키는 `--bind-repo`로 부를 때, Then "nothing was changed"로 거절되고 원장은 묶이지 않은 채다. 죽여야 할 변이: 아무것도 쓰지 않은 동사도 묶는 구현(개정 전 계약) → (e5)에서 FAIL; 밝힘을 무시하고 묶지 않는 구현 → (e5')에서 FAIL; 어긋난 밝힘을 받아들이거나 묶는 구현 → (e5'')에서 FAIL. 이 세 단계는 한 입력에서 개정 전 계약과 함께 성립할 수 없다 — 개정은 개정 전 계약(R-B)을 바꾼다.
  - (e5)의 RED 기록(이 작성자가 돌리지 않음; 수리 보고 `.moai/reports/issue-1948/af-audit-a36e-repair.md` N55 절, `1c8ee551e2`에서 되살린 사례; 명령은 그 절 "Step 1"에 원문 그대로 있다): `TestAFResidual_N55ARefusedVerbThatWroteNothingDoesNotBindTheLedger`의 하위 사례 `reap_with_nothing_to_reap`이 `--- FAIL`, `bound=".../O/.git"`. 그 뒤 `31a7c1099a`까지 코드 변경 없음(`git diff --stat 709d9fd9d4 31a7c1099a`가 보고서 한 파일만 찍음). (e5')(e5'')를 단언하는 시험 이름은 확인하지 않았다. 아래 착지 시험 가운데 개정 전 계약을 고정하는 것(밝힘 없는 빈손 `reap` 뒤 `bound == <git 디렉터리>`; 같은 보고서 N55 절): `TestAFVerbs_N38FirstBindLedgerFileInsideAnotherRepositoryIsRefused`, `TestAFVerbs_N50TheDefaultLedgerLocationIsNoFileClue`, `TestAFVerbs_N51bRowsNameCommitsTheCallerDoesNotHave`, `TestAFVerbs_N51cASymlinkedLedgerFileIsAskedWhereItReallyIs`, `TestAFVerbs_N51ADeclarationIsNotAskedWhereAnythingElseSaysWhereTheLedgerBelongs`(`the_ledger_file_sits_in_the_repository`, `the_default_location`). 그 시험들과 코드를 개정 계약에 맞추는 일은 후속 카드 몫이며 이 개정은 바꾸지 않았다.
  - 이 사례를 다루는 착지 시험(이름·줄만 확인, 본문 덮음 판정 안 함, 돌리지 않음): `internal/agenticfactory/ledger/repository_identity_test.go:30` `TestRepository_TheFirstBindingIsKeptAndSurvivesAReopen`, `:55` `TestRepository_TheSameIdentityAgainIsANoOp`, `:76` `TestRepository_ADifferentIdentityIsRefusedAndNothingChanges`, `:105` `TestRepository_TwoFirstWritesLeaveExactlyOneWinner`, `:135` `TestRepository_TheBindingCannotBeUpdatedOrDeletedByAnyoneHoldingTheFile`; `internal/cli/af_repo_n38_test.go:133` `TestAFVerbs_N38TheFirstVerbBindsTheLedgerToTheProjectsGitDirectory`, `:177` `TestAFVerbs_N38EveryVerbIsRefusedWhenTheProjectDirectoryNamesAnotherRepository`, `:253` `TestAFVerbs_N38ALeaseWorktreeInAnotherRepositoryIsRefusedWhereItIsUsed`, `:277` `TestAFVerbs_N38TheSameRepositoryByAnotherRouteIsAccepted`, `:321` `TestAFVerbs_N38AVerbThatRunsNoGitCarriesOnInADirectoryThatIsNoRepository`; `internal/cli/af_repo_n38_firstbind_test.go:216` `TestAFVerbs_N38FirstBindLeaseWorktreeNamesAnotherRepositoryRefusesEveryVerb`, `:247` `TestAFVerbs_N38FirstBindLedgerFileInsideAnotherRepositoryIsRefused`, `:311` `TestAFVerbs_N38FirstBindWithNothingKnownBindsAsBefore`, `:357` `TestAFVerbs_N38FirstBindLinkedWorktreeOfTheProjectPasses`; `internal/cli/af_land_repair_test.go:222` `TestAFVerbs_ARepoThatIsNotTheProjectsRepositoryIsRefusedBeforeAnyChange`. 훅 heartbeat 기록기의 저장소 확인: no test found(그 기록기에 저장소 확인이 없다 — progress §O 열린 위험).
- 시험 이름(하나): `TestLedgerImportGraph_NoIssueOrRemoteClient` — 소유 lane-18, 패키지 `internal/agenticfactory/ledger`. unwritten at 37b67f6d39(§B).
- RED-now·green path(0.5.0): 이 AC는 release 게이트 밖이다 — §N 사후 관측 행 AC-AFL-001. RED 미관측. 완료 계산에 들지 않는다(§D). 시험 이름은 위 한 개 그대로다.
- 이전 문구(0.4.0, RED-now·green path): «name absent at 37b67f6d39 — `TestLedgerImportGraph_NoIssueOrRemoteClient`가 `37b67f6d39`에 없다(§B). integration tip `1a552af405`(조정자가 적은 `915a6524cb`에서 움직인 tip)에서 이 이름은 없다. 착지 시험이 이 AC를 덮는 범위: 착지 시험 없음(not)(§F 행 AC-AFL-001). 착지 전 RED는 관측되지 않았고 이 칸은 판정한 RED가 아니다 — 통과로 적지 않는다.» «RED evidence not recoverable from history (landed before this SPEC); this AC adopts the landed test as a regression-guard, not a release-blocking RED (`verification-completeness.md` §2.1 undecidable disposition). 남은 일: §F에 적은 덮이지 않은 부분을 이 이름의 시험(소유 lane-18)이 더한다. 그 시험은 이미 착지한 코드 위에서 쓰이므로 처음부터 초록일 수 있고, RED가 아니라 regression-guard로 다룬다. 참고 사실(시험 아님): `1a552af405`에서 퇴역 대상 9개 패키지를 가져오는 import를 찾는 `git grep`은 출력이 없었다(progress §K).»
- 이전 문구(0.3.0, RED-now·green path): «`TestLedgerImportGraph_NoIssueOrRemoteClient`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.» «lane-18의 원장 첫 구현 커밋(계약 §4: `ledger/api.go` 형식 전용 커밋 뒤의 Open/schema 커밋)이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.»

#### AC-AFL-002 — fencing epoch와 낡은 epoch 쓰기

- 주 REQ: maps REQ-AFL-002.
- Given 카드 하나가 epoch 1로 claim되고, 그 소유자가 회수(Dead)된 뒤 다른 세션이 다시 claim해 epoch 2가 된 상태(새 소유자의 lease와 키가 있음), When epoch 1을 실은 아래 여섯 메서드를 각각 부를 때, Then 사례마다 아래 표와 같다(한 시험 안의 하위 사례 여섯). 모든 사례에서 바뀐 행은 0이다.
- 전제(0.5.0): 이 Given은 회수된 카드를 `Claim`이 다시 고르는 것을 전제한다. 회수된 카드가 다시 claim 대상이 되는 집합은 REQ-AFL-010에서 미정이다(OD-43). 코드 사실(요구사항 아님): `db02401ad1:internal/agenticfactory/ledger/claim.go:351-358`의 `pickTx`는 lease 없는 BUILDING·REVIEWING 카드도 고른다. 이전 문구(0.4.0): 이 전제 줄은 없었다.

| 사례 | 메서드(epoch 1) | 오류 | 바뀌지 않아야 하는 것 |
|---|---|---|---|
| S-Move | `Move` | `ErrStaleEpoch` | `cards`의 단계·epoch, 그 카드의 `transitions` 행 수 |
| S-ExpandKeys | `ExpandKeys` | `ErrStaleEpoch` | 그 카드의 `lease_keys` 행 |
| S-AddEvidence | `AddEvidence` | `ErrStaleEpoch` | 그 카드의 `evidence` 행 수 |
| S-Heartbeat | `Heartbeat` | `ErrStaleEpoch` | 새 소유자 lease의 `heartbeat_at`·`expires_at` |
| S-Release | `Release` | `ErrStaleEpoch` | 새 소유자(epoch 2)의 `leases` 행과 `lease_keys` 행 — 낡은 epoch의 Release가 새 소유자의 lease·키를 지우지 않는다 |
| S-Enqueue | `Enqueue` | `ErrStaleEpoch` | 그 카드의 `merge_queue` 행(없음 그대로)과 단계 |

- 죽여야 할 변이: `Release`가 `WHERE card_id=?`만으로 lease·키를 지우는 구현(epoch를 보지 않음) → S-Release에서 FAIL(새 소유자의 lease·키가 사라짐). 그 밖에 어느 쓰기든 WHERE에 epoch를 넣지 않는 구현 → 그 사례에서 FAIL.
- 범위 밖: `QueueUpdate`의 landing 잠금 epoch 울타리는 AC-AFL-014 사례 (b)가 맡는다(N14).
- 이전 문구(0.3.0): Given 카드 하나가 epoch 1로 claim된 뒤 다시 claim되어 epoch 2가 된 상태, When epoch 1을 실은 `Move`·`AddEvidence`·`Heartbeat`를 부를 때, Then 바뀐 행이 0이고 오류는 `ErrStaleEpoch`이며, 카드 상태는 그대로다.
- 시험 이름(하나): `TestStore_StaleEpochWrite_ReturnsErrStaleEpoch` — 소유 lane-18, 패키지 `internal/agenticfactory/ledger`. unwritten at 37b67f6d39(§B).
- RED-now·green path(0.5.0): 이 AC는 release 게이트 밖이다 — §N 사후 관측 행 AC-AFL-002. RED 미관측. 완료 계산에 들지 않는다(§D). 시험 이름은 위 한 개 그대로다.
- 이전 문구(0.4.0, RED-now·green path): «name absent at 37b67f6d39 — `TestStore_StaleEpochWrite_ReturnsErrStaleEpoch`가 `37b67f6d39`에 없다(§B). integration tip `1a552af405`(조정자가 적은 `915a6524cb`에서 움직인 tip)에서 이 이름은 없다. lane-18 커밋 `607bd17804`(가지 `WT-agentic-factory-core-claim-runner`; integration tip `1a552af405`에 들지 않음)의 `internal/agenticfactory/ledger/ac_names_test.go:117`에 같은 이름이 있다. 착지 시험이 이 AC를 덮는 범위: partly — 같은 이름 시험은 `Move`·`AddEvidence`·`Heartbeat` 세 메서드를 덮고, `ExpandKeys`·`Release`·`Enqueue`는 다른 착지 시험이 덮는다(§F 행 AC-AFL-002). 착지 전 RED는 관측되지 않았고 이 칸은 판정한 RED가 아니다 — 통과로 적지 않는다.» «RED evidence not recoverable from history (landed before this SPEC); this AC adopts the landed test as a regression-guard, not a release-blocking RED (`verification-completeness.md` §2.1 undecidable disposition). 같은 이름 시험이 integration에 들어오면 regression-guard로 돈다. 이름은 바꾸지 않는다(다른 메서드는 §F의 추가 시험이 덮어 이 이름 하나가 AC 전체를 덮지 않는다). 남은 일: §F에 적은 덮이지 않은 부분을 이 이름의 시험(소유 lane-18)이 더한다. 그 시험은 이미 착지한 코드 위에서 쓰이므로 처음부터 초록일 수 있고, RED가 아니라 regression-guard로 다룬다.»
- 이전 문구(0.3.0, RED-now·green path): «`TestStore_StaleEpochWrite_ReturnsErrStaleEpoch`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.» «lane-18의 Claim/Move 구현 커밋이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.»

#### AC-AFL-003 — BEGIN IMMEDIATE·busy_timeout·SQLite 버전

- 주 REQ: maps REQ-AFL-003.
- Given 새 원장 파일, When `Open`이 연결을 열 때, Then busy_timeout이 5000 ms로 설정되어 있고, SQLite 버전이 3.35보다 낮다고 보고되면 `Open`이 오류를 돌려주며 store를 내주지 않는다.
- 시험 이름(하나): `TestOpen_SetsBusyTimeoutAndChecksSQLiteVersion` — 소유 lane-18, 패키지 `internal/agenticfactory/ledger`. unwritten at 37b67f6d39(§B).
- RED-now·green path(0.5.0): 이 AC는 release 게이트 밖이다 — §N 사후 관측 행 AC-AFL-003. RED 미관측. 완료 계산에 들지 않는다(§D). 시험 이름은 위 한 개 그대로다.
- 이전 문구(0.4.0, RED-now·green path): «name absent at 37b67f6d39 — `TestOpen_SetsBusyTimeoutAndChecksSQLiteVersion`가 `37b67f6d39`에 없다(§B). integration tip `1a552af405`(조정자가 적은 `915a6524cb`에서 움직인 tip)에서 이 이름은 없다. 착지 시험이 이 AC를 덮는 범위: partly — busy_timeout과 버전 판정 함수는 덮고, 낮은 버전에서 `Open`이 store를 내주지 않는 것은 덮지 않는다(§F 행 AC-AFL-003). 착지 전 RED는 관측되지 않았고 이 칸은 판정한 RED가 아니다 — 통과로 적지 않는다.» «RED evidence not recoverable from history (landed before this SPEC); this AC adopts the landed test as a regression-guard, not a release-blocking RED (`verification-completeness.md` §2.1 undecidable disposition). 남은 일: §F에 적은 덮이지 않은 부분을 이 이름의 시험(소유 lane-18)이 더한다. 그 시험은 이미 착지한 코드 위에서 쓰이므로 처음부터 초록일 수 있고, RED가 아니라 regression-guard로 다룬다.»
- 이전 문구(0.3.0, RED-now·green path): «`TestOpen_SetsBusyTimeoutAndChecksSQLiteVersion`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.» «lane-18의 Open 구현 커밋이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.»

#### AC-AFL-004 — 사건 시각은 DB 시계

- 주 REQ: maps REQ-AFL-004.
- Given `internal/agenticfactory/ledger` 디렉터리 바로 아래의 `_test.go`가 아닌 `.go` 파일 전부(하위 디렉터리 `importer`는 넣지 않음), When 시험이 그 파일들을 `go/parser`로 읽고 (i) `INSERT`·`UPDATE` 문 문자열을 모을 때 문자열 리터럴과 **패키지 수준 문자열 상수**를 `+`로 이은 식을 `go/constant`(또는 같은 일을 하는 `go/types`의 상수 값)로 접어 하나의 SQL 문자열로 만든 뒤, 다섯 시각 열 `created_at`·`updated_at`·`at`·`heartbeat_at`·`enqueued_at` 가운데 하나에 값을 넣는 자리를 찾고, (ii) `time.Now`·`time.Since`·`time.Until`·`time.After` 선택자 호출을 모을 때, Then (i) 접은 SQL에서 그 모든 자리의 값이 `strftime('%Y-%m-%dT%H:%M:%fZ','now')`이고(자리표시자 `?`, SQL 안의 Go 쪽 시각 문자열, 접을 수 없는 피연산자가 아님), (ii)는 아래 허용 예외 밖에서 0건이다 — 시각 열에 Go 시계 값을 쓰는 곳이 0이다.
- 접기 규칙: 접을 수 없는 피연산자(함수 호출 `offsetSQL(d)` 같은 것)가 이은 식에 있으면 시험은 그 피연산자가 값을 넣는 열을 본다 — 다섯 열 가운데 하나면 FAIL, 다섯 열 밖(`expires_at` 등)이면 세지 않는다.
- 착지 사실(준수 구현이 통과해야 함): `1a552af405:internal/agenticfactory/ledger/schema.go:24`에 ``const nowSQL = `strftime('%Y-%m-%dT%H:%M:%fZ','now')` ``가 있고, `cards.go:73`·`claim.go:154`·`evidence.go:37`·`notes.go:26`·`queue.go:124` 등이 `` `+nowSQL+` ``로 그 상수를 잇는다(`git grep -n 'nowSQL' 1a552af405 -- internal/agenticfactory/ledger`). 0.3.0 문구대로 리터럴만 맞추면 이 준수 구현이 (i)에서 FAIL했으므로 접기를 넣었다(N5).
- 시험이 읽는 것: 위 파일들의 Go 구문 트리(선택자 식, 문자열 리터럴, 패키지 수준 상수 선언). DB를 열지 않는다. 빈 집합 확인: 시험은 먼저 (i)에서 다섯 열 각각에 대해 모은 자리가 1개 이상임을 단언한다 — 아무것도 모으지 못하고 통과하는 것을 막는다(`verification-completeness.md` §1.1).
- 허용 예외(명시 목록, 이 밖은 없다): (e1) `Heartbeat`의 호출자 인자 `nextWake` — `next_wake_at`에 들어간다. 다섯 시각 열이 아니며, 계약 §10은 "nextWake is informational only"라고 적지만 I4와의 관계는 열린 예외로 둔다(OD-35 (f)). 다섯 열 밖인 `expires_at`(leases·locks)의 원천도 OD-35 (f)에 열려 있다. 30분 snapshot 주기의 타이머(`time.NewTicker`·`time.AfterFunc`)는 (ii)의 네 선택자와 이름이 달라 세지 않는다. 새 예외는 SPEC 개정으로만 이 목록에 들어간다.
- 죽여야 할 변이: (m1) Go 시각 값을 시각 열에 쓰는 구현 — `?`에 Go `time.Time`이나 그 문자열을 넘기거나, SQL 리터럴 안에 Go 쪽 시각 문자열을 넣음 → (i)에서 FAIL; (m2) `time.Now().Format(...)`을 SQL 문자열에 이어 붙이는 구현 → 접을 수 없는 피연산자가 다섯 열에 값을 넣어 (i)에서 FAIL, `time.Now` 호출로 (ii)에서도 FAIL.
- 시험 이름(하나): `TestLedgerSource_TimestampColumnsUseDBClock` — 소유 lane-18, 패키지 `internal/agenticfactory/ledger`. unwritten at 37b67f6d39(§B).
- RED-now·green path(0.5.0): 이 AC는 release 게이트 밖이다 — §N 사후 관측 행 AC-AFL-004. RED 미관측. 완료 계산에 들지 않는다(§D). 시험 이름은 위 한 개 그대로다.
- 이전 문구(0.4.0, RED-now·green path): «name absent at 37b67f6d39 — `TestLedgerSource_TimestampColumnsUseDBClock`가 `37b67f6d39`에 없다(§B). integration tip `1a552af405`(조정자가 적은 `915a6524cb`에서 움직인 tip)에서 이 이름은 없다. 착지 시험이 이 AC를 덮는 범위: not — 착지한 시각 확인은 런타임 값 범위만 보고 DB 시계와 Go 시계를 가르지 못한다; 정적 검사 시험은 없다(§F 행 AC-AFL-004). 착지 전 RED는 관측되지 않았고 이 칸은 판정한 RED가 아니다 — 통과로 적지 않는다.» «RED evidence not recoverable from history (landed before this SPEC); this AC adopts the landed test as a regression-guard, not a release-blocking RED (`verification-completeness.md` §2.1 undecidable disposition). 남은 일: §F에 적은 덮이지 않은 부분을 이 이름의 시험(소유 lane-18)이 더한다. 그 시험은 이미 착지한 코드 위에서 쓰이므로 처음부터 초록일 수 있고, RED가 아니라 regression-guard로 다룬다.»
- 이전 문구(0.3.0, RED-now·green path): «`TestLedgerSource_TimestampColumnsUseDBClock`가 `37b67f6d39`에 없다(§B). 원장 패키지도 없다. 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.» «lane-18의 schema·쓰기 구현 커밋이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.»
- 이전 문구(0.3.0, Given·읽는 것·예외·변이): «Given `internal/agenticfactory/ledger` 디렉터리 바로 아래의 `_test.go`가 아닌 `.go` 파일 전부(하위 디렉터리 `importer`는 넣지 않음), When 시험이 그 파일들을 `go/parser`로 읽어 (i) 문자열 리터럴 가운데 `INSERT`·`UPDATE` 문이 다섯 시각 열 `created_at`·`updated_at`·`at`·`heartbeat_at`·`enqueued_at` 가운데 하나에 값을 넣는 자리와 (ii) `time.Now`·`time.Since`·`time.Until`·`time.After` 선택자 호출을 모을 때, Then (i)의 모든 자리가 `strftime('%Y-%m-%dT%H:%M:%fZ','now')`로 값을 넣고(자리표시자 `?`나 Go 쪽 값이 아님), (ii)는 아래 허용 예외 밖에서 0건이다. / 시험이 읽는 것: 위 파일들의 Go 구문 트리(선택자 식과 문자열 리터럴). DB를 열지 않는다. 빈 집합 확인: 시험은 먼저 (i)에서 다섯 열 각각에 대해 모은 자리가 1개 이상임을 단언한다 — 아무것도 모으지 못하고 통과하는 것을 막는다(`verification-completeness.md` §1.1). / 허용 예외(명시 목록, 이 밖은 없다): (e1) `Heartbeat`의 호출자 인자 `nextWake` — `next_wake_at`에 들어간다. 다섯 시각 열이 아니며, 계약 §10은 "nextWake is informational only"라고 적지만 I4와의 관계는 열린 예외로 둔다(OD-35 (f)). 다섯 열 밖인 `expires_at`(leases·locks)의 원천도 OD-35 (f)에 열려 있다. 30분 snapshot 주기의 타이머(`time.NewTicker`·`time.AfterFunc`)는 (ii)의 네 선택자와 이름이 달라 세지 않는다. 새 예외는 SPEC 개정으로만 이 목록에 들어간다. / 죽여야 할 변이: Go `time.Now()` 값을 문자열로 만들어 `?`로 `at`에 넣는 구현 → (i)에서 FAIL; Go 시계를 다른 계산에 쓰는 구현 → (ii)에서 FAIL.»
- 이전 문구(0.1.0): Given 호출자 쪽 시계를 먼 과거로 돌린 시험 환경, When `AddCard`·`Move`·`Heartbeat`·`Enqueue`를 부를 때, Then `created_at`·`updated_at`·`at`·`heartbeat_at`·`enqueued_at`이 호출자 시각이 아니라 DB의 `strftime('%Y-%m-%dT%H:%M:%fZ','now')` 값이다. 시험 이름 `TestStore_Timestamps_UseDBClock`.

#### AC-AFL-005 — 살아 있는 소유자는 회수하지 않음

- 주 REQ: maps REQ-AFL-005.
- Given lease 기한이 지난 카드 셋(소유자가 서로 다름)과, 소유자마다 아래 표의 판정을 돌려주는 `alive` 함수, When `Reap`을 한 번 부를 때, Then 사례마다 아래 표와 같다(한 시험 안의 하위 사례 셋).

| 사례 | `alive` 판정 | lease·키 | `reclaim_count` | `stalled` 쪽지 |
|---|---|---|---|---|
| Live | 살아 있음(계약 §10의 `Alive`) | 그대로 | 그대로 | 그 card·epoch에 정확히 하나(계약 §10 20th (d); OD-41 leader decision (leader-decisions-od36-42-20261030.md, 2026-10-30T05:56Z; reversible)) |
| Unknown | 판정 불명 | 그대로(살아 있는 것으로 다룸) | 그대로 | 하나 |
| Dead | 죽음 | 없어짐(한 번 회수, `Reap` 반환값에 그 카드 하나) | +1 | 단언하지 않음 |

- 이전 문구(0.2.0, Live 행 쪽지 칸): «그 card·epoch에 하나(계약 §10 20th (d); 조정자 지시 "no `stalled` note"와 다름 — OD-41)». OD-41이 닫혀 그 어긋남 표시를 지웠다.
- 죽여야 할 변이: 기한 지난 Live 소유자를 회수하는 구현 → Live 행에서 FAIL(I5 "A live process is never reclaimed automatically"). 그 밖에 Unknown을 회수하는 구현은 Unknown 행, Dead를 회수하지 않는 구현은 Dead 행에서 FAIL.
- 시험 이름(하나): `TestReap_LivenessCases_ReclaimOnlyDead` — 소유 lane-18, 패키지 `internal/agenticfactory/ledger`. unwritten at 37b67f6d39(§B).
- RED-now·green path(0.5.0): 이 AC는 release 게이트 밖이다 — §N 사후 관측 행 AC-AFL-005. RED 미관측. 완료 계산에 들지 않는다(§D). 시험 이름은 위 한 개 그대로다.
- 이전 문구(0.4.0, RED-now·green path): «name absent at 37b67f6d39 — `TestReap_LivenessCases_ReclaimOnlyDead`가 `37b67f6d39`에 없다(§B). integration tip `1a552af405`(조정자가 적은 `915a6524cb`에서 움직인 tip)에서 이 이름은 없다. lane-18 커밋 `607bd17804`(가지 `WT-agentic-factory-core-claim-runner`; integration tip `1a552af405`에 들지 않음)의 `internal/agenticfactory/ledger/ac_names_test.go:30`에 같은 이름이 있다. 착지 시험이 이 AC를 덮는 범위: fully — `607bd17804`의 같은 이름 시험이 세 사례를 모두 단언한다(§F 행 AC-AFL-005). 착지 전 RED는 관측되지 않았고 이 칸은 판정한 RED가 아니다 — 통과로 적지 않는다.» «RED evidence not recoverable from history (landed before this SPEC); this AC adopts the landed test as a regression-guard, not a release-blocking RED (`verification-completeness.md` §2.1 undecidable disposition). 같은 이름 시험이 integration에 들어오면 regression-guard로 돈다. 이름은 바꾸지 않는다(이미 같은 이름이다).»
- 이전 문구(0.3.0, RED-now·green path): «`TestReap_LivenessCases_ReclaimOnlyDead`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.» «lane-18의 Reap 구현 커밋(활성 판정 어댑터는 lane-20)이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.»
- 이전 문구(0.1.0): Given lease 기한이 지난 카드와, 그 소유자에 대해 판정 불명을 돌려주는 `alive` 함수, When `Reap`을 부를 때, Then lease와 키가 그대로 남고 `reclaim_count`가 늘지 않으며 `stalled` 쪽지가 하나 기록된다. 시험 이름 `TestReap_UnknownLiveness_KeepsLeaseAndWritesStalledNote`.

#### AC-AFL-006 — 단계 쓰기는 core.CheckMove를 거침

- 주 REQ: maps REQ-AFL-006.
- Given READY 카드 하나, When `core.CheckMove`가 거절하는 위치(예: READY에서 DONE)로 `Move`를 부를 때, Then 오류가 돌아오고 `cards.stage`와 `transitions`가 바뀌지 않는다.
- 시험 이름(하나): `TestMove_RefusesEdgeRejectedByCheckMove` — 소유 lane-18, 패키지 `internal/agenticfactory/ledger`. unwritten at 37b67f6d39(§B).
- RED-now·green path(0.5.0): 이 AC는 release 게이트 밖이다 — §N 사후 관측 행 AC-AFL-006. RED 미관측. 완료 계산에 들지 않는다(§D). 시험 이름은 위 한 개 그대로다.
- 이전 문구(0.4.0, RED-now·green path): «name absent at 37b67f6d39 — `TestMove_RefusesEdgeRejectedByCheckMove`가 `37b67f6d39`에 없다(§B). integration tip `1a552af405`(조정자가 적은 `915a6524cb`에서 움직인 tip)에서 이 이름은 없다. 착지 시험이 이 AC를 덮는 범위: partly — 거절 오류는 덮고, 거절 뒤 `cards.stage`·`transitions` 불변은 덮지 않는다(§F 행 AC-AFL-006). 착지 전 RED는 관측되지 않았고 이 칸은 판정한 RED가 아니다 — 통과로 적지 않는다.» «RED evidence not recoverable from history (landed before this SPEC); this AC adopts the landed test as a regression-guard, not a release-blocking RED (`verification-completeness.md` §2.1 undecidable disposition). 남은 일: §F에 적은 덮이지 않은 부분을 이 이름의 시험(소유 lane-18)이 더한다. 그 시험은 이미 착지한 코드 위에서 쓰이므로 처음부터 초록일 수 있고, RED가 아니라 regression-guard로 다룬다.»
- 이전 문구(0.3.0, RED-now·green path): «`TestMove_RefusesEdgeRejectedByCheckMove`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.» «lane-18의 Move 구현 커밋(F6 강화 `37b67f6d39` 위)이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.»

#### AC-AFL-007 — 행을 지우지 않음·세 번째 회수에서 park

- 주 REQ: maps REQ-AFL-007.
- Given 이미 두 번 회수된 카드 X와 Dead를 돌려주는 `alive` 함수, 그리고 lease와 키를 쥔 카드 Y·Z, When (i) X의 세 번째 회수가 일어나고, (ii) Y를 DONE으로, Z를 CANCELLED로 `Move`할 때, Then (i) X의 `reclaim_count`가 3이고 X가 park 사유 `reclaim-cap`으로 park되며, 앞선 `transitions` 행이 하나도 지워지지 않았고, 회수된 소유자의 `leases`·`lease_keys` 행은 없다(`Reap`이 지운 것); (ii) Y·Z의 `leases`·`lease_keys` 행이 각 Move와 같은 트랜잭션 안에서 `Release` 경로로 지워졌고, Y·Z의 `cards`·`transitions` 행은 남아 있다(REQ-AFL-007, OD-36 leader decision (leader-decisions-od36-42-20261030.md, 2026-10-30T05:56Z; reversible)).
- 전제(0.5.0): (i)의 '두 번 회수된 카드'는 회수 뒤 다시 claim된 카드다. 회수된 카드가 다시 claim 대상이 되는 집합은 REQ-AFL-010에서 미정이다(OD-43); 착지 코드는 그런 카드를 다시 고른다(`db02401ad1:internal/agenticfactory/ledger/claim.go:351-358`, `claim_test.go:296`). 이전 문구(0.4.0): 이 전제 줄은 없었다.
- 같은 시험의 사례 (iii)(0.4.0, N6): Given `evidence`·`notes`·`merge_queue` 행을 가진 카드 W, When W의 lease가 (a) `Release`로 풀릴 때, (b) `Reap`(Dead)으로 회수될 때, (c) W가 DONE으로 `Move`될 때(각각 따로 준비한 W), Then 각 경우 뒤 W의 `evidence`·`notes`·`merge_queue` 행 수가 그 전과 같다.
- 죽여야 할 변이: DONE·CANCELLED로 옮긴 뒤 lease를 남기는 구현 → (ii)에서 FAIL; Move가 `cards`나 `transitions` 행을 지우는 구현 → (ii)에서 FAIL; 셋째 회수에서 park하지 않는 구현 → (i)에서 FAIL; `Release`·`Reap`·DONE `Move`가 `evidence`·`notes`·`merge_queue` 행을 지우는 구현 → (iii)에서 FAIL.
- 이전 문구(0.3.0, 변이 줄): «DONE·CANCELLED로 옮긴 뒤 lease를 남기는 구현 → (ii)에서 FAIL; Move가 `cards`나 `transitions` 행을 지우는 구현 → (ii)에서 FAIL; 셋째 회수에서 park하지 않는 구현 → (i)에서 FAIL.» 사례 (iii)은 없었다.
- 이전 문구(0.2.0): Given 이미 두 번 회수된 카드와 Dead를 돌려주는 `alive` 함수, When 세 번째 회수가 일어날 때, Then `reclaim_count`가 3이고 카드가 park 사유 `reclaim-cap`으로 park되며, 앞선 `transitions` 행이 하나도 지워지지 않았고, 회수된 소유자의 `leases`·`lease_keys` 행은 없다(`Reap`이 지운 것; REQ-AFL-007, OD-36).
- 이전 문구(0.1.0): Given 이미 두 번 회수된 카드와 Dead를 돌려주는 `alive` 함수, When 세 번째 회수가 일어날 때, Then `reclaim_count`가 3이고 카드가 park되며, 앞선 `transitions` 행이 하나도 지워지지 않았다.
- 같은 AC의 0.6.0 사례(REQ-AFL-007 reclaim-cap park의 출구): (iv) Given 사유 `reclaim-cap`으로 park된 카드 R(`reclaim_count` 3), When 운영자가 `moai af resume`을 `--guard reclaim-cap`, 비지 않은 `--reason`, R의 현재 epoch로 부를 때, Then R은 같은 단계에서 park가 풀리고 `reclaim_count`가 0이며 그 이동의 증거에 `resume-guard:reclaim-cap`이 실린다; (v) Given 같은 R, When `--reason`이 비었거나, `--guard`가 `reclaim-cap`이 아니거나, epoch가 다를 때, 그리고 Given park되지 않은 카드, When `resume`을 부를 때, Then 모두 거절되고 원장이 그대로다. 죽여야 할 변이: park를 풀며 `reclaim_count`를 남기는 구현 → (iv)에서 FAIL(다음 회수 한 번에 다시 park); 사유 없이 풀어 주는 구현 → (v)에서 FAIL.
  - 착지 시험(이름·줄만 확인, 돌리지 않음): `internal/agenticfactory/ledger/move_test.go:178` `TestAnOperatorResumeOfAReclaimCapParkNeedsNoLeaseAndClearsTheCount`(본문 `:212`가 `got.ReclaimCount != 0`을 실패로 단언하는 것까지 읽음); `internal/cli/af_resume_test.go:99` `TestAFVerbs_N42TheResumeVerbRefusesWithoutAReasonOrWithAnotherGuardOrEpochAndChangesNothing`, `:141` `TestAFVerbs_N42TheResumeVerbRefusesACardThatIsNotParkedAndChangesNothing`, `:160` `TestAFVerbs_N42TheResumeVerbLetsAReclaimCapParkedCardFinishOnItsLandedCommit`. 다른 epoch의 거절이 원장 층에서 나는지, 동사 층에서만 나는지는 확인하지 않았다(progress §O).
- 시험 이름(하나): `TestReap_ThirdReclaim_ParksCard` — 소유 lane-18, 패키지 `internal/agenticfactory/ledger`. unwritten at 37b67f6d39(§B).
- RED-now·green path(0.5.0): 이 AC는 release 게이트 밖이다 — §N 사후 관측 행 AC-AFL-007. RED 미관측. 완료 계산에 들지 않는다(§D). 시험 이름은 위 한 개 그대로다.
- 이전 문구(0.4.0, RED-now·green path): «name absent at 37b67f6d39 — `TestReap_ThirdReclaim_ParksCard`가 `37b67f6d39`에 없다(§B). integration tip `1a552af405`(조정자가 적은 `915a6524cb`에서 움직인 tip)에서 이 이름은 없다. 착지 시험이 이 AC를 덮는 범위: partly — 셋째 회수 park, CANCELLED·DONE Move 뒤 lease 0, 네 기록 표의 DELETE 거절은 덮고, 사례 (iii)의 행 생존과 DONE 뒤 lease_keys 0은 덮지 않는다(§F 행 AC-AFL-007). 착지 전 RED는 관측되지 않았고 이 칸은 판정한 RED가 아니다 — 통과로 적지 않는다.» «RED evidence not recoverable from history (landed before this SPEC); this AC adopts the landed test as a regression-guard, not a release-blocking RED (`verification-completeness.md` §2.1 undecidable disposition). 남은 일: §F에 적은 덮이지 않은 부분을 이 이름의 시험(소유 lane-18)이 더한다. 그 시험은 이미 착지한 코드 위에서 쓰이므로 처음부터 초록일 수 있고, RED가 아니라 regression-guard로 다룬다.»
- 이전 문구(0.3.0, RED-now·green path): «`TestReap_ThirdReclaim_ParksCard`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.» «lane-18의 Reap 구현 커밋이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.»

#### AC-AFL-008 — Enqueue는 같은 commit SHA의 관문 증거를 요구함

- 주 REQ: maps REQ-AFL-008.
- Given 관문 증거 행(`subject_sha` = 커밋 A)을 가진 카드 P와 Q, When (i) P에 `Enqueue(candidateSHA=A)`, (ii) Q에 `Enqueue(candidateSHA=B)`, (iii) 증거가 있는 다른 카드를 `Move`로 MERGING에 바로 옮기려 할 때, Then (i) 받아들여져 P가 MERGING이고, (ii) `ErrEvidenceStale`이 돌아오며 Q의 단계는 그대로이고, (iii) 거절되어 단계가 그대로다.
- 죽여야 할 변이: SHA를 비교하지 않는 구현 → (ii)에서 FAIL; 아무 Enqueue도 받지 않는 구현 → (i)에서 FAIL; Move로 MERGING을 허용하는 구현 → (iii)에서 FAIL.
- 같은 AC의 0.6.0 사례(REQ-AFL-008 Enqueue가 더 거절하는 것): (iv) Given 커밋 A의 관문 증거 뒤에 커밋 B의 green 증거를 남긴 카드, When `Enqueue(candidateSHA=A)`, Then `ErrEvidenceStale`이고 카드는 그대로다 — 단 A에 대한 리뷰나 A에 대한 실패 기록은 "A를 다시 잰 것"으로 세지 않는다; (v) Given A에 대한 가장 새 green·gate·review 기록이 실패인 카드, When `Enqueue(A)`, Then `ErrEvidenceStale`; (vi) Given A를 이미 큐에 낸 뒤 BUILDING으로 되돌아간 카드, When 그 이동 뒤 A의 통과한 green·gate 실행 없이 `Enqueue(A)`, Then `ErrEvidenceStale`이고, 그 이동 뒤 통과한 실행을 기록하면 받아들여진다; (vii) Given 이전 카드 epoch에서 landed가 된 큐 행을 가진 회수된 카드, When 현재 epoch에서 아무것도 재지 않고 무엇이든 낼 때, Then `ErrInvalid`이고, 현재 epoch에서 A의 통과한 실행이 기록되어 있으면 행을 이어받거나 새로 시작한다. "이동 뒤"는 evidence id와 `evidence_mark`로 정한다(시계 아님). 죽여야 할 변이: 후보의 증거만 보고 더 새 다른 커밋 작업을 보지 않는 구현 → (iv)에서 FAIL; 실패한 최신 실행을 무시하는 구현 → (v)에서 FAIL; 되돌아간 뒤 다시 재지 않은 제출을 받는 구현 → (vi)에서 FAIL.
  - 착지 시험(이름·줄만 확인, 본문 덮음 판정 안 함, 돌리지 않음): `internal/agenticfactory/ledger/enqueue_newest_evidence_test.go:20` `TestEnqueue_TheCandidateMustBeWhatTheCardMeasuredLast`; `internal/agenticfactory/ledger/enqueue_measuring_again_test.go:37` `TestEnqueue_AReviewOrAFailureOfTheCandidateIsNotMeasuringItAgain`, `:89` `TestEnqueue_AFailingRunOfTheCandidateDoesNotHideTheNewerCommit`; `internal/agenticfactory/ledger/enqueue_rework_test.go:38` `TestEnqueue_ACommitOfferedBeforeTheMoveBackNeedsARunRecordedSince`; `internal/agenticfactory/ledger/enqueue_reclaimed_landed_test.go:58` `TestEnqueue_AReclaimedCardTakesOverOrReplacesTheLandedRowOnceItMeasuredInItsOwnEpoch`; `internal/cli/af_land_n3_test.go:178` `TestAFVerbs_N3AReclaimedCardThatMeasuredNothingInItsOwnEpochCannotOfferAnything`.
- 시험 이름(하나): `TestEnqueue_EvidenceSubjectSHAMismatch_ReturnsErrEvidenceStale` — 소유 lane-18, 패키지 `internal/agenticfactory/ledger`. unwritten at 37b67f6d39(§B).
- RED-now·green path(0.5.0): 이 AC는 release 게이트 밖이다 — §N 사후 관측 행 AC-AFL-008. RED 미관측. 완료 계산에 들지 않는다(§D). 시험 이름은 위 한 개 그대로다.
- 이전 문구(0.4.0, RED-now·green path): «name absent at 37b67f6d39 — `TestEnqueue_EvidenceSubjectSHAMismatch_ReturnsErrEvidenceStale`가 `37b67f6d39`에 없다(§B). integration tip `1a552af405`(조정자가 적은 `915a6524cb`에서 움직인 tip)에서 이 이름은 없다. 착지 시험이 이 AC를 덮는 범위: partly — 사례 (i)(ii)(iii)의 오류·행은 세 시험이 나눠 덮고, (iii)의 단계 불변은 덮지 않는다(§F 행 AC-AFL-008). 착지 전 RED는 관측되지 않았고 이 칸은 판정한 RED가 아니다 — 통과로 적지 않는다.» «RED evidence not recoverable from history (landed before this SPEC); this AC adopts the landed test as a regression-guard, not a release-blocking RED (`verification-completeness.md` §2.1 undecidable disposition). 남은 일: §F에 적은 덮이지 않은 부분을 이 이름의 시험(소유 lane-18)이 더한다. 그 시험은 이미 착지한 코드 위에서 쓰이므로 처음부터 초록일 수 있고, RED가 아니라 regression-guard로 다룬다.»
- 이전 문구(0.3.0, RED-now·green path): «`TestEnqueue_EvidenceSubjectSHAMismatch_ReturnsErrEvidenceStale`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.» «lane-18의 AddEvidence·Enqueue 구현 커밋이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.»
- 이전 문구(0.1.0): 제목 "증거는 tree SHA에 묶임"; Given tree SHA A로 기록된 증거 행, When tree SHA B를 판정할 때 그 증거를 셀 때, Then 그 증거는 세어지지 않는다. 시험 이름 `TestEvidence_DifferentTreeSHA_IsInvalid`.

#### AC-AFL-009 — 스키마와 CHECK 제약

- 주 REQ: maps REQ-AFL-009.
- Given 막 만든 원장, When `merge_queue.state`에 목록 밖 값을, `notes.kind`에 목록 밖 값을 넣으려 할 때, Then 두 쓰기 모두 CHECK 제약 위반으로 거절되고, 여덟 표(`cards`·`transitions`·`leases`·`lease_keys`·`evidence`·`merge_queue`·`locks`·`notes`)가 모두 있으며 `lease_keys.key`가 기본 키다.
- 같은 AC의 0.6.0 사례(REQ-AFL-009 스키마 판과 이전): 위 Then의 "여덟 표"는 0.5.1 문구이며 지우지 않는다; 0.6.0 착지 코드에서는 아홉째 표 `repository`(한 행)와 `transitions.evidence_mark` 열이 더 있다. (b) Given `PRAGMA user_version` 1로 쓰인 원장(evidence_mark 열 없음), When `Open`할 때, Then 한 쓰기 트랜잭션 안에서 열이 더해지고 판이 2가 되며, 각 transition의 mark는 그 시각 이전에 기록된 카드의 가장 새 evidence id다 — 같은 밀리초의 evidence는 MERGING으로 가는 이동에서는 뒤로, 다른 이동에서는 앞으로 센다; (c) Given 판 3인 파일, When `Open`, Then 거절되고 파일이 바뀌지 않는다; (d) Given 이전 판의 파일, When `Open`, Then `repository` 표가 생기고 묶음이 비어 있다. 죽여야 할 변이: 같은 밀리초를 한 방향으로만 세는 이전 → (b)의 MERGING 사례나 BUILDING 되돌림 사례 중 하나에서 FAIL; 판을 보지 않는 Open → (c)에서 FAIL.
  - 착지 시험(이름·줄만 확인, 본문 덮음 판정 안 함, 돌리지 않음): `internal/agenticfactory/ledger/rework_migration_test.go:102` `TestOpen_MigratesALedgerWrittenBeforeEvidenceMarks`, `:158` `TestOpen_RefusesALedgerNewerThanTheCode`; `internal/agenticfactory/ledger/migration_same_millisecond_test.go:77` `TestOpen_ASameMillisecondGreenAfterTheOfferIsNewerWork`, `:115` `TestOpen_ASameMillisecondMeasurementOfTheMoveBackStillRanksBeforeIt`; `internal/agenticfactory/ledger/repository_identity_test.go:157` `TestRepository_FilesOfEarlierSchemasOpenIntoTheNewOne`. (c)의 "파일이 바뀌지 않는다"는 코드가 판 비교(`schema.go:123-125`)를 DDL 실행(`:126`) 앞에 둔 것에서 읽었을 뿐 시험 본문으로 확인하지 않았다.
- 시험 이름(하나): `TestSchema_CheckConstraints_RejectUnknownQueueStateAndNoteKind` — 소유 lane-18, 패키지 `internal/agenticfactory/ledger`. unwritten at 37b67f6d39(§B).
- RED-now·green path(0.5.0): 이 AC는 release 게이트 밖이다 — §N 사후 관측 행 AC-AFL-009. RED 미관측. 완료 계산에 들지 않는다(§D). 시험 이름은 위 한 개 그대로다.
- 이전 문구(0.4.0, RED-now·green path): «name absent at 37b67f6d39 — `TestSchema_CheckConstraints_RejectUnknownQueueStateAndNoteKind`가 `37b67f6d39`에 없다(§B). integration tip `1a552af405`(조정자가 적은 `915a6524cb`에서 움직인 tip)에서 이 이름은 없다. 착지 시험이 이 AC를 덮는 범위: partly — 여덟 표와 두 CHECK 거절은 덮고, `lease_keys.key` 기본 키는 덮지 않는다(§F 행 AC-AFL-009). 착지 전 RED는 관측되지 않았고 이 칸은 판정한 RED가 아니다 — 통과로 적지 않는다.» «RED evidence not recoverable from history (landed before this SPEC); this AC adopts the landed test as a regression-guard, not a release-blocking RED (`verification-completeness.md` §2.1 undecidable disposition). 남은 일: §F에 적은 덮이지 않은 부분을 이 이름의 시험(소유 lane-18)이 더한다. 그 시험은 이미 착지한 코드 위에서 쓰이므로 처음부터 초록일 수 있고, RED가 아니라 regression-guard로 다룬다.»
- 이전 문구(0.3.0, RED-now·green path): «`TestSchema_CheckConstraints_RejectUnknownQueueStateAndNoteKind`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.» «lane-18의 schema 커밋이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.»

#### AC-AFL-010 — Claim 한 트랜잭션·등급별 시작 단계

- 주 REQ: maps REQ-AFL-010.
- Given 우선순위가 다른 READY 카드 둘(class C 하나, class A 하나), When `Claim`을 부를 때, Then 우선순위가 높은 카드 하나만 lease·키·epoch 증가와 함께 BUILDING/SPEC(C) 또는 BUILDING/IMPLEMENT(A·B)로 옮겨지고 id와 epoch가 돌아오며, 고를 카드가 없으면 `ErrNoWork`가 돌아온다.
- 전제(0.5.0): 첫 줄의 '고를 카드'는 REQ-AFL-010 0.5.0 문장대로 READY 카드와 회수로 다시 claim 대상이 된 카드다; 뒤의 집합은 미정이다(OD-43). 아래 L-Dead 행은 회수만 단언하고 H가 다시 claim되는지는 단언하지 않는다. 이전 문구(0.4.0): 이 전제 줄은 없었다.
- 같은 시험의 lazy reap 하위 사례: Given lease 기한이 지난 카드 H(소유자 O)와 키가 겹치지 않는 READY 카드 하나, 그리고 O에 대해 아래 판정을 돌려주는 `ClaimRequest`의 활성 판정 함수, When 다른 세션이 `Claim`을 부를 때, Then 사례마다 아래 표와 같다.

| 사례 | O의 판정 | H의 lease·키·`cards.epoch`·`reclaim_count` | `stalled` 쪽지 | 다른 READY 카드 |
|---|---|---|---|---|
| L-Live | 살아 있음 | 그대로(회수 안 함) | H의 그 card·epoch에 정확히 하나 — `Claim`을 두 번 불러도 하나(계약 §10 20th (d)) | claim됨 |
| L-Unknown | 판정 불명 | 그대로(살아 있는 것으로 다룸) | 단언하지 않음 | claim됨 |
| L-Dead | 죽음 | 이전 lease·키 없어지고 `reclaim_count` +1(회수) | 단언하지 않음 | 단언하지 않음 |
| L-Changed-epoch | 죽음으로 판정했으나, 판정(트랜잭션 밖) 뒤 트랜잭션 안 재확인 전에 H lease의 epoch가 바뀜 | 그대로(회수 안 함; 계약 §10 "re-checks that epoch and heartbeat_at are unchanged") | 단언하지 않음 | 단언하지 않음 |
| L-Changed-heartbeat | 죽음으로 판정했으나, 판정(트랜잭션 밖) 뒤 트랜잭션 안 재확인 전에 H lease의 `heartbeat_at`이 바뀜(epoch는 그대로) | 그대로(회수 안 함; 같은 계약 문장) | 단언하지 않음 | 단언하지 않음 |

- 죽여야 할 변이: `Claim`이 기한 지난 Live 소유자를 회수하는 구현 → L-Live에서 FAIL(I5 "A live process is never reclaimed automatically"). 그 밖에 Unknown을 회수하는 구현 → L-Unknown, Dead를 회수하지 않는 구현 → L-Dead, 트랜잭션 안에서 epoch를 다시 보지 않는 구현 → L-Changed-epoch에서 FAIL, `heartbeat_at`을 다시 보지 않는 구현 → L-Changed-heartbeat에서 FAIL(0.5.0, D8: 두 조건을 따로 잡도록 행을 나눴다).
- AC-AFL-005는 그대로 둔다(그 AC는 `Reap` 경로, 이 AC는 `Claim`의 lazy reap 경로).
- 이전 문구(0.4.0, L-Changed 행): «| L-Changed | 죽음으로 판정했으나, 판정(트랜잭션 밖) 뒤 트랜잭션 안 재확인 전에 H lease의 `heartbeat_at` 또는 epoch가 바뀜 | 그대로(회수 안 함; 계약 §10 "re-checks epoch+heartbeat_at inside") | 단언하지 않음 | 단언하지 않음 |» 변이 줄: «트랜잭션 안에서 epoch·`heartbeat_at`을 다시 보지 않는 구현 → L-Changed에서 FAIL.»
- 이전 문구(0.3.0): lazy reap 하위 사례와 변이 줄이 없었다(Given 문장은 위 첫 줄 그대로).
- 시험 이름(하나): `TestClaim_SingleTransaction_ReadyToBuildingByClass` — 소유 lane-18, 패키지 `internal/agenticfactory/ledger`. unwritten at 37b67f6d39(§B).
- RED-now·green path(0.5.0): 이 AC는 release 게이트 밖이다 — §N 사후 관측 행 AC-AFL-010. RED 미관측. 완료 계산에 들지 않는다(§D). 시험 이름은 위 한 개 그대로다.
- 이전 문구(0.4.0, RED-now·green path): «name absent at 37b67f6d39 — `TestClaim_SingleTransaction_ReadyToBuildingByClass`가 `37b67f6d39`에 없다(§B). integration tip `1a552af405`(조정자가 적은 `915a6524cb`에서 움직인 tip)에서 이 이름은 없다. 착지 시험이 이 AC를 덮는 범위: partly — 등급별 시작 단계·우선순위·`ErrNoWork`·L-Live·L-Unknown·L-Dead·heartbeat 갱신 L-Changed는 여러 시험이 나눠 덮고, epoch 변경 L-Changed는 덮지 않는다(§F 행 AC-AFL-010). 착지 전 RED는 관측되지 않았고 이 칸은 판정한 RED가 아니다 — 통과로 적지 않는다.» «RED evidence not recoverable from history (landed before this SPEC); this AC adopts the landed test as a regression-guard, not a release-blocking RED (`verification-completeness.md` §2.1 undecidable disposition). L-Live 사례는 lane-18 커밋 `607bd17804`(가지 `WT-agentic-factory-core-claim-runner`; integration tip `1a552af405`에 들지 않음)의 `TestClaim_LazyReap_DoesNotReclaimLiveOwner`가 덮는다(§F). 이름은 바꾸지 않는다 — 착지 시험 하나가 AC 전체를 덮지 않는다. 남은 일: §F에 적은 덮이지 않은 부분을 이 이름의 시험(소유 lane-18)이 더한다. 그 시험은 이미 착지한 코드 위에서 쓰이므로 처음부터 초록일 수 있고, RED가 아니라 regression-guard로 다룬다.»
- 이전 문구(0.3.0, RED-now·green path): «`TestClaim_SingleTransaction_ReadyToBuildingByClass`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.» «lane-18의 Claim 구현 커밋이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.»

#### AC-AFL-011 — 충돌 키와 ExpandKeys

- 주 REQ: maps REQ-AFL-011.
- Given 카드 X가 `shared:go.mod` 키를 잡고 있는 상태, When 카드 Y가 `ExpandKeys`로 같은 키를 요청할 때, Then `ErrKeyConflict`가 돌아오고 어느 카드도 park되지 않는다.
- 시험 이름(하나): `TestExpandKeys_Conflict_ReturnsErrKeyConflict` — 소유 lane-18, 패키지 `internal/agenticfactory/ledger`. unwritten at 37b67f6d39(§B).
- RED-now·green path(0.5.0): 이 AC는 release 게이트 밖이다 — §N 사후 관측 행 AC-AFL-011. RED 미관측. 완료 계산에 들지 않는다(§D). 시험 이름은 위 한 개 그대로다.
- 이전 문구(0.4.0, RED-now·green path): «name absent at 37b67f6d39 — `TestExpandKeys_Conflict_ReturnsErrKeyConflict`가 `37b67f6d39`에 없다(§B). integration tip `1a552af405`(조정자가 적은 `915a6524cb`에서 움직인 tip)에서 이 이름은 없다. 착지 시험이 이 AC를 덮는 범위: partly — 요청한 쪽의 `ErrKeyConflict`·키 불변·park 없음은 덮고, 키를 쥔 쪽의 park 없음은 덮지 않는다(§F 행 AC-AFL-011). 착지 전 RED는 관측되지 않았고 이 칸은 판정한 RED가 아니다 — 통과로 적지 않는다.» «RED evidence not recoverable from history (landed before this SPEC); this AC adopts the landed test as a regression-guard, not a release-blocking RED (`verification-completeness.md` §2.1 undecidable disposition). 남은 일: §F에 적은 덮이지 않은 부분을 이 이름의 시험(소유 lane-18)이 더한다. 그 시험은 이미 착지한 코드 위에서 쓰이므로 처음부터 초록일 수 있고, RED가 아니라 regression-guard로 다룬다.»
- 이전 문구(0.3.0, RED-now·green path): «`TestExpandKeys_Conflict_ReturnsErrKeyConflict`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.» «lane-18의 KeysFor·ExpandKeys 구현 커밋이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.»

#### AC-AFL-012 — 겹치지 않는 키는 동시에 진행 (D4)

- 주 REQ: maps REQ-AFL-012.
- Given 키가 겹치지 않는 READY 카드 둘, When 두 세션이 차례로 `Claim`을 부를 때, Then 두 카드가 모두 claim되어 동시에 BUILDING에 있고, 어느 쪽도 전역 직렬 slot을 기다리지 않는다.
- 시험 이름(하나): `TestClaim_DisjointConflictKeys_ClaimInParallel` — 소유 lane-18, 패키지 `internal/agenticfactory/ledger`. unwritten at 37b67f6d39(§B).
- RED-now·green path(0.5.0): 이 AC는 release 게이트 밖이다 — §N 사후 관측 행 AC-AFL-012. RED 미관측. 완료 계산에 들지 않는다(§D). 시험 이름은 위 한 개 그대로다.
- 이전 문구(0.4.0, RED-now·green path): «name absent at 37b67f6d39 — `TestClaim_DisjointConflictKeys_ClaimInParallel`가 `37b67f6d39`에 없다(§B). integration tip `1a552af405`(조정자가 적은 `915a6524cb`에서 움직인 tip)에서 이 이름은 없다. 착지 시험이 이 AC를 덮는 범위: partly — 겹치지 않는 카드가 모두 claim되는 것은 덮고, 두 카드가 동시에 BUILDING에 있는 단계는 덮지 않는다(§F 행 AC-AFL-012). 착지 전 RED는 관측되지 않았고 이 칸은 판정한 RED가 아니다 — 통과로 적지 않는다.» «RED evidence not recoverable from history (landed before this SPEC); this AC adopts the landed test as a regression-guard, not a release-blocking RED (`verification-completeness.md` §2.1 undecidable disposition). 남은 일: §F에 적은 덮이지 않은 부분을 이 이름의 시험(소유 lane-18)이 더한다. 그 시험은 이미 착지한 코드 위에서 쓰이므로 처음부터 초록일 수 있고, RED가 아니라 regression-guard로 다룬다.»
- 이전 문구(0.3.0, RED-now·green path): «`TestClaim_DisjointConflictKeys_ClaimInParallel`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.» «lane-18의 Claim 구현 커밋이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.»

#### AC-AFL-013 — 카드 흐름 claim → DONE

- 주 REQ: maps REQ-AFL-013.
- Given class A 카드 하나와 시험 저장소, When 카드를 claim → RED 증거 → GREEN → 관문 → MERGING(`Enqueue`)으로 몰 때, Then `landed` 쪽지 뒤에 카드가 DONE이고 lease가 풀려 있으며, `spec.risk`가 비고 diff가 400줄 이하라 REVIEWING을 거치지 않았다.
- 시험 이름(하나): `TestCardFlow_ClaimToDone_ThroughMergeQueue` — 소유 lane-22, 패키지 미정(통합 시험 위치, OD-35). unwritten at 37b67f6d39(§B).
- RED-now(0.5.1, release 게이트, 감사 N1 — lane report 기록이며 이 작성자가 돌린 것이 아니다): 출처는 lane-22의 lane report `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/agentic-factory-integration/.moai/reports/issue-1948/af-verbs-ac017-ac013.md`(integration 트리의 gitignore된 `.moai/reports`). 측정 커밋: `95e100e754`(RED 커밋, "a skeleton `newAFCmd` without verbs, not registered"), 기준 `03490b9b8f`. 명령: lane report af-verbs-ac017-ac013.md L32(이 SPEC에는 실행할 수 있는 형태로 옮기지 않는다). lane report가 옮긴 출력(그대로):
  ```
  === RUN   TestCardFlow_ClaimToDone_ThroughMergeQueue
      af_flow_test.go:85: moai af claim --db .../af/ledger.db --session s-t1 --worktree ...: unknown flag: --db
          stderr: Error: unknown flag: --db
  --- FAIL: TestCardFlow_ClaimToDone_ThroughMergeQueue (0.18s)
  FAIL
  FAIL	github.com/modu-ai/moai-adk/internal/cli	1.362s
  ```
  종료 코드: exit code not recorded as a field; non-zero implied by the FAIL summary line only. 명령 형식에 대한 발견(기록만 함, 여기서 고치지 않음 — `verification-completeness.md` §1.1): lane report가 옮긴 명령은 앵커 없는 `-run` 패턴(`^…$` 없이 `-run TestName`)을 쓴다. 그래서 고른 시험이 하나도 없는 실행도 `ok`를 찍었을 것이다. lane report 출력의 `=== RUN` 줄 하나가 의도한 시험이 돌았다는 유일한 표시이며, 이 작성자는 그 명령을 돌리지 않았다. 이전 문구(0.5.1 첫 쓰기): «고른 시험의 수: 명령의 `-run` 인자는 앵커가 없어 더 긴 이름의 시험도 고를 수 있다(spec lint `VacuousTestAssertion` 경고가 이 줄을 가리킨다; 명령은 lane report 그대로 두었다). lane report 출력에 `=== RUN` 줄이 하나뿐이므로 그 실행이 고른 시험은 이 이름 하나로 읽힌다 — 이 작성자가 돌려 확인한 것은 아니다.» 빨강의 이유(lane report가 적은 대로): 동사가 없는 뼈대 `moai af`라 `claim` 동사와 `--db` 플래그가 없다 — 이 작업이 더할 동사의 부재이며 손대지 않는 파일 때문이 아니다. 커밋 그래프 순서(이 작성자가 이 작업에서 돌림): `git -C <worktree> merge-base --is-ancestor 95e100e754 b58b1566b6` → `exit=0`(RED가 GREEN의 조상); 반대 방향 대조 `git -C <worktree> merge-base --is-ancestor b58b1566b6 5a9370f894` → `exit=1`. `git -C <worktree> branch --contains 95e100e754 --list WT-agentic-factory-integration` → `+ WT-agentic-factory-integration`. 시험 위치: `95e100e754`에 `internal/cli/af_flow_test.go:49`가 있다(`git grep -n`으로 확인).
- GREEN 기록(0.5.1, lane report — 이 작성자가 관측하지 않음): 커밋 `b58b1566b6`, 명령: lane report af-verbs-ac017-ac013.md L44, lane report가 옮긴 줄: `--- PASS: TestCardFlow_ClaimToDone_ThroughMergeQueue (0.40s)` / `PASS` / `ok  	github.com/modu-ai/moai-adk/internal/cli	1.587s`. lane report 자신의 Gaps: REVIEWING을 건너뛴 것을 위치 이력으로 단언하지 않는다; 3회 수리 한도와 `fail_reason` 경로를 거치지 않는다; 관문 실행기는 nil을 돌려주는 대역이다; `internal/cli` 전체는 돌리지 않았다. 그래서 Then 문장의 "REVIEWING을 거치지 않았다"는 이 GREEN 기록이 덮지 않는다.
- 이전 문구(0.5.0, RED-now): «RED-now(0.5.0, release 게이트): 시험이 쓰이지 않은 상태(unwritten test) — `TestCardFlow_ClaimToDone_ThroughMergeQueue`가 `37b67f6d39`에 없고(§B, 0.2.0 측정의 종료 코드 1) tip `db02401ad1`에도 없다(§B.2 [L-013]). 시험이 없으므로 실행 명령을 두지 않는다(§A). RED 미관측.»
- 이전 문구(0.4.0, RED-now): «name absent at 37b67f6d39 — `TestCardFlow_ClaimToDone_ThroughMergeQueue`가 `37b67f6d39`에 없다(§B). integration tip `1a552af405`(조정자가 적은 `915a6524cb`에서 움직인 tip)에서 이 이름은 없다. 착지 시험이 이 AC를 덮는 범위: 착지 시험 없음(no landed test)(§F 행 AC-AFL-013). 착지 전 RED는 관측되지 않았고 이 칸은 판정한 RED가 아니다 — 통과로 적지 않는다.»
- green path: lane-22의 CLI·통합 커밋(계약 §7 P2)이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다. (0.4.0: 이 AC의 코드는 `1a552af405`에 착지하지 않았으므로 RED 커밋이 GREEN보다 먼저라는 경로가 그대로 남는다.)
- 이전 문구(0.3.0, RED-now·green path): «`TestCardFlow_ClaimToDone_ThroughMergeQueue`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.» «lane-22의 CLI·통합 커밋(계약 §7 P2)이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.»

#### AC-AFL-014 — 병합 큐 착지와 완료 증거 (D3)

- 주 REQ: maps REQ-AFL-014.
- Given MERGING 카드와 후보 SHA, When 병합 큐가 그 카드를 처리할 때, Then 후보 worktree 검증이 잠금 밖에서 끝난 뒤 `merge-queue:landing`을 잡고, `git update-ref <trunk> <new> <old>`로 trunk를 옮기고, `landed` 쪽지를 남기고 카드를 DONE으로 옮긴다; 리더 승인 receipt는 쓰이지 않는다.
- 같은 시험의 사례(0.4.0, N14): (b) Given landing 잠금을 다른 epoch로 다시 잡은 상태(앞선 잠금 epoch는 낡음), When 낡은 잠금 epoch를 실은 `QueueUpdate`로 큐 행을 landing·landed로 옮기려 할 때, Then `ErrStaleEpoch`가 돌아오고 큐 행의 상태는 그대로다; (c) Given trunk ref가 어느 worktree에 checkout되어 있는 상태, When 착지를 시도할 때, Then 착지가 거절되고 trunk ref가 움직이지 않으며 landing 잠금을 잡지 않는다.
- 같은 시험의 사례(0.5.1, 감사 N2 — 사례마다 판정 층을 적는다): (d1) [저장소 층, `QueueUpdate`] Given landing 상태의 큐 행, When 재시작 뒤 reconcile이 그 행을 waiting으로 되돌리는 쓰기를 카드 epoch 값으로 실을 때, Then `ErrStaleEpoch`이고 행 상태가 그대로다; (d2) [저장소 층 + 병합 큐 reconcile 층] Given landing 상태의 큐 행, When 쥐고 있는 `merge-queue:landing` 잠금 epoch로 같은 되돌림을 실을 때, Then 받아들여져 행이 waiting이 된다 — 그리고 verifying 상태의 큐 행은 그 행의 `card_epoch`가 `cards.epoch`와 다르거나 카드에 살아 있는 lease가 없을 때에만 reconcile이 waiting으로 되돌린다; (d2′) [병합 큐 reconcile 층, 반대 사례] Given verifying 상태의 큐 행이 현재 카드 epoch를 싣고 그 카드의 lease가 살아 있는 상태, When reconcile이 돌 때, Then 그 행은 verifying 그대로이고 reconcile은 그 행에 아무 동작도 내지 않는다(살아 있는 착지자의 검증은 건드리지 않음); (d3) [병합 큐 reconcile 층] Given 잠금을 쥐지 못한 reconcile, When 같은 되돌림을 시도할 때, Then 행을 바꾸지 않는다. 계약 문장(§10 20th review lane-22 (a), 고정한 계약 `/Users/goos/MoAI/moai-adk-go/.moai/reports/issue-1948/local-adk-design-20261030.md` 256-257행, sha256 `f2c1078b…` 이 작업에서 다시 확인): "Reconcile reverts a `verifying` row only when its `card_epoch` differs from `cards.epoch` or the card has no live lease — a live lander's verification is left alone". 계약 문장(§10 19th review, 잠금 조건은 그대로 유효): "Reconcile reverts verifying/landing rows to waiting only while it holds the `merge-queue:landing` lock, checked by that lock epoch."
- 죽여야 할 변이(0.5.1): landing·landed 쓰기에서 잠금 epoch를 보지 않는 구현 → (b)에서 FAIL; checkout 여부를 보지 않고 `update-ref`하는 구현 → (c)에서 FAIL; reconcile의 waiting 되돌림을 카드 epoch로 받는 구현 → (d1)에서 FAIL; 잠금 없이 되돌리는 구현 → (d3)에서 FAIL; 살아 있는 lease를 가진 착지자의 verifying 행을 되돌리는 구현 → (d2′)에서 FAIL; lease가 끊긴 verifying 행을 끝내 되돌리지 않는 구현 → (d2)의 verifying 부분에서 FAIL.
- 착지 사실(0.5.1, 시험 아님 — 사례 (d2)(d2′)의 verifying 부분을 덮는 착지 시험, 이 작성자가 본문을 읽음): `b28a31d9dc:internal/agenticfactory/mergeq/real_ledger_reconcile_test.go:16`(`TestReconcile_RealLedger_AVerifyingRowWithALiveLeaseIsLeftAlone`: 실제 ledger 위에서 verifying 행의 lease가 살아 있으면 `m.Reconcile`이 동작 0개를 내고 행이 `QueueVerifying` 그대로이며, 대조로 시각을 24시간 뒤로 옮겨 lease가 끊기면 동작 1개를 내고 행이 `QueueWaiting`이 된다). 이 시험을 이 작성자가 돌려 본 것은 아니다. 이 AC의 시험 이름은 아래 한 개 그대로다.
- 이전 문구(0.5.0, 사례 (d)): «(d) Given landing 상태의 큐 행(그리고 verifying 상태의 큐 행), When 재시작 뒤 reconcile이 그 행을 waiting으로 되돌리는 `QueueUpdate`를 (d1) 카드 epoch 값으로, (d2) 쥐고 있는 `merge-queue:landing` 잠금 epoch로 실을 때, 그리고 (d3) 잠금을 쥐지 못한 reconcile이 같은 되돌림을 시도할 때, Then (d1)은 `ErrStaleEpoch`이고 행 상태가 그대로이며, (d2)는 받아들여져 행이 waiting이 되고, (d3)은 행을 바꾸지 않는다. 계약 문장(§10 19th review): "Reconcile reverts verifying/landing rows to waiting only while it holds the `merge-queue:landing` lock, checked by that lock epoch."» (0.5.0 문구는 verifying 행이 잠금 epoch만 맞으면 늘 되돌아가는 것처럼 읽혔고, 20th lane-22 (a)의 lease 조건이 빠져 있었다 — 감사 N2.)
- 이전 문구(0.5.0, 변이 줄): «landing·landed 쓰기에서 잠금 epoch를 보지 않는 구현 → (b)에서 FAIL; checkout 여부를 보지 않고 `update-ref`하는 구현 → (c)에서 FAIL; reconcile의 waiting 되돌림을 카드 epoch로 받는 구현 → (d1)에서 FAIL; 잠금 없이 되돌리는 구현 → (d3)에서 FAIL.»
- 착지 사실(0.5.0, 시험 아님 — 사례 (d)의 일부를 덮는 착지 시험): `db02401ad1:internal/agenticfactory/ledger/queue_test.go:170-176`(`TestQueueUpdateChecksTheEpochItsStateCallsFor` 안, 주석 "a reconcile sends a landing row back to waiting, under the lock": landing 행을 카드 epoch 값으로 waiting에 → `ErrStaleEpoch`, 잠금 epoch로 → 오류 없음)와 `db02401ad1:internal/agenticfactory/mergeq/mergeq_test.go:586`(`TestReconcile_PutsARowBackToWaitingOnlyWhileHoldingTheLandingLock`: 잠금이 `ErrLockHeld`면 행을 건드리지 않고, 잡으면 waiting으로 한 번 옮긴 뒤 잠금을 한 번 잡고 한 번 푼다). verifying 행의 되돌림은 두 시험 모두 단언하지 않는다(§F 행 AC-AFL-014).
- 이전 문구(0.4.0, 변이 줄): «landing·landed 쓰기에서 잠금 epoch를 보지 않는 구현 → (b)에서 FAIL; checkout 여부를 보지 않고 `update-ref`하는 구현 → (c)에서 FAIL.» 사례 (d)는 없었다.
- 이전 문구(0.3.0): 사례 (b)(c)와 그 변이 줄이 없었다(위 Given 문장은 0.1.0 그대로).
- 같은 AC의 0.6.0 사례(REQ-AFL-014 0.6.0 (i)~(v)):
  - (e) [동사 층, trunk 고정] Given 원장과 저장소, When `moai af land`나 `moai af reconcile`에 `refs/heads/af/trunk`가 아닌 `--trunk`를 줄 때, Then 아무것도 착지하기 전에 거절되고 오류가 `--trunk`와 전용 trunk를 이름 대며 원장이 그대로다. 착지 시험(본문을 읽음 — land·reconcile 두 동사를 돌며 오류 문구와 원장 스냅숏 불변을 단언; 돌리지 않음): `internal/cli/af_land_repair_test.go:188` `TestAFLand_ATrunkOtherThanTheDedicatedOneIsRefusedBeforeLandingAndTheLedgerStaysAsItWas`; 도움말: 같은 파일 `:300` `TestAFVerbs_TheHelpSaysTheTrunkCannotBeChangedAndWhichRepositoriesAreAccepted`.
  - (f) [병합 큐 층, 실패 기록] Given 실패하는 관문, 병합 충돌, 실패하는 다른 git 단계, When 착지할 때, Then 원장·쪽지·실패 사유·동사 출력에는 종류·관문 id·원인·바이트 수·로그 경로(또는 "logging failed")만 있고 관문 글이나 출력은 없으며, 출력은 `gate-logs`의 0600 파일(0700 디렉터리)에 가려지고 잘려 남는다; 로그를 쓸 수 없어도 실패는 기록된다; 링크인 디렉터리는 거절되고 그 대상은 손대지 않는다; 디렉터리 합계 상한을 넘으면 가장 오래된 파일부터 지워지고, 1분이 지난 남은 임시 파일은 지워지며 새 임시 파일은 남아 합계에 든다; macOS에서 ACL을 지울 수 없으면 출력을 버리고 사유가 그렇게 말한다. 착지 시험(이름·줄만 확인, 본문 덮음 판정 안 함, 돌리지 않음): `internal/agenticfactory/mergeq/faillog_test.go:23` `TestFailureLog_FileModeDirectoryModeCapAndLastLine`, `:100` `TestFailureLog_OneFilePerCardCandidateAndKindAndTheNextOverwrites`, `:249` `TestLand_AGateErrorsOutputIsInTheLogAndNotInTheReason`, `:311` `TestLand_AMergeConflictsGitOutputIsInTheLogAndNotInTheReason`, `:335` `TestLand_ALogThatCannotBeWrittenStillRecordsTheFailure`; `internal/agenticfactory/mergeq/faillog_repair_test.go:55` `TestFailureLog_TheCapHoldsAfterTheMaskingMadeTheTextLonger`, `:126` `TestLand_AFailingGitStepKeepsItsOutputInTheLogAndNotInTheError`, `:218` `TestFailureLog_WhenTheACLCannotBeClearedTheOutputIsDroppedAndTheReasonSaysSo`, `:258` `TestFailureLog_ADirectoryThatIsALinkIsRefusedAndItsTargetIsLeftAlone`, `:299` `TestFailureLog_TheDirectoryKeepsOneTotalCapAndTheOldestFilesGoFirst`; `internal/agenticfactory/mergeq/faillog_temp_test.go:53` `TestFailureLog_AStaleLeftoverTempIsRemovedAndTheDirectoryEndsWithinTheCap`, `:88` `TestFailureLog_AFreshTempIsLeftAloneAndCountedInTheTotal`; `internal/agenticfactory/mergeq/faillog_acl_darwin_test.go:56` `TestFailureLog_TheLogAndItsDirectoryEndWithoutAnInheritedACL`; `internal/agenticfactory/mergeq/typed_errors_test.go:116` `TestLand_TheGateTextHeadsTheGateLogMaskedAndBounded`; `internal/cli/af_land_repair_test.go:44` `TestAFLand_TheGateTextIsNotInTheReasonTheStdoutOrTheLedgerButTheLogHoldsIt`, `:123` `TestAFLand_TheLandedNoteCarriesTheGateIDAndNotTheGateText`.
  - (g) [저장소 층, merged 사실] Given landed 행을 가진 카드, When (g1) 제출 mark와 정확히 같은 id의 green만 있을 때, (g2) 두 번 BUILDING으로 되돌아갔고 두 번째 되돌림 뒤 측정이 없을 때, (g3) 착지한 커밋에 대한 가장 새 리뷰가 실패일 때, (g4) 제출 뒤 다른 커밋에 대한 실패가 있었고 후보를 다시 재지 않았을 때, Then 카드는 merged로 읽히지 않아 DONE이 거절된다; (g5) 철회 전에 기록된 다른 커밋의 통과 기록은 아무것도 막지 않는다. 착지 시험(이름·줄만 확인, 돌리지 않음): `internal/agenticfactory/ledger/done_mark_boundary_test.go:17` `TestDone_AGreenWithExactlyTheIdOfTheOffersMarkIsOlderWork`, `:56` `TestDone_ACardTakenBackTwiceNeedsAMeasurementAfterTheSecondMoveBack`; `internal/agenticfactory/ledger/done_review_and_failures_test.go:72` `TestDone_AFailingReviewOfTheLandedCommitRefusesTheCard`, `:133` `TestDone_AFailureAboutAnotherCommitAfterTheOfferNeedsTheCandidateMeasuredAgain`, `:188` `TestDone_APassingRecordAboutAnotherCommitBeforeTheWithdrawalBlocksNothing`.
  - (h) [저장소 층 + 동사 층, 이전 착지의 이어받기] Given 이전 카드 epoch에서 landed가 된 행, When 현재 보유자가 같은 커밋을 다시 낼 때, Then 호출자가 `OnTrunk(true)`라고 말할 때만 행이 landed로 남아 현재 epoch로 옮겨지고, 아니면 `waiting`으로 다시 시작한다; `enqueue` 동사는 후보가 `refs/heads/af/trunk`의 조상인지 저장소에 묻고, 저장소가 답하지 못하면 아무것도 바꾸지 않고 거절한다. 착지 시험(이름·줄만 확인, 돌리지 않음): `internal/agenticfactory/ledger/enqueue_takeover_trunk_test.go:18` `TestEnqueue_ATakeOverAsksTheCallerWhetherTheCandidateIsOnTheTrunk`; `internal/cli/af_trunk_n39_test.go:62` `TestAFVerbs_N39ARowLandedOnAnotherTrunkIsNotTakenOverAsLanded`, `:93` `TestAFVerbs_N39ATrunkThatLostTheLandedCommitIsNotTakenOverAsLanded`, `:114` `TestAFVerbs_N39ALandedRowWhoseCommitIsInTheDedicatedTrunkIsTakenOver`, `:130` `TestAFVerbs_N39EnqueueRefusesWhenTheRepositoryCannotAnswerAndChangesNothing`(본문 `:130-145`를 읽음 — trunk ref를 지운 뒤 거절, 오류가 `af/trunk`를 이름 대고 원장 스냅숏 불변).
  - (i) [동사 층, DONE의 trunk 확인] Given 큐 행이 `landed`인 카드, When `moai af move`로 DONE에 옮길 때, Then 그 커밋이 `refs/heads/af/trunk`에 없거나 저장소가 답하지 못하면 아무것도 바꾸지 않고 거절한다. 착지 시험(돌리지 않음): `internal/cli/af_trunk_n39_test.go:171` `TestAFVerbs_N39TheMoveToDoneAsksWhetherTheLandedCommitIsInTheTrunk`(본문 `:171-195`를 읽음 — trunk를 한 커밋 되돌린 뒤 DONE 거절, 오류가 `af/trunk`와 커밋을 이름 대고 원장 스냅숏 불변, 대조로 커밋을 trunk에 되돌리면 DONE 수락), `:150` `TestAFVerbs_N39ACardRefusedDoneBecauseTheTrunkLostItsCommitCanBeLandedAgain`(이름·줄만). 저장소가 답하지 못하는 경우의 DONE 거절: no test found(이 작성자가 찾은 이름 가운데 없다; 코드 `af.go:187-211`에서만 읽음).
  - 죽여야 할 변이: 동사가 `--trunk`를 그대로 넘기는 구현 → (e)에서 FAIL; 관문 글이나 git 출력을 실패 사유에 싣는 구현 → (f)에서 FAIL; 시계로 "이동 뒤"를 정하는 구현 → (g1)·(g2)에서 FAIL; 다른 trunk에 착지한 행을 landed로 이어받는 구현 → (h)에서 FAIL; trunk를 묻지 않고 DONE으로 옮기는 동사 → (i)에서 FAIL.
- 시험 이름(하나): `TestMergeq_Land_CASUpdateRefThenLandedNote` — 소유 lane-22, 패키지 `internal/agenticfactory/mergeq`. unwritten at 37b67f6d39(§B).
- RED-now·green path(0.5.0): 이 AC는 release 게이트 밖이다 — §N 사후 관측 행 AC-AFL-014. RED 미관측. 완료 계산에 들지 않는다(§D). 시험 이름은 위 한 개 그대로다.
- 이전 문구(0.4.0, RED-now·green path): «name absent at 37b67f6d39 — `TestMergeq_Land_CASUpdateRefThenLandedNote`가 `37b67f6d39`에 없다(§B). integration tip `1a552af405`(조정자가 적은 `915a6524cb`에서 움직인 tip)에서 이 이름은 없다. 착지 시험이 이 AC를 덮는 범위: partly — CAS 착지·`landed` 쪽지 하나·잠금 밖 관문·checkout 거절·잠금 epoch 거절은 여러 시험이 나눠 덮고, 카드의 DONE 이동은 mergeq 시험이 가짜 큐로 돌아 덮지 않는다(§F 행 AC-AFL-014). 착지 전 RED는 관측되지 않았고 이 칸은 판정한 RED가 아니다 — 통과로 적지 않는다.» «RED evidence not recoverable from history (landed before this SPEC); this AC adopts the landed test as a regression-guard, not a release-blocking RED (`verification-completeness.md` §2.1 undecidable disposition). 남은 일: §F에 적은 덮이지 않은 부분을 이 이름의 시험(소유 lane-22)이 더한다. 그 시험은 이미 착지한 코드 위에서 쓰이므로 처음부터 초록일 수 있고, RED가 아니라 regression-guard로 다룬다.»
- 이전 문구(0.3.0, RED-now·green path): «`TestMergeq_Land_CASUpdateRefThenLandedNote`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.» «lane-22의 mergeq 구현 커밋이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.»

#### AC-AFL-015 — Issue·원격 코드 퇴역

- 주 REQ: maps REQ-AFL-015.
- Given 퇴역 변경이 착지한 트리, When `internal/` 아래의 `_test.go`가 아닌 `.go` 파일의 import와 `moai` 명령 등록을 훑을 때, Then `forge`·`mirror`·`gitref`·`effect/ghintake`·`effect/ctxledger`·`effect/labels`·`effect/ghexec`·`effect/ghidentity`·`effect/gitexec`를 가져오는 파일이 없고, `moai factory resident` 하위 명령이 등록되어 있지 않으며, 같은 트리에 `ledger`·`mergeq`·`claim`·`core`·`runner` 패키지는 남아 있다(양성 대조: 범위 밖 패키지가 지워지는 구현을 잡음).
- 계약 문장(고정한 계약 §10 P2 retirement plan, revision 항목): "To avoid a cross-lane wait, lane-20 owns the whole remote retirement in one change: `moai factory resident` (internal/cli/factory_resident.go, its test, its registration), resident, monitor/notify remote parts (ghpr/Inbox/forge cursors), forge (+webhook), mirror, gitref, effect/{ghintake, ctxledger,labels,ghexec,ghidentity,gitexec}, docstore cursors/queue, and e2e files whose only subject is those packages." (OD-42 leader decision (leader-decisions-od36-42-20261030.md, 2026-10-30T05:56Z; reversible))
- 판정하지 않는 것: `effect` 자체와 `docstore.Store`의 퇴역(계약이 "unless reconcile or monitor/adapt has a non-remote use"라는 조건을 달았다), `identity`의 GitHub 조회와 `intake`의 GitHub 검색(§10에 없어 미정). §E에 적었다.
- 같은 시험의 사례(0.4.0, N8): (b) Then 퇴역 대상 패키지 디렉터리 `internal/agenticfactory/forge`·`mirror`·`gitref`·`resident`·`effect/ghintake`·`effect/ctxledger`·`effect/labels`·`effect/ghexec`·`effect/ghidentity`·`effect/gitexec`가 트리에 없다 — 어디서도 import되지 않더라도 디렉터리가 남아 있으면 FAIL(import 검사만으로는 이 구현을 잡지 못한다); (c) `resident` 패키지(`internal/agenticfactory/resident`, `37b67f6d39`에 있던 경로)도 (a)의 import 검사와 (b)의 부재 검사 대상에 든다.
- 죽여야 할 변이: 패키지 하나라도 남기는 구현, `factory resident`를 등록해 두는 구현 → FAIL; import는 모두 걷어냈으나 퇴역 대상 디렉터리를 남겨 둔 구현 → (b)에서 FAIL; 범위 밖 `ledger`·`mergeq`·`claim`·`core`·`runner`를 지우는 구현 → 양성 대조에서 FAIL.
- 이전 문구(0.3.0, 변이 줄): «패키지 하나라도 남기는 구현, `factory resident`를 등록해 두는 구현 → FAIL; 범위 밖 `ledger`·`mergeq`·`claim`·`core`·`runner`를 지우는 구현 → 양성 대조에서 FAIL.» 사례 (b)(c)는 없었다.
- 시험 이름(하나): `TestRetirement_NoImportOfForgeMirrorGitref` — 소유 lane-20, 패키지 미정(OD-35 (h)). unwritten at 37b67f6d39(§B).
- RED 기록(0.5.1, release 게이트, 감사 N1·N3·N4 — lane report 기록이며 이 작성자가 돌린 것이 아니다): 출처는 lane-20의 lane report `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/agentic-factory-integration/.moai/reports/issue-1948/af-retirement-ac015.md`. 시험은 `internal/agenticfactory/retirement/retirement_test.go`의 `TestRetirement_NoImportOfForgeMirrorGitref`이며 `MOAI_RETIREMENT_ROOT=<dir>`로 다른 checkout을 가리킬 수 있다. 측정 트리: `8d0c47a7f3`(퇴역 커밋 `3fd9ab6456`의 부모 — `git -C <worktree> log -1 --format='%h %p %s' 3fd9ab6456`의 출력 `3fd9ab6456 8d0c47a7f3 [issue-1948] refactor(agenticfactory): retire the Issue/remote code under the local-only design (P2, OD-42)`로 이 작성자가 확인). lane report가 적은 기준: «`git archive 8d0c47a7f3 go.mod go.sum cmd internal/{agenticfactory,cli,homestate,config,hook}` extracted to a scratch dir, test file copied in». 명령: `MOAI_RETIREMENT_ROOT=<scratch> go test -p 4 -count=1 ./internal/agenticfactory/retirement/`. lane report가 옮긴 줄: `--- FAIL: TestRetirement_NoImportOfForgeMirrorGitref (1.38s)`, "32 failure lines", 예: `effect/ghexec imports the retired package .../agenticfactory/forge`, `resident imports the retired package .../effect/ctxledger`, `the retired directory internal/agenticfactory/forge still exists`. 종료 코드: exit code not recorded as a field; non-zero implied by the FAIL summary line only. lane report 자신의 Gap(그대로): «only the first lines were read in full; the factory resident file case was exercised by the 32-line total, not read line by line.» GREEN 기록(lane report — 이 작성자가 관측하지 않음): 트리 «HEAD of branch WT-agentic-monitor-effect-reconcile (integration 03490b9b8f merged as 43759b02fb, plus this commit)», 명령 `go test -p 4 -count=1 -v ./internal/agenticfactory/retirement/`, 줄 `--- PASS: TestRetirement_NoImportOfForgeMirrorGitref (2.00s)`, `ok ... 2.232s`. 변이(lane report): 빈 퇴역 디렉터리 `mirror` → FAIL `the retired directory internal/agenticfactory/mirror still exists`; `ledger`를 지운 사본 → FAIL `the package .../agenticfactory/ledger must remain after the retirement`; `newFactoryResidentCommand`를 다시 등록하는 트리는 거치지 않았다. 착지(이 작성자가 확인): 시험 커밋 `d649512ed7`, integration 병합 `b28a31d9dc`(`merge(lane-20): AC-AFL-015 retirement import test (d649512ed7, Issue #1948)`); `git -C <worktree> merge-base --is-ancestor d649512ed7 b28a31d9dc` → `exit=0`; 두 SHA 모두 `git -C <worktree> branch --contains <sha> --list WT-agentic-factory-integration` → `+ WT-agentic-factory-integration`; `b28a31d9dc`에 `internal/agenticfactory/retirement/retirement_test.go:65`.
- 기록한 이탈(DEVIATION, 0.5.1, 감사 N4): 이 AC의 RED는 커밋 그래프로 시험-먼저 순서를 증언하지 않는다. 퇴역 변경 `3fd9ab6456`이 시험 커밋 `d649512ed7`보다 먼저 착지했고, 시험을 RED로 싣고 GREEN 커밋보다 앞선 커밋은 그래프에 없다. 대신 한 일: 퇴역 전 부모 트리 `8d0c47a7f3`의 일부(PARTIAL) `git archive`를 scratch 디렉터리에 풀고 시험 파일을 복사해 넣어 그래프 밖에서 한 번 돌렸다(lane-20). 이 실행이 증언하지 못하는 것: (1) 작성 순서(시험이 구현보다 먼저 쓰였는지) — `verification-claim-integrity.md` §2.3에 따라 순서는 커밋 그래프만 증언한다; (2) 모듈 전체 위에서의 실행 — archive에 든 경로는 `go.mod go.sum cmd internal/{agenticfactory,cli,homestate,config,hook}`뿐이다; (3) 빨강이 사례 (b)(퇴역 디렉터리 잔존)와 (a)(import)에서만 왔다는 것 — 아래 N3. 이 이탈은 0.5.0의 "§2.3에 따라 순서 조항을 두 관측으로 바꿔 적는다"는 문장을 대신한다: 두 관측은 이제 기록으로 있으나, 그것은 순서의 증언이 아니라 기록된 이탈이다. 이 기록을 통과로 적지 않는다.
- N3 상태(0.5.1): not resolved by the report. 감사 N3은 빨강의 원인이 둘일 수 있다고 적었다 — 양성 대조(`ledger`·`mergeq`·`claim`·`core`·`runner`가 남아 있음, import 검사가 간선을 봄)가 그 패키지가 없는 트리에서도 실패하기 때문이다. lane report는 32줄 가운데 첫 줄들만 읽었다고 적었으므로, 32줄 안에 양성 대조 실패가 따로 있는지는 lane report에서 읽을 수 없다. 이 작성자가 이 작업에서 읽은 트리 사실(부분 증거, 디렉터리 쪽만): `git -C <worktree> ls-tree -d --name-only 8d0c47a7f3 internal/agenticfactory/{ledger,mergeq,claim,core,runner,forge,mirror,gitref,resident}`(실제 호출은 경로 9개를 나열한 한 번의 명령) → `claim`·`core`·`forge`·`gitref`·`ledger`·`mergeq`·`mirror`·`resident`·`runner`가 모두 나왔고 `exit=0`. 그래서 `8d0c47a7f3` 트리에는 대조 디렉터리 다섯이 있다. 그러나 (i) scratch 사본은 부분 archive이고, (ii) import 간선 대조가 그 사본에서 어떻게 나왔는지는 기록되지 않았으므로, N3을 풀린 것으로 적지 않는다.
- 이전 문구(0.5.0, RED-now): «RED-now(0.5.0, release 게이트): 시험이 쓰이지 않은 상태다 — `TestRetirement_NoImportOfForgeMirrorGitref`가 `37b67f6d39`(§B)에도 tip `db02401ad1`(§B.2 [L-015d])에도 없다. 그 시험이 단언할 상태는 `37b67f6d39`에서 거짓이다: 사례 (b)의 퇴역 대상 디렉터리 열 개(`forge`·`mirror`·`gitref`·`resident`·`effect/ghintake`·`effect/ctxledger`·`effect/labels`·`effect/ghexec`·`effect/ghidentity`·`effect/gitexec`)가 그 트리에 있고(§B.2 [L-015a]), `moai factory resident`의 파일 `internal/cli/factory_resident.go`가 있다([L-015c]). 그래서 이 AC의 시험을 `37b67f6d39` 트리에서 돌리면 (b)에서 FAIL해야 한다. 이 문장은 트리를 읽어 적은 것이며 시험을 돌려 본 것이 아니다. 빨강의 이유는 이 작업(퇴역)이 지울 디렉터리가 남아 있다는 것이며, 손대지 않는 파일 때문이 아니다. RED 미관측.»
- green path(0.5.1): 완료 계산(§D)에는 위 RED 기록(기록한 이탈 포함)과 GREEN 기록이 함께 있어야 한다 — 둘 다 lane report 기록으로 있다. 순서 조항은 커밋 그래프가 증언하지 못하며, 위 DEVIATION으로 적었다. 남은 일: N3(양성 대조의 원인 분리)은 열려 있다.
- 이전 문구(0.5.0, green path): «lane-20이 이 이름의 시험을 쓴다. 완료로 세려면 (1) 그 시험을 `37b67f6d39` 트리 위에서 돌려 (b)가 FAIL하는 출력과 (2) integration tip에서 통과하는 출력을 각 tree SHA와 함께 기록해야 한다. 퇴역 변경 자체는 이미 착지했다: tip `db02401ad1`에서 같은 열 디렉터리가 없다([L-015b]). RED 커밋이 GREEN 커밋보다 먼저라는 순서는 퇴역이 이미 착지해 커밋 그래프로 얻을 수 없으므로, 이 AC의 순서 조항은 (1)(2)의 두 관측으로 바꿔 적는다(`verification-claim-integrity.md` §2.3: 순서 조항을 커밋 그래프가 증언할 수 있는 것으로 고쳐 적음).»
- 이전 문구(0.4.0, RED-now·green path): «name absent at 37b67f6d39 — `TestRetirement_NoImportOfForgeMirrorGitref`가 `37b67f6d39`에 없다(§B). integration tip `1a552af405`(조정자가 적은 `915a6524cb`에서 움직인 tip)에서 이 이름은 없다. 착지 시험이 이 AC를 덮는 범위: 착지 시험 없음(no landed test) — 퇴역 변경 자체는 `1a552af405`(merge(lane-20) 3fd9ab6456)로 착지했다(§F 행 AC-AFL-015). 착지 전 RED는 관측되지 않았고 이 칸은 판정한 RED가 아니다 — 통과로 적지 않는다.» «RED evidence not recoverable from history (landed before this SPEC); this AC adopts the landed test as a regression-guard, not a release-blocking RED (`verification-completeness.md` §2.1 undecidable disposition). 남은 일: §F에 적은 덮이지 않은 부분을 이 이름의 시험(소유 lane-20)이 더한다. 그 시험은 이미 착지한 코드 위에서 쓰이므로 처음부터 초록일 수 있고, RED가 아니라 regression-guard로 다룬다. 참고 사실(시험 아님): `1a552af405`에 `forge`·`mirror`·`gitref`·`resident`·`effect/{ghintake,ctxledger,labels,ghexec,ghidentity,gitexec,docstore}` 디렉터리가 없다(progress §K).»
- 이전 문구(0.3.0, RED-now·green path): «`TestRetirement_NoImportOfForgeMirrorGitref`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.» «lane-20의 퇴역 변경 하나(계약 §10: 게이트 integration `go build ./... && go vet ./...`)가 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.»
- 이전 문구(0.1.0): Given 퇴역이 끝난 트리, When AgenticFactory 패키지 그래프를 훑을 때, Then `forge`·`mirror`·`effect/ghintake`·`effect/ctxledger`·`effect/labels`·`gitref`를 가져오는 패키지가 없다. 소유 lane-22, green path "lane-22의 퇴역 커밋(패키지마다 한 커밋; claim·controlref는 lane-18)".

#### AC-AFL-016 — importer는 원천을 읽기만 함

- 주 REQ: maps REQ-AFL-016.
- Given backlog.db·factory.db·claim JSON 문서의 시험 사본, When importer를 dry-run으로 돌릴 때, Then 세 원천의 바이트가 실행 전과 같다.
- 시험 이름(하나): `TestImporter_DryRun_SourcesUnchanged` — 소유 lane-22, 패키지 `internal/agenticfactory/ledger/importer`. unwritten at 37b67f6d39(§B).
- RED-now·green path(0.5.0): 이 AC는 release 게이트 밖이다 — §N 사후 관측 행 AC-AFL-016. RED 미관측. 완료 계산에 들지 않는다(§D). 시험 이름은 위 한 개 그대로다.
- 이전 문구(0.4.0, RED-now·green path): «name absent at 37b67f6d39 — `TestImporter_DryRun_SourcesUnchanged`가 `37b67f6d39`에 없다(§B). integration tip `1a552af405`(조정자가 적은 `915a6524cb`에서 움직인 tip)에서 이 이름은 없다. 착지 시험이 이 AC를 덮는 범위: partly — backlog.db·factory.db 두 원천은 덮고, claim JSON 원천은 덮지 않는다(§F 행 AC-AFL-016). 착지 전 RED는 관측되지 않았고 이 칸은 판정한 RED가 아니다 — 통과로 적지 않는다.» «RED evidence not recoverable from history (landed before this SPEC); this AC adopts the landed test as a regression-guard, not a release-blocking RED (`verification-completeness.md` §2.1 undecidable disposition). 남은 일: §F에 적은 덮이지 않은 부분을 이 이름의 시험(소유 lane-22)이 더한다. 그 시험은 이미 착지한 코드 위에서 쓰이므로 처음부터 초록일 수 있고, RED가 아니라 regression-guard로 다룬다.»
- 이전 문구(0.3.0, RED-now·green path): «`TestImporter_DryRun_SourcesUnchanged`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.» «lane-22의 importer dry-run 커밋이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.»

#### AC-AFL-017 — `moai af` 동사

- 주 REQ: maps REQ-AFL-017.
- Given `moai af` 명령, When 등록된 하위 동사를 볼 때, Then claim, 단계 이동, 증거 기록, `Enqueue`, `Release` 각각에 닿는 동사가 있다(동사 이름은 미정, OD-33).
- 시험 이름(하나): `TestAFCommand_VerbsRegistered` — 소유 lane-22, 패키지 미정(CLI 패키지, OD-33). unwritten at 37b67f6d39(§B). (0.5.0: tip `db02401ad1`에서도 없다, §B.2 [L-017]; `git -C <worktree> grep -n -e 'Use: *"af' -e 'Use:  *"af' db02401ad1 -- internal/cli`의 출력도 없었다 — 종료 코드 관측 못 함.)
- RED-now(0.5.1, release 게이트, 감사 N1 — lane report 기록이며 이 작성자가 돌린 것이 아니다): 출처는 lane-22의 lane report `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/agentic-factory-integration/.moai/reports/issue-1948/af-verbs-ac017-ac013.md`. 측정 커밋: `5a9370f894`(RED 커밋), 기준 `03490b9b8f`. 명령: lane report af-verbs-ac017-ac013.md L20(이 SPEC에는 실행할 수 있는 형태로 옮기지 않는다). lane report가 옮긴 출력(그대로):
  ```
  === RUN   TestAFCommand_VerbsRegistered
      af_test.go:31: `moai af` is not registered on the root command
  --- FAIL: TestAFCommand_VerbsRegistered (0.00s)
  FAIL
  FAIL	github.com/modu-ai/moai-adk/internal/cli	1.524s
  ```
  종료 코드: exit code not recorded as a field; non-zero implied by the FAIL summary line only. 명령 형식에 대한 발견(기록만 함, 여기서 고치지 않음 — `verification-completeness.md` §1.1): lane report가 옮긴 명령은 앵커 없는 `-run` 패턴(`^…$` 없이 `-run TestName`)을 쓴다. 그래서 고른 시험이 하나도 없는 실행도 `ok`를 찍었을 것이다. lane report 출력의 `=== RUN` 줄 하나가 의도한 시험이 돌았다는 유일한 표시이며, 이 작성자는 그 명령을 돌리지 않았다. 이전 문구(0.5.1 첫 쓰기): «고른 시험의 수: 명령의 `-run` 인자는 앵커가 없어 더 긴 이름의 시험도 고를 수 있다(spec lint `VacuousTestAssertion` 경고가 이 줄을 가리킨다; 명령은 lane report 그대로 두었다). lane report 출력에 `=== RUN` 줄이 하나뿐이므로 그 실행이 고른 시험은 이 이름 하나로 읽힌다 — 이 작성자가 돌려 확인한 것은 아니다.» 빨강의 이유: `moai af`가 root 명령에 등록되어 있지 않다 — 이 작업이 더할 명령의 부재다. 커밋 그래프 순서(이 작성자가 이 작업에서 돌림): `git -C <worktree> merge-base --is-ancestor 5a9370f894 b58b1566b6` → `exit=0`; `git -C <worktree> merge-base --is-ancestor 5a9370f894 95e100e754` → `exit=0`; 반대 방향 대조 `git -C <worktree> merge-base --is-ancestor b58b1566b6 5a9370f894` → `exit=1`. `git -C <worktree> branch --contains 5a9370f894 --list WT-agentic-factory-integration` → `+ WT-agentic-factory-integration`. 시험 위치: `5a9370f894`에 `internal/cli/af_test.go:23`이 있다(`git grep -n`으로 확인).
- GREEN 기록(0.5.1, lane report — 이 작성자가 관측하지 않음): 커밋 `b58b1566b6`, 명령: lane report af-verbs-ac017-ac013.md L50, lane report가 옮긴 줄: `--- PASS: TestAFCommand_VerbsRegistered (0.00s)` / `PASS` / `ok  	github.com/modu-ai/moai-adk/internal/cli	1.232s`. lane report 자신의 Gaps: 동사 이름(`claim`·`move`·`evidence`·`enqueue`·`release`)은 lane이 고른 것이며 OD-33은 열려 있다.
- 이전 문구(0.5.0, RED-now): «RED-now(0.5.0, release 게이트): 시험이 쓰이지 않은 상태(unwritten test) — `TestAFCommand_VerbsRegistered`가 `37b67f6d39`에 없고(§B, 0.2.0 측정의 종료 코드 1) tip `db02401ad1`에도 없다(§B.2 [L-017]). 시험이 없으므로 실행 명령을 두지 않는다(§A). RED 미관측.»
- 이전 문구(0.4.0, RED-now): «name absent at 37b67f6d39 — `TestAFCommand_VerbsRegistered`가 `37b67f6d39`에 없다(§B). integration tip `1a552af405`(조정자가 적은 `915a6524cb`에서 움직인 tip)에서 이 이름은 없다. 착지 시험이 이 AC를 덮는 범위: 착지 시험 없음(no landed test)(§F 행 AC-AFL-017). 착지 전 RED는 관측되지 않았고 이 칸은 판정한 RED가 아니다 — 통과로 적지 않는다.»
- green path: lane-22의 CLI 커밋이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다. (0.4.0: 이 AC의 코드는 `1a552af405`에 착지하지 않았으므로 RED 커밋이 GREEN보다 먼저라는 경로가 그대로 남는다.)
- 이전 문구(0.3.0, RED-now·green path): «`TestAFCommand_VerbsRegistered`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.» «lane-22의 CLI 커밋이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.»

#### AC-AFL-018 — §6 기본값은 이름 있는 상수

- 주 REQ: maps REQ-AFL-018.
- Given 원장 패키지, When 그 상수를 읽을 때, Then lease 15 min, renew 5 min, reclaim cap 3, snapshot 30 min이 이름 있는 상수이고 값이 계약 §6과 같다(원장 밖 상수는 OD-35).
- 시험 이름(하나): `TestLedgerDefaults_MatchContract` — 소유 lane-18, 패키지 `internal/agenticfactory/ledger`. unwritten at 37b67f6d39(§B).
- RED-now·green path(0.5.0): 이 AC는 release 게이트 밖이다 — §N 사후 관측 행 AC-AFL-018. RED 미관측. 완료 계산에 들지 않는다(§D). 시험 이름은 위 한 개 그대로다.
- 이전 문구(0.4.0, RED-now·green path): «name absent at 37b67f6d39 — `TestLedgerDefaults_MatchContract`가 `37b67f6d39`에 없다(§B). integration tip `1a552af405`(조정자가 적은 `915a6524cb`에서 움직인 tip)에서 이 이름은 없다. 착지 시험이 이 AC를 덮는 범위: not — 상수는 `1a552af405:internal/agenticfactory/ledger/api.go:40-50`에 있으나 값을 단언하는 시험이 없다(§F 행 AC-AFL-018). 착지 전 RED는 관측되지 않았고 이 칸은 판정한 RED가 아니다 — 통과로 적지 않는다.» «RED evidence not recoverable from history (landed before this SPEC); this AC adopts the landed test as a regression-guard, not a release-blocking RED (`verification-completeness.md` §2.1 undecidable disposition). 남은 일: §F에 적은 덮이지 않은 부분을 이 이름의 시험(소유 lane-18)이 더한다. 그 시험은 이미 착지한 코드 위에서 쓰이므로 처음부터 초록일 수 있고, RED가 아니라 regression-guard로 다룬다.»
- 이전 문구(0.3.0, RED-now·green path): «`TestLedgerDefaults_MatchContract`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.» «lane-18의 원장 구현 커밋이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.»

#### AC-AFL-019 — 흐름 읽기 모형은 Reader만 씀

- 주 REQ: maps REQ-AFL-019.
- Given 멈춘 lease 하나를 담은 가짜 `ledger.Reader`, When 흐름 모형을 계산할 때, Then 그 카드가 stalled로 잡히고, 계산은 그 Reader 말고 다른 입력을 쓰지 않는다.
- 시험 이름(하나): `TestFlow_StalledDetection_FromReader` — 소유 lane-20, 패키지 `internal/agenticfactory/monitor/flow`. unwritten at 37b67f6d39(§B).
- RED-now·green path(0.5.0): 이 AC는 release 게이트 밖이다 — §N 사후 관측 행 AC-AFL-019. RED 미관측. 완료 계산에 들지 않는다(§D). 시험 이름은 위 한 개 그대로다.
- 이전 문구(0.4.0, RED-now·green path): «name absent at 37b67f6d39 — `TestFlow_StalledDetection_FromReader`가 `37b67f6d39`에 없다(§B). integration tip `1a552af405`(조정자가 적은 `915a6524cb`에서 움직인 tip)에서 이 이름은 없다. 착지 시험이 이 AC를 덮는 범위: partly — stalled 규칙은 덮고, "그 Reader 말고 다른 입력을 쓰지 않는다"는 덮지 않는다(§E 0.4.0)(§F 행 AC-AFL-019). 착지 전 RED는 관측되지 않았고 이 칸은 판정한 RED가 아니다 — 통과로 적지 않는다.» «RED evidence not recoverable from history (landed before this SPEC); this AC adopts the landed test as a regression-guard, not a release-blocking RED (`verification-completeness.md` §2.1 undecidable disposition). 남은 일: §F에 적은 덮이지 않은 부분을 이 이름의 시험(소유 lane-20)이 더한다. 그 시험은 이미 착지한 코드 위에서 쓰이므로 처음부터 초록일 수 있고, RED가 아니라 regression-guard로 다룬다.»
- 이전 문구(0.3.0, RED-now·green path): «`TestFlow_StalledDetection_FromReader`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.» «lane-20의 flow 패키지 커밋이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.»

#### AC-AFL-020 — 보드는 읽기 전용

- 주 REQ: maps REQ-AFL-020.
- Given 카드·lease·queue·쪽지마다 서로 다른 표지값을 담고, lease 하나가 `Keys []string`에 표지 키 둘을 싣는 가짜 `ledger.Reader`, When (i) 보드 페이지를 그릴 때, Then 카드·lease·키·queue·쪽지 다섯 범주의 표지값이 모두 페이지에 나오고, 키 표지값은 그 `Lease.Keys`에서 온 것이다; 그리고 When (ii) 같은 시험이 `internal/web` 디렉터리(하위 포함)의 `_test.go`가 아닌 `.go` 파일을 `go/parser`로 읽을 때, Then `ledger.Store` 선택자와 `ledger.Open` 호출이 0건이다(OD-39 leader decision (leader-decisions-od36-42-20261030.md, 2026-10-30T05:56Z; reversible)).
- 이전 문구(0.2.0): Given 카드·lease·queue·쪽지마다 서로 다른 표지값을 담은 가짜 `ledger.Reader`, When (i) 보드 페이지를 그릴 때, Then 네 범주의 표지값이 모두 페이지에 나온다; 그리고 When (ii) 같은 시험이 `internal/web` 디렉터리(하위 포함)의 `_test.go`가 아닌 `.go` 파일을 `go/parser`로 읽을 때, Then `ledger.Store` 선택자와 `ledger.Open` 호출이 0건이다.
- 판정하지 않는 것: 흐름 지표 표시. §E에 적었다. (이전 문구(0.2.0): "lease 키(출처 미정, OD-39)와 흐름 지표 표시")
- 죽여야 할 변이: 보드가 따로 `*ledger.Store`를 열어 읽거나 쓰는 구현 → (ii)에서 FAIL; Reader 값을 쓰지 않고 고정 내용을 그리는 구현 → (i)에서 FAIL.
- 시험 이름(하나): `TestBoard_ReaderOnly_NoStoreReference` — 소유 lane-21, 패키지 `internal/web`. unwritten at 37b67f6d39(§B).
- RED-now·green path(0.5.0): 이 AC는 release 게이트 밖이다 — §N 사후 관측 행 AC-AFL-020. RED 미관측. 완료 계산에 들지 않는다(§D). 시험 이름은 위 한 개 그대로다.
- 이전 문구(0.4.0, RED-now·green path): «name absent at 37b67f6d39 — `TestBoard_ReaderOnly_NoStoreReference`가 `37b67f6d39`에 없다(§B). integration tip `1a552af405`(조정자가 적은 `915a6524cb`에서 움직인 tip)에서 이 이름은 없다. 착지 시험이 이 AC를 덮는 범위: partly — 보기 전용 경로·키 표시는 덮고, (ii) 정적 `ledger.Store`·`ledger.Open` 0건은 덮지 않는다(§F 행 AC-AFL-020). 착지 전 RED는 관측되지 않았고 이 칸은 판정한 RED가 아니다 — 통과로 적지 않는다.» «RED evidence not recoverable from history (landed before this SPEC); this AC adopts the landed test as a regression-guard, not a release-blocking RED (`verification-completeness.md` §2.1 undecidable disposition). 남은 일: §F에 적은 덮이지 않은 부분을 이 이름의 시험(소유 lane-21)이 더한다. 그 시험은 이미 착지한 코드 위에서 쓰이므로 처음부터 초록일 수 있고, RED가 아니라 regression-guard로 다룬다.»
- 이전 문구(0.3.0, RED-now·green path): «`TestBoard_ReaderOnly_NoStoreReference`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.» «lane-21의 보드 커밋이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.»
- 이전 문구(0.1.0): Given 카드·lease·키·queue·쪽지를 담은 가짜 `ledger.Reader`, When 보드 페이지를 그릴 때, Then 여섯 항목(카드, lease, 키, queue, 쪽지, 흐름 지표)이 보이고 Reader에 쓰기 호출이 하나도 없다. 시험 이름 `TestBoard_FakeReader_RendersReadOnly`.

#### AC-AFL-021 — 훅 heartbeat 30 s 조절

- 주 REQ: maps REQ-AFL-021.
- Given 만든 지 30 s가 안 된 세션 stamp 파일, When 훅이 heartbeat를 쓰려 할 때, Then DB 쓰기가 일어나지 않는다.
- 시험 이름(하나): `TestLedgerHeartbeat_ThrottlesToTheStampFile` — 소유 lane-20(계약 §4 표; 착지 커밋을 쓴 lane은 확인하지 않음), 패키지 `internal/hook`(착지 위치 `1a552af405:internal/hook/ledger_heartbeat_test.go:58`). unwritten at 37b67f6d39(§B). 채택 근거(시험 본문 인용): 첫 호출 뒤 `*now = now.Add(LedgerHeartbeatEvery - time.Second)`에서 다시 부르고 `"a stamp younger than %v must skip the DB write"`로 호출 수 1을 단언하며, `+2s` 뒤에는 `"a stamp older than %v must beat again"`으로 2를 단언한다; `LedgerHeartbeatEvery = 30 * time.Second`(`1a552af405:internal/hook/ledger_heartbeat.go:20`). 패키지 위치는 착지 사실로 적었을 뿐 상위 OD-35 (k) 행은 닫지 않았다. (0.5.0: 채택 근거였던 'fully' 판정은 틀렸다 — 본문은 30 s를 직접 고정하지 않고 stamp 파일을 단언하지 않는다, §F 행 AC-AFL-021 partly. 이름은 바꾸지 않았다; 이 AC는 §N에 있어 완료 계산에 들지 않는다. 이름을 되돌릴지는 이 판에서 정하지 않았다.)
- 이전 문구(0.3.0, 시험 이름): «`TestHeartbeatWriter_SkipsWriteWhenStampFresh` — 소유 lane-20, 패키지 미정(훅 패키지, 계약은 'hook heartbeat writer'라고만 적음; OD-35 (k)).»
- RED-now·green path(0.5.0): 이 AC는 release 게이트 밖이다 — §N 사후 관측 행 AC-AFL-021. RED 미관측. 완료 계산에 들지 않는다(§D). 시험 이름은 위 한 개 그대로다.
- 이전 문구(0.4.0, RED-now·green path): «name absent at 37b67f6d39 — `TestLedgerHeartbeat_ThrottlesToTheStampFile`가 `37b67f6d39`에 없다(§B). integration tip `1a552af405`(조정자가 적은 `915a6524cb`에서 움직인 tip)에서 채택한 이름이 `internal/hook/ledger_heartbeat_test.go:58`에 있다. 착지 시험이 이 AC를 덮는 범위: fully — 채택한 이름 그 자체(§F 행 AC-AFL-021). 착지 전 RED는 관측되지 않았고 이 칸은 판정한 RED가 아니다 — 통과로 적지 않는다.» «RED evidence not recoverable from history (landed before this SPEC); this AC adopts the landed test as a regression-guard, not a release-blocking RED (`verification-completeness.md` §2.1 undecidable disposition). 채택한 시험이 이미 tip에 있고 regression-guard로 돈다.»
- 이전 문구(0.3.0, RED-now·green path): «`TestHeartbeatWriter_SkipsWriteWhenStampFresh`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.» «lane-20의 훅 heartbeat 커밋(계약 §7 P2, api.go 착지 뒤)이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.»

## §D 완료 정의

- (0.5.1, 감사 N1·N4) 완료 계산에 드는 것은 release 게이트 AC 셋(AC-AFL-013·015·017)뿐이다. 한 AC는 다음 둘이 모두 있을 때에만 완료로 센다: (1) 기록된 RED — 명령, 실패 줄, 측정 커밋 또는 트리, 그리고 순서(013·017: 커밋 그래프로 RED 커밋이 GREEN 커밋의 조상임이 확인됨; 015: 커밋 그래프가 순서를 증언하지 못하므로 기록한 이탈(DEVIATION)로 적힌 그래프 밖 실행); (2) GREEN 기록 — 명령, 통과 줄, 커밋 또는 트리. 지금 세 AC 모두 (1)(2)가 lane report 기록으로 있으며, 이 작성자는 그 시험을 돌리지 않았다. 종료 코드는 세 기록 모두 따로 적힌 칸이 없다(exit code not recorded as a field; non-zero implied by the FAIL summary line only). AC-AFL-015는 N3이 풀리지 않았고 순서가 이탈로 기록되었다. 이 문장은 판정이 아니다: 코드 판정은 고정한 integration SHA 위의 마지막 sync 감사가 한다(progress §M). §N의 18개 AC는 완료 계산에 들지 않으며, 그 시험의 통과는 release 판정이 아니다.
- 이전 문구(0.5.0): «(0.5.0) 완료 계산에 드는 것은 release 게이트 AC 셋(AC-AFL-013·015·017)뿐이다. 각각은 먼저 RED-now 관측이 있거나 시험이 쓰이지 않은 상태가 명시되어 있어야 한다 — 지금 셋 모두 '쓰이지 않은 시험' 상태이며, AC-AFL-015는 `37b67f6d39`에서 거짓인 상태를 트리를 읽어 적었다(관측한 RED는 없다). 그 뒤 그 시험이 소유 lane의 커밋으로 존재하고, RED 출력(013·017: RED 커밋이 GREEN 커밋보다 먼저 — `verification-claim-integrity.md` §2.3; 015: `37b67f6d39` 트리 위 실행의 FAIL 출력)과 GREEN 출력이 각 tree SHA와 함께 기록될 때 그 AC를 완료로 센다. §N의 18개 AC는 완료 계산에 들지 않으며, 그 시험의 통과는 release 판정이 아니다.
- 이전 문구(0.4.0): «(0.4.0) 21개 시험이 각 소유 lane의 커밋으로 존재하고 integration tip에 들어 있으며, 각 시험이 실행되어 통과한 출력이 그 tree SHA와 함께 기록된다. RED 커밋이 GREEN 커밋보다 먼저 있어야 한다는 조건(`verification-claim-integrity.md` §2.3: 순서는 커밋 그래프만 증언한다)은 코드가 아직 착지하지 않은 AC-AFL-013·017에만 남는다. 나머지 19개는 regression-guard다(§A 0.4.0, `verification-completeness.md` §2.1) — 이 19개의 통과는 release-blocking 판정이 아니다.»
- 이전 문구(0.3.0): 21개 시험이 각 소유 lane의 커밋으로 존재하고, RED 커밋이 GREEN 커밋보다 먼저 있으며(`verification-claim-integrity.md` §2.3: 순서는 커밋 그래프만 증언한다), GREEN 뒤 각 시험이 실행되어 통과한 출력이 그 tree SHA와 함께 기록된다.
- 실행 명령은 시험이 생긴 뒤에 그 이름만 고르는 형태로 적는다. 그 전에는 적지 않는다.
- (0.6.0) AC-AFL-001·007·008·009·014에 더한 "0.6.0 사례"는 모두 §N의 AC 위에 있으므로 release 게이트 밖이며 완료 계산에 들지 않는다. 그 사례 줄에 적은 착지 시험은 이 작성자가 돌리지 않았고, 통과로 적지 않는다. 위 첫 줄의 완료 계산(AC-AFL-013·015·017)은 이 판에서 바뀌지 않는다.

## §E 덮지 못하는 부분

- (0.6.0) REQ-AFL-001의 저장소 묶음: 원장이 아무것도 모를 때(묶이지 않았고, 행이 이름 대는 커밋도 없고, `--db` 원장 파일 자리도 git 저장소가 아닐 때 — 2026-11-01 정정: 같은 저장소의 lease worktree는 단서가 아니므로 lease worktree가 있어도 같다) `--db` 원장의 첫 호출은 `--bind-repo`로 저장소를 밝혀야 묶인다(2026-11-02 개정: 기본 위치의 원장과 단서가 있는 원장은 밝힘 없이도 원장에 쓰고 끝난 동사로 묶이며 — 아무것도 쓰지 않은 동사는 밝힘 없이는 묶지 않는다, AC-AFL-001 (e5) — 그 쓰는 첫 동사가 맞는 저장소에서 불렸는지는 아무 AC도 단언하지 않는다). 저장소를 디스크에서 옮기면 git common dir의 절대 경로가 바뀌어 뒤의 호출이 거절된다 — 그 뒤의 복구 경로(`moai af move-repository`, 표지 비교, 표지 없는 원장의 대체 규칙, `--from`, 첫 이동의 스키마 판 올림)는 2026-11-01 정정으로 REQ-AFL-001에 적었고, 그것을 단언하는 AC는 없다. 표지(repository identity marker) 비교와 그 대체 규칙을 단언하는 AC도 없다. 훅 heartbeat 기록기(`internal/hook/ledger_heartbeat.go`)는 저장소를 확인하지 않는다(`grep -n -e repositor -e Repo`가 아무것도 찍지 않음; 종료 코드 관측 못 함). 셋 모두 progress §O의 열린 위험이다.
- (0.6.0) REQ-AFL-014 (v)의 "저장소가 답하지 못하면 DONE 거절"과 REQ-AFL-007의 "다른 epoch의 resume 거절"이 원장 층에서 나는지는 이 작성자가 찾은 시험 이름과 읽은 본문으로 판정하지 않았다(AC-AFL-007 (v), AC-AFL-014 (i) 줄).
- (0.6.0) 0.6.0 사례 줄의 착지 시험은 대부분 이름과 줄만 확인했다; 본문을 읽은 시험은 그 줄에 따로 적었다. 덮는 범위의 판정은 §F 규칙(본문이 적게 단언하면 partly)을 적용하지 않았으며, §F 표에 0.6.0 행을 더하지 않았다.

- REQ-AFL-003의 "트랜잭션 안 외부 명령 없음, 에이전트 턴을 넘는 트랜잭션 없음": AC-AFL-003의 시험은 Open 쪽만 본다. 확인 방법 미정(OD-35 (j)).
- REQ-AFL-018(0.5.0, D5): §6 상수 여덟 개가 모두 원장 패키지에 착지했다(`db02401ad1:internal/agenticfactory/ledger/api.go:41-48`) — 상위 OD-35 (d)는 착지 코드로 답해졌다. AC-AFL-018의 Then은 그 가운데 넷(lease·renew·reclaim cap·snapshot)만 적는다; 나머지 넷과, 훅 패키지에 따로 있는 `LedgerHeartbeatEvery = 30 * time.Second`(`db02401ad1:internal/hook/ledger_heartbeat.go:20`, `ledger.HeartbeatThrottle`를 쓰지 않음)는 AC가 단언하지 않는다. 이전 문구(0.4.0): «REQ-AFL-018: 원장 밖 상수(heartbeat 30 s, fix-loop 3, review 400, landing TTL 10 min)의 패키지 미정(OD-35) — AC-AFL-018은 원장 상수만 본다.»
- REQ-AFL-013의 REVIEWING 경로(교차 모델 리뷰): AC-AFL-013은 REVIEWING을 거치지 않는 경우만 본다. 리뷰 수단 미정(OD-35).
- (0.4.0) AC-AFL-004의 정적 검사는 `ledger` 디렉터리 바로 아래 파일만 읽는다. `importer` 하위 디렉터리는 덮지 않는다. 리터럴과 패키지 수준 문자열 상수를 `+`로 이은 SQL은 접어서 읽고, 접을 수 없는 피연산자가 다섯 시각 열에 값을 넣으면 FAIL이다(N5). 함수 안 지역 변수나 `fmt.Sprintf`로 만든 SQL은 여전히 (i)의 수집이 놓친다. 이전 문구(0.2.0): «AC-AFL-004의 정적 검사는 `ledger` 디렉터리 바로 아래 파일만 읽는다. `importer` 하위 디렉터리는 덮지 않고, SQL 문을 리터럴이 아닌 방법(문자열 조합 등)으로 만들면 (i)의 수집이 놓친다.»
- (0.3.0) REQ-AFL-020의 흐름 지표 표시는 AC-AFL-020이 판정하지 않는다. lease 키는 0.3.0에서 판정에 들었다(OD-39 결정; 이전 문구(0.2.0): "lease 키(출처 미정, OD-39)와 흐름 지표 표시는 AC-AFL-020이 판정하지 않는다").
- (0.3.0) REQ-AFL-015의 `effect` 자체·`docstore.Store` 조건부 퇴역, 원격 wiring을 먼저 걷어내는 순서, 지우기 전 역할 읽기, local-only 부분 이동, 빌드·vet 게이트는 AC-AFL-015가 단언하지 않는다. `identity` GitHub 조회와 `intake` GitHub 검색의 포함 여부는 미정이다.
- (0.3.0) REQ-AFL-008 메모의 절차 의무(같은 HEAD 리뷰, 독립 감사, QA PASS)는 AC가 없다 — 메모는 새 의무가 아니다(OD-37 결정).
- (0.4.0) REQ-AFL-014의 merged·closed·settled 사실 원천과 waiting·verifying·failed 쓰기의 card epoch 원천은 AC-AFL-014가 단언하지 않는다. trunk ref checkout 거절과 landing·landed의 잠금 epoch는 0.4.0에서 사례 (c)(b)로 들었다(N14). 이전 문구(0.2.0): «REQ-AFL-014의 trunk 전용 ref, QueueUpdate epoch 원천, merged·closed·settled 사실 원천은 AC-AFL-014의 시험 하나가 단언하지 않는다(AC-AFL-014는 0.1.0 그대로).»
- (0.4.0) REQ-AFL-010의 `Reap(nil)` 오류와 REQ-AFL-006의 park 사유 집합은 AC-AFL-010·006이 단언하지 않는다. "probe 밖, 트랜잭션 안 재확인"은 0.4.0에서 AC-AFL-010 L-Changed 사례로 들었다(N1). 이전 문구(0.2.0): «REQ-AFL-010의 "probe 밖, 트랜잭션 안 재확인"과 `Reap(nil)` 오류, REQ-AFL-006의 park 사유 집합은 AC-AFL-010·006이 단언하지 않는다(두 AC는 0.1.0 그대로).»
- (0.4.0) AC-AFL-019: 착지한 `flow.Measure`는 Reader 말고 활성 판정 `Probe`도 인자로 받는다(`1a552af405:internal/agenticfactory/monitor/flow/flow_test.go:47`: `Measure(context.Background(), src, probeOf(ledger.LivenessAlive), t0, Defaults())`). AC-AFL-019의 "계산은 그 Reader 말고 다른 입력을 쓰지 않는다"와 이 사실의 관계는 이 판에서 정하지 않았다 — 감사가 볼 곳이다.
- (0.5.0) REQ-AFL-014의 reconcile 예외(사례 (d)): 착지 시험은 landing 행의 되돌림만 저장소에서 단언하고 verifying 행의 되돌림은 단언하지 않는다. 계약 §10 20th review lane-22 (a)("Reconcile reverts a `verifying` row only when its `card_epoch` differs from `cards.epoch` or the card has no live lease")는 REQ-AFL-014에도 AC-AFL-014에도 넣지 않았다 — 이 판의 범위 밖이다(progress §L 남은 부채).
- (0.5.0) §B.2의 원장 항목은 종료 코드를 관측하지 못했다. 그래서 AC-AFL-013·015·017의 RED-now 칸은 `verification-completeness.md` §2.1의 네 요소 가운데 셋만 갖는다.
- (0.5.0) §N의 통과 증거: 이 작성자가 관측한 통과 출력은 없다. lane-18이 보고한 원장 패키지 `go test -race` ok는 보고이며 관측이 아니다.
- (0.4.0) 착지 시험이 덮는 범위(§F)는 이 작성자가 시험 본문을 읽어 판정한 것이며 시험을 돌리지 않았다. lane-18이 보고한 `go test -race` ok는 이 작성자가 관측하지 않았다.

## §F AC ↔ landed test (0.5.0)

이전 문구(제목): «## §F AC ↔ landed test (0.4.0)»

읽은 곳(0.5.0): integration tip `db02401ad1`(`git branch --list WT-agentic-factory-integration --format='%(refname:short) %(objectname:short)'` → `WT-agentic-factory-integration db02401ad1`; 감사가 인용한 `c83013938a`, 0.4.0이 읽은 `1a552af405`에서 움직임). lane-18 커밋 `607bd17804`의 `ac_names_test.go`·`verifying_reentry_test.go`는 tip에 들어왔다(§B.2 [L-names]). 시험 이름과 줄은 `git grep -n '^func Test' db02401ad1 -- <패키지>`로 다시 확인했다 — 아래 인용 줄은 `ledger_board_test.go`의 `TestBoard_RouteIsReadOnlyAndRendersEveryPanel`(213 → 216) 말고 같다. 본문은 `git diff 1a552af405 db02401ad1`·`git diff 607bd17804 db02401ad1`로 바뀐 곳을 보고 다시 읽었다: 바뀐 곳은 `QueueUpdate`의 epoch 인자 형식(`CardEpoch`/`LockEpoch`), `Close` 오류 처리, 가짜 큐의 `SetState` 서명, 보드 시험의 flow 단언이며, 아래 행이 인용한 단언 내용은 바뀌지 않았다. 판정 규칙(0.5.0, 감사 B4): 시험 본문이 AC가 요구하는 것보다 적게 단언하면 partly다 — 이름만 보고 판정하지 않는다. 시험은 돌리지 않았다. 0.4.0의 채택 규칙("fully인 행만 착지 이름을 AC 이름으로 채택")은 AC-AFL-021에서 틀린 판정 위에 적용되었다(그 행 참조).

이전 문구(0.4.0, 읽은 곳): «읽은 곳: integration tip `1a552af405`(`git branch --list WT-agentic-factory-integration --format='%(refname:short) %(objectname:short)'` → `WT-agentic-factory-integration 1a552af405`; 조정자가 적은 `915a6524cb`에서 움직임)과 lane-18 커밋 `607bd17804`(가지 `WT-agentic-factory-core-claim-runner`에만 있음). 시험 이름과 줄은 `git grep -n '^func Test' <commit> -- <패키지>`로 확인했고, 덮는 범위는 시험 본문을 `git show <commit>:<파일>`로 읽어 판정했다. 시험은 돌리지 않았다. "fully"인 행만 착지 이름을 AC 이름으로 채택할 수 있다.»

| AC | SPEC 시험 이름 | 착지 시험(이름, commit:file:line) | 덮음 | 본문이 단언하는 것(인용·요약) |
|---|---|---|---|---|
| AC-AFL-001 | `TestLedgerImportGraph_NoIssueOrRemoteClient` | 없음 | no landed test | — (참고: 퇴역 대상 import 0건은 이 작성자의 `git grep`, 시험 아님) |
| AC-AFL-002 | `TestStore_StaleEpochWrite_ReturnsErrStaleEpoch` | 같은 이름 `db02401ad1:internal/agenticfactory/ledger/ac_names_test.go:117`; 추가: `TestEveryWriteAfterClaimNeedsTheCurrentEpoch` `db02401ad1:internal/agenticfactory/ledger/lease_test.go:65`; `TestRelease_StaleEpoch_KeepsNewLeaseAndKeys` `db02401ad1:…/ac_names_test.go:147`; `TestEnqueueWithoutGateEvidenceForThatCommitIsStale` `db02401ad1:…/ledger/queue_test.go:69`(`:92`); `TestMoveNeedsTheLeaseEpochAndTheFactsTheLedgerHolds` `db02401ad1:…/ledger/move_test.go:16`(`:19`) | partly | 같은 이름: 회수 뒤 epoch 2로 다시 claim, epoch 1의 `"move"`·`"evidence"`·`"heartbeat"`가 `ErrStaleEpoch`, `"the card changed"`·`"a stale write left a row behind"`(evidence 0, transitions 수 같음) — S-Move·S-AddEvidence·S-Heartbeat의 오류와 행(heartbeat_at 불변은 단언 안 함). `lease_test.go:65`: `"heartbeat"`·`"release"`·`"expand"`·`"evidence"`를 `c.Epoch+1`로 불러 `ErrStaleEpoch`, `"a stale write changed the lease"`(키 수 1) — S-ExpandKeys: partly: future-epoch only; stale (older) epoch not asserted(같은 시험의 `:90`은 회수·재claim 뒤 옛 epoch의 `Heartbeat`만 본다). `ac_names_test.go:147`: 낡은 `Release` → `ErrStaleEpoch`, `"the new holder's lease changed"`·`"the new holder's key was deleted"` — S-Release 전부. `queue_test.go:92`: `Enqueue(c.Epoch+1)` → `ErrStaleEpoch` — S-Enqueue: partly: future-epoch only; stale (older) epoch not asserted; 그 호출 뒤 큐 행·단계 단언 없음 |
| AC-AFL-003 | `TestOpen_SetsBusyTimeoutAndChecksSQLiteVersion` | `TestOpenCreatesTheDirectoryTheSchemaAndTheSettings` `db02401ad1:…/ledger/open_test.go:23`; `TestTheEngineVersionIsChecked` `…/open_test.go:83` | partly | `busy != int(BusyTimeout.Milliseconds())`면 오류(`BusyTimeout = 5 * time.Second`, `api.go:49`); `checkVersion("3.34.9")`가 오류. `Open`이 낮은 버전에서 store를 내주지 않는 것은 단언 안 함 |
| AC-AFL-004 | `TestLedgerSource_TimestampColumnsUseDBClock` | `TestAddCardRoundTripsAndTheDatabaseClockStamps` `db02401ad1:…/ledger/claim_test.go:44` | not | `c.CreatedAt.IsZero() \|\| time.Since(c.CreatedAt) > time.Minute`면 오류 — Go 시계로 써도 통과하므로 AC를 덮지 않음. 정적 검사 시험 없음 |
| AC-AFL-005 | `TestReap_LivenessCases_ReclaimOnlyDead` | 같은 이름 `db02401ad1:…/ledger/ac_names_test.go:30`; 추가: `TestAnExpiredLeaseOfAnAliveHolderIsNotedOnceAndKept` `db02401ad1:…/ledger/review20_test.go:87` | fully(같은 이름; 0.5.0에 tip 본문 다시 확인 — `git diff 607bd17804 db02401ad1`에 이 시험 줄 변화 없음) | live·unknown: `leases`·`lease_keys` 1, `ReclaimCount` 0, `stalledNoteCount` 1; dead: 0·0, `ReclaimCount` 1; `Reap` 반환은 `"dead"` 하나. `review20_test.go:87`: Alive로 `Reap` 두 번 뒤 `stalled` 쪽지 하나 |
| AC-AFL-006 | `TestMove_RefusesEdgeRejectedByCheckMove` | `TestMoveNeedsTheLeaseEpochAndTheFactsTheLedgerHolds` `db02401ad1:…/ledger/move_test.go:16` | partly | `:53` READY로 `Move` → `core.ErrTransitionInvalid`. 거절 뒤 단계·`transitions` 불변 단언 없음(주석만 "stores nothing") |
| AC-AFL-007 | `TestReap_ThirdReclaim_ParksCard` | `TestTheThirdReclaimParksTheCardForTheOperator` `db02401ad1:…/ledger/claim_test.go:296`; `TestMoveToDoneOrCancelledReleasesTheLeaseAndKeysInTheSameTransaction` `…/move_test.go:83`; `TestRowsOfTheHistoryTablesCannotBeDeleted` `…/open_test.go:119` | partly | `:296`(Claim의 lazy reap 경로): `ReclaimCount == ReclaimCap`, `Park.Reason == ParkReasonReclaimCap`, lease 0. `:83`: CANCELLED 뒤 lease 0·`lease_keys` 0, DONE 뒤 lease 0(키는 단언 안 함). `:119`: `cards`·`transitions`·`evidence`·`notes`의 `DELETE`가 `"never deleted"`로 거절. `merge_queue` 행 생존, 사례 (iii), 앞선 `transitions` 수 보존은 단언 없음 |
| AC-AFL-008 | `TestEnqueue_EvidenceSubjectSHAMismatch_ReturnsErrEvidenceStale` | `TestEnqueueBindsTheCandidateToPassingGateEvidenceAndMovesTheCardIntoMerging` `db02401ad1:…/ledger/queue_test.go:46`; `TestEnqueueWithoutGateEvidenceForThatCommitIsStale` `…/queue_test.go:69`; `TestMergingIsEnteredByEnqueueAlone` `…/move_test.go:58` | partly | (i) `:46` 단계 `StageMerging`; (ii) `:69` `"another commit": "c2"` → `ErrEvidenceStale`, `"a refused enqueue left a queue row"`, 단계 BUILDING; (iii) `:58` `Move`로 MERGING → `ErrInvalid`(단계 불변 단언 없음) |
| AC-AFL-009 | `TestSchema_CheckConstraints_RejectUnknownQueueStateAndNoteKind` | `TestTheSchemaRefusesWhatTheContractClosesOff` `db02401ad1:…/ledger/open_test.go:73`; `TestOpenCreatesTheDirectoryTheSchemaAndTheSettings` `…/open_test.go:23` | partly | `'gossip'` 쪽지 종류와 `'paused'` 큐 상태 INSERT가 `err == nil`이면 오류(오류 종류는 보지 않음); 여덟 표 존재. `lease_keys.key` 기본 키 단언 없음 |
| AC-AFL-010 | `TestClaim_SingleTransaction_ReadyToBuildingByClass` | `TestClaimPicksTheHighestPriorityReadyCardAndMovesIt` `db02401ad1:…/ledger/claim_test.go:81`; `TestClaimWritesTheTransitionsAndTheLease` `:111`; `TestClaimSkipsWhatIsParkedUnmetOrKeyed` `:156`; `TestADeadHoldersCardIsReclaimedAtTheStartOfClaim` `:244`; `TestALiveOrUnknownHolderIsNeverReclaimedByClaim` `:275`; `TestTheJudgeRunsOutsideTheTransactionAndARenewalSinceTheProbeIsRespected` `…/reap_test.go:95`; `TestClaim_LazyReap_DoesNotReclaimLiveOwner` `db02401ad1:…/ac_names_test.go:74` | partly | `:81` class C → `BUILDING/SPEC`, class A → `BUILDING/IMPLEMENT`, 우선순위 높은 카드, epoch 1. `:156` `ErrNoWork`. `:244` Dead → epoch 2로 다시 claim, `ReclaimCount` 1 (L-Dead); 다시 claim된 카드의 단계가 BUILDING이다(`:266`) — 회수된 BUILDING 카드를 다시 고르는 코드 사실(OD-43). `:275` alive·unknown → `ErrNoWork`, lease 1 — L-Unknown: partly: asserts lease count only, no other READY card(카드가 하나뿐이라 '다른 READY 카드 claim됨' 칸을 볼 수 없고, H의 키·`cards.epoch`·`reclaim_count` 불변도 단언 없음). `reap_test.go:95` 판정 중 `heartbeat_at` 갱신 → `Reap`·`Claim` 모두 회수 안 함 (L-Changed-heartbeat; L-Changed-epoch 사례는 착지 시험에서 찾지 못했다 — `reap_test.go`의 시험 네 개 이름을 보았고 본문은 `:95`만 읽음). `ac_names_test.go:74` Live 소유자 lease·키·epoch 1·`ReclaimCount` 0 그대로, 다른 카드 claim, 두 번 `Claim` 뒤 `"stalled notes %d after two Claims, want exactly one"` (L-Live 전부) |
| AC-AFL-011 | `TestExpandKeys_Conflict_ReturnsErrKeyConflict` | `TestExpandKeysAddsAllOrNothingAndConflictsParkNothing` `db02401ad1:…/ledger/lease_test.go:126` | partly | b가 a의 `KeyGoMod`를 요청 → `ErrKeyConflict`, `"a refused expansion added keys"`, `"a conflict parks nothing"`(b만). a(X)의 park 없음은 단언 없음 |
| AC-AFL-012 | `TestClaim_DisjointConflictKeys_ClaimInParallel` | `TestAConflictKeyHeldElsewhereSkipsTheCardAndParksNothing` `db02401ad1:…/ledger/claim_test.go:202`; `TestConcurrentClaimsNeverShareACard` `:333` | partly | `:202` 첫 카드를 쥔 채 겹치지 않는 `free` 카드가 claim됨; `:333` 네 worker가 여덟 카드를 하나씩 모두 claim. 두 카드가 동시에 BUILDING인 단계 단언 없음 |
| AC-AFL-013 | `TestCardFlow_ClaimToDone_ThroughMergeQueue` | 없음 | no landed test | — |
| AC-AFL-014 | `TestMergeq_Land_CASUpdateRefThenLandedNote` | `TestLand_CleanCandidateMovesTheTrunkToTheMergeResult` `db02401ad1:internal/agenticfactory/mergeq/mergeq_test.go:152`; `TestLand_GatesRunOutsideTheLock` `:203`; `TestLand_ATrunkThatMovesAfterTheRecheckIsCaughtByTheCompareAndSwap` `:327`(본문 읽지 않음, 이름만); `TestLand_RefusesATrunkThatIsCheckedOutInAWorktree` `:416`; `TestReconcile_RecordsOneLandedNoteAndIsIdempotent` `:664`; `TestQueueUpdateChecksTheEpochItsStateCallsFor` `…/ledger/queue_test.go:136`; `TestQueueUpdate_EpochSource_FencesByKind` `db02401ad1:…/ledger/ac_names_test.go:174`; `TestQueueUpdate_VerifyingReentry_AtCardEpoch` `db02401ad1:…/ledger/verifying_reentry_test.go:10`; (0.5.0, 사례 (d)) `TestReconcile_PutsARowBackToWaitingOnlyWhileHoldingTheLandingLock` `db02401ad1:…/mergeq/mergeq_test.go:586` | partly | `:152` trunk가 merge 결과로, 상태 `verifying,landing,landed`, 잠금 1/1, `landed` 쪽지 정확히 1 — 가짜 큐(`fakeQueue`)라 카드 DONE 이동은 단언 안 함. `:203` `"the gates ran while the landing lock was held"`. `:416` (c): `ErrTrunkCheckedOut`, trunk 그대로, 관문 0회, 잠금 0회, holder 제거 뒤 대조 착지. `queue_test.go:136` (b): landing을 card epoch·`lock+5`로 → `ErrStaleEpoch`. `ac_names_test.go:174` (b): landing을 `lock+1`·card epoch로 → `ErrStaleEpoch`, landed를 card epoch로 → `ErrStaleEpoch`, 낡은 card epoch 행은 어느 epoch로든 `failed` 거절. `verifying_reentry_test.go:10`: card epoch로 verifying→verifying 허용, 다른 epoch 거절, landing→verifying `ErrInvalid` (trunk checkout 거절은 mergeq·lane-22 몫, 원장 몫이 아님). (0.5.0, (d)) `queue_test.go:170-176`: landing 행을 카드 epoch 값(`LockEpoch(c.Epoch, "c1")`)으로 waiting에 → `ErrStaleEpoch`(d1), 잠금 epoch로 → 오류 없음(d2); 되돌린 뒤 행 상태는 단언 없음, verifying 행 되돌림 없음. `mergeq_test.go:586`: `ErrLockHeld`면 행 상태 변화 0(d3), 잠금을 잡으면 waiting으로 한 번, 잠금 1회 잡고 1회 풂 — 가짜 큐라 저장소의 epoch 검사는 보지 않음 |
| AC-AFL-015 | `TestRetirement_NoImportOfForgeMirrorGitref` | 없음 | no landed test | — (`db02401ad1:internal/agenticfactory/effect/guard_inventory_test.go:68` `TestEveryWritePathUnderEffectUsesTheGuard`는 쓰기 경로 guard 주제라 이 AC가 아님) |
| AC-AFL-016 | `TestImporter_DryRun_SourcesUnchanged` | `TestDry_NeverWritesToASource` `db02401ad1:internal/agenticfactory/ledger/importer/importer_test.go:136` | partly | backlog.db·factory.db의 sha256이 dry-run 전후 같음(`"a source file changed during the dry-run"`), 읽기 전용 연결 대조(`"control: the read-only connection accepted a write"`). claim JSON 원천 없음 |
| AC-AFL-017 | `TestAFCommand_VerbsRegistered` | 없음 | no landed test | — (`internal/cli/agentic_factory_test.go`의 시험은 `moai af` 동사를 보지 않음 — 이름만 확인) |
| AC-AFL-018 | `TestLedgerDefaults_MatchContract` | 없음 | not | — (상수 `LeaseDuration = 15 * time.Minute`, `RenewEvery = 5 * time.Minute`, `ReclaimCap = 3`, `SnapshotEvery = 30 * time.Minute`, `db02401ad1:…/ledger/api.go:41-48`; 0.5.1 N6: `b28a31d9dc`에서 `git grep -n`으로 `HeartbeatThrottle` 41행·`SnapshotEvery` 48행·`MinSQLite` 50행을 다시 확인 — §6 상수 여덟 개는 41-48, 이전 문구 "41-50"; 값을 단언하는 시험 없음) |
| AC-AFL-019 | `TestFlow_StalledDetection_FromReader` | `TestStalled_Rule` `db02401ad1:internal/agenticfactory/monitor/flow/flow_test.go:76`; `TestStalled_ReportsEvidence` `:105` | partly | 가짜 Reader(`fakeSource`)와 probe로: heartbeat 낡음·wake 지남·dead/unknown → stalled, `"process alive is never stalled"`. `Measure`가 probe도 입력으로 받음(§E 0.4.0) |
| AC-AFL-020 | `TestBoard_ReaderOnly_NoStoreReference` | `TestBoard_RouteIsReadOnlyAndRendersEveryPanel` `db02401ad1:internal/web/ledger_board_test.go:216`; `TestBoard_ConflictKeysNameTheirHolder` `:134` | partly | `:216` 패널 여섯 id와 표지 문자열, `"the board must be display-only"`(`<input` 등 없음), `POST /board`·`/api/board` 405; `:134` 키 `pkg:internal/web`·`shared:go.mod`·`pkg:internal/cli`의 소유 카드. (ii) 정적 `ledger.Store`·`ledger.Open` 0건 단언 없음 |
| AC-AFL-021 | `TestLedgerHeartbeat_ThrottlesToTheStampFile`(0.4.0 채택; 이전 `TestHeartbeatWriter_SkipsWriteWhenStampFresh`) | 같은 이름 `db02401ad1:internal/hook/ledger_heartbeat_test.go:58` | partly (0.4.0: fully — 틀린 판정) | `LedgerHeartbeatEvery - time.Second` 뒤 `"a stamp younger than %v must skip the DB write"`(호출 1), `+2s` 뒤 `"a stamp older than %v must beat again"`(호출 2); `LedgerHeartbeatEvery = 30 * time.Second`(`db02401ad1:internal/hook/ledger_heartbeat.go:20`). 덮지 못하는 것(0.5.0, 감사 B4): 30 s를 직접 고정하지 않는다 — 시험은 `LedgerHeartbeatEvery`를 기준으로 앞뒤만 보므로 그 상수를 1 s로 바꾼 변이도 통과한다; stamp 파일 자체(경로·내용)는 단언하지 않고 DB 호출 수만 본다; 시계는 주입한 `ledgerHeartbeatNow`다 |

## §N 사후 관측 (release 게이트 밖)

(0.6.0 메모: AC-AFL-001·007·008·009·014의 0.6.0 사례와 그 착지 시험은 §C의 각 AC에 적었다. 아래 각 AC 항목은 0.5.1 문구 그대로이며, 0.6.0 사례에 대해서도 통과 증거는 없다 — 이 작성자는 어떤 시험도 돌리지 않았다.)

(0.5.0 신설, 감사 B1.) 코드나 시험이 integration tip `db02401ad1`에 착지한 AC 18개다. 이 AC들은 release 게이트 밖이며 완료 계산에 들지 않는다(§D). 각 AC의 제목과 시험 이름 하나는 §C에 그대로 있다. 아래 '단언'은 착지 시험 본문을 읽어 적은 요약이고, 인용과 줄은 §F의 같은 AC 행에 있다. 이 작성자는 어떤 시험도 돌리지 않았다.

- **AC-AFL-001** (`TestLedgerImportGraph_NoIssueOrRemoteClient`) — 착지 시험: 없음. 단언: 없음(참고: 퇴역 대상 import 0건은 이 작성자의 `git grep`이며 시험이 아니다). 덮음: no landed test.
  - 통과 증거: 없음.
  - RED 미관측.
- **AC-AFL-002** (`TestStore_StaleEpochWrite_ReturnsErrStaleEpoch`) — 착지 시험: 같은 이름(`ac_names_test.go:117`), `lease_test.go:65`, `ac_names_test.go:147`, `queue_test.go:69`, `move_test.go:16`. 단언: 회수·재claim 뒤 옛 epoch의 Move·AddEvidence·Heartbeat·Release가 `ErrStaleEpoch`이고 행이 바뀌지 않음; ExpandKeys·Enqueue는 앞선(미래) epoch로만 시험. 덮음: partly.
  - 통과 증거: 관측 없음 — lane-18이 보고한 원장 패키지 `go test -race` ok는 보고이며 이 작성자의 관측이 아니다.
  - RED 미관측.
- **AC-AFL-003** (`TestOpen_SetsBusyTimeoutAndChecksSQLiteVersion`) — 착지 시험: `open_test.go:23`, `:83`. 단언: busy_timeout이 `BusyTimeout`과 같음, `checkVersion("3.34.9")` 오류. 낮은 버전에서 `Open`이 store를 내주지 않는 것은 단언 없음. 덮음: partly.
  - 통과 증거: 관측 없음 — lane-18 보고(위와 같음)뿐.
  - RED 미관측.
- **AC-AFL-004** (`TestLedgerSource_TimestampColumnsUseDBClock`) — 착지 시험: `claim_test.go:44`. 단언: `CreatedAt`이 0이 아니고 1분 안 — Go 시계로 써도 통과. 정적 검사 시험 없음. 덮음: not.
  - 통과 증거: 관측 없음 — lane-18 보고뿐.
  - RED 미관측.
- **AC-AFL-005** (`TestReap_LivenessCases_ReclaimOnlyDead`) — 착지 시험: 같은 이름(`ac_names_test.go:30`), `review20_test.go:87`. 단언: live·unknown은 lease·키 1, `ReclaimCount` 0, stalled 쪽지 1; dead는 0·0, `ReclaimCount` 1, `Reap` 반환 하나. 덮음: fully.
  - 통과 증거: 관측 없음 — lane-18 보고뿐.
  - RED 미관측.
- **AC-AFL-006** (`TestMove_RefusesEdgeRejectedByCheckMove`) — 착지 시험: `move_test.go:16`. 단언: 거절된 이동이 `core.ErrTransitionInvalid`. 거절 뒤 단계·`transitions` 불변은 단언 없음. 덮음: partly.
  - 통과 증거: 관측 없음 — lane-18 보고뿐.
  - RED 미관측.
- **AC-AFL-007** (`TestReap_ThirdReclaim_ParksCard`) — 착지 시험: `claim_test.go:296`, `move_test.go:83`, `open_test.go:119`. 단언: 셋째 회수에서 `reclaim-cap` park·lease 0; CANCELLED 뒤 lease·키 0, DONE 뒤 lease 0; 네 기록 표의 DELETE 거절. 사례 (iii)·DONE 뒤 키 0·`merge_queue` 행 생존은 단언 없음. 덮음: partly.
  - 통과 증거: 관측 없음 — lane-18 보고뿐.
  - RED 미관측.
- **AC-AFL-008** (`TestEnqueue_EvidenceSubjectSHAMismatch_ReturnsErrEvidenceStale`) — 착지 시험: `queue_test.go:46`, `:69`, `move_test.go:58`. 단언: 같은 커밋 증거면 MERGING; 다른 커밋이면 `ErrEvidenceStale`·큐 행 없음·단계 BUILDING; `Move`로 MERGING은 `ErrInvalid`(단계 불변은 단언 없음). 덮음: partly.
  - 통과 증거: 관측 없음 — lane-18 보고뿐.
  - RED 미관측.
- **AC-AFL-009** (`TestSchema_CheckConstraints_RejectUnknownQueueStateAndNoteKind`) — 착지 시험: `open_test.go:73`, `:23`. 단언: 목록 밖 쪽지 종류·큐 상태 INSERT가 오류(오류 종류는 보지 않음); 여덟 표 존재. 기본 키는 단언 없음. 덮음: partly.
  - 통과 증거: 관측 없음 — lane-18 보고뿐.
  - RED 미관측.
- **AC-AFL-010** (`TestClaim_SingleTransaction_ReadyToBuildingByClass`) — 착지 시험: `claim_test.go:81`·`:111`·`:156`·`:244`·`:275`, `reap_test.go:95`, `ac_names_test.go:74`. 단언: 등급별 시작 단계·우선순위·`ErrNoWork`; L-Live 전부; L-Unknown은 lease 수만; L-Dead; L-Changed-heartbeat. L-Changed-epoch는 착지 시험을 찾지 못함. 덮음: partly.
  - 통과 증거: 관측 없음 — lane-18 보고뿐.
  - RED 미관측.
- **AC-AFL-011** (`TestExpandKeys_Conflict_ReturnsErrKeyConflict`) — 착지 시험: `lease_test.go:126`. 단언: 요청한 쪽의 `ErrKeyConflict`·키 불변·park 없음. 키를 쥔 쪽의 park 없음은 단언 없음. 덮음: partly.
  - 통과 증거: 관측 없음 — lane-18 보고뿐.
  - RED 미관측.
- **AC-AFL-012** (`TestClaim_DisjointConflictKeys_ClaimInParallel`) — 착지 시험: `claim_test.go:202`, `:333`. 단언: 겹치지 않는 카드가 claim됨; 동시 claim이 카드를 나눠 갖지 않음. 두 카드가 동시에 BUILDING인 단계는 단언 없음. 덮음: partly.
  - 통과 증거: 관측 없음 — lane-18 보고뿐.
  - RED 미관측.
- **AC-AFL-014** (`TestMergeq_Land_CASUpdateRefThenLandedNote`) — 착지 시험: `mergeq_test.go:152`·`:203`·`:416`·`:586`·`:664`, `queue_test.go:136`(그 안 `:170-176`), `ac_names_test.go:174`, `verifying_reentry_test.go:10`. 단언: CAS 착지·`landed` 쪽지 하나·잠금 밖 관문·checkout 거절·landing/landed 잠금 epoch 거절·reconcile의 landing→waiting은 잠금 epoch로만·잠금 없이 되돌리지 않음. 카드 DONE 이동(가짜 큐)과 verifying 행 되돌림은 단언 없음. 덮음: partly.
  - 통과 증거: 관측 없음 — 원장 쪽은 lane-18 보고뿐, mergeq 쪽은 이 작성자가 읽은 보고도 없음.
  - RED 미관측.
- **AC-AFL-016** (`TestImporter_DryRun_SourcesUnchanged`) — 착지 시험: `importer_test.go:136`. 단언: backlog.db·factory.db의 sha256이 dry-run 전후 같음, 읽기 전용 연결 대조. claim JSON 원천은 없음. 덮음: partly.
  - 통과 증거: 관측 없음 — 이 작성자가 읽은 보고도 없음.
  - RED 미관측.
- **AC-AFL-018** (`TestLedgerDefaults_MatchContract`) — 착지 시험: 없음. 상수는 `db02401ad1:internal/agenticfactory/ledger/api.go:41-48`에 있으나 값을 단언하는 시험이 없다. 덮음: not.
  - 통과 증거: 없음.
  - RED 미관측.
- **AC-AFL-019** (`TestFlow_StalledDetection_FromReader`) — 착지 시험: `flow_test.go:76`, `:105`. 단언: 가짜 Reader와 probe로 stalled 규칙, 살아 있는 프로세스는 stalled 아님. `Measure`가 probe도 입력으로 받는다(§E 0.4.0). 덮음: partly.
  - 통과 증거: 관측 없음 — 이 작성자가 읽은 보고도 없음.
  - RED 미관측.
- **AC-AFL-020** (`TestBoard_ReaderOnly_NoStoreReference`) — 착지 시험: `ledger_board_test.go:216`, `:134`. 단언: 패널 여섯 id와 표지 문자열, 보기 전용, POST 405; 키의 소유 카드. 정적 `ledger.Store`·`ledger.Open` 0건은 단언 없음. 덮음: partly.
  - 통과 증거: 관측 없음 — 이 작성자가 읽은 보고도 없음.
  - RED 미관측.
- **AC-AFL-021** (`TestLedgerHeartbeat_ThrottlesToTheStampFile`) — 착지 시험: 같은 이름(`internal/hook/ledger_heartbeat_test.go:58`). 단언: 주입한 시계로 `LedgerHeartbeatEvery - 1s` 뒤 DB 호출 1, `+2s` 뒤 2. 30 s 고정·stamp 파일은 단언 없음. 덮음: partly.
  - 통과 증거: 관측 없음 — 이 작성자가 읽은 보고도 없음.
  - RED 미관측.
