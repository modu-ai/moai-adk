# Mobile Development Reference

## Platform Version Matrix

### Swift 6.0 (iOS 18+, macOS 15+)
- Release: September 2025
- Xcode: 16.0+
- Minimum Deployment: iOS 15.0+ (recommended iOS 17.0+)
- Key Features:
  - Complete data-race safety by default
  - Typed throws for precise error handling
  - Custom actor executors for concurrency control
  - Embedded Swift for IoT and embedded systems
  - Improved C++ interoperability

### Kotlin 2.0 (Android 14+)
- Release: May 2024
- Android Studio: Koala (2024.1.1)+
- Minimum SDK: 24 (recommended 26+)
- Key Features:
  - K2 compiler with 2x faster compilation
  - Context receivers for dependency injection
  - Data objects for singleton patterns
  - Improved smart casts
  - Better multiplatform support

### Dart 3.5 / Flutter 3.24
- Dart Release: November 2025
- Flutter Release: November 2025
- Minimum OS: iOS 12.0+, Android API 21+
- Key Features:
  - Pattern matching with exhaustiveness checking
  - Records and destructuring
  - Sealed classes and interfaces
  - Macros (experimental)
  - Enhanced extension types

## Context7 Library Mappings

### Swift/iOS Libraries
```
/apple/swift                    - Swift language and standard library
/apple/swift-package-manager    - SwiftPM package management
/apple/swift-nio                - Non-blocking I/O framework
/Alamofire/Alamofire            - HTTP networking library
/onevcat/Kingfisher             - Image downloading and caching
/realm/realm-swift              - Mobile database
/SnapKit/SnapKit                - Auto Layout DSL
/ReactiveX/RxSwift              - Reactive programming
/pointfreeco/swift-composable-architecture - TCA architecture
/Quick/Quick                    - BDD testing framework
/Quick/Nimble                   - Matcher framework
```

### Kotlin/Android Libraries
```
/JetBrains/kotlin               - Kotlin language
/android/architecture-components-samples - Jetpack samples
/google/dagger                  - Dependency injection (Hilt)
/square/retrofit                - HTTP client
/coil-kt/coil                   - Image loading
/cashapp/sqldelight             - Type-safe SQL
/InsertKoinIO/koin              - Dependency injection
/Kotlin/kotlinx.coroutines      - Coroutines
/mockk/mockk                    - Mocking library
/google/truth                   - Assertion library
```

### Flutter/Dart Libraries
```
/flutter/flutter                - Flutter framework
/dart-lang/sdk                  - Dart SDK
/rrousselGit/riverpod           - State management
/felangel/bloc                  - BLoC pattern
/fluttercommunity/get_it        - Service locator
/cfug/dio                       - HTTP client
/isar/isar                      - NoSQL database
/Sub6Resources/flutter_html     - HTML rendering
/abuanwar072/Flutter-Responsive-Admin-Panel-or-Dashboard - UI patterns
```

## Architecture Patterns

### iOS Architecture (SwiftUI + TCA)

The Composable Architecture (TCA):
```swift
// Feature.swift
@Reducer
struct UserFeature {
    @ObservableState
    struct State: Equatable {
        var user: User?
        var isLoading = false
        var error: String?
    }

    enum Action: BindableAction {
        case binding(BindingAction<State>)
        case loadUser(String)
        case userLoaded(Result<User, Error>)
        case logout
    }

    @Dependency(\.userClient) var userClient
    @Dependency(\.mainQueue) var mainQueue

    var body: some ReducerOf<Self> {
        BindingReducer()
        Reduce { state, action in
            switch action {
            case .binding:
                return .none

            case let .loadUser(id):
                state.isLoading = true
                state.error = nil
                return .run { send in
                    await send(.userLoaded(
                        Result { try await userClient.fetch(id) }
                    ))
                }

            case let .userLoaded(.success(user)):
                state.isLoading = false
                state.user = user
                return .none

            case let .userLoaded(.failure(error)):
                state.isLoading = false
                state.error = error.localizedDescription
                return .none

            case .logout:
                state.user = nil
                return .none
            }
        }
    }
}

// UserView.swift
struct UserView: View {
    @Bindable var store: StoreOf<UserFeature>

    var body: some View {
        WithPerceptionTracking {
            content
                .task { store.send(.loadUser("current")) }
        }
    }

    @ViewBuilder
    private var content: some View {
        if store.isLoading {
            ProgressView()
        } else if let error = store.error {
            ErrorView(message: error)
        } else if let user = store.user {
            UserProfileContent(user: user)
        }
    }
}
```

