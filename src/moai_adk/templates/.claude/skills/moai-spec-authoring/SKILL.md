---
name: moai-spec-authoring
version: 1.0.0
created: 2025-10-23
updated: 2025-10-23
status: active
description: Comprehensive guide for authoring MoAI-ADK SPEC documents with proper metadata, EARS syntax, and best practices.
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
| **Version** | 1.0.0 (2025-10-23) |
| **Allowed tools** | Read, Bash, Glob |
| **Auto-load** | `/alfred:1-plan`, SPEC writing tasks |
| **Tier** | Foundation |

---

## What It Does

Comprehensive guide for authoring high-quality MoAI-ADK SPEC documents. Covers YAML metadata structure (7 required + 9 optional fields), EARS requirement syntax (5 patterns), versioning lifecycle, TAG integration, and validation strategies.

**Key capabilities**:
- Step-by-step SPEC creation workflow
- Complete metadata field reference with examples
- EARS syntax templates with real-world patterns
- Pre-submission validation checklist
- Common pitfalls prevention
- Integration with `/alfred:1-plan` workflow

---

## When to Use

**Automatic triggers**:
- `/alfred:1-plan` command execution
- SPEC document creation requests
- Requirement clarification discussions
- Feature planning sessions

**Manual invocation**:
- Learning SPEC authoring best practices
- Validating existing SPEC documents
- Troubleshooting metadata issues
- Understanding EARS syntax patterns

---

## Quick Start: 5-Step SPEC Creation

### Step 1: Initialize SPEC Directory

```bash
# Create SPEC directory structure
mkdir -p .moai/specs/SPEC-{DOMAIN}-{NUMBER}
cd .moai/specs/SPEC-{DOMAIN}-{NUMBER}

# Example: Authentication feature
mkdir -p .moai/specs/SPEC-AUTH-001
```

### Step 2: Create `spec.md` with YAML Front Matter

```yaml
---
# Required Fields (7)
id: AUTH-001
version: 0.0.1
status: draft
created: 2025-10-23
updated: 2025-10-23
author: @YourGitHubHandle
priority: high

# Optional Fields (recommended)
category: feature
labels:
  - authentication
  - security
  - jwt
---
```

### Step 3: Add SPEC Title & HISTORY

```markdown
# @SPEC:AUTH-001: JWT Authentication System

## HISTORY

### v0.0.1 (2025-10-23)
- **INITIAL**: JWT-based authentication SPEC draft created
- **AUTHOR**: @YourGitHubHandle
- **SCOPE**: User authentication, token generation, token validation
- **CONTEXT**: Requirement from product roadmap
```

### Step 4: Define Environment & Assumptions

```markdown
## Environment

**Execution Context**:
- Runtime: Node.js 20.x or later
- Framework: Express.js
- Database: PostgreSQL 15+

**Technical Stack**:
- JWT library: jsonwebtoken ^9.0.0
- Hashing: bcrypt ^5.1.0

**Constraints**:
- Token lifetime: 15 minutes (access), 7 days (refresh)
- Security: HTTPS required in production

## Assumptions

1. **User Storage**: User credentials stored in PostgreSQL
2. **Secret Management**: JWT secrets managed via environment variables
3. **Clock Sync**: Server clocks synchronized (NTP)
4. **Password Policy**: Minimum 8 characters enforced at signup
```

### Step 5: Write EARS Requirements

```markdown
## Requirements

### Ubiquitous Requirements

**UR-001**: The system shall provide JWT-based authentication.

**UR-002**: The system shall support user login with email and password.

**UR-003**: The system shall issue access tokens and refresh tokens.

### Event-driven Requirements

**ER-001**: WHEN a user submits valid credentials, the system shall issue a JWT access token with 15-minute expiry.

**ER-002**: WHEN a token expires, the system shall return HTTP 401 Unauthorized.

**ER-003**: WHEN a refresh token is presented, the system shall issue a new access token IF the refresh token is valid.

### State-driven Requirements

**SR-001**: WHILE a user is authenticated, the system shall allow access to protected resources.

**SR-002**: WHILE a token is valid, the system shall extract user ID from token claims.

### Optional Features

**OF-001**: WHERE multi-factor authentication is enabled, the system can require OTP verification after password check.

**OF-002**: WHERE session logging is enabled, the system can record login timestamps and IP addresses.

### Constraints

**C-001**: IF a token is expired, the system shall deny access and return HTTP 401.

**C-002**: IF more than 5 failed login attempts occur within 10 minutes, the system should temporarily lock the account.

**C-003**: Access tokens shall not exceed 15-minute lifetime.

**C-004**: Refresh tokens shall not exceed 7-day lifetime.
```

