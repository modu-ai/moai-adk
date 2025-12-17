---
name: moai-lang-flutter
description: Flutter 3.24+ / Dart 3.5+ development specialist covering Riverpod, go_router, and cross-platform patterns. Use when building cross-platform mobile apps, desktop apps, or web applications with Flutter.
version: 1.0.0
category: language
tags: [flutter, dart, riverpod, cross-platform, mobile, desktop]
context7-libraries: [/flutter/flutter, /rrousselgit/riverpod, /flutter/packages]
related-skills: [moai-lang-swift, moai-lang-kotlin]
updated: 2025-12-07
status: active
---

## Quick Reference (30 seconds)

Flutter/Dart Development Expert - Dart 3.5+, Flutter 3.24+ with modern patterns.

Auto-Triggers: Flutter projects (`.dart` files, `pubspec.yaml`), cross-platform apps, widget development

Core Capabilities:
- Dart 3.5: Patterns, records, sealed classes, extension types
- Flutter 3.24: Widget tree, Material 3, adaptive layouts
- Riverpod: State management with code generation
- go_router: Declarative navigation and deep linking
- Platform Channels: Native iOS/Android integration
- Testing: flutter_test, widget_test, integration_test

## Implementation Guide (5 minutes)

### Dart 3.5 Language Features

Pattern Matching with Sealed Classes:
```dart
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

// Exhaustive switch expression
String handleResult(Result<User> result) => switch (result) {
  Success(:final data) => 'User: ${data.name}',
  Failure(:final error) => 'Error: $error',
};

// Guard clauses in patterns
String describeUser(User user) => switch (user) {
  User(age: var a) when a < 18 => 'Minor',
  User(age: var a) when a >= 65 => 'Senior',
  User(name: var n, age: var a) => '$n, age $a',
};
```

Records and Destructuring:
```dart
typedef UserRecord = ({String name, int age, String email});

// Multiple return values
(String name, int age) parseUser(Map<String, dynamic> json) {
  return (json['name'] as String, json['age'] as int);
}

// Destructuring
void processUser(Map<String, dynamic> json) {
  final (name, age) = parseUser(json);
  print('$name is $age years old');
}

// Record patterns in collections
void processUsers(List<UserRecord> users) {
  for (final (:name, :age, :email) in users) {
    print('$name ($age): $email');
  }
}
```

Extension Types:
```dart
extension type UserId(String value) {
  factory UserId.generate() => UserId(Uuid().v4());
  bool get isValid => value.isNotEmpty;
}

extension type Email(String value) {
  bool get isValid => value.contains('@') && value.contains('.');
  String get domain => value.split('@').last;
}
```

### Riverpod State Management

Provider Definitions:
```dart
import 'package:riverpod_annotation/riverpod_annotation.dart';
part 'providers.g.dart';

@riverpod
UserRepository userRepository(Ref ref) {
  return UserRepository(ref.read(dioProvider));
}

@riverpod
Future<User> user(Ref ref, String userId) async {
  return ref.watch(userRepositoryProvider).getUser(userId);
}

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

  Future<void> updateUser(User user) async {
    state = const AsyncLoading();
    state = await AsyncValue.guard(
      () => ref.read(userRepositoryProvider).updateUser(user),
    );
  }
}

@riverpod
Stream<List<Message>> messages(Ref ref, String chatId) {
  return ref.watch(chatRepositoryProvider).watchMessages(chatId);
}
```

Widget Integration:
```dart
class UserScreen extends ConsumerWidget {
  final String userId;
  const UserScreen({required this.userId, super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final userAsync = ref.watch(userProvider(userId));

    return Scaffold(
      appBar: AppBar(title: const Text('User Profile')),
      body: userAsync.when(
        data: (user) => UserProfile(user: user),
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (error, stack) => Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Text('Error: $error'),
              ElevatedButton(
                onPressed: () => ref.invalidate(userProvider(userId)),
                child: const Text('Retry'),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
```

