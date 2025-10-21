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
- 언어별 소스 디렉터리(e.g. `src/`, `app/`).
- 언어별 빌드/테스트 설정 파일(예: `package.json`, `pyproject.toml`, `go.mod`).
- 관련 테스트 스위트 및 샘플 데이터.

## Outputs
- 선택된 언어에 맞춘 테스트/린트 실행 계획.
- 주요 언어 관용구와 리뷰 체크포인트 목록.

## Failure Modes
- 언어 런타임이나 패키지 매니저가 설치되지 않았을 때.
- 다중 언어 프로젝트에서 주 언어를 판별하지 못했을 때.

## Dependencies
- Read/Grep 도구로 프로젝트 파일 접근이 필요합니다.
- `Skill("moai-foundation-langs")`와 함께 사용하면 교차 언어 규약 공유가 용이합니다.

## References
- Google. "Dart Language Tour." https://dart.dev/guides/language/language-tour (accessed 2025-03-29).
- Flutter. "Testing." https://docs.flutter.dev/testing (accessed 2025-03-29).

## Changelog
- 2025-03-29: 언어별 입력·출력·실패 대응·참조 정보를 명세했습니다.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Dart-specific review)
- mobile-app-expert (Flutter app development)

## Best Practices
- 언어 공식 스타일 가이드와 린터를 일치시켜 자동 검증을 활성화하세요.
- CI에서 재현 가능한 명령으로 테스트/빌드 파이프라인을 고정합니다.
