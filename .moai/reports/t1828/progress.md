# t1828 진행 기록 (lane-18, Class B run→sync)

카드: t1828 [운영자 승인 10-17·Class B·P1] 팩토리 반복 대기·중단 — 경계별 원인표와 재현되는 결함만 격리 수리
run: tn1p4e · lane: lane-18 · 트리: `.moai/worktrees/t1828` · 브랜치: `WT-10-17-class`
Baseline: 카드 트리 HEAD `93f4fc75c`(=local main), go1.26.8 darwin/arm64. 설치 moai 빌드는 `v3.2.0-rc.36-t1822c`(ee981a932-dirty, 2026-10-15T21:56Z) — 본 카드의 판정은 카드 트리 소스를 go test 로 직접 컴파일·실행해 귀속(§2.2 도구 출처).

## 1. 관측·재현 정리 (가설 vs 재현 분리)

### 재현된 결함 D1 — 임대 만료만으로 직렬 슬롯 해제(보유자 생존 미판정)

운영 팩토리 DB 스냅샷(db 단독 복사본, `-wal`/`-shm` 부재 확인) + events 테이블로 타임라인 고정(모두 UTC):

| 시각 | 사건 | 증거 |
|---|---|---|
| 10:42:02 | 리더 assign t1825→lane-1, t1826→lane-5 (mode=serial) | events `card.transition` actor=assign |
| 10:42:43 | lane-1 T3: t1825 assigned→leased (임대 만료 10:57:43) | events; workers.lane-1 heartbeat_at=10:42:43 (이후 정지) |
| ~10:44 | lane-5 `next --card t1826` → `refused serial-slot` (t1825 만료 전) | 리더 전달 + t1826 행 abandoned 10:45:04 |
| 10:57:13 | 리더 assign t1697→lane-13 | events |
| 10:57:43 | t1825 임대 만료 — lane-1 프로세스 생존(pid 21334, ps 실측) 상태에서 미갱신 | cards 행 + ps |
| 10:59:37 | lane-13 T3: t1697 assigned→leased — **슬롯 free 판정으로 성공** | events; `factorySerialSlotHeld`가 `LeaseExpired`만으로 해제 |
| 11:09:16 | lane-13, t1697 브랜치(WT-lane-collab) 커밋 `964f788c3` | git log |
| 11:22:39 | lane-1, t1825 브랜치(WT-factory-run-select) 커밋 `32f7e64be` — **두 serial 카드 실동시 진행 실측** | git log |

원인(코드): `internal/cli/factory_card.go:295` `factorySerialSlotHeld = !free(state) && !LeaseExpired(now) && owner != ""` — 만료 임대의 보유자가 살아 있는지 판정하지 않는다. 긴 턴 중인 레인은 갱신 cron 이 발화하지 않고(교훈 t1687: cron 은 idle 에만 발화), 만료 후 갱신은 룰 d-20261010T095731Z-5ddc 로 거부된다(`card_transition.go:500` ErrLeaseExpired 무기록 거부). 결과: 생존 레인의 작업 보호가 만료선에서 끊긴다.

### 동일 메커니즘의 실시간 재현 3건 (관측)

4. **lane-13 재임대 비대칭(리더 전달, 11:3x~12:0x)**: t1697 첫 임대(10:59:37, 슬롯 free 통과) → 자체 임대 만료 11:30:57(수거 11:56:06→assigned, owner 유지) → 재임대 `refused serial-slot`. 타임라인 고정(스냅샷 12:0x): t1825 임대는 10:57:43 만료에서 **12:01:03까지 leased+expired 로 미수거** 존속. 리더 가설(assignedHere/ignoreAssigned 경로 차이)과 별개로 확인된 구조적 사실: (i) 거부 상세문이 차단 카드를 명시하지 않아 리더가 점유자를 추정해야 했음 — 관측성 공백, (ii) 재임대 시점에 t1825 leased+expired(uncollected)면 ignoreAssigned 의 예외 대상(merely assigned)이 아니므로 판정에 잡힘 — 첫 임대(통과)와 재임대(거부)의 비대칭은 두 시점의 t1825 행 상태 차이로도 설명 가능. 회귀 사례 (b)(재임대 경로) + 거부문 차단자 명시를 대리자 범위에 추가 전달(11:59Z·12:04Z 메시지).

