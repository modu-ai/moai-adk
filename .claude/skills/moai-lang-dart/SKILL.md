---
name: moai-lang-dart
version: 3.0.0
created: 2025-11-06
updated: 2025-11-06
status: active
description: Dart 3.x enterprise development with Flutter 3.x, null safety, patterns, records, and Context7 MCP integration.
keywords: ['dart', 'flutter', 'null-safety', 'patterns', 'records', 'extensions', 'async', 'state-management']
allowed-tools:
  - Read
  - Bash
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
---

# 🚀 Dart 3.x & Flutter 3.x Enterprise Development Premium Skill

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-lang-dart |
| **Version** | 3.0.0 (2025-11-06) - Premium Edition |
| **Allowed tools** | Read, Bash, Context7 MCP Integration |
| **Auto-load** | Dart projects, Flutter apps, cross-platform development |
| **Tier** | Premium Language |
| **Context7 Integration** | Dart 3.x + Flutter 3.x Official Docs |

---

## 🎯 What It Does

**Enterprise-grade Dart 3.x and Flutter 3.x development** with modern language features, null safety, advanced patterns, and production-ready cross-platform applications.

### Core Capabilities

**🎯 Dart 3.x Modern Language Features**:
- **Null Safety**: Sound null safety with flow analysis and type promotion
- **Patterns & Records**: Destructuring, pattern matching, record types
- **Extensions**: Extension methods for enhanced functionality
- **Advanced Types**: Sealed classes, enhanced enums, type aliases
- **Concurrency**: Isolates, async/await, streams with sound safety

**📱 Flutter 3.x Advanced Development**:
- **Widget System**: Composition, stateful widgets, performance optimization
- **State Management**: Provider, Riverpod, BLoC, and custom solutions
- **Navigation**: Navigator 2.0, deep linking, route management
- **Cross-platform**: iOS, Android, Web, Desktop, Embedded support
- **Rendering**: Custom painters, shaders, performance profiling

**⚡ Enterprise Flutter Ecosystem**:
- **Testing**: Widget tests, integration tests, golden tests
- **Build Systems**: Flutter build runner, code generation
- **Internationalization**: flutter_localizations, l10n, i18n
- **CI/CD**: GitHub Actions, Codemagic, Firebase App Distribution

---

## 🌟 Enterprise Patterns & Best Practices

### Dart 3.x Null Safety & Modern Patterns

```dart
// Modern Dart with null safety and patterns
class UserRepository {
  final ApiService _api;
  final CacheService _cache;

  UserRepository(this._api, this._cache);

  // Async pattern with null safety
  Future<User?> getUserById(String id) async {
    // Check cache first with null-aware operators
    final cachedUser = await _cache.getUser(id);
    if (cachedUser != null) return cachedUser;

    // Fetch from API with error handling
    try {
      final response = await _api.get('/users/$id');
      if (response.statusCode == 200) {
        final user = User.fromJson(response.data);
        await _cache.saveUser(id, user);
        return user;
      }
    } on NetworkException catch (e) {
      _logger.warning('Network error fetching user $id: $e');
    }

    return null;
  }

  // Pattern matching with records
  (User?, bool) getUserWithFreshness(String id) {
    final user = _cache.getUserSync(id);
    final isFresh = user?.isFresh ?? false;
    return (user, isFresh);
  }

  // Extension method for validation
  bool isValidUserId(String id) => id.isNotEmpty && id.length >= 3;
}

// Extension on String for validation
extension UserValidation on String {
  bool get isValidEmail => RegExp(r'^[\w-\.]+@([\w-]+\.)+[\w-]{2,4}$').hasMatch(this);
  bool get isValidPassword => length >= 8 && contains(RegExp(r'[A-Z]')) && contains(RegExp(r'[0-9]'));
}

// Records for data transfer
typedef UserProfile = ({
  String id,
  String name,
  String email,
  DateTime lastLogin,
  bool isActive,
});

class UserService {
  UserProfile createProfile(User user) => (
    id: user.id,
    name: user.name,
    email: user.email,
    lastLogin: user.lastLogin ?? DateTime.now(),
    isActive: user.isActive,
  );

  // Pattern matching with switch expressions
  UserStatus determineStatus(UserProfile profile) {
    return switch (profile) {
      (isActive: true, lastLogin: final date) when DateTime.now().difference(date).inDays < 30:
        UserStatus.active,
      (isActive: true) => UserStatus.inactive,
      (isActive: false) => UserStatus.suspended,
    };
  }
}
```

### Flutter 3.x Advanced State Management

