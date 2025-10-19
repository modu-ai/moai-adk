---
name: moai-foundation-langs
description: Auto-detects project language and framework (package.json, pyproject.toml,
  etc)
allowed-tools:
- Read
- Bash
- Write
- Edit
- TodoWrite
---

# Alfred Language Detection

## What it does

Automatically detects project's primary language and framework by scanning configuration files, then recommends appropriate testing tools and linters.

## When to use

- "언어 감지", "프로젝트 언어 확인", "테스트 도구 추천"
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

## Examples

### Example 1: Auto-detect project language
User: "/alfred:0-project"
Claude: (scans config files, detects Python, recommends pytest + ruff + black)

### Example 2: Manual detection
User: "이 프로젝트는 무슨 언어?"
Claude: (analyzes config files and reports language + framework)