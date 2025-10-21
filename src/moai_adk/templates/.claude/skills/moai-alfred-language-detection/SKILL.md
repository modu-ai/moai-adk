---

name: moai-alfred-language-detection
description: Detects the project’s primary language/runtime and recommends default testing tooling when initializing or planning workflows. Use when sizing up the project language and default tooling.
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - TodoWrite
---

# Alfred Language Detection

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Write (write_file), Edit (edit_file), Bash (terminal), TodoWrite (todo_write) |
| Auto-load | /alfred:0-project bootstrap |
| Trigger cues | Repository language probing, framework detection, tooling recommendation via Alfred. |

## What it does

Automatically detects project's primary language and framework by scanning configuration files, then recommends appropriate testing tools and linters.

## When to use

- Activates when Alfred needs to detect project languages or recommend toolchains.
- "Detect language", "Check project language", "Recommend testing tools"
- Automatically invoked by `/alfred:0-project`, `/alfred:2-run`
- Setting up new project

## How it works

**Configuration File Scanning**:
- `package.json` → TypeScript/JavaScript (Jest/Vitest, ESLint/Biome)
- `pyproject.toml` → Python (pytest, ruff, black)
- `Cargo.toml` → Rust (cargo test, clippy, rustfmt)
- `go.mod` → Go (go test, golint, gofmt)
- `Gemfile` → Ruby (RSpec, RuboCop)
- `pubspec.yaml` → Dart/Flutter (flutter test, dart analyze)
- `build.gradle` → Java/Kotlin (JUnit, Checkstyle)
- `Package.swift` → Swift (XCTest, SwiftLint)

**Toolchain Recommendation**:
```json
{
  "language": "Python",
  "test_framework": "pytest",
  "linter": "ruff",
  "formatter": "black",
  "type_checker": "mypy",
  "package_manager": "uv"
}
```

**Framework Detection**:
- **Python**: FastAPI, Django, Flask
- **TypeScript**: React, Next.js, Vue
- **Java**: Spring Boot, Quarkus

**Supported Languages**: Python, TypeScript, Java, Go, Rust, Ruby, Dart, Swift, Kotlin, PHP, C#, C++, Elixir, Scala, Clojure (20+ languages)

## Best Practices
- 사용자에게 보여주는 문구는 TUI/보고서용 표현으로 작성합니다.
- 도구 실행 시 명령과 결과 요약을 함께 기록합니다.

## Examples
```markdown
- /alfred 커맨드 내부에서 이 스킬을 호출해 보고서를 생성합니다.
- Completion Report에 요약을 추가합니다.
```

## Inputs
- MoAI-ADK 프로젝트 맥락 (`.moai/project/`, `.claude/` 템플릿 등).
- 사용자 명령 또는 상위 커맨드에서 전달한 파라미터.

## Outputs
- Alfred 워크플로우에 필요한 보고서, 체크리스트 또는 추천 항목.
- 후속 서브 에이전트 호출을 위한 구조화된 데이터.

## Failure Modes
- 필수 입력 문서가 없거나 권한이 제한된 경우.
- 사용자 승인 없이 파괴적인 변경이 요구될 때.

## Dependencies
- cc-manager, project-manager 등 상위 에이전트와 협력이 필요합니다.

## References
- GitHub Linguist. "Programmatic language detection." https://github.com/github-linguist/linguist (accessed 2025-03-29).
- JetBrains. "Project language insights." https://www.jetbrains.com/help/idea/project-tool-window.html (accessed 2025-03-29).

## Changelog
- 2025-03-29: Alfred 전용 스킬에 입력/출력/실패 대응을 추가했습니다.

## Works well with

- alfred-trust-validation (language-specific tool verification)
- alfred-code-reviewer (language-specific review criteria)