1. **lane-18(본 세션)**: 조사 턴 중 임대 만료 11:16:01Z → `factory renew t1828` 거부("card lease expired") → `factory stage t1828 run` 시도가 만료를 지연 수거(이벤트 `lease.expired` 11:27:36, version 5, 행→assigned) → MCP `factory_next --card t1828` 재임대 성공. 8859 복구 순서 실측 확인.
2. **lane-9(t1752)**: 갱신 cron 이 만료선 수 초 뒤 발화 → 동일 거부 → 동일 복구(stage 수거→재임대→renew 성공, stage=plan 보존). 리더 전달(10-17).
3. **lane-5 맨 next → t1780 임대(10:42:50)**: 같은 창에서 비지명 임대도 성공 — t1780 분류(mode) 미확인, 관측으로만 기록.

### 미재현·관측만 (수리 대상 아님)

- run 경계 스코핑: 슬롯 판정이 `ListCards(ctx, runID)` 로 현재 run 행만 읽음(`factory_card.go:817/1379`) — tmvgvn(은퇴) 잔존 in-flight 행은 새 run 판정에 보이지 않음. 본 관측 2건은 이 원인 불필요(t1825·t1697 모두 tn1p4e 행). Gap 으로 기록.
- hook inbox degraded: 본 세션 메시지 경계에서 "factory messaging degraded: inbox context deadline exceeded" 실관측(전달은 결국 성공). 이전 분석의 Stop 이 degraded 상태를 버린다는 관측성 가설(user_prompt_submit.go:221 / stop.go:88)은 미측정 — 가설로 유지.
- 외부 Claude Code 런타임 worktree 가드의 복합 명령 거부: 본 세션에서 `GIT_CONFIG_GLOBAL` 주입형 2건 거부 실측(정당 차단으로 대응, 격리 수단 변경). 생산자는 외부 바이너리 → t1824 범위.

## 2. 경계별 원인표 (초판 — 최종본은 wait-boundary-report.md)

| 경계 | 상태 출처 | 대기 사유 | 다음 책임자 |
|---|---|---|---|
| 요청→배차 | backlog.db(SQLite, live 큐; backlog.json 아님) | 운영자 pick 전 queued | 운영자/리더 |
| 배차→팩토리 기록 | factory.db cards + events (T1 picked→T2 assigned→T3 leased, version fencing) | 리더 assign 후 레인 T3 전 | 리더→레인 |
| 기록→메시지 | factorymsg store (pending→claim 30s→receipt disposition+ack, TTL dead-letter) | 턴 경계 전달 | 브로커/수신 세션 |
| 메시지→착수 | dispatch assigned→delivered→**started**(수신자 명시 전이, 영수증은 delivered 넘기지 않음 — dispatch.go:124) | hook inbox degraded 창 지연 가능(관측) | 수신 세션 |
| 착수→완료/실패/질문 | stage 전이 + evidence_path; 만료 후 갱신 거부(룰) | 긴 턴이 임대기한 초과 | 레인(stage 수거→재임대 복구) |
| 완료→다음 작업 | merge-ready 이후 슬롯 해제 상태 집합(factorySerialSlotFree) | 통합 창 직렬화 | 레인/리더 |
| stale/예외→복구 | **D1 재현**: 만료+생존 보유자가 슬롯 해제 → 제2 serial 임대 | — | **본 카드가 수리** |

ack/accepted 를 착수·완료로 승격하지 않음: dispatch `started` 전이와 첫 진행 증거가 착수의 유일한 표지(코드 주석으로 보장: "a message receipt does not perform them").

## 3. 기존 시험 실행 출력 (카드 지정분)

