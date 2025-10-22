# moai-lang-dart - CLI Reference

_Last updated: 2025-10-22_

## Quick Reference

### Installation

```bash
# Install Dart SDK (standalone)
# macOS
brew tap dart-lang/dart
brew install dart

# Linux
sudo apt-get update
sudo apt-get install apt-transport-https
wget -qO- https://dl-ssl.google.com/linux/linux_signing_key.pub | sudo gpg --dearmor -o /usr/share/keyrings/dart.gpg
echo 'deb [signed-by=/usr/share/keyrings/dart.gpg arch=amd64] https://storage.googleapis.com/download.dartlang.org/linux/debian stable main' | sudo tee /etc/apt/sources.list.d/dart_stable.list
sudo apt-get update
sudo apt-get install dart

# Install Flutter (includes Dart)
# macOS
brew install --cask flutter

# Linux
sudo snap install flutter --classic

# Verify installation
dart --version    # Dart SDK version 3.6.0 (stable)
flutter --version # Flutter 3.27.0
```

### Common Commands

```bash
# Create new project
dart create -t console my_app
flutter create my_app

# Run application
dart run
flutter run

# Test
dart test
flutter test

# Test with coverage
flutter test --coverage
genhtml coverage/lcov.info -o coverage/html

# Analyze code
dart analyze
dart analyze --fatal-infos

# Format code
dart format .
dart format --fix .
dart format --output=none --set-exit-if-changed .

# Package management
dart pub get
dart pub upgrade
dart pub outdated
dart pub audit

# Build
flutter build apk
flutter build ios
flutter build web

# Code generation (mockito, json_serializable)
dart run build_runner build
dart run build_runner watch
```

## Tool Versions (2025-10-22)

- **Dart**: 3.6.0 (stable)
- **Flutter**: 3.27.0 (stable)
- **dart analyze**: 3.6.0 (built-in linter)
- **dart format**: 3.6.0 (built-in formatter)
- **test**: 1.25.8 (unit testing framework)
- **flutter_test**: included with Flutter SDK
- **mockito**: 5.4.4 (mocking framework)
- **build_runner**: 2.4.13 (code generation)
- **flutter_lints**: 5.0.0 (linting rules)

## Official Documentation Links

- **Dart Language**: https://dart.dev/
- **Flutter**: https://flutter.dev/
- **Dart Testing**: https://dart.dev/tools/testing
- **Flutter Testing**: https://docs.flutter.dev/testing
- **Package Repository**: https://pub.dev/
- **Dart Packages**: https://pub.dev/packages

## Testing Framework

### dart test (Pure Dart Projects)

```dart
import 'package:test/test.dart';

void main() {
  test('description', () {
    expect(actual, matcher);
  });

  group('group description', () {
    setUp(() {
      // Runs before each test in this group
    });

    tearDown(() {
      // Runs after each test in this group
    });

    test('test 1', () {});
    test('test 2', () {});
  });
}
```

### flutter_test (Flutter Projects)

```dart
import 'package:flutter_test/flutter_test.dart';

void main() {
  testWidgets('description', (WidgetTester tester) async {
    await tester.pumpWidget(MyWidget());
    expect(find.text('Hello'), findsOneWidget);
  });
}
```

### Common Matchers

```dart
// Equality
expect(actual, equals(expected));
expect(actual, isNot(expected));

// Type checking
expect(value, isA<Type>());
expect(value, isNull);
expect(value, isNotNull);

// Numeric
expect(value, greaterThan(10));
expect(value, lessThan(20));
expect(value, inClosedRange(1, 10));
expect(value, closeTo(15.0, 0.1));

// Strings
expect(str, contains('substring'));
expect(str, startsWith('prefix'));
expect(str, endsWith('suffix'));
expect(str, matches(RegExp(r'\d+')));

// Collections
expect(list, isEmpty);
expect(list, isNotEmpty);
expect(list, hasLength(5));
expect(list, contains(item));
expect(map, containsKey('key'));
expect(map, containsValue(value));

// Async
await expectLater(future, completes);
await expectLater(future, throwsA(isA<Exception>()));

// Streams
expect(stream, emits(value));
expect(stream, emitsInOrder([1, 2, 3]));
expect(stream, emitsDone);
expect(stream, emitsError(isA<Exception>()));
```

## Widget Testing (Flutter)

### WidgetTester Methods

```dart
// Build widget
await tester.pumpWidget(widget);

// Trigger frame
await tester.pump();
await tester.pump(Duration(seconds: 1));

// Wait for animations
await tester.pumpAndSettle();

// Find widgets
find.text('Button');
find.byType(ElevatedButton);
find.byKey(Key('my_key'));
find.byIcon(Icons.add);
find.widgetWithText(ElevatedButton, 'Click me');

// Interact
await tester.tap(finder);
await tester.longPress(finder);
await tester.drag(finder, Offset(100, 0));
await tester.enterText(finder, 'text');

// Verify
expect(find.text('Result'), findsOneWidget);
expect(find.byType(CircularProgressIndicator), findsNothing);
```

## Linting & Formatting

### analysis_options.yaml

```yaml
include: package:flutter_lints/flutter.yaml

analyzer:
  strong-mode:
    implicit-casts: false
    implicit-dynamic: false
  errors:
    missing_required_param: error
    missing_return: error
    todo: ignore
  exclude:
    - "build/**"
    - "**/*.g.dart"
    - "**/*.freezed.dart"

linter:
  rules:
    - always_declare_return_types
    - always_use_package_imports
    - avoid_print
    - avoid_relative_lib_imports
    - prefer_const_constructors
    - prefer_final_fields
    - prefer_final_locals
    - sort_constructors_first
    - use_key_in_widget_constructors
```

