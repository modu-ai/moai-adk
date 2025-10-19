---
name: moai-lang-dart
description: Dart best practices with flutter test, dart analyze, and Flutter widget
  patterns
allowed-tools:
- Read
- Bash
---

# Dart Expert

## What it does

Provides Dart-specific expertise for TDD development, including flutter test framework, dart analyze linting, and Flutter widget patterns for cross-platform app development.

## When to use

- "Dart 테스트 작성", "Flutter 위젯 패턴", "flutter test 사용법"
- Automatically invoked when working with Dart/Flutter projects
- Dart SPEC implementation (`/alfred:2-build`)

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

## Examples

### Example 1: TDD with flutter test
User: "/alfred:2-build UI-001"
Claude: (creates RED widget test, GREEN implementation, REFACTOR with const)

### Example 2: Static analysis
User: "dart analyze 실행"
Claude: (runs dart analyze and reports issues)

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Dart-specific review)
- mobile-app-expert (Flutter app development)
