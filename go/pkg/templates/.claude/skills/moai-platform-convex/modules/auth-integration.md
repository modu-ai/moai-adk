# Authentication Integration Module

Authentication patterns with Clerk, Auth0, and custom providers for Convex applications.

---

## Clerk Integration

### Client Setup with Clerk

```typescript
// src/providers/ConvexClerkProvider.tsx
import { ClerkProvider, useAuth } from '@clerk/clerk-react'
import { ConvexProviderWithClerk } from 'convex/react-clerk'
import { ConvexReactClient } from 'convex/react'

const convex = new ConvexReactClient(import.meta.env.VITE_CONVEX_URL)

export function ConvexClerkProvider({ children }: { children: React.ReactNode }) {
  return (
    <ClerkProvider publishableKey={import.meta.env.VITE_CLERK_PUBLISHABLE_KEY}>
      <ConvexProviderWithClerk client={convex} useAuth={useAuth}>
        {children}
      </ConvexProviderWithClerk>
    </ClerkProvider>
  )
}
```

### Environment Configuration

```bash
# .env.local
VITE_CONVEX_URL=https://your-deployment.convex.cloud
VITE_CLERK_PUBLISHABLE_KEY=pk_test_...
```

### Convex Auth Configuration

```json
// convex/auth.config.ts
export default {
  providers: [
    {
      domain: "https://your-clerk-domain.clerk.accounts.dev",
      applicationID: "convex"
    }
  ]
}
```

---

## Auth0 Integration

### Client Setup with Auth0

```typescript
// src/providers/ConvexAuth0Provider.tsx
import { Auth0Provider, useAuth0 } from '@auth0/auth0-react'
import { ConvexProviderWithAuth0 } from 'convex/react-auth0'
import { ConvexReactClient } from 'convex/react'

const convex = new ConvexReactClient(import.meta.env.VITE_CONVEX_URL)

function ConvexAuth0Inner({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading, getAccessTokenSilently } = useAuth0()
  
  return (
    <ConvexProviderWithAuth0
      client={convex}
      isAuthenticated={isAuthenticated}
      isLoading={isLoading}
      getAccessToken={getAccessTokenSilently}
    >
      {children}
    </ConvexProviderWithAuth0>
  )
}

export function ConvexAuth0Provider({ children }: { children: React.ReactNode }) {
  return (
    <Auth0Provider
      domain={import.meta.env.VITE_AUTH0_DOMAIN}
      clientId={import.meta.env.VITE_AUTH0_CLIENT_ID}
      authorizationParams={{
        redirect_uri: window.location.origin,
        audience: import.meta.env.VITE_AUTH0_AUDIENCE
      }}
    >
      <ConvexAuth0Inner>{children}</ConvexAuth0Inner>
    </Auth0Provider>
  )
}
```

---

## Server-Side Authentication

### Getting User Identity

```typescript
// convex/functions/users.ts
import { query, mutation } from '../_generated/server'
import { v } from 'convex/values'

export const current = query({
  handler: async (ctx) => {
    const identity = await ctx.auth.getUserIdentity()
    if (!identity) return null
    
    return await ctx.db
      .query('users')
      .withIndex('by_token', (q) =>
        q.eq('tokenIdentifier', identity.tokenIdentifier)
      )
      .first()
  }
})
```

### Ensure User Exists

```typescript
export const ensureUser = mutation({
  handler: async (ctx) => {
    const identity = await ctx.auth.getUserIdentity()
    if (!identity) throw new Error('Unauthorized')
    
    // Check if user exists
    const existing = await ctx.db
      .query('users')
      .withIndex('by_token', (q) =>
        q.eq('tokenIdentifier', identity.tokenIdentifier)
      )
      .first()
    
    if (existing) return existing._id
    
    // Create new user
    return await ctx.db.insert('users', {
      tokenIdentifier: identity.tokenIdentifier,
      email: identity.email ?? '',
      name: identity.name ?? '',
      pictureUrl: identity.pictureUrl ?? '',
      createdAt: Date.now()
    })
  }
})
```

### User Schema

```typescript
// convex/schema.ts
import { defineSchema, defineTable } from 'convex/server'
import { v } from 'convex/values'

export default defineSchema({
  users: defineTable({
    tokenIdentifier: v.string(),
    email: v.string(),
    name: v.string(),
    pictureUrl: v.optional(v.string()),
    role: v.optional(v.string()),
    createdAt: v.number()
  })
    .index('by_token', ['tokenIdentifier'])
    .index('by_email', ['email'])
})
```

---

## Authorization Patterns

### Resource Authorization

```typescript
export const getDocument = query({
  args: { id: v.id('documents') },
  handler: async (ctx, args) => {
    const identity = await ctx.auth.getUserIdentity()
    const doc = await ctx.db.get(args.id)
    
    if (!doc) throw new Error('Document not found')
    
    // Public documents are always accessible
    if (doc.isPublic) return doc
    
    // Private documents require authentication
    if (!identity) throw new Error('Unauthorized')
    
    // Check ownership
    if (doc.ownerId === identity.subject) return doc
    
    // Check collaborator access
    const collaborator = await ctx.db
      .query('collaborators')
      .withIndex('by_document_user', (q) =>
        q.eq('documentId', args.id).eq('userId', identity.subject)
      )
      .first()
    
    if (collaborator) return doc
    
    throw new Error('Forbidden')
  }
})
```

