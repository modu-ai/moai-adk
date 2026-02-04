# Claude Code Statusline - Official Documentation Reference

Source: https://code.claude.com/docs/en/statusline
Updated: 2026-01-06

## Overview

The statusline provides a customizable display area in Claude Code's interface for showing dynamic information such as project status, resource usage, and custom metrics.

## Setup

### Interactive Setup

Use the /statusline command to configure via interactive interface.

### Manual Configuration

Add statusline configuration to .claude/settings.json:

```json
{
  "statusLine": {
    "command": "path/to/statusline-script.sh"
  }
}
```

## How Statusline Works

### Execution Model

1. Claude Code invokes the statusline command
2. Script receives JSON input via stdin
3. First line of stdout becomes the status display
4. Updates occur at most every 300ms

### Input Format

The statusline script receives JSON via stdin:

```json
{
  "hook_event_name": "statusLine",
  "session_id": "abc123",
  "cwd": "/path/to/project",
  "model": "claude-sonnet-4",
  "workspace": "/path/to/workspace",
  "cost": {
    "total_usd": 0.05,
    "input_tokens": 1000,
    "output_tokens": 500
  },
  "context_window": {
    "used": 50000,
    "total": 200000
  }
}
```

### Available Input Fields

- hook_event_name: Always "statusLine"
- session_id: Current session identifier
- cwd: Current working directory
- model: Active Claude model name
- workspace: Workspace root path
- cost: Usage cost information
  - total_usd: Total cost in USD
  - input_tokens: Input token count
  - output_tokens: Output token count
- context_window: Context usage
  - used: Tokens currently used
  - total: Maximum available tokens

## ANSI Color Support

Statusline supports ANSI escape codes for styling:

### Color Examples

- Red: \033[31m
- Green: \033[32m
- Yellow: \033[33m
- Blue: \033[34m
- Reset: \033[0m

## Implementation Examples

### Bash Script

```bash
#!/bin/bash
read -r input
model=$(echo "$input" | jq -r '.model')
cost=$(echo "$input" | jq -r '.cost.total_usd')
used=$(echo "$input" | jq -r '.context_window.used')
total=$(echo "$input" | jq -r '.context_window.total')
pct=$((used * 100 / total))
echo "Model: $model | Cost: \$$cost | Context: ${pct}%"
```

### Python Script

```python
#!/usr/bin/env python3
import sys
import json

data = json.loads(sys.stdin.read())
model = data.get('model', 'unknown')
cost = data.get('cost', {}).get('total_usd', 0)
ctx = data.get('context_window', {})
used = ctx.get('used', 0)
total = ctx.get('total', 1)
pct = int(used * 100 / total)

print(f"Model: {model} | ${cost:.2f} | Ctx: {pct}%")
```

### Node.js Script

```javascript
#!/usr/bin/env node
let data = '';
process.stdin.on('data', chunk => data += chunk);
process.stdin.on('end', () => {
  const input = JSON.parse(data);
  const model = input.model || 'unknown';
  const cost = input.cost?.total_usd || 0;
  const used = input.context_window?.used || 0;
  const total = input.context_window?.total || 1;
  const pct = Math.round(used * 100 / total);
  console.log(`${model} | $${cost.toFixed(2)} | ${pct}%`);
});
```

## Context Window Usage Display

### Calculating Percentage

```bash
used=$(echo "$input" | jq -r '.context_window.used')
total=$(echo "$input" | jq -r '.context_window.total')
percentage=$((used * 100 / total))
```

### Color-Coded Display

```bash
if [ $percentage -lt 50 ]; then
  color="\033[32m"  # Green
elif [ $percentage -lt 80 ]; then
  color="\033[33m"  # Yellow
else
  color="\033[31m"  # Red
fi
echo -e "${color}Context: ${percentage}%\033[0m"
```

## Best Practices

### Keep Output Concise

Status line has limited space. Prioritize essential information:
- Model name (abbreviated if needed)
- Cost or token usage
- Context percentage
- Custom project indicators

### Use Visual Indicators

Employ emojis and colors for quick scanning:
- Green checkmark for healthy status
- Yellow warning for approaching limits
- Red alert for critical conditions

### Handle Missing Data

Always provide fallbacks for missing fields:

```bash
model=$(echo "$input" | jq -r '.model // "unknown"')
cost=$(echo "$input" | jq -r '.cost.total_usd // 0')
```

### Test with jq

Validate JSON parsing before deployment:

```bash
echo '{"model":"sonnet"}' | jq -r '.model'
```

### Update Frequency Considerations

Statusline updates at most every 300ms:
- Avoid expensive computations
- Cache values when possible
- Use efficient JSON parsing

## Configuration Options

### Custom Script Path

```json
{
  "statusLine": {
    "command": "~/.claude/scripts/my-statusline.sh"
  }
}
```

### Script with Arguments

```json
{
  "statusLine": {
    "command": "python3 ~/.claude/scripts/status.py --format=minimal"
  }
}
```

### Disable Statusline

```json
{
  "statusLine": null
}
```

## Troubleshooting

### Statusline Not Updating

Check that:
- Script is executable (chmod +x)
- Script outputs to stdout (not stderr)
- JSON parsing is correct
- No infinite loops in script

### Display Issues

If output appears garbled:
- Verify ANSI codes are correct
- Ensure single-line output
- Check for trailing newlines
- Test script manually

### Performance Issues

If statusline causes slowdown:
- Simplify JSON parsing
- Remove external command calls
- Use lightweight scripting language
- Cache computed values

## Advanced Patterns

### Project-Specific Status

Detect project type and show relevant info:

```bash
if [ -f "package.json" ]; then
  echo "Node.js Project"
elif [ -f "pyproject.toml" ]; then
  echo "Python Project"
else
  echo "Generic Project"
fi
```

### Git Integration

Show git branch in status:

```bash
branch=$(git branch --show-current 2>/dev/null || echo "no-git")
echo "Branch: $branch"
```

### Cost Alerts

Highlight when cost exceeds threshold:

```bash
cost=$(echo "$input" | jq -r '.cost.total_usd')
if (( $(echo "$cost > 1.00" | bc -l) )); then
  echo -e "\033[31mCost Alert: \$$cost\033[0m"
else
  echo "Cost: \$$cost"
fi
```
