---
name: "moai-validation-quality"
version: "4.0.0"
status: "stable"
description: "Enterprise AI-powered quality validation for 21 programming languages: Python, JavaScript, Go, Rust, Java, C/C++, C#, Ruby, PHP, Swift, Kotlin, Scala, Dart, R, Shell, SQL, HTML/CSS, Tailwind CSS, Markdown. Includes test coverage (≥85%), type checking, linting, formatting, security scanning, and Context7 integration for latest tool patterns."
allowed-tools:
  - Bash
  - Read
  - Grep
---

# AI-Powered Enterprise Quality Validation Skill v4.0.0

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-validation-quality |
| **Version** | 4.0.0 Enterprise (2025-11-12) |
| **Tier** | Essential AI-Powered Validation |
| **AI Integration** | ✅ Context7 MCP for latest tool patterns |
| **Supported Languages** | **21 languages** (see Language Categories) |
| **Tool Support** | pytest, mypy, ruff, biome, eslint, go-test, cargo, maven, gradle, phpstan, swiftlint, shellcheck, markdownlint-cli2, and more |
| **Auto-load** | On demand for /alfred:3-sync validation |
| **Trigger cues** | quality, validation, test, coverage, lint, format, sync, check |

---

## 🎯 When to Use

**Automatic Triggers**:
- `/alfred:3-sync` command execution (quality validation phase)
- Release-level quality checks (85% coverage requirement)
- Multi-language project validation
- Pre-commit quality gates

**Manual Invocation**:
- "Validate project code quality across all languages"
- "Check test coverage and type safety"
- "Run linting and formatting checks"
- "Verify TRUST 5 principle compliance"

---

## ✨ Supported Languages (21)

### Compiled Languages
- **C, C++** (gcc/g++, cppcheck, clang-format) → [LANGUAGES-COMPILED.md](LANGUAGES-COMPILED.md)
- **Rust** (cargo, clippy, rustfmt, tarpaulin) → [LANGUAGES-COMPILED.md](LANGUAGES-COMPILED.md)
- **Go** (go test, staticcheck, gofmt) → [LANGUAGES-COMPILED.md](LANGUAGES-COMPILED.md)
- **C#** (dotnet, roslyn-analyzers) → [LANGUAGES-COMPILED.md](LANGUAGES-COMPILED.md)

### Interpreted Languages
- **Python** (pytest, mypy, ruff, black) → [LANGUAGES-INTERPRETED.md](LANGUAGES-INTERPRETED.md)
- **JavaScript/TypeScript** (biome, eslint, prettier, jest, tsc) → [LANGUAGES-INTERPRETED.md](LANGUAGES-INTERPRETED.md)
- **Ruby** (rubocop, rspec, standardrb) → [LANGUAGES-INTERPRETED.md](LANGUAGES-INTERPRETED.md)
- **PHP** (phpstan, phpunit, phpcs) → [LANGUAGES-INTERPRETED.md](LANGUAGES-INTERPRETED.md)

### Mobile Languages
- **Swift** (swiftlint, swiftformat, swift test) → [LANGUAGES-MOBILE.md](LANGUAGES-MOBILE.md)
- **Kotlin** (detekt, ktlint, junit) → [LANGUAGES-MOBILE.md](LANGUAGES-MOBILE.md)
- **Dart** (dart analyze, dart format, dart test) → [LANGUAGES-MOBILE.md](LANGUAGES-MOBILE.md)

### Additional Languages
- **Java** (gradle/maven, checkstyle, spotbugs) → [LANGUAGES-INTERPRETED.md](LANGUAGES-INTERPRETED.md)
- **Scala** (scalafmt, scalatest) → [LANGUAGES-INTERPRETED.md](LANGUAGES-INTERPRETED.md)
- **R** (lintr, testthat) → [LANGUAGES-INTERPRETED.md](LANGUAGES-INTERPRETED.md)
- **Shell** (shellcheck, shfmt, bats) → [LANGUAGES-MARKUP.md](LANGUAGES-MARKUP.md)

