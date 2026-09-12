---
name: moai-workflow-worktree
description: >
  Git worktree management for parallel SPEC development with isolated workspaces,
  automatic branch registration, and seamless MoAI-ADK integration. Use when
  setting up parallel development environments.

when_to_use: >
  Use for git worktree management: parallel SPEC development with isolated
  workspaces, automatic branch registration, branch isolation, and
  seamless MoAI-ADK integration for multiple concurrent SPECs.

license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Write, Grep, Glob
user-invocable: false
metadata:
  version: "1.1.0"
  category: "workflow"
  status: "active"
  updated: "2026-07-10"
  modularized: "true"
  tags: "git, worktree, parallel, development, spec, isolation"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000
---

# MoAI Worktree Management

Git worktree management system for parallel SPEC development with isolated workspaces, automatic registration, and seamless MoAI-ADK integration.

Core Philosophy: Each SPEC deserves its own isolated workspace to enable true parallel development without context switching overhead.

## Quick Reference (30 seconds)

What is MoAI Worktree Management?
A specialized Git worktree system that creates isolated development environments for each SPEC, enabling parallel development without conflicts.

Key Features:
- Isolated Workspaces: Each SPEC gets its own worktree with independent Git state
- Automatic Registration: Worktree registry tracks all active workspaces
- Parallel Development: Multiple SPECs can be developed simultaneously
- Seamless Integration: Works with /moai plan, /moai run, /moai sync workflow
- Smart Synchronization: Automatic sync with base branch when needed
- Cleanup Automation: Automatic cleanup of merged worktrees

Quick Access:
- CLI commands: Refer to Worktree Commands Module at modules/worktree-commands.md
- Management patterns: Refer to Worktree Management Module at modules/worktree-management.md
- Parallel workflow: Refer to Parallel Development Module at modules/parallel-development.md
- Integration guide: Refer to Integration Patterns Module at modules/integration-patterns.md
- Troubleshooting: Refer to Troubleshooting Module at modules/troubleshooting.md

Use Cases:
- Multiple SPECs development in parallel
- Isolated testing environments
- Feature branch isolation
- Code review workflows
- Experimental feature development

---

## Implementation Guide (5 minutes)

### 1. Core Architecture - Worktree Management System

Purpose: Create isolated Git worktrees for parallel SPEC development.

Key Components:

1. Worktree Registry - Central registry tracking all worktrees
2. Manager Layer - Core worktree operations including create, switch, remove, and sync
3. CLI Interface - User-friendly command interface
4. Models - Data structures for worktree metadata
5. Integration Layer - MoAI-ADK workflow integration

Registry Structure:

The registry file stores worktree metadata in JSON format. Each worktree entry contains an identifier, file path, branch name, creation timestamp, last sync time, status (active or merged), and base branch reference. The config section defines the worktree root directory, auto-sync preference, and cleanup behavior for merged branches.

File Structure:

The worktree system creates a dedicated directory structure in the user's global home directory. At the worktree root (~/.moai/worktrees/{ProjectName}/), you will find the central registry JSON file and individual directories for each SPEC. Each SPEC directory contains a .git file for worktree metadata and a complete copy of all project files.

Detailed Reference: Refer to Worktree Management Module at modules/worktree-management.md

---

### 2. CLI Commands - Complete Command Interface

Purpose: Map each worktree task to the command that actually performs it.

Entering a worktree is the launcher's job — `moai worktree` has no creation verb:

- `moai cc -w <name>` — work inside the worktree
- `moai cc -w <name> --spawn` — open it in a new tmux window, keeping this session

For inspection, use git directly: `git worktree list`.

Management commands, as `moai worktree --help` lists them:

| Command | Purpose |
|---|---|
| `sync [branch-name]` | Sync worktree with base branch |
| `remove [path]` | Remove a worktree |
| `clean` | Clean stale worktree references |
| `recover` | Repair worktree registry |
| `done [branch-name]` | Complete worktree and cleanup |
| `snapshot` | Capture working tree state snapshot for guard verification |
| `verify` | Verify working tree state against snapshot + check agent response |
| `restore` | Restore working tree to a snapshot's HEAD state |

This table is a copy of the help output. `moai worktree --help` is the authoritative list; where the two differ, the help output wins.

Detailed Reference: Refer to Worktree Commands Module at modules/worktree-commands.md. Its command examples predate the current command set; where they differ from the table above, the table and `moai worktree --help` win.

---

### 3. Parallel Development Workflow - Isolated SPEC Development

Purpose: Enable true parallel development without context switching.

