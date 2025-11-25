# moai-lang-kotlin - CLI Reference

_Last updated: 2025-11-13_

## Quick Reference

### Installation

#### Using Gradle (Kotlin/JVM)

```bash
# Add to build.gradle.kts
dependencies {
    implementation("org.jetbrains.kotlin:kotlin-stdlib:2.0.20")
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-core:1.8.0")
}
```

#### Using Gradle (KMP)

```bash
# build.gradle.kts
kotlin {
    jvm()
    iosArm64()
    iosSimulatorArm64()
    js(IR) { browser() }

    sourceSets {
        commonMain.dependencies {
            implementation("org.jetbrains.kotlin:kotlin-stdlib:2.0.20")
            implementation("org.jetbrains.kotlinx:kotlinx-coroutines-core:1.8.0")
        }
    }
}
```

#### Using Maven

```xml
<dependency>
    <groupId>org.jetbrains.kotlin</groupId>
    <artifactId>kotlin-stdlib</artifactId>
    <version>2.0.20</version>
</dependency>
```

### Common Commands

#### Build and Run

```bash
# Compile Kotlin code
kotlinc HelloWorld.kt -include-runtime -d hello.jar

# Run Kotlin code
kotlin HelloWorld.kt

# Run with Gradle
gradle build
gradle run

# Run tests
gradle test
```

#### Testing

```bash
# Run all tests
gradle test

# Run specific test class
gradle test --tests UserRepositoryTest

# Run with coverage
gradle test jacocoTestReport

# Check coverage report
open build/reports/jacoco/test/html/index.html
```

#### Linting and Formatting

```bash
# Install ktlint
brew install ktlint
# or
gradle ktlintFormat

# Run ktlint
ktlint

# Auto-fix lint issues
ktlint -F

# Check specific file
ktlint src/main/kotlin/Main.kt
```

#### Type Checking

```bash
# Run Kotlin compiler (type check only)
kotlinc -no-jdk -no-stdlib HelloWorld.kt -d /tmp

# With Gradle
gradle compileKotlin
```

#### Performance Profiling

```bash
# Run with profiler (JVM)
gradle run --args="--profile"

# Memory profiling
kotlin -J-Xmx2g HelloWorld.kt

# CPU profiling
gradle run --profile
```

---

## Tool Versions (2025-11-13)

| Tool | Version | Purpose | Status |
|------|---------|---------|--------|
| **Kotlin** | 2.0.20 | Core language | Current |
| **JDK** | 17-21 | Java Runtime | Stable |
| **Gradle** | 8.5+ | Build system | Current |
| **ktlint** | 1.5.0 | Linting | Current |
| **Kotest** | 5.8.0 | Testing | Current |
| **MockK** | 1.13.8 | Mocking | Current |
| **Kotlinx-coroutines** | 1.8.0 | Async | Current |
| **Compose Multiplatform** | 1.6.10 | UI | Current |
| **Android Gradle Plugin** | 8.5.0 | Android build | Current |

---

## Kotlin Version Compatibility Matrix (2025-11-13)

| Kotlin Version | JVM Target | KMP | Compose | Status |
|---|---|---|---|---|
| **2.0.20** | JVM 8+ | Yes | 1.6.10 | Current |
| **2.0.0** | JVM 8+ | Yes | 1.6.0 | LTS |
| **1.9.24** | JVM 8+ | Yes | 1.5.x | Maintenance |
| **1.9.0** | JVM 8+ | Yes | 1.5.0 | End of life |

---

## Android API Compatibility

```kotlin
// Android SDK versions
minSdk = 24         // Android 7.0
targetSdk = 34      // Android 14.0
compileSdk = 34

// Kotlin on Android
implementation("org.jetbrains.kotlin:kotlin-stdlib-jdk8:2.0.20")
implementation("androidx.appcompat:appcompat:1.7.0")
implementation("androidx.lifecycle:lifecycle-runtime:2.7.0")
```

---

## Core Libraries Reference

### Kotlinx-Coroutines

```bash
# Build.gradle.kts
dependencies {
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-core:1.8.0")
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-android:1.8.0")
    testImplementation("org.jetbrains.kotlinx:kotlinx-coroutines-test:1.8.0")
}
```

