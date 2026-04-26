---
name: moai-meta-harness
description: >
  Meta-harness skill that designs project-specific agent teams and generates the
  skills they use. Adapts the revfactory/harness 7-Phase workflow to MoAI's agent
  ecosystem. Triggered by /moai project Socratic interview and produces dynamic
  my-harness-* skills + .claude/agents/my-harness/* + .moai/harness/* artifacts.
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

<!-- ATTRIBUTION
Original work: revfactory/harness (https://github.com/revfactory/harness)
License: Apache License 2.0
Adaptations: 7-Phase workflow integrated with MoAI agent ecosystem (manager-*, expert-*, evaluator-active)
NOTICE: This file contains modifications. See SPEC-V3R3-HARNESS-001 for derivation history.
-->

> **Apache 2.0 Attribution**: Adapted from [revfactory/harness](https://github.com/revfactory/harness) (Apache License 2.0). The 7-Phase workflow below is a MoAI adaptation of the upstream 6-Phase + Evolution Mechanism. See `.claude/rules/moai/NOTICE.md` for the full third-party notices and SPEC-V3R3-HARNESS-001 for derivation history.

Meta-factory skill that architects and generates project-specific agent teams.
Adapts the [revfactory/harness](https://github.com/revfactory/harness) 7-Phase
workflow to MoAI's agent ecosystem. Produces `my-harness-*` skills and agent
definitions tailored to each project's domain.

**Upstream**: revfactory/harness (Apache-2.0) — "A meta-skill that designs
domain-specific agent teams, defines specialized agents, and generates the skills
they use." (2905 stars, 420 forks, created 2026-03-26)

**Effectiveness data (design target)**: +60% avg quality score (49.5 → 79.3),
15/15 win rate, −32% variance (n=15, author-measured A/B, third-party
replications pending). Source: Hwang, M. (2026). "Harness: Structured
Pre-Configuration for Enhancing LLM Code Agent Output Quality."
revfactory/claude-code-harness.

---

## Quick Reference

### When to Use

- `/moai project` Phase 5+ runs and detects an absent `.moai/harness/main.md`
- CLAUDE.md contains `<!-- moai:harness-start -->` markers (installed by
  SPEC-V3R3-PROJECT-HARNESS-001, not this skill)
- User explicitly requests harness generation for their project domain

### Key Outputs

| Artifact | Location | Owner |
|----------|----------|-------|
| Harness config | `.moai/harness/main.md` + extension files | this skill |
| Agent definitions | `.claude/agents/my-harness/*.md` | this skill |
| Domain skills | `.claude/skills/my-harness-*/SKILL.md` | this skill |

All generated artifacts use the `my-harness-*` prefix — never `moai-*`.

### 6 Architectural Patterns (upstream)

Pipeline, Fan-out/Fan-in, Expert Pool, Producer-Reviewer, Supervisor,
Hierarchical Delegation. Details in modules/seven-phase-workflow.md.

---

## Implementation Guide

### 7-Phase Workflow

#### Source Mapping

Each MoAI phase maps to upstream revfactory/harness phases
(ref: https://github.com/revfactory/harness#workflow):

| MoAI Phase | Upstream Harness Phase | Owning Agent | Inputs | Outputs |
|------------|------------------------|--------------|--------|---------|
| 1. Discovery | Phase 0 (audit) + Phase 1 domain analysis (Socratic) | manager-spec | User request | `answers.yaml` |
| 2. Analysis | Phase 1 domain analysis (codebase scan) | manager-spec + manager-strategy | `answers.yaml` + repo state | Analysis report |
| 3. Synthesis | Phase 2 team architecture design | manager-spec | Analysis report | SPEC doc with EARS |
| 4. Skeleton | Phase 3 agent definition generation | meta-harness (this skill) | SPEC doc | `.moai/harness/main.md` + extensions |
| 5. Customization | Phase 4 skill generation | meta-harness (this skill) | Skeleton | `.claude/agents/my-harness/*.md` + `.claude/skills/my-harness-*/SKILL.md` |
| 6. Evaluation | Phase 5 integration + Phase 6 validation | evaluator-active | Generated artifacts | Sprint Contract score |
| 7. Iteration | Harness Evolution Mechanism + Phase 7-5 ops | LEARNING-001 (separate SPEC) | Scoring deltas | Factory feedback (out of scope) |

#### Phase 1 — Discovery

`manager-spec` conducts a 16-question Socratic interview (owned by
SPEC-V3R3-PROJECT-HARNESS-001). The interview surfaces:
- Project domain (e.g., fintech, e-commerce, IoT)
- Primary languages and frameworks
- Team size and expertise level
- Quality, security, and performance priorities

Output: `answers.yaml` written to `.moai/harness/answers.yaml`.

#### Phase 2 — Analysis

`manager-spec` and `manager-strategy` scan the repository:
- File structure patterns → infer domain boundaries
- Existing agents/skills → avoid duplication
- `go.mod`, `package.json`, `requirements.txt` → detect stack
- Test coverage baseline → set quality targets

Output: Analysis report passed to Phase 3.

#### Phase 3 — Synthesis

`manager-spec` synthesizes the analysis into a SPEC document with EARS
requirements. The SPEC specifies:
- Which of the 6 architectural patterns fits the domain
- Agent roles and their tool permissions
- Skill categories to generate
- Acceptance criteria for Phase 6 evaluation

#### Phase 4 — Skeleton

This skill (`moai-meta-harness`) generates the harness skeleton:

1. Read `answers.yaml` and the SPEC document
2. Write `.moai/harness/main.md` — the harness entry point
3. Write extension files: `.moai/harness/agents.md`, `.moai/harness/skills.md`
4. Create agent definition stubs in `.claude/agents/my-harness/`

Agents involved: `builder-agent`, `builder-skill` for artifact generation.

#### Phase 5 — Customization

This skill fills the skeleton with domain-specific content:

1. Generate agent definitions (`.claude/agents/my-harness/*.md`) referencing
   existing MoAI agents: `manager-spec`, `manager-strategy`, `manager-tdd`,
   `manager-ddd`, `manager-quality`, `manager-docs`, `manager-git`,
   `expert-backend`, `expert-frontend`, `expert-debug`, `expert-testing`,
   `expert-security`, `expert-refactoring`, `expert-performance`, `expert-devops`,
   `expert-mobile`, `builder-agent`, `builder-skill`, `builder-plugin`,
   `evaluator-active`, `plan-auditor`.
2. Generate domain skills (`.claude/skills/my-harness-*/SKILL.md`) following
   the skill-authoring.md schema with `my-harness-*` prefix.
3. All artifacts are user-owned and never overwritten by `moai update`.

#### Phase 6 — Evaluation

`evaluator-active` runs the Sprint Contract protocol
(design constitution §11.5) against generated artifacts:
- Functionality: Do agents execute their stated purpose?
- Security: No credential leaks, safe tool permissions?
- Craft: YAML frontmatter valid, CSV allowed-tools, progressive disclosure?
- Consistency: Brand/domain alignment, naming conventions followed?

Pass threshold: 0.75 (configurable via `design.yaml pass_threshold`;
FROZEN floor: 0.60 per design constitution §2).

#### Phase 7 — Iteration

Owned by SPEC-V3R3-HARNESS-LEARNING-001. Out of scope for this skill.
The evolution mechanism captures scoring deltas from Phase 6 and feeds
them back to Phase 4/5 on next harness run.

---

### MoAI Agent Cross-References

This skill orchestrates but does NOT replace existing agents. All agents
referenced below are static MoAI agents — no new agents are introduced.

**Planning & Strategy**

- `manager-spec` — Discovery (Phase 1) and Synthesis (Phase 3)
- `manager-strategy` — Analysis codebase scan (Phase 2)
- `plan-auditor` — EARS compliance check on the SPEC produced in Phase 3

**Implementation**

- `expert-backend` — Backend domain harness templates
- `expert-frontend` — Frontend domain harness templates
- `expert-mobile` — Mobile domain harness templates
- `expert-devops` — DevOps/platform domain harness templates
- `expert-security` — Security review of generated permissions
- `expert-testing` — Test harness pattern generation
- `expert-debug` — Debug agent patterns
- `expert-refactoring` — Refactoring workflow patterns
- `expert-performance` — Performance profiling patterns

**Builders**

- `builder-agent` — Generates `.claude/agents/my-harness/*.md` content
- `builder-skill` — Generates `.claude/skills/my-harness-*/SKILL.md` content
- `builder-plugin` — Optional plugin bundling of generated artifacts

**Workflow Managers**

- `manager-ddd` — DDD-flavored harness workflow templates
- `manager-tdd` — TDD-flavored harness workflow templates
- `manager-quality` — Quality gate configuration in generated harnesses
- `manager-docs` — Documentation generation patterns
- `manager-git` — Git workflow patterns for generated harnesses

**Quality**

- `evaluator-active` — Sprint Contract evaluation (Phase 6)

---

### Generated Harness Validation

After Phase 5 (Customization) emits new `my-harness-*` skills, this
meta-harness automatically hands off to `evaluator-active` using the
Sprint Contract protocol (design constitution §11.5).

**4-Dimension Sprint Contract Assessment**

| Dimension | What is Checked |
|-----------|----------------|
| Functionality | Agent definitions execute their stated purpose; skills have valid trigger conditions |
| Security | No credentials in generated files; tool permissions follow least-privilege |
| Craft | YAML frontmatter valid (CSV allowed-tools, quoted metadata); progressive disclosure configured |
| Consistency | Domain alignment with `answers.yaml`; naming follows `my-harness-*` convention |

**Scoring**

- Pass threshold: 0.75 default (configurable via `design.yaml pass_threshold`)
- FROZEN floor: 0.60 (design constitution §2, immutable)
- Scoring rubric: evaluator-active rubric anchoring (design constitution §12, Mechanism 1)

**Design Target Reference**

The +60% effectiveness figure from Hwang (2026) — average quality improvement
from 49.5 → 79.3 in a 15-run A/B study (author-measured, third-party
replications pending) — is the **design intent** for this validation hook.
REQ-HARNESS-009 explicitly states this REQ does not require runtime measurement;
the evaluator profile references the +60% target as the design goal that
motivated the Sprint Contract integration.

---

### Namespace Separation

`moai-*` namespace is reserved for MoAI-maintained skills managed by `moai update`.
The only `moai-*` skill in this system is `moai-meta-harness` (this file).

`my-harness-*` namespace is user-owned. All artifacts generated at runtime
by this meta-harness use the `my-harness-*` prefix:

- `.claude/skills/my-harness-<domain>/SKILL.md` — domain-specific skill
- `.claude/agents/my-harness/<role>.md` — agent definition
- `.moai/harness/main.md` — harness entry point

**Contract**

- `moai update` MUST NOT overwrite `my-harness-*` artifacts.
- This meta-harness MUST NOT emit any artifact with a `moai-` prefix at runtime.
- Emitting a `moai-` prefixed file during Phase 4 or 5 is a contract violation.

**Storage Roots**

| Namespace | Location | Managed by |
|-----------|----------|------------|
| `moai-*` skills | `.claude/skills/moai-*/` | `moai update` |
| `my-harness-*` skills | `.claude/skills/my-harness-*/` | User (this meta-harness) |
| `my-harness-*` agents | `.claude/agents/my-harness/` | User (this meta-harness) |
| Harness config | `.moai/harness/` | User (this meta-harness) |

---

### Trigger Mechanics

**Auto-load Conditions**

1. `/moai project` Phase 5+ runs and `.moai/harness/main.md` is absent.
2. CLAUDE.md contains `<!-- moai:harness-start -->` markers. These markers
   are installed by SPEC-V3R3-PROJECT-HARNESS-001 during project initialization;
   this skill does not install them.

**Frontmatter Triggers**

This skill loads when any of the following match:
- Keywords: `harness`, `project-init`, `meta-skill`, `agent-team`, `harness-evolve`
- Agents: `manager-spec`, `manager-strategy`, `evaluator-active`
- Phases: `plan`, `run`, `sync`

**Deferred Execution Contract**

This skill provides the workflow recipe and agent cross-references.
It does NOT execute `/moai project` Phase 5+ logic — that invocation
is owned by SPEC-V3R3-PROJECT-HARNESS-001. The separation is intentional:
- This skill = capability (what to do and how)
- PROJECT-HARNESS-001 = invocation wiring (when to do it)

---

### Out of Scope

The following capabilities are explicitly NOT implemented by this skill:

- **5-layer integration mechanism** — owned by SPEC-V3R3-PROJECT-HARNESS-001.
  The integration with `/moai project` phases, hook installation, and
  CLAUDE.md marker management are all delegated to that SPEC.
- **16-question Socratic interview** — owned by SPEC-V3R3-PROJECT-HARNESS-001.
  The `manager-spec` conducts the interview under that SPEC's control.
- **Auto-evolution loop** — owned by SPEC-V3R3-HARNESS-LEARNING-001.
  The learning feedback mechanism (Phase 7) and delta capture are
  separate work items outside Wave A.
- **Modification of `.claude/agents/moai/` or static `moai-*` skills** —
  this meta-harness generates only `my-harness-*` prefixed artifacts
  and has no write access to MoAI's own agent/skill directories.

---

## Works Well With

- `moai-foundation-core` — SPEC-First DDD and TRUST 5 quality gates
- `moai-foundation-cc` — Claude Code skill/agent authoring standards
- `manager-spec` — Conducts Discovery and Synthesis phases
- `evaluator-active` — Sprint Contract evaluation in Phase 6
- `builder-agent` / `builder-skill` — Artifact generation helpers

---

*Upstream: revfactory/harness (Apache-2.0) | MoAI adaptation: SPEC-V3R3-HARNESS-001*
*See `.claude/rules/moai/NOTICE.md` for full Apache 2.0 attribution.*
