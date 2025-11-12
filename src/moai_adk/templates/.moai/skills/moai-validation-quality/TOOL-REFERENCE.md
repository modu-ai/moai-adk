# Tool Reference: Fallback Chains & Installation

> Referenced from: `moai-validation-quality/SKILL.md`

This document provides complete tool information for all 21 languages: fallback chains, installation commands, and Context7 integration patterns.

---

## 🔄 Complete Fallback Chains (All Languages)

### Compiled Languages

```yaml
c:
  compiler: [gcc, clang]
  static_analysis: [cppcheck, splint]
  memory_check: [valgrind, address-sanitizer]
  formatter: [clang-format]

cpp:
  compiler: [g++, clang++]
  static_analysis: [cppcheck, clang-tidy, cpplint]
  memory_check: [valgrind, thread-sanitizer]
  formatter: [clang-format]

csharp:
  compiler: [dotnet]
  test: [xunit, nunit, mstest]
  analyzer: [roslyn-analyzers]
  formatter: [dotnet-format]
  style_check: [editorconfig]

go:
  test: [go-test]
  static_analysis: [staticcheck, go-vet, golint]
  linter: [golangci-lint, gometalinter]
  formatter: [gofmt, goimports]

rust:
  test: [cargo-test]
  linter: [cargo-clippy]
  formatter: [rustfmt]
  security: [cargo-audit]
  coverage: [tarpaulin, llvm-cov]
```

### Interpreted Languages

```yaml
python:
  test_coverage: [pytest, unittest]
  type_checker: [mypy, pyright, pyre]
  linter: [ruff, flake8, pylint]
  formatter: [ruff, black, autopep8]
  security: [bandit, pip-audit]

javascript:
  linter: [biome, eslint, standard]
  formatter: [biome, prettier, dprint]
  test: [jest, vitest, mocha]
  type_checker: [tsc, flow]

typescript:
  linter: [biome, eslint]
  formatter: [biome, prettier]
  type_checker: [tsc]
  test: [jest, vitest]

java:
  build: [maven, gradle]
  test: [junit5, testng]
  code_analysis: [checkstyle, spotbugs, pmd]
  coverage: [jacoco]

kotlin:
  build: [gradle, maven]
  test: [junit5, kotlin-test]
  linter: [detekt, ktlint]
  formatter: [ktlint, kotlinter]

php:
  linter: [phpcs, phpstan, psalm]
  formatter: [php-cs-fixer]
  test: [phpunit, pest]
  static_analysis: [phpstan, psalm]

ruby:
  linter: [rubocop, standardrb]
  formatter: [rubocop, standardrb]
  test: [rspec, minitest]

r:
  linter: [lintr]
  formatter: [styler]
  test: [testthat]

scala:
  compiler: [sbt]
  test: [scalatest, specs2]
  formatter: [scalafmt]
  linter: [scalafix, scalastyle]
```

### Mobile Languages

```yaml
dart:
  analyzer: [dart-analyze]
  formatter: [dart-format]
  test: [dart-test]
  coverage: [dart-coverage]
  metrics: [dart-code-metrics]

kotlin:
  build: [gradle]
  test: [junit5]
  linter: [detekt]
  formatter: [ktlint]

swift:
  build: [swift-build, xcodebuild]
  test: [swift-test, xctest]
  linter: [swiftlint]
  formatter: [swiftformat]
```

### Markup & Configuration Languages

```yaml
html:
  validator: [htmlhint, html-validate]
  formatter: [prettier]

css:
  linter: [stylelint, csslint]
  formatter: [prettier, postcss]

tailwind_css:
  linter: [stylelint]
  formatter: [prettier, prettier-plugin-tailwindcss]
  builder: [tailwindcss-cli]

markdown:
  linter: [markdownlint-cli2, markdownlint, remark-lint]
  formatter: [prettier, remark]
  mermaid: [mermaid-cli]

shell:
  linter: [shellcheck]
  formatter: [shfmt, bash-formatter]
  test: [bats, shunit2]

sql:
  linter: [sqlfluff, sqlint]
  formatter: [pgformatter, sqlfmt, sqlparse]
```

---

## 📦 Installation Commands by Language

### C / C++
```bash
# macOS
brew install gcc clang cmake cppcheck clang-format valgrind

# Linux (Ubuntu)
sudo apt-get install build-essential clang cppcheck clang-format valgrind

# Google Test
git clone https://github.com/google/googletest.git
cd googletest && mkdir build && cd build && cmake .. && make install
```

