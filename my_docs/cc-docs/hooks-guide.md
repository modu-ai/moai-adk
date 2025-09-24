# Hooks Guide

Comprehensive guide to Claude Code hooks - user-defined shell commands that execute at different points in Claude Code's lifecycle, providing deterministic control over agent behavior.

## Overview

Claude Code hooks enable powerful workflow automation by executing custom commands at specific events during Claude's operation. They provide deterministic control over the agent's behavior, ensuring specific actions always occur when certain conditions are met.

### Key Benefits

- **Workflow Automation**: Automatically execute tasks like code formatting, linting, or notifications
- **Quality Control**: Enforce coding standards and prevent unwanted modifications
- **Integration**: Connect Claude Code with external tools and services
- **Monitoring**: Log activities and track tool usage patterns
- **Security**: Block dangerous operations and validate inputs

## Hook Events

Claude Code supports multiple hook events that trigger at different points in the interaction lifecycle:

### Core Events

| Event | Description | When It Triggers | Common Use Cases |
|-------|-------------|------------------|------------------|
| `PreToolUse` | Before any tool execution | Just before a tool is called | Validation, blocking, preprocessing |
| `PostToolUse` | After tool execution completes | After tool finishes successfully | Formatting, cleanup, notifications |
| `UserPromptSubmit` | When user submits a prompt | After user input is received | Logging, preprocessing, validation |
| `Notification` | When Claude sends notifications | During permission requests or alerts | Custom notification systems |
| `Stop` | When Claude finishes responding | After Claude completes its response | Session cleanup, summaries |
| `SubagentStop` | When subagent tasks complete | After subagent finishes | Subagent result processing |
| `PreCompact` | Before conversation compacting | Before context reduction | Backup, logging |
| `SessionStart` | When starting/resuming session | At session initialization | Environment setup |
| `SessionEnd` | When session ends | At session termination | Cleanup, reporting |

### Event-Specific Data

Each hook event receives contextual information relevant to its trigger point:

```json
{
  "hook_event_name": "PreToolUse",
  "session_id": "abc123...",
  "tool_name": "Edit",
  "tool_input": {
    "file_path": "/path/to/file.ts",
    "old_string": "original code",
    "new_string": "modified code"
  },
  "workspace": {
    "current_dir": "/project/path",
    "project_dir": "/project/path"
  }
}
```

## Hook Configuration

### Interactive Configuration

Use the `/hooks` command for guided setup:

```bash
# Open hooks configuration interface
/hooks

# Configure specific event type
/hooks PreToolUse

# Quick access to hook debugging
/hooks debug
```

### Manual Configuration

Edit `.claude/settings.json` directly:

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Edit|Write|MultiEdit",
        "hooks": [
          {
            "type": "command",
            "command": "~/.claude/hooks/validate-edit.sh"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "echo 'Command executed' >> ~/.claude/bash.log"
          }
        ]
      }
    ]
  }
}
```

### Hook Structure

Each hook configuration includes:

- **matcher**: Tool name pattern (regex supported)
- **type**: Always "command" for shell commands
- **command**: Shell command to execute

## Matchers and Patterns

### Tool Name Matching

```json
{
  "matcher": "Edit",           // Exact match
  "matcher": "Edit|Write",     // Multiple tools
  "matcher": ".*",             // All tools
  "matcher": "Bash\\(git.*\\)", // Bash commands starting with git
  "matcher": "Read\\(.*\\.ts\\)" // TypeScript file reads
}
```

### Pattern Examples

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Edit|MultiEdit|Write",
        "hooks": [{"type": "command", "command": "echo 'File modification'"}]
      },
      {
        "matcher": "Bash\\(rm.*\\)",
        "hooks": [{"type": "command", "command": "echo 'Blocked dangerous rm command' && exit 1"}]
      },
      {
        "matcher": ".*\\.py$",
        "hooks": [{"type": "command", "command": "python -m py_compile"}]
      }
    ]
  }
}
```

## Practical Hook Examples

### Code Quality Automation

#### TypeScript Formatting Hook

```bash
#!/bin/bash
# ~/.claude/hooks/format-typescript.sh

# Read hook data from stdin
input=$(cat)

# Extract file path
file_path=$(echo "$input" | jq -r '.tool_input.file_path // empty')

if [[ -n "$file_path" && "$file_path" =~ \.tsx?$ ]]; then
    echo "Formatting TypeScript file: $file_path"
    npx prettier --write "$file_path"
    npx eslint --fix "$file_path" 2>/dev/null || true
fi
```

