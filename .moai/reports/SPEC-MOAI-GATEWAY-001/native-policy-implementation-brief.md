# Native 정책 본문 보강과 구현 인계

## Claim

독립 후보 검토 `native-policy-plan-review.md` PASS 이후 코어 0.9.0의 기존 요구·인수 조건 안에서 native 정책 계약을 보강했다. version/status는 0.9.0/draft를 유지했다. 제품 활성화·실제 통합 PASS를 선언하지 않는다.

수정 파일은 코어 spec.md, plan.md, acceptance.md, design.md, research.md 다섯 개다. design.md:404의 정책 표·응답 문법·title/high/keep-all/metadata/계량 계약, acceptance.md:548의 기존 AC 보강, plan.md:538의 구현 순서가 인계 기준이다. 승인된 구독 출력 분기와 receipt 본문을 보존했다. AUTH Windows·PICKER·progress·제품 코드는 수정하지 않았다.

## Evidence

이 작업 트리에서 실행:

```text
/tmp/moai-gateway-working-20260911 spec lint SPEC-MOAI-GATEWAY-001
✓ No findings — all SPEC documents are valid
```

변경 직전 `/tmp/native-policy-amend-before.json`과 Python SHA-256 비교:

```text
RECEIPT_SHA 59560f75191398a5b3e0a9a531e088b126f0f0f1a52927e23b74125913f59ad6 MATCH True
progress.md UNCHANGED
```

receipt 비교 범위는 design.md의 `선택된 설계 후보 — 활성화 게이트 유지 (0.9.0).` 마커 직후부터 `## 5.` 직전까지다. `/tmp/gateway-after19-spec-before.json`의 정렬된 ID 집합과 현재 정규식 추출 집합 비교 결과:

```text
REQ_IDS_UNCHANGED True
AC_IDS_UNCHANGED True
```

전체 REQ ID 26개와 AC ID 25개가 유지된다. 기존 묘비를 제외한 활성 요구·AC는 24/24다. lint는 문서 구조 확인이며 의미 감사나 runtime 판정이 아니다.

## Baseline-attribution

작업 트리 `.claude/worktrees/moai-proxy-unified`, WT-unified-gateway 기준의 2026-09-11 본문 보강이다. 실측 입력 근거는 native-policy-readiness-observation.md, title-policy-runtime-observation.md이며, Codex 수치 근거는 codex-context-catalog-observation.md의 pristine 공식 snapshot이다. 이번 작업은 새 provider 요청을 보내지 않았다. 실제 title 네 모델의 200은 직접 preflight이고 제품 native 변환 증거가 아니다.

## 구현 책임과 순서

1. Native/GPT 정책 작성자는 요청 문법과 provider별 wire 매핑, native thinking/signature 응답 문법, title 결과 검증, 입력 계량 투영과 그 focused tests를 맡는다. 예상 접점은 internal/gateway/anthropic.go, translate/request.go의 정책 분기, native 응답 validator, input_estimate.go다. 정확한 파일 소유는 배정 시 부모가 기존 receipt codec 작성자와 조정한다.
2. receipt 패키지·codec의 envelope 및 hash 형식·정규화·저장·수명·launcher는 이번 정책 작성자의 소유가 아니다. metadata의 세션 귀속 검증은 기존 receipt 진입점을 먼저 호출하며, 정책 작성자가 별도 receipt를 만들지 않는다. 같은 translator 파일에 접점이 있으면 지정 구간만 순차 인계한다.
3. 실제 raw003~008에서 policy 부분만 안전하게 추출한 RED로 시작한다. title/high와 main adaptive/high/keep-all 양성, 누락 대 null, 타입·중복·unknown·범위·미지원 조합의 송신 0을 검사한다. native history/응답의 빈 thinking·signature delta를 보존하고 text/tool SSE 및 nonstream 회귀 대조군을 유지한다.
4. GPT title의 정확한 schema와 고정 name, high, 공개 history/opaque 순서, Claude 계정·장치 metadata 외부 유출 0을 판정한다. 일반 schema 축소·disabled의 low 대체·manual budget의 high 대체를 금한다. 이미 구현된 승인 출력 상한 분기는 보존하고 별도 재설계하지 않는다.
5. 계량에 system/history/tools/output schema를 반영한다. 272000 기본·872000 override 최대·95% headroom은 출처가 있는 선언으로 표현한다. 실제 계정 cap, 이미지 수용 또는 tokenizer 상한으로 표시하지 않는다. Claude/GLM cap는 공식/current native 근거를 확보하는 연구 게이트이며 추측값으로 채우지 않는다.
6. 정책 단위·wire 대조 후 실제 title와 main, 네 GPT 도구/회상/재개 통합을 수행한다. receipt 음성군과 출력 상한·취소·잘림 회귀를 유지한다. AUTH Windows 작성자의 파일·CI 실행 책임을 침범하지 않는다.

## Gaps

새 본문의 독립 감사와 제품 구현·네 모델 통합 실행은 이 인계에서 수행하지 않았다. native Claude/GLM의 정확한 cap, 일반 JSON Schema 지원, 수동 reasoning budget 대응, disabled의 GPT별 동등 대응, 실계정 대형 입력 및 이미지 수용은 근거 게이트를 유지한다. 새로운 사용자 결정 없이 근거 확인과 구현을 이어갈 수 있으며 이 미확인을 전체 목표에서 삭제하지 않는다.

## Residual-risk

현재 관측된 native 입력 밖의 필수 변형이 추가로 나올 수 있다. 이를 허용되지 않은 키의 무조건 통과나 사용자 정책의 조용한 삭제로 해결하지 않는다. 출처가 있는 context 선언도 특정 계정·endpoint의 실측 한도와 다를 수 있다. native 서명의 문법 보존은 서명 진위 검증과 다르다.
