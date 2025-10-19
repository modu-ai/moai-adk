---
name: moai-essentials-review
description: Automated code review with SOLID principles, code smells, and language-specific
  best practices
allowed-tools:
- Read
- Bash
- Write
- Edit
- TodoWrite
---

# Alfred Code Reviewer

## What it does

Automated code review with language-specific best practices, SOLID principles verification, and code smell detection.

## When to use

- "코드 리뷰해줘", "이 코드 개선점은?", "코드 품질 확인"
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

User: "이 코드 리뷰해줘"
Claude: (analyzes code, detects issues, provides improvement suggestions)
## Works well with

- moai-foundation-specs
- moai-essentials-refactor
