# Frequently Asked Questions (FAQ)

**Created**: 2025-11-21
**Status**: Production Ready

---

## Alfred's SPEC Decision

### Q1: Does Alfred always make accurate decisions?

**A**: No, it's not 100% accurate. Alfred analyzes based on natural language patterns, so there may be a 5-10% error rate.

However:
- ✅ Major complexity is accurately judged
- ✅ Users can always reject
- ✅ Can propose when complexity increases during implementation
- ✅ Improves with monthly statistics

Example: Even if Alfred says "SPEC unnecessary", users can still create a SPEC if desired.

---

### Q2: Why 5 questions?

**A**: It's a balance between sufficient accuracy and simplicity.

```
3 questions   → Low accuracy (high false positives)
5 questions   → Appropriate balance ← Selected!
7+ questions  → Excessive (increased complexity)
```

5 questions:
- Adequately reflect Alfred's understanding
- Clear from user perspective
- Simple to implement

---

### Q3: How are "Possible" answers handled?

**A**: "Possible" is treated the same as "Yes".

```
0-1 (only No) → SPEC Unnecessary
2-3 (Yes/Possible included) → SPEC Recommended  ← "Possible" included!
4-5 (Yes/Possible included) → SPEC Strongly Recommended
```

Example: 2 "Possible", 1 "No" → 2 met → SPEC Recommended

---

### Q4: Can the decision criteria be customized?

**A**: May be possible in the future, but currently fixed.

Reasons:
- ✅ Ensure standardized workflow
- ✅ All users use same criteria
- ✅ Easy statistical comparison

If needed, users can reject/modify after Alfred's decision.

---

## SPEC Templates

### Q5: Why 3 levels?

**A**: To minimize user burden while including all necessary information.

```
Level 1 (Minimal)      → 5 min writing (optimal for simple tasks)
Level 2 (Standard)     → 10-15 min writing (optimal for general features)
Level 3 (Comprehensive) → 20-30 min writing (optimal for complex tasks)
```

More levels complicate selection.

---

### Q6: Can templates be modified?

**A**: Yes, anytime!

Generated SPECs:
- ✅ Can add sections
- ✅ Can remove sections
- ✅ Can modify content
- ✅ Can change order

But initially, use the template Alfred selects.

---

### Q7: Does Level 1 require writing tests?

**A**: Yes, but written differently.

```
Level 1: Simple unit tests (1-2)
  Example: Check function input/output

Level 2: Unit tests + Integration tests (5-10)
  Example: Check entire workflow

Level 3: Full test stack (20+)
  Example: Unit/Integration/E2E tests
```

Development without tests is prohibited regardless of SPEC.

---

### Q8: Can templates be changed during work?

**A**: Yes, depending on the case.

Example:
```
Initial: Start with Level 1 (Minimal)
→ Detect complexity increase during implementation
→ Can transition to Level 2 (Standard)
```

Alfred will propose transition when detected.

---

## Statistics and Analysis

### Q9: Are statistics accurate?

**A**: Major trends are accurate, but individual items are not 100% reliable.

Reasons:
- ❌ Tasks implemented without SPEC not tracked
- ❌ Not linked without SPEC-ID in Git commit messages
- ❌ Test coverage difficult to measure automatically

However:
- ✅ Overall trends highly reliable
- ✅ Monthly comparisons meaningful
- ✅ Can identify improvement directions

Usage: **Focus on trends rather than individual numbers**

---

### Q10: Can statistics be disabled?

**A**: Yes, possible.

Method:
```bash
# Remove SessionEnd Hook
rm .claude/hooks/sessionend.sh

# Or comment out Hook contents
```

Results:
- ✅ No data collection
- ❌ No monthly report generation
- ❌ Cannot measure effectiveness

Recommendation: Don't disable! (Needed to prove effectiveness)

---

### Q11: Is personal information collected?

**A**: No, only statistics.

Collected data:
- ✅ SPEC ID, creation time
- ✅ Implementation time, status
- ✅ Git commit hash (not message)
- ✅ File paths
- ✅ Test coverage

Not collected:
- ❌ Code contents
- ❌ Personal information
- ❌ Business secrets

Storage location: `.moai/logs/` (local only, no external transmission)

---

### Q12: Who manages monthly reports?

**A**: Automatically generated, no manual management needed.

Automatic:
- ✅ Auto-generated last day of each month
- ✅ Existing reports automatically archived
- ✅ Trends automatically calculated

Recommended:
- Review reports once monthly
- Identify improvements
- Plan for next month

---

## Workflow and Usage

### Q13: Is it okay to reject SPEC proposals?

**A**: Yes, perfectly fine!

Meaning of rejection:
- ✅ No penalty
- ✅ Alfred won't force
- ✅ Respects user judgment

Example:
```
Alfred: "SPEC recommended"
User: "No, implement directly"
→ No problem
```

