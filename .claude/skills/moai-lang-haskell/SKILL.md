---

name: moai-lang-haskell
description: Haskell best practices with HUnit, Stack/Cabal, and pure functional programming. Use when writing or reviewing Haskell code in project workflows.
allowed-tools:
  - Read
  - Bash
---

# Haskell Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand when language keywords are detected |
| Trigger cues | Haskell code discussions, framework guidance, or file extensions such as .hs. |
| Tier | 3 |

## What it does

Provides Haskell-specific expertise for TDD development, including HUnit testing, Stack/Cabal build tools, and pure functional programming with strong type safety.

## When to use

- Engages when the conversation references Haskell work, frameworks, or files like .hs.
- “Writing Haskell tests”, “How to use HUnit”, “Pure functional programming”
- Automatically invoked when working with Haskell projects
- Haskell SPEC implementation (`/alfred:2-run`)

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
```bash
cabal test && hlint src
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
- Haskell.org. "Haskell Language Documentation." https://www.haskell.org/documentation/ (accessed 2025-03-29).
- GitHub. "HLint." https://github.com/ndmitchell/hlint (accessed 2025-03-29).

## Changelog
- 2025-03-29: Input/output/failure response/reference information for each language has been specified.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Haskell-specific review)
- alfred-refactoring-coach (functional refactoring)

## Best Practices
- Enable automatic validation by matching your linter with the language's official style guide.
- Fix test/build pipelines with reproducible commands in CI.
