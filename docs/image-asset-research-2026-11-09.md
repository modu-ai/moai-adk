# 이미지 생성 asset 계약 조사 — 권리·비용·visual QA 분리 (#2174 → #2203, REQ-F5-01)

상위: #2158 (EP-F). 설계 근거: `references/chatgpt-space/page_e3612b0b07888191ae58c58d66bdff4a.md`(사본 sha256 `2c98c03db59d0c39e8f91929d9175201d476f295ff2c9422772a6bcea32626fc`, 2026-11-09 읽음) 이미지 workflow 절 L157.

## 조사 방법

- 외부 사실은 provider 공식 문서 기준으로 확인했습니다. 확인 경로는 Context7 문서 미러(`/websites/developers_openai_api`, `/websites/z_ai`, `/websites/playwright_dev`, 2026-11-09 질의)이며 각 사실에 공식 URL과 확인일을 붙였습니다.
- **REQ-F5-01 준수**: 이 조사는 문서 조회만 수행했습니다 — 이미지 생성 API 호출·유료 분석 호출 0건입니다. 미확인 항목은 추정으로 채우지 않고 명시합니다.
- 확인 날짜: **2026-11-09**(아래 외부 사실에 동일 적용).
- 수리 이력: agent-9의 공식 페이지 재조회 값과 초안을 #2203 comment 6734927875로 인계받았습니다. 이번 epoch 4에서는 2026-11-09T08:18:26Z~08:18:27Z에 아래 표의 공식 Markdown/HTML 원문을 직접 GET하고 응답 바이트 SHA256을 기록했습니다. `service-terms`는 HTTP 403을 관측한 뒤 재시도·우회하지 않았습니다. 기존 Context7 조사 중 직접 재조회하지 않은 R1·C1·R3·C3는 기존 조사 기록을 유지한 것으로, 이번에 다시 검증했다고 주장하지 않습니다.

## 설계 전제 (내부 근거, e361 이미지 workflow)

이미지 생성은 판단 API의 부수 효과가 아닌 별도 `media.generate` workflow다. 입력: brief·용도·**권리 있는 참조 자료**·크기/quality/count·provider·**비용 상한**·저장·공유 위치. 외부 전송 대상과 개인정보를 확인하고 승인 범위 안에서만 호출한다. 출력: 이미지 원본·content digest·prompt/version provenance·**reference 권리 정보**·생성 시각·비용·**visual QA 기록**. generated asset은 실제 화면·텍스트·크롭·투명도·접근성 검사를 거친 뒤 프로젝트에 연결한다. source asset과 derived variant를 구분하고 원본을 덮어쓰지 않는다. 비용: provider hard limit도 집계 지연으로 미세 초과 가능하므로 "절대 0원 초과 보장"으로 표현하지 않는다.

## 외부 사실 (URL + 확인일 2026-11-09)

### 권리·소유권

| # | 사실 | URL |
| --- | --- | --- |
| R1 | z.ai 이용약관: 사용자는 **자신이 제출한 prompt와 자신의 요청에 대해 생성된 output**에 대한 권리를 보유한다(당사자 간 최대 범위). 단 output이 고유하다는 보장은 없고 다른 사용자·제3자의 output과 유사·동일할 수 있어, 특정 output에 대한 권리가 타인의 output까지 미치지는 않는다 | https://docs.z.ai/legal-agreement/terms-of-use |
| R2 | OpenAI custom voices 가이드에는 output 소유권이 이용자와 OpenAI 사이 Agreement에 따른다는 설명이 있다. **확장 해석의 한계**: 해당 문장은 음성(prompt-to-voice) 절의 사실이며, 이를 이미지 output의 소유권 일반 규칙으로 읽는 것은 이 보고서의 확장 해석이다. 이미지 약관에서 직접 확인한 사실로 쓰지 않는다. `service-terms`는 이번 GET도 HTTP 403이라 원문을 확인하지 못했으므로 이미지 소유권 조건은 미확인으로 보류한다 | https://developers.openai.com/api/docs/guides/custom-voices |
| R3 | OpenAI 생성물에는 **C2PA Content Credentials과 SynthID 워터마크** provenance 신호가 붙을 수 있으며, `content_provenance_checks` API로 관측 가능하다. 단 이 API는 일반 AI 탐지기가 아니고, 메타데이터 제거·변형·구형 모델·타사 생성물은 미탐지될 수 있다고 공식 문서가 명시한다 | https://developers.openai.com/api/docs/guides/content-provenance |

