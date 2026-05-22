# Requirement Clarification Process — Detailed Workflow

5-step systematic process to translate user input into a complete SPEC.

## Step 0 — Assumption Analysis (Philosopher Framework)

Before defining scope, surface and validate underlying assumptions via AskUserQuestion.

### Assumption Categories

- Technical: Technology capabilities, API availability, performance characteristics
- Business: User behavior, market requirements, timeline feasibility
- Team: Skill availability, resource allocation, knowledge gaps
- Integration: Third-party service reliability, compatibility expectations

### Assumption Documentation Template

| Field | Content |
|-------|---------|
| Assumption Statement | Clear description of what is assumed |
| Confidence Level | High / Medium / Low based on evidence |
| Evidence Basis | What supports this assumption |
| Risk if Wrong | Consequence if assumption proves false |
| Validation Method | How to verify before committing significant effort |

## Step 0.5 — Root Cause Analysis

For feature requests or problem-driven SPECs, apply Five Whys:

| Layer | Question |
|-------|----------|
| Surface Problem | What is the user observing or requesting? |
| First Why | What immediate need drives this request? |
| Second Why | What underlying problem creates that need? |
| Third Why | What systemic factor contributes? |
| Root Cause | What fundamental issue must the solution address? |

Root-cause analysis prevents solving symptoms instead of the actual problem.

## Step 1 — Scope Definition

Identify and document:

- Supported methods (e.g., authentication mechanisms, file formats)
- Validation rules and constraints
- Failure handling strategy
- Session management approach (for user-facing features)

## Step 2 — Constraint Extraction

Capture all categories of constraints:

- Performance Requirements: Response time targets (P50, P95, P99 percentiles)
- Security Requirements: OWASP compliance, encryption standards
- Compatibility Requirements: Supported browsers and devices
- Scalability Requirements: Concurrent user targets

## Step 3 — Success Criteria Definition

Define quantifiable, observable completion metrics:

- Test Coverage: Minimum percentage target (e.g., 85%)
- Response Time: Percentile targets (P50, P95, P99)
- Functional Completion: All normal scenarios pass verification
- Quality Gates: Zero linter warnings, zero security vulnerabilities

## Step 4 — Test Scenario Creation

Generate four categories of test cases:

- Normal Cases: Valid inputs with expected outputs
- Error Cases: Invalid inputs with error handling
- Edge Cases: Boundary conditions and corner cases
- Security Cases: Injection attacks, privilege escalation attempts

Every EARS requirement must map to at least one test scenario.
