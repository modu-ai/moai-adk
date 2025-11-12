# moai-lang-dart - Working Examples

_Last updated: 2025-11-12_

Dart 3.5 and Flutter 3.24 enterprise development with practical, production-ready examples covering async programming, state management, widget architecture, and cross-platform development.

---

## Part 1: Quick Start (50 lines)

### Example 1: Dart 3.5 Project Setup

```bash
# Install Dart SDK (macOS with Homebrew)
brew install dart

# Verify installation
dart --version

# Create new console project
dart create my_console_app

# Create Flutter project
flutter create my_flutter_app

# Get dependencies
cd my_flutter_app
flutter pub get

# Run tests
flutter test
```

### Example 2: Basic Null Safety and Type System

```dart
// Dart 3.5 null safety patterns
void main() {
  // Non-nullable by default
  String name = 'Alice';

  // Nullable with ?
  String? optionalName;

  // Late initialization
  late String lazyValue;
  lazyValue = 'initialized later';

  // Null coalescing
  String displayName = optionalName ?? 'Guest';

  // Null-aware operations
  int? length = optionalName?.length;

  print('Name: $name, Display: $displayName, Length: $length');
}
```

---

## Part 2: Basic Usage (150 lines)

### Example 3: Null Safety Best Practices

```dart
// Comprehensive null safety patterns

class User {
  final String id;
  final String name;
  final String? email;
  final DateTime? lastLoginDate;

  User({
    required this.id,
    required this.name,
    this.email,
    this.lastLoginDate,
  });

  // Safe method with null checking
  String getContactInfo() {
    if (email == null) {
      return 'No email provided';
    }
    return 'Email: $email';
  }

  // Using null-coalescing operator
  String getDisplayName() => name.isEmpty ? 'Anonymous' : name;

  // Safe chaining
  bool hasRecentLogin({required Duration within}) {
    return lastLoginDate?.isAfter(
      DateTime.now().subtract(within),
    ) ?? false;
  }

  // Null-aware method call
  Future<void> syncIfNeeded() async {
    await lastLoginDate?.add(Duration(hours: 1));
  }
}

void main() async {
  final user = User(
    id: '1',
    name: 'Alice',
    email: 'alice@example.com',
  );

  print(user.getContactInfo());
  print(user.getDisplayName());
  print('Has recent login: ${user.hasRecentLogin(within: Duration(days: 7))}');
}
```

### Example 4: Future & Async/Await Patterns

```dart
// Modern async programming with Future and async/await

class DataService {
  // Simulate API call
  Future<String> fetchData(String id) async {
    await Future.delayed(Duration(seconds: 1));

    if (id.isEmpty) {
      throw ArgumentError('ID cannot be empty');
    }

    return 'Data for $id';
  }

  // Multiple concurrent requests
  Future<List<String>> fetchMultipleData(List<String> ids) async {
    try {
      // Wait for all futures simultaneously
      final results = await Future.wait(
        ids.map((id) => fetchData(id)),
        eagerError: false,
      );
      return results;
    } catch (e) {
      print('Error fetching multiple data: $e');
      return [];
    }
  }

  // Timeout handling
  Future<String> fetchWithTimeout(String id) async {
    try {
      return await fetchData(id).timeout(
        Duration(seconds: 5),
        onTimeout: () => throw TimeoutException('Request timeout'),
      );
    } on TimeoutException catch (e) {
      print('Timeout: $e');
      return 'Default data';
    }
  }
}

void main() async {
  final service = DataService();

  // Single request
  try {
    final data = await service.fetchData('123');
    print('Data: $data');
  } catch (e) {
    print('Error: $e');
  }

  // Multiple concurrent requests
  final multipleData = await service.fetchMultipleData(['1', '2', '3']);
  print('Multiple data: $multipleData');
}
```

### Example 5: Flutter First App with Counter

