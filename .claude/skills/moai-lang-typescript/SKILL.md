---

name: moai-lang-typescript
description: TypeScript best practices with Vitest, Biome, strict typing, and npm/pnpm package management. Use when writing or reviewing TypeScript code in project workflows.
allowed-tools:
  - Read
  - Bash
---

# TypeScript Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand when language keywords are detected |
| Trigger cues | TypeScript code discussions, framework guidance, or file extensions such as .ts/.tsx. |
| Tier | 3 |

## What it does

Provides TypeScript-specific expertise for TDD development, including Vitest testing, Biome linting/formatting, strict type checking, and modern npm/pnpm package management.

## When to use

- Engages when the conversation references TypeScript work, frameworks, or files like .ts/.tsx.
- "Writing TypeScript tests", "How to use Vitest", "Type safety"
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

## Examples
```bash
npm run lint && npm test
```

## Inputs
- Language-specific source directories (e.g. `src/`, `app/`).
- Language-specific build/test configuration files (e.g. `package.json`, `pyproject.toml`, `go.mod`).
- Relevant test suites and sample data.

## Outputs
- Test/lint execution plan tailored to the selected language.
- List of key language idioms and review checkpoints.

## Failure Modes
- When the language runtime or package manager is not installed.
- When the main language cannot be determined in a multilingual project.

## Dependencies
- Access to the project file is required using the Read/Grep tool.
- When used with `Skill("moai-foundation-langs")`, it is easy to share cross-language conventions.

## References
- Microsoft. "TypeScript Handbook." https://www.typescriptlang.org/docs/ (accessed 2025-03-29).
- OpenJS Foundation. "ESLint User Guide." https://eslint.org/docs/latest/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: Input/output/failure response/reference information for each language has been specified.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (TypeScript-specific review)
- alfred-refactoring-coach (type-safe refactoring)

## Best Practices
- Enable automatic validation by matching your linter with the language's official style guide.
- Fix test/build pipelines with reproducible commands in CI.
