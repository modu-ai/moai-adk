# Kotlin Production Examples

## Complete REST API with Ktor

### Application Setup

```kotlin
// Application.kt
fun main() {
    embeddedServer(Netty, port = 8080, host = "0.0.0.0") {
        configureKoin()
        configureSecurity()
        configureRouting()
        configureContentNegotiation()
        configureStatusPages()
        configureMonitoring()
    }.start(wait = true)
}

fun Application.configureKoin() {
    install(Koin) {
        modules(appModule)
    }
}

val appModule = module {
    single<Database> { DatabaseFactory.create() }
    single<UserRepository> { UserRepositoryImpl(get()) }
    single<OrderRepository> { OrderRepositoryImpl(get()) }
    single<UserService> { UserServiceImpl(get(), get()) }
    single<JwtService> { JwtServiceImpl() }
}

fun Application.configureSecurity() {
    install(Authentication) {
        jwt("auth-jwt") {
            realm = "User API"
            verifier(JwtConfig.verifier)
            validate { credential ->
                if (credential.payload.audience.contains("api"))
                    JWTPrincipal(credential.payload)
                else null
            }
            challenge { _, _ ->
                call.respond(HttpStatusCode.Unauthorized, ErrorResponse("Token invalid or expired"))
            }
        }
    }
}

fun Application.configureContentNegotiation() {
    install(ContentNegotiation) {
        json(Json {
            prettyPrint = true
            isLenient = true
            ignoreUnknownKeys = true
            encodeDefaults = true
        })
    }
}

fun Application.configureStatusPages() {
    install(StatusPages) {
        exception<ValidationException> { call, cause ->
            call.respond(HttpStatusCode.BadRequest, ErrorResponse(cause.message ?: "Validation failed"))
        }
        exception<NotFoundException> { call, cause ->
            call.respond(HttpStatusCode.NotFound, ErrorResponse(cause.message ?: "Resource not found"))
        }
        exception<DuplicateException> { call, cause ->
            call.respond(HttpStatusCode.Conflict, ErrorResponse(cause.message ?: "Resource already exists"))
        }
        exception<AuthenticationException> { call, cause ->
            call.respond(HttpStatusCode.Unauthorized, ErrorResponse(cause.message ?: "Authentication failed"))
        }
        exception<Throwable> { call, cause ->
            call.application.log.error("Unhandled exception", cause)
            call.respond(HttpStatusCode.InternalServerError, ErrorResponse("Internal server error"))
        }
    }
}

fun Application.configureMonitoring() {
    install(CallLogging) {
        level = Level.INFO
        filter { call -> call.request.path().startsWith("/api") }
        format { call ->
            val status = call.response.status()
            val method = call.request.httpMethod.value
            val path = call.request.path()
            val duration = call.processingTimeMillis()
            "$method $path - $status (${duration}ms)"
        }
    }
}
```

### Complete User Routes

