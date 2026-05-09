# Frontend Development Examples

Multishot prompting examples demonstrating practical frontend patterns covering React 19, Next.js 16, Vue 3.5, and modern component architecture.

---

## Example 1: React 19 Server Components with Data Fetching

**Scenario**: Build a user dashboard with server-side data fetching and client interactivity.

**Input**:
```tsx
// Need: User dashboard showing profile and recent activity
// Requirements:
// - Server-side data fetching for performance
// - Client-side interactivity for actions
// - Proper loading and error states
```

**Output**:
```tsx
// app/dashboard/page.tsx
import { Suspense } from 'react'
import { cache } from 'react'
import { getUserProfile, getUserActivity } from '@/lib/api/users'
import { DashboardSkeleton } from '@/components/skeletons'
import { UserProfile } from './UserProfile'
import { ActivityFeed } from './ActivityFeed'
import { DashboardActions } from './DashboardActions'

// Cache data fetching for request deduplication
const getUser = cache(async (userId: string) => {
  return await getUserProfile(userId)
})

const getActivity = cache(async (userId: string) => {
  return await getUserActivity(userId, { limit: 10 })
})

interface DashboardPageProps {
  params: { userId: string }
}

export default async function DashboardPage({ params }: DashboardPageProps) {
  const { userId } = params

  return (
    <div className="container mx-auto py-8">
      <h1 className="text-3xl font-bold mb-8">Dashboard</h1>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Server Component - Profile */}
        <div className="lg:col-span-1">
          <Suspense fallback={<DashboardSkeleton type="profile" />}>
            <UserProfileSection userId={userId} />
          </Suspense>
        </div>

        {/* Server Component - Activity with streaming */}
        <div className="lg:col-span-2">
          <Suspense fallback={<DashboardSkeleton type="activity" />}>
            <ActivitySection userId={userId} />
          </Suspense>
        </div>
      </div>

      {/* Client Component - Interactive Actions */}
      <DashboardActions userId={userId} />
    </div>
  )
}

async function UserProfileSection({ userId }: { userId: string }) {
  const user = await getUser(userId)

  return <UserProfile user={user} />
}

async function ActivitySection({ userId }: { userId: string }) {
  const activity = await getActivity(userId)

  return <ActivityFeed activities={activity} />
}

// components/dashboard/DashboardActions.tsx
'use client'

import { useState, useTransition } from 'react'
import { refreshDashboard } from '@/app/actions/dashboard'
import { Button } from '@/components/ui/button'

interface DashboardActionsProps {
  userId: string
}

export function DashboardActions({ userId }: DashboardActionsProps) {
  const [isPending, startTransition] = useTransition()
  const [message, setMessage] = useState<string | null>(null)

  const handleRefresh = () => {
    startTransition(async () => {
      const result = await refreshDashboard(userId)
      setMessage(result.message)
    })
  }

  return (
    <div className="mt-6 flex items-center gap-4">
      <Button
        onClick={handleRefresh}
        disabled={isPending}
      >
        {isPending ? 'Refreshing...' : 'Refresh Dashboard'}
      </Button>

      {message && (
        <p className="text-sm text-muted-foreground">{message}</p>
      )}
    </div>
  )
}
```

**Explanation**: This pattern demonstrates React 19's server components for efficient data fetching, request deduplication with cache(), Suspense for streaming, and client components for interactivity.

---

## Example 2: Next.js 16 Server Actions with Form Handling

**Scenario**: Implement a user registration form with server-side validation.

**Input**:
```tsx
// Need: Registration form with:
// - Server-side validation
// - Real-time feedback
// - Progressive enhancement (works without JS)
```

