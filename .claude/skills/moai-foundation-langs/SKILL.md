---
name: moai-foundation-langs
version: 2.0.0
created: 2025-10-22
updated: 2025-10-22
status: active
description: Auto-detects project language from package.json, pyproject.toml, go.mod, Cargo.toml, etc.
keywords: ['language', 'detection', 'framework', 'auto']
allowed-tools:
  - Read
  - Bash
---

# Foundation Langs Skill

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-foundation-langs |
| **Version** | 2.0.0 (2025-10-22) |
| **Allowed tools** | Read (read_file), Bash (terminal) |
| **Auto-load** | SessionStart, `/alfred:0-project`, `/alfred:2-run` |
| **Tier** | Foundation |

---

## What It Does

Auto-detects project language and framework from configuration files (package.json, pyproject.toml, go.mod, Cargo.toml, etc.). This Skill enables MoAI-ADK to automatically identify the project's technology stack and load appropriate language-specific Skills, testing frameworks, and TRUST 5 quality gates.

**Key capabilities**:
- ✅ Multi-language project detection (polyglot support)
- ✅ Framework and tooling identification (React, Django, Express, etc.)
- ✅ Package manager detection (npm, pnpm, pip, cargo, etc.)
- ✅ Testing framework inference (Jest, pytest, go test, etc.)
- ✅ Build tool identification (Webpack, Vite, Make, Gradle, etc.)
- ✅ Automatic Skill preloading for detected languages
- ✅ TRUST 5 quality gate configuration per language
- ✅ Supports 23+ programming languages

---

## When to Use

**Automatic triggers**:
- **SessionStart**: Analyze project structure and preload language Skills
- `/alfred:0-project`: Initialize project metadata and detect stack
- `/alfred:2-run`: Load language-specific testing and build tools
- Keywords: language, detect, framework, stack, dependencies

**Manual invocation**:
- Identify unknown project's technology stack
- Validate detected language matches expectations
- Troubleshoot missing dependencies or misconfigurations
- Generate technology stack documentation

---

## Supported Languages & Detection Patterns

### Language Detection Matrix

| Language | Primary File | Secondary Indicators | Package Manager | Testing Framework |
|----------|-------------|---------------------|-----------------|-------------------|
| **JavaScript** | package.json | .js, .mjs, .cjs files | npm, pnpm, yarn | Jest, Mocha, Vitest |
| **TypeScript** | package.json + tsconfig.json | .ts, .tsx files | npm, pnpm, yarn | Jest, Vitest, Mocha |
| **Python** | pyproject.toml, requirements.txt, setup.py | .py files, __init__.py | pip, poetry, uv | pytest, unittest |
| **Go** | go.mod | .go files | go modules | go test |
| **Rust** | Cargo.toml | .rs files | cargo | cargo test |
| **Java** | pom.xml, build.gradle | .java files | Maven, Gradle | JUnit, TestNG |
| **Kotlin** | build.gradle.kts | .kt, .kts files | Gradle | JUnit, Kotest |
| **Swift** | Package.swift, .xcodeproj | .swift files | SPM, CocoaPods | XCTest |
| **Dart** | pubspec.yaml | .dart files | pub | flutter test |
| **C** | Makefile, CMakeLists.txt | .c, .h files | Make, CMake | Unity, Check |
| **C++** | CMakeLists.txt | .cpp, .hpp, .cc files | CMake, Make | Google Test, Catch2 |
| **C#** | .csproj, .sln | .cs files | NuGet, dotnet | xUnit, NUnit, MSTest |
| **Ruby** | Gemfile | .rb files | bundler | RSpec, Minitest |
| **PHP** | composer.json | .php files | Composer | PHPUnit |
| **Scala** | build.sbt | .scala files | sbt | ScalaTest |
| **Haskell** | stack.yaml, .cabal | .hs files | Stack, Cabal | HUnit, Hspec |
| **Elixir** | mix.exs | .ex, .exs files | mix | ExUnit |
| **Clojure** | project.clj, deps.edn | .clj, .cljs files | Leiningen, deps | clojure.test |
| **Lua** | rockspec files | .lua files | LuaRocks | busted |
| **R** | DESCRIPTION | .R files | CRAN, devtools | testthat |
| **Julia** | Project.toml | .jl files | Pkg | Test stdlib |
| **SQL** | .sql files | DDL/DML scripts | N/A | Language-specific |
| **Shell** | .sh, .bash files | Shebang lines | N/A | bats, shunit2 |

---

## Core Principles

### 1. File Pattern Recognition

Language detection relies on a priority-ordered cascade:

**Priority 1**: Explicit configuration files (highest confidence)
- `package.json`, `pyproject.toml`, `go.mod`, `Cargo.toml`, etc.

