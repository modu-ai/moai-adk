# Android Native Development

Kotlin, Jetpack Compose, and modern Android patterns for native Android development.

## Quick Reference

Key Technologies:
- **Language**: Kotlin with Coroutines, Flow, and Kotlin Multiplatform (optional)
- **UI framework**: Jetpack Compose (primary), View system (interop)
- **Architecture**: MVVM + Clean Architecture with Hilt DI
- **Data**: Room (SQLite), DataStore (preferences), WorkManager (background tasks)
- **Minimum target**: Android 7.0 (API 24) recommended; Android 8.0 (API 26) for modern features

---

## Core Patterns

### App Entry Point

```kotlin
@HiltAndroidApp
class MyApplication : Application()

@AndroidEntryPoint
class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContent {
            MyAppTheme {
                AppNavHost()
            }
        }
    }
}
```

### Navigation with Compose Navigation

```kotlin
@Composable
fun AppNavHost(
    navController: NavHostController = rememberNavController()
) {
    NavHost(navController, startDestination = "home") {
        composable("home") { HomeScreen(navController) }
        composable("detail/{id}") { backStackEntry ->
            val id = backStackEntry.arguments?.getString("id") ?: return@composable
            DetailScreen(id = id, navController = navController)
        }
    }
}
```

### ViewModel with StateFlow

```kotlin
@HiltViewModel
class UserViewModel @Inject constructor(
    private val userRepository: UserRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow(UserUiState())
    val uiState: StateFlow<UserUiState> = _uiState.asStateFlow()

    fun loadUsers() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true) }
            userRepository.getUsers()
                .onSuccess { users ->
                    _uiState.update { it.copy(users = users, isLoading = false) }
                }
                .onFailure { error ->
                    _uiState.update { it.copy(error = error.message, isLoading = false) }
                }
        }
    }
}

// Usage in Composable
@Composable
fun UserListScreen(viewModel: UserViewModel = hiltViewModel()) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()

    LaunchedEffect(Unit) { viewModel.loadUsers() }

    if (uiState.isLoading) CircularProgressIndicator()
    else LazyColumn {
        items(uiState.users) { user -> UserItem(user) }
    }
}
```

### Repository Pattern with Coroutines

```kotlin
interface UserRepository {
    suspend fun getUsers(): Result<List<User>>
}

@Singleton
class UserRepositoryImpl @Inject constructor(
    private val api: UserApi,
    private val dao: UserDao
) : UserRepository {
    override suspend fun getUsers(): Result<List<User>> = runCatching {
        val remote = api.fetchUsers()
        dao.insertAll(remote.map { it.toEntity() })
        remote
    }.recoverCatching {
        // Fallback to cache on network failure
        dao.getAll().map { it.toDomain() }
    }
}
```

---

## Works Well With

- expert-mobile (primary agent)
- moai-domain-mobile/modules/strategy-comparison.md (paradigm selection)
- expert-security (ProGuard/R8 rules, biometric auth, EncryptedSharedPreferences)
