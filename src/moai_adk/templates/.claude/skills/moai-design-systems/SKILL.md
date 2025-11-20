---
name: moai-design-systems
version: 4.0.0
updated: 2025-11-20
status: stable
description: Design system patterns with DTCG tokens, WCAG 2.2, and Figma integration
allowed-tools: [Read, Write, WebSearch, WebFetch]
---

# Design Systems Expert

**Production-Ready Design Systems with Tokens & Accessibility**

> **Focus**: W3C DTCG 2025.10, WCAG 2.2 (4.5:1 contrast), Atomic Design  
> **Tools**: Style Dictionary, Storybook, Figma MCP, axe DevTools

---

## Overview

Build consistent, accessible UI with design tokens, component libraries, and automated testing.

### Core Pillars

1.  **Design Tokens**: Single source of truth (color, spacing, typography)
2.  **Component Library**: Reusable building blocks (Atomic Design)
3.  **Accessibility**: WCAG 2.2 compliance (contrast, keyboard, screen readers)

---

## Quick Start

### 1. Design Tokens (DTCG)

JSON-based token system for colors, spacing, and typography.

**Structure**:

```json
{
  "color": {
    "primary": {
      "500": { "$value": "#3b82f6" }
    }
  },
  "spacing": {
    "md": { "$value": "1rem" }
  }
}
```

See: [examples.md](./examples.md#dtcg-tokens) for complete setup

### 2. Component Library

Type-safe components with Class Variance Authority (CVA).

**Pattern**:

- Variants (primary, secondary, outline)
- Sizes (sm, md, lg)
- States (default, hover, disabled)

See: [examples.md](./examples.md#component-library)

### 3. WCAG 2.2 Accessibility

Ensure 4.5:1 contrast ratio for text (AA standard).

**Tools**: axe DevTools, color contrast checker

See: [examples.md](./examples.md#accessibility)

### 4. Figma Integration (MCP)

Extract design tokens from Figma variables.

**Workflow**:

1. Define tokens in Figma
2. Extract via MCP
3. Transform with Style Dictionary

See: [examples.md](./examples.md#figma-integration)

---

## Atomic Design Hierarchy

```
Atoms → Molecules → Organisms → Templates → Pages

- Atoms: Button, Input, Icon
- Molecules: FormField, Card
- Organisms: Header, DataTable
- Templates: DashboardLayout
- Pages: Dashboard (with real data)
```

---

## Best Practices

1.  **Semantic Naming**: `color.primary.500` not `color.blue`
2.  **Responsive Design**: Mobile-first approach
3.  **Dark Mode**: Use CSS variables for theme switching
4.  **Accessibility**: Test with screen readers (VoiceOver, NVDA)

---

## Validation Checklist

- [ ] **Tokens**: DTCG 2025.10 format?
- [ ] **Contrast**: 4.5:1 for text (WCAG AA)?
- [ ] **Keyboard**: All interactive elements navigable?
- [ ] **ARIA**: Labels on form fields?
- [ ] **Motion**: `prefers-reduced-motion` respected?
- [ ] **Testing**: Storybook + axe DevTools configured?

---

## Related Skills

- `moai-lib-shadcn-ui`: Pre-built accessible components
- `moai-domain-frontend`: React patterns

---

## Additional Resources

- [examples.md](./examples.md): Token setup, components, accessibility
- [reference.md](./reference.md): WCAG guidelines, Figma MCP API

---

**Last Updated**: 2025-11-20
