---
id: SPEC-AGENT-ARCH-V2-001
title: "MoAI Agent Architecture v2 — super-advisor + manager-design + No-Haiku 3-Tier Token Policy"
version: "0.2.0"
status: draft
created: 2026-07-09
updated: 2026-07-09
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/agents/moai + internal/config"
lifecycle: spec-anchored
era: V3R6
tier: L
tags: "agent-arch, super-advisor, manager-design, no-haiku, 3-tier, claude-design, token-policy"
---

# SPEC-AGENT-ARCH-V2-001 — MoAI Agent Architecture v2

## Epic Context

**Standalone architecture redesign** (NOT a Workflow-Reflex Epic member). This SPEC supersedes two Workflow-Reflex draft SPECs and reconfigures the Epic from 6 active members to 4.

- **Design authority**: `.moai/reports/agent-architecture-redesign-v2-20260709.html` (565 lines, 2026-07-09) — the architecture SSOT. Every requirement in this spec.md traces to a verbatim §NN section of that report; no architecture is re-derived or invented here.
- **v1 → v2 delta (4 changes per §01)**: (1) `advisor` → `super-advisor` rename; (2) `manager-design` new agent for Claude Design bidirectional integration; (3) Haiku全面 배제 — all haiku slots replaced by Sonnet 5 + effort low/medium; (4) 3-tier token policy (`max`/`medium`/`low`) selected at `moai init`.
- **Catalog verdict (§05)**: 9 + 1 (super-advisor + manager-design + 5 retained MoAI workers + 2 auditors + Explore built-in). CLAUDE.md §4 Retained Agents ceiling moves 8 → 10.
- **Tier**: L (per §06 — up-tier rationale: 3-tier Go routing + 2 new agent files + workflow design-phase integration).
- **era**: V3R6 (modern 3-phase close: plan → run → sync).

### Supersede footprint (REQ-AA2-017)

This SPEC absorbs and supersedes two Workflow-Reflex draft SPECs:

| Superseded SPEC | Workflow-Reflex slot | Disposition | Rationale |
|-----------------|----------------------|-------------|-----------|
| `SPEC-ADVISOR-RUNG-001` | 4 of 6 | **ABSTRACTION-LIFT** — per-spawn `Agent(general-purpose, model: opus)` pattern is promoted to a dedicated catalog agent file (`super-advisor.md`); the rung trigger (N=2 same-diagnostic failures) becomes E1 of the v2 escalation doctrine; REQ-ADV-004 (GLM carve-out) and REQ-ADV-005 (CG leader-review-as-advisor) concerns are captured natively by super-advisor's on-demand/all-phases scope |
| `SPEC-MODEL-ROUTING-WIRE-001` | 2 of 6 | **POLICY-FLIP** — the WIRE-001 `haiku-inherit` tag direction is reversed: v2 mandates Haiku exclusion (haiku → sonnet low/medium); the `moai route` CLI and the pre-spawn consultation instruction fold into M3's `RouteModelFor(specTier, phase, perfTier)` 3-arg extension |

**Workflow-Reflex Epic reconfiguration**: pre-v2 the Epic had 6 active members; post-v2 the Epic has **4 active members** — SPEC-HARNESS-RATCHET-REWIRE-001 (1 of 6), SPEC-LOOP-VERDICT-CONTRACT-001 (3 of 6), SPEC-CADENCE-BRIDGE-001 (5 of 6), SPEC-OBSERVE-HYGIENE-001 (6 of 6) remain. The two superseded members (2 of 6, 4 of 6) exit via `* → superseded` (frontmatter flip applied to both targets; their `superseded_by: SPEC-AGENT-ARCH-V2-001` field recorded).

---

## User Story

