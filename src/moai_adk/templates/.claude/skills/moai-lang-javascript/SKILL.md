---

name: moai-lang-javascript
description: JavaScript best practices with Jest, ESLint, Prettier, and npm package management. Use when writing or reviewing JavaScript code in project workflows.
allowed-tools:
  - Read
  - Bash
---

# JavaScript Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand when language keywords are detected |
| Trigger cues | JavaScript code discussions, framework guidance, or file extensions such as .js. |
| Tier | 3 |

## What it does

Provides JavaScript-specific expertise for TDD development, including Jest testing, ESLint linting, Prettier formatting, and npm package management.

## When to use

- Engages when the conversation references JavaScript work, frameworks, or files like .js.
- "Writing JavaScript tests", "How to use Jest", "ES6+ grammar"
- Automatically invoked when working with JavaScript projects
- JavaScript SPEC implementation (`/alfred:2-run`)

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
```bash
npm run test && npm run lint
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
- MDN Web Docs. "JavaScript Guide." https://developer.mozilla.org/en-US/docs/Web/JavaScript/Guide (accessed 2025-03-29).
- Jest. "Getting Started." https://jestjs.io/docs/getting-started (accessed 2025-03-29).

## Changelog
- 2025-03-29: Input/output/failure response/reference information for each language has been specified.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (JavaScript-specific review)
- alfred-debugger-pro (JavaScript debugging)

## Best Practices
- Enable automatic validation by matching your linter with the language's official style guide.
- Fix test/build pipelines with reproducible commands in CI.
