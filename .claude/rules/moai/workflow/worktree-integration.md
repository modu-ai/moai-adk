---
paths: "**/.claude/agents/**,**/.claude/worktrees/**,**/.claude/teams/**"
---

# Worktree Integration Guide

Integration guide for MoAI Worktree and Claude Code Native Worktree systems.

## Overview

MoAI-ADK supports two complementary worktree systems for isolated development:

**L1 — Claude Code Native Worktree** (`.claude/worktrees/`):
- Ephemeral, session-scoped isolation
- Automatic cleanup when session ends
- Used for subagent isolation via `isolation: "worktree"` in agent definitions (v2.1.49+)
- CLI access: `claude --worktree` or `claude -w` (user-level flag)
- **Decision authority**: Claude Code runtime — MoAI orchestrator does not mandate L1 isolation

**L2 — MoAI SPEC Worktree** (`~/.moai/worktrees/{ProjectName}/`):
- Persistent, SPEC-scoped workspaces in global home directory
- Managed via `moai worktree` CLI commands (`moai worktree new <SPEC-ID>`)
- Used for multi-session SPEC development and team collaboration
- **Activation**: User opt-in only (via `moai worktree new` or L3 `--worktree` flag)

## Terminology Glossary

| Layer | What | Owner | When Used |
|-------|------|-------|-----------|
| **L1** | Claude Code Native Worktree (ephemeral, triggered by `Agent(isolation: "worktree")` frontmatter) | Claude Code runtime | Per-call decision by Claude Code runtime; MoAI orchestrator does not mandate. Path: `.claude/worktrees/<auto-name>/`. |
| **L2** | SPEC Worktree (`moai worktree new <SPEC-ID>`) | moai CLI (`internal/cli/worktree/`) | User opt-in for SPEC isolation. Path: `~/.moai/worktrees/{project}/{SPEC-ID}/`. |
| **L3** | Plan Worktree (`--worktree` flag at plan/run time) | moai workflow skill (delegates to L2 creation mechanism) | User opt-in at plan/run time. Wraps L2 lifecycle. |
| **git worktree** | Underlying git mechanism (`git worktree add/list/remove`) | git itself | Low-level descriptions only — DO NOT use this term ambiguously when L1/L2/L3 is appropriate. |

> Per user policy 2026-05-17 (`feedback_worktree_autonomous` memory): L2/L3 are **user opt-in**. L1 is **Claude Code runtime autonomous** — MoAI orchestrator does not mandate `isolation: "worktree"`. Default SPEC development uses main checkout + feature branch.

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

### L1 `isolation: worktree` in Agent Frontmatter

For agents that need isolated execution (v2.1.49+). L1 isolation decision is made by Claude Code runtime — MoAI orchestrator does not mandate it:

```yaml
---
name: my-implementer
isolation: worktree   # L1: Agent runs in its own isolated worktree (runtime decision)
background: true      # Agent runs without blocking main conversation
---
```

When L1 `isolation: worktree` may be beneficial:
- Implementation teammates that write files (role_profiles: implementer, tester, designer)
- Prevents file conflicts between parallel teammates
- Each agent gets its own clean L1 worktree at `.claude/worktrees/<auto-name>/`

When L1 `isolation: worktree` is typically NOT needed:
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
  YES → L1: Agent(isolation: "worktree") MAY be used; Claude Code runtime decides
        Do NOT mandate isolation for read-only agents
  NO ↓

Is this a multi-session SPEC development AND user opted in?
  YES → L2: moai worktree new SPEC-XXX (user opt-in)
  NO ↓

Is this a user-initiated parallel session?
  YES → L1: claude --worktree (-w) for user-initiated isolation
  NO ↓

Is this a one-shot sub-agent task?
  YES → L1: Agent(isolation: "worktree") if agent writes files (Claude Code runtime decides)
        Agent() without isolation if agent is read-only
  NO → No L1/L2/L3 worktree needed (use main checkout + feature branch)
```

### Advisory Rules (2026-05-17 Policy)

Per user policy 2026-05-17 (`feedback_worktree_autonomous` memory), L1 isolation is Claude Code runtime autonomous — MoAI orchestrator does not mandate it:

- [SHOULD] Implementation teammates in team mode (role_profiles: implementer, tester, designer) may benefit from L1 `isolation: "worktree"` when spawned via `Agent()`; Claude Code runtime decides per-call.
- [SHOULD] Read-only teammates (role_profiles: researcher, analyst, reviewer) typically do not benefit from L1 `isolation: "worktree"` — their `mode: "plan"` already prevents writes.
- [SHOULD] One-shot sub-agents that write files (expert-backend, expert-frontend, manager-develop) may benefit from L1 `Agent(isolation: "worktree")` for cross-file changes; Claude Code runtime decides.
- [SHOULD] GitHub workflow agents (fixer agents in `/moai github issues`) may use L1 `Agent(isolation: "worktree")` for branch isolation; Claude Code runtime decides.

### When to Use Which

### Use L1 `claude --worktree` (`-w`) for:

- **User-initiated isolation**: Starting a fresh session for exploratory work
- **Parallel sessions**: Running multiple independent Claude sessions on same repo
- **Quick experiments**: Testing code changes without affecting main workspace

### Use L1 `Agent(isolation: "worktree")` for (when runtime decides to isolate):

- **Parallel team agents**: Multiple implementation teammates working simultaneously
- **File conflict prevention**: Agents that write to different file patterns
- **One-shot sub-agents**: Sub-agents making cross-file modifications
- **GitHub issue fixing**: Each issue gets isolated L1 worktree for branch safety

> Note: MoAI orchestrator does not mandate L1 isolation — Claude Code runtime decides per-call.

### Use L2 MoAI SPEC Worktree (`moai worktree new`) for (user opt-in):

- **SPEC implementation**: Multi-session development of a feature (user opt-in only)
- **PR development**: Complete feature branches with commits
- **Persistent workspaces**: Work that spans multiple Claude sessions

## Integration Pattern (Hybrid Approach)

The recommended workflow (default is main checkout + feature branch per 2026-05-17 policy):

```
PLAN PHASE (default: main checkout + plan/SPEC-XXX branch)
  L1 Claude Native (-w): Optional quick exploration, ephemeral, no persistence
  Team researchers: No L1 worktree (read-only, permissionMode: plan)

