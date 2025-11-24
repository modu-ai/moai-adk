---
name: moai-domain-frontend
description: Enterprise Frontend Development with React 19, Next.js 15, Vue 3.5 and modern frameworks
version: 1.0.0
modularized: false
tags:
  - architecture
  - enterprise
  - patterns
  - frontend
updated: 2025-11-24
status: active
---

## 📊 Skill Metadata

**version**: 1.0.0
**modularized**: false
**last_updated**: 2025-11-24
**compliance_score**: 100%
**auto_trigger_keywords**: react, domain, moai, frontend, vue, next.js, typescript
**context7_references**: ["/facebook/react", "/vuejs/core", "/vercel/next.js"]
**test_documentation**: ["Component patterns", "State management", "Hooks patterns", "Performance optimization", "Testing strategies"]


## Quick Reference (30 seconds)

# Enterprise Frontend Development

**Primary Focus**: React 19, Next.js 15, Vue 3.5, component architecture, state management, optimization
**Best For**: Web application development, component libraries, performance optimization, full-stack development
**Key Frameworks**: React 19.0, Next.js 15.x, Vue 3.5, Vite 5.2, TypeScript 5.4
**Auto-triggers**: React files, Next.js projects, Vue components, frontend architecture

| Framework | Version | Release | Support |
|-----------|---------|---------|---------|
| React | 19.0 | Oct 2024 | Active |
| Next.js | 15.x | Oct 2024 | ✅ |
| Vue | 3.5 | Aug 2024 | Active |
| Vite | 5.2 | Nov 2024 | ✅ |

---

## What It Does

Enterprise-grade frontend development with modern frameworks and patterns. React Server Components, advanced state management, performance optimization, and full-stack type safety.

**Key capabilities**:
- ✅ React 19 with Server Components and use() hook
- ✅ Next.js 15 PPR, Turbopack, App Router
- ✅ Vue 3.5 with signals-based reactivity
- ✅ Component architecture and design systems
- ✅ Advanced state management (Zustand, Redux, Pinia)
- ✅ Performance optimization and bundle analysis
- ✅ TypeScript for type-safe development

---

## When to Use

**Automatic triggers**:
- React/JSX files, Next.js projects
- Vue projects, component development
- Frontend architecture decisions
- State management patterns

**Manual invocation**:
- Design component architecture
- Implement state management strategy
- Optimize bundle size and performance
- Review frontend code for best practices

---

## Three-Level Learning Path

### Level 1: Fundamentals (See examples.md)

Core frontend concepts:
- **React Basics**: Components, hooks, state, effects
- **Next.js Setup**: Pages, App Router, API routes
- **Vue Basics**: Components, composition API, lifecycle
- **Styling**: CSS Modules, Tailwind, Styled Components
- **Testing**: Unit tests, component tests with Vitest

### Level 2: Advanced Patterns (See modules/architecture.md)

Production-ready architecture:
- **Server Components**: React Server Components in Next.js
- **State Management**: Zustand, Redux, Pinia, Context API
- **Data Fetching**: SWR, React Query, Server Actions
- **Advanced Hooks**: Custom hooks, useReducer, useMemo
- **Performance**: Code splitting, lazy loading, memoization
- **Forms**: Form libraries, validation, accessibility

### Level 3: Performance & Optimization (See modules/best-practices.md)

Production deployment and optimization:
- **Bundle Optimization**: Code splitting, dynamic imports, tree-shaking
- **Runtime Performance**: React profiler, Lighthouse, Core Web Vitals
- **SEO & Accessibility**: Meta tags, structured data, WCAG compliance
- **Deployment**: Vercel, Netlify, GitHub Pages, Docker
- **Monitoring**: Error tracking, analytics, performance metrics
- **DevOps**: CI/CD, environment variables, deployment strategies

---

## Best Practices

✅ **DO**:
- Use TypeScript for type safety
- Implement proper error boundaries
- Optimize images with next/image
- Use composition over inheritance
- Test critical user paths
- Monitor Web Vitals
- Keep components small and reusable
- Use proper key in lists

❌ **DON'T**:
- Use index as list key
- Create large monolithic components
- Mutate state directly
- Ignore performance warnings
- Ship unnecessary dependencies
- Skip accessibility features
- Inline styles in components
- Over-optimize prematurely

---

## Tool Versions (2025-11-22)

| Tool | Version | Purpose |
|------|---------|---------|
| **React** | 19.0 | UI Library |
| **Next.js** | 15.x | Full-stack framework |
| **Vue** | 3.5 | Reactive framework |
| **TypeScript** | 5.4 | Type safety |
| **Vite** | 5.2 | Build tool |
| **Tailwind** | 4.0 | CSS utility |

---

## Installation & Setup

```bash
# Create Next.js app (React)
npx create-next-app@latest --typescript

# Create Vite + React app
npm create vite@latest my-app -- --template react-ts

# Create Vue app
npm create vite@latest my-app -- --template vue-ts

# Install common dependencies
npm install zustand react-query @hookform/react axios
```

---

## Works Well With

- `moai-lang-typescript` (TypeScript patterns)
- `moai-lang-html-css` (HTML/CSS semantics)
- `moai-domain-testing` (Test strategies)
- `moai-domain-database` (API integration)

---

## Learn More

- **Practical Examples**: See `examples.md` for 20+ real-world patterns
- **Architecture Patterns**: See `modules/architecture.md` for complex applications
- **Best Practices**: See `modules/best-practices.md` for production deployment
- **React Docs**: https://react.dev/
- **Next.js Docs**: https://nextjs.org/docs
- **Vue Docs**: https://vuejs.org/

---

## Changelog

- **v4.0.0** (2025-11-22): Modularized with architecture and best practices
- **v3.5.0** (2025-11-13): React 19 Server Components, Next.js 15 PPR
- **v3.0.0** (2025-10-01): React 18 hooks, composition API patterns
- **v2.0.0** (2025-08-01): TypeScript migration, design systems

---

## Context7 Integration

### Related Libraries & Tools
- [React](/facebook/react): A JavaScript library for building UIs
- [Next.js](/vercel/next.js): React framework for production
- [Vue](/vuejs/vue): The progressive JavaScript framework
- [TypeScript](/microsoft/TypeScript): Typed superset of JavaScript
- [Tailwind CSS](/tailwindlabs/tailwindcss): Utility-first CSS

### Official Documentation
- [React Documentation](https://react.dev/)
- [Next.js Documentation](https://nextjs.org/docs)
- [Vue Documentation](https://vuejs.org/)
- [TypeScript Handbook](https://www.typescriptlang.org/docs/)

---

**Skills**: Skill("moai-lang-typescript"), Skill("moai-domain-testing"), Skill("moai-lang-html-css")
**Auto-loads**: React/Vue/Next.js projects with TypeScript

