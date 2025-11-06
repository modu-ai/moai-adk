# SPEC Authoring Reference

## Complete Metadata Field Reference

### Seven Required Fields

#### 1. `id` – Unique SPEC Identifier

**Format**: `<DOMAIN>-<NUMBER>`

**Rules**:
- Immutable after assignment
- Use uppercase domain names (e.g., `AUTH`, `PAYMENT`, `CONFIG`)
- Three-digit numbers (001–999)
- Check for duplicates: `rg "@SPEC:AUTH-001" -n .moai/specs/`

**Examples**:
- `AUTH-001` (authentication)
- `INSTALLER-SEC-001` (installer security)
- `TRUST-001` (TRUST principles)
- `CONFIG-001` (configuration schema)

**Directory Structure**:
```
.moai/specs/SPEC-AUTH-001/
  ├── spec.md          # Main SPEC document
  ├── diagrams/        # Optional: architecture diagrams
  └── examples/        # Optional: code examples
```

#### 2. `version` – Semantic Versioning

**Format**: `MAJOR.MINOR.PATCH`

**Lifecycle**:

| Version | Status | Description | Trigger |
|---------|--------|-------------|---------|
| `0.0.1` | draft | Initial draft | SPEC creation |
| `0.0.x` | draft | Draft improvements | Content editing |
| `0.1.0` | completed | Implementation complete | TDD finished + `/alfred:3-sync` |
| `0.1.x` | completed | Bug fixes, doc updates | Post-implementation patches |
| `0.x.0` | completed | Feature additions | Minor enhancements |
| `1.0.0` | completed | Production stable | Stakeholder approval |

**Version Update Example**:
```markdown
## HISTORY

### v0.2.0 (2025-11-15)
- **ADDED**: Multi-factor authentication support
- **CHANGED**: Token expiration extended from 15 to 30 minutes
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
- `draft`: Authoring phase (v0.0.x)
- `active`: Implementation in progress (v0.0.x → v0.1.0)
- `completed`: Implementation finished (v0.1.0+)
- `deprecated`: Marked for removal

#### 4. `created` – Creation Date

**Format**: `YYYY-MM-DD`

**Rules**:
- Set once, never changed
- ISO 8601 date format
- Initial draft date

**Example**: `created: 2025-10-29`

#### 5. `updated` – Last Modified Date

**Format**: `YYYY-MM-DD`

**Rules**:
- Update on every content change
- Initially same as `created`
- Reflects latest edit date

**Update Pattern**:
```yaml
created: 2025-10-29   # Never change
updated: 2025-10-31   # Update on edit
```

#### 6. `author` – Primary Author

**Format**: `@{GitHubHandle}`

**Rules**:
- Single value (not an array)
- @ prefix required
- Case-sensitive (e.g., `@Goos`, not `@goos`)
- Additional contributors documented in HISTORY section

**Examples**:
```yaml
# Correct
author: @Goos

# Incorrect
author: goos           # Missing @
authors: [@Goos]       # Array not allowed
author: @goos          # Case mismatch
```

#### 7. `priority` – Task Priority

**Values**: `critical` | `high` | `medium` | `low`

**Guidelines**:

| Priority | Description | Examples |
|----------|-------------|----------|
| `critical` | Production blocker, security vulnerability | Security patches, critical bugs |
| `high` | Major features, core functionality | Authentication, payment systems |
| `medium` | Improvements, enhancements | UI polish, performance optimization |
| `low` | Nice-to-have, documentation | README updates, minor refactoring |

---

### Nine Optional Fields

#### 8. `category` – Change Type

**Values**: `feature` | `bugfix` | `refactor` | `security` | `docs` | `perf`

**Usage**:
```yaml
category: feature       # New capability
category: bugfix        # Defect resolution
category: refactor      # Code structure improvement
category: security      # Security enhancement
category: docs          # Documentation update
category: perf          # Performance optimization
```

#### 9. `labels` – Classification Tags

**Format**: String array

**Purpose**: Search, filtering, grouping

**Best Practices**:
- Use lowercase, kebab-case
- 2-5 labels per SPEC
- Avoid duplication with `category`

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

**Meaning**: SPECs that must complete first

**Example**:
```yaml
depends_on:
  - USER-001      # User model SPEC
  - TOKEN-001     # Token generation SPEC
