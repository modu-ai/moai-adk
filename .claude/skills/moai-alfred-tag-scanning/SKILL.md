---
name: moai-alfred-tag-scanning
description: Scans all @TAG markers directly from code and generates TAG inventory (CODE-FIRST principle - no intermediate cache)
version: 0.2.0
author: MoAI Skill Factory
license: MIT
tags:
  - tag
  - tracking
  - code-first
  - spec
---

<!-- @CODE:UPDATE-004:PHASE1 | SPEC: .moai/specs/SPEC-UPDATE-004/spec.md -->

# Alfred TAG Scanning

## What it does

Scans all @TAG markers (SPEC/TEST/CODE/DOC) directly from codebase and generates TAG inventory without intermediate caching (CODE-FIRST principle). Manages TAG system integrity and traceability across the entire project.

## When to use

- "TAG 스캔", "TAG 목록", "TAG 인벤토리"
- "고아 TAG 찾아줘", "TAG 체인 확인", "TAG 무결성 검증"
- Automatically invoked by `/alfred:3-sync`
- When creating new SPEC (duplicate ID check)
- After TDD implementation (TAG chain verification)

## How it works

### CODE-FIRST Principle

**TAG의 진실은 코드 자체에만 존재**합니다. 모든 TAG는 소스 파일에서 실시간으로 추출되며, 중간 캐시나 데이터베이스를 사용하지 않습니다.

**Direct Code Scanning**:
```bash
# Full TAG scan
rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/

# Domain-specific scan
rg '@SPEC:AUTH' -n .moai/specs/

# Scope-limited scan
rg '@CODE:' -n src/
```

### TAG System (4-Core)

```
@SPEC:ID → @TEST:ID → @CODE:ID → @DOC:ID
```

**TAG Structure**:
- `@SPEC:ID`: Requirements specification (EARS format)
- `@TEST:ID`: Test cases (RED phase)
- `@CODE:ID`: Implementation (GREEN + REFACTOR phase)
- `@DOC:ID`: Documentation (Living Document)

**TAG ID Format**:
- Pattern: `<DOMAIN>-<3-DIGIT>` (e.g., `AUTH-001`)
- Domain: UPPERCASE, hyphen allowed (e.g., `INSTALLER-SEC`)
- Number: 3-digit zero-padded (001~999)
- **Immutable**: Once assigned, never changes

**Directory Naming**:
- Format: `.moai/specs/SPEC-{ID}/`
- Examples: `SPEC-AUTH-001/`, `SPEC-UPDATE-004/`
- Compound domains: `SPEC-UPDATE-REFACTOR-001/`

### TAG Integrity Verification

**Chain Completeness**:
```bash
# Verify TAG chain for specific ID
rg '@SPEC:AUTH-001' -n .moai/specs/
rg '@TEST:AUTH-001' -n tests/
rg '@CODE:AUTH-001' -n src/
rg '@DOC:AUTH-001' -n docs/
```

**Orphaned TAG Detection**:
```bash
# CODE exists but SPEC missing
rg '@CODE:AUTH-001' -n src/          # CODE found
rg '@SPEC:AUTH-001' -n .moai/specs/  # SPEC not found → orphaned
```

**Validation Checklist**:
- 4-Core TAG chain completeness (@SPEC → @TEST → @CODE → @DOC)
- No orphaned TAGs (CODE without SPEC)
- No duplicate IDs (unique ID per TAG)
- No broken references (valid TAG links)
- Consistent versioning (SPEC version matches code version)

### TAG Block Template

**SPEC Document** (.moai/specs/SPEC-{ID}/spec.md):
```yaml
---
id: AUTH-001
version: 0.0.1
status: draft
created: 2025-09-15
updated: 2025-09-15
author: @Goos
priority: high
---

# @SPEC:AUTH-001: JWT Authentication System

## HISTORY
### v0.0.1 (2025-09-15)
- **INITIAL**: JWT-based authentication system specification
```

**Source Code** (src/):
```typescript
// @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: tests/auth/service.test.ts
```

**Test Code** (tests/):
```typescript
// @TEST:AUTH-001 | SPEC: SPEC-AUTH-001.md
```

### @CODE Subcategories

Implementation details are annotated within `@CODE:ID`:
- `@CODE:ID:API` - REST API, GraphQL endpoints
- `@CODE:ID:UI` - Components, views, screens
- `@CODE:ID:DATA` - Data models, schemas, types
- `@CODE:ID:DOMAIN` - Business logic, domain rules
- `@CODE:ID:INFRA` - Infrastructure, database, external integrations

