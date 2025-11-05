# SPEC Writing Basics

Learn how to write clear, executable specifications using the EARS (Easy Approach to Requirements Syntax) format. Good specifications are the foundation of successful software development.

## Overview

A SPEC (Specification) in MoAI-ADK is a structured document that defines what needs to be built, how it should behave, and what constraints apply. SPECs follow the EARS methodology to ensure clarity, testability, and completeness.

### What Makes a Good SPEC?

- **Unambiguous**: Clear language that leaves no room for interpretation
- **Testable**: Each requirement can be validated through testing
- **Complete**: Covers all aspects of the functionality
- **Traceable**: Linked to tests, code, and documentation through @TAGs
- **Maintainable**: Easy to understand and modify as requirements evolve

### The SPEC-First Principle

> "No code without specifications, no tests without clear requirements"

SPECs are written **before** any code is created, ensuring that everyone understands exactly what needs to be built and how success will be measured.

## SPEC Structure and Components

### YAML Frontmatter

Every SPEC begins with structured metadata:

```yaml
---
id: AUTH-001
version: 0.1.0
status: draft
priority: high
created: 2025-01-15
updated: 2025-01-15
author: @developer
domain: authentication
complexity: medium
estimated_hours: 8
dependencies: [USER-001, EMAIL-001]
tags: [api, security, jwt]
reviewers: [@tech-lead, @security-expert]
milestone: "Sprint 23 - Q1 2025"
---
```

**Field Descriptions**:

- **`id`**: Unique identifier in `DOMAIN-NNN` format
- **`version`**: Semantic version following `MAJOR.MINOR.PATCH`
- **`status`**: Current state (draft, in_review, approved, in_progress, completed, stable, deprecated)
- **`priority`**: Importance level (critical, high, medium, low)
- **`domain`**: Functional area (authentication, user_management, api, database, etc.)
- **`complexity`**: Implementation difficulty (simple, medium, complex, expert)
- **`estimated_hours`**: Time estimate for implementation
- **`dependencies`**: Other SPECs this depends on
- **`tags`**: Keywords for categorization and search
- **`reviewers`**: Team members who should review this SPEC
- **`milestone`**: Development milestone or sprint

### SPEC Content Sections

A complete SPEC includes these sections:

```markdown
# @SPEC:EX-AUTH-001: User Authentication System

## Overview
Brief description of what this SPEC covers and its importance.

## EARS Requirements
The core requirements written in EARS format.

## Technical Requirements
Non-functional requirements and technical constraints.

## Acceptance Criteria
Specific conditions that must be met for completion.

## Dependencies
Other components, systems, or SPECs this depends on.

## Risk Assessment
Potential challenges and mitigation strategies.

## Implementation Notes
Technical considerations and recommendations from experts.
```

## EARS Syntax: The 5 Patterns

EARS (Easy Approach to Requirements Syntax) provides 5 clear patterns for writing requirements. Each pattern serves a specific purpose and uses consistent language.

### 1. Ubiquitous Requirements

**Purpose**: Define basic functionality that must always be available.

**Pattern**: `The system SHALL <capability>`

**Examples**:
```markdown
- The system SHALL provide user authentication via email and password
- The system SHALL support JWT token-based session management
- The system SHALL validate user input before processing
- The system SHALL log all authentication attempts
- The system SHALL maintain user session state securely
```

**Characteristics**:
- Always active functionality
- Core system capabilities
- No conditions or triggers
- Essential for system operation

### 2. Event-driven Requirements

**Purpose**: Define system behavior in response to specific events or triggers.

**Pattern**: `WHEN <trigger occurs>, the system SHALL <response>`

**Examples**:
```markdown
- WHEN valid credentials are provided, the system SHALL issue access and refresh tokens
- WHEN invalid credentials are provided, the system SHALL return a 401 error
- WHEN a refresh token is valid, the system SHALL issue a new access token
- WHEN multiple failed login attempts occur, the system SHALL implement rate limiting
- WHEN a user logs out, the system SHALL invalidate all active tokens
```

**Characteristics**:
- Clear trigger-response relationship
- Event-driven behavior
- Specific conditions and outcomes
- Testable scenarios

### 3. State-driven Requirements

**Purpose**: Define behavior that depends on system state or conditions.

**Pattern**: `WHILE <condition exists>, the system SHALL <behavior>`

