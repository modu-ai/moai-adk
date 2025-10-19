---
name: moai-foundation-trust
description: Validates TRUST 5-principles (Test 85%+, Readable, Unified, Secured, Trackable)
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - TodoWrite
tier: 0
auto-load: "true"
---

# Foundation: TRUST Validation

## What it does

Validates MoAI-ADK's TRUST 5-principles compliance to ensure code quality, testability, security, and traceability.

## When to use

- "TRUST 원칙 확인", "품질 검증", "코드 품질 체크", "TRUST 5원칙", "품질 게이트"
- "테스트 커버리지", "코드 복잡도", "보안 검증", "추적성 확인"
- "Quality gate", "Code quality check", "TRUST validation"
- Automatically invoked by `/alfred:3-sync`
- Before merging PR or releasing
- During code review process

## How it works

**T - Test First**:
- Checks test coverage ≥85% (pytest, vitest, go test, cargo test, etc.)
- Verifies TDD cycle compliance (RED → GREEN → REFACTOR)

**R - Readable**:
- File ≤300 LOC, Function ≤50 LOC, Parameters ≤5, Complexity ≤10

**U - Unified**:
- SPEC-driven architecture consistency, Clear module boundaries

**S - Secured**:
- Input validation, No hardcoded secrets, Access control

**T - Trackable**:
- TAG chain integrity (@SPEC → @TEST → @CODE → @DOC)

## TRUST Validation Commands

### T - Test First (Coverage ≥85%)

**Python (pytest)**:
```bash
# Run tests with coverage
pytest --cov=src --cov-report=term-missing --cov-fail-under=85

# Generate coverage report
pytest --cov=src --cov-report=html
open htmlcov/index.html
```

**TypeScript (Vitest)**:
```bash
# Run tests with coverage
vitest --coverage --coverage.threshold.lines=85

# Coverage report
vitest --coverage
```

**Go**:
```bash
# Run tests with coverage
go test ./... -cover -coverprofile=coverage.out

# Check coverage threshold
go tool cover -func=coverage.out | grep total | awk '{if ($3+0 < 85) exit 1}'
```

**Rust**:
```bash
# Run tests with coverage (requires tarpaulin)
cargo tarpaulin --out Html --output-dir coverage --fail-under 85
```

### R - Readable (LOC & Complexity Constraints)

**File LOC check (≤300)**:
```bash
# Find files exceeding 300 LOC
find src/ -name "*.py" -o -name "*.ts" -o -name "*.java" | xargs wc -l | awk '$1 > 300 {print $2 ": " $1 " LOC"}'
```

**Function LOC check (≤50)**:
```bash
# Python: Check function length
grep -n "^def " src/**/*.py | while read line; do
  # Extract function and count lines (simplified)
  echo "Check function length manually or use radon"
done

# Use radon for Python
radon cc src/ -a -nb --total-average
```

**Cyclomatic Complexity (≤10)**:
```bash
# Python
radon cc src/ -s -n C  # Show complexity ≥10

# TypeScript/JavaScript
npx eslint src/ --rule 'complexity: ["error", 10]'

# Java
./gradlew checkstyle
```

**Parameter count (≤5)**:
```bash
# Python: Detect functions with >5 parameters
rg "def \w+\([^)]*," src/ | awk -F, 'NF > 5 {print}'
```

### U - Unified (SPEC-driven Architecture)

**Check SPEC references in code**:
```bash
# Verify all code files reference a SPEC
for file in src/**/*.{py,ts,java,go,rs}; do
  if ! grep -q "@CODE:" "$file" 2>/dev/null; then
    echo "No @CODE tag: $file"
  fi
done
```

**Module boundary check**:
```bash
# Detect circular dependencies (Python)
pydeps src/ --show-deps

# TypeScript
madge --circular src/
```

### S - Secured (Security Checks)

**Hardcoded secrets detection**:
```bash
# Find potential secrets
rg -i "(password|secret|api_key|token)\s*=\s*['\"]" src/

# Use truffleHog for deep scan
truffleHog filesystem src/
```

**SQL Injection check**:
```bash
# Python: Find raw SQL
rg "execute\(.*%s" src/

# Look for parameterized queries
rg "execute\(.*\?|execute\(.*\$[0-9]" src/
```

**Dependency vulnerability scan**:
```bash
# Python
uv pip check
safety check

# TypeScript
pnpm audit
npm audit fix

# Rust
cargo audit
```

