---
name: "moai-security-auth"
version: "4.0.0"
status: stable
description: "Enterprise Skill for advanced development"
allowed-tools: "Read, Bash, WebSearch, WebFetch"
---

# moai-security-auth: Modern Authentication Patterns

**Advanced Authentication with MFA, FIDO2, WebAuthn & Passkeys**  
Trust Score: 9.8/10 | Version: 4.0.0 | Enterprise Mode | Last Updated: 2025-11-12

---

## Overview

Authentication is the foundation of application security. Modern patterns have evolved from passwords to passwordless authentication using FIDO2, WebAuthn, and Passkeys. This Skill covers current best practices for NextAuth.js 5.x, Passport.js, and FIDO2 implementations.

**When to use this Skill:**
- Implementing secure session management
- Adding multi-factor authentication (MFA)
- Migrating to passwordless authentication (Passkeys/WebAuthn)
- Building FIDO2 hardware key support
- Setting up OAuth 2.1 provider integration
- Implementing JWT session refresh logic
- Building secure single sign-on (SSO) systems
- Managing authentication state in distributed systems

---

## Level 1: Foundations (FREE TIER)

### What is Modern Authentication?

```
Legacy Flow (2010s):
User → Username/Password → Database Hash Comparison → Session Token

Modern Flow (2025):
User → Biometric/Hardware → WebAuthn Server → Cryptographic Verification
           OR
User → OAuth Provider → Provider Verification → Access Token + ID Token
```

**Authentication Evolution:**

| Era | Method | Security | User Experience |
|-----|--------|----------|-----------------|
| **2000-2010** | Password | Weak | Good |
| **2010-2020** | Password + 2FA | Medium | Poor |
| **2020-2025** | Passwordless | Strong | Excellent |
| **2025+** | Passkeys | Strongest | Best |

### FIDO2 & WebAuthn Fundamentals

**FIDO2 Standard (2018):**
- Published by FIDO Alliance (Google, Microsoft, Apple, etc.)
- Two-factor authentication using hardware keys or biometrics
- No password transmission needed

**WebAuthn (W3C Standard):**
- Browser API for FIDO2 authentication
- Supports USB keys, biometrics, platform authenticators
- Cryptographic proof instead of password verification

**Registration Flow:**

```
User Device          Authenticator          Relying Party (Server)
    |                    |                         |
    |--Challenge------->|                         |
    |                    |--User Verification     |
    |                    |  (Biometric/PIN)       |
    |                    |                         |
    |<--Attestation------|                         |
    |                    |                         |
    |--PublicKey + Attestation------->Verify|Store
```

**Authentication Flow:**

```
User Device          Authenticator          Relying Party (Server)
    |                    |                         |
    |                    |<--Challenge------------|
    |                    |                         |
    |--User Verification-|                         |
    |                    |--Assertion (Signed)-->|Verify Signature
    |                    |                         |
    |<--Authenticated----|<--Success--------------|
```

### Session Management (NextAuth.js 5.x)

**NextAuth.js 5.0.0 (November 2025):**
- Complete rewrite for Next.js 15
- JWT sessions by default (stateless)
- Built-in OAuth/OIDC provider support
- Passwordless email magic links

**Key Concepts:**

1. **Callbacks** (Control authentication flow)
   - `authorized`: Check if user has access
   - `signIn`: Validate credentials before session creation
   - `jwt`: Modify JWT payload
   - `session`: Customize session object

2. **Providers** (Authentication sources)
   - OAuth (Google, GitHub, Microsoft)
   - Credentials (username/password)
   - Email magic link
   - FIDO2/WebAuthn

3. **Events** (Logging & audit trail)
   - `signIn`: User logged in
   - `signOut`: User logged out
   - `createUser`: New user registered
   - `updateUser`: User updated

---

## Level 2: Intermediate Patterns (STANDARD TIER)

### NextAuth.js 5.x Implementation

**Basic Setup (JWT Sessions):**

