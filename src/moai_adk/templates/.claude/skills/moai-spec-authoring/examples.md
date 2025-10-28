# SPEC Authoring Examples

## Real-World EARS Examples

### Example 1: E-commerce Checkout

```markdown
### Ubiquitous Requirements
**UR-001**: The system shall provide a shopping cart feature.
**UR-002**: The system shall support credit card payment.

### Event-driven Requirements
**ER-001**: WHEN the user adds an item to the cart, the system shall update the cart total.
**ER-002**: WHEN payment succeeds, the system shall send a confirmation email.
**ER-003**: WHEN inventory is insufficient, the system shall display an "Out of Stock" message.

### State-driven Requirements
**SR-001**: WHILE items exist in the cart, the system shall reserve inventory for 30 minutes.
**SR-002**: WHILE payment is being processed, the UI shall display a loading indicator.

### Optional Features
**OF-001**: WHERE express shipping is selected, the system can calculate express shipping cost.
**OF-002**: WHERE gift wrapping is available, the system can offer gift wrapping option.

### Unwanted Behaviors
**UB-001**: IF cart total is less than $50, THEN the system shall add a $5 shipping fee.
**UB-002**: IF 3 payment failures occur, THEN the system shall lock the order for 1 hour.
**UB-003**: Order processing time shall not exceed 5 seconds.
```

### Example 2: Mobile App Push Notifications

```markdown
### Ubiquitous Requirements
**UR-001**: The app shall support push notifications.
**UR-002**: The app shall allow users to enable/disable notifications.

### Event-driven Requirements
**ER-001**: WHEN a new message arrives, the app shall display a push notification.
**ER-002**: WHEN the user taps a notification, the app shall navigate to the message screen.
**ER-003**: WHEN notification permission is denied, the app shall display an in-app banner.

### State-driven Requirements
**SR-001**: WHILE the app is in foreground state, the system shall display in-app banner instead of push notification.
**SR-002**: WHILE Do Not Disturb mode is enabled, the system shall mute all notifications.

### Optional Features
**OF-001**: WHERE notification sound is enabled, the system can play notification sound.
**OF-002**: WHERE notification grouping is supported, the system can group notifications by conversation.

### Unwanted Behaviors
**UB-001**: IF 10 or more notifications are pending, THEN the system shall consolidate them into a summary notification.
**UB-002**: Notification delivery latency shall not exceed 5 seconds.
```

---

## Complete SPEC Examples

### Example 1: Minimal SPEC

```markdown
---
id: HELLO-001
version: 0.0.1
status: draft
created: 2025-10-29
updated: 2025-10-29
author: @Goos
priority: low
---

# @SPEC:HELLO-001: Hello World API

## HISTORY

### v0.0.1 (2025-10-29)
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

### Unwanted Behaviors

**UB-001**: Response time shall not exceed 50ms.
```

### Example 2: Production-Grade SPEC