### T - Trackable (TAG Chain Integrity)

**TAG chain verification**:
```bash
# Check complete TAG chains
rg '@SPEC:([A-Z]+-[0-9]+)' .moai/specs/ -o -r '$1' | while read tag; do
  has_test=$(rg -q "@TEST:$tag" tests/ && echo "✅" || echo "❌")
  has_code=$(rg -q "@CODE:$tag" src/ && echo "✅" || echo "❌")
  has_doc=$(rg -q "@DOC:$tag" docs/ && echo "✅" || echo "⚠️")
  echo "$tag: SPEC=✅ TEST=$has_test CODE=$has_code DOC=$has_doc"
done
```

## TRUST Report Example

```
╔══════════════════════════════════════════════════════════╗
║              TRUST 5-Principles Report                    ║
╠══════════════════════════════════════════════════════════╣
║ T - Test First                                       ✅   ║
║   • Coverage: 92% (>85%)                                 ║
║   • Total tests: 145 passing                             ║
║   • TDD commits: RED(15) → GREEN(15) → REFACTOR(12)      ║
╠══════════════════════════════════════════════════════════╣
║ R - Readable                                         ✅   ║
║   • Files: 42/45 ≤300 LOC (93%)                          ║
║   • Functions: 180/185 ≤50 LOC (97%)                     ║
║   • Complexity: Average 5.2, Max 9                       ║
║   • Parameters: All ≤5                                   ║
╠══════════════════════════════════════════════════════════╣
║ U - Unified                                          ✅   ║
║   • SPEC-driven: 15/15 features (100%)                   ║
║   • Circular deps: 0                                     ║
║   • Module cohesion: High                                ║
╠══════════════════════════════════════════════════════════╣
║ S - Secured                                          ⚠️   ║
║   • No hardcoded secrets: ✅                             ║
║   • Input validation: 12/15 endpoints (80%)              ║
║   • Dependency scan: 2 medium vulnerabilities            ║
║   → Action: Update axios to 1.6.2+                       ║
╠══════════════════════════════════════════════════════════╣
║ T - Trackable                                        ✅   ║
║   • Complete TAG chains: 14/15 (93%)                     ║
║   • Orphaned TAGs: 0                                     ║
║   • Missing DOC: AUTH-005                                ║
╠══════════════════════════════════════════════════════════╣
║ Overall Status: PASS (4/5 ✅, 1/5 ⚠️)                    ║
║ Recommendation: Fix security issues before merge         ║
╚══════════════════════════════════════════════════════════╝
```

## Examples

### Example 1: Full TRUST validation
User: "TRUST 원칙 확인해줘"

Alfred executes all 5 validation checks and generates report

Result: PASS with 1 warning (Security - update dependencies)

### Example 2: Coverage-only check
User: "테스트 커버리지 확인"

Alfred executes:
```bash
pytest --cov=src --cov-report=term-missing
```

Result: 92% coverage (PASS ≥85%)

### Example 3: Code complexity check
User: "코드 복잡도 체크"

Alfred executes:
```bash
radon cc src/ -s -n C
```

Result: 2 functions with complexity >10 (needs refactoring)

### Example 4: Pre-merge validation
User: "/alfred:3-sync"

Alfred automatically runs TRUST validation

Result:
```
TRUST Validation: PASS
✅ Test: 92% coverage
✅ Readable: All constraints met
✅ Unified: SPEC-driven
⚠️ Secured: Update 2 dependencies
✅ Trackable: 14/15 complete chains

PR Ready: Yes (with warnings)
```

## Language-specific Tools

| Language | Test | Coverage | Linter | Complexity | Security |
|----------|------|----------|--------|------------|----------|
| **Python** | pytest | pytest-cov | ruff | radon | safety, bandit |
| **TypeScript** | Vitest | vitest --coverage | Biome | eslint-complexity | npm audit |
| **Java** | JUnit | JaCoCo | Checkstyle | PMD | OWASP Dependency-Check |
| **Go** | go test | go test -cover | golint | gocyclo | gosec |
| **Rust** | cargo test | tarpaulin | clippy | - | cargo audit |
| **Ruby** | RSpec | SimpleCov | RuboCop | flog | bundler-audit |

## Works well with

- moai-foundation-tags (TAG traceability - TRUST T)
- moai-foundation-specs (SPEC validation - TRUST U)
- moai-essentials-review (Code review - TRUST R)
- moai-essentials-refactor (Complexity reduction - TRUST R)
