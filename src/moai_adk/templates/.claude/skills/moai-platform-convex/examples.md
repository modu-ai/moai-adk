# Convex Platform Examples

Complete working examples for common Convex application patterns.

---

## Example 1: Real-time Document Collaboration

### Schema

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
    .index('by_public', ['isPublic', 'createdAt']),

  collaborators: defineTable({
    documentId: v.id('documents'),
    userId: v.string(),
    permission: v.union(v.literal('read'), v.literal('write'))
  })
    .index('by_document', ['documentId'])
    .index('by_user', ['userId'])
})
```

### Server Functions

```typescript
// convex/functions/documents.ts
import { query, mutation } from '../_generated/server'
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
    await ctx.db.patch(id, { ...updates, updatedAt: Date.now() })
  }
})
```

### React Component

```typescript
// src/components/DocumentEditor.tsx
import { useQuery, useMutation } from 'convex/react'
import { api } from '../../convex/_generated/api'
import { useState, useEffect } from 'react'

export function DocumentEditor({ documentId }: { documentId: string }) {
  const document = useQuery(api.functions.documents.getById, { id: documentId })
  const updateDocument = useMutation(api.functions.documents.update)
    .withOptimisticUpdate((localStore, args) => {
      const existing = localStore.getQuery(api.functions.documents.getById, { id: args.id })
      if (existing) {
        localStore.setQuery(api.functions.documents.getById, { id: args.id }, {
          ...existing,
          ...args,
          updatedAt: Date.now()
        })
      }
    })

  const [content, setContent] = useState('')

  useEffect(() => {
    if (document) setContent(document.content)
  }, [document])

  const handleSave = async () => {
    await updateDocument({ id: documentId, content })
  }

  if (!document) return <div>Loading...</div>

  return (
    <div>
      <h1>{document.title}</h1>
      <textarea value={content} onChange={(e) => setContent(e.target.value)} />
      <button onClick={handleSave}>Save</button>
    </div>
  )
}
```

---

## Example 2: Chat Application

### Schema

```typescript
// convex/schema.ts
export default defineSchema({
  messages: defineTable({
    channelId: v.string(),
    authorId: v.string(),
    authorName: v.string(),
    content: v.string(),
    createdAt: v.number()
  })
    .index('by_channel', ['channelId', 'createdAt']),

  channels: defineTable({
    name: v.string(),
    description: v.optional(v.string()),
    isPrivate: v.boolean(),
    createdAt: v.number()
  })
})
```

### Server Functions

```typescript
// convex/functions/messages.ts
export const list = query({
  args: { channelId: v.string(), limit: v.optional(v.number()) },
  handler: async (ctx, args) => {
    return await ctx.db
      .query('messages')
      .withIndex('by_channel', (q) => q.eq('channelId', args.channelId))
      .order('desc')
      .take(args.limit ?? 50)
  }
})

export const send = mutation({
  args: { channelId: v.string(), content: v.string() },
  handler: async (ctx, args) => {
    const identity = await ctx.auth.getUserIdentity()
    if (!identity) throw new Error('Unauthorized')

    return await ctx.db.insert('messages', {
      channelId: args.channelId,
      authorId: identity.subject,
      authorName: identity.name ?? 'Anonymous',
      content: args.content,
      createdAt: Date.now()
    })
  }
})
```

### React Component

```typescript
// src/components/ChatRoom.tsx
import { useQuery, useMutation } from 'convex/react'
import { api } from '../../convex/_generated/api'
import { useState, useRef, useEffect } from 'react'

export function ChatRoom({ channelId }: { channelId: string }) {
  const messages = useQuery(api.functions.messages.list, { channelId })
  const sendMessage = useMutation(api.functions.messages.send)
  const [input, setInput] = useState('')
  const messagesEndRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  const handleSend = async () => {
    if (!input.trim()) return
    await sendMessage({ channelId, content: input })
    setInput('')
  }

  return (
    <div className="chat-room">
      <div className="messages">
        {messages?.slice().reverse().map((msg) => (
          <div key={msg._id} className="message">
            <strong>{msg.authorName}:</strong> {msg.content}
          </div>
        ))}
        <div ref={messagesEndRef} />
      </div>
      <div className="input-area">
        <input
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyPress={(e) => e.key === 'Enter' && handleSend()}
          placeholder="Type a message..."
        />
        <button onClick={handleSend}>Send</button>
      </div>
    </div>
  )
}
```

---

## Example 3: File Upload with Gallery

### Schema

```typescript
// convex/schema.ts
files: defineTable({
  storageId: v.id('_storage'),
  fileName: v.string(),
  fileType: v.string(),
  fileSize: v.number(),
  uploaderId: v.string(),
  uploadedAt: v.number()
})
  .index('by_uploader', ['uploaderId'])
  .index('by_type', ['fileType'])
```

### Server Functions

```typescript
// convex/functions/files.ts
export const generateUploadUrl = mutation({
  handler: async (ctx) => {
    const identity = await ctx.auth.getUserIdentity()
    if (!identity) throw new Error('Unauthorized')
    return await ctx.storage.generateUploadUrl()
  }
})

export const saveFile = mutation({
  args: { storageId: v.id('_storage'), fileName: v.string(), fileType: v.string(), fileSize: v.number() },
  handler: async (ctx, args) => {
    const identity = await ctx.auth.getUserIdentity()
    if (!identity) throw new Error('Unauthorized')
    
    return await ctx.db.insert('files', {
      ...args,
      uploaderId: identity.subject,
      uploadedAt: Date.now()
    })
  }
})

