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

- "TAG 스캔", "TAG 목록", "TAG 인벤토리", "TAG 체인 확인", "TAG 검증"
- "고아 TAG 찾아줘", "중복 TAG", "끊어진 링크", "TAG 무결성"
- "Traceability check", "TAG chain verification", "Orphaned TAG detection"
- Automatically invoked by `/alfred:3-sync`
- Before merging PR or releasing

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

## TAG Scanning Commands

### Full TAG Scan
```bash
# Scan all TAG types
rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/

# Scan specific TAG type
rg '@SPEC:' -n .moai/specs/
rg '@TEST:' -n tests/
rg '@CODE:' -n src/
rg '@DOC:' -n docs/

# Count TAGs by type
rg '@SPEC:' -c .moai/specs/
rg '@TEST:' -c tests/
rg '@CODE:' -c src/
rg '@DOC:' -c docs/
```

### Orphaned TAG Detection
```bash
# Find CODE TAGs without SPEC
for tag in $(rg '@CODE:([A-Z]+-[0-9]+)' src/ -o -r '$1' | sort -u); do
  if ! rg -q "@SPEC:$tag" .moai/specs/; then
    echo "Orphaned CODE: @CODE:$tag (no SPEC)"
  fi
done

# Find TEST TAGs without SPEC
for tag in $(rg '@TEST:([A-Z]+-[0-9]+)' tests/ -o -r '$1' | sort -u); do
  if ! rg -q "@SPEC:$tag" .moai/specs/; then
    echo "Orphaned TEST: @TEST:$tag (no SPEC)"
  fi
done
```

### Duplicate TAG Detection
```bash
# Find duplicate SPEC IDs
rg '@SPEC:([A-Z]+-[0-9]+)' .moai/specs/ -o -r '$1' | sort | uniq -d

# Find duplicate TEST IDs
rg '@TEST:([A-Z]+-[0-9]+)' tests/ -o -r '$1' | sort | uniq -d
```

### TAG Chain Verification
```bash
# Verify complete chain for a specific ID
TAG_ID="AUTH-001"

echo "Checking TAG chain for $TAG_ID..."
rg "@SPEC:$TAG_ID" .moai/specs/ -l
rg "@TEST:$TAG_ID" tests/ -l
rg "@CODE:$TAG_ID" src/ -l
rg "@DOC:$TAG_ID" docs/ -l
```

### TAG Domain Analysis
```bash
# List all domains
rg '@SPEC:([A-Z]+)-' .moai/specs/ -o -r '$1' | sort -u

# Count TAGs per domain
for domain in $(rg '@SPEC:([A-Z]+)-' .moai/specs/ -o -r '$1' | sort -u); do
  count=$(rg "@SPEC:$domain-" .moai/specs/ -c | awk '{sum+=$1} END {print sum}')
  echo "$domain: $count SPECs"
done
```

## Examples

### Example 1: Full TAG scan
User: "TAG 전체 스캔해줘"

Alfred executes:
```bash
rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/
```

Result:
```
TAG Inventory:
- @SPEC: 15 tags
- @TEST: 15 tags
- @CODE: 15 tags
- @DOC: 12 tags

Complete chains: 12/15 (80%)
Orphaned TAGs: 0
Missing DOC: 3 (AUTH-002, PAYMENT-001, USER-005)
```

### Example 2: Find orphaned TAGs
User: "고아 TAG 찾아줘"

Alfred executes orphan detection script

Result:
```
Orphaned TAGs detected:
- @CODE:PAYMENT-005 (src/payment/refund.ts:42)
  → No @SPEC:PAYMENT-005 found in .moai/specs/

Recommendation:
1. Create SPEC-PAYMENT-005.md
2. Or remove @CODE:PAYMENT-005 if not needed
```

### Example 3: TAG chain verification
User: "AUTH-001 TAG 체인 확인"

Alfred executes:
```bash
rg "@(SPEC|TEST|CODE|DOC):AUTH-001" -n .moai/specs/ tests/ src/ docs/
```

Result:
```
Complete TAG chain for AUTH-001:
✅ @SPEC:AUTH-001 (.moai/specs/SPEC-AUTH-001/spec.md:1)
✅ @TEST:AUTH-001 (tests/auth/service.test.ts:3)
✅ @CODE:AUTH-001 (src/auth/service.ts:1)
✅ @DOC:AUTH-001 (docs/api/auth.md:1)

Status: Complete (4/4)
```

### Example 4: Duplicate TAG detection
User: "중복 TAG 확인"

Alfred executes:
```bash
rg '@SPEC:([A-Z]+-[0-9]+)' .moai/specs/ -o -r '$1' | sort | uniq -d
```

Result:
```
Duplicate SPEC IDs found:
- AUTH-003 (appears 2 times)
  1. .moai/specs/SPEC-AUTH-003/spec.md
  2. .moai/specs/SPEC-AUTH-003-OLD/spec.md

Action required: Remove duplicate
```

### Example 5: Domain-based TAG analysis
User: "도메인별 TAG 분석"

Alfred executes domain analysis script

Result:
```
TAG Distribution by Domain:
- AUTH: 5 SPECs (33%)
- PAYMENT: 4 SPECs (27%)
- USER: 3 SPECs (20%)
- REFACTOR: 2 SPECs (13%)
- UPDATE: 1 SPEC (7%)

Total: 15 SPECs
```

## TAG Structure

**Format**: `@TYPE:DOMAIN-NUMBER`

Examples:
- `@SPEC:AUTH-001` - Authentication SPEC
- `@TEST:AUTH-001` - Authentication test
- `@CODE:AUTH-001` - Authentication implementation
- `@DOC:AUTH-001` - Authentication documentation

**Domain Naming**:
- Single word: `AUTH-001`, `USER-001`
- Hyphen-connected: `INSTALLER-SEC-001`, `UPDATE-REFACTOR-001`
- Avoid 3+ hyphens: Complex domains should be simplified

## Works well with

- moai-foundation-trust (TRUST T: Trackable validation)
- moai-foundation-specs (SPEC metadata integration)
- moai-foundation-git (TAG-based commit messages)