---

## Metadata Reference

### 7 Required Fields

#### 1. `id` – Unique SPEC Identifier

**Format**: `<DOMAIN>-<NUMBER>`

**Rules**:
- Immutable once assigned
- Use uppercase domain (e.g., `AUTH`, `PAYMENT`, `CONFIG`)
- Three-digit number (001–999)
- Check duplicates: `rg "@SPEC:AUTH-001" -n .moai/specs/`

**Examples**:
- `AUTH-001` (Authentication feature)
- `INSTALLER-SEC-001` (Installer security)
- `TRUST-001` (TRUST principles)
- `CONFIG-001` (Configuration schema)

**Directory Structure**:
```
.moai/specs/SPEC-AUTH-001/
  ├── spec.md          # Main SPEC document
  ├── diagrams/        # Optional: Architecture diagrams
  └── examples/        # Optional: Code examples
```

#### 2. `version` – Semantic Version

**Format**: `MAJOR.MINOR.PATCH`

**Lifecycle**:

| Version | Status | Description | Trigger |
|---------|--------|-------------|---------|
| `0.0.1` | draft | Initial draft | SPEC creation |
| `0.0.x` | draft | Draft refinements | Content edits |
| `0.1.0` | completed | Implementation complete | `/alfred:3-sync` after TDD |
| `0.1.x` | completed | Bug fixes, doc updates | Post-implementation patches |
| `0.x.0` | completed | Feature additions | Minor enhancements |
| `1.0.0` | completed | Production stable | Stakeholder approval |

**Version Update Examples**:
```markdown
## HISTORY

### v0.2.0 (2025-11-15)
- **ADDED**: Multi-factor authentication support
- **CHANGED**: Token expiry extended to 30 minutes
- **AUTHOR**: @YourHandle

### v0.1.0 (2025-10-30)
- **COMPLETED**: TDD implementation finished
- **EVIDENCE**: Commits 4c66076, 34e1bd9
- **TEST COVERAGE**: 89.13%

### v0.0.2 (2025-10-25)
- **REFINED**: Added password reset flow requirements
- **AUTHOR**: @YourHandle

### v0.0.1 (2025-10-23)
- **INITIAL**: JWT authentication SPEC draft created
```

#### 3. `status` – Progress State

**Values**: `draft` | `active` | `completed` | `deprecated`

**Lifecycle Flow**:
```
draft → active → completed → [deprecated]
  ↓       ↓          ↓
/alfred:1-plan  /alfred:2-run  /alfred:3-sync
```

**Transitions**:
- `draft`: Authoring in progress (v0.0.x)
- `active`: Implementation underway (v0.0.x → v0.1.0)
- `completed`: Implementation finished (v0.1.0+)
- `deprecated`: Marked for retirement

#### 4. `created` – Creation Date

**Format**: `YYYY-MM-DD`

**Rules**:
- Set once, never changes
- ISO 8601 date format
- Represents first draft date

**Example**: `created: 2025-10-23`

#### 5. `updated` – Last Modified Date

**Format**: `YYYY-MM-DD`

**Rules**:
- Update on every content change
- Initially same as `created`
- Reflects latest edit date

**Update Pattern**:
```yaml
created: 2025-10-23   # Never changes
updated: 2025-10-25   # Changes with edits
```

#### 6. `author` – Primary Author

**Format**: `@{GitHubHandle}`

**Rules**:
- Single value (not array)
- Prefix with `@`
- Use proper casing (e.g., `@Goos`, not `@goos`)
- Additional contributors go in HISTORY section

**Examples**:
```yaml
# Correct
author: @Goos

# Incorrect
author: goos           # Missing @
authors: [@Goos]       # Array not allowed
author: @goos          # Wrong casing
```

#### 7. `priority` – Work Priority

**Values**: `critical` | `high` | `medium` | `low`

**Guidelines**:

| Priority | Description | Examples |
|----------|-------------|----------|
| `critical` | Production blockers, security vulnerabilities | Security patches, critical bugs |
| `high` | Major features, core functionality | Authentication, payment system |
| `medium` | Enhancements, improvements | UI polish, performance optimization |
| `low` | Nice-to-haves, documentation | README updates, minor refactors |

---

### 9 Optional Fields

#### 8. `category` – Change Type

**Values**: `feature` | `bugfix` | `refactor` | `security` | `docs` | `perf`

**Usage**:
```yaml
category: feature       # New functionality
category: bugfix        # Defect resolution
category: refactor      # Code structure improvement
category: security      # Security enhancement
category: docs          # Documentation update
category: perf          # Performance optimization
```

#### 9. `labels` – Classification Tags

**Format**: Array of strings

