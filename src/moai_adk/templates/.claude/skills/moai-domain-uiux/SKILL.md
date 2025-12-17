---
name: moai-domain-uiux
aliases: [moai-foundation-uiux]
category: domain
description: Domain UI/UX Expert - Enterprise design systems, component architecture, accessibility, icons, and theming integration
version: 2.0.0
modularized: true
tags:
 - domain
 - uiux
 - design-systems
 - accessibility
 - components
 - icons
 - theming
updated: 2025-11-30
status: active
---

## Quick Reference (30 seconds)

# Core UI/UX Foundation

Enterprise-grade UI/UX foundation integrating design systems (W3C DTCG 2025.10), component architecture (React 19, Vue 3.5), accessibility (WCAG 2.2), icon libraries (200K+ icons), and theming systems.

Unified Capabilities:
- Design Systems: W3C DTCG 2025.10 tokens, Style Dictionary 4.0, Figma MCP workflows
- Component Architecture: Atomic Design, React 19, Vue 3.5, shadcn/ui, Radix UI primitives
- Accessibility: WCAG 2.2 AA/AAA compliance, keyboard navigation, screen reader optimization
- Icon Libraries: 10+ ecosystems (Lucide, React Icons 35K+, Tabler 5900+, Iconify 200K+)
- Theming: CSS variables, light/dark modes, theme provider, brand customization

When to Use:
- Building modern UI component libraries with design system foundations
- Implementing accessible, enterprise-grade user interfaces
- Setting up design token architecture for multi-platform projects
- Integrating comprehensive icon systems with optimal bundle sizes
- Creating customizable theming systems with dark mode support

Module Organization:
- Components: [Component Architecture](modules/component-architecture.md) (Atomic Design, component patterns, props APIs)
- Design Systems: [Design System Tokens](modules/design-system-tokens.md) (DTCG tokens, Style Dictionary, Figma MCP)
- Accessibility: [Accessibility WCAG](modules/accessibility-wcag.md) (WCAG 2.2 compliance, testing, navigation)
- Icons: [Icon Libraries](modules/icon-libraries.md) (10+ libraries, selection guide, performance optimization)
- Theming: [Theming System](modules/theming-system.md) (theme system, CSS variables, brand customization)
- Examples: [Examples](examples.md) (practical implementation examples)
- Reference: [Reference](reference.md) (external documentation links)

---

## Implementation Guide

### Foundation Stack (November 2025)

Core Technologies:
- React 19 (Server Components, Concurrent Rendering)
- TypeScript 5.5 (Full type safety, improved inference)
- Tailwind CSS 3.4 (JIT compilation, CSS variables, dark mode)
- Radix UI (Unstyled accessible primitives)
- W3C DTCG 2025.10 (Design token specification)
- Style Dictionary 4.0 (Token transformation)
- Figma MCP (Design-to-code automation)
- Storybook 8.x (Component documentation)

Quick Decision Matrix:
| Need | Module | Key Tools |
|------|--------|-----------|
| Design tokens | [Design System Tokens](modules/design-system-tokens.md) | DTCG 2025.10, Style Dictionary 4.0 |
| Component patterns | [Component Architecture](modules/component-architecture.md) | Atomic Design, React 19, shadcn/ui |
| Accessibility | [Accessibility WCAG](modules/accessibility-wcag.md) | WCAG 2.2, jest-axe, keyboard nav |
| Icons | [Icon Libraries](modules/icon-libraries.md) | Lucide, React Icons, Tabler, Iconify |
| Theming | [Theming System](modules/theming-system.md) | CSS variables, Theme Provider |
| Examples | [Examples](examples.md) | React/Vue implementations |

---

## Quick Start Workflows

### 1. Design System Setup (30 minutes)

Step 1: Initialize design tokens
```json
{
 "$schema": "https://tr.designtokens.org/format/",
 "$tokens": {
 "color": {
 "$type": "color",
 "primary": { "500": { "$value": "#3b82f6" } }
 },
 "spacing": {
 "$type": "dimension",
 "md": { "$value": "1rem" }
 }
 }
}
```

Step 2: Transform tokens with Style Dictionary
```bash
npm install --save-dev style-dictionary
npx style-dictionary build
```

Step 3: Integrate with components
```typescript
import { colors, spacing } from '@/tokens'
```

See: [Design System Tokens](modules/design-system-tokens.md) for complete token architecture

---

### 2. Component Library Setup (45 minutes)

Step 1: Initialize shadcn/ui
```bash
npx shadcn-ui@latest init
npx shadcn-ui@latest add button form dialog
```

Step 2: Setup Atomic Design structure
```
components/
 atoms/ (Button, Input, Label)
 molecules/ (FormGroup, Card)
 organisms/ (DataTable, Modal)
```

Step 3: Implement with accessibility
```typescript
<Button aria-label="Submit form" variant="primary">
 Submit
</Button>
```

