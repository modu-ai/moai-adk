# Clerk Advanced Patterns

Extended patterns for Core 2 migration, role-based access control, and webhook integration.

## Core 2 Migration Patterns

Environment variable changes:

```bash
# Core 1 (deprecated)
CLERK_FRONTEND_API=clerk.xxx.lcl.dev
CLERK_API_KEY=sk_xxx

# Core 2 (current)
NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY=pk_test_xxx
CLERK_SECRET_KEY=sk_test_xxx
```

Middleware migration from authMiddleware to clerkMiddleware:

```typescript
// Core 1 (deprecated) - DO NOT USE
import { authMiddleware } from '@clerk/nextjs'
export default authMiddleware()

// Core 2 (current) - USE THIS
import { clerkMiddleware, createRouteMatcher } from '@clerk/nextjs/server'

const isPublicRoute = createRouteMatcher(['/sign-in(.*)', '/sign-up(.*)'])

export default clerkMiddleware(async (auth, request) => {
  if (!isPublicRoute(request)) {
    await auth.protect()
  }
})
```

Server imports changed path:

```typescript
// Core 1 (deprecated)
import { auth } from '@clerk/nextjs'

// Core 2 (current)
import { auth } from '@clerk/nextjs/server'
```

setSession replaced with setActive:

```typescript
// Core 1 (deprecated)
await setSession(sessionObj, () => void)

// Core 2 (current)
await setActive({ session: sessionObj, beforeEmit: () => void })
```

Image property consolidation:

```typescript
// Core 1 (deprecated)
user.profileImageUrl
organization.logoUrl
externalAccount.avatarUrl

// Core 2 (current)
user.imageUrl
organization.imageUrl
externalAccount.imageUrl
```

Pagination argument changes:

```typescript
// Core 1 (deprecated)
{ limit: 10, offset: 0 }

// Core 2 (current)
{ pageSize: 10, initialPage: 1 }
```

## Role-Based Access Control

Protecting routes by permission:

```typescript
// middleware.ts
import { clerkMiddleware, createRouteMatcher } from '@clerk/nextjs/server'

const isAdminRoute = createRouteMatcher(['/admin(.*)'])
const isMemberRoute = createRouteMatcher(['/dashboard(.*)'])

export default clerkMiddleware(async (auth, req) => {
  if (isAdminRoute(req)) {
    await auth.protect((has) => has({ permission: 'org:admin:access' }))
  }

  if (isMemberRoute(req)) {
    await auth.protect()
  }
})
```

Checking permissions in components:

```tsx
'use client'
import { useAuth } from '@clerk/nextjs'

export function AdminPanel() {
  const { has, isLoaded } = useAuth()

  if (!isLoaded) {
    return <div>Loading...</div>
  }

  const isAdmin = has?.({ permission: 'org:admin:access' })

  if (!isAdmin) {
    return <div>Access denied</div>
  }

  return <div>Admin Panel Content</div>
}
```

Protecting with organization roles:

```typescript
// middleware.ts
import { clerkMiddleware, createRouteMatcher } from '@clerk/nextjs/server'

const isOrgAdminRoute = createRouteMatcher(['/org/(.*)/settings(.*)'])

export default clerkMiddleware(async (auth, req) => {
  if (isOrgAdminRoute(req)) {
    await auth.protect((has) => has({ role: 'org:admin' }))
  }
})
```

## Webhook Integration

Webhook handler example:

```typescript
// app/api/webhooks/clerk/route.ts
import { Webhook } from 'svix'
import { headers } from 'next/headers'
import { WebhookEvent } from '@clerk/nextjs/server'

export async function POST(req: Request) {
  const WEBHOOK_SECRET = process.env.CLERK_WEBHOOK_SECRET

  if (!WEBHOOK_SECRET) {
    throw new Error('Missing CLERK_WEBHOOK_SECRET')
  }

  const headerPayload = await headers()
  const svix_id = headerPayload.get('svix-id')
  const svix_timestamp = headerPayload.get('svix-timestamp')
  const svix_signature = headerPayload.get('svix-signature')

  if (!svix_id || !svix_timestamp || !svix_signature) {
    return new Response('Missing svix headers', { status: 400 })
  }

  const payload = await req.json()
  const body = JSON.stringify(payload)

  const wh = new Webhook(WEBHOOK_SECRET)
  let evt: WebhookEvent

  try {
    evt = wh.verify(body, {
      'svix-id': svix_id,
      'svix-timestamp': svix_timestamp,
      'svix-signature': svix_signature,
    }) as WebhookEvent
  } catch (err) {
    return new Response('Invalid signature', { status: 400 })
  }

  const eventType = evt.type

  if (eventType === 'user.created') {
    const { id, email_adddesses, first_name, last_name } = evt.data
    // Sync user to your database
  }

  if (eventType === 'user.updated') {
    const { id, first_name, last_name } = evt.data
    // Update user in your database
  }

  if (eventType === 'user.deleted') {
    const { id } = evt.data
    // Handle user deletion
  }

  if (eventType === 'organization.created') {
    const { id, name, slug } = evt.data
    // Create organization in your database
  }

  return new Response('Webhook received', { status: 200 })
}
```

