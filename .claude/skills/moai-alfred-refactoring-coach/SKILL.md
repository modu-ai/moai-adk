---

name: moai-alfred-refactoring-coach
description: Refactoring guidance with design patterns, code smells detection, and step-by-step improvement plans. Use when outlining refactor steps that preserve functionality.
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - TodoWrite
---

# Alfred Refactoring Coach

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Write (write_file), Edit (edit_file), Bash (terminal), TodoWrite (todo_write) |
| Auto-load | /alfred:2-run refactor lane |
| Trigger cues | Refactoring retros, duplication cleanup, code health follow-ups inside Alfred. |

## What it does

Refactoring guidance with design pattern recommendations, code smell detection, and step-by-step improvement plans.

## When to use

- Activates when Alfred is asked to plan or stage refactoring work.
- “Help with refactoring”, “How can I improve this code?”, “Apply design patterns” 
- “Code organization”, “Remove duplication”, “Separate functions”

## How it works

**Refactoring Techniques**:
- **Extract Method**: Separate long methods
- **Replace Conditional with Polymorphism**: Remove conditional statements
- **Introduce Parameter Object**: Group parameters
- **Extract Class**: Massive class separation

**Design Pattern Recommendations**:
- Complex object creation → **Builder Pattern**
- Type-specific behavior → **Strategy Pattern**
- Global state → **Singleton Pattern**
- Incompatible interfaces → **Adapter Pattern**
- Delayed object creation → **Factory Pattern**

**3-Strike Rule**:
```
1st occurrence: Just implement
2nd occurrence: Notice similarity (leave as-is)
3rd occurrence: Pattern confirmed → Refactor! 🔧
```

**Refactoring Checklist**:
- [ ] All tests passing before refactoring
- [ ] Code smells identified
- [ ] Refactoring goal clear
- [ ] Change one thing at a time
- [ ] Run tests after each change
- [ ] Commit frequently

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
- Fowler, Martin. "Refactoring: Improving the Design of Existing Code." Addison-Wesley, 2018.
- IEEE Software. "Managing Technical Debt." IEEE Software, 2021.

## Changelog
- 2025-03-29: Added input/output/failure response to Alfred-specific skills.

## Works well with

- alfred-code-reviewer
- alfred-trust-validation
