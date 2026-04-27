---
name: moai-design-system
description: >
  Unified design system specialist integrating Intent-First design craft, UI/UX foundations
  (accessibility, design tokens, component architecture), and Pencil MCP tool integration.
  Absorbed from moai-design-craft, moai-domain-uiux, and moai-design-tools (Pencil portion).
  Use when establishing design intent, building design systems, or rendering Pencil designs.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Write, Edit, Grep, Glob, mcp__context7__resolve-library-id, mcp__context7__get-library-docs, mcp__pencil__batch_design, mcp__pencil__batch_get, mcp__pencil__get_screenshot, mcp__pencil__snapshot_layout, mcp__pencil__get_editor_state, mcp__pencil__get_variables, mcp__pencil__set_variables, mcp__pencil__get_guidelines, mcp__pencil__get_style_guide, mcp__pencil__get_style_guide_tags, mcp__pencil__open_document, mcp__pencil__find_empty_space_on_canvas, mcp__pencil__replace_all_matching_properties, mcp__pencil__search_all_unique_properties
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

# MoAI Extension: Triggers
triggers:
  keywords: ["intent-first", "design craft", "design direction", "design intent", "design critique", "craft review", "design memory", "design system", "design audit", "UI/UX", "accessibility", "WCAG", "ARIA", "design tokens", "component library", "theming", "dark mode", "shadcn", "Radix UI", "Storybook", "pencil", "pen frame", "pencil mcp", "web copy", "ux writing", "headline", "cta copy", "landing page copy", "domain exploration", "system.md", "why before what", "design extract", "interface design", "anti-ai writing", "icon", "Style Dictionary", "Lucide", "Iconify", "Hugeicons", "responsive design", "user experience", "Anti-AI Slop", "AI slop prevention", "design to code", "design export", "render dna", "react from design", "tailwind from design", "design context", "ui implementation", "component from design", "layout from design"]
  agents: ["expert-frontend", "team-designer"]
  phases: ["plan", "run", "review"]
---

# MoAI Design System Specialist

Unified design expertise covering three domains: Intent-First craft (design direction and critique), UI/UX foundations (accessibility, tokens, components), and Pencil MCP integration.

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

---

## Pencil MCP Integration (absorbed from moai-design-tools, Pencil portion)

### Core Pencil Tools

| Tool | Purpose |
|------|---------|
| `mcp__pencil__get_editor_state` | Load current design canvas state |
| `mcp__pencil__batch_get` | Fetch multiple design elements |
| `mcp__pencil__get_screenshot` | Capture rendered preview |
| `mcp__pencil__snapshot_layout` | Snapshot layout structure |
| `mcp__pencil__get_style_guide` | Load design tokens and style guide |
| `mcp__pencil__batch_design` | Apply design changes in batch |
| `mcp__pencil__set_variables` | Update design token variables |

### Pencil Workflow

1. Load editor state (`get_editor_state`) to understand current canvas
2. Fetch style guide (`get_style_guide`) to align with existing tokens
3. Apply batch changes (`batch_design`) for multi-element updates
4. Capture screenshot (`get_screenshot`) for visual verification
5. Export via `moai-workflow-design-import` for Pencil-to-code flow

### Pencil-to-Code Export

Pencil designs export to React + Tailwind via `moai-workflow-design-import`. The import workflow reads the exported `.pen` file and generates component code using the project's shadcn/ui configuration.

For Figma MCP integration (fetching Figma design context), see moai-design-tools full skill.

---

## DTCG 2025.10 Token Specification (SPEC-V3R3-DESIGN-PIPELINE-001)

### Specification Reference

The W3C Design Tokens Community Group (DTCG) 2025.10 snapshot defines the canonical token
format for MoAI design pipelines. A Go validator (`internal/design/dtcg/`) enforces schema
compliance before `expert-frontend` consumes any `tokens.json` produced by Path A, B1, or B2.

> **Forward reference**: `internal/design/dtcg/SPEC.md` will be created in Phase 3 of
> SPEC-V3R3-DESIGN-PIPELINE-001. Until then, the spec snapshot date is 2025-10 and the
> supported categories are as listed below.

### Supported Token Categories (2025.10)

All 14 categories below MUST be validated by `internal/design/dtcg.Validate` before code
generation:

| Category | Description | Example value format |
|----------|-------------|----------------------|
| `color` | Color values | `"#RRGGBB"`, CSS color string |
| `dimension` | Length values with units | `"16px"`, `"1rem"` |
| `font` | Font shorthand | CSS font shorthand string |
| `fontFamily` | Font family stack | `"Inter, sans-serif"` |
| `fontWeight` | Font weight | `400`, `"bold"` |
| `duration` | Time values | `"200ms"`, `"0.2s"` |
| `cubicBezier` | Easing curve | `[0.4, 0, 0.2, 1]` (4-number array) |
| `number` | Unitless number | `1.5`, `16` |
| `strokeStyle` | Stroke style keyword | `"solid"`, `"dashed"` |
| `border` | Border composite | `{ color, width, style }` |
| `transition` | Transition composite | `{ duration, timingFunction, delay }` |
| `shadow` | Shadow composite | `{ color, offsetX, offsetY, blur, spread }` |
| `gradient` | Gradient composite | `{ type, stops }` |
| `typography` | Typography composite | `{ fontFamily, fontSize, fontWeight, lineHeight }` |

### Validator Invocation Guidance

`expert-frontend` MUST call `internal/design/dtcg.Validate(tokensJSON)` before generating
any frontend code from a `tokens.json` file. This applies to all three paths (A, B1, B2).

Validation failure produces a structured error report:

```
dtcg.ValidationError{
  TokenPath: "colors.primary",    // DTCG token path
  Category:  "color",             // Expected category
  Rule:      "invalid_format",    // Violated rule
  Got:       "PRIMARY_BLUE",      // Actual value found
}
```

On validation failure: surface the structured error report to the orchestrator and block
code generation. Do not attempt to auto-correct token values.

### Brand Context Priority

Design constitution §3.1 (FROZEN): brand context (`.moai/project/brand/visual-identity.md`)
is the constitutional parent. Token values from any path that conflict with brand constraints
MUST be flagged as warnings and presented for user resolution — not silently overridden.

The DTCG validator gate runs AFTER brand context is loaded and BEFORE frontend code
generation. The validation order is:

1. Load brand context (`.moai/project/brand/`)
2. Run DTCG validator on `tokens.json`
3. Cross-check DTCG-valid tokens against brand constraints
4. Surface conflicts as warnings
5. Proceed to `expert-frontend` code generation

---

<!-- moai:evolvable-start id="verification" -->
## Verification

- [ ] Design intent documented before any visual implementation begins
- [ ] Design tokens follow W3C DTCG hierarchy (primitive → semantic → component)
- [ ] All interactive elements pass WCAG 2.2 keyboard navigation test
- [ ] Color contrast ratios verified (4.5:1 normal text, 3:1 UI components)
- [ ] Web copy free of AI-generic filler phrases
- [ ] Pencil MCP tools used via batch operations (not individual calls) for performance
- [ ] DTCG 2025.10 validator invoked before expert-frontend code generation (REQ-DPL-010)
- [ ] Brand context loaded before DTCG validation (design constitution §3.1)
- [ ] Validation errors surface structured report (not silent auto-correction)

<!-- moai:evolvable-end -->
