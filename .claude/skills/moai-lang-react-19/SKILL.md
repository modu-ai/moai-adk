---
title: React 19 Hooks & Modern Patterns
description: Master React 19 features including use() hook, useTransition, Suspense, and Server Components integration
freedom_level: high
tier: language
updated: 2025-10-31
---

# React 19 Hooks & Modern Patterns

## Overview

React 19 introduces significant improvements to hooks, Suspense, transitions, and Server Components integration. The latest version (19.2) adds View Transitions support for SSR and enhanced batching for better performance. This skill covers practical patterns for leveraging React 19's newest capabilities.

## Key Patterns

### 1. use() Hook with Promises

**Pattern**: The `use()` hook can read promises and context, enabling cleaner async data handling.

```typescript
import { use, Suspense } from 'react'

async function fetchUser(id: string) {
  const res = await fetch(`/api/users/${id}`)
  return res.json()
}

function UserProfile({ userPromise }: { userPromise: Promise<User> }) {
  // use() suspends until promise resolves
  const user = use(userPromise)
  
  return (
    <div>
      <h1>{user.name}</h1>
      <p>{user.email}</p>
    </div>
  )
}

export default function UserPage({ params }: { params: { id: string } }) {
  const userPromise = fetchUser(params.id)
  
  return (
    <Suspense fallback={<UserSkeleton />}>
      <UserProfile userPromise={userPromise} />
    </Suspense>
  )
}
```

**Key Difference from useEffect**:
- `use()` can be called conditionally and in loops
- Works seamlessly with Suspense boundaries
- No need for loading state management

### 2. useTransition for Non-Blocking Updates

**Pattern**: Use `useTransition` to keep UI responsive during expensive updates.

```typescript
'use client'

import { useState, useTransition } from 'react'

function SearchResults() {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState([])
  const [isPending, startTransition] = useTransition()
  
  const handleSearch = (value: string) => {
    setQuery(value) // Immediate update (high priority)
    
    startTransition(() => {
      // Non-blocking update (low priority)
      const filtered = expensiveFilter(allItems, value)
      setResults(filtered)
    })
  }
  
  return (
    <div>
      <input 
        value={query}
        onChange={(e) => handleSearch(e.target.value)}
        placeholder="Search..."
      />
      {isPending && <LoadingSpinner />}
      <ResultsList results={results} />
    </div>
  )
}
```

**Benefits**: Input stays responsive, no janky UI during filtering, automatic loading states.

### 3. Async Transitions (React 19 Feature)

**Pattern**: Pass async functions directly to `startTransition` for form submissions.

```typescript
'use client'

import { useTransition } from 'react'

function PostForm() {
  const [isPending, startTransition] = useTransition()
  
  const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    const formData = new FormData(e.currentTarget)
    
    startTransition(async () => {
      const result = await fetch('/api/posts', {
        method: 'POST',
        body: formData
      })
      
      if (result.ok) {
        // React automatically shows pending state
        window.location.href = '/posts'
      }
    })
  }
  
  return (
    <form onSubmit={handleSubmit}>
      <input name="title" required />
      <textarea name="content" required />
      <button disabled={isPending}>
        {isPending ? 'Publishing...' : 'Publish'}
      </button>
    </form>
  )
}
```

**React 19 Enhancement**: Async functions in transitions get automatic error boundaries and pending states.

### 4. Enhanced Suspense Batching

**Pattern**: Multiple suspended components batch together instead of cascading fallbacks.

```typescript
import { Suspense } from 'react'

// React 19.2 batches these suspensions together
export default function Dashboard() {
  return (
    <div>
      <h1>Dashboard</h1>
      <Suspense fallback={<DashboardSkeleton />}>
        {/* These load in parallel and show single fallback */}
        <UserStats />
        <RecentActivity />
        <TeamOverview />
      </Suspense>
    </div>
  )
}

async function UserStats() {
  const stats = await fetch('/api/stats').then(r => r.json())
  return <StatsCard data={stats} />
}

async function RecentActivity() {
  const activity = await fetch('/api/activity').then(r => r.json())
  return <ActivityFeed data={activity} />
}

async function TeamOverview() {
  const team = await fetch('/api/team').then(r => r.json())
  return <TeamList members={team} />
}
```

