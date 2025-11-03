---
name: "Writing SPEC Documents with Metadata"
description: "Standards for SPEC authoring including YAML metadata, EARS syntax, HISTORY section, and validation. Guides spec-builder agent through SPEC creation in .moai/specs/SPEC-{ID}/ directory. Covers 7 required metadata fields, 5 EARS patterns, version tracking, changelog format, and @SPEC:ID TAG placement. Essential for /alfred:1-plan command and duplicate detection."
allowed-tools: "Read, Write, Bash(rg:*), Bash(mkdir:*)"
---

# SPEC Metadata Standard Skill

## SPEC Directory Structure

```
.moai/specs/SPEC-{DOMAIN}-{NUMBER}/
├── spec.md          # Requirements (EARS format) + metadata
├── plan.md          # Implementation plan
└── acceptance.md    # Test scenarios, acceptance criteria
```

## YAML Metadata (7 Fields)

Every SPEC starts with frontmatter:

```yaml
---
id: AUTH-001
title: "User Authentication System"
version: 0.1.0
status: active
created: 2025-11-03
updated: 2025-11-03
author: @GOOS🪿엉아
priority: high
---
```

### Field Definitions

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `id` | string | ✅ | Domain-NUMBER (e.g., AUTH-001) |
| `title` | string | ✅ | User-facing description |
| `version` | semver | ✅ | Current status (0.x.x = draft) |
| `status` | enum | ✅ | active\|draft\|deprecated\|archived |
| `created` | ISO8601 | ✅ | YYYY-MM-DD |
| `updated` | ISO8601 | ✅ | Last modification date |
| `author` | string | ✅ | Creator identifier |
| `priority` | enum | ✅ | critical\|high\|medium\|low |

## EARS Requirements (5 Patterns)

### 1. Ubiquitous (Baseline)
```
The system shall provide [capability].
The system shall validate [input] before [action].
```

### 2. Event-driven
```
WHEN [condition], the system shall [behavior].
WHEN a user logs in, the system shall verify credentials.
```

### 3. State-driven
```
WHILE [state], the system shall [behavior].
WHILE authenticated, the system shall allow access.
```

### 4. Optional Features
```
WHERE [condition], the system may [behavior].
WHERE API key present, the system may enable caching.
```

### 5. Constraints
```
IF [condition], the system shall [constraint].
IF password invalid, the system shall deny access.
```

## HISTORY Section

Track all changes:

```markdown
## HISTORY

### v0.2.0 (2025-11-03)
- Added email verification requirement
- Updated password complexity rules

### v0.1.0 (2025-11-01)
- Initial draft of AUTH-001
```

## Validation Checklist

- [ ] ID: Domain-NUMBER format (e.g., AUTH-001)
- [ ] No duplicate @SPEC:ID (`rg "@SPEC:{ID}" .moai/specs/`)
- [ ] All 7 metadata fields present
- [ ] version follows semver (0.x.x for draft)
- [ ] 5+ EARS requirements with different patterns
- [ ] HISTORY section with at least v0.1.0
- [ ] Directory: `.moai/specs/SPEC-{ID}/`
- [ ] @SPEC:ID placed in first 50 lines of spec.md

## Common Patterns

| Scenario | Action |
|----------|--------|
| Start new SPEC | Create .moai/specs/SPEC-{DOMAIN}-{NUMBER}/ |
| Check duplicate | rg "@SPEC:{ID}" .moai/specs/ (should be 0) |
| Write requirement | Use one of 5 EARS patterns |
| Update version | Change version field, update HISTORY |
| Mark deprecated | Set status: deprecated, add note |
| Link to @CODE | In code comment: @CODE:SPEC-AUTH-001 |
