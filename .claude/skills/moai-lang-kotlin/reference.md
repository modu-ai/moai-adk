# moai-lang-kotlin - CLI Reference

_Last updated: 2025-10-22_

## Tool Versions (2025-10-22)

| Tool | Version | Official Link |
|------|---------|---------------|
| **Kotlin** | 2.1.0 | https://kotlinlang.org/docs/releases.html |
| **JUnit** | 5.11.0 | https://junit.org/junit5/ |
| **Gradle** | 8.12.0 | https://gradle.org/releases/ |
| **ktlint** | 1.5.0 | https://github.com/pinterest/ktlint |

## Quick Reference

### Project Initialization

```bash
# Initialize new Kotlin project with Gradle
gradle init --type kotlin-application --dsl kotlin

# Or with Maven
mvn archetype:generate \
  -DarchetypeGroupId=org.jetbrains.kotlin \
  -DarchetypeArtifactId=kotlin-archetype-jvm \
  -DarchetypeVersion=2.1.0
```

### Common Commands

```bash
# Build project
./gradlew build

# Run tests
./gradlew test

# Run tests with coverage
./gradlew test jacocoTestReport

# Run specific test
./gradlew test --tests "com.example.UserServiceTest"

# Run tests continuously (watch mode)
./gradlew test --continuous

# Check code style
./gradlew ktlintCheck

# Auto-fix code style
./gradlew ktlintFormat

# Run application
./gradlew run

# Clean build artifacts
./gradlew clean

# Show dependencies
./gradlew dependencies

# Check for dependency updates
./gradlew dependencyUpdates
```

### Testing Commands

```bash
# Run all tests
./gradlew test

# Run with coverage
./gradlew test jacocoTestReport

# Verify coverage threshold (≥85%)
./gradlew jacocoTestCoverageVerification

# Run tests in parallel
./gradlew test --parallel --max-workers=4

# Run only unit tests
./gradlew test --tests "*Test"

# Run only integration tests
./gradlew test --tests "*IntegrationTest"

# Debug tests
./gradlew test --debug-jvm
```

### Code Quality Commands

```bash
# Run ktlint
./gradlew ktlintCheck

# Auto-fix formatting
./gradlew ktlintFormat

# Run detekt (static analysis)
./gradlew detekt

# Generate detekt report
./gradlew detekt --build-cache
```

### Dependency Management

```bash
# Add dependency to build.gradle.kts
dependencies {
    implementation("group:artifact:version")
    testImplementation("group:artifact:version")
}

# Refresh dependencies
./gradlew build --refresh-dependencies

# Show dependency tree
./gradlew dependencies

# Check for updates
./gradlew dependencyUpdates
```

---

## Build Configuration (build.gradle.kts)

### Basic Setup

```kotlin
plugins {
    kotlin("jvm") version "2.1.0"
    application
    jacoco
    id("org.jlleitschuh.gradle.ktlint") version "12.1.1"
}

group = "com.example"
version = "1.0.0"

repositories {
    mavenCentral()
}

dependencies {
    // Kotlin standard library
    implementation(kotlin("stdlib"))

    // Testing
    testImplementation(kotlin("test"))
    testImplementation("org.junit.jupiter:junit-jupiter:5.11.0")
    testImplementation("io.mockk:mockk:1.13.13")
}

tasks.test {
    useJUnitPlatform()
}

application {
    mainClass.set("com.example.MainKt")
}
```

### Advanced Configuration

