# SPEC-AGENTIC-FACTORY-LOCAL-001 — 수용 기준 (0.2.0)

## §A 측정 규칙

- AC마다 검증 대상 시험 이름(Go 시험 함수 이름)은 **정확히 하나**다. 이름은 이 작성자가 계약의 패키지·소유 lane에 맞춰 붙인 제안 이름이다.
- 21개 이름은 모두 기준 커밋 `37b67f6d39`에 없다(§B). 그래서 **실행 명령을 두지 않는다**: 아무 시험도 고르지 않는 `-run` 명령은 0개를 훑고도 `ok`를 낸다(`.claude/rules/moai/development/verification-completeness.md` §1.1).
- 각 AC의 두 칸(`verification-completeness.md` §2): **RED-now**(지금 무엇이 왜 빨간가)와 **green path**(어느 lane의 어떤 커밋이 뒤집는가). 둘 다 글로 적은 것이며 실행 결과가 아니다.
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

## §C 수용 기준

#### AC-AFL-001 — 단일 로컬 권위, Issue 읽기·쓰기 없음

- 주 REQ: maps REQ-AFL-001.
- Given 원장 패키지와 그것이 가져오는 패키지 그래프, When 그래프를 훑을 때, Then `forge`·`mirror`·`gitref`·`effect/ghintake`·`effect/ctxledger`·`effect/labels` 가운데 어느 것도 그래프에 없고, 원장 파일 경로는 `<moaiHome>/db/<projectKey>/af/ledger.db` 하나다.
- 시험 이름(하나): `TestLedgerImportGraph_NoIssueOrRemoteClient` — 소유 lane-18, 패키지 `internal/agenticfactory/ledger`. unwritten at 37b67f6d39(§B).
- RED-now: `TestLedgerImportGraph_NoIssueOrRemoteClient`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.
- green path: lane-18의 원장 첫 구현 커밋(계약 §4: `ledger/api.go` 형식 전용 커밋 뒤의 Open/schema 커밋)이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.

#### AC-AFL-002 — fencing epoch와 낡은 epoch 쓰기

- 주 REQ: maps REQ-AFL-002.
- Given 카드 하나가 epoch 1로 claim된 뒤 다시 claim되어 epoch 2가 된 상태, When epoch 1을 실은 `Move`·`AddEvidence`·`Heartbeat`를 부를 때, Then 바뀐 행이 0이고 오류는 `ErrStaleEpoch`이며, 카드 상태는 그대로다.
- 시험 이름(하나): `TestStore_StaleEpochWrite_ReturnsErrStaleEpoch` — 소유 lane-18, 패키지 `internal/agenticfactory/ledger`. unwritten at 37b67f6d39(§B).
- RED-now: `TestStore_StaleEpochWrite_ReturnsErrStaleEpoch`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.
- green path: lane-18의 Claim/Move 구현 커밋이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.

#### AC-AFL-003 — BEGIN IMMEDIATE·busy_timeout·SQLite 버전

- 주 REQ: maps REQ-AFL-003.
- Given 새 원장 파일, When `Open`이 연결을 열 때, Then busy_timeout이 5000 ms로 설정되어 있고, SQLite 버전이 3.35보다 낮다고 보고되면 `Open`이 오류를 돌려주며 store를 내주지 않는다.
- 시험 이름(하나): `TestOpen_SetsBusyTimeoutAndChecksSQLiteVersion` — 소유 lane-18, 패키지 `internal/agenticfactory/ledger`. unwritten at 37b67f6d39(§B).
- RED-now: `TestOpen_SetsBusyTimeoutAndChecksSQLiteVersion`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.
- green path: lane-18의 Open 구현 커밋이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.

#### AC-AFL-004 — 사건 시각은 DB 시계

