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

## design-dna — Reference-Design Deconstruction Taxonomy

The following reference material is derived from the design-dna open-source skill:

**Source Repository**: https://github.com/zanwei/design-dna
**License**: MIT License — Copyright (c) 2026 the design-dna authors

### Imported Components

The three-dimension Design DNA taxonomy (imported 2026-08-19) is incorporated into the `moai-domain-design-dna` skill:

1. The `design_system` / `design_style` / `visual_effects` dimension split and its field groups + enum vocabularies -> `.claude/skills/moai-domain-design-dna/references/dna-schema.md`
2. The extraction rules (dominance-based colour roles, relative radius measurement, multi-reference conflict resolution, `enabled: false` non-invention gating) and the performance-tier technology map -> `.claude/skills/moai-domain-design-dna/SKILL.md` + `references/effects-implementation.md`

### Attribution

The design-dna skill is shared publicly under the MIT License. MoAI-ADK has distilled the taxonomy and its extraction/generation rules, rewritten the prose for MoAI skill conventions and progressive-disclosure layout, and cross-referenced overlapping material to its existing owners rather than restating it. The MIT copyright notice is retained per the license terms.

---

## motion-design — Vendor-Neutral Motion Principles

The following reference material is derived from the motion-design open-source skill:

**Source Repository**: https://github.com/LottieFiles/motion-design-skill
**License**: MIT License — Copyright (c) 2025 LottieFiles

### Imported Components

The vendor-neutral motion-principles taxonomy (imported 2026-08-19) is incorporated into the `moai-ref-ui-polish` skill:

1. The Three Pillars decision passes, the Three Motion Layers amplitude/offset model, the two 1/3 Rules, the attention budget, and the stagger-budget table -> `.claude/skills/moai-ref-ui-polish/references/motion-principles.md`
2. The four motion-personality archetypes, the UI-adapted reading of Disney's 12 animation principles with their numeric ranges, and the emotion-to-motion / path-as-language maps -> the same reference file

### Attribution

The motion-design skill is shared publicly under the MIT License. The source contains no Lottie or LottieFiles tooling — it is purely implementation-agnostic motion principle. MoAI-ADK has distilled that principle layer into a single progressive-disclosure reference file, rewritten the prose for MoAI skill conventions, dropped the publisher's branding and install framing, and cross-referenced the overlapping implementation rules to their existing owner (`moai-ref-ui-polish/SKILL.md` § Motion) rather than restating them. The MIT copyright notice is retained per the license terms.

---

## gsap-skills — Reduced-Motion and Compositor-Cost Rules

The following reference material is derived from the gsap-skills open-source skill set:

**Source Repository**: https://github.com/greensock/gsap-skills
**License**: MIT License — Copyright (c) 2026 GreenSock

### Imported Components

Two library-independent rules (imported 2026-08-19) are incorporated into the `moai-ref-ui-polish` skill:

1. The `prefers-reduced-motion` accessibility branch as a required authoring step for every non-decorative animation -> `.claude/skills/moai-ref-ui-polish/SKILL.md` § Motion Accessibility and Cost
2. The compositor-versus-layout framing for why `transform` and `opacity` outperform layout-triggering properties -> the same section

### Attribution

The gsap-skills set is shared publicly under the MIT License. MoAI-ADK imported only the two rules that hold independently of any animation library, restated in platform-neutral terms. No GSAP API surface, framework-lifecycle guidance, membership or licensing material, or vendor-recommendation trigger was carried over — `moai-ref-ui-polish` remains vendor-neutral and names no animation library as a default. The MIT copyright notice is retained per the license terms.

---

## genjutsu — Design-Audit Detection Suite

The following reference material is derived from the genjutsu open-source skill plugin:

**Source Repository**: https://github.com/AThevon/genjutsu
**License**: MIT License — Copyright (c) 2026 Adrien Thevon

### Imported Components

The design-audit detection suite and two hover-doctrine rules (imported 2026-08-19) are incorporated into the `moai-ref-ui-polish` skill:

1. The grep-based audit patterns — motion-gap detection, the three-stack reduced-motion probe, the accessibility and layout-property checks, and the duration/easing inventory method with its 3-5-value system budget -> `.claude/skills/moai-ref-ui-polish/references/design-audit.md`
2. The mobile no-hover doctrine (gate hover behind a pointer media query) and its desktop inverse (hover is a required affordance on pointer devices) -> `.claude/skills/moai-ref-ui-polish/SKILL.md` § Interaction

### Attribution

The genjutsu plugin is shared publicly under the MIT License. MoAI-ADK imported only the detection layer and the two hover rules, adapting them for MoAI conventions: the hardcoded source root was parameterized, per-pattern signal quality was added so a match is treated as a candidate rather than a defect, the source's own three-tier severity vocabulary was mapped onto the skill's existing HIGH/MEDIUM/LOW scale rather than introducing a second one, and volatile published library sizes were replaced by a measure-it instruction. Not imported: the separately vendored third-party `ui-ux-pro-max` dataset and its Python CLI, the framework-specific API cheat sheets, and the plugin's persona scaffolding. The MIT copyright notice is retained per the license terms.

---

**Import Date (harness)**: 2026-04-26
**Import Date (Karpathy)**: 2026-04-28
**Import Date (im-not-ai)**: 2026-06-15
**Import Date (design-dna)**: 2026-08-19
**Import Date (motion-design)**: 2026-08-19
**Import Date (gsap-skills)**: 2026-08-19
**Import Date (genjutsu)**: 2026-08-19
**MoAI-ADK License**: MIT
**Combined Compatibility**: Apache 2.0 imports distributed under MIT with both Apache and MIT attributions preserved.

---

## Anthropic 2026 Alignment

Anthropic Claude Code documentation — fair-use academic attribution. The agent catalog was realigned to Anthropic 2026 best practices (8 retained agents at consolidation time; now 10 per CLAUDE.md §4). Realignment details: the agent-catalog realignment SPEC.
