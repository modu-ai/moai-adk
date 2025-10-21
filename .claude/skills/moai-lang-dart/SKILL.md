---

name: moai-lang-dart
description: Dart best practices with flutter test, dart analyze, and Flutter widget patterns. Use when writing or reviewing Dart/Flutter code in project workflows.
allowed-tools:
  - Read
  - Bash
---

# Dart Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand when language keywords are detected |
| Trigger cues | Dart code discussions, framework guidance, or file extensions such as .dart. |
| Tier | 3 |

## What it does

Provides Dart-specific expertise for TDD development, including flutter test framework, dart analyze linting, and Flutter widget patterns for cross-platform app development.

## When to use

- Engages when the conversation references Dart work, frameworks, or files like .dart.
- “Writing Dart tests”, “Flutter widget patterns”, “How to use flutter tests”
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

## Examples
```bash
dart test && dart analyze
```

## Inputs
- Language-specific source directories (e.g. `src/`, `app/`).
- Language-specific build/test configuration files (e.g. `package.json`, `pyproject.toml`, `go.mod`).
- Relevant test suites and sample data.

## Outputs
- Test/lint execution plan tailored to the selected language.
- List of key language idioms and review checkpoints.

## Failure Modes
- When the language runtime or package manager is not installed.
- When the main language cannot be determined in a multilingual project.

## Dependencies
- Access to the project file is required using the Read/Grep tool.
- When used with `Skill("moai-foundation-langs")`, it is easy to share cross-language conventions.

## References
- Google. "Dart Language Tour." https://dart.dev/guides/language/language-tour (accessed 2025-03-29).
- Flutter. "Testing." https://docs.flutter.dev/testing (accessed 2025-03-29).

## Changelog
- 2025-03-29: Input/output/failure response/reference information for each language has been specified.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Dart-specific review)
- mobile-app-expert (Flutter app development)

## Best Practices
- Enable automatic validation by matching your linter with the language's official style guide.
- Fix test/build pipelines with reproducible commands in CI.