- 주 REQ: maps REQ-AFL-004.
- Given `internal/agenticfactory/ledger` 디렉터리 바로 아래의 `_test.go`가 아닌 `.go` 파일 전부(하위 디렉터리 `importer`는 넣지 않음), When 시험이 그 파일들을 `go/parser`로 읽어 (i) 문자열 리터럴 가운데 `INSERT`·`UPDATE` 문이 다섯 시각 열 `created_at`·`updated_at`·`at`·`heartbeat_at`·`enqueued_at` 가운데 하나에 값을 넣는 자리와 (ii) `time.Now`·`time.Since`·`time.Until`·`time.After` 선택자 호출을 모을 때, Then (i)의 모든 자리가 `strftime('%Y-%m-%dT%H:%M:%fZ','now')`로 값을 넣고(자리표시자 `?`나 Go 쪽 값이 아님), (ii)는 아래 허용 예외 밖에서 0건이다.
- 시험이 읽는 것: 위 파일들의 Go 구문 트리(선택자 식과 문자열 리터럴). DB를 열지 않는다. 빈 집합 확인: 시험은 먼저 (i)에서 다섯 열 각각에 대해 모은 자리가 1개 이상임을 단언한다 — 아무것도 모으지 못하고 통과하는 것을 막는다(`verification-completeness.md` §1.1).
- 허용 예외(명시 목록, 이 밖은 없다): (e1) `Heartbeat`의 호출자 인자 `nextWake` — `next_wake_at`에 들어간다. 다섯 시각 열이 아니며, 계약 §10은 "nextWake is informational only"라고 적지만 I4와의 관계는 열린 예외로 둔다(OD-35 (f)). 다섯 열 밖인 `expires_at`(leases·locks)의 원천도 OD-35 (f)에 열려 있다. 30분 snapshot 주기의 타이머(`time.NewTicker`·`time.AfterFunc`)는 (ii)의 네 선택자와 이름이 달라 세지 않는다. 새 예외는 SPEC 개정으로만 이 목록에 들어간다.
- 죽여야 할 변이: Go `time.Now()` 값을 문자열로 만들어 `?`로 `at`에 넣는 구현 → (i)에서 FAIL; Go 시계를 다른 계산에 쓰는 구현 → (ii)에서 FAIL.
- 시험 이름(하나): `TestLedgerSource_TimestampColumnsUseDBClock` — 소유 lane-18, 패키지 `internal/agenticfactory/ledger`. unwritten at 37b67f6d39(§B).
- RED-now: `TestLedgerSource_TimestampColumnsUseDBClock`가 `37b67f6d39`에 없다(§B). 원장 패키지도 없다. 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.
- green path: lane-18의 schema·쓰기 구현 커밋이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.
- 이전 문구(0.1.0): Given 호출자 쪽 시계를 먼 과거로 돌린 시험 환경, When `AddCard`·`Move`·`Heartbeat`·`Enqueue`를 부를 때, Then `created_at`·`updated_at`·`at`·`heartbeat_at`·`enqueued_at`이 호출자 시각이 아니라 DB의 `strftime('%Y-%m-%dT%H:%M:%fZ','now')` 값이다. 시험 이름 `TestStore_Timestamps_UseDBClock`.

#### AC-AFL-005 — 살아 있는 소유자는 회수하지 않음

- 주 REQ: maps REQ-AFL-005.
- Given lease 기한이 지난 카드 셋(소유자가 서로 다름)과, 소유자마다 아래 표의 판정을 돌려주는 `alive` 함수, When `Reap`을 한 번 부를 때, Then 사례마다 아래 표와 같다(한 시험 안의 하위 사례 셋).

| 사례 | `alive` 판정 | lease·키 | `reclaim_count` | `stalled` 쪽지 |
|---|---|---|---|---|
| Live | 살아 있음(계약 §10의 `Alive`) | 그대로 | 그대로 | 그 card·epoch에 하나(계약 §10 20th (d); 조정자 지시 "no `stalled` note"와 다름 — OD-41) |
| Unknown | 판정 불명 | 그대로(살아 있는 것으로 다룸) | 그대로 | 하나 |
| Dead | 죽음 | 없어짐(한 번 회수, `Reap` 반환값에 그 카드 하나) | +1 | 단언하지 않음 |

- 죽여야 할 변이: 기한 지난 Live 소유자를 회수하는 구현 → Live 행에서 FAIL(I5 "A live process is never reclaimed automatically"). 그 밖에 Unknown을 회수하는 구현은 Unknown 행, Dead를 회수하지 않는 구현은 Dead 행에서 FAIL.
- 시험 이름(하나): `TestReap_LivenessCases_ReclaimOnlyDead` — 소유 lane-18, 패키지 `internal/agenticfactory/ledger`. unwritten at 37b67f6d39(§B).
- RED-now: `TestReap_LivenessCases_ReclaimOnlyDead`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.
- green path: lane-18의 Reap 구현 커밋(활성 판정 어댑터는 lane-20)이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.
- 이전 문구(0.1.0): Given lease 기한이 지난 카드와, 그 소유자에 대해 판정 불명을 돌려주는 `alive` 함수, When `Reap`을 부를 때, Then lease와 키가 그대로 남고 `reclaim_count`가 늘지 않으며 `stalled` 쪽지가 하나 기록된다. 시험 이름 `TestReap_UnknownLiveness_KeepsLeaseAndWritesStalledNote`.

