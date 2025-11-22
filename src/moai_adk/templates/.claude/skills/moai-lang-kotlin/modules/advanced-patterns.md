# Kotlin Advanced Patterns — Enterprise Development

_Last updated: 2025-11-22_

## Reactive Streams with Flow

### Cold Flows (Lazy Evaluation)

```kotlin
// Flow is a cold stream - only emits when collected
fun getUsersFlow(): Flow<User> = flow {
    println("Starting user fetch...")
    for (i in 1..5) {
        delay(100)
        emit(User(id = i, name = "User$i"))
    }
}

// Each collection starts fresh
getUsersFlow().collect { user ->
    println("Got: ${user.name}")
}
// Prints: "Starting user fetch..." then 5 users

getUsersFlow().collect { user ->
    println("Got again: ${user.name}")
}
// Prints: "Starting user fetch..." again - flow restarts
```

### Hot Flows (StateFlow & SharedFlow)

```kotlin
class UserViewModel {
    // StateFlow - always has current value
    private val _users = MutableStateFlow<List<User>>(emptyList())
    val users: StateFlow<List<User>> = _users.asStateFlow()

    // SharedFlow - broadcast without state
    private val _events = MutableSharedFlow<UserEvent>(
        replay = 1,  // Replay last event for new collectors
        extraBufferCapacity = 10
    )
    val events: SharedFlow<UserEvent> = _events.asSharedFlow()

    fun fetchUsers() {
        viewModelScope.launch {
            val result = userRepository.fetchUsers()
            _users.value = result
            _events.emit(UserEvent.UsersFetched(result.size))
        }
    }
}

// Collect with replayCache
viewModel.events.collect { event ->
    when (event) {
        is UserEvent.UsersFetched -> showMessage("${event.count} users loaded")
        else -> {}
    }
}
```

### Flow Composition Patterns

```kotlin
// Chaining operations with efficient stream processing
fun getUsersWithFiltering(): Flow<User> {
    return getUsersFlow()
        .filter { it.isActive }  // Only active users
        .map { it.copy(name = it.name.uppercase()) }
        .distinctUntilChanged { old, new -> old.id == new.id }
        .catch { e ->
            println("Error: ${e.message}")
            emit(User(id = 0, name = "Error"))
        }
        .timeout(5000)  // Timeout after 5 seconds
}

// Combining multiple flows
fun combinedFlow(): Flow<Pair<User, String>> {
    return combine(
        getUsersFlow(),
        getStatusFlow()
    ) { user, status ->
        user to status
    }
}

// Switching flows based on events
fun switchFlow(): Flow<User> {
    return eventFlow
        .switchMap { event ->
            when (event) {
                is Event.FetchUser -> getUserFlow(event.userId)
                else -> emptyFlow()
            }
        }
}
```

## Kotlin Multiplatform (KMP) Architecture

### expect/actual Pattern

```kotlin
// commonMain/kotlin/data/Database.kt
expect class DatabaseDriver {
    suspend fun initialize(): Result<Unit>
    suspend fun saveUser(user: User): Result<Unit>
    suspend fun getUser(id: String): Result<User>
    suspend fun deleteAllUsers(): Result<Unit>
}

// androidMain/kotlin/data/DatabaseDriver.kt
actual class DatabaseDriver(context: android.content.Context) {
    private val db: AppDatabase by lazy {
        Room.databaseBuilder(context, AppDatabase::class.java, "users.db")
            .build()
    }

    actual suspend fun initialize() = try {
        Result.success(Unit)  // Room auto-initializes
    } catch (e: Exception) {
        Result.failure(e)
    }

    actual suspend fun saveUser(user: User) = try {
        db.userDao().insert(user)
        Result.success(Unit)
    } catch (e: Exception) {
        Result.failure(e)
    }

    actual suspend fun getUser(id: String) = try {
        val user = db.userDao().getUser(id)
        Result.success(user)
    } catch (e: Exception) {
        Result.failure(e)
    }

    actual suspend fun deleteAllUsers() = try {
        db.userDao().deleteAll()
        Result.success(Unit)
    } catch (e: Exception) {
        Result.failure(e)
    }
}

// iosMain/kotlin/data/DatabaseDriver.kt
actual class DatabaseDriver {
    private val sqlDriver: SqlDriver = NativeSqliteDriver(AppDatabase.Schema, "users.db")
    private val database = AppDatabase(sqlDriver)

    actual suspend fun initialize() = try {
        Result.success(Unit)
    } catch (e: Exception) {
        Result.failure(e)
    }

    actual suspend fun saveUser(user: User) = try {
        database.userQueries.insert(user.id, user.name, user.email)
        Result.success(Unit)
    } catch (e: Exception) {
        Result.failure(e)
    }

    actual suspend fun getUser(id: String) = try {
        val user = database.userQueries.getUser(id).executeAsOneOrNull()
        user?.let { Result.success(it) } ?: Result.failure(Exception("Not found"))
    } catch (e: Exception) {
        Result.failure(e)
    }

    actual suspend fun deleteAllUsers() = try {
        database.userQueries.deleteAll()
        Result.success(Unit)
    } catch (e: Exception) {
        Result.failure(e)
    }
}

// Usage in commonMain - Platform-agnostic
suspend fun initializeDatabase(): Result<Unit> {
    return DatabaseDriver().initialize()
}
```

