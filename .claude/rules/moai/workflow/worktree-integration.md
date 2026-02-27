---
paths: "**/.claude/agents/**,**/.moai/worktrees/**"
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

**MoAI Worktree** (`.moai/worktrees/`):
- Persistent, SPEC-scoped workspaces
- Managed via `moai worktree` CLI commands
- Used for multi-session SPEC development and team collaboration

## Comparison Table

| Feature | Claude Native | MoAI |
|---------|--------------|------|
| **Path** | `.claude/worktrees/<name>/` | `.moai/worktrees/{Project}/{SPEC}/` |
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
name: team-backend-dev
isolation: worktree   # Agent runs in its own isolated worktree
background: true      # Agent runs without blocking main conversation
---
```

When to use `isolation: worktree`:
- Implementation agents that write files (team-backend-dev, team-frontend-dev, team-tester, team-designer)
- Prevents file conflicts between parallel teammates
- Each agent gets its own clean worktree at `.claude/worktrees/<auto-name>/`

When NOT to use `isolation: worktree`:
- Read-only agents (team-researcher, team-analyst, team-architect, team-quality)
- `permissionMode: plan` already prevents writes; adding isolation adds overhead without benefit

### `background: true` in Agent Frontmatter

Run agent without blocking the main conversation (v2.1.46+):

```yaml
---
name: team-backend-dev
background: true   # Returns immediately; results delivered on next turn
---
```

Use with `isolation: worktree` for optimal parallel execution in team mode.

Kill background agent: Press `Ctrl+F` in Claude Code interface.

## Worktree Selection Rules [HARD]

### Decision Tree

```
Is this a team mode implementation with parallel agents?
  YES → Use Task(isolation: "worktree") for write agents
        Do NOT use isolation for read-only agents
  NO ↓

Is this a multi-session SPEC development?
  YES → Use MoAI Worktree (moai worktree new SPEC-XXX)
  NO ↓

Is this a user-initiated parallel session?
  YES → Use claude --worktree (-w)
  NO ↓

Is this a one-shot sub-agent task?
  YES → Use Task(isolation: "worktree") if agent writes files
        Use Task() without isolation if agent is read-only
  NO → No worktree needed
```

### HARD Rules

- [HARD] Implementation agents in team mode (team-backend-dev, team-frontend-dev, team-tester, team-designer) MUST use `isolation: "worktree"` when spawned via Task()
- [HARD] Read-only agents (team-researcher, team-analyst, team-architect, team-quality) MUST NOT use `isolation: "worktree"` — their `permissionMode: plan` already prevents writes
- [HARD] One-shot sub-agents that write files (expert-backend, expert-frontend, manager-ddd, manager-tdd) SHOULD use `isolation: "worktree"` when making cross-file changes
- [HARD] GitHub workflow agents (fixer agents in /moai github issues) MUST use `isolation: "worktree"` for branch isolation

### When to Use Which

### Use `claude --worktree` (`-w`) for:

- **User-initiated isolation**: Starting a fresh session for exploratory work
- **Parallel sessions**: Running multiple independent Claude sessions on same repo
- **Quick experiments**: Testing code changes without affecting main workspace

### Use `Task(isolation: "worktree")` for:

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
  Team write agents: Task(isolation: "worktree") for parallel execution
  Team read agents: No worktree (quality validation, analysis)

SYNC PHASE
  MoAI Worktree: PR creation from persistent workspace
```

## Agent Configuration by Role

### Implementation Agents (isolation: worktree + background: true)

```yaml
# team-backend-dev, team-frontend-dev, team-tester, team-designer
isolation: worktree   # Isolated worktree per agent
background: true      # Non-blocking parallel execution
permissionMode: acceptEdits
```

### Research/Analysis Agents (no isolation needed)

```yaml
# team-researcher, team-analyst, team-architect, team-quality
# No isolation: worktree (read-only, permissionMode: plan prevents writes)
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
| Run | MoAI | `.moai/worktrees/{Project}/{SPEC}/` |
| Sync | MoAI | Same as Run phase |

---

Version: 3.0.0 (HARD Rules + Decision Tree)
Source: SPEC-WORKTREE-001
