# moai-lang-kotlin - Working Examples

_Last updated: 2025-11-13_

## Quick Start (Kotlin 2.0 Minimal Setup)

```kotlin
// Modern Kotlin 2.0 with coroutines
import kotlinx.coroutines.*

class HelloWorld {
    suspend fun greetAsync(name: String): String {
        delay(100) // Simulate async work
        return "Hello, $name!"
    }
}

// Usage
suspend fun main() {
    val greeting = HelloWorld()
    val result = greeting.greetAsync("Kotlin")
    println(result)  // Output: Hello, Kotlin!
}
```

## Basic Usage Examples

### Example 1: Coroutines - Basic Async/Await

```kotlin
import kotlinx.coroutines.*

class UserService {
    suspend fun fetchUserData(userId: String): String {
        delay(1000) // Simulate API call
        return "User data for $userId"
    }
}

suspend fun main() {
    val userService = UserService()

    // Simple async operation
    val userData = userService.fetchUserData("123")
    println(userData)

    // Multiple concurrent operations
    coroutineScope {
        val user1 = async { userService.fetchUserData("1") }
        val user2 = async { userService.fetchUserData("2") }
        val user3 = async { userService.fetchUserData("3") }

        val results = listOf(user1, user2, user3).awaitAll()
        results.forEach { println(it) }
    }
}
```

### Example 2: Extension Functions

```kotlin
// Define extension functions for enhanced API
fun String.isValidEmail(): Boolean {
    val emailRegex = "^[A-Za-z0-9+_.-]+@(.+)$".toRegex()
    return this.matches(emailRegex)
}

fun <T> List<T>.secondOrNull(): T? {
    return if (this.size > 1) this[1] else null
}

fun Int.factorial(): Int {
    return if (this <= 1) 1 else this * (this - 1).factorial()
}

// Usage
fun main() {
    val email = "user@example.com"
    if (email.isValidEmail()) {
        println("Valid email")
    }

    val numbers = listOf(1, 2, 3, 4, 5)
    println(numbers.secondOrNull()) // 2

    println(5.factorial()) // 120
}
```

### Example 3: Sealed Classes and Pattern Matching

```kotlin
sealed class Result<out T> {
    data class Success<T>(val data: T) : Result<T>()
    data class Error(val exception: Throwable) : Result<Nothing>()
    object Loading : Result<Nothing>()
}

fun processResult(result: Result<String>) {
    when (result) {
        is Result.Success -> println("Success: ${result.data}")
        is Result.Error -> println("Error: ${result.exception.message}")
        is Result.Loading -> println("Loading...")
    }
}

// Usage
val result: Result<String> = Result.Success("Hello")
processResult(result)
```

### Example 4: Data Classes and Destructuring

```kotlin
data class User(
    val id: String,
    val name: String,
    val email: String,
    val age: Int
)

fun main() {
    val user = User("1", "John Doe", "john@example.com", 30)

    // Destructuring
    val (id, name, email, age) = user
    println("$name ($email) - Age: $age")

    // Copy with modifications
    val updatedUser = user.copy(age = 31)
    println(updatedUser)

    // Equality comparison (automatic for data classes)
    val sameUser = User("1", "John Doe", "john@example.com", 30)
    println(user == sameUser) // true
}
```

## Intermediate Patterns

### Example 5: Higher-Order Functions

```kotlin
// Function that takes a function as parameter
fun <T, R> transformList(
    items: List<T>,
    transform: (T) -> R
): List<R> {
    return items.map(transform)
}

// Function that returns a function
fun createMultiplier(factor: Int): (Int) -> Int {
    return { number -> number * factor }
}

// Usage
fun main() {
    val numbers = listOf(1, 2, 3, 4, 5)
    val doubled = transformList(numbers) { it * 2 }
    println(doubled) // [2, 4, 6, 8, 10]

    val times3 = createMultiplier(3)
    println(times3(5)) // 15
}
```

### Example 6: Flow - Reactive Streams

```kotlin
import kotlinx.coroutines.flow.*

fun countDownFlow(): Flow<Int> = flow {
    for (i in 5 downTo 1) {
        delay(1000)
        emit(i)
    }
}

suspend fun main() {
    countDownFlow()
        .filter { it > 2 }
        .map { "Number: $it" }
        .collect { println(it) }

    // Practical example: Processing events
    val eventFlow = (1..100).asFlow()
        .onEach { delay(10) }
        .filter { it % 2 == 0 }
        .map { it * it }

    eventFlow.collect { println(it) }
}
```

