---
name: moai-lang-mobile
description: Mobile development specialist covering Swift 6 for iOS, Kotlin for Android, and Dart/Flutter for cross-platform. Use when developing iOS apps, Android apps, cross-platform mobile solutions, or native mobile integrations.
version: 1.0.0
category: language
tags:
  - mobile
  - ios
  - android
  - flutter
  - swift
  - kotlin
  - dart
updated: 2025-12-07
status: active
---

## Quick Reference (30 seconds)

Mobile Development Expert - Swift 6, Kotlin 2.0, Dart 3.5/Flutter 3.24 with native patterns.

Auto-Triggers: Mobile-specific files (`.swift`, `.kt`, `.dart`), iOS/Android projects, Flutter apps

Core Capabilities:
- Swift 6.0: Typed throws, actors, SwiftUI 6, Combine
- Kotlin 2.0: Jetpack Compose, Coroutines, Room
- Dart 3.5/Flutter 3.24: Patterns, Riverpod, Platform Channels
- Cross-platform: Native integrations, shared business logic
- Testing: XCTest, JUnit, widget_test, integration_test

## Implementation Guide (5 minutes)

### Swift 6.0 (iOS Development)

Core Language Features:
- Typed Throws: `func fetch() throws(NetworkError) -> Data`
- Custom Actor Executors: Fine-grained concurrency control
- Embedded Swift: Bare-metal and embedded systems support
- Complete Concurrency: Data-race safety by default

SwiftUI 6 Patterns:
```swift
// Modern SwiftUI View with observation
@Observable class UserViewModel {
    var user: User?
    var isLoading = false

    func loadUser() async throws(APIError) {
        isLoading = true
        defer { isLoading = false }
        user = try await api.fetchUser()
    }
}

struct UserProfileView: View {
    @State private var viewModel = UserViewModel()

    var body: some View {
        NavigationStack {
            Group {
                if viewModel.isLoading {
                    ProgressView()
                } else if let user = viewModel.user {
                    UserDetailView(user: user)
                }
            }
            .task { try? await viewModel.loadUser() }
        }
    }
}
```

Async/Await and Actors:
```swift
actor UserCache {
    private var cache: [String: User] = [:]

    func get(_ id: String) -> User? { cache[id] }
    func set(_ id: String, user: User) { cache[id] = user }
}

@MainActor
class UserController {
    private let cache = UserCache()

    func loadUser(_ id: String) async throws -> User {
        if let cached = await cache.get(id) { return cached }
        let user = try await api.fetchUser(id)
        await cache.set(id, user: user)
        return user
    }
}
```

### Kotlin 2.0 (Android Development)

Jetpack Compose UI:
```kotlin
@Composable
fun UserProfileScreen(
    viewModel: UserViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()

    Scaffold(
        topBar = { TopAppBar(title = { Text("Profile") }) }
    ) { padding ->
        when (val state = uiState) {
            is UiState.Loading -> CircularProgressIndicator()
            is UiState.Success -> UserContent(state.user, Modifier.padding(padding))
            is UiState.Error -> ErrorMessage(state.message)
        }
    }
}

@Composable
private fun UserContent(user: User, modifier: Modifier = Modifier) {
    LazyColumn(modifier = modifier) {
        item { UserHeader(user) }
        items(user.posts) { post -> PostCard(post) }
    }
}
```

ViewModel with StateFlow:
```kotlin
@HiltViewModel
class UserViewModel @Inject constructor(
    private val userRepository: UserRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow<UiState>(UiState.Loading)
    val uiState: StateFlow<UiState> = _uiState.asStateFlow()

    init { loadUser() }

    private fun loadUser() {
        viewModelScope.launch {
            userRepository.getUser()
                .catch { e -> _uiState.value = UiState.Error(e.message) }
                .collect { user -> _uiState.value = UiState.Success(user) }
        }
    }
}

sealed interface UiState {
    data object Loading : UiState
    data class Success(val user: User) : UiState
    data class Error(val message: String?) : UiState
}
```

Room Database with Flow:
```kotlin
@Entity(tableName = "users")
data class UserEntity(
    @PrimaryKey val id: String,
    val name: String,
    val email: String,
    @ColumnInfo(name = "created_at") val createdAt: Long
)

@Dao
interface UserDao {
    @Query("SELECT * FROM users WHERE id = :id")
    fun getUserById(id: String): Flow<UserEntity?>

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertUser(user: UserEntity)

    @Query("DELETE FROM users WHERE id = :id")
    suspend fun deleteUser(id: String)
}

@Database(entities = [UserEntity::class], version = 1)
abstract class AppDatabase : RoomDatabase() {
    abstract fun userDao(): UserDao
}
```

### Dart 3.5 / Flutter 3.24

Dart 3 Patterns and Records:
```dart
// Sealed classes with pattern matching
sealed class Result<T> {
  const Result();
}
class Success<T> extends Result<T> {
  final T data;
  const Success(this.data);
}
class Failure<T> extends Result<T> {
  final String error;
  const Failure(this.error);
}

// Pattern matching with exhaustiveness
String handleResult(Result<User> result) => switch (result) {
  Success(:final data) => 'User: ${data.name}',
  Failure(:final error) => 'Error: $error',
};

// Records for multiple return values
(String name, int age) parseUser(Map<String, dynamic> json) {
  return (json['name'] as String, json['age'] as int);
}
```

