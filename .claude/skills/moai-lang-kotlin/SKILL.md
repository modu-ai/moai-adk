---
name: moai-lang-kotlin
version: 2.0.0
created: 2025-11-06
updated: 2025-11-06
status: active
description: "Kotlin best practices with Android development, Spring Boot backend, and modern coroutines for 2025"
keywords: [kotlin, programming, android, backend, spring, coroutines, kmp, mobile]
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - WebFetch
  - WebSearch
---

# Kotlin Development Mastery

**Modern Kotlin Development with 2025 Best Practices**

> Comprehensive Kotlin development guidance covering Android applications, Spring Boot backend services, KMP cross-platform development, and coroutine-based programming using the latest tools and frameworks.

## What It Does

### Android Development
- **Mobile App Development**: Jetpack Compose UI, MVVM architecture, modern Android patterns
- **Platform Integration**: Camera, location, notifications, background services
- **Performance Optimization**: Memory management, battery optimization, lifecycle awareness
- **Testing**: Unit tests, UI tests, integration tests with modern Android testing framework

### Backend Development
- **API Development**: Spring Boot with Kotlin, Ktor, Micronaut for microservices
- **Database Integration**: Exposed, jOOQ, Spring Data with coroutine support
- **Async Programming**: Coroutines, Flow, structured concurrency
- **Testing**: MockK, Kotest, TestContainers with coroutine support

### Cross-Platform Development
- **KMP Development**: Shared business logic across iOS, Android, Web, Desktop
- **Platform-Specific APIs**: Expect/actual implementations for platform differences
- **UI Development**: Compose Multiplatform for cross-platform UI
- **Testing**: Shared test suites across platforms

## When to Use

### Perfect Scenarios
- **Building Android applications with modern Jetpack Compose**
- **Developing backend microservices with Spring Boot and Ktor**
- **Creating cross-platform applications with Kotlin Multiplatform**
- **Implementing coroutine-based asynchronous programming**
- **Building reactive applications with Flow and structured concurrency**
- **Migrating Java codebases to Kotlin for modern development**
- **Developing Android libraries and SDKs**

### Common Triggers
- "Create Android app with Kotlin"
- "Build Kotlin backend API"
- "Set up KMP project"
- "Implement coroutines in Kotlin"
- "Optimize Kotlin performance"
- "Test Kotlin application"
- "Kotlin best practices"

## Tool Version Matrix (2025-11-06)

### Core Kotlin
- **Kotlin**: 2.0.20 (current) / 1.9.24 (LTS)
- **Kotlin Multiplatform**: 2.0.20
- **Package Managers**: Gradle 8.8 with Kotlin DSL, Maven
- **Runtime**: JVM 17+, Android API 21+, iOS 13+, WebAssembly

### Android Development
- **Android Studio**: Ladybug | 2024.2.1
- **Jetpack Compose**: 1.7.0
- **Compose BOM**: 2024.10.00
- **Android Gradle Plugin**: 8.7.0
- **KSP**: 2.0.20-1.0.25

### Backend Frameworks
- **Spring Boot**: 3.3.x with Kotlin support
- **Ktor**: 2.3.x - Asynchronous HTTP framework
- **Micronaut**: 4.7.x - Cloud-native framework
- **Exposed**: 0.53.x - SQL framework
- **jOOQ**: 3.19.x - Type-safe SQL

### Testing Tools
- **Kotest**: 5.9.x - Powerful testing framework
- **MockK**: 1.13.x - Mocking library for Kotlin
- **TestContainers**: 1.20.x - Integration testing
- **Espresso**: 3.6.x - Android UI testing
- **Compose Testing**: 1.7.0 - Compose UI testing

### Mobile Development
- **Compose Multiplatform**: 1.7.0
- **KMP libraries**: Ktor Client, SQLDelight, Voyager

## Ecosystem Overview

### Package Management

```bash
# Gradle Kotlin DSL project creation
gradle init --type kotlin-application --dsl kotlin --test-framework kotest --package com.example

# KMP project creation
curl -sSLO https://github.com/Kotlin/kmm-samples/archive/main.zip && unzip main.zip

# Dependency management with version catalogs
# gradle/libs.versions.toml
[versions]
kotlin = "2.0.20"
compose-bom = "2024.10.00"
ktor = "2.3.12"

[libraries]
kotlinx-coroutines-core = { module = "org.jetbrains.kotlinx:kotlinx-coroutines-core", version.ref = "kotlinx-coroutines" }
ktor-client-core = { module = "io.ktor:ktor-client-core", version.ref = "ktor" }

[bundles]
ktor = ["ktor-client-core", "ktor-client-logging"]
```

### Project Structure (2025 Best Practice)

```
kotlin-project/
├── gradle/
│   └── libs.versions.toml       # Version catalog
├── build.gradle.kts            # Root build configuration
├── settings.gradle.kts         # Gradle settings
├── gradlew                     # Gradle wrapper
└── src/
    ├── commonMain/kotlin/      # Shared KMP code
    │   └── com/example/
    │       ├── data/           # Shared data models
    │       ├── domain/         # Business logic
    │       └── network/        # API clients
    ├── androidMain/kotlin/     # Android-specific code
    │   └── com/example/
    │       ├── ui/             # Android UI
    │       ├── platform/       # Android implementations
    │       └── MainActivity.kt
    ├── iosMain/kotlin/         # iOS-specific code
    │   └── com/example/
    │       └── platform/       # iOS implementations
    ├── jvmMain/kotlin/         # JVM/Backend code
    │   └── com/example/
    │       ├── api/            # REST controllers
    │       ├── service/        # Business services
    │       └── Application.kt
    └── commonTest/kotlin/      # Shared tests
        └── com/example/
```

