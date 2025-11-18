# Alfred Proactive Suggestions - Real-World Examples

> **Main Skill**: [SKILL.md](SKILL.md)  
> **API Reference**: [reference.md](reference.md)

---

## Example 1: Database Migration Risk Detection

### Scenario

**User**: `/alfred:2-run SPEC-DATABASE-001`

**SPEC content**:
```
Title: Migrate user table to new schema
Description: Add email_verified column, rename password to password_hash
```

### Alfred Detection

**Risk pattern**: Database Migration
**Risk level**: High
**Affected**: Production database (10K users)

### Proactive Suggestion

```
High-risk operation detected: Database migration

SPEC-DATABASE-001 Impact Analysis:

Affected:
- Table: users (10,243 rows)
- Columns: 2 added, 1 renamed
- Downtime estimate: 5-10 minutes

Recommended safeguards:
1. ✅ Create database backup (pg_dump)
2. ✅ Test migration on staging environment
3. ✅ Prepare rollback script
4. ✅ Schedule maintenance window
5. ✅ Verify migration in dry-run mode

Proceed with implementation?
  [Yes, I've taken precautions] 
  [No, test on staging first] 
  [Show me detailed checklist]
```

### User Response

**Option 1: Yes, I've taken precautions**
→ Alfred proceeds with TDD implementation

**Option 2: No, test on staging first**
→ Alfred pauses, suggests staging deployment command

**Option 3: Show me detailed checklist**
→ Alfred displays full migration best practices guide

---

## Example 2: Repetitive Task Automation

### Scenario

**User behavior** (observed pattern):
```
User edits: src/components/Button.tsx (updates import)
User edits: src/components/Card.tsx (updates import)
User edits: src/components/Modal.tsx (updates import)
```

**Pattern**: Same operation on 3 files

### Alfred Detection

**Optimization pattern**: Repetitive tasks
**Files remaining**: 7 similar components
**Time saved**: ~15 minutes

### Proactive Suggestion

```
Repetitive pattern detected: Updating import statements

Observed pattern:
- Change: import { Theme } from '../theme'
- To: import { Theme } from '@/theme'
- Applied: 3/10 files

Automation opportunity:
1. Analyze your last edits
2. Generate batch refactor script
3. Apply to remaining 7 files
4. Estimated time saved: 15 minutes

Create automation script?
  [Yes, generate and apply] 
  [No, I'll continue manually]
```

### Outcome

**If Yes**:
→ Alfred generates and executes batch script
→ Updates 7 remaining files in 10 seconds
→ Reports results

**If No**:
→ Alfred silently continues
→ Won't suggest automation again this session (avoid repetition)

---

## Example 3: Breaking Change Detection

### Scenario

**User**: Modifies API endpoint signature
```diff
- POST /api/auth/login { email, password }
+ POST /api/auth/login { username, password, mfa_code }
```

### Alfred Detection

**Risk pattern**: Breaking change
**Risk level**: High
**Impact**: 3 frontend clients + 5 external integrations

### Proactive Suggestion

```
Breaking change detected: API endpoint signature modified

Impact Analysis:
- Endpoint: POST /api/auth/login
- Change: Added required field 'username', 'mfa_code'
- Affected clients:
  - Web frontend (React)
  - Mobile app (Flutter)
  - Admin dashboard (Vue)
  - 5 external integrations (webhooks)

Recommended approach:
1. Create v2 endpoint (/api/v2/auth/login)
2. Maintain v1 for 2 releases (deprecation period)
3. Document migration guide
4. Notify stakeholders 2 weeks before v1 sunset
5. Monitor v1 usage metrics

Proceed with breaking change?
  [Yes, create migration plan] 
  [No, use non-breaking approach] 
  [Show API versioning best practices]
```

---

## Example 4: Security Concern Detection

### Scenario

**User** (commits file with API key):
```python
API_KEY = "sk-proj-abc123xyz789"
client = OpenAI(api_key=API_KEY)
```

### Alfred Detection

**Risk pattern**: Security concern
**Risk level**: High
**Exposure**: Credentials in code

### Proactive Suggestion

```
Security concern detected: API key in code

Risk: 
- Exposed credentials if committed to git
- Potential unauthorized access
- Compliance violation (SOC2, GDPR)

Recommended fix:
1. Move API_KEY to .env file
2. Add .env to .gitignore
3. Use environment variable: os.getenv('OPENAI_API_KEY')
4. Rotate compromised key immediately

Automatic fix available:
- Extract to .env
- Update code to use os.getenv()
- Add .env to .gitignore
- Generate key rotation reminder

Fix automatically?
  [Yes, extract to .env now] 
  [I'll fix manually] 
  [This is a test key, ignore]
```

### Outcome

**If Yes (automatic)**:
→ Alfred creates `.env` file
→ Adds `OPENAI_API_KEY=sk-proj-abc123xyz789`
→ Updates code to `os.getenv('OPENAI_API_KEY')`
→ Adds `.env` to `.gitignore`
→ Reminds user to rotate key

---

## Example 5: Production Deployment Without Staging

### Scenario

**User**: `Deploy to production`

**Alfred checks**:
- Last staging deployment: 3 days ago
- No staging test of current changes
- Critical infrastructure change detected

### Alfred Detection

**Risk pattern**: Production operation
**Risk level**: High
**Missing**: Staging verification

### Proactive Suggestion