export const listImages = query({
  args: { uploaderId: v.string() },
  handler: async (ctx, args) => {
    const files = await ctx.db
      .query('files')
      .withIndex('by_uploader', (q) => q.eq('uploaderId', args.uploaderId))
      .collect()
    
    return Promise.all(
      files
        .filter((f) => f.fileType.startsWith('image/'))
        .map(async (f) => ({
          ...f,
          url: await ctx.storage.getUrl(f.storageId)
        }))
    )
  }
})
```

### React Component

```typescript
// src/components/ImageGallery.tsx
import { useQuery, useMutation } from 'convex/react'
import { api } from '../../convex/_generated/api'
import { useCallback } from 'react'

export function ImageGallery({ userId }: { userId: string }) {
  const images = useQuery(api.functions.files.listImages, { uploaderId: userId })
  const generateUploadUrl = useMutation(api.functions.files.generateUploadUrl)
  const saveFile = useMutation(api.functions.files.saveFile)

  const handleUpload = useCallback(async (file: File) => {
    const uploadUrl = await generateUploadUrl()
    const response = await fetch(uploadUrl, {
      method: 'POST',
      headers: { 'Content-Type': file.type },
      body: file
    })
    const { storageId } = await response.json()
    await saveFile({
      storageId,
      fileName: file.name,
      fileType: file.type,
      fileSize: file.size
    })
  }, [generateUploadUrl, saveFile])

  return (
    <div className="gallery">
      <input
        type="file"
        accept="image/*"
        onChange={(e) => e.target.files?.[0] && handleUpload(e.target.files[0])}
      />
      <div className="images">
        {images?.map((img) => (
          <img key={img._id} src={img.url ?? ''} alt={img.fileName} />
        ))}
      </div>
    </div>
  )
}
```

---

## Example 4: Task Management with Crons

### Schema

```typescript
// convex/schema.ts
tasks: defineTable({
  title: v.string(),
  description: v.optional(v.string()),
  status: v.union(v.literal('todo'), v.literal('in_progress'), v.literal('done')),
  dueDate: v.optional(v.number()),
  assigneeId: v.optional(v.string()),
  createdAt: v.number(),
  updatedAt: v.number()
})
  .index('by_status', ['status'])
  .index('by_due_date', ['dueDate'])
  .index('by_assignee', ['assigneeId'])
```

### Scheduled Cleanup

```typescript
// convex/crons.ts
import { cronJobs } from 'convex/server'
import { internal } from './_generated/api'

const crons = cronJobs()

crons.cron(
  'cleanup completed tasks',
  '0 0 * * 0',  // Every Sunday at midnight
  internal.tasks.cleanupCompleted
)

crons.interval(
  'send due date reminders',
  { hours: 1 },
  internal.tasks.sendReminders
)

export default crons
```

### Internal Functions

```typescript
// convex/functions/tasks.ts
import { internalMutation, internalAction } from '../_generated/server'

export const cleanupCompleted = internalMutation({
  handler: async (ctx) => {
    const thirtyDaysAgo = Date.now() - 30 * 24 * 60 * 60 * 1000
    
    const oldTasks = await ctx.db
      .query('tasks')
      .withIndex('by_status', (q) => q.eq('status', 'done'))
      .collect()
    
    const toDelete = oldTasks.filter((t) => t.updatedAt < thirtyDaysAgo)
    
    for (const task of toDelete) {
      await ctx.db.delete(task._id)
    }
    
    return { deleted: toDelete.length }
  }
})

export const sendReminders = internalAction({
  handler: async (ctx) => {
    const tomorrow = Date.now() + 24 * 60 * 60 * 1000
    
    const dueSoon = await ctx.runQuery(internal.tasks.getDueSoon, { before: tomorrow })
    
    for (const task of dueSoon) {
      // Send notification via external service
      await sendNotification(task.assigneeId, `Task "${task.title}" is due soon!`)
    }
  }
})
```

---

## Example 5: Webhook Integration

### HTTP Endpoint

```typescript
// convex/http.ts
import { httpRouter } from 'convex/server'
import { httpAction } from './_generated/server'
import { internal } from './_generated/api'

const http = httpRouter()

http.route({
  path: '/webhook/payment',
  method: 'POST',
  handler: httpAction(async (ctx, request) => {
    const signature = request.headers.get('x-webhook-signature')
    const body = await request.text()
    
    // Verify signature
    if (!verifyWebhookSignature(body, signature)) {
      return new Response('Invalid signature', { status: 401 })
    }
    
    const payload = JSON.parse(body)
    
    await ctx.runMutation(internal.payments.processWebhook, {
      eventType: payload.type,
      data: payload.data
    })
    
    return new Response('OK', { status: 200 })
  })
})

export default http
```

### Payment Processing

```typescript
// convex/functions/payments.ts
import { internalMutation } from '../_generated/server'
import { v } from 'convex/values'

export const processWebhook = internalMutation({
  args: { eventType: v.string(), data: v.any() },
  handler: async (ctx, args) => {
    switch (args.eventType) {
      case 'payment.completed':
        await ctx.db.insert('payments', {
          externalId: args.data.id,
          amount: args.data.amount,
          status: 'completed',
          processedAt: Date.now()
        })
        break
      
      case 'payment.failed':
        await ctx.db.insert('payments', {
          externalId: args.data.id,
          amount: args.data.amount,
          status: 'failed',
          error: args.data.error,
          processedAt: Date.now()
        })
        break
    }
  }
})
```

---

Version: 2.0.0
Type: Examples
Parent Skill: moai-platform-convex
