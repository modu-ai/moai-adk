---
title: plan_type 티어 프로필
weight: 4
draft: false
---

MoAI-ADK는 동일한 워크플로우라도 API 종량제와 구독 요금제는 최적 배분이 다르다는 점을 인식합니다. `plan_type` 축은 요금제별로 Tier × Phase 모델/effort 매트릭스를 분리 적용합니다. 이 페이지는 SPEC-MODEL-TIER-PLANTYPE-001(CLOSED)에서 구현된 60-셀 프로필 매트릭스를 공식 문서화합니다.

## plan_type 축

`plan_type`은 두 가지 값을 가집니다:

- `api` — 종량제. 달러가 유일한 제약입니다. 과제당 비용 최적화가 목표입니다.
- `subscription` — 구독 요금제. 주간 토큰 쿼터 + Opus 가중 차감이 제약입니다. 쿼터당 해결 과제 수 극대화가 목표입니다.

구독 요금제에서 Opus 시간은 별도 가중 차감됩니다(Max 5x: Sonnet 140-280h vs Opus 15-35h, 약 1/8). 따라서 구독에서는 Opus를 추론에만 배치하고 실행은 풍부한 Sonnet 시간으로 돌리는 opusplan 구조가 최적입니다.

## plan_type 설정

```bash
moai init . --plan-type api           # 초기화 시 설정
moai update --plan-type subscription  # 사후 전환
```

`llm.yaml`의 `llm.plan_type` 필드에서 현재 값을 확인할 수 있습니다.

## 60-셀 프로필 매트릭스

10개 에이전트 × 3개 티어 × 2개 plan_type = 60셀. 아래 표는 SPEC-MODEL-TIER-PLANTYPE-001의 ApplyTierProfile 구현입니다.

### Plan A — API 종량제 (rev2)

API에서는 달러가 유일한 제약입니다. rev2 수정: 단가는 Sonnet이 Opus의 절반이지만, 과제당 비용은 Opus $13.22 < Sonnet $26.40으로 역전합니다. 따라서 API에서는 실행도 Opus를 사용합니다. 추론 = Fable(품질 1위), 실행 = Opus(과제당 비용 1위), 기계 = Sonnet low.

| 에이전트 (역할) | A-max (품질) | A-medium (권장) | A-low (비용) |
|---|---|---|---|
| manager-spec (추론) | fable / high | fable / high | opus / high |
| plan-auditor (추론) | fable / high | fable / high | opus / high |
| sync-auditor (추론) | fable / high | opus / high | opus / medium |
| manager-design (추론) | fable / high | fable / high | opus / high |
| super-advisor (최고추론) | fable / xhigh | fable / high | opus / high |
| manager-develop (실행) | fable / high | opus / high | opus / medium |
| builder-harness (실행) | opus / high | opus / medium | opus / medium |
| manager-docs (기계) | sonnet / medium | sonnet / low | sonnet / low |
| manager-git (기계) | sonnet / low | sonnet / low | sonnet / low |
| Explore (탐색) | inherit / medium | inherit / low | inherit / low |

### Plan B — 구독 요금제 (가용성 최우선)

구독의 제약은 달러가 아니라 주간 토큰 쿼터 + Opus 가중 차감입니다. 목표는 "쿼터당 해결 과제 수" 극대화 = 재시도 루프 배제 + Opus는 추론에만 배정. Anthropic 공식 opusplan 패턴("계획은 Opus, 실행은 Sonnet")의 정밀 버전입니다.

| 에이전트 (역할) | B-max (권장) | B-medium | B-low (Pro) |
|---|---|---|---|
| manager-spec (추론) | opus / high | opus / high | opus / medium |
| plan-auditor (추론) | opus / high | opus / medium | sonnet / high |
| sync-auditor (추론) | opus / high | opus / medium | sonnet / high |
| manager-design (추론) | opus / high | opus / medium | sonnet / high |
| super-advisor (최고추론) | opus / xhigh | opus / high | opus / medium |
| manager-develop (실행) | sonnet / high | sonnet / high | sonnet / high |
| builder-harness (실행) | sonnet / high | sonnet / medium | sonnet / medium |
| manager-docs (기계) | sonnet / low | sonnet / low | sonnet / low |
| manager-git (기계) | sonnet / low | sonnet / low | sonnet / low |
| Explore (탐색) | inherit / medium | inherit / low | inherit / low |

## ApplyTierProfile 메커니즘

`ApplyTierProfile`은 에이전트 frontmatter의 `model`과 `effort`를 모두 교체합니다(replace-both). 전 에이전트에 `effort:` 필드가 있어 "보존" 모드가 무효이므로, 항상 replace-both로 동작합니다.

이 메커니즘은 SPEC-MODEL-TIER-PLANTYPE-001(run-phase 완료, CLOSED)에서 구현되었습니다. 위 표의 모든 셀은 라이브 동작으로 검증되었습니다.

## GLM 백엔드 effort 오버레이

{{< icon warning warn >}} **정직성 고지 (REQ-DA-060)**: GLM 백엔드 effort 오버레이의 wire 유효성은 라이브 GLM 세션 아웃바운드 관측이 필요한 실증 과제입니다.

GLM 백엔드(`moai glm` / `moai cg` GLM 패널)에서는 Claude의 5단 effort(max / xhigh / high / medium / low)를 GLM의 3단 reasoning_effort(high / max)로 collapse하여 적용합니다. 구현 내용:

- `IsGLMBackend` 감지로 GLM 세션 식별
- 5단 → 3단 collapse 매핑 (max/xhigh → max, high → high, medium/low → GLM 미지원)
- coding 작업 시 max override

**구현 + 배선 완료, wire 유효성 실증 예정** — z.ai가 Anthropic-compat shim으로 `ANTHROPIC_REASONING_EFFORT` 환경변수 값을 실제로 소비하는지는 라이브 GLM 세션 아웃바운드 관측이 필요한 run-phase 실증 과제입니다. 이 페이지에 "동작 보장"으로 서술하지 않으며, "구현 + 배선 완료, wire 유효성 실증 예정"으로 기재합니다.

## 모델 정책 보드 (moai web)

`moai web`의 `/model-policy` 보드에서 plan_type과 티어 프로필을 시각적으로 확인하고 설정할 수 있습니다. 이 보드는 SPEC-WEB-CONSOLE-013의 승인된 예외로서 plan_type 쓰기를 허용합니다.

## 로드맵

{{< icon clock >}} **spawn-time 36-셀 라우팅** (SPEC-MODEL-TIER-ROUTING-PROFILES-001) — 현재 ApplyTierProfile은 에이전트 단위 라우팅입니다. spawn-time에 phase와 SPEC Tier를 조합한 36-셀 정밀 라우팅은 descoped된 후속 SPEC입니다. 현재는 에이전트 frontmatter의 model/effort가 ApplyTierProfile에 의해 replace-both 되는 구조로 운영됩니다.

## 다음 단계

- [3-티어 에이전트 아키텍처](/ko/advanced/no-haiku-3tier/) — DeepSWE 리더보드 근거와 3-티어 정의
- [토크노믹스 개요](/ko/advanced/tokenomics-overview/) — 4-층 토크노믹스 구조의 B층 라우팅
