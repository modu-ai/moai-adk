---
name: moai-domain-frontend
description: Frontend development specialist covering React 19, Next.js 16, Vue 3.5, and modern UI/UX patterns with component architecture
version: 1.0.0
category: domain
allowed-tools: Read, Write, Edit, Grep, Glob
tags:
 - frontend
 - react
 - nextjs
 - vue
 - ui
 - components
updated: 2025-12-06
status: active
author: MoAI-ADK Team
---

# Frontend Development Specialist

## Quick Reference (30 seconds)

Modern Frontend Development - Comprehensive frontend patterns covering React 19, Next.js 16, Vue 3.5, and modern UI/UX architecture.

Core Capabilities:
- React 19: Server components, concurrent features, optimized patterns
- Next.js 16: App router, server actions, advanced optimization
- 🟢 Vue 3.5: Composition API, TypeScript integration, reactivity
- Component Architecture: Design systems, component libraries, story-driven development
- Responsive Design: Mobile-first, accessibility, performance optimization

When to Use:
- Modern web application development
- Component library creation and management
- Performance optimization for frontend
- UI/UX implementation with accessibility
- Cross-platform frontend development

---

## Implementation Guide

### React 19 Server Components

Modern React Architecture:
```tsx
// app/components/UserProfile.tsx
import { cache } from 'react'
import { getUser } from '@/lib/users'

const getUserCached = cache(getUser)

interface UserProfileProps {
 userId: string
}

export default async function UserProfile({ userId }: UserProfileProps) {
 const user = await getUserCached(userId)

 return (
 <div className="user-profile">
 <h2>{user.name}</h2>
 <p>{user.email}</p>
 <ClientActions userId={userId} />
 </div>
 )
}

'use client'

function ClientActions({ userId }: { userId: string }) {
 const [isFollowing, setIsFollowing] = useState(false)

 return (
 <button onClick={() => setIsFollowing(!isFollowing)}>
 {isFollowing ? 'Unfollow' : 'Follow'}
 </button>
 )
}
```

Concurrent Features:
```tsx
import { Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'

function App() {
 return (
 <ErrorBoundary fallback={<div>Something went wrong</div>}>
 <Suspense fallback={<Loading />}>
 <UserProfile userId="123" />
 </Suspense>
 </ErrorBoundary>
 )
}
```

### Next.js 16 App Router

Server Actions and Data Fetching:
```tsx
// app/actions/users.ts
'use server'

import { revalidatePath } from 'next/cache'
import { redirect } from 'next/navigation'

export async function createUser(formData: FormData) {
 const name = formData.get('name') as string
 const email = formData.get('email') as string

 const user = await db.user.create({
 data: { name, email }
 })

 revalidatePath('/users')
 redirect(`/users/${user.id}`)
}

// app/users/page.tsx
import { createUser } from '../actions/users'

export default function UsersPage() {
 return (
 <form action={createUser}>
 <input name="name" placeholder="Name" required />
 <input name="email" type="email" placeholder="Email" required />
 <button type="submit">Create User</button>
 </form>
 )
}
```

Advanced Route Patterns:
```tsx
// app/[category]/[slug]/page.tsx
interface PageProps {
 params: { category: string; slug: string }
 searchParams: { [key: string]: string | string[] | undefined }
}

export default async function Page({ params, searchParams }: PageProps) {
 const { category, slug } = params
 const page = searchParams.page ?? '1'

 const data = await getData(category, slug, page)

 return <ContentComponent data={data} />
}

// Generate static params for performance
export async function generateStaticParams() {
 const posts = await getAllPosts()
 return posts.map((post) => ({
 category: post.category,
 slug: post.slug
 }))
}
```

### Vue 3.5 Composition API

