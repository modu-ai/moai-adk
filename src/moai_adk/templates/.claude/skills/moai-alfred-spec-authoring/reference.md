# SPEC Authoring Reference Guide

## YAML Metadata Field Reference

### Required Fields (7)

#### `code` (String)
- **Format**: SPEC-XXX (3+ digits)
- **Example**: SPEC-001, SPEC-105, SPEC-4072
- **Auto-generated**: By MoAI-ADK when creating new SPEC
- **Validation**: Must be unique across project
- **Immutable**: Yes, after creation

#### `title` (String)
- **Length**: 50-80 characters (recommended)
- **Style**: Title Case, descriptive, specific
- **Example**: ✓ "Email Notification Service with Template Engine"
- **Anti-pattern**: ✗ "Update feature", ✗ "New SPEC"
- **Immutable**: No (can be refined)

#### `status` (String: enum)
- **Valid Values**: `draft` | `active` | `deprecated` | `archived`
- **Lifecycle**: draft → active → deprecated → archived
- **Default**: draft
- **Constraints**:
  - Only `draft` status allows metadata changes
  - `deprecated` requires replacement SPEC code
  - `archived` is immutable

#### `created_at` (ISO 8601 Date)
- **Format**: YYYY-MM-DD
- **Example**: 2025-11-12
- **Immutable**: Yes, set at creation
- **Timezone**: UTC (Z suffix optional)

#### `updated_at` (ISO 8601 Date)
- **Format**: YYYY-MM-DD
- **Example**: 2025-11-12
- **Auto-updated**: By automation on content changes
- **Manual**: Developers should NOT update
- **Frequency**: Once per change (git commit)

#### `priority` (String: enum)
- **Valid Values**: `critical` | `high` | `medium` | `low`
- **Business Mapping**:
  - `critical`: Blocks release (security, data integrity)
  - `high`: Delivers business value, planned for sprint
  - `medium`: Nice-to-have, backlog priority
  - `low`: Technical debt, future consideration
- **Affects**: Sprint planning, resource allocation

#### `effort` (Integer: 1-13 Fibonacci)
- **Valid Values**: 1, 2, 3, 5, 8, 13
- **Estimation Scale**:
  - 1: Trivial (< 2 hours)
  - 2: Small (2-4 hours)
  - 3: Medium (4-8 hours)
  - 5: Large (1-2 days)
  - 8: Extra Large (2-3 days)
  - 13: Uncertain / Needs breakdown
- **Note**: If >13, break into multiple SPECs

---

### Optional Fields (9)

#### `version` (Semantic Versioning)
- **Format**: MAJOR.MINOR.PATCH
- **Example**: 1.0.0, 1.2.3, 2.0.0
- **Increment Rules**:
  - MAJOR: Breaking changes, API incompatibility
  - MINOR: New features, backward compatible
  - PATCH: Bug fixes, documentation
- **Initial**: 0.1.0 (draft), 1.0.0 (active)

#### `deadline` (ISO 8601 Date)
- **Format**: YYYY-MM-DD
- **Example**: 2025-12-15
- **Optional**: Use only if deadline exists
- **Affects**: Release planning

#### `epic` (String)
- **Format**: Epic code (e.g., AUTH-01, NOTIFICATIONS-01)
- **Purpose**: Group related SPECs by business capability
- **Example**: AUTH-01 (Authentication epic)
- **Usage**: Filter SPECs by epic, roadmap planning

#### `depends_on` (Array of SPEC codes)
- **Format**: [SPEC-XXX, SPEC-YYY, ...]
- **Meaning**: This SPEC requires completion of referenced SPECs
- **Validation**: Prevents circular dependencies
- **Usage**: Identify critical path dependencies
- **Example**: [SPEC-102, SPEC-104]

#### `domains` (Array of domain strings)
- **Valid Values**: backend, frontend, database, devops, security, infrastructure, etc.
- **Purpose**: Route SPEC to appropriate team
- **Example**: [backend, database, security]
- **Multi-value**: Most SPECs span 2-3 domains

