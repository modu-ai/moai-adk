---
name: agency-design-system
description: >
  Visual design system patterns for AI Agency projects covering color palettes,
  typography, spacing, layout, and component design tokens. Enforces brand visual
  consistency and eliminates AI-generated design anti-patterns.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Grep, Glob
user-invocable: false
metadata:
  version: "1.0.0"
  category: "agency"
  status: "active"
  updated: "2026-04-02"
  evolution_count: "0"
  confidence_score: "0.00"
  base_context: "visual-identity.md"
  dependencies: "agency-evaluation-criteria"
  tags: "design, visual-identity, color, typography, layout, components, design-tokens"
  related-skills: "agency-copywriting, agency-frontend-patterns, agency-evaluation-criteria"

progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

triggers:
  keywords: ["design", "visual", "color", "typography", "layout", "component", "spacing", "design-system", "UI", "theme"]
  agents: ["designer", "builder"]
  phases: ["plan", "run"]
---

# Agency Design System Skill

Governs all visual design decisions for AI Agency projects. Ensures brand-aligned visual identity, consistent design tokens, and elimination of generic AI-generated design patterns.

---

## Static Zone

### Identity

**Purpose**: Define and enforce visual design system rules, component patterns, and design token standards for agency project generation.

**Input Contract**:
- Copy document (from copywriter output: `copy.md`)
- Visual identity preferences (from `.agency/context/visual-identity.md`)
- Design preferences expressed during client interview

**Output Contract**:
- `design-spec.md` containing:
  - Color tokens (primary, secondary, accent, semantic, neutral)
  - Typography scale (headings, body, caption, with responsive sizes)
  - Spacing system (base unit, scale multipliers)
  - Component specifications (buttons, cards, sections, navigation)
  - Layout grid (columns, gutters, breakpoints)
  - Animation/interaction guidelines

**Owner**: `.agency/context/visual-identity.md`

### Core Principles

> Derived from Brand Context. Never auto-modified. Manual editing only.

1. First section (hero) establishes the entire site tone — all subsequent sections chain from it
2. Design must reflect brand identity from visual-identity.md, not generic templates
3. No AI slop: avoid purple gradients + white cards + generic stock icons
4. Accessibility: minimum WCAG 2.1 AA contrast ratios
5. Responsive-first: mobile → tablet → desktop progression

---

## Dynamic Zone

> Rules, Anti-Patterns, and Heuristics below evolve via user feedback.
> The agency-learner agent modifies this zone based on accumulated learnings.

### Rules

<!-- Initially empty. Rules are added as the system learns user preferences.
Format:
### R-001 (v1 | confidence: 0.XX | evidence: N)
[Rule description]
- Context: [When this rule applies]
- Evidence: [F-XXX references]
- Added: vX.Y.Z
-->

(No rules yet. Rules will be added as the system learns from user feedback.)

### Anti-Patterns

<!-- Initially empty. Anti-patterns are added when repeated negative feedback is detected.
Format:
### AP-001 (confidence: 0.XX | evidence: N)
AVOID: [What to avoid]
- Why: [Reason]
- Evidence: [F-XXX references]
-->

(No anti-patterns yet.)

### Heuristics

<!-- Initially empty. Heuristics are soft guidelines added at 3x feedback threshold.
Format:
### H-001 (weight: 0.X | evidence: N)
[Soft guideline]
- Context: [When applicable]
- Note: [Caveats]
-->

(No heuristics yet.)

---

## Evolution Log

<!-- Tracks all changes to the Dynamic Zone.
Format:
- vX.Y.Z: [Change description]
-->

- v1.0.0: Initial creation (Static Zone from Brand Context, empty Dynamic Zone)