## Modern Development Patterns

### Kotlin Language Features 2.0

```kotlin
import kotlinx.coroutines.*
import kotlinx.coroutines.flow.*
import kotlin.coroutines.*
import kotlin.time.Duration.Companion.seconds

// Context receivers (Kotlin 2.0)
context(Logging)
suspend fun processUserData(userId: String) {
    log("Processing user: $userId")
    // Processing logic
}

// Sealed interfaces for better hierarchies
sealed interface NetworkResult<out T> {
    data class Success<T>(val data: T) : NetworkResult<T>
    data class Error(val exception: Throwable) : NetworkResult<Nothing>
    object Loading : NetworkResult<Nothing>
}

// Extension functions for better APIs
inline fun <T, R> Result<T>.flatMap(transform: (value: T) -> Result<R>): Result<R> {
    return fold(
        onSuccess = { success -> transform(success) },
        onFailure = { failure -> Result.failure(failure) }
    )
}

// Higher-order functions for functional programming
inline fun <T> List<T>.filterAndMap(
    crossinline predicate: (T) -> Boolean,
    crossinline transform: (T) -> T
): List<T> = this
    .filter { predicate(it) }
    .map { transform(it) }

// Scope functions usage
class UserRepository {
    private val users = mutableMapOf<String, User>()
    
    fun createUser(name: String, email: String): User {
        return User(name, email).also { user ->
            users[user.id] = user
            log("User created: ${user.id}")
        }
    }
    
    fun findUser(id: String): User? = users[id]?.takeIf { it.isActive }
}
```

### Coroutines and Flow Patterns

```kotlin
// Structured concurrency
class UserRepository(
    private val api: UserApi,
    private val database: UserDatabase
) {
    suspend fun getUserWithCache(id: String): User = coroutineScope {
        val cached = async(Dispatchers.IO) { database.getUser(id) }
        val network = async(Dispatchers.IO) { api.getUser(id) }
        
        when (val user = cached.await()) {
            null -> {
                val networkUser = network.await()
                database.saveUser(networkUser)
                networkUser
            }
            else -> {
                // Refresh in background
                launch { database.saveUser(network.await()) }
                user
            }
        }
    }
}

// Flow operators for reactive streams
class MessageRepository {
    private val _messages = MutableSharedFlow<Message>()
    val messages: SharedFlow<Message> = _messages.asSharedFlow()
    
    fun getFilteredMessages(filter: MessageFilter): Flow<Message> {
        return messages
            .filter { filter.matches(it) }
            .debounce(300.milliseconds)
            .distinctUntilChanged()
            .catch { error ->
                log("Error in message flow: ${error.message}")
                emit(Message.Error(error))
            }
    }
    
    // Cold flow with state
    fun getStateFlow(): StateFlow<UiState> = channelFlow {
        send(UiState.Loading)
        
        try {
            val data = fetchData()
            send(UiState.Success(data))
            
            // Continue listening for updates
            dataUpdates.collect { update ->
                send(UiState.Success(update))
            }
        } catch (e: Exception) {
            send(UiState.Error(e))
        }
    }.stateIn(
        scope = CoroutineScope(Dispatchers.Default + SupervisorJob()),
        started = SharingStarted.WhileSubscribed(5000),
        initialValue = UiState.Idle
    )
}

// Exception handling with coroutines
suspend fun safeApiCall(apiCall: suspend () -> User): Result<User> {
    return try {
        Result.success(apiCall())
    } catch (e: IOException) {
        Result.failure(NetworkException("Network error", e))
    } catch (e: HttpException) {
        Result.failure(ServerException("Server error: ${e.code()}", e))
    } catch (e: Exception) {
        Result.failure(UnknownException("Unknown error", e))
    }
}

// Timeout and retry patterns
suspend fun fetchWithRetryAndTimeout(
    maxRetries: Int = 3,
    timeout: Duration = 30.seconds
): Data {
    return retry(maxRetries) { attempt ->
        if (attempt > 0) delay(1000 * attempt) // Exponential backoff
        
        withTimeout(timeout) {
            fetchData()
        }
    }
}

// Coroutine scope management
class MyViewModel(
    private val repository: UserRepository
) : ViewModel() {
    
    private val _uiState = MutableStateFlow<UiState>(UiState.Idle)
    val uiState: StateFlow<UiState> = _uiState.asStateFlow()
    
    fun loadUser(userId: String) {
        viewModelScope.launch {
            _uiState.value = UiState.Loading
            
            repository.getUser(userId)
                .catch { error -> _uiState.value = UiState.Error(error) }
                .collect { user -> _uiState.value = UiState.Success(user) }
        }
    }
    
    override fun onCleared() {
        super.onCleared()
        viewModelScope.cancel() // Cancel all coroutines
    }
}
```

### Android Jetpack Compose Patterns

