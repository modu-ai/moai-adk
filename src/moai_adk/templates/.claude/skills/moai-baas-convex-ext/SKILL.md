# Skill: moai-baas-convex-ext

## Metadata

```yaml
skill_id: moai-baas-convex-ext
skill_name: Convex Realtime Database & Sync Framework
version: 2.0.0
created_date: 2025-11-09
updated_date: 2025-11-09
language: english
triggers:
  - keywords: ["Convex", "Realtime", "Sync", "Database", "TypeScript", "OCC", "Production"]
  - contexts: ["convex-detected", "pattern-f", "realtime-app"]
agents:
  - backend-expert
  - database-expert
  - frontend-expert
  - devops-expert
freedom_level: high
word_count: 1200
context7_references:
  - url: "https://docs.convex.dev/database"
    topic: "Database Design & Schema"
  - url: "https://docs.convex.dev/sync"
    topic: "Realtime Sync Primitives"
  - url: "https://docs.convex.dev/auth"
    topic: "Authentication & Authorization"
  - url: "https://docs.convex.dev/functions"
    topic: "Server Functions"
spec_reference: "@SPEC:BAAS-ECOSYSTEM-001"
```

---

## 📚 Content

### 1. Convex Architecture & Core Concepts (200 words)

**Convex** is a TypeScript-first realtime database platform with native synchronization and built-in backend functions.

**Core Philosophy**:
```
Traditional Approach:
  Client → REST API → Database
  (Manual sync, stale data, complexity)

Convex Approach:
  Client ⟷ Convex Sync (bidirectional, automatic)
  (Native sync, realtime, type-safe)
```

**Key Differences from Firebase/Supabase**:

| Feature | Convex | Firebase | Supabase |
|---------|--------|----------|----------|
| Database | Convex (NoSQL+Relational hybrid) | Firestore (NoSQL) | PostgreSQL (SQL) |
| Sync | Native automatic | Manual listeners | Manual subscriptions |
| Backend | Convex Functions (TypeScript) | Cloud Functions | Edge Functions |
| Type Safety | Full end-to-end | Partial (Firebase Admin SDK) | Partial (via types) |
| Realtime | Built-in | Requires Realtime DB | Via Supabase Realtime |
| Consistency | Strong (OCC) | Eventual | Strong (ACID) |

**Architecture Stack**:
```
┌─────────────────────────────────────┐
│ Convex                              │
├─────────────────────────────────────┤
│ 1. Convex Functions                 │
│    └─ Mutations, Queries, Actions   │
│                                      │
│ 2. Database (TypeScript Schema)     │
│    └─ Documents, Relationships      │
│                                      │
│ 3. Realtime Sync                    │
│    └─ Automatic client sync         │
│                                      │
│ 4. Authentication                   │
│    └─ Auth0, Clerk, Custom JWT      │
│                                      │
│ 5. File Storage                     │
│    └─ Built-in blob storage         │
└─────────────────────────────────────┘
```

---

### 2. Database Design & TypeScript Schema (250 words)

**Convex databases** use TypeScript for schema definition, providing type safety from database to client.

**Basic Schema Definition**:

```typescript
// convex/schema.ts
import { defineSchema, defineTable } from "convex/server";
import { v } from "convex/values";

export default defineSchema({
  users: defineTable({
    email: v.string(),
    name: v.string(),
    avatar: v.optional(v.string()),
    role: v.union(v.literal("admin"), v.literal("user")),
    createdAt: v.number(),
    isActive: v.boolean(),
  })
    .index("email", ["email"])
    .index("createdAt", ["createdAt"]),

  posts: defineTable({
    userId: v.id("users"),
    title: v.string(),
    content: v.string(),
    published: v.boolean(),
    createdAt: v.number(),
    updatedAt: v.number(),
  })
    .index("userId", ["userId"])
    .index("published_createdAt", ["published", "createdAt"]),

  comments: defineTable({
    postId: v.id("posts"),
    userId: v.id("users"),
    text: v.string(),
    createdAt: v.number(),
  })
    .index("postId", ["postId"])
    .index("userId", ["userId"]),
});
```

**Key Concepts**:
- **Indexes**: Define `.index()` for query efficiency
- **Relationships**: Use `v.id("tableName")` for foreign keys
- **Validation**: All fields use `v.*` validators (no runtime type errors)
- **Immutability**: Tables are append-only internally

**Query Examples**:

```typescript
// convex/functions.ts
import { query, mutation } from "./_generated/server";
import { v } from "convex/values";

// Read data (strongly typed)
export const getPosts = query({
  args: { userId: v.id("users") },
  handler: async (ctx, args) => {
    return await ctx.db
      .query("posts")
      .withIndex("userId", (q) => q.eq("userId", args.userId))
      .filter((q) => q.eq(q.field("published"), true))
      .collect();
  },
});

// Write data (strongly typed)
export const createPost = mutation({
  args: {
    userId: v.id("users"),
    title: v.string(),
    content: v.string(),
  },
  handler: async (ctx, args) => {
    const postId = await ctx.db.insert("posts", {
      userId: args.userId,
      title: args.title,
      content: args.content,
      published: false,
      createdAt: Date.now(),
      updatedAt: Date.now(),
    });
    return postId;
  },
});
```

---

### 3. Realtime Sync & Client Integration (250 words)

**Convex Sync** automatically synchronizes database state to clients, eliminating manual cache management.

**Two Sync Patterns**:

#### **Pattern 1: useQuery Hook (Read Sync)**
```typescript
// Client-side (React)
import { useQuery } from "convex/react";
import { api } from "./_generated/api";

export function UserPosts({ userId }: { userId: Id<"users"> }) {
  // Automatically synced, updates in real-time
  const posts = useQuery(api.getPosts, { userId });

  if (posts === undefined) return <div>Loading...</div>;

  return (
    <ul>
      {posts.map((post) => (
        <li key={post._id}>{post.title}</li>
      ))}
    </ul>
  );
}
```

#### **Pattern 2: useMutation Hook (Write Sync)**
```typescript
import { useMutation } from "convex/react";
import { api } from "./_generated/api";

export function CreatePostForm({ userId }: { userId: Id<"users"> }) {
  const createPost = useMutation(api.createPost);

  const handleSubmit = async (title: string, content: string) => {
    // Optimistic update + server sync
    const postId = await createPost({
      userId,
      title,
      content,
    });
    console.log("Created post:", postId);
  };

  return <form onSubmit={(e) => handleSubmit(...)} />;
}
```

**Key Advantages**:
- ✅ **Automatic sync**: No manual refetch logic
- ✅ **Optimistic updates**: UI updates before server confirmation
- ✅ **Type-safe**: TypeScript types flow from backend → client
- ✅ **Offline support**: Built-in offline queue (SyncState)
- ✅ **No N+1 queries**: Convex optimizes queries automatically

#### **Advanced: Offline Support**
```typescript
import { useSyncedQuery } from "convex/react";

export function OfflineAwarePosts() {
  const { syncState, data: posts } = useSyncedQuery(api.getPosts, {});

  return (
    <>
      {syncState === "synced" && <p>✅ All data synced</p>}
      {syncState === "syncing" && <p>🔄 Syncing...</p>}
      {syncState === "offline" && <p>⚠️ Offline mode</p>}

      {posts?.map((post) => (
        <PostCard key={post._id} post={post} />
      ))}
    </>
  );
}
```

---

### 4. Authentication & Authorization (200 words)

**Convex Auth** supports multiple authentication providers with built-in session management.

**Auth Integration Options**:

```typescript
// convex/auth.config.ts
import GitHub from "@auth/core/providers/github";
import { defineAuth } from "convex/server";

export const { auth, signIn, signOut, store } = defineAuth({
  providers: [
    GitHub({
      id: "github",
      clientId: process.env.GITHUB_ID,
      clientSecret: process.env.GITHUB_SECRET,
    }),
  ],
});
```

**User Context in Functions**:

```typescript
export const getCurrentUser = query({
  handler: async (ctx) => {
    // Automatically gets current authenticated user
    const identity = await ctx.auth.getUserIdentity();
    if (!identity) return null;

    return await ctx.db
      .query("users")
      .withIndex("email", (q) => q.eq("email", identity.email))
      .first();
  },
});

export const updateProfile = mutation({
  args: { name: v.string() },
  handler: async (ctx, args) => {
    const identity = await ctx.auth.getUserIdentity();
    if (!identity) throw new Error("Not authenticated");

    const user = await ctx.db
      .query("users")
      .withIndex("email", (q) => q.eq("email", identity.email))
      .first();

    if (!user) throw new Error("User not found");

    await ctx.db.patch(user._id, { name: args.name });
  },
});
```

**Authorization Patterns**:
- **Owner-based**: Check `userId === identity.sub`
- **Role-based**: Store role in users table, check in mutations
- **Team-based**: Join query through team membership table

---

### 5. Advanced Patterns: Actions & Scheduled Functions (120 words)

**Actions** enable multi-step operations with external API calls and transactions.