Configuration:
```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Edit|Write|MultiEdit",
        "hooks": [
          {
            "type": "command",
            "command": "~/.claude/hooks/format-typescript.sh"
          }
        ]
      }
    ]
  }
}
```

#### Python Code Validation

```bash
#!/bin/bash
# ~/.claude/hooks/validate-python.sh

input=$(cat)
file_path=$(echo "$input" | jq -r '.tool_input.file_path // empty')

if [[ -n "$file_path" && "$file_path" =~ \.py$ ]]; then
    echo "Validating Python file: $file_path"

    # Syntax check
    if ! python -m py_compile "$file_path" 2>/dev/null; then
        echo "Python syntax error detected!"
        exit 1
    fi

    # Basic linting
    if command -v black >/dev/null; then
        black "$file_path"
    fi

    if command -v flake8 >/dev/null; then
        flake8 "$file_path" --max-line-length=88
    fi
fi
```

### Security and Safety Hooks

#### Dangerous Command Blocker

```bash
#!/bin/bash
# ~/.claude/hooks/security-blocker.sh

input=$(cat)
command=$(echo "$input" | jq -r '.tool_input.command // empty')

# Block dangerous commands
dangerous_patterns=(
    "rm -rf /"
    "rm -rf \$HOME"
    "chmod 777"
    "sudo rm"
    "> /dev/sda"
)

for pattern in "${dangerous_patterns[@]}"; do
    if [[ "$command" =~ $pattern ]]; then
        echo "BLOCKED: Dangerous command detected: $pattern"
        exit 1
    fi
done

echo "Command safety check passed"
```

#### Sensitive File Protection

```bash
#!/bin/bash
# ~/.claude/hooks/protect-sensitive.sh

input=$(cat)
file_path=$(echo "$input" | jq -r '.tool_input.file_path // empty')

# Protected files and directories
protected_patterns=(
    "/etc/passwd"
    "/etc/shadow"
    "~/.ssh/"
    ".env"
    "config/secrets"
)

for pattern in "${protected_patterns[@]}"; do
    if [[ "$file_path" =~ $pattern ]]; then
        echo "BLOCKED: Attempt to modify protected file: $file_path"
        exit 1
    fi
done
```

### Development Workflow Automation

#### Git Integration Hook

```bash
#!/bin/bash
# ~/.claude/hooks/git-integration.sh

input=$(cat)
file_path=$(echo "$input" | jq -r '.tool_input.file_path // empty')

if [[ -n "$file_path" && -f "$file_path" ]]; then
    # Check if file is in git repo
    if git rev-parse --git-dir > /dev/null 2>&1; then
        # Add file to git staging
        git add "$file_path"
        echo "Added $file_path to git staging area"

        # Log modification
        echo "$(date): Modified $file_path" >> ~/.claude/file-changes.log
    fi
fi
```

#### Project Documentation Updater

```bash
#!/bin/bash
# ~/.claude/hooks/update-docs.sh

input=$(cat)
file_path=$(echo "$input" | jq -r '.tool_input.file_path // empty')

# Update documentation when certain files change
if [[ "$file_path" =~ (src/.*\.(ts|js|py)$|README\.md$) ]]; then
    echo "Updating project documentation..."

    # Generate API documentation
    if command -v typedoc >/dev/null && [[ "$file_path" =~ \.ts$ ]]; then
        npx typedoc --out docs/ src/
    fi

    # Update README if package.json changes
    if [[ "$file_path" == "package.json" ]]; then
        # Custom README update logic
        echo "Package.json modified - consider updating README"
    fi
fi
```

### Notification and Monitoring

#### Slack Notification Hook

```bash
#!/bin/bash
# ~/.claude/hooks/slack-notify.sh

input=$(cat)
tool_name=$(echo "$input" | jq -r '.tool_name // empty')
file_path=$(echo "$input" | jq -r '.tool_input.file_path // empty')

# Notify on important changes
if [[ "$tool_name" == "Edit" && "$file_path" =~ (production|config) ]]; then
    webhook_url="YOUR_SLACK_WEBHOOK_URL"

    message="{\"text\":\"🤖 Claude Code modified critical file: \`$file_path\`\"}"

    curl -X POST -H 'Content-type: application/json' \
         --data "$message" \
         "$webhook_url"
fi
```

