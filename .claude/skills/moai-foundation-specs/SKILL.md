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

### Example 1: Validate SPEC structure
User: "SPEC-AUTH-001 메타데이터 확인해줘"
Claude: (validates YAML frontmatter and HISTORY section, reports issues)

### Example 2: Batch validation
User: "모든 SPEC 메타데이터 검증"
Claude: (validates all SPEC documents and generates report)
## Works well with

- moai-foundation-ears
- moai-foundation-tags
