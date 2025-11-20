---
name: moai-security-auth
version: 4.0.0
status: stable
updated: 2025-11-20
description: Modern authentication patterns with MFA, FIDO2, WebAuthn & Passkeys
category: Security
allowed-tools: [Read, Bash, WebSearch, WebFetch]
---

# moai-security-auth: Modern Authentication Patterns

**Advanced authentication with MFA, FIDO2, WebAuthn & Passkeys**

> Trust Score: 9.8/10 | Version: 4.0.0

---

## Overview

Enterprise authentication expert covering modern security patterns:

- **Passwordless Authentication**: FIDO2, WebAuthn, and Passkeys
- **Multi-Factor Authentication**: TOTP, SMS, and hardware tokens
- **OAuth 2.1 Integration**: Social login and enterprise SSO
- **Session Management**: JWT, refresh tokens, and secure cookies
- **Advanced Security**: Rate limiting, account lockout, and audit logging

**Core Technologies**:

- NextAuth.js 5.x for Next.js applications
- Passport.js for Express.js applications
- WebAuthn API for passwordless authentication
- JWT for stateless session management

---

## Authentication Evolution

| Era       | Method         | Security  | UX        |
| --------- | -------------- | --------- | --------- |
| 2000-2010 | Password       | Weak      | Good      |
| 2010-2020 | Password + 2FA | Medium    | Poor      |
| 2020-2025 | Passwordless   | Strong    | Excellent |
| 2025+     | Passkeys       | Strongest | Best      |

---

## NextAuth.js 5.x Implementation

### Complete Configuration

```typescript
// lib/auth.ts
import NextAuth from "next-auth";
import GitHub from "next-auth/providers/github";
import Credentials from "next-auth/providers/credentials";
import { DrizzleAdapter } from "@auth/drizzle-adapter";

export const config = {
  adapter: DrizzleAdapter(db),
  providers: [
    GitHub({
      clientId: process.env.GITHUB_CLIENT_ID!,
      clientSecret: process.env.GITHUB_CLIENT_SECRET!,
    }),

    Credentials({
      credentials: {
        email: { label: "Email", type: "email" },
        password: { label: "Password", type: "password" },
        mfaCode: { label: "2FA Code", type: "text", optional: true },
      },
      async authorize(credentials) {
        const user = await db.query.users.findFirst({
          where: eq(users.email, credentials.email),
        });

        if (!user) return null;

        // Verify password
        const valid = await bcrypt.compare(
          credentials.password,
          user.passwordHash
        );

        if (!valid) return null;

        // Verify MFA if enabled
        if (user.mfaEnabled && user.mfaSecret) {
          if (!credentials.mfaCode) {
            throw new Error("MFA code required");
          }

          const mfaValid = speakeasy.totp.verify({
            secret: user.mfaSecret,
            encoding: "base32",
            token: credentials.mfaCode,
          });

          if (!mfaValid) throw new Error("Invalid MFA code");
        }

        return {
          id: user.id,
          email: user.email,
          name: user.name,
          role: user.role,
        };
      },
    }),
  ],

  session: {
    strategy: "jwt",
    maxAge: 30 * 24 * 60 * 60, // 30 days
  },

  callbacks: {
    async authorized({ auth, request }) {
      const isLoggedIn = !!auth?.user;
      const isAdmin = auth?.user?.role === "admin";
      const isAdminRoute = request.nextUrl.pathname.startsWith("/admin");

      return isAdminRoute ? isLoggedIn && isAdmin : isLoggedIn;
    },

    async jwt({ token, user }) {
      if (user) {
        token.id = user.id;
        token.role = user.role;
      }
      return token;
    },

    async session({ session, token }) {
      if (token) {
        session.user.id = token.id;
        session.user.role = token.role;
      }
      return session;
    },
  },

  pages: {
    signIn: "/auth/signin",
    error: "/auth/error",
  },
};

export const { handlers, auth, signIn, signOut } = NextAuth(config);
```

---

## MFA Implementation

### TOTP Setup

```typescript
// lib/mfa.ts
import speakeasy from "speakeasy";
import QRCode from "qrcode";

export class MFAService {
  static generateSecret(userEmail: string) {
    return speakeasy.generateSecret({
      name: `MyApp (${userEmail})`,
      issuer: "MyApp",
    });
  }

  static async generateQRCode(secret: speakeasy.GeneratedSecret) {
    const otpauthUrl = speakeasy.otpauthURL({
      secret: secret.base32,
      label: secret.name,
      issuer: secret.issuer,
    });

    return await QRCode.toDataURL(otpauthUrl);
  }

  static verifyToken(secret: string, token: string): boolean {
    return speakeasy.totp.verify({
      secret,
      encoding: "base32",
      token,
      window: 1, // Allow time drift
    });
  }

  static async enableMFA(userId: string, secret: string, token: string) {
    if (!this.verifyToken(secret, token)) {
      throw new Error("Invalid verification code");
    }

    await db
      .update(users)
      .set({
        mfaEnabled: true,
        mfaSecret: secret,
        mfaEnabledAt: new Date(),
      })
      .where(eq(users.id, userId));
  }
}
```

