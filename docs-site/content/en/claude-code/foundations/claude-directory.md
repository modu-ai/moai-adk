---
title: .claude Directory
weight: 60
draft: false
description: "The structure and scopes of the .claude directory — the per-project configuration root from which Claude Code reads CLAUDE.md, settings.json, skills, subagents, and hooks."
---

# .claude Directory

The `.claude` directory is the single configuration root from which Claude Code reads instructions, settings, and extensions for each project.

{{< callout type="info" >}}
**One-line summary**: `.claude` is the project-specific "control panel" Claude Code inspects at the start of every session; commit most of it to git to share with your team, and isolate only the personal files.
{{< /callout >}}

For most users, editing just two files — `CLAUDE.md` and `settings.json` — is enough. Add the rest — skills, rules, subagents — one at a time as needed.

## The Role of the .claude Directory

Claude Code reads configuration from two places: the `.claude/` directory of the project you are working in, and `~/.claude/` in your home directory. Files inside the project are committed to git and shared with the team, while files in `~/.claude/` remain personal settings applied to all projects.

- **Delivering project context**: guidance Claude "reads and follows", like `CLAUDE.md`
- **Enforcing behavior**: settings that are "executed" regardless of Claude's compliance, like `settings.json` permissions and hooks
- **Storing extensions**: reusable assets such as skills, subagents, and dynamic workflows

The key distinction here is **guidance** vs **configuration**. `CLAUDE.md` and rules are guidance Claude consults, with no guarantee they are always followed; hooks and permissions are enforced directly by the runtime and are therefore deterministic. When you need guaranteed behavior, implement it as a hook or permission, not as guidance. This distinction is the first design decision of harness engineering — MoAI-ADK, too, deploys orchestrator guidance (CLAUDE.md), quality-gate hooks, and agent/skill assets into this directory with a single `moai init`, composing a project-specific harness.

## Project .claude/ Directory Structure

| Item | Location | Commit | Role |
| --- | --- | --- | --- |
| `CLAUDE.md` | Project root or `.claude/` | ✓ | Project guidance loaded into context every session |
| `settings.json` | `.claude/` | ✓ | Enforced settings: permissions, hooks, env vars, default model |
| `settings.local.json` | `.claude/` | - | Personal settings override (auto-gitignored) |
| `rules/` | `.claude/` | ✓ | Guidance split by topic; can be conditionally loaded by file path |
| `skills/` | `.claude/` | ✓ | Skills invoked via `/name` or auto-invoked by Claude |
| `commands/` | `.claude/` | ✓ | Single-file prompts (same mechanism as skills) |
| `agents/` | `.claude/` | ✓ | Subagent definitions with independent context windows |
| `workflows/` | `.claude/` | ✓ | Dynamic workflow scripts coordinating multiple subagents |
| `hooks/` | `.claude/` | ✓ | Scripts run by hooks (registered in settings.json) |
| `agent-memory/` | `.claude/` | ✓ | Persistent memory dedicated to subagents |
| `.mcp.json` | Project root | ✓ | Team-shared MCP server configuration |
| `.worktreeinclude` | Project root | ✓ | Gitignore patterns to copy when creating worktrees |

### Guidance Files (What Claude Reads)

**`CLAUDE.md`**: holds the project's rules, frequent commands, and architectural context. Its full contents load into context every session, so keeping it under 200 lines is recommended; split into rules when it grows.

**`rules/*.md`**: without a `paths:` frontmatter they load at session start; with a `paths:` glob they load only when a matching file enters context. Splitting into per-topic rules is best practice as `CLAUDE.md` approaches 200 lines.

### Enforced Settings (What Claude Code Executes)

**`settings.json`**: carries the `permissions` (tool/command allow/deny), `hooks` (scripts run on events), `statusLine`, `model`, `env`, and `outputStyle` keys.

**`settings.local.json`**: the same schema but personal and never committed. Use it when you need permissions that differ from the team defaults.

### Extension Assets

**`skills/<name>/SKILL.md`**: folder-based skills that can bundle reference docs, templates, and scripts together.

**`commands/*.md`**: single-file prompts. Officially the same mechanism as skills; new workflows are recommended to be written as skills.

**`agents/*.md`**: subagents with their own system prompt and tool access. They run in a fresh context window, keeping the main conversation clean.

**`workflows/*.js`**: dynamic workflow scripts that spawn and coordinate many subagents.

## Global ~/.claude/ Directory Structure

| Item | Location | Role |
| --- | --- | --- |
| `CLAUDE.md` | `~/.claude/` | Personal guidance applied to all projects |
| `settings.json` | `~/.claude/` | Default settings for all projects (overridden by project settings) |
| `keybindings.json` | `~/.claude/` | Custom keyboard shortcuts |
| `skills/` | `~/.claude/` | Personal skills available in all projects |
| `commands/` | `~/.claude/` | Personal commands available in all projects |
| `agents/` | `~/.claude/` | Personal subagents available in all projects |
| `workflows/` | `~/.claude/` | Personal workflows available in all projects |
| `output-styles/` | `~/.claude/` | Personal output styles |
| `projects/` | `~/.claude/` | Per-project session records, conversation transcripts, auto memory |

## Setting Scopes and Priority

The same setting can exist at multiple locations, and the more specific scope wins. Scopes come in three tiers: enterprise, user, and project.

| Scope | Location | Applies to |
| --- | --- | --- |
| Enterprise | `managed-settings.json` (OS-specific system path) | Entire organization (not user-overridable; top priority) |
| User (global) | `~/.claude/` | All projects (personal defaults) |
| Project | `.claude/` | The current project (team-shared) |
| Project local | `.claude/settings.local.json` | The current project, personal (top priority among user-edited files) |

**Array settings** (like `permissions.allow`) are **merged** across all scopes. **Scalar settings** (like `model`) use the **single value** from the most specific scope.

## Versioned vs Excluded

| File | Commit | Reason |
| --- | --- | --- |
| `CLAUDE.md`, `rules/`, `settings.json` | ✓ | Team-shared context and policy |
| `skills/`, `commands/`, `agents/`, `workflows/` | ✓ | Team-shared extension assets |
| `.mcp.json` | ✓ | Team-shared MCP server configuration |
| `settings.local.json` | - | Personal overrides (auto-gitignored) |
| All of `~/.claude/` | - | Personal settings applied to all projects |
| `CLAUDE.local.md` | - | Per-project personal guidance (create manually, then add to `.gitignore`) |

Claude Code automatically adds `settings.local.json` to `.gitignore` when first creating it.

## Related Documents

- [settings.json Guide](/advanced/settings-json)
- [CLAUDE.md Guide](/advanced/claude-md-guide)
- [Statusline System](/advanced/statusline)

## References

- [Explore the .claude directory (Claude Code official docs)](https://code.claude.com/docs/en/claude-directory)

{{< callout type="tip" >}}
For a new project, fill in just `CLAUDE.md` and `settings.json` first. Put team permissions and hooks in the project `settings.json`, and permissions only you use in `settings.local.json` — a clean start with no git conflicts.
{{< /callout >}}
