---
name: moai-essentials-refactor
description: Refactoring guidance with design patterns and code improvement strategies
allowed-tools:
- Read
- Bash
- Write
- Edit
- TodoWrite
---

# Alfred Refactoring Coach

## What it does

Refactoring guidance with design pattern recommendations, code smell detection, and step-by-step improvement plans.

## When to use

- "리팩토링 도와줘", "이 코드 개선 방법은?", "디자인 패턴 적용"
- "코드 정리", "중복 제거", "함수 분리"

## How it works

**Refactoring Techniques**:
- **Extract Method**: 긴 메서드 분리
- **Replace Conditional with Polymorphism**: 조건문 제거
- **Introduce Parameter Object**: 매개변수 그룹화
- **Extract Class**: 거대한 클래스 분리

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

User: "중복 코드 제거해줘"
Claude: (identifies duplicates, suggests Extract Method, provides step-by-step plan)
## Works well with

- moai-essentials-review
