---
title: 의사결정 메모리 시스템
weight: 50
draft: false
---

MoAI의 사용자 선호도 학습 및 적응형 권장 시스템을 안내합니다.

{{< callout type="info" >}}
**한 줄 요약**: 의사결정 메모리는 사용자의 선택을 기억하고 향후 유사한 상황에서 개인화된 권장을 제공합니다.
{{< /callout >}}

## 시스템 개요

의사결정 메모리(Decision Memory)는 MoAI-ADK의 **장기 학습 계층**입니다. AskUserQuestion 라운드에서 사용자의 선택을 관찰하고, 향후 동일한 의사결정 지점에서 통계적으로 다수 선택을 기반으로 적응형 권장을 제공합니다.

### 핵심 원칙

| 원칙 | 설명 |
|------|------|
| **관찰 기반** | 사용자 선택의 통계적 다수를 학습 (정책 기본값 아님) |
| **투명성** | 권장 근거를 항상 명시 (cold-start 상태 포함) |
| **자율성** | 사용자는 권장을 언제든 거부 가능 |
| **적응형 강도** | 숙련도에 따라 권장의 강도 자동 조정 |

## 5 구성 요소

### 1. 3-Tier Memory Layer (메모리 계층)

의사결정 메모리는 3개 계층으로 구성됩니다.

#### L0: Immediate (즉시 메모리)
- **범위**: 현재 세션 내
- **용도**: 방금 사용자가 선택한 옵션 참조
- **지속성**: 세션 종료 시 소실

#### L1: Session Span (세션 범위 메모리)
- **범위**: 같은 프로젝트의 최근 3개 세션
- **용도**: 최근 선호도 기반 권장
- **지속성**: `.claude/projects/{hash}/memory/` 자동 메모리

#### L2: Long-term (장기 메모리)
- **범위**: 모든 세션 (무제한)
- **용도**: 통계적 다수 학습, 장기 트렌드
- **지속성**: MEMORY.md + topic 파일 (사용자 관리)

### 2. Adaptive Recommendation Placement (적응형 권장 배치)

권장(첫 옵션의 `(권장)` 라벨)은 **관찰된 통계적 다수**에 근거합니다.

#### Cold-Start (초기 상태)
- **관찰 < N**: 충분한 관찰 데이터 부재
- **권장 배치**: 정적 기본값 (명시적으로 공개)
- **표시 방식**: `based on static default, N observations needed for personalization`

#### Warm State (학습 중)
- **관찰 = N~M**: 부분 학습
- **권장 배치**: 관찰된 다수 + 신뢰도 신호
- **신뢰도**: 관찰 수 × 선택 일관성

#### Mature State (안정화)
- **관찰 > M**: 충분한 학습
- **권장 배치**: 강한 다수 확신 (통계적으로 유의)
- **신뢰도**: 최고 (≥95% 신뢰도)

#### 숙련도 기반 적응형 강도
- **전문가 (세션 > 50)**: 약 추천 강도 (자율성 우선, inferred preference만 공개)
- **초보자 (세션 < 10)**: 강 추천 강도 (`(권장)` 라벨 + 이유 명시)
- **중급자 (10 ≤ 세션 ≤ 50)**: 중간 강도 (정황에 따라 조정)

### 3. PostToolUse Capture Hook (의사결정 포착)

AskUserQuestion 응답이 도착하면 PostToolUse 훅이 자동으로 의사결정을 포착합니다.

#### 포착되는 데이터

```json
{
  "decision_id": "moai-ask-001",
  "timestamp": "2026-07-01T10:00:00Z",
  "question": "다음 단계를 선택하세요",
  "user_choice": "Option A (권장)",
  "all_options": ["Option A", "Option B", "Option C"],
  "context": {
    "spec_id": "SPEC-XXX-001",
    "phase": "run",
    "workflow": "/moai run"
  }
}
```

#### 저장 위치

