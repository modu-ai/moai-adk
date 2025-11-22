---
name: moai-lang-kotlin
description: Kotlin 2.0+ Multiplatform Development with KMP, Coroutines, and Compose
---

## Quick Reference (30 seconds)

# Kotlin Multiplatform Development — Enterprise

**Primary Focus**: Kotlin 2.0+ with KMP architecture, async patterns, and Compose UI
**Best For**: Multiplatform development (mobile, backend, web), coroutines, reactive programming
**Key Libraries**: Kotlin 2.0.20, Coroutines 1.8.0, Compose Multiplatform 1.6.10, Ktor 2.3.12
**Auto-triggers**: Kotlin, KMP, coroutines, multiplatform, Compose

| Version | Release | Support |
|---------|---------|---------|
| Kotlin 2.0.20 | 2025-11 | Active |
| Coroutines 1.8.0 | 2024-11 | ✅ |
| Compose MP 1.6.10 | 2024-11 | ✅ |
| Ktor 2.3.12 | 2024-11 | ✅ |

---

## What It Does

Kotlin 2.0+ multiplatform development with structured concurrency, type-safe async patterns, and cross-platform UI. Enterprise-ready patterns for mobile, backend, and web applications.

**Key capabilities**:
- ✅ Kotlin Multiplatform (KMP) architecture with expect/actual
- ✅ Structured concurrency and coroutines (Flow, StateFlow, async/await)
- ✅ Compose Multiplatform UI (Android, iOS, Web)
- ✅ Type-safe null handling and sealed classes
- ✅ Reactive streams and reactive programming
- ✅ Enterprise dependency injection and architecture patterns

---

## When to Use

**Automatic triggers**:
- Kotlin/KMP project files (*.kt, build.gradle.kts)
- Multiplatform architecture discussions
- Coroutine and async/await patterns
- Compose UI development

**Manual invocation**:
- Design multiplatform architecture
- Implement async patterns and error handling
- Optimize performance for mobile/backend
- Review enterprise Kotlin code

---

## Three-Level Learning Path

### Level 1: Fundamentals (See examples.md)

Core Kotlin 2.0 concepts with practical patterns:
- **Null Safety**: Type system, safe navigation, Elvis operator
- **Coroutines**: async/await, launch, scope, structured concurrency
- **Collections & Sequences**: Lazy evaluation, functional programming
- **Kotlin Basics**: Data classes, sealed classes, extension functions
- **Testing**: runTest, MockK, unit and integration tests

### Level 2: Advanced Patterns (See modules/advanced-patterns.md)

Production-ready enterprise patterns:
- **Reactive Streams**: Flow, StateFlow, SharedFlow, cold vs hot
- **KMP Architecture**: expect/actual, multiplatform project structure
- **Error Handling**: Result wrapper, sealed hierarchies, Railway-oriented programming
- **Dependency Injection**: Koin, service locator patterns
- **Advanced DSL**: Builder pattern, context receivers, scope functions

### Level 3: Performance & Deployment (See modules/optimization.md)

Production optimization and multiplatform deployment:
- **Memory Optimization**: Inline classes, lazy sequences, object pooling
- **Execution Speed**: Tail recursion, coroutine pooling, efficient collections
- **Multiplatform Performance**: Platform-specific optimizations, expect/actual best practices
- **Testing at Scale**: Test containers, performance benchmarking
- **CI/CD Integration**: Build pipeline optimization, multiplatform publishing

---

## Best Practices

✅ **DO**:
- Use structured concurrency (coroutineScope, supervisorScope)
- Leverage null safety over exceptions
- Use Sequence for lazy evaluation on large datasets
- Implement proper resource management with `use` blocks
- Write comprehensive tests with runTest and MockK
- Use Data classes for domain models
- Prefer sealed classes over open hierarchies

❌ **DON'T**:
- Use GlobalScope (always use structured scopes)
- Ignore null safety warnings
- Mix blocking calls in coroutines
- Use bare exceptions without context
- Skip error handling in async operations
- Create unmanaged coroutines
- Use !! operator without validation

---

## Tool Versions (2025-11-22)

| Tool | Version | Purpose |
|------|---------|---------|
| **Kotlin** | 2.0.20 | Language |
| **Coroutines** | 1.8.0 | Async |
| **Compose MP** | 1.6.10 | UI |
| **Ktor** | 2.3.12 | HTTP |
| **Serialization** | 1.7.1 | JSON |
| **Android Gradle** | 8.5.0 | Build |

---

## Installation & Setup

```bash
# Gradle with Kotlin
plugins {
    kotlin("multiplatform") version "2.0.20"
    id("com.android.library") version "8.5.0"
}

# Add dependencies
kotlin {
    sourceSets {
        commonMain.dependencies {
            implementation("org.jetbrains.kotlinx:kotlinx-coroutines-core:1.8.0")
            implementation("org.jetbrains.compose.runtime:runtime:1.6.10")
            implementation("io.ktor:ktor-client-core:2.3.12")
        }
    }
}
```

---

## Works Well With

- `moai-domain-backend` (Kotlin server development)
- `moai-essentials-debug` (Debugging coroutines)
- `moai-foundation-testing` (Test strategies)
- `moai-foundation-trust` (TRUST 5 quality gates)

---

## Learn More

- **Practical Examples**: See `examples.md` for 20+ real-world patterns
- **Advanced Patterns**: See `modules/advanced-patterns.md` for reactive, KMP, DSL patterns
- **Performance Tuning**: See `modules/optimization.md` for memory, multiplatform optimization
- **Official Docs**: https://kotlinlang.org/
- **Coroutines**: https://kotlinlang.org/docs/coroutines-overview.html
- **Compose**: https://www.jetbrains.com/help/compose-multiplatform/

---

## Changelog

- **v4.0.0** (2025-11-22): Modularized structure with advanced patterns and optimization modules
- **v3.5.0** (2025-11-13): Refactored to progressive disclosure with comprehensive examples
- **v3.0.0** (2025-03-15): Added KMP and multiplatform patterns
- **v2.0.0** (2025-01-10): Basic Kotlin patterns and best practices

---

## Context7 Integration

### Related Libraries & Tools
- [Kotlin](/jetbrains/kotlin): Modern JVM language with multiplatform support
- [Kotlin Coroutines](/kotlin/kotlinx.coroutines): Structured concurrency and async patterns
- [Compose Multiplatform](/jetbrains/compose-multiplatform): Cross-platform UI framework
- [Ktor](/ktorio/ktor): Asynchronous web framework
- [Serialization](/kotlin/kotlinx.serialization): JSON and data serialization

### Official Documentation
- [Kotlin Documentation](https://kotlinlang.org/docs/)
- [Coroutines Guide](https://kotlinlang.org/docs/coroutines-overview.html)
- [KMP Guide](https://www.jetbrains.com/help/kotlin-multiplatform-dev/)
- [Compose Docs](https://www.jetbrains.com/help/compose-multiplatform/)

### Version-Specific Guides
Latest stable version: Kotlin 2.0.20, Coroutines 1.8.0, Compose MP 1.6.10
- [Kotlin 2.0 Release Notes](https://kotlinlang.org/docs/whatsnew20.html)
- [Coroutines API](https://kotlin.github.io/kotlinx.coroutines/)
- [KMP Setup](https://www.jetbrains.com/help/kotlin-multiplatform-dev/get-started.html)

---

**Skills**: Skill("moai-essentials-debug"), Skill("moai-foundation-testing"), Skill("moai-domain-backend")
**Auto-loads**: Kotlin projects with multiplatform, coroutines, Compose patterns

