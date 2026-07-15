---
title: Decision Memory System
weight: 50
draft: false
---

The starting point of agentic loop engineering is observation — every time the loop runs, observations accumulate, and the accumulated observations become the raw material of learning. Decision memory is the layer that extends the observation target from code to the **user's choices**.

{{< callout type="info" >}}
**One-line summary**: Decision memory remembers the user's choices and provides personalized recommendations in similar future situations.
{{< /callout >}}

## System Overview

Decision Memory is MoAI-ADK's **long-term learning layer**. It observes the user's choices in AskUserQuestion rounds and provides an adaptive recommendation based on the statistical-majority choice at the same decision point in the future.

What matters is the direction. Rather than packaging a default the system wants to push as `(Recommended)`, **what the user has actually chosen repeatedly** becomes the recommendation.

### Core Principles

| Principle | Description |
|------|------|
| **Observation-based** | Learns the statistical majority of user choices (not a policy default) |
| **Transparency** | Always states the recommendation basis (including the cold-start state) |
| **Autonomy** | The user can reject a recommendation at any time |
| **Adaptive strength** | Automatically adjusts recommendation strength by proficiency |

## The 4 Components

### 1. 3-Tier Memory Layer

Decision memory consists of 3 layers. The further down, the longer they persist.

#### L0: Immediate

- **Scope**: within the current session
- **Purpose**: reference the option the user just chose
- **Persistence**: lost when the session ends

#### L1: Session Span

- **Scope**: the last 3 sessions of the same project
- **Purpose**: recommendation based on recent preference
- **Persistence**: `.claude/projects/{hash}/memory/` auto-memory

#### L2: Long-term

- **Scope**: all sessions (unlimited)
- **Purpose**: statistical-majority learning, long-term trends
- **Persistence**: MEMORY.md + topic files (user-managed)

### 2. Adaptive Recommendation Placement

Recommendation placement consists of 5 principles (SSOT: `.claude/rules/moai/core/askuser-protocol.md` § Recommendation Placement Principles).

#### Principle 1 — emission timing (information-gain alignment)

When the orchestrator estimates the uncertainty p of an upcoming decision, it emits that question via AskUserQuestion at p ≈ 0.5 (the decision boundary where Fisher information I = p(1−p) is maximal). When p is close to 0 or 1 (nearly certain), it auto-resolves to the statistical-majority option and omits the question.

#### Principle 2 — question ordering (descending information gain)

When placing multiple questions in one AskUserQuestion call, the highest-information-gain question is placed first. This lets the user complete the core decisions first and meet lower-value questions later.

#### Principle 3 — recommended option (statistical-majority rational default)