**Purpose**: Search, filtering, grouping

**Best Practices**:
- Use lowercase, kebab-case
- 2-5 labels per SPEC
- Avoid redundancy with `category`

**Examples**:
```yaml
labels:
  - authentication
  - jwt
  - security

labels:
  - performance
  - optimization
  - caching

labels:
  - installer
  - template
  - cross-platform
```

#### 10-13. Relationship Fields (Dependency Graph)

##### `depends_on` – Required SPECs

**Meaning**: SPECs that must be completed first

**Example**:
```yaml
depends_on:
  - USER-001      # User model SPEC
  - TOKEN-001     # Token generation SPEC
```

**Use Case**: Determines execution order, parallelization

##### `blocks` – Blocked SPECs

**Meaning**: SPECs that cannot proceed until this one is resolved

**Example**:
```yaml
blocks:
  - AUTH-002      # OAuth integration waits for basic auth
  - PAYMENT-001   # Payment needs authentication
```

##### `related_specs` – Associated SPECs

**Meaning**: Related items without direct dependencies

**Example**:
```yaml
related_specs:
  - SESSION-001   # Session management (related but independent)
  - AUDIT-001     # Audit logging (cross-cutting concern)
```

##### `related_issue` – Linked GitHub Issue

**Format**: Full GitHub issue URL

**Example**:
```yaml
related_issue: "https://github.com/modu-ai/moai-adk/issues/42"
```

#### 14-15. Scope Fields (Impact Analysis)

##### `scope.packages` – Impacted Packages

**Purpose**: Track which packages/modules are affected

**Example**:
```yaml
scope:
  packages:
    - src/core/auth
    - src/core/token
    - src/api/routes/auth
```

##### `scope.files` – Key Files

**Purpose**: Reference primary implementation files

**Example**:
```yaml
scope:
  files:
    - auth-service.ts
    - token-manager.ts
    - auth.routes.ts
```

---

## EARS Requirement Syntax

### The 5 EARS Patterns

EARS (Easy Approach to Requirements Syntax) provides disciplined, testable requirements using familiar keywords.

#### Pattern 1: Ubiquitous Requirements

**Template**: `The system shall [capability].`

**Purpose**: Baseline functionality always active

**Characteristics**:
- No preconditions
- Always applicable
- Defines core features

**Examples**:
```markdown
**UR-001**: The system shall provide user authentication.

**UR-002**: The system shall support HTTPS connections.

**UR-003**: The system shall store user credentials securely.

**UR-004**: The mobile app shall have a mass of less than 50 MB.

**UR-005**: The API response time shall not exceed 200ms for 95% of requests.
```

**Best Practices**:
- ✅ Use active voice
- ✅ Single responsibility per requirement
- ✅ Measurable outcomes
- ❌ Avoid vague terms ("user-friendly", "fast")

#### Pattern 2: Event-driven Requirements

**Template**: `WHEN [trigger], the system shall [response].`

**Purpose**: Define behavior triggered by specific events

**Characteristics**:
- Triggered by discrete events
- One-time response
- Cause-and-effect relationship

**Examples**:
```markdown
**ER-001**: WHEN a user submits valid credentials, the system shall issue a JWT token.

**ER-002**: WHEN a token expires, the system shall return HTTP 401 Unauthorized.

**ER-003**: WHEN a user clicks "Forgot Password", the system shall send a password reset email.

**ER-004**: WHEN the database connection fails, the system shall retry 3 times with exponential backoff.

**ER-005**: WHEN a file upload exceeds 10 MB, the system shall reject the upload with error message.
```

**Advanced Pattern** (with postcondition):
```markdown
**ER-006**: WHEN a payment transaction completes, the system shall send a confirmation email THEN update the order status to "paid".
```

**Best Practices**:
- ✅ Single trigger per requirement
- ✅ Specific, testable response
- ✅ Include error conditions
- ❌ Don't chain multiple WHENs

#### Pattern 3: State-driven Requirements

**Template**: `WHILE [state], the system shall [behavior].`

**Purpose**: Continuous behavior during a state

**Characteristics**:
- Active while state persists
- Continuous monitoring
- State-dependent behavior

**Examples**:
```markdown
**SR-001**: WHILE a user is authenticated, the system shall allow access to protected routes.

**SR-002**: WHILE a token is valid, the system shall extract user ID from token claims.

**SR-003**: WHILE the system is in maintenance mode, the system shall return HTTP 503 Service Unavailable.

**SR-004**: WHILE battery level is below 20%, the mobile app shall reduce background sync frequency.

**SR-005**: WHILE a file upload is in progress, the UI shall display a progress bar.
```

