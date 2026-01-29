# Frontend Development Reference

## API Reference

### React 19 Hooks

Server Components Data Fetching:
```tsx
import { cache } from 'react'
import { use } from 'react'

// Cached data fetching function
const getUser = cache(async (id: string) => {
  const response = await fetch(`/api/users/${id}`)
  return response.json()
})

// Server Component
async function UserProfile({ userId }: { userId: string }) {
  const user = await getUser(userId)
  return <div>{user.name}</div>
}

// Client Component with use()
'use client'
function UserData({ userPromise }: { userPromise: Promise<User> }) {
  const user = use(userPromise)
  return <div>{user.name}</div>
}
```

useOptimistic Hook:
```tsx
'use client'
import { useOptimistic, useTransition } from 'react'

function CommentList({ comments, addComment }: Props) {
  const [isPending, startTransition] = useTransition()
  const [optimisticComments, addOptimistic] = useOptimistic(
    comments,
    (state, newComment: Comment) => [...state, newComment]
  )

  async function handleSubmit(formData: FormData) {
    const newComment = { text: formData.get('text'), pending: true }
    addOptimistic(newComment)

    startTransition(async () => {
      await addComment(formData)
    })
  }

  return (
    <form action={handleSubmit}>
      <input name="text" />
      <button type="submit">Add</button>
      {optimisticComments.map(comment => (
        <Comment key={comment.id} {...comment} />
      ))}
    </form>
  )
}
```

useFormStatus Hook:
```tsx
'use client'
import { useFormStatus } from 'react-dom'

function SubmitButton() {
  const { pending, data, method, action } = useFormStatus()

  return (
    <button type="submit" disabled={pending}>
      {pending ? 'Submitting...' : 'Submit'}
    </button>
  )
}
```

### Next.js 16 APIs

Server Actions:
```tsx
// app/actions.ts
'use server'

import { revalidatePath, revalidateTag } from 'next/cache'
import { redirect } from 'next/navigation'
import { cookies } from 'next/headers'

export async function createPost(formData: FormData) {
  const title = formData.get('title') as string
  const content = formData.get('content') as string

  const post = await db.post.create({
    data: { title, content }
  })

  revalidatePath('/posts')
  revalidateTag('posts')
  redirect(`/posts/${post.id}`)
}

export async function updateUser(userId: string, data: UserUpdateData) {
  const cookieStore = cookies()
  const token = cookieStore.get('auth-token')

  if (!token) {
    throw new Error('Unauthorized')
  }

  return await db.user.update({
    where: { id: userId },
    data
  })
}
```

Parallel Routes:
```tsx
// app/dashboard/layout.tsx
export default function DashboardLayout({
  children,
  analytics,
  team,
}: {
  children: React.ReactNode
  analytics: React.ReactNode
  team: React.ReactNode
}) {
  return (
    <div className="dashboard">
      <main>{children}</main>
      <aside className="sidebar">
        {analytics}
        {team}
      </aside>
    </div>
  )
}

// app/dashboard/@analytics/page.tsx
async function AnalyticsPanel() {
  const data = await getAnalytics()
  return <AnalyticsChart data={data} />
}

// app/dashboard/@team/page.tsx
async function TeamPanel() {
  const members = await getTeamMembers()
  return <TeamList members={members} />
}
```

Intercepting Routes:
```tsx
// app/feed/@modal/(.)photo/[id]/page.tsx
import { Modal } from '@/components/modal'
import { getPhoto } from '@/lib/photos'

export default async function PhotoModal({
  params: { id },
}: {
  params: { id: string }
}) {
  const photo = await getPhoto(id)
  return (
    <Modal>
      <img src={photo.url} alt={photo.title} />
    </Modal>
  )
}
```

### Vue 3.5 Composition API