```dart
// Modern state management with Provider and Riverpod patterns
class AppState extends ChangeNotifier {
  User? _currentUser;
  List<Project> _projects = [];
  bool _isLoading = false;
  String? _error;

  User? get currentUser => _currentUser;
  List<Project> get projects => List.unmodifiable(_projects);
  bool get isLoading => _isLoading;
  String? get error => _error;

  bool get isAuthenticated => _currentUser != null;

  Future<void> loadUser() async {
    _setLoading(true);
    _error = null;

    try {
      final user = await _userRepository.getCurrentUser();
      _currentUser = user;
      await _loadProjects();
    } catch (e) {
      _error = 'Failed to load user: $e';
    } finally {
      _setLoading(false);
    }
  }

  Future<void> _loadProjects() async {
    if (_currentUser == null) return;

    try {
      final projects = await _projectRepository.getUserProjects(_currentUser!.id);
      _projects = projects;
    } catch (e) {
      _error = 'Failed to load projects: $e';
    }
  }

  void _setLoading(bool loading) {
    _isLoading = loading;
    notifyListeners();
  }
}

// Riverpod provider for modern state management
@riverpod
class UserNotifier extends _$UserNotifier {
  @override
  Future<User?> build() async {
    return _userRepository.getCurrentUser();
  }

  Future<void> refreshUser() async {
    state = const AsyncValue.loading();

    state = await AsyncValue.guard(() async {
      final user = await _userRepository.getCurrentUser();
      return user;
    });
  }

  Future<void> updateUserProfile(UserProfile profile) async {
    final previousState = state.value;

    state = await AsyncValue.guard(() async {
      final updatedUser = await _userRepository.updateProfile(profile);
      return updatedUser;
    });

    // Rollback on error
    if (state.hasError && previousState != null) {
      state = AsyncValue.data(previousState);
    }
  }
}

// Multi-provider setup for enterprise apps
class AppProviders extends ConsumerWidget {
  const AppProviders({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return MultiProvider(
      providers: [
        ChangeNotifierProvider(create: (_) => AppState()),
        Provider(create: (_) => NavigationService()),
        Provider(create: (_) => ApiService()),
        Provider(create: (_) => CacheService()),
      ],
      child: const AppNavigation(),
    );
  }
}
```

### Enterprise Flutter Architecture

```dart
// Clean architecture with dependency injection
abstract class UserRepository {
  Future<User?> getUserById(String id);
  Future<List<User>> getAllUsers();
  Future<void> saveUser(User user);
  Future<void> deleteUser(String id);
}

class UserRepositoryImpl implements UserRepository {
  final ApiService _api;
  final CacheService _cache;
  final Logger _logger;

  UserRepositoryImpl(this._api, this._cache, this._logger);

  @override
  Future<User?> getUserById(String id) async {
    try {
      // Implement caching strategy
      final cached = await _cache.get<User>('user_$id');
      if (cached != null) {
        _logger.debug('User $id found in cache');
        return cached;
      }

      // Fetch from API
      final response = await _api.get('/users/$id');
      if (response.statusCode == 200) {
        final user = User.fromJson(response.data);
        await _cache.set('user_$id', user, Duration(minutes: 15));
        return user;
      }
    } catch (e) {
      _logger.error('Error fetching user $id: $e');
      rethrow;
    }
    return null;
  }

  @override
  Future<List<User>> getAllUsers() async {
    try {
      final response = await _api.get('/users');
      if (response.statusCode == 200) {
        final users = (response.data as List)
            .map((json) => User.fromJson(json))
            .toList();

        // Cache the result
        await _cache.set('all_users', users, Duration(minutes: 5));
        return users;
      }
    } catch (e) {
      _logger.error('Error fetching all users: $e');
      rethrow;
    }
    return [];
  }

  @override
  Future<void> saveUser(User user) async {
    try {
      final response = await _api.post('/users', data: user.toJson());
      if (response.statusCode == 201 || response.statusCode == 200) {
        // Update cache
        await _cache.set('user_${user.id}', user, Duration(minutes: 15));
        _logger.info('User ${user.id} saved successfully');
      } else {
        throw ApiException('Failed to save user: ${response.statusCode}');
      }
    } catch (e) {
      _logger.error('Error saving user ${user.id}: $e');
      rethrow;
    }
  }

  @override
  Future<void> deleteUser(String id) async {
    try {
      final response = await _api.delete('/users/$id');
      if (response.statusCode == 204 || response.statusCode == 200) {
        // Remove from cache
        await _cache.delete('user_$id');
        _logger.info('User $id deleted successfully');
      } else {
        throw ApiException('Failed to delete user: ${response.statusCode}');
      }
    } catch (e) {
      _logger.error('Error deleting user $id: $e');
      rethrow;
    }
  }
}

// BLoC pattern for complex state management
class UserBloc extends Bloc<UserEvent, UserState> {
  final UserRepository _repository;

  UserBloc(this._repository) : super(UserState.initial()) {
    on<LoadUsers>(_onLoadUsers);
    on<AddUser>(_onAddUser);
    on<UpdateUser>(_onUpdateUser);
    on<DeleteUser>(_onDeleteUser);
  }

  Future<void> _onLoadUsers(LoadUsers event, Emitter<UserState> emit) async {
    emit(state.copyWith(isLoading: true, error: null));

    try {
      final users = await _repository.getAllUsers();
      emit(state.copyWith(
        isLoading: false,
        users: users,
      ));
    } catch (e) {
      emit(state.copyWith(
        isLoading: false,
        error: 'Failed to load users: $e',
      ));
    }
  }

  Future<void> _onAddUser(AddUser event, Emitter<UserState> emit) async {
    try {
      await _repository.saveUser(event.user);
      final updatedUsers = [...state.users, event.user];
      emit(state.copyWith(users: updatedUsers));
    } catch (e) {
      emit(state.copyWith(error: 'Failed to add user: $e'));
    }
  }

  Future<void> _onUpdateUser(UpdateUser event, Emitter<UserState> emit) async {
    try {
      await _repository.saveUser(event.user);
      final updatedUsers = state.users
          .map((user) => user.id == event.user.id ? event.user : user)
          .toList();
      emit(state.copyWith(users: updatedUsers));
    } catch (e) {
      emit(state.copyWith(error: 'Failed to update user: $e'));
    }
  }

  Future<void> _onDeleteUser(DeleteUser event, Emitter<UserState> emit) async {
    try {
      await _repository.deleteUser(event.userId);
      final updatedUsers = state.users.where((user) => user.id != event.userId).toList();
      emit(state.copyWith(users: updatedUsers));
    } catch (e) {
      emit(state.copyWith(error: 'Failed to delete user: $e'));
    }
  }
}
```

---

