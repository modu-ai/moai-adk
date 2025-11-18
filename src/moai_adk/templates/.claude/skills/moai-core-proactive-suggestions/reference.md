# Alfred Proactive Suggestions - Quick Reference

> **Main Skill**: [SKILL.md](SKILL.md)  
> **Examples**: [examples.md](examples.md)

---

## Risk Classification Quick Reference

| Risk Level | Operations | Beginner | Intermediate | Expert |
|------------|------------|----------|--------------|--------|
| **Low** | Read-only, docs, typos | Confirm | Skip | Skip |
| **Medium** | Code changes, config | Confirm + explain | Confirm | Skip |
| **High** | DB migration, prod deploy | Confirm + checklist | Confirm + checklist | Confirm |

---

## Risk Pattern Catalog

| Pattern | Keywords | Risk Level | Mitigation |
|---------|----------|------------|------------|
| **Database Migration** | migration, schema, ALTER | High | Backup, staging, rollback |
| **Destructive Ops** | rm, force push, reset | High | Alternative suggestion |
| **Breaking Changes** | API change, major version | High | Migration plan |
| **Production Deploy** | deploy, prod, release | High | Staging checklist |
| **Security** | credentials, API key | High | Environment vars |
| **Large File Edit** | >100 lines, no tests | Medium | TDD workflow |

---

## Optimization Pattern Catalog

| Pattern | Detection | Time Saved | Automation |
|---------|-----------|------------|------------|
| **Repetitive Tasks** | 3+ similar edits | 10-30 min | Batch script |
| **Parallel Execution** | Sequential independent tasks | 30-50% | Parallel flag |
| **Manual Workflows** | GUI-equivalent actions | 15-20 min | `/alfred:*` command |

---

## Suggestion Priority Matrix

**When multiple suggestions eligible, prioritize**:

1. **High-risk warnings** (always show)
2. **Medium-risk warnings** (if no high-risk)
3. **Optimization patterns** (if no risks)
4. **Learning opportunities** (lowest priority)

**Frequency limit**: Max 1 suggestion per 5 minutes

---

## Expertise-Based Thresholds

| Expertise | Suggestions/Session | Focus | Pattern Threshold |
|-----------|---------------------|-------|-------------------|
| **Beginner** | 3-5 | Learning + risks | Low (suggest common) |
| **Intermediate** | 2-3 | Optimizations + med risks | Medium (suggest improvements) |
| **Expert** | 1-2 | Advanced + high risks | High (suggest advanced only) |

---

## Decision Tree

```
Operation Detected
    ↓
Classify Risk (Low/Medium/High)
    ↓
Check Expertise Level (Beginner/Int/Expert)
    ↓
Determine Confirmation Threshold
    ↓
├─ High Risk → Always confirm (all levels)
├─ Medium Risk → Confirm if Beginner/Int
└─ Low Risk → Confirm if Beginner only
    ↓
Check Suggestion Frequency (1 per 5 min)
    ↓
├─ Within limit → Show suggestion
└─ Over limit → Queue for later
```

---

## Integration Points

**Called by**:
- All `/alfred:*` commands (risk detection)
- `moai-core-persona-roles` (role adaptation)
- `moai-core-expertise-detection` (threshold tuning)

**Calls**:
- `AskUserQuestion` (confirmation dialogs)
- Risk mitigation Skills (context-specific)

---

**End of Reference** | 2025-11-02
