---
name: moai-cc-hook-model-strategy
version: 4.0.0
updated: 2025-11-19
status: stable
stability: stable
description: Claude Code hook execution strategies, timing models, sequencing patterns, and orchestration for automated workflow enforcement. Covers all hook types and execution timing models.
focus_areas:
  - Hook types and timing models
  - Sequential vs parallel execution
  - Hook composition and chaining
  - Error handling in hooks
keywords:
  - hooks
  - automation
  - execution
  - validation
  - pre-tool
  - post-tool
  - session
allowed-tools: Read, Glob, Bash
---

# Claude Code Hook Execution Strategies

## Overview

Hooks provide automation triggers at strategic points in Claude Code workflows: pre-tool execution (validation), post-tool execution (processing), and session lifecycle events (initialization, cleanup). Enable systematic enforcement of quality standards and security policies.

## Hook Architecture

### Three Hook Types

```
WORKFLOW TIMELINE
│
├─ Session Start
│  └─ SessionStart hooks trigger
│
├─ Tool Invocation
│  ├─ Pre-Tool hooks trigger (validation)
│  ├─ Tool executes
│  └─ Post-Tool hooks trigger (processing)
│
├─ Task Completion
│  └─ SessionEnd hooks trigger (cleanup)
│
└─ Session Close
```

## Hook Types & Timing Models

### Pre-Tool Hooks (Validation)

Execute **BEFORE** tool runs:

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "python3 .claude/hooks/validate-command.py"
          }
        ]
      }
    ]
  }
}
```

**Execution Timing**:
```
User: bash("git commit -m 'message'")
  ↓
Pre-Tool Hook: validate-command.py
  ├─ Check command syntax
  ├─ Verify permissions
  ├─ Scan for dangerous patterns
  └─ Decision: ALLOW / DENY
       ↓
       ALLOW → Tool executes
       DENY → Block & report error
```

**Use Cases**:
- Command validation (security)
- Permission checking
- Resource verification
- Input sanitization
- Pattern matching

**Example Hook Script**:

```python
#!/usr/bin/env python3
import re
import sys
import json

DANGEROUS_PATTERNS = [
    r"rm -rf",
    r"sudo ",
    r"chmod 777",
    r"&&.*rm",
    r"\|.*sh"
]

def validate_command(command):
    for pattern in DANGEROUS_PATTERNS:
        if re.search(pattern, command):
            return False, f"Dangerous pattern: {pattern}"
    return True, "Safe"

if __name__ == "__main__":
    input_data = json.load(sys.stdin)
    command = input_data.get("command", "")
    is_safe, message = validate_command(command)

    if not is_safe:
        print(f"SECURITY BLOCK: {message}", file=sys.stderr)
        sys.exit(2)  # Non-zero = block
    sys.exit(0)  # Zero = allow
```

### Post-Tool Hooks (Processing)

Execute **AFTER** tool completes:

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Edit",
        "hooks": [
          {
            "type": "command",
            "command": "python3 .claude/hooks/post-edit-processing.py"
          }
        ]
      }
    ]
  }
}
```

**Execution Timing**:
```
Tool executes: Edit(src/auth.py)
  ↓
Post-Tool Hook: post-edit-processing.py
  ├─ Validate syntax
  ├─ Run linter
  ├─ Check formatting
  └─ Log results
       ↓
Returns to user with processing results
```

**Use Cases**:
- Result validation
- Cleanup operations
- Logging and metrics
- Side-effect handling
- Output formatting

### Session Hooks (Lifecycle)

Execute at session start/end:

```json
{
  "hooks": {
    "SessionStart": [
      {
        "type": "command",
        "command": "uv run .moai/scripts/statusline.py"
      }
    ],
    "SessionEnd": [
      {
        "type": "command",
        "command": "python3 .claude/hooks/cleanup.py"
      }
    ]
  }
}
```

**Timing**:
```
SESSION START
  ↓
SessionStart hooks trigger
  ├─ Load project config
  ├─ Initialize memory
  ├─ Check dependencies
  └─ Display status
       ↓
WORK PHASE
       ↓
SESSION END
  ↓
SessionEnd hooks trigger
  ├─ Save state
  ├─ Cleanup temp files
  ├─ Generate report
  └─ Close connections
```

**Use Cases**:
- Context initialization
- State setup
- Health checks
- Cleanup on exit
- Report generation

## Execution Models

### Sequential Execution (Dependent Hooks)

Hooks run one after another, in order:

```
Tool Call
  ↓
Pre-Hook 1 (VALIDATE INPUT)
  ├─ Check syntax → Pass/Fail
  └─ Output: validated_input
       ↓
Pre-Hook 2 (CHECK PERMISSIONS)
  ├─ Verify permissions → Pass/Fail
  └─ Output: permission_result
       ↓
Pre-Hook 3 (AUDIT LOG)
  ├─ Log operation → Always pass
  └─ Output: audit_entry
       ↓
Tool Execution
  ↓
Post-Hook 1 (VALIDATE OUTPUT)
Post-Hook 2 (FORMAT RESULTS)
Post-Hook 3 (UPDATE METRICS)
  ↓
Result
```

