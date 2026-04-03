# Default Evaluator Profile

Standard skeptical evaluation for general-purpose code review.

## Evaluation Weights

| Dimension | Weight | Threshold |
|-----------|--------|-----------|
| Functionality | 40% | All acceptance criteria PASS |
| Security | 25% | No Critical/High findings |
| Craft | 20% | Coverage >= 85% |
| Consistency | 15% | No major pattern violations |

## Hard Thresholds

- Security FAIL = Overall FAIL (regardless of other scores)
- Coverage below 85% = Craft FAIL

## Evaluation Rules

- Require concrete evidence for every PASS verdict
- Mark unverifiable criteria as UNVERIFIED, not PASS
- Report all findings with file:line references
- Provide actionable fix recommendations for every FAIL
