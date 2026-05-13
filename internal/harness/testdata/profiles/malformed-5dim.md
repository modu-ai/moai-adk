# Malformed Profile — 5th Unknown Dimension (Test Fixture)

> NEGATIVE FIXTURE: This profile intentionally includes an unknown dimension "Performance"
> alongside only 3 canonical dimensions (missing Consistency), to trigger rejection.
> Used by TestParseRubricMarkdown_RejectsFifthDimension.
>
> Design intent: Parser skips "Performance" (unknown), result has only 3 canonical dims.
> Rubric.Validate() then rejects: "must have exactly 4 dimensions, got 3"
> (ErrInvalidConfig wrapping HRN_UNKNOWN_DIMENSION via ValidationErrors).

## Evaluation Dimensions

| Dimension     | Weight | Pass Threshold |
|---------------|--------|----------------|
| Functionality | 25%    | 0.60           |
| Security      | 25%    | 0.60           |
| Craft         | 25%    | 0.60           |
| Performance   | 25%    | 0.60           |

## Must-Pass Criteria

- Functionality: All core features must work.
- Security: No security vulnerabilities.

## Scoring Rubric

### Functionality

| Score | Description                    |
|-------|-------------------------------|
| 0.25  | Major functionality missing   |
| 0.50  | Partial functionality present |
| 0.75  | Most functionality working    |
| 1.00  | Complete functionality        |

### Security

| Score | Description                    |
|-------|-------------------------------|
| 0.25  | Critical vulnerabilities      |
| 0.50  | Moderate issues               |
| 0.75  | Minor issues only             |
| 1.00  | No security issues            |

### Craft

| Score | Description                    |
|-------|-------------------------------|
| 0.25  | Poor code quality             |
| 0.50  | Acceptable quality            |
| 0.75  | Good quality                  |
| 1.00  | Excellent quality             |

### Performance

| Score | Description                    |
|-------|-------------------------------|
| 0.25  | Very slow                     |
| 0.50  | Acceptable speed              |
| 0.75  | Good performance              |
| 1.00  | Excellent performance         |
