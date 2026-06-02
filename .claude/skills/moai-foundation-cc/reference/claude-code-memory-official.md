# Claude Code Memory System - Official Reference

Source: https://code.claude.com/docs/en/memory

## Overview

Claude Code has two complementary memory systems that carry knowledge across sessions:

1. **CLAUDE.md files** — Instructions you write to give Claude persistent context
2. **Auto memory** — Notes Claude writes itself based on discoveries and patterns

Both are loaded at the start of every session and treated as context, not enforced configuration.

## CLAUDE.md Files

### Purpose

CLAUDE.md files are markdown files that give Claude persistent instructions for a project, your personal workflow, or your organization. You write these files in plain text; Claude reads them at the start of every session.

### File Locations (Load Order)

CLAUDE.md files can live in several locations, each with a different scope. Files are loaded in order from broadest to most specific:

| Scope | Location | Purpose |
|-------|----------|---------|
| **Managed policy** | `/Library/Application Support/ClaudeCode/CLAUDE.md` (macOS)<br />`/etc/claude-code/CLAUDE.md` (Linux/WSL)<br />`C:\Program Files\ClaudeCode\CLAUDE.md` (Windows) | Organization-wide instructions managed by IT/DevOps |
| **User** | `~/.claude/CLAUDE.md` | Personal preferences for all projects |
| **Project** | `./CLAUDE.md` or `./.claude/CLAUDE.md` | Team-shared instructions for the project |
| **Local** | gitignored local instructions file at the project root | Personal project-specific preferences |

Files in subdirectories load on demand when Claude reads files in those directories, not at session start.

### Writing Effective Instructions

- **Size**: Target under 200 lines per file. Longer files consume more context tokens and reduce adherence.
- **Structure**: Use markdown headers and bullets to group related instructions. Organized sections are easier to follow than dense paragraphs.
- **Specificity**: Write instructions concrete enough to verify. "Use 2-space indentation" instead of "Format code properly." "API handlers live in src/api/handlers/" instead of "Keep files organized."
- **Consistency**: If two rules contradict each other, Claude may pick one arbitrarily. Remove outdated or conflicting instructions periodically.

### Organizing Instructions with `.claude/rules/`

For larger projects, organize instructions into multiple files using the `.claude/rules/` directory. Each file covers one topic with a descriptive filename like `testing.md` or `api-design.md`. Rules without `paths` frontmatter load at session start; rules with path-scoped `paths:` frontmatter load only when matching files are encountered.

```yaml
---
paths: "src/api/**/*.ts"
---

# API Development Rules

- All API endpoints must include input validation
- Use the standard error response format
```

### Import Additional Files

CLAUDE.md files can import additional files using `@path/to/import` syntax. Imported files are expanded and loaded into context at launch.

Example:
```
See @README for project overview and @package.json for available npm commands.
```

Relative paths resolve relative to the file containing the import. Maximum import depth is four hops. The first time Claude Code encounters external imports, it shows an approval dialog. Imports stay disabled if declined.

### Manage CLAUDE.md for Large Teams

Organizations can deploy a centrally managed CLAUDE.md applied to all users on a machine. File locations are listed above under Managed policy scope. Use organization-wide configuration management systems (MDM, Group Policy, Ansible) to distribute the file.

The `claudeMd` key in managed settings puts content directly in `managed-settings.json` instead of deploying a separate file.

Exclude specific CLAUDE.md files in large monorepos using the `claudeMdExcludes` setting in `.claude/settings.local.json`:

```json
{
  "claudeMdExcludes": [
    "**/monorepo/CLAUDE.md",
    "/path/to/other-team/.claude/rules/**"
  ]
}
```

## Auto Memory

Auto memory lets Claude accumulate knowledge across sessions without you writing anything. Claude saves notes for itself as it works: build commands, debugging insights, architecture notes, code style preferences, and workflow habits.

Auto memory requires Claude Code v2.1.59 or later (check with `claude --version`).

### Enable or Disable

Auto memory is on by default. Toggle it via `/memory` command in a session, or set `autoMemoryEnabled` in settings:

```json
{
  "autoMemoryEnabled": false
}
```

Or set the environment variable: `CLAUDE_CODE_DISABLE_AUTO_MEMORY=1`

### Storage Location

Each project gets its own memory directory at `~/.claude/projects/<project>/memory/`. The `<project>` path is derived from the git repository; all worktrees and subdirectories within the same repo share one auto memory directory. Outside a git repo, the project root is used instead.

To store auto memory in a different location, set `autoMemoryDirectory`:

```json
{
  "autoMemoryDirectory": "~/my-custom-memory-dir"
}
```

The directory contains a `MEMORY.md` index and optional topic files:

```
~/.claude/projects/<project>/memory/
├── MEMORY.md                # Index, loaded every session
├── debugging.md             # Detailed notes on demand
├── api-conventions.md       # API design decisions
└── ...
```

### How It Works

The first 200 lines of `MEMORY.md` (or the first 25KB, whichever comes first) are loaded at the start of every session. Content beyond that threshold is not loaded automatically.

Topic files like `debugging.md` are not loaded at startup. Claude reads them on demand using standard file tools when the information is needed.

Auto memory is machine-local. All worktrees and subdirectories within the same git repository share one auto memory directory. Files are not synced across machines or cloud environments.

## View and Edit Memory with `/memory`

The `/memory` command lists:
- All loaded CLAUDE.md and `.claude/rules/` files in your current session
- Toggle for auto memory on/off
- Link to open the auto memory folder
- Options to select and open any file in your editor

## Troubleshoot Memory Issues

### Claude isn't following my CLAUDE.md

CLAUDE.md content is delivered as context after the system prompt, not as the system prompt itself. Claude reads it and tries to follow it, but there's no strict enforcement, especially for vague or conflicting instructions.

Debug steps:
1. Run `/memory` to verify your files are being loaded.
2. Check that the file is in a location that gets loaded for your session.
3. Make instructions more specific. "Use 2-space indentation" works better than "format code nicely."
4. Look for conflicting instructions across CLAUDE.md files. If two files give different guidance, Claude may pick one arbitrarily.
5. For instructions that must run at a specific point (before every commit, after file edits), write a [hook](/en/hooks-guide) instead. Hooks execute as shell commands at fixed lifecycle events.

Tip: Use the `InstructionsLoaded` hook to log exactly which instruction files are loaded, when they load, and why.

### I don't know what auto memory saved

Run `/memory` and select the auto memory folder to browse what Claude has saved. Everything is plain markdown you can read, edit, or delete.

### My CLAUDE.md is too large

Files over 200 lines consume more context and may reduce adherence. Use [path-scoped rules](#organizing-instructions-with-claude/rules/) to load instructions only for matching files, or trim content not needed in every session. Splitting into `@path` imports helps organization but does not reduce context, since imported files load at launch.

### Instructions seem lost after `/compact`

Project-root CLAUDE.md survives compaction: Claude re-reads it from disk and re-injects it into the session. Nested CLAUDE.md files in subdirectories are not re-injected automatically; they reload the next time Claude reads a file in that subdirectory.

## Related Resources

- [Debug your configuration](/en/debug-your-config): diagnose why CLAUDE.md or settings aren't taking effect
- [Skills](/en/skills): package repeatable workflows that load on demand
- [Settings](/en/settings): configure Claude Code behavior with settings files
- [Subagent memory](/en/sub-agents#enable-persistent-memory): let subagents maintain their own auto memory
