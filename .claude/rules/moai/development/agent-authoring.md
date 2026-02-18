# Agent Authoring

Guidelines for creating custom agents in MoAI-ADK.

## Agent Definition Location

Custom agents are defined in `.claude/agents/*.md` or `.claude/agents/**/*.md` (subdirectories supported).

Directory convention:
- User custom agents: `.claude/agents/<agent-name>.md` (root level)
- MoAI-ADK system agents: `.claude/agents/moai/<agent-name>.md` (moai subdirectory)

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
| maxTurns | No | Unlimited | Maximum agentic turns before stopping |
| skills | No | None | Skills injected into agent context at startup |
| mcpServers | No | None | MCP servers available to this agent |
| hooks | No | None | Lifecycle hooks scoped to this agent |
| memory | No | None | Persistent memory scope for cross-session learning |

### Field Details

**tools**: When specified, the agent can only use listed tools. When omitted, the agent inherits all tools from the parent. Mutually exclusive with disallowedTools.

**disallowedTools**: Denylist approach. The agent inherits all tools except those listed. Mutually exclusive with tools.

**skills**: Full skill content is injected into the agent context, not just made available for invocation. Agents do not inherit skills from the parent. Each skill listed must exist in `.claude/skills/`.

**mcpServers**: Either a server name reference (matching a key in `.mcp.json`) or an inline server definition with command and args.

**hooks**: Supports PreToolUse, PostToolUse, and SubagentStop events scoped to this agent. See @hooks-system.md for configuration format.

## Task(agent_type) Restrictions

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

### Team Agents (8) - Experimental

Agents for Claude Code Agent Teams (v2.1.32+, requires CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1):

| Agent | Model | Phase | Mode | Purpose |
|-------|-------|-------|------|---------|
| team-researcher | haiku | plan | plan (read-only) | Codebase exploration and research |
| team-analyst | inherit | plan | plan (read-only) | Requirements analysis |
| team-architect | inherit | plan | plan (read-only) | Technical design |
| team-backend-dev | inherit | run | acceptEdits | Server-side implementation |
| team-designer | inherit | run | acceptEdits | UI/UX design with Pencil/Figma MCP (requires Pencil MCP server) |
| team-frontend-dev | inherit | run | acceptEdits | Client-side implementation |
| team-tester | inherit | run | acceptEdits | Test creation with exclusive test file ownership |
| team-quality | inherit | run | plan (read-only) | TRUST 5 quality validation |

## Rules

- Write agent definitions in English
- Define expertise domain clearly in description
- Minimize tool permissions (least privilege)
- Include relevant trigger keywords
- Use permissionMode: plan for read-only agents
- Preload skills for domain expertise instead of relying on runtime loading

## Tool Permissions

Recommended tool sets by category:

Manager agents: Read, Write, Edit, Grep, Glob, Bash, Task, TaskCreate, TaskUpdate

Expert agents: Read, Write, Edit, Grep, Glob, Bash

Builder agents: Read, Write, Edit, Grep, Glob

Team implementation agents: Read, Write, Edit, Grep, Glob, Bash (+ skills preloading for domain expertise)

Team research agents: Read, Grep, Glob, Bash (read-only via permissionMode: plan)

Notes:
- Use `skills` field to preload domain-specific knowledge into team agents
- Team agents with permissionMode: plan cannot write files regardless of tools listed
- Prefer skills preloading over large tool lists for domain expertise

## Agent Invocation

Invoke agents via Task tool:

- "Use the expert-backend subagent to implement the API"
- Task tool with subagent_type parameter

For team mode invocation:
- TeamCreate to initialize team structure
- Task() with team_name and name parameters to spawn teammates
- SendMessage for inter-teammate coordination
- TeamDelete after all teammates shut down
- See team-plan.md and team-run.md for complete workflow examples

## MoAI Integration

- Use builder-agent subagent for creation
- Skill("moai-foundation-claude") for patterns
- Follow skill-authoring.md for YAML schema
- See @hooks-system.md for agent hook configuration
