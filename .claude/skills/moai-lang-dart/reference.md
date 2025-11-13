# moai-lang-dart - CLI Reference & API Guide

_Last updated: 2025-11-12_

Complete reference guide for Dart 3.5 and Flutter 3.24 development including version compatibility matrix, package management, CLI tools, and core API reference.

---

## Version Compatibility Matrix (2025-11-12)

### Dart & Flutter Versions

| Component | Version | Status | LTS | EOL Date |
|-----------|---------|--------|-----|----------|
| **Dart** | 3.6.0 | Current | 3.4 LTS | 2025-12 |
| **Flutter** | 3.24.0 | Current | 3.22 LTS | 2025-12 |
| **Dart** | 3.5.4 | Previous | - | 2025-10 |
| **Flutter** | 3.22.3 | LTS | ✅ | 2026-04 |

### Recommended Setup

```
Platform          | Dart    | Flutter | Xcode  | Android SDK
------------------|---------|---------|--------|------------
macOS (Intel)     | 3.6.0   | 3.24.0  | 15.4+  | API 34+
macOS (Apple M1)  | 3.6.0   | 3.24.0  | 15.4+  | API 34+
Linux             | 3.6.0   | 3.24.0  | -      | API 34+
Windows           | 3.6.0   | 3.24.0  | -      | API 34+
iOS (min)         | 3.6.0   | 3.24.0  | 15.4+  | -
Android (min)     | 3.6.0   | 3.24.0  | -      | API 24+
```

---

## Installation & Setup

### macOS Installation

```bash
# Using Homebrew (recommended)
brew install dart

# Using FVM (Flutter Version Manager - for multiple versions)
brew install fvm
fvm install 3.24.0
fvm global 3.24.0

# Verify installation
dart --version
flutter --version

# Run doctor to check environment
dart pub global activate fvm
fvm flutter doctor

# Get all dependencies
flutter pub get
```

### Environment Setup

```bash
# Add to ~/.zshrc or ~/.bash_profile
export PATH="$PATH:$HOME/fvm/default/bin"
export PATH="$PATH:$HOME/fvm/default/bin/cache/dart-sdk/bin"

# Verify PATH
echo $PATH | tr ':' '\n' | grep -i dart
```

---

## Common Commands

### Project Management

```bash
# Create new project
dart create my_console_app          # Console project
flutter create my_app                # Flutter app
flutter create --template=plugin my_plugin  # Plugin

# Get dependencies
flutter pub get

# Upgrade dependencies
flutter pub upgrade
flutter pub upgrade --major-versions

# Check for outdated packages
flutter pub outdated
```

### Development & Testing

```bash
# Run app (Flutter)
flutter run
flutter run -d chrome              # Web
flutter run -d macos              # macOS
flutter run --no-sound-null-safety # Unsafe (for testing)

# Run tests
flutter test
flutter test --coverage
flutter test test/specific_test.dart

# Run specific test with verbose output
flutter test -v test/unit/user_model_test.dart

# Run tests with coverage
flutter test --coverage
lcov --list coverage/lcov.info
```

### Code Quality

```bash
# Static analysis
dart analyze
flutter analyze

# Format code
dart format lib/

# Fix common issues
dart fix --apply

# Lint
flutter pub get
dart run custom_lint

# Linting with stricter rules
dart analyze --fatal-infos
```

### Build & Release

```bash
# Build APK (Android)
flutter build apk --release
flutter build apk --split-per-abi

# Build iOS
flutter build ios --release

# Build web
flutter build web --release

# Build macOS
flutter build macos --release

# Build Windows
flutter build windows --release

# Check build size
flutter build apk --analyze-size --release
```

### Debugging

```bash
# Enable verbose logging
flutter run -v

# Attach to running app
flutter attach

# Debug mode
flutter run --debug

# Profile mode
flutter run --profile

# Release mode
flutter run --release

# Enable verbose hot reload
flutter run --verbose --track-widget-creation
```

---

## Package Versions (Key Dependencies)

### State Management

| Package | Version | Purpose | Status |
|---------|---------|---------|--------|
| **provider** | 6.4.0+ | Recommended (easy) | ✅ Active |
| **riverpod** | 2.6.0+ | Advanced (powerful) | ✅ Active |
| **bloc** | 8.1.0+ | Enterprise pattern | ✅ Active |
| **getx** | 4.6.0+ | All-in-one | ✅ Active |
| **mobx** | 2.3.0+ | Reactive | ✅ Active |

