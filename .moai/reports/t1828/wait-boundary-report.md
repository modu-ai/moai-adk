# 팩토리 반복 대기·중단 — 경계별 원인표와 격리 수리 (t1828)

작성: lane-18, run tn1p4e, 2026-10-17. 정본 카드: `moai todo show t1828`.

## Claim

1. **경계 7단계 표를 작성했다**(§경계별 원인표) — 각 경계의 상태 출처·대기 사유·다음 책임자를 분리했고, ack/accepted 를 착수·완료로 승격하지 않는 기준(dispatch `started` 전이 = 착수의 유일한 표지)을 코드로 확인했다.
2. **재현된 결함은 1가지다**: 만료된 임대의 보유자가 살아 있어도 직렬 슬롯이 해제되어, 살아 있는 두 레인이 serial 카드 2장을 동시에 진행했다(실측). 이 경계만 수리했다 — 커밋 `424605284` on `WT-10-17-class`.
3. 같은 메커니즘의 실시간 관측 4건(본 세션 임대 만료·재임대, lane-9 갱신 경쟁, lane-13 첫 임대/재임대 비대칭)을 타임라인과 함께 기록했고, 거부 상세문이 차단 카드를 밝히지 않는 관측성 공백을 함께 수리했다(거부문이 카드·보유자 명시).
4. 카드가 지정한 기존 시험 5종을 실행 출력으로 재사용했고 전부 통과했다 — 브로커 경계(exactly-once·fencing·lifecycle)는 현 트리에서 재현된 결함이 아니다.

## Evidence (명령 + 출력)

### D1 결함 재현 (RED, 수리 전)

- 명령: `go test ./internal/cli/ -run TestFactoryNextRefusedBehindExpiredLeaseWithLiveHolder -count=1` (TMPDIR=/private/tmp 격리, MOAI_KANBAN_* unset)
- 출력(발췌): `--- FAIL: TestFactoryNextRefusedBehindExpiredLeaseWithLiveHolder (22.88s)` … `factory next leased a second serial card behind an expired lease whose holder is alive` / `FAIL github.com/modu-ai/moai-adk/internal/cli 26.256s`

### 운영 사건 타임라인 (스냅샷 db + events + git log, 모두 UTC)

| 시각 | 사건 |
|---|---|
| 10:42:43 | lane-1 T3 t1825 leased(만료 10:57:43); workers.lane-1 하트비트 10:42:43 정지 |
| ~10:44 | lane-5 `next --card t1826` → `refused serial-slot`(t1825 만료 전 — 올바른 거부) |
| 10:57:43 | t1825 임대 만료 — **lane-1 프로세스 생존**(pid 21334, ps 실측) |
| 10:59:37 | lane-13 T3 t1697 leased — 만료만으로 슬롯 free 판정, **통과(결함)** |
| 11:09:16 | lane-13, t1697 브랜치 커밋 `964f788c3` |
| 11:22:39 | lane-1, t1825 브랜치 커밋 `32f7e64be` — **serial 2장 동시 진행 실측** |
| 11:30:57→11:56:06 | t1697 자체 임대 만료·수거(→assigned) 후 재임대 `refused serial-slot` |
| 12:01:03 | t1825 임대 만료 수거(leased+expired 행이 64분간 존속 후) |

### 수리 후 GREEN (대리자 실행 + 레인 독립 재실행)

| -run 패턴 | 대리자 출력 | 레인 재실행 |
|---|---|---|
| TestAutoLaneCycle | `ok internal/cli 90.775s` | —(장시간, 부하 규칙상 슬롯 필요) |
| TestFactorySerialSlotHeld | `ok internal/cli 1.353s` | `--- PASS TestFactorySerialSlotHeldRequiresADriver (0.00s)` `ok 1.012s` |
| TestFactoryNext | `ok internal/cli 715.949s`(join-guard·nominate·신규 3종 포함) | —(장시간) |
| TestFactoryNextReNominationAfterOwnExpiryStillRefused | `--- PASS (1.71s)` | `--- PASS (2.10s)` `ok 3.162s` |
| TestFactoryNextRefusedBehindExpiredLeaseWithLiveHolder | RED→GREEN | —(동일 시험, 대리자 실행) |

gofmt -l 3파일 = 0행(양쪽 모두), `go build ./internal/cli/` OK, `go vet ./internal/cli/` exit 0(대리자).

### 기존 시험 재사용 (수리 전 현 트리 93f4fc75c)