However, implementation time may increase and bug possibility may rise.

---

### Q14: What's the difference between prototype and production code?

**A**: Alfred handles them differently.

Prototype:
```
User: "I want to quickly make a prototype"
→ Alfred: Skip SPEC
→ Fast iterative development
→ Collect feedback after completion
```

Production transition:
```
User: "Now let's make it production code"
→ Alfred: Re-assess SPEC necessity
→ Generate appropriate SPEC
→ Develop systematically
```

Effect:
- Prototypes are fast
- Production is stable

---

### Q15: Is SPEC mandatory for team projects?

**A**: Yes, **strongly recommended** for team collaboration.

Reasons:
```
Individual work:
  → SPEC optional (efficiency focused)

Team collaboration (2+ people):
  → SPEC required (high coordination cost)
```

Effects:
- ✅ Unified understanding among team members
- ✅ Reduced integration errors
- ✅ Shortened code review time
- ✅ Clear progress tracking

Rule: **Larger team size = Higher SPEC importance**

---

### Q16: SPEC writing takes too long

**A**: Utilize 80% automatic generation by AI!

Timeline:
```
AI generates 80% automatically: 5-10 min
User modifies 20%: 1-5 min
Total time: 10-15 min
```

Acceleration methods:
1. Clarify only SPEC title and objectives
2. AI automatically writes the rest
3. User reviews/modifies

Example:
```
"SPEC-005: Payment Module Refactoring"
"Objective: 50% processing time reduction"

→ AI automatically generates evaluation criteria, analysis, recommendations
→ User reviews in just 5 minutes
```

---

## Migration & Changes

### Q17: Can it be applied to existing projects?

**A**: Yes, and almost immediate application possible.

Steps:
```
1. Load Skill (already done)
2. Update CLAUDE.md (30 min)
3. Set up Hooks (15 min)
4. Initialize statistics data (5 min)

Total time: 1 hour
```

Existing SPECs:
- ✅ Can maintain as is
- ✅ Compatible with new system
- ✅ Gradual migration

---

### Q18: Do existing SPEC formats need to change?

**A**: No, no change needed.

Compatibility:
- ✅ Level 1/2/3 templates optional
- ✅ Existing SPECs remain valid
- ✅ Only new SPECs use new templates

Gradual transition:
```
Existing: Free-form SPECs (100)
→ Keep as is

New: 3-level template SPECs (created going forward)
→ Apply new format

Mixed: No problem (high compatibility)
```

---

### Q19: Can other teams use this system?

**A**: Yes, perfectly possible!

This Skill:
- ✅ Independent of MoAI-ADK
- ✅ Applicable to other projects
- ✅ Alfred concept reusable

Sharing method:
```
1. Copy Skill path to other projects
2. Add section to CLAUDE.md
3. Initialize statistics

→ Immediately usable
```

---

## Troubleshooting

### Q20: SPEC generation failed

**A**: Check the following:

Checklist:
```
1. Confirmed /moai:1-plan "description" command input?
2. spec-builder agent available?
3. Skill loaded in CLAUDE.md?
4. .moai/specs/ directory exists?
```

Resolution:
- Verify SPEC ID auto-generation
- Check specific error messages
- Report via `/moai:9-feedback`

---

### Q21: Statistics not being collected

**A**: Verify Hook is executing:

Verification:
```bash
# Check Hook file exists
ls -la .claude/hooks/sessionend.sh

# Check execute permission
chmod +x .claude/hooks/sessionend.sh

# Check logs
cat .moai/logs/spec-usage.json
```

Troubleshooting:
- Create if Hook file missing
- Set permissions
- Check data after session end

---

### Q22: Can I override Alfred's decision?

**A**: Yes, always possible!

Method:
```
User choice always possible regardless of Alfred's proposal

Example:
Alfred: "SPEC unnecessary"
User: "I still want to make SPEC"
→ Manually execute /moai:1-plan

Or:

Alfred: "SPEC recommended"
User: "No, implement directly"
→ Reject Alfred's proposal
```

Result: **User's choice is highest priority**

---

## Learn More

### Additional Documentation

| Document | Content |
|----------|---------|
| README.md | Skill overview (5 min read) |
| alfred-decision-logic.md | Decision algorithm (15 min) |
| templates.md | 3-level templates (30 min) |
| analytics.md | Statistics system (20 min) |
| examples.md | 10+ practical examples (30 min) |

### External Resources

- CLAUDE.md: Alfred complete overview
- .moai/specs/: Generated SPEC examples
- .moai/reports/: Monthly reports

---

### Feedback

If you have questions or suggestions:
```bash
/moai:9-feedback "Feedback on SPEC Skill"
```

Or create GitHub Issue:
https://github.com/moai-adk/moai-adk/issues

---

**Document Version**: 1.0.0
**Last Updated**: 2025-11-21
**Status**: Production Ready

---

Happy SPEC-First Development! 🚀
