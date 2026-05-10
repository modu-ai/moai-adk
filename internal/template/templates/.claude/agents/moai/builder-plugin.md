---
name: builder-plugin
description: |
  Retired (SPEC-V3R2-ORC-001) — use builder-harness with artifact_type=plugin.
  This agent has been consolidated into the unified builder-harness agent.
  See builder-harness.md for the active replacement.
retired: true
retired_replacement: builder-harness
retired_param_hint: "artifact_type=plugin"
tools: []
skills: []
---

# builder-plugin — Retired Agent

<!-- @MX:NOTE: [AUTO] retirement-pattern — matches SPEC-V3R3-RETIRED-DDD-001 stub migration -->

This agent has been retired as part of SPEC-V3R2-ORC-001 (Agent roster consolidation 22 → 17).

## Replacement

Use **builder-harness** with `artifact_type=plugin` instead.

## Migration Guide

| Old Invocation | New Invocation |
|----------------|----------------|
| `Use the builder-plugin subagent to create a plugin` | `Use the builder-harness subagent with artifact_type=plugin to create a plugin` |
| `builder-plugin: build a marketplace plugin` | `builder-harness: build a marketplace plugin (artifact_type=plugin)` |

## Active Agent

See `.claude/agents/moai/builder-harness.md` for the full agent definition.