Workflow Integration:

During the Plan Phase using /moai plan, the SPEC is created, and the launcher (`moai cc -w <name>`) enters an isolated worktree for it.

During the Development Phase, the isolated worktree environment provides independent Git state with zero context switching overhead.

During the Sync Phase using /moai sync, the worktree sync command ensures clean integration with conflict resolution support.

During the Cleanup Phase, the worktree clean command provides automatic cleanup with registry maintenance.

Parallel Development Benefits:

1. Context Isolation: Each SPEC has its own Git state, files, and environment
2. Zero Switching Cost: Instant switching between worktrees
3. Independent Development: Work on multiple SPECs simultaneously
4. Safe Experimentation: Isolated environment for experimental features
5. Clean Integration: Automatic sync and conflict resolution

Example Workflow:

First, create a worktree for SPEC-001 with a description like "User Authentication" and switch to that directory. Then run /moai run SPEC-001 to develop in isolation. Next, navigate back to the main repository and create another worktree for SPEC-002 with description "Payment Integration". Switch to that worktree and run /moai run SPEC-002 for parallel development. When needed, switch between worktrees and continue development. Finally, sync both worktrees when ready for integration.

Detailed Reference: Refer to Parallel Development Module at modules/parallel-development.md

---

### 4. Integration Patterns - MoAI-ADK Workflow Integration

Purpose: Seamless integration with MoAI-ADK Plan-Run-Sync workflow.

Integration Points:

During Plan Phase Integration with /moai plan, after SPEC creation, enter the worktree through the launcher: `moai cc -w <name>` works inside it, and `moai cc -w <name> --spawn` opens it in a new tmux window.

During Development Phase with /moai run, worktree isolation provides a clean development environment with independent Git state preventing conflicts and automatic registry tracking.

During Sync Phase with /moai sync, before PR creation run the sync command for the SPEC. After PR merge, run the clean command with the merged-only flag to remove completed worktrees.

Auto-Detection Patterns:

The system detects worktree environments by checking for the registry file in the parent directory. When detected, the SPEC ID is extracted from the current directory name.

Configuration Integration:

The MoAI configuration supports worktree settings including auto_create for automatic worktree creation, auto_sync for automatic synchronization, cleanup_merged for automatic cleanup of merged branches, and worktree_root for specifying the worktree directory location with project name substitution.

Detailed Reference: Refer to Integration Patterns Module at modules/integration-patterns.md

---

### 5. `--spawn` — Launch a Teammate Session in a New tmux Window

Purpose: start a Claude or GLM session in a worktree **without giving up the session you are in**.

The launch commands (`moai cc`, `moai glm`, `moai cg`) normally replace the running shell, which is right for "work here now" but cannot express "keep going and start a teammate alongside me". `--spawn` re-issues the same command in a new tmux window instead, then returns so the caller keeps working.

Combined with `-w <name>`, one command opens a teammate in an isolated worktree:

```bash
moai cg -w feat-auth --spawn    # GLM teammate in .claude/worktrees/feat-auth
moai cc -w feat-auth --spawn    # Claude teammate, same worktree
moai glm -w feat-auth --spawn   # all-GLM teammate
```

Behavior:

- The new window is created detached, so focus stays in the caller's pane. The printed pane id (e.g. `%7`) is the handle for switching to it.
- The spawned window starts at the project root, so a short `-w <name>` value resolves against `.claude/worktrees/<name>/`.
- `--spawn` is consumed by MoAI and never reaches Claude Code. Tokens after the `--` pass-through marker are left untouched.
- Arguments are shell-quoted, so a worktree name containing spaces or shell metacharacters reaches the spawned process intact.

Requirements — each is refused with a clear error rather than a silent fallback, because falling back would replace the caller's session, the exact outcome `--spawn` exists to avoid:

| Missing | Message |
|---------|---------|
| `$TMUX` (not inside a session) | `tmux session required for --spawn` |
| `tmux` binary | `--spawn requires the tmux binary` |
| `moai` binary in `PATH` | `--spawn needs the moai binary in PATH` |

No settings are mutated before these checks run, so a refusal leaves the environment untouched. The spawned command performs its own backend setup inside the new window.

Platform note: tmux is POSIX-only, so `--spawn` is unavailable on Windows and reports the missing binary. Entering a worktree in place with `-w` works on every platform.

Detailed Reference: the launcher's spawn entry point — flag stripping, command reconstruction with shell quoting, and the tmux window invocation.

---

## Advanced Implementation (10+ minutes)

### Synchronization Strategies

