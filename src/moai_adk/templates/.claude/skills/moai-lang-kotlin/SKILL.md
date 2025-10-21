---

name: moai-lang-kotlin
description: Kotlin best practices with JUnit, Gradle, ktlint, coroutines, and extension functions. Use when writing or reviewing Kotlin code in project workflows.
allowed-tools:
  - Read
  - Bash
---

# Kotlin Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand when language keywords are detected |
| Trigger cues | Kotlin code discussions, framework guidance, or file extensions such as .kt/.kts. |
| Tier | 3 |

## What it does

Provides Kotlin-specific expertise for TDD development, including JUnit testing, Gradle build system, ktlint linting, coroutines for concurrency, and extension functions.

## When to use

- Engages when the conversation references Kotlin work, frameworks, or files like .kt/.kts.
- “Writing Kotlin tests”, “How to use coroutines”, “Android patterns”
- Automatically invoked when working with Kotlin/Android projects
- Kotlin SPEC implementation (`/alfred:2-run`)

## How it works

**TDD Framework**:
- **JUnit 5**: Unit testing with Kotlin extensions
- **MockK**: Kotlin-friendly mocking library
- **Kotest**: Kotlin-native testing framework
- Test coverage ≥85% with JaCoCo

**Build Tools**:
- **Gradle**: build.gradle.kts with Kotlin DSL
- **Maven**: pom.xml alternative
- Multi-platform support (JVM, Native, JS)

**Code Quality**:
- **ktlint**: Kotlin linter with formatting
- **detekt**: Static code analysis
- **Android Lint**: Android-specific checks

**Kotlin Features**:
- **Coroutines**: Async programming with suspend functions
- **Extension functions**: Add methods to existing classes
- **Data classes**: Automatic equals/hashCode/toString
- **Null safety**: Non-nullable types by default
- **Smart casts**: Automatic type casting after checks

**Android Patterns**:
- **Jetpack Compose**: Declarative UI
- **ViewModel**: UI state management
- **Room**: Database abstraction
- **Retrofit**: Network requests

## Examples
```bash
./gradlew test && ./gradlew ktlintCheck
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
- JetBrains. "Kotlin Language Documentation." https://kotlinlang.org/docs/home.html (accessed 2025-03-29).
- Pinterest. "ktlint." https://pinterest.github.io/ktlint/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: 언어별 입력·출력·실패 대응·참조 정보를 명세했습니다.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Kotlin-specific review)
- mobile-app-expert (Android app development)

## Best Practices
- 언어 공식 스타일 가이드와 린터를 일치시켜 자동 검증을 활성화하세요.
- CI에서 재현 가능한 명령으로 테스트/빌드 파이프라인을 고정합니다.
