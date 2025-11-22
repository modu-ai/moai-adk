# Alfred's SPEC Decision Logic

**Created**: 2025-11-21
**Status**: Production Ready

---

## Overview

Alfred analyzes user requests and conversations in **natural language** to automatically determine the necessity of SPEC creation.

This document explains Alfred's decision criteria and how it proposes recommendations to users in detail.

---

## Decision Criteria: 5 Questions

Alfred analyzes task complexity through the following 5 questions:

### ① File Modification Scope
**Q**: Does it modify or create multiple files?

```
No: Only one file modified
  Examples: CSS style change, string modification, adding a single function

Possible: 2-3 files modified
  Examples: Login logic modification (component + service)

Yes: 4 or more files modified
  Examples: Image upload (API + Frontend + DB + Middleware)
```

### ② Architecture Impact
**Q**: Are there architecture or data model changes?

```
No: Existing structure maintained
  Examples: Logic modification in existing endpoints

Possible: Partial changes
  Examples: Adding new Service class, adding existing DB columns

Yes: Major changes
  Examples: Microservice transition, introducing new architecture patterns
```

### ③ Component Integration
**Q**: Is integration across multiple components required?

```
No: Single component only
  Examples: Changes only within one page

Possible: 2-3 components
  Examples: Login component + Profile component

Yes: 4 or more components
  Examples: Frontend + Backend + Database + Cache + Message Queue
```

### ④ Implementation Time
**Q**: Is implementation time expected to be 30 minutes or more?

```
No: 15 minutes or less
  Examples: Color change, text modification, simple function

Possible: 15-30 minutes
  Examples: Simple feature addition, partial refactoring

Yes: 30 minutes or more
  Examples: Complex features, architecture changes, integration work
```

### ⑤ Future Maintenance
**Q**: Is future maintenance or expansion needed?

```
No: One-time task
  Examples: Urgent bug fix, temporary logic

Possible: Future change possibility
  Examples: New payment module, authentication system

Yes: Clear maintenance/expansion needed
  Examples: Core features, reusable components
```

---

## Decision Logic

Alfred collects answers to the above 5 questions and makes decisions as follows:

```
Number of "Yes" or "Possible" answers:

┌─────────┬──────────────┬────────────────────────┐
│ Count   │ Decision     │ Action                 │
├─────────┼──────────────┼────────────────────────┤
│ 0-1     │ SPEC         │ Proceed with           │
│         │ Unnecessary  │ implementation         │
│ 2-3     │ SPEC         │ User choice (Yes/No)   │
│         │ Recommended  │                        │
│ 4-5     │ SPEC         │ Emphasized proposal    │
│         │ Strongly     │                        │
│         │ Recommended  │                        │
└─────────┴──────────────┴────────────────────────┘
```

---

## User Proposal Patterns

### Pattern A: SPEC Unnecessary (0-1 met)

```
GOOS, I'll proceed with implementation without SPEC for this task.

Analysis:
  • Only one file modified
  • No architecture impact
  • Implementation time under 15 minutes

→ Proceed with immediate implementation
```

### Pattern B: SPEC Recommended (2-3 met)

```
GOOS, I recommend creating a SPEC document for this task for the following reasons:

📋 Analysis Results:
  ✓ Multiple files need modification (Backend, Frontend)
  ✓ Data model changes present
  - Component integration: 2
  - Implementation time: 45 minutes
  - Future maintenance: Required

This can reduce implementation time by 30%.

Please select from the following:
```

Then call `AskUserQuestion`:

```json
{
  "questions": [
    {
      "question": "Would you like to generate a SPEC document and proceed?",
      "header": "SPEC Proposal",
      "multiSelect": false,
      "options": [
        {
          "label": "Yes, generate SPEC then implement",
          "description": "Automatically executes /moai:1-plan and delegates to spec-builder"
        },
        {
          "label": "No, start implementation now",
          "description": "Proceeds with implementation without SPEC"
        }
      ]
    }
  ]
}
```

### Pattern C: SPEC Strongly Recommended (4-5 met)

```
GOOS, a SPEC document is **strongly recommended** for this task for the following reasons:

⚠️ Complexity Analysis:
  ✓ Multiple file modifications (Backend, Frontend, DB included)
  ✓ Architecture changes required
  ✓ 3+ component integration
  ✓ Estimated implementation time: 2-3 hours
  ✓ Future maintenance required

Expected benefits with SPEC:
  • 40% implementation time reduction
  • 60% bug risk reduction
  • 50% future maintenance cost savings

Please select from the following:
```

Call `AskUserQuestion` in the same manner

---

## Automatic Workflow

### When User Selects "Yes, Generate SPEC"

**Step 1: User Feedback**
```
GOOS, I'll generate the SPEC. Please wait a moment...
```

**Step 2: Execute SPEC Generation**
```bash
/moai:1-plan "{requirements summary}"
```

