# TRUST 5 Framework — Detailed Assessment Rubric

Conceptual quality assessment model with five dimensions. Provides guidance for evaluating code quality; not an implemented module.

## Dimension 1 — Testability

Evaluate whether the code can be effectively tested:

- Are functions pure and deterministic?
- Are dependencies injectable (no hidden globals)?
- Is the code modular enough for unit testing?
- Are side effects isolated and observable?

Scoring spans from low testability requiring significant refactoring to high testability with excellent test coverage support.

## Dimension 2 — Readability

Assess how easily the code can be understood:

- Are variable and function names descriptive?
- Is the code structure logical?
- Are complex operations documented?
- Does the naming follow project convention?

Scoring evaluates naming conventions, code organization, and documentation quality.

## Dimension 3 — Understandability

Evaluate conceptual clarity of the implementation:

- Is the business logic clearly expressed?
- Are abstractions appropriate (not over- or under-engineered)?
- Can a new developer understand the code quickly?
- Does the code communicate intent beyond mechanical operation?

Goes beyond surface readability to assess architectural clarity.

## Dimension 4 — Security

Assess security posture and vulnerability exposure:

- Are inputs validated at trust boundaries?
- Are secrets properly managed (no hardcoded credentials)?
- Are common vulnerability patterns avoided (injection, XSS, CSRF, SSRF)?
- Does the code follow OWASP best practices?

Scoring evaluates adherence to security best practices.

## Dimension 5 — Transparency

Evaluate operational visibility and debuggability:

- Is error handling comprehensive (no silent failures)?
- Are logs meaningful and structured?
- Can issues be traced through the system end-to-end?
- Are metrics and observability hooks present?

Scoring assesses observability and troubleshooting capabilities.

## Overall TRUST Score Calculation

The overall TRUST score combines all five dimensions using weighted averaging. Critical issues in any dimension can override the average, ensuring security or testability problems are not masked by high scores elsewhere.

A passing score typically requires:

- Minimum threshold met in each individual dimension (no dimension below floor)
- Acceptable weighted average across all five dimensions
- Zero critical issues in security and testability

## Scoring Reference Values

| Mode | Average Threshold | Per-Dimension Floor | Critical Issues Allowed |
|------|-------------------|---------------------|-------------------------|
| Strict | 0.90 | 0.80 | 0 |
| Standard | 0.80 | 0.70 | 0 |
| Lenient | 0.70 | 0.60 | warnings only |

Use Strict for security-sensitive code, Standard for most application code, Lenient for prototypes and exploratory work.
