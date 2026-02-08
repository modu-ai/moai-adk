# Skill Authoring

Guidelines for creating MoAI skills following the Agent Skills open standard (agentskills.io).

## YAML Frontmatter Schema

MoAI skills follow the Agent Skills standard with MoAI-specific extensions.

### Standard Fields (agentskills.io)

Required fields:
- name: Skill identifier, lowercase with hyphens, max 64 characters (system: moai-{category}-{name}, user: custom-{name})
- description: Purpose description using YAML folded scalar (>), max 1024 characters

Optional standard fields:
- license: SPDX license identifier (default: Apache-2.0)
- compatibility: Target platform description, max 500 characters (default: Designed for Claude Code)
- allowed-tools: Space-delimited string of tool names the skill can use (experimental)
- user-invocable: Boolean to control slash command menu visibility (default: true, set to false to hide from / menu)

### metadata Map

Key-value pairs where both keys and values MUST be strings. Used for simple custom properties.

Common metadata keys:
- version: Semantic version as string (e.g., "1.0.0")
- category: foundation, workflow, domain, language, platform, library, tool
- status: active, experimental, deprecated
- updated: ISO date as string (e.g., "2026-01-28")
- modularized: Whether content is split into modules ("true" or "false")
- tags: Comma-separated tag list as single string
- author: Skill author name
- context7-libraries: Comma-separated library identifiers for Context7 MCP
- related-skills: Comma-separated related skill names
- aliases: Comma-separated alternative names
- argument-hint: Usage hint for user-invocable skills
- context: Contextual description for skill behavior
- agent: Target agent name

### MoAI Extension Fields

Complex structured fields kept at top level with standardized comments.

progressive_disclosure: Token optimization configuration
- enabled: boolean
- level1_tokens: approximate tokens for metadata level
- level2_tokens: approximate tokens for body level

triggers: Loading trigger conditions
- keywords: list of trigger keywords
- agents: list of agent names that load this skill
- phases: list of workflow phases
- languages: list of programming languages

### Schema Example

```yaml
---
name: moai-example-skill
description: >
  Brief description of what this skill does, max 1024 characters.
  Use YAML folded scalar (>) for multi-line descriptions.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Grep Glob Bash mcp__context7__resolve-library-id mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "1.0.0"
  category: "domain"
  status: "active"
  updated: "2026-01-28"
  modularized: "false"
  tags: "example, demo, template"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["example", "demo"]
  agents: ["expert-backend"]
  phases: ["run"]
---
```

### Key Format Rules

allowed-tools format: Space-delimited string, not YAML array.
- Correct: `allowed-tools: Read Grep Glob Bash`
- Wrong: `allowed-tools: [Read, Grep, Glob, Bash]`

metadata values: All values must be quoted strings.
- Correct: `version: "1.0.0"`
- Wrong: `version: 1.0.0`

description format: Use YAML folded scalar (>) for readability.
- Correct: `description: >\n  Multi-line description here.`
- Wrong: `description: "Long description in quotes"`

## Progressive Disclosure

Three-level system for token efficiency:

Level 1 (Metadata):
- Tokens: ~100
- Content: name, description, version, triggers
- Loading: Always for skills in agent frontmatter

Level 2 (Body):
- Tokens: ~5000
- Content: Full documentation, code examples
- Loading: When trigger conditions match

Level 3 (Bundled):
- Tokens: Variable
- Content: reference.md, modules/, examples/
- Loading: On-demand by Claude

## Tool Permissions by Category

Foundation Skills:
- Allowed: Read, Grep, Glob, Context7 MCP
- Never: Bash, Task

Workflow Skills:
- Allowed: Read, Write, Edit, Grep, Glob, Bash, TodoWrite
- Conditional: AskUserQuestion (MoAI only), Task (managers only)

Domain Skills:
- Allowed: Read, Grep, Glob, Bash
- Conditional: Write, Edit (implementation tasks only)
- Never: AskUserQuestion, Task

Language Skills:
- Allowed: Read, Grep, Glob, Bash, Context7 MCP
- Conditional: Write, Edit (implementation tasks only)
- Never: AskUserQuestion, Task

## Trigger Configuration

```yaml
triggers:
  keywords: ["api", "database", "authentication"]
  agents: ["manager-spec", "expert-backend"]
  phases: ["plan", "run"]
  languages: ["python", "typescript"]
```

## Best Practices

- Use minimum required permissions
- Prefer Read before Write/Edit operations
- Prefer Edit over Bash for file modifications
- Include 5-10 keywords per skill for accurate triggering
- Overestimate token usage by 10-20% for safety
- Use YAML folded scalar (>) for description field
- Keep all metadata values as quoted strings
- Use space-delimited format for allowed-tools
- Mark MoAI extension fields with standardized comments
