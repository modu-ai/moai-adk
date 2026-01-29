# Vue 3.5 Patterns

Comprehensive patterns for Vue 3.5 development covering Composition API, Reactivity system, TypeScript integration, and Pinia state management.

---

## Composition API Fundamentals

### Script Setup with TypeScript

```vue
<!-- components/UserCard.vue -->
<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'

// Type definitions
interface User {
  id: string
  name: string
  email: string
  avatar: string
  role: 'admin' | 'user' | 'guest'
}

interface Props {
  userId: string
  showActions?: boolean
}

interface Emits {
  (e: 'update', user: User): void
  (e: 'delete', userId: string): void
  (e: 'select', user: User): void
}

// Props with defaults
const props = withDefaults(defineProps<Props>(), {
  showActions: true
})

// Type-safe emits
const emit = defineEmits<Emits>()

// Expose methods to parent
defineExpose({
  refresh: () => fetchUser()
})

// Reactive state
const user = ref<User | null>(null)
const loading = ref(true)
const error = ref<string | null>(null)

// Computed properties
const displayName = computed(() => {
  if (!user.value) return 'Unknown'
  return user.value.name || user.value.email
})

const isAdmin = computed(() => user.value?.role === 'admin')

// Methods
async function fetchUser() {
  loading.value = true
  error.value = null

  try {
    const response = await fetch(`/api/users/${props.userId}`)
    if (!response.ok) throw new Error('Failed to fetch user')
    user.value = await response.json()
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Unknown error'
  } finally {
    loading.value = false
  }
}

function handleSelect() {
  if (user.value) {
    emit('select', user.value)
  }
}

function handleDelete() {
  emit('delete', props.userId)
}

// Watchers
watch(
  () => props.userId,
  (newId, oldId) => {
    if (newId !== oldId) {
      fetchUser()
    }
  },
  { immediate: true }
)

// Lifecycle
onMounted(() => {
  console.log('UserCard mounted')
})

onUnmounted(() => {
  console.log('UserCard unmounted')
})
</script>

<template>
  <div class="user-card" :class="{ 'is-admin': isAdmin }">
    <!-- Loading state -->
    <div v-if="loading" class="loading">
      <Spinner />
    </div>

    <!-- Error state -->
    <div v-else-if="error" class="error">
      <p>{{ error }}</p>
      <button @click="fetchUser">Retry</button>
    </div>

    <!-- User content -->
    <div v-else-if="user" class="user-content" @click="handleSelect">
      <img :src="user.avatar" :alt="displayName" class="avatar" />

      <div class="info">
        <h3 class="name">{{ displayName }}</h3>
        <p class="email">{{ user.email }}</p>
        <span class="role" :class="`role-${user.role}`">
          {{ user.role }}
        </span>
      </div>

      <div v-if="showActions" class="actions">
        <button @click.stop="$emit('update', user)">Edit</button>
        <button @click.stop="handleDelete" class="danger">Delete</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.user-card {
  @apply border rounded-lg p-4 hover:shadow-md transition-shadow;
}

.user-card.is-admin {
  @apply border-blue-500;
}

.avatar {
  @apply w-16 h-16 rounded-full object-cover;
}

.role-admin {
  @apply bg-blue-100 text-blue-800;
}

.role-user {
  @apply bg-green-100 text-green-800;
}

.danger {
  @apply text-red-600 hover:text-red-800;
}
</style>
```

---

## Composables (Reusable Logic)

### Data Fetching Composable

