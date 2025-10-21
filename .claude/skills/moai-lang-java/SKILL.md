---

name: moai-lang-java
description: Java best practices with JUnit, Maven/Gradle, Checkstyle, and Spring Boot patterns. Use when writing or reviewing Java code in project workflows.
allowed-tools:
  - Read
  - Bash
---

# Java Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand when language keywords are detected |
| Trigger cues | Java code discussions, framework guidance, or file extensions such as .java. |
| Tier | 3 |

## What it does

Provides Java-specific expertise for TDD development, including JUnit testing, Maven/Gradle build tools, Checkstyle linting, and Spring Boot patterns.

## When to use

- Engages when the conversation references Java work, frameworks, or files like .java.
- “Writing Java tests”, “How to use JUnit”, “Spring Boot patterns”
- Automatically invoked when working with Java projects
- Java SPEC implementation (`/alfred:2-run`)

## How it works

**TDD Framework**:
- **JUnit 5**: Unit testing with annotations (@Test, @BeforeEach)
- **Mockito**: Mocking framework for dependencies
- **AssertJ**: Fluent assertion library
- Test coverage ≥85% with JaCoCo

**Build Tools**:
- **Maven**: pom.xml, dependency management
- **Gradle**: build.gradle, Kotlin DSL support
- Multi-module project structures

**Code Quality**:
- **Checkstyle**: Java style checker (Google/Sun conventions)
- **PMD**: Static code analysis
- **SpotBugs**: Bug detection

**Spring Boot Patterns**:
- Dependency Injection (@Autowired, @Component)
- REST controllers (@RestController, @RequestMapping)
- Service layer separation (@Service, @Repository)
- Configuration properties (@ConfigurationProperties)

**Best Practices**:
- File ≤300 LOC, method ≤50 LOC
- Interfaces for abstraction
- Builder pattern for complex objects
- Exception handling with custom exceptions

## Examples
```bash
./mvnw test && ./mvnw checkstyle:check
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
- Oracle. "Java Language Specification." https://docs.oracle.com/javase/specs/ (accessed 2025-03-29).
- JUnit. "JUnit 5 User Guide." https://junit.org/junit5/docs/current/user-guide/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: Input/output/failure response/reference information for each language has been specified.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Java-specific review)
- database-expert (JPA/Hibernate patterns)

## Best Practices
- Enable automatic validation by matching your linter with the language's official style guide.
- Fix test/build pipelines with reproducible commands in CI.
