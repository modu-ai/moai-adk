---
name: moai-library-shadcn
description: >
  Provides shadcn/ui component library expertise for React applications with Tailwind CSS.
  Use when implementing UI components, design systems, or component composition with
  shadcn/ui, Radix primitives, or Tailwind-based component libraries.
  Do NOT use for non-React frameworks or custom CSS-only styling
  (use moai-domain-frontend instead).
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Grep Glob mcp__context7__resolve-library-id mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "2.1.0"
  category: "library"
  modularized: "true"
  status: "active"
  updated: "2026-01-11"
  tags: "library, shadcn, enterprise, development, ui"
  aliases: "moai-library-shadcn"

# MoAI Extension: Triggers
triggers:
  keywords: ["shadcn", "component library", "design system", "radix", "tailwind", "ui components"]
---

## Quick Reference

Enterprise shadcn/ui Component Library Expert

Comprehensive shadcn/ui expertise with AI-powered design system architecture, Context7 integration, and intelligent component orchestration for modern React applications.

Core Capabilities:

- AI-Powered Component Architecture using Context7 MCP
- Intelligent Design System with automated theme customization
- Advanced Component Orchestration with accessibility and performance
- Enterprise UI Framework with zero-configuration design tokens
- Predictive Component Analytics with usage insights

When to Use:

- shadcn/ui component library discussions
- React component architecture planning
- Tailwind CSS integration and design tokens
- Accessibility implementation
- Design system customization

Module Organization:

- Core Concepts: This file covers shadcn/ui overview, architecture, and ecosystem
- Components: The shadcn-components.md module covers component library and advanced patterns
- Theming: The shadcn-theming.md module covers theme system and customization
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

### Installation and Setup

Step 1: Initialize shadcn/ui by running the shadcn-ui init command with npx using the latest version.

Step 2: Configure components.json with the schema URL pointing to ui.shadcn.com/schema.json. Set the style to default and enable RSC and TSX. Configure Tailwind settings including the config path, CSS path, base color, CSS variables enabled, and optional prefix. Set up aliases for components, utils, and ui paths.

Step 3: Add components individually using the shadcn-ui add command with npx, specifying component names such as button, form, or dialog.

### Foundation Technologies

React 19 features include Server Components support, concurrent rendering features, automatic batching improvements, and streaming SSR enhancements.

TypeScript 5.5 provides full type safety across components, improved inference for generics, better error messages, and enhanced developer experience.

Tailwind CSS 3.4 includes JIT compilation, CSS variable support, dark mode variants, and container queries.

Radix UI provides unstyled accessible primitives, keyboard navigation, focus management, and ARIA attributes.

Integration Stack includes React Hook Form for form state management, Zod for schema validation, class-variance-authority for variant management, Framer Motion for animation library, and Lucide React for icon library.

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

- shadcn-components.md module for advanced component patterns and implementation
- shadcn-theming.md module for theme system and customization strategies
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
- API Reference at ui.shadcn.com/docs/components
- Radix UI Documentation at radix-ui.com
- Tailwind CSS Documentation at tailwindcss.com

Latest Versions as of November 2025:

- React 19
- TypeScript 5.5
- Tailwind CSS 3.4
- Radix UI Latest

---

Last Updated: 2026-01-11
Status: Production Ready
