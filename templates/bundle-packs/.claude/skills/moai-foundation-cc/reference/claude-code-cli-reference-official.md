# Claude Code CLI Reference - Official Documentation Reference

Source: https://code.claude.com/docs/en/cli-reference
Updated: 2026-01-06

## Overview

The Claude Code CLI provides command-line access to Claude's capabilities with comprehensive options for customization, automation, and integration.

## Basic Commands

### Interactive Mode

```bash
claude
```

Starts Claude Code in interactive terminal mode.

### Direct Query

```bash
claude "Your question or task"
```

Sends a single query and enters interactive mode.

### Prompt Mode

```bash
claude -p "Your prompt"
```

Runs prompt, outputs response, and exits.

### Continue Conversation

```bash
claude -c "Follow-up"
```

Continues the most recent conversation.

### Resume Session

```bash
claude -r session_id "Continue task"
```

Resumes a specific session by ID.

### Update CLI

```bash
claude update
```

Updates Claude Code to the latest version.

## System Prompt Options

### Replace System Prompt

```bash
claude -p "Task" --system-prompt "Custom instructions"
```

Warning: Removes Claude Code default capabilities.

### Append to System Prompt

```bash
claude -p "Task" --append-system-prompt "Additional context"
```

Recommended: Preserves Claude Code functionality.

### Load from File

```bash
claude -p "Task" --system-prompt-file prompt.txt
```

Loads system prompt from external file.

## Tool Management

### Specify Tools

```bash
claude -p "Task" --tools "Read,Write,Bash"
```

Explicitly lists available tools.

### Allow Tools (Auto-approve)

```bash
claude -p "Task" --allowedTools "Read,Grep,Glob"
```

Auto-approves specified tools without prompts.

### Tool Pattern Matching

```bash
claude -p "Task" --allowedTools "Bash(git:*)"
```

Allow specific command patterns only.

### Multiple Patterns

```bash
claude -p "Task" --allowedTools "Bash(npm:*),Bash(git:*),Read"
```

### Disallow Tools

```bash
claude -p "Task" --disallowedTools "Bash,Write"
```

Prevents Claude from using specified tools.

## Output Options

### Output Format

```bash
claude -p "Task" --output-format text
claude -p "Task" --output-format json
claude -p "Task" --output-format stream-json
```

Available formats: text (default), json, stream-json

### JSON Schema Validation

```bash
claude -p "Extract data" --json-schema '{"type": "object"}'
```

Validates output against JSON schema.

### Schema from File

```bash
claude -p "Task" --json-schema-file schema.json
```

## Session Management

### Fork Session

```bash
claude -p "Alternative approach" --fork-session session_id
```

Creates a new branch from existing session.

### Maximum Turns

```bash
claude -p "Complex task" --max-turns 15
```

Limits conversation turns.

## Agent Configuration

### Use Specific Agent

```bash
claude -p "Review code" --agent code-reviewer
```

Uses defined sub-agent.

### Dynamic Agent Definition

```bash
claude -p "Task" --agents '{
  "my-agent": {
    "description": "Agent purpose",
    "prompt": "System prompt",
    "tools": ["Read", "Grep"],
    "model": "sonnet"
  }
}'
```

Defines agents inline via JSON.

## Settings

### Override Settings

```bash
claude -p "Task" --settings '{"model": "opus"}'
```

Overrides settings for this invocation.

### Show Setting Sources

```bash
claude --setting-sources
```

Displays origin of each setting value.

## Browser Integration

### Enable Chrome

```bash
claude -p "Browse task" --chrome
```

Enables browser automation.

### Disable Chrome

```bash
claude -p "Code task" --no-chrome
```

Disables browser features.

## MCP Server Commands

### Add MCP Server

HTTP transport:
```bash
claude mcp add --transport http server-name https://url
```

Stdio transport:
```bash
claude mcp add --transport stdio server-name command args
```

SSE transport (deprecated):
```bash
claude mcp add --transport sse server-name https://url
```

### List MCP Servers

```bash
claude mcp list
```

### Get Server Details

```bash
claude mcp get server-name
```

### Remove MCP Server

```bash
claude mcp remove server-name
```

## Plugin Commands

### Install Plugin

```bash
claude plugin install plugin-name
claude plugin install owner/repo
claude plugin install https://github.com/owner/repo.git
claude plugin install plugin-name --scope project
```

### Uninstall Plugin

```bash
claude plugin uninstall plugin-name
```

### Enable/Disable Plugin

```bash
claude plugin enable plugin-name
claude plugin disable plugin-name
```

### Update Plugin

```bash
claude plugin update plugin-name
claude plugin update  # Update all
```

### List Plugins

```bash
claude plugin list
```

### Validate Plugin

```bash
claude plugin validate .
```

## Environment Variables

### Configuration Variables

- CLAUDE_API_KEY: API authentication key
- CLAUDE_MODEL: Default model selection
- CLAUDE_OUTPUT_FORMAT: Default output format
- CLAUDE_TIMEOUT: Request timeout in seconds

### Runtime Variables

- CLAUDE_PROJECT_DIR: Current project directory
- CLAUDE_CODE_REMOTE: Indicates remote execution
- CLAUDE_ENV_FILE: Path to environment file

### MCP Variables

- MAX_MCP_OUTPUT_TOKENS: Maximum MCP output (default: 25000)
- MCP_TIMEOUT: MCP server timeout in milliseconds

### Update Control

- DISABLE_AUTOUPDATER: Disable automatic updates

## Exit Codes

- 0: Success
- 1: General error
- 2: Permission denied or blocked operation

## Complete Examples

### CI/CD Code Review

```bash
claude -p "Review this PR for security issues" \
  --allowedTools "Read,Grep,Glob" \
  --append-system-prompt "Focus on OWASP Top 10 vulnerabilities" \
  --output-format json \
  --max-turns 5
```

### Automated Documentation

```bash
claude -p "Generate API documentation for src/" \
  --allowedTools "Read,Glob,Write" \
  --json-schema-file docs-schema.json
```

### Structured Data Extraction

```bash
claude -p "Extract all function signatures from codebase" \
  --allowedTools "Read,Grep,Glob" \
  --json-schema '{"type":"array","items":{"type":"object","properties":{"name":{"type":"string"},"params":{"type":"array"},"returns":{"type":"string"}}}}'
```

### Git Commit Message

```bash
git diff --staged | claude -p "Generate commit message" \
  --allowedTools "Read" \
  --output-format text
```

### Multi-Agent Workflow

```bash
claude -p "Analyze and refactor this module" \
  --agents '{
    "analyzer": {
      "description": "Code analyzer",
      "tools": ["Read", "Grep"],
      "model": "haiku"
    },
    "refactorer": {
      "description": "Code refactorer",
      "tools": ["Read", "Write", "Edit"],
      "model": "sonnet"
    }
  }'
```

## Best Practices

### Security

- Use --allowedTools to restrict capabilities
- Avoid --dangerously-skip-permissions in untrusted environments
- Validate input before passing to Claude

### Performance

- Use appropriate --max-turns for task complexity
- Consider haiku model for simple tasks
- Use --output-format json for programmatic parsing

### Debugging

- Use --setting-sources to troubleshoot configuration
- Check exit codes for error handling
- Use --output-format json for detailed response metadata

### Automation

- Always specify --allowedTools in scripts
- Use --output-format json for reliable parsing
- Handle errors with exit code checks
- Log session IDs for debugging
