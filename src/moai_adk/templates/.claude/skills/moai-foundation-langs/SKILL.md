---
name: moai-foundation-langs
tier: 1
description: Auto-detects project language and framework (package.json, pyproject.toml, etc) and provides LanguageInterface standard
allowed-tools:
- Read
- Bash
- Write
- Edit
- TodoWrite
---

# Alfred Language Detection & LanguageInterface

## What it does

Automatically detects project's primary language and framework by scanning configuration files, then recommends appropriate testing tools and linters.

**NEW in v0.4.0**: LanguageInterface standard for all moai-lang-* skills.

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

---

## LanguageInterface Definition (v0.4.0+)

**Standard fields that ALL moai-lang-* skills MUST implement**:

```yaml
interface:
  language: "Python"           # Language name
  test_framework: "pytest"      # Testing tool
  linter: "ruff"               # Linter tool
  formatter: "black"           # Formatter tool
  type_checker: "mypy"         # Type checker (optional)
  package_manager: "uv"        # Package manager
  version_requirement: ">=3.11" # Minimum version
```

**Purpose**: Ensures all moai-lang-* skills provide consistent toolchain recommendations.

**Usage by Sub-agents**:
- `language-detector` uses this interface to return standardized JSON
- `document-generator` references this for tech.md STACK section
- `feature-selector` uses this for language-specific skill selection

---

## Language Detection Patterns

**Python**:
- Config files: `pyproject.toml`, `requirements.txt`, `setup.py`
- Frameworks: FastAPI, Django, Flask
- LanguageInterface: `pytest`, `ruff`, `black`, `mypy`, `uv`

**TypeScript**:
- Config files: `package.json` + `tsconfig.json`
- Frameworks: Next.js, React, Vue
- LanguageInterface: `vitest`, `biome`, `biome`, `tsc`, `npm`

**Java**:
- Config files: `pom.xml`, `build.gradle`
- Frameworks: Spring Boot, Quarkus
- LanguageInterface: `JUnit`, `Checkstyle`, `google-java-format`, N/A, `Maven`

**Go**:
- Config files: `go.mod`
- Frameworks: Gin, Echo
- LanguageInterface: `go test`, `golint`, `gofmt`, N/A, `go mod`

**Rust**:
- Config files: `Cargo.toml`
- Frameworks: Actix, Rocket
- LanguageInterface: `cargo test`, `clippy`, `rustfmt`, N/A, `cargo`

---

## Works well with

- moai-lang-python (implements LanguageInterface for Python)
- moai-lang-typescript (implements LanguageInterface for TypeScript)
- moai-lang-java (implements LanguageInterface for Java)
- All other moai-lang-* skills (depend on this interface)

---

## Examples

### Example 1: Auto-detect project language
User: "/alfred:0-project"
Claude: (scans config files, detects Python, recommends pytest + ruff + black)

### Example 2: Manual detection
User: "이 프로젝트는 무슨 언어?"
Claude: (analyzes config files and reports language + framework)