Modern Vue Patterns:
```vue
<!-- components/UserCard.vue -->
<script setup lang="ts">
import { ref, computed, onMounted, watchEffect } from 'vue'

interface User {
 id: string
 name: string
 email: string
}

interface Props {
 userId: string
}

const props = defineProps<Props>()
const emit = defineEmits<{
 userLoaded: [user: User]
}>()

const user = ref<User | null>(null)
const loading = ref(true)
const error = ref<string | null>(null)

const initials = computed(() => {
 return user.value?.name
 .split(' ')
 .map(n => n[0])
 .join('')
 .toUpperCase() ?? '??'
})

const fetchUser = async () => {
 try {
 loading.value = true
 error.value = null
 user.value = await getUser(props.userId)
 emit('userLoaded', user.value)
 } catch (err) {
 error.value = err instanceof Error ? err.message : 'Failed to load user'
 } finally {
 loading.value = false
 }
}

watchEffect(() => {
 fetchUser()
})
</script>

<template>
 <div class="user-card">
 <div v-if="loading" class="loading">Loading...</div>
 <div v-else-if="error" class="error">{{ error }}</div>
 <div v-else-if="user" class="user-info">
 <div class="avatar">{{ initials }}</div>
 <div class="details">
 <h3>{{ user.name }}</h3>
 <p>{{ user.email }}</p>
 </div>
 </div>
 </div>
</template>

<style scoped>
.user-card {
 @apply border rounded-lg p-4 shadow-sm;
}

.avatar {
 @apply w-12 h-12 rounded-full bg-blue-500 text-white flex items-center justify-center font-semibold;
}
</style>
```

### Component Architecture

Design System Components:
```tsx
// components/Button/Button.tsx
import { forwardRef } from 'react'
import { cva, type VariantProps } from 'class-variance-authority'

const buttonVariants = cva(
 'inline-flex items-center justify-center rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:opacity-50 disabled:pointer-events-none',
 {
 variants: {
 variant: {
 default: 'bg-primary text-primary-foreground hover:bg-primary/90',
 destructive: 'bg-destructive text-destructive-foreground hover:bg-destructive/90',
 outline: 'border border-input hover:bg-accent hover:text-accent-foreground',
 secondary: 'bg-secondary text-secondary-foreground hover:bg-secondary/80',
 ghost: 'hover:bg-accent hover:text-accent-foreground',
 link: 'underline-offset-4 hover:underline text-primary',
 },
 size: {
 default: 'h-10 py-2 px-4',
 sm: 'h-9 px-3 rounded-md',
 lg: 'h-11 px-8 rounded-md',
 icon: 'h-10 w-10',
 },
 },
 defaultVariants: {
 variant: 'default',
 size: 'default',
 },
 }
)

export interface ButtonProps
 extends React.ButtonHTMLAttributes<HTMLButtonElement>,
 VariantProps<typeof buttonVariants> {
 asChild?: boolean
}

const Button = forwardRef<HTMLButtonElement, ButtonProps>(
 ({ className, variant, size, asChild = false, ...props }, ref) => {
 return (
 <button
 className={cn(buttonVariants({ variant, size, className }))}
 ref={ref}
 {...props}
 />
 )
 }
)
Button.displayName = 'Button'

export { Button, buttonVariants }
```

Compound Components Pattern:
```tsx
// components/Card/Card.tsx
import { createContext, useContext } from 'react'

interface CardContextValue {
 variant: 'default' | 'outlined'
}

const CardContext = createContext<CardContextValue>({
 variant: 'default'
})

interface CardProps {
 variant?: 'default' | 'outlined'
 children: React.ReactNode
}

export function Card({ variant = 'default', children }: CardProps) {
 return (
 <CardContext.Provider value={{ variant }}>
 <div className={`card card--${variant}`}>
 {children}
 </div>
 </CardContext.Provider>
 )
}

export function CardHeader({ children }: { children: React.ReactNode }) {
 return <div className="card__header">{children}</div>
}

export function CardContent({ children }: { children: React.ReactNode }) {
 return <div className="card__content">{children}</div>
}

export function CardFooter({ children }: { children: React.ReactNode }) {
 return <div className="card__footer">{children}</div>
}

// Usage
<Card variant="outlined">
 <CardHeader>
 <h3>Title</h3>
 </CardHeader>
 <CardContent>
 <p>Card content goes here</p>
 </CardContent>
 <CardFooter>
 <Button>Action</Button>
 </CardFooter>
</Card>
```

