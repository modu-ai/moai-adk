---
name: "moai-platform-convex"
description: "Convex real-time backend specialist covering TypeScript-first reactive patterns, optimistic updates, server functions, and file storage. Use when building real-time collaborative apps, implementing reactive queries, or integrating with Clerk/Auth0."
version: 2.0.0
category: "platform"
modularized: true
user-invocable: false
tags: ['convex', 'realtime', 'reactive', 'typescript', 'optimistic-updates']
context7-libraries: "/get-convex/convex"
related-skills: "moai-platform-supabase, moai-lang-typescript, moai-domain-frontend"
updated: 2026-01-08
status: "active"
allowed-tools:
  - Read
  - Write
  - Bash
  - Grep
  - Glob
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
---

# Convex Real-time Backend Specialist

Convex is a real-time reactive backend platform with TypeScript-first design, automatic caching, and optimistic updates.

---

## Quick Reference

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

### Context7 Library

Use mcp__context7__resolve-library-id with "convex" to get the library ID, then use mcp__context7__get-library-docs for latest documentation.

---

## Module Index

This skill is organized into specialized modules for detailed implementation guidance:

### Reactive Queries Module
Location: modules/reactive-queries.md

Covers real-time reactive query patterns including basic query structure, index-based queries, search indexes, pagination patterns, React integration with useQuery hooks, and optimistic updates.

### Server Functions Module
Location: modules/server-functions.md

Covers mutations for write operations, actions for external API integration, internal functions for server-to-server calls, scheduled functions (crons), and HTTP endpoints for webhooks.

### Authentication Integration Module
Location: modules/auth-integration.md

Covers Clerk integration, Auth0 integration, server-side authentication patterns, authorization and role-based access control, and session management.

### File Storage Module
Location: modules/file-storage.md

Covers file upload workflows, storage URL generation, client-side upload with progress, file display components, and file management operations.

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
    })
})
```

### Validators (v module)

```typescript
import { v } from 'convex/values'

// Primitives
v.string(), v.number(), v.boolean(), v.null(), v.int64(), v.bytes()

// Complex types
v.array(v.string())
v.object({ name: v.string(), age: v.number() })
v.union(v.literal('read'), v.literal('write'))
v.optional(v.string())

// References
v.id('tableName')
```

### React Client Setup

```typescript
// src/providers/ConvexProvider.tsx
import { ConvexProvider, ConvexReactClient } from 'convex/react'
import { ConvexProviderWithClerk } from 'convex/react-clerk'
import { useAuth } from '@clerk/clerk-react'

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

## Best Practices

Query Optimization:
- Use indexes for all filtered queries
- Prefer paginated queries for large datasets
- Use search indexes for full-text search
- Leverage automatic caching

Mutation Design:
- Keep mutations focused and atomic
- Use internal mutations for multi-step operations
- Validate all inputs with the v module
- Always check authorization

Error Handling:
- Use ConvexError for structured errors
- Check for undefined during loading states
- Handle optimistic update rollbacks

---

## Works Well With

- moai-platform-supabase - Alternative PostgreSQL-based backend
- moai-lang-typescript - TypeScript patterns and best practices
- moai-domain-frontend - React integration patterns
- moai-platform-clerk - Clerk authentication patterns
- moai-platform-auth0 - Auth0 authentication patterns

---

## Resources

- Official Documentation: https://docs.convex.dev
- Context7 Library: /get-convex/convex
- GitHub: https://github.com/get-convex/convex

---

Status: Production Ready
Version: 2.0.0
Last Updated: 2026-01-06
Platform: Convex Real-time Backend