`moai worktree sync` merges the base branch into the worktree by default; `--strategy rebase` rebases the worktree onto the base instead, and `--base` selects the base branch. When a sync stops on a conflict, resolve it with ordinary git inside the worktree and rerun the sync.

### Capabilities Not Provided

`moai worktree` has no shared or team registry mode, no per-developer prefixing, no selective sync patterns, no conflict auto-resolution or interactive mode, no worktree templates, and no shallow, background, parallel, or cache options. Prepare a worktree after entering it, and see `modules/worktree-commands.md` for the flags that exist.

---

## Works Well With

Commands:
- /moai plan - SPEC creation with automatic worktree setup
- /moai run - Development in isolated worktree environment
- /moai sync - Integration with automatic worktree sync
- /moai feedback - Worktree workflow improvements

Skills:
- moai-foundation-core - Parallel development patterns
- moai-workflow-project - Project management integration
- moai-workflow-spec - SPEC-driven development
- moai-ref-git-workflow - Git workflow optimization

Tools:
- Git worktree - Native Git worktree functionality
- Cobra - CLI command framework and formatted output

---

## Quick Decision Guide

For new SPEC development, use the worktree isolation pattern with auto-setup. The primary approach is worktree isolation and the supporting pattern is integration with /moai plan.

For parallel development across multiple SPECs, use one worktree and one session per SPEC. The primary approach is maintaining multiple worktrees and the supporting pattern is opening each in its own tmux window with `moai cc -w <name> --spawn`.

For team coordination, give each developer their own worktrees and branches. The primary approach is per-developer isolation and the supporting pattern is integrating through the base branch and pull requests.

For code review workflows, use isolated review worktrees. The primary approach is worktree isolation for reviews and the supporting pattern is clean sync after review completion.

For experimental features, use temporary worktrees with auto-cleanup. The primary approach is creating temporary worktrees and the supporting pattern is safe experimentation with automatic removal.

Module Deep Dives:
- Worktree Commands: Refer to modules/worktree-commands.md for complete CLI reference
- Worktree Management: Refer to modules/worktree-management.md for core architecture
- Parallel Development: Refer to modules/parallel-development.md for workflow patterns
- Integration Patterns: Refer to modules/integration-patterns.md for MoAI-ADK integration
- Troubleshooting: Refer to modules/troubleshooting.md for problem resolution

Full Examples: Refer to references/examples.md
External Resources: Refer to references/reference.md

<!-- moai:evolvable-start id="rationalizations" -->
## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "Worktree isolation is overkill for this small change" | Small changes on main cause merge conflicts when parallel work is in progress. Worktrees prevent this. |
| "I will just work on the main branch, it is faster" | Working on main blocks other agents from writing. Worktrees enable parallelism. |
| "Read-only agents need worktree isolation too, for safety" | Isolation is unnecessary only for an agent whose tools list holds no tool that can write. The spawn-time mode parameter is deprecated and ignored, so it blocks nothing. Omitting Write, Edit, and NotebookEdit is not enough: Bash writes through the shell, and so can any other file-mutating tool (another shell, an MCP tool, or an Agent that spawns a writer). While any such tool is present, read-only is not guaranteed and isolation is not wasted. |
| "I can skip worktree cleanup, git handles it" | Stale worktree branches accumulate and confuse git worktree list. Always prune after use. |
| "Absolute paths in agent prompts are fine since the worktree has the same structure" | Absolute paths to the main repo bypass worktree isolation entirely. Use relative paths. |

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="red-flags" -->
## Red Flags

- Implementation agent spawned without isolation: worktree in team mode
- Agent whose tools list holds no tool that can write, spawned with isolation: worktree (unnecessary overhead)
- Agent prompt contains absolute path to the main project directory for write targets
- Worktree not pruned after team session completes (stale branches remain)
- cd /absolute/project/path in Bash commands inside worktree-isolated agent prompts

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="verification" -->
## Verification

- [ ] Implementation teammates use isolation: worktree (check agent spawn parameters)
- [ ] Teammates spawned without isolation: worktree hold no tool that can write (check the tools list for Write, Edit, NotebookEdit, Bash, and any other file-mutating tool; with any present, read-only is not guaranteed)
- [ ] Agent prompts reference write-target files by relative paths only
- [ ] `git worktree list` shows no stale worktrees after session ends
- [ ] Worktree CWD isolation verified on Claude Code >= 2.1.97 (check version)
- [ ] Hook scripts (handle-worktree-create.sh, handle-worktree-remove.sh) are present and executable

<!-- moai:evolvable-end -->
