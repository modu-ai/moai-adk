# 이미지 생성 asset 계약 조사 — 권리·비용·visual QA 분리 (#2174 → #2203, REQ-F5-01)

상위: #2158 (EP-F). 설계 근거: `references/chatgpt-space/page_e3612b0b07888191ae58c58d66bdff4a.md`(사본 sha256 `2c98c03db59d0c39e8f91929d9175201d476f295ff2c9422772a6bcea32626fc`, 2026-11-09 읽음) 이미지 workflow 절 L157.

## 조사 방법

- 외부 사실은 provider 공식 문서 기준으로 확인했습니다. 확인 경로는 Context7 문서 미러(`/websites/developers_openai_api`, `/websites/z_ai`, `/websites/playwright_dev`, 2026-11-09 질의)이며 각 사실에 공식 URL과 확인일을 붙였습니다.
- **REQ-F5-01 준수**: 이 조사는 문서 조회만 수행했습니다 — 이미지 생성 API 호출·유료 분석 호출 0건입니다. 미확인 항목은 추정으로 채우지 않고 명시합니다.
- 확인 날짜: **2026-11-09**(아래 외부 사실에 동일 적용).

## 설계 전제 (내부 근거, e361 이미지 workflow)

이미지 생성은 판단 API의 부수 효과가 아닌 별도 `media.generate` workflow다. 입력: brief·용도·**권리 있는 참조 자료**·크기/quality/count·provider·**비용 상한**·저장·공유 위치. 외부 전송 대상과 개인정보를 확인하고 승인 범위 안에서만 호출한다. 출력: 이미지 원본·content digest·prompt/version provenance·**reference 권리 정보**·생성 시각·비용·**visual QA 기록**. generated asset은 실제 화면·텍스트·크롭·투명도·접근성 검사를 거친 뒤 프로젝트에 연결한다. source asset과 derived variant를 구분하고 원본을 덮어쓰지 않는다. 비용: provider hard limit도 집계 지연으로 미세 초과 가능하므로 "절대 0원 초과 보장"으로 표현하지 않는다.

## 외부 사실 (URL + 확인일 2026-11-09)

### 권리·소유권

| # | 사실 | URL |
| --- | --- | --- |
| R1 | z.ai 이용약관: 사용자는 **자신이 제출한 prompt와 자신의 요청에 대해 생성된 output**에 대한 권리를 보유한다(당사자 간 최대 범위). 단 output이 고유하다는 보장은 없고 다른 사용자·제3자의 output과 유사·동일할 수 있어, 특정 output에 대한 권리가 타인의 output까지 미치지는 않는다 | https://docs.z.ai/legal-agreement/terms-of-use |
| R2 | OpenAI 측 소유권은 **이용자와의 Agreement가 지배**한다(공식 문서가 "output ownership rights remain governed by your Agreement with OpenAI"로 표기 — custom voices 문서이지만 일반 원칙 표기로 인용) | https://openai.com/policies/service-terms/ |
| R3 | OpenAI 생성물에는 **C2PA Content Credentials과 SynthID 워터마크** provenance 신호가 붙을 수 있으며, `content_provenance_checks` API로 관측 가능하다. 단 이 API는 일반 AI 탐지기가 아니고, 메타데이터 제거·변형·구형 모델·타사 생성물은 미탐지될 수 있다고 공식 문서가 명시한다 | https://developers.openai.com/api/docs/guides/content-provenance |

### 비용 경계

| # | 사실 | URL |
| --- | --- | --- |
| C1 | z.ai 이미지 생성은 **장당 과금**: CogView-4 $0.01/image, GLM-Image $0.015/image (공식 pricing 페이지) | https://docs.z.ai/guides/overview/pricing |
| C2 | OpenAI 이미지 모델은 gpt-image 계열로 **토큰 과금**(입력/출력/캐시 입력 1M 토큰 단위, 예: gpt-image-2 text $5 · image $30 · cached $2)과 gpt-image-1-mini의 **장당 과금**(1024x1024 high $0.036/image)이 병존한다. gpt-image-1은 2026-10-23 종료(공식 문서 명시) | https://developers.openai.com/api/docs/pricing |
| C3 | 이미지 생성 파라미터(quality·size·background·output_format·n 등)는 **공식 API reference**에 문서화되어 있으며 quality·size 선택이 비용·출력에 직결된다 | https://developers.openai.com/api/reference/resources/images/methods/generate |

### visual QA 분리 (DOM/AX vs visual)

| # | 사실 | URL |
| --- | --- | --- |
| Q1 | Playwright는 **접근성 스냅샷 단언**(`toMatchAriaSnapshot`, `.aria.yml`)과 구조 단언(locator·getByRole)을 제공 — DOM/AX 축의 검사는 구조·접근성 검증이다 | https://playwright.dev/docs/api/class-locatorassertions |
| Q2 | Playwright의 **visual 비교는 별도 단언**(`toHaveScreenshot`) — PNG/WebP 스냅샷 pixel diff로 `maxDiffPixels`·`maxDiffPixelRatio`·YIQ 색 공간 `threshold`·`mask`·애니메이션 비활성화 옵션을 갖는다. 구조 단언과 픽셀 단언은 서로 다른 API 계열이다 | https://playwright.dev/docs/test-snapshots , https://playwright.dev/docs/api/class-locatorassertions |
| Q3 | Playwright visual comparison은 **동일 엔진에서 두 번 연속 캡처가 일치할 때까지 대기**하는 안정화 절차를 내장한다(애니메이션·캐럿 처리 포함) — 픽셀 단언의 불안정 요소를 줄이는 공식 기능 | https://playwright.dev/docs/api/class-locatorassertions |