```markdown
---
id: AUTH-001
version: 0.1.0
status: completed
created: 2025-10-29
updated: 2025-10-30
author: @Goos
priority: high
category: feature
labels:
  - authentication
  - jwt
  - security
depends_on:
  - USER-001
  - TOKEN-001
blocks:
  - AUTH-002
  - PAYMENT-001
related_issue: "https://github.com/modu-ai/moai-adk/issues/42"
scope:
  packages:
    - src/core/auth
    - src/core/token
    - src/api/routes/auth
  files:
    - auth-service.ts
    - token-manager.ts
    - auth.routes.ts
---

# @SPEC:AUTH-001: JWT Authentication System

## HISTORY

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

### v0.0.1 (2025-10-29)
- **INITIAL**: JWT authentication SPEC draft created
- **AUTHOR**: @Goos
- **SCOPE**: User authentication, token generation, token validation
- **CONTEXT**: Q4 2025 product roadmap requirements

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

1. **User Storage**: User credentials are stored in PostgreSQL
2. **Secret Management**: JWT secrets are managed via environment variables
3. **Clock Sync**: Server clock is synchronized with NTP
4. **Password Policy**: Minimum 8 characters enforced during registration

## Requirements

### Ubiquitous Requirements

**UR-001**: The system shall provide JWT-based authentication.

**UR-002**: The system shall support user login with email and password.

**UR-003**: The system shall issue both access and refresh tokens.

### Event-driven Requirements

**ER-001**: WHEN the user submits valid credentials, the system shall issue a JWT access token with 15-minute expiration.

**ER-002**: WHEN a token expires, the system shall return HTTP 401 Unauthorized.

**ER-003**: WHEN a refresh token is presented, the system shall issue a new access token if the refresh token is valid.

### State-driven Requirements

**SR-001**: WHILE the user is in an authenticated state, the system shall permit access to protected resources.

**SR-002**: WHILE a token is valid, the system shall extract the user ID from token claims.

### Optional Features

**OF-001**: WHERE multi-factor authentication is enabled, the system can require OTP verification after password confirmation.

**OF-002**: WHERE session logging is enabled, the system can record login timestamp and IP address.

### Unwanted Behaviors

**UB-001**: IF a token has expired, THEN the system shall deny access and return HTTP 401.

**UB-002**: IF 5 or more login failures occur within 10 minutes, THEN the system shall temporarily lock the account.

**UB-003**: Access token lifetime shall not exceed 15 minutes.

**UB-004**: Refresh token lifetime shall not exceed 7 days.

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

## Decision Log

### Decision 1: JWT vs Session Cookies (2025-10-29)
**Context**: Need stateless authentication for microservices
**Decision**: Use JWT tokens
**Alternatives Considered**:
  - Session cookies (rejected: stateful, not scalable)
  - OAuth 2.0 (deferred: too complex for MVP)
**Consequences**:
  - ✅ Stateless, scalable
  - ✅ Service-to-service authentication
  - ❌ Token revocation complexity

### Decision 2: Token Expiration 15 minutes (2025-10-30)
**Context**: Balance between security and UX
**Decision**: 15-minute access token, 7-day refresh token
**Rationale**: Industry standard, OWASP best practices
**References**: OWASP JWT best practices

## Requirements Traceability Matrix

| Req ID | Description | Test Cases | Status |
|--------|-------------|------------|--------|
| UR-001 | JWT authentication | test_authenticate_valid_user | ✅ |
| ER-001 | Token issuance | test_token_generation | ✅ |
| ER-002 | Token expiration | test_expired_token_rejection | ✅ |
| SR-001 | Authenticated access | test_protected_route_access | ✅ |
| UB-001 | Token lifetime | test_token_expiry_constraint | ✅ |
```

---

## Advanced Patterns

### Pattern 1: Versioned Requirements

Document requirement evolution across versions:

```markdown
### v0.2.0 (2025-11-15)
**UR-001** (CHANGED): The system shall respond within 200ms for 99% of requests.
  - Previous (v0.1.0): 95% of requests
  - Rationale: User feedback-driven performance improvement

### v0.1.0 (2025-10-30)
**UR-001**: The system shall respond within 200ms for 95% of requests.
```

### Pattern 2: Requirements Traceability Matrix

Explicitly link requirements to test cases:

```markdown
## Requirements Traceability Matrix

| Req ID | Description | Test Cases | Status |
|--------|-------------|------------|--------|
| UR-001 | JWT authentication | test_authenticate_valid_user | ✅ |
| ER-001 | Token issuance | test_token_generation | ✅ |
| ER-002 | Token expiration | test_expired_token_rejection | ✅ |
| SR-001 | Authenticated access | test_protected_route_access | ✅ |
| UB-001 | Token lifetime | test_token_expiry_constraint | ✅ |
```

### Pattern 3: Decision Log

Document architectural decisions within the SPEC:

```markdown
## Decision Log

### Decision 1: JWT vs Session Cookies (2025-10-29)
**Context**: Need stateless authentication for microservices
**Decision**: Use JWT tokens
**Alternatives Considered**:
  - Session cookies (rejected: stateful, not scalable)
  - OAuth 2.0 (deferred: too complex for MVP)
**Consequences**:
  - ✅ Stateless, scalable
  - ✅ Service-to-service authentication
  - ❌ Token revocation complexity

### Decision 2: Token Expiration 15 minutes (2025-10-30)
**Context**: Balance between security and UX
**Decision**: 15-minute access token, 7-day refresh token
**Rationale**: Industry standard, OWASP best practices
**References**: OWASP JWT best practices
```

---

## Troubleshooting

### Issue: "Duplicate SPEC ID detected"

**Symptom**: `rg "@SPEC:AUTH-001" -n` returns multiple results

**Resolution**:
```bash
# Find all occurrences
rg "@SPEC:AUTH-001" -n .moai/specs/

# Keep one SPEC, rename the other
# Update TAG references in code/tests
rg '@SPEC:AUTH-001' -l src/ tests/ | xargs sed -i 's/@SPEC:AUTH-001/@SPEC:AUTH-002/g'
```

### Issue: "Version number doesn't match status"