```kotlin
plugins {
    kotlin("jvm") version "2.1.0"
    kotlin("plugin.serialization") version "2.1.0"
    application
    jacoco
    id("org.jlleitschuh.gradle.ktlint") version "12.1.1"
    id("io.gitlab.arturbosch.detekt") version "1.23.7"
}

kotlin {
    jvmToolchain(21)

    compilerOptions {
        freeCompilerArgs.add("-Xjsr305=strict")
        allWarningsAsErrors.set(true)
    }
}

dependencies {
    // Kotlin
    implementation(kotlin("stdlib"))
    implementation(kotlin("reflect"))

    // Coroutines
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-core:1.9.0")
    testImplementation("org.jetbrains.kotlinx:kotlinx-coroutines-test:1.9.0")

    // Serialization
    implementation("org.jetbrains.kotlinx:kotlinx-serialization-json:1.7.3")

    // Testing
    testImplementation(kotlin("test"))
    testImplementation("org.junit.jupiter:junit-jupiter:5.11.0")
    testImplementation("io.mockk:mockk:1.13.13")
    testImplementation("org.assertj:assertj-core:3.26.3")
}

tasks.test {
    useJUnitPlatform()
    finalizedBy(tasks.jacocoTestReport)

    testLogging {
        events("passed", "skipped", "failed")
        exceptionFormat = org.gradle.api.tasks.testing.logging.TestExceptionFormat.FULL
        showStandardStreams = false
    }
}

tasks.jacocoTestReport {
    dependsOn(tasks.test)

    reports {
        xml.required.set(true)
        html.required.set(true)
        csv.required.set(false)
    }

    classDirectories.setFrom(
        files(classDirectories.files.map {
            fileTree(it) {
                exclude(
                    "**/config/**",
                    "**/dto/**",
                    "**/entity/**",
                    "**/*Application*"
                )
            }
        })
    )
}

tasks.jacocoTestCoverageVerification {
    dependsOn(tasks.test)

    violationRules {
        rule {
            limit {
                counter = "LINE"
                value = "COVEREDRATIO"
                minimum = "0.85".toBigDecimal()
            }
        }

        rule {
            limit {
                counter = "BRANCH"
                value = "COVEREDRATIO"
                minimum = "0.80".toBigDecimal()
            }
        }
    }
}

tasks.check {
    dependsOn(tasks.jacocoTestCoverageVerification)
}

ktlint {
    version.set("1.5.0")
    verbose.set(true)
    android.set(false)
    outputToConsole.set(true)
    ignoreFailures.set(false)

    filter {
        exclude("**/generated/**")
        include("**/kotlin/**")
    }
}

detekt {
    buildUponDefaultConfig = true
    allRules = false
    config.setFrom("$projectDir/config/detekt.yml")

    reports {
        html.required.set(true)
        xml.required.set(true)
        txt.required.set(false)
    }
}
```

---

## Testing Patterns

### JUnit 5 Test Structure

```kotlin
import org.junit.jupiter.api.*
import kotlin.test.assertEquals
import kotlin.test.assertTrue
import kotlin.test.assertFalse
import kotlin.test.assertNotNull
import kotlin.test.assertNull

class CalculatorTest {

    private lateinit var calculator: Calculator

    @BeforeEach
    fun setUp() {
        calculator = Calculator()
    }

    @AfterEach
    fun tearDown() {
        // Cleanup
    }

    @Test
    fun `should add two numbers`() {
        val result = calculator.add(2, 3)
        assertEquals(5, result)
    }

    @Test
    fun `should handle negative numbers`() {
        val result = calculator.add(-1, -2)
        assertEquals(-3, result)
    }

    @Test
    @Disabled("Not implemented yet")
    fun `should handle complex operations`() {
        // TODO
    }
}
```

### Parameterized Tests

```kotlin
import org.junit.jupiter.params.ParameterizedTest
import org.junit.jupiter.params.provider.ValueSource
import org.junit.jupiter.params.provider.CsvSource
import org.junit.jupiter.params.provider.MethodSource

class ParameterizedTests {

    @ParameterizedTest
    @ValueSource(ints = [1, 2, 3, 4, 5])
    fun `should be positive`(number: Int) {
        assertTrue(number > 0)
    }

    @ParameterizedTest
    @CsvSource(
        "1, 2, 3",
        "2, 3, 5",
        "5, 5, 10"
    )
    fun `should add correctly`(a: Int, b: Int, expected: Int) {
        assertEquals(expected, calculator.add(a, b))
    }

    @ParameterizedTest
    @MethodSource("provideTestData")
    fun `should validate email`(email: String, isValid: Boolean) {
        assertEquals(isValid, validator.isValidEmail(email))
    }

    companion object {
        @JvmStatic
        fun provideTestData() = listOf(
            Arguments.of("test@example.com", true),
            Arguments.of("invalid-email", false),
            Arguments.of("user@domain.co", true)
        )
    }
}
```

### Coroutine Testing

```kotlin
import kotlinx.coroutines.test.*
import org.junit.jupiter.api.Test
import kotlin.test.assertEquals

class CoroutineTests {

    @Test
    fun `should fetch data async`() = runTest {
        val service = ApiService()
        val result = service.fetchData("user123")

        assertEquals("Data for user123", result)
    }

    @Test
    fun `should handle delays`() = runTest {
        val start = currentTime
        val service = ApiService()

        service.delayedOperation()

        val elapsed = currentTime - start
        assertTrue(elapsed >= 1000)
    }

    @Test
    fun `should test concurrent operations`() = runTest {
        val service = ApiService()
        val results = service.fetchMultiple(listOf("1", "2", "3"))

        assertEquals(3, results.size)
    }
}
```

