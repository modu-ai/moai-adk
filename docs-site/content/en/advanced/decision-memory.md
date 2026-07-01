---
title: Decision Memory System
weight: 50
draft: false
---

MoAI's user preference learning and adaptive recommendation system.

{{< callout type="info" >}}
**One-line summary**: Decision Memory remembers user choices and provides personalized recommendations at similar decision points in the future.
{{< /callout >}}

## System Overview

Decision Memory is the **long-term learning layer** of MoAI-ADK. It observes user choices during AskUserQuestion rounds and provides adaptive recommendations based on statistically majority choices at future equivalent decision points.

### Core Principles

| Principle | Description |
|-----------|-------------|
| **Observation-Based** | Learns from statistical majority of user choices (not policy defaults) |
| **Transparency** | Always states the basis for recommendations (including cold-start state) |
| **Autonomy** | Users can always reject recommendations |
| **Adaptive Strength** | Recommendation intensity automatically adjusts based on user expertise level |

## 5 Core Components

### 1. 3-Tier Memory Layer

Decision Memory consists of three layers.

#### L0: Immediate (Current Session)
- **Scope**: Within the current session only
- **Use**: Reference the option user just selected
- **Persistence**: Discarded at session end

#### L1: Session Span (Recent Sessions)
- **Scope**: Most recent 3 sessions in the same project
- **Use**: Recommendations based on recent preferences
- **Persistence**: Auto-memory in `.claude/projects/{hash}/memory/`

#### L2: Long-term (All Sessions)
- **Scope**: All sessions (unlimited history)
- **Use**: Statistical majority learning, long-term trends
- **Persistence**: MEMORY.md + topic files (user-managed)

### 2. Adaptive Recommendation Placement

Recommendations (the `(Recommended)` label on the first option) are grounded in **observed statistical majorities**.

#### Cold-Start (Initial State)
- **Observations < N**: Insufficient data collected
- **Recommendation basis**: Static default (explicitly disclosed)
- **Display format**: `based on static default, N observations needed for personalization`

#### Warm State (Learning Phase)
- **Observations = N~M**: Partial learning
- **Recommendation basis**: Observed majority + confidence signal
- **Confidence**: Observation count × choice consistency

#### Mature State (Stable)
- **Observations > M**: Sufficient learning
- **Recommendation basis**: Strong majority with high confidence
- **Confidence**: Highest (≥95% statistical significance)

#### Expertise-Based Adaptive Strength
- **Expert (sessions > 50)**: Weak recommendation strength (autonomy-first, inferred preference disclosed only)
- **Novice (sessions < 10)**: Strong recommendation strength (`(Recommended)` label + explicit reason)
- **Intermediate (10 ≤ sessions ≤ 50)**: Medium strength (context-dependent adjustment)

### 3. PostToolUse Capture Hook

Decision capture is automatic when AskUserQuestion responses arrive via PostToolUse hook.

#### Captured Data

```json
{
  "decision_id": "moai-ask-001",
  "timestamp": "2026-07-01T10:00:00Z",
  "question": "Select your next step",
  "user_choice": "Option A (Recommended)",
  "all_options": ["Option A", "Option B", "Option C"],
  "context": {
    "spec_id": "SPEC-XXX-001",
    "phase": "run",
    "workflow": "/moai run"
  }
}
```

#### Storage Locations

- **During session**: `.moai/state/decisions/` (temporary JSON)
- **Session end**: `~/.claude/projects/{hash}/memory/decisions.jsonl` (auto-memory)

### 4. Decay Policy

Gradually reduces the weight of older decisions.

#### Decay Function

```
weight(t) = initial_weight × exp(-decay_rate × days_ago)
```

#### Defaults
- **Initial weight**: 1.0
- **Decay rate**: 0.1 (approximately 50% decay per 7 days)
- **Retention period**: 90 days (then auto-archive)

#### Examples

```
Yesterday's choice:  weight = 0.95
7 days ago:          weight = 0.50
30 days ago:         weight = 0.04
90+ days:            Archived (excluded from recommendations)
```

### 5. Recovery Controls

Manages error recovery and resets for decision memory.

#### Memory Reset

Users can reset learned preferences:

```bash
/moai memory reset
```

#### Preference Edit

Modify recommendations for a specific decision category:

```bash
/moai memory set <category> <preferred-option>
```

#### Preference Query

View currently learned preferences:

```bash
/moai memory list
```

## Decision Categories

Primary decision types tracked by memory:

| Category | Example |
|----------|---------|
| **Tier Selection** | Choose Tier S/M/L |
| **Cycle Type** | DDD vs TDD mode |
| **Worktree Strategy** | Main vs Branch vs Worktree |
| **PR Routing** | Direct-to-main vs PR-based |
| **Team Mode** | Solo vs Agent Teams |
| **Model Selection** | Model choice per task |
| **Effort Level** | Effort level (low/medium/high/xhigh) |

## Statistical Majority Learning Examples

### Scenario 1: Tier Selection

After 10 Tier selection decisions:

```
Tier S: 3 selections
Tier M: 6 selections  ← Statistical majority (60%)
Tier L: 1 selection

Learning result: Tier M labeled (Recommended)
Confidence: Moderate-High (6/10 = 60%, N=10)
Recommendation text: "Tier M (Recommended) — based on 60% recent choices"
```

### Scenario 2: Cycle Type

```
DDD: 4 selections
TDD: 5 selections  ← Statistical majority
Other: 1 selection

Learning result: TDD labeled (Recommended)
Confidence: Moderate (5/10 = 50%, N=10)
Recommendation text: "TDD (Recommended) — based on observation"
```

## Cold-Start Transparency

When observations are insufficient, explicit disclosure:

```
Option 1: Tier M (Recommended) — based on static default, 5 observations needed for personalization
Option 2: Tier L
Option 3: Tier S
```

Users clearly understand the learning-in-progress state.

## Expertise-Based Strength Adjustment Examples

### Novice User (sessions < 10)
```
Tier M (Recommended) — based on recent choices
(Strong recommendation strength)
```

### Expert User (sessions > 50)
```
Options:
- Tier M (recent choice 60%)
- Tier L
- Tier S
(Weak recommendation strength, inferred preference disclosed only)
```

## Related Documentation

- [AskUserQuestion Protocol](/advanced/agent-guide) - Recommendation placement rules (HARD)
- [Workflow Selection](/advanced/harness-v4-builder) - Tier selection and decision-making
- [Memory System](/getting-started/memory) - User preference management

{{< callout type="info" >}}
**Tip**: Decision Memory operates automatically. No explicit configuration needed. Users are automatically learned as they make decisions.
{{< /callout >}}
