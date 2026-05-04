---
id: SPEC-SKILL-DESC-001
acceptance_version: "0.1.0"
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
---

# Acceptance Criteria — SPEC-SKILL-DESC-001

## Given-When-Then Scenarios

### Scenario 1: SPEC-SKILL-TEST-001 dependency 검증

**Given** SPEC-SKILL-TEST-001 (Wave 2) framework is installed

**When** the user invokes `moai skill optimize moai-workflow-spec`

**Then** the optimizer SHALL successfully call SPEC-SKILL-TEST-001 framework
**And** the framework SHALL return FP and FN measurements
**And** the optimizer SHALL produce a report referencing the framework's sample prompts

---

### Scenario 2: 임계값 미초과 → no optimization needed

**Given** a skill with FP = 0.10 (below 0.15) and FN = 0.05 (below 0.10)

**When** the user invokes `moai skill optimize <name>`

**Then** the optimizer SHALL emit "no optimization needed" message
**And** no LLM call SHALL be made for description rewriting
**And** the skill frontmatter SHALL NOT be modified

---

### Scenario 3: FP 임계값 초과 → tightening 제안

**Given** a skill with FP = 0.22 (above 0.15) and FN = 0.05 (below 0.10)

**When** the user invokes `moai skill optimize <name>`

**Then** the optimizer SHALL invoke `SuggestTightening` LLM call
**And** the suggestion SHALL be a narrower description targeting FP reduction
**And** the report SHALL display before/after diff
**And** the optimizer SHALL pause for user approval (default behavior)

---

### Scenario 4: FN 임계값 초과 → broadening 제안

**Given** a skill with FP = 0.08 (below 0.15) and FN = 0.18 (above 0.10)

**When** the user invokes `moai skill optimize <name>`

**Then** the optimizer SHALL invoke `SuggestBroadening` LLM call
**And** the suggestion SHALL be a wider description targeting FN reduction
**And** the report SHALL display before/after diff
**And** the optimizer SHALL pause for user approval

---

### Scenario 5: 사용자 승인 → apply + regression

**Given** the optimizer has produced a tightening suggestion
**And** the user approves via AskUserQuestion or `--apply` flag

**When** the optimizer applies the change

**Then** the skill frontmatter `description:` SHALL be updated to the new description
**And** the optimizer SHALL re-run SPEC-SKILL-TEST-001 framework
**And** post-apply FP/FN metrics SHALL be measured
**And** the optimizer SHALL run full skill catalog regression
**And** if cross-effect > 5% routing variation, the optimizer SHALL automatically rollback

---

### Scenario 6: User rejection — no apply

**Given** the optimizer has produced a suggestion
**And** the user rejects via AskUserQuestion (selects "Reject")

**When** the optimizer processes the rejection

**Then** the skill frontmatter SHALL NOT be modified
**And** the report SHALL note the rejection
**And** no audit log entry SHALL be created for the rejected suggestion
**And** the optimizer SHALL exit with code 0 (clean rejection)

---

### Scenario 7: Auto-apply block (no approval, no flag)

**Given** the optimizer is invoked without `--apply` flag and without user approval

**When** the optimizer attempts to write the new description

**Then** the optimizer SHALL refuse to write and emit error: "user approval required (use --apply or interactive confirmation)"
**And** the skill frontmatter SHALL NOT be modified
**And** the optimizer SHALL exit with non-zero code

---

### Scenario 8: Sample data insufficient → halt

**Given** a skill has only 3 sample prompts in SPEC-SKILL-TEST-001 framework
**And** minimum threshold is 5

**When** the user invokes `moai skill optimize <name>`

**Then** the optimizer SHALL halt with error: "data insufficient: 3 sample prompts < 5 minimum"
**And** no LLM call SHALL be made
**And** no skill frontmatter modification SHALL occur

---

### Scenario 9: Convergence detection (loop prevention)

**Given** the optimizer has previously suggested description "X" for a skill (recorded in state file)

**When** the user invokes `moai skill optimize <name>` again
**And** the new suggestion is also "X" (same as last)

**Then** the optimizer SHALL detect convergence
**And** the optimizer SHALL emit "convergence reached" message
**And** the optimizer SHALL halt without applying

---

### Scenario 10: Per-skill threshold override

**Given** a skill frontmatter contains:
```yaml
optimization_thresholds:
  fp: 0.20
  fn: 0.15
```
**And** the skill has FP = 0.18 and FN = 0.12

**When** the user invokes `moai skill optimize <name>`

**Then** the optimizer SHALL use the override thresholds
**And** since FP (0.18) < FP threshold (0.20) AND FN (0.12) < FN threshold (0.15), the optimizer SHALL emit "no optimization needed"

---

## Edge Cases

### EC-1: Both FP and FN exceed thresholds
If both FP > 0.15 and FN > 0.10, the optimizer SHALL prioritize tightening (FP first), then run broadening in a separate invocation if needed.

### EC-2: SPEC-SKILL-TEST-001 framework unavailable
If the dependency is missing or returns errors, the optimizer SHALL emit a blocker error referencing the missing dependency. No fallback measurement is permitted.

### EC-3: Cross-effect rollback failure
If automatic rollback after cross-effect detection fails (e.g., file write error), the optimizer SHALL emit a critical error and recommend manual revert via `git checkout`.

### EC-4: LLM call failure
If the LLM call for suggestion fails (network, rate limit), the optimizer SHALL retry up to 2 times with exponential backoff. After 3 failures, halt with error.

### EC-5: --dry-run mode
If `--dry-run` flag is set, the optimizer SHALL produce the report and suggestion without prompting for approval and without modifying the frontmatter.

---

## Quality Gate Criteria

| Gate | Threshold | Evidence |
|------|-----------|----------|
| optimize command functional | E2E test on 5 skills | 5/5 PASS |
| FP threshold detection | seeded high-FP skill | flagged for tightening |
| FN threshold detection | seeded high-FN skill | flagged for broadening |
| Auto-apply block | no-approval scenario | 100% block |
| Cross-effect protection | full catalog regression | <= 5% variation or rollback |
| Convergence detection | repeated suggestion | halt |
| Sample data validation | <5 samples | halt with error |
| Per-skill override | override scenario | applied correctly |
| LLM call retry | network failure simulation | 3 attempts then halt |
| Template-First sync | clean | `make build` diff |
| plan-auditor | PASS | auditor report |

---

## Definition of Done

- [ ] All 10 Given-When-Then scenarios PASS
- [ ] All 5 edge cases (EC-1 to EC-5) documented and handled
- [ ] All 11 quality gate criteria meet threshold
- [ ] `internal/skill/optimizer/{analyzer,suggester}.go` with >= 90% coverage
- [ ] `cmd/moai/skill.go` extended with optimize subcommand
- [ ] `.claude/agents/moai/builder-skill.md` updated with optimization protocol section
- [ ] `.claude/rules/moai/development/skill-description.md` policy document
- [ ] SPEC-SKILL-TEST-001 dependency verified (blockedBy gate passed)
- [ ] Acceptance rate dogfooding: 5 skills tested with >= 60% accept rate
- [ ] FP/FN reduction validated: >= 50% reduction post-apply
- [ ] `make build` regenerates embedded.go cleanly
- [ ] CHANGELOG.md updated under Unreleased
- [ ] CI runners (ubuntu/macos/windows) PASS
- [ ] plan-auditor PASS
- [ ] No batch optimization (single skill per invocation enforced)

End of acceptance.md (SPEC-SKILL-DESC-001).