### Example 7: Scope Functions (let, apply, run, with)

```kotlin
data class Person(
    var name: String = "",
    var age: Int = 0,
    var email: String = ""
)

fun main() {
    val person = Person()

    // apply: modifies object and returns it
    person.apply {
        name = "John Doe"
        age = 30
        email = "john@example.com"
    }.also {
        println("Created: $it")
    }

    // let: transforms value safely
    person.email.let { email ->
        if (email.isNotEmpty()) {
            println("Valid email: $email")
        }
    }

    // with: calls multiple methods on same object
    with(person) {
        println("Name: $name")
        println("Age: $age")
        println("Email: $email")
    }

    // run: like let but returns result
    val message = person.run {
        "User $name is $age years old"
    }
    println(message)
}
```

### Example 8: Null Safety and Elvis Operator

```kotlin
data class Address(val street: String?, val city: String?)

fun getDisplayAddress(address: Address?): String {
    // Elvis operator: use default if null
    return address?.let { addr ->
        "${addr.street ?: "Unknown"}, ${addr.city ?: "Unknown"}"
    } ?: "No address available"
}

fun main() {
    val address1 = Address("123 Main St", "Springfield")
    println(getDisplayAddress(address1))
    // Output: 123 Main St, Springfield

    val address2 = Address(null, "Portland")
    println(getDisplayAddress(address2))
    // Output: Unknown, Portland

    println(getDisplayAddress(null))
    // Output: No address available
}
```

### Example 9: Multiplatform Code with expect/actual

```kotlin
// commonMain/kotlin/Logger.kt
expect class Logger {
    fun debug(tag: String, message: String)
    fun error(tag: String, message: String, throwable: Throwable? = null)
}

// androidMain/kotlin/Logger.kt
actual class Logger {
    actual fun debug(tag: String, message: String) {
        android.util.Log.d(tag, message)
    }

    actual fun error(tag: String, message: String, throwable: Throwable?) {
        android.util.Log.e(tag, message, throwable)
    }
}

// iosMain/kotlin/Logger.kt
actual class Logger {
    actual fun debug(tag: String, message: String) {
        println("[$tag] DEBUG: $message")
    }

    actual fun error(tag: String, message: String, throwable: Throwable?) {
        println("[$tag] ERROR: $message")
        throwable?.printStackTrace()
    }
}
```

### Example 10: Compose Multiplatform - Basic UI

```kotlin
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.foundation.layout.*

@Composable
fun UserListScreen(users: List<String>) {
    Column(modifier = Modifier.fillMaxSize()) {
        Text(
            text = "Users",
            style = MaterialTheme.typography.headlineLarge,
            modifier = Modifier.padding(16.dp)
        )

        LazyColumn(modifier = Modifier.fillMaxSize()) {
            items(users.size) { index ->
                UserItem(users[index])
            }
        }
    }
}

@Composable
fun UserItem(userName: String) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .padding(8.dp)
    ) {
        Text(
            text = userName,
            modifier = Modifier.padding(16.dp)
        )
    }
}

// Usage
@Composable
fun App() {
    val users = listOf("Alice", "Bob", "Charlie")
    UserListScreen(users)
}
```

## Advanced Patterns

### Example 11: Inline Classes for Type-Safe IDs

```kotlin
@JvmInline
value class UserId(val value: String) {
    companion object {
        fun generate() = UserId(java.util.UUID.randomUUID().toString())
    }
}

@JvmInline
value class Email(val value: String) {
    fun isValid(): Boolean = value.contains("@") && value.contains(".")
}

data class User(
    val id: UserId,
    val email: Email,
    val name: String
)

fun processUser(user: User) {
    if (user.email.isValid()) {
        println("Processing user: ${user.id.value}")
    }
}
```

### Example 12: DSL Building Pattern

```kotlin
class HtmlBuilder {
    private val elements = mutableListOf<String>()

    fun tag(name: String, attributes: String = "", content: String = "") {
        elements.add("<$name$attributes>$content</$name>")
    }

    fun div(attributes: String = "", block: HtmlBuilder.() -> Unit = {}) {
        val builder = HtmlBuilder()
        builder.block()
        elements.add("<div$attributes>${builder.elements.joinToString()}</div>")
    }

    fun p(content: String) = tag("p", content = content)
    fun h1(content: String) = tag("h1", content = content)

    fun build() = elements.joinToString("\n")
}

fun html(block: HtmlBuilder.() -> Unit): String {
    val builder = HtmlBuilder()
    builder.block()
    return "<html>${builder.build()}</html>"
}

// Usage
fun main() {
    val page = html {
        h1("Welcome")
        div {
            p("This is a paragraph")
            p("And another one")
        }
    }
    println(page)
}
```

