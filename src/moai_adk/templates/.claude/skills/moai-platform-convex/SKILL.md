---
name: moai-platform-convex
description: Convex real-time backend specialist covering TypeScript-first reactive patterns, optimistic updates, and server functions. Use when building real-time collaborative apps.
version: 1.0.0
category: platform
tags: convex, realtime, reactive, typescript, optimistic-updates
context7-libraries: /get-convex/convex
related-skills: moai-platform-supabase, moai-lang-typescript
allowed-tools: Read, Write, Bash, Grep, Glob
---

# moai-platform-convex: Convex Real-time Backend Specialist

## Quick Reference (30 seconds)

Convex is a real-time reactive backend platform with TypeScript-first design, automatic caching, and optimistic updates.

### When to Use Convex

- Real-time collaborative applications (docs, whiteboards, chat)
- Apps requiring instant UI updates without manual refetching
- TypeScript-first projects needing end-to-end type safety
- Applications with complex optimistic update requirements

### Core Concepts

Server Functions: queries (read), mutations (write), actions (external APIs)
Reactive Queries: Automatic re-execution when underlying data changes
Optimistic Updates: Instant UI updates before server confirmation
Automatic Caching: Built-in query result caching with intelligent invalidation

### Quick Start

```bash
npm create convex@latest
npx convex dev
```

### Context7 Library: /get-convex/convex

---

## Implementation Guide

### Project Structure

```
my-app/
  convex/
    _generated/         # Auto-generated types and API
    schema.ts           # Database schema definition
    functions/          # Server functions by domain
    http.ts             # HTTP endpoints (optional)
    crons.ts            # Scheduled jobs (optional)
  src/
    ConvexProvider.tsx  # Client setup
```

### Schema Definition

```typescript
// convex/schema.ts
import { defineSchema, defineTable } from 'convex/server'
import { v } from 'convex/values'

export default defineSchema({
  documents: defineTable({
    title: v.string(),
    content: v.string(),
    ownerId: v.string(),
    isPublic: v.boolean(),
    createdAt: v.number(),
    updatedAt: v.number()
  })
    .index('by_owner', ['ownerId'])
    .index('by_public', ['isPublic', 'createdAt'])
    .searchIndex('search_content', {
      searchField: 'content',
      filterFields: ['ownerId', 'isPublic']
    }),

  collaborators: defineTable({
    documentId: v.id('documents'),
    userId: v.string(),
    permission: v.union(v.literal('read'), v.literal('write'))
  })
    .index('by_document', ['documentId'])
    .index('by_user', ['userId'])
})
```

### Validators (v module)

```typescript
import { v } from 'convex/values'

// Primitives: v.string(), v.number(), v.boolean(), v.null(), v.int64(), v.bytes()
// Complex: v.array(v.string()), v.object({...}), v.union(...), v.optional(...)
// References: v.id('tableName')
```

### Query Functions (Reactive)

```typescript
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

export const getById = query({
  args: { id: v.id('documents') },
  handler: async (ctx, args) => {
    const doc = await ctx.db.get(args.id)
    if (!doc) throw new Error('Document not found')
    return doc
  }
})

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

### Mutation Functions

```typescript
import { mutation } from '../_generated/server'
import { v } from 'convex/values'

export const create = mutation({
  args: { title: v.string(), content: v.string(), isPublic: v.boolean() },
  handler: async (ctx, args) => {
    const identity = await ctx.auth.getUserIdentity()
    if (!identity) throw new Error('Unauthorized')
    return await ctx.db.insert('documents', {
      ...args,
      ownerId: identity.subject,
      createdAt: Date.now(),
      updatedAt: Date.now()
    })
  }
})

export const update = mutation({
  args: { id: v.id('documents'), title: v.optional(v.string()), content: v.optional(v.string()) },
  handler: async (ctx, args) => {
    const { id, ...updates } = args
    const existing = await ctx.db.get(id)
    if (!existing) throw new Error('Document not found')
    const identity = await ctx.auth.getUserIdentity()
    if (existing.ownerId !== identity?.subject) throw new Error('Forbidden')
    await ctx.db.patch(id, { ...updates, updatedAt: Date.now() })
  }
})

export const remove = mutation({
  args: { id: v.id('documents') },
  handler: async (ctx, args) => await ctx.db.delete(args.id)
})
```

### Action Functions (External APIs)

```typescript
import { action } from '../_generated/server'
import { internal } from '../_generated/api'
import { v } from 'convex/values'

