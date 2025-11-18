---
title: "Directive System Quick Start"
version: "1.0.0"
date: "2025-11-19"
audience: "First-time users of directive system"
---

# Directive System Quick Start

**Get oriented in 5 minutes. Find what you need in 30 seconds.**

---

## What Are These Documents?

**Directives** are specifications of how commands and features SHOULD WORK:
- ✅ What the feature does
- ✅ How users experience it
- ✅ What success looks like
- ✅ How errors are handled
- ✅ What to do when things break

These are NOT code. These are requirements that code must meet.

---

## Where to Start (By Role)

### I'm a Project Manager
**Time**: 15 minutes
**Start here**: `0-project-executive-summary.md`
**Then read**: README.md section "Success Criteria"

**What you'll learn**:
- What this feature does
- How long it takes users
- What success looks like
- How to plan timeline

---

### I'm a Developer Implementing This
**Time**: 90 minutes
**Start here**: `0-project-command-directive.md` (all sections)
**Then read**: `0-project-error-recovery-guide.md` (implementation checklist)

**What you'll learn**:
- Exact requirements
- What each section should do
- How to handle errors
- Implementation checklist to verify

---

### I'm a QA/Test Engineer
**Time**: 60 minutes
**Start here**: `0-project-error-recovery-guide.md` section "Testing Checklist"
**Then read**: `0-project-command-directive.md` section "Success Metrics"

**What you'll learn**:
- 16 test scenarios to run
- What success looks like
- All error conditions
- How to validate user experience

---

### I'm Supporting Users
**Time**: 30 minutes
**Start here**: `0-project-executive-summary.md` sections "What Is This Command" + "When Should Users Run It"
**Then read**: `0-project-error-recovery-guide.md` section "Troubleshooting Decision Tree"

**What you'll learn**:
- What to tell users when they ask
- How to troubleshoot errors
- Clear recovery steps
- Error message standards

---

### I'm Onboarding to the Team
**Time**: 2 hours
**Start here**: `README.md` (15 min)
**Then read in order**:
1. `0-project-executive-summary.md` (15 min)
2. `0-project-command-directive.md` section 1 "Core Philosophy" (20 min)
3. `0-project-command-directive.md` section 7 "Critical Rules" (10 min)
4. `0-project-error-recovery-guide.md` (30 min)

**What you'll learn**:
- How system works
- Core principles
- Non-negotiable rules
- Error handling approach

---

## 30-Second Reference

### "What is /moai:0-project?"
→ See: `0-project-executive-summary.md`, first paragraph

### "How does user experience this?"
→ See: `0-project-executive-summary.md`, section "How Does It Work"

### "What are the critical rules?"
→ See: `0-project-command-directive.md`, section 7

### "What happens if [error]?"
→ See: `0-project-error-recovery-guide.md`, section "Error Categories"

### "How do I test this?"
→ See: `0-project-error-recovery-guide.md`, section "Testing Checklist"

### "How do I implement this?"
→ See: `0-project-command-directive.md`, section 10 "Implementation Checklist"

### "How do I know when done?"
→ See: `0-project-command-directive.md`, section 11 "Success Metrics"

### "What happens if user does [X]?"
→ See: `0-project-command-directive.md`, section 3-5 "Phase Directives"

---

## Document Map

```
START HERE: README.md (navigation guide)
    ↓
Choose your path:
    ├─→ Project Manager: executive-summary.md
    ├─→ Developer: full-directive.md + error-guide.md
    ├─→ QA: error-guide.md + testing checklist
    ├─→ Support: executive-summary.md + error-guide troubleshooting
    └─→ New Team Member: Read all (2 hours)
```

---

## Key Documents at a Glance

| Document | Length | Best For | Time |
|----------|--------|----------|------|
| README.md | 433 L | Navigation | 10 min |
| Executive Summary | 441 L | Quick understanding | 15 min |
| Full Directive | 1,154 L | Implementation | 60 min |
| Error Guide | 1,212 L | Error handling + testing | 30 min |
| Delivery Summary | 447 L | Overview of package | 5 min |

---

## Core Principles (Read These First)

**If you only read 5 minutes, read this**:

### Principle 1: Language-First
→ User's language is the primary setting. Everything else in their language.

### Principle 2: Zero Direct Tools
→ Command uses ONLY Task() and AskUserQuestion(). NO Read/Write/Edit/Bash.

### Principle 3: Complete Delegation
→ Command orchestrates. Agents execute. Skills do file operations.

### Principle 4: Atomic Updates
→ All changes together or none. Backup before write. Rollback available.

### Principle 5: User Confirmation
→ Breaking changes need confirmation. Clear recovery paths for errors.

---

## Critical Rules (Check Before Coding)

**These are non-negotiable. Check them before every commit**:

✓ Execute ONLY ONE mode per invocation
✓ Read language from config before starting
✓ Use only Task() and AskUserQuestion()
✓ Validate at 3 checkpoints (Tab 1, Tab 3, before write)
✓ Create backup before any write
✓ Show user exact changes made
✓ No emojis in interactive fields
✓ All output in user's language
✓ Every error has recovery path
✓ Never leave partial config state

→ See: `0-project-command-directive.md` section 7 for full list

