---

name: moai-essentials-debug
description: Advanced debugging with stack trace analysis, error pattern detection, and fix suggestions. Use when delivering quick diagnostic support for everyday issues.
allowed-tools:
  - Read
  - Bash
  - Write
  - Edit
  - TodoWrite
---

# Alfred Debugger Pro

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Write (write_file), Edit (edit_file), Bash (terminal), TodoWrite (todo_write) |
| Auto-load | On demand during Run stage (debug-helper) |
| Trigger cues | Runtime error triage, stack trace analysis, root cause investigation requests. |

## What it does

Advanced debugging support with stack trace analysis, common error pattern detection, and actionable fix suggestions.

## When to use

- Loads when users share stack traces or ask why a failure occurred.
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
```markdown
- Checks the current diff and lists items that can be modified immediately.
- Schedule follow-up tasks with TodoWrite.
```

## Inputs
- A snapshot of the code/tests/documentation you are currently working on.
- Ongoing agent status information.

## Outputs
- Immediately actionable checklists or improvement suggestions.
- Recommendations on whether to take next steps or not.

## Failure Modes
- If you cannot find the required files or test results.
- When the scope of work is excessively large and cannot be resolved with simple support.

## Dependencies
- Mainly used in conjunction with `tdd-implementer`, `quality-gate`, etc.

## References
- Microsoft. "Debugging Techniques." https://learn.microsoft.com/visualstudio/debugger/ (accessed 2025-03-29).
- JetBrains. "Debugging Code." https://www.jetbrains.com/help/idea/debugging-code.html (accessed 2025-03-29).

## Changelog
- 2025-03-29: Overhauled input/output definitions for Essentials skills.

## Works well with

- moai-essentials-refactor

## Best Practices
- Record results, even for simple improvements, to increase traceability.
- Clearly mark items that require human review to distinguish them from automation.