**Output**:
```tsx
// app/register/page.tsx
import { RegisterForm } from './RegisterForm'

export default function RegisterPage() {
  return (
    <div className="max-w-md mx-auto py-12">
      <h1 className="text-2xl font-bold mb-6">Create Account</h1>
      <RegisterForm />
    </div>
  )
}

// app/register/RegisterForm.tsx
'use client'

import { useActionState } from 'react'
import { registerUser, type RegisterState } from '@/app/actions/auth'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Alert, AlertDescription } from '@/components/ui/alert'

const initialState: RegisterState = {
  success: false,
  errors: {}
}

export function RegisterForm() {
  const [state, formAction, isPending] = useActionState(
    registerUser,
    initialState
  )

  return (
    <form action={formAction} className="space-y-4">
      {state.errors?.general && (
        <Alert variant="destructive">
          <AlertDescription>{state.errors.general}</AlertDescription>
        </Alert>
      )}

      <div className="space-y-2">
        <Label htmlFor="name">Full Name</Label>
        <Input
          id="name"
          name="name"
          type="text"
          placeholder="John Doe"
          aria-describedby="name-error"
        />
        {state.errors?.name && (
          <p id="name-error" className="text-sm text-destructive">
            {state.errors.name}
          </p>
        )}
      </div>

      <div className="space-y-2">
        <Label htmlFor="email">Email</Label>
        <Input
          id="email"
          name="email"
          type="email"
          placeholder="john@example.com"
          aria-describedby="email-error"
        />
        {state.errors?.email && (
          <p id="email-error" className="text-sm text-destructive">
            {state.errors.email}
          </p>
        )}
      </div>

      <div className="space-y-2">
        <Label htmlFor="password">Password</Label>
        <Input
          id="password"
          name="password"
          type="password"
          minLength={8}
          aria-describedby="password-error"
        />
        {state.errors?.password && (
          <p id="password-error" className="text-sm text-destructive">
            {state.errors.password}
          </p>
        )}
      </div>

      <Button type="submit" className="w-full" disabled={isPending}>
        {isPending ? 'Creating Account...' : 'Create Account'}
      </Button>
    </form>
  )
}

// app/actions/auth.ts
'use server'

import { z } from 'zod'
import { redirect } from 'next/navigation'
import { createUser } from '@/lib/users'

const registerSchema = z.object({
  name: z.string().min(2, 'Name must be at least 2 characters'),
  email: z.string().email('Please enter a valid email'),
  password: z.string().min(8, 'Password must be at least 8 characters')
})

export interface RegisterState {
  success: boolean
  errors?: {
    name?: string
    email?: string
    password?: string
    general?: string
  }
}

export async function registerUser(
  prevState: RegisterState,
  formData: FormData
): Promise<RegisterState> {
  // Parse and validate
  const rawData = {
    name: formData.get('name'),
    email: formData.get('email'),
    password: formData.get('password')
  }

  const validatedFields = registerSchema.safeParse(rawData)

  if (!validatedFields.success) {
    const fieldErrors = validatedFields.error.flatten().fieldErrors
    return {
      success: false,
      errors: {
        name: fieldErrors.name?.[0],
        email: fieldErrors.email?.[0],
        password: fieldErrors.password?.[0]
      }
    }
  }

  try {
    await createUser(validatedFields.data)
  } catch (error) {
    if (error instanceof Error && error.message === 'EMAIL_EXISTS') {
      return {
        success: false,
        errors: { email: 'This email is already registered' }
      }
    }
    return {
      success: false,
      errors: { general: 'Something went wrong. Please try again.' }
    }
  }

  redirect('/login?registered=true')
}
```

**Explanation**: This pattern shows Next.js 16 server actions with useActionState for form handling, Zod validation, proper error messages, and progressive enhancement.

---

## Example 3: Vue 3.5 Composition API with TypeScript

**Scenario**: Create a reusable data table component with sorting and pagination.

**Input**:
```vue
<!-- Need: Reusable data table with:
- Type-safe props and emits
- Sorting by column
- Pagination
- Loading and empty states -->
```

