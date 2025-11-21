# SPEC Exception Handling Handbook

**Status**: Static Reference | **Version**: 1.0.0 | **Language**: English

---

## Overview

While the SPEC-First workflow works for most situations, certain **exceptional scenarios** warrant different handling.

This handbook documents:
- When to bypass normal SPEC workflow
- How to handle special cases
- User choice and autonomy principles
- Mid-implementation complexity escalation

---

## Exception 1: Prototype Development

### Definition
Rapid iteration focused on exploration, validation, or proof-of-concept rather than production code.

### Detection Keywords
- "**quickly prototype**" - Speed emphasis
- "**fast iteration**" - Rapid feedback loops
- "**explore**" - Experimental investigation
- "**proof-of-concept**" - Validation before building
- "**rapid mockup**" - Quick visual exploration

### What Alfred Does

```
User: "I want to quickly prototype a new UI design"
  ↓
Alfred: Detects "quickly prototype" keyword
  ↓
Action: SKIP SPEC recommendation
  ↓
Guidance: "Let's rapidly iterate on the design"
  ↓
Suggestion: "Once validated, we can formalize as SPEC for production"
```

### Why SPEC Not Recommended

| Aspect | Prototyping | Production |
|--------|-----------|-----------|
| Documentation | Optional | Critical |
| Testing | Minimal | Comprehensive |
| Review Process | Informal | Formal |
| Architecture | Loose | Strict |
| Timeline | Hours-Days | Days-Weeks |

### Process

1. **Implement Rapidly** (No SPEC)
   - Focus on speed and exploration
   - Accept technical shortcuts
   - Iterate based on feedback

2. **Validate** with Stakeholders
   - Gather feedback
   - Identify what worked
   - Understand what's needed

3. **Transition to Production** (Create SPEC)
   - Formalize requirements from prototype
   - Design proper architecture
   - Plan systematic implementation
   - Set test coverage expectations

### Example

```
Prototype Phase:
─────────────────
"Make a dashboard mockup in 2 hours to show stakeholders"
→ No SPEC created
→ Quick React components without tests
→ Iterative design feedback

Validation:
───────────
Stakeholders approve the design concept

Production Transition:
──────────────────────
"Now let's build the real dashboard"
→ /moai:1-plan creates SPEC-XXX
→ Create database models
→ Implement with tests
→ Deploy with confidence
```

---

## Exception 2: Emergency Bug Fix

### Definition
Critical production issues requiring immediate resolution without normal development process.

### Detection Keywords
- "**broken**" - Service down
- "**urgent**" - Time-sensitive
- "**critical**" - Business impact
- "**emergency**" - Immediate action
- "**crash**" - System failure

### What Alfred Does

```
User: "Payment processing is broken! Users can't complete purchases!"
  ↓
Alfred: Detects "broken", "urgent" keywords
  ↓
Action: SKIP SPEC recommendation
  ↓
Guidance: "Immediate implementation for emergency fix"
  ↓
Process: Fast diagnosis → Quick fix → Deploy
```

### Why SPEC Not Appropriate

| Phase | Normal Development | Emergency |
|-------|-------------------|-----------|
| Planning | 10-15 min | Skip |
| Documentation | Required | Minimal |
| Testing | Comprehensive | Smoke test |
| Review | Required | Expedited |
| Deployment | Scheduled | Immediate |

### Process

1. **Diagnosis** (< 5 min)
   - Identify root cause
   - Determine scope
   - Assess impact

2. **Quick Fix** (< 30 min)
   - Implement minimal fix
   - Avoid major refactoring
   - Accept temporary solutions

3. **Deploy** (< 5 min)
   - Fast deployment
   - Monitor closely
   - Have rollback ready

4. **Post-Fix Documentation** (After stabilization)
   - Document what happened
   - Plan proper fix (SPEC-based)
   - Prevent recurrence

### Example

```
Emergency Scenario:
───────────────────
"Database connection pool exhausted - API down"

Immediate Fix (5 min):
- Increase connection limit
- Restart affected services
- Verify traffic restored

Post-Emergency (Next day):
- Create SPEC for proper connection pooling
- Design monitoring and alerts
- Implement systematic solution
```

