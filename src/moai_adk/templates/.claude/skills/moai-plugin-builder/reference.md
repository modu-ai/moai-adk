# Plugin Builder Reference

## Complete Plugin Templates

### Minimal Plugin Template

Directory Structure:
```
minimal-plugin/
  .claude-plugin/
    plugin.json
  commands/
    hello.md
  README.md
```

plugin.json:
```json
{
  "name": "minimal-plugin",
  "version": "1.0.0",
  "description": "Minimal Claude Code plugin example",
  "commands": ["./commands"]
}
```

commands/hello.md:
```markdown
---
description: Simple hello world command
---

Say hello to the user with a friendly greeting.
If arguments provided: $ARGUMENTS
```

### Full-Featured Plugin Template

Directory Structure:
```
full-plugin/
  .claude-plugin/
    plugin.json
  commands/
    analyze.md
    generate.md
  agents/
    analyzer.md
    generator.md
  skills/
    domain-knowledge/
      SKILL.md
      reference.md
  hooks/
    hooks.json
    validate.sh
  output-styles/
    custom-style.md
  .mcp.json
  .lsp.json
  LICENSE
  CHANGELOG.md
  README.md
```

plugin.json:
```json
{
  "name": "full-plugin",
  "version": "1.0.0",
  "description": "Full-featured Claude Code plugin with all components",
  "author": {
    "name": "Plugin Author",
    "email": "author@example.com",
    "url": "https://example.com"
  },
  "homepage": "https://github.com/user/full-plugin",
  "repository": "https://github.com/user/full-plugin",
  "license": "MIT",
  "keywords": ["claude-code", "analysis", "generation"],
  "commands": ["./commands"],
  "agents": ["./agents"],
  "skills": ["./skills"],
  "hooks": ["./hooks/hooks.json"],
  "mcpServers": ["./.mcp.json"],
  "lspServers": ["./.lsp.json"],
  "outputStyles": ["./output-styles"]
}
```

## Component Templates

### Command Template

commands/analyze.md:
```markdown
---
description: Analyze codebase for patterns and issues
---

# Code Analysis Command

Analyze the codebase for:
1. Code quality patterns
2. Potential issues
3. Improvement opportunities

Target path: $1
Analysis type: $2

If no arguments provided, analyze the current project root.

Use the following approach:
- First, scan the directory structure
- Then, identify key files and patterns
- Finally, provide actionable recommendations

Reference configuration: @.claude/settings.json
```

### Agent Template

agents/analyzer.md:
```markdown
---
name: analyzer
description: Code analysis agent for identifying patterns and issues
tools: Read, Grep, Glob
model: sonnet
permissionMode: default
skills:
  - domain-knowledge
---

# Code Analyzer Agent

You are a code analysis specialist. Your role is to:

1. Identify code patterns and anti-patterns
2. Detect potential bugs and issues
3. Suggest improvements and optimizations
4. Evaluate code quality metrics

## Analysis Approach

When analyzing code:
- Start with high-level structure review
- Drill down into specific modules
- Cross-reference related components
- Provide specific, actionable feedback

## Output Format

Provide analysis in structured format:
- Summary of findings
- Detailed issues with file locations
- Prioritized recommendations
- Code quality score
```

### Skill Template

skills/domain-knowledge/SKILL.md:
```markdown
---
name: domain-knowledge
description: Domain-specific knowledge for analysis tasks
allowed-tools: Read, Grep, Glob
---

# Domain Knowledge Skill

## Quick Reference (30 seconds)

Key domain concepts and terminology for effective analysis.

## Implementation Guide

### Core Concepts

Define domain-specific patterns and conventions.

### Analysis Patterns

Standard approaches for domain analysis.

## Advanced Patterns

Expert-level domain knowledge and edge cases.
```

### Hook Template

hooks/hooks.json:
```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Write",
        "hooks": [
          {
            "type": "command",
            "command": "bash ${CLAUDE_PLUGIN_ROOT}/hooks/validate.sh \"$tool_input\""
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "prompt",
            "prompt": "Verify the command executed successfully and check for any errors in the output."
          }
        ]
      }
    ],
    "SessionStart": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "echo 'Plugin session initialized'"
          }
        ]
      }
    ]
  }
}
```

hooks/validate.sh:
```bash
#!/bin/bash

# Validate write operations
INPUT="$1"

# Check for dangerous patterns
if echo "$INPUT" | grep -qE "(rm -rf|sudo|chmod 777)"; then
  echo "BLOCKED: Potentially dangerous operation detected"
  exit 1
fi

# Check path traversal
if echo "$INPUT" | grep -qE "\.\.\/"; then
  echo "BLOCKED: Path traversal attempt detected"
  exit 1
fi

exit 0
```

### MCP Server Template

