# Philosopher Framework Examples

Practical applications of the Philosopher Framework in real development scenarios.

## Example 1: Technology Selection

### Scenario
Team needs to choose a state management solution for a React application.

### Phase 1: Assumption Audit

Assumptions identified:
- Assumption 1: Application will grow significantly in complexity
  - Confidence: Medium
  - Risk if wrong: Over-engineering with heavyweight solution

- Assumption 2: Team is familiar with Redux patterns
  - Confidence: Low (need to verify)
  - Risk if wrong: Steep learning curve, slower delivery

- Assumption 3: Performance is critical for this application
  - Confidence: High (based on requirements)
  - Risk if wrong: Premature optimization

AskUserQuestion applied:
- Verified team has 2 developers with Redux experience, 3 without
- Confirmed application expected to have 50+ components with shared state
- Clarified performance requirement is for initial load, not state updates

### Phase 2: First Principles Decomposition

Surface Problem: Need state management for React app

First Why: Multiple components need to share and update the same data
Second Why: Component prop drilling becomes unwieldy at scale
Third Why: Application has deeply nested component hierarchy
Root Cause: Need centralized, predictable state access pattern

Constraint Analysis:
- Hard Constraint: Must work with React 18
- Soft Constraint: Team preference for familiar patterns
- Degree of Freedom: Can choose any state management approach

### Phase 3: Alternative Generation

Option A - Redux Toolkit:
- Pros: Industry standard, extensive ecosystem, DevTools
- Cons: Boilerplate, learning curve, potentially overkill

Option B - Zustand:
- Pros: Minimal boilerplate, easy to learn, lightweight
- Cons: Smaller ecosystem, less structured for large apps

Option C - React Context + useReducer:
- Pros: Built-in, no dependencies, familiar React patterns
- Cons: Performance concerns at scale, no DevTools

Option D - Jotai:
- Pros: Atomic approach, minimal re-renders, simple API
- Cons: Different mental model, smaller community

### Phase 4: Trade-off Analysis

Criteria and Weights (confirmed via AskUserQuestion):
- Learning Curve: 25% (team has mixed experience)
- Scalability: 25% (app expected to grow)
- Performance: 20% (important but not critical)
- Ecosystem: 15% (tooling matters)
- Bundle Size: 15% (initial load is priority)

Scores:
- Redux Toolkit: Learning 5, Scale 9, Perf 7, Eco 9, Size 5 = 7.0
- Zustand: Learning 9, Scale 7, Perf 8, Eco 6, Size 9 = 7.8
- Context: Learning 8, Scale 5, Perf 5, Eco 5, Size 10 = 6.5
- Jotai: Learning 7, Scale 8, Perf 9, Eco 5, Size 8 = 7.4

### Phase 5: Cognitive Bias Check

Anchoring: Initial instinct was Redux due to familiarity - verified alternatives genuinely considered
Confirmation: Actively searched for Zustand limitations and Redux advantages
Sunk Cost: No prior investment influencing decision
Overconfidence: Acknowledged uncertainty in scalability predictions

### Recommendation

Selected: Zustand
Rationale: Best balance of learning curve and scalability given team composition
Trade-off Accepted: Smaller ecosystem in exchange for faster team productivity
Mitigation: Plan migration path to Redux if complexity exceeds Zustand capabilities
Review Trigger: If state logic exceeds 20 stores or requires complex middleware

---

## Example 2: Performance Optimization Decision

### Scenario
API endpoint is slow (P95: 2 seconds, target: 200ms).

### Phase 1: Assumption Audit

Assumptions identified:
- Assumption: Database query is the bottleneck
  - Confidence: Medium (based on intuition)
  - Risk if wrong: Optimizing wrong component

AskUserQuestion applied:
- Requested profiling data before proceeding
- Discovered: 60% time in DB, 30% in serialization, 10% in network

### Phase 2: First Principles Decomposition

Surface Problem: API is slow (2 seconds)

First Why: Response takes too long to generate
Second Why: Database query returns too much data
Third Why: Query fetches all columns when only 5 needed
Fourth Why: Using ORM default select without optimization
Root Cause: ORM usage pattern not optimized for this use case

### Phase 3: Alternative Generation

Option A - Optimize existing query:
- Add column selection, indexes
- Effort: Low, Risk: Low

Option B - Add caching layer:
- Redis cache for frequently accessed data
- Effort: Medium, Risk: Low

Option C - Denormalize data:
- Create read-optimized view
- Effort: High, Risk: Medium

Option D - Architecture change:
- Implement CQRS pattern
- Effort: Very High, Risk: High

### Phase 4: Trade-off Analysis

Given that root cause is query optimization, Option A adddesses root cause directly.

Recommendation: Start with Option A, add Option B if insufficient

### Phase 5: Cognitive Bias Check

Avoided availability bias: Did not jump to caching because of recent project
Avoided overconfidence: Requested profiling data before assuming cause

---

## Example 3: Refactoring Scope Decision

### Scenario
Legacy module needs updates. Team debates full rewrite vs incremental improvement.

### Phase 1: Assumption Audit

Assumptions identified:
- Assumption: Full rewrite will take 3 months
  - Confidence: Low (estimates often wrong)
  - Risk if wrong: Project overrun

- Assumption: Legacy code is unmaintainable
  - Confidence: Medium (need metrics)
  - Risk if wrong: Unnecessary rewrite

AskUserQuestion applied:
- Requested code metrics (complexity, test coverage, change frequency)
- Results: 40% test coverage, cyclomatic complexity 25 (high), 3 bugs/month

### Phase 2: First Principles Decomposition

Surface Problem: Legacy module is problematic

First Why: Bugs are frequent and hard to fix
Second Why: Code is complex and poorly tested
Third Why: Original design didn't anticipate current requirements
Root Cause: Design-requirement mismatch accumulated over time

Key insight: Not all parts equally problematic. Core algorithm is solid, interface layer is messy.

### Phase 3: Alternative Generation

Option A - Full rewrite:
- Start fresh, modern patterns
- Risk: Second-system effect, feature parity challenges

Option B - Strangler pattern:
- Gradually replace pieces
- Risk: Prolonged hybrid state, complexity

Option C - Targeted refactoring:
- Fix highest-impact areas only
- Risk: Technical debt remains in untouched areas

Option D - Interface wrapper:
- Clean interface over legacy internals
- Risk: Hiding problems, not solving them

### Phase 4: Trade-off Analysis

Decided via AskUserQuestion: Team has 6 weeks available, not 3 months.

Given time constraint, Option C (targeted refactoring) selected.
Focus areas: Interface layer (source of most bugs), add tests first.

### Recommendation

Approach: Targeted refactoring with test-first approach
Scope: Interface layer only (60% of bugs, 30% of code)
Trade-off: Core algorithm technical debt remains
Mitigation: Comprehensive tests around core algorithm
Review Trigger: If bug rate doesn't decrease 50% in 2 months

---

## Key Learnings

1. Always verify assumptions with data before major decisions
2. Root cause analysis often reveals simpler solutions
3. Time and resource constraints legitimately narrow options
4. Document trade-offs for future reference
5. Set clear review triggers for decisions

---

Version: 1.0.0
Parent Skill: moai-foundation-philosopher
