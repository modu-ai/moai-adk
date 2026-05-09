# Kotlin 2.0+ Reference Guide

## Complete Language Coverage

### Kotlin 2.0

Version Information:
- Latest: 2.0.20 (November 2025)
- K2 Compiler: 2x faster compilation, better type inference
- Multiplatform: JVM, JS, Native, WASM targets
- Minimum JVM: Java 8 (Java 21 recommended)

Core Features:

K2 Compiler Features:
- Stable - New compiler frontend with 2x performance improvement
- Better type inference and error messages
- Improved smart casts and control flow analysis
- Enhanced kapt alternative: KSP 2.0

Context Receivers:
- Experimental feature for multiple implicit receivers
- Enable with `-Xcontext-receivers` compiler flag
- Useful for dependency injection patterns

Data Objects:
- Stable singleton data classes
- Automatically generated `toString()` returns object name
- No `copy()` or `componentN()` methods

Value Classes:
- Stable inline wrapper types (formerly inline classes)
- Zero runtime overhead for primitive wrappers
- Use `@JvmInline` annotation

Sealed Interfaces:
- Stable restricted interface implementations
- Enable exhaustive `when` expressions
- Support hierarchical type modeling

Explicit API Mode:
- Stable mode for library development
- Enforces explicit visibility modifiers
- Enable with `explicitApi()` in Gradle

---

## Ecosystem Libraries

### Coroutines (kotlinx.coroutines 1.9)

Core Components:
- `CoroutineScope` - Structured concurrency boundary
- `Job` - Cancellable background operation
- `Deferred` - Future with result value
- `Flow` - Cold asynchronous stream
- `StateFlow` / `SharedFlow` - Hot observable streams

Dispatchers:
- `Dispatchers.Default` - CPU-intensive work
- `Dispatchers.IO` - Blocking I/O operations
- `Dispatchers.Main` - UI thread (Android/Desktop)
- `Dispatchers.Unconfined` - Immediate execution

Virtual Thread Integration (JVM 21+):
```kotlin
val vtDispatcher = Executors.newVirtualThreadPerTaskExecutor()
    .asCoroutineDispatcher()

withContext(vtDispatcher) {
    // Runs on virtual thread
}
```

### Ktor 3.0

Server Features:
- Async HTTP with Netty, CIO, Jetty engines
- WebSocket support with routing
- JWT, OAuth, Session authentication
- Content negotiation with kotlinx.serialization
- Request validation and error handling
- OpenAPI/Swagger integration

Client Features:
- Multiplatform HTTP client
- Connection pooling and retry
- Serialization support
- WebSocket client
- Logging and monitoring

Plugins Ecosystem:
- `ContentNegotiation` - JSON/XML serialization
- `Authentication` - JWT, OAuth, Basic, Form
- `StatusPages` - Exception handling
- `CallLogging` - Request/response logging
- `CORS` - Cross-origin resource sharing
- `Compression` - Gzip, Deflate
- `RateLimit` - Request throttling

### Exposed 0.55

Features:
- Type-safe SQL DSL
- DAO pattern support
- Transaction management
- Connection pooling integration
- Coroutines support via `exposed-kotlin-datetime`

Supported Databases:
- PostgreSQL, MySQL, MariaDB
- SQLite, H2, Oracle
- SQL Server

Table Types:
- `Table` - Base table definition
- `IdTable<T>` - Table with typed ID column
- `IntIdTable` - Table with Int ID
- `LongIdTable` - Table with Long ID
- `UUIDTable` - Table with UUID ID

### Arrow 2.0

Core Modules:
- `arrow-core` - Either, Option, Validated
- `arrow-fx-coroutines` - Effectful programming
- `arrow-optics` - Lens, Prism, Traversal
- `arrow-resilience` - Retry, Circuit breaker

Key Types:
- `Either<A, B>` - Disjoint union type
- `Option<A>` - Nullable alternative
- `Validated<E, A>` - Accumulating errors
- `Ior<A, B>` - Inclusive or

### Compose Multiplatform 1.7

Supported Platforms:
- Android (native Compose)
- Desktop (JVM with Skia)
- iOS (Kotlin/Native with Skia)
- Web (Kotlin/JS or WASM)

UI Components:
- Material 3 design system
- Custom component creation
- Animation framework
- Gesture handling
- Navigation library

State Management:
- `remember` - Composition-scoped state
- `mutableStateOf` - Observable state
- `derivedStateOf` - Computed state
- `collectAsState` - Flow to state conversion

---

## Context7 Library Mappings

