# Status Line Configuration

Customize Claude Code with a dynamic status line that displays contextual information at the bottom of the interface, similar to terminal prompts in shells like Oh-my-zsh.

## Overview

The status line system provides real-time contextual information about your Claude Code session, including:

- Current model and version
- Working directory and project information
- Git branch status
- Cost and performance metrics
- Session duration and API usage
- Custom styling with ANSI colors

## Configuration Methods

### Interactive Setup

Run the `/statusline` command for guided setup:

```bash
# Basic setup - reproduces your terminal prompt
/statusline

# Custom instructions
/statusline show the model name in orange
/statusline include git branch and cost information
```

### Manual Configuration

Add a `statusLine` configuration to your `.claude/settings.json`:

```json
{
  "statusLine": {
    "type": "command",
    "command": "~/.claude/statusline.sh",
    "padding": 0
  }
}
```

## How Status Lines Work

### Update Behavior

- **Trigger**: Updates when conversation messages change
- **Rate Limiting**: Updates run at most every 300ms to prevent performance issues
- **Output**: First line of stdout becomes the status line text
- **Styling**: Supports ANSI color codes and emojis

### Script Execution

1. Claude Code executes your specified command
2. Session data is passed as JSON via stdin
3. Your script processes the data and outputs formatted text
4. The output appears as the status line

## JSON Input Structure

Your status line script receives comprehensive session metadata:

```json
{
  "hook_event_name": "Status",
  "session_id": "abc123...",
  "transcript_path": "/path/to/transcript.json",
  "cwd": "/current/working/directory",
  "model": {
    "id": "claude-opus-4-1",
    "display_name": "Opus"
  },
  "workspace": {
    "current_dir": "/current/working/directory",
    "project_dir": "/original/project/directory"
  },
  "version": "1.0.80",
  "output_style": {
    "name": "default"
  },
  "cost": {
    "total_cost_usd": 0.01234,
    "total_duration_ms": 45000,
    "total_api_duration_ms": 2300,
    "total_lines_added": 156,
    "total_lines_removed": 23
  }
}
```

### Available Data Fields

| Field | Description | Example Value |
|-------|-------------|---------------|
| `hook_event_name` | Always "Status" for status line | `"Status"` |
| `session_id` | Unique session identifier | `"abc123..."` |
| `transcript_path` | Path to conversation transcript | `"/path/to/transcript.json"` |
| `cwd` | Current working directory | `"/Users/user/project"` |
| `model.id` | Full model identifier | `"claude-opus-4-1"` |
| `model.display_name` | Human-readable model name | `"Opus"` |
| `workspace.current_dir` | Current directory path | `"/Users/user/project"` |
| `workspace.project_dir` | Original project directory | `"/Users/user/project"` |
| `version` | Claude Code version | `"1.0.80"` |
| `output_style.name` | Current output style | `"default"` |
| `cost.total_cost_usd` | Total session cost in USD | `0.01234` |
| `cost.total_duration_ms` | Total session duration | `45000` |
| `cost.total_api_duration_ms` | API call duration | `2300` |
| `cost.total_lines_added` | Lines of code added | `156` |
| `cost.total_lines_removed` | Lines of code removed | `23` |

## Example Scripts

### Basic Bash Script

```bash
#!/bin/bash
# Read JSON input from stdin
input=$(cat)

# Extract values using jq
MODEL_DISPLAY=$(echo "$input" | jq -r '.model.display_name')
CURRENT_DIR=$(echo "$input" | jq -r '.workspace.current_dir')

echo "[$MODEL_DISPLAY] 📁 ${CURRENT_DIR##*/}"
```

### Advanced Bash with Helper Functions

```bash
#!/bin/bash
# Read JSON input once
input=$(cat)

# Helper functions for common extractions
get_model_name() { echo "$input" | jq -r '.model.display_name'; }
get_current_dir() { echo "$input" | jq -r '.workspace.current_dir'; }
get_project_dir() { echo "$input" | jq -r '.workspace.project_dir'; }
get_version() { echo "$input" | jq -r '.version'; }
get_cost() { echo "$input" | jq -r '.cost.total_cost_usd'; }
get_duration() { echo "$input" | jq -r '.cost.total_duration_ms'; }
get_lines_added() { echo "$input" | jq -r '.cost.total_lines_added'; }
get_lines_removed() { echo "$input" | jq -r '.cost.total_lines_removed'; }

# Use the helpers
MODEL=$(get_model_name)
DIR=$(get_current_dir)
echo "[$MODEL] 📁 ${DIR##*/}"
```