### MFA API Routes

```typescript
// app/api/auth/mfa/enable/route.ts
import { auth } from "@/lib/auth";
import { MFAService } from "@/lib/mfa";

export async function POST(request: Request) {
  const session = await auth();
  if (!session?.user?.id) {
    return Response.json({ error: "Unauthorized" }, { status: 401 });
  }

  const { token } = await request.json();
  const secret = MFAService.generateSecret(session.user.email);

  await MFAService.enableMFA(session.user.id, secret.base32, token);

  return Response.json({ success: true });
}
```

---

## WebAuthn & Passkeys

### WebAuthn Service

```typescript
// lib/webauthn.ts
import {
  generateRegistrationOptions,
  verifyRegistrationResponse,
  generateAuthenticationOptions,
  verifyAuthenticationResponse,
} from "@simplewebauthn/server";

const rpID = process.env.WEBAUTHN_RP_ID!;
const rpName = "MyApp";
const origin = process.env.WEBAUTHN_ORIGIN!;

export class WebAuthnService {
  static async startRegistration(userId: string, email: string, name: string) {
    const options = generateRegistrationOptions({
      rpID,
      rpName,
      userID: new TextEncoder().encode(userId),
      userName: email,
      userDisplayName: name,
      authenticatorSelection: {
        residentKey: "preferred", // Passkey support
        userVerification: "preferred", // Biometric
      },
    });

    // Store challenge in session
    await redis.setex(
      `webauthn:register:${userId}`,
      900, // 15 minutes
      JSON.stringify(options.challenge)
    );

    return options;
  }

  static async completeRegistration(userId: string, response: any) {
    const challengeStr = await redis.get(`webauthn:register:${userId}`);
    if (!challengeStr) {
      throw new Error("Registration challenge expired");
    }

    const expectedChallenge = JSON.parse(challengeStr);
    const verification = await verifyRegistrationResponse({
      response,
      expectedChallenge,
      expectedOrigin: origin,
      expectedRPID: rpID,
      requireUserVerification: true,
    });

    if (!verification.verified || !verification.registrationInfo) {
      throw new Error("Registration verification failed");
    }

    // Store credential
    await db.webauthnCredentials.create({
      userId,
      credentialId: verification.registrationInfo.credentialID,
      credentialPublicKey: verification.registrationInfo.credentialPublicKey,
      counter: verification.registrationInfo.counter,
      transports: response.response.transports,
    });

    await redis.del(`webauthn:register:${userId}`);

    return verification;
  }

  static async startAuthentication(email: string) {
    const user = await db.query.users.findFirst({
      where: eq(users.email, email),
      with: { credentials: true },
    });

    if (!user) throw new Error("User not found");

    const options = generateAuthenticationOptions({
      rpID,
      allowCredentials: user.credentials.map((cred) => ({
        id: cred.credentialId,
        type: "public-key",
        transports: cred.transports,
      })),
      userVerification: "preferred",
    });

    await redis.setex(
      `webauthn:auth:${user.id}`,
      900,
      JSON.stringify(options.challenge)
    );

    return options;
  }

  static async completeAuthentication(email: string, response: any) {
    const user = await db.query.users.findFirst({
      where: eq(users.email, email),
      with: { credentials: true },
    });

    if (!user) throw new Error("User not found");

    const challengeStr = await redis.get(`webauthn:auth:${user.id}`);
    if (!challengeStr) {
      throw new Error("Authentication challenge expired");
    }

    const expectedChallenge = JSON.parse(challengeStr);
    const credential = user.credentials.find(
      (cred) => Buffer.compare(cred.credentialId, response.id) === 0
    );

    if (!credential) throw new Error("Credential not found");

    const verification = await verifyAuthenticationResponse({
      response,
      expectedChallenge,
      expectedOrigin: origin,
      expectedRPID: rpID,
      credential: {
        id: credential.credentialId,
        publicKey: credential.credentialPublicKey,
        counter: credential.counter,
        transports: credential.transports,
      },
      requireUserVerification: true,
    });

    if (!verification.verified) {
      throw new Error("Authentication verification failed");
    }

    // Update counter
    await db.webauthnCredentials
      .update({
        counter: verification.authenticationInfo.newCounter,
      })
      .where(eq(webauthnCredentials.id, credential.id));

    await redis.del(`webauthn:auth:${user.id}`);

    return user;
  }
}
```

---

## Security Middleware

### Rate Limiting

```typescript
// middleware.ts
import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

const rateLimit = new Map<string, { count: number; resetTime: number }>();

export async function middleware(request: NextRequest) {
  const ip = request.ip || "unknown";

  // Rate limit auth endpoints
  if (request.nextUrl.pathname.startsWith("/api/auth/")) {
    const now = Date.now();
    const windowMs = 15 * 60 * 1000; // 15 minutes
    const maxRequests = 5;

    const record = rateLimit.get(ip);

    if (!record || now > record.resetTime) {
      rateLimit.set(ip, {
        count: 1,
        resetTime: now + windowMs,
      });
    } else {
      record.count++;

      if (record.count > maxRequests) {
        return NextResponse.json(
          { error: "Too many requests" },
          { status: 429, headers: { "Retry-After": "60" } }
        );
      }
    }
  }

  return NextResponse.next();
}
```

