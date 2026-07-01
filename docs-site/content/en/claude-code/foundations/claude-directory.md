---
title: The .claude Directory
weight: 60
draft: false
description: "Documents the structure and scopes of the .claude directory, the per-project configuration root from which Claude Code reads CLAUDE.md, settings.json, skills, sub-agents, and hooks."
---

The `.claude` directory is the single configuration root from which Claude Code reads each project's instructions, settings, and extensions.

{{< callout type="info" >}}
**TL;DR**: `.claude` is the project-specific "control panel" that Claude Code consults at the start of every session. Most of it is committed to git and shared with the team, while only personal files are kept separate.
{{< /callout >}}

For most users, editing just two files — `CLAUDE.md` and `settings.json` — is enough. The rest (skills, rules, sub-agents) can be added one at a time as the need arises.

## What the .claude Directory Does

Claude Code reads settings from two places: the `.claude/` directory of the project you are working in, and `~/.claude/` in your home directory. Files inside the project are committed to git and shared with the team, while files in `~/.claude/` remain personal settings that apply to every project.

- **Delivering project context**: instructions that Claude "reads and follows," such as `CLAUDE.md`
- **Enforcing behavior**: settings that are "enforced" regardless of whether Claude complies, such as `settings.json` permissions and hooks
- **Storing extensions**: reusable assets such as skills, sub-agents, and dynamic workflows

The key distinction here is between **guidance** and **configuration**. `CLAUDE.md` and rules are advisory notes that Claude consults, so there is no guarantee they are always honored, whereas hooks and permissions are enforced directly by the runtime and are therefore deterministic. When you need guaranteed behavior, implement it with hooks or permissions rather than guidance.

## Project .claude/ Directory Structure

| Item | Location | Commit | Role |
| --- | --- | --- | --- |
| `CLAUDE.md` | project root or `.claude/` | ✓ | Project instructions loaded as context at every session |
| `settings.json` | `.claude/` | ✓ | Enforced settings: permissions, hooks, environment variables, default model, etc. |
| `settings.local.json` | `.claude/` | - | Personal settings override (automatically gitignored) |
| `rules/` | `.claude/` | ✓ | Instructions split by topic, can be loaded conditionally by file path |
| `skills/` | `.claude/` | ✓ | Skills invoked with `/name` or called automatically by Claude |
| `commands/` | `.claude/` | ✓ | Single-file prompts (same mechanism as skills) |
| `agents/` | `.claude/` | ✓ | Sub-agent definitions with their own independent context windows |
| `workflows/` | `.claude/` | ✓ | Dynamic workflow scripts that orchestrate multiple sub-agents |
| `hooks/` | `.claude/` | ✓ | Scripts that hooks execute (registered in settings.json) |
| `agent-memory/` | `.claude/` | ✓ | Persistent memory dedicated to sub-agents |
| `.mcp.json` | project root | ✓ | Team-shared MCP server configuration |
| `.worktreeinclude` | project root | ✓ | Gitignore patterns to copy when creating worktrees |

### Guidance Files (what Claude reads)

**`CLAUDE.md`**: holds the project's rules, frequently used commands, and architectural context. Because the entire file is loaded as context every session, 200 lines or fewer is recommended; when it grows longer, split it into rules.

**`rules/*.md`**: loaded at session start when there is no `paths:` frontmatter, and loaded only when the matching file enters context when a `paths:` glob is present. When `CLAUDE.md` approaches 200 lines, the best practice is to split it into topic-based rules.

### Enforced Configuration (what Claude Code enforces)

**`settings.json`**: holds the `permissions` (allow/deny tools and commands), `hooks` (run scripts at event points), `statusLine`, `model`, `env`, and `outputStyle` keys.

**`settings.local.json`**: the same schema but personal and not committed. Use it when you need permissions that differ from the team defaults.

### Extension Assets

**`skills/<name>/SKILL.md`**: folder-based skills that can bundle reference docs, templates, and scripts together.

**`commands/*.md`**: single-file prompts. Officially the same mechanism as skills; writing new workflows as skills is recommended.

**`agents/*.md`**: sub-agents with their own system prompt and tool access. They run in a fresh context window, keeping the main conversation clean.

**`workflows/*.js`**: dynamic workflow scripts that spawn and orchestrate multiple sub-agents.

## Global ~/.claude/ Directory Structure

| Item | Location | Role |
| --- | --- | --- |
| `CLAUDE.md` | `~/.claude/` | Personal instructions applied to all projects |
| `settings.json` | `~/.claude/` | Default settings for all projects (overridden by project settings) |
| `keybindings.json` | `~/.claude/` | Custom keyboard shortcuts |
| `skills/` | `~/.claude/` | Personal skills available in all projects |
| `commands/` | `~/.claude/` | Personal commands available in all projects |
| `agents/` | `~/.claude/` | Personal sub-agents available in all projects |
| `workflows/` | `~/.claude/` | Personal workflows available in all projects |
| `output-styles/` | `~/.claude/` | Personal output styles |

## Configuration Scopes and Precedence

The same setting can exist in multiple locations, and the more specific scope wins. Scopes are divided into three levels: enterprise, user, and project.

| Scope | Location | Applies To |
| --- | --- | --- |
| Enterprise | `managed-settings.json` (OS-specific system path) | Entire organization (cannot be overridden; highest priority) |
| User (global) | `~/.claude/` | All projects (personal defaults) |
| Project | `.claude/` | Current project (team-shared) |
| Project local | `.claude/settings.local.json` | Current project, personal (highest priority among user-edited files) |

**Array settings** (such as `permissions.allow`) **combine** values from all scopes. **Scalar settings** (such as `model`) **use the single value** from the most specific scope.

## Version-Controlled vs Excluded

| File | Commit | Reason |
| --- | --- | --- |
| `CLAUDE.md`, `rules/`, `settings.json` | ✓ | Context and policy shared by the team |
| `skills/`, `commands/`, `agents/`, `workflows/` | ✓ | Extension assets shared by the team |
| `.mcp.json` | ✓ | Team-shared MCP server configuration |
| `settings.local.json` | - | Personal override (automatically gitignored) |
| All of `~/.claude/` | - | Personal settings applied to every project; never commit |
| `CLAUDE.local.md` | - | Per-project personal instructions (create manually and add to `.gitignore`) |

Claude Code automatically adds `settings.local.json` to `.gitignore` when it first creates the file.

## Related Docs

- [settings.json Guide](/advanced/settings-json)
- [CLAUDE.md Guide](/advanced/claude-md-guide)
- [Statusline System](/advanced/statusline)

## References

- [Explore the .claude directory (Claude Code official docs)](https://code.claude.com/docs/en/claude-directory)

{{< callout type="tip" >}}
For a new project, fill in just the two files `CLAUDE.md` and `settings.json` first; put team permissions and hooks in the project `settings.json` and permissions only you use in `settings.local.json`, and you can start cleanly without git conflicts.
{{< /callout >}}