```typescript
// lib/auth.ts
import NextAuth, { type NextAuthConfig } from 'next-auth';
import GitHub from 'next-auth/providers/github';
import Credentials from 'next-auth/providers/credentials';
import { encode as defaultEncode, decode as defaultDecode } from 'next-auth/jwt';

const config = {
  providers: [
    // 1. OAuth Provider (GitHub)
    GitHub({
      clientId: process.env.GITHUB_ID,
      clientSecret: process.env.GITHUB_SECRET,
      allowDangerousEmailAccountLinking: false
    }),
    
    // 2. Passwordless Email Magic Link
    Email({
      server: {
        host: process.env.EMAIL_SERVER_HOST,
        port: parseInt(process.env.EMAIL_SERVER_PORT),
        auth: {
          user: process.env.EMAIL_SERVER_USER,
          pass: process.env.EMAIL_SERVER_PASSWORD
        }
      },
      from: process.env.EMAIL_FROM
    }),
    
    // 3. Credentials (for custom auth with MFA)
    Credentials({
      credentials: {
        email: { label: 'Email', type: 'email' },
        password: { label: 'Password', type: 'password' },
        mfaCode: { label: '2FA Code', type: 'text', optional: true }
      },
      async authorize(credentials) {
        // 1. Verify password
        const user = await db.users.findByEmail(credentials.email);
        if (!user) return null;
        
        const passwordMatch = await bcrypt.compare(
          credentials.password,
          user.passwordHash
        );
        
        if (!passwordMatch) return null;
        
        // 2. Check MFA if enabled
        if (user.mfaEnabled) {
          if (!credentials.mfaCode) {
            throw new Error('MFA code required');
          }
          
          const mfaValid = await verifyTOTP(
            credentials.mfaCode,
            user.mfaSecret
          );
          
          if (!mfaValid) {
            throw new Error('Invalid MFA code');
          }
        }
        
        return {
          id: user.id,
          email: user.email,
          name: user.name
        };
      }
    })
  ],
  
  // JWT configuration
  jwt: {
    encode: async (params) => {
      // Use RS256 for better security
      if (params.token?.mfa === true) {
        // Short-lived token for MFA verification
        return defaultEncode({ ...params, exp: Date.now() / 1000 + 300 });
      }
      return defaultEncode(params);
    },
    decode: defaultDecode
  },
  
  // Session configuration
  session: {
    strategy: 'jwt',         // Stateless sessions
    maxAge: 30 * 24 * 60 * 60, // 30 days
    updateAge: 24 * 60 * 60   // Refresh every 24h
  },
  
  // Callbacks for customization
  callbacks: {
    // Control who can sign in
    authorized({ auth, request }) {
      const isLoggedIn = !!auth?.user;
      const isOnAdminPage = request.nextUrl.pathname.startsWith('/admin');
      
      if (isOnAdminPage) {
        return isLoggedIn && auth.user.role === 'admin';
      }
      return true;
    },
    
    // Validate credentials
    async signIn({ user, account, profile }) {
      // Verify email if required
      if (!user.emailVerified) {
        throw new Error('Email not verified');
      }
      
      // Check if account is locked
      if (user.locked) {
        throw new Error('Account locked');
      }
      
      return true;
    },
    
    // Modify JWT token
    async jwt({ token, user, account }) {
      // Initial sign in
      if (user) {
        token.id = user.id;
        token.role = user.role;
      }
      
      // Refresh token data on each session
      if (account?.access_token) {
        token.accessToken = account.access_token;
        token.accessTokenExpires = account.expires_at * 1000;
      }
      
      // Check if token needs refresh
      if (token.accessTokenExpires && 
          Date.now() < token.accessTokenExpires) {
        return token;
      }
      
      // Refresh access token if expired
      return refreshAccessToken(token);
    },
    
    // Customize session object
    async session({ session, token }) {
      session.user.id = token.id;
      session.user.role = token.role;
      session.accessToken = token.accessToken;
      return session;
    }
  },
  
  // Pages customization
  pages: {
    signIn: '/auth/signin',
    error: '/auth/error',
    newUser: '/auth/welcome'
  }
} satisfies NextAuthConfig;

export const { handlers, auth, signIn, signOut } = NextAuth(config);
```

### FIDO2/WebAuthn Implementation

**Registration (User Setup):**