```kotlin
@Composable
fun UserListScreen(
    users: LazyPagingItems<User>,
    onUserClick: (User) -> Unit,
    modifier: Modifier = Modifier
) {
    LazyColumn(
        modifier = modifier.fillMaxSize(),
        contentPadding = PaddingValues(16.dp),
        verticalArrangement = Arrangement.spacedBy(8.dp)
    ) {
        items(
            count = users.itemCount,
            key = users.itemKey { it.id }
        ) { index ->
            val user = users[index]
            
            if (user != null) {
                UserItem(
                    user = user,
                    onClick = { onUserClick(user) }
                )
            } else {
                UserItemPlaceholder()
            }
        }
        
        when (users.loadState.append) {
            is LoadState.Loading -> item {
                LoadingIndicator()
            }
            is LoadState.Error -> item {
                ErrorMessage(
                    message = "Failed to load users",
                    onRetry = { users.retry() }
                )
            }
            else -> {}
        }
    }
}

@Composable
private fun UserItem(
    user: User,
    onClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    Card(
        onClick = onClick,
        modifier = modifier.fillMaxWidth(),
        elevation = CardDefaults.cardElevation(defaultElevation = 4.dp)
    ) {
        Row(
            modifier = Modifier
                .padding(16.dp)
                .fillMaxWidth(),
            verticalAlignment = Alignment.CenterVertically
        ) {
            AsyncImage(
                model = user.avatarUrl,
                contentDescription = "Avatar",
                modifier = Modifier
                    .size(48.dp)
                    .clip(CircleShape),
                contentScale = ContentScale.Crop
            )
            
            Spacer(modifier = Modifier.width(16.dp))
            
            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = user.name,
                    style = MaterialTheme.typography.titleMedium
                )
                Text(
                    text = user.email,
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }
            
            Icon(
                imageVector = Icons.Default.ChevronRight,
                contentDescription = null,
                tint = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }
    }
}

// State management with MVI
@Stable
sealed interface UserListUiState {
    object Idle : UserListUiState
    object Loading : UserListUiState
    data class Success(val users: LazyPagingItems<User>) : UserListUiState
    data class Error(val message: String) : UserListUiState
}

@Composable
fun rememberUserListUiState(
    viewModel: UserListViewModel = hiltViewModel()
): UserListUiState {
    val uiState by viewModel.uiState.collectAsState()
    return uiState
}

// Custom composables
@Composable
fun LoadingIndicator(
    modifier: Modifier = Modifier
) {
    Box(
        modifier = modifier.fillMaxWidth(),
        contentAlignment = Alignment.Center
    ) {
        CircularProgressIndicator()
    }
}

@Composable
fun ErrorMessage(
    message: String,
    onRetry: () -> Unit,
    modifier: Modifier = Modifier
) {
    Column(
        modifier = modifier
            .fillMaxWidth()
            .padding(16.dp),
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        Text(
            text = message,
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.error,
            textAlign = TextAlign.Center
        )
        
        Spacer(modifier = Modifier.height(8.dp))
        
        Button(onClick = onRetry) {
            Text("Retry")
        }
    }
}
```

### Spring Boot with Kotlin Patterns

```kotlin
@SpringBootApplication
@EnableTransactionManagement
class Application {
    @Bean
    fun objectMapper(): ObjectMapper = jacksonObjectMapper()
        .registerModules(JavaTimeModule(), KotlinModule.Builder().build())
        .disable(SerializationFeature.WRITE_DATES_AS_TIMESTAMPS)
}

@Configuration
@EnableWebFluxSecurity
class SecurityConfig {
    
    @Bean
    fun securityWebFilterChain(
        http: ServerHttpSecurity,
        jwtDecoder: ReactiveJwtDecoder
    ): SecurityWebFilterChain {
        return http
            .csrf { it.disable() }
            .authorizeExchange { exchanges ->
                exchanges
                    .pathMatchers("/api/auth/**").permitAll()
                    .pathMatchers("/api/public/**").permitAll()
                    .pathMatchers("/actuator/health").permitAll()
                    .pathMatchers("/api/admin/**").hasRole("ADMIN")
                    .anyExchange().authenticated()
            }
            .oauth2ResourceServer { oauth2 ->
                oauth2.jwt { jwt -> jwt.jwtDecoder(jwtDecoder) }
            }
            .build()
    }
}

// Repository with Coroutines
@Repository
interface UserRepository : ReactiveCrudRepository<User, Long> {
    suspend fun findByUsername(username: String): User?
    fun findByEmailContaining(email: String): Flow<User>
    fun findByActiveTrue(): Flow<User>
}

@Service
@Transactional
class UserService(
    private val userRepository: UserRepository,
    private val eventPublisher: ApplicationEventPublisher
) {
    suspend fun createUser(request: CreateUserRequest): User {
        return userEntity.toUser().also { user ->
            userRepository.save(userEntity)
            eventPublisher.publishEvent(UserCreatedEvent(user.id))
        }
    }
    
    fun searchUsers(query: String): Flow<User> {
        return userRepository.findByEmailContaining(query)
            .flowOn(Dispatchers.IO)
            .catch { error ->
                log.error("Error searching users", error)
                throw UserSearchException("Failed to search users", error)
            }
    }
}

// REST Controller with Coroutines
@RestController
@RequestMapping("/api/users")
@Validated
class UserController(
    private val userService: UserService
) {
    @GetMapping("/{id}")
    suspend fun getUser(@PathVariable id: Long): ResponseEntity<UserDto> {
        val user = userService.getUserById(id)
            ?: throw UserNotFoundException("User not found: $id")
        
        return ResponseEntity.ok(user.toDto())
    }
    
    @GetMapping
    fun searchUsers(
        @RequestParam query: String?,
        @RequestParam(defaultValue = "0") page: Int,
        @RequestParam(defaultValue = "20") size: Int
    ): Flow<UserDto> {
        return userService.searchUsers(query ?: "")
            .map { it.toDto() }
            .take(size.toLong())
            .skip((page * size).toLong())
    }
    
    @PostMapping
    @ResponseStatus(HttpStatus.CREATED)
    suspend fun createUser(
        @Valid @RequestBody request: CreateUserRequest
    ): UserDto {
        val user = userService.createUser(request)
        return user.toDto()
    }
}

// Domain models with data classes
@Document
data class User(
    @Id val id: String? = null,
    val username: String,
    val email: String,
    val profile: UserProfile,
    val preferences: UserPreferences,
    val createdAt: Instant = Instant.now(),
    val updatedAt: Instant = Instant.now(),
    val active: Boolean = true
)

@Document
data class UserProfile(
    val firstName: String,
    val lastName: String,
    val avatar: String? = null,
    val bio: String? = null
)

// Extension functions for DTOs
fun User.toDto() = UserDto(
    id = id!!,
    username = username,
    email = email,
    firstName = profile.firstName,
    lastName = profile.lastName,
    avatar = profile.avatar,
    bio = profile.bio,
    active = active
)
```