**As the** MoAI maintainer consolidating the agent catalog around a single worker-class model (Sonnet 5) with on-demand Opus reasoning,
**I want** (a) a dedicated super-advisor agent file for high-reasoning consultation across all phases, (b) a manager-design agent that bidirectionally integrates with Claude Design via the DesignSync MCP tool, (c) a No-Haiku 3-tier token policy (`max`/`medium`/`low`) selected at `moai init` and mechanically routed through an extended `RouteModelFor(specTier, phase, perfTier)` accessor, and (d) the two superseded Workflow-Reflex draft SPECs formally retired,
**so that** the cost-asymmetry between cheap execution and expensive reasoning is exploited systematically (Opus deployed only at reasoning-intensive points; Haiku's quality variance and `[1m]`-entitlement edge cases eliminated), design-to-code handoff becomes a first-class workflow phase rather than a role_profile side-channel, and the catalog stops carrying two draft SPECs whose concerns v2 subsumes.

---

## Problem — Measurable Gap Definition (vci §2 attribution)

All gaps measured 2026-07-09 by this agent via Read/Bash against the live tree. The design SSOT (`.moai/reports/agent-architecture-redesign-v2-20260709.html`) is the architecture authority; gaps below cite live-tree evidence of the v1 → v2 delta's absence.

### GAP-1 — No super-advisor agent file exists (§01 change ①, §05)

- **Measured source**: `ls .claude/agents/moai/` and `ls internal/template/templates/.claude/agents/moai/` — no `super-advisor.md` (nor legacy `advisor.md`) in either tree; `grep -rn "super-advisor" .claude/ CLAUDE.md internal/template/templates/` → 0 matches; CLAUDE.md §4 Retained Agents table (lines 95-106) lists 8 agents, none named `super-advisor`.
- **Observed pattern**: the v1 advisor concept lived only as SPEC-ADVISOR-RUNG-001 draft prose (per-spawn `Agent(general-purpose, model: opus)`); no dedicated catalog entry. v2 promotes this to a named agent with Opus injection (max/medium tiers) + effort xhigh FIXED (frontmatter) + read-only + on-demand across all phases.

### GAP-2 — No manager-design agent file exists (§01 change ②, §04)

- **Measured source**: `ls .claude/agents/moai/` — no `manager-design.md`; `grep -rn "manager-design\|/design-sync\|DesignSync" .claude/agents/ .claude/skills/ .claude/rules/` → 0 matches in agent bodies; workflow.yaml `role_profiles.designer` (lines 146-153 area) is the only design-surface primitive, and it is a team-mode role_profile (not a catalog agent); `.mcp.json` does NOT register a `DesignSync` server (verified 2026-07-09 — Gap, see research.md §H).
- **Observed pattern**: Claude Design integration has no first-class agent owner; the D1-D5 pipeline (§04) and the H1-H9 handoff contract (§04 table) have no agent body to embed into; `spec-workflow.md` has no `plan → design → run` conditional route for UI-surfaced SPECs.

### GAP-3 — Haiku references persist across the catalog and config (§01 change ③, §05)

- **Measured source**: `grep -rn "haiku" .claude/agents/moai/ .moai/config/sections/llm.yaml .moai/config/sections/workflow.yaml` → matches in: `llm.yaml:9` (`low: haiku` in `claude_models`); `workflow.yaml:154` (`read-only-extract: { model: haiku, ... }` in `workflow_agents`); `workflow.yaml:174,175,179` (`S-sync`, `S-mx`, `M-mx` matrix entries in `model_routing`); `model-policy.md` § Inherit-by-Default ("except manager-docs and manager-git which use model: haiku"); `team-protocol.md` role matrix (`researcher | plan | haiku`); `model_routing.go:31` (`"haiku": true` in `validRoutingModels` closed set).
- **Observed pattern**: Haiku is wired into 5+ surfaces. v2 mandates Haiku = 0 across the catalog (agent frontmatter, routing matrices, role profiles, workflow_agents, claude_models); all haiku slots replaced by Sonnet 5 + effort low/medium.

### GAP-4 — Single-tier model_routing matrix; no per-tier routing profiles (§01 change ④, §02-D, §2-E)

- **Measured source**: `internal/config/model_routing.go:89` — `func (c *Config) RouteModelFor(tier, phase string) (ModelRoutingEntry, error)` (current 2-arg signature); `workflow.yaml:171-183` — single `model_routing:` block (12 entries, S/M/L × plan/run/sync/mx); `grep -n "model_routing_profiles\|performance_tier" internal/config/*.go` → only `llm.yaml:5` (`performance_tier: ""` — field exists but unused); `llm.yaml:6-9` — `claude_models: {high: opus, medium: sonnet, low: haiku}` (3-tier names exist but low=haiku, contradicting v2 No-Haiku).
- **Observed pattern**: the current accessor takes `(tier, phase)` and returns a single `{model, effort}` entry. v2 extends to `RouteModelFor(specTier, phase, perfTier)` returning a per-performance-tier entry, backed by 3 matrices (`model_routing_profiles.{max,medium,low}` — §2-D tables).

### GAP-5 — Two superseded Workflow-Reflex SPECs still carry `status: draft` (REQ-AA2-017)

- **Measured source**: `.moai/specs/SPEC-ADVISOR-RUNG-001/spec.md:5` → `status: draft`; `.moai/specs/SPEC-MODEL-ROUTING-WIRE-001/spec.md:5` → `status: draft`; both lack a `superseded_by` frontmatter field.
- **Observed pattern**: v2 subsumes both SPECs' concerns (super-advisor absorbs ADVISOR-RUNG; No-Haiku 3-tier flips MODEL-ROUTING-WIRE's haiku-inherit direction and absorbs its `moai route` + pre-spawn consultation). Both targets must transition `draft → superseded` with `superseded_by: SPEC-AGENT-ARCH-V2-001` per the Status Transition Ownership Matrix.

