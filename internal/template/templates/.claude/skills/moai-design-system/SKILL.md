---
name: moai-design-system
description: >
  Unified design system specialist integrating Intent-First design craft and UI/UX foundations
  (accessibility, design tokens, component architecture). Use when establishing design intent
  and building design systems.

when_to_use: >
  Use for design-system authoring and audits: Intent-First craft,
  WCAG/ARIA accessibility, design tokens, theming and dark mode, component
  libraries (shadcn, Radix UI, Storybook), Style Dictionary, icon sets
  (Lucide, Iconify, Hugeicons), UX writing, and responsive UI
  implementation.

license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Write, Edit, Grep, Glob, mcp__context7__resolve-library-id, mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "1.0.0"
  category: "domain"
  status: "active"
  updated: "2026-04-25"
  modularized: "false"
  tags: "design, craft, intent-first, design-system, accessibility, WCAG, design-tokens, theming, pencil, shadcn, Radix, icons, uiux"
  related-skills: "moai-design-tools, moai-domain-uiux, moai-design-craft, moai-domain-frontend"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000
---

# MoAI Design System Specialist

Unified design expertise covering two domains: Intent-First craft (design direction and critique) and UI/UX foundations (accessibility, tokens, components).

---

## Design Craft (absorbed from moai-design-craft)

### Core Philosophy: Intent-First

Before any visual or component decision, establish *why* — the domain, the user, the interaction contract, and the craft principles that apply.

**Design Direction Workflow:**
1. Domain Exploration: Name the domain, list key entities, establish vocabulary
2. Design Memory: Load `.moai/design/system.md` for existing design decisions
3. Intent Statement: Write "This interface should feel [adjective] because [reason]"
4. Principle Derivation: Extract 3-5 craft principles from the intent statement
5. Apply and Critique: Evaluate implementation against each principle

**Post-Build Critique Checklist:**
- Does the visual hierarchy match the domain vocabulary?
- Does every animation serve an interaction contract?
- Is web copy free of AI-generic phrases ("elevate", "seamless", "leverage")?
- Does the information density match the user's cognitive load?

**Anti-AI Slop Writing Rules for UI Copy:**
- Replace "Unlock" → use specific action ("Publish", "Deploy", "Launch")
- Replace "Seamless" → describe the actual experience ("One-step setup")
- Replace "Leverage" → use direct verbs ("Use", "Apply", "Run")
- Headlines must state what happens, not how it feels

---

## UI/UX Foundations (absorbed from moai-domain-uiux)

### Design Token Architecture

W3C DTCG 2025.10 standard. Use Style Dictionary 4.0 for multi-platform output.

Token hierarchy: `{category}.{concept}.{property}.{variant}.{state}`
- Primitive tokens: raw values (colors, spacing scales)
- Semantic tokens: context-bound references (`color.surface.primary`)
- Component tokens: component-specific mappings (`button.background.default`)

CSS custom properties for runtime theming. Dark mode via `[data-theme="dark"]` selector.

### Component Architecture

Atomic Design hierarchy: Atoms → Molecules → Organisms → Templates → Pages.
Use shadcn/ui with Radix UI primitives for accessible, unstyled base components.
Extend with Tailwind CSS utility classes. Never override Radix accessibility attributes.

### Accessibility Standards

WCAG 2.2 AA minimum. AAA for text-heavy and public-facing interfaces.

Required patterns:
- All interactive elements: keyboard navigable, visible focus ring
- Images: descriptive `alt` text or `aria-hidden="true"` for decorative
- Forms: label association via `htmlFor` or `aria-labelledby`
- Color contrast: 4.5:1 for normal text, 3:1 for large text and UI components
- Reduced motion: `prefers-reduced-motion` media query for animations

### Icon Library Selection

| Library | Count | Use Case |
|---------|-------|----------|
| Lucide | 1500+ | Default, clean, consistent stroke |
| Hugeicons | 27K+ | Rich product iconography |
| Iconify | 200K+ | Multi-library access layer |
| Tabler | 5900+ | Technical / developer tools |

<!-- moai:evolvable-start id="verification" -->
## Verification

- [ ] Design intent documented before any visual implementation begins
- [ ] Design tokens follow W3C DTCG hierarchy (primitive → semantic → component)
- [ ] All interactive elements pass WCAG 2.2 keyboard navigation test
- [ ] Color contrast ratios verified (4.5:1 normal text, 3:1 UI components)
- [ ] Web copy free of AI-generic filler phrases

<!-- moai:evolvable-end -->