---

## Success Checklist (How to Know When Done)

**Implementation is complete when**:

From Full Directive Implementation Checklist:
- [ ] All phases understand mode detection logic
- [ ] All phases use only Task() and AskUserQuestion()
- [ ] Language handling is consistent
- [ ] Error messages translate to user's language
- [ ] Skill dependencies documented
- [ ] Validation at 3 checkpoints
- [ ] No emojis in interactive fields
- [ ] Tab schema format validated
- [ ] Backup/rollback tested
- [ ] Context save mechanism specified

From Error Guide Testing Checklist:
- [ ] Config missing (first time) → works
- [ ] Config corrupted → recovery works
- [ ] Invalid language → user can fix
- [ ] Git conflict → user sees suggestion
- [ ] All 16 error scenarios tested
- [ ] User can resume from interruption
- [ ] Backups and rollbacks work
- [ ] All error paths tested

---

## Common Questions

**Q: Which document should I read first?**
A: Start with README.md for navigation. Then go to your role-specific document.

**Q: Can I skip some sections?**
A: Yes. README.md shows what each section covers. Skip what's not relevant to you.

**Q: How detailed are these?**
A: Executive Summary is conceptual. Full Directive is implementation-ready. Error Guide has specific scenarios.

**Q: What if my code conflicts with a directive?**
A: Directives are specification. Code must change. Report conflict so we can clarify.

**Q: How do I update these?**
A: See README.md section "Version Control & Updates"

**Q: Where's the code?**
A: These are specifications, not code. Code should follow these directives.

---

## Directive Quick Reference

### When User Runs /moai:0-project (No Arguments)
→ Check: `.moai/config/config.json` exists?
→ If NO: Run INITIALIZATION mode (new project setup)
→ If YES: Run AUTO-DETECT mode (show current status)

### When User Runs /moai:0-project setting
→ Run SETTINGS mode (tab-based config editor)
→ Load 5 tabs, execute selected tab(s)
→ Validate at checkpoints

### When User Runs /moai:0-project update
→ Run UPDATE mode (smart template merging)
→ Preserve language from existing config
→ Auto-translate announcements

### When Error Occurs
→ Show error in user's language
→ Explain what went wrong
→ Offer recovery options
→ Log for debugging

---

## Finding Specific Information

### "I need to know..."

**How mode detection works**
→ `0-project-command-directive.md` section 2 + section 12 (decision tree)

**What Tab 1 should do**
→ `0-project-command-directive.md` section 3.3 (Tab 1 example)

**What Tab 3 validation is**
→ `0-project-command-directive.md` section 3.3 (Tab 3 validation example)

**How to handle invalid language**
→ `0-project-error-recovery-guide.md` section 2.1 + troubleshooting tree

**How to handle git conflict**
→ `0-project-error-recovery-guide.md` section 2.3 + recovery steps

**How to test everything**
→ `0-project-error-recovery-guide.md` section "Testing Checklist"

**How to implement completely**
→ `0-project-command-directive.md` section 10 "Implementation Checklist"

**What success looks like**
→ `0-project-command-directive.md` section 11 "Success Metrics"

**What to tell users**
→ `0-project-executive-summary.md` section "User Journey Examples"

---

## File Locations

**All in**: `/Users/goos/MoAI/MoAI-ADK/.moai/directives/`

**Start with**: `README.md`

**Then use**: Whichever document is relevant to your role

---

## Next Steps

### If You're Implementing
1. Read `0-project-command-directive.md` (1 hour)
2. Review `Implementation Checklist` section
3. Create detailed plan
4. Start coding, reference directives daily
5. Run testing checklist before submitting
6. Verify against success metrics

### If You're Testing
1. Read `0-project-error-recovery-guide.md` "Testing Checklist" (20 min)
2. Create test cases from checklist
3. Execute 16 test scenarios
4. Document any deviations
5. Verify user experience matches summary

### If You're Supporting Users
1. Read `0-project-executive-summary.md` (15 min)
2. Bookmark `Troubleshooting Decision Tree` (error guide)
3. Save `Error Message Quality Standards` as reference
4. Create FAQ from directives
5. Use decision tree to troubleshoot issues

### If You're Managing/Leading
1. Read `README.md` (15 min)
2. Read `0-project-executive-summary.md` (15 min)
3. Review `Implementation Checklist` for timeline
4. Share executive summary with stakeholders
5. Use success metrics to track progress

---

## Getting Help

**Question about directives?**
→ See `README.md` section FAQ

**Can't find something?**
→ Use README.md navigation table

**Need clarification?**
→ Check related sections (referenced in each document)

**Found an error in directives?**
→ Report to project leadership with specific location

---

## Summary

This directive package has **everything you need** to:
- ✅ Understand what feature does
- ✅ Implement it correctly
- ✅ Test it thoroughly
- ✅ Support users
- ✅ Know when you're done

**Start with**: Your role's quick-start path above
**Reference**: Specific sections as needed
**Verify**: Against checklists before shipping

You're all set!

---

**Quick Start Version**: 1.0.0
**Date**: 2025-11-19
**Status**: Ready to use

**All directives in**: `.moai/directives/`
