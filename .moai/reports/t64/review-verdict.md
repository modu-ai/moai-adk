# t64 Review Verdict — PASS

- Card: t64 — GLM 1M 컨텍스트 창 표시 실측 (측정 카드, 코드 변경 0 — verified: 커밋 `7aa7d22ea`는 `.moai/reports/t64/` 4파일 추가만 포함)
- Criteria (dispatch-specified): 측정 설계의 타당성 + 증거 재현성
- Reviewer session: release-v311 (2acd4be4), 2026-08-17

## Claim (주장)

배포 바이너리 v3.1.0(d6b80a01c, t65 수정 225a51e24 포함 — ancestry 확인됨)에서
4개 표면이 전부 1,000,000을 보고한다 — 원 카드 주장("200K로 보임")은 재현되지 않는다.

## Evidence (증거) — reviewer-executed re-verification

| 표면 | 원시 증거 재검증 | 관측 |
|---|---|---|
| ① CC 원시 statusline | `grep context_window_size raw/statusline-stdin.jsonl \| sort \| uniq -c` | **28/28 이벤트 전부 `1000000`** — 단일 값만 존재 |
| ② MoAI statusline 스냅샷 | `cat raw/context-usage.json` | `context_window_size: 1000000`, `raw_pct: 0`, `band: large`, session a5bfa9c6 — 리포트 일치 |
| ④ 세션 결과 JSON | t77 raw 재검증 활용 | `modelUsage["glm-5.3"].contextWindow: 1000000` (런 A/B 양쪽) |
| (배경) live env | t77 리포트 §증거2 (`ps eww`) | `CLAUDE_CODE_MAX_CONTEXT_TOKENS=1000000`, `CLAUDE_CODE_AUTO_COMPACT_WINDOW=1000000` |
| 설정 원본 | `cat raw/llm-yaml-used.yaml` | 전 슬롯 glm-5.3 미러 확인 |

표면 ③(TUI `/context` "Auto-compact window: 1m tokens")은 tmux capture 인용 —
원시파일로는 미보존되나 재현 절차가 인라인 문서화됨(하단 Gaps).

## 설계 타당성 판정

- 원 주장의 비재현을 "현재 바이넌리 4표면 1M"로 간접 입증하는 구조가 정직함
  (pre-t65 재현 생략 사유를 Gaps에 명시 — 구 바이너리 빌드 비용 대비 이력이 명확).
- 교정 경로(glmContextWindows)가 원시 페이로드와 "뒤집지 않고 일치하는" 상태로
  관측됐다는 점은 두 독립 경로의 교차확인으로, 설계상 강점.
- 원 카드 전제가 이미 머지된 t65로 해소된 상태임을 확인 — 코드 불변경 판단이 옳음.

## Gaps (미검증)

- 표면 ③ tmux capture 원시파일 미보존(절차는 재현 가능).
- "잔여 15,000,000 tokens" 예산 표시: 원인 미규명(리포트 Gaps 명시) — 단,
  **본 리뷰 세션이 동일 표시를 라이브로 관측**(세션 토큰 카운터 15,000,000에서
  감소 중 — 13,809,436 → 13,777,353)하여 표시의 실재만은 독립 뒷받침.
  후속 카드 후보 지위 유지가 타당.
- glm-5.3 이외 티어·실측 1M 스케일 오토컴팩트 발동 여부 미측정(카드 범위 밖 명시).

## Residual-risk (잔여 위험)

- `CLAUDE_CODE_MAX_CONTEXT_TOKENS`는 클라이언트측 선언 — 서버측 하드 한도와
  어긋나면 표시만 1M일 수 있음(실측 범위 97K+ 토큰 정상 처리 관측).
- CC 버전업에 따른 스키마 변경 가능 — 교정 경로가 이중 안전망(리포트 서술 동의).

**Verdict: PASS** — 카드 목적(t65 잔무: 1M 표시 실측) 완전 충족, 잔여 운영자 몫 없음.

## 독립성 고지

t77과 동일 — 동일 계보 /clear 이후, 커밋된 증거만으로 재유도.
