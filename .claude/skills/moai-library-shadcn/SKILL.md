---
name: moai-library-shadcn
description: >
  Provides shadcn/ui component library expertise for React applications
  with shadcn/cli v4, preset system, RTL support, unified radix-ui,
  and OKLCH theming. Use when implementing UI components, design
  systems, CLI workflows, or registry authoring with shadcn/ui.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Grep, Glob, mcp__context7__resolve-library-id, mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "3.0.0"
  category: "library"
  modularized: "true"
  status: "active"
  updated: "2026-03-28"
  tags: "library, shadcn, enterprise, development, ui, preset, RTL, OKLCH, CLI v4, radix-ui, registry"
  aliases: "moai-library-shadcn"

# MoAI Extension: Triggers
triggers:
  keywords: ["shadcn", "component library", "design system", "radix", "tailwind", "ui components", "preset", "RTL", "OKLCH", "radix-ui", "migrate", "shadcn skills", "registry"]
---

## Quick Reference

Enterprise shadcn/ui Component Library Expert

Comprehensive shadcn/ui expertise covering shadcn/cli v4, preset-based design systems, RTL support, unified radix-ui package, OKLCH theming, and AI integration for modern React applications.

Core Capabilities:

- CLI v4: Preset system, framework templates, migration commands, registry management
- Components: Radix UI and Base UI primitives with unified radix-ui package
- Theming: OKLCH color system, CSS variables, dark mode, preset-based configuration
- RTL: First-class right-to-left support with logical CSS properties
- AI Integration: shadcn/skills context system, MCP server for AI assistants
- Registry: Custom component distribution, font system, block library

When to Use:

- shadcn/ui component library setup and customization
- Preset-based design system configuration
- RTL layout implementation or migration
- Registry authoring for component distribution
- AI-assisted component scaffolding with shadcn/skills

Module Organization:

- Core Concepts: This file covers shadcn/ui overview, architecture, and ecosystem
- CLI and Registry: The cli-registry.md module covers CLI v4 commands, presets, RTL, migrations, and registry
- Components: The shadcn-components.md module covers component library and patterns
- Theming: The shadcn-theming.md module covers OKLCH color system and customization
- Advanced Patterns: The advanced-patterns.md module covers complex implementations
- Optimization: The optimization.md module covers performance tuning

---

## Implementation Guide

### shadcn/ui Overview

shadcn/ui is a collection of re-usable components built with Radix UI and Tailwind CSS. Unlike traditional component libraries, it is not an npm package but rather a collection of components you copy into your project.

Key Benefits include full control and ownership of components, zero dependencies beyond Radix UI primitives, complete customization with Tailwind CSS, TypeScript-first design with excellent type safety, and built-in accessibility with WCAG 2.1 AA compliance.

Architecture Philosophy: shadcn/ui components are built on top of Radix UI Primitives which provide unstyled accessible primitives. Tailwind CSS provides utility-first styling. TypeScript ensures type safety throughout. Your customization layer provides full control over the final implementation.

### Core Component Categories

Form Components include Input, Select, Checkbox, Radio, and Textarea. Form validation integrates with react-hook-form and Zod. Accessibility is ensured through proper ARIA labels.

Display Components include Card, Dialog, Sheet, Drawer, and Popover. Responsive design patterns are built in. Dark mode support is included.

Navigation Components include Navigation Menu, Breadcrumb, Tabs, and Pagination. Keyboard navigation support is built in. Focus management is handled automatically.

Data Components include Table, Calendar, DatePicker, and Charts. Virtual scrolling is available for large datasets. TanStack Table integration is supported.

Feedback Components include Alert, Toast, Progress, Badge, and Avatar. Loading states and skeletons are available. Error boundaries are supported.

New Components in 2026 include Empty, Field, Item, Button Group, Spinner, Kbd, and Input Group.

### Installation and Setup

Step 1: Initialize project using shadcn/cli v4. Choose a framework template (Next.js, Vite, TanStack Start, React Router, Astro, or Laravel) and optionally apply a preset for bundled design system configuration.

Step 2: The CLI generates a components.json configuration file. It detects the framework, configures Tailwind CSS (v3 or v4), sets up path aliases, and selects the base primitive library (Radix UI or Base UI).

Step 3: Add components individually using the add command. Use --dry-run to preview changes, --diff to see registry updates. The shadcn info command shows installed components and project configuration.

### Foundation Technologies

React 19 features include Server Components support, concurrent rendering, automatic batching, and streaming SSR.

TypeScript 5.9 provides full type safety, improved inference, and enhanced developer experience.

Tailwind CSS v4 includes CSS-first configuration, CSS variables, OKLCH color support, and container queries.

Radix UI uses the unified radix-ui package with single imports. Base UI is available as an alternative primitive library.

Integration Stack includes React Hook Form for form state, Zod for schema validation, class-variance-authority for variants, Framer Motion for animations, and Lucide React for icons.

### AI-Powered Architecture Design

The ShadcnUIArchitectOptimizer class uses Context7 MCP integration to design optimal shadcn/ui architectures. It initializes a Context7 client, component analyzer, and theme optimizer. The design_optimal_shadcn_architecture method takes design system requirements and fetches latest shadcn/ui and React documentation via Context7. It then optimizes component selection based on UI components and user needs, optimizes theme configuration based on brand guidelines and accessibility requirements, and returns a complete ShadcnUIArchitecture including component library, theme system, accessibility compliance, performance optimization, integration patterns, and customization strategy.