## Performance Considerations

### Memory Optimization

```kotlin
// Efficient collections usage
class PerformanceOptimized {
    
    // Use primitive collections for performance
    private val userIdSet = Object2IntOpenHashMap<User>()
    private val activeUsers = BooleanArrayList()
    
    // Lazy initialization with delegation
    private val expensiveResource: ExpensiveResource by lazy {
        createExpensiveResource()
    }
    
    // Efficient string processing
    fun processLargeText(text: String): List<String> {
        return text.lineSequence()
            .filter { it.isNotBlank() }
            .map { it.trim() }
            .toList()
    }
    
    // Efficient data processing with sequences
    fun processUsers(users: List<User>): List<UserDto> {
        return users.asSequence()
            .filter { it.active }
            .sortedBy { it.username }
            .map { it.toDto() }
            .toList()
    }
}

// Coroutine performance patterns
class PerformanceService {
    
    // Limit concurrency
    private val dispatcher = Dispatchers.IO.limitedParallelism(10)
    
    suspend fun processItems(items: List<Item>): List<Result> {
        return coroutineScope {
            items.map { item ->
                async(dispatcher) {
                    processItem(item)
                }
            }.awaitAll()
        }
    }
    
    // Efficient flow processing
    fun processLargeStream(): Flow<Result> {
        return generateSequence { generateNextItem() }
            .asFlow()
            .flowOn(Dispatchers.Default)
            .buffer(Channel.CONFLATED) // Keep only latest
            .conflate() // Skip intermediate values if consumer is slow
    }
}
```

### Kotlin/Native and KMP Performance

```kotlin
// Expect/actual for platform-specific optimizations
expect class PlatformLogger() {
    fun log(message: String)
    fun logError(error: Throwable)
}

// JVM implementation
actual class PlatformLogger actual constructor() {
    private val logger = LoggerFactory.getLogger(PlatformLogger::class.java)
    
    actual fun log(message: String) {
        logger.info(message)
    }
    
    actual fun logError(error: Throwable) {
        logger.error("Error occurred", error)
    }
}

// Native implementation
actual class PlatformLogger actual constructor() {
    actual fun log(message: String) {
        println("[INFO] $message")
    }
    
    actual fun logError(error: Throwable) {
        println("[ERROR] ${error.message}")
    }
}

// Memory management in KMP
class SharedCache<K, V> {
    private val cache = mutableMapOf<K, V>()
    private val maxSize: Int
    
    constructor(maxSize: Int = 100) {
        this.maxSize = maxSize
    }
    
    fun get(key: K): V? = cache[key]
    
    fun put(key: K, value: V) {
        if (cache.size >= maxSize) {
            val oldestKey = cache.keys.first()
            cache.remove(oldestKey)
        }
        cache[key] = value
    }
    
    fun clear() {
        cache.clear()
    }
}
```

### Profiling and Monitoring

```bash
# Kotlin coroutines debugging
# Add to build.gradle.kts
dependencies {
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-debug")
}

# Enable coroutine debugging
# VM options: -Dkotlinx.coroutines.debug=on

# Android Studio profiler
# CPU Profiler, Memory Profiler, Network Profiler

# JVM debugging with jstat/jmap
jstat -gc <pid> 1000 10
jmap -dump:format=b,file=heap.hprof <pid>

# Gradle build performance
./gradlew build --scan
./gradlew build --profile
```

## Testing Strategy

### Kotest Configuration

```kotlin
// build.gradle.kts
tasks.test {
    useTestNG() // or useJUnitPlatform()
    
    // Configure test execution
    testLogging {
        events("passed", "skipped", "failed")
        showExceptions = true
        showCauses = true
        showStackTraces = true
    }
    
    // Parallel test execution
    maxParallelForks = (Runtime.getRuntime().availableProcessors() / 2).takeIf { it > 0 } ?: 1
}

// Test configuration
@TestConfiguration
class TestConfig {
    @Bean
    @Primary
    fun userRepository(): UserRepository = mockk()
}

// Kotest test classes
class UserRepositoryTest : FunSpec({
    val repository = mockk<UserRepository>()
    val service = UserService(repository)
    
    test("should create user successfully") {
        // Given
        val request = CreateUserRequest("testuser", "test@example.com")
        val expectedUser = User("1", "testuser", "test@example.com")
        
        coEvery { repository.findByUsername("testuser") } returns null
        coEvery { repository.save(any()) } returns expectedUser
        
        // When
        val result = service.createUser(request)
        
        // Then
        result shouldBe expectedUser
        coVerify { repository.save(any()) }
    }
    
    context("user search") {
        test("should return users matching query") {
            // Given
            val users = flowOf(
                User("1", "john@example.com"),
                User("2", "jane@example.com")
            )
            every { repository.findByEmailContaining("john") } returns users
            
            // When
            val results = service.searchUsers("john").toList()
            
            // Then
            results shouldHaveSize 1
            results.first().email shouldBe "john@example.com"
        }
    }
})
```

