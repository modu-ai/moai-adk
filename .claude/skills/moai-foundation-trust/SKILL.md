---
name: moai-foundation-trust
description: Validates TRUST 5-principles (Test 85%+, Readable, Unified, Secured, Trackable). Use when aligning with TRUST governance.
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - TodoWrite
version: 2.0.0
created: 2025-10-22
---

# Foundation: TRUST Validation

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Version | 2.0.0 |
| Created | 2025-10-22 |
| Allowed tools | Read, Write, Edit, Bash, TodoWrite |
| Auto-load | SessionStart (foundation bootstrap), `/alfred:3-sync` |
| Trigger cues | TRUST compliance checks, release readiness reviews, quality gate enforcement |

## What it does

Validates MoAI-ADK's TRUST 5-principles compliance using the latest 2025 testing frameworks, SAST tools, and CI/CD automation to ensure code quality, testability, security, and traceability.

## When to use

- Activates when TRUST compliance or release readiness needs evaluation
- "Check the TRUST principle", "Quality verification", "Check code quality"
- Automatically invoked by `/alfred:3-sync`
- Before merging PR or releasing
- During CI/CD pipeline execution

## How it works

### T - Test First (Coverage ≥85%)

**Supported Testing Frameworks (2025-10-22)**:

| Language | Framework | Version | Coverage Tool | Command |
| --- | --- | --- | --- | --- |
| Python | pytest | 8.4.2 | pytest-cov | `pytest --cov=src --cov-report=term-missing --cov-fail-under=85` |
| TypeScript/JS | Vitest | 2.0.5 | @vitest/coverage-v8 | `vitest run --coverage --coverage.thresholds.lines=85` |
| JavaScript | Jest | 29.x | built-in | `jest --coverage --coverageThreshold='{"global":{"lines":85}}'` |
| Go | testing | 1.23 | built-in | `go test ./... -cover -coverprofile=coverage.out` |
| Rust | cargo test | 1.82.0 | tarpaulin | `cargo tarpaulin --out Xml --output-dir coverage/ --fail-under 85` |
| Java/Kotlin | JUnit | 5.10.x | JaCoCo | `./gradlew test jacocoTestReport` |
| Dart/Flutter | flutter test | 3.x | built-in | `flutter test --coverage --coverage-path=coverage/lcov.info` |

**TDD Cycle Verification**:
- RED: Failing test exists before implementation (`@TEST:ID`)
- GREEN: Implementation passes test (`@CODE:ID`)
- REFACTOR: Code quality improvements documented

**Quality Gates**:
- Line coverage ≥85%
- Branch coverage ≥80%
- No skipped tests in production code
- Test execution time <5 minutes for unit tests

### R - Readable (Code Quality)

**Constraints**:
- File ≤300 LOC
- Function ≤50 LOC
- Parameters ≤5
- Cyclomatic complexity ≤10

**Linting Tools (2025-10-22)**:

| Language | Linter | Version | Command |
| --- | --- | --- | --- |
| Python | ruff | 0.6.x | `ruff check . --fix` |
| TypeScript/JS | Biome | 1.9.x | `biome check --apply .` |
| JavaScript | ESLint | 9.x | `eslint . --fix` |
| Go | golangci-lint | 1.60.x | `golangci-lint run` |
| Rust | clippy | 1.82.0 | `cargo clippy -- -D warnings` |
| Java | Checkstyle | 10.x | `./gradlew checkstyleMain` |

**Formatting Standards**:
- Consistent indentation (language-specific)
- Meaningful variable/function names
- Early return pattern (guard clauses)
- Comments only for "why", not "what"

### U - Unified (Architecture)

**Verification Points**:
- SPEC-driven architecture consistency
- Clear module boundaries and responsibilities
- Type safety (TypeScript strict mode, mypy strict, etc.)
- Interface/Protocol compliance
- Dependency direction (inward toward domain)

**Type Checking Tools**:

| Language | Tool | Version | Command |
| --- | --- | --- | --- |
| Python | mypy | 1.11.x | `mypy src/ --strict` |
| TypeScript | tsc | 5.6.x | `tsc --noEmit --strict` |
| Go | built-in | 1.23 | `go vet ./...` |
| Rust | built-in | 1.82.0 | `cargo check` |

### S - Secured (Security & SAST)

**SAST Tools (2025-10-22)**:

| Tool | Version | Purpose | Command |
| --- | --- | --- | --- |
| detect-secrets | 1.4.x | Secret detection | `detect-secrets scan --baseline .secrets.baseline` |
| trivy | 0.56.x | Vulnerability scanning | `trivy fs --severity HIGH,CRITICAL .` |
| semgrep | 1.94.x | Static analysis | `semgrep --config=auto --error .` |
| bandit | 1.7.x | Python security | `bandit -r src/ -ll` |
| npm audit | latest | JS dependencies | `npm audit --audit-level=high` |
| gosec | 2.x | Go security | `gosec ./...` |

**Security Checklist**:
- No hardcoded secrets (API keys, passwords, tokens)
- Input validation on all external data
- Proper error handling (no sensitive info in errors)
- Dependency vulnerability scan passed
- Access control enforced
- HTTPS/TLS for network calls

### T - Trackable (TAG Chain Integrity)

**TAG Structure Validation**:
- `@SPEC:ID` in specs (`.moai/specs/SPEC-<ID>/spec.md`)
- `@TEST:ID` in tests (`tests/`)
- `@CODE:ID` in source (`src/`)
- `@DOC:ID` in docs (`docs/`)

