# Core Sub-agents Detailed Guide

A complete reference for Alfred's 10 core agents.

## Overview

| #   | Agent                | Role              | Skills | Optimal Size     |
| --- | -------------------- | ----------------- | ------ | ---------------- |
| 1   | project-manager      | Project init      | 5      | 1-person team    |
| 2   | spec-builder         | SPEC writing      | 8      | All teams        |
| 3   | implementation-planner | Planning       | 6      | Team projects    |
| 4   | tdd-implementer      | TDD execution     | 12     | All teams        |
| 5   | doc-syncer           | Doc sync          | 8      | All teams        |
| 6   | tag-agent            | TAG validation    | 4      | Medium+ projects |
| 7   | git-manager          | Git automation    | 10     | All teams        |
| 8   | trust-checker        | Quality validation| 7      | Release stage    |
| 9   | quality-gate         | Release prep      | 6      | Production       |
| 10  | debug-helper         | Error resolution  | 9      | When issues occur|

______________________________________________________________________

## 1. project-manager

**Role**: Project initialization and metadata management

### Activation Conditions

```
/alfred:0-project [setting|update]
```

### Key Responsibilities

- Set project metadata (name, description, team size)
- Select and apply conversation language
- Determine development mode (solo/team/org)
- Initialize `.moai/config.json`
- Set TRUST 5 principle defaults

### Interaction Format

```
User: /alfred:0-project

Alfred: Project name?
→ project-manager: Validate and save input

Alfred: Development mode?
→ project-manager: Determine settings based on team size

Alfred: Conversation language?
→ project-manager: Set language for all subsequent communication

Result: .moai/config.json created
```

### Generated File Structure

```
.moai/
├── config.json           # Project settings
├── specs/               # SPEC storage directory
├── docs/                # Generated documentation
├── reports/             # Analysis reports
└── scripts/             # Utilities
```

### Usage Scenarios

- **New project start**: First-time Alfred initialization
- **Settings change**: Modify language, team mode, test coverage goals
- **Multiple projects**: Independent settings per project

______________________________________________________________________

## 2. spec-builder

**Role**: Write SPEC documents in EARS format

### Activation Conditions

```
/alfred:1-plan "Title1" "Title2" ...
/alfred:1-plan SPEC-ID "modifications"
```

### Key Responsibilities

- Structure user requirements in EARS format
- Auto-generate SPEC IDs (SPEC-001, SPEC-002...)
- Validate requirement clarity
- Draft test plans
- Define implementation scope

### EARS Format Structure

```
GIVEN:     Initial situation description
WHEN:      User action
THEN:      Expected result
```

### Example

**User Input**:

```
/alfred:1-plan "User authentication system"
```

**Generated SPEC**:

```markdown
# SPEC-001: User Authentication System

## Requirements

### Login Feature
- GIVEN: User visits login page
  WHEN: Valid email and password entered
  THEN: Session created and dashboard redirect

### Password Error Handling
- GIVEN: Login page
  WHEN: Incorrect password entered
  THEN: "Password error" message displayed

## Test Plan
- [ ] Successful login
- [ ] Password error
- [ ] Account lock (5 failures)
```

### Quality Criteria

- Clear requirements (5+)
- Unambiguous expressions
- Testable conditions
- Implementable scope

______________________________________________________________________

## 3. implementation-planner

**Role**: Architecture and execution planning

### Activation Conditions

```
/alfred:2-run SPEC-ID (at start)
```

### Key Responsibilities

- Break down SPEC into implementation steps
- Design file and directory structure
- Analyze task dependencies
- Identify parallelization opportunities
- Estimate time and difficulty

### Planning Process

```
SPEC Analysis
    ↓
Task Breakdown (5-10 steps)
    ↓
Dependency Mapping
    ↓
Identify Parallelization Opportunities
    ↓
List Affected Files
    ↓
Time Estimation
    ↓
Request User Approval
```

### Planning Document Example