### 비용 경계

| # | 사실 | URL |
| --- | --- | --- |
| C1 | z.ai 이미지 생성은 **장당 과금**: CogView-4 $0.01/image, GLM-Image $0.015/image (공식 pricing 페이지) | https://docs.z.ai/guides/overview/pricing |
| C2 | OpenAI gpt-image 계열 가격표는 **Standard / Batch 요금방식**과 **Modality별 Input / Output / Cached input**을 나눈다(USD/1M tokens). gpt-image-2 — Standard: Image $8 / $30 / $2, Text $5 / – / $1.25. Batch: Image $4 / $15 / $1, Text $2.50 / – / $0.625. **캐시 요율 적용 범위**: GPT Image 2의 cached input은 Responses API image generation tool에만 적용되며 직접 Images API 요청에는 적용되지 않는다 | https://developers.openai.com/api/docs/pricing , https://developers.openai.com/api/docs/guides/image-generation#cached-input-pricing |
| C2a | 모델 페이지의 장당 출력 가격 예시: gpt-image-1-mini, Quality High, 1024x1024 **$0.036/image**. 이것만으로 전체 요청 비용을 뜻하지 않으며 prompt text 입력과 편집용 image 입력 토큰 비용은 별도로 더해야 한다 | https://developers.openai.com/api/docs/models/gpt-image-1-mini , https://developers.openai.com/api/docs/guides/image-generation#historical-output-costs |
| C2b | 모델 수명주기: gpt-image-1은 2026-10-23 종료(대체 gpt-image-2.5-sunburst / -flare). gpt-image-1-mini·gpt-image-1.5·chatgpt-image-latest는 2026-12-01 종료 예정(대체 gpt-image-2.6-sunburst / -flare). image-generation 가이드도 "gpt-image-1 was shut down on October 23, 2026"이라고 적는다 | https://developers.openai.com/api/docs/deprecations , https://developers.openai.com/api/docs/guides/image-generation |
| C3 | 이미지 생성 파라미터(quality·size·background·output_format·n 등)는 **공식 API reference**에 문서화되어 있으며 quality·size 선택이 비용·출력에 직결된다 | https://developers.openai.com/api/reference/resources/images/methods/generate |

### visual QA 분리 (DOM/AX vs visual)

| # | 사실 | URL |
| --- | --- | --- |
| Q1 | Playwright는 **접근성 스냅샷 단언**(`toMatchAriaSnapshot`, `.aria.yml`)과 구조 단언(locator·getByRole)을 제공 — DOM/AX 축의 검사는 구조·접근성 검증이다 | https://playwright.dev/docs/api/class-locatorassertions |
| Q2 | Playwright의 **visual 비교는 별도 단언**(`toHaveScreenshot`) — PNG/WebP 스냅샷 pixel diff로 `maxDiffPixels`·`maxDiffPixelRatio`·YIQ 색 공간 `threshold`·`mask`·애니메이션 비활성화 옵션을 갖는다. 구조 단언과 픽셀 단언은 서로 다른 API 계열이다 | https://playwright.dev/docs/test-snapshots , https://playwright.dev/docs/api/class-locatorassertions |
| Q3 | Playwright visual comparison은 **두 번 연속 캡처가 일치할 때까지 대기**하는 안정화 절차를 내장한다(애니메이션·캐럿 처리 포함) — 픽셀 단언의 불안정 요소를 줄이는 공식 기능 | https://playwright.dev/docs/api/class-locatorassertions |

