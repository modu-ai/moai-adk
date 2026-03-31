# Research: Agent Teams Dynamic Generation Architecture

## Research Sources

- [Official Docs: Orchestrate teams of Claude Code sessions](https://code.claude.com/docs/en/agent-teams)
- [GitHub Issue #24316: Allow custom .claude/agents/ definitions as teammates](https://github.com/anthropics/claude-code/issues/24316)
- [GitHub Issue #31977: In-process teammates lack Agent tool](https://github.com/anthropics/claude-code/issues/31977)
- [GitHub Issue #23506: Custom agents can't spawn into teams](https://github.com/anthropics/claude-code/issues/23506)
- [TeammateTool System Prompts](https://github.com/Piebald-AI/claude-code-system-prompts)

## Key Findings

### 1. Agent Teams Architecture (Official)

Agent Teams use two mechanisms:
- **High-Level (Official)**: Natural language → Claude spawns general-purpose teammates
- **Low-Level (Programmatic)**: TeamCreate → Agent(team_name, name) → SendMessage → TeamDelete

MoAI uses the programmatic approach, which is valid but not the documented primary path.

### 2. Agent tool Parameters for Teams

The Agent tool supports:
- `subagent_type`: References .claude/agents/ OR built-in types (general-purpose, Explore, Plan)
- `team_name`: Links agent to team (requires TeamCreate first)
- `name`: Makes agent addressable via SendMessage
- `model`: Runtime override (sonnet, opus, haiku)
- `mode`: Permission override (plan, acceptEdits, etc.)
- `isolation`: "worktree" for file isolation

Key insight: `model`, `mode`, `isolation` can ALL be overridden at spawn time without agent definition files.

### 3. What Agent Definitions Provide That Runtime Can't

| Feature | In Definition | Runtime Override? |
|---------|--------------|-------------------|
| tools (allowlist) | Yes | NO |
| skills (preload) | Yes | NO (but Skill() tool available) |
| hooks (agent-scoped) | Yes | NO |
| model | Yes | YES (Agent model param) |
| permissionMode | Yes | YES (Agent mode param) |
| isolation | Yes | YES (Agent isolation param) |

### 4. Impact Analysis of Removing team-* Files

- **tools**: general-purpose inherits ALL tools → mode: "plan" enforces read-only
- **skills**: Teammates can self-load via Skill() tool or receive in prompt
- **hooks**: Use global TeammateIdle/TaskCompleted hooks + prompt instructions
- **memory**: Team agents are ephemeral → memory field unused in practice

### 5. CG Mode Compatibility

| Mode | Team Mode Support | Mechanism |
|------|------------------|-----------|
| cc | Full | All Claude API |
| glm | Full | All GLM via env override |
| cg | split-pane only | tmux session env isolation (unverified) |

## Conclusion

Static team-* agent definitions provide marginal value over dynamic generation via Agent() runtime parameters. The only real losses (agent-scoped hooks, skill preloading) have practical workarounds.