### Android Testing with Compose

```kotlin
@RunWith(AndroidJUnit4::class)
class UserProfileScreenTest {
    
    @get:Rule
    val composeTestRule = createComposeRule()
    
    @Test
    fun userProfileScreen_displaysUserInformation() {
        // Given
        val user = User(
            id = "1",
            name = "John Doe",
            email = "john@example.com",
            avatar = "https://example.com/avatar.jpg"
        )
        
        // When
        composeTestRule.setContent {
            UserProfileScreen(
                user = user,
                onEditClick = {}
            )
        }
        
        // Then
        composeTestRule
            .onNodeWithText("John Doe")
            .assertIsDisplayed()
        
        composeTestRule
            .onNodeWithText("john@example.com")
            .assertIsDisplayed()
        
        composeTestRule
            .onNodeWithContentDescription("User avatar")
            .assertIsDisplayed()
    }
    
    @Test
    fun userProfileScreen_whenEditClicked_callsOnEditClick() {
        // Given
        var editClicked = false
        val user = User("1", "John Doe", "john@example.com")
        
        // When
        composeTestRule.setContent {
            UserProfileScreen(
                user = user,
                onEditClick = { editClicked = true }
            )
        }
        
        composeTestRule
            .onNodeWithText("Edit")
            .performClick()
        
        // Then
        assertTrue(editClicked)
    }
}

// ViewModel testing
@ExtendWith(MockitoExtension::class)
class UserListViewModelTest {
    
    @Mock
    private lateinit var userRepository: UserRepository
    
    private lateinit var viewModel: UserListViewModel
    
    @Before
    fun setup() {
        viewModel = UserListViewModel(userRepository)
    }
    
    @Test
    fun `loadUsers should update UI state with users`() = runTest {
        // Given
        val users = flowOf(
            User("1", "User 1"),
            User("2", "User 2")
        )
        whenever(userRepository.getUsers()).thenReturn(users)
        
        // When
        viewModel.loadUsers()
        
        // Then
        val state = viewModel.uiState.value
        assertTrue(state is UserListUiState.Success)
        assertEquals(2, (state as UserListUiState.Success).users.count())
    }
}
```

### Integration Testing

```kotlin
@SpringBootTest
@Testcontainers
@Transactional
class UserControllerIntegrationTest {
    
    companion object {
        @Container
        @JvmStatic
        val postgres = PostgreSQLContainer<Nothing>("postgres:16-alpine")
            .apply {
                withDatabaseName("testdb")
                withUsername("test")
                withPassword("test")
            }
        
        @JvmStatic
        @DynamicPropertySource
        fun postgresProperties(registry: DynamicPropertyRegistry) {
            registry.add("spring.r2dbc.url") { 
                "r2dbc:postgresql://${postgres.host}:${postgres.firstMappedPort}/testdb" 
            }
            registry.add("spring.r2dbc.username") { postgres.username }
            registry.add("spring.r2dbc.password") { postgres.password }
        }
    }
    
    @Autowired
    private lateinit var webTestClient: WebTestClient
    
    @Test
    fun `should create and retrieve user`() {
        // Given
        val createUserRequest = CreateUserRequest(
            username = "testuser",
            email = "test@example.com"
        )
        
        // When
        val createResponse = webTestClient.post()
            .uri("/api/users")
            .bodyValue(createUserRequest)
            .exchange()
            .expectStatus().isCreated
            .expectBody(UserDto::class.java)
            .returnResult()
            .responseBody!!
        
        // Then
        val retrievedUser = webTestClient.get()
            .uri("/api/users/${createResponse.id}")
            .exchange()
            .expectStatus().isOk
            .expectBody(UserDto::class.java)
            .returnResult()
            .responseBody!!
        
        assertEquals(createResponse.id, retrievedUser.id)
        assertEquals(createResponse.username, retrievedUser.username)
    }
}
```

## Security Best Practices

### Input Validation and Sanitization

```kotlin
// Data class with validation
data class CreateUserRequest(
    @field:NotBlank(message = "Username is required")
    @field:Size(min = 3, max = 50, message = "Username must be between 3 and 50 characters")
    @field:Pattern(regexp = "^[a-zA-Z0-9_]+$", message = "Username can only contain letters, numbers, and underscores")
    val username: String,
    
    @field:NotBlank(message = "Email is required")
    @field:Email(message = "Invalid email format")
    @field:Size(max = 100, message = "Email must be less than 100 characters")
    val email: String,
    
    @field:NotBlank(message = "Password is required")
    @field:Size(min = 8, max = 128, message = "Password must be between 8 and 128 characters")
    @field:Pattern(
        regexp = "^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d)(?=.*[@$!%*?&])[A-Za-z\\d@$!%*?&]",
        message = "Password must contain at least one lowercase letter, one uppercase letter, one digit, and one special character"
    )
    val password: String
) {
    init {
        require(!email.lowercase().contains("@admin.com")) {
            "Admin email domains are not allowed"
        }
    }
}

// Security service
@Service
class SecurityService {
    private val passwordEncoder = BCryptPasswordEncoder(12)
    
    fun encodePassword(rawPassword: String): String {
        return passwordEncoder.encode(rawPassword)
    }
    
    fun matches(rawPassword: String, encodedPassword: String): Boolean {
        return passwordEncoder.matches(rawPassword, encodedPassword)
    }
    
    fun validateInput(input: String): String {
        return input
            .replace(Regex("<script.*?>.*?</script>"), "")
            .replace(Regex("javascript:"), "")
            .trim()
    }
}
```

