---
name: moai-platform-clerk
description: Clerk modern authentication specialist covering WebAuthn, passkeys, passwordless, and beautiful UI components. Use when implementing modern auth with great UX.
version: 2.0.0
category: platform
tags: clerk, webauthn, passkeys, passwordless, authentication
context7-libraries: /clerk/clerk-docs
related-skills: moai-platform-auth0, moai-lang-typescript
updated: 2025-12-30
status: active
allowed-tools: Read, Write, Bash, Grep, Glob
---

# Clerk Modern Authentication Specialist

Modern authentication platform with WebAuthn, passkeys, passwordless flows, beautiful pre-built UI components, and multi-tenant organization support.

SDK Versions (as of December 2025):
- @clerk/nextjs: 6.x (Core 2, requires Next.js 13.0.4+, React 18+)
- @clerk/clerk-react: 5.x (Core 2, requires React 18+)
- @clerk/express: 1.x
- Node.js: 18.17.0+ required

## Quick Reference (30 seconds)

Environment Variables:

```bash
# .env.local
NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY=pk_test_...
CLERK_SECRET_KEY=sk_test_...
NEXT_PUBLIC_CLERK_SIGN_IN_URL=/sign-in
NEXT_PUBLIC_CLERK_SIGN_UP_URL=/sign-up
```

ClerkProvider Setup (app/layout.tsx):

```tsx
import { ClerkProvider } from '@clerk/nextjs'

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <ClerkProvider>
      <html lang="en">
        <body>{children}</body>
      </html>
    </ClerkProvider>
  )
}
```

Basic Middleware (middleware.ts):

```typescript
import { clerkMiddleware } from '@clerk/nextjs/server'

export default clerkMiddleware()

export const config = {
  matcher: [
    '/((?!_next|[^?]*\\.(?:html?|css|js(?!on)|jpe?g|webp|png|gif|svg|ttf|woff2?|ico|csv|docx?|xlsx?|zip|webmanifest)).*)',
    '/(api|trpc)(.*)',
  ],
}
```

Context7 Access:
- Library: /clerk/clerk-docs
- Resolution: Use resolve-library-id with "clerk" then get-library-docs

---

## Implementation Guide

### ClerkProvider with Authentication Components

Full layout with sign-in/sign-out controls:

```tsx
// app/layout.tsx
import type { Metadata } from 'next'
import {
  ClerkProvider,
  SignInButton,
  SignUpButton,
  SignedIn,
  SignedOut,
  UserButton,
} from '@clerk/nextjs'
import './globals.css'

export default function RootLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  return (
    <ClerkProvider>
      <html lang="en">
        <body>
          <header className="flex justify-end items-center p-4 gap-4 h-16">
            <SignedOut>
              <SignInButton />
              <SignUpButton />
            </SignedOut>
            <SignedIn>
              <UserButton />
            </SignedIn>
          </header>
          {children}
        </body>
      </html>
    </ClerkProvider>
  )
}
```

### Protecting Routes with Middleware

Route protection with createRouteMatcher:

```typescript
// middleware.ts
import { clerkMiddleware, createRouteMatcher } from '@clerk/nextjs/server'

const isProtectedRoute = createRouteMatcher([
  '/dashboard(.*)',
  '/forum(.*)',
  '/api/private(.*)',
])

export default clerkMiddleware(async (auth, req) => {
  if (isProtectedRoute(req)) {
    await auth.protect()
  }
})

export const config = {
  matcher: [
    '/((?!_next|[^?]*\\.(?:html?|css|js(?!on)|jpe?g|webp|png|gif|svg|ttf|woff2?|ico|csv|docx?|xlsx?|zip|webmanifest)).*)',
    '/(api|trpc)(.*)',
  ],
}
```

Protecting all routes except public:

```typescript
// middleware.ts
import { clerkMiddleware, createRouteMatcher } from '@clerk/nextjs/server'

const isPublicRoute = createRouteMatcher([
  '/sign-in(.*)',
  '/sign-up(.*)',
  '/',
  '/about',
])

export default clerkMiddleware(async (auth, req) => {
  if (!isPublicRoute(req)) {
    await auth.protect()
  }
})
```

### useAuth Hook

Access authentication state and tokens:

```tsx
'use client'
import { useAuth } from '@clerk/nextjs'

export default function ExternalDataPage() {
  const { userId, sessionId, getToken, isLoaded, isSignedIn } = useAuth()

  const fetchExternalData = async () => {
    const token = await getToken()
    const response = await fetch('https://api.example.com/data', {
      headers: { Authorization: `Bearer ${token}` },
    })
    return response.json()
  }

  if (!isLoaded) return <div>Loading...</div>
  if (!isSignedIn) return <div>Sign in to view this page</div>

  return (
    <div>
      <p>User ID: {userId}</p>
      <p>Session ID: {sessionId}</p>
      <button onClick={fetchExternalData}>Fetch Data</button>
    </div>
  )
}
```

### useUser Hook

Access user profile data:

```tsx
'use client'
import { useUser } from '@clerk/nextjs'

export default function ProfilePage() {
  const { isSignedIn, user, isLoaded } = useUser()

  if (!isLoaded) return <div>Loading...</div>
  if (!isSignedIn) return <div>Sign in to view your profile</div>

  return (
    <div>
      <h1>Welcome, {user.firstName}!</h1>
      <p>Email: {user.primaryEmailAddress?.emailAddress}</p>
      <img src={user.imageUrl} alt="Profile" width={100} height={100} />
    </div>
  )
}
```

### SignIn and SignUp Pages

Dedicated authentication pages:

```tsx
// app/sign-in/[[...sign-in]]/page.tsx
import { SignIn } from '@clerk/nextjs'

export default function SignInPage() {
  return (
    <div className="flex min-h-screen items-center justify-center">
      <SignIn />
    </div>
  )
}
```

```tsx
// app/sign-up/[[...sign-up]]/page.tsx
import { SignUp } from '@clerk/nextjs'

export default function SignUpPage() {
  return (
    <div className="flex min-h-screen items-center justify-center">
      <SignUp />
    </div>
  )
}
```

### Server-Side Authentication

App Router server components:

```tsx
// app/dashboard/page.tsx
import { auth, currentUser } from '@clerk/nextjs/server'
import { redirect } from 'next/navigation'

export default async function DashboardPage() {
  const { userId } = await auth()
  if (!userId) redirect('/sign-in')

  const user = await currentUser()

  return (
    <div>
      <h1>Dashboard</h1>
      <p>Welcome, {user?.firstName}!</p>
    </div>
  )
}
```

Route Handler authentication:

```typescript
// app/api/user/route.ts
import { auth } from '@clerk/nextjs/server'
import { NextResponse } from 'next/server'

export async function GET() {
  const { userId } = await auth()
  if (!userId) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  }
  return NextResponse.json({ userId })
}
```

### Organization Management

OrganizationSwitcher component:

```tsx
// app/dashboard/layout.tsx
import { OrganizationSwitcher } from '@clerk/nextjs'

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <div>
      <nav className="flex items-center gap-4 p-4">
        <OrganizationSwitcher />
      </nav>
      {children}
    </div>
  )
}
```

Custom organization switcher with useOrganizationList:

```tsx
'use client'
import { useOrganizationList } from '@clerk/nextjs'

export function CustomOrganizationSwitcher() {
  const { isLoaded, setActive, userMemberships } = useOrganizationList({
    userMemberships: { infinite: true },
  })

  if (!isLoaded) return <p>Loading...</p>

  return (
    <div>
      <h2>Your Organizations</h2>
      <ul>
        {userMemberships.data?.map((membership) => (
          <li key={membership.id}>
            <span>{membership.organization.name}</span>
            <button
              onClick={() => setActive({ organization: membership.organization.id })}
            >
              Select
            </button>
          </li>
        ))}
      </ul>
      {userMemberships.hasNextPage && (
        <button onClick={() => userMemberships.fetchNext()}>Load more</button>
      )}
    </div>
  )
}
```

Access current organization with useOrganization:

```tsx
'use client'
import { useOrganization } from '@clerk/nextjs'

export function OrganizationInfo() {
  const { organization, isLoaded } = useOrganization()

  if (!isLoaded) return <div>Loading...</div>
  if (!organization) return <div>No organization selected</div>

  return (
    <div>
      <h2>{organization.name}</h2>
      <img src={organization.imageUrl} alt={organization.name} width={64} />
      <p>Members: {organization.membersCount}</p>
    </div>
  )
}
```

---

## Advanced Patterns

For advanced implementation patterns including Core 2 migration, role-based access control, webhooks, custom auth flows, and external service integration, see:

- [Advanced Patterns](modules/advanced-patterns.md)

Key Core 2 Migration Changes:
- Environment: CLERK_FRONTEND_API to NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY
- Middleware: authMiddleware() to clerkMiddleware()
- Imports: '@clerk/nextjs' to '@clerk/nextjs/server' for server-side
- Session: setSession() to setActive()
- Images: profileImageUrl/logoUrl to imageUrl

Migration Tool:

```bash
npx @clerk/upgrade --from=core-1
```

---

## Resources

Official Documentation:
- Quickstart: https://clerk.com/docs/quickstarts/nextjs
- SDK Reference: https://clerk.com/docs/reference/nextjs/overview
- Core 2 Migration: https://clerk.com/docs/guides/development/upgrading/upgrade-guides/core-2/nextjs
- Webhooks: https://clerk.com/docs/integrations/webhooks

Works Well With:
- moai-platform-auth0: Alternative enterprise SSO solution
- moai-platform-supabase: Supabase authentication integration
- moai-platform-vercel: Vercel deployment with Clerk
- moai-lang-typescript: TypeScript development patterns
- moai-domain-frontend: React and Next.js integration

---

Status: Production Ready
Version: 2.0.0
Last Updated: 2025-12-30
SDK Version: @clerk/nextjs 6.x (Core 2)
