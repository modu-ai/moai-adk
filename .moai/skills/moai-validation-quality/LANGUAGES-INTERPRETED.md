# Interpreted Languages Validation Patterns

> Referenced from: `moai-validation-quality/SKILL.md`

This document covers validation patterns for interpreted languages: **Python, JavaScript, TypeScript, Java, Ruby, PHP, Scala, R**.

---

## Python

### Technology Stack
- **Language**: Python 3.13.9 (Latest stable, 2025)
- **Package Manager**: pip, uv
- **Test Framework**: pytest, unittest
- **Type Checker**: mypy, pyright, pyre
- **Linter**: ruff (modern), flake8 (traditional)
- **Formatter**: ruff, black

### Validation Tools

#### Test Coverage (≥85% required)
```bash
# pytest with coverage
pytest --cov=src tests/ --cov-report=term --cov-fail-under=85

# With detailed output
pytest --cov=src --cov-report=html --cov-report=term-missing tests/

# Coverage minimum check
pytest --cov=src --cov-fail-under=85 --cov-report=xml
```

#### Type Checking (mypy - strict)
```bash
# Strict mode required
mypy src/ --strict

# With specific checks
mypy src/ --strict --warn-unused-ignores --disallow-untyped-calls

# Check specific file
mypy src/module.py --strict
```

#### Linting (Ruff - modern, Flake8 - fallback)
```bash
# Modern: Ruff (fast, comprehensive)
ruff check src/

# With specific rules
ruff check src/ --select E,W,F

# Traditional fallback: Flake8
flake8 src/ --max-line-length=100
```

#### Formatting (Ruff - modern, Black - traditional)
```bash
# Modern: Ruff
ruff format --check src/
ruff format src/  # Auto-format

# Traditional fallback: Black
black --check src/
black src/  # Auto-format
```

#### Security Scanning (Bandit)
```bash
# Security scan
bandit -r src/ -f json

# Or pip-audit for dependencies
pip-audit
```

### Example Validation Script
```bash
#!/bin/bash
set -e

echo "🐍 Python Validation Starting..."

echo "▶ Testing + Coverage"
pytest --cov=src tests/ --cov-fail-under=85

echo "▶ Type Checking"
mypy src/ --strict

echo "▶ Linting"
ruff check src/ || flake8 src/

echo "▶ Formatting"
ruff format --check src/ || black --check src/

echo "✅ All validations passed!"
```

### Fallback Chain
```yaml
python:
  test_coverage: [pytest, unittest]
  type_checker: [mypy, pyright, pyre]
  linter: [ruff, flake8, pylint]
  formatter: [ruff, black, autopep8]
  security: [bandit, pip-audit]
```

### Context7 Integration
```python
context7_patterns = await context7.get_library_docs(
    context7_library_id="/quality-validation/python-2025",
    topic="pytest coverage mypy ruff black bandit testing patterns"
)
```

---

## JavaScript / TypeScript

### Technology Stack
- **Node.js**: 22.x LTS
- **Package Manager**: npm, pnpm, bun
- **Test Framework**: Jest, Vitest, Mocha
- **Linter**: Biome (modern), ESLint (traditional)
- **Formatter**: Biome, Prettier
- **Type Checker**: TypeScript (tsc)

### Validation Tools

#### Testing
```bash
# Jest with coverage
npm test -- --coverage --coverageThreshold='{"global":{"lines":85}}'

# Vitest
npm run test -- --coverage

# Mocha
mocha --require ts-node/register tests/**/*.test.ts --coverage
```

#### Linting & Formatting (Biome - Modern)
```bash
# Biome (combines linter + formatter)
biome check src/ --apply-unsafe=false

# Or with formatting
biome format src/ --check
```

#### Linting & Formatting (ESLint + Prettier - Traditional)
```bash
# ESLint
npx eslint src/ --max-warnings=0

# Prettier
npx prettier --check src/

# Combined
npx eslint src/ && npx prettier --check src/
```

#### Type Checking (TypeScript)
```bash
# TypeScript compilation check
tsc --noEmit

# With strict mode
tsc --strict --noEmit

# Check specific file
tsc src/index.ts --noEmit
```

### Example Validation Script
```bash
#!/bin/bash
set -e

echo "📘 JavaScript/TypeScript Validation Starting..."

echo "▶ Testing"
npm test -- --coverage

echo "▶ Type Checking"
tsc --noEmit

echo "▶ Linting & Formatting"
if command -v biome &> /dev/null; then
  biome check src/
else
  npx eslint src/ && npx prettier --check src/
fi

echo "✅ All validations passed!"
```

### Fallback Chain
```yaml
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
```

### Context7 Integration
```python
context7_patterns = await context7.get_library_docs(
    context7_library_id="/quality-validation/javascript-2025",
    topic="jest vitest biome eslint prettier typescript testing patterns"
)
```

---

## Java

### Technology Stack
- **Language**: Java 21 (LTS, September 2023)
- **Build System**: Maven 3.9+, Gradle 8.x
- **Test Framework**: JUnit 5, TestNG
- **Code Analysis**: Checkstyle, SpotBugs, PMD