The recommendation (the first option's `(Recommended)` label) is grounded in the **observed statistical majority**. Rather than a policy default the system wants to push, the option the user has actually chosen repeatedly becomes the recommendation. It moves between three states depending on the observation count.

##### Cold-Start (initial state)

- **Observations < N**: insufficient observation data
- **Recommendation placement**: the static default (explicitly disclosed)
- **Display form**: `based on static default, N observations needed for personalization`

##### Warm State (learning)

- **Observations = N~M**: partial learning
- **Recommendation placement**: the observed majority + a confidence signal
- **Confidence**: observation count × selection consistency

##### Mature State (stabilized)

- **Observations > M**: sufficient learning
- **Recommendation placement**: strong majority conviction (statistically significant)
- **Confidence**: highest (≥95% confidence)

#### Principle 4 — precondition statement

A recommended option's description must state the preconditions under which the recommendation holds. It is presented in the `"Recommended when <precondition>"` form so the user can immediately reject it when a precondition is violated. A recommendation without stated preconditions is a design defect.

#### Principle 5 — proficiency-based adaptive strength

The same recommendation has different strength depending on the recipient. Strong recommendations erode an expert's autonomy, while weak recommendations to a beginner only add decision fatigue.

- **Expert (sessions > 50)**: weak recommendation strength (autonomy-first, disclose only the inferred preference)
- **Beginner (sessions < 10)**: strong recommendation strength (`(Recommended)` label + stated rationale)
- **Intermediate (10 ≤ sessions ≤ 50)**: medium strength (adjusted by context)

### 3. PostToolUse Capture Hook

When an AskUserQuestion response arrives, the PostToolUse hook automatically captures the decision. There is nothing for the user to record separately.

#### Captured Data

```json
{
  "decision_id": "moai-ask-001",
  "timestamp": "2026-07-01T10:00:00Z",
  "question": "Choose the next step",
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

Preference memory persists the decisions captured on the orchestrator's user-question channel to `~/.claude/projects/{slug}/memory/user_decisions/` (SPEC-V3R6-ASKUSER-DECISION-MEMORY-001).

**3-tier hierarchy:**

- **core**: highest-priority hot cache. On a core hit, recall/archival are not accessed
- **recall** (`recall.jsonl`): recent-session facts
- **archival** (`archival/`): the full-search target

### 4. Decay Policy

Past choices do not represent today's preference verbatim. Transient entries are subject to **power-law decay + a 28-day TTL** (REQ-ADM-011, REQ-ADM-012), so their weight gradually decreases and they are evicted once the TTL elapses. Core-tier eviction demotes the lowest-weight items first.

Management commands:

```bash
moai preference decay-scan   # one background decay pass (at most once per day, timestamp-gated)
moai preference toggle       # per-session personalization on/off (non-persistent, reset each session)
```

{{< callout type="info" >}}
**Note**: the exact decay exponent, half-life constant, and other detailed parameters are runtime implementation details (only the power-law + 28-day TTL contract is fixed); this page describes them only at the contract level.
{{< /callout >}}

## Decision Categories

The main decision types the memory tracks.

| Category | Example |
|----------|------|
| **Tier Selection** | Tier S/M/L selection |
| **Cycle Type** | DDD vs TDD mode |
| **Worktree Strategy** | Main vs Branch vs Worktree |
| **PR Routing** | Direct-to-main vs PR-based |
| **Model Selection** | Model choice per task |
| **Effort Level** | Effort level (low/medium/high/xhigh) |

It is worth noting that Model Selection and Effort Level are included here — since the preferences that decision memory learns ultimately feed into model and reasoning-depth assignment, this system is also the personalization layer of Tokenomics.

## Examples of Statistical-Majority Learning

### Scenario 1: Tier Selection

If the user has made 10 Tier selections:

```
Tier S: chosen 3 times
Tier M: chosen 6 times  ← statistical majority (60%)
Tier L: chosen 1 time

Learning result: Tier M shown as (Recommended)
Confidence: medium-high (6/10 = 60%, N=10)
Recommendation text: "Tier M (Recommended) — based on 60% of recent choices"
```

### Scenario 2: Cycle Type

```
DDD: 4 times
TDD: chosen 5 times  ← statistical majority
Other: 1 time

Learning result: TDD is (Recommended)
Confidence: medium (5/10 = 50%, N=10)
Recommendation text: "TDD (Recommended) — observation-based"
```

## Cold-Start Transparency

When observations are insufficient, that fact is disclosed explicitly rather than hidden.

```
Option 1: Tier M (Recommended) — based on static default, 5 observations needed for personalization
Option 2: Tier L
Option 3: Tier S
```

The user can clearly recognize that the system is still learning.

## Examples of Proficiency-Based Strength Adjustment

### Beginner User (sessions < 10)

```
Tier M (Recommended) — presented based on recent choices
(strong recommendation strength)
```

### Expert User (sessions > 50)

```
Options:
- Tier M (60% of recent choices)
- Tier L
- Tier S
(weak recommendation strength, disclose the inferred preference only)
```

## Related Documents

- [Agent Guide](/en/advanced/agent-guide) - AskUserQuestion recommendation-placement rules (HARD)
- [Harness v4 Builder Deep Dive](/en/advanced/harness-v4-builder) - Tier selection and decision-making
- [Memory System](/en/claude-code/context-memory/memory) - user-preference management

{{< callout type="info" >}}
**Tip**: Decision memory works automatically. No explicit configuration is needed — every time you make a decision, the system quietly learns.
{{< /callout >}}