**Output**:
```vue
<!-- components/DataTable.vue -->
<script setup lang="ts" generic="T extends Record<string, unknown>">
import { computed, ref } from 'vue'

interface Column<T> {
  key: keyof T
  label: string
  sortable?: boolean
  render?: (value: T[keyof T], row: T) => string
}

interface Props {
  data: T[]
  columns: Column<T>[]
  loading?: boolean
  pageSize?: number
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  pageSize: 10
})

const emit = defineEmits<{
  rowClick: [row: T]
  sort: [column: keyof T, direction: 'asc' | 'desc']
}>()

// Sorting state
const sortColumn = ref<keyof T | null>(null)
const sortDirection = ref<'asc' | 'desc'>('asc')

// Pagination state
const currentPage = ref(1)

// Computed: Sorted data
const sortedData = computed(() => {
  if (!sortColumn.value) return props.data

  return [...props.data].sort((a, b) => {
    const aVal = a[sortColumn.value!]
    const bVal = b[sortColumn.value!]

    if (aVal === bVal) return 0

    const comparison = aVal < bVal ? -1 : 1
    return sortDirection.value === 'asc' ? comparison : -comparison
  })
})

// Computed: Paginated data
const paginatedData = computed(() => {
  const start = (currentPage.value - 1) * props.pageSize
  return sortedData.value.slice(start, start + props.pageSize)
})

// Computed: Total pages
const totalPages = computed(() => {
  return Math.ceil(props.data.length / props.pageSize)
})

// Methods
function handleSort(column: Column<T>) {
  if (!column.sortable) return

  if (sortColumn.value === column.key) {
    sortDirection.value = sortDirection.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortColumn.value = column.key
    sortDirection.value = 'asc'
  }

  emit('sort', column.key, sortDirection.value)
}

function handleRowClick(row: T) {
  emit('rowClick', row)
}

function goToPage(page: number) {
  if (page >= 1 && page <= totalPages.value) {
    currentPage.value = page
  }
}

function getCellValue(row: T, column: Column<T>): string {
  const value = row[column.key]
  if (column.render) {
    return column.render(value, row)
  }
  return String(value ?? '')
}
</script>

<template>
  <div class="data-table">
    <!-- Loading overlay -->
    <div v-if="loading" class="loading-overlay">
      <div class="spinner" />
    </div>

    <!-- Table -->
    <table class="w-full">
      <thead>
        <tr class="border-b">
          <th
            v-for="column in columns"
            :key="String(column.key)"
            class="px-4 py-3 text-left"
            :class="{ 'cursor-pointer hover:bg-muted': column.sortable }"
            @click="handleSort(column)"
          >
            <div class="flex items-center gap-2">
              {{ column.label }}
              <span v-if="column.sortable && sortColumn === column.key">
                {{ sortDirection === 'asc' ? '↑' : '↓' }}
              </span>
            </div>
          </th>
        </tr>
      </thead>

      <tbody>
        <!-- Empty state -->
        <tr v-if="!loading && data.length === 0">
          <td :colspan="columns.length" class="px-4 py-8 text-center text-muted-foreground">
            No data available
          </td>
        </tr>

        <!-- Data rows -->
        <tr
          v-for="(row, index) in paginatedData"
          :key="index"
          class="border-b hover:bg-muted/50 cursor-pointer"
          @click="handleRowClick(row)"
        >
          <td
            v-for="column in columns"
            :key="String(column.key)"
            class="px-4 py-3"
          >
            {{ getCellValue(row, column) }}
          </td>
        </tr>
      </tbody>
    </table>

    <!-- Pagination -->
    <div v-if="totalPages > 1" class="flex items-center justify-between px-4 py-3 border-t">
      <div class="text-sm text-muted-foreground">
        Showing {{ (currentPage - 1) * pageSize + 1 }} to
        {{ Math.min(currentPage * pageSize, data.length) }} of {{ data.length }}
      </div>

      <div class="flex gap-1">
        <button
          class="px-3 py-1 rounded hover:bg-muted disabled:opacity-50"
          :disabled="currentPage === 1"
          @click="goToPage(currentPage - 1)"
        >
          Previous
        </button>

        <button
          v-for="page in totalPages"
          :key="page"
          class="px-3 py-1 rounded"
          :class="page === currentPage ? 'bg-primary text-primary-foreground' : 'hover:bg-muted'"
          @click="goToPage(page)"
        >
          {{ page }}
        </button>

        <button
          class="px-3 py-1 rounded hover:bg-muted disabled:opacity-50"
          :disabled="currentPage === totalPages"
          @click="goToPage(currentPage + 1)"
        >
          Next
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.data-table {
  @apply relative overflow-hidden rounded-lg border;
}

.loading-overlay {
  @apply absolute inset-0 bg-background/80 flex items-center justify-center z-10;
}

.spinner {
  @apply w-8 h-8 border-4 border-primary border-t-transparent rounded-full animate-spin;
}
</style>

<!-- Usage Example -->
<!--
<script setup lang="ts">
import DataTable from '@/components/DataTable.vue'

interface User {
  id: number
  name: string
  email: string
  status: 'active' | 'inactive'
}

const users: User[] = [
  { id: 1, name: 'John Doe', email: 'john@example.com', status: 'active' },
  { id: 2, name: 'Jane Smith', email: 'jane@example.com', status: 'inactive' }
]

const columns = [
  { key: 'name' as const, label: 'Name', sortable: true },
  { key: 'email' as const, label: 'Email', sortable: true },
  {
    key: 'status' as const,
    label: 'Status',
    render: (value: 'active' | 'inactive') =>
      value === 'active' ? 'Active' : 'Inactive'
  }
]

function handleRowClick(user: User) {
  console.log('Clicked:', user)
}
</script>

<template>
  <DataTable
    :data="users"
    :columns="columns"
    :page-size="10"
    @row-click="handleRowClick"
  />
</template>
-->
```