#### `acceptance_difficulty` (String: enum)
- **Valid Values**: low | medium | high | critical
- **Definition**: How hard is it to verify acceptance criteria?
- **Affects**: QA resource planning
- **Example**: `high` (complex state management, many edge cases)

#### `rollback_risk` (String: enum)
- **Valid Values**: low | medium | high | critical
- **Definition**: Severity if this change must be reverted
- **Critical**: Data migration, breaking API, security changes
- **High**: Database schema change, configuration change
- **Example**: `high` (database migration is difficult to reverse)

#### `risks` (String: multi-line)
- **Format**: Bullet list of risk statements
- **Example**:
  ```
  - Security: JWT key rotation must be tested
  - Performance: Token validation on every request
  - Compatibility: Breaks existing auth tokens
  ```
- **Purpose**: Identify risks for planning mitigation

#### `tags` (Array of strings)
- **Format**: [tag1, tag2, ...]
- **Purpose**: Enable cross-cutting search and filtering
- **Example**: [authentication, security, jwt, users]
- **Naming**: lowercase, hyphenated

---

## EARS Requirement Syntax Reference

### Pattern 1: Universal Requirements

**When**: Core behavior, non-negotiable functionality

**Template**:
```
SPEC: The [System] SHALL [Action]
```

**Complete Syntax**:
```
SPEC-XXX-REQ-001: The [System Component] SHALL [Action Description]
[when applicable: under [Conditions]]
[expected outcome: to achieve [Objective]]
```

**Real Examples**:

```
SPEC-105-REQ-001: The notification service SHALL send emails asynchronously.

SPEC-001-REQ-002: The authentication service SHALL validate all JWT tokens
using RS256 algorithm against the published public key.

SPEC-204-REQ-003: The database adapter SHALL support connection pooling
to maintain optimal performance under load.
```

**Test Pattern**:
```python
def test_universal_requirement():
    """Verify the system SHALL [Action]."""
    # Arrange
    # Act
    # Assert
    assert <requirement_met>
```

---

### Pattern 2: Conditional Requirements

**When**: Behavior depends on specific conditions

**Template**:
```
SPEC: If [Condition], then the [System] SHALL [Action]
```

**Complete Syntax**:
```
SPEC-XXX-REQ-001: If [Condition Description],
then the [System Component] SHALL [Action Description]
[expected outcome: returning [Outcome]]
```

**Real Examples**:

```
SPEC-105-REQ-002: If email delivery fails, the notification service
SHALL retry up to 3 times with exponential backoff (1s, 2s, 4s)
before marking as failed.

SPEC-001-REQ-004: If a JWT token has expired, the authentication service
SHALL reject the request and return HTTP 401 Unauthorized
with error code TOKEN_EXPIRED.

SPEC-204-REQ-005: If connection pool reaches max capacity,
the adapter SHALL queue requests with a 30-second timeout.
```

**Test Pattern**:
```python
def test_conditional_requirement():
    """Verify: If [Condition], then SHALL [Action]."""
    # Arrange: Set up condition
    # Act: Trigger condition
    # Assert: Verify expected action
    assert <action_performed_correctly>
```

---

### Pattern 3: Unwanted Behavior (Negative Requirements)

**When**: Security constraints, forbidden operations, anti-patterns

**Template**:
```
SPEC: The [System] SHALL NOT [Forbidden Action]
```

**Complete Syntax**:
```
SPEC-XXX-REQ-001: The [System Component] SHALL NOT [Forbidden Action]
[reason: because [Security Risk | Performance Issue | Data Integrity Issue]]
```

**Real Examples**:

```
SPEC-001-REQ-006: The authentication service SHALL NOT accept JWT tokens
signed with symmetric algorithms (HS256, HS384, HS512)
in a production environment.

SPEC-105-REQ-003: The notification service SHALL NOT send duplicate emails
to the same recipient within a 5-minute window.

SPEC-204-REQ-007: The database adapter SHALL NOT store plaintext passwords
or connection credentials in logs or error messages.
```

**Test Pattern**:
```python
def test_unwanted_behavior():
    """Verify system SHALL NOT [Forbidden Action]."""
    # Arrange: Set up scenario for forbidden action
    # Act: Attempt forbidden action
    # Assert: Verify action is prevented/rejected
    with pytest.raises(ForbiddenError):
        <forbidden_action>
```

