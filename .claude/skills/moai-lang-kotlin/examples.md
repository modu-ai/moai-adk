# moai-lang-kotlin - Working Examples

_Last updated: 2025-10-22_

## Example 1: Project Setup with Gradle & JUnit 5

```bash
# Initialize Kotlin project with Gradle
gradle init --type kotlin-application --dsl kotlin

# Project structure created:
# app/
# ├── build.gradle.kts
# ├── src/
# │   ├── main/kotlin/
# │   └── test/kotlin/
# ├── settings.gradle.kts
# └── gradle.properties
```

**build.gradle.kts configuration**:
```kotlin
plugins {
    kotlin("jvm") version "2.1.0"
    application
    id("org.jlleitschuh.gradle.ktlint") version "12.1.1"
}

repositories {
    mavenCentral()
}

dependencies {
    // Kotlin standard library
    implementation(kotlin("stdlib"))

    // Coroutines
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-core:1.9.0")

    // Testing
    testImplementation(kotlin("test"))
    testImplementation("org.junit.jupiter:junit-jupiter:5.11.0")
    testImplementation("io.mockk:mockk:1.13.13")
    testImplementation("org.jetbrains.kotlinx:kotlinx-coroutines-test:1.9.0")
}

tasks.test {
    useJUnitPlatform()

    // Coverage configuration
    finalizedBy(tasks.jacocoTestReport)
}

tasks.jacocoTestReport {
    dependsOn(tasks.test)

    reports {
        xml.required.set(true)
        html.required.set(true)
    }

    classDirectories.setFrom(
        files(classDirectories.files.map {
            fileTree(it) {
                exclude("**/config/**", "**/dto/**")
            }
        })
    )
}

ktlint {
    version.set("1.5.0")
    verbose.set(true)
    android.set(false)
}

application {
    mainClass.set("com.example.AppKt")
}
```

## Example 2: TDD Workflow with JUnit 5

**RED: Write failing test**
```kotlin
// src/test/kotlin/com/example/domain/UserServiceTest.kt
package com.example.domain

import io.mockk.every
import io.mockk.mockk
import io.mockk.verify
import kotlinx.coroutines.test.runTest
import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows
import kotlin.test.assertEquals
import kotlin.test.assertNotNull

class UserServiceTest {

    @Test
    fun `should create user with valid data`() = runTest {
        // @TEST:USER-001
        val repository = mockk<UserRepository>()
        val service = UserService(repository)

        val userId = "user123"
        val email = "test@example.com"

        every { repository.save(any()) } returns User(userId, email)

        val result = service.createUser(email)

        assertNotNull(result)
        assertEquals(userId, result.id)
        assertEquals(email, result.email)
        verify(exactly = 1) { repository.save(any()) }
    }

    @Test
    fun `should reject invalid email`() = runTest {
        val repository = mockk<UserRepository>()
        val service = UserService(repository)

        val exception = assertThrows<IllegalArgumentException> {
            service.createUser("invalid-email")
        }

        assertEquals("Invalid email format", exception.message)
    }

    @Test
    fun `should handle duplicate email`() = runTest {
        val repository = mockk<UserRepository>()
        val service = UserService(repository)

        every { repository.findByEmail(any()) } returns User("existing", "test@example.com")

        val exception = assertThrows<DuplicateUserException> {
            service.createUser("test@example.com")
        }

        assertEquals("User with email test@example.com already exists", exception.message)
    }
}
```

**GREEN: Implement feature**
```kotlin
// src/main/kotlin/com/example/domain/UserService.kt
package com.example.domain

/**
 * @CODE:USER-001 | SPEC: SPEC-USER-001.md | TEST: UserServiceTest.kt
 * User management service with email validation
 */
class UserService(private val repository: UserRepository) {

    suspend fun createUser(email: String): User {
        validateEmail(email)

        repository.findByEmail(email)?.let {
            throw DuplicateUserException("User with email $email already exists")
        }

        val user = User(
            id = generateUserId(),
            email = email
        )

        return repository.save(user)
    }

    private fun validateEmail(email: String) {
        val emailRegex = "^[A-Za-z0-9+_.-]+@[A-Za-z0-9.-]+\\.[A-Z|a-z]{2,}$".toRegex()
        require(emailRegex.matches(email)) { "Invalid email format" }
    }

    private fun generateUserId(): String = "user${System.currentTimeMillis()}"
}

data class User(val id: String, val email: String)

interface UserRepository {
    suspend fun save(user: User): User
    suspend fun findByEmail(email: String): User?
}

class DuplicateUserException(message: String) : Exception(message)
```

