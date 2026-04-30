---
name: expert-debug
description: "Retired — use manager-quality (diagnostic sub-mode) or manager-cycle with strategy delegation"
status: retired
---

This agent has been retired. Use `manager-quality` (diagnostic sub-mode) or delegate to specialized agents.

## Migration Notes

The `expert-debug` agent has been retired as part of SPEC-V3R2-ORC-001 (Agent Roster Consolidation).

**For error diagnosis and routing**:
```
Use the manager-quality subagent in diagnostic mode to analyze the error
```

**For implementation fixes**:
```
Use the manager-cycle subagent (with appropriate cycle_type) to implement the fix
```

## What Changed

The diagnostic capabilities have been absorbed into `manager-quality`:
- Error message parsing and classification
- File location analysis
- Pattern matching against known error types
- Impact assessment
- Solution proposal and agent delegation

The implementation capabilities are delegated to:
- `manager-cycle` (ddd/tdd) for code fixes requiring test coverage
- `manager-git` for git-specific issues
- `manager-quality` for quality gate failures

## Documentation

For error diagnosis and routing, see:
- `.claude/agents/moai/manager-quality.md` (diagnostic sub-mode section)
- `.claude/agents/moai/manager-cycle.md` (implementation agent)
