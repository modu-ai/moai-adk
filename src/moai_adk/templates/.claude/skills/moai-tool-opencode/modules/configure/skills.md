---
source: https://opencode.ai/docs/skills/
fetched: 2026-01-08
title: Agent Skills
---

# Agent Skills Documentation

## Overview

Agent Skills enables OpenCode to discover reusable instructions from repositories or home directories. These skills are loaded on-demand through a native `skill` tool, allowing agents to access available skills and retrieve full content when needed.

## File Organization

Skills are placed in dedicated folders with a `SKILL.md` file inside.

### Search Locations

OpenCode searches multiple locations for skills:

| Location | Scope |
|----------|-------|
| `.opencode/skill/<name>/SKILL.md` | Project |
| `~/.config/opencode/skill/<name>/SKILL.md` | Global |
| `.claude/skills/<name>/SKILL.md` | Claude Code compatible |
| `~/.claude/skills/<name>/SKILL.md` | Claude Code global |

## Required Frontmatter

Each `SKILL.md` must include YAML frontmatter:

```markdown
---
name: my-skill-name
description: A brief description of what this skill does
license: MIT
compatibility: opencode>=1.0.0
metadata:
  author: Your Name
  version: 1.0.0
---

# Skill Content

Instructions and knowledge go here...
```

### Frontmatter Fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Unique skill identifier |
| `description` | Yes | Brief description |
| `license` | No | License type |
| `compatibility` | No | Version requirements |
| `metadata` | No | Additional metadata |

## Name Validation Rules

Skill names must follow strict patterns:

- Be 1-64 characters
- Lowercase alphanumeric with single hyphens
- Not starting/ending with `-`
- No consecutive `--`

**Regex:** `^[a-z0-9]+(-[a-z0-9]+)*$`

**Important:** Names must match their containing directory.

### Valid Examples

- `my-skill`
- `code-review`
- `python-testing`
- `api-documentation`

### Invalid Examples

- `My-Skill` (uppercase)
- `-my-skill` (starts with hyphen)
- `my--skill` (consecutive hyphens)
- `my_skill` (underscore)

## Permission Control

Admins can restrict skill access in `opencode.json`:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "permission": {
    "skill": {
      "*": "allow",
      "dangerous-*": "deny",
      "experimental-*": "ask"
    }
  }
}
```

### Permission Levels

| Permission | Behavior |
|------------|----------|
| `allow` | Immediate access |
| `deny` | Hidden from agents |
| `ask` | User approval required |

Supports wildcards for pattern matching.

## Disabling Skills

### Per-Agent Disabling

Disable skills for specific agents:

```json
{
  "agent": {
    "simple-agent": {
      "skill": false
    }
  }
}
```

Or in agent markdown frontmatter:

```markdown
---
description: Agent without skills
skill: false
---
```

This completely removes the `<available_skills>` section.

## Creating Skills

### Directory Structure

```
.opencode/skill/my-skill/
├── SKILL.md
├── examples/
│   └── example.md
└── templates/
    └── template.md
```

### Example SKILL.md

```markdown
---
name: code-review
description: Comprehensive code review guidelines and checklists
---

# Code Review Skill

## Quick Start

When reviewing code, focus on:
1. Correctness
2. Security
3. Performance
4. Maintainability

## Detailed Guidelines

### Security Checklist
- [ ] Input validation
- [ ] Authentication checks
- [ ] Authorization verification
- [ ] SQL injection prevention
- [ ] XSS prevention

### Performance Checklist
- [ ] Efficient algorithms
- [ ] Proper caching
- [ ] Database query optimization
- [ ] Memory management
```

## Best Practices

1. **Clear naming**: Use descriptive, hyphenated names
2. **Good descriptions**: Help agents understand when to use skills
3. **Modular content**: Break complex skills into sections
4. **Version control**: Track skills in your repository
5. **Test skills**: Verify agents use skills correctly