## 🔧 Modern Development Workflow

### pubspec.yaml Configuration (Flutter 3.x)

```yaml
name: enterprise_flutter_app
description: Enterprise Flutter application with modern architecture

# Prevent accidental publishing to pub.dev
publish_to: 'none'

version: 1.0.0+1

environment:
  sdk: '>=3.0.0 <4.0.0'
  flutter: ">=3.10.0"

dependencies:
  flutter:
    sdk: flutter
  flutter_localizations:
    sdk: flutter

  # State Management
  flutter_riverpod: ^2.4.9
  provider: ^6.1.1
  bloc: ^8.1.2
  flutter_bloc: ^8.1.3

  # Networking & API
  dio: ^5.3.2
  retrofit: ^4.0.3
  json_annotation: ^4.8.1

  # Database & Storage
  hive: ^2.2.3
  hive_flutter: ^1.1.0
  shared_preferences: ^2.2.2
  secure_storage: ^8.0.0

  # Architecture
  get_it: ^7.6.4
  injectable: ^2.3.2
  auto_route: ^7.9.2

  # UI Components
  cupertino_icons: ^1.0.6
  material_symbols_icons: ^4.2719.3
  flutter_svg: ^2.0.9
  cached_network_image: ^3.3.0

  # Utilities
  logger: ^2.0.2+1
  intl: ^0.18.1
  uuid: ^4.2.1
  equatable: ^2.0.5

  # Testing (dev_dependencies)
dev_dependencies:
  flutter_test:
    sdk: flutter
  integration_test:
    sdk: flutter

  # Code Generation
  build_runner: ^2.4.7
  retrofit_generator: ^8.0.4
  json_serializable: ^6.7.1
  hive_generator: ^2.0.1
  injectable_generator: ^2.4.1
  auto_route_generator: ^7.3.2

  # Testing Tools
  mockito: ^5.4.2
  golden_toolkit: ^0.15.0
  bloc_test: ^9.1.4
  flutter_driver:
    sdk: flutter

  # Code Quality
  very_good_analysis: ^5.1.0
  dart_code_metrics: ^5.7.6

flutter:
  uses-material-design: true
  generate: true

  assets:
    - assets/images/
    - assets/icons/
    - assets/config/

  fonts:
    - family: CustomIcons
      fonts:
        - asset: assets/fonts/CustomIcons.ttf
```

### Advanced Testing Patterns

```dart
// Widget testing with test helpers
class UserListScreenTestHelpers {
  static Widget createWidgetUnderTest({
    required UserBloc userBloc,
    bool isLoading = false,
  }) {
    return MaterialApp(
      home: BlocProvider<UserBloc>.value(
        value: userBloc,
        child: UserListScreen(isLoading: isLoading),
      ),
    );
  }

  static User createMockUser({
    String id = '1',
    String name = 'Test User',
    String email = 'test@example.com',
  }) {
    return User(
      id: id,
      name: name,
      email: email,
      createdAt: DateTime.now(),
    );
  }
}

void main() {
  group('UserListScreen Tests', () {
    late UserBloc userBloc;
    late MockUserRepository mockRepository;

    setUp(() {
      mockRepository = MockUserRepository();
      userBloc = UserBloc(mockRepository);
    });

    tearDown(() {
      userBloc.close();
    });

    testWidgets('displays loading indicator when loading', (tester) async {
      when(() => mockRepository.getAllUsers())
          .thenAnswer((_) async => []);

      await tester.pumpWidget(
        UserListScreenTestHelpers.createWidgetUnderTest(
          userBloc: userBloc,
          isLoading: true,
        ),
      );

      expect(find.byType(CircularProgressIndicator), findsOneWidget);
    });

    testWidgets('displays user list when loaded', (tester) async {
      final users = [
        UserListScreenTestHelpers.createMockUser(name: 'User 1'),
        UserListScreenTestHelpers.createMockUser(name: 'User 2'),
      ];

      when(() => mockRepository.getAllUsers())
          .thenAnswer((_) async => users);

      await tester.pumpWidget(
        UserListScreenTestHelpers.createWidgetUnderTest(
          userBloc: userBloc,
        ),
      );

      await tester.pumpAndSettle();

      expect(find.text('User 1'), findsOneWidget);
      expect(find.text('User 2'), findsOneWidget);
      expect(find.byType(UserListView), findsOneWidget);
    });

    testWidgets('displays error message on failure', (tester) async {
      when(() => mockRepository.getAllUsers())
          .thenThrow(Exception('Network error'));

      await tester.pumpWidget(
        UserListScreenTestHelpers.createWidgetUnderTest(
          userBloc: userBloc,
        ),
      );

      await tester.pumpAndSettle();

      expect(find.text('Failed to load users'), findsOneWidget);
      expect(find.byType(ErrorWidget), findsOneWidget);
    });

    testWidgets('tapping user navigates to detail screen', (tester) async {
      final users = [
        UserListScreenTestHelpers.createMockUser(name: 'Test User'),
      ];

      when(() => mockRepository.getAllUsers())
          .thenAnswer((_) async => users);

      await tester.pumpWidget(
        UserListScreenTestHelpers.createWidgetUnderTest(
          userBloc: userBloc,
        ),
      );

      await tester.pumpAndSettle();

      await tester.tap(find.text('Test User'));
      await tester.pumpAndSettle();

      verify(() => mockRepository.getUserById('1')).called(1);
    });
  });
}

// Integration tests
void main() {
  group('User Management Integration Tests', () {
    integrationTest('Complete user management flow', (tester) async {
      // Launch app
      app.main();
      await tester.pumpAndSettle();

      // Navigate to user list
      await tester.tap(find.byIcon(Icons.people));
      await tester.pumpAndSettle();

      // Verify initial state
      expect(find.byType(UserListScreen), findsOneWidget);

      // Add new user
      await tester.tap(find.byIcon(Icons.add));
      await tester.pumpAndSettle();

      await tester.enterText(
        find.byKey(Key('name_field')),
        'Integration Test User',
      );
      await tester.enterText(
        find.byKey(Key('email_field')),
        'integration@test.com',
      );
      await tester.tap(find.byKey(Key('save_button')));
      await tester.pumpAndSettle();

      // Verify user appears in list
      expect(find.text('Integration Test User'), findsOneWidget);
      expect(find.text('integration@test.com'), findsOneWidget);

      // Navigate to user details
      await tester.tap(find.text('Integration Test User'));
      await tester.pumpAndSettle();

      // Verify details screen
      expect(find.byType(UserDetailScreen), findsOneWidget);
      expect(find.text('Integration Test User'), findsOneWidget);

      // Edit user
      await tester.tap(find.byIcon(Icons.edit));
      await tester.pumpAndSettle();

      await tester.enterText(
        find.byKey(Key('name_field')),
        'Updated User',
      );
      await tester.tap(find.byKey(Key('save_button')));
      await tester.pumpAndSettle();

      // Verify update
      expect(find.text('Updated User'), findsOneWidget);
      expect(find.text('Integration Test User'), findsNothing);

      // Delete user
      await tester.tap(find.byIcon(Icons.delete));
      await tester.pumpAndSettle();

      await tester.tap(find.text('Delete'));
      await tester.pumpAndSettle();

      // Verify deletion
      expect(find.text('Updated User'), findsNothing);
    });
  });
}
```