### Aggregate defect claim

**The v2 architecture (4 changes) has zero footprint in the live tree, and two draft SPECs whose concerns v2 subsumes remain active.** This SPEC lands the v2 architecture across 4 milestones (M1 super-advisor, M2 manager-design, M3 No-Haiku 3-tier Go code, M4 doctrine refresh), retires the 2 superseded SPECs, and leaves the Workflow-Reflex Epic at 4 active members.

---

## Requirements (GEARS notation)

> **Subject convention**: generalized subjects ("the super-advisor agent file", "the CLAUDE.md catalog", "the RouteModelFor accessor", "the moai init CLI", "the model-policy doctrine"). No legacy `IF/THEN` modality (per GEARS — `When <undesired>` form is permitted).

### M1 — super-advisor agent file

#### REQ-AA2-001 — Ubiquitous — super-advisor agent file

The super-advisor agent file (`.claude/agents/moai/super-advisor.md` + template mirror `internal/template/templates/.claude/agents/moai/super-advisor.md`) SHALL exist with: (a) `model: inherit` frontmatter (tier-routing injects Opus at spawn time — NOT a frontmatter pin, preserving the `[1m]`-safety rationale per model-policy.md § Inherit-by-Default); (b) `effort: xhigh` FIXED in frontmatter (per §05: "자문 = 최대 추론 (frontmatter 고정)"); (c) read-only tool whitelist (no Write/Edit — `tools: Read, Grep, Glob, Bash, WebFetch, Skill`); (d) `permissionMode: plan`; (e) `description:` carrying an explicit `NOT for:` mutual-exclusion clause vs auditors ("NOT for: gate verdicts (plan-auditor/sync-auditor own binding PASS/FAIL judgment)").

#### REQ-AA2-002 — Ubiquitous — CLAUDE.md §4 ceiling 8 → 10

The CLAUDE.md §4 Retained Agents table (live + template mirror) SHALL move from "8 retained agents" to "10 retained agents" by adding rows for `super-advisor` (class: `meta/advisor`, phase scope: "on-demand high-reasoning consultation across all phases") and `manager-design` (class: `core/manager`, phase scope: "design-phase Claude Design integration"). The Selection Decision Tree SHALL add: "10. On-demand high-reasoning advisor consultation? Use the `super-advisor` subagent" and "11. Design-phase Claude Design integration? Use the `manager-design` subagent".

#### REQ-AA2-003 — Event-driven (When) — super-advisor escalation doctrine E1-E4

**When** one of four escalation triggers fires — E1 bug-deadlock (3+ consecutive same-diagnostic failures), E2 architecture/design decision point, E3 second-opinion request (orchestrator uncertainty), E4 loop deadlock (`/moai loop` or `/moai fix` ceiling-exit per SPEC-LOOP-VERDICT-CONTRACT-001) — the orchestrator SHALL consult the super-advisor (per-spawn `Agent(general-purpose)` with the super-advisor role, Opus model arg per performance tier, xhigh effort), receive a non-binding prescription, and either re-seed the executor or escalate to the user. The doctrine SHALL be embedded in `.claude/rules/moai/core/agent-common-protocol.md` § Error Recovery Pattern with a cross-reference to the super-advisor agent file. E1-E4 entry conditions are exhaustive; expansion is M4 doctrine-refresh territory.

