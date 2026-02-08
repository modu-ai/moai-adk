# SPEC Workflow

MoAI's three-phase development workflow with token budget management.

## Phase Overview

| Phase | Command | Agent | Token Budget | Purpose |
|-------|---------|-------|--------------|---------|
| Plan | /moai plan | manager-spec | 30K | Create SPEC document |
| Run | /moai run | manager-ddd/tdd (per quality.yaml) | 180K | DDD/TDD implementation |
| Sync | /moai sync | manager-docs | 40K | Documentation sync |

## Plan Phase

Create comprehensive specification using EARS format.

Token Strategy:
- Allocation: 30,000 tokens
- Load requirements only
- Execute /clear after completion
- Saves 45-50K tokens for implementation

Output:
- SPEC document at `.moai/specs/SPEC-XXX/spec.md`
- EARS format requirements
- Acceptance criteria
- Technical approach

## Run Phase

Implement specification using configured development methodology.

Token Strategy:
- Allocation: 180,000 tokens
- Selective file loading
- Enables 70% larger implementations

Development Methodology:
- Configured in quality.yaml (development_mode: ddd, tdd, or hybrid)
- See @workflow-modes.md for detailed methodology cycles

Success Criteria:
- All SPEC requirements implemented
- Methodology-specific tests passing
- 85%+ code coverage
- TRUST 5 quality gates passed

## Sync Phase

Generate documentation and prepare for deployment.

Token Strategy:
- Allocation: 40,000 tokens
- Result caching
- 60% fewer redundant file reads

Output:
- API documentation
- Updated README
- CHANGELOG entry
- Pull request

## Completion Markers

AI uses markers to signal task completion:
- `<moai>DONE</moai>` - Task complete
- `<moai>COMPLETE</moai>` - Full completion

## Context Management

/clear Strategy:
- After /moai plan completion (mandatory)
- When context exceeds 150K tokens
- Before major phase transitions

Progressive Disclosure:
- Level 1: Metadata only (~100 tokens)
- Level 2: Skill body when triggered (~5000 tokens)
- Level 3: Bundled files on-demand

## Phase Transitions

Plan to Run:
- Trigger: SPEC document approved
- Action: Execute /clear, then /moai run SPEC-XXX

Run to Sync:
- Trigger: Implementation complete, tests passing
- Action: Execute /moai sync SPEC-XXX

## Agent Teams Variant

When team mode is enabled (workflow.team.enabled and AGENT_TEAMS env), phases can execute with Agent Teams instead of sub-agents.

### Team Mode Phase Overview

| Phase | Sub-agent Mode | Team Mode | Condition |
|-------|---------------|-----------|-----------|
| Plan | manager-spec (single) | researcher + analyst + architect (parallel) | Complexity >= threshold |
| Run | manager-ddd/tdd (sequential) | backend-dev + frontend-dev + tester (parallel) | Domains >= 3 or files >= 10 |
| Sync | manager-docs (single) | manager-docs (always sub-agent) | N/A |

### Team Mode Plan Phase
- TeamCreate for parallel research team
- Teammates explore codebase, analyze requirements, design approach
- MoAI synthesizes into SPEC document
- Shutdown team, /clear before Run phase

### Team Mode Run Phase
- TeamCreate for implementation team
- Task decomposition with file ownership boundaries
- Teammates self-claim tasks from shared list
- Quality validation after all implementation completes
- Shutdown team

### Mode Selection
- --team flag: Force team mode
- --solo flag: Force sub-agent mode
- No flag (default): Complexity-based selection
- See workflow.yaml team.auto_selection for thresholds

### Fallback
If team mode fails or is unavailable:
- Graceful fallback to sub-agent mode
- Continue from last completed task
- No data loss or state corruption