**Examples**:
```markdown
- WHILE a user is authenticated, the system SHALL allow access to protected resources
- WHILE a session is active, the system SHALL maintain user context
- WHILE rate limiting is active, the system SHALL reject excess requests
- WHILE maintenance mode is enabled, the system SHALL allow admin access only
- WHILE a password reset is in progress, the system SHALL block login attempts
```

**Characteristics**:
- Continuous behavior while condition exists
- State-dependent functionality
- Context-aware operations
- Ongoing system behavior

### 4. Optional Requirements

**Purpose**: Define nice-to-have features or conditional functionality.

**Pattern**: `WHERE <condition is met>, the system MAY <optional behavior>`

**Examples**:
```markdown
- WHERE multi-factor authentication is enabled, the system MAY require additional verification
- WHERE social login providers are configured, the system MAY support OAuth authentication
- WHERE device fingerprinting is available, the system MAY track login sessions by device
- WHERE user preferences allow, the system MAY remember login location
- WHERE analytics are enabled, the system MAY track authentication patterns
```

**Characteristics**:
- Conditional functionality
- Optional features
- Enhancement capabilities
- Configuration-dependent

### 5. Unwanted Behaviors (Constraints)

**Purpose**: Define what the system should NOT do and constraints that must be followed.

**Pattern**: `The system SHALL NOT <undesired behavior>` or `<parameter> SHALL NOT <constraint>`

**Examples**:
```markdown
- The system SHALL NOT store passwords in plain text
- The system SHALL NOT reveal whether an email address is registered
- Passwords SHALL NOT be less than 8 characters
- The system SHALL NOT allow concurrent sessions with the same credentials
- JWT tokens SHALL NOT expire after more than 24 hours
- Login attempts SHALL NOT exceed 5 per minute per IP address
- The system SHALL NOT accept special characters in usernames
```

**Characteristics**:
- Security constraints
- Business rules
- Technical limitations
- Quality requirements

## Writing Effective Requirements

### Requirement Quality Checklist

For each requirement, verify:

**Clarity**:
- [ ] Language is unambiguous
- [ ] Technical terms are defined
- [ ] Acronyms are explained
- [ ] Context is clear

**Testability**:
- [ ] Success criteria are defined
- [ ] Test scenarios can be created
- [ ] Expected outcomes are specified
- [ ] Edge cases are considered

**Completeness**:
- [ ] All functional aspects covered
- [ ] Error conditions included
- [ ] Performance requirements specified
- [ ] Security considerations addressed

**Consistency**:
- [ ] Language follows EARS patterns
- [ ] Terminology is consistent
- [ ] Requirements don't conflict
- [ ] Dependencies are identified

### Common Mistakes to Avoid

**<span class="material-icons">cancel</span> Vague Language**:
```
The system should handle user authentication well.
```

**<span class="material-icons">check_circle</span> Specific Language**:
```
The system SHALL authenticate users via email/password credentials within 500ms.
```

**<span class="material-icons">cancel</span> Multiple Requirements in One**:
```
WHEN users login, the system SHALL issue tokens and log the attempt and update the user profile.
```

**<span class="material-icons">check_circle</span> Separate Requirements**:
```
WHEN valid credentials are provided, the system SHALL issue JWT tokens.
WHEN authentication occurs, the system SHALL log the attempt with timestamp.
WHEN users authenticate successfully, the system SHALL update last login timestamp.
```

**<span class="material-icons">cancel</span> Implementation Details**:
```
The system SHALL use bcrypt with 12 rounds to hash passwords in the PostgreSQL database.
```

**<span class="material-icons">check_circle</span> Behavioral Requirements**:
```
The system SHALL hash passwords using a secure algorithm with minimum 12 rounds.
The system SHALL store password hashes securely in the database.
```

**<span class="material-icons">cancel</span> Missing Error Conditions**:
```
WHEN users login with valid credentials, the system SHALL issue tokens.
```

**<span class="material-icons">check_circle</span> Complete Coverage**:
```
WHEN valid credentials are provided, the system SHALL issue JWT tokens.
WHEN invalid credentials are provided, the system SHALL return a 401 error.
WHEN the authentication service is unavailable, the system SHALL return a 503 error.
```

## SPEC ID Assignment and Management

### ID Format and Rules

**Format**: `DOMAIN-NNN`

**Examples**: `AUTH-001`, `USER-002`, `API-003`, `DB-001`

