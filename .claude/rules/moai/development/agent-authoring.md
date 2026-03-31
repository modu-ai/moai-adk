---
paths: "**/.claude/agents/**"
---

# Agent Authoring

Guidelines for creating custom agents in MoAI-ADK.

## Agent Definition Location

Custom agents are defined in `.claude/agents/*.md` or `.claude/agents/**/*.md` (subdirectories supported).

Directory convention:
- User custom agents: `.claude/agents/<agent-name>.md` (root level)
- MoAI-ADK system agents: `.claude/agents/moai/<agent-name>.md` (moai subdirectory)

Platform Support: Windows ARM64 (`win32-arm64`) is natively supported as of Claude Code v2.1.41. No WSL required for ARM-based Windows devices.

## Supported Frontmatter Fields

All agent definitions use YAML frontmatter. The following fields are available:

| Field | Required | Default | Description |
|-------|----------|---------|-------------|
| name | Yes | - | Unique identifier, lowercase with hyphens |
| description | Yes | - | When Claude should delegate to this agent |
| tools | No | Inherits all | Tools the agent can use (allowlist approach) |
| disallowedTools | No | None | Tools to deny (denylist approach, alternative to tools) |
| model | No | inherit | Model selection: sonnet, opus, haiku, or inherit |
| permissionMode | No | default | Permission behavior for the agent |
| maxTurns | No | Unlimited | Maximum agentic turns before stopping (deprecated since v2.1.69+, use maxContextSize instead) |
| skills | No | None | Skills injected into agent context at startup |
| mcpServers | No | None | MCP servers available to this agent |
| hooks | No | None | Lifecycle hooks scoped to this agent |
| memory | No | None | Persistent memory scope for cross-session learning |
| background | No | false | Run agent in background without blocking conversation (v2.1.46+) |
| isolation | No | none | Isolation mode: "worktree" creates isolated git worktree (v2.1.49+) |

### Field Details

**tools**: When specified, the agent can only use listed tools. When omitted, the agent inherits all tools from the parent. Mutually exclusive with disallowedTools.

**disallowedTools**: Denylist approach. The agent inherits all tools except those listed. Mutually exclusive with tools.

**skills**: Full skill content is injected into the agent context, not just made available for invocation. Agents do not inherit skills from the parent. Each skill listed must exist in `.claude/skills/`.

**mcpServers**: Either a server name reference (matching a key in `.mcp.json`) or an inline server definition with command and args.

**hooks**: Supports PreToolUse, PostToolUse, and SubagentStop events scoped to this agent. See @hooks-system.md for configuration format.

**background**: When set to true, the agent runs in the background without blocking the main conversation. Results are delivered asynchronously on the next turn. Available since Claude Code v2.1.46.

**isolation**: Controls agent execution isolation. When set to "worktree", the agent runs in an isolated git worktree, preventing conflicts with the main working directory. Available since Claude Code v2.1.49.

## Agent(agent_type) Restrictions

The `tools` field supports `Task(worker, researcher)` syntax to restrict which subagents an agent can spawn.

- Only applies to agents running as the main thread via `claude --agent`
- Has no effect on subagent definitions (subagents cannot spawn other subagents)
- MoAI agents run as subagents, so this restriction is not currently applicable
- Useful for creating coordinator agents that run as the main thread

## Permission Modes

The `permissionMode` field controls how the agent handles permission checks:

| Mode | Behavior | Use Case |
|------|----------|----------|
| default | Standard permission checking with user prompts | General-purpose agents |
| acceptEdits | Auto-accept file edit operations | Trusted implementation agents |
| delegate | Coordination-only mode, restricts to team management tools | Team lead agents |
| dontAsk | Auto-deny all permission prompts | Strict sandbox agents |
| bypassPermissions | Skip all permission checks (use with caution) | Fully trusted automation |
| plan | Read-only exploration mode, no write operations | Research and analysis agents |

## Persistent Memory

The `memory` field enables cross-session learning for agents. Three scope levels:

| Scope | Storage Location | Shared via VCS | Use Case |
|-------|-----------------|----------------|----------|
| user | ~/.claude/agent-memory/\<name\>/ | No | Cross-project learnings, personal preferences |
| project | .claude/agent-memory/\<name\>/ | Yes | Project-specific knowledge, team-shared context |
| local | .claude/agent-memory-local/\<name\>/ | No | Project-specific knowledge, not shared |

## Agent Categories

### Manager Agents (8)

Coordinate workflows and multi-step processes:

- manager-spec: SPEC document creation
- manager-ddd: DDD implementation cycle
- manager-tdd: TDD implementation cycle
- manager-docs: Documentation generation
- manager-quality: Quality gates validation
- manager-project: Project configuration
- manager-strategy: System design, architecture decisions
- manager-git: Git operations, branching strategy

### Expert Agents (8)

Domain-specific implementation:

- expert-backend: API and server development
- expert-frontend: UI and client development
- expert-security: Security analysis
- expert-devops: CI/CD and infrastructure
- expert-performance: Performance optimization
- expert-debug: Debugging and troubleshooting
- expert-testing: Test creation and strategy
- expert-refactoring: Code refactoring

### Builder Agents (3)

Create new MoAI components:

- builder-agent: New agent definitions
- builder-skill: New skill creation
- builder-plugin: Plugin creation

### Dynamic Team Generation (Experimental)

**Architecture**: Agent Teams teammates are spawned dynamically using `Agent(subagent_type: "general-purpose")` with runtime parameter overrides. No static team-* agent definition files are used.

**Key distinction from regular subagents**:
- Regular subagents: spawned from main conversation, return results, cannot communicate with each other
- Dynamic teammates: spawned with `team_name` + `name` parameters, get Agent Teams tools (SendMessage, TaskList etc.) automatically injected by the framework

**Spawn pattern** (Agent Teams only):
```
Agent(subagent_type: "general-purpose", team_name: "...", name: "researcher", model: "haiku", mode: "plan")
```

Role profiles are defined in `.moai/config/sections/workflow.yaml` under `team.role_profiles`:

| Role Profile | Default Model | Mode | Isolation | Purpose |
|-------------|---------------|------|-----------|---------|
| researcher | haiku | plan (read-only) | none | Codebase exploration, analysis |
| analyst | sonnet | plan (read-only) | none | Requirements analysis, validation |
| architect | sonnet | plan (read-only) | none | Solution design, architecture |
| implementer | sonnet | acceptEdits | worktree | Backend, frontend, full-stack code |
| tester | sonnet | acceptEdits | worktree | Test creation, coverage validation |
| designer | sonnet | acceptEdits | worktree | UI/UX design with MCP tools |
| reviewer | haiku | plan (read-only) | none | Code review, quality validation |

Requires: `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` in settings.json env

## Rules

- Write agent definitions in English
- Define expertise domain clearly in description
- Minimize tool permissions (least privilege)
- Include relevant trigger keywords
- Use permissionMode: plan for read-only agents
- Preload skills for domain expertise instead of relying on runtime loading

## Tool Permissions

Recommended tool sets by category:

Manager agents: Read, Write, Edit, Grep, Glob, Bash, Skill, TodoWrite (NOTE: Agent tool is NOT included by default for regular subagents. However, Agent Teams teammates CAN spawn other teammates using Agent() with the team_name parameter, v2.1.50+)

Expert agents: Read, Write, Edit, Grep, Glob, Bash

Builder agents: Read, Write, Edit, Grep, Glob

Dynamic teammates (general-purpose): Inherit all tools from parent session. Permission control via `mode` parameter at spawn time.

Notes:
- Dynamic teammates use `mode: "plan"` for read-only enforcement instead of tool restrictions
- Project-specific context is included in the spawn prompt, not preloaded skills
- Teammates can self-load skills via Skill() tool when deeper documentation is needed

## Agent Invocation

Invoke agents via Agent tool:

- "Use the expert-backend subagent to implement the API"
- Agent tool with subagent_type parameter

For team mode invocation:
- TeamCreate to initialize team structure
- Agent() with team_name and name parameters to spawn teammates
- SendMessage for inter-teammate coordination
- TeamDelete after all teammates shut down
- See team-plan.md and team-run.md for complete workflow examples
