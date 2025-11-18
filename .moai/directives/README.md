---
title: "MoAI-ADK Directive Guides"
version: "1.0.0"
updated: "2025-11-19"
audience: "Everyone (users, implementers, support)"
scope: "Index and guide to directive documents"
---

# MoAI-ADK Directive Guides

**Official directives defining HOW commands and features should work, not implementation details.**

---

## Overview

Directive guides are **prescriptive specifications** of user experience and behavior, from both user and implementer perspectives. They define:

- ✅ What the command should do
- ✅ How users should experience it
- ✅ What success looks like
- ✅ How errors should be handled
- ✅ What to do when things don't work

**They do NOT contain**:
- ❌ Code implementation details
- ❌ Internal algorithms
- ❌ Technical architecture
- ❌ Programming language specifics

---

## Document Structure

Each directive document has a clear audience and purpose:

### 1. **0-project-command-directive.md** (Comprehensive)

**Audience**: Implementers, architects, advanced users

**Size**: ~800 lines

**Contents**:
- Core philosophy and principles
- Smart entry point directive
- Essential configuration directives
- User interaction model
- Configuration philosophy
- Phase-based guidelines
- Critical rules
- Implementation checklist
- Success metrics

**When to read**:
- Implementing the `/moai:0-project` command
- Understanding design decisions
- Training developers
- Creating new commands (as template)

**Key sections**:
- Section 1: Core Philosophy (8 core directives)
- Section 2-5: Phase breakdown (Entry → Config → Completion)
- Section 6-7: Rules and documentation
- Section 10: Implementation checklist

---

### 2. **0-project-executive-summary.md** (Quick Reference)

**Audience**: Project managers, product leads, quick reference users

**Size**: ~400 lines

**Contents**:
- What the command does (30-second version)
- When to use it (by scenario)
- How it works (user experience flow)
- Key features (5 main features)
- Configuration outcomes
- Success criteria
- Error handling philosophy
- Technical architecture (high-level)
- User journey examples
- Time investment breakdown

**When to read**:
- Quick understanding of `/moai:0-project`
- Explaining feature to stakeholders
- Planning implementation timeline
- Setting success metrics

**Key sections**:
- What It Is (30-second overview)
- How Does It Work (5-step user flow)
- Key Features (5 main selling points)
- Success Criteria (user vs. system)
- User Journey Examples (3 realistic scenarios)

---

### 3. **0-project-error-recovery-guide.md** (Detailed Handling)

**Audience**: Implementers, support teams, quality engineers

**Size**: ~700 lines

**Contents**:
- Error classification system (by severity)
- 6 error categories with 20+ specific error types
- Detection strategies
- Recovery steps for each error
- User messages (by language)
- Prevention strategies
- Troubleshooting decision tree
- Error message quality standards
- Testing checklist
- Logging and debugging

**When to read**:
- Implementing error handling
- Creating support documentation
- Testing error paths
- Handling user support requests
- Setting up logging strategy

**Key sections**:
- Error Classification System (severity levels)
- Error Categories 1-6 (detailed error specs)
- Error Prevention Strategies (4 strategies)
- Troubleshooting Decision Tree (visual flow)
- Testing Checklist (16 test cases)

---

## Quick Navigation

### For Different Roles

**Product Manager**:
1. Read: Executive Summary (15 min)
2. Review: Success Criteria section
3. Check: Timeline estimates

**Implementer**:
1. Read: Full Directive (60 min)
2. Read: Error Recovery Guide (30 min)
3. Use: Implementation Checklist
4. Reference: As you code

**Support Team**:
1. Read: Executive Summary (15 min)
2. Read: Error Recovery Guide, troubleshooting section (20 min)
3. Bookmark: Error Message Quality Standards
4. Reference: During support tickets

**QA Engineer**:
1. Read: Full Directive, Critical Rules section (20 min)
2. Read: Error Recovery Guide, Testing Checklist (30 min)
3. Use: Testing Checklist for test planning
4. Reference: User journey examples

**End User**:
1. Read: Executive Summary, "When Should Users Run It" section
2. Read: User journey examples relevant to their scenario
3. Reference: CLAUDE.md for quick steps

---

## Directive Principles

All directives follow these principles:

### 1. Language-First Design
- Directives are written to be understood by non-technical readers
- Technical content includes context and explanation
- Examples use simple, clear language

### 2. User-Centric Perspective
- Written from user's point of view first
- User experience takes priority
- Implementation details are secondary

### 3. Actionable Content
- Every section has clear actions or outcomes
- Decision trees are provided where relevant
- Success criteria are explicit and measurable

### 4. Progressive Disclosure
- Essential content first
- Recommended content next
- Advanced content last
- Reader can stop at any level and have complete picture

### 5. Prescriptive, Not Descriptive
- Directives say "MUST do X" (not "might consider X")
- Clear rules and exceptions
- Standards are enforced through checklists

### 6. Completeness
- Covers happy path and all error paths
- Includes recovery strategies
- Anticipates questions users might have

---

## How to Use These Documents

### Creating a New Feature

1. **Review Full Directive** (60 min)
   - Understand what should happen
   - Learn critical rules and constraints
   - Identify success criteria

2. **Create Implementation Plan**
   - Use "Implementation Checklist" as template
   - Map features to code components
   - Define test scenarios

3. **Implement**
   - Follow directives closely
   - Reference error handling guide
   - Validate against checklist

4. **Test**
   - Use testing checklist from error guide
   - Test all error paths
   - Verify user experience matches summary

