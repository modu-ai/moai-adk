---

name: moai-baas-clerk-ext
description: Enterprise Clerk Authentication Platform with AI-powered modern identity

---

# Quick Reference (Level 1)

## Clerk Authentication Platform (November 2025)

### Core Features
- **Modern Authentication**: Passwordless, social login, biometric authentication
- **Multi-Platform Support**: Web, mobile, native applications with unified auth
- **Organizations**: Built-in multi-tenant workspace management
- **WebAuthn Integration**: Hardware security keys and biometric authentication
- **Real-time Sessions**: Advanced session management with cross-device sync

### Latest Versions (November 2025)
- **@clerk/nextjs**: v6.35.0 - Enhanced Next.js integration
- **@clerk/clerk-js**: v5.107.0 - Core JavaScript SDK improvements
- **@clerk/chrome-extension**: v2.7.14 - Chrome extension authentication
- **Android SDK**: Generally available with full feature parity

### Key Authentication Methods
- **Email/Password**: Traditional authentication with enhanced security
- **Social Login**: 30+ providers including Google, GitHub, Discord
- **Passwordless**: Magic links, email/SMS OTP
- **WebAuthn**: Hardware security keys, Windows Hello, Touch ID
- **M2M Tokens**: Machine-to-machine authentication

### When to Use
**Automatic triggers**:
- Clerk authentication architecture and modern identity discussions
- Multi-platform user management and organization implementation
- WebAuthn and modern authentication method integration
- Real-time user experience and session management

**Manual invocation**:
- Designing enterprise Clerk architectures with optimal user experience
- Implementing organization management and multi-tenant authentication
- Planning migrations from traditional authentication systems
- Optimizing user onboarding and security configurations


# Core Implementation (Level 2)

## Clerk Architecture Intelligence

```python
# AI-powered Clerk architecture optimization with Context7
class ClerkArchitectOptimizer:
    def __init__(self):
        self.context7_client = Context7Client()
        self.auth_analyzer = AuthenticationAnalyzer()
        self.ux_optimizer = UserExperienceOptimizer()
    
    async def design_optimal_clerk_architecture(self, 
                                              requirements: AuthenticationRequirements) -> ClerkArchitecture:
        """Design optimal Clerk architecture using AI analysis."""
        
        # Get latest Clerk and authentication documentation via Context7
        clerk_docs = await self.context7_client.get_library_docs(
            context7_library_id='/clerk/docs',
            topic="authentication user management organizations webauthn 2025",
            tokens=3000
        )
        
        # Optimize user experience flows
        ux_design = self.ux_optimizer.optimize_user_flows(
            requirements.user_preferences,
            requirements.platform_requirements,
            clerk_docs
        )
        
        # Configure security framework
        security_config = self.auth_analyzer.configure_security(
            requirements.security_level,
            requirements.threat_model,
            clerk_docs
        )
        
        return ClerkArchitecture(
            authentication_flows=self._design_auth_flows(requirements),
            organization_setup=self._configure_organizations(requirements),
            security_framework=security_config,
            user_experience=ux_design
        )
```

## Multi-Platform Authentication Setup

```typescript
// Next.js application with Clerk integration
import { ClerkProvider, SignIn, SignUp, UserButton } from '@clerk/nextjs';
import { dark } from '@clerk/themes';

function MyApp({ Component, pageProps }) {
  return (
    <ClerkProvider
      appearance={{
        baseTheme: dark,
        variables: {
          colorPrimary: '#ffffff',
          colorBackground: '#1a1a1a',
        }
      }}
      publishableKey={process.env.NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY}
    >
      <Component {...pageProps} />
    </ClerkProvider>
  );
}

// Protected route component
import { useAuth, RedirectToSignIn } from '@clerk/nextjs';

export default function ProtectedPage() {
  const { isSignedIn, user, isLoaded } = useAuth();

  if (!isLoaded) return <div>Loading...</div>;
  if (!isSignedIn) return <RedirectToSignIn />;

  return (
    <div>
      <h1>Welcome, {user.firstName}!</h1>
      <UserButton afterSignOutUrl="/" />
    </div>
  );
}
```

## Organization Management Implementation

```typescript
import { useOrganization, useUser, OrganizationList, CreateOrganization } from '@clerk/nextjs';

export function OrganizationManagement() {
  const { organization, isLoaded, membership } = useOrganization();
  const { user } = useUser();

  if (!isLoaded) return <div>Loading organization...</div>;

  return (
    <div className="organization-management">
      {organization ? (
        <div className="current-organization">
          <h2>{organization.name}</h2>
          <p>Role: {membership?.role}</p>
          
          {membership?.role === 'admin' && (
            <div className="admin-panel">
              <h3>Admin Controls</h3>
              <OrganizationInvitation />
              <MemberList />
            </div>
          )}
        </div>
      ) : (
        <div className="no-organization">
          <h3>Join or Create an Organization</h3>
          <OrganizationList hidePersonal />
          <CreateOrganization />
        </div>
      )}
    </div>
  );
}
```


