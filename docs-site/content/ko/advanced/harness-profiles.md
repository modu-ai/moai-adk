---
title: 하네스 프로필과 평가 시스템
weight: 75
draft: false
---

모든 변경에 같은 깊이로 검증을 걸면 토큰이 낭비되고, 검증을 얕게 통일하면 품질이 떨어집니다. MoAI-ADK의 답은 **적응형 검증**입니다 — SPEC의 복잡도에 맞춰 검증 깊이를 자동으로 조절하고, 평가는 만든 쪽이 아니라 독립 평가자에게 맡깁니다.

## 개요

MoAI-ADK의 하네스(Harness)는 **3계층 적응형 품질 검증 시스템**입니다. SPEC의 복잡도에 따라 검증 깊이를 자동으로 조절하고, sync-auditor 에이전트가 4차원 스코어링으로 독립적이고 회의적인 품질 평가를 맡습니다. "된 것 같다"가 아니라 점수와 근거로 완료를 판정하는 구조입니다.

## 3계층 하네스 레벨

| 레벨 | 설명 | 적용 시점 | sync-auditor |
|------|------|----------|-----------------|
| **minimal** | 빠른 검증 | 단순 변경 (typos, 설정 수정) | 생략 가능 |
| **standard** | 기본 품질 검증 | 대부분의 작업 | 선택적 |
| **thorough** | 전체 검증 + TRUST 5 | 복잡한 SPEC, 대규모 변경 | 필수 |

하네스 레벨은 SPEC scope를 보고 **복잡도 추정기**(Complexity Estimator)가 자동으로 정합니다. 오타 수정에 thorough 검증을 돌리지 않는 것 — 그 자체가 토크노믹스입니다.

## 4차원 스코어링

sync-auditor는 네 가지 차원으로 점수를 매깁니다.

| 차원 | 설명 | 기본 Must-Pass |
|------|------|---------------|
| **Functionality** | 기능 완성도 — 의도된 목적을 달성했는가 | 예 |
| **Security** | 보안 — OWASP, 인증, 권한, 입력 검증 | 예 |
| **Craft** | 코드 품질 — 가독성, 구조, 테스트 커버리지 | 아니오 |
| **Consistency** | 일관성 — 프로젝트 규칙, 코드 스타일 준수 | 아니오 |

### 점수 범위

각 차원은 0.0 ~ 1.0 점수를 받습니다.

### 루브릭 앵커

점수가 평가자의 기분에 따라 흔들리지 않도록, 모든 평가 기준에는 4단계 루브릭 앵커를 붙입니다.

| 점수 | 수준 | 의미 |
|------|------|------|
| 0.25 | 미달 | 기본 요구사항 미충족 |
| 0.50 | 부분 | 일부 충족, 개선 필요 |
| 0.75 | 충족 | 대부분 충족, 소규모 개선 |
| 1.00 | 우수 | 모든 기준 완벽 충족 |

## 평가 프로필

`.moai/config/evaluator-profiles/`에 네 가지 프로필이 들어 있습니다. 작업 성격에 따라 평가 기준의 엄격함을 바꿔 쓰면 됩니다.

| 프로필 | 설명 | 적합한 경우 |
|--------|------|------------|
| `default.md` | 균형 잡힌 기본 프로필 | 대부분의 작업 |
| `strict.md` | 엄격한 기준 | 보안 중요 작업 |
| `lenient.md` | 관대한 기준 | 프로토타이핑 |
| `frontend.md` | 프론트엔드 특화 | UI/UX 작업 |

## 평가자 편향 방지 (5가지 메커니즘)

LLM 평가자는 그냥 두면 후해집니다. 이를 구조적으로 눌러두려고 다섯 가지 메커니즘이 함께 돕니다.

| # | 메커니즘 | 설명 |
|---|---------|------|
| 1 | **루브릭 앵커링** | 점수마다 루브릭 근거를 반드시 달아야 함 |
| 2 | **회귀 베이스라인** | 이전 프로젝트보다 점수가 지나치게 뛰면 감지 |
| 3 | **Must-Pass 방화벽** | 필수 기준은 다른 영역 점수로 메울 수 없음 |
| 4 | **독립 재평가** | 다섯 번마다 독립 재평가 (편차가 0.10을 넘으면 재조정) |
| 5 | **안티패턴 교차 검사** | 알려진 안티패턴이 나오면 해당 차원 점수를 0.50으로 제한 |

## Evaluator Memory Scope

평가자의 판단 기억은 **반복 하나짜리**입니다. GAN Loop가 한 바퀴 돌 때마다 sync-auditor는 새 컨텍스트로 다시 시작하고, 직전 반복의 판단 근거는 새 프롬프트에 실리지 않습니다. 반복 사이에 넘어가는 것은 Sprint Contract 상태뿐입니다. 평가자가 제 이전 판단에 붙들려 점수를 관성적으로 매기지 않게 하려는 설계입니다.

## 설정

`.moai/config/sections/harness.yaml`에서 설정합니다.

```yaml
harness:
  default_profile: "default"        # evaluator_profile 미지정 SPEC의 기본값
  evaluator:
    memory_scope: per_iteration     # FROZEN — 변경 불가
  mode_defaults:
    solo: auto                      # sub-agent 모드: 자동 감지
    team: auto                      # team 모드: 자동 감지
    cg: thorough                    # CG 모드: 항상 thorough
  auto_detection:
    enabled: true
    rules:
      minimal:
        conditions:
          - "file_count <= 3 AND single_domain"
      thorough:
        conditions:
          - "security_keywords OR payment_keywords present"
  escalation:
    enabled: true
    max_escalations: 2
  effort_mapping:
    minimal:  "low"
    standard: "medium"
    thorough: "high"
  levels:
    thorough:
      evaluator: true
```

## 관련 문서

- [하네스 엔지니어링](/ko/core-concepts/harness-engineering) — 하네스 개념 개요
- [TRUST 5 품질](/ko/core-concepts/trust-5) — 5가지 품질 기준
- [Constitution 시스템](/ko/core-concepts/constitution) — FROZEN/Evolvable 규칙
- GAN Loop — 디자인 품질 검증 반복 (GAN Loop는 adversarial 평가자-판별자 루프로 품질 개선을 위한 반복 검증 패턴입니다)
