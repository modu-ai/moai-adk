# SPEC-EVO-001 Execution Plan

## Master Initiative: Skill Evolution Preservation System

This document is the execution plan for the 6-SPEC initiative to add skill evolution capability to moai-adk while preserving user modifications across `moai update`.

## SPEC Index

| SPEC | Wave | Status | Blocks | Blocked By |
|------|------|--------|--------|------------|
| [SPEC-EVO-001](./spec.md) | 0 | Draft | All others | - |
| [SPEC-SKILL-ENHANCE-001](../SPEC-SKILL-ENHANCE-001/spec.md) | 1 | Draft | - | EVO-001 |
| [SPEC-CORE-BEHAV-001](../SPEC-CORE-BEHAV-001/spec.md) | 1 | Draft | - | EVO-001 |
| [SPEC-TELEMETRY-001](../SPEC-TELEMETRY-001/spec.md) | 1 | Draft | REFLECT-001 | EVO-001 |
| [SPEC-REFLECT-001](../SPEC-REFLECT-001/spec.md) | 2 | Draft | - | TELEMETRY-001, EVO-001 |
| [SPEC-THIN-CMDS-001](../SPEC-THIN-CMDS-001/spec.md) | 2 | Draft | - | EVO-001 |

## Wave Execution Order

### Wave 0 — Foundation (Solo)
**SPEC-EVO-001**: Must complete before Wave 1 can start.

Key deliverables:
- `.moai/evolution/` directory infrastructure
- `internal/cli/update.go` protected paths extension
- `internal/merge/evolvable_zone.go` marker-aware merge
- SessionStart symlink creation for new skills
- Manifest schema

### Wave 1 — Parallel (4 SPECs)
After Wave 0 completes, these can run in parallel (ideally via Agent Teams):

1. **SPEC-SKILL-ENHANCE-001**: Add Anti-Rationalization/Red Flags/Verification to all 41 skills
2. **SPEC-CORE-BEHAV-001**: Constitution Agent Core Behaviors + HUMAN GATE markers
3. **SPEC-TELEMETRY-001**: Skill usage JSONL recorder + outcome heuristics + report command

Wave 1 outputs are independent — no cross-SPEC file conflicts expected.

### Wave 2 — Parallel (2 SPECs)
After Wave 1 completes:

1. **SPEC-REFLECT-001**: Reflective Write Hook (depends on TELEMETRY-001)
2. **SPEC-THIN-CMDS-001**: Command audit + test enforcement (depends on EVO-001 only, parallelizable with REFLECT-001)

## Dependency Graph

```
       SPEC-EVO-001 (Wave 0)
              │
   ┌──────────┼──────────┬──────────┐
   │          │          │          │
   ▼          ▼          ▼          ▼
SKILL-    CORE-      TELEMETRY-   THIN-CMDS
ENHANCE   BEHAV                   (can wait)
   │          │          │          │
   │          │          ▼          │
   │          │      REFLECT-001    │
   │          │                     │
   └──────────┴──────┬──────────────┘
                     │
                     ▼
              (All merged to main)
```

## Estimated Scope

| SPEC | Go LOC | Template LOC | Test LOC | Files Modified |
|------|--------|--------------|----------|----------------|
| EVO-001 | ~800 | ~50 | ~400 | 7 Go + 2 template |
| SKILL-ENHANCE-001 | 0 | ~4000 | 0 | 41 SKILL.md + `make build` |
| CORE-BEHAV-001 | 0 | ~300 | 0 | 5 template files |
| TELEMETRY-001 | ~600 | 0 | ~400 | 8 new Go files |
| REFLECT-001 | ~1200 | 0 | ~600 | 10 new Go files |
| THIN-CMDS-001 | ~200 | ~50 | ~100 | 1 new test + audit only |

**Total**: ~2800 Go LOC, ~4400 template LOC, ~1500 test LOC, ~100 files touched

## Execution Mode Recommendation

**Wave 0**: Single manager-ddd agent (sub-agent mode, foreground) — foundational, needs careful integration testing.

**Wave 1**: Agent Teams with 3 parallel teammates — independent work, benefits from isolation.
- Teammate 1 (role: implementer): SKILL-ENHANCE-001 template work
- Teammate 2 (role: implementer): CORE-BEHAV-001 template work
- Teammate 3 (role: implementer): TELEMETRY-001 Go code

**Wave 2**: Two sequential or parallel manager-ddd agents:
- REFLECT-001 first (higher complexity, more risky)
- THIN-CMDS-001 can run anytime after EVO-001 (very low risk)

## Safety Protocol

Per user requirements and CLAUDE.md Section 7:

1. **Approach-First (Rule 1)**: Each SPEC's implementation must present approach before coding. Done via SPEC docs themselves.
2. **Multi-File Decomposition (Rule 2)**: Each SPEC has TodoList broken by file. Enforced by manager agents.
3. **Post-Implementation Review (Rule 3)**: Every SPEC lists potential issues + test suggestions in Risk Assessment section.
4. **Reproduction-First Bug Fix (Rule 4)**: N/A for new feature SPECs; applies if bugs emerge.
5. **Context-First Discovery (Rule 5)**: Exploration already completed (4 parallel exploration agents).

## Test Strategy Overview

- **EVO-001**: Unit tests for marker parser, integration test for update preservation
- **SKILL-ENHANCE-001**: Template build test, YAML frontmatter validator
- **CORE-BEHAV-001**: Rules file consistency test (no duplicate rules)
- **TELEMETRY-001**: Unit tests for recorder/outcome/report, integration test for hook flow
- **REFLECT-001**: Unit tests for each safety layer, integration test for proposal flow
- **THIN-CMDS-001**: Command pattern audit test (this IS the main deliverable)

## Rollback Plan

Each SPEC is designed to be rollback-safe:

- **EVO-001**: Evolution directory is additive; removing it is safe
- **SKILL-ENHANCE-001**: Added sections are wrapped in markers; can be removed via merge
- **CORE-BEHAV-001**: Added constitution section; revert preserves old behavior
- **TELEMETRY-001**: Recording only; no behavioral changes to existing flows
- **REFLECT-001**: L5 gates all proposals; can be disabled via config flag
- **THIN-CMDS-001**: Pure audit + test; no behavior change

## Success Criteria (Initiative-Level)

- [ ] All 6 SPECs completed with their individual acceptance criteria
- [ ] `moai update` preserves all evolvable zones and .moai/evolution/ content
- [ ] All 41 skills have anti-rationalization/red-flags/verification sections
- [ ] moai-constitution.md has 6 core behaviors section
- [ ] Skill usage telemetry writes to .moai/evolution/telemetry/
- [ ] Reflective Write generates proposals requiring user approval
- [ ] Command thin-pattern test enforces pattern
- [ ] `go test ./...` passes
- [ ] `make build` succeeds
- [ ] `make install` succeeds

## Open Questions

None — user confirmed all design choices in previous turn:
- Hybrid marker + directory preservation
- Wave-based parallel execution
- All 41 skills for enhancement scope
- User approval required for Reflective Write (never auto-apply)