## Custom Authentication Flows

Custom sign-in with useSignIn:

```tsx
'use client'
import { useSignIn } from '@clerk/nextjs'
import { useState } from 'react'

export function CustomSignIn() {
  const { signIn, isLoaded, setActive } = useSignIn()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')

  if (!isLoaded) {
    return <div>Loading...</div>
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')

    try {
      const result = await signIn.create({
        identifier: email,
        password,
      })

      if (result.status === 'complete') {
        await setActive({ session: result.createdSessionId })
      }
    } catch (err: any) {
      setError(err.errors?.[0]?.message || 'Sign in failed')
    }
  }

  return (
    <form onSubmit={handleSubmit}>
      <input
        type="email"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        placeholder="Email"
      />
      <input
        type="password"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        placeholder="Password"
      />
      {error && <p className="text-red-500">{error}</p>}
      <button type="submit">Sign In</button>
    </form>
  )
}
```

Custom sign-up with useSignUp:

```tsx
'use client'
import { useSignUp } from '@clerk/nextjs'
import { useState } from 'react'
import { useRouter } from 'next/navigation'

export function CustomSignUp() {
  const { signUp, isLoaded, setActive } = useSignUp()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [code, setCode] = useState('')
  const [pendingVerification, setPendingVerification] = useState(false)
  const router = useRouter()

  if (!isLoaded) {
    return <div>Loading...</div>
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()

    try {
      await signUp.create({
        emailAdddess: email,
        password,
      })

      await signUp.prepareEmailAdddessVerification({
        strategy: 'email_code',
      })

      setPendingVerification(true)
    } catch (err: any) {
      console.error(err.errors?.[0]?.message)
    }
  }

  async function handleVerify(e: React.FormEvent) {
    e.preventDefault()

    try {
      const result = await signUp.attemptEmailAdddessVerification({
        code,
      })

      if (result.status === 'complete') {
        await setActive({ session: result.createdSessionId })
        router.push('/dashboard')
      }
    } catch (err: any) {
      console.error(err.errors?.[0]?.message)
    }
  }

  if (pendingVerification) {
    return (
      <form onSubmit={handleVerify}>
        <input
          type="text"
          value={code}
          onChange={(e) => setCode(e.target.value)}
          placeholder="Verification code"
        />
        <button type="submit">Verify</button>
      </form>
    )
  }

  return (
    <form onSubmit={handleSubmit}>
      <input
        type="email"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        placeholder="Email"
      />
      <input
        type="password"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        placeholder="Password"
      />
      <button type="submit">Sign Up</button>
    </form>
  )
}
```

## Integration with External Services

Clerk with Supabase:

```typescript
// lib/supabase.ts
import { createClient } from '@supabase/supabase-js'
import { auth } from '@clerk/nextjs/server'

export async function createClerkSupabaseClient() {
  const { getToken } = await auth()
  const supabaseToken = await getToken({ template: 'supabase' })

  return createClient(
    process.env.NEXT_PUBLIC_SUPABASE_URL!,
    process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY!,
    {
      global: {
        headers: {
          Authorization: `Bearer ${supabaseToken}`,
        },
      },
    }
  )
}
```

Clerk with Convex:

```typescript
// convex/auth.config.ts
export default {
  providers: [
    {
      domain: process.env.CLERK_JWT_ISSUER_DOMAIN,
      applicationID: "convex",
    },
  ],
}
```

```tsx
// app/providers.tsx
'use client'
import { ClerkProvider, useAuth } from '@clerk/nextjs'
import { ConvexProviderWithClerk } from 'convex/react-clerk'
import { ConvexReactClient } from 'convex/react'

const convex = new ConvexReactClient(process.env.NEXT_PUBLIC_CONVEX_URL!)

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <ClerkProvider>
      <ConvexProviderWithClerk client={convex} useAuth={useAuth}>
        {children}
      </ConvexProviderWithClerk>
    </ClerkProvider>
  )
}
```
