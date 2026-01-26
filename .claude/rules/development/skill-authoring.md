# Skill Authoring

Guidelines for creating MoAI skills, agents, and commands.

## YAML Frontmatter Schema

Required fields:
- name: Skill identifier (pattern: moai-{category}-{name})
- description: Purpose description (50-1024 characters)
- version: Semantic version (MAJOR.MINOR.PATCH)
- category: foundation, workflow, domain, language, platform, library, tool
- status: active, experimental, deprecated
- allowed-tools: Tools this skill can use

Optional fields:
- modularized: Whether content is split into modules
- user-invocable: Whether users can invoke via slash command
- progressive_disclosure: Token optimization config
- triggers: Loading trigger conditions

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
- Conditional: AskUserQuestion (Alfred only), Task (managers only)

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
