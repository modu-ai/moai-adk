# First Principles Module

Deep dive into root cause analysis and fundamental decomposition techniques.

## The Five Whys Technique

### Application Process

Start with the observed problem and ask "why" repeatedly:

Level 1 - Surface Problem:
- What is the user or system experiencing?
- What are the visible symptoms?
- When and where does it occur?

Level 2 - First Why:
- What is the immediate cause?
- What triggers the symptom?
- What component is directly involved?

Level 3 - Second Why:
- Why does that immediate cause exist?
- What conditions enable it?
- What upstream factor contributes?

Level 4 - Third Why:
- Why do those conditions exist?
- What systemic factor is at play?
- What process or design decision led here?

Level 5 - Fourth/Fifth Why (Root Cause):
- Why was that decision made?
- What fundamental constraint or assumption drove it?
- What would need to change to prevent recurrence?

### Common Pitfalls

Stopping too early:
- Accepting symptom as cause
- Not digging into systemic factors
- Blaming individuals instead of systems

Going too far:
- Reaching philosophical or unchangeable causes
- Losing actionable specificity
- Expanding scope beyond project control

Branching confusion:
- Multiple valid answers at each level
- Explore most impactful branch first
- Document alternative branches for later

## Constraint Analysis

### Hard Constraints

Definition: Non-negotiable requirements that cannot be changed.

Examples:
- Security compliance requirements (SOC2, GDPR, HIPAA)
- Physical limitations (network latency, storage capacity)
- Legal requirements (data retention, accessibility)
- Budget ceiling (approved funding limit)
- Compatibility requirements (existing system integration)

Handling: Design solutions that work within these constraints.

### Soft Constraints

Definition: Preferences that can be adjusted if trade-offs are acceptable.

Examples:
- Timeline preferences (desired but negotiable dates)
- Feature scope (nice-to-have vs must-have)
- Technology preferences (familiar vs optimal)
- Quality level (good enough vs perfect)
- Team preferences (how vs what)

Handling: Use AskUserQuestion to clarify which are truly negotiable.

### Self-Imposed Constraints

Definition: Assumptions disguised as requirements.

Common self-imposed constraints:
- "We have to use technology X" (when alternatives exist)
- "It needs to be done this way" (when other approaches work)
- "We don't have time to..." (when prioritization could help)
- "That's how we've always done it" (legacy process inertia)

Handling: Question whether these are real constraints or habits.

## Decomposition Patterns

### Goal Decomposition

Break high-level goals into actionable sub-goals:

High-Level Goal: Improve application performance

Sub-Goal 1: Reduce page load time
- Metric: Time to first contentful paint
- Target: Under 1.5 seconds

Sub-Goal 2: Improve API response time
- Metric: P95 response time
- Target: Under 200ms

Sub-Goal 3: Reduce resource consumption
- Metric: Memory and CPU usage
- Target: 20% reduction

### Solution Space Mapping

Identify all possible solution directions:

Problem: Database queries are slow

Solution Space:
- Query optimization (indexes, query rewriting)
- Caching layer (Redis, in-memory)
- Database scaling (read replicas, sharding)
- Architecture change (CQRS, event sourcing)
- Data model redesign (denormalization, aggregation)
- Hardware upgrade (faster disks, more memory)

Use AskUserQuestion to explore which directions are viable given constraints.

## Integration with AskUserQuestion

When decomposing problems:
- Use AskUserQuestion to verify understanding of the problem
- Use AskUserQuestion to explore why certain constraints exist
- Use AskUserQuestion to confirm root cause identification
- Use AskUserQuestion to validate decomposition completeness

---

Version: 1.0.0
Parent Skill: moai-foundation-philosopher
