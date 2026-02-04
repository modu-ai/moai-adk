---
name: moai-domain-uiux
description: >
  UI/UX design systems specialist covering accessibility, icons, theming,
  design tokens, and user experience patterns.
  Use when user asks about design systems, WCAG accessibility compliance, ARIA patterns,
  icon libraries, dark mode theming, design tokens, or user experience research.
  Do NOT use for React component coding or frontend implementation
  (use moai-domain-frontend instead) or shadcn/ui specifics
  (use moai-library-shadcn instead).
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Grep Glob mcp__context7__resolve-library-id mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "2.0.0"
  category: "domain"
  status: "active"
  updated: "2026-01-11"
  modularized: "true"
  tags: "domain, uiux, design-systems, accessibility, components, icons, theming"

# MoAI Extension: Triggers
triggers:
  keywords:
    - UI/UX
    - design system
    - accessibility
    - WCAG
    - ARIA
    - icon
    - theming
    - dark mode
    - design tokens
    - component library
    - Radix UI
    - shadcn
    - Storybook
    - Pencil
    - design tokens
    - Style Dictionary
    - Lucide
    - Iconify
    - responsive design
    - user experience
---

## Quick Reference

Core UI/UX Foundation - Enterprise-grade UI/UX foundation integrating design systems (W3C DTCG 2025.10), component architecture (React 19, Vue 3.5), accessibility (WCAG 2.2), icon libraries (200K+ icons), and theming systems.

Unified Capabilities:

- Design Systems: W3C DTCG 2025.10 tokens, Style Dictionary 4.0, Pencil MCP workflows
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

- Components: modules/component-architecture.md (Atomic Design, component patterns, props APIs)
- Design Systems: modules/design-system-tokens.md (DTCG tokens, Style Dictionary, Pencil MCP)
- Accessibility: modules/accessibility-wcag.md (WCAG 2.2 compliance, testing, navigation)
- Icons: modules/icon-libraries.md (10+ libraries, selection guide, performance optimization)
- Theming: modules/theming-system.md (theme system, CSS variables, brand customization)
- Web Interface Guidelines: modules/web-interface-guidelines.md (Vercel Labs comprehensive UI/UX compliance)
- Examples: examples.md (practical implementation examples)
- Reference: reference.md (external documentation links)

---

## Implementation Guide

### Foundation Stack

Core Technologies:

- React 19 with Server Components and Concurrent Rendering
- TypeScript 5.5 with full type safety and improved inference
- Tailwind CSS 3.4 with JIT compilation, CSS variables, and dark mode
- Radix UI for unstyled accessible primitives
- W3C DTCG 2025.10 for design token specification
- Style Dictionary 4.0 for token transformation
- Pencil MCP for design-to-code automation
- Storybook 8.x for component documentation

Quick Decision Guide:

For design tokens, use modules/design-system-tokens.md with DTCG 2025.10 and Style Dictionary 4.0.

For component patterns, use modules/component-architecture.md with Atomic Design, React 19, and shadcn/ui.

For accessibility, use modules/accessibility-wcag.md with WCAG 2.2, jest-axe, and keyboard navigation.

For icons, use modules/icon-libraries.md with Lucide, React Icons, Tabler, and Iconify.

For theming, use modules/theming-system.md with CSS variables and Theme Provider.

For practical examples, use examples.md with React and Vue implementations.

---

## Quick Start Workflows

### Design System Setup

Step 1: Initialize design tokens by creating a JSON file with DTCG schema URL. Define color tokens with type color and primary 500 value. Define spacing tokens with type dimension and md value of 1rem.

Step 2: Transform tokens with Style Dictionary by installing the package and running the build command.

Step 3: Integrate with components by importing colors and spacing from the tokens directory.

See modules/design-system-tokens.md for complete token architecture.

### Component Library Setup

Step 1: Initialize shadcn/ui by running the init command, then add button, form, and dialog components.

Step 2: Set up Atomic Design structure with atoms directory for Button, Input, and Label components, molecules directory for FormGroup and Card components, and organisms directory for DataTable and Modal components.

