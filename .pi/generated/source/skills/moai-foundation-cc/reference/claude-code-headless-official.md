# Claude Code Headless Mode - Official Documentation Reference

Source: https://code.claude.com/docs/en/headless
Updated: 2026-01-06

## Overview

Headless mode allows programmatic interaction with Claude Code without an interactive terminal interface. This enables CI/CD integration, automated workflows, and script-based usage.

## Basic Usage

### Simple Prompt

```bash
claude -p "Your prompt here"
```

The -p flag runs Claude with the given prompt and exits after completion.

### Continue Previous Conversation

```bash
claude -c "Follow-up question"
```

The -c flag continues the most recent conversation.

### Resume Specific Session

```bash
claude -r session_id "Continue this task"
```

The -r flag resumes a specific session by ID.

## Output Formats

### Plain Text (default)

```bash
claude -p "Explain this code" --output-format text
```

Returns response as plain text.

### JSON Output

```bash
claude -p "Analyze this" --output-format json
```

Returns structured JSON:
```json
{
  "result": "Response text",
  "session_id": "abc123",
  "usage": {
    "input_tokens": 100,
    "output_tokens": 200
  },
  "structured_output": null
}
```

### Streaming JSON

```bash
claude -p "Long task" --output-format stream-json
```

Returns JSON objects as they are generated, useful for real-time processing.

## Structured Output

### JSON Schema Validation

```bash
claude -p "Extract data" --json-schema '{"type": "object", "properties": {"name": {"type": "string"}}}'
```

Claude validates output against the provided JSON schema.

### Schema from File

```bash
claude -p "Process this" --json-schema-file schema.json
```

Loads schema from a file for complex structures.

## Tool Management

### Allow Specific Tools

```bash
claude -p "Build the project" --allowedTools "Bash,Read,Write"
```

Auto-approves the specified tools without prompts.

### Tool Pattern Matching

```bash
claude -p "Check git status" --allowedTools "Bash(git:*)"
```

Allow only specific command patterns.

### Multiple Patterns

```bash
claude -p "Review changes" --allowedTools "Bash(git diff:*),Bash(git status:*),Read"
```

Combine multiple tool patterns.

### Disallow Specific Tools

```bash
claude -p "Analyze code" --disallowedTools "Bash,Write"
```

Prevent Claude from using specified tools.

## System Prompt Configuration

### Replace System Prompt

```bash
claude -p "Task" --system-prompt "You are a code reviewer"
```

Completely replaces the default system prompt.

Warning: This removes Claude Code capabilities. Use --append-system-prompt instead unless you have specific requirements.

### Append to System Prompt

```bash
claude -p "Task" --append-system-prompt "Focus on security issues"
```

Adds instructions while preserving Claude Code functionality.

### System Prompt from File

```bash
claude -p "Task" --system-prompt-file prompt.txt
```

Loads system prompt from a file.

## Session Management

### Get Session ID

JSON output includes session_id for later reference:

```bash
result=$(claude -p "Start task" --output-format json)
session_id=$(echo $result | jq -r '.session_id')
```

### Fork Session

```bash
claude -p "Alternative approach" --fork-session abc123
```

Creates a new conversation branch from an existing session.

## Advanced Options

### Maximum Turns

```bash
claude -p "Complex task" --max-turns 10
```

Limits the number of conversation turns.

### Custom Agents

```bash
claude -p "Review code" --agent code-reviewer
```

Uses a specific sub-agent for the task.

### Dynamic Agent Definition

```bash
claude -p "Task" --agents '{
  "reviewer": {
    "description": "Code review specialist",
    "prompt": "You are an expert code reviewer",
    "tools": ["Read", "Grep", "Glob"],
    "model": "sonnet"
  }
}'
```

Defines sub-agents dynamically via JSON.

### Settings Override

```bash
claude -p "Task" --settings '{"model": "opus"}'
```

Overrides settings for this invocation.

### Show Setting Sources

```bash
claude --setting-sources
```

Displays where each setting value comes from.

## Browser Integration

### Enable Chrome Integration

```bash
claude -p "Browse this page" --chrome
```

Enables browser automation capabilities.

### Disable Chrome Integration

```bash
claude -p "Code task" --no-chrome
```

Explicitly disables browser features.

## CI/CD Integration Examples

### GitHub Actions

```yaml
- name: Code Review
  run: |
    claude -p "Review the changes in this PR" \
      --allowedTools "Read,Grep,Glob" \
      --output-format json > review.json
```

### Automated Commit Messages

```bash
git diff --staged | claude -p "Generate commit message for these changes" \
  --allowedTools "Read" \
  --append-system-prompt "Output only the commit message, no explanation"
```

### PR Description Generation

```bash
claude -p "Generate PR description" \
  --allowedTools "Bash(git diff:*),Bash(git log:*),Read" \
  --output-format json
```

### Structured Data Extraction

```bash
claude -p "Extract API endpoints from this codebase" \
  --allowedTools "Read,Grep,Glob" \
  --json-schema '{"type": "array", "items": {"type": "object", "properties": {"path": {"type": "string"}, "method": {"type": "string"}}}}'
```

## Agent SDK

For more programmatic control, use the Agent SDK:

### Python

```python
from anthropic import Claude

agent = Claude()
result = agent.run("Your task", tools=["Read", "Write"])
```

### TypeScript

```typescript
import { Claude } from '@anthropic-ai/sdk';

const agent = new Claude();
const result = await agent.run("Your task", { tools: ["Read", "Write"] });
```

### SDK Features

- Native structured outputs
- Tool approval callbacks
- Stream-based real-time output
- Full programmatic control
- Error handling and retry logic

## Environment Variables

### Configuration via Environment

```bash
export CLAUDE_MODEL=opus
export CLAUDE_OUTPUT_FORMAT=json
claude -p "Task"
```

### Available Variables

- CLAUDE_MODEL: Default model selection
- CLAUDE_OUTPUT_FORMAT: Default output format
- CLAUDE_TIMEOUT: Request timeout in seconds
- CLAUDE_API_KEY: API authentication

## Best Practices

### Use Append for System Prompts

Prefer --append-system-prompt over --system-prompt to retain Claude Code capabilities.

### Specify Tool Restrictions

Always use --allowedTools in CI/CD to prevent unintended actions.

### Handle Errors

Check exit codes and parse JSON output for error handling:

```bash
result=$(claude -p "Task" --output-format json 2>&1)
if [ $? -ne 0 ]; then
  echo "Error: $result"
  exit 1
fi
```

### Use Structured Output

For data extraction, use --json-schema to ensure consistent output format.

### Log Sessions

Store session IDs for debugging and continuity:

```bash
session_id=$(claude -p "Task" --output-format json | jq -r '.session_id')
echo "Session: $session_id" >> sessions.log
```

## Troubleshooting

### Command Hangs

If headless mode appears to hang:
- Check for permission prompts (use --allowedTools)
- Verify network connectivity
- Check API key configuration

### Unexpected Output Format

If output format is wrong:
- Verify --output-format flag spelling
- Check for conflicting environment variables
- Ensure JSON schema is valid if using --json-schema

### Tool Permission Denied

If tools are blocked:
- Verify tool names in --allowedTools
- Check pattern syntax for command restrictions
- Review enterprise policy restrictions
