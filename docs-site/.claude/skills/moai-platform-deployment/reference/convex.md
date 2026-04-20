# Convex Platform Reference

Comprehensive guide to Convex real-time reactive backend platform.

---

## Reactive Queries

### Basic Query Structure

**Define Query**:
```typescript
// convex/messages.ts
import { query } from "./_generated/server"
import { v } from "convex/values"

export const list = query({
  args: { limit: v.optional(v.number()) },
  handler: async (ctx, args) => {
    const messages = await ctx.db.query("messages")
      .order("desc")
      .take(args.limit ?? 50)

    return messages
  }
})
```

**Use in React**:
```typescript
import { useQuery } from "convex/react"
import { api } from "../convex/_generated/api"

export function MessageList() {
  const messages = useQuery(api.messages.list, { limit: 20 })

  if (messages === undefined) {
    return <div>Loading...</div>
  }

  return (
    <ul>
      {messages.map((message) => (
        <li key={message._id}>{message.text}</li>
      ))}
    </ul>
  )
}
```

### Index-Based Queries

**Define Index**:
```typescript
// convex/schema.ts
import { defineSchema, defineTable } from "convex/server"
import { v } from "convex/values"

export default defineSchema({
  messages: defineTable({
    text: v.string(),
    author: v.string(),
    timestamp: v.number(),
  })
    .index("by_author", ["author"])
    .index("by_timestamp", ["timestamp"]),
})
```

**Query with Index**:
```typescript
export const listByAuthor = query({
  args: { author: v.string() },
  handler: async (ctx, args) => {
    const messages = await ctx.db
      .query("messages")
      .withIndex("by_author", (q) =>
        q.eq("author", args.author)
      )
      .collect()

    return messages
  }
})
```

### Pagination Patterns

**Cursor-Based Pagination**:
```typescript
export const listPaginated = query({
  args: {
    paginationOpts: v.object({
      numItems: v.number(),
      cursor: v.optional(v.string()),
    }),
  },
  handler: async (ctx, args) => {
    const results = await ctx.db
      .query("messages")
      .paginate(args.paginationOpts)

    return {
      page: results.page,
      continueCursor: results.continueCursor,
      isDone: results.isDone,
    }
  }
})
```

**React Pagination Component**:
```typescript
import { useQuery } from "convex/react"
import { api } from "../convex/_generated/api"
import { useState } from "react"

export function PaginatedMessages() {
  const [cursor, setCursor] = useState<string | undefined>()
  const { page, isDone, continueCursor } = useQuery(
    api.messages.listPaginated,
    { paginationOpts: { numItems: 20, cursor } }
  )

  return (
    <div>
      {page?.map((msg) => <div key={msg._id}>{msg.text}</div>)}
      {!isDone && (
        <button onClick={() => setCursor(continueCursor)}>
          Load More
        </button>
      )}
    </div>
  )
}
```

### Search Indexes

**Define Search Index**:
```typescript
// convex/schema.ts
export default defineSchema({
  posts: defineTable({
    title: v.string(),
    body: v.string(),
    author: v.string(),
    published: v.boolean(),
  })
    .searchIndex("search_title_body", {
      searchField: "title",
      filterFields: ["published"],
    })
    .searchIndex("search_full_text", {
      searchField: "body",
      filterFields: ["author", "published"],
    }),
})
```

**Search Query**:
```typescript
export const searchPosts = query({
  args: {
    searchQuery: v.string(),
    onlyPublished: v.boolean(),
  },
  handler: async (ctx, args) => {
    const results = await ctx.db
      .query("posts")
      .withSearchIndex("search_title_body", (q) =>
        q
          .search("title", args.searchQuery)
          .eq("published", args.onlyPublished)
      )
      .take(20)

    return results
  }
})
```

---

## Server Functions

### Mutations (Write Operations)

**Simple Mutation**:
```typescript
import { mutation } from "./_generated/server"

export const send = mutation({
  args: {
    text: v.string(),
    author: v.string(),
  },
  handler: async (ctx, args) => {
    const messageId = await ctx.db.insert("messages", {
      text: args.text,
      author: args.author,
      timestamp: Date.now(),
    })

    return messageId
  }
})
```

**Use in React**:
```typescript
import { useMutation } from "convex/react"
import { api } from "../convex/_generated/api"

export function MessageForm() {
  const sendMessage = useMutation(api.messages.send)

  const handleSubmit = (text: string) => {
    sendMessage({ text, author: "user" })
  }

  return <form onSubmit={(e) => handleSubmit(/* ... */)}>...</form>
}
```

### Actions (External API Calls)

**HTTP Request**:
```typescript
import { action } from "./_generated/server"

export const fetchWeather = action({
  args: { city: v.string() },
  handler: async (_, args) => {
    const response = await fetch(
      `https://api.weather.com/${args.city}`
    )
    const data = await response.json()
    return data
  }
})
```

**Action with Mutation**:
```typescript
export const updateFromAPI = action({
  args: { id: v.id("todos") },
  handler: async (_, args) => {
    // Fetch from external API
    const data = await fetch(`https://api.example.com/${args.id}`)
    const json = await data.json()

    // Run mutation (internal call)
    const result = await ctx.runMutation(
      internal.todos.updateFromAPI,
      { id: args.id, data: json }
    )

    return result
  }
})
```

### Scheduled Functions (Crons)

**Define Cron**:
```typescript
// convex/crons.ts
import { cronJobs } from "./_generated/server"

