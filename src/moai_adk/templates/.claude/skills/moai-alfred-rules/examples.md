# Practical Examples: MoAI-ADK Rules in Action

## Example 1: Mandatory Skill Invocation

### ❌ WRONG: Skipping Skill for TRUST Validation
```
User: "Check code quality of my new feature"
Alfred: "Looking at your code...linting...checking coverage...
Actually, looks good! No issues found."
```

**Problem**: Direct linting without `Skill("moai-foundation-trust")` → misses quality gates, context overhead

### ✅ CORRECT: Using Skill Invocation
```
User: "Check code quality of my new feature"
Alfred: Skill("moai-foundation-trust")
Result: ✓ Structured TRUST 5 validation (Test, Readable, Unified, Secured, Trackable)
```

---

## Example 2: AskUserQuestion for Ambiguous Requirements

### ❌ WRONG: Making Assumptions
```
User: "Add user authentication to the app"
Alfred: [Assumes OAuth] "I'll implement Google OAuth..."
Reality: User wanted basic email/password auth
Result: Wasted effort, wrong implementation
```

### ✅ CORRECT: Using AskUserQuestion
```
User: "Add user authentication to the app"
Alfred: AskUserQuestion([
  "What auth method?": [Email/Password, OAuth, JWT, Multi-factor],
  "What scope?": [Admin users only, All users, Public signup],
  "Session duration?": [24 hours, 7 days, Remember me]
])
User selects: [Email/Password, All users, 7 days]
Alfred: "I'll implement email/password with 7-day sessions..."
Result: ✓ Correct implementation, zero rework
```

---

## Example 3: TAG Chain Validation Rules

### ❌ WRONG: Orphan TAGs
```python
# src/auth/user.py
def verify_email(token):  # @CODE:AUTH-015
    """Verify user email with token"""
    ...

# PROBLEM: No corresponding @TEST:AUTH-015 in tests/
# Result: TAG chain broken, no traceability
```

### ✅ CORRECT: Complete TAG Chain
```python
# tests/auth/test_email_verify.py
def test_verify_email_with_valid_token():  # @TEST:AUTH-015
    """Verify user email with token"""
    ...
    assert user.email_verified = True

# src/auth/user.py
def verify_email(token):  # @CODE:AUTH-015
    """Verify user email with token"""
    ...

# README.md
## Email Verification (AUTH-015)
User can verify their email... # @DOC:AUTH-015

# HISTORY
- @TEST:AUTH-015: Added email verification test (2025-11-02)
- @CODE:AUTH-015: Implemented token validation (2025-11-02)
- @DOC:AUTH-015: Updated README (2025-11-02)

# Result: ✓ Complete traceability from SPEC → TEST → CODE → DOC
```

---

## Example 4: TRUST 5 Quality Gates

### Scenario: Code Review with TRUST Validation

**Before TRUST Check**:
```python
# user_service.py - 60% test coverage, no input validation
def create_user(name, email, password):
    user = User(name=name, email=email, password=password)
    db.save(user)  # ❌ No validation, SQL injection risk
    return user
```

**After Skill("moai-foundation-trust")**:
```
⚠️ TRUST 5 Validation Results:

1. Test (85%+ required): ❌ FAIL
   - Current: 60% coverage
   - Issue: Missing edge case tests (empty email, duplicate email)

2. Readable: ⚠️ WARNING
   - Issue: No input validation docstring
   - Fix: Add parameter validation comments

3. Unified: ✅ PASS
   - Pattern matches other services

4. Secured: ❌ FAIL
   - OWASP A04: SQL injection risk (no parameterized query)
   - Fix: Use ORM or prepared statements

5. Trackable: ❌ FAIL
   - No @TAG assigned
   - Fix: Add @TEST:USER-002 and @CODE:USER-002

Action: Skill("moai-essentials-refactor") to fix security + add tests
```

---

## Example 5: Skill Tier Usage Pattern

### Scenario: Implementing "Add authentication dashboard"

