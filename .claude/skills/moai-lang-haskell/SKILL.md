---
name: moai-lang-haskell
description: Haskell best practices with HUnit, Stack/Cabal, and pure functional programming
allowed-tools:
- Read
- Bash
---

# Haskell Expert

## What it does

Provides Haskell-specific expertise for TDD development, including HUnit testing, Stack/Cabal build tools, and pure functional programming with strong type safety.

## When to use

- "Haskell 테스트 작성", "HUnit 사용법", "순수 함수형 프로그래밍"
- Automatically invoked when working with Haskell projects
- Haskell SPEC implementation (`/alfred:2-build`)

## How it works

**TDD Framework**:
- **HUnit**: Unit testing framework
- **QuickCheck**: Property-based testing
- **Hspec**: BDD-style testing
- Test coverage with hpc

**Build Tools**:
- **Stack**: Reproducible builds, dependency resolution
- **Cabal**: Haskell package system
- **hpack**: Alternative package description

**Code Quality**:
- **hlint**: Haskell linter
- **stylish-haskell**: Code formatting
- **GHC warnings**: Compiler-level checks

**Functional Programming**:
- **Pure functions**: No side effects
- **Monads**: IO, Maybe, Either, State
- **Functors/Applicatives**: Abstraction patterns
- **Type classes**: Polymorphism
- **Lazy evaluation**: Infinite data structures

**Best Practices**:
- File ≤300 LOC, function ≤50 LOC
- Prefer total functions (avoid partial)
- Type-driven development
- Point-free style (when readable)
- Avoid do-notation overuse

## Examples

### Example 1: TDD with HUnit
User: "/alfred:2-build PARSE-001"
Claude: (creates RED test with HUnit, GREEN implementation with pure functions, REFACTOR)

### Example 2: Property testing
User: "QuickCheck 속성 테스트"
Claude: (creates property-based tests for invariants)

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Haskell-specific review)
- alfred-refactoring-coach (functional refactoring)