### Mocking with MockK

```kotlin
import io.mockk.*
import org.junit.jupiter.api.Test
import kotlin.test.assertEquals

class MockingTests {

    @Test
    fun `should mock repository`() {
        val repository = mockk<UserRepository>()
        val service = UserService(repository)

        every { repository.findById("123") } returns User("123", "test@example.com")

        val result = service.getUser("123")

        assertEquals("123", result.id)
        verify(exactly = 1) { repository.findById("123") }
    }

    @Test
    fun `should mock suspend functions`() = runTest {
        val repository = mockk<UserRepository>()

        coEvery { repository.save(any()) } returns User("123", "test@example.com")

        val result = repository.save(User("123", "test@example.com"))

        assertEquals("123", result.id)
        coVerify(exactly = 1) { repository.save(any()) }
    }

    @Test
    fun `should spy on real object`() {
        val calculator = spyk<Calculator>()

        every { calculator.add(any(), any()) } returns 100

        assertEquals(100, calculator.add(1, 2))
        verify { calculator.add(1, 2) }
    }
}
```

---

## Kotlin Best Practices

### Null Safety

```kotlin
// Use nullable types explicitly
var name: String? = null

// Safe call operator
val length = name?.length

// Elvis operator
val length = name?.length ?: 0

// Safe cast
val stringValue = value as? String

// Not-null assertion (use sparingly)
val length = name!!.length

// Let function for null checks
name?.let {
    println("Name is $it")
}
```

### Extension Functions

```kotlin
// String extensions
fun String.isValidEmail(): Boolean {
    return matches(Regex("^[A-Za-z0-9+_.-]+@[A-Za-z0-9.-]+\\.[A-Z|a-z]{2,}$"))
}

// Usage
if (email.isValidEmail()) {
    // Process email
}

// Collection extensions
fun <T> List<T>.secondOrNull(): T? = if (size >= 2) this[1] else null

// Number extensions
fun Int.isEven(): Boolean = this % 2 == 0
```

### Sealed Classes

```kotlin
sealed class Result<out T> {
    data class Success<T>(val data: T) : Result<T>()
    data class Error(val message: String) : Result<Nothing>()
    data object Loading : Result<Nothing>()
}

// Exhaustive when expression
fun <T> handleResult(result: Result<T>) = when (result) {
    is Result.Success -> println("Success: ${result.data}")
    is Result.Error -> println("Error: ${result.message}")
    is Result.Loading -> println("Loading...")
}
```

### Data Classes

```kotlin
// Auto-generates equals(), hashCode(), toString(), copy()
data class User(
    val id: String,
    val email: String,
    val name: String
)

// Copy with modifications
val updatedUser = user.copy(name = "New Name")

// Destructuring
val (id, email, name) = user
```

---

## TRUST 5 Checklist

### T - Test Coverage (≥85%)

```bash
./gradlew test jacocoTestReport
# View: build/reports/jacoco/test/html/index.html
```

### R - Readable Code

```bash
./gradlew ktlintCheck
./gradlew ktlintFormat
./gradlew detekt
```

### U - Unified Types

- Use Kotlin's type system
- Leverage sealed classes
- Use data classes for immutability
- Avoid any and explicit casts

### S - Security

```bash
# Dependency vulnerability scanning
./gradlew dependencyCheckAnalyze

# Check: build/reports/dependency-check-report.html
```

### T - Trackable with @TAG

```bash
# Find all TAG references
rg '@(CODE|TEST|SPEC):' -n src/ --type kotlin

# Verify TAG coverage
rg '@CODE:' src/main/kotlin/ --count
rg '@TEST:' src/test/kotlin/ --count
```

---

## Official Resources

- **Kotlin Docs**: https://kotlinlang.org/docs/home.html
- **Kotlin Style Guide**: https://kotlinlang.org/docs/coding-conventions.html
- **JUnit 5 User Guide**: https://junit.org/junit5/docs/current/user-guide/
- **Gradle Kotlin DSL**: https://docs.gradle.org/current/userguide/kotlin_dsl.html
- **ktlint**: https://pinterest.github.io/ktlint/
- **MockK**: https://mockk.io/
- **Coroutines Guide**: https://kotlinlang.org/docs/coroutines-guide.html

---

_For working examples, see [examples.md](examples.md)_