**REFACTOR: Improve with extension functions and better error handling**
```kotlin
// src/main/kotlin/com/example/domain/UserService.kt
package com.example.domain

import java.util.UUID

/**
 * @CODE:USER-001 | SPEC: SPEC-USER-001.md | TEST: UserServiceTest.kt
 * User management service with email validation
 */
class UserService(private val repository: UserRepository) {

    suspend fun createUser(email: String): User {
        email.validateEmail()

        repository.findByEmail(email)?.let {
            throw DuplicateUserException("User with email $email already exists")
        }

        return User(
            id = UUID.randomUUID().toString(),
            email = email
        ).also { repository.save(it) }
    }
}

// Extension function for email validation
private fun String.validateEmail() {
    val emailRegex = "^[A-Za-z0-9+_.-]+@[A-Za-z0-9.-]+\\.[A-Z|a-z]{2,}$".toRegex()
    require(emailRegex.matches(this)) { "Invalid email format: $this" }
}

data class User(val id: String, val email: String)

interface UserRepository {
    suspend fun save(user: User): User
    suspend fun findByEmail(email: String): User?
}

sealed class UserException(message: String) : Exception(message)
class DuplicateUserException(email: String) : UserException("User with email $email already exists")
class InvalidEmailException(email: String) : UserException("Invalid email format: $email")
```

## Example 3: Coroutines & Async Operations

```kotlin
// src/test/kotlin/com/example/api/ApiClientTest.kt
package com.example.api

import kotlinx.coroutines.test.runTest
import kotlinx.coroutines.async
import kotlinx.coroutines.delay
import org.junit.jupiter.api.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class ApiClientTest {

    @Test
    fun `should fetch data concurrently`() = runTest {
        // @TEST:API-001
        val client = ApiClient()

        val start = System.currentTimeMillis()
        val results = client.fetchMultiple(listOf("user1", "user2", "user3"))
        val duration = System.currentTimeMillis() - start

        assertEquals(3, results.size)
        assertTrue(duration < 1500) // Should complete in parallel, not sequentially
    }

    @Test
    fun `should handle partial failures`() = runTest {
        val client = ApiClient()

        val results = client.fetchWithFallback(listOf("valid", "invalid", "valid"))

        assertEquals(3, results.size)
        assertTrue(results[0].isSuccess)
        assertTrue(results[1].isFailure)
        assertTrue(results[2].isSuccess)
    }
}

// src/main/kotlin/com/example/api/ApiClient.kt
package com.example.api

import kotlinx.coroutines.*

/**
 * @CODE:API-001 | SPEC: SPEC-API-001.md | TEST: ApiClientTest.kt
 * Async API client with coroutines
 */
class ApiClient {

    suspend fun fetchMultiple(ids: List<String>): List<String> = coroutineScope {
        ids.map { id ->
            async {
                fetchData(id)
            }
        }.awaitAll()
    }

    suspend fun fetchWithFallback(ids: List<String>): List<Result<String>> = coroutineScope {
        ids.map { id ->
            async {
                runCatching { fetchData(id) }
            }
        }.awaitAll()
    }

    private suspend fun fetchData(id: String): String {
        delay(500) // Simulate API call
        if (id == "invalid") throw IllegalArgumentException("Invalid ID")
        return "Data for $id"
    }
}
```

## Example 4: Quality Gate Check

```bash
# Run all tests with coverage
./gradlew test jacocoTestReport

# Check coverage threshold
./gradlew jacocoTestCoverageVerification

# Run ktlint check
./gradlew ktlintCheck

# Auto-fix formatting issues
./gradlew ktlintFormat

# Build project
./gradlew build

# TRUST 5 validation
echo "T - Test coverage"
./gradlew test jacocoTestReport
# Verify coverage ≥85% in build/reports/jacoco/test/html/index.html

echo "R - Readable code"
./gradlew ktlintCheck

echo "U - Unified types"
# Kotlin has strong type system by default

echo "S - Security scan"
./gradlew dependencyCheckAnalyze

echo "T - Trackable with @TAG"
rg '@(CODE|TEST|SPEC):' -n src/ --type kotlin
```