### M2 — manager-design agent file + design phase

#### REQ-AA2-004 — Ubiquitous — manager-design agent file with H1-H9 verbatim

The manager-design agent file (`.claude/agents/moai/manager-design.md` + template mirror) SHALL exist with the frontmatter per §04 codeblock (verbatim: `name: manager-design`, `tools: Read, Write, Edit, Grep, Glob, Bash, DesignSync`, `model: inherit`, `effort: xhigh` FIXED across all tiers, `permissionMode: acceptEdits`, `isolation: worktree`, `memory: project`, `skills: [moai-domain-frontend]`) AND the H1-H9 handoff contract (§04 D4 Handoff Contract table) embedded VERBATIM in the agent body — H1 수신 경로, H2 배치 규약, H3 1:1 충실도, H4 브랜드 우선, H5 주석 변환, H6 검증, H7 보안, H8 재위임 패키지, H9 숨김 폴더 안내. Each H-clause SHALL carry its "위반·실패 시 행동" (violation/failure action) verbatim.

#### REQ-AA2-005 — Event-driven (When) — D1-D5 design pipeline workflow

**When** a UI-surfaced SPEC enters the design phase, the manager-design agent SHALL execute the 5-step pipeline verbatim from §04 flow diagram: D1 연결 준비 (login/project setup via `list_projects`/`create_project`/`get_project`), D2 디자인 시스템 생성·동기화 (`finalize_plan(planId)` user-approval → `write_files(localPath)` component-unit increment), D3 화면 결과물 생성 (Claude Design canvas + `report_validate` metrics), D4 핸드오프 수신·붙여넣기 (`/design-sync` pull user guidance OR `get_file` receive → paste to reserved paths), D5 구현 연결 (Section A-E delegation package to manager-develop). The workflow skill (`.claude/skills/moai/workflows/design.md`) SHALL carry the D1-D5 prose; the agent file references it.

#### REQ-AA2-006 — Capability gate (Where) — spec-workflow plan→design→run conditional route

**Where** a SPEC declares a UI surface (heuristic: explicit frontend-component / view / page deliverable in acceptance.md, OR `tier: L` + frontend module), the spec-workflow (`spec-workflow.md` § SPEC Phase Discipline) SHALL route `plan → design → run` (design enters AFTER plan-audit PASS + Implementation Kickoff Approval, BEFORE run-phase M1 commit). **Where** no UI surface is declared, the route remains the standard `plan → run → sync`. The conditional route is additive — it does not change the plan→run→sync ordering for non-UI SPECs.

#### REQ-AA2-007 — Ubiquitous — designer role_profile + pencil MCP absorption

The `workflow.yaml role_profiles.designer` entry and the `pencil` MCP server registration SHALL be documented as absorbed into manager-design (cross-referenced from `team-protocol.md` role matrix and `settings-management.md` MCP catalog). The absorption is doc-layer: `role_profiles.designer` gains a `# Absorbed by manager-design (SPEC-AGENT-ARCH-V2-001 M2)` annotation; the pencil MCP server entry remains (it is the `.pen` file editor) but its primary consumer becomes manager-design.

### M3 — No-Haiku 3-tier token policy (Go code)

#### REQ-AA2-008 — Ubiquitous — RouteModelFor 3-arg extension

The `RouteModelFor` accessor (`internal/config/model_routing.go:89`) SHALL extend from 2-arg `RouteModelFor(tier, phase string)` to 3-arg `RouteModelFor(specTier, phase, perfTier string)` returning the entry from `model_routing_profiles[perfTier][specTier-phase]`. The current 2-arg call surface (zero external call sites per research.md §A baseline) makes this signature change safe. The closed-set validation (`validRoutingTiers`, `validRoutingPhases`) SHALL add `validRoutingPerfTiers = {max, medium, low}`; the `defaultRoutingEntry` fallback semantics SHALL be preserved (absent pair → `FallbackApplied: true`).

#### REQ-AA2-009 — Ubiquitous — model_routing_profiles 3 matrices

