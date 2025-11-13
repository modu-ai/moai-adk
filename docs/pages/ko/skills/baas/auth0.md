# Auth0 Enterprise Authentication 완전 가이드

## 개요

Auth0는 Okta의 엔터프라이즈급 Identity 플랫폼입니다. SAML, OIDC, JWT 표준을 완벽하게 지원하며, 글로벌 기업들이 신뢰하는 보안과 규정 준수 기능을 제공합니다.

**핵심 장점**:
- **Universal Login**: 중앙화된 인증 페이지
- **Enterprise Connections**: SAML, AD/LDAP, Azure AD
- **Advanced Security**: Anomaly Detection, Breached Password Detection
- **Compliance Ready**: SOC2, GDPR, HIPAA, ISO 27001
- **Flexible Integration**: 모든 언어/프레임워크 지원

## 왜 Auth0인가?

### 1. Enterprise 요구사항 완벽 지원 (Pattern H)

```typescript
// lib/auth0.ts
import { Auth0Provider } from '@auth0/nextjs-auth0'

export const auth0Config = {
  domain: process.env.AUTH0_DOMAIN!,
  clientId: process.env.AUTH0_CLIENT_ID!,
  clientSecret: process.env.AUTH0_CLIENT_SECRET!,
  baseURL: process.env.AUTH0_BASE_URL!,
  issuerBaseURL: `https://${process.env.AUTH0_DOMAIN}`,
  secret: process.env.AUTH0_SECRET!,
  routes: {
    callback: '/api/auth/callback',
    postLogoutRedirect: '/',
  },
}

// app/api/auth/[auth0]/route.ts
import { handleAuth, handleLogin } from '@auth0/nextjs-auth0'

export const GET = handleAuth({
  login: handleLogin({
    returnTo: '/dashboard',
    authorizationParams: {
      // OIDC scopes
      scope: 'openid profile email offline_access',
      // Custom parameters
      audience: process.env.AUTH0_AUDIENCE,
      // Enterprise connection
      connection: 'saml-enterprise',
    },
  }),
})
```

### 2. SAML/OIDC 엔터프라이즈 SSO

```typescript
// middleware.ts - Enterprise SSO
import { withMiddlewareAuthRequired } from '@auth0/nextjs-auth0/edge'

export default withMiddlewareAuthRequired({
  returnTo: '/dashboard',
})

export const config = {
  matcher: ['/dashboard/:path*', '/api/protected/:path*'],
}

// app/dashboard/page.tsx
import { getSession } from '@auth0/nextjs-auth0'

export default async function Dashboard() {
  const session = await getSession()

  // SAML Assertion에서 가져온 사용자 정보
  const user = session?.user

  return (
    <div>
      <h1>Welcome, {user?.name}</h1>
      <p>Email: {user?.email}</p>
      <p>Organization: {user?.org_id}</p>
      <p>Roles: {user?.['https://myapp.com/roles']?.join(', ')}</p>
    </div>
  )
}
```

## 주요 기능

### 1. Universal Login

**중앙화된 인증 페이지**:

```typescript
// app/api/auth/[auth0]/route.ts
import { handleAuth, handleLogin } from '@auth0/nextjs-auth0'

export const GET = handleAuth({
  login: handleLogin({
    authorizationParams: {
      // Universal Login 페이지 커스터마이징
      screen_hint: 'signup', // 회원가입 화면으로 시작
      prompt: 'login', // 재인증 강제
      ui_locales: 'ko', // 한국어 UI
      // Custom parameters
      organization: 'org_abc123', // 특정 조직으로 로그인
    },
  }),
  signup: handleLogin({
    authorizationParams: {
      screen_hint: 'signup',
    },
    returnTo: '/onboarding',
  }),
})
```

**New Universal Login 커스터마이징** (Auth0 Dashboard):
```html
<!DOCTYPE html>
<html>
  <head>
    <title>Login - My App</title>
    {%- auth0:head -%}
  </head>
  <body>
    <div id="custom-login">
      <!-- 커스텀 로고, 브랜딩 -->
      <img src="https://myapp.com/logo.png" alt="Logo" />

      {%- auth0:widget -%}

      <!-- 커스텀 푸터 -->
      <footer>
        <a href="https://myapp.com/privacy">Privacy Policy</a>
        <a href="https://myapp.com/terms">Terms of Service</a>
      </footer>
    </div>
  </body>
