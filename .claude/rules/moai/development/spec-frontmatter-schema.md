---
description: "SPEC 파일 frontmatter canonical 12-field 스키마 — Single Source of Truth (SSOT)"
paths: "**/*.md,.moai/specs/**/*.md"
---

# SPEC Frontmatter Schema — SSOT

> **Single Source of Truth** for the canonical SPEC frontmatter schema.
> Enforcement: `internal/spec/lint.go` `FrontmatterSchemaRule` (REQ-SPC-003-006).
> Cross-referenced by: `.claude/skills/moai/workflows/plan.md` § Pre-Write Frontmatter Checklist,
> `.claude/skills/moai/team/plan.md` § Pre-Write Frontmatter Checklist.

## Canonical 12 Required Fields

All SPEC documents (`spec.md`) MUST contain exactly these 12 fields in YAML frontmatter.
Missing any field or using a snake_case alias causes `FrontmatterInvalid` lint findings.

```yaml
---
id: SPEC-{DOMAIN}-{NUM}
title: "Human-readable title"
version: "X.Y.Z"
status: draft
created: YYYY-MM-DD
updated: YYYY-MM-DD
author: Author Name
priority: P1
phase: "vX.Y.Z target"
module: "path/to/module"
lifecycle: spec-anchored
tags: "tag1, tag2, tag3"
---
```

## Field Reference

| Field | Type | Constraints | Notes |
|-------|------|-------------|-------|
| `id` | string | `^SPEC-[A-Z][A-Z0-9]+-[0-9]{3}$` | Domain-namespaced identifier |
| `title` | string | non-empty, quoted | Human-readable description |
| `version` | string | semver `X.Y.Z`, quoted | Start at `"0.1.0"` |
| `status` | enum | See Status Enum below | Lifecycle state |
| `created` | date | `YYYY-MM-DD` ISO format | Creation date |
| `updated` | date | `YYYY-MM-DD` ISO format | Last update date |
| `author` | string | non-empty | Author name |
| `priority` | enum | `P0`\|`P1`\|`P2`\|`P3` or `High`\|`Medium`\|`Low`\|`Critical` | Default `P1` |
| `phase` | string | non-empty, typically release target | e.g. `"v3.0.0"` |
| `module` | string | non-empty, path-like | Affected Go module or directory |
| `lifecycle` | enum | `spec-anchored`\|`spec-lite`\|`exploratory` | Default `spec-anchored` |
| `tags` | string | comma-separated, non-empty | Searchable labels |

## Status Enum (8 values)

```
draft → planned → in-progress → implemented → completed
                                         ↓
                               superseded | archived | rejected
```

Valid values: `draft`, `planned`, `in-progress`, `implemented`, `completed`, `superseded`, `archived`, `rejected`

## Optional Fields

These fields may be included when needed but are NOT required by `FrontmatterSchemaRule`:

| Field | Type | Notes |
|-------|------|-------|
| `issue_number` | integer or null | GitHub issue number. Omit entirely when not tracking. |
| `depends_on` | list | SPEC IDs this SPEC depends on. Used by BODP signal A. |
| `lint.skip` | list | Lint rule codes to skip. Use only for documented debt. |
| `bc_id` | string | Backward-compatibility tracking ID. |
| `tier` | enum (`S` \| `M` \| `L`) | SPEC complexity tier (LEAN workflow, SPEC-V3R5-WORKFLOW-LEAN-001). Determines artifact set (S=2 files, M=3, L=5), Section A-E delegation template applicability (S=optional), and plan-auditor PASS threshold (S=0.75 / M=0.80 / L=0.85). See `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier. Backward compat rule: absence = **Tier L** (5-artifact default behavior for pre-LEAN SPECs). |

## Rejected Snake_Case Aliases

The YAML struct decoder in `internal/spec/lint.go` uses `yaml:"created"`, `yaml:"updated"`, `yaml:"tags"` tags.
Snake_case aliases are silently dropped by the decoder, causing empty-value `FrontmatterInvalid` findings.

| Do NOT use | Use instead |
|------------|-------------|
| `created_at:` | `created:` |
| `updated_at:` | `updated:` |
| `labels:` | `tags:` |
| `spec_id:` | `id:` |

## Lint Rule Implementation

`FrontmatterSchemaRule` in `internal/spec/lint.go`:

- **Rule code**: `FrontmatterInvalid`
- **Severity**: Warning
- **REQ coverage**: REQ-SPC-003-006
- **Check**: Iterates all 12 required fields; emits one finding per missing/empty field.
- **YAML binding**: `SPECFrontmatter` struct uses canonical field names (`created`, `updated`, `tags`).
  Snake_case aliases in the source YAML file are not recognized — they produce empty values.

See `internal/spec/lint.go` `FrontmatterSchemaRule.Check()` for the authoritative implementation.

## Examples

### Correct (all 12 fields, canonical names)

```yaml
---
id: SPEC-AUTH-001
title: "OAuth2 Authentication"
version: "0.1.0"
status: draft
created: 2026-05-16
updated: 2026-05-16
author: GOOS Kim
priority: P1
phase: "v3.0.0"
module: "internal/auth"
lifecycle: spec-anchored
tags: "auth, oauth2, security"
---
```

### Wrong (snake_case aliases — produces 3 FrontmatterInvalid findings)

```yaml
---
id: SPEC-AUTH-001
title: "OAuth2 Authentication"
version: "0.1.0"
status: draft
created_at: 2026-05-16   # WRONG — use created:
updated_at: 2026-05-16   # WRONG — use updated:
author: GOOS Kim
priority: P1
phase: "v3.0.0"
module: "internal/auth"
lifecycle: spec-anchored
labels: [auth, oauth2]   # WRONG — use tags: "auth, oauth2"
---
```

## Version History

| Date | Author | Change |
|------|--------|--------|
| (initial) | maintainer | Initial creation — resolves dual-schema drift between plan.md (9-field) and lint.go (12-field) |