```
SPEC-001: User Authentication System

📋 Task Breakdown:
1. Design data models (User, Session)
2. Create database schema
3. Implement password hashing function
4. Implement login endpoint
5. Write session management middleware
6. Logout endpoint
7. Password reset
8. Account lock mechanism

🔄 Dependencies:
1 → 2 → 3 → 4
     ↓
     5 → 6, 7 → 8

⚡ Parallelization:
- 4 and 5 can be parallelized
- 6, 7, 8 can be parallelized

📁 Affected Files:
- models/user.py (NEW)
- models/session.py (NEW)
- api/auth.py (NEW)
- middleware/session.py (NEW)
- tests/test_auth.py (NEW)
- docs/auth.md (NEW)

⏱️ Estimated Time: 2 hours (3 phases: RED/GREEN/REFACTOR)
```

______________________________________________________________________

## 4. tdd-implementer

**Role**: Execute RED-GREEN-REFACTOR cycle

### Activation Conditions

```
/alfred:2-run SPEC-ID (during execution)
```

### Key Responsibilities

- RED phase: Write failing tests
- GREEN phase: Minimal implementation
- REFACTOR phase: Improve code quality
- Update TodoWrite after each phase completion
- Track test status

### TDD 3-Phase Implementation

#### Phase 1: RED

```python
# Write only failing tests
def test_user_registration():
    user = register_user("user@example.com", "password123")
    assert user.email == "user@example.com"
    assert user.is_verified == False

# Execute → FAIL :x:
```

#### Phase 2: GREEN

```python
# Minimal implementation
def register_user(email, password):
    user = User(email=email)
    user.set_password(password)
    db.session.add(user)
    db.session.commit()
    return user

# Execute → PASS ✅
```

#### Phase 3: REFACTOR

```python
# Improve code quality (maintain tests)
def register_user(email, password):
    """User registration"""
    # Input validation
    if not is_valid_email(email):
        raise ValueError("Invalid email")
    if len(password) < 8:
        raise ValueError("Password too short")

    # Duplicate check
    if User.query.filter_by(email=email).first():
        raise ValueError("User already exists")

    # Create user
    user = User(email=email)
    user.set_password(password)
    db.session.add(user)
    db.session.commit()

    return user
```

### TodoWrite Tracking

```
[in_progress] RED: SPEC-001 test writing
[completed]   RED: SPEC-001 test writing
[in_progress] GREEN: SPEC-001 minimal implementation
[completed]   GREEN: SPEC-001 minimal implementation
[in_progress] REFACTOR: SPEC-001 code improvement
[completed]   REFACTOR: SPEC-001 code improvement
```

______________________________________________________________________

## 5. doc-syncer

**Role**: Automatic documentation generation and synchronization

### Activation Conditions

```
/alfred:3-sync auto [SPEC-ID]
```

### Key Responsibilities

- Auto-generate API documentation (OpenAPI/Swagger)
- Generate architecture diagrams
- Write deployment guides
- Generate change summary documents
- Validate document links

### Generated Document Types

| Document     | Content              | Format      |
| ------------ | -------------------- | ----------- |
| API Spec     | RESTful endpoints    | OpenAPI 3.1 |
| Architecture | System diagrams      | Mermaid     |
| Deployment   | Deployment procedures| Markdown    |
| Changelog    | Changes              | Markdown    |
| Migration    | Data migration       | SQL + description |

### Generation Location

```
docs/
├── api/
│   └── SPEC-001.md          # API documentation
├── architecture/
│   └── SPEC-001.md          # Architecture
├── deployment/
│   └── SPEC-001.md          # Deployment guide
├── migrations/
│   └── 001_create_users.sql # Migration
└── changelog/
    └── v1.0.0.md            # Changes
```

______________________________________________________________________

## 6. tag-agent

**Role**: TAG validation and traceability management

### Activation Conditions

```
/alfred:3-sync auto [SPEC-ID]
```

### Key Responsibilities

- Validate SPEC → TEST → CODE → DOC TAG chain
- Detect and remove orphaned TAGs
- Validate TAG naming rules
- Verify traceability integrity

### TAG Chain

```
SPEC-001 (Requirements)
    ↓
@TEST:SPEC-001:* (Tests)
    ↓
@CODE:SPEC-001:* (Implementation)
    ↓
@DOC:SPEC-001:* (Documentation)
    ↓
Cross-reference (Complete traceability)
```

### Example

