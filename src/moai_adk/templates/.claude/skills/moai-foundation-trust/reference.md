# TRUST Validation Reference Guide

## Quick Reference

This document provides comprehensive CLI commands, configuration templates, and tool setup instructions for implementing TRUST 5-principles validation across all supported languages.

---

## T - Test First: CLI Command Matrix

### Python (pytest 8.4.2)

**Installation**:
```bash
pip install pytest==8.4.2 pytest-cov==5.0.0 pytest-xdist==3.6.1
```

**Basic Coverage**:
```bash
pytest --cov=src --cov-report=term-missing
```

**Coverage with Threshold**:
```bash
pytest --cov=src --cov-report=term-missing --cov-fail-under=85
```

**Coverage Reports (Multiple Formats)**:
```bash
pytest --cov=src --cov-report=html --cov-report=xml --cov-report=term
```

**Parallel Execution**:
```bash
pytest -n auto --cov=src --cov-report=term-missing
```

**Configuration** (`pyproject.toml`):
```toml
[tool.pytest.ini_options]
testpaths = ["tests"]
python_files = ["test_*.py", "*_test.py"]
addopts = [
    "--cov=src",
    "--cov-report=term-missing",
    "--cov-fail-under=85",
    "--strict-markers",
    "--disable-warnings"
]

[tool.coverage.run]
branch = true
source = ["src"]
omit = ["*/tests/*", "*/migrations/*", "*/__pycache__/*"]

[tool.coverage.report]
exclude_lines = [
    "pragma: no cover",
    "def __repr__",
    "raise AssertionError",
    "raise NotImplementedError",
    "if __name__ == .__main__.:",
    "if TYPE_CHECKING:",
]
```

---

### TypeScript/JavaScript (Vitest 2.0.5)

**Installation**:
```bash
npm install -D vitest@2.0.5 @vitest/coverage-v8@2.0.5
```

**Basic Coverage**:
```bash
vitest run --coverage
```

**Coverage with Threshold**:
```bash
vitest run --coverage --coverage.thresholds.lines=85 --coverage.thresholds.branches=80
```

**Watch Mode**:
```bash
vitest --coverage --watch
```

**Configuration** (`vitest.config.ts`):
```typescript
import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    globals: true,
    environment: 'node',
    coverage: {
      provider: 'v8',
      reporter: ['text', 'html', 'lcov'],
      exclude: [
        'node_modules/',
        'dist/',
        '**/*.spec.ts',
        '**/*.test.ts',
        '**/types/**',
      ],
      thresholds: {
        lines: 85,
        branches: 80,
        functions: 85,
        statements: 85,
      },
    },
  },
});
```

---

### JavaScript (Jest 29.x)

**Installation**:
```bash
npm install -D jest@29 @types/jest
```

**Basic Coverage**:
```bash
jest --coverage
```

**Coverage with Threshold**:
```bash
jest --coverage --coverageThreshold='{"global":{"lines":85,"branches":80}}'
```

**Configuration** (`jest.config.js`):
```javascript
module.exports = {
  testEnvironment: 'node',
  collectCoverageFrom: [
    'src/**/*.{js,jsx}',
    '!src/**/*.test.{js,jsx}',
    '!src/**/__tests__/**',
  ],
  coverageThreshold: {
    global: {
      lines: 85,
      branches: 80,
      functions: 85,
      statements: 85,
    },
  },
  coverageReporters: ['text', 'html', 'lcov'],
};
```

---

### Go (testing 1.23)

**Basic Tests**:
```bash
go test ./...
```

**Coverage Report**:
```bash
go test ./... -cover -coverprofile=coverage.out
```

**Coverage by Package**:
```bash
go test ./... -coverprofile=coverage.out -covermode=atomic
go tool cover -html=coverage.out -o coverage.html
```

**Coverage Threshold Check**:
```bash
# Custom script for threshold enforcement
go test ./... -coverprofile=coverage.out -covermode=atomic
go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//' | awk '{if($1<85)exit 1}'
```

