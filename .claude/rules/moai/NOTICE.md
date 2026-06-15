# MoAI-ADK Third-Party Notices

This product includes software developed by revfactory/harness and redistributed under the Apache License 2.0.

## Apache License 2.0

The following source material is licensed under Apache License 2.0:

**Source Repository**: https://github.com/revfactory/harness  
**License**: Apache License 2.0 (https://www.apache.org/licenses/LICENSE-2.0)

### Imported Components

The following reference documents from `revfactory/harness` (imported 2026-04-26) are incorporated into MoAI-ADK as pattern cookbook rules:

1. `agent-design-patterns.md` → `.claude/rules/moai/development/agent-patterns.md`
2. `qa-agent-guide.md` → `.claude/rules/moai/quality/boundary-verification.md`
3. `skill-testing-guide.md` → `.claude/rules/moai/development/skill-ab-testing.md`
4. `team-examples.md` → `.claude/rules/moai/workflow/team-pattern-cookbook.md`
5. `orchestrator-template.md` → `.claude/rules/moai/development/orchestrator-templates.md`
6. `skill-writing-guide.md` → `.claude/rules/moai/development/skill-writing-craft.md`

### Attribution

This product includes software developed by revfactory/harness contributors. The original works and any modifications are provided under the terms of the Apache License 2.0.

The imported documents have been adapted for MoAI-ADK terminology and 16-language neutrality while preserving the original technical content and design patterns. Original source authorship is retained.

### Full Apache License 2.0 Text

For the complete Apache License 2.0 text, visit: https://www.apache.org/licenses/LICENSE-2.0

---

## Karpathy Coding Principles

The following reference material is derived from Andrej Karpathy's coding philosophy:

**Source Repository**: https://github.com/forrestchang/andrej-karpathy-skills

### Imported Concepts

The following concepts from Karpathy's 4 coding principles and anti-pattern catalog (imported 2026-04-28) are incorporated into MoAI-ADK:

1. **4 Coding Principles** → `.claude/rules/moai/development/karpathy-quickref.md`
   - Think Before Coding, Simplicity First, Surgical Changes, Goal-Driven Execution
   - Mapped to MoAI's 6 Agent Core Behaviors with checkpoint questions

2. **Anti-Pattern Catalog (8 categories)** → `.claude/skills/moai/references/anti-patterns.md`
   - Premature Abstraction, Over-Engineering, Drive-By Refactoring, Style Drift
   - Silent Assumption, Guessing Over Clarifying, Sycophantic Agreement, Claiming Without Evidence
   - Adapted with Go/Python/TypeScript code examples for MoAI agent context

3. **Constitution Amendments (3 additions)** → `.claude/rules/moai/core/moai-constitution.md`
   - Behavior 4: Quantitative LOC trigger (Simplicity First)
   - Behavior 5: Style-matching directive (Surgical Changes)
   - Behavior 6: Goal-to-test pattern (Goal-Driven Execution)

### Attribution

Andrej Karpathy's coding principles are shared publicly as educational material. The `forrestchang/andrej-karpathy-skills` repository packages these principles into a structured reference. MoAI-ADK has adapted the concepts, mapped them to existing Agent Core Behaviors, and created concrete code examples specific to MoAI's orchestration context.

---

## im-not-ai (Humanize KR) — Korean AI-Tell Taxonomy

The following reference material is derived from the im-not-ai (Humanize KR) open-source skill:

**Source Repository**: https://github.com/epoko77-ai/im-not-ai
**License**: MIT License — Copyright (c) 2026 epoko77-ai

### Imported Components

The Korean AI-tell taxonomy (imported 2026-06-15) is incorporated into the `moai-domain-humanize` skill:

1. 10-category (A–J) Korean AI-tell detection taxonomy → `.claude/skills/moai-domain-humanize/modules/korean.md`
2. S1/S2/S3 severity model, A–D quality grades, and 30%/50% over-editing guardrails → shared across `.claude/skills/moai-domain-humanize/` (SKILL.md + all four language modules)

The English, Japanese, and Chinese modules of the same skill are independently web-researched catalogues modeled on this architecture, not ports of the source.

### Attribution

The im-not-ai skill is shared publicly under the MIT License. MoAI-ADK has ported the Korean taxonomy structure and adapted it for MoAI skill conventions and progressive-disclosure layout while preserving the original technical content. The MIT copyright notice is retained per the license terms.

---

**Import Date (harness)**: 2026-04-26
**Import Date (Karpathy)**: 2026-04-28
**Import Date (im-not-ai)**: 2026-06-15
**MoAI-ADK License**: MIT
**Combined Compatibility**: Apache 2.0 imports distributed under MIT with both Apache and MIT attributions preserved.

---

## Anthropic 2026 Alignment (SPEC-V3R6-AGENT-TEAM-REBUILD-001)

The MoAI agent catalog and orchestration patterns were realigned to Anthropic 2026 best practices via SPEC-V3R6-AGENT-TEAM-REBUILD-001 (plan-phase commit `b957a4d04`, run-phase milestones M1-M8). The realignment consolidated the agent catalog from 17 entries to 8 retained agents (7 MoAI-custom + 1 Anthropic built-in `Explore`) and archived 12 phantom/domain-expert agents.

### Audit 3 Findings A1-A6 (verbatim Anthropic sources cited in spec.md §B.1)

The architectural pivot was grounded in 6 verbatim findings from Anthropic's official documentation (deep SRP audit conducted 2026-05-25 via 3 parallel audit agents: 17-agent SRP audit + workflow agent-to-phase ownership audit + Anthropic 2026 verbatim citation audit):

1. **Finding A1 — Subagent spawning ceiling**: *"Subagents cannot spawn other subagents."* (Source: https://claude.com/docs/en/sub-agents)
   - Implication: The previous MoAI architecture's `manager-strategy → manager-develop` hierarchical chain was architecturally impossible. Resolution: collapse the chain by routing `/moai run` directly to `manager-develop`.

2. **Finding A2 — Agent Teams team-size ceiling**: *"Start with 3-5 teammates for most workflows. This balances parallel work with manageable coordination."* (Source: https://claude.com/docs/en/agent-teams)
   - Implication: MoAI's 17-agent catalog exceeded Anthropic's recommended 3-5 ceiling by 2-5×. Resolution: archive 12 agents and retain 8.

3. **Finding A3 — Subagent definition discipline**: *"Define a custom subagent when you keep spawning the same kind of worker."* (Source: https://claude.com/docs/en/best-practices)
   - Implication: 12 of 17 MoAI agents had 0 invocations across the 4 most recent SPECs (phantom agents). Resolution: archive phantom agents to `.moai/backups/agent-archive-2026-05-25/`.

4. **Finding A4 — Coding-task parallelism caveat**: *"most coding tasks involve fewer truly parallelizable tasks than research, and LLM agents are not yet great at coordinating and delegating to other agents in real time."* (Source: https://anthropic.com/engineering/built-multi-agent-research-system)
   - Implication: Parallel multi-spawn (Mode 4) is preferred for research-heavy work but not for coding-heavy work. Resolution: orchestration-mode-selection.md Mode 5 (sequential sub-agent) is the default fallback for coding tasks.

5. **Finding A5 — Hook event vocabulary**: Stop, PostToolUse, SubagentStop, TaskCompleted hook events are first-class observability surfaces. (Source: https://claude.com/docs/en/hooks)
   - Implication: 3 NEW hook scripts authored at M4 (PostToolUse Status Transition + Stop sync-phase quality gate + TaskCompleted team-mode) integrate with this vocabulary.

6. **Finding A6 — Opus 4.7 Adaptive Thinking**: Opus 4.7 introduces Adaptive Thinking that dynamically allocates reasoning tokens based on task complexity, replacing fixed `budget_tokens` from older models. (Source: https://platform.claude.com/docs/en/about-claude/models/whats-new-claude-4-7)
   - Implication: MoAI agent prompts removed fixed thinking budgets; `ultrathink` keyword is the canonical deep-reasoning trigger.

### Archive Summary (M3 milestone, 2026-05-25)

- **Archive count**: 11 actual archived + 1 originally absent (`researcher.md` — never present as a MoAI file in this repo)
- **Archive location**: `.moai/backups/agent-archive-2026-05-25/` (preserves `core/`, `meta/`, `expert/` substructure with per-agent README)
- **Archived agents (11 actual + 1 absent)**: `manager-strategy`, `manager-quality`, `manager-brain`, `manager-project`, `claude-code-guide`, `researcher` (originally absent), `expert-backend`, `expert-frontend`, `expert-security`, `expert-devops`, `expert-performance`, `expert-refactoring`
- **Retained agents (8 total)**: `manager-spec`, `manager-develop`, `manager-docs`, `manager-git`, `plan-auditor`, `sync-auditor`, `builder-harness`, plus Anthropic built-in `Explore`

### Migration Guidance

When a paste-ready resume message or `Agent()` invocation references one of the 12 archived agents, the MoAI orchestrator rejects the spawn per `.claude/rules/moai/workflow/archived-agent-rejection.md` and consults the per-archived-agent migration table for the retained-agent replacement pattern.

### Attribution

Anthropic Claude Code documentation is publicly available at https://claude.com/docs/en/. The verbatim citations in Findings A1-A6 are reproduced under fair-use academic-attribution conventions; no source code is incorporated. MoAI-ADK's agent catalog realignment is an independent implementation derived from analysis of Anthropic's published guidance.

**Import Date (Anthropic 2026 verbatim citations)**: 2026-05-25
**SPEC Reference**: `.moai/specs/SPEC-V3R6-AGENT-TEAM-REBUILD-001/` (5-artifact set: spec.md + plan.md + acceptance.md + design.md + research.md; plan-phase commit `b957a4d04`)