Alfred automatically calls:
```python
Task(
  subagent_type="spec-builder",
  description="Generate SPEC document",
  prompt="Analyze user requirements and generate SPEC"
)
```

**Step 3: Automatic Template Selection**
spec-builder automatically selects from:
```
Complexity LOW → Level 1 (Minimal, 5 sections)
Complexity MEDIUM → Level 2 (Standard, 7 sections, EARS)
Complexity HIGH → Level 3 (Comprehensive, 10+ sections)
```

**Step 4: SPEC Generation Complete**
```
SPEC-001 has been generated.
```

**Step 5: Context Initialization**
```
Execute /clear to start new session
```

**Step 6: Implementation Proposal**
```
📋 Next Steps:
  1. /clear (initialize context)
  2. /moai:2-run SPEC-001 (TDD implementation)

Would you like to proceed?
```

---

## Decision Examples

### Example 1: Bug Fix (SPEC Unnecessary)

**User Request**:
> "The error message on the login page is only displayed in English. Please change it to Korean."

**Alfred Analysis**:
```
① File modification: Only i18n configuration → No
② Architecture: No changes → No
③ Components: Single → No
④ Time: 10 minutes → No
⑤ Maintenance: None → No

Result: 0 met → SPEC Unnecessary
```

**Alfred Decision**:
```
Proceed with immediate implementation

Implementation: Add Korean translation to i18n/messages.json
```

---

### Example 2: Feature Addition (SPEC Recommended)

**User Request**:
> "I want to add a feature where users can upload profile pictures. It should include image optimization, caching, and DB storage."

**Alfred Analysis**:
```
① File modification: Backend API + Frontend Form + DB Schema + Middleware
   → 4 or more → Yes

② Architecture: Add file upload flow, image processing layer
   → Major changes → Yes

③ Components: Frontend + Backend + Database + Cache
   → 4 or more → Yes

④ Time: Image optimization, S3 integration, frontend UI
   → 2+ hours → Yes

⑤ Maintenance: Future profile-related features possible
   → Certain → Yes

Result: 5 met → SPEC Strongly Recommended
```

**Alfred Decision**:
```
Propose SPEC strong recommendation to user

User selects "Yes"

→ Automatically execute /moai:1-plan
→ spec-builder selects Level 2 (Standard) template
→ Auto-generate SPEC-003
→ TDD implementation with /moai:2-run SPEC-003
```

---

### Example 3: Prototype (Exception)

**User Request**:
> "I want to quickly make a prototype. Fast implementation is more important than accurate design."

**Alfred Analysis**:
```
Prototype keyword detected
→ Recognize words like "quickly", "prototype", "temporary"

Special handling:
  Skip SPEC regardless of complexity
  Recommend immediate implementation
```

**Alfred Decision**:
```
GOOS, it's a prototype. I'll proceed quickly.

→ Start immediate implementation
→ Can propose SPEC when transitioning to production
```

---

## Exception Handling

### Case 1: User Rejects SPEC

```
GOOS, I'll proceed with immediate implementation without SPEC.

⚠️ Note: If complexity increases during implementation,
we can transition to SPEC.
```

→ Proceed with immediate implementation

### Case 2: Complexity Increases During Implementation

Initial decision: Simple modification (SPEC unnecessary)
During implementation: Complexity unexpectedly increases

```
GOOS, this task is becoming more complex than expected.

Current situation:
  • Additional file modifications needed
  • Possible architecture changes
  • Expected time increased to 2 hours

Would you like to create a SPEC for systematic progress?
```

→ Can immediately execute `/moai:1-plan` based on user choice

### Case 3: Emergency Situation

```
User: "This is a production bug. Immediate fix needed!"

Alfred: Skip SPEC proposal

→ Immediate implementation
→ SPEC creation possible after completion
```

---

## Alfred's Advantages

### 1. Natural Workflow
```
❌ Before: User decides "Is SPEC needed?" every time
✅ After: Alfred automatically decides and proposes
```

### 2. Minimize False Positives
```
Conservative decision with 5 conditions
→ Reduce unnecessary proposals
→ High user trust
```

### 3. Flexible Response
```
Prototypes, emergency fixes, changes during implementation
Can respond to various situations
```

### 4. Data-Driven Improvement
```
Measure effectiveness with monthly statistics
→ Continuously improve decision criteria
```

---

## Implementation Checklist

When implementing Alfred's SPEC decision:

- [ ] Implement 5 question prompts
- [ ] Condition fulfillment count calculation logic
- [ ] AskUserQuestion integration
- [ ] Automatic /moai:1-plan trigger
- [ ] Template automatic selection logic
- [ ] Exception handling (prototype, emergency)
- [ ] Detect complexity increase during implementation
- [ ] Statistics data collection integration

---

**Document Version**: 1.0.0
**Last Updated**: 2025-11-21
**Status**: Production Ready
