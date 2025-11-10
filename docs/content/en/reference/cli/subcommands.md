# Alfred Workflow Command Guide

Alfred commands that control MoAI-ADK's core 4-step workflow.

> **Important**: Alfred commands are only available **within the Claude Code environment**.

## Workflow Overview

```
/alfred:0-project (Initialization)
    ↓
/alfred:1-plan (Planning: SPEC Writing)
    ↓
/alfred:2-run (Execution: TDD Development)
    ↓
/alfred:3-sync (Synchronization: Documentation/Validation)
    ↓
Complete and PR Creation
```

______________________________________________________________________

## 1. /alfred:0-project

**Project Setup and Initialization**

### Syntax

```
/alfred:0-project [option]
```

### Options

```
setting     View current settings
update      Modify project settings
```

### Key Features

- :bullseye: Set project metadata (name, description, language)
- 🌐 Select conversation language (Korean, English, Japanese, Chinese)
- 🔧 Select development mode (solo/team/org)
- 📋 Initialize SPEC-First TDD checklist
- 🏷️ Activate TAG system
- 📊 Set test coverage goal (default 85%)

### Interaction Example

```
/alfred:0-project

> Alfred: Please enter project name
User: Hello World API

> Alfred: Project description?
User: Simple REST API tutorial

> Alfred: Primary language?
User: [1] Python  [2] TypeScript  [3] Go
Select: 1

> Alfred: Conversation language?
User: [1] Korean  [2] English
Select: 1

> Alfred: Development mode?
User: [1] Solo  [2] Team  [3] Organization
Select: 1

✅ Project initialization complete!
```

### Generated Settings

`.moai/config.json`:

```json
{
  "project": {
    "name": "Hello World API",
    "description": "Simple REST API tutorial",
    "language": "python"
  },
  "language": {
    "conversation_language": "ko",
    "conversation_language_name": "Korean"
  },
  "constitution": {
    "test_coverage_minimum": 85
  }
}
```

______________________________________________________________________

## 2. /alfred:1-plan

**SPEC Writing and Planning**

### Syntax

```
/alfred:1-plan "Title1" "Title2" ... | SPEC-ID modifications
```

### Use Cases

#### Create New SPEC

```
/alfred:1-plan "User authentication system"
```

Or multiple SPECs:

```
/alfred:1-plan "Login feature" "Registration" "Password reset"
```

#### Modify Existing SPEC

```
/alfred:1-plan SPEC-001 "Login feature (Modified: Add OAuth 2.0)"
```

### Alfred's Planning Process

1. **Intent Understanding**: Request analysis and clarification

   - If ambiguous, collect additional information via AskUserQuestion

2. **Plan Creation**: Structured execution strategy

   - Task decomposition (Decomposition)
   - Dependency analysis (Dependency Analysis)
   - Identify parallelization opportunities (Parallelization)
   - Specify affected files (File List)
   - Estimate time (Time Estimation)

3. **User Approval**: Present plan and request approval

   ```
   Alfred: I've planned as follows. Proceed?

   📋 Plan Summary:
   - SPEC-001: Login feature
   - SPEC-002: Registration
   - Affected files: 5
   - Estimated time: 30 minutes

   [Proceed] [Modify] [Cancel]
   ```

4. **TodoWrite Initialization**: Start tracking all task items

### Generated Files

```
.moai/specs/SPEC-001/
├── spec.md              # SPEC document (EARS format)
├── requirements.md      # Detailed requirements
└── tests.md            # Test plan
```

### SPEC Document Structure

```markdown
# SPEC-001: Login Feature

## Requirements (EARS Format)

### Basic Requirements
- GIVEN: User visits login page
  WHEN: Valid email and password entered
  THEN: Session created and dashboard redirect

### Error Handling
- GIVEN: Login page
  WHEN: Incorrect password entered
  THEN: "Password incorrect" message

## Test Plan
- [ ] Successful login with valid credentials
- [ ] Failed login with invalid credentials
- [ ] Login after new user registration
```

______________________________________________________________________

## 3. /alfred:2-run

**TDD Implementation Execution**

### Syntax

```
/alfred:2-run [SPEC-ID | "all"]
```

### Use Cases

#### Develop Specific SPEC

```
/alfred:2-run SPEC-001
```

#### Develop All SPECs

```
/alfred:2-run all
```

### Execution Workflow

Alfred strictly follows TDD's 3 phases:

#### Phase 1: RED - Write Failing Tests

```
Alfred: Starting RED phase
- Create test file: tests/test_login.py
- Write tests (SPEC-based)
- Execute → All fail :x:

✅ RED phase complete
All tests are in failing state.

[Proceed to GREEN phase]
```

**Sample Test**:

```python
# tests/test_login.py @TEST:SPEC-001:*
import pytest
from app import login

def test_valid_login():
    """GIVEN: Login page
       WHEN: Valid credentials
       THEN: Session created"""
    result = login("user@example.com", "password123")
    assert result["status"] == "success"
    assert result["session"] is not None

def test_invalid_password():
    """GIVEN: Login page
       WHEN: Incorrect password
       THEN: Error message"""
    result = login("user@example.com", "wrong")
    assert result["status"] == "error"
    assert "password" in result["message"].lower()
```

#### Phase 2: GREEN - Pass Tests with Minimal Implementation

```
Alfred: Starting GREEN phase
- Add minimal implementation: app.py
- Execute tests → All pass ✅

✅ GREEN phase complete
All tests pass.

[Proceed to REFACTOR phase]
```

**Sample Implementation**:

