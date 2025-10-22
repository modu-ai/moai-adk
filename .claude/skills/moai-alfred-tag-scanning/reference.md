# TAG Scanning Reference

_Last updated: 2025-10-22_

## TAG Structure

### TAG Format
```
@<TYPE>:<ID>[:<SUBCATEGORY>]
```

### Components

| Component | Description | Example |
|-----------|-------------|---------|
| TYPE | TAG category | SPEC, CODE, TEST, DOC |
| ID | Domain + number | AUTH-001, USER-042 |
| SUBCATEGORY | Optional classification | API, UI, DATA, DOMAIN, INFRA |

---

## TAG Types

### @SPEC:ID
**Location**: SPEC files (`.moai/specs/`)

**Purpose**: Mark specification documents

**Example**:
```markdown
# Authentication SPEC

@SPEC:AUTH-001

## Requirements
...
```

### @CODE:ID
**Location**: Source code (`src/`)

**Purpose**: Mark implementation

**Example**:
```python
# @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: tests/test_auth.py
def generate_token(user_id):
    pass
```

### @TEST:ID
**Location**: Test files (`tests/`)

**Purpose**: Mark test coverage

**Example**:
```python
# @TEST:AUTH-001 | CODE: src/auth/token.py | SPEC: SPEC-AUTH-001.md
def test_generate_token():
    pass
```

### @DOC:ID
**Location**: Documentation (`docs/`)

**Purpose**: Mark documentation

**Example**:
```markdown
# Authentication Guide

@DOC:AUTH-001

This guide explains JWT authentication.
```

---

## CODE Subcategories

### @CODE:ID:API
REST/GraphQL endpoints, API routes

### @CODE:ID:UI
UI components, views, templates

### @CODE:ID:DATA
Data models, schemas, database entities

### @CODE:ID:DOMAIN
Business logic, domain services

### @CODE:ID:INFRA
Infrastructure, configuration, deployment

---

## TAG Scanning Commands

### Scan All TAGs
```bash
rg '@(SPEC|CODE|TEST|DOC):[A-Z]+-\d{3}' -n
```

### Scan Specific Domain
```bash
rg '@(SPEC|CODE|TEST):AUTH' -n
```

### Count TAG Occurrences
```bash
rg '@CODE:AUTH-001' -c
```

### Find Orphan TAGs (CODE without TEST)
```bash
# List all @CODE TAGs
rg '@CODE:' -o | sort -u > /tmp/code_tags.txt

# List all @TEST TAGs
rg '@TEST:' -o | sort -u > /tmp/test_tags.txt

# Find CODE TAGs without corresponding TEST
comm -23 /tmp/code_tags.txt /tmp/test_tags.txt
```

---

## TAG Inventory Generation

### Complete Inventory Script
```bash
#!/bin/bash
set -euo pipefail

OUTPUT_FILE=".moai/TAG-INVENTORY.md"

echo "# TAG Inventory" > "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"
echo "_Generated: $(date -u +"%Y-%m-%d %H:%M:%S UTC")_" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"

# Extract all unique TAG IDs
TAG_IDS=$(rg '@(SPEC|CODE|TEST|DOC):([A-Z]+-\d{3})' -o --no-filename | \
          sed 's/@[A-Z]*://' | \
          sort -u)

for TAG_ID in $TAG_IDS; do
    echo "## $TAG_ID" >> "$OUTPUT_FILE"
    echo "" >> "$OUTPUT_FILE"

    # SPEC
    SPEC_MATCHES=$(rg "@SPEC:$TAG_ID" -n || true)
    if [ -n "$SPEC_MATCHES" ]; then
        echo "### @SPEC:$TAG_ID" >> "$OUTPUT_FILE"
        echo '```' >> "$OUTPUT_FILE"
        echo "$SPEC_MATCHES" >> "$OUTPUT_FILE"
        echo '```' >> "$OUTPUT_FILE"
        echo "" >> "$OUTPUT_FILE"
    fi

    # CODE
    CODE_MATCHES=$(rg "@CODE:$TAG_ID" -n || true)
    if [ -n "$CODE_MATCHES" ]; then
        echo "### @CODE:$TAG_ID" >> "$OUTPUT_FILE"
        echo '```' >> "$OUTPUT_FILE"
        echo "$CODE_MATCHES" >> "$OUTPUT_FILE"
        echo '```' >> "$OUTPUT_FILE"
        echo "" >> "$OUTPUT_FILE"
    fi

    # TEST
    TEST_MATCHES=$(rg "@TEST:$TAG_ID" -n || true)
    if [ -n "$TEST_MATCHES" ]; then
        echo "### @TEST:$TAG_ID" >> "$OUTPUT_FILE"
        echo '```' >> "$OUTPUT_FILE"
        echo "$TEST_MATCHES" >> "$OUTPUT_FILE"
        echo '```' >> "$OUTPUT_FILE"
        echo "" >> "$OUTPUT_FILE"
    fi

    echo "---" >> "$OUTPUT_FILE"
    echo "" >> "$OUTPUT_FILE"
done

echo "✅ TAG inventory generated: $OUTPUT_FILE"
```

---

## TAG Validation Rules

### Required Relationships
1. Every @CODE must have corresponding @TEST
2. Every @CODE must reference a @SPEC
3. Every @TEST must reference @CODE and @SPEC
4. SPEC files must exist for all referenced TAGs

### Best Practices
- ✅ Place TAG at top of function/class
- ✅ Include references (SPEC, TEST, CODE)
- ✅ Use subcategories for clarity (@CODE:AUTH-001:API)
- ✅ Update TAGs when code moves
- ❌ Don't duplicate TAGs unnecessarily
- ❌ Don't use TAGs in generated code

---

## References

- [MoAI-ADK TAG Guide](/.moai/memory/development-guide.md#tag-lifecycle)
- [Ripgrep Documentation](https://github.com/BurntSushi/ripgrep)

---

_For TAG scanning examples, see examples.md_
