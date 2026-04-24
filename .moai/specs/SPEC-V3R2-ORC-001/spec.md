---
id: SPEC-V3R2-ORC-001
title: "Agent roster consolidation (22 → 17)"
version: "0.1.0"
status: draft
created: 2026-04-23
updated: 2026-04-23
author: GOOS
priority: P0 Critical
phase: "v3.0.0 — Phase 3 — Agent Cleanup"
module: ".claude/agents/moai/, internal/template/templates/.claude/agents/moai/"
dependencies:
  - SPEC-V3R2-CON-001
related_problem:
  - P-A05
  - P-A06
  - P-A07
  - P-A08
  - P-A09
  - P-A10
  - P-A12
  - P-A14
  - P-A15
  - P-A16
  - P-A17
  - P-A20
  - P-A21
  - P-A23
related_theme: "Layer 4 — Orchestration, Master §4.4, §7.2, §8 BC-V3R2-005, BC-V3R2-009, BC-V3R2-016"
breaking: true
bc_id: [BC-V3R2-005, BC-V3R2-009, BC-V3R2-016]
lifecycle: spec-anchored
tags: "agent, roster, consolidation, manager-cycle, builder-platform, manager-quality, v3r2"
---

# SPEC-V3R2-ORC-001: Agent roster consolidation (22 → 17)

## HISTORY

| Version | Date       | Author | Description                          |
|---------|------------|--------|--------------------------------------|
| 0.1.0   | 2026-04-23 | GOOS   | Initial draft (Wave 4 SPEC writer, round 2) |

---

## 1. Goal (목적)

MoAI v2.13.2 agent roster (22 agents) exhibits structural over-decomposition: 3 near-identical builders, 2 managers with 60% body overlap (ddd/tdd), 4 advisor-only experts that delegate every real activity. R5 audit (`.moai/design/v3-redesign/research/r5-agent-audit.md`) grades the roster at mean 28.8/33 with systemic overlap. v3 Round 2 consolidates 22 agents into **17 agents** by (a) merging manager-ddd + manager-tdd into a single parameterized `manager-cycle`, (b) collapsing builder-agent + builder-skill + builder-plugin into `builder-platform`, (c) retiring expert-debug as a diagnostic sub-mode of manager-quality, (d) retiring expert-testing by splitting strategy responsibility into manager-cycle and load-test execution into expert-performance, (e) scope-shrinking manager-project to `.moai/project/` document generation only.

This SPEC owns the agent-file deltas (creation of new agents, retirement of merged agents, stub redirects) and the migration-time SPEC reference rewriting. Common-Protocol CI lint is deferred to SPEC-V3R2-ORC-002; effort-level population to SPEC-V3R2-ORC-003; worktree flag application to SPEC-V3R2-ORC-004.

### 1.1 Background

R5 audit §Recommended v3 agent inventory catalogs the current 22 agents and their destiny. The consolidation rationale for each merge/retire decision is grounded in concrete R5 evidence:

- **manager-ddd + manager-tdd overlap (P-A09)**: ≈60% body is identical (LSP baseline, checkpoint, @MX, decision guide, steps 1-5). Real differences are three phase-name substitutions (ANALYZE-PRESERVE-IMPROVE vs RED-GREEN-REFACTOR) and one coverage heuristic. Merge into `manager-cycle` with `cycle_type: ddd|tdd` parameter.
- **3-builder near-identity (P-A05)**: `builder-agent`, `builder-skill`, `builder-plugin` share the same frontmatter (sonnet / bypassPermissions / memory:user), same tool list, same 5-phase workflow body. Artifact-template differences are fully representable by a single `artifact_type` parameter.
- **expert-debug as router (P-A06)**: 70% of the ~100 LOC body is a delegation table pointing at manager-ddd, manager-quality, expert-refactoring. Agent declares no Write/Edit tools. Fold into manager-quality as a diagnostic sub-mode.
- **expert-testing as strategy-only (P-A07)**: Self-declared OUT OF SCOPE includes unit-test execution (delegated to manager-ddd) and load-test execution (delegated to expert-performance). Remaining scope is a strategy doc that belongs in manager-cycle RED planning.
- **manager-project over-scoped (P-A10, P-A12)**: Body runs 6 routing modes; 3 of them (settings_modification, glm_configuration, template_update_optimization) belong in the `moai` CLI binary, not in an agent. Frontmatter scope line "File creation restricted to .moai/project/ only" already contradicts body workflow.
- **expert-frontend Pencil split (P-A20)**: Expert-frontend body mixes React/Vue code implementation with Pencil MCP design workflow (17 MCP tools). v3 keeps expert-frontend for code authorship and documents the Pencil scope recommendation for Phase 6 (`/moai design` path B); the split itself is deferred to WF-003 since the design-flow readiness gates it.

