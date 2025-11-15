# Clerk Modern User Management 완전 가이드

## 개요

Clerk는 현대적인 사용자 인증 및 관리 플랫폼입니다. 완전한 UI 컴포넌트, 세션 관리, 다중 인증 제공자, 그리고 엔터프라이즈급 보안을 5분 안에 통합할 수 있습니다.

**핵심 장점**:
- **Drop-in Components**: 즉시 사용 가능한 React 컴포넌트
- **Multi-tenant Ready**: Organizations & RBAC 내장
- **Passwordless Options**: 이메일 코드, SMS, Social Login
- **Modern Security**: WebAuthn, MFA, Session Management
- **Developer Experience**: Next.js, React, Node.js 완벽 통합

## 왜 Clerk인가?

### 1. 5분 안에 인증 시스템 구축 (Pattern H)

```typescript
// app/layout.tsx
import { ClerkProvider } from '@clerk/nextjs'

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <ClerkProvider>
      <html lang="ko">
        <body>{children}</body>
      </html>
    </ClerkProvider>
  )
}

// app/sign-in/[[...sign-in]]/page.tsx
import { SignIn } from '@clerk/nextjs'

export default function SignInPage() {
  return (
    <div className="flex h-screen items-center justify-center">
      <SignIn />
    </div>
  )
}

// 완료! 이제 다음 기능이 모두 작동합니다:
// - 이메일/비밀번호 로그인
// - 소셜 로그인 (Google, GitHub, etc.)
// - 비밀번호 재설정
// - 이메일 인증
// - 세션 관리
```

### 2. Multi-tenant SaaS 완벽 지원

```typescript
// app/dashboard/page.tsx
import { auth, currentUser } from '@clerk/nextjs'
import { redirect } from 'next/navigation'

export default async function DashboardPage() {
  const { userId, orgId, orgRole } = auth()

  if (!userId) redirect('/sign-in')

  const user = await currentUser()

  return (
    <div>
      <h1>Welcome, {user?.firstName}!</h1>
      {orgId && (
        <div>
          <p>Organization: {orgId}</p>
          <p>Your Role: {orgRole}</p>
        </div>
      )}
    </div>
  )
}
```

## 주요 기능

### 1. Authentication Methods

**소셜 로그인**:
```typescript
// app/sign-up/[[...sign-up]]/page.tsx
import { SignUp } from '@clerk/nextjs'

export default function SignUpPage() {
  return (
    <SignUp
      appearance={{
        elements: {
          socialButtonsBlockButton: 'custom-social-button',
        },
      }}
      redirectUrl="/onboarding"
    />
  )
}
```

**Clerk Dashboard에서 설정**:
- Google OAuth
- GitHub OAuth
- Microsoft OAuth
- Facebook, Twitter, LinkedIn, etc.

**Passwordless (이메일 코드)**:
```typescript
import { SignIn } from '@clerk/nextjs'

export default function SignInPage() {
  return (
    <SignIn
      appearance={{
        variables: {
          colorPrimary: '#0F172A',
        },
      }}
      // 이메일 코드 활성화 (Dashboard에서 설정)
    />
  )
}
```

### 2. Session Management

```typescript
// middleware.ts - Protected Routes
import { authMiddleware } from '@clerk/nextjs'

export default authMiddleware({
  publicRoutes: ['/', '/pricing', '/about'],
  ignoredRoutes: ['/api/webhook'],
})

export const config = {
  matcher: ['/((?!.+\\.[\\w]+$|_next).*)', '/', '/(api|trpc)(.*)'],
}

// app/api/protected/route.ts
import { auth } from '@clerk/nextjs'
import { NextResponse } from 'next/server'

export async function GET() {
  const { userId } = auth()

  if (!userId) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  }

  // 보호된 데이터 반환
  return NextResponse.json({ data: 'Secret data' })
}
```

### 3. Organizations (Multi-tenancy)

```typescript
// app/org/[orgId]/page.tsx
import { auth, clerkClient } from '@clerk/nextjs'

export default async function OrganizationPage({
  params,
}: {
  params: { orgId: string }
}) {
  const { userId, orgRole } = auth()

  // 조직 정보 가져오기
  const organization = await clerkClient.organizations.getOrganization({
    organizationId: params.orgId,
  })

  // 멤버 목록
  const members = await clerkClient.organizations.getOrganizationMembershipList({
    organizationId: params.orgId,
  })

  return (
    <div>
      <h1>{organization.name}</h1>
      <p>Your Role: {orgRole}</p>

      <h2>Members</h2>
      <ul>
        {members.map((member) => (
          <li key={member.id}>
            {member.publicUserData.firstName} - {member.role}
          </li>
        ))}
      </ul>
    </div>
  )
}
```