### KMP Project Structure

```
kmp-app/
├── shared/
│   ├── src/
│   │   ├── commonMain/kotlin/
│   │   │   ├── domain/
│   │   │   │   ├── User.kt
│   │   │   │   └── UseCase.kt
│   │   │   ├── data/
│   │   │   │   ├── Database.kt (expect class)
│   │   │   │   └── UserRepository.kt
│   │   │   └── presentation/
│   │   │       ├── UserViewModel.kt
│   │   │       └── UserState.kt
│   │   ├── androidMain/kotlin/
│   │   │   ├── data/
│   │   │   │   └── DatabaseDriver.kt (actual)
│   │   │   └── di/
│   │   │       └── AndroidModule.kt
│   │   ├── iosMain/kotlin/
│   │   │   ├── data/
│   │   │   │   └── DatabaseDriver.kt (actual)
│   │   │   └── di/
│   │   │       └── IosModule.kt
│   │   └── commonTest/kotlin/
│   │       ├── domain/
│   │       └── data/
│   └── build.gradle.kts (KMP configuration)
├── androidApp/
├── iosApp/
└── webApp/
```

## Error Handling Patterns

### Result Type with Railway-Oriented Programming

```kotlin
// Sealed class for explicit error handling
sealed class Result<out T> {
    data class Success<T>(val data: T) : Result<T>()
    data class Failure(val exception: Throwable) : Result<Nothing>()
    object Loading : Result<Nothing>()
}

// Extension functions for Result
inline fun <T, R> Result<T>.map(transform: (T) -> R): Result<R> =
    when (this) {
        is Result.Success -> Result.Success(transform(data))
        is Result.Failure -> this
        else -> this
    }

inline fun <T> Result<T>.onSuccess(action: (T) -> Unit): Result<T> {
    if (this is Result.Success) action(data)
    return this
}

inline fun <T> Result<T>.onFailure(action: (Throwable) -> Unit): Result<T> {
    if (this is Result.Failure) action(exception)
    return this
}

// Usage - Railway-oriented
suspend fun processUser(userId: String): Result<UserDto> {
    return getUserFromApi(userId)
        .map { it.toDomain() }
        .map { validateUser(it) }
        .onSuccess { cache.save(it) }
        .onFailure { logger.error("Failed to process user", it) }
}

// Chain operations with error context
suspend fun fetchAndProcessUsers(): Result<List<User>> {
    return getUsersFromDatabase()
        .map { users ->
            users.filter { it.isActive }
                .map { it.copy(updated = Clock.System.now()) }
        }
        .onFailure { database.recover() }
}
```

### Exception Context with Custom Types