Riverpod State Management:
```dart
// Provider definitions
@riverpod
class UserNotifier extends _$UserNotifier {
  @override
  FutureOr<User?> build() => null;

  Future<void> loadUser(String id) async {
    state = const AsyncLoading();
    state = await AsyncValue.guard(
      () => ref.read(userRepositoryProvider).getUser(id),
    );
  }
}

@riverpod
UserRepository userRepository(Ref ref) {
  return UserRepository(ref.read(dioProvider));
}

// Widget usage
class UserScreen extends ConsumerWidget {
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final userAsync = ref.watch(userNotifierProvider);

    return userAsync.when(
      data: (user) => user != null
        ? UserProfile(user: user)
        : const EmptyState(),
      loading: () => const CircularProgressIndicator(),
      error: (e, st) => ErrorWidget(e.toString()),
    );
  }
}
```

Platform Channels:
```dart
// Dart side
class NativeBridge {
  static const _channel = MethodChannel('com.app/native');

  Future<String> getPlatformVersion() async {
    final version = await _channel.invokeMethod<String>('getPlatformVersion');
    return version ?? 'Unknown';
  }

  Future<void> shareContent(String text) async {
    await _channel.invokeMethod('share', {'text': text});
  }
}

// Swift side (iOS)
@main
class AppDelegate: FlutterAppDelegate {
  override func application(
    _ application: UIApplication,
    didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?
  ) -> Bool {
    let controller = window?.rootViewController as! FlutterViewController
    let channel = FlutterMethodChannel(
      name: "com.app/native",
      binaryMessenger: controller.binaryMessenger
    )

    channel.setMethodCallHandler { call, result in
      switch call.method {
      case "getPlatformVersion":
        result("iOS \(UIDevice.current.systemVersion)")
      case "share":
        if let args = call.arguments as? [String: Any],
           let text = args["text"] as? String {
          // Share implementation
          result(nil)
        }
      default:
        result(FlutterMethodNotImplemented)
      }
    }

    return super.application(application, didFinishLaunchingWithOptions: launchOptions)
  }
}
```

### Testing Strategies

Swift XCTest:
```swift
@MainActor
final class UserViewModelTests: XCTestCase {
    var sut: UserViewModel!
    var mockAPI: MockUserAPI!

    override func setUp() {
        mockAPI = MockUserAPI()
        sut = UserViewModel(api: mockAPI)
    }

    func testLoadUserSuccess() async throws {
        mockAPI.mockUser = User(id: "1", name: "Test")
        try await sut.loadUser()
        XCTAssertEqual(sut.user?.name, "Test")
        XCTAssertFalse(sut.isLoading)
    }
}
```

Kotlin JUnit with Coroutines:
```kotlin
@OptIn(ExperimentalCoroutinesApi::class)
class UserViewModelTest {
    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private lateinit var viewModel: UserViewModel
    private val mockRepository = mockk<UserRepository>()

    @Before
    fun setup() {
        viewModel = UserViewModel(mockRepository)
    }

    @Test
    fun `loadUser emits success state`() = runTest {
        coEvery { mockRepository.getUser() } returns flowOf(testUser)

        viewModel.uiState.test {
            assertEquals(UiState.Loading, awaitItem())
            assertEquals(UiState.Success(testUser), awaitItem())
        }
    }
}
```

Flutter Widget Test:
```dart
void main() {
  testWidgets('UserScreen shows user data', (tester) async {
    final container = ProviderContainer(overrides: [
      userNotifierProvider.overrideWith(() => MockUserNotifier()),
    ]);

    await tester.pumpWidget(
      UncontrolledProviderScope(
        container: container,
        child: const MaterialApp(home: UserScreen()),
      ),
    );

    expect(find.byType(CircularProgressIndicator), findsOneWidget);
    await tester.pumpAndSettle();
    expect(find.text('Test User'), findsOneWidget);
  });
}
```

## Advanced Patterns

For comprehensive coverage including:
- Advanced concurrency patterns (Actors, Combine, Coroutines, Isolates)
- Architecture patterns (MVVM, Clean Architecture, BLoC)
- Performance optimization and profiling
- CI/CD and app distribution

See: [reference.md](reference.md) and [examples.md](examples.md)

## Context7 Library Mappings

Swift/iOS: `/apple/swift`, `/apple/swift-package-manager`
Kotlin/Android: `/JetBrains/kotlin`, `/android/architecture-components-samples`
Flutter/Dart: `/flutter/flutter`, `/dart-lang/sdk`
Testing: `/Quick/Quick`, `/pointfreeco/swift-composable-architecture`

## Works Well With

- `moai-domain-backend` - API integration and backend communication
- `moai-quality-security` - Mobile security best practices
- `moai-essentials-debug` - Mobile debugging and profiling
- `moai-foundation-trust` - TRUST 5 quality principles for mobile
- `moai-context7-integration` - Latest mobile documentation access

---

Version: 1.0.0
Last Updated: 2025-12-07
Status: Production Ready