**JaCoCo coverage verification**:
```kotlin
// Add to build.gradle.kts
tasks.jacocoTestCoverageVerification {
    violationRules {
        rule {
            limit {
                minimum = "0.85".toBigDecimal()
            }
        }

        rule {
            element = "CLASS"
            limit {
                counter = "LINE"
                value = "COVEREDRATIO"
                minimum = "0.80".toBigDecimal()
            }
            excludes = listOf(
                "*.config.*",
                "*.dto.*",
                "*Application*"
            )
        }
    }
}

tasks.check {
    dependsOn(tasks.jacocoTestCoverageVerification)
}
```

## Example 5: Sealed Classes & When Expressions

```kotlin
// src/test/kotlin/com/example/domain/PaymentProcessorTest.kt
package com.example.domain

import org.junit.jupiter.api.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class PaymentProcessorTest {

    @Test
    fun `should process credit card payment`() {
        // @TEST:PAY-001
        val processor = PaymentProcessor()
        val payment = Payment.CreditCard("1234-5678", 100.0)

        val result = processor.process(payment)

        assertTrue(result is PaymentResult.Success)
        assertEquals(100.0, (result as PaymentResult.Success).amount)
    }

    @Test
    fun `should handle insufficient funds`() {
        val processor = PaymentProcessor()
        val payment = Payment.BankTransfer("ACC123", 1000.0, insufficientFunds = true)

        val result = processor.process(payment)

        assertTrue(result is PaymentResult.Failure)
        assertEquals("Insufficient funds", (result as PaymentResult.Failure).reason)
    }
}

// src/main/kotlin/com/example/domain/PaymentProcessor.kt
package com.example.domain

/**
 * @CODE:PAY-001 | SPEC: SPEC-PAY-001.md | TEST: PaymentProcessorTest.kt
 * Payment processing with sealed classes
 */
sealed class Payment {
    data class CreditCard(val cardNumber: String, val amount: Double) : Payment()
    data class BankTransfer(val accountNumber: String, val amount: Double, val insufficientFunds: Boolean = false) : Payment()
    data class Cryptocurrency(val wallet: String, val amount: Double, val currency: String) : Payment()
}

sealed class PaymentResult {
    data class Success(val transactionId: String, val amount: Double) : PaymentResult()
    data class Failure(val reason: String) : PaymentResult()
    data object Pending : PaymentResult()
}

class PaymentProcessor {

    fun process(payment: Payment): PaymentResult = when (payment) {
        is Payment.CreditCard -> {
            if (payment.cardNumber.isValid()) {
                PaymentResult.Success("TXN-${System.currentTimeMillis()}", payment.amount)
            } else {
                PaymentResult.Failure("Invalid card number")
            }
        }
        is Payment.BankTransfer -> {
            if (payment.insufficientFunds) {
                PaymentResult.Failure("Insufficient funds")
            } else {
                PaymentResult.Success("TXN-${System.currentTimeMillis()}", payment.amount)
            }
        }
        is Payment.Cryptocurrency -> {
            PaymentResult.Pending // Crypto payments need confirmation
        }
    }

    private fun String.isValid(): Boolean = this.matches(Regex("\\d{4}-\\d{4}"))
}
```

---

## TRUST 5 Integration

### Test Coverage (≥85%)
```bash
./gradlew test jacocoTestReport
# Check: build/reports/jacoco/test/html/index.html
```

### Readable Code
```bash
./gradlew ktlintCheck
./gradlew ktlintFormat
```

### Unified Types
- Use Kotlin's type system
- Leverage sealed classes for exhaustive when expressions
- Use data classes for immutability

### Security
```bash
./gradlew dependencyCheckAnalyze
# Check: build/reports/dependency-check-report.html
```

### Trackable with @TAG
```bash
rg '@(CODE|TEST|SPEC):' -n src/ --type kotlin
```

---

_For detailed CLI reference, see [reference.md](reference.md)_