**Key Functions**:
- `launch()` - Fire-and-forget coroutine
- `async()` - Coroutine with result
- `withContext()` - Suspend with context switching
- `runBlocking()` - Bridge sync/async
- `supervisorScope()` - Error isolation
- `Flow<T>` - Reactive streams
- `StateFlow<T>` - State management

### Kotlinx-Serialization

```bash
# Build.gradle.kts
plugins {
    kotlin("plugin.serialization") version "2.0.20"
}

dependencies {
    implementation("org.jetbrains.kotlinx:kotlinx-serialization-json:1.7.1")
}
```

**Key Features**:
- JSON serialization
- Protocol buffers
- CBOR format
- Polymorphic types

### Arrow-kt

```bash
# Build.gradle.kts
dependencies {
    implementation("io.arrow-kt:arrow-core:1.2.0")
    implementation("io.arrow-kt:arrow-fx-coroutines:1.2.0")
}
```

**Key Types**:
- `Either<E, A>` - Error handling
- `Option<A>` - Nullable values
- `Effect<A>` - Computation builder

---

## Syntax Reference

### Type Declarations

```kotlin
// Immutable property
val name: String = "John"

// Mutable property
var age: Int = 30

// Type inference
val derived = "Hello"  // Inferred as String

// Nullable types
val nullable: String? = null
val notNull: String = "Required"

// Function type
val fn: (Int) -> String = { num -> num.toString() }
```

### Functions

```kotlin
// Regular function
fun greet(name: String): String = "Hello, $name"

// Suspend function (coroutine)
suspend fun fetchData(): String = withContext(Dispatchers.IO) { /* ... */ }

// Extension function
fun String.duplicate() = this + this

// Infix function
infix fun Int.times(other: Int) = this * other

// Operator overloading
operator fun String.plus(other: String) = "$this$other"

// Default parameters
fun connect(host: String, port: Int = 8080) { /* ... */ }

// Varargs
fun sum(vararg numbers: Int) = numbers.sum()
```

### Control Flow

```kotlin
// When expression (like switch, but better)
val result = when (x) {
    1 -> "one"
    2 -> "two"
    else -> "other"
}

// If expression
val status = if (age >= 18) "adult" else "minor"

// Try-catch expression
val value = try {
    Integer.parseInt("123")
} catch (e: NumberFormatException) {
    0
}

// For loop
for (i in 1..10) { /* ... */ }
for (item in list) { /* ... */ }

// While loop
while (count < 10) { /* ... */ }

// Do-while loop
do {
    println(count)
    count++
} while (count < 10)
```

### Collections

```kotlin
// List
val list = listOf(1, 2, 3)
val mutable = mutableListOf(1, 2, 3)

// Set
val set = setOf(1, 2, 3)

// Map
val map = mapOf("key" to "value")

// Pair
val pair: Pair<String, Int> = "name" to 25

// Sequence (lazy)
val seq = (1..1000).asSequence().filter { it % 2 == 0 }
```

### Lambda Expressions

```kotlin
// Basic lambda
val add = { a: Int, b: Int -> a + b }

// With receiver
val format: String.(String) -> String = { suffix -> "$this-$suffix" }

// It parameter
listOf(1, 2, 3).map { it * 2 }

// Trailing lambda
list.fold(0) { acc, num -> acc + num }

// Destructuring
pairs.forEach { (key, value) -> println("$key: $value") }
```

### Classes and Objects

```kotlin
// Data class
data class User(val id: String, val name: String, val age: Int)

// Sealed class
sealed class Result<T> {
    data class Success<T>(val data: T) : Result<T>()
    data class Error(val exception: Throwable) : Result<Nothing>()
}

// Object (singleton)
object Config {
    const val API_URL = "https://api.example.com"
}

// Interface
interface Repository {
    suspend fun fetch(id: String): Result<String>
}

// Enum
enum class Status {
    PENDING, ACTIVE, COMPLETED
}

// Value class (inline)
@JvmInline
value class UserId(val value: String)
```

---

## Best Practices

### 1. Null Safety

```kotlin
// GOOD: Use safe calls and let
val length = name?.length ?: 0

// GOOD: Use null coalescing
val username = user?.name ?: "Anonymous"

// AVOID: Not-null assertion (!)
val length = name!!.length  // Can throw

// GOOD: Smart casting
if (obj is String) {
    println(obj.length)  // obj is String here
}
```