### Example 13: Dependency Injection with Koin

```kotlin
import org.koin.core.context.startKoin
import org.koin.core.module.dsl.singleOf
import org.koin.dsl.module

class UserRepository {
    fun getUser(id: String) = "User: $id"
}

class UserService(private val repository: UserRepository) {
    fun fetchUser(id: String) = repository.getUser(id)
}

val appModule = module {
    singleOf(::UserRepository)
    singleOf(::UserService)
}

fun main() {
    startKoin {
        modules(appModule)
    }

    val userService: UserService = org.koin.java.KoinJavaComponent.get()
    println(userService.fetchUser("123"))
}
```

### Example 14: Error Handling with Result Type

```kotlin
suspend fun fetchUserSafely(userId: String): Result<String> = try {
    val user = fetchUser(userId) // May throw
    Result.success(user)
} catch (e: Exception) {
    Result.failure(e)
}

suspend fun fetchUser(userId: String): String {
    delay(100)
    if (userId.isEmpty()) throw IllegalArgumentException("ID cannot be empty")
    return "User: $userId"
}

suspend fun main() {
    // Chain operations with Result
    fetchUserSafely("123")
        .onSuccess { user ->
            println("Got user: $user")
        }
        .onFailure { error ->
            println("Error: ${error.message}")
        }

    // Or use getOrNull for optional handling
    val user = fetchUserSafely("456").getOrNull()
    println(user ?: "User not found")
}
```

### Example 15: Generics with Type Variance

```kotlin
// Covariant: Producer (out)
interface Producer<out T> {
    fun produce(): T
}

// Contravariant: Consumer (in)
interface Consumer<in T> {
    fun consume(item: T)
}

class NumberProducer : Producer<Int> {
    override fun produce() = 42
}

class NumberConsumer : Consumer<Number> {
    override fun consume(item: Number) {
        println("Consumed: $item")
    }
}

fun main() {
    val intProducer: Producer<Int> = NumberProducer()
    val numberProducer: Producer<Number> = intProducer // OK: covariance

    val numberConsumer: Consumer<Number> = NumberConsumer()
    val intConsumer: Consumer<Int> = numberConsumer // OK: contravariance
}
```

### Example 16: Sequence for Lazy Evaluation

```kotlin
fun main() {
    val numbers = (1..1_000_000).asSequence()
        .filter { it % 2 == 0 }
        .map { it * it }
        .take(5)
        .toList()

    println(numbers) // [4, 16, 36, 64, 100]

    // Compare with List (eager evaluation)
    val listResult = (1..1_000_000)
        .filter { it % 2 == 0 }
        .map { it * it }
        .take(5)
    // This processes all elements, even though we only take 5
}
```

### Example 17: Advanced Coroutine Patterns

```kotlin
import kotlinx.coroutines.*

suspend fun parallelProcessing() {
    coroutineScope {
        val job1 = async { fetchData("API1") }
        val job2 = async { fetchData("API2") }
        val job3 = async { fetchData("API3") }

        val results = awaitAll(job1, job2, job3)
        println("All results: $results")
    }
}

suspend fun fetchData(source: String): String {
    delay(1000)
    return "Data from $source"
}

suspend fun retryWithBackoff(
    maxRetries: Int = 3,
    initialDelay: Long = 100,
    block: suspend () -> String
): String {
    var delay = initialDelay
    repeat(maxRetries) { attempt ->
        try {
            return block()
        } catch (e: Exception) {
            if (attempt == maxRetries - 1) throw e
            delay(delay)
            delay *= 2
        }
    }
    throw Exception("Max retries exceeded")
}

suspend fun main() {
    parallelProcessing()

    val data = retryWithBackoff { fetchData("API") }
    println(data)
}
```

### Example 18: State Management with StateFlow

```kotlin
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow

class CounterViewModel {
    private val _count = MutableStateFlow(0)
    val count: StateFlow<Int> = _count.asStateFlow()

    fun increment() {
        _count.value++
    }

    fun decrement() {
        _count.value--
    }
}

suspend fun main() {
    val viewModel = CounterViewModel()

    // Collect state changes
    coroutineScope {
        launch {
            viewModel.count.collect { count ->
                println("Count: $count")
            }
        }

        launch {
            repeat(5) {
                delay(100)
                viewModel.increment()
            }
        }
    }
}
```

### Example 19: Contract-Based Programming

