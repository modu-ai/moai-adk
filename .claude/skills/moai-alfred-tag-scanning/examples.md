# TAG Scanning Examples

_Last updated: 2025-10-22_

## Example 1: TAG Structure

### Complete TAG Block
```python
# @CODE:AUTH-001 | SPEC: SPEC-AUTH-001.md | TEST: tests/test_auth.py
# @CODE:AUTH-001:API | Implements JWT token generation endpoint
class AuthTokenAPI:
    def generate_token(self, user_id: str) -> str:
        """
        Generate JWT token for authenticated user.

        @CODE:AUTH-001:API
        SPEC: SPEC-AUTH-001.md
        TEST: tests/test_auth.py:test_generate_token
        """
        payload = {"user_id": user_id, "exp": datetime.utcnow() + timedelta(hours=24)}
        return jwt.encode(payload, SECRET_KEY, algorithm="HS256")
```

### TAG Components
- **Primary TAG**: `@CODE:AUTH-001`
- **Subcategory**: `@CODE:AUTH-001:API`
- **References**: SPEC and TEST file paths
- **Purpose**: Brief description

---

## Example 2: TAG Scanning Script

### Python Implementation
```python
#!/usr/bin/env python3
import re
from pathlib import Path
from collections import defaultdict

TAG_PATTERN = r'@(SPEC|CODE|TEST|DOC):([A-Z]+-\d{3})(?::([A-Z]+))?'

def scan_tags(directory):
    """
    Scan directory for @TAG markers and generate inventory.
    """
    inventory = defaultdict(list)

    for file_path in Path(directory).rglob('*'):
        if file_path.is_file() and not file_path.name.startswith('.'):
            content = file_path.read_text(errors='ignore')
            matches = re.finditer(TAG_PATTERN, content)

            for match in matches:
                tag_type = match.group(1)  # SPEC, CODE, TEST, DOC
                tag_id = match.group(2)    # AUTH-001
                subcategory = match.group(3) or ''  # API, UI, etc.

                tag_full = f"@{tag_type}:{tag_id}"
                if subcategory:
                    tag_full += f":{subcategory}"

                inventory[tag_id].append({
                    'tag': tag_full,
                    'type': tag_type,
                    'subcategory': subcategory,
                    'file': str(file_path),
                    'line': content[:match.start()].count('\n') + 1
                })

    return inventory

# Usage
inventory = scan_tags('.')
for tag_id, occurrences in sorted(inventory.items()):
    print(f"\n{tag_id}:")
    for occ in occurrences:
        print(f"  {occ['tag']} in {occ['file']}:{occ['line']}")
```

### Output Example
```
AUTH-001:
  @SPEC:AUTH-001 in .moai/specs/SPEC-AUTH-001/spec.md:5
  @CODE:AUTH-001 in src/auth/token.py:12
  @CODE:AUTH-001:API in src/auth/api.py:25
  @TEST:AUTH-001 in tests/test_auth.py:8
  @TEST:AUTH-001 in tests/test_token.py:15

USER-002:
  @SPEC:USER-002 in .moai/specs/SPEC-USER-002/spec.md:5
  @CODE:USER-002 in src/user/profile.py:18
  @TEST:USER-002 in tests/test_profile.py:12
```

---

## Example 3: Orphan TAG Detection

### Scenario
Finding TAGs with missing links (e.g., CODE exists but no TEST).

### Detection Script
```python
def detect_orphans(inventory):
    """
    Detect TAGs with incomplete coverage.
    """
    orphans = []

    for tag_id, occurrences in inventory.items():
        types_present = {occ['type'] for occ in occurrences}

        # Required: SPEC must exist
        if 'SPEC' not in types_present:
            orphans.append(f"{tag_id}: Missing @SPEC")

        # If CODE exists, TEST should exist
        if 'CODE' in types_present and 'TEST' not in types_present:
            orphans.append(f"{tag_id}: @CODE exists but no @TEST")

        # If TEST exists, CODE should exist
        if 'TEST' in types_present and 'CODE' not in types_present:
            orphans.append(f"{tag_id}: @TEST exists but no @CODE")

    return orphans

# Usage
orphans = detect_orphans(inventory)
if orphans:
    print("⚠️  Orphan TAGs detected:")
    for orphan in orphans:
        print(f"  - {orphan}")
else:
    print("✅ No orphan TAGs")
```

### Output
```
⚠️  Orphan TAGs detected:
  - AUTH-003: @CODE exists but no @TEST
  - USER-005: Missing @SPEC
  - API-007: @TEST exists but no @CODE
```

---

## Example 4: TAG Chain Verification

### Scenario
Verify that all TAG references point to existing files.

### Verification Script
```python
def verify_tag_chain(inventory):
    """
    Verify TAG chain integrity (SPEC → CODE → TEST).
    """
    issues = []

    for tag_id, occurrences in inventory.items():
        # Group by type
        by_type = defaultdict(list)
        for occ in occurrences:
            by_type[occ['type']].append(occ)

        # Check SPEC references CODE/TEST files
        for spec_occ in by_type.get('SPEC', []):
            spec_file = Path(spec_occ['file'])
            content = spec_file.read_text()

            # Extract referenced TEST file
            test_match = re.search(r'TEST:\s*([^\s\)]+)', content)
            if test_match:
                test_file = Path(test_match.group(1))
                if not test_file.exists():
                    issues.append(
                        f"{tag_id}: SPEC references non-existent TEST: {test_file}"
                    )

            # Extract referenced CODE file
            code_match = re.search(r'CODE:\s*([^\s\)]+)', content)
            if code_match:
                code_file = Path(code_match.group(1))
                if not code_file.exists():
                    issues.append(
                        f"{tag_id}: SPEC references non-existent CODE: {code_file}"
                    )

    return issues
```

---

_For TAG patterns and conventions, see reference.md_