### HTTP & Networking

| Package | Version | Purpose | Status |
|---------|---------|---------|--------|
| **http** | 1.2.0+ | Simple HTTP | ✅ Active |
| **dio** | 5.6.0+ | Advanced HTTP | ✅ Active |
| **retrofit** | 4.1.0+ | REST client | ✅ Active |
| **socket_io_client** | 2.0.0+ | WebSocket | ✅ Active |

### Data Storage

| Package | Version | Purpose | Status |
|---------|---------|---------|--------|
| **sqflite** | 2.3.0+ | SQLite (mobile) | ✅ Active |
| **hive** | 2.2.0+ | NoSQL (fast) | ✅ Active |
| **firebase_core** | 2.28.0+ | Firebase | ✅ Active |
| **shared_preferences** | 2.2.0+ | Key-value | ✅ Active |

### UI & Widgets

| Package | Version | Purpose | Status |
|---------|---------|---------|--------|
| **flutter_riverpod** | 2.6.0+ | UI integration | ✅ Active |
| **cached_network_image** | 3.4.0+ | Image caching | ✅ Active |
| **shimmer** | 3.0.0+ | Loading effects | ✅ Active |
| **get_it** | 7.6.0+ | Service locator | ✅ Active |

### Testing

| Package | Version | Purpose | Status |
|---------|---------|---------|--------|
| **test** | 1.25.0+ | Unit tests | ✅ Active |
| **mockito** | 5.4.0+ | Mocking | ✅ Active |
| **integration_test** | Latest | Widget tests | ✅ Active |
| **fake_async** | 1.3.0+ | Async testing | ✅ Active |

---

## Core API Reference (Dart)

### Null Safety Keywords

```dart
// Non-nullable (default)
String name = 'Alice';      // Cannot be null

// Nullable
String? email;              // Can be null

// Late initialization
late String value;          // Initialize later

// Required parameter
void greet(required String name) {}

// Default value
void greet(String name = 'Guest') {}
```

### Async Programming

```dart
// Future - represents async result
Future<String> fetchData() async => 'Data';

// Await - wait for future
final data = await fetchData();

// Then - callback style
fetchData().then((data) => print(data));

// Stream - continuous values
Stream<int> countStream() async* {
  for (int i = 0; i < 5; i++) {
    yield i;
  }
}

// Listen to stream
countStream().listen((value) => print(value));

// Timeout
await fetchData().timeout(Duration(seconds: 5));

// Error handling
try {
  await fetchData();
} catch (e) {
  print('Error: $e');
}
```

### Collections

```dart
// List
List<String> items = ['a', 'b', 'c'];
items.add('d');
items.where((item) => item.length > 1);
items.map((item) => item.toUpperCase());

// Map
Map<String, int> scores = {'Alice': 100, 'Bob': 90};
scores['Charlie'] = 95;
scores.forEach((name, score) => print('$name: $score'));

// Set
Set<String> unique = {'a', 'b', 'c'};
unique.add('d');
unique.contains('a');
```

### Extension Methods

```dart
// Define extension
extension StringExtension on String {
  String capitalize() => '${this[0].toUpperCase()}${substring(1)}';
}

// Use extension
final name = 'john'.capitalize();  // 'John'
```

### Pattern Matching (Dart 3.0+)

```dart
// Switch patterns
switch (number) {
  case 0:
    print('Zero');
  case > 0:
    print('Positive');
  case < 0:
    print('Negative');
  default:
    print('Other');
}

// Object patterns
switch (point) {
  case Point(x: 0, y: 0):
    print('Origin');
  case Point(:var x, :var y):
    print('Point at $x, $y');
}
```

---

## Core API Reference (Flutter)

### Material Design Widgets

```dart
// Layout
Scaffold              // App structure
AppBar                // Top bar
FloatingActionButton  // Action button
BottomNavigationBar   // Bottom nav

// Input
TextField             // Text input
Checkbox              // Checkbox
Radio                 // Radio button
Switch                // Toggle
DropdownButton        // Dropdown

// Display
Text                  // Text display
Image                 // Image
Icon                  // Icon
Card                  // Card container
ListTile              // List item

// Lists
ListView              // Scrollable list
ListView.builder      // Lazy list
GridView              // Grid layout
GridView.builder      // Lazy grid
```

### State Management