```typescript
// composables/useFetch.ts
import { ref, shallowRef, watchEffect, toValue, type MaybeRefOrGetter } from 'vue'

interface UseFetchOptions<T> {
  immediate?: boolean
  refetch?: boolean
  initialData?: T
  onSuccess?: (data: T) => void
  onError?: (error: Error) => void
}

interface UseFetchReturn<T> {
  data: Ref<T | null>
  error: Ref<Error | null>
  loading: Ref<boolean>
  execute: () => Promise<void>
  refresh: () => Promise<void>
}

export function useFetch<T>(
  url: MaybeRefOrGetter<string>,
  options: UseFetchOptions<T> = {}
): UseFetchReturn<T> {
  const {
    immediate = true,
    refetch = false,
    initialData = null,
    onSuccess,
    onError
  } = options

  const data = shallowRef<T | null>(initialData as T | null)
  const error = ref<Error | null>(null)
  const loading = ref(false)

  async function execute() {
    loading.value = true
    error.value = null

    try {
      const response = await fetch(toValue(url))

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`)
      }

      const result = await response.json()
      data.value = result
      onSuccess?.(result)
    } catch (err) {
      const fetchError = err instanceof Error ? err : new Error('Fetch failed')
      error.value = fetchError
      onError?.(fetchError)
    } finally {
      loading.value = false
    }
  }

  if (refetch) {
    watchEffect(() => {
      toValue(url) // Track URL changes
      execute()
    })
  } else if (immediate) {
    execute()
  }

  return {
    data,
    error,
    loading,
    execute,
    refresh: execute
  }
}

// Usage
const { data: users, loading, error, refresh } = useFetch<User[]>('/api/users')
```

### Form Handling Composable

```typescript
// composables/useForm.ts
import { reactive, ref, computed } from 'vue'
import { z, type ZodSchema } from 'zod'

interface UseFormOptions<T> {
  initialValues: T
  validationSchema?: ZodSchema<T>
  onSubmit: (values: T) => Promise<void>
}

export function useForm<T extends Record<string, unknown>>(
  options: UseFormOptions<T>
) {
  const { initialValues, validationSchema, onSubmit } = options

  const values = reactive({ ...initialValues }) as T
  const errors = reactive<Partial<Record<keyof T, string>>>({})
  const touched = reactive<Partial<Record<keyof T, boolean>>>({})
  const submitting = ref(false)
  const submitted = ref(false)

  const isValid = computed(() => {
    if (!validationSchema) return true
    const result = validationSchema.safeParse(values)
    return result.success
  })

  const isDirty = computed(() => {
    return Object.keys(initialValues).some(
      key => values[key as keyof T] !== initialValues[key as keyof T]
    )
  })

  function validate(): boolean {
    if (!validationSchema) return true

    const result = validationSchema.safeParse(values)

    // Clear all errors
    Object.keys(errors).forEach(key => {
      delete errors[key as keyof T]
    })

    if (!result.success) {
      result.error.errors.forEach(err => {
        const path = err.path[0] as keyof T
        if (path) {
          errors[path] = err.message
        }
      })
      return false
    }

    return true
  }

  function validateField(field: keyof T) {
    if (!validationSchema) return

    touched[field] = true
    const result = validationSchema.safeParse(values)

    if (!result.success) {
      const fieldError = result.error.errors.find(
        err => err.path[0] === field
      )
      errors[field] = fieldError?.message
    } else {
      delete errors[field]
    }
  }

  function setFieldValue<K extends keyof T>(field: K, value: T[K]) {
    values[field] = value
    validateField(field)
  }

  function reset() {
    Object.assign(values, initialValues)
    Object.keys(errors).forEach(key => delete errors[key as keyof T])
    Object.keys(touched).forEach(key => delete touched[key as keyof T])
    submitted.value = false
  }

  async function handleSubmit() {
    if (!validate()) return

    submitting.value = true
    try {
      await onSubmit(values)
      submitted.value = true
    } finally {
      submitting.value = false
    }
  }

  return {
    values,
    errors,
    touched,
    submitting,
    submitted,
    isValid,
    isDirty,
    validate,
    validateField,
    setFieldValue,
    reset,
    handleSubmit
  }
}
```

---

## Reactivity Deep Dive

### Reactive References

```typescript
import {
  ref,
  reactive,
  shallowRef,
  shallowReactive,
  readonly,
  toRef,
  toRefs,
  computed,
  watch,
  watchEffect
} from 'vue'

// ref - for primitives and single values
const count = ref(0)
const user = ref<User | null>(null)

// reactive - for objects
const state = reactive({
  users: [] as User[],
  loading: false,
  filters: {
    search: '',
    role: 'all'
  }
})