</html>
```

### 2. Actions (Rules/Hooks)

**사용자 정의 인증 플로우**:

```javascript
// Auth0 Action: Add User Roles to Token
exports.onExecutePostLogin = async (event, api) => {
  const namespace = 'https://myapp.com'

  // 데이터베이스에서 역할 가져오기
  const roles = await fetchUserRoles(event.user.user_id)

  // ID Token에 역할 추가
  api.idToken.setCustomClaim(`${namespace}/roles`, roles)

  // Access Token에 역할 추가
  api.accessToken.setCustomClaim(`${namespace}/roles`, roles)

  // 사용자 메타데이터 업데이트
  api.user.setUserMetadata('last_login', new Date())
}

// Auth0 Action: Block Suspicious Logins
exports.onExecutePostLogin = async (event, api) => {
  // 비정상적인 위치에서 로그인 시도
  const userCountry = event.user.app_metadata?.usual_country
  const loginCountry = event.request.geoip.country_code

  if (userCountry && userCountry !== loginCountry) {
    // MFA 강제 또는 로그인 차단
    api.multifactor.enable('any', { allowRememberBrowser: false })

    // 알림 전송
    await sendSecurityAlert(event.user.email, {
      location: loginCountry,
      ip: event.request.ip,
    })
  }
}

// Auth0 Action: Progressive Profiling
exports.onExecutePostLogin = async (event, api) => {
  const { user } = event

  // 첫 로그인 시 추가 정보 요청
  if (user.logins_count === 1) {
    api.redirect.sendUserTo('https://myapp.com/complete-profile', {
      query: {
        user_id: user.user_id,
        token: api.redirect.encodeToken({ expiresInSeconds: 300 }),
      },
    })
  }
}
```

### 3. Organizations (B2B SaaS)

```typescript
// lib/auth0-orgs.ts
import { ManagementClient } from 'auth0'

const management = new ManagementClient({
  domain: process.env.AUTH0_DOMAIN!,
  clientId: process.env.AUTH0_CLIENT_ID!,
  clientSecret: process.env.AUTH0_CLIENT_SECRET!,
})

export async function createOrganization(name: string, displayName: string) {
  return await management.organizations.create({
    name,
    display_name: displayName,
    branding: {
      logo_url: 'https://myapp.com/org-logo.png',
      colors: {
        primary: '#0F172A',
        page_background: '#FFFFFF',
      },
    },
  })
}

export async function addMemberToOrganization(orgId: string, userId: string, roles: string[]) {
  await management.organizations.addMembers(
    { id: orgId },
    { members: [userId] }
  )

  await management.organizations.addMemberRoles(
    { id: orgId, user_id: userId },
    { roles }
  )
}

export async function getOrganizationMembers(orgId: string) {
  return await management.organizations.getMembers({ id: orgId })
}

// app/api/orgs/route.ts
import { createOrganization } from '@/lib/auth0-orgs'

export async function POST(request: Request) {
  const { name, displayName } = await request.json()

  const org = await createOrganization(name, displayName)

  return Response.json(org)
}
```

**조직별 로그인**:
```typescript
// app/login/[organization]/page.tsx
export default function OrganizationLogin({
  params,
}: {
  params: { organization: string }
}) {
  return (
    <div>
      <h1>Sign in to {params.organization}</h1>
      <a
        href={`/api/auth/login?organization=${params.organization}`}
        className="btn"
      >
        Sign In
      </a>
    </div>
  )
}
```

### 4. Multi-Factor Authentication

```typescript
// app/api/auth/[auth0]/route.ts
import { handleAuth, handleLogin } from '@auth0/nextjs-auth0'

export const GET = handleAuth({
  login: handleLogin({
    authorizationParams: {
      // MFA 강제
      acr_values: 'http://schemas.openid.net/pape/policies/2007/06/multi-factor',
    },
  }),
})

// Auth0 Action: Adaptive MFA
exports.onExecutePostLogin = async (event, api) => {
  const riskScore = await calculateRiskScore(event)

  if (riskScore > 0.7) {
    // 높은 위험도 - MFA 강제
    api.multifactor.enable('any', { allowRememberBrowser: false })
  } else if (riskScore > 0.4) {
    // 중간 위험도 - MFA 권장
    api.multifactor.enable('any', { allowRememberBrowser: true })
  }
}

