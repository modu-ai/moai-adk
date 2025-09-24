# CLI Reference

Complete command-line interface reference for Claude Code with all available options, flags, and usage patterns.

## Core Commands

### Interactive Mode

| Command | Description | Example |
|---------|-------------|---------|
| `claude` | Start interactive REPL | `claude` |
| `claude "query"` | Start REPL with initial prompt | `claude "explain this project"` |
| `claude -c` / `claude --continue` | Continue most recent conversation | `claude -c` |
| `claude -r "<session-id>"` / `claude --resume` | Resume specific session | `claude -r "abc123" "continue task"` |

### Non-Interactive Mode

| Command | Description | Example |
|---------|-------------|---------|
| `claude -p "query"` / `claude --print` | Query via SDK and exit | `claude -p "explain this function"` |
| `cat file \| claude -p "query"` | Process piped content | `cat error.log \| claude -p "analyze"` |

### System Commands

| Command | Description | Example |
|---------|-------------|---------|
| `claude update` | Update to latest version | `claude update` |
| `claude mcp` | Configure Model Context Protocol servers | `claude mcp list` |
| `claude config` | Manage configuration settings | `claude config list` |
| `claude commit` | Direct git commit with AI assistance | `claude commit` |

## CLI Flags and Options

### Essential Flags

| Flag | Long Form | Description | Example |
|------|-----------|-------------|---------|
| `-p` | `--print` | Print response without interactive mode | `claude -p "explain this code"` |
| `-c` | `--continue` | Load most recent conversation | `claude -c` |
| `-r` | `--resume` | Resume specific session | `claude -r abc123 "continue task"` |
| `-v` | `--verbose` | Enable detailed logging | `claude -v` |
| `-h` | `--help` | Show help information | `claude --help` |

### Session Management

| Flag | Description | Example |
|------|-------------|---------|
| `--resume <session-id>` | Resume specific session by ID | `claude --resume abc123 "query"` |
| `--continue` | Load most recent conversation | `claude --continue` |
| `--max-turns <number>` | Limit agentic turns | `claude --max-turns 5` |
| `--no-interactive` | Run in non-interactive mode | `claude --resume abc123 "query" --no-interactive` |

### Model and Behavior Control

| Flag | Description | Example |
|------|-------------|---------|
| `--model <model>` | Set session model | `claude --model claude-sonnet-4-20250514` |
| `--append-system-prompt <text>` | Add to system prompt (print mode only) | `claude --append-system-prompt "Custom instruction"` |
| `--permission-mode <mode>` | Set permission mode | `claude --permission-mode plan` |
| `--dangerously-skip-permissions` | Skip all permission prompts | `claude --dangerously-skip-permissions` |

### Input/Output Control

| Flag | Description | Options | Example |
|------|-------------|---------|---------|
| `--output-format <format>` | Set output format | `text`, `json`, `stream-json` | `claude -p "query" --output-format json` |
| `--input-format <format>` | Set input format | `text`, `stream-json` | `claude -p --input-format stream-json` |
| `--verbose` | Enable detailed logging | - | `claude --verbose` |

### Tool Management

| Flag | Description | Example |
|------|-------------|---------|
| `--allowedTools <tools>` | Specify permitted tools | `claude --allowedTools "Bash(git log:*)" "Read"` |
| `--disallowedTools <tools>` | Block specific tools | `claude --disallowedTools "Bash(rm:*)" "Edit"` |
| `--permission-prompt-tool <tool>` | Designate MCP tool for permission prompts | `claude -p --permission-prompt-tool mcp_auth_tool "query"` |

### Directory and Context

| Flag | Description | Example |
|------|-------------|---------|
| `--add-dir <path>` | Add working directories | `claude --add-dir ../docs/` |

## Permission Modes

Claude Code supports several permission modes that control how it interacts with tools and your system:

| Mode | Description | Use Case | Example |
|------|-------------|----------|---------|
| `default` | Standard permission behavior | Normal development | `claude --permission-mode default` |
| `acceptEdits` | Auto-accept file edits, ask for others | Safe development workflow | `claude --permission-mode acceptEdits` |
| `plan` | Read-only mode for analysis and planning | Code exploration and planning | `claude --permission-mode plan` |
| `bypassPermissions` | Bypass permission system entirely | Advanced automation | `claude --permission-mode bypassPermissions` |
| `ask` | Ask before each tool use | Maximum control and security | `claude --permission-mode ask` |