*Source: R5 §Detailed findings; Master §4.4, §7.2; problem-catalog.md P-A05..P-A20.*

### 1.2 Non-Goals

- Common-Protocol CI lint implementation (SPEC-V3R2-ORC-002 owns the lint and the AskUserQuestion scrubs).
- Effort-level matrix population across all 17 agents (SPEC-V3R2-ORC-003).
- `isolation: worktree` MUST upgrade (SPEC-V3R2-ORC-004).
- Common-Protocol skeptical-evaluator block DRY extraction (SPEC-V3R2-ORC-002 §Duplicate Skeptical Mandate).
- Adding `manager-design` agent (P-A21 deferred per Master §6 deferred list and §12 open question #6 — re-evaluate after WF-003 ships).
- Creating new agents outside the merge targets (builder-platform and manager-cycle are the only new agents).
- Changes to evaluator-active / plan-auditor / expert-security beyond frontmatter additions specified elsewhere.

---

## 2. Scope (범위)

### 2.1 In Scope

- Create `.claude/agents/moai/manager-cycle.md` replacing `manager-ddd.md` + `manager-tdd.md`; accept `cycle_type: ddd|tdd` parameter documented in the body; preserve the UNION of both agents' trigger keywords.
- Create `.claude/agents/moai/builder-platform.md` replacing `builder-agent.md` + `builder-skill.md` + `builder-plugin.md`; accept `artifact_type: agent|skill|plugin|command|hook|mcp-server|lsp-server` parameter; preserve UNION of trigger keywords across the three originals.
- Delete `.claude/agents/moai/expert-debug.md`; extend `.claude/agents/moai/manager-quality.md` with a diagnostic sub-mode section that absorbs expert-debug's `--deepthink debug` routing.
- Delete `.claude/agents/moai/expert-testing.md`; extend `manager-cycle.md` with a RED-phase test-strategy subsection; extend `expert-performance.md` with a `--deepthink load-test` mode covering k6/Locust/JMeter execution.
- Delete `.claude/agents/moai/manager-ddd.md`, `manager-tdd.md`, `builder-agent.md`, `builder-skill.md`, `builder-plugin.md` (replaced by new merged agents).
- Scope-shrink `.claude/agents/moai/manager-project.md`: body restricted to `.moai/project/{product,structure,tech}.md` generation; remove settings_modification / glm_configuration / template_update_optimization modes; remove self-contradicting AskUserQuestion workflow lines.
- Trim `expert-backend.md` trigger list to 12-15 dedup keywords per P-A17; trim `manager-git.md` trigger list per P-A16.
- Remove `mcp__context7__*` preload from agents that never use library docs per P-A15 (manager-git, manager-quality; validate against R5 table).
- Align `memory:` field defaults across roster per P-A14 / P-A23 (plan-auditor gains `memory: project`; manager-project / plan-auditor / researcher all preload `moai-foundation-core`).
- Template-first: all deletes and creates MUST occur under `internal/template/templates/.claude/agents/moai/` first, then `make build` regenerates embedded copy, then local `.claude/agents/moai/` is kept byte-identical.
- Emit stub agents (one-line redirect bodies) at deprecated names for one v3.x minor cycle per BC-V3R2-009 / BC-V3R2-016 so that legacy SPEC references do not break on rollout.

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- AskUserQuestion scrub text changes inside agent bodies (delegated to SPEC-V3R2-ORC-002 where CI lint also lands).
- `effort:` field population on any of the 17 final agents (delegated to SPEC-V3R2-ORC-003).
- `isolation: worktree` additions on implementer/tester/designer agents (delegated to SPEC-V3R2-ORC-004).
- Common-Protocol Skeptical Mandate block DRY extraction (delegated to SPEC-V3R2-ORC-002).
- Adding a new `manager-design` agent (P-A21 deferred per Master §13 — revisit after WF-003).
- Changing frontmatter schema (handled by v3-legacy SPEC-V3-AGT-001 which continues into v3r2 via CON-002 amendment protocol).
- Builder-platform code implementation (agent body authorship only; the Go-side dispatcher for `artifact_type:` lives in CLI, out of scope here).
- expert-frontend split of Pencil/code authorship (defer to WF-003 when `/moai design` path B is first-class).
- Adjusting evaluator-active or plan-auditor body text beyond what OR C-002/003 prescribe.
- Breaking changes beyond those listed in BC-V3R2-005, 006, 009, 016.

---

## 3. Environment (환경)

- Runtime: moai-adk-go v3.0.0-alpha.3+ (target phase per Master §9)
- Claude Code v2.1.111+ with Agent Teams support (team mode file-ownership implications)
- Opus 4.7 Adaptive Thinking compatible; agents run on sonnet / opus model mix per existing frontmatter
- Target directory (template first): `internal/template/templates/.claude/agents/moai/`
- Target directory (rendered): `.claude/agents/moai/`
- Local and template trees MUST remain byte-identical after `make build` (verified via `diff -r` gate from CLAUDE.local.md §2)
- Affected roster size: 22 → 17 (net −5)
- Dependent SPEC migrator: SPEC-V3R2-MIG-001 rewrites `manager-ddd` / `manager-tdd` / `builder-*` / `expert-debug` / `expert-testing` references inside legacy SPEC bodies
- Stub retention window: one v3.x minor cycle (per Master §8 deprecation policy)

---

## 4. Assumptions (가정)

- SPEC-V3R2-CON-001 (FROZEN/EVOLVABLE codification) has landed; the 7 FROZEN invariants (Master §1.3) cover agent-common-protocol rules that this SPEC preserves verbatim.
- SPEC-V3R2-ORC-002 (CI lint) will merge in the same v3.0.0-alpha.3 phase and will enforce the "no AskUserQuestion in agent body" rule against the new files created here.
- SPEC-V3R2-ORC-003 will populate `effort` fields on the 17 final agents produced by this SPEC.
- SPEC-V3R2-ORC-004 will add `isolation: worktree` to new write-heavy agents (manager-cycle, builder-platform).
- The union of `manager-ddd` and `manager-tdd` trigger keywords can be accepted without false-positive routing (R1 evidence: R5 language-trigger coverage table shows both agents at 7×4 = 28 keywords each, union ≤ 40 with deduping).
- Stub agents emitting a single "This agent is retired; use {new-name}" line cause no Claude Code runtime error (builder-agent.md still in v2 skill registry tolerates minimal bodies per existing template pattern).
- `diff -r` CI gate between template and local trees can be satisfied in a single PR because `make build` regenerates embedded.go deterministically.
- expert-frontend Pencil-split deferral does not block Phase 3 since `/moai design` path B does not yet ship.

---

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous (항시)

**REQ-ORC-001-001 (Ubiquitous) — 최종 에이전트 수**
The consolidated MoAI v3 agent roster **shall** contain exactly 17 active agent files in `.claude/agents/moai/`: manager-spec, manager-strategy, manager-cycle, manager-quality, manager-docs, manager-git, manager-project, expert-backend, expert-frontend, expert-security, expert-devops, expert-performance, expert-refactoring, builder-platform, evaluator-active, plan-auditor, researcher.

**REQ-ORC-001-002 (Ubiquitous) — 병합 에이전트 정의**
The `manager-cycle.md` agent body **shall** declare `cycle_type: ddd|tdd` as the first required input parameter and **shall** preserve every workflow step previously present in either `manager-ddd.md` or `manager-tdd.md`, parameterizing phase names (ANALYZE-PRESERVE-IMPROVE vs RED-GREEN-REFACTOR) via the cycle_type selector.

**REQ-ORC-001-003 (Ubiquitous) — 빌더 플랫폼 정의**
The `builder-platform.md` agent body **shall** declare `artifact_type: agent|skill|plugin|command|hook|mcp-server|lsp-server` as a required parameter and **shall** retain the 5-phase workflow structure (Requirements → Research → Architecture → Implementation → Validation) from the three retired builder agents.

**REQ-ORC-001-004 (Ubiquitous) — manager-quality 진단 모드**
The `manager-quality.md` agent body **shall** contain a "Diagnostic Sub-Mode" section covering the intake-routing responsibilities previously owned by `expert-debug.md`, including the delegation table that mapped diagnosed failure categories to manager-cycle / expert-refactoring / manager-git.

**REQ-ORC-001-005 (Ubiquitous) — 템플릿 우선 원칙 준수**
Every agent file add/modify/delete performed by this SPEC **shall** first be applied under `internal/template/templates/.claude/agents/moai/`, followed by `make build`, resulting in a byte-identical local `.claude/agents/moai/` tree (CLAUDE.local.md §2 Template-First Rule).

### 5.2 Event-Driven (이벤트 기반)

**REQ-ORC-001-006 (Event-Driven) — 은퇴 에이전트 스텁**
**When** the v3.0.0-alpha.3 release ships, the template tree **shall** contain single-file stubs at the retired paths (`manager-ddd.md`, `manager-tdd.md`, `builder-agent.md`, `builder-skill.md`, `builder-plugin.md`, `expert-debug.md`, `expert-testing.md`) whose body contains exactly one instruction: "This agent has been retired; use {new-name} with parameter {param} instead" — the stub **shall** also carry a minimal frontmatter (`name:`, `description:`, `status: retired`).

**REQ-ORC-001-007 (Event-Driven) — SPEC 참조 마이그레이션**
**When** SPEC-V3R2-MIG-001 rewrites legacy SPEC bodies, occurrences of the retired agent names **shall** be replaced according to the migration table: `manager-ddd` → `manager-cycle` (with `cycle_type: ddd`); `manager-tdd` → `manager-cycle` (with `cycle_type: tdd`); `builder-agent|builder-skill|builder-plugin` → `builder-platform` (with `artifact_type: agent|skill|plugin`); `expert-debug` → `manager-quality` (diagnostic mode); `expert-testing` → `manager-cycle` (strategy) or `expert-performance` (load).

**REQ-ORC-001-008 (Event-Driven) — manager-project 범위 축소**
**When** the manager-project agent is invoked, the spawn prompt **shall** reject any task that writes outside `.moai/project/{product,structure,tech}.md`; the retired modes (settings_modification, glm_configuration, template_update_optimization) **shall** return a blocker report pointing at the `moai` CLI subcommand (`moai glm`, `moai update`, `moai cc`).

### 5.3 State-Driven (상태 기반)

**REQ-ORC-001-009 (State-Driven) — 언어 트리거 보존**
**While** the v3.0.0-alpha.3 release is active, the union of EN/KO/JA/ZH trigger keywords declared on `manager-cycle.md` **shall** contain every trigger keyword previously declared on either `manager-ddd.md` or `manager-tdd.md`; similarly `builder-platform.md` **shall** union all trigger keywords from builder-agent + builder-skill + builder-plugin.

**REQ-ORC-001-010 (State-Driven) — 트리거 중복 제거**
**While** the roster is in its consolidated v3r2 state, `expert-backend.md` EN trigger list **shall** contain no duplicate tokens (including near-duplicates such as "Oracle" appearing as both an EN standalone token and inside a KO translation) and **shall** be bounded at 12-15 EN triggers per P-A17; `manager-git.md` **shall** be reduced to approximately 8 high-precision triggers per P-A16.

**REQ-ORC-001-011 (State-Driven) — MCP 프리로드 감사**
**While** the roster is in its consolidated v3r2 state, `mcp__context7__*` MCP tools **shall** NOT appear in the tools list of any agent that does not invoke Context7 in its body workflow; specifically, `manager-git.md` and `manager-quality.md` **shall** drop the Context7 preload.

### 5.4 Optional (선택)

**REQ-ORC-001-012 (Optional) — manager-quality 메모리**
**Where** the harness configuration enables cross-SPEC pattern retention, `manager-quality.md` **may** declare `memory: project` so that diagnostic patterns persist per project; if declared, the memory path **shall** be `.moai/agents/manager-quality/memory/`.

**REQ-ORC-001-013 (Optional) — plan-auditor 메모리**
**Where** the roster is consolidated, `plan-auditor.md` **shall** gain a `memory: project` frontmatter field (addressing P-A14 — plan-auditor is the most memory-worthy agent across SPECs and currently declares no memory field).

**REQ-ORC-001-014 (Optional) — expert-performance 쓰기 권한**
**Where** the retired `expert-testing.md` load-test duties are absorbed, `expert-performance.md` **may** gain `Write` tool access scoped to `.moai/docs/performance-analysis-*.md` only, resolving the P-A08 contradiction (body says "Create analysis file" but tools list excludes Write).

### 5.5 Unwanted Behavior

**REQ-ORC-001-015 (Unwanted Behavior) — 새로운 advisor-only 에이전트 금지**
**If** a v3r2 agent definition's tool list excludes every write-capable tool (Write, Edit, MultiEdit) AND the body workflow prescribes a file-write side effect, **then** the agent definition **shall** be rejected by CI (carried out by SPEC-V3R2-ORC-002 lint rule). This prevents reintroducing the P-A06/P-A07/P-A08 pattern.

**REQ-ORC-001-016 (Unwanted Behavior) — 삭제 에이전트 미 은퇴 금지**
**If** a PR deletes any file in `.claude/agents/moai/` without adding either a stub or a template-tree mirror of the deletion, **then** CI **shall** fail with error `ORC_AGENT_DELETE_WITHOUT_STUB` pointing at the orphan deletion.

**REQ-ORC-001-017 (Unwanted Behavior) — 트리거 누락 금지**
**If** the merged agent's trigger set fails to include a keyword that appeared in any retired source agent's trigger set, **then** CI **shall** fail with error `ORC_TRIGGER_DROP` listing the dropped keyword(s) to prevent silent routing regressions.

---

## 6. Acceptance Criteria (수용 기준 요약)

Detailed Given-When-Then scenarios are in `acceptance.md`.

Core criteria:

- **AC-ORC-001-01**: Counting active (non-stub) agent files under `.claude/agents/moai/` yields exactly 17, matching the REQ-001 list.
- **AC-ORC-001-02**: `manager-cycle.md` exists, declares `cycle_type` parameter, and its body contains every phase step found in either `manager-ddd.md` or `manager-tdd.md` (parameterized).
- **AC-ORC-001-03**: `builder-platform.md` exists, declares `artifact_type`, and covers all 7 artifact kinds listed in REQ-003.
- **AC-ORC-001-04**: `manager-quality.md` body contains a "Diagnostic Sub-Mode" section whose content matches the expert-debug delegation table semantically.
- **AC-ORC-001-05**: All 7 stub files exist at deprecated paths with status=retired frontmatter.
- **AC-ORC-001-06**: `diff -r internal/template/templates/.claude/agents/moai/ .claude/agents/moai/` returns no differences.
- **AC-ORC-001-07**: Legacy SPEC that previously referenced `manager-ddd` routes successfully to `manager-cycle` via MIG-001 rewriter (integration test).
- **AC-ORC-001-08**: Manual invocation of `manager-project` with a task writing to `.moai/config/sections/language.yaml` returns a blocker report citing the CLI alternative.
- **AC-ORC-001-09**: Script counting duplicate tokens in `expert-backend.md` EN trigger list returns 0; token count is within 12-15.
- **AC-ORC-001-10**: `mcp__context7__*` absent from `manager-git.md` and `manager-quality.md` tools lists.
- **AC-ORC-001-11**: `plan-auditor.md` frontmatter contains `memory: project`.
- **AC-ORC-001-12**: Union-preservation test: every trigger keyword present in retired `manager-ddd.md` or `manager-tdd.md` is present in `manager-cycle.md` (same for builder-platform vs three retired builders).

---

## 7. Constraints (제약)

- [HARD] Template-First (CLAUDE.local.md §2): every delete/create operation starts in `internal/template/templates/.claude/agents/moai/`.
- [HARD] Common-Protocol preservation: no agent body introduced by this SPEC may contain the literal string `AskUserQuestion` (enforced by ORC-002 CI lint; prerequisite for merge).
- [HARD] 16-language neutrality (CLAUDE.local.md §15): agent bodies remain language-agnostic; no language-specific assertions beyond existing per-language rule pointers.
- [HARD] FROZEN constitution (Master §1.3): agent-common-protocol and CLAUDE.md Section 8 (AskUserQuestion-Only-for-Orchestrator) are preserved verbatim.
- [HARD] Stub retention window: retired-agent stubs MUST remain available for one full v3.x minor cycle per Master §8 BC-V3R2-009 / BC-V3R2-016 deprecation rows.
- [HARD] Trigger-union preservation: merging two agents MUST produce the trigger-keyword UNION, not the intersection (prevents routing regressions per REQ-017).
- [HARD] No scope creep: this SPEC changes only agent files, SPEC migrator rules (via MIG-001 dependency), and the manager-project body scope; it does not modify skills, rules, or commands.
- [HARD] No new breaking changes beyond BC-V3R2-005/006/009/016 listed in Master §8.

---

## 8. Risks & Mitigations (리스크 및 완화)

| Risk                                                          | Impact | Mitigation                                                                                               |
|---------------------------------------------------------------|--------|----------------------------------------------------------------------------------------------------------|
| Merging manager-ddd + manager-tdd drops a phase-specific rule | HIGH   | Union-preservation test (AC-12) + human code-review diff pass; staged rollout in alpha.3                 |
| builder-platform parameter dispatch loses an artifact variant | HIGH   | Explicit artifact_type enum table in body; CI test spawning builder-platform per artifact variant        |
| Retired-agent references in legacy SPEC break routing         | HIGH   | SPEC-V3R2-MIG-001 rewriter; stub redirects retained one v3.x cycle                                       |
| manager-project blocker-report pattern over-triggers          | MEDIUM | Explicit scope allowlist in body; CLI subcommand referenced in blocker text                              |
| expert-frontend Pencil split deferred creates technical debt  | MEDIUM | Documented in §1.2 Non-Goals; revisit in WF-003 (Phase 6)                                                |
| Trigger-union dedup miscounts near-duplicates                 | MEDIUM | AC-09 uses a deterministic token-normalization script; reviewed PR-time                                  |
| Phase-name parameterization confuses users of `/moai run`     | LOW    | `/moai run` command skills route to manager-cycle; cycle_type surfaces in skill frontmatter argument-hint |
| Stub frontmatter strictness fails lint                        | LOW    | Minimal-stub template co-designed with ORC-002 lint exemption pattern                                    |
| Legacy dev-only SPEC-V3-AGT-001 conflicts with this SPEC      | LOW    | v3r2 namespace isolates the two; MIG-001 is the single authoritative rewriter                            |

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- **SPEC-V3R2-CON-001** (FROZEN/EVOLVABLE codification) — constitution zone model must land before agent-file deletes (stubs and deletes are zone-sensitive).
- **SPEC-V3R2-ORC-002** (Common Protocol CI lint) — prerequisite for merging new agent files since lint rejects any body containing literal `AskUserQuestion`.
- **SPEC-V3R2-ORC-003** (Effort matrix) — populates `effort` on the 17 final agents this SPEC creates.
- **SPEC-V3R2-ORC-004** (Worktree MUST) — adds `isolation: worktree` to new write-heavy agents (manager-cycle, builder-platform).

### 9.2 Blocks

- **SPEC-V3R2-MIG-001** (v2→v3 user migrator) — depends on the final 17-agent roster to generate the rewriter table.
- **SPEC-V3R2-WF-003** (Multi-mode router) — depends on manager-cycle as the target for `/moai run --mode autopilot|loop|team`.
- **SPEC-V3R2-HRN-001** (Harness routing) — references the 17 agents when routing harness levels to agents.

### 9.3 Related

- **SPEC-V3-AGT-001** (legacy v3 Agent Frontmatter Expansion) — complementary; v3r2 inherits the frontmatter schema while this SPEC addresses roster structure.
- **SPEC-V3R2-CON-003** (Constitution consolidation) — merges CLAUDE.md §4 Agent Catalog counts post-consolidation.

---

## 10. Traceability (추적성)

- REQ-to-AC mapping: REQ-001 → AC-01; REQ-002 → AC-02, AC-12; REQ-003 → AC-03, AC-12; REQ-004 → AC-04; REQ-005 → AC-06; REQ-006 → AC-05; REQ-007 → AC-07; REQ-008 → AC-08; REQ-009 → AC-12; REQ-010 → AC-09; REQ-011 → AC-10; REQ-013 → AC-11; REQ-015, REQ-016, REQ-017 → CI-only enforcement (fail-path tests in `.moai/specs/SPEC-V3R2-ORC-002/acceptance.md`).
- Total REQ count: 17 (Ubiquitous 5, Event-Driven 3, State-Driven 3, Optional 3, Unwanted 3)
- Expected AC count: 12
- Wave 1/2 sources:
  - `r5-agent-audit.md` §Recommended v3 agent inventory (17-agent target)
  - `r5-agent-audit.md` §Per-agent audit table (detailed verdicts)
  - `problem-catalog.md` P-A05..P-A20 cluster (roster problems)
  - `pattern-library.md` O-6 (Agent count matches task structure)
  - `design-principles.md` Principle 10 (Parallelism via Explicit DAG implies well-bounded role set) and Principle 11 (Agent Count Matches Task Structure)
  - `major-v3-master.md` §4.4 Layer 4 Orchestration, §7.2 Agent inventory table, §8 BC-V3R2-005/006/009/016
- **File:line anchors** (per D5 traceability requirement):
  - `docs/design/major-v3-master.md:L1050-1056` (§11.4 Orchestration — ORC-001..005 definitions)
  - `docs/design/major-v3-master.md:L963-976` (§8 BC catalog — BC-V3R2-005/009/016)
  - `docs/design/major-v3-master.md:L989` (§9 Release Plan — Phase 3 roster)
  - `.moai/design/v3-redesign/research/r5-agent-audit.md` §Recommended v3 agent inventory
  - `.moai/design/v3-redesign/synthesis/problem-catalog.md` Cluster 6 (P-A05..P-A20)
  - `.moai/design/v3-redesign/synthesis/pattern-library.md` §O-6
- Code-side paths:
  - `.claude/agents/moai/manager-cycle.md` (new, REQ-001, REQ-002, REQ-009)
  - `.claude/agents/moai/builder-platform.md` (new, REQ-001, REQ-003, REQ-009)
  - `.claude/agents/moai/manager-quality.md` (modified, REQ-004, REQ-011, REQ-012)
  - `.claude/agents/moai/manager-project.md` (scope-shrunk, REQ-008)
  - `.claude/agents/moai/plan-auditor.md` (modified, REQ-013)
  - `.claude/agents/moai/expert-backend.md` (trigger dedup, REQ-010)
  - `.claude/agents/moai/expert-performance.md` (load-test absorption, REQ-014)
  - `.claude/agents/moai/expert-testing.md` (retired, REQ-006)
  - `.claude/agents/moai/expert-debug.md` (retired, REQ-006)
  - `.claude/agents/moai/manager-ddd.md`, `manager-tdd.md`, `builder-agent.md`, `builder-skill.md`, `builder-plugin.md` (retired stubs, REQ-006)
  - `internal/template/templates/.claude/agents/moai/` (template-first mirrors, REQ-005)

### 10.1 Consolidation destiny table (all 22 current agents)

| # | Current agent | Category | v3r2 Destiny | Notes |
|---|---------------|----------|--------------|-------|
| 1 | manager-spec | manager | KEEP | REQ-001 roster member |
| 2 | manager-strategy | manager | KEEP | REQ-001 roster member |
| 3 | manager-ddd | manager | MERGE → manager-cycle | Retired; REQ-006 stub; BC-V3R2-009, BC-V3R2-016 |
| 4 | manager-tdd | manager | MERGE → manager-cycle | Retired; REQ-006 stub; BC-V3R2-016 |
| 5 | manager-quality | manager | REFACTOR | Absorbs expert-debug diagnostic mode (REQ-004); drops Context7 (REQ-011) |
| 6 | manager-docs | manager | KEEP | REQ-001 roster member |
| 7 | manager-git | manager | REFACTOR | Trim trigger set (REQ-010); drop Context7 (REQ-011) |
| 8 | manager-project | manager | REFACTOR | Scope-shrunk (REQ-008); retains REQ-001 roster slot |
| 9 | expert-backend | expert | REFACTOR | Trigger dedup (REQ-010); retains roster slot |
| 10 | expert-frontend | expert | KEEP | Pencil-split deferred to WF-003 per §1.2 |
| 11 | expert-security | expert | KEEP | Agent tool removal in ORC-002 |
| 12 | expert-devops | expert | KEEP | AskUserQuestion scrub in ORC-002 |
| 13 | expert-performance | expert | REFACTOR | Optional Write scope (REQ-014); absorbs expert-testing load-test |
| 14 | expert-debug | expert | RETIRE → manager-quality | Merged into diagnostic sub-mode (REQ-004, REQ-006) |
| 15 | expert-testing | expert | RETIRE (split) | Strategy → manager-cycle; load → expert-performance (REQ-006) |
| 16 | expert-refactoring | expert | KEEP | Worktree add in ORC-004 |
| 17 | builder-agent | builder | MERGE → builder-platform | Retired (REQ-006); BC-V3R2-009 |
| 18 | builder-skill | builder | MERGE → builder-platform | Retired (REQ-006) |
| 19 | builder-plugin | builder | MERGE → builder-platform | Retired (REQ-006) |
| 20 | evaluator-active | evaluator | KEEP | Effort upgrade in ORC-003 |
| 21 | plan-auditor | evaluator | REFACTOR | Gains `memory: project` (REQ-013); effort upgrade in ORC-003 |
| 22 | researcher | meta | KEEP (or RETIRE) | KEEP selected per Master §7.2; worktree add in ORC-004 |

**Net delta: 22 → 17 (−5): manager-ddd, manager-tdd, expert-debug, expert-testing, 3 builders retire; manager-cycle and builder-platform appear net +2; arithmetic: −4 − 3 + 2 = −5 (correct).**

---

End of SPEC.
