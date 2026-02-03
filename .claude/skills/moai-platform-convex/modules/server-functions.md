# Server Functions Module

Mutations, actions, internal functions, and scheduled jobs for Convex server-side operations.

---

## Mutation Functions

### Basic Mutation Structure

Mutations are transactional write operations that modify the database.

```typescript
// convex/functions/documents.ts
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
```

### Update Mutation

```typescript
export const update = mutation({
  args: {
    id: v.id('documents'),
    title: v.optional(v.string()),
    content: v.optional(v.string())
  },
  handler: async (ctx, args) => {
    const { id, ...updates } = args
    const existing = await ctx.db.get(id)
    if (!existing) throw new Error('Document not found')
    
    const identity = await ctx.auth.getUserIdentity()
    if (existing.ownerId !== identity?.subject) {
      throw new Error('Forbidden')
    }
    
    await ctx.db.patch(id, { ...updates, updatedAt: Date.now() })
  }
})
```

### Delete Mutation

```typescript
export const remove = mutation({
  args: { id: v.id('documents') },
  handler: async (ctx, args) => {
    const doc = await ctx.db.get(args.id)
    if (!doc) throw new Error('Document not found')
    
    const identity = await ctx.auth.getUserIdentity()
    if (doc.ownerId !== identity?.subject) {
      throw new Error('Forbidden')
    }
    
    await ctx.db.delete(args.id)
  }
})
```

### Batch Operations

```typescript
export const batchUpdate = mutation({
  args: {
    ids: v.array(v.id('documents')),
    updates: v.object({ status: v.string() })
  },
  handler: async (ctx, args) => {
    const results = await Promise.all(
      args.ids.map(async (id) => {
        await ctx.db.patch(id, { ...args.updates, updatedAt: Date.now() })
        return id
      })
    )
    return results
  }
})
```

---

## Action Functions

### External API Integration

Actions can call external APIs and perform side effects.

```typescript
// convex/functions/ai.ts
import { action } from '../_generated/server'
import { internal } from '../_generated/api'
import { v } from 'convex/values'

export const generateSummary = action({
  args: { documentId: v.id('documents') },
  handler: async (ctx, args) => {
    // Fetch document using internal query
    const doc = await ctx.runQuery(internal.documents.getById, {
      id: args.documentId
    })
    
    // Call external API
    const response = await fetch('https://api.openai.com/v1/chat/completions', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${process.env.OPENAI_API_KEY}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        model: 'gpt-4',
        messages: [
          { role: 'system', content: 'Summarize the following text.' },
          { role: 'user', content: doc.content }
        ]
      })
    })
    
    const result = await response.json()
    const summary = result.choices[0].message.content
    
    // Update document with summary
    await ctx.runMutation(internal.documents.updateSummary, {
      id: args.documentId,
      summary
    })
    
    return summary
  }
})
```

### Webhook Handler Action

```typescript
export const processWebhook = action({
  args: { body: v.string(), signature: v.string() },
  handler: async (ctx, args) => {
    // Verify webhook signature
    const isValid = verifySignature(args.body, args.signature)
    if (!isValid) throw new Error('Invalid signature')
    
    const payload = JSON.parse(args.body)
    
    // Process webhook based on event type
    switch (payload.type) {
      case 'payment.completed':
        await ctx.runMutation(internal.payments.markCompleted, {
          paymentId: payload.data.id
        })
        break
      case 'subscription.cancelled':
        await ctx.runMutation(internal.subscriptions.cancel, {
          subscriptionId: payload.data.id
        })
        break
    }
    
    return { processed: true }
  }
})
```

---

## Internal Functions

### Internal Query

Internal functions are only callable from other server functions.

```typescript
// convex/functions/documents.ts
import { internalQuery, internalMutation } from '../_generated/server'
import { v } from 'convex/values'

export const getById = internalQuery({
  args: { id: v.id('documents') },
  handler: async (ctx, args) => {
    return await ctx.db.get(args.id)
  }
})
```

### Internal Mutation

```typescript
export const updateSummary = internalMutation({
  args: { id: v.id('documents'), summary: v.string() },
  handler: async (ctx, args) => {
    await ctx.db.patch(args.id, {
      summary: args.summary,
      updatedAt: Date.now()
    })
  }
})
```

### Calling Internal Functions

```typescript
import { internal } from '../_generated/api'

// From an action
export const processDocument = action({
  args: { documentId: v.id('documents') },
  handler: async (ctx, args) => {
    // Call internal query
    const doc = await ctx.runQuery(internal.documents.getById, {
      id: args.documentId
    })
    
    // Call internal mutation
    await ctx.runMutation(internal.documents.updateSummary, {
      id: args.documentId,
      summary: 'Generated summary'
    })
  }
})
```

