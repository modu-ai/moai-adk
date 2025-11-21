---

name: moai-cc-hook-model-strategy
description: Claude Code hooks lifecycle, model strategy patterns, and security automation
allowed-tools: [Read, Bash]

---

# Claude Code Hook Model Strategy

## Quick Reference

Claude Code hooks are shell commands that execute at strategic points in Claude's lifecycle - transforming Claude from a reactive assistant into a proactive, security-aware development partner. This skill covers the 10 lifecycle events, model strategy patterns, and security automation frameworks for enterprise-grade Claude Code deployments.

**Core Hook Events** (November 2025):
- **PreToolUse**: Block dangerous operations before execution (security gate)
- **PostToolUse**: Validate output after tool completion (quality gate)
- **UserPromptSubmit**: Filter/enhance prompts before Claude processes (intent validation)
- **PermissionRequest**: Custom permission logic with allow/deny capability
- **Stop**: Execute cleanup when Claude finishes responding
- **SessionStart/SessionEnd**: Initialize/teardown session-level resources
- **SubagentStop**: Coordinate multi-agent workflows
- **PreCompact**: Prepare for context compression
- **Notification**: React to Claude Code system events

**Security-First Principle**: Hooks run with your environment's credentials - malicious code can exfiltrate data. Always review implementations before registration.


## Implementation Guide

### Phase 1: Security-First Hook Architecture

**File Protection Pattern** (Block sensitive file modifications):
```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Edit|MultiEdit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "python3 -c \"import json, sys; data=json.load(sys.stdin); path=data.get('tool_input',{}).get('file_path',''); sys.exit(2 if any(p in path for p in ['.env', 'package-lock.json', '.git/', 'secrets.yaml']) else 0)\""
          }
        ]
      }
    ]
  }
}
```

**Exit Code Strategy**:
- `0`: Allow operation (success)
- `1`: Warning (operation proceeds with logged message)
- `2`: **Block operation** (hard stop - tool execution canceled)

**Dangerous Command Detection**:
```python
#!/usr/bin/env python3
import json
import sys
import re

DANGEROUS_PATTERNS = [
    (r'rm\s+-rf\s+/', 'System deletion attempt'),
    (r'sudo\s+rm', 'Privileged deletion'),
    (r'chmod\s+777', 'Insecure permissions'),
    (r'curl.*\|\s*sh', 'Remote script execution'),
    (r'eval\s*\(', 'Code evaluation risk'),
]

input_data = json.loads(sys.stdin.read())
command = input_data.get('tool_input', {}).get('command', '')

for pattern, reason in DANGEROUS_PATTERNS:
    if re.search(pattern, command, re.IGNORECASE):
        print(f"BLOCKED: {reason}", file=sys.stderr)
        sys.exit(2)  # Hard block

sys.exit(0)  # Allow
```

### Phase 2: Model Strategy Patterns

**Deterministic Quality Gates** (Post-execution formatting):
```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "if [[ \"$(jq -r '.tool_input.file_path' | grep -E '\\.(ts|js)$')\" ]]; then prettier --write $(jq -r '.tool_input.file_path'); fi"
          }
        ]
      }
    ]
  }
}
```

**Compliance Logging** (Audit trail for regulated environments):
```python
#!/usr/bin/env python3
from datetime import datetime
import json
import sys
from pathlib import Path

input_data = json.loads(sys.stdin.read())
tool_name = input_data.get('tool_name', 'unknown')
tool_input = input_data.get('tool_input', {})
session_id = input_data.get('session', {}).get('id', 'unknown')

log_entry = {
    "timestamp": datetime.utcnow().isoformat() + 'Z',
    "session_id": session_id,
    "tool": tool_name,
    "input": tool_input
}

log_file = Path("logs/compliance-audit.jsonl")
log_file.parent.mkdir(parents=True, exist_ok=True)

with open(log_file, 'a') as f:
    f.write(json.dumps(log_entry) + '\n')

sys.exit(0)
```

### Phase 3: Advanced Orchestration

**Multi-Hook Coordination** (Chain validation + logging + notification):
```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "uv run .claude/hooks/validate_bash.py"
          },
          {
            "type": "command",
            "command": "uv run .claude/hooks/log_bash.py"
          }
        ]
      }
    ],
    "Stop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "uv run .claude/hooks/notify_completion.py"
          }
        ]
      }
    ]
  }
}
```

