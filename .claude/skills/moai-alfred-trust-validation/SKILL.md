---
name: moai-alfred-trust-validation
description: Validates TRUST 5-principles compliance (Test coverage 85%+, Code constraints, Architecture unity, Security, TAG trackability)
version: 0.2.0
author: MoAI Skill Factory
license: MIT
tags:
  - trust
  - quality
  - validation
  - tdd
---

<!-- @CODE:UPDATE-004:PHASE1 | SPEC: .moai/specs/SPEC-UPDATE-004/spec.md -->

# Alfred TRUST Validation

## What it does

Validates MoAI-ADK's TRUST 5-principles compliance to ensure code quality, testability, security, and traceability. Uses differential scanning (Level 1→2→3) for efficient and accurate quality assurance.

## When to use

- "TRUST 원칙 확인", "품질 검증", "코드 품질 체크"
- "테스트 커버리지 확인", "보안 검사", "코드 표준 준수 확인"
- Automatically invoked by `/alfred:3-sync`
- Before merging PR or releasing
- After major refactoring or feature addition

## How it works

### Differential Scanning System (3-Level)

**Fast Scan First**: Perform lightweight checks first, proceed to deeper analysis only if needed.

**Scanning Strategy**:
- **Level 1 (1-3s)**: File structure, basic configuration
- **Level 2 (5-10s)**: Code quality, test execution
- **Level 3 (20-30s)**: Full analysis, dependency checks

**Early Termination**: If critical issues found at Level 1, report immediately and skip deeper analysis.

### Level 1 - Quick Structure Check (1-3s)

**File Structure**:
```bash
find . -name "*.ts" -o -name "*.js" | wc -l    # Count source files
ls package.json tsconfig.json pyproject.toml  # Check config files
find tests/ -name "*test*" -o -name "*spec*"  # Check test files
```

**What's checked**:
- Basic file structure (source file count)
- Configuration file existence (package.json, tsconfig.json, pyproject.toml)
- Test file presence (test, spec patterns)

### Level 2 - Medium Quality Check (5-10s)

**Test & Quality Scripts**:
```bash
npm run test --silent         # Run tests
npm run lint --silent         # Run linter
npm run test:coverage        # Check coverage
```

**What's checked**:
- Test execution (success rate)
- Linter results (code style compliance)
- Basic coverage (≥85% target)

### Level 3 - Deep Analysis (20-30s)

**Comprehensive TRUST Check**:
```bash
rg '@TAG' -n                 # TAG traceability
rg 'TODO|FIXME' -n           # Incomplete work detection
rg 'import.*from' -n         # Architecture dependency analysis
```

**What's checked**:
- TAG chain integrity (full verification)
- Incomplete work detection (TODO, FIXME patterns)
- Architecture dependency analysis (import structure)

### TRUST 5-Principles Validation

Reference: `@.moai/memory/development-guide.md#TRUST-5원칙`

#### T - Test First (테스트 우선)

**Level 1 Quick Check**:
- `test/` directory exists
- `*test*.ts`, `*spec*.ts` file count
- `package.json` has test script

**Level 2 Medium Check**:
- `npm test` execution and results
- Basic test success rate
- Jest/Vitest config file verification

**Level 3 Deep Check**:
- Test coverage ≥85% precise measurement
- TDD Red-Green-Refactor pattern analysis
- Test independence and determinism verification
- TypeScript type safety in tests

**Language-Specific Tools**:
- **Python**: pytest + coverage.py + mypy
- **TypeScript**: Vitest/Jest + c8/istanbul
- **Java**: JUnit + JaCoCo
- **Go**: go test -cover
- **Rust**: cargo test + tarpaulin
- **Ruby**: RSpec + SimpleCov

#### R - Readable (읽기 쉽게)

**Level 1 Quick Check**:
- File size ≤300 LOC (`wc -l`)
- TypeScript/JavaScript file count
- ESLint/Prettier config file exists

**Level 2 Medium Check**:
- Function size ≤50 LOC
- Parameter count ≤5
- `npm run lint` results

**Level 3 Deep Check**:
- Cyclomatic complexity ≤10 precise calculation
- Readability pattern analysis (naming conventions, comment quality)
- TypeScript strict mode compliance

**Code Constraints**:
```yaml
File: ≤300 LOC
Function: ≤50 LOC
Parameters: ≤5
Complexity: ≤10
```

**Language-Specific Linters**:
- **Python**: ruff (linter + formatter)
- **TypeScript**: Biome or ESLint + Prettier
- **Java**: Checkstyle + PMD
- **Go**: golint + gofmt
- **Rust**: clippy + rustfmt
- **Ruby**: RuboCop

#### U - Unified (통합 설계)

**Level 1 Quick Check**:
- import/export statement basic analysis
- Directory structure consistency
- tsconfig.json path settings

**Level 2 Medium Check**:
- Module dependency directionality
- Layer separation structure
- Interface definition consistency

**Level 3 Deep Check**:
- Circular dependency detection and analysis
- Architecture boundary verification
- Domain model consistency check

**SPEC-Driven Architecture**:
- Each SPEC defines complexity thresholds
- Domain boundaries defined by SPEC (not language conventions)
- Cross-language traceability via @TAG system

#### S - Secured (안전하게)

**Level 1 Quick Check**:
- `.env` file in `.gitignore`
- Basic try-catch block presence
- `package-lock.json` security settings

**Level 2 Medium Check**:
- Input validation logic basic analysis
- Logging system usage patterns
- `npm audit` basic execution

**Level 3 Deep Check**:
- Sensitive data protection pattern verification
- SQL injection prevention pattern check
- Security vulnerability deep analysis

