---
name: moai-lang-dart
description: Dart best practices with flutter test, dart analyze, and Flutter widget patterns
allowed-tools:
  - Read
  - Bash
tier: 3
auto-load: "true"
---

# Dart Expert

## What it does

Provides Dart-specific expertise for TDD development, including flutter test framework, dart analyze linting, and Flutter widget patterns for cross-platform app development.

## When to use

- "Dart 테스트 작성", "Flutter 위젯 패턴", "flutter test 사용법", "모바일 앱", "크로스 플랫폼", "위젯 테스트"
- "Material Design", "Cupertino", "상태 관리", "네비게이션"
- "Provider", "GetX", "Riverpod", "BLoC", "MobX"
- Automatically invoked when working with Dart/Flutter projects
- Dart SPEC implementation (`/alfred:2-run`)

## How it works

**TDD Framework**:
- **flutter test**: Built-in test framework
- **mockito**: Mocking library for Dart
- **Widget testing**: Test Flutter widgets
- Test coverage with `flutter test --coverage`

**Code Quality**:
- **dart analyze**: Static analysis tool
- **dart format**: Code formatting
- **very_good_analysis**: Strict lint rules

**Package Management**:
- **pub**: Package manager (pub.dev)
- **pubspec.yaml**: Dependency configuration
- Flutter SDK version management

**Flutter Patterns**:
- **StatelessWidget/StatefulWidget**: UI components
- **Provider/Riverpod**: State management
- **BLoC**: Business logic separation
- **Navigator**: Routing and navigation

**Best Practices**:
- File ≤300 LOC, function ≤50 LOC
- Prefer `const` constructors for immutable widgets
- Use `final` for immutable fields
- Widget composition over inheritance

## Modern Dart (3.0+)

**Recommended Version**: Dart 3.2+ (stable), 3.0+ for null safety features

**Modern Features**:
- **Null safety** (2.12+): Non-nullable by default with ?
- **Sealed classes** (3.0+): Exhaustive pattern matching
- **Macros** (3.5 preview): Compile-time code generation
- **Extensions** (2.7+): Add methods to existing types
- **Enhanced enums** (3.0+): Enums with fields and methods
- **Class modifiers** (3.0+): final, base, interface, sealed

**Version Check**:
```bash
dart --version
flutter --version
```

## Package Management Commands

### Using pub (Built-in Package Manager)
```bash
# Initialize new project
flutter create my_app
dart create -t console-full my_console_app

# Add dependencies
flutter pub add provider
flutter pub add --dev mockito test

# Get dependencies
flutter pub get
dart pub get

# Update dependencies
flutter pub upgrade
dart pub upgrade --major-versions

# Remove dependencies
flutter pub remove provider

# Check dependencies
dart pub deps
dart pub outdated

# Run pub commands
dart pub publish --dry-run  # Prepare for publishing

# Build Flutter app
flutter build apk
flutter build ios
flutter build web

# Run tests
flutter test
dart test

# Format and analyze
dart format .
dart analyze

# Generate code
dart run build_runner build
```

### pubspec.yaml Configuration
```yaml
name: my_app
description: My Flutter application
version: 1.0.0+1

environment:
  sdk: '>=3.0.0 <4.0.0'
  flutter: '>=3.13.0'

dependencies:
  flutter:
    sdk: flutter
  provider: ^6.0.0
  go_router: ^10.0.0

dev_dependencies:
  flutter_test:
    sdk: flutter
  mockito: ^5.0.0
  test: ^1.21.0

flutter:
  uses-material-design: true
  assets:
    - assets/
  fonts:
    - family: CustomFont
      fonts:
        - asset: assets/fonts/CustomFont-Regular.ttf
        - asset: assets/fonts/CustomFont-Bold.ttf
          weight: 700
```

## Examples

### Example 1: TDD with flutter test
User: "/alfred:2-run UI-001"
Claude: (creates RED widget test, GREEN implementation, REFACTOR with const)

### Example 2: Static analysis
User: "dart analyze 실행"
Claude: (runs dart analyze and reports issues)

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Dart-specific review)
- mobile-app-expert (Flutter app development)
