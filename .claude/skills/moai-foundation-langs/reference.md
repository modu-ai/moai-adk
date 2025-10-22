# Language Detection Reference

> Auto-detect project languages and recommend tooling

_Last updated: 2025-10-22_

---

## Language Detection Patterns

### Detection Strategy

MoAI-ADK auto-detects languages by scanning for manifest files and configuration patterns:

| Language | Manifest Files | Config Files | Indicators |
|----------|---------------|--------------|------------|
| **JavaScript/TypeScript** | `package.json`, `package-lock.json` | `tsconfig.json`, `.eslintrc`, `jest.config.js` | `.js`, `.ts`, `.jsx`, `.tsx` |
| **Python** | `pyproject.toml`, `setup.py`, `requirements.txt` | `pytest.ini`, `ruff.toml`, `mypy.ini` | `.py`, `.pyi` |
| **Go** | `go.mod`, `go.sum` | `go.work` | `.go` |
| **Rust** | `Cargo.toml`, `Cargo.lock` | `rust-toolchain.toml` | `.rs` |
| **Java/Kotlin** | `pom.xml`, `build.gradle`, `gradle.properties` | `settings.gradle` | `.java`, `.kt` |
| **Ruby** | `Gemfile`, `Gemfile.lock` | `.rubocop.yml`, `spec_helper.rb` | `.rb`, `.rake` |
| **PHP** | `composer.json`, `composer.lock` | `phpunit.xml` | `.php` |
| **C/C++** | `CMakeLists.txt`, `Makefile`, `meson.build` | `compile_commands.json` | `.c`, `.cpp`, `.h`, `.hpp` |
| **Swift** | `Package.swift` | `.swiftpm/` | `.swift` |
| **Dart/Flutter** | `pubspec.yaml`, `pubspec.lock` | `analysis_options.yaml` | `.dart` |

### Multi-Language Projects

**Priority Order** (when multiple languages detected):
1. Primary language (most files)
2. Framework-specific (e.g., TypeScript for Next.js)
3. Ecosystem dominance (e.g., Python for ML projects)

**Example Detection Result**:
```json
{
  "primary": "typescript",
  "detected": ["typescript", "python", "shell"],
  "frameworks": ["nextjs", "fastapi"],
  "test_tools": ["vitest", "pytest"],
  "recommended_skills": [
    "moai-lang-typescript",
    "moai-lang-python",
    "moai-domain-frontend",
    "moai-domain-backend"
  ]
}
```

---

## Recommended Tooling by Language

### JavaScript/TypeScript

**Package Manager**:
- `npm` (Node.js default)
- `pnpm` (faster, disk-efficient)
- `yarn` (stable alternative)

**Testing**:
- `vitest` (modern, fast, Vite-native)
- `jest` (established, comprehensive)

**Linting/Formatting**:
- `biome` (all-in-one, fast)
- `eslint` + `prettier` (traditional)

**Type Checking**:
- `tsc` (TypeScript compiler)

### Python

**Package Manager**:
- `uv` (fastest, modern)
- `pip` (standard)
- `poetry` (dependency management)

**Testing**:
- `pytest` (de facto standard)

**Linting/Formatting**:
- `ruff` (all-in-one, blazing fast)
- `black` (opinionated formatting)

**Type Checking**:
- `mypy` (static type checker)

### Go

**Tooling** (built-in):
- `go test` (testing)
- `go fmt` (formatting)
- `go vet` (linting)
- `golangci-lint` (comprehensive linting)

### Rust

**Tooling** (built-in):
- `cargo test` (testing)
- `cargo fmt` (formatting with rustfmt)
- `cargo clippy` (linting)

---

## Detection Algorithm

### File Scanning Process

```
1. Scan root directory for manifest files
   ├─ package.json → JavaScript/TypeScript
   ├─ pyproject.toml → Python
   ├─ Cargo.toml → Rust
   ├─ go.mod → Go
   └─ (check all language patterns)

2. Count source files by extension
   ├─ .ts, .tsx → TypeScript
   ├─ .py → Python
   ├─ .go → Go
   └─ (analyze file distribution)

3. Detect frameworks from dependencies
   ├─ package.json: "next": "^14.0.0" → Next.js
   ├─ requirements.txt: fastapi → FastAPI
   └─ (parse dependency manifests)

4. Recommend Skill packs
   ├─ Primary language → lang-* Skill
   ├─ Detected frameworks → domain-* Skills
   └─ Testing tools → essentials-* Skills
```

### Shell Script Example

```bash
#!/bin/bash
# Language detection script

detect_language() {
  local dir="${1:-.}"

  # Check for TypeScript/JavaScript
  if [[ -f "$dir/package.json" ]]; then
    if [[ -f "$dir/tsconfig.json" ]] || grep -q '"typescript"' "$dir/package.json"; then
      echo "typescript"
    else
      echo "javascript"
    fi
    return
  fi

  # Check for Python
  if [[ -f "$dir/pyproject.toml" ]] || [[ -f "$dir/setup.py" ]] || [[ -f "$dir/requirements.txt" ]]; then
    echo "python"
    return
  fi

  # Check for Go
  if [[ -f "$dir/go.mod" ]]; then
    echo "go"
    return
  fi

  # Check for Rust
  if [[ -f "$dir/Cargo.toml" ]]; then
    echo "rust"
    return
  fi

  echo "unknown"
}

# Usage
LANG=$(detect_language "/path/to/project")
echo "Detected language: $LANG"
```

---

## Skill Preloading Strategy

Based on detected language, MoAI-ADK preloads relevant Skills:

### TypeScript Project

**Auto-loaded Skills**:
- `moai-lang-typescript` (primary)
- `moai-essentials-debug` (debugging)
- `moai-essentials-review` (code review)
- `moai-foundation-trust` (quality gates)

**Conditional Skills** (based on frameworks):
- Next.js → `moai-domain-frontend`
- Express → `moai-domain-backend`
- Jest/Vitest → Testing skills

### Python Project

**Auto-loaded Skills**:
- `moai-lang-python` (primary)
- `moai-essentials-debug`
- `moai-essentials-perf`
- `moai-foundation-trust`

**Conditional Skills**:
- FastAPI/Django → `moai-domain-backend`
- pytest → Testing skills
- NumPy/Pandas → `moai-domain-data-science`
- TensorFlow/PyTorch → `moai-domain-ml`

---

## Framework Detection Patterns

### Frontend Frameworks

```json
{
  "next": "Next.js",
  "react": "React",
  "vue": "Vue.js",
  "@angular/core": "Angular",
  "svelte": "Svelte"
}
```

### Backend Frameworks

```json
{
  "express": "Express.js",
  "fastify": "Fastify",
  "fastapi": "FastAPI",
  "django": "Django",
  "flask": "Flask",
  "gin": "Gin (Go)",
  "axum": "Axum (Rust)"
}
```

### Testing Frameworks

```json
{
  "vitest": "Vitest",
  "jest": "Jest",
  "pytest": "pytest",
  "go test": "Go testing",
  "cargo test": "Rust testing",
  "junit": "JUnit (Java)"
}
```

---

## Resources

**Language Documentation**:
- JavaScript/TypeScript: https://www.typescriptlang.org/docs/
- Python: https://docs.python.org/3/
- Go: https://go.dev/doc/
- Rust: https://doc.rust-lang.org/

**Package Managers**:
- npm: https://docs.npmjs.com/
- uv (Python): https://github.com/astral-sh/uv
- cargo: https://doc.rust-lang.org/cargo/

---

**Last Updated**: 2025-10-22
**Maintained by**: MoAI-ADK Foundation Team
