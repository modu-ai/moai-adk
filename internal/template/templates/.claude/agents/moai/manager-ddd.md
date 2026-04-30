---
name: manager-ddd
description: "Retired — use manager-cycle with cycle_type=ddd"
status: retired
---

This agent has been retired. Use `manager-cycle` with `cycle_type=ddd` instead.

## Migration Notes

The `manager-ddd` agent has been consolidated into `manager-cycle` as part of SPEC-V3R2-ORC-001 (Agent Roster Consolidation).

**Old invocation**:
```
Use the manager-ddd subagent to refactor the authentication module
```

**New invocation**:
```
Use the manager-cycle subagent with cycle_type=ddd to refactor the authentication module
```

## What Changed

- All DDD-specific capabilities (ANALYZE-PRESERVE-IMPROVE cycle) are preserved in manager-cycle
- Trigger keywords remain the same
- Hooks, tools, and skills configuration unchanged
- The only difference is the `cycle_type` parameter in the spawn prompt

## Documentation

For the complete DDD workflow, see:
- `.claude/agents/moai/manager-cycle.md` (unified agent)
- `.claude/skills/moai-workflow-ddd/SKILL.md` (DDD workflow skill)