### Markup & Configuration Languages
- **HTML/CSS** (htmlhint, stylelint, prettier) → [LANGUAGES-MARKUP.md](LANGUAGES-MARKUP.md)
- **Tailwind CSS** (stylelint with Tailwind config) → [LANGUAGES-MARKUP.md](LANGUAGES-MARKUP.md)
- **Markdown** (markdownlint-cli2, prettier, mermaid-cli) → [LANGUAGES-MARKUP.md](LANGUAGES-MARKUP.md)
- **SQL** (sqlfluff) → [LANGUAGES-MARKUP.md](LANGUAGES-MARKUP.md)

---

## 🔄 VALID Framework (Validation Methodology)

### V - **Verify Tool Availability**
Check if required validation tools are installed. If missing, provide installation guidance (pip, npm, cargo, brew, etc).

### A - **Analyze Project Language**
Detect project language by checking for: `pyproject.toml`, `package.json`, `Cargo.toml`, `go.mod`, `pom.xml`, `build.gradle`, `composer.json`, etc.

### L - **Load Context7 Patterns**
Query Context7 for latest validation patterns:
```python
context7_patterns = await context7.get_library_docs(
    context7_library_id="/quality-validation/standards",
    topic=f"{detected_language} validation test coverage linting formatting 2025",
    tokens=5000
)
```

### I - **Implement Validation** (Direct Tool Execution)
Execute validation tools directly via Bash (no Python wrappers):
```bash
# Python example
pytest --cov=src tests/ --cov-fail-under=85
mypy src/ --strict
ruff check src/
```

### D - **Deliver Results** (Screen First)
1. Report validation results to user FIRST (screen output)
2. Parse tool output and format for readability
3. Generate background documentation (if config enables)

---

## 🎯 Core Validation Workflow

```
1. Load this Skill: Skill("moai-validation-quality")
   ↓
2. Detect Language: Check for language-specific config files
   ↓
3. Get Context7 Patterns: Fetch latest tool patterns from Context7
   ↓
4. Execute Tools (Direct Bash):
   - Test coverage (≥85% required)
   - Type checking (strict mode)
   - Linting (modern tools with fallbacks)
   - Formatting checks
   - Security scanning
   ↓
5. Parse Results: Extract pass/fail status and metrics
   ↓
6. Report to User (Screen Priority):
   ✅ Test coverage: 87% (required: 85%)
   ✅ Type checking: Passed
   ✅ Linting: Passed
   ✅ Formatting: Passed
   ↓
7. Delegate to Agents:
   - spec-status-agent: Update SPEC status to "completed"
   ↓
8. Background Report (Config Check):
   - Only if config.reporting.enabled = true
   - Only if config.reporting.on_completion_only = true
```

---

## 📋 Reference Files

For detailed tool patterns, see:

- **[LANGUAGES-COMPILED.md](LANGUAGES-COMPILED.md)**: C, C++, Rust, Go, C#
- **[LANGUAGES-INTERPRETED.md](LANGUAGES-INTERPRETED.md)**: Python, JavaScript, TypeScript, Java, Ruby, PHP, Scala, R
- **[LANGUAGES-MOBILE.md](LANGUAGES-MOBILE.md)**: Swift, Kotlin, Dart
- **[LANGUAGES-MARKUP.md](LANGUAGES-MARKUP.md)**: HTML/CSS, Tailwind CSS, Markdown, Shell, SQL
- **[TOOL-REFERENCE.md](TOOL-REFERENCE.md)**: Fallback chains, installation commands, Context7 patterns

---

## 🛠️ Tool Fallback Chains (Summary)

Each language has prioritized tools with automatic fallbacks:

```yaml
# Example: Python
python:
  linter: [ruff, flake8, pylint]           # Try ruff first, fallback to flake8
  formatter: [ruff, black, autopep8]       # Modern: ruff, Traditional: black
  type_checker: [mypy, pyright, pyre]      # Type checking options
  test_coverage: [pytest, unittest]        # Testing frameworks

# Example: JavaScript/TypeScript
javascript:
  linter: [biome, eslint, standard]        # Modern: biome, Traditional: eslint
  formatter: [biome, prettier, dprint]     # Modern: biome, Traditional: prettier
  type_checker: [tsc, flow]                # TypeScript or Flow
  test: [jest, vitest, mocha]              # Testing frameworks
```

