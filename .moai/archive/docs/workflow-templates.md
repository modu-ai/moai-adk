# Workflow Templates Guide

## Overview

MoAI-ADK provides language-specific GitHub Actions workflow templates optimized for Test-Driven Development (TDD) with automated tag validation, testing, linting, and coverage reporting.

## Available Templates

### 1. Python (`python-tag-validation.yml`)

**Location**: `src/moai_adk/templates/workflows/python-tag-validation.yml`

**Features**:
- **Test Framework**: pytest with 85% coverage target
- **Type Checking**: mypy (static type checking)
- **Linting**: ruff (fast Python linter)
- **Code Quality**: Automated quality checks
- **Python Versions**: 3.11, 3.12, 3.13 (matrix testing)

**Workflow Triggers**:
- Push to any branch
- Pull request events

**Example Usage**:
```yaml
name: Python TAG Validation
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        python-version: ['3.11', '3.12', '3.13']
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with:
          python-version: ${{ matrix.python-version }}
      - run: pip install -e ".[dev]"
      - run: pytest --cov=. --cov-report=term-missing --cov-fail-under=85
      - run: mypy src/
      - run: ruff check .
```

---

### 2. JavaScript (`javascript-tag-validation.yml`)

**Location**: `src/moai_adk/templates/workflows/javascript-tag-validation.yml`

**Features**:
- **Package Manager**: Auto-detect (npm, yarn, pnpm, bun)
- **Test Framework**: npm test (supports Jest, Vitest, Mocha)
- **Linting**: eslint or biome
- **Coverage Target**: 80%
- **Node Versions**: 20 LTS, 22 LTS (matrix testing)

**Package Manager Detection**:
- Automatically detects based on lock files
- Priority: bun.lockb > pnpm-lock.yaml > yarn.lock > package-lock.json

**Example Usage**:
```yaml
name: JavaScript TAG Validation
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        node-version: ['20', '22']
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: ${{ matrix.node-version }}
      - run: npm ci  # or: pnpm install, yarn install, bun install
      - run: npm test
      - run: npm run lint
```

---

### 3. TypeScript (`typescript-tag-validation.yml`)

**Location**: `src/moai_adk/templates/workflows/typescript-tag-validation.yml`

**Features**:
- **Type Checking**: tsc --noEmit (strict type checking)
- **Test Framework**: npm test (Vitest, Jest)
- **Linting**: biome or eslint + prettier
- **Coverage Target**: 85%
- **Node Versions**: 20 LTS, 22 LTS (matrix testing)

**TypeScript-Specific Checks**:
- Strict type validation before tests
- Build verification (no emit mode)
- Type coverage reporting

**Example Usage**:
```yaml
name: TypeScript TAG Validation
on: [push, pull_request]
jobs:
  type-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
      - run: npm ci
      - run: tsc --noEmit
  test:
    needs: type-check
    runs-on: ubuntu-latest
    steps:
      - run: npm test -- --coverage
```

---

### 4. Go (`go-tag-validation.yml`)

**Location**: `src/moai_adk/templates/workflows/go-tag-validation.yml`

**Features**:
- **Test Framework**: go test -v -cover
- **Linting**: golangci-lint (comprehensive Go linter)
- **Format Checking**: gofmt
- **Coverage Target**: 75%
- **Go Versions**: 1.21, 1.22 (matrix testing)

**Go-Specific Checks**:
- Formatting verification (gofmt)
- Comprehensive linting (golangci-lint)
- Race condition detection
- Build verification

**Example Usage**:
```yaml
name: Go TAG Validation
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: ['1.21', '1.22']
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go-version }}
      - run: go test -v -cover -race ./...
      - run: golangci-lint run
      - run: gofmt -d .
```

---

## Template Selection

### Automatic Selection

MoAI-ADK automatically selects the appropriate workflow template based on detected project language:

```python
from moai_adk.core.project.detector import LanguageDetector

detector = LanguageDetector()
language = detector.detect()  # Detect project language

if language in ["python", "javascript", "typescript", "go"]:
    template_path = detector.get_workflow_template_path(language)
    # Copy template to .github/workflows/tag-validation.yml
```

### Manual Selection

You can manually specify a workflow template by copying it to your project:

```bash
# Copy Python workflow
cp src/moai_adk/templates/workflows/python-tag-validation.yml .github/workflows/tag-validation.yml

# Copy TypeScript workflow
cp src/moai_adk/templates/workflows/typescript-tag-validation.yml .github/workflows/tag-validation.yml
```

## Customization

### Modifying Coverage Targets

Edit the workflow file to adjust coverage thresholds:

**Python**:
```yaml
- run: pytest --cov-fail-under=90  # Changed from 85%
```

**JavaScript/TypeScript**:
```yaml
- run: npm test -- --coverage --coverageThreshold='{"global":{"lines":90}}'
```

**Go**:
```yaml
- run: go test -coverprofile=coverage.out -covermode=atomic ./...
- run: go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//' | awk '{if ($1 < 80) exit 1}'
```

### Adding Custom Steps

Insert custom steps in the workflow:

```yaml
jobs:
  test:
    steps:
      - uses: actions/checkout@v4
      - name: Custom pre-test setup
        run: ./scripts/setup-test-db.sh
      - run: npm test
      - name: Custom post-test cleanup
        if: always()
        run: ./scripts/cleanup.sh
```

## Troubleshooting

### Problem: Workflow not running

**Cause**: Template not copied to `.github/workflows/`

**Solution**:
```bash
mkdir -p .github/workflows
cp src/moai_adk/templates/workflows/python-tag-validation.yml .github/workflows/tag-validation.yml
```

### Problem: Coverage check failing

**Cause**: Coverage below threshold

**Solution**: Either improve test coverage or adjust threshold temporarily:
```yaml
# Python
- run: pytest --cov-fail-under=80  # Lower threshold

# JavaScript
- run: npm test -- --coverage --coverageThreshold='{"global":{"lines":75}}'
```

### Problem: Linting errors

**Cause**: Code doesn't meet style guidelines

**Solution**:
```bash
# Python - auto-fix with ruff
ruff check . --fix

# JavaScript/TypeScript - auto-fix with biome
npm run lint:fix

# Go - format code
gofmt -w .
```

## Best Practices

1. **Keep coverage thresholds realistic** - Start at 75-80%, increase gradually
2. **Use matrix testing** - Test across multiple language versions
3. **Enable branch protection** - Require workflow success before merge
4. **Cache dependencies** - Speed up workflow execution
5. **Run locally first** - Test workflows before pushing to CI

## Related Documentation

- [Language Detection Guide](./language-detection-guide.md) - How languages are detected
- [tdd-implementer Agent](./.claude/agents/alfred/tdd-implementer.md) - Workflow generation
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