See: [Component Architecture](modules/component-architecture.md) for patterns and examples

---

### 3. Icon System Integration (15 minutes)

Step 1: Choose icon library
```bash
# General purpose
npm install lucide-react

# Maximum variety
npm install @iconify/react

# Dashboard optimized
npm install @tabler/icons-react
```

Step 2: Implement type-safe icons
```typescript
import { Heart, Search } from 'lucide-react'

<Search className="w-5 h-5 text-gray-600" />
```

See: [Icon Libraries](modules/icon-libraries.md) for library comparison and optimization

---

### 4. Theme System Setup (30 minutes)

Step 1: Configure CSS variables
```css
:root {
 --primary: 222.2 47.4% 11.2%;
 --background: 0 0% 100%;
}

.dark {
 --primary: 210 40% 98%;
 --background: 222.2 84% 4.9%;
}
```

Step 2: Implement Theme Provider
```typescript
<ThemeProvider attribute="class" defaultTheme="system">
 <App />
</ThemeProvider>
```

See: [Theming System](modules/theming-system.md) for complete theme system

---

## Key Principles

1. Design Token First:
- Single source of truth for design decisions
- Semantic naming (`color.primary.500` not `blue-500`)
- Multi-theme support (light/dark)
- Platform-agnostic transformation

2. Accessibility by Default:
- WCAG 2.2 AA minimum (4.5:1 text contrast)
- Keyboard navigation for all interactive elements
- ARIA attributes for screen readers
- Focus management and visible indicators

3. Component Composition:
- Atomic Design hierarchy (Atoms → Molecules → Organisms)
- Props API for reusability
- Variant-based styling (not separate components)
- Type-safe with TypeScript

4. Performance Optimization:
- Tree-shaking for icons (import specific, not *)
- Lazy loading for large components
- React.memo for expensive renders
- Bundle size monitoring

---

## Tool Ecosystem

| Category | Tool | Version | Purpose |
|----------|------|---------|---------|
| Design Tokens | W3C DTCG | 2025.10 | Token specification |
| | Style Dictionary | 4.0+ | Token transformation |
| Components | React | 19 | UI framework |
| | shadcn/ui | Latest | Component library |
| | Radix UI | Latest | Accessible primitives |
| Icons | Lucide | Latest | 1000+ modern icons |
| | React Icons | Latest | 35K+ multi-library |
| | Iconify | Latest | 200K+ universal |
| Theming | Tailwind CSS | 3.4 | Utility-first CSS |
| | CSS Variables | Native | Theme tokens |
| Accessibility | axe DevTools | Latest | Accessibility testing |
| | jest-axe | Latest | Automated a11y tests |
| Documentation | Storybook | 8.x | Component docs |
| | Figma MCP | Latest | Design-to-code |

---

## Module Cross-Reference

### [Component Architecture](modules/component-architecture.md)
Focus: Component architecture and implementation patterns

Key Topics:
- Atomic Design (Atoms, Molecules, Organisms)
- React 19 + Server Components
- Vue 3.5 + Composition API
- shadcn/ui component patterns
- Props API design
- Storybook integration

When to Use: Building or architecting UI component libraries

---

### [Design System Tokens](modules/design-system-tokens.md)
Focus: Design token architecture and tooling

Key Topics:
- W3C DTCG 2025.10 token structure
- Style Dictionary configuration
- Multi-theme support
- Figma MCP workflow
- Semantic naming conventions

When to Use: Setting up design system foundations

---

### [Accessibility WCAG](modules/accessibility-wcag.md)
Focus: WCAG 2.2 compliance and accessibility testing

Key Topics:
- Color contrast validation (4.5:1 AA, 7:1 AAA)
- Keyboard navigation patterns
- Screen reader optimization (ARIA)
- Focus management
- Automated testing (jest-axe)

When to Use: Ensuring accessibility compliance

---

### [Icon Libraries](modules/icon-libraries.md)
Focus: Icon library selection and integration

Key Topics:
- 10+ library comparison (Lucide, React Icons, Tabler, Iconify)
- Bundle size optimization
- Tree-shaking strategies
- Type-safe icon components
- Performance patterns

When to Use: Integrating icon systems with optimal bundle sizes

---

### [Theming System](modules/theming-system.md)
Focus: Theme system implementation

Key Topics:
- CSS variable architecture
- Light/dark mode switching
- System preference detection
- Brand customization
- Tailwind CSS integration

When to Use: Implementing customizable theming

---

### [Examples](examples.md)
Focus: Practical code examples

Key Topics:
- Button component (React, Vue)
- Form validation (Zod + React Hook Form)
- Data table (TanStack Table)
- Modal dialog (focus trap)
- Theme provider
- Icon usage patterns

When to Use: Reference implementations

---

### [Reference](reference.md)
Focus: External documentation links