### Android Architecture (MVVM + Clean Architecture)

Clean Architecture Layers:
```kotlin
// Domain Layer - Use Cases
class GetUserUseCase @Inject constructor(
    private val userRepository: UserRepository
) {
    suspend operator fun invoke(userId: String): Result<User> {
        return runCatching { userRepository.getUser(userId) }
    }
}

// Data Layer - Repository Implementation
class UserRepositoryImpl @Inject constructor(
    private val remoteDataSource: UserRemoteDataSource,
    private val localDataSource: UserLocalDataSource,
    private val dispatcher: CoroutineDispatcher
) : UserRepository {

    override suspend fun getUser(userId: String): User = withContext(dispatcher) {
        try {
            val remoteUser = remoteDataSource.fetchUser(userId)
            localDataSource.saveUser(remoteUser.toEntity())
            remoteUser.toDomain()
        } catch (e: IOException) {
            localDataSource.getUser(userId)?.toDomain()
                ?: throw UserNotFoundException(userId)
        }
    }

    override fun observeUser(userId: String): Flow<User> {
        return localDataSource.observeUser(userId)
            .map { it.toDomain() }
            .distinctUntilChanged()
    }
}

// Presentation Layer - ViewModel
@HiltViewModel
class UserViewModel @Inject constructor(
    private val getUserUseCase: GetUserUseCase,
    private val savedStateHandle: SavedStateHandle
) : ViewModel() {

    private val userId = savedStateHandle.get<String>("userId") ?: ""

    val uiState: StateFlow<UserUiState> = flow {
        emit(UserUiState.Loading)
        getUserUseCase(userId)
            .onSuccess { emit(UserUiState.Success(it)) }
            .onFailure { emit(UserUiState.Error(it.message)) }
    }
    .stateIn(
        scope = viewModelScope,
        started = SharingStarted.WhileSubscribed(5000),
        initialValue = UserUiState.Loading
    )
}
```

### Flutter Architecture (Clean Architecture + Riverpod)

Layer Organization:
```dart
// Domain Layer
abstract class UserRepository {
  Future<User> getUser(String id);
  Stream<User> watchUser(String id);
}

class GetUserUseCase {
  final UserRepository _repository;
  GetUserUseCase(this._repository);

  Future<User> call(String id) => _repository.getUser(id);
}

// Data Layer
class UserRepositoryImpl implements UserRepository {
  final UserRemoteDataSource _remote;
  final UserLocalDataSource _local;

  UserRepositoryImpl(this._remote, this._local);

  @override
  Future<User> getUser(String id) async {
    try {
      final user = await _remote.fetchUser(id);
      await _local.cacheUser(user);
      return user;
    } on NetworkException {
      final cached = await _local.getCachedUser(id);
      if (cached != null) return cached;
      rethrow;
    }
  }

  @override
  Stream<User> watchUser(String id) => _local.watchUser(id);
}

// Presentation Layer (Riverpod)
@riverpod
UserRepository userRepository(Ref ref) {
  return UserRepositoryImpl(
    ref.read(userRemoteDataSourceProvider),
    ref.read(userLocalDataSourceProvider),
  );
}

@riverpod
class UserController extends _$UserController {
  @override
  FutureOr<User?> build(String userId) async {
    return ref.read(userRepositoryProvider).getUser(userId);
  }

  Future<void> refresh() async {
    state = const AsyncLoading();
    state = await AsyncValue.guard(
      () => ref.read(userRepositoryProvider).getUser(arg),
    );
  }
}
```

## Concurrency Patterns

### Swift Structured Concurrency