```kotlin
// UserRoutes.kt
fun Application.configureRouting() {
    val userService by inject<UserService>()
    val jwtService by inject<JwtService>()

    routing {
        route("/api/v1") {
            // Health check
            get("/health") {
                call.respond(mapOf("status" to "healthy", "timestamp" to Instant.now().toString()))
            }

            // Public authentication routes
            route("/auth") {
                post("/register") {
                    val request = call.receive<CreateUserRequest>()
                    request.validate()
                    val user = userService.create(request)
                    call.respond(HttpStatusCode.Created, user.toDto())
                }

                post("/login") {
                    val request = call.receive<LoginRequest>()
                    val user = userService.authenticate(request.email, request.password)
                    val token = jwtService.generateToken(user)
                    call.respond(TokenResponse(token, user.toDto()))
                }

                post("/refresh") {
                    val request = call.receive<RefreshRequest>()
                    val newToken = jwtService.refreshToken(request.token)
                    call.respond(TokenResponse(newToken, null))
                }
            }

            // Protected user routes
            authenticate("auth-jwt") {
                route("/users") {
                    get {
                        val page = call.parameters["page"]?.toIntOrNull() ?: 0
                        val size = call.parameters["size"]?.toIntOrNull()?.coerceIn(1, 100) ?: 20
                        val search = call.parameters["search"]

                        val users = if (search != null) {
                            userService.search(search, page, size)
                        } else {
                            userService.findAll(page, size)
                        }
                        call.respond(users.map { it.toDto() })
                    }

                    get("/{id}") {
                        val id = call.parameters["id"]?.toLongOrNull()
                            ?: throw ValidationException("Invalid user ID")
                        val user = userService.findById(id)
                            ?: throw NotFoundException("User not found")
                        call.respond(user.toDto())
                    }

                    get("/me") {
                        val userId = call.principal<JWTPrincipal>()!!
                            .payload.getClaim("userId").asLong()
                        val user = userService.findById(userId)
                            ?: throw NotFoundException("User not found")
                        call.respond(user.toDto())
                    }

                    put("/{id}") {
                        val id = call.parameters["id"]?.toLongOrNull()
                            ?: throw ValidationException("Invalid user ID")
                        val request = call.receive<UpdateUserRequest>()
                        request.validate()
                        val user = userService.update(id, request)
                            ?: throw NotFoundException("User not found")
                        call.respond(user.toDto())
                    }

                    delete("/{id}") {
                        val id = call.parameters["id"]?.toLongOrNull()
                            ?: throw ValidationException("Invalid user ID")
                        if (userService.delete(id)) {
                            call.respond(HttpStatusCode.NoContent)
                        } else {
                            throw NotFoundException("User not found")
                        }
                    }

                    // User's orders
                    get("/{id}/orders") {
                        val id = call.parameters["id"]?.toLongOrNull()
                            ?: throw ValidationException("Invalid user ID")
                        val orders = userService.getUserOrders(id)
                        call.respond(orders.map { it.toDto() })
                    }
                }
            }
        }
    }
}
```

### Service Layer with Coroutines

```kotlin
// UserService.kt
interface UserService {
    suspend fun findAll(page: Int, size: Int): List<User>
    suspend fun findById(id: Long): User?
    suspend fun search(query: String, page: Int, size: Int): List<User>
    suspend fun create(request: CreateUserRequest): User
    suspend fun update(id: Long, request: UpdateUserRequest): User?
    suspend fun delete(id: Long): Boolean
    suspend fun authenticate(email: String, password: String): User
    suspend fun getUserOrders(userId: Long): List<Order>
}

class UserServiceImpl(
    private val userRepository: UserRepository,
    private val orderRepository: OrderRepository
) : UserService {

    override suspend fun findAll(page: Int, size: Int): List<User> =
        userRepository.findAll(page * size, size)

    override suspend fun findById(id: Long): User? =
        userRepository.findById(id)

    override suspend fun search(query: String, page: Int, size: Int): List<User> =
        userRepository.search(query, page * size, size)

    override suspend fun create(request: CreateUserRequest): User {
        if (userRepository.existsByEmail(request.email)) {
            throw DuplicateException("Email already registered")
        }

        val user = User(
            id = 0,
            name = request.name,
            email = request.email,
            passwordHash = BCrypt.hashpw(request.password, BCrypt.gensalt()),
            status = UserStatus.PENDING,
            createdAt = Instant.now()
        )
        return userRepository.save(user)
    }

    override suspend fun update(id: Long, request: UpdateUserRequest): User? {
        val existing = userRepository.findById(id) ?: return null

        if (request.email != null && request.email != existing.email) {
            if (userRepository.existsByEmail(request.email)) {
                throw DuplicateException("Email already registered")
            }
        }

        val updated = existing.copy(
            name = request.name ?: existing.name,
            email = request.email ?: existing.email,
            updatedAt = Instant.now()
        )
        return userRepository.update(updated)
    }

    override suspend fun delete(id: Long): Boolean =
        userRepository.delete(id)

    override suspend fun authenticate(email: String, password: String): User {
        val user = userRepository.findByEmail(email)
            ?: throw AuthenticationException("Invalid credentials")

        if (!BCrypt.checkpw(password, user.passwordHash)) {
            throw AuthenticationException("Invalid credentials")
        }

        if (user.status != UserStatus.ACTIVE) {
            throw AuthenticationException("Account not active")
        }

        return user
    }

    override suspend fun getUserOrders(userId: Long): List<Order> = coroutineScope {
        // Verify user exists
        val user = async { userRepository.findById(userId) }
        val orders = async { orderRepository.findByUserId(userId) }

        if (user.await() == null) {
            throw NotFoundException("User not found")
        }
        orders.await()
    }
}
```

