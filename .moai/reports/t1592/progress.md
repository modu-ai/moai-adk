# t1592 — windows 전용 bind-budget 적색 수리

lane-13 (generation 3) · run tmhxo0 · 트리 `.claude/worktrees/t1592` · 브랜치 `WT-win-bind-budget`
베이스: a2a184ad3 ([t1575] Review gate M2, #1802)

## 카드

[windows 전용 적색·리더 발행 10-08] internal/hook TestRebindPathStaysInsideBindBudget/production_bind_budget —
Multi-OS Test (windows-latest)에서만 FAIL(ubuntu 동일 패키지 초록). gh run 37641344677 job 112860999078.

## 수령 경위 (기록용)

1. 리더 배차 수령 → `factory next --card t1592` 거부(serial-slot).
2. 원인 판독: tmhxo0 cards 14행이 2026-10-08T01:22:11~27Z 일괄 rewrite로 F1 enum 밖 legacy
   `completed` 리터럴이 됨 → `factorySerialSlotFree` positive enumeration이 해제 못 함
   (factory_card.go:245) + `LeaseExpired` false(card_record.go:146) → 14행 전부 슬롯 보유자 판정.
   마지막 성공 임대 events seq 280 (2026-10-07T16:49Z). 리더 보고 후 리더가 14행을 `done`으로
   재기록 — 기계 재관측 완료(14행 done, t1588 임대 관통 seq 282-284). 웨지는 리더가 청산.
3. 재임대 재시도 → t1588(lane-10)의 살아 있는 정당 임대(만료 01:53Z)가 슬롯 보유. 웨지가 아닌
   설계 동작(REQ-TCD-008)이라 무임대 집행(t1498 선례는 보유자 부재 웨지 전용)을 적용하지 않고
   **명시적 대기 + 대기 중 읽기 전용 조사**로 판정. 재점검 크론 f9b6639b (:07/:27/:47).
   슬롯 해제 시 임대 후 구현 착수. (리더 수령 확인 10-08)

## RED 증거 (관측분)

- CI (windows-latest, gh run 37641344677 job 112860999078):
  ```
  --- FAIL: TestRebindPathStaysInsideBindBudget/production_bind_budget (5.01s)
      factory_rebind_test.go:596: rebind path elapsed 3.079591s against the 2s bind budget
      factory_rebind_test.go:598: the rebind did not complete inside the production bind budget:
        "factory messaging degraded: initialize factory message broker: context deadline exceeded"
  ```
- 로컬 darwin 기선 (2026-10-08, 이 트리 a2a184ad3):
  `go test ./internal/hook/ -run 'TestRebindPathStaysInsideBindBudget|TestRebindBrokerInitializationHonorsCallerBudget' -v -count=1`
  → production_bind_budget **elapsed 256.514916ms** PASS (2s 대비 8배 여유), 나머지 서브테스트 전부 PASS.
- windows/darwin 비: 같은 신규-초기화 경로가 windows에서 ~12배 — **2s 예산은 windows 축에서 한 번도
  보정된 적 없음**(SPEC-FACTORY-STALE-RUN-HEAL-001 acceptance.md:15의 보정 근거 "약 75ms"는 darwin).

## 메커니즘 (원인 확정)

- 실패 문면이 `context deadline exceeded`(SQLITE_BUSY 아님 — 대조: 잠금 경로는
  "database is locked (5) (SQLITE_BUSY)"로 나옴, 로컬 locked-init 테스트가 그 문면을 핀함).
- 에러 래퍼 위치: internal/factorymsg/store.go:316-318 — `db.ExecContext(ctx, schema)`.
  신규 broker.db의 schema DDL(peers/messages/dead_letters + handoff 2종 + index, 6문 이상)이
  2s 예산 안에 못 끝남.
- 가속 요인 3종:
  1. modernc.org/sqlite v1.60.1 — 다중 문 ExecContext를 문 단위 실행, 각 DDL 자동커밋 → 커밋당 flush.
  2. DSN 프라그마에 `synchronous` 부재 → 기본 FULL: WAL에서도 커밋마다 WAL sync.
     두 열린 경로 모두(openWithContext store.go:293-304, OpenExistingWithDeadline store.go:226-228).
  3. windows-latest 콜드 FS + Defender 실시간 검사 — 파일 생성·플러시당 수백 ms.
- 카드 본문의 초기 가설(경로 구분자·권한 축)은 기각 — 실패 본체는 시간 예산.
  brokerDSN은 이미 filepath.ToSlash로 드라이브 경로를 처리(store.go:163-171).

## 수리 설계 (결정 기록)

- **채택: `synchronous(NORMAL)` 프라그마를 두 열린 경로의 DSN에 추가.**
  - 근거: 브로커는 조정 상태(메시지 만료·턴마다 훅 재시도·dead_letters best-effort)라
    전원 손실 시 마지막 커밋 유실을 허용 — WAL+NORMAL은 SQLite 권장 쌍.
    커밋당 fsync 제거로 신규 초기화 DDL과 이후 쓰기 모두 절감.
  - 대안 기각: (a) 예산 상향 — AC-SRH-014의 2s는 훅 5s 타임아웃 안쪽 여유분까지 설계된 값이라
    상향은 톱니를 무디게 함, windows 분산(런처별 수 초 스파이크)을 수치로 못 흡수.
    (b) schema 단일 트랜잭션화 — ctx 취소 시 드라이버 커넥션에 explicit tx 잔류 위험
    (MaxOpenConns(1)이라 오염 커넥션 재사용), NORMAL 적용 시 이득 소멸.
    (c) windows skip — AC의 경과시간 톱니 포기, 기각.
- 핀: pragma 조립을 두 경로가 공유하는 순수 헬퍼로 추출, 유닛 테스트로
  synchronous(NORMAL) 상용 존재를 기계 고정(무단 제거 회귀 방지).

## 계획

1. [진행중] 임대 확보(슬롯 해제 대기, 크론 재시도) — 확보 후 구현 착수
2. store.go 두 DSN pragma 세트에 synchronous(NORMAL) + 헬퍼 추출 + 핀 테스트
3. 검증: internal/factorymsg + internal/hook 영향계열 재실행 (트리 전체 스위트 금지 — t1542 교훈)
4. windows GREEN 판정은 PR CI(Multi-OS windows 잡)가 주체 — diff 경로가 잡 트리거에 걸림
5. card-review → push·PR (github-flow, main) — 기록 정산(assign/stage/complete)은 임대 경로로

## Gaps (미관측)

- windows 로컬 실행 불가 — windows GREEN은 PR CI 판정으로 대체. 로컬 검증은 darwin 한정.
- synchronous(NORMAL)의 windows 실효치(개선 폭)는 CI 경과 로그로만 관측 가능. 만약 여전히 2s 초과 시
  후속 결정: (i) 초기화 DDL 단일 트랜잭션 재검토 (ii) 예산 재보정(운영자 합의 축).

## Residual-risk

- windows 러너 편차(Defender·콜드 캐시)가 크면 NORMAL 만으로 2s 안에 못 들 수 있음 — 이 경우
  위 후속 결정 축. AC-SRH-014의 예산값 자체는 이 카드에서 바꾸지 않음.

## Run-phase 증거 (2026-10-08, 커밋 30b9befdb 기준)

**Claim** — windows-latest의 2s bind-budget 실패는 신규 broker.db schema DDL의 문당 자동커밋이
`synchronous` 프라그마 부재로 커밋마다 FULL fsync를 지불하기 때문이다. 두 열린 경로 DSN에
`synchronous(NORMAL)`을 넣으면 per-commit flush가 사라져 예산 안에 들어온다(WAL+NORMAL은
crash-safe — 권한 상실 축 아님).

**Evidence** (본 트리 30b9befdb에서 관측):
- `go build ./internal/factorymsg/ && go vet ./internal/factorymsg/` → BUILD_OK
- `go test ./internal/factorymsg/ -run 'TestOpenPathsSetSynchronousNormal' -v -count=1`
  → `--- PASS` (init/existing 양 경로 실연결 `PRAGMA synchronous` = 1 NORMAL)
- `go test ./internal/factorymsg/ -count=1 -timeout 300s` → `ok ... 165.109s` (패키지 전체)
- `go test ./internal/hook/ -run 'Rebind|BindBudget|BrokerInit|FactoryMessage|FactoryHook'
  -count=1 -timeout 300s` → `ok ... 122.357s`
- 변경: internal/factorymsg/store.go(2 DSN pragma) + store_synchronous_test.go(핀, 2파일 +56행)

**Baseline-attribution** — RED: gh run 37641344677 job 112860999078(windows, 3.08s vs 2s);
darwin 기선 256.514916ms(수리 전 동일 테스트, 본 문서 RED 증거 절). 수리 후 관측치는 모두
본 트리 30b9befdb에서 본 세션 실행.

**Gaps** — windows 네이티브 실행 없음(PR CI가 windows 판정 주체 — diff 경로가 Multi-OS
잡 트리거에 걸림). synchronous(NORMAL)의 windows 실효 개선폭은 CI 경과 로그로만 관측.

**Residual-risk** — windows 러너 분산이 크면 NORMAL만으로 2s 미달일 수 있음 → 후속 축:
(i) 초기화 DDL 단일 트랜잭션 재검토 (ii) 예산 재보정(운영자 합의). AC-SRH-014의 2s 값은
본 카드에서 불변.

## 대기·결정 기록 (2026-10-08T02:14Z 기준)

- 거부 3차 원인 변경: 14행 웨지(해소) → t1588 live lease(만료 해소) → **assigned 형제 행 2개**.
  t1592는 기록 행 없는 신규 카드라 지명 임대가 `ignoreAssigned=false`(factory_card.go:1216)를
  통과 — lease 만료로 `assigned` 회수된 t1588(lane-10, 01:53Z 만료)·t1598(lane-5, 02:08Z 만료)이
  소유자 있음+해제 집합 밖으로 슬롯 보유. assigned엔 만료 안전망이 없어(LeaseExpired는
  lease-holding 상태에만 참) 두 레인이 재임대 안 하면 신규 카드 임대는 무기한 막힘.
- 결정 보드: `moai decision read --scope card:t1592` → board=absent (사다리 ② 통과).
- 사다리 종착: ① 디스크 증거(기계 판독 완료) → ⑤ 리드 채팅 — 필요한 것은 레인이 할 수 없는
  기록 행위(큐 승격/assign)뿐.

```text
wait record: id=t1592-w1 reason=serial-slot held by assigned siblings t1588(lane-10)·t1598(lane-5) — leases lapsed 01:53Z/02:08Z, assigned rows carry no expiry net against a new-card nomination (factory_card.go:1216 ignoreAssigned=false) target=leader surgical unblock: moai factory assign t1592 --lane lane-13 (t1542 own-assigned nominated arm passes ignoreAssigned=true) or clear/re-lease the stale rows recheck=cron f9b6639b :07/:27/:47
decision record: decided_by=lane-13(claude·lane) evidence_refs=.claude/worktrees/t1592/.moai/reports/t1592/progress.md + factory.db holders query (t1588·t1598 assigned, leases lapsed 01:53Z/02:08Z) + internal/cli/factory_card.go:1209-1217 ladder_path=①→⑤
```

- 02:3xZ 갱신: 리더가 t1592 assign 완료(lane-13, v2) — own-assigned 팔 준비됨. 잔여 블로커는
  t1588의 살아 있는 sync-audit 임대(02:51:58Z 만료, lane-10 진행 중·events seq 305 증거 전진) —
  own-assigned 팔은 assigned 형제만 건너뛰므로 sync-audit 행은 여전 계수. 해제 시점 =
  t1588 merge-ready 전이 또는 02:51:58Z 중 빠른 쪽. 크론 재시도 유지.

- 04:1xZ 리더 판정 접수: 재분류 동사 설치 빌드 미존재(t1480 미착지)·보유 레인(t1598 plan-audit
  게이트 재시도 중·t1587 sync) 활동 중이라 갱신 보류 레버 미사용 — **감시 v2 저격 유지** 확정.
  굶음 패턴의 구조적 해소는 t1600 D4(작업 무임대 병렬·직렬화는 리더 병합 큐) 착지 후.

## 이벤트 로그

- 2026-10-08 ~11:00KST Stop 훅 codex 게이트가 본 레인 턴 종료에 발화(Verdict: fail, P2 3건) —
  지적 파일(internal/userassets/lock_guard_windows.go:23·install.go:211,
  internal/cli/update/reconcile_classify.go:220)은 **카드 diff 밖·본 트리 베이스 대비 완전 클린**
  (`git status --short` 0행 관측, 진행 기록은 gitignore). 세션 이동 후에도
  CLAUDE_PROJECT_DIR가 primary로 스폰-동결(MOAI_PROJECT_DIR 재도장 false)되어 primary의
  타 세션 WIP를 검사한 것으로 판독 — t1383 클래스 잔여 변형. P2 3건은 codex 리뷰 보고 그대로
  리더 전달(본 레인이 검증한 것 아님 — 귀속·처분은 리더/발행 축). 레인은 카드 범위 밖 수리 안 함.
- 2026-10-08 ~02:2xZ 동일 게이트 오발화 재발(2회차, P2 4건 — update_dryrun_preview.go:49·
  codex_review_gate.go:157·userassets/install.go:211·atomicfile/section.go:283). 카드 트리
  여전히 베이스 대비 클린. primary 타 세션 WIP가 진화 중으로 매 레인 턴 종료마다 게이트가
  foreign WIP를 검사·블록 — 리더 재통지 완료, 레인 처분 불변(범위 밖 미수리).
- 02:5xZ **귀속 정정**: 3회차 게이트 판정문이 스스로 "HEAD a2a184ad3·깨끗한 작업 트리·동일
  diff 해시 재확인"을 밝혀 — 게이트는 primary 타 세션 WIP가 아니라 **베이스 트리 자체의
  기존 결함**(userassets·atomicfile 축, 재현 프로브 /tmp/t1592-review-probes/)을 검사하고
  있었다. 1~2회차 보고의 'primary WIP' 귀속은 기각. 리더 정정 통지. 슬롯 경과: t1588
  merge-ready 전이(02:52:46Z) 직후 lane-5가 t1598을 plan으로 잠금(02:52:51Z, 초 단위 경합) —
  20분 크론으로는 열림 창을 놓치므로 20초 간 감시 루프(/tmp/t1592-slotwatch.sh, 읽기 전용,
  상한 ~2h)를 배경 실행해 열리는 즉시 임대 시도로 전환.
- 03:3xZ 감시 루프 v1은 개방을 잡았으나 통지 왕복 지연 사이 lane-5가 t1598을 재잠금(03:28:47Z,
  events 313→314) — 알림 경로의 수십 초 갭으로는 경합 승률 0. v2(/tmp/t1592-slotwatch2.sh,
  배경 bx4huillm): 3초 간 폴링, 개방 감지와 임대 시도를 같은 반복에서 실행(부모 체크아웃
  서브셸 cd — 동사가 부모 체크아웃 강제 실측, 워크트리 cwd 거부). 성공 임대 시 즉시 종료
  신호 → 10분 임대 창 안에 leased→run 전이로 갱신하며 구현 착수. 실패(경합 패배) 시 루프
  지속. 상한 ~6h.
- 02:5xZ 게이트 4회차: 지적 경로가 본 워크트리 파일로 표기 + "기준 커밋 구현으로는 통과" 문구.
  가설 검증: `git rev-list --count a2a184ad3..origin/main` = **0** (베이스가 현재 — 역방향
  델타 가설 기각). 결론: 게이트가 빈 diff 트리를 전수 리뷰하는 것이 본체(t1383 수리축).
  좌표 실재 확인 — section.go:283 `os.ReadFile(path)`·codex_readiness.go:220
  `os.UserHomeDir()` 실측 일치. 결함은 실재하나 카드 범위 밖(리더 원장 보유). 리더가
  재전달 불요 통지에 따라 추가 전달 없음.