// shallowRef/shallowReactive - for performance
// Only tracks top-level reactivity
const largeData = shallowRef<LargeObject[]>([])
const config = shallowReactive({
  settings: {} as Settings // settings changes won't trigger updates
})

// readonly - prevent mutations
const immutableState = readonly(state)

// toRef/toRefs - destructure reactive objects safely
const { search, role } = toRefs(state.filters)
const loadingRef = toRef(state, 'loading')

// Computed with getter and setter
const fullName = computed({
  get: () => `${state.firstName} ${state.lastName}`,
  set: (value: string) => {
    const [first, last] = value.split(' ')
    state.firstName = first
    state.lastName = last || ''
  }
})
```

### Advanced Watchers

```typescript
import { watch, watchEffect, watchPostEffect, watchSyncEffect } from 'vue'

// Watch single source
watch(count, (newVal, oldVal) => {
  console.log(`Count changed from ${oldVal} to ${newVal}`)
})

// Watch multiple sources
watch(
  [count, () => state.loading],
  ([newCount, newLoading], [oldCount, oldLoading]) => {
    console.log('Count or loading changed')
  }
)

// Deep watch with options
watch(
  () => state.filters,
  (newFilters) => {
    fetchData(newFilters)
  },
  {
    deep: true,       // Watch nested properties
    immediate: true,  // Run immediately
    flush: 'post',    // Run after DOM update
    once: true        // Run only once
  }
)

// watchEffect - auto-tracks dependencies
watchEffect(async () => {
  const query = state.filters.search
  if (query) {
    const results = await searchUsers(query)
    state.users = results
  }
})

// watchPostEffect - runs after DOM updates
watchPostEffect(() => {
  // Safe to access DOM here
  document.title = `Users (${state.users.length})`
})

// Cleanup function
watchEffect((onCleanup) => {
  const controller = new AbortController()

  fetch('/api/data', { signal: controller.signal })
    .then(/* ... */)

  onCleanup(() => {
    controller.abort()
  })
})
```

---

## Pinia State Management

### Store Definition

```typescript
// stores/userStore.ts
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

interface User {
  id: string
  name: string
  email: string
  role: 'admin' | 'user'
}

// Composition API style (recommended)
export const useUserStore = defineStore('user', () => {
  // State
  const currentUser = ref<User | null>(null)
  const users = ref<User[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Getters
  const isAuthenticated = computed(() => !!currentUser.value)
  const isAdmin = computed(() => currentUser.value?.role === 'admin')
  const userCount = computed(() => users.value.length)

  const getUserById = computed(() => {
    return (id: string) => users.value.find(u => u.id === id)
  })

  // Actions
  async function login(email: string, password: string) {
    loading.value = true
    error.value = null

    try {
      const response = await fetch('/api/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
        headers: { 'Content-Type': 'application/json' }
      })

      if (!response.ok) {
        throw new Error('Login failed')
      }

      currentUser.value = await response.json()
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Unknown error'
      throw err
    } finally {
      loading.value = false
    }
  }

  function logout() {
    currentUser.value = null
  }

  async function fetchUsers() {
    loading.value = true

    try {
      const response = await fetch('/api/users')
      users.value = await response.json()
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to fetch'
    } finally {
      loading.value = false
    }
  }

  function updateUser(id: string, updates: Partial<User>) {
    const index = users.value.findIndex(u => u.id === id)
    if (index !== -1) {
      users.value[index] = { ...users.value[index], ...updates }
    }
  }

  // Subscribe to actions for side effects
  function $reset() {
    currentUser.value = null
    users.value = []
    loading.value = false
    error.value = null
  }

  return {
    // State
    currentUser,
    users,
    loading,
    error,
    // Getters
    isAuthenticated,
    isAdmin,
    userCount,
    getUserById,
    // Actions
    login,
    logout,
    fetchUsers,
    updateUser,
    $reset
  }
})
```

### Store with Persistence

```typescript
// stores/settingsStore.ts
import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

