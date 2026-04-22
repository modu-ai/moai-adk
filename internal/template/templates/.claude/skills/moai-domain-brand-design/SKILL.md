---
name: moai-domain-brand-design
description: >
  Brand-aligned visual design system specialist for web projects. Enforces
  hero-first layout chaining, WCAG 2.1 AA accessibility, Lighthouse >= 80,
  and design token extraction from brand identity files. Covers color palettes,
  typography, spacing, and component specifications.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Write, Edit, Grep, Glob
user-invocable: false
metadata:
  version: "1.0.0"
  category: "domain"
  status: "active"
  updated: "2026-04-20"
  tags: "design, brand, visual-identity, design-tokens, wcag, typography, color-palette, hero-section"
  related-skills: "moai-domain-copywriting, moai-workflow-gan-loop, moai-domain-uiux"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["design-tokens", "color-palette", "typography", "hero-section", "wcag", "visual-identity", "design system", "brand design", "spacing system", "component spec"]
  agents: ["expert-frontend"]
  phases: ["run"]
---

# moai-domain-brand-design

Visual design system skill for brand-aligned web projects. Absorbed from agency-design-system (v1.0.0). Enforces hero-first chaining, WCAG 2.1 AA contrast, and structured design token output for downstream implementation.

---

## Quick Reference

### Entry Conditions

Before generating design output, verify:

1. `.moai/project/brand/visual-identity.md` exists and contains no `_TBD_` markers.
2. Copy scope is defined (from `moai-domain-copywriting` JSON output or inline brief).
3. Target framework is confirmed (from `.moai/config/sections/design.yaml` `default_framework`).

If `visual-identity.md` has unresolved `_TBD_` markers, stop and request brand interview completion.

If the defined color palette conflicts with generated design tokens, execution halts and a conflict report is returned (see Error Handling below).

### Figma Integration

Figma integration is disabled by default. Check `.moai/config/sections/design.yaml`:

```
figma:
  enabled: false
```

If `figma.enabled: true` and a public Figma file URL is provided, extract design tokens from the Figma file. Otherwise use `visual-identity.md` as the sole source of truth.

---

## Implementation Guide

### Hero-First Chaining

The hero section establishes the visual tone for the entire site. All subsequent sections chain from it:

1. Extract hero background color, typography, and spacing from `visual-identity.md`.
2. Derive complementary section colors using the established contrast ratio rules.
3. Apply consistent spacing scale across all sections (do not reset per section).
4. Navigation and footer inherit hero's typographic scale.

