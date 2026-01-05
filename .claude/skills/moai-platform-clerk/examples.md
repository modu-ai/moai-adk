# Clerk Authentication Examples

Production-ready code examples for Clerk authentication in Next.js 13+ App Router with React 18+ and TypeScript.

## Table of Contents

1. [Project Setup](#1-project-setup)
2. [ClerkProvider Configuration](#2-clerkprovider-configuration)
3. [Public and Protected Routes](#3-public-and-protected-routes)
4. [Sign-In/Sign-Up Pages](#4-sign-insign-up-pages)
5. [Client-Side Auth Hooks](#5-client-side-auth-hooks)
6. [User Profile Management](#6-user-profile-management)
7. [Organization Features](#7-organization-features)
8. [Server-Side Authentication](#8-server-side-authentication)
9. [Middleware Route Protection](#9-middleware-route-protection)
10. [Logout Implementation](#10-logout-implementation)

---

## 1. Project Setup

### 1.1 Initialize Next.js Project

```bash
npx create-next-app@latest my-clerk-app \
  --typescript \
  --tailwind \
  --eslint \
  --app \
  --src-dir \
  --import-alias "@/*"

cd my-clerk-app
```

### 1.2 Install Clerk SDK

```bash
npm install @clerk/nextjs
```

### 1.3 Configure Environment Variables

```bash
# .env.local
NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY=pk_test_xxxxxxxxxxxxxxxxxxxxxxxx
CLERK_SECRET_KEY=sk_test_xxxxxxxxxxxxxxxxxxxxxxxx
NEXT_PUBLIC_CLERK_SIGN_IN_URL=/sign-in
NEXT_PUBLIC_CLERK_SIGN_UP_URL=/sign-up
```

Get your keys from: https://dashboard.clerk.com/

### 1.4 TypeScript Types

Create `types/clerk.ts` for enhanced type safety:

```typescript
// types/clerk.ts
import { ClerkLoadingState } from '@clerk/nextjs/server'

declare global {
  interface Window {
    Clerk?: any
  }
}

export interface AuthUser {
  id: string
  firstName?: string | null
  lastName?: string | null
  fullName?: string | null
  emailAddress: string
  imageUrl: string
}

export interface OrganizationMember {
  id: string
  userId: string
  organizationId: string
  role: 'admin' | 'basic_member' | 'guest_member'
}

export {}
```

---

## 2. ClerkProvider Configuration

### 2.1 Root Layout with ClerkProvider

```tsx
// app/layout.tsx
import type { Metadata } from 'next'
import { ClerkProvider } from '@clerk/nextjs'
import { Inter } from 'next/font/google'
import './globals.css'

const inter = Inter({ subsets: ['latin'] })

export const metadata: Metadata = {
  title: 'My Clerk App',
  description: 'Authentication with Clerk',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <ClerkProvider>
      <html lang="en">
        <body className={inter.className}>{children}</body>
      </html>
    </ClerkProvider>
  )
}
```

### 2.2 ClerkProvider with Custom Appearance

```tsx
// app/layout.tsx
import { ClerkProvider } from '@clerk/nextjs'
import { dark } from '@clerk/themes'

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <ClerkProvider
      appearance={{
        baseTheme: dark,
        elements: {
          formButtonPrimary: 'bg-slate-900 hover:bg-slate-800 text-sm',
          card: 'shadow-none',
        },
      }}
    >
      <html lang="en">
        <body>{children}</body>
      </html>
    </ClerkProvider>
  )
}
```

### 2.3 Layout with Auth Components

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
import Link from 'next/link'

export const metadata: Metadata = {
  title: 'Clerk App',
  description: 'Modern authentication',
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <ClerkProvider>
      <html lang="en">
        <body>
          <nav className="flex items-center justify-between p-4 border-b">
            <Link href="/" className="text-xl font-bold">
              MyApp
            </Link>
            <div className="flex items-center gap-4">
              <SignedOut>
                <SignInButton mode="modal">
                  <button className="px-4 py-2 text-sm font-medium">
                    Sign In
                  </button>
                </SignInButton>
                <SignUpButton mode="modal">
                  <button className="px-4 py-2 text-sm font-medium bg-blue-600 text-white rounded">
                    Sign Up
                  </button>
                </SignUpButton>
              </SignedOut>
              <SignedIn>
                <UserButton
                  appearance={{
                    elements: {
                      avatarBox: 'w-10 h-10',
                    },
                  }}
                />
              </SignedIn>
            </div>
          </nav>
          <main className="container mx-auto py-8">{children}</main>
        </body>
      </html>
    </ClerkProvider>
  )
}
```

---

## 3. Public and Protected Routes

### 3.1 Public Home Page

```tsx
// app/page.tsx
import Link from 'next/link'

export default function HomePage() {
  return (
    <div className="text-center py-12">
      <h1 className="text-4xl font-bold mb-4">Welcome to MyApp</h1>
      <p className="text-gray-600 mb-8">
        The best way to manage your projects
      </p>
      <div className="flex gap-4 justify-center">
        <Link
          href="/sign-in"
          className="px-6 py-3 bg-blue-600 text-white rounded-lg font-medium"
        >
          Get Started
        </Link>
        <Link
          href="/dashboard"
          className="px-6 py-3 border border-gray-300 rounded-lg font-medium"
        >
          View Dashboard
        </Link>
      </div>
    </div>
  )
}
```

### 3.2 Protected Dashboard Page

```tsx
// app/dashboard/page.tsx
import { auth } from '@clerk/nextjs/server'
import { redirect } from 'next/navigation'

export default async function DashboardPage() {
  const { userId } = await auth()

  if (!userId) {
    redirect('/sign-in?redirectUrl=/dashboard')
  }

  return (
    <div>
      <h1 className="text-3xl font-bold mb-6">Dashboard</h1>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div className="p-6 bg-white rounded-lg shadow">
          <h2 className="text-xl font-semibold mb-2">Projects</h2>
          <p className="text-3xl font-bold">12</p>
        </div>
        <div className="p-6 bg-white rounded-lg shadow">
          <h2 className="text-xl font-semibold mb-2">Tasks</h2>
          <p className="text-3xl font-bold">48</p>
        </div>
        <div className="p-6 bg-white rounded-lg shadow">
          <h2 className="text-xl font-semibold mb-2">Team Members</h2>
          <p className="text-3xl font-bold">8</p>
        </div>
      </div>
    </div>
  )
}
```

### 3.3 Route Group for Protected Pages

```tsx
// app/(protected)/layout.tsx
import { auth } from '@clerk/nextjs/server'
import { redirect } from 'next/navigation'

export default async function ProtectedLayout({
  children,
}: {
  children: React.ReactNode
}) {
  const { userId } = await auth()

  if (!userId) {
    redirect('/sign-in')
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <aside className="w-64 bg-white shadow-lg p-6">
        <nav className="space-y-2">
          <a href="/dashboard" className="block p-2 rounded hover:bg-gray-100">
            Dashboard
          </a>
          <a href="/projects" className="block p-2 rounded hover:bg-gray-100">
            Projects
          </a>
          <a href="/settings" className="block p-2 rounded hover:bg-gray-100">
            Settings
          </a>
        </nav>
      </aside>
      <main>{children}</main>
    </div>
  )
}
```

---

## 4. Sign-In/Sign-Up Pages

### 4.1 Sign-In Page

```tsx
// app/sign-in/[[...sign-in]]/page.tsx
import { SignIn } from '@clerk/nextjs'

export default function SignInPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50">
      <SignIn
        appearance={{
          elements: {
            card: 'shadow-xl rounded-2xl',
          },
        }}
      />
    </div>
  )
}
```

### 4.2 Sign-Up Page

```tsx
// app/sign-up/[[...sign-up]]/page.tsx
import { SignUp } from '@clerk/nextjs'

export default function SignUpPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50">
      <SignUp
        appearance={{
          elements: {
            card: 'shadow-xl rounded-2xl',
          },
        }}
      />
    </div>
  )
}
```

### 4.3 Custom Sign-In Page with Redirect

```tsx
// app/sign-in/[[...sign-in]]/page.tsx
import { SignIn } from '@clerk/nextjs'

export default function SignInPage({
  searchParams,
}: {
  searchParams: { redirectUrl?: string }
}) {
  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50">
      <SignIn
        afterSignInUrl={searchParams.redirectUrl || '/dashboard'}
        afterSignUpUrl={searchParams.redirectUrl || '/dashboard'}
        appearance={{
          elements: {
            card: 'shadow-xl rounded-2xl',
          },
        }}
      />
    </div>
  )
}
```

### 4.4 Sign-In/Sign-Up Combined Flow

```tsx
// app/sign-in/[[...sign-in]]/page.tsx
import { SignIn } from '@clerk/nextjs'

export default function SignInPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50">
      <SignIn
        signUpUrl="/sign-up"
        forceRedirectUrl={false}
        appearance={{
          elements: {
            card: 'shadow-xl rounded-2xl',
            footer: 'hidden',
          },
        }}
      />
    </div>
  )
}
```

---

## 5. Client-Side Auth Hooks

### 5.1 useAuth Hook for Authentication State

```tsx
// app/components/AuthStatus.tsx
'use client'

import { useAuth } from '@clerk/nextjs'

export function AuthStatus() {
  const { isLoaded, isSignedIn, userId } = useAuth()

  if (!isLoaded) {
    return <div>Loading...</div>
  }

  if (!isSignedIn) {
    return <div>Not signed in</div>
  }

  return (
    <div>
      <p>Signed in as: {userId}</p>
    </div>
  )
}
```

### 5.2 Token Management with getToken

```tsx
// app/components/ExternalApi.tsx
'use client'

import { useAuth } from '@clerk/nextjs'
import { useState } from 'react'

export function ExternalApiButton() {
  const { getToken, userId } = useAuth()
  const [data, setData] = useState<any>(null)
  const [loading, setLoading] = useState(false)

  const fetchData = async () => {
    setLoading(true)
    try {
      const token = await getToken({ template: 'supabase' })
      const response = await fetch('https://api.example.com/user', {
        headers: {
          Authorization: `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      })
      const result = await response.json()
      setData(result)
    } catch (error) {
      console.error('Error fetching data:', error)
    } finally {
      setLoading(false)
    }
  }

  if (!userId) return null

  return (
    <div className="p-4 border rounded-lg">
      <button
        onClick={fetchData}
        disabled={loading}
        className="px-4 py-2 bg-blue-600 text-white rounded"
      >
        {loading ? 'Loading...' : 'Fetch External Data'}
      </button>
      {data && (
        <pre className="mt-4 p-4 bg-gray-100 rounded overflow-auto">
          {JSON.stringify(data, null, 2)}
        </pre>
      )}
    </div>
  )
}
```

### 5.3 Session Management with useSession

```tsx
// app/components/SessionInfo.tsx
'use client'

import { useSession } from '@clerk/nextjs'

export function SessionInfo() {
  const { isLoaded, isSignedIn, session } = useSession()

  if (!isLoaded || !isSignedIn) {
    return null
  }

  return (
    <div className="p-4 border rounded-lg">
      <h3 className="font-semibold mb-2">Session Information</h3>
      <p>Session ID: {session.id}</p>
      <p>Status: {session.status}</p>
      <p>Last Active: {session.lastActiveAt.toLocaleString()}</p>
      <p>Expire At: {session.expireAt.toLocaleString()}</p>
    </div>
  )
}
```

---

## 6. User Profile Management

### 6.1 useUser Hook for Profile Data

```tsx
// app/components/UserProfile.tsx
'use client'

import { useUser } from '@clerk/nextjs'

export function UserProfile() {
  const { isLoaded, isSignedIn, user } = useUser()

  if (!isLoaded) {
    return <div>Loading...</div>
  }

  if (!isSignedIn) {
    return <div>Please sign in</div>
  }

  return (
    <div className="p-6 bg-white rounded-lg shadow">
      <div className="flex items-center gap-4 mb-4">
        <img
          src={user.imageUrl}
          alt={user.fullName || 'User'}
          className="w-16 h-16 rounded-full"
        />
        <div>
          <h2 className="text-xl font-bold">
            {user.fullName || 'Anonymous User'}
          </h2>
          <p className="text-gray-600">
            {user.primaryEmailAddress?.emailAddress}
          </p>
        </div>
      </div>
      <div className="space-y-2 text-sm">
        <p>
          <strong>First Name:</strong> {user.firstName || 'Not set'}
        </p>
        <p>
          <strong>Last Name:</strong> {user.lastName || 'Not set'}
        </p>
        <p>
          <strong>Created:</strong>{' '}
          {user.createdAt.toLocaleDateString()}
        </p>
      </div>
    </div>
  )
}
```

### 6.2 Update User Profile

```tsx
// app/components/UpdateProfile.tsx
'use client'

import { useUser } from '@clerk/nextjs'
import { useState } from 'react'

export function UpdateProfile() {
  const { user } = useUser()
  const [firstName, setFirstName] = useState(user?.firstName || '')
  const [lastName, setLastName] = useState(user?.lastName || '')
  const [loading, setLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    try {
      await user?.update({
        firstName: firstName || undefined,
        lastName: lastName || undefined,
      })
      alert('Profile updated successfully!')
    } catch (error) {
      console.error('Error updating profile:', error)
      alert('Failed to update profile')
    } finally {
      setLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4 max-w-md">
      <div>
        <label className="block text-sm font-medium mb-1">
          First Name
        </label>
        <input
          type="text"
          value={firstName}
          onChange={(e) => setFirstName(e.target.value)}
          className="w-full px-3 py-2 border rounded"
          placeholder="Enter first name"
        />
      </div>
      <div>
        <label className="block text-sm font-medium mb-1">
          Last Name
        </label>
        <input
          type="text"
          value={lastName}
          onChange={(e) => setLastName(e.target.value)}
          className="w-full px-3 py-2 border rounded"
          placeholder="Enter last name"
        />
      </div>
      <button
        type="submit"
        disabled={loading}
        className="w-full px-4 py-2 bg-blue-600 text-white rounded disabled:bg-gray-400"
      >
        {loading ? 'Updating...' : 'Update Profile'}
      </button>
    </form>
  )
}
```

### 6.3 Server-Side User Data

```tsx
// app/profile/page.tsx
import { auth, currentUser } from '@clerk/nextjs/server'
import { redirect } from 'next/navigation'

export default async function ProfilePage() {
  const { userId } = await auth()

  if (!userId) {
    redirect('/sign-in')
  }

  const user = await currentUser()

  return (
    <div className="max-w-2xl mx-auto p-6">
      <h1 className="text-3xl font-bold mb-6">Your Profile</h1>
      <div className="bg-white rounded-lg shadow p-6 space-y-4">
        <div className="flex items-center gap-4">
          <img
            src={user?.imageUrl}
            alt={user?.fullName || 'User'}
            className="w-20 h-20 rounded-full"
          />
          <div>
            <h2 className="text-2xl font-bold">
              {user?.fullName || 'Anonymous User'}
            </h2>
            <p className="text-gray-600">
              {user?.primaryEmailAddress?.emailAddress}
            </p>
          </div>
        </div>
        <hr />
        <div className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <p className="font-medium">User ID</p>
            <p className="text-gray-600">{user?.id}</p>
          </div>
          <div>
            <p className="font-medium">First Name</p>
            <p className="text-gray-600">{user?.firstName || 'N/A'}</p>
          </div>
          <div>
            <p className="font-medium">Last Name</p>
            <p className="text-gray-600">{user?.lastName || 'N/A'}</p>
          </div>
          <div>
            <p className="font-medium">Joined</p>
            <p className="text-gray-600">
              {user?.createdAt.toLocaleDateString()}
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}
```

---

## 7. Organization Features

### 7.1 Organization Switcher Component

```tsx
// app/components/OrganizationSwitcher.tsx
'use client'

import { OrganizationSwitcher } from '@clerk/nextjs'

export function CustomOrgSwitcher() {
  return (
    <OrganizationSwitcher
      appearance={{
        elements: {
          organizationSwitcherTrigger: 'px-4 py-2 border rounded-lg',
          organizationSwitcherPopover: 'shadow-xl rounded-lg',
        },
      }}
      afterCreateOrganizationUrl="/organization"
      afterLeaveOrganizationUrl="/select-org"
    />
  )
}
```

### 7.2 useOrganization Hook for Current Organization

```tsx
// app/components/OrganizationInfo.tsx
'use client'

import { useOrganization } from '@clerk/nextjs'

export function OrganizationInfo() {
  const { organization, isLoaded } = useOrganization()

  if (!isLoaded) {
    return <div>Loading organization...</div>
  }

  if (!organization) {
    return <div>No organization selected</div>
  }

  return (
    <div className="p-6 bg-white rounded-lg shadow">
      <div className="flex items-center gap-4 mb-4">
        <img
          src={organization.imageUrl}
          alt={organization.name}
          className="w-16 h-16 rounded-lg"
        />
        <div>
          <h2 className="text-xl font-bold">{organization.name}</h2>
          <p className="text-gray-600">
            {organization.membersCount} members
          </p>
        </div>
      </div>
      <p className="text-sm">
        <strong>Created:</strong>{' '}
        {organization.createdAt.toLocaleDateString()}
      </p>
      <p className="text-sm">
        <strong>Admin:</strong>{' '}
        {organization.adminDisabled ? 'Disabled' : 'Enabled'}
      </p>
    </div>
  )
}
```

### 7.3 Custom Organization List

```tsx
// app/components/OrganizationList.tsx
'use client'

import { useOrganizationList } from '@clerk/nextjs'

export function CustomOrganizationList() {
  const { isLoaded, setActive, userMemberships } = useOrganizationList({
    userMemberships: {
      infinite: true,
      keepPreviousData: true,
    },
  })

  if (!isLoaded) {
    return <div>Loading...</div>
  }

  return (
    <div className="space-y-2">
      <h3 className="font-semibold mb-2">Your Organizations</h3>
      <ul className="space-y-2">
        {userMemberships.data?.map((membership) => (
          <li
            key={membership.id}
            className="p-3 border rounded-lg hover:bg-gray-50 cursor-pointer"
            onClick={() =>
              setActive({ organization: membership.organization.id })
            }
          >
            <div className="flex items-center gap-3">
              <img
                src={membership.organization.imageUrl}
                alt={membership.organization.name}
                className="w-10 h-10 rounded"
              />
              <div>
                <p className="font-medium">
                  {membership.organization.name}
                </p>
                <p className="text-sm text-gray-600">
                  {membership.role.replace(/_/g, ' ')}
                </p>
              </div>
            </div>
          </li>
        ))}
      </ul>
      {userMemberships.hasNextPage && (
        <button
          onClick={() => userMemberships.fetchNext()}
          className="w-full px-4 py-2 text-sm bg-gray-100 rounded hover:bg-gray-200"
        >
          Load more
        </button>
      )}
    </div>
  )
}
```

### 7.4 Organization Members Management

```tsx
// app/organization/[id]/members/page.tsx
import { auth } from '@clerk/nextjs/server'
import { notFound } from 'next/navigation'

export default async function OrganizationMembersPage({
  params,
}: {
  params: { id: string }
}) {
  const { userId, orgId } = await auth()

  if (!userId || !orgId) {
    return <div>Access denied</div>
  }

  try {
    const response = await fetch(
      `https://api.clerk.com/v1/organizations/${orgId}/memberships`,
      {
        headers: {
          Authorization: `Bearer ${process.env.CLERK_SECRET_KEY}`,
        },
      }
    )

    if (!response.ok) {
      notFound()
    }

    const memberships = await response.json()

    return (
      <div className="max-w-4xl mx-auto p-6">
        <h1 className="text-3xl font-bold mb-6">Organization Members</h1>
        <div className="bg-white rounded-lg shadow overflow-hidden">
          <table className="w-full">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-sm font-medium">
                  User
                </th>
                <th className="px-6 py-3 text-left text-sm font-medium">
                  Role
                </th>
                <th className="px-6 py-3 text-left text-sm font-medium">
                  Joined
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-200">
              {memberships.map((membership: any) => (
                <tr key={membership.id}>
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-3">
                      <img
                        src={membership.publicUserData.imageUrl}
                        alt={membership.publicUserData.firstName}
                        className="w-8 h-8 rounded-full"
                      />
                      <div>
                        <p className="font-medium">
                          {membership.publicUserData.firstName}{' '}
                          {membership.publicUserData.lastName}
                        </p>
                        <p className="text-sm text-gray-600">
                          {membership.publicUserData.identifier}
                        </p>
                      </div>
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <span className="px-2 py-1 text-sm bg-blue-100 text-blue-800 rounded">
                      {membership.role}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-sm text-gray-600">
                    {new Date(membership.createdAt).toLocaleDateString()}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    )
  } catch (error) {
    return <div>Error loading members</div>
  }
}
```

---

## 8. Server-Side Authentication

### 8.1 Server Component with auth()

```tsx
// app/dashboard/page.tsx
import { auth } from '@clerk/nextjs/server'
import { redirect } from 'next/navigation'

export default async function DashboardPage() {
  const { userId, orgId } = await auth()

  if (!userId) {
    redirect('/sign-in')
  }

  return (
    <div>
      <h1 className="text-2xl font-bold mb-4">Dashboard</h1>
      <p>Welcome! Your user ID: {userId}</p>
      {orgId && <p>Organization ID: {orgId}</p>}
    </div>
  )
}
```

### 8.2 Server Component with currentUser()

```tsx
// app/settings/page.tsx
import { currentUser } from '@clerk/nextjs/server'

export default async function SettingsPage() {
  const user = await currentUser()

  if (!user) {
    return <div>Please sign in</div>
  }

  return (
    <div className="max-w-2xl mx-auto p-6">
      <h1 className="text-3xl font-bold mb-6">Settings</h1>
      <div className="bg-white rounded-lg shadow p-6">
        <h2 className="text-xl font-semibold mb-4">Account Information</h2>
        <dl className="space-y-3">
          <div className="flex justify-between">
            <dt className="font-medium">Email</dt>
            <dd>{user.primaryEmailAddress?.emailAddress}</dd>
          </div>
          <div className="flex justify-between">
            <dt className="font-medium">Name</dt>
            <dd>{user.fullName || 'Not set'}</dd>
          </div>
          <div className="flex justify-between">
            <dt className="font-medium">User ID</dt>
            <dd className="font-mono text-sm">{user.id}</dd>
          </div>
        </dl>
      </div>
    </div>
  )
}
```

### 8.3 Protected Route Handler

```typescript
// app/api/protected/route.ts
import { auth } from '@clerk/nextjs/server'
import { NextResponse } from 'next/server'

export async function GET() {
  const { userId } = await auth()

  if (!userId) {
    return NextResponse.json(
      { error: 'Unauthorized' },
      { status: 401 }
    )
  }

  const data = {
    userId,
    message: 'This is protected data',
    timestamp: new Date().toISOString(),
  }

  return NextResponse.json(data)
}

export async function POST(request: Request) {
  const { userId } = await auth()

  if (!userId) {
    return NextResponse.json(
      { error: 'Unauthorized' },
      { status: 401 }
    )
  }

  const body = await request.json()

  // Process protected data
  const result = {
    userId,
    processedData: body,
    timestamp: new Date().toISOString(),
  }

  return NextResponse.json(result, { status: 201 })
}
```

### 8.4 Organization-Specific Route Handler

```typescript
// app/api/organization/data/route.ts
import { auth } from '@clerk/nextjs/server'
import { NextResponse } from 'next/server'

export async function GET() {
  const { userId, orgId } = await auth()

  if (!userId) {
    return NextResponse.json(
      { error: 'Unauthorized' },
      { status: 401 }
    )
  }

  if (!orgId) {
    return NextResponse.json(
      { error: 'No organization selected' },
      { status: 400 }
    )
  }

  // Fetch organization-specific data
  const orgData = {
    organizationId: orgId,
    userId,
    data: {
      // Your organization data here
    },
  }

  return NextResponse.json(orgData)
}
```

### 8.5 Token Generation in API Routes

```typescript
// app/api/token/route.ts
import { auth } from '@clerk/nextjs/server'
import { NextResponse } from 'next/server'

export async function GET() {
  const { userId } = await auth()

  if (!userId) {
    return NextResponse.json(
      { error: 'Unauthorized' },
      { status: 401 }
    )
  }

  // Generate JWT token for external services
  const token = await (
    await auth()
  ).getToken({ template: 'supabase' })

  return NextResponse.json({
    token,
    userId,
  })
}
```

---

## 9. Middleware Route Protection

### 9.1 Basic Middleware Configuration

```typescript
// middleware.ts
import { clerkMiddleware } from '@clerk/nextjs/server'

export default clerkMiddleware()

export const config = {
  matcher: [
    '/((?!_next|[^?]*\\.(?:html?|css|js(?!on)|jpe?g|webp|png|gif|svg|ttf|woff2?|ico|csv|docx?|xlsx?|zip|webmanifest)).*)',
    '/(api|trpc)(.*)',
  ],
}
```

### 9.2 Protect Specific Routes

```typescript
// middleware.ts
import { clerkMiddleware, createRouteMatcher } from '@clerk/nextjs/server'

// Define protected routes
const isProtectedRoute = createRouteMatcher([
  '/dashboard(.*)',
  '/projects(.*)',
  '/settings(.*)',
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

### 9.3 Define Public Routes Only

```typescript
// middleware.ts
import { clerkMiddleware, createRouteMatcher } from '@clerk/nextjs/server'

// Define public routes (all others protected)
const isPublicRoute = createRouteMatcher([
  '/sign-in(.*)',
  '/sign-up(.*)',
  '/api/webhooks(.*)',
  '/',
  '/about(.*)',
  '/pricing(.*)',
])

export default clerkMiddleware(async (auth, req) => {
  if (!isPublicRoute(req)) {
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

### 9.4 Role-Based Access Control

```typescript
// middleware.ts
import { clerkMiddleware, createRouteMatcher } from '@clerk/nextjs/server'

const isProtectedRoute = createRouteMatcher([
  '/dashboard(.*)',
  '/admin(.*)',
])

export default clerkMiddleware(async (auth, req) => {
  if (isProtectedRoute(req)) {
    const { has } = await auth()

    // Check for admin role
    if (req.nextUrl.pathname.startsWith('/admin')) {
      if (!(await has({ role: 'org:admin' }))) {
        return new Response('Unauthorized', { status: 403 })
      }
    }

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

### 9.5 Organization-Based Protection

```typescript
// middleware.ts
import { clerkMiddleware, createRouteMatcher } from '@clerk/nextjs/server'

const isOrgRoute = createRouteMatcher(['/organization(.*)', '/org(.*)'])

export default clerkMiddleware(async (auth, req) => {
  if (isOrgRoute(req)) {
    const { orgId } = await auth()

    if (!orgId) {
      // Redirect to org selection if no org selected
      const url = req.nextUrl.clone()
      url.pathname = '/select-org'
      return Response.redirect(url)
    }
  }
})

export const config = {
  matcher: [
    '/((?!_next|[^?]*\\.(?:html?|css|js(?!on)|jpe?g|webp|png|gif|svg|ttf|woff2?|ico|csv|docx?|xlsx?|zip|webmanifest)).*)',
    '/(api|trpc)(.*)',
  ],
}
```

---

## 10. Logout Implementation

### 10.1 SignOutButton Component

```tsx
// app/components/SignOutButton.tsx
'use client'

import { SignOutButton } from '@clerk/nextjs'

export function LogoutButton() {
  return (
    <SignOutButton signOutUrl="/">
      <button className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700">
        Sign Out
      </button>
    </SignOutButton>
  )
}
```

### 10.2 SignOut with Redirect

```tsx
// app/components/SignOutButton.tsx
'use client'

import { SignOutButton } from '@clerk/nextjs'
import { useRouter } from 'next/navigation'

export function SignOutWithRedirect() {
  const router = useRouter()

  return (
    <SignOutButton
      signOutCallback={() => {
        router.push('/goodbye')
      }}
    >
      <button className="px-4 py-2 bg-red-600 text-white rounded">
        Sign Out
      </button>
    </SignOutButton>
  )
}
```

### 10.3 Custom SignOut Handler

```tsx
// app/components/CustomSignOut.tsx
'use client'

import { useClerk } from '@clerk/nextjs'

export function CustomSignOutButton() {
  const { signOut } = useClerk()

  const handleSignOut = async () => {
    // Perform cleanup operations
    await fetch('/api/cleanup', { method: 'POST' })

    // Sign out
    signOut(() => {
      // Callback after sign out
      window.location.href = '/'
    })
  }

  return (
    <button
      onClick={handleSignOut}
      className="px-4 py-2 bg-red-600 text-white rounded"
    >
      Sign Out
    </button>
  )
}
```

### 10.4 SignOut in User Menu

```tsx
// app/components/UserMenu.tsx
'use client'

import {
  UserButton,
  SignOutButton,
  useAuth,
} from '@clerk/nextjs'

export function UserMenu() {
  const { userId } = useAuth()

  if (!userId) return null

  return (
    <UserButton
      appearance={{
        elements: {
          userButtonOuterIdentifier: 'text-sm font-medium',
        },
      }}
    >
      <UserButton.MenuItems>
        <UserButton.Link
          label="Profile"
          href="/profile"
          labelIcon={<ProfileIcon />}
        />
        <UserButton.Link
          label="Settings"
          href="/settings"
          labelIcon={<SettingsIcon />}
        />
        <UserButton.Action label="Sign Out" />
      </UserButton.MenuItems>
    </UserButton>
  )
}

// Helper components for icons
function ProfileIcon() {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      className="h-4 w-4"
      fill="none"
      viewBox="0 0 24 24"
      stroke="currentColor"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
      />
    </svg>
  )
}

function SettingsIcon() {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      className="h-4 w-4"
      fill="none"
      viewBox="0 0 24 24"
      stroke="currentColor"
    >
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"
      />
      <path
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth={2}
        d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
      />
    </svg>
  )
}
```

### 10.5 Server-Side SignOut

```tsx
// app/sign-out/page.tsx
import { redirect } from 'next/navigation'

export default function SignOutPage() {
  // Server-side redirect to Clerk's sign-out
  redirect('/sign-out')
}
```

---

## Additional Resources

### Testing Authentication

```typescript
// tests/auth.test.tsx
import { render, screen } from '@testing-library/react'
import { ClerkProvider, useAuth } from '@clerk/nextjs'

// Mock auth state
function TestComponent() {
  const { isSignedIn, userId } = useAuth()
  return <div>{isSignedIn ? `Signed in as ${userId}` : 'Not signed in'}</div>
}

test('displays user info when signed in', () => {
  render(
    <ClerkProvider>
      <TestComponent />
    </ClerkProvider>
  )
  // Add your test assertions
})
```

### Error Handling

```tsx
// app/components/ErrorBoundary.tsx
'use client'

import { useAuth } from '@clerk/nextjs'
import { useEffect } from 'react'

export function AuthErrorBoundary({ children }: { children: React.ReactNode }) {
  const { isLoaded } = useAuth()

  useEffect(() => {
    if (!isLoaded) {
      console.warn('Auth not loaded')
    }
  }, [isLoaded])

  if (!isLoaded) {
    return <div>Loading authentication...</div>
  }

  return <>{children}</>
}
```

### Performance Optimization

```typescript
// next.config.js
const nextConfig = {
  // Inline Clerk scripts for better performance
  experimental: {
    // Optimize Clerk's JavaScript bundles
  },
}

module.exports = nextConfig
```

---

## Version Compatibility

- **Next.js**: 13.0.4+ (App Router)
- **React**: 18+
- **Node.js**: 18.17.0+
- **@clerk/nextjs**: 6.x (Core 2)
- **TypeScript**: 5.x

## Official Documentation

- Clerk Docs: https://clerk.com/docs
- Next.js SDK: https://clerk.com/docs/sdk/nextjs
- Core 2 Migration: https://clerk.com/docs/guides/development/upgrading/upgrade-guides/core-2/nextjs
- Components Reference: https://clerk.com/docs/components/overview

---

Last Updated: 2025-12-30
Examples Version: 1.0.0
