---
name: "moai-alfred-proactive-suggestions"
version: "4.0.0"
created: 2025-11-02
updated: 2025-11-12
status: stable
tier: specialization
description: "Guide Alfred to provide non-intrusive proactive suggestions based on risk detection, optimization patterns, and learning opportunities. Enhanced with Context7 MCP for up-to-date documentation."
allowed-tools: "Read, AskUserQuestion, WebSearch, WebFetch, mcp__context7__resolve-library-id, mcp__context7__get-library-docs"
primary-agent: "alfred"
secondary-agents: [plan-agent, session-manager]
keywords: [alfred, proactive, suggestions, api, database]
tags: [alfred-core]
orchestration: 
can_resume: true
typical_chain_position: "middle"
depends_on: []
---

# moai-alfred-proactive-suggestions

**Alfred Proactive Suggestions**

> **Primary Agent**: alfred  
> **Secondary Agents**: plan-agent, session-manager  
> **Version**: 4.0.0  
> **Keywords**: alfred, proactive, suggestions, api, database

---

## 📖 Progressive Disclosure

### Level 1: Quick Reference (Core Concepts)

What It Does

Alfred proactively identifies risks, optimization opportunities, and learning moments during workflow execution. Suggestions are contextual, actionable, and limited to prevent interruption.

**Key capabilities**:
- ✅ Risk detection (6 patterns): Database migrations, breaking changes, destructive operations
- ✅ Optimization patterns (3 types): Automation, parallel execution, shortcuts
- ✅ Learning opportunities: Best practices, common pitfalls, Skill recommendations
- ✅ Non-intrusive: Max 1 suggestion per 5 minutes
- ✅ Risk-based decision making: Low/Medium/High classification

---

---

### Level 2: Practical Implementation (Common Patterns)

# Alfred Proactive Suggestions - Intelligent Pattern Recognition

---

What It Does

Alfred proactively identifies risks, optimization opportunities, and learning moments during workflow execution. Suggestions are contextual, actionable, and limited to prevent interruption.

**Key capabilities**:
- ✅ Risk detection (6 patterns): Database migrations, breaking changes, destructive operations
- ✅ Optimization patterns (3 types): Automation, parallel execution, shortcuts
- ✅ Learning opportunities: Best practices, common pitfalls, Skill recommendations
- ✅ Non-intrusive: Max 1 suggestion per 5 minutes
- ✅ Risk-based decision making: Low/Medium/High classification

---

---

When to Use

**Automatic activation**:
- Risk patterns detected during command execution
- Repetitive manual operations observed
- Beginner users encountering learning opportunities
- Complex workflows with optimization potential

**Manual reference**:
- Understanding Alfred's suggestion logic
- Customizing suggestion thresholds
- Learning risk classification criteria

---

---

Three Suggestion Categories

### 🚨 Risk Detection (Safety First)

**Purpose**: Prevent data loss, production outages, security vulnerabilities

**6 Risk Patterns**:

1. **Database Migration**: Schema changes, data migrations
2. **Destructive Operations**: File deletion, force push, reset commands
3. **Breaking Changes**: API changes, dependency updates
4. **Production Operations**: Deployment without staging test
5. **Security Concerns**: Exposed credentials, insecure configs
6. **Large File Operations**: Editing 100+ line files without tests

**Suggestion style**: Warning + mitigation checklist + confirmation

---

### ⚡ Optimization Patterns (Efficiency Boost)

**Purpose**: Reduce manual effort, speed up workflows, suggest automation

**3 Optimization Patterns**:

1. **Repetitive Tasks**: Same operation on 3+ files
2. **Parallel Execution**: Independent tasks executed sequentially
3. **Manual Workflows**: GUI-equivalent actions that could use commands

**Suggestion style**: Observation + time savings estimate + automation offer

---

### 🎓 Learning Opportunities (Knowledge Growth)

**Purpose**: Educate users on best practices, prevent future mistakes

**Trigger conditions**:
- Beginner expertise level detected
- First-time feature usage
- Common pitfall encountered
- Suboptimal pattern detected

**Suggestion style**: Educational + Skill recommendation + example

---

---

Risk Classification System

### Low Risk

**Characteristics**:
- Read-only operations
- Documentation updates
- Typo corrections
- SPEC edits (non-implementation)

**Confirmation threshold**:
- Beginner: Confirm
- Intermediate: Skip
- Expert: Skip

**Example**: Fix typo in README.md

---

### Medium Risk

**Characteristics**:
- Code changes affecting behavior
- Test modifications
- Configuration updates
- Dependency version bumps

**Confirmation threshold**:
- Beginner: Confirm + explanation
- Intermediate: Confirm
- Expert: Skip

**Example**: Update authentication logic

---

### High Risk