### Repository with Exposed

```kotlin
// UserRepository.kt
interface UserRepository {
    suspend fun findAll(offset: Int, limit: Int): List<User>
    suspend fun findById(id: Long): User?
    suspend fun findByEmail(email: String): User?
    suspend fun search(query: String, offset: Int, limit: Int): List<User>
    suspend fun existsByEmail(email: String): Boolean
    suspend fun save(user: User): User
    suspend fun update(user: User): User
    suspend fun delete(id: Long): Boolean
}

class UserRepositoryImpl(private val database: Database) : UserRepository {

    override suspend fun findAll(offset: Int, limit: Int): List<User> = dbQuery {
        UserEntity.all()
            .orderBy(Users.createdAt to SortOrder.DESC)
            .limit(limit, offset.toLong())
            .map { it.toModel() }
    }

    override suspend fun findById(id: Long): User? = dbQuery {
        UserEntity.findById(id)?.toModel()
    }

    override suspend fun findByEmail(email: String): User? = dbQuery {
        UserEntity.find { Users.email eq email }.singleOrNull()?.toModel()
    }

    override suspend fun search(query: String, offset: Int, limit: Int): List<User> = dbQuery {
        UserEntity.find {
            (Users.name.lowerCase() like "%${query.lowercase()}%") or
            (Users.email.lowerCase() like "%${query.lowercase()}%")
        }
        .orderBy(Users.createdAt to SortOrder.DESC)
        .limit(limit, offset.toLong())
        .map { it.toModel() }
    }

    override suspend fun existsByEmail(email: String): Boolean = dbQuery {
        UserEntity.find { Users.email eq email }.count() > 0
    }

    override suspend fun save(user: User): User = dbQuery {
        UserEntity.new {
            name = user.name
            email = user.email
            passwordHash = user.passwordHash
            status = user.status
        }.toModel()
    }

    override suspend fun update(user: User): User = dbQuery {
        val entity = UserEntity.findById(user.id)
            ?: throw NotFoundException("User not found")
        entity.name = user.name
        entity.email = user.email
        entity.updatedAt = Instant.now()
        entity.toModel()
    }

    override suspend fun delete(id: Long): Boolean = dbQuery {
        val entity = UserEntity.findById(id) ?: return@dbQuery false
        entity.delete()
        true
    }

    private suspend fun <T> dbQuery(block: suspend () -> T): T =
        newSuspendedTransaction(Dispatchers.IO, database) { block() }
}
```

---

## Android with Jetpack Compose

### Complete Screen with ViewModel

