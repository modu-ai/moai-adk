---
name: manager-cycle
description: |
  Retired (SPEC-V3R2-ORC-001 follow-up rename) — use manager-develop with cycle_type=ddd or cycle_type=tdd.
  This agent has been renamed to manager-develop for clearer naming.
  See manager-develop.md for the active replacement.
retired: true
retired_replacement: manager-develop
retired_param_hint: "cycle_type=ddd|tdd"
tools: []
skills: []
---

<!-- @MX:NOTE: [AUTO] retirement-pattern — manager-cycle renamed to manager-develop (ORC-001 follow-up); canonical name is now manager-develop -->

# manager-cycle — Retired Agent

This agent has been retired as part of the ORC-001 follow-up rename.
The name `manager-cycle` has been renamed to `manager-develop` for clearer, more intuitive naming.

## Replacement

Use **manager-develop** with `cycle_type=ddd` or `cycle_type=tdd` instead.

## Migration Guide

| Old Invocation | New Invocation |
|----------------|----------------|
| `Use the manager-cycle subagent to implement the feature` | `Use the manager-develop subagent with cycle_type=ddd to implement the feature` |
| `manager-cycle: run DDD cycle (cycle_type=ddd)` | `manager-develop: run DDD cycle (cycle_type=ddd)` |
| `manager-cycle: run TDD cycle (cycle_type=tdd)` | `manager-develop: run TDD cycle (cycle_type=tdd)` |

## Why This Change

`manager-cycle` was an intermediate name. `manager-develop` is the canonical name that
clearly expresses the role: the unified development implementation agent supporting both
DDD (ANALYZE-PRESERVE-IMPROVE) and TDD (RED-GREEN-REFACTOR) cycles through the
`cycle_type` parameter.

## Active Agent

See `.claude/agents/moai/manager-develop.md` for the full agent definition.
