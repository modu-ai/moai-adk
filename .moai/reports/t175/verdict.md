# t175 — 카드 판정 기록 (verdict — 2026-08-22)

worktree t175 / branch `WT-glm-effort-max` (base origin/main @ 1519f2660).
상세 §A 측정은 동봐 `measurements.md` (probe_shim.py 포함).

## plan

SPEC-GLM-EFFORT-MAX-001 (Tier S, 6 REQ/8 AC): 1차 FAIL 0.875(마커 게이팅 +
정정 3건) → 리드 비준 2건(세션 기본 max + REQ-GER-004 슈퍼세션 기록 / 미러
hunk) + D2-D6 반영 → 재감사 **PASS 1.00** (0.875→1.0). 커밋 f49117968(§A
실측) · 6d12df688(SPEC) · 46ddbd838(수정 v0.1.1) · 5ac73c885(감사+§F serial).

## run (리드 판정 PASS — AC-GEM-008 종단 확인은 릴리즈 바이너리 설치 시점 지연)

AC 7/8 PASS + AC-GEM-008 지연(풀 운영 중 전역 바이너리 스왑 리스크 회피 —
단위 증거 TestGLMReasoningEnvVars_SessionMax 확보; 리드가 설치 직후 신규 GLM
세션 env 확인 결과를 §E.2 같은 행에 추가). RED 17실패 verbatim · GREEN
4커밋(a0340aa22→9685b35ab: RED→GREEN→미러 문서+make build→§E) · 커버리지
internal/template 85.7% · 3플랫폼 빌드 0 · lint 신규 0 · must-not-flip
가드 무변경(9파일 diff 확인).

## sync + 종결

경량 close: CHANGELOG [Unreleased] Changed · `in-progress → completed` ·
§E.4 + sync_commit_sha 7803fab90 (백필 91fc1f108) + 인용 정정 0c7df92c0
(슈퍼세션 대상 = REBALANCE-001, TUNE-001 아님).

## 핵심 실측 (§A — AC-MTP-032b 잔여 폐쇄)

- shim: Anthropic `thinking` 파라미터 수용·깊이 제어 가능(P1/P3), 최상위
  z.ai식 `reasoning_effort`는 무시(P2) — 유효 통로는 thinking-budget 매핑
- 제 세션 env=high(현행 하드코딩 일치 — 주입 배선 실증), 세션 응답의
  thinking 블록 = 종단 관측
- llm.yaml glm.effort 블록 stored-only 확정(통로 가설 P2로 반증)

## 잔여

- AC-GEM-008 종단 확인 — 리드, 릴리즈 바이너리 설치(t151 절차 rm+cp) 직후
- glm.go:377-384 "thinking toggle (thinking-off state)" 선재 진부 — 마이크로
  카드 후보(리드 큐 기록됨)
- max-high reasoning 토큰 증분 미정량(spec §5 수용) · REBALANCE-001 정체
  초안 처분 — 리드 배치 시점 별도 질의