- `go test ./internal/factorymsg/ -run TestDispatch -count=1 -v` → 5종 전부 PASS: TestDispatchResultExactlyOnce(하위 8 포함) 5.36s, TestDispatchResultOrderTable 0.36s, TestDispatchLifecycleTransitions 0.34s, TestDispatchStaleFencingAndSuperseded 0.30s, TestDispatchReassignmentFencesLateResult 0.62s — `ok 7.784s`
- `go test ./internal/cli/ -run TestAutoLaneCycleAssignedCardBeforeQueuedRanking -count=1 -v` → PASS 10.45s(`ok 12.078s`)
- `go test ./internal/hook/ -run TestFactory -count=1` → 1차 FAIL 456.686s(부하) → 재실행 PASS `ok 112.137s` — 비결정적 타이밍 플레이크로 판정

## Baseline-attribution

- 분석·수리 대상 트리: 카드 워크트리 `.moai/worktrees/t1828`, 브랜치 `WT-10-17-class`, base `93f4fc75c`(=당시 local main), 수리 커밋 `424605284`. 판정 빌드: go1.26.8 darwin/arm64, 카드 트리 소스를 직접 컴파일(`go test`)해 귀속 — 설치 moai 바이너리(v3.2.0-rc.36-t1822c, ee981a932-dirty)는 판정에 쓰지 않았다.
- 운영 증거: `~/.moai/db/moai-adk-go-1bd3d038/factory/factory.db`의 **복사본 스냅샷**(원본 -wal/-shm 부재 확인, db 단독 복사 = 완전 스냅샷) — 운영 DB 직접 수정·재주입 없음. ps(lane-1 pid 21334 생존)·git log(t1825 `32f7e64be`, t1697 `964f788c3`).
- 이전 분석 `/tmp/moai-adk-factory-wait-analysis-20261017.md`(primary 478285ab4·main c3aa6b057 기준)의 원인 후보 표는 임시 파일이므로, 본 보고서가 필요 내용(경계표·원인·가설 구분)을 자급한다.

## 경계별 원인표 (최종)

| 경계 | 상태 출처 | 대기 사유 | 다음 책임자 |
|---|---|---|---|
| 요청→배차 | backlog.db(SQLite live 큐; backlog.json 아님) | 운영자 pick 전 queued | 운영자/리더 |
| 배차→팩토리 기록 | factory.db cards+events(T1 picked→T2 assigned→T3 leased, version fencing) | 리더 assign 후 레인 T3 전 | 리더→레인 |
| 기록→메시지 | factorymsg store(pending→claim 30s→receipt disposition+ack, TTL dead-letter) | 턴 경계 전달 지연(degraded 창 관측 2회) | 브로커/수신 세션 |
| 메시지→착수 | dispatch assigned→delivered→**started**(수신자 명시 전이; 영수증은 delivered 넘기지 않음) | 착수 전이 미발생 | 수신 세션 |
| 착수→완료/실패/질문 | stage 전이+evidence_path; 만료 후 갱신은 룰링 d-20261010T095731Z-5ddc 로 거부 | 긴 턴이 임대기한(15분) 초과 + 갱신 cron 은 idle 발화 | 레인(stage 수거→재임대 복구, 실측) |
| 완료→다음 작업 | factorySerialSlotFree 상태 집합(merge-ready+ 종료 상태 해제) | 통합 창 직렬화 | 레인/리더 |
| stale/예외→복구 | **D1(수리됨)**: 만료+생존 보유자 슬롯 해제 → 제2 serial 임대 | — | 본 카드(완료) |

## 원인 정리 (가설 vs 재현)

**재현(수리 대상)**: D1 — `factorySerialSlotHeld`가 `!LeaseExpired(now)`만으로 해제. RED 시험으로 재현, 수리로 소거.

**가설(미재현 — 기록만)**:
- run 스코핑: 슬롯 판정이 `ListCards(ctx, runID)` 현재 run 행만 읽음 — 은퇴 run 잔존 in-flight 행이 새 run 판정에 안 보임. 본 사건 4건은 모두 같은 run(tn1p4e) 행이라 이 원인 불필요. 재현되면 t1480 연결 제안.
- Stop 훅이 degraded 상태를 버림(user_prompt_submit.go/stop.go) — 관측성 가설, 미측정.
- hook budget 시험의 부하 민감성(1차 FAIL/2차 PASS) — 비결정성만 확인, 원인 미상.

