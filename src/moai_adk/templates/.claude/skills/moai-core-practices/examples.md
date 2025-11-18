# Practical Examples: Workflow Execution

## Example 1: JIT Context Retrieval

### Task: "Add email verification feature"

**Phase 1: High-level Brief**
```markdown
## Email Verification Feature
- Goal: User can verify email after signup
- Success: User receives email, clicks link, marked verified
- Stakeholders: User (receiver), Admin (monitoring)
```

**Phase 2: Technical Core**
```markdown
## Architecture
- Entry point: src/api/auth.py - POST /auth/signup
- Domain model: models/user.py - User.email_verified
- Email service: infra/email_service.py - send_verification_email()
```

**Phase 3: Edge Cases**
```markdown
## Known Gotchas
- Token expires in 24h
- Duplicate email prevents signup
- Test mode uses mock email service (doesn't send)
```

---

## Example 2: Feature Implementation Workflow

```bash
# Step 1: Create SPEC
/alfred:1-plan "Email Verification"

# Step 2: TDD RED phase
/alfred:2-run SPEC-AUTH-015
# Write tests: test_verify_email_valid_token, test_token_expired, test_duplicate_email
# RED: All 3 tests fail

# Step 3: TDD GREEN phase
# Implement: User.verify_email(token)
# GREEN: All 3 tests pass

# Step 4: TDD REFACTOR phase
# Improve: Extract token validation logic
# REFACTOR: Tests still pass, code cleaner

# Step 5: Sync
/alfred:3-sync
# Update README with email verification docs
# Update CHANGELOG with SPEC-AUTH-015 reference
```

---

## Example 3: Explore Agent for Large Codebase

### ❌ WRONG: Manual file hunting
```
User: "How is authentication currently implemented?"
Alfred:
  grep -r "authenticate" src/
  grep -r "login" src/
  grep -r "jwt" src/
  # … 20 files to read, context bloated
```

### ✅ CORRECT: Use Explore Agent
```
User: "How is authentication currently implemented?"
Alfred: Task(subagent_type="Explore", prompt="Find authentication flow including entry points, models, middleware")
Explore:
  - Found: src/api/auth.py (login endpoint)
  - Found: models/user.py (User model, password_hash)
  - Found: middleware/auth.py (JWT validation)
  - Found: test/test_auth.py (test patterns)
Result: Clear architecture summary without bloated context
```

---

## Example 4: Problem Diagnosis

### Scenario: Tests failing unexpectedly

```
Error: "test_email_verification failed - connection timeout"

Debugging Steps:
1. Check stack trace → Email service timeout
2. Skill("moai-essentials-debug") → "Is test mode configured?"
3. Diagnosis → Production email service called in tests
4. Fix → Add mock for test environment
5. Verify → Tests pass again
```

---

## Example 5: Multi-step Workflow with Agents

```
User: "Implement search feature with 95%+ test coverage"

Alfred:
1. AskUserQuestion → Clarify search scope (users? products? all?)
2. Skill("moai-core-spec-metadata-extended") → Create SPEC-SEARCH-001
3. Skill("moai-foundation-trust") → Enforce 95% coverage target
4. Skill("moai-essentials-debug") → Handle search performance
5. Skill("moai-foundation-tags") → Validate TAG chain
6. Skill("moai-foundation-git") → Proper commit messages

Result: Complete feature with TRUST 5 + full traceability
```

---

Learn more in `reference.md` for complete workflow patterns and advanced scenarios.
