---
name: moai-security-auth
description: Modern authentication patterns with OAuth 2.1, WebAuthn, Passkeys, and MFA
modularized: true
modules:
  - jwt-oauth-2-1
  - webauthn-passkeys
  - mfa-patterns
---

## Quick Reference (30 seconds)

# Modern Authentication Security

**OAuth 2.1, WebAuthn, Passkeys, and Multi-Factor Authentication (MFA)**

**Core Capabilities**:
- OAuth 2.1 with mandatory PKCE
- WebAuthn/FIDO2 passwordless authentication
- Passkeys (synced across devices)
- TOTP/SMS Multi-Factor Authentication
- NextAuth.js 5.x and Passport.js 0.7.x

**When to Use**:
- Implementing secure user authentication
- Adding passwordless login (Face ID, Touch ID, Passkeys)
- Migrating from OAuth 2.0 to OAuth 2.1
- Adding multi-factor authentication (MFA)
- Building session management systems

---

## Modules

### 1. [JWT & OAuth 2.1](SKILL-jwt-oauth.md)
Modern OAuth 2.1 authentication with mandatory PKCE, secure JWT patterns, and token rotation.

**Topics**:
- OAuth 2.1 authorization flow with PKCE
- JWT best practices (HS256, RS256)
- Token refresh and rotation
- Secure token storage

### 2. [WebAuthn & Passkeys](SKILL-webauthn.md)
Passwordless authentication with WebAuthn, FIDO2, and Passkeys.

**Topics**:
- WebAuthn registration and authentication
- Passkeys (synced credentials)
- Platform authenticators (Face ID, Touch ID)
- Security key support (YubiKey)

### 3. Multi-Factor Authentication (MFA)
Time-based One-Time Passwords (TOTP), SMS verification, and backup codes.

**Topics**:
- TOTP setup and verification (RFC 6238)
- Backup codes generation
- Rate limiting and brute-force prevention
- MFA recovery flows

---

## Core Implementation

### NextAuth.js 5.x Setup

```typescript
// lib/auth.ts
import NextAuth, { type NextAuthConfig } from 'next-auth';
import GitHub from 'next-auth/providers/github';
import Credentials from 'next-auth/providers/credentials';

const config = {
  providers: [
    // OAuth Provider (GitHub)
    GitHub({
      clientId: process.env.GITHUB_ID,
      clientSecret: process.env.GITHUB_SECRET
    }),
    
    // Credentials Provider (email/password + MFA)
    Credentials({
      credentials: {
        email: { label: 'Email', type: 'email' },
        password: { label: 'Password', type: 'password' },
        mfaCode: { label: '2FA Code', type: 'text', optional: true }
      },
      async authorize(credentials) {
        const user = await db.users.findByEmail(credentials.email);
        
        if (!user) return null;
        
        // Verify password
        const passwordValid = await bcrypt.compare(
          credentials.password,
          user.passwordHash
        );
        
        if (!passwordValid) return null;
        
        // Check MFA if enabled
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
        
        return { id: user.id, email: user.email, name: user.name };
      }
    })
  ],
  
  session: {
    strategy: 'jwt',
    maxAge: 30 * 24 * 60 * 60 // 30 days
  },
  
  callbacks: {
    async jwt({ token, user }) {
      if (user) {
        token.id = user.id;
        token.email = user.email;
      }
      return token;
    },
    
    async session({ session, token }) {
      session.user.id = token.id;
      return session;
    }
  }
} satisfies NextAuthConfig;

export const { handlers, auth, signIn, signOut } = NextAuth(config);
```

### Multi-Factor Authentication (TOTP)

```typescript
// lib/totp-service.ts
import { authenticator } from 'otplib';
import QRCode from 'qrcode';

export class TOTPService {
  async setupTOTP(user: User) {
    // Generate secret
    const secret = authenticator.generateSecret({
      name: `MyApp (${user.email})`
    });
    
    // Generate QR code
    const qrCode = await QRCode.toDataURL(secret);
    
    // Store temporarily
    await redis.setex(`totp:pending:${user.id}`, 600, secret);
    
    return { secret, qrCode };
  }
  
  async verifyTOTPSetup(user: User, token: string) {
    const secret = await redis.get(`totp:pending:${user.id}`);
    if (!secret) throw new Error('No pending TOTP');
    
    // Verify token
    const isValid = authenticator.check(token, secret);
    if (!isValid) throw new Error('Invalid token');
    
    // Generate backup codes
    const backupCodes = Array.from({ length: 10 })
      .map(() => crypto.randomBytes(4).toString('hex').toUpperCase());
    
    // Store permanently
    await db.users.update(user.id, {
      mfaEnabled: true,
      mfaSecret: secret,
      mfaBackupCodes: backupCodes.map(code => bcrypt.hashSync(code, 10))
    });
    
    await redis.del(`totp:pending:${user.id}`);
    
    return backupCodes;
  }
  
  async verifyToken(user: User, token: string) {
    // Check backup codes
    const isBackup = user.mfaBackupCodes.some(code =>
      bcrypt.compareSync(token, code)
    );
    
    if (isBackup) {
      // Mark as used
      const codes = user.mfaBackupCodes.filter(
        code => !bcrypt.compareSync(token, code)
      );
      await db.users.update(user.id, { mfaBackupCodes: codes });
      return true;
    }
    
    // Check TOTP
    return authenticator.check(token, user.mfaSecret);
  }
}
```

---

## Security Standards

**OWASP Top 10 2024 Compliance**:
- A02:2021 - Cryptographic Failures (bcrypt password hashing)
- A07:2021 - Authentication Failures (MFA, rate limiting)

**NIST SP 800-63B Compliance**:
- Password storage (Argon2id, bcrypt)
- Multi-factor authentication
- Session management

**FIDO2 & WebAuthn Compliance**:
- FIDO2 specifications
- W3C WebAuthn Level 2

---

## Context7 Integration

### Related Libraries & Tools
- [bcryptjs](/dcodeIO/bcrypt.js): Bcrypt password hashing
- [argon2](/P-H-C/phc-winner-argon2): Argon2 password hashing
- [jwt](/auth0/node-jsonwebtoken): JWT authentication
- [oauth2-server](/oauthjs/node-oauth2-server): OAuth 2.0 authorization server
- [passport](/jaredhanson/passport): Authentication middleware

### Official Documentation
- [NIST SP 800-63B](https://pages.nist.gov/800-63-3/sp800-63b.html)
- [JWT.io](https://jwt.io/)
- [Passport.js](http://www.passportjs.org/)

### Version-Specific Guides
Latest stable versions: bcryptjs, argon2, JWT, OAuth 2.0
- [Password Storage Best Practices](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html)
- [MFA Implementation Guide](https://cheatsheetseries.owasp.org/cheatsheets/Multifactor_Authentication_Cheat_Sheet.html)
- [Session Management Guide](https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html)

---

## Version History

**v5.0.0** (2025-11-22)
- OAuth 2.1 compliance (PKCE mandatory)
- Removed deprecated Implicit Flow
- Added Passkeys module
- Updated NextAuth.js 5.x patterns

**v4.0.0** (2025-11-12)
- Added WebAuthn passwordless authentication
- TOTP MFA implementation
- NextAuth.js 5.x migration

---

**Last Updated**: 2025-11-22  
**Classification**: Enterprise Authentication Security  
**Compliance**: OAuth 2.1, FIDO2, OWASP Top 10 2024
