# Lenient Evaluator Profile

Relaxed evaluation for prototypes, experiments, and non-production code.

## Evaluation Dimensions

| Dimension | Weight | Pass Threshold |
|-----------|--------|----------------|
| Functionality | 60% | Core acceptance criteria PASS |
| Security | 20% | No Critical findings only |
| Craft | 10% | Coverage >= 60% |
| Consistency | 10% | Basic pattern compliance |

## Must-Pass Criteria

- Security: No Critical severity findings (all other severities are warnings only)
- Functionality: Core happy-path acceptance criteria must pass

## Hard Thresholds

- Security Critical = Overall FAIL
- High/Medium security findings = WARNING only (not FAIL)

## Relaxed Rules

- UNVERIFIED criteria acceptable (prototype stage)
- Lower coverage threshold (60%)
- Pattern deviations acceptable with rationale
- Focus on "does it work" over "is it perfect"

## Scoring Rubric

### Functionality (60%)

| Score | Description |
|-------|-------------|
| 1.00 | All acceptance criteria pass including edge cases |
| 0.75 | Core acceptance criteria pass; some edge cases untested |
| 0.50 | Happy-path works; error paths unverified (acceptable for prototype) |
| 0.25 | Basic intent demonstrated; multiple core criteria fail or are unverified |

### Security (20%)

| Score | Description |
|-------|-------------|
| 1.00 | No findings of any severity |
| 0.75 | No Critical/High; Medium findings present |
| 0.50 | No Critical; High findings present (WARNING, not FAIL) |
| 0.25 | Critical findings present (triggers overall FAIL) |

### Craft (10%)

| Score | Description |
|-------|-------------|
| 1.00 | Coverage >= 85%, clean and readable code |
| 0.75 | Coverage >= 70%, generally readable |
| 0.50 | Coverage >= 60%, some issues but functional |
| 0.25 | Coverage below 60% or significantly unclear code |

### Consistency (10%)

| Score | Description |
|-------|-------------|
| 1.00 | Fully consistent with project conventions |
| 0.75 | Minor deviations; still recognizably following project style |
| 0.50 | Some deviations; basic conventions followed |
| 0.25 | Significantly different from existing patterns (acceptable with rationale) |
