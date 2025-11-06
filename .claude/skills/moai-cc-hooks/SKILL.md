---
name: moai-cc-hooks
description: "Configuring Claude Code Hooks System. Design, implement, and manage PreToolUse/PostToolUse/SessionStart/Notification/Stop hooks. Use when enforcing safety checks, auto-formatting, running linters, or triggering automated workflows based on development events."
allowed-tools: "Read, Write, Edit, Glob, Bash"
---

## Skill Metadata

| Field | Value |
| ----- | ----- |
| Version | 1.0.0 |
| Tier | Ops |
| Auto-load | When configuring hook system |

## What It Does

Hook system 설정 및 관리를 위한 전체 가이드를 제공합니다. PreToolUse, PostToolUse, SessionStart 등 다양한 hook type의 설계와 구현 방법을 다룹니다.

## When to Use

- Safety check나 validation hook을 추가할 때
- Auto-formatting이나 linting을 자동화할 때
- Session lifecycle event에 응답하는 로직을 작성할 때
- 개발 이벤트 기반 workflow를 트리거할 때


# Configuring Claude Code Hooks System

Hooks are lightweight (<100ms) automated scripts triggered by Claude Code lifecycle events. They enforce guardrails, run quality checks, and seed context without blocking user workflow.

## Hook Types & Events

| Hook Type | Trigger Event | Execution Time | Use Cases |
|-----------|---------------|-----------------|-----------|
| **PreToolUse** | Before tool execution (Read, Edit, Bash, etc.) | <100ms | Command validation, permission checks, safety gates |
| **PostToolUse** | After tool execution succeeds | <100ms | Auto-formatting, linting, permission restoration |
| **SessionStart** | Session initialization | <500ms | Project summary, context seeding, status card |
| **Notification** | User notification event | N/A | macOS notifications, alerts, wait status |
| **Stop** | Session termination | N/A | Cleanup, final checks, task summary |

## Hook Configuration in settings.json

### High-Freedom Approach: Events-Driven Hooks

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "node ~/.claude/hooks/pre-bash-validator.js"
          }
        ]
      },
      {
        "matcher": "Edit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "bash ~/.claude/hooks/pre-edit-guard.sh"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Edit",
        "hooks": [
          {
            "type": "command",
            "command": "bash ~/.claude/hooks/post-edit-lint.sh"
          }
        ]
      }
    ],
    "SessionStart": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "node ~/.claude/hooks/session-status-card.js"
          }
        ]
      }
    ]
  }
}
```

## Medium-Freedom: Hook Patterns

### Pattern 1: Pre-Command Validation
```bash
#!/bin/bash
# pre-bash-validator.sh: Block dangerous patterns

FORBIDDEN_PATTERNS=(
  "rm -rf /"
  "sudo rm"
  "chmod 777 /"
  "eval \$(curl"
)

COMMAND="$1"
for pattern in "${FORBIDDEN_PATTERNS[@]}"; do
  if [[ "$COMMAND" =~ $pattern ]]; then
    echo "🔴 Blocked: $pattern detected" >&2
    exit 2  # Block execution
  fi
done

exit 0  # Allow execution
```

### Pattern 2: Post-Edit Auto-Formatting
```bash
#!/bin/bash
# post-edit-format.sh: Auto-format after edits

FILE="$1"
EXT="${FILE##*.}"

case "$EXT" in
  js|ts)
    npx prettier --write "$FILE" 2>/dev/null
    ;;
  py)
    python3 -m black "$FILE" 2>/dev/null
    ;;
  go)
    gofmt -w "$FILE" 2>/dev/null
    ;;
esac

exit 0
```

### Pattern 3: SessionStart Status Card
```bash
#!/bin/bash
# session-status-card.sh: Show project status

echo "🚀 Claude Code Session Started"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ -f ".moai/config.json" ]; then
  PROJECT=$(jq -r '.name' .moai/config.json 2>/dev/null || echo "Unknown")
  echo "📦 Project: $PROJECT"
  echo "🏗️  Framework: $(jq -r '.tech_stack' .moai/config.json 2>/dev/null || echo "Auto-detect")"
fi

echo "📋 Recent SPECS:"
ls .moai/specs/ 2>/dev/null | head -3 | sed 's/^/  ✓ /'

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
```

## Low-Freedom: Security-Focused Hook Scripts

### Permission Preservation Hook
```bash
#!/bin/bash
set -euo pipefail
# preserve-permissions.sh: Save/restore file permissions

HOOK_TYPE="${1:-pre}"  # 'pre' or 'post'
FILE="${2:-.}"

PERMS_FILE="/tmp/perms_${FILE//\//_}.txt"

if [[ "$HOOK_TYPE" == "pre" ]]; then
  stat -c "%a %u:%g" "$FILE" > "$PERMS_FILE" 2>/dev/null || true
  exit 0
elif [[ "$HOOK_TYPE" == "post" ]]; then
  if [[ -f "$PERMS_FILE" ]]; then
    SAVED_PERMS=$(cat "$PERMS_FILE")
    chmod ${SAVED_PERMS%% *} "$FILE" 2>/dev/null || true
    rm "$PERMS_FILE"
  fi
  exit 0
fi
```

### Dangerous Command Blocker
```python
#!/usr/bin/env python3
import json
import sys
import re

BLOCKED_PATTERNS = [
    (r"rm\s+-rf\s+/", "Blocking rm -rf /"),
    (r"sudo\s+rm", "Blocking sudo rm without confirmation"),
    (r">\s*/etc/\w+", "Blocking writes to system files"),
    (r"curl.*\|\s*bash", "Blocking curl | bash (code injection risk)"),
]

try:
    data = json.load(sys.stdin)
    command = data.get("tool_input", {}).get("command", "")

    for pattern, msg in BLOCKED_PATTERNS:
        if re.search(pattern, command):
            print(f"🔴 BLOCKED: {msg}", file=sys.stderr)
            sys.exit(2)  # Block

    sys.exit(0)  # Allow
except Exception as e:
    print(f"Hook error: {e}", file=sys.stderr)
    sys.exit(0)  # Allow on error (fail open)
```

## Hook Exit Codes

| Code | Meaning | Behavior |
|------|---------|----------|
| `0` | Success | Tool proceeds normally |
| `1` | Warning + Stderr | Warning logged, tool proceeds |
| `2` | Blocked + Error | Tool execution blocked |

## Best Practices

✅ **DO**:
- Keep hook scripts < 100ms execution time
- Use specific matchers (not wildcard `*` unless necessary)
- Log errors clearly to stderr
- Test hooks before deploying to team

❌ **DON'T**:
- Make network calls in PreToolUse hooks
- Block common operations (use warnings instead)
- Write to user files from hooks
- Create complex logic (delegate to sub-agents)

## Hook Validation Checklist

- [ ] All scripts have proper shebang (`#!/bin/bash` or `#!/usr/bin/env python3`)
- [ ] Scripts are executable: `chmod +x hook.sh`
- [ ] Exit codes are correct (0, 1, or 2)
- [ ] Paths are absolute (not relative)
- [ ] Tested for < 100ms latency
- [ ] Error messages are user-friendly
- [ ] No hardcoded secrets or credentials

---

**Reference**: Context7 Claude Code Hooks documentation
**Version**: 1.0.0
