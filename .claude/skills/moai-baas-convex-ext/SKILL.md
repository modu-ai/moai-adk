---
name: moai-baas-convex-ext
description: Moai Baas Convex Ext - Professional implementation guide
version: 1.0.0
modularized: false
tags:
  - backend-as-a-service
  - platform
  - convex
  - enterprise
updated: 2025-11-24
status: active
---

## 📊 Skill Metadata

**version**: 1.0.0  
**modularized**: false  
**last_updated**: 2025-11-22  
**compliance_score**: 75%  
**auto_trigger_keywords**: ext, convex, moai, baas  


$(head -n 30 /tmp/parent.md | head -n 20)

---

## Authentication & Authorization

### Row-Level Security

```typescript
// convex/schema.ts
export default defineSchema({
  messages: defineTable({
    text: v.string(),
    userId: v.id("users"),
    // Automatic RLS via userId
  }).index("by_user", ["userId"])
});
```

### Function Authentication

```typescript
// convex/messages.ts
export const sendMessage = mutation({
  args: { text: v.string() },
  handler: async (ctx, args) => {
    const identity = await ctx.auth.getUserIdentity();
    if (!identity) throw new Error("Unauthenticated");

    await ctx.db.insert("messages", {
      text: args.text,
      userId: identity.subject
    });
  }
});
```

### Security Best Practices
- Token-based authentication via Clerk/Auth0
- Automatic HTTPS enforcement
- Input validation via Zod schemas

---

## Implementation

For detailed patterns:
- **Core Implementation**: modules/core.md

---

**End of Skill** | Updated 2025-11-21
