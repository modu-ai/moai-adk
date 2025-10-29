---
name: moai-spec-authoring
version: 1.2.0
created: 2025-10-23
updated: 2025-10-29
status: active
description: SPEC document authoring guide - YAML metadata, EARS syntax (5 patterns with Unwanted Behaviors), validation checklist
keywords: ['spec', 'authoring', 'ears', 'metadata', 'requirements', 'tdd', 'planning']
allowed-tools:
  - Read
  - Bash
  - Glob
---

# SPEC Authoring Skill

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-spec-authoring |
| **Version** | 1.2.0 (2025-10-29) |
| **Allowed tools** | Read, Bash, Glob |
| **Auto-load** | `/alfred:1-plan`, SPEC authoring tasks |
| **Tier** | Foundation |

---

## What It Does

Comprehensive guide for authoring SPEC documents in MoAI-ADK. Provides YAML metadata structure (7 required + 9 optional fields), official EARS requirement syntax (5 patterns including Unwanted Behaviors), version management lifecycle, TAG integration, and validation strategies.

**Key capabilities**:
- Step-by-step SPEC creation workflow
- Complete metadata field reference with lifecycle rules
- EARS syntax templates and real-world patterns
- Pre-submission validation checklist
- Common pitfalls prevention guide
- `/alfred:1-plan` workflow integration

---

## When to Use

**Automatic triggers**:
- `/alfred:1-plan` command execution
- SPEC document creation requests
- Requirements clarification discussions
- Feature planning sessions

**Manual invocation**:
- Learn SPEC authoring best practices
- Validate existing SPEC documents
- Troubleshoot metadata issues
- Understand EARS syntax patterns

---

## Quick Start: 5-Step SPEC Creation

### Step 1: Initialize SPEC Directory

```bash
mkdir -p .moai/specs/SPEC-{DOMAIN}-{NUMBER}
# Example: Authentication feature
mkdir -p .moai/specs/SPEC-AUTH-001
```

### Step 2: Write YAML Front Matter

```yaml
---
id: AUTH-001
version: 0.0.1
status: draft
created: 2025-10-29
updated: 2025-10-29
author: @YourGitHubHandle
priority: high
---
```

### Step 3: Add SPEC Title & HISTORY

```markdown
# @SPEC:AUTH-001: JWT Authentication System

## HISTORY

### v0.0.1 (2025-10-29)
- **INITIAL**: JWT authentication SPEC draft created
- **AUTHOR**: @YourHandle
```

### Step 4: Define Environment & Assumptions

```markdown
## Environment

**Runtime**: Node.js 20.x or later
**Framework**: Express.js
**Database**: PostgreSQL 15+

## Assumptions

1. User credentials stored in PostgreSQL
2. JWT secrets managed via environment variables
3. Server clock synchronized with NTP
```

### Step 5: Write EARS Requirements

```markdown
## Requirements

### Ubiquitous Requirements
**UR-001**: The system shall provide JWT-based authentication.

### Event-driven Requirements
**ER-001**: WHEN the user submits valid credentials, the system shall issue a JWT token with 15-minute expiration.

### State-driven Requirements
**SR-001**: WHILE the user is in an authenticated state, the system shall permit access to protected resources.

### Optional Features
**OF-001**: WHERE multi-factor authentication is enabled, the system can require OTP verification after password confirmation.

### Unwanted Behaviors
**UB-001**: IF a token has expired, THEN the system shall deny access and return HTTP 401.
```

---

## Five EARS Pattern Overview

| Pattern | Keyword | Purpose | Example |
|---------|---------|---------|---------|
| **Ubiquitous** | shall | Core functionality always active | "The system shall provide login capability" |
| **Event-driven** | WHEN | Response to specific events | "WHEN login fails, display error" |
| **State-driven** | WHILE | Persistent behavior during state | "WHILE in authenticated state, permit access" |
| **Optional** | WHERE | Conditional features based on flags | "WHERE premium enabled, unlock feature" |
| **Unwanted Behaviors** | IF-THEN | Error handling, quality gates, business rules | "IF token expires, deny access + return 401" |

---

## Seven Required Metadata Fields

1. **id**: `<DOMAIN>-<NUMBER>` (e.g., `AUTH-001`) - Immutable identifier
2. **version**: `MAJOR.MINOR.PATCH` (e.g., `0.0.1`) - Semantic versioning
3. **status**: `draft` | `active` | `completed` | `deprecated`
4. **created**: `YYYY-MM-DD` - Initial creation date
5. **updated**: `YYYY-MM-DD` - Final modification date
6. **author**: `@GitHubHandle` - Primary author (@ prefix required)
7. **priority**: `critical` | `high` | `medium` | `low`

**Version Lifecycle**:
- `0.0.x` → draft (authoring phase)
- `0.1.0` → completed (implementation done)
- `1.0.0` → stable (production-ready)

---

## Validation Checklist

### Metadata Validation
- [ ] All 7 required fields present
- [ ] `author` field includes @ prefix
- [ ] `version` format is `0.x.y`
- [ ] `id` is not duplicated (`rg "@SPEC:AUTH-001" -n .moai/specs/`)

### Content Validation
- [ ] YAML Front Matter complete
- [ ] Title includes `@SPEC:{ID}` TAG block
- [ ] HISTORY section has v0.0.1 INITIAL entry
- [ ] Environment section defined
- [ ] Assumptions section defined (minimum 3 items)
- [ ] Requirements section uses EARS patterns
- [ ] Traceability section shows TAG chain structure

### EARS Syntax Validation
- [ ] Ubiquitous: "shall" + capability
- [ ] Event-driven: Starts with "WHEN [trigger]"
- [ ] State-driven: Starts with "WHILE [state]"
- [ ] Optional: Starts with "WHERE [feature]", uses "can"
- [ ] Unwanted Behaviors: "IF-THEN" or direct constraint expression

---

## Common Pitfalls

1. ❌ **Changing SPEC ID after assignment** → Breaks TAG chain
2. ❌ **Skipping HISTORY updates** → Content changes without audit trail
3. ❌ **Jumping version numbers** → v0.0.1 → v1.0.0 without intermediate steps
4. ❌ **Ambiguous requirements** → "Fast and user-friendly" (unmeasurable)
5. ❌ **Missing @ prefix in author** → `author: Goos` instead of `author: @Goos`
6. ❌ **Mixing EARS patterns** → Multiple keywords in single requirement

---

## Related Skills

- `moai-foundation-ears` - Official EARS syntax patterns
- `moai-foundation-specs` - Metadata validation automation
- `moai-foundation-tags` - TAG system integration
- `moai-alfred-spec-metadata-validation` - Automated validation

---

## Detailed Reference

- **Full Metadata Reference**: [reference.md](./reference.md)
- **Practical Examples**: [examples.md](./examples.md)

---

**Last Updated**: 2025-10-29
**Version**: 1.2.0
**Maintained By**: MoAI-ADK Team
**Support**: Use `/alfred:1-plan` command for guided SPEC creation