**Chain Verification**:
```bash
# Scan all TAGs
rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/

# Detect orphans
rg '@CODE:AUTH-001' -n src/          # CODE exists
rg '@SPEC:AUTH-001' -n .moai/specs/  # SPEC missing → orphan
```

**Traceability Requirements**:
- Every `@CODE:ID` must reference `@SPEC:ID` and `@TEST:ID`
- TAG block includes file paths: `SPEC: <path> | TEST: <path>`
- HISTORY section tracks all changes with dates
- No duplicate TAG IDs across project

## Inputs
- Project configuration (`.moai/config.json`, `CLAUDE.md`)
- Source code (`src/`, `tests/`)
- SPEC documents (`.moai/specs/`)
- CI/CD configuration (`.github/workflows/`)

## Outputs
- TRUST compliance report (pass/fail per principle)
- Coverage metrics with delta from previous run
- Security scan results
- TAG chain validation report
- Quality gate decision (block/allow merge)

## CI/CD Integration

**GitHub Actions Example** (`.github/workflows/trust-validation.yml`):

```yaml
name: TRUST Validation

on: [pull_request, push]

jobs:
  trust-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      # T - Test First (Python example)
      - name: Run tests with coverage
        run: |
          pip install pytest pytest-cov
          pytest --cov=src --cov-report=xml --cov-fail-under=85

      # R - Readable
      - name: Lint code
        run: |
          pip install ruff
          ruff check .

      # U - Unified
      - name: Type check
        run: |
          pip install mypy
          mypy src/ --strict

      # S - Secured
      - name: Security scan
        uses: aquasecurity/trivy-action@master
        with:
          scan-type: 'fs'
          severity: 'HIGH,CRITICAL'
          exit-code: '1'

      - name: Secret detection
        run: |
          pip install detect-secrets
          detect-secrets scan --baseline .secrets.baseline

      # T - Trackable
      - name: TAG validation
        run: |
          # Check for orphaned TAGs
          ./scripts/validate-tags.sh
```

## Quality Gate Enforcement

**Merge Requirements**:
- All 5 TRUST principles must pass
- Test coverage ≥85% (no regression from baseline)
- Zero high/critical security vulnerabilities
- All TAGs have valid chains
- Code review approved

**Failure Actions**:
- Block PR merge
- Post detailed failure report as PR comment
- Suggest remediation steps
- Link to relevant documentation

## Failure Modes
- Missing standard files (`.moai/config.json`, `CLAUDE.md`)
- Insufficient file access permissions
- Conflicting policies requiring coordination
- Tool version mismatches in CI vs local
- Network failures during dependency scans

## Dependencies
- Works synergistically with `moai-foundation-tags` (TAG traceability)
- Requires `moai-foundation-specs` (SPEC validation)
- Integrates with `cc-manager` for session management

## References
- pytest Documentation. "Coverage Reporting." https://docs.pytest.org/en/stable/how-to/coverage.html (accessed 2025-10-22).
- Vitest Documentation. "Coverage Guide." https://vitest.dev/guide/coverage.html (accessed 2025-10-22).
- Go Testing Package. "Coverage Testing." https://go.dev/blog/coverage (accessed 2025-10-22).
- Trivy Documentation. "Vulnerability Scanning." https://trivy.dev/ (accessed 2025-10-22).
- detect-secrets. "Secret Detection Tool." https://github.com/Yelp/detect-secrets (accessed 2025-10-22).
- SemGrep. "Lightweight SAST." https://semgrep.dev/docs/ (accessed 2025-10-22).
- SonarSource. "Quality Gate: Developer's Guide." https://www.sonarsource.com/company/newsroom/white-papers/quality-gate/ (accessed 2025-03-29).
- ISO/IEC 25010. "Systems and software quality models." (accessed 2025-03-29).

## Changelog
- 2025-10-22: v2.0.0 - Added latest tool versions (pytest 8.4.2, Vitest 2.0.5, trivy 0.56.x, etc.), CI/CD automation, SAST tools, quality gate enforcement
- 2025-03-29: v1.0.0 - Foundation skill templates enhanced to align with best practice structures

## Works well with

- `moai-foundation-tags` (TAG traceability)
- `moai-foundation-specs` (SPEC validation)
- `moai-alfred-trust-validation` (Alfred workflow integration)
- `cc-manager` (session management)

## Examples

**Scenario 1: Pre-merge TRUST validation**
```bash
# Run full TRUST check before merging
pytest --cov=src --cov-fail-under=85
ruff check .
mypy src/ --strict
trivy fs --severity HIGH,CRITICAL .
detect-secrets scan
rg '@(SPEC|TEST|CODE):' -n .moai/specs/ tests/ src/
```

**Scenario 2: CI/CD pipeline integration**
```yaml
# GitHub Actions workflow snippet
- name: TRUST Gate
  run: |
    chmod +x scripts/trust-check.sh
    ./scripts/trust-check.sh || exit 1
```

**Scenario 3: Local development pre-commit**
```bash
# .git/hooks/pre-commit
#!/bin/bash
pytest --cov=src --cov-fail-under=85 --quiet || exit 1
ruff check . --quiet || exit 1
mypy src/ --strict --no-error-summary || exit 1
```

## Best Practices
- Run TRUST validation locally before pushing
- Configure IDE to show coverage inline
- Automate quality gates in CI/CD (never manual)
- Track coverage trends over time
- Document reasons for TRUST exceptions (with approval)
- Update tool versions quarterly
- Use pre-commit hooks for fast feedback
- Fail fast on security violations
- Generate coverage reports for review
- Archive TRUST reports per release