```typescript
// lib/webauthn.ts
import { generateRegistrationOptions, verifyRegistrationResponse } from '@simplewebauthn/server';
import { isoBase64URL } from '@simplewebauthn/server/helpers/iso';

export async function startRegistration(user: User) {
  // 1. Generate challenge
  const options = generateRegistrationOptions({
    rpID: process.env.WEBAUTHN_RP_ID,        // domain.com
    rpName: 'My Application',
    userID: isoBase64URL.fromBuffer(Buffer.from(user.id)),
    userName: user.email,
    userDisplayName: user.name,
    
    // 2. Require user verification (biometric/PIN)
    authenticatorSelection: {
      authenticatorAttachment: 'cross-platform',  // USB key
      residentKey: 'preferred',                    // Passkey support
      userVerification: 'preferred'                // Biometric
    },
    
    // 3. Attestation for registration verification
    attestationType: 'direct',
    
    // 4. Support multiple algorithms
    supportedAlgos: [-7, -257]  // ES256, RS256
  });
  
  // 3. Store challenge in session (expires in 15 minutes)
  await redis.setex(
    `webauthn:challenge:${user.id}`,
    900,
    JSON.stringify(options.challenge)
  );
  
  return options;
}

export async function completeRegistration(
  user: User,
  attestationResponse: PublicKeyCredential
) {
  // 1. Get stored challenge
  const challengeStr = await redis.get(`webauthn:challenge:${user.id}`);
  const challenge = JSON.parse(challengeStr);
  
  // 2. Verify attestation response
  const verification = await verifyRegistrationResponse({
    response: attestationResponse,
    expectedChallenge: challenge,
    expectedRPID: process.env.WEBAUTHN_RP_ID,
    expectedOrigin: process.env.WEBAUTHN_ORIGIN,
    
    // Verify authenticator is certified
    requireResidentKey: false,
    requireUserVerification: true
  });
  
  if (!verification.verified) {
    throw new Error('WebAuthn registration verification failed');
  }
  
  // 3. Store public key and counter
  await db.webauthnCredentials.create({
    user_id: user.id,
    credential_id: isoBase64URL.toBuffer(
      verification.registrationInfo.credentialID
    ),
    public_key: verification.registrationInfo.credentialPublicKey,
    counter: verification.registrationInfo.counter,
    transports: attestationResponse.response.getTransports(),
    created_at: new Date()
  });
  
  // 4. Clean up challenge
  await redis.del(`webauthn:challenge:${user.id}`);
  
  return verification;
}
```

**Authentication (Sign In):**

```typescript
import { generateAuthenticationOptions, verifyAuthenticationResponse } from '@simplewebauthn/server';

export async function startAuthentication(email: string) {
  // 1. Get user's credentials
  const user = await db.users.findByEmail(email);
  if (!user) throw new Error('User not found');
  
  const credentials = await db.webauthnCredentials.findByUserId(user.id);
  
  // 2. Generate challenge
  const options = generateAuthenticationOptions({
    rpID: process.env.WEBAUTHN_RP_ID,
    
    // User must verify with same credential
    allowCredentials: credentials.map(cred => ({
      id: cred.credential_id,
      transports: cred.transports
    })),
    
    userVerification: 'required'  // Must verify identity
  });
  
  // 3. Store challenge for verification
  await redis.setex(
    `webauthn:auth:${user.id}`,
    900,
    JSON.stringify(options.challenge)
  );
  
  return options;
}

export async function completeAuthentication(
  email: string,
  assertionResponse: PublicKeyCredential
) {
  // 1. Get user
  const user = await db.users.findByEmail(email);
  
  // 2. Get credential used
  const credentialID = assertionResponse.id;
  const credential = await db.webauthnCredentials.findByCredentialID(
    Buffer.from(credentialID, 'utf-8')
  );
  
  // 3. Get challenge
  const challengeStr = await redis.get(`webauthn:auth:${user.id}`);
  const challenge = JSON.parse(challengeStr);
  
  // 4. Verify assertion
  const verification = await verifyAuthenticationResponse({
    response: assertionResponse,
    expectedChallenge: challenge,
    expectedRPID: process.env.WEBAUTHN_RP_ID,
    expectedOrigin: process.env.WEBAUTHN_ORIGIN,
    
    // Verify counter prevents cloning
    authenticator: {
      credentialID: credential.credential_id,
      credentialPublicKey: credential.public_key,
      counter: credential.counter
    }
  });
  
  if (!verification.verified) {
    throw new Error('WebAuthn authentication failed');
  }
  
  // 5. Update counter (prevents cloning)
  await db.webauthnCredentials.update(credential.id, {
    counter: verification.authenticationInfo.newCounter
  });
  
  return user;
}
```

