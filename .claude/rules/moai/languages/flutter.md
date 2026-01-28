---
paths:
  - "**/*.dart"
  - "**/pubspec.yaml"
  - "**/pubspec.lock"
---

# Flutter/Dart Rules

Version: Flutter 3.24+ / Dart 3.5+

## Tooling

- Linting: dart analyze with flutter_lints
- Formatting: dart format
- Testing: flutter test with coverage >= 85%

## Best Practices (2026)

- Use Riverpod 3.0 for state management
- Use go_router for type-safe navigation
- Prefer const constructors for performance
- Use freezed for immutable data classes
- Use drift for offline-first SQLite persistence

## Riverpod 3.0 Patterns

```dart
// Code generation for type safety
@riverpod
class Counter extends _$Counter {
  @override
  int build() => 0;

  void increment() => state++;
}

// Async data with automatic caching
@riverpod
Future<User> user(UserRef ref, {required int id}) async {
  return ref.watch(apiProvider).fetchUser(id);
}
```

## Offline-First Patterns

```dart
// Drift for local SQLite database
@DriftDatabase(tables: [Users, Tasks])
class AppDatabase extends _$AppDatabase {
  AppDatabase() : super(_openConnection());

  // Auto-sync when online
  Future<void> syncWithServer() async {
    final localChanges = await getUnsynced();
    await api.sync(localChanges);
  }
}
```

## Navigation

```dart
// Type-safe routing with go_router
final router = GoRouter(
  routes: [
    GoRoute(
      path: '/user/:id',
      builder: (context, state) => UserScreen(
        id: int.parse(state.pathParameters['id']!),
      ),
    ),
  ],
);
```

## Performance

- Use `const` widgets to prevent rebuilds
- Use `ListView.builder` for long lists
- Profile with DevTools
- Minimize widget rebuilds with selective providers

## MoAI Integration

- Use Skill("moai-lang-flutter") for detailed patterns
- Follow TRUST 5 quality gates