**Makefile Example**:
```makefile
.PHONY: test coverage

test:
	go test -v ./...

coverage:
	go test -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

coverage-check:
	@go test -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -func=coverage.out | grep total | awk '{print "Total coverage: " $$3}'
	@go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//' | \
		awk '{if($$1<85){print "FAIL: Coverage below 85%"; exit 1}else{print "PASS: Coverage >= 85%"}}'
```

---

### Rust (cargo test 1.82.0)

**Basic Tests**:
```bash
cargo test
```

**Coverage with tarpaulin**:
```bash
cargo install cargo-tarpaulin
cargo tarpaulin --out Xml --output-dir coverage/ --fail-under 85
```

**Coverage with HTML Report**:
```bash
cargo tarpaulin --out Html --output-dir coverage/
```

**Configuration** (`Cargo.toml`):
```toml
[package]
name = "my-project"
version = "0.1.0"

[dependencies]

[dev-dependencies]

[profile.test]
opt-level = 0

# Coverage configuration
[package.metadata.tarpaulin]
exclude-files = ["target/*", "tests/*"]
out = ["Html", "Xml"]
output-dir = "coverage/"
fail-under = 85
```

---

### Java/Kotlin (JUnit 5.10.x + JaCoCo)

**Gradle Configuration** (`build.gradle.kts`):
```kotlin
plugins {
    kotlin("jvm") version "1.9.22"
    jacoco
}

jacoco {
    toolVersion = "0.8.11"
}

tasks.test {
    useJUnitPlatform()
    finalizedBy(tasks.jacocoTestReport)
}

tasks.jacocoTestReport {
    dependsOn(tasks.test)
    reports {
        xml.required.set(true)
        html.required.set(true)
        csv.required.set(false)
    }
}

tasks.jacocoTestCoverageVerification {
    violationRules {
        rule {
            limit {
                minimum = "0.85".toBigDecimal()
            }
        }
    }
}
```

**Commands**:
```bash
./gradlew test jacocoTestReport
./gradlew jacocoTestCoverageVerification
```

---

## R - Readable: Linting & Formatting

### Python (ruff 0.6.x)

**Installation**:
```bash
pip install ruff==0.6.8
```

**Commands**:
```bash
ruff check .                    # Check only
ruff check . --fix              # Auto-fix
ruff format .                   # Format code
ruff check . --statistics       # Show statistics
```

**Configuration** (`pyproject.toml`):
```toml
[tool.ruff]
line-length = 100
target-version = "py311"

[tool.ruff.lint]
select = [
    "E",   # pycodestyle errors
    "W",   # pycodestyle warnings
    "F",   # pyflakes
    "I",   # isort
    "N",   # pep8-naming
    "UP",  # pyupgrade
    "B",   # flake8-bugbear
    "C4",  # flake8-comprehensions
    "SIM", # flake8-simplify
]
ignore = ["E501"]  # line too long (handled by formatter)

[tool.ruff.lint.per-file-ignores]
"__init__.py" = ["F401"]  # unused imports
"tests/**" = ["S101"]     # allow assert in tests

[tool.ruff.lint.mccabe]
max-complexity = 10
```

---

### TypeScript/JavaScript (Biome 1.9.x)

**Installation**:
```bash
npm install -D @biomejs/biome@1.9.0
```

**Commands**:
```bash
biome check .                   # Check all
biome check --apply .           # Auto-fix
biome format .                  # Format only
biome lint .                    # Lint only
```

**Configuration** (`biome.json`):
```json
{
  "$schema": "https://biomejs.dev/schemas/1.9.0/schema.json",
  "organizeImports": {
    "enabled": true
  },
  "linter": {
    "enabled": true,
    "rules": {
      "recommended": true,
      "complexity": {
        "noExtraBooleanCast": "error",
        "noMultipleSpacesInRegularExpressionLiterals": "error",
        "noUselessCatch": "error",
        "noUselessConstructor": "error"
      },
      "correctness": {
        "noUnusedVariables": "error",
        "useExhaustiveDependencies": "warn"
      },
      "suspicious": {
        "noExplicitAny": "warn"
      }
    }
  },
  "formatter": {
    "enabled": true,
    "formatWithErrors": false,
    "indentStyle": "space",
    "indentWidth": 2,
    "lineWidth": 100
  },
  "javascript": {
    "formatter": {
      "quoteStyle": "single",
      "semicolons": "always"
    }
  }
}
```

