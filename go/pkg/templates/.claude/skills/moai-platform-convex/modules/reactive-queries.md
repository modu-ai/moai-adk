# Reactive Queries Module

Real-time reactive query patterns for Convex applications with automatic re-execution, intelligent caching, and optimistic updates.

---

## Query Fundamentals

### Basic Query Structure

Queries are reactive functions that automatically re-execute when underlying data changes.

```typescript
// convex/functions/documents.ts
import { query } from '../_generated/server'
import { v } from 'convex/values'

export const list = query({
  args: { ownerId: v.string() },
  handler: async (ctx, args) => {
    return await ctx.db
      .query('documents')
      .withIndex('by_owner', (q) => q.eq('ownerId', args.ownerId))
      .order('desc')
      .collect()
  }
})
```

### Query with Single Document

```typescript
export const getById = query({
  args: { id: v.id('documents') },
  handler: async (ctx, args) => {
    const doc = await ctx.db.get(args.id)
    if (!doc) throw new Error('Document not found')
    return doc
  }
})
```

---

## Index-Based Queries

### Single Field Index

```typescript
// Schema definition
documents: defineTable({
  title: v.string(),
  ownerId: v.string(),
  createdAt: v.number()
}).index('by_owner', ['ownerId'])

// Query using index
export const byOwner = query({
  args: { ownerId: v.string() },
  handler: async (ctx, args) => {
    return await ctx.db
      .query('documents')
      .withIndex('by_owner', (q) => q.eq('ownerId', args.ownerId))
      .collect()
  }
})
```

### Compound Index

```typescript
// Schema with compound index
documents: defineTable({
  ownerId: v.string(),
  status: v.string(),
  createdAt: v.number()
}).index('by_owner_status', ['ownerId', 'status'])

// Query with compound conditions
export const byOwnerAndStatus = query({
  args: { ownerId: v.string(), status: v.string() },
  handler: async (ctx, args) => {
    return await ctx.db
      .query('documents')
      .withIndex('by_owner_status', (q) =>
        q.eq('ownerId', args.ownerId).eq('status', args.status)
      )
      .collect()
  }
})
```

### Range Queries

```typescript
export const recentByOwner = query({
  args: { ownerId: v.string(), since: v.number() },
  handler: async (ctx, args) => {
    return await ctx.db
      .query('documents')
      .withIndex('by_owner_created', (q) =>
        q.eq('ownerId', args.ownerId).gte('createdAt', args.since)
      )
      .order('desc')
      .collect()
  }
})
```

---

## Search Indexes

### Full-Text Search

```typescript
// Schema with search index
documents: defineTable({
  title: v.string(),
  content: v.string(),
  ownerId: v.string(),
  isPublic: v.boolean()
}).searchIndex('search_content', {
  searchField: 'content',
  filterFields: ['ownerId', 'isPublic']
})

// Search query
export const searchContent = query({
  args: { searchQuery: v.string(), limit: v.optional(v.number()) },
  handler: async (ctx, args) => {
    return await ctx.db
      .query('documents')
      .withSearchIndex('search_content', (q) =>
        q.search('content', args.searchQuery).eq('isPublic', true)
      )
      .take(args.limit ?? 10)
  }
})
```

---

## Pagination Patterns

### Cursor-Based Pagination

```typescript
export const paginated = query({
  args: {
    paginationOpts: v.object({
      numItems: v.number(),
      cursor: v.union(v.string(), v.null())
    })
  },
  handler: async (ctx, args) => {
    return await ctx.db
      .query('documents')
      .order('desc')
      .paginate(args.paginationOpts)
  }
})
```

### React Hook for Pagination

```typescript
import { usePaginatedQuery } from 'convex/react'
import { api } from '../../convex/_generated/api'

export function PaginatedDocumentList() {
  const { results, status, loadMore } = usePaginatedQuery(
    api.functions.documents.paginated,
    {},
    { initialNumItems: 10 }
  )

  return (
    <div>
      {results.map((doc) => <DocumentCard key={doc._id} document={doc} />)}
      {status === 'CanLoadMore' && (
        <button onClick={() => loadMore(10)}>Load More</button>
      )}
      {status === 'LoadingMore' && <Loading />}
    </div>
  )
}
```

---

## React Integration

### Basic useQuery Hook

```typescript
import { useQuery } from 'convex/react'
import { api } from '../../convex/_generated/api'

export function DocumentList({ userId }: { userId: string }) {
  const documents = useQuery(api.functions.documents.list, { ownerId: userId })

  if (documents === undefined) return <Loading />

  return (
    <div>
      {documents.map((doc) => <DocumentCard key={doc._id} document={doc} />)}
    </div>
  )
}
```

### Conditional Queries

```typescript
export function ConditionalDocument({ documentId }: { documentId: string | null }) {
  const document = useQuery(
    api.functions.documents.getById,
    documentId ? { id: documentId } : 'skip'
  )

  if (documentId === null) return <SelectDocument />
  if (document === undefined) return <Loading />

  return <DocumentView document={document} />
}
```

---

## Optimistic Updates

### Basic Optimistic Update

```typescript
import { useMutation } from 'convex/react'
import { api } from '../../convex/_generated/api'

export function useOptimisticUpdate() {
  return useMutation(api.functions.documents.update)
    .withOptimisticUpdate((localStore, args) => {
      const { id, ...updates } = args
      const existing = localStore.getQuery(api.functions.documents.getById, { id })
      if (existing) {
        localStore.setQuery(api.functions.documents.getById, { id }, {
          ...existing,
          ...updates,
          updatedAt: Date.now()
        })
      }
    })
}
```

### Optimistic Create

```typescript
export function useOptimisticCreate() {
  return useMutation(api.functions.documents.create)
    .withOptimisticUpdate((localStore, args) => {
      const currentList = localStore.getQuery(
        api.functions.documents.list,
        { ownerId: args.ownerId }
      )
      if (currentList) {
        const optimisticDoc = {
          _id: `optimistic_${Date.now()}` as any,
          _creationTime: Date.now(),
          ...args,
          createdAt: Date.now(),
          updatedAt: Date.now()
        }
        localStore.setQuery(
          api.functions.documents.list,
          { ownerId: args.ownerId },
          [optimisticDoc, ...currentList]
        )
      }
    })
}
```

---

## Query Best Practices

Performance Optimization:
- Always use indexes for filtered queries
- Prefer specific indexes over table scans
- Use pagination for large result sets
- Limit search results with take()

Reactivity Considerations:
- Queries automatically re-run when data changes
- Use skip parameter to conditionally disable queries
- Combine multiple queries in single components when related

---

Version: 1.0.0
Module: Reactive Queries
Parent Skill: moai-platform-convex
