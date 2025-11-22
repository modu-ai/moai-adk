---
name: moai-baas-convex-ext
description: Enterprise Convex Real-Time Backend with AI-powered reactive database
version: 1.0.0
modularized: false
last_updated: 2025-11-22
compliance_score: 61
auto_trigger_keywords:
  - baas
  - convex
  - ext
  - typescript
category_tier: 1
---

## Quick Reference (30 seconds)

# Enterprise Convex Real-Time Backend Expert 

---

## When to Use

**Automatic triggers**:
- Convex real-time backend architecture and synchronization discussions
- Full-stack TypeScript application development with real-time features
- Collaborative application design and real-time data management
- Real-time database optimization and performance tuning

**Manual invocation**:
- Designing enterprise Convex architectures with optimal real-time patterns
- Implementing collaborative features and real-time synchronization
- Planning migrations from traditional backends to Convex
- Optimizing real-time performance and data consistency

---

# Quick Reference (Level 1)

## Advanced Real-Time Patterns

```typescript
// Collaborative editing with operational transformation
export const applyEdit = mutation({
  args: {
    documentId: v.id("documents"),
    operation: v.object({
      type: v.union(v.literal("insert"), v.literal("delete"), v.literal("replace")),
      position: v.number(),
      content: v.optional(v.string()),
      length: v.optional(v.number()),
    }),
    clientVersion: v.number(),
  },
  handler: async (ctx, args) => {
    const document = await ctx.db.get(args.documentId);
    if (!document) throw new Error("Document not found");

    // Apply operational transformation
    const transformedOp = await transformOperation(
      args.operation,
      document.pendingOperations || []
    );

    // Apply transformed operation to document
    const newContent = applyOperation(document.content || "", transformedOp);
    
    // Update document with new content and clear pending operations
    await ctx.db.patch(args.documentId, {
      content: newContent,
      updatedAt: Date.now(),
      version: document.version + 1,
      pendingOperations: [],
    });

    // Broadcast operation to other clients
    await ctx.scheduler.runAfter(0, internal.realtime.broadcastOperation, {
      documentId: args.documentId,
      operation: transformedOp,
      author: args.operation.author,
    });

    return { success: true, newVersion: document.version + 1 };
  },
});

// Real-time cursor tracking for collaborative cursors
export const updateCursor = mutation({
  args: {
    roomId: v.id("rooms"),
    cursor: v.object({ x: v.number(), y: v.number() }),
  },
  handler: async (ctx, args) => {
    const identity = await ctx.auth.getUserIdentity();
    if (!identity) throw new Error("Unauthorized");

    const user = await ctx.db
      .query("users")
      .withIndex("by_email", (q) => q.eq("email", identity.email!))
      .unique();

    if (!user) throw new Error("User not found");

    // Update user's cursor position
    await ctx.db.patch(user._id, {
      presence: {
        currentRoom: args.roomId,
        lastActivity: Date.now(),
        cursor: args.cursor,
      },
    });

    // Broadcast cursor position to other users in the room
    await ctx.scheduler.runAfter(0, internal.realtime.broadcastCursor, {
      roomId: args.roomId,
      userId: user._id,
      cursor: args.cursor,
    });

    return { success: true };
  },
});
```

### Self-Hosting Configuration

```typescript
// Self-hosted Convex configuration with PostgreSQL
import { ConvexHttpClient } from "convex/browser";

export class SelfHostedConvex {
  private client: ConvexHttpClient;
  private postgresConfig: PostgresConfig;

  constructor(config: SelfHostedConfig) {
    this.client = new ConvexHttpClient(config.convexUrl);
    this.postgresConfig = config.postgres;
  }

  // Initialize self-hosted Convex with PostgreSQL
  async initialize(): Promise<void> {
    // Configure PostgreSQL persistence
    await this.configurePostgresPersistence();
    
    // Set up replication for high availability
    await this.configureReplication();
    
    // Initialize monitoring and metrics
    await this.setupMonitoring();
  }

  private async configurePostgresPersistence(): Promise<void> {
    const persistenceConfig = {
      host: this.postgresConfig.host,
      port: this.postgresConfig.port,
      database: this.postgresConfig.database,
      username: this.postgresConfig.username,
      password: this.postgresConfig.password,
      ssl: this.postgresConfig.ssl,
      
      // Optimization settings for Convex
      poolSize: 20,
      connectionTimeout: 5000,
      statementTimeout: 10000,
      
      // Replication settings
      replicationMode: "streaming",
      walLevel: "logical",
      maxWalSenders: 5,
    };

    await this.client.action(internal.persistence.configure, persistenceConfig);
  }

  // Configure multi-region deployment
  async configureMultiRegion(deploymentConfig: MultiRegionConfig): Promise<void> {
    const regions = [
      { name: "us-east-1", primary: true },
      { name: "eu-west-1", primary: false },
      { name: "ap-southeast-1", primary: false },
    ];

    for (const region of regions) {
      await this.deployRegion(region, deploymentConfig);
    }

    // Configure cross-region replication
    await this.setupCrossRegionReplication(regions);
  }
}
```

### Performance Optimization

```typescript
// Real-time performance optimization strategies
export class ConvexPerformanceOptimizer {
  // Optimize function execution with batching
  static async batchUpdates<T>(
    updates: Array<{ id: Id<any>; data: Partial<T> }>,
    batch: any
  ): Promise<void> {
    const BATCH_SIZE = 100;
    
    for (let i = 0; i < updates.length; i += BATCH_SIZE) {
      const batchUpdates = updates.slice(i, i + BATCH_SIZE);
      
      await batch.run(async (ctx) => {
        const promises = batchUpdates.map(({ id, data }) =>
          ctx.db.patch(id, data)
        );
        await Promise.all(promises);
      });
    }
  }

  // Optimize real-time subscriptions with smart filtering
  static optimizeRealtimeSubscriptions(
    roomId: Id<"rooms">,
    userId: Id<"users">,
    subscriptionFilters: SubscriptionFilters
  ) {
    return {
      // Only subscribe to relevant data
      messages: subscriptionFilters.includeMessages
        ? q => q.eq("roomId", roomId).gte("timestamp", subscriptionFilters.since)
        : null,
      
      // Only track presence for active users
      users: subscriptionFilters.trackPresence
        ? q => q.eq("status", "online").eq("presence.currentRoom", roomId)
        : null,
      
      // Optimize document updates for collaboration
      documents: subscriptionFilters.trackDocumentChanges
        ? q => q.eq("roomId", roomId).gt("updatedAt", subscriptionFilters.lastSeen)
        : null,
    };
  }
}
```

---

# Reference & Integration (Level 4)

---

## Core Implementation

## What It Does

Enterprise Convex Real-Time Backend expert with AI-powered reactive database architecture, Context7 integration, and intelligent synchronization orchestration for collaborative applications.

**Revolutionary  capabilities**:
- 🤖 **AI-Powered Convex Architecture** using Context7 MCP for latest real-time patterns
- 📊 **Intelligent Synchronization Orchestration** with automated real-time optimization
- 🚀 **Advanced TypeScript Integration** with AI-driven type safety and performance
- 🔗 **Enterprise Reactive Patterns** with zero-configuration real-time workflows
- 📈 **Predictive Performance Analytics** with usage forecasting and optimization

---

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

---

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

---

# Advanced Implementation (Level 3)



---

## Reference & Resources

See [reference.md](reference.md) for detailed API reference and official documentation.


---

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