**When to use sequential**:
- When hooks depend on previous results
- When order matters (e.g., validation before execution)
- When failure of one hook should block others

### Parallel Execution (Independent Hooks)

Hooks run simultaneously when independent:

```
Tool Call
  ↓
Pre-Hooks (all in parallel):
  ├─ Thread 1: Validate syntax
  ├─ Thread 2: Check permissions
  ├─ Thread 3: Scan for secrets
  └─ Thread 4: Verify resources
       ↓
       (All must pass for tool to execute)
  ↓
Tool Execution
  ↓
Post-Hooks (all in parallel):
  ├─ Thread 1: Validate output
  ├─ Thread 2: Update metrics
  ├─ Thread 3: Log operation
  └─ Thread 4: Format results
  ↓
Result
```

**When to use parallel**:
- When hooks are independent
- For performance optimization
- When order doesn't matter
- For I/O operations

## Hook Composition & Chaining

### Hook Chain Pattern

```python
# Pre-execution chain:
# 1. Validation Hook    (input validation)
# 2. Permission Hook    (authorization)
# 3. Resource Hook      (availability check)
# 4. Audit Hook         (logging)
# 5. Execution         (actual tool)

# Post-execution chain:
# 1. Output Hook        (result validation)
# 2. Processing Hook    (transformation)
# 3. Cleanup Hook       (resource release)
# 4. Metrics Hook       (usage tracking)
```

### Conditional Hook Execution

```bash
#!/bin/bash
# Hook only runs if specific conditions met

if [[ "$TOOL_TYPE" == "Edit" ]]; then
  # Run linter for edit operations
  python3 -m pylint --score=no "$FILE" || exit 1
fi

if [[ "$COMMAND" == "rm -rf" ]]; then
  echo "BLOCKED: Destructive command"
  exit 2
fi

exit 0
```

## Error Handling in Hooks

### Hook Failure Handling

```python
# Hook exit codes:
# 0 = Success (allow/continue)
# 1 = Warning (log but continue)
# 2 = Error (block/stop)

#!/usr/bin/env python3
import sys

try:
    # Perform validation
    result = validate_operation()
    
    if not result.valid:
        # Block operation
        print(f"BLOCK: {result.error}", file=sys.stderr)
        sys.exit(2)  # Exit code 2 = block
    
except Exception as e:
    # Unexpected error
    print(f"ERROR: {e}", file=sys.stderr)
    sys.exit(1)  # Exit code 1 = warning

# Success
sys.exit(0)  # Exit code 0 = allow
```

### Hook Timeout Handling

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "command": "python3 .claude/hooks/fast-validation.py",
        "timeout": 5000  # 5 second timeout
      }
    ]
  }
}
```

## Best Practices

### Hook Design

- ✅ Keep hooks lightweight (< 1 second)
- ✅ Implement proper error handling
- ✅ Use sequential for dependent hooks
- ✅ Use parallel for independent hooks
- ✅ Log hook execution
- ✅ Test hooks thoroughly
- ✅ Document hook purpose

### Performance

- ✅ Minimize hook execution time
- ✅ Cache validation results
- ✅ Use parallel execution where possible
- ✅ Avoid blocking I/O in hooks
- ✅ Monitor hook performance

### Security

- ✅ Validate all hook inputs
- ✅ Handle edge cases
- ✅ Log security decisions
- ✅ Test with malicious input
- ✅ Use secure exit codes

## Common Hook Scenarios

### Scenario 1: Pre-Commit Security Check

```bash
#!/bin/bash
# Pre-commit hook: validate before git commit

# Check for secrets
if grep -r "PASSWORD\|API_KEY\|SECRET" --include="*.py" . ; then
  echo "ERROR: Secrets found in code"
  exit 2
fi

# Check code formatting
python3 -m black --check . || exit 2

# All checks pass
exit 0
```

### Scenario 2: Post-Edit Code Formatting

```bash
#!/bin/bash
# Post-edit hook: auto-format edited files

# Get edited file from input
FILE="$1"

# Auto-format
python3 -m black "$FILE" || exit 1
python3 -m ruff check --fix "$FILE" || exit 1

echo "Formatted: $FILE"
exit 0
```

### Scenario 3: Session Initialization

```bash
#!/bin/bash
# SessionStart hook: initialize project context

# Load project status
echo "Loading project configuration..."
python3 .moai/scripts/statusline.py

# Check dependencies
echo "Verifying dependencies..."
uv sync --quiet

echo "Session ready!"
exit 0
```

## TRUST 5 Compliance

- **Test-First**: Hook patterns tested with real execution scenarios
- **Readable**: Clear hook structure, well-documented intent, self-explanatory
- **Unified**: Consistent hook implementation across all types
- **Secured**: Validation and security hooks, proper error handling
- **Trackable**: Full hook execution logging and audit trail

## Related Skills

- `moai-cc-permission-mode` - Permission validation in hooks
- `moai-cc-subagent-lifecycle` - Lifecycle hooks for subagents
- `moai-core-workflow` - Overall workflow orchestration

---

**Last Updated**: 2025-11-19
**Version**: 4.0.0
**Enterprise Production Ready**: Yes ✅
**Maturity**: Stable