```kotlin
// UserListScreen.kt
@Composable
fun UserListScreen(
    viewModel: UserListViewModel = hiltViewModel(),
    onUserClick: (Long) -> Unit,
    onAddUserClick: () -> Unit
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    val snackbarHostState = remember { SnackbarHostState() }

    LaunchedEffect(Unit) {
        viewModel.events.collect { event ->
            when (event) {
                is UserListEvent.ShowError -> {
                    snackbarHostState.showSnackbar(event.message)
                }
                is UserListEvent.UserDeleted -> {
                    snackbarHostState.showSnackbar("User deleted successfully")
                }
            }
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("Users") },
                actions = {
                    IconButton(onClick = { viewModel.refresh() }) {
                        Icon(Icons.Default.Refresh, contentDescription = "Refresh")
                    }
                }
            )
        },
        floatingActionButton = {
            FloatingActionButton(onClick = onAddUserClick) {
                Icon(Icons.Default.Add, contentDescription = "Add User")
            }
        },
        snackbarHost = { SnackbarHost(snackbarHostState) }
    ) { padding ->
        Box(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
        ) {
            when (val state = uiState) {
                is UserListUiState.Loading -> {
                    CircularProgressIndicator(
                        modifier = Modifier.align(Alignment.Center)
                    )
                }
                is UserListUiState.Success -> {
                    if (state.users.isEmpty()) {
                        EmptyState(
                            message = "No users found",
                            onAddClick = onAddUserClick
                        )
                    } else {
                        UserList(
                            users = state.users,
                            onUserClick = onUserClick,
                            onDeleteClick = { viewModel.deleteUser(it) }
                        )
                    }
                }
                is UserListUiState.Error -> {
                    ErrorState(
                        message = state.message,
                        onRetryClick = { viewModel.retry() }
                    )
                }
            }

            // Pull to refresh
            PullToRefreshContainer(
                state = rememberPullToRefreshState(),
                isRefreshing = uiState is UserListUiState.Loading,
                onRefresh = { viewModel.refresh() }
            )
        }
    }
}

@Composable
fun UserList(
    users: List<User>,
    onUserClick: (Long) -> Unit,
    onDeleteClick: (Long) -> Unit
) {
    LazyColumn(
        contentPadding = PaddingValues(16.dp),
        verticalArrangement = Arrangement.spacedBy(8.dp)
    ) {
        items(users, key = { it.id }) { user ->
            UserListItem(
                user = user,
                onClick = { onUserClick(user.id) },
                onDelete = { onDeleteClick(user.id) },
                modifier = Modifier.animateItem()
            )
        }
    }
}

@Composable
fun UserListItem(
    user: User,
    onClick: () -> Unit,
    onDelete: () -> Unit,
    modifier: Modifier = Modifier
) {
    var showDeleteDialog by remember { mutableStateOf(false) }

    if (showDeleteDialog) {
        AlertDialog(
            onDismissRequest = { showDeleteDialog = false },
            title = { Text("Delete User") },
            text = { Text("Are you sure you want to delete ${user.name}?") },
            confirmButton = {
                TextButton(onClick = { onDelete(); showDeleteDialog = false }) {
                    Text("Delete", color = MaterialTheme.colorScheme.error)
                }
            },
            dismissButton = {
                TextButton(onClick = { showDeleteDialog = false }) {
                    Text("Cancel")
                }
            }
        )
    }

    Card(
        modifier = modifier
            .fillMaxWidth()
            .clickable(onClick = onClick),
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Row(
            modifier = Modifier.padding(16.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            AsyncImage(
                model = ImageRequest.Builder(LocalContext.current)
                    .data(user.avatarUrl)
                    .crossfade(true)
                    .placeholder(R.drawable.avatar_placeholder)
                    .build(),
                contentDescription = user.name,
                modifier = Modifier
                    .size(48.dp)
                    .clip(CircleShape)
            )
            Spacer(Modifier.width(16.dp))
            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = user.name,
                    style = MaterialTheme.typography.titleMedium
                )
                Text(
                    text = user.email,
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }
            IconButton(onClick = { showDeleteDialog = true }) {
                Icon(
                    Icons.Default.Delete,
                    contentDescription = "Delete",
                    tint = MaterialTheme.colorScheme.error
                )
            }
        }
    }
}
```

### ViewModel with StateFlow

