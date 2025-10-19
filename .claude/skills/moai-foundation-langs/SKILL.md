---
name: moai-foundation-langs
description: Auto-detects project language and framework (package.json, pyproject.toml, etc)
allowed-tools:
  - Read
  - Bash
  - Write
  - Edit
  - TodoWrite
tier: 0
auto-load: "true"
---

# Alfred Language Detection

## What it does

Automatically detects project's primary language and framework by scanning configuration files, then recommends appropriate testing tools and linters.

## When to use

- "언어 감지", "프로젝트 언어 확인", "테스트 도구 추천", "프레임워크 감지"
- "빌드 도구", "패키지 관리자", "린터 추천", "포맷터 추천"
- "Language detection", "Framework identification", "Tool recommendation"
- Automatically invoked by `/alfred:0-project`, `/alfred:2-run`
- Setting up new project
- Migrating to MoAI-ADK from existing codebase

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

**Supported Languages**: Python, TypeScript, JavaScript, Java, Go, Rust, Ruby, Dart, Swift, Kotlin, PHP, C#, C++, Elixir, Scala, Clojure, Haskell, Lua, Julia, R, Shell, SQL (23 languages)

## Detection Commands

### File-based Language Detection
```bash
# TypeScript/JavaScript
ls package.json tsconfig.json 2>/dev/null

# Python
ls pyproject.toml setup.py requirements.txt 2>/dev/null

# Rust
ls Cargo.toml 2>/dev/null

# Go
ls go.mod 2>/dev/null

# Java/Kotlin
ls build.gradle pom.xml 2>/dev/null

# Ruby
ls Gemfile 2>/dev/null

# Dart/Flutter
ls pubspec.yaml 2>/dev/null

# Swift
ls Package.swift 2>/dev/null

# C/C++
ls CMakeLists.txt Makefile 2>/dev/null
```

### Framework Detection
```bash
# Check for specific frameworks
grep "\"react\":" package.json
grep "fastapi" pyproject.toml
grep "spring-boot" pom.xml
grep "gin-gonic/gin" go.mod
```

### Tool Version Check
```bash
# Python
python --version
pytest --version
ruff --version
mypy --version

# TypeScript
node --version
tsc --version
vitest --version

# Rust
rustc --version
cargo --version

# Go
go version
```

## Examples

### Example 1: Auto-detect Python project
User: "/alfred:0-project"

Alfred detects:
```json
{
  "language": "Python",
  "version": "3.12",
  "framework": "FastAPI",
  "test_framework": "pytest",
  "linter": "ruff",
  "formatter": "black",
  "type_checker": "mypy",
  "package_manager": "uv"
}
```

Result: Recommends `uv add --dev pytest ruff black mypy`

### Example 2: TypeScript with React
User: "이 프로젝트는 무슨 언어?"

Alfred detects:
```json
{
  "language": "TypeScript",
  "framework": "React + Next.js",
  "test_framework": "Vitest",
  "linter": "Biome",
  "package_manager": "pnpm"
}
```

Result: Reports "TypeScript (React + Next.js) project with Vitest + Biome"

### Example 3: Multi-language project
User: "프로젝트 언어 분석"

Alfred detects:
```json
{
  "primary_language": "TypeScript",
  "secondary_languages": ["Python", "Shell"],
  "frontend": "React",
  "backend": "FastAPI (Python)",
  "scripts": "Shell (DevOps)"
}
```

Result: Recommends language-specific tools for each layer

### Example 4: Monorepo detection
User: "언어 감지"

Alfred detects:
```
Monorepo structure:
- packages/frontend/ → TypeScript (React)
- packages/backend/ → Python (FastAPI)
- packages/mobile/ → Dart (Flutter)
```

Result: Recommends tools for each package

## Language-Framework Mapping

| Language | Common Frameworks | Test Framework | Linter | Package Manager |
|----------|------------------|----------------|--------|-----------------|
| **Python** | FastAPI, Django, Flask | pytest | ruff | uv, poetry, pip |
| **TypeScript** | React, Next.js, Vue | Vitest, Jest | Biome, ESLint | pnpm, npm, yarn |
| **Java** | Spring Boot, Quarkus | JUnit | Checkstyle | Maven, Gradle |
| **Go** | Gin, Echo, Fiber | go test | golint | go mod |
| **Rust** | Axum, Actix | cargo test | clippy | cargo |
| **Ruby** | Rails, Sinatra | RSpec | RuboCop | Bundler |
| **Dart** | Flutter | flutter test | dart analyze | pub |
| **Swift** | SwiftUI, Vapor | XCTest | SwiftLint | SPM |
| **Kotlin** | Spring Boot, Ktor | JUnit | detekt | Gradle |

## Works well with

- moai-lang-* (Language-specific skills)
- moai-foundation-trust (Language-agnostic quality checks)