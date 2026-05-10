---
name: builder-skill
description: |
  Retired (SPEC-V3R2-ORC-001) — use builder-platform with artifact_type=skill.
  This agent has been consolidated into the unified builder-platform agent.
  See builder-platform.md for the active replacement.
retired: true
retired_replacement: builder-platform
retired_param_hint: "artifact_type=skill"
tools: []
skills: []
---

# builder-skill — Retired Agent

<!-- @MX:NOTE: [AUTO] retirement-pattern — matches SPEC-V3R3-RETIRED-DDD-001 stub migration -->

This agent has been retired as part of SPEC-V3R2-ORC-001 (Agent roster consolidation 22 → 17).

## Replacement

Use **builder-platform** with `artifact_type=skill` instead.

## Migration Guide

| Old Invocation | New Invocation |
|----------------|----------------|
| `Use the builder-skill subagent to create a skill` | `Use the builder-platform subagent with artifact_type=skill to create a skill` |
| `builder-skill: design a knowledge skill` | `builder-platform: design a knowledge skill (artifact_type=skill)` |

## Active Agent

See `.claude/agents/moai/builder-platform.md` for the full agent definition.