#### AC-AFL-006 — 단계 쓰기는 core.CheckMove를 거침

- 주 REQ: maps REQ-AFL-006.
- Given READY 카드 하나, When `core.CheckMove`가 거절하는 위치(예: READY에서 DONE)로 `Move`를 부를 때, Then 오류가 돌아오고 `cards.stage`와 `transitions`가 바뀌지 않는다.
- 시험 이름(하나): `TestMove_RefusesEdgeRejectedByCheckMove` — 소유 lane-18, 패키지 `internal/agenticfactory/ledger`. unwritten at 37b67f6d39(§B).
- RED-now: `TestMove_RefusesEdgeRejectedByCheckMove`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.
- green path: lane-18의 Move 구현 커밋(F6 강화 `37b67f6d39` 위)이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.

#### AC-AFL-007 — 행을 지우지 않음·세 번째 회수에서 park

- 주 REQ: maps REQ-AFL-007.
- Given 이미 두 번 회수된 카드와 Dead를 돌려주는 `alive` 함수, When 세 번째 회수가 일어날 때, Then `reclaim_count`가 3이고 카드가 park 사유 `reclaim-cap`으로 park되며, 앞선 `transitions` 행이 하나도 지워지지 않았고, 회수된 소유자의 `leases`·`lease_keys` 행은 없다(`Reap`이 지운 것; REQ-AFL-007, OD-36).
- 이전 문구(0.1.0): Given 이미 두 번 회수된 카드와 Dead를 돌려주는 `alive` 함수, When 세 번째 회수가 일어날 때, Then `reclaim_count`가 3이고 카드가 park되며, 앞선 `transitions` 행이 하나도 지워지지 않았다.
- 시험 이름(하나): `TestReap_ThirdReclaim_ParksCard` — 소유 lane-18, 패키지 `internal/agenticfactory/ledger`. unwritten at 37b67f6d39(§B).
- RED-now: `TestReap_ThirdReclaim_ParksCard`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.
- green path: lane-18의 Reap 구현 커밋이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.

#### AC-AFL-008 — Enqueue는 같은 commit SHA의 관문 증거를 요구함

- 주 REQ: maps REQ-AFL-008.
- Given 관문 증거 행(`subject_sha` = 커밋 A)을 가진 카드 P와 Q, When (i) P에 `Enqueue(candidateSHA=A)`, (ii) Q에 `Enqueue(candidateSHA=B)`, (iii) 증거가 있는 다른 카드를 `Move`로 MERGING에 바로 옮기려 할 때, Then (i) 받아들여져 P가 MERGING이고, (ii) `ErrEvidenceStale`이 돌아오며 Q의 단계는 그대로이고, (iii) 거절되어 단계가 그대로다.
- 죽여야 할 변이: SHA를 비교하지 않는 구현 → (ii)에서 FAIL; 아무 Enqueue도 받지 않는 구현 → (i)에서 FAIL; Move로 MERGING을 허용하는 구현 → (iii)에서 FAIL.
- 시험 이름(하나): `TestEnqueue_EvidenceSubjectSHAMismatch_ReturnsErrEvidenceStale` — 소유 lane-18, 패키지 `internal/agenticfactory/ledger`. unwritten at 37b67f6d39(§B).
- RED-now: `TestEnqueue_EvidenceSubjectSHAMismatch_ReturnsErrEvidenceStale`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.
- green path: lane-18의 AddEvidence·Enqueue 구현 커밋이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.
- 이전 문구(0.1.0): 제목 "증거는 tree SHA에 묶임"; Given tree SHA A로 기록된 증거 행, When tree SHA B를 판정할 때 그 증거를 셀 때, Then 그 증거는 세어지지 않는다. 시험 이름 `TestEvidence_DifferentTreeSHA_IsInvalid`.

#### AC-AFL-009 — 스키마와 CHECK 제약