### Recovery Process

```
Emergency Fix → Stabilization → Analysis → Formal SPEC → Production Release
   (2 hours)      (1 hour)    (2 hours)    (15 min)      (1 week)
```

---

## Exception 3: Team Collaboration

### Definition
Distributed team work requiring explicit coordination on larger features.

### When Collaboration Emphasis Matters

- **Remote Teams**: Multiple time zones
- **Large Teams**: 3+ people working together
- **Complex Features**: Multiple integration points
- **Long Duration**: Multi-week development

### What Alfred Does

```
User: "Our team is building a payment system (3 backend, 2 frontend)"
  ↓
Alfred: Analyzes complexity + team size
  ↓
Action: STRONGLY RECOMMEND SPEC
  ↓
Guidance: "SPEC essential for team coordination"
```

### Why SPEC Critical for Teams

| Aspect | Individual | Team |
|--------|-----------|------|
| Communication | Implicit | Explicit (SPEC) |
| Integration Risk | Low | High |
| Testing Overlap | Possible | Likely |
| Schedule Sync | Flexible | Coordinated |
| Documentation | Personal | Shared |

### Team Workflow

```
1. SPEC Creation
   - Define architecture together
   - Assign responsibilities
   - Set integration points

2. Parallel Development
   - Backend team works on API
   - Frontend team works on UI
   - Both reference same SPEC

3. Integration Testing
   - Verify API contracts
   - Test data flows
   - Validate workflows

4. Deployment
   - Coordinated release
   - Zero downtime strategies
   - Rollback procedures
```

### Example

```
Payment System (Team: 5 developers)
──────────────────────────────────

SPEC-005: Payment Processing Integration

Team Assignments:
- Backend (A, B): Payment API, transaction management
- Frontend (C, D): UI components, form handling
- DevOps (E): Deployment pipeline, monitoring

Integration Points:
- API contract (documented in SPEC)
- Database schema (shared)
- Error handling (standardized)

Timeline:
- Week 1: Parallel development (reference SPEC)
- Week 2: Integration and testing
- Week 3: Deployment and monitoring
```

### Why Individual Exceptions Don't Apply

**Prototype in Team**:
```
✗ WRONG: "Let's prototype the payment flow quickly"
         → Leads to rework, wasted effort

✓ RIGHT: Create SPEC first, then code in parallel
         → Each team member understands the whole picture
```

**Emergency in Team**:
```
✗ WRONG: "Quick fix without coordination"
         → Other team member's work breaks

✓ RIGHT: "Emergency hotfix with SPEC context"
         → Everyone knows what changed
         → Coordinated rollout/rollback
```

---

## Exception 4: Complex Implementation Discovery

### Definition
Work starts simple (no SPEC needed), but complexity increases during implementation.

### Detection

**Initial Assessment**: "Change button color" (0 criteria met)
→ No SPEC created
→ Work begins

**During Implementation**: Discover button color affects:
- Design system tokens
- Multiple components
- Accessibility contrast
- Dark mode variant
→ Complexity increases to 4-5 criteria

### What Alfred Does

```
During Implementation:
  Developer: "This is more complex than expected"

  Alfred: Analyzes current status
  ├─ Files changed: 5 (vs expected 1)
  ├─ Time spent: 40 min (vs estimated 5 min)
  └─ Architecture impact: Design system changes

  Analysis: New criteria met = 4

  Action: OFFER SPEC CONVERSION

  Proposal: "This has become more complex.
             Should we create a SPEC to plan properly?"

  Options:
  ├─ Yes: Create SPEC-XXX, plan implementation strategy
  └─ No: Continue ad-hoc, accept higher risk
```

### Process

1. **Continue Current Work** (if simple enough)
   - Finish what you started
   - Document what changed
   - Note complexity increase

2. **Evaluate Complexity** (if work ongoing)
   - Assess new scope
   - Estimate remaining work
   - Consider coordination needs

