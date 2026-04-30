---
id: SPEC-EVAL-LOOP-001
acceptance_version: "0.1.0"
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
---

# Acceptance Criteria — SPEC-EVAL-LOOP-001

## Given-When-Then Scenarios

### Scenario 1: Standard harness activates feedback-loop by default

**Given** a SPEC with no explicit `evaluator_profile` field
**And** harness level is auto-detected as `standard`
**And** `harness.yaml.levels.standard.evaluator_mode = "feedback-loop"`

**When** `/moai run SPEC-XXX` reaches Phase 2.8a (Active Evaluator Pass)

**Then** evaluator-active SHALL execute up to 2 iterations
**And** state file `.moai/state/evaluator-loop/SPEC-XXX.json` SHALL be created
**And** the state file SHALL contain at least one iteration record after the first evaluation

---

### Scenario 2: Pass on first iteration terminates loop early

**Given** standard harness with feedback-loop enabled
**And** the artifact achieves score >= pass_threshold (0.75) on iteration 1

**When** evaluator-active completes iteration 1

**Then** the loop SHALL terminate immediately with PASS verdict
**And** no second iteration SHALL be executed
**And** state file SHALL record `status: "passed"` and `iterations` length = 1

---

### Scenario 3: Iteration 1 fails, iteration 2 passes

**Given** standard harness with feedback-loop enabled
**And** iteration 1 produces score 0.65 (below pass_threshold)

**When** evaluator-active completes iteration 1
**And** generator (manager-ddd/tdd) consumes feedback and produces revised artifact
**And** iteration 2 produces score 0.80

**Then** evaluator-active SHALL execute exactly 2 iterations
**And** the final verdict SHALL be PASS
**And** state file iterations array SHALL have length 2 with scores [0.65, 0.80]

---

### Scenario 4: Max iterations reached without passing → escalation

**Given** standard harness with feedback-loop enabled
**And** iteration 1 score = 0.55, iteration 2 score = 0.68

**When** evaluator-active completes iteration 2 (max_iterations)

**Then** evaluator-active SHALL emit a structured escalation report including final score, dimension breakdown, and unresolved findings
**And** the orchestrator SHALL receive the escalation and surface it to the user
**And** state file SHALL record `status: "escalated_max_iterations"`

---

### Scenario 5: Stagnation triggers early escalation

**Given** standard harness with feedback-loop enabled
**And** improvement_threshold = 0.10
**And** stagnation_consecutive = 1
**And** iteration 1 score = 0.60, iteration 2 score = 0.65 (delta 0.05 < 0.10)

**When** evaluator-active computes the iteration-2 delta

**Then** the loop SHALL be flagged as stagnating
**And** evaluator-active SHALL escalate to orchestrator without performing further iterations
**And** state file SHALL record `status: "escalated_stagnation"`

---

### Scenario 6: SPEC opts out via `evaluator_loop_disabled: true`

**Given** a SPEC with frontmatter `evaluator_loop_disabled: true`
**And** harness level is `standard`

**When** `/moai run SPEC-XXX` reaches Phase 2.8a

**Then** evaluator-active SHALL execute final-pass behavior (single evaluation)
**And** no iteration loop SHALL be entered
**And** no state file SHALL be created

---

### Scenario 7: Backward compatibility — thorough mode unchanged

**Given** a SPEC with harness level `thorough`
**And** sprint_contract = true (existing thorough behavior)

**When** `/moai run SPEC-XXX` runs

**Then** Phase 2.0 contract negotiation SHALL execute as before
**And** Phase 2.8a per-sprint evaluation SHALL execute as before
**And** the new feedback-loop config keys SHALL NOT affect thorough behavior
**And** no `.moai/state/evaluator-loop/` file SHALL be created for thorough SPEC

---

### Scenario 8: Token cost stays within budget

**Given** baseline measurement on 5 standard SPEC samples (M0 milestone)
**And** baseline average token cost per evaluation = T_baseline

**When** the same 5 SPECs are re-evaluated under feedback-loop mode

**Then** the average token cost across the 5 samples SHALL NOT exceed `T_baseline * 1.25`
**And** the cost increase SHALL be documented in the validation report at `.moai/reports/eval-loop-validation/<DATE>.md`

---

## Edge Cases

### EC-1: Configuration absent

If `harness.yaml.levels.standard.evaluator_mode` is missing, the system SHALL fall back to final-pass with a non-blocking warning logged to stderr.

### EC-2: State file corruption

If `.moai/state/evaluator-loop/<SPEC-ID>.json` is malformed, evaluator-active SHALL ignore the corrupt state and start fresh from iteration 1, logging a warning.

### EC-3: Generator unavailable for feedback hand-off

If the generator (manager-ddd/tdd) cannot consume feedback (e.g., agent crashed), evaluator-active SHALL escalate immediately with `status: "escalated_handoff_failure"`.

### EC-4: Both `evaluator_loop_disabled: true` AND thorough harness

`evaluator_loop_disabled` is a no-op for thorough harness. thorough behavior is unchanged.

---

## Quality Gate Criteria

| Gate | Threshold | Evidence |
|------|-----------|----------|
| feedback-loop activation rate (standard) | 100% (when default ON) | E2E test |
| token cost increase | < +25% | M0 baseline + M6 measurement |
| stagnation false-positive rate | < 20% | controlled test cases |
| backward compatibility (thorough/minimal) | zero diff | regression suite |
| Template-First sync | clean | `make build` diff |
| plan-auditor validation | PASS | auditor report |

---

## Definition of Done

- [ ] All 8 Given-When-Then scenarios PASS
- [ ] All 4 edge cases documented and handled (EC-1 through EC-4)
- [ ] All 6 quality gate criteria meet threshold
- [ ] M0 baseline measurement report committed to `.moai/reports/eval-loop-baseline/`
- [ ] M6 validation report committed to `.moai/reports/eval-loop-validation/`
- [ ] CHANGELOG.md updated under Unreleased
- [ ] docs-site 4개국어 reference 업데이트 (별도 PR via /moai sync)
- [ ] plan-auditor PASS
- [ ] Template-First diff = 0 after `make build`
- [ ] Hard rule audit: no frontmatter custom field violation

End of acceptance.md.
