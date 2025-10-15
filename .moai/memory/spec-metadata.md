# SPEC Metadata Structure Guide

> **MoAI-ADK SPEC Metadata Standard**
>
> Every SPEC document MUST follow this metadata structure.

---

## 📋 Metadata Structure Overview

SPEC metadata consists of **7 required fields** and **9 optional fields**.

### Full Structure Example

```yaml
---
# Required fields (7)
id: AUTH-001
version: 0.0.1
status: draft
created: 2025-09-15
updated: 2025-09-15
author: @Goos
priority: high

# Optional fields - classification/meta
category: security
labels:
  - authentication
  - jwt

# Optional fields - relationships (dependency graph)
depends_on:
  - USER-001
blocks:
  - AUTH-002
related_specs:
  - TOKEN-002
related_issue: "https://github.com/modu-ai/moai-adk/issues/123"

# Optional fields - scope (impact analysis)
scope:
  packages:
    - src/core/auth
  files:
    - auth-service.ts
    - jwt-manager.ts
---
```

---

## Required Fields

### 1. `id` - Unique SPEC ID
- **Type**: string
- **Format**: `<DOMAIN>-<NUMBER>`
- **Examples**: `AUTH-001`, `INSTALLER-SEC-001`
- **Rules**:
  - Permanent identifier (cannot change once assigned)
  - Use three digits (001-999)
  - Domain is uppercase; hyphens allowed
  - Directory name: `.moai/specs/SPEC-{ID}/` (e.g., `.moai/specs/SPEC-AUTH-001/`)

### 2. `version` - Version
- **Type**: string (Semantic Version)
- **Format**: `MAJOR.MINOR.PATCH`
- **Default**: `0.0.1` (starting point for all SPECs, status: draft)
- **Version scheme**:
  - **v0.0.1**: INITIAL - first draft of the SPEC (status: draft)
  - **v0.0.x**: Draft revisions/improvements (increment patch for document updates)
  - **v0.1.0**: TDD implementation complete (status: completed, updated manually or via `/alfred:3-sync`)
  - **v0.1.x**: Bug fixes, documentation improvements (patch version)
  - **v0.x.0**: Feature additions, major enhancements (minor version)
  - **v1.0.0**: Stable release (production ready, explicit user approval required)

### 3. `status` - Progress State
- **Type**: enum
- **Values**:
  - `draft`: authoring in progress
  - `active`: implementation in progress
  - `completed`: implementation finished
  - `deprecated`: scheduled for retirement

### 4. `created` - Created Date
- **Type**: date (string)
- **Format**: `YYYY-MM-DD`
- **Example**: `2025-10-06`

### 5. `updated` - Last Updated Date
- **Type**: date (string)
- **Format**: `YYYY-MM-DD`
- **Rule**: Update whenever the SPEC content changes.

### 6. `author` - Author
- **Type**: string
- **Format**: `@{GitHub ID}`
- **Example**: `@Goos`
- **Rules**:
  - Use a single author (do not use an `authors` array)
  - Prepend the GitHub ID with `@`
  - Record additional contributors in the HISTORY section

### 7. `priority` - Priority
- **Type**: enum
- **Values**:
  - `critical`: Immediate action required (security, critical bugs)
  - `high`: High priority (major features)
  - `medium`: Medium priority (enhancements)
  - `low`: Low priority (optimisation, docs)

---

## Optional Fields

### Classification / Meta Fields

#### 8. `category` - Change Type
- **Type**: enum
- **Values**:
  - `feature`: New feature
  - `bugfix`: Bug fix
  - `refactor`: Refactoring
  - `security`: Security improvement
  - `docs`: Documentation
  - `perf`: Performance optimisation

#### 9. `labels` - Tags
- **Type**: array of string
- **Purpose**: Searching, filtering, grouping
- **Example**:
  ```yaml
  labels:
    - installer
    - template
    - security
  ```

### Relationship Fields (Dependency Graph)

#### 10. `depends_on` - Prerequisite SPECs
- **Type**: array of string
- **Meaning**: SPECs that must be completed before this one
- **Example**:
  ```yaml
  depends_on:
    - USER-001
    - AUTH-001
  ```
- **Use case**: Determine execution order and parallelism

#### 11. `blocks` - Blocked SPECs
- **Type**: array of string
- **Meaning**: SPECs that cannot proceed until this one is complete
- **Example**:
  ```yaml
  blocks:
    - PAYMENT-003
  ```

#### 12. `related_specs` - Related SPECs
- **Type**: array of string
- **Meaning**: SPECs without direct dependencies but contextually related
- **Example**:
  ```yaml
  related_specs:
    - TOKEN-002
    - SESSION-001
  ```

#### 13. `related_issue` - Related GitHub Issue
- **Type**: string (URL)
- **Format**: Full GitHub Issue URL
- **Example**:
  ```yaml
  related_issue: "https://github.com/modu-ai/moai-adk/issues/123"
  ```

### Scope / Impact Fields

#### 14. `scope.packages` - Affected Packages
- **Type**: array of string
- **Meaning**: Packages/modules impacted by the SPEC
- **Example**:
  ```yaml
  scope:
    packages:
      - moai-adk-ts/src/core/installer
      - moai-adk-ts/src/core/git
  ```

#### 15. `scope.files` - Key Files
- **Type**: array of string
- **Meaning**: Reference list of key files to change
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
# Verify that every SPEC file includes all required fields
rg "^(id|version|status|created|updated|author|priority):" .moai/specs/SPEC-*/spec.md

# Detect SPECs missing the priority field
rg -L "^priority:" .moai/specs/SPEC-*/spec.md
```

### Format Checks
```bash
# Validate the author field format (e.g., @Goos)
rg "^author: @[A-Z]" .moai/specs/SPEC-*/spec.md

# Validate the version field format (0.x.y)
rg "^version: 0\\.\d+\\.\d+" .moai/specs/SPEC-*/spec.md
```

---

## Migration Guide

### Updating Existing SPECs

#### 1. Add the priority field
Add `priority` if the SPEC does not already include it:
```yaml
priority: medium  # or low|high|critical
```

#### 2. Standardise the author field
- `authors: ["@goos"]` → `author: @Goos`
- Convert lowercase IDs to capitalised forms

#### 3. Add optional fields (recommended)
```yaml
category: refactor
labels:
  - code-quality
  - maintenance
```

---

## Design Principles

### 1. DRY (Don't Repeat Yourself)
- ❌ **Remove**: The `reference` field (every SPEC referenced the same masterplan → redundant)
- ✅ **Alternative**: Document project-level references in README.md

### 2. Context-Aware
- Include only the context that is needed
- Use optional fields only when they add value

### 3. Traceable
- Use `depends_on`, `blocks`, and `related_specs` to express SPEC relationships
- Allow automation to detect circular dependencies

### 4. Maintainable
- Ensure every field can be validated by tooling
- Keep formats consistent for easier parsing

### 5. Simple First
- Minimise complexity
- Limit the structure to 7 required + 9 optional fields
- Expand incrementally when necessary

---

**Last Updated**: 2025-10-06
**Author**: @Alfred
