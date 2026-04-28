---
description: "Five team execution patterns with role profiles and communication protocols"
paths: ".moai/config/sections/workflow.yaml,.claude/rules/moai/workflow/worktree-integration.md"
---

<!-- Source: revfactory/harness — Apache License 2.0 — see .claude/rules/moai/NOTICE.md -->

# Team Pattern Cookbook

Five proven patterns for orchestrating Agent Teams in MoAI-ADK, with role compositions, file ownership, and communication protocols.

---

## Research Team

**Purpose**: Parallel exploration of a topic from multiple angles, with cross-team learning and synthesis.

**Composition**:
- Research Analyst A: Technical deep dive
- Research Analyst B: Community/user perspective
- Analyst C: Market/competitive analysis

**File Ownership**:
- research-deep-dive.md — Analyst A (technical research)
- research-community.md — Analyst B (user/community insights)
- research-market.md — Analyst C (market competitive landscape)
- research-synthesis.md — Team Lead (combines all perspectives)

**Communication Protocol**:
1. Each analyst runs independently on their angle
2. Analysts share findings via SendMessage when insights contradict or amplify
3. Team Lead synthesizes into unified research.md
4. On conflict: present both perspectives with evidence, let user decide

**When Discovery Helps**:
- Large research topics where single perspective misses nuance
- Competing claims in sources (user research vs product goals)
- Need to quickly validate from multiple angles before committing to architecture

**When Solo Agent Better**:
- Research is straightforward (single source, clear answer)
- Team coordination overhead > quality gain

**Example Coordination**:
```
Analyst A: "Current framework adoption is X%"
Analyst B: "But users in forums complain about X's learning curve"
Analyst C: "Competitors using X report Y% migration cost"
Team Lead: "Synthesize these into: X is popular but has adoption risk, competitors mitigating via Z"
```

**Shutdown Sequence**:
1. All analysts TaskUpdate their findings as "complete"
2. Team Lead assembles synthesis.md
3. Team Lead sends synthesis to user
4. Team sends shutdown_request approval to leader
5. Leader runs TeamDelete

---

## Implementation Team

**Purpose**: Parallel development of independent modules with file-level ownership to prevent conflicts.

**Composition**:
- Backend Developer: API endpoints, database, business logic
- Frontend Developer: UI components, client-side logic, hooks
- Test Specialist: Unit + integration tests, coverage

**File Ownership**:
- `src/api/**/*.go` (or equivalent backend) — Backend Dev
- `src/components/**/*.tsx`, `src/hooks/**/*.ts` — Frontend Dev
- `**/*_test.go`, `**/__tests__/**` — Test Specialist (read-only, but owns test files)

**Communication Protocol**:
1. Backend Dev creates API contract document (types, endpoints)
2. Frontend Dev reads contract, creates matching hooks and components
3. Test Specialist writes tests for both layers
4. Boundary issues escalate via SendMessage (e.g., "API response shape doesn't match hook expectation")
5. Backend Dev and Frontend Dev synchronize via SendMessage, then re-implement

**When Team Works**:
- Clear file boundaries (backend/frontend separation)
- API contract can be defined upfront
- Parallel development saves time over sequential