#### Command Usage Analytics

```bash
#!/bin/bash
# ~/.claude/hooks/analytics.sh

input=$(cat)
tool_name=$(echo "$input" | jq -r '.tool_name // empty')
session_id=$(echo "$input" | jq -r '.session_id // empty')

# Log tool usage for analytics
log_entry="{\"timestamp\":\"$(date -Iseconds)\",\"tool\":\"$tool_name\",\"session\":\"$session_id\"}"
echo "$log_entry" >> ~/.claude/usage-analytics.jsonl

# Weekly summary generation
if [[ $(date +%u) == "1" ]]; then  # Monday
    echo "Generating weekly usage report..."
    # Process analytics data
fi
```

### Advanced Integration Patterns

#### MCP Tool Integration

```bash
#!/bin/bash
# ~/.claude/hooks/mcp-integration.sh

input=$(cat)
hook_event=$(echo "$input" | jq -r '.hook_event_name')

case "$hook_event" in
    "SessionStart")
        echo "Session started - checking MCP servers..."
        # Verify MCP server connections
        claude mcp list | grep -q "running" || echo "Warning: MCP servers not running"
        ;;
    "PreToolUse")
        tool_name=$(echo "$input" | jq -r '.tool_name')
        if [[ "$tool_name" =~ ^mcp__ ]]; then
            echo "MCP tool usage detected: $tool_name"
            # Log MCP tool usage
        fi
        ;;
esac
```

#### Multi-Hook Coordination

```bash
#!/bin/bash
# ~/.claude/hooks/coordinator.sh

input=$(cat)
hook_event=$(echo "$input" | jq -r '.hook_event_name')

# Create coordination file
coord_file="/tmp/claude-hooks-coordination"
touch "$coord_file"

case "$hook_event" in
    "PreToolUse")
        echo "pre-tool-$(date +%s)" >> "$coord_file"
        ;;
    "PostToolUse")
        echo "post-tool-$(date +%s)" >> "$coord_file"

        # Check if we have matching pre/post events
        pre_count=$(grep -c "pre-tool" "$coord_file")
        post_count=$(grep -c "post-tool" "$coord_file")

        if [[ $pre_count -eq $post_count ]]; then
            echo "All tools completed successfully"
            # Trigger batch operations
        fi
        ;;
esac
```

## Hook Input/Output

### Input Structure

Hooks receive comprehensive JSON data via stdin:

```json
{
  "hook_event_name": "PreToolUse",
  "session_id": "550e8400-e29b-41d4-a716-446655440000",
  "transcript_path": "/path/to/transcript.json",
  "cwd": "/current/working/directory",
  "tool_name": "Edit",
  "tool_input": {
    "file_path": "/path/to/file.ts",
    "old_string": "const value = 1;",
    "new_string": "const value = 2;"
  },
  "workspace": {
    "current_dir": "/current/working/directory",
    "project_dir": "/original/project/directory"
  },
  "model": {
    "id": "claude-opus-4-1",
    "display_name": "Opus"
  },
  "version": "1.0.80",
  "output_style": {
    "name": "default"
  }
}
```

### Output Options

Hooks can produce different types of output:

#### Success (Exit Code 0)

```bash
echo "Hook executed successfully"
exit 0
```

#### Block Operation (Exit Code 1)

```bash
echo "Operation blocked: Invalid file modification"
exit 1
```

#### Structured JSON Response

```bash
response='{
  "status": "success",
  "message": "File processed successfully",
  "metadata": {
    "lines_changed": 15,
    "warnings": []
  }
}'
echo "$response"
```

#### Multi-line Output

```bash
cat << 'EOF'
File modification completed:
- Applied formatting
- Fixed linting issues
- Updated documentation
EOF
```

## Security Best Practices

### Input Validation

Always validate and sanitize inputs:

```bash
#!/bin/bash
# Secure hook template

input=$(cat)

# Validate JSON input
if ! echo "$input" | jq empty 2>/dev/null; then
    echo "Invalid JSON input"
    exit 1
fi

# Safely extract values with defaults
file_path=$(echo "$input" | jq -r '.tool_input.file_path // empty')
tool_name=$(echo "$input" | jq -r '.tool_name // empty')

# Validate file paths
if [[ -n "$file_path" ]]; then
    # Check for path traversal attacks
    if [[ "$file_path" =~ \.\./|\.\.\\ ]]; then
        echo "Blocked: Path traversal attempt"
        exit 1
    fi

    # Ensure absolute path
    if [[ ! "$file_path" =~ ^/ ]]; then
        echo "Blocked: Relative path not allowed"
        exit 1
    fi
fi
```