cronJobs({
  sendDailyDigest: {
    every: { hours: 24 },
    handler: async (ctx) => {
      const users = await ctx.db.query("users").collect()
      // Send digest emails...
    }
  },
  cleanupOldData: {
    every: { days: 7 },
    handler: async (ctx) => {
      const oldData = await ctx.db
        .query("logs")
        .withIndex("by_timestamp", (q) =>
          q.lt("timestamp", Date.now() - 30 * 24 * 60 * 60 * 1000)
        )
        .collect()

      for (const log of oldData) {
        await ctx.db.delete(log._id)
      }
    }
  }
})
```

### HTTP Endpoints (Webhooks)

**Define HTTP Endpoint**:
```typescript
// convex/http.ts
import { httpRouter } from "convex/server"
import { Webhook } from "svix"

const http = httpRouter()

http.route({
  path: "/stripe/webhook",
  method: "POST",
  handler: async (ctx, request) => {
    const payload = await request.json()
    const webhook = new Webhook(process.env.STRIPE_WEBHOOK_SECRET!)

    try {
      const evt = webhook.verify(JSON.stringify(payload), {
        "svix-id": request.headers.get("svix-id")!,
        "svix-timestamp": request.headers.get("svix-timestamp")!,
        "svix-signature": request.headers.get("svix-signature")!,
      })

      // Handle webhook event
      await ctx.runMutation(internal.stripe.handleWebhook, {
        event: evt.type,
        data: evt.data,
      })

      return new Response(null, { status: 200 })
    } catch (err) {
      return new Response("Invalid signature", { status: 401 })
    }
  }
})

export default http
```

---

## Authentication Integration

### Clerk Integration

**Convex Provider Setup**:
```typescript
// src/ConvexProvider.tsx
import { ConvexProviderWithClerk } from "convex/react-clerk"
import { ConvexReactClient } from "convex/react"
import { useAuth } from "@clerk/clerk-react"

const convex = new ConvexReactClient(
  process.env.NEXT_PUBLIC_CONVEX_URL!
)

export function ConvexClerkProvider({ children }) {
  const { isLoaded, userId } = useAuth()

  return (
    <ConvexProviderWithClerk client={convex} useAuth={useAuth}>
      {children}
    </ConvexProviderWithClerk>
  )
}
```

**Authenticated Queries**:
```typescript
export const myMessages = query({
  args: {},
  handler: async (ctx) => {
    const identity = await ctx.auth.getUserIdentity()
    if (!identity) {
      throw new Error("Not authenticated")
    }

    const messages = await ctx.db
      .query("messages")
      .withIndex("by_author", (q) =>
        q.eq("author", identity.subject)
      )
      .collect()

    return messages
  }
})
```

**Authenticated Mutations**:
```typescript
export const send = mutation({
  args: { text: v.string() },
  handler: async (ctx, args) => {
    const identity = await ctx.auth.getUserIdentity()
    if (!identity) {
      throw new Error("Not authenticated")
    }

    await ctx.db.insert("messages", {
      text: args.text,
      author: identity.subject,
      timestamp: Date.now(),
    })
  }
})
```

### Auth0 Integration

**Similar to Clerk, use Auth0 token**:
```typescript
// In Convex functions
const identity = await ctx.auth.getUserIdentity()
const token = identity?.tokenIdentifier
const userId = token?.replace("Auth0|", "")
```

### Role-Based Access Control

**Define Roles**:
```typescript
// convex/schema.ts
export default defineSchema({
  users: defineTable({
    name: v.string(),
    email: v.string(),
    role: v.union(
      v.literal("admin"),
      v.literal("user"),
      v.literal("guest")
    ),
  }),
})
```

**Check Permissions**:
```typescript
export const deleteUser = mutation({
  args: { userId: v.id("users") },
  handler: async (ctx, args) => {
    const identity = await ctx.auth.getUserIdentity()
    const currentUser = await ctx.db
      .query("users")
      .withIndex("by_token", (q) =>
        q.eq("tokenIdentifier", identity!.tokenIdentifier)
      )
      .unique()

    if (currentUser?.role !== "admin") {
      throw new Error("Unauthorized: Admin only")
    }

    await ctx.db.delete(args.userId)
  }
})
```

---

## File Storage

### File Upload Workflow

**Server Function**:
```typescript
// convex/files.ts
import { mutation } from "./_generated/server"

export const generateUploadUrl = mutation({
  args: {},
  handler: async (ctx) => {
    return await ctx.storage.generateUploadUrl()
  }
})

