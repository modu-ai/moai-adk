---
name: moai-foundation-tags
description: Scans @TAG markers directly from code and generates inventory (CODE-FIRST)
allowed-tools:
  - Read
  - Bash
  - Write
  - Edit
  - TodoWrite
tier: 0
auto-load: "true"
---

# Alfred TAG Scanning

## What it does

Scans all @TAG markers (SPEC/TEST/CODE/DOC) directly from codebase and generates TAG inventory without intermediate caching (CODE-FIRST principle).

## When to use

- "TAG 스캔", "TAG 목록", "TAG 인벤토리", "고아 TAG 찾아줘", "TAG 체인 확인", "추적성 검증"
- "TAG scan", "TAG inventory", "Orphan detection", "TAG chain", "Code-first", "Traceability"
- Automatically invoked by `/alfred:3-sync`
- When verifying TAG integrity

## How it works

**CODE-FIRST Scanning**:
```bash
# Direct code scan without intermediate cache
rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/
```

**TAG Inventory Generation**:
- Lists all TAGs with file locations
- Detects orphaned TAGs (no corresponding SPEC/TEST/CODE)
- Identifies broken links in TAG chain
- Reports duplicate IDs

**TAG Chain Verification**:
- @SPEC → @TEST → @CODE → @DOC connection check
- Ensures traceability across all artifacts

## TAG Inventory Format

```
SPEC TAGs (n):
- ID-001: path/to/spec.md:7
- ID-002: path/to/spec.md:8

TEST TAGs (n):
- ID-001: tests/file.py:12
- ID-002: tests/file.py:25

CODE TAGs (n):
- ID-001: src/file.py:42
- ID-002: src/file.py:89

TAG Chain Completeness:
- ✅ n complete chains
- ⚠️ n incomplete chains
- ❌ n orphaned TAGs
```

## Examples

### Example 1: Full TAG inventory scan
User: "TAG 전체 스캔해줘"
Claude:
```
🔍 CODE-FIRST TAG Scanning...

SPEC TAGs (15):
✅ AUTH-001: .moai/specs/SPEC-AUTH-001/spec.md:7
✅ USER-001: .moai/specs/SPEC-USER-001/spec.md:7
✅ PAYMENT-001: .moai/specs/SPEC-PAYMENT-001/spec.md:7
... (12 more)

TEST TAGs (15):
✅ AUTH-001: tests/auth/test_service.py:1
✅ USER-001: tests/user/test_model.py:1
✅ PAYMENT-001: tests/payment/test_processor.py:1
... (12 more)

CODE TAGs (15):
✅ AUTH-001: src/auth/service.py:1
✅ USER-001: src/user/model.py:1
✅ PAYMENT-001: src/payment/processor.py:1
... (12 more)

DOC TAGs (12):
✅ AUTH-001: docs/api/auth.md:5
✅ USER-001: docs/api/user.md:5
✅ PAYMENT-001: docs/api/payment.md:5
... (9 more)

═══════════════════════════════════════════
TAG Chain Completeness:
✅ 12 complete chains (@SPEC → @TEST → @CODE → @DOC)
⚠️ 3 incomplete chains (missing @DOC)
❌ 0 orphaned TAGs

Overall: 15/15 TAGs linked (100%)
```

### Example 2: Find orphaned TAGs
User: "고아 TAG 찾아줘"
Claude:
```
🔍 Orphan TAG Detection...

❌ Orphaned: @CODE:PAYMENT-003
  Location: src/payment/processor.py:45
  Issue: No matching @SPEC:PAYMENT-003
  Recommendation: Create .moai/specs/SPEC-PAYMENT-003/spec.md

❌ Orphaned: @TEST:BILLING-005
  Location: tests/billing/test_invoice.py:12
  Issue: No matching @CODE:BILLING-005
  Recommendation: Implement src/billing/invoice.py

✅ Verification: 2 orphaned TAGs detected
Next step: Create missing SPEC or remove orphaned TAG
```