**Priority 2**: Build system files (high confidence)
- `Makefile`, `CMakeLists.txt`, `build.gradle`, `pom.xml`

**Priority 3**: Source file extensions (medium confidence)
- `.py`, `.js`, `.go`, `.rs`, `.java`, etc.

**Priority 4**: Directory structure patterns (low confidence)
- `src/`, `tests/`, `lib/`, `cmd/`, etc.

**Detection Algorithm**:
```pseudocode
1. Scan root directory for explicit config files
2. If multiple configs found → polyglot project
3. Parse config files to identify frameworks
4. Scan src/ and tests/ directories for file extensions
5. Build confidence matrix:
   - Config file match: +10 points
   - Build file match: +5 points
   - Source file match: +1 point per file (max 10)
6. Language with highest score is primary
7. Languages with score >5 are secondary
8. Report detected stack with confidence levels
```

### 2. Polyglot Project Support

MoAI-ADK supports multi-language projects (e.g., Python backend + TypeScript frontend):

**Detection Strategy**:
- Identify all language markers in repository
- Classify as primary (main application) or secondary (scripts, tools, tests)
- Load appropriate Skills for each detected language
- Configure separate TRUST 5 gates per language

**Example: Full-stack TypeScript + Python project**:
```
project-root/
├── package.json          ← TypeScript (frontend)
├── pyproject.toml        ← Python (backend)
├── frontend/
│   ├── src/*.tsx         ← React components
│   └── tests/*.test.ts   ← Jest tests
└── backend/
    ├── src/*.py          ← FastAPI app
    └── tests/*.py        ← pytest tests
```

**Detection Result**:
- Primary: TypeScript (frontend), Python (backend)
- Frameworks: React (TypeScript), FastAPI (Python)
- Testing: Jest (TypeScript), pytest (Python)
- Package Managers: pnpm (TypeScript), uv (Python)

### 3. Framework Identification

After language detection, parse configuration files to identify frameworks:

**JavaScript/TypeScript Frameworks** (from package.json dependencies):
- React: `react`, `react-dom`
- Vue: `vue`
- Angular: `@angular/core`
- Next.js: `next`
- Express: `express`
- NestJS: `@nestjs/core`

**Python Frameworks** (from pyproject.toml dependencies):
- Django: `django`
- FastAPI: `fastapi`
- Flask: `flask`
- Streamlit: `streamlit`

**Java/Kotlin Frameworks** (from pom.xml or build.gradle):
- Spring Boot: `spring-boot-starter`
- Quarkus: `quarkus-core`
- Micronaut: `micronaut-core`

**Go Frameworks** (from go.mod):
- Gin: `github.com/gin-gonic/gin`
- Echo: `github.com/labstack/echo`
- Fiber: `github.com/gofiber/fiber`

### 4. Testing Framework Inference

MoAI-ADK infers testing framework from package dependencies or file patterns:

**Inference Rules**:

| Language | Test File Pattern | Package Indicators | Default Framework |
|----------|------------------|-------------------|-------------------|
| JavaScript/TypeScript | `*.test.js`, `*.spec.ts` | `jest`, `vitest`, `mocha` | Jest (most common) |
| Python | `test_*.py`, `*_test.py` | `pytest`, `unittest` | pytest (recommended) |
| Go | `*_test.go` | Built-in | go test |
| Rust | `tests/*.rs` | Built-in | cargo test |
| Java/Kotlin | `*Test.java`, `*Test.kt` | `junit-jupiter`, `kotest` | JUnit 5 |

**Fallback Strategy**:
- If no test files or packages found → recommend default framework
- If multiple frameworks detected → use most recent in package.json/pyproject.toml
- If conflicting frameworks → warn user and prompt for clarification

---

## Detection Workflow

### Phase 1: Initial Scan

**Objective**: Identify all potential language markers in repository.

**Steps**:
```bash
# 1. List all configuration files
ls -la package.json pyproject.toml go.mod Cargo.toml composer.json

# 2. Count source files by extension
find . -type f -name "*.js" | wc -l
find . -type f -name "*.py" | wc -l
find . -type f -name "*.go" | wc -l
# ... repeat for all supported languages

# 3. Identify build systems
ls -la Makefile CMakeLists.txt build.gradle pom.xml

# 4. Check for framework-specific directories
ls -d frontend/ backend/ src/ tests/ cmd/ pkg/
```

**Output**: Raw scan results with file counts.

### Phase 2: Parsing & Classification

**Objective**: Parse configuration files and classify languages by priority.