**Best Practices**:
- ✅ Clearly define state boundaries
- ✅ Specify state entry/exit conditions
- ✅ Test state transitions
- ❌ Avoid overlapping states

#### Pattern 4: Optional Features

**Template**: `WHERE [feature], the system can [behavior].`

**Purpose**: Conditional functionality based on feature flags

**Characteristics**:
- Applies only if feature is present
- Configuration-dependent
- Product variation support

**Examples**:
```markdown
**OF-001**: WHERE multi-factor authentication is enabled, the system can require OTP verification.

**OF-002**: WHERE session logging is enabled, the system can record login timestamps.

**OF-003**: WHERE premium subscription is active, the system can allow unlimited API calls.

**OF-004**: WHERE dark mode is selected, the UI can render with dark color scheme.

**OF-005**: WHERE analytics consent is granted, the system can track user behavior.
```

**Best Practices**:
- ✅ Use "can" (permissive) not "shall" (mandatory)
- ✅ Clearly define feature flag conditions
- ✅ Specify default behavior when feature is absent
- ❌ Don't make core features optional

#### Pattern 5: Constraints

**Template**: `IF [condition], THEN the system shall [constraint].`

**Purpose**: Enforce quality attributes and business rules

**Characteristics**:
- Conditional enforcement
- Quality gates
- Business rule validation

**Examples**:
```markdown
**C-001**: IF a token is expired, THEN the system shall deny access and return HTTP 401.

**C-002**: IF more than 5 failed login attempts occur within 10 minutes, THEN the system should temporarily lock the account.

**C-003**: Access tokens shall not exceed 15-minute lifetime.

**C-004**: IF a password is less than 8 characters, THEN the system shall reject registration.

**C-005**: IF API rate limit is exceeded, THEN the system shall return HTTP 429 Too Many Requests.
```

**Simplified Constraint** (no condition):
```markdown
**C-006**: The system shall not store plaintext passwords.

**C-007**: All API endpoints shall require authentication except /health and /login.
```

**Best Practices**:
- ✅ Use SHALL for hard constraints, SHOULD for soft recommendations
- ✅ Quantify limits (time, size, count)
- ✅ Specify enforcement mechanism
- ❌ Avoid vague constraints

---

### EARS Pattern Selection Guide

| Pattern | Keyword | Use When | Example Context |
|---------|---------|----------|-----------------|
| **Ubiquitous** | shall | Core functionality, always active | "System shall provide login" |
| **Event-driven** | WHEN | Response to discrete event | "WHEN login fails, show error" |
| **State-driven** | WHILE | Continuous behavior in state | "WHILE logged in, allow access" |
| **Optional** | WHERE | Feature flag or configuration | "WHERE premium, unlock features" |
| **Constraints** | IF-THEN | Quality gates, business rules | "IF expired, deny access" |

---

### Real-World EARS Examples

#### Example 1: E-commerce Checkout

```markdown
### Ubiquitous Requirements
**UR-001**: The system shall provide a shopping cart.
**UR-002**: The system shall support credit card payments.

### Event-driven Requirements
**ER-001**: WHEN a user adds an item to cart, the system shall update the cart total.
**ER-002**: WHEN payment is successful, the system shall send order confirmation email.
**ER-003**: WHEN inventory is insufficient, the system shall display "out of stock" message.

### State-driven Requirements
**SR-001**: WHILE items are in cart, the system shall reserve inventory for 30 minutes.
**SR-002**: WHILE payment is processing, the UI shall display a loading indicator.

### Optional Features
**OF-001**: WHERE express shipping is selected, the system can calculate expedited shipping cost.
**OF-002**: WHERE gift wrapping is available, the system can offer gift wrap option.

### Constraints
**C-001**: IF cart total is below $50, THEN the system shall add $5 shipping fee.
**C-002**: IF payment fails after 3 attempts, THEN the system shall lock the order for 1 hour.
**C-003**: Order processing time shall not exceed 5 seconds.
```

#### Example 2: Mobile App Push Notifications

```markdown
### Ubiquitous Requirements
**UR-001**: The app shall support push notifications.
**UR-002**: The app shall allow users to enable/disable notifications.

### Event-driven Requirements
**ER-001**: WHEN a new message arrives, the app shall display a push notification.
**ER-002**: WHEN a notification is tapped, the app shall navigate to the message screen.
**ER-003**: WHEN notification permission is denied, the app shall show in-app notification banner.

### State-driven Requirements
**SR-001**: WHILE the app is in foreground, the system shall display in-app banners instead of push notifications.
**SR-002**: WHILE Do Not Disturb mode is active, the system shall silence all notifications.

### Optional Features
**OF-001**: WHERE notification sounds are enabled, the system can play notification sound.
**OF-002**: WHERE notification grouping is supported, the system can group notifications by conversation.

### Constraints
**C-001**: IF more than 10 notifications are pending, THEN the system shall group them into a summary notification.
**C-002**: Notification delivery latency shall not exceed 5 seconds.
```

