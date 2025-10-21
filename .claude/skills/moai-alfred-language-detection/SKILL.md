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
- The text shown to the user is written using TUI/report expressions.
- When running the tool, a summary of commands and results are recorded.

## Examples
```markdown
- Call this skill inside the /alfred command to generate a report.
- Add summary to Completion Report.
```

## Inputs
- MoAI-ADK project context (`.moai/project/`, `.claude/` templates, etc.).
- Parameters passed from user commands or higher commands.

## Outputs
- Reports, checklists or recommendations for your Alfred workflow.
- Structured data for subsequent subagent calls.

## Failure Modes
- When required input documents are missing or permissions are limited.
- When disruptive changes are required without user approval.

## Dependencies
- Cooperation with higher-level agents such as cc-manager and project-manager is required.

## References
- GitHub Linguist. "Programmatic language detection." https://github.com/github-linguist/linguist (accessed 2025-03-29).
- JetBrains. "Project language insights." https://www.jetbrains.com/help/idea/project-tool-window.html (accessed 2025-03-29).

## Changelog
- 2025-03-29: Added input/output/failure response to Alfred-specific skills.

## Works well with

- alfred-trust-validation (language-specific tool verification)
- alfred-code-reviewer (language-specific review criteria)
