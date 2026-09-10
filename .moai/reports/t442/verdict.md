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

`spec.md:45`(§1.2)와 `CHANGELOG.md:256`의 구 방향은 **만료된 기록**이다 — t175 시점(2026-08-22) 이후 엔드포인트 동작이 바뀌었거나 t175 측정 자체가 결함이었거나 둘 중 하나인데, 어느 쪽인지는 이 관측으로 구분되지 않는다(구분하려면 t175의 원측정 프롬브를 같은 조건으로 재현해야 하고, 본 카드는 그 재현을 하지 않았다). 어느 쪽이든 현재의 참이 아니므로 t225/t240 방향으로 정정한다. (리드 검토 2026-09-02: 초기본의 "(시차 2일뿐이라 전자가 유력)" 괄호는 근거 없는 원인 지목이라 삭제하고 위 문장으로 교체 — 원인 구분 시도 없이 "구분되지 않는다"만 남긴다.) **t175 원기록은 지우지 않는다** — `.moai/reports/t175/measurements.md`와 HISTORY 행에 superseded 이력으로 남는다. 슬롯·모델 키잉 문장(t360 축)은 일절 건드리지 않는다.

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

---

## 판정 보충 — 원인 구분 (2026-09-03, 운영자 지시 "판정 요청")

원 결론의 열린 절 — "엔드포인트가 바뀌었거나 t175 측정 자체가 결함이었거나, 구분되지 않는다(구분하려면 t175 원측정 재현 필요)" — 에 대한 추후 판정. **"t175 측정 설계 결함" 가설이 시간무관으로 입증됐고, "엔드포인트 변화" 가설은 이제 모순 설명에 불필요하다.**

### 설계 결함 3건 — t175 기록 자체의 숫자로 입증 (과거 관측 불요)

1. **P2는 구성상 무정보 프로브였다.** 시험값 `reasoning_effort: "max"`는 z.ai의 omit-default다(본 카드 측정: 양 모델 max ≈ null; REQ-GEM-005의 리드비준 근거 "max is z.ai's own omit-default"와도 일치). 요청값이 기본값과 같으면 "존중됨"과 "무시됨"을 그 프로브로는 구분할 수 없다. 프롬프트("Reply with the single word OK.")는 추론 수요가 0이라 깊이 판별도 불가 — 이중 무정보. 그런데도 P2 output=3이 "무시됨(MEASURED)"으로 기록됐다.
2. **"budget에 깊이 비례"는 t175 자기 숫자가 반대 방향이다.** P1(budget 2048) output=48 > P3(budget 32768) output=23 — 비례가 참이라면 P3 ≥ P1이어야 한다. t175는 §6 갭에서 스스로 "P1 vs P3 output 차이는 태스크 난이도로 설명됨"이라 기록했다 — 즉 당시에도 비례는 관측된 적이 없었고, 그럼에도 "depth scales with budget"가 MEASURED로 적혔다.
3. **thinking 블록 반환은 판별자가 아니다.** 항상-추론 바닥(본 카드 측정: null 조건에서도 thinking 블록 상시; SPEC C-5 "thinking cannot be disabled") 때문에 블록 존재는 thinking 파라미터 수용의 증거가 못 된다. P1의 "추론 블록 반환이염"은 무귀결 관측이었다.

### 오늘의 원 프로브 동일 재실행 (3회 호출)

`probe_shim.py`를 2026-09-03 그대로 실행 — 원문 출력 동봉: `.moai/reports/t442/t175-probe-rerun-2026-09-03.txt`

| 프로브 | t175 (2026-08-22) | 오늘 재실행 (2026-09-03) |
|---|---|---|
| P1 (budget 2048) | output=48, blocks=[thinking,text] | output=42, blocks=[thinking,text] |
| P2 (effort=max) | output=3, **blocks=[text]** | output=14, **blocks=[thinking,text]** |
| P3 (budget 32768) | output=23 | output=46 |

동일 프로브·동일 조건에서 **P2의 thinking 블록 유무가 실행 간에 뒤집힌다** — P2 관측이 판별자가 아니라 잡음이었다는 실증. budget 비례는 오늘도 부재(42 vs 46, 평탄).

### 판정

- **원인 구분: "측정 결함"으로 충분하다.** t175 프로브는 2026-08-22 당시에도 두 주장을 입증할 수 없는 설계였다 — 구 기록의 "MEASURED" 지위는 작성 시점에 이미 무근거였다. "엔드포인트가 바뀌었다"는 여전히 직접 관측 불가(08-22 상태는 재현 불가)하며, 참일 수 있으나 모순을 설명하는 데 필요하지 않다.
- **본 카드의 정정(spec.md v0.1.2 + CHANGELOG)은 불변** — 어느 원인이든 구 기록이 현재의 참이 아니라는 본 판정서의 결론과 무관하다.
- t175 원기록(`.moai/reports/t175/measurements.md`)은 계속 보존한다 — 이제 "결함 설계의 기록"으로서의 역사 가치를 가진다.
- 리드 검토 이력(2026-09-02): 초기본의 근거 없는 원인 지목 괄호 삭제 경위는 앞 절에 기록됨 — 본 보충은 그 열린 절을 닫는 것이지 삭제된 추론을 되살리는 것이 아니다.
