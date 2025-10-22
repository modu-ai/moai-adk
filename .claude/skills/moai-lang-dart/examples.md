# moai-lang-dart - Working Examples

_Last updated: 2025-10-22_

## Example 1: Project Setup with Flutter Test & dart test

```bash
# Create new Flutter project
flutter create my_app
cd my_app

# For pure Dart project (no Flutter)
dart create -t console my_dart_app
cd my_dart_app
```

**pubspec.yaml configuration**:
```yaml
name: my_app
description: A new Flutter project
version: 1.0.0+1

environment:
  sdk: '>=3.0.0 <4.0.0'

dependencies:
  flutter:
    sdk: flutter

dev_dependencies:
  flutter_test:
    sdk: flutter
  flutter_lints: ^5.0.0
  mockito: ^5.4.4
  build_runner: ^2.4.13
  test: ^1.25.8

flutter:
  uses-material-design: true
```

## Example 2: TDD Workflow with dart test

**RED: Write failing test**
```dart
// test/calculator_test.dart
import 'package:test/test.dart';
import 'package:my_app/calculator.dart';

void main() {
  // @TEST:CALC-001
  group('Calculator', () {
    late Calculator calculator;

    setUp(() {
      calculator = Calculator();
    });

    test('should add two positive numbers', () {
      expect(calculator.add(2, 3), equals(5));
    });

    test('should add negative numbers', () {
      expect(calculator.add(-2, -3), equals(-5));
    });

    test('should subtract numbers', () {
      expect(calculator.subtract(5, 3), equals(2));
    });

    test('should divide numbers', () {
      expect(calculator.divide(10, 2), equals(5.0));
    });

    test('should throw on division by zero', () {
      expect(
        () => calculator.divide(10, 0),
        throwsA(isA<ArgumentError>()),
      );
    });
  });
}
```

**GREEN: Implement feature**
```dart
// lib/calculator.dart

/**
 * @CODE:CALC-001 | SPEC: SPEC-CALC-001.md | TEST: calculator_test.dart
 * Basic calculator with arithmetic operations
 */
class Calculator {
  int add(int a, int b) => a + b;

  int subtract(int a, int b) => a - b;

  double divide(double a, double b) {
    if (b == 0) {
      throw ArgumentError('Cannot divide by zero');
    }
    return a / b;
  }

  int multiply(int a, int b) => a * b;
}
```

**REFACTOR: Add type safety and extensions**
```dart
// lib/calculator.dart

/**
 * @CODE:CALC-001 | SPEC: SPEC-CALC-001.md | TEST: calculator_test.dart
 * Calculator with enhanced type safety
 */
class Calculator {
  T add<T extends num>(T a, T b) {
    if (a is int && b is int) {
      return (a + b) as T;
    }
    return (a.toDouble() + b.toDouble()) as T;
  }

  T subtract<T extends num>(T a, T b) {
    if (a is int && b is int) {
      return (a - b) as T;
    }
    return (a.toDouble() - b.toDouble()) as T;
  }

  double divide(num a, num b) {
    if (b == 0) {
      throw ArgumentError('Cannot divide by zero');
    }
    return a / b;
  }
}

// Extension methods
extension NumExtensions on num {
  bool isEven() => this % 2 == 0;
  bool isOdd() => !isEven();
  num squared() => this * this;
}
```

## Example 3: Widget Testing with flutter_test

```dart
// test/widget/counter_widget_test.dart
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:my_app/widgets/counter_widget.dart';

void main() {
  // @TEST:UI-001
  testWidgets('Counter increments smoke test', (WidgetTester tester) async {
    // Build the widget
    await tester.pumpWidget(
      const MaterialApp(
        home: CounterWidget(),
      ),
    );

    // Verify initial state
    expect(find.text('0'), findsOneWidget);
    expect(find.text('1'), findsNothing);

    // Tap the '+' icon and trigger a frame
    await tester.tap(find.byIcon(Icons.add));
    await tester.pump();

    // Verify that counter has incremented
    expect(find.text('0'), findsNothing);
    expect(find.text('1'), findsOneWidget);
  });

  testWidgets('Counter decrements', (WidgetTester tester) async {
    await tester.pumpWidget(
      const MaterialApp(
        home: CounterWidget(initialValue: 5),
      ),
    );

    expect(find.text('5'), findsOneWidget);

    await tester.tap(find.byIcon(Icons.remove));
    await tester.pump();

    expect(find.text('4'), findsOneWidget);
  });

  testWidgets('Counter resets', (WidgetTester tester) async {
    await tester.pumpWidget(
      const MaterialApp(
        home: CounterWidget(initialValue: 10),
      ),
    );

    await tester.tap(find.byKey(const Key('reset_button')));
    await tester.pump();

    expect(find.text('0'), findsOneWidget);
  });
}
```

