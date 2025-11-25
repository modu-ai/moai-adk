---
name: moai-lang-dart
description: Dart 3.5 enterprise development with Flutter 3.24, advanced async programming, state management, and cross-platform mobile development. Enterprise patterns for scalable applications with Context7 MCP integration.
version: 1.0.0
modularized: false
tags:
  - programming-language
  - enterprise
  - dart
  - development
updated: 2025-11-24
status: active
---

## 📊 Skill Metadata

**version**: 1.0.0  
**modularized**: false  
**last_updated**: 2025-11-22  
**compliance_score**: 75%  
**auto_trigger_keywords**: lang, moai, dart  


## Quick Reference (30 seconds)

# Dart - Enterprise 


## Quick Reference

### Null Safety (Dart 3.5)

```dart
// Non-nullable by default
String name = 'Alice';

// Nullable with explicit ?
String? email;

// Late initialization
late String value;

// Null-coalescing operator
String display = email ?? 'No email';

// Safe navigation
int? length = email?.length;
```

### Async Programming

```dart
// Modern async/await
Future<String> fetchData() async {
  await Future.delayed(Duration(seconds: 1));
  return 'Data';
}

// Stream for continuous values
Stream<int> counter() async* {
  for (int i = 0; i < 5; i++) {
    yield i;
  }
}

// Error handling
try {
  final data = await fetchData();
} catch (e) {
  print('Error: $e');
}
```

### Flutter Widgets

```dart
// StatelessWidget - immutable UI
class MyWidget extends StatelessWidget {
  @override
  Widget build(BuildContext context) => Text('Hello');
}

// StatefulWidget - mutable state
class Counter extends StatefulWidget {
  @override
  State<Counter> createState() => _CounterState();
}

class _CounterState extends State<Counter> {
  int count = 0;

  @override
  Widget build(BuildContext context) {
    return FloatingActionButton(
      onPressed: () => setState(() => count++),
      child: Text('$count'),
    );
  }
}
```

### State Management Pattern Selection

```
Simple UI          → setState
Single Feature     → Provider / GetX
Complex App        → Riverpod / BLoC
Enterprise Scale   → BLoC + Repository Pattern
```


## Version History

| Version | Date | Status | Notes |
|---------|------|--------|-------|
| 4.0.0 | 2025-11-12 | Current | Enterprise v4 restructure, Context7 integration |
| 3.5.4 | 2025-10-22 | Previous | Advanced patterns focus |
| 3.0.0 | 2025-09-01 | Legacy | Initial release |


**For detailed examples** → See `examples.md`
**For API reference** → See `reference.md`
**For hands-on patterns** → See `examples.md` Part 3-4


## Implementation Guide

## When to Use

**Automatic triggers**:
- Dart and Flutter development discussions
- Cross-platform mobile application development
- State management pattern implementation
- Async programming and stream handling
- Mobile app architecture and design
- Enterprise Flutter application development

**Manual invocation**:
- Design mobile application architecture
- Implement advanced async patterns
- Optimize Flutter app performance
- Review enterprise Dart/Flutter code
- Implement state management solutions
- Troubleshoot mobile development issues


## Technology Stack (2025-11-12)

| Component | Version | Purpose | Status |
|-----------|---------|---------|--------|
| **Dart** | 3.6.0 | Core language | ✅ Current |
| **Flutter** | 3.24.0 | UI framework | ✅ Current |
| **Riverpod** | 2.6.1 | State management | ✅ Current |
| **BLoC** | 8.1.6 | State management | ✅ Current |
| **Dio** | 5.6.0 | HTTP client | ✅ Current |
| **Provider** | 6.4.0 | State management | ✅ Current |
| **Firebase Core** | 2.28.0 | Backend | ✅ Current |


## Implementation Patterns

### Pattern 1: Null Safety Best Practices

```dart
// Type safety ensures null checks at compile time
class User {
  final String id;
  final String? optionalEmail;

  User({required this.id, this.optionalEmail});

  // Safe method with null-coalescing
  String getEmail() => optionalEmail ?? 'no-email@example.com';

  // Null-aware property access
  bool hasEmail() => optionalEmail?.isNotEmpty ?? false;
}

// Usage ensures type safety
final user = User(id: '1', optionalEmail: 'test@example.com');
print(user.getEmail());     // Type-safe string
print(user.hasEmail());     // Type-safe boolean
```

### Pattern 2: Async Stream Handling

```dart
// Generator function for streams
Stream<String> dataStream() async* {
  for (int i = 0; i < 5; i++) {
    yield 'Item $i';
    await Future.delayed(Duration(seconds: 1));
  }
}

// Listening with error handling
dataStream().listen(
  (value) => print(value),
  onError: (error) => print('Error: $error'),
  onDone: () => print('Complete'),
  cancelOnError: false,
);
```

### Pattern 3: Provider State Management

```dart
// Simple provider for state
class CounterProvider extends ChangeNotifier {
  int _count = 0;
  int get count => _count;

  void increment() {
    _count++;
    notifyListeners();
  }
}

// Using Provider in widget
Consumer<CounterProvider>(
  builder: (context, counter, child) {
    return Text('Count: ${counter.count}');
  },
)
```

### Pattern 4: Repository Pattern