### Multi-Factor Authentication (TOTP)

**Time-based One-Time Password:**

```typescript
import { authenticator } from 'otplib';
import QRCode from 'qrcode';

export async function setupTOTP(user: User) {
  // 1. Generate secret
  const secret = authenticator.generateSecret({
    name: `My App (${user.email})`
  });
  
  // 2. Generate QR code
  const qrCode = await QRCode.toDataURL(secret);
  
  // 3. Store temporary (not yet verified)
  await redis.setex(
    `totp:pending:${user.id}`,
    600,  // 10 minutes
    secret
  );
  
  return { secret, qrCode };
}

export async function verifyTOTPSetup(user: User, token: string) {
  // 1. Get pending secret
  const secret = await redis.get(`totp:pending:${user.id}`);
  if (!secret) throw new Error('No pending TOTP setup');
  
  // 2. Verify token
  const isValid = authenticator.check(token, secret);
  if (!isValid) throw new Error('Invalid token');
  
  // 3. Verify backup codes
  const backupCodes = Array.from({ length: 10 }).map(() =>
    crypto.randomBytes(4).toString('hex').toUpperCase()
  );
  
  // 4. Store permanently
  await db.users.update(user.id, {
    mfaEnabled: true,
    mfaSecret: secret,
    mfaBackupCodes: backupCodes.map(code =>
      bcrypt.hashSync(code, 10)
    )
  });
  
  // Clean up
  await redis.del(`totp:pending:${user.id}`);
  
  return backupCodes;
}

export async function verifyTOTPToken(user: User, token: string) {
  // Check if token is backup code
  const isBackup = user.mfaBackupCodes.some(hashedCode =>
    bcrypt.compareSync(token, hashedCode)
  );
  
  if (isBackup) {
    // Mark backup code as used
    const codes = user.mfaBackupCodes.filter(
      code => !bcrypt.compareSync(token, code)
    );
    await db.users.update(user.id, { mfaBackupCodes: codes });
    return true;
  }
  
  // Check TOTP token
  const isValid = authenticator.check(token, user.mfaSecret);
  return isValid;
}

// Rate limit TOTP attempts (prevent brute force)
const totpAttempts = new Map<string, number>();

export async function verifyTOTPWithRateLimit(
  user: User,
  token: string
) {
  const key = `totp:${user.id}`;
  const attempts = totpAttempts.get(key) || 0;
  
  if (attempts >= 5) {
    throw new Error('Too many failed attempts. Try again in 15 minutes.');
  }
  
  const isValid = await verifyTOTPToken(user, token);
  
  if (!isValid) {
    totpAttempts.set(key, attempts + 1);
    setTimeout(() => totpAttempts.delete(key), 900000); // 15 minutes
    throw new Error('Invalid token');
  }
  
  totpAttempts.delete(key);
  return true;
}
```

### Passport.js with Custom Strategy

**Passport.js 0.7.x:**