### Example 3: Domain-specific TAG scan
User: "AUTH 도메인 TAG만 스캔"
Claude:
```
🔍 Domain-specific TAG Scan: AUTH

SPEC: 3 TAGs
├── AUTH-001: .moai/specs/SPEC-AUTH-001/spec.md:7
├── AUTH-002: .moai/specs/SPEC-AUTH-002/spec.md:7
└── AUTH-003: .moai/specs/SPEC-AUTH-003/spec.md:7

TEST: 3 TAGs
├── AUTH-001: tests/auth/test_service.py:1 ✅
├── AUTH-002: tests/auth/test_jwt.py:1 ✅
└── AUTH-003: tests/auth/test_refresh.py:1 ✅

CODE: 3 TAGs
├── AUTH-001: src/auth/service.py:1 ✅
├── AUTH-002: src/auth/jwt.py:1 ✅
└── AUTH-003: src/auth/refresh.py:1 ✅

DOC: 2 TAGs (⚠️ AUTH-003 missing)
├── AUTH-001: docs/api/auth.md:5 ✅
└── AUTH-002: docs/api/jwt.md:5 ✅

Chain Status:
✅ AUTH-001: Complete (@SPEC→@TEST→@CODE→@DOC)
✅ AUTH-002: Complete (@SPEC→@TEST→@CODE→@DOC)
⚠️ AUTH-003: Incomplete (missing @DOC)

Action required: Add @DOC:AUTH-003 to docs/api/refresh.md
```

### Example 4: Duplicate ID detection
User: "중복 TAG ID 찾아줘"
Claude:
```
🔍 Duplicate TAG ID Detection...

❌ AUTH-001 appears 3 times:
  1. .moai/specs/SPEC-AUTH-001/spec.md:7
  2. .moai/specs/SPEC-AUTH-001-backup/spec.md:7  ← Duplicate!
  3. src/auth/legacy.py:15                        ← Duplicate!

Recommendation:
- Delete backup SPEC file
- Rename legacy.py tag to AUTH-004
- Re-run verification

Result: Found 1 duplicate ID
Next step: Resolve conflicts
```

### Example 5: TAG chain verification workflow
User: "TAG 체인 검증 실행"
Claude:
```
🔍 Complete TAG Chain Verification...

Step 1: Scan SPEC directory
✅ Found 15 SPEC TAGs

Step 2: Scan TEST directory
✅ Found 15 @TEST TAGs (15/15 matched)

Step 3: Scan CODE directory
✅ Found 15 @CODE TAGs (15/15 matched)

Step 4: Scan DOC directory
✅ Found 12 @DOC TAGs (12/15 matched)

Step 5: Verify chain integrity
✅ SPEC→TEST links: 15/15 (100%)
✅ SPEC→CODE links: 15/15 (100%)
✅ SPEC→DOC links: 12/15 (80%)

Step 6: Generate report
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Overall: 14/15 complete chains (93%)

Issues:
⚠️ 3 missing @DOC tags (AUTH-002, USER-003, BILLING-001)

Status: Ready for /alfred:3-sync
Recommendation: Add 3 missing @DOC tags, then re-run
```

## Common TAG Operations

```bash
# Find all TAGs in specific file
rg '@(SPEC|TEST|CODE|DOC):' path/to/file.py

# Find TAGs for specific ID
rg 'AUTH-001' -n

# Find orphaned @CODE (has no matching @SPEC)
rg '@CODE:(\w+-\d+)' -o -r '$1' src/ | while read id; do
  rg "@SPEC:$id" .moai/specs/ -q || echo "Orphan: $id"
done

# Find duplicate TAGs
rg '@(SPEC|TEST|CODE|DOC):(\w+-\d+)' -o src/ | sort | uniq -d

# Full TAG chain check
rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/ | wc -l
```

## Keywords

"TAG 스캔", "TAG 목록", "TAG 인벤토리", "orphan detection", "TAG chain", "code-first", "traceability"

## Reference

- TAG lifecycle: `.moai/memory/spec-metadata.md#TAG-lifecycle`
- SPEC naming: `.moai/specs/SPEC-{ID}/`
- TAG validation: CLAUDE.md#@TAG-시스템

## Works well with

- moai-foundation-trust (TAG chain validation)
- moai-foundation-specs (SPEC metadata)
- `/alfred:3-sync` (TAG synchronization)