### Account Lockout

```typescript
// lib/auth-security.ts
export class AuthSecurityService {
  private static readonly MAX_ATTEMPTS = 5;
  private static readonly LOCKOUT_DURATION = 15 * 60 * 1000;

  static async checkLockout(identifier: string): Promise<boolean> {
    const attempts = await redis.get(`auth:attempts:${identifier}`);
    if (!attempts) return false;

    const { count, lockUntil } = JSON.parse(attempts);

    if (count >= this.MAX_ATTEMPTS && Date.now() < lockUntil) {
      return true;
    }

    if (Date.now() > lockUntil) {
      await redis.del(`auth:attempts:${identifier}`);
    }

    return false;
  }

  static async recordFailedAttempt(identifier: string): Promise<void> {
    const key = `auth:attempts:${identifier}`;
    const current = await redis.get(key);

    if (!current) {
      await redis.setex(
        key,
        this.LOCKOUT_DURATION / 1000,
        JSON.stringify({
          count: 1,
          lockUntil: Date.now() + this.LOCKOUT_DURATION,
        })
      );
      return;
    }

    const { count } = JSON.parse(current);
    const newCount = count + 1;

    if (newCount >= this.MAX_ATTEMPTS) {
      await this.logSecurityEvent({
        type: "ACCOUNT_LOCKED",
        identifier,
        timestamp: new Date(),
      });
    }

    await redis.setex(
      key,
      this.LOCKOUT_DURATION / 1000,
      JSON.stringify({
        count: newCount,
        lockUntil: Date.now() + this.LOCKOUT_DURATION,
      })
    );
  }

  static async resetAttempts(identifier: string): Promise<void> {
    await redis.del(`auth:attempts:${identifier}`);
  }

  static async logSecurityEvent(event: {
    type: string;
    identifier: string;
    timestamp: Date;
    userAgent?: string;
    ip?: string;
  }) {
    await db.securityEvents.create(event);
  }
}
```

---

## Database Schema

```sql
-- Users table
CREATE TABLE users (
  id TEXT PRIMARY KEY,
  email TEXT UNIQUE NOT NULL,
  name TEXT NOT NULL,
  password_hash TEXT,
  role TEXT NOT NULL DEFAULT 'user',
  mfa_enabled BOOLEAN DEFAULT false,
  mfa_secret TEXT,
  mfa_enabled_at TIMESTAMP,
  active BOOLEAN DEFAULT true,
  created_at TIMESTAMP DEFAULT NOW()
);

-- WebAuthn credentials
CREATE TABLE webauthn_credentials (
  id SERIAL PRIMARY KEY,
  user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  credential_id BYTEA UNIQUE NOT NULL,
  credential_public_key BYTEA NOT NULL,
  counter BIGINT NOT NULL,
  transports TEXT[],
  created_at TIMESTAMP DEFAULT NOW()
);

-- Security events
CREATE TABLE security_events (
  id SERIAL PRIMARY KEY,
  type TEXT NOT NULL,
  identifier TEXT NOT NULL,
  timestamp TIMESTAMP NOT NULL,
  user_agent TEXT,
  ip TEXT,
  metadata JSON
);
```

---

## Best Practices

### Password Security

```typescript
import bcrypt from "bcryptjs";

// Hash password
const hashPassword = async (password: string) => {
  return await bcrypt.hash(password, 12);
};

// Verify password
const verifyPassword = async (password: string, hash: string) => {
  return await bcrypt.compare(password, hash);
};
```

### JWT Security

```typescript
import jwt from "jsonwebtoken";

const generateToken = (payload: any) => {
  return jwt.sign(payload, process.env.JWT_SECRET!, {
    expiresIn: "7d",
    algorithm: "HS256",
  });
};

const verifyToken = (token: string) => {
  return jwt.verify(token, process.env.JWT_SECRET!);
};
```

---

## Related Skills

- `moai-security-devsecops`: Security automation
- `moai-security-encryption`: Data encryption
- `moai-security-ssrf`: SSRF prevention

---

## Validation Checklist

**Authentication**:

- [ ] NextAuth.js configured
- [ ] OAuth providers set up
- [ ] Credentials provider implemented
- [ ] Session management configured

**MFA**:

- [ ] TOTP implementation complete
- [ ] QR code generation working
- [ ] MFA verification tested
- [ ] Backup codes generated

**WebAuthn**:

- [ ] Passkey registration working
- [ ] Passkey authentication working
- [ ] Credential storage configured
- [ ] Challenge management implemented

**Security**:

- [ ] Rate limiting active
- [ ] Account lockout configured
- [ ] Security events logged
- [ ] Password hashing secure

---

**Last Updated**: 2025-11-20