```typescript
import { Strategy as LocalStrategy } from 'passport-local';
import { Strategy as JwtStrategy, ExtractJwt } from 'passport-jwt';
import bcrypt from 'bcryptjs';
import passport from 'passport';

// 1. Local Strategy (username/password)
passport.use(new LocalStrategy(
  {
    usernameField: 'email',
    passwordField: 'password',
    passReqToCallback: true
  },
  async (req, email, password, done) => {
    try {
      const user = await db.users.findByEmail(email);
      
      if (!user) {
        return done(null, false, { 
          message: 'Invalid credentials' 
        });
      }
      
      // Check if account is locked
      if (user.loginAttempts >= 5 && 
          Date.now() < user.lockUntil) {
        return done(null, false, { 
          message: 'Account locked. Try again later.' 
        });
      }
      
      const isPasswordValid = await bcrypt.compare(
        password,
        user.passwordHash
      );
      
      if (!isPasswordValid) {
        // Increment failed attempts
        await db.users.update(user.id, {
          loginAttempts: (user.loginAttempts || 0) + 1,
          lockUntil: user.loginAttempts >= 4 
            ? new Date(Date.now() + 30 * 60000) // 30 min lock
            : undefined
        });
        
        return done(null, false, { 
          message: 'Invalid credentials' 
        });
      }
      
      // Reset lock on successful login
      if (user.loginAttempts > 0) {
        await db.users.update(user.id, {
          loginAttempts: 0,
          lockUntil: null
        });
      }
      
      return done(null, user);
    } catch (err) {
      return done(err);
    }
  }
));

// 2. JWT Strategy
passport.use(new JwtStrategy(
  {
    jwtFromRequest: ExtractJwt.fromAuthHeaderAsBearerToken(),
    secretOrKey: process.env.JWT_SECRET,
    algorithms: ['HS256']
  },
  async (jwtPayload, done) => {
    try {
      const user = await db.users.findById(jwtPayload.id);
      
      if (!user) {
        return done(null, false);
      }
      
      // Check if token is blacklisted (logout)
      const isBlacklisted = await redis.get(
        `jwt:blacklist:${jwtPayload.jti}`
      );
      
      if (isBlacklisted) {
        return done(null, false);
      }
      
      return done(null, user, jwtPayload);
    } catch (err) {
      return done(err);
    }
  }
));

// 3. Serialization
passport.serializeUser((user, done) => {
  done(null, user.id);
});

passport.deserializeUser(async (id, done) => {
  try {
    const user = await db.users.findById(id);
    done(null, user);
  } catch (err) {
    done(err);
  }
});

// 4. Express middleware
app.post('/login', 
  passport.authenticate('local', { session: false }),
  (req, res) => {
    const token = jwt.sign(
      { id: req.user.id, jti: uuid() },
      process.env.JWT_SECRET,
      { expiresIn: '7d' }
    );
    
    res.json({ token, user: req.user });
  }
);

// Protected route
app.get('/profile',
  passport.authenticate('jwt', { session: false }),
  (req, res) => {
    res.json(req.user);
  }
);
```

---

## Level 3: Enterprise Patterns (PREMIUM TIER)

### Passkeys (Passwordless Future)

**Passkey Registration & Authentication:**

```typescript
// Passkeys = WebAuthn + Backup Sync (iCloud, Google Password Manager)
// Registration same as WebAuthn, but with different UX

export async function registerPasskey(user: User) {
  const options = generateRegistrationOptions({
    rpID: process.env.WEBAUTHN_RP_ID,
    rpName: 'My Application',
    userID: isoBase64URL.fromBuffer(Buffer.from(user.id)),
    userName: user.email,
    userDisplayName: user.name,
    
    // Passkey-specific settings
    authenticatorSelection: {
      authenticatorAttachment: 'platform',  // Device built-in (not USB)
      residentKey: 'required',               // Passkey must be resident
      userVerification: 'required'           // Biometric/PIN required
    },
    
    attestationType: 'direct'
  });
  
  // Passkey will be synced by platform (iCloud, Google, etc.)
  return options;
}

// Authenticate with any passkey (phone, laptop, shared device)
export async function authenticateWithPasskey(email: string) {
  const user = await db.users.findByEmail(email);
  const credentials = await db.webauthnCredentials.findByUserId(
    user.id,
    { type: 'passkey' }
  );
  
  const options = generateAuthenticationOptions({
    rpID: process.env.WEBAUTHN_RP_ID,
    allowCredentials: []  // Any passkey works
  });
  
  return options;
}
```

### Session Refresh & Sliding Windows

**NextAuth.js JWT Refresh:**

```typescript
async function refreshAccessToken(token: JWT) {
  try {
    // Refresh token with OAuth provider
    const response = await fetch(
      `https://oauth-provider.com/token`,
      {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: new URLSearchParams({
          client_id: process.env.OAUTH_CLIENT_ID,
          client_secret: process.env.OAUTH_CLIENT_SECRET,
          grant_type: 'refresh_token',
          refresh_token: token.refreshToken
        })
      }
    );
    
    const refreshedTokens = await response.json();
    
    if (!response.ok) throw refreshedTokens;
    
    return {
      ...token,
      accessToken: refreshedTokens.access_token,
      accessTokenExpires: Date.now() + refreshedTokens.expires_in * 1000,
      refreshToken: refreshedTokens.refresh_token ?? token.refreshToken
    };
  } catch (error) {
    return { ...token, error: 'RefreshAccessTokenError' };
  }
}

