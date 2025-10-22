# SPEC Metadata Validation Examples

_Last updated: 2025-10-22_

## Example 1: Valid SPEC Structure

### Complete SPEC File
```markdown
---
id: AUTH-001
version: 0.1.0
status: draft
created: 2025-10-22
updated: 2025-10-22
owner: @john
tags: [authentication, security]
---

# Authentication Service SPEC

@SPEC:AUTH-001

## Overview
JWT-based authentication system for API access.

## Requirements (EARS Format)

### Ubiquitous
- The system shall hash all passwords using bcrypt with cost factor 12.
- The system shall enforce minimum password length of 12 characters.

### Event-Driven
- When login succeeds, the system shall create a JWT token valid for 24 hours.
- When login fails, the system shall log the attempt with IP address.

### State-Driven
- While user session is active, the system shall refresh token every 15 minutes.

## HISTORY

### v0.1.0 (2025-10-22)
- **INITIAL**: Draft authentication SPEC with JWT requirements.
```

### Validation Result
```
✅ SPEC-AUTH-001 validation passed

Checks:
- ✅ YAML frontmatter present
- ✅ Required fields: id, version, status, created
- ✅ Valid semantic version (0.1.0)
- ✅ Status in allowed values: draft
- ✅ HISTORY section present
- ✅ Version matches latest HISTORY entry
- ✅ @SPEC:AUTH-001 TAG present in content
```

---

## Example 2: Invalid SPEC (Missing Fields)

### Invalid SPEC File
```markdown
---
id: USER-001
status: draft
---

# User Management SPEC

Some content here...
```

### Validation Result
```
❌ SPEC-USER-001 validation failed

Errors:
- ❌ Missing required field: version
- ❌ Missing required field: created
- ❌ Missing HISTORY section
- ❌ @SPEC:USER-001 TAG not found in content
```

---

## Example 3: Version Mismatch Detection

### SPEC File
```markdown
---
id: API-001
version: 0.2.0        # ← Frontmatter version
status: active
created: 2025-10-20
updated: 2025-10-22
---

# API Gateway SPEC

## HISTORY

### v0.1.0 (2025-10-20)
- **INITIAL**: Initial API gateway design.

### v0.2.1 (2025-10-22)    # ← HISTORY version (mismatch!)
- **UPDATE**: Add rate limiting requirements.
```

### Validation Result
```
⚠️  SPEC-API-001 validation warning

Issues:
- ⚠️  Version mismatch:
  Frontmatter: 0.2.0
  Latest HISTORY: 0.2.1
  → Update frontmatter to match HISTORY

Recommendation:
Update frontmatter version to 0.2.1 or add HISTORY entry for 0.2.0.
```

---

## Example 4: Automated Validation in CI/CD

### GitHub Actions Workflow
```yaml
name: Validate SPECs
on:
  pull_request:
    paths:
      - '.moai/specs/**/*.md'

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Python
        uses: actions/setup-python@v4
        with:
          python-version: '3.11'

      - name: Install dependencies
        run: |
          pip install pyyaml semantic-version

      - name: Validate SPEC files
        run: |
          python scripts/validate_specs.py .moai/specs/

      - name: Report
        if: failure()
        run: |
          echo "❌ SPEC validation failed"
          echo "See above for details"
          exit 1
```

### Validation Script
```python
#!/usr/bin/env python3
import re
import sys
from pathlib import Path
import yaml
from semantic_version import Version

def validate_spec(spec_path):
    content = spec_path.read_text()

    # Extract YAML frontmatter
    match = re.search(r'^---\n(.*?)\n---', content, re.DOTALL)
    if not match:
        return False, "Missing YAML frontmatter"

    frontmatter = yaml.safe_load(match.group(1))

    # Check required fields
    required = ['id', 'version', 'status', 'created']
    missing = [f for f in required if f not in frontmatter]
    if missing:
        return False, f"Missing fields: {', '.join(missing)}"

    # Validate version format
    try:
        Version(frontmatter['version'])
    except:
        return False, f"Invalid version format: {frontmatter['version']}"

    # Check HISTORY section
    if '## HISTORY' not in content:
        return False, "Missing HISTORY section"

    # Check @SPEC TAG
    spec_tag = f"@SPEC:{frontmatter['id']}"
    if spec_tag not in content:
        return False, f"Missing {spec_tag} in content"

    return True, "Valid"

# Usage
for spec_file in Path('.moai/specs').rglob('*.md'):
    valid, message = validate_spec(spec_file)
    if not valid:
        print(f"❌ {spec_file}: {message}")
        sys.exit(1)
    else:
        print(f"✅ {spec_file}: {message}")
```

---

_For validation rules and schema, see reference.md_