---

## HISTORY Section Format

The HISTORY section documents all SPEC versions and their changes.

### Structure

```markdown
## HISTORY

### v{MAJOR}.{MINOR}.{PATCH} ({YYYY-MM-DD})
- **{CHANGE_TYPE}**: {Description}
- **AUTHOR**: {GitHub handle}
- **{Additional context}**: {Details}
```

### Change Types

| Type | Description | Example |
|------|-------------|---------|
| **INITIAL** | First draft | `v0.0.1: INITIAL draft created` |
| **REFINED** | Content updates during draft | `v0.0.2: REFINED requirements based on review` |
| **COMPLETED** | Implementation finished | `v0.1.0: COMPLETED TDD implementation` |
| **ADDED** | New requirements/features | `v0.2.0: ADDED multi-factor authentication` |
| **CHANGED** | Modified requirements | `v0.2.0: CHANGED token expiry from 15min to 30min` |
| **FIXED** | Bug fixes post-implementation | `v0.1.1: FIXED token refresh logic` |
| **DEPRECATED** | Marked for retirement | `v1.5.0: DEPRECATED legacy auth endpoint` |

### Complete HISTORY Example

```markdown
## HISTORY

### v0.2.0 (2025-11-15)
- **ADDED**: Multi-factor authentication support via OTP
- **CHANGED**: Token expiry extended to 30 minutes based on user feedback
- **AUTHOR**: @Goos
- **REVIEWER**: @SecurityTeam
- **RATIONALE**: Improved UX while maintaining security posture

### v0.1.1 (2025-11-01)
- **FIXED**: Token refresh race condition
- **EVIDENCE**: Commit 3f9a2b7
- **AUTHOR**: @Goos

### v0.1.0 (2025-10-30)
- **COMPLETED**: TDD implementation finished
- **AUTHOR**: @Goos
- **EVIDENCE**: Commits 4c66076, 34e1bd9, 1dec08f
- **TEST COVERAGE**: 89.13% (target: 85%)
- **QUALITY METRICS**:
  - Test Pass Rate: 100% (42/42 tests)
  - Linting: ruff ✅
  - Type Checking: mypy ✅
- **TAG CHAIN**:
  - @SPEC:AUTH-001: 1 occurrence
  - @TEST:AUTH-001: 8 occurrences
  - @CODE:AUTH-001: 12 occurrences

### v0.0.2 (2025-10-25)
- **REFINED**: Added password reset flow requirements
- **REFINED**: Clarified token lifetime constraints
- **AUTHOR**: @Goos

### v0.0.1 (2025-10-23)
- **INITIAL**: JWT authentication SPEC draft created
- **AUTHOR**: @Goos
- **SCOPE**: User authentication, token generation, token validation
- **CONTEXT**: Requirement from Q4 2025 product roadmap
```

---

## TAG Integration

### TAG Block Format

Every SPEC document starts with a TAG block after the title:

```markdown
# @SPEC:AUTH-001: JWT Authentication System
```

### TAG Chain References

Link related TAGs in the SPEC:

```markdown
## Traceability (@TAG Chain)

### TAG Chain Structure
```
@SPEC:AUTH-001 (this document)
  ↓
@TEST:AUTH-001 (tests/auth/service.test.ts)
  ↓
@CODE:AUTH-001 (src/auth/service.ts, src/auth/token-manager.ts)
  ↓
@DOC:AUTH-001 (docs/api/authentication.md)
```

### Verification Commands
```bash
# Verify SPEC TAG
rg '@SPEC:AUTH-001' -n .moai/specs/

# Check duplicate IDs
rg '@SPEC:AUTH' -n .moai/specs/
rg 'AUTH-001' -n

# Full TAG chain scan
rg '@(SPEC|TEST|CODE|DOC):AUTH-001' -n
```
```

---

## Pre-Submission Validation Checklist

### Metadata Validation

```bash
# Check all required fields present
rg "^(id|version|status|created|updated|author|priority):" .moai/specs/SPEC-AUTH-001/spec.md

# Verify author format (@Handle)
rg "^author: @[A-Z]" .moai/specs/SPEC-AUTH-001/spec.md

# Verify version format (0.x.y)
rg "^version: 0\.\d+\.\d+" .moai/specs/SPEC-AUTH-001/spec.md

# Check duplicate IDs
rg "@SPEC:AUTH-001" -n .moai/specs/
```

### Content Validation