- `go test ./internal/factorymsg/ -run TestDispatch -count=1 -timeout 15m -v` (TMPDIR 격리) → **전부 PASS**: TestDispatchResultExactlyOnce(하위 8: duplicate_delivery, lost_receipt_redelivery, sender_restart_after_result, sender_restart_lane_scope, broker_restart, old_generation_redelivery, collision 포함) 5.36s, TestDispatchResultOrderTable 0.36s, TestDispatchLifecycleTransitions 0.34s, TestDispatchStaleFencingAndSuperseded 0.30s, TestDispatchReassignmentFencesLateResult 0.62s — `ok internal/factorymsg 7.784s`, exit 0.
- 판정: 브로커 경계(exactly-once·fencing·lifecycle)는 현 트리에서 재현된 결함 아님.
- 전체 패키지 실행 1건은 10m 타임아웃으로 실패(`TestFactoryLaneHandoffRebindVsUserPromptRegisterRace/unforced_200_iterations` 단독 6m45s+, 부하 중 기계 측정) — 제품 결함 아님, 범위 재실행으로 대체.
- `go test ./internal/cli/ -run TestAutoLaneCycleAssignedCardBeforeQueuedRanking -count=1 -timeout 15m -v` (TMPDIR 격리) → **PASS 10.45s**, `ok internal/cli 12.078s`, exit 0. 판정: 자기 assigned 우선 임대 경계는 현 트리에서 재현된 결함 아님(카드 지시 (a) 항과 일치).
- `go test ./internal/hook/ -run TestFactory -count=1 -timeout 15m` (TMPDIR 격리) → 1차 **FAIL 456.686s**(부하 중, 실패 시험명 미포함) → 재실행 2차 **PASS `ok internal/hook 112.137s`** (같은 명령, 같은 환경). 판정: 비결정적 timing 플레이크(부하 시) — 재현된 결함 아님, 수리하지 않음. 두 실행 모두 기록(§3 baseline). timing 민감성은 잔여 위험으로 보고서 기재.
- 본 세션 메시지 경계에서 "factory messaging degraded: inbox context deadline exceeded" **2회 실관측**(10:5x, 11:3x 근방) — 전달은 모두 결국 성공. degraded 창이 모델 턴 시작을 지연시키는지는 미측정(가설 유지).
- 임대 갱신: 11:35Z 선제 갱신 성공(만료 11:53:00Z) — 긴 턴 사이 갱신 규율 준수. 11:40:39Z 워치독 갱신 재성공(만료 11:55:39Z).

## 4. 결정 기록

- D-20261017-1: 수리 범위를 D1(만료 임대 보유자 생존 판정 추가)로 확정. 근거: 유일하게 재현된 결함, 리더 관측 2건 + 실시간 재현과 정합. `factory.FactoryProcessAlive`(internal/factory/factory_alive_unix.go:22) + workers.pid 재사용, 새 상수·스키마 변경 없음, worker 행 부재/pid 사망 시 기존 해제 유지(t1407 no-wedge 보존).
- D-20261017-2: 만료 후 갱신 허용 여부는 기존 룰링(d-20261010T095731Z-5ddc)이 있으므로 변경하지 않음 — 남은 결정으로 기록.
- D-20261017-3: run 경계 스코핑은 재현 없으므로 수리하지 않음(관측 기록만). 필요성이 재현으로 입증되면 t1480(큐·기록 투영) 연결 제안.
- t1821 기록: (b) lease-less assigned 의 직렬 슬롯 점유 — 본 관측은 역방향(만료로 점유가 풀림)이며, assigned 무임대 행의 신규 카드 arm 점유는 기존 설계(t1407 룰링)로 유지. 본 카드 착지 시 t1821 에 결과 기록 예정.

## 6. 워치독 대기 기록

