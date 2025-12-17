# Flutter/Dart Reference Guide

## Platform Version Matrix

### Dart 3.5 (November 2025)
- Release: November 2025
- Key Features:
  - Pattern matching with exhaustiveness checking
  - Records and destructuring
  - Sealed classes and interfaces
  - Extension types with zero runtime cost
  - Macros (experimental)
  - Enhanced class modifiers (sealed, final, base, interface)

### Flutter 3.24 (November 2025)
- Release: November 2025
- Minimum OS: iOS 12.0+, Android API 21+
- Key Features:
  - Material 3 design system
  - Impeller rendering engine (default on iOS)
  - Adaptive layouts and responsive design
  - Enhanced platform views
  - Improved hot reload and DevTools

## Context7 Library Mappings

### Core Framework
```
/flutter/flutter              - Flutter framework and engine
/dart-lang/sdk                - Dart language SDK
/flutter/packages             - Official Flutter packages (go_router, etc.)
```

### State Management
```
/rrousselGit/riverpod         - Riverpod state management
/felangel/bloc                - BLoC pattern library
/jonataslaw/getx              - GetX framework
/alibaba/flutter_redux        - Redux for Flutter
```

### Navigation
```
/flutter/packages             - go_router official package
/theyakka/fluro               - Fluro router
```

### Networking
```
/cfug/dio                     - Dio HTTP client
/nickmeinhold/chopper         - Chopper HTTP client generator
```

### Storage
```
/isar/isar                    - Isar NoSQL database
/simonbengtsson/drift         - Drift SQL database
/nickmeinhold/hive            - Hive key-value storage
/nickmeinhold/shared_preferences - SharedPreferences wrapper
```

### UI Components
```
/Sub6Resources/flutter_html   - HTML rendering
/nickmeinhold/cached_network_image - Image caching
/nickmeinhold/flutter_svg     - SVG rendering
```

### Testing
```
/flutter/flutter              - flutter_test (built-in)
/dart-lang/mockito            - Mockito for Dart
/nickmeinhold/mocktail        - Mocktail testing
```

## Architecture Patterns

### Clean Architecture with Riverpod

Layer Organization:
```dart
// lib/
// ├── core/
// │   ├── error/
// │   │   ├── exceptions.dart
// │   │   └── failures.dart
// │   ├── network/
// │   │   └── api_client.dart
// │   └── utils/
// │       └── extensions.dart
// ├── features/
// │   └── user/
// │       ├── data/
// │       │   ├── datasources/
// │       │   │   ├── user_local_datasource.dart
// │       │   │   └── user_remote_datasource.dart
// │       │   ├── models/
// │       │   │   └── user_model.dart
// │       │   └── repositories/
// │       │       └── user_repository_impl.dart
// │       ├── domain/
// │       │   ├── entities/
// │       │   │   └── user.dart
// │       │   ├── repositories/
// │       │   │   └── user_repository.dart
// │       │   └── usecases/
// │       │       ├── get_user.dart
// │       │       └── update_user.dart
// │       └── presentation/
// │           ├── providers/
// │           │   └── user_provider.dart
// │           ├── screens/
// │           │   └── user_screen.dart
// │           └── widgets/
// │               └── user_card.dart
// └── main.dart

// Domain Layer - Entity
class User {
  final String id;
  final String name;
  final String email;
  final DateTime? createdAt;

  const User({
    required this.id,
    required this.name,
    required this.email,
    this.createdAt,
  });
}

// Domain Layer - Repository Interface
abstract class UserRepository {
  Future<User> getUser(String id);
  Future<User> updateUser(User user);
  Stream<User> watchUser(String id);
}

// Domain Layer - Use Case
class GetUserUseCase {
  final UserRepository _repository;

  const GetUserUseCase(this._repository);

  Future<User> call(String id) => _repository.getUser(id);
}

// Data Layer - Model
class UserModel extends User {
  const UserModel({
    required super.id,
    required super.name,
    required super.email,
    super.createdAt,
  });

  factory UserModel.fromJson(Map<String, dynamic> json) => UserModel(
    id: json['id'] as String,
    name: json['name'] as String,
    email: json['email'] as String,
    createdAt: json['created_at'] != null
        ? DateTime.parse(json['created_at'] as String)
        : null,
  );

  Map<String, dynamic> toJson() => {
    'id': id,
    'name': name,
    'email': email,
    if (createdAt != null) 'created_at': createdAt!.toIso8601String(),
  };
}

// Data Layer - Repository Implementation
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
  Future<User> updateUser(User user) async {
    final updated = await _remote.updateUser(user as UserModel);
    await _local.cacheUser(updated);
    return updated;
  }

  @override
  Stream<User> watchUser(String id) => _local.watchUser(id);
}

// Presentation Layer - Provider
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

  Future<void> update(User user) async {
    state = const AsyncLoading();
    state = await AsyncValue.guard(
      () => ref.read(userRepositoryProvider).updateUser(user),
    );
  }
}
```

