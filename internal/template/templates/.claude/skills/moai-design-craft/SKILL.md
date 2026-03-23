---
name: moai-design-craft
description: >
  Intent-First design craft specialist covering design direction, domain vocabulary,
  design memory, and post-build critique. Use when establishing design intent before
  implementation, extracting design patterns from existing code, auditing code against
  a design system, or performing a post-build craft review and rebuild.
  Do NOT use for design token management (use moai-domain-uiux) or Figma/Pencil tool
  mechanics (use moai-design-tools).
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Grep, Glob
user-invocable: false
metadata:
  version: "1.0.0"
  category: "domain"
  status: "active"
  updated: "2026-03-23"
  modularized: "true"
  tags: "design, craft, intent-first, design-direction, domain-exploration, design-memory, critique"
  related-skills: "moai-domain-uiux, moai-design-tools, moai-domain-frontend"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 4500

# MoAI Extension: Triggers
triggers:
  keywords: ["intent-first", "design craft", "design direction", "design intent", "domain exploration", "design critique", "craft review", "design memory", "design system", "system.md", "design audit", "why before what", "design extract", "interface design"]
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
- `modules/critique-workflow.md` — Post-build critique: observe → diagnose → rebuild

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

## Works Well With

- `moai-domain-uiux` — Design tokens, WCAG, accessibility (complementary, not overlapping)
- `moai-design-tools` — Figma/Pencil tool mechanics (complementary, not overlapping)
- `moai-domain-frontend` — Component implementation patterns

---

Version: 1.0.0
Last Updated: 2026-03-23