**조직 스위처 컴포넌트**:
```typescript
// components/org-switcher.tsx
'use client'

import { OrganizationSwitcher } from '@clerk/nextjs'

export function OrgSwitcher() {
  return (
    <OrganizationSwitcher
      hidePersonal={false}
      afterCreateOrganizationUrl="/org/:id"
      afterSelectOrganizationUrl="/org/:id"
      appearance={{
        elements: {
          rootBox: 'custom-org-switcher',
        },
      }}
    />
  )
}
```

### 4. Role-Based Access Control (RBAC)

```typescript
// lib/rbac.ts
import { auth } from '@clerk/nextjs'

export function hasPermission(permission: string): boolean {
  const { orgRole } = auth()

  const permissions = {
    'org:admin': ['read', 'write', 'delete', 'invite', 'manage'],
    'org:member': ['read', 'write'],
    'org:guest': ['read'],
  }

  return permissions[orgRole as keyof typeof permissions]?.includes(permission) || false
}

// app/api/posts/[id]/route.ts
import { hasPermission } from '@/lib/rbac'

export async function DELETE(
  request: Request,
  { params }: { params: { id: string } }
) {
  if (!hasPermission('delete')) {
    return Response.json({ error: 'Forbidden' }, { status: 403 })
  }

  // 삭제 로직
  return Response.json({ success: true })
}
```

### 5. User Metadata

```typescript
// lib/user-metadata.ts
import { auth, clerkClient } from '@clerk/nextjs'

export async function updateUserMetadata(data: any) {
  const { userId } = auth()

  if (!userId) throw new Error('Unauthorized')

  await clerkClient.users.updateUser(userId, {
    publicMetadata: {
      // 공개 메타데이터 (다른 사용자에게 보임)
      displayName: data.displayName,
      avatar: data.avatar,
    },
    privateMetadata: {
      // 비공개 메타데이터 (본인만 보임)
      preferences: data.preferences,
    },
    unsafeMetadata: {
      // 클라이언트에서 수정 가능
      theme: data.theme,
    },
  })
}

// 사용 예제
await updateUserMetadata({
  displayName: 'John Doe',
  avatar: 'https://...',
  preferences: { notifications: true },
  theme: 'dark',
})
```

### 6. Webhooks

```typescript
// app/api/webhooks/clerk/route.ts
import { Webhook } from 'svix'
import { headers } from 'next/headers'
import { WebhookEvent } from '@clerk/nextjs/server'

export async function POST(req: Request) {
  const WEBHOOK_SECRET = process.env.CLERK_WEBHOOK_SECRET

  if (!WEBHOOK_SECRET) {
    throw new Error('Please add CLERK_WEBHOOK_SECRET to .env')
  }

  // 헤더 가져오기
  const headerPayload = headers()
  const svix_id = headerPayload.get('svix-id')
  const svix_timestamp = headerPayload.get('svix-timestamp')
  const svix_signature = headerPayload.get('svix-signature')

  if (!svix_id || !svix_timestamp || !svix_signature) {
    return Response.json({ error: 'Missing headers' }, { status: 400 })
  }

  // 페이로드 가져오기
  const payload = await req.json()
  const body = JSON.stringify(payload)

  // 서명 검증
  const wh = new Webhook(WEBHOOK_SECRET)

  let evt: WebhookEvent

  try {
    evt = wh.verify(body, {
      'svix-id': svix_id,
      'svix-timestamp': svix_timestamp,
      'svix-signature': svix_signature,
    }) as WebhookEvent
  } catch (err) {
    console.error('Error verifying webhook:', err)
    return Response.json({ error: 'Invalid signature' }, { status: 400 })
  }

  // 이벤트 처리
  switch (evt.type) {
    case 'user.created':
      await handleUserCreated(evt.data)
      break
    case 'user.updated':
      await handleUserUpdated(evt.data)
      break
    case 'organization.created':
      await handleOrganizationCreated(evt.data)
      break
    default:
      console.log(`Unhandled event type: ${evt.type}`)
  }

  return Response.json({ received: true })
}

async function handleUserCreated(data: any) {
  // 사용자 생성 시 로직 (예: 데이터베이스에 추가)
  console.log('User created:', data.id)
}
```

### 7. Multi-Factor Authentication (MFA)