Hero section requirements:
- CTA button is visible above the fold on mobile (375px viewport, 667px height).
- Headline contrast ratio against background: minimum 4.5:1 (WCAG AA).
- Hero image or background: never pure white (#FFFFFF) unless brand explicitly specifies.

### Design Token Extraction

Extract and output the following token categories from `visual-identity.md`:

**Color tokens**:
- `color.primary`: Brand primary color (hex or OKLCH)
- `color.primary.foreground`: Text on primary background (must pass 4.5:1 contrast)
- `color.secondary`: Secondary brand color
- `color.accent`: Call-to-action and highlight color
- `color.neutral.*`: Scale from 50 to 950 (gray shades)
- `color.semantic.success`, `color.semantic.warning`, `color.semantic.error`: Status colors
- `color.background`: Page background
- `color.surface`: Card and component background

**Typography tokens**:
- `font.family.sans`: Primary sans-serif stack
- `font.family.mono`: Code and technical content
- `font.size.*`: Scale: xs (12px), sm (14px), base (16px), lg (18px), xl (20px), 2xl (24px), 3xl (30px), 4xl (36px)
- `font.weight.normal`, `font.weight.medium`, `font.weight.bold`, `font.weight.black`
- `line.height.tight` (1.25), `line.height.normal` (1.5), `line.height.relaxed` (1.75)

**Spacing tokens**:
- Base unit: 4px
- Scale: `space.1` (4px) through `space.24` (96px), following 4px grid
- `space.section`: Vertical section padding (default 80px desktop, 48px mobile)
- `space.container.max`: Maximum content width (default 1280px)
- `space.container.padding`: Horizontal page padding (default 24px mobile, 48px desktop)

**Border radius tokens**:
- `radius.sm` (4px), `radius.md` (8px), `radius.lg` (12px), `radius.xl` (16px), `radius.full` (9999px)

**Shadow tokens**:
- `shadow.sm`, `shadow.md`, `shadow.lg`, `shadow.xl`

Output all tokens as a structured JSON file compatible with CSS custom properties and Tailwind CSS v4 theme configuration.

---

### WCAG 2.1 AA Compliance

All color combinations must pass these contrast ratios:

| Use case | Minimum ratio | Requirement |
| --- | --- | --- |
| Body text (< 18px or < 14px bold) | 4.5:1 | WCAG AA |
| Large text (>= 18px or >= 14px bold) | 3:1 | WCAG AA |
| UI components and graphical objects | 3:1 | WCAG AA |
| Focus indicators | 3:1 | WCAG AA |

If the brand's `visual-identity.md` specifies a color combination that fails contrast, execution halts and a conflict report is returned. The report includes:
- Failing pair (foreground + background)
- Actual contrast ratio
- Minimum required ratio
- Three alternative foreground colors that pass the required ratio

**AI slop detection** — Reject these visual patterns without brand justification:
- Purple gradient (#8B5CF6 to #6D28D9) as primary visual element
- White card (`#FFFFFF`) on light gray (`#F9FAFB`) background without border or shadow
- Generic stock icon sets (feather-icons, heroicons without customization)

---

### Component Specifications

Define the following component specifications in the design output:

**Button**:
- Primary: `color.primary` background, `color.primary.foreground` text
- Secondary: `color.secondary` background or transparent with border
- Destructive: `color.semantic.error` background
- States: default, hover (10% darker), focus (3px outline in `color.accent`), disabled (40% opacity)
- Size: sm (h-8), md (h-10, default), lg (h-12)
- Touch target: minimum 44x44px on mobile

**Card**:
- Background: `color.surface`
- Border: 1px solid `color.neutral.200` (light mode)
- Radius: `radius.lg`
- Padding: `space.6` (24px)
- Shadow: `shadow.sm` (default), `shadow.md` (on hover)

**Navigation**:
- Height: 64px desktop, 56px mobile
- Background: transparent on hero, solid on scroll
- Logo: maximum 32px height
- Links: `font.size.sm`, `font.weight.medium`
- Mobile: hamburger menu trigger at 768px breakpoint

**Section layout**:
- Vertical padding: `space.section` (see spacing tokens)
- Max content width: `space.container.max`
- Horizontal padding: `space.container.padding`
- Alternating backgrounds: `color.background` and `color.surface` for visual rhythm

---

### Layout Grid

Default responsive grid:
- Mobile (< 768px): 4 columns, 16px gutter, 24px margin
- Tablet (768px - 1024px): 8 columns, 24px gutter, 32px margin
- Desktop (>= 1024px): 12 columns, 32px gutter, 48px margin

Hero layout options (select based on `visual-identity.md` preference):
- `centered`: Content centered, full-width background image or gradient
- `split-left`: Copy left (7 cols), visual right (5 cols)
- `split-right`: Visual left (5 cols), copy right (7 cols)

---

### Error Handling

**Color palette conflict**: When generated design tokens conflict with `visual-identity.md` defined palette, halt execution and return:

```
BRAND_DESIGN_CONFLICT: Color token mismatch detected.
- Defined in visual-identity.md: <color>
- Generated token: <color>
- Conflict: <explanation>
- Resolution options: [list 2-3 adjustments]
```

**WCAG contrast failure**: When brand colors fail contrast requirements, halt and return the conflict report described in the WCAG section. Do not generate fallback colors silently.

**Missing identity file**: When `visual-identity.md` does not exist or contains only `_TBD_` values, return:

```
BRAND_DESIGN_MISSING_IDENTITY: visual-identity.md is incomplete.
- Unresolved markers found: <list of _TBD_ fields>
- Action required: Run brand interview via /moai design
```

---

## Advanced Patterns

### Dark Mode Support

When brand context specifies dark mode support:
- Define `color.*.dark` variants for each semantic color
- Use CSS `prefers-color-scheme` for automatic switching
- Ensure all token pairs pass contrast in both modes
- Navigation and cards must have distinct dark mode backgrounds

### Animation and Interaction

Keep interaction guidelines minimal and purposeful:
- Transition duration: 150ms (micro), 300ms (standard), 500ms (entrance)
- Easing: `ease-out` for entrances, `ease-in` for exits, `ease-in-out` for state changes
- Avoid animations that trigger on scroll by default (accessibility)
- Respect `prefers-reduced-motion` media query

### Performance Budget

Generated design must meet:
- Lighthouse Performance >= 80
- Lighthouse Accessibility >= 90
- Lighthouse Best Practices >= 80
- Lighthouse SEO >= 80
- Core Web Vitals: LCP < 2.5s, CLS < 0.1
- Font files: maximum 2 custom font families, subset to used character ranges

---

## Works Well With

- `moai-domain-copywriting`: Copy length constraints inform layout choices
- `moai-workflow-design-import`: Replaces code-based design when Claude Design bundle is available
- `moai-workflow-gan-loop`: Design Quality dimension evaluates token compliance and WCAG
- `moai-domain-uiux`: Extends with accessibility audit patterns

---

Source: Absorbed from agency-design-system v1.0.0 on 2026-04-20.
REQ coverage: REQ-SKILL-004, REQ-SKILL-005, REQ-SKILL-006, REQ-FALLBACK-003
Version: 1.0.0
