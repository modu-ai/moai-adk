---
title: Constitution System
weight: 35
draft: false
---

The constitutional constraint system that manages MoAI-ADK's immutable rules (FROZEN) and evolvable rules (Evolvable).

## Overview

As we saw in [Harness Engineering](/en/core-concepts/harness-engineering), MoAI-ADK's harness evolves its own guidance from the observations the loop accumulates. So what governs that evolution? The answer is the **Constitution** system.

The Constitution separates immutable constraints that AI agents can never change on their own (the FROZEN Zone) from evolvable constraints that can be improved through learning (the Evolvable Zone). Keeping the evaluation criteria and safety rules **outside** the evolution loop — this is why a self-evolving harness does not run away, and it is the core safety mechanism of harness engineering.

## FROZEN vs Evolvable

### FROZEN Zone (Immutable)

Rules that AI agents can never modify. Only human developers may change them.

**Representative items**:

| Item | Description | Source |
|------|------|------|
| TRUST 5 | The five quality criteria | moai-constitution.md |
| SPEC + EARS | The specification format | spec-workflow.md |
| AskUserQuestion monopoly | The user-question channel | agent-common-protocol.md |
| 4 evaluation dimensions | Functionality/Security/Craft/Consistency | harness/scorer.go |
| 4 rubric anchors | 0.25/0.50/0.75/1.00 | harness/rubric.go |
| Pass-threshold floor | Minimum 0.60 (cannot be lowered) | design-constitution.md |
| Design pipeline order | manager-spec first, sync-auditor last | design-constitution.md |

### Evolvable Zone (Evolvable)

Rules for which improvement proposals are possible through lessons and research.

**Representative items**:

| Item | Description |
|------|------|
| Skill body content | The details of moai-domain-* skills |
| Pipeline weights | phase_weights in design.yaml |
| Iteration limits | iteration_limits in design.yaml |
| Agent behavior rules | Surface Assumptions, Enforce Simplicity, etc. |

## Zone Registry

The **Single Source of Truth** enumerating every HARD clause.

### ID Allocation Rules

```
CONST-V3R2-NNN (zero-padded to 3+ digits)

001-050: existing HARD clauses
051-099: design constitution mirror entries
100-149: design overflow (auto expansion)
150+: new additions
```

### Canary Gate

FROZEN clauses carry `canary_gate: true`. Canary verification is mandatory before any change.

```yaml
# Example Zone Registry entry
- id: CONST-V3R2-154
  zone: Frozen
  file: internal/harness/scorer.go
  anchor: "#dimension-enum"
  clause: "Dimension enum FROZEN at 4 values"
  canary_gate: true
```

## Safety Architecture (5 Layers)

The Constitution system is protected by a 5-layer safety architecture. No matter how much learning the harness accumulates, a change must pass through the following five gates in order:

### Layer 1: Frozen Guard

Before any write operation, verifies the target file is not in the FROZEN zone. On violation: block the write + log +
notify the user.

### Layer 2: Canary Check

Applies the proposed change in memory and re-evaluates the 3 most recent projects. If the score drop
exceeds 0.10, the change is rejected.

### Layer 3: Contradiction Detector

When a new learning conflicts with an existing rule, both sides are presented to the user. Automatic overwriting
never happens.

### Layer 4: Rate Limiter

Limits the speed of evolution:

| Parameter | Default | Description |
|-----------|--------|------|
| `max_evolution_rate_per_week` | 3 | Maximum evolutions per week |
| `cooldown_hours` | 24 | Minimum wait between evolutions |
| `max_active_learnings` | 50 | Maximum number of active learning items |

### Layer 5: Human Oversight

When `require_approval: true`, every evolution proposal requires user approval.

## Using It from the CLI

```bash
# Query the full registry
moai constitution list

# Filter by Frozen zone
moai constitution list --zone frozen

# Query clauses for a specific file only
moai constitution list --file internal/harness/scorer.go

# Output in JSON format
moai constitution list --format json
```

## Related Documents

- [TRUST 5 Quality](/en/core-concepts/trust-5) — The five quality criteria
- [Harness Engineering](/en/core-concepts/harness-engineering) — Harness concept overview
- [SPEC-Based Development](/en/core-concepts/spec-based-dev) — The SPEC workflow
