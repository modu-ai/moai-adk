---
paths:
  - "**/*.kt"
  - "**/*.kts"
  - "**/build.gradle.kts"
---

# Kotlin Rules

Version: Kotlin 2.0+

## Tooling

- Build: Gradle with Kotlin DSL
- Linting: ktlint or detekt
- Testing: JUnit 5, MockK
- Coverage: Kover >= 85%

## MUST

- Use data classes for DTOs and value objects
- Use sealed classes for restricted hierarchies
- Use coroutines for async operations
- Use extension functions for utilities
- Prefer immutability (val over var)
- Use null safety features (?., ?:, !!)

## MUST NOT

- Use !! without prior null check
- Use lateinit for nullable types
- Suppress warnings without justification
- Use Java-style getters/setters
- Block the main thread with runBlocking in production
- Use mutable collections in public APIs

## File Conventions

- *Test.kt for test files
- Multiple classes allowed per file if related
- Use PascalCase for classes
- Use camelCase for functions and properties
- Use SCREAMING_CASE for constants

## Testing

- Use JUnit 5 with Kotlin extensions
- Use MockK for mocking (Kotlin-native)
- Use Kotest for property-based testing
- Use Testcontainers for integration tests

## MoAI Integration

- Use Skill("moai-lang-kotlin") for detailed patterns
- Follow TRUST 5 quality gates
