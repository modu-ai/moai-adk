---
paths: "**/.claude/agents/**,**/.claude/worktrees/**,**/.claude/teams/**"
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
| **Lifetime** | Ephemeral (session-scoped) | Persistent |
| **Purpose** | Session isolation for subagents | SPEC development, PR creation |
| **CLI** | `claude -w` (user) or `isolation: worktree` (agent) | `moai cc -w <name>` to enter; `moai worktree clean/done/remove` to dispose |
| **Cleanup** | Automatic on session end | Manual via `moai worktree remove` |
| **Branch Strategy** | Temporary branches | Feature branches linked to SPEC |
| **Team Use** | Single agent isolation | Multi-developer collaboration |
| **State Persistence** | None | SPEC state, progress tracking |
| **Hook Support** | WorktreeCreate/WorktreeRemove hooks | WorktreeCreate/WorktreeRemove hooks |

## Terminology Glossary

This glossary is the canonical definition surface for the L1 / L2 worktree-layer terms used across the MoAI rule set. (The former L3 "launch action" tier is retired with `/moai plan --worktree`; a worktree is now entered, not provisioned by a workflow step.) Other rules (`spec-workflow.md`, `worktree-state-guard.md`, `session-handoff.md`, and `CLAUDE.md` §14) cross-reference `§ Terminology Glossary` for these definitions.

| Layer | Name | What it is | Path / Trigger | Lifetime | Owner |
|-------|------|-----------|----------------|----------|-------|
| **L1** | Claude-native ephemeral worktree | Session-scoped isolation materialized by the Claude Code runtime for a subagent spawned with `Agent(isolation: "worktree")` (or `claude --worktree`). The runtime decides whether to materialize it. | `.claude/worktrees/<auto-name>/` | Ephemeral — auto-cleaned on session end | Claude Code runtime (autonomous; MoAI orchestrator does not mandate it per the opt-in policy) |
| **L2** | MoAI persistent SPEC worktree | A persistent, SPEC-scoped working directory entered with `moai cc -w <name>`. Used for multi-session SPEC development (run + sync phases reuse the same L2 worktree). | `~/.moai/worktrees/<project>/<SPEC>/` | Persistent — disposed only via `moai worktree done SPEC-XXX` after both run + sync PRs merge | MoAI (user-managed via `moai worktree` CLI) |

Relationships:
- An **L2** persistent worktree is created by entering one with `moai cc -w <name>` (the former `/moai plan --worktree` launch action and `moai worktree new` command are both retired).
- An **L1** ephemeral worktree is materialized autonomously by the Claude Code runtime for an isolated subagent; it is independent of L2 and may occur inside either the main checkout or an L2 worktree.
- When work happens inside an L2 worktree, the paste-ready resume MUST anchor the next session there (Block 0) per `session-handoff.md` § Worktree-Anchored Resume Pattern.

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
- Implementation teammates that write files (write-capable implementation roles: implementer / tester / designer)
- Prevents file conflicts between parallel teammates
- Each agent gets its own clean worktree at `.claude/worktrees/<auto-name>/`

When NOT to use `isolation: worktree`:
- Read-only teammates (read-only research/review roles: researcher / analyst / reviewer)
- `permissionMode: plan` already prevents writes; adding isolation adds overhead without benefit

#### L1 ephemeral vs L2 persistent — `isolation: worktree` is NOT a re-entry mechanism

`Agent(isolation: "worktree")` creates a NEW **L1 ephemeral** worktree scoped to a single subagent invocation under `.claude/worktrees/<auto-name>/`. It is categorically distinct from an **L2 persistent** SPEC worktree entered with `moai cc -w <name>`. Conflating the two produces the worktree-masked flaky failure mode documented in `worktree-state-guard.md` — an L1 ephemeral worktree diverges from the L2 base and silently breaks parallel-session coordination.

