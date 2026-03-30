---
name: moai-ref-react-patterns
description: >
  React/Next.js component design patterns, state management strategies, and project
  structure reference for frontend development. Agent-extending skill that amplifies
  expert-frontend expertise with production-grade React patterns.
  NOT for: backend API design, database modeling, DevOps, mobile apps.
user-invocable: false
metadata:
  version: "1.0.0"
  category: "domain"
  status: "active"
  updated: "2026-03-30"
  tags: "react, nextjs, component, patterns, frontend, reference"
  agent: "expert-frontend"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 3000

# MoAI Extension: Triggers
triggers:
  keywords: ["react", "component", "nextjs", "frontend", "ui", "state"]
  agents: ["expert-frontend"]
  phases: ["run"]
---

# React Patterns Reference

## Target Agent

`expert-frontend` - Applies these patterns directly to component design and state management.

## Component Design Patterns

### 1. Compound Components
Parent and child share implicit state via Context.

Suited for: Tab, Accordion, Dropdown, Select
Structure: `<Select>` + `<Select.Trigger>` + `<Select.Option>`

### 2. Custom Hooks (Extraction Pattern)
Extract state logic into reusable hooks.

Suited for: Form management, API calls, localStorage, debounce
Naming: `use` prefix required - `useForm`, `useDebounce`, `useAuth`

### 3. Container/Presentational Separation
Separate data logic (Container) from UI (Presentational).

Suited for: Large apps, when testability is needed
Container: Data fetch, state management, event handlers
Presentational: Renders only from props, functionally pure

### 4. Headless Component
Provides behavior/state without UI.

Suited for: Design system-independent logic
Examples: headless `useCombobox`, `useDialog`, `useTable`

## State Management Selection Guide

| State Type | Tool | Rationale |
|-----------|------|-----------|
| UI Local | useState, useReducer | Component-internal |
| Server State | React Query / TanStack Query | Caching, refetch, optimistic |
| Global Client | Zustand | Concise, minimal boilerplate |
| Complex Global | Zustand + Immer | Immutability convenience |
| URL State | nuqs / useSearchParams | Filters, pagination |
| Form State | React Hook Form + Zod | Integrated validation |
| Theme/i18n | Context + Provider | Low change frequency |

### Decision Flow
```
Restorable from URL? -> URL state (nuqs)
Server data? -> React Query
Shared across components? -> Zustand
Component-internal? -> useState
Complex transitions? -> useReducer
```

## Next.js App Router Structure

```
src/
├── app/                    # App Router
│   ├── (auth)/             # Auth route group
│   │   ├── login/page.tsx
│   │   └── register/page.tsx
│   ├── (main)/             # Main route group
│   │   ├── dashboard/page.tsx
│   │   └── settings/page.tsx
│   ├── api/                # API Routes
│   ├── layout.tsx          # Root layout
│   └── page.tsx            # Home
├── components/
│   ├── ui/                 # Base UI (Button, Input, Modal)
│   └── features/           # Feature components
│       ├── auth/
│       └── dashboard/
├── hooks/                  # Custom hooks
├── lib/                    # Utilities, config
├── stores/                 # Zustand stores
├── types/                  # TypeScript types
└── styles/                 # Global styles
```

## Component Quality Standards

| Item | Standard |
|------|----------|
| Component Size | Under 200 lines (split if exceeded) |
| Props | 5 or fewer (group into object if exceeded) |
| Custom Hooks | Always extract when reusing logic |
| Error Boundaries | Set at the page level |
| Loading States | Provide loading UI for all async ops |
| Form Validation | Validate on both client and server |

## Performance Patterns

| Pattern | When | Tool |
|---------|------|------|
| Memoization | Expensive computation | `useMemo`, `React.memo` |
| Lazy Loading | Bundle size | `React.lazy`, `next/dynamic` |
| Virtualization | 1000+ item lists | `@tanstack/react-virtual` |
| Image Optimization | Image loading | `next/image` |
| Optimistic Updates | Immediate feedback | React Query `onMutate` |
| Debounce | Search, input | `useDeferredValue` or custom hook |

## Error Handling

### Hierarchical Error Boundaries
```
RootErrorBoundary (global)
  └── LayoutErrorBoundary (per section)
      └── ComponentErrorFallback (individual)
```

### API Error Handling
| HTTP Status | Client Handling |
|------------|----------------|
| 401 | Auto logout + redirect |
| 403 | Unauthorized UI |
| 404 | Not Found page |
| 422 | Per-field form error |
| 429 | Retry + wait notice |
| 500 | Generic error + retry button |

## Accessibility Checklist

- [ ] Alt text on all images
- [ ] Keyboard navigation (Tab, Enter, Escape)
- [ ] ARIA labels (aria-label, role)
- [ ] Color contrast 4.5:1 or above
- [ ] Visible focus indicator
- [ ] Semantic HTML (button, nav, main, section)