### Go
```bash
# Already includes: go test, gofmt, go vet
# Install additional:
go install honnef.co/go/tools/cmd/staticcheck@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### Rust
```bash
# Install Rust (includes cargo, rustfmt, clippy)
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh

# Update
rustup update

# Tarpaulin (coverage)
cargo install cargo-tarpaulin
```

### C#
```bash
# macOS
brew install dotnet

# Linux
sudo apt-get install dotnet-sdk-9.0

# Windows: Download from https://dotnet.microsoft.com/download
```

### Python
```bash
# All tools
pip install pytest pytest-cov mypy ruff black bandit pip-audit

# Modern tools only (recommended)
pip install pytest pytest-cov mypy ruff

# Traditional tools (fallback)
pip install pytest pytest-cov mypy flake8 black
```

### JavaScript / TypeScript
```bash
# Modern (Biome)
npm install -g @biomejs/biome

# Traditional
npm install -g eslint prettier jest typescript

# All-in-one
npm install -g @biomejs/biome eslint prettier jest typescript
```

### Java
```bash
# macOS
brew install maven gradle default-jdk

# Linux
sudo apt-get install maven gradle default-jdk-headless

# Windows: Download from https://maven.apache.org, https://gradle.org
```

### Kotlin
```bash
# Gradle (usually included in projects)
brew install gradle

# ktlint
brew install ktlint

# Detekt (via Gradle plugin in project)
```

### PHP
```bash
# Composer + tools
composer require --dev \
  phpunit/phpunit \
  phpstan/phpstan \
  php-cs-fixer/php-cs-fixer \
  squizlabs/php_codesniffer
```

### Ruby
```bash
# All tools
gem install rubocop rspec standardrb

# Or with bundler (recommended)
bundle add rubocop rspec standardrb --group development
```

### Swift
```bash
# macOS
brew install swiftlint swiftformat

# Linux: Build from source
# https://github.com/realm/SwiftLint/releases
# https://github.com/nicklockwood/SwiftFormat/releases
```

### Kotlin
```bash
# Via Gradle (in project)
# Or directly:
brew install ktlint

# Detekt via Gradle plugin
```

### Dart
```bash
# Install Dart SDK
brew install dart

# Global packages
dart pub global activate dart_code_metrics

# Flutter (includes Dart)
flutter pub global activate dart_code_metrics
```

### R
```bash
# Install R packages
Rscript -e 'install.packages(c("lintr", "styler", "testthat", "devtools"))'

# Or in R console:
# install.packages(c("lintr", "styler", "testthat"))
```

### HTML / CSS / Markdown
```bash
# All tools
npm install -g htmlhint stylelint prettier markdownlint-cli2 @mermaid-js/mermaid-cli

# Specific:
npm install -g htmlhint          # HTML
npm install -g stylelint         # CSS
npm install -g markdownlint-cli2 # Markdown
npm install -g @mermaid-js/mermaid-cli  # Mermaid
```

### Tailwind CSS
```bash
# Install Tailwind
npm install -g tailwindcss

# Prettier + Tailwind plugin
npm install -g prettier prettier-plugin-tailwindcss

# Stylelint
npm install -g stylelint stylelint-config-standard
```

### Shell Script
```bash
# ShellCheck
brew install shellcheck  # macOS
sudo apt-get install shellcheck  # Linux

# shfmt
brew install shfmt
# Linux: Download from https://github.com/mvdan/sh/releases

# BATS
brew install bats-core
npm install -g bats
```

### SQL
```bash
# sqlfluff
pip install sqlfluff

# pgformatter
brew install pgformatter  # macOS
pip install pgformatter   # Python