---

### Pattern 4: Stakeholder Requirements

**When**: User stories, feature requirements, stakeholder concerns

**Template**:
```
SPEC: As a [User Role], I want [Feature] so that [Benefit]
```

**Complete Syntax**:
```
SPEC-XXX-REQ-001: As a [User Role],
I want [Feature Description]
so that [Business Benefit or Outcome]
```

**Real Examples**:

```
SPEC-001-REQ-008: As an API consumer, I want to pass JWT tokens
in the Authorization header so that my requests are authenticated
without exposing tokens in URLs or query parameters.

SPEC-105-REQ-004: As a business user, I want to receive email notifications
for account activities so that I stay informed without checking the app constantly.

SPEC-204-REQ-009: As a database administrator, I want connection pooling
so that I can optimize resource utilization and reduce connection overhead.
```

**Test Pattern**:
```python
def test_stakeholder_requirement():
    """Verify: As [User], I want [Feature] so that [Benefit]."""
    # Arrange: Set up user context
    # Act: User performs action to get feature
    # Assert: Verify benefit is achieved
    assert <benefit_achieved>
```

---

### Pattern 5: Boundary Condition (Edge Cases)

**When**: Performance limits, resource constraints, edge cases

**Template**:
```
SPEC: [System] SHALL [Action] when [Boundary Condition]
```

**Complete Syntax**:
```
SPEC-XXX-REQ-001: The [System Component] SHALL [Action Description]
when [Boundary Condition: specific value, threshold, or edge case],
achieving [Performance Metric or Outcome]
```

**Real Examples**:

```
SPEC-105-REQ-005: The notification service SHALL process at least 1,000 emails/second
and SHALL NOT exceed 500MB memory usage under sustained load.

SPEC-001-REQ-010: The authentication service SHALL return HTTP 429 Too Many Requests
when a single IP address attempts more than 10 failed authentication attempts within 5 minutes.

SPEC-204-REQ-011: The database adapter SHALL maintain ≤50ms response time
when processing 100+ concurrent connections with query complexity score ≥7.
```

**Test Pattern**:
```python
@pytest.mark.benchmark
def test_boundary_condition(benchmark):
    """Verify system handles boundary [Condition]."""
    # Arrange: Set up boundary condition
    result = benchmark(<action_under_load>)
    # Assert: Verify metric is met
    assert result.p99_latency <= 50  # milliseconds
    assert result.memory_peak <= 500  # MB
```

---

## EARS Anti-Patterns to Avoid

### Anti-Pattern 1: Vague Language

**Bad**:
```
SPEC: The system should authenticate users securely.
SPEC: The service may retry failed requests.
SPEC: The API might return errors in some cases.
```

**Good**:
```
SPEC: The system SHALL authenticate users using JWT tokens.
SPEC: The service SHALL retry failed requests up to 3 times.
SPEC: The API SHALL return HTTP 400 Bad Request for invalid input.
```

**Rule**: Replace "should", "may", "might", "could" with "SHALL" or "SHALL NOT"

---

### Anti-Pattern 2: Unobservable Requirements

**Bad**:
```
SPEC: The system shall be user-friendly.
SPEC: The API shall be fast.
SPEC: Authentication shall be secure.
```

**Good**:
```
SPEC: The API response time SHALL be ≤200ms (p99) for common operations.
SPEC: Authentication SHALL use RS256 algorithm with key rotation every 90 days.
SPEC: The UI SHALL have <3 clicks to reach primary feature.
```

**Rule**: Requirements must be testable/measurable

---

### Anti-Pattern 3: Multiple Requirements in One Statement

**Bad**:
```
SPEC: The system SHALL support multiple authentication methods
and validate tokens securely.
```

**Good**:
```
SPEC-001-REQ-001: The system SHALL support JWT, OAuth2, and API key authentication.
SPEC-001-REQ-002: The system SHALL validate all authentication tokens
using cryptographic signatures before accepting requests.
```

**Rule**: One requirement = one observable behavior