### Authentication and Authorization

```kotlin
@Configuration
@EnableWebFluxSecurity
class SecurityConfig {
    
    @Bean
    fun securityWebFilterChain(http: ServerHttpSecurity): SecurityWebFilterChain {
        return http
            .csrf { it.disable() }
            .authorizeExchange { exchanges ->
                exchanges
                    .pathMatchers("/api/auth/**").permitAll()
                    .pathMatchers("/api/public/**").permitAll()
                    .pathMatchers("/actuator/health").permitAll()
                    .pathMatchers("/api/admin/**").hasRole("ADMIN")
                    .anyExchange().authenticated()
            }
            .oauth2ResourceServer { oauth2 ->
                oauth2.jwt { jwt -> jwt.jwtDecoder(jwtDecoder()) }
            }
            .build()
    }
    
    @Bean
    fun reactiveAuthenticationManager(
        userRepository: UserRepository,
        passwordService: PasswordService
    ): ReactiveAuthenticationManager {
        return UserDetailsRepositoryReactiveAuthenticationManager(
            userRepository.toUserDetailsService()
        ).apply {
            setPasswordEncoder(passwordService.encoder)
        }
    }
}

// Method-level security
@Service
class DocumentService {
    
    @PreAuthorize("hasRole('USER') and @documentSecurity.canRead(#documentId, authentication.name)")
    suspend fun getDocument(documentId: Long): Document {
        return documentRepository.findById(documentId)
            ?: throw DocumentNotFoundException("Document not found: $documentId")
    }
    
    @PreAuthorize("hasRole('ADMIN') or @documentSecurity.isOwner(#documentId, authentication.name)")
    suspend fun updateDocument(documentId: Long, updateDto: DocumentUpdateDto): Document {
        val document = getDocument(documentId)
        if (!canUpdate(document, SecurityContextHolder.getContext().authentication.name)) {
            throw AccessDeniedException("Not authorized to update document")
        }
        return documentRepository.save(document.update(updateDto))
    }
}
```

## Integration Patterns

### Database Integration with Exposed

```kotlin
object Users : Table("users") {
    val id = uuid("id").autoGenerate()
    val username = varchar("username", 50).uniqueIndex()
    val email = varchar("email", 100).uniqueIndex()
    val createdAt = datetime("created_at").defaultExpression(CurrentDateTime())
    val updatedAt = datetime("updated_at").defaultExpression(CurrentDateTime())
    val active = bool("active").default(true)
    
    override val primaryKey = PrimaryKey(id)
}

@Entity
data class User(
    val id: UUID = UUID.randomUUID(),
    val username: String,
    val email: String,
    val createdAt: Instant = Instant.now(),
    val updatedAt: Instant = Instant.now(),
    val active: Boolean = true
)

// Repository with Exposed
class UserRepository(
    private val database: Database
) {
    suspend fun findById(id: UUID): User? = withContext(Dispatchers.IO) {
        transaction(database) {
            Users.select { Users.id eq id }
                .map { it.toUser() }
                .singleOrNull()
        }
    }
    
    suspend fun save(user: User): User = withContext(Dispatchers.IO) {
        transaction(database) {
            Users.upsert {
                it[id] = user.id
                it[username] = user.username
                it[email] = user.email
                it[createdAt] = user.createdAt
                it[updatedAt] = Instant.now()
                it[active] = user.active
            }
            user.copy(updatedAt = Instant.now())
        }
    }
    
    fun findAllActive(): Flow<User> = flow {
        transaction(database) {
            Users.select { Users.active eq true }
                .map { it.toUser() }
                .forEach { user -> emit(user) }
        }
    }.flowOn(Dispatchers.IO)
}

private fun ResultRow.toUser() = User(
    id = this[Users.id],
    username = this[Users.username],
    email = this[Users.email],
    createdAt = this[Users.createdAt].toInstant(),
    updatedAt = this[Users.updatedAt].toInstant(),
    active = this[Users.active]
)
```

### Ktor Client Integration

