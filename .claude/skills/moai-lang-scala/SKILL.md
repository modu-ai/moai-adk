---
name: moai-lang-scala
description: Scala best practices with ScalaTest, sbt, and functional programming
  patterns
allowed-tools:
- Read
- Bash
---

# Scala Expert

## What it does

Provides Scala-specific expertise for TDD development, including ScalaTest framework, sbt build tool, and functional programming patterns.

## When to use

- "Scala 테스트 작성", "ScalaTest 사용법", "함수형 프로그래밍"
- Automatically invoked when working with Scala projects
- Scala SPEC implementation (`/alfred:2-build`)

## How it works

**TDD Framework**:
- **ScalaTest**: Flexible testing framework
- **specs2**: BDD-style testing
- **ScalaCheck**: Property-based testing
- Test coverage with sbt-scoverage

**Build Tools**:
- **sbt**: Scala build tool
- **build.sbt**: Build configuration
- Multi-project builds

**Code Quality**:
- **Scalafmt**: Code formatting
- **Scalafix**: Linting and refactoring
- **WartRemover**: Code linting

**Functional Programming**:
- **Immutable data structures**
- **Higher-order functions**
- **Pattern matching**
- **For-comprehensions**
- **Monads (Option, Either, Try)**

**Best Practices**:
- File ≤300 LOC, method ≤50 LOC
- Prefer immutable vals over mutable vars
- Case classes for data modeling
- Tail recursion for loops
- Avoid null, use Option

## Examples

### Example 1: TDD with ScalaTest
User: "/alfred:2-build PARSER-001"
Claude: (creates RED test with ScalaTest, GREEN implementation with immutability, REFACTOR)

### Example 2: Property testing
User: "ScalaCheck 속성 테스트 작성"
Claude: (creates property-based tests for edge cases)

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Scala-specific review)
- alfred-refactoring-coach (functional refactoring)