**관측(원인 판정 아님)**: 본 세션 임대 만료→갱신 거부→stage 수거→재임대 복구 전 과정(8859 순서 실측), lane-9 갱신 경쟁(리더 전달), lane-13 재임대 비대칭(두 시점 t1825 행 상태 차이로도 설명 가능 — 리더의 assignedHere 가설과 병립), lane-5 맨 next t1780 임대(mode 미확인).

## 수정 파일 (커밋 424605284)

- `internal/cli/factory_card.go` — `factorySerialSlotHeld`에 보유자 생존 조항(만료 임대는 roster pid 생존 시 유지, 부재/사망/비양수 해제 — t1407 no-wedge 보존), `factorySerialInFlightExcluding`→`(차단ID, bool)` 반환, `factoryWorkerPidAlive`(workers.pid + `factory.FactoryProcessAlive` 재사용), db 소유 호출부 2곳에서 클로저 스레딩, `factoryKeepSetRefusal` 거부문에 차단 카드·보유자 명시.
- `internal/cli/factory_serial_slot_liveness_test.go`(신규) — 사례 (a) 생존 보유자 만료 임대 거부(RED 근거), (b) 자체 만료 수거 후 재임대도 대칭 거부(실제 `db.Transition` T27/T28 구동), 사망 보유자 해제 고정(roster 부재 + pid −1), 차단자 명시 단언.
- `internal/cli/factory_lane_join_guard_test.go` — 신규 서명 행렬 갱신(생존/사망 만료 임대행, 만료 시에만 판독자 호출 프로브).

새 상수·스키마·설정 키 없음. 범위: internal/cli 3파일 + 본 보고서.

## 남은 결정 · 다음 담당자

| 항목 | 담당자 |
|---|---|
| 만료 후 갱신 허용 여부(룰링 d-20261010T095731Z-5ddc 유지) 재검토 | 운영자/리더 — 본 카드는 현행 유지 |
| run 스코핑(은퇴 run 행 가시성) — 재현 시 t1480 연결 | 리더(카드 연결 제안) |
| t1821 기록: (a) 자기 assigned 우선은 현 main·설치 빌드에 이미 존재(옛 결함), (b) 본 관측은 역방향(만료로 점유 해제)이며 assigned 무임대 점유는 t1407 룰링대로 유지 | 본 카드 착지 시 t1821 에 기록 |
| CI(원격) 검증 — 본 카드는 merge-ready 정지로 push 없음 | 리더(통합 시점) |
| 외부 런타임 가드 차단(t1824)·감독기(t1482)·레인 자가복구(t1715) | 별도 카드 |

## Gaps

- **실제 외부 런타임(Claude Code 바이너리)은 재현 대상이 아니다** — t1824 범위. 본 카드는 ADK 수정 가능 범위만 다뤘다(거부 사유·오류·재개 안내 분리는 거부문 명시로 일부 반영).
- 커버리지 백분율 미측정(-cover 미실행 — 범위 시험 4종이 검증 증거). 원격 CI 미실행(merge-ready 정지, push 금지).
- TestAutoLaneCycle·TestFactoryNext 전체 계열의 재실행은 부하 규칙(슬롯)상 생략, 대리자 1회 실행 출력만 귀속.
- lane-13 재임대 비대칭의 정확한 차단자 특정 불가 — 거부문이 그 시점엔 카드를 명시하지 않았음(수리 후에는 명시됨).
- hook 패키지 FAIL 1회의 실패 시험명 미확보(재실행 PASS로 판정 마무리).

## Residual-risk

- 생존 판정이 roster(workers) pid 기반: 레인 프로세스가 살아 있되 그 카드를 버린 경우(예: 재부팅 후 새 세션이 같은 pid 없음)는 해제되나, **pid 재사용** 시 죽은 보유자가 살아 있는 것으로 읽힐 가능성(프로세스 시작시각 미확인 — factorymsg `ownerCurrent`는 start 대조하나 본 수리는 pid 신호만 사용). 빈도 낮음, t1407 방향과의 균형에서 선택.
- 만료 임대가 살아 있는 보유자에게 무한 유지되면, 보유자가 실제로 죽어도 pid 사망 전까지 다른 serial 카드가 대기 — 갱신 거부 룰링과 결합 시 회복이 stage 수거 순서에 의존(본 세션에서 실측한 복구 경로).
- 18-레인 부하에서 타이밍 민감 시험의 비결정성(hook 1차 FAIL)은 환경 잔여 위험으로 남는다.
