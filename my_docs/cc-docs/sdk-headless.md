# Claude Code Headless SDK

## Overview

The Headless SDK allows you to run Claude Code programmatically without the interactive UI, perfect for automation, CI/CD pipelines, and scripting. It provides the same powerful capabilities as the interactive mode through command-line interfaces and configuration options.

## Key Features

- **Non-Interactive Mode**: Run Claude Code without user prompts
- **Multiple Output Formats**: Text, JSON, or streaming JSON
- **Session Management**: Resume conversations and manage context
- **Tool Control**: Fine-grained permission management
- **Automation Ready**: Perfect for scripts and CI/CD pipelines

## Installation

```bash
# Via native installer (recommended)
curl -fsSL claude.ai/install.sh | bash

# Via NPM (requires Node.js 18+)
npm install -g @anthropic-ai/claude-code
```

## Authentication

Set up your API key:

```bash
export ANTHROPIC_API_KEY="your-api-key-here"
```

**Alternative Providers:**

```bash
export CLAUDE_CODE_USE_BEDROCK=1    # For AWS Bedrock
export CLAUDE_CODE_USE_VERTEX=1     # For Google Vertex AI
```

## Basic Usage

### One-Shot Queries

```bash
# Simple query
claude -p "explain this error message"

# With piped input
cat error.log | claude -p "analyze this error and suggest fixes"

# Specify output format
claude -p "list all TODO comments" --output-format json
```

### Session Management

```bash
# Start with initial prompt
> "help me debug the authentication flow"

# Continue most recent session
claude --continue

# Resume specific session by ID
claude --resume "550e8400-e29b-41d4-a716-446655440000" "continue debugging"
```

## Command-Line Options

### Essential Flags

| Flag                | Description               | Example                         |
| ------------------- | ------------------------- | ------------------------------- |
| `-p, --print`       | Non-interactive mode      | `claude -p "task"`              |
| `-c, --continue`    | Continue recent session   | `claude -c`                     |
| `-r, --resume`      | Resume specific session   | `claude -r session-id`          |
| `--output-format`   | Set output format         | `--output-format json`          |
| `--allowedTools`    | Restrict tool access      | `--allowedTools "Read,Bash"`    |
| `--permission-mode` | Set permission level      | `--permission-mode acceptEdits` |
| `--max-turns`       | Limit conversation rounds | `--max-turns 5`                 |
| `--verbose`         | Enable debug output       | `--verbose`                     |

### Output Formats

#### Text (Default)

```bash
claude -p "summarize this file" < README.md
```

Human-readable output, perfect for terminal display.

#### JSON

```bash
claude -p "analyze code complexity" --output-format json
```

Structured output with metadata:

```json
{
  "result": "The code has moderate complexity...",
  "usage": {
    "input_tokens": 150,
    "output_tokens": 75
  },
  "session_id": "abc123..."
}
```

#### Streaming JSON

```bash
claude -p "refactor this function" --output-format stream-json
```

Line-by-line JSON messages for real-time processing:

```json
{"type": "thinking", "content": "Analyzing function..."}
{"type": "tool_use", "tool": "Read", "status": "started"}
{"type": "result", "content": "Here's the refactored version..."}
```

## Advanced Configuration

### Permission Modes

#### `ask` - Request Approval

```bash
claude -p "modify configuration files" --permission-mode ask
```

Prompts for each tool use (not practical for headless).

#### `acceptEdits` - Auto-Approve File Changes

```bash
claude -p "fix formatting issues" --permission-mode acceptEdits
```

Automatically approves Read/Edit/Write, asks for Bash/WebFetch.

#### `acceptAll` - Full Automation

```bash
claude -p "run tests and fix failures" --permission-mode acceptAll
```

Approves all tool uses automatically.

#### `bypassPermissions` - Advanced Use

```bash
claude -p "system maintenance tasks" --permission-mode bypassPermissions
```

Bypasses permission system entirely (use with caution).

### Tool Restrictions

#### Specific Tools Only

```bash
claude -p "analyze codebase structure" --allowedTools "Read,Grep,Glob"
```

#### Safe Development Mode

```bash
claude -p "refactor components" --allowedTools "Read,Edit,MultiEdit"
```

#### Read-Only Analysis

```bash
claude -p "security audit" --allowedTools "Read,Grep,Bash(git:*)"
```

## Practical Examples

### Code Review Automation

```bash
#!/bin/bash
# review_changes.sh

# Get list of changed files
CHANGED_FILES=$(git diff --name-only HEAD~1)

# Review each file
echo "$CHANGED_FILES" | while read file; do
  echo "Reviewing $file..."
  claude -p "Review this file for security issues and code quality: $file" \
    --allowedTools "Read,Grep" \
    --output-format json \
    --max-turns 2 > "review_$file.json"
done
```

### Documentation Generation

```bash
#!/bin/bash
# generate_docs.sh

# Find all TypeScript files
find src/ -name "*.ts" -not -path "*/node_modules/*" | while read file; do
  claude -p "Generate comprehensive JSDoc comments for all functions in this file" \
    --allowedTools "Read,Edit" \
    --permission-mode acceptEdits \
    "$file"
done
```

### CI/CD Integration

```bash
# .github/workflows/ai-review.yml
name: AI Code Review

on: [pull_request]

jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Install Claude Code
        run: curl -fsSL claude.ai/install.sh | bash

      - name: Review Changes
        env:
          ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
        run: |
          claude -p "Review the changes in this PR for potential issues" \
            --allowedTools "Read,Bash(git:*)" \
            --output-format json \
            --max-turns 3 > review.json
```

