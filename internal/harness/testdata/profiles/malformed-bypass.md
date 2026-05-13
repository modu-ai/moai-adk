# Malformed Profile — Must-Pass Bypass (Test Fixture)

> NEGATIVE FIXTURE: This profile intentionally excludes Security from Must-Pass criteria,
> attempting to bypass the must-pass firewall. Used by TestParseRubricMarkdown_MustPassBypassRejected.
>
> Design intent: Parser reads the Must-Pass section and finds Security is absent.
> Rubric.Validate() returns ErrMustPassBypassProhibited because Security is a floor dimension
> that cannot be bypassed (design-constitution §12 Mechanism 3, REQ-HRN-003-018).

## Evaluation Dimensions

| Dimension     | Weight | Pass Threshold |
|---------------|--------|----------------|
| Functionality | 25%    | 0.60           |
| Security      | 25%    | 0.60           |
| Craft         | 25%    | 0.60           |
| Consistency   | 25%    | 0.60           |

## Must-Pass Criteria

- Functionality: All core features must work.

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

### Consistency

| Score | Description                    |
|-------|-------------------------------|
| 0.25  | Major inconsistencies         |
| 0.50  | Some inconsistencies          |
| 0.75  | Mostly consistent             |
| 1.00  | Fully consistent              |
