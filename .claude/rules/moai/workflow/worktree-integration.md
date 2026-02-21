# Worktree Integration Guide

Integration guide for MoAI Worktree and Claude Code Native Worktree systems.

## Overview

MoAI-ADK supports two complementary worktree systems for isolated development:

**Claude Code Native Worktree** (`.claude/worktrees/`):
- Ephemeral, session-scoped isolation
- Automatic cleanup when session ends
- Used for subagent isolation via `isolation: worktree` in agent definitions

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
| **CLI** | `claude -w` (automatic) | `moai worktree new/list/remove` |
| **Cleanup** | Automatic on session end | Manual via `moai worktree remove` |
| **Branch Strategy** | Temporary branches | Feature branches linked to SPEC |
| **Team Use** | Single agent isolation | Multi-developer collaboration |
| **State Persistence** | None | SPEC state, progress tracking |
| **Hook Support** | Standard Claude hooks | WorktreeCreate/Remove hooks |

## When to Use Which

### Use Claude Native Worktree for:

- **Quick experiments**: Testing code changes without affecting main workspace
- **Debugging sessions**: Isolated environment for troubleshooting
- **One-time analysis**: Read-only codebase exploration
- **Subagent isolation**: Implementation agents with `isolation: worktree`
- **Parallel agent execution**: Multiple teammates working independently

### Use MoAI Worktree for:

- **SPEC implementation**: Multi-session development of a feature
- **PR development**: Complete feature branches with commits
- **Team collaboration**: Shared workspace for team members
- **Long-running features**: Development spanning multiple sessions
- **Cross-session state**: Progress tracking and resumption

## Integration Pattern (Hybrid Approach)

The recommended workflow combines both worktree systems:

```
┌─────────────────────────────────────────────────────────────────┐
│  PLAN PHASE                                                      │
│  └── Claude Native (-w): Quick exploration, no persistence      │
├─────────────────────────────────────────────────────────────────┤
│  RUN PHASE                                                       │
│  ├── MoAI Worktree: SPEC implementation, persistent state        │
│  └── Teammates: isolation: worktree for parallel execution      │
├─────────────────────────────────────────────────────────────────┤
│  SYNC PHASE                                                      │
│  └── MoAI Worktree: PR creation from persistent workspace        │
└─────────────────────────────────────────────────────────────────┘
```

### Phase-Specific Usage

**Plan Phase:**
```bash
# Exploration uses Claude native worktree (ephemeral)
# No state needs to persist between sessions
claude -w  # Automatic for plan-mode agents
```

**Run Phase:**
```bash
# SPEC implementation uses MoAI worktree (persistent)
moai worktree new SPEC-AUTH-001
cd .moai/worktrees/MyProject/SPEC-AUTH-001

# Teammates work in their own isolated worktrees
# Each implementation agent spawns with isolation: worktree
```

**Sync Phase:**
```bash
# PR creation from MoAI worktree
cd .moai/worktrees/MyProject/SPEC-AUTH-001
gh pr create --title "feat: Add JWT authentication"

# Cleanup after merge
moai worktree remove SPEC-AUTH-001
```

## Agent Configuration

### Implementation Agents

For agents that modify code, use `isolation: worktree`:

```yaml
---
name: team-backend-dev
description: Backend implementation specialist
model: inherit
isolation: worktree         # Creates isolated worktree
background: true            # Enables parallel execution
permissionMode: acceptEdits
tools: Read, Write, Edit, Grep, Glob, Bash
---
```

### Read-Only Agents

For research and analysis agents, use `isolation: none`:

```yaml
---
name: team-researcher
description: Codebase exploration specialist
model: haiku
isolation: none             # No worktree needed (read-only)
permissionMode: plan        # Read-only mode
tools: Read, Grep, Glob, Bash
---
```

### Architecture Agents

For design and planning agents that need workspace but minimal changes:

```yaml
---
name: team-architect
description: Technical architecture specialist
model: opus
isolation: none             # Usually read-only analysis
permissionMode: plan
tools: Read, Grep, Glob, Bash
---
```

