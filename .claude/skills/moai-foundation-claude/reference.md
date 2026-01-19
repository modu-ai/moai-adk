# moai-foundation-claude Reference

## API Reference

### Skill Definition API

Frontmatter Fields:
- `name` (required): Skill identifier in kebab-case, max 64 characters
- `description` (required): One-line description, max 1024 characters
- `version`: Semantic version (e.g., "2.0.0")
- `tools`: Comma-separated list of allowed tools
- `modularized`: Boolean indicating modular file structure
- `category`: Skill category (foundation, domain, workflow, library, integration)
- `tags`: Array of searchable keywords
- `aliases`: Alternative names for skill invocation

### Sub-agent Delegation API

Task Invocation:
- `Task(subagent_type, prompt)`: Invoke specialized sub-agent
- `Task(subagent_type, prompt, context)`: Invoke with context from previous task
- Returns structured result object for chaining

Available Sub-agent Types:
- `spec-builder`: EARS format specification generation
- `ddd-implementer`: ANALYZE-PRESERVE-IMPROVE DDD execution
- `backend-expert`: Backend architecture and API development
- `frontend-expert`: Frontend UI implementation
- `security-expert`: Security analysis and validation
- `docs-manager`: Technical documentation generation
- `quality-gate`: TRUST 5 validation
- `agent-factory`: Create new sub-agents
- `skill-factory`: Create compliant skills

### Command Parameter API

Parameter Types:
- `$1`, `$2`, `$3`: Positional arguments
- `$ARGUMENTS`: All arguments as single string
- `@filename`: File content injection

Command Location:
- Personal: `~/.claude/commands/`
- Project: `.claude/commands/`

---

## Configuration Options

### Settings Hierarchy

Priority Order (highest to lowest):
1. Enterprise settings (`/etc/claude/settings.json`)
2. User settings (`~/.claude/settings.json`)
3. Project settings (`.claude/settings.json`)
4. Local settings (`.claude/settings.local.json`)

### Tool Permissions

Permission Levels:
- `Read, Grep, Glob`: Read-only access for analysis
- `Read, Write, Edit, Grep, Glob`: Full file manipulation
- `Bash`: System command execution (requires explicit grant)
- `WebFetch, WebSearch`: External web access

### Memory Configuration

Memory File Locations:
- Enterprise: `/etc/claude/CLAUDE.md`
- User: `~/.claude/CLAUDE.md`
- Project: `./CLAUDE.md` or `.claude/CLAUDE.md`

Memory Import Syntax:
```markdown
@import path/to/file.md
```

---

## Integration Patterns

### Command-Agent-Skill Orchestration

Sequential Pattern:
1. Command receives user input with `$ARGUMENTS`
2. Command loads relevant Skills via `Skill("skill-name")`
3. Command delegates to sub-agent via `Task(subagent_type, prompt)`
4. Sub-agent executes with loaded skill context
5. Result returned to command for presentation

Parallel Pattern:
- Multiple independent `Task()` calls execute concurrently
- Results aggregated after all complete
- Use when tasks have no dependencies

### Hook Integration

PreToolUse Hooks:
- Execute before any tool invocation
- Can block or modify tool execution
- Use for validation, logging, security checks

PostToolUse Hooks:
- Execute after tool completion
- Can process or modify results
- Use for backup, audit, notification

Hook Configuration (settings.json):
```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "ToolName",
        "hooks": [{"type": "command", "command": "hook-script"}]
      }
    ]
  }
}
```

### MCP Server Integration

Context7 Integration:
- Use for real-time documentation lookup
- Two-step pattern: resolve library ID, then fetch docs
- Supports token-limited responses

MCP Tool Invocation:
- Tools prefixed with `mcp__` for MCP-provided capabilities
- Server configuration in settings.json

---

## Troubleshooting

### Skill Not Loading

Symptoms: Skill not recognized, missing context

Solutions:
1. Verify file location (`~/.claude/skills/` or `.claude/skills/`)
2. Check SKILL.md frontmatter syntax (valid YAML)
3. Confirm name follows kebab-case, max 64 chars
4. Verify file size under 500 lines

### Sub-agent Delegation Failures

Symptoms: Task() returns error, incomplete results

Solutions:
1. Verify subagent_type is valid
2. Check prompt clarity and specificity
3. Ensure required context is provided
4. Review token budget (each Task() gets 200K)

### Hook Not Executing

Symptoms: PreToolUse/PostToolUse not triggering

Solutions:
1. Check matcher pattern matches tool name exactly
2. Verify hook script exists and is executable
3. Review settings.json syntax
4. Check command permissions

### Memory File Issues

Symptoms: CLAUDE.md content not applied

Solutions:
1. Verify file location in correct hierarchy
2. Check file encoding (UTF-8 required)
3. Review @import paths (relative to file)
4. Ensure file permissions allow reading

---

## External Resources

### Official Documentation

- [Claude Code Skills Guide](https://docs.anthropic.com/claude-code/skills)
- [Sub-agents Documentation](https://docs.anthropic.com/claude-code/agents)
- [Custom Commands Reference](https://docs.anthropic.com/claude-code/commands)
- [Hooks System Guide](https://docs.anthropic.com/claude-code/hooks)
- [Memory Management](https://docs.anthropic.com/claude-code/memory)
- [Settings Configuration](https://docs.anthropic.com/claude-code/settings)
- [IAM and Permissions](https://docs.anthropic.com/claude-code/iam)

### Best Practices

- Keep SKILL.md under 500 lines
- Use progressive disclosure (Quick, Implementation, Advanced)
- Apply least-privilege tool permissions
- Document trigger scenarios in description
- Include working examples for each pattern

### Related Skills

- `moai-foundation-core`: Core execution patterns and SPEC workflow
- `moai-foundation-context`: Token budget and session management
- `moai-workflow-project`: Project initialization and configuration
- `moai-docs-generation`: Documentation automation

---

Version: 2.0.0
Last Updated: 2025-12-06
