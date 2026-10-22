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
- 12:42-12:47Z — 수리 커밋 `424605284` 착지(RED 재현→GREEN 범위 시험 4종 통과 보고). 레인 독립 검증: 신규 시험 2종 PASS+gofmt 클린. 보고서 `wait-boundary-report.md` 작성, 증거 커밋 `c741e22ce`(force-add).
- 12:5xZ — card-review round 1 (codex scope=card): **FAIL(advisory) P2 1건** — label 재할당 시 새 프로세스 생존이 옛 보유자 생존으로 오인(factoryWorkerPidAlive 가 label 만 조회). 지적 타당으로 판단, 대리자 재위임: card.HeartbeatAt vs workers.registered_at 세대 판별 최소 수리(RED 먼저: 재등록 세대에서 슬롯 해제). `.moai/reports/t1828/card-review.md` round 1 기록. 재심사(round 2) 대기.
- 13:0xZ — P2 수리 커밋 `34212a4be`(세대 판별, RED 양방향→GREEN 6종). 레인 독립 검증 PASS(재등록 세대 시험 1.55s). **card-review round 2: PASS(결함 0건)**. 다음: 문서 마감 커밋 → 단계 전이 run→sync→merge-ready 정지.
- 13:1xZ — 문서 마감 커밋 `4576923e7`(card-review round 1-2 기록·보고서 수정). 단계 전이: `run` v7 → `sync` v8 → `sync-audit` v9(증거 4576923e7:.moai/reports/t1828/wait-boundary-report.md). merge-ready 전이는 T13 게이트에서 sync-audit 판정 파일 요구 → sync-audit-4dim 워크플로 백그라운드 실행 중(Class B 표면: 카드 정본=수용기준, 무거운 재실행 금지·기록 출력 판정).
- 13:2xZ — sync-audit-4dim 판정 **INCOMPLETE**(무-SPEC 카드는 spec_id 주입 불가 — 지정 대체 경로: 콜드 sync-auditor가 카드 델타 표면의 판정 소유자). 콜드 sync-auditor 스폰(93f4fc75c..HEAD 4커밋, 수용기준=카드 정본, 저렴 시험만·기록 출력 판정, 판정 파일 .moai/reports/t1828/sync-audit-cold.md). 통보 대기.
- 15:32-15:34Z(10-18 00:3x KST) — 긴 공백 복구: sync-auditor가 5시간 사용량 한도(재설정 23:30:34 KST)로 사망, 판정 파일 미작성. 임대 13:19:11Z 만료 → renew refused-expired → stage 시도가 **committedRefusal 형태로 만료 수거 동반**(행→assigned, stage=sync-audit 보존, 15:33:01) → MCP factory_next 재임대 성공(stage=sync-audit). 8859 복구 순서 2회차 실측(이번엔 거부+수거 동시 형태). 한도 재설정 경과로 sync-auditor SendMessage 재개(교차모형 codex REQUIRED 계획 유지, 판정 파일 대기).
- 15:49-15:5xZ — 콜드 sync-audit 판정 파일 `.moai/reports/t1828/sync-audit-cold.md`: **FAIL — F1[P1, blocking]**(필수 codex 게이트 재현 + 감사 코드 확인): 타 레인 stage 로 만료 수거 → assigned 행이 ignoreAssigned skip 에서 제외 → 세대 인지 생존 판정 미실행 → 제2 serial 임대 우회. 카드 정본 t1821-b 항("재현되면 여기서 고친다")의 재현 형태. F2[P3]: card-review.md 의 P2 수리 RED verbatim 미기록(보강 예정). F3[P4]: 조치 불요. 재심사 범위 F1 델타로 한정 — 대리자 3차 수리 재위임(RED 먼저: 외부 stage 수거→assignedHere 지정→거부; 순수 배정 행 skip 유지로 t1407/t1542 웨징 측정 보존).
- 15:56-16:01Z — 3차 수리 커밋 `7b2e9cd2c`(skip 순수 배정 행으로 축소 + assigned-with-heartbeat 조항, RED 재현→GREEN 12종·인접 5종 개별). 레인 독립 검증 PASS 3.14s. F2 디스크 보강 완료(card-review.md RED verbatim 추가, 커밋 대기).
- 16:1xZ — 재심사(round 2): **FAIL — F4[P2, blocking]**(감사+필수 codex 이중 확증, 이번 델타 회귀): ReassignCard(card_reassign.go:125-127)가 소유자 변경 시 HeartbeatAt 유지 → 재배정 미착수 행의 낡은 하트비트가 점유 판정 발동 → 만료 serial 2장 상호 재배정 시 자가 회복 불가 웨징. 처방: 재배정 시 HeartbeatAt 초기화(소비자 열거 확인 후)+RED 먼저, 재심사 F4 델타 한정. F1 해소 확인(라이브 시험 2종 PASS·근원 확인). 대리자 4차 수리 재위임.
- 16:2xZ — 4차 수리 커밋 `31be384b6`(재배정 시 HeartbeatAt 클리어 — 소비자 열거 선행 완료, RED 자기참조 웨징 재현→GREEN cli 8종+homestate 이웃 10종). 레인 독립 검증 PASS 4.13s. 범위 판단: F4 처방 파일이 internal/homestate 소속으로 원래 internal/cli 한정 브리프를 카드 소유자(레인)가 이 건에 한해 명시 완화.
- 16:2xZ — **재심사(round 3): PASS-WITH-DEBT**(audited_sha 31be384b6, round 1/1b/2/2b/3/3b 기록): F4 해소, F1·F2 해소 유지, F3 불요, codex 필수 게이트 pass(receipts rcpt-5ad639d77a87b1d9619e574a). 부채 5(비차단): docs 커밋 대기·stage-ownership 리더 후속·커버리지/-race 미측정(부하 규칙)·게이트 도구 빌드 시차·pid 재사용 한계. 오케스트레이터 audit-plan 확인 실시.
- 최종 커밋 목록(WT-10-17-class): `424605284`(D1 생존 보유자 슬롯 유지) → `c741e22ce`(보고서 1차) → `34212a4be`(P2 세대 판별) → `4576923e7`(card-review 1-2 기록) → `7b2e9cd2c`(F1 수거 assigned 점유) → `31be384b6`(F4 재배정 하트비트 클리어) → `d208d9e37`(sync-audit 판정 기록·최종 체인).
- 16:3x-16:5xZ(10-18 01:5x KST) — **T13 증거 매듭**(이 카드 주제의 실측 사례): 카드의 기록 증거(EvidenceSHA=4576923e7, T11 진입 시점)가 감사 판정의 audited_sha(31be384b6, 수리 후)와 불일치 → T12/T13 모두 접두사 검사에서 차단. 기계 해명: T11 증거는 `<HEAD>:<HEAD에 추적된 파일>`(verifyAuditEntry), T12/T13 은 파일명 반복번호 최신 판정 파일의 audited_sha 가 기록 증거의 접두사여야 함(readAuditVerdict/newestVerdictFile). 설계된 수리 루프: FAIL 판정 → T12(sync 복귀) → 수리 → T11(새 증거) → 재심사 → T13. 해법(전부 정당): 감사자가 이미 수행한 round-1 감사(4576923e7, FAIL)를 반복 파일로 재발행 → T12 → T11(증거 d208d9e37:sync-audit-cold.md, HEAD 추적) → 감사자가 docs-only 델타(31be384b6..d208d9e37) 확인 후 최신 판정(audited d208d9e37, PASS-WITH-DEBT) 발행 → T13. 임대 3차 만료·복구(16:37:56 만료 → refused-not-holder 형태 수거 → 재임대 → v15 sync-audit 재개)도 실측 추가.

