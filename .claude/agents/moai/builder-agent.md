---
name: builder-agent
description: "Retired — use builder-platform with artifact_type=agent"
status: retired
---

This agent has been retired. Use `builder-platform` with `artifact_type=agent` instead.

## Migration Notes

The `builder-agent` agent has been consolidated into `builder-platform` as part of SPEC-V3R2-ORC-001 (Agent Roster Consolidation).

**Old invocation**:
```
Use the builder-agent subagent to create a new sub-agent
```

**New invocation**:
```
Use the builder-platform subagent with artifact_type=agent to create a new sub-agent
```

## What Changed

- All agent creation capabilities are preserved in builder-platform
- Trigger keywords remain the same
- Tools, skills, and configuration unchanged
- The only difference is the `artifact_type` parameter in the spawn prompt

## Documentation

For the complete agent creation workflow, see:
- `.claude/agents/moai/builder-platform.md` (unified builder)
- `.claude/rules/moai/development/agent-authoring.md` (agent authoring guide)