**React 19.2 Improvement**: Previously, each component would show fallback sequentially (waterfall). Now they batch and reveal together, enabling smoother View Transitions.

### 5. useOptimistic for Instant Feedback

**Pattern**: Show optimistic UI updates before server confirmation.

```typescript
'use client'

import { useOptimistic } from 'react'

type Todo = { id: string; text: string; completed: boolean }

function TodoList({ todos }: { todos: Todo[] }) {
  const [optimisticTodos, addOptimisticTodo] = useOptimistic(
    todos,
    (state, newTodo: Todo) => [...state, newTodo]
  )
  
  const handleAdd = async (text: string) => {
    const tempId = `temp-${Date.now()}`
    
    // Immediately show in UI
    addOptimisticTodo({ id: tempId, text, completed: false })
    
    // Submit to server
    await fetch('/api/todos', {
      method: 'POST',
      body: JSON.stringify({ text })
    })
  }
  
  return (
    <ul>
      {optimisticTodos.map(todo => (
        <li key={todo.id}>{todo.text}</li>
      ))}
    </ul>
  )
}
```

**Use Cases**: Like/favorite buttons, adding comments, creating items, toggling settings.

### 6. useFormStatus for Form State

**Pattern**: Access parent form's submission status without prop drilling.

```typescript
'use client'

import { useFormStatus } from 'react-dom'

function SubmitButton() {
  const { pending } = useFormStatus()
  
  return (
    <button disabled={pending}>
      {pending ? 'Submitting...' : 'Submit'}
    </button>
  )
}

function ContactForm() {
  return (
    <form action={submitContactForm}>
      <input name="email" type="email" required />
      <textarea name="message" required />
      <SubmitButton /> {/* Automatically knows form state */}
    </form>
  )
}
```

**Benefits**: No manual state management, works with Server Actions, progressive enhancement.

### 7. Avoiding Common Suspense Pitfalls

**Pattern**: Structure data fetching to prevent waterfalls.

```typescript
// ❌ BAD: Creates waterfall (sequential loading)
async function BadUserPage() {
  const user = await fetchUser()
  const posts = await fetchPosts(user.id) // Waits for user first
  return <div>{/* render */}</div>
}

// ✅ GOOD: Parallel fetching with separate Suspense boundaries
function GoodUserPage() {
  const userPromise = fetchUser()
  const postsPromise = fetchPosts()
  
  return (
    <>
      <Suspense fallback={<UserSkeleton />}>
        <UserInfo userPromise={userPromise} />
      </Suspense>
      
      <Suspense fallback={<PostsSkeleton />}>
        <PostsList postsPromise={postsPromise} />
      </Suspense>
    </>
  )
}

function UserInfo({ userPromise }: { userPromise: Promise<User> }) {
  const user = use(userPromise)
  return <div>{user.name}</div>
}
```

## Checklist

- [ ] Replace `useEffect` data fetching with `use()` hook + Suspense
- [ ] Add `useTransition` to expensive filtering/sorting operations
- [ ] Implement optimistic UI for user actions (like, add, delete)
- [ ] Use `useFormStatus` for form submission feedback
- [ ] Audit Suspense boundaries: ensure parallel loading, not waterfalls
- [ ] Test async transitions with network throttling (DevTools)
- [ ] Verify View Transitions work in React 19.2+ SSR apps

## Resources

- **React 19 Official Release**: https://react.dev/blog/2024/12/05/react-19
- **React 19.2 Updates**: https://react.dev/blog/2025/10/01/react-19-2
- **useTransition Documentation**: https://react.dev/reference/react/useTransition
- **Suspense Documentation**: https://react.dev/reference/react/Suspense
- **use() Hook Guide**: https://mittalkartik1.medium.com/susexploring-react-19-the-new-use-api-with-suspense-4be658cf7ee2
- **React 19 New Hooks**: https://marmelab.com/blog/2024/01/23/react-19-new-hooks.html

---

**Version**: 1.0.0  
**Last Updated**: 2025-10-31  
**Model Recommendation**: Sonnet (deep reasoning for hook patterns and Suspense boundaries)