RUN PHASE (default: main checkout + feat/SPEC-XXX branch)
  L2 MoAI SPEC Worktree: User opt-in via moai worktree new — persistent state
  Team write agents: L1 Agent(isolation: "worktree") per Claude Code runtime decision
  Team read agents: No L1 worktree (quality validation, analysis)

SYNC PHASE (default: same branch as run phase)
  L2 MoAI SPEC Worktree: PR creation from persistent workspace (if L2 was used)
```

## Agent Configuration by Role

### Implementation Agents (L1 isolation: worktree + background: true)

```yaml
# Implementation teammates (role_profiles: implementer, tester, designer)
# Spawned via: Agent(subagent_type: "general-purpose", mode: "acceptEdits", isolation: "worktree")
# Note: Claude Code runtime decides whether to materialize an L1 worktree per-call.
isolation: worktree   # L1 isolated worktree per agent (runtime decision)
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

## Prompt Path Rules for L1 Worktree-Isolated Agents

When the orchestrator generates prompts for agents spawned with L1 `isolation: "worktree"`, paths in the prompt determine where the agent operates. Incorrect paths bypass L1 worktree isolation entirely.

### HARD Rules (Path Handling)

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

Per-step worktree applicability is governed by `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Phase Discipline (canonical source). This table summarizes the mapping for quick reference; on conflict, spec-workflow.md wins.

> Per user policy 2026-05-17, L2 worktree (Steps 2-4) is **user opt-in** — default flow uses main checkout + feature branch. See `feedback_worktree_autonomous` memory.

| Step | Phase   | L2 Worktree?                              | Location                              | Lifecycle event              |
|------|---------|-------------------------------------------|---------------------------------------|------------------------------|
| 1    | Plan    | **NO** (main checkout)                    | n/a — `plan/SPEC-XXX` branch on main  | plan PR merged               |
| 2    | Run     | **User opt-in** (L2 SPEC worktree)        | `~/.moai/worktrees/{project}/{SPEC}/` | run PR merged                |
| 3    | Sync    | **User opt-in** — same as Step 2 if used | same L2 path as Step 2 (do NOT recreate) | sync PR merged            |
| 4    | Cleanup | n/a (only if L2 was created)              | host checkout                         | `moai worktree done SPEC-XXX` |

Disposal contract: When L2 worktree was created, `moai worktree done SPEC-XXX` SHOULD run only after BOTH run PR AND sync PR are merged. Premature disposal between Step 2 merge and Step 3 merge breaks Sync.

## Team Protocol

Shared protocol for all MoAI Agent Teams teammates. Supplements role-specific instructions.

### Team Discovery
- Read `~/.claude/teams/{team-name}/config.json` to discover teammates
- Always refer to teammates by their `name` field when using SendMessage

### Communication
- Use direct messages (type: "message") by default
- NEVER broadcast unless a critical blocking issue affects ALL teammates
- Send findings and results to the team lead via SendMessage when complete
- Report blockers to the team lead immediately
- Update task status via TaskUpdate

### Task Management
After completing each task:
- Mark task as completed via TaskUpdate (MANDATORY - prevents infinite waiting)
- Check TaskList for available unblocked tasks
- Claim the next available unblocked task (prefer lowest ID first) or wait for team lead instructions

### Error Recovery
- If you encounter an error, do NOT stop working. Try an alternative approach first
- If the error persists after 3 attempts, report it to the team lead via SendMessage with the error details, file path, and what you tried
- Continue with remaining tasks even if one task fails
- If blocked by another teammate's work, report the blocker and move to the next unblocked task

### Shutdown Handling
When you receive a shutdown_request JSON message:
- If all work is complete: SendMessage(type: "shutdown_response", request_id: "<from message>", approve: true)
- If work is in progress: SendMessage(type: "shutdown_response", request_id: "<from message>", approve: false, content: "Still working on [task]")

### Idle States
- Going idle is NORMAL - it means you are waiting for input from the team lead
- After completing work, you will go idle while waiting for the next assignment
- The team lead will either send new work or a shutdown request
- NEVER assume work is done until you receive shutdown_request from the lead

### Context Isolation
- You do NOT have access to the team lead's conversation history
- All necessary context must come from your spawn prompt or teammate messages
- If context is insufficient, ask the team lead for clarification via SendMessage

---

Version: 4.0.0 (Team Protocol merged from team-protocol.md via SPEC-V3R2-CON-003 OP-3)
Source: SPEC-WORKTREE-001, SPEC-TEAM-PROTOCOL-001