---

## 📊 Performance Optimization Strategies

### Dart Concurrency & Performance

```dart
// Isolate-based background processing
class BackgroundProcessor {
  static final ReceivePort _receivePort = ReceivePort();
  static Isolate? _isolate;

  static Future<void> start() async {
    _isolate = await Isolate.spawn(_backgroundWorker, _receivePort.sendPort);
  }

  static Future<void> stop() async {
    _isolate?.kill(priority: Isolate.immediate);
    _isolate = null;
  }

  static Future<T> process<T>(T data) async {
    final completer = Completer<T>();

    _receivePort.listen((result) {
      if (result is T) {
        completer.complete(result);
      }
    });

    _receivePort.sendPort.send(data);
    return completer.future;
  }

  static void _backgroundWorker(SendPort sendPort) {
    final receivePort = ReceivePort();
    sendPort.send(receivePort.sendPort);

    receivePort.listen((data) async {
      // Process data in background
      final result = await _processData(data);
      sendPort.send(result);
    });
  }

  static Future<T> _processData<T>(T data) async {
    // Simulate heavy computation
    await Future.delayed(Duration(milliseconds: 100));

    // Return processed data
    if (data is List) {
      return data.map((item) => _transformItem(item)).toList() as T;
    }

    return data;
  }

  static dynamic _transformItem(dynamic item) {
    // Transform logic here
    return item;
  }
}

// Stream optimization for real-time data
class OptimizedStreamManager {
  final StreamController<DataEvent> _controller = StreamController<DataEvent>.broadcast();
  final Map<String, StreamSubscription> _subscriptions = {};
  final Debouncer _debouncer = Debouncer(Duration(milliseconds: 300));

  Stream<DataEvent> get events => _controller.stream;

  void subscribeToDataStream(String id, Stream<dynamic> stream) {
    // Cancel existing subscription if any
    _subscriptions[id]?.cancel();

    _subscriptions[id] = stream
        .where((event) => _isValidEvent(event))
        .transform(_eventTransformer)
        .listen(
          (event) => _debouncer.run(() => _handleEvent(id, event)),
          onError: (error) => _handleError(id, error),
        );
  }

  void unsubscribe(String id) {
    _subscriptions[id]?.cancel();
    _subscriptions.remove(id);
  }

  bool _isValidEvent(dynamic event) {
    return event != null && _isRecentEvent(event);
  }

  bool _isRecentEvent(dynamic event) {
    final timestamp = event['timestamp'] as int?;
    if (timestamp == null) return false;

    final now = DateTime.now().millisecondsSinceEpoch;
    return (now - timestamp) < Duration.minutes(5).inMilliseconds;
  }

  StreamTransformer<dynamic, DataEvent> get _eventTransformer {
    return StreamTransformer.fromHandlers(
      handleData: (data, sink) {
        try {
          final event = DataEvent.fromJson(data);
          sink.add(event);
        } catch (e) {
          sink.addError(DataTransformationError('Invalid event format: $e'));
        }
      },
    );
  }

  void _handleEvent(String id, dynamic event) {
    _controller.add(DataEvent(
      source: id,
      data: event,
      timestamp: DateTime.now(),
    ));
  }

  void _handleError(String id, dynamic error) {
    _controller.addError(SubscriptionError(id, error));
  }

  void dispose() {
    for (final subscription in _subscriptions.values) {
      subscription.cancel();
    }
    _subscriptions.clear();
    _controller.close();
  }
}
```

### Flutter Performance Optimization

