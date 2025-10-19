---
name: moai-lang-javascript
description: JavaScript best practices with Jest, ESLint, Prettier, and npm package
  management
allowed-tools:
- Read
- Bash
---

# JavaScript Expert

## What it does

Provides JavaScript-specific expertise for TDD development, including Jest testing, ESLint linting, Prettier formatting, and npm package management.

## When to use

- "JavaScript 테스트 작성", "Jest 사용법", "ES6+ 문법"
- Automatically invoked when working with JavaScript projects
- JavaScript SPEC implementation (`/alfred:2-build`)

## How it works

**TDD Framework**:
- **Jest**: Unit testing with mocking, snapshots
- **@testing-library**: DOM/React testing
- Test coverage ≥85% enforcement

**Code Quality**:
- **ESLint**: JavaScript linting with recommended rules
- **Prettier**: Code formatting (opinionated)
- **JSDoc**: Type hints via comments (for type safety)

**Package Management**:
- **npm**: Standard package manager
- **package.json** for dependencies and scripts
- Semantic versioning

**Modern JavaScript**:
- ES6+ features (arrow functions, destructuring, spread/rest)
- Async/await over callbacks
- Module imports (ESM) over CommonJS

**Best Practices**:
- File ≤300 LOC, function ≤50 LOC
- Prefer `const` over `let`, avoid `var`
- Guard clauses for early returns
- Meaningful names, avoid abbreviations

## Examples

### Example 1: TDD with Jest
User: "/alfred:2-build API-001"
Claude: (creates RED test with Jest, GREEN implementation, REFACTOR with JSDoc)

### Example 2: Linting
User: "ESLint 실행"
Claude: (runs eslint . and reports linting errors)

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (JavaScript-specific review)
- alfred-debugger-pro (JavaScript debugging)