# Advanced Patterns (Level 3)

## WebAuthn Security Implementation

```typescript
import { useAuth } from '@clerk/nextjs';
import { startAuthentication } from '@simplewebauthn/browser';

export function WebAuthnSecurity() {
  const { user } = useAuth();

  const enableWebAuthn = async () => {
    try {
      const authResp = await startAuthentication({
        challenge: 'random_challenge_string',
        allowCredentials: [],
        userVerification: 'required',
        timeout: 60000,
      });

      await fetch('/api/auth/webauthn/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          credential: authResp,
          userId: user.id,
        }),
      });

      console.log('WebAuthn security key added successfully');
    } catch (error) {
      console.error('WebAuthn registration failed:', error);
    }
  };

  return (
    <div className="webauthn-security">
      <h3>Security Keys</h3>
      <button onClick={enableWebAuthn}>
        Add Security Key (WebAuthn)
      </button>
      <SecurityKeyList />
    </div>
  );
}
```

## Real-time User Experience

```typescript
import { useAuth, useUser } from '@clerk/nextjs';
import { useState, useEffect } from 'react';

export function RealtimeUserExperience() {
  const { user, isLoaded } = useUser();
  const [onlineStatus, setOnlineStatus] = useState<'online' | 'away' | 'offline'>('online');
  const [lastActivity, setLastActivity] = useState(Date.now());

  useEffect(() => {
    const handleActivity = () => {
      setLastActivity(Date.now());
      setOnlineStatus('online');
    };

    const checkInactivity = () => {
      const inactiveTime = Date.now() - lastActivity;
      if (inactiveTime > 300000) setOnlineStatus('away');
      if (inactiveTime > 900000) setOnlineStatus('offline');
    };

    window.addEventListener('mousemove', handleActivity);
    window.addEventListener('keydown', handleActivity);
    
    const inactivityTimer = setInterval(checkInactivity, 60000);

    return () => {
      window.removeEventListener('mousemove', handleActivity);
      window.removeEventListener('keydown', handleActivity);
      clearInterval(inactivityTimer);
    };
  }, [lastActivity]);

  if (!isLoaded) return <div>Loading user experience...</div>;

  return (
    <div className="user-experience">
      <div className={`status-indicator ${onlineStatus}`}>
        <span className="status-dot"></span>
        <span className="status-text">{onlineStatus}</span>
      </div>
      
      {user?.publicMetadata?.theme && (
        <div className="theme-applied">
          Theme: {user.publicMetadata.theme}
        </div>
      )}
      
      <PersonalizedFeatures user={user} />
    </div>
  );
}
```

## M2M (Machine-to-Machine) Authentication

```typescript
export function M2MAuthentication() {
  const [m2mToken, setM2mToken] = useState<string | null>(null);

  const generateM2MToken = async () => {
    try {
      const response = await fetch('/api/auth/m2m/token', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${process.env.CLERK_SECRET_KEY}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          template: 'service_account',
          expires_in_seconds: 3600,
        }),
      });

      const data = await response.json();
      setM2mToken(data.jwt);
    } catch (error) {
      console.error('M2M token generation failed:', error);
    }
  };

  return (
    <div className="m2m-authentication">
      <h3>Machine-to-Machine Authentication</h3>
      <button onClick={generateM2MToken}>
        Generate M2M Token
      </button>
      {m2mToken && (
        <div className="token-display">
          <p>Token generated successfully</p>
          <code>{m2mToken.substring(0, 20)}...</code>
        </div>
      )}
    </div>
  );
}
```


**End of Skill** | Updated 2025-11-21

## Context7 Integration

### Related Libraries & Tools
- [Clerk](/clerk/javascript): Complete user management platform with authentication
- [@clerk/nextjs](/clerk/javascript): Clerk SDK for Next.js applications
- [@clerk/clerk-js](/clerk/javascript): Clerk JavaScript SDK for web applications
- [@clerk/backend](/clerk/javascript): Clerk backend SDK for server-side operations
- [@clerk/chrome-extension](/clerk/javascript): Clerk SDK for Chrome extensions

### Official Documentation
- [Clerk Documentation](https://clerk.com/docs)
- [Clerk Next.js](https://clerk.com/docs/quickstarts/nextjs)
- [Clerk API Reference](https://clerk.com/docs/references/backend/overview)
- [Organizations](https://clerk.com/docs/organizations/overview)
- [WebAuthn with Clerk](https://clerk.com/docs/custom-flows/webauthn)

### Version-Specific Guides
Latest stable version: @clerk/nextjs v6.35.0, @clerk/clerk-js v5.107.0
- [Clerk v6 Migration](https://clerk.com/docs/upgrade-guides/v6)
- [Next.js 15 with Clerk](https://clerk.com/docs/quickstarts/nextjs)
- [Organizations Setup](https://clerk.com/docs/organizations/setup)
- [M2M Authentication](https://clerk.com/docs/backend-requests/handling/overview)

