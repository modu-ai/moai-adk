---
name: moai-alfred-debugger-pro
description: Advanced debugging support with stack trace analysis, error pattern detection, and fix suggestions
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - TodoWrite
---

# Alfred Debugger Pro

## What it does

Advanced debugging support with stack trace analysis, common error pattern detection, and actionable fix suggestions.

## When to use

- “Resolve the error”, “What is the cause of this error?”, “Stack trace analysis”
- Automatically invoked on runtime errors (via debug-helper sub-agent)
- "Why not?", "Solving NullPointerException"

## How it works

**Stack Trace Analysis**:
```python
# Error example
jwt.exceptions.ExpiredSignatureError: Signature has expired

# Alfred Analysis
📍 Error Location: src/auth/service.py:142
🔍 Root Cause: JWT token has expired
💡 Fix Suggestion:
   1. Implement token refresh logic
   2. Check expiration before validation
   3. Handle ExpiredSignatureError gracefully
```

**Common Error Patterns**:
- `NullPointerException` → Optional usage, guard clauses
- `IndexError` → Boundary checks
- `KeyError` → `.get()` with defaults
- `TypeError` → Type hints, input validation
- `ConnectionError` → Retry logic, timeouts

**Debugging Checklist**:
- [ ] Reproducible?
- [ ] Log messages?
- [ ] Input data?
- [ ] Recent changes?
- [ ] Dependency versions?

**Language-specific Tips**:
- **Python**: Logging, type guards
- **TypeScript**: Type guards, null checks
- **Java**: Optional, try-with-resources

## Examples

User: "Please solve JWT ExpiredSignatureError"
Claude: (analyzes stack trace, identifies root cause, suggests fix)

## Works well with

- alfred-code-reviewer
- alfred-trust-validation