**When Sequential Better**:
- API contract is ambiguous or evolving
- Heavy backend-frontend coupling
- Files are tightly interdependent (can't work in parallel)

**Example Workflow**:
```
Day 1:
  - Backend: Define API schema + endpoints
  - Frontend: Read schema, start building hooks
  - Tests: Write integration tests against API contract

Day 2:
  - Backend: Implement API
  - Frontend: Implement UI with hooks
  - Tests: Run tests, report boundary issues

Day 3:
  - Backend + Frontend: Sync on boundary defects, iterate
  - Tests: Verify fixes
```

**Shutdown Sequence**:
1. Backend Dev marks implementation complete
2. Frontend Dev marks implementation complete
3. Test Specialist verifies all tests green, marks complete
4. Team shuts down after all members idle/complete

---

## Review Team

**Purpose**: Parallel code review from multiple perspectives (security, performance, architecture).

**Composition**:
- Security Reviewer: Auth, input validation, OWASP
- Performance Reviewer: Algorithms, caching, bottlenecks
- Architecture Reviewer: Design patterns, modularity, scalability

**File Ownership**:
- All read-only (mode: plan)
- Each reviewer focuses on their dimension

**Communication Protocol**:
1. All reviewers read full changeset independently
2. Each creates review report in their dimension
3. Reviewers find conflicting opinions via SendMessage (e.g., "Architecture recommends caching, Performance says caching adds complexity here")
4. Discuss and reach consensus on each conflict
5. Single unified review output

**When Parallel Review Helps**:
- Large changes affecting multiple dimensions
- Security and performance often have tradeoffs (need expert negotiation)
- Parallel review faster than sequential

**When Single Reviewer Better**:
- Small, focused change (one dimension)
- Reviewer specialization not clear-cut

**Example Conflict Resolution**:
```
Sec Reviewer: "Add input validation before query"
Perf Reviewer: "Validation adds latency, query sanitization sufficient"
Arch Reviewer: "Input validation should be middleware, separate concern"
Consensus: Input validation in middleware (architecture clean, performance acceptable, security strengthened)
```

**Shutdown Sequence**:
1. All reviewers complete review reports
2. Team Lead synthesizes into single PR feedback
3. Team applies feedback or marks as "reviewed with concerns"
4. TeamDelete

---

## Design Team

**Purpose**: Collaborative design exploration with design, user experience, and stakeholder perspectives.

**Composition**:
- Visual Designer: UI/visual language, component design
- UX Designer: User flow, interaction patterns, accessibility
- Product Lead: Stakeholder requirements, business goals alignment

**File Ownership**:
- `design-visual.md` — Visual Designer (color, typography, component spec)
- `design-ux.md` — UX Designer (flows, interactions, accessibility)
- `design-product.md` — Product Lead (requirements, success metrics)

**Communication Protocol**:
1. All start with the same design brief
2. Visual Designer proposes component system
3. UX Designer reviews flows against components
4. Product Lead validates against business goals
5. Conflicts resolved via SendMessage + negotiation
6. Iterate until all three approve

**When Collaborative Design Helps**:
- Complex user flows requiring visual + interaction + product alignment
- Stakeholder conflict needs real-time discussion
- Design iterations benefit from multiple perspectives

**When Solo Designer Better**:
- Simple UI updates (button style, color tweak)
- Designer has full context and authority

**Example Iteration**:
```
Visual: "Dark mode with minimalist components"
UX: "Dark mode affects contrast for accessibility, reconsider"
Product: "Accessibility is requirement, let's support both modes"
Visual: "Update spec: light + dark modes, WCAG AAA contrast"
UX: "Flows work with both modes, approved"
Product: "Meets requirements, approved"
```

**Shutdown Sequence**:
1. All three approve design
2. Team produces final design.md with visual, UX, and product sections
3. TeamDelete

---

## Debug Team

**Purpose**: Parallel investigation of a complex bug from multiple angles.

**Composition**:
- Reproducer: Isolates and documents the bug with test case
- Analyzer: Traces code paths to identify likely root causes
- Verifier: Tests hypotheses and validates fixes

**File Ownership**:
- `bug-reproduction.md` + `test_repro.go` — Reproducer
- `bug-analysis.md` — Analyzer
- `bug-verification.md` + test runs — Verifier

**Communication Protocol**:
1. Reproducer creates minimal test case that fails
2. Analyzer reads test and code, identifies likely causes
3. Verifier runs hypothesis tests (e.g., "What if we change X?")
4. Analyzer synthesizes hypotheses into ranked list: most likely → least likely
5. Team tests ranked list systematically
6. First successful hypothesis is root cause

**When Parallel Debug Helps**:
- Complex bug with multiple possible causes
- Reproducing bug requires specialist skills
- Multiple hypotheses to test (can test in parallel)

**When Single Debugger Better**:
- Bug is simple (obvious from symptom)
- Reproducing requires deep system context

**Example Investigation**:
```
Reproducer: "GET /api/users returns 500 on parameter X=Y"
Analyzer: "Could be: (1) type mismatch, (2) DB error, (3) missing handler"
Verifier: Tests (1) with different X types → not the issue
Verifier: Tests (2) with DB connection check → found it: DB timeout
Fix: Increase timeout, verify test passes
```

**Shutdown Sequence**:
1. Root cause identified and documented
2. Fix implemented and tested
3. Verifier confirms bug no longer reproduces
4. TeamDelete

---

## Pattern Selection Decision Tree

```
What is the team task?
├── Parallel research/exploration → Research Team
├── Feature implementation (backend + frontend + tests) → Implementation Team
├── Multi-dimension code review (security + perf + arch) → Review Team
├── Collaborative design with stakeholder input → Design Team
├── Complex bug investigation with multiple hypotheses → Debug Team
└── Other → Assess whether clear file boundaries + communication model exists
           If yes, adapt a pattern; if no, use sequential sub-agents
```

---

## Shutdown Protocol (All Teams)

All teams follow this sequence:

1. **Members Complete Work**: Each teammate TaskUpdates their tasks as complete
2. **Idle Status**: Teammates go idle, waiting for coordination
3. **Lead Initiates Shutdown**: Team Lead sends `shutdown_request` with request_id
4. **Members Approve/Reject**: Each teammate responds with `shutdown_response { request_id, approve }`
5. **Final Sync**: Lead synthesizes outputs into final artifact
6. **TeamDelete**: Leader calls TeamDelete to clean up team

If a teammate rejects shutdown (still working), Lead can:
- Extend deadline (send message, wait for completion)
- Collect partial output and proceed (accept incomplete work)
- Escalate to user (team is stuck)

---

## Communication Guidelines

**Via SendMessage**:
- Share findings when they contradict other team members
- Escalate conflicts that need resolution
- Coordinate on interdependent work
- Ask clarifying questions about other member's approach

**Via TaskUpdate**:
- Mark work as complete or blocked
- Provide brief status (no long messages — use file artifacts)

**When to NOT Communicate**:
- Individual analysis that doesn't affect other team members (work independently)
- Status that doesn't block anyone (save for final report)

**Frequency**:
- First day: High communication (establish coordination)
- Middle days: Medium (share discoveries that affect others)
- Final day: High (synthesis and shutdown)

This cookbook provides practical patterns for the most common team scenarios in MoAI-ADK.