### Domain Categories

| Domain | Description | Examples |
|--------|-------------|----------|
| **AUTH** | Authentication and authorization | Login, registration, permissions |
| **USER** | User management and profiles | Profile creation, user settings |
| **API** | REST API endpoints and interfaces | HTTP endpoints, request/response formats |
| **DB** | Database schemas and operations | Tables, queries, migrations |
| **UI** | User interface components | Forms, pages, interactions |
| **SEC** | Security features and controls | Encryption, auditing, compliance |
| **PERF** | Performance and optimization | Caching, load balancing, monitoring |
| **INT** | Integration with external systems | Third-party APIs, webhooks |
| **CONFIG** | Configuration and settings | Environment variables, feature flags |

### ID Assignment Process

1. **Select Domain**: Choose appropriate domain for the feature
2. **Check Existing IDs**: Find the next available number in that domain
3. **Assign ID**: Use format `DOMAIN-NNN` (e.g., `AUTH-001`)
4. **Record in Registry**: Update `.moai/specs/registry.json`

**Example Registry Entry**:
```json
{
  "AUTH-001": {
    "title": "User Authentication System",
    "status": "completed",
    "created": "2025-01-15",
    "assigned_to": "@developer"
  },
  "AUTH-002": {
    "title": "Password Reset Functionality",
    "status": "draft",
    "created": "2025-01-16",
    "assigned_to": "@developer"
  }
}
```

## Acceptance Criteria

### Writing Effective Acceptance Criteria

Acceptance criteria define when a SPEC is considered complete and working correctly.

#### Gherkin-style Format

```gherkin
Feature: User Authentication

Scenario: Successful login with valid credentials
  GIVEN a registered user with valid credentials
  WHEN the user submits correct email and password
  THEN the system SHALL authenticate the user
  AND the system SHALL issue JWT tokens
  AND the tokens SHALL be valid for 15 minutes

Scenario: Failed login with invalid credentials
  GIVEN a registered user account
  WHEN the user submits incorrect password
  THEN the system SHALL reject authentication
  AND the system SHALL return 401 error
  AND the system SHALL log the failed attempt
```

#### Checklist Format

```markdown
## Acceptance Criteria

### Functional Requirements
- [ ] Users can authenticate with email and password
- [ ] System issues JWT access tokens (15 min expiry)
- [ ] System issues JWT refresh tokens (7 days expiry)
- [ ] Invalid credentials return 401 error
- [ ] Rate limiting prevents brute force attacks

### Security Requirements
- [ ] Passwords are hashed with bcrypt (12+ rounds)
- [ ] JWT tokens use RS256 signing algorithm
- [ ] All endpoints use HTTPS only
- [ ] Input validation prevents injection attacks
- [ ] Authentication attempts are logged

### Performance Requirements
- [ ] Login response time < 500ms
- [ ] Token validation < 100ms
- [ ] System supports 1000 concurrent authentications
- [ ] Database queries are optimized

### Integration Requirements
- [ ] Integrates with user management system
- [ ] Sends email notifications for security events
- [ ] Connects to monitoring and alerting
- [ ] Supports multiple environments (dev/staging/prod)
```

### Testable Acceptance Criteria

Each acceptance criterion should be:

1. **Specific**: Clear and unambiguous
2. **Measurable**: Can be quantified or verified
3. **Achievable**: Realistic and attainable
4. **Relevant**: Aligns with business goals
5. **Time-bound**: Has clear completion criteria

## Dependencies and Relationships

### Types of Dependencies

**Functional Dependencies**:
```markdown
## Dependencies
- USER-001: User management system (required for authentication)
- EMAIL-001: Email service for notifications (required for password reset)
- RATE-001: Rate limiting service (required for security)
```

**Technical Dependencies**:
```markdown
## Technical Dependencies
- PostgreSQL database (version 13+)
- Redis for session storage
- SMTP server for email delivery
- SSL certificates for HTTPS
```

**External Dependencies**:
```markdown
## External Dependencies
- OAuth providers (Google, GitHub) for social login
- SMS service for two-factor authentication
- Monitoring service for security events
```

### Dependency Management

1. **Identify Dependencies**: List all required components
2. **Assess Impact**: Understand how dependencies affect implementation
3. **Plan Integration**: Define how dependencies will be integrated
4. **Document Interfaces**: Specify integration points and contracts
5. **Monitor Changes**: Track dependency updates and compatibility