## 5. 산출물

- 수리 커밋: `424605284` on `WT-10-17-class` (fix(factory): hold the serial slot behind an expired lease while its holder lives) — RED 재현→GREEN(범위 시험 4종 통과)→커밋. 대리자 보고서 원문은 세션 기록, 요지 본 파일 §1·보고서에 반영.
- 최종 보고서: `.moai/reports/t1828/wait-boundary-report.md` — 5단 구조(Claim/Evidence/Baseline/Gaps/Residual-risk) + 경계별 원인표 + 가설/재현 분리 + 수정 파일 + 남은 결정·다음 담당자.
- card-review: `.moai/reports/t1828/card-review.md` (진행 예정)
- 레인 독립 검증(Trust-but-verify): TestFactorySerialSlotHeldRequiresADriver PASS 0.00s(ok 1.012s), TestFactoryNextReNominationAfterOwnExpiryStillRefused PASS 2.10s(ok 3.162s), gofmt 3파일 0행 — 12:47Z.
- 정지 지점: merge-ready (factory complete 호출 금지, main 병합·push 금지 — 리더 지시 10-17)
- 20:4xZ(10-18 05:4x KST) — **카드 종결**: T13 증거 매듭 해소(iter2 FAIL@4576923e7 재발행 → T12 v16 → T11 v17 증거 d208d9e37 재기록 → iter3 PASS-WITH-DEBT@d208d9e37 → **T13 merge-ready v18 착지**). 최종 docs 커밋 `564f57022`. 리더 완료 보고 발송(msg 8af8223e), 재점검 크론 삭제.
- **세션 대기 기록**: 이유=다음 카드 배정 대기(리더 무중단 규칙: 배정분만 next --card), 대상=리더, 재점검점=리더 메시지 도착(세션 웨이크). t1828 merge-ready 정지·워크트리 보존·complete 미호출.

