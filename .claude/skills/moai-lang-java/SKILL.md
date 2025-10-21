---
name: moai-lang-java
description: Java best practices with JUnit, Maven/Gradle, Checkstyle, and Spring Boot patterns
allowed-tools:
  - Read
  - Bash
tier: 3
auto-load: "true"
---

# Java Expert

## What it does

Provides Java-specific expertise for TDD development, including JUnit testing, Maven/Gradle build tools, Checkstyle linting, and Spring Boot patterns.

## When to use

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

### Example 1: TDD with JUnit
User: "/alfred:2-run SERVICE-001"
Claude: (creates RED test with JUnit 5, GREEN implementation, REFACTOR with interfaces)

### Example 2: Build execution
User: "Run Maven build"
Claude: (runs mvn clean test and reports results)

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Java-specific review)
- database-expert (JPA/Hibernate patterns)
