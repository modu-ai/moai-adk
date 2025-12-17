---
name: moai-lang-kotlin
description: Kotlin 2.0+ development specialist covering Ktor, coroutines, Compose Multiplatform, and Kotlin-idiomatic patterns. Use when building Kotlin server apps, Android apps, or multiplatform projects.
version: 1.0.0
category: language
tags: kotlin, ktor, coroutines, compose, android, multiplatform
context7-libraries: /ktorio/ktor, /jetbrains/compose-multiplatform, /jetbrains/exposed
related-skills: moai-lang-java, moai-lang-swift
updated: 2025-12-07
status: active
---

## Quick Reference (30 seconds)

Kotlin 2.0+ Expert - K2 compiler, coroutines, Ktor, Compose Multiplatform with Context7 integration.

Auto-Triggers: Kotlin files (`.kt`, `.kts`), Gradle Kotlin DSL (`build.gradle.kts`, `settings.gradle.kts`)

Core Capabilities:
- Kotlin 2.0: K2 compiler, coroutines, Flow, sealed classes, value classes
- Ktor 3.0: Async HTTP server/client, WebSocket, JWT authentication
- Exposed 0.55: Kotlin SQL framework with coroutines support
- Spring Boot (Kotlin): Kotlin-idiomatic Spring with WebFlux
- Compose Multiplatform: Desktop, iOS, Web, Android UI
- Testing: JUnit 5, MockK, Kotest, Turbine for Flow testing

---

## Implementation Guide (5 minutes)

### Kotlin 2.0 Features

Coroutines and Flow:
```kotlin
// Structured concurrency with async/await
suspend fun fetchUserWithOrders(userId: Long): UserWithOrders = coroutineScope {
    val userDeferred = async { userRepository.findById(userId) }
    val ordersDeferred = async { orderRepository.findByUserId(userId) }
    UserWithOrders(userDeferred.await(), ordersDeferred.await())
}

// Flow for reactive streams
fun observeUsers(): Flow<User> = flow {
    while (true) {
        emit(userRepository.findLatest())
        delay(1000)
    }
}.flowOn(Dispatchers.IO)
```

Sealed Classes and Value Classes:
```kotlin
sealed interface Result<out T> {
    data class Success<T>(val data: T) : Result<T>
    data class Error(val exception: Throwable) : Result<Nothing>
    data object Loading : Result<Nothing>
}

@JvmInline
value class UserId(val value: Long) {
    init { require(value > 0) { "UserId must be positive" } }
}

@JvmInline
value class Email(val value: String) {
    init { require(value.contains("@")) { "Invalid email format" } }
}
```

### Ktor 3.0 Server

Application Setup:
```kotlin
fun main() {
    embeddedServer(Netty, port = 8080, host = "0.0.0.0") {
        configureKoin()
        configureSecurity()
        configureRouting()
        configureContentNegotiation()
    }.start(wait = true)
}

fun Application.configureKoin() {
    install(Koin) { modules(appModule) }
}

val appModule = module {
    single<Database> { DatabaseFactory.create() }
    single<UserRepository> { UserRepositoryImpl(get()) }
    single<UserService> { UserServiceImpl(get()) }
}

fun Application.configureSecurity() {
    install(Authentication) {
        jwt("auth-jwt") {
            realm = "User API"
            verifier(JwtConfig.verifier)
            validate { credential ->
                if (credential.payload.audience.contains("api"))
                    JWTPrincipal(credential.payload) else null
            }
        }
    }
}

fun Application.configureContentNegotiation() {
    install(ContentNegotiation) {
        json(Json { prettyPrint = true; ignoreUnknownKeys = true })
    }
}
```