### Training Team Members

1. **New Implementer**:
   - Read: Full Directive (60 min)
   - Read: Error Recovery Guide (30 min)
   - Discuss: Philosophy section together
   - Review: Critical Rules section
   - Practice: Implementation checklist

2. **New Support Person**:
   - Read: Executive Summary (15 min)
   - Read: Error Recovery Guide (30 min)
   - Study: Error Message Quality Standards
   - Practice: Troubleshooting decision tree
   - Roleplay: Common error scenarios

3. **New QA Engineer**:
   - Read: Full Directive, Critical Rules (20 min)
   - Read: Error Recovery Guide (30 min)
   - Study: Success Metrics section
   - Build: Test plan from checklist
   - Execute: Test scenarios and edge cases

### During Development

**Implementers**:
- Reference: Section 3-5 (Phase directives)
- Check: Section 7 (Critical Rules) before every code change
- Use: Implementation Checklist to verify completion

**QA Engineers**:
- Use: Testing Checklist as test plan
- Reference: Error categories for edge cases
- Verify: Success Criteria are met

**Support Team**:
- Reference: Error Message Quality Standards when translating
- Use: Troubleshooting Decision Tree when helping users
- Check: Recovery steps for accuracy

---

## Relationship to Other Documents

### CLAUDE.md
- **CLAUDE.md**: Quick reference, 3 levels of detail (5/15/30 min)
- **Directives**: Comprehensive specifications, complete coverage

**Usage**:
- CLAUDE.md: User refers to for quick reminder
- Directives: Developer refers to for implementation detail

### .moai/memory/ Files
- **Memory files**: Extended context on specific topics
- **Directives**: Prescriptive specifications for behavior

**Usage**:
- Memory files: How-to and background information
- Directives: What the system should do and how to know it works

### Agent/Skill Documentation
- **Agent docs**: What an agent can do, examples
- **Directives**: What an agent MUST do, error handling

**Usage**:
- Agent docs: Reference during agent use
- Directives: Requirements that agents must meet

---

## Version Control & Updates

**Current Version**: 1.0.0 (2025-11-19)

**When to Update Directives**:
- Major feature changes
- New error categories discovered
- User experience improvements
- Process changes
- Rule clarifications

**How to Update**:
1. Update relevant directive document
2. Update version number
3. Document change in git commit message
4. Notify relevant teams (implementers, support, QA)

**Backward Compatibility**:
- Current implementations follow these directives
- Changes to directives may require code updates
- Version field in each document for tracking

---

## Checklist for Using Directives

### Before Implementing

- [ ] Read full directive for the feature
- [ ] Understand core philosophy (Section 1)
- [ ] Review success criteria
- [ ] Read error recovery guide
- [ ] Create implementation checklist
- [ ] Identify skills/agents needed
- [ ] Plan test scenarios

### During Implementation

- [ ] Reference phase directives
- [ ] Check critical rules before commits
- [ ] Verify error handling completeness
- [ ] Test against checklist
- [ ] Validate user experience

### Before Release

- [ ] All checklist items completed
- [ ] All error paths tested
- [ ] User experience matches summary
- [ ] Success criteria verified
- [ ] Support team trained
- [ ] Documentation updated

---

## FAQ About Directives

**Q: Why are there separate executive summary and full directive?**

A: Different audiences need different levels of detail. Executive summary is 400 lines for quick understanding. Full directive is 800 lines with all implementation detail. QA uses error recovery guide (700 lines). Together they provide complete specification.

**Q: Can I use directives as user documentation?**

A: Executive summary is good for users. Full directive is technical (for implementers). Use CLAUDE.md and .moai/memory/ files for user documentation.

**Q: What if directive conflicts with my code?**

A: Directives are the specification. If your code conflicts, the code needs to change. Report the conflict so we can understand the issue and update directive if needed.

**Q: How detailed should directives be?**

A: Specific enough to implement without ambiguity. Not so detailed that they become code comments. Include examples and decision trees where helpful.

**Q: Who writes directives?**

A: Directives are written collaboratively:
- Product lead: Features and user experience
- Lead developer: Technical feasibility
- QA lead: Testing and edge cases
- Support lead: Error handling and recovery

---

## Directory Structure

```
.moai/directives/
├── README.md (this file)
├── 0-project-command-directive.md (full specification)
├── 0-project-executive-summary.md (quick reference)
├── 0-project-error-recovery-guide.md (error handling)
├── 1-plan-command-directive.md (future)
├── 2-run-command-directive.md (future)
├── 3-sync-command-directive.md (future)
└── [feature]-directive.md (as features are added)
```

---

## Maintenance

**These directives are living documents**. They evolve as:
- User needs change
- New error patterns emerge
- Better solutions are discovered
- Features are enhanced
- Teams provide feedback

**Review cycle**: Quarterly review with team
**Update policy**: Changes require consensus
**Version tracking**: Every change documented

---

## Summary

Use these directives to:
- ✅ Understand how features should work
- ✅ Implement features correctly
- ✅ Handle errors gracefully
- ✅ Provide good support
- ✅ Know when implementation is complete
- ✅ Test comprehensively

These are the **source of truth** for feature behavior in MoAI-ADK.

---

**Document Version**: 1.0.0
**Last Updated**: 2025-11-19
**Status**: Ready for use
**Maintenance**: Reviewed quarterly

**Questions?** Refer to relevant directive section or contact project leadership.
