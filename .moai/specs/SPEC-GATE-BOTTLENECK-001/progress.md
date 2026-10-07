# progress.md — SPEC-GATE-BOTTLENECK-001 (card t1575)

## §E.1 Plan-phase Audit-Ready Signal
plan_status: audit-ready (self — lane-authored SPEC, card t1575의 4종 계약 그대로)
plan_complete_at: 2026-10-07

## §F Phase 4 Mode Selection
Input: tier M, scope 3 files (gate + cache + tests), 1 domain (cli), Go.
Decision: direct (orchestrator/lane-direct, no Agent spawn) — 단일 파일 중심 수리형 구현으로 serial 대용 direct가 정합.

## §E.2 Run-phase Evidence

- M1 (cache) 착지: `internal/cli/codex_review_cache.go`(신설) + `codex_review_gate.go` (4a) 조회·(4b) 기록 삽입.
  - 재판정: `go test ./internal/cli/ -run 'TestReviewGate' -count=1` → ok (7.1s)
  - 신규 2종: TestReviewGate_CacheHitReusesFailVerdict(캐시 히트 → BLOCK·RPC 0회·skip 카운트)·TestReviewGate_StaleKeyRunsLiveReview(키 이동 → 라이브 리뷰 전환 — 실패 픽스처로 판별력 강화)
  - fixture 교훈: receipt 기록이 porcelain을 더럽히면 키가 이동 — fixture가 실repo의 `.moai/state/` ignore를 반영해야 함(실측).
- M3(스코프 분리): 계약 문서화 완료(SPEC § Scope split — 카드 diff 축은 t1383 소관).
- M4(재현 조건부): 캐시 히트 경로 = 전체 재현 생략(REQ-GBN-003, 구현됨). 라이브 리뷰의 0건 조건부는 review request 파라미터 확장이 필요 — M2와 함께 잔여.

## §E.3 Run-phase Audit-Ready Signal
run_status: partial — M1 착지, M2(지연 블록)·라이브 리뷰 0건 조건부 잔여.

## 체크포인트 (다음 세션 재개 지점)

- M2 지연 블록: Stop에서 수신 영수증 부재 시 백그라운드 리뷰 기동(`moai verify codex-review` 재사용) + ALLOW, 판정은 다음 턴 진입 훅(UserPromptSubmit/PreToolUse 신설 배선 — settings.json.tmpl + 훅 매니페스트 M4)에서 receipt 조회해 집행. async:true 불채택(블록 능력 상실).
- 라이브 리뷰 0건 조건부: reviewRequestParams에 재현 지시자 추가(codex 측 지원 확인 선행).
- 검증 남은 것: 턴당 벽시간 전후 비교 실측, no-edit 자체허용 회귀 없음(기존 TestReviewGate_NoEditTurnAllows가 계속 green으로 보호 중), fail-open 불변(신규 테스트 2종이 담보).

## 턴종료 게이트 발견 처분 (r9 — 본 카드 diff 발견 0건)

턴종료 게이트가 merge-ref 전체 diff를 보며 보고한 14건(patch-id P1=t1561 발행·after 힌트 재계산·bundle 3건·picked 스킵·todo_issuance 6건·todo.go 2건·backlog_store/relation)은 **전부 t1542 card-review 원장과 리더 원장(t1533 라인·t1561·t1562·t1559)에 기재된 기존 행의 재관측**이다. 본 M1 diff(codex_review_cache.go·gate 조회/기록·테스트)에서의 발견은 0건. 본 카드는 이들 수리의 소관이 아니며 원장 행이 소유한다.

## 턴종료 게이트 발견 처분 (r10 — 본 카드 diff 2건 수리, 타 소관 11건 재관측)