```
Production deployment without staging verification

Current status:
- Last staging test: 3 days ago (stale)
- Changes since then: 47 commits
- Risk level: High (database schema change included)

Pre-deployment checklist:
- [ ] Tested on staging environment
- [ ] Rollback plan prepared
- [ ] Monitoring alerts configured
- [ ] Team notified
- [ ] Database backup created
- [ ] Feature flags configured

Recommended action:
1. Deploy to staging first
2. Run integration tests
3. Verify database migration
4. Then proceed to production

Deploy to production?
  [Cancel, deploy to staging first] 
  [Continue to production (risky)] 
  [Show deployment best practices]
```

---

## Example 6: Parallel Execution Optimization

### Scenario

**User workflow** (sequential):
```
Running unit tests... (2 min)
Running integration tests... (3 min)
Running E2E tests... (5 min)
Total: 10 minutes
```

### Alfred Detection

**Optimization pattern**: Parallel execution
**Tasks**: 3 independent test suites
**Time savings**: 50% (5 minutes)

### Proactive Suggestion

```
Parallel execution opportunity detected

Current workflow:
1. pytest tests/unit (2 min)
2. pytest tests/integration (3 min)
3. playwright test (5 min)
Total: 10 minutes sequential

Optimized workflow:
Run all test suites in parallel:
- pytest tests/unit &
- pytest tests/integration &
- playwright test &
Total: 5 minutes (max of 3 durations)

Time saved: 5 minutes (50% improvement)

Enable parallel execution?
  [Yes, run in parallel now] 
  [No, keep sequential] 
  [Set as default for future runs]
```

---

## Example 7: Learning Opportunity (Beginner)

### Scenario

**User** (Beginner): Creates test after implementation
```
User writes: src/auth.py (implementation)
User writes: tests/test_auth.py (tests)
```

**Pattern**: Not following TDD (Test-First)

### Alfred Detection

**Learning opportunity**: TDD workflow
**Expertise level**: Beginner
**Trigger**: First-time implementation

### Proactive Suggestion

```
Observation: Tests written after implementation

Best practice: Test-Driven Development (TDD)

Current approach:
1. Write implementation
2. Write tests
Risk: Tests may pass incorrectly or miss edge cases

TDD approach:
1. Write failing test (RED) - defines behavior
2. Implement code (GREEN) - satisfies test
3. Refactor (REFACTOR) - improve with safety net

Benefits:
- 40% fewer bugs (industry data)
- Better code design
- Confidence in refactoring

Next time, try:
- Use /alfred:2-run for automated TDD workflow
- Skill("moai-foundation-trust") - TRUST 5 principles

Would you like to retry with TDD workflow?
  [Yes, revert and start with tests] 
  [No, I understand but prefer current approach] 
  [Show me TDD tutorial]
```

---

## Example 8: Large File Edit Without Tests

### Scenario

**User**: Edits `src/payment/processor.py` (350 lines)

**File analysis**:
- Lines modified: 150
- Test coverage: 0%
- Complexity: High (payment logic)

### Alfred Detection

**Risk pattern**: Large file operation
**Risk level**: Medium
**Missing**: Test coverage

### Proactive Suggestion

```
Large file edit detected: 150 lines modified

File: src/payment/processor.py (350 lines)
Test coverage: 0%
Complexity: High (payment processing logic)

Risk: 
- Regression without tests
- Payment failures in production
- Difficult debugging

Recommendation:
1. Pause current edit
2. Write tests first (TDD)
3. Use /alfred:2-run SPEC-PAYMENT-001
4. Ensures payment logic correctness

Alternative:
- Continue edit but write tests immediately after
- Minimum 80% coverage required for payment code

Proceed?
  [Pause, write tests first (recommended)] 
  [Continue, I'll write tests after] 
  [Show payment testing best practices]
```

---

## Example 9: Expert User - Advanced Technique Suggestion

### Scenario

**User** (Expert): Manually performs complex git operations repeatedly
```
git checkout -b feature/auth
git add .
git commit -m "feat: add auth"
git push -u origin feature/auth
gh pr create --title "Auth" --body "..."
```

**Pattern**: Repeated 5 times in past week

### Alfred Detection

**Optimization pattern**: Manual workflow
**Expertise level**: Expert
**Candidate**: Custom workflow automation

### Proactive Suggestion

```
Advanced automation opportunity

Observed pattern:
- 5 PRs created this week
- Same workflow: branch → commit → push → PR
- Time spent: ~5 min each = 25 min/week

Suggested optimization:
Create custom Alfred command: /alfred:create-pr

Workflow automation:
1. Auto-generate branch from SPEC ID
2. Create commit with conventional format
3. Push to origin with upstream tracking
4. Generate PR with SPEC context
5. Request reviewers automatically

Time saved: 20 min/week
Setup time: 10 minutes

Would you like guidance on custom command creation?
  [Yes, show me how] 
  [No, I prefer manual control] 
  [Email me documentation later]
```

---

## Summary: Suggestion Types by Scenario

| Scenario | Risk Level | Suggestion Type | User Action |
|----------|------------|-----------------|-------------|
| Database migration | High | Risk warning + checklist | Confirm precautions |
| Repetitive edits | N/A | Optimization (automation) | Accept/decline script |
| Breaking API change | High | Risk + migration plan | Create v2 or non-breaking |
| Credentials in code | High | Security fix | Auto-extract to .env |
| Prod without staging | High | Risk + checklist | Deploy staging first |
| Sequential tests | N/A | Optimization (parallel) | Enable parallel |
| Test after impl | N/A | Learning (TDD) | Retry with TDD |
| Large file no tests | Medium | Risk + TDD suggestion | Pause, write tests |
| Advanced workflow | N/A | Optimization (custom cmd) | Learn custom automation |

---

**End of Examples** | 2025-11-02