## 미확인 항목

| 항목 | 상태 |
| --- | --- |
| 생성 이미지의 **상업적 이용 가능성에 대한 법적 판단** — 약관 조항 인용은 법률 자문이 아니다 | 미확인(법률 검토 필요) |
| OpenAI 이미지 output 소유권에 적용되는 service-terms 원문 | 미확인(이번 직접 GET HTTP 403, 재시도·우회 없음). R2의 음성 문서를 이미지 권리 보증으로 사용하지 않는다 |
| OpenAI `content_provenance_checks` API의 과금 여부 | 미확인(호출하지 않았음 — REQ-F5-01) |
| z.ai 이미지 생성 output의 C2PA/SynthID 등 provenance 신호 부착 여부 | 미확인(공식 문서에서 해당 절을 확인하지 못했다) |
| 참조 자료에 제3자 저작물이 포함될 때의 2차적저작물성 판단 | 미확인(법률 검토 필요) |

## 양립 분석 — 설계 media.generate 요구와 외부 사실 매핑

| 설계 요구 | 외부 근거 | 계약 반영 방식(제안) |
| --- | --- | --- |
| reference 권리 정보 입력 | R1: z.ai 조항은 사용자 prompt·output 권리를 인정하되 타인 output과의 유사 가능성을 명시 | asset 레코드에 `reference_rights` 필드(사용자 확인 스트링)와 `provider_terms_ref`(약관 URL·확인일)를 의무 기록 |
| 생성 시각·비용 기록 | C1·C2·C2a: provider·모델별 장당 요율·토큰 단가를 구분함. OpenAI는 Standard/Batch 요금방식·modality별 토큰 단가와 장당 출력가 예시를 병기함 | asset 레코드에 `cost`를 provider 과금 단위 그대로 기록(환산 금지) — provider 가격표 URL·확인일 동봉 |
| prompt/version provenance | R3: C2PA·SynthID 신호 존재, 변형·제거 가능성 명시 | `provenance`에 C2PA/SynthID 관측 결과를 **관측값으로만** 기록(보증 아님), 미탐지 가능성도 함께 기록 |
| visual QA 기록 | Q1·Q2: DOM/AX 단언과 픽셀 단언이 별도 API 계열 | QA 기록을 두 절로 분리 — (1) 구조·접근성(aria snapshot·locator 단언), (2) visual(스냅샷 pixel diff, threshold·mask 기록) |
| 실제 화면·텍스트·크롭·투명도·접근성 검사 | Q2의 `omitBackground`(투명도)·`mask`·애니메이션 처리 등 공식 옵션 | 검사 항목별 사용 옵션을 QA 기록에 남긴다 |
| 비용 상한·절대 0원 보장 금지 | C1·C2·C2b: 과금 단위가 provider마다 다르고 모델 수명주기가 있다(gpt-image-1 종료, mini 등 2026-12-01 종료 예정) | 승인 상한 도달 시 중단 규칙 유지 + provider 모델 수명주기(종료·대체) 확인을 비용 확인 절차에 추가 |

## 결론 (채택/보류)

**채택**:
1. **권리 확인 방법** — (a) 확인된 provider 약관의 소유권 조항과 출처를 asset 레코드에 기록한다(R1). OpenAI 이미지 약관은 R2의 한계에 따라 직접 확인될 때까지 미확인으로 둔다. (b) 참조 자료의 권리는 사용자 확인 스트링으로 기록한다(자동 판정 아님 — 법적 판단은 보류). (c) provenance 신호(C2PA/SynthID)는 관측값으로만 기록하고 보증으로 쓰지 않는다(R3).
2. **비용 경계** — provider 공식 가격표 URL·확인일을 asset 레코드에 남기고, 과금 단위(장당/토큰)를 환산 없이 그대로 기록한다. 승인 상한 초과 시 중단은 설계대로 유지하되 "절대 0원 보장" 표현은 설계 문구대로 금지한다.
3. **visual QA 분리** — DOM/AX(구조·접근성) 검사와 visual(픽셀) 검사를 별도 단계·별도 기록으로 수행한다(Q1·Q2). 생성물 검사는 실제 화면·텍스트·크롭·투명도·접근성 항목을 모두 포함하고 사용한 옵션을 기록한다(Q3).