Task Groups for Parallel Execution:
```swift
func loadDashboard() async throws -> Dashboard {
    async let user = api.fetchUser()
    async let posts = api.fetchPosts()
    async let notifications = api.fetchNotifications()

    return try await Dashboard(
        user: user,
        posts: posts,
        notifications: notifications
    )
}

// TaskGroup for dynamic parallelism
func loadAllUsers(_ ids: [String]) async throws -> [User] {
    try await withThrowingTaskGroup(of: User.self) { group in
        for id in ids {
            group.addTask { try await api.fetchUser(id) }
        }
        return try await group.reduce(into: []) { $0.append($1) }
    }
}
```

Actor Isolation:
```swift
actor ImageCache {
    private var cache: [URL: UIImage] = [:]
    private var inProgress: [URL: Task<UIImage, Error>] = [:]

    func image(for url: URL) async throws -> UIImage {
        if let cached = cache[url] { return cached }

        if let task = inProgress[url] {
            return try await task.value
        }

        let task = Task { try await downloadImage(url) }
        inProgress[url] = task

        do {
            let image = try await task.value
            cache[url] = image
            inProgress[url] = nil
            return image
        } catch {
            inProgress[url] = nil
            throw error
        }
    }

    private func downloadImage(_ url: URL) async throws -> UIImage {
        let (data, _) = try await URLSession.shared.data(from: url)
        guard let image = UIImage(data: data) else {
            throw ImageError.invalidData
        }
        return image
    }
}
```

### Kotlin Coroutines

Flow Operators:
```kotlin
fun observeUserWithPosts(userId: String): Flow<UserWithPosts> {
    return combine(
        userDao.observeUser(userId),
        postDao.observePostsByUser(userId)
    ) { user, posts ->
        UserWithPosts(user, posts)
    }
    .distinctUntilChanged()
    .flowOn(Dispatchers.IO)
}

// Retry with exponential backoff
fun <T> Flow<T>.retryWithBackoff(
    maxRetries: Int = 3,
    initialDelay: Long = 1000
): Flow<T> = retryWhen { cause, attempt ->
    if (cause is IOException && attempt < maxRetries) {
        delay(initialDelay * 2.0.pow(attempt.toDouble()).toLong())
        true
    } else {
        false
    }
}

// Debounce for search
fun SearchViewModel.observeSearch() {
    searchQuery
        .debounce(300)
        .filter { it.length >= 2 }
        .distinctUntilChanged()
        .flatMapLatest { query ->
            searchRepository.search(query)
                .catch { emit(SearchResult.Error(it)) }
        }
        .onEach { _searchResults.value = it }
        .launchIn(viewModelScope)
}
```

### Dart Isolates

Compute-Heavy Operations:
```dart
// Simple compute
Future<List<ProcessedItem>> processItems(List<RawItem> items) async {
  return compute(_processItemsIsolate, items);
}

List<ProcessedItem> _processItemsIsolate(List<RawItem> items) {
  return items.map((item) => ProcessedItem.from(item)).toList();
}

// Long-running isolate with bidirectional communication
class ImageProcessor {
  late final Isolate _isolate;
  late final ReceivePort _receivePort;
  late final SendPort _sendPort;

  Future<void> spawn() async {
    _receivePort = ReceivePort();
    _isolate = await Isolate.spawn(
      _isolateEntry,
      _receivePort.sendPort,
    );
    _sendPort = await _receivePort.first as SendPort;
  }

  Future<Uint8List> processImage(Uint8List imageData) async {
    final responsePort = ReceivePort();
    _sendPort.send([imageData, responsePort.sendPort]);
    return await responsePort.first as Uint8List;
  }

  static void _isolateEntry(SendPort sendPort) {
    final receivePort = ReceivePort();
    sendPort.send(receivePort.sendPort);

    receivePort.listen((message) {
      final imageData = message[0] as Uint8List;
      final replyPort = message[1] as SendPort;

      // Heavy image processing
      final processed = _heavyImageProcessing(imageData);
      replyPort.send(processed);
    });
  }
}
```

## Navigation Patterns