```python
# @CODE:SPEC-001:register_user
def register_user(email: str, password: str) -> User:
    """User registration"""
    # @CODE:SPEC-001:validate_email
    if not is_valid_email(email):
        raise ValueError("Invalid email")

    # @CODE:SPEC-001:hash_password
    hashed = hash_password(password)

    # @CODE:SPEC-001:create_user
    user = User(email=email, password_hash=hashed)
    db.session.add(user)
    db.session.commit()

    return user

# @TEST:SPEC-001:test_register_success
def test_register_success():
    user = register_user("test@example.com", "password123")
    assert user.email == "test@example.com"
```

______________________________________________________________________

## 7. git-manager

**Role**: Git workflow automation

### Activation Conditions

Automatically activated at all stages

### Key Responsibilities

- Create feature branches (feature/SPEC-001)
- Auto-generate commit messages
- Commit by RED/GREEN/REFACTOR phase
- Create and manage PRs
- Validate before merge

### Git Workflow

```
main
    ↓
develop (base branch)
    ↓
feature/SPEC-001 (feature branch)
    │
    ├── feat: RED phase (commit)
    ├── feat: GREEN phase (commit)
    ├── refactor: code quality (commit)
    │
    ↓
PR #23 (develop ← feature/SPEC-001)
    ├── Test validation
    ├── Code review
    └── Merge
    ↓
develop (merge complete)
    ↓
main (on release)
```

### Commit Message Format

```
<type>: <description>

🤖 Generated by Claude Code

Co-Authored-By: 🎩 Alfred@MoAI
```

**Types**:

- `feat`: New feature
- `fix`: Bug fix
- `refactor`: Code improvement
- `test`: Test addition
- `docs`: Documentation update

______________________________________________________________________

## 8. trust-checker

**Role**: TRUST 5 principle validation

### Activation Conditions

```
/alfred:2-run SPEC-ID (after completion)
```

### TRUST 5 Principles

| Principle      | Description        | Validation        |
| -------------- | ------------------ | ----------------- |
| **T**est First | Test-driven development | Coverage 85%+ |
| **R**eadable   | Readable code      | Linting pass      |
| **U**nified    | Consistent structure | Naming rules compliance |
| **S**ecured    | Security           | Security scan pass |
| **T**rackable  | Traceability       | TAG integrity     |

### Validation Results

```
✅ Test First: 92% coverage (target: 85%)
✅ Readable: MyPy complete, ruff pass
✅ Unified: Naming rules compliance
✅ Secured: Dependency security check pass
✅ Trackable: 12 TAGs validated

:bullseye: TRUST 5 Compliance: PASS ✅
```

______________________________________________________________________

## 9. quality-gate

**Role**: Release readiness check

### Activation Conditions

```
/alfred:3-sync auto all (final stage)
```

### Validation Items

- ✅ All SPECs complete
- ✅ Test coverage 85% or higher
- ✅ All tests pass
- ✅ 0 security vulnerabilities
- ✅ 100% documentation completeness
- ✅ TAG integrity

### Release Decision

```
All items pass → PR Merge → Release ready

Failed items exist → Detailed report → Improvement needed
```

______________________________________________________________________

## 10. debug-helper

**Role**: Error analysis and automatic fixes

### Activation Conditions

```
Automatically activated when errors or exceptions occur
```

### Key Responsibilities

- Analyze error stack traces
- Identify root causes
- Suggest solutions
- Determine if auto-fix is possible
- Suggest temporary workarounds

### Error Handling Process

```
Error occurs
    ↓
debug-helper: Analysis
    ├─ Identify type
    ├─ Trace cause
    ├─ Search similar cases
    └─ Suggest solution
    ↓
[Auto-fix possible?]
    ├─ YES → Fix and re-execute
    └─ NO → Provide detailed guide
```

______________________________________________________________________

## Agent Collaboration Examples

### Complete Workflow Example

```
SPEC-001 creation (spec-builder)
    ↓
Implementation plan (implementation-planner)
    ↓
RED phase tests (tdd-implementer)
    ↓
GREEN phase implementation (tdd-implementer)
    ↓
REFACTOR phase (tdd-implementer)
    ↓
TRUST 5 validation (trust-checker)
    ↓
Git commit (git-manager)
    ↓
Documentation generation (doc-syncer)
    ↓
TAG validation (tag-agent)
    ↓
Release preparation (quality-gate)
    ↓
Complete!
```

______________________________________________________________________

**Next**: [Expert Agents](experts.md) or [Agents Overview](index.md)



