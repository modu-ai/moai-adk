# Lenient Evaluator Profile

Relaxed evaluation for prototypes, experiments, and non-production code.

## Evaluation Weights

| Dimension | Weight | Threshold |
|-----------|--------|-----------|
| Functionality | 60% | Core acceptance criteria PASS |
| Security | 20% | No Critical findings only |
| Craft | 10% | Coverage >= 60% |
| Consistency | 10% | Basic pattern compliance |

## Hard Thresholds

- Security Critical = Overall FAIL
- High/Medium security findings = WARNING only (not FAIL)

## Relaxed Rules

- UNVERIFIED criteria acceptable (prototype stage)
- Lower coverage threshold (60%)
- Pattern deviations acceptable with rationale
- Focus on "does it work" over "is it perfect"