## Concurrency Patterns

### Isolates for Heavy Computation

Simple Compute:
```dart
// Using compute for one-off operations
Future<List<ProcessedItem>> processItems(List<RawItem> items) async {
  return compute(_processItemsIsolate, items);
}

List<ProcessedItem> _processItemsIsolate(List<RawItem> items) {
  return items.map((item) => ProcessedItem.from(item)).toList();
}
```

Long-Running Isolate:
```dart
class ImageProcessor {
  late final Isolate _isolate;
  late final ReceivePort _receivePort;
  late final SendPort _sendPort;
  bool _isInitialized = false;

  Future<void> spawn() async {
    if (_isInitialized) return;

    _receivePort = ReceivePort();
    _isolate = await Isolate.spawn(
      _isolateEntry,
      _receivePort.sendPort,
    );
    _sendPort = await _receivePort.first as SendPort;
    _isInitialized = true;
  }

  Future<Uint8List> processImage(Uint8List imageData) async {
    if (!_isInitialized) {
      throw StateError('ImageProcessor not initialized. Call spawn() first.');
    }

    final responsePort = ReceivePort();
    _sendPort.send(_ProcessImageRequest(imageData, responsePort.sendPort));
    final result = await responsePort.first;

    if (result is _ProcessImageError) {
      throw result.error;
    }
    return result as Uint8List;
  }

  void dispose() {
    _isolate.kill(priority: Isolate.immediate);
    _receivePort.close();
  }

  static void _isolateEntry(SendPort sendPort) {
    final receivePort = ReceivePort();
    sendPort.send(receivePort.sendPort);

    receivePort.listen((message) {
      if (message is _ProcessImageRequest) {
        try {
          final processed = _heavyImageProcessing(message.imageData);
          message.replyPort.send(processed);
        } catch (e) {
          message.replyPort.send(_ProcessImageError(e));
        }
      }
    });
  }

  static Uint8List _heavyImageProcessing(Uint8List data) {
    // Heavy computation here
    return data;
  }
}

class _ProcessImageRequest {
  final Uint8List imageData;
  final SendPort replyPort;
  _ProcessImageRequest(this.imageData, this.replyPort);
}

class _ProcessImageError {
  final Object error;
  _ProcessImageError(this.error);
}
```

## Navigation Patterns

### Advanced go_router Configuration

Nested Navigation:
```dart
final router = GoRouter(
  initialLocation: '/',
  routes: [
    // Root route
    GoRoute(
      path: '/',
      redirect: (context, state) => '/home',
    ),

    // Main app with bottom navigation
    StatefulShellRoute.indexedStack(
      builder: (context, state, navigationShell) {
        return MainScaffold(navigationShell: navigationShell);
      },
      branches: [
        StatefulShellBranch(
          routes: [
            GoRoute(
              path: '/home',
              name: 'home',
              builder: (context, state) => const HomeScreen(),
              routes: [
                GoRoute(
                  path: 'detail/:id',
                  name: 'home-detail',
                  builder: (context, state) => DetailScreen(
                    id: state.pathParameters['id']!,
                  ),
                ),
              ],
            ),
          ],
        ),
        StatefulShellBranch(
          routes: [
            GoRoute(
              path: '/search',
              name: 'search',
              builder: (context, state) => const SearchScreen(),
            ),
          ],
        ),
        StatefulShellBranch(
          routes: [
            GoRoute(
              path: '/profile',
              name: 'profile',
              builder: (context, state) => const ProfileScreen(),
            ),
          ],
        ),
      ],
    ),

    // Auth routes (outside main scaffold)
    GoRoute(
      path: '/login',
      name: 'login',
      builder: (context, state) => const LoginScreen(),
    ),
    GoRoute(
      path: '/register',
      name: 'register',
      builder: (context, state) => const RegisterScreen(),
    ),
  ],

  // Global redirect for auth
  redirect: (context, state) {
    final isLoggedIn = ref.read(authStateProvider).valueOrNull != null;
    final isAuthRoute = state.matchedLocation.startsWith('/login') ||
        state.matchedLocation.startsWith('/register');

    if (!isLoggedIn && !isAuthRoute) {
      return '/login?redirect=${state.uri}';
    }

    if (isLoggedIn && isAuthRoute) {
      final redirect = state.uri.queryParameters['redirect'];
      return redirect ?? '/home';
    }

    return null;
  },
);
```

Deep Linking Configuration:
```dart
// android/app/src/main/AndroidManifest.xml
// Add intent filters for deep linking

// iOS: ios/Runner/Info.plist
// Add URL schemes and associated domains

// Router configuration for deep links
GoRouter(
  routes: [
    GoRoute(
      path: '/product/:id',
      builder: (context, state) {
        final productId = state.pathParameters['id']!;
        final query = state.uri.queryParameters;
        return ProductScreen(
          id: productId,
          variant: query['variant'],
        );
      },
    ),
  ],
);
```

