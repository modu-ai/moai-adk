---
name: moai-alfred-language-detection
description: Detects project primary language and framework based on config files, recommends appropriate testing tools and linters
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - TodoWrite
---

# Alfred Language Detection

## What it does

Automatically detects project's primary language and framework by scanning configuration files, then recommends appropriate testing tools and linters.

## When to use

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

## Examples

### Example 1: Auto-detect project language
User: "/alfred:0-project"
Claude: (scans config files, detects Python, recommends pytest + ruff + black)

### Example 2: Manual detection
User: "What language is this project in?"
Claude: (analyzes config files and reports language + framework)

## Works well with

- alfred-trust-validation (language-specific tool verification)
- alfred-code-reviewer (language-specific review criteria)