### Role-Based Access Control

```typescript
import { internalQuery } from '../_generated/server'

export const getUserRole = internalQuery({
  args: { userId: v.string() },
  handler: async (ctx, args) => {
    const user = await ctx.db
      .query('users')
      .withIndex('by_token', (q) => q.eq('tokenIdentifier', args.userId))
      .first()
    
    return user?.role ?? 'user'
  }
})

export const adminOnlyOperation = mutation({
  args: { action: v.string() },
  handler: async (ctx, args) => {
    const identity = await ctx.auth.getUserIdentity()
    if (!identity) throw new Error('Unauthorized')
    
    const role = await ctx.runQuery(internal.users.getUserRole, {
      userId: identity.tokenIdentifier
    })
    
    if (role !== 'admin') {
      throw new Error('Admin access required')
    }
    
    // Perform admin action
  }
})
```

### Permission-Based Access

```typescript
// convex/schema.ts
permissions: defineTable({
  userId: v.string(),
  resourceId: v.id('documents'),
  permission: v.union(
    v.literal('read'),
    v.literal('write'),
    v.literal('delete'),
    v.literal('admin')
  )
})
  .index('by_user_resource', ['userId', 'resourceId'])

// Authorization helper
export const hasPermission = internalQuery({
  args: {
    userId: v.string(),
    resourceId: v.id('documents'),
    permission: v.string()
  },
  handler: async (ctx, args) => {
    const permissionLevels = ['read', 'write', 'delete', 'admin']
    const requiredLevel = permissionLevels.indexOf(args.permission)
    
    const userPermission = await ctx.db
      .query('permissions')
      .withIndex('by_user_resource', (q) =>
        q.eq('userId', args.userId).eq('resourceId', args.resourceId)
      )
      .first()
    
    if (!userPermission) return false
    
    const userLevel = permissionLevels.indexOf(userPermission.permission)
    return userLevel >= requiredLevel
  }
})
```

---

## Auth Hooks Pattern

### useCurrentUser Hook

```typescript
// src/hooks/useCurrentUser.ts
import { useQuery, useMutation } from 'convex/react'
import { api } from '../../convex/_generated/api'
import { useEffect } from 'react'

export function useCurrentUser() {
  const user = useQuery(api.functions.users.current)
  const ensureUser = useMutation(api.functions.users.ensureUser)
  
  useEffect(() => {
    if (user === null) {
      ensureUser()
    }
  }, [user, ensureUser])
  
  return {
    user,
    isLoading: user === undefined,
    isAuthenticated: user !== null
  }
}
```

### Protected Route Component

```typescript
// src/components/ProtectedRoute.tsx
import { useCurrentUser } from '../hooks/useCurrentUser'
import { Navigate } from 'react-router-dom'

export function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { user, isLoading, isAuthenticated } = useCurrentUser()
  
  if (isLoading) return <LoadingSpinner />
  if (!isAuthenticated) return <Navigate to="/login" />
  
  return <>{children}</>
}
```

### Role-Protected Component

```typescript
export function AdminRoute({ children }: { children: React.ReactNode }) {
  const { user, isLoading } = useCurrentUser()
  
  if (isLoading) return <LoadingSpinner />
  if (!user || user.role !== 'admin') {
    return <Navigate to="/unauthorized" />
  }
  
  return <>{children}</>
}
```

---

## Session Management

### Session Tracking

```typescript
// convex/schema.ts
sessions: defineTable({
  userId: v.string(),
  deviceInfo: v.string(),
  lastActiveAt: v.number(),
  createdAt: v.number()
})
  .index('by_user', ['userId'])
  .index('by_last_active', ['lastActiveAt'])

// Update session activity
export const updateSession = mutation({
  args: { deviceInfo: v.string() },
  handler: async (ctx, args) => {
    const identity = await ctx.auth.getUserIdentity()
    if (!identity) return
    
    const existing = await ctx.db
      .query('sessions')
      .withIndex('by_user', (q) => q.eq('userId', identity.subject))
      .first()
    
    if (existing) {
      await ctx.db.patch(existing._id, { lastActiveAt: Date.now() })
    } else {
      await ctx.db.insert('sessions', {
        userId: identity.subject,
        deviceInfo: args.deviceInfo,
        lastActiveAt: Date.now(),
        createdAt: Date.now()
      })
    }
  }
})
```

---

## Best Practices

Authentication Setup:
- Use official Convex auth provider integrations
- Configure auth providers in convex/auth.config.ts
- Store minimal user data in Convex database

Authorization:
- Always validate authentication in server functions
- Use internal queries for authorization checks
- Implement least-privilege access patterns
- Cache authorization results when appropriate

Security:
- Never trust client-provided user IDs
- Use identity.tokenIdentifier for user identification
- Validate permissions for every protected operation
- Log security-relevant events

---

Version: 1.0.0
Module: Authentication Integration
Parent Skill: moai-platform-convex
