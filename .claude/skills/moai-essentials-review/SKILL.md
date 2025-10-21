---

name: moai-essentials-review
description: Automated code review with SOLID principles, code smells, and language-specific best practices. Use when preparing concise review checklists for code changes.
allowed-tools:
  - Read
  - Bash
  - Write
  - Edit
  - TodoWrite
---

# Alfred Code Reviewer

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Write (write_file), Edit (edit_file), Bash (terminal), TodoWrite (todo_write) |
| Auto-load | On demand during Sync stage (review gate) |
| Trigger cues | Code review requests, quality checklist preparation, merge readiness checks. |

## What it does

Automated code review with language-specific best practices, SOLID principles verification, and code smell detection.

## When to use

- Loads when someone asks for a code review or a pre-merge quality assessment.
- “Please review the code”, “How can this code be improved?”, “Check the code quality”
- Optionally invoked after `/alfred:3-sync`
- Before merging PR

## How it works

**Code Constraints Check**:
- File ≤300 LOC
- Function ≤50 LOC
- Parameters ≤5
- Cyclomatic complexity ≤10

**SOLID Principles**:
- Single Responsibility
- Open/Closed
- Liskov Substitution
- Interface Segregation
- Dependency Inversion

**Code Smell Detection**:
- Long Method
- Large Class
- Duplicate Code
- Dead Code
- Magic Numbers

**Language-specific Best Practices**:
- Python: List comprehension, type hints, PEP 8
- TypeScript: Strict typing, async/await, error handling
- Java: Streams API, Optional, Design patterns

**Review Report**:
```markdown
## Code Review Report

### 🔴 Critical Issues (3)
1. **src/auth/service.py:45** - Function too long (85 > 50 LOC)
2. **src/api/handler.ts:120** - Missing error handling
3. **src/db/repository.java:200** - Magic number

### ⚠️ Warnings (5)
1. **src/utils/helper.py:30** - Unused import

### ✅ Good Practices Found
- Test coverage: 92%
- Consistent naming
```

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
- IEEE. "Recommended Practice for Software Reviews." IEEE 1028-2008.
- Cisco. "Peer Review Best Practices." https://www.cisco.com/c/en/us/support/docs/optical/ons-15454-esc/15114-peer-review.html (accessed 2025-03-29).

## Changelog
- 2025-03-29: Overhauled input/output definitions for Essentials skills.

## Works well with

- moai-foundation-specs
- moai-essentials-refactor

## Best Practices
- Record results, even for simple improvements, to increase traceability.
- Clearly mark items that require human review to distinguish them from automation.