Composables:
```typescript
// composables/useUser.ts
import { ref, computed, watch, onMounted } from 'vue'

export function useUser(userId: Ref<string>) {
  const user = ref<User | null>(null)
  const loading = ref(false)
  const error = ref<Error | null>(null)

  const fullName = computed(() => {
    if (!user.value) return ''
    return `${user.value.firstName} ${user.value.lastName}`
  })

  async function fetchUser() {
    loading.value = true
    error.value = null
    try {
      user.value = await api.getUser(userId.value)
    } catch (e) {
      error.value = e as Error
    } finally {
      loading.value = false
    }
  }

  watch(userId, fetchUser, { immediate: true })

  return { user, loading, error, fullName, refetch: fetchUser }
}

// Usage in component
const { user, loading, error } = useUser(toRef(props, 'userId'))
```

Provide/Inject Pattern:
```typescript
// context/theme.ts
import { provide, inject, ref, readonly } from 'vue'
import type { InjectionKey, Ref } from 'vue'

interface ThemeContext {
  theme: Ref<'light' | 'dark'>
  toggleTheme: () => void
}

const ThemeKey: InjectionKey<ThemeContext> = Symbol('theme')

export function provideTheme() {
  const theme = ref<'light' | 'dark'>('light')

  function toggleTheme() {
    theme.value = theme.value === 'light' ? 'dark' : 'light'
  }

  provide(ThemeKey, {
    theme: readonly(theme),
    toggleTheme
  })

  return { theme, toggleTheme }
}

export function useTheme() {
  const context = inject(ThemeKey)
  if (!context) {
    throw new Error('useTheme must be used within ThemeProvider')
  }
  return context
}
```

---

## Configuration Options

### Next.js Configuration

```javascript
// next.config.js
/** @type {import('next').NextConfig} */
const nextConfig = {
  experimental: {
    ppr: true,                    // Partial Prerendering
    reactCompiler: true,          // React Compiler
    serverActions: {
      bodySizeLimit: '2mb',
      allowedOrigins: ['my-domain.com']
    }
  },

  images: {
    remotePatterns: [
      {
        protocol: 'https',
        hostname: 'cdn.example.com',
        port: '',
        pathname: '/images/**'
      }
    ],
    formats: ['image/avif', 'image/webp'],
    deviceSizes: [640, 750, 828, 1080, 1200, 1920, 2048, 3840],
    imageSizes: [16, 32, 48, 64, 96, 128, 256, 384]
  },

  compiler: {
    removeConsole: process.env.NODE_ENV === 'production',
    reactRemoveProperties: true
  },

  headers: async () => [
    {
      source: '/:path*',
      headers: [
        { key: 'X-Frame-Options', value: 'DENY' },
        { key: 'X-Content-Type-Options', value: 'nosniff' },
        { key: 'Referrer-Policy', value: 'strict-origin-when-cross-origin' }
      ]
    }
  ],

  redirects: async () => [
    {
      source: '/old-path',
      destination: '/new-path',
      permanent: true
    }
  ]
}

module.exports = nextConfig
```

### Vite Configuration

```typescript
// vite.config.ts
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react-swc'
import path from 'path'

export default defineConfig({
  plugins: [
    react({
      jsxImportSource: '@emotion/react'
    })
  ],

  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
      '@components': path.resolve(__dirname, './src/components'),
      '@hooks': path.resolve(__dirname, './src/hooks'),
      '@utils': path.resolve(__dirname, './src/utils')
    }
  },

  build: {
    target: 'es2022',
    minify: 'terser',
    sourcemap: true,
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['react', 'react-dom'],
          router: ['react-router-dom'],
          ui: ['@radix-ui/react-dialog', '@radix-ui/react-dropdown-menu']
        }
      }
    }
  },

  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8000',
        changeOrigin: true
      }
    }
  },

  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: './src/test/setup.ts',
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html']
    }
  }
})
```

### Tailwind CSS Configuration

