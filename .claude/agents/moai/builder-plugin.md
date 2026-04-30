---
name: builder-plugin
description: "Retired — use builder-platform with artifact_type=plugin"
status: retired
---

This agent has been retired. Use `builder-platform` with `artifact_type=plugin` instead.

## Migration Notes

The `builder-plugin` agent has been consolidated into `builder-platform` as part of SPEC-V3R2-ORC-001 (Agent Roster Consolidation).

**Old invocation**:
```
Use the builder-plugin subagent to create a new plugin
```

**New invocation**:
```
Use the builder-platform subagent with artifact_type=plugin to create a new plugin
```

## What Changed

- All plugin creation capabilities are preserved in builder-platform
- Trigger keywords remain the same
- Tools, skills, and configuration unchanged
- The only difference is the `artifact_type` parameter in the spawn prompt

## Documentation

For the complete plugin creation workflow, see:
- `.claude/agents/moai/builder-platform.md` (unified builder)
- `.claude/rules/moai/development/plugin-authoring.md` (plugin authoring guide)