## 재임대 세션 상태 (10-22, lane-5 / run tn8rpj)

- 카드가 신규 run(tn8rpj) 큐에서 재배차되어 `factory next --card t1828` 임대(스테이지 표기 `-` —
  구 run tn1p4e 기록의 merge-ready v18과 별개 행으로 보임). **재수행 불요 판정**: 선행 레인의
  작업 체인 전부 본 브랜치 존재 확인(424605284·34212a4be·7b2e9cd2c·31be384b6·564f57022·
  d208d9e37 전부 HEAD 조상, merge-base 실측) + 판정 파일 4종(sync-audit-cold/iter2/iter3 +
  wait-boundary-report) 디스크 보존.
- 착지 준비로 main 흡수 1회(28b727065b → 머지 커밋 264ec6296a, 카드 트레일러 포함, 충돌 없음).
- **대기 기록**: 사유 = 완료 카드(merge-ready 실질)의 착지 처분, 대상 = 리더 tn8rpj, 선택지 =
  (가) 레인이 착지 수행(준비성 삼중+re-measure+push+PR — t1913 절차 준용), (나) 리더 직접 처분,
  (다) 흡수 후 트리에 대한 신규 감사 1회 후 착지. 재점검점 = 리더 응답.

## 기계 위기 국면 기록 (10-22 10:2x KST, lane-5)

- **부하 크리시스**: load 444~489(역대 과부하 영역) — go build/test 전반 600s+ 지연, MCP 서버
  2517s 무응답 타임아웃, 메시지 경계 degraded(inbox context deadline exceeded). 레인의 상태 확인
  명령조차 백그라운드 전환. 조치: **모든 도구 호출 일시 중단**(서버 워커 점유 방지), 부하 배수 후 재개.
- **t1828 수렴 상태(무결, 커밋됨)**: `36978859a1 fix(t1828): reconcile card repairs with evolved
  main signatures` — factory_card.go(+32/-7, 호출부 현 시그니처 정렬) + liveness 테스트(+49).
  대리자 사전 보고: 빌드·gofmt 정상화 후 커밋. **레인 독립 검증(빌드+스코프 시험)은 부하 배수까지
  보류** — 미실측이며 이는 Gaps로 유지.
- **잔여 단계(부하 복구 후 순서)**: ① 레인 독립 검증(빌드+TestFactorySerialSlot 스코프) ② 흡수
  HEAD 재감사(PASS 목표 — 리더 병합 조건) ③ push + birth-draft PR(--undo 이중 적용, complete 호출
  금지 유지) ④ 리더 보고.
- **t1913 임대 상태**: 만료(02:12:25Z) — 복구 시도가 MCP 타임아웃으로 미완료. PR #1867 draft 보존
  은 원격에 안전(병합 불가 상태). 복구 재개 시 8859: stage 만료 거부(회수 확인) → next 재임대 →
  renew. 카드 기록은 #1867 exact 링크 보존 확인됨(10-22 재임대 시 실측).
- **재개 절차**: 부하 < ~20 확인 후 — ① 이 파일의 잔여 단계 순서로 t1828 진행 ② t1913 8859 복구
  ③ 크론·갱신 루프 재무장(현재 크론 983d9ecb 생존, 루프 만료).

## 착지 준비 완료 (10-22, lane-5)

