---
title: Decision Memory System
weight: 50
draft: false
---

Agentic loop engineering begins with observation — every turn of the loop accumulates observations, and those observations become the raw material of learning. Decision memory is the layer that extends the object of observation from code to **the user's choices**.

{{< callout type="info" >}}
**One-line summary**: Decision memory remembers the user's choices and provides personalized recommendations in similar situations in the future.
{{< /callout >}}

## System Overview

Decision Memory is MoAI-ADK's **long-term learning layer**. It observes user choices in AskUserQuestion rounds and, at the same decision points in the future, provides adaptive recommendations based on the statistical majority of those choices.

Direction is what matters. The system does not wrap the default it wants to push in a `(Recommended)` label — what **the user has actually chosen repeatedly** becomes the recommendation.

### Core Principles

| Principle | Description |
|------|------|
| **Observation-based** | Learns the statistical majority of user choices (not policy defaults) |
| **Transparency** | The rationale for a recommendation is always stated (including cold-start status) |
| **Autonomy** | The user can reject a recommendation at any time |
| **Adaptive strength** | Recommendation strength is auto-adjusted by proficiency |

## The 5 Components

### 1. 3-Tier Memory Layer

Decision memory consists of three tiers. The lower the tier, the longer it persists.

#### L0: Immediate
- **Scope**: within the current session
- **Purpose**: referencing options the user just selected
- **Persistence**: lost when the session ends

#### L1: Session Span
- **Scope**: the last 3 sessions of the same project
- **Purpose**: recommendations based on recent preferences
- **Persistence**: auto-memory in `.claude/projects/{hash}/memory/`

#### L2: Long-term
- **Scope**: all sessions (unlimited)
- **Purpose**: statistical-majority learning, long-term trends
- **Persistence**: MEMORY.md + topic files (user-managed)

### 2. Adaptive Recommendation Placement

The recommendation (the `(Recommended)` label on the first option) is grounded in the **observed statistical majority**. It moves between three states depending on the amount of observation.

#### Cold-Start
- **Observations < N**: insufficient observation data
- **Recommendation placement**: static default (explicitly disclosed)
- **Display form**: `based on static default, N observations needed for personalization`

#### Warm State
- **Observations = N~M**: partial learning
- **Recommendation placement**: observed majority + confidence signal
- **Confidence**: observation count × choice consistency

#### Mature State
- **Observations > M**: sufficient learning
- **Recommendation placement**: strong majority conviction (statistically significant)
- **Confidence**: highest (≥95% confidence)

#### Proficiency-Based Adaptive Strength

The same recommendation is delivered with different strength depending on the audience. A strong recommendation to an expert erodes autonomy, and a weak recommendation to a beginner only adds decision fatigue.

- **Expert (sessions > 50)**: weak recommendation strength (autonomy first, only the inferred preference is disclosed)
- **Beginner (sessions < 10)**: strong recommendation strength (`(Recommended)` label + stated rationale)
- **Intermediate (10 ≤ sessions ≤ 50)**: medium strength (adjusted by context)

### 3. PostToolUse Capture Hook

When an AskUserQuestion response arrives, the PostToolUse hook automatically captures the decision. The user never has to record anything manually.

#### Captured Data

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

#### Storage Locations

- **During the session**: `.moai/state/decisions/` (temporary JSON)
- **At session end**: `~/.claude/projects/{hash}/memory/decisions.jsonl` (auto-memory)

### 4. Decay Policy

A choice made three months ago does not represent today's preference. The weight of older decisions gradually decreases.

#### Decay Function

```
weight(t) = initial_weight × exp(-decay_rate × days_ago)
```

#### Defaults
- **Initial weight**: 1.0
- **Decay rate**: 0.1 (about 50% decay every 7 days)
- **Retention period**: 90 days (auto-archived afterward)

#### Example

```
어제 선택: weight = 0.95
7일 전 선택: weight = 0.50
30일 전 선택: weight = 0.04
90일 이상: 아카이브 (권장 반영 제외)
```

### 5. Recovery Controls

For when learning has hardened in the wrong direction, error-recovery and reset tools are provided.

#### Memory Reset

The user can reset learned preferences.

```bash
/moai memory reset
```

#### Preference Editing

Modify the recommendation for a specific decision category.

```bash
/moai memory set <category> <preferred-option>
```

#### Preference Inspection

Check the currently learned preferences.

```bash
/moai memory list
```

## Decision Categories

The main decision types the memory tracks.

| Category | Example |
|----------|------|
| **Tier Selection** | Choosing Tier S/M/L |
| **Cycle Type** | DDD vs TDD mode |
| **Worktree Strategy** | Main vs Branch vs Worktree |
| **PR Routing** | Direct-to-main vs PR-based |
| **Team Mode** | Solo vs Agent Teams |
| **Model Selection** | Model choice per task |
| **Effort Level** | Effort level (low/medium/high/xhigh) |

It is worth noting that Model Selection and Effort Level are included here — the preferences decision memory learns ultimately feed into model and reasoning-depth assignment, so this system is also the personalization layer of tokenomics.

## Examples of Statistical-Majority Learning

### Scenario 1: Tier Selection

If the user has made 10 Tier selections:

```
Tier S: 3회 선택
Tier M: 6회 선택  ← 통계적 다수 (60%)
Tier L: 1회 선택

학습 결과: Tier M이 (권장)으로 표시
신뢰도: 중상 (6/10 = 60%, N=10)
권장 문구: "Tier M (권장) — 최근 선택 60% 기반"
```

### Scenario 2: Cycle Type

```
DDD: 4회
TDD: 5회 선택  ← 통계적 다수
기타: 1회

학습 결과: TDD가 (권장)
신뢰도: 중 (5/10 = 50%, N=10)
권장 문구: "TDD (권장) — 관찰 기반"
```

## Cold-Start Transparency

When observations are insufficient, the fact is disclosed explicitly rather than hidden.

```
선택지 1: Tier M (권장) — based on static default, 5 observations needed for personalization
선택지 2: Tier L
선택지 3: Tier S
```

The user can clearly recognize that the system is still learning.

## Examples of Proficiency-Based Strength Adjustment

### Beginner User (sessions < 10)
```
Tier M (권장) — 최근 선택 기반 제시
(강 추천 강도)
```

### Expert User (sessions > 50)
```
선택지들:
- Tier M (최근 선택 60%)
- Tier L
- Tier S
(약 추천 강도, inferred preference 공개만)
```

## Related Documents

- [Agent Guide](/en/advanced/agent-guide) - AskUserQuestion recommendation placement rules (HARD)
- [Harness v4 Builder Advanced Guide](/en/advanced/harness-v4-builder) - tier selection and decision-making
- [Memory System](/en/getting-started/memory) - managing user preferences

{{< callout type="info" >}}
**Tip**: Decision memory works automatically. No explicit configuration is needed — the system learns quietly every time you make a decision.
{{< /callout >}}
