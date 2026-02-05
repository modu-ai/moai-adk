---
paths:
  - "**/*.dart"
  - "**/pubspec.yaml"
  - "**/pubspec.lock"
---

# Flutter/Dart Rules

Version: Flutter 3.24+ / Dart 3.5+

## Tooling

- Build: flutter CLI
- Linting: dart analyze, flutter_lints
- Formatting: dart format
- Testing: flutter test
- Package management: pub

## MUST

- Use Riverpod or Provider for state management
- Use go_router for navigation
- Use freezed for immutable models
- Use const constructors when possible
- Handle null safety properly
- Separate business logic from UI

## MUST NOT

- Use setState in complex widgets
- Use BuildContext across async gaps
- Ignore analyzer warnings
- Use dynamic type
- Block the UI thread
- Hardcode strings (use l10n)

## File Conventions

- *_test.dart for test files
- Use snake_case for file names
- Use PascalCase for classes
- Use camelCase for functions and variables
- lib/ for source, test/ for tests

## Testing

- Use flutter_test for widget tests
- Use mockito for mocking
- Use golden tests for UI verification
- Use integration_test for E2E

## MoAI Integration

- Use Skill("moai-lang-flutter") for detailed patterns
- Follow TRUST 5 quality gates