Routing with Authentication:
```kotlin
fun Application.configureRouting() {
    val userService by inject<UserService>()

    routing {
        route("/api/v1") {
            post("/auth/register") {
                val request = call.receive<CreateUserRequest>()
                val user = userService.create(request)
                call.respond(HttpStatusCode.Created, user.toDto())
            }

            authenticate("auth-jwt") {
                route("/users") {
                    get {
                        val page = call.parameters["page"]?.toIntOrNull() ?: 0
                        val size = call.parameters["size"]?.toIntOrNull() ?: 20
                        call.respond(userService.findAll(page, size).map { it.toDto() })
                    }

                    get("/{id}") {
                        val id = call.parameters["id"]?.toLongOrNull()
                            ?: return@get call.respond(HttpStatusCode.BadRequest)
                        userService.findById(id)?.let { call.respond(it.toDto()) }
                            ?: call.respond(HttpStatusCode.NotFound)
                    }
                }
            }
        }
    }
}
```

### Exposed SQL Framework

Table and Entity:
```kotlin
object Users : LongIdTable("users") {
    val name = varchar("name", 100)
    val email = varchar("email", 255).uniqueIndex()
    val passwordHash = varchar("password_hash", 255)
    val status = enumerationByName<UserStatus>("status", 20)
    val createdAt = timestamp("created_at").defaultExpression(CurrentTimestamp())
}

class UserEntity(id: EntityID<Long>) : LongEntity(id) {
    companion object : LongEntityClass<UserEntity>(Users)
    var name by Users.name
    var email by Users.email
    var passwordHash by Users.passwordHash
    var status by Users.status
    var createdAt by Users.createdAt

    fun toModel() = User(id.value, name, email, passwordHash, status, createdAt)
}
```

Repository with Coroutines:
```kotlin
class UserRepositoryImpl(private val database: Database) : UserRepository {
    override suspend fun findById(id: Long): User? = dbQuery {
        UserEntity.findById(id)?.toModel()
    }

    override suspend fun save(user: User): User = dbQuery {
        UserEntity.new {
            name = user.name
            email = user.email
            passwordHash = user.passwordHash
            status = user.status
        }.toModel()
    }

    private suspend fun <T> dbQuery(block: suspend () -> T): T =
        newSuspendedTransaction(Dispatchers.IO, database) { block() }
}
```

### Spring Boot with Kotlin

WebFlux Controller:
```kotlin
@RestController
@RequestMapping("/api/users")
class UserController(private val userService: UserService) {

    @GetMapping
    suspend fun listUsers(
        @RequestParam(defaultValue = "0") page: Int,
        @RequestParam(defaultValue = "20") size: Int
    ): Flow<UserDto> = userService.findAll(page, size).map { it.toDto() }

    @GetMapping("/{id}")
    suspend fun getUser(@PathVariable id: Long): ResponseEntity<UserDto> =
        userService.findById(id)?.let { ResponseEntity.ok(it.toDto()) }
            ?: ResponseEntity.notFound().build()

    @PostMapping
    suspend fun createUser(@Valid @RequestBody request: CreateUserRequest): ResponseEntity<UserDto> {
        val user = userService.create(request)
        return ResponseEntity.created(URI.create("/api/users/${user.id}")).body(user.toDto())
    }
}
```

---

## Advanced Patterns

### Compose Multiplatform

Shared UI Component:
```kotlin
@Composable
fun UserListScreen(viewModel: UserListViewModel, onUserClick: (Long) -> Unit) {
    val uiState by viewModel.uiState.collectAsState()

    when (val state = uiState) {
        is UiState.Loading -> LoadingIndicator()
        is UiState.Success -> UserList(state.users, onUserClick)
        is UiState.Error -> ErrorMessage(state.message, onRetry = viewModel::retry)
    }
}

@Composable
fun UserCard(user: User, onClick: () -> Unit) {
    Card(
        modifier = Modifier.fillMaxWidth().clickable(onClick = onClick),
        elevation = CardDefaults.cardElevation(4.dp)
    ) {
        Row(modifier = Modifier.padding(16.dp)) {
            AsyncImage(model = user.avatarUrl, contentDescription = user.name,
                modifier = Modifier.size(48.dp).clip(CircleShape))
            Spacer(Modifier.width(16.dp))
            Column {
                Text(user.name, style = MaterialTheme.typography.titleMedium)
                Text(user.email, style = MaterialTheme.typography.bodySmall)
            }
        }
    }
}
```