### Best Practices

Requirements include using CSS variables for theme customization, implementing proper TypeScript types, following accessibility guidelines for WCAG 2.1 AA compliance, using Radix UI primitives for complex interactions, testing components with React Testing Library, optimizing bundle size with tree-shaking, and implementing responsive design patterns.

Critical Implementation Standards:

[HARD] Use CSS variables exclusively for color values. This enables dynamic theming, supports dark mode transitions, and maintains design system consistency across all components. Without CSS variables, theme changes require code modifications, dark mode fails, and brand customization becomes unmaintainable.

[HARD] Include accessibility attributes on all interactive elements. This ensures WCAG 2.1 AA compliance, screen reader compatibility, and inclusive user experience for users with disabilities. Missing accessibility attributes excludes users with disabilities, violates legal compliance requirements, and reduces application usability.

[HARD] Implement keyboard navigation for all interactive components. This provides essential navigation method for keyboard users, supports assistive technologies, and improves overall user experience efficiency. Without keyboard navigation, power users cannot efficiently use the application and accessibility compliance fails.

[SOFT] Provide loading states for asynchronous operations. This communicates operation progress to users, reduces perceived latency, and improves user confidence in application responsiveness.

[HARD] Implement error boundaries around component trees. This prevents entire application crashes from isolated component failures, enables graceful error recovery, and maintains application stability.

[HARD] Apply Tailwind CSS classes instead of inline styles. This maintains consistency with design system, enables JIT compilation benefits, supports responsive design variants, and improves bundle size optimization.

[SOFT] Implement dark mode support across all components. This provides user preference respect, reduces eye strain in low-light environments, and aligns with modern UI expectations.

### Performance Optimization

Bundle Size optimization includes tree-shaking to remove unused components, code splitting for large components, lazy loading with React.lazy, and dynamic imports for heavy dependencies.

Runtime Performance optimization includes React.memo for expensive components, useMemo and useCallback for computations, virtual scrolling for large lists, and debouncing user interactions.

Accessibility includes ARIA attributes for all interactive elements, keyboard navigation support, focus management, and screen reader testing.

---

## Advanced Patterns

### Component Composition

The composable pattern involves importing Card, CardHeader, CardTitle, and CardContent from the ui/card components. A DashboardCard component accepts a title and children props, wrapping them in the Card structure with CardHeader containing CardTitle and CardContent containing the children.

### Form Validation

The Zod and React Hook Form integration pattern involves importing useForm from react-hook-form, zodResolver from hookform/resolvers/zod, and z from zod. Define a formSchema with z.object containing field validations such as z.string().email() for email and z.string().min(8) for password. Infer the FormValues type from the schema. The form component uses useForm with zodResolver passing the formSchema. The form element uses form.handleSubmit with an onSubmit handler.

---

## Works Well With

- modules/cli-registry.md for CLI v4 commands, presets, RTL, migrations, and registry
- shadcn-components.md module for advanced component patterns and implementation
- shadcn-theming.md module for OKLCH theme system and customization strategies
- moai-domain-uiux for design system architecture and principles
- moai-lang-typescript for TypeScript best practices
- code-frontend for frontend development patterns

---

## Context7 Integration

Related Libraries:

- shadcn/ui at /shadcn-ui/ui provides re-usable components built with Radix UI and Tailwind
- Radix UI at /radix-ui/primitives provides unstyled accessible component primitives
- Tailwind CSS at /tailwindlabs/tailwindcss provides utility-first CSS framework

Official Documentation:

- shadcn/ui Documentation at ui.shadcn.com/docs
- CLI Reference at ui.shadcn.com/docs/cli
- MCP Server at ui.shadcn.com/docs/mcp
- Radix UI Documentation at radix-ui.com
- Tailwind CSS Documentation at tailwindcss.com

Latest Versions as of March 2026:

- shadcn/cli v4
- React 19
- TypeScript 5.9+
- Tailwind CSS v4
- Unified radix-ui package

---

Last Updated: 2026-03-28
Status: Production Ready

<!-- moai:evolvable-start id="rationalizations" -->
## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "I will customize the component after copying it" | Customization without understanding the original breaks accessibility and keyboard interaction. Read the source first. |
| "I do not need the CLI, I will copy-paste from the docs" | The CLI handles dependency resolution and file placement. Manual copy misses peer dependencies and path conventions. |
| "Any color palette works, I will pick one later" | shadcn/ui uses CSS variables tied to the theme. Changing colors later means updating every component that references them. |
| "I will skip the dark mode variant for now" | shadcn components are designed for light/dark from the start. Skipping dark mode means retrofitting every component later. |
| "This component is too complex, I will build my own" | shadcn components are battle-tested for accessibility and keyboard navigation. Building your own risks regressions in both. |

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="red-flags" -->
## Red Flags

- shadcn component modified without preserving keyboard navigation behavior
- Hardcoded color values instead of CSS variables from the theme
- Component installed manually without using shadcn CLI
- Missing dark mode styles on newly added components
- Radix UI accessibility primitives removed during customization

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="verification" -->
## Verification

- [ ] Components installed via shadcn CLI (show command history)
- [ ] All color values use CSS variables from the theme (no hardcoded hex)
- [ ] Dark mode renders correctly for every new component
- [ ] Keyboard navigation works (Tab, Enter, Escape, Arrow keys) on interactive components
- [ ] Accessibility attributes preserved from the original shadcn source

<!-- moai:evolvable-end -->