.mcp.json:
```json
{
  "mcpServers": {
    "plugin-server": {
      "command": "node",
      "args": ["${CLAUDE_PLUGIN_ROOT}/servers/mcp-server.js"],
      "env": {
        "NODE_ENV": "production",
        "PLUGIN_ROOT": "${CLAUDE_PLUGIN_ROOT}",
        "PROJECT_DIR": "${CLAUDE_PROJECT_DIR}"
      }
    },
    "external-api": {
      "command": "npx",
      "args": ["-y", "@example/mcp-server"],
      "env": {
        "API_KEY": "${API_KEY}",
        "API_SECRET": "${API_SECRET}"
      }
    }
  }
}
```

### LSP Server Template

.lsp.json:
```json
{
  "lspServers": {
    "typescript": {
      "command": "typescript-language-server",
      "args": ["--stdio"],
      "extensionToLanguage": {
        ".ts": "typescript",
        ".tsx": "typescriptreact",
        ".js": "javascript",
        ".jsx": "javascriptreact"
      },
      "env": {
        "TSC_WATCHFILE": "PriorityPollingInterval"
      }
    },
    "python": {
      "command": "pylsp",
      "args": [],
      "extensionToLanguage": {
        ".py": "python",
        ".pyi": "python"
      },
      "env": {
        "PYTHONPATH": "${CLAUDE_PROJECT_DIR}:${CLAUDE_PROJECT_DIR}/src"
      }
    }
  }
}
```

### Output Style Template

output-styles/custom-style.md:
```markdown
---
name: custom-style
description: Custom output formatting style
---

# Custom Output Style

## Response Format

Structure responses with:
- Clear section headers
- Bulleted key points
- Code blocks for technical content
- Summary at the end

## Tone

Professional yet approachable. Focus on clarity and actionability.

## Formatting Rules

- Use markdown formatting
- Include file paths in code blocks
- Highlight important warnings
- Provide next steps
```

## Validation Rules

### plugin.json Validation

Required Structure:
- Must be valid JSON
- Must contain "name" field
- Name must be kebab-case (lowercase, hyphens, numbers only)
- Name must be unique across installed plugins

Optional Field Validation:
- version: Must follow semver (X.Y.Z)
- author.email: Must be valid email format
- homepage, repository: Must be valid URLs
- license: Should be SPDX identifier

Path Validation:
- All paths must start with "./"
- Paths must be relative to plugin root
- Paths must not contain ".." traversal
- Referenced files/directories must exist

### Component Validation

Commands:
- Must have .md extension
- Must contain YAML frontmatter
- Must have description in frontmatter
- Markdown body must not be empty

Agents:
- Must have .md extension
- Must contain name in frontmatter
- Tools must be valid tool names
- Model must be: sonnet, opus, haiku, inherit
- permissionMode must be: default, bypassPermissions, plan, passthrough

Skills:
- Must have SKILL.md file
- Must contain name in frontmatter
- Must contain description in frontmatter
- Must be under 500 lines

Hooks:
- Must be valid JSON
- Event names must be valid hook events
- Matcher must be string or array
- Hook type must be: command, prompt, agent

### Server Validation

MCP Servers:
- Must contain command field
- Command must be executable
- Args must be array of strings
- Env values must be strings

LSP Servers:
- Must contain command field
- Must contain extensionToLanguage mapping
- Extensions must start with "."
- Language values must be strings
- Optional advanced fields: transport, initializationOptions, settings, workspaceFolder, startupTimeout, shutdownTimeout, restartOnCrash, maxRestarts, loggingConfig

## Environment Variables

Available Variables:
- `${CLAUDE_PLUGIN_ROOT}` - Absolute path to plugin directory
- `${CLAUDE_PROJECT_DIR}` - Absolute path to current project
- `${HOME}` - User home directory
- `${USER}` - Current username

Variable Usage:
- Supported in all path configurations
- Supported in env values
- Supported in command arguments
- Resolved at plugin load time

## File Organization Best Practices

Recommended Structure:
```
plugin-name/
  .claude-plugin/
    plugin.json         # Manifest only
  commands/
    feature-one.md      # Group by feature
    feature-two.md
  agents/
    specialist.md       # Named by role
  skills/
    knowledge-area/     # Directory per skill
      SKILL.md
      reference.md
      examples.md
  hooks/
    hooks.json          # Main config
    scripts/            # Hook scripts
      validate.sh
      notify.sh
  servers/
    mcp-server.js       # Custom servers
  output-styles/
    style.md
  .mcp.json             # MCP config
  .lsp.json             # LSP config
  LICENSE
  CHANGELOG.md
  README.md
```

Naming Conventions:
- Plugin name: kebab-case, descriptive
- Command files: action-noun.md
- Agent files: role-name.md
- Skill directories: knowledge-area
- Script files: action-verb.sh