---

## Validation Checklist (Pre-Submission)

### Metadata Validation

| Check | Pass | Fail |
|-------|------|------|
| `code` matches SPEC-XXX format | ✓ | Fix format |
| `title` is 50-80 characters | ✓ | Adjust length |
| `status` is valid enum value | ✓ | Use draft/active/deprecated/archived |
| `created_at` is ISO 8601 | ✓ | Use YYYY-MM-DD |
| `updated_at` is ISO 8601 | ✓ | Use YYYY-MM-DD |
| `priority` is valid enum | ✓ | Use critical/high/medium/low |
| `effort` is Fibonacci (1-13) | ✓ | Use 1,2,3,5,8,13 |
| Optional fields correct type | ✓ | Fix type/format |

### Requirement Syntax

| Check | Pass | Fail |
|-------|------|------|
| At least 3 EARS patterns used | ✓ | Add more patterns |
| Each REQ follows EARS template | ✓ | Rewrite using pattern |
| Requirements are testable | ✓ | Add measurable criteria |
| No vague language ("should", "may") | ✓ | Replace with SHALL/SHALL NOT |
| No compound requirements | ✓ | Split into separate REQs |
| Numbering is sequential | ✓ | Renumber REQ-001, REQ-002, ... |

### Unwanted Behaviors

| Check | Pass | Fail |
|-------|------|------|
| Security constraints listed | ✓ | Add if applicable |
| Performance constraints listed | ✓ | Add if applicable |
| Reliability constraints listed | ✓ | Add if applicable |
| Each UB has test approach | ✓ | Define how to verify non-occurrence |
| UBs are specific (not generic) | ✓ | Avoid "The system shall be safe" |

### Acceptance Criteria

| Check | Pass | Fail |
|-------|------|------|
| All 5 EARS patterns implemented | ✓ | Add missing patterns |
| All UBs have test cases | ✓ | Add test coverage |
| Code coverage target ≥85% | ✓ | Increase target if needed |
| Performance baseline defined | ✓ | Add metrics (latency, throughput) |
| Security scan scope defined | ✓ | Specify OWASP Top 10 coverage |

### TAG Integration

| Check | Pass | Fail |
|-------|------|------|

### Documentation Quality

| Check | Pass | Fail |
|-------|------|------|
| Overview is 2-3 sentences | ✓ | Shorten/clarify |
| Architecture impact explained | ✓ | Add design decisions |
| Database changes documented | ✓ | Add schema changes if applicable |
| Configuration parameters listed | ✓ | Add config details if applicable |
| Related SPECs referenced | ✓ | Check depends_on array |
| No TODO or placeholder text | ✓ | Complete all sections |

### Final Review

| Check | Pass | Fail |
|-------|------|------|
| Markdown formatting correct | ✓ | Fix syntax |
| All links are valid | ✓ | Test links |
| No confidential info exposed | ✓ | Remove sensitive details |
| Ready for team review | ✓ | Complete all checks above |

---

## Official References

### EARS Syntax Origins
- EARS Paper: https://ieeexplore.ieee.org/document/5586873
- Stanford EARS Guide: https://requirements.readthedocs.io/

### Requirement Best Practices
- IEEE 829 Standard: https://standards.ieee.org/standard/829-2024.html
- IREB Glossary: https://www.ireb.org/
- ISO/IEC/IEEE 29148: https://standards.ieee.org/standard/29148-2024.html

### YAML Specification
- YAML Official: https://yaml.org/
- YAML v1.2 Spec: https://yaml.org/spec/1.2/spec.html

### Testing Frameworks
- Pytest: https://docs.pytest.org/
- Jest: https://jestjs.io/
- Go Testing: https://pkg.go.dev/testing
- Rust Testing: https://doc.rust-lang.org/book/ch11-00-testing.html

---

## Summary

This reference provides:
- Complete YAML field documentation
- EARS pattern templates with examples
- Unwanted Behavior patterns
- Validation checklist (pre-submission)
- Anti-patterns to avoid
- Official documentation links

**Use when**: Authoring SPECs, reviewing requirements, or validating metadata.