### Git-Aware Bash Script

```bash
#!/bin/bash
# Read JSON input from stdin
input=$(cat)

# Extract values using jq
MODEL_DISPLAY=$(echo "$input" | jq -r '.model.display_name')
CURRENT_DIR=$(echo "$input" | jq -r '.workspace.current_dir')

# Show git branch if in a git repo
GIT_BRANCH=""
if git rev-parse --git-dir > /dev/null 2>&1; then
    BRANCH=$(git branch --show-current 2>/dev/null)
    if [ -n "$BRANCH" ]; then
        GIT_BRANCH=" | 🌿 $BRANCH"
    fi
fi

echo "[$MODEL_DISPLAY] 📁 ${CURRENT_DIR##*/}$GIT_BRANCH"
```

### Python Script

```python
#!/usr/bin/env python3
import json
import sys
import os

# Read JSON from stdin
data = json.load(sys.stdin)

# Extract values
model = data['model']['display_name']
current_dir = os.path.basename(data['workspace']['current_dir'])

# Check for git branch
git_branch = ""
if os.path.exists('.git'):
    try:
        with open('.git/HEAD', 'r') as f:
            ref = f.read().strip()
            if ref.startswith('ref: refs/heads/'):
                git_branch = f" | 🌿 {ref.replace('ref: refs/heads/', '')}"
    except:
        pass

print(f"[{model}] 📁 {current_dir}{git_branch}")
```

### Node.js Script

```javascript
#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

// Read JSON from stdin
let input = '';
process.stdin.on('data', chunk => input += chunk);
process.stdin.on('end', () => {
    const data = JSON.parse(input);

    // Extract values
    const model = data.model.display_name;
    const currentDir = path.basename(data.workspace.current_dir);

    // Check for git branch
    let gitBranch = '';
    try {
        const headContent = fs.readFileSync('.git/HEAD', 'utf8').trim();
        if (headContent.startsWith('ref: refs/heads/')) {
            gitBranch = ` | 🌿 ${headContent.replace('ref: refs/heads/', '')}`;
        }
    } catch (e) {
        // Not a git repo or can't read HEAD
    }

    console.log(`[${model}] 📁 ${currentDir}${gitBranch}`);
});
```

## Advanced Status Line Examples

### Cost and Performance Tracking

```bash
#!/bin/bash
input=$(cat)

MODEL=$(echo "$input" | jq -r '.model.display_name')
COST=$(echo "$input" | jq -r '.cost.total_cost_usd')
DURATION=$(echo "$input" | jq -r '.cost.total_duration_ms')
LINES_ADDED=$(echo "$input" | jq -r '.cost.total_lines_added')

# Convert duration to seconds
DURATION_SEC=$((DURATION / 1000))

echo "[$MODEL] 💰 \$${COST} | ⏱️ ${DURATION_SEC}s | 📝 +${LINES_ADDED} lines"
```

### Colorized Status Line

```bash
#!/bin/bash
input=$(cat)

MODEL=$(echo "$input" | jq -r '.model.display_name')
DIR=$(echo "$input" | jq -r '.workspace.current_dir')
COST=$(echo "$input" | jq -r '.cost.total_cost_usd')

# ANSI color codes
BLUE='\033[34m'
GREEN='\033[32m'
YELLOW='\033[33m'
RESET='\033[0m'

echo -e "${BLUE}[${MODEL}]${RESET} ${GREEN}📁 ${DIR##*/}${RESET} ${YELLOW}💰 \$${COST}${RESET}"
```

### Multi-Information Status Line

