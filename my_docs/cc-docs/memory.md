# Memory System

Manage Claude's persistent memory across different contexts and hierarchies.

## Overview

Claude Code's memory system provides layered, hierarchical memory management through CLAUDE.md files. This system enables teams and individuals to share context, instructions, and preferences across different scopes while maintaining clear precedence rules.

## Memory Hierarchy

Claude Code uses a 4-level memory hierarchy, loaded in priority order:

| Priority | Memory Type | Location | Purpose | Shared With |
|----------|-------------|----------|---------|-------------|
| **1** | **Enterprise Policy** | System paths (see below) | Organization-wide instructions managed by IT/DevOps | All users in organization |
| **2** | **Project Memory** | `./CLAUDE.md` | Team-shared instructions for the project | Team members via source control |
| **3** | **User Memory** | `~/.claude/CLAUDE.md` | Personal preferences for all projects | Just you (all projects) |
| **4** | **Project Memory (Local)** | `./CLAUDE.local.md` | **Deprecated** - Personal project-specific preferences | Just you (current project) |

### Enterprise Policy Locations

Enterprise managed policy settings are deployed at:

- **macOS**: `/Library/Application Support/ClaudeCode/CLAUDE.md`
- **Linux/WSL**: `/etc/claude-code/CLAUDE.md`
- **Windows**: `C:\ProgramData\ClaudeCode\CLAUDE.md`

## CLAUDE.md Import System

CLAUDE.md files support powerful import functionality using `@path/to/import` syntax:

### Basic Import Examples

```markdown
See @README for project overview and @package.json for available npm commands for this project.

# Additional Instructions
- git workflow @docs/git-instructions.md
```

### Advanced Import Features

```markdown
# Individual Preferences
- @~/.claude/my-project-instructions.md

# Import from home directory
@~/.claude/coding-standards.md

# Relative path imports
@docs/api-guidelines.md
@../shared/team-conventions.md
```

### Import Rules and Limitations

- **Path Support**: Both relative and absolute paths are supported
- **Recursive Imports**: Maximum depth of 5 hops to prevent infinite loops
- **Code Block Exclusion**: Imports are not evaluated inside code spans or code blocks
- **Home Directory**: Can import files from user's home directory (`~/`)

```markdown
This code span will not be treated as an import: `@anthropic-ai/claude-code`

```bash
# This will also not be treated as an import
@some/path/file.md
```
```

## Memory Lookup Process

Claude Code discovers and loads memory files through a systematic process:

1. **Start Location**: Current working directory
2. **Tree Traversal**: Recursively searches up directory tree (excluding root)
3. **File Discovery**: Locates any CLAUDE.md or CLAUDE.local.md files
4. **Subtree Loading**: Subtree CLAUDE.md files are loaded only when working in those specific directories
5. **Hierarchy Application**: Applies memory hierarchy rules with proper precedence

## Memory Management Commands

### Interactive Memory Management

```bash
# View and edit loaded memories
/memory

# Initialize project memory
/init

# View memory status and loaded files
/memory
```

### Quick Memory Addition

```bash
# Start input with '#' for quick memory addition
# Always use descriptive variable names
```

When you start an input with `#`, Claude Code prompts you to select which memory file to store the instruction in.

### Configuration Commands

```bash
# List all settings
claude config list

# Get specific setting
claude config get <key>

# Set configuration value
claude config set <key> <value>

# Add to list settings
claude config add <key> <value>

# Remove from list settings
claude config remove <key> <value>
```

## Advanced Memory Patterns

### Multi-File Organization

Break complex instructions into manageable, focused files:

```markdown
# Main CLAUDE.md
@docs/coding-standards.md
@docs/git-workflow.md
@docs/testing-guidelines.md
@docs/deployment-procedures.md

# Project-specific instructions
When working on authentication features, refer to @docs/auth-security.md
```

### Team Collaboration Patterns