**Context Injection Pattern** (Enhance prompts with project-specific context):
```python
#!/usr/bin/env python3
import json
import sys

input_data = json.loads(sys.stdin.read())
prompt = input_data.get('prompt', '')

# Inject context
enhanced_prompt = f"""
Project Context:
- Tech Stack: FastAPI, PostgreSQL, Redis
- Test Coverage Target: ≥85%
- Security: OWASP Top 10 compliance required

Original Request:
{prompt}
"""

# Return enhanced prompt (Claude will see this instead)
output = input_data.copy()
output['prompt'] = enhanced_prompt
print(json.dumps(output))
sys.exit(0)
```


## Advanced Patterns

### Enterprise Security Framework

**Multi-Layer Defense**:
```
Layer 1: UserPromptSubmit
  ├─ Validate intent (block malicious requests)
  ├─ Check against policy database
  └─ Log for audit trail

Layer 2: PreToolUse
  ├─ Validate tool arguments (SQL injection, path traversal)
  ├─ Check file permissions
  └─ Block dangerous patterns

Layer 3: PostToolUse
  ├─ Validate output (no credentials leaked)
  ├─ Format code (deterministic quality)
  └─ Update metrics dashboard

Layer 4: Stop
  ├─ Generate session report
  ├─ Archive logs
  └─ Notify stakeholders
```

**Configuration Storage**: `~/.claude/settings.json`
```json
{
  "permissions": {
    "allow": [
      "Bash(uv:*)",
      "Write",
      "Read"
    ],
    "deny": [
      "Bash(rm:*)",
      "Bash(sudo:*)"
    ]
  },
  "hooks": {
    "PreToolUse": [...],
    "PostToolUse": [...],
    "UserPromptSubmit": [...],
    "Stop": [...]
  }
}
```

### Performance Optimization

**Hook Execution Budget**:
- Target: <100ms per hook (prevent UX lag)
- Use compiled binaries for hot paths (Rust/Go)
- Cache expensive validations (file existence checks)
- Parallelize independent hooks (async execution)

**Monitoring Pattern**:
```python
import time
start = time.time()
# ... hook logic ...
elapsed = time.time() - start
if elapsed > 0.1:  # 100ms threshold
    print(f"WARN: Hook slow ({elapsed:.2f}s)", file=sys.stderr)
```


## Best Practices

### ✅ DO
- Review all hooks before deployment (security audit)
- Use exit code 2 for hard blocks (vs 0 for allow)
- Log security events to immutable storage (compliance)
- Test hooks in isolation before enabling
- Version control your hooks (`~/.claude/hooks/` directory)
- Document hook behavior in `.claude/HOOKS.md`
- Monitor hook performance (<100ms target)
- Use Python `uv run` for dependency management

### ❌ DON'T
- Hardcode credentials in hooks (use environment variables)
- Trust user input without validation (sanitize all paths)
- Block legitimate operations (test thoroughly)
- Ignore hook failures (monitor stderr output)
- Skip performance testing (avoid UX degradation)
- Modify Claude's response without user consent
- Use hooks for business logic (keep them lightweight)


## Lifecycle Event Reference

| Event | When It Fires | Use Cases | Exit Code Behavior |
|-------|---------------|-----------|-------------------|
| **UserPromptSubmit** | Before Claude sees prompt | Intent validation, context injection, security filtering | 2 = block prompt |
| **PreToolUse** | Before tool execution | Validate arguments, block dangerous ops, log commands | 2 = block tool |
| **PostToolUse** | After tool completes | Format code, validate output, update metrics | N/A (informational) |
| **PermissionRequest** | Permission dialog shown | Custom permission logic | 2 = deny permission |
| **Stop** | Claude finishes response | Session summary, cleanup, notifications | N/A |
| **SessionStart** | Session begins/resumes | Initialize resources, load context | N/A |
| **SessionEnd** | Session terminates | Archive logs, teardown resources | N/A |
| **SubagentStop** | Subagent task completes | Multi-agent coordination | N/A |
| **PreCompact** | Before context compression | Save critical context | N/A |
| **Notification** | Claude sends notification | Custom notification handling | N/A |


## Works Well With

- `moai-cc-permission-mode` (Permission system integration)
- `moai-cc-subagent-lifecycle` (Multi-agent coordination)
- `moai-security-api` (Security validation patterns)
- `moai-foundation-trust` (TRUST 5 quality gates)


**Version**: 1.0.0  
**Last Updated**: 2025-11-21  
**Status**: Production Ready (Enterprise)  
**Official Reference**: https://code.claude.com/docs/en/hooks-guide  
**Community Patterns**: https://github.com/disler/claude-code-hooks-mastery
