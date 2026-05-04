# SPEC-V3R3-RETIRED-DDD-001 Compact (auto-generated)

> Auto-generated extract from spec.md v0.3.0 / plan.md v0.3.0 / acceptance.md v0.3.0 / research.md v0.3.0.
> Includes: REQ list (sequential REQ-RD-001..REQ-RD-012), G/W/T scenarios, files-to-modify (3-category disposition), Out-of-Scope, Open Items.
> Excludes: overview prose, technical approach, research references, annotation history.

## HISTORY

| Version | Date       | Description                                                                                                                                                |
|---------|------------|------------------------------------------------------------------------------------------------------------------------------------------------------------|
| 0.1.0   | 2026-05-04 | Auto-generated initial extract.                                                                                                                            |
| 0.2.0   | 2026-05-04 | Audit iter 1 fixes: REQ 순서 sequential REQ-RD-001..012; file taxonomy Cat A 30 / Cat B 3 / Cat C 2 (research.md §6 ground truth); REQ-RD-NNN namespace 정합. |
| 0.3.0   | 2026-05-04 | Audit iter 2 atomic sync (D-NEW-1 ~ D-NEW-5): version bumped to current; orphan out-of-range REQ-RD reference replaced with sequential REQ-RD-012; canonical file count Cat A 30 consistent across all artifacts. |

---

## Frontmatter (canonical)

```yaml
id: SPEC-V3R3-RETIRED-DDD-001
version: "0.3.0"
status: completed
created_at: 2026-05-04
updated_at: 2026-05-04
author: Goos Kim
priority: Medium
labels: [retire, agent-runtime, ddd, standardization, follow-up]
issue_number: 778
depends_on: [SPEC-V3R3-RETIRED-AGENT-001]
related_specs: [SPEC-V3R2-ORC-001, SPEC-V3R3-RETIRED-AGENT-001]
breaking: false
bc_id: []
lifecycle: spec-anchored
```

---

## REQ List (12 EARS Requirements, sequential REQ-RD-001..REQ-RD-012)

### Ubiquitous (4)

- **REQ-RD-001**: standardize manager-ddd.md as retired stub with 5-field frontmatter (`retired: true`, `retired_replacement: manager-cycle`, `retired_param_hint: "cycle_type=ddd"`, `tools: []`, `skills: []`); body cites both predecessor and current SPEC.
- **REQ-RD-002**: extend `agent_frontmatter_audit_test.go` with manager-ddd assertions (2 subtests + 1 new top-level function + 1 helper).
- **REQ-RD-003**: `make build` includes standardized retired stub; existing generic test loops automatically cover manager-ddd.
- **REQ-RD-004**: predecessor `agentStartHandler` (no modification) blocks manager-ddd spawn with `block` decision + replacement message + exit 2.

### Event-Driven (3)

- **REQ-RD-005**: SubagentStart hook for `agentName == "manager-ddd"` retirement → block decision JSON + ≤500ms response (predecessor REQ-RA-012 budget).
- **REQ-RD-006**: factory.go `case "ddd":` preserved + switch-level @MX:NOTE expanded to cite SPEC-V3R3-RETIRED-DDD-001.
- **REQ-RD-007**: `go test ./internal/template/...` validates manager-ddd retirement; emits `RETIREMENT_INCOMPLETE_manager-ddd` or `ORPHANED_MANAGER_DDD_REFERENCE` sentinels.

### State-Driven (3)

- **REQ-RD-008**: retired stub body has 5 H2 sections (mirror manager-tdd) describing reason, replacement, migration, active agent pointer.
- **REQ-RD-009**: ≤500ms guard performance preserved (predecessor REQ-RA-012; observed 0.056ms).
- **REQ-RD-010**: 3-category disposition taxonomy: **Cat A SUBSTITUTE-TO-CYCLE (30 files)** + **Cat B KEEP-AS-IS (3 files)** + **Cat C UPDATE-WITH-ANNOTATION (2 files)**. `TestNoOrphanedManagerDDDReference.checkFiles` slice = 30 Cat A files only.

### Optional (1)

- **REQ-RD-011**: optional `moai agents list --retired` extension — DEFERRED to follow-up SPEC.

### Unwanted Behavior (1)

- **REQ-RD-012 (composite)**: CI fails with `RETIREMENT_INCOMPLETE_manager-ddd` or `ORPHANED_MANAGER_DDD_REFERENCE` if any retirement criterion missing. Predecessor agent infrastructure provides infrastructure (agentStartHandler is generic).

---

## Acceptance Criteria (Given/When/Then summary)

### AC-RD-01 (Positive): Retirement frontmatter conformance
- Given: manager-ddd.md after M2 with 5-field frontmatter
- When: `go test ./internal/template/ -run "TestAgentFrontmatterAudit|TestRetirementCompletenessAssertion" -v`
- Then: ALL retirement-related subtests PASS; body has 5 H2 sections; cites both SPECs in `## Why This Change`; file size between **1300–1700 bytes** (audit D7 fix tolerance band)

