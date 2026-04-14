# Strict Evaluator Profile

Enhanced security and reliability evaluation for critical systems (auth, payment, migration).

## Evaluation Dimensions

| Dimension | Weight | Pass Threshold |
|-----------|--------|----------------|
| Functionality | 35% | All acceptance criteria PASS + edge cases verified |
| Security | 35% | Full OWASP Top 10 audit, no findings of any severity |
| Craft | 20% | Coverage >= 90%, mutation testing score >= 80% |
| Consistency | 10% | Strict pattern adherence |

## Must-Pass Criteria

- All four dimensions must individually score >= 80% (0.80) to pass
- Security: Zero findings at any severity level (Critical/High/Medium/Low/Info)
- Functionality: All edge cases must be explicitly tested and verified
- No UNVERIFIED criteria allowed — everything must be demonstrated

## Hard Thresholds

- Security ANY finding = Overall FAIL (even Low/Info severity)
- Coverage below 90% = Craft FAIL
- Any UNVERIFIED criterion = Overall FAIL (must verify everything)
- Any dimension below 0.80 = Overall FAIL

## Additional Checks

- Input validation on all external boundaries
- Error handling for all failure paths
- No hardcoded credentials or secrets
- Parameterized queries for all database access
- Rate limiting on authentication endpoints

## Scoring Rubric

### Functionality (35%)

| Score | Description |
|-------|-------------|
| 1.00 | All acceptance criteria pass; all edge cases verified with tests; no UNVERIFIED items |
| 0.75 | All primary criteria pass; edge cases documented but not all tested (triggers FAIL in strict) |
| 0.50 | Core functionality verified; some edge cases missing (FAIL) |
| 0.25 | Multiple criteria unverified or failing (FAIL) |

### Security (35%)

| Score | Description |
|-------|-------------|
| 1.00 | Zero findings at any severity; full OWASP Top 10 audit passed; threat model reviewed |
| 0.75 | No Critical/High; Medium findings present with mitigations (triggers FAIL in strict) |
| 0.50 | No Critical; High or Medium findings present (FAIL) |
| 0.25 | Critical findings present (FAIL) |

### Craft (20%)

| Score | Description |
|-------|-------------|
| 1.00 | Coverage >= 90%, mutation score >= 80%, zero linting warnings, excellent naming |
| 0.75 | Coverage >= 85%, mutation score >= 70%, minor issues (triggers FAIL in strict) |
| 0.50 | Coverage >= 80%, mutation score below 70% (FAIL) |
| 0.25 | Coverage below 80% or no mutation testing (FAIL) |

### Consistency (10%)

| Score | Description |
|-------|-------------|
| 1.00 | All code strictly follows established project patterns and conventions |
| 0.75 | Minor deviations with documented rationale (triggers FAIL in strict) |
| 0.50 | Multiple deviations from conventions (FAIL) |
| 0.25 | Significant structural inconsistencies (FAIL) |
