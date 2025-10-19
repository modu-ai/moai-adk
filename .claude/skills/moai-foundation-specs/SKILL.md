---
name: moai-foundation-specs
description: Validates SPEC YAML frontmatter (7 required fields) and HISTORY section
allowed-tools:
  - Read
  - Bash
  - Write
  - Edit
  - TodoWrite
---

# Alfred SPEC Metadata Validation

## What it does

Validates SPEC document structure including YAML frontmatter (7 required fields) and HISTORY section compliance.

## When to use

- "SPEC 검증", "메타데이터 확인", "SPEC 구조 체크"
- Automatically invoked by `/alfred:1-plan`
- Before creating SPEC document

## How it works

**YAML Frontmatter Validation (7 required fields)**:
- `id`: SPEC ID (e.g., AUTH-001)
- `version`: Semantic Version (e.g., 0.0.1)
- `status`: draft|active|completed|deprecated
- `created`: YYYY-MM-DD format
- `updated`: YYYY-MM-DD format
- `author`: @{GitHub ID} format
- `priority`: low|medium|high|critical

**HISTORY Section Validation**:
- Checks existence of HISTORY section
- Verifies version history (INITIAL/ADDED/CHANGED/FIXED tags)
- Validates author and date consistency

**Format Validation**:
```bash
# Check required fields
rg "^(id|version|status|created|updated|author|priority):" .moai/specs/SPEC-*/spec.md

# Verify HISTORY section
rg "^## HISTORY" .moai/specs/SPEC-*/spec.md
```

## Examples

### Example 1: Validate single SPEC
User: "SPEC-AUTH-001 메타데이터 확인해줘"
Claude:
```
SPEC Metadata Validation Report

✅ PASS: id (AUTH-001)
✅ PASS: version (0.0.1)
✅ PASS: status (draft)
✅ PASS: created (2025-10-06)
✅ PASS: updated (2025-10-06)
✅ PASS: author (@Goos)
✅ PASS: priority (high)
✅ PASS: HISTORY section (3 entries)

Overall: 7/7 required fields present
```

### Example 2: Detect missing fields
User: "SPEC-USER-001 메타데이터 검증"
Claude:
```
❌ FAIL: Missing required fields
  - ❌ priority: NOT FOUND (expected: low|medium|high|critical)

⚠️ WARNING: author format incorrect
  - Expected: @{GitHub_ID} (e.g., @Goos)
  - Found: goos (missing @ prefix)

→ Fix recommendations:
  1. Add 'priority: medium' to YAML frontmatter
  2. Change 'author: goos' → 'author: @Goos'
  3. Re-run validation
```

### Example 3: Batch validation with summary
User: "모든 SPEC 메타데이터 검증"
Claude:
```
Batch Metadata Validation: 15 SPEC documents scanned

✅ PASS: 12 specs (80%)
⚠️ WARNING: 2 specs (13%)
  - SPEC-PAYMENT-001: priority missing
  - SPEC-BILLING-002: HISTORY format incorrect

❌ FAIL: 1 spec (7%)
  - SPEC-LEGACY-001: author format invalid

Summary:
- All required fields: 13/15 (87%)
- Valid HISTORY sections: 14/15 (93%)
- Overall compliance: 87%

Next steps:
1. Fix 2 missing priorities
2. Update 1 HISTORY format
3. Correct 1 author format
→ Re-run validation to confirm
```

### Example 4: HISTORY validation
User: "SPEC-AUTH-001의 HISTORY 섹션 검증"
Claude:
```
HISTORY Section Validation

✅ PASS: HISTORY section exists
✅ PASS: Version entries present:
  - v0.0.1 (2025-10-06): INITIAL ✅
  - v0.0.2 (2025-10-10): ADDED - Token refresh logic ✅
  - v0.1.0 (2025-10-15): CHANGED - API response format ✅

✅ PASS: Chronological order maintained
✅ PASS: Author tags present in all entries
✅ PASS: Valid change types (INITIAL, ADDED, CHANGED)

Result: HISTORY section fully compliant
```

## Common Validation Errors

- **Missing required fields**: Add all 7 required fields to YAML frontmatter
  - Minimum fields: id, version, status, created, updated, author, priority

- **Invalid date format**: Use YYYY-MM-DD (e.g., 2025-10-06)
  - ❌ Wrong: 10/06/2025, Oct 6 2025
  - ✅ Correct: 2025-10-06

- **Invalid status value**: Must be one of: draft, active, completed, deprecated
  - ❌ Wrong: wip, pending, in-progress
  - ✅ Correct: draft

- **Author format**: Use @{GitHub_ID}
  - ❌ Wrong: Goos, goos, user123
  - ✅ Correct: @Goos

- **Missing HISTORY section**: Add `## HISTORY` section with version entries
  - Must include at least v0.0.1 (INITIAL)

## Reference

- Detailed metadata guide: `.moai/memory/spec-metadata.md`
- SPEC directory structure: `.moai/specs/SPEC-{ID}/spec.md`
- Related: moai-foundation-ears (requirement writing)

## Works well with

- moai-foundation-ears (SPEC authoring)
- moai-foundation-tags (TAG tracking)
- moai-foundation-trust (TRUST validation)