- 주 REQ: maps REQ-AFL-009.
- Given 막 만든 원장, When `merge_queue.state`에 목록 밖 값을, `notes.kind`에 목록 밖 값을 넣으려 할 때, Then 두 쓰기 모두 CHECK 제약 위반으로 거절되고, 여덟 표(`cards`·`transitions`·`leases`·`lease_keys`·`evidence`·`merge_queue`·`locks`·`notes`)가 모두 있으며 `lease_keys.key`가 기본 키다.
- 시험 이름(하나): `TestSchema_CheckConstraints_RejectUnknownQueueStateAndNoteKind` — 소유 lane-18, 패키지 `internal/agenticfactory/ledger`. unwritten at 37b67f6d39(§B).
- RED-now: `TestSchema_CheckConstraints_RejectUnknownQueueStateAndNoteKind`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.
- green path: lane-18의 schema 커밋이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.

#### AC-AFL-010 — Claim 한 트랜잭션·등급별 시작 단계

- 주 REQ: maps REQ-AFL-010.
- Given 우선순위가 다른 READY 카드 둘(class C 하나, class A 하나), When `Claim`을 부를 때, Then 우선순위가 높은 카드 하나만 lease·키·epoch 증가와 함께 BUILDING/SPEC(C) 또는 BUILDING/IMPLEMENT(A·B)로 옮겨지고 id와 epoch가 돌아오며, 고를 카드가 없으면 `ErrNoWork`가 돌아온다.
- 시험 이름(하나): `TestClaim_SingleTransaction_ReadyToBuildingByClass` — 소유 lane-18, 패키지 `internal/agenticfactory/ledger`. unwritten at 37b67f6d39(§B).
- RED-now: `TestClaim_SingleTransaction_ReadyToBuildingByClass`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.
- green path: lane-18의 Claim 구현 커밋이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.

#### AC-AFL-011 — 충돌 키와 ExpandKeys

- 주 REQ: maps REQ-AFL-011.
- Given 카드 X가 `shared:go.mod` 키를 잡고 있는 상태, When 카드 Y가 `ExpandKeys`로 같은 키를 요청할 때, Then `ErrKeyConflict`가 돌아오고 어느 카드도 park되지 않는다.
- 시험 이름(하나): `TestExpandKeys_Conflict_ReturnsErrKeyConflict` — 소유 lane-18, 패키지 `internal/agenticfactory/ledger`. unwritten at 37b67f6d39(§B).
- RED-now: `TestExpandKeys_Conflict_ReturnsErrKeyConflict`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.
- green path: lane-18의 KeysFor·ExpandKeys 구현 커밋이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.

#### AC-AFL-012 — 겹치지 않는 키는 동시에 진행 (D4)

- 주 REQ: maps REQ-AFL-012.
- Given 키가 겹치지 않는 READY 카드 둘, When 두 세션이 차례로 `Claim`을 부를 때, Then 두 카드가 모두 claim되어 동시에 BUILDING에 있고, 어느 쪽도 전역 직렬 slot을 기다리지 않는다.
- 시험 이름(하나): `TestClaim_DisjointConflictKeys_ClaimInParallel` — 소유 lane-18, 패키지 `internal/agenticfactory/ledger`. unwritten at 37b67f6d39(§B).
- RED-now: `TestClaim_DisjointConflictKeys_ClaimInParallel`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.
- green path: lane-18의 Claim 구현 커밋이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.

#### AC-AFL-013 — 카드 흐름 claim → DONE

- 주 REQ: maps REQ-AFL-013.
- Given class A 카드 하나와 시험 저장소, When 카드를 claim → RED 증거 → GREEN → 관문 → MERGING(`Enqueue`)으로 몰 때, Then `landed` 쪽지 뒤에 카드가 DONE이고 lease가 풀려 있으며, `spec.risk`가 비고 diff가 400줄 이하라 REVIEWING을 거치지 않았다.
- 시험 이름(하나): `TestCardFlow_ClaimToDone_ThroughMergeQueue` — 소유 lane-22, 패키지 미정(통합 시험 위치, OD-35). unwritten at 37b67f6d39(§B).
- RED-now: `TestCardFlow_ClaimToDone_ThroughMergeQueue`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.
- green path: lane-22의 CLI·통합 커밋(계약 §7 P2)이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.

#### AC-AFL-014 — 병합 큐 착지와 완료 증거 (D3)

- 주 REQ: maps REQ-AFL-014.
- Given MERGING 카드와 후보 SHA, When 병합 큐가 그 카드를 처리할 때, Then 후보 worktree 검증이 잠금 밖에서 끝난 뒤 `merge-queue:landing`을 잡고, `git update-ref <trunk> <new> <old>`로 trunk를 옮기고, `landed` 쪽지를 남기고 카드를 DONE으로 옮긴다; 리더 승인 receipt는 쓰이지 않는다.
- 시험 이름(하나): `TestMergeq_Land_CASUpdateRefThenLandedNote` — 소유 lane-22, 패키지 `internal/agenticfactory/mergeq`. unwritten at 37b67f6d39(§B).
- RED-now: `TestMergeq_Land_CASUpdateRefThenLandedNote`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.
- green path: lane-22의 mergeq 구현 커밋이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.

