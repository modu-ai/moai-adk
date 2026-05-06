---
id: SPEC-LOOP-TERM-001
acceptance_version: "0.1.0"
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
---

# Acceptance Criteria — SPEC-LOOP-TERM-001

## Given-When-Then Scenarios

### Scenario 1: Canonical reference exists and is referenced by all 4 workflows

**Given** the file `.claude/rules/moai/workflow/iteration-termination.md` exists
**And** workflows loop, fix, coverage, e2e are present in `.claude/skills/moai/workflows/`

**When** a grep search for `iteration-termination.md` is executed across the 4 workflows

**Then** all 4 workflows SHALL contain at least one reference to the canonical file
**And** all 4 workflows SHALL declare their `max_iterations` override value (5, 3, 3, 2 respectively)

---

### Scenario 1b: Termination schema declares all required fields with dot notation

**Given** the canonical reference `.claude/rules/moai/workflow/iteration-termination.md` is parsed for the termination schema
**And** the schema is the single source of truth (REQ-LT-001)

**When** the schema fields are enumerated

**Then** the schema SHALL include exactly these fields with the following dot-notation keys:
- `max_iterations`
- `stagnation.detect_after`
- `stagnation.improvement_min`
- `stagnation.consecutive`
- `escalation.target`
- `escalation.reason_required`
- `state_file`

**And** no underscore-flattened variants (e.g., `stagnation_consecutive`) SHALL appear in any moai workflow document
**And** plan.md §4.1 schema example SHALL match this field set verbatim

---

### Scenario 2: Loop workflow respects max_iterations boundary

**Given** `/moai loop` is invoked with a goal that requires more than 5 iterations
**And** `max_iterations: 5` (default for loop)

**When** the workflow executes 5 iterations without success

**Then** at iteration 5 completion, escalation SHALL be triggered
**And** the orchestrator SHALL invoke `ToolSearch(query: "select:AskUserQuestion")` before the question
**And** AskUserQuestion SHALL present 3 standardized options + Other
**And** state file `.moai/state/loop/<run_id>.json` SHALL record `status: "escalated_max_iterations"`

---

### Scenario 3: Fix workflow stops at max_iterations 3

**Given** `/moai fix` is invoked
**And** `max_iterations: 3` (default for fix)

**When** 3 iterations complete without resolving the failure

**Then** the workflow SHALL escalate at iteration 3 boundary
**And** state file `.moai/state/fix/<run_id>.json` SHALL contain 3 iteration records

---

### Scenario 4: Coverage workflow detects stagnation

**Given** `/moai coverage` is invoked
**And** `improvement_min: 0.05` interpreted as 1.0 percent coverage delta
**And** iteration 1 coverage = 75.0%, iteration 2 coverage = 75.5%

**When** stagnation detection runs at iteration 2

**Then** the delta SHALL be measured as 0.5 percent (below threshold)
**And** the workflow SHALL flag stagnation and escalate
**And** state file SHALL record `status: "escalated_stagnation"`

---

### Scenario 5: E2E workflow with stagnation detection on passing scenario count

**Given** `/moai e2e` is invoked
**And** iteration 1 produces 5 passing scenarios, iteration 2 produces 5 passing scenarios (delta = 0)

**When** stagnation detection runs at iteration 2

**Then** the workflow SHALL flag stagnation
**And** `max_iterations: 2` for e2e means escalation is triggered
**And** state file `.moai/state/e2e/<run_id>.json` SHALL be persisted

---

### Scenario 6: Subagent attempts AskUserQuestion → blocker report instead

**Given** a subagent (e.g., manager-quality) is performing a fix iteration
**And** the subagent encounters a stuck condition

**When** the subagent decides to escalate

**Then** the subagent SHALL NOT invoke AskUserQuestion directly
**And** the subagent SHALL return a structured blocker report to the orchestrator
**And** the orchestrator SHALL handle ToolSearch + AskUserQuestion sequence

---

### Scenario 7: Resume from existing state file

**Given** a previous workflow run was aborted at iteration 2
**And** state file `.moai/state/loop/run-abc.json` exists with `status: "aborted"` and 2 iteration records

**When** `/moai loop` is invoked again with the same context

**Then** the workflow SHALL detect the existing state file
**And** the orchestrator SHALL invoke AskUserQuestion offering: Resume from iteration 3, Start fresh, or Abort
**And** if user selects Resume, iteration 3 SHALL be the next iteration executed

---

### Scenario 8: Backward compatibility — design `§11` GAN Loop unchanged

**Given** the `/moai design` workflow uses GAN Loop Contract from `design.yaml §11`
**And** `gan_loop.max_iterations: 5`, `improvement_threshold: 0.05`, `escalation_after: 3`

**When** the new termination schema is introduced

**Then** the `/moai design` workflow SHALL continue using `gan_loop.*` keys
**And** no `termination.*` keys SHALL be applied to design domain
**And** the canonical reference SHALL explicitly note the design domain exception

---

### Scenario 9: Escalation reason_required produces structured rationale

**Given** `escalation.reason_required: true` (default)
**And** a workflow triggers escalation

**When** the escalation report is produced

**Then** the report SHALL include a structured "Rationale" section with:
- Trigger type (max_iterations OR stagnation OR explicit_abort)
- Current state summary
- Suggested next action

---

## Edge Cases

### EC-1: state file directory missing

If `.moai/state/<workflow>/` does not exist, the workflow SHALL create it with `0700` permissions before writing the state file.

### EC-2: Concurrent runs of same workflow

Multiple `run_id` UUIDs prevent collision. If two workflows share the same `run_id` (impossible in practice), the second SHALL detect existing file and offer overwrite/resume choice.

### EC-3: AskUserQuestion deferred tool not preloaded

If orchestrator forgets to invoke ToolSearch before AskUserQuestion, the resulting `InputValidationError` SHALL be caught, ToolSearch invoked, and AskUserQuestion retried once.

### EC-4: state file corruption

If JSON is malformed, the workflow SHALL log a warning, archive the corrupt file as `<run_id>.json.corrupt`, and start fresh from iteration 1.

### EC-5: workflow without explicit override

If a future workflow inherits termination schema without specifying `max_iterations`, the canonical default 5 SHALL apply.

---

## Quality Gate Criteria

| Gate | Threshold | Evidence |
|------|-----------|----------|
| canonical reference adoption | 4/4 workflows | grep result |
| state file persistence | 100% of iteration boundaries | E2E test |
| escalation reach user | 100% of max_iterations or stagnation triggers | E2E test |
| design `§11` regression | zero diff | regression test |
| Template-First sync | clean | `make build` diff |
| plan-auditor | PASS | auditor report |
| AskUserQuestion compliance (orchestrator-only) | 100% | code audit |

---

## Definition of Done

- [ ] All 9 Given-When-Then scenarios PASS
- [ ] All 5 edge cases documented and handled (EC-1 through EC-5)
- [ ] All 7 quality gate criteria meet threshold
- [ ] M0 baseline captured for regression
- [ ] Canonical reference file exists with all 4 fields documented
- [ ] All 4 workflows reference the canonical file
- [ ] State file schema documented
- [ ] CHANGELOG.md updated
- [ ] docs-site 4개국어 reference (별도 PR via /moai sync)
- [ ] plan-auditor PASS
- [ ] Template-First diff = 0 after `make build`

End of acceptance.md.
