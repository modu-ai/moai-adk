# t442 판정서 — SPEC-GLM-EFFORT-MAX-001 §1.2 + CHANGELOG:256 구 방향 MEASURED 기록 정정

- 판정: **CONFIRMED — t225 신 방향이 현재 참 (직접 재현, 2026-09-02)**. 배차 전제("둘 중 하나는 틀렸다") 확정: t175 구 방향의 두 판별식이 모두 오늘 관측에서 반증됐다. DROPPED 아님.
- 충돌 기록 2벌: `SPEC-GLM-EFFORT-MAX-001/spec.md:45`(§1.2) + `CHANGELOG.md:256` — "shim은 Anthropic `thinking` 파라미터를 존중(depth가 budget에 비례)하고 top-level `reasoning_effort`는 조용히 무시한다" (MEASURED, 근거 `.moai/reports/t175/measurements.md`, 2026-08-22) ⟷ t225 실측(2026-08-24, glm_audit 경로, null-제어 differential — "top-level `reasoning_effort` 실효, thinking-budget 객체 무시") + t240 착지(2026-09-02, 문서 3종 정정).
- 본 측정이 채운 Gap: t240 판정서의 "z.ai 엔드포인트의 물리적 재측정 미수행".

## 경로 동일성 (변호 불성립 확인)

t240 판정서 Residual은 "t225는 audit 경로, t175는 세션 경로 — 물리적 동일성 미증명"을 남겼다. 확인 결과 **두 기록은 같은 대상을 재었다**: `internal/cli/mcp_glm.go:5`와 `internal/cli/mcp_server.go:389` — glm_audit 도 같은 Anthropic-compat 엔드포인트(`https://api.z.ai/api/anthropic/v1/messages`)를 친다. `internal/config/defaults.go:124` `DefaultGLMBaseURL` 동일. 경로 차이로 양쪽을 다 옳게 만드는 변호는 성립하지 않는다.

## 측정 설계

- 엔드포인트: `https://api.z.ai/api/anthropic/v1/messages`, 비스트리밍 HTTP POST, `Authorization: Bearer`
- 모델: `glm-5.3-flash`(기본 슬롯 모델, `defaults.go:157`) + 교차 확인 `glm-5.3`
- 프롬프트(전 조건 동일, 깊이 유발형): 동전 세기 문제 ("3 boxes × 4 bags × 5 coins, 2 coins removed from every second bag …")
- 조건: null / `reasoning_effort`: low·high·max / `thinking: {type: enabled, budget_tokens}`: 512·4096
- 관측치: `usage.output_tokens` + thinking 블록 총 길이(chars)
- 반복: flash n=5(null·low·t512·t4096), glm-5.3 n=3(null·low·high·max), 1차 탐색 라운드 4. 총 36 호출.
- raw 응답 36 파일: `.moai/reports/t442/raw/*.json` (키·비밀값 미포함 — 요청 본문에는 자격증명 없음)

## 결과 — thinking-chars (중앙값 | 평균)

| 조건 | glm-5.3-flash (n=5) | glm-5.3 (n=3) |
|---|---|---|
| null | 1358 \| 1560 | 839 \| 866 |
| effort=low | 983 \| 914 | 1286 \| 1349 |
| effort=high | — | 1406 \| 1534 |
| effort=max | — | 923 \| 933 |
| thinking 512 | 1221 \| 1125 | 미측정 |
| thinking 4096 | 698 \| 907 | 미측정 |

output_tokens 중앙값도 같은 순서(flash: null 664 / low 541 / t512 555 / t4096 409).

## 판별식 판정

1. **"thinking 존중, budget에 깊이 비례" (t175) → 반증.** flash t512 vs t4096: 중앙 1221 vs 698 — 비례 없음(오히려 역방향). budget을 늘려도 깊이가 늘지 않는다.
2. **"top-level `reasoning_effort` 무시(조용히 버림)" (t175) → 반증.** 두 모델 모두 필드 투입이 유의미한 깊이 변화를 만든다 — flash: low가 null 대비 감소(중앙 비 1.38, 평균 비 1.71), glm-5.3: low/high가 null 대비 증가. 무시라면 어느 쪽도 일어나지 않는다.
3. **t225 방향("`reasoning_effort` 실효, thinking-budget 객체 무시") → 확인.** 부가 관측: 양 모델에서 max ≈ null("max = omit-default", SPEC C-5와 정합).

## 결론

`spec.md:45`(§1.2)와 `CHANGELOG.md:256`의 구 방향은 **만료된 기록**이다 — t175 시점(2026-08-22) 이후 엔드포인트 동작이 바뀌었거나 t175 측정 자체가 결함이었거나 둘 중 하나인데, 어느 쪽인지는 이 관측으로 구분되지 않는다(시차 2일뿐이라 전자가 유력). 어느 쪽이든 현재의 참이 아니므로 t225/t240 방향으로 정정한다. **t175 원기록은 지우지 않는다** — `.moai/reports/t175/measurements.md`와 HISTORY 행에 superseded 이력으로 남는다. 슬롯·모델 키잉 문장(t360 축)은 일절 건드리지 않는다.

## Gaps

- thinking budget 판별식(1번)은 flash에서만 측정 — glm-5.3 budget 미측정.
- 단일 프롬프트, n=3~5, 비스트리밍 — 절대 깊이의 소표본 통계. 방향 판정에는 충분, 효과 크기 추정에는 불충분.
- 엔드포인트가 언제 구 동작에서 바뀌었는지 미확인.
- `ANTHROPIC_REASONING_EFFORT` env → Claude Code 실제 요청 본문 주입 체인의 end-to-end 재측정은 범위 밖(세션이 보내는 요청 본문은 미포착; 본 측정은 엔드포인트 직호출).
- flash에서 `reasoning_effort=max` 조건 미측정(1차 라운드 t512 프롬프트에서 간접 확인에 그침).

## Residual-risk

- **값별 효과 방향이 모델마다 다르다**(flash: low < null / glm-5.3: low > null, high 최대). "low=낮음"이라는 어휘 직관이 glm-5.3에서 성립하지 않는다. 정정 문장은 **"필드가 읽힌다"까지만 단언**하고 값별 순서·등급은 단언하지 않아야 한다 — 그 부분은 미해결 관측이다.
- flash에서 null > t4096(기대와 반대 방향의 유의미해 보이는 차이): thinking 객체의 존재 자체가 동작을 미세하게 바꿀 가능성. 소표본, 원인 미판명.
- `defaults.go:164`의 "flash는 `reasoning_effort` 'max'만 받는다" 주석과 긴장 — 오늘 flash는 low를 HTTP 200 + 실효로 수용했다. **코드 미수정**(모델·슬롯 어휘는 t360 축 인접), 관측만 기록.
- 단일 API 키·단일 리전의 관측 — z.ai 측 라우팅 변동 가능성은 통제되지 않았다.

## 5-섹션 요약

- **Claim**: 오늘(2026-09-02) 기준 z.ai Anthropic-compat 엔드포인트는 top-level `reasoning_effort`를 읽고, thinking-budget 객체는 깊이 조절 채널이 아니다. → t225 신 방향 참.
- **Evidence**: 위 결과 표 + raw 36 파일(`.moai/reports/t442/raw/`).
- **Baseline-attribution**: 본 워크트리(`.claude/worktrees/t442`, 분기 `WT-glm-effort-measured`, develop `349cadc93` 흡수 기준), 2026-09-02 본 런, 직호출 curl.
- **Gaps**: 상기 Gaps 절.
- **Residual-risk**: 상기 Residual-risk 절.