**Security by Design**:
- Security controls implemented during TDD (not after)
- Input validation based on SPEC interface definitions
- Audit logging for SPEC-defined critical operations
- Access control following SPEC permission model

#### T - Trackable (추적 가능)

**Level 1 Quick Check**:
- `package.json` version field
- `CHANGELOG.md` existence
- Git tag basic status

**Level 2 Medium Check**:
- @TAG comment usage patterns
- Commit message convention compliance
- Semantic versioning basic verification

**Level 3 Deep Check**:
- @TAG system complete analysis
- Requirements traceability matrix verification
- Release management system comprehensive evaluation

**SPEC-Code Traceability**:
- All code changes reference SPEC ID via @TAG system
- 3-step workflow tracking:
  - `/alfred:1-spec`: @SPEC:ID tag creation
  - `/alfred:2-build`: @TEST:ID → @CODE:ID TDD implementation
  - `/alfred:3-sync`: @DOC:ID documentation + full TAG verification
- Code-scan based traceability: `rg '@(SPEC|TEST|CODE|DOC):' -n`

### Validation Output Format

**Standard TRUST Validation Report**:
```markdown
🧭 TRUST 5원칙 검증 결과
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
📊 전체 준수율: XX% | 스캔 레벨: X | 소요시간: X초

🎯 원칙별 점수:
┌─────────────────┬──────┬────────┬─────────────────────┐
│ 원칙            │ 점수 │ 상태   │ 핵심 이슈           │
├─────────────────┼──────┼────────┼─────────────────────┤
│ T (Test First)  │ XX%  │ ✅/⚠️/❌ │ [핵심 문제]         │
│ R (Readable)    │ XX%  │ ✅/⚠️/❌ │ [핵심 문제]         │
│ U (Unified)     │ XX%  │ ✅/⚠️/❌ │ [핵심 문제]         │
│ S (Secured)     │ XX%  │ ✅/⚠️/❌ │ [핵심 문제]         │
│ T (Trackable)   │ XX%  │ ✅/⚠️/❌ │ [핵심 문제]         │
└─────────────────┴──────┴────────┴─────────────────────┘

❌ 긴급 수정 필요 (Critical):

1. [T] 테스트 커버리지 부족
   - 현재: XX% (목표: ≥85%)
   - 파일: [test가 없는 파일들]
   - 해결: 누락된 테스트 케이스 작성

2. [S] 보안 취약점 발견
   - 위치: [파일:라인]
   - 내용: [구체적 취약점]
   - 해결: [권장 수정 방법]

⚠️ 개선 권장 (Warning):

1. [R] 함수 크기 초과
   - 현재: XX LOC (권장: ≤50 LOC)
   - 함수: [함수명 in 파일명]
   - 해결: 함수 분해 및 리팩토링

✅ 준수 사항 (Pass):

- [T] TDD 사이클 정상 동작 ✓
- [U] 모듈 구조 일관성 ✓
- [T] 시맨틱 버전 체계 준수 ✓

🎯 개선 우선순위:

1. 🔥 긴급 (24시간 내): [Critical 이슈들]
2. ⚡ 중요 (1주일 내): [Warning 이슈들]
3. 🔧 권장 (2주일 내): [Enhancement 제안들]
```

## Examples

### Example 1: Quality gate check
**User**: "/alfred:3-sync"

**Alfred**:
```bash
# Level 1 (Quick Check)
ls -la .moai/specs/ tests/ src/
wc -l src/**/*.ts

# Level 2 (Medium Check)
npm test --silent
npm run lint --silent

# Level 3 (Deep Check)
rg '@TAG' -n .moai/specs/ tests/ src/
rg 'TODO|FIXME' -n
```

**Output**: (Generates TRUST validation report as shown above)

### Example 2: Manual validation
**User**: "TRUST 원칙 준수도 확인해줘"

**Alfred**: (Executes 3-level scanning and reports compliance status)

**Output**:
```
✅ TRUST 5원칙 전체 준수율: 92%

원칙별 상세:
- T (Test First): 95% ✅ (커버리지 87%)
- R (Readable): 90% ✅ (2개 함수 크기 초과)
- U (Unified): 100% ✅ (모듈 구조 완벽)
- S (Secured): 85% ⚠️ (입력 검증 일부 누락)
- T (Trackable): 90% ✅ (TAG 체인 완전)

권장 조치:
- 입력 검증 로직 추가 (auth/login.ts)
- 2개 함수 리팩토링 (≤50 LOC)
```

### Example 3: Focused validation
**User**: "테스트 커버리지만 확인해줘"

**Alfred**:
```bash
# Level 2 focused check
npm run test:coverage
```

**Output**:
```
Test Coverage Report:
- Statements: 87.5% (target: ≥85%) ✅
- Branches: 82.1% (target: ≥85%) ⚠️
- Functions: 91.3% ✅
- Lines: 86.9% ✅

Missing coverage:
- src/auth/password-reset.ts (12 uncovered branches)
- src/payment/refund.ts (5 uncovered lines)
```

## Performance Metrics

**Validation Quality**:
- Validation accuracy: ≥95%
- False positive rate: ≤5%
- Scan completion time: Level 1(3s), Level 2(10s), Level 3(30s)

**Efficiency**:
- Appropriate scan level selection rate: ≥90%
- Unnecessary deep scan prevention: ≥80%
- Clear improvement direction: 100%

## Works well with

- moai-alfred-tag-scanning (TAG traceability verification)
- moai-alfred-spec-metadata-validation (SPEC compliance check)
- moai-alfred-code-reviewer (code quality analysis)
- moai-alfred-git-conventional-commits (commit message validation)

## Files included

- templates/trust-report-template.md
- templates/trust-detailed-report.md
- templates/trust-summary.md