Key Topics:
- Official documentation (DTCG, WCAG, Figma, Storybook)
- Library references (React, Tailwind, Radix UI)
- Tool documentation (Style Dictionary, jest-axe)
- Best practice guides

When to Use: Finding official resources

---

## Best Practices

DO:
- Use semantic design tokens (`color.primary.500` not `blue-500`)
- Follow Atomic Design hierarchy (Atoms → Molecules → Organisms)
- Verify 4.5:1 contrast ratio for all text (WCAG AA)
- Implement keyboard navigation for all interactive elements
- Tree-shake icons (import specific, avoid `import *`)
- Use CSS variables for theme customization
- Document all props with TypeScript types
- Test components with jest-axe for accessibility

Required Practices:

[HARD] Use design tokens exclusively for all color, spacing, and typography values
WHY: Design tokens provide a single source of truth, enabling consistent theming, multi-platform support, and scalable design systems
IMPACT: Hardcoded values create maintenance debt, break theme switching, and violate design system principles

[HARD] Include ARIA labels on all icon-only interactive elements
WHY: Screen readers cannot interpret visual icons without text alternatives, making content inaccessible to users with visual impairments
IMPACT: Missing ARIA labels violate WCAG 2.2 AA compliance and exclude users who depend on assistive technologies

[HARD] Import icons individually rather than using namespace imports
WHY: Namespace imports (`import * from 'lucide-react'`) bundle entire libraries, defeating tree-shaking optimization
IMPACT: Bundle sizes increase by 500KB-2MB per icon library, degrading load performance and user experience

[HARD] Test all components in both light and dark modes
WHY: Theme switching affects color contrast, readability, and accessibility compliance across all UI states
IMPACT: Untested dark mode implementations may fail WCAG contrast requirements and create unusable interfaces

[HARD] Implement keyboard navigation for all interactive components
WHY: Keyboard-only users and assistive technology users require Tab, Enter, Escape, and Arrow key support
IMPACT: Missing keyboard support violates WCAG 2.2 AA and excludes users who cannot use pointing devices

[HARD] Provide visible focus indicators for all focusable elements
WHY: Focus indicators communicate current keyboard position, essential for navigation and accessibility
IMPACT: Invisible focus states create confusion, violate WCAG 2.2 AA, and make keyboard navigation unusable

[SOFT] Use Tailwind utility classes instead of inline styles
WHY: Tailwind provides consistent spacing scale, responsive design, and automatic purging for optimal bundle sizes
IMPACT: Inline styles bypass design system constraints, create inconsistent spacing, and increase CSS bundle size

[SOFT] Include loading states for all asynchronous operations
WHY: Loading states provide feedback during data fetching, preventing user uncertainty and duplicate actions
IMPACT: Missing loading states create poor user experience with unclear interface states and potential duplicate submissions

---

## Works Well With

Skills:
- `moai-lang-typescript` - TypeScript and JavaScript best practices
- `moai-foundation-core` - TRUST 5 quality validation
- `moai-library-nextra` - Documentation generation
- `moai-library-shadcn` - shadcn/ui specialized patterns (complementary)

Agents:
- `code-frontend` - Frontend component implementation
- `design-uiux` - Design system architecture
- `mcp-figma` - Figma integration workflows
- `core-quality` - Accessibility and quality validation

Commands:
- `/moai:2-run` - TDD implementation cycle
- `/moai:3-sync` - Documentation generation

---

## Migration from Legacy Skills

This skill consolidates 4 previous skills:

moai-component-designer → [Component Architecture](modules/component-architecture.md)
- Atomic Design patterns
- React 19 / Vue 3.5 examples
- Component architecture

moai-design-systems → [Design System Tokens](modules/design-system-tokens.md) + [Accessibility WCAG](modules/accessibility-wcag.md)
- DTCG token architecture
- Figma MCP workflows
- WCAG 2.2 compliance

moai-icons-vector → [Icon Libraries](modules/icon-libraries.md)
- Icon library comparison
- Performance optimization
- Integration patterns

moai-library-shadcn (partially) → [Component Architecture](modules/component-architecture.md) + [Theming System](modules/theming-system.md)
- shadcn/ui patterns
- Theme system
- Component composition

Note: `moai-library-shadcn` remains as a complementary skill for shadcn/ui-specific advanced patterns.

---

## Official Resources

- W3C DTCG: https://designtokens.org
- WCAG 2.2: https://www.w3.org/WAI/WCAG22/quickref/
- React 19: https://react.dev
- Tailwind CSS: https://tailwindcss.com
- Radix UI: https://www.radix-ui.com
- shadcn/ui: https://ui.shadcn.com
- Storybook: https://storybook.js.org
- Figma MCP: https://help.figma.com/hc/en-us/articles/32132100833559
- Style Dictionary: https://styledictionary.com
- Lucide Icons: https://lucide.dev
- Iconify: https://iconify.design

---

Last Updated: 2025-11-26
Status: Production Ready
Version: 1.0.0