#### AC-AFL-015 — Issue·원격 코드 퇴역

- 주 REQ: maps REQ-AFL-015.
- Given 퇴역이 끝난 트리, When AgenticFactory 패키지 그래프를 훑을 때, Then `forge`·`mirror`·`effect/ghintake`·`effect/ctxledger`·`effect/labels`·`gitref`를 가져오는 패키지가 없다.
- 시험 이름(하나): `TestRetirement_NoImportOfForgeMirrorGitref` — 소유 lane-22, 패키지 미정(OD-35). unwritten at 37b67f6d39(§B).
- RED-now: `TestRetirement_NoImportOfForgeMirrorGitref`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.
- green path: lane-22의 퇴역 커밋(패키지마다 한 커밋; claim·controlref는 lane-18)이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.

#### AC-AFL-016 — importer는 원천을 읽기만 함

- 주 REQ: maps REQ-AFL-016.
- Given backlog.db·factory.db·claim JSON 문서의 시험 사본, When importer를 dry-run으로 돌릴 때, Then 세 원천의 바이트가 실행 전과 같다.
- 시험 이름(하나): `TestImporter_DryRun_SourcesUnchanged` — 소유 lane-22, 패키지 `internal/agenticfactory/ledger/importer`. unwritten at 37b67f6d39(§B).
- RED-now: `TestImporter_DryRun_SourcesUnchanged`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.
- green path: lane-22의 importer dry-run 커밋이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.

#### AC-AFL-017 — `moai af` 동사

- 주 REQ: maps REQ-AFL-017.
- Given `moai af` 명령, When 등록된 하위 동사를 볼 때, Then claim, 단계 이동, 증거 기록, `Enqueue`, `Release` 각각에 닿는 동사가 있다(동사 이름은 미정, OD-33).
- 시험 이름(하나): `TestAFCommand_VerbsRegistered` — 소유 lane-22, 패키지 미정(CLI 패키지, OD-33). unwritten at 37b67f6d39(§B).
- RED-now: `TestAFCommand_VerbsRegistered`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.
- green path: lane-22의 CLI 커밋이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.

#### AC-AFL-018 — §6 기본값은 이름 있는 상수

- 주 REQ: maps REQ-AFL-018.
- Given 원장 패키지, When 그 상수를 읽을 때, Then lease 15 min, renew 5 min, reclaim cap 3, snapshot 30 min이 이름 있는 상수이고 값이 계약 §6과 같다(원장 밖 상수는 OD-35).
- 시험 이름(하나): `TestLedgerDefaults_MatchContract` — 소유 lane-18, 패키지 `internal/agenticfactory/ledger`. unwritten at 37b67f6d39(§B).
- RED-now: `TestLedgerDefaults_MatchContract`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.
- green path: lane-18의 원장 구현 커밋이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.

#### AC-AFL-019 — 흐름 읽기 모형은 Reader만 씀

- 주 REQ: maps REQ-AFL-019.
- Given 멈춘 lease 하나를 담은 가짜 `ledger.Reader`, When 흐름 모형을 계산할 때, Then 그 카드가 stalled로 잡히고, 계산은 그 Reader 말고 다른 입력을 쓰지 않는다.
- 시험 이름(하나): `TestFlow_StalledDetection_FromReader` — 소유 lane-20, 패키지 `internal/agenticfactory/monitor/flow`. unwritten at 37b67f6d39(§B).
- RED-now: `TestFlow_StalledDetection_FromReader`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.
- green path: lane-20의 flow 패키지 커밋이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.

#### AC-AFL-020 — 보드는 읽기 전용

