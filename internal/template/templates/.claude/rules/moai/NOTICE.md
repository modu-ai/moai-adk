---
paths: "**/NOTICE.md"
---

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
4. `team-examples.md` → (retired — the derived team-pattern cookbook rule was removed in the Agent Teams static-layer retirement; no longer distributed)
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

## Anthropic 2026 Alignment

Anthropic Claude Code documentation — fair-use academic attribution. The agent catalog was realigned to Anthropic 2026 best practices (8 retained agents at consolidation time; now 10 per CLAUDE.md §4). Realignment details: the agent-catalog realignment SPEC.