StatefulWidget with Riverpod:
```dart
class EditUserScreen extends ConsumerStatefulWidget {
  const EditUserScreen({super.key});
  @override
  ConsumerState<EditUserScreen> createState() => _EditUserScreenState();
}

class _EditUserScreenState extends ConsumerState<EditUserScreen> {
  final _formKey = GlobalKey<FormState>();
  late TextEditingController _nameController;

  @override
  void initState() {
    super.initState();
    _nameController = TextEditingController();
  }

  @override
  void dispose() {
    _nameController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    ref.listen(userNotifierProvider, (prev, next) {
      next.whenOrNull(
        error: (e, _) => ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('Error: $e')),
        ),
      );
    });

    final isLoading = ref.watch(userNotifierProvider).isLoading;

    return Form(
      key: _formKey,
      child: Column(
        children: [
          TextFormField(
            controller: _nameController,
            validator: (v) => v?.isEmpty ?? true ? 'Required' : null,
          ),
          ElevatedButton(
            onPressed: isLoading ? null : () async {
              if (_formKey.currentState!.validate()) {
                await ref.read(userNotifierProvider.notifier)
                    .updateUser(User(name: _nameController.text));
              }
            },
            child: isLoading
                ? const CircularProgressIndicator()
                : const Text('Save'),
          ),
        ],
      ),
    );
  }
}
```

### go_router Navigation

Router Configuration:
```dart
final router = GoRouter(
  initialLocation: '/',
  routes: [
    GoRoute(
      path: '/',
      name: 'home',
      builder: (context, state) => const HomeScreen(),
      routes: [
        GoRoute(
          path: 'user/:id',
          name: 'user-detail',
          builder: (context, state) => UserDetailScreen(
            userId: state.pathParameters['id']!,
          ),
        ),
      ],
    ),
    ShellRoute(
      builder: (context, state, child) => MainShell(child: child),
      routes: [
        GoRoute(
          path: '/feed',
          pageBuilder: (_, __) => const NoTransitionPage(child: FeedScreen()),
        ),
        GoRoute(
          path: '/search',
          pageBuilder: (_, __) => const NoTransitionPage(child: SearchScreen()),
        ),
        GoRoute(
          path: '/profile',
          pageBuilder: (_, __) => const NoTransitionPage(child: ProfileScreen()),
        ),
      ],
    ),
  ],
  redirect: (context, state) {
    final isLoggedIn = authNotifier.isLoggedIn;
    final isLoggingIn = state.matchedLocation == '/login';
    if (!isLoggedIn && !isLoggingIn) return '/login';
    if (isLoggedIn && isLoggingIn) return '/';
    return null;
  },
  errorBuilder: (context, state) => ErrorScreen(error: state.error),
);

// Navigation methods
void navigateToUser(BuildContext context, String userId) {
  context.go('/user/$userId');
}

void goBack(BuildContext context) {
  if (context.canPop()) context.pop();
  else context.go('/');
}
```

### Platform Channels

Dart Implementation:
```dart
class NativeBridge {
  static const _channel = MethodChannel('com.example.app/native');
  static const _eventChannel = EventChannel('com.example.app/events');

  Future<String> getPlatformVersion() async {
    try {
      final version = await _channel.invokeMethod<String>('getPlatformVersion');
      return version ?? 'Unknown';
    } on PlatformException catch (e) {
      throw NativeBridgeException('Failed: ${e.message}');
    }
  }

  Future<void> shareContent({required String text, String? title}) async {
    await _channel.invokeMethod('share', {
      'text': text,
      if (title != null) 'title': title,
    });
  }

  Stream<BatteryState> watchBatteryState() {
    return _eventChannel.receiveBroadcastStream().map((event) {
      final data = event as Map<dynamic, dynamic>;
      return BatteryState(
        level: data['level'] as int,
        isCharging: data['isCharging'] as bool,
      );
    });
  }

  void setupMethodCallHandler() {
    _channel.setMethodCallHandler((call) async {
      switch (call.method) {
        case 'onNativeEvent':
          // Handle native event
          return true;
        default:
          throw MissingPluginException('Not implemented: ${call.method}');
      }
    });
  }
}

class BatteryState {
  final int level;
  final bool isCharging;
  const BatteryState({required this.level, required this.isCharging});
}
```

