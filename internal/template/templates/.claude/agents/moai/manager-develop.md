---
name: manager-develop
description: |
  Retired (SPEC-V3R2-ORC-001) — use manager-cycle with cycle_type=ddd or cycle_type=tdd.
  This agent has been renamed and consolidated into the unified manager-cycle agent.
  See manager-cycle.md for the active replacement.
retired: true
retired_replacement: manager-cycle
retired_param_hint: "cycle_type=ddd|tdd"
tools: []
skills: []
---

<!-- @MX:NOTE: [AUTO] retirement-pattern — manager-develop renamed to manager-cycle (SPEC-V3R2-ORC-001); canonical name is now manager-cycle -->

# manager-develop — Retired Agent

This agent has been retired as part of SPEC-V3R2-ORC-001 (Agent Roster Consolidation 22→17).
The name `manager-develop` was an intermediate name used during SPEC-V3R2-RT-005; the canonical
name is `manager-cycle` per SPEC-V3R2-ORC-001 naming conventions.

## Replacement

Use **manager-cycle** with `cycle_type=ddd` or `cycle_type=tdd` instead.

## Migration Guide

| Old Invocation | New Invocation |
|----------------|----------------|
| `Use the manager-develop subagent to implement the feature` | `Use the manager-cycle subagent with cycle_type=ddd to implement the feature` |
| `manager-develop: run DDD cycle` | `manager-cycle: run DDD cycle (cycle_type=ddd)` |
| `manager-develop: run TDD cycle` | `manager-cycle: run TDD cycle (cycle_type=tdd)` |

## Why This Change

`manager-develop` was a transitional name introduced during SPEC-V3R2-RT-005 for the unified
DDD+TDD implementation agent. SPEC-V3R2-ORC-001 canonizes `manager-cycle` as the name to match
the `cycle_type` parameter and the consistent `manager-*` naming pattern used across the roster.

## Active Agent

See `.claude/agents/moai/manager-cycle.md` for the full agent definition.
