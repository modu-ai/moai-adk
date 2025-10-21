# SPEC Metadata Structure Guide

> **MoAI-ADK SPEC Metadata Standard**
>
> Every SPEC document must follow this structure.

---

## 📋 Metadata Overview

SPEC metadata contains **7 required fields** and **9 optional fields**.

### Full Example

```yaml
---
# Required Fields (7)
id: AUTH-001                    # Unique SPEC ID
version: 0.0.1                  # Semantic version (v0.0.1 = INITIAL, draft start)
status: draft                   # draft|active|completed|deprecated
created: 2025-09-15             # Creation date (YYYY-MM-DD)
updated: 2025-09-15             # Last updated (YYYY-MM-DD; initially same as created)
author: @Goos                   # Author (single GitHub handle)
priority: high                  # low|medium|high|critical

# Optional Fields – Classification/Meta
category: security              # feature|bugfix|refactor|security|docs|perf
labels:                         # Tags for search and grouping
  - authentication
  - jwt

# Optional Fields – Relationships (Dependency Graph)
depends_on:                     # SPECs this one depends on (optional)
  - USER-001
blocks:                         # SPECs blocked by this one (optional)
  - AUTH-002
related_specs:                  # Related SPECs (optional)
  - TOKEN-002
related_issue: "https://github.com/modu-ai/moai-adk/issues/123"

# Optional Fields – Scope/Impact
scope:
  packages:                     # Impacted packages
    - src/core/auth
  files:                        # Key files (optional)
    - auth-service.ts
    - jwt-manager.ts
---
```

---

## Required Fields

### 1. `id` – Unique SPEC Identifier
- **Type**: string
- **Format**: `<DOMAIN>-<NUMBER>`
- **Examples**: `AUTH-001`, `INSTALLER-SEC-001`
- **Rules**:
  - Immutable once assigned
  - Use three digits (001–999)
  - Domain in uppercase; hyphens allowed
  - Directory name: `.moai/specs/SPEC-{ID}/` (e.g., `.moai/specs/SPEC-AUTH-001/`)

### 2. `version` – Semantic Version
- **Type**: string (`MAJOR.MINOR.PATCH`)
- **Default**: `0.0.1` (all SPECs start here, status: draft)
- **Version Lifecycle**:
  - **v0.0.1**: INITIAL – SPEC first draft (status: draft)
  - **v0.0.x**: Draft refinements (increment PATCH when editing the SPEC)
  - **v0.1.0**: TDD implementation complete (status: completed, updated via `/alfred:3-sync`)
  - **v0.1.x**: Bug fixes or doc improvements (PATCH increment)
  - **v0.x.0**: Feature additions or major enhancements (MINOR increment)
  - **v1.0.0**: Stable release (production ready, explicit stakeholder approval required)

### 3. `status` – Progress State
- **Type**: enum
- **Values**:
  - `draft`: Authoring in progress
  - `active`: Implementation underway
  - `completed`: Implementation finished
  - `deprecated`: Planned for retirement

### 4. `created` – Creation Date
- **Type**: date string
- **Format**: `YYYY-MM-DD`
- **Example**: `2025-10-06`

### 5. `updated` – Last Modified Date
- **Type**: date string
- **Format**: `YYYY-MM-DD`
- **Rule**: Update whenever the SPEC content changes.

### 6. `author` – Primary Author
- **Type**: string
- **Format**: `@{GitHub ID}`
- **Example**: `@Goos`
- **Rules**:
  - Single value only (no `authors` array)
  - Prefix the GitHub handle with `@`
  - Additional contributors belong in the HISTORY section

### 7. `priority` – Work Priority
- **Type**: enum
- **Values**:
  - `critical`: Immediate attention (security, severe defects)
  - `high`: Major feature work
  - `medium`: Enhancements
  - `low`: Optimizations or documentation

---

## Optional Fields

### Classification / Meta

#### 8. `category` – Change Type
- **Type**: enum
- **Values**:
  - `feature`: New functionality
  - `bugfix`: Defect resolution
  - `refactor`: Structural improvements
  - `security`: Security enhancements
  - `docs`: Documentation updates
  - `perf`: Performance optimizations

#### 9. `labels` – Classification Tags
- **Type**: array of strings
- **Purpose**: Search, filtering, grouping
- **Example**:
  ```yaml
  labels:
    - installer
    - template
    - security
  ```

### Relationship Fields (Dependency Graph)

#### 10. `depends_on` – Required SPECs
- **Type**: array of strings
- **Meaning**: SPECs that must be completed first
- **Example**:
  ```yaml
  depends_on:
    - USER-001
    - AUTH-001
  ```
- **Use Case**: Determines execution order and parallelization.

#### 11. `blocks` – Blocked SPECs
- **Type**: array of strings
- **Meaning**: SPECs that cannot proceed until this one is resolved
- **Example**:
  ```yaml
  blocks:
    - PAYMENT-003
  ```

#### 12. `related_specs` – Associated SPECs
- **Type**: array of strings
- **Meaning**: Related items without direct dependencies
- **Example**:
  ```yaml
  related_specs:
    - TOKEN-002
    - SESSION-001
  ```