- 11:40:39Z — 명시적 대기: awaited-delegate(manager-develop RED+수리 진행 중, 커밋 미착지), 재점검점 = 다음 크론 발화. renewal=renewed(만료 11:55:39Z). decision record: decided_by=lane-18-watchdog evidence_refs=.moai/reports/t1828/progress.md ladder_path=§3.5
- 11:45:36Z — 대기 1회차: HEAD·작업트리 무변경(스폰 후 12분), 대리자 종료 징후 없음. 갱신 renewed(만료 12:00:31Z). 다음 발화 무기록 시 2회 연속 → 구조화 블로커. hook 재실행 PASS로 hook 경계 판정 확정(플레이크).
- 11:49:59Z — 대기 2회차(연속 무기록) → §3.5 블로커: 대리자에게 SendMessage 상태 확인·재개 요청 발송(요청: RED+최소 수정+범위 시험+커밋; 증거: git status 비어 있음, HEAD 93f4fc75c 불변, 완료 통보 없음). 갱신 renewed(만료 12:04:58Z). 응답 없고 다음 발화까지 무기록이면 재위임 판정.
- 11:54:53Z — 대리자 활동 감지: RED 시험 파일 `internal/cli/factory_serial_slot_liveness_test.go` 등장(untracked, 커밋 전). §3.5 증거 이동 중 → 명시적 대기 유지, 재점검점 = 다음 크론. 갱신 renewed(만료 12:09:53Z).
- 11:59:27Z — 대리자 GREEN 진행 중: `factory_card.go` 수정(서명에 생존 판정자 파라미터 추가, 호출부 2곳 과도 미갱신 상태 — 대리자 작업 중). 명시적 대기 유지. 갱신 renewed(만료 12:14:25Z).
- 12:04:24Z — 대리자 GREEN 지속: 호출부 컴파일 오류 소멸, `factory_lane_join_guard_test.go` 갱신 완료(서명 전파). 범위 시험 단계 진입으로 판단. 갱신 renewed(만료 12:19:22Z). 관측 4(재임대 비대칭) 반영·회귀 사례 (b)+거부문 차단자 명시 대리자 전달 완료.
- 12:09:03Z — 상태 동결(파일 집합 불변·커밋 없음): 대리자 범위 시험 실행 중으로 판단(스폰 후 36분, 정상 범위). 명시적 대기 유지. 갱신 renewed(만료 12:24:02Z). 다음 발화 동결 지속 시 재문의.
- 12:1xZ — 대리자 429 레이트 리밋으로 조기 종료(accidental-stop): 최종 보고 "gofmt·vet 클린, T27/T28 만료 수거 전이 확인(leased→assigned), 시험 구동법 확인 중". SendMessage 재개 1회 실행(남은 단계 명시: 시험 (a)/(b)+사망 방향 고정, 거부문 차단자 명시, 범위 시험, gofmt, 커밋). decision record: decided_by=lane-18-watchdog evidence_refs=delegate-last-report ladder_path=§3.5-revive-1
- 12:14:12Z — 재개 후 3분, 파일 집합 동일(시험 실행·커밋 전 단계 판단). 명시적 대기. 갱신 renewed(만료 12:29:10Z).
- 12:16Z — 재개 1회 직후 대리자 429 재발(두 번째, glm-5.3 계정 수준 한도 활성). 즉시 재개 보류 — 다음 크론 발화(~12:20Z)에 냉각 후 재개 2/2 실행 예정(마지막 재개 기회). 본 세션 모델도 같은 한도라 자체 수행은 대안 아님. 작업물은 디스크에 안전(3파일, gofmt·vet 클린 보고).
- 12:18:48Z — 냉각 후 재개 2/2 실행(마지막). 트리 상태 재확인: 3파일 변동 동일, 커밋 없음, 갱신 renewed(만료 12:33:47Z). 이 재개도 429로 끝나면 구조화 블로커를 리더에게 보고(남은 판단: 재위임 시점·모델 대안).
- 12:23:12Z — 재개 2/2 생존·작업 중: `factoryKeepSetRefusal` 서명 확장(차단자 명시 파라미터) 과도 상태 관측. 명시적 대기. 갱신 renewed(만료 12:38:11Z).
- 12:27:54Z — 호출부 전이 완료(과도 오류 소멸). 커밋 전 — 범위 시험 실행 단계 판단. 명시적 대기. 갱신 renewed(만료 12:42:53Z).
- 12:32:42Z — 파일 집합 ~10분 동결·중단 통보 없음(재개 2/2 생존 = 시험·검증 실행 국면 판단). 명시적 대기. 갱신 renewed(만료 12:47:42Z). 다음 발화에서도 동결·통보 없으면 상태 문의 1회.
- 12:37:30Z — 동결 ~15분 지속·통보 없음 → 상태 문의 1회 발송(단계 확인·거부/레이트리밋 여부). 갱신 renewed(만료 12:52:28Z). 응답·통보 없이 다음 발화까지 동결이면: 산출물 부분 상태로 판단해 인수(in-take) 검토 개시.
- 12:4xZ — 대리자 응답: 자체 백그라운드 시험 실행 대기 중(출력 버퍼링, 종료 시 자동 재개·재통보). 파일 동결 원인 해명 — 범위 시험 백그라운드 실행. 별도 조치 없음, 통보 대기.

## 5. 산출물

- 수리 커밋: `424605284` on `WT-10-17-class` (fix(factory): hold the serial slot behind an expired lease while its holder lives) — RED 재현→GREEN(범위 시험 4종 통과)→커밋. 대리자 보고서 원문은 세션 기록, 요지 본 파일 §1·보고서에 반영.
- 최종 보고서: `.moai/reports/t1828/wait-boundary-report.md` — 5단 구조(Claim/Evidence/Baseline/Gaps/Residual-risk) + 경계별 원인표 + 가설/재현 분리 + 수정 파일 + 남은 결정·다음 담당자.
- card-review: `.moai/reports/t1828/card-review.md` (진행 예정)
- 레인 독립 검증(Trust-but-verify): TestFactorySerialSlotHeldRequiresADriver PASS 0.00s(ok 1.012s), TestFactoryNextReNominationAfterOwnExpiryStillRefused PASS 2.10s(ok 3.162s), gofmt 3파일 0행 — 12:47Z.
- 정지 지점: merge-ready (factory complete 호출 금지, main 병합·push 금지 — 리더 지시 10-17)
