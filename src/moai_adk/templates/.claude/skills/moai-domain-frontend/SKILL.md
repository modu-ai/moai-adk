---
name: "moai-domain-frontend"
description: "Frontend development specialist covering React 19, Next.js 16, Vue 3.5, and modern UI/UX patterns with component architecture. Use when building web UIs, implementing components, optimizing frontend performance, or integrating state management."
version: 2.0.0
category: "domain"
modularized: true
user-invocable: false
tags: ['frontend', 'react', 'nextjs', 'vue', 'ui', 'components']
context7-libraries: ['/facebook/react', '/vercel/next.js', '/vuejs/vue']
updated: 2026-01-08
allowed-tools:
  - Read
  - Grep
  - Glob
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
status: "active"
author: "MoAI-ADK Team"
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

[React 19 Patterns](modules/react19-patterns.md):
- Server Components, Concurrent features, cache() API, Form handling

[Next.js 16 Patterns](modules/nextjs16-patterns.md):
- App Router, Server Actions, ISR, Route Handlers, Parallel Routes

[Vue 3.5 Patterns](modules/vue35-patterns.md):
- Composition API, Composables, Reactivity, Pinia, Provide/Inject

### Architecture Patterns

[Component Architecture](modules/component-architecture.md):
- Design tokens, CVA variants, Compound components, Accessibility

[State Management](modules/state-management.md):
- Zustand, Redux Toolkit, React Context, Pinia

[Performance Optimization](modules/performance-optimization.md):
- Code splitting, Dynamic imports, Image optimization, Memoization

---

## Implementation Quickstart

### React 19 Server Component

```tsx
import { cache } from 'react'
import { Suspense } from 'react'

const getData = cache(async (id: string) => {
  const res = await fetch(`/api/data/${id}`)
  return res.json()
})

export default async function Page({ params }: { params: { id: string } }) {
  return (
    <Suspense fallback={<Skeleton />}>
      <DataDisplay data={await getData(params.id)} />
    </Suspense>
  )
}
```

### Next.js Server Action

```tsx
'use server'

import { revalidatePath } from 'next/cache'
import { z } from 'zod'

const schema = z.object({
  title: z.string().min(1),
  content: z.string().min(10)
})

export async function createPost(formData: FormData) {
  const result = schema.safeParse({
    title: formData.get('title'),
    content: formData.get('content')
  })
  if (!result.success) return { errors: result.error.flatten().fieldErrors }
  await db.post.create({ data: result.data })
  revalidatePath('/posts')
}
```

### Vue Composable

```typescript
import { ref, computed, watchEffect } from 'vue'

export function useUser(userId: Ref<string>) {
  const user = ref<User | null>(null)
  const loading = ref(false)
  const fullName = computed(() =>
    user.value ? `${user.value.firstName} ${user.value.lastName}` : ''
  )
  watchEffect(async () => {
    loading.value = true
    user.value = await fetchUser(userId.value)
    loading.value = false
  })
  return { user, loading, fullName }
}
```

### CVA Component

```tsx
import { cva, type VariantProps } from 'class-variance-authority'

const buttonVariants = cva(
  'inline-flex items-center justify-center rounded-md font-medium',
  {
    variants: {
      variant: {
        default: 'bg-primary text-white hover:bg-primary/90',
        outline: 'border border-input hover:bg-accent',
      },
      size: {
        sm: 'h-9 px-3 text-sm',
        default: 'h-10 px-4',
        lg: 'h-11 px-8',
      },
    },
    defaultVariants: { variant: 'default', size: 'default' },
  }
)

export function Button({ variant, size, children }: ButtonProps) {
  return <button className={buttonVariants({ variant, size })}>{children}</button>
}
```

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

Module files in modules/ directory contain detailed patterns.
- React: https://react.dev/
- Next.js: https://nextjs.org/docs
- Vue: https://vuejs.org/

---

Version: 2.0.0 | Last Updated: 2026-01-06