### dart format Options

```bash
# Format all files
dart format .

# Check formatting without modifying
dart format --output=none .

# Fail if changes needed (CI)
dart format --output=none --set-exit-if-changed .

# Fix formatting issues
dart format --fix .

# Specify line length
dart format --line-length=120 .
```

## Testing CLI Commands

### Running Tests

```bash
# Run all tests
dart test
flutter test

# Run specific file
dart test test/widget_test.dart
flutter test test/widget_test.dart

# Run tests matching pattern
dart test --name="Calculator"
flutter test --name="Counter"

# Run with coverage
flutter test --coverage

# Generate HTML coverage
genhtml coverage/lcov.info -o coverage/html
open coverage/html/index.html

# Run tests in watch mode
dart test --watch

# Parallel test execution
dart test --concurrency=4
flutter test --concurrency=4

# Test with verbose output
dart test --reporter=expanded
flutter test --verbose
```

### Coverage Configuration

```bash
# Install lcov (for HTML reports)
# macOS
brew install lcov

# Ubuntu/Debian
sudo apt-get install lcov

# Generate coverage
flutter test --coverage

# Convert to HTML
genhtml coverage/lcov.info -o coverage/html --no-function-coverage

# View in browser
open coverage/html/index.html
```

## Package Management

### pubspec.yaml Structure

```yaml
name: my_app
description: A new Flutter project
version: 1.0.0+1
publish_to: 'none'

environment:
  sdk: '>=3.0.0 <4.0.0'

dependencies:
  flutter:
    sdk: flutter
  http: ^1.2.2
  provider: ^6.1.2

dev_dependencies:
  flutter_test:
    sdk: flutter
  flutter_lints: ^5.0.0
  mockito: ^5.4.4
  build_runner: ^2.4.13
  test: ^1.25.8

flutter:
  uses-material-design: true
  assets:
    - assets/images/
```

### pub Commands

```bash
# Install dependencies
dart pub get
flutter pub get

# Update dependencies
dart pub upgrade
dart pub upgrade --major-versions

# Check for outdated packages
dart pub outdated

# Security audit
dart pub audit

# Add package
dart pub add package_name
flutter pub add package_name

# Remove package
dart pub remove package_name

# Publish package
dart pub publish --dry-run
dart pub publish

# Run executable from package
dart pub global activate package_name
dart pub global run package_name
```

## Code Generation

### build_runner (Mockito, JSON Serialization)

```bash
# One-time generation
dart run build_runner build

# Delete conflicting outputs and rebuild
dart run build_runner build --delete-conflicting-outputs

# Watch mode (regenerate on file changes)
dart run build_runner watch

# Clean generated files
dart run build_runner clean
```

### Mockito Mock Generation

```dart
// In test file
import 'package:mockito/annotations.dart';
import 'package:mockito/mockito.dart';

@GenerateMocks([MyRepository, MyService])
import 'my_test.mocks.dart';

void main() {
  test('example', () {
    final mockRepo = MockMyRepository();
    when(mockRepo.fetch()).thenReturn('data');

    expect(mockRepo.fetch(), equals('data'));
    verify(mockRepo.fetch()).called(1);
  });
}
```

## Quality Gate Commands (TRUST 5)

### Test Coverage (≥85%)

```bash
# Generate coverage
flutter test --coverage

# View HTML report
genhtml coverage/lcov.info -o coverage/html
open coverage/html/index.html

# Check coverage percentage
lcov --summary coverage/lcov.info
```

### Readable Code

```bash
# Run analyzer
dart analyze --fatal-infos

# Format code
dart format --fix .

# Check format (CI)
dart format --output=none --set-exit-if-changed .
```

### Unified Types

```dart
// Dart has strong type system by default
// Use null safety features

String? nullableString;
String nonNullableString = 'value';

late String lateString;
final String finalString = 'immutable';
const String constString = 'compile-time constant';
```

### Security

```bash
# Check for security vulnerabilities
dart pub audit

# Check for outdated dependencies
dart pub outdated

# Update dependencies
dart pub upgrade
```

### Trackable (@TAG)

```bash
# Search for TAG markers
rg '@(CODE|TEST|SPEC):' -n lib/ test/ --type dart

# Count TAG usage
rg '@CODE:' lib/ --type dart | wc -l
rg '@TEST:' test/ --type dart | wc -l
```

## Build & Deployment

### Flutter Build Commands

```bash
# Android
flutter build apk --release
flutter build appbundle --release

# iOS
flutter build ios --release
flutter build ipa --release

# Web
flutter build web --release

# Desktop
flutter build macos --release
flutter build windows --release
flutter build linux --release
```

### Flutter Run Modes

```bash
# Debug mode (default)
flutter run

# Profile mode (performance profiling)
flutter run --profile

# Release mode
flutter run --release

# Specific device
flutter run -d chrome
flutter run -d emulator-5554
```

## Debugging

### Debug Commands

```bash
# Run with DevTools
flutter run --observatory-port=8888
dart devtools

# Print debug info
flutter run --verbose

# Dump widget hierarchy
flutter screenshot --type=renderTree
```

### dart:developer Utilities

```dart
import 'dart:developer';

void debugFunction() {
  // Breakpoint
  debugger();

  // Logging
  log('Debug message', name: 'my.app');

  // Timing
  Timeline.startSync('operation');
  // ... operation ...
  Timeline.finishSync();
}
```

---

_For working examples, see [examples.md](examples.md)_
