# React 19 Patterns

Comprehensive patterns for React 19 development covering Server Components, Concurrent features, Suspense, and the cache() API.

---

## Server Components Architecture

### Understanding Server Components

Server Components execute exclusively on the server, enabling direct database access and keeping sensitive logic server-side. They reduce client bundle size and improve initial load performance.

Key Characteristics:
- Zero client-side JavaScript overhead
- Direct access to backend resources
- Automatic code splitting
- Cannot use hooks or browser APIs

### Server Component Pattern

```tsx
// app/components/UserProfile.tsx
import { cache } from 'react'
import { getUser, getUserPosts } from '@/lib/api'

// Cache data fetching for request deduplication
const getUserCached = cache(async (userId: string) => {
  return await getUser(userId)
})

const getUserPostsCached = cache(async (userId: string) => {
  return await getUserPosts(userId, { limit: 10 })
})

interface UserProfileProps {
  userId: string
}

export default async function UserProfile({ userId }: UserProfileProps) {
  // Parallel data fetching
  const [user, posts] = await Promise.all([
    getUserCached(userId),
    getUserPostsCached(userId)
  ])

  return (
    <article className="user-profile">
      <header className="profile-header">
        <img src={user.avatar} alt={user.name} className="avatar" />
        <div className="info">
          <h1>{user.name}</h1>
          <p className="bio">{user.bio}</p>
          <p className="stats">
            {user.followersCount} followers | {posts.length} posts
          </p>
        </div>
      </header>

      <section className="posts">
        <h2>Recent Posts</h2>
        {posts.map(post => (
          <PostCard key={post.id} post={post} />
        ))}
      </section>

      {/* Client Component for interactivity */}
      <ProfileActions userId={userId} initialFollowing={user.isFollowing} />
    </article>
  )
}
```

### Client Components Integration

```tsx
// components/ProfileActions.tsx
'use client'

import { useState, useTransition, useOptimistic } from 'react'
import { followUser, unfollowUser } from '@/app/actions/users'
import { Button } from '@/components/ui/button'

interface ProfileActionsProps {
  userId: string
  initialFollowing: boolean
}

export function ProfileActions({ userId, initialFollowing }: ProfileActionsProps) {
  const [isPending, startTransition] = useTransition()
  const [optimisticFollowing, setOptimisticFollowing] = useOptimistic(
    initialFollowing,
    (current, newValue: boolean) => newValue
  )

  const handleToggleFollow = () => {
    const newValue = !optimisticFollowing
    setOptimisticFollowing(newValue)

    startTransition(async () => {
      if (newValue) {
        await followUser(userId)
      } else {
        await unfollowUser(userId)
      }
    })
  }

  return (
    <div className="profile-actions">
      <Button
        onClick={handleToggleFollow}
        variant={optimisticFollowing ? 'outline' : 'default'}
        disabled={isPending}
      >
        {optimisticFollowing ? 'Following' : 'Follow'}
      </Button>
    </div>
  )
}
```

---

## Concurrent Features

### Suspense for Data Fetching

```tsx
import { Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'

function DashboardPage() {
  return (
    <div className="dashboard">
      <h1>Dashboard</h1>

      {/* Independent Suspense boundaries for parallel loading */}
      <div className="grid grid-cols-3 gap-4">
        <ErrorBoundary fallback={<WidgetError />}>
          <Suspense fallback={<WidgetSkeleton />}>
            <AnalyticsWidget />
          </Suspense>
        </ErrorBoundary>

        <ErrorBoundary fallback={<WidgetError />}>
          <Suspense fallback={<WidgetSkeleton />}>
            <RevenueWidget />
          </Suspense>
        </ErrorBoundary>

        <ErrorBoundary fallback={<WidgetError />}>
          <Suspense fallback={<WidgetSkeleton />}>
            <UsersWidget />
          </Suspense>
        </ErrorBoundary>
      </div>
    </div>
  )
}

// Each widget fetches data independently
async function AnalyticsWidget() {
  const data = await getAnalytics()
  return (
    <div className="widget">
      <h3>Analytics</h3>
      <p>Page views: {data.pageViews}</p>
      <p>Unique visitors: {data.uniqueVisitors}</p>
    </div>
  )
}
```

### useTransition for Non-Blocking Updates