`Agent(isolation: "worktree")` is NOT a re-entry mechanism for existing L2 persistent worktrees. To re-enter an existing worktree:
- **Current-session re-entry** (no `/clear`, same session continuing): use the Claude Code runtime tool `EnterWorktree(<path>)` — see `EnterWorktree` / `ExitWorktree` Tools below.
- **New-session launch** (post-`/clear` or new terminal): use the launcher flag `moai cc -w <name-or-abs-path>` (see the `-w` L2 absolute-path extension below).

### `background: true` in Agent Frontmatter

Run agent without blocking the main conversation (v2.1.46+):

```yaml
---
name: team-coder
background: true   # Returns immediately; results delivered on next turn
---
```

Use with `isolation: worktree` for optimal parallel execution in team mode.

Background-execution policy for write-capable agents is owned by `.claude/rules/moai/core/agent-common-protocol.md` § Background Agent Execution. As of Claude Code v2.1.198 subagents run in the background by default, and a background write surfaces a permission prompt in the main session naming the asking subagent; MoAI aligns with that runtime default rather than forcing foreground. The retained safeguard is concurrency, not backgrounding: MoAI does not run two write-capable agents concurrently. Use `background: true` for:
- Read-only research and analysis agents
- Agents whose write paths are pre-approved in settings.json `permissions.allow`

Kill background agent: Press `Ctrl+X Ctrl+K` in Claude Code interface (v2.1.83+).

### Worktree Base Branch (`worktree.baseRef`)

Native worktrees (`--worktree` and subagent `isolation: worktree`) branch from the repository's default branch (`origin/HEAD`) by default, so they start from a clean tree matching the remote. If no remote is configured or the fetch fails, the worktree falls back to the current local `HEAD`. To always branch from local `HEAD` instead (carrying unpushed commits and feature-branch state), set `worktree.baseRef` to `"head"` in settings (accepts only `"fresh"` or `"head"`, not arbitrary refs):

```json
{
  "worktree": {
    "baseRef": "head"
  }
}
```

Use `"head"` when isolating subagents that must operate on in-progress work. To branch a native worktree from a specific pull request, pass the PR number prefixed with `#` (e.g. `claude --worktree "#1234"`); Claude Code fetches `pull/<number>/head` and creates the worktree at `.claude/worktrees/pr-<number>`.

This setting governs **Claude-native** worktrees only, which is now every worktree the launcher creates — `moai cc -w <name>` passes `-w` straight through to `claude`.

### `.worktreeinclude` (Copy Gitignored Files into Native Worktrees)

A native worktree is a fresh checkout, so untracked files (`.env`, `.env.local`, local config) are not present. Add a `.worktreeinclude` file at the project root to copy them automatically when Claude creates a worktree. It uses `.gitignore` syntax; only files that match a pattern AND are gitignored are copied (tracked files are never duplicated):

```text
.env
.env.local
.moai/config/sections/*.local.yaml
```

Applies to `claude --worktree`, subagent `isolation: worktree` worktrees, and desktop parallel sessions. NOT processed when a custom `WorktreeCreate` hook replaces the default git behavior — copy local files inside the hook script instead.

### `EnterWorktree` / `ExitWorktree` Tools

`EnterWorktree(<path>)` is the canonical mechanism for entering an existing worktree in the current session. The orchestrator's emitted guidance (paste-ready resume messages, Block 0 of the Worktree-Anchored Resume Pattern, in-session instructions) SHALL use `EnterWorktree(<path>)` for current-session worktree re-entry, replacing the shell-`cd`, `git -C <path>`, and subshell-`cd` patterns. A bare `cd` instruction SHALL NOT appear in orchestrator-emitted current-session-entry guidance; it remains valid only for human-typed, manual-shell contexts.

Claude can move the session into a worktree mid-session via the `EnterWorktree` tool (e.g. when the user says "work in a worktree"), creating one under `.claude/worktrees/`. Once inside, Claude can switch directly to another worktree by calling `EnterWorktree` with a target path; the previous worktree stays on disk untouched. `ExitWorktree` returns to the originating checkout. These are Claude Code runtime tools — MoAI does not mandate their use; they are the interactive counterpart to the launcher `-w` flag and `isolation: worktree` frontmatter.