```kotlin
// UserListViewModel.kt
@HiltViewModel
class UserListViewModel @Inject constructor(
    private val userRepository: UserRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow<UserListUiState>(UserListUiState.Loading)
    val uiState: StateFlow<UserListUiState> = _uiState.asStateFlow()

    private val _events = MutableSharedFlow<UserListEvent>()
    val events: SharedFlow<UserListEvent> = _events.asSharedFlow()

    init {
        loadUsers()
    }

    fun loadUsers() {
        viewModelScope.launch {
            _uiState.value = UserListUiState.Loading
            userRepository.getUsers()
                .catch { e ->
                    _uiState.value = UserListUiState.Error(e.message ?: "Unknown error")
                }
                .collect { users ->
                    _uiState.value = UserListUiState.Success(users)
                }
        }
    }

    fun refresh() = loadUsers()
    fun retry() = loadUsers()

    fun deleteUser(id: Long) {
        viewModelScope.launch {
            userRepository.deleteUser(id)
                .onSuccess {
                    _events.emit(UserListEvent.UserDeleted(id))
                    loadUsers()
                }
                .onFailure { e ->
                    _events.emit(UserListEvent.ShowError(e.message ?: "Failed to delete user"))
                }
        }
    }
}

sealed interface UserListUiState {
    data object Loading : UserListUiState
    data class Success(val users: List<User>) : UserListUiState
    data class Error(val message: String) : UserListUiState
}

sealed interface UserListEvent {
    data class ShowError(val message: String) : UserListEvent
    data class UserDeleted(val id: Long) : UserListEvent
}
```

### Repository with Room and Retrofit

```kotlin
// UserRepository.kt
interface UserRepository {
    fun getUsers(): Flow<List<User>>
    suspend fun getUser(id: Long): Result<User>
    suspend fun deleteUser(id: Long): Result<Unit>
    suspend fun createUser(request: CreateUserRequest): Result<User>
    suspend fun syncUsers(): Result<Unit>
}

class UserRepositoryImpl @Inject constructor(
    private val api: UserApi,
    private val dao: UserDao,
    @IoDispatcher private val ioDispatcher: CoroutineDispatcher
) : UserRepository {

    override fun getUsers(): Flow<List<User>> =
        dao.getAllUsers()
            .onStart { syncUsersFromNetwork() }
            .flowOn(ioDispatcher)

    private suspend fun syncUsersFromNetwork() {
        try {
            val users = api.getUsers()
            dao.insertAll(users.map { it.toEntity() })
        } catch (e: Exception) {
            // Log error, continue with cached data
            Timber.e(e, "Failed to sync users from network")
        }
    }

    override suspend fun getUser(id: Long): Result<User> = withContext(ioDispatcher) {
        runCatching {
            dao.getUser(id)
                ?: api.getUser(id).also { dao.insert(it.toEntity()) }.toDomain()
        }
    }

    override suspend fun deleteUser(id: Long): Result<Unit> = withContext(ioDispatcher) {
        runCatching {
            api.deleteUser(id)
            dao.delete(id)
        }
    }

    override suspend fun createUser(request: CreateUserRequest): Result<User> =
        withContext(ioDispatcher) {
            runCatching {
                val user = api.createUser(request)
                dao.insert(user.toEntity())
                user.toDomain()
            }
        }

    override suspend fun syncUsers(): Result<Unit> = withContext(ioDispatcher) {
        runCatching {
            val users = api.getUsers()
            dao.deleteAll()
            dao.insertAll(users.map { it.toEntity() })
        }
    }
}
```

---

## Compose Multiplatform

### Shared UI Module