### Automated Testing

```bash
#!/bin/bash
# smart_test.sh

# Run tests and analyze failures
claude -p "Run the test suite and fix any failures you find" \
  --allowedTools "Bash,Read,Edit" \
  --permission-mode acceptAll \
  --max-turns 10 \
  --verbose
```

### Log Analysis

```bash
# Analyze application logs
tail -n 1000 /var/log/app.log | \
  claude -p "Analyze these logs for errors and performance issues" \
    --allowedTools "Read" \
    --output-format json
```

## Session Management

### Understanding Sessions

- Each conversation creates a unique session ID
- Sessions maintain context between turns
- Use `--continue` for most recent session
- Use `--resume session-id` for specific sessions

### Session Discovery

```bash
# Find session IDs
claude config list | grep session

# Or use status command in interactive mode
claude
> /status
```

### Long-Running Sessions

```bash
# Start a debugging session
SESSION_ID=$(> "Start debugging the authentication system" --output-format json | jq -r '.session_id')

# Continue working
claude --resume "$SESSION_ID" "Check the login endpoint"
claude --resume "$SESSION_ID" "Fix any security vulnerabilities"
claude --resume "$SESSION_ID" "Add comprehensive tests"
```

## Error Handling

### Exit Codes

- `0`: Success
- `1`: General error (authentication, invalid arguments)
- `2`: Interrupted by user
- `130`: System interrupt (SIGINT)

### Error Recovery

```bash
#!/bin/bash
# robust_task.sh

set -e  # Exit on error

if claude -p "perform complex task" --max-turns 5; then
    echo "Task completed successfully"
else
    echo "Task failed, attempting recovery"
    claude -p "diagnose and fix the previous task failure" --max-turns 3
fi
```

### Timeout Management

```bash
# Set timeouts for long operations
timeout 300s claude -p "run comprehensive analysis" || echo "Analysis timed out"
```

## Integration Patterns

### With Make

```makefile
# Makefile
.PHONY: ai-review ai-docs ai-test

ai-review:
	claude -p "Review recent changes for issues" \
		--allowedTools "Read,Bash(git:*)" \
		--output-format json > review.json

ai-docs:
	claude -p "Update documentation based on recent code changes" \
		--allowedTools "Read,Edit,Glob" \
		--permission-mode acceptEdits

ai-test:
	claude -p "Run tests and fix failures" \
		--allowedTools "Bash,Read,Edit" \
		--permission-mode acceptAll
```

### With Docker

```dockerfile
# Dockerfile
FROM node:18-alpine

RUN curl -fsSL claude.ai/install.sh | bash

COPY scripts/ /scripts/
WORKDIR /app

CMD ["claude", "-p", "perform automated code review", "--allowedTools", "Read,Grep"]
```

### Environment Configuration

```bash
# .env for headless operations
export ANTHROPIC_API_KEY="your-key"
export CLAUDE_CODE_PERMISSION_MODE="acceptEdits"
export CLAUDE_CODE_MAX_TURNS="5"
export CLAUDE_CODE_ALLOWED_TOOLS="Read,Edit,Bash"
```

## Best Practices

### Security

1. **Limit Tool Access**: Use `--allowedTools` to restrict capabilities
2. **Permission Modes**: Choose appropriate automation level
3. **Credential Management**: Never log API keys
4. **Input Validation**: Sanitize inputs in scripts

### Performance

1. **Set Max Turns**: Prevent infinite loops with `--max-turns`
2. **Use Specific Tools**: Avoid broad permissions
3. **Optimize Prompts**: Clear, specific instructions work better
4. **Stream Large Outputs**: Use streaming JSON for real-time feedback

### Reliability

1. **Error Handling**: Check exit codes and handle failures
2. **Timeouts**: Set reasonable limits for operations
3. **Logging**: Use `--verbose` for debugging
4. **Session Management**: Save session IDs for complex workflows

## Troubleshooting

### Common Issues

#### Permission Denied

```bash
# Check current permission mode
claude config get defaultMode

# Override for single operation
claude -p "task" --permission-mode acceptAll
```

#### Tool Not Available

```bash
# List available tools
claude -p "what tools are available?" --allowedTools "TodoWrite"

# Check tool restrictions
claude config get permissions.deny
```

#### Session Not Found

```bash
# List recent sessions
ls ~/.claude/projects/$(basename $(pwd))/

# Start new session if needed
> "restart the debugging process"
```

### Debug Commands

```bash
claude --verbose -p "task"          # Detailed logging
claude /status                      # Check system status (interactive)
claude /doctor                      # Diagnose installation (interactive)
```

## Advanced Usage

### Parallel Processing

```bash
# Process multiple files in parallel
find src/ -name "*.js" | xargs -P 4 -I {} \
  claude -p "optimize this file: {}" --allowedTools "Read,Edit"
```

### Custom Workflows

```bash
#!/bin/bash
# ai_workflow.sh

# Multi-step automation
echo "Step 1: Analysis"
claude -p "analyze project structure" --allowedTools "Read,Glob" > analysis.txt

echo "Step 2: Implementation"
ANALYSIS=$(cat analysis.txt)
claude -p "Based on this analysis: $ANALYSIS, implement improvements" \
  --allowedTools "Read,Edit,Bash" \
  --permission-mode acceptAll

echo "Step 3: Testing"
claude -p "run tests and verify improvements" \
  --allowedTools "Bash,Read" \
  --max-turns 3
```

This headless SDK enables powerful automation while maintaining the same capabilities as interactive Claude Code sessions.
