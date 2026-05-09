---
description: MoAI-specific coding standards for instruction documents and configuration files
globs: .pi/generated/source/**/*.md, .pi/generated/source/**/*.yaml, .moai/**/*.yaml, .pi/generated/source/CLAUDE.md
---

# Coding Standards

MoAI-specific coding standards. General coding conventions are not included as Claude already knows them.

## Language Policy

All instruction documents must be in English:
- .pi/generated/source/CLAUDE.md
- Agent definitions (.pi/agents/overrides/**/*.md)
- Slash commands (.pi/generated/source/commands/**/*.md)
- Skill definitions (.pi/generated/source/skills/**/*.md)
- Hook scripts (.pi/generated/source/hooks/**/*.py, *.sh)
- Configuration files (.moai/config/**/*.yaml)

User-facing documentation may use multiple languages:
- README.md, CHANGELOG.md
- User guides, API documentation

## File Size Limits

.pi/generated/source/CLAUDE.md must not exceed 40,000 characters.

When approaching limit:
- Move detailed content to .pi/generated/source/rules/moai/
- Use @import references
- Keep only core identity and hard rules in .pi/generated/source/CLAUDE.md

## Content Restrictions

Prohibited in instruction documents:
- Code examples for conceptual explanations
- Flow control as code syntax
- Decision trees as code structures
- Emoji characters (except output styles)
- Time estimates or duration predictions

## Duplicate Prevention

Single source of truth principle:
- Each piece of information exists in exactly one location
- Use references (@file) instead of copying content
- Update source file, not copies

## Paths Frontmatter

Use paths frontmatter for conditional rule loading:

```yaml
---
paths: "**/*.py,**/pyproject.toml"
---
```

This ensures rules load only when working with matching files.