---

### Go (golangci-lint 1.60.x)

**Installation**:
```bash
# Binary installation
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.60.3
```

**Commands**:
```bash
golangci-lint run                    # Run all linters
golangci-lint run --fix              # Auto-fix
golangci-lint run --enable-all       # Enable all linters
```

**Configuration** (`.golangci.yml`):
```yaml
run:
  timeout: 5m
  tests: true

linters:
  enable:
    - gofmt
    - goimports
    - govet
    - staticcheck
    - errcheck
    - gosimple
    - ineffassign
    - unused
    - misspell
    - gocyclo
    - dupl
    - goconst
    - gocognit

linters-settings:
  gocyclo:
    min-complexity: 10
  dupl:
    threshold: 100
  goconst:
    min-len: 3
    min-occurrences: 3
  gocognit:
    min-complexity: 15

issues:
  exclude-use-default: false
  max-issues-per-linter: 0
  max-same-issues: 0
```

---

### Rust (clippy 1.82.0)

**Commands**:
```bash
cargo clippy                          # Basic check
cargo clippy -- -D warnings           # Treat warnings as errors
cargo clippy --all-targets            # Check all targets
cargo fmt                             # Format code
cargo fmt -- --check                  # Check formatting
```

**Configuration** (`clippy.toml`):
```toml
# Clippy configuration
disallowed-methods = [
    "std::env::set_var",
]

cognitive-complexity-threshold = 10
```

---

## U - Unified: Type Checking

### Python (mypy 1.11.x)

**Installation**:
```bash
pip install mypy==1.11.2
```

**Commands**:
```bash
mypy src/                         # Check src directory
mypy src/ --strict                # Strict mode
mypy src/ --no-error-summary      # Quiet mode
mypy src/ --install-types         # Install missing type stubs
```

**Configuration** (`pyproject.toml`):
```toml
[tool.mypy]
python_version = "3.11"
strict = true
warn_return_any = true
warn_unused_configs = true
disallow_untyped_defs = true
disallow_any_generics = true
check_untyped_defs = true
no_implicit_optional = true
warn_redundant_casts = true
warn_unused_ignores = true
warn_no_return = true
warn_unreachable = true

[[tool.mypy.overrides]]
module = "tests.*"
disallow_untyped_defs = false
```

---

### TypeScript (tsc 5.6.x)

**Commands**:
```bash
tsc --noEmit                      # Type check only
tsc --noEmit --strict             # Strict mode
tsc --noEmit --watch              # Watch mode
```

**Configuration** (`tsconfig.json`):
```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "ESNext",
    "lib": ["ES2022"],
    "strict": true,
    "noImplicitAny": true,
    "strictNullChecks": true,
    "strictFunctionTypes": true,
    "strictBindCallApply": true,
    "strictPropertyInitialization": true,
    "noImplicitThis": true,
    "alwaysStrict": true,
    "noUnusedLocals": true,
    "noUnusedParameters": true,
    "noImplicitReturns": true,
    "noFallthroughCasesInSwitch": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true
  },
  "include": ["src/**/*"],
  "exclude": ["node_modules", "dist"]
}
```

---

## S - Secured: SAST Tools

### detect-secrets 1.4.x

**Installation**:
```bash
pip install detect-secrets==1.4.0
```

**Commands**:
```bash
# Create baseline
detect-secrets scan > .secrets.baseline

# Scan against baseline
detect-secrets scan --baseline .secrets.baseline

# Audit findings
detect-secrets audit .secrets.baseline

# Update baseline
detect-secrets scan --baseline .secrets.baseline --update
```

**Pre-commit Hook** (`.pre-commit-config.yaml`):
```yaml
repos:
  - repo: https://github.com/Yelp/detect-secrets
    rev: v1.4.0
    hooks:
      - id: detect-secrets
        args: ['--baseline', '.secrets.baseline']
```

---

### Trivy 0.56.x