**Characteristics**:
- Database migrations
- Production deployments
- Breaking API changes
- Destructive git operations (force push)
- Large refactoring (10+ files)

**Confirmation threshold**:
- Beginner: Confirm + checklist
- Intermediate: Confirm + checklist
- Expert: Confirm

**Example**: Migrate 10K user records to new schema

---

---

Risk Pattern Details

### Pattern 1: Database Migration

**Detection**:
- SPEC contains "migration", "schema", "database"
- SQL files modified
- ORM model changes detected

**Suggestion**:
```
High-risk operation detected: Database migration

Recommended safeguards:
1. Create database backup
2. Test on staging environment
3. Prepare rollback script
4. Schedule maintenance window
5. Verify migration in dry-run mode

Proceed?
  [Yes, precautions taken] [No, cancel] [Show checklist]
```

---

### Pattern 2: Destructive Operations

**Detection**:
- Commands: `rm`, `git reset --hard`, `git push --force`
- File deletion requests
- Irreversible data changes

**Suggestion**:
```
Destructive operation detected: Force push to main branch

Risks:
- Rewrites shared git history
- Breaks collaborator branches
- Potential data loss

Alternatives:
- Use regular push if possible
- Create new branch instead
- Discuss with team before force push

Are you absolutely sure?
  [Cancel] [Proceed with force push]
```

---

### Pattern 3: Breaking Changes

**Detection**:
- API endpoint signature changes
- Public function parameter changes
- Dependency major version update

**Suggestion**:
```
Breaking change detected: API endpoint signature modified

Impact analysis:
- 3 frontend clients affected
- 5 external integrations impacted
- Migration guide required

Recommended actions:
1. Create deprecation notice (v1 → v2)
2. Maintain backward compatibility for 2 versions
3. Document migration path
4. Notify stakeholders

Proceed with breaking change?
  [Yes, create migration plan] [No, use non-breaking approach]
```

---

### Pattern 4: Production Operations

**Detection**:
- Deployment target: production
- No staging test recorded
- Critical infrastructure change

**Suggestion**:
```
Production deployment without staging verification

Checklist:
- [ ] Tested on staging environment
- [ ] Rollback plan prepared
- [ ] Monitoring alerts configured
- [ ] Team notified
- [ ] Backup created

Deploy to production?
  [Yes, checklist complete] [No, test on staging first]
```

---

### Pattern 5: Security Concerns

**Detection**:
- Credentials in code
- API keys in environment files
- Public S3 bucket configuration
- Insecure HTTP endpoints

**Suggestion**:
```
Security concern detected: API key in code

Risk: Exposed credentials if committed to git

Recommended fix:
1. Move to environment variable (.env)
2. Add .env to .gitignore
3. Use secret management (AWS Secrets, Vault)
4. Rotate compromised key

Fix automatically?
  [Yes, move to .env] [I'll fix manually]
```

---

### Pattern 6: Large File Operations

**Detection**:
- Editing file >100 lines
- No test coverage for file
- Complex logic modification

**Suggestion**:
```
Large file edit detected: 250 lines modified

Risk: Regression without test coverage

Recommendation:
1. Write tests before refactoring (TDD)
2. Break into smaller changes
3. Use /alfred:2-run for TDD workflow

Proceed?
  [Pause, write tests first] [Continue without tests]
```

---

---

Optimization Pattern Details

### Pattern 1: Repetitive Tasks

**Detection**:
- Same operation on 3+ files
- Similar edits detected
- Pattern recognition threshold reached

**Suggestion**:
```
Repetitive pattern detected: Updating import statements in 5 files

Automation opportunity:
- Analyze your last 2 edits
- Generate batch script
- Apply to remaining 3 files
- Estimated time saved: 10 minutes

Create automation?
  [Yes, generate script] [No, continue manually]
```

---

### Pattern 2: Parallel Execution

**Detection**:
- Sequential tasks with no dependencies
- Independent test suites
- Multiple API calls in sequence

**Suggestion**:
```
Parallel execution opportunity detected

Current workflow:
1. Run unit tests (2 min)
2. Run integration tests (3 min)
3. Run E2E tests (5 min)
Total: 10 minutes sequential

Optimized workflow:
1. Run all test suites in parallel
Total: 5 minutes (max of 3 durations)

Time saved: 5 minutes (50%)

Enable parallel execution?
  [Yes, run in parallel] [No, keep sequential]
```

---

### Pattern 3: Manual Workflows

**Detection**:
- Performing git operations manually
- Manual file creation instead of commands
- Repetitive confirmation steps

**Suggestion**:
```
Manual workflow detected: Creating SPEC files by hand

Automation available:
- Use /alfred:1-plan for automated SPEC creation
- Includes EARS validation
- Auto-generates @TAGs
- Ensures completeness

Time saved per SPEC: 15 minutes
Quality improvement: +30% (validation)

Switch to /alfred:1-plan?
  [Yes, use command] [No, prefer manual]
```