```dart
// Basic Flutter app: Counter with State Management

import 'package:flutter/material.dart';

void main() {
  runApp(const CounterApp());
}

class CounterApp extends StatelessWidget {
  const CounterApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'Counter App',
      theme: ThemeData(
        colorScheme: ColorScheme.fromSeed(
          seedColor: Colors.deepPurple,
        ),
        useMaterial3: true,
      ),
      home: const CounterPage(),
    );
  }
}

class CounterPage extends StatefulWidget {
  const CounterPage({super.key});

  @override
  State<CounterPage> createState() => _CounterPageState();
}

class _CounterPageState extends State<CounterPage> {
  int _counter = 0;

  void _incrementCounter() {
    setState(() {
      _counter++;
    });
  }

  void _decrementCounter() {
    setState(() {
      if (_counter > 0) {
        _counter--;
      }
    });
  }

  void _resetCounter() {
    setState(() {
      _counter = 0;
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Counter App'),
        centerTitle: true,
        elevation: 2,
      ),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Text(
              'Current Count:',
              style: TextStyle(fontSize: 16),
            ),
            const SizedBox(height: 16),
            Text(
              '$_counter',
              style: Theme.of(context).textTheme.displayLarge?.copyWith(
                color: Colors.deepPurple,
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 48),
            Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                FloatingActionButton(
                  onPressed: _decrementCounter,
                  tooltip: 'Decrement',
                  child: const Icon(Icons.remove),
                ),
                const SizedBox(width: 16),
                FloatingActionButton(
                  onPressed: _resetCounter,
                  tooltip: 'Reset',
                  child: const Icon(Icons.refresh),
                ),
                const SizedBox(width: 16),
                FloatingActionButton(
                  onPressed: _incrementCounter,
                  tooltip: 'Increment',
                  child: const Icon(Icons.add),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
```

---

## Part 3: Intermediate Patterns (200 lines)

### Example 6: Stream Programming

```dart
// Stream and reactive programming patterns

class StreamService {
  // Create a simple stream
  Stream<int> countdownStream(int count) async* {
    for (int i = count; i >= 0; i--) {
      yield i;
      await Future.delayed(Duration(seconds: 1));
    }
  }

  // Stream with error handling
  Stream<List<String>> fetchDataStream(String url) async* {
    try {
      for (int page = 1; page <= 3; page++) {
        await Future.delayed(Duration(seconds: 1));

        if (page == 2) {
          throw Exception('Network error on page 2');
        }

        yield ['Item $page-1', 'Item $page-2', 'Item $page-3'];
      }
    } catch (e) {
      yield* Stream.error(e);
    }
  }

  // Transform stream
  Stream<String> transformStream(Stream<int> stream) {
    return stream.map((number) => 'Count: $number');
  }

  // Filter stream
  Stream<int> filterStream(Stream<int> stream) {
    return stream.where((number) => number.isEven);
  }
}

void main() {
  final service = StreamService();

  // Listen to countdown
  service.countdownStream(5).listen(
    (count) => print('Countdown: $count'),
    onError: (error) => print('Error: $error'),
    onDone: () => print('Countdown complete'),
  );
}
```

### Example 7: Provider State Management (Popular Pattern)

```dart
// State management with Provider package

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

// Simple state model
class UserModel extends ChangeNotifier {
  String _username = '';
  List<String> _favoriteItems = [];
  bool _isLoading = false;

  String get username => _username;
  List<String> get favoriteItems => List.unmodifiable(_favoriteItems);
  bool get isLoading => _isLoading;

  Future<void> loadUser(String id) async {
    _isLoading = true;
    notifyListeners();

    try {
      // Simulate API call
      await Future.delayed(Duration(seconds: 2));

      _username = 'User #$id';
      _favoriteItems = ['Item 1', 'Item 2', 'Item 3'];
    } finally {
      _isLoading = false;
      notifyListeners();
    }
  }

  void toggleFavorite(String item) {
    if (_favoriteItems.contains(item)) {
      _favoriteItems.remove(item);
    } else {
      _favoriteItems.add(item);
    }
    notifyListeners();
  }

  void logout() {
    _username = '';
    _favoriteItems.clear();
    notifyListeners();
  }
}

// UI using Provider
class UserProfile extends StatelessWidget {
  const UserProfile({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('User Profile')),
      body: Consumer<UserModel>(
        builder: (context, userModel, child) {
          if (userModel.isLoading) {
            return const Center(child: CircularProgressIndicator());
          }

          return Column(
            children: [
              Padding(
                padding: const EdgeInsets.all(16.0),
                child: Text(
                  'Username: ${userModel.username}',
                  style: Theme.of(context).textTheme.headlineSmall,
                ),
              ),
              Expanded(
                child: ListView.builder(
                  itemCount: userModel.favoriteItems.length,
                  itemBuilder: (context, index) {
                    final item = userModel.favoriteItems[index];
                    return ListTile(
                      title: Text(item),
                      trailing: IconButton(
                        icon: const Icon(Icons.favorite),
                        color: Colors.red,
                        onPressed: () => userModel.toggleFavorite(item),
                      ),
                    );
                  },
                ),
              ),
              Padding(
                padding: const EdgeInsets.all(16.0),
                child: ElevatedButton(
                  onPressed: userModel.logout,
                  child: const Text('Logout'),
                ),
              ),
            ],
          );
        },
      ),
    );
  }
}

// App setup with Provider
class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return ChangeNotifierProvider<UserModel>(
      create: (_) => UserModel()..loadUser('1'),
      child: MaterialApp(
        title: 'Provider Example',
        home: const UserProfile(),
      ),
    );
  }
}

void main() {
  runApp(const MyApp());
}
```