**Installation**:
```bash
# macOS
brew install aquasecurity/trivy/trivy

# Linux
wget -qO - https://aquasecurity.github.io/trivy-repo/deb/public.key | sudo apt-key add -
echo "deb https://aquasecurity.github.io/trivy-repo/deb $(lsb_release -sc) main" | sudo tee -a /etc/apt/sources.list.d/trivy.list
sudo apt-get update
sudo apt-get install trivy
```

**Commands**:
```bash
# Filesystem scan
trivy fs .

# Scan with severity filter
trivy fs --severity HIGH,CRITICAL .

# Exit on vulnerabilities
trivy fs --severity HIGH,CRITICAL --exit-code 1 .

# Output formats
trivy fs --format json --output result.json .
trivy fs --format sarif --output result.sarif .
```

**Configuration** (`trivy.yaml`):
```yaml
severity: HIGH,CRITICAL
exit-code: 1
timeout: 5m
cache:
  dir: .trivy-cache
```

---

### Semgrep 1.94.x

**Installation**:
```bash
pip install semgrep==1.94.0
```

**Commands**:
```bash
# Auto configuration (recommended)
semgrep --config=auto .

# Specific rulesets
semgrep --config=p/owasp-top-ten .
semgrep --config=p/security-audit .

# Exit on findings
semgrep --config=auto --error .

# Output formats
semgrep --config=auto --json --output=results.json .
semgrep --config=auto --sarif --output=results.sarif .
```

**Configuration** (`.semgrep.yml`):
```yaml
rules:
  - id: hardcoded-secret
    pattern: |
      password = "..."
    message: Hardcoded password detected
    severity: ERROR
    languages: [python, javascript, typescript]
```

---

### Bandit 1.7.x (Python)

**Installation**:
```bash
pip install bandit==1.7.9
```

**Commands**:
```bash
bandit -r src/                    # Scan src directory
bandit -r src/ -ll                # High severity only
bandit -r src/ -f json            # JSON output
```

**Configuration** (`.bandit`):
```yaml
exclude_dirs:
  - /tests
  - /venv

tests:
  - B201  # flask_debug_true
  - B301  # pickle
  - B302  # marshal
  - B303  # md5
  - B304  # ciphers
  - B305  # cipher_modes
  - B306  # mktemp_q
  - B307  # eval
  - B308  # mark_safe
  - B309  # httpsconnection
  - B310  # urllib_urlopen
```

---

## T - Trackable: TAG Validation

### TAG Scanning Commands

**Find all TAGs**:
```bash
rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/
```

**Find specific TAG**:
```bash
rg '@CODE:AUTH-001' -n src/
```

**Find orphaned TAGs** (CODE without SPEC):
```bash
# 1. Extract all CODE TAGs
rg '@CODE:([A-Z]+-\d+)' -or '$1' src/ | sort -u > code_tags.txt

# 2. Extract all SPEC TAGs
rg '@SPEC:([A-Z]+-\d+)' -or '$1' .moai/specs/ | sort -u > spec_tags.txt

# 3. Find orphans (in CODE but not in SPEC)
comm -23 code_tags.txt spec_tags.txt
```

**TAG Chain Validation Script** (`scripts/validate-tags.sh`):
```bash
#!/bin/bash
set -e

echo "Validating TAG chains..."

# Find all TAGs
ALL_TAGS=$(rg '@(SPEC|TEST|CODE|DOC):([A-Z]+-\d+)' -or '$2' .moai/specs/ tests/ src/ docs/ 2>/dev/null | sort -u)

ERRORS=0

for TAG in $ALL_TAGS; do
  SPEC_COUNT=$(rg "@SPEC:$TAG" -c .moai/specs/ 2>/dev/null | awk -F: '{sum+=$2} END {print sum}')
  CODE_COUNT=$(rg "@CODE:$TAG" -c src/ 2>/dev/null | awk -F: '{sum+=$2} END {print sum}')
  TEST_COUNT=$(rg "@TEST:$TAG" -c tests/ 2>/dev/null | awk -F: '{sum+=$2} END {print sum}')

  echo "TAG $TAG: SPEC=$SPEC_COUNT CODE=$CODE_COUNT TEST=$TEST_COUNT"

  if [ "$CODE_COUNT" -gt 0 ] && [ "$SPEC_COUNT" -eq 0 ]; then
    echo "  ERROR: CODE exists but SPEC missing"
    ERRORS=$((ERRORS + 1))
  fi

  if [ "$CODE_COUNT" -gt 0 ] && [ "$TEST_COUNT" -eq 0 ]; then
    echo "  WARNING: CODE exists but TEST missing"
  fi
done

if [ $ERRORS -gt 0 ]; then
  echo "TAG validation FAILED with $ERRORS errors"
  exit 1
fi

echo "TAG validation PASSED"
```