```dart
// Performance-optimized widget with const constructors and keys
class OptimizedUserList extends StatelessWidget {
  final List<User> users;
  final Function(User) onUserTap;
  final Function(User) onUserDelete;

  const OptimizedUserList({
    super.key,
    required this.users,
    required this.onUserTap,
    required this.onUserDelete,
  });

  @override
  Widget build(BuildContext context) {
    return ListView.builder(
      itemCount: users.length,
      // Use item extent for better performance
      itemExtent: 80.0,
      cacheExtent: 500.0,
      itemBuilder: (context, index) {
        final user = users[index];
        // Use ValueKey for efficient rebuilding
        return UserTile(
          key: ValueKey(user.id),
          user: user,
          onTap: onUserTap,
          onDelete: onUserDelete,
        );
      },
    );
  }
}

// Optimized individual tile widget
class UserTile extends StatelessWidget {
  final User user;
  final Function(User) onTap;
  final Function(User) onDelete;

  const UserTile({
    super.key,
    required this.user,
    required this.onTap,
    required this.onDelete,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
      child: InkWell(
        onTap: () => onTap(user),
        borderRadius: BorderRadius.circular(8),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Row(
            children: [
              // Use Hero for smooth navigation
              Hero(
                tag: 'user_avatar_${user.id}',
                child: CircleAvatar(
                  radius: 24,
                  backgroundColor: Theme.of(context).primaryColor.withOpacity(0.1),
                  child: Text(
                    user.name.isNotEmpty ? user.name[0].toUpperCase() : '?',
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                ),
              ),
              const SizedBox(width: 16),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text(
                      user.name,
                      style: Theme.of(context).textTheme.titleMedium,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                    const SizedBox(height: 4),
                    Text(
                      user.email,
                      style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                        color: Theme.of(context).textTheme.bodyMedium?.color?.withOpacity(0.7),
                      ),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 8),
              // Use Consumer for expensive operations only
              Consumer<UserCacheService>(
                builder: (context, cache, child) {
                  return IconButton(
                    icon: Icon(
                      cache.isFavorite(user.id) ? Icons.favorite : Icons.favorite_border,
                      color: cache.isFavorite(user.id) ? Colors.red : null,
                    ),
                    onPressed: () => cache.toggleFavorite(user.id),
                    tooltip: cache.isFavorite(user.id) ? 'Remove from favorites' : 'Add to favorites',
                  );
                },
              ),
              const SizedBox(width: 8),
              DeleteButton(
                onPressed: () => onDelete(user),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

// Custom performance monitoring widget
class PerformanceMonitor extends StatefulWidget {
  final Widget child;
  final String? name;

  const PerformanceMonitor({
    super.key,
    required this.child,
    this.name,
  });

  @override
  State<PerformanceMonitor> createState() => _PerformanceMonitorState();
}

class _PerformanceMonitorState extends State<PerformanceMonitor> {
  final Stopwatch _stopwatch = Stopwatch();
  int _buildCount = 0;
  Duration _totalBuildTime = Duration.zero;

  @override
  Widget build(BuildContext context) {
    if (kDebugMode) {
      _stopwatch.start();
      _buildCount++;

      // Schedule a post-frame callback to measure build time
      WidgetsBinding.instance.addPostFrameCallback((_) {
        _stopwatch.stop();
        _totalBuildTime += _stopwatch.elapsed;

        if (_buildCount % 10 == 0) {
          final avgBuildTime = _totalBuildTime.inMicroseconds / _buildCount;
          debugPrint(
            'PerformanceMonitor ${widget.name ?? 'Widget'}: '
            'Avg build time: ${avgBuildTime.toStringAsFixed(2)}μs '
            '($_buildCount builds)',
          );
        }

        _stopwatch.reset();
      });
    }

    return widget.child;
  }
}

// Usage example
class UserListScreen extends StatelessWidget {
  const UserListScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Users')),
      body: PerformanceMonitor(
        name: 'UserList',
        child: Consumer<UserBloc>(
          builder: (context, userBloc, child) {
            final state = userBloc.state;

            if (state.isLoading) {
              return const Center(child: CircularProgressIndicator());
            }

            if (state.error != null) {
              return ErrorWidget(
                error: state.error!,
                onRetry: () => userBloc.add(LoadUsers()),
              );
            }

            return OptimizedUserList(
              users: state.users,
              onUserTap: (user) => _navigateToUserDetail(context, user),
              onUserDelete: (user) => _deleteUser(context, userBloc, user),
            );
          },
        ),
      ),
    );
  }

  void _navigateToUserDetail(BuildContext context, User user) {
    Navigator.of(context).push(
      MaterialPageRoute(
        builder: (context) => UserDetailScreen(userId: user.id),
      ),
    );
  }

  void _deleteUser(BuildContext context, UserBloc userBloc, User user) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Delete User'),
        content: Text('Are you sure you want to delete ${user.name}?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () {
              Navigator.of(context).pop();
              userBloc.add(DeleteUser(user.id));
            },
            child: const Text('Delete'),
          ),
        ],
      ),
    );
  }
}
```

---

## 🔒 Security Best Practices

### Secure Flutter Development

