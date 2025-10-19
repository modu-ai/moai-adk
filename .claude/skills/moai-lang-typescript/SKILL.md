---
name: moai-lang-typescript
description: TypeScript best practices with Vitest, Biome, strict typing, and npm/pnpm package management
allowed-tools:
  - Read
  - Bash
tier: 3
auto-load: "true"
---

# TypeScript Expert

## What it does

Provides TypeScript-specific expertise for TDD development, including Vitest testing, Biome linting/formatting, strict type checking, and modern npm/pnpm package management.

## When to use

- "TypeScript 테스트 작성", "Vitest 사용법", "타입 안전성", "Node.js 백엔드", "프론트엔드 개발", "웹 API"
- "React", "Vue.js", "Angular", "Express.js", "NestJS", "Fastify"
- "Next.js", "SvelteKit", "Astro", "Remix"
- Automatically invoked when working with TypeScript projects
- TypeScript SPEC implementation (`/alfred:2-run`)

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

## Modern TypeScript (5.0+)

**Recommended Version**: TypeScript 5.3+ for production, 5.0+ for modern features

**Modern Features**:
- **Decorators** (5.0+): ES decorator syntax support
- **const type parameters** (5.0+): Generic constraints
- **Export maps** (5.0+): Conditional exports in package.json
- **Module resolution** (5.0+): Node 16+ resolution modes
- **Type parameter defaults** (5.0+): Default generic types
- **Inferred type predicates** (5.3+): Smarter type narrowing

**Version Check**:
```bash
tsc --version  # Check TypeScript version
tsc --noEmit   # Type check without emitting
```

## Package Management Commands

### Using pnpm (Recommended - Fast & Efficient)
```bash
# Initialize project
pnpm init
pnpm create vite@latest my-app -- --template react-ts

# Install dependencies
pnpm install vitest @testing-library/react biome

# Add dependencies
pnpm add express @types/express
pnpm add -D typescript @types/node

# Remove dependencies
pnpm remove express

# List dependencies
pnpm list
pnpm list --recursive

# Update dependencies
pnpm update
pnpm up --latest

# Run scripts
pnpm run test
pnpm run dev
pnpm run build

# Install and run in one command
pnpm dlx vitest
```

### Using npm
```bash
# Initialize
npm init -y
npm create vite@latest my-app -- --template react-ts

# Install
npm install vitest @testing-library/react biome
npm install --save express @types/express
npm install --save-dev typescript @types/node

# Install specific version
npm install react@18.2.0

# Remove packages
npm uninstall express

# Update
npm update
npm update --latest

# Run
npm run test
npm run dev
```

### Using yarn
```bash
yarn init
yarn create react-app my-app
yarn add vitest @testing-library/react

# Installation
yarn install
yarn add express @types/express
yarn add --dev typescript @types/node

# Run
yarn test
yarn dev
```

## Examples

### Example 1: TDD with Vitest
User: "/alfred:2-run USER-001"
Claude: (creates RED test with Vitest, GREEN implementation with strict types, REFACTOR)

### Example 2: Type checking
User: "TypeScript 타입 오류 확인"
Claude: (runs tsc --noEmit and reports type errors)

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (TypeScript-specific review)
- alfred-refactoring-coach (type-safe refactoring)
