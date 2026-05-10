---
name: builder-plugin
description: |
  Retired (SPEC-V3R2-ORC-001) — use builder-platform with artifact_type=plugin.
  This agent has been consolidated into the unified builder-platform agent.
  See builder-platform.md for the active replacement.
retired: true
retired_replacement: builder-platform
retired_param_hint: "artifact_type=plugin"
tools: []
skills: []
---

# builder-plugin — Retired Agent

<!-- @MX:NOTE: [AUTO] retirement-pattern — matches SPEC-V3R3-RETIRED-DDD-001 stub migration -->

This agent has been retired as part of SPEC-V3R2-ORC-001 (Agent roster consolidation 22 → 17).

## Replacement

Use **builder-platform** with `artifact_type=plugin` instead.

## Migration Guide

| Old Invocation | New Invocation |
|----------------|----------------|
| `Use the builder-plugin subagent to create a plugin` | `Use the builder-platform subagent with artifact_type=plugin to create a plugin` |
| `builder-plugin: build a marketplace plugin` | `builder-platform: build a marketplace plugin (artifact_type=plugin)` |

## Active Agent

See `.claude/agents/moai/builder-platform.md` for the full agent definition.
