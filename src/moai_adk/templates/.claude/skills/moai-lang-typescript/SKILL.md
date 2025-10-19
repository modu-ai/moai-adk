---
name: moai-lang-typescript
description: TypeScript best practices with Vitest, Biome, strict typing, and npm/pnpm
  package management
allowed-tools:
- Read
- Bash
---

# TypeScript Expert

## What it does

Provides TypeScript-specific expertise for TDD development, including Vitest testing, Biome linting/formatting, strict type checking, and modern npm/pnpm package management.

## When to use

- "TypeScript 테스트 작성", "Vitest 사용법", "타입 안전성"
- Automatically invoked when working with TypeScript projects
- TypeScript SPEC implementation (`/alfred:2-build`)

## How it works

**TDD Framework**:
- **Vitest**: Fast unit testing (Jest-compatible API)
- **@testing-library**: Component testing for React/Vue
- Test coverage ≥85% with c8/istanbul

**Type Safety**:
- **strict: true** in tsconfig.json
- **noImplicitAny**, **strictNullChecks**, **strictFunctionTypes**
- Interface definitions, Generics, Type guards

**Code Quality**:
- **Biome**: Fast linter + formatter (replaces ESLint + Prettier)
- Type-safe configurations
- Import organization, unused variable detection

**Package Management**:
- **pnpm**: Fast, disk-efficient package manager (preferred)
- **npm**: Fallback option
- `package.json` + `tsconfig.json` configuration

**Best Practices**:
- File ≤300 LOC, function ≤50 LOC
- Prefer interfaces over types for public APIs
- Use const assertions for literal types
- Avoid `any`, prefer `unknown` or proper types

## Examples

### Example 1: TDD with Vitest
User: "/alfred:2-build USER-001"
Claude: (creates RED test with Vitest, GREEN implementation with strict types, REFACTOR)

### Example 2: Type checking
User: "TypeScript 타입 오류 확인"
Claude: (runs tsc --noEmit and reports type errors)

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (TypeScript-specific review)
- alfred-refactoring-coach (type-safe refactoring)
