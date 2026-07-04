---
id: SPEC-DISCOVERY-UNKNOWNS-001
title: "Unknowns framework Tier-1 for Context-First Discovery — Blind Spot Pass + decision-reversibility ordering + 4-quadrant lens"
version: "0.1.0"
status: draft
created: 2026-07-05
updated: 2026-07-05
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/rules/moai/core"
lifecycle: spec-anchored
tier: M
tags: "discovery, unknowns, blind-spot, gears, askuser, planning, context-first, doc"
---

# SPEC-DISCOVERY-UNKNOWNS-001 — Acceptance Criteria

## §D. AC Matrix

Each AC is independently testable. Commands are run from the project root (`/Users/goos/MoAI/moai-adk-go`) after run-phase edits + `make build`. `<local>` / `<mirror>` shorthand: the local file and its `internal/template/templates/` mirror per spec.md §A.4.

| AC | REQ | Verification command (independently testable) | PASS condition |
|----|-----|-----------------------------------------------|----------------|
| AC-DU-001 | REQ-DU-001 | `grep -cE '^#{2,3} .*Blind Spot Pass' .claude/rules/moai/core/askuser-protocol.md` | ≥ 1 (a named H2/H3 Blind Spot Pass subsection exists) |
| AC-DU-002 | REQ-DU-003 | `grep -A40 'Blind Spot Pass' .claude/rules/moai/core/askuser-protocol.md \| grep -E 'Agent\(Explore\)' && grep -A40 'Blind Spot Pass' .claude/rules/moai/core/askuser-protocol.md \| grep -iE 'read-only'` | both match (mechanism = Explore read-only scan) |
| AC-DU-003 | REQ-DU-004 | `grep -A40 'Blind Spot Pass' .claude/rules/moai/core/askuser-protocol.md \| grep -iE 'optional\|not a mandatory gate'` | ≥ 1 match (optionality stated) |
| AC-DU-004 | REQ-DU-005 | `grep -A40 'Blind Spot Pass' .claude/rules/moai/core/askuser-protocol.md \| grep -iE 'AskUserQuestion' && grep -A40 'Blind Spot Pass' .claude/rules/moai/core/askuser-protocol.md \| grep -iE 'not prompt\|never prompt\|does not prompt'` | both match (boundary reaffirmed: surface via AskUserQuestion; Explore does not prompt) |
| AC-DU-005 | REQ-DU-006 | `grep -c 'Blind Spot Pass' .claude/rules/moai/workflow/spec-workflow.md` and `grep -c 'Blind Spot Pass' CLAUDE.md` | both ≥ 1 (wired from Plan Phase + §7 Rule 5) |
| AC-DU-006 | REQ-DU-007 | `grep -iE 'most likely to change\|decision-reversibility' .claude/agents/moai/manager-spec.md && grep -iE 'mechanical\|refactor' .claude/agents/moai/manager-spec.md \| grep -iE 'bottom\|defer\|last'` | both match (lead-with-change-likelihood + defer-mechanical) |
| AC-DU-007 | REQ-DU-008 | `grep -A2 'Rule 1 — Approach-First' CLAUDE.md \| grep -iE 'most likely to change\|change[ -]likelihood\|first'` | ≥ 1 match (Approach-First leads with high-change decisions) |
| AC-DU-008 | REQ-DU-009, REQ-DU-013 | `find internal pkg cmd -newer .moai/specs/SPEC-DISCOVERY-UNKNOWNS-001/spec.md -name '*.go' -not -path '*/testdata/*'` scoped to this SPEC's change set | empty (no new/modified Go from this SPEC — ordering rule only, no machinery) |
| AC-DU-009 | REQ-DU-010 | `grep -iE 'Known-Knowns' .claude/rules/moai/core/askuser-protocol.md && grep -iE 'Known-Unknowns' … && grep -iE 'Unknown-Knowns' … && grep -iE 'Unknown-Unknowns' …` (all four terms) | all four quadrant terms present in § Ambiguity Triggers |
| AC-DU-010 | REQ-DU-011 | `grep -iE 'Unknown-Unknowns' .claude/rules/moai/core/askuser-protocol.md \| grep -iE 'Blind Spot Pass'` OR proximity grep (`grep -A6 'Unknown-Unknowns' … \| grep 'Blind Spot Pass'`) | ≥ 1 (suspected Unknown-Unknowns routes to Blind Spot Pass) |
| AC-DU-011 | REQ-DU-012 | `grep -A3 'Rule 5 — Context-First' CLAUDE.md \| grep -iE 'Known-Unknowns\|4-quadrant\|unknown-unknown'` AND the §7 Rule 5 net addition for T3 is ≤ 1 line + pointer (diet check via diff review) | lens framing + SSOT pointer present; ≤ 1 line + pointer added |
| AC-DU-012 | REQ-DU-014 | For each of the 4 files: the M1-M3 phrase present in BOTH `<local>` AND `<mirror>` — e.g. `grep -c 'Blind Spot Pass' internal/template/templates/.claude/rules/moai/core/askuser-protocol.md` ≥ 1, and same for manager-spec/CLAUDE.md/spec-workflow mirrors | every new phrase present in local AND mirror (parity) |
| AC-DU-013 | REQ-DU-014 | `make build` | exit 0 (embedded template FS regenerated) |
| AC-DU-014 | REQ-DU-015 | `grep -rn 'SPEC-DISCOVERY-UNKNOWNS' internal/template/templates/` and `grep -rnE '2026-07-05\|Finding [0-9]\|Audit [0-9]' internal/template/templates/.claude/rules/moai/core/askuser-protocol.md internal/template/templates/.claude/agents/moai/manager-spec.md` | both return zero matches introduced by this SPEC (§25 neutrality) |
| AC-DU-015 | REQ-DU-013 | `git status --porcelain` filtered to this SPEC's change set; assert no `.go` path | no `.go` file modified/added by this SPEC (doc/rule/agent-body only) |

