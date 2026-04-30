---
name: builder-skill
description: "Retired — use builder-platform with artifact_type=skill"
status: retired
---

This agent has been retired. Use `builder-platform` with `artifact_type=skill` instead.

## Migration Notes

The `builder-skill` agent has been consolidated into `builder-platform` as part of SPEC-V3R2-ORC-001 (Agent Roster Consolidation).

**Old invocation**:
```
Use the builder-skill subagent to create a new skill
```

**New invocation**:
```
Use the builder-platform subagent with artifact_type=skill to create a new skill
```

## What Changed

- All skill creation capabilities are preserved in builder-platform
- Trigger keywords remain the same
- Tools, skills, and configuration unchanged
- The only difference is the `artifact_type` parameter in the spawn prompt

## Documentation

For the complete skill creation workflow, see:
- `.claude/agents/moai/builder-platform.md` (unified builder)
- `.claude/rules/moai/development/skill-authoring.md` (skill authoring guide)
