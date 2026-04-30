---
id: SPEC-SKILL-TEST-001
acceptance_version: "0.1.0"
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
---

# Acceptance Criteria — SPEC-SKILL-TEST-001

## Given-When-Then Scenarios

### Scenario 1: New skill creation triggers test prompt request

**Given** a user invokes builder-skill to create a new skill `moai-domain-foo`

**When** builder-skill begins skill scaffolding

**Then** the orchestrator SHALL receive a request for test prompts from builder-skill
**And** the orchestrator SHALL invoke AskUserQuestion to ask the user for at least 3 sample test prompts
**And** if the user provides prompts, they SHALL be saved to `.claude/skills/moai-domain-foo/tests/sample-prompts.yaml`

---

### Scenario 2: Skill modification without existing tests → prompt for tests

**Given** a user invokes builder-skill to modify an existing skill that has no `tests/` subfolder

**When** builder-skill checks for existing tests

**Then** builder-skill SHALL emit a request to the orchestrator
**And** the orchestrator SHALL ask via AskUserQuestion: "Add tests now?", "Skip (warning logged)", "Skip permanently for this skill"
**And** if user selects "Add tests now", builder-skill SHALL guide test creation

---

### Scenario 3: Test execution measures pass_rate, time, tokens

**Given** a skill `moai-foundation-core` has `tests/sample-prompts.yaml` (3 prompts) and `tests/expected-outcomes.yaml`

**When** builder-skill executes the test run

**Then** the test result file `.moai/research/skill-tests/moai-foundation-core/<run-id>.md` SHALL be created
**And** the file SHALL contain:
- pass_rate (e.g., "2/3" or "67%")
- elapsed_time_avg in seconds
- token_consumption_avg in tokens
- per-prompt result table

---

### Scenario 4: A/B comparator runs in blind mode

**Given** a skill has version `v_current` and version `v_proposed`
**And** both versions produce outputs for the same sample prompt

**When** the A/B comparator agent is spawned

**Then** the comparator prompt SHALL present outputs as `Output A` and `Output B` (random order)
**And** no version label, file path, modification time, or commit hash SHALL appear in the comparator prompt
**And** the comparator output SHALL be one of: "A", "B", "inconclusive" with rationale
**And** the orchestrator SHALL un-shuffle the result to determine winner = current OR proposed

---

### Scenario 5: Description optimization suggests but does not apply

**Given** a skill description has 65% keyword overlap with sample prompts (below 70% threshold)

**When** builder-skill runs description optimization analysis

**Then** builder-skill SHALL emit a suggestion containing: current description, suggested description, rationale, predicted impact
**And** builder-skill SHALL NOT modify the skill frontmatter directly
**And** the orchestrator SHALL surface the suggestion via AskUserQuestion with options: "Apply", "Modify suggestion", "Reject"
**And** only on "Apply" SHALL the frontmatter be edited

---

### Scenario 6: Pass rate below 80% flags skill for refinement

**Given** a skill test run produces pass_rate = 65% (below 80% threshold)

**When** the test result file is generated

**Then** the metadata section SHALL include `needs_refinement: true`
**And** the result file SHALL list specific failing prompts and reasons
**And** builder-skill SHALL recommend next steps in the result summary

---

### Scenario 7: User skips sample prompts → non-blocking warning

**Given** a user creates a new skill but declines to provide sample test prompts

**When** builder-skill proceeds with skill creation

**Then** the skill creation SHALL succeed without error
**And** a non-blocking warning SHALL be logged: "Skill <id> created without tests. Testing is opt-in initially."
**And** the skill SHALL NOT have a `tests/` subfolder

---

### Scenario 8: Legacy skill modification triggers migration prompt

**Given** a legacy skill `moai-foundation-core` exists without `tests/` subfolder
**And** the user modifies the skill body

**When** builder-skill detects the modification

**Then** the orchestrator SHALL receive a migration prompt request
**And** the user SHALL be asked: "This skill has no tests. Add tests now?", "Skip (warning logged)", "Skip permanently"
**And** the bulk migration of all 100+ legacy skills SHALL NOT be triggered

---

### Scenario 9: A/B comparator returns inconclusive → keep current

**Given** A/B comparator evaluates current vs proposed and returns `winner: null` (inconclusive)

**When** builder-skill receives the result

**Then** builder-skill SHALL retain the current version
**And** the inconclusive result SHALL be surfaced to the orchestrator with full comparator rationale
**And** no automatic deployment of the proposed version SHALL occur

---

## Edge Cases

### EC-1: tests/ directory exists but YAML files malformed

If `sample-prompts.yaml` or `expected-outcomes.yaml` is malformed JSON/YAML, builder-skill SHALL log an error, archive the corrupt file as `<filename>.corrupt`, and prompt the user to recreate.

### EC-2: Sample prompts and expected outcomes count mismatch

If `sample-prompts.yaml` has N prompts but `expected-outcomes.yaml` has M (N != M), builder-skill SHALL emit a schema error and refuse to execute tests until corrected.

### EC-3: Skill has external MCP dependencies that are unavailable

If the skill depends on an MCP tool that is offline, the test SHALL log "dependency unavailable" for affected prompts and exclude them from pass_rate calculation.

### EC-4: Test results directory permission denied

If `.moai/research/skill-tests/` cannot be written, builder-skill SHALL emit an error with permission diagnosis and skip persistence (test result available in session only).

### EC-5: A/B comparator prompt accidentally leaks version

If a future change accidentally includes version label in the comparator prompt, the result SHALL be flagged as `non_blind_compromised` and the verdict SHALL be discarded.

---

## Quality Gate Criteria

| Gate | Threshold | Evidence |
|------|-----------|----------|
| Test framework adoption (modified skills, first quarter) | > 50% | quarterly report |
| Pass rate threshold (passing skills) | >= 80% | per-skill measurement |
| Elapsed time per test | < 30s avg | E2E benchmark |
| Token consumption per test | < 5,000 tokens avg | E2E benchmark |
| A/B comparator blind compliance | 100% (no leakage) | prompt audit on 10 runs |
| Description optimization acceptance | > 70% (10 suggestions) | manual review |
| False-trigger reduction (after optimization) | -30% | controlled before/after test |
| Auto-modification of frontmatter | 0% (zero violations) | code audit |
| Template-First sync | clean diff | `make build` |
| plan-auditor | PASS | auditor report |

---

## Definition of Done

- [ ] All 9 Given-When-Then scenarios PASS
- [ ] All 5 edge cases documented and handled (EC-1 through EC-5)
- [ ] All 10 quality gate criteria meet threshold
- [ ] M0 baseline (skill inventory + 5 sample skill state) captured
- [ ] builder-skill agent body contains "Skill Testing Framework" section
- [ ] Skill test schema (sample-prompts.yaml + expected-outcomes.yaml) documented
- [ ] Test result Markdown schema documented
- [ ] A/B comparator blind protocol documented + audit checklist
- [ ] Description optimization suggestion-only enforcement verified
- [ ] Sample 5 skills dogfooded with testing framework
- [ ] CHANGELOG.md updated (opt-in this quarter, mandatory next)
- [ ] docs-site 4개국어 reference (별도 PR via /moai sync)
- [ ] plan-auditor PASS
- [ ] Template-First diff = 0 after `make build`
- [ ] `.moai/research/skill-tests/` added to `.gitignore`

End of acceptance.md.
