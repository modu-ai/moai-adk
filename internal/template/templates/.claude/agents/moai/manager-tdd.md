---
name: manager-tdd
description: |
  Retired (SPEC-V3R3-RETIRED-AGENT-001) — use manager-develop with cycle_type=tdd.
  This agent has been consolidated into the unified manager-develop agent.
  See manager-develop.md for the active replacement.
retired: true
retired_replacement: manager-develop
retired_param_hint: "cycle_type=tdd"
tools: []
skills: []
---

# manager-tdd — Retired Agent

This agent has been retired as part of SPEC-V3R3-RETIRED-AGENT-001.

## Replacement

Use **manager-develop** with `cycle_type=tdd` instead.

## Migration Guide

| Old Invocation | New Invocation |
|----------------|----------------|
| `Use the manager-tdd subagent to implement the feature` | `Use the manager-develop subagent with cycle_type=tdd to implement the feature` |
| `manager-tdd: run RED-GREEN-REFACTOR cycle` | `manager-develop: run TDD cycle (cycle_type=tdd)` |

## Why This Change

The `manager-tdd` and `manager-ddd` agents have been consolidated into a single `manager-develop` agent that supports both DDD (ANALYZE-PRESERVE-IMPROVE) and TDD (RED-GREEN-REFACTOR) cycles through the `cycle_type` parameter. This unification:

- Eliminates duplication between the two agents
- Provides a single entry point for all implementation cycles
- Enables future cycle types without additional agent proliferation

## Active Agent

See `.claude/agents/moai/manager-develop.md` for the full agent definition.