### TAG Operations

**Duplicate Prevention**:
```bash
# Before creating new TAG
rg "@SPEC:NEW-ID" -n .moai/specs/    # Must return empty
rg "@CODE:NEW-ID" -n src/            # Must return empty
```

**TAG Reuse (Recommended)**:
```bash
# Search existing TAGs by keyword
rg '@SPEC:AUTH' -n .moai/specs/      # Find AUTH domain TAGs
rg -i 'authentication' -n .moai/specs/  # Case-insensitive keyword search
```

**TAG Statistics**:
```bash
# Count TAGs by category
rg '@SPEC:' -c .moai/specs/
rg '@TEST:' -c tests/
rg '@CODE:' -c src/
rg '@DOC:' -c docs/
```

### Performance

**Scan Speed**:
- Small project (<100 files): <50ms
- Medium project (100-1000 files): <200ms
- Large project (1000+ files): <1s

**Why Fast**:
- Direct `rg` (ripgrep) usage - no intermediate processing
- Minimal regex patterns - simple string matching
- Scope-limited scanning - only relevant directories

## Examples

### Example 1: Full TAG scan
**User**: "TAG 전체 스캔해줘"

**Alfred**:
```bash
rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/
```

**Output**:
```
Found 42 TAGs:
- @SPEC: 12
- @TEST: 10
- @CODE: 15
- @DOC: 5

Orphaned TAGs: 0
Duplicate IDs: 0
Chain integrity: 100%
```

### Example 2: Find orphaned TAGs
**User**: "고아 TAG 찾아줘"

**Alfred**:
```bash
# Find all CODE TAGs
rg '@CODE:' -n src/

# For each CODE TAG, verify SPEC exists
rg '@SPEC:{ID}' -n .moai/specs/
```

**Output**:
```
Orphaned TAGs found: 2

1. @CODE:PAYMENT-005
   - Location: src/payment/processor.ts:12
   - Issue: No corresponding @SPEC:PAYMENT-005
   - Action: Create SPEC or remove CODE TAG

2. @CODE:ANALYTICS-002
   - Location: src/analytics/tracker.ts:45
   - Issue: No corresponding @SPEC:ANALYTICS-002
   - Action: Create SPEC or remove CODE TAG
```

### Example 3: Check duplicate IDs
**User**: "AUTH-001 TAG 중복 확인"

**Alfred**:
```bash
rg 'AUTH-001' -n .moai/specs/ tests/ src/ docs/
```

**Output**:
```
.moai/specs/SPEC-AUTH-001/spec.md:7:# @SPEC:AUTH-001: JWT Authentication
tests/auth/service.test.ts:1:// @TEST:AUTH-001 | SPEC: SPEC-AUTH-001.md
src/auth/jwt-manager.ts:1:// @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md
docs/auth/jwt.md:1:<!-- @DOC:AUTH-001 -->

No duplicates found. TAG chain is complete.
```

### Example 4: Domain-specific TAG search
**User**: "INSTALLER 도메인 TAG 목록"

**Alfred**:
```bash
rg '@SPEC:INSTALLER' -n .moai/specs/
```

**Output**:
```
Found 5 INSTALLER domain TAGs:
- INSTALLER-001: Template installer core
- INSTALLER-002: Version compatibility check
- INSTALLER-003: Security validation
- INSTALLER-004: Rollback mechanism
- INSTALLER-005: Progress reporting
```

## Success Criteria

**TAG Format Accuracy**:
- TAG format errors: 0 violations
- ID format compliance: 100%

**Duplicate Prevention**:
- Duplicate TAG prevention rate: ≥95%
- Pre-creation duplicate checks: 100%

**Chain Integrity**:
- TAG chain completeness: 100%
- Orphaned TAG detection: 100%
- Broken reference detection: 100%

**Performance**:
- Scan speed (small project): <50ms
- Scan speed (medium project): <200ms
- Scan speed (large project): <1s

## Works well with

- moai-alfred-trust-validation (TAG traceability verification)
- moai-alfred-spec-metadata-validation (SPEC ID validation)
- moai-alfred-git-conventional-commits (TAG-based commit messages)

## Files included

- templates/tag-inventory-template.md
- templates/tag-chain-report.md
