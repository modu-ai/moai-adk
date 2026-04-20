# Deep Questioning Module

Detailed guide for the 6-layer progressive inquiry framework.

## Layer-by-Layer Guide

### Layer 1: Surface Understanding

Purpose: Establish baseline comprehension of the stated problem.

Key Questions:
- What exactly is being requested?
- What is the expected output or deliverable?
- What are the obvious inputs and data sources?
- Who made this request and why?
- What is the timeline expectation?

Verification Test: Can you explain this request to a non-technical person in one paragraph? If not, you need more information.

Common Pitfalls at This Layer:
- Assuming you understand when you only recognize keywords
- Conflating the request with a similar past request
- Skipping clarification because the request "seems obvious"

Output: Clear, restated problem definition.

### Layer 2: Problem Depth

Purpose: Understand why the problem exists and what has been tried before.

Key Questions:
- Why does this problem exist now? (What changed?)
- What is the root cause vs the symptom?
- How long has this been a problem?
- What workarounds currently exist?
- What has been tried before and why did it fail?
- What would happen if we did nothing?

Root Cause Techniques:
- Ask "Why?" five times to drill past symptoms
- Distinguish between proximate cause and root cause
- Look for systemic patterns (is this a recurring issue?)

Historical Context:
- Previous attempts and their outcomes
- Lessons learned from related problems
- Organizational or technical debt contributing to the issue

Output: Root cause identification and problem history.

### Layer 3: Context and Constraints

Purpose: Map the full constraint landscape surrounding the problem.

Technical Constraints:
- What technology stack is in use?
- What are the performance requirements?
- What are the scalability expectations?
- What backward compatibility is required?
- What integration points exist?

Organizational Constraints:
- What team skills are available?
- What approval processes are required?
- What compliance or regulatory requirements apply?
- What political or cultural factors influence the solution?

Resource Constraints:
- What is the budget (time, money, people)?
- What infrastructure is available?
- What third-party services can be used?

Hidden Constraints Discovery:
- Ask: "What would make this solution unacceptable?"
- Ask: "What assumptions are embedded in the current architecture?"
- Ask: "What would a new team member find surprising about this system?"

Output: Comprehensive constraint map.

### Layer 4: User Perspective

Purpose: Understand the actual end-user experience and needs.

User Research Questions:
- Who are the actual end users? (Not who we assume)
- What is their current workflow step by step?
- What are their biggest pain points?
- What do they value most? (Speed? Accuracy? Simplicity?)
- What would they consider a disappointing solution?
- What would delight them beyond expectations?

Empathy Mapping:
- What do users think about when using the current system?
- What do users feel frustrated by?
- What do users say when describing the problem?
- What do users actually do (observed behavior vs stated behavior)?

Edge User Consideration:
- Power users: What advanced features matter?
- New users: What onboarding experience is needed?
- Occasional users: What must be immediately intuitive?
- Accessibility users: What accommodation is required?

Output: User needs analysis and empathy map.

### Layer 5: Solution Exploration

Purpose: Define the solution space boundaries and technical requirements.

Boundary Conditions:
- What is the minimum viable solution?
- What is the maximum scope we should consider?
- What edge cases could break the solution?
- What error conditions must be handled?

Performance Requirements:
- Expected load (requests per second, concurrent users)
- Latency requirements (p50, p95, p99)
- Data volume expectations (current and projected)
- Availability requirements (uptime SLA)

Integration Requirements:
- What existing systems must the solution work with?
- What APIs must be consumed or provided?
- What data formats are required?
- What authentication and authorization is needed?

Quality Attributes:
- Maintainability: How easy should it be to modify?
- Testability: How should it be tested?
- Observability: What monitoring is needed?
- Security: What threat model applies?

Output: Detailed solution requirements specification.

### Layer 6: Validation and Risk

Purpose: Define how we will know if the solution works and what could go wrong.

Validation Strategy:
- How will we measure success?
- What metrics will we track?
- What is the acceptance testing plan?
- What user feedback mechanism will we use?
- When will we know if the solution failed?

Risk Assessment:

Risk Categories:
- Technical Risk: Implementation may not work as expected
- Integration Risk: May not work with existing systems
- Performance Risk: May not meet load requirements
- Security Risk: May introduce vulnerabilities
- Schedule Risk: May take longer than estimated
- Adoption Risk: Users may not adopt the solution

Risk Evaluation Matrix:
- Probability: Low, Medium, High
- Impact: Low, Medium, High, Critical
- Priority: Probability multiplied by Impact

Mitigation Planning:
- For each High/Critical risk, define a mitigation strategy
- Identify monitoring and early warning signals
- Define rollback procedure
- Establish "kill criteria" (when to abandon the approach)

Rollback Strategy:
- Can we revert to the previous state?
- How long would rollback take?
- What data would be lost during rollback?
- Who needs to be involved in a rollback decision?

Output: Risk register and validation plan.

---

## Depth Indicators

Use these guidelines to determine appropriate questioning depth:

Shallow (Layers 1-2):
- Quick bug fixes
- Minor UI adjustments
- Configuration changes
- Well-understood tasks

Moderate (Layers 1-4):
- New feature implementation
- API design decisions
- Database schema changes
- Performance optimization

Deep (All 6 Layers):
- Architecture decisions
- Technology selection
- System redesign
- Security-critical features

Exhaustive (All Layers, Multiple Iterations):
- Platform migration
- Core infrastructure changes
- Compliance-critical systems
- Multi-team initiatives

---

## Deep Questioning Output Template

Understanding Level: [Shallow / Moderate / Deep / Exhaustive]

Layer 1 - Surface:
- Stated Goal: [What is requested]
- Success Definition: [What success looks like]

Layer 2 - Depth:
- Root Cause: [Why this problem exists]
- Prior Attempts: [What has been tried]

Layer 3 - Context:
- Key Constraints: [Critical limitations]
- Hidden Constraints: [Discovered limitations]

Layer 4 - Users:
- Primary Users: [Who they are]
- Key Pain Points: [What matters most]

Layer 5 - Solution:
- Boundary Conditions: [Min/max scope]
- Edge Cases: [Potential failure points]

Layer 6 - Validation:
- Success Metrics: [How we measure]
- Top Risks: [Highest priority risks]

Open Questions: [Remaining unknowns]
Recommended Next Steps: [What to do next]