---

## CI/CD Integration Examples

### GitHub Actions (Python Project)

**File**: `.github/workflows/trust-check.yml`

```yaml
name: TRUST Validation

on:
  pull_request:
  push:
    branches: [main, develop]

jobs:
  trust-gate:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Python
        uses: actions/setup-python@v5
        with:
          python-version: '3.11'
          cache: 'pip'

      - name: Install dependencies
        run: |
          pip install -r requirements.txt
          pip install pytest pytest-cov ruff mypy detect-secrets

      # T - Test First
      - name: Run tests with coverage
        run: |
          pytest --cov=src --cov-report=xml --cov-report=term --cov-fail-under=85

      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v4
        with:
          file: ./coverage.xml
          fail_ci_if_error: true

      # R - Readable
      - name: Lint with ruff
        run: ruff check .

      # U - Unified
      - name: Type check with mypy
        run: mypy src/ --strict

      # S - Secured
      - name: Scan for secrets
        run: detect-secrets scan --baseline .secrets.baseline

      - name: Security scan with Trivy
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          scan-ref: '.'
          severity: 'HIGH,CRITICAL'
          exit-code: '1'

      - name: SAST with Semgrep
        run: |
          pip install semgrep
          semgrep --config=auto --error .

      # T - Trackable
      - name: Validate TAG chains
        run: |
          chmod +x scripts/validate-tags.sh
          ./scripts/validate-tags.sh

      - name: TRUST gate summary
        if: always()
        run: |
          echo "## TRUST Validation Results" >> $GITHUB_STEP_SUMMARY
          echo "- Test coverage: $(grep -oP 'TOTAL.*\K\d+%' coverage.txt || echo 'N/A')" >> $GITHUB_STEP_SUMMARY
          echo "- Linting: $(ruff check . --statistics | tail -1)" >> $GITHUB_STEP_SUMMARY
```

---

### GitHub Actions (TypeScript Project)

**File**: `.github/workflows/trust-check.yml`

```yaml
name: TRUST Validation

on:
  pull_request:
  push:
    branches: [main, develop]

jobs:
  trust-gate:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'

      - name: Install dependencies
        run: npm ci

      # T - Test First
      - name: Run tests with coverage
        run: npm run test:coverage

      - name: Check coverage threshold
        run: |
          COVERAGE=$(npx vitest run --coverage --reporter=json | jq '.coverage.lines.pct')
          if (( $(echo "$COVERAGE < 85" | bc -l) )); then
            echo "Coverage $COVERAGE% is below 85%"
            exit 1
          fi

      # R - Readable
      - name: Lint with Biome
        run: npx biome check .

      # U - Unified
      - name: Type check
        run: npm run type-check

      # S - Secured
      - name: Audit dependencies
        run: npm audit --audit-level=high

      - name: Security scan with Trivy
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          scan-ref: '.'
          severity: 'HIGH,CRITICAL'
          exit-code: '1'

      # T - Trackable
      - name: Validate TAG chains
        run: |
          chmod +x scripts/validate-tags.sh
          ./scripts/validate-tags.sh
```

---

## Local Pre-commit Hooks

### Python Project (`.git/hooks/pre-commit`)

