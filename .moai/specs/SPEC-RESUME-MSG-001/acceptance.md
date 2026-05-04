---
id: SPEC-RESUME-MSG-001
acceptance_version: "0.1.0"
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
---

# Acceptance Criteria — SPEC-RESUME-MSG-001

## Given-When-Then Scenarios

### Scenario 1: 정책 문서 보강

**Given** the SPEC-RESUME-MSG-001 implementation completes
**And** Template-First sync runs

**When** the user inspects `context-window-management.md`

**Then** the file SHALL exist with 5 new sections added (Applies To, Format Extensions, State JSON Schema, Auto-Save Trigger, AskUserQuestion Prohibition)
**And** existing 75% / 90% threshold definitions SHALL be preserved
**And** the corresponding template file SHALL be identical (same hash)

---

### Scenario 2: 6+ Workflow 적용

**Given** the policy document exists

**When** the user reads "Applies To" section

**Then** the section SHALL list at least 6 workflows: /moai plan, /moai run, /moai sync, /moai loop, /moai fix, /moai design (GAN loop), Wave-style multi-SPEC delegation
**And** each workflow SHALL have at least 1 cross-reference back from its skill body

---

### Scenario 3: 5 Format Extensions

**Given** the policy document exists

**When** the user reads §"Workflow-Specific Resume Format Extensions"

**Then** the section SHALL contain exactly 5 extensions: Ext1 (loop), Ext2 (design GAN), Ext3 (sync multi-PR), Ext4 (run agent chain), Ext5 (Wave)
**And** each extension SHALL preserve the standard 7-line format and add 1-2 lines of extension_data
**And** each extension SHALL include a complete example resume message

---

### Scenario 4: `.moai/state/<session-id>.json` Schema

**Given** the policy document exists

**When** the user reads §".moai/state/ Recommended Schema"

**Then** the schema SHALL include fields: session_id, workflow, started_at, last_saved_at, context_usage_percent, active_spec, applied_lessons, current_step, next_action, extension_data
**And** the schema SHALL be marked as RECOMMENDED (not hard rule)
**And** `.moai/state/.gitkeep` SHALL exist

---

### Scenario 5: 75% Threshold → State Save + Resume Message

**Given** an active workflow at 76% context usage

**When** the orchestrator self-monitoring detects threshold crossing

**Then** the orchestrator SHALL persist state to `.moai/state/<session-id>.json`
**And** the orchestrator SHALL emit a structured resume message via natural-language status (NOT AskUserQuestion)
**And** the resume message SHALL contain all 7 standard lines
**And** the resume message SHALL include applicable extension lines for the active workflow

---

### Scenario 6: 90% Threshold → /clear Recommendation

**Given** an active workflow at 91% context usage

**When** the orchestrator self-monitoring detects 90% crossing

**Then** the orchestrator SHALL announce `/clear` recommendation as natural-language status
**And** the orchestrator SHALL NOT use AskUserQuestion for this announcement
**And** the resume message SHALL be available BEFORE the announcement (so user can paste after /clear)

---

### Scenario 7: User Paste-and-Resume

**Given** the user runs `/clear` after receiving a resume message
**And** the user pastes the resume message in a new session

**When** the orchestrator processes the input

**Then** the orchestrator SHALL recognize the `ultrathink. <workflow_name>` prefix
**And** the orchestrator SHALL load referenced state from `.moai/state/<session-id>.json` if present
**And** the orchestrator SHALL continue the workflow at the indicated next step

---

### Scenario 8: Cross-References (6+)

**Given** the SPEC implementation completes

**When** the user inspects skill files

**Then** at minimum 6 files SHALL contain cross-reference to `context-window-management.md`:
  - `.claude/skills/moai/workflows/plan.md`
  - `.claude/skills/moai/workflows/run.md`
  - `.claude/skills/moai/workflows/sync.md`
  - `.claude/skills/moai/workflows/loop.md`
  - `.claude/skills/moai/workflows/fix.md`
  - `.claude/skills/moai-team-design/SKILL.md`
**And** each cross-reference SHALL be in a relevant section (e.g., "Resume protocol" or "Context Management")

---

## Edge Cases

### EC-1: `.moai/state/<session-id>.json` Missing or Unreadable
The workflow SHALL emit a degraded resume message (best-effort) with explicit "state unavailable" note. The user can manually reconstruct context.

### EC-2: Wave-Style Without Active SPEC
For Wave-style multi-SPEC delegation, the resume message SHALL include Wave N/M and completed SPEC list, even if individual SPEC progress.md is absent.

### EC-3: GAN Loop Iteration N/5 Without Score
If design GAN loop fails before scoring (e.g., builder error), the Ext2 resume message SHALL note "score: pending" and last error.

### EC-4: Self-Monitoring Inaccurate (Under-Estimate)
If orchestrator under-estimates and 75% is reached without state save, the user CAN manually trigger save by writing `progress.md`. This is acceptable per under-estimate-preferred policy.

### EC-5: AskUserQuestion 시도 → 위반
If any orchestrator implementation tries to use AskUserQuestion for /clear trigger, the policy SHALL flag as anti-pattern. Status announcement is the only allowed channel.

---

## Quality Gate Criteria

| Gate | Threshold | Evidence |
|------|-----------|----------|
| Policy document | both local + template | file existence |
| Existing 75%/90% preserved | grep verification | preserved |
| 5 new sections | grep | all present |
| 6+ workflow listed | grep | >= 6 |
| 5 format extensions | grep | exactly 5 |
| State JSON schema | 10 fields documented | grep |
| `.moai/state/.gitkeep` | file existence | EXISTS |
| AskUserQuestion prohibition | grep | EXISTS |
| Cross-references | >= 6 | grep count |
| Verbatim Anthropic quotes | >= 3 | grep blockquote |
| Template-First sync | clean | `make build` diff |
| plan-auditor | PASS | auditor report |

---

## Definition of Done

- [ ] All 8 Given-When-Then scenarios PASS
- [ ] All 5 edge cases (EC-1 to EC-5) documented and handled
- [ ] All 12 quality gate criteria meet threshold
- [ ] Policy document boosted at `.claude/rules/moai/workflow/context-window-management.md` and template
- [ ] 5 new sections added (Applies To, Format Extensions, State JSON Schema, Auto-Save Trigger, AskUserQuestion Prohibition)
- [ ] Existing 75% / 90% threshold definitions preserved
- [ ] 6+ workflows listed (plan, run, sync, loop, fix, design, Wave-style)
- [ ] 5 format extensions documented (loop, design GAN, sync multi-PR, run agent chain, Wave)
- [ ] `.moai/state/<session-id>.json` recommended schema (10 fields)
- [ ] `.moai/state/.gitkeep` present
- [ ] Self-monitoring detection heuristic documented (under-estimate preferred)
- [ ] AskUserQuestion prohibition (status announcement only)
- [ ] 6+ cross-references in workflow skills + design skill
- [ ] 3+ verbatim Anthropic citations
- [ ] `make build` regenerates embedded.go cleanly
- [ ] CHANGELOG.md updated under Unreleased
- [ ] No Go code change (documentation + cross-ref only verified by `git diff`)
- [ ] plan-auditor PASS
- [ ] dogfooding: 1 long-running workflow (e.g., /moai loop) tested with resume cycle

End of acceptance.md (SPEC-RESUME-MSG-001).