**JavaScript/TypeScript Detection**:
```bash
# Read package.json
cat package.json | jq '.dependencies, .devDependencies'

# Detect framework
if grep -q '"react"' package.json; then
  echo "Framework: React"
elif grep -q '"vue"' package.json; then
  echo "Framework: Vue"
elif grep -q '"@angular/core"' package.json; then
  echo "Framework: Angular"
fi

# Detect testing framework
if grep -q '"jest"' package.json; then
  echo "Testing: Jest"
elif grep -q '"vitest"' package.json; then
  echo "Testing: Vitest"
fi

# Detect package manager
if [ -f "pnpm-lock.yaml" ]; then
  echo "Package Manager: pnpm"
elif [ -f "yarn.lock" ]; then
  echo "Package Manager: yarn"
else
  echo "Package Manager: npm"
fi
```

**Python Detection**:
```bash
# Read pyproject.toml
cat pyproject.toml | grep -A 10 '\[tool.poetry.dependencies\]'

# Detect framework
if grep -q 'django' pyproject.toml; then
  echo "Framework: Django"
elif grep -q 'fastapi' pyproject.toml; then
  echo "Framework: FastAPI"
elif grep -q 'flask' pyproject.toml; then
  echo "Framework: Flask"
fi

# Detect testing framework
if grep -q 'pytest' pyproject.toml; then
  echo "Testing: pytest"
else
  echo "Testing: unittest (default)"
fi

# Detect package manager
if grep -q '\[tool.poetry\]' pyproject.toml; then
  echo "Package Manager: Poetry"
elif [ -f "uv.lock" ]; then
  echo "Package Manager: uv"
else
  echo "Package Manager: pip"
fi
```

**Go Detection**:
```bash
# Read go.mod
cat go.mod | grep 'require'

# Detect framework
if grep -q 'github.com/gin-gonic/gin' go.mod; then
  echo "Framework: Gin"
elif grep -q 'github.com/labstack/echo' go.mod; then
  echo "Framework: Echo"
fi

# Testing framework (always go test)
echo "Testing: go test (built-in)"

# Package manager (always go modules)
echo "Package Manager: go modules"
```

**Rust Detection**:
```bash
# Read Cargo.toml
cat Cargo.toml | grep -A 10 '\[dependencies\]'

# Detect framework
if grep -q 'actix-web' Cargo.toml; then
  echo "Framework: Actix Web"
elif grep -q 'rocket' Cargo.toml; then
  echo "Framework: Rocket"
elif grep -q 'axum' Cargo.toml; then
  echo "Framework: Axum"
fi

# Testing framework (always cargo test)
echo "Testing: cargo test (built-in)"

# Package manager (always cargo)
echo "Package Manager: cargo"
```

### Phase 3: Confidence Scoring

**Objective**: Assign confidence scores to each detected language.

**Scoring Formula**:
```
Score = (Config_Match × 10) + (Build_Match × 5) + min(Source_Files, 10) + (Test_Files × 2)

Where:
- Config_Match: 1 if explicit config file exists, 0 otherwise
- Build_Match: 1 if build system file exists, 0 otherwise
- Source_Files: Count of source files (capped at 10)
- Test_Files: Count of test files
```

**Classification**:
- Score ≥ 15: **Primary language** (high confidence)
- Score 5-14: **Secondary language** (medium confidence)
- Score < 5: **Incidental** (low confidence, scripts/tools only)

**Example Scoring**:
```
Project: Full-stack TypeScript + Python app

TypeScript:
- package.json exists: +10
- 45 .ts files: +10 (capped)
- 12 .test.ts files: +24
- Total: 44 → Primary language

Python:
- pyproject.toml exists: +10
- 8 .py files: +8
- 5 test_*.py files: +10
- Total: 28 → Primary language

Shell:
- 3 .sh files: +3
- Total: 3 → Incidental (scripts only)
```

### Phase 4: Report Generation

**Objective**: Generate structured detection report for Alfred.

**Report Format**:
```json
{
  "primary_languages": [
    {
      "language": "TypeScript",
      "confidence": 95,
      "framework": "React",
      "testing": "Jest",
      "package_manager": "pnpm",
      "source_files": 45,
      "test_files": 12,
      "config_files": ["package.json", "tsconfig.json"]
    },
    {
      "language": "Python",
      "confidence": 92,
      "framework": "FastAPI",
      "testing": "pytest",
      "package_manager": "uv",
      "source_files": 8,
      "test_files": 5,
      "config_files": ["pyproject.toml"]
    }
  ],
  "secondary_languages": [],
  "incidental_languages": [
    {
      "language": "Shell",
      "confidence": 15,
      "source_files": 3,
      "purpose": "Build scripts"
    }
  ],
  "recommended_skills": [
    "moai-lang-typescript",
    "moai-lang-python",
    "moai-domain-frontend",
    "moai-domain-backend"
  ]
}
```

