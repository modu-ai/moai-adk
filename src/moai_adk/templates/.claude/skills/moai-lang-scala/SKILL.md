---

name: moai-lang-scala
description: Scala best practices with ScalaTest, sbt, and functional programming patterns. Use when writing or reviewing Scala code in project workflows.
allowed-tools:
  - Read
  - Bash
---

# Scala Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand when language keywords are detected |
| Trigger cues | Scala code discussions, framework guidance, or file extensions such as .scala. |
| Tier | 3 |

## What it does

Provides Scala-specific expertise for TDD development, including ScalaTest framework, sbt build tool, and functional programming patterns.

## When to use

- Engages when the conversation references Scala work, frameworks, or files like .scala.
- “Writing Scala tests”, “How to use ScalaTest”, “Functional programming”
- Automatically invoked when working with Scala projects
- Scala SPEC implementation (`/alfred:2-run`)

## How it works

**TDD Framework**:
- **ScalaTest**: Flexible testing framework
- **specs2**: BDD-style testing
- **ScalaCheck**: Property-based testing
- Test coverage with sbt-scoverage

**Build Tools**:
- **sbt**: Scala build tool
- **build.sbt**: Build configuration
- Multi-project builds

**Code Quality**:
- **Scalafmt**: Code formatting
- **Scalafix**: Linting and refactoring
- **WartRemover**: Code linting

**Functional Programming**:
- **Immutable data structures**
- **Higher-order functions**
- **Pattern matching**
- **For-comprehensions**
- **Monads (Option, Either, Try)**

**Best Practices**:
- File ≤300 LOC, method ≤50 LOC
- Prefer immutable vals over mutable vars
- Case classes for data modeling
- Tail recursion for loops
- Avoid null, use Option

## Examples
```bash
sbt test && sbt scalafmtCheck
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
- Lightbend. "Scala Documentation." https://docs.scala-lang.org/ (accessed 2025-03-29).
- Scalameta. "scalafmt." https://scalameta.org/scalafmt/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: 언어별 입력·출력·실패 대응·참조 정보를 명세했습니다.

## Works well with

- alfred-trust-validation (coverage verification)
- alfred-code-reviewer (Scala-specific review)
- alfred-refactoring-coach (functional refactoring)

## Best Practices
- 언어 공식 스타일 가이드와 린터를 일치시켜 자동 검증을 활성화하세요.
- CI에서 재현 가능한 명령으로 테스트/빌드 파이프라인을 고정합니다.
