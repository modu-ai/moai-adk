---
paths:
  - "**/*.kt"
  - "**/*.kts"
---

# Kotlin Rules

Version: Kotlin 2.0+

## Tooling

- Build: Gradle with Kotlin DSL
- Testing: Kotest or JUnit 5
- Linting: ktlint, detekt

## Best Practices (2026)

- Use structured concurrency - child tasks auto-cancel with parent
- Never hardcode Dispatchers - inject them for testability
- Never use GlobalScope - makes testing difficult
- Trigger coroutines from ViewModel, not Views
- Use data classes for DTOs, sealed classes for state

## Ktor 3.4.0 Features

- Structured concurrency: client disconnect cancels request handling
- Zstd compression support
- OpenAPI generation
- Duplex streaming for OkHttp

## Coroutine Patterns

```kotlin
// Inject dispatchers
class UserRepository(
    private val ioDispatcher: CoroutineDispatcher = Dispatchers.IO
) {
    suspend fun fetchUser(id: String) = withContext(ioDispatcher) {
        api.getUser(id)
    }
}

// Concurrent requests with Ktor client
val responses = coroutineScope {
    urls.map { url ->
        async { client.get(url) }
    }.awaitAll()
}
```

## Android Patterns

- Use Jetpack Compose for UI
- Use Hilt for dependency injection
- Use Flow for reactive streams

## MoAI Integration

- Use Skill("moai-lang-kotlin") for detailed patterns
- Follow TRUST 5 quality gates
