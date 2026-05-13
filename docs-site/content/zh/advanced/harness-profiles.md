---
title: 하네스 프로필과 평가 시스템
weight: 75
draft: false
---

3계층 하네스 레벨과 4차원 평가 프로필을 통한 적응형 품질 검증 시스템입니다.

## 개요

MoAI-ADK의 하네스(Harness)는 **3계층 적응형 품질 검증 시스템**입니다. SPEC의 복잡도에
따라 자동으로 검증 깊이를 조절합니다. evaluator-active 에이전트가 4차원 스코어링으로
독립적이고 회의적인 품질 평가를 수행합니다.

## 3계층 하네스 레벨

| 레벨 | 설명 | 적용 시점 | evaluator-active |
|------|------|----------|-----------------|
| **minimal** | 빠른 검증 | 단순 변경 (typos, 설정 수정) | 생략 가능 |
| **standard** | 기본 품질 검증 | 대부분의 작업 | 선택적 |
| **thorough** | 전체 검증 + TRUST 5 | 복잡한 SPEC, 대규모 변경 | 필수 |

하네스 레벨은 SPEC scope를 기반으로 **복잡도 추정기**(Complexity Estimator)가 자동으로
결정합니다.

## 4차원 스코어링

evaluator-active는 4개 차원으로 점수를 매깁니다:

| 차원 | 설명 | 기본 Must-Pass |
|------|------|---------------|
| **Functionality** | 기능 완성도 — 의도된 목적을 달성했는가 | 예 |
| **Security** | 보안 — OWASP, 인증, 권한, 입력 검증 | 예 |
| **Craft** | 코드 품질 — 가독성, 구조, 테스트 커버리지 | 아니오 |
| **Consistency** | 일관성 — 프로젝트 규칙, 코드 스타일 준수 | 아니오 |

### 점수 범위

각 차원은 0.0 ~ 1.0 점수를 받습니다.

### 루브릭 앵커

모든 평가 기준은 4단계 루브릭 앵커를 가집니다:

| 점수 | 수준 | 의미 |
|------|------|------|
| 0.25 | 미달 | 기본 요구사항 미충족 |
| 0.50 | 부분 | 일부 충족, 개선 필요 |
| 0.75 | 충족 | 대부분 충족, 소규모 개선 |
| 1.00 | 우수 | 모든 기준 완벽 충족 |

## 평가 프로필

`.moai/config/evaluator-profiles/`에 4개 프로필이 제공됩니다:

| 프로필 | 설명 | 적합한 경우 |
|--------|------|------------|
| `default.md` | 균형 잡힌 기본 프로필 | 대부분의 작업 |
| `strict.md` | 엄격한 기준 | 보안 중요 작업 |
| `lenient.md` | 관대한 기준 | 프로토타이핑 |
| `frontend.md` | 프론트엔드 특화 | UI/UX 작업 |

## 평가자 편향 방지 (5가지 메커니즘)

평가자의 관대함을 방지하기 위해 5가지 메커니즘이 작동합니다:

| # | 메커니즘 | 설명 |
|---|---------|------|
| 1 | **루브릭 앵커링** | 점수에 루브릭 정당화 필수 |
| 2 | **회귀 베이스라인** | 이전 프로젝트 대비 과도한 점수 상승 감지 |
| 3 | **Must-Pass 방화벽** | 필수 기준은 다른 영역 점수로 보상 불가 |
| 4 | **독립 재평가** | 5번째마다 독립 재평가 (편차 > 0.10 시 재조정) |
| 5 | **안티패턴 교차 검사** | 알려진 안티패턴 발견 시 해당 차원 점수 0.50 제한 |

## Evaluator Memory Scope

평가자의 판단 기억은 **반복별로 일시적**입니다. GAN Loop의 각 반복에서 evaluator-active는
새 컨텍스트로 재시작되며, 이전 반복의 판단 근거는 새 프롬프트에 포함되지 않습니다.
Sprint Contract 상태만이 반복 간에 유지됩니다.

## 설정

`.moai/config/sections/harness.yaml`에서 설정합니다:

```yaml
harness:
  level: auto              # auto | minimal | standard | thorough
  evaluator:
    memory_scope: per_iteration   # FROZEN — 변경 불가
    profiles:
      default: .moai/config/evaluator-profiles/default.md
      strict: .moai/config/evaluator-profiles/strict.md
    aggregation: min              # min | mean
    must_pass_dimensions:
      - Functionality
      - Security
```

## 관련 문서

- [하네스 엔지니어링](/ko/core-concepts/harness-engineering) — 하네스 개념 개요
- [TRUST 5 품질](/ko/core-concepts/trust-5) — 5가지 품질 기준
- [Constitution 시스템](/ko/core-concepts/constitution) — FROZEN/Evolvable 규칙
- [GAN Loop](/ko/design/gan-loop) — 디자인 품질 검증 반복