Kotlin Core:
```
/JetBrains/kotlin - Kotlin language reference
/Kotlin/kotlinx.coroutines - Coroutines library
/Kotlin/kotlinx.serialization - Serialization framework
```

Server Frameworks:
```
/ktorio/ktor - Ktor server and client
/JetBrains/Exposed - SQL framework
/koin/koin - Dependency injection
/insert-koin/koin - Alternative Koin path
```

Functional Programming:
```
/arrow-kt/arrow - Arrow FP library
/cashapp/sqldelight - Type-safe SQL
/google/ksp - Kotlin Symbol Processing
```

Multiplatform:
```
/jetbrains/compose-multiplatform - Compose UI
/touchlab/SKIE - Swift-Kotlin interop
/AAkira/Napier - Multiplatform logging
```

Android:
```
/android/architecture-components - Jetpack
/coil-kt/coil - Image loading
/square/okhttp - HTTP client
/square/retrofit - REST client
```

Testing:
```
/mockk/mockk - Mocking library
/kotest/kotest - Testing framework
/app.cash/turbine - Flow testing
```

---

## Testing Patterns

### MockK with Coroutines

```kotlin
class UserServiceTest {
    private val repository = mockk<UserRepository>()
    private val service = UserService(repository)

    @Test
    fun `should create user`() = runTest {
        // Arrange
        val request = CreateUserRequest("John", "john@example.com", "password")
        val expectedUser = User(1L, "John", "john@example.com")

        coEvery { repository.existsByEmail(any()) } returns false
        coEvery { repository.save(any()) } returns expectedUser

        // Act
        val result = service.create(request)

        // Assert
        assertThat(result).isEqualTo(expectedUser)
        coVerify(exactly = 1) { repository.save(any()) }
    }

    @Test
    fun `should throw on duplicate email`() = runTest {
        coEvery { repository.existsByEmail("existing@example.com") } returns true

        assertThrows<DuplicateEmailException> {
            service.create(CreateUserRequest("John", "existing@example.com", "pass"))
        }

        coVerify(exactly = 0) { repository.save(any()) }
    }
}
```

### Flow Testing with Turbine

```kotlin
@Test
fun `should emit user updates`() = runTest {
    val users = listOf(
        User(1L, "John", "john@example.com"),
        User(2L, "Jane", "jane@example.com")
    )

    service.observeUsers().test {
        assertThat(awaitItem()).isEqualTo(users[0])
        assertThat(awaitItem()).isEqualTo(users[1])
        awaitComplete()
    }
}

@Test
fun `should handle errors in flow`() = runTest {
    coEvery { repository.findAllAsFlow() } throws RuntimeException("DB error")

    service.streamUsers().test {
        val error = awaitError()
        assertThat(error).isInstanceOf(RuntimeException::class.java)
        assertThat(error.message).contains("DB error")
    }
}
```

### Kotest Specification Style

```kotlin
class UserServiceSpec : FunSpec({
    val repository = mockk<UserRepository>()
    val service = UserService(repository)

    beforeTest {
        clearAllMocks()
    }

    context("create user") {
        test("should create user successfully") {
            coEvery { repository.existsByEmail(any()) } returns false
            coEvery { repository.save(any()) } returns User(1L, "John", "john@example.com")

            val result = service.create(CreateUserRequest("John", "john@example.com", "pass"))

            result.name shouldBe "John"
            result.email shouldBe "john@example.com"
        }

        test("should reject duplicate email") {
            coEvery { repository.existsByEmail("taken@example.com") } returns true

            shouldThrow<DuplicateEmailException> {
                service.create(CreateUserRequest("John", "taken@example.com", "pass"))
            }
        }
    }
})
```

### Ktor Test Host

```kotlin
class UserRoutesTest {
    @Test
    fun `GET users returns list`() = testApplication {
        val mockService = mockk<UserService>()
        coEvery { mockService.findAll(any(), any()) } returns listOf(
            User(1L, "John", "john@example.com")
        )

        application {
            install(ContentNegotiation) { json() }
            routing {
                route("/api/users") {
                    get { call.respond(mockService.findAll(0, 20)) }
                }
            }
        }

        client.get("/api/users").apply {
            assertEquals(HttpStatusCode.OK, status)
            val users = body<List<UserDto>>()
            assertEquals(1, users.size)
            assertEquals("John", users[0].name)
        }
    }
}
```

---

## Performance Characteristics

### Compilation Times

Build Performance (Kotlin 2.0 with K2):
- Clean build: 30-60% faster than Kotlin 1.9
- Incremental build: 10-20% faster
- KSP processing: 2-3x faster than kapt