### Widget Patterns

Adaptive Layouts:
```dart
class AdaptiveScaffold extends StatelessWidget {
  final Widget child;
  final List<NavigationDestination> destinations;
  final int selectedIndex;
  final ValueChanged<int> onDestinationSelected;

  const AdaptiveScaffold({
    required this.child,
    required this.destinations,
    required this.selectedIndex,
    required this.onDestinationSelected,
    super.key,
  });

  @override
  Widget build(BuildContext context) {
    final width = MediaQuery.sizeOf(context).width;

    if (width < 600) {
      return Scaffold(
        body: child,
        bottomNavigationBar: NavigationBar(
          selectedIndex: selectedIndex,
          onDestinationSelected: onDestinationSelected,
          destinations: destinations,
        ),
      );
    }

    if (width < 840) {
      return Scaffold(
        body: Row(children: [
          NavigationRail(
            selectedIndex: selectedIndex,
            onDestinationSelected: onDestinationSelected,
            destinations: destinations.map((d) => NavigationRailDestination(
              icon: d.icon, selectedIcon: d.selectedIcon, label: Text(d.label),
            )).toList(),
          ),
          const VerticalDivider(thickness: 1, width: 1),
          Expanded(child: child),
        ]),
      );
    }

    return Scaffold(
      body: Row(children: [
        NavigationDrawer(
          selectedIndex: selectedIndex,
          onDestinationSelected: onDestinationSelected,
          children: destinations.map((d) => NavigationDrawerDestination(
            icon: d.icon, selectedIcon: d.selectedIcon ?? d.icon, label: Text(d.label),
          )).toList(),
        ),
        const VerticalDivider(thickness: 1, width: 1),
        Expanded(child: child),
      ]),
    );
  }
}
```

### Testing

Widget Test Example:
```dart
void main() {
  testWidgets('UserScreen displays data', (tester) async {
    final container = ProviderContainer(overrides: [
      userRepositoryProvider.overrideWithValue(MockUserRepository()),
    ]);

    await tester.pumpWidget(
      UncontrolledProviderScope(
        container: container,
        child: const MaterialApp(home: UserScreen(userId: '1')),
      ),
    );

    expect(find.byType(CircularProgressIndicator), findsOneWidget);
    await tester.pumpAndSettle();
    expect(find.text('Test User'), findsOneWidget);
  });
}
```

For comprehensive testing patterns, see [examples.md](examples.md).

## Advanced Patterns

For comprehensive coverage including:
- Clean Architecture with Riverpod
- Isolates for compute-heavy operations
- Custom render objects and painting
- FFI and platform-specific plugins
- Performance optimization and profiling

See: [reference.md](reference.md) and [examples.md](examples.md)

## Context7 Library Mappings

Flutter/Dart Core:
- `/flutter/flutter` - Flutter framework
- `/dart-lang/sdk` - Dart SDK

State Management:
- `/rrousselGit/riverpod` - Riverpod state management
- `/felangel/bloc` - BLoC pattern

Navigation and Storage:
- `/flutter/packages` - go_router and official packages
- `/cfug/dio` - HTTP client
- `/isar/isar` - NoSQL database

## Works Well With

- `moai-lang-swift` - iOS native integration for platform channels
- `moai-lang-kotlin` - Android native integration for platform channels
- `moai-domain-backend` - API integration and backend communication
- `moai-quality-security` - Mobile security best practices
- `moai-essentials-debug` - Flutter debugging and DevTools

---

Version: 1.0.0
Last Updated: 2025-12-07
Status: Production Ready