---

## Language-Specific Detection Details

### JavaScript/TypeScript

**Primary Detection**:
```bash
# Check for package.json
if [ -f "package.json" ]; then
  echo "✓ JavaScript/TypeScript project detected"

  # Determine if TypeScript
  if [ -f "tsconfig.json" ] || grep -q '"typescript"' package.json; then
    echo "Language: TypeScript"
  else
    echo "Language: JavaScript"
  fi
fi
```

**Framework Detection** (from package.json):
```bash
# Parse dependencies
DEPS=$(jq -r '.dependencies // {} | keys[]' package.json)
DEV_DEPS=$(jq -r '.devDependencies // {} | keys[]' package.json)

# React
if echo "$DEPS" | grep -q "react"; then
  echo "Framework: React"
  # Check for Next.js
  if echo "$DEPS" | grep -q "next"; then
    echo "Meta-framework: Next.js"
  fi
fi

# Vue
if echo "$DEPS" | grep -q "vue"; then
  echo "Framework: Vue"
  # Check for Nuxt
  if echo "$DEPS" | grep -q "nuxt"; then
    echo "Meta-framework: Nuxt"
  fi
fi

# Angular
if echo "$DEPS" | grep -q "@angular/core"; then
  echo "Framework: Angular"
fi

# Svelte
if echo "$DEPS" | grep -q "svelte"; then
  echo "Framework: Svelte"
  # Check for SvelteKit
  if echo "$DEPS" | grep -q "@sveltejs/kit"; then
    echo "Meta-framework: SvelteKit"
  fi
fi

# Express (backend)
if echo "$DEPS" | grep -q "express"; then
  echo "Backend Framework: Express"
fi

# NestJS (backend)
if echo "$DEPS" | grep -q "@nestjs/core"; then
  echo "Backend Framework: NestJS"
fi
```

**Testing Framework Detection**:
```bash
# Check devDependencies for test frameworks
if echo "$DEV_DEPS" | grep -q "jest"; then
  echo "Testing: Jest"
elif echo "$DEV_DEPS" | grep -q "vitest"; then
  echo "Testing: Vitest"
elif echo "$DEV_DEPS" | grep -q "mocha"; then
  echo "Testing: Mocha"
elif echo "$DEV_DEPS" | grep -q "@playwright/test"; then
  echo "Testing: Playwright (E2E)"
elif echo "$DEV_DEPS" | grep -q "cypress"; then
  echo "Testing: Cypress (E2E)"
fi
```

**Package Manager Detection**:
```bash
if [ -f "pnpm-lock.yaml" ]; then
  echo "Package Manager: pnpm"
elif [ -f "yarn.lock" ]; then
  echo "Package Manager: yarn"
elif [ -f "bun.lockb" ]; then
  echo "Package Manager: bun"
else
  echo "Package Manager: npm (default)"
fi
```

### Python

**Primary Detection**:
```bash
# Check for Python-specific files (priority order)
if [ -f "pyproject.toml" ]; then
  echo "✓ Python project detected (pyproject.toml)"
elif [ -f "requirements.txt" ]; then
  echo "✓ Python project detected (requirements.txt)"
elif [ -f "setup.py" ]; then
  echo "✓ Python project detected (setup.py)"
elif [ -f "Pipfile" ]; then
  echo "✓ Python project detected (Pipfile)"
fi
```

**Framework Detection** (from pyproject.toml):
```bash
# Parse dependencies
if [ -f "pyproject.toml" ]; then
  # Django
  if grep -q 'django' pyproject.toml; then
    echo "Framework: Django"
  fi

  # FastAPI
  if grep -q 'fastapi' pyproject.toml; then
    echo "Framework: FastAPI"
  fi

  # Flask
  if grep -q 'flask' pyproject.toml; then
    echo "Framework: Flask"
  fi

  # Streamlit
  if grep -q 'streamlit' pyproject.toml; then
    echo "Framework: Streamlit"
  fi

  # Celery (task queue)
  if grep -q 'celery' pyproject.toml; then
    echo "Task Queue: Celery"
  fi
fi
```

**Testing Framework Detection**:
```bash
if grep -q 'pytest' pyproject.toml; then
  echo "Testing: pytest"
elif grep -q 'unittest' pyproject.toml; then
  echo "Testing: unittest"
else
  echo "Testing: pytest (recommended default)"
fi
```

**Package Manager Detection**:
```bash
if grep -q '\[tool.poetry\]' pyproject.toml; then
  echo "Package Manager: Poetry"
elif [ -f "uv.lock" ]; then
  echo "Package Manager: uv"
elif [ -f "Pipfile" ]; then
  echo "Package Manager: pipenv"
else
  echo "Package Manager: pip (default)"
fi
```