**Widget implementation**:
```dart
// lib/widgets/counter_widget.dart
import 'package:flutter/material.dart';

/**
 * @CODE:UI-001 | SPEC: SPEC-UI-001.md | TEST: counter_widget_test.dart
 * Counter widget with increment/decrement
 */
class CounterWidget extends StatefulWidget {
  final int initialValue;

  const CounterWidget({
    super.key,
    this.initialValue = 0,
  });

  @override
  State<CounterWidget> createState() => _CounterWidgetState();
}

class _CounterWidgetState extends State<CounterWidget> {
  late int _counter;

  @override
  void initState() {
    super.initState();
    _counter = widget.initialValue;
  }

  void _increment() {
    setState(() {
      _counter++;
    });
  }

  void _decrement() {
    setState(() {
      _counter--;
    });
  }

  void _reset() {
    setState(() {
      _counter = 0;
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Counter'),
      ),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Text(
              '$_counter',
              style: Theme.of(context).textTheme.headlineMedium,
            ),
            const SizedBox(height: 20),
            Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                IconButton(
                  icon: const Icon(Icons.remove),
                  onPressed: _decrement,
                ),
                IconButton(
                  key: const Key('reset_button'),
                  icon: const Icon(Icons.refresh),
                  onPressed: _reset,
                ),
                IconButton(
                  icon: const Icon(Icons.add),
                  onPressed: _increment,
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

## Example 4: Async Testing with Futures and Streams

```dart
// test/api_client_test.dart
import 'package:test/test.dart';
import 'package:my_app/api_client.dart';

void main() {
  // @TEST:API-001
  group('ApiClient', () {
    late ApiClient client;

    setUp(() {
      client = ApiClient();
    });

    test('should fetch user data', () async {
      final user = await client.fetchUser(1);

      expect(user.id, equals(1));
      expect(user.name, isNotEmpty);
    });

    test('should handle errors', () async {
      expect(
        () => client.fetchUser(-1),
        throwsA(isA<ApiException>()),
      );
    });

    test('should stream data updates', () async {
      final stream = client.watchUser(1);

      expect(
        stream,
        emitsInOrder([
          isA<User>().having((u) => u.id, 'id', equals(1)),
          isA<User>().having((u) => u.name, 'name', contains('Updated')),
          emitsDone,
        ]),
      );
    });

    test('should fetch multiple users concurrently', () async {
      final ids = [1, 2, 3, 4, 5];
      final stopwatch = Stopwatch()..start();

      final users = await Future.wait(
        ids.map((id) => client.fetchUser(id)),
      );

      stopwatch.stop();

      expect(users.length, equals(5));
      // Should complete in parallel, not sequentially
      expect(stopwatch.elapsedMilliseconds, lessThan(2000));
    });
  });
}
```

**Implementation**:
```dart
// lib/api_client.dart
import 'dart:async';

/**
 * @CODE:API-001 | SPEC: SPEC-API-001.md | TEST: api_client_test.dart
 * API client with async operations
 */
class ApiClient {
  Future<User> fetchUser(int id) async {
    if (id < 0) {
      throw ApiException('Invalid user ID: $id');
    }

    // Simulate network delay
    await Future.delayed(const Duration(milliseconds: 500));

    return User(id: id, name: 'User $id', email: 'user$id@example.com');
  }

  Stream<User> watchUser(int id) async* {
    yield await fetchUser(id);

    await Future.delayed(const Duration(seconds: 1));

    yield User(id: id, name: 'User $id Updated', email: 'user$id@example.com');
  }
}

class User {
  final int id;
  final String name;
  final String email;

  User({required this.id, required this.name, required this.email});
}

class ApiException implements Exception {
  final String message;
  ApiException(this.message);

  @override
  String toString() => 'ApiException: $message';
}
```

## Example 5: Mockito for Dependency Injection

```dart
// test/user_service_test.dart
import 'package:mockito/mockito.dart';
import 'package:mockito/annotations.dart';
import 'package:test/test.dart';
import 'package:my_app/user_service.dart';
import 'package:my_app/user_repository.dart';

// Generate mocks: dart run build_runner build
@GenerateMocks([UserRepository])
import 'user_service_test.mocks.dart';