The workflow config (`workflow.yaml` + template mirror) SHALL carry `model_routing_profiles.{max, medium, low}` — 3 matrices of 12 cells each per §2-D table (S/M/L × plan/run/sync/mx × max/medium/low). Every cell value SHALL be `{model, effort}` with model ∈ `{inherit, sonnet, opus, glm}` (haiku REMOVED from the closed set per REQ-AA2-012) and effort ∈ `{low, medium, high, xhigh, max}`. The default performance tier is `medium`.

#### REQ-AA2-010 — Event-detected (When) — moai init --model-policy flag rename

**When** `moai init` runs, the CLI SHALL accept `--model-policy max|medium|low` (redefining the legacy `high/medium/low` flag's semantics: the flag names are reused but their meaning shifts from "model class" to "performance tier per §2-A"). Invalid values outside the 3-enum SHALL exit non-zero with a stderr usage error. The selected tier SHALL persist to `llm.yaml performance_tier` (existing field, default `medium`).

#### REQ-AA2-011 — Ubiquitous — claude_models low: haiku → sonnet

The `llm.yaml claude_models.low` value SHALL flip from `haiku` to `sonnet` (live + template mirror). The `high: opus` and `medium: sonnet` values are unchanged. The haiku key is REMOVED from `claude_models` (per §2-E: "haiku 항목 제거").

#### REQ-AA2-012 — Unwanted behavior — haiku-residual-0 lint rule (HARD SUCCESS METRIC)

The lint ruleset SHALL include a `HaikuResidualRule` that fails when any of: agent frontmatter `model: haiku`, `claude_models` haiku key, `model_routing_profiles` cell with `model: haiku`, `workflow_agents` entry with `model: haiku`, `role_profiles` entry with `model: haiku`, OR `validRoutingModels["haiku": true]` in `model_routing.go`. **This is the HARD success metric (§08 row 1): haiku 참조 잔존 0건 — Target: lint 0건.** The rule MUST NOT be skip-able via `lint.skip` (it is a HARD gate, not advisory). **Scope clause (binds AC-AA2-012 / AC-AA2-016 grep):** the rule's detection surface is the four in-scope surfaces enumerated above (agent frontmatter, `claude_models` block, routing/workflow_agents/role_profiles config, `validRoutingModels` Go map). The following four exemption surfaces carry the "haiku" string but are NOT violations (per Constraint #3): (X1) `_test.go` fixtures; (X2) `glm.models.haiku` in llm.yaml `glm:` block (Out of Scope); (X3) `model-policy.md` Model Aliases definition (closed-set `inherit|opus|sonnet|haiku` — the alias stays lexically valid by design per research.md §E.1; the haiku-exception PROSE in model-policy.md is removed by REQ-AA2-014 M4, verified by AC-AA2-014); (X4) the HaikuResidualRule's own source file in `internal/spec/` (the rule references "haiku" to detect it). The rule's implementation MUST scope its grep to the four in-scope surfaces with the four exemptions carved out so the gate is mechanically satisfiable.

#### REQ-AA2-013 — Ubiquitous — workflow_agents + role_profiles haiku→sonnet substitution

The `workflow.yaml workflow_agents.read-only-extract` entry (`model: haiku, effort: low`) SHALL become `model: sonnet, effort: low`; the `role_profiles.researcher` (`model: haiku`) SHALL become `model: sonnet`; the `role_profiles.reviewer` model SHALL be verified non-haiku. All haiku references in `team-protocol.md` role matrix and `team-pattern-cookbook.md` SHALL be updated to sonnet (live + template mirrors).

### M4 — doctrine refresh

#### REQ-AA2-014 — Ubiquitous — doctrine refresh (model-policy.md + agent-authoring.md + agent-patterns.md)

The doctrine files SHALL be refreshed (live + template mirrors) per SSOT §06 M4 (three sub-items verbatim): **(a)** `model-policy.md` — (i) fable enum · v2.1.196 모델 우선순위 · v2.1.198 Explore 상속 반영 (the `fable` model alias, CC v2.1.196 model-priority updates, and CC v2.1.198 Explore session-model inheritance are reflected in the model doctrine); (ii) § Model Policy Tiers replaced by §2-B agent×tier matrix; (iii) § Inherit-by-Default haiku-exception prose removed (No-Haiku renders the exception obsolete); (iv) § Effort Calibration Matrix superseded by §2-B 표로 대체 (per SSOT §06 M4 verbatim: "Effort Calibration Matrix를 §2-B 표로 대체"); **(b)** `agent-authoring.md` updated to reference the 10-agent catalog and the super-advisor/manager-design patterns; **(c)** `.claude/skills/moai/references/agent-patterns.md` (or equivalent) updated with the 4-loop mapping (orchestrator 4-Loop mechanism → catalog) and the 4 explicitly rejected alternatives (전면 동적화 / auditor 통합 / 정적 핀 / Time-루프 에이전트) per §06 M4.

### Cross-cutting

#### REQ-AA2-015 — Capability gate (Where) — template-first boundary

**Where** an edited surface has a template mirror under `internal/template/templates/` (verified present 2026-07-09 for: CLAUDE.md, agent files, spec-workflow.md, model-policy.md, agent-authoring.md, workflow.yaml, llm.yaml, team-protocol.md, team-pattern-cookbook.md, settings-management.md), the run-phase SHALL apply edits template-first (edit template source, `make build`) or identically in both trees. New Go code (`internal/config/model_routing.go` extension, `internal/cli/init.go` flag rename, lint rule) is NOT templated.

#### REQ-AA2-016 — Unwanted behavior — haiku-residual-0 HARD success metric

The SPEC SHALL NOT close with any haiku reference remaining in the live or template tree. This is the §08 row 1 success metric (Target: lint 0건) and is co-equal with REQ-AA2-012 — REQ-AA2-012 authors the lint rule; REQ-AA2-016 binds the SPEC's closure to its enforcement.

#### REQ-AA2-017 — Ubiquitous — supersede ADVISOR-RUNG-001 + MODEL-ROUTING-WIRE-001

The two Workflow-Reflex draft SPECs SHALL transition `draft → superseded` with `superseded_by: SPEC-AGENT-ARCH-V2-001` and `updated: 2026-07-09`. Their Epic Context sections SHALL carry an inline supersede note ("SUPERSEDED by SPEC-AGENT-ARCH-V2-001 — concerns absorbed: <list>"). The Workflow-Reflex Epic is hereby reconfigured from 6 active members to 4 (remaining: HARNESS-RATCHET-REWIRE-001, LOOP-VERDICT-CONTRACT-001, CADENCE-BRIDGE-001, OBSERVE-HYGIENE-001).

---

## Constraints

1. **Design SSOT is architecture authority (HARD)** — every requirement traces to a verbatim §NN section of `.moai/reports/agent-architecture-redesign-v2-20260709.html`. Re-deriving or inventing architecture is prohibited; if a gap between the SSOT and the live tree is found at run-phase, halt and return a blocker report.
2. **`[1m]` workaround preserved (HARD)** — super-advisor and manager-design both use `model: inherit` frontmatter (tier-routing injects the model at spawn); NO frontmatter model pin. The per-spawn runtime-arg channel (per `model-policy.md` § Inherit-by-Default) is the wiring mechanism; changing the `[1m]` bug workaround itself is out of scope.
3. **Haiku residual = 0 is HARD (§08 row 1)** — the SPEC MUST NOT close with haiku references in the four in-scope surfaces: (1) agent frontmatter `model:`; (2) `claude_models` block in llm.yaml (NOT the `glm.models.*` block — Out of Scope per "Out of Scope — CG mode / GLM model tables"); (3) `model_routing_profiles` / `workflow_agents` / `role_profiles` in workflow.yaml; (4) `validRoutingModels` Go map in model_routing.go. REQ-AA2-012 / REQ-AA2-016 bind. **Exemption surfaces (carry "haiku" string but are NOT violations):** (X1) `_test.go` fixtures — test fixtures may reference haiku for regression-test purposes; (X2) `glm.models.haiku` in llm.yaml `glm:` block — Out of Scope per CG mode; (X3) `model-policy.md` Model Aliases definition (`inherit|opus|sonnet|haiku` closed-set + agent-schema rule) — the alias remains lexically valid by design (research.md §E.1); model-policy.md haiku-exception PROSE removal is M4-scope (REQ-AA2-014), verified by AC-AA2-014; (X4) `internal/spec/` HaikuResidualRule own source file — the rule references "haiku" to detect it. The AC-AA2-012 / AC-AA2-016 closure gates are scoped to the four in-scope surfaces with the four exemptions carved out so the gate is mechanically satisfiable post-M3.
4. **DesignSync MCP tool availability is a run-phase precondition for M2** — the DesignSync MCP server is NOT registered in `.mcp.json` at plan-phase (research.md §H). M2's D1-D5 pipeline + H1-H9 contract are authred against the §04 documented tool contract (11 methods), but M2 run-phase execution MUST verify the tool is registered and operationally available before exercising D2 (`finalize_plan` / `write_files`). Tool absence triggers the H1 blocker-report path (graceful degradation).
5. **Subagent boundary (HARD)** — super-advisor and manager-design are subagents: they return diagnoses/handoff packages to the orchestrator and NEVER prompt the user (askuser-protocol.md § Orchestrator–Subagent Boundary). `/design-login` and `/design-sync` are user-only TUI commands; the agents guide their use, never invoke them.
6. **Pre-existing uncommitted edits PRESERVED VERBATIM** — the working tree carries unrelated modified files (`llm.yaml`, `workflow.yaml`, `manager-docs.md`, `manager-git.md`, `internal/statusline/*`, `pkg/version/version.go`, `system.yaml`). These are NOT in this SPEC's scope. `llm.yaml team_mode: glm` is RUNTIME state from the current `moai glm` session and MUST NOT be touched. This SPEC's run-phase commits MUST NOT `git add` these files; only the v2-owned changes (new agent files, config extensions, doctrine refreshes, the 2 supersede flips) enter commits.
7. **Template-First Rule (CLAUDE.local.md §2 [HARD])** — every doc/config edit lands in `internal/template/templates/` first, then `make build`, then sync to live tree.
8. **GEARS notation; era V3R6; 12 canonical frontmatter fields; no snake_case aliases** (created/updated/tags canonical).
9. **Implementation Kickoff Approval (plan→run HUMAN GATE) is mandatory** — per `.claude/rules/moai/workflow/orchestration-mode-selection.md` header + CLAUDE.local.md §19.1. No run-phase work begins without explicit user approval post plan-audit verdict.

---

## Out of Scope

> Per the `OutOfScopeRule` lint, this section uses `### Out of Scope — <topic>` H3 sub-headings with `-` bullets.

### Out of Scope — `[1m]` bug workaround redesign

- Changing the inherit-by-default frontmatter convention, the Anthropic #45847/#51060/#36670 workaround mechanics, or the baseline-refill-breaker doctrine. This SPEC routes through the existing per-spawn-arg safe channel; it does not redesign the channel.

### Out of Scope — Sonnet 5 / Opus 4.8 lineup change

- The "전 워커 Sonnet 단일화 전제" (§08 row 6: "2026-07-09 선행 완료") is treated as a frozen precondition. This SPEC does not change which Sonnet/Opus variant ships; it only changes how effort is tiered within Sonnet.

### Out of Scope — Agent Teams default flip

- `workflow.team.enabled` stays `false` (per the Sonnet 5 / Opus 4.8 re-design, SPEC-SONNET5-1M-TEAM-DISABLE). manager-design and super-advisor are sub-agent spawns (Mode 4/5), NOT agent-team teammates.

### Out of Scope — CG mode / GLM model tables

- `moai cg` / GLM teammate routing, GLM model tables, `team_mode: glm` runtime state. The `glm` value in the routing closed set is untouched; GLM-context carve-outs (CLAUDE.md §15) remain as-is.

### Out of Scope — Workflow-Reflex Epic re-numbering

- The 4 remaining Workflow-Reflex members keep their historical "N of 6" labels (1/3/5/6). Renumbering them to "N of 4" is cosmetic and out of scope; their labels remain factually true as historical positioning.

### Out of Scope — super-advisor mechanical spawn trigger

- A Go/hook layer that mechanically detects E1-E4 conditions and spawns the super-advisor. This SPEC's wiring is prompt-layer (doctrine in agent-common-protocol.md); runtime auto-spawn is follow-up territory.

### Out of Scope — manager-design sync-auditor brand-consistency gate

- M2 adds manager-design as a design-phase agent; M5 (sync-auditor brand must-pass) is referenced from H8 but the gate itself is owned by sync-auditor's existing 4-dimension scoring. This SPEC does not author a new sync-auditor dimension.

### Out of Scope — v1 → v2 migration tooling

- A automated migration script that rewrites existing SPEC frontmatter or agent files to the v2 conventions. Run-phase applies edits directly; no migration CLI is authored.

---

## Cross-References

- **Design SSOT (architecture authority)**: `.moai/reports/agent-architecture-redesign-v2-20260709.html` — §01 v1→v2 delta, §02 No-Haiku 3-tier policy (2-A through 2-E), §03 target architecture, §04 manager-design (D1-D5 + H1-H9), §05 catalog verdict 9+1, §06 milestones M1-M4, §07 risks, §08 success metrics.
- **Extension points (Go)**: `internal/config/model_routing.go:89` (`RouteModelFor` 2-arg → 3-arg); `internal/config/types.go` (ModelRouting map structure); `internal/cli/init.go` (`--model-policy` flag); `internal/spec/lint.go` (new `HaikuResidualRule`).
- **EXTEND base (doc)**: `CLAUDE.md` §4 Retained Agents (ceiling 8→10); `.claude/rules/moai/core/agent-common-protocol.md` § Error Recovery Pattern (super-advisor E1-E4); `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Phase Discipline (plan→design→run route); `.claude/rules/moai/development/model-policy.md` (§2-B supersede per SSOT §06 M4 verbatim); `.claude/rules/moai/development/agent-authoring.md`; `.claude/rules/moai/workflow/team-protocol.md` role matrix; `.claude/rules/moai/workflow/team-pattern-cookbook.md`.
- **EXTEND base (config)**: `.moai/config/sections/workflow.yaml` (model_routing_profiles, workflow_agents, role_profiles); `.moai/config/sections/llm.yaml` (claude_models, performance_tier).
- **New agent files**: `.claude/agents/moai/super-advisor.md`; `.claude/agents/moai/manager-design.md` (+ template mirrors).
- **New workflow skill**: `.claude/skills/moai/workflows/design.md` (D1-D5 pipeline).
- **Superseded SPECs**: SPEC-ADVISOR-RUNG-001 (frontmatter flip + inline note); SPEC-MODEL-ROUTING-WIRE-001 (frontmatter flip + inline note).
- **Workflow-Reflex Epic remaining (4 active)**: SPEC-HARNESS-RATCHET-REWIRE-001 (1 of 6); SPEC-LOOP-VERDICT-CONTRACT-001 (3 of 6); SPEC-CADENCE-BRIDGE-001 (5 of 6); SPEC-OBSERVE-HYGIENE-001 (6 of 6).
- **Primitives cited (unmodified)**: `model-policy.md` § Inherit-by-Default (`[1m]`-safe per-spawn-arg rationale); `archived-agent-rejection.md` §C (per-spawn `Agent(general-purpose)` pattern — the basis super-advisor promotes to a catalog agent).
- **External (Claude Design Labs)**: anthropic.com/news/claude-design-anthropic-labs; support.claude.com/en/articles/14604416; code.claude.com/docs/en/sub-agents; code.claude.com/docs/en/best-practices.

---

## History

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-09 | manager-spec | Initial draft — plan-phase artifacts (spec + plan + acceptance + design + research + progress). v2 architecture: super-advisor + manager-design + No-Haiku 3-tier token policy. Tier L. Supersedes SPEC-ADVISOR-RUNG-001 + SPEC-MODEL-ROUTING-WIRE-001. |
| 0.2.0 | 2026-07-09 | manager-spec | plan-audit iter-2 revision. D1 BLOCKING resolved: AC-AA2-012/016 closure gate narrowed to 4 in-scope surfaces (agent frontmatter, claude_models block, routing config, validRoutingModels map) with 4 exemption surfaces carved out (X1 _test.go, X2 glm.models.haiku, X3 model-policy.md alias-def, X4 HaikuResidualRule own source) — gate now mechanically satisfiable. Constraint #3 + REQ-AA2-012 scope clause expanded with exemption enumeration. D2: REQ-AA2-014 Effort Calibration Matrix target aligned to SSOT §06 M4 verbatim (§2-C → §2-B). D3: REQ-AA2-014 sub-item (a)(i) added per SSOT §06 M4 (fable enum · v2.1.196 · v2.1.198). D4: AC-AA2-001 tools-whitelist verification added. D5: AC-AA2-002 "8 retained agents" absence check added. D6: AC-AA2-009 verification-block grep indentation aligned to When-clause. |
