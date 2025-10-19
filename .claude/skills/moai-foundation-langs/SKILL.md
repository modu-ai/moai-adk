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

- "언어 감지", "프로젝트 언어 확인", "테스트 도구 추천", "프레임워크 감지", "도구 체인 추천"
- "Language detection", "Framework detection", "Toolchain recommendation", "Auto-detect language"
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

## Detection Priority (Multiple Config Files)

When multiple config files exist:
1. `package.json` (TypeScript/JavaScript) - Highest priority
2. `pyproject.toml` (Python) - Second priority
3. Language-specific files (Cargo.toml, go.mod, etc.)
4. Source code analysis - Last resort

## Examples

### Example 1: Python FastAPI Project
User: "/alfred:0-project"
Claude:
```
🔍 Scanning configuration files...

Detected: pyproject.toml ✅
Language: Python 3.11+
Framework: FastAPI

Recommended tools:
├── Test: pytest
├── Linter: ruff
├── Formatter: black
├── Type: mypy
└── Package: uv

Setup complete! Ready for TDD workflow.
```

### Example 2: TypeScript Next.js Project
User: "프로젝트 언어 감지"
Claude:
```
🔍 Scanning configuration files...

Detected: package.json ✅
Language: TypeScript 5.x
Framework: Next.js 14

Recommended tools:
├── Test: Vitest
├── Linter: Biome
├── Formatter: Biome
├── Type: tsc (strict mode)
└── Package: pnpm

Current status:
✅ package.json found
✅ tsconfig.json found (strict: true)
⚠️ vitest not installed
  → Run: pnpm add -D vitest

Ready for development!
```

### Example 3: Go HTTP Server
User: "이 프로젝트는 무슨 언어?"
Claude:
```
Detected: go.mod ✅
Language: Go 1.21
Framework: Gin (detected from imports)

Recommended tools:
├── Test: go test
├── Linter: golint, staticcheck
├── Formatter: gofmt
└── Package: Go modules

Installation:
✅ Go 1.21 (installed)
✅ Module support (built-in)

Next: /alfred:2-run SPEC-ID
```

### Example 4: Multi-language Project (Python + TypeScript)
User: "이 프로젝트 분석해줘"
Claude:
```
🔍 Multi-language project detected:

Primary: Python (72% LOC)
  ├── Framework: Django
  ├── Config: pyproject.toml
  ├── Test: pytest
  └── Tools: ruff, black, mypy

Secondary: TypeScript (28% LOC)
  ├── Framework: React
  ├── Config: package.json
  ├── Test: Vitest
  └── Tools: Biome, tsc

Recommended approach:
1. Use Python as primary language (/alfred:2-build)
2. Configure TypeScript for frontend
3. Run both test suites in CI/CD

Toolchain:
$ pip install pytest ruff black mypy  # Python
$ pnpm add -D vitest @biomejs/biome  # TypeScript
```

### Example 5: Rust Cargo Project
User: "/alfred:2-run SPEC-001"
Claude:
```
Detected: Cargo.toml ✅
Language: Rust (Edition 2021)
Framework: None (standard library preferred)

Recommended tools:
├── Test: cargo test
├── Linter: clippy
├── Formatter: rustfmt
└── Security: cargo audit

Toolchain ready:
✅ Rust 1.73+
✅ Cargo (installed)
✅ Clippy (installed)
✅ Rustfmt (installed)

Starting TDD workflow...
```

### Example 6: Unsupported/Ambiguous Language
User: "/alfred:0-project"
Claude:
```
⚠️ No recognized config file found

Fallback: Analyzing source files...
- .R files: 45%
- .py files: 35%
- .sh files: 20%

Primary: R (detected)
Recommended tools:
├── Test: testthat
├── Linter: lintr
└── Tools: styler

Note: R is supported but requires manual TDD setup.
Reference: moai-lang-r skill
```

## Supported Languages & Config Files

| Language | Config File | Min Version | Tools |
|----------|-------------|-------------|-------|
| Python | `pyproject.toml` | 3.10+ | pytest, ruff, black, mypy |
| TypeScript | `package.json` | 5.x | Vitest, Biome, tsc |
| JavaScript | `package.json` | ES2020+ | Jest, ESLint, Prettier |
| Go | `go.mod` | 1.21+ | go test, golint, gofmt |
| Rust | `Cargo.toml` | 2021 | cargo test, clippy, rustfmt |
| Java | `build.gradle` | 17+ | JUnit, Maven, Checkstyle |
| Kotlin | `build.gradle.kts` | 1.9+ | JUnit, Gradle, ktlint |
| C# | `*.csproj` | .NET 8+ | xUnit, Roslyn, msbuild |
| Ruby | `Gemfile` | 3.0+ | RSpec, RuboCop, Bundler |
| Dart | `pubspec.yaml` | 3.1+ | flutter test, dart analyze |
| Swift | `Package.swift` | 5.9+ | XCTest, SwiftLint |
| PHP | `composer.json` | 8.0+ | PHPUnit, PHP-CS-Fixer |

## Keywords

"언어 감지", "프로젝트 언어 확인", "테스트 도구 추천", "language detection", "framework detection", "toolchain recommendation"

## Reference

- Language tier skills: moai-lang-{language}
- Framework guides: CLAUDE.md#다중-언어-지원
- TRUST tools: moai-foundation-trust

## Works well with

- moai-lang-* (language-specific implementation)
- moai-foundation-trust (TRUST validation)
- `/alfred:0-project` (project initialization)