### 2. Coroutine Safety

```kotlin
// GOOD: Use structured concurrency
withContext(Dispatchers.IO) {
    val result = fetchData()
}

// AVOID: GlobalScope
GlobalScope.launch { /* ... */ }  // Do not use

// GOOD: Launch in coroutineScope
coroutineScope {
    val job = launch { /* ... */ }
}
```

### 3. Resource Management

```kotlin
// GOOD: Use use() for resources
File("data.txt").bufferedReader().use { reader ->
    reader.readLines()
}

// GOOD: Try-finally for cleanup
try {
    // Use resource
} finally {
    // Clean up
}
```

### 4. Error Handling

```kotlin
// GOOD: Result wrapper
suspend fun fetchData(): Result<String> = try {
    Result.success(apiCall())
} catch (e: Exception) {
    Result.failure(e)
}

// GOOD: Arrow Either
import arrow.core.Either

suspend fun fetchData(): Either<ApiError, String> {
    // Implementation
}
```

### 5. Testing Best Practices

```kotlin
// GOOD: Use runTest for coroutines
@Test
fun testAsync() = runTest {
    val result = someAsyncFunction()
    assertEquals("expected", result)
}

// GOOD: Mock with MockK
@Test
fun testWithMock() {
    val repository = mockk<Repository>()
    coEvery { repository.fetch("1") } returns "data"
}
```

---

## Performance Optimization

### Memory Efficiency

1. **Use Sequence instead of List for long chains**:
   ```kotlin
   (1..1000000).asSequence()
       .filter { it % 2 == 0 }
       .map { it * 2 }
       .toList()
   ```

2. **Use inline classes for zero-overhead wrappers**:
   ```kotlin
   @JvmInline
   value class UserId(val value: String)
   ```

3. **Avoid boxing with primitive types**:
   ```kotlin
   IntArray(1000)  // More efficient than Array<Int>
   ```

### Execution Speed

1. **Use tailrec for recursion**:
   ```kotlin
   tailrec fun factorial(n: Int, acc: Int = 1): Int =
       if (n <= 1) acc else factorial(n - 1, n * acc)
   ```

2. **Avoid unnecessary allocations**:
   ```kotlin
   // Create once, reuse many times
   val regex = Regex("pattern")
   ```

---

## Troubleshooting

### Common Issues

**Issue**: `java.lang.NoClassDefFoundError`
- **Cause**: Missing dependency or classpath issue
- **Solution**: Verify gradle dependencies, rebuild project

**Issue**: `SuspendFunction0 cannot be applied`
- **Cause**: Calling suspend function without suspend context
- **Solution**: Wrap in `runBlocking` or `GlobalScope.launch`

**Issue**: `Unresolved reference` in IDE but builds fine
- **Cause**: IDE cache issue
- **Solution**: Invalidate caches and restart IDE

**Issue**: Memory leaks with coroutines
- **Cause**: Not cancelling coroutines properly
- **Solution**: Use structured concurrency, proper scope management

---

## Context7 Integration

### Available Documentation

1. **Kotlin Language**: `/kotlin/kotlin`
2. **Coroutines**: `/kotlin/kotlinx.coroutines`
3. **KMP**: `/kotlin/kotlin.multiplatform`
4. **Compose**: `/jetbrains/compose-multiplatform`
5. **Ktor**: `/ktor/ktor`
6. **Serialization**: `/kotlin/kotlinx.serialization`

### Usage Example

```kotlin
// Get latest coroutine docs from Context7
val docs = mcp__context7__get-library-docs(
    context7CompatibleLibraryID = "/kotlin/kotlinx.coroutines"
)
```

---

## Resources

- **Official Docs**: https://kotlinlang.org
- **Coroutines Guide**: https://kotlinlang.org/docs/coroutines-overview.html
- **KMP Documentation**: https://kotlinlang.org/docs/multiplatform.html
- **Compose Multiplatform**: https://www.jetbrains.com/lp/compose-multiplatform/
- **Kotlinx Libraries**: https://github.com/Kotlin/kotlinx

---

## Changelog

- ** .0** (2025-11-13): Reference documentation update
- **v3.0.0** (2025-03-15): Multiplatform patterns
- **v2.0.0** (2025-01-10): Basic patterns
- **v1.0.0** (2024-12-01): Initial release