Step 3: Implement with accessibility by adding aria-label attributes to interactive elements.

See modules/component-architecture.md for patterns and examples.

### Icon System Integration

Step 1: Choose icon library based on needs. Install lucide-react for general purpose, iconify/react for maximum variety, or tabler/icons-react for dashboard optimization.

Step 2: Implement type-safe icons by importing specific icons and applying className for sizing and color.

See modules/icon-libraries.md for library comparison and optimization.

### Theme System Setup

Step 1: Configure CSS variables in root selector for primary and background colors. Define dark class with inverted values for dark mode.

Step 2: Implement Theme Provider by wrapping the application with attribute set to class and defaultTheme set to system.

See modules/theming-system.md for complete theme system.

---

## Key Principles

Design Token First:

- Single source of truth for design decisions
- Semantic naming using color.primary.500 format rather than blue-500
- Multi-theme support for light and dark modes
- Platform-agnostic transformation

Accessibility by Default:

- WCAG 2.2 AA minimum with 4.5:1 text contrast
- Keyboard navigation for all interactive elements
- ARIA attributes for screen readers
- Focus management and visible indicators

Component Composition:

- Atomic Design hierarchy from Atoms to Molecules to Organisms
- Props API for reusability
- Variant-based styling rather than separate components
- Type-safe with TypeScript

Performance Optimization:

- Tree-shaking for icons by importing specific icons rather than all
- Lazy loading for large components
- React.memo for expensive renders
- Bundle size monitoring

---

## Best Practices

Required Practices:

Use design tokens exclusively for all color, spacing, and typography values. Design tokens provide a single source of truth, enabling consistent theming, multi-platform support, and scalable design systems. Hardcoded values create maintenance debt and break theme switching.

Include ARIA labels on all icon-only interactive elements. Screen readers cannot interpret visual icons without text alternatives. Missing ARIA labels violate WCAG 2.2 AA compliance.

Import icons individually rather than using namespace imports. Namespace imports bundle entire libraries, defeating tree-shaking optimization. Bundle sizes increase by 500KB-2MB per icon library.

Test all components in both light and dark modes. Theme switching affects color contrast, readability, and accessibility compliance.

Implement keyboard navigation for all interactive components. Keyboard-only users require Tab, Enter, Escape, and Arrow key support.

Provide visible focus indicators for all focusable elements. Focus indicators communicate current keyboard position for navigation and accessibility.

Use Tailwind utility classes instead of inline styles. Tailwind provides consistent spacing scale, responsive design, and automatic purging for optimal bundle sizes.

Include loading states for all asynchronous operations. Loading states provide feedback during data fetching, preventing user uncertainty.

---

## Works Well With

Skills:

- moai-lang-typescript - TypeScript and JavaScript best practices
- moai-foundation-core - TRUST 5 quality validation
- moai-library-nextra - Documentation generation
- moai-library-shadcn - shadcn/ui specialized patterns

Agents:

- code-frontend - Frontend component implementation
- design-uiux - Design system architecture
- mcp-pencil - Pencil MCP design workflows
- core-quality - Accessibility and quality validation

Commands:

- /moai:2-run - DDD implementation cycle
- /moai:3-sync - Documentation generation

---

## Resources

For detailed module documentation, see the modules directory.

For practical code examples, see examples.md.

For external documentation links, see reference.md.

Official Resources:

- W3C DTCG: https://designtokens.org
- WCAG 2.2: https://www.w3.org/WAI/WCAG22/quickref/
- React 19: https://react.dev
- Tailwind CSS: https://tailwindcss.com
- Radix UI: https://www.radix-ui.com
- shadcn/ui: https://ui.shadcn.com
- Storybook: https://storybook.js.org
- Pencil: https://docs.pencil.dev
- Style Dictionary: https://styledictionary.com
- Lucide Icons: https://lucide.dev
- Iconify: https://iconify.design
- Vercel Web Interface Guidelines: https://github.com/vercel-labs/web-interface-guidelines

---

Last Updated: 2026-01-11
Status: Production Ready
Version: 2.0.0
