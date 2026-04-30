---
name: expert-testing
description: "Retired — use manager-cycle (strategy) or expert-performance (load testing)"
status: retired
---

This agent has been retired. Test strategy is now handled by specialist agents.

## Migration Notes

The `expert-testing` agent has been retired as part of SPEC-V3R2-ORC-001 (Agent Roster Consolidation).

**For test strategy design**:
```
Use the manager-cycle subagent with appropriate cycle_type to implement test strategy
```

**For load testing**:
```
Use the expert-performance subagent for load testing and benchmarking
```

**For E2E testing**:
```
Use the expert-frontend or expert-backend subagent for E2E test implementation
```

## What Changed

The testing capabilities have been redistributed to specialist agents:
- **Unit/Integration tests**: `manager-cycle` (tdd) for test-first development
- **Load testing**: `expert-performance` for k6/Locust/JMeter execution
- **E2E testing**: `expert-frontend` for Playwright/Cypress, `expert-backend` for API tests
- **Test strategy**: Coordinated by `manager-cycle` with appropriate cycle_type

## Documentation

For testing workflows, see:
- `.claude/agents/moai/manager-cycle.md` (test implementation)
- `.claude/agents/moai/expert-performance.md` (load testing)
- `.claude/skills/moai-workflow-testing/SKILL.md` (testing workflow)