- **수렴 완료**: 36978859a1(호출부 현 시그니처 + holderAlive 생존 경로 배선, RED-first 배선 시험) →
  5c224aa972(픽스처 4건을 #1859 게이트 헬퍼에 정렬 — 통제실험으로 생산 코드 무관 입증). 카드 수리
  4건 전부 생존 판정(대리자 §E + 감사자 재검증).
- **독립 검증**: 레인 백그라운드 관찰 build exit 0 + TestFactorySerialSlot ok 7.0s(커밋 콘텐츠 동일
  바이트) + 감사자 스코프 회귀 cli ok 263.6s·homestate ok 120.0s + 빌드 2종·vet·gofmt 클린.
- **재감사 PASS**: sync-audit-reconciled.md — rcpt-1d1ccbffd16e5534f8ede8a2, codex required pass,
  장부 체크 `convergence_check.ok: true` exit 0(오케스트레이터 digest 기록 완료). 감사 체인 합계:
  cold FAIL → iter2 FAIL → iter3 PASS-WITH-DEBT(pre-absorb) → absorb FAIL(빌드 파손) →
  **reconciled PASS**.
- **pr-open 달성**: push(origin, 신규 브랜치) → **Draft PR #1893**
  (https://github.com/modu-ai/moai-adk/pull/1893 — 제목에 카드 id, birth-draft, complete 호출 없음 —
  리더 병합 보류 지시 준수). 병합 조건 충족: 흡수 HEAD 독립 감사 PASS(rcpt-1d1ccbf) — 필수 CI는
  리더 라운드에서 관측.
- **리더 후보 과제(보고 동봉)**: 4개 픽스처가 main 자체에서도 실패(통제실험 입증) — main 측 후속
  큐잉 대상. pid-reuse 부채 불변.
- **CI 통제실험 판정 (10-23, dc6527c44e)**: PR #1893 실패 8건 전부 origin/main(bf0a35f10e) 재현 —
  **브랜치 도입 실패 0건**, DATA RACE/panic/timeout 0건, graph-freshness는 공유 codemaps 부채로
  리더 공통 수리 대상(수리 제외). main 고유 결함 5종 리더 큐잉(AGENTS.md 천장 초과·카탈로그 해시
  drift·ALB 장부 행·ubuntu Bubblewrap·TodoSelection 조항).
- **merged-pr (10-23)**: 리더가 #1869 공유 CI 수리(a1762cabe1)를 main에 병합 후 본 PR 병합 —
  **merge commit 734f11a686** (리더 실측 보고). 병합 게이트: 독립 감사 PASS(rcpt-1d1ccbf) +
  리더 라운드. 이 커밋으로 카드 생애주기 종료(merged-pr).

## CI 실패 통제실험 — 브랜치 도입분 0건 (10-23, 재임대 라운드)

PR #1893 실패 체크 4종(runs 39786048290/39786047765)을 CI 로그 → 실패 테스트 8건 확정 →
main(bf0a35f10e) 임시 읽기전용 worktree 통제실험. **전부 main 고유 결함 — #1893 diff 기인 0건,
카드 브랜치 수리 없음**(리더 범위 축소 지시: graph-freshness 제외·main 재현분 제외).

| 실패 테스트 | main 재현 | 증거 |
|---|---|---|
| TestTodoSelectionPredicatesPositivelyEnumerateStates (internal/cli) | YES | main 4건 위반(브랜치는 5c224aa972로 2건 수리, 잔여 2건은 factory_card.go — main과 동일 바이트) |
| TestCodexContractByteCeiling (internal/config) | YES | main AGENTS.md 24807B > 천장 24576B(main이 13줄 추가, 브랜치 미편집 — 병합이 main판 채택) |
| TestCodexNestedTemplateDiscoveryBudget (internal/config) | YES | 동일 원인(24807B, +231 초과) |
| TestBubblewrap_ExecHello (internal/sandbox) | YES(정적 통제) | internal/sandbox 브랜치-무차이 + CI(ubuntu) 환경 동일 → main CI에서도 동일 실패 예측(로컬 darwin은 skip으로 판별 불가 — Gaps) |
| TestCatalogHashParity (internal/spec) | YES | 동일 해시쌍(734677c3 vs 98079f07) 양측 재현 — moai-lane-watchdog 스킬 트리 브랜치-무차이 |
| TestBindingLedgerIntegrity (internal/template) | YES | ALB-0139/0148 after-text 소실, 양측 동일 — .claude/rules/ 브랜치-무차이 |
| TestCatalogHashCoversSkillSubfiles (internal/template) | YES | 동일 해시쌍 |
| TestManifestHashFormat (internal/template) | YES | 동일 해시쌍 |
| graph-freshness | — | 리더 지시로 제외(공유 codemaps 부채, 24개 PR 동일 집합 — 리더 공통 수리 대기) |

- DATA RACE / panic / test-timeout 서명: 3개 로그 전부 0건(고측: grep 실측).
- main 고유 결함 5종은 리더 큐잉 대상: ① REQ-THS-012 잔여 위반(factory_card.go) ② AGENTS.md
  Codex 천장 초과 ③ moai-lane-watchdog catalog 해시 drift(gen-catalog-hashes --all) ④
  ALB-0139/0148 장부 행(moai-mcp-tools 이관 잔여) ⑤ CI ubuntu bubblewrap 실행 실패.
- 임시 main worktree(/tmp/t1828-main)는 통제실험 후 제거 완료. HEAD 불변 66371bebd7, 수리 커밋 없음.