- 주 REQ: maps REQ-AFL-020.
- Given 카드·lease·queue·쪽지마다 서로 다른 표지값을 담은 가짜 `ledger.Reader`, When (i) 보드 페이지를 그릴 때, Then 네 범주의 표지값이 모두 페이지에 나온다; 그리고 When (ii) 같은 시험이 `internal/web` 디렉터리(하위 포함)의 `_test.go`가 아닌 `.go` 파일을 `go/parser`로 읽을 때, Then `ledger.Store` 선택자와 `ledger.Open` 호출이 0건이다.
- 판정하지 않는 것: lease 키(출처 미정, OD-39)와 흐름 지표 표시. §E에 적었다.
- 죽여야 할 변이: 보드가 따로 `*ledger.Store`를 열어 읽거나 쓰는 구현 → (ii)에서 FAIL; Reader 값을 쓰지 않고 고정 내용을 그리는 구현 → (i)에서 FAIL.
- 시험 이름(하나): `TestBoard_ReaderOnly_NoStoreReference` — 소유 lane-21, 패키지 `internal/web`. unwritten at 37b67f6d39(§B).
- RED-now: `TestBoard_ReaderOnly_NoStoreReference`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.
- green path: lane-21의 보드 커밋이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.
- 이전 문구(0.1.0): Given 카드·lease·키·queue·쪽지를 담은 가짜 `ledger.Reader`, When 보드 페이지를 그릴 때, Then 여섯 항목(카드, lease, 키, queue, 쪽지, 흐름 지표)이 보이고 Reader에 쓰기 호출이 하나도 없다. 시험 이름 `TestBoard_FakeReader_RendersReadOnly`.

#### AC-AFL-021 — 훅 heartbeat 30 s 조절

- 주 REQ: maps REQ-AFL-021.
- Given 만든 지 30 s가 안 된 세션 stamp 파일, When 훅이 heartbeat를 쓰려 할 때, Then DB 쓰기가 일어나지 않는다.
- 시험 이름(하나): `TestHeartbeatWriter_SkipsWriteWhenStampFresh` — 소유 lane-20, 패키지 미정(훅 패키지, 계약은 'hook heartbeat writer'라고만 적음; OD-35 (k)). unwritten at 37b67f6d39(§B).
- RED-now: `TestHeartbeatWriter_SkipsWriteWhenStampFresh`가 `37b67f6d39`에 없다(§B). 이 기준은 지금 판정할 수 없어 통과로 적지 않는다.
- green path: lane-20의 훅 heartbeat 커밋(계약 §7 P2, api.go 착지 뒤)이 이 시험을 RED로 먼저 넣고(계약 머리말: RED 커밋이 GREEN보다 먼저) 같은 lane의 GREEN 커밋이 뒤집는다.

## §D 완료 정의

- 21개 시험이 각 소유 lane의 커밋으로 존재하고, RED 커밋이 GREEN 커밋보다 먼저 있으며(`verification-claim-integrity.md` §2.3: 순서는 커밋 그래프만 증언한다), GREEN 뒤 각 시험이 실행되어 통과한 출력이 그 tree SHA와 함께 기록된다.
- 실행 명령은 시험이 생긴 뒤에 그 이름만 고르는 형태로 적는다. 그 전에는 적지 않는다.

## §E 덮지 못하는 부분

- REQ-AFL-003의 "트랜잭션 안 외부 명령 없음, 에이전트 턴을 넘는 트랜잭션 없음": AC-AFL-003의 시험은 Open 쪽만 본다. 확인 방법 미정(OD-35 (j)).
- REQ-AFL-018: 원장 밖 상수(heartbeat 30 s, fix-loop 3, review 400, landing TTL 10 min)의 패키지 미정(OD-35) — AC-AFL-018은 원장 상수만 본다.
- REQ-AFL-013의 REVIEWING 경로(교차 모델 리뷰): AC-AFL-013은 REVIEWING을 거치지 않는 경우만 본다. 리뷰 수단 미정(OD-35).
- (0.2.0) AC-AFL-004의 정적 검사는 `ledger` 디렉터리 바로 아래 파일만 읽는다. `importer` 하위 디렉터리는 덮지 않고, SQL 문을 리터럴이 아닌 방법(문자열 조합 등)으로 만들면 (i)의 수집이 놓친다.
- (0.2.0) REQ-AFL-020의 lease 키(출처 미정, OD-39)와 흐름 지표 표시는 AC-AFL-020이 판정하지 않는다.
- (0.2.0) REQ-AFL-014의 trunk 전용 ref, QueueUpdate epoch 원천, merged·closed·settled 사실 원천은 AC-AFL-014의 시험 하나가 단언하지 않는다(AC-AFL-014는 0.1.0 그대로).
- (0.2.0) REQ-AFL-010의 "probe 밖, 트랜잭션 안 재확인"과 `Reap(nil)` 오류, REQ-AFL-006의 park 사유 집합은 AC-AFL-010·006이 단언하지 않는다(두 AC는 0.1.0 그대로).