**보류**:
1. 생성 API 호출 자체 — REQ-F5-01(비용 호출 금지). 구현은 별도 카드에서 승인 절차를 거친다.
2. 자동 권리·적합성 판정 — 법률 검토 대상(미확인 표).
3. 특정 provider 지정 — 가격·모델 수명주기가 빠르게 변하므로(C2b의 종료 사례 — gpt-image-1 종료, mini 등 2026-12-01 종료 예정) 계약 확정 시 재확인.
4. `content_provenance_checks` API 사용 — 과금 여부 미확인(호출하지 않음).


## 직접 조회 원문과 재현 정보

조회 시각은 각 응답을 읽은 시각이며 문서 발행일·모델 접근 권한·실제 과금 결과를 뜻하지 않는다. 아래 SHA256은 내려받은 응답 바이트 기준이다. Markdown에는 Standard/Batch 표가 모두 포함되어 있어 화면의 접힌 행이나 다른 요금방식의 열을 섞지 않고 대조할 수 있다.

| 원문 | 조회 시각(UTC) | HTTP | SHA256 |
| --- | --- | --- | --- |
| https://developers.openai.com/api/docs/pricing.md | 2026-11-09T08:18:26.890810Z | 200 | f1ade44afc62fd8a12dd6e199bcecaa989a94c9b691aa620a34e1504ff62ff57 |
| https://developers.openai.com/api/docs/guides/custom-voices.md | 2026-11-09T08:18:26.891579Z | 200 | 4002579024bb717313900837691eba530968b1602b028f00409aca8794cef1e2 |
| https://developers.openai.com/api/docs/models/gpt-image-1-mini.md | 2026-11-09T08:18:26.891986Z | 200 | 4d9ea36557d4b9800c46b5ff30f4e137f61a028a624d3910c2e1ca3949384495 |
| https://developers.openai.com/api/docs/deprecations.md | 2026-11-09T08:18:26.893475Z | 200 | 49e382e18bf8e14ea82d099a21a1f2b818b72603424518519eb908b419b45be1 |
| https://developers.openai.com/api/docs/guides/image-generation.md | 2026-11-09T08:18:27.027393Z | 200 | 52c43f91f992d61b2d7b3557c091771a0cc5cd62ff71972fb86de052d5751f4e |
| https://openai.com/policies/service-terms/ | 2026-11-09T08:18:27.036918Z | 403 | 원문 읽기 거부 — 해시 없음 |
| https://playwright.dev/docs/test-snapshots | 2026-11-09T08:18:27.047028Z | 200 | b9af9343958380a360222d6b53126eb978845e99a60e261fc19bc52c8f8c4c9b |
| https://playwright.dev/docs/api/class-locatorassertions | 2026-11-09T08:18:27.129254Z | 200 | e7ffe3bea09cf4b804be49ed838969101a3cd629b48738d64420bc25d0a8cec3 |

R2·C2·C2a·C2b는 이 원문에서 직접 대조했다. Playwright 두 원문에서도 PNG/WebP·연속 두 캡처 안정화 설명을 확인했다. 이미지 생성·provenance API 호출·실제 visual QA는 실행하지 않았다.

## not_run

- 이미지 생성 API 호출 — REQ-F5-01 위반이므로 0건
- `content_provenance_checks` 호출 — 과금 미확인
- 실제 visual QA 실행(생성물 부재) — 절차 설계만 확정
- z.ai output의 provenance 신호 실측 — 공식 문서 미확인