async function calculateRiskScore(event) {
  let score = 0

  // 새로운 기기
  if (!event.user.app_metadata?.known_devices?.includes(event.request.fingerprint)) {
    score += 0.3
  }

  // 새로운 IP
  const knownIPs = event.user.app_metadata?.known_ips || []
  if (!knownIPs.includes(event.request.ip)) {
    score += 0.3
  }

  // 평소와 다른 시간대
  const hour = new Date().getHours()
  const usualHours = event.user.app_metadata?.usual_hours || []
  if (!usualHours.includes(hour)) {
    score += 0.2
  }

  return score
}
```

### 5. API Authorization

```typescript
// lib/auth0-api.ts
import { initAuth0 } from '@auth0/nextjs-auth0'

export const auth0 = initAuth0({
  audience: process.env.AUTH0_AUDIENCE,
  scope: 'openid profile email read:users write:users',
})

// middleware.ts - API 보호
import { withApiAuthRequired } from '@auth0/nextjs-auth0'

export default withApiAuthRequired(async function middleware(req) {
  const session = await getSession(req)

  // Access Token에서 permissions 확인
  const permissions = session?.user['permissions'] || []

  if (!permissions.includes('read:users')) {
    return Response.json({ error: 'Insufficient permissions' }, { status: 403 })
  }

  return Response.next()
})

// app/api/users/route.ts
import { withApiAuthRequired } from '@auth0/nextjs-auth0'
import { hasPermission } from '@/lib/permissions'

export const GET = withApiAuthRequired(async function handler(req) {
  if (!hasPermission(req, 'read:users')) {
    return Response.json({ error: 'Forbidden' }, { status: 403 })
  }

  const users = await fetchUsers()
  return Response.json(users)
})
```

### 6. Anomaly Detection

Auth0 Dashboard에서 자동으로 활성화:

- **Breached Password Detection**: 유출된 비밀번호 자동 차단
- **Brute Force Protection**: 무차별 대입 공격 방어
- **Suspicious IP Throttling**: 의심스러운 IP 제한
- **Bot Detection**: 봇 트래픽 차단

```javascript
// Auth0 Action: Custom Anomaly Response
exports.onExecutePostLogin = async (event, api) => {
  // Auth0가 감지한 이상 징후
  if (event.authentication?.methods?.some(m => m.name === 'mfa')) {
    // MFA 통과 시 정상 처리
    return
  }

  // 의심스러운 활동 감지
  if (event.stats.logins_count > 100 && event.user.logins_count < 10) {
    // 비정상적으로 많은 로그인 시도
    api.access.deny('suspicious_activity', 'Suspicious login pattern detected')

    // 관리자 알림
    await notifyAdmin({
      user: event.user.email,
      reason: 'Abnormal login pattern',
    })
  }
}
```

## 시작하기

### 1. Auth0 Application 생성

1. [Auth0 Dashboard](https://manage.auth0.com/) 접속
2. Applications → Create Application
3. 애플리케이션 타입 선택: **Regular Web Application**
4. Allowed Callback URLs: `http://localhost:3000/api/auth/callback`
5. Allowed Logout URLs: `http://localhost:3000`

### 2. Next.js 통합

```bash
# Auth0 SDK 설치
npm install @auth0/nextjs-auth0
```

```bash
# 환경 변수 설정
cat > .env.local << EOF
AUTH0_SECRET=$(openssl rand -hex 32)
AUTH0_BASE_URL=http://localhost:3000
AUTH0_ISSUER_BASE_URL=https://your-tenant.auth0.com
AUTH0_CLIENT_ID=your-client-id
AUTH0_CLIENT_SECRET=your-client-secret
AUTH0_AUDIENCE=https://your-api.example.com
EOF
```

### 3. API Routes 설정

```typescript
// app/api/auth/[auth0]/route.ts
import { handleAuth } from '@auth0/nextjs-auth0'

export const GET = handleAuth()
```

### 4. 사용자 정보 가져오기

```typescript
// app/profile/page.tsx
import { getSession } from '@auth0/nextjs-auth0'

export default async function Profile() {
  const session = await getSession()

  if (!session) {
    return <div>Please log in</div>
  }

  return (
    <div>
      <img src={session.user.picture} alt="Profile" />
      <h1>{session.user.name}</h1>
      <p>{session.user.email}</p>
    </div>
  )
}
```