See **[TOOL-REFERENCE.md](TOOL-REFERENCE.md)** for complete fallback chains for all 21 languages.

---

## 📦 Installation Guidance

Tools are not bundled; users must install them:

### Python Tools
```bash
pip install pytest pytest-cov mypy ruff black bandit
```

### JavaScript Tools
```bash
npm install -g @biomejs/biome eslint prettier jest
```

### Go Tools
```bash
go install honnef.co/go/tools/cmd/staticcheck@latest
```

### Rust Tools
```bash
rustup component add clippy rustfmt
cargo install cargo-tarpaulin
```

### Ruby Tools
```bash
gem install rubocop rspec standardrb
```

### Complete Installation → See **[TOOL-REFERENCE.md](TOOL-REFERENCE.md)**

---

## 🧠 Context7 Integration

This Skill leverages Context7 MCP for latest validation patterns:

```python
# sync-manager agent loads Context7 patterns for detected language
context7_docs = await context7.get_library_docs(
    context7_library_id="/quality-validation/standards",
    topic="pytest coverage mypy ruff biome patterns enterprise 2025",
    tokens=5000
)

# Apply latest best practices automatically
apply_context7_patterns(context7_docs, detected_language)
```

---

## ✅ Best Practices

### ✅ DO:
- ✅ Use Context7 for latest tool patterns
- ✅ Check tool availability before execution
- ✅ Apply fallback chains automatically
- ✅ Report results to screen FIRST
- ✅ Respect config.json reporting policies
- ✅ Guide users on missing tools
- ✅ Validate ≥85% test coverage (mandatory)
- ✅ Enforce strict type checking

### ❌ DON'T:
- ❌ Assume tools are installed
- ❌ Skip language detection
- ❌ Ignore Context7 patterns
- ❌ Generate reports without config check
- ❌ Accept <85% test coverage
- ❌ Run validation without confirmation

---

## 🔗 Integration with Alfred Workflow

```
/alfred:3-sync command
    ↓
sync-manager agent
    ↓
Load: Skill("moai-validation-quality")
    ↓
Detect language → Load reference file → Get Context7 patterns
    ↓
Execute tools (Bash) → Parse results → Report to screen
    ↓
Delegate: tag-agent, spec-status-agent
    ↓
Background report (config check)
```

---

## 📖 Related Skills & Documentation

Works well with:
- `moai-lang-python` (Python patterns)
- `moai-lang-javascript` (JavaScript/TypeScript patterns)
- `moai-lang-go` (Go patterns)
- `moai-lang-rust` (Rust patterns)
- `moai-domain-backend` (API testing patterns)
- `moai-domain-frontend` (UI component testing)
- `moai-essentials-debug` (Debugging validation issues)
- Context7 MCP (Latest tool documentation)

---

## 🚀 Quick Example

**User runs**: `/alfred:3-sync auto SPEC-001`

**sync-manager agent**:
1. Loads `Skill("moai-validation-quality")`
2. Detects Python project (pyproject.toml exists)
3. Reads `LANGUAGES-INTERPRETED.md` (Python section)
4. Gets Context7 patterns (latest pytest, mypy, ruff)
5. Executes:
   ```bash
   pytest --cov=src tests/ --cov-fail-under=85  # 87% ✅
   mypy src/ --strict                             # Passed ✅
   ruff check src/                                # Passed ✅
   ```
6. Reports to user:
   ```
   ✅ Quality validation passed
   - Test coverage: 87% (required: 85%)
   - Type checking: Passed
   - Linting: Passed
   ```
7. Updates SPEC status to "completed"

---

**Last Updated**: 2025-11-12
**Status**: Enterprise v4.0 Stable
**Maintained by**: Alfred SuperAgent
