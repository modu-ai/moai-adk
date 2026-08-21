# t175 §A 측정 기록 — GLM reasoning effort max 상향

> 카드 t175 plan-phase §A ground truth. 모든 주장은 (a) 코드 인용, (b) 세션 환경
> 직접 관측, 또는 (c) 프로브 응답 usage 관측 중 하나로 뒷받침된다.
> 코드 앵커: origin/main @ 1519f2660 (본 워크트리 HEAD; glm_effort_overlay.go는
> 이번 풀 무변경 — 카드 전제 확인).

## 1. 현재 코드 상태 (internal/template/glm_effort_overlay.go)

| 함수 | 현행 | 행 |
|---|---|---|
| `CollapseClaudeEffortToGLM` | low→`low`, **medium/high→`high`**, xhigh/max→`max`, 미인식→`max`(전체성 조항) | :117-129 |
| `SessionGLMReasoningState` | **하드코딩 `high`** (서브에이전트·빈 effort 폴백의 세션 기본값) | :197-199 |
| `SessionGLMReasoningStateForEffort` | 비빈 effort→collapse(effort), 빈→세션 기본 | :210-215 |
| 비용 정책 주석 | "세션-글로벌 값은 세션의 모든 스폰이 지불 — 코드 생산 에이전트만이 아니라. 그래서 바닥은 `high`, 천장 `max` 아님" | :186-192 |
| shim 소비 미확인 주석 | "z.ai의 Anthropic-compat shim이 ANTHROPIC_REASONING_EFFORT를 실제로 소비하는지는 UNVERIFIED" | :194-196 |

지시(운영자 원문): high/medium→max, low→low — 즉 collapse의 medium/high 케이스를
`glmReasoningHigh`에서 `glmReasoningMax`로. low와 전체성 조항(기본 max)은 무변경.

## 2. 세션 환경 직접 관측 (AC-MTP-032b 주입 절반 — 관측)

본 세션(팩토리 lane-9, `moai glm` 계열 GLM 세션)의 env:

```
ANTHROPIC_BASE_URL=https://api.z.ai/api/anthropic
ANTHROPIC_REASONING_EFFORT=high        ← 현행 SessionGLMReasoningState() 값과 정확히 일치
ANTHROPIC_AUTH_TOKEN=SET (값 미출력)
```

→ Branch-B 명시-쓰기 배선이 살아 있고 현재 값 `high`는 코드의 하드코딩과 일치
(주입 메커니즘 실증). 변경 후 신규 세션에서 값이 `max`로 바뀌는지가 run-phase
확인 항목 (본 세션은 이미 뜬 세션이라 env가 고정 — 바이너리 출력 경로 테스트로
대체 가능).

## 3. z.ai shim 프로브 (AC-MTP-032b 소비 절반 — 3회 한정 실측)

프로브: `probe_shim.py` (동봉) — 동일 tiny 프롬프트, 세션 자격증명, usage만 관측:

| 프로브 | 요청 차이 | 응답 | 관측 |
|---|---|---|---|
| P1 | `thinking:{type:enabled,budget_tokens:2048}` | `blocks=['thinking','text']` output=48 | **shim이 Anthropic thinking 파라미터 수용 — 추론 블록 반환이염** |
| P2 | 최상위 `reasoning_effort:"max"` (thinking 없음) | `blocks=['text']` output=3 | **무시됨** — z.ai 네이티브 reasoning_effort는 이 엔드포인트의 최상위 필드로 작동하지 않음(무언 무시, t91 §1 패턴) |
| P3 | `thinking:{...budget_tokens:32768}` | `blocks=['thinking','text']` output=23 | budget 상향에도 thinking 유지 — 와이어 깊이 제어 가능 |

종단 관측: 본 세션의 응답 자체가 thinking 블록을 포함 — env(high) → Claude Code
effort→thinking 매핑 → shim 수용 → 추론 관측의 사슬이 산 채로 돌아가는 것이
본 세션 그 자체.

**판정**: AC-MTP-032b의 미확인 잔여("ANTHROPIC_REASONING_EFFORT가 shim을 통과하는가")
는 다음과 같이 좁혀진다 — (a) shim은 Anthropic 형태 `thinking` 파라미터를 수용하고
깊이(budget)를 따른다(실증), (b) 최상위 z.ai식 `reasoning_effort`는 무시된다(실증),
(c) env→요청 매핑은 Claude Code 내부 동작으로 직접 관측 범위 밖이나 세션 응답의
thinking 블록이 간접 종단 증거. **"reasoning_effort라는 z.ai 필드를 통과"가 아니라
"thinking budget를 통과"가 실증된 통로다** — llm.yaml effort 블록의 `reasoning-max`
레이블은 이 통로와 무관하게 저장만 존재(아래 §4).

## 4. llm.yaml glm.effort 블록 (stored-only 사실 확인)

`.moai/config/sections/llm.yaml` glm.effort = `{high: reasoning-max, medium:
reasoning-max, low: ...}` (리드가 기록). `internal/template/glm_effort_overlay.go`
및 호출경로는 이 블록을 읽지 않는다 — overlay는 Claude effort 상수에서 파생
(§1 표). 즉 블록은 문서적 의도 기록이고 코드 경로는 무관 — 카드 조사 항목으로
"설정-코드 불일치"로 기록. 변경이 코드(collapse)에 적용되므로 블록과의 정합은
`high/medium→reasoning-max` 레이블과 방향 일치(레이블의 통로 가정은 §3 P2로 반증).

## 5. SessionGLMReasoningState 처분 — 판정 근거 재료 (plan이 결정)

- 현행 `high`의 근거는 비용 정책 주석(:186-192): 세션-글로벌 값은 모든 스폰이 지불.
- 운영자 지시 원문은 collapse 매핑(high/medium→max)에 관한 것 — 세션 기본값(서브에이전트·빈 effort 폴백)의 처분은 명시되지 않음.
- 비용 관측 재료: t127 실측에서 무도구 trivial 스폰의 subagent_tokens=0 — 스폰당 고정 reasoning 비용은 작으나, 대형 호출에서 max의 reasoning 토큰 증분은 실재(P1/P3 구조).
- 후보: (a) `max` 상향 — 지시 정신과 일관, 비용 정책 주석은 운영자가 뒤집은 것으로 봄 (b) `high` 유지 — 서브에이전트·프로브의 비용 보존, 지시는 매핑에 한정 해석 (c) 폴백만 `low` — 가장 보수적. plan이 근거를 세우고 리드 비준.

## 6. Gaps

- env→요청 본문 매핑(Claude Code 내부)의 직접 캡처는 불가 — 간접 종단 증거(세션 thinking 블록)로 대체.
- P2의 무시가 "에러 없이 무시"인지 "다른 필드명으로는 수용"인지(예: `thinking.type:"adaptive"` + z.ai 확장)는 미탐 — 카드 범위 밖(변경은 env 값 상이지 요청 형태 변경이 아님).
- 고부하 태스크에서의 high vs max reasoning 토큰 증분 정량은 미측정 (trivial 태스크는 깊이가 수요 결정이라 차이가 없음 — P1 vs P3 output 차이는 태스크 난이도로 설명됨).
