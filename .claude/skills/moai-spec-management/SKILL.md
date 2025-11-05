---
name: moai-spec-management
version: 1.0.0
created: 2025-11-06
updated: 2025-11-06
status: active
description: Complete SPEC lifecycle management from authoring to validation with EARS syntax, YAML metadata, version control, and quality gates. End-to-end SPEC document creation, validation, and maintenance with TAG integration. Use when authoring SPECs, validating requirements, managing SPEC versions, or ensuring SPEC quality.
keywords: ['spec', 'lifecycle', 'ears', 'yaml', 'validation', 'authoring', 'requirements', 'metadata', 'tdd']
allowed-tools:
  - Read
  - Bash
  - Glob
---

# SPEC Management - Complete Lifecycle System

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-spec-management |
| **Version** | 1.0.0 (2025-11-06) |
| **Status** | Active |
| **Tier** | Foundation |
| **Purpose** | Complete SPEC lifecycle management |

---

## What It Does

End-to-end SPEC document management system covering creation, validation, versioning, and quality assurance.

**Core capabilities**:
- ✅ SPEC authoring with complete workflow guidance
- ✅ EARS requirement syntax (5 official patterns)
- ✅ YAML metadata validation (7 required fields)
- ✅ Version lifecycle management
- ✅ Quality gate validation
- ✅ TAG system integration
- ✅ History tracking and audit trails

---

## When to Use

**SPEC Creation**:
- `/alfred:1-plan` command execution
- New feature planning sessions
- Requirements clarification discussions
- Project initialization phases

**SPEC Validation**:
- Before implementation starts
- During code review processes
- Quality gate validation
- Compliance checking

**SPEC Maintenance**:
- Version updates and lifecycle management
- History tracking and audit trails
- Migration to newer versions
- Deprecation and archiving

---

## Quick Start: 5-Step SPEC Creation

### Step 1: Initialize SPEC Directory

```bash
# Create SPEC directory with proper naming
mkdir -p .moai/specs/SPEC-{DOMAIN}-{NUMBER}

# Examples:
mkdir -p .moai/specs/SPEC-AUTH-001  # Authentication
mkdir -p .moai/specs/SPEC-API-042   # API endpoints
mkdir -p .moai/specs/SPEC-UI-017    # UI components
```

### Step 2: Write Complete YAML Front Matter

```yaml
---
id: AUTH-001
version: 0.0.1
status: draft
created: 2025-11-06
updated: 2025-11-06
author: @YourGitHubHandle
priority: high
title: JWT Authentication System
description: Complete JWT-based authentication with refresh tokens
dependencies: []
related_specs: []
tags: ['authentication', 'jwt', 'security']
---

# @SPEC:AUTH-001: JWT Authentication System
```

### Step 3: Add History Section

```markdown
## HISTORY

### v0.0.1 (2025-11-06)
- **INITIAL**: JWT authentication SPEC draft created
- **AUTHOR**: @YourHandle
- **SCOPE**: Basic JWT authentication with refresh tokens
- **DEPENDENCIES**: PostgreSQL for user storage, Redis for token blacklist
```

### Step 4: Define Environment & Assumptions

```markdown
## Environment

**Runtime**: Node.js 20.x or later
**Framework**: Express.js 4.x
**Database**: PostgreSQL 15+
**Cache**: Redis 7+
**Authentication**: JWT with RS256 signing

## Assumptions

1. User credentials stored securely in PostgreSQL with bcrypt hashing
2. JWT secrets managed via environment variables or KMS
3. Server clock synchronized with NTP for token expiration
4. SSL/TLS termination handled by reverse proxy
5. Rate limiting implemented at application level
```

### Step 5: Write EARS Requirements

```markdown
## Requirements

### Ubiquitous Requirements
**UR-001**: The system shall provide JWT-based authentication for all protected resources.

### Event-driven Requirements
**ER-001**: WHEN the user submits valid credentials, the system shall issue a JWT access token with 15-minute expiration AND a refresh token with 7-day expiration.

**ER-002**: WHEN a refresh token is presented within its validity period, the system shall issue a new access token without requiring re-authentication.

### State-driven Requirements
**SR-001**: WHILE the user is in an authenticated state, the system shall permit access to protected resources based on their role permissions.

**SR-002**: WHILE a token blacklist entry exists, the system shall reject access tokens even if they are cryptographically valid.

### Optional Features
**OF-001**: WHERE multi-factor authentication is enabled, the system shall require OTP verification after password confirmation.

**OF-002**: WHERE social login is configured, the system can authenticate users via OAuth 2.0 providers.

### Unwanted Behaviors
**UB-001**: IF a token has expired, THEN the system shall deny access and return HTTP 401 with appropriate error message.

**UB-002**: IF a token signature is invalid, THEN the system shall deny access and log the security event.

**UB-003**: IF more than 5 failed login attempts occur within 15 minutes, THEN the system shall temporarily lock the account for 30 minutes.
```

