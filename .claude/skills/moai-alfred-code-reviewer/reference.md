# Code Review Reference Documentation

_Last updated: 2025-10-22_

## SOLID Principles Reference

### 1. Single Responsibility Principle (SRP)
**Definition**: A class should have one, and only one, reason to change.

**Example Violation**:
```python
class User:
    def save_to_database(self):  # Database concern
        pass
    def send_email(self):        # Email concern
        pass
```

**Corrected**:
```python
class User:
    pass

class UserRepository:
    def save(self, user):
        pass

class EmailService:
    def send(self, user):
        pass
```

### 2. Open/Closed Principle (OCP)
**Definition**: Software entities should be open for extension, but closed for modification.

**Example**: Use interfaces/protocols instead of modifying existing code.

### 3. Liskov Substitution Principle (LSP)
**Definition**: Subtypes must be substitutable for their base types.

### 4. Interface Segregation Principle (ISP)
**Definition**: Clients should not depend on interfaces they don't use.

### 5. Dependency Inversion Principle (DIP)
**Definition**: Depend on abstractions, not concretions.

---

## Static Analysis Tools by Language (2025)

### Python
- **Linter**: ruff (replaces pylint, flake8)
- **Formatter**: black or ruff format
- **Type Checker**: mypy
- **Security**: semgrep, bandit
- **Complexity**: radon

```bash
# Install
pip install ruff mypy semgrep bandit radon

# Run
ruff check .
mypy src/
semgrep scan --config=auto
bandit -r src/
radon cc src/ -a
```

### TypeScript/JavaScript
- **Linter**: Biome (replaces ESLint)
- **Type Checker**: tsc (built-in)
- **Security**: semgrep
- **Formatter**: Biome

```bash
# Install
npm install -g @biomejs/biome

# Run
biome check --write .
tsc --noEmit
semgrep scan --config=auto
```

### Go
- **Linter**: golangci-lint (includes golint, staticcheck, etc.)
- **Formatter**: gofmt (built-in)
- **Security**: gosec
- **Type Checker**: built-in

```bash
# Install
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Run
golangci-lint run
gofmt -d .
gosec ./...
```

### Rust
- **Linter**: clippy (official)
- **Formatter**: rustfmt (official)
- **Security**: cargo-audit

```bash
# Install
rustup component add clippy rustfmt
cargo install cargo-audit

# Run
cargo clippy -- -D warnings
cargo fmt --check
cargo audit
```

---

## Clean Code Principles

### 1. Meaningful Names
- Use intention-revealing names
- Avoid disinformation
- Make meaningful distinctions
- Use pronounceable names

**Bad**: `getUserDataArr()`
**Good**: `getUserProfiles()`

### 2. Functions
- Small (≤50 lines)
- Do one thing
- One level of abstraction
- Descriptive names
- Minimal arguments (≤3)

### 3. Comments
- Explain WHY, not WHAT
- Don't use comments to compensate for bad code
- Use code as documentation

### 4. Error Handling
- Use exceptions, not error codes
- Don't return null (use Optional/Result)
- Don't pass null

---

## Code Complexity Metrics

### Cyclomatic Complexity
**Target**: ≤10 per function

**Calculation**: Number of decision points + 1

```python
# Complexity = 3
def process(status):
    if status == "active":      # +1
        return "ok"
    elif status == "pending":   # +1
        return "wait"
    else:
        return "error"
```

### Cognitive Complexity
**Target**: ≤15 per function

Measures how difficult code is to understand (nesting penalty).

---

## Automated Code Review CI/CD Integration

### GitHub Actions Template
```yaml
name: Code Quality
on: [pull_request]

jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Static Analysis
        run: |
          # Language-specific checks
          ruff check .           # Python
          biome check .          # TS/JS
          golangci-lint run      # Go

      - name: Security Scan
        run: semgrep scan --config=auto

      - name: Coverage Check
        run: |
          pytest --cov=src --cov-report=xml
          # Fail if coverage < 85%
          coverage report --fail-under=85
```

### GitLab CI Template
```yaml
code_review:
  stage: test
  script:
    - ruff check .
    - mypy src/
    - semgrep scan --config=auto
    - pytest --cov=src --cov-fail-under=85
  only:
    - merge_requests
```

---

## Code Review Checklist

### Pre-Review (Automated)
- [ ] All tests pass
- [ ] Coverage ≥85%
- [ ] No linting errors
- [ ] No security vulnerabilities
- [ ] Type checking passes

### During Review (Human)
- [ ] SOLID principles followed
- [ ] Clear, descriptive naming
- [ ] Proper error handling
- [ ] No code duplication
- [ ] Adequate test coverage
- [ ] Documentation updated
- [ ] @TAG references correct

### Post-Review
- [ ] All review comments addressed
- [ ] CI/CD pipeline passes
- [ ] TRUST 5 principles validated
- [ ] Living docs updated

---

## References

- [SOLID Principles - Uncle Bob](https://blog.cleancoder.com/uncle-bob/2020/10/18/Solid-Relevance.html)
- [Clean Code - Robert C. Martin](https://www.oreilly.com/library/view/clean-code-a/9780136083238/)
- [Code Review Best Practices - Google](https://google.github.io/eng-practices/review/)
- [Static Analysis Tools 2025](https://www.aikido.dev/blog/static-code-analysis-tools)
- [Refactoring Catalog - Martin Fowler](https://refactoring.com/catalog/)

---

_For practical examples, see examples.md_
