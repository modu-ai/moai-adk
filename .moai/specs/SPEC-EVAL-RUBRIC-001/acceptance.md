---
id: SPEC-EVAL-RUBRIC-001
acceptance_version: "0.1.0"
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
---

# Acceptance Criteria — SPEC-EVAL-RUBRIC-001

## Given-When-Then Scenarios

### Scenario 1: Default profile has 4-anchor for all 4 dimensions

**Given** the file `.moai/config/evaluator-profiles/default.md` exists

**When** the file is parsed for rubric anchors

**Then** the file SHALL contain anchor entries at scores 0.25, 0.50, 0.75, 1.00 for each of the 4 dimensions
**And** the total anchor count SHALL be 16 (4 dimensions × 4 levels)
**And** each anchor SHALL be expressed in observable terms (no subjective adjectives without qualifying detail)

---

### Scenario 2: Strict profile is more demanding than default

**Given** both `default.md` and `strict.md` profiles exist
**And** both contain Functionality dimension at 1.00 anchor

**When** anchors are compared side-by-side

**Then** the strict profile's 1.00 anchor SHALL be more demanding than the default's 1.00 anchor (e.g., requires additional property/fuzz tests)
**And** at least 2 of the 4 dimensions SHALL show stricter tone in strict.md vs default.md

---

### Scenario 3: Lenient profile is more permissive than default

**Given** both `default.md` and `lenient.md` profiles exist

**When** anchors are compared

**Then** the lenient profile's 0.50 anchor for at least 2 dimensions SHALL be more permissive than default's 0.50 anchor
**And** the lenient profile SHALL retain a 1.00 anchor that prevents leniency drift

---

### Scenario 4: Frontend profile applies 4-anchor to its dimensions

**Given** `frontend.md` profile exists
**And** it may have design-specific dimensions (Design Quality, Originality, Completeness, Functionality)

**When** the file is parsed

**Then** every dimension defined in the frontend profile SHALL have anchors at 0.25, 0.50, 0.75, 1.00
**And** no dimension SHALL be missing an anchor

---

### Scenario 5: Evaluator-active body contains "Rubric Anchoring is Mandatory" section

**Given** `.claude/agents/moai/evaluator-active.md` is read

**When** the body is searched for the mandatory anchoring section

**Then** a section titled "Rubric Anchoring is Mandatory" SHALL be present
**And** the section SHALL describe: (1) the requirement to cite, (2) the citation format, (3) the fallback to default profile, (4) the invalidation policy

---

### Scenario 6: Evaluator output includes anchor citations

**Given** evaluator-active is invoked on a sample SPEC
**And** default profile is loaded

**When** the evaluator produces the dimensional score breakdown

**Then** every score row SHALL include an "Anchor Citation" field with text quoted from the profile's rubric
**And** the format SHALL match: "{Dimension} {score} — matches anchor: '{anchor text}'"
**And** all 4 dimensions SHALL have citations (4/4)

---

### Scenario 7: Profile missing anchor → fallback + warning

**Given** a custom profile is loaded that lacks the 0.75 anchor for the Security dimension

**When** evaluator-active assigns Security score = 0.75

**Then** the evaluator SHALL fall back to the default profile's Security 0.75 anchor
**And** the evaluator SHALL emit a non-blocking warning in the report metadata: "Anchor fallback used: Security 0.75 from default profile"

---

### Scenario 8: Token overhead remains within budget

**Given** baseline measurement on 5 sample evaluations (M0)
**And** baseline average evaluation tokens = T_baseline

**When** the same 5 evaluations are re-run with anchor citation enforcement

**Then** the average evaluation tokens SHALL NOT exceed `T_baseline * 1.05` (5% overhead)
**And** the overhead SHALL be documented in the validation report

---

## Edge Cases

### EC-1: Score not in {0.25, 0.50, 0.75, 1.00}

If the evaluator assigns an intermediate score (e.g., 0.65), it SHALL cite the closest anchor below (0.50) AND the closest anchor above (0.75), with the format: "between '<lower anchor>' and '<higher anchor>', closer to lower".

### EC-2: Profile file unreadable

If `.moai/config/evaluator-profiles/<profile>.md` is unreadable (permissions, corruption), the evaluator SHALL fall back to the built-in default profile and emit a warning.

### EC-3: All 4 anchors missing for a dimension

If a profile has zero anchors for a dimension, that profile SHALL be flagged as schema-violating in the evaluator output, and the dimension SHALL inherit anchors from the built-in default.

### EC-4: Evaluator output explicitly skips citation

If the evaluator omits citation despite the mandatory rule, the report SHALL include "Unanchored scores: N" in the footer and severity SHALL be `warning` (initial hardening).

---

## Quality Gate Criteria

| Gate | Threshold | Evidence |
|------|-----------|----------|
| Default profile 4×4 anchor count | 16 | manual verification |
| All 4 profiles have 4-anchor for each dimension | 100% | grep audit |
| Strict tone differentiation | >= 2 dimensions stricter | manual review |
| Lenient tone differentiation | >= 2 dimensions more permissive | manual review |
| Anchor citation rate in evaluator output | 100% per dimension | E2E test |
| Token overhead | < +5% | M0 baseline + M6 measurement |
| Unanchored score warning visibility | 100% when violation occurs | controlled test |
| Template-First sync | clean diff | `make build` |
| plan-auditor | PASS | auditor report |

---

## Definition of Done

- [ ] All 8 Given-When-Then scenarios PASS
- [ ] All 4 edge cases documented and handled (EC-1 through EC-4)
- [ ] All 9 quality gate criteria meet threshold
- [ ] M0 baseline captured (5 sample evaluations + 4 profile audit)
- [ ] All 4 profiles have 4-anchor per dimension (verified)
- [ ] evaluator-active.md "Rubric Anchoring is Mandatory" section present
- [ ] CHANGELOG.md updated
- [ ] docs-site 4개국어 reference (별도 PR via /moai sync)
- [ ] plan-auditor PASS
- [ ] Template-First diff = 0 after `make build`
- [ ] design `§12` Mechanism 1 not regressed (separate domain, independent verification)

End of acceptance.md.