```
User Request: "Add authentication dashboard"
     ↓
Step 1: Intent unclear → AskUserQuestion
- Framework choice?
- Admin vs user dashboard?
- Real-time monitoring needed?
     ↓
Step 2: Create SPEC → Skill("moai-alfred-spec-authoring")
     ↓
Step 3: TDD Implementation → Skill("moai-foundation-trust")
     ↓
Step 4: Code Review → Skill("moai-essentials-review")
     ↓
Step 5: Git Workflow → Skill("moai-foundation-git")
     ↓
Step 6: Sync Documentation → Skill("moai-foundation-tags")
     ↓
Result: Complete implementation with TRUST 5 + TAG traceability
```

---

## Example 6: EARS Requirement Syntax with Rules

### ❌ WRONG: Vague Requirement
```
"User can view dashboard"
```

### ✅ CORRECT: EARS Syntax
```
@TAG:DASHBOARD-001

Given a logged-in user with admin role
When the user navigates to /admin/dashboard
Then the system displays real-time metrics (CPU, memory, requests)
And metrics update every 5 seconds
And unauthorized users receive 403 Forbidden

Optional:
- User can export metrics as CSV
- Metrics are cached for 60 seconds

Unwanted Behaviors:
- Dashboard does NOT expose raw logs
- Dashboard does NOT allow system shutdown
```

**Why**: Skill("moai-foundation-ears") enforces clarity, reduces ambiguity

---

## Example 7: Command Sequence with Rules

### Workflow: "I want to add email verification feature"

```bash
# Step 1: Initialize project (creates CLAUDE.md, config.json)
/alfred:0-project

# Step 2: Write SPEC
/alfred:1-plan "Email Verification"
→ Skill("moai-alfred-spec-metadata") validates metadata
→ Skill("moai-foundation-ears") validates requirement syntax
→ Creates .moai/specs/SPEC-AUTH-015/

# Step 3: TDD Implementation
/alfred:2-run SPEC-AUTH-015
→ Skill("moai-foundation-trust") enforces 85%+ test coverage
→ RED → GREEN → REFACTOR cycle
→ Creates @TEST, @CODE TAGs

# Step 4: Sync Documentation
/alfred:3-sync
→ Skill("moai-foundation-git") enforces commit message format
→ Skill("moai-foundation-tags") validates TAG chain
→ Updates README, CHANGELOG with SPEC reference

# Result: Complete feature with SPEC → TEST → CODE → DOC traceability
```

---

## Example 8: When NOT to Use Skills (Exceptions)

### ✅ LEGITIMATE: Direct tools without Skill()

```
# 1. Reading single file for context
Read(file_path="/path/to/config.json")  ← Direct read OK

# 2. Quick list of files
Bash("ls src/")  ← Direct bash OK

# 3. Git operations already wrapped in Skill
Bash("git log")  ← Called from Skill("moai-foundation-git")  ← OK

# 4. JIT context retrieval for current task
Grep("authenticate" type:py)  ← OK for immediate task context
```

### ❌ WRONG: Skipping Skill When It Exists

```
# 1. TRUST validation
❌ Direct: Bash("pytest --cov")
✅ Correct: Skill("moai-foundation-trust")

# 2. TAG validation
❌ Direct: Bash("rg '@TAG:' -n")
✅ Correct: Skill("moai-foundation-tags")

# 3. SPEC authoring
❌ Direct: Manual YAML writing
✅ Correct: Skill("moai-foundation-specs")
```

---

## Key Takeaways

1. **Always use Skill() explicitly** for 55+ knowledge domains
2. **Use AskUserQuestion** when user intent is ambiguous
3. **Enforce TRUST 5** before every code completion
4. **Maintain TAG chains** from SPEC → TEST → CODE → DOC
5. **Use EARS syntax** for all requirements
6. **Progressive disclosure** guides skill content layers
7. **JIT retrieval** manages context efficiently

---

**Learn More**: See `reference.md` for complete rule definitions, decision trees, and validation procedures.