## Output Formats

### Available Formats

| Format | Description | Use Case | Example |
|--------|-------------|----------|---------|
| `text` | Human-readable text (default) | Interactive use | `claude -p "query" --output-format text` |
| `json` | Structured JSON with metadata | Scripting and automation | `claude -p "query" --output-format json` |
| `stream-json` | Streaming JSON with incremental messages | Real-time processing | `claude -p "query" --output-format stream-json` |

### JSON Output Structure

When using `--output-format json`, Claude Code returns structured data:

```json
{
  "type": "result",
  "subtype": "success",
  "total_cost_usd": 0.003,
  "is_error": false,
  "duration_ms": 1234,
  "duration_api_ms": 800,
  "num_turns": 6,
  "result": "The response text here...",
  "session_id": "abc123"
}
```

## Basic Usage Examples

### Interactive Sessions

```bash
# Start interactive session
claude

# Start with initial prompt
claude "what files changed recently?"

# Continue previous conversation
claude --continue

# Resume specific session
claude --resume abc123 "continue working on authentication"
```

### Non-Interactive Queries

```bash
# Quick question
claude -p "explain this function"

# Process file content
cat error.log | claude -p "analyze this error"

# With specific output format
claude -p "list all functions" --output-format json
```

### Session Management

```bash
# Start legal document review with session persistence
session_id=$(claude -p "Start legal review session" --output-format json | jq -r '.session_id')

# Continue review in multiple steps
claude -p --resume "$session_id" "Review contract.pdf for liability clauses"
claude -p --resume "$session_id" "Check compliance with GDPR requirements"
claude -p --resume "$session_id" "Generate executive summary of risks"
```

## Advanced Usage Examples

### Permission Control

```bash
# Safe development mode
claude -p "Stage my changes and write commits" \
  --permission-mode acceptEdits \
  --allowedTools "Bash(git:*),Read,Write"

# Read-only analysis
claude --permission-mode plan \
  --disallowedTools "Write,Edit,Bash(rm:*)"

# Maximum automation
claude --permission-mode bypassPermissions \
  --dangerously-skip-permissions
```

### Tool-Specific Permissions

```bash
# Allow specific git operations
claude --allowedTools "Bash(git log:*)" "Bash(git diff:*)" "Read"

# Block dangerous operations
claude --disallowedTools "Bash(rm:*)" "Bash(sudo:*)" "Edit"

# Complex tool rules
claude --allowedTools "Bash(npm run test:*)" "Bash(git:*)" "Read(**/*.js)"
```

### Output Format Control

```bash
# Text output for human reading
cat data.txt | claude -p 'summarize this data' --output-format text > summary.txt

# JSON for scripting
cat code.py | claude -p 'analyze this code for bugs' --output-format json > analysis.json

# Streaming JSON for real-time processing
cat log.txt | claude -p 'parse this log file for errors' --output-format stream-json
```

### Model Selection

```bash
# Use specific model
claude --model claude-sonnet-4-20250514 "review this code"

# Model aliases
claude --model sonnet "quick analysis"
claude --model opus "complex reasoning task"
claude --model haiku "simple formatting"
```

## Configuration Commands

### Global Configuration

```bash
# List all global settings
claude config list -g

# Set global configuration
claude config set -g theme dark
claude config set -g autoUpdates false

# Get specific setting
claude config get autoUpdates
```

### Project Configuration

```bash
# List project settings
claude config list

# Set project-specific model
claude config set model opus

# Add permissions
claude config add permissions.allow "Bash(npm:*)"
claude config add permissions.allow "Read(**/*.ts)"

# Remove permissions
claude config remove permissions.deny "WebFetch"
```

### Configuration Examples

```bash
# Add allowed tool patterns
claude config add permissions.allow "Bash(git log:*)"
claude config add permissions.allow "Bash(npm run test:*)"

# Set default permission mode
claude config set permissions.defaultMode "acceptEdits"

# Configure model preferences
claude config set model "claude-3-5-sonnet-20241022"
```

## MCP Integration

### Server Management

```bash
# List all configured MCP servers
claude mcp list

# Get details for specific server
claude mcp get github

# Remove a server
claude mcp remove github

# Test server connection
claude mcp test linear
```

### MCP Command Usage

```bash
# Use MCP slash commands (format: /mcp__<server>__<prompt>)
/mcp__github__create-issue
/mcp__linear__update-ticket
/mcp__notion__create-page
```

## Session Management