```kotlin
sealed class DomainException(message: String, cause: Throwable? = null) :
    Exception(message, cause) {

    class NetworkException(message: String, cause: Throwable? = null) :
        DomainException(message, cause)

    class ValidationException(message: String, val field: String) :
        DomainException("Validation error in $field: $message")

    class NotFoundException(message: String, val resourceId: String) :
        DomainException(message)
}

// Usage with type-safe exception handling
suspend fun fetchUserSafe(userId: String): Result<User> = try {
    val user = apiClient.getUser(userId)
    if (user == null) {
        Result.Failure(
            DomainException.NotFoundException(
                "User not found",
                userId
            )
        )
    } else {
        Result.Success(user)
    }
} catch (e: IOException) {
    Result.Failure(DomainException.NetworkException(e.message ?: "Unknown", e))
} catch (e: Exception) {
    Result.Failure(e)
}
```

## Dependency Injection with Koin

### Koin Setup for KMP

```kotlin
// commonMain/kotlin/di/AppModule.kt
val appModule = module {
    single<HttpClient> {
        HttpClient {
            install(JsonFeature) {
                serializer = KotlinxSerializer()
            }
            install(Logging) {
                level = LogLevel.INFO
            }
        }
    }

    single<UserRepository> { UserRepositoryImpl(get()) }
    single<UserDatabase> { UserDatabase(get()) }

    viewModel { UserListViewModel(get(), get()) }
}

// androidMain/kotlin/di/AndroidModule.kt
val androidModule = module {
    single<DatabaseDriver> {
        DatabaseDriver(androidContext())
    }
}

// iOS and Web modules similarly...

// In commonMain/kotlin/App.kt
fun initializeKoin() {
    startKoin {
        modules(
            appModule,
            androidModule,  // Automatically selects based on platform
            iosModule,
            webModule
        )
    }
}
```

## Advanced Coroutine Patterns

### Structured Concurrency with Proper Scoping

```kotlin
class UserListViewModel : ViewModel() {
    private val _state = MutableStateFlow<State>(State.Loading)
    val state: StateFlow<State> = _state.asStateFlow()

    fun loadUsers() {
        viewModelScope.launch {
            try {
                _state.value = State.Loading

                // Parallel fetching with proper scoping
                val usersDeferred = async { userRepository.fetchUsers() }
                val statsDeferred = async { statsRepository.fetchStats() }

                val users = usersDeferred.await()
                val stats = statsDeferred.await()

                _state.value = State.Success(users, stats)
            } catch (e: CancellationException) {
                // Clean up on cancellation
                _state.value = State.Cancelled
                throw e
            } catch (e: Exception) {
                _state.value = State.Error(e.message ?: "Unknown error")
            }
        }
    }

    sealed class State {
        object Loading : State()
        data class Success(val users: List<User>, val stats: Stats) : State()
        data class Error(val message: String) : State()
        object Cancelled : State()
    }
}

// Timeout with fallback
suspend fun fetchUserWithTimeout(userId: String): User = try {
    withTimeoutOrNull(5000) {
        userRepository.fetchUser(userId)
    } ?: getCachedUser(userId) ?: User.EMPTY
} catch (e: TimeoutCancellationException) {
    logger.warn("Fetch timeout for $userId")
    getCachedUser(userId) ?: User.EMPTY
}
```

### Recursive Coroutine with Tail Recursion

```kotlin
// Tail-recursive coroutine for pagination
suspend fun fetchAllUsersPaginated(
    pageSize: Int = 20,
    pageNum: Int = 1,
    accumulated: List<User> = emptyList()
): List<User> {
    val page = userRepository.fetchPage(pageSize, pageNum)

    return if (page.isEmpty()) {
        accumulated
    } else {
        // Tail recursion - same stack frame
        fetchAllUsersPaginated(pageSize, pageNum + 1, accumulated + page)
    }
}

// Iterative approach (preferred for coroutines)
suspend fun fetchAllUsersPaginatedIterative(pageSize: Int = 20): List<User> {
    val result = mutableListOf<User>()
    var pageNum = 1

    while (true) {
        val page = userRepository.fetchPage(pageSize, pageNum)
        if (page.isEmpty()) break
        result.addAll(page)
        pageNum++
    }

    return result
}
```

## DSL and Builder Patterns

### Custom DSL with Context Receivers (Kotlin 1.7+)

