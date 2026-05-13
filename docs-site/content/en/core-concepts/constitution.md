---
title: Constitution 시스템
weight: 35
draft: false
---

MoAI-ADK의 불변 규칙(FROZEN)과 진화 가능한 규칙(Evolvable)을 관리하는 헌법적 제약 시스템입니다.

## 개요

MoAI-ADK는 **Constitution(헌법)** 시스템을 통해 AI 에이전트가 임의로 변경할 수 없는
불변 제약(FROZEN Zone)과 학습을 통해 개선할 수 있는 진화 가능 제약(Evolvable Zone)을
구분합니다. 이는 하네스 엔지니어링의 핵심 안전 메커니즘입니다.

## FROZEN vs Evolvable

### FROZEN Zone (불변)

AI 에이전트가 절대 수정할 수 없는 규칙입니다. 인간 개발자만 변경할 수 있습니다.

**대표 항목**:

| 항목 | 설명 | 소스 |
|------|------|------|
| TRUST 5 | 5가지 품질 기준 | moai-constitution.md |
| SPEC + EARS | 명세서 형식 | spec-workflow.md |
| AskUserQuestion 독점 | 사용자 질문 채널 | agent-common-protocol.md |
| 평가 차원 4개 | Functionality/Security/Craft/Consistency | harness/scorer.go |
| 루브릭 앵커 4단계 | 0.25/0.50/0.75/1.00 | harness/rubric.go |
| 통과 임계값 하한 | 최소 0.60 (낮출 수 없음) | design-constitution.md |
| 디자인 파이프라인 순서 | manager-spec 먼저, evaluator-active 마지막 | design-constitution.md |

### Evolvable Zone (진화 가능)

학습(lessons)과 연구(research)를 통해 개선 제안이 가능한 규칙입니다.

**대표 항목**:

| 항목 | 설명 |
|------|------|
| 스킬 본문 내용 | moai-domain-* 스킬의 세부 내용 |
| 파이프라인 가중치 | design.yaml의 phase_weights |
| 반복 한계 | design.yaml의 iteration_limits |
| 에이전트 행동 규칙 | Surface Assumptions, Enforce Simplicity 등 |

## Zone Registry

모든 HARD 조항을 열거하는 **단일 진실 공급원**(Single Source of Truth)입니다.

### ID 할당 규칙

```
CONST-V3R2-NNN (3자리 이상 zero-padding)

001-050: 기존 HARD 조항
051-099: design constitution 미러 엔트리
100-149: design overflow (자동 확장)
150+: 신규 추가
```

### Canary Gate

FROZEN 조항은 `canary_gate: true`를 가집니다. 변경 전 canary 검증이 필수입니다.

```yaml
# Zone Registry 엔트리 예시
- id: CONST-V3R2-154
  zone: Frozen
  file: internal/harness/scorer.go
  anchor: "#dimension-enum"
  clause: "Dimension enum FROZEN at 4 values"
  canary_gate: true
```

## 안전 아키텍처 (5계층)

Constitution 시스템은 5계층 안전 아키텍처로 보호됩니다:

### Layer 1: Frozen Guard

쓰기 작업 전 대상 파일이 FROZEN zone이 아닌지 확인합니다. 위반 시 쓰기 차단 + 로깅 +
사용자 알림.

### Layer 2: Canary Check

제안된 변경을 메모리에 적용하고 최근 3개 프로젝트를 재평가합니다. 점수 하락이
0.10 초과하면 변경을 거부합니다.

### Layer 3: Contradiction Detector

새 학습이 기존 규칙과 충돌하면 양쪽을 모두 사용자에게 제시합니다. 자동 덮어쓰기는
절대 발생하지 않습니다.

### Layer 4: Rate Limiter

진화 속도를 제한합니다:

| 파라미터 | 기본값 | 설명 |
|-----------|--------|------|
| `max_evolution_rate_per_week` | 3 | 주간 최대 진화 횟수 |
| `cooldown_hours` | 24 | 진화 간 최소 대기 시간 |
| `max_active_learnings` | 50 | 활성 학습 항목 최대 수 |

### Layer 5: Human Oversight

`require_approval: true`인 경우 모든 진화 제안은 사용자 승인이 필요합니다.

## CLI에서 활용

```bash
# 전체 registry 조회
moai constitution list

# Frozen zone 필터
moai constitution list --zone frozen

# 특정 파일 조항만 조회
moai constitution list --file internal/harness/scorer.go

# JSON 형식 출력
moai constitution list --format json
```

## 관련 문서

- [TRUST 5 품질](/ko/core-concepts/trust-5) — 5가지 품질 기준
- [하네스 엔지니어링](/ko/core-concepts/harness-engineering) — 하네스 개념 개요
- [SPEC 기반 개발](/ko/core-concepts/spec-based-dev) — SPEC 워크플로우