```javascript
// tailwind.config.js
/** @type {import('tailwindcss').Config} */
module.exports = {
  darkMode: 'class',
  content: [
    './src/**/*.{js,ts,jsx,tsx,mdx}',
    './components/**/*.{js,ts,jsx,tsx,mdx}'
  ],
  theme: {
    extend: {
      colors: {
        border: 'hsl(var(--border))',
        background: 'hsl(var(--background))',
        foreground: 'hsl(var(--foreground))',
        primary: {
          DEFAULT: 'hsl(var(--primary))',
          foreground: 'hsl(var(--primary-foreground))'
        },
        secondary: {
          DEFAULT: 'hsl(var(--secondary))',
          foreground: 'hsl(var(--secondary-foreground))'
        }
      },
      borderRadius: {
        lg: 'var(--radius)',
        md: 'calc(var(--radius) - 2px)',
        sm: 'calc(var(--radius) - 4px)'
      },
      keyframes: {
        'accordion-down': {
          from: { height: 0 },
          to: { height: 'var(--radix-accordion-content-height)' }
        },
        'accordion-up': {
          from: { height: 'var(--radix-accordion-content-height)' },
          to: { height: 0 }
        }
      },
      animation: {
        'accordion-down': 'accordion-down 0.2s ease-out',
        'accordion-up': 'accordion-up 0.2s ease-out'
      }
    }
  },
  plugins: [
    require('tailwindcss-animate'),
    require('@tailwindcss/typography'),
    require('@tailwindcss/forms')
  ]
}
```

---

## Integration Patterns

### State Management with Zustand

```typescript
// stores/useStore.ts
import { create } from 'zustand'
import { devtools, persist, subscribeWithSelector } from 'zustand/middleware'
import { immer } from 'zustand/middleware/immer'

interface AppState {
  user: User | null
  theme: 'light' | 'dark'
  notifications: Notification[]

  // Actions
  setUser: (user: User | null) => void
  toggleTheme: () => void
  addNotification: (notification: Notification) => void
  removeNotification: (id: string) => void
}

export const useStore = create<AppState>()(
  devtools(
    persist(
      subscribeWithSelector(
        immer((set, get) => ({
          user: null,
          theme: 'light',
          notifications: [],

          setUser: (user) => set({ user }),

          toggleTheme: () =>
            set((state) => {
              state.theme = state.theme === 'light' ? 'dark' : 'light'
            }),

          addNotification: (notification) =>
            set((state) => {
              state.notifications.push(notification)
            }),

          removeNotification: (id) =>
            set((state) => {
              state.notifications = state.notifications.filter(n => n.id !== id)
            })
        }))
      ),
      {
        name: 'app-storage',
        partialize: (state) => ({ user: state.user, theme: state.theme })
      }
    )
  )
)

// Selectors
export const useUser = () => useStore((state) => state.user)
export const useTheme = () => useStore((state) => state.theme)
```

### React Query Integration

```typescript
// lib/queries.ts
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'

export const userKeys = {
  all: ['users'] as const,
  lists: () => [...userKeys.all, 'list'] as const,
  list: (filters: UserFilters) => [...userKeys.lists(), filters] as const,
  details: () => [...userKeys.all, 'detail'] as const,
  detail: (id: string) => [...userKeys.details(), id] as const
}

export function useUsers(filters: UserFilters) {
  return useQuery({
    queryKey: userKeys.list(filters),
    queryFn: () => api.getUsers(filters),
    staleTime: 5 * 60 * 1000, // 5 minutes
    gcTime: 30 * 60 * 1000,   // 30 minutes
  })
}

export function useUser(id: string) {
  return useQuery({
    queryKey: userKeys.detail(id),
    queryFn: () => api.getUser(id),
    enabled: !!id
  })
}

export function useUpdateUser() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UserUpdate }) =>
      api.updateUser(id, data),
    onMutate: async ({ id, data }) => {
      await queryClient.cancelQueries({ queryKey: userKeys.detail(id) })
      const previous = queryClient.getQueryData(userKeys.detail(id))
      queryClient.setQueryData(userKeys.detail(id), (old: User) => ({
        ...old,
        ...data
      }))
      return { previous }
    },
    onError: (err, { id }, context) => {
      queryClient.setQueryData(userKeys.detail(id), context?.previous)
    },
    onSettled: (data, error, { id }) => {
      queryClient.invalidateQueries({ queryKey: userKeys.detail(id) })
    }
  })
}
```

