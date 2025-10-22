# TRUST 5 Principles Reference

> Test-first, Readable, Unified, Secured, Trackable

_Last updated: 2025-10-22_

---

## The TRUST 5 Principles

### T — Test First (Coverage ≥85%)

**Goal**: Ensure every feature is tested before deployment.

**Requirements**:
- Unit test coverage ≥85%
- Integration tests for critical paths
- TDD workflow (RED → GREEN → REFACTOR)

**Language-Specific Tools**:
- **JavaScript/TypeScript**: Vitest, Jest
- **Python**: pytest, coverage.py
- **Go**: `go test -cover`
- **Rust**: `cargo test` + `cargo tarpaulin`
- **Java/Kotlin**: JUnit, JaCoCo

**Validation**:
```bash
# JavaScript/TypeScript
vitest --coverage
# Coverage threshold: 85%

# Python
pytest --cov=src --cov-report=term-missing
# Ensure coverage >= 85%

# Go
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
# Check total coverage >= 85%
```

---

### R — Readable (Linting + Formatting)

**Goal**: Maintain consistent, readable code across the team.

**Requirements**:
- Automated linting (no warnings)
- Consistent formatting (enforced by CI)
- Intent-revealing names
- Functions ≤50 LOC, Files ≤300 LOC

**Language-Specific Tools**:
- **JavaScript/TypeScript**: Biome, ESLint + Prettier
- **Python**: Ruff (all-in-one)
- **Go**: `gofmt`, `golangci-lint`
- **Rust**: `rustfmt`, `clippy`
- **Java/Kotlin**: ktlint, Spotless

**Validation**:
```bash
# JavaScript/TypeScript
biome check .
biome format .

# Python
ruff check .
ruff format .

# Go
gofmt -l .
golangci-lint run

# Rust
cargo fmt -- --check
cargo clippy -- -D warnings
```

---

### U — Unified (Type Safety + Validation)

**Goal**: Ensure type correctness and data validation.

**Requirements**:
- Static type checking (where applicable)
- Runtime validation for external data
- Schema validation for APIs
- Consistent error handling

**Language-Specific Tools**:
- **TypeScript**: `tsc --noEmit`
- **Python**: mypy, pydantic
- **Go**: Built-in strong typing
- **Rust**: Built-in strong typing
- **Java/Kotlin**: Built-in strong typing

**Validation**:
```bash
# TypeScript
tsc --noEmit

# Python
mypy src/

# Go (type checking is automatic)
go build ./...

# Rust (type checking is automatic)
cargo build
```

---

### S — Secured (Security Scanning)

**Goal**: Prevent security vulnerabilities.

**Requirements**:
- No secrets in code
- Dependency vulnerability scanning
- SAST (Static Application Security Testing)
- OWASP Top 10 compliance

**Language-Specific Tools**:
- **JavaScript/TypeScript**: npm audit, Snyk
- **Python**: safety, bandit
- **Go**: gosec
- **Rust**: cargo-audit
- **Java/Kotlin**: OWASP Dependency-Check

**Validation**:
```bash
# JavaScript/TypeScript
npm audit
# Fix vulnerabilities: npm audit fix

# Python
safety check
bandit -r src/

# Go
gosec ./...

# Rust
cargo audit
```

---

### T — Trackable (TAG Coverage)

**Goal**: Maintain full traceability from SPEC to CODE to TEST.

**Requirements**:
- Every feature has TAG references (@CODE, @TEST, @SPEC)
- No orphan TAGs (CODE without SPEC, etc.)
- TAG chain integrity verified

**Validation**:
```bash
# Scan all TAGs
rg '@(SPEC|CODE|TEST):' -n .moai/ src/ tests/

# Detect orphan TAGs
./scripts/verify-tags.sh

# Generate TAG inventory
./scripts/tag-inventory.sh
```

---

## TRUST Validation Checklist

Before marking PR as "Ready":

```markdown
## TRUST Validation

- [ ] **Test**: Coverage ≥85% (`pytest --cov` / `vitest --coverage`)
- [ ] **Readable**: Linting passed (`ruff check` / `biome check`)
- [ ] **Unified**: Type checking passed (`mypy` / `tsc`)
- [ ] **Secured**: No security warnings (`bandit` / `npm audit`)
- [ ] **Trackable**: TAG chain complete (SPEC ↔ TEST ↔ CODE)
```

---

## Automated TRUST Gate Script

```bash
#!/bin/bash
# trust-gate.sh - Run all TRUST validation checks

set -e

echo "🔍 Running TRUST validation..."

# T - Test Coverage
echo "✓ Test coverage..."
pytest --cov=src --cov-report=term-missing --cov-fail-under=85

# R - Readable
echo "✓ Linting and formatting..."
ruff check .
ruff format --check .

# U - Unified
echo "✓ Type checking..."
mypy src/

# S - Secured
echo "✓ Security scanning..."
bandit -r src/ -ll
safety check

# T - Trackable
echo "✓ TAG validation..."
./scripts/verify-tags.sh

echo "✅ TRUST validation passed!"
exit 0
```

**Usage in CI/CD**:
```yaml
# .github/workflows/trust-gate.yml
name: TRUST Validation

on: [push, pull_request]

jobs:
  trust-gate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run TRUST validation
        run: ./scripts/trust-gate.sh
```

---

## Language-Specific TRUST Configurations

### TypeScript (Next.js)

```json
// package.json
{
  "scripts": {
    "test": "vitest --coverage",
    "lint": "biome check .",
    "format": "biome format .",
    "type-check": "tsc --noEmit",
    "security": "npm audit",
    "trust-gate": "npm run test && npm run lint && npm run type-check && npm run security"
  }
}
```

### Python (FastAPI)

```toml
# pyproject.toml
[tool.pytest.ini_options]
addopts = "--cov=src --cov-report=term-missing --cov-fail-under=85"

[tool.ruff]
line-length = 100

[tool.mypy]
strict = true
```

---

## TRUST Quality Gates

### Gate 1: Pre-Commit
- Linting (R)
- Formatting (R)
- Type checking (U)

### Gate 2: Pre-Push
- Test coverage (T)
- Security scan (S)
- TAG validation (T)

### Gate 3: PR Ready
- All TRUST checks pass
- Code review approved
- Documentation updated

---

## Resources

**Testing Tools**:
- Vitest: https://vitest.dev/
- pytest: https://docs.pytest.org/
- JUnit: https://junit.org/

**Linting/Formatting**:
- Biome: https://biomejs.dev/
- Ruff: https://docs.astral.sh/ruff/
- ESLint: https://eslint.org/

**Security**:
- OWASP Top 10: https://owasp.org/Top10/
- Bandit: https://bandit.readthedocs.io/
- npm audit: https://docs.npmjs.com/cli/audit

---

**Last Updated**: 2025-10-22
**Maintained by**: MoAI-ADK Foundation Team
