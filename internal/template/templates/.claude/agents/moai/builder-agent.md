---
name: builder-agent
description: |
  Retired (SPEC-V3R2-ORC-001) — use builder-harness with artifact_type=agent.
  This agent has been consolidated into the unified builder-harness agent.
  See builder-harness.md for the active replacement.
retired: true
retired_replacement: builder-harness
retired_param_hint: "artifact_type=agent"
tools: []
skills: []
---

# builder-agent — Retired Agent

<!-- @MX:NOTE: [AUTO] retirement-pattern — matches SPEC-V3R3-RETIRED-DDD-001 stub migration -->

This agent has been retired as part of SPEC-V3R2-ORC-001 (Agent roster consolidation 22 → 17).

## Replacement

Use **builder-harness** with `artifact_type=agent` instead.

## Migration Guide

| Old Invocation | New Invocation |
|----------------|----------------|
| `Use the builder-agent subagent to create an agent` | `Use the builder-harness subagent with artifact_type=agent to create an agent` |
| `builder-agent: design a custom agent` | `builder-harness: design a custom agent (artifact_type=agent)` |

## Active Agent

See `.claude/agents/moai/builder-harness.md` for the full agent definition.