### Example 8: HTTP Requests and JSON Parsing

```dart
// REST API calls with error handling and JSON parsing

import 'dart:convert';
import 'package:http/http.dart' as http;

// Data models
class Post {
  final int id;
  final String title;
  final String body;

  Post({
    required this.id,
    required this.title,
    required this.body,
  });

  // JSON parsing
  factory Post.fromJson(Map<String, dynamic> json) {
    return Post(
      id: json['id'] as int,
      title: json['title'] as String,
      body: json['body'] as String,
    );
  }

  Map<String, dynamic> toJson() => {
    'id': id,
    'title': title,
    'body': body,
  };
}

// API service
class ApiService {
  static const String baseUrl = 'https://jsonplaceholder.typicode.com';

  // GET request
  Future<Post> getPost(int id) async {
    try {
      final response = await http.get(
        Uri.parse('$baseUrl/posts/$id'),
      ).timeout(const Duration(seconds: 10));

      if (response.statusCode == 200) {
        final json = jsonDecode(response.body) as Map<String, dynamic>;
        return Post.fromJson(json);
      } else {
        throw Exception('Failed to load post: ${response.statusCode}');
      }
    } on http.ClientException catch (e) {
      throw Exception('Network error: $e');
    }
  }

  // GET with list response
  Future<List<Post>> getPosts({int limit = 10}) async {
    final response = await http.get(
      Uri.parse('$baseUrl/posts?_limit=$limit'),
    );

    if (response.statusCode == 200) {
      final list = jsonDecode(response.body) as List;
      return list.map((item) => Post.fromJson(item)).toList();
    } else {
      throw Exception('Failed to load posts');
    }
  }

  // POST request
  Future<Post> createPost(String title, String body) async {
    final response = await http.post(
      Uri.parse('$baseUrl/posts'),
      headers: {'Content-Type': 'application/json'},
      body: jsonEncode({
        'title': title,
        'body': body,
        'userId': 1,
      }),
    );

    if (response.statusCode == 201) {
      return Post.fromJson(jsonDecode(response.body));
    } else {
      throw Exception('Failed to create post');
    }
  }
}

void main() async {
  final service = ApiService();

  try {
    // Get single post
    final post = await service.getPost(1);
    print('Post: ${post.title}');

    // Get multiple posts
    final posts = await service.getPosts(limit: 5);
    print('Posts count: ${posts.length}');

    // Create new post
    final newPost = await service.createPost(
      'My Post',
      'This is my post content',
    );
    print('Created post: ${newPost.title}');
  } catch (e) {
    print('Error: $e');
  }
}
```

---

## Part 4: Advanced Patterns (200 lines)

### Example 9: BLoC Pattern for Complex State

```dart
// Business Logic Component pattern for enterprise apps

import 'package:flutter/material.dart';

// Events
abstract class CounterEvent {}

class IncrementEvent extends CounterEvent {}

class DecrementEvent extends CounterEvent {}

class ResetEvent extends CounterEvent {}

// State
class CounterState {
  final int count;
  final bool isLoading;
  final String? error;

  CounterState({
    this.count = 0,
    this.isLoading = false,
    this.error,
  });

  CounterState copyWith({
    int? count,
    bool? isLoading,
    String? error,
  }) {
    return CounterState(
      count: count ?? this.count,
      isLoading: isLoading ?? this.isLoading,
      error: error,
    );
  }
}

// BLoC
class CounterBloc {
  int _count = 0;
  final _stateController = ValueNotifier<CounterState>(CounterState());

  ValueNotifier<CounterState> get state => _stateController;

  void handle(CounterEvent event) {
    if (event is IncrementEvent) {
      _count++;
      _updateState();
    } else if (event is DecrementEvent) {
      if (_count > 0) {
        _count--;
      }
      _updateState();
    } else if (event is ResetEvent) {
      _count = 0;
      _updateState();
    }
  }

  void _updateState() {
    _stateController.value = CounterState(count: _count);
  }

  void dispose() {
    _stateController.dispose();
  }
}

// UI with BLoC
class CounterBlocWidget extends StatefulWidget {
  const CounterBlocWidget({super.key});

  @override
  State<CounterBlocWidget> createState() => _CounterBlocWidgetState();
}

class _CounterBlocWidgetState extends State<CounterBlocWidget> {
  late CounterBloc _bloc;

  @override
  void initState() {
    super.initState();
    _bloc = CounterBloc();
  }

  @override
  void dispose() {
    _bloc.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: const Text('BLoC Counter')),
      body: Center(
        child: ValueListenableBuilder<CounterState>(
          valueListenable: _bloc.state,
          builder: (context, state, child) {
            return Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                const Text('Current Count:'),
                Text(
                  '${state.count}',
                  style: Theme.of(context).textTheme.displayLarge,
                ),
                const SizedBox(height: 48),
                Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    FloatingActionButton(
                      onPressed: () => _bloc.handle(DecrementEvent()),
                      child: const Icon(Icons.remove),
                    ),
                    const SizedBox(width: 16),
                    FloatingActionButton(
                      onPressed: () => _bloc.handle(ResetEvent()),
                      child: const Icon(Icons.refresh),
                    ),
                    const SizedBox(width: 16),
                    FloatingActionButton(
                      onPressed: () => _bloc.handle(IncrementEvent()),
                      child: const Icon(Icons.add),
                    ),
                  ],
                ),
              ],
            );
          },
        ),
      ),
    );
  }
}

void main() {
  runApp(
    MaterialApp(
      title: 'BLoC Example',
      home: const CounterBlocWidget(),
    ),
  );
}
```