### AC-RD-02 (Positive + Edge): SubagentStart runtime guard regression check
- Given: manager-ddd.md retirement frontmatter deployed; predecessor agentStartHandler unchanged
- When: SubagentStart hook fires for `manager-ddd`
- Then: handler returns `block` decision + replacement message + exit 2 within ≤500ms

### AC-RD-03 (Boundary): factory.go case "ddd" backward compat
- Given: legacy user project with active manager-ddd hook config
- When: `factory.go.CreateHandler("ddd-pre-transformation")` invoked
- Then: routes to `NewDDDHandler` (preserved); switch-level @MX:NOTE cites SPEC-V3R3-RETIRED-DDD-001 backward compat rationale

### AC-RD-04 (Negative): CI assertion failure semantics
- Given: incomplete retirement (5 violation branches: missing field / legacy status / malformed tools / orphan substring in Cat A file / missing replacement)
- When: `go test ./internal/template/ -run "TestAgentFrontmatterAudit|TestRetirementCompletenessAssertion|TestNoOrphanedManagerDDDReference" -v`
- Then: appropriate sentinel emitted citing **REQ-RD-NNN** (NOT predecessor REQ-RA-NNN) per audit D3 fix; CI fails; PR merge blocked; predecessor manager-tdd tests continue PASS independently

---

## Files to Modify (Authoritative — research.md §6.5 single source of truth)

### File Disposition Summary

| Category | Files | Description |
|----------|-------|-------------|
| **Cat A — SUBSTITUTE-TO-CYCLE** | **30** | `manager-ddd` substring replaced with `manager-cycle`; file modified |
| **Cat B — KEEP-AS-IS** | **3** | Intentional retirement-related references; preserved (B1 manager-ddd.md frontmatter `name:` is preserved while body is rewritten) |
| **Cat C — UPDATE-WITH-ANNOTATION** | **2** | File modified (add @MX:NOTE) but `manager-ddd` substring (where present) preserved |
| **Total grep-detected files containing `manager-ddd`** | **34** | (= 30 Cat A + 3 Cat B + 1 Cat C-with-substring agent-hooks.md; Cat C2 handle-agent-hook.sh.tmpl has 0 manager-ddd substrings) |

### Cat A — SUBSTITUTE-TO-CYCLE (30 files)

**Cat A1 — Rule files (3 files)**:
1. `rules/moai/development/agent-authoring.md` line 104 — REMOVE list entry
2. `rules/moai/workflow/spec-workflow.md` line 219 — substitute
3. `rules/moai/workflow/worktree-integration.md` line 135 — substitute

**Cat A2 — Agent definition files (11 files)**:
1. `agents/moai/manager-strategy.md` line 136
2. `agents/moai/manager-quality.md` lines 64, 107, 113
3. `agents/moai/manager-spec.md` line 58
4. `agents/moai/expert-backend.md` lines 62, 119
5. `agents/moai/expert-frontend.md` line 119
6. `agents/moai/expert-testing.md` lines 55, 59, 98
7. `agents/moai/expert-debug.md` lines 59, 90
8. `agents/moai/expert-devops.md` lines 59, 114
9. `agents/moai/expert-mobile.md` line 105
10. `agents/moai/expert-refactoring.md` line 54
11. `agents/moai/evaluator-active.md` line 94

**Cat A3 — Output-style files (1 file)**:
1. `output-styles/moai/moai.md` line 127 (iter 2 newly discovered)

**Cat A4 — Skill files (15 files)**:
1. `skills/moai/SKILL.md` lines 117, 219
2. `skills/moai/references/mx-tag.md`
3. `skills/moai/references/reference.md`
4. `skills/moai/workflows/moai.md` lines 78, 148, 239
5. `skills/moai/workflows/run.md` lines 5, 24, 540, 599, 603, 619, 983
6. `skills/moai-foundation-cc/SKILL.md`
7. `skills/moai-foundation-core/SKILL.md` line 32
8. `skills/moai-foundation-quality/SKILL.md` line 31
9. `skills/moai-meta-harness/SKILL.md` lines 155, 215
10. `skills/moai-workflow-ddd/SKILL.md` line 20 (`agent: "manager-ddd"` → `agent: "manager-cycle"`)
11. `skills/moai-workflow-loop/SKILL.md` line 153
12. `skills/moai-workflow-spec/SKILL.md` lines 223, 356
13. `skills/moai-workflow-spec/references/reference.md` lines 19, 70, 272, 533
14. `skills/moai-workflow-spec/references/examples.md` lines 17, 334
15. `skills/moai-workflow-testing/SKILL.md` line 20 (`agent: "manager-ddd"` → `agent: "manager-cycle"`)

### Cat B — KEEP-AS-IS (3 files)

- B1: `agents/moai/manager-ddd.md` line 2 frontmatter `name:` (M2 rewrites body but preserves `name:` field)
- B2: `agents/moai/manager-cycle.md` lines 61, 65, 70 (intended migration table)
- B3: `agents/moai/manager-tdd.md` body line 31 (consolidation cross-reference)