```kotlin
// HTML DSL builder
class HtmlBuilder {
    private val elements = mutableListOf<String>()

    fun h1(text: String) {
        elements.add("<h1>$text</h1>")
    }

    fun p(text: String) {
        elements.add("<p>$text</p>")
    }

    fun div(block: HtmlBuilder.() -> Unit) {
        elements.add("<div>")
        block()
        elements.add("</div>")
    }

    fun build() = elements.joinToString("\n")
}

fun html(block: HtmlBuilder.() -> Unit): String {
    return HtmlBuilder().apply(block).build()
}

// Usage
val page = html {
    h1("Welcome")
    p("This is a page")
    div {
        p("Nested content")
    }
}

// JSON DSL
class JsonBuilder {
    private val map = mutableMapOf<String, Any?>()

    operator fun set(key: String, value: Any?) {
        map[key] = value
    }

    fun build(): String = Json.encodeToString(map)
}

fun json(block: JsonBuilder.() -> Unit): String {
    return JsonBuilder().apply(block).build()
}

val userJson = json {
    set("name", "John")
    set("email", "john@example.com")
    set("active", true)
}
```

## Sealed Class Hierarchies

### Domain Modeling with Sealed Classes

```kotlin
// Sealed hierarchy for type-safe domain modeling
sealed class Event {
    abstract val timestamp: Long

    data class UserCreated(
        val userId: String,
        val name: String,
        override val timestamp: Long
    ) : Event()

    data class UserDeleted(
        val userId: String,
        override val timestamp: Long
    ) : Event()

    data class UserUpdated(
        val userId: String,
        val changes: Map<String, Any>,
        override val timestamp: Long
    ) : Event()

    // Unseal warning for exhaustive when
    data class UnknownEvent(override val timestamp: Long) : Event()
}

// Pattern matching with exhaustive when
fun handleEvent(event: Event) {
    when (event) {
        is Event.UserCreated -> notifyUserCreated(event)
        is Event.UserDeleted -> notifyUserDeleted(event)
        is Event.UserUpdated -> updateUI(event)
        is Event.UnknownEvent -> logger.warn("Unknown event")
    }
}

// Sealed hierarchy for API responses
sealed class ApiResponse<out T> {
    data class Success<T>(val data: T, val statusCode: Int = 200) : ApiResponse<T>()
    data class Failure(val error: ApiError) : ApiResponse<Nothing>()
    object Loading : ApiResponse<Nothing>()
}

data class ApiError(
    val code: Int,
    val message: String,
    val details: Map<String, Any>? = null
)
```

## Performance Best Practices

### Memory Optimization

```kotlin
// 1. Use inline classes for zero-overhead wrappers
@JvmInline
value class UserId(val value: String)

@JvmInline
value class Email(val value: String)

// 2. Use Sequence for lazy evaluation
fun processLargeList(items: List<Int>): List<Int> {
    // Creates intermediate lists - inefficient
    return items
        .filter { it > 100 }          // Creates List<Int>
        .map { it * 2 }               // Creates another List<Int>
        .take(10)                     // Finally takes 10
}

fun processLargeListEfficient(items: List<Int>): List<Int> {
    // Lazy evaluation - no intermediate lists
    return items.asSequence()
        .filter { it > 100 }           // Lazy
        .map { it * 2 }                // Lazy
        .take(10)                      // Evaluates only 10 items
        .toList()                      // Single allocation
}

// 3. Use primitive arrays for collections
val intArray = IntArray(1000)          // Primitive array
val boxedArray = Array(1000) { 0 }     // Boxed integers - 8x memory

// 4. Object pooling for frequently allocated objects
class ObjectPool<T>(val factory: () -> T) {
    private val available = mutableListOf<T>()
    private val inUse = mutableSetOf<T>()

    fun acquire(): T {
        val obj = if (available.isNotEmpty()) available.removeAt(0) else factory()
        inUse.add(obj)
        return obj
    }

    fun release(obj: T) {
        inUse.remove(obj)
        available.add(obj)
    }
}

// 5. Use inline lambdas to avoid allocation
inline fun <T, R> List<T>.mapInline(crossinline transform: (T) -> R): List<R> {
    val result = mutableListOf<R>()
    for (item in this) {
        result.add(transform(item))
    }
    return result
}
```

---

**Last Updated**: 2025-11-22
**Related**: moai-lang-kotlin/SKILL.md, examples.md, modules/optimization.md