```bash
#!/bin/bash
input=$(cat)

# Extract all relevant information
MODEL=$(echo "$input" | jq -r '.model.display_name')
DIR=$(echo "$input" | jq -r '.workspace.current_dir')
VERSION=$(echo "$input" | jq -r '.version')
STYLE=$(echo "$input" | jq -r '.output_style.name')
COST=$(echo "$input" | jq -r '.cost.total_cost_usd')

# Get git branch
GIT_BRANCH=""
if git rev-parse --git-dir > /dev/null 2>&1; then
    BRANCH=$(git branch --show-current 2>/dev/null)
    if [ -n "$BRANCH" ]; then
        GIT_BRANCH=" | 🌿 $BRANCH"
    fi
fi

echo "[$MODEL v$VERSION] 📁 ${DIR##*/}$GIT_BRANCH | 🎨 $STYLE | 💰 \$$COST"
```

## Configuration Options

### Padding Control

```json
{
  "statusLine": {
    "type": "command",
    "command": "~/.claude/statusline.sh",
    "padding": 0  // Set to 0 to let status line go to screen edge
  }
}
```

### Script Permissions

Ensure your status line script is executable:

```bash
chmod +x ~/.claude/statusline.sh
```

## Best Practices

### Performance Optimization

- **Keep scripts fast**: Status lines update frequently, so optimize for speed
- **Cache expensive operations**: Store git branch info in variables
- **Limit external commands**: Minimize subprocess calls
- **Handle errors gracefully**: Use try-catch blocks for file operations

### Content Guidelines

- **One line output**: Only the first line of stdout is used
- **Concise information**: Display most relevant data for your workflow
- **Visual hierarchy**: Use colors and icons to organize information
- **Responsive design**: Consider terminal width limitations

### Error Handling

```bash
#!/bin/bash
input=$(cat)

# Safe JSON parsing with fallbacks
MODEL=$(echo "$input" | jq -r '.model.display_name // "Unknown"')
DIR=$(echo "$input" | jq -r '.workspace.current_dir // "~"')

# Safe git operations
GIT_BRANCH=""
if command -v git >/dev/null 2>&1 && git rev-parse --git-dir >/dev/null 2>&1; then
    BRANCH=$(git branch --show-current 2>/dev/null || echo "")
    if [ -n "$BRANCH" ]; then
        GIT_BRANCH=" | 🌿 $BRANCH"
    fi
fi

echo "[$MODEL] 📁 ${DIR##*/}$GIT_BRANCH"
```

## Troubleshooting

### Common Issues

1. **Script not executing**: Check file permissions with `chmod +x`
2. **JSON parsing errors**: Ensure `jq` is installed and test with mock data
3. **No output**: Verify script outputs to stdout, not stderr
4. **Performance issues**: Profile script execution time and optimize

### Testing Your Status Line

Test your script with mock JSON data:

```bash
echo '{
  "model": {"display_name": "Opus"},
  "workspace": {"current_dir": "/test/project"},
  "cost": {"total_cost_usd": 0.01}
}' | ~/.claude/statusline.sh
```

### Debug Mode

Add debug output to understand script behavior:

```bash
#!/bin/bash
input=$(cat)

# Debug: log input to file
echo "$input" >> /tmp/statusline-debug.log

# Your status line logic here...
```

## Integration with Other Features

### Hooks Integration

Status lines work alongside Claude Code hooks but serve different purposes:

- **Status lines**: Display information
- **Hooks**: Execute actions and validation

### MCP Integration

Access MCP server information in your status line:

```bash
# Check if specific MCP servers are running
if claude mcp list | grep -q "github"; then
    echo "[$MODEL] 📁 ${DIR##*/} | 🔗 GitHub MCP"
else
    echo "[$MODEL] 📁 ${DIR##*/}"
fi
```

### Output Style Coordination

Coordinate status line styling with your output style:

```bash
STYLE=$(echo "$input" | jq -r '.output_style.name')

case "$STYLE" in
    "explanatory")
        echo "[$MODEL] 🎓 ${DIR##*/} | Learning Mode"
        ;;
    "default")
        echo "[$MODEL] 📁 ${DIR##*/}"
        ;;
    *)
        echo "[$MODEL] 🎨 ${DIR##*/} | $STYLE"
        ;;
esac
```

The status line system provides a powerful way to maintain awareness of your Claude Code session context, helping you stay oriented and informed throughout your development workflow.