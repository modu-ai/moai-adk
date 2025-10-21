---

name: moai-lang-kotlin
description: Kotlin best practices with JUnit, Gradle, ktlint, coroutines, and extension functions. Use when writing or reviewing Kotlin code in project workflows.
allowed-tools:
  - Read
  - Bash
---

# Kotlin Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand when language keywords are detected |
| Trigger cues | Kotlin code discussions, framework guidance, or file extensions such as .kt/.kts. |
| Tier | 3 |

## What it does

Provides Kotlin-specific expertise for TDD development, including JUnit testing, Gradle build system, ktlint linting, coroutines for concurrency, and extension functions.

## When to use

- Engages when the conversation references Kotlin work, frameworks, or files like .kt/.kts.
- “Writing Kotlin tests”, “How to use coroutines”, “Android patterns”
- Automatically invoked when working with Kotlin/Android projects
- Kotlin SPEC implementation (`/alfred:2-run`)

## How it works

**TDD Framework**:
- **JUnit 5**: Unit testing with Kotlin extensions
- **MockK**: Kotlin-friendly mocking library
- **Kotest**: Kotlin-native testing framework
- Test coverage ≥85% with JaCoCo

**Build Tools**:
- **Gradle**: build.gradle.kts with Kotlin DSL
- **Maven**: pom.xml alternative
- Multi-platform support (JVM, Native, JS)

**Code Quality**:
- **ktlint**: Kotlin linter with formatting
- **detekt**: Static code analysis
- **Android Lint**: Android-specific checks

**Kotlin Features**:
- **Coroutines**: Async programming with suspend functions
- **Extension functions**: Add methods to existing classes
- **Data classes**: Automatic equals/hashCode/toString
- **Null safety**: Non-nullable types by default
- **Smart casts**: Automatic type casting after checks

**Android Patterns**:
- **Jetpack Compose**: Declarative UI
- **ViewModel**: UI state management
- **Room**: Database abstraction
- **Retrofit**: Network requests

## Examples
```bash
./gradlew test && ./gradlew ktlintCheck
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
- JetBrains. "Kotlin Language Documentation." https://kotlinlang.org/docs/home.html (accessed 2025-03-29).
- Pinterest. "ktlint." https://pinterest.github.io/ktlint/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: Input/output/failure response/reference information for each language has been specified.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Kotlin-specific review)
- mobile-app-expert (Android app development)

## Best Practices
- Enable automatic validation by matching your linter with the language's official style guide.
- Fix test/build pipelines with reproducible commands in CI.