## §D.1 Given-When-Then scenarios

### Scenario 1 — unfamiliar-domain plan entry triggers an optional Blind Spot Pass (T1)

- **Given** the user requests plan-phase work in a subsystem they have not worked in before, AND the orchestrator suspects unknown-unknowns (unfamiliar design/library territory),
- **When** the orchestrator reaches the plan-phase entry,
- **Then** the orchestrator runs a Blind Spot Pass — spawns `Agent(Explore)` in read-only mode to scan the relevant domain, then surfaces the user's likely unknown-unknowns through an `AskUserQuestion` round — before authoring the SPEC,
- **And** the surfacing goes only through the orchestrator's AskUserQuestion channel (Explore never prompts the user directly).

### Scenario 2 — plan.md leads with the decisions most likely to change (T2)

- **Given** a plan.md is authored for a SPEC that changes a data model, adds a new type interface, and also performs some mechanical refactors,
- **When** manager-spec orders the plan.md milestones/sections,
- **Then** the data-model, type-interface, and user-facing/UX-flow decisions appear before the mechanical refactoring steps,
- **And** the mechanical/refactoring steps are deferred to the bottom so human review focuses on the high-change-likelihood decisions.

### Scenario 3 — the 4-quadrant lens routes suspected Unknown-Unknowns to a Blind Spot Pass (T3 → T1)

- **Given** the orchestrator applies the § Ambiguity Triggers with the Known-Knowns / Known-Unknowns / Unknown-Knowns / Unknown-Unknowns lens and suspects Unknown-Unknowns,
- **When** it classifies the ambiguity,
- **Then** the lens directs it to initiate a Blind Spot Pass (per REQ-DU-002) rather than proceeding directly to plan.

## §D.2 Edge cases

- **Familiar domain / no suspected unknown-unknowns** → the Blind Spot Pass is NOT run; optionality is preserved and no forced overhead is incurred (guards REQ-DU-004). Verifiable: the subsection text explicitly gates on "suspected unknown-unknowns", not on every plan entry.
- **Template-mirror divergence** → `askuser-protocol.md` and `manager-spec.md` already diverge local-vs-template by ~1 line (§25 SPEC-ref stripping). The paired edit targets the semantically-equivalent anchor in each surface; the new phrase must still appear in BOTH (AC-DU-012).
- **CLAUDE.md active diet** → §7 Rule 5 gains ≤ ~2 lines total (T1 ref + T3 lens framing + pointer); §7 Rule 1 gains 1 line (T2). No section-block expansion (AC-DU-011 diet check).
- **Parallel-session working tree** → ~52 uncommitted files from other SPECs are present; edits touch only the 8 enumerated targets + this SPEC dir; AC-DU-015 asserts no unrelated `.go` change is attributed to this SPEC.

## §D.3 Quality gate criteria / Definition of Done

- [ ] All 15 ACs (AC-DU-001 .. AC-DU-015) PASS with observed command output recorded in the run-phase E1 matrix.
- [ ] `make build` exits 0 (AC-DU-013).
- [ ] Template parity holds for all 4 surfaces — every new phrase present in local AND mirror (AC-DU-012).
- [ ] §25 neutrality: zero internal SPEC ID / date / SHA / audit-citation leakage into template content (AC-DU-014).
- [ ] No Go source, no new agent, no new skill, no new subsystem introduced (AC-DU-008 + AC-DU-015).
- [ ] Blind Spot Pass is optional (AC-DU-003) and preserves the subagent boundary (AC-DU-004).
- [ ] CLAUDE.md diet respected — §7 Rule 5 addition ≤ ~2 lines + pointer (AC-DU-011).
- [ ] `spec-lint` clean on the new SPEC: 12 canonical frontmatter fields present; `### Out of Scope — <topic>` H3 sub-headings with `-` bullets satisfy `OutOfScopeRule`.
- [ ] plan-auditor verdict PASS at the Tier M threshold (0.80).