3. **Offer SPEC Conversion**
   - Present analysis
   - Show what would change
   - Let user decide

4. **If User Chooses SPEC**
   - Create SPEC based on current progress
   - Incorporate discovered requirements
   - Plan remaining work systematically
   - Continue from where you left off

### Example

```
Initial Request:
────────────────
"Add a 'remember me' checkbox to login form"

Alfred Analysis:
- 1 file (form component)
- No architecture change
- Simple feature
→ No SPEC needed

Implementation:
────────────────
Developer finds:
- Need session management changes
- Database schema for persistent sessions
- Security considerations (token rotation)
- Multiple affected components
→ Complexity increased!

Mid-Implementation Discovery:
────────────────────────────
Alfred: "This has grown to 4 criteria.
         Want to create a SPEC to organize remaining work?"

User Options:
├─ Yes: Create SPEC for systematic implementation
└─ No: Continue current approach with known risks

If Yes:
─ SPEC-XXX: Session Persistence with Remember-Me
 - Incorporates current progress
 - Plans architecture changes
 - Designs test strategy
```

---

## Exception 5: User Override

### Definition
User decides **not to follow** Alfred's SPEC recommendation.

### Why This Matters

```
User Autonomy Principle:
└─ Alfred proposes, user decides
   └─ No penalties for declining
   └─ User's judgment is respected
   └─ Frequent overrides are tracked (for improvement)
```

### Scenarios

**Scenario A: User Declines Recommended SPEC**
```
Alfred: "This warrants a SPEC (4 criteria met)"
User: "No thanks, let's just build it"

Result:
✅ No SPEC created
✅ Development proceeds
⚠️  Tracked for monthly review
   → If issues arise, informs future recommendations
```

**Scenario B: User Requests SPEC Despite Recommendation Against**
```
Alfred: "SPEC not needed (0-1 criteria met)"
User: "/moai:1-plan 'I want a SPEC anyway'"

Result:
✅ SPEC-XXX created (user-requested)
✅ Systematic planning proceeds
✅ User's judgment respected
```

**Scenario C: User Changes Mind Mid-Implementation**
```
Initially: No SPEC (seemed simple)
Mid-way: "Actually, let's create a SPEC"

Result:
✅ SPEC created for remaining work
✅ Current progress incorporated
✅ Systematic completion from this point
```

### Tracking User Overrides

Monthly report includes override patterns:

```markdown
## User Decision Overrides

### Declined Recommendations (vs Alfred)
- Dec 1: Recommended SPEC, user declined → Issue found later
- Dec 5: Recommended SPEC, user declined → Completed fine
- Dec 8: Recommended SPEC, user declined → Blocker encountered

### Requested SPECs Against Recommendation
- Dec 3: Alfred said "skip", user requested SPEC → Glad they did

Pattern Recognition:
- When users override (accept risks)
- When overrides lead to issues
- Recommendation accuracy trending
```

---

## Exception 6: Experimental Research

### Definition
Spike or research work where the goal is **learning**, not production code.

### Detection Keywords
- "**research**" - Investigation
- "**spike**" - Time-boxed exploration
- "**investigate**" - Problem analysis
- "**evaluate**" - Technology assessment
- "**POC**" - Proof of concept

### What Alfred Does

```
User: "Let's research authentication options (OAuth, JWT, Sessions)"
  ↓
Alfred: Detects "research", "evaluate" keywords
  ↓
Action: SKIP SPEC recommendation
  ↓
Guidance: "Research phase - plan systematically once decided"
```

### Why SPEC Not Appropriate for Research

| Aspect | Research | Implementation |
|--------|----------|-----------------|
| Goal | Learn options | Build feature |
| Outcome | Decision | Code |
| Quality Bar | Exploratory | Production |
| Documentation | Notes | SPEC |

### Process

1. **Research Phase** (No SPEC)
   - Create prototype/spike
   - Evaluate options
   - Document findings
   - Make recommendation

2. **Decision**
   - Choose approach
   - Document rationale
   - Get buy-in