### Cat C — UPDATE-WITH-ANNOTATION (2 files)

- C1: `rules/moai/core/agent-hooks.md` lines 48, 79 (preserved + ADD @MX:NOTE)
- C2: `internal/template/templates/.claude/hooks/moai/handle-agent-hook.sh.tmpl` (no manager-ddd substring; ADD @MX:NOTE only)

### Other Files Modified (Test infra + Hook factory + CHANGELOG)

- MAIN: retired stub body rewrite — `internal/template/templates/.claude/agents/moai/manager-ddd.md` (REWRITE 7628→1500 bytes target; size tolerance 1300–1700 bytes per AC-RD-01; mirror manager-tdd; intersects with Cat B1 frontmatter `name:` preservation)
- MAIN: audit test extension — `internal/template/agent_frontmatter_audit_test.go` (EXTEND +118 lines: 2 subtests + 1 top-level function + 1 helper)
- MAIN: factory @MX:NOTE expansion — `internal/hook/agents/factory.go` (MODIFY +5 lines comment-only)
- CHANGELOG — `CHANGELOG.md` (APPEND Unreleased entry, dual-language, includes `moai update` user action)

---

## Out of Scope (Exclusions)

1. mo.ai.kr direct modification (`moai update` is canonical sync; CHANGELOG calls out user action)
2. `agentStartHandler` modification (predecessor generic implementation already covers manager-ddd)
3. `manager-cycle.md` body changes (already names manager-ddd as deprecated)
4. `factory.go case "ddd":` removal (research.md §4.2 D1 KEEP backward compat)
5. `ddd_handler.go` deletion (symmetric with case keep)
6. New retirement workflow / governance SPEC
7. agency/* retired agents (SPEC-AGENCY-ABSORB-001 영역)
8. `docs-site/` 4-locale sync (CLAUDE.local.md §17, v3.0 release window)
9. Claude Code Agent runtime changes (external Anthropic dependency)
10. 5-layer defect chain layer 5 (`stream_idle_partial`, `feedback_large_spec_wave_split.md` 영역)
11. `text/template` migration scope expansion (predecessor REQ-RA-006 scope)
12. Other active agent retirements (single-case follow-up only)
13. `agent-hooks.md` table row removal (legacy backward compat doc, KEEP via Cat C1)
14. `handle-agent-hook.sh.tmpl ddd-*` references removal (legacy doc, KEEP via Cat C2)

---

## Open Items / Deferred

| # | Item | Reason | Disposition |
|---|------|--------|-------------|
| 1 | REQ-RD-011: `moai agents list --retired` subcommand | Predecessor REQ-RA-014 already deferred; CLI surface SPEC scope | DEFERRED to follow-up CLI SPEC |
| 2 | factory.go `case "ddd"` removal | Backward compat preservation; telemetry-driven cleanup | DEFERRED to telemetry-after cleanup SPEC (≥6 month observation) |
| 3 | docs-site 4-locale (ko/en/zh/ja) reference sync | CLAUDE.local.md §17 영역, large-scale translation | DEFERRED to v3.0 release window SPEC |
| 4 | mo.ai.kr active monitoring / alert system | Out-of-scope; user-managed | DEFERRED indefinitely |

---

## Decision Highlights

| # | Decision | Source |
|---|----------|--------|
| D1 | factory.go `case "ddd"` KEEP with @MX:NOTE | research.md §4.2 Option A — symmetric with predecessor `case "tdd"` |
| D2 | NEW subtests + NEW top-level function in audit test | research.md §5.2 — clean diff, targeted sentinels |
| D3 | manager-cycle.md migration table KEEP (Cat B2) | research.md §6.3 — intentional retirement notice |
| D4 | handle-agent-hook.sh.tmpl `ddd-*` references KEEP (Cat C2) | research.md §6.4 — symmetric with `tdd-*` |
| D5 | docs-site 4-locale sync DEFER | research.md §7.3 |
| D6 | mo.ai.kr direct modification OUT OF SCOPE | research.md §7.1 + spec.md §2.2 |
| D7 | moai-workflow-ddd/SKILL.md `agent:` substitution | research.md §6.2 A4 — skill metadata field, no functional change |
| D8 | manager-ddd.md body EXACT mirror manager-tdd | research.md §3.3 — consistency reduces maintenance |
| D9 | New REQ namespace `REQ-RD-NNN`, **sequential 001..012 with NO gaps** (iter 2 audit D1 fix) | research.md §9 D9 |
| D10 | TDD methodology per quality.yaml | M1 RED → M2-M4 GREEN → M5 REFACTOR |
| D11 | output-styles/moai/moai.md added to Cat A3 (iter 2 newly discovered) | research.md §9 D11 |
| D12 | 3-category disposition taxonomy (Cat A SUBSTITUTE / Cat B KEEP / Cat C UPDATE-WITH-ANNOTATION) | research.md §9 D12 — replaces iter 1's binary KEEP/SUBSTITUTE |

---

End of spec-compact.md.
