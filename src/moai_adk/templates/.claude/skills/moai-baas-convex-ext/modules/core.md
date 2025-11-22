

## Convex Real-Time Backend Platform (November 2025)

### Core Features Overview
- **Reactive Database**: Automatic real-time synchronization across clients
- **Type-Safe Backend**: End-to-end TypeScript with compile-time guarantees
- **Serverless Functions**: Backend functions with automatic scaling
- **Built-in Authentication**: User management and access control
- **Real-time Subscriptions**: Live data updates with automatic conflict resolution

### Latest Features (November 2025)
- **Self-Hosted Convex**: On-premises deployment with PostgreSQL support
- **Dashboard for Self-Hosted**: Management interface for self-hosted deployments
- **Open-Source Reactive Database**: Community-driven development
- **Enhanced TypeScript Support**: Improved type inference and performance

### Key Benefits
- **Zero Configuration**: Automatic deployment and scaling
- **Type Safety**: Compile-time error prevention
- **Real-time by Default**: Automatic synchronization without extra code
- **Developer Experience**: Modern TypeScript tooling and debugging

### Performance Characteristics
- **Real-time Sync**: P95 < 100ms latency
- **Function Execution**: Sub-second cold starts
- **Automatic Scaling**: Handles millions of concurrent users
- **Conflict Resolution**: Automatic merge and resolution strategies


# Core Implementation (Level 2)

## Real-Time Schema Design

```typescript
// Convex schema definition with TypeScript
import { defineSchema, defineTable } from "convex/server";
import { v } from "convex/values";

export default defineSchema({
  // User management with real-time presence
  users: defineTable({
    name: v.string(),
    email: v.string(),
    avatar: v.optional(v.string()),
    status: v.union(v.literal("online"), v.literal("offline"), v.literal("away")),
    lastSeen: v.number(),
    presence: v.optional(v.object({
      currentRoom: v.id("rooms"),
      lastActivity: v.number(),
      cursor: v.optional(v.object({
        x: v.number(),
        y: v.number(),
      })),
    })),
  })
    .index("by_email", ["email"])
    .index("by_status", ["status"]),

  // Collaborative rooms with real-time state
  rooms: defineTable({
    name: v.string(),
    description: v.optional(v.string()),
    type: v.union(v.literal("document"), v.literal("whiteboard"), v.literal("chat")),
    createdBy: v.id("users"),
    createdAt: v.number(),
    isPublic: v.boolean(),
    maxParticipants: v.optional(v.number()),
  })
    .index("by_created_by", ["createdBy"])
    .index("by_type", ["type"]),

  // Real-time documents with collaborative editing
  documents: defineTable({
    title: v.string(),
    content: v.optional(v.string()),
    roomId: v.id("rooms"),
    createdBy: v.id("users"),
    createdAt: v.number(),
    updatedAt: v.number(),
    version: v.number(),
    isLocked: v.boolean(),
    lockedBy: v.optional(v.id("users")),
  })
    .index("by_room", ["roomId"])
    .index("by_updated_at", ["updatedAt"]),

  // Real-time messages and activities
  messages: defineTable({
    content: v.string(),
    author: v.id("users"),
    roomId: v.id("rooms"),
    timestamp: v.number(),
    type: v.union(v.literal("text"), v.literal("system"), v.literal("file")),
    metadata: v.optional(v.any()),
  })
    .index("by_room_timestamp", ["roomId", "timestamp"])
    .index("by_author", ["author"]),
});
```

## Real-Time Function Implementation

```typescript
// Convex functions with real-time capabilities
import { mutation, query, action } from "./_generated/server";
import { v } from "convex/values";

// Real-time presence updates
export const updatePresence = mutation({
  args: {
    status: v.union(v.literal("online"), v.literal("offline"), v.literal("away")),
    currentRoom: v.optional(v.id("rooms")),
    cursor: v.optional(v.object({ x: v.number(), y: v.number() })),
  },
  handler: async (ctx, args) => {
    const identity = await ctx.auth.getUserIdentity();
    if (!identity) throw new Error("Unauthorized");

    const user = await ctx.db
      .query("users")
      .withIndex("by_email", (q) => q.eq("email", identity.email!))
      .unique();

    if (!user) throw new Error("User not found");

    const updateData: any = {
      status: args.status,
      lastSeen: Date.now(),
    };

    if (args.currentRoom || args.cursor) {
      updateData.presence = {
        currentRoom: args.currentRoom,
        lastActivity: Date.now(),
        cursor: args.cursor,
      };
    }

    await ctx.db.patch(user._id, updateData);
    return user._id;
  },
});

// Real-time document collaboration
export const updateDocument = mutation({
  args: {
    documentId: v.id("documents"),
    content: v.string(),
    version: v.number(),
  },
  handler: async (ctx, args) => {
    const identity = await ctx.auth.getUserIdentity();
    if (!identity) throw new Error("Unauthorized");

    const document = await ctx.db.get(args.documentId);
    if (!document) throw new Error("Document not found");

    // Check if document is locked by another user
    if (document.isLocked && document.lockedBy !== document.createdBy) {
      throw new Error("Document is locked by another user");
    }

    // Version conflict detection
    if (document.version !== args.version) {
      throw new Error("Version conflict - document was modified");
    }

    const updatedDocument = await ctx.db.patch(args.documentId, {
      content: args.content,
      updatedAt: Date.now(),
      version: args.version + 1,
    });

    // Trigger real-time update notification
    await ctx.scheduler.runAfter(0, internal.notifications.notifyDocumentUpdate, {
      documentId: args.documentId,
      roomId: document.roomId,
      updatedBy: document.createdBy,
    });

    return updatedDocument;
  },
});

// Real-time room activity monitoring
export const getRoomActivity = query({
  args: { roomId: v.id("rooms") },
  handler: async (ctx, args) => {
    const room = await ctx.db.get(args.roomId);
    if (!room || !room.isPublic) throw new Error("Room not accessible");

    const [messages, activeUsers] = await Promise.all([
      // Get recent messages
      ctx.db
        .query("messages")
        .withIndex("by_room_timestamp", (q) => 
          q.eq("roomId", args.roomId).order("desc").take(50)
        ),
      
      // Get active users in room
      ctx.db
        .query("users")
        .withIndex("by_status", (q) => q.eq("status", "online"))
        .collect()
        .then(users => users.filter(user => 
          user.presence?.currentRoom === args.roomId
        )),
    ]);

    return {
      messages,
      activeUsers: activeUsers.map(user => ({
        id: user._id,
        name: user.name,
        avatar: user.avatar,
        presence: user.presence,
      })),
    };
  },
});
```


# Advanced Implementation (Level 3)




## Reference & Resources

See [reference.md](reference.md) for detailed API reference and official documentation.



## Context7 Integration

### Related Libraries & Tools
- [Convex](/get-convex/convex-backend): Real-time backend platform

### Official Documentation
- [Documentation](https://docs.convex.dev/)
- [API Reference](https://docs.convex.dev/api/)

### Version-Specific Guides
Latest stable version: Latest
- [Release Notes](https://news.convex.dev/)
- [Migration Guide](https://docs.convex.dev/production/)
