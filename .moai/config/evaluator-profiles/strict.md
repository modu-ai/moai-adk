# Strict Evaluator Profile

Enhanced security and reliability evaluation for critical systems (auth, payment, migration).

## Evaluation Weights

| Dimension | Weight | Threshold |
|-----------|--------|-----------|
| Functionality | 35% | All acceptance criteria PASS + edge cases verified |
| Security | 35% | Full OWASP Top 10 audit, no findings of any severity |
| Craft | 20% | Coverage >= 90%, mutation testing score >= 80% |
| Consistency | 10% | Strict pattern adherence |

## Hard Thresholds

- Security ANY finding = Overall FAIL (even Low/Info severity)
- Coverage below 90% = Craft FAIL
- Any UNVERIFIED criterion = Overall FAIL (must verify everything)

## Additional Checks

- Input validation on all external boundaries
- Error handling for all failure paths
- No hardcoded credentials or secrets
- Parameterized queries for all database access
- Rate limiting on authentication endpoints
