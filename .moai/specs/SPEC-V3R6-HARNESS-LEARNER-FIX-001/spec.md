---
id: SPEC-V3R6-HARNESS-LEARNER-FIX-001
title: "moai-harness-learner subagent boundary fix"
version: "0.2.0"
status: implemented
created: 2026-05-22
updated: 2026-05-22
author: GOOS Kim
priority: P0
phase: "v3.0.0"
module: ".claude/skills/moai-harness-learner"
lifecycle: spec-anchored
tier: S
tags: "harness, subagent-boundary, askuser-protocol, v3, hard-violation"
---

# SPEC-V3R6-HARNESS-LEARNER-FIX-001: moai-harness-learner subagent boundary fix

## 1. Problem Statement

`.claude/skills/moai-harness-learner/SKILL.md` (162 LOC, last modified 2026-05-15) violates the canonical subagent boundary HARD rule defined in `.claude/rules/moai/core/askuser-protocol.md § Orchestrator-Subagent Boundary`. The skill currently asserts that it (not the orchestrator) invokes `AskUserQuestion` to surface Tier 4 auto-update proposals to the user. This contradicts CONST-V3R2-026 and CONST-V3R5-001/002/003 (zone-registry.md), which reserve `AskUserQuestion` exclusively for the MoAI orchestrator.

Two concrete defects identified by the `.claude/skills/` audit (2026-05-21, P0-S1):

- **Defect 1 (frontmatter)**: Line 20 lists `AskUserQuestion` in `allowed-tools`. Skills are subagent-execution scope; including `AskUserQuestion` in their allowed-tools declares a capability that subagents must never exercise.
- **Defect 2 (body assertion)**: Lines 82, 46, and 35 of the body assert "This skill calls AskUserQuestion" / "Surface via AskUserQuestion (approve / reject)" / "this skill MUST receive that payload and surface it via AskUserQuestion." These directly contradict the orchestrator-only contract.

These defects create a chicken-and-egg situation: the canonical contract says subagents must not invoke `AskUserQuestion`, but the skill body instructs the skill-resident agent to do exactly that. If the skill is executed faithfully, it produces an `InputValidationError: tool not in schema` at runtime; if the skill is executed pragmatically (ignoring the body), the documented protocol diverges from observed behavior. Either path erodes contract trust.

The original V3R4 design (per `SPEC-V3R4-HARNESS-001 §10 exclusion #10`) preserved the skill body verbatim to defer this fix. V3R6 — the v3.0.0 redesign Wave 1 — is the appropriate cycle to resolve it.

## 2. Goals

- **G-1**: Eliminate the frontmatter declaration of `AskUserQuestion` in `allowed-tools` for `moai-harness-learner`.
- **G-2**: Rewrite the body's AskUserQuestion-handling section so the skill produces a structured payload that the orchestrator consumes, instead of asserting skill-direct invocation.
- **G-3**: Cite `.claude/rules/moai/core/askuser-protocol.md § Orchestrator-Subagent Boundary` as the canonical contract in the rewritten section, removing duplicated/restated rules.
- **G-4**: Preserve all other content in the skill body (4-Tier observation ladder, safety layers L1-L5, downstream contract with `moai-workflow-gan-loop`, etc.) — this is a narrow boundary-compliance fix, not a redesign.

## 3. EARS Requirements

- **REQ-HLF-001 (Ubiquitous)**: The `moai-harness-learner` skill SHALL NOT list `AskUserQuestion` in its `allowed-tools` frontmatter field. The post-fix `allowed-tools` value SHALL be `Bash,Read,Write,Edit`.

- **REQ-HLF-002 (Event-driven)**: When a harness Tier 4 auto-update proposal requires user approval, the skill SHALL produce a structured payload containing the fields `proposal_id`, `target_path`, `field_key`, `current_value`, `new_value`, `observation_count`, `confidence`, and `recommended_action` — for the MoAI orchestrator to consume and surface via `AskUserQuestion`. The skill SHALL NOT invoke `AskUserQuestion` itself.

- **REQ-HLF-003 (Where)**: Where the skill body references the user-approval interaction model, the reference SHALL cite `.claude/rules/moai/core/askuser-protocol.md § Orchestrator-Subagent Boundary` as the canonical contract and SHALL NOT restate the rule.

- **REQ-HLF-004 (Unwanted)**: The skill body SHALL NOT assert "This skill calls AskUserQuestion", "Surface payload via AskUserQuestion", or any equivalent claim of skill-direct user-prompting capability.

- **REQ-HLF-005 (Ubiquitous)**: The skill body SHALL preserve the 4-Tier observation ladder (Observation → Heuristic → Rule → Auto-Update), the L1-L5 safety architecture references, and the downstream `moai-workflow-gan-loop` contract — modifying only the AskUserQuestion-handling section (lines 80-95 of the current body).

## 4. Binary Acceptance Criteria