- **세션 중**: `.moai/state/decisions/` (임시 JSON)
- **세션 종료**: `~/.claude/projects/{hash}/memory/decisions.jsonl` (자동 메모리)

### 4. Decay Policy (감쇠 정책)

오래된 의사결정의 가중치를 점진적으로 감소시킵니다.

#### 감쇠 함수

```
weight(t) = initial_weight × exp(-decay_rate × days_ago)
```

#### 기본값
- **Initial weight**: 1.0
- **Decay rate**: 0.1 (7일마다 약 50% 감쇠)
- **Retention period**: 90일 (이후 자동 아카이빙)

#### 예시

```
어제 선택: weight = 0.95
7일 전 선택: weight = 0.50
30일 전 선택: weight = 0.04
90일 이상: 아카이브 (권장 반영 제외)
```

### 5. Recovery Controls (복구 제어)

의사결정 메모리의 오류 복구 및 재설정을 관리합니다.

#### 메모리 초기화

사용자가 학습된 선호도를 초기화할 수 있습니다:

```bash
/moai memory reset
```

#### 선호도 편집

특정 의사결정 카테고리의 권장을 수정:

```bash
/moai memory set <category> <preferred-option>
```

#### 선호도 조회

현재 학습된 선호도 확인:

```bash
/moai memory list
```

## 의사결정 카테고리

메모리가 추적하는 주요 의사결정 유형:

| 카테고리 | 예시 |
|----------|------|
| **Tier Selection** | Tier S/M/L 선택 |
| **Cycle Type** | DDD vs TDD 모드 |
| **Worktree Strategy** | Main vs Branch vs Worktree |
| **PR Routing** | Direct-to-main vs PR-based |
| **Team Mode** | Solo vs Agent Teams |
| **Model Selection** | Model choice per task |
| **Effort Level** | Effort 레벨 (low/medium/high/xhigh) |

## 통계적 다수 학습의 예시

### 시나리오 1: Tier Selection

사용자가 10회의 Tier 선택을 했다면:

```
Tier S: 3회 선택
Tier M: 6회 선택  ← 통계적 다수 (60%)
Tier L: 1회 선택

학습 결과: Tier M이 (권장)으로 표시
신뢰도: 중상 (6/10 = 60%, N=10)
권장 문구: "Tier M (권장) — 최근 선택 60% 기반"
```

### 시나리오 2: Cycle Type

```
DDD: 4회
TDD: 5회 선택  ← 통계적 다수
기타: 1회

학습 결과: TDD가 (권장)
신뢰도: 중 (5/10 = 50%, N=10)
권장 문구: "TDD (권장) — 관찰 기반"
```

## Cold-Start 투명성

관찰 부족 시 명시적 공개:

```
선택지 1: Tier M (권장) — based on static default, 5 observations needed for personalization
선택지 2: Tier L
선택지 3: Tier S
```

사용자는 아직 학습 중인 상태임을 명확히 인식합니다.

## 숙련도 기반 강도 조정의 예

### 초보 사용자 (세션 < 10)
```
Tier M (권장) — 최근 선택 기반 제시
(강 추천 강도)
```

### 전문가 사용자 (세션 > 50)
```
선택지들:
- Tier M (최근 선택 60%)
- Tier L
- Tier S
(약 추천 강도, inferred preference 공개만)
```

## 관련 문서

- [AskUserQuestion 프로토콜](/advanced/agent-guide) - 권장 배치 규칙 (HARD)
- [워크플로우 선택](/advanced/harness-v4-builder) - Tier 선택 및 의사결정
- [메모리 시스템](/getting-started/memory) - 사용자 선호도 관리

{{< callout type="info" >}}
**팁**: 의사결정 메모리는 자동으로 작동합니다. 명시적 설정이 필요 없습니다. 사용자는 의사결정할 때마다 자동으로 학습됩니다.
{{< /callout >}}