- **본 소관 2건 수리(커밋 대상)**: ① fail receipt에 ExitCode=1 기록(기존 produceCodexReviewReceipt와 동일 매핑 — pass형 0이 다른 소비자에게 성공 증거로 읽히는 결함) ② 캐시 차단 메시지가 "무엇을 고칠지"를 잃는 문제 — fail 시 요약·발견을 `.moai/state/verify/codex-review/<head>-<digest>.md`(트리키 동일·런타임 관리 영역)에 보존하고 캐시 차단 이유에 첨부. 판별 테스트 TestReviewGate_LiveFailPreservesDetailForCachedBlock 추가. `go test ./internal/cli/ -run 'TestReviewGate' -count=1` → ok (24.0s).
- **타 소관 11건 재관측**: protected_zone_shell(1·t1500/1510 계열)·todo_issuance(4·t1559)·factory_bundle(3·t1561/62)·todo.go(1·t1554)·backlog_relation/store(2·t1542 card-review 원장). 전부 r9 원장 행의 재관측 — 본 카드 수리 소관 아님.

## 턴종료 게이트 발견 처분 (r11 — 본 카드 diff 발견 0건 연속)

r10 수리 판정 통과(리뷰어 명시: 캐시 회귀 테스트 3개 통과). 13건 전부 기존 원장 행의 재관측 — factory_card 3·factory_bundle 3·todo_issuance 4·todo.go 2·backlog_store/relation 2의 계열 분포는 r9/r10과 동일. 소관 카드(t1542 원장·t1561·t1562·t1559·t1454·t1554)가 소유하며 본 카드 수리 소관 아님.

## CI 적색 2건 수리 (리더 지시 2026-10-07 — M1 PR #1793)

- **① TestStopChainEffectParityGolden /codex_review_gate 2 leg** — REQ-GBN-001이 바꾼 공유 저장소 계약의 갱신: Claude 게이트가 라이브 판정을 기록하므로(의도된 M1 동작) "Claude 통과 뒤 영수증 부재 → unmeasured" 골든이 더 이상 성립하지 않는다. 두 leg(codex installed no receipt·stale receipt HEAD moved)를 gate_failed 기대로 갱신하고, 영수증 기록 자체를 전제 단언(receiptForCurrentState 헬퍼 — 기록이 말없이 빠지면 leg가 공허해지는 것을 막음)으로 고정. 판정 일치(Deny=Deny)는 유지 — parity 위반 아닌 관측 클래스 갱신. unmeasured 행위 커버리지는 cap 서브테스트가 유지.
- **② TestCheckProtectedZonePosixBackslashConvertedAbsoluteness** — origin/main(5e5ff8e31, t1570 #1797) 흡수 후 green. 코드 변경 0건.
- 흡수: origin/main → 병합 HEAD 64a13b837. 재판정: `go test ./internal/cli/ -run 'TestStopChainEffectParityGolden' -count=1` → ok (40.9s) · `-run 'TestReviewGate|TestCheckProtectedZone'` → cli ok (7.0s)·hook ok (5.0s).

## 최종 판정 (2026-10-07 — PR #1793 @ 92897ca84 전 체크 초록)

- **CI**: 29 pass·0 fail·0 pending — Test (ubuntu-latest) 12m37s·Race Test 1/2 (17m17s/18m40s)·Build×5·Lint·Constitution·Integration×3·spec-lint·spec-status-sync·graph-freshness-conflict-guard·CodeRabbit 포함. CodeRabbit 판독은 리더 몫(AH §4 병합 판정).
- **CI 수리 이력**: 골든 2 leg 계약 갱신(7986851b9) → LiveAxis 오탐 제거(1a3bf99a7) → inconclusive exit 2(92897ca84, 쌍둥이 포함).
- **게이트 라운드 원장 (r12~r17)**: 본 카드 diff 발견 0건 6라운드 연속. 전부 기존 원장 계열의 재관측/신규 인스턴스 — integration 병합 창 재검증 4건·factory_card/bundle 선행의존성 5건·todo_issuance 4건·backlog relation/store 2건·todo.go 2건·landing_predicate 공백 1건·audit_receipt_guard 재시작 표식 1건·todo_auto_lane 1건·integration remeasure 2건. 소관: t1538/t1542 원장·t1561·t1562·t1559·t1454·t1554·t1509.
- **잔여(M2/M3, 체크포인트)**: 지연 블록(Stop 백그라운드 기동 → 다음 턴 진입 훅 집행)·라이브 리뷰 0건 조건부 재현(reviewRequestParams 확장) — 신선 세션 continuation 권장.
