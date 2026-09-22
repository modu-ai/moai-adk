# t396 — AC-BLE-004 두 번째 공허화 조건 판정 (verdict)

- **브랜치**: WT-ble4-vacuity-count (develop 기반)
- **카드**: t396 lane-10 (G2b 순차 6장째), Tier S · 선행 t379 착지(develop 5928095ea) — 현재 develop은 그 후속이므로 충족

## Claim

AC-BLE-004("실패 집합 불변")의 두 번째 공허화 조건은 **축별로 갈린다**: 세 소비자 축 중 2축(board 락·backlog 락)은 경합 테스트가 실재해 비공허, **mutation 락 축은 `//go:build windows` 전용 테스트 2개뿐이라 darwin/linux 스윕에서는 0** — 이 축에 한해 공허화 조건이 성립한다. 처우: AC 재배치(신규 테스트 작성)가 아니라 **공허화 조건을 SPEC에 추가**하는 것으로 닫는다.

## Evidence

### 세 소비자 식별 (this tree)

| 축 | 소비자 | 상위 공개 API |
|---|---|---|
| 1 | `acquireBoardLockSerialized` (board_store.go:165) | `WriteBoardState`(:226), `RecoverBoard`(board_recover.go:69) |
| 2 | `acquireIntegrationMutationLock` (integration_lock_mutation.go:95) | `withIntegrationLockMutation`(:82) |
| 3 | `(*BacklogStore).acquireLock` (backlog_store.go:728) | backlog mutate 경로(:618, :639) |

### 경합 테스트 계수 (internal/kanban/*_test.go, 이번 실행)

| 축 | 지나는 경합 테스트 | 수 |
|---|---|---|
| 1 | TestBoardMutation_SerializedAcrossProcesses · TestBoardMutation_ConcurrencyPositiveControl · TestWriteBoardState_ConcurrentReaderSeparateProcess (board_lock_cross_test.go — 별도 프로세스 재실행, unix) · TestWriteBoardState_ConcurrentReaderSeesWholeBoards (board_store_test.go:158, goroutine) | **4** |
| 2 | integration_lock_mutation_windows_test.go 2개 — 파일 선두 `//go:build windows` 직접 관측 | **darwin/linux 0 · windows 2** |
| 3 | TestBacklogMutate_LoserSerializedNotFailed · TestBacklogConcurrentAdd_UniqueIDs · TestBacklogLock_TimeoutNamesLockPath (backlog_store_test.go) | **3** |

t379의 AC-BLE-004 측정 방법 확인(progress.md:175): `go test ./internal/kanban/... -count=1` 패키지 전체 스윕 전후, 실패 집합 전후 모두 공집합. **이 스윕이 darwin에서 돌면 축 2는 아무것도 지나지 않는다** — lane-5 자가 신고의 "세지 않았다"는 지적이 축 2에서는 정확하다.

## 판정 — 처우

- **"0이면 공허"의 성립 여부**: 전체로는 9개(부분 플랫폼 차이), 그러나 AC가 단언하는 것은 "세 소비자를 지나는" 실패 집합이므로 **축 2는 darwin/linux 기준 0** → 조건 성립.
- **처우 선택 — 공허화 조건 추가(채택), AC 재배치(기각)**: 신규 경합 테스트 작성은 mutation 락의 darwin 경합 시나리오 설계(프로세스 재실행 plumbing)를 요하는 별개 작업이고, 카드의 두 갈래 중 "공허화 조건을 SPEC에 추가"가 계수 사실을 기록하는 최소 변경. 재배치(테스트 보강)는 후속 카드 축.
- SPEC-BOARDLOCK-ERRNO-001/acceptance.md §AC-BLE-004에 조건 2 추가: *"세 소비자를 지나는 경합 테스트가 판정 플랫폼에서 0이면 이 AC는 공허하다 — 현재 acquireIntegrationMutationLock 축은 windows 전용 테스트 2개뿐이므로 darwin/linux 스윕은 이 축을 지나지 않는다(계수: board 4 · backlog 3 · mutation windows 2/darwin-linux 0)."*

## Gaps (미검증)

- windows 빌드에서 mutation 락 테스트 2개가 실제로 통과하는지는 실행 안 함(GOOS cross-compile 검증은 t385 계열 카드 층) — 카드 판정에 불요.
- RecoverBoard 경유 경합 테스트는 별도 미계수 — WriteBoardState 경유 4개로 축 1이 이미 비공허 판정이라 판정에 영향 없음.

## Residual-risk

- mutation 락 darwin 경합 테스트 부재는 이 카드가 남기는 실 부채 — 통합 창 유실 방지용 커버리지 축이라 필요해지면 별도 카드.
