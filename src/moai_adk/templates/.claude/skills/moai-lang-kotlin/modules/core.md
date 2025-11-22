// Chain operations safely
fetchData()
    .onSuccess { data -> updateUI(data) }
    .onFailure { error -> showError(error) }
```

### Testing

```kotlin
@Test
fun testAsync() = runTest {
    val result = someAsyncFunction()
    assertEquals("expected", result)
}

@Test
fun testWithMock() {
    val repo = mockk<Repository>()
    coEvery { repo.fetch("1") } returns "data"
    // Verify behavior
    coVerify { repo.fetch("1") }
}
```


## Context7 MCP Integration

This skill integrates with Context7 for real-time access to official documentation:

### Available Resources

1. **Kotlin Language**: `/kotlin/kotlin`
2. **Coroutines**: `/kotlin/kotlinx.coroutines`
3. **KMP**: `/kotlin/kotlin.multiplatform`
4. **Compose**: `/jetbrains/compose-multiplatform`
5. **Ktor**: `/ktor/ktor`
6. **Serialization**: `/kotlin/kotlinx.serialization`

### Usage Example

```kotlin
// Get latest coroutine patterns
val docs = mcp__context7__get-library-docs(
    context7CompatibleLibraryID = "/kotlin/kotlinx.coroutines"
)
```


## Performance Optimization

### Memory Efficiency

1. **Use Sequence for lazy evaluation**:
   ```kotlin
   (1..1_000_000).asSequence()
       .filter { it % 2 == 0 }
       .map { it * 2 }
       .toList()  // Only processes what's needed
   ```

2. **Inline classes for zero-overhead**:
   ```kotlin
   @JvmInline
   value class UserId(val value: String)  // No allocation at runtime
   ```

3. **Primitive arrays instead of boxed**:
   ```kotlin
   val intArray = IntArray(1000)  // More efficient than Array<Int>
   ```

### Execution Speed

1. **Tail recursion**:
   ```kotlin
   tailrec fun factorial(n: Int, acc: Int = 1): Int =
       if (n <= 1) acc else factorial(n - 1, n * acc)
   ```

2. **Coroutine pooling** (automatic with structured concurrency)


## Best Practices

### 1. Always Use Structured Concurrency

```kotlin
// GOOD
coroutineScope {
    val result = async { fetchData() }
}

// AVOID
GlobalScope.launch { fetchData() }  // Never use
```

### 2. Null Safety Over Exceptions

```kotlin
// GOOD
val user = repository.findUser(id)
    ?.let { updateUI(it) }
    ?: showNotFound()

// AVOID
val user = repository.findUser(id)!!  // Unsafe
```

### 3. Resource Management

```kotlin
// GOOD
File("data.txt").bufferedReader().use { reader ->
    reader.readLines()
}  // Auto-closes

// GOOD
try {
    // Use resource
} finally {
    resource.close()
}
```


## Security Considerations

### Input Validation

```kotlin
data class SecureUserInput(val email: String) {
    init {
        require(email.contains("@")) { "Invalid email" }
        require(email.length <= 254) { "Email too long" }
    }
}
```

### Secure Storage

```kotlin
expect class SecureStorage {
    suspend fun store(key: String, value: String)
    suspend fun retrieve(key: String): String?
}
```

### Network Security

```kotlin
// Certificate pinning
val client = HttpClient {
    install(Auth) {
        bearer {
            loadTokens { getBearerTokens() }
        }
    }
}
```


## Testing Strategy

| Category | Target | Tools |
|----------|--------|-------|
| **Unit Tests** | 80% | Kotest, MockK, runTest |
| **Integration Tests** | 15% | Kotest, testcontainers |
| **UI Tests** | 5% | Compose Test |


## Works Well With

- `moai-foundation-trust` (TRUST 5 quality gates)
- `moai-foundation-security` (Enterprise security)
- `moai-foundation-testing` (Testing strategies)
- `moai-cc-mcp-integration` (MCP integration)
- `moai-essentials-debug` (Debugging)



## Advanced Patterns

## Skill Overview

**Kotlin 2.0** Multiplatform enterprise development with advanced async patterns, KMP architecture, and Compose Multiplatform UI. This skill provides patterns for mobile, backend, and cross-platform development with full Context7 MCP integration for real-time documentation access.

### Core Capabilities

- ✅ Kotlin 2.0 Multiplatform (KMP) enterprise architecture
- ✅ Advanced coroutines and structured concurrency patterns
- ✅ Compose Multiplatform UI development (Android, iOS, Web)
- ✅ Enterprise testing strategies with Kotest and MockK
- ✅ Context7 MCP integration for latest documentation
- ✅ Performance optimization and memory management
- ✅ Modern architecture patterns (MVI, Clean Architecture)
- ✅ Security best practices for production systems


## Advanced Async Patterns

### Flow - Reactive Streams

```kotlin
// Create cold flow (lazy)
fun getUsersFlow(): Flow<User> = flow {
    while (true) {
        val users = repository.fetchUsers()
        emit(users)
        delay(5000)  // Refresh every 5 seconds
    }
}

// Compose flows
getUsersFlow()
    .map { it.copy(name = it.name.uppercase()) }
    .filter { it.isActive }
    .distinctUntilChanged()
    .collect { updateUI(it) }
```

### StateFlow - Mutable State

```kotlin
class CounterViewModel {
    private val _count = MutableStateFlow(0)
    val count: StateFlow<Int> = _count.asStateFlow()

    fun increment() { _count.value++ }
    fun decrement() { _count.value-- }
}

// Observe state changes
viewModel.count.collect { count ->
    updateUI("Count: $count")
}
```





## Context7 Integration

### Related Libraries & Tools
- [Kotlin](/jetbrains/kotlin): Modern JVM language
- [Ktor](/ktorio/ktor): Asynchronous web framework
- [Exposed](/jetbrains/exposed): SQL framework

### Official Documentation
- [Documentation](https://kotlinlang.org/docs/)
- [API Reference](https://kotlinlang.org/api/latest/jvm/stdlib/)

### Version-Specific Guides
Latest stable version: 1.9
- [Release Notes](https://github.com/JetBrains/kotlin/releases)
- [Migration Guide](https://kotlinlang.org/docs/whatsnew19.html)