---

## Troubleshooting

### React Issues

Issue: Hydration mismatch errors
Symptoms: Console warnings about hydration, content flicker
Solution:
- Ensure server and client render identical content
- Use suppressHydrationWarning for dynamic content
- Wrap browser-only code in useEffect
- Use dynamic imports with ssr: false for client-only components

Issue: Infinite re-renders
Symptoms: Maximum update depth exceeded error
Solution:
- Check useEffect dependencies for object/array references
- Use useMemo/useCallback for reference stability
- Avoid setting state unconditionally in effects
- Use functional state updates when depending on previous state

Issue: Memory leaks
Symptoms: "Can't perform state update on unmounted component"
Solution:
- Use AbortController for fetch requests
- Cleanup subscriptions in useEffect return
- Use refs to track component mount state
- Cancel async operations on unmount

### Next.js Issues

Issue: Build fails with "Module not found"
Symptoms: Import errors during build
Solution:
- Check for circular dependencies
- Verify file extensions and case sensitivity
- Ensure server-only code isn't imported in client components
- Use next/dynamic for conditional imports

Issue: Slow page loads
Symptoms: High TTFB, large bundle sizes
Solution:
- Analyze bundle with @next/bundle-analyzer
- Use dynamic imports for large components
- Implement proper caching strategies
- Enable PPR for partial prerendering

Issue: Server Action errors
Symptoms: "Error: Functions cannot be passed directly to Client Components"
Solution:
- Only pass serializable props to client components
- Use bind() to pass additional arguments to server actions
- Move complex logic to server-side utilities
- Use proper error boundaries

### Performance Issues

Issue: Slow initial render
Symptoms: High LCP, layout shifts
Solution:
- Prioritize above-the-fold content
- Use next/image with priority for hero images
- Implement skeleton loading states
- Optimize font loading with next/font

Issue: Janky animations
Symptoms: Dropped frames, stuttering
Solution:
- Use CSS transforms instead of layout properties
- Add will-change for animated elements
- Use requestAnimationFrame for JS animations
- Reduce paint complexity with layer promotion

---

## External Resources

### React
- React Documentation: https://react.dev/
- React 19 Changelog: https://react.dev/blog/2024/04/25/react-19
- React Compiler: https://react.dev/learn/react-compiler
- Server Components: https://react.dev/reference/rsc/server-components

### Next.js
- Next.js Documentation: https://nextjs.org/docs
- App Router: https://nextjs.org/docs/app
- Server Actions: https://nextjs.org/docs/app/api-reference/functions/server-actions
- Vercel AI SDK: https://sdk.vercel.ai/docs

### Vue
- Vue 3 Documentation: https://vuejs.org/
- Vue Composition API: https://vuejs.org/guide/extras/composition-api-faq.html
- Pinia: https://pinia.vuejs.org/
- Nuxt 3: https://nuxt.com/docs

### Styling
- Tailwind CSS: https://tailwindcss.com/docs
- shadcn/ui: https://ui.shadcn.com/
- Radix UI: https://www.radix-ui.com/primitives/docs
- CSS Modules: https://github.com/css-modules/css-modules

### Testing
- Vitest: https://vitest.dev/
- Testing Library: https://testing-library.com/
- Playwright: https://playwright.dev/
- Cypress: https://www.cypress.io/

### Performance
- Web Vitals: https://web.dev/vitals/
- Lighthouse: https://developer.chrome.com/docs/lighthouse/
- Bundle Analyzer: https://www.npmjs.com/package/@next/bundle-analyzer

---

Version: 1.0.0
Last Updated: 2025-12-06
