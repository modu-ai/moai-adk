---
name: manager-ddd
description: |
  Retired (SPEC-V3R3-RETIRED-DDD-001) — use manager-cycle with cycle_type=ddd.
  This agent has been consolidated into the unified manager-cycle agent.
  See manager-cycle.md for the active replacement.
retired: true
retired_replacement: manager-cycle
retired_param_hint: "cycle_type=ddd"
tools: []
skills: []
---

# manager-ddd — Retired Agent

This agent has been retired as part of SPEC-V3R3-RETIRED-DDD-001 (following SPEC-V3R3-RETIRED-AGENT-001).

## Replacement

Use **manager-cycle** with `cycle_type=ddd` instead.

## Migration Guide

| Old Invocation | New Invocation |
|----------------|----------------|
| `Use the manager-ddd subagent to implement the feature` | `Use the manager-cycle subagent with cycle_type=ddd to implement the feature` |
| `manager-ddd: run ANALYZE-PRESERVE-IMPROVE cycle` | `manager-cycle: run DDD cycle (cycle_type=ddd)` |

## Why This Change

The `manager-ddd` agent has been consolidated into the `manager-cycle` agent, which supports DDD (ANALYZE-PRESERVE-IMPROVE) cycles through the `cycle_type` parameter. This unification:

- Eliminates duplication between the two agents
- Provides a single entry point for all implementation cycles
- Enables future cycle types without additional agent proliferation

## Active Agent

See `.claude/agents/moai/manager-cycle.md` for the full agent definition.
