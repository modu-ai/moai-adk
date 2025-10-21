---

name: moai-alfred-code-reviewer
description: Automated code review with language-specific best practices, SOLID principles, and actionable improvement suggestions. Use when reviewing code changes for issues and strengths.
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - TodoWrite
---

# Alfred Code Reviewer

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Write (write_file), Edit (edit_file), Bash (terminal), TodoWrite (todo_write) |
| Auto-load | /alfred:3-sync review phase |
| Trigger cues | Alfred-driven review summaries, diff inspection, merge gate decisions. |

## What it does

Automated code review with language-specific best practices, SOLID principles verification, and code smell detection.

## When to use

- Activates when Alfred needs to summarize diffs or prepare merge feedback.
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

## Best Practices
- The text shown to the user is written using TUI/report expressions.
- When running the tool, a summary of commands and results are recorded.

## Examples
```markdown
- Call this skill inside the /alfred command to generate a report.
- Add summary to Completion Report.
```

## Inputs
- MoAI-ADK project context (`.moai/project/`, `.claude/` templates, etc.).
- Parameters passed from user commands or higher commands.

## Outputs
- Reports, checklists or recommendations for your Alfred workflow.
- Structured data for subsequent subagent calls.

## Failure Modes
- When required input documents are missing or permissions are limited.
- When disruptive changes are required without user approval.

## Dependencies
- Cooperation with higher-level agents such as cc-manager and project-manager is required.

## References
- Google. "Code Review Developer Guide." https://google.github.io/eng-practices/review/ (accessed 2025-03-29).
- IEEE. "Code Review Best Practices." IEEE Software, 2022.

## Changelog
- 2025-03-29: Added input/output/failure response to Alfred-specific skills.

## Works well with

- alfred-trust-validation
- alfred-refactoring-coach
