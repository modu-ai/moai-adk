---

name: moai-lang-scala
description: Scala best practices with ScalaTest, sbt, and functional programming patterns. Use when writing or reviewing Scala code in project workflows.
allowed-tools:
  - Read
  - Bash
---

# Scala Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand when language keywords are detected |
| Trigger cues | Scala code discussions, framework guidance, or file extensions such as .scala. |
| Tier | 3 |

## What it does

Provides Scala-specific expertise for TDD development, including ScalaTest framework, sbt build tool, and functional programming patterns.

## When to use

- Engages when the conversation references Scala work, frameworks, or files like .scala.
- “Writing Scala tests”, “How to use ScalaTest”, “Functional programming”
- Automatically invoked when working with Scala projects
- Scala SPEC implementation (`/alfred:2-run`)

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
```bash
sbt test && sbt scalafmtCheck
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
- Lightbend. "Scala Documentation." https://docs.scala-lang.org/ (accessed 2025-03-29).
- Scalameta. "scalafmt." https://scalameta.org/scalafmt/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: Input/output/failure response/reference information for each language has been specified.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Scala-specific review)
- alfred-refactoring-coach (functional refactoring)

## Best Practices
- Enable automatic validation by matching your linter with the language's official style guide.
- Fix test/build pipelines with reproducible commands in CI.