3. **Implementation** (Create SPEC)
   - Build production version
   - Follow SPEC-First workflow
   - Implement systematically

### Example

```
Research Phase:
───────────────
"Compare 3 auth approaches:
1. JWT tokens
2. OAuth2 with social login
3. Session-based auth"

Spike work (1-2 days):
- Prototype each approach
- Test with sample app
- Document pros/cons
- No test coverage needed
- No SPEC created

Decision:
─────────
"OAuth2 is best for our use case"
→ Document decision rationale

Implementation:
───────────────
/moai:1-plan "OAuth2 Integration"
→ Create SPEC-XXX
→ Systematic implementation
→ Full test coverage
→ Production quality
```

---

## Decision Flow Diagram

```
User Request
    ↓
━━━━━━━━━━━━━━━━━━━━━━━━━
Prototype?      ←─→ Skip SPEC, iterate fast
"quickly", "explore"     │
              ├─ YES ─→ PROTOTYPE PHASE
              │         (later: transition to SPEC)
              │
Emergency?     ↓
"broken", "urgent"  ←─→ Skip SPEC, fix fast
              │         (later: formalize with SPEC)
              ├─ YES ─→ EMERGENCY PHASE
              │         (post-incident: SPEC for prevention)
              │
Research?      ↓
"investigate", "spike" ←─→ Skip SPEC, explore
              │         (later: implement with SPEC)
              ├─ YES ─→ RESEARCH PHASE
              │         (post-decision: implement with SPEC)
              │
Team Size > 3?  ↓
Multi-team work ←─→ STRONGLY RECOMMEND SPEC
              │     (essential for coordination)
              ├─ YES ─→ REQUIRE SPEC
              │
              ↓
Analyze Complexity
(5 criteria assessment)
              │
        ┌─────┼─────┐
        ↓     ↓     ↓
    0-1  2-3  4-5
     │    │    │
 SKIP  REC  STR_REC
 (1)   (2)   (3)
     │    │    │
     ├────┴────┤
     ↓         ↓
   User     User
  Choice   Choice
    │         │
  YES │ NO   YES │ NO
    ├──→ PROCEED ←──┤
    │               │
    └─ No SPEC ─────┘

    Or

    └─ Create SPEC ─┘


Legend:
(1) Skip SPEC, immediate implementation
(2) SPEC recommended, user decides
(3) SPEC strongly recommended, user decides
```

---

## User Autonomy Principles

### Core Rule
> **The user always decides.** Alfred proposes; user chooses.

### Implementation
1. **Present Options**: Show what SPEC recommendation is
2. **Explain Reasoning**: Why SPEC recommended or not
3. **Respect Choice**: If user declines, proceed without penalty
4. **Track Patterns**: Use overrides to improve future recommendations

### No Enforcement
- ❌ Never force SPEC creation
- ❌ Never block work without SPEC
- ❌ Never penalize SPEC rejection
- ❌ Never override user decision

### Continuous Learning
- 📊 Track recommendation accuracy
- 📈 Analyze override patterns
- 🔄 Adjust criteria based on data
- ✅ Improve over time

---

## Common Exception Patterns

### Pattern 1: "Quick Fix" Becomes Complex
```
Initial: "Just fix the typo"
  → No SPEC created

Mid-way: "Wait, this affects multiple pages"
  → Complexity increased
  → Offer SPEC conversion
  → User can accept/decline

Result: Intentional vs accidental complexity
        → Informs future recommendations
```

### Pattern 2: Team Expansion
```
Initial: Solo developer, 1 criterion → No SPEC
  ↓
Change: Colleague joins to help
  ↓
New situation: Team of 2, 3+ criteria → SPEC recommended
  ↓
Action: "Since we're now a team, SPEC would help coordination"
```

### Pattern 3: Prototype Success
```
Initial: Rapid UI mockup created
  ↓
Discovery: Stakeholders love the design
  ↓
Decision: "Let's production-ize this"
  ↓
Action: Create SPEC for systematic implementation
```

---

**Document Version**: 1.0.0
**Last Updated**: 2025-11-21
**Status**: Production Ready