```typescript
// app/security/page.tsx
'use client'

import { useUser } from '@clerk/nextjs'

export default function SecurityPage() {
  const { user } = useUser()

  const enableMFA = async () => {
    // Clerk가 자동으로 MFA 설정 플로우 제공
    await user?.createTOTP()
  }

  return (
    <div>
      <h1>Security Settings</h1>

      <button onClick={enableMFA} className="btn">
        Enable Two-Factor Authentication
      </button>

      {user?.twoFactorEnabled && (
        <p className="text-green-600">✓ 2FA is enabled</p>
      )}
    </div>
  )
}
```

## 시작하기

### 1. 설치 및 설정

```bash
# Clerk 패키지 설치
npm install @clerk/nextjs

# 환경 변수 설정
echo "NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY=pk_test_..." >> .env.local
echo "CLERK_SECRET_KEY=sk_test_..." >> .env.local
```

### 2. Provider 설정

```typescript
// app/layout.tsx
import { ClerkProvider } from '@clerk/nextjs'
import './globals.css'

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <ClerkProvider>
      <html lang="ko">
        <body>{children}</body>
      </html>
    </ClerkProvider>
  )
}
```

### 3. 인증 페이지 생성

```typescript
// app/sign-in/[[...sign-in]]/page.tsx
import { SignIn } from '@clerk/nextjs'

export default function Page() {
  return <SignIn />
}

// app/sign-up/[[...sign-up]]/page.tsx
import { SignUp } from '@clerk/nextjs'

export default function Page() {
  return <SignUp />
}
```

### 4. Middleware 설정

```typescript
// middleware.ts
import { authMiddleware } from '@clerk/nextjs'

export default authMiddleware({
  publicRoutes: ['/', '/api/public'],
})

export const config = {
  matcher: ['/((?!.+\\.[\\w]+$|_next).*)', '/', '/(api|trpc)(.*)'],
}
```

## 사용 가이드

### Pattern H: Enterprise Security

```typescript
// app/admin/page.tsx
import { auth } from '@clerk/nextjs'
import { redirect } from 'next/navigation'

export default async function AdminPage() {
  const { userId, sessionClaims } = auth()

  // 관리자 권한 확인
  const isAdmin = sessionClaims?.metadata?.role === 'admin'

  if (!isAdmin) {
    redirect('/unauthorized')
  }

  return <div>Admin Dashboard</div>
}

// lib/clerk-roles.ts
export const ROLES = {
  ADMIN: 'admin',
  MODERATOR: 'moderator',
  USER: 'user',
} as const

export function hasRole(role: string): boolean {
  const { sessionClaims } = auth()
  return sessionClaims?.metadata?.role === role
}

export function requireRole(role: string) {
  if (!hasRole(role)) {
    throw new Error('Insufficient permissions')
  }
}
```

## 코드 예제

### 1. 커스텀 Sign Up Flow

```typescript
// app/sign-up/custom/page.tsx
'use client'

import { useSignUp } from '@clerk/nextjs'
import { useState } from 'react'

export default function CustomSignUp() {
  const { signUp, setActive } = useSignUp()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [code, setCode] = useState('')
  const [verifying, setVerifying] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()

    try {
      // 1. 회원가입 시작
      await signUp?.create({
        emailAddress: email,
        password,
      })

      // 2. 이메일 인증 코드 전송
      await signUp?.prepareEmailAddressVerification({
        strategy: 'email_code',
      })

      setVerifying(true)
    } catch (err) {
      console.error('Error:', err)
    }
  }

  const handleVerify = async (e: React.FormEvent) => {
    e.preventDefault()

    try {
      // 3. 인증 코드 확인
      const result = await signUp?.attemptEmailAddressVerification({
        code,
      })

      if (result?.status === 'complete') {
        // 4. 세션 활성화
        await setActive({ session: result.createdSessionId })
        // 성공! 리다이렉트
      }
    } catch (err) {
      console.error('Error:', err)
    }
  }

  if (verifying) {
    return (
      <form onSubmit={handleVerify}>
        <input
          type="text"
          value={code}
          onChange={(e) => setCode(e.target.value)}
          placeholder="Enter verification code"
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

### 2. Server-Side User Management

```typescript
// app/api/users/route.ts
import { clerkClient } from '@clerk/nextjs'
import { NextResponse } from 'next/server'

export async function GET() {
  // 모든 사용자 가져오기
  const users = await clerkClient.users.getUserList({
    limit: 100,
    offset: 0,
  })

  return NextResponse.json(users)
}