**Explanation**: This Vue 3.5 pattern demonstrates generic components with TypeScript, the Composition API with computed properties, and comprehensive data table functionality.

---

## Common Patterns

### Pattern 1: Compound Components

Build flexible, composable component APIs:

```tsx
// components/Card/index.tsx
import { createContext, useContext, ReactNode } from 'react'

interface CardContextValue {
  variant: 'default' | 'outlined' | 'elevated'
}

const CardContext = createContext<CardContextValue>({ variant: 'default' })

interface CardProps {
  variant?: 'default' | 'outlined' | 'elevated'
  children: ReactNode
  className?: string
}

export function Card({ variant = 'default', children, className }: CardProps) {
  return (
    <CardContext.Provider value={{ variant }}>
      <div className={cn('rounded-lg', variantStyles[variant], className)}>
        {children}
      </div>
    </CardContext.Provider>
  )
}

export function CardHeader({ children, className }: { children: ReactNode; className?: string }) {
  return (
    <div className={cn('px-6 py-4 border-b', className)}>
      {children}
    </div>
  )
}

export function CardTitle({ children, className }: { children: ReactNode; className?: string }) {
  return (
    <h3 className={cn('text-lg font-semibold', className)}>
      {children}
    </h3>
  )
}

export function CardContent({ children, className }: { children: ReactNode; className?: string }) {
  return (
    <div className={cn('px-6 py-4', className)}>
      {children}
    </div>
  )
}

export function CardFooter({ children, className }: { children: ReactNode; className?: string }) {
  const { variant } = useContext(CardContext)
  return (
    <div className={cn('px-6 py-4 border-t', className)}>
      {children}
    </div>
  )
}

// Usage
<Card variant="elevated">
  <CardHeader>
    <CardTitle>Card Title</CardTitle>
  </CardHeader>
  <CardContent>
    <p>Card content goes here.</p>
  </CardContent>
  <CardFooter>
    <Button>Action</Button>
  </CardFooter>
</Card>
```

### Pattern 2: Custom Hooks for Data Fetching

```tsx
// hooks/useQuery.ts
import { useState, useEffect, useCallback } from 'react'

interface UseQueryOptions<T> {
  enabled?: boolean
  refetchInterval?: number
  onSuccess?: (data: T) => void
  onError?: (error: Error) => void
}

interface UseQueryResult<T> {
  data: T | null
  isLoading: boolean
  isError: boolean
  error: Error | null
  refetch: () => Promise<void>
}

export function useQuery<T>(
  queryFn: () => Promise<T>,
  options: UseQueryOptions<T> = {}
): UseQueryResult<T> {
  const { enabled = true, refetchInterval, onSuccess, onError } = options

  const [data, setData] = useState<T | null>(null)
  const [isLoading, setIsLoading] = useState(enabled)
  const [error, setError] = useState<Error | null>(null)

  const fetchData = useCallback(async () => {
    setIsLoading(true)
    setError(null)

    try {
      const result = await queryFn()
      setData(result)
      onSuccess?.(result)
    } catch (err) {
      const error = err instanceof Error ? err : new Error('Unknown error')
      setError(error)
      onError?.(error)
    } finally {
      setIsLoading(false)
    }
  }, [queryFn, onSuccess, onError])

  useEffect(() => {
    if (enabled) {
      fetchData()
    }
  }, [enabled, fetchData])

  useEffect(() => {
    if (refetchInterval && enabled) {
      const interval = setInterval(fetchData, refetchInterval)
      return () => clearInterval(interval)
    }
  }, [refetchInterval, enabled, fetchData])

  return {
    data,
    isLoading,
    isError: error !== null,
    error,
    refetch: fetchData
  }
}

// Usage
function UserProfile({ userId }: { userId: string }) {
  const { data: user, isLoading, error } = useQuery(
    () => fetchUser(userId),
    {
      enabled: !!userId,
      onSuccess: (user) => console.log('User loaded:', user.name)
    }
  )

  if (isLoading) return <Skeleton />
  if (error) return <ErrorMessage error={error} />

  return <div>{user?.name}</div>
}
```

### Pattern 3: Performance Optimization with Memoization