- [ ] **YAML Front Matter**: All 7 required fields present
- [ ] **TAG Block**: `@SPEC:{ID}` in title matches metadata `id`
- [ ] **HISTORY Section**: Present with v0.0.1 INITIAL entry
- [ ] **Environment Section**: Execution context, tech stack, constraints defined
- [ ] **Assumptions Section**: At least 3 key assumptions documented
- [ ] **Requirements Section**: All 5 EARS patterns used (or justified omissions)
- [ ] **Acceptance Criteria**: Testable criteria for each requirement
- [ ] **Traceability Section**: TAG chain structure documented

### EARS Syntax Validation

- [ ] **Ubiquitous**: Uses "shall" + capability
- [ ] **Event-driven**: Starts with "WHEN [trigger]"
- [ ] **State-driven**: Starts with "WHILE [state]"
- [ ] **Optional**: Starts with "WHERE [feature]", uses "can" not "shall"
- [ ] **Constraints**: Uses "IF-THEN" or direct constraint statement

### Quality Checks

- [ ] **No duplicate IDs**: `rg "AUTH-001" -n` returns only this SPEC
- [ ] **Consistent naming**: Directory name matches `SPEC-{ID}`
- [ ] **Complete HISTORY**: Each version has date, author, description
- [ ] **Clear requirements**: Each requirement is testable and unambiguous
- [ ] **Proper versioning**: Follows semantic versioning lifecycle

---

## Common Pitfalls & Prevention

### Pitfall 1: Changing ID After Assignment

**Problem**: Changing `id: AUTH-001` to `id: AUTH-002` after SPEC is created

**Impact**:
- Breaks TAG chain references
- Orphans existing code/tests
- Git history loss

**Prevention**:
```bash
# Always check for existing IDs before assignment
rg "@SPEC:AUTH" -n .moai/specs/
rg "AUTH-001" -n

# IDs are IMMUTABLE once assigned
```

### Pitfall 2: Skipping HISTORY Updates

**Problem**: Updating SPEC content without adding HISTORY entry

**Impact**:
- Lost change rationale
- Unclear version progression
- Audit trail gaps

**Prevention**:
```markdown
# ALWAYS update HISTORY when changing content
## HISTORY

### v0.0.2 (2025-10-25)  ← New entry
- **REFINED**: Added XYZ requirement
- **AUTHOR**: @YourHandle

### v0.0.1 (2025-10-23)
- **INITIAL**: First draft
```

### Pitfall 3: Wrong Version Number Progression

**Problem**: Jumping from v0.0.1 to v1.0.0 without intermediate versions

**Impact**:
- Version lifecycle broken
- Unclear maturity signal
- Stakeholder confusion

**Prevention**:
```markdown
# Follow version lifecycle strictly
v0.0.1 → v0.0.2 → ... → v0.1.0 → v0.1.1 → ... → v1.0.0
(draft)  (draft)       (completed)  (patches)     (stable)

# Never skip stages without justification in HISTORY
```

### Pitfall 4: Vague EARS Requirements

**Problem**: "The system shall be fast and user-friendly"

**Impact**:
- Untestable requirements
- Ambiguous acceptance criteria
- Implementation confusion

**Prevention**:
```markdown
# Bad
**UR-001**: The system shall be fast.

# Good
**UR-001**: The system shall respond to API requests within 200ms for 95% of requests.

# Bad
**ER-001**: WHEN user logs in, system should work.

# Good
**ER-001**: WHEN a user submits valid credentials, the system shall issue a JWT token with 15-minute expiry.
```

### Pitfall 5: Missing Author `@` Prefix

**Problem**: `author: Goos` instead of `author: @Goos`

**Impact**:
- Validation failures
- Inconsistent attribution
- GitHub linking broken

**Prevention**:
```yaml
# Wrong
author: Goos
author: goos

# Correct
author: @Goos
```

### Pitfall 6: Duplicate SPEC IDs

**Problem**: Creating `SPEC-AUTH-001` when it already exists

**Impact**:
- TAG chain collisions
- Ambiguous references
- Merge conflicts

**Prevention**:
```bash
# ALWAYS check before creating new SPEC
rg "@SPEC:AUTH-001" -n .moai/specs/
rg "AUTH-001" -n

# Use Explore agent to find similar SPECs
# Increment number if domain exists
```

### Pitfall 7: Mixing EARS Patterns

**Problem**: "WHEN user is logged in, WHILE session is active, the system shall..."

**Impact**:
- Confusing logic
- Hard to test
- Maintenance burden

