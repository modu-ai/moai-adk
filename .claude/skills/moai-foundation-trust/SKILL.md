---
name: moai-foundation-trust
description: Validates TRUST 5-principles (Test 85%+, Readable, Unified, Secured, Trackable)
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - TodoWrite
---

# Foundation: TRUST Validation

## What it does

Validates MoAI-ADK's TRUST 5-principles compliance to ensure code quality, testability, security, and traceability.

## When to use

- "TRUST 원칙 확인", "품질 검증", "코드 품질 체크"
- Automatically invoked by `/alfred:3-sync`
- Before merging PR or releasing

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

## How to Verify TRUST Principles

**T - Test First**:
```bash
# Python: pytest coverage
pytest --cov=src --cov-report=term-missing
# Target: ≥85% coverage

# TypeScript: vitest coverage
vitest run --coverage
# Target: ≥85% coverage

# Go: go test coverage
go test -cover ./...
# Target: ≥85% coverage

# Rust: cargo tarpaulin
cargo tarpaulin --out Html
# Target: ≥85% coverage
```

**R - Readable**:
```bash
# Check file LOC
wc -l src/**/*.py | awk '$1 > 300 {print "❌ " $0}'

# Check function complexity (Python)
radon cc src/ -s -a

# Check parameters count (TypeScript)
rg "(\w+)\(([^)]{100,})" src/

# ESLint check
npm run lint
```

**U - Unified**:
```bash
# Check circular dependencies
dpdm src/index.ts --circular

# Verify SPEC-driven architecture
ls -la .moai/specs/SPEC-*/
```

**S - Secured**:
```bash
# Check for hardcoded secrets
rg "(password|api_key|secret)\s*=\s*['\"]" src/

# Check input validation
rg "request\.(get|post)" src/ -A 5 | rg "validate"

# SAST scanning (optional)
semgrep --config=p/security-audit src/
```

**T - Trackable**:
```bash
# TAG chain integrity
rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/

# Orphaned TAGs
rg '@CODE:AUTH-001' src/ && rg '@SPEC:AUTH-001' .moai/specs/ || echo "Orphan detected"
```

## Validation Result Format

```markdown
✅ PASS: [항목] ([상세])
⚠️ WARNING: [항목] ([상세])
❌ FAIL: [항목] ([상세])
→ Fix: [해결 방법]
```

## Examples

### Example 1: Full TRUST validation
User: "TRUST 원칙 확인해줘"
Claude:
```
🔍 TRUST 5-Principles Validation Report

T - Test First:
✅ PASS: Coverage 87% (target: 85%+)
✅ PASS: TDD commits detected (RED → GREEN → REFACTOR)
   - 15 RED commits, 15 GREEN commits, 8 REFACTOR commits

R - Readable:
✅ PASS: Max file LOC: 287 (limit: 300)
⚠️ WARNING: 1 function exceeds 50 LOC
   - calculate_total() at src/billing.py:142 (65 LOC)
✅ PASS: Max parameters: 4 (limit: 5)
✅ PASS: Max complexity: 8 (limit: 10)

U - Unified:
✅ PASS: All modules follow SPEC architecture
✅ PASS: No circular dependencies detected

S - Secured:
❌ FAIL: Hardcoded secret found: src/config.py:23
   - api_key = "sk_test_xxxxx"
✅ PASS: Input validation: 45/45 endpoints validated

T - Trackable:
✅ PASS: TAG chain: @SPEC:15 → @TEST:15 → @CODE:15 → @DOC:12
⚠️ WARNING: 3 @DOC tags missing

═══════════════════════════════════════════════════════════
Overall: 4 PASS, 2 WARNING, 1 FAIL
Status: 🔴 BLOCKED (Fix hardcoded secret before merge)

Recommended actions:
1. Remove hardcoded secret from src/config.py:23
   → Use environment variables instead
2. Reduce calculate_total() to ≤50 LOC
3. Add 3 missing @DOC tags
4. Re-run validation to confirm
```

