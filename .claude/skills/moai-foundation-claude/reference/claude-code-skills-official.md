# Claude Code Skills - Official Documentation Reference

Source: https://code.claude.com/docs/en/skills

## Key Concepts

### What are Claude Code Skills?

Skills are reusable pieces of functionality that extend Claude Code's capabilities. They can be automatically loaded based on context or manually invoked by users.

### Skill Types

1. Automatic Skills: Load automatically based on conversation context
2. Manual Skills: User must explicitly invoke using `/skill <skill-name>`
3. Plugin Skills: Bundled with installed packages

### Storage Locations (Priority Order)

1. Personal Skills: `~/.claude/skills/` (highest priority)
2. Project Skills: `.claude/skills/` (team-shared)
3. Plugin Skills: Bundled with installed packages (lowest priority)

### Required Frontmatter Fields

```yaml
---
name: my-skill # Required: kebab-case, max 64 chars
description: Brief description of what the skill does and when to trigger it
allowed-tools: Read, Write, Bash # Required: comma-separated tool list
---
```

### Optional Frontmatter Fields

```yaml
---
name: my-skill
description: What the skill does and when to trigger it
version: 1.0.0 # Semantic versioning
modularized: false # Whether skill can be split into modules
tags: [tag1, tag2] # For categorization
updated: 2025-11-25 # Last update date
status: active # active, deprecated, experimental
---
```

### Skill Discovery and Loading

- Automatic Loading: Skills load when Claude detects relevant context
- Progressive Loading: Supporting files load on-demand
- Tool Restrictions: Skills only have access to declared tools

### Best Practices

1. Single Responsibility: Each skill should have one clear purpose
2. Specific Descriptions: Include trigger scenarios in description
3. Minimal Tools: Declare only necessary tools using least privilege
4. Progressive Disclosure: Keep main SKILL.md concise; expand with supporting files
5. Working Examples: Include practical, tested examples

### File Structure Standards

```
skill-name/
 SKILL.md # Always create (main file, ≤ 500 lines)
 reference.md # Optional: Extended documentation
 examples.md # Optional: Code examples
 scripts/ # Optional: Utility scripts
 helper.sh
 templates/ # Optional: File templates
 template.md
```

### Skill Naming Conventions

- Format: kebab-case (lowercase with hyphens)
- Length: Maximum 64 characters
- Uniqueness: Must be unique across all storage locations
- Descriptive: Name should indicate skill purpose

### Tool Permissions

Skills only have access to tools declared in `allowed-tools`. Common tool categories:

- Information: `Read`, `Grep`, `Glob` (no user approval needed)
- Modification: `Write`, `Edit`, `Bash` (user approval required)
- External: `WebFetch`, `WebSearch` (user approval required)

### Integration Patterns

Skill Dependencies:
```yaml
skills: moai-foundation-claude, other-skill-name
```

MCP Server Integration:
```yaml
mcpServers:
 - context7
 - figma-dev-mode-mcp-server
```

### Validation and Debugging

Common Issues:
- Skill not loading: Check YAML syntax and file path
- Tool prompts unexpected: Review `allowed-tools` declaration
- Duplicate triggers: Rename skill or tighten description

Debug Commands:
- `claude --debug`: Check skill loading and configuration
- Validate YAML: Use online YAML validators or local tools
- Test skills: Create test cases for skill functionality