void main() {
  // @TEST:USER-001
  group('UserService', () {
    late MockUserRepository mockRepo;
    late UserService service;

    setUp(() {
      mockRepo = MockUserRepository();
      service = UserService(mockRepo);
    });

    test('should create user with valid email', () async {
      final email = 'test@example.com';
      final user = User(id: '123', email: email);

      when(mockRepo.save(any)).thenAnswer((_) async => user);
      when(mockRepo.findByEmail(email)).thenAnswer((_) async => null);

      final result = await service.createUser(email);

      expect(result.email, equals(email));
      verify(mockRepo.save(any)).called(1);
    });

    test('should reject invalid email', () async {
      expect(
        () => service.createUser('invalid-email'),
        throwsA(isA<ArgumentError>()),
      );

      verifyNever(mockRepo.save(any));
    });

    test('should handle duplicate email', () async {
      final email = 'existing@example.com';
      final existingUser = User(id: '999', email: email);

      when(mockRepo.findByEmail(email)).thenAnswer((_) async => existingUser);

      expect(
        () => service.createUser(email),
        throwsA(isA<DuplicateUserException>()),
      );

      verifyNever(mockRepo.save(any));
    });
  });
}
```

**Service implementation**:
```dart
// lib/user_service.dart
import 'user_repository.dart';

/**
 * @CODE:USER-001 | SPEC: SPEC-USER-001.md | TEST: user_service_test.dart
 * User management with validation
 */
class UserService {
  final UserRepository _repository;

  UserService(this._repository);

  Future<User> createUser(String email) async {
    _validateEmail(email);

    final existing = await _repository.findByEmail(email);
    if (existing != null) {
      throw DuplicateUserException('User with email $email already exists');
    }

    final user = User(
      id: DateTime.now().millisecondsSinceEpoch.toString(),
      email: email,
    );

    return await _repository.save(user);
  }

  void _validateEmail(String email) {
    final emailRegex = RegExp(r'^[\w-\.]+@([\w-]+\.)+[\w-]{2,4}$');
    if (!emailRegex.hasMatch(email)) {
      throw ArgumentError('Invalid email format: $email');
    }
  }
}

class User {
  final String id;
  final String email;

  User({required this.id, required this.email});
}

class DuplicateUserException implements Exception {
  final String message;
  DuplicateUserException(this.message);
}
```

## Example 6: Quality Gate Check

```bash
# Run all tests
flutter test

# For pure Dart projects
dart test

# Run with coverage
flutter test --coverage

# Generate HTML coverage report
genhtml coverage/lcov.info -o coverage/html
open coverage/html/index.html

# Run analyzer (linting)
dart analyze

# Fix formatting
dart format .

# Check formatting
dart format --output=none --set-exit-if-changed .

# TRUST 5 validation
echo "T - Test coverage"
flutter test --coverage
# Verify coverage ≥85% in coverage/html/index.html

echo "R - Readable code"
dart analyze
dart format --output=none --set-exit-if-changed .

echo "U - Unified types"
# Dart has strong type system by default

echo "S - Security"
dart pub outdated
dart pub audit

echo "T - Trackable with @TAG"
rg '@(CODE|TEST|SPEC):' -n lib/ test/ --type dart
```

## Example 7: Parameterized Tests

```dart
// test/math_test.dart
import 'package:test/test.dart';
import 'package:my_app/math_utils.dart';

void main() {
  group('MathUtils parameterized tests', () {
    final testCases = [
      {'a': 2, 'b': 3, 'expected': 5},
      {'a': -1, 'b': 1, 'expected': 0},
      {'a': 0, 'b': 0, 'expected': 0},
      {'a': 100, 'b': 200, 'expected': 300},
    ];

    for (final testCase in testCases) {
      test('should add ${testCase['a']} + ${testCase['b']} = ${testCase['expected']}', () {
        final result = MathUtils.add(
          testCase['a'] as int,
          testCase['b'] as int,
        );
        expect(result, equals(testCase['expected']));
      });
    }
  });

  group('Division edge cases', () {
    test('should handle large numbers', () {
      expect(MathUtils.divide(1e15, 1e5), closeTo(1e10, 0.1));
    });

    test('should handle very small numbers', () {
      expect(MathUtils.divide(1e-10, 1e-5), closeTo(1e-5, 1e-10));
    });
  });
}
```

---

## TRUST 5 Integration

### Test Coverage (≥85%)
```bash
flutter test --coverage
genhtml coverage/lcov.info -o coverage/html
```

### Readable Code
```bash
dart analyze
dart format .
```

### Unified Types
- Use Dart's strong type system
- Leverage null safety
- Use `late`, `final`, `const` appropriately

### Security
```bash
dart pub audit
dart pub outdated
```

### Trackable with @TAG
```bash
rg '@(CODE|TEST|SPEC):' -n lib/ test/ --type dart
```

---

_For detailed CLI reference, see [reference.md](reference.md)_
