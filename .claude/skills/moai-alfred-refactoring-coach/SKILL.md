---
name: moai-alfred-refactoring-coach
description: Refactoring guidance with design patterns, code smells detection, and step-by-step improvement plans
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - TodoWrite
---

# Alfred Refactoring Coach

## What it does

Refactoring guidance with design pattern recommendations, code smell detection, and step-by-step improvement plans.

## When to use

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

User: "Remove duplicate code"
Claude: (identifies duplicates, suggests Extract Method, provides step-by-step plan)

## Works well with

- alfred-code-reviewer
- alfred-trust-validation
