# SPEC-TODO-QUEUE-HOME-CANON-001 — plan-audit 보고서 (카드 t658)

- Audit surface: plan-auditor (독립 감사 에이전트) — 3회 반복 (Tier M 2회 + 조정자 승인 상한 연장 1회)
- 최종 판정 기준 트리: WT-home-queue-canonical @ e37659c00
- 이 파일은 리드 요청(2026-09-13, 세션 로그는 인용 불가)으로 감사 판정 전문을 영속화한 것

## Iteration 1 (v1.0.0 @ 68246a364) — FAIL 0.95

- 차단 D1: plan.md:69 `[NEEDS CLARIFICATION: none at authoring — …]` — 자기 해소 선언도 괄호 토큰 형식이면 MP-7 게이트 적중(t530 사례와 동일 실패 양상). 나머지 must-pass 전부 PASS, 점수는 Tier M 0.80 상회.
- D2 [optional]: 설문 A16 행 인용 `session_start_kanban.go:219` → 실제 호출 지점 :212
- D3 [optional]: 설문 B9 행 — screens_templ.go:1180을 wal/-shm 처리로 오기재(실제는 렌더 코멘트)
- D4 [optional]: REQ가 내부 심볼명 인용 — 통합 SPEC급에서는 불변식의 대상 자체라 수용 기록
- 트리 검증: 카드 전제 만료 확인(statusline/backlog.go:47-50이 BacklogCountsForRoot 위임 — 직접 JSON 독자 없음), foreman SKILL.md:95 죽은 경로 고정 실측(템플릿·로컬 byte 동일), todo_queue_read.go 단일 읽기 봉인 확인, REQ-003 가드의 도달 집합이 예외 2곳과 정확히 일치함 확인.
- cross-model: 미수령(백엔드 0 참여 — fail-open)

## Iteration 2 (v1.1.0 @ 3d6b6279b) — FAIL 0.95 (delta 결함 2건 신규)

- D1-D3 회귀 없음, MP-7 PASS 전환.
- N1 [blocking-delta]: D3 수정 커밋이 스스로 잘못 고친 범위 — events.go SHM :167-168 / WAL :169-172(실질 본문 :170-172)인데 `:167-169`로 축약 인용
- N2 [blocking-delta]: 1.1.0 HISTORY 행의 "bracket-token form never used in any artifact" — 부모 커밋이 실제로 토큰을 담고 있었으므로 검증 가능한 거짓 기술
- cross-model: codex 참여 — N1·N2를 독립 확인(claude 앵커와 수렴), GLM fail-open

## Iteration 3 (v1.2.0 @ e37659c00) — **PASS 0.95** (조정자 승인 상한 연장, 좁은 delta 재확인)

- N1 해소: B9 행 = `events.go (SHM :167-168; WAL :169-172)` — 코드 정밀 대조 일치, 구 범위 grep 0건
- N2 해소: HISTORY가 실제 제거 내용을 정확히 기술("none 페이로드조차 게이트를 트리거함")
- must-pass 7종 전부 PASS. 잔여 선택 기록 1건(HISTORY 산문 문구 차이 — 추가 반복 유발 금지 규정으로 미수정)
- 5-섹션 판정 근거: Claim(진입 가능)/Evidence(grep·diff 원문)/Baseline(t658 트리 3회 반복별 HEAD 명시)/Gaps(런 단계 행동 미관찰 — 정의상)/Residual(PASS는 문서 품질 인증, 구현 검증은 가드 테스트의 변이 제어에서)

## 최종: **run 진입 PASS** — skip-eligible 3조건 충족 (판정 PASS·0.95≥0.80·해시=e37659c00 시점). 구현 시작 승인은 별도 필수(운영자 승인 2026-09-13 수령).