```dart
// Data layer abstraction
abstract class UserRepository {
  Future<User> getUser(String id);
  Future<void> saveUser(User user);
}

// Implementation
class UserRepositoryImpl implements UserRepository {
  final apiClient = ApiClient();

  @override
  Future<User> getUser(String id) async {
    final response = await apiClient.get('/users/$id');
    return User.fromJson(response);
  }

  @override
  Future<void> saveUser(User user) async {
    await apiClient.post('/users', data: user.toJson());
  }
}
```

### Pattern 5: Widget Architecture

```dart
// Composition-based widget tree
class AppScreen extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('Title')),
      body: ListView(
        children: [
          HeaderSection(),
          ContentSection(),
          FooterSection(),
        ],
      ),
    );
  }
}
```


## Common Use Cases

### Mobile App with API Integration

1. Define models and repository layer
2. Implement state management (Riverpod/Provider)
3. Create service layer for API calls
4. Build UI widgets consuming state
5. Add comprehensive error handling
6. Test with mocks and fixtures

### Cross-Platform App (iOS + Android + Web)

1. Use Flutter for shared UI code
2. Platform channels for native code
3. State management for feature consistency
4. Responsive design with MediaQuery
5. Test on all target platforms

### Real-Time Data (Streams)

1. Define stream-based providers
2. Listen to data updates in UI
3. Implement unsubscribe on disposal
4. Handle connection errors gracefully
5. Buffer/throttle high-frequency updates


## Enterprise Checklist

- [ ] Null safety enabled (`sdk: '>=3.0.0'`)
- [ ] Clean architecture with separation of concerns
- [ ] Repository pattern for data access
- [ ] State management with proper scoping
- [ ] Comprehensive error handling
- [ ] Logging and monitoring in place
- [ ] Unit test coverage ≥85%
- [ ] Widget tests for critical UI
- [ ] Integration tests for key flows
- [ ] Performance profiling completed
- [ ] Security review (data encryption, API auth)
- [ ] Documentation and code examples provided
- [ ] CI/CD pipeline configured
- [ ] Crash reporting integrated


## Common Pitfalls

### ❌ Avoid

```dart
// Throwing unhandled exceptions
Future<Data> fetchData() async => apiCall();  // No error handling

// Memory leaks in streams
controller.stream.listen(...);  // Not cancelled

// Mutable state in build()
int counter = 0;  // Recreated on each build
```

### ✅ Do Instead

```dart
// Handle errors gracefully
Future<Data> fetchData() async {
  try { return await apiCall(); }
  catch (e) { return Data.empty(); }
}

// Cancel subscriptions
StreamSubscription sub = stream.listen(...);
@override void dispose() { sub.cancel(); }

// Use State/Provider for mutable data
class Counter extends StatefulWidget { ... }
```


## Resources

**Official Documentation**:
- Dart Guide: https://dart.dev
- Flutter Guide: https://flutter.dev
- Packages: https://pub.dev

**Learning**:
- Effective Dart: https://dart.dev/guides/language/effective-dart
- Flutter Codelabs: https://flutter.dev/codelabs
- Riverpod Tutorial: https://riverpod.dev

**Tools**:
- Flutter DevTools: DevTools for debugging
- Dart Analyzer: Static analysis
- Coverage Tools: lcov for coverage reporting


## Context7 Integration

This skill integrates with **Context7 MCP** for real-time documentation:

```
Keyword Detection → Context7 Query → Real-Time API Docs
  "async/await"   → Dart async guide
  "Flutter widget" → Flutter widget docs
  "Riverpod"     → Riverpod state management
  "Firebase"     → Firebase integration guide
```

**Enable with**: `mcp__context7__resolve-library-id` + `mcp__context7__get-library-docs`



## Advanced Patterns

## What It Does

Dart 3.5 enterprise development featuring Flutter 3.24, advanced async programming patterns, modern state management (BLoC, Riverpod, Provider), and enterprise-grade cross-platform mobile development. Context7 MCP integration provides real-time access to official Dart and Flutter documentation.

**Key capabilities**:
- ✅ Dart 3.5 with advanced type system and patterns
- ✅ Flutter 3.24 enterprise mobile development
- ✅ Advanced async programming with Isolates and Streams
- ✅ Modern state management (BLoC, Riverpod, Provider)
- ✅ Cross-platform development (iOS, Android, Web, Desktop)
- ✅ Enterprise architecture patterns (Clean Architecture, MVVM)
- ✅ Performance optimization and memory management
- ✅ Testing strategies with unit, widget, and integration tests
- ✅ Context7 MCP integration for real-time documentation


## Advanced Topics

### Error Handling Strategy

```
User Input    → Validation errors → Show in UI
Network Call  → Timeout/connection → Retry with exponential backoff
Data Parse    → Format errors → Log and use default
State Update  → Concurrent updates → Use immutable state pattern
```

### Performance Optimization

- Use `const` constructors for immutable widgets
- Implement `ListView.builder` instead of `ListView` for large lists
- Cache expensive computations with memoization
- Profile with DevTools to identify bottlenecks
- Use `RepaintBoundary` for complex widget trees

### Testing Strategy

- **Unit Tests**: Business logic (90%+ coverage target)
- **Widget Tests**: UI components and interactions
- **Integration Tests**: Full user flows and scenarios
- **Mocking**: External dependencies with Mockito


