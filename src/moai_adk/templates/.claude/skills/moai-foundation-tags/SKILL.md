---
name: moai-foundation-tags
description: Scans @TAG markers directly from code and generates inventory (CODE-FIRST)
allowed-tools:
- Read
- Bash
- Write
- Edit
- TodoWrite
---

# Alfred TAG Scanning

## What it does

Scans all @TAG markers (SPEC/TEST/CODE/DOC) directly from codebase and generates TAG inventory without intermediate caching (CODE-FIRST principle).

## When to use

- "TAG 스캔", "TAG 목록", "TAG 인벤토리"
- Automatically invoked by `/alfred:3-sync`
- "고아 TAG 찾아줘", "TAG 체인 확인"

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

## Examples

### Example 1: Full TAG scan
User: "TAG 전체 스캔해줘"
Claude: (scans all files and generates TAG inventory report)

### Example 2: Find orphaned TAGs
User: "고아 TAG 찾아줘"
Claude: (identifies TAGs without complete chain)
## Works well with

- moai-foundation-trust
- moai-foundation-specs