---

## Advanced Patterns

### Performance Optimization

React Optimization Patterns:
```tsx
import { memo, useMemo, useCallback, useMemo, useDeferredValue } from 'react'

const ExpensiveList = memo(({ items, onItemClick }: {
 items: Item[]
 onItemClick: (item: Item) => void
}) => {
 const expensiveValue = useMemo(() => {
 return items.reduce((sum, item) => sum + item.value, 0)
 }, [items])

 const handleClick = useCallback((item: Item) => {
 onItemClick(item)
 }, [onItemClick])

 return (
 <div>
 <p>Total: {expensiveValue}</p>
 {items.map(item => (
 <div key={item.id} onClick={() => handleClick(item)}>
 {item.name}
 </div>
 ))}
 </div>
 )
})
```

Next.js Performance:
```tsx
// Dynamic imports for code splitting
import dynamic from 'next/dynamic'

const DynamicChart = dynamic(
 () => import('@/components/Chart'),
 {
 loading: () => <div>Loading chart...</div>,
 ssr: false // Client-side only for heavy components
 }
)

// Image optimization
import Image from 'next/image'

function OptimizedImage({ src, alt, ...props }) {
 return (
 <Image
 src={src}
 alt={alt}
 {...props}
 placeholder="blur"
 blurDataURL="data:image/jpeg;base64,..."
 sizes="(max-width: 768px) 100vw, (max-width: 1200px) 50vw, 33vw"
 />
 )
}
```

### State Management

Zustand for Modern State Management:
```tsx
import { create } from 'zustand'
import { devtools, persist } from 'zustand/middleware'

interface UserState {
 user: User | null
 isLoading: boolean
 login: (email: string, password: string) => Promise<void>
 logout: () => void
 updateUser: (updates: Partial<User>) => void
}

export const useUserStore = create<UserState>()(
 devtools(
 persist(
 (set, get) => ({
 user: null,
 isLoading: false,

 login: async (email: string, password: string) => {
 set({ isLoading: true })
 try {
 const user = await authService.login(email, password)
 set({ user, isLoading: false })
 } catch (error) {
 set({ isLoading: false })
 throw error
 }
 },

 logout: () => {
 set({ user: null })
 },

 updateUser: (updates: Partial<User>) => {
 set(state => ({
 user: state.user ? { ...state.user, ...updates } : null
 }))
 }
 }),
 {
 name: 'user-storage',
 partialize: (state) => ({ user: state.user })
 }
 )
 )
)
```

---

## Works Well With

- moai-domain-backend - Full-stack development
- moai-library-shadcn - Component library integration
- moai-domain-uiux - UI/UX design principles
- moai-quality-security - Frontend security and accessibility
- moai-system-universal - Cross-platform optimization

---

## Technology Stack

Primary Technologies:
- Frameworks: React 19, Next.js 16, Vue 3.5, Nuxt 3
- Languages: TypeScript 5.9+, JavaScript ES2024
- Styling: Tailwind CSS 3.4+, CSS Modules, Styled Components
- State Management: Zustand, Redux Toolkit, Pinia
- Testing: Vitest, Testing Library, Playwright
- Build Tools: Vite 5, Turbopack, SWC

Component Libraries:
- shadcn/ui, Material-UI, Ant Design
- Headless UI, Radix UI
- Custom design systems

---

Status: Production Ready
Last Updated: 2025-11-30
Maintained by: MoAI-ADK Frontend Team
