---
source: https://opencode.ai/docs/rules/
fetched: 2026-01-08
title: Rules and Instructions
---

# Rules Documentation - OpenCode

## Overview

OpenCode allows developers to customize LLM behavior through custom instructions. The primary mechanism is an `AGENTS.md` file that functions similarly to "CLAUDE.md" or Cursor's rules, containing project-specific guidance.

## Key Features

### Initialization

Running the `/init` command scans your project and generates an `AGENTS.md` file with auto-detected context.

```bash
/init
```

### File Locations

| Type | Location | Purpose |
|------|----------|---------|
| Project-level | `AGENTS.md` in repository root | Version-controlled team rules |
| Global | `~/.config/opencode/AGENTS.md` | Personal rules across projects |

### Configuration Support

The `opencode.json` file accepts an `instructions` field that references external rule files:

```json
{
  "$schema": "https://opencode.ai/config.json",
  "instructions": [
    "AGENTS.md",
    "packages/*/AGENTS.md",
    ".opencode/instructions/*.md"
  ]
}
```

Supports glob patterns for flexible file discovery.

## Implementation Example

A typical project `AGENTS.md` establishes:
- Monorepo structure
- Code standards
- Naming conventions
- Import conventions
- Framework-specific guidance

### Example AGENTS.md

```markdown
# Project Instructions

## Architecture
This is an SST v3 monorepo with the following structure:
- `packages/core/` - Shared business logic
- `packages/functions/` - Lambda handlers
- `packages/web/` - React frontend

## Code Standards
- Use TypeScript strict mode
- Prefer functional components
- Use barrel exports from index.ts

## Conventions
- Import order: external, internal, relative
- Use named exports over default exports
- File names: kebab-case
```

## Precedence Rules

OpenCode searches hierarchically:

1. Local files first (traversing upward from current directory)
2. Then global configuration

Both global and project rules combine when present.

## Advanced Patterns

### Lazy Loading

Developers can implement "lazy loading" of external references by instructing the AI to read specific files "on a need-to-know basis" rather than preemptively.

```markdown
## Extended Documentation
When working on authentication, read @docs/auth-patterns.md
When working on database, read @docs/db-conventions.md
```

This keeps configuration modular while maintaining focused context.

### Multiple Instruction Files

Use glob patterns to load multiple instruction files:

```json
{
  "instructions": [
    "AGENTS.md",
    ".opencode/rules/*.md",
    "packages/*/AGENTS.md"
  ]
}
```

### Conditional Rules

Structure rules for specific contexts:

```markdown
## When Working on API
- Use zod for validation
- Follow REST conventions
- Include error handling

## When Working on Frontend
- Use React Query for data fetching
- Follow component composition patterns
- Include loading states
```