// Sliding window: Extend session if used recently
const jwtCallback = async ({ token, account, user, isNewUser, trigger, session }) => {
  if (trigger === 'update' && session?.name) {
    token.name = session.name;
  }
  
  // Auto-extend session if used in last 24 hours
  if (token.exp && Date.now() < token.exp * 1000 - 24 * 60 * 60 * 1000) {
    token.exp = Date.now() / 1000 + 30 * 24 * 60 * 60; // +30 days
  }
  
  return token;
};
```

### Audit Logging & Suspicious Activity

**Authentication Event Logging:**

```typescript
export async function logAuthEvent(
  userId: string,
  event: string,
  metadata: Record<string, any>
) {
  const ip = metadata.ip;
  const userAgent = metadata.userAgent;
  const geoLocation = await getGeolocation(ip);
  
  // Check for suspicious activity
  const lastLogin = await db.authLogs
    .findLastByUser(userId)
    .select('geoLocation', 'timestamp');
  
  const isSuspicious = lastLogin && (
    // Login from new country in short time
    lastLogin.geoLocation.country !== geoLocation.country &&
    Date.now() - lastLogin.timestamp < 3600000 // 1 hour
  );
  
  await db.authLogs.create({
    user_id: userId,
    event,
    ip,
    userAgent,
    geoLocation,
    suspicious: isSuspicious,
    timestamp: new Date()
  });
  
  // Alert user if suspicious
  if (isSuspicious) {
    await sendEmail(userId, {
      template: 'suspicious-login',
      data: {
        location: geoLocation.city,
        timestamp: new Date().toLocaleString()
      }
    });
  }
}

// Middleware to log all auth events
export function authAuditMiddleware(req, res, next) {
  const originalSend = res.send;
  
  res.send = function(data) {
    if (req.path.includes('/auth/')) {
      logAuthEvent(req.user?.id, req.method, {
        path: req.path,
        status: res.statusCode,
        ip: req.ip,
        userAgent: req.headers['user-agent']
      }).catch(console.error);
    }
    return originalSend.call(this, data);
  };
  
  next();
}
```

---

## Reference

### Official Documentation
- NextAuth.js: https://next-auth.js.org/
- FIDO Alliance: https://fidoalliance.org/
- WebAuthn Specification: https://www.w3.org/TR/webauthn-2/
- SimpleWebAuthn: https://simplewebauthn.dev/
- Passport.js: http://www.passportjs.org/
- OWASP Authentication Cheat Sheet: https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html
- OWASP Password Storage Cheat Sheet: https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html
- RFC 4226 (TOTP): https://tools.ietf.org/html/rfc4226
- RFC 6238 (HOTP): https://tools.ietf.org/html/rfc6238

### Tools & Libraries (November 2025 Versions)
- **next-auth**: 5.0.x
- **passport**: 0.7.x
- **@simplewebauthn/server**: 10.0.x
- **otplib**: 12.x
- **bcryptjs**: 2.4.x
- **jsonwebtoken**: 9.x
- **redis**: 5.x

### Common Vulnerabilities & Mitigations

| Vulnerability | OWASP | Mitigation |
|---|---|---|
| **Weak Password** | A02:2021 | TOTP/WebAuthn instead |
| **Session Fixation** | A02:2021 | Rotate session ID on login |
| **Brute Force** | A07:2021 | Rate limit + account lockout |
| **Token Exposure** | A02:2021 | Store in httpOnly cookie |
| **Credential Stuffing** | A02:2021 | Use bcrypt + salting |
| **MFA Bypass** | A07:2021 | Enforce MFA verification |

---

**Version**: 4.0.0 Enterprise
**Skill Category**: Security (Authentication & Authorization)
**Complexity**: Medium-Advanced
**Time to Implement**: 3-5 hours per component
**Prerequisites**: Node.js, React/Vue.js, OAuth concepts, WebAuthn API