```dart
// Secure storage service
class SecureStorageService {
  static const _storage = FlutterSecureStorage();
  static const _keyPrefix = 'secure_';

  static Future<void> storeSecureData(String key, String value) async {
    try {
      await _storage.write(
        key: '$_keyPrefix$key',
        value: value,
        aOptions: _getAndroidOptions(),
        iOptions: _getIOSOptions(),
      );
    } catch (e) {
      throw SecureStorageException('Failed to store secure data: $e');
    }
  }

  static Future<String?> readSecureData(String key) async {
    try {
      return await _storage.read(
        key: '$_keyPrefix$key',
        aOptions: _getAndroidOptions(),
        iOptions: _getIOSOptions(),
      );
    } catch (e) {
      throw SecureStorageException('Failed to read secure data: $e');
    }
  }

  static Future<void> deleteSecureData(String key) async {
    try {
      await _storage.delete(
        key: '$_keyPrefix$key',
        aOptions: _getAndroidOptions(),
        iOptions: _getIOSOptions(),
      );
    } catch (e) {
      throw SecureStorageException('Failed to delete secure data: $e');
    }
  }

  static AndroidOptions _getAndroidOptions() => const AndroidOptions(
    encryptedSharedPreferences: true,
  );

  static IOSOptions _getIOSOptions() => const IOSOptions(
    accessibility: KeychainAccessibility.first_unlock_this_device,
  );
}

// Network security with certificate pinning
class SecureApiService {
  late Dio _dio;

  SecureApiService() {
    _dio = Dio(BaseOptions(
      baseUrl: 'https://api.enterprise.com',
      connectTimeout: Duration(seconds: 30),
      receiveTimeout: Duration(seconds: 30),
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
      },
    ));

    _setupInterceptors();
    _setupCertificatePinning();
  }

  void _setupInterceptors() {
    _dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (options, handler) {
          // Add authentication token
          final token = SecureStorageService.readSecureData('auth_token');
          if (token != null) {
            options.headers['Authorization'] = 'Bearer $token';
          }

          // Add request ID for tracking
          options.headers['X-Request-ID'] = _generateRequestId();

          handler.next(options);
        },
        onResponse: (response, handler) {
          // Log successful responses (in debug mode only)
          if (kDebugMode) {
            debugPrint('API Response: ${response.statusCode} - ${response.requestOptions.path}');
          }
          handler.next(response);
        },
        onError: (error, handler) {
          // Handle API errors
          if (error.response?.statusCode == 401) {
            // Handle authentication error
            _handleAuthenticationError();
          } else if (error.response?.statusCode == 403) {
            // Handle authorization error
            _handleAuthorizationError();
          }

          handler.next(error);
        },
      ),
    );
  }

  void _setupCertificatePinning() {
    (dio.httpClientAdapter as DefaultHttpClientAdapter).onHttpClientCreate = (client) {
      client.badCertificateCallback = (cert, host, port) {
        // Implement certificate pinning logic
        return _validateCertificate(cert, host);
      };
      return client;
    };
  }

  bool _validateCertificate(X509Certificate cert, String host) {
    // Compare certificate with pinned certificate
    // This is a simplified example - implement proper certificate validation
    return cert.endValidity.isAfter(DateTime.now());
  }

  String _generateRequestId() {
    return '${DateTime.now().millisecondsSinceEpoch}-${Random().nextInt(10000)}';
  }

  void _handleAuthenticationError() {
    // Clear stored credentials
    SecureStorageService.deleteSecureData('auth_token');

    // Navigate to login screen
    NavigationService.instance?.navigateToAndClearStack('/login');
  }

  void _handleAuthorizationError() {
    // Show access denied message
    NavigationService.instance?.showErrorSnackBar('Access denied');
  }

  Future<Response<T>> get<T>(
    String path, {
    Map<String, dynamic>? queryParameters,
    Options? options,
  }) async {
    try {
      return await _dio.get<T>(
        path,
        queryParameters: queryParameters,
        options: options,
      );
    } catch (e) {
      throw ApiException('GET request failed: $e');
    }
  }

  Future<Response<T>> post<T>(
    String path, {
    dynamic data,
    Map<String, dynamic>? queryParameters,
    Options? options,
  }) async {
    try {
      return await _dio.post<T>(
        path,
        data: data,
        queryParameters: queryParameters,
        options: options,
      );
    } catch (e) {
      throw ApiException('POST request failed: $e');
    }
  }
}

// Input validation and sanitization
class InputValidator {
  static const _emailRegex = r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$';
  static const _passwordRegex = r'^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)[a-zA-Z\d@$!%*?&]{8,}$';
  static const _nameRegex = r'^[a-zA-Z\s]{2,50}$';

  static String? validateEmail(String? value) {
    if (value == null || value.isEmpty) {
      return 'Email is required';
    }

    if (!RegExp(_emailRegex).hasMatch(value)) {
      return 'Please enter a valid email address';
    }

    return null;
  }

  static String? validatePassword(String? value) {
    if (value == null || value.isEmpty) {
      return 'Password is required';
    }

    if (value.length < 8) {
      return 'Password must be at least 8 characters long';
    }

    if (!RegExp(_passwordRegex).hasMatch(value)) {
      return 'Password must contain uppercase, lowercase, and numbers';
    }

    return null;
  }

  static String? validateName(String? value) {
    if (value == null || value.isEmpty) {
      return 'Name is required';
    }

    if (!RegExp(_nameRegex).hasMatch(value)) {
      return 'Please enter a valid name (2-50 characters, letters only)';
    }

    return null;
  }

  static String sanitizeInput(String input) {
    return input
        .trim()
        .replaceAll(RegExp(r'<[^>]*>'), '') // Remove HTML tags
        .replaceAll(RegExp(r'[^\w\s@.-]'), ''); // Remove special characters except allowed ones
  }
}
```

---

## 📈 Monitoring & Observability

### Flutter Performance & Analytics

