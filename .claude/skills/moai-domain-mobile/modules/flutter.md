# Flutter Development

Dart, Riverpod, go_router, and cross-platform patterns for Flutter development.

## Quick Reference

Key Technologies:
- **Language**: Dart 3.x (sound null safety, records, patterns)
- **Framework**: Flutter 3.x
- **State**: Riverpod 2.x (primary), BLoC 8.x (complex state machines), Provider (legacy)
- **Navigation**: go_router 13.x (URL-based, deep links)
- **Networking**: Dio 5.x with optional Retrofit-style code generation
- **Platform bridge**: `MethodChannel` / `EventChannel` for iOS/Android native code

---

## Core Patterns

### App Entry Point

```dart
// main.dart
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'router.dart';

void main() {
  runApp(const ProviderScope(child: MyApp()));
}

class MyApp extends ConsumerWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final router = ref.watch(routerProvider);
    return MaterialApp.router(
      routerConfig: router,
      title: 'My App',
    );
  }
}
```

### Navigation with go_router

```dart
// router.dart
final routerProvider = Provider<GoRouter>((ref) {
  return GoRouter(
    initialLocation: '/home',
    routes: [
      GoRoute(
        path: '/home',
        builder: (context, state) => const HomeScreen(),
      ),
      GoRoute(
        path: '/user/:id',
        builder: (context, state) {
          final id = state.pathParameters['id']!;
          return UserDetailScreen(userId: id);
        },
      ),
    ],
  );
});
```

### State with Riverpod AsyncNotifier

```dart
// user_provider.dart
@riverpod
class UserList extends _$UserList {
  @override
  Future<List<User>> build() async {
    return ref.read(userRepositoryProvider).getUsers();
  }

  Future<void> refresh() async {
    state = const AsyncLoading();
    state = await AsyncValue.guard(() =>
      ref.read(userRepositoryProvider).getUsers()
    );
  }
}

// Usage in widget
class UserListScreen extends ConsumerWidget {
  const UserListScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final usersAsync = ref.watch(userListProvider);

    return usersAsync.when(
      data: (users) => ListView.builder(
        itemCount: users.length,
        itemBuilder: (_, i) => ListTile(title: Text(users[i].name)),
      ),
      loading: () => const CircularProgressIndicator(),
      error: (e, _) => Text('Error: $e'),
    );
  }
}
```

### Platform Channel Bridge

```dart
// Native feature access via MethodChannel
class BiometricService {
  static const _channel = MethodChannel('com.example.app/biometric');

  Future<bool> authenticate() async {
    try {
      return await _channel.invokeMethod<bool>('authenticate') ?? false;
    } on PlatformException catch (e) {
      debugPrint('Biometric error: ${e.message}');
      return false;
    }
  }
}
```

---

## Works Well With

- expert-mobile (primary agent)
- moai-domain-mobile/modules/strategy-comparison.md (paradigm selection)
- expert-security (flutter_secure_storage, certificate pinning with dio_certificate_pinning)
- moai-platform-auth (Firebase Auth integration with flutter_riverpod)