### Command Injection Prevention

```bash
#!/bin/bash
# Safe command execution

# DON'T: Direct substitution (vulnerable)
# eval "some_command $user_input"

# DO: Proper quoting and validation
if [[ "$user_input" =~ ^[a-zA-Z0-9._/-]+$ ]]; then
    some_command "$user_input"
else
    echo "Invalid input format"
    exit 1
fi
```

### File System Security

```bash
#!/bin/bash
# File system protection

file_path="$1"

# Check file exists and is readable
if [[ ! -f "$file_path" ]]; then
    echo "File does not exist: $file_path"
    exit 1
fi

# Check permissions
if [[ ! -r "$file_path" ]]; then
    echo "File not readable: $file_path"
    exit 1
fi

# Restrict to project directory
project_root="/project/root"
if [[ ! "$file_path" =~ ^"$project_root" ]]; then
    echo "File outside project scope: $file_path"
    exit 1
fi
```

### Environment Isolation

```bash
#!/bin/bash
# Secure environment setup

# Clear potentially dangerous environment variables
unset IFS
unset CDPATH
unset ENV
unset BASH_ENV

# Set secure PATH
export PATH="/usr/local/bin:/usr/bin:/bin"

# Set secure umask
umask 022

# Your hook logic here
```

## Debugging and Troubleshooting

### Hook Development Workflow

1. **Start Simple**: Begin with basic echo commands
2. **Test Isolation**: Run hooks manually with sample JSON
3. **Add Logging**: Include detailed logging for debugging
4. **Gradual Complexity**: Add features incrementally

### Debug Logging

```bash
#!/bin/bash
# Debug-enabled hook

# Enable debugging
DEBUG=${CLAUDE_HOOK_DEBUG:-false}
LOG_FILE="$HOME/.claude/hook-debug.log"

debug_log() {
    if [[ "$DEBUG" == "true" ]]; then
        echo "[$(date)] $*" >> "$LOG_FILE"
    fi
}

# Your hook code with debug calls
debug_log "Hook started with input: $input"
debug_log "Extracted file_path: $file_path"
```

### Testing Hooks

#### Manual Testing

```bash
# Create test input
cat > test-input.json << 'EOF'
{
  "hook_event_name": "PreToolUse",
  "tool_name": "Edit",
  "tool_input": {
    "file_path": "/test/file.ts",
    "old_string": "old",
    "new_string": "new"
  }
}
EOF

# Test your hook
cat test-input.json | ~/.claude/hooks/your-hook.sh
```

#### Automated Testing

```bash
#!/bin/bash
# test-hooks.sh

test_cases=(
    "test-cases/edit-typescript.json"
    "test-cases/bash-command.json"
    "test-cases/security-block.json"
)

for test_case in "${test_cases[@]}"; do
    echo "Testing: $test_case"

    if cat "$test_case" | ~/.claude/hooks/your-hook.sh; then
        echo "✓ PASS: $test_case"
    else
        echo "✗ FAIL: $test_case"
    fi
done
```

### Common Issues and Solutions

| Issue | Symptoms | Solution |
|-------|----------|----------|
| Hook not triggering | No output or logs | Check matcher pattern and event type |
| Permission denied | "Permission denied" errors | Verify hook script permissions (`chmod +x`) |
| JSON parsing errors | jq command failures | Validate input JSON structure |
| Path issues | File not found errors | Use absolute paths in hook commands |
| Performance problems | Slow Claude responses | Optimize hook execution time |

### Performance Optimization

#### Async Operations

```bash
#!/bin/bash
# Background processing for slow operations

input=$(cat)

# Quick validation
if validate_input "$input"; then
    echo "Input validated"

    # Start slow processing in background
    {
        perform_slow_operation "$input"
    } &

    # Don't wait for background job
    exit 0
else
    echo "Validation failed"
    exit 1
fi
```

#### Caching