```kotlin
// commonMain/UserScreen.kt
@Composable
fun UserScreen(
    viewModel: UserViewModel,
    modifier: Modifier = Modifier
) {
    val uiState by viewModel.uiState.collectAsState()

    Column(modifier = modifier.fillMaxSize()) {
        // Search bar
        SearchBar(
            query = uiState.searchQuery,
            onQueryChange = viewModel::onSearchQueryChange,
            modifier = Modifier.fillMaxWidth().padding(16.dp)
        )

        // User list
        when (val state = uiState.listState) {
            is ListState.Loading -> LoadingIndicator()
            is ListState.Success -> UserList(
                users = state.users,
                onUserClick = viewModel::onUserClick
            )
            is ListState.Error -> ErrorMessage(
                message = state.message,
                onRetry = viewModel::retry
            )
        }
    }
}

@Composable
expect fun SearchBar(
    query: String,
    onQueryChange: (String) -> Unit,
    modifier: Modifier = Modifier
)

// androidMain/SearchBar.kt
@Composable
actual fun SearchBar(
    query: String,
    onQueryChange: (String) -> Unit,
    modifier: Modifier
) {
    OutlinedTextField(
        value = query,
        onValueChange = onQueryChange,
        modifier = modifier,
        placeholder = { Text("Search users...") },
        leadingIcon = { Icon(Icons.Default.Search, null) },
        singleLine = true
    )
}

// desktopMain/SearchBar.kt
@Composable
actual fun SearchBar(
    query: String,
    onQueryChange: (String) -> Unit,
    modifier: Modifier
) {
    TextField(
        value = query,
        onValueChange = onQueryChange,
        modifier = modifier,
        placeholder = { Text("Search users...") },
        leadingIcon = { Icon(Icons.Default.Search, null) }
    )
}
```

### Shared ViewModel

```kotlin
// commonMain/UserViewModel.kt
class UserViewModel(
    private val userRepository: UserRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow(UserUiState())
    val uiState: StateFlow<UserUiState> = _uiState.asStateFlow()

    private val searchQuery = MutableStateFlow("")

    init {
        loadUsers()
        observeSearch()
    }

    private fun loadUsers() {
        viewModelScope.launch {
            _uiState.update { it.copy(listState = ListState.Loading) }
            userRepository.getUsers()
                .catch { e ->
                    _uiState.update { it.copy(listState = ListState.Error(e.message ?: "Error")) }
                }
                .collect { users ->
                    _uiState.update { it.copy(listState = ListState.Success(users)) }
                }
        }
    }

    private fun observeSearch() {
        viewModelScope.launch {
            searchQuery
                .debounce(300)
                .distinctUntilChanged()
                .collectLatest { query ->
                    if (query.isBlank()) {
                        loadUsers()
                    } else {
                        searchUsers(query)
                    }
                }
        }
    }

    private suspend fun searchUsers(query: String) {
        _uiState.update { it.copy(listState = ListState.Loading) }
        userRepository.searchUsers(query)
            .catch { e ->
                _uiState.update { it.copy(listState = ListState.Error(e.message ?: "Error")) }
            }
            .collect { users ->
                _uiState.update { it.copy(listState = ListState.Success(users)) }
            }
    }

    fun onSearchQueryChange(query: String) {
        _uiState.update { it.copy(searchQuery = query) }
        searchQuery.value = query
    }

    fun onUserClick(userId: Long) {
        // Handle navigation
    }

    fun retry() = loadUsers()
}

data class UserUiState(
    val searchQuery: String = "",
    val listState: ListState = ListState.Loading
)

sealed interface ListState {
    data object Loading : ListState
    data class Success(val users: List<User>) : ListState
    data class Error(val message: String) : ListState
}
```

---

## Build Configuration

### Gradle Kotlin DSL