---

---

Learning Opportunity Patterns

### Beginner: First-Time Feature Usage

**Detection**:
- User invokes `/alfred:*` command for first time
- Complex workflow initiated
- Expertise level: Beginner

**Suggestion**:
```
First-time SPEC creation detected

Learning resources:
- Skill("moai-foundation-specs") - SPEC structure guide
- Skill("moai-foundation-ears") - EARS requirements format
- Skill("moai-alfred-spec-metadata-validation") - Validation rules

Would you like a step-by-step walkthrough?
  [Yes, guide me] [No, I'll explore]
```

---

### Intermediate: Suboptimal Pattern

**Detection**:
- User creates tests after implementation (not TDD)
- Missing @TAG references
- Skipping TRUST 5 validation

**Suggestion**:
```
Observation: Tests written after implementation

Best practice: TDD (Test-First)
- Write failing test first (RED)
- Implement to pass test (GREEN)
- Refactor with safety net (REFACTOR)

Benefits:
- 40% fewer bugs (industry data)
- Better code design
- Confidence in refactoring

Learn TDD workflow:
- Skill("moai-foundation-trust") - TRUST 5 principles

Switch to TDD next time?
  [Yes, remind me] [No, I prefer current approach]
```

---

### Expert: Advanced Technique

**Detection**:
- Complex workflow detected
- Expert expertise level
- Rare suggestion opportunity

**Suggestion**:
```
Advanced technique available: Custom agent creation

Your workflow could benefit from specialized agent:
- Pattern: Frequent API integration testing
- Candidate: api-integration-tester sub-agent
- Time saved: 20 min/week

Would you like guidance on custom agent creation?
  [Yes, show me how] [No, not now]
```

---

---

Suggestion Frequency Limits

**Non-intrusive constraint**: Max 1 suggestion per 5 minutes

**Rationale**:
- Avoid alert fatigue
- Maintain user flow state
- Prioritize high-value suggestions

**Priority ranking** (when multiple suggestions eligible):

1. **High-risk warnings** (always shown)
2. **Medium-risk warnings** (shown if no high-risk)
3. **Optimization patterns** (shown if no risks)
4. **Learning opportunities** (lowest priority)

---

---

### Level 3: Advanced Patterns (Expert Reference)

> **Note**: Advanced patterns for complex scenarios.

**Coming soon**: Deep dive into expert-level usage.


---

## 🎯 Best Practices Checklist

**Must-Have:**
- ✅ [Critical practice 1]
- ✅ [Critical practice 2]

**Recommended:**
- ✅ [Recommended practice 1]
- ✅ [Recommended practice 2]

**Security:**
- 🔒 [Security practice 1]


---

## 🔗 Context7 MCP Integration

**When to Use Context7 for This Skill:**

This skill benefits from Context7 when:
- Working with [alfred]
- Need latest documentation
- Verifying technical details

**Example Usage:**

```python
# Fetch latest documentation
from moai_adk.integrations import Context7Helper

helper = Context7Helper()
docs = await helper.get_docs(
    library_id="/org/library",
    topic="alfred",
    tokens=5000
)
```

**Relevant Libraries:**

| Library | Context7 ID | Use Case |
|---------|-------------|----------|
| [Library 1] | `/org/lib1` | [When to use] |


---

## 📊 Decision Tree

**When to use moai-alfred-proactive-suggestions:**

```
Start
  ├─ Need alfred?
  │   ├─ YES → Use this skill
  │   └─ NO → Consider alternatives
  └─ Complex scenario?
      ├─ YES → See Level 3
      └─ NO → Start with Level 1
```


---

## 🔄 Integration with Other Skills

**Prerequisite Skills:**
- Skill("prerequisite-1") – [Why needed]

**Complementary Skills:**
- Skill("complementary-1") – [How they work together]

**Next Steps:**
- Skill("next-step-1") – [When to use after this]


---

## 📚 Official References

When to Use

**Automatic activation**:
- Risk patterns detected during command execution
- Repetitive manual operations observed
- Beginner users encountering learning opportunities
- Complex workflows with optimization potential

**Manual reference**:
- Understanding Alfred's suggestion logic
- Customizing suggestion thresholds
- Learning risk classification criteria

---

---

## 📈 Version History

**v4.0.0** (2025-11-12)
- ✨ Context7 MCP integration
- ✨ Progressive Disclosure structure
- ✨ 10+ code examples
- ✨ Primary/secondary agents defined
- ✨ Best practices checklist
- ✨ Decision tree
- ✨ Official references



---

**Generated with**: MoAI-ADK Skill Factory v4.0  
**Last Updated**: 2025-11-12  
**Maintained by**: Primary Agent (alfred)