### Session Information

Sessions are identified by unique IDs and stored in `~/.claude/projects/<project>/`:

```bash
# View session status
> /status
> /cost

# Clear current session (keep context)
> /clear

# Compact conversation history
> /compact
```

### Session ID Usage

```bash
# Resume with explicit session ID
claude --resume "550e8400-e29b-41d4-a716-446655440000" "continue debugging"

# Extract session ID from JSON output
session_id=$(claude -p "start task" --output-format json | jq -r '.session_id')
claude --resume "$session_id" "continue task"
```

## Error Handling and Exit Codes

### Exit Codes

| Code | Meaning | When It Occurs |
|------|---------|----------------|
| `0` | Success | Normal completion |
| `1` | General error | Invalid arguments, auth failures |
| `2` | Interrupted | User cancelled with Ctrl+C |
| `130` | SIGINT | Interrupted by system signal |

### Error Debugging

```bash
# Enable verbose logging for debugging
claude --verbose

# Check system status
> /status
> /doctor

# Verify configuration
claude config list
claude config get permissions
```

## Integration Examples

### CI/CD Pipeline

```bash
# Non-interactive code review with constraints
claude -p "review changes in this PR" \
  --output-format json \
  --max-turns 2 \
  --permission-mode plan \
  --allowedTools "Read,Grep,Bash(git:*)"
```

### Automated Documentation

```bash
# Generate docs for changed files
git diff --name-only | xargs -I {} \
  claude -p "document this file: {}" \
  --output-format text \
  --permission-mode plan
```

### Git Workflow Integration

```bash
# AI-assisted commit with specific permissions
claude -p "analyze staged changes and create commit message" \
  --allowedTools "Bash(git:*)" \
  --permission-mode acceptEdits \
  --output-format json
```

### Stream Processing

```bash
# Stream JSON messages for multi-turn conversations
echo '{"type":"user","message":{"role":"user","content":[{"type":"text","text":"Explain this code"}]}}' | \
  claude -p --output-format=stream-json --input-format=stream-json --verbose
```

## Shell Integration

### Useful Aliases

```bash
# Add to ~/.bashrc or ~/.zshrc
alias cc='claude'
alias ccq='claude -p'      # Quick query
alias ccc='claude -c'      # Continue session
alias ccp='claude --permission-mode plan'  # Plan mode
alias ccj='claude --output-format json'    # JSON output
```

### Environment Setup

```bash
# IDE integration
claude /ide

# Auto-detect VS Code context
claude  # Automatically detects running VS Code
```

## Best Practices

### Security Considerations

1. **Use appropriate permission modes** - Start with `plan` mode for exploration
2. **Limit tool access** - Use `--allowedTools` to restrict capabilities
3. **Avoid dangerous flags** - Use `--dangerously-skip-permissions` only when necessary
4. **Review permissions** - Regularly audit your permission settings

### Performance Optimization

1. **Use specific models** - Choose appropriate model for task complexity
2. **Limit turns** - Use `--max-turns` for simple tasks
3. **Cache sessions** - Resume sessions instead of starting fresh
4. **Use JSON output** - More efficient for programmatic processing

### Workflow Integration

1. **Session management** - Use descriptive initial prompts
2. **Output formats** - Choose format based on use case (text vs JSON)
3. **Tool constraints** - Define clear tool boundaries for safety
4. **Error handling** - Always check exit codes in scripts

## Troubleshooting

### Common Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| Permission denied | Insufficient tool permissions | Check `--allowedTools` or adjust permission mode |
| Command not found | Claude Code not in PATH | Verify installation and PATH configuration |
| Session not found | Invalid session ID | Use `claude config list` to verify session IDs |
| API errors | Network or authentication issues | Use `--verbose` for detailed error information |
| Tool blocked | Disallowed by configuration | Review `--disallowedTools` and permission settings |

### Debug Commands

```bash
# Comprehensive debugging
claude --verbose --output-format json

# System diagnostics
> /status
> /doctor
> /memory

# Configuration verification
claude config list
claude config get permissions
claude mcp list
```

### Testing Configuration

```bash
# Test permission setup
claude -p "echo test" --allowedTools "Bash(echo:*)" --verbose

# Test model selection
claude --model sonnet -p "simple test" --output-format json

# Test MCP integration
claude mcp test <server-name>
```

This CLI reference provides comprehensive coverage of Claude Code's command-line capabilities, enabling efficient automation and integration into development workflows.
