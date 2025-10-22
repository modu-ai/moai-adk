# SPEC Metadata Validation Reference

_Last updated: 2025-10-22_

## Required YAML Frontmatter Fields

### Minimal Required Fields (7)
```yaml
---
id: DOMAIN-###        # Unique identifier (e.g., AUTH-001)
version: 0.0.1        # Semantic version
status: draft         # draft | active | completed | archived
created: YYYY-MM-DD   # ISO 8601 date
updated: YYYY-MM-DD   # ISO 8601 date (optional but recommended)
owner: @username      # Responsible person
tags: [tag1, tag2]    # Categorization tags
---
```

### Field Specifications

#### `id` (Required)
- **Format**: `<DOMAIN>-<###>`
- **Example**: `AUTH-001`, `USER-042`, `API-123`
- **Rules**:
  - Domain: UPPERCASE letters only
  - Number: Zero-padded 3 digits
  - Must be unique across all SPECs

#### `version` (Required)
- **Format**: Semantic Versioning (SemVer)
- **Example**: `0.1.0`, `1.2.3`
- **Rules**:
  - MAJOR.MINOR.PATCH
  - Start with `0.0.1` for initial draft
  - Must match latest HISTORY entry version

#### `status` (Required)
- **Allowed values**:
  - `draft`: Work in progress
  - `active`: Being implemented
  - `completed`: Implementation finished
  - `archived`: Deprecated or superseded

#### `created` (Required)
- **Format**: ISO 8601 date (`YYYY-MM-DD`)
- **Example**: `2025-10-22`
- **Rules**: Cannot be in the future

#### `updated` (Optional but recommended)
- **Format**: ISO 8601 date (`YYYY-MM-DD`)
- **Rules**: Must be >= `created` date

#### `owner` (Required)
- **Format**: `@username` or email
- **Example**: `@john`, `john@example.com`

#### `tags` (Required)
- **Format**: YAML list
- **Example**: `[authentication, security, api]`
- **Rules**: At least one tag required

---

## HISTORY Section Requirements

### Format
```markdown
## HISTORY

### v0.0.1 (YYYY-MM-DD)
- **INITIAL**: Brief description of initial version.

### v0.1.0 (YYYY-MM-DD)
- **UPDATE**: Description of changes.
- **ADDED**: New requirements or features.

### v0.2.0 (YYYY-MM-DD)
- **BREAKING**: Breaking changes description.
- **FIXED**: Bug fixes or corrections.
```

### HISTORY Entry Types

| Type | Usage | Example |
|------|-------|---------|
| INITIAL | First version | First draft of SPEC |
| UPDATE | Non-breaking changes | Added clarifications |
| ADDED | New requirements | New authentication method |
| BREAKING | Breaking changes | Changed API contract |
| FIXED | Corrections | Fixed typo in requirement |
| DEPRECATED | Deprecation notice | Old method marked for removal |

### Validation Rules
- HISTORY must be present
- Latest entry version must match frontmatter version
- Entries must be in descending order (newest first)
- Each entry must have date in format (YYYY-MM-DD)

---

## @SPEC TAG Requirements

### Format
```markdown
@SPEC:<ID>
```

### Rules
- Must appear at least once in SPEC body
- ID must match frontmatter `id` field
- Typically placed right after title

### Example
```markdown
---
id: AUTH-001
---

# Authentication Service SPEC

@SPEC:AUTH-001

## Overview
...
```

---

## Validation Script

### Python Implementation
```python
import re
import yaml
from pathlib import Path
from datetime import datetime
from semantic_version import Version

class SPECValidator:
    REQUIRED_FIELDS = ['id', 'version', 'status', 'created', 'owner', 'tags']
    ALLOWED_STATUS = ['draft', 'active', 'completed', 'archived']

    def __init__(self, spec_path):
        self.path = Path(spec_path)
        self.content = self.path.read_text()
        self.errors = []
        self.warnings = []

    def validate(self):
        self.validate_frontmatter()
        self.validate_history()
        self.validate_tags()
        return len(self.errors) == 0

    def validate_frontmatter(self):
        match = re.search(r'^---\n(.*?)\n---', self.content, re.DOTALL)
        if not match:
            self.errors.append("Missing YAML frontmatter")
            return

        try:
            self.frontmatter = yaml.safe_load(match.group(1))
        except yaml.YAMLError as e:
            self.errors.append(f"Invalid YAML: {e}")
            return

        # Check required fields
        for field in self.REQUIRED_FIELDS:
            if field not in self.frontmatter:
                self.errors.append(f"Missing required field: {field}")

        # Validate ID format
        if 'id' in self.frontmatter:
            if not re.match(r'^[A-Z]+-\d{3}$', self.frontmatter['id']):
                self.errors.append(f"Invalid ID format: {self.frontmatter['id']}")

        # Validate version
        if 'version' in self.frontmatter:
            try:
                Version(self.frontmatter['version'])
            except ValueError:
                self.errors.append(f"Invalid version: {self.frontmatter['version']}")

        # Validate status
        if 'status' in self.frontmatter:
            if self.frontmatter['status'] not in self.ALLOWED_STATUS:
                self.errors.append(f"Invalid status: {self.frontmatter['status']}")

        # Validate dates
        if 'created' in self.frontmatter:
            try:
                created = datetime.fromisoformat(self.frontmatter['created'])
                if created > datetime.now():
                    self.errors.append("Created date is in the future")
            except ValueError:
                self.errors.append(f"Invalid created date: {self.frontmatter['created']}")

    def validate_history(self):
        if '## HISTORY' not in self.content:
            self.errors.append("Missing HISTORY section")
            return

        # Extract version from latest HISTORY entry
        history_match = re.search(r'### v([\d.]+) \(', self.content)
        if history_match:
            history_version = history_match.group(1)
            fm_version = self.frontmatter.get('version', '')
            if history_version != fm_version:
                self.warnings.append(
                    f"Version mismatch: frontmatter={fm_version}, history={history_version}"
                )

    def validate_tags(self):
        if 'id' in self.frontmatter:
            spec_tag = f"@SPEC:{self.frontmatter['id']}"
            if spec_tag not in self.content:
                self.errors.append(f"Missing {spec_tag} in content")

    def report(self):
        if self.errors:
            print(f"❌ {self.path}")
            for error in self.errors:
                print(f"   Error: {error}")
        if self.warnings:
            for warning in self.warnings:
                print(f"   Warning: {warning}")
        if not self.errors and not self.warnings:
            print(f"✅ {self.path}")
```

---

## References

- [YAML Specification](https://yaml.org/spec/1.2.2/)
- [Semantic Versioning](https://semver.org/)
- [ISO 8601 Date Format](https://en.wikipedia.org/wiki/ISO_8601)
- [MoAI-ADK SPEC Guide](/.moai/memory/development-guide.md)

---

_For validation examples, see examples.md_
