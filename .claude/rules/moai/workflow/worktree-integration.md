---
description: Worktree integration guide with path isolation rules for agents using isolation worktree
globs: "**/.claude/agents/**,**/.claude/worktrees/**"
---

# Worktree Integration Guide

Integration guide for MoAI Worktree and Claude Code Native Worktree systems.

## Overview

MoAI-ADK supports two complementary worktree systems for isolated development:

**Claude Code Native Worktree** (`.claude/worktrees/`):
- Ephemeral, session-scoped isolation
- Automatic cleanup when session ends
- Used for subagent isolation via `isolation: worktree` in agent definitions (v2.1.49+)
- CLI access: `claude --worktree` or `claude -w` (user-level flag)

**MoAI Worktree** (`~/.moai/worktrees/{ProjectName}/`):
- Persistent, SPEC-scoped workspaces in global home directory
- Managed via `moai worktree` CLI commands
- Used for multi-session SPEC development and team collaboration

## Comparison Table

| Feature | Claude Native | MoAI |
|---------|--------------|------|
| **Path** | `.claude/worktrees/<name>/` | `~/.moai/worktrees/{Project}/{SPEC}/` |
| **Lifetime** | Ephemeral (session-scoped) | Persistent (SPEC-scoped) |
| **Purpose** | Session isolation for subagents | SPEC development, PR creation |
| **CLI** | `claude -w` (user) or `isolation: worktree` (agent) | `moai worktree new/list/remove` |
| **Cleanup** | Automatic on session end | Manual via `moai worktree remove` |
| **Branch Strategy** | Temporary branches | Feature branches linked to SPEC |
| **Team Use** | Single agent isolation | Multi-developer collaboration |
| **State Persistence** | None | SPEC state, progress tracking |
| **Hook Support** | WorktreeCreate/WorktreeRemove hooks | WorktreeCreate/WorktreeRemove hooks |

## Claude Code 2.1.50+ Worktree Features

### `claude --worktree` (`-w`) Flag

For users starting isolated sessions:

```bash
# Start new isolated session in worktree
claude --worktree

# With custom name
claude --worktree my-feature

# With tmux for split-pane display (tmux or iTerm2 required)
claude --worktree --tmux
```

Behavior:
- Creates `.claude/worktrees/<name>/` automatically
- Branches from default remote branch
- On session end: prompts to keep (with commits) or auto-deletes (no changes)

tmux flag notes:
- Requires tmux or iTerm2
- NOT supported in VS Code integrated terminal, Windows Terminal, or Ghostty
- Useful for parallel team mode where viewing multiple teammates' output is beneficial

### `isolation: worktree` in Agent Frontmatter

For agents that need isolated execution (v2.1.49+):

```yaml
---
name: my-implementer
isolation: worktree   # Agent runs in its own isolated worktree
background: true      # Agent runs without blocking main conversation
---
```

When to use `isolation: worktree`:
- Implementation teammates that write files (role_profiles: implementer, tester, designer)
- Prevents file conflicts between parallel teammates
- Each agent gets its own clean worktree at `.claude/worktrees/<auto-name>/`

When NOT to use `isolation: worktree`:
- Read-only teammates (role_profiles: researcher, analyst, reviewer)
- `permissionMode: plan` already prevents writes; adding isolation adds overhead without benefit

### `background: true` in Agent Frontmatter

Run agent without blocking the main conversation (v2.1.46+):

```yaml
---
name: team-coder
background: true   # Returns immediately; results delivered on next turn
---
```

Use with `isolation: worktree` for optimal parallel execution in team mode.

[HARD] Background agents auto-deny Write/Edit operations. Only use `background: true` for:
- Read-only research and analysis agents
- Agents whose write paths are pre-approved in settings.json `permissions.allow`

For write-heavy agents without pre-approval, use `background: false` (foreground, sequential).

Kill background agent: Press `Ctrl+X Ctrl+K` in Claude Code interface (v2.1.83+).

## Worktree Selection Rules [HARD]

### Decision Tree