```kotlin
// build.gradle.kts
plugins {
    kotlin("jvm") version "2.0.20"
    kotlin("plugin.serialization") version "2.0.20"
    id("io.ktor.plugin") version "3.0.0"
    id("com.google.devtools.ksp") version "2.0.20-1.0.24"
}

group = "com.example"
version = "1.0.0"

kotlin {
    jvmToolchain(21)
}

application {
    mainClass.set("com.example.ApplicationKt")
}

ktor {
    fatJar {
        archiveFileName.set("app.jar")
    }
}

dependencies {
    // Ktor Server
    implementation("io.ktor:ktor-server-core-jvm")
    implementation("io.ktor:ktor-server-netty-jvm")
    implementation("io.ktor:ktor-server-content-negotiation-jvm")
    implementation("io.ktor:ktor-serialization-kotlinx-json-jvm")
    implementation("io.ktor:ktor-server-auth-jvm")
    implementation("io.ktor:ktor-server-auth-jwt-jvm")
    implementation("io.ktor:ktor-server-status-pages-jvm")
    implementation("io.ktor:ktor-server-call-logging-jvm")

    // Coroutines
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-core:1.9.0")

    // Koin DI
    implementation("io.insert-koin:koin-ktor:3.5.6")
    implementation("io.insert-koin:koin-logger-slf4j:3.5.6")

    // Exposed
    implementation("org.jetbrains.exposed:exposed-core:0.55.0")
    implementation("org.jetbrains.exposed:exposed-dao:0.55.0")
    implementation("org.jetbrains.exposed:exposed-jdbc:0.55.0")
    implementation("org.jetbrains.exposed:exposed-kotlin-datetime:0.55.0")

    // Database
    implementation("org.postgresql:postgresql:42.7.3")
    implementation("com.zaxxer:HikariCP:5.1.0")

    // Logging
    implementation("ch.qos.logback:logback-classic:1.5.6")

    // Testing
    testImplementation("io.ktor:ktor-server-test-host-jvm")
    testImplementation("org.jetbrains.kotlin:kotlin-test-junit5")
    testImplementation("io.mockk:mockk:1.13.12")
    testImplementation("org.jetbrains.kotlinx:kotlinx-coroutines-test:1.9.0")
    testImplementation("app.cash.turbine:turbine:1.1.0")
    testImplementation("org.testcontainers:postgresql:1.20.1")
}

tasks.test {
    useJUnitPlatform()
}
```

---

## Models and DTOs

```kotlin
// Models.kt
@Serializable
data class User(
    val id: Long = 0,
    val name: String,
    val email: String,
    val passwordHash: String,
    val status: UserStatus,
    @Serializable(with = InstantSerializer::class)
    val createdAt: Instant = Instant.now(),
    @Serializable(with = InstantSerializer::class)
    val updatedAt: Instant? = null
) {
    fun toDto() = UserDto(id, name, email, status, createdAt)
}

@Serializable
data class UserDto(
    val id: Long,
    val name: String,
    val email: String,
    val status: UserStatus,
    @Serializable(with = InstantSerializer::class)
    val createdAt: Instant
)

@Serializable
data class CreateUserRequest(
    val name: String,
    val email: String,
    val password: String
) {
    fun validate() {
        require(name.isNotBlank() && name.length in 2..100) { "Name must be 2-100 characters" }
        require(email.matches(Regex("^[\\w-.]+@[\\w-]+\\.[a-z]{2,}$"))) { "Invalid email format" }
        require(password.length >= 8) { "Password must be at least 8 characters" }
    }
}

@Serializable
data class UpdateUserRequest(
    val name: String? = null,
    val email: String? = null
) {
    fun validate() {
        name?.let { require(it.isNotBlank() && it.length in 2..100) { "Name must be 2-100 characters" } }
        email?.let { require(it.matches(Regex("^[\\w-.]+@[\\w-]+\\.[a-z]{2,}$"))) { "Invalid email format" } }
    }
}

@Serializable
data class LoginRequest(
    val email: String,
    val password: String
)

@Serializable
data class TokenResponse(
    val token: String,
    val user: UserDto?
)

@Serializable
data class ErrorResponse(
    val message: String,
    val code: String? = null
)

@Serializable
enum class UserStatus {
    PENDING, ACTIVE, SUSPENDED
}

// Exceptions
class ValidationException(message: String) : RuntimeException(message)
class NotFoundException(message: String) : RuntimeException(message)
class DuplicateException(message: String) : RuntimeException(message)
class AuthenticationException(message: String) : RuntimeException(message)
```

---

Last Updated: 2025-12-07
Version: 1.0.0
