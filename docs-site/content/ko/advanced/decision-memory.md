---
title: 의사결정 메모리 시스템
weight: 50
draft: false
---

에이전틱 루프 엔지니어링의 출발점은 관찰입니다 — 루프가 돌 때마다 관찰이 쌓이고, 쌓인 관찰이 학습의 원료가 됩니다. 의사결정 메모리는 그 관찰 대상을 코드가 아니라 **사용자의 선택**으로 확장한 계층입니다.

{{< callout type="info" >}}
**한 줄 요약**: 의사결정 메모리는 사용자의 선택을 기억하고, 향후 유사한 상황에서 개인화된 권장을 제공합니다.
{{< /callout >}}

## 시스템 개요

의사결정 메모리(Decision Memory)는 MoAI-ADK의 **장기 학습 계층**입니다. AskUserQuestion 라운드에서 사용자의 선택을 관찰하고, 향후 동일한 의사결정 지점에서 통계적 다수 선택을 기반으로 적응형 권장을 제공합니다.

중요한 건 방향입니다. 시스템이 밀고 싶은 기본값을 `(권장)`으로 포장하는 게 아니라, **사용자가 실제로 반복 선택해 온 것**이 권장이 됩니다.

### 핵심 원칙

| 원칙 | 설명 |
|------|------|
| **관찰 기반** | 사용자 선택의 통계적 다수를 학습 (정책 기본값 아님) |
| **투명성** | 권장 근거를 항상 명시 (cold-start 상태 포함) |
| **자율성** | 사용자는 권장을 언제든 거부 가능 |
| **적응형 강도** | 숙련도에 따라 권장의 강도 자동 조정 |

## 4 구성 요소

### 1. 3-Tier Memory Layer (메모리 계층)

의사결정 메모리는 3개 계층으로 구성됩니다. 아래로 갈수록 오래 남습니다.

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

권장 배치는 5원칙으로 구성됩니다 (SSOT: `.claude/rules/moai/core/askuser-protocol.md` § Recommendation Placement Principles).

#### 원칙 1 — 방출 시점 (정보 이익 정렬)

오케스트레이터가 다가오는 의사결정의 불확실성 p를 추정할 때, p ≈ 0.5 (Fisher information I = p(1−p)가 최대인 결정 경계)에서 AskUserQuestion으로 해당 질문을 방출합니다. p가 0 또는 1에 가까울 때 (거의 확실)는 통계적 다수 옵션으로 자동 해결하고 질문을 생략합니다.

#### 원칙 2 — 질문 순서 (정보 이익 내림차순)

하나의 AskUserQuestion 호출에 여러 질문을 배치할 때, 정보 이익이 가장 높은 질문을 먼저 배치합니다. 사용자가 핵심 결정을 먼저 완료하고 낮은 가치의 질문은 나중에 만나도록 합니다.

#### 원칙 3 — 권장 옵션 (통계적 다수 합리적 기본값)

권장(첫 옵션의 `(권장)` 라벨)은 **관찰된 통계적 다수**에 근거합니다. 시스템이 밀고 싶은 정책 기본값이 아니라 사용자가 실제로 반복 선택한 옵션이 권장이 됩니다. 관찰량에 따라 세 상태를 오갑니다.

##### Cold-Start (초기 상태)
- **관찰 < N**: 충분한 관찰 데이터 부재
- **권장 배치**: 정적 기본값 (명시적으로 공개)
- **표시 방식**: `based on static default, N observations needed for personalization`

##### Warm State (학습 중)
- **관찰 = N~M**: 부분 학습
- **권장 배치**: 관찰된 다수 + 신뢰도 신호
- **신뢰도**: 관찰 수 × 선택 일관성

##### Mature State (안정화)
- **관찰 > M**: 충분한 학습
- **권장 배치**: 강한 다수 확신 (통계적으로 유의)
- **신뢰도**: 최고 (≥95% 신뢰도)

#### 원칙 4 — 전제 조건 명시

권장 옵션의 설명은 해당 권장이 성립하는 전제 조건을 명시해야 합니다. 사용자가 전제가 위배되었을 때 즉시 거부할 수 있도록 `"Recommended when <전제조건>"` 형태로 제시합니다. 전제가 명시되지 않은 권장은 설계 결함입니다.

#### 원칙 5 — 숙련도 기반 적응형 강도

같은 권장이라도 상대에 따라 강도가 달라집니다. 전문가에게 강한 권장은 자율성을 침해하고, 초보자에게 약한 권장은 결정 피로만 늘리기 때문입니다.

- **전문가 (세션 > 50)**: 약 추천 강도 (자율성 우선, inferred preference만 공개)
- **초보자 (세션 < 10)**: 강 추천 강도 (`(권장)` 라벨 + 이유 명시)
- **중급자 (10 ≤ 세션 ≤ 50)**: 중간 강도 (정황에 따라 조정)

### 3. PostToolUse Capture Hook (의사결정 포착)

AskUserQuestion 응답이 도착하면 PostToolUse 훅이 자동으로 의사결정을 포착합니다. 사용자가 따로 기록할 일은 없습니다.

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

3개월 전의 선택이 오늘의 선호를 대변하지는 않습니다. 오래된 의사결정의 가중치는 점진적으로 감소합니다.

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

## 의사결정 카테고리

메모리가 추적하는 주요 의사결정 유형입니다.

| 카테고리 | 예시 |
|----------|------|
| **Tier Selection** | Tier S/M/L 선택 |
| **Cycle Type** | DDD vs TDD 모드 |
| **Worktree Strategy** | Main vs Branch vs Worktree |
| **PR Routing** | Direct-to-main vs PR-based |
| **Model Selection** | Model choice per task |
| **Effort Level** | Effort 레벨 (low/medium/high/xhigh) |

Model Selection과 Effort Level이 여기에 포함된다는 점에 주목할 만합니다 — 의사결정 메모리가 학습한 선호가 결국 모델·추론 깊이 배정으로 이어지므로, 이 시스템은 토크노믹스의 개인화 계층이기도 합니다.

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

관찰이 부족할 때는 그 사실을 숨기지 않고 명시적으로 공개합니다.

```
선택지 1: Tier M (권장) — based on static default, 5 observations needed for personalization
선택지 2: Tier L
선택지 3: Tier S
```

사용자는 아직 학습 중인 상태임을 명확히 인식할 수 있습니다.

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

- [에이전트 가이드](/ko/advanced/agent-guide) - AskUserQuestion 권장 배치 규칙 (HARD)
- [Harness v4 Builder 심화 가이드](/ko/advanced/harness-v4-builder) - Tier 선택 및 의사결정
- [메모리 시스템](/ko/claude-code/context-memory/memory) - 사용자 선호도 관리

{{< callout type="info" >}}
**팁**: 의사결정 메모리는 자동으로 작동합니다. 명시적 설정이 필요 없습니다 — 의사결정을 내릴 때마다 시스템이 조용히 학습합니다.
{{< /callout >}}
