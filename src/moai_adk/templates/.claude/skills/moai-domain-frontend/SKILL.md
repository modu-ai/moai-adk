---
name: moai-domain-frontend
description: >
  Frontend development specialist covering React 19, Next.js 16, Vue 3.5,
  and modern UI/UX patterns with component architecture. Use when building
  web UIs, implementing components, optimizing frontend performance, or
  integrating state management.
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
  tags: "frontend, react, nextjs, vue, ui, components"
  author: "MoAI-ADK Team"
  context7-libraries: "/facebook/react, /vercel/next.js, /vuejs/vue"

# MoAI Extension: Triggers
triggers:
  keywords:
    - frontend
    - UI
    - component
    - React
    - Next.js
    - Vue
    - user interface
    - responsive
    - TypeScript
    - JavaScript
    - state management
    - hooks
    - props
    - JSX
    - TSX
    - client-side
    - browser
    - DOM
    - CSS
    - Tailwind
---

# Frontend Development Specialist

## Quick Reference

Modern Frontend Development - Comprehensive patterns for React 19, Next.js 16, Vue 3.5.

Core Capabilities:

- React 19: Server components, concurrent features, cache(), Suspense
- Next.js 16: App Router, Server Actions, ISR, Route handlers
- Vue 3.5: Composition API, TypeScript, Pinia state management
- Component Architecture: Design systems, compound components, CVA
- Performance: Code splitting, dynamic imports, memoization

When to Use:

- Modern web application development
- Component library creation
- Frontend performance optimization
- UI/UX with accessibility

---

## Module Index

Load specific modules for detailed patterns:

### Framework Patterns

React 19 Patterns in modules/react19-patterns.md:

- Server Components, Concurrent features, cache() API, Form handling

Next.js 16 Patterns in modules/nextjs16-patterns.md:

- App Router, Server Actions, ISR, Route Handlers, Parallel Routes

Vue 3.5 Patterns in modules/vue35-patterns.md:

- Composition API, Composables, Reactivity, Pinia, Provide/Inject

### Architecture Patterns

Component Architecture in modules/component-architecture.md:

- Design tokens, CVA variants, Compound components, Accessibility

State Management in modules/state-management.md:

- Zustand, Redux Toolkit, React Context, Pinia

Performance Optimization in modules/performance-optimization.md:

- Code splitting, Dynamic imports, Image optimization, Memoization

Vercel React Best Practices in modules/vercel-react-best-practices.md:

- 45 rules across 8 categories from Vercel Engineering
- Eliminating waterfalls, bundle optimization, server-side performance
- Client-side data fetching, re-render optimization, rendering performance

---

## Implementation Quickstart

### React 19 Server Component

Create an async page component that uses the cache function from React to memoize data fetching. Import Suspense for loading states. Define a getData function that fetches from the API endpoint with an id parameter and returns JSON. In the page component, wrap the DataDisplay component with Suspense using a Skeleton fallback, and pass the awaited getData result as the data prop.

### Next.js Server Action

Create a server action file with the use server directive. Import revalidatePath from next/cache and z from zod for validation. Define a schema with title (minimum 1 character) and content (minimum 10 characters). The createPost function accepts FormData, validates with safeParse, returns errors on failure, creates the post in the database, and calls revalidatePath for the posts page.

### Vue Composable

Create a useUser composable that accepts a userId ref parameter. Define user as a nullable ref, loading as a boolean ref, and fullName as a computed property that concatenates firstName and lastName. Use watchEffect to set loading true, fetch the user data asynchronously, assign to user ref, and set loading false. Return the user, loading, and fullName refs.

### CVA Component

Import cva and VariantProps from class-variance-authority. Define buttonVariants with base classes for inline-flex, items-center, justify-center, rounded-md, and font-medium. Add variants object with variant options for default (primary background with hover) and outline (border with hover accent). Add size options for sm (h-9, px-3, text-sm), default (h-10, px-4), and lg (h-11, px-8). Set defaultVariants for variant and size. Export a Button component that applies the variants to a button element className.

---

## Works Well With

- moai-domain-backend - Full-stack development
- moai-library-shadcn - Component library integration
- moai-domain-uiux - UI/UX design principles
- moai-lang-typescript - TypeScript patterns
- moai-workflow-testing - Frontend testing

---

## Technology Stack

Frameworks: React 19, Next.js 16, Vue 3.5, Nuxt 3

Languages: TypeScript 5.9+, JavaScript ES2024

Styling: Tailwind CSS 3.4+, CSS Modules, shadcn/ui

State: Zustand, Redux Toolkit, Pinia

Testing: Vitest, Testing Library, Playwright

---

## Resources

Module files in the modules directory contain detailed patterns.

For working code examples, see [examples.md](examples.md).

Official documentation:

- React: https://react.dev/
- Next.js: https://nextjs.org/docs
- Vue: https://vuejs.org/

---

Version: 2.0.0
Last Updated: 2026-01-11