```bash
#!/bin/bash
# Cache expensive operations

cache_file="$HOME/.claude/hook-cache"
cache_ttl=3600  # 1 hour

input=$(cat)
input_hash=$(echo "$input" | sha256sum | cut -d' ' -f1)
cache_entry="$cache_file.$input_hash"

# Check cache
if [[ -f "$cache_entry" ]] && [[ $(($(date +%s) - $(stat -c %Y "$cache_entry"))) -lt $cache_ttl ]]; then
    cat "$cache_entry"
    exit 0
fi

# Compute result and cache
result=$(compute_expensive_result "$input")
echo "$result" | tee "$cache_entry"
```

## Advanced Hook Patterns

### State Management

```bash
#!/bin/bash
# Stateful hook with persistence

state_file="$HOME/.claude/hook-state.json"

# Load existing state
if [[ -f "$state_file" ]]; then
    state=$(cat "$state_file")
else
    state='{"counter": 0, "last_file": ""}'
fi

# Update state
counter=$(echo "$state" | jq '.counter + 1')
file_path=$(echo "$input" | jq -r '.tool_input.file_path // empty')

new_state=$(echo "$state" | jq --argjson counter "$counter" --arg file "$file_path" \
    '.counter = $counter | .last_file = $file')

# Save state
echo "$new_state" > "$state_file"

echo "Hook execution #$counter for file: $file_path"
```

### Hook Composition

```bash
#!/bin/bash
# Composable hook system

input=$(cat)

# Run multiple sub-hooks
hooks=(
    "~/.claude/hooks/validate.sh"
    "~/.claude/hooks/format.sh"
    "~/.claude/hooks/notify.sh"
)

for hook in "${hooks[@]}"; do
    if [[ -x "$hook" ]]; then
        echo "Running: $hook"
        if ! echo "$input" | "$hook"; then
            echo "Hook failed: $hook"
            exit 1
        fi
    fi
done

echo "All hooks completed successfully"
```

### Conditional Logic

```bash
#!/bin/bash
# Context-aware hook behavior

input=$(cat)
tool_name=$(echo "$input" | jq -r '.tool_name')
file_path=$(echo "$input" | jq -r '.tool_input.file_path // empty')
session_id=$(echo "$input" | jq -r '.session_id')

# Different behavior based on context
case "$tool_name" in
    "Edit"|"Write"|"MultiEdit")
        if [[ "$file_path" =~ \.tsx?$ ]]; then
            echo "Processing TypeScript file"
            format_typescript "$file_path"
        elif [[ "$file_path" =~ \.py$ ]]; then
            echo "Processing Python file"
            format_python "$file_path"
        fi
        ;;
    "Bash")
        command=$(echo "$input" | jq -r '.tool_input.command')
        if [[ "$command" =~ ^git ]]; then
            echo "Git command detected"
            log_git_operation "$command" "$session_id"
        fi
        ;;
esac
```

## Integration with External Tools

### CI/CD Integration

```bash
#!/bin/bash
# CI/CD webhook trigger

input=$(cat)
file_path=$(echo "$input" | jq -r '.tool_input.file_path // empty')

# Trigger CI/CD on specific file changes
if [[ "$file_path" =~ (Dockerfile|\.github/workflows/|package\.json)$ ]]; then
    echo "Triggering CI/CD pipeline for: $file_path"

    # GitHub Actions workflow dispatch
    curl -X POST \
        -H "Authorization: token $GITHUB_TOKEN" \
        -H "Accept: application/vnd.github.v3+json" \
        "https://api.github.com/repos/$GITHUB_REPO/actions/workflows/build.yml/dispatches" \
        -d '{"ref":"main"}'
fi
```

### Database Logging

```bash
#!/bin/bash
# Database activity logging

input=$(cat)

# Extract relevant information
session_id=$(echo "$input" | jq -r '.session_id')
tool_name=$(echo "$input" | jq -r '.tool_name')
timestamp=$(date -Iseconds)

# Log to SQLite database
sqlite3 ~/.claude/activity.db << EOF
CREATE TABLE IF NOT EXISTS hook_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp TEXT,
    session_id TEXT,
    tool_name TEXT,
    event_data TEXT
);

INSERT INTO hook_events (timestamp, session_id, tool_name, event_data)
VALUES ('$timestamp', '$session_id', '$tool_name', '$(echo "$input" | sed "s/'/''/'")');
EOF

echo "Logged hook event to database"
```

This comprehensive hooks guide provides everything needed to effectively use Claude Code hooks for workflow automation, security, and integration. The examples range from simple logging to complex multi-step workflows, all while maintaining security best practices.