**Prevention**:
```markdown
# Bad (mixed patterns)
**ER-001**: WHEN user logs in, WHILE session is active, the system shall allow access.

# Good (separate requirements)
**ER-001**: WHEN a user logs in successfully, the system shall create a session.
**SR-001**: WHILE a session is active, the system shall allow access to protected resources.
```

---

## Integration with `/alfred:1-plan`

### Workflow Handoff

When `/alfred:1-plan` is invoked, the `spec-builder` agent uses this Skill to:

1. **Analyze** user request and project context
2. **Generate** SPEC candidates with proper structure
3. **Validate** metadata completeness
4. **Create** `.moai/specs/SPEC-{ID}/spec.md` with EARS requirements
5. **Initialize** Git workflow (feature branch, Draft PR)

### spec-builder Integration Points

```markdown
Phase 1: SPEC Candidate Generation
  ↓ (uses moai-spec-authoring for metadata structure)
Phase 2: User Approval
  ↓
Phase 3: SPEC File Creation
  ↓ (applies EARS templates from this Skill)
Phase 4: Git Workflow Initialization
  ↓
Phase 5: Handoff to /alfred:2-run
```

### Agent Collaboration

- **spec-builder**: Generates SPEC using this Skill's templates
- **tag-agent**: Validates TAG format and uniqueness
- **trust-checker**: Verifies metadata completeness
- **git-manager**: Creates feature branch and Draft PR

---

## Validation Commands

### Quick Validation Script

```bash
#!/usr/bin/env bash
# validate-spec.sh - SPEC validation helper

SPEC_DIR="$1"

echo "Validating SPEC: $SPEC_DIR"

# Check required fields
echo -n "Required fields... "
rg "^(id|version|status|created|updated|author|priority):" "$SPEC_DIR/spec.md" | wc -l | grep -q "7" && echo "✅" || echo "❌"

# Check author format
echo -n "Author format... "
rg "^author: @[A-Z]" "$SPEC_DIR/spec.md" > /dev/null && echo "✅" || echo "❌"

# Check version format
echo -n "Version format... "
rg "^version: 0\.\d+\.\d+" "$SPEC_DIR/spec.md" > /dev/null && echo "✅" || echo "❌"

# Check HISTORY section
echo -n "HISTORY section... "
rg "^## HISTORY" "$SPEC_DIR/spec.md" > /dev/null && echo "✅" || echo "❌"

# Check TAG block
echo -n "TAG block... "
rg "^# @SPEC:" "$SPEC_DIR/spec.md" > /dev/null && echo "✅" || echo "❌"

# Check duplicate IDs
SPEC_ID=$(basename "$SPEC_DIR" | sed 's/SPEC-//')
DUPLICATE_COUNT=$(rg "@SPEC:$SPEC_ID" -n .moai/specs/ | wc -l)
echo -n "Duplicate ID check... "
[ "$DUPLICATE_COUNT" -eq 1 ] && echo "✅" || echo "❌ (found $DUPLICATE_COUNT occurrences)"

echo "Validation complete!"
```

### Usage

```bash
# Validate single SPEC
./validate-spec.sh .moai/specs/SPEC-AUTH-001

# Validate all SPECs
for spec in .moai/specs/SPEC-*/; do
  ./validate-spec.sh "$spec"
done
```

---

## Advanced Patterns

### Pattern: Versioned Requirements

When a requirement changes across versions, document the evolution:

```markdown
### v0.2.0 (2025-11-15)
**UR-001** (CHANGED): The system shall respond within 200ms for 99% of requests.
  - Previous (v0.1.0): 95% of requests
  - Rationale: Performance improvement based on user feedback

### v0.1.0 (2025-10-30)
**UR-001**: The system shall respond within 200ms for 95% of requests.
```

### Pattern: Requirement Traceability Matrix

Link requirements to test cases explicitly:

```markdown
## Requirements Traceability Matrix

| Req ID | Description | Test Cases | Status |
|--------|-------------|------------|--------|
| UR-001 | JWT authentication | test_authenticate_valid_user | ✅ |
| ER-001 | Token issuance | test_token_generation | ✅ |
| ER-002 | Token expiry | test_expired_token_rejection | ✅ |
| SR-001 | Authenticated access | test_protected_route_access | ✅ |
| C-001 | Token lifetime | test_token_expiry_constraint | ✅ |
```

### Pattern: Decision Log

Document architectural decisions within SPEC:

```markdown
## Decision Log

### Decision 1: JWT vs Session Cookies (2025-10-23)
**Context**: Need stateless authentication for microservices
**Decision**: Use JWT tokens
**Alternatives Considered**: 
  - Session cookies (rejected: stateful, not scalable)
  - OAuth 2.0 (deferred: too complex for MVP)
**Consequences**: 
  - ✅ Stateless, scalable
  - ✅ Cross-service authentication
  - ❌ Token revocation complexity

### Decision 2: Token Expiry 15min (2025-10-24)
**Context**: Balance security vs UX
**Decision**: 15-minute access token, 7-day refresh token
**Rationale**: Industry standard, security best practice
**References**: OWASP JWT best practices
```