```bash
#!/bin/bash
set -e

echo "Running TRUST validation..."

# T - Test First
echo "→ Running tests..."
pytest --cov=src --cov-fail-under=85 --quiet || exit 1

# R - Readable
echo "→ Linting..."
ruff check . || exit 1

# U - Unified
echo "→ Type checking..."
mypy src/ --strict || exit 1

# S - Secured
echo "→ Checking for secrets..."
detect-secrets scan --baseline .secrets.baseline || exit 1

echo "✓ TRUST validation passed"
```

### TypeScript Project (`.git/hooks/pre-commit`)

```bash
#!/bin/bash
set -e

echo "Running TRUST validation..."

# T - Test First
echo "→ Running tests..."
npm run test:coverage || exit 1

# R - Readable
echo "→ Linting..."
npx biome check . || exit 1

# U - Unified
echo "→ Type checking..."
npm run type-check || exit 1

# S - Secured
echo "→ Auditing dependencies..."
npm audit --audit-level=high || exit 1

echo "✓ TRUST validation passed"
```

---

## Makefile Automation

### Universal TRUST Makefile

```makefile
.PHONY: trust trust-test trust-lint trust-type trust-security trust-tags

# Run all TRUST checks
trust: trust-test trust-lint trust-type trust-security trust-tags
	@echo "✓ All TRUST principles validated"

# T - Test First
trust-test:
	@echo "→ T - Test First (Coverage ≥85%)"
	@pytest --cov=src --cov-fail-under=85 --quiet || exit 1
	@echo "✓ Tests passed with adequate coverage"

# R - Readable
trust-lint:
	@echo "→ R - Readable (Linting)"
	@ruff check . || exit 1
	@echo "✓ Code is readable and well-formatted"

# U - Unified
trust-type:
	@echo "→ U - Unified (Type checking)"
	@mypy src/ --strict || exit 1
	@echo "✓ Type safety verified"

# S - Secured
trust-security:
	@echo "→ S - Secured (Security scans)"
	@detect-secrets scan --baseline .secrets.baseline || exit 1
	@trivy fs --severity HIGH,CRITICAL --exit-code 1 . || exit 1
	@echo "✓ Security scans passed"

# T - Trackable
trust-tags:
	@echo "→ T - Trackable (TAG validation)"
	@bash scripts/validate-tags.sh || exit 1
	@echo "✓ TAG chains validated"

# Quick check (fast subset)
trust-quick: trust-lint trust-type
	@echo "✓ Quick TRUST check passed"
```

---

## Tool Version Management

### Python (requirements-dev.txt)

```txt
# TRUST validation tools (2025-10-22)

# T - Test First
pytest==8.4.2
pytest-cov==5.0.0
pytest-xdist==3.6.1

# R - Readable
ruff==0.6.8

# U - Unified
mypy==1.11.2

# S - Secured
detect-secrets==1.4.0
bandit==1.7.9
semgrep==1.94.0
```

### TypeScript (package.json)

```json
{
  "devDependencies": {
    "@biomejs/biome": "^1.9.0",
    "@vitest/coverage-v8": "^2.0.5",
    "vitest": "^2.0.5",
    "typescript": "^5.6.0"
  },
  "scripts": {
    "test": "vitest run",
    "test:coverage": "vitest run --coverage",
    "test:watch": "vitest --watch",
    "lint": "biome check .",
    "lint:fix": "biome check --apply .",
    "type-check": "tsc --noEmit",
    "trust": "npm run test:coverage && npm run lint && npm run type-check"
  }
}
```

---

## Additional Resources

- **pytest**: https://docs.pytest.org/
- **Vitest**: https://vitest.dev/
- **Jest**: https://jestjs.io/
- **Go testing**: https://go.dev/doc/tutorial/add-a-test
- **Rust testing**: https://doc.rust-lang.org/book/ch11-00-testing.html
- **ruff**: https://docs.astral.sh/ruff/
- **Biome**: https://biomejs.dev/
- **golangci-lint**: https://golangci-lint.run/
- **clippy**: https://github.com/rust-lang/rust-clippy
- **mypy**: https://mypy.readthedocs.io/
- **TypeScript**: https://www.typescriptlang.org/
- **Trivy**: https://trivy.dev/
- **detect-secrets**: https://github.com/Yelp/detect-secrets
- **Semgrep**: https://semgrep.dev/