# sqlparse
pip install sqlparse
```

---

## 🧠 Context7 Integration Patterns

### General Pattern
```python
# Sync-manager agent calls this for each language
context7_patterns = await context7.get_library_docs(
    context7_library_id="/quality-validation/standards",
    topic=f"{detected_language} {tool_names} testing patterns 2025",
    tokens=5000
)
```

### Language-Specific Examples

#### Python
```python
context7_patterns = await context7.get_library_docs(
    context7_library_id="/quality-validation/standards",
    topic="Python pytest mypy ruff coverage testing patterns enterprise 2025"
)
```

#### JavaScript
```python
context7_patterns = await context7.get_library_docs(
    context7_library_id="/quality-validation/standards",
    topic="JavaScript TypeScript biome eslint jest testing patterns 2025"
)
```

#### Go
```python
context7_patterns = await context7.get_library_docs(
    context7_library_id="/quality-validation/standards",
    topic="Go testing coverage staticcheck gofmt patterns 2025"
)
```

#### Rust
```python
context7_patterns = await context7.get_library_docs(
    context7_library_id="/quality-validation/standards",
    topic="Rust cargo clippy rustfmt testing patterns safety 2025"
)
```

#### Multi-Language
```python
context7_patterns = await context7.get_library_docs(
    context7_library_id="/quality-validation/standards",
    topic="multi-language testing coverage linting formatting patterns 2025"
)
```

---

## ⚙️ Configuration Files

### Python (.pylintrc, pyproject.toml)
```toml
# pyproject.toml
[tool.pytest.ini_options]
minversion = "7.0"
testpaths = ["tests"]
pythonpath = ["src"]
addopts = "--strict-markers --tb=short"

[tool.mypy]
python_version = "3.13"
strict = true
warn_unused_ignores = true

[tool.ruff]
line-length = 100
target-version = "py313"
```

### JavaScript (package.json)
```json
{
  "scripts": {
    "lint": "biome check src/ || eslint src/",
    "format": "biome format src/ || prettier --write src/",
    "test": "jest --coverage"
  },
  "devDependencies": {
    "@biomejs/biome": "latest",
    "eslint": "latest",
    "prettier": "latest",
    "jest": "latest"
  }
}
```

### Go (go.mod, .golangci.yml)
```yaml
# .golangci.yml
linters:
  enable:
    - staticcheck
    - golint
    - vet

output:
  format: colored-line-number
```

### Rust (Cargo.toml)
```toml
[dev-dependencies]
criterion = "0.5"

[profile.test]
opt-level = 1
```

### Java (pom.xml, build.gradle)
```xml
<!-- pom.xml -->
<plugin>
  <groupId>org.apache.maven.plugins</groupId>
  <artifactId>maven-surefire-plugin</artifactId>
  <version>3.0.0</version>
</plugin>
```

### Markdown (.markdownlint.json)
```json
{
  "default": true,
  "line-length": false,
  "no-multiple-blanks": false,
  "MD003": {
    "style": "consistent"
  }
}
```

---

## 🔗 Tool Links & Resources

| Language | Primary Tool | Homepage |
|----------|--------------|----------|
| Python | pytest | https://pytest.org |
| Python | mypy | https://mypy.readthedocs.io |
| Python | ruff | https://docs.astral.sh/ruff |
| JavaScript | biome | https://biomejs.dev |
| JavaScript | eslint | https://eslint.org |
| JavaScript | prettier | https://prettier.io |
| Go | go test | https://golang.org/pkg/testing |
| Go | staticcheck | https://staticcheck.dev |
| Rust | cargo | https://doc.rust-lang.org/cargo |
| Rust | clippy | https://doc.rust-lang.org/clippy |
| Java | maven | https://maven.apache.org |
| Java | gradle | https://gradle.org |
| PHP | phpstan | https://phpstan.org |
| Ruby | rubocop | https://docs.rubocop.org |
| Swift | swiftlint | https://realm.github.io/SwiftLint |
| Kotlin | detekt | https://detekt.dev |
| Dart | dart | https://dart.dev |
| Markdown | markdownlint | https://github.com/DavidAnson/markdownlint |
| Shell | shellcheck | https://www.shellcheck.net |
| SQL | sqlfluff | https://www.sqlfluff.com |

---

## 🎯 Quick Decision Tree

```
Detected Language?
├─ Python
│  ├─ Modern? → ruff + mypy + pytest
│  └─ Traditional? → flake8 + mypy + pytest
├─ JavaScript/TypeScript
│  ├─ Modern? → biome
│  └─ Traditional? → eslint + prettier
├─ Go
│  └─ staticcheck + gofmt + go test
├─ Rust
│  └─ clippy + rustfmt + cargo test
├─ Java
│  └─ maven/gradle + checkstyle + junit
├─ C/C++
│  └─ gcc/g++ + cppcheck + clang-format
├─ Shell
│  └─ shellcheck + shfmt + bats
├─ Markdown
│  └─ markdownlint-cli2 + prettier
└─ Other
   └─ Detect and apply appropriate tools
```

---

**Last Updated**: 2025-11-12
**Related**: [SKILL.md](SKILL.md), [LANGUAGES-*.md files](.)