## Hook Integration

### Worktree Lifecycle Hooks

MoAI-ADK provides hook templates for worktree lifecycle events:

**WorktreeCreate Hook** (`.claude/hooks/moai/handle-worktree-create.sh`):
- Triggered when MoAI worktree is created
- Use for: Setting up environment, installing dependencies
- Template location: `internal/template/templates/.claude/hooks/moai/`

**WorktreeRemove Hook** (`.claude/hooks/moai/handle-worktree-remove.sh`):
- Triggered when MoAI worktree is removed
- Use for: Cleanup, finalizing commits, notifications
- Template location: `internal/template/templates/.claude/hooks/moai/`

### Hook Configuration

Add worktree hooks to `.claude/settings.json`:

```json
{
  "hooks": {
    "SessionStart": [{
      "command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-session-start.sh\"",
      "timeout": 5
    }],
    "PostToolUse": [{
      "matcher": "Write|Edit",
      "command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-post-tool.sh\"",
      "timeout": 60
    }]
  }
}
```

### Teammate Isolation Hooks

Teammates with `isolation: worktree` trigger these hooks automatically:

| Hook Event | When Triggered | Purpose |
|------------|---------------|---------|
| SubagentStart | Teammate spawns | Log worktree creation |
| TeammateIdle | Teammate goes idle | Validate worktree state |
| SubagentStop | Teammate terminates | Cleanup temporary worktree |

## Worktree Commands Reference

### MoAI CLI Commands

```bash
# Create new worktree for SPEC
moai worktree new SPEC-XXX

# List all worktrees
moai worktree list

# Remove worktree
moai worktree remove SPEC-XXX

# Show worktree status
moai worktree status SPEC-XXX
```

### Claude Code Native Commands

```bash
# Start Claude in isolated worktree (automatic for agents)
claude -w

# Check current worktree status
claude status
```

## Best Practices

### Worktree Naming

- MoAI worktrees: Use SPEC-ID format (`SPEC-AUTH-001`)
- Claude native: Auto-generated (no manual naming)

### Cleanup Strategy

1. **Claude Native**: Automatic cleanup on session end
2. **MoAI Worktrees**: Clean up after PR merge or SPEC completion

### Branch Management

```
main
├── feature/SPEC-AUTH-001 (MoAI worktree)
│   └── commits from implementation
├── feature/SPEC-USER-002 (MoAI worktree)
│   └── commits from implementation
└── temporary/agent-xxx (Claude native, auto-deleted)
```

### Conflict Prevention

- Each teammate works in isolated worktree
- File ownership prevents write conflicts
- MoAI worktrees are SPEC-specific (one SPEC per worktree)

## Troubleshooting

### Common Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| Worktree not found | Removed manually | Run `moai worktree list` to verify |
| Agent worktree conflicts | Multiple agents same file | Check file ownership in team config |
| Stale worktree branches | Incomplete cleanup | Run `git worktree prune` |
| Hooks not firing | Missing wrapper script | Check `.claude/hooks/moai/` directory |

### Recovery Commands

```bash
# Prune stale git worktrees
git worktree prune

# List all git worktrees
git worktree list

# Force remove MoAI worktree
moai worktree remove SPEC-XXX --force
```

## Integration with SPEC Workflow

For complete SPEC workflow documentation, see @spec-workflow.md.

### SPEC-to-Worktree Mapping

| SPEC Phase | Worktree Type | Location |
|------------|--------------|----------|
| Plan | Claude Native | `.claude/worktrees/` (ephemeral) |
| Run | MoAI | `.moai/worktrees/{Project}/{SPEC}/` |
| Sync | MoAI | Same as Run phase |

### State Persistence

- **Plan phase**: No state persistence needed
- **Run phase**: SPEC progress tracked in `.moai/worktrees/{Project}/{SPEC}/state.yaml`
- **Sync phase**: Final state before PR creation

---

Version: 1.0.0
Source: SPEC-WORKTREE-001
