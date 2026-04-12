---
name: moai-design-craft
description: >
  Intent-First design craft specialist covering design direction, domain vocabulary,
  design memory, and post-build critique. Use when establishing design intent or
  auditing code against design principles.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Grep, Glob
user-invocable: false
metadata:
  version: "1.2.0"
  category: "domain"
  status: "active"
  updated: "2026-03-30"
  modularized: "true"
  tags: "design, craft, intent-first, design-direction, domain-exploration, design-memory, critique, web-copy, ux-writing, headline, cta"
  related-skills: "moai-domain-uiux, moai-design-tools, moai-domain-frontend"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 4500

# MoAI Extension: Triggers
triggers:
  keywords: ["intent-first", "design craft", "design direction", "design intent", "domain exploration", "design critique", "craft review", "design memory", "design system", "system.md", "design audit", "why before what", "design extract", "interface design", "web copy", "ux writing", "headline", "cta copy", "landing page copy", "anti-ai writing"]
  agents: ["expert-frontend", "team-designer"]
  phases: ["plan", "run", "review"]
---

# Design Craft Specialist

Intent-First design philosophy integrated into MoAI workflows. Ensures design decisions flow from intent and domain understanding, not from visual impulse.

## Core Philosophy

**Intent-First**: Before any visual or component decision, establish *why* — the domain, the user, the interaction contract, and the craft principles that apply.

The three craft operations:

| Operation | When | What It Does |
|-----------|------|--------------|
| Design Direction | At `/moai plan` (design keywords) | Domain exploration, intent capture, vocabulary alignment |
| Design Audit | At `/moai review --design` | Checks implementation against `.moai/design/system.md` |
| Design Critique | At `/moai review --critique` | Post-build craft review: observe, diagnose, rebuild decision |

## Module Index

- `modules/intent-first.md` — Intent-First process: domain exploration, design direction, vocabulary
- `modules/design-memory.md` — `.moai/design/system.md` read/write protocol
- `modules/critique-workflow.md` — Post-build critique: observe → diagnose → rebuild, hard rules and rejection criteria
- `modules/web-copy-craft.md` — Web copy guidelines: anti-AI writing, headline formulas, CTA patterns, body copy rhythm

## Quick Reference

### Design Direction (plan phase)

When manager-spec detects design-relevant keywords, trigger Design Direction:

1. Read `.moai/design/system.md` (if it exists) for established vocabulary and intent
2. Explore domain: What is the user doing? What is the mental model? What does success feel like?
3. Define design intent in 1–3 sentences
4. Identify 3–5 domain vocabulary terms
5. Write design direction to `.moai/design/system.md`

### Design Audit (review phase)

When `/moai review --design` is invoked:

1. Read `.moai/design/system.md` for current design system rules
2. Scan UI components against the rules
3. Report violations with file:line references
4. Suggest minimal fixes preserving existing structure

### Design Critique (review phase)

When `/moai review --critique` is invoked:

1. Observe: What does the built interface actually do? (not what it was supposed to do)
2. Diagnose: Where does the implementation drift from intent?
3. Decide: Patch (small drift) or rebuild (fundamental misalignment)

### Web Copy Craft (run phase)

When expert-frontend or team-designer generates web pages, apply copy craft rules:

1. Use headline formulas: Number Anchor, Reversal, Direct Question, Empathy Hook, Declaration
2. Vary sentence rhythm — never three consecutive sentences with the same structure
3. Replace vague intensifiers with specific facts (numbers, names, dates)
4. Eliminate AI filler phrases ("In today's fast-paced world", "Unlock the potential")
5. CTA buttons: verb-first, outcome-oriented, one per viewport

## Works Well With

- `moai-domain-uiux` — Design tokens, WCAG, accessibility (complementary, not overlapping)
- `moai-design-tools` — Figma/Pencil tool mechanics (complementary, not overlapping)
- `moai-domain-frontend` — Component implementation patterns

---

Version: 1.2.0
Last Updated: 2026-03-30

<!-- moai:evolvable-start id="rationalizations" -->
## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "Design direction is just aesthetics, the code works the same" | Design intent drives user perception and brand consistency. Code without design direction produces generic, forgettable interfaces. |
| "I will figure out the design vocabulary during implementation" | Naming components without a design vocabulary produces inconsistent names. Establish vocabulary before building. |
| "Design memory is unnecessary, the design system is the reference" | Design systems define what to use. Design memory captures why decisions were made and what was rejected. |
| "Post-build critique is just a formality" | Critique reveals drift between intent and execution. Without it, the gap accumulates with each iteration. |
| "This is an internal tool, design does not matter" | Internal users have the same cognitive load as external users. Poor design increases training cost and error rates. |

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="red-flags" -->
## Red Flags

- Implementation diverges from stated design direction without documented reason
- Component naming inconsistent with established design vocabulary
- Design decisions made without referencing design memory or prior decisions
- No post-build critique performed after major UI implementation
- Brand voice or visual identity not consulted before UI copy changes

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="verification" -->
## Verification

- [ ] Design direction documented before implementation begins
- [ ] Component names match the established design vocabulary
- [ ] Design memory consulted for relevant prior decisions
- [ ] Post-build critique completed comparing intent vs execution
- [ ] Brand visual identity referenced in color, typography, and spacing choices

<!-- moai:evolvable-end -->