## 사용 가이드

### Pattern H: Enterprise Security

```typescript
// app/api/auth/[auth0]/route.ts
import { handleAuth, handleLogin, handleCallback } from '@auth0/nextjs-auth0'

export const GET = handleAuth({
  login: handleLogin({
    authorizationParams: {
      // Enterprise connection
      connection: 'google-oauth2',
      // SAML connection
      // connection: 'saml-enterprise',
      // AD/LDAP connection
      // connection: 'ad-directory',

      // Custom domain (Professional+ plan)
      // domain: 'login.mycompany.com',

      // Audience for API access
      audience: process.env.AUTH0_AUDIENCE,

      // Scopes
      scope: 'openid profile email offline_access read:users',
    },
  }),

  callback: handleCallback({
    afterCallback: async (req, session) => {
      // 콜백 후 추가 로직
      await logUserLogin(session.user.sub)

      return session
    },
  }),

  onError(req, error) {
    console.error('Auth error:', error)
    // 커스텀 에러 페이지로 리다이렉트
    return Response.redirect('/error?message=' + error.message)
  },
})
```

## 코드 예제

### 1. Machine-to-Machine (M2M) Authentication

```typescript
// lib/auth0-m2m.ts
import { ManagementClient } from 'auth0'

const management = new ManagementClient({
  domain: process.env.AUTH0_DOMAIN!,
  clientId: process.env.AUTH0_M2M_CLIENT_ID!,
  clientSecret: process.env.AUTH0_M2M_CLIENT_SECRET!,
})

export async function getAccessToken() {
  const response = await fetch(`https://${process.env.AUTH0_DOMAIN}/oauth/token`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      grant_type: 'client_credentials',
      client_id: process.env.AUTH0_M2M_CLIENT_ID,
      client_secret: process.env.AUTH0_M2M_CLIENT_SECRET,
      audience: process.env.AUTH0_AUDIENCE,
    }),
  })

  const data = await response.json()
  return data.access_token
}

// 사용 예제
const token = await getAccessToken()
const response = await fetch('https://api.example.com/data', {
  headers: {
    Authorization: `Bearer ${token}`,
  },
})
```

### 2. Silent Authentication (Refresh Tokens)

```typescript
// lib/auth0-refresh.ts
export async function refreshAccessToken(refreshToken: string) {
  const response = await fetch(`https://${process.env.AUTH0_DOMAIN}/oauth/token`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      grant_type: 'refresh_token',
      client_id: process.env.AUTH0_CLIENT_ID,
      client_secret: process.env.AUTH0_CLIENT_SECRET,
      refresh_token: refreshToken,
    }),
  })

  return await response.json()
}

// middleware.ts - 자동 토큰 갱신
import { getSession, updateSession } from '@auth0/nextjs-auth0'

export async function middleware(request: Request) {
  const session = await getSession()

  if (session?.accessTokenExpiresAt) {
    const expiresAt = new Date(session.accessTokenExpiresAt)
    const now = new Date()

    // 토큰이 5분 이내에 만료되면 갱신
    if (expiresAt.getTime() - now.getTime() < 5 * 60 * 1000) {
      const newTokens = await refreshAccessToken(session.refreshToken)

      await updateSession(request, {
        ...session,
        accessToken: newTokens.access_token,
        accessTokenExpiresAt: Date.now() + newTokens.expires_in * 1000,
      })
    }
  }

  return NextResponse.next()
}
```

### 3. Custom Password Reset

```typescript
// app/api/auth/reset-password/route.ts
import { ManagementClient } from 'auth0'

const management = new ManagementClient({
  domain: process.env.AUTH0_DOMAIN!,
  clientId: process.env.AUTH0_CLIENT_ID!,
  clientSecret: process.env.AUTH0_CLIENT_SECRET!,
})

export async function POST(request: Request) {
  const { email } = await request.json()

  // 비밀번호 재설정 이메일 전송
  await management.tickets.changePassword({
    email,
    connection_id: 'con_YOUR_CONNECTION_ID',
    // 또는 email을 사용
    // user_id: 'auth0|123456',
  })

  return Response.json({ message: 'Password reset email sent' })
}
```

### 4. User Impersonation (Support)

```typescript
// app/api/admin/impersonate/route.ts
import { ManagementClient } from 'auth0'