export const saveFileMetadata = mutation({
  args: {
    storageId: v.id("_storage"),
    name: v.string(),
    size: v.number(),
    type: v.string(),
  },
  handler: async (ctx, args) => {
    const fileId = await ctx.db.insert("files", {
      storageId: args.storageId,
      name: args.name,
      size: args.size,
      type: args.type,
      uploadedAt: Date.now(),
    })

    return fileId
  }
})
```

**Client Upload**:
```typescript
import { useMutation } from "convex/react"
import { api } from "../convex/_generated/api"

export function FileUploader() {
  const generateUploadUrl = useMutation(api.files.generateUploadUrl)
  const saveFileMetadata = useMutation(api.files.saveFileMetadata)

  const handleUpload = async (file: File) => {
    // 1. Get upload URL
    const uploadUrl = await generateUploadUrl()

    // 2. Upload file to storage
    const response = await fetch(uploadUrl, {
      method: "POST",
      headers: { "Content-Type": file.type },
      body: file,
    })

    const { storageId } = await response.json()

    // 3. Save metadata to database
    await saveFileMetadata({
      storageId,
      name: file.name,
      size: file.size,
      type: file.type,
    })
  }

  return <input type="file" onChange={(e) => handleUpload(e.target.files![0])} />
}
```

### File Display

**Serve File**:
```typescript
// convex/files.ts
export const getFileUrl = query({
  args: { fileId: v.id("files") },
  handler: async (ctx, args) => {
    const file = await ctx.db.get(args.fileId)
    if (!file) throw new Error("File not found")

    return await ctx.storage.getUrl(file.storageId)
  }
})
```

**Display Image**:
```typescript
import { useQuery } from "convex/react"
import { api } from "../convex/_generated/api"

export function ImageView({ fileId }: { fileId: string }) {
  const url = useQuery(api.files.getFileUrl, { fileId })

  if (!url) return <div>Loading...</div>

  return <img src={url} alt="Uploaded file" />
}
```

---

## Optimistic Updates

### Basic Pattern

**Update UI Immediately**:
```typescript
import { useMutation, useQuery } from "convex/react"
import { api } from "../convex/_generated/api"

export function TodoList() {
  const todos = useQuery(api.todos.list)
  const addTodo = useMutation(api.todos.add)

  const handleAdd = async (text: string) => {
    // Optimistic update
    const optimisticId = `optimistic-${Date.now()}`
    const optimisticTodo = {
      _id: optimisticId,
      text,
      completed: false,
    }

    // Update local state immediately
    setTodos((prev) => [...prev, optimisticTodo])

    try {
      // Server mutation
      await addTodo({ text })
    } catch (error) {
      // Rollback on error
      setTodos((prev) => prev.filter((t) => t._id !== optimisticId))
    }
  }

  return (
    <ul>
      {todos?.map((todo) => (
        <li key={todo._id}>{todo.text}</li>
      ))}
    </ul>
  )
}
```

### Using React Query for Complex Optimistic Updates

```typescript
import { useMutation, useQueryClient } from "@tanstack/react-query"

export function useAddTodo() {
  const queryClient = useQueryClient()
  const mutation = useMutation({
    mutationFn: api.todos.add,
    onMutate: async (newTodo) => {
      // Cancel outgoing refetches
      await queryClient.cancelQueries({ queryKey: ["todos"] })

      // Snapshot previous value
      const previousTodos = queryClient.getQueryData(["todos"])

      // Optimistically update
      queryClient.setQueryData(["todos"], (old) => [
        ...old,
        { ...newTodo, _id: "optimistic" },
      ])

      // Return context for rollback
      return { previousTodos }
    },
    onError: (err, newTodo, context) => {
      // Rollback on error
      queryClient.setQueryData(["todos"], context?.previousTodos)
    },
    onSettled: () => {
      // Refetch after mutation
      queryClient.invalidateQueries({ queryKey: ["todos"] })
    },
  })

  return mutation
}
```

---

## Best Practices

### Query Optimization

1. **Use indexes for all filtered queries**
2. **Limit query results with `.take()`**
3. **Use pagination for large datasets**
4. **Avoid deep nesting in queries**
5. **Cache expensive computations in database**

### Mutation Design

1. **Keep mutations focused and atomic**
2. **Validate all inputs with validators**
3. **Check authentication/authorization**
4. **Return useful data to client**
5. **Handle errors gracefully**

### Schema Design

1. **Normalize data where appropriate**
2. **Use search indexes for full-text search**
3. **Define indexes for common query patterns**
4. **Use appropriate validators for all fields**
5. **Consider denormalization for read-heavy workloads**

---

## Context7 Integration

For latest Convex documentation:

```typescript
// Step 1: Resolve library
const libraries = await mcp__context7__resolve_library_id({
  libraryName: "convex",
  query: "reactive queries optimistic updates"
})

// Step 2: Get documentation
const docs = await mcp__context7__get_library_docs({
  libraryId: libraries[0].id,
  topic: "reactive-queries",
  maxTokens: 8000
})
```

---

**Status**: Production Ready
**Platform**: Convex Real-Time Backend
**Updated**: 2026-02-09