```python
# app.py @CODE:SPEC-001:*
def login(email, password):
    """Login processing"""
    if password == "password123":
        return {
            "status": "success",
            "session": "session_123"
        }
    else:
        return {
            "status": "error",
            "message": "Password incorrect"
        }
```

#### Phase 3: REFACTOR - Improve Code Quality

```
Alfred: Starting REFACTOR phase
- Improve error handling
- Add data validation
- Code cleanup

✅ REFACTOR phase complete
All tests pass, code quality improved.

[All tasks complete]
```

**Improved Implementation**:

```python
# app.py (Improved)
from flask import session
from werkzeug.security import check_password_hash
from models import User

def login(email, password):
    """Login processing (Improved version)"""
    # Input validation
    if not email or not password:
        raise ValueError("Email and password are required")

    # User lookup
    user = User.query.filter_by(email=email).first()
    if not user:
        return {
            "status": "error",
            "message": "User not registered"
        }

    # Password verification
    if not check_password_hash(user.password_hash, password):
        return {
            "status": "error",
            "message": "Password incorrect"
        }

    # Session creation
    session['user_id'] = user.id
    return {
        "status": "success",
        "session": session.sid,
        "user": {
            "id": user.id,
            "email": user.email,
            "name": user.name
        }
    }
```

### TodoWrite Tracking

Alfred automatically tracks each phase:

```
[in_progress] RED: SPEC-001 test writing
[completed]   RED: SPEC-001 test writing
[in_progress] GREEN: SPEC-001 minimal implementation
[completed]   GREEN: SPEC-001 minimal implementation
[in_progress] REFACTOR: SPEC-001 code improvement
[completed]   REFACTOR: SPEC-001 code improvement
```

______________________________________________________________________

## 4. /alfred:3-sync

**Documentation Synchronization and Validation**

### Syntax

```
/alfred:3-sync [Mode] [Target] [Path]
```

### Mode

```
auto         Auto sync (recommended)
force        Force sync
status       View current status only
project      Full project validation
```

### Target

```
SPEC-001         Sync specific SPEC
all              Sync all SPECs
```

### Key Features

1. **Document Generation** (based on generation settings)

   - Auto-generate API documentation
   - Architecture diagrams
   - Deployment guides

2. **TAG Validation**

   - Verify SPEC → TEST → CODE → DOC connections
   - Detect and remove orphaned TAGs
   - Verify traceability integrity

3. **Quality Validation**

   - Verify test coverage ≥ 85%
   - Verify all tests pass
   - Code style check

4. **PR Creation** (Team mode)

   - Create PR targeting develop branch
   - Change summary
   - Include validation results

### Synchronization Process

```
/alfred:3-sync auto SPEC-001

➡️ Step 1: SPEC Validation
✅ SPEC-001 structure normal
✅ 8 requirements confirmed

➡️ Step 2: TAG Validation
✅ @TEST:SPEC-001 tags: 12
✅ @CODE:SPEC-001 tags: 15
✅ @DOC:SPEC-001 tags: 3
⚠️ 2 orphaned TAGs removed

➡️ Step 3: Quality Validation
✅ Test coverage: 92%
✅ All tests pass
✅ Code style normal

➡️ Step 4: Document Generation
✅ API documentation generated: docs/api/SPEC-001.md
✅ Architecture diagram generated

➡️ Step 5: PR Creation
✅ PR #23 created
📝 Title: "feat: SPEC-001 Login feature implementation"
```

______________________________________________________________________

## 5. /alfred:9-feedback

**GitHub Issue Creation (Feedback)**

### Syntax

```
/alfred:9-feedback
```

### Features

- 🐛 Bug report
- 💡 Feature suggestion
- 📝 Improvement
- ❓ Question

### Interaction Example

```
/alfred:9-feedback

> Alfred: Feedback type?
Select: [1] Bug  [2] Feature  [3] Improvement  [4] Question

> Select: 1

> Title?
Input: "Session not maintained after login"

> Description?
Input: "Logs out after refresh after login"

> Reproduction steps?
Input: "1. Login 2. Refresh 3. Access dashboard"

> Expected behavior?
Input: "Session should be maintained"

✅ GitHub Issue #24 created
📝 Title: "🐛 Session not maintained after login"
```

______________________________________________________________________

## Command Quick Reference

### Complete Workflow

```bash
# 1. Project initialization
/alfred:0-project

# 2. SPEC writing
/alfred:1-plan "Login feature" "Registration"

# 3. TDD implementation
/alfred:2-run all

# 4. Synchronization and validation
/alfred:3-sync auto all

# Complete! PR automatically created
```

### Partial Workflow

```bash
# Modify specific SPEC only
/alfred:1-plan SPEC-001 "Login feature (Add OAuth)"

# Develop that SPEC only
/alfred:2-run SPEC-001

# Sync that SPEC only
/alfred:3-sync auto SPEC-001
```

______________________________________________________________________

## Error Handling

### "Alfred command not recognized"

```bash
# Restart Claude Code
exit

# Start new session
claude

# Reinitialize project
/alfred:0-project
```

### "SPEC file not found"

```bash
# Check project status
moai-adk status

# Reinitialize
moai-adk init . --force
/alfred:0-project
```

### "Insufficient test coverage"

```bash
# Check current coverage
moai-adk status

# Add missing tests
# Add tests in tests/ directory

# Sync again
/alfred:3-sync auto SPEC-001
```

______________________________________________________________________

**Next**: [moai-adk Command Reference](moai-adk.md) or [Alfred Concepts](guides/alfred/index.md)