---

## Examples Gallery

### Example 1: Complete Minimal SPEC

```markdown
---
id: HELLO-001
version: 0.0.1
status: draft
created: 2025-10-23
updated: 2025-10-23
author: @Goos
priority: low
---

# @SPEC:HELLO-001: Hello World API

## HISTORY

### v0.0.1 (2025-10-23)
- **INITIAL**: Hello World API SPEC draft created
- **AUTHOR**: @Goos

## Environment

**Runtime**: Node.js 20.x
**Framework**: Express.js

## Assumptions

1. Single endpoint required
2. No authentication needed
3. JSON response format

## Requirements

### Ubiquitous Requirements

**UR-001**: The system shall provide a GET /hello endpoint.

### Event-driven Requirements

**ER-001**: WHEN a GET request is sent to /hello, the system shall return JSON `{"message": "Hello, World!"}`.

### Constraints

**C-001**: Response time shall not exceed 50ms.
```

### Example 2: Production-Grade SPEC

See `SPEC-BRAND-001` or `SPEC-TRUST-001` in `.moai/specs/` for comprehensive examples with all optional fields, dependency graphs, and detailed HISTORY sections.

---

## Troubleshooting

### Issue: "Duplicate SPEC ID detected"

**Symptoms**: `rg "@SPEC:AUTH-001" -n` returns multiple results

**Solution**:
```bash
# Find all occurrences
rg "@SPEC:AUTH-001" -n .moai/specs/

# Choose one SPEC to keep, rename others
# Update TAG references in code/tests
rg '@SPEC:AUTH-001' -l src/ tests/ | xargs sed -i 's/@SPEC:AUTH-001/@SPEC:AUTH-002/g'
```

### Issue: "Version number doesn't match status"

**Symptoms**: `status: completed` but `version: 0.0.1`

**Solution**:
```yaml
# Update version to reflect completion
version: 0.1.0  # Implementation complete
status: completed
```

### Issue: "HISTORY section missing version"

**Symptoms**: Content changed but no new HISTORY entry

**Solution**:
```markdown
## HISTORY

### v0.0.2 (2025-10-25)  ← Add new entry
- **REFINED**: Updated XYZ requirement
- **AUTHOR**: @YourHandle

### v0.0.1 (2025-10-23)
- **INITIAL**: First draft
```

---

## References

### Internal Documentation

- `.moai/memory/spec-metadata.md` - Complete metadata reference
- `.moai/memory/development-guide.md` - SPEC-First TDD workflow
- `moai-foundation-ears` Skill - EARS syntax guide
- `moai-foundation-specs` Skill - Metadata validation

### External Resources

- EARS Official: Mavin, Wilkinson, et al. (2009) "Easy Approach to Requirements Syntax"
- QRA Guide: https://qracorp.com/guides_checklists/the-easy-approach-to-requirements-syntax-ears/
- Jama Software: EARS Notation Guide

---

## Changelog

- **v1.0.0** (2025-10-23): Initial comprehensive SPEC authoring Skill
  - 7 required + 9 optional metadata fields
  - 5 EARS patterns with examples
  - Version lifecycle guide
  - Pre-submission validation checklist
  - Common pitfalls prevention
  - Integration with `/alfred:1-plan`

---

## Works Well With

- `moai-foundation-ears` - EARS syntax patterns
- `moai-foundation-specs` - Metadata validation
- `moai-foundation-tags` - TAG system integration
- `moai-alfred-spec-metadata-validation` - Automated validation
- `spec-builder` agent - SPEC generation workflow

---

## Best Practices Summary

✅ **DO**:
- Check for duplicate IDs before creating new SPEC
- Update HISTORY with every content change
- Follow version lifecycle strictly (0.0.1 → 0.1.0 → 1.0.0)
- Use `@` prefix for author field
- Write testable, measurable requirements
- Include all 7 required metadata fields
- Use EARS patterns consistently

❌ **DON'T**:
- Change SPEC ID after assignment
- Skip HISTORY updates
- Jump version numbers without justification
- Write vague requirements ("fast", "user-friendly")
- Mix multiple EARS patterns in one requirement
- Forget to validate before submission
- Create duplicate SPEC IDs

---

**Last Updated**: 2025-10-23
**Maintained By**: MoAI-ADK Team
**Support**: Invoke `/alfred:1-plan` for guided SPEC creation