```kotlin
@kotlin.contracts.ExperimentalContracts
fun String?.isNotNullOrEmpty(): Boolean {
    contract {
        returns(true) implies (this@isNotNullOrEmpty != null)
    }
    return !this.isNullOrEmpty()
}

@kotlin.contracts.ExperimentalContracts
fun main() {
    val str: String? = "Hello"

    if (str.isNotNullOrEmpty()) {
        // Compiler knows str is not null here
        println(str.length)
    }
}
```

### Example 20: Reified Type Parameters

```kotlin
inline fun <reified T> parseJson(json: String): T {
    // Can use T::class because it's reified
    return when (T::class) {
        String::class -> json as T
        Int::class -> json.toInt() as T
        else -> throw IllegalArgumentException("Unsupported type")
    }
}

fun main() {
    val str = parseJson<String>("\"hello\"")
    val num = parseJson<Int>("42")
    println("$str, $num")
}
```

## Context7 MCP Integration Examples

### Example 21: Access Real-Time Documentation

```kotlin
// Using Context7 MCP to get latest Kotlin docs
import moai.context7.Context7

suspend fun getLatestCoroutinePatterns() {
    val context7 = Context7("kotlin-coroutines")
    val docs = context7.getLibraryDocs()

    // docs contains latest patterns from official documentation
    println("Latest patterns: ${docs.bestPractices}")
}
```

### Example 22: Integration with Context7

```kotlin
// Example of accessing latest Flow patterns
val flowDocs = Context7("kotlin/kotlinx.coroutines")
    .getTopic("flow-operators")
    .let { docs ->
        docs.examples.filter { it.isProduction }
    }
```

---

## Testing Examples

### Example 23: Testing Coroutines

```kotlin
import kotlinx.coroutines.test.runTest
import org.junit.Test
import kotlin.test.assertEquals

class UserRepositoryTest {
    @Test
    fun testFetchUser() = runTest {
        val repository = UserRepository()
        val user = repository.fetchUser("123")
        assertEquals("User: 123", user)
    }
}
```

### Example 24: Testing Flows

```kotlin
import kotlinx.coroutines.flow.flow
import kotlinx.coroutines.test.runTest
import app.cash.turbine.test

@Test
fun flowTest() = runTest {
    val flow = flow { (1..5).forEach { emit(it) } }

    flow.test {
        assertEquals(1, awaitItem())
        assertEquals(2, awaitItem())
        awaitComplete()
    }
}
```

---

## Migration Guides

### Example 25: Java to Kotlin - Collections

```kotlin
// Java
List<String> javaList = new ArrayList<>();
javaList.add("hello");
javaList.forEach(System.out::println);

// Kotlin (more concise)
val kotlinList = mutableListOf("hello")
kotlinList.forEach(::println)

// Or even simpler with functional style
listOf("hello").forEach(::println)
```

### Example 26: Java to Kotlin - Null Safety

```kotlin
// Java (nullable)
String maybeName = getName();
if (maybeName != null) {
    System.out.println(maybeName.length());
}

// Kotlin (null-safe)
val name = getName()
name?.let { println(it.length) }

// Or with Elvis operator
println(name?.length ?: "Unknown")
```

### Example 27: Java to Kotlin - Classes

```kotlin
// Java
public class User {
    private String id;
    private String name;

    public User(String id, String name) {
        this.id = id;
        this.name = name;
    }

    public String getId() { return id; }
    public String getName() { return name; }
}

// Kotlin (one line!)
data class User(val id: String, val name: String)

// Kotlin with mutable properties
class UserMutable(var id: String, var name: String)
```

---

## Performance Tips

### Example 28: Lazy Initialization

```kotlin
class ExpensiveComputation {
    val result: String by lazy {
        println("Computing...")
        Thread.sleep(1000)
        "Result"
    }
}

fun main() {
    val computation = ExpensiveComputation()
    println("Created")
    println(computation.result) // Computed here
    println(computation.result) // Cached, no recomputation
}
```

### Example 29: Sequence vs List

```kotlin
fun main() {
    // List: eager, creates intermediate lists
    (1..10)
        .filter { println("Filtering $it"); it % 2 == 0 }
        .map { println("Mapping $it"); it * 2 }
        .toList()

    println("---")

    // Sequence: lazy, only processes needed elements
    (1..10)
        .asSequence()
        .filter { println("Filtering $it"); it % 2 == 0 }
        .map { println("Mapping $it"); it * 2 }
        .toList()
}
```

---

_For more detailed patterns, see reference.md_
