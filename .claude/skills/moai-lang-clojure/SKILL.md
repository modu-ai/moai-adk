---
name: moai-lang-clojure
description: Clojure best practices with clojure.test, Leiningen, and immutable data
  structures
allowed-tools:
- Read
- Bash
---

# Clojure Expert

## What it does

Provides Clojure-specific expertise for TDD development, including clojure.test framework, Leiningen build tool, and immutable data structures with functional programming.

## When to use

- "Clojure 테스트 작성", "clojure.test 사용법", "불변 데이터 구조"
- Automatically invoked when working with Clojure projects
- Clojure SPEC implementation (`/alfred:2-build`)

## How it works

**TDD Framework**:
- **clojure.test**: Built-in testing library
- **midje**: BDD-style testing
- **test.check**: Property-based testing
- Test coverage with cloverage

**Build Tools**:
- **Leiningen**: Project automation, dependency management
- **deps.edn**: Official dependency tool
- **Boot**: Alternative build tool

**Code Quality**:
- **clj-kondo**: Linter for Clojure
- **cljfmt**: Code formatting
- **eastwood**: Additional linting

**Clojure Patterns**:
- **Immutable data structures**: Persistent collections
- **Pure functions**: Functional core, imperative shell
- **Threading macros**: -> and ->> for readability
- **Lazy sequences**: Infinite data processing
- **Transducers**: Composable transformations

**Best Practices**:
- File ≤300 LOC, function ≤50 LOC
- Prefer let bindings over def
- Use namespaces for organization
- Destructuring for data access
- Avoid mutable state

## Examples

### Example 1: TDD with clojure.test
User: "/alfred:2-build TRANSFORM-001"
Claude: (creates RED test with clojure.test, GREEN implementation with threading macros, REFACTOR)

### Example 2: Property testing
User: "test.check 속성 테스트"
Claude: (creates generative tests with test.check)

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Clojure-specific review)
- alfred-refactoring-coach (functional refactoring)