### Go

**Primary Detection**:
```bash
if [ -f "go.mod" ]; then
  echo "✓ Go project detected (go.mod)"

  # Parse module name
  MODULE=$(grep '^module' go.mod | awk '{print $2}')
  echo "Module: $MODULE"
fi
```

**Framework Detection**:
```bash
if [ -f "go.mod" ]; then
  # Gin (HTTP framework)
  if grep -q 'github.com/gin-gonic/gin' go.mod; then
    echo "Framework: Gin"
  fi

  # Echo (HTTP framework)
  if grep -q 'github.com/labstack/echo' go.mod; then
    echo "Framework: Echo"
  fi

  # Fiber (HTTP framework)
  if grep -q 'github.com/gofiber/fiber' go.mod; then
    echo "Framework: Fiber"
  fi

  # gRPC
  if grep -q 'google.golang.org/grpc' go.mod; then
    echo "Framework: gRPC"
  fi
fi
```

**Testing Framework** (always built-in):
```bash
echo "Testing: go test (built-in)"
```

### Rust

**Primary Detection**:
```bash
if [ -f "Cargo.toml" ]; then
  echo "✓ Rust project detected (Cargo.toml)"

  # Parse package name
  PACKAGE=$(grep '^\[package\]' -A 5 Cargo.toml | grep '^name' | awk -F'"' '{print $2}')
  echo "Package: $PACKAGE"
fi
```

**Framework Detection**:
```bash
if [ -f "Cargo.toml" ]; then
  # Actix Web
  if grep -q 'actix-web' Cargo.toml; then
    echo "Framework: Actix Web"
  fi

  # Rocket
  if grep -q 'rocket' Cargo.toml; then
    echo "Framework: Rocket"
  fi

  # Axum
  if grep -q 'axum' Cargo.toml; then
    echo "Framework: Axum"
  fi

  # Tokio (async runtime)
  if grep -q 'tokio' Cargo.toml; then
    echo "Async Runtime: Tokio"
  fi
fi
```

**Testing Framework** (always built-in):
```bash
echo "Testing: cargo test (built-in)"
```

### Java/Kotlin

**Primary Detection**:
```bash
# Maven (Java/Kotlin)
if [ -f "pom.xml" ]; then
  echo "✓ Java/Kotlin project detected (Maven)"
  if grep -q '<kotlin.version>' pom.xml; then
    echo "Language: Kotlin"
  else
    echo "Language: Java"
  fi
fi

# Gradle (Java/Kotlin)
if [ -f "build.gradle" ] || [ -f "build.gradle.kts" ]; then
  echo "✓ Java/Kotlin project detected (Gradle)"
  if [ -f "build.gradle.kts" ]; then
    echo "Language: Kotlin (Gradle KTS)"
  else
    echo "Language: Java (Gradle)"
  fi
fi
```

**Framework Detection** (Maven):
```bash
if [ -f "pom.xml" ]; then
  # Spring Boot
  if grep -q 'spring-boot-starter' pom.xml; then
    echo "Framework: Spring Boot"
  fi

  # Quarkus
  if grep -q 'quarkus-core' pom.xml; then
    echo "Framework: Quarkus"
  fi

  # Micronaut
  if grep -q 'micronaut-core' pom.xml; then
    echo "Framework: Micronaut"
  fi
fi
```

**Testing Framework Detection**:
```bash
if grep -q 'junit-jupiter' pom.xml; then
  echo "Testing: JUnit 5"
elif grep -q 'kotest' pom.xml; then
  echo "Testing: Kotest (Kotlin)"
elif grep -q 'testng' pom.xml; then
  echo "Testing: TestNG"
else
  echo "Testing: JUnit 4 (legacy)"
fi
```

### Swift

**Primary Detection**:
```bash
# Swift Package Manager
if [ -f "Package.swift" ]; then
  echo "✓ Swift project detected (SPM)"
fi

# Xcode project
if [ -d "*.xcodeproj" ]; then
  echo "✓ Swift project detected (Xcode)"
fi

# CocoaPods
if [ -f "Podfile" ]; then
  echo "Package Manager: CocoaPods"
fi
```

**Framework Detection**:
```bash
if [ -f "Package.swift" ]; then
  # Vapor (server-side)
  if grep -q 'vapor' Package.swift; then
    echo "Framework: Vapor"
  fi

  # SwiftUI (iOS/macOS)
  if grep -q 'SwiftUI' Package.swift; then
    echo "Framework: SwiftUI"
  fi
fi
```