---

## Scheduled Functions (Crons)

### Cron Configuration

```typescript
// convex/crons.ts
import { cronJobs } from 'convex/server'
import { internal } from './_generated/api'

const crons = cronJobs()

// Run every hour
crons.interval(
  'cleanup old drafts',
  { hours: 1 },
  internal.documents.cleanupOldDrafts
)

// Run at specific time (cron syntax)
crons.cron(
  'daily analytics',
  '0 0 * * *',  // Midnight UTC
  internal.analytics.generateDailyReport
)

// Run every 5 minutes
crons.interval(
  'sync external data',
  { minutes: 5 },
  internal.sync.fetchExternalUpdates
)

export default crons
```

### Cleanup Job Implementation

```typescript
// convex/functions/documents.ts
export const cleanupOldDrafts = internalMutation({
  handler: async (ctx) => {
    const thirtyDaysAgo = Date.now() - 30 * 24 * 60 * 60 * 1000
    
    const oldDrafts = await ctx.db
      .query('documents')
      .withIndex('by_status_created', (q) =>
        q.eq('status', 'draft').lt('createdAt', thirtyDaysAgo)
      )
      .collect()
    
    for (const draft of oldDrafts) {
      await ctx.db.delete(draft._id)
    }
    
    return { deleted: oldDrafts.length }
  }
})
```

---

## HTTP Endpoints

### HTTP Router Configuration

```typescript
// convex/http.ts
import { httpRouter } from 'convex/server'
import { httpAction } from './_generated/server'
import { internal } from './_generated/api'

const http = httpRouter()

// Webhook endpoint
http.route({
  path: '/webhook/stripe',
  method: 'POST',
  handler: httpAction(async (ctx, request) => {
    const body = await request.text()
    const signature = request.headers.get('stripe-signature')
    
    await ctx.runMutation(internal.payments.processWebhook, {
      body,
      signature
    })
    
    return new Response('OK', { status: 200 })
  })
})

// API endpoint with JSON response
http.route({
  path: '/api/documents/:id',
  method: 'GET',
  handler: httpAction(async (ctx, request) => {
    const url = new URL(request.url)
    const id = url.pathname.split('/').pop()
    
    const doc = await ctx.runQuery(internal.documents.getById, { id })
    
    if (!doc) {
      return new Response(JSON.stringify({ error: 'Not found' }), {
        status: 404,
        headers: { 'Content-Type': 'application/json' }
      })
    }
    
    return new Response(JSON.stringify(doc), {
      status: 200,
      headers: { 'Content-Type': 'application/json' }
    })
  })
})

export default http
```

---

## Error Handling

### ConvexError for Structured Errors

```typescript
import { ConvexError } from 'convex/values'

export const secureOperation = mutation({
  args: { id: v.id('documents') },
  handler: async (ctx, args) => {
    const identity = await ctx.auth.getUserIdentity()
    if (!identity) {
      throw new ConvexError('UNAUTHORIZED')
    }
    
    const doc = await ctx.db.get(args.id)
    if (!doc) {
      throw new ConvexError({
        code: 'NOT_FOUND',
        message: 'Document not found',
        documentId: args.id
      })
    }
    
    if (doc.ownerId !== identity.subject) {
      throw new ConvexError({
        code: 'FORBIDDEN',
        message: 'You do not have access to this document'
      })
    }
  }
})
```

### Client-Side Error Handling

```typescript
import { useMutation } from 'convex/react'
import { api } from '../../convex/_generated/api'

export function useSecureOperation() {
  const mutation = useMutation(api.functions.documents.secureOperation)
  
  return async (id: string) => {
    try {
      await mutation({ id })
    } catch (error) {
      if (error instanceof ConvexError) {
        const data = error.data
        switch (data.code) {
          case 'UNAUTHORIZED':
            redirectToLogin()
            break
          case 'NOT_FOUND':
            showNotification('Document not found')
            break
          case 'FORBIDDEN':
            showNotification('Access denied')
            break
        }
      }
      throw error
    }
  }
}
```

---

## Best Practices

Mutation Design:
- Keep mutations focused and atomic
- Use internal mutations for multi-step operations
- Validate all inputs with the v module
- Always check authorization before writes

Action Design:
- Use actions only for external API calls
- Call mutations from actions for database writes
- Handle external API errors gracefully
- Use retries for transient failures

Scheduled Jobs:
- Keep job handlers idempotent
- Log job results for monitoring
- Use appropriate intervals to avoid rate limits
- Clean up temporary data regularly

---

Version: 1.0.0
Module: Server Functions
Parent Skill: moai-platform-convex
