---
name: moai-lang-kotlin
description: Kotlin best practices with JUnit, Gradle, ktlint, coroutines, and extension
  functions
allowed-tools:
- Read
- Bash
---

# Kotlin Expert

## What it does

Provides Kotlin-specific expertise for TDD development, including JUnit testing, Gradle build system, ktlint linting, coroutines for concurrency, and extension functions.

## When to use

- "Kotlin 테스트 작성", "코루틴 사용법", "Android 패턴"
- Automatically invoked when working with Kotlin/Android projects
- Kotlin SPEC implementation (`/alfred:2-build`)

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

### Example 1: TDD with coroutines
User: "/alfred:2-build API-001"
Claude: (creates RED test with runTest, GREEN implementation with suspend functions, REFACTOR)

### Example 2: ktlint check
User: "ktlint 실행"
Claude: (runs ktlint and reports style violations)

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Kotlin-specific review)
- mobile-app-expert (Android app development)