## 미확인 항목

| 항목 | 상태 |
| --- | --- |
| 생성 이미지의 **상업적 이용 가능성에 대한 법적 판단** — 약관 조항 인용은 법률 자문이 아니다 | 미확인(법률 검토 필요) |
| OpenAI `content_provenance_checks` API의 과금 여부 | 미확인(호출하지 않았음 — REQ-F5-01) |
| z.ai 이미지 생성 output의 C2PA/SynthID 등 provenance 신호 부착 여부 | 미확인(공식 문서에서 해당 절을 확인하지 못했다) |
| 참조 자료에 제3자 저작물이 포함될 때의 2차적저작물성 판단 | 미확인(법률 검토 필요) |

## 양립 분석 — 설계 media.generate 요구와 외부 사실 매핑

| 설계 요구 | 외부 근거 | 계약 반영 방식(제안) |
| --- | --- | --- |
| reference 권리 정보 입력 | R1: z.ai 조항은 사용자 prompt·output 권리를 인정하되 타인 output과의 유사 가능성을 명시 | asset 레코드에 `reference_rights` 필드(사용자 확인 스트링)와 `provider_terms_ref`(약관 URL·확인일)를 의무 기록 |
| 생성 시각·비용 기록 | C1·C2: provider별 과금 단위가 다름(장당 vs 토큰) | asset 레코드에 `cost`를 provider 과금 단위 그대로 기록(환산 금지) — provider 가격표 URL·확인일 동봉 |
| prompt/version provenance | R3: C2PA·SynthID 신호 존재, 변형·제거 가능성 명시 | `provenance`에 C2PA/SynthID 관측 결과를 **관측값으로만** 기록(보증 아님), 미탐지 가능성도 함께 기록 |
| visual QA 기록 | Q1·Q2: DOM/AX 단언과 픽셀 단언이 별도 API 계열 | QA 기록을 두 절로 분리 — (1) 구조·접근성(aria snapshot·locator 단언), (2) visual(스냅샷 pixel diff, threshold·mask 기록) |
| 실제 화면·텍스트·크롭·투명도·접근성 검사 | Q2의 `omitBackground`(투명도)·`mask`·애니메이션 처리 등 공식 옵션 | 검사 항목별 사용 옵션을 QA 기록에 남긴다 |
| 비용 상한·절대 0원 보장 금지 | C1·C2: 과금 단위가 provider마다 다르고 모델 수명주기가 있다(gpt-image-1 종료) | 승인 상한 도달 시 중단 규칙 유지 + provider 모델 수명주기(종료·대체) 확인을 비용 확인 절차에 추가 |

## 결론 (채택/보류)

**채택**:
1. **권리 확인 방법** — (a) provider 이용약관의 소유권 조항 인용을 asset 레코드에 기록(R1·R2), (b) 참조 자료의 권리는 사용자 확인 스트링으로 기록(자동 판정 아님 — 법적 판단은 보류), (c) provenance 신호(C2PA/SynthID)는 관측값으로만 기록하고 보증으로 쓰지 않는다(R3).
2. **비용 경계** — provider 공식 가격표 URL·확인일을 asset 레코드에 남기고, 과금 단위(장당/토큰)를 환산 없이 그대로 기록한다. 승인 상한 초과 시 중단은 설계대로 유지하되 "절대 0원 보장" 표현은 설계 문구대로 금지한다.
3. **visual QA 분리** — DOM/AX(구조·접근성) 검사와 visual(픽셀) 검사를 별도 단계·별도 기록으로 수행한다(Q1·Q2). 생성물 검사는 실제 화면·텍스트·크롭·투명도·접근성 항목을 모두 포함하고 사용한 옵션을 기록한다(Q3).

**보류**:
1. 생성 API 호출 자체 — REQ-F5-01(비용 호출 금지). 구현은 별도 카드에서 승인 절차를 거친다.
2. 자동 권리·적합성 판정 — 법률 검토 대상(미확인 표).
3. 특정 provider 지정 — 가격·모델 수명주기가 빠르게 변하므로(C2의 종료 사례) 계약 확정 시 재확인.
4. `content_provenance_checks` API 사용 — 과금 여부 미확인(호출하지 않음).

## not_run

- 이미지 생성 API 호출 — REQ-F5-01 위반이므로 0건
- `content_provenance_checks` 호출 — 과금 미확인
- 실제 visual QA 실행(생성물 부재) — 절차 설계만 확정
- z.ai output의 provenance 신호 실측 — 공식 문서 미확인
