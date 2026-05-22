# EARS Format Deep Dive

Five patterns cover all requirement types. Each pattern has a specific use case, examples, and test strategy.

## Ubiquitous Requirements — Always Active

Format: "The system shall always X"

Use case: System-wide quality attributes that apply unconditionally.

Examples:

- Logging
- Input validation
- Error handling

Test strategy: Include in all feature test suites as common verification. These checks run regardless of the feature being tested.

## Event-Driven Requirements — Trigger-Response

Format: "WHEN event occurs THEN action executes"

Use case: User interactions and inter-system communication.

Examples:

- Button clicks → form submission
- File uploads → validation pipeline
- Payment completions → confirmation email

Test strategy: Event simulation with expected response verification. Trigger the event, observe the response, assert correctness.

## State-Driven Requirements — Conditional Behavior

Format: "IF condition is true THEN action executes"

Use case: Access control, state machines, conditional business logic.

Examples:

- Account status checks (IF active THEN allow login)
- Inventory verification (IF stock > 0 THEN allow purchase)
- Permission checks (IF admin role THEN show admin panel)

Test strategy: State setup with conditional behavior verification. Set up each state, verify the conditional branch.

## Unwanted Requirements — Prohibited Actions

Format: "The system shall not X"

Use case: Security vulnerabilities, data integrity protection.

Examples:

- No plaintext passwords stored
- No unauthorized access to admin endpoints
- No PII in logs

Test strategy: Negative test cases with prohibited behavior verification. Attempt the prohibited action, assert rejection or absence.

## Optional Requirements — Enhancement Features

Format: "Where possible, provide X"

Use case: MVP scope definition, feature prioritization.

Examples:

- OAuth login (in addition to email/password)
- Dark mode
- Offline mode

Test strategy: Conditional test execution based on implementation status. Tests skip when feature is not implemented, run when it is.

## Pattern Selection Guide

| Question | Pattern |
|----------|---------|
| Does it apply always, regardless of input? | Ubiquitous |
| Is there an external trigger? | Event-Driven |
| Does it depend on system state? | State-Driven |
| Is it forbidden behavior? | Unwanted |
| Is it nice-to-have but not required? | Optional |

Mix patterns within a single SPEC as needed. A complete feature usually combines 3-5 patterns.
