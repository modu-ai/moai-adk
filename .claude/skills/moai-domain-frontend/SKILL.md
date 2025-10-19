---
name: moai-domain-frontend
description: React/Vue/Angular development with state management, performance optimization,
  and accessibility
allowed-tools:
- Read
- Bash
---

# Frontend Expert

## What it does

Provides expertise in modern frontend development using React, Vue, or Angular, including state management patterns, performance optimization techniques, and accessibility (a11y) best practices.

## When to use

- "프론트엔드 개발", "React 컴포넌트", "상태 관리", "성능 최적화"
- Automatically invoked when working with frontend projects
- Frontend SPEC implementation (`/alfred:2-build`)

## How it works

**React Development**:
- **Functional components**: Hooks (useState, useEffect, useMemo)
- **State management**: Redux, Zustand, Jotai
- **Performance**: React.memo, useCallback, code splitting
- **Testing**: React Testing Library

**Vue Development**:
- **Composition API**: setup(), reactive(), computed()
- **State management**: Pinia, Vuex
- **Performance**: Virtual scrolling, lazy loading
- **Testing**: Vue Test Utils

**Angular Development**:
- **Components**: TypeScript classes with decorators
- **State management**: NgRx, Akita
- **Performance**: OnPush change detection, lazy loading
- **Testing**: Jasmine, Karma

**Performance Optimization**:
- **Code splitting**: Dynamic imports, route-based splitting
- **Lazy loading**: Images, components
- **Bundle optimization**: Tree shaking, minification
- **Web Vitals**: LCP, FID, CLS optimization

**Accessibility (a11y)**:
- **Semantic HTML**: Proper use of HTML5 elements
- **ARIA attributes**: Roles, labels, descriptions
- **Keyboard navigation**: Focus management
- **Screen reader support**: Alt text, aria-live

## Examples

### Example 1: React component with performance optimization
User: "/alfred:2-build UI-001"
Claude: (creates RED component test, GREEN implementation with React.memo, REFACTOR)

### Example 2: Accessibility audit
User: "접근성 검사"
Claude: (runs axe-core or Lighthouse and reports issues)

## Works well with

- alfred-trust-validation (frontend testing)
- typescript-expert (type-safe React/Vue)
- alfred-performance-optimizer (performance profiling)
