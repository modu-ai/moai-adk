---

name: moai-lang-clojure
description: Clojure best practices with clojure.test, Leiningen, and immutable data structures. Use when writing or reviewing Clojure code in project workflows.
allowed-tools:
  - Read
  - Bash
---

# Clojure Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand when language keywords are detected |
| Trigger cues | Clojure code discussions, framework guidance, or file extensions such as .clj/.cljc. |
| Tier | 3 |

## What it does

Provides Clojure-specific expertise for TDD development, including clojure.test framework, Leiningen build tool, and immutable data structures with functional programming.

## When to use

- Engages when the conversation references Clojure work, frameworks, or files like .clj/.cljc.
- "Writing Clojure tests", "How to use clojure.test", "Immutable data structures"
- Automatically invoked when working with Clojure projects
- Clojure SPEC implementation (`/alfred:2-run`)

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
```bash
lein test && clj-kondo --lint src
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
- Clojure.org. "Clojure Documentation." https://clojure.org/guides/getting_started (accessed 2025-03-29).
- clj-kondo. "User Guide." https://github.com/clj-kondo/clj-kondo/blob/master/doc/usage.md (accessed 2025-03-29).

## Changelog
- 2025-03-29: Input/output/failure response/reference information for each language has been specified.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Clojure-specific review)
- alfred-refactoring-coach (functional refactoring)

## Best Practices
- Enable automatic validation by matching your linter with the language's official style guide.
- Fix test/build pipelines with reproducible commands in CI.