### C/C++

**Primary Detection**:
```bash
# CMake
if [ -f "CMakeLists.txt" ]; then
  echo "✓ C/C++ project detected (CMake)"
  if grep -q 'CXX' CMakeLists.txt; then
    echo "Language: C++"
  else
    echo "Language: C"
  fi
fi

# Makefile
if [ -f "Makefile" ]; then
  echo "✓ C/C++ project detected (Makefile)"
  if grep -q '\.cpp' Makefile || grep -q '\.cc' Makefile; then
    echo "Language: C++"
  else
    echo "Language: C"
  fi
fi
```

**Testing Framework Detection**:
```bash
if [ -f "CMakeLists.txt" ]; then
  # Google Test (C++)
  if grep -q 'gtest' CMakeLists.txt; then
    echo "Testing: Google Test"
  fi

  # Catch2 (C++)
  if grep -q 'Catch2' CMakeLists.txt; then
    echo "Testing: Catch2"
  fi

  # Unity (C)
  if grep -q 'unity' CMakeLists.txt; then
    echo "Testing: Unity"
  fi
fi
```

---

## Polyglot Project Handling

### Detection Strategy

When multiple primary languages are detected (score ≥ 15 each), classify the project as polyglot:

**Classification Rules**:
1. **Monorepo**: Single repository with multiple language subdirectories
2. **Full-stack**: Frontend (TypeScript) + Backend (Python/Go/Java/etc.)
3. **Microservices**: Multiple services in different languages
4. **Language Extensions**: Core app (Python) + native extensions (C/Rust)

**Example: Full-stack TypeScript + Python**:
```
project-root/
├── package.json              ← TypeScript frontend
├── pyproject.toml            ← Python backend
├── frontend/
│   ├── src/
│   │   ├── components/*.tsx  ← React components
│   │   └── pages/*.tsx       ← Next.js pages
│   └── tests/*.test.ts       ← Jest tests
├── backend/
│   ├── src/
│   │   ├── api/*.py          ← FastAPI routes
│   │   └── models/*.py       ← Database models
│   └── tests/*.py            ← pytest tests
└── shared/
    └── types/*.ts            ← Shared TypeScript types
```

**Detection Output**:
```json
{
  "project_type": "polyglot",
  "classification": "full-stack",
  "languages": [
    {
      "name": "TypeScript",
      "role": "frontend",
      "directory": "frontend/",
      "framework": "Next.js",
      "testing": "Jest"
    },
    {
      "name": "Python",
      "role": "backend",
      "directory": "backend/",
      "framework": "FastAPI",
      "testing": "pytest"
    }
  ],
  "shared_components": ["shared/types/"]
}
```

### Polyglot Testing Strategy

For polyglot projects, configure separate test suites per language:

**Example: TypeScript + Python test execution**:
```bash
# Run all tests
npm run test:all

# Frontend tests (TypeScript + Jest)
cd frontend && npm test

# Backend tests (Python + pytest)
cd backend && pytest tests/

# Integration tests (cross-language)
npm run test:integration
```

**TRUST 5 Coverage Gates** (per language):
- TypeScript frontend: ≥85% coverage (Jest)
- Python backend: ≥85% coverage (pytest)
- Integration tests: ≥70% coverage (cross-language scenarios)

### Monorepo Detection

**Indicators**:
- Root-level `lerna.json`, `nx.json`, or `turbo.json`
- Multiple `package.json` files in subdirectories
- Workspaces configuration in root `package.json`

**Example: pnpm monorepo**:
```yaml
# pnpm-workspace.yaml
packages:
  - 'packages/*'
  - 'apps/*'
```

**Detection Output**:
```json
{
  "project_type": "monorepo",
  "monorepo_tool": "pnpm-workspace",
  "packages": [
    {
      "name": "@myorg/ui",
      "path": "packages/ui",
      "language": "TypeScript",
      "framework": "React"
    },
    {
      "name": "@myorg/backend",
      "path": "packages/backend",
      "language": "Python",
      "framework": "FastAPI"
    }
  ]
}
```

---

## Skill Preloading Strategy

After language detection, Alfred automatically preloads relevant Skills:

### Preload Rules

**Language Skills** (always preload for detected primary languages):
- TypeScript → `moai-lang-typescript`
- Python → `moai-lang-python`
- Go → `moai-lang-go`
- Rust → `moai-lang-rust`
- ... (repeat for all 23 supported languages)

**Domain Skills** (conditional preload based on frameworks):
- React/Vue/Angular → `moai-domain-frontend`
- Express/FastAPI/Django → `moai-domain-backend`
- Flutter/React Native → `moai-domain-mobile-app`
- REST/GraphQL → `moai-domain-web-api`

