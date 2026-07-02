---
title: Constitution System
weight: 35
draft: false
---

The constitutional constraint system that manages MoAI-ADK's immutable rules (FROZEN) and rules that can evolve (Evolvable).

## Overview

Through its **Constitution** system, MoAI-ADK distinguishes between immutable
constraints (FROZEN Zone) that an AI agent can never change on its own, and
evolvable constraints (Evolvable Zone) that can be improved through learning.
This is the core safety mechanism of harness engineering.

## FROZEN vs Evolvable

### FROZEN Zone (immutable)

Rules that an AI agent can never modify. Only a human developer can change them.

**Representative entries**:

| Item | Description | Source |
|------|------|------|
| TRUST 5 | The 5 quality criteria | moai-constitution.md |
| SPEC + EARS | Specification format | spec-workflow.md |
| AskUserQuestion exclusivity | The sole user-question channel | agent-common-protocol.md |
| 4 evaluation dimensions | Functionality/Security/Craft/Consistency | harness/scorer.go |
| 4-tier rubric anchors | 0.25/0.50/0.75/1.00 | harness/rubric.go |
| Pass-threshold floor | Minimum 0.60 (cannot be lowered) | design-constitution.md |
| Design pipeline order | manager-spec first, sync-auditor last | design-constitution.md |

### Evolvable Zone (can evolve)

Rules that can be improved through lessons and research.

**Representative entries**:

| Item | Description |
|------|------|
| Skill body content | The detailed content of moai-domain-* skills |
| Pipeline weights | The phase_weights in design.yaml |
| Iteration limits | The iteration_limits in design.yaml |
| Agent behavior rules | Surface Assumptions, Enforce Simplicity, etc. |

## Zone Registry

The **Single Source of Truth** that enumerates every HARD clause.

### ID Assignment Rules

```
CONST-V3R2-NNN (zero-padded to at least 3 digits)

001-050: existing HARD clauses
051-099: design constitution mirror entries
100-149: design overflow (auto-expanded)
150+: new additions
```

### Canary Gate

A FROZEN clause carries `canary_gate: true`. A canary verification is required before any change.

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

The Constitution system is protected by a 5-layer safety architecture:

### Layer 1: Frozen Guard

Before a write operation, checks whether the target file is in a FROZEN zone.
On violation: blocks the write + logs it + notifies the user.

### Layer 2: Canary Check

Applies the proposed change in memory and re-evaluates the 3 most recent projects.
Rejects the change if the score drop exceeds 0.10.

### Layer 3: Contradiction Detector

When new learning conflicts with an existing rule, presents both to the user.
Automatic overwriting never happens.

### Layer 4: Rate Limiter

Limits the speed of evolution:

| Parameter | Default | Description |
|-----------|--------|------|
| `max_evolution_rate_per_week` | 3 | Maximum evolutions per week |
| `cooldown_hours` | 24 | Minimum wait time between evolutions |
| `max_active_learnings` | 50 | Maximum number of active learning entries |

### Layer 5: Human Oversight

When `require_approval: true`, every evolution proposal requires user approval.

## Using the CLI

```bash
# List the full registry
moai constitution list

# Filter to the Frozen zone
moai constitution list --zone frozen

# List clauses for a specific file only
moai constitution list --file internal/harness/scorer.go

# Output in JSON format
moai constitution list --format json
```

## Related Documentation

- [TRUST 5 Quality](/core-concepts/trust-5) — The 5 quality criteria
- [Harness Engineering](/core-concepts/harness-engineering) — Overview of the harness concept
- [SPEC-Based Development](/core-concepts/spec-based-dev) — The SPEC workflow