#### 13. `related_issue` – Linked GitHub Issue
- **Type**: string (URL)
- **Format**: Full GitHub issue URL
- **Example**:
  ```yaml
  related_issue: "https://github.com/modu-ai/moai-adk/issues/123"
  ```

### Scope Fields (Impact Analysis)

#### 14. `scope.packages` – Impacted Packages
- **Type**: array of strings
- **Meaning**: Packages or modules touched by the SPEC
- **Example**:
  ```yaml
  scope:
    packages:
      - moai-adk-ts/src/core/installer
      - moai-adk-ts/src/core/git
  ```

#### 15. `scope.files` – Key Files
- **Type**: array of strings
- **Meaning**: Primary files involved (for reference)
- **Example**:
  ```yaml
  scope:
    files:
      - template-processor.ts
      - template-security.ts
  ```

---

## Metadata Validation

### Required Field Checks
```bash
# Verify that every SPEC includes the required fields
rg "^(id|version|status|created|updated|author|priority):" .moai/specs/SPEC-*/spec.md

# Identify SPECs missing the priority field
rg -L "^priority:" .moai/specs/SPEC-*/spec.md
```

### Format Checks
```bash
# Ensure the author field uses @Handle format
rg "^author: @[A-Z]" .moai/specs/SPEC-*/spec.md

# Ensure the version field follows 0.x.y
rg "^version: 0\.\d+\.\d+" .moai/specs/SPEC-*/spec.md
```

---

## Migration Guide

### Updating Existing SPECs

#### 1. Add the `priority` Field
Add it if missing:
```yaml
priority: medium  # or low|high|critical
```

#### 2. Normalize the `author` Field
- `authors: ["@goos"]` → `author: @Goos`
- Convert lowercase handles to the canonical casing.

#### 3. Add Optional Fields (Recommended)
```yaml
category: refactor
labels:
  - code-quality
  - maintenance
```

### Updating config.json for Language Support (v0.4.2+)

**Background**: MoAI-ADK v0.4.2 introduces conversation language selection in `/alfred:0-project`. Existing projects need to add language metadata to `.moai/config.json`.

#### Migration Steps

**For Existing Projects** (before v0.4.2):

Current config.json structure:
```json
{
  "project": {
    "locale": "en",
    "mode": "personal",
    "language": "python"
  }
}
```

**Updated Structure** (v0.4.2+):
```json
{
  "project": {
    "locale": "en",
    "mode": "personal",
    "language": "python",
    "conversation_language": "en",
    "conversation_language_name": "English",
    "codebase_languages": ["python"]
  }
}
```

#### New Fields

| Field | Type | Required | Description | Example |
|-------|------|----------|-------------|---------|
| `conversation_language` | string (ISO 639-1 code) | ✅ Yes | Two-letter language code for Alfred dialogs | `"ko"`, `"en"`, `"ja"`, `"zh"` |
| `conversation_language_name` | string | ✅ Yes | Display name of conversation language | `"Korean"`, `"English"` |
| `codebase_languages` | array of strings | ✅ Yes | List of programming languages detected | `["python"]`, `["typescript", "python"]` |

#### Manual Update Process

1. Open `.moai/config.json`
2. Add the three new fields under `project`:
   ```json
   "conversation_language": "en",
   "conversation_language_name": "English",
   "codebase_languages": ["python"]
   ```
3. Save and commit:
   ```bash
   git add .moai/config.json
   git commit -m "chore: add language metadata to config.json for v0.4.2+"
   ```

#### Automated Update (via `/alfred:0-project`)

Running `/alfred:0-project` on an existing project will:
1. Detect current language settings
2. Add new fields automatically
3. Preserve existing values

**No manual action required if running `/alfred:0-project` after upgrade.**

#### Field Mapping (Legacy → New)

| Old Field | New Field | Migration Rule |
|-----------|-----------|-----------------|
| `locale` | `conversation_language` | Keep as-is (or run `/alfred:0-project` to re-select) |
| (none) | `conversation_language_name` | Auto-populate from locale mapping |
| `language` | `codebase_languages` | Wrap in array: `"python"` → `["python"]` |

#### Backward Compatibility

- ✅ Projects without new fields will continue working
- ⚠️ New language features (multilingual documentation) unavailable without migration
- ✅ `/alfred:0-project` automatically migrates on next run
- ✅ Auto-detection will prefer new fields if present

---

## Design Principles

### 1. DRY (Don't Repeat Yourself)
- ❌ **Remove**: the `reference` field (every SPEC referenced the same master plan)
- ✅ **Instead**: document project-level resources in README.md

### 2. Context-Aware
- Include only the necessary context.
- Use optional fields only when they add value.

### 3. Traceable
- Use `depends_on`, `blocks`, and `related_specs` to map dependencies.
- Automated tooling can detect cyclic references.

### 4. Maintainable
- Every field must be machine-verifiable.
- Maintain consistent formatting for easy parsing.

### 5. Simple First
- Keep complexity low.
- Limit to 7 required + 9 optional fields.
- Expand gradually when justified.

---

**Last Updated**: 2025-10-06  
**Author**: @Alfred