```dart
// Performance monitoring service
class PerformanceMonitor {
  static final PerformanceMonitor _instance = PerformanceMonitor._internal();
  factory PerformanceMonitor() => _instance;
  PerformanceMonitor._internal();

  final Map<String, List<Duration>> _metrics = {};
  final Stopwatch _stopwatch = Stopwatch();

  void startTiming(String operation) {
    _stopwatch.reset();
    _stopwatch.start();
  }

  void endTiming(String operation) {
    _stopwatch.stop();

    _metrics.putIfAbsent(operation, () => []).add(_stopwatch.elapsed);

    // Log if operation takes too long
    if (_stopwatch.elapsed > Duration(seconds: 2)) {
      Logger().warning('Slow operation detected: $operation took ${_stopwatch.elapsed.inMilliseconds}ms');
    }
  }

  Map<String, PerformanceStats> getStats() {
    return _metrics.map((key, values) {
      final totalMs = values.fold<int>(0, (sum, duration) => sum + duration.inMilliseconds);
      final avgMs = totalMs / values.length;
      final maxMs = values.map((d) => d.inMilliseconds).reduce(math.max);
      final minMs = values.map((d) => d.inMilliseconds).reduce(math.min);

      return MapEntry(key, PerformanceStats(
        count: values.length,
        averageMs: avgMs,
        maxMs: maxMs,
        minMs: minMs,
        totalMs: totalMs,
      ));
    });
  }

  void resetMetrics() {
    _metrics.clear();
  }
}

class PerformanceStats {
  final int count;
  final double averageMs;
  final int maxMs;
  final int minMs;
  final int totalMs;

  PerformanceStats({
    required this.count,
    required this.averageMs,
    required this.maxMs,
    required this.minMs,
    required this.totalMs,
  });

  @override
  String toString() {
    return 'PerformanceStats(count: $count, avg: ${averageMs.toStringAsFixed(2)}ms, '
           'max: ${maxMs}ms, min: ${minMs}ms, total: ${totalMs}ms)';
  }
}

// Custom analytics service
class AnalyticsService {
  static final AnalyticsService _instance = AnalyticsService._internal();
  factory AnalyticsService() => _instance;
  AnalyticsService._internal();

  final FirebaseAnalytics? _analytics = FirebaseAnalytics.instance;

  void logScreenView(String screenName, {Map<String, String>? parameters}) {
    _analytics?.logScreenView(
      screenClass: screenName,
      screenName: screenName,
      parameters: parameters,
    );

    // Custom logging
    Logger().info('Screen view: $screenName');
  }

  void logEvent(String name, {Map<String, Object?>? parameters}) {
    _analytics?.logEvent(
      name: name,
      parameters: parameters,
    );

    // Custom logging
    Logger().info('Event: $name - $parameters');
  }

  void logUserProperty(String name, String value) {
    _analytics?.setUserProperty(name: name, value: value);

    // Custom logging
    Logger().info('User property: $name = $value');
  }

  void logError(String error, {String? stackTrace}) {
    _analytics?.logEvent(
      name: 'app_error',
      parameters: {
        'error_message': error,
        'stack_trace': stackTrace ?? '',
      },
    );

    // Custom error logging
    Logger().severe('Error: $error', stackTrace);
  }

  void logPerformance(String operation, int durationMs) {
    _analytics?.logEvent(
      name: 'performance_metric',
      parameters: {
        'operation': operation,
        'duration_ms': durationMs,
      },
    );
  }
}

// Performance monitoring widget
class PerformanceAwareWidget extends StatefulWidget {
  final Widget child;
  final String? widgetName;

  const PerformanceAwareWidget({
    super.key,
    required this.child,
    this.widgetName,
  });

  @override
  State<PerformanceAwareWidget> createState() => _PerformanceAwareWidgetState();
}

class _PerformanceAwareWidgetState extends State<PerformanceAwareWidget> {
  final PerformanceMonitor _monitor = PerformanceMonitor();
  final String _widgetName = 'Widget';

  @override
  void initState() {
    super.initState();
    _monitor.startTiming(widget.widgetName ?? _widgetName);
  }

  @override
  Widget build(BuildContext context) {
    return widget.child;
  }

  @override
  void dispose() {
    _monitor.endTiming(widget.widgetName ?? _widgetName);
    super.dispose();
  }
}
```

---

## 🔄 Context7 MCP Integration

### Real-time Documentation Access

```dart
// Service for accessing latest Dart and Flutter documentation
class FlutterDocumentationService {
  static const dartApi = '/websites/dart_dev';
  static const flutterApi = '/websites/flutter_dev';

  Future<String> getDartNullSafetyDocumentation() async {
    return await Context7Client.shared.getLibraryDocs(
      dartApi,
      topic: 'Dart 3.x null safety flow analysis type promotion',
    );
  }

  Future<String> getFlutterStateManagementDocumentation() async {
    return await Context7Client.shared.getLibraryDocs(
      flutterApi,
      topic: 'Flutter 3.x state management Provider Riverpod BLoC',
    );
  }

  Future<String> getFlutterTestingDocumentation() async {
    return await Context7Client.shared.getLibraryDocs(
      flutterApi,
      topic: 'Flutter widget testing integration tests golden tests',
    );
  }

  Future<String> getLatestDartFeature(String feature) async {
    return await Context7Client.shared.getLibraryDocs(
      dartApi,
      topic: 'Dart 3.x $feature patterns records extensions',
    );
  }

  Future<String> getLatestFlutterFeature(String feature) async {
    return await Context7Client.shared.getLibraryDocs(
      flutterApi,
      topic: 'Flutter 3.x $feature widgets navigation performance',
    );
  }
}

// Documentation-driven development helper
class DocumentationHelper {
  final FlutterDocumentationService _docService = FlutterDocumentationService();

  Future<String> getImplementationGuidance(String pattern) async {
    switch (pattern.toLowerCase()) {
      case 'null safety':
        return await _docService.getDartNullSafetyDocumentation();
      case 'state management':
        return await _docService.getFlutterStateManagementDocumentation();
      case 'testing':
        return await _docService.getFlutterTestingDocumentation();
      case 'patterns':
        return await _docService.getLatestDartFeature('patterns');
      case 'records':
        return await _docService.getLatestDartFeature('records');
      case 'extensions':
        return await _docService.getLatestDartFeature('extensions');
      default:
        return await _docService.getLatestFlutterFeature(pattern);
    }
  }

  Future<void> provideCodeExample(String pattern) async {
    final documentation = await getImplementationGuidance(pattern);

    // Parse documentation for code examples
    final examples = _extractCodeExamples(documentation);

    for (final example in examples) {
      Logger().info('Code example for $pattern:\n$example');
    }
  }

  List<String> _extractCodeExamples(String documentation) {
    // Extract code blocks from documentation
    final regex = RegExp(r'```dart\n(.*?)\n```', dotAll: true);
    final matches = regex.allMatches(documentation);

    return matches.map((match) => match.group(1) ?? '').toList();
  }
}
```