```kotlin
// Ktor client configuration
class ApiClient {
    private val httpClient = HttpClient(CIO) {
        install(ContentNegotiation) {
            json(
                Json {
                    ignoreUnknownKeys = true
                    isLenient = true
                }
            )
        }
        
        install(Auth) {
            bearer {
                loadTokens {
                    BearerTokens(accessToken = getAccessToken(), refreshToken = getRefreshToken())
                }
                refreshTokens {
                    val refreshedTokens = refreshAccessToken()
                    BearerTokens(
                        accessToken = refreshedTokens.accessToken,
                        refreshToken = refreshedTokens.refreshToken
                    )
                }
            }
        }
        
        defaultRequest {
            url("https://api.example.com/")
        }
    }
    
    suspend fun getUsers(): List<User> = try {
        httpClient.get("users").body()
    } catch (e: ClientRequestException) {
        when (e.response.status) {
            HttpStatusCode.NotFound -> emptyList()
            HttpStatusCode.Unauthorized -> throw AuthenticationException()
            else -> throw ApiException("Failed to fetch users", e)
        }
    } catch (e: Exception) {
        throw NetworkException("Network error", e)
    }
    
    suspend fun createUser(user: CreateUserRequest): User = httpClient.post("users") {
        contentType(ContentType.Application.Json)
        setBody(user)
    }.body()
}

// Multiplatform API client
class SharedApiClient(
    private val platform: Platform
) {
    private val httpClient = HttpClient {
        if (platform.isDebug) {
            install(Logging) {
                logger = Logger.DEFAULT
                level = LogLevel.INFO
            }
        }
        
        expectSuccess = true
        install(ContentNegotiation) {
            json(Json {
                ignoreUnknownKeys = true
                useAlternativeNames = false
            })
        }
    }
    
    suspend fun fetchData(): ApiResponse = httpClient.get("data").body()
}

// Platform-specific implementations
// JVM
class AndroidPlatform : Platform {
    override val name: String = "Android ${Build.VERSION.SDK_INT}"
    override val isDebug: Boolean = BuildConfig.DEBUG
}

// iOS
class IOSPlatform : Platform {
    override val name: String = UIDevice.currentDevice.systemName()
    override val isDebug: Boolean = Platform.isDebugBinary
}

actual fun getPlatform(): Platform = 
    if (Platform.osFamily == OsFamily.IOS) IOSPlatform() else AndroidPlatform()
```

### Message Queue Integration

```kotlin
// RabbitMQ with Spring Boot
@Configuration
class RabbitConfig {
    
    @Bean
    fun userQueue(): Queue = QueueBuilder.durable("user.queue").build()
    
    @Bean
    fun userExchange(): TopicExchange = TopicExchange("user.exchange")
    
    @Bean
    fun userBinding(): Binding = BindingBuilder
        .bind(userQueue())
        .to(userExchange())
        .with("user.created")
}

@Service
class UserEventPublisher(
    private val rabbitTemplate: RabbitTemplate
) {
    fun publishUserCreated(user: User) {
        val event = UserCreatedEvent(
            userId = user.id,
            username = user.username,
            email = user.email,
            timestamp = Instant.now()
        )
        
        rabbitTemplate.convertAndSend(
            "user.exchange",
            "user.created",
            event
        )
    }
}

@Component
class UserEventConsumer {
    
    private val logger = LoggerFactory.getLogger(UserEventConsumer::class.java)
    
    @RabbitListener(queues = ["user.queue"])
    fun handleUserCreated(event: UserCreatedEvent) {
        logger.info("Received user created event: $event")
        
        // Process event
        sendWelcomeEmail(event.userId, event.email)
        createAuditRecord(event)
    }
    
    private fun sendWelcomeEmail(userId: UUID, email: String) {
        // Implementation
    }
    
    private fun createAuditRecord(event: UserCreatedEvent) {
        // Implementation
    }
}
```

## Modern Development Workflow

### Gradle Configuration

```kotlin
// build.gradle.kts
plugins {
    kotlin("multiplatform") version "2.0.20"
    kotlin("plugin.serialization") version "2.0.20"
    application
    id("org.springframework.boot") version "3.3.0"
    id("io.spring.dependency-management") version "1.1.5"
    id("com.google.cloud.tools.jib") version "3.4.1"
}

kotlin {
    jvmToolchain(21)
    
    jvm {
        withJava()
        testRuns.named("test") {
            executionTask.configure {
                useJUnitPlatform()
            }
        }
    }
    
    js(IR) {
        browser {
            commonWebpackConfig {
                cssSupport {
                    enabled.set(true)
                }
            }
        }
    }
    
    listOf(
        iosX64(),
        iosArm64(),
        iosSimulatorArm64()
    ).forEach { iosTarget ->
        iosTarget.binaries.framework {
            baseName = "Shared"
            isStatic = true
        }
    }
    
    sourceSets {
        commonMain.dependencies {
            implementation(kotlin("stdlib"))
            implementation("org.jetbrains.kotlinx:kotlinx-coroutines-core:1.8.0")
            implementation("org.jetbrains.kotlinx:kotlinx-serialization-json:1.7.0")
            implementation("io.ktor:ktor-client-core:2.3.12")
            implementation("io.ktor:ktor-client-content-negotiation:2.3.12")
            implementation("io.ktor:ktor-serialization-kotlinx-json:2.3.12")
        }
        
        commonTest.dependencies {
            implementation(kotlin("test"))
            implementation(kotlin("test-coroutines"))
        }
        
        jvmMain.dependencies {
            implementation("org.springframework.boot:spring-boot-starter-webflux")
            implementation("org.springframework.boot:spring-boot-starter-data-r2dbc")
            implementation("org.springframework.boot:spring-boot-starter-amqp")
            implementation("io.ktor:ktor-client-okhttp:2.3.12")
        }
        
        androidMain.dependencies {
            implementation("io.ktor:ktor-client-android:2.3.12")
        }
        
        iosMain.dependencies {
            implementation("io.ktor:ktor-client-darwin:2.3.12")
        }
    }
}

application {
    mainClass.set("com.example.ApplicationKt")
}

tasks.withType<KotlinCompile> {
    kotlinOptions {
        jvmTarget = "21"
        freeCompilerArgs += listOf(
            "-Xjsr305=strict",
            "-Xjvm-default=all",
            "-opt-in=kotlin.RequiresOptIn"
        )
    }
}

// Jib configuration for containerization
jib {
    from {
        image = "eclipse-temurin:21-jre-alpine"
    }
    to {
        image = "my-registry.com/kotlin-app:latest"
    }
    container {
        jvmFlags = listOf("-Xms512m", "-Xmx2g")
        ports = listOf("8080")
        environment = mapOf(
            "SPRING_PROFILES_ACTIVE" to "prod"
        )
    }
}
```