```

**Use Case**: Execution order, parallelization decisions

##### `blocks` – Blocked SPECs

**Meaning**: SPECs blocked until this SPEC completes

**Example**:
```yaml
blocks:
  - AUTH-002      # OAuth integration waits for base auth
  - PAYMENT-001   # Payment requires authentication
```

##### `related_specs` – Related SPECs

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

##### `scope.packages` – Affected Packages

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

**Purpose**: Reference principal implementation files

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

### Five EARS Patterns

EARS (Easy Approach to Requirements Syntax) uses familiar keywords to provide systematic, testable requirements.

#### Pattern 1: Ubiquitous Requirements

**Template**: `The system shall [capability].`

**Purpose**: Always-active base functionality

**Characteristics**:
- No preconditions
- Always applicable
- Core functionality definition

**Examples**:
```markdown
**UR-001**: The system shall provide user authentication.

**UR-002**: The system shall support HTTPS connections.

**UR-003**: The system shall securely store user credentials.

**UR-004**: The mobile app size shall not exceed 50 MB.

**UR-005**: API response time shall not exceed 200ms for 95% of requests.
```

**Best Practices**:
- ✅ Use active voice
- ✅ Single responsibility per requirement
- ✅ Measurable outcomes
- ❌ Avoid ambiguous terms ("user-friendly", "fast")

#### Pattern 2: Event-driven Requirements

**Template**: `WHEN [trigger], the system shall [response].`

**Purpose**: Behavior triggered by specific events

**Characteristics**:
- Triggered by discrete events
- Single-shot responses
- Cause-effect relationships

**Examples**:
```markdown
**ER-001**: WHEN the user submits valid credentials, the system shall issue a JWT token.

**ER-002**: WHEN a token expires, the system shall return HTTP 401 Unauthorized.

**ER-003**: WHEN the user clicks "Forgot Password", the system shall send a password reset email.

**ER-004**: WHEN database connection fails, the system shall retry 3 times with exponential backoff.

**ER-005**: WHEN file upload exceeds 10 MB, the system shall reject the upload with an error message.
```

**Advanced Pattern** (with postconditions):
```markdown
**ER-006**: WHEN payment transaction completes, the system shall send confirmation email, then update order status to "paid".
```

**Best Practices**:
- ✅ Single trigger per requirement
- ✅ Concrete, testable response
- ✅ Include error conditions
- ❌ Avoid chaining multiple WHEN clauses

#### Pattern 3: State-driven Requirements

**Template**: `WHILE [state], the system shall [behavior].`

**Purpose**: Persistent behavior during state

**Characteristics**:
- Active while state persists
- Continuous monitoring
- State-dependent behavior

**Examples**:
```markdown
**SR-001**: WHILE the user is in an authenticated state, the system shall permit access to protected routes.

**SR-002**: WHILE a token is valid, the system shall extract user ID from token claims.

**SR-003**: WHILE the system is in maintenance mode, the system shall return HTTP 503 Service Unavailable.

**SR-004**: WHILE battery level is below 20%, the mobile app shall reduce background sync frequency.

**SR-005**: WHILE file upload is in progress, the UI shall display a progress bar.
```

**Best Practices**:
- ✅ Clearly define state boundaries
- ✅ Specify state entry/exit conditions
- ✅ Test state transitions
- ❌ Avoid overlapping states

#### Pattern 4: Optional Features

**Template**: `WHERE [feature], the system can [behavior].`

**Purpose**: Feature-flag-based conditional functionality

**Characteristics**:
- Only applies when feature exists
- Configuration-dependent
- Product variant support

**Examples**:
```markdown
**OF-001**: WHERE multi-factor authentication is enabled, the system can require OTP verification after password confirmation.

**OF-002**: WHERE session logging is enabled, the system can record login timestamp and IP address.

**OF-003**: WHERE premium subscription is enabled, the system can permit unlimited API calls.