### Example 2: Language-specific TRUST validation
User: "TypeScript 프로젝트 TRUST 검증"
Claude:
```
Language: TypeScript
Version: 5.x (tsconfig: strict=true)

T - Test First:
✅ vitest: 92% coverage (src/: 95%, tests/: 88%)
✅ TDD commits: 24 detected
   npm run test:coverage
   Result: 92% PASS

R - Readable:
✅ Max file: 298 LOC (src/services/auth.ts)
✅ Max function: 48 LOC (authenticate())
✅ Biome linter: 0 issues
   npx biome check src/
   Result: All files OK

U - Unified:
✅ Strict mode enabled
✅ No implicit any
✅ Dependencies: 23 direct, 145 transitive

S - Secured:
✅ No hardcoded secrets
✅ All inputs validated (express.json middleware)
✅ CORS configured restrictively

T - Trackable:
✅ TAG chain: 8/8 complete
   @SPEC:8 @TEST:8 @CODE:8 @DOC:8

Result: ✅ ALL PASS - Ready for production
```

### Example 3: Pre-merge validation (BLOCKED)
User: "/alfred:3-sync"
Claude:
```
🔄 TRUST Validation: Pre-merge quality gate

Scanning: src/ (45 files, 12,450 LOC)
         tests/ (42 files, 8,920 LOC)
         docs/ (15 files)

Results:
✅ PASS: Test coverage: 89% (target: 85%+)
✅ PASS: All functions ≤50 LOC
⚠️ WARNING: 1 potential secret: line 234 in config.py
   - Contains: api_key variable
   - Suggestion: Move to .env file

Status: 🟡 CONDITIONAL PASS

Action required:
→ Move api_key to environment variable
→ Then re-run validation

After fix:
- Run: python scripts/migrate_secrets.py
- Verify: TRUST validation passes
- Proceed with: PR merge
```

### Example 4: TRUST metric progression
User: "TRUST 원칙 진행도 확인"
Claude:
```
TRUST Metrics Over Time

Week 1:
T: 72% coverage → ⚠️ Need improvement
R: 5 files over 300 LOC → ⚠️ Refactor needed
U: 2 circular dependencies → ⚠️ Architecture review
S: 3 hardcoded secrets → ❌ Security issue
T: TAG chain 60% complete → ⚠️ Documentation needed

Week 2 (Current):
T: 89% coverage → ✅ PASS
R: 0 files over 300 LOC → ✅ PASS
U: 0 circular dependencies → ✅ PASS
S: 0 hardcoded secrets → ✅ PASS
T: TAG chain 100% complete → ✅ PASS

Overall Progress: 🟢 Ready for production
```

## Common TRUST Violations

- ❌ **Low test coverage** (< 85%)
  → Add unit tests targeting untested branches
  → Run: `pytest --cov=src --cov-report=term-missing`

- ❌ **Function too long** (> 50 LOC)
  → Extract smaller functions using Extract Method
  → Each function should have single responsibility

- ❌ **Hardcoded secrets**
  → Move to `.env` file or environment variables
  → Never commit credentials

- ❌ **Circular dependencies**
  → Restructure imports to follow dependency hierarchy
  → Use dependency injection if needed

- ❌ **Orphaned TAGs**
  → Every @CODE should have matching @SPEC
  → Every @TEST should reference @SPEC

## Keywords

"TRUST 원칙", "품질 검증", "코드 품질 체크", "test coverage", "code quality", "compliance validation", "production ready"

## Reference

- TRUST 5-principles detailed guide: `.moai/memory/development-guide.md#TRUST-5원칙`
- Code quality standards: CLAUDE.md#코드-제약
- TDD workflow: `/alfred:3-sync` (runs TRUST validation)

## Works well with

- moai-foundation-tags (TAG traceability validation)
- moai-foundation-specs (SPEC-driven architecture)
- moai-essentials-review (code quality review)
- moai-lang-* (language-specific tools)