```tsx
'use client'

import { useState, useTransition, useDeferredValue } from 'react'

interface SearchableListProps {
  items: Item[]
}

export function SearchableList({ items }: SearchableListProps) {
  const [query, setQuery] = useState('')
  const [isPending, startTransition] = useTransition()

  // Defer the filtered results for smooth typing
  const deferredQuery = useDeferredValue(query)

  const filteredItems = items.filter(item =>
    item.name.toLowerCase().includes(deferredQuery.toLowerCase())
  )

  const handleSearch = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value
    // Immediate update for input
    setQuery(value)

    // Or use startTransition for expensive operations
    startTransition(() => {
      // Expensive filtering or state updates
    })
  }

  return (
    <div>
      <input
        type="search"
        value={query}
        onChange={handleSearch}
        placeholder="Search..."
        className="search-input"
      />

      {/* Show loading indicator without blocking input */}
      <div className={isPending || query !== deferredQuery ? 'opacity-70' : ''}>
        {filteredItems.map(item => (
          <ItemCard key={item.id} item={item} />
        ))}
      </div>
    </div>
  )
}
```

### useOptimistic for Instant Feedback

```tsx
'use client'

import { useOptimistic, useTransition } from 'react'
import { addComment, deleteComment } from '@/app/actions/comments'

interface Comment {
  id: string
  text: string
  author: string
  pending?: boolean
}

interface CommentListProps {
  postId: string
  initialComments: Comment[]
}

export function CommentList({ postId, initialComments }: CommentListProps) {
  const [isPending, startTransition] = useTransition()
  const [optimisticComments, updateOptimistic] = useOptimistic(
    initialComments,
    (state, action: { type: 'add' | 'delete'; comment?: Comment; id?: string }) => {
      switch (action.type) {
        case 'add':
          return [...state, { ...action.comment!, pending: true }]
        case 'delete':
          return state.filter(c => c.id !== action.id)
        default:
          return state
      }
    }
  )

  const handleAddComment = async (formData: FormData) => {
    const text = formData.get('text') as string
    const tempId = `temp-${Date.now()}`

    const newComment: Comment = {
      id: tempId,
      text,
      author: 'Current User',
      pending: true
    }

    updateOptimistic({ type: 'add', comment: newComment })

    startTransition(async () => {
      await addComment(postId, text)
    })
  }

  const handleDelete = (commentId: string) => {
    updateOptimistic({ type: 'delete', id: commentId })

    startTransition(async () => {
      await deleteComment(commentId)
    })
  }

  return (
    <div className="comments">
      <form action={handleAddComment} className="comment-form">
        <textarea name="text" placeholder="Add a comment..." required />
        <button type="submit" disabled={isPending}>
          {isPending ? 'Posting...' : 'Post Comment'}
        </button>
      </form>

      <ul className="comment-list">
        {optimisticComments.map(comment => (
          <li
            key={comment.id}
            className={comment.pending ? 'opacity-60' : ''}
          >
            <p>{comment.text}</p>
            <span className="author">{comment.author}</span>
            {!comment.pending && (
              <button onClick={() => handleDelete(comment.id)}>
                Delete
              </button>
            )}
          </li>
        ))}
      </ul>
    </div>
  )
}
```

---

## Cache API Patterns

### Request Deduplication

```tsx
import { cache } from 'react'

// Define cached functions at module level
const getUser = cache(async (id: string) => {
  console.log(`Fetching user ${id}`)
  const response = await fetch(`/api/users/${id}`)
  return response.json()
})

const getTeam = cache(async (teamId: string) => {
  console.log(`Fetching team ${teamId}`)
  const response = await fetch(`/api/teams/${teamId}`)
  return response.json()
})

// Multiple components can call the same cached function
// Only one request is made per request lifecycle

async function UserHeader({ userId }: { userId: string }) {
  const user = await getUser(userId)
  return <h1>{user.name}</h1>
}

async function UserStats({ userId }: { userId: string }) {
  const user = await getUser(userId) // Reuses cached result
  return <p>Member since: {user.createdAt}</p>
}

async function UserPage({ userId }: { userId: string }) {
  return (
    <div>
      <UserHeader userId={userId} />
      <UserStats userId={userId} />
    </div>
  )
}
```

### Preloading Data

