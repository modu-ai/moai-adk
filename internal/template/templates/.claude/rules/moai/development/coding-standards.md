---
paths: ".claude/**/*.md,.claude/**/*.yaml,.moai/**/*.yaml,CLAUDE.md"
---

# Coding Standards

MoAI-specific coding standards. General coding conventions are not included as Claude already knows them.

## Language Policy

All instruction documents must be in English:
- CLAUDE.md
- Agent definitions (.claude/agents/**/*.md)
- Slash commands (.claude/commands/**/*.md)
- Skill definitions (.claude/skills/**/*.md)
- Hook scripts (.claude/hooks/**/*.py, *.sh)
- Configuration files (.moai/config/**/*.yaml)

User-facing documentation may use multiple languages:
- README.md, CHANGELOG.md
- User guides, API documentation

## File Size Limits

CLAUDE.md should stay under 40,000 characters. This is a MoAI CI-enforceable heuristic; the official Claude Code spec instead targets "under 200 lines per CLAUDE.md" and loads the file in full regardless of length. Any project-local instruction file that also loads in full at every session launch follows the same size discipline.

When approaching the limit, reduce launch-time context (priority order):
- Move detailed content to path-scoped rules (.claude/rules/ with `paths:` frontmatter) so it loads only when matching files are touched
- Move stable doctrine to .moai/docs/ and reference it with a plain prose pointer ("See: .moai/docs/<file>.md")
- Trim content not needed in every session
- Keep only core identity and hard rules inline

Note: `@import` (`@path/to/file`) does NOT reduce context — imported files are expanded and loaded in full at launch (see `.claude/skills/moai-foundation-cc/reference/claude-code-memory-official.md`). Use it for organization, never for size reduction.

## Content Restrictions

Prohibited in instruction documents:
- Code examples for conceptual explanations
- Flow control as code syntax
- Decision trees as code structures
- Emoji characters (except output styles)
- Time estimates or duration predictions

## Footer Convention

Rule files follow a consistent-by-absence footer policy: a `Version` / `Status` / `Classification` footer is OPTIONAL, not required. SSOT-owning canonical-reference rules (files that declare "Single Source of Truth" or carry a `Classification: Canonical Reference` line) SHOULD include a footer stating version, status, and classification. Short path-scoped rules MAY omit a footer entirely — absence is a valid consistent state, not a gap. Do not bulk-add footers to rules that currently lack one; the policy statement (this section) is the deliverable, not uniform footer insertion.

## Duplicate Prevention

Single source of truth principle:
- Each piece of information exists in exactly one location
- Use references (@file) instead of copying content
- Update source file, not copies

## Thin Command Pattern

All slash command files MUST be thin routing wrappers (under 20 LOC body).

Rules:
- Commands route to skills via `Skill("moai")` -- they never contain workflow logic
- All workflow logic belongs in `.claude/skills/moai/workflows/` or skill body
- YAML frontmatter must include: description, argument-hint, allowed-tools (CSV string)
- Root commands may contain router tables but no implementation logic
- Custom commands and skills are merged into one namespace: a `.claude/commands/X.md` and a `.claude/skills/X/SKILL.md` with the same name are the same invocation, and the skill form wins when both exist. Author the workflow once as a skill rather than duplicating it across a command and a skill.

Template:
```
---
description: [One-sentence action description]
argument-hint: "[Optional arg]"
allowed-tools: Skill
---

Use Skill("moai") with arguments: [subcommand] $ARGUMENTS
```

Enforcement: `internal/template/commands_audit_test.go` verifies this pattern on every `go test`.


## Claude Code Version Compatibility

Settings fields introduced by specific Claude Code versions:

| Field | Version | Notes |
|-------|---------|-------|
| `effortLevel` | v2.1.110 | Sets CLAUDE_CODE_EFFORT_LEVEL; values: low/medium/high/xhigh/max |
| `disableBypassPermissionsMode` | v2.1.111 | Prevents agents from using bypassPermissions mode when true |
| `Bash(timeout=N)` | v2.1.110 | Per-command Bash timeout in ms; max 600,000ms |

When adding new settings fields, update `internal/template/templates/.claude/settings.json.tmpl`
and this compatibility table.

## Paths Frontmatter

Use paths frontmatter for conditional rule loading:

```yaml
---
paths: "**/*.py,**/pyproject.toml"
---
```

This ensures rules load only when working with matching files.
