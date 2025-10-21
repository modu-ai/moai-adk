---

name: moai-essentials-refactor
description: Refactoring guidance with design patterns and code improvement strategies. Use when planning incremental refactors with safety nets.
allowed-tools:
  - Read
  - Bash
  - Write
  - Edit
  - TodoWrite
---

# Alfred Refactoring Coach

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Write (write_file), Edit (edit_file), Bash (terminal), TodoWrite (todo_write) |
| Auto-load | On demand during Run stage (refactor planning) |
| Trigger cues | Refactoring plans, code smell cleanup, design pattern coaching. |

## What it does

Refactoring guidance with design pattern recommendations, code smell detection, and step-by-step improvement plans.

## When to use

- Loads when the user asks how to restructure code or apply design patterns.
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
- Fowler, Martin. "Refactoring: Improving the Design of Existing Code." Addison-Wesley, 2018.
- IEEE Software. "Managing Technical Debt." IEEE Software, 2021.

## Changelog
- 2025-03-29: Overhauled input/output definitions for Essentials skills.

## Works well with

- moai-essentials-review

## Best Practices
- Record results, even for simple improvements, to increase traceability.
- Clearly mark items that require human review to distinguish them from automation.