export async function POST(request: Request) {
  const { email, firstName, lastName } = await request.json()

  // 서버에서 사용자 생성
  const user = await clerkClient.users.createUser({
    emailAddress: [email],
    firstName,
    lastName,
    password: 'temporaryPassword123!',
    publicMetadata: {
      source: 'admin-panel',
    },
  })

  return NextResponse.json(user)
}

// app/api/users/[userId]/route.ts
export async function DELETE(
  request: Request,
  { params }: { params: { userId: string } }
) {
  await clerkClient.users.deleteUser(params.userId)
  return NextResponse.json({ success: true })
}
```

### 3. Organization Invitations

```typescript
// app/api/org/invite/route.ts
import { auth, clerkClient } from '@clerk/nextjs'
import { NextResponse } from 'next/server'

export async function POST(request: Request) {
  const { orgId } = auth()

  if (!orgId) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  }

  const { email, role } = await request.json()

  // 초대 생성
  const invitation = await clerkClient.organizations.createOrganizationInvitation({
    organizationId: orgId,
    emailAddress: email,
    role: role || 'org:member',
  })

  return NextResponse.json(invitation)
}
```

### 4. Session Claims

```typescript
// app/api/auth/session/route.ts
import { auth } from '@clerk/nextjs'
import { NextResponse } from 'next/server'

export async function GET() {
  const { userId, sessionClaims, sessionId } = auth()

  return NextResponse.json({
    userId,
    sessionId,
    claims: sessionClaims,
    metadata: sessionClaims?.metadata,
  })
}
```

## Best Practices

### 1. Protect API Routes

```typescript
// middleware.ts - API 라우트 보호
import { authMiddleware } from '@clerk/nextjs'

export default authMiddleware({
  publicRoutes: ['/api/public'],
  ignoredRoutes: ['/api/webhook'],
})
```

### 2. Type-Safe User Data

```typescript
// types/clerk.ts
declare global {
  interface CustomJwtSessionClaims {
    metadata: {
      role?: 'admin' | 'moderator' | 'user'
      organizationId?: string
    }
  }
}

export {}
```

### 3. Error Handling

```typescript
// lib/auth-errors.ts
import { ClerkAPIError } from '@clerk/types'

export function handleClerkError(error: unknown) {
  if (error instanceof ClerkAPIError) {
    return {
      message: error.message,
      code: error.code,
      meta: error.meta,
    }
  }

  return {
    message: 'An unexpected error occurred',
  }
}
```

### 4. Development vs Production

```typescript
// lib/config.ts
export const clerkConfig = {
  signInUrl: process.env.NEXT_PUBLIC_CLERK_SIGN_IN_URL || '/sign-in',
  signUpUrl: process.env.NEXT_PUBLIC_CLERK_SIGN_UP_URL || '/sign-up',
  afterSignInUrl: process.env.NEXT_PUBLIC_CLERK_AFTER_SIGN_IN_URL || '/dashboard',
  afterSignUpUrl: process.env.NEXT_PUBLIC_CLERK_AFTER_SIGN_UP_URL || '/onboarding',
}
```

## 문제 해결

### 1. Hydration Mismatch

```typescript
// components/auth-guard.tsx
'use client'

import { useUser } from '@clerk/nextjs'
import { useEffect, useState } from 'react'

export function AuthGuard({ children }: { children: React.ReactNode }) {
  const { isLoaded } = useUser()
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    setMounted(true)
  }, [])

  if (!mounted || !isLoaded) {
    return <div>Loading...</div>
  }

  return <>{children}</>
}
```

### 2. Session Refresh

```typescript
// lib/session.ts
import { auth } from '@clerk/nextjs'

export async function refreshSession() {
  const { getToken } = auth()

  // 새 토큰 가져오기
  const token = await getToken()

  return token
}
```

## 다음 단계

- [Auth0 가이드](/ko/skills/baas/auth0) - Enterprise 대안
- [Vercel 가이드](/ko/skills/baas/vercel) - 배포 플랫폼
- [Pattern H: Enterprise Security](/ko/skills/patterns/pattern-h) - 보안 아키텍처
- [BaaS 개요](/ko/skills/baas) - 플랫폼 비교

## 참고 자료

- [Clerk 공식 문서](https://clerk.com/docs)
- [Next.js 통합 가이드](https://clerk.com/docs/quickstarts/nextjs)
- [Organizations 가이드](https://clerk.com/docs/organizations/overview)
- [Clerk Community](https://discord.com/invite/clerk)