### Example 10: Performance Optimization

```dart
// Performance optimization techniques

class PerformanceOptimizationExample {
  // Use const constructors for immutable widgets
  static final constButtonWidget = const Text('Tap me');

  // Cache expensive computations
  Map<String, int> _computeCache = {};

  int expensiveComputation(String key, int Function() compute) {
    if (_computeCache.containsKey(key)) {
      return _computeCache[key]!;
    }

    final result = compute();
    _computeCache[key] = result;
    return result;
  }

  // Use ListView.builder for large lists (not ListView)
  Widget buildOptimizedList(List<String> items) {
    return ListView.builder(
      itemCount: items.length,
      itemBuilder: (context, index) {
        return ListTile(
          title: Text(items[index]),
        );
      },
    );
  }

  // Lazy initialization with lazy getters
  late final expensiveResource = _initializeResource();

  String _initializeResource() {
    print('Initializing expensive resource...');
    return 'Resource ready';
  }

  // Isolate for heavy computation
  static Future<int> heavyComputation(int value) async {
    // This runs in a separate isolate
    int result = 0;
    for (int i = 0; i < 1000000000; i++) {
      result += i;
    }
    return result;
  }
}
```

---

## Part 5: Context7 Integration (Additional)

### Using Context7 for Real-Time Documentation

```dart
// When implementing Dart/Flutter features:
// 1. Detect keywords (async, Future, Stream, Widget, State)
// 2. Automatically load Context7 documentation
// 3. Reference official APIs and examples
// 4. Stay up-to-date with latest Dart/Flutter versions

// Example: Getting latest documentation
// Skill("moai-context7-lang-integration") + Skill("moai-lang-dart")
// → mcp__context7__get-library-docs("/dart/docs")
// → Real-time API reference with examples
```

---

## Part 6: Migration Guides (100 lines)

### Dart 2.x → 3.x Migration

```dart
// OLD: Dart 2.x style
void oldStyle() {
  String name = 'Alice';
  int? age; // Still need ? for null
}

// NEW: Dart 3.x with null safety by default
void newStyle() {
  // Non-nullable by default
  String name = 'Alice';

  // Explicitly nullable
  String? optionalValue;

  // Pattern matching (Dart 3.0+)
  final value = age;
  switch (value) {
    case int n when n > 18:
      print('Adult');
    case int n:
      print('Minor');
    case null:
      print('Age not set');
  }
}

// Future → async/await
// OLD style
Future<String> oldFuture() {
  return Future.delayed(
    Duration(seconds: 1),
    () => 'Result',
  );
}

// NEW style
Future<String> newAsync() async {
  await Future.delayed(Duration(seconds: 1));
  return 'Result';
}
```

### Migration Checklist

- [x] Enable null safety (`pubspec.yaml`: `sdk: ">=3.0.0"`)
- [x] Use non-nullable types by default
- [x] Replace `.length > 0` with `.isNotEmpty`
- [x] Use `const` constructors where possible
- [x] Replace callbacks with async/await
- [x] Migrate to Flutter 3.24 (or latest)
- [x] Update package dependencies
- [x] Run `dart fix --apply`

---

## Quick Links

- **Official Dart Documentation**: https://dart.dev
- **Flutter Documentation**: https://flutter.dev
- **Pub Package Repository**: https://pub.dev
- **Dart Language Tour**: https://dart.dev/guides/language/language-tour
- **Flutter Widget Catalog**: https://flutter.dev/docs/development/ui/widgets
- **Effective Dart**: https://dart.dev/guides/language/effective-dart

---

_For more detailed reference, see `reference.md` and `SKILL.md`_
