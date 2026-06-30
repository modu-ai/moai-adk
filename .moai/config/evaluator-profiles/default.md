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

## Finding-Stage Reporting (coverage before filtering)

At the finding stage, report every issue you find, including ones you are uncertain about or consider low-severity, each with a confidence level and an estimated severity. Do not filter for importance or confidence while finding — the verdict stage (must-pass thresholds + harmonic scoring) does the filtering downstream. The goal at this stage is coverage: surfacing a finding that later gets filtered out is preferable to silently dropping a real bug.

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

## D7/D8 Plan-Phase Dimensions

The following two dimensions are evaluated by plan-auditor during plan-phase
audit, not by sync-auditor during run-phase. They are listed here for
cross-reference visibility — the weights apply ONLY to plan-auditor's PASS/FAIL
decision and do NOT contribute to sync-auditor's 4-dimension overall score
(Functionality/Security/Craft/Consistency above).

| Plan-Auditor Dimension | Weight in Plan Verdict | Pass Threshold |
|------------------------|-----------------------:|----------------|
| D7 Cross-SPEC Reconciliation | 50% | No BLOCKING finding |
| D8 Cross-Platform Discipline | 50% | No BLOCKING finding |

Both dimensions use binary BLOCKING/PASS semantics — any single BLOCKING
finding fails plan-auditor's overall verdict regardless of other scores.

See `.claude/agents/moai/plan-auditor.md` Group 7 + Group 8 for the full
verification verbs and rubric.
