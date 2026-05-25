---
name: moai-meta-harness
description: >
  Meta-harness skill that designs project-specific agent teams and generates the
  skills they use. Adapts the revfactory/harness 7-Phase workflow to MoAI's agent
  ecosystem. Triggered by /moai project Socratic interview and produces dynamic
  moai-harness-* skills + .claude/agents/harness/* + .moai/harness/* artifacts.
license: Apache-2.0
compatibility: Designed for Claude Code (v2.1.111+)
allowed-tools: Read, Write, Edit, Grep, Glob, Bash
user-invocable: false
metadata:
  version: "0.1.0"
  category: "meta"
  status: "active"
  updated: "2026-04-27"
  modularized: "false"
  tags: "meta-skill, harness, project-init, agent-team-architect, apache-2-0-attribution"
  upstream_source: "revfactory/harness"
  generated_by: "moai-adk"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["harness", "project-init", "meta-skill", "agent-team", "harness-evolve"]
  agents: ["manager-spec", "manager-strategy", "evaluator-active"]
  phases: ["plan", "run", "sync"]
---

# moai-meta-harness

<!-- @MX:NOTE: [AUTO] V3R4 contract — this skill body is preserved unchanged per SPEC-V3R4-HARNESS-001 §10 exclusion #10 (text annotation only, no behavioral change). The meta-harness 7-Phase workflow that generates project-specific moai-harness-* skills and .claude/agents/harness/* definitions is governed by REQ-HRN-FND-015 (orchestrator-only AskUserQuestion contract) — any subagent generated under .claude/agents/harness/ MUST NOT invoke AskUserQuestion; if user input is required, the subagent returns a structured blocker report and the orchestrator runs the AskUser round. Cross-reference: .claude/rules/moai/core/agent-common-protocol.md § User Interaction Boundary. -->

<!-- ATTRIBUTION
Original work: revfactory/harness (https://github.com/revfactory/harness)
License: Apache License 2.0
Adaptations: 7-Phase workflow integrated with MoAI agent ecosystem (manager-*, expert-*, evaluator-active)
NOTICE: This file contains modifications. See SPEC-V3R3-HARNESS-001 for derivation history.
-->

> **Apache 2.0 Attribution**: Adapted from [revfactory/harness](https://github.com/revfactory/harness) (Apache License 2.0). The 7-Phase workflow below is a MoAI adaptation of the upstream 6-Phase + Evolution Mechanism. See `.claude/rules/moai/NOTICE.md` for the full third-party notices and SPEC-V3R3-HARNESS-001 for derivation history.

Meta-factory skill that architects and generates project-specific agent teams. Adapts the [revfactory/harness](https://github.com/revfactory/harness) 7-Phase workflow to MoAI's agent ecosystem. Produces `moai-harness-*` skills and agent definitions tailored to each project's domain.

**Upstream**: revfactory/harness (Apache-2.0) — "A meta-skill that designs domain-specific agent teams, defines specialized agents, and generates the skills they use." (2905 stars, 420 forks, created 2026-03-26)

**Effectiveness data (design target)**: +60% avg quality score (49.5 → 79.3), 15/15 win rate, −32% variance (n=15, author-measured A/B, third-party replications pending). Source: Hwang, M. (2026). "Harness: Structured Pre-Configuration for Enhancing LLM Code Agent Output Quality." revfactory/claude-code-harness.

---

## Quick Reference

### When to Use

- `/moai project` Phase 5+ runs and detects an absent `.moai/harness/main.md`
- CLAUDE.md contains `<!-- moai:harness-start -->` markers (installed by SPEC-V3R3-PROJECT-HARNESS-001, not this skill)
- User explicitly requests harness generation for their project domain

### Key Outputs

| Artifact | Location | Owner |
|----------|----------|-------|
| Harness config | `.moai/harness/main.md` + extension files | this skill |
| Agent definitions | `.claude/agents/harness/*.md` | this skill |
| Domain skills | `.claude/skills/moai-harness-*/SKILL.md` | this skill |

All generated artifacts use the `moai-harness-*` prefix — never `moai-*`.

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
| 2. Analysis | Phase 1 domain analysis (codebase scan) | manager-spec + manager-strategy | `answers.yaml` + repo state | Analysis report |
| 3. Synthesis | Phase 2 team architecture design | manager-spec | Analysis report | SPEC doc with EARS |
| 4. Skeleton | Phase 3 agent definition generation | meta-harness (this skill) | SPEC doc | `.moai/harness/main.md` + extensions |
| 5. Customization | Phase 4 skill generation | meta-harness (this skill) | Skeleton | `.claude/agents/harness/*.md` + `.claude/skills/moai-harness-*/SKILL.md` |
| 6. Evaluation | Phase 5 integration + Phase 6 validation | evaluator-active | Generated artifacts | Sprint Contract score |
| 7. Iteration | Harness Evolution Mechanism + Phase 7-5 ops | LEARNING-001 (separate SPEC) | Scoring deltas | Factory feedback (out of scope) |

### Phase Summaries

- Phase 1 (Discovery): `manager-spec` conducts 16-question Socratic interview (owned by SPEC-V3R3-PROJECT-HARNESS-001). Output: `.moai/harness/answers.yaml`
- Phase 2 (Analysis): `manager-spec` + `manager-strategy` scan repo (file structure, existing agents/skills, dependency files, test coverage)
- Phase 3 (Synthesis): `manager-spec` produces SPEC with EARS requirements selecting one of 6 architectural patterns, defining agent roles, skill categories, acceptance criteria
- Phase 4 (Skeleton): This skill generates harness skeleton — main.md, agents.md, skills.md extensions, agent definition stubs
- Phase 5 (Customization): This skill fills the skeleton with domain-specific content referencing existing MoAI agents (manager-*, expert-*, builder-harness, evaluator-active)
- Phase 6 (Evaluation): `evaluator-active` runs Sprint Contract protocol (design constitution §11.5) — 4 dimensions, pass threshold 0.75 (FROZEN floor 0.60)
- Phase 7 (Iteration): Owned by SPEC-V3R3-HARNESS-LEARNING-001 (out of scope for this skill)

See [Phase 1-7 detailed walkthrough + agent involvement](references/seven-phase-workflow.md) for full per-phase activity, inputs, outputs, and cross-reference notes.

### MoAI Agent Cross-References

This skill orchestrates but does NOT replace existing agents. All agents referenced are static MoAI agents — no new agents are introduced. Categories: Planning & Strategy (manager-spec, manager-strategy, plan-auditor), Implementation (expert-*, manager-develop, manager-quality), Builders (builder-harness with artifact_type=agent|skill|plugin), Workflow Managers (manager-develop, manager-quality, manager-docs, manager-git), Quality (evaluator-active).

See [agent cross-references full inventory](references/agent-cross-references.md) for per-agent role and phase mapping.

### Generated Harness Validation

After Phase 5 (Customization) emits new `moai-harness-*` skills, this meta-harness automatically hands off to `evaluator-active` using the Sprint Contract protocol (design constitution §11.5).

**4-Dimension Sprint Contract Assessment**:

| Dimension | What is Checked |
|-----------|----------------|
| Functionality | Agent definitions execute their stated purpose; skills have valid trigger conditions |
| Security | No credentials in generated files; tool permissions follow least-privilege |
| Craft | YAML frontmatter valid (CSV allowed-tools, quoted metadata); progressive disclosure configured |
| Consistency | Domain alignment with `answers.yaml`; naming follows `moai-harness-*` convention |

**Scoring**:

- Pass threshold: 0.75 default (configurable via `design.yaml pass_threshold`)
- FROZEN floor: 0.60 (design constitution §2, immutable)
- Scoring rubric: evaluator-active rubric anchoring (design constitution §12, Mechanism 1)

For Phase 3b — HRN-003 Hierarchical Scoring (when `harness.yaml` sets `evaluator_mode: hierarchical`), see [HRN-003 hierarchical scoring detail](references/hrn-003-hierarchical-scoring.md).

**Design Target Reference**: The +60% effectiveness figure from Hwang (2026) — 49.5 → 79.3 in a 15-run A/B study (author-measured, third-party replications pending) — is the design intent for this validation hook. REQ-HARNESS-009 explicitly states this REQ does not require runtime measurement.

---

## Namespace Separation

[HARD] Skills + Agents namespace는 **"범용 배포"** vs **"사용자 생성"** 으로 명확히 분리된다.

### Distributed (template-managed)

`moai-*` namespace (모든 prefix 포함: `moai-foundation-*`, `moai-workflow-*`, `moai-domain-*`, `moai-ref-*`, `moai-meta-*`, `moai-harness-*`) is moai-adk distributed. `moai update` 가 sync (삭제 후 신규 설치). 사용자 직접 수정은 다음 update로 overwrite.

본 namespace의 하네스 자산:
- `moai-meta-harness` (this skill — 7-Phase generator)
- `moai-harness-learner` (lifecycle 관리 빌더, project-agnostic)

### User-Generated (this meta-harness emits)

**`my-harness-*` namespace and `.claude/agents/harness/` directory** are user-owned. Created by this meta-harness during `/moai project` Phase 5+ interview, tailored to the user's project domain.

User-generated artifacts:
- `.claude/skills/my-harness-<domain>/SKILL.md` — domain-specific skill (e.g., `my-harness-trading`, `my-harness-llm-cascade`)
- `.claude/agents/harness/<role>.md` — agent definition (e.g., `.claude/agents/harness/trading-specialist.md`)
- `.moai/harness/main.md` — harness entry point + extensions

### Contract

- [HARD] This meta-harness MUST emit user-generated skills with `my-harness-*` prefix ONLY. Emitting a `moai-*` (including `moai-harness-*`) prefixed file during Phase 4 or 5 is a **contract violation**.
- [HARD] `moai update` MUST NOT delete, modify, or sync `my-harness-*` skills or `.claude/agents/harness/*` files. Backup before update is mandatory.
- [HARD] Template (`internal/template/templates/`) MUST NOT contain `my-harness-*` skills or `.claude/agents/harness/*-specialist.md` files. Leak detection triggers cleanup chore.

### Storage Roots

| Namespace / Path | Location | Source | `moai update` 동작 |
|------------------|----------|--------|---------------------|
| `moai-*` skills (incl. `moai-harness-*` builders) | `.claude/skills/moai-*/` | template | 삭제 후 신규 설치 (overwrite) |
| **`my-harness-*` skills** | `.claude/skills/my-harness-*/` | **user project (this meta-harness emits)** | **절대 삭제/modify 금지 + 백업 보존** |
| MoAI agents (core/expert/meta) | `.claude/agents/{core,expert,meta}/` | template | 삭제 후 신규 설치 (overwrite) |
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
2. CLAUDE.md contains `<!-- moai:harness-start -->` markers. These markers are installed by SPEC-V3R3-PROJECT-HARNESS-001 during project initialization; this skill does not install them.

**Frontmatter Triggers**:

This skill loads when any of the following match:

- Keywords: `harness`, `project-init`, `meta-skill`, `agent-team`, `harness-evolve`
- Agents: `manager-spec`, `manager-strategy`, `evaluator-active`
- Phases: `plan`, `run`, `sync`

**Deferred Execution Contract**:

This skill provides the workflow recipe and agent cross-references. It does NOT execute `/moai project` Phase 5+ logic — that invocation is owned by SPEC-V3R3-PROJECT-HARNESS-001. The separation is intentional:

- This skill = capability (what to do and how)
- PROJECT-HARNESS-001 = invocation wiring (when to do it)

---

## Out of Scope

The following capabilities are explicitly NOT implemented by this skill:

- **5-layer integration mechanism** — owned by SPEC-V3R3-PROJECT-HARNESS-001. The integration with `/moai project` phases, hook installation, and CLAUDE.md marker management are all delegated to that SPEC.
- **16-question Socratic interview** — owned by SPEC-V3R3-PROJECT-HARNESS-001. The `manager-spec` conducts the interview under that SPEC's control.
- **Auto-evolution loop** — owned by SPEC-V3R3-HARNESS-LEARNING-001. The learning feedback mechanism (Phase 7) and delta capture are separate work items outside Wave A.
- **Modification of `.claude/agents/{core,expert,meta,harness}/` or static `moai-*` skills** — this meta-harness generates only `moai-harness-*` prefixed artifacts and has no write access to MoAI's own agent/skill directories.

---

## Works Well With

- `moai-foundation-core` — SPEC-First DDD and TRUST 5 quality gates
- `moai-foundation-cc` — Claude Code skill/agent authoring standards
- `manager-spec` — Conducts Discovery and Synthesis phases
- `evaluator-active` — Sprint Contract evaluation in Phase 6
- `builder-harness` (artifact_type=agent|skill|plugin) — Artifact generation helper

---

*Upstream: revfactory/harness (Apache-2.0) | MoAI adaptation: SPEC-V3R3-HARNESS-001*
*See `.claude/rules/moai/NOTICE.md` for full Apache 2.0 attribution.*