```markdown
# Project CLAUDE.md (shared via git)
# Team coding standards and project conventions
@docs/team-standards.md

# Individual developer preferences (not committed)
@~/.claude/personal-preferences.md
```

### Conditional Instructions

```markdown
# Environment-specific instructions
For staging deployments: @deployment/staging-guide.md
For production deployments: @deployment/production-guide.md

# Role-based instructions
Frontend developers: @guides/frontend-setup.md
Backend developers: @guides/backend-setup.md
```

## Best Practices

### Organize by Scope

1. **Enterprise Level**: Company-wide policies, security requirements, compliance
2. **Project Level**: Team conventions, architecture decisions, shared workflows
3. **User Level**: Personal preferences, coding style, tool configurations
4. **Deprecated Local**: Avoid using CLAUDE.local.md, use home directory imports instead

### Memory Content Guidelines

- **Be Specific**: Use precise, actionable instructions rather than vague guidelines
- **Use Structure**: Organize with clear markdown headers and sections
- **Import Strategically**: Break large instructions into focused, reusable files
- **Version Control**: Include project memories in git, exclude personal preferences

### Effective Import Strategies

```markdown
# Good: Focused, specific imports
@coding-standards/typescript.md
@workflows/testing.md
@deployment/docker.md

# Avoid: Overly broad or nested imports
@everything/all-instructions.md
```

### Memory Maintenance

- **Regular Review**: Periodically review and update memory files
- **Remove Outdated**: Clean up deprecated instructions and broken imports
- **Team Sync**: Ensure project memories reflect current team agreements
- **Documentation**: Document memory organization for team members

## Common Use Cases

### Development Workflow Integration

```markdown
# Git workflow integration
@git-conventions.md

# Testing requirements
When writing tests, follow @testing/jest-patterns.md
Always run @scripts/pre-commit-checks.sh before committing
```

### Code Quality Standards

```markdown
# Code review checklist
@quality/review-checklist.md

# Performance guidelines
@performance/optimization-guide.md

# Security requirements
@security/coding-standards.md
```

### Project Onboarding

```markdown
# New developer setup
1. Review @onboarding/getting-started.md
2. Install tools from @setup/dev-environment.md
3. Understand architecture via @docs/system-overview.md
```

## Integration with Other Features

### MCP Server Integration

Memory files can reference MCP server configurations:

```markdown
# MCP integration instructions
Use Linear MCP for issue tracking: @mcp/linear-workflow.md
Notion integration for documentation: @mcp/notion-procedures.md
```

### Subagent Coordination

```markdown
# Subagent-specific instructions
For code reviews, the reviewer agent should follow @agents/review-standards.md
Testing agents should use @agents/test-patterns.md
```

### Hooks Integration

```markdown
# Hook-related memory
Pre-commit hooks configuration: @hooks/pre-commit-setup.md
Post-tool execution procedures: @hooks/post-tool-guidelines.md
```

## Migration from Deprecated Features

### Replacing CLAUDE.local.md

Instead of using the deprecated `CLAUDE.local.md`, use home directory imports:

```markdown
# Old approach (deprecated)
# Instructions in ./CLAUDE.local.md

# New approach (recommended)
# Personal preferences
@~/.claude/project-specific-preferences.md
```

This approach provides better organization and avoids git conflicts while maintaining personal customization capabilities.

## Troubleshooting Memory Issues

### Common Problems

1. **Import Not Loading**: Check file paths and ensure maximum depth (5 hops) not exceeded
2. **Memory Conflicts**: Review hierarchy and ensure proper precedence
3. **Performance Issues**: Reduce deep import chains and circular references

### Debugging Commands

```bash
# Check loaded memory files
/memory

# Verify memory hierarchy
claude config list

# Test import resolution
# Add debug imports to verify path resolution
```

The memory system provides a powerful foundation for maintaining context and consistency across all Claude Code interactions, enabling seamless collaboration and personalized development experiences.