```
Is this a team mode implementation with parallel agents?
  YES → Use Agent(isolation: "worktree") for write agents
        Do NOT use isolation for read-only agents
  NO ↓

Is this a multi-session SPEC development?
  YES → Use MoAI Worktree (moai worktree new SPEC-XXX)
  NO ↓

Is this a user-initiated parallel session?
  YES → Use claude --worktree (-w)
  NO ↓

Is this a one-shot sub-agent task?
  YES → Use Agent(isolation: "worktree") if agent writes files
        Use Agent() without isolation if agent is read-only
  NO → No worktree needed
```

### HARD Rules

- [HARD] Implementation teammates in team mode (role_profiles: implementer, tester, designer) MUST use `isolation: "worktree"` when spawned via Agent()
- [HARD] Read-only teammates (role_profiles: researcher, analyst, reviewer) MUST NOT use `isolation: "worktree"` — their `mode: "plan"` already prevents writes
- [HARD] One-shot sub-agents that write files (expert-backend, expert-frontend, manager-ddd, manager-tdd) SHOULD use `isolation: "worktree"` when making cross-file changes
- [HARD] GitHub workflow agents (fixer agents in /moai github issues) MUST use `isolation: "worktree"` for branch isolation

### When to Use Which

### Use `claude --worktree` (`-w`) for:

- **User-initiated isolation**: Starting a fresh session for exploratory work
- **Parallel sessions**: Running multiple independent Claude sessions on same repo
- **Quick experiments**: Testing code changes without affecting main workspace

### Use `Agent(isolation: "worktree")` for:

- **Parallel team agents**: Multiple implementation teammates working simultaneously
- **File conflict prevention**: Agents that write to different file patterns
- **One-shot sub-agents**: Sub-agents making cross-file modifications
- **GitHub issue fixing**: Each issue gets isolated worktree for branch safety

### Use MoAI Worktree (`moai worktree`) for:

- **SPEC implementation**: Multi-session development of a feature
- **PR development**: Complete feature branches with commits
- **Persistent workspaces**: Work that spans multiple Claude sessions

## Integration Pattern (Hybrid Approach)

The recommended workflow combines both worktree systems:

```
PLAN PHASE
  Claude Native (-w): Quick exploration, ephemeral, no persistence
  Team researchers: No worktree (read-only, permissionMode: plan)

RUN PHASE
  MoAI Worktree: SPEC implementation, persistent state
  Team write agents: Agent(isolation: "worktree") for parallel execution
  Team read agents: No worktree (quality validation, analysis)

SYNC PHASE
  MoAI Worktree: PR creation from persistent workspace
```

## Agent Configuration by Role

### Implementation Agents (isolation: worktree + background: true)

```yaml
# Implementation teammates (role_profiles: implementer, tester, designer)
# Spawned via: Agent(subagent_type: "general-purpose", mode: "acceptEdits", isolation: "worktree")
isolation: worktree   # Isolated worktree per agent
background: true      # Non-blocking parallel execution
permissionMode: acceptEdits
```

### Research/Analysis Agents (no isolation needed)

```yaml
# Read-only teammates (role_profiles: researcher, analyst, reviewer)
# Spawned via: Agent(subagent_type: "general-purpose", mode: "plan")
# No isolation: worktree (read-only, mode: plan prevents writes)
permissionMode: plan  # Read-only mode already provides safety
```

## WorktreeCreate and WorktreeRemove Hooks

MoAI-ADK implements hook handlers for worktree lifecycle events:

| Hook Event | Triggered When | MoAI Handler |
|-----------|---------------|--------------|
| WorktreeCreate | Agent with isolation: worktree spawns | `moai hook worktree-create` |
| WorktreeRemove | Agent with isolation: worktree terminates | `moai hook worktree-remove` |

Hook scripts are located at:
- `.claude/hooks/moai/handle-worktree-create.sh`
- `.claude/hooks/moai/handle-worktree-remove.sh`

Currently the handlers log worktree creation and removal for session tracking.

## Prompt Path Rules for Worktree-Isolated Agents

When the orchestrator generates prompts for agents spawned with `isolation: "worktree"`, paths in the prompt determine where the agent operates. Incorrect paths bypass worktree isolation entirely.

### HARD Rules

- [HARD] Do NOT include absolute paths to the main project directory in agent prompts for write-target files
- [HARD] Do NOT include `cd /absolute/project/path &&` in Bash commands within agent prompts
- [HARD] Reference write-target files by project-root-relative paths (e.g., `src/domains/auth/handler.go`) and let the agent resolve from its own CWD
- [HARD] `$CLAUDE_PROJECT_DIR` in hook commands is acceptable — Claude Code resolves this to the correct directory for the agent's context