export const useSettingsStore = defineStore('settings', () => {
  // Initialize from localStorage
  const theme = ref<'light' | 'dark'>(
    (localStorage.getItem('theme') as 'light' | 'dark') || 'light'
  )
  const language = ref(localStorage.getItem('language') || 'en')
  const notifications = ref(
    JSON.parse(localStorage.getItem('notifications') || 'true')
  )

  // Persist on change
  watch(theme, (newTheme) => {
    localStorage.setItem('theme', newTheme)
    document.documentElement.classList.toggle('dark', newTheme === 'dark')
  }, { immediate: true })

  watch(language, (newLang) => {
    localStorage.setItem('language', newLang)
  })

  watch(notifications, (enabled) => {
    localStorage.setItem('notifications', JSON.stringify(enabled))
  })

  function toggleTheme() {
    theme.value = theme.value === 'light' ? 'dark' : 'light'
  }

  function setLanguage(lang: string) {
    language.value = lang
  }

  return {
    theme,
    language,
    notifications,
    toggleTheme,
    setLanguage
  }
})
```

### Using Stores in Components

```vue
<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { useUserStore } from '@/stores/userStore'
import { useSettingsStore } from '@/stores/settingsStore'

const userStore = useUserStore()
const settingsStore = useSettingsStore()

// Destructure with reactivity preserved
const { currentUser, isAuthenticated, loading } = storeToRefs(userStore)
const { theme } = storeToRefs(settingsStore)

// Actions can be destructured directly
const { login, logout, fetchUsers } = userStore
const { toggleTheme } = settingsStore

async function handleLogin(email: string, password: string) {
  try {
    await login(email, password)
    await fetchUsers()
  } catch (error) {
    console.error('Login failed:', error)
  }
}
</script>

<template>
  <div :class="{ dark: theme === 'dark' }">
    <header>
      <button @click="toggleTheme">
        {{ theme === 'dark' ? 'Light' : 'Dark' }} Mode
      </button>

      <template v-if="isAuthenticated">
        <span>{{ currentUser?.name }}</span>
        <button @click="logout">Logout</button>
      </template>
    </header>
  </div>
</template>
```

---

## Provide/Inject Pattern

```typescript
// context/notification.ts
import { provide, inject, ref, readonly } from 'vue'
import type { InjectionKey, Ref } from 'vue'

interface Notification {
  id: string
  type: 'success' | 'error' | 'warning' | 'info'
  message: string
}

interface NotificationContext {
  notifications: Readonly<Ref<Notification[]>>
  add: (notification: Omit<Notification, 'id'>) => void
  remove: (id: string) => void
  clear: () => void
}

const NotificationKey: InjectionKey<NotificationContext> = Symbol('notifications')

export function provideNotifications() {
  const notifications = ref<Notification[]>([])

  function add(notification: Omit<Notification, 'id'>) {
    const id = `${Date.now()}-${Math.random()}`
    notifications.value.push({ ...notification, id })

    // Auto-remove after 5 seconds
    setTimeout(() => remove(id), 5000)
  }

  function remove(id: string) {
    const index = notifications.value.findIndex(n => n.id === id)
    if (index !== -1) {
      notifications.value.splice(index, 1)
    }
  }

  function clear() {
    notifications.value = []
  }

  const context: NotificationContext = {
    notifications: readonly(notifications),
    add,
    remove,
    clear
  }

  provide(NotificationKey, context)

  return context
}

export function useNotifications(): NotificationContext {
  const context = inject(NotificationKey)

  if (!context) {
    throw new Error('useNotifications must be used within NotificationProvider')
  }

  return context
}
```

---

## Best Practices

### Component Organization

Single File Component Structure:
- Script setup at the top with TypeScript
- Template in the middle
- Scoped styles at the bottom

Naming Conventions:
- Components: PascalCase (UserCard.vue)
- Composables: useCamelCase (useAuth.ts)
- Stores: useCamelCaseStore (useUserStore.ts)

### Performance Tips

Use shallowRef for Large Objects:
- When deep reactivity is not needed
- For large arrays of objects

Prefer computed over methods:
- Cached and only re-evaluated when dependencies change

Use v-once for static content:
- Content that never changes after initial render

Lazy load routes and components:
- Use defineAsyncComponent for large components
- Configure route-level code splitting

---

Version: 2.0.0
Last Updated: 2026-01-06