```typescript
import { action } from "./_generated/server";

export const publishPostWithNotification = action({
  args: { postId: v.id("posts") },
  handler: async (ctx, args) => {
    // Step 1: Update post status
    await ctx.runMutation(api.updatePostStatus, {
      postId: args.postId,
      published: true
    });

    // Step 2: Call external API (email notification)
    await fetch("https://api.sendgrid.com/v3/mail/send", {
      method: "POST",
      headers: { "Authorization": `Bearer ${process.env.SENDGRID_KEY}` },
      body: JSON.stringify({
        to: "subscribers@example.com",
        subject: "New post published!",
      }),
    });

    // Step 3: Log analytics
    await ctx.runMutation(api.logAnalytics, { event: "post_published" });
  },
});

// Scheduled function (daily cleanup)
export const cleanupOldDrafts = internalAction({
  handler: async (ctx) => {
    const oneDayAgo = Date.now() - 24 * 60 * 60 * 1000;
    const oldDrafts = await ctx.runQuery(api.getOldDrafts, { before: oneDayAgo });

    for (const draft of oldDrafts) {
      await ctx.runMutation(api.deleteDraft, { draftId: draft._id });
    }
  },
});
```

**Key Use Cases**:
- ✅ Multi-table transactions
- ✅ External API calls (webhooks, notifications)
- ✅ Complex business logic with multiple steps

---

### 6. Common Patterns & Best Practices (100 words)

| Pattern | Implementation | Use Case |
|---------|-----------------|----------|
| **Pagination** | `query().skip(n).take(10)` | Large datasets |
| **Full-text search** | String index + `.filter()` | Search functionality |
| **Aggregation** | `.collect()` then client-side | Analytics |
| **Cleanup** | Use `scheduled()` functions | Data maintenance |
| **File upload** | Use `generateUploadUrl()` | Media handling |

**Performance Tips**:
- ✅ Define indexes for frequently queried fields
- ✅ Use `.collect()` for small result sets only
- ✅ Batch mutations with Actions for transactions
- ✅ Leverage automatic Convex query optimization

---

### 7. Production Deployment & Cost Optimization (150 words)

**Deployment Strategy**:

```bash
# 1. Environment setup (production vs development)
npm install convex

# 2. Configure production credentials
convex env set GITHUB_ID=...
convex env set GITHUB_SECRET=...

# 3. Deploy to production
convex deploy

# 4. Monitor via dashboard
# https://dashboard.convex.dev → Logs → Monitor real-time activity
```

**Scaling & Cost Considerations**:

| Metric | Free Tier | Pro Tier | Notes |
|--------|-----------|----------|-------|
| **Database size** | 500MB | Unlimited | Auto-scaling |
| **Monthly function calls** | 1M free | $0.30 per 1M | Most cost driver |
| **Concurrent users** | Unlimited | Unlimited | No seat-based pricing |
| **Realtime subscriptions** | Unlimited | Unlimited | Included in calls |

**Cost Optimization Strategies**:
- ✅ **Debounce queries**: Reduce frequency of `useQuery()` with debouncing
- ✅ **Batch operations**: Use Actions for multi-step operations instead of separate calls
- ✅ **Archive old data**: Move historical data to separate tables
- ✅ **Index optimization**: Proper indexes reduce query iterations
- ✅ **Monitor usage**: Dashboard → Billing → Monitor function call trends

**Production Monitoring**:
```typescript
// Log important events for monitoring
export const logMetric = mutation({
  args: { event: v.string(), duration: v.number() },
  handler: async (ctx, args) => {
    await ctx.db.insert("metrics", {
      event: args.event,
      duration: args.duration,
      timestamp: Date.now(),
    });
  },
});
```

---

### 8. Troubleshooting (50 words)

| Issue | Solution |
|-------|----------|
| **Client stale data** | Convex sync is automatic; check auth |
| **Slow queries** | Add missing index via schema |
| **Auth failing** | Verify environment variables |
| **Offline mode stuck** | Check network; Convex auto-reconnects |

---

## 🎯 Usage

### Invocation from Agents
```python
Skill("moai-baas-convex-ext")
# Load when Pattern F (Convex Realtime) detected
```

### Context7 Integration
When Convex platform detected:
- Database schema design guide
- Realtime sync patterns
- Authentication flows
- Server functions

---

## 📚 Reference Materials

- [Convex Database Documentation](https://docs.convex.dev/database)
- [Sync & Realtime Guide](https://docs.convex.dev/sync)
- [Authentication Setup](https://docs.convex.dev/auth)
- [Server Functions](https://docs.convex.dev/functions)

---

## ✅ Validation Checklist

- [x] Architecture overview & core concepts
- [x] Database schema design (TypeScript)
- [x] Realtime sync patterns (useQuery/useMutation)
- [x] Authentication & authorization
- [x] Advanced patterns (Actions, Scheduled Functions)
- [x] Common patterns & best practices
- [x] Production deployment & cost optimization
- [x] Troubleshooting section
- [x] 1200-word target (from 1000)
- [x] English language (policy compliant)