## Risk Assessment

### Common Risk Categories

**Technical Risks**:
- Implementation complexity
- Performance bottlenecks
- Security vulnerabilities
- Integration challenges

**Business Risks**:
- Changing requirements
- Timeline constraints
- Resource availability
- User adoption

**Operational Risks**:
- Deployment challenges
- Monitoring gaps
- Maintenance overhead
- Scalability issues

### Risk Assessment Template

```markdown
## Risk Assessment

### High Risk
| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Token security implementation | High | Medium | Use established libraries, security review |
| Performance under load | Medium | High | Load testing, caching strategy |

### Medium Risk
| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Email service reliability | Medium | Medium | Multiple providers, fallback mechanism |
| Database schema changes | Low | High | Migration strategy, backward compatibility |

### Low Risk
| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| UI/UX design consistency | Low | Low | Design system, component library |
```

## SPEC Review Process

### Review Checklist

**Content Review**:
- [ ] Requirements are clear and unambiguous
- [ ] All use cases are covered
- [ ] Error conditions are specified
- [ ] Acceptance criteria are testable
- [ ] Dependencies are identified

**Format Review**:
- [ ] YAML frontmatter is complete and valid
- [ ] EARS patterns are used correctly
- [ ] Language is consistent
- [ ] Structure follows template
- [ ] TAG references are correct

**Quality Review**:
- [ ] Technical feasibility is confirmed
- [ ] Security considerations are addressed
- [ ] Performance requirements are realistic
- [ ] Integration points are defined
- [ ] Risk assessment is complete

### Review Workflow

1. **Author Review**: Self-review for completeness
2. **Peer Review**: Technical review by team member
3. **Expert Review**: Domain expert validation
4. **Stakeholder Review**: Business requirement validation
5. **Final Approval**: Sign-off for implementation

## Tools and Templates

### SPEC Templates

MoAI-ADK provides templates for different types of specifications:

```markdown
# API Endpoint Template
# @SPEC:EX-API-001: [Feature Name]

## Overview
Brief description of the API endpoint functionality.

## EARS Requirements
- The system SHALL provide [HTTP method] [endpoint path]
- WHEN [request condition], the system SHALL [response]
- WHILE [state condition], the system SHALL [behavior]

## Technical Requirements
- Request format: [JSON schema]
- Response format: [JSON schema]
- Status codes: [list of expected codes]
- Authentication: [requirements]
- Rate limiting: [limits]

## Acceptance Criteria
- [ ] Endpoint accepts valid requests
- [ ] Proper error handling for invalid requests
- [ ] Response format matches specification
- [ ] Authentication requirements enforced
```

### Validation Tools

**Built-in Validation**:
```bash
# Validate SPEC syntax and structure
moai-adk validate-spec .moai/specs/SPEC-AUTH-001/spec.md

# Check for common issues
moai-adk lint-specs .moai/specs/

# Generate SPEC statistics
moai-adk spec-stats
```

**Integration with Alfred**:
```bash
# Alfred automatically validates SPECs during creation
/alfred:1-plan "Feature description"

# Check SPEC quality before implementation
/alfred:validate-spec AUTH-001
```

## Best Practices Summary

### Writing SPECs

1. **Start with User Value**: Focus on what users need to accomplish
2. **Be Specific**: Use precise language and avoid ambiguity
3. **Think in Scenarios**: Consider all possible use cases and edge cases
4. **Define Success**: Clear acceptance criteria that can be tested
5. **Consider Constraints**: Identify technical and business limitations

### Managing SPECs

1. **Version Control**: Track changes with semantic versioning
2. **Regular Reviews**: Keep SPECs updated as requirements evolve
3. **Link Everything**: Maintain traceability with @TAG system
4. **Document Decisions**: Record why decisions were made
5. **Plan for Evolution**: Design for future changes and extensions

### Team Collaboration

1. **Shared Understanding**: Ensure team alignment on requirements
2. **Early Feedback**: Get input before implementation starts
3. **Continuous Refinement**: Update SPECs as understanding improves
4. **Knowledge Sharing**: Use SPECs as learning documents
5. **Quality Standards**: Maintain high standards for all SPECs

Remember: A well-written SPEC is an investment in project success. It prevents misunderstandings, reduces rework, and ensures that everyone is building the same thing! <span class="material-icons">target</span>