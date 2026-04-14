# Default Evaluator Profile

Standard skeptical evaluation for general-purpose code review.

## Evaluation Dimensions

| Dimension | Weight | Pass Threshold |
|-----------|--------|----------------|
| Functionality | 40% | All acceptance criteria PASS |
| Security | 25% | No Critical/High findings |
| Craft | 20% | Coverage >= 85% |
| Consistency | 15% | No major pattern violations |

## Must-Pass Criteria

- Functionality: All SPEC acceptance criteria must be met (no partial credit)
- Security: No Critical or High severity findings (FAIL overrides overall score)

## Hard Thresholds

- Security FAIL = Overall FAIL (regardless of other scores)
- Coverage below 85% = Craft FAIL

## Evaluation Rules

- Require concrete evidence for every PASS verdict
- Mark unverifiable criteria as UNVERIFIED, not PASS
- Report all findings with file:line references
- Provide actionable fix recommendations for every FAIL

## Scoring Rubric

### Functionality (40%)

| Score | Description |
|-------|-------------|
| 1.00 | All acceptance criteria pass with edge cases verified |
| 0.75 | All primary acceptance criteria pass; minor edge cases missing |
| 0.50 | Core functionality works; 1-2 acceptance criteria fail or are unverified |
| 0.25 | Basic skeleton present but multiple acceptance criteria fail |

### Security (25%)

| Score | Description |
|-------|-------------|
| 1.00 | No findings of any severity; OWASP Top 10 checked |
| 0.75 | No Critical/High findings; Medium findings documented with mitigations |
| 0.50 | No Critical findings; High findings present but contained |
| 0.25 | Critical or multiple High findings present (triggers overall FAIL) |

### Craft (20%)

| Score | Description |
|-------|-------------|
| 1.00 | Coverage >= 85%, clean code, no duplication, clear naming |
| 0.75 | Coverage >= 80%, minor style issues, acceptable naming |
| 0.50 | Coverage >= 70%, some duplication or unclear naming |
| 0.25 | Coverage < 70% or significant code quality issues |

### Consistency (15%)

| Score | Description |
|-------|-------------|
| 1.00 | Fully consistent with project conventions and existing patterns |
| 0.75 | Minor deviations from conventions; no structural inconsistencies |
| 0.50 | Some pattern violations; deviations are localized |
| 0.25 | Significant inconsistencies with existing codebase patterns |