**OF-004**: WHERE dark mode is selected, the UI can render in dark color scheme.

**OF-005**: WHERE analytics consent is granted, the system can track user behavior.
```

**Best Practices**:
- ✅ Use "can" (permissive) not "shall" (mandatory)
- ✅ Clearly define feature flag condition
- ✅ Specify default behavior without feature
- ❌ Don't make core functionality optional

#### Pattern 5: Unwanted Behaviors

**Template**: `IF [condition], THEN the system shall [respond appropriately].`

**Purpose**: Error handling, quality gates, business rule enforcement

**Characteristics**:
- Conditional enforcement
- Quality gates and constraints
- Business rule validation

**Examples**:
```markdown
**UB-001**: IF a token has expired, THEN the system shall deny access and return HTTP 401.

**UB-002**: IF 5 or more login failures occur within 10 minutes, THEN the system shall temporarily lock the account.

**UB-003**: Response processing time shall not exceed 5 seconds.

**UB-004**: IF password length is less than 8 characters, THEN the system shall reject registration.

**UB-005**: IF API rate limit is exceeded, THEN the system shall return HTTP 429 Too Many Requests.
```

**Simplified Constraints** (no condition):
```markdown
**UB-006**: The system shall never store passwords in plaintext.

**UB-007**: All API endpoints except /health and /login shall require authentication.
```

**Best Practices**:
- ✅ Use SHALL for strict constraints, SHOULD for recommendations
- ✅ Quantify limits (time, size, count)
- ✅ Specify enforcement mechanism
- ❌ Avoid vague constraints

---

## EARS Pattern Selection Guide

| Pattern | Keyword | Use When | Context Example |
|---------|---------|----------|-----------------|
| **Ubiquitous** | shall | Core feature, always active | "System shall provide login" |
| **Event-driven** | WHEN | Response to specific event | "WHEN login fails, show error" |
| **State-driven** | WHILE | Continuous during state | "WHILE logged in, allow access" |
| **Optional** | WHERE | Feature flag or config | "WHERE premium enabled, unlock" |
| **Unwanted Behaviors** | IF-THEN | Error handling, quality gates | "IF expired, deny and return 401" |

---

## HISTORY Section Format

The HISTORY section documents all SPEC versions and changes.

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
| **REFINED** | Content update during draft | `v0.0.2: REFINED requirements based on review` |
| **COMPLETED** | Implementation finished | `v0.1.0: COMPLETED TDD implementation` |
| **ADDED** | New requirements/features | `v0.2.0: ADDED multi-factor authentication` |
| **CHANGED** | Modified requirement | `v0.2.0: CHANGED token expiration 15→30 minutes` |
| **FIXED** | Post-implementation bug fix | `v0.1.1: FIXED token refresh race condition` |
| **DEPRECATED** | Mark for removal | `v1.5.0: DEPRECATED legacy auth endpoint` |

### Complete HISTORY Example

```markdown
## HISTORY

### v0.2.0 (2025-11-15)
- **ADDED**: Multi-factor authentication via OTP
- **CHANGED**: Token expiration extended to 30 minutes based on user feedback
- **AUTHOR**: @Goos
- **REVIEWER**: @SecurityTeam
- **RATIONALE**: Maintain security posture while improving UX

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
- **CONTEXT**: Q4 2025 product roadmap requirements
```

---

## TAG Integration

### TAG Block Format

All SPEC documents start with a TAG block after the title:

```markdown
# @SPEC:AUTH-001: JWT Authentication System
```

### TAG Chain Reference

Link related TAGs in SPEC:

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

### Validation Commands
```bash
# Validate SPEC TAG
rg '@SPEC:AUTH-001' -n .moai/specs/

# Check for duplicate IDs
rg '@SPEC:AUTH' -n .moai/specs/
rg 'AUTH-001' -n

# Scan full TAG chain
rg '@(SPEC|TEST|CODE|DOC):AUTH-001' -n
```
```

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

# Check for duplicate IDs
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

**Last Updated**: 2025-10-29
**Version**: 1.2.0
