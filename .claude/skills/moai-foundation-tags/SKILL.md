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

- "TAG Scan", "TAG List", "TAG Inventory"
- Automatically invoked by `/alfred:3-sync`
- “Find orphan TAG”, “Check TAG chain”

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
User: "Scan the entire TAG"
Claude: (scans all files and generates TAG inventory report)

### Example 2: Find orphaned TAGs
User: "Find Orphan TAG"
Claude: (identifies TAGs without complete chain)
## Works well with

- moai-foundation-trust
- moai-foundation-specs