```tsx
import { cache } from 'react'

const getProducts = cache(async (category: string) => {
  const response = await fetch(`/api/products?category=${category}`)
  return response.json()
})

// Preload function to start fetching early
function preloadProducts(category: string) {
  void getProducts(category)
}

// In a parent component or layout
function CategoryNav({ categories }: { categories: string[] }) {
  return (
    <nav>
      {categories.map(category => (
        <Link
          key={category}
          href={`/products/${category}`}
          onMouseEnter={() => preloadProducts(category)}
        >
          {category}
        </Link>
      ))}
    </nav>
  )
}
```

---

## Form Handling with Actions

### useActionState Pattern

```tsx
'use client'

import { useActionState } from 'react'
import { submitForm, type FormState } from '@/app/actions/form'

const initialState: FormState = {
  success: false,
  errors: {}
}

export function ContactForm() {
  const [state, formAction, isPending] = useActionState(
    submitForm,
    initialState
  )

  return (
    <form action={formAction} className="space-y-4">
      {state.errors?.general && (
        <div className="alert alert-error">{state.errors.general}</div>
      )}

      {state.success && (
        <div className="alert alert-success">Message sent successfully!</div>
      )}

      <div>
        <label htmlFor="name">Name</label>
        <input
          id="name"
          name="name"
          type="text"
          required
          aria-describedby={state.errors?.name ? 'name-error' : undefined}
        />
        {state.errors?.name && (
          <p id="name-error" className="error">{state.errors.name}</p>
        )}
      </div>

      <div>
        <label htmlFor="email">Email</label>
        <input
          id="email"
          name="email"
          type="email"
          required
          aria-describedby={state.errors?.email ? 'email-error' : undefined}
        />
        {state.errors?.email && (
          <p id="email-error" className="error">{state.errors.email}</p>
        )}
      </div>

      <div>
        <label htmlFor="message">Message</label>
        <textarea
          id="message"
          name="message"
          required
          rows={4}
          aria-describedby={state.errors?.message ? 'message-error' : undefined}
        />
        {state.errors?.message && (
          <p id="message-error" className="error">{state.errors.message}</p>
        )}
      </div>

      <button type="submit" disabled={isPending}>
        {isPending ? 'Sending...' : 'Send Message'}
      </button>
    </form>
  )
}
```

### useFormStatus for Submit Buttons

```tsx
'use client'

import { useFormStatus } from 'react-dom'

export function SubmitButton({ children }: { children: React.ReactNode }) {
  const { pending, data, method, action } = useFormStatus()

  return (
    <button
      type="submit"
      disabled={pending}
      className="btn btn-primary"
    >
      {pending ? (
        <span className="flex items-center gap-2">
          <Spinner className="w-4 h-4" />
          Submitting...
        </span>
      ) : (
        children
      )}
    </button>
  )
}

// Usage
function MyForm() {
  return (
    <form action={submitAction}>
      <input name="email" type="email" />
      <SubmitButton>Subscribe</SubmitButton>
    </form>
  )
}
```

---

## Best Practices

### Component Boundary Guidelines

When to Use Server Components:
- Data fetching and backend access
- Large dependencies that should not be sent to client
- Security-sensitive operations
- Static or rarely changing content

When to Use Client Components:
- Event handlers and user interactions
- Browser APIs (localStorage, geolocation)
- State management with hooks
- Real-time updates and WebSocket connections

### Performance Optimization

```tsx
// Streaming with nested Suspense
function ProductPage({ productId }: { productId: string }) {
  return (
    <div className="product-page">
      {/* Critical content loads first */}
      <Suspense fallback={<ProductHeaderSkeleton />}>
        <ProductHeader productId={productId} />
      </Suspense>

      {/* Reviews can stream in later */}
      <Suspense fallback={<ReviewsSkeleton />}>
        <ProductReviews productId={productId} />
      </Suspense>

      {/* Recommendations have lowest priority */}
      <Suspense fallback={<RecommendationsSkeleton />}>
        <RelatedProducts productId={productId} />
      </Suspense>
    </div>
  )
}
```

---

## Common Patterns Summary

Server Data Fetching:
- Use cache() for request deduplication
- Parallel fetching with Promise.all()
- Preload data on hover for faster navigation

Client Interactivity:
- useTransition for non-blocking updates
- useOptimistic for instant feedback
- useDeferredValue for expensive computations

Form Handling:
- useActionState for form state management
- useFormStatus for submit button states
- Server Actions for form processing

Error Boundaries:
- Wrap Suspense with ErrorBoundary
- Provide meaningful fallback UI
- Log errors for monitoring

---

Version: 2.0.0
Last Updated: 2026-01-06