export const generateSummary = action({
  args: { documentId: v.id('documents') },
  handler: async (ctx, args) => {
    const doc = await ctx.runQuery(internal.documents.getById, { id: args.documentId })
    const response = await fetch('https://api.openai.com/v1/completions', {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${process.env.OPENAI_API_KEY}`, 'Content-Type': 'application/json' },
      body: JSON.stringify({ model: 'gpt-4', prompt: `Summarize: ${doc.content}`, max_tokens: 150 })
    })
    const result = await response.json()
    await ctx.runMutation(internal.documents.updateSummary, { id: args.documentId, summary: result.choices[0].text })
    return result.choices[0].text
  }
})
```

### React Client Setup

```typescript
import { ConvexProvider, ConvexReactClient } from 'convex/react'
import { ConvexProviderWithClerk } from 'convex/react-clerk'

const convex = new ConvexReactClient(import.meta.env.VITE_CONVEX_URL)

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <ConvexProviderWithClerk client={convex} useAuth={useAuth}>
      {children}
    </ConvexProviderWithClerk>
  )
}
```

### React Hooks Usage

```typescript
import { useQuery, useMutation } from 'convex/react'
import { api } from '../../convex/_generated/api'

export function DocumentList({ userId }: { userId: string }) {
  const documents = useQuery(api.functions.documents.list, { ownerId: userId })
  const createDocument = useMutation(api.functions.documents.create)

  if (documents === undefined) return <Loading />

  return (
    <div>
      <button onClick={() => createDocument({ title: 'New', content: '', isPublic: false })}>
        New Document
      </button>
      {documents.map((doc) => <DocumentCard key={doc._id} document={doc} />)}
    </div>
  )
}
```

---

## Advanced Patterns

### Optimistic Updates

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
          ...existing, ...updates, updatedAt: Date.now()
        })
      }
    })
}
```

### File Storage

```typescript
// Server-side
export const generateUploadUrl = mutation({
  handler: async (ctx) => await ctx.storage.generateUploadUrl()
})

export const saveFile = mutation({
  args: { storageId: v.id('_storage'), fileName: v.string() },
  handler: async (ctx, args) => await ctx.db.insert('files', { ...args, uploadedAt: Date.now() })
})

export const getFileUrl = query({
  args: { storageId: v.id('_storage') },
  handler: async (ctx, args) => await ctx.storage.getUrl(args.storageId)
})
```

```typescript
// Client-side upload
export function useFileUpload() {
  const generateUploadUrl = useMutation(api.functions.files.generateUploadUrl)
  const saveFile = useMutation(api.functions.files.saveFile)

  return async (file: File) => {
    const uploadUrl = await generateUploadUrl()
    const response = await fetch(uploadUrl, { method: 'POST', headers: { 'Content-Type': file.type }, body: file })
    const { storageId } = await response.json()
    await saveFile({ storageId, fileName: file.name })
    return storageId
  }
}
```

### Scheduled Functions (Crons)

```typescript
import { cronJobs } from 'convex/server'
import { internal } from './_generated/api'

const crons = cronJobs()
crons.interval('cleanup old drafts', { hours: 24 }, internal.documents.cleanupOldDrafts)
crons.cron('daily analytics', '0 0 * * *', internal.analytics.generateDailyReport)
export default crons
```

### HTTP Endpoints

```typescript
import { httpRouter } from 'convex/server'
import { httpAction } from './_generated/server'

const http = httpRouter()
http.route({
  path: '/webhook/stripe',
  method: 'POST',
  handler: httpAction(async (ctx, request) => {
    const body = await request.text()
    await ctx.runMutation(internal.payments.processWebhook, { body, signature: request.headers.get('stripe-signature') })
    return new Response('OK', { status: 200 })
  })
})
export default http
```

### Authentication (Clerk)

```typescript
export const current = query({
  handler: async (ctx) => {
    const identity = await ctx.auth.getUserIdentity()
    if (!identity) return null
    return await ctx.db.query('users').withIndex('by_token', (q) => q.eq('tokenIdentifier', identity.tokenIdentifier)).first()
  }
})

export const ensureUser = mutation({
  handler: async (ctx) => {
    const identity = await ctx.auth.getUserIdentity()
    if (!identity) throw new Error('Unauthorized')
    const existing = await ctx.db.query('users').withIndex('by_token', (q) => q.eq('tokenIdentifier', identity.tokenIdentifier)).first()
    if (existing) return existing._id
    return await ctx.db.insert('users', { tokenIdentifier: identity.tokenIdentifier, email: identity.email, name: identity.name, createdAt: Date.now() })
  }
})
```

### Error Handling

```typescript
import { ConvexError } from 'convex/values'

export const secureOperation = mutation({
  args: { id: v.id('documents') },
  handler: async (ctx, args) => {
    const identity = await ctx.auth.getUserIdentity()
    if (!identity) throw new ConvexError('UNAUTHORIZED')
    const doc = await ctx.db.get(args.id)
    if (!doc) throw new ConvexError({ code: 'NOT_FOUND', message: 'Document not found' })
  }
})
```

---

## Best Practices

Query Optimization:
- Use indexes for all filtered queries
- Prefer paginated queries for large datasets
- Use search indexes for full-text search

Mutation Design:
- Keep mutations focused and atomic
- Use internal mutations for multi-step operations
- Validate all inputs with the v module

---

## Works Well With

- moai-platform-supabase - Alternative PostgreSQL-based backend
- moai-lang-typescript - TypeScript patterns and best practices
- moai-domain-frontend - React integration patterns
- moai-quality-security - Authentication and authorization patterns

---

Status: Production Ready
Generated with: MoAI-ADK Skill Factory v2.0
Last Updated: 2025-12-07
Platform: Convex Real-time Backend