**Symptom**: `status: completed` but `version: 0.0.1`

**Resolution**:
```yaml
# Update version to reflect completion
version: 0.1.0  # Implementation completed
status: completed
```

### Issue: "HISTORY section missing version entry"

**Symptom**: Content changed but no new HISTORY entry

**Resolution**:
```markdown
## HISTORY

### v0.0.2 (2025-10-25)  ← Add new entry
- **REFINED**: XYZ requirement updated
- **AUTHOR**: @YourHandle

### v0.0.1 (2025-10-23)
- **INITIAL**: Initial draft
```

### Issue: "Author field missing @ prefix"

**Symptom**: `author: Goos` instead of `author: @Goos`

**Resolution**:
```yaml
# Incorrect
author: Goos
author: goos

# Correct
author: @Goos
```

### Issue: "EARS pattern mixing"

**Symptom**: "WHEN user logs in, WHILE session is active, the system shall..."

**Resolution**:
```markdown
# Bad (pattern mixing)
**ER-001**: WHEN user logs in, WHILE session is active, the system shall permit access.

# Good (separate requirements)
**ER-001**: WHEN user successfully logs in, the system shall create a session.
**SR-001**: WHILE session is active, the system shall permit access to protected resources.
```

---

## Best Practices Summary

### ✅ DO (Best Practices)

1. **Check for duplicate IDs before creating**
   ```bash
   rg "@SPEC:AUTH-001" -n .moai/specs/
   rg "AUTH-001" -n
   ```

2. **Update HISTORY on every content change**
   ```markdown
   ### v0.0.2 (2025-10-25)
   - **REFINED**: XYZ added
   - **AUTHOR**: @YourHandle
   ```

3. **Follow version lifecycle strictly**
   ```
   0.0.1 → 0.0.2 → ... → 0.1.0 → 0.1.1 → ... → 1.0.0
   (draft)  (draft)       (completed)  (patches)     (stable)
   ```

4. **Use @ prefix in author field**
   ```yaml
   author: @Goos  # Correct
   ```

5. **Write testable, measurable requirements**
   ```markdown
   # Good
   **UR-001**: API response time shall not exceed 200ms for 95% of requests.

   # Bad
   **UR-001**: The system should be fast.
   ```

6. **Include all 7 required metadata fields**
   ```yaml
   id: AUTH-001
   version: 0.0.1
   status: draft
   created: 2025-10-29
   updated: 2025-10-29
   author: @Goos
   priority: high
   ```

7. **Use EARS patterns consistently**

### ❌ DON'T (Anti-Patterns)

1. **Don't change SPEC ID after assignment**
   - Breaks TAG chain
   - Orphans existing code/tests
   - Loses Git history

2. **Don't skip HISTORY updates**
   - Loses change rationale
   - Unclear version progression
   - Audit trail gaps

3. **Don't jump version numbers without reason**
   ```markdown
   # Bad: 0.0.1 → 1.0.0
   # Good: 0.0.1 → 0.0.2 → ... → 0.1.0 → 1.0.0
   ```

4. **Don't write ambiguous requirements**
   - Avoid "fast", "user-friendly", "good"
   - Use measurable criteria

5. **Don't mix EARS patterns in one requirement**

6. **Don't skip validation before submission**
   ```bash
   ./validate-spec.sh .moai/specs/SPEC-AUTH-001
   ```

7. **Don't create duplicate SPEC IDs**

---

## Integration Workflow

### `/alfred:1-plan` Integration

When `/alfred:1-plan` is called, the `spec-builder` agent uses this Skill to:

1. **Analyze**: User request and project context
2. **Generate**: SPEC candidates with appropriate structure
3. **Validate**: Metadata completeness
4. **Create**: `.moai/specs/SPEC-{ID}/spec.md` with EARS requirements
5. **Initialize**: Git workflow (feature branch, Draft PR)

### spec-builder Integration Points

```markdown
Phase 1: SPEC candidate generation
  ↓ (uses moai-spec-authoring for metadata structure)
Phase 2: User approval
  ↓
Phase 3: SPEC file creation
  ↓ (applies EARS templates from this Skill)
Phase 4: Git workflow initialization
  ↓
Phase 5: Handoff to /alfred:2-run
```

### Agent Collaboration

- **spec-builder**: Creates SPEC using this Skill's templates
- **tag-agent**: Validates TAG format and uniqueness
- **trust-checker**: Verifies metadata completeness
- **git-manager**: Creates feature branch and Draft PR

---

**Last Updated**: 2025-10-29
**Version**: 1.2.0