**Essentials Skills** (always preload):
- `moai-essentials-debug`
- `moai-essentials-refactor`
- `moai-essentials-review`

**Example Preload Sequence** (TypeScript + React project):
```
1. moai-lang-typescript (primary language)
2. moai-domain-frontend (React framework detected)
3. moai-essentials-debug (always)
4. moai-essentials-refactor (always)
5. moai-essentials-review (always)
```

---

## TRUST 5 Configuration Per Language

Each language has specific TRUST 5 quality gates:

### Test Coverage Requirements

| Language | Recommended Framework | Min Coverage | Command |
|----------|---------------------|-------------|---------|
| TypeScript/JavaScript | Jest, Vitest | 85% | `npm test -- --coverage` |
| Python | pytest | 85% | `pytest --cov=src tests/` |
| Go | go test | 80% | `go test -cover ./...` |
| Rust | cargo test | 80% | `cargo tarpaulin` |
| Java/Kotlin | JUnit | 85% | `mvn test jacoco:report` |
| C++ | Google Test | 80% | `lcov --capture` |

### Linter Configuration

| Language | Recommended Linter | Config File | Command |
|----------|------------------|------------|---------|
| TypeScript | ESLint + Biome | `.eslintrc`, `biome.json` | `npm run lint` |
| Python | Ruff | `pyproject.toml` | `ruff check src/` |
| Go | golangci-lint | `.golangci.yml` | `golangci-lint run` |
| Rust | Clippy | `clippy.toml` | `cargo clippy` |
| Java/Kotlin | Ktlint, Checkstyle | `checkstyle.xml` | `mvn checkstyle:check` |

### Type Safety

| Language | Type System | Strict Mode | Config |
|----------|------------|------------|--------|
| TypeScript | Static + structural | `"strict": true` | `tsconfig.json` |
| Python | Gradual (mypy) | `strict = true` | `pyproject.toml` |
| Go | Static + structural | Always strict | Built-in |
| Rust | Static + ownership | Always strict | Built-in |
| Java/Kotlin | Static + nominal | Null safety (Kotlin) | `kotlinOptions` |

---

## Common Detection Failures & Fixes

### Failure Mode 1: No Configuration File

**Symptoms**:
```bash
# No package.json, pyproject.toml, go.mod, etc. found
# Only source files exist
```

**Detection Strategy**:
- Scan for source file extensions only
- Count files per extension
- Assign primary language based on highest count
- Warn user about missing configuration

**Fix**:
```bash
# User should create appropriate config file
# TypeScript: npm init
# Python: touch pyproject.toml
# Go: go mod init <module-name>
# Rust: cargo init
```

### Failure Mode 2: Conflicting Indicators

**Symptoms**:
```bash
# Both package.json and pyproject.toml exist
# But no clear directory separation
```

**Detection Strategy**:
- Classify as polyglot project
- Scan subdirectories for language clustering
- Prompt user to confirm primary vs. secondary languages

**Fix**:
```bash
# User should organize into subdirectories:
mkdir -p frontend backend
mv package.json tsconfig.json frontend/
mv pyproject.toml backend/
```

### Failure Mode 3: Legacy Configuration

**Symptoms**:
```bash
# Python: setup.py without pyproject.toml
# Node.js: bower.json (deprecated)
```

**Detection Strategy**:
- Detect legacy files
- Warn user about deprecated configuration
- Recommend migration path

**Fix**:
```bash
# Python: Migrate to pyproject.toml
# Node.js: Remove bower.json, use package.json only
```

### Failure Mode 4: Monorepo Misdetection

**Symptoms**:
```bash
# Multiple package.json files detected
# But no monorepo configuration (lerna, nx, pnpm-workspace)
```

**Detection Strategy**:
- Count package.json files
- If >1 but no monorepo config → warn user
- Recommend monorepo tools or directory restructuring

**Fix**:
```bash
# Option 1: Set up monorepo (pnpm-workspace, lerna, nx)
# Option 2: Restructure into separate repositories
```

---

## Integration with MoAI-ADK Workflow

### Phase 0: Bootstrap (`/alfred:0-project`)

**Language detection during project initialization**:
```bash
# 1. Detect languages
alfred-lang-detect

# 2. Generate detection report
alfred-lang-report > .moai/config/lang-detection.json

# 3. Preload language Skills
alfred-skill-preload --from-detection

# 4. Configure TRUST 5 gates per language
alfred-trust-config --per-language
```