```tsx
import { memo, useMemo, useCallback, useDeferredValue } from 'react'

interface SearchListProps {
  items: Item[]
  searchTerm: string
  onSelect: (item: Item) => void
}

export const SearchList = memo(function SearchList({
  items,
  searchTerm,
  onSelect
}: SearchListProps) {
  // Defer search for smoother typing
  const deferredSearchTerm = useDeferredValue(searchTerm)

  // Memoize filtered results
  const filteredItems = useMemo(() => {
    return items.filter(item =>
      item.name.toLowerCase().includes(deferredSearchTerm.toLowerCase())
    )
  }, [items, deferredSearchTerm])

  // Memoize callback
  const handleSelect = useCallback((item: Item) => {
    onSelect(item)
  }, [onSelect])

  // Show stale indicator while updating
  const isStale = searchTerm !== deferredSearchTerm

  return (
    <div className={isStale ? 'opacity-70' : ''}>
      {filteredItems.map(item => (
        <SearchItem
          key={item.id}
          item={item}
          onSelect={handleSelect}
        />
      ))}
    </div>
  )
})

// Memoized child component
const SearchItem = memo(function SearchItem({
  item,
  onSelect
}: {
  item: Item
  onSelect: (item: Item) => void
}) {
  return (
    <div onClick={() => onSelect(item)}>
      {item.name}
    </div>
  )
})
```

---

## Anti-Patterns (Patterns to Avoid)

### Anti-Pattern 1: Prop Drilling

**Problem**: Passing props through many intermediate components.

```tsx
// Incorrect approach
function App() {
  const [user, setUser] = useState<User | null>(null)
  return <Layout user={user} setUser={setUser} />
}

function Layout({ user, setUser }) {
  return <Sidebar user={user} setUser={setUser} />
}

function Sidebar({ user, setUser }) {
  return <UserMenu user={user} setUser={setUser} />
}
```

**Solution**: Use context or state management.

```tsx
// Correct approach
const UserContext = createContext<{
  user: User | null
  setUser: (user: User | null) => void
} | null>(null)

function UserProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  return (
    <UserContext.Provider value={{ user, setUser }}>
      {children}
    </UserContext.Provider>
  )
}

function useUser() {
  const context = useContext(UserContext)
  if (!context) throw new Error('useUser must be used within UserProvider')
  return context
}

function UserMenu() {
  const { user, setUser } = useUser()
  // Direct access without prop drilling
}
```

### Anti-Pattern 2: Inline Function Definitions

**Problem**: Creating new function references on every render.

```tsx
// Incorrect approach
function List({ items }) {
  return items.map(item => (
    // New function created on every render
    <Item key={item.id} onClick={() => handleClick(item.id)} />
  ))
}
```

**Solution**: Use useCallback or component-level handlers.

```tsx
// Correct approach
function List({ items }) {
  const handleClick = useCallback((id: string) => {
    // Handle click
  }, [])

  return items.map(item => (
    <Item
      key={item.id}
      id={item.id}
      onClick={handleClick}
    />
  ))
}

// Or with proper memoization
const Item = memo(function Item({
  id,
  onClick
}: {
  id: string
  onClick: (id: string) => void
}) {
  return <div onClick={() => onClick(id)}>Item {id}</div>
})
```

### Anti-Pattern 3: Fetching in useEffect

**Problem**: Client-side fetching when server-side would be better.

```tsx
// Incorrect approach for Next.js
'use client'
function UserPage({ userId }: { userId: string }) {
  const [user, setUser] = useState(null)

  useEffect(() => {
    fetch(`/api/users/${userId}`)
      .then(res => res.json())
      .then(setUser)
  }, [userId])

  return <div>{user?.name}</div>
}
```

**Solution**: Use server components for initial data.

```tsx
// Correct approach
// app/users/[userId]/page.tsx
async function UserPage({ params }: { params: { userId: string } }) {
  const user = await getUser(params.userId)
  return <UserProfile user={user} />
}
```

---

## Accessibility Patterns

### Focus Management

```tsx
function Modal({ isOpen, onClose, children }: ModalProps) {
  const modalRef = useRef<HTMLDivElement>(null)
  const previousActiveElement = useRef<HTMLElement | null>(null)

  useEffect(() => {
    if (isOpen) {
      previousActiveElement.current = document.activeElement as HTMLElement
      modalRef.current?.focus()
    } else {
      previousActiveElement.current?.focus()
    }
  }, [isOpen])

  return (
    <div
      ref={modalRef}
      role="dialog"
      aria-modal="true"
      tabIndex={-1}
      onKeyDown={(e) => {
        if (e.key === 'Escape') onClose()
      }}
    >
      {children}
    </div>
  )
}
```

---

*For additional patterns and framework-specific optimizations, see the related skills and documentation.*