### Pre-commit Configuration

```yaml
# .pre-commit-config.yaml
repos:
  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v5.0.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-yaml
      - id: check-added-large-files
      - id: check-merge-conflict

  - repo: https://github.com/pinterest/ktlint
    rev: 1.2.1
    hooks:
      - id: ktlint
        args: ["--android", "--color"]

  - repo: local
    hooks:
      - id: gradle-detekt
        name: Static analysis with Detekt
        entry: ./gradlew detekt
        language: system
        files: '\.(kt|kts)$'
        pass_filenames: false

      - id: gradle-test
        name: Run tests
        entry: ./gradlew test
        language: system
        pass_filenames: false
        always_run: true

      - id: gradle-check
        name: Run all checks
        entry: ./gradlew check
        language: system
        pass_filenames: false
        always_run: true
```

### Docker Configuration

```dockerfile
# Multi-stage Docker build
FROM eclipse-temurin:21-jdk-alpine AS builder

WORKDIR /app

# Copy gradle wrapper and cache dependencies
COPY gradlew gradlew.bat gradle/ ./
COPY build.gradle.kts settings.gradle.kts ./

RUN ./gradlew --no-daemon dependencies

# Copy source code
COPY src/ ./src/

# Build application
RUN ./gradlew --no-daemon bootJar -x test

# Runtime stage
FROM eclipse-temurin:21-jre-alpine

RUN addgroup --system --gid 1001 appgroup && \
    adduser --system --uid 1001 --ingroup appgroup appuser && \
    mkdir -p /app && \
    chown -R appuser:appgroup /app

WORKDIR /app

COPY --from=builder /app/build/libs/*.jar app.jar

RUN chmod +x app.jar

USER appuser

EXPOSE 8080

ENV JAVA_OPTS="-Xms512m -Xmx2g -XX:+UseG1GC -XX:+UseStringDeduplication"

ENTRYPOINT ["sh", "-c", "java $JAVA_OPTS -jar app.jar"]
```

## Cross-Platform Development Patterns

### KMP Shared Business Logic

```kotlin
// commonMain/kotlin/com/example/data/models/User.kt
@Serializable
data class User(
    val id: String,
    val username: String,
    val email: String,
    val avatar: String? = null,
    val createdAt: Instant = Clock.System.now()
)

// commonMain/kotlin/com/example/data/repository/UserRepository.kt
interface UserRepository {
    suspend fun getUsers(): Result<List<User>>
    suspend fun getUserById(id: String): Result<User>
    suspend fun createUser(user: User): Result<User>
}

// commonMain/kotlin/com/example/data/service/UserService.kt
class UserService(
    private val repository: UserRepository,
    private val logger: Logger
) {
    suspend fun getAllUsers(): Result<List<User>> {
        return try {
            val users = repository.getUsers().getOrElse { error ->
                logger.error("Failed to fetch users", error)
                return Result.failure(error)
            }
            Result.success(users)
        } catch (e: Exception) {
            logger.error("Unexpected error fetching users", e)
            Result.failure(e)
        }
    }
    
    fun getUsersFlow(): Flow<List<User>> = flow {
        while (currentCoroutineContext()[Job]?.isActive == true) {
            val result = getAllUsers()
            if (result.isSuccess) {
                emit(result.getOrNull() ?: emptyList())
            }
            delay(30.seconds) // Refresh every 30 seconds
        }
    }
}

// Platform-specific implementations
// androidMain/kotlin/com/example/data/repository/AndroidUserRepository.kt
class AndroidUserRepository(
    private val context: Context
) : UserRepository {
    
    private val httpClient = HttpClient(OkHttp) {
        install(ContentNegotiation) {
            json(Json {
                ignoreUnknownKeys = true
            })
        }
    }
    
    override suspend fun getUsers(): Result<List<User>> {
        return try {
            val response = httpClient.get("https://api.example.com/users")
            val users = response.body<List<User>>()
            Result.success(users)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    override suspend fun getUserById(id: String): Result<User> {
        // Implementation
    }
    
    override suspend fun createUser(user: User): Result<User> {
        // Implementation
    }
}

// iosMain/kotlin/com/example/data/repository/IosUserRepository.kt
class IosUserRepository : UserRepository {
    
    private val httpClient = HttpClient(Darwin) {
        install(ContentNegotiation) {
            json(Json {
                ignoreUnknownKeys = true
            })
        }
    }
    
    override suspend fun getUsers(): Result<List<User>> {
        return try {
            val response = httpClient.get("https://api.example.com/users")
            val users = response.body<List<User>>()
            Result.success(users)
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    override suspend fun getUserById(id: String): Result<User> {
        // Implementation
    }
    
    override suspend fun createUser(user: User): Result<User> {
        // Implementation
    }
}
```

---

**Created by**: MoAI Language Skill Factory  
**Last Updated**: 2025-11-06  
**Version**: 2.0.0  
**Kotlin Target**: 2.0.20 with modern coroutines, KMP, and Jetpack Compose  

This skill provides comprehensive Kotlin development guidance with 2025 best practices, covering everything from Android development with Compose to backend services with Spring Boot and cross-platform KMP development.