Gradle Configuration:
```kotlin
// gradle.properties
kotlin.experimental.tryK2=true
kotlin.incremental=true
kotlin.daemon.jvmargs=-Xmx4g
org.gradle.parallel=true
org.gradle.caching=true
```

### Runtime Performance

JVM Performance:
- Inline functions: Zero overhead
- Value classes: Zero runtime allocation
- Coroutines: ~100 bytes per suspended coroutine
- Flow: Minimal allocation per emission

Ktor Performance:
- Throughput: ~200K requests/sec (Netty)
- Latency P99: ~1ms
- Memory: ~256MB base heap

### Memory Usage

Coroutine Memory:
- Continuation: ~100-200 bytes
- Job: ~300 bytes
- Channel: ~500 bytes per buffer slot

---

## Development Environment

### IDE Support

IntelliJ IDEA (Recommended):
- Full K2 compiler support
- Advanced refactoring tools
- Debugging with coroutine visualization
- Profiling integration

Android Studio:
- Based on IntelliJ IDEA
- Android-specific tooling
- Compose preview support
- Layout inspector

VS Code:
- Kotlin extension available
- Basic syntax highlighting
- Limited refactoring support

### Recommended Plugins

IntelliJ IDEA:
- Kotlin (bundled)
- Ktor (official)
- Exposed (database tooling)
- Detekt (static analysis)
- ktlint (formatting)

### Linting and Formatting

Detekt Configuration:
```yaml
# detekt.yml
build:
  maxIssues: 10
  excludeCorrectable: false

style:
  MaxLineLength:
    maxLineLength: 120
  MagicNumber:
    ignoreNumbers: ['-1', '0', '1', '2']
  UnusedPrivateMember:
    active: true

complexity:
  ComplexMethod:
    threshold: 15
  LongParameterList:
    functionThreshold: 6
```

ktlint Configuration:
```properties
# .editorconfig
[*.kt]
indent_style = space
indent_size = 4
max_line_length = 120
ktlint_code_style = ktlint_official
```

---

## Container Optimization

### Docker Multi-Stage Build

```dockerfile
FROM gradle:8.5-jdk21 AS builder
WORKDIR /app
COPY build.gradle.kts settings.gradle.kts ./
COPY gradle ./gradle
RUN gradle dependencies --no-daemon

COPY src ./src
RUN gradle shadowJar --no-daemon

FROM eclipse-temurin:21-jre-alpine
RUN addgroup -g 1000 app && adduser -u 1000 -G app -s /bin/sh -D app
WORKDIR /app
COPY --from=builder /app/build/libs/*-all.jar app.jar
USER app
EXPOSE 8080
ENTRYPOINT ["java", "-jar", "app.jar"]
```

### GraalVM Native Image

```dockerfile
FROM ghcr.io/graalvm/native-image-community:21 AS builder
WORKDIR /app
COPY . .
RUN ./gradlew nativeCompile

FROM gcr.io/distroless/base-debian12
COPY --from=builder /app/build/native/nativeCompile/app /app
ENTRYPOINT ["/app"]
```

### JVM Tuning for Containers

```yaml
containers:
  - name: kotlin-app
    image: myapp:latest
    resources:
      requests:
        memory: "256Mi"
        cpu: "250m"
      limits:
        memory: "512Mi"
        cpu: "500m"
    env:
      - name: JAVA_OPTS
        value: >-
          -XX:+UseContainerSupport
          -XX:MaxRAMPercentage=75.0
          -XX:+UseG1GC
          -XX:+UseStringDeduplication
```

---

## Migration Guides

### Kotlin 1.9 to 2.0

Enable K2 Compiler:
```properties
# gradle.properties
kotlin.experimental.tryK2=true
```

Key Changes:
- Improved smart casts in complex conditions
- Better type inference for builders
- Enhanced SAM conversion
- New K2 compiler diagnostics

### Java to Kotlin

Data Classes:
```kotlin
// Java
public class User {
    private final String name;
    private final String email;
    // constructor, getters, equals, hashCode, toString...
}

// Kotlin
data class User(val name: String, val email: String)
```

Null Safety:
```kotlin
// Java nullable parameter
fun process(value: String?) {
    value?.let { println(it) } ?: println("null value")
}

// Platform types from Java
val javaResult: String! = javaMethod() // Platform type
val safeResult: String = javaResult ?: "default" // Make safe
```

Extension Functions:
```kotlin
// Replace utility classes
fun String.toTitleCase(): String =
    split(" ").joinToString(" ") { it.capitalize() }

fun <T> List<T>.secondOrNull(): T? = getOrNull(1)
```

---

Last Updated: 2025-12-07
Version: 1.0.0
