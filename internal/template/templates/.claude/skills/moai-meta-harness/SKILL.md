---
name: moai-meta-harness
description: >
  DEPRECATED — legacy 7-Phase meta-harness. Redirects to the v4 harness Builder
  (/moai:harness <natural-language request>) which replaces the static 7-Phase
  workflow with an orchestrator-direct 4-phase Builder (ANALYZE / PLAN /
  GENERATE / ACTIVATE) + a manifest-driven dynamic-workflow Runner. Retained as
  the redirect source for backward-compat invocation paths; the 7-Phase body
  below is preserved as historical reference, NOT for new harness creation.

when_to_use: >
  Use the v4 Builder instead: issue /moai:harness <natural-language request> to
  enter Context-First Discovery (domain / goal / constraints / scope extraction)
  then the orchestrator-direct Builder. This legacy skill fires ONLY on
  backward-compat invocation paths that still reference the 7-Phase workflow; on
  any such invocation it surfaces a deprecation notice and redirects to v4.

license: Apache-2.0
compatibility: Designed for Claude Code (v2.1.111+)
allowed-tools: Read, Write, Edit, Grep, Glob, Bash
user-invocable: false
metadata:
  version: "0.2.0"
  category: "meta"
  status: "deprecated"
  updated: "2026-06-20"
  modularized: "false"
  tags: "meta-skill, harness, deprecated, v4-redirect, agent-team-architect, apache-2-0-attribution"
  upstream_source: "revfactory/harness"
  generated_by: "moai-adk"
  superseded_by: "/moai:harness v4 Builder"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000
---

# moai-meta-harness (DEPRECATED — redirect to v4)

