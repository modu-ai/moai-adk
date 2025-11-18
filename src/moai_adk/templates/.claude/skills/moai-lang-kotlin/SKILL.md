---
name: moai-lang-kotlin
version: 4.0.0
updated: '2025-11-19'
status: stable
stability: stable
description: Kotlin 2.0 Multiplatform Enterprise Development with KMP, Coroutines,
allowed-tools:
- Read
- Bash
- WebSearch
- WebFetch
---



# Lang Kotlin Skill - Enterprise 

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

---

## Quick Reference

### When to Use This Skill

**Automatic activation**:
- Kotlin/KMP development discussions
- Multiplatform project architecture and design
- Mobile development with shared business logic
- Async programming and coroutine patterns
- Compose Multiplatform UI development

**Manual invocation**:
- Design multiplatform architecture
- Implement advanced async patterns
- Optimize performance and memory usage
- Review enterprise Kotlin code

---

## Technology Stack (2025-11-13)

| Component | Version | Purpose | Status |
|-----------|---------|---------|--------|
| **Kotlin** | 2.0.20 | Core language | Current |
| **Coroutines** | 1.8.0 | Async programming | Current |
| **Compose Multiplatform** | 1.6.10 | UI framework | Current |
| **Serialization** | 1.7.1 | JSON/data serialization | Current |
| **Ktor** | 2.3.12 | HTTP client/server | Current |
| **Android Gradle Plugin** | 8.5.0 | Android build | Current |

---

## Core Language Features

### 1. Null Safety & Type System

Kotlin's null safety is a game-changer for enterprise development:

```kotlin
// Smart null handling
val name: String? = "John"
val length = name?.length ?: 0  // Safe navigation + Elvis operator

// Type-safe IDs with inline classes
@JvmInline
value class UserId(val value: String)

// Sealed hierarchies for domain modeling
sealed class Result<T> {
    data class Success<T>(val data: T) : Result<T>()
    data class Error(val exception: Throwable) : Result<Nothing>()
}
```

**Key Benefits**:
- Eliminates `NullPointerException` at compile time
- Zero runtime overhead with inline classes
- Type-safe domain models

### 2. Coroutines - Structured Concurrency

Enterprise async patterns with built-in safety:

```kotlin
// Structured concurrency scope
coroutineScope {
    val result1 = async { fetchData("API1") }
    val result2 = async { fetchData("API2") }
    awaitAll(result1, result2)
} // All coroutines cancelled if scope exits

// Context switching
withContext(Dispatchers.IO) {
    val data = blockingIoCall()  // Safe switching
}

// Error isolation with supervisor scope
supervisorScope {
    launch { riskyOperation1() }  // Failure doesn't affect others
    launch { riskyOperation2() }
}
```

**Key Benefits**:
- Automatic cancellation and resource cleanup
- Prevents memory leaks and zombie coroutines
- Context-aware execution

### 3. Extension Functions & DSLs

Powerful language extension without inheritance:

```kotlin
// Domain-specific languages (DSLs)
fun html(block: HtmlBuilder.() -> Unit): String {
    val builder = HtmlBuilder()
    builder.block()
    return builder.build()
}

// Usage
val page = html {
    h1("Welcome")
    p("This is a paragraph")
}

// See examples.md for complete DSL patterns
```

---

## Multiplatform Architecture

### Project Structure

```
kmp-enterprise-app/
├── shared/
│   ├── src/
│   │   ├── commonMain/kotlin/
│   │   │   ├── domain/        # Business logic
│   │   │   ├── data/          # Data layer
│   │   │   └── presentation/  # State management
│   │   ├── androidMain/kotlin/
│   │   ├── iosMain/kotlin/
│   │   └── commonTest/kotlin/
├── androidApp/
├── iosApp/
└── webApp/
```

### expect/actual Pattern

```kotlin
// commonMain: Define interface
expect class PlatformDatabase {
    suspend fun saveData(data: String): Result<Unit>
}

// androidMain: Android implementation
actual class PlatformDatabase {
    actual suspend fun saveData(data: String) = try {
        room.insert(data)
        Result.success(Unit)
    } catch (e: Exception) {
        Result.failure(e)
    }
}

// iosMain: iOS implementation
actual class PlatformDatabase {
    actual suspend fun saveData(data: String) = try {
        coreData.save(data)
        Result.success(Unit)
    } catch (e: Exception) {
        Result.failure(e)
    }
}
```

**Key Pattern**:
- Share business logic in `commonMain`
- Isolate platform specifics in platform modules
- Compile-time platform selection

---

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

---

## Compose Multiplatform UI

### Basic Components

```kotlin
@Composable
fun UserListScreen(users: List<User>) {
    Column(modifier = Modifier.fillMaxSize()) {
        Text("Users", style = MaterialTheme.typography.headlineLarge)

        LazyColumn {
            items(users) { user ->
                UserCard(user)
            }
        }
    }
}

@Composable
fun UserCard(user: User) {
    Card(modifier = Modifier.fillMaxWidth().padding(8.dp)) {
        Column(modifier = Modifier.padding(16.dp)) {
            Text(user.name, style = MaterialTheme.typography.titleMedium)
            Text(user.email, color = Color.Gray)
        }
    }
}
```

**Key Concepts**:
- Declarative UI
- Reusable components
- State management integration

---

## Enterprise Patterns

### Dependency Injection with Koin

```kotlin
val appModule = module {
    single { HttpClient() }
    single { UserRepository(get()) }
    viewModel { UserListViewModel(get()) }
}

// Usage
val userService: UserService = get()  // Injected automatically
```

### Error Handling

```kotlin
// Result wrapper for safe operations
suspend fun fetchData(): Result<String> = try {
    Result.success(apiCall())
} catch (e: Exception) {
    Result.failure(e)
}

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

---

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

---

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

---

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

---

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

---

## Testing Strategy

| Category | Target | Tools |
|----------|--------|-------|
| **Unit Tests** | 80% | Kotest, MockK, runTest |
| **Integration Tests** | 15% | Kotest, testcontainers |
| **UI Tests** | 5% | Compose Test |

---

## Works Well With

- `moai-foundation-trust` (TRUST 5 quality gates)
- `moai-foundation-security` (Enterprise security)
- `moai-foundation-testing` (Testing strategies)
- `moai-cc-mcp-integration` (MCP integration)
- `moai-essentials-debug` (Debugging)

---

## For Complete Information

- **Working Examples**: See `examples.md` (29 examples covering all patterns)
- **CLI Reference**: See `reference.md` (installation, commands, troubleshooting)
- **Official Docs**: https://kotlinlang.org
- **Coroutines Guide**: https://kotlinlang.org/docs/coroutines-overview.html

---

## Changelog

- ** .0** (2025-11-13): Refactored to Progressive Disclosure with comprehensive examples.md and reference.md
- **v3.0.0** (2025-03-15): Added KMP and multiplatform patterns
- **v2.0.0** (2025-01-10): Basic Kotlin patterns and best practices
- **v1.0.0** (2024-12-01): Initial release

---

_Last updated: 2025-11-18_