---

## 📚 Progressive Disclosure Examples

### High Freedom (Quick Answer - 15 tokens)
"Use Dart 3.x with Flutter 3.x for cross-platform enterprise applications."

### Medium Freedom (Detailed Guidance - 35 tokens)
"Implement null safety with flow analysis, use Provider/Riverpod for state management, and adopt BLoC for complex state logic."

### Low Freedom (Comprehensive Implementation - 80 tokens)
"Build enterprise apps with sound null safety, pattern matching, records for DTOs, multi-provider architecture, comprehensive widget testing, CI/CD pipeline with Firebase App Distribution, and performance monitoring with custom analytics integration."

---

## 🎯 Works Well With

### Core MoAI Skills
- `Skill("moai-domain-mobile-app")` - Mobile app architecture patterns
- `Skill("moai-domain-security")` - Security best practices
- `Skill("moai-foundation-trust")` - TRUST 5 compliance

### Flutter Ecosystem Technologies
- **State Management**: Provider, Riverpod, BLoC, GetX
- **Networking**: Dio, Retrofit, HTTP
- **Database**: Hive, Moor (Drift), Floor, SQLite
- **UI Components**: Material Design, CupertinoWidgets, Flutter SVG
- **Testing**: Flutter Test, Mockito, Golden Toolkit, Integration Tests

---

## 🚀 Production Deployment

### CI/CD Pipeline Configuration

```yaml
# .github/workflows/flutter.yml
name: Flutter CI/CD

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v4

    - name: Setup Flutter
      uses: subosito/flutter-action@v2
      with:
        flutter-version: '3.10.0'
        channel: 'stable'

    - name: Install dependencies
      run: flutter pub get

    - name: Generate code
      run: flutter packages pub run build_runner build --delete-conflicting-outputs

    - name: Analyze code
      run: flutter analyze

    - name: Run tests
      run: flutter test --coverage

    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v3
      with:
        file: coverage/lcov.info

  build_android:
    needs: test
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v4

    - name: Setup Flutter
      uses: subosito/flutter-action@v2
      with:
        flutter-version: '3.10.0'

    - name: Setup Java
      uses: actions/setup-java@v3
      with:
        distribution: 'temurin'
        java-version: '17'

    - name: Setup Android SDK
      uses: android-actions/setup-android@v2

    - name: Install dependencies
      run: flutter pub get

    - name: Build APK
      run: flutter build apk --release

    - name: Build App Bundle
      run: flutter build appbundle --release

    - name: Upload APK
      uses: actions/upload-artifact@v3
      with:
        name: android-apk
        path: build/app/outputs/flutter-apk/app-release.apk

    - name: Upload App Bundle
      uses: actions/upload-artifact@v3
      with:
        name: android-appbundle
        path: build/app/outputs/bundle/release/app-release.aab

  build_ios:
    needs: test
    runs-on: macos-latest

    steps:
    - uses: actions/checkout@v4

    - name: Setup Flutter
      uses: subosito/flutter-action@v2
      with:
        flutter-version: '3.10.0'

    - name: Install dependencies
      run: flutter pub get

    - name: Build iOS
      run: flutter build ios --release --no-codesign

    - name: Upload iOS build
      uses: actions/upload-artifact@v3
      with:
        name: ios-build
        path: build/ios/iphoneos/Runner.app

  deploy_firebase:
    needs: [build_android, build_ios]
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'

    steps:
    - uses: actions/checkout@v4

    - name: Download Android artifacts
      uses: actions/download-artifact@v3
      with:
        name: android-apk

    - name: Setup Firebase CLI
      uses: wzieba/Firebase-Distribution-Github-Action@v1
      with:
        appId: ${{ secrets.FIREBASE_ANDROID_APP_ID }}
        serviceCredentialsFileContent: ${{ secrets.FIREBASE_SERVICE_ACCOUNT }}
        groups: testers
        releaseNotes: "Automatic deployment from GitHub Actions"
        file: app-release.apk
```

---

## ✅ Quality Assurance Checklist

- [ ] **Dart 3.x Features**: Null safety, patterns, records, extensions
- [ ] **Flutter 3.x Integration**: Modern widgets, Navigator 2.0, performance optimization
- [ ] **Context7 Integration**: Real-time documentation from official sources
- [ ] **State Management**: Provider/Riverpod/BLoC implementation
- [ ] **Security Implementation**: Secure storage, certificate pinning, input validation
- [ ] **Testing Coverage**: Unit tests, widget tests, integration tests, golden tests
- [ ] **Code Quality**: Flutter analyze, custom linting rules, code generation
- [ ] **Production Readiness**: CI/CD pipeline, Firebase distribution, performance monitoring

---

**Last Updated**: 2025-11-06
**Version**: 3.0.0 (Premium Edition - Dart 3.x + Flutter 3.x)
**Context7 Integration**: Fully integrated with Dart and Flutter official APIs
**Status**: Production Ready - Enterprise Grade