### SwiftUI Navigation (iOS 17+)
```swift
@Observable
class NavigationRouter {
    var path = NavigationPath()

    func push<D: Hashable>(_ destination: D) {
        path.append(destination)
    }

    func pop() {
        guard !path.isEmpty else { return }
        path.removeLast()
    }

    func popToRoot() {
        path.removeLast(path.count)
    }
}

// Usage
struct ContentView: View {
    @State private var router = NavigationRouter()

    var body: some View {
        NavigationStack(path: $router.path) {
            HomeView()
                .navigationDestination(for: User.self) { user in
                    UserDetailView(user: user)
                }
                .navigationDestination(for: Post.self) { post in
                    PostDetailView(post: post)
                }
        }
        .environment(router)
    }
}
```

### Jetpack Compose Navigation
```kotlin
// NavGraph setup
@Composable
fun AppNavGraph(
    navController: NavHostController = rememberNavController()
) {
    NavHost(
        navController = navController,
        startDestination = "home"
    ) {
        composable("home") {
            HomeScreen(
                onUserClick = { navController.navigate("user/${it.id}") }
            )
        }

        composable(
            route = "user/{userId}",
            arguments = listOf(navArgument("userId") { type = NavType.StringType })
        ) { backStackEntry ->
            val userId = backStackEntry.arguments?.getString("userId") ?: ""
            UserScreen(
                userId = userId,
                onBack = { navController.popBackStack() }
            )
        }

        navigation(
            startDestination = "settings/main",
            route = "settings"
        ) {
            composable("settings/main") { SettingsMainScreen() }
            composable("settings/profile") { SettingsProfileScreen() }
        }
    }
}
```

### Flutter Navigation (go_router)
```dart
final router = GoRouter(
  initialLocation: '/',
  routes: [
    GoRoute(
      path: '/',
      builder: (context, state) => const HomeScreen(),
      routes: [
        GoRoute(
          path: 'user/:id',
          builder: (context, state) => UserScreen(
            userId: state.pathParameters['id']!,
          ),
        ),
      ],
    ),
    ShellRoute(
      builder: (context, state, child) => ScaffoldWithNavBar(child: child),
      routes: [
        GoRoute(path: '/feed', builder: (_, __) => const FeedScreen()),
        GoRoute(path: '/search', builder: (_, __) => const SearchScreen()),
        GoRoute(path: '/profile', builder: (_, __) => const ProfileScreen()),
      ],
    ),
  ],
  redirect: (context, state) {
    final isLoggedIn = ref.read(authProvider).isLoggedIn;
    final isLoggingIn = state.matchedLocation == '/login';

    if (!isLoggedIn && !isLoggingIn) return '/login';
    if (isLoggedIn && isLoggingIn) return '/';
    return null;
  },
);
```

## Performance Optimization

### iOS Performance
- Use `@Observable` instead of `@ObservedObject` for better SwiftUI performance
- Implement lazy loading with `LazyVStack`/`LazyHStack`
- Use `task(id:)` for cancellable async operations
- Profile with Instruments: Time Profiler, Allocations, Leaks

### Android Performance
- Use `remember` and `derivedStateOf` to minimize recompositions
- Implement pagination with `LazyColumn` and `Paging 3`
- Use `collectAsStateWithLifecycle` for lifecycle-aware collection
- Profile with Android Studio Profiler: CPU, Memory, Network

### Flutter Performance
- Use `const` constructors where possible
- Implement `RepaintBoundary` for complex widgets
- Use `ListView.builder` for long lists
- Profile with Flutter DevTools: Timeline, Memory

## Testing Frameworks

### iOS Testing Stack
- XCTest: Built-in unit and UI testing
- Quick/Nimble: BDD-style testing
- ViewInspector: SwiftUI view testing
- SnapshotTesting: UI snapshot tests

### Android Testing Stack
- JUnit 5: Unit testing
- MockK: Kotlin-first mocking
- Turbine: Flow testing
- Robolectric: Android framework simulation
- Espresso: UI testing

### Flutter Testing Stack
- flutter_test: Built-in testing
- mockito: Mocking
- bloc_test: BLoC testing
- golden_toolkit: Golden tests
- integration_test: Integration testing

---

Version: 1.0.0
Last Updated: 2025-12-07
