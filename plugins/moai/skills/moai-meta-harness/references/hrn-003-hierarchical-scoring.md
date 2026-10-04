# Phase 3b — HRN-003 Hierarchical Scoring

When `harness.yaml` sets `evaluator_mode: hierarchical` (the hierarchical-scoring policy), the Sprint Contract evaluation uses 4-dimension × sub-criteria hierarchical scoring instead of flat 0-100 integers.

## Flat Mode vs Hierarchical Mode

| Aspect | Flat Mode (default) | Hierarchical Mode (HRN-003) |
|--------|---------------------|-----------------------------|
| Score type | Integer 0-100 | Float anchor: 0.25 / 0.50 / 0.75 / 1.00 |
| Criteria granularity | Per-dimension | Per sub-criterion within dimension |
| Aggregation | Implicit average | `min` (default) or `mean` per profile |
| Citation | Optional | Required (ErrRubricCitationMissing if absent) |
| Must-pass | Security hard threshold | Per `must_pass_dimensions` in profile |

## Active Profile Loading

1. Check SPEC frontmatter `evaluator_profile` field.
2. If present: load `.moai/config/evaluator-profiles/{evaluator_profile}.md`
3. If absent: load `.moai/config/evaluator-profiles/{harness.default_profile}.md`
4. If profile file not found: fall back to built-in defaults (Functionality / Security / Craft / Consistency)

Profiles are parsed by `internal/harness.ParseRubricMarkdown()`. Unknown dimension names are skipped (lenient parsing) to support non-canonical profiles like `frontend.md`. Validate() enforces that exactly 4 canonical dimensions are present after parsing.

## When to Enable Hierarchical Mode

Hierarchical mode is recommended when:

- Sub-criteria need independent visibility (e.g., "Security → input validation" vs "Security → secret management")
- Project has compliance requirements requiring per-criterion audit trail
- Evaluator leniency is a concern (citation requirement reduces drift)

Flat mode is sufficient for:

- Simple harnesses with limited domain scope
- Projects where the 4-dimension aggregate is the primary signal
- Initial harness generation before profiles are tuned
