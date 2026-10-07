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