### Path Categories

| Category | Example | Absolute Path OK? | Reason |
|----------|---------|-------------------|--------|
| Write-target files | Source code, tests | NO — use relative | Agent CWD is worktree root; relative paths resolve correctly |
| Read-only references | Skills, configs via `${CLAUDE_SKILL_DIR}` | YES | Content is identical in main repo; read-only access is safe |
| SPEC documents | `.moai/specs/SPEC-XXX/spec.md` | Relative preferred | SPEC files are copied to worktree during checkout |
| Bash commands | `go test ./...` | NO `cd` prefix | Agent CWD is already set to worktree root |

### How It Works

When `isolation: "worktree"` is set, Claude Code:
1. Creates a temporary worktree from the current branch
2. Sets the agent's CWD to the worktree root
3. The agent constructs absolute paths from its own CWD

```
Main repo:  /Users/user/project/src/auth/handler.go
Worktree:   /Users/user/project/.claude/worktrees/abc123/src/auth/handler.go
```

Both share the same project structure. `src/auth/handler.go` resolves correctly in either context.

### Anti-Pattern Examples

```
# WRONG: Absolute path in prompt bypasses worktree
"Read /Users/user/project/src/auth/handler.go and fix the bug"

# WRONG: cd to main project in Bash command
"Run: cd /Users/user/project && go test ./..."

# CORRECT: Relative path — agent resolves from its own CWD
"The bug is in src/auth/handler.go. Read the file and fix it."

# CORRECT: No cd prefix — agent CWD is already worktree root
"Run: go test ./..."
```

## Minimum Version Requirements

| Feature | Minimum Version | Notes |
|---------|----------------|-------|
| `isolation: worktree` in Agent frontmatter | 2.1.49 | Basic worktree isolation |
| `background: true` in Agent frontmatter | 2.1.46 | Non-blocking agent execution |
| `claude --worktree` user flag | 2.1.50 | User-initiated worktree sessions |
| `Ctrl+X Ctrl+K` to kill background agent | 2.1.83 | Kill stuck background agents |
| Worktree CWD isolation fix | **2.1.97** | Prior versions leaked agent CWD back to parent session |
| Stop/SubagentStop hook stability | **2.1.97** | Prior versions failed on long-running sessions |
| `moai doctor` MCP scope duplicate detection | **2.1.110** | Warns on MCP server duplication across `.mcp.json` + settings.json |
| Bash tool timeout ceiling enforcement | **2.1.110** | Maximum 600,000ms (10 min) enforced by runtime |
| `effortLevel` setting for Opus 4.7 | **2.1.110** | Supports `low`/`medium`/`high`/`xhigh`/`max` effort levels |
| `CLAUDE_ENV_FILE` on Windows | **2.1.111** | Prior versions: no-op on Windows; fixed to inject env as on macOS/Linux |
| `disableBypassPermissionsMode` policy | **2.1.111** | Prevents agents from requesting `bypassPermissions` when `true` |

**Recommended**: Claude Code **2.1.111 or later** for Opus 4.7 support, MCP doctor warnings, and Windows CLAUDE_ENV_FILE parity. Minimum baseline: **2.1.97** for worktree isolation.

## Troubleshooting

| Issue | Cause | Solution |
|-------|-------|----------|
| Worktree not found | Removed manually | Run `moai worktree list` to verify |
| Agent worktree conflicts | Multiple agents same file | Check file ownership in team config |
| Stale worktree branches | Incomplete cleanup | Run `git worktree prune` |
| Hooks not firing | Missing wrapper script | Check `.claude/hooks/moai/` directory |
| `--tmux` not working | Unsupported terminal | Use tmux or iTerm2 (not VS Code, Ghostty) |

## SPEC-to-Worktree Mapping

| SPEC Phase | Worktree Type | Location |
|------------|--------------|----------|
| Plan | Claude Native | `.claude/worktrees/` (ephemeral) |
| Run | MoAI | `~/.moai/worktrees/{Project}/{SPEC}/` |
| Sync | MoAI | Same as Run phase |

---

Version: 3.0.0 (HARD Rules + Decision Tree)
Source: SPEC-WORKTREE-001