## Performance Optimization

### Widget Optimization

Const Constructors:
```dart
// Good - uses const
class MyWidget extends StatelessWidget {
  const MyWidget({super.key}); // const constructor

  @override
  Widget build(BuildContext context) {
    return const Column(
      children: [
        Text('Static text'), // const widget
        Icon(Icons.star),    // const widget
      ],
    );
  }
}

// Usage
const MyWidget(); // const instantiation
```

RepaintBoundary:
```dart
class ComplexAnimatedWidget extends StatelessWidget {
  const ComplexAnimatedWidget({super.key});

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        // Static header - doesn't need repaint
        const Header(),

        // Animated content - isolated repaint
        RepaintBoundary(
          child: AnimatedContent(),
        ),

        // Static footer - doesn't need repaint
        const Footer(),
      ],
    );
  }
}
```

ListView Optimization:
```dart
class OptimizedList extends StatelessWidget {
  final List<Item> items;
  const OptimizedList({required this.items, super.key});

  @override
  Widget build(BuildContext context) {
    return ListView.builder(
      itemCount: items.length,
      // Add extent for better performance
      itemExtent: 72.0,
      // Or use prototypeItem for variable heights
      // prototypeItem: const ItemCard(item: prototypeItem),
      itemBuilder: (context, index) {
        final item = items[index];
        return ItemCard(
          key: ValueKey(item.id), // Stable key
          item: item,
        );
      },
    );
  }
}
```

### Memory Management

Image Caching:
```dart
class ImageCacheManager {
  static final ImageCacheManager _instance = ImageCacheManager._();
  static ImageCacheManager get instance => _instance;

  ImageCacheManager._();

  void configureCache() {
    // Configure image cache size
    PaintingBinding.instance.imageCache.maximumSize = 100;
    PaintingBinding.instance.imageCache.maximumSizeBytes = 50 << 20; // 50 MB
  }

  void clearCache() {
    PaintingBinding.instance.imageCache.clear();
    PaintingBinding.instance.imageCache.clearLiveImages();
  }
}
```

Provider Disposal:
```dart
@riverpod
class ResourceController extends _$ResourceController {
  StreamSubscription? _subscription;
  Timer? _timer;

  @override
  FutureOr<Resource?> build() {
    // Setup
    _subscription = someStream.listen(_onData);
    _timer = Timer.periodic(
      const Duration(seconds: 30),
      (_) => _refresh(),
    );

    // Cleanup on disposal
    ref.onDispose(() {
      _subscription?.cancel();
      _timer?.cancel();
    });

    return null;
  }

  void _onData(dynamic data) {
    // Handle data
  }

  void _refresh() {
    // Refresh logic
  }
}
```

## Testing Frameworks

### flutter_test Configuration
```dart
// test/widget_test.dart
void main() {
  group('UserScreen', () {
    late ProviderContainer container;

    setUp(() {
      container = ProviderContainer(overrides: [
        userRepositoryProvider.overrideWithValue(MockUserRepository()),
      ]);
    });

    tearDown(() => container.dispose());

    testWidgets('renders user name', (tester) async {
      await tester.pumpWidget(
        UncontrolledProviderScope(
          container: container,
          child: const MaterialApp(home: UserScreen(userId: '1')),
        ),
      );

      await tester.pumpAndSettle();
      expect(find.text('Test User'), findsOneWidget);
    });
  });
}
```

### Integration Testing
```dart
// integration_test/app_test.dart
import 'package:integration_test/integration_test.dart';

void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  group('end-to-end test', () {
    testWidgets('login and view profile', (tester) async {
      await tester.pumpWidget(const MyApp());

      // Enter credentials
      await tester.enterText(
        find.byKey(const Key('email_field')),
        'test@example.com',
      );
      await tester.enterText(
        find.byKey(const Key('password_field')),
        'password123',
      );

      // Tap login button
      await tester.tap(find.byKey(const Key('login_button')));
      await tester.pumpAndSettle();

      // Verify navigation to home
      expect(find.byType(HomeScreen), findsOneWidget);

      // Navigate to profile
      await tester.tap(find.byIcon(Icons.person));
      await tester.pumpAndSettle();

      // Verify profile screen
      expect(find.text('Test User'), findsOneWidget);
    });
  });
}
```

### Golden Tests
```dart
// test/golden_test.dart
void main() {
  testWidgets('UserCard golden test', (tester) async {
    await tester.pumpWidget(
      MaterialApp(
        theme: ThemeData.light(),
        home: Scaffold(
          body: UserCard(
            user: const User(
              id: '1',
              name: 'Test User',
              email: 'test@example.com',
            ),
          ),
        ),
      ),
    );

    await expectLater(
      find.byType(UserCard),
      matchesGoldenFile('goldens/user_card.png'),
    );
  });
}

// Run with: flutter test --update-goldens
```

---

Version: 1.0.0
Last Updated: 2025-12-07