---

## EARS Requirements Syntax

### Five Official Patterns

| Pattern | Keyword | Purpose | Template | Example |
|---------|---------|---------|----------|---------|
| **Ubiquitous** | shall | Core functionality always active | `The system shall [capability]` | "The system shall provide login capability" |
| **Event-driven** | WHEN | Response to specific events | `WHEN [trigger], the system shall [response]` | "WHEN login fails, display error message" |
| **State-driven** | WHILE | Persistent behavior during state | `WHILE [state], the system shall [behavior]` | "WHILE authenticated, permit access to resources" |
| **Optional** | WHERE | Conditional features | `WHERE [condition], the system can [feature]` | "WHERE premium enabled, unlock advanced features" |
| **Unwanted Behaviors** | IF-THEN | Error handling, constraints | `IF [condition], THEN the system shall [action]` | "IF token expires, deny access" |

### Pattern Writing Guidelines

**Ubiquitous Requirements**:
- Describe core system capabilities
- Always active, no trigger needed
- Use "shall" for mandatory behavior
- Focus on what system provides, not how

**Event-driven Requirements**:
- Start with "WHEN [specific trigger]"
- Describe immediate response
- Include success and failure paths
- Be precise about trigger conditions

**State-driven Requirements**:
- Start with "WHILE [state persists]"
- Describe continuous behavior
- Include state entry and exit conditions
- Cover duration and persistence

**Optional Features**:
- Start with "WHERE [feature flag/condition]"
- Use "can" instead of "shall"
- Describe conditional capabilities
- Include enable/disable criteria

**Unwanted Behaviors**:
- Use "IF-THEN" or direct constraints
- Cover error conditions and edge cases
- Include security constraints
- Describe quality gates and limits

---

## YAML Metadata Complete Reference

### 7 Required Fields

| Field | Format | Example | Validation Rules |
|-------|--------|---------|-----------------|
| **id** | `<DOMAIN>-<NUMBER>` | `AUTH-001` | Unique, immutable, domain prefix |
| **version** | `MAJOR.MINOR.PATCH` | `0.0.1` | Semantic versioning, start with 0.0.1 |
| **status** | `draft\|active\|completed\|deprecated` | `draft` | Must be one of 4 values |
| **created** | `YYYY-MM-DD` | `2025-11-06` | Creation date, immutable |
| **updated** | `YYYY-MM-DD` | `2025-11-06` | Last modification date |
| **author** | `@GitHubHandle` | `@username` | @ prefix required, GitHub handle |
| **priority** | `critical\|high\|medium\|low` | `high` | Must be one of 4 values |

### 9 Optional Fields

| Field | Format | Example | Purpose |
|-------|--------|---------|---------|
| **title** | String | "JWT Authentication System" | Human-readable title |
| **description** | Text | "Complete JWT auth..." | Brief SPEC description |
| **dependencies** | List | `["SPEC-DB-001"]` | Required SPEC dependencies |
| **related_specs** | List | `["SPEC-UI-005"]` | Related but optional SPECs |
| **tags** | List | `["authentication", "security"]` | Categorization tags |
| **estimated_effort** | String | "2 weeks" | Implementation effort estimate |
| **assigned_to** | String | `@developer` | Developer assignment |
| **reviewers** | List | `["@reviewer1", "@reviewer2"]` | SPEC reviewer assignments |
| **target_date** | Date | `2025-12-01` | Target completion date |

### Version Lifecycle Management

**Version Progression**:
```
0.0.1 → 0.0.2 → ... → 0.1.0 → 0.2.0 → ... → 1.0.0
  ↓         ↓             ↓         ↓             ↓
draft    draft        completed  completed    stable
```

**Status Transitions**:
- `draft` → `active`: Ready for implementation
- `active` → `completed`: Implementation finished
- `completed` → `deprecated`: Replaced by newer version
- Any → `draft`: Major revision needed

**Version Bumping Rules**:
- `PATCH` (0.0.1 → 0.0.2): Minor corrections, typos
- `MINOR` (0.0.1 → 0.1.0): Feature additions
- `MAJOR` (0.1.0 → 1.0.0): Breaking changes, production-ready

---

## Validation System

### Pre-Submission Checklist

**Metadata Validation**:
- [ ] All 7 required fields present and valid
- [ ] `author` field includes @ prefix
- [ ] `version` follows semantic versioning
- [ ] `id` is unique across all SPECs
- [ ] Dates are in YYYY-MM-DD format