```dart
// StatefulWidget lifecycle
class MyWidget extends StatefulWidget {
  @override
  State<MyWidget> createState() => _MyWidgetState();
}

class _MyWidgetState extends State<MyWidget> {
  @override
  void initState() {}        // Initialize

  @override
  void didUpdateWidget(MyWidget oldWidget) {}  // Widget updated

  @override
  void dispose() {}          // Cleanup

  @override
  Widget build(BuildContext context) => Container();
}

// Setstate for rebuilds
setState(() {
  _count++;
});
```

### Navigation

```dart
// Push route
Navigator.of(context).push(
  MaterialPageRoute(builder: (_) => NextScreen()),
);

// Pop route
Navigator.of(context).pop();

// Named routes
Navigator.of(context).pushNamed('/detail', arguments: id);

// Go Router (modern)
context.push('/detail/$id');
context.pop();
```

### Theming

```dart
// Theme data
ThemeData(
  colorScheme: ColorScheme.fromSeed(seedColor: Colors.blue),
  textTheme: TextTheme(
    displayLarge: TextStyle(fontSize: 32),
    bodyMedium: TextStyle(fontSize: 14),
  ),
  useMaterial3: true,
);

// Access theme
Theme.of(context).colorScheme.primary
Theme.of(context).textTheme.headlineSmall
```

---

## pubspec.yaml Structure

```yaml
name: my_app
description: A Flutter app example.
version: 1.0.0+1

environment:
  sdk: '>=3.0.0 <4.0.0'
  flutter: '>=3.16.0'

dependencies:
  flutter:
    sdk: flutter
  provider: ^6.4.0
  dio: ^5.6.0
  shared_preferences: ^2.2.0

dev_dependencies:
  flutter_test:
    sdk: flutter
  test: ^1.25.0
  mockito: ^5.4.0

flutter:
  uses-material-design: true
  assets:
    - assets/images/
    - assets/data/
  fonts:
    - family: Roboto
      fonts:
        - asset: assets/fonts/Roboto-Regular.ttf
        - asset: assets/fonts/Roboto-Bold.ttf
          weight: 700
```

---

## Best Practices Checklist

### Code Quality

- [ ] Enable strict null safety (`sdk: '>=3.0.0'`)
- [ ] Run `dart analyze` without warnings
- [ ] Achieve ≥85% test coverage
- [ ] Use `const` constructors
- [ ] Follow Effective Dart guidelines
- [ ] Use linting with `analysis_options.yaml`

### Performance

- [ ] Use `ListView.builder` for large lists
- [ ] Implement image caching
- [ ] Profile with DevTools
- [ ] Minimize rebuilds with `const`
- [ ] Use `RepaintBoundary` for complex widgets
- [ ] Profile memory usage

### Security

- [ ] Validate input
- [ ] Use HTTPS for APIs
- [ ] Secure sensitive data with Keychain
- [ ] Keep dependencies updated
- [ ] Use code obfuscation for release
- [ ] Follow OWASP guidelines

### Testing

- [ ] Write unit tests for business logic
- [ ] Write widget tests for UI
- [ ] Write integration tests for flows
- [ ] Mock external dependencies
- [ ] Test error cases
- [ ] Aim for ≥85% coverage

---

## Troubleshooting

### Common Issues

**Issue**: `No devices available`
```bash
# Solution
flutter devices
flutter config --android-sdk /path/to/sdk
```

**Issue**: `Gradle build failed`
```bash
# Solution
flutter clean
rm -rf pubspec.lock
flutter pub get
flutter run
```

**Issue**: `Hot reload not working`
```bash
# Solution
flutter run --no-fast-start
flutter run --verbose
```

**Issue**: `Version conflict in pubspec.lock`
```bash
# Solution
rm pubspec.lock
flutter pub get
```

---

## Official Resources

- **Dart Official**: https://dart.dev
- **Flutter Official**: https://flutter.dev
- **Pub Packages**: https://pub.dev
- **Dart Docs**: https://dart.dev/guides
- **Flutter API**: https://api.flutter.dev
- **Dart API**: https://api.dart.dev
- **Effective Dart**: https://dart.dev/guides/language/effective-dart
- **Flutter DevTools**: https://flutter.dev/docs/development/tools/devtools

---

## Version History

| Date | Version | Notes |
|------|---------|-------|
| 2025-11-12 | 3.6.0 | Latest stable |
| 2025-10-22 | 3.5.4 | Previous release |
| 2025-09-01 | 3.24.0 | Flutter LTS |
| 2025-08-15 | 3.22.3 | Flutter LTS |

_For detailed examples, see `examples.md` and `SKILL.md`_
