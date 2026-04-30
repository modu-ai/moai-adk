---
name: manager-tdd
description: "Retired — use manager-cycle with cycle_type=tdd"
status: retired
---

This agent has been retired. Use `manager-cycle` with `cycle_type=tdd` instead.

## Migration Notes

The `manager-tdd` agent has been consolidated into `manager-cycle` as part of SPEC-V3R2-ORC-001 (Agent Roster Consolidation).

**Old invocation**:
```
Use the manager-tdd subagent to implement the new feature
```

**New invocation**:
```
Use the manager-cycle subagent with cycle_type=tdd to implement the new feature
```

## What Changed

- All TDD-specific capabilities (RED-GREEN-REFACTOR cycle) are preserved in manager-cycle
- Trigger keywords remain the same
- Hooks, tools, and skills configuration unchanged
- The only difference is the `cycle_type` parameter in the spawn prompt

## Documentation

For the complete TDD workflow, see:
- `.claude/agents/moai/manager-cycle.md` (unified agent)
- `.claude/skills/moai-workflow-tdd/SKILL.md` (TDD workflow skill)