> **DEPRECATION NOTICE** — This legacy 7-Phase meta-harness is
> **superseded by the v4 harness Builder** (`/moai:harness <natural-language
> request>`). The v4 design replaces the static 7-Phase workflow with:
>
> - An **orchestrator-direct 4-phase Builder** (ANALYZE -> PLAN -> GENERATE ->
>   ACTIVATE) that holds the plan in Claude's session context and can call
>   AskUserQuestion at the PLAN->GENERATE approval gate (first-class boundary).
> - A **manifest-driven dynamic-workflow Runner** (`harness-<name>-run.js`) that
>   reads `manifest.json` and dispatches specialists per their declared
>   `primitive` (sub-agent / dynamic-workflow / worktree / /goal /
>   adversarial-fan-out) — no heuristic re-derivation.
> - **Conditional sub-agent-granular worktree isolation** (no mandatory
>   top-level worktree; worktree only for conflict-prone parallel generation).
> - **Signal-driven phase synthesis** (not a fixed pipeline); the evaluator is
>   conditional (skipped for tasks within the model's solo reliable range).
>
> **To create a new harness**: issue `/moai:harness <natural-language request>`
> (e.g., `/moai:harness build a harness for my-project's API development`). The
> orchestrator runs Context-First Discovery on the request, derives a harness
> `<name>`, and enters the Builder. Do NOT use the 7-Phase workflow below for
> new harness creation.
>
> **What is preserved here**: the 7-Phase body below is retained as the
> **redirect source** for backward-compat — so existing invocation paths that
> reference the legacy workflow land on a deprecation notice + redirect rather
> than a dead link. The 7-Phase content is historical reference material, NOT
> active workflow. The revfactory 7-Phase residual grep excludes this body
> precisely because it IS the redirect source.

<!-- @MX:NOTE: [AUTO] V3R4 contract SUPERSEDED — the original V3R4 contract preserved this skill body unchanged per the harness foundation policy §10 exclusion #10 (text annotation only, no behavioral change). That contract is hereby EXPLICITLY SUPERSEDED by the v4 harness redesign: the body is converted to a v4 redirect. Rationale for supersession: (1) the V3R4 contract protected against behavioral change, but v4 IS a deliberate behavioral change — the 7-Phase workflow is retired in favor of the orchestrator-direct Builder + manifest-driven Runner; (2) preserving the body unchanged would leave a dead-path 7-Phase workflow that contradicts the v4 design; (3) the AskUserQuestion-only contract (REQ-HRN-FND-015) is itself preserved verbatim under v4 — any generated subagent under .claude/agents/harness/ still MUST NOT invoke AskUserQuestion (this is reaffirmed, not weakened). The supersession is narrow: the 7-Phase workflow is retired; the AskUserQuestion boundary + namespace separation + Apache-2.0 attribution are all preserved. Cross-reference: the v4 design's Migration Path (revfactory 7-Phase -> v4 mapping) in the companion harness-builder.md workflow. -->

<!-- ATTRIBUTION
Original work: revfactory/harness (https://github.com/revfactory/harness)
License: Apache License 2.0
Adaptations: 7-Phase workflow integrated with MoAI agent ecosystem (manager-*, expert-*, sync-auditor)
NOTICE: This file contains modifications. See the harness policy for derivation history. The 7-Phase workflow below is superseded by the v4 Builder; it is retained as the redirect source for backward-compat.
-->

> **Apache 2.0 Attribution**: Adapted from [revfactory/harness](https://github.com/revfactory/harness) (Apache License 2.0). The 7-Phase workflow below is a MoAI adaptation of the upstream 6-Phase + Evolution Mechanism. See `.claude/rules/moai/NOTICE.md` for the full third-party notices and the harness policy for derivation history. **The v4 Builder (the replacement) is documented in `.claude/skills/moai/workflows/harness-builder.md`.**

---

## v4 Redirect (the active path)

**New harness creation**: `/moai:harness <natural-language request>`

This routes to the v4 Builder (orchestrator-direct). The Builder:

1. **Context-First Discovery** on the request (extract domain / goal / constraints / scope).
2. **AskUserQuestion Socratic rounds** if clarity <100%; **derives** harness `<name>` from confirmed intent.
3. **Explicit approval gate** (AskUserQuestion, PLAN->GENERATE boundary).
4. **ANALYZE / PLAN / GENERATE / ACTIVATE** as orchestrator-direct phases.
5. Emits 5 artifact types: entry command, Runner Workflow, specialists, companion skills, `manifest.json`.

**Harness lifecycle** (list / edit / remove): `/moai:harness list|edit|remove <name>`.

**Harness execution**: `/harness:<name>` (auto-generated thin-wrapper command -> the harness's Runner Workflow).

See `harness-builder.md` (companion workflow under `moai/workflows/`) for the full Builder contract and `.claude/rules/moai/workflow/dynamic-workflows.md` for the Runner's dynamic-workflow primitive.

---

## Legacy 7-Phase Body (HISTORICAL REFERENCE — redirect source, NOT active)

> The content below this separator is the original 7-Phase workflow, preserved
> verbatim as the redirect source. It is NOT the active harness-creation path.
> New harness creation MUST use `/moai:harness <NL request>` (v4) above. The
> revfactory 7-Phase residual grep excludes this body because it IS the legacy
> source being redirected away from.

Meta-factory skill that architects and generates project-specific agent teams. Adapts the [revfactory/harness](https://github.com/revfactory/harness) 7-Phase workflow to MoAI's agent ecosystem. Produces `harness-*` skills and agent definitions tailored to each project's domain.

**Upstream**: revfactory/harness (Apache-2.0) — "A meta-skill that designs domain-specific agent teams, defines specialized agents, and generates the skills they use." (2905 stars, 420 forks, created 2026-03-26)

**Effectiveness data (design target)**: +60% avg quality score (49.5 → 79.3), 15/15 win rate, −32% variance (n=15, author-measured A/B, third-party replications pending). Source: Hwang, M. (2026). "Harness: Structured Pre-Configuration for Enhancing LLM Code Agent Output Quality." revfactory/claude-code-harness.

---

## Quick Reference

### When to Use

- `/moai project` Phase 5+ runs and detects an absent `.moai/harness/main.md`
- CLAUDE.md contains `<!-- moai:harness-start -->` markers (installed by the project-harness generation policy, not this skill)
- User explicitly requests harness generation for their project domain

### Key Outputs

| Artifact | Location | Owner |
|----------|----------|-------|
| Harness config | `.moai/harness/main.md` + extension files | this skill |
| Agent definitions | `.claude/agents/harness/*.md` | this skill |
| Domain skills | `.claude/skills/harness-*/SKILL.md` | this skill |

All generated artifacts use the `harness-*` prefix — never `moai-*` (which is template-managed). The `moai-harness-*` prefix specifically denotes template-managed harness builders (`moai-meta-harness`, `moai-harness-learner`) and is NOT used by this generator's emissions.

### 6 Architectural Patterns (upstream)

Pipeline, Fan-out/Fan-in, Expert Pool, Producer-Reviewer, Supervisor, Hierarchical Delegation.

See [phase walkthrough detail](references/seven-phase-workflow.md) for pattern semantics and selection guidance.

---

## Implementation Guide

### 7-Phase Workflow — Source Mapping

Each MoAI phase maps to upstream revfactory/harness phases (ref: https://github.com/revfactory/harness#workflow):

| MoAI Phase | Upstream Harness Phase | Owning Agent | Inputs | Outputs |
|------------|------------------------|--------------|--------|---------|
| 1. Discovery | Phase 0 (audit) + Phase 1 domain analysis (Socratic) | manager-spec | User request | `answers.yaml` |
| 2. Analysis | Phase 1 domain analysis (codebase scan) | manager-spec | `answers.yaml` + repo state | Analysis report |
| 3. Synthesis | Phase 2 team architecture design | manager-spec | Analysis report | SPEC doc with EARS |
| 4. Skeleton | Phase 3 agent definition generation | meta-harness (this skill) | SPEC doc | `.moai/harness/main.md` + extensions |
| 5. Customization | Phase 4 skill generation | meta-harness (this skill) | Skeleton | `.claude/agents/harness/*.md` + `.claude/skills/harness-*/SKILL.md` |
| 6. Evaluation | Phase 5 integration + Phase 6 validation | sync-auditor | Generated artifacts | Sprint Contract score |
| 7. Iteration | Harness Evolution Mechanism + Phase 7-5 ops | LEARNING-001 (separate SPEC) | Scoring deltas | Factory feedback (out of scope) |

### Phase Summaries

- Phase 1 (Discovery): `manager-spec` conducts 16-question Socratic interview (owned by the project-harness generation policy). Output: `.moai/harness/answers.yaml`
- Phase 2 (Analysis): `manager-spec` scans repo (file structure, existing agents/skills, dependency files, test coverage) — strategic analysis is absorbed into manager-spec
- Phase 3 (Synthesis): `manager-spec` produces SPEC with EARS requirements selecting one of 6 architectural patterns, defining agent roles, skill categories, acceptance criteria
- Phase 4 (Skeleton): This skill generates harness skeleton — main.md, agents.md, skills.md extensions, agent definition stubs
- Phase 5 (Customization): This skill fills the skeleton with domain-specific content referencing the retained MoAI agents (manager-*, builder-harness, sync-auditor) plus per-spawn Agent(general-purpose) domain delegations for domain-specific work
- Phase 6 (Evaluation): `sync-auditor` runs Sprint Contract protocol (design constitution §11.5) — 4 dimensions, pass threshold 0.75 (FROZEN floor 0.60)
- Phase 7 (Iteration): Owned by the harness-learning policy (out of scope for this skill)

See [Phase 1-7 detailed walkthrough + agent involvement](references/seven-phase-workflow.md) for full per-phase activity, inputs, outputs, and cross-reference notes.

### MoAI Agent Cross-References

This skill orchestrates but does NOT replace existing agents. All named agents are retained MoAI agents — no new agents are introduced; domain-specific work is delegated via per-spawn Agent(general-purpose) with domain instructions. Categories: Planning & Strategy (manager-spec, plan-auditor), Implementation (manager-develop, plus per-spawn Agent(general-purpose) domain delegations), Builders (builder-harness with artifact_type=agent|skill|plugin), Workflow Managers (manager-develop, manager-docs, manager-git; quality gate via /moai gate or the sync-phase-quality-gate.sh Stop hook), Quality (sync-auditor).

See [agent cross-references full inventory](references/agent-cross-references.md) for per-agent role and phase mapping.

### Generated Harness Validation

After Phase 5 (Customization) emits new `harness-*` skills, this meta-harness automatically hands off to `sync-auditor` using the Sprint Contract protocol (design constitution §11.5).

**4-Dimension Sprint Contract Assessment**:

| Dimension | What is Checked |
|-----------|----------------|
| Functionality | Agent definitions execute their stated purpose; skills have valid trigger conditions |
| Security | No credentials in generated files; tool permissions follow least-privilege |
| Craft | YAML frontmatter valid (CSV allowed-tools, quoted metadata); progressive disclosure configured |
| Consistency | Domain alignment with `answers.yaml`; naming follows `harness-*` convention |

**Scoring**:

- Pass threshold: 0.75 default (configurable via `design.yaml pass_threshold`)
- FROZEN floor: 0.60 (design constitution §2, immutable)
- Scoring rubric: sync-auditor rubric anchoring (design constitution §12, Mechanism 1)

For Phase 3b — HRN-003 Hierarchical Scoring (when `harness.yaml` sets `evaluator_mode: hierarchical`), see [HRN-003 hierarchical scoring detail](references/hrn-003-hierarchical-scoring.md).

**Design Target Reference**: The +60% effectiveness figure from Hwang (2026) — 49.5 → 79.3 in a 15-run A/B study (author-measured, third-party replications pending) — is the design intent for this validation hook. The governing requirement explicitly states this does not require runtime measurement.

---

## Namespace Separation

[HARD] Skills + Agents namespace는 **"범용 배포"** vs **"사용자 생성"** 으로 명확히 분리된다.

### Distributed (template-managed)

`moai-*` namespace (모든 prefix 포함: `moai-foundation-*`, `moai-workflow-*`, `moai-domain-*`, `moai-ref-*`, `moai-meta-*`, `moai-harness-*`) is moai-adk distributed. `moai update` 가 sync (삭제 후 신규 설치). 사용자 직접 수정은 다음 update로 overwrite.

본 namespace의 하네스 자산:
- `moai-meta-harness` (this skill — 7-Phase generator)
- `moai-harness-learner` (lifecycle 관리 빌더, project-agnostic)

### User-Generated (this meta-harness emits)

**`harness-*` namespace and `.claude/agents/harness/` directory** are user-owned. Created by this meta-harness during `/moai project` Phase 5+ interview, tailored to the user's project domain.

User-generated artifacts:
- `.claude/skills/harness-<domain>/SKILL.md` — domain-specific skill (e.g., `harness-trading`, `harness-llm-cascade`)
- `.claude/agents/harness/<role>.md` — agent definition (e.g., `.claude/agents/harness/trading-specialist.md`)
- `.moai/harness/main.md` — harness entry point + extensions

### Contract

- [HARD] This meta-harness MUST emit user-generated skills with `harness-*` prefix ONLY. Emitting a `moai-*` (including `moai-harness-*`) prefixed file during Phase 4 or 5 is a **contract violation**.
- [HARD] `moai update` MUST NOT delete, modify, or sync `harness-*` skills or `.claude/agents/harness/*` files. Backup before update is mandatory.
- [HARD] Template (`internal/template/templates/`) MUST NOT contain `harness-*` skills or `.claude/agents/harness/*-specialist.md` files. Leak detection triggers cleanup chore.
- [HARD] `harness-*` (user-owned) vs `moai-harness-*` (template builder) substring 구분: prefix 매칭은 정확한 startsWith 비교를 사용 (`*harness-*` substring 패턴은 false positive 위험으로 금지).
- [HARD] Generator emits `harness-*` prefix ONLY. The build enforcement recognizes `harness-*` as user-owned, with the legacy prefixed form retained during a backward-compat deprecation window. SSOT: the harness namespace separation policy.

### Generated-Agent Self-Activation Contract

[HARD] Each generated `.claude/agents/harness/<role>.md` agent MUST be emitted with both of the following frontmatter fields so the generated harness self-activates when the agent is delegated:

- A `skills:` frontmatter entry preloading the agent's companion `harness-<domain>-*` skill. This makes the domain skill load deterministically when the agent runs, rather than relying on auto-discovery which fails silently when the companion skill is absent from the agent's context.
- A non-empty, trigger-shaped `description` frontmatter field naming the domain + the observable task-shape, so the orchestrator's `.moai/harness/main.md` Task-Shape Routing table can dispatch to it.

Both fields are enforced at runtime by the Phase-6 post-generation smoke gate (`moai doctor harness`, see the `project/meta-harness.md` workflow Phase 7.3): a generated agent with an empty `description`, a dangling `skills:` reference (pointing at a non-existent `harness-*` dir), or NO `skills:` key at all causes the gate to FAIL. A `skills:`-less agent must not pass silently — that is the auto-discovery failure mode this contract closes. Full emission template + example: `project/meta-harness.md` § 6.4.1.

### Storage Roots

| Namespace / Path | Location | Source | `moai update` 동작 |
|------------------|----------|--------|---------------------|
| `moai-*` skills (incl. `moai-harness-*` builders) | `.claude/skills/moai-*/` | template | 삭제 후 신규 설치 (overwrite) |
| **`harness-*` skills** | `.claude/skills/harness-*/` | **user project (this meta-harness emits — intent declaration)** | **절대 삭제/modify 금지 + 백업 보존** |
| MoAI agents (retained 7, FLAT) | `.claude/agents/moai/` | template | 삭제 후 신규 설치 (overwrite) |
| **Generated harness agents** | `.claude/agents/harness/` | **user project (this meta-harness emits)** | **절대 삭제/modify 금지 + 백업 보존** |
| Harness config | `.moai/harness/` | user project | 절대 삭제 금지 + 백업 보존 |

### Cross-References

- `.claude/skills/moai-meta-harness/SKILL.md` § Namespace Separation (this file — canonical generator-side namespace contract)
- `.claude/rules/moai/development/skill-authoring.md` § Skills Namespace Policy
- `.claude/rules/moai/development/agent-authoring.md` § Agent Directory Convention

---

## Trigger Mechanics

**Auto-load Conditions**:

1. `/moai project` Phase 5+ runs and `.moai/harness/main.md` is absent
2. CLAUDE.md contains `<!-- moai:harness-start -->` markers. These markers are installed by the project-harness generation policy during project initialization; this skill does not install them.

**Frontmatter Triggers**:

This skill loads when any of the following match:

- Keywords: `harness`, `project-init`, `meta-skill`, `agent-team`, `harness-evolve`
- Agents: `manager-spec`, `sync-auditor`
- Phases: `plan`, `run`, `sync`

**Deferred Execution Contract**:

This skill provides the workflow recipe and agent cross-references. It does NOT execute `/moai project` Phase 5+ logic — that invocation is owned by the project-harness generation policy. The separation is intentional:

- This skill = capability (what to do and how)
- PROJECT-HARNESS-001 = invocation wiring (when to do it)

---

## Out of Scope

The following capabilities are explicitly NOT implemented by this skill:

- **5-layer integration mechanism** — owned by the project-harness generation policy. The integration with `/moai project` phases, hook installation, and CLAUDE.md marker management are all delegated to that SPEC.
- **16-question Socratic interview** — owned by the project-harness generation policy. The `manager-spec` conducts the interview under that SPEC's control.
- **Auto-evolution loop** — owned by the harness-learning policy. The learning feedback mechanism (Phase 7) and delta capture are separate work items outside Wave A.
- **Modification of `.claude/agents/{moai,harness}/` or static `moai-*` skills** — this meta-harness generates only `harness-*` prefixed artifacts and has no write access to MoAI's own agent/skill directories.

---

## Works Well With

- `moai-foundation-core` — SPEC-First DDD and TRUST 5 quality gates
- `moai-foundation-cc` — Claude Code skill/agent authoring standards
- `manager-spec` — Conducts Discovery and Synthesis phases
- `sync-auditor` — Sprint Contract evaluation in Phase 6
- `builder-harness` (artifact_type=agent|skill|plugin) — Artifact generation helper

---

*Upstream: revfactory/harness (Apache-2.0) | MoAI adaptation: the harness policy | v4 supersession: the v4 harness Builder redesign*
*See `.claude/rules/moai/NOTICE.md` for full Apache 2.0 attribution.*
