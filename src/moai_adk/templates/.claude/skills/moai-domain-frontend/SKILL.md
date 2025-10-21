---

name: moai-domain-frontend
description: React/Vue/Angular development with state management, performance optimization, and accessibility. Use when working on frontend interfaces scenarios.
allowed-tools:
  - Read
  - Bash
---

# Frontend Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand for frontend delivery |
| Trigger cues | Component architecture, design systems, accessibility, performance budgets. |
| Tier | 4 |

## What it does

Provides expertise in modern frontend development using React, Vue, or Angular, including state management patterns, performance optimization techniques, and accessibility (a11y) best practices.

## When to use

- Engages when building or reviewing UI/front-end experiences.
- “Front-end development”, “React components”, “state management”, “performance optimization”
- Automatically invoked when working with frontend projects
- Frontend SPEC implementation (`/alfred:2-run`)

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
```bash
$ npm run lint && npm run test
$ npm run build -- --profiling
```

## Inputs
- Domain-specific design documents and user requirements.
- Project technology stack and operational constraints.

## Outputs
- Domain-specific architecture or implementation guidelines.
- Recommended list of associated sub-agents/skills.

## Failure Modes
- When the domain document does not exist or is ambiguous.
- When the project strategy is unconfirmed and cannot be specified.

## Dependencies
- `.moai/project/` document and latest technical briefing are required.

## References
- Google. "Web.dev Performance Guidelines." https://web.dev/fast/ (accessed 2025-03-29).
- W3C. "Web Content Accessibility Guidelines (WCAG) 2.2." https://www.w3.org/TR/WCAG22/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: Codified input/output and failure responses for domain skills.

## Works well with

- alfred-trust-validation (frontend testing)
- typescript-expert (type-safe React/Vue)
- alfred-performance-optimizer (performance profiling)

## Best Practices
- Record supporting documentation (version/link) for each domain decision.
- Review performance, security, and operational requirements simultaneously at an early stage.