| AC ID | Verification command | Pass condition |
|-------|---------------------|----------------|
| **AC-HLF-001** | `grep "^allowed-tools" .claude/skills/moai-harness-learner/SKILL.md \| grep -c "AskUserQuestion"` | Returns `0` (frontmatter compliance per REQ-HLF-001) |
| **AC-HLF-002** | `grep -n "askuser-protocol.md" .claude/skills/moai-harness-learner/SKILL.md` | Returns at least 1 match (canonical reference present per REQ-HLF-003) |
| **AC-HLF-003** | `grep -c "This skill.*calls AskUserQuestion\|This skill (not the CLI) calls AskUserQuestion" .claude/skills/moai-harness-learner/SKILL.md` | Returns `0` (anti-pattern claim removed per REQ-HLF-004) |
| **AC-HLF-004** | `grep -c "Surface.*via AskUserQuestion\|Surface payload via .*AskUserQuestion" .claude/skills/moai-harness-learner/SKILL.md` | Returns `0` (anti-pattern step removed per REQ-HLF-004) |
| **AC-HLF-005** | `grep -c "4-Tier\|Tier 4 auto-update\|L1.*L5\|safety architecture" .claude/skills/moai-harness-learner/SKILL.md` | Returns ≥ 4 (preservation per REQ-HLF-005) |
| **AC-HLF-006** | `grep -n "proposal_id" .claude/skills/moai-harness-learner/SKILL.md` | Returns at least 1 match (structured payload field per REQ-HLF-002) |
| **AC-HLF-007** | `cd /Users/goos/MoAI/moai-adk-go && go test ./internal/template/...` | All package tests PASS (no embedded.go regression — the skill file is embedded in moai-adk-go via go:embed) |

## 5. Out of Scope

- **OOS-1**: Renaming `moai-harness-learner` to `my-harness-learner` (deferred to Wave 2 SPEC-V3R6-HARNESS-RENAME-001).
- **OOS-2**: agentskills.io schema overhaul, JSON Schema migration, or skill registry redesign (deferred to Wave 3 SPEC-V3R6-SKILL-SLIM-001, Tier L).
- **OOS-3**: Modifications to the 4-Tier observation ladder, safety layers L1-L5, GAN loop integration, or any non-AskUserQuestion section of the skill body. REQ-HLF-005 explicitly fences the change surface to the AskUserQuestion handling section.
- **OOS-4**: Changes to the CLI side (`moai harness apply`, `moai harness rollback`, `cmd/moai/harness.go`). The CLI's JSON-payload-returning behavior is already correct; this SPEC only fixes the skill body's mischaracterization of that contract.
- **OOS-5**: Other audit findings beyond P0-S1. P0-S2 through P0-Sn from `.claude/skills/` audit 2026-05-21 are tracked separately.

## 6. Risks

- **R-1 (Medium)**: The 4-Tier observation ladder and safety architecture references in the body may be entangled with the AskUserQuestion section via section numbering. Mitigation: Read full body via Read tool before Edit; identify the exact line range (current estimate: 80-95); apply Edit with sufficient `old_string` context to avoid cross-section drift. Verify via AC-HLF-005 (preservation grep count).

- **R-2 (Low)**: Embedded template regeneration via `make build` may produce a different `embedded.go` hash. Mitigation: Run `go test ./internal/template/...` (AC-HLF-007) after edit to confirm no regression. The change is content-only; embed metadata (paths, names) unchanged.

- **R-3 (Low)**: Downstream skills (`moai-workflow-gan-loop`, `evaluator-active`) may have implicit dependencies on the current AskUserQuestion-handling phrasing. Mitigation: Grep for cross-references via `grep -rn "moai-harness-learner.*AskUserQuestion" .claude/` during run-phase pre-flight. If matches found, fold into the same fix or escalate as a blocker.

## 7. REQ ↔ AC Traceability

| Requirement | Acceptance Criteria | Coverage |
|-------------|--------------------|----------|
| REQ-HLF-001 | AC-HLF-001 | 100% |
| REQ-HLF-002 | AC-HLF-006 | 100% |
| REQ-HLF-003 | AC-HLF-002 | 100% |
| REQ-HLF-004 | AC-HLF-003, AC-HLF-004 | 100% |
| REQ-HLF-005 | AC-HLF-005 | 100% |
| (build integrity) | AC-HLF-007 | (cross-cutting) |

**Coverage**: 5/5 requirements verified by binary AC = **100%**. AC-HLF-007 is a cross-cutting build-integrity gate not tied to a specific REQ but required by Section D of the delegation contract (no embedded.go regression).

---

**SPEC tier**: S (LEAN minimal — 2 artifacts: spec.md + plan.md; acceptance inline in §4).
**Workflow**: Late-Branch per SPEC-V3R5-LATE-BRANCH-001 REQ-LB-005 — commits stay on main until PR creation.
**Audit reference**: `.claude/skills/` audit 2026-05-21, P0-S1 (subagent boundary violation in moai-harness-learner).
**Canonical contract**: `.claude/rules/moai/core/askuser-protocol.md § Orchestrator-Subagent Boundary`; CONST-V3R2-026, CONST-V3R5-001/002/003.