**Content Validation**:
- [ ] YAML Front Matter is valid
- [ ] Title includes `@SPEC:{ID}` TAG block
- [ ] HISTORY section has v0.0.1 INITIAL entry
- [ ] Environment section clearly defined
- [ ] Assumptions section has minimum 3 items
- [ ] Requirements section uses proper EARS patterns
- [ ] Traceability section shows TAG chain structure

**Quality Validation**:
- [ ] Requirements are measurable and testable
- [ ] No ambiguous terms ("fast", "user-friendly")
- [ ] Each requirement has a unique identifier
- [ ] Dependencies clearly documented
- [ ] Edge cases covered

### EARS Syntax Validation

**Pattern Compliance**:
- [ ] Ubiquitous: "shall" + clear capability
- [ ] Event-driven: Starts with "WHEN [trigger]"
- [ ] State-driven: Starts with "WHILE [state]"
- [ ] Optional: Starts with "WHERE [condition]", uses "can"
- [ ] Unwanted Behaviors: "IF-THEN" structure or direct constraint

**Quality Checks**:
- [ ] One concept per requirement
- [ ] No mixed patterns in single requirement
- [ ] Clear success/failure criteria
- [ ] Measurable outcomes
- [ ] Complete trigger-response pairs

---

## Common Pitfalls & Solutions

### Critical Mistakes

1. **Changing SPEC ID after assignment**
   - ❌ Breaking change to TAG chain
   - ✅ Keep ID immutable, create new SPEC if needed

2. **Skipping HISTORY updates**
   - ❌ Content changes without audit trail
   - ✅ Update HISTORY for every content change

3. **Jumping version numbers**
   - ❌ v0.0.1 → v1.0.0 without intermediate steps
   - ✅ Follow progression: 0.0.1 → 0.1.0 → 1.0.0

4. **Ambiguous requirements**
   - ❌ "Fast and user-friendly interface"
   - ✅ "Response time < 200ms, 3-click navigation"

5. **Missing @ prefix in author**
   - ❌ `author: Goos`
   - ✅ `author: @Goos`

### Pattern Violations

1. **Mixing EARS patterns**
   - ❌ "WHEN user logs in AND is admin, shall see dashboard"
   - ✅ Split into separate requirements

2. **Missing trigger conditions**
   - ❌ "System shall send notifications"
   - ✅ "WHEN event occurs, system shall send notification"

3. **Unmeasurable criteria**
   - ❌ "System shall be secure"
   - ✅ "System shall pass OWASP Top 10 security scan"

---

## Integration with TAG System

### TAG Chain Integration

**SPEC Creation**:
```
1. Create @SPEC:AUTH-001 during planning
2. Reference in implementation with @TEST:AUTH-001, @CODE:AUTH-001
3. Complete with @DOC:AUTH-001 in documentation
```

**Cross-Reference Validation**:
```bash
# Check for orphan TAGs
rg "@TAG-[A-Z]+-[0-9]+" .moai/specs/ -A 2 -B 2

# Validate TAG chain integrity
rg "@SPEC:AUTH-001" . --type md
rg "@TEST:AUTH-001" . --type md
rg "@CODE:AUTH-001" . --type md
rg "@DOC:AUTH-001" . --type md
```

### Traceability Management

**Requirements to Implementation**:
```
@SPEC:AUTH-001 (UR-001) → @TEST:AUTH-001 (test-login) → @CODE:AUTH-001 (auth-service)
```

**Change Impact Analysis**:
- When SPEC changes, identify affected TEST/CODE/DOC TAGs
- Update all downstream references
- Document changes in HISTORY section

---

## Quality Gates

### Before Implementation
- [ ] SPEC validation passed
- [ ] All stakeholders reviewed
- [ ] Dependencies available
- [ ] Environment constraints documented

### During Implementation
- [ ] Tests written for all requirements
- [ ] Code matches SPEC requirements
- [ ] TAG references properly maintained
- [ ] Deviation risks documented

### After Implementation
- [ ] All requirements implemented
- [ ] Tests achieve required coverage
- [ ] Documentation updated
- [ ] SPEC status updated to completed

---

## Automation & Tools

### Validation Scripts

```bash
# Validate SPEC YAML frontmatter
python3 .moai/scripts/spec_validator.py .moai/specs/SPEC-AUTH-001/spec.md

# Check EARS syntax compliance
python3 .moai/scripts/ears_validator.py .moai/specs/SPEC-AUTH-001/spec.md

# Generate SPEC quality report
python3 .moai/scripts/spec_quality_report.py .moai/specs/
```

### Quality Metrics

- **Completeness**: % of required sections filled
- **Clarity**: Ambiguity score based on measurable criteria
- **Testability**: % of requirements with testable criteria
- **Traceability**: TAG chain completeness score

---

**End of Skill** | Consolidated from moai-alfred-spec-authoring + moai-foundation-specs + moai-foundation-ears
