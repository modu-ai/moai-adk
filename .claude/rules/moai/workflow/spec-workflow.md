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
| Plan | manager-spec (single) | team-researcher + team-analyst + team-architect (parallel) | Complexity >= threshold |
| Run | manager-ddd/tdd (sequential) | team-backend-dev + team-frontend-dev + team-tester (parallel) | Domains >= 3 or files >= 10 |
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

### Token Cost Awareness

Agent teams use significantly more tokens than a single session. Each teammate has its own independent context window, so token usage scales linearly with the number of active teammates.

Estimated token multipliers by team pattern:
- plan_research (3 teammates): ~3x plan phase tokens
- implementation (3 teammates): ~3x run phase tokens
- design_implementation (4 teammates): ~4x run phase tokens
- investigation (3 teammates): ~2x (haiku model reduces cost)
- review (3 teammates): ~2x (read-only, shorter sessions)

When to prefer team mode over sub-agent mode:
- Research and review tasks where parallel exploration adds real value
- Cross-layer features (frontend + backend + tests)
- Complex debugging with multiple potential root causes
- Tasks where teammates need to communicate and coordinate

When to prefer sub-agent mode:
- Sequential tasks with heavy dependencies
- Same-file edits or tightly coupled changes
- Routine tasks with clear single-domain scope
- Token budget is a concern

### Team Workflow References

Detailed team orchestration steps are defined in dedicated workflow files:

- Plan phase: @.claude/skills/moai/workflows/team-plan.md
- Run phase: @.claude/skills/moai/workflows/team-run.md
- Fix phase: @.claude/skills/moai/workflows/team-debug.md
- Review: @.claude/skills/moai/workflows/team-review.md

### Known Limitations

Agent teams are experimental. Current limitations from Claude Code documentation:

- No session resumption: /resume and /rewind do not restore in-process teammates. After resuming a session, the lead may attempt to message teammates that no longer exist. Spawn new teammates if this occurs.
- Task status can lag: Teammates sometimes fail to mark tasks as completed, which blocks dependent tasks. Check whether work is actually done and update task status manually if needed.
- Shutdown can be slow: Teammates finish their current request or tool call before shutting down.
- One team per session: A lead can only manage one team at a time. Clean up the current team before starting a new one.
- No nested teams: Teammates cannot spawn their own teams or teammates. Only the lead can manage the team.
- Lead is fixed: The session that creates the team is the lead for its lifetime. Leadership cannot be transferred.
- Permissions set at spawn: All teammates start with the lead's permission mode. Individual teammate modes can be changed after spawning via permissionMode in agent definitions.
- Split panes require tmux or iTerm2: The default in-process mode works in any terminal. Split-pane mode is not supported in VS Code integrated terminal, Windows Terminal, or Ghostty.

### Prerequisites

Both conditions must be met for team mode:
- `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` in environment or settings.json env
- `workflow.team.enabled: true` in `.moai/config/sections/workflow.yaml`

If prerequisites are not met, all subcommands gracefully fall back to sub-agent mode.

### Mode Selection
- --team flag: Force team mode
- --solo flag: Force sub-agent mode
- No flag (default): Complexity-based selection
- See workflow.yaml team.auto_selection for thresholds

### Fallback
If team mode fails or prerequisites are not met:
- Graceful fallback to sub-agent mode
- Continue from last completed task
- No data loss or state corruption
- Trigger conditions: AGENT_TEAMS env not set, workflow.team.enabled false, TeamCreate failure, teammate spawn failure