`EnterWorktree` is complementary to, not replaced by, the launcher flag `moai cc -w <name-or-abs-path>`:

- **`EnterWorktree(<path>)`** — current-session re-entry (no `/clear`, same session continuing). Use this when the orchestrator is mid-turn and needs to move the active session into an existing worktree.
- **`moai cc -w <name>` (or `moai glm -w` / `moai cg -w`)** — new-session launch (post-`/clear` or new terminal). Use this as the Block 0 new-terminal launcher of the paste-ready resume. The `-w` flag accepts BOTH short names (resolved against `.claude/worktrees/<name>/`) AND absolute paths under `~/.moai/worktrees/<project>/...` (L2 persistent worktrees — see the `claude --worktree` (`-w`) Flag section above for the L2 absolute-path extension).

The shell-`cd` form (`cd <path> && <launcher>`), the `git -C <path>` form, and the subshell-`cd` form (`(cd <path> && ...)`) are DEPRECATED for orchestrator-emitted current-session worktree entry guidance. They break `Agent(isolation: "worktree")` CWD isolation (the agent's CWD is the worktree root; a `cd /absolute/path` bypasses it) and were the root cause of prior incidents where a sub-agent used `git -C` instead of `EnterWorktree` and was corrected mid-run.

## Worktree Selection Rules [ZONE:Evolvable] [HARD]

### Decision Tree

```
Is this a parallel write workers within a hierarchical team (e.g., manager-kanban fan-out)?
  YES → Use Agent(isolation: "worktree") for write agents
        Do NOT use isolation for read-only agents
  NO ↓

Is this a multi-session SPEC development?
  YES → Enter a worktree: moai cc -w <name>
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

- [ZONE:Evolvable] [HARD] Implementation leaf workers spawned in parallel by `manager-kanban` (or any parallel-write fan-out shape) MUST use `isolation: "worktree"` when spawned via Agent()
- [ZONE:Evolvable] [HARD] Read-only teammates (read-only research/review roles: researcher / analyst / reviewer) MUST NOT use `isolation: "worktree"` — read-only enforcement rests on tool restriction (`Explore`, or a `tools:` list omitting Write/Edit); the spawn-time `mode` parameter is deprecated and ignored since Claude Code v2.1.213, so a teammate is read-only only when its tools cannot write
- [ZONE:Evolvable] [HARD] One-shot sub-agents that write files across 3 or more paths per invocation MUST use `isolation: "worktree"`. This includes write-heavy retained agents (manager-develop), per-spawn `Agent(general-purpose)` specialists with a write-heavy domain whitelist (e.g. backend / frontend / devops / refactoring), and team-mode role profiles (implementer, tester, designer).
<!-- @MX:ANCHOR: WorktreeMUSTRule — invariant contract; all write-heavy agents MUST declare isolation:worktree; enforced by LR-05 lint rule -->
<!-- @MX:REASON: MUST level required to eliminate silent file-write conflict failure mode in parallel Agent() execution. -->
- [ZONE:Evolvable] [HARD] GitHub workflow agents (fixer agents in /moai github issues) MUST use `isolation: "worktree"` for branch isolation

### Parallel-Session Branch Conflict Auto-Isolation

[ZONE:Evolvable] [HARD] **When** the orchestrator detects (via the Pre-Spawn Sync Check active-sessions registry OR the Pre-Edit Sync Check, `.moai/state/active-sessions.json`) that ≥1 foreign active session is on the same checkout during **any write work** — whether worktree entry was chosen OR the orchestrator is editing the shared tree directly (direct main-session Edit/Write/Bash, which bypasses the spawn gate; see `.claude/rules/moai/core/agent-common-protocol.md` § Pre-Edit Sync Check) — the orchestrator SHALL auto-create one worktree per foreign registry entry (or isolate the direct-edit work into a worktree) to prevent cross-session branch-state interference. This auto-isolation procedure resolves the parallel-session branch conflict mechanically rather than surfacing it as a manual race. The "worktree entry is chosen" conjunct is no longer required: a foreign active session during direct-edit work triggers isolation too, because direct edits share the same branch-state mutable surface as spawned-agent writes.

**Conservative predicate** — ANY foreign active-session registry entry triggers auto-isolation. False positives are cheap (an extra worktree is inexpensive and user-deletable); false negatives corrupt the working tree (a genuine conflict goes unresolved and produces cross-session branch-state interference). Stale-registry false positives MAY produce a worktree that the user later deletes.

**Naming scheme** — each auto-created worktree is named `auto-<session-short>-<spec-id>` where `<session-short>` is the first 8 characters of THAT foreign session entry's UUID and `<spec-id>` is the active SPEC identifier. The naming is deterministic so the auto-created worktree is greppable and traceable to the originating foreign session. No "or equivalent" clause — the scheme is fixed.

**Landing paths** — each auto-created worktree SHALL land under `.claude/worktrees/auto-<session-short>-<spec-id>/` (L1 Claude-native) OR `~/.moai/worktrees/<project>/auto-<session-short>-<spec-id>/` (L2 persistent), so the two sessions do NOT share a branch-state mutable surface. The primary-checkout branch guard exempts worktree paths from its deny, so the auto-isolation procedure does NOT trip the branch guard.

**Surface** — the orchestrator surfaces the auto-isolation as an info log (NOT an `AskUserQuestion` round — the procedure auto-resolves the race, it does not ask the user to resolve it). The info log notes the registry entry's age so a stale-registry false positive is visible.

**Multiple foreign sessions (≥2)** — the procedure auto-creates N worktrees, one per foreign registry entry (Edge-3: each session gets its own isolated branch-state surface).

## Sentinel Key Glossary

Structured error codes emitted by `moai agent lint` and `moai workflow lint` for programmatic detection:

| Sentinel Key | Source | Meaning |
|---|---|---|
| `ORC_WORKTREE_MISSING` | LR-05 (agent lint) | Write-heavy agent lacks `isolation: worktree` in frontmatter |
| `ORC_WORKTREE_ON_READONLY` | LR-09 (agent lint) | Read-only agent (`permissionMode: plan`) has `isolation: worktree` — prohibited overhead |
| `ORC_WORKTREE_REQUIRED` | `moai workflow lint` | Legacy sentinel from the retired Agent Teams `role_profiles` isolation check — inert since the `team:` config block was removed from workflow.yaml |

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
# Implementation teammates (write-capable implementation roles: implementer / tester / designer)
# Spawned via: Agent(subagent_type: "general-purpose", mode: "acceptEdits", isolation: "worktree")
isolation: worktree   # Isolated worktree per agent
background: true      # Non-blocking parallel execution
permissionMode: acceptEdits
```

### Research/Analysis Agents (no isolation needed)

```yaml
# Read-only teammates (read-only research/review roles: researcher / analyst / reviewer)
# Spawned via: Agent(subagent_type: "general-purpose") with a read-only tools list
# No isolation: worktree (read-only via tool restriction — the spawn-time mode
# parameter is deprecated/ignored since v2.1.213; a parent bypassPermissions/
# acceptEdits mode takes precedence over child permission settings)
permissionMode: plan  # advisory frontmatter; the tool restriction is the guarantee
```

## WorktreeCreate and WorktreeRemove Hooks (Not Registered by Default)

Claude Code v2.1.49+ defines `WorktreeCreate` / `WorktreeRemove` hooks that **replace** Claude Code's default git worktree behavior — not extend it. Per the official contract (https://code.claude.com/docs/en/hooks):

| Hook | Role | stdout contract | Failure mode |
|---|---|---|---|
| WorktreeCreate | Active creator — MUST actually create the worktree directory and echo its absolute path to stdout (plain text only, no JSON; HTTP hooks use `{"hookSpecificOutput": {"worktreePath": "..."}}`). | Single line: `/absolute/path/to/worktree` | Empty stdout OR any non-zero exit aborts creation |
| WorktreeRemove | Observer — runs during/after removal for cleanup. | No output required | Failures logged in debug mode only |

The stdin JSON for both events includes `worktree_path` (Claude Code's proposed path), `name`, `cwd`, `session_id`, `transcript_path`, `hook_event_name`.

**MoAI-ADK does NOT register these hooks by default.** Claude Code's default git worktree handling is sufficient for our agent isolation use case — write-heavy work is declared `isolation: worktree` by the retained `manager-develop` agent, by per-spawn `Agent(general-purpose)` specialists with a write-heavy domain whitelist, and by team-mode role profiles (implementer, tester, designer) per the Worktree Selection Rules above. Registering observer-only hooks here would replace the default behavior with non-functional stubs and produce `"WorktreeCreate hook returned a path that is not a directory: {}"` because an empty JSON object cannot be parsed as a path.

If a future use case requires custom worktree creation (e.g., non-git VCS, shared-file symlinks, per-worktree database setup), implement an active creator hook that:

1. Reads stdin JSON (fields: `worktree_path`, `name`, `cwd`, `session_id`).
2. Performs `git worktree add` (or equivalent for the VCS), redirecting its stdout to `/dev/null` so it does not pollute the hook stdout.
3. Prints **only** the absolute worktree path to stdout. All progress/diagnostic output goes to stderr.
4. Exits 0 on success; any non-zero exit aborts creation.

Handler files at `internal/hook/worktree_{create,remove}.go` and `internal/cli/hook.go` `worktree-create` / `worktree-remove` subcommands are preserved as opt-in infrastructure for future active-creator implementations. They are not registered in `.claude/settings.json` until such an implementation lands. Likewise, `.claude/hooks/moai/handle-worktree-{create,remove}.sh` wrapper scripts exist but are not invoked by any settings.json entry.

## Prompt Path Rules for Worktree-Isolated Agents

When the orchestrator generates prompts for agents spawned with `isolation: "worktree"`, paths in the prompt determine where the agent operates. Incorrect paths bypass worktree isolation entirely.

### HARD Rules

- [ZONE:Frozen] [HARD] Do NOT include absolute paths to the main project directory in agent prompts for write-target files
- [ZONE:Frozen] [HARD] Do NOT include `cd /absolute/project/path &&` in Bash commands within agent prompts
- [ZONE:Frozen] [HARD] Reference write-target files by project-root-relative paths (e.g., `src/domains/auth/handler.go`) and let the agent resolve from its own CWD
- [ZONE:Frozen] [HARD] `$CLAUDE_PROJECT_DIR` in hook commands is acceptable — Claude Code resolves this to the correct directory for the agent's context

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
Main repo:  $HOME/project/src/auth/handler.go
Worktree:   $HOME/project/.claude/worktrees/abc123/src/auth/handler.go
```

Both share the same project structure. `src/auth/handler.go` resolves correctly in either context.

### Anti-Pattern Examples

```
# WRONG: Absolute path in prompt bypasses worktree
"Read $HOME/project/src/auth/handler.go and fix the bug"

# WRONG: cd to main project in Bash command
"Run: cd $HOME/project && go test ./..."

# CORRECT: Relative path — agent resolves from its own CWD
"The bug is in src/auth/handler.go. Read the file and fix it."

# CORRECT: No cd prefix — agent CWD is already worktree root
"Run: go test ./..."
```

## Teammate Session Launch (`--spawn`)

`moai cc -w <name> --spawn` (likewise `moai glm` / `moai cg`) opens a session in the named worktree in a NEW tmux window and returns, so the caller keeps its own session. Without `--spawn` the same command enters the worktree in place by replacing the current process. See `.claude/skills/moai-workflow-worktree/SKILL.md` § `--spawn` for requirements, error messages, and example invocations.

> **Two distinct `teammateMode` fields — do not conflate.** MoAI's own `.claude/settings.local.json` launcher-selection field (values `"tmux"` / `"glm"` / `"claude"`) is set by `moai cg` / `moai glm` / `moai cc` and selects which launcher a session runs. This is SEPARATE from the Claude Code runtime `teammateMode` setting, whose default changed from `auto` to `in-process` as of Claude Code v2.1.179 — with the in-process default, split panes no longer auto-open. Additionally, as of Claude Code v2.1.181, an idle teammate's agent-panel row hides after 30 seconds and reappears on the next turn. These two CC-runtime behaviors govern how teammates are displayed. Both fields happen to share the name `teammateMode`.

### HARD Rules

[ZONE:Frozen] [HARD] CLI launch decisions MUST NOT invoke `AskUserQuestion`. Every launch outcome is decided from observable state (tmux session presence, `teammateMode`, GLM env vars) and reported through exit codes and stderr. This satisfies the Branch Origin Decision Protocol (see `.claude/rules/moai/development/branch-origin-protocol.md` § HARD Rules).

Static guard: `internal/cli/worktree/new_test.go` `TestNew_NoAskUserQuestion` scans the worktree-creation source for `AskUserQuestion` / `mcp__askuser` references.

[ZONE:Evolvable] [HARD] `--spawn` refuses rather than degrades. Outside tmux, or without the `tmux` / `moai` binaries, it returns a non-zero exit instead of falling back to an in-place launch — a silent fallback would replace the caller's session, the outcome the flag exists to avoid. Refusal happens before any settings mutation.

### Retired: `moai worktree new --team` and the swarm registry

The `--team` flag and its four launch patterns are retired. Entering a worktree is `-w`; spawning a teammate window is `--spawn`. The write-only `.moai/state/swarm/<SPEC-ID>.json` registry was retired with it — no code ever read it, and the `moai swarm status / done / kill-all` commands it was a baseline for were never built.

### Cross-references

- `.claude/skills/moai-workflow-worktree/SKILL.md` § `--spawn` (requirements + examples)
- `internal/cli/spawn.go`, `internal/cli/spawn_test.go`
- Branch Origin Decision Protocol (BODP)

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

**Recommended**: Claude Code **2.1.186 or later** for current background-agent permission-prompt semantics, Opus 4.7+ / 4.8 / Opus 5 support, MCP doctor warnings, and Windows CLAUDE_ENV_FILE parity. Minimum baseline: **2.1.97** for worktree isolation.

## Troubleshooting

| Issue | Cause | Solution |
|-------|-------|----------|
| Worktree not found | Removed manually | Run `git worktree list` to verify |
| Agent worktree conflicts | Multiple agents same file | Check file ownership in team config |
| Stale worktree branches | Incomplete cleanup | Run `git worktree prune` |
| Hooks not firing | Missing wrapper script | Check `.claude/hooks/moai/` directory |
| `--tmux` not working | Unsupported terminal | Use tmux or iTerm2 (not VS Code, Ghostty) |

## SPEC-to-Worktree Mapping

[ZONE:Frozen] [HARD] Per-step worktree applicability is governed by `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Phase Discipline (canonical source). This table summarizes the mapping for quick reference; on conflict, spec-workflow.md wins.

| Step | Phase   | Worktree?                | Location                              | Lifecycle event              |
|------|---------|--------------------------|---------------------------------------|------------------------------|
| 1    | Plan    | **NO** (main checkout)   | n/a — `plan/SPEC-XXX` branch on main  | plan PR merged               |
| 2    | Run     | **opt-in (`moai cc -w <name>`)** | the entered worktree                  | run PR merged                |
| 3    | Sync    | **opt-in** — same as Step 2            | same path as Step 2 (do NOT recreate) | sync PR merged               |
| 4    | Cleanup | n/a                      | host checkout                         | `moai worktree done SPEC-XXX` |

Worktree usage is user opt-in; the default flow runs all phases on a `feat/SPEC-XXX` branch in the main checkout. To use one, enter it with `moai cc -w <name>` before invoking the phase.

[ZONE:Frozen] [HARD] Disposal contract: `moai worktree done SPEC-XXX` MUST run only after BOTH run PR AND sync PR are merged. Premature disposal between Step 2 merge and Step 3 merge breaks Sync.

---

Version: 4.1.0 (Team Protocol section removed — Agent Teams static layer retired)
