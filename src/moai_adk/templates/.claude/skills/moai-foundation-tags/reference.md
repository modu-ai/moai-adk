# moai-foundation-tags: Reference & Resources

## Official TAG System Documentation

### Primary References
- **MoAI-ADK Configuration**: `.moai/config.json` (tags section)
- **TAG Storage**: Source code scanning (.claude/skills/.moai/specs/)
- **Validation Scripts**: `.moai/scripts/validation/tag_*.py`

## TAG Type Specifications

- **Location**: `.moai/specs/SPEC-XXX/spec.md`
- **Purpose**: Define requirements and acceptance criteria
- **Validation**: Must be approved by SPEC reviewer

- **Location**: `tests/test_*.py`
- **Purpose**: Implement test cases validating @SPEC
- **Validation**: Must link to existing @SPEC

- **Location**: `src/**/*.py`
- **Purpose**: Mark code that implements @SPEC requirements
- **Validation**: Must reference valid @SPEC

- **Location**: `docs/**/*.md`
- **Purpose**: Link documentation to implementation
- **Validation**: Must reference @SPEC or @CODE

## Chain Relationships

### Valid Chain Links (November 2025)

```
    ↓ depends_on
    ↓ validates
    ↓ described_by
```

### Invalid Patterns (Validation Errors)

```
PATTERN 1: Floating @TEST (no @SPEC)

PATTERN 2: Orphan @SPEC (no @TEST/@CODE)

PATTERN 3: Circular reference

PATTERN 4: Version mismatch
```

## Scanning & Detection Commands

### Find TAGs in Codebase
```bash
# All TAGs by type

# Count TAGs

# Find specific TAG

# Find all references to TAG
rg 'AUTH-001' -n
```

### Python TAG Analysis

```python
import re
import subprocess

def scan_tags(tag_type='SPEC'):
    """Scan for specific TAG type"""
    cmd = f'rg @{tag_type}:[A-Z]+-\d+ --no-filename -o'
    result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
    return set(result.stdout.strip().split('\n'))

def find_references(tag):
    """Find all references to a TAG"""
    cmd = f'rg "{tag}" -n'
    result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
    return result.stdout.strip().split('\n')

def validate_chain(spec_tag):
    """Validate complete chain for a SPEC"""
    chains = {
        'spec': False,
        'test': False,
        'code': False,
        'doc': False
    }
    refs = find_references(spec_tag)
    for ref in refs:
        if '@SPEC' in ref: chains['spec'] = True
        if '@TEST' in ref: chains['test'] = True
        if '@CODE' in ref: chains['code'] = True
        if '@DOC' in ref: chains['doc'] = True
    return all(chains.values())  # All parts present
```

## TAG Validation Rules

### Syntax Rules
- Must start with `@`
- Must have TYPE (SPEC, TEST, CODE, DOC, DEPRECATED)
- Must have DOMAIN (alphabetic, max 20 chars)
- Must have NUMBER (001-999)
- Optional: Subtype (numeric suffix)

### Semantic Rules
- TAGs must be unique across codebase
- @SPEC must exist before linking
- @TEST must reference valid @SPEC
- @CODE must reference valid @SPEC
- @DOC must reference @SPEC or @CODE
- No circular dependencies allowed

### Coverage Rules
- Every @SPEC should have @TEST (≥1)
- Every @TEST should validate a @SPEC (≥1)
- Every @CODE should implement a @SPEC (≥1)
- Test coverage ≥85% for all @CODE

## Orphan Detection Patterns

### Query: Find Orphan SPECs

```sql
-- Pseudo SQL for concept
SELECT spec_tag
FROM tags
WHERE type = 'SPEC'
  AND NOT EXISTS (
    SELECT 1 FROM tags
    WHERE type IN ('TEST', 'CODE')
      AND parent_ref = spec_tag
  )
ORDER BY created_date;
```

### Query: Find Floating TESTSs

```sql
SELECT test_tag
FROM tags
WHERE type = 'TEST'
  AND parent_ref NOT IN (
    SELECT tag FROM tags WHERE type = 'SPEC'
  );
```

### Cleanup Decision Tree

```
Found Orphan TAG?
    ├─ Type A: Unused @SPEC
    │  ├─ Feature cancelled?
    │  │  └─ Mark @DEPRECATED
    │  └─ Feature pending?
    │     └─ Link to @TEST/@CODE
    │
    ├─ Type B: Floating @TEST
    │  ├─ Valid test?
    │  │  └─ Create missing @SPEC
    │  └─ Obsolete?
    │     └─ Mark @DEPRECATED
    │
    ├─ Type C: Stray @CODE
    │  ├─ Active feature?
    │  │  └─ Create SPEC→TEST chain
    │  └─ Legacy code?
    │     └─ Document with @DOC
    │
    └─ Type D: Orphan @DOC
       ├─ Source still exists?
       │  └─ Link to @SPEC/@CODE
       └─ Outdated doc?
          └─ Mark @DEPRECATED or delete
```

## Version Compatibility

### Supported Versions

| Component | Version | Release Date | Status |
|-----------|---------|--------------|--------|
| MoAI-ADK | 0.22.5+ | 2025-11-12 | Active |
| TAG System | 4.0.0 | 2025-11-12 | Current |
| Python | 3.12+ | 2023-10-02 | Required |
| Git | 2.40+ | 2023-03-13 | Required |

### Breaking Changes (v3 → v4)

| Feature | v3 Pattern | v4 Pattern | Action |
|---------|-----------|-----------|--------|

## Integration Checklist

### Setup Phase
- [ ] Review TAG system architecture (this Skill Level 1)
- [ ] Understand four-chain model (@SPEC→@TEST→@CODE→@DOC)
- [ ] Review naming conventions and syntax rules
- [ ] Configure `.moai/config.json` TAG settings

### Development Phase
- [ ] Create @SPEC for new feature (Day 1)
- [ ] Add @TEST tags in test files (Days 2-3)
- [ ] Link @CODE tags in implementation (Days 4-5)
- [ ] Add @DOC tags in documentation (Day 6)
- [ ] Run validation script before commit

### Validation Phase
- [ ] Run tag scanner: `rg '@(SPEC|TEST|CODE|DOC):'`
- [ ] Check chains: `.moai/scripts/validation/tag_chain_validator.py`
- [ ] Detect orphans: `.moai/scripts/validation/orphan_detector.py`
- [ ] Verify coverage: Coverage report >= 85%
- [ ] Generate audit report

## Common Issues & Solutions

### Issue: TAG not found during validation
**Cause**: TAG not scanned yet or ripgrep not installed  
**Solution**: 
```bash
# Install ripgrep
brew install ripgrep  # macOS
apt-get install ripgrep  # Linux

# Rescan
```

### Issue: Circular reference detected
**Cause**: @SPEC references @CODE that references @TEST that references @SPEC  
**Solution**: Restructure to avoid loops. Use linear chain only.

### Issue: Coverage < 85% for @SPEC
**Cause**: @SPEC has @TEST but code coverage metrics are low  
**Solution**: Add more test cases or improve test quality. Use mutation testing.

### Issue: Orphan TAGs prevent merge
**Cause**: CI/CD validation fails on orphan detection  
**Solution**: Either link orphan or mark as @DEPRECATED before committing.

## Next Steps

1. **Understand Level 2**: Review practical implementation patterns
2. **Implement Patterns**: Start tagging new features following examples
3. **Run Validation**: Execute tag chain validator in CI/CD
4. **Review Governance**: Implement enterprise audit procedures
5. **Monitor Metrics**: Track TAG health monthly