### Validation Tools

#### Build + Compile
```bash
# Maven
mvn clean compile -DfailOnWarning=true

# Gradle
./gradlew build
```

#### Testing
```bash
# Maven
mvn test

# Gradle
./gradlew test --info

# With coverage (JaCoCo)
mvn test jacoco:report
```

#### Code Analysis (Checkstyle)
```bash
# Maven
mvn checkstyle:check

# Gradle
./gradlew checkstyleMain

# Or SpotBugs
mvn spotbugs:check
```

### Fallback Chain
```yaml
java:
  build: [maven, gradle]
  test: [junit5, testng]
  code_analysis: [checkstyle, spotbugs, pmd]
  coverage: [jacoco]
```

---

## Ruby

### Technology Stack
- **Language**: Ruby 3.3.x
- **Package Manager**: Bundler, RubyGems
- **Test Framework**: RSpec, Minitest
- **Linter**: Rubocop, StandardRB

### Validation Tools

#### Linting
```bash
# Rubocop
rubocop --fail-level W

# StandardRB (opinionated)
standardrb
```

#### Testing
```bash
# RSpec
rspec --format documentation

# With coverage
rspec --format documentation --coverage
```

### Fallback Chain
```yaml
ruby:
  linter: [rubocop, standardrb]
  formatter: [rubocop, standardrb]
  test: [rspec, minitest]
```

---

## PHP

### Technology Stack
- **Language**: PHP 8.4
- **Package Manager**: Composer
- **Test Framework**: PHPUnit, Pest
- **Static Analysis**: PHPStan, Psalm

### Validation Tools

#### Syntax Check
```bash
# Find all PHP files and check syntax
find src -name "*.php" -exec php -l {} \;
```

#### Linting (PHP_CodeSniffer)
```bash
# Check PSR-12 standard
phpcs --standard=PSR12 src/

# Auto-fix
phpcbf --standard=PSR12 src/
```

#### Static Analysis (PHPStan)
```bash
# Analyze code
phpstan analyse src/ --level=max

# With config
phpstan analyse --configuration phpstan.neon
```

#### Testing
```bash
# PHPUnit with coverage
phpunit --coverage-text --coverage-html coverage/
```

### Fallback Chain
```yaml
php:
  linter: [phpcs, phpstan, psalm]
  formatter: [php-cs-fixer]
  test: [phpunit, pest]
  static_analysis: [phpstan, psalm]
```

---

## Scala

### Technology Stack
- **Language**: Scala 3.3.x
- **Build System**: sbt
- **Test Framework**: ScalaTest, Specs2

### Validation Tools

#### Compilation
```bash
# sbt compile
sbt compile
```

#### Testing
```bash
# Run tests
sbt test
```

#### Formatting (Scalafmt)
```bash
# Check format
sbt scalafmtCheck

# Auto-format
sbt scalafmt
```

#### Linting (Scalafix)
```bash
# Run linting
sbt 'scalafix --check'
```

### Fallback Chain
```yaml
scala:
  compiler: [sbt]
  test: [scalatest, specs2]
  formatter: [scalafmt]
  linter: [scalafix, scalastyle]
```

---

## R

### Technology Stack
- **Language**: R 4.3.x
- **Package Manager**: CRAN, renv
- **Test Framework**: testthat

### Validation Tools

#### Linting (lintr)
```bash
# Lint all R files
Rscript -e "lintr::lint_dir('.')"
```

#### Formatting (styler)
```bash
# Format code
Rscript -e "styler::style_dir('.')"

# Check format
Rscript -e "styler::style_file('script.R')"
```

#### Testing
```bash
# Run tests with testthat
Rscript -e "testthat::test_dir('tests')"
```

### Fallback Chain
```yaml
r:
  linter: [lintr]
  formatter: [styler]
  test: [testthat]
```

---

## Installation Commands

### Python Tools
```bash
pip install pytest pytest-cov mypy ruff black bandit pip-audit
```

### JavaScript Tools
```bash
npm install -g @biomejs/biome eslint prettier jest typescript
```

### Java Tools
```bash
# Maven (macOS)
brew install maven

# Gradle - included in projects
./gradlew build

# Linux
sudo apt-get install maven gradle default-jdk-headless
```

### Ruby Tools
```bash
gem install rubocop rspec standardrb
```

### PHP Tools
```bash
composer require --dev phpunit/phpunit phpstan/phpstan php-cs-fixer/php-cs-fixer squizlabs/php_codesniffer
```

### Scala Tools
```bash
# sbt
brew install sbt

# Or from https://www.scala-sbt.org/download.html
```

### R Tools
```bash
# R packages
Rscript -e 'install.packages(c("lintr", "styler", "testthat"))'
```

---

**Last Updated**: 2025-11-12
**Related**: [SKILL.md](SKILL.md), [LANGUAGES-MOBILE.md](LANGUAGES-MOBILE.md), [TOOL-REFERENCE.md](TOOL-REFERENCE.md)