**Output** (`.moai/config/lang-detection.json`):
```json
{
  "detection_date": "2025-10-22",
  "primary_languages": [
    {
      "language": "TypeScript",
      "confidence": 95,
      "framework": "Next.js",
      "testing": "Jest",
      "package_manager": "pnpm",
      "preloaded_skills": [
        "moai-lang-typescript",
        "moai-domain-frontend"
      ]
    }
  ],
  "trust_config": {
    "typescript": {
      "coverage_min": 85,
      "linter": "eslint + biome",
      "type_check": true
    }
  }
}
```

### Phase 1: Plan (`/alfred:1-plan`)

**Use detected language for SPEC templates**:
```bash
# Generate SPEC with language-specific examples
alfred-spec-create SPEC-001 \
  --language typescript \
  --framework next.js \
  --template web-api
```

### Phase 2: Run (`/alfred:2-run`)

**Use detected testing framework for TDD cycle**:
```bash
# RED: Create failing test (language-specific)
alfred-test-create AUTH-001 --framework jest

# GREEN: Implement feature
alfred-code-implement AUTH-001 --language typescript

# REFACTOR: Lint and format
alfred-lint --language typescript
```

### Phase 3: Sync (`/alfred:3-sync`)

**Run language-specific quality gates**:
```bash
# Run tests for all detected languages
for lang in $(alfred-lang-list --primary); do
  alfred-test-run --language "$lang" --coverage
done

# Validate coverage ≥85% per language
alfred-coverage-check --min 85 --per-language
```

---

## Best Practices Summary

### DO:
- ✅ Use explicit configuration files (package.json, pyproject.toml, go.mod, etc.)
- ✅ Organize polyglot projects into clear subdirectories (frontend/, backend/)
- ✅ Declare all dependencies explicitly in config files
- ✅ Use standard testing frameworks recommended per language
- ✅ Maintain separate TRUST 5 coverage gates per language
- ✅ Document polyglot architecture in project README
- ✅ Use monorepo tools (pnpm-workspace, lerna, nx) for multi-package projects
- ✅ Keep language-specific configuration files at appropriate directory levels
- ✅ Validate detection results during `/alfred:0-project`

### DON'T:
- ❌ Mix source files from multiple languages in the same directory
- ❌ Use deprecated configuration files (setup.py without pyproject.toml)
- ❌ Rely on source file detection alone (always provide config files)
- ❌ Skip testing framework declaration in config files
- ❌ Use conflicting package managers (npm + pnpm in the same project)
- ❌ Ignore language detection warnings during project initialization
- ❌ Forget to update .moai/config/lang-detection.json after adding new languages
- ❌ Commit node_modules/, __pycache__/, or other language-specific build artifacts

---

## Dependencies

- Access to project files via Read/Bash tools
- Integration with `moai-foundation-trust` for quality gates
- Integration with language-specific Skills (moai-lang-*)
- Integration with domain Skills (moai-domain-*)
- JSON parsing tools (jq) for configuration file analysis

---

## Works Well With

- `moai-foundation-trust` — TRUST 5 quality gates per language
- `moai-lang-typescript` — TypeScript-specific best practices
- `moai-lang-python` — Python-specific best practices
- `moai-lang-go` — Go-specific best practices
- `moai-lang-rust` — Rust-specific best practices
- `moai-alfred-language-detection` — Alfred sub-agent integration
- `moai-domain-frontend` — Frontend framework detection
- `moai-domain-backend` — Backend framework detection

---

## References (Latest Documentation)

**Language Detection Libraries**:
- [Polyglot Language Detection](https://polyglot.readthedocs.io/en/latest/Detection.html)
- [Linguist (GitHub)](https://github.com/github-linguist/linguist) — GitHub's language detection
- [Tokei](https://github.com/XAMPPRocky/tokei) — Language statistics

**Package Management**:
- [npm Documentation](https://docs.npmjs.com/)
- [pnpm Documentation](https://pnpm.io/)
- [Poetry Documentation](https://python-poetry.org/docs/)
- [uv Documentation](https://github.com/astral-sh/uv)
- [Cargo Book](https://doc.rust-lang.org/cargo/)
- [Go Modules Reference](https://go.dev/ref/mod)

**Testing Frameworks**:
- [Jest Documentation](https://jestjs.io/)
- [pytest Documentation](https://docs.pytest.org/)
- [Go Testing Package](https://pkg.go.dev/testing)
- [Google Test Documentation](https://google.github.io/googletest/)

_Documentation links updated 2025-10-22_

---

## Changelog

- **v2.0.0** (2025-10-22): Major expansion with 23+ language support, polyglot project detection, framework identification, automatic Skill preloading, TRUST 5 configuration per language, and monorepo support
- **v1.0.0** (2025-03-29): Initial Skill release with basic language detection

---

## License

This Skill is part of the MoAI-ADK project and follows the same license terms.