const management = new ManagementClient({
  domain: process.env.AUTH0_DOMAIN!,
  clientId: process.env.AUTH0_CLIENT_ID!,
  clientSecret: process.env.AUTH0_CLIENT_SECRET!,
})

export async function POST(request: Request) {
  const { userId, adminId } = await request.json()

  // 관리자 권한 확인
  const admin = await management.users.get({ id: adminId })
  if (!admin.app_metadata?.roles?.includes('admin')) {
    return Response.json({ error: 'Forbidden' }, { status: 403 })
  }

  // Impersonation 토큰 생성
  const impersonation = await management.users.createImpersonationToken({
    user_id: userId,
    protocol: 'oauth2',
    impersonator_id: adminId,
    additionalParameters: {
      response_type: 'code',
      state: 'STATE_VALUE',
    },
  })

  return Response.json({ url: impersonation })
}
```

## Best Practices

### 1. 보안 강화

```javascript
// Auth0 Action: Security Headers
exports.onExecutePostLogin = async (event, api) => {
  // Custom claims에 보안 정보 추가
  api.idToken.setCustomClaim('https://myapp.com/security', {
    mfa_enabled: event.user.multifactor?.length > 0,
    last_password_reset: event.user.last_password_reset,
    risk_score: await calculateRisk(event),
  })
}
```

### 2. 감사 로그

```typescript
// lib/audit-log.ts
export async function logAuthEvent(event: any) {
  await prisma.auditLog.create({
    data: {
      userId: event.user.sub,
      action: event.type,
      ip: event.request.ip,
      userAgent: event.request.userAgent,
      metadata: event,
      timestamp: new Date(),
    },
  })
}
```

### 3. Rate Limiting

```javascript
// Auth0 Action: Custom Rate Limiting
const redis = require('redis')
const client = redis.createClient({ url: event.secrets.REDIS_URL })

exports.onExecutePostLogin = async (event, api) => {
  await client.connect()

  const key = `login:${event.user.user_id}`
  const count = await client.incr(key)

  if (count === 1) {
    await client.expire(key, 3600) // 1시간
  }

  if (count > 10) {
    api.access.deny('rate_limit_exceeded', 'Too many login attempts')
  }

  await client.disconnect()
}
```

## 문제 해결

### 1. Token Expiration

```typescript
// lib/token-validation.ts
import { jwtVerify } from 'jose'

export async function validateToken(token: string) {
  try {
    const JWKS = createRemoteJWKSet(
      new URL(`https://${process.env.AUTH0_DOMAIN}/.well-known/jwks.json`)
    )

    const { payload } = await jwtVerify(token, JWKS, {
      issuer: `https://${process.env.AUTH0_DOMAIN}/`,
      audience: process.env.AUTH0_AUDIENCE,
    })

    return payload
  } catch (error) {
    console.error('Token validation failed:', error)
    return null
  }
}
```

### 2. CORS Issues

```typescript
// next.config.js
module.exports = {
  async headers() {
    return [
      {
        source: '/api/auth/:path*',
        headers: [
          { key: 'Access-Control-Allow-Credentials', value: 'true' },
          { key: 'Access-Control-Allow-Origin', value: process.env.ALLOWED_ORIGIN },
          { key: 'Access-Control-Allow-Methods', value: 'GET,POST' },
          { key: 'Access-Control-Allow-Headers', value: 'Content-Type, Authorization' },
        ],
      },
    ]
  },
}
```

## 다음 단계

- [Clerk 가이드](/ko/skills/baas/clerk) - Modern 대안
- [Pattern H: Enterprise Security](/ko/skills/patterns/pattern-h) - 보안 아키텍처
- [Vercel 가이드](/ko/skills/baas/vercel) - 배포 플랫폼
- [BaaS 개요](/ko/skills/baas) - 플랫폼 비교

## 참고 자료

- [Auth0 공식 문서](https://auth0.com/docs)
- [Next.js SDK](https://auth0.com/docs/quickstart/webapp/nextjs)
- [Actions 가이드](https://auth0.com/docs/customize/actions)
- [Auth0 Community](https://community.auth0.com/)