### Testing with MockK

```kotlin
class UserServiceTest {
    private val userRepository = mockk<UserRepository>()
    private val userService = UserService(userRepository)

    @Test
    fun `should fetch user concurrently`() = runTest {
        val testUser = User(1L, "John", "john@example.com")
        coEvery { userRepository.findById(1L) } coAnswers { delay(100); testUser }

        val result = userService.findById(1L)
        assertThat(result).isEqualTo(testUser)
    }

    @Test
    fun `should handle Flow emissions`() = runTest {
        val users = listOf(User(1L, "John", "john@example.com"))
        coEvery { userRepository.findAllAsFlow() } returns users.asFlow()

        userService.streamUsers().toList().also { result ->
            assertThat(result).hasSize(1)
        }
    }
}
```

### Gradle Build Configuration

```kotlin
plugins {
    kotlin("jvm") version "2.0.20"
    kotlin("plugin.serialization") version "2.0.20"
    id("io.ktor.plugin") version "3.0.0"
}

kotlin { jvmToolchain(21) }

dependencies {
    implementation("io.ktor:ktor-server-core-jvm")
    implementation("io.ktor:ktor-server-netty-jvm")
    implementation("io.ktor:ktor-server-content-negotiation-jvm")
    implementation("io.ktor:ktor-server-auth-jwt-jvm")
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-core:1.9.0")
    implementation("org.jetbrains.exposed:exposed-core:0.55.0")
    implementation("org.jetbrains.exposed:exposed-dao:0.55.0")
    implementation("org.postgresql:postgresql:42.7.3")

    testImplementation("io.mockk:mockk:1.13.12")
    testImplementation("org.jetbrains.kotlinx:kotlinx-coroutines-test:1.9.0")
    testImplementation("app.cash.turbine:turbine:1.1.0")
}
```

---

## Context7 Integration

Library mappings for latest documentation:
- `/ktorio/ktor` - Ktor 3.0 server/client documentation
- `/jetbrains/exposed` - Exposed SQL framework
- `/JetBrains/kotlin` - Kotlin 2.0 language reference
- `/Kotlin/kotlinx.coroutines` - Coroutines library
- `/jetbrains/compose-multiplatform` - Compose Multiplatform
- `/arrow-kt/arrow` - Arrow functional programming

Usage:
```python
docs = await mcp__context7__get_library_docs(
    context7CompatibleLibraryID="/ktorio/ktor",
    topic="routing authentication jwt",
    tokens=8000
)
```

---

## When to Use Kotlin

Use Kotlin When:
- Developing Android applications (official language)
- Building modern server applications with Ktor
- Preferring concise, expressive syntax
- Building reactive services with coroutines and Flow
- Creating multiplatform applications (iOS, Desktop, Web)
- Full Java interoperability required

Consider Alternatives When:
- Legacy Java codebase requiring minimal changes
- Big data pipelines (prefer Scala with Spark)

---

## Works Well With

- `moai-lang-java` - Java interoperability and Spring Boot patterns
- `moai-domain-backend` - REST API, GraphQL, microservices architecture
- `moai-domain-database` - JPA, Exposed, R2DBC patterns
- `moai-quality-testing` - JUnit 5, MockK, TestContainers integration
- `moai-infra-docker` - JVM container optimization

---

## Troubleshooting

K2 Compiler: Add `kotlin.experimental.tryK2=true` to gradle.properties; clear `.gradle` for rebuild

Coroutines: Avoid `runBlocking` in suspend contexts; use `Dispatchers.IO` for blocking operations

Ktor: Ensure `ContentNegotiation` installed; check JWT verifier configuration; verify routing hierarchy

Exposed: Ensure all DB operations in transaction; be aware of entity loading outside transaction

---

## Advanced Documentation

For comprehensive reference materials:
- [reference.md](reference.md) - Complete ecosystem, Context7 mappings, testing patterns, performance
- [examples.md](examples.md) - Production-ready code examples, Ktor, Compose, Android patterns

---

Last Updated: 2025-12-07
Status: Production Ready (v1.0.0)
